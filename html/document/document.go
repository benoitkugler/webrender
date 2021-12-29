// This package implements the high level parts of
// the document generation, but is still backend independant.
// It is meant to be used together with a `backend.Drawer`.
package document

import (
	"fmt"
	"log"
	"net/url"
	"path"

	"github.com/benoitkugler/webrender/html/layout/text"
	"github.com/benoitkugler/webrender/logger"
	mt "github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/version"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/layout"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/utils"
)

type (
	Color = parser.RGBA
	Box   = bo.Box
)

type fl = utils.Fl

func toF(v pr.Dimension) fl { return fl(v.Value) }

// Return the matrix for the CSS transform properties on this box (possibly nil).
func getMatrix(box_ Box) (mt.Transform, bool) {
	// "Transforms apply to block-level and atomic inline-level elements,
	//  but do not apply to elements which may be split into
	//  multiple inline-level boxes."
	// http://www.w3.org/TR/css3-2d-transforms/#introduction
	box := box_.Box()
	trans := box.Style.GetTransform()
	if len(trans) == 0 || bo.InlineBoxT.IsInstance(box_) {
		return mt.Transform{}, false
	}

	borderWidth := box.BorderWidth()
	borderHeight := box.BorderHeight()
	or := box.Style.GetTransformOrigin()
	offsetX := pr.ResoudPercentage(or[0].ToValue(), borderWidth).V()
	offsetY := pr.ResoudPercentage(or[1].ToValue(), borderHeight).V()
	originX := fl(box.BorderBoxX() + offsetX)
	originY := fl(box.BorderBoxY() + offsetY)

	matrix := mt.New(1, 0, 0, 1, originX, originY)
	for _, t := range trans {
		name, args := t.String, t.Dimensions
		// The length of args depends on `name`, see package validation for details.
		rightMat := mt.Identity()
		switch name {
		case "scale":
			sx, sy := toF(args[0]), toF(args[1])
			rightMat.Scale(sx, sy)
		case "rotate":
			angle := toF(args[0])
			rightMat.Rotate(angle)
		case "translate":
			translateX, translateY := args[0], args[1]
			rightMat.Translate(
				fl(pr.ResoudPercentage(translateX.ToValue(), borderWidth).V()),
				fl(pr.ResoudPercentage(translateY.ToValue(), borderHeight).V()),
			)
		case "skew":
			rightMat.Skew(toF(args[0]), toF(args[1]))
		case "matrix":
			rightMat = mt.New(toF(args[0]), toF(args[1]), toF(args[2]),
				toF(args[3]), toF(args[4]), toF(args[5]))
		default:
			panic(fmt.Sprintf("unexpected name for CSS transform property : %s", name))
		}
		matrix.Mult(rightMat) // same as matrix = mt.Mul(matrix, rightMat)
	}
	matrix.Translate(-originX, -originY) // same as matrix = mt.Mul(matrix, mt.New(1, 0, 0, 1, -originX, -originY))
	return matrix, true
}

// Apply a transformation matrix to an axis-aligned rectangle
// and return its axis-aligned bounding box as ``(x_min, y_min, x_max, y_max)``
func rectangleAabb(matrix mt.Transform, posX, posY, width, height fl) [4]fl {
	x1, y1 := matrix.TransformPoint(posX, posY)
	x2, y2 := matrix.TransformPoint(posX+width, posY)
	x3, y3 := matrix.TransformPoint(posX, posY+height)
	x4, y4 := matrix.TransformPoint(posX+width, posY+height)
	boxX1 := utils.Mins(x1, x2, x3, x4)
	boxY1 := utils.Mins(y1, y2, y3, y4)
	boxX2 := utils.Maxs(x1, x2, x3, x4)
	boxY2 := utils.Maxs(y1, y2, y3, y4)
	return [4]fl{boxX1, boxY1, boxX2, boxY2}
}

//  Link is a positionned link in a page.
type Link struct {
	// Type is one of three strings :
	// - "external": `target` is an absolute URL
	// - "internal": `target` is an anchor name
	//   The anchor might be defined in another page,
	//   in multiple pages (in which case the first occurence is used),
	//   or not at all.
	// - "attachment": `target` is an absolute URL and points
	//   to a resource to attach to the document.
	Type string

	Target string

	// [x_min, y_min, x_max, y_max] in CSS
	// pixels from the top-left of the page.
	Rectangle [4]fl
}

type bookmarkData struct {
	label    string
	open     bool
	position [2]fl
	level    int
}

func gatherLinksAndBookmarks(box_ bo.Box, bookmarks *[]bookmarkData, links *[]Link, anchors anchors, matrix *mt.Transform) {
	if transform, hasTransform := getMatrix(box_); hasTransform {
		if matrix != nil {
			t := mt.Mul(*matrix, transform)
			matrix = &t
		} else {
			matrix = &transform
		}
	}
	box := box_.Box()
	bookmarkLabel := box.BookmarkLabel
	bookmarkLevel := 0
	if lvl := box.Style.GetBookmarkLevel(); lvl.String != "none" {
		bookmarkLevel = lvl.Int
	}
	state := box.Style.GetBookmarkState()
	link := box.Style.GetLink()
	anchorName := string(box.Style.GetAnchor())
	hasBookmark := bookmarkLabel != "" && bookmarkLevel != 0
	// "link" is inherited but redundant on text boxes
	hasLink := !link.IsNone() && !(bo.TextBoxT.IsInstance(box_) || bo.LineBoxT.IsInstance(box_))
	// In case of duplicate IDs, only the first is an anchor.
	_, inAnchors := anchors[anchorName]
	hasAnchor := anchorName != "" && !inAnchors
	isAttachment := box.IsAttachment

	if hasBookmark || hasLink || hasAnchor {
		posX, posY, width, height := bo.HitArea(box_).Unpack()
		if hasLink {
			linkType, target := link.Name, link.String
			if linkType == "external" && isAttachment {
				linkType = "attachment"
			}
			linkS := Link{Type: linkType, Target: target}
			if matrix != nil {
				linkS.Rectangle = rectangleAabb(*matrix, posX, posY, width, height)
			} else {
				linkS.Rectangle = [4]fl{posX, posY, posX + width, posY + height}
			}
			*links = append(*links, linkS)
		}
		if matrix != nil && (hasBookmark || hasAnchor) {
			posX, posY = matrix.TransformPoint(posX, posY)
		}
		if hasBookmark {
			*bookmarks = append(*bookmarks, bookmarkData{
				level: bookmarkLevel, label: bookmarkLabel,
				position: [2]fl{posX, posY}, open: state == "open",
			})
		}
		if hasAnchor {
			anchors[anchorName] = [2]fl{posX, posY}
		}
	}

	for _, child := range box_.AllChildren() {
		gatherLinksAndBookmarks(child, bookmarks, links, anchors, matrix)
	}
}

type anchors = map[string][2]fl

// Page represents a single rendered page.
type Page struct {
	pageBox *bo.PageBox

	// The `dict` mapping each anchor name to its target, an
	// `(x, y)` point in CSS pixels from the top-left of the page.
	anchors anchors

	// `bookmarkLevel` and `bookmarkLabel` are based on
	// the CSS properties of the same names. `target` is an `(x, y)`
	// point in CSS pixels from the top-left of the page.
	bookmarks []bookmarkData

	links []Link

	// The page bleed widths with values in CSS pixels.
	Bleed Bleed

	// The page width, including margins, in CSS pixels.
	Width fl

	// The page height, including margins, in CSS pixels.
	Height fl
}

// newPage post-process a laid out `PageBox`.
func newPage(pageBox *bo.PageBox) Page {
	d := Page{}
	d.Width = fl(pageBox.MarginWidth())
	d.Height = fl(pageBox.MarginHeight())

	d.Bleed = Bleed{
		Top:    fl(pageBox.Style.GetBleedTop().Value),
		Right:  fl(pageBox.Style.GetBleedRight().Value),
		Bottom: fl(pageBox.Style.GetBleedBottom().Value),
		Left:   fl(pageBox.Style.GetBleedLeft().Value),
	}
	d.anchors = anchors{}

	gatherLinksAndBookmarks(
		pageBox, &d.bookmarks, &d.links, d.anchors, nil)
	d.pageBox = pageBox
	return d
}

// Paint the page on `dst`.
// leftX is the X coordinate of the left of the page, in user units.
// topY is the Y coordinate of the top of the page, in user units.
// scale is the Zoom scale in user units per CSS pixel.
// clip : whether to clip/cut content outside the page. If false, content can overflow.
// (leftX=0, topY=0, scale=1, clip=false)
func (d Page) Paint(dst backend.OutputPage, fc *text.FontConfiguration, leftX, topY, scale fl, clip bool) {
	dst.OnNewStack(func() {
		// Make (0, 0) the top-left corner and make user units CSS pixels
		dst.Transform(mt.New(scale, 0, 0, scale, leftX, topY))
		if clip {
			width := d.Width
			height := d.Height
			dst.Rectangle(0, 0, width, height)
			dst.Clip(false)
		}
		ctx := drawContext{dst: dst, fonts: fc}
		ctx.drawPage(d.pageBox)
	})
}

// Document is a rendered document ready to be painted on a drawing target.
//
// It is obtained by calling the `Render()` function.
type Document struct {
	// A list of `Page` objects.
	Pages []Page

	// A function called to fetch external resources such
	// as stylesheets and images.
	urlFetcher utils.UrlFetcher

	fontconfig *text.FontConfiguration

	// A `DocumentMetadata` object.
	// Contains information that does not belong to a specific page
	// but to the whole document.
	Metadata utils.DocumentMetadata
}

// Render performs the layout of the whole document and returns a document
// ready to be painted.
//
// fontConfig is mandatory
// presentationalHints should default to `false`
func Render(html *tree.HTML, stylesheets []tree.CSS, presentationalHints bool, fontConfig *text.FontConfiguration) Document {
	pageBoxes := layout.Layout(html, stylesheets, presentationalHints, fontConfig)
	pages := make([]Page, len(pageBoxes))
	for i, pageBox := range pageBoxes {
		pages[i] = newPage(pageBox)
	}
	return Document{Pages: pages, Metadata: html.GetMetadata(), urlFetcher: html.UrlFetcher, fontconfig: fontConfig}
}

// Take a subset of the pages.
//
// Examples:
// Write two PDF files for odd-numbered and even-numbered pages:
//     document.Copy(document.pages[::2]).writePdf("oddPages.pdf")
//     document.Copy(document.pages[1::2]).writePdf("evenPages.pdf")
// Combine multiple documents into one PDF file, using metadata from the first:
//		var allPages []Page
// 		for _, doc := range documents {
//		 	for _, p := range doc.pages {
//		 		allPages = append(allPages, p)
//		 	}
//		 }
//		documents[0].Copy(allPages).writePdf("combined.pdf")
// func (d Document) Copy(pages []Page, all bool) Document {
// 	if all {
// 		pages = d.Pages
// 	}
// 	return Document{Pages: pages, Metadata: d.Metadata, urlFetcher: d.urlFetcher}
// }

// Resolve internal hyperlinks.
// Links to a missing anchor are removed with a warning.
// If multiple anchors have the same name, the first one is used.
// Returns lists (one per page) like :attr:`Page.links`,
// except that ``target`` for internal hyperlinks is
// ``(pageNumber, x, y)`` instead of an anchor name.
// The page number is a 0-based index into the :attr:`pages` list,
// and ``x, y`` have been scaled (origin is at the top-left of the page).
func (d *Document) resolveLinks() ([][]Link, [][]backend.Anchor) {
	anchors := utils.NewSet()
	pagedAnchors := make([][]backend.Anchor, len(d.Pages))
	for i, page := range d.Pages {
		var current []backend.Anchor
		for anchorName, pos := range page.anchors {
			if !anchors.Has(anchorName) {
				current = append(current, backend.Anchor{Name: anchorName, X: pos[0], Y: pos[1]})
				anchors.Add(anchorName)
			}
		}
		pagedAnchors[i] = current
	}
	pagedLinks := make([][]Link, len(d.Pages))
	for i, page := range d.Pages {
		var pageLinks []Link
		for _, link := range page.links {
			// linkType, anchorName, rectangle = link
			if link.Type == "internal" {
				if !anchors.Has(link.Target) {
					log.Printf("No anchor #%s for internal URI reference\n", link.Target)
				} else {
					pageLinks = append(pageLinks, link)
				}
			} else {
				// External link
				pageLinks = append(pageLinks, link)
			}
		}
		pagedLinks[i] = pageLinks
	}
	return pagedLinks, pagedAnchors
}

// Make a tree of all bookmarks in the document.
func (d Document) makeBookmarkTree() []backend.BookmarkNode {
	// At one point in the document, for each "output" depth, how much
	// to add to get the source level (CSS values of bookmark-level).
	// E.g. with <h1> then <h3>, levelShifts == [0, 1]
	// 1 means that <h3> has depth 3 - 1 = 2 in the output.
	var (
		skippedLevels []int
		root          []backend.BookmarkNode
	)
	lastByDepth := []*[]backend.BookmarkNode{&root} // initialise with the root
	previousLevel := 0
	for pageNumber, page := range d.Pages {
		for _, bk := range page.bookmarks {
			level, label, pos, open := bk.level, bk.label, bk.position, bk.open
			if level > previousLevel {
				// Example: if the previous bookmark is a <h2>, the next
				// depth "should" be for <h3>. If now we get a <h6> we’re
				// skipping two levels: append 6 - 3 - 1 = 2
				skippedLevels = append(skippedLevels, level-previousLevel-1)
			} else {
				temp := level
				for temp < previousLevel {
					pop := skippedLevels[len(skippedLevels)-1]
					skippedLevels = skippedLevels[:len(skippedLevels)-1]
					temp += 1 + pop
				}
				if temp > previousLevel {
					// We remove too many "skips", add some back:
					skippedLevels = append(skippedLevels, temp-previousLevel-1)
				}
			}
			sum := 0
			for _, l := range skippedLevels {
				sum += l
			}
			previousLevel = level
			depth := level - sum
			if depth != len(skippedLevels) || depth < 1 {
				panic(fmt.Sprintf("expected depth >= 1 and depth == len(skippedLevels) got %d", depth))
			}
			subtree := backend.BookmarkNode{Label: label, PageIndex: pageNumber, X: pos[0], Y: pos[1], Open: open}
			(*lastByDepth[depth-1]) = append((*lastByDepth[depth-1]), subtree)
			lastByDepth = lastByDepth[:depth]
			tmp := *lastByDepth[depth-1]
			lastByDepth = append(lastByDepth, &tmp[len(tmp)-1].Children)
		}
	}
	return root
}

// Include hyperlinks in current PDF page.
func (d Document) addHyperlinks(links []Link, context backend.OutputPage, scale mt.Transform) {
	for _, link := range links {
		linkType, linkTarget, rectangle := link.Type, link.Target, link.Rectangle
		xMin, yMin := scale.TransformPoint(rectangle[0], rectangle[1])
		xMax, yMax := scale.TransformPoint(rectangle[2], rectangle[3])
		if linkType == "external" {
			context.AddExternalLink(xMin, yMin, xMax, yMax, linkTarget)
		} else if linkType == "internal" {
			context.AddInternalLink(xMin, yMin, xMax, yMax, linkTarget)
		} else if linkType == "attachment" {
			// actual embedding has be done previously
			context.AddFileAnnotation(xMin, yMin, xMax, yMax, linkTarget)
		}
	}
}

func (d *Document) scaleAnchors(anchors []backend.Anchor, matrix mt.Transform) {
	for i, a := range anchors {
		anchors[i].X, anchors[i].Y = matrix.TransformPoint(a.X, a.Y)
	}
}

func (d *Document) fetchAttachment(attachmentUrl string) backend.Attachment {
	// Attachments from document links like <link> or <a> can only be URLs.
	tmp, err := utils.FetchSource(utils.InputUrl(attachmentUrl), "", d.urlFetcher, false)
	if err != nil {
		log.Printf("Failed to load attachment at url %s: %s\n", attachmentUrl, err)
		return backend.Attachment{}
	}
	source, baseurl := tmp.Content, tmp.BaseUrl
	filename := getFilenameFromResult(baseurl)
	return backend.Attachment{Content: source, Title: filename}
}

// Derive a filename from a fetched resource.
// This is either the filename returned by the URL fetcher, the last URL path
// component or a synthetic name if the URL has no path.
func getFilenameFromResult(rawurl string) string {
	var filename string

	// The URL path likely contains a filename, which is a good second guess
	if rawurl != "" {
		u, err := url.Parse(rawurl)
		if err == nil {
			if u.Scheme != "data" {
				filename = path.Base(u.Path)
			}
		}
	}

	if filename == "" {
		// The URL lacks a path altogether. Use a synthetic name.

		// Using guessExtension is a great idea, but sadly the extension is
		// probably random, depending on the alignment of the stars, which car
		// you're driving and which software has been installed on your machine.
		extension := ".bin"
		filename = "attachment" + extension
	} else {
		filename = utils.Unquote(filename)
	}

	return filename
}

// Write paints the pages in the given `target`, with meta-data.
//
// The zoom factor is in PDF units per CSS units, and should default to 1.
// Warning : all CSS units are affected, including physical units like
// `cm` and named sizes like `A4`.  For values other than
// 1, the physical CSS units will thus be "wrong".
//
// `attachments` is an optional list of additional file attachments for the
// generated PDF document, added to those collected from the metadata.
func (d *Document) Write(target backend.Output, zoom pr.Fl, attachments []backend.Attachment) {
	// 0.75 = 72 PDF point per inch / 96 CSS pixel per inch
	scale := zoom * 0.75

	// Links and anchors
	pagedLinks, pagedAnchors := d.resolveLinks()

	// files must be embedded before being used on the pages
	d.embedFileAnnotations(pagedLinks, target)

	logger.ProgressLogger.Println("Step 6 - Drawing pages")

	for i, page := range d.Pages {
		pageWidth := scale * (page.Width + fl(page.Bleed.Left) + fl(page.Bleed.Right))
		pageHeight := scale * (page.Height + fl(page.Bleed.Top) + fl(page.Bleed.Bottom))
		left := -scale * page.Bleed.Left
		top := -scale * page.Bleed.Top
		right := left + pageWidth
		bottom := top + pageHeight

		outputPage := target.AddPage(left/scale, top/scale, (right-left)/scale, (bottom-top)/scale)
		outputPage.Transform(mt.New(1, 0, 0, -1, 0, page.Height*scale))
		page.Paint(outputPage, d.fontconfig, 0, 0, scale, false)

		// Draw from the top-left corner
		matrix := mt.New(scale, 0, 0, -scale, 0, page.Height*scale)

		d.addHyperlinks(pagedLinks[i], outputPage, matrix)
		d.scaleAnchors(pagedAnchors[i], matrix)
		page.Bleed.setMediaBoxes([4]fl{left, top, right, bottom}, outputPage)
	}

	target.CreateAnchors(pagedAnchors)

	logger.ProgressLogger.Println("Step 7 - Adding PDF metadata")

	// embedded files
	as := attachments
	for _, a := range d.Metadata.Attachments {
		t := d.fetchAttachment(a.URL)
		if len(t.Content) != 0 {
			as = append(as, t)
		}
	}
	target.SetAttachments(as)

	// Set bookmarks
	target.SetBookmarks(d.makeBookmarkTree())

	// Set document information
	target.SetTitle(d.Metadata.Title)
	target.SetDescription(d.Metadata.Description)
	target.SetCreator(d.Metadata.Generator)
	target.SetAuthors(d.Metadata.Authors)
	target.SetKeywords(d.Metadata.Keywords)
	target.SetProducer(version.VersionString)
	target.SetDateCreation(d.Metadata.Created)
	target.SetDateModification(d.Metadata.Modified)
}

func (d *Document) embedFileAnnotations(pagedLinks [][]Link, context backend.Output) {
	// A single link can be split in multiple regions.
	for _, rl := range pagedLinks {
		for _, link := range rl {
			if link.Type == "attachment" {
				a := d.fetchAttachment(link.Target)
				if len(a.Content) != 0 {
					context.EmbedFile(link.Target, a)
				}
			}
		}
	}
}

func (bleed Bleed) setMediaBoxes(mediaBox [4]fl, target backend.OutputPage) {
	bleed.Top *= 0.75
	bleed.Bottom *= 0.75
	bleed.Left *= 0.75
	bleed.Right *= 0.75

	// Add bleed box
	left, top, right, bottom := mediaBox[0], mediaBox[1], mediaBox[2], mediaBox[3]

	trimLeft := left + bleed.Left
	trimTop := top + bleed.Top
	trimRight := right - bleed.Right
	trimBottom := bottom - bleed.Bottom

	// Arbitrarly set PDF BleedBox between CSS bleed box (PDF MediaBox) and
	// CSS page box (PDF TrimBox), at most 10 px from the TrimBox.
	bleedLeft := trimLeft - utils.MinF(10, bleed.Left)
	bleedTop := trimTop - utils.MinF(10, bleed.Top)
	bleedRight := trimRight + utils.MinF(10, bleed.Right)
	bleedBottom := trimBottom + utils.MinF(10, bleed.Bottom)

	target.SetMediaBox(left, top, right, bottom)
	target.SetTrimBox(trimLeft, trimTop, trimRight, trimBottom)
	target.SetBleedBox(bleedLeft, bleedTop, bleedRight, bleedBottom)
}
