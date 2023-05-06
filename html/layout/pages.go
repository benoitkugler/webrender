package layout

import (
	"fmt"
	"strings"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"
)

// Layout for pages and CSS3 margin boxes.

type contentSizer interface {
	minContentSize() pr.Float
	maxContentSize() pr.Float
}

type orientedBoxITF interface {
	baseBox() *orientedBox
	restoreBoxAttributes()
}

type orientedBox struct {
	// abstract, must be implemented by subclasses
	contentSizer

	context                 *layoutContext
	box                     Box // either *bo.PageBox or *bo.MarginBox
	marginA, marginB, inner pr.MaybeFloat
	paddingPlusBorder       pr.Float
}

func (o *orientedBox) baseBox() *orientedBox {
	return o
}

func (o orientedBox) sugar() pr.Float {
	return o.paddingPlusBorder + o.marginA.V() + o.marginB.V()
}

func (o orientedBox) outer() pr.Float {
	return o.sugar() + o.inner.V()
}

func (o orientedBox) outerMinContentSize() pr.Float {
	if o.inner == pr.AutoF {
		return o.sugar() + o.minContentSize()
	}
	return o.sugar() + o.inner.V()
}

func (o orientedBox) outerMaxContentSize() pr.Float {
	if o.inner == pr.AutoF {
		return o.sugar() + o.maxContentSize()
	}
	return o.sugar() + o.inner.V()
}

func (o *orientedBox) shrinkToFit(available pr.Float) {
	o.inner = pr.Min(pr.Max(o.minContentSize(), available), o.maxContentSize())
}

type verticalBox struct {
	orientedBox
}

func newVerticalBox(context *layoutContext, box Box) *verticalBox {
	self := new(verticalBox)
	self.context = context
	self.box = box
	// Inner dimension: that of the content area, as opposed to the
	// outer dimension: that of the margin area.
	box_ := box.Box()
	self.inner = box_.Height
	self.marginA = box_.MarginTop
	self.marginB = box_.MarginBottom
	self.paddingPlusBorder = box_.PaddingTop.V() + box_.PaddingBottom.V() +
		box_.BorderTopWidth.V() + box_.BorderBottomWidth.V()
	self.orientedBox.contentSizer = self
	return self
}

func (self *verticalBox) restoreBoxAttributes() {
	box := self.box.Box()
	box.Height = self.inner
	box.MarginTop = self.marginA
	box.MarginBottom = self.marginB
}

// TODO: Define what are the min-content && max-content heights
func (self *verticalBox) minContentSize() pr.Float {
	return 0
}

func (self *verticalBox) maxContentSize() pr.Float {
	return 1e6
}

type horizontalBox struct {
	_minContentSize, _maxContentSize pr.MaybeFloat
	orientedBox
}

func newHorizontalBox(context *layoutContext, box Box) *horizontalBox {
	self := new(horizontalBox)
	self.context = context
	self.box = box
	box_ := box.Box()
	self.inner = box_.Width
	self.marginA = box_.MarginLeft
	self.marginB = box_.MarginRight
	self.paddingPlusBorder = box_.PaddingLeft.V() + box_.PaddingRight.V() +
		box_.BorderLeftWidth.V() + box_.BorderRightWidth.V()
	self.orientedBox.contentSizer = self
	return self
}

func (self *horizontalBox) restoreBoxAttributes() {
	box := self.box.Box()
	box.Width = self.inner
	box.MarginLeft = self.marginA
	box.MarginRight = self.marginB
}

func (self *horizontalBox) minContentSize() pr.Float {
	if self._minContentSize == nil {
		self._minContentSize = minContentWidth(self.context, self.box, false)
	}
	return self._minContentSize.V()
}

func (self *horizontalBox) maxContentSize() pr.Float {
	if self._maxContentSize == nil {
		self._maxContentSize = maxContentWidth(self.context, self.box, false)
	}
	return self._maxContentSize.V()
}

func countAuto(v1, v2, v3 pr.MaybeFloat) int {
	out := 0
	if v1 == pr.AutoF {
		out++
	}
	if v2 == pr.AutoF {
		out++
	}
	if v3 == pr.AutoF {
		out++
	}
	return out
}

// Compute and set a margin box fixed dimension on “box“.
//
// Described in: https://drafts.csswg.org/css-page-3/#margin-constraints
//
//   - box: The margin box to work on
//   - outer: The target outer dimension (value of a page margin)
//   - vertical: true to set height, margin-top and margin-bottom; false for width,
//     margin-left and margin-right
//   - topOrLeft: true if the margin box in if the top half (for vertical==true) or
//     left half (for vertical==false) of the page.
//     This determines which margin should be "auto" if the values are
//     over-constrained. (Rule 3 of the algorithm.)
func computeFixedDimension(context *layoutContext, box_ *bo.MarginBox, outer pr.Float, vertical, topOrLeft bool) {
	var boxOriented orientedBoxITF
	if vertical {
		boxOriented = newVerticalBox(context, box_)
	} else {
		boxOriented = newHorizontalBox(context, box_)
	}
	box := boxOriented.baseBox()

	// Rule 2
	total := box.paddingPlusBorder
	for _, value := range [3]pr.MaybeFloat{box.marginA, box.marginB, box.inner} {
		if value != pr.AutoF {
			total += value.V()
		}
	}
	if total > outer {
		if box.marginA == pr.AutoF {
			box.marginA = pr.Float(0)
		}
		if box.marginB == pr.AutoF {
			box.marginB = pr.Float(0)
		}
		if box.inner == pr.AutoF {
			// XXX this is not in the spec, but without it box.inner
			// would end up with a negative value.
			// Instead, this will trigger rule 3 below.
			// https://lists.w3.org/Archives/Public/www-style/2012Jul/0006.html
			box.inner = pr.Float(0)
		}
	}
	// Rule 3
	if countAuto(box.marginA, box.marginB, box.inner) == 0 {
		// Over-constrained
		if topOrLeft {
			box.marginA = pr.AutoF
		} else {
			box.marginB = pr.AutoF
		}
	}
	// Rule 4
	if countAuto(box.marginA, box.marginB, box.inner) == 1 {
		if box.inner == pr.AutoF {
			box.inner = outer - box.paddingPlusBorder - box.marginA.V() - box.marginB.V()
		} else if box.marginA == pr.AutoF {
			box.marginA = outer - box.paddingPlusBorder - box.marginB.V() - box.inner.V()
		} else if box.marginB == pr.AutoF {
			box.marginB = outer - box.paddingPlusBorder - box.marginA.V() - box.inner.V()
		}
	}
	// Rule 5
	if box.inner == pr.AutoF {
		if box.marginA == pr.AutoF {
			box.marginA = pr.Float(0)
		}
		if box.marginB == pr.AutoF {
			box.marginB = pr.Float(0)
		}
		box.inner = outer - box.paddingPlusBorder - box.marginA.V() - box.marginB.V()
	}
	// Rule 6
	if box.marginA == pr.AutoF && box.marginB == pr.AutoF {
		v := (outer - box.paddingPlusBorder - box.inner.V()) / 2
		box.marginA = v
		box.marginB = v
	}

	if countAuto(box.marginA, box.marginB, box.inner) > 0 {
		panic(fmt.Sprintf("unexpected auto value in %v", box))
	}

	boxOriented.restoreBoxAttributes()
}

// Compute and set a margin box fixed dimension on “box“
//
// Described in: https://drafts.csswg.org/css-page-3/#margin-dimension
//
//   - sideBoxes: Three boxes on a same side (as opposed to a corner.)
//     A list of:
//
//     -- A @*-left or @*-top margin box
//     -- A @*-center or @*-middle margin box
//     -- A @*-right or @*-bottom margin box
//
//   - vertical:
//     true to set height, margin-top and margin-bottom; false for width,
//     margin-left and margin-right
//
//   - outerSum:
//     The target total outer dimension (max box width or height)
func computeVariableDimension(context *layoutContext, sideBoxes_ [3]*bo.MarginBox, vertical bool, outerSum pr.Float) {
	var sideBoxes [3]orientedBoxITF
	for i, box_ := range sideBoxes_ {
		if vertical {
			sideBoxes[i] = newVerticalBox(context, box_)
		} else {
			sideBoxes[i] = newHorizontalBox(context, box_)
		}
	}
	boxA, boxB, boxC := sideBoxes[0].baseBox(), sideBoxes[1].baseBox(), sideBoxes[2].baseBox()

	for _, box_ := range sideBoxes {
		box := box_.baseBox()
		if box.marginA == pr.AutoF {
			box.marginA = pr.Float(0)
		}
		if box.marginB == pr.AutoF {
			box.marginB = pr.Float(0)
		}
	}

	if boxB.box.(*bo.MarginBox).IsGenerated {
		if boxB.inner == pr.AutoF {
			acMaxContentSize := 2 * pr.Max(boxA.outerMaxContentSize(), boxC.outerMaxContentSize())
			if outerSum >= (boxB.outerMaxContentSize() + acMaxContentSize) {
				boxB.inner = boxB.maxContentSize()
			} else {
				acMinContentSize := 2 * pr.Max(boxA.outerMinContentSize(), boxC.outerMinContentSize())
				boxB.inner = boxB.minContentSize()
				available := outerSum - boxB.outer() - acMinContentSize
				if available > 0 {
					weightAc := acMaxContentSize - acMinContentSize
					weightB := boxB.maxContentSize() - boxB.minContentSize()
					weightSum := weightAc + weightB
					// By definition of maxContentSize and minContentSize,
					// weights can not be negative. weightSum == 0 implies that
					// maxContentSize == minContentSize for each box, in
					// which case the sum can not be both <= and > outerSum
					// Therefore, one of the last two "if" statements would not
					// have lead us here.
					if weightSum <= 0 {
						panic(fmt.Sprintf("weightSum must be > 0, got %f", weightSum))
					}
					boxB.inner = boxB.inner.V() + available*weightB/weightSum
				}
			}
		}
		if boxA.inner == pr.AutoF {
			boxA.shrinkToFit((outerSum-boxB.outer())/2 - boxA.sugar())
		}
		if boxC.inner == pr.AutoF {
			boxC.shrinkToFit((outerSum-boxB.outer())/2 - boxC.sugar())
		}
	} else {
		// Non-generated boxes get zero for every box-model property
		if boxB.inner.V() != 0 {
			panic(fmt.Sprintf("expected boxB.inner == 0, got %v", boxB.inner))
		}
		if boxA.inner == pr.AutoF && boxC.inner == pr.AutoF {
			if outerSum >= (boxA.outerMaxContentSize() + boxC.outerMaxContentSize()) {
				boxA.inner = boxA.maxContentSize()
				boxC.inner = boxC.maxContentSize()
			} else {
				boxA.inner = boxA.minContentSize()
				boxC.inner = boxC.minContentSize()
				available := outerSum - boxA.outer() - boxC.outer()
				if available > 0 {
					weightA := boxA.maxContentSize() - boxA.minContentSize()
					weightC := boxC.maxContentSize() - boxC.minContentSize()
					weightSum := weightA + weightC
					// By definition of maxContentSize and minContentSize,
					// weights can ! be negative. weightSum == 0 implies that
					// maxContentSize == minContentSize for each box, in
					// which case the sum can ! be both <= and > outerSum
					// Therefore, one of the last two "if" statements would not
					// have lead us here.
					if weightSum <= 0 {
						panic(fmt.Sprintf("weightSum must be > 0, got %f", weightSum))
					}
					boxA.inner = boxA.inner.V() + available*weightA/weightSum
					boxC.inner = boxC.inner.V() + available*weightC/weightSum
				}
			}
		} else if boxA.inner == pr.AutoF {
			boxA.shrinkToFit(outerSum - boxC.outer() - boxA.sugar())
		} else if boxC.inner == pr.AutoF {
			boxC.shrinkToFit(outerSum - boxA.outer() - boxC.sugar())
		}
	}

	// And, we’re done!
	if countAuto(boxA.inner, boxB.inner, boxC.inner) > 0 {
		panic("unexpected auto value")
	}

	// Set the actual attributes back.
	for _, box := range sideBoxes {
		box.restoreBoxAttributes()
	}
}

// Drop "pages" counter from style in @page and @margin context.
// Ensure `counter-increment: page` for @page context if not otherwise
// manipulated by the style.
func standardizePageBasedCounters(style pr.ElementStyle, pseudoType string) {
	pageCounterTouched := false
	for _, propname := range [...]string{"counter_set", "counter_reset", "counter_increment"} {
		prop := style.Get(propname).(pr.SIntStrings)
		if prop.String == "auto" {
			style.Set(propname, pr.SIntStrings{Values: pr.IntStrings{}})
			continue
		}
		var justifiedValues pr.IntStrings
		for _, v := range prop.Values {
			if v.String == "page" {
				pageCounterTouched = true
			}
			if v.String != "pages" {
				justifiedValues = append(justifiedValues, v)
			}
		}
		style.Set(propname, pr.SIntStrings{Values: justifiedValues})
	}

	if pseudoType == "" && !pageCounterTouched {
		current := style.GetCounterIncrement()
		newInc := append(pr.IntStrings{{String: "page", Int: 1}}, current.Values...)
		style.SetCounterIncrement(pr.SIntStrings{Values: newInc})
	}
}

// Yield laid-out margin boxes for this page.
// “state“ is the actual, up-to-date page-state from
// “context.pageMaker[context.currentPage]“.
func makeMarginBoxes(context *layoutContext, page *bo.PageBox, state tree.PageState) []Box {
	// This is a closure only to make calls shorter

	// Return a margin box with resolved percentages.
	// The margin box may still have "auto" values.
	// Return ``None`` if this margin box should not be generated.
	// :param atKeyword: which margin box to return, eg. "@top-left"
	// :param containingBlock: as expected by :func:`resolvePercentages`.
	makeBox := func(atKeyword string, containingBlock bo.MaybePoint) *bo.MarginBox {
		style := context.styleFor.Get(page.PageType, atKeyword)
		if style == nil {
			// doesn't affect counters
			style = tree.ComputedFromCascaded(nil, nil, page.Style, nil, "", "", nil, context)
		}
		standardizePageBasedCounters(style, atKeyword)
		box := bo.NewMarginBox(atKeyword, style)
		// Empty boxes should not be generated, but they may be needed for
		// the layout of their neighbors.
		// TODO: should be the computed value.
		ct := style.GetContent().String
		box.IsGenerated = !(ct == "normal" || ct == "inhibit" || ct == "none")
		// TODO: get actual counter values at the time of the last page break
		if box.IsGenerated {
			// @margins mustn't manipulate page-context counters
			marginState := state.Copy()
			// quoteDepth, counterValues, counterScopes = marginState
			// TODO: check this, probably useless
			marginState.CounterScopes = append(marginState.CounterScopes, utils.NewSet())
			bo.UpdateCounters(&marginState, box.Style)
			box.Children = bo.ContentToBoxes(
				box.Style, box, marginState.QuoteDepth, marginState.CounterValues,
				context.resolver, &context.TargetCollector, context.counterStyle, context,
				page)
			bo.ProcessWhitespace(box, false)
			bo.ProcessTextTransform(box)
			box = bo.CreateAnonymousBox(box).(*bo.MarginBox) // type stable
		}
		resolvePercentages(box, containingBlock, "")
		boxF := box.Box()
		if !box.IsGenerated {
			boxF.Width = pr.Float(0)
			boxF.Height = pr.Float(0)
			for _, side := range [4]bo.Side{bo.STop, bo.SRight, bo.SBottom, bo.SLeft} {
				boxF.ResetSpacing(side)
			}
		}
		return box
	}

	marginTop := page.MarginTop.V()
	marginBottom := page.MarginBottom
	marginLeft := page.MarginLeft.V()
	marginRight := page.MarginRight
	maxBoxWidth := page.BorderWidth()
	maxBoxHeight := page.BorderHeight()

	// bottom right corner of the border box
	pageEndX := marginLeft + maxBoxWidth
	pageEndY := marginTop + maxBoxHeight

	// Margin box dimensions, described in
	// https://drafts.csswg.org/csswg/css3-page/#margin-box-dimensions
	var generatedBoxes []*bo.MarginBox
	prefixs := [4]string{"top", "bottom", "left", "right"}
	verticals := [4]bool{false, false, true, true}
	containingBlocks := [4]bo.MaybePoint{
		{maxBoxWidth, marginTop},
		{maxBoxWidth, marginBottom},
		{marginLeft, maxBoxHeight},
		{marginRight, maxBoxHeight},
	}
	positionXs := [4]pr.Float{marginLeft, marginLeft, 0, pageEndX}
	positionYs := [4]pr.Float{0, pageEndY, marginTop, marginTop}

	for i := range prefixs {
		prefix, vertical, containingBlock, positionX, positionY := prefixs[i], verticals[i], containingBlocks[i], positionXs[i], positionYs[i]

		suffixes := [3]string{"left", "center", "right"}
		variableOuter, fixedOuter := containingBlock[0], containingBlock[1]
		if vertical {
			suffixes = [3]string{"top", "middle", "bottom"}
			fixedOuter, variableOuter = containingBlock[0], containingBlock[1]
		}
		var sideBoxes [3]*bo.MarginBox
		anyIsGenerated := false
		for i, suffix := range suffixes {
			sideBoxes[i] = makeBox(fmt.Sprintf("@%s-%s", prefix, suffix), containingBlock)
			if sideBoxes[i].IsGenerated {
				anyIsGenerated = true
			}
		}

		if !anyIsGenerated {
			continue
		}
		// We need the three boxes together for the variable dimension:
		computeVariableDimension(context, sideBoxes, vertical, variableOuter.V())
		offsets := [...]pr.Float{0, 0.5, 1}
		for i := range sideBoxes {
			box, offset := sideBoxes[i], offsets[i]
			if !box.IsGenerated {
				continue
			}
			box.PositionY = positionY
			box.PositionX = positionX
			if vertical {
				box.PositionY += offset * (variableOuter.V() - box.MarginHeight())
			} else {
				box.PositionX += offset * (variableOuter.V() - box.MarginWidth())
			}
			computeFixedDimension(context, box, fixedOuter.V(), !vertical, prefix == "top" || prefix == "left")
			generatedBoxes = append(generatedBoxes, box)
		}
	}

	atKeywords := [4]string{"@top-left-corner", "@top-right-corner", "@bottom-left-corner", "@bottom-right-corner"}
	cbWidths := [4]pr.MaybeFloat{marginLeft, marginRight, marginLeft, marginRight}
	cbHeights := [4]pr.MaybeFloat{marginTop, marginTop, marginBottom, marginBottom}
	positionXs = [4]pr.Float{0, pageEndX, 0, pageEndX}
	positionYs = [4]pr.Float{0, 0, pageEndY, pageEndY}
	// Corner boxes
	for i := range atKeywords {
		atKeyword, cbWidth, cbHeight, positionX, positionY := atKeywords[i], cbWidths[i], cbHeights[i], positionXs[i], positionYs[i]
		box := makeBox(atKeyword, bo.MaybePoint{cbWidth, cbHeight})
		if !box.IsGenerated {
			continue
		}
		box.PositionX = positionX
		box.PositionY = positionY
		computeFixedDimension(context, box, cbHeight.V(), true, strings.Contains(atKeyword, "top"))
		computeFixedDimension(context, box, cbWidth.V(), false, strings.Contains(atKeyword, "left"))
		generatedBoxes = append(generatedBoxes, box)
	}

	out := make([]Box, len(generatedBoxes))
	for i, box := range generatedBoxes {
		out[i] = marginBoxContentLayout(context, box)
	}
	return out
}

// Layout a margin box’s content once the box has dimensions.
func marginBoxContentLayout(context *layoutContext, mBox *bo.MarginBox) Box {
	var positionedBoxes []*AbsolutePlaceholder
	newBox_, tmp, _ := blockContainerLayout(context, mBox, -pr.Inf, nil, true,
		&positionedBoxes, &positionedBoxes, new([]pr.Float), false, -1)

	if tmp.resumeAt != nil {
		panic(fmt.Sprintf("resumeAt should be nil, got %v", tmp.resumeAt))
	}

	for _, absBox := range positionedBoxes {
		absoluteLayout(context, absBox, mBox, &positionedBoxes, 0, nil)
	}

	box := newBox_.Box()
	verticalAlign := box.Style.GetVerticalAlign()
	// Every other value is read as "top", ie. no change.
	if L := len(box.Children); (verticalAlign.String == "middle" || verticalAlign.String == "bottom") && L != 0 {
		firstChild := box.Children[0]
		lastChild := box.Children[L-1].Box()
		top := firstChild.Box().PositionY
		// Not always exact because floating point errors
		// assert top == box.contentBoxY()
		bottom := lastChild.PositionY + lastChild.MarginHeight()
		contentHeight := bottom - top
		offset := box.Height.V() - contentHeight
		if verticalAlign.String == "middle" {
			offset /= 2
		}
		for _, child := range box.Children {
			child.Translate(child, 0, offset, false)
		}
	}
	return newBox_
}

// Take a :class:`OrientedBox` object and set either width, margin-left
// and margin-right; or height, margin-top and margin-bottom.
// "The width and horizontal margins of the page box are then calculated
//
//	exactly as for a non-replaced block element := range normal flow. The height
//	and vertical margins of the page box are calculated analogously (instead
//	of using the block height formulas). In both cases if the values are
//	over-constrained, instead of ignoring any margins, the containing block
//	is resized to coincide with the margin edges of the page box."
//
// https://drafts.csswg.org/csswg/css3-page/#page-box-page-rule
// https://www.w3.org/TR/CSS21/visudet.html#blockwidth
func pageWidthOrHeight(box_ orientedBoxITF, containingBlockSize pr.Float) {
	box := box_.baseBox()
	remaining := containingBlockSize - box.paddingPlusBorder
	if box.inner == pr.AutoF {
		if box.marginA == pr.AutoF {
			box.marginA = pr.Float(0)
		}
		if box.marginB == pr.AutoF {
			box.marginB = pr.Float(0)
		}
		box.inner = remaining - box.marginA.V() - box.marginB.V()
	} else if box.marginA == pr.AutoF && box.marginB == pr.AutoF {
		box.marginA = (remaining - box.inner.V()) / 2
		box.marginB = box.marginA
	} else if box.marginA == pr.AutoF {
		box.marginA = remaining - box.inner.V() - box.marginB.V()
	} else if box.marginB == pr.AutoF {
		box.marginB = remaining - box.inner.V() - box.marginA.V()
	}
	box_.restoreBoxAttributes()
}

var (
	pageWidth  = handleMinMaxWidth(pageWidth_)
	pageHeight = handleMinMaxHeight(pageHeight_)
)

// @handleMinMaxWidth
// containingBlock must be block
func pageWidth_(box Box, context *layoutContext, containingBlock containingBlock) (bool, pr.Float) {
	pageWidthOrHeight(newHorizontalBox(context, box), containingBlock.(block).Width)
	return false, 0
}

// @handleMinMaxHeight
// containingBlock must be block
func pageHeight_(box Box, context *layoutContext, containingBlock containingBlock) (bool, pr.Float) {
	pageWidthOrHeight(newVerticalBox(context, box), containingBlock.(block).Height)
	return false, 0
}

// Take just enough content from the beginning to fill one page.
//
// Return “(page, finished)“. “page“ is a laid out PageBox object
// and “resumeAt“ indicates where in the document to start the next page,
// or is “None“ if this was the last page.
//
// pageNumber: integer, start at 1 for the first page
// resumeAt: as returned by “makePage()“ for the previous page,
//
//	or ``None`` for the first page.
func (context *layoutContext) makePage(rootBox bo.BlockLevelBoxITF, pageType utils.PageElement, resumeAt tree.ResumeStack,
	pageNumber int, pageState *tree.PageState,
) (*bo.PageBox, tree.ResumeStack, tree.PageBreak) {
	style := context.styleFor.Get(pageType, "")

	// Propagated from the root or <body>.
	style.SetOverflow(pr.String(rootBox.Box().ViewportOverflow))
	page := bo.NewPageBox(pageType, style)

	deviceSize_ := page.Style.GetSize()
	cbWidth, cbHeight := deviceSize_[0].Value, deviceSize_[1].Value
	resolvePercentages(page, bo.MaybePoint{cbWidth, cbHeight}, "")

	page.PositionX = 0
	page.PositionY = 0
	pageWidth(page, context, block{Width: cbWidth})
	pageHeight(page, context, block{Height: cbHeight})

	rootBox.Box().PositionX = page.ContentBoxX()
	rootBox.Box().PositionY = page.ContentBoxY()
	context.pageBottom = rootBox.Box().PositionY + page.Height.V()
	initialContainingBlock := page

	footnoteAreaStyle := context.styleFor.Get(pageType, "@footnote")
	footnoteArea := bo.NewFootnoteAreaBox(page, footnoteAreaStyle)
	resolvePercentages(footnoteArea, bo.MaybePoint{page.Width, page.Height}, "")
	footnoteArea.PositionX = page.ContentBoxX()
	footnoteArea.PositionY = context.pageBottom

	var previousResumeAt tree.ResumeStack
	if pageType.Blank {
		previousResumeAt = resumeAt
		rootBox = bo.CopyWithChildren(rootBox, nil).(bo.BlockLevelBoxITF) // CopyWithChildren is type stable
	}

	// TODO: handle cases where the root element is something else.
	// See https://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
	if !(bo.BlockT.IsInstance(rootBox) || bo.FlexContainerT.IsInstance(rootBox)) {
		panic(fmt.Sprintf("expected Block or FlexContainer, got %s", rootBox))
	}
	context.createBlockFormattingContext()
	context.currentPage = pageNumber
	context.currentPageFootnotes = nil
	context.currentFootnoteArea = footnoteArea

	reportedFootnotes := context.reportedFootnotes
	context.reportedFootnotes = nil

	for i, reportedFootnote := range reportedFootnotes {
		context.footnotes = append(context.footnotes, reportedFootnote)
		overflow := context.layoutFootnote(reportedFootnote)
		if overflow && i != 0 {
			context.reportFootnote(reportedFootnote)
			context.reportedFootnotes = context.reportedFootnotes[i:]
			break
		}
	}

	var (
		adjoiningMargins []pr.Float
		positionedBoxes  []*AbsolutePlaceholder // Mixed absolute and fixed
		outOfFlowBoxes   []Box
		contextOutOfFlow = context.brokenOutOfFlow
	)
	context.brokenOutOfFlow = make(map[Box]brokenBox) // new map
	for _, v := range contextOutOfFlow {
		box, containingBlock := v.box, v.containingBlock
		box.Box().PositionY = rootBox.Box().ContentBoxY()

		var (
			outOfFlowBox      Box
			outOfFlowResumeAt tree.ResumeStack
		)
		if box.Box().IsFloated() {
			outOfFlowBox, outOfFlowResumeAt = floatLayout(context, box, containingBlock.Box(),
				&positionedBoxes, &positionedBoxes, 0, v.resumeAt)
		} else {
			if !box.Box().IsAbsolutelyPositioned() {
				panic("internal error: box should be absolutely positioned")
			}
			outOfFlowBox, outOfFlowResumeAt = absoluteBoxLayout(context, box, containingBlock,
				&positionedBoxes, 0, v.resumeAt)
		}
		outOfFlowBoxes = append(outOfFlowBoxes, outOfFlowBox)
		if outOfFlowResumeAt != nil {
			context.brokenOutOfFlow[outOfFlowBox] = brokenBox{box, containingBlock, outOfFlowResumeAt}
		}
	}

	rootBox, tmp, _ := blockLevelLayout(context, rootBox, 0, resumeAt,
		&initialContainingBlock.BoxFields, true, &positionedBoxes, &positionedBoxes, &adjoiningMargins, false, -1)
	resumeAt = tmp.resumeAt
	if rootBox == nil {
		panic("expected non nil box for the root element")
	}

	rootBox.Box().Children = append(outOfFlowBoxes, rootBox.Box().Children...)

	footnoteArea = bo.CreateAnonymousBox(bo.Deepcopy(footnoteArea)).(*bo.FootnoteAreaBox)
	tmpBox, _, _ := blockLevelLayout(
		context, footnoteArea, -pr.Inf, nil, &footnoteArea.Page.BoxFields,
		true, &positionedBoxes, &positionedBoxes, new([]pr.Float), false, -1)
	footnoteArea = tmpBox.(*bo.FootnoteAreaBox)
	footnoteArea.Translate(footnoteArea, 0, -footnoteArea.MarginHeight(), false)

	for _, placeholder := range positionedBoxes {
		if placeholder.Box().Style.GetPosition().String == "fixed" {
			page.FixedBoxes = append(page.FixedBoxes, placeholder.AliasBox) // page.FixedBox is empty before this loop
		}
	}
	for i := 0; i < len(positionedBoxes); i++ { // note that positionedBoxes may grow over the loop
		absoluteLayout(context, positionedBoxes[i], page, &positionedBoxes, 0, nil)
	}

	context.finishBlockFormattingContext(rootBox)

	page.Children = []Box{rootBox, footnoteArea}
	descendants := bo.Descendants(page)
	for _, child := range positionedBoxes {
		descendants = append(descendants, bo.Descendants(child)...)
	}

	// Update page counter values
	standardizePageBasedCounters(style, "")
	bo.UpdateCounters(pageState, style)
	pageCounterValues := pageState.CounterValues
	// pageCounterValues will be cached in the pageMaker

	targetCollector := context.TargetCollector
	pageMaker := context.pageMaker

	// remakeState tells the makeAllPages-loop in layoutDocument()
	// whether and what to re-make.
	remakeState := &pageMaker[pageNumber-1].RemakeState

	// Evaluate and cache page values only once (for the first LineBox)
	// otherwise we suffer endless loops when the target/pseudo-element
	// spans across multiple pages
	cachedAnchors := utils.NewSet()
	cachedLookups := map[*tree.CounterLookupItem]bool{}

	for _, v := range pageMaker[:pageNumber-1] {
		cachedAnchors.Extend(v.RemakeState.Anchors)
		for _, u := range v.RemakeState.ContentLookups {
			cachedLookups[u] = true
		}
	}

	for _, child := range descendants {
		// Cache target's page counters
		anchor := string(child.Box().Style.GetAnchor())
		if anchor != "" && !cachedAnchors.Has(anchor) {
			remakeState.Anchors = append(remakeState.Anchors, anchor)
			cachedAnchors.Add(anchor)
			// Re-make of affected targeting boxes is inclusive
			targetCollector.CacheTargetPageCounters(anchor, pageCounterValues, pageNumber-1, pageMaker)
		}

		// string-set and bookmark-labels don't create boxes, only `content`
		// requires another call to makePage. There is maximum one "content"
		// item per box.
		var counterLookup *tree.CounterLookupItem
		if missingLink := child.MissingLink(); missingLink != nil {
			// A CounterLookupItem exists for the css-token "content"
			key := tree.NewFunctionKey(missingLink, "content")
			counterLookup = targetCollector.CounterLookupItems[key]
		}

		// Resolve missing (page based) counters
		if counterLookup != nil {
			callParseAgain := false

			// Prevent endless loops
			refreshMissingCounters := !cachedLookups[counterLookup]
			if refreshMissingCounters {
				remakeState.ContentLookups = append(remakeState.ContentLookups, counterLookup)
				cachedLookups[counterLookup] = true
				counterLookup.PageMakerIndex = tree.NewOptionnalInt(pageNumber - 1)
			}

			// Step 1: page based back-references
			// Marked as pending by targetCollector.cacheTargetPageCounters
			if counterLookup.Pending {
				if !pageCounterValues.Equal(counterLookup.CachedPageCounterValues) {
					counterLookup.CachedPageCounterValues = pageCounterValues.Copy()
				}
				counterLookup.Pending = false
				callParseAgain = true
			}

			// Step 2: local counters
			// If the box mixed-in page counters changed, update the content
			// and cache the new values.
			missingCounters := counterLookup.MissingCounters
			if len(missingCounters) != 0 {
				if missingCounters.Has("pages") {
					remakeState.PagesWanted = true
				}
				if refreshMissingCounters && !pageCounterValues.Equal(counterLookup.CachedPageCounterValues) {
					counterLookup.CachedPageCounterValues = pageCounterValues.Copy()
					for counterName := range missingCounters {
						counterValue := pageCounterValues[counterName]
						if counterValue != nil {
							callParseAgain = true
							// no need to loop them all
							break
						}
					}
				}
			}

			// Step 3: targeted counters
			targetMissing := counterLookup.MissingTargetCounters
			for anchorName, missedCounters := range targetMissing {
				if !missedCounters.Has("pages") {
					continue
				}
				// Adjust "pagesWanted"
				item := targetCollector.TargetLookupItems[anchorName]
				pageMakerIndex := item.PageMakerIndex
				if pageMakerIndex >= 0 && cachedAnchors.Has(anchorName) {
					pageMaker[pageMakerIndex].RemakeState.PagesWanted = true
				}
				// "contentChanged" is triggered in
				// targets.cacheTargetPageCounters()
			}

			if callParseAgain {
				remakeState.ContentChanged = true
				counterLookup.ParseAgain(pageCounterValues)
			}
		}
	}

	if pageType.Blank {
		resumeAt = previousResumeAt
	}

	if traceMode {
		traceLogger.DumpTree(page, "makePage done")
	}

	return page, resumeAt, tmp.nextPage
}

// Return one laid out page without margin boxes.
// Start with the initial values from “context.pageMaker[index]“.
// The resulting values / initial values for the next page are stored in
// the “pageMaker“.
// As the function"s name suggests: the plan is not to make all pages
// repeatedly when a missing counter was resolved, but rather re-make the
// single page where the “contentChanged“ happened.
func (context *layoutContext) remakePage(index int, rootBox bo.BlockLevelBoxITF, html *tree.HTML) (*bo.PageBox, tree.ResumeStack) {
	tmp := context.pageMaker[index]

	// PageType for current page, values for pageMaker[index + 1].
	// Don't modify actual pageMaker[index] values!
	pageState := tmp.InitialPageState.Copy()
	nextPageName := tmp.InitialNextPage.Page.String
	first := index == 0
	var nextPageSide string
	switch tmp.InitialNextPage.Break {
	case "left", "right":
		nextPageSide = tmp.InitialNextPage.Break
	case "recto", "verso":
		directionLtr := rootBox.Box().Style.GetDirection() == "ltr"
		breakVerso := tmp.InitialNextPage.Break == "verso"
		nextPageSide = "left"
		if directionLtr != breakVerso {
			nextPageSide = "right"
		}
	}
	blank := (nextPageSide == "left" && tmp.RightPage) || (nextPageSide == "right" && !tmp.RightPage) ||
		(len(context.reportedFootnotes) != 0 && tmp.InitialResumeAt == nil)

	if blank {
		nextPageName = ""
	}
	side := "left"
	if tmp.RightPage {
		side = "right"
	}
	pageType := utils.PageElement{Side: side, Blank: blank, First: first, Index: index, Name: nextPageName}
	context.styleFor.SetPageComputedStylesT(pageType, html)

	context.forcedBreak = tmp.InitialNextPage.Break != "any" || tmp.InitialNextPage.Page.String != ""
	context.marginClearance = false

	// makePage wants a pageNumber of index + 1
	pageNumber := index + 1
	page, resumeAt, nextPage := context.makePage(rootBox, pageType, tmp.InitialResumeAt,
		pageNumber, &pageState)

	if (nextPage == tree.PageBreak{}) {
		panic("expected nextPage")
	}
	if blank {
		nextPage.Page = tmp.InitialNextPage.Page
	}
	rightPage := !tmp.RightPage

	// Check whether we need to append or update the next pageMaker item
	var pageMakerNextChanged bool
	if index+1 >= len(context.pageMaker) {
		// New page
		pageMakerNextChanged = true
	} else {
		// Check whether something changed
		// TODO: Find what we need to compare. Is resumeAt enough?
		tmp := context.pageMaker[index+1]
		// (nextResumeAt, nextNextPage, nextRightPage,nextPageState, )
		pageMakerNextChanged = !tmp.InitialResumeAt.Equals(resumeAt) ||
			tmp.InitialNextPage != nextPage ||
			tmp.RightPage != rightPage ||
			!tmp.InitialPageState.Equal(pageState)
	}

	if pageMakerNextChanged {
		// Reset remakeState
		remakeState := tree.RemakeState{}
		// Setting contentChanged to true ensures remake.
		// If resumeAt  == nil  (last page) it must be false to prevent endless
		// loops && list index out of range (see #794).
		remakeState.ContentChanged = resumeAt != nil
		// pageState is already a deepcopy
		item := tree.PageMaker{
			InitialResumeAt: resumeAt, InitialNextPage: nextPage, RightPage: rightPage,
			InitialPageState: pageState, RemakeState: remakeState,
		}
		if index+1 >= len(context.pageMaker) {
			context.pageMaker = append(context.pageMaker, item)
		} else {
			context.pageMaker[index+1] = item
		}
	}

	return page, resumeAt
}

// Return a list of laid out pages without margin boxes.
// Re-make pages only if necessary.
func (context *layoutContext) makeAllPages(rootBox bo.BlockLevelBoxITF, html *tree.HTML, pages []*bo.PageBox) []*bo.PageBox {
	var (
		out               []*bo.PageBox
		reportedFootnotes []Box
	)
	i := 0
	for {
		remakeState := context.pageMaker[i].RemakeState
		var (
			resumeAt tree.ResumeStack
			page     *bo.PageBox
		)
		if len(pages) == 0 || remakeState.ContentChanged || remakeState.PagesWanted {
			logger.ProgressLogger.Printf("Step 5 - Creating layout - Page %d", i+1)
			// Reset remakeState
			context.pageMaker[i].RemakeState = tree.RemakeState{}
			page, resumeAt = context.remakePage(i, rootBox, html)
			reportedFootnotes = context.reportedFootnotes
			out = append(out, page)
		} else {
			logger.ProgressLogger.Printf("Step 5 - Creating layout - Page %d (up-to-date)", i+1)
			resumeAt = context.pageMaker[i+1].InitialResumeAt
			reportedFootnotes = nil
			out = append(out, pages[i])
		}

		i += 1
		if resumeAt == nil && len(reportedFootnotes) == 0 {
			// Throw away obsolete pages and content
			context.pageMaker = context.pageMaker[:i+1]
			for k := range context.brokenOutOfFlow {
				delete(context.brokenOutOfFlow, k)
			}
			context.reportedFootnotes = context.reportedFootnotes[:0]
			return out
		}
	}
}
