package text

import (
	"strings"

	"github.com/benoitkugler/textlayout/language"
	"github.com/benoitkugler/textprocessing/pango"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/text/hyphen"
)

func PangoUnitsFromFloat(v pr.Fl) int32 { return int32(v*pango.Scale + 0.5) }

func PangoUnitsToFloat(v pango.Unit) pr.Fl { return pr.Fl(v) / pango.Scale }

type TextLayoutContext interface {
	Fonts() *FontConfiguration
	HyphenCache() map[HyphenDictKey]hyphen.Hyphener
	StrutLayoutsCache() map[StrutLayoutKey][2]pr.Float
}

// TextLayout wraps a pango.Layout object
type TextLayout struct {
	style   *TextStyle
	Metrics *LineMetrics // optional

	MaxWidth pr.MaybeFloat

	Context TextLayoutContext // will be a *LayoutContext; to avoid circular dependency

	Layout pango.Layout

	JustificationSpacing pr.Fl
	FirstLineRTL         bool // true is the first line direction is RTL
}

func newTextLayout(context TextLayoutContext, style *TextStyle, justificationSpacing pr.Fl, maxWidth pr.MaybeFloat) *TextLayout {
	var layout TextLayout

	layout.JustificationSpacing = justificationSpacing
	layout.setup(context, style)
	layout.MaxWidth = maxWidth

	return &layout
}

// Text returns a readonly slice of the text used in the layout.
func (p *TextLayout) Text() []rune { return p.Layout.Text }

func (p *TextLayout) setup(context TextLayoutContext, style *TextStyle) {
	p.Context = context
	p.style = style
	p.FirstLineRTL = false
	fontmap := context.Fonts().Fontmap
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
		p.Metrics = &LineMetrics{
			Ascent:                 PangoUnitsToFloat(metrics.Ascent),
			UnderlinePosition:      PangoUnitsToFloat(metrics.UnderlinePosition),
			UnderlineThickness:     PangoUnitsToFloat(metrics.UnderlineThickness),
			StrikethroughPosition:  PangoUnitsToFloat(metrics.StrikethroughPosition),
			StrikethroughThickness: PangoUnitsToFloat(metrics.StrikethroughThickness),
		}
	} else {
		p.Metrics = nil
	}

	if len(style.FontFeatures) != 0 {
		attr := pango.NewAttrFontFeatures(pangoFontFeatures(style.FontFeatures))
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

	wordSpacing := p.style.WordSpacing
	if justify {
		// Justification is needed when drawing text but is useless during
		// layout, when it can be ignored.
		wordSpacing += p.JustificationSpacing
	}

	letterSpacing := p.style.LetterSpacing

	wordBreaking := p.style.OverflowWrap == OAnywhere || p.style.OverflowWrap == OBreakWord

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

func (p *TextLayout) setTabs() {
	tabSize := p.style.TabSize
	width := tabSize.Width
	if tabSize.IsMultiple { // no unit, means a multiple of the advance width of the space character
		layout := newTextLayout(p.Context, p.style, p.JustificationSpacing, nil)
		layout.SetText(strings.Repeat(" ", width))
		line, _ := layout.GetFirstLine()
		widthTmp, _ := lineSize(line, p.style.LetterSpacing)
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

	p.FirstLineRTL = firstLine.ResolvedDir%2 != 0

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

func defaultFontFeature(f string) string {
	if f == "" {
		return "normal"
	}
	return f
}
