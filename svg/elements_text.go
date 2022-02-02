package svg

import (
	"math"
	"strings"

	"github.com/benoitkugler/webrender/backend"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/text"
	drawText "github.com/benoitkugler/webrender/text/draw"
)

// text tags
type span struct {
	style  pr.Properties
	text   string
	rotate []Fl // angles in degrees

	dx, dy Value

	letterSpacing Value

	textAnchor, displayAnchor anchor

	baseline baseline
}

func newText(node *cascadedNode, tree *svgContext) (drawable, error) {
	var out span

	out.text = string(node.text)
	out.style = pr.InitialValues.Copy()

	family := "sans-serif"
	if f, has := node.attrs["font-family"]; has {
		family = f
	}
	out.style.SetFontFamily(strings.Split(family, ","))

	if w, has := node.attrs["font-weight"]; has {
		out.style.SetFontWeight(pr.IntString{Int: parseFontWeight(w)})
	}

	if s, has := node.attrs["font-style"]; has {
		out.style.SetFontStyle(pr.String(s))
	}

	// get rotations and translations
	var err error
	out.dx, err = parseValue(node.attrs["dx"])
	if err != nil {
		return nil, err
	}
	out.dy, err = parseValue(node.attrs["dy"])
	if err != nil {
		return nil, err
	}

	out.rotate, err = parsePoints(node.attrs["rotate"], nil, false)
	if err != nil {
		return nil, err
	}

	out.letterSpacing, err = parseValue(node.attrs["letter-spacing"])
	if err != nil {
		return nil, err
	}

	out.textAnchor = parseAnchor(node.attrs["text-anchor"])
	out.displayAnchor = parseAnchor(node.attrs["display-anchor"])

	baseline, has := node.attrs["dominant-baseline"]
	if !has {
		baseline = node.attrs["alignment-baseline"]
	}
	out.baseline = parseBaseline(baseline)

	return out, nil
}

func (t span) draw(dst backend.Canvas, attrs *attributes, svg *SVGImage, dims drawingDims) []vertex {
	t.style.SetFontSize(pr.FToV(dims.fontSize))

	splitted := text.SplitFirstLine(t.text, t.style, svg.textContext, pr.Inf, 0, false)
	x, y := dims.point(attrs.x, attrs.y)
	dx, dy := dims.point(t.dx, t.dy)

	// return early when there’s no text,
	// update the cursor position though
	if t.text == "" {
		if attrs.x.U == 0 { // no x specified
			x = svg.cursorPosition.x
		}
		if attrs.y.U == 0 { // no y specified
			y = svg.cursorPosition.y
		}
		svg.cursorPosition = point{x + dx, y + dy}
		return nil
	}

	var xAlign, yAlign, xBearing, yBearing Fl

	// align text box horizontally
	letterSpacing := dims.length(t.letterSpacing)
	ascentL, descentL := dims.fontSize*.8, dims.fontSize*.2

	width, height := Fl(splitted.Width), Fl(splitted.Height)
	switch t.textAnchor {
	case middle:
		xAlign = -(width/2. + xBearing)
		if letterSpacing != 0 && t.text != "" {
			xAlign -= Fl(len(splitted.Layout.Layout.Text)-1) * letterSpacing / 2
		}
	case end:
		xAlign = -(width + xBearing)
		if letterSpacing != 0 && t.text != "" {
			xAlign -= Fl(len(splitted.Layout.Layout.Text)-1) * letterSpacing
		}

	}

	// align text box vertically
	if t.displayAnchor == middle {
		yAlign = -height/2 - yBearing
	} else if t.displayAnchor == top {
		yAlign = -yBearing
	} else if t.displayAnchor == bottom {
		yAlign = -height - yBearing
	} else if t.baseline == central {
		yAlign = (ascentL+descentL)/2 - descentL
	} else if t.baseline == ascent {
		yAlign = ascentL
	} else if t.baseline == descent {
		yAlign = -descentL
	}

	chars := []rune(t.text)
	var bbox Rectangle
	for i, r := range chars {
		var rot Fl // en radians
		if i < len(t.rotate) {
			rot = t.rotate[i] * math.Pi / 180
		} else if L := len(t.rotate); L != 0 {
			rot = t.rotate[L-1] * math.Pi / 180
		}

		if attrs.x.U != 0 { // x specified
			svg.cursorDPosition.x = 0
		}
		if attrs.y.U != 0 { // y specified
			svg.cursorDPosition.y = 0
		}
		svg.cursorDPosition.x += dx
		svg.cursorDPosition.y += dy

		splitted := text.SplitFirstLine(string(r), t.style, svg.textContext, pr.Inf, 0, false)
		layout := splitted.Layout
		width, height = Fl(splitted.Width), Fl(splitted.Height)

		letterX, letterY := svg.cursorPosition.x, svg.cursorPosition.y
		if i != 0 && attrs.x.U != 0 {
			letterX = x
		}
		if i != 0 && attrs.y.U != 0 {
			letterY = y
		}

		if i != 0 {
			letterX += letterSpacing
		}

		xPosition := letterX + svg.cursorDPosition.x + xAlign
		yPosition := letterY + svg.cursorDPosition.y + yAlign

		cursorPosition := point{letterX + width, letterY}

		bb := Rectangle{
			cursorPosition.x + xAlign + svg.cursorDPosition.x,
			cursorPosition.y + yAlign + svg.cursorDPosition.y,
			width,
			height,
		}
		if i == 0 {
			bbox = bb
		} else {
			bbox.union(bb)
		}

		layout.ApplyJustification()

		dst.OnNewStack(func() {
			dst.MoveTo(xPosition, yPosition)

			if rot != 0 {
				dst.Transform(matrix.Translation(xPosition, yPosition))
				dst.Transform(matrix.Rotation(rot))
				dst.Transform(matrix.Translation(-xPosition, -yPosition))
			}

			// svg.paintNode(dst, node, dims) // TODO: check whether it is needed
			textContext := drawText.Context{Output: dst, Fonts: svg.textContext.Fonts()}
			textContext.DrawFirstLine(layout, t.style, "none", pr.NamedString{Name: "none"}, xPosition, yPosition)
		})

		svg.cursorPosition = cursorPosition
	}
	return nil
}
