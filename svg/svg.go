// Package svg implements parsing of SVG images.
// It transforms SVG text files into an in-memory structure
// that is easy to draw.
// CSS is supported via the style and cascadia packages.
package svg

import (
	"fmt"
	"io"
	"log"
	"math"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/utils"
)

// convert from an svg tree to the final form

// nodes that are not directly draw but may be referenced by ID
// from other nodes
type definitions struct {
	filters      map[string][]filter
	clipPaths    map[string]clipPath
	masks        map[string]mask
	paintServers map[string]paintServer
	markers      map[string]*marker
}

func newDefinitions() definitions {
	return definitions{
		filters:      make(map[string][]filter),
		clipPaths:    make(map[string]clipPath),
		masks:        make(map[string]mask),
		paintServers: make(map[string]paintServer),
		markers:      make(map[string]*marker),
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
	var ctx drawingDims
	ctx.concreteWidth, ctx.concreteHeight = width, height
	if vb := svg.ViewBox(); vb != nil {
		ctx.innerWidth, ctx.innerHeight = vb.Width, vb.Height
	} else {
		ctx.innerWidth, ctx.innerHeight = width, height
	}
	ctx.fontSize = defaultFontSize
	ctx.setupDiagonal()
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
			dst = dst.AddOpacityGroup(x, y, width, height)
		}

		applyTransform(dst, node.attributes.transforms, dims)

		// clip
		if cp, has := svg.definitions.clipPaths[node.clipPathID]; has {
			applyClipPath(dst, cp, node, dims)
		}

		// manage display and visibility
		display := node.attributes.display
		visible := node.attributes.visible

		// draw the node itself.
		var vertices []vertex
		if visible && node.graphicContent != nil {
			vertices = node.graphicContent.draw(dst, &node.attributes, dims)
		}

		// then recurse
		if display {
			for _, child := range node.children {
				svg.drawNode(dst, child, dims, paint)
			}
		}

		// apply mask
		if ma, has := svg.definitions.masks[node.maskID]; has {
			applyMask(dst, ma, node, dims)
		}

		// do the actual painting
		if paint {
			svg.paintNode(dst, node, dims)
		}

		// draw markers
		if vertices != nil {
			// svg.drawMarkers(dst, vertices, node, dims, paint)
		}

		// apply opacity group and restore original target
		if paint && 0 <= opacity && opacity < 1 {
			originalDst.DrawOpacityGroup(opacity, dst)
			dst = originalDst
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
			clipBox                                *Rectangle
			scaleX, scaleY, translateX, translateY Fl
		)
		markerWidth, markerHeight := dims.point(marker.markerWidth, marker.markerHeight)
		if vb := node.attributes.viewbox; vb != nil {
			scaleX, scaleY, translateX, translateY = marker.preserveAspectRatio.resolveTransforms(dims, markerWidth, markerHeight, nil)

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
			translateX, translateY = dims.point(marker.refX, marker.refY)
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
func (pr preserveAspectRatio) resolveTransforms(dims drawingDims, width, height Fl, viewbox *Rectangle) (scaleX, scaleY, translateX, translateY Fl) {
	// FIXME:
	// viewbox = viewbox or node.get_viewbox()
	// if viewbox:
	//     viewbox_width, viewbox_height = viewbox[2:]
	// elif svg.tree == node:
	//     viewbox_width, viewbox_height = svg.get_intrinsic_size(font_size)
	//     if None in (viewbox_width, viewbox_height):
	//         return 1, 1, 0, 0
	// else:
	//     return 1, 1, 0, 0

	// scale_x = width / viewbox_width if viewbox_width else 1
	// scale_y = height / viewbox_height if viewbox_height else 1

	// aspect_ratio = node.get('preserveAspectRatio', 'xMidYMid').split()
	// align = aspect_ratio[0]
	// if align == 'none':
	//     x_position = 'min'
	//     y_position = 'min'
	// else:
	//     meet_or_slice = aspect_ratio[1] if len(aspect_ratio) > 1 else None
	//     if meet_or_slice == 'slice':
	//         scale_value = max(scale_x, scale_y)
	//     else:
	//         scale_value = min(scale_x, scale_y)
	//     scale_x = scale_y = scale_value
	//     x_position = align[1:4].lower()
	//     y_position = align[5:].lower()

	// if node.tag == 'marker':
	//     translate_x, translate_y = svg.point(
	//         node.get('refX'), node.get('refY', '0'), font_size)
	// else:
	//     translate_x = 0
	//     if x_position == 'mid':
	//         translate_x = (width - viewbox_width * scale_x) / 2
	//     elif x_position == 'max':
	//         translate_x = width - viewbox_width * scale_x

	//     translate_y = 0
	//     if y_position == 'mid':
	//         translate_y += (height - viewbox_height * scale_y) / 2
	//     elif y_position == 'max':
	//         translate_y += height - viewbox_height * scale_y

	// if viewbox:
	//     translate_x -= viewbox[0] * scale_x
	//     translate_y -= viewbox[1] * scale_y

	return
}

func applyTransform(dst backend.CanvasNoFill, transforms []transform, dims drawingDims) {
	if len(transforms) == 0 { // do not apply a useless identity transform
		return
	}

	// aggregate the transformations
	mat := matrix.Identity()
	for _, transform := range transforms {
		transform.applyTo(&mat, dims)
	}
	if mat.Determinant() != 0 {
		dst.Transform(mat)
	}
}

func applyFilters(dst backend.CanvasNoFill, filters []filter, node *svgNode, dims drawingDims) {
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
			log.Println("blend filter not implemented")
		}
	}
}

func applyClipPath(dst backend.CanvasNoFill, clipPath clipPath, node *svgNode, dims drawingDims) {
	// old_ctm = self.stream.ctm
	if clipPath.isUnitsBBox {
		x, y := dims.point(node.attributes.x, node.attributes.y)
		width, height := dims.point(node.attributes.width, node.attributes.height)
		dst.Transform(matrix.New(width, 0, 0, height, x, y))
	}

	// FIXME:
	log.Println("applying clip path is not supported")
	// clip_path._etree_node.tag = 'g'
	// self.draw_node(clip_path, font_size, fill_stroke=False)

	// At least set the clipping area to an empty path, so that it’s
	// totally clipped when the clipping path is empty.
	dst.Rectangle(0, 0, 0, 0)
	dst.Clip(false)
	// new_ctm = self.stream.ctm
	// if new_ctm.determinant:
	//     self.stream.transform(*(old_ctm @ new_ctm.invert).values)
}

func applyMask(dst backend.CanvasNoFill, mask mask, node *svgNode, dims drawingDims) {
	// mask._etree_node.tag = 'g'

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
		// TODO: update viewbox if needed
		//     mask.attrib['viewBox'] = f'{x} {y} {width} {height}'
	}

	// FIXME:
	log.Println("mask not implemented")
	// alpha_stream = svg.stream.add_group([x, y, width, height])
	// state = pydyf.Dictionary({
	//     'Type': '/ExtGState',
	//     'SMask': pydyf.Dictionary({
	//         'Type': '/Mask',
	//         'S': '/Luminance',
	//         'G': alpha_stream,
	//     }),
	//     'ca': 1,
	//     'AIS': 'false',
	// })
	// svg.stream.set_state(state)

	// svg_stream = svg.stream
	// svg.stream = alpha_stream
	// svg.draw_node(mask, font_size)
	// svg.stream = svg_stream
}

// ImageLoader is used to resolve and process image url found in SVG files.
type ImageLoader = func(url string) (backend.Image, error)

// Parse parsed the given SVG data. Warnings are
// logged for unsupported elements.
// An error is returned for invalid documents.
// `baseURL` is used as base path for url resources.
// `imageLoader` is required to handle inner images.
func Parse(svg io.Reader, baseURL string, imageLoader ImageLoader) (*SVGImage, error) {
	tree, err := buildSVGTree(svg, baseURL)
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
}

// update `innerDiagonal` from `innerWidth` and `innerHeight`.
func (dims *drawingDims) setupDiagonal() {
	dims.innerDiagonal = Fl(math.Hypot(float64(dims.innerWidth), float64(dims.innerHeight)) / math.Sqrt2)
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

	stroke, fill painter

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
	children := make([]*svgNode, len(node.children))
	for i, c := range node.children {
		var err error
		children[i], err = tree.processNode(c, defs)
		if err != nil {
			return nil, err
		}
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
		defs.clipPaths[id] = newClipPath(node, children)
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
		pat, err := newPattern(node)
		if err != nil {
			return nil, err
		}
		defs.paintServers[id] = pat
	case "defs":
		// children has been processed and registred,
		// so we discard the node, which is not needed anymore
	default:
		return tree.processGraphicNode(node, children)
	}

	return nil, nil
}

// process a node to be displayed by building its content
func (tree *svgContext) processGraphicNode(node *cascadedNode, children []*svgNode) (*svgNode, error) {
	out := svgNode{children: children}
	builder := elementBuilders[node.tag]
	if builder == nil {
		// this node is not drawn, return an empty node
		// with its children
		return &out, nil
	}

	err := node.attrs.parseCommonAttributes(&out.attributes)
	if err != nil {
		return nil, err
	}

	out.graphicContent, err = builder(node, tree)
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
	out.fill, err = newPainter(na["fill"])
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
