package svg

import (
	"math"

	"github.com/benoitkugler/webrender/utils"
)

// if withStroke is true, add the stroke shape
func (node *svgNode) resolveBoundingBox(dims drawingDims, withStroke bool) (Rectangle, bool) {
	if node.graphicContent == nil {
		return Rectangle{}, false
	}
	bbox, ok := node.graphicContent.boundingBox(&node.attributes, dims)
	if !ok {
		return Rectangle{}, false
	}

	if withStroke && node.attributes.stroke.valid {
		strokeWidth := dims.length(node.attributes.strokeWidth)
		bbox.X -= strokeWidth / 2
		bbox.Y -= strokeWidth / 2
		bbox.Width += strokeWidth
		bbox.Height += strokeWidth
	}

	return bbox, true
}

// compute the union of all the node bounding boxes or return false if
// no one is valid
func boundingBoxUnion(nodes []*svgNode, dims drawingDims) (Rectangle, bool) {
	foundValid := false
	out := Rectangle{}
	for _, node := range nodes {
		bb, ok := node.resolveBoundingBox(dims, true)
		if !ok {
			continue
		}

		if !foundValid { // start with this box
			out = bb
			foundValid = true
		} else { // perform the union
			out.union(bb)
		}
	}
	return out, foundValid
}

type Rectangle struct {
	X, Y, Width, Height Fl
}

// increase the rectangle to contain (x,y)
func (r *Rectangle) add(x, y Fl) {
	minX, minY := utils.MinF(r.X, x), utils.MinF(r.Y, y)
	maxX, maxY := utils.MaxF(r.X+r.Width, x), utils.MaxF(r.Y+r.Height, y)
	r.X, r.Y, r.Width, r.Height = minX, minY, maxX-minX, maxY-minY
}

// increase `r` to contain `other`
func (r *Rectangle) union(other Rectangle) {
	r.add(other.X, other.Y)
	r.add(other.X+other.Width, other.Y+other.Height)
}

func (r rect) boundingBox(attrs *attributes, dims drawingDims) (Rectangle, bool) {
	x, y := dims.point(attrs.x, attrs.y)
	width, height := dims.point(attrs.width, attrs.height)
	return Rectangle{x, y, width, height}, true
}

func (e ellipse) boundingBox(_ *attributes, dims drawingDims) (Rectangle, bool) {
	rx, ry := dims.point(e.rx, e.ry)
	cx, cy := dims.point(e.cx, e.cy)
	return Rectangle{cx - rx, cy - ry, 2 * rx, 2 * ry}, true
}

func (l line) boundingBox(_ *attributes, dims drawingDims) (Rectangle, bool) {
	x1, y1 := dims.point(l.x1, l.y1)
	x2, y2 := dims.point(l.x2, l.y2)
	x, y := utils.MinF(x1, x2), utils.MinF(y1, y2)
	width, height := utils.MaxF(x1, x2)-x, utils.MaxF(y1, y2)-y
	return Rectangle{x, y, width, height}, true
}

func (pl polyline) boundingBox(_ *attributes, dims drawingDims) (Rectangle, bool) {
	if len(pl.points) == 0 {
		return Rectangle{}, false
	}
	bbox := Rectangle{X: pl.points[0].x, Y: pl.points[0].y}
	for _, point := range pl.points[1:] {
		bbox.add(point.x, point.y)
	}
	return bbox, true
}

func (p path) boundingBox(_ *attributes, _ drawingDims) (Rectangle, bool) {
	if len(p) == 0 {
		return Rectangle{}, false
	}
	boundingBox, currentPoint := p[0].boundingBox(point{})
	for _, item := range p[1:] {
		var itemBoudingBox Rectangle
		itemBoudingBox, currentPoint = item.boundingBox(currentPoint)
		boundingBox.union(itemBoudingBox)
	}
	return boundingBox, true
}

func (image) boundingBox(_ *attributes, _ drawingDims) (Rectangle, bool) {
	return Rectangle{}, false
}

func (svg) boundingBox(_ *attributes, _ drawingDims) (Rectangle, bool) {
	return Rectangle{}, false
}

func (span) boundingBox(_ *attributes, _ drawingDims) (Rectangle, bool) {
	return Rectangle{}, false
}

// bounding box for bezier curves

type lineBezier [2]point

func (l lineBezier) criticalPoints() (tX, tY []Fl) {
	return nil, nil
}

func evaluateBezierLine(p0, p1, t Fl) Fl {
	return (p1-p0)*t + p0
}

func (l lineBezier) evaluateCurve(t Fl) (x, y Fl) {
	p0x, p0y := l[0].x, l[0].y
	p1x, p1y := l[1].x, l[1].y

	return evaluateBezierLine(p0x, p1x, t), evaluateBezierLine(p0y, p1y, t)
}

type quadBezier [3]point

// quadratic polinomial
// x = At^2 + Bt + C
// where
// A = p0 + p2 - 2p1
// B = 2(p1 - p0)
// C = p0
func bezierQuad(p0, p1, p2, t Fl) Fl {
	return (p0+p2-2*p1)*t*t + 2*(p1-p0)*t + p0
}

// derivative as at + b where a,b :
func quadraticDerivative(p0, p1, p2 Fl) (a, b Fl) {
	return 2 * (p2 - p1 - (p1 - p0)), 2 * (p1 - p0)
}

// handle the case where a = 0
func linearRoots(a, b Fl) []Fl {
	if a == 0 {
		return nil
	}
	return []Fl{-b / a}
}

func (cu quadBezier) criticalPoints() (tX, tY []Fl) {
	p0x, p0y := cu[0].x, cu[0].y
	p1x, p1y := cu[1].x, cu[1].y
	p2x, p2y := cu[2].x, cu[2].y

	aX, bX := quadraticDerivative(p0x, p1x, p2x)
	aY, bY := quadraticDerivative(p0y, p1y, p2y)

	return linearRoots(aX, bX), linearRoots(aY, bY)
}

func (cu quadBezier) evaluateCurve(t Fl) (x, y Fl) {
	p0x, p0y := cu[0].x, cu[0].y
	p1x, p1y := cu[1].x, cu[1].y
	p2x, p2y := cu[2].x, cu[2].y
	return bezierQuad(p0x, p1x, p2x, t), bezierQuad(p0y, p1y, p2y, t)
}

type cubicBezier [4]point

func (cu cubicBezier) criticalPoints() (tX, tY []Fl) {
	p1x, p1y := cu[0].x, cu[0].y
	c1x, c1y := cu[1].x, cu[1].y
	c2x, c2y := cu[2].x, cu[2].y
	p2x, p2y := cu[3].x, cu[3].y

	aX, bX, cX := cubicDerivative(p1x, c1x, c2x, p2x)
	aY, bY, cY := cubicDerivative(p1y, c1y, c2y, p2y)

	return quadraticRoots(aX, bX, cX), quadraticRoots(aY, bY, cY)
}

func (cu cubicBezier) evaluateCurve(t Fl) (x, y Fl) {
	p0x, p0y := cu[0].x, cu[0].y
	p1x, p1y := cu[1].x, cu[1].y
	p2x, p2y := cu[2].x, cu[2].y
	p3x, p3y := cu[3].x, cu[3].y
	return bezierSpline(p0x, p1x, p2x, p3x, t), bezierSpline(p0y, p1y, p2y, p3y, t)
}

// cubic polinomial
// x = At^3 + Bt^2 + Ct + D
// where A,B,C,D:
// A = p3 -3 * p2 + 3 * p1 - p0
// B = 3 * p2 - 6 * p1 +3 * p0
// C = 3 * p1 - 3 * p0
// D = p0
func bezierSpline(p0, p1, p2, p3, t Fl) Fl {
	return (p3-3*p2+3*p1-p0)*t*t*t +
		(3*p2-6*p1+3*p0)*t*t +
		(3*p1-3*p0)*t +
		(p0)
}

// We would like to know the values of t where X = 0
// X  = (p3-3*p2+3*p1-p0)t^3 + (3*p2-6*p1+3*p0)t^2 + (3*p1-3*p0)t + (p0)
// Derivative :
// X' = 3(p3-3*p2+3*p1-p0)t^(3-1) + 2(6*p2-12*p1+6*p0)t^(2-1) + 1(3*p1-3*p0)t^(1-1)
// simplified:
// X' = (3*p3-9*p2+9*p1-3*p0)t^2 + (6*p2-12*p1+6*p0)t + (3*p1-3*p0)
// taken as aX^2 + bX + c  a,b and c are:
func cubicDerivative(p0, p1, p2, p3 Fl) (a, b, c Fl) {
	return 3*p3 - 9*p2 + 9*p1 - 3*p0, 6*p2 - 12*p1 + 6*p0, 3*p1 - 3*p0
}

// b^2 - 4ac = Determinant
func determinant(a, b, c Fl) Fl { return b*b - 4*a*c }

func solve(a, b, c Fl, s bool) Fl {
	var sign Fl = 1.
	if !s {
		sign = -1.
	}
	return (-b + (Fl(math.Sqrt(float64((b*b)-(4*a*c)))) * sign)) / (2 * a)
}

func quadraticRoots(a, b, c Fl) []Fl {
	d := determinant(a, b, c)
	if d < 0 {
		return nil
	}

	if a == 0 {
		// aX^2 + bX + c well then then this is a simple line
		// x= -c / b
		return []Fl{-c / b}
	}

	if d == 0 {
		return []Fl{solve(a, b, c, true)}
	} else {
		return []Fl{
			solve(a, b, c, true),
			solve(a, b, c, false),
		}
	}
}

type bezier interface {
	// compute the t zeroing the derivative
	criticalPoints() (tX, tY []Fl)
	// compute the point a time t
	evaluateCurve(t Fl) (x, y Fl)
}

func computeBezierBoundingBox(curve bezier) Rectangle {
	resX, resY := curve.criticalPoints()

	// draw min and max
	var bbox []point

	// add begin and end point
	for _, t := range append(append(resX, 0, 1), resY...) {
		// filter invalid value
		if !(0 <= t && t <= 1) {
			continue
		}
		x, y := curve.evaluateCurve(t)

		bbox = append(bbox, point{x, y})
	}

	// bbox is never empty since it always contains the begin and end point
	minX := bbox[0].x
	minY := bbox[0].y
	maxX, maxY := minX, minY

	for _, e := range bbox[1:] {
		minX = utils.MinF(e.x, minX)
		minY = utils.MinF(e.y, minY)
		maxX = utils.MaxF(e.x, maxX)
		maxY = utils.MaxF(e.y, maxY)
	}
	return Rectangle{minX, minY, maxX - minX, maxY - minY}
}

// return the bounding box for the segment and the updated current point
func (pi pathItem) boundingBox(currentPoint point) (Rectangle, point) {
	switch pi.op {
	case moveTo:
		newPoint := pi.args[0]
		return Rectangle{X: newPoint.x, Y: newPoint.y}, newPoint
	case lineTo, close:
		newPoint := pi.args[0]
		return computeBezierBoundingBox(lineBezier{currentPoint, newPoint}), newPoint
	case cubicTo:
		newPoint := pi.args[2]
		return computeBezierBoundingBox(cubicBezier{currentPoint, pi.args[0], pi.args[1], newPoint}), newPoint
	default:
		panic("unreachable")
	}
}
