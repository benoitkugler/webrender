package utils

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// http://whatwg.org/C#space-character
const htmlWhitespace = " \t\n\f\r"

var (
	htmlSpacesRe               = regexp.MustCompile(fmt.Sprintf("[%s]+", htmlWhitespace))
	htmlSpaceSeparatedTokensRe = regexp.MustCompile(fmt.Sprintf("[^%s]+", htmlWhitespace))
)

type PageElement struct {
	Side  string
	Name  string
	Index int
	Blank bool
	First bool
}

type ElementKey struct {
	Element    *HTMLNode
	PseudoType string
	PageType   PageElement
}

func (e ElementKey) IsPageType() bool {
	return e.Element == nil
}

func (p PageElement) ToKey(pseudoType string) ElementKey {
	return ElementKey{PseudoType: pseudoType, PageType: p}
}

type HTMLNode html.Node

func (h *HTMLNode) AsHtmlNode() *html.Node {
	return (*html.Node)(h)
}

func (h *HTMLNode) ToKey(pseudoType string) ElementKey {
	return ElementKey{PseudoType: pseudoType, Element: h}
}

// Get returns the attribute `name` or ""
// See HasAttr if you need to distinguish between no attribute
// and an attribute with an empty string value.
func (h HTMLNode) Get(name string) string {
	for _, attr := range h.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

// HasAttr returns true if `name` is among the attributes (possibly empty).
func (h HTMLNode) HasAttr(name string) bool {
	for _, attr := range h.Attr {
		if attr.Key == name {
			return true
		}
	}
	return false
}

// Iter return an iterator over the html tree.
// If tags are given, only the node matching them
// will be returned by the iterator.
func (h *HTMLNode) Iter(tags ...atom.Atom) HTMLIterator {
	return NewHtmlIterator((*html.Node)(h), tags...)
}

// ------------------------------------ html walk utilities ------------------------------------

// HTMLIterator simplify the (depth first) walk on an HTML tree.
type HTMLIterator struct {
	tagsMap map[atom.Atom]bool // if nil, means all
	result  *html.Node         // valid element to return
	toVisit []*html.Node
}

// NewHtmlIterator use `root` as start point.
// If `tags` is given, only node matching one of them are returned.
func NewHtmlIterator(root *html.Node, tags ...atom.Atom) HTMLIterator {
	tagsMap := make(map[atom.Atom]bool)
	for _, tag := range tags {
		tagsMap[tag] = true
	}
	return HTMLIterator{toVisit: []*html.Node{root}, tagsMap: tagsMap}
}

// pop the stack
func (h *HTMLIterator) popNode() *html.Node {
	L := len(h.toVisit) - 1
	next := h.toVisit[L]
	h.toVisit = h.toVisit[:L]
	return next
}

// HasNext returns true if a node still has to be visited.
func (h *HTMLIterator) HasNext() bool {
	if h.result != nil { // empty the stack
		return true
	}

	if len(h.toVisit) == 0 { // walk is done
		return false
	}

	next := h.popNode()
	if next.NextSibling != nil {
		h.toVisit = append(h.toVisit, next.NextSibling)
	}
	if next.FirstChild != nil {
		h.toVisit = append(h.toVisit, next.FirstChild)
	}

	if len(h.tagsMap) == 0 || h.tagsMap[next.DataAtom] { // found one element
		h.result = next
		return true
	}

	// check the remaining nodes
	return h.HasNext()
}

func (h *HTMLIterator) Next() *HTMLNode {
	out := (*HTMLNode)(h.result)
	h.result = nil
	return out
}

// NodeChildren returns the direct children of `element`.
// Skip empty text nodes
func (element HTMLNode) NodeChildren(skipBlank bool) (children []*HTMLNode) {
	child := element.FirstChild
	for child != nil {
		if !(skipBlank && child.Type == html.TextNode && strings.TrimSpace(child.Data) == "") {
			children = append(children, (*HTMLNode)(child))
		}
		child = child.NextSibling
	}
	return
}

// IsText returns true if the node is a non empty text node.
func (element HTMLNode) IsText() (bool, string) {
	if text := element.Data; element.Type == html.TextNode && text != "" {
		return true, text
	}
	return false, ""
}

// GetChildrenText returns the text directly in the element, but not descendants.
// It's the concatenation of all children's TextNodes.
func (element HTMLNode) GetChildrenText() (content []byte) {
	if element.Type == html.TextNode {
		content = []byte(element.Data)
	}

	for _, child := range element.NodeChildren(false) {
		if child.Type == html.TextNode {
			content = append(content, child.Data...)
		}
	}
	return content
}

// GetText returns the content of the first text node child.
// Due to Go html.Parse() behavior, this method mimic Python xml.etree.text
// attribute.
func (element HTMLNode) GetText() string {
	if c := element.FirstChild; c != nil && c.Type == html.TextNode {
		return c.Data
	}
	return ""
}

// Transform (only) ASCII letters to lower case: A-Z is mapped to a-z.
//     This is used for `ASCII case-insensitive
//     <http://whatwg.org/C#ascii-case-insensitive>`_ matching.
//     This is different from the strings.ToLower function
//     which also affect non-ASCII characters,
//     sometimes mapping them into the ASCII range:
//     		keyword = u"Bac\u212Aground"
//     		assert strings.ToLower(keyword) == u"background"
//     		assert asciiLower(keyword) != strings.ToLower(keyword)
//     		assert asciiLower(keyword) == u"bac\u212Aground"
//
func AsciiLower(s string) string {
	rs := []rune(s)
	out := make([]rune, len(rs))
	for index, c := range rs {
		if c < unicode.MaxASCII {
			c = unicode.ToLower(c)
		}
		out[index] = c
	}
	return string(out)
}

// Return whether the given element has a `rel` attribute with the
// given link type.
// `linkType` must be a lower-case string.
func (element HTMLNode) HasLinkType(linkType string) bool {
	attr := element.Get("rel")
	matchs := htmlSpaceSeparatedTokensRe.FindAllString(attr, -1)
	for _, token := range matchs {
		if AsciiLower(token) == linkType {
			return true
		}
	}
	return false
}

// Return the base URL for the document.
// See http://www.w3.org/TR/html5/urls.html#document-base-url
func FindBaseUrl(htmlDocument *html.Node, fallbackBaseUrl string) string {
	iter := NewHtmlIterator(htmlDocument, atom.Base)
	firstBaseElement := iter.Next()
	if firstBaseElement != nil {
		href := strings.TrimSpace(firstBaseElement.Get("href"))
		if href != "" {
			out, err := SafeUrljoin(fallbackBaseUrl, href, true)
			if err != nil {
				return ""
			}
			return out
		}
	}
	return fallbackBaseUrl
}

// Return whether the given element has a `rel` attribute with the
// given link type (must be a lower-case string).
func ElementHasLinkType(element *HTMLNode, linkType string) bool {
	for _, token := range htmlSpaceSeparatedTokensRe.FindAllString(element.Get("rel"), -1) {
		if AsciiLower(token) == linkType {
			return true
		}
	}
	return false
}

// Meta-information belonging to a whole `Document`.
type DocumentMetadata struct {
	// The title of the document, as a string.
	// Extracted from the `<title>` element in HTML
	// and written to the `/Title` info field in PDF.
	Title string

	// The description of the document, as a string.
	// Extracted from the `<meta name=description>` element in HTML
	// and written to the `/Subject` info field in PDF.
	Description string

	// The name of one of the software packages
	// used to generate the document, as a string.
	// Extracted from the `<meta name=generator>` element in HTML
	// and written to the `/Creator` info field in PDF.
	Generator string

	// Keywords associated with the document, as a list of strings.
	// (Defaults to the empty list.)
	// Extracted from `<meta name=keywords>` elements in HTML
	// and written to the `/Keywords` info field in PDF.
	Keywords []string

	// The authors of the document, as a list of strings.
	// (Defaults to the empty list.)
	// Extracted from the `<meta name=author>` elements in HTML
	// and written to the `/Author` info field in PDF.
	Authors []string

	// The creation date of the document, as a string.
	// Dates are in one of the six formats specified in
	// `W3C’s profile of ISO 8601 <http://www.w3.org/TR/NOTE-datetime>`.
	// Extracted from the `<meta name=dcterms.created>` element in HTML
	// and written to the `/CreationDate` info field in PDF.
	Created time.Time

	// The modification date of the document, as a string.
	// Dates are in one of the six formats specified in
	// `W3C’s profile of ISO 8601 <http://www.w3.org/TR/NOTE-datetime>`.
	// Extracted from the `<meta name=dcterms.modified>` element in HTML
	// and written to the `/ModDate` info field in PDF.
	Modified time.Time

	// File attachments, as a list of tuples of URL and a description.
	// (Defaults to the empty list.)
	// Extracted from the `<link rel=attachment>` elements in HTML
	// and written to the `/EmbeddedFiles` dictionary in PDF.
	Attachments []Attachment
}

type Attachment struct {
	URL, Title string
}

// Relevant specs:
//     http://www.whatwg.org/html#the-title-element
//     http://www.whatwg.org/html#standard-metadata-names
//     http://wiki.whatwg.org/wiki/MetaExtensions
//     http://microformats.org/wiki/existing-rel-values#HTML5LinkExtensionsT
//
func GetHtmlMetadata(wrapperElement *HTMLNode, baseUrl string) DocumentMetadata {
	var (
		title, description, generator string
		authors, keywords             = []string{}, []string{}
		created, modified             time.Time
		keywordsSet                   = map[string]bool{}
		attachments                   = []Attachment{}
	)
	iter := wrapperElement.Iter(atom.Title, atom.Meta, atom.Link)
	for iter.HasNext() {
		element := iter.Next()
		switch element.DataAtom {
		case atom.Title:
			if title == "" {
				title = string(element.GetChildrenText())
			}
		case atom.Meta:
			name := AsciiLower(element.Get("name"))
			content := element.Get("content")
			switch name {
			case "keywords":
				for _, keyword := range strings.Split(content, ",") {
					keyword = stripWhitespace(keyword)
					if !keywordsSet[keyword] {
						keywords = append(keywords, keyword)
						keywordsSet[keyword] = true
					}
				}
			case "author":
				authors = append(authors, content)
			case "description":
				if description == "" {
					description = content
				}
			case "generator":
				if generator == "" {
					generator = content
				}
			case "dcterms.created":
				if created.IsZero() {
					created = parseW3cDate(name, content)
				}
			case "dcterms.modified":
				if modified.IsZero() {
					modified = parseW3cDate(name, content)
				}
			}
		case atom.Link:
			if ElementHasLinkType(element, "attachment") {
				url := element.GetUrlAttribute("href", baseUrl, false)
				attTitle := element.Get("title")
				if url == "" {
					log.Println("Missing href in <link rel='attachment'>")
				} else {
					attachments = append(attachments, Attachment{URL: url, Title: attTitle})
				}
			}
		}
	}
	return DocumentMetadata{
		Title:       title,
		Description: description,
		Generator:   generator,
		Keywords:    keywords,
		Authors:     authors,
		Created:     created,
		Modified:    modified,
		Attachments: attachments,
	}
}

// Use the HTML definition of "space character",
//     not all Unicode Whitespace.
//     http://www.whatwg.org/html#strip-leading-and-trailing-whitespace
//     http://www.whatwg.org/html#space-character
//
func stripWhitespace(s string) string {
	return strings.Trim(s, htmlWhitespace)
}

// YYYY (eg 1997)
// YYYY-MM (eg 1997-07)
// YYYY-MM-DD (eg 1997-07-16)
// YYYY-MM-DDThh:mmTZD (eg 1997-07-16T19:20+01:00)
// YYYY-MM-DDThh:mm:ssTZD (eg 1997-07-16T19:20:30+01:00)
// YYYY-MM-DDThh:mm:ss.sTZD (eg 1997-07-16T19:20:30.45+01:00)
var (
	w3CDateRe = regexp.MustCompile(
		`^` +
			"[ \t\n\f\r]*" +
			`(?P<year>\d\d\d\d)` +
			`(?:` +
			`-(?P<month>0\d|1[012])` +
			`(?:` +
			`-(?P<day>[012]\d|3[01])` +
			`(?:` +
			`T(?P<hour>[01]\d|2[0-3])` +
			`:(?P<minute>[0-5]\d)` +
			`(?:` +
			`:(?P<second>[0-5]\d)` +
			`(?:\.\d+)?` + // Second fraction, ignored
			`)?` +
			`(?:Z|(?P<tzHour>[+-](?:[01]\d|2[0-3])):(?P<tzMinute>[0-5]\d))` + // UTC
			`)?` +
			`)?` +
			`)?` +
			"[ \t\n\f\r]*" +
			`$`)

	W3CDateReGroupsIndexes = map[string]int{}
)

func init() {
	for i, name := range w3CDateRe.SubexpNames() {
		if i != 0 && name != "" {
			W3CDateReGroupsIndexes[name] = i
		}
	}
}

func toInt(s string, defaut ...int) int {
	if s == "" && len(defaut) > 0 {
		return defaut[0]
	}
	out, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("unexpected string for int : %s", s)
	}
	return out
}

// http://www.w3.org/TR/NOTE-datetime
func parseW3cDate(metaName, str string) time.Time {
	match := w3CDateRe.FindStringSubmatch(str)
	if len(match) == 0 {
		log.Printf("Invalid %s date: %s", metaName, str)
		return time.Time{}
	}
	year := toInt(match[W3CDateReGroupsIndexes["year"]])
	month := toInt(match[W3CDateReGroupsIndexes["month"]], 1)
	day := toInt(match[W3CDateReGroupsIndexes["day"]], 1)
	hour := toInt(match[W3CDateReGroupsIndexes["hour"]], 0)
	minute := toInt(match[W3CDateReGroupsIndexes["minute"]], 0)
	second := toInt(match[W3CDateReGroupsIndexes["second"]], 0)
	var tzHour, tzMinute int
	if match[W3CDateReGroupsIndexes["hour"]] != "" {
		if match[W3CDateReGroupsIndexes["minute"]] == "" {
			log.Fatalf("minute shouldn't be empty when hour is present")
		}
		if match[W3CDateReGroupsIndexes["tzHour"]] != "" {
			if !(strings.HasPrefix(match[W3CDateReGroupsIndexes["tzHour"]], "+") || strings.HasPrefix(match[W3CDateReGroupsIndexes["tzHour"]], "-")) {
				log.Fatalf("tzHour should start by + or -, got %s", match[W3CDateReGroupsIndexes["tzHour"]])
			}
			if match[W3CDateReGroupsIndexes["tzMinute"]] == "" {
				log.Fatalf("tzMinute shouldn't be empty when tzHour is present")
			}
			tzHour = toInt(match[W3CDateReGroupsIndexes["tzHour"]])
			tzMinute = toInt(match[W3CDateReGroupsIndexes["tzMinute"]])
		}
	}
	loc := time.UTC
	if tzHour != 0 || tzMinute != 0 {
		loc = time.FixedZone("w3c", tzHour*3600+tzMinute*60)
	}
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, loc)
}
