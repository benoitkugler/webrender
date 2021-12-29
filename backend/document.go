package backend

import (
	"time"

	"github.com/benoitkugler/textlayout/fonts"
)

type Anchor struct {
	Name string
	// Origin at the top-left of the page
	X, Y Fl
}

type Attachment struct {
	Title, Description string
	Content            []byte
}

// GlyphExtents exposes glyph metrics, normalized by the font size.
type GlyphExtents struct {
	Width  int
	Y      int
	Height int
}

// Font stores some metadata used in the output document.
type Font struct {
	Cmap    map[fonts.GID][]rune
	Extents map[fonts.GID]GlyphExtents
	Bbox    [4]int
}

// IsFixedPitch returns true if only one width is used,
// that is if the font is monospaced.
func (f *Font) IsFixedPitch() bool {
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

// BookmarkNode exposes the outline hierarchy of the document
type BookmarkNode struct {
	Label     string
	Children  []BookmarkNode
	Open      bool // state of the outline item
	PageIndex int  // page index (0-based) to link to
	X, Y      Fl   // position in the page
}

// Document is the main target to whole the laid out document,
// consisting in pages, metadata and embedded files.
type Document interface {
	// AddPage creates a new page with the given dimensions and returns
	// it to be paint on.
	// The y axis grows downward, meaning bottom > top
	AddPage(left, top, right, bottom Fl) Page

	// CreateAnchors register a list of anchors per page, which are named targets of internal links.
	// `anchors` is a 0-based list, meaning anchors in page 1 are at index 0.
	// The origin of internal link has been be added by `OutputPage.AddInternalLink`.
	// `CreateAnchors` is called after all the pages have been created and processed
	CreateAnchors(anchors [][]Anchor)

	// Add global attachments to the file
	SetAttachments(as []Attachment)

	// Embed a file. Calling this method twice with the same id
	// won't embed the content twice.
	// `fileID` will be passed to `OutputPage.AddFileAnnotation`
	EmbedFile(fileID string, a Attachment)

	// Metadatas

	SetTitle(title string)
	SetDescription(description string)
	SetCreator(creator string)
	SetAuthors(authors []string)
	SetKeywords(keywords []string)
	SetProducer(producer string)
	SetDateCreation(d time.Time)
	SetDateModification(d time.Time)

	// SetBookmarks setup the document outline
	SetBookmarks(root []BookmarkNode)
}

// Page is the target of one laid out page,
// composed of a Canvas and link supports.
type Page interface {
	// AddInternalLink shows a link on the page, pointing to the
	// named anchor, which will be registered with `Output.CreateAnchors`
	AddInternalLink(xMin, yMin, xMax, yMax Fl, anchorName string)

	// AddExternalLink shows a link on the page, pointing to
	// the given url
	AddExternalLink(xMin, yMin, xMax, yMax Fl, url string)

	// AddFileAnnotation adds a file annotation on the current page.
	// The file content has been added with `Output.EmbedFile`.
	AddFileAnnotation(xMin, yMin, xMax, yMax Fl, fileID string)

	// Adjust the media boxes

	SetMediaBox(left, top, right, bottom Fl)
	SetTrimBox(left, top, right, bottom Fl)
	SetBleedBox(left, top, right, bottom Fl)

	Canvas
}
