package text

import (
	"fmt"
	"strings"

	"github.com/benoitkugler/textlayout/language"
	"github.com/benoitkugler/textlayout/pango"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/html/layout/text/hyphen"
	"github.com/benoitkugler/webrender/utils"
)

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
		lang = utils.LST_TO_ISO[strings.ToLower(string(flo))]
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

	fontDesc.SetAbsoluteSize(utils.PangoUnitsFromFloat(fontSize))

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

	if text != "" && (wordSpacing != 0 || letterSpacing != 0) {
		letterSpacingInt := utils.PangoUnitsFromFloat(letterSpacing)
		spaceSpacingInt := utils.PangoUnitsFromFloat(wordSpacing) + letterSpacingInt
		attrList := p.Layout.Attributes

		addAttr := func(start, end int, spacing int32) {
			attr := pango.NewAttrLetterSpacing(spacing)
			attr.StartIndex, attr.EndIndex = start, end
			attrList.Change(attr)
		}

		textRunes := p.Layout.Text
		addAttr(0, len(textRunes), letterSpacingInt)
		for position, c := range textRunes {
			if c == ' ' {
				addAttr(position, position+1, spaceSpacingInt)
			}
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
	width := utils.PangoUnitsToFloat(logicalExtents.Width)
	height := utils.PangoUnitsToFloat(logicalExtents.Height)
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
