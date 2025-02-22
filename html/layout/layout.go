// Transform a "before layout" box tree into an "after layout" tree,
// by breaking boxes across lines and pages; and determining the position and dimension
// of each box fragment.
//
// Boxes in the new tree have `used values` in their PositionX,
// PositionY, Width and Height attributes, amongst others.
// (see https://www.w3.org/TR/CSS21/cascade.html#used-value)
//
// The laid out pages are ready to be printed or display on screen,
// which is done by the higher level `document` package.
package layout

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/benoitkugler/webrender/css/counters"
	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/images"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/text/hyphen"
	"github.com/benoitkugler/webrender/utils"
	"github.com/benoitkugler/webrender/utils/testutils/tracer"
	"golang.org/x/net/html"
)

// if true, save a structured trace in an external file
// this is a costly operation, and should only be used for debugging purpose
const traceMode = false

var traceLogger tracer.Tracer // used only when traceMode is true

func init() {
	if traceMode {
		traceLogger = tracer.NewTracer(filepath.Join(os.TempDir(), "trace_go.txt"))
	}
}

type Box = bo.Box

// Layout lay out the whole document, returning one box per pages.
//
// This includes line breaks, page breaks, absolute size and position for all
// boxes.
func Layout(html *tree.HTML, stylesheets []tree.CSS, presentationalHints bool, fontConfig text.FontConfiguration) []*bo.PageBox {
	counterStyle := make(counters.CounterStyle)
	context := newLayoutContext(html, stylesheets, presentationalHints, fontConfig, counterStyle)

	logger.ProgressLogger.Println("Step 4 - Creating formatting structure")

	rootBox := bo.BuildFormattingStructure(html.Root, context.styleFor, context.resolver,
		html.BaseUrl, &context.TargetCollector, counterStyle, &context.footnotes)

	return layoutDocument(html, rootBox, context, -1)
}

// Initialize “context.pageMaker“.
// Collect the pagination's states required for page based counters.
func initializePageMaker(context *layoutContext, rootBox bo.BoxFields) {
	context.pageMaker = nil

	// Special case the root box
	pageBreak := rootBox.Style.GetBreakBefore()

	// TODO: take care of text direction and writing mode
	// https://www.w3.org/TR/css3-page/#progression
	var rightPage bool
	switch pageBreak {
	case "right":
		rightPage = true
	case "left":
		rightPage = false
	case "recto":
		rightPage = rootBox.Style.GetDirection() == "ltr"
	case "verso":
		rightPage = rootBox.Style.GetDirection() == "rtl"
	default:
		rightPage = rootBox.Style.GetDirection() == "ltr"
	}
	pv, _ := rootBox.PageValues()
	nextPage := tree.PageBreak{Break: "any", Page: pv}

	// pageState is prerequisite for filling in missing page based counters
	// although neither a variable quoteDepth nor counterScopes are needed
	// in page-boxes -- reusing
	// `formattingStructure.bo.updateCounters()` to avoid redundant
	// code requires a full `state`.
	// The value of **pages**, of course, is unknown until we return and
	// might change when "contentChanged" triggers re-pagination...
	// So we start with an empty state
	pageState := tree.PageState{
		// Shared mutable objects:
		QuoteDepth:    []int{0}, // quoteDepth: single integer
		CounterValues: tree.CounterValues{"pages": []int{0}},
		CounterScopes: []utils.Set{utils.NewSet("pages")}, // counterScopes
	}

	// Initial values
	remakeState := tree.RemakeState{}
	context.pageMaker = append(context.pageMaker, tree.PageMaker{
		InitialResumeAt: nil, InitialNextPage: nextPage, RightPage: rightPage,
		InitialPageState: pageState, RemakeState: remakeState,
	})
}

// Lay out and yield the fixed boxes of “pages“.
func layoutFixedBoxes(context *layoutContext, pages []*bo.PageBox, containingPage *bo.PageBox) []Box {
	var out []Box
	for _, page := range pages {
		for _, box := range page.FixedBoxes {
			// As replaced boxes are never copied during layout, ensure that we
			// have different boxes (with a possibly different layout) for
			// each pages.
			if bo.ReplacedT.IsInstance(box) {
				box = box.Copy()
			}
			// Absolute boxes in fixed boxes are rendered as fixed boxes'
			// children, even when they are fixed themselves.
			var absoluteBoxes []*AbsolutePlaceholder
			b, _ := absoluteBoxLayout(context, box, containingPage, &absoluteBoxes, -pr.Inf, nil)
			out = append(out, b)
			for len(absoluteBoxes) != 0 {
				var newAbsoluteBoxes []*AbsolutePlaceholder
				for _, absBox := range absoluteBoxes {
					absoluteLayout(context, absBox, containingPage, &newAbsoluteBoxes, -pr.Inf, nil)
				}
				absoluteBoxes = newAbsoluteBoxes
			}
		}
	}
	return out
}

func layoutDocument(doc *tree.HTML, rootBox bo.BlockLevelBoxITF, context *layoutContext, maxLoops int) []*bo.PageBox {
	initializePageMaker(context, *rootBox.Box())
	if maxLoops == -1 {
		maxLoops = 8 // default value
	}
	var (
		pages             []*bo.PageBox
		originalFootnotes = append([]Box(nil), context.footnotes...) // copy
	)
	actualTotalPages := 0

	for loop := 0; loop < maxLoops; loop += 1 {
		if loop > 0 {
			logger.ProgressLogger.Printf("Step 5 - Creating layout - Repagination #%d \n", loop)
			context.footnotes = append([]Box(nil), originalFootnotes...)
		}

		initialTotalPages := actualTotalPages
		pages = context.makeAllPages(rootBox, doc, pages)
		actualTotalPages = len(pages)

		// Check whether another round is required
		reloopContent := false
		reloopPages := false
		for _, pageData := range context.pageMaker {
			// Update pages
			pageCounterValues := pageData.InitialPageState.CounterValues
			pageCounterValues["pages"] = []int{actualTotalPages}
			if pageData.RemakeState.ContentChanged {
				reloopContent = true
			}
			if pageData.RemakeState.PagesWanted {
				reloopPages = initialTotalPages != actualTotalPages
			}
		}

		// No need for another loop, stop here
		if !reloopContent && !reloopPages {
			break
		}
	}

	// Calculate string-sets and bookmark-label containing page based counters
	// when pagination is finished. No need to do that (maybe multiple times) in
	// makePage because they dont create boxes, only appear in MarginBoxes and
	// in the final PDF.

	// Prevent repetition of bookmarks (see #1145).
	watchElements, watchElementsBefore, watchElementsAfter := map[*html.Node]bool{}, map[*html.Node]bool{}, map[*html.Node]bool{}

	for i, page := range pages {
		// We need the updated pageCounterValues
		pageCounterValues := context.pageMaker[i+1].InitialPageState.CounterValues

		for _, child := range bo.Descendants(page) {
			childBox := child.Box()
			// Only one bookmark per original box
			if childBox.BookmarkLabel != "" {
				var checklist map[*html.Node]bool
				if childBox.PseudoType == "before" {
					checklist = watchElementsBefore
				} else if childBox.PseudoType == "after" {
					checklist = watchElementsAfter
				} else {
					checklist = watchElements
				}

				if checklist[childBox.Element] {
					childBox.BookmarkLabel = ""
				} else {
					checklist[childBox.Element] = true
				}
			}

			if mLink := child.MissingLink(); mLink != nil {
				for key, item := range context.TargetCollector.CounterLookupItems {
					box, cssToken := key.SourceBox, key.CssToken
					if mLink == box && cssToken != "content" {
						if cssToken == "bookmark-label" && childBox.BookmarkLabel == "" {
							// don't refill it!
							continue
						}

						item.ParseAgain(pageCounterValues)

						if cssToken == "bookmark-label" {
							childBox.BookmarkLabel = box.GetBookmarkLabel()
						}
					}
				}
			}
			// Collect the stringSets in the LayoutContext
			stringSets := childBox.StringSet
			for _, stringSet := range stringSets {
				stringName, text := stringSet.Type, string(stringSet.Content.(pr.String))
				dict := context.stringSet[stringName]
				if dict == nil {
					dict = make(map[int][]string)
				}
				dict[i+1] = append(dict[i+1], text)
				context.stringSet[stringName] = dict
			}
		}
	}

	out := make([]*bo.PageBox, len(pages))
	// Add margin boxes
	for i, page := range pages {
		var rootChildren []Box
		root, footnoteArea := page.Box().Children[0], page.Box().Children[1]
		rootChildren = append(rootChildren, layoutFixedBoxes(context, pages[:i], page)...)
		rootChildren = append(rootChildren, root.Box().Children...)
		rootChildren = append(rootChildren, layoutFixedBoxes(context, pages[i+1:], page)...)
		root.Box().Children = rootChildren
		context.currentPage = i + 1 // pageNumber starts at 1

		// pageMaker's pageState is ready for the MarginBoxes
		state := context.pageMaker[context.currentPage].InitialPageState
		page.Children = []Box{root}
		if len(footnoteArea.Box().Children) != 0 {
			page.Children = append(page.Children, footnoteArea)
		}
		page.Children = append(page.Children, makeMarginBoxes(context, page, state)...)
		layoutBackgrounds(page, context.resolver.FetchImage)
		out[i] = page

		if traceMode {
			traceLogger.DumpTree(page, fmt.Sprintf("Final page %d", i))
		}
	}
	return out
}

var _ text.TextLayoutContext = (*layoutContext)(nil)

type brokenBox struct {
	box             Box
	containingBlock Box
	resumeAt        tree.ResumeStack
}

// layoutContext stores the global context needed during layout,
// such as various caches.
type layoutContext struct {
	// caches
	stringSet       map[string]map[int][]string
	runningElements map[string]map[int][]Box
	strutLayouts    map[text.StrutLayoutKey][2]pr.Float
	tables          map[*bo.TableBox]map[bool]tableContentWidths

	resolver            bo.URLResolver
	fontConfig          text.FontConfiguration
	TargetCollector     tree.TargetCollector
	counterStyle        counters.CounterStyle
	dictionaries        map[text.HyphenDictKey]hyphen.Hyphener
	styleFor            *tree.StyleFor
	pageMaker           []tree.PageMaker
	excludedShapes      *[]*bo.BoxFields
	excludedShapesLists [][]*bo.BoxFields
	brokenOutOfFlow     map[Box]brokenBox

	footnotes            []Box
	currentPageFootnotes []Box
	reportedFootnotes    []Box
	currentFootnoteArea  *bo.FootnoteAreaBox

	currentPage int
	pageBottom  pr.Float

	marginClearance bool
	forcedBreak     bool
	inColumn        bool
}

// presentationalHints=false,
func newLayoutContext(html *tree.HTML, stylesheets []tree.CSS,
	presentationalHints bool, fontConfig text.FontConfiguration, counterStyle counters.CounterStyle,
) *layoutContext {
	var (
		pageRules       []tree.PageRule
		userStylesheets = stylesheets
	)

	cache := images.NewCache()
	getImageFromUri := func(url, forcedMimeType string, orientation pr.SBoolFloat) images.Image {
		return images.GetImageFromUri(cache, html.UrlFetcher, false, url, forcedMimeType, orientation)
	}

	self := layoutContext{}
	self.resolver = bo.URLResolver{Fetch: html.UrlFetcher, FetchImage: getImageFromUri}
	self.fontConfig = fontConfig
	self.TargetCollector = tree.NewTargetCollector()
	self.counterStyle = counterStyle
	self.runningElements = make(map[string]map[int][]Box)
	self.brokenOutOfFlow = make(map[Box]brokenBox)

	// Cache
	self.stringSet = make(map[string]map[int][]string)
	self.dictionaries = make(map[text.HyphenDictKey]hyphen.Hyphener)
	self.strutLayouts = make(map[text.StrutLayoutKey][2]pr.Float)
	self.tables = map[*bo.TableBox]map[bool]tableContentWidths{}

	self.styleFor = tree.GetAllComputedStyles(html, userStylesheets, presentationalHints, fontConfig,
		counterStyle, &pageRules, &self.TargetCollector, false, &self)
	return &self
}

func (l *layoutContext) CurrentPage() int { return l.currentPage }

func (l *layoutContext) Fonts() text.FontConfiguration { return l.fontConfig }

func (l *layoutContext) HyphenCache() map[text.HyphenDictKey]hyphen.Hyphener {
	return l.dictionaries
}

func (l *layoutContext) StrutLayoutsCache() map[text.StrutLayoutKey][2]pr.Float {
	return l.strutLayouts
}

func (l *layoutContext) overflowsPage(bottomSpace, positionY pr.Float) bool {
	// Use a small fudge factor to avoid floating numbers errors.
	// The 1e-9 value comes from PEP 485.
	return positionY > (l.pageBottom-bottomSpace)*(1+1e-9)
}

func (l *layoutContext) createBlockFormattingContext() {
	l.excludedShapesLists = append(l.excludedShapesLists, nil)
	l.excludedShapes = &l.excludedShapesLists[len(l.excludedShapesLists)-1]
}

func (l *layoutContext) finishBlockFormattingContext(rootBox_ Box) {
	// See https://www.w3.org/TR/CSS2/visudet.html#root-height
	rootBox := rootBox_.Box()
	if rootBox.Style.GetHeight().S == "auto" && len(*l.excludedShapes) != 0 {
		boxBottom := rootBox.ContentBoxY() + rootBox.Height.V()
		maxShapeBottom := boxBottom
		for _, shape := range *l.excludedShapes {
			v := shape.PositionY + shape.MarginHeight()
			if v > maxShapeBottom {
				maxShapeBottom = v
			}
		}
		rootBox.Height = rootBox.Height.V() + maxShapeBottom - boxBottom
	}
	l.excludedShapesLists = l.excludedShapesLists[:len(l.excludedShapesLists)-1]
	if L := len(l.excludedShapesLists); L != 0 {
		l.excludedShapes = &l.excludedShapesLists[L-1]
	} else {
		l.excludedShapes = nil
	}
}

func resolveKeyword(keyword, name string, page Box) string {
	switch keyword {
	case "first":
		return "first"
	case "start":
		element := page
		for element != nil {
			if element.Box().Style.GetStringSet().String != "none" {
				for _, v := range element.Box().Style.GetStringSet().Contents {
					if v.String == name {
						return "first"
					}
				}
			}
			if len(element.Box().Children) > 0 {
				element = element.Box().Children[0]
				continue
			}
			break
		}
	case "last":
		return "last"
	case "first-except":
		return "return"
	}
	return ""
}

// Resolve value of string function (as set by string set).
// We'll have something like this that represents all assignments on a
// given page:
//
// {1: [u"First Header"], 3: [u"Second Header"],
//
//	4: [u"Third Header", u"3.5th Header"]}
//
// Value depends on current page.
// https://drafts.csswg.org/csswg/css-gcpm/#funcdef-string
//
// `keyword` indicates which value of the named string to use.
// Default is the first assignment on the current page
// else the most recent assignment (entry value)
// keyword="first"
func (lc *layoutContext) GetStringSetFor(page Box, name, keyword string) string {
	if currentS, in := lc.stringSet[name][lc.currentPage]; in {
		// A value was assigned on this page
		switch resolveKeyword(keyword, name, page) {
		case "first":
			return currentS[0]
		case "last":
			return currentS[len(currentS)-1]
		case "return":
			return ""
		}
	}
	// Search backwards through previous pages
	for previousPage := lc.currentPage - 1; previousPage > 0; previousPage -= 1 {
		if currentS, in := lc.stringSet[name][previousPage]; in {
			return currentS[len(currentS)-1]
		}
	}
	return ""
}

func (lc *layoutContext) GetRunningElementFor(page Box, name, keyword string) Box {
	if currentS, in := lc.runningElements[name][lc.currentPage]; in {
		// A value was assigned on this page
		switch resolveKeyword(keyword, name, page) {
		case "first":
			return currentS[0]
		case "last":
			return currentS[len(currentS)-1]
		case "return":
			return nil
		}
	}
	// Search backwards through previous pages
	for previousPage := lc.currentPage - 1; previousPage > 0; previousPage -= 1 {
		if currentS, in := lc.runningElements[name][previousPage]; in {
			return currentS[len(currentS)-1]
		}
	}
	return nil
}

// Add a footnote to the layout for this page.
func (l *layoutContext) layoutFootnote(footnote Box) bool {
	removeFromBoxes(&l.footnotes, footnote)
	l.currentPageFootnotes = append(l.currentPageFootnotes, footnote)
	return l.updateFootnoteArea()
}

// Remove a footnote from the layout and return it to the waitlist.
func (l *layoutContext) unlayoutFootnote(footnote Box) {
	if !isInBoxes(footnote, l.footnotes) {
		l.footnotes = append(l.footnotes, footnote)
		if isInBoxes(footnote, l.currentPageFootnotes) {
			removeFromBoxes(&l.currentPageFootnotes, footnote)
		} else if isInBoxes(footnote, l.reportedFootnotes) {
			removeFromBoxes(&l.reportedFootnotes, footnote)
		}
		l.updateFootnoteArea()
	}
}

// Mark a footnote as being moved to the next page.
func (l *layoutContext) reportFootnote(footnote Box) {
	removeFromBoxes(&l.currentPageFootnotes, footnote)
	l.reportedFootnotes = append(l.reportedFootnotes, footnote)
	l.updateFootnoteArea()
}

// Update the page bottom size and our footnote area height.
func (l *layoutContext) updateFootnoteArea() bool {
	if l.currentFootnoteArea.Height != pr.AutoF && !l.inColumn {
		l.pageBottom += l.currentFootnoteArea.MarginHeight()
	}
	l.currentFootnoteArea.Children = l.currentPageFootnotes
	if len(l.currentFootnoteArea.Children) != 0 {
		footnoteArea := bo.CreateAnonymousBox(bo.Deepcopy(l.currentFootnoteArea)).(bo.BlockLevelBoxITF)
		footnoteArea, _, _ = blockLevelLayout(
			l, footnoteArea, -pr.Inf, nil,
			&l.currentFootnoteArea.Page.BoxFields, true, nil, nil, nil, false, -1)
		l.currentFootnoteArea.Height = footnoteArea.Box().Height
		if !l.inColumn {
			l.pageBottom -= footnoteArea.Box().MarginHeight()
		}
		lastChild := footnoteArea.Box().Children[len(footnoteArea.Box().Children)-1]
		overflow := (lastChild.Box().PositionY+lastChild.Box().MarginHeight() >
			footnoteArea.Box().PositionY+footnoteArea.Box().MarginHeight()-
				footnoteArea.Box().MarginBottom.V())
		return overflow
	} else {
		l.currentFootnoteArea.Height = pr.Float(0)
		return false
	}
}
