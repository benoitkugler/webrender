package svg

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/utils"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// convert from html nodes to an intermediate svg tree

// svgContext is an intermediated representation of an SVG file,
// where CSS has been applied, and text has been processed
type svgContext struct {
	root *cascadedNode // with tag svg

	baseURL     string
	imageLoader ImageLoader
	urlFetcher  utils.UrlFetcher

	// to handle use tags
	defs map[string]*cascadedNode
	// ID of the current <use> target being resolved,
	// to prevent infinite recursion
	inUseIDs utils.Set

	// cache
	pathParser pathParser
}

// cascadedNode is a node in an SVG document.
// we use this intermediate representation to
// ease the cascading of the properties
// and the text handling
type cascadedNode struct {
	tag      string
	text     []byte
	attrs    nodeAttributes
	children []*cascadedNode
}

// returns a copy
// attrs id deepcopied, but the children and text are shallow copies
func (c *cascadedNode) copy() cascadedNode {
	out := *c
	out.attrs = make(nodeAttributes, len(c.attrs))
	for k, v := range c.attrs {
		out.attrs[k] = v
	}
	return out
}

// raw attributes value of a node
// attibutes will be updated in the post processing
// step due to the cascade
type nodeAttributes map[string]string

func newNodeAttributes(attrs []html.Attribute) nodeAttributes {
	out := make(nodeAttributes, len(attrs))
	for _, attr := range attrs {
		out[attr.Key] = attr.Val
	}
	return out
}

func (na nodeAttributes) viewBox() (*Rectangle, error) {
	if attrValue := na["viewBox"]; attrValue != "" {
		v, err := parseViewbox(attrValue)
		return &v, err
	}
	return nil, nil
}

func (na nodeAttributes) fontSize() (Value, error) {
	attrValue, has := na["font-size"]
	if !has {
		attrValue = "1em"
	}
	return parseValue(attrValue)
}

func (na nodeAttributes) strokeWidth() (Value, error) {
	attrValue, has := na["stroke-width"]
	if !has {
		attrValue = "1px"
	}
	return parseValue(attrValue)
}

func (na nodeAttributes) aspectRatio() preserveAspectRatio {
	preserveAspectRatio := "xMidYMid"
	if s, has := na["preserveAspectRatio"]; has {
		preserveAspectRatio = s
	}
	return parsePreserveAspectRatio(preserveAspectRatio)
}

// default to black
func (na nodeAttributes) fill() (painter, error) {
	attrValue, has := na["fill"]
	if !has {
		attrValue = "black"
	}
	return newPainter(attrValue)
}

func (na nodeAttributes) lineCap() backend.StrokeCapMode {
	switch na["stroke-linecap"] {
	case "round":
		return backend.RoundCap
	case "square":
		return backend.SquareCap
	default:
		return backend.ButtCap
	}
}

func (na nodeAttributes) lineJoin() backend.StrokeJoinMode {
	switch na["stroke-linejoin"] {
	case "round":
		return backend.Round
	case "bevel":
		return backend.Bevel
	default:
		return backend.Miter
	}
}

func (na nodeAttributes) miterLimit() (Fl, error) {
	attrValue, has := na["stroke-miterlimit"]
	if !has {
		attrValue = "4"
	}
	v, err := strconv.ParseFloat(attrValue, 32)
	if v < 0 {
		v = 4
	}
	return Fl(v), err
}

func (na nodeAttributes) markerWidth() (Value, error) {
	attrValue := na["markerWidth"]
	return parseValue(attrValue)
}

func (na nodeAttributes) markerHeight() (Value, error) {
	attrValue := na["markerHeight"]
	return parseValue(attrValue)
}

func (na nodeAttributes) markerUnitsUserSpace() bool {
	attrValue := na["markerUnits"]
	return attrValue == "userSpaceOnUse"
}

func (na nodeAttributes) display() bool {
	attrValue := na["display"]
	return attrValue != "none"
}

func (na nodeAttributes) visible() bool {
	attrValue := na["visibility"]
	visible := attrValue != "hidden"
	return na.display() && visible
}

func (na nodeAttributes) strokeDasharray() ([]Value, error) {
	attrValue := na["stroke-dasharray"]
	return parseValues(attrValue)
}

func (na nodeAttributes) strokeDashoffset() (Value, error) {
	attrValue := na["stroke-dashoffset"]
	return parseValue(attrValue)
}

func (na nodeAttributes) spacePreserve() bool {
	return na["space"] == "preserve"
}

// walk the tree to extract content needed to build the SVG tree
func fetchStyleAndTextRefs(root *utils.HTMLNode) ([][]byte, map[string][]byte) {
	var (
		stylesheets [][]byte
		trefs       = make(map[string][]byte)
	)
	iter := root.Iter()
	for iter.HasNext() {
		node := iter.Next()
		if css := handleStyleElement(node); len(css) != 0 {
			stylesheets = append(stylesheets, css)
			continue
		}

		// register text refs
		if id := node.Get("id"); id != "" {
			trefs[id] = node.GetChildrenText()
		}
	}
	return stylesheets, trefs
}

// Convert from the html representation to an internal,
// simplified form, suitable for post-processing.
// The stylesheets are processed and applied, the values
// of the CSS properties begin stored as attributes
// Inheritable attributes are cascaded and 'inherit' special values are resolved.
func buildSVGTreeReader(svg io.Reader, baseURL string, urlFetcher utils.UrlFetcher) (*svgContext, error) {
	root, err := html.Parse(svg)
	if err != nil {
		return nil, err
	}

	return buildSVGTree(root, baseURL, urlFetcher)
}

func buildSVGTree(root *html.Node, baseURL string, urlFetcher utils.UrlFetcher) (*svgContext, error) {
	// extract the root svg node, which is not
	// always the first one
	iter := utils.NewHtmlIterator(root, atom.Svg)
	if !iter.HasNext() {
		return nil, errors.New("missing <svg> element")
	}
	svgRoot := iter.Next()

	stylesheets, trefs := fetchStyleAndTextRefs(svgRoot)
	normalMatcher, importantMatcher := parseStylesheets(stylesheets, baseURL)

	// build the SVG tree and apply style attribute
	var out svgContext
	out.baseURL = baseURL
	out.urlFetcher = urlFetcher
	out.defs = make(map[string]*cascadedNode)
	out.inUseIDs = make(utils.Set)

	// may return nil to discard the node
	var buildTree func(node *html.Node, parentAttrs nodeAttributes) *cascadedNode

	buildTree = func(node *html.Node, parentAttrs nodeAttributes) *cascadedNode {
		// text is handled by the parent
		// style elements are no longer useful
		if node.Type != html.ElementNode || node.DataAtom == atom.Style {
			return nil
		}

		attrs := newNodeAttributes(node.Attr)
		// Cascade attributes
		for key, value := range parentAttrs {
			if _, isNotInherited := notInheritedAttributes[key]; !isNotInherited {
				if _, isSet := attrs[key]; !isSet {
					attrs[key] = value
				}
			}
		}

		// Apply style
		attrs.applyStyle(baseURL, (*html.Node)(node), normalMatcher, importantMatcher)

		// Replace 'currentColor' value
		for key := range colorAttributes {
			if attrs[key] == "currentColor" {
				if c, has := attrs["color"]; has {
					attrs[key] = c
				} else {
					attrs[key] = "black"
				}
			}
		}

		// Handle 'inherit' values
		for key, value := range attrs {
			if value == "inherit" {
				attrs[key] = parentAttrs[key]
			}
		}

		nodeSVG := &cascadedNode{
			tag:   node.Data,
			text:  (*utils.HTMLNode)(node).GetChildrenText(),
			attrs: attrs,
		}

		// recurse
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if childSVG := buildTree(child, attrs); childSVG != nil {
				nodeSVG.children = append(nodeSVG.children, childSVG)
			}
		}

		// Fix text in text tags
		if node.Data == "text" || node.Data == "textPath" || node.Data == "a" {
			handleText(nodeSVG, true, true, trefs)
		}

		if ID := attrs["id"]; ID != "" {
			out.defs[ID] = nodeSVG
		}

		return nodeSVG
	}

	out.root = buildTree((*html.Node)(svgRoot), nil)

	return &out, nil
}

var (
	replacerPreserve   = strings.NewReplacer("\n", " ", "\r", " ", "\t", " ")
	replacerNoPreserve = strings.NewReplacer("\n", "", "\r", "", "\t", " ")
)

// replace newlines by spaces, and merge spaces if not preserved.
func processWhitespace(text []byte, preserveSpace bool) []byte {
	if preserveSpace {
		return []byte(replacerPreserve.Replace(string(text)))
	}
	return []byte(replacerNoPreserve.Replace(string(text)))
}

// handle text node by fixing whitespaces and flattening tails,
// updating node 'children' and 'text'
func handleText(node *cascadedNode, trailingSpace, textRoot bool, trefs map[string][]byte) bool {
	preserve := node.attrs.spacePreserve()
	node.text = processWhitespace(node.text, preserve)
	if trailingSpace && !preserve {
		node.text = bytes.TrimLeft(node.text, " ")
	}

	if len(node.text) != 0 {
		trailingSpace = bytes.HasSuffix(node.text, []byte{' '})
	}

	var newChildren []*cascadedNode
	for _, child := range node.children {
		if child.tag == "tref" {
			// Retrieve the referenced node and get its flattened text
			// and remove the node children.
			id := parseURLFragment(child.attrs["href"])
			node.text = append(node.text, trefs[id]...)
			continue
		}

		trailingSpace = handleText(child, trailingSpace, false, trefs)

		newChildren = append(newChildren, child)
	}

	if textRoot && len(newChildren) == 0 && !preserve {
		node.text = bytes.TrimRight(node.text, " ")
	}

	node.children = newChildren

	return trailingSpace
}

// these attributes are not cascaded
var notInheritedAttributes = utils.NewSet(
	"clip",
	"clip-path",
	"filter",
	"height",
	"id",
	"mask",
	"opacity",
	"overflow",
	"rotate",
	"stop-color",
	"stop-opacity",
	"style",
	"transform",
	"transform-origin",
	"viewBox",
	"width",
	"x",
	"y",
	"dx",
	"dy",
	"href",
)

var colorAttributes = utils.NewSet(
	"fill",
	"flood-color",
	"lighting-color",
	"stop-color",
	"stroke",
)
