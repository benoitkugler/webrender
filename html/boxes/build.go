package boxes

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/benoitkugler/webrender/images"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/text"

	"github.com/benoitkugler/webrender/css/parser"

	"github.com/benoitkugler/webrender/css/counters"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/utils"
	"golang.org/x/net/html"
)

// Turn an element tree with associated CSS style (computed values)
// into a "before layout" formatting structure / box tree.
// This includes creating anonymous boxes and processing whitespace
// as necessary.

var (
	TableFirstLetter = []*unicode.RangeTable{unicode.Ps, unicode.Pe, unicode.Pi, unicode.Pf, unicode.Po}

	styleScores = map[pr.String]float64{}
	styleMap    = map[pr.String]pr.String{
		"inset":  "ridge",
		"outset": "groove",
	}

	transparent = pr.Color(parser.ParseColorString("transparent"))
)

func init() {
	styles := []pr.String{
		"hidden", "double", "solid", "dashed", "dotted", "ridge",
		"outset", "groove", "inset", "none",
	}
	N := len(styles) - 1
	for i, v := range styles {
		styleScores[v] = float64(N - i)
	}
}

type Context interface {
	// RunningElements() map[string]map[int]Box
	CurrentPage() int
	GetStringSetFor(page Box, name, keyword string) string
	GetRunningElementFor(page Box, name, keyword string) Box
}

type URLResolver struct {
	Fetch      utils.UrlFetcher
	FetchImage ImageFetcher
}

type ImageFetcher = func(url, forcedMimeType string, orientation pr.SBoolFloat) images.Image

type styleForI interface {
	Get(element tree.Element, pseudoType string) pr.ElementStyle
}

type rootStyleFor struct {
	elementTree *utils.HTMLNode
	tree.StyleFor
}

func (r rootStyleFor) Get(element tree.Element, pseudoType string) pr.ElementStyle {
	style := r.StyleFor.Get(element, pseudoType)
	if style != nil {
		if element == r.elementTree {
			style.SetDisplay(pr.Display{"block", "flow"})
		} else {
			style.SetDisplay(pr.Display{"none"})
		}
	}
	return style
}

var (
	lineFeedRe = regexp.MustCompile(`\r\n?`)
	tabRe      = regexp.MustCompile(`[\t ]*\n[\t ]*`)
	spaceRe    = regexp.MustCompile(`[\t ]+`)
)

// CreateAnonymousBox create anonymous boxes in box descendants according to layout rules.
func CreateAnonymousBox(box Box) Box {
	box = AnonymousTableBoxes(box)
	box = FlexBoxes(box)
	box = GridBoxes(box)
	box = InlineInBlock(box)
	box = BlockInInline(box)
	return box
}

// Build a formatting structure (box tree) from an element tree.
func BuildFormattingStructure(elementTree *utils.HTMLNode, styleFor *tree.StyleFor, resolver URLResolver,
	baseUrl string, targetCollector *tree.TargetCollector, cs counters.CounterStyle, footnotes *[]Box,
) BlockLevelBoxITF {
	boxList := elementToBox(elementTree, styleFor, resolver, baseUrl, targetCollector, cs, nil, footnotes)

	var box Box
	if len(boxList) > 0 {
		box = boxList[0]
	} else { //  No root element
		rsf := rootStyleFor{elementTree: elementTree, StyleFor: *styleFor}
		box = elementToBox(elementTree, rsf, resolver, baseUrl, targetCollector, cs, nil, footnotes)[0]
	}

	targetCollector.CheckPendingTargets()

	box.Box().IsForRootElement = true
	// If this is changed, maybe update layout.pages.makeMarginBoxes()
	box = CreateAnonymousBox(box)
	box = setViewportOverflow(box)
	return box.(BlockLevelBoxITF)
}

// Maps values of the “display“ CSS property to box types.
func makeBox(style pr.ElementStyle, content []Box, element *utils.HTMLNode, pseudoType string) (b Box, _ error) {
	tmp := style.GetDisplay()
	display := [2]string{tmp[0], tmp[1]}
	switch display {
	case [2]string{"block", "flow"}:
		b = NewBlockBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"inline", "flow"}:
		b = NewInlineBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"block", "flow-root"}:
		b = NewBlockBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"inline", "flow-root"}:
		b = NewInlineBlockBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"block", "table"}:
		b = NewTableBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"inline", "table"}:
		b = NewInlineTableBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"block", "flex"}:
		b = NewFlexBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"inline", "flex"}:
		b = NewInlineFlexBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"block", "grid"}:
		b = NewGridBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"inline", "grid"}:
		b = NewInlineGridBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"table-row"}:
		b = NewTableRowBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"table-row-group"}:
		b = NewTableRowGroupBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"table-header-group"}:
		b = NewTableRowGroupBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"table-footer-group"}:
		b = NewTableRowGroupBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"table-column"}:
		b = NewTableColumnBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"table-column-group"}:
		b = NewTableColumnGroupBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"table-cell"}:
		b = NewTableCellBox(style, (*html.Node)(element), pseudoType, content)
	case [2]string{"table-caption"}:
		b = NewTableCaptionBox(style, (*html.Node)(element), pseudoType, content)
	default:
		return nil, fmt.Errorf("ignored box %s: display property %s not supported", element.Data, tmp)
	}
	return b, nil
}

// Convert an element and its children into a box with children.
//
// Return a list of boxes. Most of the time the list will have one item but
// may have zero or more than one.
//
//	Eg.::
//
//	    <p>Some <em>emphasised</em> text.</p>
//
//	gives (not actual syntax)::
//
//	    BlockBox[
//	        TextBox["Some "],
//	        InlineBox[
//	            TextBox["emphasised"],
//	        ],
//	        TextBox[" text."],
//	    ]
//
//	``TextBox``es are anonymous inline boxes:
//	See https://www.w3.org/TR/CSS21/visuren.html#anonymous
func elementToBox(element *utils.HTMLNode, styleFor styleForI,
	resolver URLResolver, baseUrl string, targetCollector *tree.TargetCollector,
	cs counters.CounterStyle, state *tree.PageState, footnotes *[]Box,
) []Box {
	if element.Type != html.TextNode && element.Type != html.ElementNode && element.Type != html.DocumentNode {
		// Here we ignore comments and XML processing instructions.
		return nil
	}

	style := styleFor.Get(element, "")

	display := style.GetDisplay()
	if display == (pr.Display{"none"}) {
		return nil
	}

	if style.GetFloat() == "footnote" {
		if style.GetFootnoteDisplay() == "block" {
			style.SetDisplay(pr.Display{"block", "flow"})
		} else {
			style.SetDisplay(pr.Display{"inline", "flow"})
		}
	}

	box, err := makeBox(style, nil, element, "")
	if err != nil {
		logger.WarningLogger.Println(err)
		return nil
	}

	if state == nil {
		// use a list to have a shared mutable object
		state = &tree.PageState{
			// Shared mutable objects:
			QuoteDepth: []int{0}, // single integer
			CounterValues: tree.CounterValues{
				"footnote": []int{0},
			}, // name -> stacked/scoped values
			CounterScopes: []utils.Set{ //  element tree depths -> counter names
				utils.NewSet("footnote"),
			},
		}
	}

	counterValues := state.CounterValues

	UpdateCounters(state, style)
	// If this element’s direct children create new scopes, the counter
	// names will be in this new list
	state.CounterScopes = append(state.CounterScopes, utils.Set{})

	box.Box().FirstLetterStyle = styleFor.Get(element, "first-letter")
	box.Box().firstLineStyle = styleFor.Get(element, "first-line")

	var children, markerBoxes []Box
	if display.Has("list-item") {
		mb := markerToBox(element, state, style, styleFor, resolver, targetCollector, cs)
		if mb != nil {
			markerBoxes = []Box{mb}
		}
		children = append(children, markerBoxes...)
	}

	vs := beforeAfterToBox(element, "before", state, styleFor, resolver, targetCollector, cs)
	children = append(children, vs...)

	// collect anchor's counter_values, maybe it's a target.
	// to get the spec-conform counter_values we must do it here,
	// after the ::before is parsed and before the ::after is
	if anchor := string(style.GetAnchor()); anchor != "" {
		targetCollector.StoreTarget(anchor, counterValues, box)
	}

	if text := element.Data; element.Type == html.TextNode && text != "" {
		children = append(children, TextBoxAnonymousFrom(box, text))
	}

	for _, childElement := range element.NodeChildren(false) {
		// utils.HTMLNode as no notion of tail. Instead, text are converted in text nodes
		if ok, text := childElement.IsText(); ok && text != "" {
			textBox := TextBoxAnonymousFrom(box, text)
			if len(children) > 0 {
				// TextBox is a leaf in inheritance tree, so we can type assert against the concrete type
				// instead of using interfaces.
				if ct, ok := children[len(children)-1].(*TextBox); ok {
					ct.Text = append(ct.Text, textBox.Text...)
				} else {
					children = append(children, textBox)
				}
			} else {
				children = append(children, textBox)
			}
		} else {
			childBoxes := elementToBox(childElement, styleFor, resolver, baseUrl, targetCollector, cs, state, footnotes)
			if len(childBoxes) != 0 && childBoxes[0].Box().Style.GetFloat() == "footnote" {
				footnote := childBoxes[0]
				footnote.Box().Style.SetFloat("none")
				(*footnotes) = append(*footnotes, footnote)
				callStyle := styleFor.Get(element, "footnote-call")
				footnoteCall, err := makeBox(callStyle, nil, element, "footnote-call")
				if err != nil {
					logger.WarningLogger.Println(err)
					return nil
				}
				footnoteCall.Box().Children = ContentToBoxes(
					callStyle, footnoteCall, state.QuoteDepth, state.CounterValues,
					resolver, targetCollector, cs, nil, nil)
				footnoteCall.Box().Footnote = footnote
				childBoxes = []Box{footnoteCall}
			}
			children = append(children, childBoxes...)
		}
	}
	children = append(children, beforeAfterToBox(element, "after", state, styleFor, resolver, targetCollector, cs)...)

	// Scopes created by this element’s children stop here.
	counterScope := state.CounterScopes[len(state.CounterScopes)-1]
	state.CounterScopes = state.CounterScopes[:len(state.CounterScopes)-1]
	for name := range counterScope {
		counterValues[name] = counterValues[name][:len(counterValues[name])-1]
		if len(counterValues[name]) == 0 {
			delete(counterValues, name)
		}
	}

	box.Box().Children = children
	if style.GetAppearance() != "none" {
		box.Box().Children = nil
	}

	ProcessWhitespace(box, false)
	setContentLists(element, box, style, counterValues, targetCollector, cs)
	ProcessTextTransform(box)

	if len(markerBoxes) > 0 && len(box.Box().Children) == 1 {
		// See https://www.w3.org/TR/css-lists-3/#list-style-position-outside
		//
		// "The size or contents of the marker box may affect the height of the
		//  principal block box and/or the height of its first line box, and in
		//  some cases may cause the creation of a new line box; this
		//  interaction is also not defined."
		//
		// We decide here to add a zero-width space to have a minimum
		// height. Adding text boxes is not the best idea, but it's not a good
		// moment to add an empty line box, and the specification lets us do
		// almost what we want, so…
		if style.GetListStylePosition() == "outside" {
			box.Box().Children = append(box.Box().Children, TextBoxAnonymousFrom(box, "\u200b"))
		}
	}

	if style.GetFloat() == "footnote" {
		state.CounterValues["footnote"][len(state.CounterValues["footnote"])-1] += 1
		markerStyle := styleFor.Get(element, "footnote-marker")
		marker, err := makeBox(markerStyle, nil, element, "footnote-marker")
		if err != nil {
			logger.WarningLogger.Println(err)
			return nil
		}
		marker.Box().Children = ContentToBoxes(
			markerStyle, box, state.QuoteDepth, state.CounterValues, resolver,
			targetCollector, cs, nil, nil)
		box.Box().Children = append([]Box{marker}, box.Box().Children...)
	}
	// Specific handling for the element. (eg. replaced element)
	return handleElement(element, box, resolver, baseUrl)
}

// Yield the box for ::before or ::after pseudo-element.
func beforeAfterToBox(element *utils.HTMLNode, pseudoType string, state *tree.PageState, styleFor styleForI,
	resolver URLResolver, targetCollector *tree.TargetCollector, cs counters.CounterStyle,
) []Box {
	style := styleFor.Get(element, pseudoType)
	if pseudoType != "" && style == nil {
		// Pseudo-elements with no style at all do not get a style dict.
		// Their initial content property computes to "none".
		return nil
	}

	display := style.GetDisplay()
	content := style.GetContent()
	if display == (pr.Display{"none"}) || content.String == "none" || content.String == "normal" || content.String == "inhibit" {
		return nil
	}

	box, err := makeBox(style, nil, element, pseudoType)
	if err != nil {
		logger.WarningLogger.Println(err)
		return nil
	}

	UpdateCounters(state, style)

	var children []Box
	if display.Has("list-item") {
		mb := markerToBox(element, state, style, styleFor, resolver, targetCollector, cs)
		if mb != nil {
			children = append(children, mb)
		}
	}
	children = append(children, ContentToBoxes(
		style, box, state.QuoteDepth, state.CounterValues, resolver, targetCollector, cs, nil, nil)...)

	box.Box().Children = children

	// calculate the bookmark-label
	if style.GetBookmarkLevel().Tag != pr.None {
		computeBookmarkLabel(element, box, style.GetBookmarkLabel(), state.CounterValues, targetCollector, cs)
	}

	return []Box{box}
}

// Yield the box for ::marker pseudo-element if there is one.
// https://drafts.csswg.org/css-lists-3/#marker-pseudo
func markerToBox(element *utils.HTMLNode, state *tree.PageState, parentStyle pr.ElementStyle, styleFor styleForI,
	resolver URLResolver, targetCollector *tree.TargetCollector, cs counters.CounterStyle,
) Box {
	style := styleFor.Get(element, "marker")

	box, err := makeBox(style, nil, element, "marker")
	if err != nil {
		logger.WarningLogger.Println(err)
		return nil
	}

	children := &box.Box().Children

	if style.GetDisplay() == (pr.Display{"none"}) {
		return nil
	}

	image := style.GetListStyleImage()

	if content := style.GetContent().String; content != "normal" && content != "inhibit" {
		*children = append(*children, ContentToBoxes(style, box, state.QuoteDepth, state.CounterValues,
			resolver, targetCollector, cs, nil, nil)...)
	} else {
		if imageUrl, ok := image.(pr.UrlImage); ok {
			// image may be None here too, in case the image is not available.
			image_ := resolver.FetchImage(string(imageUrl), "", style.GetImageOrientation())
			if image_ != nil {
				markerBox := InlineReplacedBoxAnonymousFrom(box, image_)
				*children = append(*children, markerBox)
			}
		}
		if len(*children) == 0 && style.GetListStyleType().Name != "none" {
			counterValue_, has := state.CounterValues["list-item"]
			if !has {
				counterValue_ = []int{0}
			}
			counterValue := counterValue_[len(counterValue_)-1]
			markerText := cs.RenderMarker(style.GetListStyleType(), counterValue)
			markerBox := TextBoxAnonymousFrom(box, markerText)
			markerBox.Box().Style.SetWhiteSpace("pre-wrap")
			*children = append(*children, markerBox)
		}
	}

	if len(*children) == 0 {
		return nil
	}
	var markerBox Box
	if parentStyle.GetListStylePosition() == "outside" {
		markerBox = BlockBoxAnonymousFrom(box, *children)
		// We can safely edit everything that can't be changed by user style
		// See https://drafts.csswg.org/css-pseudo-4/#marker-pseudo
		markerBox.Box().Style.SetPosition(pr.BoolString{String: "absolute"})
		translateX := pr.Dimension{Value: 100, Unit: pr.Perc}
		if parentStyle.GetDirection() == "ltr" {
			translateX = pr.Dimension{Value: -100, Unit: pr.Perc}
		}
		translateY := pr.ZeroPixels
		markerBox.Box().Style.SetTransform(pr.Transforms{{String: "translate", Dimensions: pr.Dimensions{translateX, translateY}}})
	} else {
		markerBox = InlineBoxAnonymousFrom(box, *children)
	}
	return markerBox
}

// Collect missing counters.
func collectMissingCounter(counterName string, counterValues tree.CounterValues, missingCounters utils.Set) {
	for s := range counterValues {
		if s == counterName {
			return
		}
	}
	missingCounters.Add(counterName)
}

// Collect missing target counters.
//
// The corresponding TargetLookupItem caches the target"s page based
// counter values during pagination.
func collectMissingTargetCounter(counterName string, lookupCounterValues tree.CounterValues,
	anchorName string, missingTargetCounters map[string]utils.Set,
) {
	if _, in := lookupCounterValues[counterName]; !in {
		missingCounters := missingTargetCounters[anchorName]
		if missingCounters == nil {
			missingCounters = make(utils.Set)
		}
		for s := range missingCounters {
			if counterName == s {
				return
			}
		}
		missingCounters.Add(counterName)
		missingTargetCounters[anchorName] = missingCounters
	}
}

// Compute and return the boxes corresponding to the “content_list“.
//
// “parseAgain“ is called to compute the “content_list“ again when
// “target_collector.lookup_target()“ detected a pending target.
//
// “build_formatting_structure“ calls
// “target_collector.check_pending_targets()“ after the first pass to do
// required reparsing.
func computeContentList(contentList pr.ContentProperties, parentBox Box, counterValues tree.CounterValues,
	cssToken string, parseAgain tree.ParseFunc, targetCollector *tree.TargetCollector, cs counters.CounterStyle,
	resolver URLResolver, quoteDepth []int, quoteStyle pr.Quotes, lang pr.TaggedString, context Context, page Box,
) []Box {
	contentBoxes := []Box{}

	missingCounters := utils.Set{}
	missingTargetCounters := map[string]utils.Set{}
	inPageContext := context != nil && page != nil

	// Collect missing counters during build_formatting_structure.
	// Pointless to collect missing target counters in MarginBoxes.
	needCollectMissing := targetCollector.IsCollecting() && !inPageContext

	if parentBox.Box().cachedCounterValues == nil {
		// Store the counter_values in the parent_box to make them accessible
		// in @page context. Obsoletes the parse_again function's deepcopy.
		parentBox.Box().cachedCounterValues = counterValues.Copy()
	}

	var hasText bool
	addText := func(text string) {
		hasText = true
		if text == "" {
			return
		}
		if L := len(contentBoxes); L != 0 {
			if textBox, ok := contentBoxes[L-1].(*TextBox); ok {
				textBox.Text = append(textBox.Text, []rune(text)...)
				return
			}
		}
		contentBoxes = append(contentBoxes, TextBoxAnonymousFrom(parentBox, text))
	}

outerLoop:
	for _, content := range contentList {
		switch content.Type {
		case "string":
			addText(content.AsString())
		case "url":
			if resolver.FetchImage == nil {
				continue
			}
			value := content.Content.(pr.NamedString)
			if value.Name != "external" {
				// Embedding internal references is impossible
				continue
			}
			image := resolver.FetchImage(value.String, "", parentBox.Box().Style.GetImageOrientation())
			if image != nil {
				contentBoxes = append(contentBoxes, InlineReplacedBoxAnonymousFrom(parentBox, image))
			}
		case "content()":
			addedText := extractText(content.AsString(), parentBox)
			// Simulate the step of white space processing
			// (normally done during the layout)
			addedText = strings.TrimSpace(addedText)
			addText(addedText)
		case "string()":
			value := content.AsStrings()
			if !inPageContext {
				// string() is currently only valid in @page context
				// See https://github.com/Kozea/WeasyPrint/issues/723
				logger.WarningLogger.Printf("'string(%s)' is only allowed in page margins", strings.Join(value, " "))
				continue
			}
			if len(value) == 1 {
				value = append(value, "first")
			}
			addText(context.GetStringSetFor(page, value[0], value[1]))
		case "counter()":
			counterName, counterStyle := content.AsCounter()
			if needCollectMissing {
				collectMissingCounter(counterName, counterValues, missingCounters)
			}
			if counterStyle.Name == "none" {
				continue
			}
			cv, has := counterValues[counterName]
			if !has {
				cv = []int{0}
			}
			counterValue := cv[len(cv)-1]
			addText(cs.RenderValueStyle(counterValue, counterStyle))
		case "counters()":
			counterName, separator, counterStyle := content.AsCounters()
			if needCollectMissing {
				collectMissingCounter(counterName, counterValues, missingCounters)
			}
			if counterStyle.Name == "none" {
				continue
			}
			vs, has := counterValues[counterName]
			if !has {
				vs = []int{0}
			}
			styles := make([]string, len(vs))
			for i, counterValue := range vs {
				styles[i] = cs.RenderValueStyle(counterValue, counterStyle)
			}
			addText(strings.Join(styles, separator))
		case "target-counter()":
			anchorToken, counterName, counterStyle := content.AsTargetCounter()
			lookupTarget := targetCollector.LookupTarget(anchorToken, parentBox, cssToken, parseAgain)
			if lookupTarget.IsUpToDate() {
				targetValues := lookupTarget.TargetBox.CachedCounterValues()
				if needCollectMissing {
					collectMissingTargetCounter(counterName, targetValues,
						tree.AnchorNameFromToken(anchorToken),
						missingTargetCounters)
				}
				// Mixin target"s cached page counters.
				// cachedPageCounterValues are empty during layout.
				localCounters := lookupTarget.CachedPageCounterValues.Copy()
				localCounters.Update(targetValues)
				vs, has := localCounters[counterName]
				if !has {
					vs = []int{0}
				}
				counterValue := vs[len(vs)-1]
				addText(cs.RenderValue(counterValue, counterStyle))
			} else {
				break outerLoop
			}
		case "target-counters()":
			anchorToken, counterName, separator, counterStyle := content.AsTargetCounters()
			lookupTarget := targetCollector.LookupTarget(
				anchorToken, parentBox, cssToken, parseAgain)
			if lookupTarget.IsUpToDate() {
				if separator.Type != "string" {
					break outerLoop
				}
				separatorString := separator.AsString()
				targetValues := lookupTarget.TargetBox.CachedCounterValues()
				if needCollectMissing {
					collectMissingTargetCounter(
						counterName, targetValues,
						tree.AnchorNameFromToken(anchorToken),
						missingTargetCounters)
				}
				// Mixin target"s cached page counters.
				// cachedPageCounterValues are empty during layout.
				localCounters := lookupTarget.CachedPageCounterValues.Copy()
				localCounters.Update(targetValues)
				vs, has := localCounters[counterName]
				if !has {
					vs = []int{0}
				}
				tmps := make([]string, len(vs))
				for j, counterValue := range vs {
					tmps[j] = cs.RenderValue(counterValue, counterStyle)
				}
				addText(strings.Join(tmps, separatorString))
			} else {
				break outerLoop
			}
		case "target-text()":
			anchorToken, textStyle := content.AsTargetText()
			lookupTarget := targetCollector.LookupTarget(
				anchorToken, parentBox, cssToken, parseAgain)
			if lookupTarget.IsUpToDate() {
				targetBox := lookupTarget.TargetBox
				text := extractText(textStyle, targetBox.(Box))
				// Simulate the step of white space processing
				// (normally done during the layout)
				addText(strings.TrimSpace(text))
			} else {
				break outerLoop
			}
		case "quote":
			if quoteDepth != nil && !quoteStyle.IsNone() {
				value := content.AsQuote()
				isOpen := value.Open
				insert := value.Insert && quoteStyle.Tag != pr.None
				if !isOpen {
					quoteDepth[0] = utils.MaxInt(0, quoteDepth[0]-1)
				}
				if insert {
					openQuotes, closeQuotes := quoteStyle.Open, quoteStyle.Close
					if quoteStyle.Tag == pr.Auto {
						openQuotes, closeQuotes = text.GetLangQuotes(lang.S)
					}
					quotes := closeQuotes
					if isOpen {
						quotes = openQuotes
					}
					addText(quotes[utils.MinInt(quoteDepth[0], len(quotes)-1)])
				}
				if isOpen {
					quoteDepth[0] += 1
				}
			}
		case "element()":
			value := content.AsStrings()
			if !inPageContext {
				logger.WarningLogger.Printf("element(%s) is only allowed in page margins", strings.Join(value, " "))
				continue
			}
			if len(value) == 1 {
				value = append(value, "first")
			}
			newBox := context.GetRunningElementFor(page, value[0], value[1])
			if newBox == nil {
				continue
			}

			newBox = Deepcopy(newBox)

			newBox.Box().Style.SetPosition(pr.BoolString{String: "static"})

			if ParentT.IsInstance(newBox) {
				for _, child := range Descendants(newBox) {
					if content := child.Box().Style.GetContent(); content.String == "normal" || content.String == "none" {
						continue
					}
					child.Box().Children = ContentToBoxes(
						child.Box().Style, child, quoteDepth, counterValues,
						resolver, targetCollector, cs, context, page)
				}
			}
			contentBoxes = append(contentBoxes, newBox)
		case "leader()":
			ld := content.AsLeader()
			if ld == "" {
				continue
			}

			textBox := TextBoxAnonymousFrom(parentBox, ld)
			leaderBox := InlineBoxAnonymousFrom(parentBox, []Box{textBox})
			// Avoid breaks inside the leader box
			leaderBox.Style.SetWhiteSpace("pre")
			// Prevent whitespaces from being removed from the text box
			textBox.Style.SetWhiteSpace("pre")
			leaderBox.IsLeader = true
			contentBoxes = append(contentBoxes, leaderBox)
		}
	}

	if hasText || len(contentBoxes) > 0 {
		// Only add CounterLookupItem if the content_list actually produced text
		targetCollector.CollectMissingCounters(parentBox, cssToken, parseAgain, missingCounters, missingTargetCounters)
		return contentBoxes
	}
	return nil
}

// Takes the value of a “content“ property and yield boxes.
func ContentToBoxes(style pr.ElementStyle, parentBox Box, quoteDepth []int, counterValues tree.CounterValues,
	resolver URLResolver, targetCollector *tree.TargetCollector, cs counters.CounterStyle, context Context, page Box,
) []Box {
	origQuoteDepth := make([]int, len(quoteDepth))

	// Closure to parse the ``parentBoxes`` children all again.
	parseAgain := func(mixinPagebasedCounters tree.CounterValues) {
		// Neither alters the mixed-in nor the cached counter values, no
		// need to deepcopy here
		localCounters := mixinPagebasedCounters.Copy()
		for k, v := range parentBox.Box().cachedCounterValues {
			localCounters[k] = v
		}

		var localChildren []Box
		localChildren = append(localChildren, ContentToBoxes(
			style, parentBox, origQuoteDepth, localCounters,
			resolver, targetCollector, cs, nil, nil)...)

		parentChildren := parentBox.Box().Children
		if len(parentChildren) == 1 && LineT.IsInstance(parentChildren[0]) {
			parentChildren[0].Box().Children = localChildren
		} else {
			parentBox.Box().Children = localChildren
		}
	}

	if style.GetContent().String == "inhibit" {
		return nil
	}

	copy(origQuoteDepth, quoteDepth)

	cssToken := "content"
	boxList := computeContentList(
		style.GetContent().Contents, parentBox, counterValues, cssToken, parseAgain,
		targetCollector, cs, resolver, quoteDepth, style.GetQuotes(), style.GetLang(),
		context, page)
	return boxList
}

// Parse the content-list value of “stringName“ for “string-set“.
func computeStringSet(element *utils.HTMLNode, box Box, stringName string, contentList pr.ContentProperties,
	counterValues tree.CounterValues, targetCollector *tree.TargetCollector, cs counters.CounterStyle,
) {
	// Closure to parse the string-set string value all again.
	parseAgain := func(mixinPagebasedCounters tree.CounterValues) {
		// Neither alters the mixed-in nor the cached counter values, no
		// need to deepcopy here
		localCounters := mixinPagebasedCounters.Copy()
		for k, v := range box.Box().cachedCounterValues {
			localCounters[k] = v
		}

		computeStringSet(element, box, stringName, contentList, localCounters, targetCollector, cs)
	}

	cssToken := "string-set::" + stringName
	boxList := computeContentList(contentList, box, counterValues, cssToken, parseAgain,
		targetCollector, cs, URLResolver{}, nil, pr.Quotes{}, pr.TaggedString{}, nil, nil)
	if boxList != nil {
		var builder []rune
		for _, box1 := range boxList {
			if textBox, ok := box1.(*TextBox); ok {
				builder = append(builder, textBox.Text...)
			}
		}
		string_ := string(builder)
		// Avoid duplicates, care for parseAgain and missing counters, don't
		// change the pointer
		newStringSet := make(pr.ContentProperties, 0, len(box.Box().StringSet))
		for i, stringSet := range box.Box().StringSet {
			if stringSet.Type == stringName {
				newStringSet = append(newStringSet, box.Box().StringSet[i+1:]...)
				break
			}
			newStringSet = append(newStringSet, stringSet)
		}
		newStringSet = append(newStringSet, pr.ContentProperty{Type: stringName, Content: pr.String(string_)})
		box.Box().StringSet = newStringSet
	}
}

// Parses the content-list value for “bookmark-label“.
func computeBookmarkLabel(element *utils.HTMLNode, box Box, contentList pr.ContentProperties, counterValues tree.CounterValues,
	targetCollector *tree.TargetCollector, cs counters.CounterStyle,
) {
	// Closure to parse the bookmark-label all again..
	parseAgain := func(mixinPagebasedCounters tree.CounterValues) {
		// Neither alters the mixed-in nor the cached counter values, no
		// need to deepcopy here
		localCounters := mixinPagebasedCounters.Copy()
		for k, v := range box.Box().cachedCounterValues {
			localCounters[k] = v
		}
		computeBookmarkLabel(element, box, contentList, localCounters, targetCollector, cs)
	}

	cssToken := "bookmark-label"
	boxList := computeContentList(contentList, box, counterValues, cssToken, parseAgain, targetCollector, cs,
		URLResolver{}, nil, pr.Quotes{}, pr.TaggedString{}, nil, nil)

	var builder []rune
	for _, box := range boxList {
		if textBox, ok := box.(*TextBox); ok {
			builder = append(builder, textBox.Text...)
		}
	}
	box.Box().BookmarkLabel = string(builder)
}

// Set the content-lists values.
// These content-lists are used in GCPM properties like “string-set“ and
// “bookmark-label“.
func setContentLists(element *utils.HTMLNode, box Box, style pr.ElementStyle, counterValues tree.CounterValues,
	targetCollector *tree.TargetCollector, cs counters.CounterStyle,
) {
	if sss := style.GetStringSet(); sss.String != "none" {
		for _, c := range sss.Contents {
			stringName, stringValues := c.String, c.Contents
			computeStringSet(element, box, stringName, stringValues, counterValues, targetCollector, cs)
		}
	}
	if style.GetBookmarkLevel().Tag != pr.None {
		computeBookmarkLabel(element, box, style.GetBookmarkLabel(), counterValues, targetCollector, cs)
	}
}

// Handle the “counter-*“ properties.
func UpdateCounters(state *tree.PageState, style pr.ElementStyle) {
	_, counterValues, counterScopes := state.QuoteDepth, state.CounterValues, state.CounterScopes
	siblingScopes := counterScopes[len(counterScopes)-1]

	for _, nv := range style.GetCounterReset().Values {
		slice := counterValues[nv.String]
		if siblingScopes.Has(nv.String) {
			slice = slice[:len(slice)-1]
		} else {
			siblingScopes.Add(nv.String)
		}
		counterValues[nv.String] = append(slice, nv.Int)
	}

	for _, nv := range style.GetCounterSet().Values {
		values := counterValues[nv.String]
		if len(values) == 0 {
			if siblingScopes.Has(nv.String) {
				logger.WarningLogger.Println("ci.String shoud'nt be in siblingScopes")
			}
			siblingScopes.Add(nv.String)
			values = append(values, 0)
		}
		values[len(values)-1] = nv.Int
		counterValues[nv.String] = values
	}

	counterIncrement := style.GetCounterIncrement()
	if counterIncrement.String == "auto" {
		// "auto" is the initial value but is not valid in stylesheet:
		// there was no counter-increment declaration for this element.
		// (Or the winning value was "initial".)
		// https://drafts.csswg.org/css-lists-3/#declaring-a-list-item
		if style.GetDisplay().Has("list-item") {
			counterIncrement = pr.SIntStrings{Values: pr.IntStrings{{String: "list-item", Int: 1}}}
		} else {
			counterIncrement = pr.SIntStrings{}
		}
	}
	for _, ci := range counterIncrement.Values {
		values := counterValues[ci.String]
		if len(values) == 0 {
			if siblingScopes.Has(ci.String) {
				logger.WarningLogger.Println("ci.String shoud'nt be in siblingScopes")
			}
			siblingScopes.Add(ci.String)
			values = append(values, 0)
		}
		values[len(values)-1] += ci.Int
		counterValues[ci.String] = values
	}
}

var reHasNonWhitespace = regexp.MustCompile(`\S`)

func hasNonWhitespaceDefault(text string) bool { return reHasNonWhitespace.MatchString(text) }

// Return true if “box“ is a TextBox with only whitespace.
func isWhitespace(box Box, hasNonWhitespace func(string) bool) bool {
	if hasNonWhitespace == nil {
		hasNonWhitespace = hasNonWhitespaceDefault
	}
	textBox, is := box.(*TextBox)
	return is && !hasNonWhitespace(textBox.TextS())
}

type wrapImproperIterator struct {
	box      Box
	children boxIterator

	currentBox Box // box to return
	stackBox   Box // may store an additional Box to yield

	test           func(Box) bool
	improper       []Box
	wrapperBoxType BoxType
}

func (iter *wrapImproperIterator) Next() bool {
	if iter.stackBox != nil {
		iter.currentBox = iter.stackBox
		iter.stackBox = nil
		iter.improper = iter.improper[:0]
		return true
	}

	for iter.children.Next() { // process the next input box
		child := iter.children.Box()
		if iter.test(child) {
			if len(iter.improper) > 0 {
				wrapper := iter.wrapperBoxType.AnonymousFrom(iter.box, nil)
				// Apply the rules again on the new wrapper
				iter.currentBox = tableBoxesChildren(wrapper, iter.improper)
				iter.stackBox = child
			} else {
				iter.currentBox = child
			}
			return true
		} else {
			// Whitespace either fail the test or were removed earlier,
			// so there is no need to take special care with the definition
			// of "consecutive".
			if FlexContainerT.IsInstance(iter.box) {
				// The display value of a flex item must be "blockified", see
				// https://www.w3.org/TR/css-flexbox-1/#flex-items
			} else {
				iter.improper = append(iter.improper, child)
			}
		}
	}

	if len(iter.improper) > 0 {
		wrapper := iter.wrapperBoxType.AnonymousFrom(iter.box, nil)
		// Apply the rules again on the new wrapper
		iter.currentBox = tableBoxesChildren(wrapper, iter.improper)
		iter.improper = iter.improper[:0]
		return true
	}

	return false
}

func (iter *wrapImproperIterator) Box() Box { return iter.currentBox }

//	Wrap consecutive children that do not pass ``test`` in a box of type
//
// “test“ defaults to children being of the same type as “wrapperType“.
func wrapImproper(box Box, children boxIterator, wrapperBoxType BoxType, test func(Box) bool) *wrapImproperIterator {
	if test == nil {
		test = wrapperBoxType.IsInstance
	}
	return &wrapImproperIterator{box: box, children: children, wrapperBoxType: wrapperBoxType, test: test}
}

// Remove and add boxes according to the table model.
//
// Take and return a “Box“ object.
//
// See https://www.w3.org/TR/CSS21/tables.html#anonymous-boxes
func AnonymousTableBoxes(box Box) Box {
	if !ParentT.IsInstance(box) || box.Box().IsRunning() {
		return box
	}

	// Do recursion.
	boxChildren := box.Box().Children
	children := make([]Box, len(boxChildren))
	for index, child := range boxChildren {
		children[index] = AnonymousTableBoxes(child)
	}
	return tableBoxesChildren(box, children)
}

// Internal implementation of AnonymousTableBoxes().box
func tableBoxesChildren(box Box, children []Box) Box {
	if TableColumnT.IsInstance(box) { // rule 1.1
		// Remove all children.
		children = nil
	} else if tableBox, ok := box.(*TableColumnGroupBox); ok { // rule 1.2
		// Remove children other than table-column.
		newChildren := make([]Box, 0, len(children))
		for _, child := range children {
			if TableColumnT.IsInstance(child) {
				newChildren = append(newChildren, child)
			}
		}
		children = newChildren

		// Rule XXX (not in the spec): column groups have at least
		// one column child.
		if len(children) == 0 {
			span := tableBox.span()
			if span < 1 {
				span = 1
			}

			for i := 0; i < span; i++ {
				children = append(children, TableColumnBoxAnonymousFrom(box, nil))
			}
		}
	}

	// rule 1.3
	if box.Box().tabularContainer && len(children) >= 2 {
		// Last child
		internal, text := children[len(children)-2], children[len(children)-1]

		if internal.Box().internalTableOrCaption && isWhitespace(text, nil) {
			children = children[:len(children)-1]
		}
		// First child
		if len(children) >= 2 {
			text, internal = children[0], children[1]
			if internal.Box().internalTableOrCaption && isWhitespace(text, nil) {
				children = children[1:]
			}
		}
		// Children other than first and last that would be removed by
		// rule 1.3 are also removed by rule 1.4 below.
	}

	newChildren, maxIndex := make([]Box, 0, len(children)), len(children)-1
	for index, child := range children {
		// Ignore some whitespace: rule 1.4
		var prevChild, nextChild Box
		if index != 0 {
			prevChild = children[index-1]
		}
		if index != maxIndex {
			nextChild = children[index+1]
		}
		if !(prevChild != nil && prevChild.Box().internalTableOrCaption && nextChild != nil && nextChild.Box().internalTableOrCaption && isWhitespace(child, nil)) {
			newChildren = append(newChildren, child)
		}
	}
	children = newChildren

	var childrenIter boxIterator = newBoxIter(children)
	if TableT.IsInstance(box) {
		// Rule 2.1
		childrenIter = wrapImproper(box, childrenIter, TableRowT,
			func(child Box) bool { return child.Box().properTableChild })
	} else if TableRowGroupT.IsInstance(box) {
		// Rule 2.2
		childrenIter = wrapImproper(box, childrenIter, TableRowT, nil)
	}

	if TableRowT.IsInstance(box) {
		// Rule 2.3
		childrenIter = wrapImproper(box, childrenIter, TableCellT, nil)
	} else {
		// Rule 3.1
		childrenIter = wrapImproper(box, childrenIter, TableRowT, func(child Box) bool {
			return !TableCellT.IsInstance(child)
		})
	}
	// Rule 3.2
	if InlineT.IsInstance(box) {
		childrenIter = wrapImproper(box, childrenIter, InlineTableT,
			func(child Box) bool {
				return !child.Box().properTableChild
			})
	} else {
		parentType := box.Type()
		childrenIter = wrapImproper(box, childrenIter, TableT,
			func(child Box) bool {
				return !child.Box().properTableChild || parentType.IsInProperParents(child.Type())
			})
	}

	if tableBox, ok := box.(TableBoxITF); ok {
		return wrapTable(tableBox, childrenIter)
	}
	box.Box().Children = collectBoxes(childrenIter)
	return box
}

// Take a table box and return it in its table wrapper box.
// Also re-order children and assign grid positions to each column and cell.
// Because of colspan/rowspan works, gridY is implicitly the index of a row,
// but gridX is an explicit attribute on cells, columns and column group.
// https://www.w3.org/TR/CSS21/tables.html#model
// https://www.w3.org/TR/CSS21/tables.html#table-layout
//
// wrapTable will panic if box's children are not table boxes
func wrapTable(box TableBoxITF, children boxIterator) Box {
	// Group table children by type
	var columns, rows, allCaptions []Box
	byType := map[BoxType]*[]Box{
		TableColumnT:      &columns,
		TableColumnGroupT: &columns,
		TableRowT:         &rows,
		TableRowGroupT:    &rows,
		TableCaptionT:     &allCaptions,
	}

	for children.Next() {
		child := children.Box()
		*byType[child.Type()] = append(*byType[child.Type()], child)
	}

	// Split top and bottom captions
	var captionTop, captionBottom []Box
	for _, caption := range allCaptions {
		switch caption.Box().Style.GetCaptionSide() {
		case "top":
			captionTop = append(captionTop, caption)
		case "bottom":
			captionBottom = append(captionBottom, caption)
		}
	}
	// Assign X positions on the grid to column boxes
	columnGroups := collectTableColumnGroupBoxs(wrapImproper(box, newBoxIter(columns), TableColumnGroupT, nil))
	gridX := 0
	for _, group := range columnGroups {
		group.GridX = gridX
		if len(group.Children) > 0 {
			for _, column := range group.Children {
				// There's no need to take care of group's span, as "span=x"
				// already generates x TableColumnBox children
				column.Box().GridX = gridX
				gridX += 1
			}
		} else {
			gridX += group.span()
		}
	}
	gridWidth := gridX

	rowGroups := collectBoxes(wrapImproper(box, newBoxIter(rows), TableRowGroupT, nil))
	// Extract the optional header and footer groups.
	var (
		bodyRowGroups []Box
		header        Box
		footer        Box
	)
	for _, _group := range rowGroups {
		group := _group.Box()
		display := group.Style.GetDisplay()
		if display == (pr.Display{"table-header-group"}) && header == nil {
			group.IsHeader = true
			header = _group
		} else if display == (pr.Display{"table-footer-group"}) && footer == nil {
			group.IsFooter = true
			footer = _group
		} else {
			bodyRowGroups = append(bodyRowGroups, _group)
		}
	}

	rowGroups = nil
	if header != nil {
		rowGroups = []Box{header}
	}
	rowGroups = append(rowGroups, bodyRowGroups...)
	if footer != nil {
		rowGroups = append(rowGroups, footer)
	}

	// Assign a (x,y) position in the grid to each cell.
	// rowspan can not extend beyond a row group, so each row group
	// is independent.
	// https://www.w3.org/TR/CSS21/tables.html#table-layout
	// Column 0 is on the left if direction is ltr, right if rtl.
	// This algorithm does not change.
	gridHeight := 0
	for _, group := range rowGroups {
		// Indexes: row number in the group.
		// Values: set of cells already occupied by row-spanning cells.
		groupChildren := group.Box().Children
		occupiedCellsByRow := make([]map[int]bool, len(groupChildren))
		// init the maps
		for i := range occupiedCellsByRow {
			occupiedCellsByRow[i] = make(map[int]bool)
		}
		for _, row := range groupChildren {
			occupiedCellsInThisRow := occupiedCellsByRow[0]
			occupiedCellsByRow = occupiedCellsByRow[1:]
			// The list is now about rows after this one.
			gridX = 0
			for _, _cell := range row.Box().Children {
				cell := _cell.Box()
				// Make sure that the first grid cell is free.
				for occupiedCellsInThisRow[gridX] {
					gridX += 1
				}
				cell.GridX = gridX
				newGridX := gridX + cell.Colspan
				// https://www.w3.org/TR/html401/struct/tables.html#adef-rowspan
				if cell.Rowspan != 1 {
					maxRowspan := len(occupiedCellsByRow) + 1
					var spannedRows []map[int]bool
					if cell.Rowspan == 0 {
						// All rows until the end of the group
						spannedRows = occupiedCellsByRow
						cell.Rowspan = maxRowspan
					} else {
						cell.Rowspan = utils.MinInt(cell.Rowspan, maxRowspan)
						spannedRows = occupiedCellsByRow[:cell.Rowspan-1]
					}
					for _, occupiedCells := range spannedRows {
						for i := gridX; i < newGridX; i++ {
							occupiedCells[i] = true
						}
					}
				}
				gridX = newGridX
				gridWidth = utils.MaxInt(gridWidth, gridX)
			}
		}
		gridHeight += len(groupChildren)
	}
	table := CopyWithChildren(box, rowGroups).(TableBoxITF)
	table.Box().Style = table.Box().Style.Copy()
	tableBox := table.Table()
	tableBox.ColumnGroups = columnGroups
	if tableBox.Style.GetBorderCollapse() == "collapse" {
		tableBox.CollapsedBorderGrid = collapseTableBorders(table, gridWidth, gridHeight)
	}
	var wrapperAFT func(Box, []Box) Box
	if InlineTableT.IsInstance(box) {
		wrapperAFT = InlineBlockT.AnonymousFrom
	} else {
		wrapperAFT = BlockT.AnonymousFrom
	}
	wrapper := wrapperAFT(box, append(append(captionTop, table), captionBottom...))
	wrapperBox := wrapper.Box()
	wrapperBox.Style = wrapperBox.Style.Copy()
	wrapperBox.IsTableWrapper = true
	// Non-inherited properties of the table element apply to one
	// of the wrapper and the table. The other get the initial value.
	wbStyle, tbStyle := wrapperBox.Style, table.Box().Style
	for name := range pr.TableWrapperBoxProperties {
		key := name.Key()
		wbStyle.Set(key, tbStyle.Get(key))
		tbStyle.Set(key, pr.InitialValues[name])
	}

	return wrapper
}

type Score [3]float64

func (s Score) Lower(other Score) bool {
	return s[0] < other[0] || (s[0] == other[0] && (s[1] < other[1] || (s[1] == other[1] && s[2] < other[2])))
}

type Border struct {
	Style pr.String
	Score Score
	Width utils.Fl
	Color pr.Color
}

type BorderGrids struct {
	Vertical, Horizontal [][]Border
}

// Resolve border conflicts for a table in the collapsing border model.
//
// Set appropriate border widths on the table,
// column group, column, row group, row and cell boxes; and return
// a data structure for the resolved collapsed border grid.
func collapseTableBorders(table TableBoxITF, gridWidth, gridHeight int) BorderGrids {
	if gridWidth == 0 || gridHeight == 0 {
		// Don’t bother with empty tables
		return BorderGrids{}
	}

	weakNullBorder := Border{Score: Score{0, 0, styleScores["none"]}, Style: "none", Width: 0, Color: transparent}

	verticalBorders, horizontalBorders := make([][]Border, gridHeight), make([][]Border, gridHeight+1)
	for y := range horizontalBorders {
		l1, l2 := make([]Border, gridWidth+1), make([]Border, gridWidth)
		for x := range l2 {
			l1[x] = weakNullBorder
			l2[x] = weakNullBorder
		}
		l1[gridWidth] = weakNullBorder
		if y < gridHeight {
			verticalBorders[y] = l1
		}
		horizontalBorders[y] = l2
	}

	// side follows the [bottom, left, right, top] convention
	const (
		bottom pr.KnownProp = iota
		left
		right
		top
	)
	setOneBorder := func(borderGrid [][]Border, boxStyle pr.ElementStyle, side pr.KnownProp, gridX, gridY int) {
		style := boxStyle.Get((pr.PBorderBottomStyle + side*5).Key()).(pr.String)
		width := boxStyle.Get((pr.PBorderBottomWidth + side*5).Key()).(pr.DimOrS)
		color := tree.ResolveColor(boxStyle, pr.PBorderBottomColor+side*5)

		// https://www.w3.org/TR/CSS21/tables.html#border-conflict-resolution
		score := Score{0, float64(width.Value), styleScores[style]}
		if style == "hidden" {
			score[0] = 1
		}

		_style, in := styleMap[style]
		if in {
			style = _style
		}

		previousScore := borderGrid[gridY][gridX].Score
		// Strict < so that the earlier call wins in case of a tie.
		if previousScore.Lower(score) {
			borderGrid[gridY][gridX] = Border{Score: score, Style: style, Width: utils.Fl(width.Value), Color: color}
		}
	}

	setBorders := func(box Box, x, y, w, h int) {
		style := box.Box().Style
		for yy := y; yy < y+h; yy++ {
			setOneBorder(verticalBorders, style, left, x, yy)
			setOneBorder(verticalBorders, style, right, x+w, yy)
		}
		for xx := x; xx < x+w; xx++ {
			setOneBorder(horizontalBorders, style, top, xx, y)
			setOneBorder(horizontalBorders, style, bottom, xx, y+h)
		}
	}

	// The order is important here:
	// "A style set on a cell wins over one on a row, which wins over a
	//  row group, column, column group and, lastly, table"
	// See https://www.w3.org/TR/CSS21/tables.html#border-conflict-resolution
	strongNullBorder := Border{Score: Score{1, 0, styleScores["hidden"]}, Style: "hidden", Width: 0, Color: transparent}

	gridY := 0
	for _, rowGroup := range table.Box().Children {
		for _, row := range rowGroup.Box().Children {
			for _, _cell := range row.Box().Children {
				cell := _cell.Box()
				// No border inside of a cell with rowspan || colspan
				for xx := cell.GridX + 1; xx < cell.GridX+cell.Colspan; xx++ {
					for yy := gridY; yy < gridY+cell.Rowspan; yy++ {
						verticalBorders[yy][xx] = strongNullBorder
					}
				}
				for xx := cell.GridX; xx < cell.GridX+cell.Colspan; xx++ {
					for yy := gridY + 1; yy < gridY+cell.Rowspan; yy++ {
						horizontalBorders[yy][xx] = strongNullBorder
					}
				}
				// The cell’s own borders
				setBorders(_cell, cell.GridX, gridY, cell.Colspan, cell.Rowspan)
			}
			gridY += 1
		}
	}

	gridY = 0
	for _, rowGroup := range table.Box().Children {
		for _, row := range rowGroup.Box().Children {
			setBorders(row, 0, gridY, gridWidth, 1)
			gridY += 1
		}
	}

	gridY = 0
	for _, rowGroup := range table.Box().Children {
		rowspan := len(rowGroup.Box().Children)
		setBorders(rowGroup, 0, gridY, gridWidth, rowspan)
		gridY += rowspan
	}

	for _, columnGroup := range table.Table().ColumnGroups {
		for _, column := range columnGroup.Children {
			setBorders(column, column.Box().GridX, 0, 1, gridHeight)
		}
	}

	for _, columnGroup := range table.Table().ColumnGroups {
		setBorders(columnGroup, columnGroup.GridX, 0, columnGroup.span(), gridHeight)
	}

	setBorders(table, 0, 0, gridWidth, gridHeight)

	// Now that all conflicts are resolved, set transparent borders of
	// the correct widths on each box. The actual border grid will be
	// painted separately.
	setBorderUsedWidth := func(box Box, side pr.KnownProp, twiceWidth utils.Fl) {
		box_, value := box.Box(), pr.Float(twiceWidth/2)
		switch side {
		case top:
			box_.BorderTopWidth = value
		case right:
			box_.BorderRightWidth = value
		case bottom:
			box_.BorderBottomWidth = value
		case left:
			box_.BorderLeftWidth = value
		}
	}

	removeBorders := func(box Box) {
		setBorderUsedWidth(box, top, 0)
		setBorderUsedWidth(box, right, 0)
		setBorderUsedWidth(box, bottom, 0)
		setBorderUsedWidth(box, left, 0)
	}

	maxVerticalWidth := func(x, y, h int) utils.Fl {
		var max utils.Fl
		for _, gridRow := range verticalBorders[y : y+h] {
			width := gridRow[x].Width
			if width > max {
				max = width
			}
		}
		return max
	}

	maxHorizontalWidth := func(x, y, w int) utils.Fl {
		var max utils.Fl
		for _, _s := range horizontalBorders[y][x : x+w] {
			width := _s.Width
			if width > max {
				max = width
			}
		}
		return max
	}

	gridY = 0
	for _, rowGroup := range table.Box().Children {
		removeBorders(rowGroup)
		for _, row := range rowGroup.Box().Children {
			removeBorders(row)
			for _, _cell := range row.Box().Children {
				cell := _cell.Box()
				setBorderUsedWidth(_cell, top, maxHorizontalWidth(cell.GridX, gridY, cell.Colspan))
				setBorderUsedWidth(_cell, bottom, maxHorizontalWidth(cell.GridX, gridY+cell.Rowspan, cell.Colspan))
				setBorderUsedWidth(_cell, left, maxVerticalWidth(cell.GridX, gridY, cell.Rowspan))
				setBorderUsedWidth(_cell, right, maxVerticalWidth(cell.GridX+cell.Colspan, gridY, cell.Rowspan))
			}
			gridY += 1
		}
	}

	for _, columnGroup := range table.Table().ColumnGroups {
		removeBorders(columnGroup)
		for _, column := range columnGroup.Children {
			removeBorders(column)
		}
	}

	setBorderUsedWidth(table, top, maxHorizontalWidth(0, 0, gridWidth))
	setBorderUsedWidth(table, bottom, maxHorizontalWidth(0, gridHeight, gridWidth))
	// "UAs must compute an initial left && right border width for the table
	// by examining the first && last cells in the first row of the table."
	// https://www.w3.org/TR/CSS21/tables.html#collapsing-borders
	// ... so h=1, not gridHeight :
	setBorderUsedWidth(table, left, maxVerticalWidth(0, 0, 1))
	setBorderUsedWidth(table, right, maxVerticalWidth(gridWidth, 0, 1))

	return BorderGrids{Vertical: verticalBorders, Horizontal: horizontalBorders}
}

// Remove and add boxes according to the flex model.
// See https://www.w3.org/TR/css-flexbox-1/#flex-items
func FlexBoxes(box Box) Box {
	if !ParentT.IsInstance(box) || box.Box().IsRunning() {
		return box
	}

	// Do recursion.
	children := make([]Box, len(box.Box().Children))
	for i, child := range box.Box().Children {
		children[i] = FlexBoxes(child)
	}
	box.Box().Children = flexChildren(box, children)
	return box
}

func flexChildren(box Box, children []Box) []Box {
	if _, isFlexCont := box.(FlexContainerBoxITF); isFlexCont {
		var flexChildren []Box
		for _, child := range children {
			if !child.Box().IsAbsolutelyPositioned() {
				child.Box().IsFlexItem = true
			}

			if textBox, ok := child.(*TextBox); ok {
				// https://www.w3.org/TR/css-flexbox-1/#flex-items
				if strings.Trim(textBox.TextS(), " ") == "" {
					continue
				}
			}

			if _, ok := child.(InlineLevelBoxITF); ok {
				anonymous := BlockBoxAnonymousFrom(box, []Box{child})
				anonymous.IsFlexItem = true
				flexChildren = append(flexChildren, anonymous)
			} else {
				flexChildren = append(flexChildren, child)
			}
		}
		return flexChildren
	}
	return children
}

// Remove and add boxes according to the grid model.
// See https://drafts.csswg.org/css-grid-2/#grid-item
func GridBoxes(box Box) Box {
	if !ParentT.IsInstance(box) || box.Box().IsRunning() {
		return box
	}
	// Do recursion.
	children := make([]Box, len(box.Box().Children))
	for i, child := range box.Box().Children {
		children[i] = GridBoxes(child)
	}
	box.Box().Children = gridChildren(box, children)
	return box
}

func gridChildren(box Box, children []Box) []Box {
	if GridContainerT.IsInstance(box) {
		var gridChildren []Box
		for _, child := range children {
			if !child.Box().IsAbsolutelyPositioned() {
				child.Box().IsGridItem = true
			}
			if text, ok := child.(*TextBox); ok && strings.Trim(text.TextS(), " ") == "" {
				// TODO: ignore texts only containing "characters that can be
				// affected by the white-space property"
				// https://drafts.csswg.org/css-grid-2/#grid-item
				continue
			}
			if InlineLevelT.IsInstance(child) {
				anonymous := BlockBoxAnonymousFrom(child, []Box{child})
				anonymous.Box().Style = child.Box().Style
				child.Box().IsGridItem = false
				anonymous.Box().IsGridItem = true
				gridChildren = append(gridChildren, anonymous)
			} else {
				gridChildren = append(gridChildren, child)
			}
		}
		return gridChildren
	}
	return children
}

// ProcessWhitespace executes the first part of "The 'white-space' processing model".
// See https://www.w3.org/TR/CSS21/text.html#white-space-model
// and https://drafts.csswg.org/css-text-3/#white-space-rules
// The default value of followingCollapsibleSpace shoud be `false`.
func ProcessWhitespace(box Box, followingCollapsibleSpace bool) bool {
	if box_, isTextBox := box.(*TextBox); isTextBox {
		text := box_.TextS()
		if len(text) == 0 {
			return followingCollapsibleSpace
		}

		// Normalize line feeds
		text = lineFeedRe.ReplaceAllString(text, "\n")

		styleWhiteSpace := box_.Style.GetWhiteSpace()
		newLineCollapse := styleWhiteSpace == "normal" || styleWhiteSpace == "nowrap"
		spaceCollapse := styleWhiteSpace == "normal" || styleWhiteSpace == "nowrap" || styleWhiteSpace == "pre-line"

		if spaceCollapse {
			// \r characters were removed/converted earlier
			text = tabRe.ReplaceAllString(text, "\n")
		}

		if newLineCollapse {
			// Could also replace with a zero width space character (U+200B),
			// or no character
			// CSS3: https://www.w3.org/TR/css3-text/#line-break-transform
			text = strings.ReplaceAll(text, "\n", " ")
		}
		if spaceCollapse {
			text = spaceRe.ReplaceAllString(text, " ")
			previousText := text
			if followingCollapsibleSpace && strings.HasPrefix(text, " ") {
				text = text[1:]
				box_.LeadingCollapsibleSpace = true
			}
			followingCollapsibleSpace = strings.HasSuffix(previousText, " ")
		} else {
			followingCollapsibleSpace = false
		}
		box_.Text = []rune(text)
	} else {
		for _, child := range box.Box().Children {
			switch child.(type) {
			case *TextBox, *InlineBox: // leaf
				chidlCollapsibleSpace := ProcessWhitespace(child, followingCollapsibleSpace)
				if box.Box().IsInNormalFlow() && child.Box().IsInNormalFlow() {
					followingCollapsibleSpace = chidlCollapsibleSpace
				}
			default:
				if child.Box().IsInNormalFlow() {
					followingCollapsibleSpace = false
				}
			}
		}
	}

	return followingCollapsibleSpace && !box.Box().IsRunning()
}

// Capitalize words according to CSS’s "text-transform: capitalize".
func capitalize(text string) string {
	letterFound := false
	var out strings.Builder
	for _, letter := range text {
		isLN := unicode.Is(unicode.L, letter) || unicode.Is(unicode.N, letter)
		if !letterFound && isLN {
			letterFound = true
			letter = unicode.ToUpper(letter)
		} else if isZ := unicode.Is(unicode.Z, letter); isZ {
			letterFound = false
		}
		out.WriteRune(letter)
	}
	return out.String()
}

func ProcessTextTransform(box Box) {
	if tb, ok := box.(*TextBox); ok {
		text, style := tb.TextS(), tb.Style
		textTransform := tb.Style.GetTextTransform()
		if textTransform != "none" {
			switch textTransform {
			case "uppercase":
				text = strings.ToUpper(text)
			case "lowercase":
				text = strings.ToLower(text)
			// Python’s unicode.captitalize is not the same.
			case "capitalize":
				text = capitalize(text)
			case "full-width":
				text = strings.Map(func(u rune) rune {
					rep, in := asciiToWide[u]
					if !in {
						return u
					}
					return rep
				}, text)
			}
		}
		if style.GetHyphens() == "none" {
			text = strings.ReplaceAll(text, "\u00AD", "") //  U+00AD SOFT HYPHEN (SHY)
		}
		tb.Text = []rune(text)
	} else if !box.Box().IsRunning() {
		// recursion
		for _, child := range box.Box().Children {
			if TextT.IsInstance(child) || InlineT.IsInstance(child) {
				ProcessTextTransform(child)
			}
		}
	}
}

// Build the structure of lines inside blocks and return a new box tree.
//
// Consecutive inline-level boxes in a block container box are wrapped into a
// line box, itself wrapped into an anonymous block box.
//
// This line box will be broken into multiple lines later.
//
// This is the first case in
// https://www.w3.org/TR/CSS21/visuren.html#anonymous-block-level
//
// Example:
//
//	BlockBox[
//	    TextBox["Some "],
//	    InlineBox[TextBox["text"]],
//	    BlockBox[
//	        TextBox["More text"],
//	    ]
//	]
//
// is turned into::
//
//	BlockBox[
//	    AnonymousBlockBox[
//	        LineBox[
//	            TextBox["Some "],
//	            InlineBox[TextBox["text"]],
//	        ]
//	    ]
//	    BlockBox[
//	        LineBox[
//	            TextBox["More text"],
//	        ]
//	    ]
//	]
func InlineInBlock(box Box) Box {
	if len(box.Box().Children) == 0 || box.Box().IsRunning() {
		return box
	}
	baseBox := box.Box()
	boxChildren := baseBox.Children

	if len(boxChildren) > 0 && !baseBox.LeadingCollapsibleSpace {
		baseBox.LeadingCollapsibleSpace = boxChildren[0].Box().LeadingCollapsibleSpace
	}

	var children []Box
	trailingCollapsibleSpace := false
	for _, child := range boxChildren {
		// Keep track of removed collapsing spaces for wrap opportunities, and
		// remove empty text boxes.
		// (They may have been emptied by ProcessWhitespace().)

		if trailingCollapsibleSpace {
			child.Box().LeadingCollapsibleSpace = true
		}

		if textBox, isTextBox := child.(*TextBox); isTextBox && len(textBox.Text) == 0 {
			trailingCollapsibleSpace = child.Box().LeadingCollapsibleSpace
		} else {
			trailingCollapsibleSpace = false
			children = append(children, InlineInBlock(child))
		}
	}
	if !baseBox.TrailingCollapsibleSpace {
		baseBox.TrailingCollapsibleSpace = trailingCollapsibleSpace
	}

	if !BlockContainerT.IsInstance(box) {
		baseBox.Children = children
		return box
	}

	var newLineChildren, newChildren []Box
	for _, childBox := range children {
		if LineT.IsInstance(childBox) {
			panic("childBox can't be a LineBox")
		}
		if len(newLineChildren) > 0 && childBox.Box().IsAbsolutelyPositioned() {
			newLineChildren = append(newLineChildren, childBox)
		} else if InlineLevelT.IsInstance(childBox) || (len(newLineChildren) > 0 && !childBox.Box().IsInNormalFlow()) {
			// Do not append white space at the start of a line :
			// it would be removed during layout.
			childTextBox, isTextBox := childBox.(*TextBox)
			st := childBox.Box().Style.GetWhiteSpace()
			// Sequence of white-space was collapsed to a single space by ProcessWhitespace().
			if len(newLineChildren) > 0 || !(isTextBox && childTextBox.TextS() == " " && (st == "normal" || st == "nowrap" || st == "pre-line")) {
				newLineChildren = append(newLineChildren, childBox)
			}
		} else {
			if len(newLineChildren) > 0 {
				// Inlines are consecutive no more: add this line box
				// and create a new one.
				lineBox := LineBoxAnonymousFrom(box, newLineChildren)
				anonymous := BlockBoxAnonymousFrom(box, []Box{lineBox})
				newChildren = append(newChildren, anonymous)
				newLineChildren = nil
			}
			newChildren = append(newChildren, childBox)
		}
	}
	if len(newLineChildren) > 0 {
		// There were inlines at the end
		lineBox := LineBoxAnonymousFrom(box, newLineChildren)
		if len(newChildren) > 0 {
			anonymous := BlockBoxAnonymousFrom(box, []Box{lineBox})
			newChildren = append(newChildren, anonymous)
		} else {
			// Only inline-level children: one line box
			newChildren = append(newChildren, lineBox)
		}
	}

	baseBox.Children = newChildren
	return box
}

// Build the structure of blocks inside lines.
//
// Inline boxes containing block-level boxes will be broken in two
// boxes on each side on consecutive block-level boxes, each side wrapped
// in an anonymous block-level box.
//
// This is the second case in
// https://www.w3.org/TR/CSS21/visuren.html#anonymous-block-level
//
//	Eg. if this is given::
//
//	    BlockBox[
//	        LineBox[
//	            InlineBox[
//	                TextBox["Hello."],
//	            ],
//	            InlineBox[
//	                TextBox["Some "],
//	                InlineBox[
//	                    TextBox["text"]
//	                    BlockBox[LineBox[TextBox["More text"]]],
//	                    BlockBox[LineBox[TextBox["More text again"]]],
//	                ],
//	                BlockBox[LineBox[TextBox["And again."]]],
//	            ]
//	        ]
//	    ]
//
//	this is returned::
//
//	    BlockBox[
//	        AnonymousBlockBox[
//	            LineBox[
//	                InlineBox[
//	                    TextBox["Hello."],
//	                ],
//	                InlineBox[
//	                    TextBox["Some "],
//	                    InlineBox[TextBox["text"]],
//	                ]
//	            ]
//	        ],
//	        BlockBox[LineBox[TextBox["More text"]]],
//	        BlockBox[LineBox[TextBox["More text again"]]],
//	        AnonymousBlockBox[
//	            LineBox[
//	                InlineBox[
//	                ]
//	            ]
//	        ],
//	        BlockBox[LineBox[TextBox["And again."]]],
//	        AnonymousBlockBox[
//	            LineBox[
//	                InlineBox[
//	                ]
//	            ]
//	        ],
//	    ]
func BlockInInline(box Box) Box {
	if len(box.Box().Children) == 0 || box.Box().IsRunning() {
		return box
	}

	var newChildren []Box
	changed := false

	for _, child := range box.Box().Children {
		var newChild Box
		if LineT.IsInstance(child) {
			if len(box.Box().Children) != 1 {
				panic(fmt.Sprintf("Line boxes should have no siblings at this stage, got %v.", box.Box().Children))
			}

			var (
				stack          tree.ResumeStack
				newLine, block Box
			)
			for {
				newLine, block, stack = innerBlockInInline(child, stack)
				if block == nil {
					break
				}
				anon := BlockBoxAnonymousFrom(box, []Box{newLine})
				newChildren = append(newChildren, anon)
				newChildren = append(newChildren, BlockInInline(block))
				// Loop with the same child and the new stack.
			}

			if len(newChildren) > 0 {
				// Some children were already added, this became a block
				// context.
				newChild = BlockBoxAnonymousFrom(box, []Box{newLine})
			} else {
				// Keep the single line box as-is, without anonymous blocks.
				newChild = newLine
			}
		} else {
			// Not in an inline formatting context.
			newChild = BlockInInline(child)
		}

		if newChild != child {
			changed = true
		}
		newChildren = append(newChildren, newChild)
	}
	if changed {
		box.Box().Children = newChildren
	}
	return box
}

// Find a block-level box in an inline formatting context.
// If one is found, return “(newBox, blockLevelBox, resumeAt)“.
// “newBox“ contains all of “box“ content before the block-level box.
// “resumeAt“ can be passed as “skipStack“ in a new call to
// this function to resume the search just after the block-level box.
// If no block-level box is found after the position marked by
// “skipStack“, return “(newBox, None, None)“
func innerBlockInInline(box Box, skipStack tree.ResumeStack) (Box, Box, tree.ResumeStack) {
	var newChildren []Box
	var blockLevelBox Box
	var resumeAt tree.ResumeStack
	changed := false

	isStart := skipStack == nil
	var skip int
	if isStart {
		skip = 0
	} else {
		skip, skipStack = skipStack.Unpack()
	}

	hasBroken := false
	for i, child := range box.Box().Children[skip:] {
		index := i + skip
		if BlockLevelT.IsInstance(child) && child.Box().IsInNormalFlow() {
			if skipStack != nil {
				panic("Should not skip here")
			}
			blockLevelBox = child
			index += 1 // Resume *after* the block
		} else {
			var newChild Box
			if InlineT.IsInstance(child) {
				newChild, blockLevelBox, resumeAt = innerBlockInInline(child, skipStack)
				skipStack = nil
			} else {
				if skipStack != nil {
					panic("Should not skip here")
				}
				newChild = BlockInInline(child)
				// blockLevelBox is still None.
			}

			if newChild != child {
				changed = true
			}
			newChildren = append(newChildren, newChild)
		}

		if blockLevelBox != nil {
			resumeAt = tree.ResumeStack{index: resumeAt}
			box = CopyWithChildren(box, newChildren)
			hasBroken = true
			break
		}
	}
	if !hasBroken {
		if changed || skip > 0 {
			box = CopyWithChildren(box, newChildren)
		}
	}

	return box, blockLevelBox, resumeAt
}

// Set a “ViewportOverflow“ attribute on the box for the root element.
//
//	Like backgrounds, ``overflow`` on the root element must be propagated
//	to the viewport.
//
//	See https://www.w3.org/TR/CSS21/visufx.html#overflow
func setViewportOverflow(rootBox Box) Box {
	chosenBox := rootBox
	if strings.ToLower(rootBox.Box().ElementTag()) == "html" &&
		rootBox.Box().Style.GetOverflow() == "visible" {

		for _, child := range rootBox.Box().Children {
			if strings.ToLower(child.Box().ElementTag()) == "body" {
				chosenBox = child
				break
			}
		}
	}
	rootBox.Box().ViewportOverflow = string(chosenBox.Box().Style.GetOverflow())
	chosenBox.Box().Style.SetOverflow("visible")
	return rootBox
}

func boxText(box Box) string {
	if tBox, is := box.(*TextBox); is {
		return tBox.TextS()
	}
	var builder []rune
	if ParentT.IsInstance(box) {
		for _, child := range Descendants(box) {
			pt := child.Box().PseudoType
			if child, ok := child.(*TextBox); ok && pt != "before" && pt != "after" && pt != "marker" {
				builder = append(builder, child.Text...)
			}
		}
	}
	return string(builder)
}

func extractText(textPart string, box Box) string {
	switch textPart {
	case "text", "content":
		return boxText(box)
	case "before", "after":
		var builder strings.Builder
		if ParentT.IsInstance(box) {
			for _, child := range Descendants(box) {
				et := child.Box().PseudoType
				if et == textPart && !ParentT.IsInstance(child) {
					builder.WriteString(boxText(child))
				}
			}
		}
		return builder.String()
	case "first-letter":
		characterFound := false
		firstLetter := ""
		for _, letter := range boxText(box) {
			isPunc := unicode.In(letter, TableFirstLetter...)
			if !isPunc {
				if characterFound {
					break
				}
				characterFound = true
			}
			firstLetter += string(letter)
		}
		return firstLetter
	default:
		panic("exhaustive switch")
	}
}
