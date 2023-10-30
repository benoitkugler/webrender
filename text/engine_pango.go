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
	fc "github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/textprocessing/pango"
	"github.com/benoitkugler/textprocessing/pango/fcfonts"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"
)

// FontConfiguration holds information about the
// available fonts on the system.
// It is used for text layout at various steps of the process.
type FontConfigurationPango struct {
	fontmap *fcfonts.FontMap

	userFonts    map[fonts.FaceID]fonts.Face
	fontsContent map[string][]byte // to be embedded in the target
}

// NewFontConfigurationPango uses a fontconfig database to create a new
// font configuration
func NewFontConfigurationPango(fontmap *fcfonts.FontMap) *FontConfigurationPango {
	out := &FontConfigurationPango{
		fontmap:      fontmap,
		userFonts:    make(map[fonts.FaceID]fonts.Face),
		fontsContent: make(map[string][]byte),
	}
	out.fontmap.SetFaceLoader(out)
	return out
}

func (f *FontConfigurationPango) LoadFace(key fonts.FaceID, format fc.FontFormat) (fonts.Face, error) {
	if face, has := f.userFonts[key]; has {
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
		key := fonts.FaceID{
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
	fontconfigStyle, ok := FONTCONFIG_STYLE[ruleDescriptors.FontStyle]
	if !ok {
		fontconfigStyle = "roman"
	}
	fontconfigWeight, ok := FONTCONFIG_WEIGHT[ruleDescriptors.FontWeight]
	if !ok {
		fontconfigWeight = "regular"
	}
	fontconfigStretch, ok := FONTCONFIG_STRETCH[ruleDescriptors.FontStretch]
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
	FONTCONFIG_WEIGHT = map[pr.IntString]string{
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
	FONTCONFIG_STYLE = map[pr.String]string{
		"normal":  "roman",
		"italic":  "italic",
		"oblique": "oblique",
	}
	FONTCONFIG_STRETCH = map[pr.String]string{
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
