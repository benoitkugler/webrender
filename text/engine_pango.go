package text

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/language"
	fc "github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/textprocessing/pango"
	"github.com/benoitkugler/textprocessing/pango/fcfonts"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/text/hyphen"
	"github.com/benoitkugler/webrender/utils"
)

func PangoUnitsFromFloat(v pr.Fl) int32 { return int32(v*pango.Scale + 0.5) }

func PangoUnitsToFloat(v pango.Unit) pr.Fl { return pr.Fl(v) / pango.Scale }

// FontConfigurationPango holds information about the
// available fonts on the system.
// It is used for text layout at various steps of the process.
type FontConfigurationPango struct {
	fontmap *fcfonts.FontMap

	userFonts    map[FontOrigin]fonts.Face
	fontsContent map[string][]byte // to be embedded in the target
}

// NewFontConfigurationPango uses a fontconfig database to create a new
// font configuration
func NewFontConfigurationPango(fontmap *fcfonts.FontMap) *FontConfigurationPango {
	out := &FontConfigurationPango{
		fontmap:      fontmap,
		userFonts:    make(map[FontOrigin]fonts.Face),
		fontsContent: make(map[string][]byte),
	}
	out.fontmap.SetFaceLoader(out)
	return out
}

func (f *FontConfigurationPango) LoadFace(key fonts.FaceID, format fc.FontFormat) (fonts.Face, error) {
	if face, has := f.userFonts[FontOrigin(key)]; has {
		return face, nil
	}
	return fcfonts.DefaultLoadFace(key, format)
}

func (fc *FontConfigurationPango) spaceHeight(style *TextStyle) (height, baseline pr.Float) {
	layout := newTextLayout(fc, style, nil)
	layout.SetText(" ")
	line, _ := layout.GetFirstLine()
	sp := firstLineMetrics(line, nil, layout, -1, false, style, false, "")
	return sp.Height, sp.Baseline
}

func (fc *FontConfigurationPango) width0(style *TextStyle) pr.Fl {
	p := newTextLayout(fc, style, nil)

	p.Layout.SetText("0") // avoid recursion for letter-spacing and word-spacing properties
	line, _ := p.GetFirstLine()
	var logicalExtents pango.Rectangle
	line.GetExtents(nil, &logicalExtents)
	return PangoUnitsToFloat(logicalExtents.Width)
}

// var styles []*TextStyle

// func dumpStyle(s *TextStyle) {
// 	styles = append(styles, s)
// 	f, _ := os.Create("styles.json")
// 	enc := json.NewEncoder(f)
// 	enc.SetIndent(" ", " ")
// 	enc.Encode(styles)
// 	f.Close()
// }

func (fc *FontConfigurationPango) heightx(style *TextStyle) pr.Fl {
	p := newTextLayout(fc, style, nil)

	p.Layout.SetText("x") // avoid recursion for letter-spacing and word-spacing properties
	line, _ := p.GetFirstLine()
	var inkExtents pango.Rectangle
	line.GetExtents(&inkExtents, nil)
	return -PangoUnitsToFloat(inkExtents.Y)
}

type runeProp uint8

// bit mask
const (
	isWordEnd runeProp = 1 << iota
	isWordStart
	isLineBreak
)

func (fc *FontConfigurationPango) runeProps(text []rune) []runeProp {
	text = []rune(bidiMarkReplacer.Replace(string(text)))
	logAttrs := pango.ComputeCharacterAttributes(text, -1)
	out := make([]runeProp, len(logAttrs))
	for i, p := range logAttrs {
		if p.IsWordStart() {
			out[i] |= isWordStart
		}
		if p.IsWordEnd() {
			out[i] |= isWordEnd
		}
		if p.IsLineBreak() {
			out[i] |= isLineBreak
		}
	}
	return out
}

func (fc FontConfigurationPango) CanBreakText(t []rune) pr.MaybeBool {
	if len(t) < 2 {
		return nil
	}
	logs := fc.runeProps(t)
	for _, l := range logs[1 : len(logs)-1] {
		if l&isLineBreak != 0 {
			return pr.True
		}
	}
	return pr.False
}

// FontContent returns the content of the given face, which may be needed
// in the final output.
func (f *FontConfigurationPango) FontContent(font FontOrigin) []byte {
	// either the font is loaded at run time or is loaded from disk
	if content, has := f.fontsContent[font.File]; has {
		return content
	}

	b, err := os.ReadFile(font.File)
	if err != nil {
		logger.WarningLogger.Println(err)
	}
	// cache the result to avoid loading the same file over and over
	f.fontsContent[font.File] = b

	return b
}

func (f *FontConfigurationPango) AddFontFace(ruleDescriptors validation.FontFaceDescriptors, urlFetcher utils.UrlFetcher) string {
	if f.fontmap == nil {
		return ""
	}

	for _, url := range ruleDescriptors.Src {
		if url.String == "" {
			continue
		}
		if !(url.Name == "external" || url.Name == "local") {
			continue
		}

		filename, err := f.loadOneFont(url, ruleDescriptors, urlFetcher)
		if err != nil {
			logger.WarningLogger.Println(err)
			continue
		}

		return filename
	}

	logger.WarningLogger.Printf("Font-face %s cannot be loaded", ruleDescriptors.FontFamily)
	return ""
}

// make `s` a valid xml string content
func escapeXML(s string) string {
	var b strings.Builder
	xml.EscapeText(&b, []byte(s))
	return b.String()
}

func (f *FontConfigurationPango) loadOneFont(url pr.NamedString, ruleDescriptors validation.FontFaceDescriptors, urlFetcher utils.UrlFetcher) (string, error) {
	config := f.fontmap.Config

	if url.Name == "local" {
		fontName := url.String
		pattern := fc.NewPattern()
		config.Substitute(pattern, nil, fc.MatchResult)
		pattern.SubstituteDefault()
		pattern.AddString(fc.FULLNAME, fontName)
		pattern.AddString(fc.POSTSCRIPT_NAME, fontName)
		matchingPattern := f.fontmap.Database.Match(pattern, config)

		// prevent RuntimeError, see issue #677
		if matchingPattern == nil {
			return "", fmt.Errorf("failed to get matching local font for %s", fontName)
		}

		family, _ := matchingPattern.GetString(fc.FULLNAME)
		postscript, _ := matchingPattern.GetString(fc.POSTSCRIPT_NAME)
		if fn := strings.ToLower(fontName); fn == strings.ToLower(family) || fn == strings.ToLower(postscript) {
			filename := matchingPattern.FaceID().File
			var err error
			url.String, err = filepath.Abs(filename)
			if err != nil {
				return "", fmt.Errorf("failed to load local font %s: %s", fontName, err)
			}
		} else {
			return "", fmt.Errorf("failed to load local font %s", fontName)
		}
	}

	result, err := urlFetcher(url.String)
	if err != nil {
		return "", fmt.Errorf("failed to load font at: %s", err)
	}
	fontFilename := escapeXML(url.String)

	content, err := io.ReadAll(result.Content)
	if err != nil {
		return "", fmt.Errorf("failed to load font at %s", url.String)
	}

	faces, format := fc.ReadFontFile(bytes.NewReader(content))
	if format == "" {
		return "", fmt.Errorf("failed to load font at %s : unsupported format", fontFilename)
	}

	if len(faces) != 1 {
		return "", fmt.Errorf("font collections are not supported (%s)", url.String)
	}

	if url.Name == "external" {
		key := FontOrigin{
			File: fontFilename,
		}
		f.userFonts[key] = faces[0]
		f.fontsContent[key.File] = content
	}

	featuresString := ""
	for _, v := range getFontFaceFeatures(ruleDescriptors) {
		featuresString += fmt.Sprintf("<string>%s=%d</string>", v.Tag[:], v.Value)
	}
	fontconfigStyle, ok := fcStyle[ruleDescriptors.FontStyle]
	if !ok {
		fontconfigStyle = "roman"
	}
	fontconfigWeight, ok := fcWeight[ruleDescriptors.FontWeight]
	if !ok {
		fontconfigWeight = "regular"
	}
	fontconfigStretch, ok := fcStretch[ruleDescriptors.FontStretch]
	if !ok {
		fontconfigStretch = "normal"
	}

	xmlConfig := fmt.Sprintf(`<?xml version="1.0"?>
		<!DOCTYPE fontconfig SYSTEM "fonts.dtd">
		<fontconfig>
		  <match target="scan">
			<test name="file" compare="eq">
			  <string>%s</string>
			</test>
			<edit name="family" mode="assign_replace">
			  <string>%s</string>
			</edit>
			<edit name="slant" mode="assign_replace">
			  <const>%s</const>
			</edit>
			<edit name="weight" mode="assign_replace">
			  <const>%s</const>
			</edit>
			<edit name="width" mode="assign_replace">
			  <const>%s</const>
			</edit>
		  </match>
		  <match target="font">
			<test name="family" compare="eq">
			  <string>%s</string>
			</test>
			<edit name="fontfeatures" mode="assign_replace">%s</edit>
		  </match>
		</fontconfig>`, fontFilename, ruleDescriptors.FontFamily, fontconfigStyle,
		fontconfigWeight, fontconfigStretch, ruleDescriptors.FontFamily, featuresString)

	err = config.LoadFromMemory(bytes.NewReader([]byte(xmlConfig)))
	if err != nil {
		return "", fmt.Errorf("failed to load fontconfig config: %s", err)
	}

	fs, err := config.ScanFontRessource(bytes.NewReader(content), fontFilename)
	if err != nil {
		return "", fmt.Errorf("failed to load font at %s", url.String)
	}

	f.fontmap.Database = append(f.fontmap.Database, fs...)
	f.fontmap.SetConfig(config, f.fontmap.Database)
	return fontFilename, nil
}

// Fontconfig features
var (
	fcWeight = map[pr.IntString]string{
		{String: "normal"}: "regular",
		{String: "bold"}:   "bold",
		{Int: 100}:         "thin",
		{Int: 200}:         "extralight",
		{Int: 300}:         "light",
		{Int: 400}:         "regular",
		{Int: 500}:         "medium",
		{Int: 600}:         "demibold",
		{Int: 700}:         "bold",
		{Int: 800}:         "extrabold",
		{Int: 900}:         "black",
	}
	fcStyle = map[pr.String]string{
		"normal":  "roman",
		"italic":  "italic",
		"oblique": "oblique",
	}
	fcStretch = map[pr.String]string{
		"normal":          "normal",
		"ultra-condensed": "ultracondensed",
		"extra-condensed": "extracondensed",
		"condensed":       "condensed",
		"semi-condensed":  "semicondensed",
		"semi-expanded":   "semiexpanded",
		"expanded":        "expanded",
		"extra-expanded":  "extraexpanded",
		"ultra-expanded":  "ultraexpanded",
	}
)

func getFontDescription(fd FontDescription) pango.FontDescription {
	fontDesc := pango.NewFontDescription()
	fontDesc.SetFamily(strings.Join(fd.Family, ","))

	fontDesc.SetStyle(pango.Style(fd.Style))
	fontDesc.SetStretch(pango.Stretch(fd.Stretch))
	fontDesc.SetWeight(pango.Weight(fd.Weight))

	fontDesc.SetAbsoluteSize(PangoUnitsFromFloat(fd.Size))

	fontDesc.SetVariations(pangoFontVariations(fd.VariationSettings))

	return fontDesc
}

func pangoFontVariations(vs []Variation) string {
	chunks := make([]string, len(vs))
	for i, v := range vs {
		chunks[i] = fmt.Sprintf("%s=%f", v.Tag[:], v.Value)
	}
	return strings.Join(chunks, ",")
}

func pangoFontFeatures(vs []Feature) string {
	chunks := make([]string, len(vs))
	for i, v := range vs {
		chunks[i] = fmt.Sprintf("%s=%d", v.Tag[:], v.Value)
	}
	return strings.Join(chunks, ",")
}

func firstLineMetrics(firstLine *pango.LayoutLine, text []rune, layout *TextLayoutPango, resumeAt int, spaceCollapse bool,
	style *TextStyle, hyphenated bool, hyphenationCharacter string,
) FirstLine {
	length := firstLine.Length
	if hyphenated {
		length -= len([]rune(hyphenationCharacter))
	} else if resumeAt != -1 && resumeAt != 0 {
		// Set an infinite width as we don't want to break lines when drawing,
		// the lines have already been split and the size may differ. Rendering
		// is also much faster when no width is set.
		layout.Layout.SetWidth(-1)

		// Create layout with final text
		if length > len(text) {
			length = len(text)
		}
		firstLineText := string(text[:length])

		// Remove trailing spaces if spaces collapse
		if spaceCollapse {
			firstLineText = strings.TrimRight(firstLineText, " ")
		}

		layout.SetText(firstLineText)

		firstLine, _ = layout.GetFirstLine()
		length = 0
		if firstLine != nil {
			length = firstLine.Length
		}
	}

	// FIXME:
	if resumeAt > len(text) {
		resumeAt = len(text)
	}

	width, height := lineSize(firstLine, style.LetterSpacing)
	baseline := PangoUnitsToFloat(layout.Layout.GetBaseline())
	return FirstLine{
		Layout: layout,
		Length: length, ResumeAt: resumeAt,
		Width: pr.Float(width), Height: pr.Float(height), Baseline: pr.Float(baseline),
		FirstLineRTL: firstLine.ResolvedDir%2 != 0,
	}
}

// TextLayoutPango wraps a pango.Layout object
type TextLayoutPango struct {
	Style   *TextStyle
	metrics *LineMetrics // optional

	MaxWidth pr.MaybeFloat

	fonts FontConfiguration // will be a *LayoutContext; to avoid circular dependency

	Layout pango.Layout

	justificationSpacing pr.Fl
}

func newTextLayout(fonts FontConfiguration, style *TextStyle, maxWidth pr.MaybeFloat) *TextLayoutPango {
	var layout TextLayoutPango

	layout.setup(fonts, style)
	layout.MaxWidth = maxWidth

	return &layout
}

// createLayout returns a pango.Layout with default Pango line-breaks.
// `style` is a style dict of computed values.
// `maxWidth` is the maximum available width in the same unit as style.FontSize,
// or `nil` for unlimited width.
func createLayout(text string, style *TextStyle, fonts FontConfiguration, maxWidth pr.MaybeFloat) *TextLayoutPango {
	layout := newTextLayout(fonts, style, maxWidth)
	textWrap := style.textWrap()
	if maxWidth, ok := maxWidth.(pr.Float); ok && textWrap && maxWidth < 2<<21 {
		// Make sure that maxWidth * Pango.SCALE == maxWidth * 1024 fits in a
		// signed integer. Treat bigger values same as None: unconstrained width.
		layout.Layout.SetWidth(pango.Unit(PangoUnitsFromFloat(utils.Maxs(0, pr.Fl(maxWidth)))))
	}

	layout.SetText(text)
	return layout
}

// Text returns a readonly slice of the text used in the layout.
func (p *TextLayoutPango) Text() []rune { return p.Layout.Text }

func (p *TextLayoutPango) Metrics() *LineMetrics { return p.metrics }

func (p *TextLayoutPango) Justification() pr.Float           { return pr.Float(p.justificationSpacing) }
func (p *TextLayoutPango) SetJustification(spacing pr.Float) { p.justificationSpacing = pr.Fl(spacing) }

func (p *TextLayoutPango) setup(fonts FontConfiguration, style *TextStyle) {
	p.fonts = fonts
	p.Style = style
	fontmap := fonts.(*FontConfigurationPango).fontmap
	pc := pango.NewContext(fontmap)
	pc.SetRoundGlyphPositions(false)

	var lang pango.Language
	if flo := style.FontLanguageOverride; (flo != fontLanguageOverride{}) {
		lang = language.NewLanguage(lstToISO[flo])
	} else if lg := style.Lang; lg != "" {
		lang = language.NewLanguage(lg)
	} else {
		lang = pango.DefaultLanguage()
	}
	pc.SetLanguage(lang)

	fontDesc := getFontDescription(style.FontDescription)
	p.Layout = *pango.NewLayout(pc)
	p.Layout.SetFontDescription(&fontDesc)

	if style.TextDecorationLine != 0 {
		metrics := pc.GetMetrics(&fontDesc, lang)
		p.metrics = &LineMetrics{
			Ascent:                 PangoUnitsToFloat(metrics.Ascent),
			UnderlinePosition:      PangoUnitsToFloat(metrics.UnderlinePosition),
			UnderlineThickness:     PangoUnitsToFloat(metrics.UnderlineThickness),
			StrikethroughPosition:  PangoUnitsToFloat(metrics.StrikethroughPosition),
			StrikethroughThickness: PangoUnitsToFloat(metrics.StrikethroughThickness),
		}
	} else {
		p.metrics = nil
	}

	if len(style.FontFeatures) != 0 {
		attr := pango.NewAttrFontFeatures(pangoFontFeatures(style.FontFeatures))
		p.Layout.SetAttributes(pango.AttrList{attr})
	}
}

func (p *TextLayoutPango) SetText(text string) { p.setText(text, false) }

// ApplyJustification re-layout the text, applying justification.
func (p *TextLayoutPango) ApplyJustification() {
	p.Layout.SetWidth(-1)
	p.setText(string(p.Layout.Text), true)
}

func (p *TextLayoutPango) setText(text string, justify bool) {
	if index := strings.IndexByte(text, '\n'); index != -1 && len(text) >= index+2 {
		// Keep only the first line plus one character, we don't need more
		text = text[:index+2]
	}

	p.Layout.SetText(text)

	wordSpacing := p.Style.WordSpacing
	if justify {
		// Justification is needed when drawing text but is useless during
		// layout, when it can be ignored.
		wordSpacing += p.justificationSpacing
	}

	letterSpacing := p.Style.LetterSpacing

	wordBreaking := p.Style.OverflowWrap == OAnywhere || p.Style.OverflowWrap == OBreakWord

	if text != "" && (wordSpacing != 0 || letterSpacing != 0 || wordBreaking) {
		letterSpacingInt := PangoUnitsFromFloat(letterSpacing)
		spaceSpacingInt := PangoUnitsFromFloat(wordSpacing) + letterSpacingInt
		attrList := p.Layout.Attributes

		addAttr := func(start, end int, spacing int32) {
			attr := pango.NewAttrLetterSpacing(spacing)
			attr.StartIndex, attr.EndIndex = start, end
			attrList.Change(attr)
		}

		textRunes := p.Layout.Text

		if letterSpacing != 0 {
			addAttr(0, len(textRunes), letterSpacingInt)
		}

		if wordSpacing != 0 {
			if len(textRunes) == 1 && textRunes[0] == ' ' {
				// We need more than one space to set word spacing
				p.Layout.SetText(" \u200b") // Space + zero-width space
			}

			for position, c := range textRunes {
				if c == ' ' {
					// Pango gives only half of word-spacing on boundaries
					factor := int32(1)
					if position == 0 || position == len(textRunes)-1 {
						factor = 2
					}
					addAttr(position, position+1, factor*spaceSpacingInt)
				}
			}
		}

		if wordBreaking {
			attr := pango.NewAttrInsertHyphens(false)
			attr.StartIndex, attr.EndIndex = 0, len(textRunes)
			attrList.Change(attr)
		}

		p.Layout.SetAttributes(attrList)
	}

	// Tabs width
	if strings.ContainsRune(text, '\t') {
		p.setTabs()
	}
}

func (p *TextLayoutPango) setTabs() {
	tabSize := p.Style.TabSize
	width := tabSize.Width
	if tabSize.IsMultiple { // no unit, means a multiple of the advance width of the space character
		layout := newTextLayout(p.fonts, p.Style, nil)
		layout.SetText(strings.Repeat(" ", width))
		line, _ := layout.GetFirstLine()
		widthTmp, _ := lineSize(line, p.Style.LetterSpacing)
		width = int(widthTmp + 0.5)
	}
	// 0 is not handled correctly by Pango
	if width == 0 {
		width = 1
	}
	tabs := &pango.TabArray{Tabs: []pango.Tab{{Alignment: pango.TAB_LEFT, Location: pango.Unit(width)}}, PositionsInPixels: true}
	p.Layout.SetTabs(tabs)
}

// GetFirstLine returns the first line and the index of the second line, or -1.
func (p *TextLayoutPango) GetFirstLine() (*pango.LayoutLine, int) {
	firstLine := p.Layout.GetLine(0)
	secondLine := p.Layout.GetLine(1)
	index := -1
	if secondLine != nil {
		index = secondLine.StartIndex
	}

	return firstLine, index
}

// lineSize gets the logical width and height of the given `line`.
// [letterSpacing] is added, a value of 0 has no impact
func lineSize(line *pango.LayoutLine, letterSpacing pr.Fl) (pr.Fl, pr.Fl) {
	var logicalExtents pango.Rectangle
	line.GetExtents(nil, &logicalExtents)
	width := PangoUnitsToFloat(logicalExtents.Width)
	height := PangoUnitsToFloat(logicalExtents.Height)
	width += letterSpacing
	return width, height
}

func (fc *FontConfigurationPango) splitFirstLine(hyphenCache map[HyphenDictKey]hyphen.Hyphener, text []rune, style *TextStyle,
	maxWidth pr.MaybeFloat, minimum, isLineStart bool,
) FirstLine {
	// See https://www.w3.org/TR/css-text-3/#white-space-property
	var (
		ws               = style.WhiteSpace
		textWrap         = style.textWrap()
		spaceCollapse    = ws == WNormal || ws == WNowrap || ws == WPreLine
		originalMaxWidth = maxWidth
		layout           *TextLayoutPango
		fontSize         = pr.Float(style.Size)
		firstLine        *pango.LayoutLine
		resumeIndex      int
		text_            = string(text)
	)
	if !textWrap {
		maxWidth = nil
	}
	// Step #1: Get a draft layout with the first line
	if maxWidth, ok := maxWidth.(pr.Float); ok && maxWidth != pr.Inf && fontSize != 0 {
		// Try to use a small amount of text instead of the whole text
		shortText := shortTextHint(text, maxWidth, fontSize)

		layout = createLayout(string(shortText), style, fc, maxWidth)
		firstLine, resumeIndex = layout.GetFirstLine()
		if resumeIndex == -1 && len(shortText) != len(text) {
			// The small amount of text fits in one line, give up and use the whole text
			layout.SetText(text_)
			firstLine, resumeIndex = layout.GetFirstLine()
		}
	} else {
		layout = createLayout(text_, style, fc, originalMaxWidth)
		firstLine, resumeIndex = layout.GetFirstLine()
	}

	// Step #2: Don't split lines when it's not needed
	if maxWidth == nil {
		// The first line can take all the place needed
		return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	}
	maxWidthV := pr.Fl(maxWidth.V())

	firstLineWidth, _ := lineSize(firstLine, style.LetterSpacing)

	if resumeIndex == -1 && firstLineWidth <= maxWidthV {
		// The first line fits in the available width
		return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	}

	// Step #3: Try to put the first word of the second line on the first line
	// https://mail.gnome.org/archives/gtk-i18n-list/2013-September/msg00006
	// is a good thread related to this problem.

	firstLineText := text_
	if resumeIndex != -1 && resumeIndex <= len(text) {
		firstLineText = string(text[:resumeIndex])
	}
	firstLineFits := (firstLineWidth <= maxWidthV ||
		strings.ContainsRune(strings.TrimSpace(firstLineText), ' ') ||
		fc.CanBreakText([]rune(strings.TrimSpace(firstLineText))) == pr.True)
	var secondLineText []rune
	if firstLineFits {
		// The first line fits but may have been cut too early by Pango
		if resumeIndex == -1 {
			secondLineText = text
		} else {
			secondLineText = text[resumeIndex:]
		}
	} else {
		// The line can't be split earlier, try to hyphenate the first word.
		firstLineText = ""
		secondLineText = text
	}

	nextWord := strings.SplitN(string(secondLineText), " ", 2)[0]
	if nextWord != "" {
		if spaceCollapse {
			// nextWord might fit without a space afterwards
			// only try when space collapsing is allowed
			newFirstLineText := firstLineText + nextWord
			layout.SetText(newFirstLineText)
			firstLine, resumeIndex = layout.GetFirstLine()
			// firstLineWidth, _ = lineSize(firstLine, style.GetLetterSpacing())
			if resumeIndex == -1 {
				if firstLineText != "" {
					// The next word fits in the first line, keep the layout
					resumeIndex = len([]rune(newFirstLineText)) + 1
					return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
				} else {
					// Second line is none
					resumeIndex = firstLine.Length + 1
					if resumeIndex >= len(text) {
						resumeIndex = -1
					}
				}
			}
		}
	} else if firstLineText != "" {
		// We found something on the first line but we did not find a word on
		// the next line, no need to hyphenate, we can keep the current layout
		return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	}

	// Step #4: Try to hyphenate
	hyphens := style.Hyphens
	lang := language.NewLanguage(style.Lang)
	if lang != "" {
		lang = hyphen.LanguageFallback(lang)
	}
	limit := style.HyphenateLimitChars
	hyphenateCharacter := style.HyphenateCharacter
	hyphenated := false
	softHyphen := '\u00ad'

	autoHyphenation, manualHyphenation := false, false
	if hyphens != HNone {
		manualHyphenation = strings.ContainsRune(firstLineText, softHyphen) || strings.ContainsRune(nextWord, softHyphen)
	}

	var startWord, stopWord int
	if hyphens == HAuto && lang != "" {
		nextWordBoundaries := fc.wordBoundaries(secondLineText)
		if len(nextWordBoundaries) == 2 {
			// We have a word to hyphenate
			startWord, stopWord = nextWordBoundaries[0], nextWordBoundaries[1]
			nextWord = string(secondLineText[startWord:stopWord])
			if stopWord-startWord >= limit.Total {
				// This word is long enough
				firstLineWidth, _ = lineSize(firstLine, style.LetterSpacing)
				space := maxWidthV - firstLineWidth
				zone := style.HyphenateLimitZone
				limitZone := zone.Limit
				if zone.IsPercentage {
					limitZone = (maxWidthV * zone.Limit / 100.)
				}
				if space > limitZone || space < 0 {
					// Available space is worth the try, or the line is even too
					// long to fit: try to hyphenate
					autoHyphenation = true
				}
			}
		}
	}

	// Automatic hyphenation opportunities within a word must be ignored if the
	// word contains a conditional hyphen, in favor of the conditional
	// hyphen(s).
	// See https://drafts.csswg.org/css-text-3/#valdef-hyphens-auto
	var dictionaryIterations []string
	if manualHyphenation {
		// Manual hyphenation: check that the line ends with a soft
		// hyphen and add the missing hyphen
		if strings.HasSuffix(firstLineText, string(softHyphen)) {
			// The first line has been split on a soft hyphen
			if id := strings.LastIndexByte(firstLineText, ' '); id != -1 {
				firstLineText, nextWord = firstLineText[:id], firstLineText[id+1:]
				nextWord = " " + nextWord
				layout.SetText(firstLineText)
				firstLine, _ = layout.GetFirstLine()
				resumeIndex = len([]rune(firstLineText + " "))
			} else {
				firstLineText, nextWord = "", firstLineText
			}
		}
		dictionaryIterations = hyphenDictionaryIterationsOld(nextWord, softHyphen)
	} else if autoHyphenation {
		dictionaryKey := HyphenDictKey{lang, limit}
		dictionary, ok := hyphenCache[dictionaryKey]
		if !ok {
			dictionary = hyphen.NewHyphener(lang, limit.Left, limit.Right)
			hyphenCache[dictionaryKey] = dictionary
		}
		dictionaryIterations = dictionary.Iterate(nextWord)
	}

	if len(dictionaryIterations) != 0 {
		var newFirstLineText, hyphenatedFirstLineText string
		for _, firstWordPart := range dictionaryIterations {
			newFirstLineText = (firstLineText + string(secondLineText[:startWord]) + firstWordPart)
			hyphenatedFirstLineText = (newFirstLineText + hyphenateCharacter)
			newLayout := createLayout(hyphenatedFirstLineText, style, fc, maxWidth)
			newFirstLine, newIndex := newLayout.GetFirstLine()
			newFirstLineWidth, _ := lineSize(newFirstLine, style.LetterSpacing)
			newSpace := maxWidthV - newFirstLineWidth
			hyphenated = newIndex == -1 && (newSpace >= 0 || firstWordPart == dictionaryIterations[len(dictionaryIterations)-1])
			if hyphenated {
				layout = newLayout
				firstLine = newFirstLine
				resumeIndex = len([]rune(newFirstLineText))
				break
			}
		}

		if !hyphenated && firstLineText == "" {
			// Recreate the layout with no maxWidth to be sure that
			// we don't break before or inside the hyphenate character
			hyphenated = true
			layout.SetText(hyphenatedFirstLineText)
			layout.Layout.SetWidth(-1)
			firstLine, _ = layout.GetFirstLine()
			resumeIndex = len([]rune(newFirstLineText))
			if text[resumeIndex] == softHyphen {
				resumeIndex += 1
			}
		}
	}

	if !hyphenated && strings.HasSuffix(firstLineText, string(softHyphen)) {
		// Recreate the layout with no maxWidth to be sure that
		// we don't break inside the hyphenate-character string
		hyphenated = true
		hyphenatedFirstLineText := firstLineText + hyphenateCharacter
		layout.SetText(hyphenatedFirstLineText)
		layout.Layout.SetWidth(-1)
		firstLine, _ = layout.GetFirstLine()
		resumeIndex = len([]rune(firstLineText))
	}

	// Step 5: Try to break word if it's too long for the line
	overflowWrap, wordBreak := style.OverflowWrap, style.WordBreak
	firstLineWidth, _ = lineSize(firstLine, style.LetterSpacing)
	space := maxWidthV - firstLineWidth
	// If we can break words and the first line is too long
	canBreak := wordBreak == WBBreakAll ||
		(isLineStart && (overflowWrap == OAnywhere || (overflowWrap == OBreakWord && !minimum)))
	if space < 0 && canBreak {
		// Is it really OK to remove hyphenation for word-break ?
		hyphenated = false
		layout.SetText(string(text))
		layout.Layout.SetWidth(pango.Unit(PangoUnitsFromFloat(maxWidthV)))
		layout.Layout.SetWrap(pango.WRAP_CHAR)
		var index int
		firstLine, index = layout.GetFirstLine()
		resumeIndex = index
		if resumeIndex == 0 {
			resumeIndex = firstLine.Length
		}
		if resumeIndex >= len(text) {
			resumeIndex = -1
		}
	}

	return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, hyphenated, hyphenateCharacter)
}

// split word on each hyphen occurence, starting by the end
func hyphenDictionaryIterationsOld(word string, hyphen rune) (out []string) {
	wordRunes := []rune(word)
	for i := len(wordRunes) - 1; i >= 0; i-- {
		if wordRunes[i] == hyphen {
			out = append(out, string(wordRunes[:i+1]))
		}
	}
	return out
}

var bidiMarkReplacer = strings.NewReplacer(
	"\u202a", "\u200b",
	"\u202b", "\u200b",
	"\u202c", "\u200b",
	"\u202d", "\u200b",
	"\u202e", "\u200b",
)

// returns nil or [wordStart, wordEnd]
func (fc *FontConfigurationPango) wordBoundaries(t []rune) *[2]int {
	if len(t) < 2 {
		return nil
	}
	var out [2]int
	hasBroken := false
	for i, attr := range fc.runeProps(t) {
		if attr&isWordEnd != 0 {
			out[1] = i // word end
			hasBroken = true
			break
		}
		if attr&isWordStart != 0 {
			out[0] = i // word start
		}
	}
	if !hasBroken {
		return nil
	}
	return &out
}

// GetLastWordEnd returns the index in `t` of the last word,
// or -1
func GetLastWordEnd(fc *FontConfigurationPango, t []rune) int {
	if len(t) < 2 {
		return -1
	}
	attrs := fc.runeProps(t)
	for i := 0; i < len(attrs); i++ {
		item := attrs[len(attrs)-1-i]
		if i != 0 && item&isWordEnd != 0 {
			return len(t) - i
		}
	}
	return -1
}
