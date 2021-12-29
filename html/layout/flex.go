package layout

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/benoitkugler/webrender/html/tree"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
)

// Layout for flex containers && flex-items.

type indexedBox struct {
	box   Box
	index int
}

type flexLine struct {
	line                     []indexedBox
	crossSize, lowerBaseline pr.Float
}

func (f flexLine) reverse() {
	for left, right := 0, len(f.line)-1; left < right; left, right = left+1, right-1 {
		f.line[left], f.line[right] = f.line[right], f.line[left]
	}
}

func (f flexLine) sum() pr.Float {
	var sum pr.Float
	for _, child := range f.line {
		sum += child.box.Box().HypotheticalMainSize
	}
	return sum
}

func (f flexLine) allFrozen() bool {
	for _, child := range f.line {
		if !child.box.Box().Frozen {
			return false
		}
	}
	return true
}

func (f flexLine) adjustements() pr.Float {
	var sum pr.Float
	for _, child := range f.line {
		sum += child.box.Box().Adjustment
	}
	return sum
}

func reverse(f []flexLine) {
	for left, right := 0, len(f)-1; left < right; left, right = left+1, right-1 {
		f[left], f[right] = f[right], f[left]
	}
}

func sumCross(f []flexLine) pr.Float {
	var sumCross pr.Float
	for _, line := range f {
		sumCross += line.crossSize
	}
	return sumCross
}

func getAttr(box *bo.BoxFields, axis, min string) pr.MaybeFloat {
	var boxAxis pr.MaybeFloat
	if axis == "width" {
		boxAxis = box.Width
		if min == "min" {
			boxAxis = box.MinWidth
		} else if min == "max" {
			boxAxis = box.MaxWidth
		}
	} else {
		boxAxis = box.Height
		if min == "min" {
			boxAxis = box.MinHeight
		} else if min == "max" {
			boxAxis = box.MaxHeight
		}
	}
	return boxAxis
}

func getCrossMargins(child *bo.BoxFields, cross string) bo.MaybePoint {
	crossMargins := bo.MaybePoint{child.MarginLeft, child.MarginRight}
	if cross == "height" {
		crossMargins = bo.MaybePoint{child.MarginTop, child.MarginBottom}
	}
	return crossMargins
}

func getCross(box *bo.BoxFields, cross string) pr.Value {
	out, _ := box.Style.Get(cross).(pr.Value)
	return out
}

func setDirection(box *bo.BoxFields, position string, value pr.Float) {
	if position == "positionX" {
		box.PositionX = value
	} else {
		box.PositionY = value
	}
}

// the returned box as same concrete type than box_
func flexLayout(context *layoutContext, box_ Box, maxPositionY pr.Float, skipStack tree.ResumeStack, containingBlock containingBlock,
	pageIsEmpty bool, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder) (bo.Box, blockLayout) {

	context.createBlockFormattingContext()
	var resumeAt tree.ResumeStack
	box := box_.Box()
	// Step 1 is done in formattingStructure.Boxes
	// Step 2
	axis, cross := "height", "width"
	if strings.HasPrefix(string(box.Style.GetFlexDirection()), "row") {
		axis, cross = "width", "height"
	}

	var marginLeft pr.Float
	if box.MarginLeft != pr.Auto {
		marginLeft = box.MarginLeft.V()
	}
	var marginRight pr.Float
	if box.MarginRight != pr.Auto {
		marginRight = box.MarginRight.V()
	}
	var marginTop pr.Float
	if box.MarginTop != pr.Auto {
		marginTop = box.MarginTop.V()
	}
	var marginBottom pr.Float
	if box.MarginBottom != pr.Auto {
		marginBottom = box.MarginBottom.V()
	}
	var availableMainSpace pr.Float
	cbWidth, cbHeight := containingBlock.ContainingBlock()
	boxAxis := getAttr(box, axis, "")
	if boxAxis != pr.Auto {
		availableMainSpace = boxAxis.V()
	} else {
		if axis == "width" {
			availableMainSpace = cbWidth.V() - marginLeft - marginRight -
				box.PaddingLeft.V() - box.PaddingRight.V() - box.BorderLeftWidth.V() - box.BorderRightWidth.V()
		} else {
			mainSpace := maxPositionY - box.PositionY
			if cbHeight != pr.Auto {
				// if hasattr(cbHeight, "unit") {
				//     assert cbHeight.unit == "px"
				//     mainSpace = min(mainSpace, cbHeight.value)
				mainSpace = pr.Min(mainSpace, cbHeight.V())
			}
			availableMainSpace = mainSpace - marginTop - marginBottom -
				box.PaddingTop.V() - box.PaddingBottom.V() - box.BorderTopWidth.V() - box.BorderBottomWidth.V()
		}
	}
	var availableCrossSpace pr.Float
	boxCross := getAttr(box, cross, "")
	if boxCross != pr.Auto {
		availableCrossSpace = boxCross.V()
	} else {
		if cross == "height" {
			mainSpace := maxPositionY - box.ContentBoxY()
			if he := cbHeight; he != pr.Auto {
				mainSpace = pr.Min(mainSpace, he.V())
			}
			availableCrossSpace = mainSpace - marginTop - marginBottom -
				box.PaddingTop.V() - box.PaddingBottom.V() - box.BorderTopWidth.V() - box.BorderBottomWidth.V()
		} else {
			availableCrossSpace = cbWidth.V() - marginLeft - marginRight -
				box.PaddingLeft.V() - box.PaddingRight.V() - box.BorderLeftWidth.V() - box.BorderRightWidth.V()
		}
	}

	// Step 3
	children := box.Children
	parentBox_ := bo.CopyWithChildren(box_, children)
	parentBox := parentBox_.Box()
	resolvePercentagesBox(parentBox_, containingBlock, "")

	if parentBox.MarginTop == pr.Auto {
		box.MarginTop = pr.Float(0)
		parentBox.MarginTop = pr.Float(0)
	}
	if parentBox.MarginBottom == pr.Auto {
		box.MarginBottom = pr.Float(0)
		parentBox.MarginBottom = pr.Float(0)
	}
	if parentBox.MarginLeft == pr.Auto {
		box.MarginLeft = pr.Float(0)
		parentBox.MarginLeft = pr.Float(0)
	}
	if parentBox.MarginRight == pr.Auto {
		box.MarginRight = pr.Float(0)
		parentBox.MarginRight = pr.Float(0)
	}
	if bo.FlexBoxT.IsInstance(parentBox_) {
		blockLevelWidth(parentBox_, nil, containingBlock)
	} else {
		parentBox.Width = flexMaxContentWidth(context, parentBox_, true)
	}
	originalSkipStack := skipStack
	if skipStack != nil {
		var index int
		index, skipStack = skipStack.Unpack()
		if strings.HasSuffix(string(box.Style.GetFlexDirection()), "-reverse") {
			children = children[:index+1]
		} else {
			children = children[index:]
		}
	} else {
		skipStack = nil
	}

	childSkipStack := skipStack
	for _, child_ := range children {
		child := child_.Box()
		if !child.IsFlexItem {
			continue
		}

		// See https://www.W3.org/TR/css-flexbox-1/#min-size-auto

		mainFlexDirection := ""
		if child.Style.GetOverflow() == "visible" {
			mainFlexDirection = axis
		}

		resolvePercentagesBox(child_, containingBlock, mainFlexDirection)
		child.PositionX = parentBox.ContentBoxX()
		child.PositionY = parentBox.ContentBoxY()
		if child.MinWidth == pr.Auto {
			specifiedSize := pr.Inf
			if child.Width != pr.Auto {
				specifiedSize = child.Width.V()
			}
			newChild := child_.Copy()
			if bo.ParentBoxT.IsInstance(child_) {
				newChild = bo.CopyWithChildren(child_, child.Children)
			}
			newChild.Box().Style = child.Style.Copy()
			newChild.Box().Style.SetWidth(pr.SToV("auto"))
			newChild.Box().Style.SetMinWidth(pr.ZeroPixels.ToValue())
			newChild.Box().Style.SetMaxWidth(pr.Dimension{Value: pr.Inf, Unit: pr.Px}.ToValue())
			contentSize := minContentWidth(context, newChild, false)
			child.MinWidth = pr.Min(specifiedSize, contentSize)
		} else if child.MinHeight == pr.Auto {
			specifiedSize := pr.Inf
			if child.Height != pr.Auto {
				specifiedSize = child.Height.V()
			}
			newChild := child_.Copy()
			if bo.ParentBoxT.IsInstance(child_) {
				newChild = bo.CopyWithChildren(child_, child.Children)
			}
			newChild.Box().Style = child.Style.Copy()
			newChild.Box().Style.SetHeight(pr.SToV("auto"))
			newChild.Box().Style.SetMinHeight(pr.ZeroPixels.ToValue())
			newChild.Box().Style.SetMaxHeight(pr.Dimension{Value: pr.Inf, Unit: pr.Px}.ToValue())
			newChild, _ = blockLevelLayout(context, newChild.(bo.BlockLevelBoxITF),
				pr.Inf, childSkipStack, parentBox, pageIsEmpty, nil, nil, nil, false)
			contentSize := newChild.Box().Height.V()
			child.MinHeight = pr.Min(specifiedSize, contentSize)
		}

		child.Style = child.Style.Copy()
		var flexBasis pr.Value
		if child.Style.GetFlexBasis().String == "content" {
			flexBasis = pr.SToV("content")
			child.FlexBasis = flexBasis
		} else {
			child.FlexBasis = pr.MaybeFloatToValue(resolveOnePercentage(child.Style.GetFlexBasis(), "flex_basis", availableMainSpace, ""))
			flexBasis = child.FlexBasis
		}

		// "If a value would resolve to auto for width, it instead resolves
		// to content for flex-basis." Let's do this for height too.
		// See https://www.W3.org/TR/css-flexbox-1/#propdef-flex-basis
		target, val := &child.Height, child.Style.GetHeight()
		if axis == "width" {
			target, val = &child.Width, child.Style.GetWidth()
		}
		*target = resolveOnePercentage(val, axis, availableMainSpace, "")
		if flexBasis.String == "auto" {
			if getCross(child, axis).String == "auto" {
				flexBasis = pr.SToV("content")
			} else {
				if axis == "width" {
					flexBasis_ := child.BorderWidth()
					if child.MarginLeft != pr.Auto {
						flexBasis_ += child.MarginLeft.V()
					}
					if child.MarginRight != pr.Auto {
						flexBasis_ += child.MarginRight.V()
					}
					flexBasis = flexBasis_.ToValue()
				} else {
					flexBasis_ := child.BorderHeight()
					if child.MarginTop != pr.Auto {
						flexBasis_ += child.MarginTop.V()
					}
					if child.MarginBottom != pr.Auto {
						flexBasis_ += child.MarginBottom.V()
					}
					flexBasis = flexBasis_.ToValue()
				}
			}
		}

		// Step 3.A
		if flexBasis.String != "content" {
			child.FlexBaseSize = flexBasis.Value

			// TODO: Step 3.B
			// TODO: Step 3.C

			// Step 3.D is useless, as we never have infinite sizes on paged media

			// Step 3.E
		} else {
			child.Style.Set(axis, pr.SToV("max-content"))
			styleAxis := child.Style.Get(axis).(pr.Value)
			// TODO: don"t set style value, support *-content values instead
			if styleAxis.String == "max-content" {
				child.Style.Set(axis, pr.SToV("auto"))
				if axis == "width" {
					child.FlexBaseSize = maxContentWidth(context, child_, true)
				} else {
					newChild := child_.Copy()
					if bo.ParentBoxT.IsInstance(child_) {
						newChild = bo.CopyWithChildren(child_, child.Children)
					}
					newChild.Box().Width = pr.Inf
					newChild, _ = blockLevelLayout(context, newChild.(bo.BlockLevelBoxITF), pr.Inf, childSkipStack,
						parentBox, pageIsEmpty, absoluteBoxes, fixedBoxes, new([]pr.Float), false)
					child.FlexBaseSize = newChild.Box().MarginHeight()
				}
			} else if styleAxis.String == "min-content" {
				child.Style.Set(axis, pr.SToV("auto"))
				if axis == "width" {
					child.FlexBaseSize = minContentWidth(context, child_, true)
				} else {
					newChild := child_.Copy()
					if bo.ParentBoxT.IsInstance(child_) {
						newChild = bo.CopyWithChildren(child_, child.Children)
					}
					newChild.Box().Width = pr.Float(0)
					newChild, _ = blockLevelLayout(context, newChild.(bo.BlockLevelBoxITF), pr.Inf, childSkipStack,
						parentBox, pageIsEmpty, absoluteBoxes, fixedBoxes, nil, false)
					child.FlexBaseSize = newChild.Box().MarginHeight()
				}
			} else if styleAxis.Unit == pr.Px {
				// TODO: should we add padding, borders and margins?
				child.FlexBaseSize = styleAxis.Value
			} else {
				panic(fmt.Sprintf("unexpected Style[axis] : %v", styleAxis))
			}
		}
		if axis == "width" {
			child.HypotheticalMainSize = pr.Max(child.MinWidth.V(), pr.Min(child.FlexBaseSize, child.MaxWidth.V()))
		} else {
			child.HypotheticalMainSize = pr.Max(child.MinHeight.V(), pr.Min(child.FlexBaseSize, child.MaxHeight.V()))
		}

		// Skip stack is only for the first child
		childSkipStack = nil
	}

	// Step 4
	// TODO: the whole step has to be fixed
	if axis == "width" {
		blockLevelWidth(box_, nil, containingBlock)
	} else {
		if he := box.Style.GetHeight(); he.String != "auto" {
			box.Height = he.Value
		} else {
			box.Height = pr.Float(0)
			for i, child_ := range children {
				child := child_.Box()
				if !child.IsFlexItem {
					continue
				}
				childHeight := child.HypotheticalMainSize + child.BorderTopWidth.V() + child.BorderBottomWidth.V() +
					child.PaddingTop.V() + child.PaddingBottom.V()
				if getAttr(box, axis, "") == pr.Auto && childHeight+box.Height.V() > availableMainSpace {
					resumeAt = tree.ResumeStack{i: nil}
					children = children[:i+1]
					break
				}
				box.Height = box.Height.V() + childHeight
			}
		}
	}

	// Step 5
	var flexLines []flexLine

	var line flexLine
	var lineSize pr.Float
	axisSize := getAttr(box, axis, "")

	sortedChildren := append([]Box{}, children...)
	sort.Slice(sortedChildren, func(i, j int) bool {
		return sortedChildren[i].Box().Style.GetOrder() < sortedChildren[j].Box().Style.GetOrder()
	})
	for i, child_ := range sortedChildren {
		child := child_.Box()
		if !child.IsFlexItem {
			continue
		}
		lineSize += child.HypotheticalMainSize
		if box.Style.GetFlexWrap() != "nowrap" && lineSize > axisSize.V() {
			if len(line.line) != 0 {
				flexLines = append(flexLines, line)
				line = flexLine{line: []indexedBox{{index: i, box: child_}}}
				lineSize = child.HypotheticalMainSize
			} else {
				line.line = append(line.line, indexedBox{index: i, box: child_})
				flexLines = append(flexLines, line)
				line.line = nil
				lineSize = 0
			}
		} else {
			line.line = append(line.line, indexedBox{index: i, box: child_})
		}
	}
	if len(line.line) != 0 {
		flexLines = append(flexLines, line)
	}

	// TODO: handle *-reverse using the terminology from the specification
	if box.Style.GetFlexWrap() == "wrap-reverse" {
		reverse(flexLines)
	}
	if strings.HasSuffix(string(box.Style.GetFlexDirection()), "-reverse") {
		for _, line := range flexLines {
			line.reverse()
		}
	}

	// Step 6
	// See https://www.W3.org/TR/css-flexbox-1/#resolve-flexible-lengths
	for _, line := range flexLines {
		// Step 6 - 9.7.1
		hypotheticalMainSize := line.sum()
		flexFactorType := "shrink"
		if hypotheticalMainSize < availableMainSpace {
			flexFactorType = "grow"
		}

		// Step 6 - 9.7.2
		for _, v := range line.line {
			child := v.box.Box()
			if flexFactorType == "grow" {
				child.FlexFactor = child.Style.GetFlexGrow()
			} else {
				child.FlexFactor = child.Style.GetFlexShrink()
			}
			if child.FlexFactor == 0 ||
				(flexFactorType == "grow" && child.FlexBaseSize > child.HypotheticalMainSize) ||
				(flexFactorType == "shrink" && child.FlexBaseSize < child.HypotheticalMainSize) {
				child.TargetMainSize = child.HypotheticalMainSize
				child.Frozen = true
			} else {
				child.Frozen = false
			}
		}

		// Step 6 - 9.7.3
		initialFreeSpace := availableMainSpace
		for _, v := range line.line {
			child := v.box.Box()
			if child.Frozen {
				initialFreeSpace -= child.TargetMainSize
			} else {
				initialFreeSpace -= child.FlexBaseSize
			}
		}

		// Step 6 - 9.7.4
		for !line.allFrozen() {
			var unfrozenFactorSum pr.Float
			remainingFreeSpace := availableMainSpace

			// Step 6 - 9.7.4.B
			for _, v := range line.line {
				child := v.box.Box()
				if child.Frozen {
					remainingFreeSpace -= child.TargetMainSize
				} else {
					remainingFreeSpace -= child.FlexBaseSize
					unfrozenFactorSum += child.FlexFactor
				}
			}

			if unfrozenFactorSum < 1 {
				initialFreeSpace *= unfrozenFactorSum
			}

			if initialFreeSpace == pr.Inf {
				initialFreeSpace = math.MaxInt32
			}
			if remainingFreeSpace == pr.Inf {
				remainingFreeSpace = math.MaxInt32
			}

			initialMagnitude := -pr.Inf
			if initialFreeSpace > 0 {
				initialMagnitude = pr.Float(math.Round(math.Log10(float64(initialFreeSpace))))
			}
			remainingMagnitude := -pr.Inf
			if remainingFreeSpace > 0 {
				remainingMagnitude = pr.Float(math.Round(math.Log10(float64(remainingFreeSpace))))
			}
			if initialMagnitude < remainingMagnitude {
				remainingFreeSpace = initialFreeSpace
			}

			// Step 6 - 9.7.4.c
			if remainingFreeSpace == 0 {
				// "Do nothing", but we at least set the flexBaseSize as
				// targetMainSize for next step.
				for _, v := range line.line {
					child := v.box.Box()
					if !child.Frozen {
						child.TargetMainSize = child.FlexBaseSize
					}
				}
			} else {
				var scaledFlexShrinkFactorsSum, flexGrowFactorsSum pr.Float
				for _, v := range line.line {
					child := v.box.Box()
					if !child.Frozen {
						child.ScaledFlexShrinkFactor = child.FlexBaseSize * child.Style.GetFlexShrink()
						scaledFlexShrinkFactorsSum += child.ScaledFlexShrinkFactor
						flexGrowFactorsSum += child.Style.GetFlexGrow()
					}
				}
				for _, v := range line.line {
					child := v.box.Box()
					if !child.Frozen {
						if flexFactorType == "grow" {
							ratio := child.Style.GetFlexGrow() / flexGrowFactorsSum
							child.TargetMainSize = child.FlexBaseSize + remainingFreeSpace*ratio
						} else if flexFactorType == "shrink" {
							if scaledFlexShrinkFactorsSum == 0 {
								child.TargetMainSize = child.FlexBaseSize
							} else {
								ratio := child.ScaledFlexShrinkFactor / scaledFlexShrinkFactorsSum
								child.TargetMainSize = child.FlexBaseSize + remainingFreeSpace*ratio
							}
						}
					}
				}
			}

			// Step 6 - 9.7.4.d
			// TODO: First part of this step is useless until 3.E is correct
			for _, v := range line.line {
				child := v.box.Box()
				child.Adjustment = 0
				if !child.Frozen && child.TargetMainSize < 0 {
					child.Adjustment = -child.TargetMainSize
					child.TargetMainSize = 0
				}
			}

			// Step 6 - 9.7.4.e
			adjustments := line.adjustements()
			for _, v := range line.line {
				child := v.box.Box()
				if adjustments == 0 {
					child.Frozen = true
				} else if adjustments > 0 && child.Adjustment > 0 {
					child.Frozen = true
				} else if adjustments < 0 && child.Adjustment < 0 {
					child.Frozen = true
				}
			}
		}
		// Step 6 - 9.7.5
		for _, v := range line.line {
			child := v.box.Box()
			if axis == "width" {
				child.Width = child.TargetMainSize - child.PaddingLeft.V() - child.PaddingRight.V() -
					child.BorderLeftWidth.V() - child.BorderRightWidth.V()
				if child.MarginLeft != pr.Auto {
					child.Width = child.Width.V() - child.MarginLeft.V()
				}
				if child.MarginRight != pr.Auto {
					child.Width = child.Width.V() - child.MarginRight.V()
				}
			} else {
				child.Height = child.TargetMainSize - child.PaddingTop.V() - child.PaddingBottom.V() -
					child.BorderTopWidth.V() - child.BorderTopWidth.V()
				if child.MarginLeft != pr.Auto {
					child.Height = child.Height.V() - child.MarginLeft.V()
				}
				if child.MarginRight != pr.Auto {
					child.Height = child.Height.V() - child.MarginRight.V()
				}
			}
		}
	}

	// Step 7
	// TODO: Fix TODO in build.FlexChildren
	// TODO: Handle breaks
	var newFlexLines []flexLine
	childSkipStack = skipStack
	for _, line := range flexLines {
		var newFlexLine flexLine
		for _, v := range line.line {
			child_ := v.box
			child := child_.Box()
			// TODO: Find another way than calling blockLevelLayoutSwitch to
			// get baseline and child.Height
			if child.MarginTop == pr.Auto {
				child.MarginTop = pr.Float(0)
			}
			if child.MarginBottom == pr.Auto {
				child.MarginBottom = pr.Float(0)
			}
			childCopy := child_.Copy()
			if bo.ParentBoxT.IsInstance(child_) {
				childCopy = bo.CopyWithChildren(child_, child.Children)
			}

			blockLevelWidth(childCopy, nil, parentBox)
			newChild, tmp := blockLevelLayoutSwitch(context, childCopy.(bo.BlockLevelBoxITF), pr.Inf, childSkipStack,
				parentBox, pageIsEmpty, absoluteBoxes, fixedBoxes, new([]pr.Float), false)
			adjoiningMargins := tmp.adjoiningMargins
			child.Baseline = pr.Float(0)
			if bl := findInFlowBaseline(newChild, false); bl != nil {
				child.Baseline = bl.V()
			}
			if cross == "height" {
				child.Height = newChild.Box().Height
				// As flex items margins never collapse (with other flex items
				// or with the flex container), we can add the adjoining margins
				// to the child bottom margin.
				child.MarginBottom = child.MarginBottom.V() + collapseMargin(adjoiningMargins)
			} else {
				child.Width = minContentWidth(context, child_, false)
			}

			newFlexLine.line = append(newFlexLine.line, indexedBox{index: v.index, box: child_})

			// Skip stack is only for the first child
			childSkipStack = nil
		}
		if len(newFlexLine.line) != 0 {
			newFlexLines = append(newFlexLines, newFlexLine)
		}
	}
	flexLines = newFlexLines

	// Step 8
	crossSize := getAttr(box, cross, "")
	if len(flexLines) == 1 && crossSize != pr.Auto {
		flexLines[0].crossSize = crossSize.V()
	} else {
		for index, line := range flexLines {
			var collectedItems, notCollectedItems []*bo.BoxFields
			for _, v := range line.line {
				child := v.box.Box()
				alignSelf := child.Style.GetAlignSelf()
				if strings.HasPrefix(string(box.Style.GetFlexDirection()), "row") && alignSelf == "baseline" &&
					child.MarginTop != pr.Auto && child.MarginBottom != pr.Auto {
					collectedItems = append(collectedItems, child)
				} else {
					notCollectedItems = append(notCollectedItems, child)
				}
			}
			var crossStartDistance, crossEndDistance pr.Float
			for _, child := range collectedItems {
				baseline := child.Baseline.V() - child.PositionY
				crossStartDistance = pr.Max(crossStartDistance, baseline)
				crossEndDistance = pr.Max(crossEndDistance, child.MarginHeight()-baseline)
			}
			collectedCrossSize := crossStartDistance + crossEndDistance
			var nonCollectedCrossSize pr.Float
			if len(notCollectedItems) != 0 {
				nonCollectedCrossSize = -pr.Inf
				for _, child := range notCollectedItems {
					var childCrossSize pr.Float
					if cross == "height" {
						childCrossSize = child.BorderHeight()
						if child.MarginTop != pr.Auto {
							childCrossSize += child.MarginTop.V()
						}
						if child.MarginBottom != pr.Auto {
							childCrossSize += child.MarginBottom.V()
						}
					} else {
						childCrossSize = child.BorderWidth()
						if child.MarginLeft != pr.Auto {
							childCrossSize += child.MarginLeft.V()
						}
						if child.MarginRight != pr.Auto {
							childCrossSize += child.MarginRight.V()
						}
					}
					nonCollectedCrossSize = pr.Max(childCrossSize, nonCollectedCrossSize)
				}
			}
			line.crossSize = pr.Max(collectedCrossSize, nonCollectedCrossSize)
			flexLines[index] = line
		}
	}

	if len(flexLines) == 1 {
		line := flexLines[0]
		minCrossSize := getAttr(box, cross, "min")
		if minCrossSize == pr.Auto {
			minCrossSize = -pr.Inf
		}
		maxCrossSize := getAttr(box, cross, "max")
		if maxCrossSize == pr.Auto {
			maxCrossSize = pr.Inf
		}
		line.crossSize = pr.Max(minCrossSize.V(), pr.Min(line.crossSize, maxCrossSize.V()))
	}

	// Step 9
	if box.Style.GetAlignContent() == "stretch" {
		var definiteCrossSize pr.MaybeFloat
		if he := box.Style.GetHeight(); cross == "height" && he.String != "auto" {
			definiteCrossSize = he.Value
		} else if cross == "width" {
			if bo.FlexBoxT.IsInstance(box_) {
				if box.Style.GetWidth().String == "auto" {
					definiteCrossSize = availableCrossSpace
				} else {
					definiteCrossSize = box.Style.GetWidth().Value
				}
			}
		}
		if definiteCrossSize != nil {
			extraCrossSize := definiteCrossSize.V()
			for _, line := range flexLines {
				extraCrossSize -= line.crossSize
			}
			if extraCrossSize != 0 {
				for i, line := range flexLines {
					line.crossSize += extraCrossSize / pr.Float(len(flexLines))
					flexLines[i] = line
				}
			}
		}
	}

	// TODO: Step 10

	// Step 11
	for _, line := range flexLines {
		for _, v := range line.line {
			child := v.box.Box()
			alignSelf := child.Style.GetAlignSelf()
			if alignSelf == "auto" {
				alignSelf = box.Style.GetAlignItems()
			}
			if alignSelf == "stretch" && getCross(child, cross).String == "auto" {
				crossMargins := getCrossMargins(child, cross)
				if getCross(child, cross).String == "auto" {
					if !(crossMargins[0] == pr.Auto || crossMargins[1] == pr.Auto) {
						crossSize := line.crossSize
						if cross == "height" {
							crossSize -= child.MarginTop.V() + child.MarginBottom.V() +
								child.PaddingTop.V() + child.PaddingBottom.V() + child.BorderTopWidth.V() + child.BorderBottomWidth.V()
						} else {
							crossSize -= child.MarginLeft.V() + child.MarginRight.V() +
								child.PaddingLeft.V() + child.PaddingRight.V() + child.BorderLeftWidth.V() + child.BorderRightWidth.V()
						}
						if cross == "width" {
							child.Width = crossSize
						} else {
							child.Height = crossSize
						}
						// TODO: redo layout?
					}
				}
			} // else: Cross size has been set by step 7
		}
	}

	// Step 12
	// TODO: handle rtl
	originalPositionAxis := box.ContentBoxY()
	if axis == "width" {
		originalPositionAxis = box.ContentBoxX()
	}

	justifyContent := box.Style.GetJustifyContent()
	if strings.HasSuffix(string(box.Style.GetFlexDirection()), "-reverse") {
		if justifyContent == "flex-start" {
			justifyContent = "flex-end"
		} else if justifyContent == "flex-end" {
			justifyContent = "flex-start"
		}
	}

	for _, line := range flexLines {
		positionAxis := originalPositionAxis
		var freeSpace pr.Float
		if axis == "width" {
			freeSpace = box.Width.V()
			for _, v := range line.line {
				child := v.box.Box()
				freeSpace -= child.BorderWidth()
				if child.MarginLeft != pr.Auto {
					freeSpace -= child.MarginLeft.V()
				}
				if child.MarginRight != pr.Auto {
					freeSpace -= child.MarginRight.V()
				}
			}
		} else {
			freeSpace = box.Height.V()
			for _, v := range line.line {
				child := v.box.Box()
				freeSpace -= child.BorderHeight()
				if child.MarginTop != pr.Auto {
					freeSpace -= child.MarginTop.V()
				}
				if child.MarginBottom != pr.Auto {
					freeSpace -= child.MarginBottom.V()
				}
			}
		}

		var margins pr.Float
		for _, v := range line.line {
			child := v.box.Box()
			if axis == "width" {
				if child.MarginLeft == pr.Auto {
					margins += 1
				}
				if child.MarginRight == pr.Auto {
					margins += 1
				}
			} else {
				if child.MarginTop == pr.Auto {
					margins += 1
				}
				if child.MarginBottom == pr.Auto {
					margins += 1
				}
			}
		}
		if margins != 0 {
			freeSpace /= margins
			for _, v := range line.line {
				child := v.box.Box()
				if axis == "width" {
					if child.MarginLeft == pr.Auto {
						child.MarginLeft = freeSpace
					}
					if child.MarginRight == pr.Auto {
						child.MarginRight = freeSpace
					}
				} else {
					if child.MarginTop == pr.Auto {
						child.MarginTop = freeSpace
					}
					if child.MarginBottom == pr.Auto {
						child.MarginBottom = freeSpace
					}
				}
			}
			freeSpace = 0
		}

		if box.Style.GetDirection() == "rtl" && axis == "width" {
			freeSpace = -freeSpace
		}

		if justifyContent == "flex-end" {
			positionAxis += freeSpace
		} else if justifyContent == "center" {
			positionAxis += freeSpace / 2
		} else if justifyContent == "space-around" {
			positionAxis += freeSpace / pr.Float(len(line.line)) / 2
		} else if justifyContent == "space-evenly" {
			positionAxis += freeSpace / (pr.Float(len(line.line)) + 1)
		}

		for _, v := range line.line {
			child := v.box.Box()
			if axis == "width" {
				child.PositionX = positionAxis
				if justifyContent == "stretch" {
					child.Width = child.Width.V() + freeSpace/pr.Float(len(line.line))
				}

			} else {
				child.PositionY = positionAxis
			}

			if axis == "width" {
				if box.Style.GetDirection() == "rtl" {
					positionAxis += -child.MarginWidth()
				} else {
					positionAxis += child.MarginWidth()
				}
			} else {
				positionAxis += child.MarginHeight()
			}

			if justifyContent == "space-around" {
				positionAxis += freeSpace / pr.Float(len(line.line))
			} else if justifyContent == "space-between" {
				if len(line.line) > 1 {
					positionAxis += freeSpace / (pr.Float(len(line.line)) - 1)
				}
			} else if justifyContent == "space-evenly" {
				positionAxis += freeSpace / (pr.Float(len(line.line)) + 1)
			}
		}
	}

	// Step 13
	positionCross := box.ContentBoxX()
	if cross == "height" {
		positionCross = box.ContentBoxY()
	}
	for index, line := range flexLines {
		line.lowerBaseline = 0
		// TODO: don't duplicate this loop
		for _, v := range line.line {
			child := v.box.Box()
			alignSelf := child.Style.GetAlignSelf()
			if alignSelf == "auto" {
				alignSelf = box.Style.GetAlignItems()
			}
			if alignSelf == "baseline" && axis == "width" {
				// TODO: handle vertical text
				child.Baseline = child.Baseline.V() - positionCross
				line.lowerBaseline = pr.Max(line.lowerBaseline, child.Baseline.V())
			}
		}
		for _, v := range line.line {
			child := v.box.Box()
			crossMargins := getCrossMargins(child, cross)
			var autoMargins pr.Float
			if crossMargins[0] == pr.Auto {
				autoMargins += 1
			}
			if crossMargins[1] == pr.Auto {
				autoMargins += 1
			}
			if autoMargins != 0 {
				extraCross := line.crossSize
				if cross == "height" {
					extraCross -= child.BorderHeight()
					if child.MarginTop != pr.Auto {
						extraCross -= child.MarginTop.V()
					}
					if child.MarginBottom != pr.Auto {
						extraCross -= child.MarginBottom.V()
					}
				} else {
					extraCross -= child.BorderWidth()
					if child.MarginLeft != pr.Auto {
						extraCross -= child.MarginLeft.V()
					}
					if child.MarginRight != pr.Auto {
						extraCross -= child.MarginRight.V()
					}
				}
				if extraCross > 0 {
					extraCross /= autoMargins
					if cross == "height" {
						if child.MarginTop == pr.Auto {
							child.MarginTop = extraCross
						}
						if child.MarginBottom == pr.Auto {
							child.MarginBottom = extraCross
						}
					} else {
						if child.MarginLeft == pr.Auto {
							child.MarginLeft = extraCross
						}
						if child.MarginRight == pr.Auto {
							child.MarginRight = extraCross
						}
					}
				} else {
					if cross == "height" {
						if child.MarginTop == pr.Auto {
							child.MarginTop = pr.Float(0)
						}
						child.MarginBottom = extraCross
					} else {
						if child.MarginLeft == pr.Auto {
							child.MarginLeft = pr.Float(0)
						}
						child.MarginRight = extraCross
					}
				}
			} else {
				// Step 14
				alignSelf := child.Style.GetAlignSelf()
				if alignSelf == "auto" {
					alignSelf = box.Style.GetAlignItems()
				}
				if cross == "height" {
					child.PositionY = positionCross
				} else {
					child.PositionX = positionCross
				}
				if alignSelf == "flex-end" {
					if cross == "height" {
						child.PositionY += line.crossSize - child.MarginHeight()
					} else {
						child.PositionX += line.crossSize - child.MarginWidth()
					}
				} else if alignSelf == "center" {
					if cross == "height" {
						child.PositionY += (line.crossSize - child.MarginHeight()) / 2
					} else {
						child.PositionX += (line.crossSize - child.MarginWidth()) / 2
					}
				} else if alignSelf == "baseline" {
					if cross == "height" {
						child.PositionY += line.lowerBaseline - child.Baseline.V()
					} else {
						// Handle vertical text
					}
				} else if alignSelf == "stretch" {
					if getCross(child, cross).String == "auto" {
						var margins pr.Float
						if cross == "height" {
							margins = child.MarginTop.V() + child.MarginBottom.V()
						} else {
							margins = child.MarginLeft.V() + child.MarginRight.V()
						}
						if child.Style.GetBoxSizing() == "content-box" {
							if cross == "height" {
								margins += child.BorderTopWidth.V() + child.BorderBottomWidth.V() +
									child.PaddingTop.V() + child.PaddingBottom.V()
							} else {
								margins += child.BorderLeftWidth.V() + child.BorderRightWidth.V() +
									child.PaddingLeft.V() + child.PaddingRight.V()
							}
						}
						// TODO: don't set style width, find a way to avoid
						// width re-calculation after Step 16
						child.Style.Set(cross, pr.Dimension{Value: line.crossSize - margins, Unit: pr.Px}.ToValue())
					}
				}
			}
		}
		positionCross += line.crossSize
		flexLines[index] = line
	}

	sc := sumCross(flexLines)
	// Step 15
	if getCross(box, cross).String == "auto" {
		// TODO: handle min-max
		if cross == "height" {
			box.Height = sc
		} else {
			box.Width = sc
		}
	} else if len(flexLines) > 1 { // Step 16
		extraCrossSize := getAttr(box, cross, "").V() - sc
		direction := "positionX"
		if cross == "height" {
			direction = "positionY"
		}
		boxAlignContent := box.Style.GetAlignContent()
		if extraCrossSize > 0 {
			var crossTranslate pr.Float
			for _, line := range flexLines {
				for _, v := range line.line {
					child := v.box.Box()
					if child.IsFlexItem {
						currentValue := child.PositionX
						if direction == "positionY" {
							currentValue = child.PositionY
						}
						currentValue += crossTranslate
						setDirection(child, direction, currentValue)

						switch boxAlignContent {
						case "flex-end":
							setDirection(child, direction, currentValue+extraCrossSize)
						case "center":
							setDirection(child, direction, currentValue+extraCrossSize/2)
						case "space-around":
							setDirection(child, direction, currentValue+extraCrossSize/pr.Float(len(flexLines))/2)
						case "space-evenly":
							setDirection(child, direction, currentValue+extraCrossSize/(pr.Float(len(flexLines))+1))
						}
					}
				}
				switch boxAlignContent {
				case "space-between":
					crossTranslate += extraCrossSize / (pr.Float(len(flexLines)) - 1)
				case "space-around":
					crossTranslate += extraCrossSize / pr.Float(len(flexLines))
				case "space-evenly":
					crossTranslate += extraCrossSize / (pr.Float(len(flexLines)) + 1)
				}
			}
		}
	}

	// TODO: don't use blockBoxLayout, see TODOs in Step 14 and
	// build.FlexChildren.
	box_ = box_.Copy()
	box = box_.Box()
	box.Children = nil
	childSkipStack = skipStack
	for _, line := range flexLines {
		for _, v := range line.line {
			i, child := v.index, v.box.Box()
			if child.IsFlexItem {
				newChild, tmp := blockLevelLayoutSwitch(context, v.box.(bo.BlockLevelBoxITF), maxPositionY, childSkipStack, box,
					pageIsEmpty, absoluteBoxes, fixedBoxes, new([]pr.Float), false)
				childResumeAt := tmp.resumeAt
				if newChild == nil {
					if resumeAt != nil {
						if resumeIndex, _ := resumeAt.Unpack(); resumeIndex != 0 {
							resumeAt = tree.ResumeStack{resumeIndex + i - 1: nil}
						}
					}
				} else {
					box.Children = append(box.Children, newChild)
					if childResumeAt != nil {
						firstLevelSkip := 0
						if originalSkipStack != nil {
							firstLevelSkip, _ = originalSkipStack.Unpack()
						}
						if resumeAt != nil {
							resumeIndex, _ := resumeAt.Unpack()
							firstLevelSkip += resumeIndex
						}
						resumeAt = tree.ResumeStack{firstLevelSkip + i: childResumeAt}
					}
				}
				if resumeAt != nil {
					break
				}
			}

			// Skip stack is only for the first child
			childSkipStack = nil
		}
		if resumeAt != nil {
			break
		}
	}

	// Set box height
	// TODO: this is probably useless because of step #15
	if axis == "width" && box.Height == pr.Auto {
		if len(flexLines) != 0 {
			box.Height = sumCross(flexLines)
		} else {
			box.Height = pr.Float(0)
		}
	}

	// Set baseline
	// See https://www.W3.org/TR/css-flexbox-1/#flex-baselines
	// TODO: use the real algorithm
	if bo.InlineFlexBoxT.IsInstance(box_) {
		if axis == "width" { // and main text direction is horizontal
			if len(flexLines) != 0 {
				box.Baseline = flexLines[0].lowerBaseline
			} else {
				box.Baseline = pr.Float(0)
			}
		} else {
			var val pr.MaybeFloat
			if len(box.Children) != 0 {
				val = findInFlowBaseline(box.Children[0], false)
			}
			if val != nil {
				box.Baseline = val.V()
			} else {
				box.Baseline = pr.Float(0)
			}
		}
	}

	context.finishBlockFormattingContext(box_)

	// TODO: check these returned values
	return box_, blockLayout{
		resumeAt:          resumeAt,
		nextPage:          tree.PageBreak{Break: "any"},
		adjoiningMargins:  nil,
		collapsingThrough: false,
	}
}
