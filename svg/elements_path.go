package svg

// handle <path d="XXX"> elements
// adapted from https://github.com/srwiley/oksvg

import (
	"errors"
	"log"
	"math"
	"unicode"

	"github.com/benoitkugler/webrender/backend"
)

var errParamMismatch = errors.New("svg path: param mismatch")

type vertex struct {
	x, y Fl // point position
	// orientation given by the shape, used
	// with orient="auto" or "auto-start-reverse"
	angle Fl
}

type point struct {
	x, y Fl
}

type pathItem struct {
	op   pathOperation
	args [3]point // (x, y) up to three arguments
}

type pathOperation uint8

const (
	// moveTo moves the current point.
	moveTo pathOperation = iota
	// lineTo draws a line from the current point, and updates it.
	lineTo
	// cubicTo draws a cubic Bezier curve from the current point, and updates it.
	cubicTo

	// NOTE: to simplify the backend requirement, we do not use
	// quadratic curves
)

func (pi pathItem) draw(dst backend.GraphicTarget) {
	switch pi.op {
	case moveTo:
		dst.MoveTo(pi.args[0].x, pi.args[0].y)
	case lineTo:
		dst.LineTo(pi.args[0].x, pi.args[0].y)
	case cubicTo:
		dst.CubicTo(pi.args[0].x, pi.args[0].y, pi.args[1].x, pi.args[1].y, pi.args[2].x, pi.args[2].y)
	}
}

// convert to a cubic
// CP1 = QP0 + 2/3 *(QP1-QP0)
// CP2 = QP2 + 2/3 *(QP1-QP2)
func quadraticToCubic(x0, y0, x1, y1, x2, y2 Fl) [3]point {
	cp1 := point{x0 + 2/3*(x1-x0), y0 + 2/3*(y1-y0)}
	cp2 := point{x2 + 2/3*(x1-x2), y2 + 2/3*(y1-y2)}
	cp3 := point{x2, y2}
	return [3]point{cp1, cp2, cp3}
}

// pathParser is used to parse SVG format path strings into a Path
type pathParser struct {
	path []pathItem // currently parsed

	points                 []Fl
	currentX, currentY     Fl // current position after the last drawing operation
	cntlPtX, cntlPtY       Fl
	pathStartX, pathStartY Fl
	lastKey                uint8
	inPath                 bool
}

func (c *pathParser) reset() {
	c.currentX = 0
	c.currentY = 0
	c.points = c.points[0:0]
	c.lastKey = ' '
	c.path = c.path[:0]
	c.inPath = false
}

// parsePath translates the svgPath description string into a path.
func (c *pathParser) parsePath(svgPath string) ([]pathItem, error) {
	c.reset()
	lastIndex := -1
	for i, v := range svgPath {
		if unicode.IsLetter(v) && v != 'e' {
			if lastIndex != -1 {
				if err := c.addSeg([]byte(svgPath[lastIndex:i])); err != nil {
					return nil, err
				}
			}
			lastIndex = i
		}
	}
	if lastIndex != -1 {
		if err := c.addSeg([]byte(svgPath[lastIndex:])); err != nil {
			return nil, err
		}
	}
	// return a copy so that the subsequent calls wont
	// alter the path

	return append([]pathItem(nil), c.path...), nil
}

func (c *pathParser) close() {
	c.path = append(c.path, pathItem{op: lineTo, args: [3]point{{c.pathStartX, c.pathStartY}}})
}

func (c *pathParser) moveTo(x, y Fl) {
	c.path = append(c.path, pathItem{op: moveTo, args: [3]point{{x, y}}})
}

func (c *pathParser) lineTo(x, y Fl) {
	c.path = append(c.path, pathItem{op: lineTo, args: [3]point{{x, y}}})
}

// also update the current point
func (c *pathParser) quadTo(x1, y1, x2, y2 Fl) {
	c.path = append(c.path, pathItem{op: cubicTo, args: quadraticToCubic(c.currentX, c.currentY, x1, y1, x2, y2)})
	c.currentX, c.currentY = x2, y2
}

func (c *pathParser) cubicTo(x1, y1, x2, y2, x3, y3 Fl) {
	c.path = append(c.path, pathItem{op: cubicTo, args: [3]point{
		{x1, y1}, {x2, y2}, {x3, y3},
	}})
}

func reflection(px, py, rx, ry Fl) (x, y Fl) {
	return px*2 - rx, py*2 - ry
}

func (c *pathParser) valsToAbs(last Fl) {
	for i := 0; i < len(c.points); i++ {
		last += c.points[i]
		c.points[i] = last
	}
}

func (c *pathParser) pointsToAbs(sz int) {
	lastX := c.currentX
	lastY := c.currentY
	for j := 0; j < len(c.points); j += sz {
		for i := 0; i < sz; i += 2 {
			c.points[i+j] += lastX
			c.points[i+1+j] += lastY
		}
		lastX = c.points[(j+sz)-2]
		lastY = c.points[(j+sz)-1]
	}
}

func (c *pathParser) hasSetsOrMore(sz int, rel bool) bool {
	if !(len(c.points) >= sz && len(c.points)%sz == 0) {
		return false
	}
	if rel {
		c.pointsToAbs(sz)
	}
	return true
}

// getPoints reads a set of floating point values from the SVG format number string,
// and add them to the cursor's points slice.
func (c *pathParser) getPoints(dataPoints []byte) (err error) {
	c.points = c.points[:0]
	c.points, err = parsePoints(string(dataPoints), c.points)
	return err
}

func (c *pathParser) reflectControlQuad() {
	switch c.lastKey {
	case 'q', 'Q', 'T', 't':
		c.cntlPtX, c.cntlPtY = reflection(c.currentX, c.currentY, c.cntlPtX, c.cntlPtY)
	default:
		c.cntlPtX, c.cntlPtY = c.currentX, c.currentY
	}
}

func (c *pathParser) reflectControlCube() {
	switch c.lastKey {
	case 'c', 'C', 's', 'S':
		c.cntlPtX, c.cntlPtY = reflection(c.currentX, c.currentY, c.cntlPtX, c.cntlPtY)
	default:
		c.cntlPtX, c.cntlPtY = c.currentX, c.currentY
	}
}

// addSeg decodes an SVG segment string into equivalent raster path commands saved
// in the cursor's Path
func (c *pathParser) addSeg(segString []byte) error {
	// Parse the string describing the numeric points in SVG format
	if err := c.getPoints(segString[1:]); err != nil {
		return err
	}
	l := len(c.points)
	k := segString[0]
	rel := false
	switch k {
	case 'z':
		fallthrough
	case 'Z':
		if len(c.points) != 0 {
			return errParamMismatch
		}
		if c.inPath {
			c.close()
			c.currentX = c.pathStartX
			c.currentY = c.pathStartY
			c.inPath = false
		}
	case 'm':
		rel = true
		fallthrough
	case 'M':
		if !c.hasSetsOrMore(2, rel) {
			return errParamMismatch
		}
		c.pathStartX, c.pathStartY = c.points[0], c.points[1]
		c.inPath = true
		c.moveTo(c.pathStartX, c.pathStartY)
		for i := 2; i < l-1; i += 2 {
			c.lineTo(c.points[i], c.points[i+1])
		}
		c.currentX = c.points[l-2]
		c.currentY = c.points[l-1]
	case 'l':
		rel = true
		fallthrough
	case 'L':
		if !c.hasSetsOrMore(2, rel) {
			return errParamMismatch
		}
		for i := 0; i < l-1; i += 2 {
			c.lineTo(c.points[i], c.points[i+1])
		}
		c.currentX = c.points[l-2]
		c.currentY = c.points[l-1]
	case 'v':
		c.valsToAbs(c.currentY)
		fallthrough
	case 'V':
		if !c.hasSetsOrMore(1, false) {
			return errParamMismatch
		}
		for _, p := range c.points {
			c.lineTo(c.currentX, p)
		}
		c.currentY = c.points[l-1]
	case 'h':
		c.valsToAbs(c.currentX)
		fallthrough
	case 'H':
		if !c.hasSetsOrMore(1, false) {
			return errParamMismatch
		}
		for _, p := range c.points {
			c.lineTo(p, c.currentY)
		}
		c.currentX = c.points[l-1]
	case 'q':
		rel = true
		fallthrough
	case 'Q':
		if !c.hasSetsOrMore(4, rel) {
			return errParamMismatch
		}
		for i := 0; i < l-3; i += 4 {
			c.quadTo(
				c.points[i], c.points[i+1],
				c.points[i+2], c.points[i+3],
			)
		}
		c.cntlPtX, c.cntlPtY = c.points[l-4], c.points[l-3]
		c.currentX = c.points[l-2]
		c.currentY = c.points[l-1]
	case 't':
		rel = true
		fallthrough
	case 'T':
		if !c.hasSetsOrMore(2, rel) {
			return errParamMismatch
		}
		for i := 0; i < l-1; i += 2 {
			c.reflectControlQuad()
			c.quadTo(
				c.cntlPtX, c.cntlPtY,
				c.points[i], c.points[i+1],
			)
			c.lastKey = k
		}
	case 'c':
		rel = true
		fallthrough
	case 'C':
		if !c.hasSetsOrMore(6, rel) {
			return errParamMismatch
		}
		for i := 0; i < l-5; i += 6 {
			c.cubicTo(
				c.points[i], c.points[i+1],
				c.points[i+2], c.points[i+3],
				c.points[i+4], c.points[i+5],
			)
		}
		c.cntlPtX, c.cntlPtY = c.points[l-4], c.points[l-3]
		c.currentX = c.points[l-2]
		c.currentY = c.points[l-1]
	case 's':
		rel = true
		fallthrough
	case 'S':
		if !c.hasSetsOrMore(4, rel) {
			return errParamMismatch
		}
		for i := 0; i < l-3; i += 4 {
			c.reflectControlCube()
			c.cubicTo(
				c.cntlPtX, c.cntlPtY,
				c.points[i], c.points[i+1],
				c.points[i+2], c.points[i+3],
			)
			c.lastKey = k
			c.cntlPtX, c.cntlPtY = c.points[i], c.points[i+1]
			c.currentX = c.points[i+2]
			c.currentY = c.points[i+3]
		}
	case 'a', 'A':
		if !c.hasSetsOrMore(7, false) {
			return errParamMismatch
		}
		for i := 0; i < l-6; i += 7 {
			if k == 'a' {
				c.points[i+5] += c.currentX
				c.points[i+6] += c.currentY
			}
			c.addArcFromA(c.points[i:])
		}
	default:
		log.Println("Ignoring svg command " + string(k))
	}
	// So we know how to extend some segment types
	c.lastKey = k
	return nil
}

// // ellipseAt adds a path of an elipse centered at cx, cy of radius rx and ry
// // to the pathCursor
// func (c *pathCursor) ellipseAt(cx, cy, rx, ry Fl) {
// 	c.placeX, c.placeY = cx+rx, cy
// 	c.points = c.points[0:0]
// 	c.points = append(c.points, rx, ry, 0.0, 1.0, 0.0, c.placeX, c.placeY)
// 	c.moveTo(c.placeX, c.placeY)
// 	c.placeX, c.placeY = c.addArc(c.points, cx, cy, c.placeX, c.placeY)
// 	c.close()
// }

// addArcFromA adds a path of an arc element to the cursor path to the pathCursor
func (c *pathParser) addArcFromA(points []Fl) {
	ra, rb := float64(points[0]), float64(points[1])
	cx, cy := findEllipseCenter(&ra, &rb, float64(points[2])*math.Pi/180, float64(c.currentX),
		float64(c.currentY), float64(points[5]), float64(points[6]), points[4] == 0, points[3] == 0)
	points[0], points[1] = Fl(ra), Fl(rb)

	c.currentX, c.currentY = c.addArc(c.points, Fl(cx), Fl(cy), c.currentX, c.currentY)
}

// addArc adds an arc to the adder p
func (p *pathParser) addArc(points []Fl, cx, cy, px, py Fl) (lx, ly Fl) {
	// maxDx is the maximum radians a cubic splice is allowed to span
	// in ellipse parametric when approximating an off-axis ellipse.
	const maxDx = math.Pi / 8

	rotX := float64(points[2]) * math.Pi / 180 // Convert degress to radians
	largeArc := points[3] != 0
	sweep := points[4] != 0
	startAngle := math.Atan2(float64(py-cy), float64(px-cx)) - rotX
	endAngle := math.Atan2(float64(points[6]-cy), float64(points[5]-cx)) - rotX
	deltaTheta := endAngle - startAngle
	arcBig := math.Abs(deltaTheta) > math.Pi

	// Approximate ellipse using cubic bezeir splines
	etaStart := math.Atan2(math.Sin(startAngle)/float64(points[1]), math.Cos(startAngle)/float64(points[0]))
	etaEnd := math.Atan2(math.Sin(endAngle)/float64(points[1]), math.Cos(endAngle)/float64(points[0]))
	deltaEta := etaEnd - etaStart
	if (arcBig && !largeArc) || (!arcBig && largeArc) { // Go has no boolean XOR
		if deltaEta < 0 {
			deltaEta += math.Pi * 2
		} else {
			deltaEta -= math.Pi * 2
		}
	}
	// This check might be needed if the center point of the elipse is
	// at the midpoint of the start and end lines.
	if deltaEta < 0 && sweep {
		deltaEta += math.Pi * 2
	} else if deltaEta >= 0 && !sweep {
		deltaEta -= math.Pi * 2
	}

	// Round up to determine number of cubic splines to approximate bezier curve
	segs := int(math.Abs(deltaEta)/maxDx) + 1
	dEta := deltaEta / float64(segs) // span of each segment
	// Approximate the ellipse using a set of cubic bezier curves by the method of
	// L. Maisonobe, "Drawing an elliptical arc using polylines, quadratic
	// or cubic Bezier curves", 2003
	// https://www.spaceroots.org/documents/elllipse/elliptical-arc.pdf
	tde := math.Tan(dEta / 2)
	alpha := Fl(math.Sin(dEta) * (math.Sqrt(4+3*tde*tde) - 1) / 3) // Math is fun!
	lx, ly = px, py
	sinTheta, cosTheta := math.Sin(rotX), math.Cos(rotX)
	ldx, ldy := ellipsePrime(float64(points[0]), float64(points[1]), sinTheta, cosTheta, etaStart)
	for i := 1; i <= segs; i++ {
		eta := etaStart + dEta*float64(i)
		var px, py Fl
		if i == segs {
			px, py = points[5], points[6] // Just makes the end point exact; no roundoff error
		} else {
			px, py = ellipsePointAt(float64(points[0]), float64(points[1]), sinTheta, cosTheta, eta, float64(cx), float64(cy))
		}
		dx, dy := ellipsePrime(float64(points[0]), float64(points[1]), sinTheta, cosTheta, eta)
		p.cubicTo(lx+alpha*ldx, ly+alpha*ldy, px-alpha*dx, py-alpha*dy, px, py)
		lx, ly, ldx, ldy = px, py, dx, dy
	}
	return lx, ly
}

// ellipsePrime gives tangent vectors for parameterized elipse; a, b, radii, eta parameter, center cx, cy
func ellipsePrime(a, b, sinTheta, cosTheta, eta float64) (px, py Fl) {
	bCosEta := b * math.Cos(eta)
	aSinEta := a * math.Sin(eta)
	px = Fl(-aSinEta*cosTheta - bCosEta*sinTheta)
	py = Fl(-aSinEta*sinTheta + bCosEta*cosTheta)
	return
}

// ellipsePointAt gives points for parameterized elipse; a, b, radii, eta parameter, center cx, cy
func ellipsePointAt(a, b, sinTheta, cosTheta, eta, cx, cy float64) (px, py Fl) {
	aCosEta := a * math.Cos(eta)
	bSinEta := b * math.Sin(eta)
	px = Fl(cx + aCosEta*cosTheta - bSinEta*sinTheta)
	py = Fl(cy + aCosEta*sinTheta + bSinEta*cosTheta)
	return
}

// findEllipseCenter locates the center of the Ellipse if it exists. If it does not exist,
// the radius values will be increased minimally for a solution to be possible
// while preserving the ra to rb ratio.  ra and rb arguments are pointers that can be
// checked after the call to see if the values changed. This method uses coordinate transformations
// to reduce the problem to finding the center of a circle that includes the origin
// and an arbitrary point. The center of the circle is then transformed
// back to the original coordinates and returned.
func findEllipseCenter(ra, rb *float64, rotX, startX, startY, endX, endY float64, sweep, smallArc bool) (cx, cy float64) {
	cos, sin := math.Cos(rotX), math.Sin(rotX)

	// Move origin to start point
	nx, ny := endX-startX, endY-startY

	// Rotate ellipse x-axis to coordinate x-axis
	nx, ny = nx*cos+ny*sin, -nx*sin+ny*cos
	// Scale X dimension so that ra = rb
	nx *= *rb / *ra // Now the ellipse is a circle radius rb; therefore foci and center coincide

	midX, midY := nx/2, ny/2
	midlenSq := midX*midX + midY*midY

	var hr float64
	if *rb**rb < midlenSq {
		// Requested ellipse does not exist; scale ra, rb to fit. Length of
		// span is greater than max width of ellipse, must scale *ra, *rb
		nrb := math.Sqrt(midlenSq)
		if *ra == *rb {
			*ra = nrb // prevents roundoff
		} else {
			*ra = *ra * nrb / *rb
		}
		*rb = nrb
	} else {
		hr = math.Sqrt(*rb**rb-midlenSq) / math.Sqrt(midlenSq)
	}
	// Notice that if hr is zero, both answers are the same.
	if (sweep && smallArc) || (!sweep && !smallArc) {
		cx = midX + midY*hr
		cy = midY - midX*hr
	} else {
		cx = midX - midY*hr
		cy = midY + midX*hr
	}

	// reverse scale
	cx *= *ra / *rb
	// Reverse rotate and translate back to original coordinates
	return cx*cos - cy*sin + startX, cx*sin + cy*cos + startY
}
