package text

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
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

// textWrap returns true if the "white-space" property allows wrapping
func (ts *TextStyle) textWrap() bool {
	ws := ts.WhiteSpace
	return ws == WNormal || ws == WPreWrap || ws == WPreLine
}

func (ts *TextStyle) spaceCollapse() bool {
	ws := ts.WhiteSpace
	return ws == WNormal || ws == WNowrap || ws == WPreLine
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
	HyphenateLimitChars pr.Limits
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
	out.FontDescription.Weight = newFontWeight(style.GetFontWeight())
	out.FontDescription.Stretch = newFontStretch(style.GetFontStretch())
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

func (ft Feature) String() string {
	return fmt.Sprintf("'%s'=%d", ft.Tag[:], ft.Value)
}

type featureSet map[[4]byte]int

func newFeatureSet(fs []Feature) featureSet {
	out := make(featureSet)
	for _, f := range fs {
		out[f.Tag] = f.Value
	}
	return out
}

// other is applied on top of fs
func (fs featureSet) merge(other []Feature) {
	for _, f := range other {
		fs[f.Tag] = f.Value
	}
}

func (fs featureSet) list() []Feature {
	out := make([]Feature, 0, len(fs))
	for k, v := range fs {
		out = append(out, Feature{k, v})
	}
	sort.Slice(out, func(i, j int) bool { return bytes.Compare(out[i].Tag[:], out[j].Tag[:]) == -1 })
	return out
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

func getFontFaceFeatures(ruleDescriptors validation.FontFaceDescriptors) []Feature {
	props := pr.Properties{}
	// avoid nil values
	props.SetFontKerning("")
	props.SetFontVariantLigatures(pr.SStrings{})
	props.SetFontVariantPosition("")
	props.SetFontVariantCaps("")
	props.SetFontVariantNumeric(pr.SStrings{})
	props.SetFontVariantAlternates("")
	props.SetFontVariantEastAsian(pr.SStrings{})
	props.SetFontFeatureSettings(pr.SIntStrings{})
	for _, rules := range ruleDescriptors.FontVariant {
		if rules.Property.SpecialProperty != nil {
			continue
		}
		if cascaded := rules.Property.ToCascaded(); cascaded.Default == 0 {
			props[rules.Name.KnownProp] = cascaded.ToCSS()
		}
	}
	if !ruleDescriptors.FontFeatureSettings.IsNone() {
		props.SetFontFeatureSettings(ruleDescriptors.FontFeatureSettings)
	}

	return getFontFeatures(props)
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

func newFontWeight(weight pr.IntString) uint16 {
	if weight.String == "normal" {
		weight.Int = 400
	} else if weight.String == "bold" {
		weight.Int = 700
	}
	return uint16(weight.Int)
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

// Language system tags
// From https://docs.microsoft.com/typography/opentype/spec/languagetags
var lstToISO = map[fontLanguageOverride]string{
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
