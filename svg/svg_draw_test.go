package svg

import (
	"io"
	"log"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/matrix"
)

var outputLog = log.New(io.Discard, "svg_test: ", 0)

type fl = backend.Fl

var _ backend.Canvas = outputPage{}

type outputPage struct{}

func (outputPage) GetBoundingBox() (left, top, right, bottom fl) {
	outputLog.Println("GetBoundingBox")
	return 0, 0, 10, 10
}

func (outputPage) SetBoundingBox(left, top, right, bottom fl) {
	outputLog.Println("SetBoundingBox")
}

func (outputPage) OnNewStack(f func()) {
	outputLog.Println("OnNewStack")
	f()
}

func (outputPage) Rectangle(x fl, y fl, width fl, height fl) {
	outputLog.Println("Rectangle")
}

func (outputPage) Clip(evenOdd bool) {
	outputLog.Println("Clip")
}

func (outputPage) SetColorRgba(color parser.RGBA, stroke bool) {
	outputLog.Println("SetColorRgba")
}

func (outputPage) SetLineWidth(width fl) {
	outputLog.Println("SetLineWidth")
}

func (outputPage) SetDash(dashes []fl, offset fl) {
	outputLog.Println("SetDash")
}

func (outputPage) Paint(backend.PaintOp) {
	outputLog.Println("Paint")
}

func (outputPage) Transform(mt matrix.Transform) {
	outputLog.Println("Transform")
}

func (outputPage) GetTransform() matrix.Transform {
	outputLog.Println("GetTransform")
	return matrix.Identity()
}

func (outputPage) MoveTo(x fl, y fl) {
	outputLog.Println("MoveTo")
}

func (outputPage) LineTo(x fl, y fl) {
	outputLog.Println("LineTo")
}

func (outputPage) CubicTo(x1, y1, x2, y2, x3, y3 fl) {
	outputLog.Println("CubicTo")
}

func (outputPage) ClosePath() {
	outputLog.Println("ClosePath")
}

func (outputPage) SetTextPaint(backend.PaintOp) {
	outputLog.Println("SetTextPaint")
}

func (outputPage) SetBlendingMode(mode string) {
	outputLog.Println("SetTextPaint")
}

func (outputPage) DrawText(text []backend.TextDrawing) {
	outputLog.Println("DrawText", text)
}

func (outputPage) AddFont(backend.Font, []byte) *backend.FontChars {
	outputLog.Println("AddFont")
	return &backend.FontChars{Cmap: make(map[backend.GID][]rune), Extents: make(map[backend.GID]backend.GlyphExtents)}
}

func (outputPage) NewGroup(x, y, width, height Fl) backend.Canvas {
	outputLog.Println("NewGroup")
	return outputPage{}
}

func (outputPage) DrawRasterImage(img backend.RasterImage, width, height fl) {
	outputLog.Println("DrawRasterImage")
}

func (outputPage) SetAlphaMask(mask backend.Canvas) {
	outputLog.Println("SetAlphaMask")
}

func (outputPage) DrawGradient(gradient backend.GradientLayout, width, height fl) {
	outputLog.Println("DrawGradient")
}

func (outputPage) DrawWithOpacity(opacity fl, group backend.Canvas) {
	outputLog.Println("DrawGroup")
}

func (outputPage) SetStrokeOptions(backend.StrokeOptions) {
	outputLog.Println("SetStrokeOptions")
}

func (outputPage) SetColorPattern(backend.Canvas, fl, fl, matrix.Transform, bool) {
	outputLog.Println("SetColorPattern")
}

func (outputPage) State() backend.GraphicState {
	return outputPage{}
}
