package svg

import (
	"fmt"
	"log"
	"strings"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/utils"
)

// handle painter for fill and stroke values

// painter is either a simple RGBA color,
// or a reference to a more complex `paintServer`
type painter struct {
	// value of the url attribute, refering
	// to a paintServer element
	refID string

	color parser.RGBA

	// if 'false', no painting occurs (not the same as black)
	valid bool
}

// parse a fill or stroke attribute
func newPainter(attr string) (painter, error) {
	attr = strings.TrimSpace(attr)
	if attr == "" || attr == "none" {
		return painter{}, nil
	}

	var out painter
	if strings.HasPrefix(attr, "url(") {
		if i := strings.IndexByte(attr, ')'); i != -1 {
			out.refID = parseURLFragment(attr[:i])
			attr = attr[i+1:] // skip the )
		} else {
			return out, fmt.Errorf("invalid url in color '%s'", attr)
		}
	}

	color := parser.ParseColorString(attr)
	// currentColor has been resolved during tree building
	out.color = color.RGBA
	out.valid = true

	return out, nil
}

// ensure that v is positive and equal to offset modulo total
func clampModulo(offset, total Fl) Fl {
	if offset < 0 { // shift to [0, dashesLength]
		offset = -offset
		quotient := utils.Floor(offset / total)
		remainder := offset - quotient*total
		return total - remainder
	}
	return offset
}

func (dims drawingDims) resolveDashes(dashArray []Value, dashOffset Value) ([]Fl, Fl) {
	dashes := make([]Fl, len(dashArray))
	var dashesLength Fl
	for i, v := range dashArray {
		dashes[i] = dims.length(v)
		if dashes[i] < 0 {
			return nil, 0
		}
		dashesLength += dashes[i]
	}
	if dashesLength == 0 {
		return nil, 0
	}
	offset := dims.length(dashOffset)
	offset = clampModulo(offset, dashesLength)
	return dashes, offset
}

// paint by filling and stroking the given node onto the graphic target
func (svg *SVGImage) paintNode(dst backend.CanvasNoFill, node *svgNode, dims drawingDims) {
	// TODO: handle text
	// if node.tag in ('text', 'textPath', 'a') and not text:
	// return

	strokeWidth := dims.length(node.strokeWidth)
	doFill := node.fill.valid
	doStroke := node.stroke.valid && strokeWidth > 0

	// fill
	if doFill {
		svg.applyPainter(dst, node, node.fill, node.fillOpacity, dims, false)
	}

	// stroke
	if doStroke {
		svg.applyPainter(dst, node, node.stroke, node.strokeOpacity, dims, true)

		// stroke options
		dashes, offset := dims.resolveDashes(node.dashArray, node.dashOffset)
		dst.SetDash(dashes, offset)

		dst.SetLineWidth(strokeWidth)
		dst.SetStrokeOptions(node.strokeOptions)
	}

	if doFill {
		dst.Fill(node.isFillEvenOdd)
	}
	if doStroke {
		dst.Stroke()
	}
}

// apply the given painter to the given node, outputing the
// the result in `dst`
// opacity is an additional opacity factor
func (svg *SVGImage) applyPainter(dst backend.CanvasNoFill, node *svgNode, pt painter, opacity Fl, dims drawingDims, stroke bool) {
	if !pt.valid {
		return
	}

	// check for a paintServer
	if ps := svg.definitions.paintServers[pt.refID]; ps != nil {
		wasPainted := ps.paint(dst, node, opacity, dims, stroke)
		if wasPainted {
			return
		} // else default to a plain color
	}

	pt.color.A *= opacity // apply the opacity factor
	dst.SetColorRgba(pt.color, stroke)
}

// gradient or pattern
type paintServer interface { // TODO:
	paint(dst backend.CanvasNoFill, node *svgNode, opacity Fl, dims drawingDims, stroke bool) bool
}

// either linear or radial
type gradientKind interface {
	isGradient()
}

func (gradientLinear) isGradient() {}
func (gradientRadial) isGradient() {}

type gradientLinear struct {
	x1, y1, x2, y2 Value
}

func newGradientLinear(node *cascadedNode) (out gradientLinear, err error) {
	out.x1, err = parseValue(node.attrs["x1"])
	if err != nil {
		return out, err
	}
	out.y1, err = parseValue(node.attrs["y1"])
	if err != nil {
		return out, err
	}
	out.x2, err = parseValue(node.attrs["x2"])
	if err != nil {
		return out, err
	}
	out.y2, err = parseValue(node.attrs["y2"])
	if err != nil {
		return out, err
	}

	// default values
	if out.x2.U == 0 {
		out.x2 = Value{100, Perc} // 100%
	}

	return out, nil
}

type gradientRadial struct {
	cx, cy, r, fx, fy, fr Value
}

func newGradientRadial(node *cascadedNode) (out gradientRadial, err error) {
	cx, cy, r := node.attrs["cx"], node.attrs["cy"], node.attrs["r"]
	if cx == "" {
		cx = "50%"
	}
	if cy == "" {
		cy = "50%"
	}
	if r == "" {
		r = "50%"
	}
	fx, fy, fr := node.attrs["fx"], node.attrs["fy"], node.attrs["fr"]
	if fx == "" {
		fx = cx
	}
	if fy == "" {
		fy = cy
	}

	out.cx, err = parseValue(cx)
	if err != nil {
		return out, err
	}
	out.cy, err = parseValue(cy)
	if err != nil {
		return out, err
	}
	out.r, err = parseValue(r)
	if err != nil {
		return out, err
	}
	out.fx, err = parseValue(fx)
	if err != nil {
		return out, err
	}
	out.fy, err = parseValue(fy)
	if err != nil {
		return out, err
	}
	out.fr, err = parseValue(fr)
	if err != nil {
		return out, err
	}

	return out, nil
}

// gradient specification, prior to resolving units
type gradient struct {
	kind gradientKind

	spreadMethod GradientSpread // default to NoRepeat

	positions []Value
	colors    []parser.RGBA

	transforms []transform

	isUnitsUserSpace bool
}

// parse a linearGradient or radialGradient node
func newGradient(node *cascadedNode) (out gradient, err error) {
	out.positions = make([]Value, len(node.children))
	out.colors = make([]parser.RGBA, len(node.children))
	for i, child := range node.children {
		out.positions[i], err = parseValue(child.attrs["offset"])
		if err != nil {
			return out, err
		}

		sc, has := child.attrs["stop-color"]
		if !has {
			sc = "black"
		}
		stopColor := parser.ParseColorString(sc).RGBA

		stopColor.A, err = parseOpacity(child.attrs["stop-opacity"])
		if err != nil {
			return out, err
		}

		out.colors[i] = stopColor
	}

	out.isUnitsUserSpace = node.attrs["gradientUnits"] == "userSpaceOnUse"
	switch node.attrs["spreadMethod"] {
	case "repeat":
		out.spreadMethod = Repeat
	case "reflect":
		out.spreadMethod = Reflect
		// default NoRepeat
	}

	out.transforms, err = parseTransform(node.attrs["gradientTransform"])
	if err != nil {
		return out, err
	}

	switch node.tag {
	case "linearGradient":
		out.kind, err = newGradientLinear(node)
		if err != nil {
			return out, fmt.Errorf("invalid linear gradient: %s", err)
		}
	case "radialGradient":
		out.kind, err = newGradientRadial(node)
		if err != nil {
			return out, fmt.Errorf("invalid radial gradient: %s", err)
		}
	default:
		panic("unexpected node tag " + node.tag)
	}

	return out, nil
}

func (gr gradient) paint(dst backend.CanvasNoFill, node *svgNode, opacity Fl, dims drawingDims, stroke bool) bool {
	if len(gr.colors) == 0 {
		return false
	}

	if len(gr.colors) == 1 { // actually solid
		dst.SetColorRgba(gr.colors[0], stroke)
		return true
	}

	bbox, ok := node.resolveBoundingBox(dims, stroke)
	if !ok {
		return false
	}

	x, y := bbox.X, bbox.Y
	width, height := bbox.Width, bbox.Height
	if gr.isUnitsUserSpace {
		width, height = dims.innerWidth, dims.innerHeight
	}

	// resolve positions values
	positions := make([]Fl, len(gr.positions))
	var previousPos Fl
	for i, p := range gr.positions {
		pos := p.Resolve(dims.fontSize, 1)
		positions[i] = utils.MaxF(pos, previousPos) // ensure positions is increasing
		previousPos = pos
	}
	colors := append([]parser.RGBA(nil), gr.colors...)

	switch gr.spreadMethod {
	case Repeat, Reflect:
		if positions[0] > 0 {
			positions = append([]Fl{0}, positions...)
			colors = append([]parser.RGBA{colors[0]}, colors...)
		}
		if positions[len(positions)-1] < 1 {
			positions = append(positions, 1)
			colors = append(colors, colors[len(colors)-1])
		}
	default:
		// Add explicit colors at boundaries if needed, because PDF doesn’t
		// extend color stops that are not displayed
		if positions[0] == positions[1] {
			if _, isRadial := gr.kind.(gradientRadial); isRadial {
				// avoid negative radius for radial gradients
				positions = append([]Fl{0}, positions...)
			} else {
				positions = append([]Fl{positions[0] - 1}, positions...)
			}
			colors = append([]parser.RGBA{colors[0]}, colors...)
		}
		if L := len(positions); positions[L-2] == positions[L-1] {
			positions = append(positions, positions[L-1]+1)
			colors = append(colors, colors[len(colors)-1])
		}
	}

	var laidOutGradient backend.GradientLayout
	mt := matrix.Translation(x, y)
	switch kind := gr.kind.(type) {
	case gradientLinear:
		x1 := kind.x1.Resolve(dims.fontSize, 1)
		y1 := kind.y1.Resolve(dims.fontSize, 1)
		x2 := kind.x2.Resolve(dims.fontSize, 1)
		y2 := kind.y2.Resolve(dims.fontSize, 1)
		if gr.isUnitsUserSpace {
			x1 -= x
			y1 -= y
			x2 -= x
			y2 -= y
		} else {
			length := utils.MinF(width, height)
			x1 *= length
			y1 *= length
			x2 *= length
			y2 *= length

			// update the transformation matrix
			var a, d Fl = 1, 1
			if height < width {
				a = width / height
			} else {
				d = height / width
			}
			mt.LeftMultBy(matrix.Scaling(a, d))
		}
		vectorLength := utils.Hypot(x2-x1, y2-y1)
		laidOutGradient = gr.spreadMethod.LinearGradient(positions, colors, x1, y2, x2, y2, vectorLength)
	case gradientRadial:
		cx := kind.cx.Resolve(dims.fontSize, 1)
		cy := kind.cy.Resolve(dims.fontSize, 1)
		r := kind.r.Resolve(dims.fontSize, 1)
		fx := kind.fx.Resolve(dims.fontSize, width)
		fy := kind.fy.Resolve(dims.fontSize, height)
		fr := kind.fr.Resolve(dims.fontSize, 1)
		if gr.isUnitsUserSpace {
			cx -= x
			cy -= y
			fx -= x
			fy -= y
		} else {
			length := utils.MinF(width, height)
			cx *= length
			cy *= length
			r *= length
			fx *= length
			fy *= length
			fr *= length

			// update the transformation matrix
			var a, d Fl = 1, 1
			if height < width {
				a = width / height
			} else {
				d = height / width
			}
			mt.LeftMultBy(matrix.Scaling(a, d))
		}

		laidOutGradient = gr.spreadMethod.RadialGradient(positions, colors, fx, fy, fr, cx, cy, r, width, height)
	}

	laidOutGradient.Reapeating = gr.spreadMethod != NoRepeat

	if trs := gr.transforms; len(trs) != 0 {
		mat := aggregateTransforms(trs, dims.fontSize, dims.normalizedDiagonal)
		mt.LeftMultBy(mat)
	}

	if laidOutGradient.Kind == "solid" {
		dst.Rectangle(0, 0, width, height)
		dst.SetColorRgba(laidOutGradient.Colors[0], false)
		dst.Fill(false)
		return true
	}

	// TODO: handle transformation
	dst.DrawGradient(laidOutGradient, width, height)
	return true
}

type pattern struct { // TODO:
	transforms []transform

	box

	isUnitsUserSpace   bool // patternUnits
	isContentUnitsBBox bool // patternContentUnits
}

func newPattern(node *cascadedNode) (out pattern, err error) {
	out.isUnitsUserSpace = node.attrs["patternUnits"] == "userSpaceOnUse"
	out.isContentUnitsBBox = node.attrs["patternContentUnits"] == "objectBoundingBox"
	err = node.attrs.parseBox(&out.box)
	if err != nil {
		return out, fmt.Errorf("parsing pattern element: %s", err)
	}

	out.transforms, err = parseTransform(node.attrs["patternTransform"])
	if err != nil {
		return out, fmt.Errorf("parsing pattern element: %s", err)
	}

	return out, nil
}

func (pt pattern) paint(dst backend.CanvasNoFill, node *svgNode, opacity Fl, dims drawingDims, stroke bool) bool {
	log.Println("painting with pattern is not implemented")
	return false
}
