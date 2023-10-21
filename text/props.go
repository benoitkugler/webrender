package text

import (
	"strings"

	"github.com/benoitkugler/textlayout/language"
	pr "github.com/benoitkugler/webrender/css/properties"
)

// TextStyle exposes the subset of a [pr.Style]
// required to layout text.
type TextStyle struct {
	FontFamily            []string
	FontStyle             FontStyle
	FontStretch           FontStretch
	FontWeight            int
	FontSize              pr.Fl
	FontVariationSettings []Variation // empty for 'normal'

	FontLanguageOverride FontLanguageOverride
	Lang                 string

	TextDecorationLine pr.Decorations

	WhiteSpace    Whitespace
	LetterSpacing pr.Fl // 0 for 'normal'

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
}

func NewTextStyle(style pr.StyleAccessor) *TextStyle {
	var out TextStyle

	out.FontFamily = style.GetFontFamily()
	out.FontStyle = newFontStyle(style.GetFontStyle())
	out.FontStretch = newFontStretch(style.GetFontStretch())
	out.FontWeight = style.GetFontWeight().Int
	out.FontSize = pr.Fl(style.GetFontSize().Value)
	out.FontVariationSettings = newFontVariationSettings(style.GetFontVariationSettings())

	out.FontLanguageOverride = newFontLanguageOverrride(style.GetFontLanguageOverride())
	out.Lang = style.GetLang().String

	out.TextDecorationLine = style.GetTextDecorationLine()

	out.WhiteSpace = newWhiteSpace(style.GetWhiteSpace())
	if ls := style.GetLetterSpacing(); ls.String != "normal" {
		out.LetterSpacing = pr.Fl(ls.Value)
	}

	out.FontFeatures = getFontFeatures(style)

	return &out
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

// FontLanguageOverride is either 'normal' (the zero value)
// or a 4 byte tag, normalized to lower case
type FontLanguageOverride [4]byte

func newFontLanguageOverrride(flo pr.String) FontLanguageOverride {
	if flo == "normal" {
		return [4]byte{}
	}

	var out [4]byte
	copy(out[:], strings.ToLower(string(flo)))
	return out
}

// Language system tags
// From https://docs.microsoft.com/typography/opentype/spec/languagetags
var lstToISO = map[string]language.Language{
	"aba":  "abq",
	"afk":  "afr",
	"afr":  "aar",
	"agw":  "ahg",
	"als":  "gsw",
	"alt":  "atv",
	"ari":  "aiw",
	"ark":  "mhv",
	"ath":  "apk",
	"avr":  "ava",
	"bad":  "bfq",
	"bad0": "bad",
	"bag":  "bfy",
	"bal":  "krc",
	"bau":  "bci",
	"bch":  "bcq",
	"bgr":  "bul",
	"bil":  "byn",
	"bkf":  "bla",
	"bli":  "bal",
	"bln":  "bjt",
	"blt":  "bft",
	"bmb":  "bam",
	"bri":  "bra",
	"brm":  "mya",
	"bsh":  "bak",
	"bti":  "btb",
	"chg":  "sgw",
	"chh":  "hne",
	"chi":  "nya",
	"chk":  "ckt",
	"chk0": "chk",
	"chu":  "chv",
	"chy":  "chy",
	"cmr":  "swb",
	"crr":  "crx",
	"crt":  "crh",
	"csl":  "chu",
	"csy":  "ces",
	"dcr":  "cwd",
	"dgr":  "doi",
	"djr":  "dje",
	"djr0": "djr",
	"dng":  "ada",
	"dnk":  "din",
	"dri":  "prs",
	"dun":  "dng",
	"dzn":  "dzo",
	"ebi":  "igb",
	"ecr":  "crj",
	"edo":  "bin",
	"erz":  "myv",
	"esp":  "spa",
	"eti":  "est",
	"euq":  "eus",
	"evk":  "evn",
	"evn":  "eve",
	"fan":  "acf",
	"fan0": "fan",
	"far":  "fas",
	"fji":  "fij",
	"fle":  "vls",
	"fne":  "enf",
	"fos":  "fao",
	"fri":  "fry",
	"frl":  "fur",
	"frp":  "frp",
	"fta":  "fuf",
	"gad":  "gaa",
	"gae":  "gla",
	"gal":  "glg",
	"gaw":  "gbm",
	"gil":  "niv",
	"gil0": "gil",
	"gmz":  "guk",
	"grn":  "kal",
	"gro":  "grt",
	"gua":  "grn",
	"hai":  "hat",
	"hal":  "flm",
	"har":  "hoj",
	"hbn":  "amf",
	"hma":  "mrj",
	"hnd":  "hno",
	"ho":   "hoc",
	"hri":  "har",
	"hye0": "hye",
	"ijo":  "ijc",
	"ing":  "inh",
	"inu":  "iku",
	"iri":  "gle",
	"irt":  "gle",
	"ism":  "smn",
	"iwr":  "heb",
	"jan":  "jpn",
	"jii":  "yid",
	"jud":  "lad",
	"jul":  "dyu",
	"kab":  "kbd",
	"kab0": "kab",
	"kac":  "kfr",
	"kal":  "kln",
	"kar":  "krc",
	"keb":  "ktb",
	"kge":  "kat",
	"kha":  "kjh",
	"khk":  "kca",
	"khs":  "kca",
	"khv":  "kca",
	"kis":  "kqs",
	"kkn":  "kex",
	"klm":  "xal",
	"kmb":  "kam",
	"kmn":  "kfy",
	"kmo":  "kmw",
	"kms":  "kxc",
	"knr":  "kau",
	"kod":  "kfa",
	"koh":  "okm",
	"kon":  "ktu",
	"kon0": "kon",
	"kop":  "koi",
	"koz":  "kpv",
	"kpl":  "kpe",
	"krk":  "kaa",
	"krm":  "kdr",
	"krn":  "kar",
	"krt":  "kqy",
	"ksh":  "kas",
	"ksh0": "ksh",
	"ksi":  "kha",
	"ksm":  "sjd",
	"kui":  "kxu",
	"kul":  "kfx",
	"kuu":  "kru",
	"kuy":  "kdt",
	"kyk":  "kpy",
	"lad":  "lld",
	"lah":  "bfu",
	"lak":  "lbe",
	"lam":  "lmn",
	"laz":  "lzz",
	"lcr":  "crm",
	"ldk":  "lbj",
	"lma":  "mhr",
	"lmb":  "lif",
	"lmw":  "ngl",
	"lsb":  "dsb",
	"lsm":  "smj",
	"lth":  "lit",
	"luh":  "luy",
	"lvi":  "lav",
	"maj":  "mpe",
	"mak":  "vmw",
	"man":  "mns",
	"map":  "arn",
	"maw":  "mwr",
	"mbn":  "kmb",
	"mch":  "mnc",
	"mcr":  "crm",
	"mde":  "men",
	"men":  "mym",
	"miz":  "lus",
	"mkr":  "mak",
	"mle":  "mdy",
	"mln":  "mlq",
	"mlr":  "mal",
	"mly":  "msa",
	"mnd":  "mnk",
	"mng":  "mon",
	"mnk":  "man",
	"mnx":  "glv",
	"mok":  "mdf",
	"mon":  "mnw",
	"mth":  "mai",
	"mts":  "mlt",
	"mun":  "unr",
	"nan":  "gld",
	"nas":  "nsk",
	"ncr":  "csw",
	"ndg":  "ndo",
	"nhc":  "csw",
	"nis":  "dap",
	"nkl":  "nyn",
	"nko":  "nqo",
	"nor":  "nob",
	"nsm":  "sme",
	"nta":  "nod",
	"nto":  "epo",
	"nyn":  "nno",
	"ocr":  "ojs",
	"ojb":  "oji",
	"oro":  "orm",
	"paa":  "sam",
	"pal":  "pli",
	"pap":  "plp",
	"pap0": "pap",
	"pas":  "pus",
	"pgr":  "ell",
	"pil":  "fil",
	"plg":  "pce",
	"plk":  "pol",
	"ptg":  "por",
	"qin":  "bgr",
	"rbu":  "bxr",
	"rcr":  "atj",
	"rms":  "roh",
	"rom":  "ron",
	"roy":  "rom",
	"rsy":  "rue",
	"rua":  "kin",
	"sad":  "sck",
	"say":  "chp",
	"sek":  "xan",
	"sel":  "sel",
	"sgo":  "sag",
	"sgs":  "sgs",
	"sib":  "sjo",
	"sig":  "xst",
	"sks":  "sms",
	"sky":  "slk",
	"sla":  "scs",
	"sml":  "som",
	"sna":  "seh",
	"sna0": "sna",
	"snh":  "sin",
	"sog":  "gru",
	"srb":  "srp",
	"ssl":  "xsl",
	"ssm":  "sma",
	"sur":  "suq",
	"sve":  "swe",
	"swa":  "aii",
	"swk":  "swa",
	"swz":  "ssw",
	"sxt":  "ngo",
	"taj":  "tgk",
	"tcr":  "cwd",
	"tgn":  "ton",
	"tgr":  "tig",
	"tgy":  "tir",
	"tht":  "tah",
	"tib":  "bod",
	"tkm":  "tuk",
	"tmn":  "tem",
	"tna":  "tsn",
	"tne":  "enh",
	"tng":  "toi",
	"tod":  "xal",
	"tod0": "tod",
	"trk":  "tur",
	"tsg":  "tso",
	"tua":  "tru",
	"tul":  "tcy",
	"tuv":  "tyv",
	"twi":  "aka",
	"usb":  "hsb",
	"uyg":  "uig",
	"vit":  "vie",
	"vro":  "vro",
	"wa":   "wbm",
	"wag":  "wbr",
	"wcr":  "crk",
	"wel":  "cym",
	"wlf":  "wol",
	"xbd":  "khb",
	"xhs":  "xho",
	"yak":  "sah",
	"yba":  "yor",
	"ycr":  "cre",
	"yim":  "iii",
	"zhh":  "zho",
	"zhp":  "zho",
	"zhs":  "zho",
	"zht":  "zho",
	"znd":  "zne",
}
