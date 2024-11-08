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
type textSpan struct {
	style  pr.Properties
	text   string
	rotate []Fl // angles in degrees

	x, y, dx, dy []Value

	letterSpacing, textLength Value
	lengthAdjust              bool // true for spacingAndGlyphs

	textAnchor, displayAnchor anchor

	baseline baseline
}

func newTextSpan(node *cascadedNode, tree *svgContext) (drawable, error) {
	var out textSpan

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
	out.textLength, err = parseValue(node.attrs["textLength"])
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

	out.lengthAdjust = node.attrs["lengthAdjust"] == "spacingAndGlyphs"

	return out, nil
}

func (t textSpan) draw(dst backend.Canvas, attrs *attributes, svg *SVGImage, dims drawingDims) []vertex {
	t.style.SetFontSize(pr.FToV(dims.fontSize))

	splitted := text.SplitFirstLine(t.text, t.style, svg.textContext, pr.Inf, false, true)
	width, height := Fl(splitted.Width), Fl(splitted.Height)

	// Get rotations and translations
	var xs, ys, dxs, dys []Fl
	for _, v := range t.x {
		xs = append(xs, v.Resolve(dims.fontSize, dims.innerWidth))
	}
	for _, v := range t.y {
		ys = append(ys, v.Resolve(dims.fontSize, dims.innerHeight))
	}
	for _, v := range t.dx {
		dxs = append(dxs, v.Resolve(dims.fontSize, dims.innerWidth))
	}
	for _, v := range t.dy {
		dys = append(dys, v.Resolve(dims.fontSize, dims.innerHeight))
	}

	var yAlign Fl

	letterSpacing := dims.length(t.letterSpacing)
	textLength := dims.length(t.textLength)
	scaleX := Fl(1.)
	if textLength != 0 && t.text == "" {
		// calculate the number of spaces to be considered for the text
		spacesCount := Fl(len(t.text) - 1)
		if t.lengthAdjust {
			// scale letterSpacing up/down to textLength
			widthWithSpacing := width + spacesCount*letterSpacing
			letterSpacing *= textLength / widthWithSpacing
			// calculate the glyphs scaling factor by:
			// - deducting the scaled letterSpacing from textLength
			// - dividing the calculated value by the original width
			spacelessTextLength := textLength - spacesCount*letterSpacing
			scaleX = spacelessTextLength / width
		} else if spacesCount != 0 {
			// adjust letter spacing to fit textLength
			letterSpacing = (textLength - width) / spacesCount
		}
		width = textLength
	}
	ascentL, descentL := dims.fontSize*.8, dims.fontSize*.2

	// align text box vertically
	if t.displayAnchor == middle {
		yAlign = -height / 2
	} else if t.displayAnchor == top {
		// pass
	} else if t.displayAnchor == bottom {
		yAlign = -height
	} else if t.baseline == central {
		yAlign = (ascentL+descentL)/2 - descentL
	} else if t.baseline == ascent {
		yAlign = ascentL
	} else if t.baseline == descent {
		yAlign = -descentL
	}

	// return early when there’s no text,
	// update the cursor position though
	if t.text == "" {
		x0 := svg.cursorPosition.x
		if len(xs) != 0 {
			x0 = xs[0]
		}
		y0 := svg.cursorPosition.y
		if len(ys) != 0 {
			y0 = ys[0]
		}
		var dx0, dy0 Fl
		if len(dxs) != 0 {
			dx0 = dxs[0]
		}
		if len(dys) != 0 {
			dy0 = dys[0]
		}
		svg.cursorPosition = point{x0 + dx0, y0 + dy0}
		return nil
	}

	// Draw letters
	chars := []rune(t.text)
	var (
		bbox   Rectangle
		texts  []backend.TextDrawing
		drawer = drawText.Context{Output: dst, Fonts: svg.textContext.Fonts()}
	)
	for i, r := range chars {
		hasX, hasY := i < len(xs), i < len(ys)

		var angle Fl // en radians
		if i < len(t.rotate) {
			angle = t.rotate[i] * math.Pi / 180
		} else if L := len(t.rotate); L != 0 {
			angle = t.rotate[L-1] * math.Pi / 180
		}

		if hasX && xs[i] != 0 { // x specified
			svg.cursorDPosition.x = 0
		}
		if hasY && ys[i] != 0 { // y specified
			svg.cursorDPosition.y = 0
		}
		if i < len(dxs) {
			svg.cursorDPosition.x += dxs[i]
		}
		if i < len(dys) {
			svg.cursorDPosition.y += dys[i]
		}

		splitted := text.SplitFirstLine(string(r), t.style, svg.textContext, pr.Inf, false, true)
		layout := splitted.Layout
		width, height = Fl(splitted.Width), Fl(splitted.Height)

		x, y := svg.cursorPosition.x, svg.cursorPosition.y
		if hasX {
			x = xs[i]
		}
		if hasY {
			y = ys[i]
		}

		width *= scaleX
		if i != 0 {
			x += letterSpacing
		}
		svg.cursorPosition = point{x + width, y}

		xPosition := x + svg.cursorDPosition.x
		yPosition := y + svg.cursorDPosition.y + yAlign

		pointsBb := Rectangle{
			xPosition, yPosition,
			width, -height,
		}
		if i == 0 {
			bbox = pointsBb
		} else {
			bbox.union(pointsBb)
		}

		layout.ApplyJustification()

		doFill, doStroke := svg.setupPaint(dst, &svgNode{graphicContent: t, attributes: *attrs}, dims)
		dst.State().SetTextPaint(newPaintOp(doFill, doStroke, false))
		texts = append(texts,
			drawer.CreateFirstLine(layout, "none", pr.TaggedString{Tag: pr.None}, scaleX, xPosition, yPosition, angle))
	}

	dst.OnNewStack(func() {
		dst.DrawText(texts)
	})

	return nil
}
