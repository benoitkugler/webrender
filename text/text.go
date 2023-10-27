package text

import (
	"fmt"
	"math"
	"strings"

	"github.com/benoitkugler/textlayout/language"
	"github.com/benoitkugler/textprocessing/pango"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/text/hyphen"
	"github.com/benoitkugler/webrender/utils"
)

// Splitted exposes the result of laying out
// one line of text
type Splitted struct {
	// pango Layout with the first line
	Layout *TextLayout

	// length in runes of the first line
	Length int

	// the number of runes to skip for the next line.
	// May be -1 if the whole text fits in one line.
	// This may be greater than `Length` in case of preserved
	// newline characters.
	ResumeAt int

	// Width is the width in pixels of the first line
	Width pr.Float

	// Height is the height in pixels of the first line
	Height pr.Float

	// Baselineis the baseline in pixels of the first line
	Baseline pr.Float
}

// CreateLayout returns a pango.Layout with default Pango line-breaks.
// `style` is a style dict of computed values.
// `maxWidth` is the maximum available width in the same unit as style.GetFontSize(),
// or `nil` for unlimited width.
func CreateLayout(text string, style *TextStyle, context TextLayoutContext, maxWidth pr.MaybeFloat, justificationSpacing pr.Float) *TextLayout {
	layout := newTextLayout(context, style, pr.Fl(justificationSpacing), maxWidth)
	ws := style.WhiteSpace
	textWrap := ws == WNormal || ws == WPreWrap || ws == WPreLine
	if maxWidth, ok := maxWidth.(pr.Float); ok && textWrap && maxWidth < 2<<21 {
		// Make sure that maxWidth * Pango.SCALE == maxWidth * 1024 fits in a
		// signed integer. Treat bigger values same as None: unconstrained width.
		layout.Layout.SetWidth(pango.Unit(PangoUnitsFromFloat(utils.Maxs(0, pr.Fl(maxWidth)))))
	}

	layout.SetText(text)
	return layout
}

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

// Fit as much as possible in the available width for one line of text.
// minimum=False
func SplitFirstLine(text_ string, style_ pr.StyleAccessor, context TextLayoutContext,
	maxWidth pr.MaybeFloat, justificationSpacing pr.Float, minimum, isLineStart bool,
) Splitted {
	style := NewTextStyle(style_, false)
	// See https://www.w3.org/TR/css-text-3/#white-space-property
	var (
		ws               = style.WhiteSpace
		textWrap         = ws == WNormal || ws == WPreWrap || ws == WPreLine
		spaceCollapse    = ws == WNormal || ws == WNowrap || ws == WPreLine
		originalMaxWidth = maxWidth
		layout           *TextLayout
		fontSize         = pr.Float(style.FontSize)
		firstLine        *pango.LayoutLine
		resumeIndex      int
	)
	if !textWrap {
		maxWidth = nil
	}
	// Step #1: Get a draft layout with the first line
	if maxWidth, ok := maxWidth.(pr.Float); ok && maxWidth != pr.Inf && fontSize != 0 {
		// shortText := text_
		cut := len(text_)
		if maxWidth <= 0 {
			// Trying to find minimum size, let's naively split on spaces and
			// keep one word + one letter

			if spaceIndex := strings.IndexByte(text_, ' '); spaceIndex != -1 {
				cut = spaceIndex + 2 // index + space + one letter
			}
		} else {
			cut = int(maxWidth / fontSize * 2.5)
		}

		if cut > len(text_) {
			cut = len(text_)
		}
		shortText := text_[:cut]

		// Try to use a small amount of text instead of the whole text
		layout = CreateLayout(shortText, style, context, maxWidth, justificationSpacing)
		firstLine, resumeIndex = layout.GetFirstLine()
		if resumeIndex == -1 && shortText != text_ {
			// The small amount of text fits in one line, give up and use the whole text
			layout.SetText(text_)
			firstLine, resumeIndex = layout.GetFirstLine()
		}
	} else {
		layout = CreateLayout(text_, style, context, originalMaxWidth, justificationSpacing)
		firstLine, resumeIndex = layout.GetFirstLine()
	}

	text := []rune(text_)

	// Step #2: Don't split lines when it's not needed
	if maxWidth == nil {
		// The first line can take all the place needed
		return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	}
	maxWidthV := pr.Fl(maxWidth.V())

	firstLineWidth, _ := lineSize(firstLine, style.LetterSpacing)

	if resumeIndex == -1 && firstLineWidth <= maxWidthV {
		// The first line fits in the available width
		return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	}

	// Step #3: Try to put the first word of the second line on the first line
	// https://mail.gnome.org/archives/gtk-i18n-list/2013-September/msg00006
	// is a good thread related to this problem.

	firstLineText := text_
	if resumeIndex != -1 && resumeIndex <= len(text) {
		firstLineText = string(text[:resumeIndex])
	}
	firstLineFits := (firstLineWidth <= maxWidthV ||
		strings.ContainsRune(strings.TrimSpace(firstLineText), ' ') ||
		CanBreakText([]rune(strings.TrimSpace(firstLineText))) == pr.True)
	var secondLineText []rune
	if firstLineFits {
		// The first line fits but may have been cut too early by Pango
		if resumeIndex == -1 {
			secondLineText = text
		} else {
			secondLineText = text[resumeIndex:]
		}
	} else {
		// The line can't be split earlier, try to hyphenate the first word.
		firstLineText = ""
		secondLineText = text
	}

	nextWord := strings.SplitN(string(secondLineText), " ", 2)[0]
	if nextWord != "" {
		if spaceCollapse {
			// nextWord might fit without a space afterwards
			// only try when space collapsing is allowed
			newFirstLineText := firstLineText + nextWord
			layout.SetText(newFirstLineText)
			firstLine, resumeIndex = layout.GetFirstLine()
			// firstLineWidth, _ = lineSize(firstLine, style.GetLetterSpacing())
			if resumeIndex == -1 {
				if firstLineText != "" {
					// The next word fits in the first line, keep the layout
					resumeIndex = len([]rune(newFirstLineText)) + 1
					return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
				} else {
					// Second line is none
					resumeIndex = firstLine.Length + 1
					if resumeIndex >= len(text) {
						resumeIndex = -1
					}
				}
			}
		}
	} else if firstLineText != "" {
		// We found something on the first line but we did not find a word on
		// the next line, no need to hyphenate, we can keep the current layout
		return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	}

	// Step #4: Try to hyphenate
	hyphens := style.Hyphens
	lang := language.NewLanguage(style.Lang)
	if lang != "" {
		lang = hyphen.LanguageFallback(lang)
	}
	limit := style.HyphenateLimitChars
	hyphenateCharacter := style.HyphenateCharacter
	total, left, right := limit[0], limit[1], limit[2]
	hyphenated := false
	softHyphen := '\u00ad'

	autoHyphenation, manualHyphenation := false, false
	if hyphens != HNone {
		manualHyphenation = strings.ContainsRune(firstLineText, softHyphen) || strings.ContainsRune(nextWord, softHyphen)
	}

	var startWord, stopWord int
	if hyphens == HAuto && lang != "" {
		nextWordBoundaries := getNextWordBoundaries(secondLineText)
		if len(nextWordBoundaries) == 2 {
			// We have a word to hyphenate
			startWord, stopWord = nextWordBoundaries[0], nextWordBoundaries[1]
			nextWord = string(secondLineText[startWord:stopWord])
			if stopWord-startWord >= total {
				// This word is long enough
				firstLineWidth, _ = lineSize(firstLine, style.LetterSpacing)
				space := maxWidthV - firstLineWidth
				zone := style.HyphenateLimitZone
				limitZone := zone.Limit
				if zone.IsPercentage {
					limitZone = (maxWidthV * zone.Limit / 100.)
				}
				if space > limitZone || space < 0 {
					// Available space is worth the try, or the line is even too
					// long to fit: try to hyphenate
					autoHyphenation = true
				}
			}
		}
	}

	// Automatic hyphenation opportunities within a word must be ignored if the
	// word contains a conditional hyphen, in favor of the conditional
	// hyphen(s).
	// See https://drafts.csswg.org/css-text-3/#valdef-hyphens-auto
	var dictionaryIterations []string
	if manualHyphenation {
		// Manual hyphenation: check that the line ends with a soft
		// hyphen and add the missing hyphen
		if strings.HasSuffix(firstLineText, string(softHyphen)) {
			// The first line has been split on a soft hyphen
			if id := strings.LastIndexByte(firstLineText, ' '); id != -1 {
				firstLineText, nextWord = firstLineText[:id], firstLineText[id+1:]
				nextWord = " " + nextWord
				layout.SetText(firstLineText)
				firstLine, _ = layout.GetFirstLine()
				resumeIndex = len([]rune(firstLineText + " "))
			} else {
				firstLineText, nextWord = "", firstLineText
			}
		}
		dictionaryIterations = hyphenDictionaryIterations(nextWord, softHyphen)
	} else if autoHyphenation {
		dictionaryKey := HyphenDictKey{lang, left, right, total}
		dictionary, ok := context.HyphenCache()[dictionaryKey]
		if !ok {
			dictionary = hyphen.NewHyphener(lang, left, right)
			context.HyphenCache()[dictionaryKey] = dictionary
		}
		dictionaryIterations = dictionary.Iterate(nextWord)
	}

	if len(dictionaryIterations) != 0 {
		var newFirstLineText, hyphenatedFirstLineText string
		for _, firstWordPart := range dictionaryIterations {
			newFirstLineText = (firstLineText + string(secondLineText[:startWord]) + firstWordPart)
			hyphenatedFirstLineText = (newFirstLineText + hyphenateCharacter)
			newLayout := CreateLayout(hyphenatedFirstLineText, style, context, maxWidth, justificationSpacing)
			newFirstLine, newIndex := newLayout.GetFirstLine()
			newFirstLineWidth, _ := lineSize(newFirstLine, style.LetterSpacing)
			newSpace := maxWidthV - newFirstLineWidth
			hyphenated = newIndex == -1 && (newSpace >= 0 || firstWordPart == dictionaryIterations[len(dictionaryIterations)-1])
			if hyphenated {
				layout = newLayout
				firstLine = newFirstLine
				resumeIndex = len([]rune(newFirstLineText))
				if text[resumeIndex] == softHyphen {
					// Recreate the layout with no maxWidth to be sure that
					// we don't break before the soft hyphen
					layout.Layout.SetWidth(-1)
					resumeIndex += 1
				}
				break
			}
		}

		if !hyphenated && firstLineText == "" {
			// Recreate the layout with no maxWidth to be sure that
			// we don't break before or inside the hyphenate character
			hyphenated = true
			layout.SetText(hyphenatedFirstLineText)
			layout.Layout.SetWidth(-1)
			firstLine, _ = layout.GetFirstLine()
			resumeIndex = len([]rune(newFirstLineText))
			if text[resumeIndex] == softHyphen {
				resumeIndex += 1
			}
		}
	}

	if !hyphenated && strings.HasSuffix(firstLineText, string(softHyphen)) {
		// Recreate the layout with no maxWidth to be sure that
		// we don't break inside the hyphenate-character string
		hyphenated = true
		hyphenatedFirstLineText := firstLineText + hyphenateCharacter
		layout.SetText(hyphenatedFirstLineText)
		layout.Layout.SetWidth(-1)
		firstLine, _ = layout.GetFirstLine()
		resumeIndex = len([]rune(firstLineText))
	}

	// Step 5: Try to break word if it's too long for the line
	overflowWrap, wordBreak := style.OverflowWrap, style.WordBreak
	firstLineWidth, _ = lineSize(firstLine, style.LetterSpacing)
	space := maxWidthV - firstLineWidth
	// If we can break words and the first line is too long
	canBreak := wordBreak == WBBreakAll ||
		(isLineStart && (overflowWrap == OAnywhere || (overflowWrap == OBreakWord && !minimum)))
	if space < 0 && canBreak {
		// Is it really OK to remove hyphenation for word-break ?
		hyphenated = false
		layout.SetText(string(text))
		layout.Layout.SetWidth(pango.Unit(PangoUnitsFromFloat(maxWidthV)))
		layout.Layout.SetWrap(pango.WRAP_CHAR)
		var index int
		firstLine, index = layout.GetFirstLine()
		resumeIndex = index
		if resumeIndex == 0 {
			resumeIndex = firstLine.Length
		}
		if resumeIndex >= len(text) {
			resumeIndex = -1
		}
	}

	return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, hyphenated, hyphenateCharacter)
}

func firstLineMetrics(firstLine *pango.LayoutLine, text []rune, layout *TextLayout, resumeAt int, spaceCollapse bool,
	style *TextStyle, hyphenated bool, hyphenationCharacter string,
) Splitted {
	length := firstLine.Length
	if hyphenated {
		length -= len([]rune(hyphenationCharacter))
	} else if resumeAt != -1 && resumeAt != 0 {
		// Set an infinite width as we don't want to break lines when drawing,
		// the lines have already been split and the size may differ. Rendering
		// is also much faster when no width is set.
		layout.Layout.SetWidth(-1)

		// Create layout with final text
		if length > len(text) {
			length = len(text)
		}
		firstLineText := string(text[:length])

		// Remove trailing spaces if spaces collapse
		if spaceCollapse {
			firstLineText = strings.TrimRight(firstLineText, " ")
		}

		layout.SetText(firstLineText)

		firstLine, _ = layout.GetFirstLine()
		length = 0
		if firstLine != nil {
			length = firstLine.Length
		}
	}

	width, height := lineSize(firstLine, style.LetterSpacing)
	baseline := PangoUnitsToFloat(layout.Layout.GetBaseline())
	return Splitted{Layout: layout, Length: length, ResumeAt: resumeAt, Width: pr.Float(width), Height: pr.Float(height), Baseline: pr.Float(baseline)}
}

var rp = strings.NewReplacer(
	"\u202a", "\u200b",
	"\u202b", "\u200b",
	"\u202c", "\u200b",
	"\u202d", "\u200b",
	"\u202e", "\u200b",
)

func getLogAttrs(text []rune) []pango.CharAttr {
	text = []rune(rp.Replace(string(text)))
	logAttrs := pango.ComputeCharacterAttributes(text, -1)
	return logAttrs
}

// returns nil or [wordStart, wordEnd]
func getNextWordBoundaries(t []rune) []int {
	if len(t) < 2 {
		return nil
	}
	out := make([]int, 2)
	hasBroken := false
	for i, attr := range getLogAttrs(t) {
		if attr.IsWordEnd() {
			out[1] = i // word end
			hasBroken = true
			break
		}
		if attr.IsWordBoundary() {
			out[0] = i // word start
		}
	}
	if !hasBroken {
		return nil
	}
	return out
}

// GetLastWordEnd returns the index in `t` if the last word,
// or -1
func GetLastWordEnd(t []rune) int {
	if len(t) < 2 {
		return -1
	}
	attrs := getLogAttrs(t)
	for i := 0; i < len(attrs); i++ {
		item := attrs[len(attrs)-1-i]
		if i != 0 && item.IsWordEnd() {
			return len(t) - i
		}
	}
	return -1
}

func CanBreakText(t []rune) pr.MaybeBool {
	if len(t) < 2 {
		return nil
	}
	logs := getLogAttrs(t)
	for _, l := range logs[1 : len(logs)-1] {
		if l.IsLineBreak() {
			return pr.True
		}
	}
	return pr.False
}

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
// `context` is mandatory for the text layout.
func StrutLayout(style_ pr.StyleAccessor, context TextLayoutContext) [2]pr.Float {
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

	layouts := context.StrutLayoutsCache()
	if v, ok := layouts[key]; ok {
		return v
	}

	layout := newTextLayout(context, style, 0, nil)
	layout.SetText(" ")
	line, _ := layout.GetFirstLine()
	sp := firstLineMetrics(line, nil, layout, -1, false, style, false, "")
	if lineHeight.String == "normal" {
		result := [2]pr.Float{sp.Height, sp.Baseline}
		if context != nil {
			context.StrutLayoutsCache()[key] = result
		}
		return result
	}
	lineHeightV := lineHeight.Value
	if lineHeight.Unit == pr.Scalar {
		lineHeightV *= pr.Float(fontSize)
	}
	result := [2]pr.Float{lineHeightV, sp.Baseline + (lineHeightV-sp.Height)/2}
	if context != nil {
		context.StrutLayoutsCache()[key] = result
	}
	return result
}

// CharacterRatio returns the ratio 1ex/font_size or 1ch/font_size, according to given style.
// It should be used with a valid text context to get accurate result.
// Otherwise, if context is `nil`, it returns 1 as a default value.
// It does not query WordSpacing or LetterSpacing from the style.
func CharacterRatio(style_ pr.ElementStyle, cache pr.TextRatioCache, isCh bool, context TextLayoutContext) pr.Float {
	if context == nil {
		return 1
	}

	style := NewTextStyle(style_, true) // avoid recursion for letter-spacing and word-spacing properties
	key := fontStyleCacheKey(style)
	if f, ok := cache.Get(key, isCh); ok {
		return f
	}

	// Random big value
	const fontSize pr.Fl = 1000
	style.FontSize = fontSize

	layout := newTextLayout(context, style, 0, nil)
	character := "x"
	if isCh {
		character = "0"
	}
	layout.Layout.SetText(character) // avoid recursion for letter-spacing and word-spacing properties
	line, _ := layout.GetFirstLine()

	var inkExtents, logicalExtents pango.Rectangle
	line.GetExtents(&inkExtents, &logicalExtents)
	var measure pr.Fl
	if isCh {
		measure = PangoUnitsToFloat(logicalExtents.Width)
	} else {
		measure = -PangoUnitsToFloat(inkExtents.Y)
	}

	// Zero means some kind of failure, fallback is 0.5.
	// We round to try keeping exact values that were altered by Pango.
	v := math.Round(float64(measure/fontSize)*100000) / 100000
	if v == 0 {
		return 0.5
	}
	out := pr.Float(v)
	cache.Set(key, isCh, out)
	return out
}

func fontStyleCacheKey(style *TextStyle) string {
	return fmt.Sprint(
		style.FontFamily,
		style.FontStyle,
		style.FontStretch,
		style.FontWeight,
		style.FontFeatures,
		style.FontVariationSettings,
	)
}
