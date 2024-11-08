package svg

import (
	"fmt"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/utils"
)

// this file exposes gradient functions
// shared by svg and css rendering engine.

func reverseColors(a []parser.RGBA) []parser.RGBA {
	n := len(a)
	out := make([]parser.RGBA, n)
	for i := range a {
		out[n-1-i] = a[i]
	}
	return out
}

func reverseFloats(a []Fl) []Fl {
	n := len(a)
	out := make([]Fl, n)
	for i := range a {
		out[n-1-i] = a[i]
	}
	return out
}

// Normalize to [0..1].
// Write on positions.
func normalizeStopPositions(positions []Fl) (Fl, Fl) {
	first := positions[0]
	last := positions[len(positions)-1]
	totalLength := last - first
	if totalLength != 0 {
		for i, pos := range positions {
			positions[i] = (pos - first) / totalLength
		}
	} else {
		for i := range positions {
			positions[i] = 0
		}
	}
	return first, last
}

// http://drafts.csswg.org/csswg/css-images-3/#find-the-average-color-of-a-gradient
func gradientAverageColor(colors []parser.RGBA, positions []Fl) parser.RGBA {
	nbStops := len(positions)
	if nbStops <= 1 || nbStops != len(colors) {
		panic(fmt.Sprintf("expected same length, at least 2, got %d, %d", nbStops, len(colors)))
	}
	totalLength := positions[nbStops-1] - positions[0]
	if totalLength == 0 {
		for i := range positions {
			positions[i] = Fl(i)
		}
		totalLength = Fl(nbStops - 1)
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
		return parser.RGBA{
			R: resultR / resultA,
			G: resultG / resultA,
			B: resultB / resultA,
			A: resultA,
		}
	}
	return parser.RGBA{}
}

// GradientSpread defines how a gradient should be repeated.
type GradientSpread uint8

const (
	NoRepeat GradientSpread = iota
	Repeat
	Reflect
)

// LinearGradient handle spread (repeat) for linear gradients
// It is used for SVG and CSS gradient rendering.
func (spread GradientSpread) LinearGradient(positions []Fl, colors []parser.RGBA, x1, y1, dx, dy, vectorLength Fl) backend.GradientLayout {
	first, last := normalizeStopPositions(positions)
	if spread != NoRepeat {
		// Render as a solid color if the first and last positions are equal
		// See https://drafts.csswg.org/css-images-3/#repeating-gradients
		if first == last {
			color := gradientAverageColor(colors, positions)
			return backend.GradientLayout{ScaleY: 1, GradientKind: backend.GradientKind{Kind: "solid"}, Colors: []parser.RGBA{color}}
		}

		// Define defined gradient length and steps between positions
		stopLength := last - first
		// assert stopLength > 0
		positionSteps := make([]Fl, len(positions)-1)
		for i := range positionSteps {
			positionSteps[i] = positions[i+1] - positions[i]
		}

		// Create cycles used to add colors
		var (
			nextSteps, previousSteps   []Fl
			nextColors, previousColors []parser.RGBA
		)
		if spread == Repeat {
			nextSteps = append([]Fl{0}, positionSteps...)
			nextColors = colors
			previousSteps = append([]Fl{0}, reverseFloats(positionSteps)...)
			previousColors = reverseColors(colors)
		} else { // Reflect
			nextSteps = append(append(append([]Fl{0}, reverseFloats(positionSteps)...), 0), positionSteps...)
			nextColors = append(reverseColors(colors), colors...)
			previousSteps = append(append(append([]Fl{0}, positionSteps...), 0), reverseFloats(positionSteps)...)
			previousColors = append(colors, reverseColors(colors)...)
		}

		// Add colors after last step
		for i := 0; last < vectorLength; i++ {
			step := nextSteps[i%len(nextSteps)]
			colors = append(colors, nextColors[i%len(nextColors)])
			positions = append(positions, positions[len(positions)-1]+step)
			last += step * stopLength
		}

		// Add colors before last step
		for i := 0; first > 0; i++ {
			step := previousSteps[i%len(previousSteps)]
			colors = append([]parser.RGBA{previousColors[i%len(previousColors)]}, colors...)
			positions = append([]Fl{positions[0] - step}, positions...)
			first -= step * stopLength
		}
	}

	points := [6]Fl{
		x1 + dx*first,
		y1 + dy*first,
		x1 + dx*last,
		y1 + dy*last,
	}
	return backend.GradientLayout{ScaleY: 1, GradientKind: backend.GradientKind{Kind: "linear", Coords: points}, Positions: positions, Colors: colors}
}

func (spread GradientSpread) RadialGradient(positions []Fl, colors []parser.RGBA, fx, fy, fr, cx, cy, r, width, height Fl) backend.GradientLayout {
	first, last := normalizeStopPositions(positions)

	if spread != NoRepeat && first == last {
		// Render as a solid color if the first and last positions are equal
		// See https://drafts.csswg.org/css-images-3/#repeating-gradients
		color := gradientAverageColor(colors, positions)
		return backend.GradientLayout{ScaleY: 1, GradientKind: backend.GradientKind{Kind: "solid"}, Colors: []parser.RGBA{color}}
	}

	// Define the coordinates of the gradient circles
	fr, r = fr+(r-fr)*first, fr+(r-fr)*last

	circles := [6]Fl{fx, fy, fr, cx, cy, r}

	if spread != NoRepeat {
		circles, positions, colors = spread.repeatRadial(width, height, circles, positions, colors)
	}

	return backend.GradientLayout{ScaleY: 1, GradientKind: backend.GradientKind{Kind: "radial", Coords: circles}, Positions: positions, Colors: colors}
}

// points = [fx, fy, fr, cx, cy, r]
func (spread GradientSpread) repeatRadial(width, height Fl, points [6]Fl, positions []Fl, colors []parser.RGBA) ([6]Fl, []Fl, []parser.RGBA) {
	// Keep original lists and values, they’re useful
	originalColors := append([]parser.RGBA{}, colors...)
	originalPositions := append([]Fl{}, positions...)
	gradientLength := points[5] - points[2]

	// Get the maximum distance between the center and the corners, to find
	// how many times we have to repeat the colors outside
	maxDistance := utils.Maxs(
		utils.Hypot(width-points[0], height-points[1]),
		utils.Hypot(width-points[0], -points[1]),
		utils.Hypot(-points[0], height-points[1]),
		utils.Hypot(-points[0], -points[1]),
	)
	repeatAfter := int(utils.Ceil((maxDistance - points[5]) / gradientLength))
	if repeatAfter > 0 {
		// Repeat colors and extrapolate positions
		repeat := 1 + repeatAfter

		colors = make([]parser.RGBA, len(colors)*repeat)
		reversedColors := reverseColors(originalColors)
		tmpPositions := make([]Fl, 0, len(positions)*repeat)
		for i := 0; i < repeat; i++ {
			// reverse originalColors for reflect and even index
			if spread == Reflect && i%2 == 0 { // Reflect
				copy(colors[i*len(originalColors):], reversedColors)
			} else {
				copy(colors[i*len(originalColors):], originalColors)
			}
			for _, position := range positions {
				tmpPositions = append(tmpPositions, Fl(i)+position)
			}
		}
		positions = tmpPositions

		points[5] = points[5] + gradientLength*Fl(repeatAfter)
	}

	if points[2] == 0 {
		// Inner circle has 0 radius, no need to repeat inside, return
		return points, positions, colors
	}

	// Find how many times we have to repeat the colors inside
	repeatBefore := points[2] / gradientLength

	// Set the inner circle size to 0
	points[2] = 0

	// Find how many times the whole gradient can be repeated
	fullRepeat := int(repeatBefore)
	if fullRepeat != 0 {
		// Repeat colors and extrapolate positions

		reversedColors := reverseColors(originalColors)
		positionsTmp := make([]Fl, 0, len(positions)+len(originalPositions)*fullRepeat)
		for i := 0; i < fullRepeat; i++ {
			// reverse originalColors for reflect and even index
			if spread == Reflect && i%2 == 0 { // Reflect
				colors = append(colors, reversedColors...)
			} else {
				colors = append(colors, originalColors...)
			}
			for _, position := range originalPositions {
				positionsTmp = append(positionsTmp, Fl(i-fullRepeat)+position)
			}
		}

		positions = append(positionsTmp, positions...)
	}

	// Find the ratio of gradient that must be added to reach the center
	partialRepeat := repeatBefore - Fl(fullRepeat)
	if partialRepeat == 0 {
		// No partial repeat, return
		return points, positions, colors
	}

	// Iterate through positions in reverse order, from the outer
	// circle to the original inner circle, to find positions from
	// the inner circle (including full repeats) to the center
	// assert (originalPositions[0], originalPositions[-1]) == (0, 1)
	// assert 0 < partialRepeat < 1
	reverse := reverseFloats(originalPositions)
	ratio := 1 - partialRepeat
	LC, LP := len(originalColors), len(originalPositions)

	for i_, position := range reverse {
		i := i_ + 1
		if position == ratio {
			// The center is a color of the gradient, truncate original
			// colors and positions and prepend them
			colors = append(originalColors[LC-i:], colors...)
			tmp := originalPositions[LP-i:]
			newPositions := make([]Fl, len(tmp))
			for j, position := range tmp {
				newPositions[j] = position - Fl(fullRepeat) - 1
			}
			positions = append(newPositions, positions...)
			return points, positions, colors
		}
		if position < ratio {
			// The center is between two colors of the gradient,
			// define the center color as the average of these two
			// gradient colors
			color := originalColors[LC-i]
			nextColor := originalColors[LC-(i-1)]
			nextPosition := originalPositions[LP-(i-1)]
			averageColors := []parser.RGBA{color, color, nextColor, nextColor}
			averagePositions := []Fl{position, ratio, ratio, nextPosition}
			zeroColor := gradientAverageColor(averageColors, averagePositions)
			colors = append(append([]parser.RGBA{zeroColor}, originalColors[LC-(i-1):]...), colors...)
			tmp := originalPositions[LP-(i-1):]
			newPositions := make([]Fl, len(tmp))
			for j, position := range tmp {
				newPositions[j] = position - Fl(fullRepeat) - 1
			}
			positions = append(append([]Fl{ratio - 1 - Fl(fullRepeat)}, newPositions...), positions...)
			return points, positions, colors
		}
	}

	return points, positions, colors
}
