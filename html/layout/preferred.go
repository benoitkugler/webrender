package layout

import (
	"fmt"
	"math"
	"strings"

	"github.com/benoitkugler/webrender/html/layout/text"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/tree"
)

// Preferred and minimum preferred width, aka. max-content and min-content
// width, aka. the shrink-to-fit algorithm.

// Terms used (max-content width, min-content width) are defined in David
// Baron's unofficial draft (http://dbaron.org/css/intrinsic/).

// Return the shrink-to-fit width of ``box``.
// *Warning:* both availableOuterWidth and the return value are
// for width of the *content area*, not margin area.
// http://www.w3.org/TR/CSS21/visudet.html#float-width
func shrinkToFit(context *layoutContext, box Box, availableContentWidth pr.Float) pr.Float {
	return pr.Min(
		pr.Max(minContentWidth(context, box, false), availableContentWidth),
		maxContentWidth(context, box, false),
	)
}

// Return the min-content width for ``box``.
// This is the width by breaking at every line-break opportunity.
// outer should default to true
func minContentWidth(context *layoutContext, box Box, outer bool) pr.Float {
	if box.Box().IsTableWrapper {
		return tableAndColumnsPreferredWidths(context, box, outer).tableMinContentWidth
	} else if bo.TableCellBoxT.IsInstance(box) {
		return tableCellMinContentWidth(context, box, outer)
	} else if bo.BlockContainerBoxT.IsInstance(box) || bo.TableColumnBoxT.IsInstance(box) || bo.FlexBoxT.IsInstance(box) {
		return blockMinContentWidth(context, box, outer)
	} else if bo.TableColumnGroupBoxT.IsInstance(box) {
		return columnGroupContentWidth(box)
	} else if IsLine(box) {
		return inlineMinContentWidth(context, box, outer, nil, false, true)
	} else if rep, isReplaced := box.(bo.ReplacedBoxITF); isReplaced {
		return replacedMinContentWidth(rep.Replaced(), outer)
	} else if bo.FlexContainerBoxT.IsInstance(box) {
		return flexMinContentWidth(context, box, outer)
	} else {
		panic(fmt.Sprintf("min-content width for %T not handled yet", box))
	}
}

// Return the max-content width for ``box``.
// This is the width by only breaking at forced line breaks.
// outer=true
func maxContentWidth(context *layoutContext, box Box, outer bool) pr.Float {
	rep, isReplaced := box.(bo.ReplacedBoxITF)
	if box.Box().IsTableWrapper {
		return tableAndColumnsPreferredWidths(context, box, outer).tableMaxContentWidth
	} else if bo.TableCellBoxT.IsInstance(box) {
		return tableCellMaxContentWidth(context, box, outer)
	} else if bo.BlockContainerBoxT.IsInstance(box) || bo.TableColumnBoxT.IsInstance(box) || bo.FlexBoxT.IsInstance(box) {
		return blockMaxContentWidth(context, box, outer)
	} else if bo.TableColumnGroupBoxT.IsInstance(box) {
		return columnGroupContentWidth(box)
	} else if IsLine(box) {
		return inlineMaxContentWidth(context, box, outer, true)
	} else if isReplaced {
		return replacedMaxContentWidth(rep.Replaced(), outer)
	} else if bo.FlexContainerBoxT.IsInstance(box) {
		return flexMaxContentWidth(context, box, outer)
	} else {
		logger.WarningLogger.Printf("max-content width for %T not handled yet", box)
		return 0
	}
}

type fnBlock = func(*layoutContext, Box, bool) pr.Float

// Helper to create ``block_*_ContentWidth.``
func blockContentWidth(context *layoutContext, box Box, function fnBlock, outer bool) pr.Float {
	width := box.Box().Style.GetWidth()
	var widthValue pr.Float
	if width.String == "auto" || width.Unit == pr.Percentage {
		// "percentages on the following properties are treated instead as
		// though they were the following: width: auto"
		// http://dbaron.org/css/intrinsic/#outer-intrinsic
		var max pr.Float = 0
		for _, child := range box.Box().Children {
			if !child.Box().IsAbsolutelyPositioned() {
				v := function(context, child, true)
				if v > max {
					max = v
				}
			}
		}
		widthValue = max
	} else {
		if width.Unit != pr.Px {
			panic(fmt.Sprintf("expected Px got %d", width.Unit))
		}
		widthValue = width.Value
	}

	return adjust(box, outer, widthValue, true, true)
}

// Get box width from given width and box min- and max-widths.
func minMax(box Box, width pr.Float) pr.Float {
	minWidth := box.Box().Style.GetMinWidth()
	maxWidth := box.Box().Style.GetMaxWidth()
	var resMin, resMax pr.Float
	if minWidth.String == "auto" || minWidth.Unit == pr.Percentage {
		resMin = 0
	} else {
		resMin = minWidth.Value
	}
	if maxWidth.String == "auto" || maxWidth.Unit == pr.Percentage {
		resMax = pr.Inf
	} else {
		resMax = maxWidth.Value
	}

	if replaced, ok := box.(bo.ReplacedBoxITF); ok {
		_, _, ratio := replaced.Replaced().Replacement.GetIntrinsicSize(1, box.Box().Style.GetFontSize().Value)
		if ratio != nil {
			minHeight := box.Box().Style.GetMinHeight()
			if minHeight.String != "auto" && minHeight.Unit != pr.Percentage {
				resMin = pr.Max(resMin, minHeight.Value*ratio.V())
			}
			maxHeight := box.Box().Style.GetMaxHeight()
			if maxHeight.String != "auto" && maxHeight.Unit != pr.Percentage {
				resMax = pr.Min(resMax, maxHeight.Value*ratio.V())
			}
		}
	}

	return pr.Max(resMin, pr.Min(width, resMax))
}

// Add box paddings, borders and margins to ``width``.
// left=true, right=true
func marginWidth(box *bo.BoxFields, width pr.Float, left, right bool) pr.Float {
	var percentages pr.Float
	var cases []string
	if left {
		cases = append(cases, "margin_left", "padding_left")
	}
	if right {
		cases = append(cases, "margin_right", "padding_right")
	}
	for _, value := range cases {
		styleValue := box.Style.Get(value).(pr.Value)
		if styleValue.String != "auto" {
			switch styleValue.Unit {
			case pr.Px:
				width += styleValue.Value
			case pr.Percentage:
				percentages += styleValue.Value
			default:
				panic(fmt.Sprintf("expected Px or Percentage, got %d", styleValue.Unit))
			}
		}
	}

	if left {
		width += box.Style.GetBorderLeftWidth().Value
	}
	if right {
		width += box.Style.GetBorderRightWidth().Value
	}

	if percentages < 100 {
		return width / (1 - percentages/100.)
	} else {
		// Pathological case, ignore
		return 0
	}
}

// Respect min/max && adjust width depending on ``outer``.
// If ``outer`` is set to ``true``, return margin width, else return content
// width.
// left=true, right=true
func adjust(box Box, outer bool, width pr.Float, left, right bool) pr.Float {
	fixed := minMax(box, width)
	if outer {
		return marginWidth(box.Box(), fixed, left, right)
	} else {
		return fixed
	}
}

// Return the min-content width for a ``BlockBox``.
// outer=true
func blockMinContentWidth(context *layoutContext, box Box, outer bool) pr.Float {
	v := blockContentWidth(context, box, minContentWidth, outer)
	return v
}

// Return the max-content width for a ``BlockBox``.
// outer=true
func blockMaxContentWidth(context *layoutContext, box Box, outer bool) pr.Float {
	return blockContentWidth(context, box, maxContentWidth, outer)
}

// Return the min-content width for an ``InlineBox``.
//
// The width is calculated from the lines from ``skipStack``. If
// ``firstLine`` is ``true``, only the first line minimum width is
// calculated.
// outer=true, skipStack=None, firstLine=false, isLineStart=false
func inlineMinContentWidth(context *layoutContext, box_ Box, outer bool, skipStack tree.ResumeStack,
	firstLine, isLineStart bool) pr.Float {
	widths := inlineLineWidths(context, box_, outer, isLineStart, true, skipStack, firstLine)

	if firstLine {
		widths = widths[0:1]
	} else {
		widths[len(widths)-1] -= trailingWhitespaceSize(context, box_)
	}
	return adjust(box_, outer, pr.Maxs(widths...), true, true)
}

// Return the max-content width for an ``InlineBox``.
// outer=true, isLineStart=false
func inlineMaxContentWidth(context *layoutContext, box_ Box, outer, isLineStart bool) pr.Float {
	widths := inlineLineWidths(context, box_, outer, isLineStart, false, nil, false)
	widths[len(widths)-1] -= trailingWhitespaceSize(context, box_)
	return adjust(box_, outer, pr.Maxs(widths...), true, true)
}

// Return the *-content width for a ``TableColumnGroupBox``.
func columnGroupContentWidth(box Box) pr.Float {
	width := box.Box().Style.GetWidth()
	var width_ pr.Float
	if width.String == "auto" || width.Unit == pr.Percentage {
		width_ = 0
	} else if width.Unit == pr.Px {
		width_ = width.Value
	} else {
		panic(fmt.Sprintf("expected Px got %d", width.Unit))
	}

	return adjust(box, false, width_, true, true)
}

// Return the min-content width for a ``TableCellBox``.
func tableCellMinContentWidth(context *layoutContext, box_ Box, outer bool) pr.Float {
	box := box_.Box()
	var maxChildrenWidths pr.Float
	for _, child := range box.Children {
		if !child.Box().IsAbsolutelyPositioned() {
			v := minContentWidth(context, child, true)
			if v > maxChildrenWidths {
				maxChildrenWidths = v
			}
		}
	}
	childrenMinWidth := marginWidth(box, maxChildrenWidths, true, true)

	width := box.Style.GetWidth()
	var cellMinWidth pr.Float
	if width.String != "auto" && width.Unit == pr.Px {
		cellMinWidth = adjust(box_, outer, width.Value, true, true)
	}

	return pr.Max(childrenMinWidth, cellMinWidth)
}

// Return the max-content width for a ``TableCellBox``.
func tableCellMaxContentWidth(context *layoutContext, box Box, outer bool) pr.Float {
	return pr.Max(tableCellMinContentWidth(context, box, outer), blockMaxContentWidth(context, box, outer))
}

// firstLine=false
func inlineLineWidths(context *layoutContext, box_ Box, outer, isLineStart,
	minimum bool, skipStack tree.ResumeStack, firstLine bool) []pr.Float {
	var (
		textIndent, currentLine pr.Float
		skip                    int
		out                     []pr.Float
		box                     = box_.Box()
	)
	if bo.LineBoxT.IsInstance(box_) {
		if box.Style.GetTextIndent().Unit == pr.Percentage {
			// TODO: this is wrong, text-indent percentages should be resolved
			// before calling this function.
			textIndent = 0
		} else {
			textIndent = box.Style.GetTextIndent().Value
		}
	}
	currentLine = 0
	if skipStack != nil {
		skip, skipStack = skipStack.Unpack()
	}
	for _, child := range box.Children[skip:] {
		if child.Box().IsAbsolutelyPositioned() {
			continue // Skip
		}
		textBox, isTextBox := child.(*bo.TextBox)
		var lines []pr.Float
		if bo.InlineBoxT.IsInstance(child) {
			lines = inlineLineWidths(context, child, outer, isLineStart, minimum,
				skipStack, firstLine)
			if firstLine {
				lines = lines[0:1]
			}
			if len(lines) == 1 {
				lines[0] = adjust(child, outer, lines[0], true, true)
			} else {
				lines[0] = adjust(child, outer, lines[0], true, false)
				lines[len(lines)-1] = adjust(child, outer, lines[len(lines)-1], false, true)
			}
		} else if isTextBox {
			wp := textBox.Style.GetWhiteSpace()
			spaceCollapse := wp == "normal" || wp == "nowrap" || wp == "pre-line"
			if skipStack == nil {
				skip = 0
			} else {
				skip, skipStack = skipStack.Unpack()
				if skipStack != nil {
					panic(fmt.Sprintf("expected empty SkipStack, got %v", skipStack))
				}
			}
			childText := string([]rune(textBox.Text)[skip:])
			if isLineStart && spaceCollapse {
				childText = strings.TrimLeft(childText, " ")
			}
			if minimum && childText == " " {
				lines = []pr.Float{0, 0}
			} else {
				var maxWidth pr.MaybeFloat
				if minimum {
					maxWidth = pr.Float(0)
				}
				resumeAt := 0
				newResumeAt := 0
				textRunes := []rune(childText)
				for newResumeAt != -1 {
					resumeAt += newResumeAt
					tmp := text.SplitFirstLine(string(textRunes[resumeAt:]), textBox.Style, context, maxWidth, textBox.JustificationSpacing, true)
					newResumeAt = tmp.ResumeAt
					lines = append(lines, tmp.Width)
					if firstLine {
						break
					}
				}
				if firstLine && newResumeAt != -1 && newResumeAt != 0 {
					currentLine += lines[0]
					break
				}
			}
		} else {
			// http://www.w3.org/TR/css3-text/#line-break-details
			// "The line breaking behavior of a replaced element
			//  or other atomic inline is equivalent to that
			//  of the Object Replacement Character (U+FFFC)."
			// http://www.unicode.org/reports/tr14/#DescriptionOfProperties
			// "By default, there is a break opportunity
			//  both before and after any inline object."
			if minimum {
				lines = []pr.Float{0, maxContentWidth(context, child, true), 0}
			} else {
				lines = []pr.Float{maxContentWidth(context, child, true)}
			}
		}
		// The first text line goes on the current line
		currentLine += lines[0]
		if len(lines) > 1 {
			// Forced line break
			out = append(out, currentLine+textIndent)
			textIndent = 0
			if len(lines) > 2 {
				out = append(out, lines[1:len(lines)-1]...)
			}
			currentLine = lines[len(lines)-1]
		}
		isLineStart = lines[len(lines)-1] == 0
		skipStack = nil
	}
	out = append(out, currentLine+textIndent)
	return out
}

// Return the percentage contribution of a cell, column or column group.
// http://dbaron.org/css/intrinsic/#pct-contrib
func percentageContribution(box bo.BoxFields) pr.Float {
	var minWidth, width pr.Float
	maxWidth := pr.Inf
	miw, maw, w := box.Style.GetMinWidth(), box.Style.GetMaxWidth(), box.Style.GetWidth()
	if miw.String != "auto" && miw.Unit == pr.Percentage {
		minWidth = miw.Value
	}
	if maw.String != "auto" && maw.Unit == pr.Percentage {
		maxWidth = maw.Value
	}
	if w.String != "auto" && w.Unit == pr.Percentage {
		width = w.Value
	}
	return pr.Max(minWidth, pr.Min(width, maxWidth))
}

type tableContentWidths struct {
	grid                         [][]Box
	columnMinContentWidths       []pr.Float
	columnMaxContentWidths       []pr.Float
	columnIntrinsicPercentages   []pr.Float
	constrainedness              []bool
	tableMinContentWidth         pr.Float
	totalHorizontalBorderSpacing pr.Float
	tableMaxContentWidth         pr.Float
}

// Return content widths for the auto layout table and its columns.
// http://dbaron.org/css/intrinsic/
// outer = true
func tableAndColumnsPreferredWidths(context *layoutContext, box_ Box, outer bool) tableContentWidths {
	box := box_.Box()
	table_ := box.GetWrappedTable()
	table := table_.Table()

	if result := context.tables[table]; result != nil {
		return result[outer]
	}

	// Create the grid
	var gridWidth, gridHeight int
	rowNumber := 0
	for _, rowGroup := range table.Children {
		for _, row := range rowGroup.Box().Children {
			for _, cell := range row.Box().Children {
				gridWidth = utils.MaxInt(cell.Box().GridX+cell.Box().Colspan, gridWidth)
				gridHeight = utils.MaxInt(rowNumber+cell.Box().Rowspan, gridHeight)
			}
			rowNumber += 1
		}
	}
	// zippedGrid = list(zip(*grid)), which is the transpose of grid
	grid, zippedGrid := make([][]Box, gridHeight), make([][]Box, gridWidth)
	for i := range grid {
		grid[i] = make([]Box, gridWidth)
	}
	for j := range zippedGrid {
		zippedGrid[j] = make([]Box, gridHeight)
	}
	rowNumber = 0
	for _, rowGroup := range table.Children {
		for _, row := range rowGroup.Box().Children {
			for _, cell := range row.Box().Children {
				grid[rowNumber][cell.Box().GridX] = cell
				zippedGrid[cell.Box().GridX][rowNumber] = cell
			}
			rowNumber += 1
		}
	}

	// Define the total horizontal border spacing
	var totalHorizontalBorderSpacing pr.Float
	if table.Style.GetBorderCollapse() == "separate" && gridWidth > 0 {
		var tot pr.Float = 1
		for _, column := range zippedGrid {
			any := false
			for _, b := range column {
				if b != nil {
					any = true
					break
				}
			}
			if any {
				tot += 1
			}
		}
		totalHorizontalBorderSpacing = table.Style.GetBorderSpacing()[0].Value * tot
	}

	if gridWidth == 0 || gridHeight == 0 {
		table.Children = nil
		minWidth := blockMinContentWidth(context, table_, false)
		maxWidth := blockMaxContentWidth(context, table_, false)
		outerMinWidth := adjust(box_, true, blockMinContentWidth(context, table_, true), true, true)
		outerMaxWidth := adjust(box_, true, blockMaxContentWidth(context, table_, true), true, true)
		context.tables[table] = map[bool]tableContentWidths{
			false: {
				tableMinContentWidth:         minWidth,
				tableMaxContentWidth:         maxWidth,
				totalHorizontalBorderSpacing: totalHorizontalBorderSpacing,
			},
			true: {
				tableMinContentWidth:         outerMinWidth,
				tableMaxContentWidth:         outerMaxWidth,
				totalHorizontalBorderSpacing: totalHorizontalBorderSpacing,
			},
		}
		return context.tables[table][outer]
	}

	columnGroups := make([]Box, gridWidth)
	columns := make([]Box, gridWidth)
	columnNumber := 0
outerLoop:
	for _, columnGroup := range table.ColumnGroups {
		for _, column := range columnGroup.Box().Children {
			columnGroups[columnNumber] = columnGroup
			columns[columnNumber] = column
			columnNumber += 1
			if columnNumber == gridWidth {
				break outerLoop
			}
		}
	}

	var colspanCells []Box

	// Define the intermediate content widths
	minContentWidths := make([]pr.Float, gridWidth)
	maxContentWidths := make([]pr.Float, gridWidth)
	intrinsicPercentages := make([]pr.Float, gridWidth)

	groupss := [2]*[]Box{&columnGroups, &columns}

	// Intermediate content widths for span 1
	for i := range minContentWidths {
		for _, groups := range groupss {
			if b := (*groups)[i]; b != nil {
				minContentWidths[i] = pr.Max(minContentWidths[i], minContentWidth(context, b, true))
				maxContentWidths[i] = pr.Max(maxContentWidths[i], maxContentWidth(context, b, true))
				intrinsicPercentages[i] = pr.Max(intrinsicPercentages[i], percentageContribution(*b.Box()))
			}
		}
		for _, cell := range zippedGrid[i] {
			if cell != nil {
				if cell.Box().Colspan == 1 {
					minContentWidths[i] = pr.Max(minContentWidths[i], minContentWidth(context, cell, true))
					maxContentWidths[i] = pr.Max(maxContentWidths[i], maxContentWidth(context, cell, true))
					intrinsicPercentages[i] = pr.Max(intrinsicPercentages[i], percentageContribution(*cell.Box()))
				} else {
					colspanCells = append(colspanCells, cell)
				}
			}
		}
	}

	// Intermediate content widths for span > 1 is wrong in the 4.1 section, as
	// explained in its third issue. Min- and max-content widths are handled by
	// the excess width distribution method, and percentages do not distribute
	// widths to columns that have originating cells.

	// Intermediate intrinsic percentage widths for span > 1
	for span := 1; span < gridWidth; span += 1 {
		var percentageContributions []pr.Float
		for i, percentageContribution_ := range intrinsicPercentages {
			for j := range zippedGrid[i] {
				var indexes []int
				for k := 0; k < i+1; k += 1 {
					if grid[j][k] != nil {
						indexes = append(indexes, k)
					}
				}
				if len(indexes) == 0 {
					continue
				}
				origin := indexes[len(indexes)-1] // = max
				originCell := grid[j][origin]
				ocColspan := originCell.Box().Colspan
				if ocColspan-1 != span {
					continue
				}
				var baselinePercentage pr.Float
				for u := origin; u < origin+ocColspan; u += 1 {
					baselinePercentage += intrinsicPercentages[u]
				}

				// Cell contribution to intrinsic percentage width
				if intrinsicPercentages[i] == 0 {
					diff := pr.Max(0, percentageContribution(*originCell.Box())-baselinePercentage)
					var (
						otherColumnsContributions    []pr.Float
						otherColumnsContributionsSum pr.Float
					)
					for s := origin; s < origin+ocColspan; s += 1 {
						if intrinsicPercentages[s] == 0 {
							otherColumnsContributions = append(otherColumnsContributions, maxContentWidths[s])
							otherColumnsContributionsSum += maxContentWidths[s]
						}
					}
					var ratio pr.Float
					if otherColumnsContributionsSum == 0 {
						if len(otherColumnsContributions) > 0 {
							ratio = 1. / pr.Float(len(otherColumnsContributions))
						} else {
							ratio = 1
						}
					} else {
						ratio = maxContentWidths[i] / otherColumnsContributionsSum
					}
					percentageContribution_ = pr.Max(percentageContribution_, diff*ratio)
				}
			}
			percentageContributions = append(percentageContributions, percentageContribution_)
		}
		intrinsicPercentages = percentageContributions
	}

	// Define constrainedness
	constrainedness := make([]bool, gridWidth)
	for i := range constrainedness {
		if columnGroups[i] != nil {
			if wid := columnGroups[i].Box().Style.GetWidth(); wid.String != "auto" && wid.Unit != pr.Percentage {
				constrainedness[i] = true
				continue
			}
		}
		if columns[i] != nil {
			if wid := columns[i].Box().Style.GetWidth(); wid.String != "auto" && wid.Unit != pr.Percentage {
				constrainedness[i] = true
				continue
			}
		}
		for _, cell := range zippedGrid[i] {
			if cell != nil {
				if wid := cell.Box().Style.GetWidth(); cell.Box().Colspan == 1 && wid.String != "auto" && wid.Unit != pr.Percentage {
					constrainedness[i] = true
					break
				}
			}
		}
	}
	var cumsum pr.Float
	for i, percentage := range intrinsicPercentages {
		u := pr.Min(percentage, 100-cumsum)
		cumsum += percentage
		intrinsicPercentages[i] = u
	}

	// Max- and min-content widths for span > 1
	for _, cell_ := range colspanCells {
		cell := cell_.Box()
		minContent := minContentWidth(context, cell_, true)
		maxContent := maxContentWidth(context, cell_, true)
		columnSlice := [2]int{cell.GridX, cell.GridX + cell.Colspan}
		var columnsMinContent, columnsMaxContent pr.Float
		for s := columnSlice[0]; s < columnSlice[1]; s += 1 {
			columnsMinContent += minContentWidths[s]
			columnsMaxContent += maxContentWidths[s]
		}
		var spacing pr.Float
		if table.Style.GetBorderCollapse() == "separate" {
			spacing = pr.Float(cell.Colspan-1) * table.Style.GetBorderSpacing()[0].Value
		}

		if minContent > columnsMinContent+spacing {
			excessWidth := minContent - (columnsMinContent + spacing)
			distributeExcessWidth(context, zippedGrid, excessWidth, minContentWidths,
				constrainedness, intrinsicPercentages, maxContentWidths, columnSlice)
		}

		if maxContent > columnsMaxContent+spacing {
			excessWidth := maxContent - (columnsMaxContent + spacing)
			distributeExcessWidth(context, zippedGrid, excessWidth, maxContentWidths,
				constrainedness, intrinsicPercentages, maxContentWidths, columnSlice)
		}
	}

	// Calculate the max- and min-content widths of table and columns
	var (
		smallpercentageContributions               []pr.Float
		largepercentageContributionNumerator, sum_ pr.Float
	)
	for i, v := range intrinsicPercentages {
		sum_ += v
		if v != 0 {
			smallpercentageContributions = append(smallpercentageContributions, maxContentWidths[i]/(v/100.))
		} else {
			largepercentageContributionNumerator += maxContentWidths[i]
		}
	}
	largepercentageContributionDenominator := (100. - sum_) / 100.
	var largepercentageContribution pr.Float
	if largepercentageContributionDenominator == 0 {
		if largepercentageContributionNumerator == 0 {
			largepercentageContribution = 0
		} else {
			// "the large percentage contribution of the table [is] an
			// infinitely large number if the numerator is nonzero [and] the
			// denominator of that ratio is 0."
			//
			// http://dbaron.org/css/intrinsic/#autotableintrinsic
			//
			// Please note that "an infinitely large number" is not "infinite",
			// and that"s probably not a coincindence: putting "inf" here breaks
			// some cases (see #305).
			largepercentageContribution = math.MaxInt32
		}
	} else {
		largepercentageContribution = largepercentageContributionNumerator / largepercentageContributionDenominator
	}

	var sumMin, sumMax pr.Float
	for i := range minContentWidths {
		sumMin += minContentWidths[i]
		sumMax += maxContentWidths[i]
	}
	tableMinContentWidth := totalHorizontalBorderSpacing + sumMin
	tableMaxContentWidth := totalHorizontalBorderSpacing + pr.Max(
		pr.Max(sumMax, largepercentageContribution),
		pr.Maxs(smallpercentageContributions...))

	tableMinWidth := tableMinContentWidth
	tableMaxWidth := tableMaxContentWidth
	if wid := table.Style.GetWidth(); wid.String != "auto" && wid.Unit == pr.Px {
		// "percentages on the following properties are treated instead as
		// though they were the following: width: auto"
		// http://dbaron.org/css/intrinsic/#outer-intrinsic
		tableMinWidth = wid.Value
		tableMaxWidth = wid.Value
	}

	tableMinContentWidth = pr.Max(tableMinContentWidth, adjust(table, false, tableMinWidth, true, true))
	tableMaxContentWidth = pr.Max(tableMaxContentWidth, adjust(table, false, tableMaxWidth, true, true))
	tableOuterMinContentWidth := marginWidth(&table.BoxFields, marginWidth(box, tableMinContentWidth, true, true), true, true)
	tableOuterMaxContentWidth := marginWidth(&table.BoxFields, marginWidth(box, tableMaxContentWidth, true, true), true, true)

	result := tableContentWidths{
		columnMinContentWidths:       minContentWidths,
		columnMaxContentWidths:       maxContentWidths,
		columnIntrinsicPercentages:   intrinsicPercentages,
		constrainedness:              constrainedness,
		totalHorizontalBorderSpacing: totalHorizontalBorderSpacing,
		grid:                         zippedGrid,
	}
	resultFalse := result
	resultFalse.tableMinContentWidth = tableMinContentWidth
	resultFalse.tableMaxContentWidth = tableMaxContentWidth
	resultTrue := result
	resultTrue.tableMinContentWidth = tableOuterMinContentWidth
	resultTrue.tableMaxContentWidth = tableOuterMaxContentWidth

	context.tables[table] = map[bool]tableContentWidths{
		false: resultFalse,
		true:  resultTrue,
	}
	return context.tables[table][outer]
}

// Return the min-content width for an ``InlineReplacedBox``.
// outer=true
func replacedMinContentWidth(box *bo.ReplacedBox, outer bool) pr.Float {
	width := box.Style.GetWidth()
	var (
		h pr.MaybeFloat
		w pr.Float
	)
	if width.String == "auto" {
		height := box.Style.GetHeight()
		if height.String == "auto" || height.Unit == pr.Percentage {
			h = pr.Auto
		} else if height.Unit == pr.Px {
			h = height.Value
		} else {
			panic(fmt.Sprintf("expected Px got %d", height.Unit))
		}
		if mw := box.Style.GetMaxWidth(); mw.String != "auto" && mw.Unit == pr.Percentage {
			// See https://drafts.csswg.org/css-sizing/#intrinsic-contribution
			w = pr.Float(0)
		} else {
			image := box.Replacement
			iwidth, iheight, ratio := image.GetIntrinsicSize(box.Style.GetImageResolution().Value, box.Style.GetFontSize().Value)
			w, _ = defaultImageSizing(iwidth, iheight, ratio, pr.Auto, h, 300, 150)
		}
	} else if width.Unit == pr.Percentage {
		// See https://drafts.csswg.org/css-sizing/#intrinsic-contribution
		w = 0
	} else if width.Unit == pr.Px {
		w = width.Value
	} else {
		panic(fmt.Sprintf("expected Px got %d", width.Unit))
	}
	return adjust(box, outer, w, true, true)
}

// Return the max-content width for an ``InlineReplacedBox``.
func replacedMaxContentWidth(box *bo.ReplacedBox, outer bool) pr.Float {
	width := box.Style.GetWidth()
	var (
		h pr.MaybeFloat
		w pr.Float
	)
	if width.String == "auto" {
		height := box.Style.GetHeight()
		if height.String == "auto" || height.Unit == pr.Percentage {
			h = pr.Auto
		} else if height.Unit == pr.Px {
			h = height.Value
		} else {
			panic(fmt.Sprintf("expected Px got %d", height.Unit))
		}

		image := box.Replacement
		iwidth, iheight, ratio := image.GetIntrinsicSize(box.Style.GetImageResolution().Value, box.Style.GetFontSize().Value)
		w, _ = defaultImageSizing(iwidth, iheight, ratio, pr.Auto, h, 300, 150)

	} else if width.Unit == pr.Percentage {
		// See https://drafts.csswg.org/css-sizing/#intrinsic-contribution
		w = 0
	} else if width.Unit == pr.Px {
		w = width.Value
	} else {
		panic(fmt.Sprintf("expected Px got %d", width.Unit))
	}
	return adjust(box, outer, w, true, true)
}

// Return the min-content width for an ``FlexContainerBox``.
// outer=true
func flexMinContentWidth(context *layoutContext, box Box, outer bool) pr.Float {
	// TODO: use real values, see
	// https://www.w3.org/TR/css-flexbox-1/#intrinsic-sizes
	var sum, max pr.Float
	for _, child := range box.Box().Children {
		if child.Box().IsFlexItem {
			v := minContentWidth(context, child, true)
			sum += v
			if v > max {
				max = v
			}
		}
	}
	if strings.HasPrefix(string(box.Box().Style.GetFlexDirection()), "row") && box.Box().Style.GetFlexWrap() == "nowrap" {
		return adjust(box, outer, sum, true, true)
	} else {
		return adjust(box, outer, max, true, true)
	}
}

// Return the max-content width for an ``FlexContainerBox``.
func flexMaxContentWidth(context *layoutContext, box Box, outer bool) pr.Float {
	// TODO: use real values, see
	// https://www.w3.org/TR/css-flexbox-1/#intrinsic-sizes
	var sum, max pr.Float
	for _, child := range box.Box().Children {
		if child.Box().IsFlexItem {
			v := maxContentWidth(context, child, true)
			sum += v
			if v > max {
				max = v
			}
		}
	}
	if strings.HasPrefix(string(box.Box().Style.GetFlexDirection()), "row") && box.Box().Style.GetFlexWrap() == "nowrap" {
		return adjust(box, outer, sum, true, true)
	} else {
		return adjust(box, outer, max, true, true)
	}
}

// Return the size of the trailing whitespace of ``box``.
func trailingWhitespaceSize(context *layoutContext, box Box) pr.Float {
	for IsLine(box) {
		ch := box.Box().Children
		if len(ch) == 0 {
			return 0
		}
		box = ch[len(ch)-1]
	}
	textBox, ok := box.(*bo.TextBox)
	if ws := box.Box().Style.GetWhiteSpace(); !(ok && textBox.Text != "" &&
		(ws == "normal" || ws == "nowrap" || ws == "pre-line")) {
		return 0
	}
	strippedText := strings.TrimRight(textBox.Text, " ")
	if box.Box().Style.GetFontSize() == pr.FToV(0) || len(strippedText) == len(textBox.Text) {
		return 0
	}
	if len(strippedText) != 0 {
		resume, oldResume := 0, -1
		var oldBox *bo.TextBox
		for resume != -1 {
			oldResume = resume
			oldBox, resume, _ = splitTextBox(context, textBox, nil, resume)
		}
		if oldBox == nil {
			panic("oldBox can't be nil")
		}
		strippedBox := textBox.CopyWithText(strippedText)
		strippedBox, resume, _ = splitTextBox(context, strippedBox, nil, oldResume)
		if strippedBox == nil || resume != -1 {
			panic("invalid strippedBox or resume")
		}
		return oldBox.Box().Width.V() - strippedBox.Box().Width.V()
	} else {
		spli := text.SplitFirstLine(textBox.Text, textBox.Style, context, nil, textBox.JustificationSpacing, false)
		return spli.Width
	}
}
