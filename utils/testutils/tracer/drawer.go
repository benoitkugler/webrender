package tracer

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/matrix"
)

// implements a logging backend, used for debugging

var _ backend.Document = &Drawer{}

type Drawer struct {
	out    io.Writer
	indent int
}

func NewDrawerNoOp() *Drawer { return &Drawer{out: io.Discard} }

// NewDrawerFile panics if an error occurs.
func NewDrawerFile(outFile string) *Drawer {
	f, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}

	return &Drawer{out: f}
}

type fl = backend.Fl

func (dr Drawer) println(args ...interface{}) {
	fmt.Fprint(dr.out, strings.Repeat("  ", dr.indent))
	fmt.Fprintln(dr.out, args...)
}

func (dr Drawer) printf(f string, args ...interface{}) {
	fmt.Fprintf(dr.out, strings.Repeat(" ", dr.indent)+f+"\n", args...)
}

func (dr *Drawer) AddPage(left, top, right, bottom fl) backend.Page {
	dr.println("AddPage :")
	return dr
}

func (dr Drawer) CreateAnchors(anchors [][]backend.Anchor) {
	dr.println("CreateAnchors :")
}

func (dr Drawer) SetAttachments(as []backend.Attachment) {
	dr.println("SetAttachments :")
}

func (dr Drawer) EmbedFile(id string, a backend.Attachment) {
	dr.println("EmbedFile :")
}

func (dr Drawer) SetTitle(title string) {
	dr.println("SetTitle :")
}

func (dr Drawer) SetDescription(description string) {
	dr.println("SetDescription :")
}

func (dr Drawer) SetCreator(creator string) {
	dr.println("SetCreator :")
}

func (dr Drawer) SetAuthors(authors []string) {
	dr.println("SetAuthors :")
}

func (dr Drawer) SetKeywords(keywords []string) {
	dr.println("SetKeywords :")
}

func (dr Drawer) SetProducer(producer string) {
	dr.println("SetProducer :")
}

func (dr Drawer) SetDateCreation(d time.Time) {
	dr.println("SetDateCreation :")
}

func (dr Drawer) SetDateModification(d time.Time) {
	dr.println("SetDateModification :")
}

func (dr Drawer) SetBookmarks([]backend.BookmarkNode) {
	dr.println("AddBookmark :")
}

func (dr Drawer) AddInternalLink(x, y, w, h fl, anchorName string) {
	dr.println("AddInternalLink :")
}

func (dr Drawer) AddExternalLink(x, y, w, h fl, url string) {
	dr.println("AddExternalLink :")
}

func (dr Drawer) AddFileAnnotation(x, y, w, h fl, id string) {
	dr.println("AddFileAnnotation :")
}

func (dr Drawer) GetBoundingBox() (left, top, right, bottom fl) {
	dr.println("GetBoundingBox :")
	return 0, 0, 10, 10
}

func (dr Drawer) SetBoundingBox(left, top, right, bottom fl) {
	dr.printf("SetBoundingBox : %.2f %.2f %.2f %.2f", left, top, right, bottom)
}

func (dr Drawer) SetMediaBox(left, top, right, bottom fl) {
	dr.printf("SetMediaBox : %.2f %.2f %.2f %.2f", left, top, right, bottom)
}

func (dr Drawer) SetTrimBox(left, top, right, bottom fl) {
	dr.printf("SetTrimBox : %.2f %.2f %.2f %.2f", left, top, right, bottom)
}

func (dr Drawer) SetBleedBox(left, top, right, bottom fl) {
	dr.printf("SetBleedBox : %.2f %.2f %.2f %.2f", left, top, right, bottom)
}

func (dr *Drawer) OnNewStack(f func()) {
	dr.println("OnNewStack :")
	dr.indent++
	f()
	dr.indent--
}

func (dr Drawer) Rectangle(x fl, y fl, width fl, height fl) {
	dr.println("Rectangle :", x, y, width, height)
}

func (dr Drawer) Clip(evenOdd bool) {
	dr.println("Clip :")
}

func (dr Drawer) SetAlpha(alpha fl, stroke bool) {
	if stroke {
		dr.printf("SetAlpha : stroke %.2f", alpha)
	} else {
		dr.printf("SetAlpha : fill %.2f", alpha)
	}
}

func (dr Drawer) SetColorRgba(color parser.RGBA, stroke bool) {
	if stroke {
		dr.printf("SetColorRgba : stroke %.2f %.2f %.2f %.2f", color.R, color.G, color.B, color.A)
	} else {
		dr.printf("SetColorRgba : fill %.2f %.2f %.2f %.2f", color.R, color.G, color.B, color.A)
	}
}

func (dr Drawer) SetLineWidth(width fl) {
	dr.println("SetLineWidth :", width)
}

func (dr Drawer) SetDash(dashes []fl, offset fl) {
	dr.println("SetDash :", dashes, offset)
}

func (dr Drawer) Paint(op backend.PaintOp) {
	dr.println("Paint :", op)
}

func (dr Drawer) GetTransform() matrix.Transform {
	dr.println("GetTransform :")
	return matrix.Identity()
}

func (dr Drawer) Transform(mt matrix.Transform) {
	dr.println("Transform :", mt)
}

func (dr Drawer) MoveTo(x fl, y fl) {
	dr.println("MoveTo :", x, y)
}

func (dr Drawer) LineTo(x fl, y fl) {
	dr.println("LineTo :", x, y)
}

func (dr Drawer) CubicTo(x1, y1, x2, y2, x3, y3 fl) {
	dr.println("CubicTo :", x1, y1, x2, y2, x3, y3)
}

func (dr Drawer) ClosePath() {
	dr.println("ClosePath :")
}

func (dr Drawer) SetTextPaint(p backend.PaintOp) {
	dr.println("SetTextPaint :", p.String())
}

func (dr Drawer) SetBlendingMode(mode string) {
	dr.println("SetTextPaint :")
}

func (dr Drawer) DrawText(text []backend.TextDrawing) {
	for _, chunk := range text {
		dr.println("DrawText :", chunk.Matrix(), "font size:", chunk.FontSize)
		for _, run := range chunk.Runs {
			dr.println("->", run.Font.Origin(), ":", run.Glyphs)
		}
	}
}

func (dr Drawer) AddFont(backend.Font, []byte) *backend.FontChars {
	dr.println("AddFont :")
	return &backend.FontChars{Cmap: make(map[backend.GID][]rune), Extents: make(map[backend.GID]backend.GlyphExtents)}
}

func (dr *Drawer) NewGroup(x, y, width, height fl) backend.Canvas {
	dr.println("NewGroup :", x, y, width, height)
	return dr
}

func (dr Drawer) DrawRasterImage(img backend.RasterImage, width, height fl) {
	dr.println("DrawRasterImage :")
}

func (dr Drawer) DrawGradient(gradient backend.GradientLayout, width, height fl) {
	dr.println("DrawGradient :")
}

func (dr Drawer) DrawWithOpacity(opacity fl, group backend.Canvas) {
	dr.println("DrawWithOpacity :", opacity)
}

func (dr Drawer) SetStrokeOptions(opt backend.StrokeOptions) {
	dr.println("SetStrokeOptions :", opt)
}

func (dr Drawer) SetColorPattern(backend.Canvas, fl, fl, matrix.Transform, bool) {
	dr.println("SetColorPattern :")
}

func (dr Drawer) SetAlphaMask(mask backend.Canvas) {
	dr.println("SetAlphaMask :")
}

func (dr Drawer) State() backend.GraphicState {
	return dr
}
