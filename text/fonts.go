package text

import (
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
// It is implemented by text engines, using pango or go-text.
type FontConfiguration interface {
	// FontContent returns the content of the given font, which may be needed
	// in the final output.
	FontContent(font FontOrigin) []byte

	// AddFontFace load a font file from an external source, using
	// the given [urlFetcher], which must be valid.
	//
	// It returns the file name of the loaded file.
	AddFontFace(ruleDescriptors validation.FontFaceDescriptors, urlFetcher utils.UrlFetcher) string
}
