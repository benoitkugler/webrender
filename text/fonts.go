package text

import (
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/utils"
)

// FontOrigin is a reference to a binary font file, either
// on disk or stored in memory.
type FontOrigin struct {
	File string // The filename or identifier of the font file.

	// The index of the face in a collection. It is always 0 for
	// single font files.
	Index uint16

	// For variable fonts, stores 1 + the instance index.
	// (0 to ignore variations).
	Instance uint16
}

// FontConfiguration holds information about the
// available fonts on the system.
// It is used for text layout at various steps of the rendering process.
//
// It is implemented by and totaly tighted to text engines, either pango or go-text.
type FontConfiguration interface {
	// FontContent returns the content of the given font, which may be needed
	// in the final output.
	FontContent(font FontOrigin) []byte

	// AddFontFace load a font file from an external source, using
	// the given [urlFetcher], which must be valid.
	//
	// It returns the file name of the loaded file.
	AddFontFace(ruleDescriptors validation.FontFaceDescriptors, urlFetcher utils.UrlFetcher) string

	// CanBreakText returns True if there is a line break strictly inside [t], False otherwise.
	// It should return nil if t has length < 2.
	CanBreakText(t []rune) pr.MaybeBool

	// returns the advance of the '0' char, using the font described by the given [style]
	width0(style *TextStyle) pr.Fl
	// returns the height of the 'x' char, using the font described by the given [style]
	heightx(style *TextStyle) pr.Fl
	// returns the height and baseline of a line containing a single space (" ")
	spaceHeight(style *TextStyle) (height, baseline pr.Float)

	// compute the unicode propery of the given runes,
	// returning a slice of length L + 1
	// the returned slice is readonly, and valid only until the
	// next call to runeProps
	// runeProps([]rune) []runeProp
}
