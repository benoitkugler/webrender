// Package svg implements parsing of SVG images.
// It transforms SVG text files into an in-memory structure
// that is easy to draw.
// CSS is supported via the style and cascadia packages.
package svg

import (
	"fmt"
	"io"
	"math"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/utils"
	"golang.org/x/net/html"
)

// convert from an svg tree to the final form

// nodes that are not directly draw but may be referenced by ID
// from other nodes
type definitions struct {
	filters      map[string][]filter
	clipPaths    map[string]*clipPath
	masks        map[string]mask
	paintServers map[string]paintServer
	markers      map[string]*marker
	nodes        map[string]*svgNode
}

func newDefinitions() definitions {
	return definitions{
		filters:      make(map[string][]filter),
		clipPaths:    make(map[string]*clipPath),
		masks:        make(map[string]mask),
		paintServers: make(map[string]paintServer),
		markers:      make(map[string]*marker),
		nodes:        make(map[string]*svgNode),
	}
}

type SVGImage struct {
	root *svgNode

	definitions definitions
}

// DisplayedSize returns the value of the "width" and "height" attributes
// of the <svg> root element, which discribe the displayed size of the rectangular viewport.
// If no value is specified, it default to 100% (auto).
func (svg *SVGImage) DisplayedSize() (width, height Value) {
	w, h := svg.root.width, svg.root.height
	if w.U == 0 {
		w = Value{100, Perc}
	}
	if h.U == 0 {
		h = Value{100, Perc}
	}
	return w, h
}

// ViewBox returns the optional value of the "viewBox" attribute
func (svg *SVGImage) ViewBox() *Rectangle { return svg.root.viewbox }

// Draw draws the parsed SVG image into the given `dst` output,
// with the given `width` and `height`.
func (svg *SVGImage) Draw(dst backend.Canvas, width, height Fl) {
	var dims drawingDims

	dims.concreteWidth, dims.concreteHeight = width, height

	if vb := svg.ViewBox(); vb != nil {
		dims.innerWidth, dims.innerHeight = vb.Width, vb.Height
	} else {
		dims.innerWidth, dims.innerHeight = width, height
	}

	dims.fontSize = defaultFontSize

	dims.setupDiagonal()

	svg.drawNode(dst, svg.root, dims, true)
}

// if paint is false, only the path operations are executed, not the actual filling or drawing
// moreover, no new graphic stack is created
func (svg *SVGImage) drawNode(dst backend.Canvas, node *svgNode, dims drawingDims, paint bool) {
	dims.fontSize = node.attributes.fontSize.Resolve(dims.fontSize, dims.fontSize)

	paintTask := func() {
		// apply filters
		if filters := svg.definitions.filters[node.filterID]; filters != nil {
			applyFilters(dst, filters, node, dims)
		}

		// create sub group for opacity
		opacity := node.attributes.opacity
		var originalDst backend.Canvas
		if paint && 0 <= opacity && opacity < 1 {
			originalDst = dst
			var x, y, width, height Fl = 0, 0, dims.concreteWidth, dims.concreteHeight
			if box, ok := node.resolveBoundingBox(dims, true); ok {
				x, y, width, height = box.X, box.Y, box.Width, box.Height
			}
			dst = dst.NewGroup(x, y, width, height)
		}

		applyTransform(dst, node.attributes.transforms, dims)

		// clip
		if cp, has := svg.definitions.clipPaths[node.clipPathID]; has {
			svg.applyClipPath(dst, cp, node, dims)
		}

		// manage display and visibility
		display := node.attributes.display
		visible := node.attributes.visible

		// draw the node itself.
		var vertices []vertex
		if visible {
			// special case for <use> tags, which require the context
			if u, isUse := node.graphicContent.(use); isUse {
				svg.drawUse(dst, u, &node.attributes, dims)
			} else if node.graphicContent != nil {
				vertices = node.graphicContent.draw(dst, &node.attributes, dims)
			}
		}

		// then recurse
		if display {
			for _, child := range node.children {
				svg.drawNode(dst, child, dims, paint)
			}
		}

		// apply mask
		if ma, has := svg.definitions.masks[node.maskID]; has {
			svg.applyMask(dst, ma, node, dims)
		}

		// do the actual painting
		if paint {
			svg.paintNode(dst, node, dims)
		}

		// draw markers
		if len(vertices) != 0 {
			svg.drawMarkers(dst, vertices, node, dims, paint)
		}

		// apply opacity group and restore original target
		if paint && 0 <= opacity && opacity < 1 {
			originalDst.DrawWithOpacity(opacity, dst)
			dst = originalDst // actually not used
		}
	}

	if paint {
		dst.OnNewStack(paintTask)
	} else {
		paintTask()
	}
}

// vertices are the resolved vertices computed when drawing the shape
func (svg *SVGImage) drawMarkers(dst backend.Canvas, vertices []vertex, node *svgNode, dims drawingDims, paint bool) {
	const (
		start uint8 = iota
		mid
		end
	)
	commonMarker := svg.definitions.markers[node.markerID]

	// [start, mid, end] defautling to the common marker
	markers := [3]*marker{commonMarker, commonMarker, commonMarker}

	if marker := svg.definitions.markers[node.markerStartID]; marker != nil {
		markers[start] = marker
	}
	if marker := svg.definitions.markers[node.markerMidID]; marker != nil {
		markers[mid] = marker
	}
	if marker := svg.definitions.markers[node.markerEndID]; marker != nil {
		markers[end] = marker
	}

	for i, vertex := range vertices {
		position := mid
		if i == 0 {
			position = start
		} else if i == len(vertices)-1 {
			position = end
		}

		marker := markers[position]
		if marker == nil {
			continue
		}

		// calculate position, scale and clipping
		var (
			clipBox        *Rectangle
			scaleX, scaleY Fl
		)
		markerWidth, markerHeight := dims.point(marker.markerWidth, marker.markerHeight)
		translateX, translateY := dims.point(marker.refX, marker.refY)
		if vb := node.attributes.viewbox; vb != nil {
			scaleX, scaleY, translateX, translateY = marker.preserveAspectRatio.resolveTransforms(markerWidth, markerHeight, marker.viewbox, &point{translateX, translateY})

			clipViewbox := *vb
			if marker.viewbox != nil {
				clipViewbox = *marker.viewbox
			}

			xPosition, yPosition := marker.preserveAspectRatio.xPosition, marker.preserveAspectRatio.yPosition

			if xPosition == "mid" {
				clipViewbox.X += (clipViewbox.Width - markerWidth/scaleX) / 2
			} else if xPosition == "max" {
				clipViewbox.X += clipViewbox.Width - markerWidth/scaleX
			}

			if yPosition == "mid" {
				clipViewbox.Y += (clipViewbox.Height - markerHeight/scaleY) / 2
			} else if yPosition == "max" {
				clipViewbox.Y += clipViewbox.Height - markerHeight/scaleY
			}

			clipBox = &Rectangle{clipViewbox.X, clipViewbox.Y, markerWidth / scaleX, markerHeight / scaleY}
		} else {
			if box, ok := boundingBoxUnion(marker.children, dims); ok {
				scaleX = utils.MinF(markerWidth/box.Width, markerHeight/box.Height)
				scaleY = scaleX
			} else {
				scaleX, scaleY = 1, 1
			}
			clipBox = nil
		}

		// scale
		if !marker.isUnitsUserSpace {
			scale := dims.length(node.attributes.strokeWidth)
			scaleX *= scale
			scaleY *= scale
		}

		// override angle
		angle := vertex.angle
		nodeAngle := marker.orient
		if nodeAngle.U != auto && nodeAngle.U != autoStartReverse {
			angle = nodeAngle.V * math.Pi / 180 // convert from degrees to radians
		} else if nodeAngle.U == autoStartReverse && position == start {
			angle += math.Pi
		}

		// draw marker path
		for _, child := range marker.children {
			dst.OnNewStack(func() {
				mat := matrix.Rotation(angle)
				mat.LeftMultBy(matrix.Scaling(scaleX, scaleY))
				mat.LeftMultBy(matrix.Translation(vertex.x, vertex.y))
				mat.LeftMultBy(matrix.Translation(translateX, translateY))
				dst.Transform(mat)

				overflow := marker.overflow
				if clipBox != nil && (overflow == "hidden" || overflow == "scroll") {
					dst.OnNewStack(func() {
						dst.Rectangle(clipBox.X, clipBox.Y, clipBox.Width, clipBox.Height)
					})
					dst.Clip(false)
				}

				svg.drawNode(dst, child, dims, paint)
			})
		}

	}
}

// compute scale and translation needed to preserve ratio
// translate is optional
func (pr preserveAspectRatio) resolveTransforms(width, height Fl, viewbox *Rectangle, translate *point) (scaleX, scaleY, translateX, translateY Fl) {
	if viewbox == nil {
		return 1, 1, 0, 0
	}
	// }else if svg.tree == node{
	//     viewboxWidth, viewboxHeight = svg.get_intrinsic_size(font_size)
	//     if None in (viewboxWidth, viewboxHeight):
	//         return 1, 1, 0, 0
	viewboxWidth, viewboxHeight := viewbox.Width, viewbox.Height

	scaleX, scaleY = 1, 1
	if viewboxWidth != 0 {
		scaleX = width / viewboxWidth
	}
	if viewboxHeight != 0 {
		scaleY = height / viewboxHeight
	}

	if !pr.none {
		if pr.slice {
			scaleX = utils.MaxF(scaleX, scaleY)
		} else {
			scaleX = utils.MaxF(scaleX, scaleY)
		}
		scaleY = scaleX
	}

	if translate != nil {
		translateX, translateY = translate.x, translate.y
	} else {
		if pr.xPosition == "mid" {
			translateX = (width - viewboxWidth*scaleX) / 2
		} else if pr.xPosition == "max" {
			translateX = width - viewboxWidth*scaleX
		}

		if pr.yPosition == "mid" {
			translateY += (height - viewboxHeight*scaleY) / 2
		} else if pr.yPosition == "max" {
			translateY += height - viewboxHeight*scaleY
		}
	}

	if viewbox != nil {
		translateX -= viewbox.X * scaleX
		translateY -= viewbox.Y * scaleY
	}

	return
}

// resolve units and compose the transforms
func aggregateTransforms(transforms []transform, fontSize, diagonal Fl) matrix.Transform {
	// aggregate the transformations
	mat := matrix.Identity()
	for _, transform := range transforms {
		transform.applyTo(&mat, fontSize, diagonal)
	}
	return mat
}

func applyTransform(dst backend.Canvas, transforms []transform, dims drawingDims) {
	if len(transforms) == 0 { // do not apply a useless identity transform
		return
	}

	// aggregate the transformations
	mat := aggregateTransforms(transforms, dims.fontSize, dims.innerDiagonal)
	if mat.Determinant() != 0 {
		dst.Transform(mat)
	}
}

func applyFilters(dst backend.Canvas, filters []filter, node *svgNode, dims drawingDims) {
	for _, filter := range filters {
		switch filter := filter.(type) {
		case filterOffset:
			var dx, dy Fl
			if filter.isUnitsBBox {
				bbox, _ := node.resolveBoundingBox(dims, true)
				dx = filter.dx.Resolve(dims.fontSize, 1) * bbox.Width
				dy = filter.dy.Resolve(dims.fontSize, 1) * bbox.Height
			} else {
				dx, dy = dims.point(filter.dx, filter.dy)
			}
			dst.Transform(matrix.New(1, 0, 0, 1, dx, dy))
		case filterBlend:
			// TODO:
			logger.WarningLogger.Println("blend filter not implemented")
		}
	}
}

func (svg *SVGImage) applyClipPath(dst backend.Canvas, clipPath *clipPath, node *svgNode, dims drawingDims) {
	oldCtm := dst.GetTransform()

	if clipPath.isUnitsBBox {
		x, y := dims.point(node.attributes.x, node.attributes.y)
		width, height := dims.point(node.attributes.width, node.attributes.height)
		dst.Transform(matrix.New(width, 0, 0, height, x, y))
	}

	svg.drawNode(dst, &clipPath.svgNode, dims, false)

	// At least set the clipping area to an empty path, so that it’s
	// totally clipped when the clipping path is empty.
	dst.Rectangle(0, 0, 0, 0)
	dst.Clip(false)
	newCtm := dst.GetTransform()
	if err := newCtm.Invert(); err == nil {
		dst.Transform(matrix.Mul(oldCtm, newCtm))
	}
}

func (svg *SVGImage) applyMask(dst backend.Canvas, mask mask, node *svgNode, dims drawingDims) {
	widthRef, heightRef := dims.innerWidth, dims.innerHeight
	if mask.isUnitsBBox {
		widthRef, heightRef = dims.point(node.width, node.height)
	}

	x := mask.x.Resolve(dims.fontSize, widthRef)
	y := mask.y.Resolve(dims.fontSize, heightRef)
	width := mask.width.Resolve(dims.fontSize, widthRef)
	height := mask.height.Resolve(dims.fontSize, heightRef)

	mask.x = Value{x, Px}
	mask.y = Value{y, Px}
	mask.width = Value{width, Px}
	mask.height = Value{height, Px}

	if mask.isUnitsBBox {
		x, y, width, height = 0, 0, widthRef, heightRef
	} else {
		mask.viewbox = &Rectangle{X: x, Y: y, Width: width, Height: height}
	}

	alpha := dst.NewGroup(x, y, width, height)
	svg.drawNode(alpha, &mask.svgNode, dims, true)
	dst.DrawMask(alpha)
}

// ImageLoader is used to resolve and process image url found in SVG files.
type ImageLoader = func(url string) (backend.Image, error)

// Parse parsed the given SVG source data. Warnings are
// logged for unsupported elements.
// An error is returned for invalid documents.
// `baseURL` is used as base path for url resources.
// `urlFetcher` is required to handle linked SVG documents like in <use> tags.
// `imageLoader` is required to handle inner images.
func Parse(svg io.Reader, baseURL string, imageLoader ImageLoader, urlFetcher utils.UrlFetcher) (*SVGImage, error) {
	root, err := html.Parse(svg)
	if err != nil {
		return nil, err
	}

	return ParseNode(root, baseURL, imageLoader, urlFetcher)
}

// ParseNode is the same as Parse but works with an already parsed
// svg input.
func ParseNode(root *html.Node, baseURL string, imageLoader ImageLoader, urlFetcher utils.UrlFetcher) (*SVGImage, error) {
	tree, err := buildSVGTree(root, baseURL, urlFetcher)
	if err != nil {
		return nil, err
	}

	tree.imageLoader = imageLoader

	var out SVGImage
	out.definitions = newDefinitions()
	// Build the drawable items by parsing attributes
	out.root, err = tree.processNode(tree.root, out.definitions)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

type svgNode struct {
	graphicContent drawable
	children       []*svgNode
	attributes
}

// drawingDims stores the configuration to use
// when drawing
type drawingDims struct {
	// width and height as requested by the user
	// when calling Draw.
	concreteWidth, concreteHeight Fl

	fontSize Fl

	// either the root viewbox width and height,
	// or the concreteWidth, concreteHeight if
	// no viewBox is provided
	innerWidth, innerHeight Fl

	// cached value of norm(innerWidth, innerHeight) / sqrt(2)
	innerDiagonal Fl

	// cached value of norm(concreteWidth, concreteHeight) / sqrt(2)
	normalizedDiagonal Fl
}

// update `innerDiagonal` and `normalizedDiagonal`
func (dims *drawingDims) setupDiagonal() {
	dims.innerDiagonal = Fl(math.Hypot(float64(dims.innerWidth), float64(dims.innerHeight)) / math.Sqrt2)
	dims.normalizedDiagonal = Fl(math.Hypot(float64(dims.concreteWidth), float64(dims.concreteHeight)) / math.Sqrt2)
}

// resolve the size of an x/y or width/height couple.
func (dims drawingDims) point(xv, yv Value) (x, y Fl) {
	x = xv.Resolve(dims.fontSize, dims.innerWidth)
	y = yv.Resolve(dims.fontSize, dims.innerHeight)
	return
}

// resolve a length
func (dims drawingDims) length(length Value) Fl {
	return length.Resolve(dims.fontSize, dims.innerDiagonal)
}

// box is a shared type for dimensions
// found in several SVG elements, which may be expressed with units
type box struct {
	x, y, width, height Value
}

// attributes stores the SVG attributes
// shared by all node types in the final rendering tree
type attributes struct {
	viewbox *Rectangle

	transforms []transform

	clipPathID, maskID, filterID                      string
	markerID, markerStartID, markerMidID, markerEndID string

	dashArray []Value

	fill, stroke painter // fill default to black, stroke to nothing

	box

	fontSize      Value
	strokeWidth   Value // default to 1px
	strokeOptions backend.StrokeOptions

	dashOffset Value

	opacity, strokeOpacity, fillOpacity Fl // default to 1

	isFillEvenOdd    bool
	display, visible bool
}

func (tree *svgContext) processNode(node *cascadedNode, defs definitions) (*svgNode, error) {
	var children []*svgNode
	for _, c := range node.children {
		child, err := tree.processNode(c, defs)
		if err != nil {
			return nil, err
		}
		if child == nil {
			continue // do not add useless node to the tree
		}
		children = append(children, child)
	}

	// actual processing of the node, with the following cases
	//	- node used as definition, extracted from the svg tree
	//	- graphic element to display -> processGraphicNode

	id := node.attrs["id"]
	switch node.tag {
	case "filter":
		filters, err := newFilter(node)
		if err != nil {
			return nil, err
		}
		defs.filters[id] = filters
	case "clipPath":
		cp, err := newClipPath(node, children)
		if err != nil {
			return nil, err
		}
		defs.clipPaths[id] = cp
	case "mask":
		ma, err := newMask(node, children)
		if err != nil {
			return nil, err
		}
		defs.masks[id] = ma
	case "marker":
		ma, err := newMarker(node, children)
		if err != nil {
			return nil, err
		}
		defs.markers[id] = ma
	case "linearGradient", "radialGradient":
		grad, err := newGradient(node)
		if err != nil {
			return nil, err
		}
		defs.paintServers[id] = grad
	case "pattern":
		pat, err := newPattern(node, children)
		if err != nil {
			return nil, err
		}
		defs.paintServers[id] = pat
	case "use": // special case
		resolved, err := tree.resolveUse(node, defs)
		if err != nil {
			return nil, err
		}
		return resolved, nil
	case "defs":
		// children has been processed and registred,
		// so we discard the node, which is not needed anymore
	default:
		out, err := tree.processGraphicNode(node, children)
		if err != nil {
			return nil, err
		}
		// register node with id
		if id != "" {
			defs.nodes[id] = out
		}
		return out, nil
	}

	return nil, nil
}

// process a node to be displayed by building its content
func (tree *svgContext) processGraphicNode(node *cascadedNode, children []*svgNode) (*svgNode, error) {
	out := svgNode{children: children}

	err := node.attrs.parseCommonAttributes(&out.attributes)
	if err != nil {
		return nil, err
	}

	switch node.tag {
	case "circle", "ellipse":
		out.graphicContent, err = newEllipse(node, tree)
	case "image":
		out.graphicContent, err = newImage(node, tree)
	case "line":
		out.graphicContent, err = newLine(node, tree)
	case "path":
		out.graphicContent, err = newPath(node, tree)
	case "polyline":
		out.graphicContent, err = newPolyline(node, tree)
	case "polygon":
		out.graphicContent, err = newPolygon(node, tree)
	case "rect":
		out.graphicContent, err = newRect(node, tree)
	case "svg":
		out.graphicContent, err = newSvg(node, tree)
		// case "a", "text", "textPath","tspan":
		// out.graphicContent, err = newText(node, tree)
	}

	if err != nil {
		return nil, fmt.Errorf("invalid element %s: %s", node.tag, err)
	}

	return &out, nil
}

func (na nodeAttributes) parseBox(out *box) (err error) {
	out.x, err = parseValue(na["x"])
	if err != nil {
		return err
	}
	out.y, err = parseValue(na["y"])
	if err != nil {
		return err
	}
	out.width, err = parseValue(na["width"])
	if err != nil {
		return err
	}
	out.height, err = parseValue(na["height"])
	if err != nil {
		return err
	}

	return nil
}

func (na nodeAttributes) parseCommonAttributes(out *attributes) error {
	err := na.parseBox(&out.box)
	if err != nil {
		return err
	}
	out.viewbox, err = na.viewBox()
	if err != nil {
		return err
	}
	out.fontSize, err = na.fontSize()
	if err != nil {
		return err
	}
	out.strokeWidth, err = na.strokeWidth()
	if err != nil {
		return err
	}

	out.opacity, err = parseOpacity(na["opacity"])
	if err != nil {
		return err
	}
	out.strokeOpacity, err = parseOpacity(na["stroke-opacity"])
	if err != nil {
		return err
	}
	out.fillOpacity, err = parseOpacity(na["fill-opacity"])
	if err != nil {
		return err
	}

	out.transforms, err = parseTransform(na["transform"])
	if err != nil {
		return err
	}

	out.stroke, err = newPainter(na["stroke"])
	if err != nil {
		return err
	}
	out.fill, err = na.fill()
	if err != nil {
		return err
	}
	out.isFillEvenOdd = na["fill-rull"] == "evenodd"

	out.dashOffset, err = parseValue(na["stroke-dashoffset"])
	if err != nil {
		return err
	}
	out.dashArray, err = parseValues(na["stroke-dasharray"])
	if err != nil {
		return err
	}

	out.strokeOptions.LineCap = na.lineCap()
	out.strokeOptions.LineJoin = na.lineJoin()
	out.strokeOptions.MiterLimit, err = na.miterLimit()
	if err != nil {
		return err
	}

	out.filterID = parseURLFragment(na["filter"])
	out.clipPathID = parseURLFragment(na["clip-path"])
	out.maskID = parseURLFragment(na["mask"])

	out.markerID = parseURLFragment(na["marker"])
	out.markerStartID = parseURLFragment(na["marker-start"])
	out.markerMidID = parseURLFragment(na["marker-mid"])
	out.markerEndID = parseURLFragment(na["marker-end"])

	out.display = na.display()
	out.visible = na.visible()
	return nil
}
