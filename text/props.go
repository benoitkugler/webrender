package text

import (
	"encoding/binary"
	"math"
	"strings"

	pr "github.com/benoitkugler/webrender/css/properties"
)

type LineMetrics struct {
	// Distance from the baseline to the logical top of a line of text.
	// (The logical top may be above or below the top of the
	// actual drawn ink. It is necessary to lay out the text to figure
	// where the ink will be.)
	Ascent pr.Fl

	// Distance above the baseline of the top of the underline.
	// Since most fonts have underline positions beneath the baseline, this value is typically negative.
	UnderlinePosition pr.Fl

	// Suggested thickness to draw for the underline.
	UnderlineThickness pr.Fl

	// Distance above the baseline of the top of the strikethrough.
	StrikethroughPosition pr.Fl
	// Suggested thickness to draw for the strikethrough.
	StrikethroughThickness pr.Fl
}

// FontDescription stores the settings influencing
// font resolution and metrics.
type FontDescription struct {
	Family            []string
	Style             FontStyle
	Stretch           FontStretch
	Weight            uint16
	Size              pr.Fl
	VariationSettings []Variation // empty for 'normal'
}

func (fd FontDescription) hash(includeSize bool) []byte {
	var hash []byte
	for _, f := range fd.Family {
		hash = append(hash, f...)
	}
	hash = append(hash, byte(fd.Style), byte(fd.Stretch))
	hash = binary.BigEndian.AppendUint16(hash, fd.Weight)
	if includeSize {
		hash = binary.BigEndian.AppendUint32(hash, math.Float32bits(fd.Size))
	}
	for _, v := range fd.VariationSettings {
		hash = append(hash, v.Tag[:]...)
		hash = binary.BigEndian.AppendUint32(hash, math.Float32bits(v.Value))
	}
	return hash
}

// TextStyle exposes the subset of a [pr.Style]
// required to layout text.
type TextStyle struct {
	FontDescription

	TextDecorationLine pr.Decorations

	// FontFeatures stores the resolved value
	// for all the CSS properties related :
	// 	"font-kerning"
	// 	"font-variant-ligatures"
	// 	"font-variant-position"
	// 	"font-variant-caps"
	// 	"font-variant-numeric"
	// 	"font-variant-alternates"
	// 	"font-variant-east-asian"
	// 	"font-feature-settings"
	FontFeatures []Feature

	FontLanguageOverride fontLanguageOverride
	Lang                 string

	WhiteSpace   Whitespace
	OverflowWrap OverflowWrap
	WordBreak    WordBreak

	Hyphens             Hyphens
	HyphenateCharacter  string
	HyphenateLimitChars pr.Ints3
	HyphenateLimitZone  HyphenateZone

	WordSpacing   pr.Fl
	LetterSpacing pr.Fl // 0 for 'normal'
	TabSize       TabSize
}

// If ignoreSpacing is true, 'word-spacing' and 'letter-spacing' are
// not queried from [style]
func NewTextStyle(style pr.StyleAccessor, ignoreSpacing bool) *TextStyle {
	var out TextStyle

	out.FontDescription.Family = style.GetFontFamily()
	out.FontDescription.Style = newFontStyle(style.GetFontStyle())
	out.FontDescription.Stretch = newFontStretch(style.GetFontStretch())
	out.FontDescription.Weight = uint16(style.GetFontWeight().Int)
	out.FontDescription.Size = pr.Fl(style.GetFontSize().Value)
	out.FontDescription.VariationSettings = newFontVariationSettings(style.GetFontVariationSettings())

	out.FontLanguageOverride = newFontLanguageOverrride(style.GetFontLanguageOverride())
	out.Lang = style.GetLang().String

	out.TextDecorationLine = style.GetTextDecorationLine()

	out.WhiteSpace = newWhiteSpace(style.GetWhiteSpace())
	out.OverflowWrap = newOverflowWrap(style.GetOverflowWrap())
	out.WordBreak = newWordBreak(style.GetWordBreak())

	out.Hyphens = newHyphens(style.GetHyphens())
	out.HyphenateLimitChars = style.GetHyphenateLimitChars()
	out.HyphenateCharacter = string(style.GetHyphenateCharacter())
	out.HyphenateLimitZone = newHyphenateZone(style.GetHyphenateLimitZone())

	if !ignoreSpacing {
		out.WordSpacing = pr.Fl(style.GetWordSpacing().Value)
		if ls := style.GetLetterSpacing(); ls.String != "normal" {
			out.LetterSpacing = pr.Fl(ls.Value)
		}
	}

	out.TabSize = newTabSize(style.GetTabSize())

	out.FontFeatures = getFontFeatures(style)

	return &out
}

type TabSize struct {
	Width      int
	IsMultiple bool // true to use Width * <space character width>
}

func newTabSize(ts pr.Value) TabSize {
	return TabSize{
		Width:      int(ts.Value),
		IsMultiple: ts.Unit == 0,
	}
}

type Feature struct {
	Tag   [4]byte
	Value int
}

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

func defaultFontFeature(f string) string {
	if f == "" {
		return "normal"
	}
	return f
}

// Get the font features from the different properties in style.
// See https://www.w3.org/TR/css-fonts-3/#feature-precedence
// default value is "normal"
// pass nil for default ("normal") on fontFeatureSettings
func getFontFeatures(style pr.StyleAccessor) []Feature {
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

	if len(features) == 0 {
		return nil
	}

	out := make([]Feature, 0, len(features))
	for k, v := range features {
		var item Feature
		copy(item.Tag[:], k)
		item.Value = v
		out = append(out, item)
	}

	return out
}

type WordBreak uint8

const (
	WBNormal WordBreak = iota
	WBBreakAll
)

func newWordBreak(w pr.String) WordBreak {
	switch w {
	case "", "normal":
		return WBNormal
	case "break-all":
		return WBBreakAll
	default:
		return WBNormal
	}
}

type Whitespace uint8

const (
	WNormal Whitespace = iota
	WNowrap
	WPre
	WPreWrap
	WPreLine
	WBreakSpaces
)

func newWhiteSpace(w pr.String) Whitespace {
	switch w {
	case "", "normal":
		return WNormal
	case "nowrap":
		return WNowrap
	case "pre":
		return WPre
	case "pre-wrap":
		return WPreWrap
	case "pre-line":
		return WPreLine
	case "break-spaces":
		return WBreakSpaces
	default:
		return WNormal
	}
}

type OverflowWrap uint8

const (
	ONormal OverflowWrap = iota
	OAnywhere
	OBreakWord
)

func newOverflowWrap(w pr.String) OverflowWrap {
	switch w {
	case "", "normal":
		return ONormal
	case "anywhere":
		return OAnywhere
	case "break-word":
		return OBreakWord
	default:
		return ONormal
	}
}

type Hyphens uint8

const (
	HManual Hyphens = iota
	HNone
	HAuto
)

func newHyphens(h pr.String) Hyphens {
	switch h {
	case "", "manual":
		return HManual
	case "none":
		return HNone
	case "auto":
		return HAuto
	default:
		return HManual
	}
}

type HyphenateZone struct {
	Limit        pr.Fl
	IsPercentage bool
}

func newHyphenateZone(zone pr.Value) HyphenateZone {
	return HyphenateZone{
		Limit:        pr.Fl(zone.Value),
		IsPercentage: zone.Unit == pr.Perc,
	}
}

type FontStyle uint8

const (
	FSyNormal FontStyle = iota
	FSyOblique
	FSyItalic
)

func newFontStyle(style pr.String) FontStyle {
	switch strings.ToLower(string(style)) {
	case "", "roman", "normal":
		return FSyNormal
	case "oblique":
		return FSyOblique
	case "italic":
		return FSyItalic
	default:
		return FSyNormal
	}
}

type FontStretch uint8

const (
	FSeUltraCondensed FontStretch = iota // ultra condensed width
	FSeExtraCondensed                    // extra condensed width
	FSeCondensed                         // condensed width
	FSeSemiCondensed                     // semi condensed width
	FSeNormal                            // the normal width
	FSeSemiExpanded                      // semi expanded width
	FSeExpanded                          // expanded width
	FSeExtraExpanded                     // extra expanded width
	FSeUltraExpanded                     // ultra expanded width
)

func newFontStretch(stretch pr.String) FontStretch {
	switch strings.ToLower(string(stretch)) {
	case "", "normal":
		return FSeNormal
	case "ultra-condensed":
		return FSeUltraCondensed
	case "extra-condensed":
		return FSeExtraCondensed
	case "condensed":
		return FSeCondensed
	case "semi-condensed":
		return FSeSemiCondensed
	case "semi-expanded":
		return FSeSemiExpanded
	case "expanded":
		return FSeExpanded
	case "extra-expanded":
		return FSeExtraExpanded
	case "ultra-expanded":
		return FSeUltraExpanded
	default:
		return FSeNormal
	}
}

type Variation struct {
	Tag   [4]byte
	Value pr.Fl
}

func newFontVariationSettings(vs pr.SFloatStrings) []Variation {
	if vs.String == "normal" {
		return nil
	}
	out := make([]Variation, len(vs.Values))
	for i, v := range vs.Values {
		copy(out[i].Tag[:], v.String)
		out[i].Value = v.Float
	}
	return out
}

// fontLanguageOverride is either 'normal' (coded as the zero value)
// or a 4 byte tag, normalized to lower case
type fontLanguageOverride [4]byte

func newFontLanguageOverrride(flo pr.String) fontLanguageOverride {
	if flo == "normal" {
		return [4]byte{}
	}

	var out [4]byte
	copy(out[:], strings.ToLower(string(flo)))
	return out
}
