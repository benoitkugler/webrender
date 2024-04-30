package text

import (
	"fmt"
	"math"
	"strings"

	"github.com/benoitkugler/textlayout/language"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/text/hyphen"
)

type TextLayoutContext interface {
	Fonts() FontConfiguration
	HyphenCache() map[HyphenDictKey]hyphen.Hyphener
	StrutLayoutsCache() map[StrutLayoutKey][2]pr.Float
}

// EngineLayout stores the text engine dependant version of the laid out
// text.
//
// Is is only meant to be consumed by the text/draw package.
type EngineLayout interface {
	// Text returns a readonly slice of the text in the layout
	Text() []rune

	// Metrics may return nil when [TextDecorationLine] is empty
	Metrics() *LineMetrics

	// Justification returns the current justification
	Justification() pr.Float
	// SetJustification add an additional spacing between words
	// to justify text. Depending on the implementation, it
	// may be ignored until [ApplyJustification] is called.
	SetJustification(spacing pr.Float)

	ApplyJustification()
}

// Splitted exposes the result of laying out
// one line of text
type Splitted struct {
	// Output layout containing (at least) the first line
	Layout EngineLayout

	// Length in runes of the first line
	Length int

	// ResumeAt is the number of runes to skip for the next line.
	// May be -1 if the whole text fits in one line.
	// This may be greater than [Length] in case of preserved
	// newline characters.
	ResumeAt int

	// Width is the width in pixels of the first line
	Width pr.Float

	// Height is the height in pixels of the first line
	Height pr.Float

	// Baseline is the baseline in pixels of the first line
	Baseline pr.Float

	FirstLineRTL bool // true is the first line direction is RTL
}

// split word on each hyphen occurence, starting by the end
func hyphenDictionaryIterations(word string, hyphen rune) (out []string) {
	wordRunes := []rune(word)
	for i := len(wordRunes) - 1; i >= 0; i-- {
		if wordRunes[i] == hyphen {
			out = append(out, string(wordRunes[:i+1]))
		}
	}
	return out
}

type HyphenDictKey struct {
	lang               language.Language
	left, right, total int
}

// SplitFirstLine fit as much text from [text_] as possible in the available width given by [maxWidth]
// minimum=False
func SplitFirstLine(text_ string, style_ pr.StyleAccessor, context TextLayoutContext,
	maxWidth pr.MaybeFloat, minimum, isLineStart bool,
) Splitted {
	return splitFirstLine(text_, style_, context, maxWidth, minimum, isLineStart)
}

func CanBreakText(fc FontConfiguration, t []rune) pr.MaybeBool {
	if len(t) < 2 {
		return nil
	}
	logs := fc.runeProps(t)
	for _, l := range logs[1 : len(logs)-1] {
		if l&isLineBreak != 0 {
			return pr.True
		}
	}
	return pr.False
}

type runeProp uint8

// bit mask
const (
	isWordEnd runeProp = 1 << iota
	isWordBoundary
	isLineBreak
)

type StrutLayoutKey struct {
	lang                 string
	fontFamily           string // joined
	lineHeight           pr.Value
	fontWeight           int
	fontSize             pr.Fl
	fontLanguageOverride fontLanguageOverride
	fontStretch          FontStretch
	fontStyle            FontStyle
}

// StrutLayout returns a tuple of the used value of `line-height` and the baseline.
// The baseline is given from the top edge of line height.
// [context] is mandatory for the text layout.
func StrutLayout(style_ pr.StyleAccessor, context TextLayoutContext) (result [2]pr.Float) {
	style := NewTextStyle(style_, false)

	fontSize := style.FontSize
	if fontSize == 0 {
		return [2]pr.Float{}
	}

	lineHeight := style_.GetLineHeight()

	key := StrutLayoutKey{
		fontSize:             fontSize,
		fontLanguageOverride: style.FontLanguageOverride,
		lang:                 style.Lang,
		fontFamily:           strings.Join(style.FontFamily, ""),
		fontStyle:            style.FontStyle,
		fontStretch:          style.FontStretch,
		fontWeight:           style.FontWeight,
		lineHeight:           lineHeight,
	}

	cache := context.StrutLayoutsCache()
	if v, ok := cache[key]; ok {
		return v
	}

	height, baseline := context.Fonts().spaceHeight(style)

	if lineHeight.String == "normal" {
		result = [2]pr.Float{height, baseline}
	} else {
		lineHeightV := lineHeight.Value
		if lineHeight.Unit == pr.Scalar {
			lineHeightV *= pr.Float(fontSize)
		}
		result = [2]pr.Float{lineHeightV, baseline + (lineHeightV-height)/2}
	}

	cache[key] = result
	return result
}

// CharacterRatio returns the ratio 1ex/font_size or 1ch/font_size, according to given style.
// It should be used with a valid text context to get accurate result.
// Otherwise, if context is `nil`, it returns 1 as a default value.
// It does not query WordSpacing or LetterSpacing from the style.
func CharacterRatio(style_ pr.ElementStyle, cache pr.TextRatioCache, isCh bool, fonts FontConfiguration) pr.Float {
	if fonts == nil {
		return 1
	}

	style := NewTextStyle(style_, true) // avoid recursion for letter-spacing and word-spacing properties
	key := style.cacheKey()
	if f, ok := cache.Get(key, isCh); ok {
		return f
	}

	// Random big value
	const fontSize pr.Fl = 1000
	style.FontSize = fontSize

	var measure pr.Fl
	if isCh {
		measure = fonts.width0(style)
	} else {
		measure = fonts.heightx(style)
	}

	// Zero means some kind of failure, fallback is 0.5.
	// We round to try keeping exact values that were altered by the engine.
	v := pr.Float(math.Round(float64(measure/fontSize)*100000) / 100000)
	if v == 0 {
		v = 0.5
	}
	cache.Set(key, isCh, v)
	return v
}

func (style *TextStyle) cacheKey() string {
	return fmt.Sprint(
		style.FontFamily,
		style.FontStyle,
		style.FontStretch,
		style.FontWeight,
		style.FontFeatures,
		style.FontVariationSettings,
	)
}
