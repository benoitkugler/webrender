package svg

import (
	"bytes"
	"errors"
	"strconv"
	"strings"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
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
func buildSVGTree(root *html.Node, baseURL string) (*svgContext, error) {
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
		var normalAttr, importantAttr []declaration
		if styleAttr := attrs["style"]; styleAttr != "" {
			normalAttr, importantAttr = parseDeclarations(parser.Tokenize([]byte(styleAttr), false))
		}
		delete(attrs, "style") // not useful anymore

		var allProps []declaration
		allProps = append(allProps, normalMatcher.match((*html.Node)(node))...)
		allProps = append(allProps, normalAttr...)
		allProps = append(allProps, importantMatcher.match((*html.Node)(node))...)
		allProps = append(allProps, importantAttr...)
		for _, d := range allProps {
			attrs[d.property] = strings.TrimSpace(d.value)
		}

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

		return nodeSVG
	}

	out.root = buildTree((*html.Node)(svgRoot), nil)
	out.baseURL = baseURL

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
