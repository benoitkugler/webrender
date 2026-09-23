package text

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/text/hyphen"
	"github.com/benoitkugler/webrender/utils"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/fontscan"
	"github.com/go-text/typesetting/language"

	"github.com/go-text/typesetting/segmenter"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

var (
	_ FontConfiguration = (*FontConfigurationGotext)(nil)
	_ EngineLayout      = (*TextLayoutGotext)(nil)
)

type FontConfigurationGotext struct {
	fm         *fontscan.FontMap
	shaper     shaping.HarfbuzzShaper
	unicodeSeg segmenter.Segmenter
	inputSeg   shaping.Segmenter

	lineWrapper shaping.LineWrapper

	fontsContent  map[string][]byte        // to be embedded in the target
	fontsFeatures map[*font.Font][]Feature // as requested by @font-face

	textLayoutCache map[textKey]FirstLine
}

func NewFontConfigurationGotext(fm *fontscan.FontMap) *FontConfigurationGotext {
	out := FontConfigurationGotext{
		fm:              fm,
		fontsContent:    make(map[string][]byte),
		fontsFeatures:   make(map[*font.Font][]Feature), // as loaded by loadOneFont
		textLayoutCache: make(map[textKey]FirstLine),
	}
	out.shaper.SetFontCacheSize(64)
	return &out
}

// AddFontFace load a font file from an external source, using
// the given [urlFetcher], which must be valid.
//
// It returns the key of the loaded file.
func (f *FontConfigurationGotext) AddFontFace(ruleDescriptors validation.FontFaceDescriptors, urlFetcher utils.UrlFetcher) string {
	for _, url := range ruleDescriptors.Src {
		if url.S == "" {
			continue
		}
		if !(url.Tag == pr.External || url.Tag == pr.Local) {
			continue
		}

		filename, err := f.loadOneFont(url, ruleDescriptors, urlFetcher)
		if err != nil {
			logger.WarningLogger.Printf("AddFonfFace: %s\n", err)
			continue
		}

		return filename
	}

	logger.WarningLogger.Printf("Font face %s (src : %v) cannot be loaded", ruleDescriptors.FontFamily, ruleDescriptors.Src)
	return ""
}

// returns an error if the font is not found or has failed to be downloaded.
func (f *FontConfigurationGotext) loadOneFont(url pr.TaggedString, ruleDescriptors validation.FontFaceDescriptors, urlFetcher utils.UrlFetcher) (string, error) {
	if url.Tag == pr.Local {
		family := url.S
		// search through the system fonts, returning the filepath of the font, or an empty string
		// if no font matches the given family.
		location, ok := f.fm.FindSystemFont(family)
		if !ok {
			return "", fmt.Errorf("failed to load local font %s: not found", family)
		}

		// replace the family by an actual path
		var err error
		url.S, err = filepath.Abs(location.File)
		if err != nil {
			return "", fmt.Errorf("failed to load local font %s: %s", family, err)
		}
	}

	result, err := urlFetcher(url.S)
	if err != nil {
		return "", fmt.Errorf("failed to load font at %s: %s", url.S, err)
	}

	content, err := io.ReadAll(result.Content)
	if err != nil {
		return "", fmt.Errorf("failed to load font at %s", url.S)
	}

	lds, err := opentype.NewLoaders(result.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse font at %s : %s", url.S, err)
	}
	if len(lds) != 1 {
		return "", fmt.Errorf("font collections are not supported (at %s)", url.S)
	}

	ft, err := font.NewFont(lds[0])
	if err != nil {
		return "", fmt.Errorf("failed to parse font at %s : %s", url.S, err)
	}

	desc := font.Description{
		Family: string(ruleDescriptors.FontFamily),
		Aspect: newAspect(
			newFontStyle(ruleDescriptors.FontStyle),
			newFontWeight(ruleDescriptors.FontWeight),
			newFontStretch(ruleDescriptors.FontStretch),
		),
	}
	// add the face with a "unique" ID
	key := fmt.Sprintf("%s+%v", url.S, desc)
	f.fm.AddFace(font.NewFace(ft), fontscan.Location{File: key}, desc)

	if url.Tag == pr.External {
		f.fontsContent[key] = content
	}

	// track the font features to apply
	f.fontsFeatures[ft] = getFontFaceFeatures(ruleDescriptors)

	return key, nil
}

func (f *FontConfigurationGotext) FontLocation(font *font.Font) fontscan.Location {
	// TODO: simplify FontContent impl.
	return f.fm.FontLocation(font)
}

// FontContent returns the content of the given face, which may be needed
// in the final output.
func (f *FontConfigurationGotext) FontContent(font FontOrigin) []byte {
	// either the font is registred at run time or is loaded from disk
	// if registred at run time, its content has already been written in fontsContent
	if content, has := f.fontsContent[font.File]; has {
		return content
	}

	b, err := os.ReadFile(font.File)
	if err != nil {
		logger.WarningLogger.Println("FontContent:", err)
	}
	// cache the result to avoid loading the same file over and over
	f.fontsContent[font.File] = b

	return b
}

type TextLayoutGotext struct {
	text  []rune
	Style *TextStyle
	Line  shaping.Line

	MaxWidth pr.Float // input constraint, not always respected. Inf for no constraint

	justification pr.Float
}

// Text returns a readonly slice of the text in the layout
func (l TextLayoutGotext) Text() []rune { return l.text }

// Metrics may return nil when [TextDecorationLine] is empty
func (l TextLayoutGotext) Metrics() LineMetrics {
	if l.Style.TextDecorationLine == 0 {
		return LineMetrics{}
	}

	face := l.Line[0].Face
	factor := pr.Fl(l.Style.FontDescription.Size) / pr.Fl(face.Upem())

	extents, _ := face.FontHExtents()
	return LineMetrics{
		Ascent:                 pr.Fl(extents.Ascender) * factor,
		UnderlinePosition:      face.LineMetric(font.UnderlinePosition) * factor,
		UnderlineThickness:     face.LineMetric(font.UnderlineThickness) * factor,
		StrikethroughPosition:  face.LineMetric(font.StrikethroughPosition) * factor,
		StrikethroughThickness: face.LineMetric(font.StrikethroughThickness) * factor,
	}
}

// Justification returns the current justification
func (l TextLayoutGotext) Justification() pr.Float { return l.justification }

// SetJustification add an additional spacing between words
// to justify text. Depending on the implementation, it
// may be ignored until [ApplyJustification] is called.
func (l *TextLayoutGotext) SetJustification(spacing pr.Float) {
	l.justification = spacing
}

func (l *TextLayoutGotext) ApplyJustification() {
	shaping.AddSpacing(l.Line, l.text, floatToFixed(pr.Fl(l.justification)), 0)
}

func newAspect(style FontStyle, weight uint16, stretch FontStretch) font.Aspect {
	aspect := font.Aspect{
		Style:  font.StyleNormal,
		Weight: font.Weight(weight),
	}
	if style == FSty_Italic || style == FSty_Oblique {
		aspect.Style = font.StyleItalic
	}
	switch stretch {
	case FStr_UltraCondensed:
		aspect.Stretch = font.StretchUltraCondensed
	case FStr_ExtraCondensed:
		aspect.Stretch = font.StretchExtraCondensed
	case FStr_Condensed:
		aspect.Stretch = font.StretchCondensed
	case FStr_SemiCondensed:
		aspect.Stretch = font.StretchSemiCondensed
	case FStr_Normal:
		aspect.Stretch = font.StretchNormal
	case FStr_SemiExpanded:
		aspect.Stretch = font.StretchSemiExpanded
	case FStr_Expanded:
		aspect.Stretch = font.StretchExpanded
	case FStr_ExtraExpanded:
		aspect.Stretch = font.StretchExtraExpanded
	case FStr_UltraExpanded:
		aspect.Stretch = font.StretchUltraExpanded
	}
	return aspect
}

func newQuery(fd FontDescription) fontscan.Query {
	return fontscan.Query{
		Families: fd.Family,
		Aspect:   newAspect(fd.Style, fd.Weight, fd.Stretch),
	}
}

func newFeatures(features []Feature) []shaping.FontFeature {
	fts := make([]shaping.FontFeature, len(features))
	for i, f := range features {
		fts[i] = shaping.FontFeature{
			Tag:   opentype.NewTag(f.Tag[0], f.Tag[1], f.Tag[2], f.Tag[3]),
			Value: f.Value,
		}
	}
	return fts
}

func (fc *FontConfigurationGotext) resolveFace(r rune, font FontDescription) *font.Face {
	query := newQuery(font)
	fc.fm.SetQuery(query)
	return fc.fm.ResolveFace(r)
}

// uses sizeFactor * font.Size and returns sizeFactor
func (fc *FontConfigurationGotext) shapeRune(r rune, desc FontDescription, features []Feature) ([]shaping.Glyph, shaping.Bounds, pr.Float) {
	face := fc.resolveFace(r, desc)
	if face == nil { // fontmap is broken
		return nil, shaping.Bounds{}, 1
	}

	// sizeFactor is used to get a better precision
	sizeFactor := fixed.Int26_6(1000)
	if desc.Size >= 400 { // avoid overflowing go-text Int26_6 upper limit
		sizeFactor = 1
	} else if desc.Size >= 200 {
		sizeFactor = 5
	} else if desc.Size >= 100 {
		sizeFactor = 10
	}

	out := fc.shaper.Shape(shaping.Input{
		Text:     []rune{r},
		RunStart: 0, RunEnd: 1,
		Direction:    di.DirectionLTR,
		FontFeatures: newFeatures(features),
		Script:       language.Latin,
		Language:     language.NewLanguage("en"),
		Face:         face,
		// float to fixed, the size factor is to get a better precision
		Size: floatToFixed(desc.Size) * sizeFactor,
	})
	return out.Glyphs, out.LineBounds, pr.Float(sizeFactor)
}

func (fc *FontConfigurationGotext) heightx(style *TextStyle) pr.Fl {
	glyphs, _, sizeFactor := fc.shapeRune('x', style.FontDescription, style.FontFeatures)

	if len(glyphs) == 0 { // fontmap is broken, return a 'reasonnable' value
		return style.FontDescription.Size
	}

	return pr.Fl(fixedToFloat(glyphs[0].YBearing) / sizeFactor) // fixed to float
}

func (fc *FontConfigurationGotext) width0(style *TextStyle) pr.Fl {
	glyphs, _, sizeFactor := fc.shapeRune('0', style.FontDescription, style.FontFeatures)

	if len(glyphs) == 0 { // fontmap is broken, return a 'reasonnable' value
		return style.FontDescription.Size
	}
	return pr.Fl(fixedToFloat(glyphs[0].Advance) / sizeFactor) // fixed to float
}

func (fc *FontConfigurationGotext) spaceHeight(style *TextStyle) (height, baseline pr.Float) {
	_, bounds, sizeFactor := fc.shapeRune(' ', style.FontDescription, style.FontFeatures)

	height = fixedToFloat(bounds.Ascent-bounds.Descent) / sizeFactor
	baseline = fixedToFloat(bounds.Ascent) / sizeFactor

	return height, baseline
}

func (fc *FontConfigurationGotext) widthSpace(style *TextStyle) pr.Fl {
	glyphs, _, sizeFactor := fc.shapeRune(' ', style.FontDescription, style.FontFeatures)

	if len(glyphs) == 0 { // fontmap is broken, return a 'reasonnable' value
		return style.FontDescription.Size
	}

	return pr.Fl(fixedToFloat(glyphs[0].Advance) / sizeFactor) // fixed to float
}

func (fc *FontConfigurationGotext) CanBreakText(t []rune) pr.MaybeBool {
	if len(t) < 2 {
		return nil
	}
	fc.unicodeSeg.Init(t)
	iter := fc.unicodeSeg.LineIterator()
	if iter.Next() {
		line := iter.Line()
		end := line.Offset + len(line.Text)
		if end < len(t) {
			return pr.True
		}
	}
	return pr.False
}

// returns nil or a slice [wordStart:wordEnd]
func (fc *FontConfigurationGotext) wordBoundaries(t []rune) *[2]int {
	if len(t) < 2 {
		return nil
	}
	fc.unicodeSeg.Init(t)
	iter := fc.unicodeSeg.WordIterator()
	if iter.Next() {
		word := iter.Word()
		return &[2]int{word.Offset, word.Offset + len(word.Text)}
	}
	return nil
}

// return -1 if not found
func (fc *FontConfigurationGotext) getNextBreakPoint(text []rune) int {
	fc.unicodeSeg.Init(text)
	iter := fc.unicodeSeg.LineIterator()
	if iter.Next() {
		line := iter.Line()
		end := line.Offset + len(line.Text)
		return end
	}
	return -1
}

// GetLastWordEnd returns the index in `t` of the end of the before-last word,
// or -1
func (fc *FontConfigurationGotext) GetLastWordEnd(t []rune) int {
	if len(t) < 2 {
		return -1
	}
	fc.unicodeSeg.Init(t)
	iter := fc.unicodeSeg.WordIterator()
	var word1, word2 segmenter.Word
	for iter.Next() {
		word1 = word2
		word2 = iter.Word()
	}
	return word1.Offset + len(word1.Text)
}

// returns the first occurence of c, or -1 if not found
func index(text []rune, c rune) int {
	for i, r := range text {
		if r == c {
			return i
		}
	}
	return -1
}

// returns the last occurence of c, or -1 if not found
func lastIndex(text []rune, c rune) int {
	for i := len(text) - 1; i >= 0; i-- {
		if text[i] == c {
			return i
		}
	}
	return -1
}

func hasSuffix(text []rune, c rune) bool {
	return len(text) != 0 && text[len(text)-1] == c
}

// equivalent of python rstrip(' ')
func trimTrailingSpaces(text []rune) []rune {
	i := len(text) - 1
	for ; i >= 0; i-- {
		if text[i] != ' ' {
			break
		}
	}
	return text[:i+1]
}

// secondLineIndex is -1 if the whole [text] fits into the first line
// pass pr.Inf to remove width constraint
func (fc *FontConfigurationGotext) wrap(text []rune, style *TextStyle, maxWidth pr.Float) FirstLine {
	return fc.WrapWordBreak(text, style, maxWidth, false)
}

func floatToFixed(v pr.Fl) fixed.Int26_6    { return fixed.Int26_6(v * 64) }
func fixedToFloat(v fixed.Int26_6) pr.Float { return pr.Float(v) / 64 }

type textKey struct {
	text           string
	style          styleKey
	maxWidth       pr.Float
	allowWordBreak bool
}

// same as wrap, but may allows break inside words
func (fc *FontConfigurationGotext) WrapWordBreak(text []rune, style *TextStyle, maxWidth pr.Float, allowWordBreak bool) FirstLine {
	if len(text) == 0 {
		return FirstLine{
			Layout:   &TextLayoutGotext{Style: style, MaxWidth: maxWidth},
			Length:   0,
			ResumeAt: -1,
			Width:    0, Height: 0, Baseline: 0,
			FirstLineRTL: false,
		}
	}

	key := textKey{string(text), style.key(), maxWidth, allowWordBreak}
	if l, ok := fc.textLayoutCache[key]; ok {
		return l
	}

	var lang language.Language
	if flo := style.FontLanguageOverride; (flo != fontLanguageOverride{}) {
		lang = language.NewLanguage(lstToISO[flo])
	} else if lg := style.Lang; lg != "" {
		lang = language.NewLanguage(lg)
	} else {
		lang = language.DefaultLanguage()
	}

	// select the proper fonts
	fc.fm.SetQuery(newQuery(style.FontDescription))

	// segment the input text, with proper lang and size
	dir := di.DirectionLTR
	if style.Direction == pr.Rtl {
		dir = di.DirectionRTL
	}
	inputs := fc.inputSeg.Split(shaping.Input{
		Text:      text,
		RunEnd:    len(text),
		Language:  lang,
		Size:      floatToFixed(style.FontDescription.Size),
		Direction: dir, // default, will be overriden
	}, fc.fm)

	// TODO: lazy iterator
	outputs := make(shaping.Line, len(inputs))
	for i, input := range inputs {
		// the features are comming either from the style,
		// or registred via CSS @font-face rule
		defaults := newFeatureSet(fc.fontsFeatures[input.Face.Font])
		defaults.merge(style.FontFeatures)
		input.FontFeatures = newFeatures(defaults.list())

		// shape !
		output := fc.shaper.Shape(input)
		outputs[i] = output
	}

	// do we have some tabs ?
	if indexRune(text, '\t') != -1 {
		tabWidth := fixed.I(style.TabSize.Width)
		if style.TabSize.IsMultiple {
			spaceW := fc.widthSpace(style)
			tabWidth = floatToFixed(spaceW * pr.Fl(style.TabSize.Width))
		}
		AlignTabs(outputs, text, tabWidth, 0)
	}

	if style.LetterSpacing != 0 || style.WordSpacing != 0 {
		ws, ls := floatToFixed(style.WordSpacing), floatToFixed(style.LetterSpacing)
		shaping.AddSpacing(outputs, text, ws, ls)
		// add letter spacing at the end, like other browers do
		lastRun := &outputs[len(outputs)-1]
		lastRun.Glyphs[len(lastRun.Glyphs)-1].Advance += ls
		lastRun.RecomputeAdvance()
	}

	// now we can wrap the runs
	textWrap, spaceCollapse := style.textWrap(), style.spaceCollapse()
	// resolve the wraping width, ignoring the constraint when not wraping
	wrapingMaxWidth := maxWidth
	if !textWrap {
		wrapingMaxWidth = pr.Inf
	}
	mw, wLine, fitsOnFirstLine := fc.LineWrap(text, style, outputs, wrapingMaxWidth, allowWordBreak)
	line := wLine.Line

	if len(line) == 0 {
		return FirstLine{
			Layout:   &TextLayoutGotext{Style: style, MaxWidth: maxWidth},
			Length:   0,
			ResumeAt: -1,
			Width:    0, Height: 0, Baseline: 0,
			FirstLineRTL: false,
		}
	}

	resumeAt := wLine.NextLine
	firstLineLength := resumeAt
	if resumeAt == len(text) {
		resumeAt = -1
	}

	// handle mandatory break properly : go-text keeps it on the first line,
	// while weasyprint skips it
	// See https://unicode.org/reports/tr14/#BK
	// 000B LINE TABULATION (VT)
	// 000C FORM FEED (FF)
	// 2028 LINE SEPARATOR
	// 2029 PARAGRAPH SEPARATOR
	if r := text[wLine.NextLine-1]; r == '\n' || r == '\u000B' || r == '\u000C' || r == '\u2028' || r == '\u2029' {
		resumeAt = wLine.NextLine
		firstLineLength = wLine.NextLine - 1
		lastRun := &line[len(line)-1]
		lastRun.Glyphs = lastRun.Glyphs[:len(lastRun.Glyphs)-1]
		lastRun.RecalculateAll()
		wLine.TrimmedTrailingWhitespace = 0 // not to interfere with latter space handling
	}

	// sort the line by visual order
	sort.Slice(line, func(i, j int) bool { return line[i].VisualIndex < line[j].VisualIndex })

	if !fitsOnFirstLine && spaceCollapse {
		// remove the space runes...
		text = trimTrailingSpaces(text[:firstLineLength])
		trimmed := firstLineLength - len(text)
		firstLineLength = len(text)
		// and the matching glyphs
		lastRun := &line[len(line)-1]
		lastRun.Runes.Count -= trimmed // we assume all spaces are on the same run
		i := len(lastRun.Glyphs) - 1
		for ; i >= 0; i-- {
			if lastRun.Glyphs[i].Width != 0 {
				break
			}
		}
		lastRun.Glyphs = lastRun.Glyphs[:i+1]
		lastRun.RecalculateAll()
	}

	// copy the line, owned by lineWrapper
	outLine := slices.Clone(line)

	var (
		width fixed.Int26_6
		// directly fetch line bounds, to avoid rounding errors
		height, top pr.Float
	)
	for _, run := range outLine {
		width += run.Advance

		extents, _ := run.Face.FontHExtents()
		ascent := pr.Float(extents.Ascender) * pr.Float(style.FontDescription.Size) / pr.Float(run.Face.Upem())
		descent := pr.Float(extents.Descender) * pr.Float(style.FontDescription.Size) / pr.Float(run.Face.Upem())

		bottom := top - height

		if ascent > top {
			top = ascent
		}
		if descent < bottom {
			bottom = descent
		}
		height = top - bottom
	}

	if fitsOnFirstLine && spaceCollapse {
		// weasyprint puts the collapsed space on the line...
		if (width + wLine.TrimmedTrailingWhitespace) <= mw {
			width += wLine.TrimmedTrailingWhitespace
			// in RTL direction, the collapsed char must be expanded again
			// Note that the line is already sorted in visual order
			if dir.Progression() == di.TowardTopLeft {
				glyphs := outLine[0].Glyphs
				if len(glyphs) != 0 {
					glyphs[0].Advance += wLine.TrimmedTrailingWhitespace
				}
			}
		}
	}

	out := FirstLine{
		Layout:       &TextLayoutGotext{text: text[:firstLineLength], Style: style, Line: outLine, MaxWidth: maxWidth},
		Length:       firstLineLength,
		ResumeAt:     resumeAt,
		FirstLineRTL: style.Direction == pr.Rtl,
		Width:        fixedToFloat(width),
		Height:       height,
		Baseline:     top,
	}

	fc.textLayoutCache[key] = out

	return out
}

func (fc *FontConfigurationGotext) LineWrap(text []rune, style *TextStyle, runs shaping.Line, maxWidth pr.Float, allowWordBreak bool) (usedMaxWidth fixed.Int26_6, _ shaping.WrappedLine, _ bool) {
	dir := di.DirectionLTR
	if style.Direction == pr.Rtl {
		dir = di.DirectionRTL
	}

	// convert to fixed; properly handling large overflow
	const maxFixed = math.MaxInt32 >> 6 // max value for fixed.Int26_6
	mw := fixed.I(maxFixed)
	if maxWidth <= maxFixed {
		mw = floatToFixed(pr.Fl(maxWidth))
	}

	spaceCollapse := style.spaceCollapse()
	config := shaping.WrapConfig{
		Direction:                     dir,           // overall direction of the text
		BreakPolicy:                   shaping.Never, // mimic the default pango behavior
		DisableTrailingWhitespaceTrim: !spaceCollapse,
	}
	if allowWordBreak {
		config.BreakPolicy = shaping.Always
	}

	fc.lineWrapper.Prepare(config, text, shaping.NewSliceIterator(runs))
	line, done := fc.lineWrapper.WrapNextLineF(mw)
	return mw, line, done
}

// splitFirstLineGotext fit as much text from [text_] as possible in the available width given by [maxWidth].
// minimum should defaults to false
func (fc *FontConfigurationGotext) splitFirstLine(hyphenCache map[HyphenDictKey]hyphen.Hyphener, text []rune, style *TextStyle,
	maxWidth pr.MaybeFloat, minimum, isLineStart bool,
) FirstLine {
	// See https://www.w3.org/TR/css-text-3/#white-space-property
	var (
		textWrap         = style.textWrap()
		originalMaxWidth = maxWidth
		fontSize         = pr.Float(style.Size)
		firstLine        FirstLine
	)
	if !textWrap {
		maxWidth = nil
	}
	// Step #1: Get a draft layout with the first line
	if maxWidth, ok := maxWidth.(pr.Float); ok && maxWidth != pr.Inf && fontSize != 0 {
		// Try to use a small amount of text instead of the whole text
		shortText := shortTextHint(text, maxWidth, fontSize)

		firstLine = fc.wrap(shortText, style, maxWidth)
		if firstLine.ResumeAt == -1 && len(shortText) != len(text) {
			// The small amount of text fits in one line, give up and use the whole text
			firstLine = fc.wrap(text, style, maxWidth)
		} else {
			// If the second line of the short text can break, we have the next
			// line break point required for step #3 in it, drop the end of the text.
			if firstLine.ResumeAt != -1 && firstLine.ResumeAt != len(shortText) {
				firstLineText := shortText[:firstLine.ResumeAt]
				start := len(firstLineText) + 1
				if fc.getNextBreakPoint(shortText[start:]) != -1 {
					text = shortText
				}
			}
		}
	} else {
		originalMaxW := pr.Inf
		if originalMaxWidth != nil {
			originalMaxW = originalMaxWidth.V()
		}
		firstLine = fc.wrap(text, style, originalMaxW)
	}

	// Step #2: Don't split lines when it's not needed
	if maxWidth == nil || len(text) == 0 {
		// The first line can take all the place needed
		return firstLine
	}
	maxWidthV := maxWidth.V()

	if firstLine.ResumeAt == -1 && firstLine.Width <= maxWidthV {
		// The first line really fits in the available width
		return firstLine
	}

	// Step #3: Try to put the first word of the second line on the first line
	// https://mail.gnome.org/archives/gtk-i18n-list/2013-September/msg00006
	// is a good thread related to this problem.
	var firstLineText, secondLineText []rune
	if firstLine.Width <= maxWidthV {
		// the first line fits, but,
		//	since we wrap without using work breaks, the first word of the second line
		// 	could, after hyphenation, fit (partially) on the first line
		if firstLine.ResumeAt != -1 {
			firstLineText, secondLineText = text[:firstLine.ResumeAt], text[firstLine.ResumeAt:]
		}
	} else {
		// the first line is too long (only possible with one word)
		firstLineText = []rune{}
		secondLineText = text
	}

	breakPoint := len(secondLineText)
	if len(secondLineText) != 0 {
		// Find the second line’s first break point.
		breakPoint = fc.getNextBreakPoint(secondLineText)
		if breakPoint == -1 {
			breakPoint = len(secondLineText)
		}
	}

	nextWord := trimTrailingSpaces(secondLineText[:breakPoint])
	if len(nextWord) != 0 {
		if style.spaceCollapse() && hasSuffix(secondLineText[:breakPoint], ' ') {
			// Next word might fit without a space afterwards; only try when
			// space collapsing is allowed.
			newFirstLineText := slices.Concat(firstLineText, nextWord)
			firstLine = fc.wrap(newFirstLineText, style, maxWidthV)
			if firstLine.ResumeAt == -1 {
				if len(firstLineText) != 0 {
					// The next word fits in the first line, keep the layout.
					return firstLine
				} else {
					// Second line is None.
					firstLine.ResumeAt = firstLine.Length + 1
					if firstLine.ResumeAt >= len(text) {
						firstLine.ResumeAt = -1
					}
				}
			}
		}
	} else if len(firstLineText) != 0 {
		// We found something on the first line but we did not find a word on
		// the next line, no need to hyphenate, we can keep the current layout.
		return firstLine
	}

	// Step #4: Try to hyphenate
	hyphens := style.Hyphens
	lang := language.NewLanguage(style.Lang)
	if lang != "" {
		lang = hyphen.LanguageFallback(lang)
	}
	hyphenLimit := style.HyphenateLimitChars
	hyphenateCharacter := []rune(style.HyphenateCharacter)
	hyphenated := false
	const softHyphen = '\u00ad'

	autoHyphenation, manualHyphenation := false, false
	if hyphens != HNone {
		manualHyphenation = index(firstLineText, softHyphen) != -1 || index(nextWord, softHyphen) != -1
	}

	var startWord, stopWord, nextTextIndex int
	if hyphens == HAuto && lang != "" {
		// Get text until next line break opportunity.
		nextText := secondLineText
		if nextBreakPoint := fc.getNextBreakPoint(secondLineText); nextBreakPoint != -1 {
			nextText = nextText[:nextBreakPoint]
		}

		// Try all words included in this text.
		for len(nextText) != 0 {
			nextWordBoundaries := fc.wordBoundaries(nextText)
			if nextWordBoundaries != nil {
				// We have a word to hyphenate
				startWord, stopWord = nextWordBoundaries[0], nextWordBoundaries[1]
				nextWord = nextText[startWord:stopWord]
				if stopWord-startWord >= hyphenLimit.Total {
					// This word is long enough
					space := pr.Fl(maxWidthV - firstLine.Width)
					zone := style.HyphenateLimitZone
					limitZone := zone.Limit
					if zone.IsPercentage {
						limitZone = (pr.Fl(maxWidthV) * zone.Limit / 100.)
					}
					if space > limitZone || space < 0 {
						// Available space is worth the try, or the line is even too
						// long to fit: try to hyphenate
						autoHyphenation = true
						nextTextIndex += startWord
						break
					}
				}

				// This word doesn’t work, try next one.
				nextText = nextText[stopWord:]
				nextTextIndex += stopWord
			} else {
				break
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
		if hasSuffix(firstLineText, softHyphen) {
			// The first line has been split on a soft hyphen
			if i := lastIndex(firstLineText, ' '); i != -1 {
				firstLineText, nextWord = firstLineText[:i], firstLineText[i:] // next word start with a space
				firstLine = fc.wrap(firstLineText, style, maxWidthV)
				firstLine.ResumeAt = len(firstLineText) + 1 // track the space we have remove
			} else {
				firstLineText, nextWord = nil, firstLineText
			}
		}
		dictionaryIterations = hyphenDictionaryIterations(nextWord, softHyphen)
	} else if autoHyphenation {
		dictionaryKey := HyphenDictKey{lang, hyphenLimit}
		dictionary, ok := hyphenCache[dictionaryKey]
		if !ok {
			dictionary = hyphen.NewHyphener(lang, hyphenLimit.Left, hyphenLimit.Right)
			hyphenCache[dictionaryKey] = dictionary
		}
		previousWords := secondLineText[:nextTextIndex]
		dictionaryIterations = dictionary.IterateRunes(nextWord, string(previousWords))
	}

	var hyphenatedFirstLineText []rune
	if len(dictionaryIterations) != 0 {
		var newFirstLineText []rune
		for _, firstWordPart := range dictionaryIterations {
			newFirstLineText = slices.Concat(firstLineText, []rune(firstWordPart))
			hyphenatedFirstLineText = append(newFirstLineText, hyphenateCharacter...)
			newFirstLine := fc.wrap(hyphenatedFirstLineText, style, maxWidthV)
			newSpace := maxWidthV - newFirstLine.Width
			hyphenated = newFirstLine.ResumeAt == -1 && (newSpace >= 0 || firstWordPart == dictionaryIterations[len(dictionaryIterations)-1])
			if hyphenated {
				firstLine = newFirstLine
				firstLine.ResumeAt = len(newFirstLineText)
				break
			}
		}

		if !hyphenated && len(firstLineText) == 0 {
			// Recreate the layout with no maxWidth to be sure that
			// we don't break before or inside the hyphenate character
			hyphenated = true
			firstLine = fc.wrap(hyphenatedFirstLineText, style, pr.Inf)
			firstLine.ResumeAt = len(newFirstLineText)
			if text[firstLine.ResumeAt] == softHyphen {
				firstLine.ResumeAt += 1
			}
		}
	}

	if !hyphenated && hasSuffix(firstLineText, softHyphen) {
		// Recreate the layout with no maxWidth to be sure that
		// we don't break inside the hyphenate-character string
		hyphenated = true
		hyphenatedFirstLineText = slices.Concat(firstLineText, hyphenateCharacter)
		firstLine = fc.wrap(hyphenatedFirstLineText, style, pr.Inf)
		firstLine.ResumeAt = len(firstLineText)
	}

	// Step #5: Try to break word if it's too long for the line
	overflowWrap, wordBreak := style.OverflowWrap, style.WordBreak
	space := maxWidthV - firstLine.Width
	// If we can break words and the first line is too long
	canBreak := wordBreak == WBBreakAll ||
		(isLineStart && (overflowWrap == OAnywhere || (overflowWrap == OBreakWord && !minimum)))
	if space < 0 && canBreak {
		// Is it really OK to remove hyphenation for word-break ?
		hyphenated = false
		firstLine = fc.WrapWordBreak(text, style, maxWidthV, true)
	}

	if hyphenated {
		firstLine.Length -= len(hyphenateCharacter)
	}

	return firstLine
}

// tabs support

func applyTabs(out *shaping.Output, text []rune, columnWidth, runStart fixed.Int26_6) {
	columnWidthF := float64(columnWidth) / 64
	var advance fixed.Int26_6
	for i, g := range out.Glyphs {
		gAdvance := g.Advance
		isTab := g.RunesCount() == 1 && g.GlyphsCount() == 1 && text[g.TextIndex()] == '\t'
		if !isTab {
			advance += gAdvance
			continue
		}

		var updatedTabAdvance fixed.Int26_6
		if columnWidth == 0 {
			// simply trim the advance, nothing else to do
		} else {
			// update the advance of the glyph so that the next glyph is "tab-aligned" :
			// we want the "end" of the tab to be a multiple of columnWidth, that is :
			// (runStart + advance + updatedTabAdvance) % columnWith == 0
			glyphStartF := float64(runStart+advance) / 64
			remainder := math.Mod(glyphStartF, columnWidthF)
			updatedTabAdvance = fixed.Int26_6((columnWidthF - remainder) * 64)
		}

		out.Glyphs[i].Advance = updatedTabAdvance

		advance += updatedTabAdvance
	}

	// no need to call RecomputeAdvance
	out.Advance = advance
}

// AlignTabs updates the advance of glyphs mapped to '\t' runes,
// so that tabs are aligned on columns defined by [columnWidth].
//
// [lineOffset] may be non zero if the line starts after the first column.
//
// As a special case, if [columnWidth] is zero,
// tabs are trimmed (their advance is set to 0).
func AlignTabs(l shaping.Line, text []rune, columnWidth, lineOffset fixed.Int26_6) {
	runsAdvance := lineOffset // the position of the start of the current run
	for i := range l {
		applyTabs(&l[i], text, columnWidth, runsAdvance)
		runsAdvance += l[i].Advance
	}
}
