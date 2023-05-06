package boxes

import (
	"strings"

	"github.com/benoitkugler/webrender/images"
	"github.com/benoitkugler/webrender/logger"

	"github.com/benoitkugler/webrender/utils"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type handlerFunction = func(element *utils.HTMLNode, box Box, resolver URLResolver, baseUrl string) []Box

// htmlHandlers map a tag name to a callback creating the boxes needed.
var htmlHandlers = map[string]handlerFunction{
	"img":      handleImg,
	"embed":    handleEmbed,
	"object":   handleObject,
	"colgroup": handleColgroup,
	"col":      handleCol,
	"svg":      handleSVG,
}

// HandleElement handle HTML elements that need special care.
func handleElement(element *utils.HTMLNode, box Box, resolver URLResolver, baseUrl string) []Box {
	handler, in := htmlHandlers[box.Box().ElementTag()]
	if in {
		ls := handler(element, box, resolver, baseUrl)
		return ls
	}
	return []Box{box}
}

// Wrap an image in a replaced box.
//
// That box is either block-level or inline-level, depending on what the
// element should be.
func makeReplacedBox(element *utils.HTMLNode, box Box, image images.Image) Box {
	var newBox Box
	if box.Box().Style.GetDisplay().Has("block") {
		b := NewBlockReplacedBox(box.Box().Style, (*html.Node)(element), "", image)
		newBox = &b
	} else {
		b := NewInlineReplacedBox(box.Box().Style, (*html.Node)(element), "", image)
		newBox = &b
	}
	newBox.Box().StringSet = box.Box().StringSet
	newBox.Box().BookmarkLabel = box.Box().BookmarkLabel
	return newBox
}

// Handle “<img>“ elements, return either an image or the alt-text.
// See: http://www.w3.org/TR/html5/embedded-content-1.html#the-img-element
func handleImg(element *utils.HTMLNode, box Box, resolver URLResolver, baseUrl string) []Box {
	src := element.GetUrlAttribute("src", baseUrl, false)
	alt := element.Get("alt")
	if src != "" {
		image := resolver.FetchImage(src, "", box.Box().Style.GetImageOrientation())
		if image != nil {
			return []Box{makeReplacedBox(element, box, image)}
		}
		// Invalid image, use the alt-text.
		if alt != "" {
			box.Box().Children = []Box{TextBoxAnonymousFrom(box, alt)}
			return []Box{box}
		}
	} else {
		if alt != "" {
			box.Box().Children = []Box{TextBoxAnonymousFrom(box, alt)}
			return []Box{box}
		}
	}
	// The element represents nothing
	return nil
}

// Handle “<embed>“ elements, return either an image or nothing.
// See: https://www.w3.org/TR/html5/embedded-content-0.html#the-embed-element
func handleEmbed(element *utils.HTMLNode, box Box, resolver URLResolver, baseUrl string) []Box {
	src := element.GetUrlAttribute("src", baseUrl, false)
	type_ := strings.TrimSpace(element.Get("type"))
	if src != "" {
		image := resolver.FetchImage(src, type_, box.Box().Style.GetImageOrientation())
		if image != nil {
			return []Box{makeReplacedBox(element, box, image)}
		}
	}
	// No fallback.
	return nil
}

// Handle “<object>“ elements, return either an image or the fallback
// content.
// See: https://www.w3.org/TR/html5/embedded-content-0.html#the-object-element
func handleObject(element *utils.HTMLNode, box Box, resolver URLResolver, baseUrl string) []Box {
	data := element.GetUrlAttribute("data", baseUrl, false)
	type_ := strings.TrimSpace(element.Get("type"))
	if data != "" {
		image := resolver.FetchImage(data, type_, box.Box().Style.GetImageOrientation())
		if image != nil {
			return []Box{makeReplacedBox(element, box, image)}
		}
	}
	// The element’s children are the fallback.
	return []Box{box}
}

// Handle the “span“ attribute.
func handleColgroup(element *utils.HTMLNode, box Box, _ URLResolver, _ string) []Box {
	if box, ok := box.(*TableColumnGroupBox); ok { // leaf
		hasCol := false
		for _, child := range element.NodeChildren(true) {
			if child.DataAtom == atom.Col {
				hasCol = true
			}
		}
		if !hasCol {
			children := make([]Box, box.span())
			for i := range children {
				children[i] = TableColumnBoxAnonymousFrom(box, nil)
			}
			box.Box().Children = children
		}
	}
	return []Box{box}
}

// Handle the “span“ attribute.
func handleCol(_ *utils.HTMLNode, box Box, _ URLResolver, _ string) []Box {
	if box, ok := box.(*TableColumnBox); ok { // leaf
		if span := box.span(); span > 1 {
			// Generate multiple boxes
			// http://lists.w3.org/Archives/Public/www-style/2011Nov/0293.html
			out := make([]Box, span)
			for i := range out {
				out[i] = box.Copy()
			}
			return out
		}
	}
	return []Box{box}
}

// handle the inline <svg> elements
// Return either an image or the fallback content.
func handleSVG(element *utils.HTMLNode, box Box, resolver URLResolver, baseUrl string) []Box {
	img, err := images.NewSVGImageFromNode((*html.Node)(element), baseUrl, resolver.Fetch)
	if err != nil {
		logger.WarningLogger.Printf("Failed to load inline SVG: %s", err)
		return nil
	}
	return []Box{makeReplacedBox(element, box, img)}
}
