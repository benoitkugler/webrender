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

	Context TextLayoutContext // will be a *LayoutContext; to avoid circular dependency

	Layout pango.Layout

	justificationSpacing pr.Fl
}

func newTextLayout(context TextLayoutContext, style *TextStyle, justificationSpacing pr.Fl, maxWidth pr.MaybeFloat) *TextLayoutPango {
	var layout TextLayoutPango

	layout.justificationSpacing = justificationSpacing
	layout.setup(context, style)
	layout.MaxWidth = maxWidth

	return &layout
}

// Text returns a readonly slice of the text used in the layout.
func (p *TextLayoutPango) Text() []rune { return p.Layout.Text }

func (p *TextLayoutPango) Metrics() *LineMetrics { return p.metrics }

func (p *TextLayoutPango) setup(context TextLayoutContext, style *TextStyle) {
	p.Context = context
	p.Style = style
	fontmap := context.Fonts().(*FontConfigurationPango).fontmap
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
		layout := newTextLayout(p.Context, p.Style, p.justificationSpacing, nil)
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
