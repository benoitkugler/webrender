package svg

import (
	"math"
	"strings"

	"github.com/benoitkugler/webrender/backend"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/text"
	drawText "github.com/benoitkugler/webrender/text/draw"
)

// text tags
type span struct {
	style  pr.Properties
	text   string
	rotate []Fl // angles in degrees

	x, y, dx, dy []Value

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
	out.x, err = parseValues(node.attrs["x"])
	if err != nil {
		return nil, err
	}
	out.y, err = parseValues(node.attrs["y"])
	if err != nil {
		return nil, err
	}

	out.dx, err = parseValues(node.attrs["dx"])
	if err != nil {
		return nil, err
	}
	out.dy, err = parseValues(node.attrs["dy"])
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

	splitted := text.SplitFirstLine(t.text, t.style, svg.textContext, pr.Inf, false, true)

	var x, y, dx, dy []Fl
	for _, v := range t.x {
		x = append(x, v.Resolve(dims.fontSize, dims.innerWidth))
	}
	for _, v := range t.y {
		y = append(y, v.Resolve(dims.fontSize, dims.innerHeight))
	}
	for _, v := range t.dx {
		dx = append(dx, v.Resolve(dims.fontSize, dims.innerWidth))
	}
	for _, v := range t.dy {
		dy = append(dy, v.Resolve(dims.fontSize, dims.innerHeight))
	}

	// return early when there’s no text,
	// update the cursor position though
	if t.text == "" {
		x0 := svg.cursorPosition.x
		if len(x) != 0 {
			x0 = x[0]
		}
		y0 := svg.cursorPosition.y
		if len(y) != 0 {
			y0 = y[0]
		}
		var dx0, dy0 Fl
		if len(dx) != 0 {
			dx0 = dx[0]
		}
		if len(dy) != 0 {
			dy0 = dy[0]
		}
		svg.cursorPosition = point{x0 + dx0, y0 + dy0}
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
			xAlign -= Fl(len(splitted.Layout.Text())-1) * letterSpacing / 2
		}
	case end:
		xAlign = -(width + xBearing)
		if letterSpacing != 0 && t.text != "" {
			xAlign -= Fl(len(splitted.Layout.Text())-1) * letterSpacing
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
	var (
		bbox   Rectangle
		texts  []backend.TextDrawing
		drawer = drawText.Context{Output: dst, Fonts: svg.textContext.Fonts()}
	)
	for i, r := range chars {
		hasX, hasY := i < len(x), i < len(y)

		var angle Fl // en radians
		if i < len(t.rotate) {
			angle = t.rotate[i] * math.Pi / 180
		} else if L := len(t.rotate); L != 0 {
			angle = t.rotate[L-1] * math.Pi / 180
		}

		if hasX && x[i] != 0 { // x specified
			svg.cursorDPosition.x = 0
		}
		if hasY && y[i] != 0 { // y specified
			svg.cursorDPosition.y = 0
		}
		if i < len(dx) {
			svg.cursorDPosition.x += dx[i]
		}
		if i < len(dy) {
			svg.cursorDPosition.y += dy[i]
		}

		splitted := text.SplitFirstLine(string(r), t.style, svg.textContext, pr.Inf, false, true)
		layout := splitted.Layout
		width, height = Fl(splitted.Width), Fl(splitted.Height)

		letterX, letterY := svg.cursorPosition.x, svg.cursorPosition.y
		if hasX {
			letterX = x[i]
		}
		if hasY {
			letterY = y[i]
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

		doFill, doStroke := svg.setupPaint(dst, &svgNode{graphicContent: t, attributes: *attrs}, dims)
		dst.State().SetTextPaint(newPaintOp(doFill, doStroke, false))
		texts = append(texts,
			drawer.CreateFirstLine(layout, "none", pr.TaggedString{Tag: pr.None}, xPosition, yPosition, angle))

		svg.cursorPosition = cursorPosition
	}

	dst.OnNewStack(func() {
		dst.DrawText(texts)
	})

	return nil
}
