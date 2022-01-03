package layout

import (
	"fmt"
	"math"

	pr "github.com/benoitkugler/webrender/css/properties"

	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/tree"
)

type AliasBox = bo.Box

// AbsolutePlaceholder is left where an absolutely-positioned box was taken out of the flow.
type AbsolutePlaceholder struct {
	AliasBox
	layoutDone bool
}

func NewAbsolutePlaceholder(box Box) *AbsolutePlaceholder {
	out := AbsolutePlaceholder{AliasBox: box, layoutDone: false}
	return &out
}

func (AbsolutePlaceholder) IsClassicalBox() bool { return false }

func (abs *AbsolutePlaceholder) setLaidOutBox(newBox Box) {
	abs.AliasBox = newBox
	abs.layoutDone = true
}

func (abs *AbsolutePlaceholder) Translate(box Box, dx, dy pr.Float, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	if abs.layoutDone {
		abs.AliasBox.Translate(box, dx, dy, ignoreFloats)
	} else {
		// Descendants do not have a position yet.
		abs.AliasBox.Box().PositionX += dx
		abs.AliasBox.Box().PositionY += dy
	}
}

func (abs AbsolutePlaceholder) Copy() Box {
	out := abs
	out.AliasBox = abs.AliasBox.Copy()
	return &out
}

func (abs AbsolutePlaceholder) String() string {
	return fmt.Sprintf("<Placeholder %s (%s)>", abs.AliasBox.Type(), abs.AliasBox.Box().ElementTag())
}

func toBoxes(children []*AbsolutePlaceholder) []Box {
	asBox := make([]Box, len(children))
	for i, v := range children {
		asBox[i] = v
	}
	return asBox
}

func fromBoxes(children []Box) []*AbsolutePlaceholder {
	asPlac := make([]*AbsolutePlaceholder, len(children))
	for i, v := range children {
		asPlac[i] = v.(*AbsolutePlaceholder)
	}
	return asPlac
}

var absoluteWidth = handleMinMaxWidth(_absoluteWidth)

// @handleMinMaxWidth
// containingBlock must be block
func _absoluteWidth(box_ Box, context *layoutContext, containingBlock containingBlock) (bool, pr.Float) {
	// http://www.w3.org/TR/CSS2/visudet.html#abs-replaced-width
	box := box_.Box()
	// These names are waaay too long
	marginL := box.MarginLeft
	marginR := box.MarginRight
	paddingL := box.PaddingLeft
	paddingR := box.PaddingRight
	borderL := box.BorderLeftWidth
	borderR := box.BorderRightWidth
	width := box.Width
	left := box.Left
	right := box.Right

	cb_ := containingBlock.(block)
	cbX, cbWidth := cb_.X, cb_.Width

	paddingPlusBordersX := paddingL.V() + paddingR.V() + borderL.V() + borderR.V()
	var translateX pr.Float = 0
	translateBoxWidth := false
	defaultTranslateX := cbX - box.PositionX
	if left == pr.Auto && right == pr.Auto && width == pr.Auto {
		if marginL == pr.Auto {
			box.MarginLeft = pr.Float(0)
		}
		if marginR == pr.Auto {
			box.MarginRight = pr.Float(0)
		}
		availableWidth := cbWidth - (paddingPlusBordersX + box.MarginLeft.V() + box.MarginRight.V())
		box.Width = shrinkToFit(context, box_, availableWidth)
	} else if left != pr.Auto && right != pr.Auto && width != pr.Auto {
		widthForMargins := cbWidth - (right.V() + left.V() + width.V() + paddingPlusBordersX)
		if marginL == pr.Auto && marginR == pr.Auto {
			if width.V()+paddingPlusBordersX+right.V()+left.V() <= cbWidth {
				box.MarginLeft = widthForMargins / 2
				box.MarginRight = box.MarginLeft
			} else {
				box.MarginLeft = pr.Float(0)
				box.MarginRight = widthForMargins
			}
		} else if marginL == pr.Auto {
			box.MarginLeft = widthForMargins
		} else if marginR == pr.Auto {
			box.MarginRight = widthForMargins
		} else {
			box.MarginRight = widthForMargins
		}
		translateX = left.V() + defaultTranslateX
	} else {
		if marginL == pr.Auto {
			box.MarginLeft = pr.Float(0)
		}
		if marginR == pr.Auto {
			box.MarginRight = pr.Float(0)
		}
		spacing := paddingPlusBordersX + box.MarginLeft.V() + box.MarginRight.V()
		if left == pr.Auto && width == pr.Auto {
			box.Width = shrinkToFit(context, box_, cbWidth-spacing-right.V())
			translateX = cbWidth - right.V() - spacing + defaultTranslateX
			translateBoxWidth = true
		} else if left == pr.Auto && right == pr.Auto {
			// Keep the static position
		} else if width == pr.Auto && right == pr.Auto {
			box.Width = shrinkToFit(context, box_, cbWidth-spacing-left.V())
			translateX = left.V() + defaultTranslateX
		} else if left == pr.Auto {
			translateX = cbWidth + defaultTranslateX - right.V() - spacing - width.V()
		} else if width == pr.Auto {
			box.Width = cbWidth.V() - right.V() - left.V() - spacing
			translateX = left.V() + defaultTranslateX
		} else if right == pr.Auto {
			translateX = left.V() + defaultTranslateX
		}
	}
	return translateBoxWidth, translateX
}

func absoluteHeight(box_ Box, containingBlock block) (bool, pr.Float) {
	box := box_.Box()
	// These names are waaay too long
	marginT := box.MarginTop
	marginB := box.MarginBottom
	paddingT := box.PaddingTop
	paddingB := box.PaddingBottom
	borderT := box.BorderTopWidth
	borderB := box.BorderBottomWidth
	height := box.Height
	top := box.Top
	bottom := box.Bottom

	cbY, cbHeight := containingBlock.Y, containingBlock.Height

	// http://www.w3.org/TR/CSS2/visudet.html#abs-non-replaced-height

	paddingsPlusBordersY := paddingT.V() + paddingB.V() + borderT.V() + borderB.V()
	var translateY pr.Float = 0
	translateBoxHeight := false
	defaultTranslateY := cbY - box.PositionY
	if top == pr.Auto && bottom == pr.Auto && height == pr.Auto {
		// Keep the static position
		if marginT == pr.Auto {
			box.MarginTop = pr.Float(0)
		}
		if marginB == pr.Auto {
			box.MarginBottom = pr.Float(0)
		}
	} else if top != pr.Auto && bottom != pr.Auto && height != pr.Auto {
		heightForMargins := cbHeight - (top.V() + bottom.V() + height.V() + paddingsPlusBordersY)
		if marginT == pr.Auto && marginB == pr.Auto {
			box.MarginTop = heightForMargins / 2
			box.MarginBottom = box.MarginTop
		} else if marginT == pr.Auto {
			box.MarginTop = heightForMargins
		} else if marginB == pr.Auto {
			box.MarginBottom = heightForMargins
		} else {
			box.MarginBottom = heightForMargins
		}
		translateY = top.V() + defaultTranslateY
	} else {
		if marginT == pr.Auto {
			box.MarginTop = pr.Float(0)
		}
		if marginB == pr.Auto {
			box.MarginBottom = pr.Float(0)
		}
		spacing := paddingsPlusBordersY + box.MarginTop.V() + box.MarginBottom.V()
		if top == pr.Auto && height == pr.Auto {
			translateY = cbHeight.V() - bottom.V() - spacing + defaultTranslateY
			translateBoxHeight = true
		} else if top == pr.Auto && bottom == pr.Auto {
			// Keep the static position
		} else if height == pr.Auto && bottom == pr.Auto {
			translateY = top.V() + defaultTranslateY
		} else if top == pr.Auto {
			translateY = cbHeight.V() + defaultTranslateY - bottom.V() - spacing - height.V()
		} else if height == pr.Auto {
			box.Height = cbHeight.V() - bottom.V() - top.V() - spacing
			translateY = top.V() + defaultTranslateY
		} else if bottom == pr.Auto {
			translateY = top.V() + defaultTranslateY
		}
	}
	return translateBoxHeight, translateY
}

// performs either blockContainerLayout or flexLayout on box_
func absoluteLayoutDriver(context *layoutContext, box_ Box, containingBlock block, fixedBoxes *[]*AbsolutePlaceholder, bottomSpace pr.Float, skipStack tree.ResumeStack, isBlock bool) (Box, tree.ResumeStack) {
	box := box_.Box()
	cbWidth, cbHeight := containingBlock.Width, containingBlock.Height

	translateBoxWidth, translateX := absoluteWidth(box_, context, containingBlock)
	translateBoxHeight, translateY := absoluteHeight(box_, containingBlock)

	// This box is the containing block for absolute descendants.
	var absoluteBoxes []*AbsolutePlaceholder

	if box.IsTableWrapper {
		tableWrapperWidth(context, box_, bo.MaybePoint{cbWidth, cbHeight})
	}

	var (
		newBox Box
		bl     blockLayout
	)
	if isBlock {
		newBox, bl = blockContainerLayout(context, box_, bottomSpace, skipStack, true, &absoluteBoxes, fixedBoxes, new([]pr.Float), false)
	} else {
		newBox, bl = flexLayout(context, box_, bottomSpace, skipStack, containingBlock, true, &absoluteBoxes, fixedBoxes)
	}

	for _, childPlaceholder := range absoluteBoxes {
		absoluteLayout(context, childPlaceholder, newBox, fixedBoxes, bottomSpace, skipStack)
	}

	if translateBoxWidth {
		translateX -= newBox.Box().Width.V()
	}
	if translateBoxHeight {
		translateY -= newBox.Box().Height.V()
	}

	newBox.Translate(newBox, translateX, translateY, false)

	return newBox, bl.resumeAt
}

func absoluteBlock(context *layoutContext, box_ Box, containingBlock block, fixedBoxes *[]*AbsolutePlaceholder, bottomSpace pr.Float, skipStack tree.ResumeStack) (Box, tree.ResumeStack) {
	return absoluteLayoutDriver(context, box_, containingBlock, fixedBoxes, bottomSpace, skipStack, true)
}

func absoluteFlex(context *layoutContext, box_ Box, containingBlock block, fixedBoxes *[]*AbsolutePlaceholder, bottomSpace pr.Float, skipStack tree.ResumeStack) (Box, tree.ResumeStack) {
	return absoluteLayoutDriver(context, box_, containingBlock, fixedBoxes, bottomSpace, skipStack, false)
}

// Set the width of absolute positioned ``box``.
func absoluteLayout(context *layoutContext, placeholder *AbsolutePlaceholder, containingBlock Box,
	fixedBoxes *[]*AbsolutePlaceholder, bottomSpace pr.Float, skipStack tree.ResumeStack) {
	if placeholder.layoutDone {
		panic("placeholder can't have its layout done.")
	}
	box := placeholder.AliasBox
	newBox, resumeAt := absoluteBoxLayout(context, box, containingBlock, fixedBoxes, bottomSpace, skipStack)
	placeholder.setLaidOutBox(newBox)
	if resumeAt != nil {
		context.brokenOutOfFlow = append(context.brokenOutOfFlow, brokenBox{
			box:             box,
			containingBlock: containingBlock,
			resumeAt:        resumeAt,
		})
	}
}

func absoluteBoxLayout(context *layoutContext, box Box, cb_ Box, fixedBoxes *[]*AbsolutePlaceholder,
	bottomSpace pr.Float, skipStack tree.ResumeStack) (Box, tree.ResumeStack) {
	// http://www.w3.org/TR/CSS2/visudet.html1#containing-block-details

	if traceMode {
		traceLogger.DumpTree(box, "absoluteBoxLayout")
	}

	var containingBlock block
	cb := cb_.Box()
	if _, isPageBox := cb_.(*bo.PageBox); isPageBox {
		containingBlock.X = cb.ContentBoxX()
		containingBlock.Y = cb.ContentBoxY()
		containingBlock.Width = cb.Width.V()
		containingBlock.Height = cb.Height.V()
	} else {
		containingBlock.X = cb.PaddingBoxX()
		containingBlock.Y = cb.PaddingBoxY()
		containingBlock.Width = cb.PaddingWidth()
		containingBlock.Height = cb.PaddingHeight()
	}

	resolvePercentages(box, bo.MaybePoint{containingBlock.Width, containingBlock.Height}, "")
	resolvePositionPercentages(box.Box(), bo.Point{containingBlock.Width, containingBlock.Height})

	context.createBlockFormattingContext()
	// Absolute tables are wrapped into block boxes
	var (
		newBox   Box
		resumeAt tree.ResumeStack
	)
	if bo.BlockBoxT.IsInstance(box) {
		newBox, resumeAt = absoluteBlock(context, box, containingBlock, fixedBoxes, bottomSpace, skipStack)
	} else if bo.FlexContainerBoxT.IsInstance(box) {
		newBox, resumeAt = absoluteFlex(context, box, containingBlock, fixedBoxes, bottomSpace, skipStack)
	} else {
		if !bo.BlockReplacedBoxT.IsInstance(box) {
			panic(fmt.Sprintf("box should be a BlockReplaced, got %T", box))
		}
		newBox = absoluteReplaced(box, containingBlock)
	}
	context.finishBlockFormattingContext(newBox)
	return newBox, resumeAt
}

func intDiv(a pr.Float, b int) pr.Float {
	return pr.Float(int(math.Floor(float64(a))) / b)
}

func absoluteReplaced(box_ Box, containingBlock block) Box {
	inlineReplacedBoxWidthHeight(box_, containingBlock)
	box := box_.Box()
	cbX, cbY, cbWidth, cbHeight := containingBlock.X, containingBlock.Y, containingBlock.Width, containingBlock.Height
	ltr := box.Style.GetDirection() == "ltr"

	// http://www.w3.org/TR/CSS21/visudet.html#abs-replaced-width
	if box.Left == pr.Auto && box.Right == pr.Auto {
		// static position:
		if ltr {
			box.Left = box.PositionX - cbX
		} else {
			box.Right = cbX + cbWidth - box.PositionX
		}
	}
	if box.Left == pr.Auto || box.Right == pr.Auto {
		if box.MarginLeft == pr.Auto {
			box.MarginLeft = pr.Float(0)
		}
		if box.MarginRight == pr.Auto {
			box.MarginRight = pr.Float(0)
		}
		remaining := cbWidth - box.MarginWidth()
		if box.Left == pr.Auto {
			box.Left = remaining - box.Right.V()
		}
		if box.Right == pr.Auto {
			box.Right = remaining - box.Left.V()
		}
	} else if pr.Auto == box.MarginLeft || pr.Auto == box.MarginRight {
		remaining := cbWidth - (box.BorderWidth() + box.Left.V() + box.Right.V())
		if box.MarginLeft == pr.Auto && box.MarginRight == pr.Auto {
			if remaining >= 0 {
				box.MarginLeft = intDiv(remaining, 2)
				box.MarginRight = box.MarginLeft
			} else if ltr {
				box.MarginLeft = pr.Float(0)
				box.MarginRight = remaining
			} else {
				box.MarginLeft = remaining
				box.MarginRight = pr.Float(0)
			}
		} else if box.MarginLeft == pr.Auto {
			box.MarginLeft = remaining
		} else {
			box.MarginRight = remaining
		}
	} else {
		// Over-constrained
		if ltr {
			box.Right = cbWidth - (box.MarginWidth() + box.Left.V())
		} else {
			box.Left = cbWidth - (box.MarginWidth() + box.Right.V())
		}
	}

	// http://www.w3.org/TR/CSS21/visudet.html#abs-replaced-height
	if box.Top == pr.Auto && box.Bottom == pr.Auto {
		box.Top = box.PositionY - cbY
	}
	if box.Top == pr.Auto || box.Bottom == pr.Auto {
		if box.MarginTop == pr.Auto {
			box.MarginTop = pr.Float(0)
		}
		if box.MarginBottom == pr.Auto {
			box.MarginBottom = pr.Float(0)
		}
		remaining := cbHeight - box.MarginHeight()
		if box.Top == pr.Auto {
			box.Top = remaining
		}
		if box.Bottom == pr.Auto {
			box.Bottom = remaining
		}
	} else if box.MarginTop == pr.Auto || box.MarginBottom == pr.Auto {
		remaining := cbHeight - (box.BorderHeight() + box.Top.V() + box.Bottom.V())
		if box.MarginTop == pr.Auto && box.MarginBottom == pr.Auto {
			box.MarginTop = intDiv(remaining, 2)
			box.MarginBottom = box.MarginTop
		} else if box.MarginTop == pr.Auto {
			box.MarginTop = remaining
		} else {
			box.MarginBottom = remaining
		}
	} else {
		// Over-constrained
		box.Bottom = cbHeight - (box.MarginHeight() + box.Top.V())
	}

	// No children for replaced boxes, no need to .translate()
	box.PositionX = cbX + box.Left.V()
	box.PositionY = cbY + box.Top.V()
	return box_
}
