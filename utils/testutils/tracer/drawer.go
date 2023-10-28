package tracer

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/matrix"
)

// implements a logging backend, used for debugging

var _ backend.Document = Drawer{}

type Drawer struct {
	out io.Writer
}

func NewDrawerNoOp() Drawer { return Drawer{out: io.Discard} }

// NewDrawerFile panics if an error occurs.
func NewDrawerFile(outFile string) Drawer {
	f, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}

	return Drawer{out: f}
}

type fl = backend.Fl

func (dr Drawer) AddPage(left, top, right, bottom fl) backend.Page {
	fmt.Fprintln(dr.out, "AddPage :")
	return dr
}

func (dr Drawer) CreateAnchors(anchors [][]backend.Anchor) {
	fmt.Fprintln(dr.out, "CreateAnchors :")
}

func (dr Drawer) SetAttachments(as []backend.Attachment) {
	fmt.Fprintln(dr.out, "SetAttachments :")
}

func (dr Drawer) EmbedFile(id string, a backend.Attachment) {
	fmt.Fprintln(dr.out, "EmbedFile :")
}

func (dr Drawer) SetTitle(title string) {
	fmt.Fprintln(dr.out, "SetTitle :")
}

func (dr Drawer) SetDescription(description string) {
	fmt.Fprintln(dr.out, "SetDescription :")
}

func (dr Drawer) SetCreator(creator string) {
	fmt.Fprintln(dr.out, "SetCreator :")
}

func (dr Drawer) SetAuthors(authors []string) {
	fmt.Fprintln(dr.out, "SetAuthors :")
}

func (dr Drawer) SetKeywords(keywords []string) {
	fmt.Fprintln(dr.out, "SetKeywords :")
}

func (dr Drawer) SetProducer(producer string) {
	fmt.Fprintln(dr.out, "SetProducer :")
}

func (dr Drawer) SetDateCreation(d time.Time) {
	fmt.Fprintln(dr.out, "SetDateCreation :")
}

func (dr Drawer) SetDateModification(d time.Time) {
	fmt.Fprintln(dr.out, "SetDateModification :")
}

func (dr Drawer) SetBookmarks([]backend.BookmarkNode) {
	fmt.Fprintln(dr.out, "AddBookmark :")
}

func (dr Drawer) AddInternalLink(x, y, w, h fl, anchorName string) {
	fmt.Fprintln(dr.out, "AddInternalLink :")
}

func (dr Drawer) AddExternalLink(x, y, w, h fl, url string) {
	fmt.Fprintln(dr.out, "AddExternalLink :")
}

func (dr Drawer) AddFileAnnotation(x, y, w, h fl, id string) {
	fmt.Fprintln(dr.out, "AddFileAnnotation :")
}

func (dr Drawer) GetRectangle() (left, top, right, bottom fl) {
	fmt.Fprintln(dr.out, "GetRectangle :")
	return 0, 0, 10, 10
}

func (dr Drawer) SetMediaBox(left, top, right, bottom fl) {
	fmt.Fprintf(dr.out, "SetMediaBox : %.2f %.2f %.2f %.2f\n", left, top, right, bottom)
}

func (dr Drawer) SetTrimBox(left, top, right, bottom fl) {
	fmt.Fprintf(dr.out, "SetTrimBox : %.2f %.2f %.2f %.2f\n", left, top, right, bottom)
}

func (dr Drawer) SetBleedBox(left, top, right, bottom fl) {
	fmt.Fprintf(dr.out, "SetBleedBox : %.2f %.2f %.2f %.2f\n", left, top, right, bottom)
}

func (dr Drawer) OnNewStack(f func()) {
	fmt.Fprintln(dr.out, "OnNewStack :")
	f()
}

func (dr Drawer) Rectangle(x fl, y fl, width fl, height fl) {
	fmt.Fprintln(dr.out, "Rectangle :", x, y, width, height)
}

func (dr Drawer) Clip(evenOdd bool) {
	fmt.Fprintln(dr.out, "Clip :")
}

func (dr Drawer) SetColorRgba(color parser.RGBA, stroke bool) {
	if stroke {
		fmt.Fprintf(dr.out, "SetColorRgba : stroke %.2f %.2f %.2f\n", color.R, color.G, color.B)
	} else {
		fmt.Fprintf(dr.out, "SetColorRgba : fill %.2f %.2f %.2f\n", color.R, color.G, color.B)
	}
}

func (dr Drawer) SetLineWidth(width fl) {
	fmt.Fprintln(dr.out, "SetLineWidth :")
}

func (dr Drawer) SetDash(dashes []fl, offset fl) {
	fmt.Fprintln(dr.out, "SetDash :")
}

func (dr Drawer) Paint(op backend.PaintOp) {
	fmt.Fprintln(dr.out, "Paint :", op)
}

func (dr Drawer) GetTransform() matrix.Transform {
	fmt.Fprintln(dr.out, "GetTransform :")
	return matrix.Identity()
}

func (dr Drawer) Transform(mt matrix.Transform) {
	fmt.Fprintln(dr.out, "Transform :")
}

func (dr Drawer) MoveTo(x fl, y fl) {
	fmt.Fprintln(dr.out, "MoveTo :", x, y)
}

func (dr Drawer) LineTo(x fl, y fl) {
	fmt.Fprintln(dr.out, "LineTo :", x, y)
}

func (dr Drawer) CubicTo(x1, y1, x2, y2, x3, y3 fl) {
	fmt.Fprintln(dr.out, "CubicTo :", x1, y1, x2, y2, x3, y3)
}

func (dr Drawer) ClosePath() {
	fmt.Fprintln(dr.out, "ClosePath :")
}

func (dr Drawer) SetTextPaint(backend.PaintOp) {
	fmt.Fprintln(dr.out, "SetTextPaint :")
}

func (dr Drawer) SetBlendingMode(mode string) {
	fmt.Fprintln(dr.out, "SetTextPaint :")
}

func (dr Drawer) DrawText(text []backend.TextDrawing) {
	fmt.Fprintln(dr.out, "DrawText :", text)
}

func (dr Drawer) AddFont(backend.Font, []byte) *backend.FontChars {
	fmt.Fprintln(dr.out, "AddFont :")
	return &backend.FontChars{Cmap: make(map[backend.GID][]rune), Extents: make(map[backend.GID]backend.GlyphExtents)}
}

func (dr Drawer) NewGroup(x, y, width, height fl) backend.Canvas {
	fmt.Fprintln(dr.out, "NewGroup :")
	return dr
}

func (dr Drawer) DrawRasterImage(img backend.RasterImage, width, height fl) {
	fmt.Fprintln(dr.out, "DrawRasterImage :")
}

func (dr Drawer) DrawGradient(gradient backend.GradientLayout, width, height fl) {
	fmt.Fprintln(dr.out, "DrawGradient :")
}

func (dr Drawer) DrawWithOpacity(opacity fl, group backend.Canvas) {
	fmt.Fprintln(dr.out, "DrawGroup :")
}

func (dr Drawer) SetStrokeOptions(backend.StrokeOptions) {
	fmt.Fprintln(dr.out, "SetStrokeOptions :")
}

func (dr Drawer) SetColorPattern(backend.Canvas, fl, fl, matrix.Transform, bool) {
	fmt.Fprintln(dr.out, "SetColorPattern :")
}

func (dr Drawer) SetAlphaMask(mask backend.Canvas) {
	fmt.Fprintln(dr.out, "SetAlphaMask :")
}

func (dr Drawer) State() backend.GraphicState {
	return dr
}
