package text

import (
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

// FirstLine exposes the result of laying out
// one line of text
type FirstLine struct {
	// Output layout containing (at least) the first line
	Layout EngineLayout

	// Length in runes of the first line
	Length int

	// ResumeAt is the number of runes to skip for the next line.
	// May be -1 if the whole text fits in one line.
	// This may be greater than [Length] in case of preserved
	// newline characters or white space collapse.
	ResumeAt int

	// Width is the width in pixels of the first line
	Width pr.Float

	// Height is the height in pixels of the first line,
	// computed by merging the font extents (y position and height)
	// of each run in the line.
	Height pr.Float

	// Baseline is the baseline in pixels of the first line
	Baseline pr.Float

	FirstLineRTL bool // true is the first line direction is RTL
}

// split word on each hyphen occurence, starting by the end
func hyphenDictionaryIterations(word []rune, hyphen rune) (out []string) {
	for i := len(word) - 1; i >= 0; i-- {
		if word[i] == hyphen {
			out = append(out, string(word[:i+1]))
		}
	}
	return out
}

type HyphenDictKey struct {
	lang  language.Language
	limit pr.Limits
}

// returns a prefix of text
func shortTextHint(text string, maxWidth, fontSize pr.Float) string {
	cut := len(text)
	if maxWidth <= 0 {
		// Trying to find minimum size, let's naively split on spaces and
		// keep one word + one letter

		if spaceIndex := strings.IndexByte(text, ' '); spaceIndex != -1 {
			cut = spaceIndex + 2 // index + space + one letter
		}
	} else {
		cut = int(maxWidth / fontSize * 2.5)
	}

	if cut > len(text) {
		cut = len(text)
	}

	return text[:cut]
}

// SplitFirstLine fit as much text from [text] as possible in the available width given by [maxWidth].
// minimum should default to [false]
func SplitFirstLine(text string, style_ pr.StyleAccessor, context TextLayoutContext,
	maxWidth pr.MaybeFloat, minimum, isLineStart bool,
) FirstLine {
	style := NewTextStyle(style_, false)
	return context.Fonts().splitFirstLine(context.HyphenCache(), text, style, maxWidth, minimum, isLineStart)
}

type StrutLayoutKey struct {
	lang                 string
	fontFamily           string // joined
	lineHeight           pr.DimOrS
	fontWeight           uint16
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

	fontSize := style.Size
	if fontSize == 0 {
		return [2]pr.Float{}
	}

	lineHeight := style_.GetLineHeight()

	key := StrutLayoutKey{
		fontSize:             fontSize,
		fontLanguageOverride: style.FontLanguageOverride,
		lang:                 style.Lang,
		fontFamily:           strings.Join(style.Family, ""),
		fontStyle:            style.Style,
		fontStretch:          style.Stretch,
		fontWeight:           style.Weight,
		lineHeight:           lineHeight,
	}

	cache := context.StrutLayoutsCache()
	if v, ok := cache[key]; ok {
		return v
	}

	height, baseline := context.Fonts().spaceHeight(style)

	if lineHeight.S == "normal" {
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
	style.FontDescription.Size = fontSize

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
	return string(append(style.FontDescription.binary(nil, false), featuresBinary(style.FontFeatures)...))
}
