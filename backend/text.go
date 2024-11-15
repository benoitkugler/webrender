package backend

import (
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/text"
)

// TextDrawing exposes the positionned text glyphs to draw
// and the associated font, in a backend independent manner
type TextDrawing struct {
	Runs []TextRun

	FontSize, ScaleX Fl
	X, Y             Fl // origin
	Angle            Fl // (optional) rotation
}

// Matrix return the transformation scaling the text by [FontSize],
// translating if to (X, Y)  and applying the [Angle] rotation
func (td TextDrawing) Matrix() matrix.Transform {
	mat := matrix.New(td.FontSize*td.ScaleX, 0, 0, -td.FontSize, td.X, td.Y)
	if td.Angle != 0 { // avoid useless multiplication if angle == 0
		mat.RightMultBy(matrix.Rotation(td.Angle))
	}
	return mat
}

// TextRun is a serie of glyphs with constant font.
type TextRun struct {
	Font   Font
	Glyphs []TextGlyph
}

type GID = uint32

// TextGlyph stores a glyph and it's position
type TextGlyph struct {
	Kerning  int // normalized by FontSize
	Glyph    GID
	Offset   Fl // normalized by FontSize
	Rise     Fl
	XAdvance Fl // how much to move before drawing, used for emojis
}

// GlyphExtents exposes glyph metrics, normalized by the font size.
type GlyphExtents struct {
	Width  int
	Y      int
	Height int
}

// FontChars stores some metadata that may be required in the output document.
type FontChars struct {
	Cmap    map[GID][]rune
	Extents map[GID]GlyphExtents
	Bbox    [4]int
}

// IsFixedPitch returns true if only one width is used,
// that is if the font is monospaced.
func (f *FontChars) IsFixedPitch() bool {
	seen := -1
	for _, w := range f.Extents {
		if seen == -1 {
			seen = w.Width
			continue
		}
		if w.Width != seen {
			return false
		}
	}
	return true
}

type FontDescription struct {
	Family string
	Style  text.FontStyle
	Weight int

	Ascent  Fl
	Descent Fl

	Size int // the font size used with this font

	IsOpentype bool
	// IsOpentype is true for an OpenType file containing a PostScript Type 2 font
	IsOpentypeOpentype bool
}

// Font are implemented by valid
// map keys
type Font interface {
	Origin() text.FontOrigin
	Description() FontDescription
}
