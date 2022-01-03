package document

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"text/template"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/pango"
	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/html/layout/text"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/matrix"

	"github.com/benoitkugler/webrender/html/layout"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/images"
	"github.com/benoitkugler/webrender/utils"

	bo "github.com/benoitkugler/webrender/html/boxes"
)

// Take an "after layout" box tree and draw it onto a cairo context.

var sides = [4]string{"top", "right", "bottom", "left"}

const (
	pi = math.Pi

	headerSVG = `
	<svg height="{{ .Height }}" width="{{ .Width }}"
		 fill="transparent" stroke="black" stroke-width="1"
		 xmlns="http://www.w3.org/2000/svg"
		 xmlns:xlink="http://www.w3.org/1999/xlink">
  	`

	crop = `
  <!-- horizontal top left -->
  <path d="M0,{{ .Bleed.Top }} h{{ .HalfBleed.Left }}" />
  <!-- horizontal top right -->
  <path d="M0,{{ .Bleed.Top }} h{{ .HalfBleed.Right }}"
        transform="translate({{ .Width }},0) scale(-1,1)" />
  <!-- horizontal bottom right -->
  <path d="M0,{{ .Bleed.Bottom }} h{{ .HalfBleed.Right }}"
        transform="translate({{ .Width }},{{ .Height }}) scale(-1,-1)" />
  <!-- horizontal bottom left -->
  <path d="M0,{{ .Bleed.Bottom }} h{{ .HalfBleed.Left }}"
        transform="translate(0,{{ .Height }}) scale(1,-1)" />
  <!-- vertical top left -->
  <path d="M{{ .Bleed.Left }},0 v{{ .HalfBleed.Top }}" />
  <!-- vertical bottom right -->
  <path d="M{{ .Bleed.Right }},0 v{{ .HalfBleed.Bottom }}"
        transform="translate({{ .Width }},{{ .Height }}) scale(-1,-1)" />
  <!-- vertical bottom left -->
  <path d="M{{ .Bleed.Left }},0 v{{ .HalfBleed.Bottom }}"
        transform="translate(0,{{ .Height }}) scale(1,-1)" />
  <!-- vertical top right -->
  <path d="M{{ .Bleed.Right }},0 v{{ .HalfBleed.Top }}"
        transform="translate({{ .Width }},0) scale(-1,1)" />
`
	cross = `
  <!-- top -->
  <circle r="{{ .HalfBleed.Top }}"
          transform="scale(0.5)
                     translate({{ .Width }},{{ .HalfBleed.Top }}) scale(0.5)" />
  <path d="M-{{ .HalfBleed.Top }},{{ .HalfBleed.Top }} h{{ .Bleed.Top }}
           M0,0 v{{ .Bleed.Top }}"
        transform="scale(0.5) translate({{ .Width }},0)" />
  <!-- bottom -->
  <circle r="{{ .HalfBleed.Bottom }}"
          transform="translate(0,{{ .Height }}) scale(0.5)
                     translate({{ .Width }},-{{ .HalfBleed.Bottom }}) scale(0.5)" />
  <path d="M-{{ .HalfBleed.Bottom }},-{{ .HalfBleed.Bottom }} h{{ .Bleed.Bottom }}
           M0,0 v-{{ .Bleed.Bottom }}"
        transform="translate(0,{{ .Height }}) scale(0.5) translate({{ .Width }},0)" />
  <!-- left -->
  <circle r="{{ .HalfBleed.Left }}"
          transform="scale(0.5)
                     translate({{ .HalfBleed.Left }},{{ .Height }}) scale(0.5)" />
  <path d="M{{ .HalfBleed.Left }},-{{ .HalfBleed.Left }} v{{ .Bleed.Left }}
           M0,0 h{{ .Bleed.Left }}"
        transform="scale(0.5) translate(0,{{ .Height }})" />
  <!-- right -->
  <circle r="{{ .HalfBleed.Right }}"
          transform="translate({{ .Width }},0) scale(0.5)
                     translate(-{{ .HalfBleed.Right }},{{ .Height }}) scale(0.5)" />
  <path d="M-{{ .HalfBleed.Right }},-{{ .HalfBleed.Right }} v{{ .Bleed.Right }}
           M0,0 h-{{ .Bleed.Right }}"
        transform="translate({{ .Width }},0)
                   scale(0.5) translate(0,{{ .Height }})" />
`
)

type Bleed struct {
	Top, Bottom, Left, Right fl
}

// FIXME: check gofpdf support for SVG
type svgArgs struct {
	Width, Height    fl
	Bleed, HalfBleed Bleed
}

// Transform a HSV color to a RGB color.
func hsv2rgb(hue, saturation, value fl) (r, g, b fl) {
	c := value * saturation
	x := c * fl(1-math.Abs(float64(utils.FloatModulo(hue/60, 2))-1))
	m := value - c
	switch {
	case 0 <= hue && hue < 60:
		return c + m, x + m, m
	case 60 <= hue && hue < 120:
		return x + m, c + m, m
	case 120 <= hue && hue < 180:
		return m, c + m, x + m
	case 180 <= hue && hue < 240:
		return m, x + m, c + m
	case 240 <= hue && hue < 300:
		return x + m, m, c + m
	case 300 <= hue && hue < 360:
		return c + m, m, x + m
	default:
		logger.WarningLogger.Printf("invalid hue %f", hue)
		return 0, 0, 0
	}
}

// Transform a RGB color to a HSV color.
func rgb2hsv(red, green, blue fl) (h, s, c fl) {
	cmax := utils.Maxs(red, green, blue)
	cmin := utils.Mins(red, green, blue)
	delta := cmax - cmin
	var hue fl
	if delta == 0 {
		hue = 0
	} else if cmax == red {
		hue = 60 * utils.FloatModulo((green-blue)/delta, 6)
	} else if cmax == green {
		hue = 60 * ((blue-red)/delta + 2)
	} else if cmax == blue {
		hue = 60 * ((red-green)/delta + 4)
	}
	var saturation fl
	if delta != 0 {
		saturation = delta / cmax
	}
	return hue, saturation, cmax
}

// Return a darker color.
func darken(color Color) Color {
	hue, saturation, value := rgb2hsv(color.R, color.G, color.B)
	value /= 1.5
	saturation /= 1.25
	r, g, b := hsv2rgb(hue, saturation, value)
	return Color{R: r, G: g, B: b, A: color.A}
}

// Return a lighter color.
func lighten(color Color) Color {
	hue, saturation, value := rgb2hsv(color.R, color.G, color.B)
	value = 1 - (1-value)/1.5
	if saturation != 0 {
		saturation = 1 - (1-saturation)/1.25
	}
	r, g, b := hsv2rgb(hue, saturation, value)
	return Color{R: r, G: g, B: b, A: color.A}
}

type drawContext struct {
	dst   backend.Canvas
	fonts *text.FontConfiguration
}

// Draw the given PageBox.
func (ctx drawContext) drawPage(page *bo.PageBox) {
	bleed := Bleed{
		Top:    fl(page.Style.GetBleedTop().Value),
		Bottom: fl(page.Style.GetBleedBottom().Value),
		Left:   fl(page.Style.GetBleedLeft().Value),
		Right:  fl(page.Style.GetBleedRight().Value),
	}
	marks := page.Style.GetMarks()
	stackingContext := NewStackingContextFromPage(page)
	ctx.drawBackground(stackingContext.box.Box().Background, false, bleed, marks)
	ctx.drawBackground(page.CanvasBackground, false, Bleed{}, pr.Marks{})
	ctx.drawBorder(page)
	ctx.drawStackingContext(stackingContext)
}

func (ctx drawContext) drawBoxBackgroundAndBorder(box Box) error {
	ctx.drawBackground(box.Box().Background, true, Bleed{}, pr.Marks{})
	if box_, ok := box.(bo.TableBoxITF); ok {
		box := box_.Table()
		ctx.drawTableBackgrounds(box_)
		if box.Style.GetBorderCollapse() == "separate" {
			ctx.drawBorder(box)
			for _, rowGroup := range box.Children {
				for _, row := range rowGroup.Box().Children {
					for _, cell := range row.Box().Children {
						if cell.Box().Style.GetEmptyCells() == "show" || !cell.Box().Empty {
							ctx.drawBorder(cell)
						}
					}
				}
			}
		} else {
			ctx.drawCollapsedBorders(box)
		}
	} else {
		ctx.drawBorder(box)
	}
	return nil
}

// Draw a ``stackingContext`` on ``context``.
func (ctx drawContext) drawStackingContext(stackingContext StackingContext) {
	// See http://www.w3.org/TR/CSS2/zindex.html
	ctx.dst.OnNewStack(func() {
		box_ := stackingContext.box
		box := box_.Box()

		// apply the viewport_overflow to the html box, see #35
		if box.IsForRootElement && (stackingContext.page.Style.GetOverflow() != "visible") {
			roundedBoxPath(
				ctx.dst, stackingContext.page.RoundedPaddingBox())
			ctx.dst.Clip(false)
		}

		if clips := box.Style.GetClip(); box.IsAbsolutelyPositioned() && len(clips) != 0 {
			top, right, bottom, left := clips[0], clips[1], clips[2], clips[3]
			if top.String == "auto" {
				top.Value = 0
			}
			if right.String == "auto" {
				right.Value = 0
			}
			if bottom.String == "auto" {
				bottom.Value = box.BorderHeight()
			}
			if left.String == "auto" {
				left.Value = box.BorderWidth()
			}
			ctx.dst.Rectangle(
				fl(box.BorderBoxX()+right.Value),
				fl(box.BorderBoxY()+top.Value),
				fl(left.Value-right.Value),
				fl(bottom.Value-top.Value),
			)
			ctx.dst.Clip(false)
		}

		originalDst := ctx.dst
		opacity := fl(box.Style.GetOpacity())
		if opacity < 1 { // we draw all the following to a separate group
			ctx.dst = ctx.dst.AddOpacityGroup(pr.Fl(box.BorderBoxX()), pr.Fl(box.BorderBoxY()),
				pr.Fl(box.BorderWidth()), pr.Fl(box.BorderHeight()))
		}

		if mat, ok := getMatrix(box_); ok {
			if mat.Determinant() != 0 {
				ctx.dst.Transform(mat)
			} else {
				logger.WarningLogger.Printf("non invertible transformation matrix %v\n", mat)
				return
			}
		}

		// Point 1 is done in drawPage

		// Point 2
		if bo.BlockBoxT.IsInstance(box_) || bo.MarginBoxT.IsInstance(box_) ||
			bo.InlineBlockBoxT.IsInstance(box_) || bo.TableCellBoxT.IsInstance(box_) ||
			bo.FlexContainerBoxT.IsInstance(box_) {
			// The canvas background was removed by setCanvasBackground
			ctx.drawBoxBackgroundAndBorder(box_)
		}

		ctx.dst.OnNewStack(func() {
			// dont clip the PageBox, see #35
			if box.Style.GetOverflow() != "visible" && !bo.PageBoxT.IsInstance(box_) {
				// Only clip the content and the children:
				// - the background is already clipped
				// - the border must *not* be clipped
				roundedBoxPath(ctx.dst, box.RoundedPaddingBox())
				ctx.dst.Clip(false)
			}

			// Point 3
			for _, childContext := range stackingContext.negativeZContexts {
				ctx.drawStackingContext(childContext)
			}

			// Point 4
			for _, block := range stackingContext.blockLevelBoxes {
				ctx.drawBoxBackgroundAndBorder(block)
			}

			// Point 5
			for _, childContext := range stackingContext.floatContexts {
				ctx.drawStackingContext(childContext)
			}

			// Point 6
			if bo.InlineBoxT.IsInstance(box_) {
				ctx.drawInlineLevel(stackingContext.page, box_, 0, "clip", pr.NamedString{Name: "none"})
			}

			// Point 7
			for _, block := range append([]Box{box_}, stackingContext.blocksAndCells...) {
				if blockRep, ok := block.(bo.ReplacedBoxITF); ok {
					ctx.drawReplacedbox(blockRep)
				} else {
					for _, child := range block.Box().Children {
						if bo.LineBoxT.IsInstance(child) {
							ctx.drawInlineLevel(stackingContext.page, child, 0, "clip", pr.NamedString{Name: "none"})
						}
					}
				}
			}

			// Point 8
			for _, childContext := range stackingContext.zeroZContexts {
				ctx.drawStackingContext(childContext)
			}

			// Point 9
			for _, childContext := range stackingContext.positiveZContexts {
				ctx.drawStackingContext(childContext)
			}

			return
		})

		// Point 10
		ctx.drawOutlines(box_)

		if opacity < 1 {
			group := ctx.dst
			ctx.dst = originalDst
			ctx.dst.OnNewStack(func() {
				ctx.dst.DrawOpacityGroup(opacity, group)
			})
		}
	})
}

// Draw the path of the border radius box.
// ``widths`` is a tuple of the inner widths (top, right, bottom, left) from
// the border box. Radii are adjusted from these values. Default is (0, 0, 0,
// 0).
func roundedBoxPath(context backend.Canvas, radii bo.RoundedBox) {
	x, y, w, h, tl, tr, br, bl := pr.Fl(radii.X), pr.Fl(radii.Y), pr.Fl(radii.Width), pr.Fl(radii.Height), radii.TopLeft, radii.TopRight, radii.BottomRight, radii.BottomLeft
	if (tl[0] == 0 || tl[1] == 0) && (tr[0] == 0 || tr[1] == 0) &&
		(br[0] == 0 || br[1] == 0) && (bl[0] == 0 || bl[1] == 0) {
		// No radius, draw a rectangle
		context.Rectangle(x, y, w, h)
		return
	}

	var r pr.Fl = 0.45

	context.MoveTo(x+pr.Fl(tl[0]), y)
	context.LineTo(x+w-pr.Fl(tr[0]), y)
	context.CubicTo(
		x+w-pr.Fl(tr[0])*r, y, x+w, y+pr.Fl(tr[1])*r, x+w, y+pr.Fl(tr[1]))
	context.LineTo(x+w, y+h-pr.Fl(br[1]))
	context.CubicTo(
		x+w, y+h-pr.Fl(br[1])*r, x+w-pr.Fl(br[0])*r, y+h, x+w-pr.Fl(br[0]),
		y+h)
	context.LineTo(x+pr.Fl(bl[0]), y+h)
	context.CubicTo(
		x+pr.Fl(bl[0])*r, y+h, x, y+h-pr.Fl(bl[1])*r, x, y+h-pr.Fl(bl[1]))
	context.LineTo(x, y+pr.Fl(tl[1]))
	context.CubicTo(
		x, y+pr.Fl(tl[1])*r, x+pr.Fl(tl[0])*r, y, x+pr.Fl(tl[0]), y)
}

func formatSVG(svg string, data svgArgs) (string, error) {
	tmp, err := template.New("svg").Parse(svg)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := tmp.Execute(&b, data); err != nil {
		return "", fmt.Errorf("unexpected template error : %s", err)
	}
	return b.String(), nil
}

func reversed(in []bo.BackgroundLayer) []bo.BackgroundLayer {
	N := len(in)
	out := make([]bo.BackgroundLayer, N)
	for i, v := range in {
		out[N-1-i] = v
	}
	return out
}

func (ctx drawContext) drawBackgroundDefaut(bg *bo.Background) {
	ctx.drawBackground(bg, true, Bleed{}, pr.Marks{})
}

// Draw the background color and image
// If ``clipBox`` is set to ``false``, the background is not clipped to the
// border box of the background, but only to the painting area
// clipBox=true bleed=nil marks=()
func (ctx drawContext) drawBackground(bg *bo.Background, clipBox bool, bleed Bleed, marks pr.Marks) {
	if bg == nil {
		return
	}

	ctx.dst.OnNewStack(func() {
		if clipBox {
			for _, box := range bg.Layers[len(bg.Layers)-1].ClippedBoxes {
				roundedBoxPath(ctx.dst, box)
			}
			ctx.dst.Clip(false)
		}

		// Background color
		if bg.Color.A > 0 {
			ctx.dst.OnNewStack(func() {
				paintingArea := bg.Layers[len(bg.Layers)-1].PaintingArea
				if !paintingArea.IsNone() {
					ptx, pty, ptw, pth := paintingArea.Unpack()
					if (bleed != Bleed{}) {
						// Painting area is the PDF BleedBox
						ptx -= bleed.Left
						pty -= bleed.Top
						ptw += bleed.Left + bleed.Right
						pth += bleed.Top + bleed.Bottom
					}
					ctx.dst.Rectangle(ptx, pty, ptw, pth)
					ctx.dst.Clip(false)
				}
				ctx.dst.Rectangle(ctx.dst.GetRectangle())
				ctx.dst.SetColorRgba(bg.Color, false)
				ctx.dst.Fill(false)
			})
		}

		if (bleed != Bleed{}) && !marks.IsNone() {
			x, y, width, height := bg.Layers[len(bg.Layers)-1].PaintingArea.Unpack()
			x -= bleed.Left
			y -= bleed.Top
			width += bleed.Left + bleed.Right
			height += bleed.Top + bleed.Bottom
			svg := headerSVG
			if marks.Crop {
				svg += crop
			}
			if marks.Cross {
				svg += cross
			}
			svg += "</svg>"
			halfBleed := Bleed{
				Top:    bleed.Top * 0.5,
				Bottom: bleed.Bottom * 0.5,
				Left:   bleed.Left * 0.5,
				Right:  bleed.Right * 0.5,
			}
			svg, err := formatSVG(svg, svgArgs{Width: width, Height: height, Bleed: bleed, HalfBleed: halfBleed})
			if err != nil {
				logger.WarningLogger.Println(err)
				return
			}
			image, err := images.NewSVGImage(strings.NewReader(svg), "", nil)
			if err != nil {
				logger.WarningLogger.Println(err)
				return
			}

			// Painting area is the PDF media box
			size := [2]pr.Float{pr.Float(width), pr.Float(height)}
			position := bo.Position{Point: bo.MaybePoint{pr.Float(x), pr.Float(y)}}
			repeat := bo.Repeat{Reps: [2]string{"no-repeat", "no-repeat"}}
			unbounded := true
			paintingArea := pr.Rectangle{pr.Float(x), pr.Float(y), pr.Float(width), pr.Float(height)}
			positioningArea := pr.Rectangle{0, 0, pr.Float(width), pr.Float(height)}
			layer := bo.BackgroundLayer{
				Image: image, Size: size, Position: position, Repeat: repeat, Unbounded: unbounded,
				PaintingArea: paintingArea, PositioningArea: positioningArea,
			}
			bg.Layers = append([]bo.BackgroundLayer{layer}, bg.Layers...)
		}
		// Paint in reversed order: first layer is "closest" to the viewer.
		for _, layer := range reversed(bg.Layers) {
			ctx.drawBackgroundImage(layer, bg.ImageRendering)
		}
	})
}

// Draw the background color and image of the table children.
func (ctx drawContext) drawTableBackgrounds(table_ bo.TableBoxITF) {
	table := table_.Table()
	for _, columnGroup := range table.ColumnGroups {
		ctx.drawBackgroundDefaut(columnGroup.Box().Background)

		for _, column := range columnGroup.Box().Children {
			ctx.drawBackgroundDefaut(column.Box().Background)
		}
	}
	for _, rowGroup := range table.Children {
		ctx.drawBackgroundDefaut(rowGroup.Box().Background)

		for _, row := range rowGroup.Box().Children {
			ctx.drawBackgroundDefaut(row.Box().Background)

			for _, cell := range row.Box().Children {
				cell := cell.Box()
				if table.Style.GetBorderCollapse() == "collapse" ||
					cell.Style.GetEmptyCells() == "show" || !cell.Empty {
					ctx.drawBackgroundDefaut(cell.Background)
				}
			}
		}
	}
}

func (ctx drawContext) drawBackgroundImage(layer bo.BackgroundLayer, imageRendering pr.String) {
	if layer.Image == nil || layer.Size[0] == 0 || layer.Size[1] == 0 {
		return
	}

	paintingX, paintingY, paintingWidth, paintingHeight := layer.PaintingArea.Unpack()
	positioningX, positioningY, positioningWidth, positioningHeight := layer.PositioningArea.Unpack()
	positionX, positionY := layer.Position.Point[0], layer.Position.Point[1]
	repeatX, repeatY := layer.Repeat.Reps[0], layer.Repeat.Reps[1]
	imageWidth, imageHeight := pr.Fl(layer.Size[0]), pr.Fl(layer.Size[1])
	var repeatWidth, repeatHeight pr.Fl
	switch repeatX {
	case "no-repeat":
		// We want at least the whole imageWidth drawn on subSurface, but we
		// want to be sure it will not be repeated on the paintingWidth.
		repeatWidth = utils.Maxs(imageWidth, paintingWidth)
	case "repeat", "round":
		// We repeat the image each imageWidth.
		repeatWidth = imageWidth
	case "space":
		nRepeats := pr.Fl(math.Floor(float64(positioningWidth / imageWidth)))
		if nRepeats >= 2 {
			// The repeat width is the whole positioning width with one image
			// removed, divided by (the number of repeated images - 1). This
			// way, we get the width of one image + one space. We ignore
			// background-position for this dimension.
			repeatWidth = (positioningWidth - imageWidth) / (nRepeats - 1)
			positionX = pr.Float(0)
		} else {
			// We don't repeat the image.
			repeatWidth = imageWidth
		}
	default:
		panic(fmt.Sprintf("unexpected repeatX %s", repeatX))
	}

	// Comments above apply here too.
	switch repeatY {
	case "no-repeat":
		repeatHeight = utils.Maxs(imageHeight, paintingHeight)
	case "repeat", "round":
		repeatHeight = imageHeight
	case "space":
		nRepeats := fl(math.Floor(float64(positioningHeight / imageHeight)))
		if nRepeats >= 2 {
			repeatHeight = (positioningHeight - imageHeight) / (nRepeats - 1)
			positionY = pr.Float(0)
		} else {
			repeatHeight = imageHeight
		}
	default:
		panic(fmt.Sprintf("unexpected repeatY %s", repeatY))
	}

	if repeatX == "no-repeat" && repeatY == "no-repeat" {
		// PDF patterns always repeat, use a big number to hide repetition
		_, _, w, h := ctx.dst.GetRectangle()
		repeatWidth = 2 * w
		repeatHeight = 2 * h
	}

	options := backend.BackgroundImageOptions{
		ImageWidth:   imageWidth,
		ImageHeight:  imageHeight,
		RepeatWidth:  repeatWidth,
		RepeatHeight: repeatHeight,
		X:            pr.Fl(positionX.V()) + positioningX,
		Y:            pr.Fl(positionY.V()) + positioningY,
		Rendering:    string(imageRendering),
	}
	ctx.dst.OnNewStack(func() {
		if layer.Unbounded {
			x1, y1, x2, y2 := ctx.dst.GetRectangle()
			ctx.dst.Rectangle(x1, y1, x2-x1, y2-y1)
		} else {
			ctx.dst.Rectangle(paintingX, paintingY, paintingWidth, paintingHeight)
		}
		ctx.dst.FillWithImage(layer.Image, options)
	})
}

// Increment X and Y coordinates by the given offsets.
func xyOffset(x, y, offsetX, offsetY, offset float64) (float64, float64) {
	return x + offsetX*offset, y + offsetY*offset
}

func styledColor(style pr.String, color Color, side string) []Color {
	if style == "inset" || style == "outset" {
		doLighten := (side == "top" || side == "left") != (style == "inset")
		if doLighten {
			return []Color{lighten(color)}
		}
		return []Color{darken(color)}
	} else if style == "ridge" || style == "groove" {
		if (side == "top" || side == "left") != (style == "ridge") {
			return []Color{lighten(color), darken(color)}
		} else {
			return []Color{darken(color), lighten(color)}
		}
	}
	return []Color{color}
}

// Draw the box border
func (ctx drawContext) drawBorder(box_ Box) {
	// We need a plan to draw beautiful borders, and that's difficult, no need
	// to lie. Let's try to find the cases that we can handle in a smart way.
	box := box_.Box()

	// Draw column borders.
	drawColumnBorder := func() {
		columns := bo.BlockContainerBoxT.IsInstance(box_) && (box.Style.GetColumnWidth().String != "auto" || box.Style.GetColumnCount().String != "auto")
		if crw := box.Style.GetColumnRuleWidth(); columns && !crw.IsNone() {
			borderWidths := pr.Rectangle{0, 0, 0, crw.Value}
			for _, child := range box.Children[1:] {
				ctx.dst.OnNewStack(func() {
					positionX := child.Box().PositionX - (crw.Value+
						box.Style.GetColumnGap().Value)/2
					borderBox := pr.Rectangle{
						positionX, child.Box().PositionY,
						crw.Value, box.Height.V(),
					}
					clipBorderSegment(ctx.dst, box.Style.GetColumnRuleStyle(),
						fl(crw.Value), "left", borderBox, &borderWidths, nil)
					ctx.drawRectBorder(borderBox, borderWidths,
						box.Style.GetColumnRuleStyle(), styledColor(
							box.Style.GetColumnRuleStyle(),
							tree.ResolveColor(box.Style, "column_rule_color").RGBA, "left"))
				})
			}
		}
	}

	// The box is hidden, easy.
	if box.Style.GetVisibility() != "visible" {
		drawColumnBorder()
		return
	}

	widths := pr.Rectangle{box.BorderTopWidth.V(), box.BorderRightWidth.V(), box.BorderBottomWidth.V(), box.BorderLeftWidth.V()}

	// No border, return early.
	if widths.IsNone() {
		drawColumnBorder()
		return
	}
	var (
		colors    [4]Color
		colorsSet = map[Color]bool{}
		styles    [4]pr.String
		stylesSet = utils.NewSet()
	)
	for i, side := range sides {
		colors[i] = tree.ResolveColor(box.Style, fmt.Sprintf("border_%s_color", side)).RGBA
		colorsSet[colors[i]] = true
		if colors[i].A != 0 {
			styles[i] = box.Style.Get(fmt.Sprintf("border_%s_style", side)).(pr.String)
		}
		stylesSet.Add(string(styles[i]))
	}

	// The 4 sides are solid or double, and they have the same color. Oh yeah!
	// We can draw them so easily!
	if len(stylesSet) == 1 && (stylesSet.Has("solid") || stylesSet.Has("double")) && len(colorsSet) == 1 {
		ctx.drawRoundedBorder(box, styles[0], []Color{colors[0]})
		drawColumnBorder()
		return
	}

	// We"re not smart enough to find a good way to draw the borders :/. We must
	// draw them side by side.
	for i, side := range sides {
		width, color, style := widths[i], colors[i], styles[i]
		if width == 0 || color.IsNone() {
			continue
		}
		ctx.dst.OnNewStack(func() {
			rb := box.RoundedBorderBox()
			roundedBox := pr.Rectangle{rb.X, rb.Y, rb.Width, rb.Height}
			radii := [4]bo.Point{rb.TopLeft, rb.TopRight, rb.BottomRight, rb.BottomLeft}
			clipBorderSegment(ctx.dst, style, fl(width), side,
				roundedBox, &widths, &radii)
			ctx.drawRoundedBorder(box, style, styledColor(style, color, side))
		})
	}

	drawColumnBorder()
}

// Clip one segment of box border (border_widths=nil, radii=nil).
// The strategy is to remove the zones not needed because of the style or the
// side before painting.
func clipBorderSegment(context backend.Canvas, style pr.String, width fl, side string,
	borderBox pr.Rectangle, borderWidths *pr.Rectangle, radii *[4]bo.Point) {

	bbx, bby, bbw, bbh := borderBox.Unpack()
	var tlh, tlv, trh, trv, brh, brv, blh, blv fl
	if radii != nil {
		tlh, tlv, trh, trv, brh, brv, blh, blv = fl((*radii)[0][0]), fl((*radii)[0][1]), fl((*radii)[1][0]), fl((*radii)[1][1]), fl((*radii)[2][0]), fl((*radii)[2][1]), fl((*radii)[3][0]), fl((*radii)[3][1])
	}
	bt, br, bb, bl := width, width, width, width
	if borderWidths != nil {
		bt, br, bb, bl = borderWidths.Unpack()
	}

	// Get the point use for border transition.
	// The extra boolean returned is ``true`` if the point is in the padding
	// box (ie. the padding box is rounded).
	// This point is not specified. We must be sure to be inside the rounded
	// padding box, and in the zone defined in the "transition zone" allowed
	// by the specification. We chose the corner of the transition zone. It"s
	// easy to get and gives quite good results, but it seems to be different
	// from what other browsers do.
	transitionPoint := func(x1, y1, x2, y2 fl) (fl, fl, bool) {
		if math.Abs(float64(x1)) > math.Abs(float64(x2)) && math.Abs(float64(y1)) > math.Abs(float64(y2)) {
			return x1, y1, true
		}
		return x2, y2, false
	}

	// Return the length of the half of one ellipsis corner.

	// Inspired by [Ramanujan, S., "Modular Equations and Approximations to
	// pi" Quart. J. Pure. Appl. Math., vol. 45 (1913-1914), pp. 350-372],
	// wonderfully explained by Dr Rob.

	// http://mathforum.org/dr.math/faq/formulas/
	cornerHalfLength := func(a, b fl) fl {
		x := (a - b) / (a + b)
		return pi / 8 * (a + b) * (1 + 3*x*x/(10+fl(math.Sqrt(float64(4-3*x*x)))))
	}
	var (
		px1, px2, py1, py2, way, angle, mainOffset fl
		rounded1, rounded2                         bool
	)
	if side == "top" {
		px1, py1, rounded1 = transitionPoint(tlh, tlv, bl, bt)
		px2, py2, rounded2 = transitionPoint(-trh, trv, -br, bt)
		width = bt
		way = 1
		angle = 1
		mainOffset = bby
	} else if side == "right" {
		px1, py1, rounded1 = transitionPoint(-trh, trv, -br, bt)
		px2, py2, rounded2 = transitionPoint(-brh, -brv, -br, -bb)
		width = br
		way = 1
		angle = 2
		mainOffset = bbx + bbw
	} else if side == "bottom" {
		px1, py1, rounded1 = transitionPoint(blh, -blv, bl, -bb)
		px2, py2, rounded2 = transitionPoint(-brh, -brv, -br, -bb)
		width = bb
		way = -1
		angle = 3
		mainOffset = bby + bbh
	} else if side == "left" {
		px1, py1, rounded1 = transitionPoint(tlh, tlv, bl, bt)
		px2, py2, rounded2 = transitionPoint(blh, -blv, bl, -bb)
		width = bl
		way = -1
		angle = 4
		mainOffset = bbx
	}

	var a1, b1, a2, b2, lineLength, length fl
	if side == "top" || side == "bottom" {
		a1, b1 = px1-bl/2, way*py1-width/2
		a2, b2 = -px2-br/2, way*py2-width/2
		lineLength = bbw - px1 + px2
		length = bbw
		context.MoveTo(bbx+bbw, mainOffset)
		context.LineTo(bbx, mainOffset)
		context.LineTo(bbx+px1, mainOffset+py1)
		context.LineTo(bbx+bbw+px2, mainOffset+py2)
	} else if side == "left" || side == "right" {
		a1, b1 = -way*px1-width/2, py1-bt/2
		a2, b2 = -way*px2-width/2, -py2-bb/2
		lineLength = bbh - py1 + py2
		length = bbh
		context.MoveTo(mainOffset, bby+bbh)
		context.LineTo(mainOffset, bby)
		context.LineTo(mainOffset+px1, bby+py1)
		context.LineTo(mainOffset+px2, bby+bbh+py2)
	}

	if style == "dotted" || style == "dashed" {
		dash := 3 * width
		if style == "dotted" {
			dash = width
		}
		if rounded1 || rounded2 {
			// At least one of the two corners is rounded
			chl1 := cornerHalfLength(a1, b1)
			chl2 := cornerHalfLength(a2, b2)
			length = lineLength + chl1 + chl2
			dashLength := fl(math.Round(float64(length / dash)))
			if rounded1 && rounded2 {
				// 2x dashes
				dash = length / (dashLength + utils.FloatModulo(dashLength, 2))
			} else {
				// 2x - 1/2 dashes
				dash = length / (dashLength + utils.FloatModulo(dashLength, 2) - 0.5)
			}
			dashes1 := int(math.Ceil(float64((chl1 - dash/2) / dash)))
			dashes2 := int(math.Ceil(float64((chl2 - dash/2) / dash)))
			line := int(math.Floor(float64(lineLength / dash)))

			drawDots := func(dashes, line int, way, x, y, px, py, chl fl) (int, fl) {
				if dashes == 0 {
					return line + 1, 0
				}
				var (
					hasBroken              bool
					offset, angle1, angle2 fl
				)
				for i_ := 0; i_ < dashes; i_ += 2 {
					i := fl(i_) + 0.5 // half dash
					angle1 = ((2*angle - way) + i*way*dash/chl) / 4 * pi

					fn := utils.MaxF
					if way > 0 {
						fn = utils.MinF
					}
					angle2 = fn(
						((2*angle-way)+(i+1)*way*dash/chl)/4*pi,
						angle*pi/2,
					)
					if side == "top" || side == "bottom" {
						context.MoveTo(x+px, mainOffset+py)
						context.LineTo(x+px-way*px*1/fl(math.Tan(float64(angle2))), mainOffset)
						context.LineTo(x+px-way*px*1/fl(math.Tan(float64(angle1))), mainOffset)
					} else if side == "left" || side == "right" {
						context.MoveTo(mainOffset+px, y+py)
						context.LineTo(mainOffset, y+py+way*py*fl(math.Tan(float64(angle2))))
						context.LineTo(mainOffset, y+py+way*py*fl(math.Tan(float64(angle1))))
					}
					if angle2 == angle*pi/2 {
						offset = (angle1 - angle2) / ((((2*angle - way) + (i+1)*way*dash/chl) /
							4 * pi) - angle1)
						line += 1
						hasBroken = true
						break
					}
				}
				if !hasBroken {
					offset = 1 - (angle*pi/2-angle2)/(angle2-angle1)
				}
				return line, offset
			}
			var offset fl
			line, offset = drawDots(dashes1, line, way, bbx, bby, px1, py1, chl1)
			line, _ = drawDots(dashes2, line, -way, bbx+bbw, bby+bbh, px2, py2, chl2)

			if lineLength > 1e-6 {
				for i_ := 0; i_ < line; i_ += 2 {
					i := fl(i_) + offset
					var x1, x2, y1, y2 fl
					if side == "top" || side == "bottom" {
						x1 = utils.MaxF(bbx+px1+i*dash, bbx+px1)
						x2 = utils.MinF(bbx+px1+(i+1)*dash, bbx+bbw+px2)
						y1 = mainOffset
						if way < 0 {
							y1 -= width
						}
						y2 = y1 + width
					} else if side == "left" || side == "right" {
						y1 = utils.MaxF(bby+py1+i*dash, bby+py1)
						y2 = utils.MinF(bby+py1+(i+1)*dash, bby+bbh+py2)
						x1 = mainOffset
						if way > 0 {
							x1 -= width
						}
						x2 = x1 + width
					}
					context.Rectangle(x1, y1, x2-x1, y2-y1)
				}
			}
		} else {
			// 2x + 1 dashes
			context.Clip(true)
			ld := fl(math.Round(float64(length / dash)))
			denom := ld - utils.FloatModulo(ld+1, 2)
			dash = length
			if denom != 0 {
				dash /= denom
			}
			maxI := int(math.Round(float64(length / dash)))
			for i_ := 0; i_ < maxI; i_ += 2 {
				i := fl(i_)
				switch side {
				case "top":
					context.Rectangle(bbx+i*dash, bby, dash, width)
				case "right":
					context.Rectangle(bbx+bbw-width, bby+i*dash, width, dash)
				case "bottom":
					context.Rectangle(bbx+i*dash, bby+bbh-width, dash, width)
				case "left":
					context.Rectangle(bbx, bby+i*dash, width, dash)
				}
			}
		}
	}
	context.Clip(true)
}

func (ctx drawContext) drawRoundedBorder(box *bo.BoxFields, style pr.String, colors []Color) {
	roundedBoxPath(ctx.dst, box.RoundedPaddingBox())
	if style == "ridge" || style == "groove" {
		roundedBoxPath(ctx.dst, box.RoundedBoxRatio(1./2))
		ctx.dst.SetColorRgba(colors[0], false)
		ctx.dst.Fill(true)
		roundedBoxPath(ctx.dst, box.RoundedBoxRatio(1./2))
		roundedBoxPath(ctx.dst, box.RoundedBorderBox())
		ctx.dst.SetColorRgba(colors[1], false)
		ctx.dst.Fill(true)
		return
	}
	if style == "double" {
		roundedBoxPath(ctx.dst, box.RoundedBoxRatio(1./3))
		roundedBoxPath(ctx.dst, box.RoundedBoxRatio(2./3))
	}
	roundedBoxPath(ctx.dst, box.RoundedBorderBox())
	ctx.dst.SetColorRgba(colors[0], false)
	ctx.dst.Fill(true)
}

func (ctx drawContext) drawRectBorder(box, widths pr.Rectangle, style pr.String, color []Color) {
	bbx, bby, bbw, bbh := box.Unpack()
	bt, br, bb, bl := widths.Unpack()
	ctx.dst.Rectangle(box.Unpack())
	if style == "ridge" || style == "groove" {
		ctx.dst.Rectangle(bbx+bl/2, bby+bt/2, bbw-(bl+br)/2, bbh-(bt+bb)/2)
		ctx.dst.SetColorRgba(color[0], false)
		ctx.dst.Fill(true)
		ctx.dst.Rectangle(bbx+bl/2, bby+bt/2, bbw-(bl+br)/2, bbh-(bt+bb)/2)
		ctx.dst.Rectangle(bbx+bl, bby+bt, bbw-bl-br, bbh-bt-bb)
		ctx.dst.SetColorRgba(color[1], false)
		ctx.dst.Fill(true)
		return
	}
	if style == "double" {
		ctx.dst.Rectangle(bbx+bl/3, bby+bt/3, bbw-(bl+br)/3, bbh-(bt+bb)/3)
		ctx.dst.Rectangle(bbx+bl*2/3, bby+bt*2/3, bbw-(bl+br)*2/3, bbh-(bt+bb)*2/3)
	}
	ctx.dst.Rectangle(bbx+bl, bby+bt, bbw-bl-br, bbh-bt-bb)
	ctx.dst.SetColorRgba(color[0], false)
	ctx.dst.Fill(true)
}

func (ctx drawContext) drawOutlines(box_ Box) {
	box := box_.Box()
	width_ := box.Style.GetOutlineWidth()
	color := tree.ResolveColor(box.Style, "outline_color").RGBA
	style := box.Style.GetOutlineStyle()
	if box.Style.GetVisibility() == "visible" && width_.Value != 0 && color.A != 0 {
		width := width_.Value
		outlineBox := pr.Rectangle{
			box.BorderBoxX() - width, box.BorderBoxY() - width,
			box.BorderWidth() + 2*width, box.BorderHeight() + 2*width,
		}
		for _, side := range sides {
			ctx.dst.OnNewStack(func() {
				clipBorderSegment(ctx.dst, style, fl(width), side, outlineBox, nil, nil)
				ctx.drawRectBorder(outlineBox, pr.Rectangle{width, width, width, width},
					style, styledColor(style, color, side))
			})
		}
	}

	if bo.ParentBoxT.IsInstance(box_) {
		for _, child := range box.Children {
			if child.IsClassicalBox() {
				ctx.drawOutlines(child)
			}
		}
	}
}

type segment struct {
	side string
	bo.Border
	borderBox pr.Rectangle
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Draw borders of table cells when they collapse.
func (ctx drawContext) drawCollapsedBorders(table *bo.TableBox) {
	var rowHeights, rowPositions []pr.Float
	for _, rowGroup := range table.Children {
		for _, row := range rowGroup.Box().Children {
			rowHeights = append(rowHeights, row.Box().Height.V())
			rowPositions = append(rowPositions, row.Box().PositionY)
		}
	}
	columnWidths := table.ColumnWidths
	if len(rowHeights) == 0 || len(columnWidths) == 0 {
		// One of the list is empty: don’t bother with empty tables
		return
	}
	columnPositions := table.ColumnPositions // shallow copy
	gridHeight := len(rowHeights)
	gridWidth := len(columnWidths)

	if gridWidth != len(columnPositions) {
		panic(fmt.Sprintf("expected same gridWidth and columnPositions length, got %d, %d", gridWidth, len(columnPositions)))
	}

	// Add the end of the last column, but make a copy from the table attr.
	if table.Style.GetDirection() == "ltr" {
		columnPositions = append(columnPositions, columnPositions[len(columnPositions)-1]+columnWidths[len(columnWidths)-1])
	} else {
		columnPositions = append([]pr.Float{columnPositions[0] + columnWidths[0]}, columnPositions...)
	}

	// Add the end of the last row.
	rowPositions = append(rowPositions, rowPositions[len(rowPositions)-1]+rowHeights[len(rowHeights)-1])
	verticalBorders, horizontalBorders := table.CollapsedBorderGrid.Vertical, table.CollapsedBorderGrid.Horizontal

	headerRows := 0
	if table.Children[0].Box().IsHeader {
		headerRows = len(table.Children[0].Box().Children)
	}

	footerRows := 0
	if L := len(table.Children); table.Children[L-1].Box().IsFooter {
		footerRows = len(table.Children[L-1].Box().Children)
	}

	skippedRows := table.SkippedRows
	bodyRowsOffset := 0
	if skippedRows != 0 {
		bodyRowsOffset = skippedRows - headerRows
	}

	originalGridHeight := len(verticalBorders)
	footerRowsOffset := originalGridHeight - gridHeight

	rowNumber := func(y int, horizontal bool) int {
		// fmt.Println(headerRows, footerRows, y, horizontal)
		// Examples in comments for 2 headers rows, 5 body rows, 3 footer rows
		if headerRows != 0 && y < (headerRows+boolToInt(horizontal)) {
			// Row in header: y < 2 for vertical, y < 3 for horizontal
			return y
		} else if footerRows != 0 && y >= (gridHeight-footerRows-boolToInt(horizontal)) {
			// Row in footer: y >= 7 for vertical, y >= 6 for horizontal
			return y + footerRowsOffset
		} else {
			// Row in body: 2 >= y > 7 for vertical, 3 >= y > 6 for horizontal
			return y + bodyRowsOffset
		}
	}

	var segments []segment

	// vertical=true
	halfMaxWidth := func(borderList [][]bo.Border, yxPairs [2][2]int, vertical bool) pr.Float {
		var result pr.Float
		for _, tmp := range yxPairs {
			y, x := tmp[0], tmp[1]
			cond := 0 <= y && y <= gridHeight && 0 <= x && x < gridWidth
			if vertical {
				cond = 0 <= y && y < gridHeight && 0 <= x && x <= gridWidth
			}
			if cond {
				yy := rowNumber(y, !vertical)
				width := pr.Float(borderList[yy][x].Width)
				result = pr.Max(result, width)
			}
		}
		return result / 2
	}

	addVertical := func(x, y int) {
		yy := rowNumber(y, false)
		border := verticalBorders[yy][x]
		if border.Width == 0 || border.Color.RGBA.A == 0 {
			return
		}
		posX := columnPositions[x]
		posY1 := rowPositions[y]
		if y != 0 || !table.SkipCellBorderTop {
			posY1 -= halfMaxWidth(horizontalBorders, [2][2]int{{y, x - 1}, {y, x}}, false)
		}
		posY2 := rowPositions[y+1]
		if y != gridHeight-1 || !table.SkipCellBorderBottom {
			posY2 += halfMaxWidth(horizontalBorders, [2][2]int{{y + 1, x - 1}, {y + 1, x}}, false)
		}
		segments = append(segments, segment{
			Border: border, side: "left",
			borderBox: pr.Rectangle{posX - pr.Float(border.Width)/2, posY1, 0, posY2 - posY1},
		})
	}

	addHorizontal := func(x, y int) {
		if y == 0 && table.SkipCellBorderTop {
			return
		}
		if y == gridHeight && table.SkipCellBorderBottom {
			return
		}

		yy := rowNumber(y, true)
		border := horizontalBorders[yy][x]
		if border.Width == 0 || border.Color.RGBA.A == 0 {
			return
		}
		posY := rowPositions[y]
		shiftBefore := halfMaxWidth(verticalBorders, [2][2]int{{y - 1, x}, {y, x}}, true)
		shiftAfter := halfMaxWidth(verticalBorders, [2][2]int{{y - 1, x + 1}, {y, x + 1}}, true)
		var posX1, posX2 pr.Float
		if table.Style.GetDirection() == "ltr" {
			posX1 = columnPositions[x] - shiftBefore
			posX2 = columnPositions[x+1] + shiftAfter
		} else {
			posX1 = columnPositions[x+1] - shiftAfter
			posX2 = columnPositions[x] + shiftBefore
		}
		segments = append(segments, segment{
			Border: border, side: "top",
			borderBox: pr.Rectangle{posX1, posY - pr.Float(border.Width)/2, posX2 - posX1, 0},
		})
	}

	for x := 0; x < gridWidth; x++ {
		addHorizontal(x, 0)
	}
	for y := 0; y < gridHeight; y++ {
		addVertical(0, y)
		for x := 0; x < gridWidth; x++ {
			addVertical(x+1, y)
			addHorizontal(x, y+1)
		}
	}

	// Sort bigger scores last (painted later, on top)
	// Since the number of different scores is expected to be small compared
	// to the number of segments, there should be little changes and Timsort
	// should be closer to O(n) than O(n * log(n))
	sort.SliceStable(segments, func(i, j int) bool {
		return segments[i].Border.Score.Lower(segments[j].Border.Score)
	})

	for _, segment := range segments {
		widths := pr.Rectangle{0, 0, 0, pr.Float(segment.Width)}
		if segment.side == "top" {
			widths = pr.Rectangle{pr.Float(segment.Width), 0, 0, 0}
		}
		ctx.dst.OnNewStack(func() {
			clipBorderSegment(ctx.dst, segment.Style, segment.Width, segment.side, segment.borderBox,
				&widths, nil)
			ctx.drawRectBorder(segment.borderBox, widths, segment.Style,
				styledColor(segment.Style, segment.Color.RGBA, segment.side))
		})
	}
}

// Draw the given :class:`boxes.ReplacedBox`
func (ctx drawContext) drawReplacedbox(box_ bo.ReplacedBoxITF) {
	box := box_.Replaced()
	if box.Style.GetVisibility() != "visible" || !pr.Is(box.Width) || !pr.Is(box.Height) {
		return
	}

	drawWidth, drawHeight, drawX, drawY := layout.LayoutReplacedBox(box_)
	if drawWidth <= 0 || drawHeight <= 0 {
		return
	}

	ctx.dst.OnNewStack(func() {
		roundedBoxPath(ctx.dst, box.RoundedContentBox())
		ctx.dst.Clip(false)
		ctx.dst.Transform(matrix.New(1, 0, 0, 1, pr.Fl(drawX), pr.Fl(drawY)))
		ctx.dst.OnNewStack(func() {
			box.Replacement.Draw(ctx.dst, pr.Fl(drawWidth), pr.Fl(drawHeight), string(box.Style.GetImageRendering()))
		})
	})
}

// offsetX=0, textOverflow="clip"
func (ctx drawContext) drawInlineLevel(page *bo.PageBox, box_ Box, offsetX fl, textOverflow string, blockEllipsis pr.NamedString) {
	if stackingContext, ok := box_.(StackingContext); ok {
		if !(bo.InlineBlockBoxT.IsInstance(stackingContext.box) || bo.InlineFlexBoxT.IsInstance(stackingContext.box)) {
			panic(fmt.Sprintf("expected InlineBlock or InlineFlex, got %v", stackingContext.box))
		}
		ctx.drawStackingContext(stackingContext)
	} else {
		box := box_.Box()
		ctx.drawBackgroundDefaut(box.Background)
		ctx.drawBorder(box_)
		textBox, isTextBox := box_.(*bo.TextBox)
		replacedBox, isReplacedBox := box_.(bo.ReplacedBoxITF)
		if layout.IsLine(box_) {
			if lineBox, ok := box_.(*bo.LineBox); ok {
				textOverflow = lineBox.TextOverflow
				blockEllipsis = lineBox.BlockEllipsis
			}
			for _, child := range box.Children {
				childOffsetX := offsetX
				if _, ok := child.(StackingContext); !ok {
					childOffsetX = offsetX + fl(child.Box().PositionX) - fl(box.PositionX)
				}
				if childT, ok := child.(*bo.TextBox); ok {
					ctx.drawText(childT, childOffsetX, textOverflow, blockEllipsis)
				} else {
					ctx.drawInlineLevel(page, child, childOffsetX, textOverflow, blockEllipsis)
				}
			}
		} else if isReplacedBox {
			ctx.drawReplacedbox(replacedBox)
		} else if isTextBox {
			// Should only happen for list markers
			ctx.drawText(textBox, offsetX, textOverflow, blockEllipsis)
		} else {
			panic(fmt.Sprintf("unexpected box %s", box_.Type()))
		}
	}
}

// Draw ``textbox`` to a ``cairo.Context`` from ``PangoCairo.Context``
// 	(offsetX=0,textOverflow="clip")
func (ctx drawContext) drawText(textbox *bo.TextBox, offsetX fl, textOverflow string, blockEllipsis pr.NamedString) {
	if textbox.Style.GetVisibility() != "visible" {
		return
	}

	x, y := pr.Fl(textbox.PositionX), pr.Fl(textbox.PositionY+textbox.Baseline.V())
	ctx.dst.SetColorRgba(textbox.Style.GetColor().RGBA, false)

	textbox.PangoLayout.ApplyJustification()
	ctx.drawFirstLine(textbox, textOverflow, blockEllipsis, x, y)

	// Draw text decoration

	decoration := textbox.Style.GetTextDecorationLine()
	color := textbox.Style.GetTextDecorationColor()
	if color.Type == parser.ColorCurrentColor {
		color = textbox.Style.GetColor()
	}

	thickness := pr.Fl(textbox.Style.GetFontSize().Value / 18) // Like other browsers do
	var offsetY pr.Float

	metrics := textbox.PangoLayout.Metrics

	if utils.Set(decoration).Has("overline") {
		thickness = text.PangoUnitsToFloat((metrics.UnderlineThickness))
		offsetY = textbox.Baseline.V() - pr.Float(text.PangoUnitsToFloat(metrics.Ascent)) + pr.Float(thickness)/2
	}
	if utils.Set(decoration).Has("underline") {
		thickness = text.PangoUnitsToFloat((metrics.UnderlineThickness))
		offsetY = textbox.Baseline.V() - pr.Float(text.PangoUnitsToFloat(metrics.UnderlinePosition)) + pr.Float(thickness)/2
	}
	if utils.Set(decoration).Has("line-through") {
		thickness = text.PangoUnitsToFloat((metrics.StrikethroughThickness))
		offsetY = textbox.Baseline.V() - pr.Float(text.PangoUnitsToFloat(metrics.StrikethroughPosition))
	}

	if !decoration.IsNone() {
		ctx.drawTextDecoration(textbox, offsetX, pr.Fl(offsetY), thickness, color.RGBA)
	}
}

func (ctx drawContext) drawFirstLine(textbox *bo.TextBox, textOverflow string, blockEllipsis pr.NamedString, x, y pr.Fl) {
	layout := &textbox.PangoLayout.Layout
	layout.SetSingleParagraphMode(true)

	var ellipsis string
	if textOverflow == "ellipsis" || blockEllipsis.Name != "none" {
		// assert textbox.PangoLayout.maxWidth is not nil
		maxWidth := textbox.PangoLayout.MaxWidth.V()
		layout.SetWidth(pango.Unit(text.PangoUnitsFromFloat(pr.Fl(maxWidth))))
		if textOverflow == "ellipsis" {
			layout.SetEllipsize(pango.ELLIPSIZE_END)
		} else {
			ellipsis = blockEllipsis.String
			if blockEllipsis.Name == "auto" {
				ellipsis = "…"
			}
			// Remove last word if hyphenated
			newText := layout.Text
			if hyph := string(textbox.Style.GetHyphenateCharacter()); strings.HasSuffix(string(newText), hyph) {
				lastWordEnd := text.GetLastWordEnd(newText[:len(newText)-len([]rune(hyph))])
				if lastWordEnd != -1 && lastWordEnd != 0 {
					newText = newText[:lastWordEnd]
				}
			}
			textbox.PangoLayout.SetText(string(newText) + ellipsis)
		}
	}

	firstLine, secondLine := textbox.PangoLayout.GetFirstLine()
	if blockEllipsis.Name != "none" {
		for secondLine != 0 && secondLine != -1 {
			lastWordEnd := text.GetLastWordEnd(layout.Text[:len(layout.Text)-len([]rune(ellipsis))])
			if lastWordEnd == -1 {
				break
			}
			newText := layout.Text[:lastWordEnd]
			textbox.PangoLayout.SetText(string(newText) + ellipsis)
			firstLine, secondLine = textbox.PangoLayout.GetFirstLine()
		}
	}

	var (
		output               backend.TextDrawing
		inkRect, logicalRect pango.Rectangle
		lastFont             *backend.Font
		xAdvance             pr.Fl
	)
	fontSize := pr.Fl(textbox.Style.GetFontSize().Value)
	output.FontSize = fontSize
	output.X, output.Y = x, y

	textRunes := layout.Text
	for run := firstLine.Runs; run != nil; run = run.Next {

		// Pango objects
		glyphItem := run.Data
		glyphString := glyphItem.Glyphs
		runStart := glyphItem.Item.Offset

		// Font content
		pangoFont := glyphItem.Item.Analysis.Font
		content := ctx.fonts.FontContent(pangoFont.FaceID())
		outFont := ctx.dst.AddFont(pangoFont, content)

		if outFont != lastFont { // add a new "run"
			var outRun backend.TextRun
			outRun.Font = pangoFont
			output.Runs = append(output.Runs, outRun)
		} else { // use the last one
		}
		runDst := &output.Runs[len(output.Runs)-1]

		runDst.Glyphs = make([]backend.TextGlyph, len(glyphString.Glyphs))
		for i, glyphInfo := range glyphString.Glyphs {
			outGlyph := &runDst.Glyphs[i]
			width := glyphInfo.Geometry.Width
			glyph := glyphInfo.Glyph

			if glyph == pango.GLYPH_EMPTY {
				outGlyph.Offset = pr.Fl(width) / fontSize
				outGlyph.Glyph = fonts.EmptyGlyph
				continue
			}

			outGlyph.Offset = pr.Fl(glyphInfo.Geometry.XOffset) / fontSize
			outGlyph.Glyph = glyph.GID()

			// Ink bounding box and logical widths in font
			if _, in := outFont.Extents[outGlyph.Glyph]; !in {
				pangoFont.GlyphExtents(glyph, &inkRect, &logicalRect)
				x1, y1, x2, y2 := inkRect.X, -inkRect.Y-inkRect.Height,
					inkRect.X+inkRect.Width, -inkRect.Y
				if int(x1) < outFont.Bbox[0] {
					outFont.Bbox[0] = int(text.PangoUnitsToFloat(x1*1000) / fontSize)
				}
				if int(y1) < outFont.Bbox[1] {
					outFont.Bbox[1] = int(text.PangoUnitsToFloat(y1*1000) / fontSize)
				}
				if int(x2) > outFont.Bbox[2] {
					outFont.Bbox[2] = int(text.PangoUnitsToFloat(x2*1000) / fontSize)
				}
				if int(y2) > outFont.Bbox[3] {
					outFont.Bbox[3] = int(text.PangoUnitsToFloat(y2*1000) / fontSize)
				}
				outFont.Extents[outGlyph.Glyph] = backend.GlyphExtents{
					Width:  int(text.PangoUnitsToFloat(logicalRect.Width*1000) / fontSize),
					Y:      int(text.PangoUnitsToFloat(logicalRect.Y*1000) / fontSize),
					Height: int(text.PangoUnitsToFloat(logicalRect.Height*1000) / fontSize),
				}
			}

			// Kerning, word spacing, letter spacing
			outGlyph.Kerning = int(pr.Fl(outFont.Extents[outGlyph.Glyph].Width) - text.PangoUnitsToFloat(width*1000)/fontSize + outGlyph.Offset)

			// Mapping between glyphs and characters
			startPos := runStart + glyphString.LogClusters[i] // Positions of the glyphs in the UTF-8 string
			endPos := runStart + glyphItem.Item.Length
			if i < len(glyphString.Glyphs)-1 {
				endPos = runStart + glyphString.LogClusters[i+1]
			}
			if _, in := outFont.Cmap[outGlyph.Glyph]; !in {
				outFont.Cmap[outGlyph.Glyph] = textRunes[startPos:endPos]
			}

			// advance
			outGlyph.XAdvance = xAdvance
			xAdvance += pr.Fl(outFont.Extents[outGlyph.Glyph].Width) + outGlyph.Offset
		}
	}

	ctx.dst.DrawText(output)
}

func (ctx drawContext) drawWave(x, y, width, offsetX, radius pr.Fl) {
	var up pr.Fl = 1
	maxX := x + width
	ctx.dst.Rectangle(x, y-2*radius, width, 4*radius)
	ctx.dst.Clip(false)

	x -= offsetX
	ctx.dst.MoveTo(x, y)

	for x < maxX {
		ctx.dst.CubicTo(x+radius/2, y+up*radius,
			x+3*radius/2, y+up*radius,
			x+2*radius, y)
		x += 2 * radius
		up *= -1
	}
}

// Draw text-decoration of ``textbox`` to a ``context``.
func (ctx drawContext) drawTextDecoration(textbox *bo.TextBox, offsetX, offsetY, thickness pr.Fl, color Color) {
	style := textbox.Style.GetTextDecorationStyle()

	ctx.dst.OnNewStack(func() {
		ctx.dst.SetColorRgba(color, true)
		ctx.dst.SetLineWidth(thickness)

		if style == "dashed" {
			ctx.dst.SetDash([]fl{5 * thickness}, offsetX)
		} else if style == "dotted" {
			ctx.dst.SetDash([]fl{thickness}, offsetX)
		}

		posX, posY, width := fl(textbox.PositionX), fl(textbox.PositionY), fl(textbox.Width.V())

		if style == "wavy" {
			thickness *= 0.75
			ctx.drawWave(posX, posY+offsetY, width, offsetX, thickness)
		} else {
			ctx.dst.MoveTo(posX, posY+offsetY)
			ctx.dst.LineTo(posX+width, posY+offsetY)
		}

		ctx.dst.SetLineWidth(thickness)

		if style == "double" {
			delta := 2 * thickness
			ctx.dst.MoveTo(posX, posY+offsetY+delta)
			ctx.dst.LineTo(posX+width, posY+offsetY+delta)
		}

		ctx.dst.Stroke()
	})
}
