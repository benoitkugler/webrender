package svg

import (
	"io"
	"log"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/pango"
	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/matrix"
)

var outputLog = log.New(io.Discard, "svg_test: ", 0)

type fl = backend.Fl

var _ backend.Canvas = outputPage{}

type outputPage struct{}

func (outputPage) GetRectangle() (left, top, right, bottom fl) {
	outputLog.Println("GetRectangle")
	return 0, 0, 10, 10
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

func (outputPage) SetAlpha(alpha fl, stroke bool) {
	outputLog.Println("SetAlpha")
}

func (outputPage) SetLineWidth(width fl) {
	outputLog.Println("SetLineWidth")
}

func (outputPage) SetDash(dashes []fl, offset fl) {
	outputLog.Println("SetDash")
}

func (outputPage) Fill(evenOdd bool) {
	outputLog.Println("Fill")
}

func (outputPage) FillWithImage(backend.Image, backend.BackgroundImageOptions) {
	outputLog.Println("Fill")
}

func (outputPage) Stroke() {
	outputLog.Println("Stroke")
}

func (outputPage) Transform(mt matrix.Transform) {
	outputLog.Println("Transform")
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

func (outputPage) DrawText(text backend.TextDrawing) {
	outputLog.Println("DrawText", text)
}

func (outputPage) AddFont(pango.Font, []byte) *backend.Font {
	outputLog.Println("AddFont")
	return &backend.Font{Cmap: make(map[fonts.GID][]rune), Extents: make(map[fonts.GID]backend.GlyphExtents)}
}

func (outputPage) DrawRasterImage(img backend.RasterImage, width, height fl) {
	outputLog.Println("DrawRasterImage")
}

func (outputPage) DrawGradient(gradient backend.GradientLayout, width, height fl) {
	outputLog.Println("DrawGradient")
}

func (outputPage) AddOpacityGroup(x, y, width, height fl) backend.Canvas {
	outputLog.Println("AddGroup")
	return outputPage{}
}

func (outputPage) DrawOpacityGroup(opacity fl, group backend.Canvas) {
	outputLog.Println("DrawGroup")
}

func (outputPage) SetStrokeOptions(backend.StrokeOptions) {
	outputLog.Println("SetStrokeOptions")
}
