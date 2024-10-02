package layout

import (
	"fmt"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"
)

// Layout for tables and internal table boxes.

// Layout for a table box.
func tableLayout(context *layoutContext, table_ bo.TableBoxITF, bottomSpace pr.Float, skipStack tree.ResumeStack,
	pageIsEmpty bool, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder,
) (bo.BlockLevelBoxITF, blockLayout) {
	table := table_.Table()

	hasHeader := len(table.Children) != 0 && table.Children[0].Box().IsHeader
	hasFooter := len(table.Children) != 0 && table.Children[len(table.Children)-1].Box().IsFooter
	collapse := table.Style.GetBorderCollapse() == "collapse"
	removeStartDecoration := skipStack != nil && !hasHeader

	table_.RemoveDecoration(&table.BoxFields, removeStartDecoration, false)

	columnWidths := table.ColumnWidths
	var borderSpacingX, borderSpacingY pr.Float
	if !collapse {
		tmp := table.Style.GetBorderSpacing()
		borderSpacingX, borderSpacingY = tmp[0].Value, tmp[1].Value
	}

	table.ColumnPositions = nil
	positionX := table.ContentBoxX()
	rowsLeftX := positionX + borderSpacingX
	var rowsWidth pr.Float
	if table.Style.GetDirection() == "ltr" {
		positionX := table.ContentBoxX()
		rowsX := positionX + borderSpacingX
		for _, width := range columnWidths {
			positionX += borderSpacingX
			table.ColumnPositions = append(table.ColumnPositions, positionX)
			positionX += width
		}
		rowsWidth = positionX - rowsX
	} else {
		positionX := table.ContentBoxX() + table.Width.V()
		rowsX := positionX - borderSpacingX
		for _, width := range columnWidths {
			positionX -= borderSpacingX
			positionX -= width
			table.ColumnPositions = append(table.ColumnPositions, positionX)
		}
		rowsWidth = rowsX - positionX
	}

	var skippedRows int
	if collapse {
		table.SkipCellBorderTop = false
		table.SkipCellBorderBottom = false
		splitCells := false
		if skipStack != nil {
			skippedGroups, groupSkipStack := skipStack.Unpack()
			skippedRows = 0
			if groupSkipStack != nil {
				var cellSkipStack tree.ResumeStack
				skippedRows, cellSkipStack = groupSkipStack.Unpack()
				if cellSkipStack != nil {
					splitCells = true
				}
			}
			for _, group := range table.Children[:skippedGroups] {
				skippedRows += len(group.Box().Children)
			}
		}

		if !splitCells && !hasHeader {
			horizontalBorders := table.CollapsedBorderGrid.Horizontal
			if len(horizontalBorders) != 0 {
				var max pr.Fl
				for _, tmp := range horizontalBorders[skippedRows] {
					if tmp.Width > max {
						max = tmp.Width
					}
				}
				table.BorderTopWidth = pr.Float(max / 2)
			}
		}
	}

	// Make this a sub-function so that many local variables like rowsX
	// don't need to be passed as parameters.
	groupLayout := func(group_ Box, positionY, bottomSpace pr.Float, pageIsEmpty bool, skipStack tree.ResumeStack) (Box, tree.ResumeStack, tree.PageBreak) {
		var resumeAt tree.ResumeStack
		nextPage := tree.PageBreak{Break: "any"}
		originalPageIsEmpty := pageIsEmpty
		resolvePercentagesBox(group_, &table.BoxFields, 0)
		group := group_.Box()
		group.PositionX = rowsLeftX
		group.PositionY = positionY
		group.Width = rowsWidth
		newGroupChildren := []Box{}
		// For each rows, cells for which this is the last row (with rowspan)
		endingCellsByRow := make([][]Box, len(group.Children))

		isGroupStart := skipStack == nil
		skip := 0
		if !isGroupStart {
			skip, skipStack = skipStack.Unpack()
		}

		for i, row_ := range group.Children[skip:] {
			row := row_.Box()
			indexRow := i + skip
			row.Index = indexRow
			if len(newGroupChildren) != 0 {
				pageBreak := blockLevelPageBreak(newGroupChildren[len(newGroupChildren)-1], row_)
				if forcePageBreak(pageBreak, context) {
					nextPage.Break = pageBreak
					resumeAt = tree.ResumeStack{indexRow: nil}
					break
				}
			}

			resolvePercentagesBox(row_, &table.BoxFields, 0)
			row.PositionX = rowsLeftX
			row.PositionY = positionY
			row.Width = rowsWidth
			// Place cells at the top of the row and layout their content
			var newRowChildren []Box
			for indexCell, cell_ := range row.Children {
				cell := cell_.Box()
				var spannedWidths []pr.Float
				if cell.GridX < len(columnWidths) {
					spannedWidths = columnWidths[cell.GridX:]
				}
				if cell.Colspan < len(spannedWidths) {
					spannedWidths = spannedWidths[:cell.Colspan]
				}
				// In the fixed layout the grid width is set by cells in
				// the first row and column elements.
				// This may be less than the previous value of cell.colspan
				// if that would bring the cell beyond the grid width.
				cell.Colspan = len(spannedWidths)
				if cell.Colspan == 0 {
					// The cell is entierly beyond the grid width, remove it
					// entierly. Subsequent cells in the same row have greater
					// gridX, so they are beyond too.
					ignoredCells := row.Children[indexCell:]
					logger.WarningLogger.Printf("This table row has more columns than the table, ignored %d cells: %v",
						len(ignoredCells), ignoredCells)
					break
				}
				resolvePercentagesBox(cell_, &table.BoxFields, 0)
				if table.Style.GetDirection() == "ltr" {
					cell.PositionX = table.ColumnPositions[cell.GridX]
				} else {
					cell.PositionX = table.ColumnPositions[cell.GridX+cell.Colspan-1]
				}
				cell.PositionY = row.PositionY
				cell.MarginTop = pr.Float(0)
				cell.MarginLeft = pr.Float(0)
				cell.Width = pr.Float(0)
				bordersPlusPadding := cell.BorderWidth() // with width==0
				// TODO: we should remove the number of columns with no
				// originating cells to cell.colspan, see testLayoutTableAuto49
				width := borderSpacingX*pr.Float(cell.Colspan-1) - bordersPlusPadding
				for _, sw := range spannedWidths {
					width += sw
				}
				cell.Width = width
				// The computed height is a minimum
				cell.ComputedHeight = cell.Height
				cell.Height = pr.AutoF

				var cellSkipStack tree.ResumeStack
				if len(skipStack) != 0 {
					if rs, has := skipStack[indexCell]; has {
						cellSkipStack = rs
					} else {
						cellSkipStack = tree.ResumeStack{len(cell.Children): nil}
					}
				}

				// Adapt cell and table collapsing borders when a row is split
				if cellSkipStack != nil && collapse {
					if hasHeader {
						// We have a header, we have to adapt the position of
						// the split cell to match the header’s bottom border
						headerRows := table.Children[0].Box().Children
						if L := len(headerRows); L != 0 && len(headerRows[L-1].Box().Children) != 0 {
							max := -pr.Inf
							for _, header := range headerRows[L-1].Box().Children {
								if w := header.Box().BorderBottomWidth; w > max {
									max = w
								}
							}
							cell.PositionY += max
						}
					} else {
						// We don’t have a header, we have to skip the
						// decoration at the top of the table when it’s drawn
						table.SkipCellBorderTop = true
					}
				}

				// First try to render content as if there was already something
				// on the page to avoid hitting block_level_layout’s TODO. Then
				// force to render something if the page is actually empty, or
				// just draw an empty cell otherwise. See
				// test_table_break_children_margin.
				newCell, tmp, _ := blockContainerLayout(context, cell_, bottomSpace, cellSkipStack, pageIsEmpty,
					absoluteBoxes, fixedBoxes, new([]pr.Float), false, -1)
				cellResumeAt := tmp.resumeAt
				if newCell == nil {
					cell_ = bo.CopyWithChildren(cell_, nil)
					cell_, _, _ = blockContainerLayout(context, cell_, bottomSpace, cellSkipStack, true,
						new([]*AbsolutePlaceholder), new([]*AbsolutePlaceholder), new([]pr.Float), false, -1)
					cellResumeAt = tree.ResumeStack{0: nil}
				} else {
					cell_ = newCell
				}
				cell = cell_.Box()
				cell_.RemoveDecoration(cell, cellSkipStack != nil, cellResumeAt != nil)
				if cellResumeAt != nil {
					if resumeAt == nil {
						resumeAt = tree.ResumeStack{indexRow: tree.ResumeStack{}}
					}
					resumeAt[indexRow][indexCell] = cellResumeAt
				}

				any := false
				for _, child := range cell.Children {
					if child.Box().IsFloated() || child.Box().IsInNormalFlow() {
						any = true
						break
					}
				}
				cell.Empty = !any

				cell.ContentHeight = cell.Height.V()
				if cell.ComputedHeight != pr.AutoF {
					cell.Height = pr.Max(cell.Height.V(), cell.ComputedHeight.V())
				}
				newRowChildren = append(newRowChildren, cell_)
			}

			if resumeAt != nil && !pageIsEmpty {
				if bi := row.Style.GetBreakInside(); avoidPageBreak(string(bi), context) {
					resumeAt = tree.ResumeStack{indexRow: {}}
					break
				}
			}

			row_ = bo.CopyWithChildren(row_, newRowChildren)
			row = row_.Box()

			// Table height algorithm
			// http://www.w3.org/TR/CSS21/tables.html#height-layout

			// cells with vertical-align: baseline
			var baselineCells []Box
			for _, cell_ := range row.Children {
				cell := cell_.Box()
				verticalAlign := cell.Style.GetVerticalAlign()
				if verticalAlign.S == "top" || verticalAlign.S == "middle" || verticalAlign.S == "bottom" {
					cell.VerticalAlign = verticalAlign.S
				} else {
					// Assume "baseline" for any other value
					cell.VerticalAlign = "baseline"
					cell.Baseline = cellBaseline(cell_)
					baselineCells = append(baselineCells, cell_)
				}
			}
			if len(baselineCells) != 0 {
				maxBaseline := -pr.Inf
				for _, cell := range baselineCells {
					if bs := cell.Box().Baseline.V(); bs > maxBaseline {
						maxBaseline = bs
					}
				}
				row.Baseline = maxBaseline
				for _, cell := range baselineCells {
					extra := row.Baseline.V() - cell.Box().Baseline.V()
					if cell.Box().Baseline != row.Baseline && extra != 0 {
						addTopPadding(cell.Box(), extra)
					}
				}
			}

			// row height
			for _, cell := range row.Children {
				endingCellsByRow[cell.Box().Rowspan-1] = append(endingCellsByRow[cell.Box().Rowspan-1], cell)
			}
			endingCells := endingCellsByRow[0]
			endingCellsByRow = endingCellsByRow[1:]
			var rowBottomY pr.Float
			if len(endingCells) != 0 { // in this row
				if row.Height == pr.AutoF {
					for _, cell := range endingCells {
						if v := cell.Box().PositionY + cell.Box().BorderHeight(); v > rowBottomY {
							rowBottomY = v
						}
					}
					row.Height = pr.Max(rowBottomY-row.PositionY, 0)
				} else {
					var m pr.Float
					for _, rowCell := range endingCells {
						if v := rowCell.Box().Height.V(); v > m {
							m = v
						}
					}
					row.Height = pr.Max(row.Height.V(), m)
					rowBottomY = row.PositionY + row.Height.V()
				}
			} else {
				rowBottomY = row.PositionY
				row.Height = pr.Float(0)
			}

			if len(baselineCells) == 0 {
				row.Baseline = rowBottomY
			}

			// Add extra padding to make the cells the same height as the row
			// and honor vertical-align
			for _, cell_ := range endingCells {
				cell := cell_.Box()
				cellBottomY := cell.PositionY + cell.BorderHeight()
				extra := rowBottomY - cellBottomY
				if extra != 0 {
					if cell.VerticalAlign == "bottom" {
						addTopPadding(cell, extra)
					} else if cell.VerticalAlign == "middle" {
						extra /= 2.
						addTopPadding(cell, extra)
						cell.PaddingBottom = cell.PaddingBottom.V() + extra
					} else {
						cell.PaddingBottom = cell.PaddingBottom.V() + extra
					}
				}
				if cell.ComputedHeight != pr.AutoF {
					var verticalAlignShift pr.Float
					if cell.VerticalAlign == "middle" {
						verticalAlignShift = (cell.ComputedHeight.V() - cell.ContentHeight.V()) / 2
					} else if cell.VerticalAlign == "bottom" {
						verticalAlignShift = cell.ComputedHeight.V() - cell.ContentHeight.V()
					}
					if verticalAlignShift > 0 {
						for _, child := range cell.Children {
							child.Translate(child, 0, verticalAlignShift, false)
						}
					}
				}
			}

			nextPositionY := row.PositionY + row.Height.V()
			if resumeAt == nil {
				nextPositionY += borderSpacingY
			}

			// Break if one cell was broken
			breakCell := false
			if resumeAt != nil {
				allEmpty := true
				for _, child := range row.Children {
					if !child.Box().Empty {
						allEmpty = false
						break
					}
				}

				if allEmpty {
					// No cell was displayed, give up row
					nextPositionY = pr.Inf
					pageIsEmpty = false
					resumeAt = nil
				} else {
					breakCell = true
				}
			}

			// Break if this row overflows the page, unless there is no
			// other content on the page.
			if !pageIsEmpty && context.overflowsPage(bottomSpace, nextPositionY) {
				if len(newGroupChildren) != 0 {
					previousRow := newGroupChildren[len(newGroupChildren)-1]
					pageBreak := blockLevelPageBreak(previousRow, row_)
					if avoidPageBreak(pageBreak, context) {
						newGroupChildrenTmp, resumeAtTmp := findEarlierPageBreak(context, newGroupChildren, absoluteBoxes, fixedBoxes)
						if newGroupChildrenTmp != nil || resumeAtTmp != nil {
							newGroupChildren, resumeAt = newGroupChildrenTmp, resumeAtTmp
							break
						}
					} else {
						resumeAt = tree.ResumeStack{indexRow: nil}
						break
					}
				}

				if originalPageIsEmpty {
					resumeAt = tree.ResumeStack{indexRow: nil}
				} else {
					return nil, nil, nextPage
				}

				break
			}

			newGroupChildren = append(newGroupChildren, row_)
			positionY = nextPositionY
			pageIsEmpty = false
			skipStack = nil

			if breakCell && collapse && !hasFooter {
				table.SkipCellBorderBottom = true
			}

			if breakCell || resumeAt != nil {
				break
			}
		}

		// Do not keep the row group if we made a page break
		// before any of its rows or with "avoid"
		if bi := group.Style.GetBreakInside(); resumeAt != nil && !originalPageIsEmpty && (avoidPageBreak(string(bi), context) || len(newGroupChildren) == 0) {
			return nil, nil, nextPage
		}

		group_ = bo.CopyWithChildren(group_, newGroupChildren)
		group = group_.Box()
		group_.RemoveDecoration(group, !isGroupStart, resumeAt != nil)
		// Set missing baselines in a second loop because of rowspan
		for _, row_ := range group.Children {
			row := row_.Box()
			if row.Baseline == nil {
				if len(row.Children) != 0 {
					// lowest bottom content edge
					var max pr.Float
					for _, cell := range row.Children {
						if v := cell.Box().ContentBoxY() + cell.Box().Height.V(); v > max {
							max = v
						}
					}
					row.Baseline = max - row.PositionY
				} else {
					row.Baseline = pr.Float(0)
				}
			}

		}
		group.Height = positionY - group.PositionY
		if len(group.Children) != 0 {
			// The last border spacing is outside of the group.
			group.Height = group.Height.V() - borderSpacingY
		}

		return group_, resumeAt, nextPage
	}

	bodyGroupsLayout := func(skipStack tree.ResumeStack, positionY, bottomSpace pr.Float, pageIsEmpty bool) ([]Box, tree.ResumeStack, tree.PageBreak, pr.Float) {
		skip := 0
		if skipStack != nil {
			skip, skipStack = skipStack.Unpack()
		}
		newTableChildren := []Box{}
		var resumeAt tree.ResumeStack
		nextPage := tree.PageBreak{Break: "any"}

		for i, group_ := range table.Children[skip:] {
			group := group_.Box()
			indexGroup := i + skip
			group.Index = indexGroup

			if group.IsHeader || group.IsFooter {
				continue
			}

			if L := len(newTableChildren); L != 0 {
				pageBreak := blockLevelPageBreak(newTableChildren[L-1], group_)
				if forcePageBreak(pageBreak, context) {
					nextPage.Break = pageBreak
					resumeAt = tree.ResumeStack{indexGroup: nil}
					break
				}
			}
			var newGroup Box
			newGroup, resumeAt, nextPage = groupLayout(group_, positionY, bottomSpace, pageIsEmpty, skipStack)
			skipStack = nil

			if newGroup == nil {
				if L := len(newTableChildren); L != 0 {
					previousGroup := newTableChildren[L-1]
					pageBreak := blockLevelPageBreak(previousGroup, group_)
					if avoidPageBreak(pageBreak, context) {
						v1, v2 := findEarlierPageBreak(context, newTableChildren, absoluteBoxes, fixedBoxes)
						if v1 != nil && v2 != nil {
							newTableChildren, resumeAt = v1, v2
							break
						}
					}
					resumeAt = tree.ResumeStack{indexGroup: nil}
				} else {
					return nil, nil, nextPage, positionY
				}
				break
			}

			newTableChildren = append(newTableChildren, newGroup)
			positionY += newGroup.Box().Height.V() + borderSpacingY
			pageIsEmpty = false

			if resumeAt != nil {
				resumeAt = tree.ResumeStack{indexGroup: resumeAt}
				break
			}
		}

		return newTableChildren, resumeAt, nextPage, positionY
	}

	// Layout for row groups, rows and cells
	positionY := table.ContentBoxY()
	if skipStack == nil {
		positionY += borderSpacingY
	}

	initialPositionY := positionY
	var tableRows []Box
	for _, rowGroup := range table.Children {
		if !rowGroup.Box().IsHeader && !rowGroup.Box().IsFooter {
			tableRows = append(tableRows, rowGroup)
		}
	}

	allGroupsLayout := func() (Box, []Box, Box, pr.Float, tree.ResumeStack, tree.PageBreak) {
		var (
			header, footer             Box
			headerHeight, footerHeight pr.Float
			resumeAt                   tree.ResumeStack
			nextPage                   tree.PageBreak
		)

		// If the page is not empty, we try to render the header and the footer
		// on it. If the table does not fit on the page, we try to render it on
		// the next page.
		//
		// If the page is empty and the header and footer are too big, there
		// are not rendered. If no row can be rendered because of the header and
		// the footer, the header and/or the footer are not rendered.

		headerFooterMaxBottomSpace := -pr.Inf
		if pageIsEmpty {
			headerFooterMaxBottomSpace = bottomSpace
		}

		if hasHeader {
			header = table.Children[0]
			header, resumeAt, _ = groupLayout(header, positionY, headerFooterMaxBottomSpace, false, nil)
			if header != nil && resumeAt == nil {
				headerHeight = header.Box().Height.V() + borderSpacingY
			} else { // Header too big for the page
				header = nil
			}
		} // else header = nil

		if hasFooter {
			footer = table.Children[len(table.Children)-1]
			footer, resumeAt, _ = groupLayout(footer, positionY, headerFooterMaxBottomSpace, false, nil)
			if footer != nil && resumeAt == nil {
				footerHeight = footer.Box().Height.V() + borderSpacingY
			} else { // Footer too big for the page
				footer = nil
			}
		} // else footer = nil

		// Don't remove headers and footers if breaks are avoided in line groups
		skip := 0
		if skipStack != nil {
			skip, _ = skipStack.Unpack()
		}
		avoidBreaks := false
		for _, group_ := range table.Children[skip:] {
			group := group_.Box()
			if bi := group.Style.GetBreakInside(); !group.IsHeader && !group.IsFooter {
				avoidBreaks = avoidPageBreak(string(bi), context)
				break
			}
		}

		if header != nil && footer != nil {
			// Try with both the header and footer
			var (
				newTableChildren []Box
				endPositionY     pr.Float
			)
			newTableChildren, resumeAt, nextPage, endPositionY = bodyGroupsLayout(skipStack, positionY+headerHeight,
				bottomSpace+footerHeight, avoidBreaks)
			if len(newTableChildren) != 0 || len(tableRows) == 0 || !pageIsEmpty {
				footer.Translate(footer, 0, endPositionY-footer.Box().PositionY, false)
				endPositionY += footerHeight
				return header, newTableChildren, footer, endPositionY, resumeAt, nextPage
			} else {
				// We could not fit any content, drop the footer
				footer = nil
			}
		}

		if header != nil && footer == nil {
			// Try with just the header
			var (
				newTableChildren []Box
				endPositionY     pr.Float
			)
			newTableChildren, resumeAt, nextPage, endPositionY = bodyGroupsLayout(skipStack, positionY+headerHeight, bottomSpace, avoidBreaks)
			if len(newTableChildren) != 0 || len(tableRows) == 0 || !pageIsEmpty {
				return header, newTableChildren, footer, endPositionY, resumeAt, nextPage
			} else {
				// We could not fit any content, drop the header
				header = nil
			}
		}

		if footer != nil && header == nil {
			// Try with just the footer
			var (
				newTableChildren []Box
				endPositionY     pr.Float
			)
			newTableChildren, resumeAt, nextPage, endPositionY = bodyGroupsLayout(skipStack, positionY, bottomSpace+footerHeight, avoidBreaks)
			if len(newTableChildren) != 0 || len(tableRows) == 0 || !pageIsEmpty {
				footer.Translate(footer, 0, endPositionY-footer.Box().PositionY, false)
				endPositionY += footerHeight
				return header, newTableChildren, footer, endPositionY, resumeAt, nextPage
			} else {
				// We could not fit any content, drop the footer
				footer = nil
			}
		}

		if header != nil || footer != nil {
			panic(fmt.Sprintf("expected empty header and footer, got %v %v", header, footer))
		}
		newTableChildren, resumeAt, nextPage, endPositionY := bodyGroupsLayout(skipStack, positionY, bottomSpace, pageIsEmpty)
		return header, newTableChildren, footer, endPositionY, resumeAt, nextPage
	}

	// Closure getting the column cells.
	getColumnCells := func(table *bo.TableBox, column *bo.BoxFields) func() []Box {
		return func() []Box {
			var out []Box
			for _, rowGroup := range table.Children {
				for _, row := range rowGroup.Box().Children {
					for _, cell := range row.Box().Children {
						if cell.Box().GridX == column.GridX {
							out = append(out, cell)
						}
					}
				}
			}
			return out
		}
	}

	header, newTableChildren, footer, positionY, resumeAt, nextPage := allGroupsLayout()
	if newTableChildren == nil {
		if resumeAt != nil {
			panic(fmt.Sprintf("resumeAt should be nil, got %v", resumeAt))
		}
		return nil, blockLayout{resumeAt: resumeAt, nextPage: nextPage, adjoiningMargins: nil, collapsingThrough: false}
	}

	var newChildren []Box
	if header != nil {
		newChildren = append(newChildren, header)
	}
	newChildren = append(newChildren, newTableChildren...)
	if footer != nil {
		newChildren = append(newChildren, footer)
	}
	table_ = bo.CopyWithChildren(table_, newChildren).(bo.TableBoxITF) // CopyWithChildren is type stable
	table = table_.Table()
	removeEndDecoration := resumeAt != nil && !hasFooter
	table_.RemoveDecoration(&table.BoxFields, removeStartDecoration, removeEndDecoration)
	if collapse {
		table.SkippedRows = skippedRows
	}

	// If the height property has a bigger value, just add blank space
	// below the last row group.
	var th pr.Float
	if table.Height != pr.AutoF {
		th = table.Height.V()
	}
	table.Height = pr.Max(th, positionY-table.ContentBoxY())

	// Layout for column groups and columns
	columnsHeight := positionY - initialPositionY
	if len(table.Children) != 0 {
		// The last border spacing is below the columns.
		columnsHeight -= borderSpacingY
	}
	for _, group := range table.ColumnGroups {
		for _, column_ := range group.Children {
			column := column_.Box()
			resolvePercentagesBox(column_, &table.BoxFields, 0)
			if column.GridX < len(table.ColumnPositions) {
				column.PositionX = table.ColumnPositions[column.GridX]
				column.PositionY = initialPositionY
				column.Width = columnWidths[column.GridX]
				column.Height = columnsHeight
			} else {
				// Ignore extra empty columns
				column.PositionX = 0
				column.PositionY = 0
				column.Width = pr.Float(0)
				column.Height = pr.Float(0)
			}
			resolvePercentagesBox(group, &table.BoxFields, 0)
			column.GetCells = getColumnCells(table, column)
		}
		first := group.Children[0].Box()
		last := group.Children[len(group.Children)-1].Box()
		group.PositionX = first.PositionX
		group.PositionY = initialPositionY
		group.Width = last.PositionX + last.Width.V() - first.PositionX
		group.Height = columnsHeight
	}

	avoidBreak := avoidPageBreak(string(table.Style.GetBreakInside()), context)
	if resumeAt != nil && !pageIsEmpty && avoidBreak {
		table_ = nil
		resumeAt = nil
	}

	if traceMode {
		traceLogger.DumpTree(table_, "tableLayout-end")
	}

	return table_, blockLayout{resumeAt: resumeAt, nextPage: nextPage, adjoiningMargins: nil, collapsingThrough: false}
}

// Increase the top padding of a box. This also translates the children.
func addTopPadding(box *bo.BoxFields, extraPadding pr.Float) {
	box.PaddingTop = box.PaddingTop.V() + extraPadding
	for _, child := range box.Children {
		child.Translate(child, 0, extraPadding, false)
	}
}

// Run the fixed table layout and return a list of column widths
// http://www.w3.org/TR/CSS21/tables.html#fixed-table-layout
func fixedTableLayout(box *bo.BoxFields) {
	table := box.GetWrappedTable().Table()
	if table.Width == pr.AutoF {
		panic("table width can't be auto here")
	}
	var allColumns []Box
	for _, columnGroup := range table.ColumnGroups {
		allColumns = append(allColumns, columnGroup.Box().Children...)
	}

	var firstRowCells []Box
	if len(table.Children) != 0 && len(table.Children[0].Box().Children) != 0 {
		firstRowgroup := table.Children[0].Box()
		firstRowCells = firstRowgroup.Children[0].Box().Children
	}
	var sum int
	for _, cell := range firstRowCells {
		sum += cell.Box().Colspan
	}
	numColumns := utils.MaxInt(len(allColumns), sum)
	// ``None`` means not know yet.
	columnWidths := make([]pr.MaybeFloat, numColumns)

	// `width` on column boxes
	for i, column_ := range allColumns {
		column := column_.Box()
		column.Width = resolveOnePercentage(column.Style.GetWidth(), pr.PWidth, table.Width.V(), 0)
		if column.Width != pr.AutoF {
			columnWidths[i] = column.Width
		}
	}

	var borderSpacingX pr.Float
	if table.Style.GetBorderCollapse() == "separate" {
		borderSpacingX = table.Style.GetBorderSpacing()[0].Value
	}

	// `width` on cells of the first row.
	i := 0
	for _, cell_ := range firstRowCells {
		cell := cell_.Box()
		resolvePercentagesBox(cell_, &table.BoxFields, 0)
		if cell.Width != pr.AutoF {
			width := cell.BorderWidth()
			width -= borderSpacingX * pr.Float(cell.Colspan-1)
			// In the general case, this width affects several columns (through
			// colspan) some of which already have a width. Subtract these
			// known widths and divide among remaining columns.
			var columnsWithoutWidth []int // and occupied by this cell
			for j := i; j < i+cell.Colspan; j++ {
				if columnWidths[j] == nil {
					columnsWithoutWidth = append(columnsWithoutWidth, j)
				} else {
					width -= columnWidths[j].V()
				}
			}
			if len(columnsWithoutWidth) != 0 {
				widthPerColumn := width / pr.Float(len(columnsWithoutWidth))
				for _, j := range columnsWithoutWidth {
					columnWidths[j] = widthPerColumn
				}
			}
		}
		i += cell.Colspan
	}

	// Distribute the remaining space equally on columns that do not have
	// a width yet.
	allBorderSpacing := borderSpacingX * pr.Float(numColumns+1)
	var columnsWithoutWidth []int
	minTableWidth := allBorderSpacing
	for i, w := range columnWidths {
		if w != nil {
			minTableWidth += w.V()
		} else {
			columnsWithoutWidth = append(columnsWithoutWidth, i)
		}
	}
	if len(columnsWithoutWidth) != 0 && table.Width.V() >= minTableWidth {
		remainingWidth := table.Width.V() - minTableWidth
		widthPerColumn := remainingWidth / pr.Float(len(columnsWithoutWidth))
		for _, i := range columnsWithoutWidth {
			columnWidths[i] = widthPerColumn
		}
	} else {
		// XXX this is bad, but we were given a broken table to work with...
		for _, i := range columnsWithoutWidth {
			columnWidths[i] = pr.Float(0)
		}
	}
	outCW := make([]pr.Float, len(columnWidths))
	var sumColumnWidths pr.Float
	for i, v := range columnWidths {
		sumColumnWidths += v.V()
		outCW[i] = v.V()
	}
	// If the sum is less than the table width,
	// distribute the remaining space equally
	extraWidth := table.Width.V() - sumColumnWidths - allBorderSpacing
	if extraWidth <= 0 {
		// substract a negative: widen the table
		table.Width = table.Width.V() - extraWidth
	} else if numColumns != 0 {
		extraPerColumn := extraWidth / pr.Float(numColumns)
		for i, w := range outCW {
			outCW[i] = w + extraPerColumn
		}
	}

	// Now we have table.Width == sum(columnWidths) + allBorderSpacing
	// with possible floating point rounding errors.
	// (unless there is zero column)
	table.ColumnWidths = outCW
}

func sum(l []pr.Float) pr.Float {
	var out pr.Float
	for _, v := range l {
		out += v
	}
	return out
}

// Run the auto table layout and return a list of column widths.
// http://www.w3.org/TR/CSS21/tables.html#auto-table-layout
func autoTableLayout(context *layoutContext, box_ Box, containingBlock bo.Point) {
	box := box_.Box()
	table_ := box.GetWrappedTable()
	table := table_.Table()
	tmp := tableAndColumnsPreferredWidths(context, box_, false)
	var margins pr.Float
	if box.MarginLeft != pr.AutoF {
		margins += box.MarginLeft.V()
	}
	if box.MarginRight != pr.AutoF {
		margins += box.MarginRight.V()
	}
	paddings := table.PaddingLeft.V() + table.PaddingRight.V()
	borders := table.BorderLeftWidth.V() + table.BorderRightWidth.V()

	cbWidth := containingBlock[0]
	availableWidth := cbWidth - margins - paddings - borders

	if table.Width == pr.AutoF {
		if availableWidth <= tmp.tableMinContentWidth {
			table.Width = tmp.tableMinContentWidth
		} else if availableWidth < tmp.tableMaxContentWidth {
			table.Width = availableWidth
		} else {
			table.Width = tmp.tableMaxContentWidth
		}
	} else {
		if table.Width.V() < tmp.tableMinContentWidth {
			table.Width = tmp.tableMinContentWidth
		}
	}

	if len(tmp.grid) == 0 {
		table.ColumnWidths = nil
		return
	}

	assignableWidth := table.Width.V() - tmp.totalHorizontalBorderSpacing
	minContentGuess := append([]pr.Float{}, tmp.columnMinContentWidths...)
	minContentPercentageGuess := append([]pr.Float{}, tmp.columnMinContentWidths...)
	minContentSpecifiedGuess := append([]pr.Float{}, tmp.columnMinContentWidths...)
	maxContentGuess := append([]pr.Float{}, tmp.columnMaxContentWidths...)
	L := 4
	guesses := [4]*[]pr.Float{
		&minContentGuess, &minContentPercentageGuess,
		&minContentSpecifiedGuess, &maxContentGuess,
	}
	for i := range tmp.grid {
		if tmp.columnIntrinsicPercentages[i] != 0 {
			minContentPercentageGuess[i] = pr.Max(
				tmp.columnIntrinsicPercentages[i]/100*assignableWidth,
				tmp.columnMinContentWidths[i])
			minContentSpecifiedGuess[i] = minContentPercentageGuess[i]
			maxContentGuess[i] = minContentPercentageGuess[i]
		} else if tmp.constrainedness[i] {
			minContentSpecifiedGuess[i] = tmp.columnMinContentWidths[i]
		}
	}

	if assignableWidth <= sum(maxContentGuess) {
		// Default values shouldn't be used, but we never know.
		// See https://github.com/Kozea/WeasyPrint/issues/770
		lowerGuess := guesses[0]
		upperGuess := guesses[L-1]

		// We have to work around floating point rounding errors here.
		// The 1e-9 value comes from PEP 485.
		for _, guess := range guesses {
			if sum(*guess) <= assignableWidth*(1+1e-9) {
				lowerGuess = guess
			} else {
				break
			}
		}
		for i := range guesses {
			guess := guesses[L-1-i]
			if sum(*guess) >= assignableWidth*(1-1e-9) {
				upperGuess = guess
			} else {
				break
			}
		}
		if upperGuess == lowerGuess {
			table.ColumnWidths = *upperGuess
		} else {
			addedWidths := make([]pr.Float, len(tmp.grid))
			var sl, saw, availableRatio pr.Float
			for i := range tmp.grid {
				addedWidths[i] = (*upperGuess)[i] - (*lowerGuess)[i]
				sl += (*lowerGuess)[i]
				saw += addedWidths[i]
			}
			if saw != 0 {
				availableRatio = (assignableWidth - sl) / saw
			}
			cw := make([]pr.Float, len(tmp.grid))
			for i := range tmp.grid {
				cw[i] = (*lowerGuess)[i] + addedWidths[i]*availableRatio
			}
			table.ColumnWidths = cw
		}
	} else {
		table.ColumnWidths = maxContentGuess
		excessWidth := assignableWidth - sum(maxContentGuess)
		excessWidth = distributeExcessWidth(context, tmp.grid, excessWidth, table.ColumnWidths, tmp.constrainedness,
			tmp.columnIntrinsicPercentages, tmp.columnMaxContentWidths, [2]int{0, len(tmp.grid)})
		if excessWidth != 0 {
			if tmp.tableMinContentWidth < table.Width.V()-excessWidth {
				// Reduce the width of the size from the excess width that has
				// not been distributed.
				table.Width = table.Width.V() - excessWidth
			} else {
				// Break rules
				var columns []int
				for i, column := range tmp.grid {
					anyColumn := false
					for _, b := range column {
						if b != nil {
							anyColumn = true
							break
						}
					}
					if anyColumn {
						columns = append(columns, i)
					}
				}
				for _, i := range columns {
					table.ColumnWidths[i] += excessWidth / pr.Float(len(columns))
				}
			}
		}
	}
}

// Find the width of each column and derive the wrapper width.
func tableWrapperWidth(context *layoutContext, wrapper_ Box, containingBlock bo.MaybePoint) {
	wrapper := wrapper_.Box()
	table := wrapper.GetWrappedTable()
	resolvePercentages(table, containingBlock, 0)

	if table.Box().Style.GetTableLayout() == "fixed" && table.Box().Width != pr.AutoF {
		fixedTableLayout(wrapper)
	} else {
		autoTableLayout(context, wrapper_, containingBlock.V())
	}
	wrapper.Width = table.Box().BorderWidth()
}

// Return the y position of a cell’s baseline from the top of its border box.
// See http://www.w3.org/TR/CSS21/tables.html#height-layout
func cellBaseline(cell Box) pr.Float {
	result := findInFlowBaseline(cell, false, bo.LineT, bo.TableRowT)
	if result != nil {
		return result.V() - cell.Box().PositionY
	} else {
		// Default to the bottom of the content area.
		return cell.Box().BorderTopWidth.V() + cell.Box().PaddingTop.V() + cell.Box().Height.V()
	}
}

// Return the absolute Y position for the first (or last) in-flow baseline
// if any or nil. Can't return "auto".
// [last] defaults to false, baselinesT to nothing
func findInFlowBaseline(box Box, last bool, baselinesT ...bo.BoxType) pr.MaybeFloat {
	if len(baselinesT) == 0 {
		baselinesT = []bo.BoxType{bo.LineT}
	}
	// TODO: synthetize baseline when needed
	// See https://www.w3.org/TR/css-align-3/#synthesize-baseline
	for _, type_ := range baselinesT { // if isinstance(box, baselinesT)
		if type_.IsInstance(box) {
			return box.Box().PositionY + box.Box().Baseline.V()
		}
	}
	if bo.ParentT.IsInstance(box) && !bo.TableCaptionT.IsInstance(box) {
		children := box.Box().Children
		if last {
			children = reversedBoxes(children)
		}
		for _, child := range children {
			if child.Box().IsInNormalFlow() {
				result := findInFlowBaseline(child, last, baselinesT...)
				if result != nil {
					return result
				}
			}
		}
	}
	return nil
}

type indexedCol struct {
	column []Box
	i      int
}

// Distribute available width to columns.
//
// Return excess width left (>0) when it's impossible without breaking rules, or 0
//
// See http://dbaron.org/css/intrinsic/#distributetocols
func distributeExcessWidth(context *layoutContext, grid [][]bo.Box, excessWidth pr.Float, columnWidths []pr.Float,
	constrainedness []bool, columnIntrinsicPercentages, columnMaxContentWidths []pr.Float, columnSlice [2]int,
) pr.Float {
	// First group
	var (
		columns       []indexedCol
		currentWidths []pr.Float
	)
	for i, column := range grid[columnSlice[0]:columnSlice[1]] {
		if !constrainedness[i+columnSlice[0]] && columnIntrinsicPercentages[i+columnSlice[0]] == 0 &&
			columnMaxContentWidths[i+columnSlice[0]] > 0 {
			v := indexedCol{i: i + columnSlice[0], column: column}
			columns = append(columns, v)
			currentWidths = append(currentWidths, columnWidths[v.i])
		}
	}
	if len(columns) != 0 {
		L := utils.MinInt(len(columnMaxContentWidths), len(currentWidths))
		var (
			sumDifferences pr.Float
			differences    = make([]pr.Float, L)
		)
		for i := 0; i < L; i++ {
			v := pr.Max(0, columnMaxContentWidths[i]-currentWidths[i])
			differences[i] = v
			sumDifferences += v
		}
		if sumDifferences > excessWidth {
			for i, difference := range differences {
				differences[i] = difference / sumDifferences * excessWidth
			}
		}
		excessWidth -= sumDifferences
		for i, difference := range differences {
			columnWidths[columns[i].i] += difference
		}
	}
	if excessWidth <= 0 {
		return 0
	}

	// Second group
	var columns_ []int
	for i := range grid[columnSlice[0]:columnSlice[1]] {
		if !constrainedness[i+columnSlice[0]] && columnIntrinsicPercentages[i+columnSlice[0]] == 0 {
			columns_ = append(columns_, i+columnSlice[0])
		}
	}

	if l := pr.Float(len(columns_)); l != 0 {
		for _, i := range columns_ {
			columnWidths[i] += excessWidth / l
		}
		return 0
	}

	// Third group
	columns, currentWidths = nil, nil
	for i, column := range grid[columnSlice[0]:columnSlice[1]] {
		if constrainedness[i+columnSlice[0]] && columnIntrinsicPercentages[i+columnSlice[0]] == 0 &&
			columnMaxContentWidths[i+columnSlice[0]] > 0 {
			v := indexedCol{column, i + columnSlice[0]}
			columns = append(columns, v)
			currentWidths = append(currentWidths, columnWidths[v.i])
		}
	}
	if len(columns) != 0 {
		L := utils.MinInt(len(columnMaxContentWidths), len(currentWidths))
		var (
			sumDifferences pr.Float
			differences    = make([]pr.Float, L)
		)
		for i := 0; i < L; i++ {
			v := pr.Max(0, columnMaxContentWidths[i]-currentWidths[i])
			differences[i] = v
			sumDifferences += v
		}
		if sumDifferences > excessWidth {
			for i, difference := range differences {
				differences[i] = difference / sumDifferences * excessWidth
			}
		}
		excessWidth -= sumDifferences
		for i, difference := range differences {
			columnWidths[columns[i].i] += difference
		}
	}
	if excessWidth <= 0 {
		return 0
	}

	// Fourth group
	columns = nil
	mapIndex := map[int]bool{}
	for i, column := range grid[columnSlice[0]:columnSlice[1]] {
		if columnIntrinsicPercentages[i+columnSlice[0]] > 0 {
			v := indexedCol{i: i + columnSlice[0], column: column}
			columns = append(columns, v)
			mapIndex[v.i] = true
		}
	}
	if L := len(columns); L != 0 {
		var fixedWidth pr.Float
		for j := range grid {
			if !mapIndex[j] {
				fixedWidth += columnWidths[j]
			}
		}
		var percentageWidth pr.Float
		for _, tmp := range columns {
			percentageWidth += columnIntrinsicPercentages[tmp.i]
		}
		var ratio pr.Float
		if fixedWidth != 0 && percentageWidth >= 100 {
			// Sum of the percentages are greater than 100%
			ratio = excessWidth
		} else if fixedWidth == 0 {
			// No fixed width, let's take the whole excess width
			ratio = excessWidth
		} else {
			ratio = fixedWidth / (100 - percentageWidth)
		}

		widths, currentWidths, differences := make([]pr.Float, L), make([]pr.Float, L), make([]pr.Float, L)
		var sumDifferences pr.Float
		for index, tmp := range columns {
			widths[index] = columnIntrinsicPercentages[tmp.i] * ratio
			currentWidths[index] = columnWidths[tmp.i]
			// Allow to reduce the size of the columns to respect the percentage
			differences[index] = widths[index] - currentWidths[index]
			sumDifferences += differences[index]
		}

		if sumDifferences > excessWidth {
			for i, difference := range differences {
				differences[i] = difference / sumDifferences * excessWidth
			}
		}
		excessWidth -= sumDifferences
		for i, difference := range differences {
			columnWidths[columns[i].i] += difference
		}
	}
	if excessWidth <= 0 {
		return 0
	}

	// Bonus: we've tried our best to distribute the extra size, but we
	// failed. Instead of blindly distributing the size among all the colums
	// and breaking all the rules (as said in the draft), let's try to
	// change the columns with no constraint at all, then resize the table,
	// and at least break the rules to make the columns fill the table.

	// Fifth group, part 1
	columns_ = nil
	for i, column := range grid[columnSlice[0]:columnSlice[1]] {
		anyColumn, anyMaxContent := false, false
		for _, cell := range column {
			if cell != nil {
				anyColumn = true
				if maxContentWidth(context, cell, true) != 0 {
					anyMaxContent = true
				}
			}
		}
		if anyColumn && columnIntrinsicPercentages[i+columnSlice[0]] == 0 && !anyMaxContent {
			columns_ = append(columns_, i+columnSlice[0])
		}
	}
	if L := pr.Float(len(columns_)); L != 0 {
		for _, i := range columns_ {
			columnWidths[i] += excessWidth / L
		}
		return 0
	}
	// Fifth group, part 2, aka abort
	return excessWidth
}
