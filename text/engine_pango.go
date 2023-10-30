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

	var inkExtents, logicalExtents pango.Rectangle
	line.GetExtents(&inkExtents, &logicalExtents)
	return PangoUnitsToFloat(logicalExtents.Width)
}

func (fc *FontConfigurationPango) heightx(style *TextStyle) pr.Fl {
	p := newTextLayout(fc, style, nil)

	p.Layout.SetText("x") // avoid recursion for letter-spacing and word-spacing properties
	line, _ := p.GetFirstLine()

	var inkExtents, logicalExtents pango.Rectangle
	line.GetExtents(&inkExtents, &logicalExtents)
	return -PangoUnitsToFloat(inkExtents.Y)
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

	features := pr.Properties{}
	// avoid nil values
	features.SetFontKerning("")
	features.SetFontVariantLigatures(pr.SStrings{})
	features.SetFontVariantPosition("")
	features.SetFontVariantCaps("")
	features.SetFontVariantNumeric(pr.SStrings{})
	features.SetFontVariantAlternates("")
	features.SetFontVariantEastAsian(pr.SStrings{})
	features.SetFontFeatureSettings(pr.SIntStrings{})
	for _, rules := range ruleDescriptors.FontVariant {
		if rules.Property.SpecialProperty != nil {
			continue
		}
		if cascaded := rules.Property.ToCascaded(); cascaded.Default == 0 {
			features[rules.Name.KnownProp] = cascaded.ToCSS()
		}
	}
	if !ruleDescriptors.FontFeatureSettings.IsNone() {
		features.SetFontFeatureSettings(ruleDescriptors.FontFeatureSettings)
	}
	featuresString := ""
	for _, v := range getFontFeatures(features) {
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

func getFontDescription(style *TextStyle) pango.FontDescription {
	fontDesc := pango.NewFontDescription()
	fontDesc.SetFamily(strings.Join(style.FontFamily, ","))

	fontDesc.SetStyle(pango.Style(style.FontStyle))
	fontDesc.SetStretch(pango.Stretch(style.FontStretch))
	fontDesc.SetWeight(pango.Weight(style.FontWeight))

	fontDesc.SetAbsoluteSize(PangoUnitsFromFloat(style.FontSize))

	fontDesc.SetVariations(pangoFontVariations(style.FontVariationSettings))

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
) Splitted {
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

	width, height := lineSize(firstLine, style.LetterSpacing)
	baseline := PangoUnitsToFloat(layout.Layout.GetBaseline())
	return Splitted{
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
	ws := style.WhiteSpace
	textWrap := ws == WNormal || ws == WPreWrap || ws == WPreLine
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
		lang = lstToISO[flo]
	} else if lg := style.Lang; lg != "" {
		lang = language.NewLanguage(lg)
	} else {
		lang = pango.DefaultLanguage()
	}
	pc.SetLanguage(lang)

	fontDesc := getFontDescription(style)
	p.Layout = *pango.NewLayout(pc)
	p.Layout.SetFontDescription(&fontDesc)

	if !style.TextDecorationLine.IsNone() {
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

// Language system tags
// From https://docs.microsoft.com/typography/opentype/spec/languagetags
var lstToISO = map[fontLanguageOverride]language.Language{
	{'a', 'b', 'a'}:      "abq",
	{'a', 'f', 'k'}:      "afr",
	{'a', 'f', 'r'}:      "aar",
	{'a', 'g', 'w'}:      "ahg",
	{'a', 'l', 's'}:      "gsw",
	{'a', 'l', 't'}:      "atv",
	{'a', 'r', 'i'}:      "aiw",
	{'a', 'r', 'k'}:      "mhv",
	{'a', 't', 'h'}:      "apk",
	{'a', 'v', 'r'}:      "ava",
	{'b', 'a', 'd'}:      "bfq",
	{'b', 'a', 'd', '0'}: "bad",
	{'b', 'a', 'g'}:      "bfy",
	{'b', 'a', 'l'}:      "krc",
	{'b', 'a', 'u'}:      "bci",
	{'b', 'c', 'h'}:      "bcq",
	{'b', 'g', 'r'}:      "bul",
	{'b', 'i', 'l'}:      "byn",
	{'b', 'k', 'f'}:      "bla",
	{'b', 'l', 'i'}:      "bal",
	{'b', 'l', 'n'}:      "bjt",
	{'b', 'l', 't'}:      "bft",
	{'b', 'm', 'b'}:      "bam",
	{'b', 'r', 'i'}:      "bra",
	{'b', 'r', 'm'}:      "mya",
	{'b', 's', 'h'}:      "bak",
	{'b', 't', 'i'}:      "btb",
	{'c', 'h', 'g'}:      "sgw",
	{'c', 'h', 'h'}:      "hne",
	{'c', 'h', 'i'}:      "nya",
	{'c', 'h', 'k'}:      "ckt",
	{'c', 'h', 'k', '0'}: "chk",
	{'c', 'h', 'u'}:      "chv",
	{'c', 'h', 'y'}:      "chy",
	{'c', 'm', 'r'}:      "swb",
	{'c', 'r', 'r'}:      "crx",
	{'c', 'r', 't'}:      "crh",
	{'c', 's', 'l'}:      "chu",
	{'c', 's', 'y'}:      "ces",
	{'d', 'c', 'r'}:      "cwd",
	{'d', 'g', 'r'}:      "doi",
	{'d', 'j', 'r'}:      "dje",
	{'d', 'j', 'r', '0'}: "djr",
	{'d', 'n', 'g'}:      "ada",
	{'d', 'n', 'k'}:      "din",
	{'d', 'r', 'i'}:      "prs",
	{'d', 'u', 'n'}:      "dng",
	{'d', 'z', 'n'}:      "dzo",
	{'e', 'b', 'i'}:      "igb",
	{'e', 'c', 'r'}:      "crj",
	{'e', 'd', 'o'}:      "bin",
	{'e', 'r', 'z'}:      "myv",
	{'e', 's', 'p'}:      "spa",
	{'e', 't', 'i'}:      "est",
	{'e', 'u', 'q'}:      "eus",
	{'e', 'v', 'k'}:      "evn",
	{'e', 'v', 'n'}:      "eve",
	{'f', 'a', 'n'}:      "acf",
	{'f', 'a', 'n', '0'}: "fan",
	{'f', 'a', 'r'}:      "fas",
	{'f', 'j', 'i'}:      "fij",
	{'f', 'l', 'e'}:      "vls",
	{'f', 'n', 'e'}:      "enf",
	{'f', 'o', 's'}:      "fao",
	{'f', 'r', 'i'}:      "fry",
	{'f', 'r', 'l'}:      "fur",
	{'f', 'r', 'p'}:      "frp",
	{'f', 't', 'a'}:      "fuf",
	{'g', 'a', 'd'}:      "gaa",
	{'g', 'a', 'e'}:      "gla",
	{'g', 'a', 'l'}:      "glg",
	{'g', 'a', 'w'}:      "gbm",
	{'g', 'i', 'l'}:      "niv",
	{'g', 'i', 'l', '0'}: "gil",
	{'g', 'm', 'z'}:      "guk",
	{'g', 'r', 'n'}:      "kal",
	{'g', 'r', 'o'}:      "grt",
	{'g', 'u', 'a'}:      "grn",
	{'h', 'a', 'i'}:      "hat",
	{'h', 'a', 'l'}:      "flm",
	{'h', 'a', 'r'}:      "hoj",
	{'h', 'b', 'n'}:      "amf",
	{'h', 'm', 'a'}:      "mrj",
	{'h', 'n', 'd'}:      "hno",
	{'h', 'o'}:           "hoc",
	{'h', 'r', 'i'}:      "har",
	{'h', 'y', 'e', '0'}: "hye",
	{'i', 'j', 'o'}:      "ijc",
	{'i', 'n', 'g'}:      "inh",
	{'i', 'n', 'u'}:      "iku",
	{'i', 'r', 'i'}:      "gle",
	{'i', 'r', 't'}:      "gle",
	{'i', 's', 'm'}:      "smn",
	{'i', 'w', 'r'}:      "heb",
	{'j', 'a', 'n'}:      "jpn",
	{'j', 'i', 'i'}:      "yid",
	{'j', 'u', 'd'}:      "lad",
	{'j', 'u', 'l'}:      "dyu",
	{'k', 'a', 'b'}:      "kbd",
	{'k', 'a', 'b', '0'}: "kab",
	{'k', 'a', 'c'}:      "kfr",
	{'k', 'a', 'l'}:      "kln",
	{'k', 'a', 'r'}:      "krc",
	{'k', 'e', 'b'}:      "ktb",
	{'k', 'g', 'e'}:      "kat",
	{'k', 'h', 'a'}:      "kjh",
	{'k', 'h', 'k'}:      "kca",
	{'k', 'h', 's'}:      "kca",
	{'k', 'h', 'v'}:      "kca",
	{'k', 'i', 's'}:      "kqs",
	{'k', 'k', 'n'}:      "kex",
	{'k', 'l', 'm'}:      "xal",
	{'k', 'm', 'b'}:      "kam",
	{'k', 'm', 'n'}:      "kfy",
	{'k', 'm', 'o'}:      "kmw",
	{'k', 'm', 's'}:      "kxc",
	{'k', 'n', 'r'}:      "kau",
	{'k', 'o', 'd'}:      "kfa",
	{'k', 'o', 'h'}:      "okm",
	{'k', 'o', 'n'}:      "ktu",
	{'k', 'o', 'n', '0'}: "kon",
	{'k', 'o', 'p'}:      "koi",
	{'k', 'o', 'z'}:      "kpv",
	{'k', 'p', 'l'}:      "kpe",
	{'k', 'r', 'k'}:      "kaa",
	{'k', 'r', 'm'}:      "kdr",
	{'k', 'r', 'n'}:      "kar",
	{'k', 'r', 't'}:      "kqy",
	{'k', 's', 'h'}:      "kas",
	{'k', 's', 'h', '0'}: "ksh",
	{'k', 's', 'i'}:      "kha",
	{'k', 's', 'm'}:      "sjd",
	{'k', 'u', 'i'}:      "kxu",
	{'k', 'u', 'l'}:      "kfx",
	{'k', 'u', 'u'}:      "kru",
	{'k', 'u', 'y'}:      "kdt",
	{'k', 'y', 'k'}:      "kpy",
	{'l', 'a', 'd'}:      "lld",
	{'l', 'a', 'h'}:      "bfu",
	{'l', 'a', 'k'}:      "lbe",
	{'l', 'a', 'm'}:      "lmn",
	{'l', 'a', 'z'}:      "lzz",
	{'l', 'c', 'r'}:      "crm",
	{'l', 'd', 'k'}:      "lbj",
	{'l', 'm', 'a'}:      "mhr",
	{'l', 'm', 'b'}:      "lif",
	{'l', 'm', 'w'}:      "ngl",
	{'l', 's', 'b'}:      "dsb",
	{'l', 's', 'm'}:      "smj",
	{'l', 't', 'h'}:      "lit",
	{'l', 'u', 'h'}:      "luy",
	{'l', 'v', 'i'}:      "lav",
	{'m', 'a', 'j'}:      "mpe",
	{'m', 'a', 'k'}:      "vmw",
	{'m', 'a', 'n'}:      "mns",
	{'m', 'a', 'p'}:      "arn",
	{'m', 'a', 'w'}:      "mwr",
	{'m', 'b', 'n'}:      "kmb",
	{'m', 'c', 'h'}:      "mnc",
	{'m', 'c', 'r'}:      "crm",
	{'m', 'd', 'e'}:      "men",
	{'m', 'e', 'n'}:      "mym",
	{'m', 'i', 'z'}:      "lus",
	{'m', 'k', 'r'}:      "mak",
	{'m', 'l', 'e'}:      "mdy",
	{'m', 'l', 'n'}:      "mlq",
	{'m', 'l', 'r'}:      "mal",
	{'m', 'l', 'y'}:      "msa",
	{'m', 'n', 'd'}:      "mnk",
	{'m', 'n', 'g'}:      "mon",
	{'m', 'n', 'k'}:      "man",
	{'m', 'n', 'x'}:      "glv",
	{'m', 'o', 'k'}:      "mdf",
	{'m', 'o', 'n'}:      "mnw",
	{'m', 't', 'h'}:      "mai",
	{'m', 't', 's'}:      "mlt",
	{'m', 'u', 'n'}:      "unr",
	{'n', 'a', 'n'}:      "gld",
	{'n', 'a', 's'}:      "nsk",
	{'n', 'c', 'r'}:      "csw",
	{'n', 'd', 'g'}:      "ndo",
	{'n', 'h', 'c'}:      "csw",
	{'n', 'i', 's'}:      "dap",
	{'n', 'k', 'l'}:      "nyn",
	{'n', 'k', 'o'}:      "nqo",
	{'n', 'o', 'r'}:      "nob",
	{'n', 's', 'm'}:      "sme",
	{'n', 't', 'a'}:      "nod",
	{'n', 't', 'o'}:      "epo",
	{'n', 'y', 'n'}:      "nno",
	{'o', 'c', 'r'}:      "ojs",
	{'o', 'j', 'b'}:      "oji",
	{'o', 'r', 'o'}:      "orm",
	{'p', 'a', 'a'}:      "sam",
	{'p', 'a', 'l'}:      "pli",
	{'p', 'a', 'p'}:      "plp",
	{'p', 'a', 'p', '0'}: "pap",
	{'p', 'a', 's'}:      "pus",
	{'p', 'g', 'r'}:      "ell",
	{'p', 'i', 'l'}:      "fil",
	{'p', 'l', 'g'}:      "pce",
	{'p', 'l', 'k'}:      "pol",
	{'p', 't', 'g'}:      "por",
	{'q', 'i', 'n'}:      "bgr",
	{'r', 'b', 'u'}:      "bxr",
	{'r', 'c', 'r'}:      "atj",
	{'r', 'm', 's'}:      "roh",
	{'r', 'o', 'm'}:      "ron",
	{'r', 'o', 'y'}:      "rom",
	{'r', 's', 'y'}:      "rue",
	{'r', 'u', 'a'}:      "kin",
	{'s', 'a', 'd'}:      "sck",
	{'s', 'a', 'y'}:      "chp",
	{'s', 'e', 'k'}:      "xan",
	{'s', 'e', 'l'}:      "sel",
	{'s', 'g', 'o'}:      "sag",
	{'s', 'g', 's'}:      "sgs",
	{'s', 'i', 'b'}:      "sjo",
	{'s', 'i', 'g'}:      "xst",
	{'s', 'k', 's'}:      "sms",
	{'s', 'k', 'y'}:      "slk",
	{'s', 'l', 'a'}:      "scs",
	{'s', 'm', 'l'}:      "som",
	{'s', 'n', 'a'}:      "seh",
	{'s', 'n', 'a', '0'}: "sna",
	{'s', 'n', 'h'}:      "sin",
	{'s', 'o', 'g'}:      "gru",
	{'s', 'r', 'b'}:      "srp",
	{'s', 's', 'l'}:      "xsl",
	{'s', 's', 'm'}:      "sma",
	{'s', 'u', 'r'}:      "suq",
	{'s', 'v', 'e'}:      "swe",
	{'s', 'w', 'a'}:      "aii",
	{'s', 'w', 'k'}:      "swa",
	{'s', 'w', 'z'}:      "ssw",
	{'s', 'x', 't'}:      "ngo",
	{'t', 'a', 'j'}:      "tgk",
	{'t', 'c', 'r'}:      "cwd",
	{'t', 'g', 'n'}:      "ton",
	{'t', 'g', 'r'}:      "tig",
	{'t', 'g', 'y'}:      "tir",
	{'t', 'h', 't'}:      "tah",
	{'t', 'i', 'b'}:      "bod",
	{'t', 'k', 'm'}:      "tuk",
	{'t', 'm', 'n'}:      "tem",
	{'t', 'n', 'a'}:      "tsn",
	{'t', 'n', 'e'}:      "enh",
	{'t', 'n', 'g'}:      "toi",
	{'t', 'o', 'd'}:      "xal",
	{'t', 'o', 'd', '0'}: "tod",
	{'t', 'r', 'k'}:      "tur",
	{'t', 's', 'g'}:      "tso",
	{'t', 'u', 'a'}:      "tru",
	{'t', 'u', 'l'}:      "tcy",
	{'t', 'u', 'v'}:      "tyv",
	{'t', 'w', 'i'}:      "aka",
	{'u', 's', 'b'}:      "hsb",
	{'u', 'y', 'g'}:      "uig",
	{'v', 'i', 't'}:      "vie",
	{'v', 'r', 'o'}:      "vro",
	{'w', 'a'}:           "wbm",
	{'w', 'a', 'g'}:      "wbr",
	{'w', 'c', 'r'}:      "crk",
	{'w', 'e', 'l'}:      "cym",
	{'w', 'l', 'f'}:      "wol",
	{'x', 'b', 'd'}:      "khb",
	{'x', 'h', 's'}:      "xho",
	{'y', 'a', 'k'}:      "sah",
	{'y', 'b', 'a'}:      "yor",
	{'y', 'c', 'r'}:      "cre",
	{'y', 'i', 'm'}:      "iii",
	{'z', 'h', 'h'}:      "zho",
	{'z', 'h', 'p'}:      "zho",
	{'z', 'h', 's'}:      "zho",
	{'z', 'h', 't'}:      "zho",
	{'z', 'n', 'd'}:      "zne",
}

func canBreakTextPango(t []rune) pr.MaybeBool {
	if len(t) < 2 {
		return nil
	}
	logs := getLogAttrs(t)
	for _, l := range logs[1 : len(logs)-1] {
		if l.IsLineBreak() {
			return pr.True
		}
	}
	return pr.False
}
