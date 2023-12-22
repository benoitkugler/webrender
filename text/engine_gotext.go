package text

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	bkLang "github.com/benoitkugler/textlayout/language"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/text/hyphen"
	"github.com/benoitkugler/webrender/utils"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/fontscan"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/opentype/api/metadata"
	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/segmenter"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

var _ FontConfiguration = (*FontConfigurationGotext)(nil)

type FontConfigurationGotext struct {
	fm         *fontscan.FontMap
	shaper     shaping.HarfbuzzShaper
	unicodeSeg segmenter.Segmenter
	inputSeg   shaping.Segmenter

	lineWrapper shaping.LineWrapper

	fontsContent  map[string][]byte        // to be embedded in the target
	fontsFeatures map[*font.Font][]Feature // as requested by @font-face
}

func NewFontConfigurationGotext(fm *fontscan.FontMap) *FontConfigurationGotext {
	out := FontConfigurationGotext{
		fm:            fm,
		fontsContent:  make(map[string][]byte),
		fontsFeatures: make(map[*font.Font][]Feature), // as loaded by loadOneFont
	}
	out.shaper.SetFontCacheSize(64)
	return &out
}

// AddFontFace load a font file from an external source, using
// the given [urlFetcher], which must be valid.
//
// It returns the file name of the loaded file.
func (f *FontConfigurationGotext) AddFontFace(ruleDescriptors validation.FontFaceDescriptors, urlFetcher utils.UrlFetcher) string {
	for _, url := range ruleDescriptors.Src {
		if url.String == "" {
			continue
		}
		if !(url.Name == "external" || url.Name == "local") {
			continue
		}

		filename, err := f.loadOneFont(url, ruleDescriptors, urlFetcher)
		if err != nil {
			logger.WarningLogger.Println(err)
			continue
		}

		return filename
	}

	logger.WarningLogger.Printf("Font face %s (src : %s) cannot be loaded", ruleDescriptors.FontFamily, ruleDescriptors.Src)
	return ""
}

// returns an error if the font is not found or has failed to be downloaded.
func (f *FontConfigurationGotext) loadOneFont(url pr.NamedString, ruleDescriptors validation.FontFaceDescriptors, urlFetcher utils.UrlFetcher) (string, error) {
	if url.Name == "local" {
		family := url.String
		// search through the system fonts, returning the filepath of the font, or an empty string
		// if no font matches the given family.
		location, ok := f.fm.FindSystemFont(family)
		if !ok {
			return "", fmt.Errorf("failed to load local font %s: not found", family)
		}

		// replace the family by an actual path
		var err error
		url.String, err = filepath.Abs(location.File)
		if err != nil {
			return "", fmt.Errorf("failed to load local font %s: %s", family, err)
		}
	}

	result, err := urlFetcher(url.String)
	if err != nil {
		return "", fmt.Errorf("failed to load font at %s: %s", url.String, err)
	}

	content, err := io.ReadAll(result.Content)
	if err != nil {
		return "", fmt.Errorf("failed to load font at %s", url.String)
	}

	lds, err := loader.NewLoaders(result.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse font at %s : %s", url.String, err)
	}
	if len(lds) != 1 {
		return "", fmt.Errorf("font collections are not supported (at %s)", url.String)
	}

	ft, err := font.NewFont(lds[0])
	if err != nil {
		return "", fmt.Errorf("failed to parse font at %s : %s", url.String, err)
	}

	if url.Name == "external" {
		f.fontsContent[url.String] = content
	}

	desc := metadata.Description{
		Family: string(ruleDescriptors.FontFamily),
		Aspect: newAspect(
			newFontStyle(ruleDescriptors.FontStyle),
			newFontWeight(ruleDescriptors.FontWeight),
			newFontStretch(ruleDescriptors.FontStretch),
		),
	}
	f.fm.AddFace(&font.Face{Font: ft}, fontscan.Location{File: url.String}, desc)

	// track the font features to apply
	f.fontsFeatures[ft] = getFontFaceFeatures(ruleDescriptors)

	return url.String, nil
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
		logger.WarningLogger.Println(err)
	}
	// cache the result to avoid loading the same file over and over
	f.fontsContent[font.File] = b

	return b
}

type layoutGotext struct {
	text []rune
}

// Text returns a readonly slice of the text in the layout
func (l layoutGotext) Text() []rune { return l.text }

// Metrics may return nil when [TextDecorationLine] is empty
func (layoutGotext) Metrics() *LineMetrics { return nil }

// Justification returns the current justification
func (layoutGotext) Justification() pr.Float { return 0 }

// SetJustification add an additional spacing between words
// to justify text. Depending on the implementation, it
// may be ignored until [ApplyJustification] is called.
func (layoutGotext) SetJustification(spacing pr.Float) {}

func (layoutGotext) ApplyJustification() {}

func newAspect(style FontStyle, weight uint16, stretch FontStretch) metadata.Aspect {
	aspect := metadata.Aspect{
		Style:  metadata.StyleNormal,
		Weight: metadata.Weight(weight),
	}
	if style == FSyItalic || style == FSyOblique {
		aspect.Style = metadata.StyleItalic
	}
	switch stretch {
	case FSeUltraCondensed:
		aspect.Stretch = metadata.StretchUltraCondensed
	case FSeExtraCondensed:
		aspect.Stretch = metadata.StretchExtraCondensed
	case FSeCondensed:
		aspect.Stretch = metadata.StretchCondensed
	case FSeSemiCondensed:
		aspect.Stretch = metadata.StretchSemiCondensed
	case FSeNormal:
		aspect.Stretch = metadata.StretchNormal
	case FSeSemiExpanded:
		aspect.Stretch = metadata.StretchSemiExpanded
	case FSeExpanded:
		aspect.Stretch = metadata.StretchExpanded
	case FSeExtraExpanded:
		aspect.Stretch = metadata.StretchExtraExpanded
	case FSeUltraExpanded:
		aspect.Stretch = metadata.StretchUltraExpanded
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
			Tag:   loader.NewTag(f.Tag[0], f.Tag[1], f.Tag[2], f.Tag[3]),
			Value: uint32(f.Value),
		}
	}
	return fts
}

func (fc *FontConfigurationGotext) resolveFace(r rune, font FontDescription) *font.Face {
	query := newQuery(font)
	fc.fm.SetQuery(query)
	return fc.fm.ResolveFace(r)
}

const sizeFactor = 100

// uses sizeFactor * font.Size
func (fc *FontConfigurationGotext) shape(r rune, font FontDescription, features []Feature) ([]shaping.Glyph, shaping.Bounds) {
	face := fc.resolveFace(r, font)
	if face == nil { // fontmap is broken
		return nil, shaping.Bounds{}
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
		Size: fixed.Int26_6(font.Size*64) * sizeFactor,
	})

	return out.Glyphs, out.LineBounds
}

func (fc *FontConfigurationGotext) heightx(style *TextStyle) pr.Fl {
	glyphs, _ := fc.shape('x', style.FontDescription, style.FontFeatures)

	if len(glyphs) == 0 { // fontmap is broken, return a 'reasonnable' value
		return style.FontDescription.Size
	}

	return pr.Fl(glyphs[0].YBearing) / 64 / sizeFactor // fixed to float
}

func (fc *FontConfigurationGotext) width0(style *TextStyle) pr.Fl {
	glyphs, _ := fc.shape('0', style.FontDescription, style.FontFeatures)

	if len(glyphs) == 0 { // fontmap is broken, return a 'reasonnable' value
		return style.FontDescription.Size
	}

	return pr.Fl(glyphs[0].XAdvance) / 64 / sizeFactor // fixed to float
}

func (fc *FontConfigurationGotext) spaceHeight(style *TextStyle) (height, baseline pr.Float) {
	_, bounds := fc.shape(' ', style.FontDescription, style.FontFeatures)

	height = pr.Float(bounds.Ascent-bounds.Descent) / 64 / sizeFactor
	baseline = pr.Float(bounds.Ascent) / 64 / sizeFactor

	return height, baseline
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
	var out [2]int
	// TODO: add word attr in typesetting
	out[1] = len(t)
	return &out
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
	return fc.wrapWordBreak(text, style, maxWidth, false)
}

// same as wrap, but may allows break inside words
func (fc *FontConfigurationGotext) wrapWordBreak(text []rune, style *TextStyle, maxWidth pr.Float, allowWordBreak bool) FirstLine {
	textWrap, spaceCollapse := style.textWrap(), style.spaceCollapse()
	mw := math.MaxInt
	if textWrap && maxWidth != pr.Inf {
		// use maxWidth
		mw = int(utils.MaxF(0, pr.Fl(maxWidth)))
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
	inputs := fc.inputSeg.Split(shaping.Input{
		Text:      text,
		RunEnd:    len(text),
		Language:  lang,
		Size:      fixed.Int26_6(style.FontDescription.Size * 64),
		Direction: di.DirectionLTR, // default, will be overriden
	}, fc.fm)

	// TODO: lazy iterator
	outputs := make([]shaping.Output, len(inputs))
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

	// now we can wrap the runs
	config := shaping.WrapConfig{BreakPolicy: shaping.Never} // mimic the default pango behavior
	if allowWordBreak {
		config.BreakPolicy = shaping.Always
	}
	fc.lineWrapper.Prepare(config, text, shaping.NewSliceIterator(outputs))
	wLine, fitsOnFirstLine := fc.lineWrapper.WrapNextLine(mw)
	line := wLine.Line

	if len(line) == 0 {
		return FirstLine{
			Layout:   layoutGotext{},
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

	if !fitsOnFirstLine && spaceCollapse {
		// remove the space runes...
		text = trimTrailingSpaces(text[:firstLineLength])
		firstLineLength = len(text)
		// and the matching glyphs
		lastRun := &line[len(line)-1]
		i := len(lastRun.Glyphs) - 1
		for ; i >= 0; i-- {
			if lastRun.Glyphs[i].Width != 0 {
				break
			}
		}
		lastRun.Glyphs = lastRun.Glyphs[:i+1]
		lastRun.RecalculateAll()
	}

	firstLineRTL := line[0].Direction.Progression() == di.TowardTopLeft

	var width, height, maxAscent fixed.Int26_6
	for _, run := range line {
		width += run.Advance
		if a := run.LineBounds.Ascent; a > maxAscent {
			maxAscent = a
		}
		if h := run.LineBounds.Ascent - run.LineBounds.Descent; h > height {
			height = h
		}
	}

	// TODO: properly handle letter spacing

	return FirstLine{
		Layout:       layoutGotext{text: text},
		Length:       firstLineLength,
		ResumeAt:     resumeAt,
		FirstLineRTL: firstLineRTL,
		Width:        pr.Float(width) / 64,
		Height:       pr.Float(height) / 64,
		Baseline:     pr.Float(maxAscent) / 64,
	}
}

// splitFirstLineGotext fit as much text from [text_] as possible in the available width given by [maxWidth].
// minimum should defaults to false
func (fc *FontConfigurationGotext) splitFirstLine(hyphenCache map[HyphenDictKey]hyphen.Hyphener, text_ string, style *TextStyle,
	maxWidth pr.MaybeFloat, minimum, isLineStart bool,
) FirstLine {
	// See https://www.w3.org/TR/css-text-3/#white-space-property
	var (
		textWrap         = style.textWrap()
		originalMaxWidth = maxWidth
		fontSize         = pr.Float(style.Size)
		firstLine        FirstLine
		text             = []rune(text_)
	)
	if !textWrap {
		maxWidth = nil
	}
	// Step #1: Get a draft layout with the first line
	if maxWidth, ok := maxWidth.(pr.Float); ok && maxWidth != pr.Inf && fontSize != 0 {
		// Try to use a small amount of text instead of the whole text
		shortText := shortTextHint(text_, maxWidth, fontSize)

		firstLine = fc.wrap([]rune(shortText), style, maxWidth)
		if firstLine.ResumeAt == -1 && len(shortText) != len(text_) {
			// The small amount of text fits in one line, give up and use the whole text
			firstLine = fc.wrap(text, style, maxWidth)
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

	firstLineText, secondLineText := text, []rune(nil)
	if firstLine.ResumeAt != -1 {
		firstLineText, secondLineText = text[:firstLine.ResumeAt], text[firstLine.ResumeAt:]
	}

	// Now, there is two cases :
	//	- firstLine.Width > maxWidthV : the first line is too long (only possible with one word)
	//	- firstLine.Width <= maxWidthV : the first line fits, but,
	//	since we wrap without using work breaks, the first word of the second line
	// 	could, after hyphenation, fit (partially) on the first line
	// That's why we either try to hyphenate the end of the first line or
	// the start of the second
	nextWord := secondLineText
	if firstLine.Width > maxWidthV {
		nextWord = firstLineText
	}

	// cut at the first space
	if i := index(secondLineText, ' '); i != -1 {
		nextWord = secondLineText[:i]
	}

	// Step #3: Try to hyphenate
	hyphens := style.Hyphens
	lang := bkLang.NewLanguage(style.Lang)
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

	var startWord, stopWord int
	if hyphens == HAuto && lang != "" {
		nextWordBoundaries := fc.wordBoundaries(nextWord)
		if nextWordBoundaries != nil {
			// We have a word to hyphenate
			startWord, stopWord = nextWordBoundaries[0], nextWordBoundaries[1]
			nextWord = secondLineText[startWord:stopWord]
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
		if hasSuffix(firstLineText, softHyphen) {
			// The first line has been split on a soft hyphen
			if id := lastIndex(firstLineText, ' '); id != -1 {
				firstLineText, nextWord = firstLineText[:id], firstLineText[id:] // next word start with a space
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
		dictionaryIterations = dictionary.IterateRunes(nextWord)
	}

	var hyphenatedFirstLineText []rune
	if len(dictionaryIterations) != 0 {
		var newFirstLineText []rune
		for _, firstWordPart := range dictionaryIterations {
			newFirstLineText = append(append(append([]rune(nil), firstLineText...), secondLineText[:startWord]...), []rune(firstWordPart)...)
			hyphenatedFirstLineText = append(newFirstLineText, hyphenateCharacter...)
			newFirstLine := fc.wrap(hyphenatedFirstLineText, style, maxWidthV)
			newSpace := maxWidthV - newFirstLine.Width
			hyphenated = newFirstLine.ResumeAt == -1 && (newSpace >= 0 || firstWordPart == dictionaryIterations[len(dictionaryIterations)-1])
			if hyphenated {
				firstLine = newFirstLine
				firstLine.Length -= len(hyphenateCharacter) // do not consider hyphen for length
				firstLine.ResumeAt = len(newFirstLineText)
				if text[firstLine.ResumeAt] == softHyphen {
					// Recreate the layout with no maxWidth to be sure that
					// we don't break before the soft hyphen
					firstLine.Layout = fc.wrap(hyphenatedFirstLineText, style, pr.Inf).Layout
					firstLine.ResumeAt += 1
				}
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
		hyphenatedFirstLineText = append(append([]rune(nil), firstLineText...), hyphenateCharacter...)
		firstLine = fc.wrap(hyphenatedFirstLineText, style, pr.Inf)
		firstLine.ResumeAt = len(firstLineText)
	}

	// Step #4: Try to break word if it's too long for the line
	overflowWrap, wordBreak := style.OverflowWrap, style.WordBreak
	space := maxWidthV - firstLine.Width
	// If we can break words and the first line is too long
	canBreak := wordBreak == WBBreakAll ||
		(isLineStart && (overflowWrap == OAnywhere || (overflowWrap == OBreakWord && !minimum)))
	if space < 0 && canBreak {
		// Is it really OK to remove hyphenation for word-break ?
		hyphenated = false
		firstLine = fc.wrapWordBreak(text, style, maxWidthV, true)
	}

	return firstLine
}
