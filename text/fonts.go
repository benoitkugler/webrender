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
type FontConfiguration struct {
	Fontmap *fcfonts.FontMap

	userFonts    map[fonts.FaceID]fonts.Face
	fontsContent map[string][]byte // to be embedded in the target
}

// NewFontConfiguration uses a fontconfig database to create a new
// font configuration
func NewFontConfiguration(fontmap *fcfonts.FontMap) *FontConfiguration {
	out := &FontConfiguration{
		Fontmap:      fontmap,
		userFonts:    make(map[fonts.FaceID]fonts.Face),
		fontsContent: make(map[string][]byte),
	}
	out.Fontmap.SetFaceLoader(out)
	return out
}

func (f *FontConfiguration) LoadFace(key fonts.FaceID, format fc.FontFormat) (fonts.Face, error) {
	if face, has := f.userFonts[key]; has {
		return face, nil
	}
	return fcfonts.DefaultLoadFace(key, format)
}

// FontContent returns the content of the given face, which may be needed
// in the final output.
func (f *FontConfiguration) FontContent(faceID fonts.FaceID) []byte {
	// either the font is loaded at run time or is loaded from disk
	if content, has := f.fontsContent[faceID.File]; has {
		return content
	}

	b, err := os.ReadFile(faceID.File)
	if err != nil {
		logger.WarningLogger.Println(err)
	}
	// cache the result to avoid loading the same file over and over
	f.fontsContent[faceID.File] = b

	return b
}

func (f *FontConfiguration) AddFontFace(ruleDescriptors validation.FontFaceDescriptors, urlFetcher utils.UrlFetcher) string {
	if f.Fontmap == nil {
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

func (f *FontConfiguration) loadOneFont(url pr.NamedString, ruleDescriptors validation.FontFaceDescriptors, urlFetcher utils.UrlFetcher) (string, error) {
	config := f.Fontmap.Config

	if url.Name == "local" {
		fontName := url.String
		pattern := fc.NewPattern()
		config.Substitute(pattern, nil, fc.MatchResult)
		pattern.SubstituteDefault()
		pattern.AddString(fc.FULLNAME, fontName)
		pattern.AddString(fc.POSTSCRIPT_NAME, fontName)
		matchingPattern := f.Fontmap.Database.Match(pattern, config)

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
	for k, v := range getFontFeatures(features) {
		featuresString += fmt.Sprintf("<string>%s=%d</string>", k, v)
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

	f.Fontmap.Database = append(f.Fontmap.Database, fs...)
	f.Fontmap.SetConfig(config, f.Fontmap.Database)
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

var (
	ligatureKeys = map[string][]string{
		"common-ligatures":        {"liga", "clig"},
		"historical-ligatures":    {"hlig"},
		"discretionary-ligatures": {"dlig"},
		"contextual":              {"calt"},
	}
	capsKeys = map[string][]string{
		"small-caps":      {"smcp"},
		"all-small-caps":  {"c2sc", "smcp"},
		"petite-caps":     {"pcap"},
		"all-petite-caps": {"c2pc", "pcap"},
		"unicase":         {"unic"},
		"titling-caps":    {"titl"},
	}
	numericKeys = map[string]string{
		"lining-nums":        "lnum",
		"oldstyle-nums":      "onum",
		"proportional-nums":  "pnum",
		"tabular-nums":       "tnum",
		"diagonal-fractions": "frac",
		"stacked-fractions":  "afrc",
		"ordinal":            "ordn",
		"slashed-zero":       "zero",
	}
	eastAsianKeys = map[string]string{
		"jis78":              "jp78",
		"jis83":              "jp83",
		"jis90":              "jp90",
		"jis04":              "jp04",
		"simplified":         "smpl",
		"traditional":        "trad",
		"full-width":         "fwid",
		"proportional-width": "pwid",
		"ruby":               "ruby",
	}
)

// Get the font features from the different properties in style.
// See https://www.w3.org/TR/css-fonts-3/#feature-precedence
// default value is "normal"
// pass nil for default ("normal") on fontFeatureSettings
func getFontFeatures(style pr.StyleAccessor) map[string]int {
	fontKerning := defaultFontFeature(string(style.GetFontKerning()))
	fontVariantPosition := defaultFontFeature(string(style.GetFontVariantPosition()))
	fontVariantCaps := defaultFontFeature(string(style.GetFontVariantCaps()))
	fontVariantAlternates := defaultFontFeature(string(style.GetFontVariantAlternates()))

	features := map[string]int{}

	// Step 1: getting the default, we rely on Pango for this
	// Step 2: @font-face font-variant, done in fonts.addFontFace
	// Step 3: @font-face font-feature-settings, done in fonts.addFontFace

	// Step 4: font-variant && OpenType features

	if fontKerning != "auto" {
		features["kern"] = 0
		if fontKerning == "normal" {
			features["kern"] = 1
		}
	}

	fontVariantLigatures := style.GetFontVariantLigatures()
	if fontVariantLigatures.String == "none" {
		for _, keys := range ligatureKeys {
			for _, key := range keys {
				features[key] = 0
			}
		}
	} else if fontVariantLigatures.String != "normal" {
		for _, ligatureType := range fontVariantLigatures.Strings {
			value := 1
			if strings.HasPrefix(ligatureType, "no-") {
				value = 0
				ligatureType = ligatureType[3:]
			}
			for _, key := range ligatureKeys[ligatureType] {
				features[key] = value
			}
		}
	}

	if fontVariantPosition == "sub" {
		// https://www.w3.org/TR/css-fonts-3/#font-variant-position-prop
		features["subs"] = 1
	} else if fontVariantPosition == "super" {
		features["sups"] = 1
	}

	if fontVariantCaps != "normal" {
		// https://www.w3.org/TR/css-fonts-3/#font-variant-caps-prop
		for _, key := range capsKeys[fontVariantCaps] {
			features[key] = 1
		}
	}

	if fv := style.GetFontVariantNumeric(); fv.String != "normal" {
		for _, key := range fv.Strings {
			features[numericKeys[key]] = 1
		}
	}

	if fontVariantAlternates != "normal" {
		// See https://www.w3.org/TR/css-fonts-3/#font-variant-caps-prop
		if fontVariantAlternates == "historical-forms" {
			features["hist"] = 1
		}
	}

	if fv := style.GetFontVariantEastAsian(); fv.String != "normal" {
		for _, key := range fv.Strings {
			features[eastAsianKeys[key]] = 1
		}
	}

	// Step 5: incompatible non-OpenType features, already handled by Pango

	// Step 6: font-feature-settings
	for _, pair := range style.GetFontFeatureSettings().Values {
		features[pair.String] = pair.Int
	}

	return features
}

func getFontDescription(style pr.StyleAccessor) pango.FontDescription {
	fontDesc := pango.NewFontDescription()
	fontDesc.SetFamily(strings.Join(style.GetFontFamily(), ","))

	sty, _ := pango.StyleMap.FromString(string(style.GetFontStyle()))
	fontDesc.SetStyle(pango.Style(sty))

	str, _ := pango.StretchMap.FromString(string(style.GetFontStretch()))
	fontDesc.SetStretch(pango.Stretch(str))

	fontDesc.SetWeight(pango.Weight(style.GetFontWeight().Int))

	fontDesc.SetAbsoluteSize(PangoUnitsFromFloat(pr.Fl(style.GetFontSize().Value)))

	if vs := style.GetFontVariationSettings(); vs.String != "normal" {
		chunks := make([]string, len(vs.Values))
		for i, v := range vs.Values {
			chunks[i] = fmt.Sprintf("%s=%f", v.String, v.Float)
		}
		variations := strings.Join(chunks, ",")
		fontDesc.SetVariations(variations)
	}
	return fontDesc
}
