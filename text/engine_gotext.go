package text

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/logger"
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
	fontFilename := escapeXML(url.String)

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
		f.fontsContent[fontFilename] = content
	}

	desc := metadata.Description{
		Family: string(ruleDescriptors.FontFamily),
		Aspect: newAspect(
			newFontStyle(ruleDescriptors.FontStyle),
			newFontWeight(ruleDescriptors.FontWeight),
			newFontStretch(ruleDescriptors.FontStretch),
		),
	}
	f.fm.AddFace(&font.Face{Font: ft}, desc)

	// track the font features to apply
	f.fontsFeatures[ft] = getFontFaceFeatures(ruleDescriptors)

	return fontFilename, nil
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

// secondLineIndex is -1 if the whole [text] fits into the first line
// pass pr.Inf to remove width constraint
func (fc *FontConfigurationGotext) wrap(text []rune, style *TextStyle, maxWidth pr.Float) (firstLine shaping.Line, secondLineIndex int) {
	textWrap := style.textWrap()
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

	// segment the input text
	inputs := fc.inputSeg.Split(text, fc.fm, di.DirectionLTR)

	// TODO: lazy iterator
	var outputs []shaping.Output
	for _, input := range inputs {
		// apply lang, size and features ...
		input.Language = lang
		input.Size = fixed.Int26_6(style.FontDescription.Size * 64)

		// the features are comming either from the style,
		// or registred via CSS @font-face rule
		defaults := newFeatureSet(fc.fontsFeatures[input.Face.Font])
		defaults.merge(style.FontFeatures)
		input.FontFeatures = newFeatures(defaults.list())

		// ... and shape !
		output := fc.shaper.Shape(input)
		outputs = append(outputs, output)
	}

	fc.lineWrapper.Prepare(shaping.WrapConfig{}, text, shaping.NewSliceIterator(outputs))
	line, _, _ := fc.lineWrapper.WrapNextLine(mw)

	if len(line) == 0 {
		return line, -1
	}
	lastRun := line[len(line)-1]
	return line, lastRun.Runes.Offset + lastRun.Runes.Count
}

// SplitFirstLine2 fit as much text from [text_] as possible in the available width given by [maxWidth].
// minimum should defaults to false
func SplitFirstLine2(text_ string, style_ pr.StyleAccessor, context TextLayoutContext,
	maxWidth pr.MaybeFloat, minimum, isLineStart bool,
) Splitted {
	style := NewTextStyle(style_, false)
	// See https://www.w3.org/TR/css-text-3/#white-space-property
	var (
		ws               = style.WhiteSpace
		textWrap         = style.textWrap()
		spaceCollapse    = ws == WNormal || ws == WNowrap || ws == WPreLine
		originalMaxWidth = maxWidth
		fontSize         = pr.Float(style.Size)
		resumeIndex      int
		firstLine        shaping.Line
		fc               = context.Fonts().(*FontConfigurationGotext)
		text             = []rune(text_)
	)
	if !textWrap {
		maxWidth = nil
	}
	// Step #1: Get a draft layout with the first line
	if maxWidth, ok := maxWidth.(pr.Float); ok && maxWidth != pr.Inf && fontSize != 0 {
		// Try to use a small amount of text instead of the whole text
		shortText := shortTextHint(text_, maxWidth, fontSize)

		firstLine, resumeIndex = fc.wrap([]rune(shortText), style, maxWidth)
		if resumeIndex == -1 && len(shortText) != len(text_) {
			// The small amount of text fits in one line, give up and use the whole text
			firstLine, resumeIndex = fc.wrap(text, style, maxWidth)
		}
	} else {
		originalMaxW := pr.Inf
		if originalMaxWidth != nil {
			originalMaxW = originalMaxWidth.V()
		}
		firstLine, resumeIndex = fc.wrap(text, style, originalMaxW)
	}

	fmt.Println(spaceCollapse, len(firstLine), resumeIndex)

	// // Step #2: Don't split lines when it's not needed
	// if maxWidth == nil {
	// 	// The first line can take all the place needed
	// 	return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	// }
	// maxWidthV := pr.Fl(maxWidth.V())

	// firstLineWidth, _ := lineSize(firstLine, style.LetterSpacing)

	// if resumeIndex == -1 && firstLineWidth <= maxWidthV {
	// 	// The first line fits in the available width
	// 	return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	// }

	// // Step #3: Try to put the first word of the second line on the first line
	// // https://mail.gnome.org/archives/gtk-i18n-list/2013-September/msg00006
	// // is a good thread related to this problem.

	// firstLineText := text_
	// if resumeIndex != -1 && resumeIndex <= len(text) {
	// 	firstLineText = string(text[:resumeIndex])
	// }
	// firstLineFits := (firstLineWidth <= maxWidthV ||
	// 	strings.ContainsRune(strings.TrimSpace(firstLineText), ' ') ||
	// 	fc.CanBreakText([]rune(strings.TrimSpace(firstLineText))) == pr.True)
	// var secondLineText []rune
	// if firstLineFits {
	// 	// The first line fits but may have been cut too early by Pango
	// 	if resumeIndex == -1 {
	// 		secondLineText = text
	// 	} else {
	// 		secondLineText = text[resumeIndex:]
	// 	}
	// } else {
	// 	// The line can't be split earlier, try to hyphenate the first word.
	// 	firstLineText = ""
	// 	secondLineText = text
	// }

	// nextWord := strings.SplitN(string(secondLineText), " ", 2)[0]
	// if nextWord != "" {
	// 	if spaceCollapse {
	// 		// nextWord might fit without a space afterwards
	// 		// only try when space collapsing is allowed
	// 		newFirstLineText := firstLineText + nextWord
	// 		layout.SetText(newFirstLineText)
	// 		firstLine, resumeIndex = layout.GetFirstLine()
	// 		// firstLineWidth, _ = lineSize(firstLine, style.GetLetterSpacing())
	// 		if resumeIndex == -1 {
	// 			if firstLineText != "" {
	// 				// The next word fits in the first line, keep the layout
	// 				resumeIndex = len([]rune(newFirstLineText)) + 1
	// 				return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	// 			} else {
	// 				// Second line is none
	// 				resumeIndex = firstLine.Length + 1
	// 				if resumeIndex >= len(text) {
	// 					resumeIndex = -1
	// 				}
	// 			}
	// 		}
	// 	}
	// } else if firstLineText != "" {
	// 	// We found something on the first line but we did not find a word on
	// 	// the next line, no need to hyphenate, we can keep the current layout
	// 	return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	// }

	// // Step #4: Try to hyphenate
	// hyphens := style.Hyphens
	// lang := language.NewLanguage(style.Lang)
	// if lang != "" {
	// 	lang = hyphen.LanguageFallback(lang)
	// }
	// limit := style.HyphenateLimitChars
	// hyphenateCharacter := style.HyphenateCharacter
	// total, left, right := limit[0], limit[1], limit[2]
	// hyphenated := false
	// softHyphen := '\u00ad'

	// autoHyphenation, manualHyphenation := false, false
	// if hyphens != HNone {
	// 	manualHyphenation = strings.ContainsRune(firstLineText, softHyphen) || strings.ContainsRune(nextWord, softHyphen)
	// }

	// var startWord, stopWord int
	// if hyphens == HAuto && lang != "" {
	// 	nextWordBoundaries := getNextWordBoundaries(fc, secondLineText)
	// 	if len(nextWordBoundaries) == 2 {
	// 		// We have a word to hyphenate
	// 		startWord, stopWord = nextWordBoundaries[0], nextWordBoundaries[1]
	// 		nextWord = string(secondLineText[startWord:stopWord])
	// 		if stopWord-startWord >= total {
	// 			// This word is long enough
	// 			firstLineWidth, _ = lineSize(firstLine, style.LetterSpacing)
	// 			space := maxWidthV - firstLineWidth
	// 			zone := style.HyphenateLimitZone
	// 			limitZone := zone.Limit
	// 			if zone.IsPercentage {
	// 				limitZone = (maxWidthV * zone.Limit / 100.)
	// 			}
	// 			if space > limitZone || space < 0 {
	// 				// Available space is worth the try, or the line is even too
	// 				// long to fit: try to hyphenate
	// 				autoHyphenation = true
	// 			}
	// 		}
	// 	}
	// }

	// // Automatic hyphenation opportunities within a word must be ignored if the
	// // word contains a conditional hyphen, in favor of the conditional
	// // hyphen(s).
	// // See https://drafts.csswg.org/css-text-3/#valdef-hyphens-auto
	// var dictionaryIterations []string
	// if manualHyphenation {
	// 	// Manual hyphenation: check that the line ends with a soft
	// 	// hyphen and add the missing hyphen
	// 	if strings.HasSuffix(firstLineText, string(softHyphen)) {
	// 		// The first line has been split on a soft hyphen
	// 		if id := strings.LastIndexByte(firstLineText, ' '); id != -1 {
	// 			firstLineText, nextWord = firstLineText[:id], firstLineText[id+1:]
	// 			nextWord = " " + nextWord
	// 			layout.SetText(firstLineText)
	// 			firstLine, _ = layout.GetFirstLine()
	// 			resumeIndex = len([]rune(firstLineText + " "))
	// 		} else {
	// 			firstLineText, nextWord = "", firstLineText
	// 		}
	// 	}
	// 	dictionaryIterations = hyphenDictionaryIterations(nextWord, softHyphen)
	// } else if autoHyphenation {
	// 	dictionaryKey := HyphenDictKey{lang, left, right, total}
	// 	dictionary, ok := context.HyphenCache()[dictionaryKey]
	// 	if !ok {
	// 		dictionary = hyphen.NewHyphener(lang, left, right)
	// 		context.HyphenCache()[dictionaryKey] = dictionary
	// 	}
	// 	dictionaryIterations = dictionary.Iterate(nextWord)
	// }

	// if len(dictionaryIterations) != 0 {
	// 	var newFirstLineText, hyphenatedFirstLineText string
	// 	for _, firstWordPart := range dictionaryIterations {
	// 		newFirstLineText = (firstLineText + string(secondLineText[:startWord]) + firstWordPart)
	// 		hyphenatedFirstLineText = (newFirstLineText + hyphenateCharacter)
	// 		newLayout := createLayout(hyphenatedFirstLineText, style, fc, maxWidth)
	// 		newFirstLine, newIndex := newLayout.GetFirstLine()
	// 		newFirstLineWidth, _ := lineSize(newFirstLine, style.LetterSpacing)
	// 		newSpace := maxWidthV - newFirstLineWidth
	// 		hyphenated = newIndex == -1 && (newSpace >= 0 || firstWordPart == dictionaryIterations[len(dictionaryIterations)-1])
	// 		if hyphenated {
	// 			layout = newLayout
	// 			firstLine = newFirstLine
	// 			resumeIndex = len([]rune(newFirstLineText))
	// 			if text[resumeIndex] == softHyphen {
	// 				// Recreate the layout with no maxWidth to be sure that
	// 				// we don't break before the soft hyphen
	// 				layout.Layout.SetWidth(-1)
	// 				resumeIndex += 1
	// 			}
	// 			break
	// 		}
	// 	}

	// 	if !hyphenated && firstLineText == "" {
	// 		// Recreate the layout with no maxWidth to be sure that
	// 		// we don't break before or inside the hyphenate character
	// 		hyphenated = true
	// 		layout.SetText(hyphenatedFirstLineText)
	// 		layout.Layout.SetWidth(-1)
	// 		firstLine, _ = layout.GetFirstLine()
	// 		resumeIndex = len([]rune(newFirstLineText))
	// 		if text[resumeIndex] == softHyphen {
	// 			resumeIndex += 1
	// 		}
	// 	}
	// }

	// if !hyphenated && strings.HasSuffix(firstLineText, string(softHyphen)) {
	// 	// Recreate the layout with no maxWidth to be sure that
	// 	// we don't break inside the hyphenate-character string
	// 	hyphenated = true
	// 	hyphenatedFirstLineText := firstLineText + hyphenateCharacter
	// 	layout.SetText(hyphenatedFirstLineText)
	// 	layout.Layout.SetWidth(-1)
	// 	firstLine, _ = layout.GetFirstLine()
	// 	resumeIndex = len([]rune(firstLineText))
	// }

	// // Step 5: Try to break word if it's too long for the line
	// overflowWrap, wordBreak := style.OverflowWrap, style.WordBreak
	// firstLineWidth, _ = lineSize(firstLine, style.LetterSpacing)
	// space := maxWidthV - firstLineWidth
	// // If we can break words and the first line is too long
	// canBreak := wordBreak == WBBreakAll ||
	// 	(isLineStart && (overflowWrap == OAnywhere || (overflowWrap == OBreakWord && !minimum)))
	// if space < 0 && canBreak {
	// 	// Is it really OK to remove hyphenation for word-break ?
	// 	hyphenated = false
	// 	layout.SetText(string(text))
	// 	layout.Layout.SetWidth(pango.Unit(PangoUnitsFromFloat(maxWidthV)))
	// 	layout.Layout.SetWrap(pango.WRAP_CHAR)
	// 	var index int
	// 	firstLine, index = layout.GetFirstLine()
	// 	resumeIndex = index
	// 	if resumeIndex == 0 {
	// 		resumeIndex = firstLine.Length
	// 	}
	// 	if resumeIndex >= len(text) {
	// 		resumeIndex = -1
	// 	}
	// }

	// return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, hyphenated, hyphenateCharacter)

	return Splitted{}
}
