package text

import (
	"fmt"
	"strings"

	"github.com/benoitkugler/textlayout/language"
	"github.com/benoitkugler/textlayout/pango"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/html/layout/text/hyphen"
)

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

func PangoUnitsFromFloat(v pr.Fl) int32 { return int32(v*pango.Scale + 0.5) }

func PangoUnitsToFloat(v pango.Unit) pr.Fl { return pr.Fl(v) / pango.Scale }

type TextLayoutContext interface {
	Fontmap() pango.FontMap
	HyphenCache() map[HyphenDictKey]hyphen.Hyphener
	StrutLayoutsCache() map[StrutLayoutKey][2]pr.Float
}

// TextLayout wraps a pango.Layout object
type TextLayout struct {
	Style   pr.StyleAccessor
	Metrics *pango.FontMetrics // optional

	MaxWidth pr.MaybeFloat

	Context TextLayoutContext // will be a *LayoutContext; to avoid circular dependency

	Layout pango.Layout

	JustificationSpacing pr.Fl
	FirstLineDirection   pango.Direction
}

func NewTextLayout(context TextLayoutContext, fontSize pr.Fl, style pr.StyleAccessor, justificationSpacing pr.Fl, maxWidth pr.MaybeFloat) *TextLayout {
	var layout TextLayout

	layout.JustificationSpacing = justificationSpacing
	layout.setup(context, fontSize, style)
	layout.MaxWidth = maxWidth

	return &layout
}

func (p *TextLayout) setup(context TextLayoutContext, fontSize pr.Fl, style pr.StyleAccessor) {
	p.Context = context
	p.Style = style
	p.FirstLineDirection = 0
	fontmap := context.Fontmap()
	pc := pango.NewContext(fontmap)
	pc.SetRoundGlyphPositions(false)

	var lang pango.Language
	if flo := style.GetFontLanguageOverride(); flo != "normal" {
		lang = lstToISO[strings.ToLower(string(flo))]
	} else if lg := style.GetLang().String; lg != "" {
		lang = language.NewLanguage(lg)
	} else {
		lang = pango.DefaultLanguage()
	}
	pc.SetLanguage(lang)

	fontDesc := pango.NewFontDescription()
	fontDesc.SetFamily(strings.Join(style.GetFontFamily(), ","))

	sty, _ := pango.StyleMap.FromString(string(style.GetFontStyle()))
	fontDesc.SetStyle(pango.Style(sty))

	str, _ := pango.StretchMap.FromString(string(style.GetFontStretch()))
	fontDesc.SetStretch(pango.Stretch(str))

	fontDesc.SetWeight(pango.Weight(style.GetFontWeight().Int))

	fontDesc.SetAbsoluteSize(PangoUnitsFromFloat(fontSize))

	if !style.GetTextDecorationLine().IsNone() {
		metrics := pc.GetMetrics(&fontDesc, lang)
		p.Metrics = &metrics
	} else {
		p.Metrics = nil
	}

	p.Layout = *pango.NewLayout(pc)
	p.Layout.SetFontDescription(&fontDesc)

	features := getFontFeatures(style)
	if len(features) != 0 {
		var chunks []string
		for k, v := range features {
			chunks = append(chunks, fmt.Sprintf("%s=%d", k, v))
		}
		featuresString := strings.Join(chunks, ",")
		attr := pango.NewAttrFontFeatures(featuresString)
		p.Layout.SetAttributes(pango.AttrList{attr})
	}
}

func (p *TextLayout) SetText(text string) { p.setText(text, false) }

// ApplyJustification re-layout the text, applying justification.
func (p *TextLayout) ApplyJustification() {
	p.Layout.SetWidth(-1)
	p.setText(string(p.Layout.Text), true)
}

func (p *TextLayout) setText(text string, justify bool) {
	if index := strings.IndexByte(text, '\n'); index != -1 && len(text) >= index+2 {
		// Keep only the first line plus one character, we don't need more
		text = text[:index+2]
	}

	p.Layout.SetText(text)

	wordSpacing := pr.Fl(p.Style.GetWordSpacing().Value)
	if justify {
		// Justification is needed when drawing text but is useless during
		// layout. Ignore it before layout is reactivated before the drawing
		// step.
		wordSpacing += p.JustificationSpacing
	}

	var letterSpacing pr.Fl
	if ls := p.Style.GetLetterSpacing(); ls.String != "normal" {
		letterSpacing = pr.Fl(ls.Value)
	}

	wordBreaking := p.Style.GetOverflowWrap() == "anywhere" || p.Style.GetOverflowWrap() == "break-word"

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
			for position, c := range textRunes {
				if c == ' ' {
					addAttr(position, position+1, spaceSpacingInt)
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

func (p *TextLayout) setTabs() {
	tabSize := p.Style.GetTabSize()
	width := int(tabSize.Value)
	if tabSize.Unit == 0 { // no unit, means a multiple of the advance width of the space character
		layout := NewTextLayout(p.Context, pr.Fl(p.Style.GetFontSize().Value), p.Style, p.JustificationSpacing, nil)
		layout.SetText(strings.Repeat(" ", width))
		line, _ := layout.GetFirstLine()
		widthTmp, _ := lineSize(line, p.Style.GetLetterSpacing())
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
func (p *TextLayout) GetFirstLine() (*pango.LayoutLine, int) {
	firstLine := p.Layout.GetLine(0)
	secondLine := p.Layout.GetLine(1)
	index := -1
	if secondLine != nil {
		index = secondLine.StartIndex
	}

	p.FirstLineDirection = firstLine.ResolvedDir

	return firstLine, index
}

// lineSize gets the logical width and height of the given `line`.
// `style` is used to add letter spacing (if needed).
func lineSize(line *pango.LayoutLine, letterSpacing pr.Value) (pr.Fl, pr.Fl) {
	var logicalExtents pango.Rectangle
	line.GetExtents(nil, &logicalExtents)
	width := PangoUnitsToFloat(logicalExtents.Width)
	height := PangoUnitsToFloat(logicalExtents.Height)
	if letterSpacing.String != "normal" {
		width += pr.Fl(letterSpacing.Value)
	}
	return width, height
}

func defaultFontFeature(f string) string {
	if f == "" {
		return "normal"
	}
	return f
}
