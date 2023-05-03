package layout

import (
	"fmt"

	"github.com/benoitkugler/webrender/html/tree"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
)

// Page breaking and layout for block-level and block-container boxes.

type blockLayout struct {
	resumeAt          tree.ResumeStack
	adjoiningMargins  []pr.Float
	nextPage          tree.PageBreak
	collapsingThrough bool
}

// Lay out the block-level “box“.
//
// `maxPositionY` is the absolute vertical position (as in
// “someBox.PositionY“) of the bottom of the
// content box of the current page area.
func blockLevelLayout(context *layoutContext, box_ bo.BlockLevelBoxITF, bottomSpace pr.Float, skipStack tree.ResumeStack,
	containingBlock *bo.BoxFields, pageIsEmpty bool, absoluteBoxes,
	fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins *[]pr.Float,
	discard bool, maxLines int,
) (bo.BlockLevelBoxITF, blockLayout, int) {
	box := box_.Box()
	if !bo.TableT.IsInstance(box_) {
		resolvePercentagesBox(box_, containingBlock, "")

		if box.MarginTop == pr.AutoF {
			box.MarginTop = pr.Float(0)
		}
		if box.MarginBottom == pr.AutoF {
			box.MarginBottom = pr.Float(0)
		}

		if context.currentPage > 1 && pageIsEmpty {
			// When an unforced break occurs before or after a block-level box,
			// any margins adjoining the break are truncated to zero.
			if collapseWithPage := containingBlock.IsForRootElement || len(*adjoiningMargins) != 0; collapseWithPage {
				if box.Style.GetMarginBreak() == "discard" {
					box.MarginTop = pr.Float(0)
				} else if box.Style.GetMarginBreak() == "auto" {
					if !context.forcedBreak {
						box.MarginTop = pr.Float(0)
					}
				}
			}
		}

		collapsedMargin := collapseMargin(append(*adjoiningMargins, box.MarginTop.V()))
		bl := box_.BlockLevel()
		bl.Clearance = getClearance(context, box, collapsedMargin)
		if bl.Clearance != nil {
			topBorderEdge := box.PositionY + collapsedMargin + bl.Clearance.V()
			box.PositionY = topBorderEdge - box.MarginTop.V()
			adjoiningMargins = new([]pr.Float)
		}
	}
	return blockLevelLayoutSwitch(context, box_, bottomSpace, skipStack, containingBlock,
		pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins, discard, maxLines)
}

// Call the layout function corresponding to the “box“ type.
func blockLevelLayoutSwitch(context *layoutContext, box_ bo.BlockLevelBoxITF, bottomSpace pr.Float, skipStack tree.ResumeStack,
	containingBlock *bo.BoxFields, pageIsEmpty bool, absoluteBoxes,
	fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins *[]pr.Float, discard bool,
	maxLines int,
) (bo.BlockLevelBoxITF, blockLayout, int) {
	if debugMode {
		debugLogger.LineWithIndent("Layout BLOCK-LEVEL <%s> (%s) (resume at : %s)", box_.Box().ElementTag(), box_.Type(), skipStack)
		defer debugLogger.LineWithDedent("")
	}

	if traceMode {
		traceLogger.Dump(fmt.Sprintf("skipStack %s", skipStack))
		traceLogger.DumpTree(box_, "blockLevelLayoutSwitch")
	}

	blockBox, isBlockBox := box_.(bo.BlockBoxITF)
	replacedBox, isReplacedBox := box_.(bo.ReplacedBoxITF)
	if table, ok := box_.(bo.TableBoxITF); ok {
		out1, out2 := tableLayout(context, table, bottomSpace, skipStack, pageIsEmpty, absoluteBoxes, fixedBoxes)
		return out1, out2, -1
	} else if isBlockBox {
		out1, out2 := blockBoxLayout(context, blockBox, bottomSpace, skipStack, containingBlock,
			pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins, discard, maxLines)
		return out1, out2, -1
	} else if isReplacedBox && bo.BlockReplacedT.IsInstance(box_) {
		b, v := blockReplacedBoxLayout(context, replacedBox, containingBlock)
		return b.(bo.BlockLevelBoxITF), v, -1 // blockReplacedBoxLayout is type stable
	} else if bo.FlexT.IsInstance(box_) {
		box_, layout := flexLayout(context, box_, bottomSpace, skipStack, containingBlock,
			pageIsEmpty, absoluteBoxes, fixedBoxes)
		return box_.(bo.BlockLevelBoxITF), layout, -1 // flexLayout is type stable
	} else { // pragma: no cover
		panic(fmt.Sprintf("Layout for %s not handled yet", box_))
	}
}

// Lay out the block “box“.
func blockBoxLayout(context *layoutContext, box_ bo.BlockBoxITF, bottomSpace pr.Float, skipStack tree.ResumeStack,
	containingBlock *bo.BoxFields, pageIsEmpty bool, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins *[]pr.Float,
	discard bool, maxLines int,
) (bo.BlockLevelBoxITF, blockLayout) {
	box := box_.Box()
	if box.Style.GetColumnWidth().String != "auto" || box.Style.GetColumnCount().String != "auto" {
		newBox_, result := columnsLayout(context, box_, bottomSpace, skipStack, containingBlock,
			pageIsEmpty, absoluteBoxes, fixedBoxes, *adjoiningMargins)
		resumeAt := result.resumeAt
		if resumeAt == nil {
			newBox := newBox_.Box()
			columnsBottomSpace := newBox.MarginBottom.V() + newBox.PaddingBottom.V() + newBox.BorderBottomWidth.V()
			if columnsBottomSpace != 0 {
				removePlaceholders(context, []Box{newBox_}, absoluteBoxes, fixedBoxes)
				bottomSpace += columnsBottomSpace
				newBox_, result = columnsLayout(context, box_, bottomSpace, skipStack,
					containingBlock, pageIsEmpty, absoluteBoxes, fixedBoxes, *adjoiningMargins)
			}
		}
		return newBox_, result
	} else if box.IsTableWrapper {
		tableWrapperWidth(context, box_, bo.MaybePoint{containingBlock.Width, containingBlock.Height})
	}
	blockLevelWidth(box_, nil, containingBlock)

	newBox__, result, maxLines := blockContainerLayout(context, box_, bottomSpace, skipStack, pageIsEmpty,
		absoluteBoxes, fixedBoxes, adjoiningMargins, discard, maxLines)
	newBox, _ := newBox__.(bo.BlockBoxITF) // blockContainerLayout is type stable
	if newBox != nil && newBox.Box().IsTableWrapper {
		// Don't collide with floats
		// https://www.w3.org/TR/CSS21/visuren.html#floats
		positionX, positionY, _ := avoidCollisions(context, newBox, containingBlock, false)
		newBox.Translate(newBox, positionX-newBox.Box().PositionX, positionY-newBox.Box().PositionY, false)
	}

	return newBox, result
}

var blockReplacedWidth = handleMinMaxWidth(blockReplacedWidth_)

// @handleMinMaxWidth
func blockReplacedWidth_(box Box, _ *layoutContext, containingBlock containingBlock) (bool, pr.Float) {
	// https://www.w3.org/TR/CSS21/visudet.html#block-replaced-width
	replacedBoxWidth_(box, nil, containingBlock)
	blockLevelWidth_(box, nil, containingBlock)
	return false, 0
}

var blockLevelWidth = handleMinMaxWidth(blockLevelWidth_)

// @handleMinMaxWidth
// Set the “box“ width.
// containingBlock must be bo.BoxFields
func blockLevelWidth_(box_ Box, _ *layoutContext, containingBlock_ containingBlock) (bool, pr.Float) {
	box := box_.Box()
	// "cb" stands for "containing block"
	var (
		cbWidth   pr.Float
		direction pr.String
	)
	switch cb := containingBlock_.(type) {
	case *bo.BoxFields:
		direction = cb.Style.GetDirection()
		cbWidth = cb.Width.V()
	case block:
		cbWidth = cb.Width
		direction = "ltr"
	}
	// https://www.w3.org/TR/CSS21/visudet.html#blockwidth

	// These names are waaay too long
	marginL := box.MarginLeft
	marginR := box.MarginRight
	width := box.Width
	paddingL := box.PaddingLeft.V()
	paddingR := box.PaddingRight.V()
	borderL := box.BorderLeftWidth.V()
	borderR := box.BorderRightWidth.V()

	// Only margin-left, margin-right and width can be "auto".
	// We want:  width of containing block ==
	//               margin-left + border-left-width + padding-left + width
	//               + padding-right + border-right-width + margin-right

	paddingsPlusBorders := paddingL + paddingR + borderL + borderR
	if width != pr.AutoF {
		total := paddingsPlusBorders + width.V()
		if marginL != pr.AutoF {
			total += marginL.V()
		}
		if marginR != pr.AutoF {
			total += marginR.V()
		}
		if total > cbWidth {
			if marginL == pr.AutoF {
				marginL = pr.Float(0)
				box.MarginLeft = pr.Float(0)
			}
			if marginR == pr.AutoF {
				marginR = pr.Float(0)
				box.MarginRight = pr.Float(0)
			}
		}
	}
	if width != pr.AutoF && marginL != pr.AutoF && marginR != pr.AutoF {
		// The equation is over-constrained.
		if direction == "rtl" && !box.IsColumn {
			box.PositionX += cbWidth - paddingsPlusBorders - width.V() - marginR.V() - marginL.V()
		} // Do nothing in ltr.
	}
	if width == pr.AutoF {
		if marginL == pr.AutoF {
			marginL = pr.Float(0)
			box.MarginLeft = pr.Float(0)
		}
		if marginR == pr.AutoF {
			marginR = pr.Float(0)
			box.MarginRight = pr.Float(0)
		}
		width = cbWidth - (paddingsPlusBorders + marginL.V() + marginR.V())
		box.Width = width
	}
	marginSum := cbWidth - paddingsPlusBorders - width.V()
	if marginL == pr.AutoF && marginR == pr.AutoF {
		box.MarginLeft = marginSum / 2.
		box.MarginRight = marginSum / 2.
	} else if marginL == pr.AutoF && marginR != pr.AutoF {
		box.MarginLeft = marginSum - marginR.V()
	} else if marginL != pr.AutoF && marginR == pr.AutoF {
		box.MarginRight = marginSum - marginL.V()
	}
	return false, 0
}

// Translate the “box“ if it is relatively positioned.
func relativePositioning(box_ Box, containingBlock bo.Point) {
	box := box_.Box()
	if box.Style.GetPosition().String == "relative" {
		resolvePositionPercentages(box, containingBlock)
		var translateX, translateY pr.Float
		if box.Left != pr.AutoF && box.Right != pr.AutoF {
			if box.Style.GetDirection() == "ltr" {
				translateX = box.Left.V()
			} else {
				translateX = -box.Right.V()
			}
		} else if box.Left != pr.AutoF {
			translateX = box.Left.V()
		} else if box.Right != pr.AutoF {
			translateX = -box.Right.V()
		} else {
			translateX = 0
		}

		if box.Top != pr.AutoF {
			translateY = box.Top.V()
		} else if box.Bottom != pr.AutoF {
			translateY = -box.Bottom.V()
		} else {
			translateY = 0
		}

		box_.Translate(box_, translateX, translateY, false)
	}
	if IsLine(box_) {
		for _, child := range box.Children {
			relativePositioning(child, containingBlock)
		}
	}
}

func reversedPl(f []*AbsolutePlaceholder) []*AbsolutePlaceholder {
	L := len(f)
	out := make([]*AbsolutePlaceholder, L)
	for i, v := range f {
		out[L-1-i] = v
	}
	return out
}

func reversedBoxes(in []Box) []Box {
	N := len(in)
	out := make([]Box, N)
	for i, v := range in {
		out[N-1-i] = v
	}
	return out
}

// reverse in place
func reverseB(a []Box) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}

type childrenBlockLevel interface {
	Box
	Children() []bo.BlockLevelBoxITF
}

// Set the “box“ height.
func blockContainerLayout(context *layoutContext, box_ Box, bottomSpace pr.Float, skipStack tree.ResumeStack,
	pageIsEmpty bool, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins *[]pr.Float, discard bool, maxLines int,
) (Box, blockLayout, int) {
	box := box_.Box()
	if !(bo.BlockContainerT.IsInstance(box_) || bo.FlexT.IsInstance(box_)) {
		panic(fmt.Sprintf("expected BlockContainer or Flex, got %T", box_))
	}

	// See https://www.w3.org/TR/CSS21/visuren.html#block-formatting
	if establishesFormattingContext(box_) {
		context.createBlockFormattingContext()
	}

	isStart := skipStack == nil
	box_.RemoveDecoration(box, !isStart, false)

	discard = discard || box.Style.GetContinue() == "discard"
	drawBottomDecoration := discard || box.Style.GetBoxDecorationBreak() == "clone"

	if drawBottomDecoration {
		bottomSpace += box.PaddingBottom.V() + box.BorderBottomWidth.V() + box.MarginBottom.V()
	}

	*adjoiningMargins = append(*adjoiningMargins, box.MarginTop.V())
	thisBoxAdjoiningMargins := adjoiningMargins

	collapsingWithChildren := !(pr.Is(box.BorderTopWidth) || pr.Is(box.PaddingTop) || box.IsFlexItem ||
		establishesFormattingContext(box_) || box.IsForRootElement)
	var positionY pr.Float
	if collapsingWithChildren {
		positionY = box.PositionY
	} else {
		box.PositionY += collapseMargin(*adjoiningMargins) - box.MarginTop.V()
		adjoiningMargins = new([]pr.Float)
		positionY = box.ContentBoxY()
	}

	positionX := box.ContentBoxX()

	if box.Style.GetPosition().String == "relative" {
		// New containing block, use a new absolute list
		absoluteBoxes = &[]*AbsolutePlaceholder{}
	}

	var (
		newChildren, allFootnotes []Box
		nextPage                  = tree.PageBreak{Break: "any"}
		resumeAt                  tree.ResumeStack
		brokenOutOfFlow           = make(map[Box]brokenBox)
		lastInFlowChild           Box
	)

	if ml := box.Style.GetMaxLines(); ml.Tag != pr.None {
		if maxLines <= 0 || ml.I <= maxLines {
			maxLines = ml.I
		}
	}

	skip := 0
	firstLetterStyle := box.FirstLetterStyle
	if !isStart {
		skip, skipStack = skipStack.Unpack()
		firstLetterStyle = nil
	}
	L := len(box.Children[skip:])
	var i int
	for i = 0; i < L; i++ {
		index := i + skip
		child_ := box.Children[skip:][i]
		child := child_.Box()

		if debugMode {
			debugLogger.LineWithIndent("Block container layout child %d: <%s> (%s), in normal flow: %v", i, child.ElementTag(), child_.Type(), child.IsInNormalFlow())
		}

		child.PositionX = positionX
		child.PositionY = positionY // does not count margins in adjoiningMargins
		var newFootnotes []Box

		var abort, stop bool
		if !child.IsInNormalFlow() {
			abort = false
			var (
				outOfFlowResumeAt tree.ResumeStack
				newChild          Box
			)
			stop, resumeAt, newChild, outOfFlowResumeAt = outOfFlowLayout(context, box, index, child_,
				&newChildren, pageIsEmpty, absoluteBoxes, fixedBoxes, *adjoiningMargins, bottomSpace)
			if outOfFlowResumeAt != nil {
				brokenOutOfFlow[newChild] = brokenBox{child_, box_, outOfFlowResumeAt}
			}
		} else if childLineBox, ok := child_.(*bo.LineBox); ok { // LineBox is a final type
			abort, stop, resumeAt, positionY, newChildren, newFootnotes, maxLines = lineBoxLayout(context, box_, index, childLineBox,
				newChildren, pageIsEmpty, absoluteBoxes, fixedBoxes, *adjoiningMargins,
				bottomSpace, positionY, skipStack, firstLetterStyle, maxLines)
			drawBottomDecoration = drawBottomDecoration || resumeAt == nil
			adjoiningMargins = new([]pr.Float)
			allFootnotes = append(allFootnotes, newFootnotes...)
		} else {
			var (
				adjoiningMarginsV []pr.Float
				newMaxLines       int
			)
			abort, stop, resumeAt, positionY, adjoiningMarginsV, nextPage, newChildren, newMaxLines = inFlowLayout(context, box, index, child_,
				newChildren, pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins,
				bottomSpace, positionY, skipStack, firstLetterStyle, collapsingWithChildren, discard, maxLines)
			skipStack = nil
			adjoiningMargins = &adjoiningMarginsV

			if newMaxLines != -1 && maxLines != -1 {
				maxLines = newMaxLines
				if maxLines <= 0 {
					stop = true
					lastChild := (child_ == box.Children[len(box.Children)-1])
					if !lastChild {
						children := newChildren
						for len(children) != 0 {
							lastChild := children[len(children)-1]
							if lastChildF, ok := lastChild.(*bo.LineBox); ok {
								lastChildF.BlockEllipsis = box.Style.GetBlockEllipsis()
							} else if bo.ParentT.IsInstance(lastChild) {
								children = lastChild.Box().Children
								continue
							}
							break
						}
					}
				}
			}

		}

		if debugMode {
			debugLogger.LineWithDedent("--> block child done (resume at: %s, positionY: %g)", resumeAt, positionY)
		}

		if traceMode {
			origin := "inFlow"
			if !child.IsInNormalFlow() {
				origin = "outOfFlow"
			} else if _, ok := child_.(*bo.LineBox); ok {
				origin = "lineBox"
			}
			traceLogger.Dump(fmt.Sprintf("Block container layout child %d (%s) resumeAt %s", i, origin, resumeAt))
		}

		if abort {
			page_, _ := child.PageValues()
			removePlaceholders(context, box.Children[skip:], absoluteBoxes, fixedBoxes)
			for _, footnote := range newFootnotes {
				context.unlayoutFootnote(footnote)
			}

			return nil, blockLayout{nextPage: tree.PageBreak{Break: "any", Page: page_}}, maxLines
		} else if stop {
			break
		}
	}

	if i == L {
		resumeAt = nil
	}

	boxIsFragmented := resumeAt != nil
	if box.Style.GetContinue() == "discard" {
		resumeAt = nil
	}

	if bi := string(box.Style.GetBreakInside()); boxIsFragmented && avoidPageBreak(bi, context) && !pageIsEmpty {
		for _, footnote := range allFootnotes {
			context.unlayoutFootnote(footnote)
		}

		return nil, blockLayout{nextPage: tree.PageBreak{Break: "any"}}, maxLines
	}

	for k, v := range brokenOutOfFlow {
		context.brokenOutOfFlow[k] = v
	}

	if collapsingWithChildren {
		box.PositionY += collapseMargin(*thisBoxAdjoiningMargins) - box.MarginTop.V()
	}

	lastInFlowChild = nil
	for _, previousChild := range reversedBoxes(newChildren) {
		if previousChild.Box().IsInNormalFlow() {
			lastInFlowChild = previousChild
			break
		}
	}
	collapsingThrough := false
	if lastInFlowChild == nil {
		collapsedMargin := collapseMargin(*adjoiningMargins)
		// top && bottom margin of this box
		if (box.Height == pr.AutoF || box.Height == pr.Float(0)) &&
			getClearance(context, box, collapsedMargin) == nil &&
			box.MinHeight == pr.Float(0) && box.BorderTopWidth == pr.Float(0) && box.PaddingTop == pr.Float(0) &&
			box.BorderBottomWidth == pr.Float(0) && box.PaddingBottom == pr.Float(0) {
			collapsingThrough = true
		} else {
			positionY += collapsedMargin
			adjoiningMargins = new([]pr.Float)
		}
	} else {
		// bottom margin of the last child && bottom margin of this box ...
		if box.Height != pr.AutoF {
			// not adjoining. (positionY is not used afterwards.)
			adjoiningMargins = new([]pr.Float)
		}
	}

	if pr.Is(box.BorderBottomWidth) || pr.Is(box.PaddingBottom) ||
		establishesFormattingContext(box_) || box.IsForRootElement || box.IsTableWrapper {
		positionY += collapseMargin(*adjoiningMargins)
		adjoiningMargins = new([]pr.Float)
	}

	newBox_ := bo.CopyWithChildren(box_, newChildren)
	newBox := newBox_.Box()
	newBox_.RemoveDecoration(newBox, !isStart, boxIsFragmented && !discard)

	if newBox.Height == pr.AutoF {
		if len(*context.excludedShapes) != 0 && newBox.Style.GetOverflow() != "visible" {
			maxFloatPositionY := -pr.Inf
			for _, floatBox := range *context.excludedShapes {
				v := floatBox.PositionY + floatBox.MarginHeight()
				if v > maxFloatPositionY {
					maxFloatPositionY = v
				}
			}
			positionY = pr.Max(maxFloatPositionY, positionY)
		}
		newBox.Height = positionY - newBox.ContentBoxY()
	}

	if newBox.Style.GetPosition().String == "relative" {
		// New containing block, resolve the layout of the absolute descendants
		for _, absoluteBox := range *absoluteBoxes {
			absoluteLayout(context, absoluteBox, newBox_, fixedBoxes, bottomSpace, nil)
		}
	}

	for _, child := range newBox.Children {
		relativePositioning(child, bo.Point{newBox.Width.V(), newBox.Height.V()})
	}

	if establishesFormattingContext(newBox_) {
		context.finishBlockFormattingContext(newBox_)
	}

	if discard || !boxIsFragmented {
		// After finishBlockFormattingContext which may increment
		// newBox.Height
		newBox.Height = pr.Max(pr.Min(newBox.Height.V(), newBox.MaxHeight.V()), newBox.MinHeight.V())
	} else if bottomSpace > -pr.Inf && !newBox.IsColumn {
		// Make the box fill the blank space at the bottom of the page
		// https://www.w3.org/TR/css-break-3/#box-splitting
		newBoxHeight := context.pageBottom - bottomSpace - newBox.PositionY - (newBox.MarginHeight() - newBox.Height.V())
		if newBoxHeight > newBox.Height.V() {
			newBox.Height = newBoxHeight
			if drawBottomDecoration {
				newBox.Height = newBox.Height.V() + box.PaddingBottom.V() + box.BorderBottomWidth.V() + box.MarginBottom.V()
			}
		}
	}

	if nextPage.Page.IsNone() {
		_, nextPage.Page = newBox.PageValues()
	}

	return newBox_, blockLayout{
		resumeAt: resumeAt, nextPage: nextPage,
		adjoiningMargins: *adjoiningMargins, collapsingThrough: collapsingThrough,
	}, maxLines
}

func findLastInFlowChild(children []Box) Box {
	for _, previousChild := range reversedBoxes(children) {
		if previousChild.Box().IsInNormalFlow() {
			return previousChild
		}
	}
	return nil
}

func outOfFlowLayout(context *layoutContext, box *bo.BoxFields, index int, child_ Box, newChildren *[]Box,
	pageIsEmpty bool, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins []pr.Float, bottomSpace pr.Float,
) (stop bool, resumeAt tree.ResumeStack, newChild Box, outOfFlowLayoutResumeAt tree.ResumeStack) {
	child := child_.Box()
	child.PositionY += collapseMargin(adjoiningMargins)

	if child.IsAbsolutelyPositioned() {
		placeholder := NewAbsolutePlaceholder(child_)
		newChild = placeholder
		placeholder.Box().Index = index
		*newChildren = append(*newChildren, placeholder)
		if child.Style.GetPosition().String == "absolute" {
			*absoluteBoxes = append(*absoluteBoxes, placeholder)
		} else {
			*fixedBoxes = append(*fixedBoxes, placeholder)
		}
	} else if child.IsFloated() {
		newChild, outOfFlowLayoutResumeAt = floatLayout(context, child_, box, absoluteBoxes, fixedBoxes, bottomSpace, nil)
		newChild_ := newChild.Box()
		// New page if overflow
		pageOverflow := context.overflowsPage(bottomSpace, newChild_.PositionY+newChild_.Height.V())
		if (pageIsEmpty && len(*newChildren) == 0) || !pageOverflow {
			asPlaceholder := AbsolutePlaceholder{AliasBox: newChild}
			asPlaceholder.Box().Index = index
			*newChildren = append(*newChildren, &asPlaceholder)
		} else {
			lastInFlowChild := findLastInFlowChild(*newChildren)
			pageBreak := blockLevelPageBreak(lastInFlowChild, child_)
			resumeAt = tree.ResumeStack{index: nil}
			if len(*newChildren) != 0 && avoidPageBreak(pageBreak, context) {
				r1, r2 := findEarlierPageBreak(context, *newChildren, absoluteBoxes, fixedBoxes)
				if r1 != nil || r2 != nil {
					_, resumeAt = r1, r2
				}
			}
			stop = true
		}
	} else if child.IsRunning() {
		context.addRunning(child_)
	}
	return stop, resumeAt, newChild, outOfFlowLayoutResumeAt
}

func breakLine(context *layoutContext, box *bo.BoxFields, line *bo.LineBox, newChildren *[]Box, iter *lineBoxeIterator,
	pageIsEmpty bool, index int, skipStack, resumeAt tree.ResumeStack,
	absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder,
) (bool, bool, tree.ResumeStack) {
	overOrphans := len(*newChildren) - int(box.Style.GetOrphans())
	if overOrphans < 0 && !pageIsEmpty {
		// Reached the bottom of the page before we had
		// enough lines for orphans, cancel the whole box.
		return true, false, resumeAt
	}
	// How many lines we need on the next page to satisfy widows
	// -1 for the current line.
	needed := int(box.Style.GetWidows() - 1)
	if needed != 0 {
		for iter.Has() {
			iter.Next()
			needed -= 1
			if needed == 0 {
				break
			}
		}
	}
	if needed > overOrphans && !pageIsEmpty {
		// Total number of lines < orphans + widows
		return true, false, resumeAt
	}
	if needed != 0 && needed <= overOrphans {
		// Remove lines to keep them for the next page
		cut := len(*newChildren) - needed
		for _, child := range (*newChildren)[cut:] {
			removePlaceholders(context, child.Box().Children, absoluteBoxes, fixedBoxes)
		}
		(*newChildren) = (*newChildren)[:cut]
	}
	// Page break here, resume before this line
	removePlaceholders(context, line.Children, absoluteBoxes, fixedBoxes)
	return false, true, tree.ResumeStack{index: skipStack}
}

// newChildren and the returned newChildren are actually *LineBox
func lineBoxLayout(context *layoutContext, box_ Box, index int, child_ *bo.LineBox, newChildren []Box,
	pageIsEmpty bool, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins []pr.Float, bottomSpace, positionY pr.Float,
	skipStack tree.ResumeStack, firstLetterStyle pr.ElementStyle, maxLines int) (
	abort, stop bool, resumeAt tree.ResumeStack, _ pr.Float, _ []Box, newFootnotes []Box, _ int,
) {
	box := box_.Box()
	if len(box.Children) != 1 {
		panic("line box with siblings before layout")
	}
	if len(adjoiningMargins) != 0 {
		positionY += collapseMargin(adjoiningMargins)
		adjoiningMargins = nil
	}
	linesIterator := iterLineBoxes(context, child_, positionY, bottomSpace, skipStack,
		box_, absoluteBoxes, fixedBoxes, firstLetterStyle)
	for i := 0; linesIterator.Has(); i++ {
		tmp := linesIterator.Next()
		line_ := tmp.line
		resumeAt = tmp.resumeAt
		line := line_.Box()

		// Break box if we reached max-lines
		if maxLines != -1 {
			if maxLines == 0 {
				newChildren[len(newChildren)-1].(*bo.LineBox).BlockEllipsis = box.Style.GetBlockEllipsis()
				break
			}
			maxLines -= 1
		}
		// Update line resume_at and position_y
		line_.ResumeAt = resumeAt
		newPositionY := line.PositionY + line.Height.V()
		// Add bottom padding and border to the bottom position of the box if needed
		var offsetY pr.Float
		if resumeAt == nil || box.Style.GetBoxDecorationBreak() == "clone" {
			offsetY = box.BorderBottomWidth.V() + box.PaddingBottom.V()
		}

		// Allow overflow if the first line of the page is higher
		// than the page itself so that we put *something* on this
		// page and can advance in the context.
		overflow := (len(newChildren) != 0 || !pageIsEmpty) && context.overflowsPage(bottomSpace, newPositionY+offsetY)
		if overflow {
			abort, stop, resumeAt = breakLine(context, box, line_, &newChildren, linesIterator,
				pageIsEmpty, index, skipStack, resumeAt, absoluteBoxes, fixedBoxes)
			break
			// See https://drafts.csswg.org/css-page-3/#allowed-pg-brk
			// "When an unforced page break occurs here, both the adjoining
			//  ‘margin-top’ and ‘margin-bottom’ are set to zero."
			// See https://github.com/Kozea/WeasyPrint/issues/115
		} else if pageIsEmpty && context.overflowsPage(bottomSpace, newPositionY) {
			// Remove the top border when a page is empty && the box is
			// too high to be drawn := range one page
			newPositionY -= box.MarginTop.V()
			line_.Translate(line_, 0, -box.MarginTop.V(), false)
			box.MarginTop = pr.Float(0)
		}

		if len(context.footnotes) != 0 {
			breakLinebox := false
			var footnotes []Box
			for _, descendant := range bo.Descendants(line_) {
				if ftn := descendant.Box().Footnote; isInBoxes(ftn, context.footnotes) {
					footnotes = append(footnotes, ftn)
				}
			}
			for _, footnote := range footnotes {
				overflow := context.layoutFootnote(footnote)
				newFootnotes = append(newFootnotes, footnote)
				overflow = overflow || len(context.reportedFootnotes) != 0 || context.overflowsPage(bottomSpace, newPositionY+offsetY)
				if overflow {
					context.reportFootnote(footnote)
					// If we've put other content on this page, then we may want
					// to push this line or block to the next page. Otherwise,
					// we can't (and would loop forever if we tried), so don't
					// even try.
					if len(newChildren) != 0 || !pageIsEmpty {
						if footnote.Box().Style.GetFootnotePolicy() == "line" {
							abort, stop, resumeAt = breakLine(
								context, box, line_, &newChildren, linesIterator, pageIsEmpty,
								index, skipStack, resumeAt, absoluteBoxes, fixedBoxes)
							breakLinebox = true
						} else if footnote.Box().Style.GetFootnotePolicy() == "block" {
							abort, breakLinebox = true, true
						}
						break
					}
				}
			}
			if breakLinebox {
				break
			}
		}

		newChildren = append(newChildren, line_)
		positionY = newPositionY
		skipStack = resumeAt

		if traceMode {
			traceLogger.Dump(fmt.Sprintf("lineBoxLayout at line %d -> %s", i, skipStack))
		}
	}

	if len(newChildren) != 0 {
		resumeAt = tree.ResumeStack{index: newChildren[len(newChildren)-1].(*bo.LineBox).ResumeAt}
	}

	return abort, stop, resumeAt, positionY, newChildren, newFootnotes, maxLines
}

func inFlowLayout(context *layoutContext, box *bo.BoxFields, index int, child_ Box, newChildren []Box,
	pageIsEmpty bool, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins *[]pr.Float, bottomSpace, positionY pr.Float,
	skipStack tree.ResumeStack, firstLetterStyle pr.ElementStyle, collapsingWithChildren, discard bool,
	maxLines int) (
	abort, stop bool, resumeAt tree.ResumeStack, _ pr.Float, _ []pr.Float, nextPage tree.PageBreak, _ []Box, _ int,
) {
	lastInFlowChild := findLastInFlowChild(newChildren)
	child := child_.Box()
	pageBreak := "auto"
	if lastInFlowChild != nil {
		// Between in-flow siblings
		pageBreak = blockLevelPageBreak(lastInFlowChild, child_)
		pageName_ := blockLevelPageName(lastInFlowChild, child_)
		if (pageName_.String != "" || pageName_.Page != 0) || forcePageBreak(pageBreak, context) {
			pageName, _ := child.PageValues()
			nextPage = tree.PageBreak{Break: pageBreak, Page: pageName}
			resumeAt = tree.ResumeStack{index: nil}
			stop = true
			return abort, stop, resumeAt, positionY, *adjoiningMargins, nextPage, newChildren, maxLines
		}
	}

	newContainingBlock := box

	if !newContainingBlock.IsTableWrapper {
		resolvePercentagesBox(child_, newContainingBlock, "")
		if lastInFlowChild == nil && collapsingWithChildren {
			oldCollapsedMargin := collapseMargin(*adjoiningMargins)
			childMarginTop := child.MarginTop
			if childMarginTop == pr.AutoF {
				childMarginTop = pr.Float(0)
			} else if context.currentPage > 1 && pageIsEmpty {
				if mb := box.Style.GetMarginBreak(); mb == "discard" || (mb == "auto" && !context.forcedBreak) {
					childMarginTop = pr.Float(0)
				}
			}
			newCollapsedMargin := collapseMargin(append(*adjoiningMargins, childMarginTop.V()))
			collapsedMarginDifference := newCollapsedMargin - oldCollapsedMargin
			for _, previousNewChild := range newChildren {
				previousNewChild.Translate(previousNewChild, 0, collapsedMarginDifference, false)
			}

			if clearance := getClearance(context, child, newCollapsedMargin); clearance != nil {
				for _, previousNewChild := range newChildren {
					previousNewChild.Translate(previousNewChild, 0, -collapsedMarginDifference, false)
				}

				collapsedMargin := collapseMargin(*adjoiningMargins)
				box.PositionY += collapsedMargin - box.MarginTop.V()
				// Count box.MarginTop as we emptied adjoiningMargins
				adjoiningMargins = new([]pr.Float)
				positionY = box.ContentBoxY()
			}
		}
	}
	if len(*adjoiningMargins) != 0 && box.IsTableWrapper {
		collapsedMargin := collapseMargin(*adjoiningMargins)
		child.PositionY += collapsedMargin
		positionY += collapsedMargin
		adjoiningMargins = new([]pr.Float)
	}

	atLeastOneNotPlaceholder := false
	for _, child := range newChildren {
		if _, isAbsPlac := child.(*AbsolutePlaceholder); !isAbsPlac {
			atLeastOneNotPlaceholder = true
			break
		}
	}
	pageIsEmptyWithNoChildren := pageIsEmpty && !atLeastOneNotPlaceholder

	if child.FirstLetterStyle == nil {
		child.FirstLetterStyle = firstLetterStyle
	}

	newChild_, tmp, maxLines := blockLevelLayout(context, child_.(bo.BlockLevelBoxITF), bottomSpace, skipStack,
		newContainingBlock, pageIsEmptyWithNoChildren, absoluteBoxes, fixedBoxes, adjoiningMargins, discard, maxLines)
	resumeAt, nextPage = tmp.resumeAt, tmp.nextPage
	nextAdjoiningMargins, collapsingThrough := tmp.adjoiningMargins, tmp.collapsingThrough

	if traceMode {
		traceLogger.Dump(fmt.Sprintf("in inFlowLayout: blockLevelLayout -> %s", resumeAt))
	}

	if newChild_ != nil {
		newChild := newChild_.Box()

		// We need to do this after the child layout to have the
		// used value for marginTop (eg. it might be a percentage.)
		if !(bo.BlockT.IsInstance(newChild_) || bo.TableT.IsInstance(newChild_)) {
			*adjoiningMargins = append(*adjoiningMargins, newChild.MarginTop.V())
			offsetY := collapseMargin(*adjoiningMargins) - newChild.MarginTop.V()
			newChild_.Translate(newChild_, 0, offsetY, false)
		}
		// else: blocks handle that themselves.

		if !collapsingThrough {
			newContentPositionY := newChild.ContentBoxY() + newChild.Height.V()
			newPositionY := newChild.BorderBoxY() + newChild.BorderHeight()
			pageOverflow := context.overflowsPage(bottomSpace, newContentPositionY)
			if pageOverflow && !pageIsEmptyWithNoChildren {
				// The child content overflows the page area, display it on the
				// next page.
				removePlaceholders(context, []Box{newChild_}, absoluteBoxes, fixedBoxes)
				newChild_ = nil
			} else if context.overflowsPage(bottomSpace, newPositionY) && !pageIsEmptyWithNoChildren {
				// The child border/padding overflows the page area, do the
				// layout again with a higher bottomSpace value.
				removePlaceholders(context, []Box{newChild_}, absoluteBoxes, fixedBoxes)
				bottomSpace += newChild.PaddingBottom.V() + newChild.BorderBottomWidth.V()

				newChild_, tmp, maxLines = blockLevelLayout(context, child_.(bo.BlockLevelBoxITF), bottomSpace, skipStack,
					newContainingBlock, pageIsEmptyWithNoChildren, absoluteBoxes, fixedBoxes, adjoiningMargins, discard, maxLines)
				resumeAt, nextPage = tmp.resumeAt, tmp.nextPage
				nextAdjoiningMargins, collapsingThrough = tmp.adjoiningMargins, tmp.collapsingThrough
				if newChild_ != nil {
					newChild = newChild_.Box()
					positionY = (newChild.BorderBoxY() + newChild.BorderHeight())
				}
			} else {
				positionY = newPositionY
			}
		}

		adjoiningMargins = &nextAdjoiningMargins
		if newChild_ != nil {
			*adjoiningMargins = append(*adjoiningMargins, newChild.MarginBottom.V())
		}

		if newChild_ != nil && newChild_.BlockLevel().Clearance != nil {
			positionY = newChild.BorderBoxY() + newChild.BorderHeight()
		}
	}

	skipStack = nil

	if newChild_ == nil {
		// Nothing fits in the remaining space of this page: break
		if avoidPageBreak(pageBreak, context) {
			r1, r2 := findEarlierPageBreak(context, newChildren, absoluteBoxes, fixedBoxes)
			if r1 != nil || r2 != nil {
				newChildren, resumeAt = r1, r2
				stop = true
				return abort, stop, resumeAt, positionY, *adjoiningMargins, nextPage, newChildren, maxLines
			} else {
				// We did not find any page break opportunity
				if !pageIsEmpty {
					// The page has content *before* this block:
					// cancel the block and try to find a break
					// in the parent.
					abort = true
					return abort, stop, resumeAt, positionY, *adjoiningMargins, nextPage, newChildren, maxLines
				}
				// else : ignore this "avoid" and break anyway.
			}
		}
		allAbsPos := true
		for _, child := range newChildren {
			if !child.Box().IsAbsolutelyPositioned() {
				allAbsPos = false
				break
			}
		}
		if allAbsPos {
			// This box has only rendered absolute children, keep them
			// for the next page. This is for example useful for list
			// markers.
			removePlaceholders(context, newChildren, absoluteBoxes, fixedBoxes)
			newChildren = nil
		}
		if len(newChildren) != 0 {
			resumeAt = tree.ResumeStack{index: nil}
			stop = true
		} else {
			// This was the first child of this box, cancel the box
			// completly
			abort = true
		}
		return abort, stop, resumeAt, positionY, *adjoiningMargins, nextPage, newChildren, maxLines
	}

	// index in its non-laid-out parent, not in future new parent
	// May be used in findEarlierPageBreak()
	newChild_.Box().Index = index
	newChildren = append(newChildren, newChild_)
	if resumeAt != nil {
		resumeAt = tree.ResumeStack{index: resumeAt}
		stop = true
	}

	return abort, stop, resumeAt, positionY, *adjoiningMargins, nextPage, newChildren, maxLines
}

// Return the amount of collapsed margin for a list of adjoining margins.
func collapseMargin(adjoiningMargins []pr.Float) pr.Float {
	var maxPos, minNeg pr.Float
	for _, m := range adjoiningMargins {
		if m > maxPos {
			maxPos = m
		} else if m < minNeg {
			minNeg = m
		}
	}
	return maxPos + minNeg
}

// Return wether a box establishes a block formatting context.
// See https://www.w3.org/TR/CSS2/visuren.html#block-formatting
func establishesFormattingContext(box_ Box) bool {
	box := box_.Box()
	return box.IsFloated() ||
		box.IsAbsolutelyPositioned() ||
		box.IsColumn ||
		(bo.BlockContainerT.IsInstance(box_) && !bo.BlockT.IsInstance(box_)) ||
		(bo.BlockT.IsInstance(box_) && box.Style.GetOverflow() != "visible") ||
		box.Style.GetDisplay().Has("flow-root")
}

// https://drafts.csswg.org/css-break-3/#possible-breaks
func isParallel(box Box) bool {
	return bo.BlockLevelT.IsInstance(box) || bo.TableRowGroupT.IsInstance(box) || bo.TableRowT.IsInstance(box)
}

func reverseStrings(f []pr.String) {
	for left, right := 0, len(f)-1; left < right; left, right = left+1, right-1 {
		f[left], f[right] = f[right], f[left]
	}
}

// Return the value of “page-break-before“ or “page-break-after“
// that "wins" for boxes that meet at the margin between two sibling boxes.
// For boxes before the margin, the "page-break-after" value is considered;
// for boxes after the margin the "page-break-before" value is considered.
// * "avoid" takes priority over "auto"
// * "page" takes priority over "avoid" or "auto"
// * "left" or "right" take priority over "always", "avoid" or "auto"
// * Among "left" && "right", later values in the tree take priority.
//
// See https://drafts.csswg.org/css-page-3/#allowed-pg-brk
func blockLevelPageBreak(siblingBefore, siblingAfter Box) string {
	var values []pr.String

	box_ := siblingBefore
	for isParallel(box_) {
		box := box_.Box()
		values = append(values, box.Style.GetBreakAfter())
		if !(bo.ParentT.IsInstance(box_) && len(box.Children) != 0) {
			break
		}
		box_ = box.Children[len(box.Children)-1]
	}
	reverseStrings(values) // Have them in tree order

	box_ = siblingAfter
	for isParallel(box_) {
		box := box_.Box()
		values = append(values, box.Style.GetBreakBefore())
		if !(bo.ParentT.IsInstance(box_) && len(box.Children) != 0) {
			break
		}
		box_ = box.Children[0]
	}
	choices := map[[2]pr.String]bool{
		{"page", "auto"}:           true,
		{"page", "avoid"}:          true,
		{"page", "avoid-page"}:     true,
		{"page", "avoid-column"}:   true,
		{"column", "auto"}:         true,
		{"column", "avoid"}:        true,
		{"column", "avoid-page"}:   true,
		{"column", "avoid-column"}: true,
		{"avoid", "auto"}:          true,
		{"avoid-page", "auto"}:     true,
		{"avoid-column", "auto"}:   true,
	}
	var result pr.String = "auto"
	for _, value := range values {
		tmp := [2]pr.String{value, result}
		if value == "left" || value == "right" || value == "recto" || value == "verso" || choices[tmp] {
			result = value
		}

	}

	return string(result)
}

// Return the next page name when siblings don't have the same names,
// or the zero value.
func blockLevelPageName(siblingBefore, siblingAfter Box) pr.Page {
	_, beforePage := siblingBefore.PageValues()
	afterPage, _ := siblingAfter.PageValues()
	if beforePage != afterPage {
		return afterPage
	}
	return pr.Page{}
}

// Find the last possible page break in “children“
// Because of a `page-break-before: avoid` or a `page-break-after: avoid`
// we need to find an earlier page break opportunity inside `children`.
// Absolute or fixed placeholders removed from children should also be
// removed from `absoluteBoxes` or `fixedBoxes`.
func findEarlierPageBreak(context *layoutContext, children []Box, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder) (newChildren []Box, resumeAt tree.ResumeStack) {
	if len(children) != 0 && bo.LineT.IsInstance(children[0]) {
		// Normally `orphans` and `widows` apply to the block container, but
		// line boxes inherit them.
		orphans := int(children[0].Box().Style.GetOrphans())
		widows := int(children[0].Box().Style.GetWidows())
		index := len(children) - widows // how many lines we keep
		if index < orphans {
			return nil, nil
		}
		newChildren = children[:index]
		resumeAt = tree.ResumeStack{0: newChildren[len(newChildren)-1].(*bo.LineBox).ResumeAt}
		removePlaceholders(context, children[index:], absoluteBoxes, fixedBoxes)
		return newChildren, resumeAt
	}

	var (
		previousInFlow Box
		index          int
		i_, L          = 0, len(children)
	)
	for i_ = 0; i_ < L; i_++ { // reversed(list(enumerate(children)))
		index = L - i_ - 1
		child_ := children[index]
		child := child_.Box()

		if bo.TableRowGroupT.IsInstance(child_) && (child.IsHeader || child.IsFooter) {
			// We don’t want to break pages before table headers or footers.
			continue
		} else if child.IsColumn {
			// We don’t want to break pages between columns.
			continue
		}

		if child.IsInNormalFlow() {
			pageBreak := blockLevelPageBreak(child_, previousInFlow)
			if previousInFlow != nil && !avoidPageBreak(pageBreak, context) {
				index += 1 // break after child
				newChildren = children[:index]
				// Get the index in the original parent
				resumeAt = tree.ResumeStack{children[index].Box().Index: nil}
				break
			}
			previousInFlow = child_
		}
		if bi := string(child.Style.GetBreakInside()); child.IsInNormalFlow() && !avoidPageBreak(bi, context) {
			if bo.BlockT.IsInstance(child_) || bo.TableT.IsInstance(child_) || bo.TableRowGroupT.IsInstance(child_) {
				newGrandChildren, resumeAtTmp := findEarlierPageBreak(context, child.Children, absoluteBoxes, fixedBoxes)
				if newGrandChildren != nil || resumeAtTmp != nil {
					resumeAt = resumeAtTmp
					newChild := bo.CopyWithChildren(child_, newGrandChildren)
					newChildren = append(children[:index], newChild)

					// Re-add footer at the end of split table
					if bo.TableRowGroupT.IsInstance(child_) {
						for _, nextChild := range children[index:] {
							if nextChild.Box().IsFooter {
								newChildren = append(newChildren, nextChild)
							}
						}
					}

					// Index in the original parent
					resumeAt = tree.ResumeStack{newChild.Box().Index: resumeAt}
					index += 1 // Remove placeholders after child
					break
				}
			}
		}
	}
	if i_ == L {
		return nil, nil
	}

	removePlaceholders(context, children[index:], absoluteBoxes, fixedBoxes)
	return newChildren, resumeAt
}

func removeFromAbsolutes(list *[]*AbsolutePlaceholder, box Box) {
	b := (*list)[:0]
	for _, x := range *list {
		if x != box {
			b = append(b, x)
		}
	}
	*list = b
}

// update boxes in place
func removeFromBoxes(list *[]Box, box Box) {
	b := (*list)[:0]
	for _, x := range *list {
		if x != box {
			b = append(b, x)
		}
	}
	*list = b
}

// For boxes that have been removed in findEarlierPageBreak(),
// also remove the matching placeholders in absoluteBoxes and fixedBoxes.
//
// Also takes care of removed footnotes and floats.
func removePlaceholders(context *layoutContext, boxList []Box, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder) {
	for _, box_ := range boxList {
		box := box_.Box()
		if bo.ParentT.IsInstance(box_) {
			removePlaceholders(context, box.Children, absoluteBoxes, fixedBoxes)
		}
		if box.Style.GetPosition().String == "absolute" {
			removeFromAbsolutes(absoluteBoxes, box_)
		} else if box.Style.GetPosition().String == "fixed" {
			removeFromAbsolutes(fixedBoxes, box_)
		}

		if box.Footnote != nil {
			context.unlayoutFootnote(box.Footnote)
		}
		delete(context.brokenOutOfFlow, box_)
	}
}

// Test whether we should avoid breaks.
func avoidPageBreak(pageBreak string, context *layoutContext) bool {
	if context.inColumn {
		return pageBreak == "avoid" || pageBreak == "avoid-page" || pageBreak == "avoid-column"
	}
	return pageBreak == "avoid" || pageBreak == "avoid-page"
}

// Test whether we should force breaks.
func forcePageBreak(pageBreak string, context *layoutContext) bool {
	if context.inColumn {
		return pageBreak == "page" || pageBreak == "left" || pageBreak == "right" || pageBreak == "recto" || pageBreak == "verso" || pageBreak == "column"
	}
	return pageBreak == "page" || pageBreak == "left" || pageBreak == "right" || pageBreak == "recto" || pageBreak == "verso"
}
