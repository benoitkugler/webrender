package images

import (
	"fmt"
	"math"
	"strings"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/svg"
	"github.com/benoitkugler/webrender/utils"
)

// Gradient line size: distance between the starting point and ending point.
// Positions: list of Dimension in px or % (possibliy zero)
// 0 is the starting point, 1 the ending point.
// http://dev.w3.org/csswg/css-images-3/#color-stop-syntax
// Return processed color stops, as a list of floats in px.
func processColorStops(gradientLineSize pr.Float, positions_ []pr.Dimension) []pr.Fl {
	L := len(positions_)
	positions := make([]pr.MaybeFloat, L)
	for i, position := range positions_ {
		positions[i] = pr.ResoudPercentage(position.ToValue(), gradientLineSize)
	}
	// First and last default to 100%
	if positions[0] == nil {
		positions[0] = pr.Float(0)
	}
	if positions[L-1] == nil {
		positions[L-1] = gradientLineSize
	}

	// Make sure positions are increasing.
	previousPos := positions[0].V()
	for i, position := range positions {
		if position != nil {
			if position.V() < previousPos {
				positions[i] = previousPos
			} else {
				previousPos = position.V()
			}
		}
	}

	// Assign missing values
	previousI := L - 1
	for i, position := range positions {
		if position != nil {
			base := positions[previousI]
			increment := (position.V() - base.V()) / pr.Float(i-previousI)
			for j := previousI + 1; j < i; j += 1 {
				positions[j] = base.V() + pr.Float(j)*increment
			}
			previousI = i
		}
	}
	out := make([]pr.Fl, L)
	for i, v := range positions {
		out[i] = pr.Fl(v.V())
	}
	return out
}

// http://dev.w3.org/csswg/css-images-3/#find-the-average-color-of-a-gradient
func gradientAverageColor(colors []Color, positions []pr.Fl) Color {
	nbStops := len(positions)
	if nbStops <= 1 || nbStops != len(colors) {
		panic(fmt.Sprintf("expected same length, at least 2, got %d, %d", nbStops, len(colors)))
	}
	totalLength := positions[nbStops-1] - positions[0]
	if totalLength == 0 {
		for i := range positions {
			positions[i] = pr.Fl(i)
		}
		totalLength = pr.Fl(nbStops - 1)
	}
	premulR := make([]utils.Fl, nbStops)
	premulG := make([]utils.Fl, nbStops)
	premulB := make([]utils.Fl, nbStops)
	alpha := make([]utils.Fl, nbStops)
	for i, col := range colors {
		premulR[i] = col.R * col.A
		premulG[i] = col.G * col.A
		premulB[i] = col.B * col.A
		alpha[i] = col.A
	}
	var resultR, resultG, resultB, resultA utils.Fl
	totalWeight := 2 * totalLength
	for i_, position := range positions[1:] {
		i := i_ + 1
		weight := utils.Fl((position - positions[i-1]) / totalWeight)
		j := i - 1
		resultR += premulR[j] * weight
		resultG += premulG[j] * weight
		resultB += premulB[j] * weight
		resultA += alpha[j] * weight
		j = i
		resultR += premulR[j] * weight
		resultG += premulG[j] * weight
		resultB += premulB[j] * weight
		resultA += alpha[j] * weight
	}
	// Un-premultiply:
	if resultA != 0 {
		return Color{
			R: resultR / resultA,
			G: resultG / resultA,
			B: resultB / resultA,
			A: resultA,
		}
	}
	return Color{}
}

type layouter interface {
	// width, height: Gradient box. Top-left is at coordinates (0, 0).
	Layout(width, height pr.Float) backend.GradientLayout
}

type gradient struct {
	layouter

	colors        []Color
	stopPositions []pr.Dimension
	repeating     bool
}

func newGradient(colorStops []pr.ColorStop, repeating bool) gradient {
	self := gradient{}
	self.colors = make([]Color, len(colorStops))
	self.stopPositions = make([]pr.Dimension, len(colorStops))
	for i, v := range colorStops {
		self.colors[i] = v.Color.RGBA
		self.stopPositions[i] = v.Position
	}
	self.repeating = repeating
	return self
}

func (g gradient) GetIntrinsicSize(_, _ pr.Float) (pr.MaybeFloat, pr.MaybeFloat, pr.MaybeFloat) {
	// Gradients are not affected by image resolution, parent or font size.
	return nil, nil, nil
}

func (g gradient) Draw(dst backend.CanvasNoFill, concreteWidth, concreteHeight pr.Fl, imageRendering string) {
	layout := g.layouter.Layout(pr.Float(concreteWidth), pr.Float(concreteHeight))
	layout.Reapeating = g.repeating

	if layout.Kind == "solid" {
		dst.Rectangle(0, 0, concreteWidth, concreteHeight)
		dst.SetColorRgba(layout.Colors[0], false)
		dst.Fill(false)
		return
	}

	dst.DrawGradient(layout, concreteWidth, concreteHeight)
}

type LinearGradient struct {
	direction pr.DirectionType
	gradient
}

func (LinearGradient) isImage() {}

func NewLinearGradient(from pr.LinearGradient) LinearGradient {
	self := LinearGradient{gradient: newGradient(from.ColorStops, from.Repeating)}
	self.layouter = &self
	// ("corner", keyword) or ("angle", radians)
	self.direction = from.Direction
	return self
}

func toData(c Color) [6]pr.Fl {
	return [6]pr.Fl{pr.Fl(c.R), pr.Fl(c.G), pr.Fl(c.B), pr.Fl(c.A), 0, 0}
}

func (self LinearGradient) Layout(width, height pr.Float) backend.GradientLayout {
	// Only one color, render the gradient as a solid color
	if len(self.colors) == 1 {
		return backend.GradientLayout{ScaleY: 1, GradientKind: backend.GradientKind{Kind: "solid"}, Colors: []parser.RGBA{self.colors[0]}}
	}
	// (dx, dy) is the unit vector giving the direction of the gradient.
	// Positive dx: right, positive dy: down.
	var dx, dy pr.Fl
	if self.direction.Corner != "" {
		var factorX, factorY pr.Float
		switch self.direction.Corner {
		case "top_left":
			factorX, factorY = -1, -1
		case "top_right":
			factorX, factorY = 1, -1
		case "bottom_left":
			factorX, factorY = -1, 1
		case "bottom_right":
			factorX, factorY = 1, 1
		}
		diagonal := pr.Hypot(width, height)
		// Note the direction swap: dx based on height, dy based on width
		// The gradient line is perpendicular to a diagonal.
		dx = pr.Fl(factorX * height / diagonal)
		dy = pr.Fl(factorY * width / diagonal)
	} else {
		angle := float64(self.direction.Angle) // 0 upwards, then clockwise
		dx = pr.Fl(math.Sin(angle))
		dy = pr.Fl(-math.Cos(angle))
	}

	// Round dx and dy to avoid floating points errors caused by
	// trigonometry and angle units conversions
	dx, dy = utils.Round(dx), utils.Round(dy)

	// Distance between center && ending point,
	// ie. half of between the starting point && ending point :
	colors := self.colors
	vectorLength := pr.Fl(pr.Abs(width*pr.Float(dx)) + pr.Abs(height*pr.Float(dy)))
	positions := processColorStops(pr.Float(vectorLength), self.stopPositions)

	if !self.repeating {
		// Add explicit colors at boundaries if needed, because PDF doesn’t
		// extend color stops that are not displayed
		if positions[0] == positions[1] {
			positions = append([]pr.Fl{positions[0] - 1}, positions...)
			colors = append([]parser.RGBA{colors[0]}, colors...)
		}
		if positions[len(positions)-2] == positions[len(positions)-1] {
			positions = append(positions, positions[len(positions)-1]+1)
			colors = append(colors, colors[len(colors)-1])
		}
	}

	spread := svg.NoRepeat
	if self.repeating {
		spread = svg.Repeat
	}
	startX := (pr.Fl(width) - dx*vectorLength) / 2
	startY := (pr.Fl(height) - dy*vectorLength) / 2

	return spread.LinearGradient(positions, colors, startX, startY, dx, dy, vectorLength)
}

type RadialGradient struct {
	gradient
	shape  string
	size   pr.GradientSize
	center pr.Center
}

func (RadialGradient) isImage() {}

func NewRadialGradient(from pr.RadialGradient) RadialGradient {
	self := RadialGradient{gradient: newGradient(from.ColorStops, from.Repeating)}
	self.layouter = &self
	//  Type of ending shape: "circle" || "ellipse"
	self.shape = from.Shape
	// sizeType: "keyword"
	//   size: "closest-corner", "farthest-corner",
	//         "closest-side", || "farthest-side"
	// sizeType: "explicit"
	//   size: (radiusX, radiusY)
	self.size = from.Size
	// Center of the ending shape. (originX, posX, originY, posY)
	self.center = from.Center
	return self
}

func handleDegenerateRadial(sizeX, sizeY pr.Float) (pr.Float, pr.Float) {
	// http://dev.w3.org/csswg/css-images-3/#degenerate-radials
	if sizeX == 0 && sizeY == 0 {
		sizeX = 1e-7
		sizeY = 1e-7
	} else if sizeX == 0 {
		sizeX = 1e-7
		sizeY = 1e7
	} else if sizeY == 0 {
		sizeX = 1e7
		sizeY = 1e-7
	}
	return sizeX, sizeY
}

func (self RadialGradient) Layout(width, height pr.Float) backend.GradientLayout {
	if len(self.colors) == 1 {
		return backend.GradientLayout{ScaleY: 1, GradientKind: backend.GradientKind{Kind: "solid"}, Colors: []parser.RGBA{self.colors[0]}}
	}
	originX, centerX_, originY, centerY_ := self.center.OriginX, self.center.Pos[0], self.center.OriginY, self.center.Pos[1]
	centerX := pr.ResoudPercentage(centerX_.ToValue(), width).V()
	centerY := pr.ResoudPercentage(centerY_.ToValue(), height).V()
	if originX == "right" {
		centerX = width - centerX
	}
	if originY == "bottom" {
		centerY = height - centerY
	}

	sizeX, sizeY := handleDegenerateRadial(self.resolveSize(width, height, centerX, centerY))
	scaleY := pr.Fl(sizeY / sizeX)

	colors := self.colors
	positions := processColorStops(sizeX, self.stopPositions)
	if !self.repeating {
		// Add explicit colors at boundaries if needed, because PDF doesn’t
		// extend color stops that are not displayed
		if positions[0] > 0 && positions[0] == positions[1] {
			positions = append([]pr.Fl{0}, positions...)
			colors = append([]parser.RGBA{colors[0]}, colors...)
		}
		if positions[len(positions)-2] == positions[len(positions)-1] {
			positions = append(positions, positions[len(positions)-1]+1)
			colors = append(colors, colors[len(colors)-1])
		}
	}

	if positions[0] < 0 {
		// PDF does not like negative radiuses,
		// shift into the positive realm.
		if self.repeating {
			// Add vector lengths to first position until positive
			vectorLength := positions[len(positions)-1] - positions[0]
			offset := vectorLength * pr.Fl(1+math.Floor(float64(-positions[0]/vectorLength)))
			for i, p := range positions {
				positions[i] = p + offset
			}
		} else {
			// only keep colors with position >= 0, interpolate if needed
			if positions[len(positions)-1] <= 0 {
				// All stops are negatives,
				// everything is "padded" with the last color.
				return backend.GradientLayout{ScaleY: 1, GradientKind: backend.GradientKind{Kind: "solid"}, Colors: []parser.RGBA{self.colors[len(self.colors)-1]}}
			}

			for i, position := range positions {
				if position == 0 {
					// Keep colors and positions from this rank
					colors, positions = colors[i:], positions[i:]
					break
				}

				if position > 0 {
					// Interpolate with the previous to get the color at 0.
					color := colors[i]
					negColor := colors[i-1]
					negPosition := positions[i-1]
					if negPosition >= 0 {
						panic(fmt.Sprintf("expected non positive negPosition, got %f", negPosition))
					}
					intermediateColor := gradientAverageColor(
						[]Color{negColor, negColor, color, color},
						[]pr.Fl{negPosition, 0, 0, position})
					colors = append([]Color{intermediateColor}, colors[i:]...)
					positions = append([]pr.Fl{0}, positions[i:]...)
					break
				}
			}

		}
	}

	spread := svg.NoRepeat
	if self.repeating {
		spread = svg.Repeat
	}

	// spread.RadialGradient works with absolute lengths: apply scaleY
	fx := pr.Fl(centerX)
	fy := pr.Fl(centerY) / scaleY
	cx, cy := fx, fy
	var fr, r pr.Fl = 0, 1
	out := spread.RadialGradient(positions, colors, fx, fy, fr, cx, cy, r, pr.Fl(width)/scaleY, pr.Fl(height)/scaleY)

	out.ScaleY = scaleY // restore the scale
	return out
}

func (self RadialGradient) resolveSize(width, height, centerX, centerY pr.Float) (pr.Float, pr.Float) {
	if self.size.IsExplicit() {
		sizeX, sizeY := self.size.Explicit[0], self.size.Explicit[1]
		sizeX_ := pr.ResoudPercentage(sizeX.ToValue(), width).V()
		sizeY_ := pr.ResoudPercentage(sizeY.ToValue(), height).V()
		return sizeX_, sizeY_
	}
	left := pr.Abs(centerX)
	right := pr.Abs(width - centerX)
	top := pr.Abs(centerY)
	bottom := pr.Abs(height - centerY)
	pick := pr.Maxs
	if strings.HasPrefix(self.size.Keyword, "closest") {
		pick = pr.Mins
	}
	if strings.HasSuffix(self.size.Keyword, "side") {
		if self.shape == "circle" {
			sizeXy := pick(left, right, top, bottom)
			return sizeXy, sizeXy
		}
		// else: ellipse
		return pick(left, right), pick(top, bottom)
	}
	// else: corner
	if self.shape == "circle" {
		sizeXy := pick(pr.Hypot(left, top), pr.Hypot(left, bottom),
			pr.Hypot(right, top), pr.Hypot(right, bottom))
		return sizeXy, sizeXy
	}
	// else: ellipse
	keys := [4]pr.Float{pr.Hypot(left, top), pr.Hypot(left, bottom), pr.Hypot(right, top), pr.Hypot(right, bottom)}
	m := map[pr.Float][2]pr.Float{
		keys[0]: {left, top},
		keys[1]: {left, bottom},
		keys[2]: {right, top},
		keys[3]: {right, bottom},
	}
	c := m[pick(keys[0], keys[1], keys[2], keys[3])]
	cornerX, cornerY := c[0], c[1]
	return cornerX * pr.Float(math.Sqrt(2)), cornerY * pr.Float(math.Sqrt(2))
}
