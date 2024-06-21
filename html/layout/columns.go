package layout

import (
	"fmt"

	"github.com/benoitkugler/webrender/html/tree"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
)

// Layout for columns.

func isInFloats(v pr.Float, l []pr.Float) bool {
	for _, vl := range l {
		if vl == v {
			return true
		}
	}
	return false
}

// if box is nil, then represents a list
type boxOrList struct {
	box  bo.BlockLevelBoxITF
	list []Box
}

// Lay out a multi-column “box“.
func columnsLayout(context *layoutContext, box_ bo.BlockBoxITF, bottomSpace pr.Float, skipStack tree.ResumeStack, containingBlock *bo.BoxFields,
	pageIsEmpty bool, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins []pr.Float,
) (bo.BlockLevelBoxITF, blockLayout) {
	style := box_.Box().Style
	width_ := style.GetColumnWidth()
	count_ := style.GetColumnCount()
	gap := style.GetColumnGap().Value
	height_ := style.GetHeight()
	originalBottomSpace := bottomSpace

	context.inColumn = true

	if style.GetPosition().String == "relative" {
		// New containing block, use a new absolute list
		absoluteBoxes = &[]*AbsolutePlaceholder{}
	}

	box_ = bo.CopyWithChildren(box_, box_.Box().Children).(bo.BlockBoxITF) // CopyWithChildren preserves the concrete type of box_
	box := box_.Box()
	box.PositionY += collapseMargin(adjoiningMargins) - box.MarginTop.V()

	// Set height if defined
	heightDefined := false
	if height_.S != "auto" && height_.Unit != pr.Perc {
		if height_.Unit != pr.Px {
			panic(fmt.Sprintf("expected Px got %v", height_))
		}
		heightDefined = true
		emptySpace := context.pageBottom - box.ContentBoxY() - height_.Value
		bottomSpace = pr.Max(bottomSpace, emptySpace)
	}

	blockLevelWidth(box_, nil, containingBlock)
	// Define the number of columns and their widths

	availableWidth := box.Width.V()
	var (
		width pr.Float
		count int
	)
	if width_.S == "auto" && count_.String != "auto" {
		count = count_.Int
		width = pr.Max(0, availableWidth-(pr.Float(count)-1)*gap) / pr.Float(count)
	} else if width_.S != "auto" && count_.String == "auto" {
		count = int(pr.Max(1, pr.Floor((availableWidth+gap)/(width_.Value+gap))))
		width = (availableWidth+gap)/pr.Float(count) - gap
	} else { // overconstrained, with width != 'auto' and count != 'auto'
		count = int(pr.Min(pr.Float(count_.Int), pr.Floor((availableWidth+gap)/(width_.Value+gap))))
		width = (availableWidth+gap)/pr.Float(count) - gap
	}

	// Handle column-span property with the following structure:
	// columnsAndBlocks = [
	//     [columnChild1, columnChild2],
	//     spanningBlock,
	//     …
	// ]
	type ibl struct {
		index int
		bl    boxOrList
	}
	var (
		columnsAndBlocks []ibl
		columnChildren   []Box
	)
	skip := 0
	if skipStack != nil {
		skip, _ = skipStack.Unpack()
	}
	for i_, child := range box.Children[skip:] {
		index := i_ + skip
		if child.Box().Style.GetColumnSpan() == "all" {
			if len(columnChildren) != 0 {
				columnsAndBlocks = append(columnsAndBlocks, ibl{index - len(columnChildren), boxOrList{list: columnChildren}})
			}
			columnsAndBlocks = append(columnsAndBlocks, ibl{index, boxOrList{box: child.Copy().(bo.BlockLevelBoxITF)}})
			columnChildren = nil
			continue
		}
		columnChildren = append(columnChildren, child.Copy())
	}
	if len(columnChildren) != 0 {
		columnsAndBlocks = append(columnsAndBlocks, ibl{len(box.Children) - len(columnChildren), boxOrList{list: columnChildren}})
	}

	if skipStack != nil {
		skipStack = tree.ResumeStack{0: skipStack[skip]}
	}

	var nextPage tree.PageBreak
	if len(box.Children) == 0 {
		nextPage = tree.PageBreak{Break: "any"}
		skipStack = nil
	}

	// Find height and balance.
	//
	// The current algorithm starts from the total available height, to check
	// whether the whole content can fit. If it doesn’t fit, we keep the partial
	// rendering. If it fits, we try to balance the columns starting from the
	// ideal height (the total height divided by the number of columns). We then
	// iterate until the last column is not the highest one. At the end of each
	// loop, we add the minimal height needed to make one direct child at the
	// top of one column go to the end of the previous column.
	//
	// We rely on a real rendering for each loop, and with a stupid algorithm
	// like this it can last minutes…

	adjoiningMargins = nil
	var (
		currentPositionY    = box.ContentBoxY()
		newChildren         []Box
		columnSkipStack     tree.ResumeStack
		lastLoop            = false
		breakPage           = false
		lastFootnotesHeight pr.Float
		footnoteAreaHeights = []pr.Float{0}
		index               int
	)
	if h := context.currentFootnoteArea.Height; h != pr.AutoF {
		footnoteAreaHeights = []pr.Float{context.currentFootnoteArea.MarginHeight()}
	}

	for _, pair := range columnsAndBlocks {
		index = pair.index
		columnChildrenOrBlock := pair.bl
		if block := columnChildrenOrBlock.box; block != nil {
			// We have a spanning block, we display it like other blocks.
			resolvePercentagesBox(block, containingBlock, 0)
			block.Box().PositionX = box.ContentBoxX()
			block.Box().PositionY = currentPositionY
			newChild, tmp, _ := blockLevelLayout(context, block, originalBottomSpace, skipStack,
				containingBlock, pageIsEmpty, absoluteBoxes, fixedBoxes, &adjoiningMargins, false, -1)
			skipStack = nil
			if newChild == nil {
				lastLoop = true
				breakPage = true
				break
			}
			adjoiningMargins = tmp.adjoiningMargins
			newChildren = append(newChildren, newChild)
			currentPositionY = newChild.Box().BorderHeight() + newChild.Box().BorderBoxY()
			adjoiningMargins = append(adjoiningMargins, newChild.Box().MarginBottom.V())
			if tmp.resumeAt != nil {
				lastLoop = true
				breakPage = true
				columnSkipStack = tmp.resumeAt
				break
			}
			pageIsEmpty = false
			continue
		}

		// We have a list of children that we have to balance between columns.
		columnChildren := columnChildrenOrBlock.list

		// Find the total height available for the first run
		currentPositionY += collapseMargin(adjoiningMargins)
		adjoiningMargins = nil
		columnBox := createColumnBox(box_, containingBlock, columnChildren, width, currentPositionY)
		height := context.pageBottom - currentPositionY - originalBottomSpace
		maxHeight := height

		// Try to render columns until the content fits, increase the column
		// height step by step
		columnSkipStack = skipStack
		lostSpace := pr.Inf
		originalExcludedShapes := append([]*bo.BoxFields(nil), *context.excludedShapes...) // copy
		originalPageIsEmpty := pageIsEmpty
		pageIsEmpty = false
		stopRendering, balancing := false, false
		for {
			// Remove extra excluded shapes introduced during the previous loop
			*context.excludedShapes = (*context.excludedShapes)[:len(originalExcludedShapes)]

			// Render the columns
			columnSkipStack = skipStack
			var (
				consumedHeights []pr.Float
				newBoxes        []Box
			)
			for i := 0; i < count; i += 1 {
				// Render one column
				newBox, tmp := blockBoxLayout(context, columnBox, context.pageBottom-currentPositionY-height,
					columnSkipStack, containingBlock, pageIsEmpty || !balancing, new([]*AbsolutePlaceholder), new([]*AbsolutePlaceholder), new([]pr.Float),
					false, -1)
				resumeAt := tmp.resumeAt
				nextPage = tmp.nextPage
				if newBox == nil {
					// We didn't render anything, retry
					columnSkipStack = tree.ResumeStack{0: nil}
					break
				}
				newBoxes = append(newBoxes, newBox)
				columnSkipStack = resumeAt

				// Calculate consumed height, empty space and next box height
				var lastInFlowChildren *bo.BoxFields
				for _, child := range newBox.Box().Children {
					if ch := child.Box(); ch.IsInNormalFlow() {
						lastInFlowChildren = ch
					}
				}

				var consumedHeight, emptySpace, nextBoxHeight pr.Float
				if lastInFlowChildren != nil {
					// Get the empty space at the bottom of the column box
					consumedHeight = lastInFlowChildren.MarginHeight() + lastInFlowChildren.PositionY - currentPositionY
					emptySpace = height - consumedHeight

					// Get the minimum size needed to render the next box
					if columnSkipStack != nil {
						nextBox, _ := blockBoxLayout(context, columnBox, pr.Inf,
							columnSkipStack, containingBlock, true, new([]*AbsolutePlaceholder), new([]*AbsolutePlaceholder), new([]pr.Float),
							false, -1)
						for _, child := range nextBox.Box().Children {
							if child.Box().IsInNormalFlow() {
								nextBoxHeight = child.Box().MarginHeight()
								break
							}
						}
						removePlaceholders(context, []bo.Box{nextBox}, new([]*AbsolutePlaceholder), new([]*AbsolutePlaceholder))
					}
				}
				// else
				// consumedHeight = 0
				// nextBoxHeight = 0
				// emptySpace = 0

				consumedHeights = append(consumedHeights, consumedHeight)

				// Append the size needed to render the next box in this
				// column.
				//
				// The next box size may be smaller than the empty space, for
				// example when the next box can't be separated from its own
				// next box. In this case we don't try to find the real value
				// and let the workaround below fix this for us.
				//
				// We also want to avoid very small values that may have been
				// introduced by rounding errors. As the workaround below at
				// least adds 1 pixel for each loop, we can ignore lost spaces
				// lower than 1px.
				if nextBoxHeight-emptySpace > 1 {
					lostSpace = pr.Min(lostSpace, nextBoxHeight-emptySpace)
				}

				// Stop if we already rendered the whole content
				if resumeAt == nil {
					break
				}
			}

			// Remove placeholders but keep the current footnote area height
			lastFootnotesHeight = 0
			if h := context.currentFootnoteArea.Height; h != pr.AutoF {
				lastFootnotesHeight = context.currentFootnoteArea.MarginHeight()
			}
			removePlaceholders(context, newBoxes, new([]*AbsolutePlaceholder), new([]*AbsolutePlaceholder))

			if lastLoop {
				break
			}

			if balancing {
				if columnSkipStack == nil {
					// We rendered the whole content, stop
					break
				}

				// Increase the column heights and render them again
				addHeight := lostSpace
				if lostSpace == pr.Inf {
					addHeight = 1
				}
				height += addHeight

				if height > maxHeight {
					// We reached max height, stop rendering
					height = maxHeight
					stopRendering = true
					break
				}
			} else {
				L := len(footnoteAreaHeights)
				if !isInFloats(lastFootnotesHeight, footnoteAreaHeights) {
					// Footnotes have been rendered, try to re-render with the
					// new footnote area height
					height -= lastFootnotesHeight - footnoteAreaHeights[L-1]
					footnoteAreaHeights = append(footnoteAreaHeights, lastFootnotesHeight)
					continue
				}
				everythingFits := (columnSkipStack == nil && pr.Maxs(consumedHeights...) <= maxHeight)
				if everythingFits {
					// Everything fits, start expanding columns at the average
					// of the column heights
					maxHeight -= lastFootnotesHeight
					if style.GetColumnFill() == "balance" {
						balancing = true
						height = sum(consumedHeights) / pr.Float(count)
					} else {
						break
					}
				} else {
					// Content overflows even at maximum height, stop now and
					// let the columns continue on the next page
					height += footnoteAreaHeights[L-1]
					if L > 2 {
						lastFootnotesHeight = pr.Min(lastFootnotesHeight, footnoteAreaHeights[L-1])
					}
					height -= lastFootnotesHeight
					stopRendering = true
					break
				}
			}
		}

		bottomSpace = pr.Max(bottomSpace, context.pageBottom-currentPositionY-height)

		// Replace the current box children with real columns
		i := 0
		var maxColumnHeight pr.Float
		var columns []Box
		for {
			i_ := pr.Float(i)

			columnBox = createColumnBox(box_, containingBlock, columnChildren, width, currentPositionY)
			if style.GetDirection() == "rtl" {
				columnBox.Box().PositionX += box.Width.V() - (i_+1)*width - i_*gap
			} else {
				columnBox.Box().PositionX += i_ * (width + gap)
			}
			newChild, tmp := blockBoxLayout(context, columnBox, bottomSpace, skipStack,
				containingBlock, originalPageIsEmpty, absoluteBoxes, fixedBoxes, new([]pr.Float),
				false, -1)
			columnSkipStack = tmp.resumeAt
			columnNextPage := tmp.nextPage

			if traceMode {
				traceLogger.Dump(fmt.Sprintf("column %d -> %s", i, columnSkipStack))
			}

			if newChild == nil {
				breakPage = true
				break
			}
			nextPage = columnNextPage
			skipStack = columnSkipStack
			columns = append(columns, newChild)
			maxColumnHeight = pr.Max(maxColumnHeight, newChild.Box().MarginHeight())
			if skipStack == nil {
				bottomSpace = originalBottomSpace
				break
			}
			i += 1
			if i == count && !heightDefined {
				// [If] a declaration that constrains the column height
				// (e.g., using height || max-height). In this case,
				// additional column boxes are created in the inline
				// direction.
				break
			}
		}

		// Update the current y position and set the columns’ height
		currentPositionY += pr.Min(maxHeight, maxColumnHeight)
		for _, column := range columns {
			column.Box().Height = maxColumnHeight
			newChildren = append(newChildren, column)
		}

		skipStack = nil
		pageIsEmpty = false

		if stopRendering {
			break
		}
	}

	// Report footnotes above the defined footnotes height
	reportFootnotes(context, lastFootnotesHeight)

	if len(box.Children) != 0 && len(newChildren) == 0 {
		// The box has children but none can be drawn, let's skip the whole box
		context.inColumn = false

		if traceMode {
			traceLogger.Dump("columnLayout: exit early")
		}

		return nil, blockLayout{resumeAt: tree.ResumeStack{0: nil}, nextPage: tree.PageBreak{Break: "any"}}
	}

	// Set the height of box and the containing box
	box.Children = newChildren
	currentPositionY += collapseMargin(adjoiningMargins)
	height := currentPositionY - box.ContentBoxY()
	var heightDifference pr.Float
	if box.Height == pr.AutoF {
		box.Height = height
		heightDifference = 0
	} else {
		heightDifference = box.Height.V() - height
	}

	// Update the latest columns’ height to respect min-height
	if box.MinHeight != pr.AutoF && box.MinHeight.V() > box.Height.V() {
		heightDifference += box.MinHeight.V() - box.Height.V()
		box.Height = box.MinHeight
	}
	for _, child := range reversedBoxes(newChildren) {
		if child.Box().IsColumn {
			child.Box().Height = child.Box().Height.V() + heightDifference
		} else {
			break
		}
	}

	if style.GetPosition().String == "relative" {
		// New containing block, resolve the layout of the absolute descendants
		for _, absoluteBox := range *absoluteBoxes {
			absoluteLayout(context, absoluteBox, box_, fixedBoxes, bottomSpace, nil)
		}
	}

	// Calculate skip stack
	if columnSkipStack != nil {
		skip, _ = columnSkipStack.Unpack()
		skipStack = tree.ResumeStack{index + skip: columnSkipStack[skip]}
	} else if breakPage {
		skipStack = tree.ResumeStack{index: nil}
	}

	// Update page bottom according to the new footnotes
	if context.currentFootnoteArea.Height != pr.AutoF {
		context.pageBottom += footnoteAreaHeights[0]
		context.pageBottom -= context.currentFootnoteArea.MarginHeight()
	}

	context.inColumn = false

	if traceMode {
		traceLogger.Dump(fmt.Sprintf("columnsLayout -> %s", skipStack))
	}

	return box_, blockLayout{resumeAt: skipStack, nextPage: nextPage}
}

// Report footnotes above the defined footnotes height.
func reportFootnotes(context *layoutContext, footnotesHeight pr.Float) {
	if len(context.currentPageFootnotes) == 0 {
		return
	}
	// Report and count footnotes
	reportedFootnotes := 0
	for context.currentFootnoteArea.MarginHeight() > footnotesHeight {
		context.reportFootnote(context.currentPageFootnotes[len(context.currentPageFootnotes)-1])
		reportedFootnotes += 1
	}
	// Revert reported footnotes, as they’ve been reported starting from the
	// last one
	if reportedFootnotes >= 2 {
		L := len(context.reportedFootnotes)
		reverseB(context.reportedFootnotes[L-reportedFootnotes:])
	}
}

// Create a column box including given children.
func createColumnBox(box_ Box, containingBlock containingBlock, children []Box, width, positionY pr.Float) bo.BlockBoxITF {
	columnBox := box_.Type().AnonymousFrom(box_, children).(bo.BlockBoxITF) // AnonymousFrom preserves concrete types
	resolvePercentagesBox(columnBox, containingBlock, 0)
	columnBox.Box().IsColumn = true
	columnBox.Box().Width = width
	columnBox.Box().PositionX = box_.Box().ContentBoxX()
	columnBox.Box().PositionY = positionY
	return columnBox
}
