// Package backend defines a common interface, providing graphics primitives.
//
// It aims at supporting the operations defined in HTML and SVG files in an output-agnostic manner,
// so that various output formats may be generated (GUI canvas, raster image or PDF files for instance).
//
// The types implementing this interface will be used to convert a document.Document to the final output,
// or to draw an svg.SVGImage
package backend

import (
	"fmt"
	"io"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/pango"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/utils"
)

type Fl = utils.Fl

// TextDrawing exposes the positionned text glyphs to draw
// and the associated font, in a backend independent manner
type TextDrawing struct {
	Runs []TextRun

	FontSize Fl
	X, Y     Fl // origin of the text
}

// TextRun is a serie of glyphs with constant font.
type TextRun struct {
	Font   pango.Font
	Glyphs []TextGlyph
}

// TextGlyph stores a glyph and it's position
type TextGlyph struct {
	Glyph    fonts.GID
	Offset   Fl  // normalized by FontSize
	Kerning  int // normalized by FontSize
	XAdvance Fl  // how much to move before drawing
}

type GradientKind struct {
	// Kind is either:
	// 	"solid": Colors is then a one element array and Positions and Coords are empty.
	// 	"linear": Coords is (x0, y0, x1, y1)
	// 			  coordinates of the starting and ending points.
	// 	"radial": Coords is (cx0, cy0, radius0, cx1, cy1, radius1)
	// 			  coordinates of the starting end ending circles
	Kind   string
	Coords [6]Fl
}

type GradientLayout struct {
	// Positions is a list of floats in [0..1].
	// 0 at the starting point, 1 at the ending point.
	Positions []Fl
	Colors    []parser.RGBA

	GradientKind

	// used for ellipses radial gradients. 1 otherwise.
	ScaleY     utils.Fl
	Reapeating bool
}

// RasterImage is an image to be included in the ouput.
type RasterImage struct {
	Content  io.Reader
	MimeType string

	// Rendering is the CSS property for this image.
	Rendering string

	// ID is a unique identifier which permits caching
	// image content when possible.
	ID int
}

// Image groups all possible image format,
// like raster image, svg, or gradients.
type Image interface {
	GetIntrinsicSize(imageResolution, fontSize properties.Float) (width, height, ratio properties.MaybeFloat)

	// Draw shall write the image on the given `canvas`
	Draw(canvas Canvas, textContext text.TextLayoutContext, concreteWidth, concreteHeight Fl, imageRendering string)
}

// StrokeJoinMode type to specify how segments join when stroking.
type StrokeJoinMode uint8

// StrokeJoinMode constants determine how stroke segments bridge the gap at a join
const (
	Miter StrokeJoinMode = iota
	Round
	Bevel
)

func (s StrokeJoinMode) String() string {
	switch s {
	case Round:
		return "Round"
	case Bevel:
		return "Bevel"
	case Miter:
		return "Miter"
	default:
		return fmt.Sprintf("<unknown JoinMode %s>", string(s))
	}
}

// StrokeCapMode defines how to draw caps on the ends of lines
// when stroking.
type StrokeCapMode uint8

const (
	ButtCap StrokeCapMode = iota // default value
	RoundCap
	SquareCap
)

func (c StrokeCapMode) String() string {
	switch c {
	case ButtCap:
		return "ButtCap"
	case SquareCap:
		return "SquareCap"
	case RoundCap:
		return "RoundCap"
	default:
		return fmt.Sprintf("<unknown CapMode %s>", string(c))
	}
}

// StrokeOptions specifies advanced stroking options.
type StrokeOptions struct {
	LineCap  StrokeCapMode
	LineJoin StrokeJoinMode

	// MiterLimit is the miter cutoff value for `Miter`, `Arc`, `Miterclip` and `ArcClip` join modes
	MiterLimit Fl
}

// PaintOp specifies the graphic operation applied to the current path
type PaintOp uint8

const (
	// Clear does not show anything on the output, but reset the current path
	Stroke PaintOp = 1 << iota
	FillEvenOdd
	FillNonZero // mutually exclusive with FillEvenOdd
)

// Canvas represents a 2D surface which is the target of graphic operations.
// It may be used as the final output (like a PDF page or the screen),
// or as intermediate container (see for instance DrawWithOpacity or DrawAsMask)
type Canvas interface {
	// Returns the current canvas rectangle
	GetRectangle() (left, top, right, bottom Fl)

	// OnNewStack save the current graphic stack,
	// execute the given closure, and restore the stack.
	OnNewStack(func())

	// NewGroup creates a new drawing target with the given
	// bounding box. It may be filled by graphic operations
	// before being passed to the `DrawWithOpacity`, `SetColorPattern`
	// and `DrawAsMask` methods.
	NewGroup(x, y, width, height Fl) Canvas

	// DrawWithOpacity draw the given target to the main target, applying the given opacity (in [0,1]).
	DrawWithOpacity(opacity Fl, group Canvas)

	// DrawAsMask inteprets `mask` as an alpha mask
	DrawAsMask(mask Canvas)

	// Establishes a new clip region
	// by intersecting the current clip region
	// with the current path as it would be filled by `Fill`
	// and according to the fill rule given in `evenOdd`.
	//
	// After `Clip`, the current path will be cleared (or closed).
	//
	// The current clip region affects all drawing operations
	// by effectively masking out any changes to the surface
	// that are outside the current clip region.
	//
	// Calling `Clip` can only make the clip region smaller,
	// never larger, but you can call it in the `OnNewStack` closure argument,
	// so that the original clip region is restored afterwards.
	Clip(evenOdd bool)

	// Sets the color which will be used for any subsequent drawing operation.
	//
	// The color and alpha components are
	// floating point numbers in the range 0 to 1.
	// If the values passed in are outside that range, they will be clamped.
	// `stroke` controls whether stroking or filling operations are concerned.
	SetColorRgba(color parser.RGBA, stroke bool)

	// SetColorPattern set the current paint color to the given pattern.
	// A pattern acts as a fill or stroke color, but permits complex textures.
	// It consists of a rectangle, fill with arbitrary content, which will be replicated
	// at fixed horizontal and vertical intervals to fill an area.
	// (contentWidth, contentHeight) define the size of the pattern content.
	// `mat` maps the pattern’s internal coordinate system to the one
	// in which it will painted.
	// `stroke` controls whether stroking or filling operations are concerned.
	SetColorPattern(pattern Canvas, contentWidth, contentHeight Fl, mat matrix.Transform, stroke bool)

	// Sets the current line width to be used by `Stroke`.
	// The line width value specifies the diameter of a pen
	// that is circular in user space,
	// (though device-space pen may be an ellipse in general
	// due to scaling / shear / rotation of the CTM).
	SetLineWidth(width Fl)

	// Sets the dash pattern to be used by `Stroke`.
	// A dash pattern is specified by dashes, a list of positive values.
	// Each value provides the length of alternate "on" and "off"
	// portions of the stroke.
	// `offset` specifies a non negative offset into the pattern
	// at which the stroke begins.
	//
	// Each "on" segment will have caps applied
	// as if the segment were a separate sub-path.
	// In particular, it is valid to use an "on" length of 0
	// with `RoundCap` or `SquareCap`
	// in order to distribute dots or squares along a path.
	//
	// If `dashes` is empty dashing is disabled.
	// If it is of length 1 a symmetric pattern is assumed
	// with alternating on and off portions of the size specified
	// by the single value.
	SetDash(dashes []Fl, offset Fl)

	// SetStrokeOptions sets additionnal options to be used when stroking
	// (in addition to SetLineWidth and SetDash)
	SetStrokeOptions(StrokeOptions)

	// Paint actually shows the current path on the target,
	// either stroking, filling or doing both, according to `op`.
	// The result of the operation depends on the current fill and
	// stroke settings.
	// After this call, the current path will be cleared.
	Paint(op PaintOp)

	// GetTransform returns the current transformation matrix (CTM).
	GetTransform() matrix.Transform

	// Modifies the current transformation matrix (CTM)
	// by applying `mt` as an additional transformation.
	// The new transformation of user space takes place
	// after any existing transformation.
	Transform(mt matrix.Transform)

	// Adds a rectangle of the given size to the current path,
	// at position ``(x, y)`` in user-space coordinates.
	// (X,Y) coordinates are the top left corner of the rectangle.
	// Note that this method may be expressed using MoveTo and LineTo,
	// but may be implemented more efficiently.
	Rectangle(x Fl, y Fl, width Fl, height Fl)

	// Begin a new sub-path.
	// After this call the current point will be (x, y).
	MoveTo(x Fl, y Fl)

	// Adds a line to the path from the current point
	// to position (x, y) in user-space coordinates.
	// After this call the current point will be (x, y).
	// A current point must be defined before using this method.
	LineTo(x Fl, y Fl)

	// Add cubic Bézier curve to current path.
	// The curve shall extend to (x3, y3) using (x1, y1) and (x2,
	// y2) as the Bézier control points.
	CubicTo(x1, y1, x2, y2, x3, y3 Fl)

	// ClosePath add a straight line to the beginning of
	// the current sub-path (specified by MoveTo)
	// It is somewhat equivalent to adding a LineTo instruction,
	// but some backends may optimize the corner rendering, applying line join style.
	ClosePath()

	// AddFont register a new font to be used in the output and return
	// an object used to store associated metadata.
	// This method will be called several times with the same `font` argument,
	// so caching is advised.
	AddFont(font pango.Font, content []byte) *Font

	// DrawText draws the given text using the current fill color.
	// The fonts of the runs have been registred with `AddFont`.
	DrawText(TextDrawing)

	// DrawRasterImage draws the given image at the current point, with the given dimensions.
	// Typical format for image.Content are PNG, JPEG, GIF.
	DrawRasterImage(image RasterImage, width, height Fl)

	// DrawGradient draws the given gradient at the current point.
	// Solid gradient are already handled, meaning that only linear and radial
	// must be taken care of.
	DrawGradient(gradient GradientLayout, width, height Fl)
}
