package text

import (
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/utils"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/fontscan"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/opentype/api/metadata"
	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/segmenter"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

var _ FontConfiguration = (*FontConfigurationGotext)(nil)

type FontConfigurationGotext struct {
	fm        *fontscan.FontMap
	shaper    shaping.HarfbuzzShaper
	segmenter segmenter.Segmenter
}

func NewFontConfigurationGotext(fm *fontscan.FontMap) *FontConfigurationGotext {
	out := FontConfigurationGotext{fm: fm}
	out.shaper.SetFontCacheSize(64)
	return &out
}

// FontContent returns the content of the given font, which may be needed
// in the final output.
func (FontConfigurationGotext) FontContent(font FontOrigin) []byte {
	// TODO:
	return nil
}

// AddFontFace load a font file from an external source, using
// the given [urlFetcher], which must be valid.
//
// It returns the file name of the loaded file.
func (FontConfigurationGotext) AddFontFace(ruleDescriptors validation.FontFaceDescriptors, urlFetcher utils.UrlFetcher) string {
	// TODO:
	return ""
}

func newQuery(fd FontDescription) fontscan.Query {
	aspect := metadata.Aspect{
		Style:  metadata.StyleNormal,
		Weight: metadata.Weight(fd.Weight),
	}
	if fd.Style == FSyItalic || fd.Style == FSyOblique {
		aspect.Style = metadata.StyleItalic
	}
	switch fd.Stretch {
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
	return fontscan.Query{
		Families: fd.Family,
		Aspect:   aspect,
	}
}

const sizeFactor = 100

// uses sizeFactor * font.Size
func (fc *FontConfigurationGotext) shape(r rune, font FontDescription, features []Feature) ([]shaping.Glyph, shaping.Bounds) {
	query := newQuery(font)
	fc.fm.SetQuery(query)
	face := fc.fm.ResolveFace(r)
	if face == nil { // fontmap is broken
		return nil, shaping.Bounds{}
	}

	fts := make([]shaping.FontFeature, len(features))
	for i, f := range features {
		fts[i] = shaping.FontFeature{
			Tag:   loader.NewTag(f.Tag[0], f.Tag[1], f.Tag[2], f.Tag[3]),
			Value: uint32(f.Value),
		}
	}
	out := fc.shaper.Shape(shaping.Input{
		Text:     []rune{r},
		RunStart: 0, RunEnd: 1,
		Direction:    di.DirectionLTR,
		FontFeatures: fts,
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
	fc.segmenter.Init(t)
	iter := fc.segmenter.LineIterator()
	if iter.Next() {
		line := iter.Line()
		end := line.Offset + len(line.Text)
		if end < len(t) {
			return pr.True
		}
	}
	return pr.False
}
