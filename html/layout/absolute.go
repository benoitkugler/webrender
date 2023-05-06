package layout

import (
	"fmt"
	"math"

	pr "github.com/benoitkugler/webrender/css/properties"

	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/tree"
)

// ---------------------- Absolutely positioned boxes management. ----------------

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

var absoluteWidth = handleMinMaxWidth(_absoluteWidth)

// @handleMinMaxWidth
// containingBlock must be block
func _absoluteWidth(box_ Box, context *layoutContext, containingBlock containingBlock) (bool, pr.Float) {
	// https://www.w3.org/TR/CSS2/visudet.html#abs-replaced-width
	box := box_.Box()

	ltr := box.Style.ParentStyle() == nil || box.Style.ParentStyle().GetDirection() == "ltr"
	paddingsBorders := box.PaddingLeft.V() + box.PaddingRight.V() + box.BorderLeftWidth.V() + box.BorderRightWidth.V()

	marginL := box.MarginLeft
	marginR := box.MarginRight
	width := box.Width
	left := box.Left
	right := box.Right

	cb_ := containingBlock.(block)
	cbX, cbWidth := cb_.X, cb_.Width

	var translateX pr.Float = 0
	translateBoxWidth := false
	defaultTranslateX := cbX - box.PositionX
	if left == pr.AutoF && right == pr.AutoF && width == pr.AutoF {
		if marginL == pr.AutoF {
			box.MarginLeft = pr.Float(0)
		}
		if marginR == pr.AutoF {
			box.MarginRight = pr.Float(0)
		}
		availableWidth := cbWidth - (paddingsBorders + box.MarginLeft.V() + box.MarginRight.V())
		box.Width = shrinkToFit(context, box_, availableWidth)
		if !ltr {
			translateBoxWidth = true
			translateX = defaultTranslateX + availableWidth
		}
	} else if left != pr.AutoF && right != pr.AutoF && width != pr.AutoF {
		widthForMargins := cbWidth - (right.V() + left.V() + width.V() + paddingsBorders)
		if marginL == pr.AutoF && marginR == pr.AutoF {
			if width.V()+paddingsBorders+right.V()+left.V() <= cbWidth {
				box.MarginLeft = widthForMargins / 2
				box.MarginRight = box.MarginLeft
			} else {
				if ltr {
					box.MarginLeft = pr.Float(0)
					box.MarginRight = widthForMargins
				} else {
					box.MarginLeft = widthForMargins
					box.MarginRight = pr.Float(0)
				}
			}
		} else if marginL == pr.AutoF {
			box.MarginLeft = widthForMargins
		} else if marginR == pr.AutoF {
			box.MarginRight = widthForMargins
		} else if ltr {
			box.MarginRight = widthForMargins
		} else {
			box.MarginLeft = widthForMargins
		}
		translateX = left.V() + defaultTranslateX
	} else {
		if marginL == pr.AutoF {
			box.MarginLeft = pr.Float(0)
		}
		if marginR == pr.AutoF {
			box.MarginRight = pr.Float(0)
		}
		spacing := paddingsBorders + box.MarginLeft.V() + box.MarginRight.V()
		if left == pr.AutoF && width == pr.AutoF {
			box.Width = shrinkToFit(context, box_, cbWidth-spacing-right.V())
			translateX = cbWidth - right.V() - spacing + defaultTranslateX
			translateBoxWidth = true
		} else if left == pr.AutoF && right == pr.AutoF {
			if !ltr {
				availableWidth := cbWidth - (paddingsBorders + box.MarginLeft.V() + box.MarginRight.V())
				translateBoxWidth = true
				translateX = defaultTranslateX + availableWidth
			}
		} else if width == pr.AutoF && right == pr.AutoF {
			box.Width = shrinkToFit(context, box_, cbWidth-spacing-left.V())
			translateX = left.V() + defaultTranslateX
		} else if left == pr.AutoF {
			translateX = cbWidth + defaultTranslateX - right.V() - spacing - width.V()
		} else if width == pr.AutoF {
			box.Width = cbWidth.V() - right.V() - left.V() - spacing
			translateX = left.V() + defaultTranslateX
		} else if right == pr.AutoF {
			translateX = left.V() + defaultTranslateX
		}
	}
	return translateBoxWidth, translateX
}

func absoluteHeight(box_ Box, containingBlock block) (bool, pr.Float) {
	box := box_.Box()

	paddingsBorders := box.PaddingTop.V() + box.PaddingBottom.V() + box.BorderTopWidth.V() + box.BorderBottomWidth.V()

	marginT := box.MarginTop
	marginB := box.MarginBottom
	height := box.Height
	top := box.Top
	bottom := box.Bottom

	cbY, cbHeight := containingBlock.Y, containingBlock.Height

	// https://www.w3.org/TR/CSS2/visudet.html#abs-non-replaced-height

	var translateY pr.Float = 0
	translateBoxHeight := false
	defaultTranslateY := cbY - box.PositionY
	if top == pr.AutoF && bottom == pr.AutoF && height == pr.AutoF {
		// Keep the static position
		if marginT == pr.AutoF {
			box.MarginTop = pr.Float(0)
		}
		if marginB == pr.AutoF {
			box.MarginBottom = pr.Float(0)
		}
	} else if top != pr.AutoF && bottom != pr.AutoF && height != pr.AutoF {
		heightForMargins := cbHeight - (top.V() + bottom.V() + height.V() + paddingsBorders)
		if marginT == pr.AutoF && marginB == pr.AutoF {
			box.MarginTop = heightForMargins / 2
			box.MarginBottom = box.MarginTop
		} else if marginT == pr.AutoF {
			box.MarginTop = heightForMargins
		} else if marginB == pr.AutoF {
			box.MarginBottom = heightForMargins
		} else {
			box.MarginBottom = heightForMargins
		}
		translateY = top.V() + defaultTranslateY
	} else {
		if marginT == pr.AutoF {
			box.MarginTop = pr.Float(0)
		}
		if marginB == pr.AutoF {
			box.MarginBottom = pr.Float(0)
		}
		spacing := paddingsBorders + box.MarginTop.V() + box.MarginBottom.V()
		if top == pr.AutoF && height == pr.AutoF {
			translateY = cbHeight.V() - bottom.V() - spacing + defaultTranslateY
			translateBoxHeight = true
		} else if top == pr.AutoF && bottom == pr.AutoF {
			// Keep the static position
		} else if height == pr.AutoF && bottom == pr.AutoF {
			translateY = top.V() + defaultTranslateY
		} else if top == pr.AutoF {
			translateY = cbHeight.V() + defaultTranslateY - bottom.V() - spacing - height.V()
		} else if height == pr.AutoF {
			box.Height = cbHeight.V() - bottom.V() - top.V() - spacing
			translateY = top.V() + defaultTranslateY
		} else if bottom == pr.AutoF {
			translateY = top.V() + defaultTranslateY
		}
	}
	return translateBoxHeight, translateY
}

// performs either blockContainerLayout or flexLayout on box_
func absoluteLayoutDriver(context *layoutContext, box_ Box, containingBlock block, fixedBoxes *[]*AbsolutePlaceholder,
	bottomSpace pr.Float, skipStack tree.ResumeStack,
	isBlock bool,
) (Box, tree.ResumeStack) {
	box := box_.Box()
	cbWidth, cbHeight := containingBlock.Width, containingBlock.Height

	translateBoxWidth, translateX := absoluteWidth(box_, context, containingBlock)
	translateBoxHeight, translateY := false, pr.Float(0)
	if len(skipStack) == 0 {
		translateBoxHeight, translateY = absoluteHeight(box_, containingBlock)
	}

	if translateBoxHeight {
		bottomSpace -= box.PositionY
	} else {
		bottomSpace += translateY
	}

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
		newBox, bl, _ = blockContainerLayout(context, box_, bottomSpace, skipStack, true,
			&absoluteBoxes, fixedBoxes, new([]pr.Float), false, -1)
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

// Set the width of absolute positioned “box“.
func absoluteLayout(context *layoutContext, placeholder *AbsolutePlaceholder, containingBlock Box,
	fixedBoxes *[]*AbsolutePlaceholder, bottomSpace pr.Float, skipStack tree.ResumeStack,
) {
	if placeholder.layoutDone {
		panic("placeholder can't have its layout done.")
	}
	box := placeholder.AliasBox
	newBox, resumeAt := absoluteBoxLayout(context, box, containingBlock, fixedBoxes, bottomSpace, skipStack)
	placeholder.setLaidOutBox(newBox)
	if resumeAt != nil {
		context.brokenOutOfFlow[placeholder] = brokenBox{
			box:             box,
			containingBlock: containingBlock,
			resumeAt:        resumeAt,
		}
	}
}

func absoluteBoxLayout(context *layoutContext, box Box, cb_ Box, fixedBoxes *[]*AbsolutePlaceholder,
	bottomSpace pr.Float, skipStack tree.ResumeStack,
) (Box, tree.ResumeStack) {
	// https://www.w3.org/TR/CSS2/visudet.html1#containing-block-details

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
	if bo.BlockT.IsInstance(box) {
		newBox, resumeAt = absoluteBlock(context, box, containingBlock, fixedBoxes, bottomSpace, skipStack)
	} else if bo.FlexContainerT.IsInstance(box) {
		newBox, resumeAt = absoluteFlex(context, box, containingBlock, fixedBoxes, bottomSpace, skipStack)
	} else {
		if !bo.BlockReplacedT.IsInstance(box) {
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

	ltr := box.Style.ParentStyle() == nil || box.Style.ParentStyle().GetDirection() == "ltr"

	// https://www.w3.org/TR/CSS21/visudet.html#abs-replaced-width
	if box.Left == pr.AutoF && box.Right == pr.AutoF {
		// static position:
		if ltr {
			box.Left = box.PositionX - cbX
		} else {
			box.Right = cbX + cbWidth - box.PositionX
		}
	}
	if box.Left == pr.AutoF || box.Right == pr.AutoF {
		if box.MarginLeft == pr.AutoF {
			box.MarginLeft = pr.Float(0)
		}
		if box.MarginRight == pr.AutoF {
			box.MarginRight = pr.Float(0)
		}
		remaining := cbWidth - box.MarginWidth()
		if box.Left == pr.AutoF {
			box.Left = remaining - box.Right.V()
		}
		if box.Right == pr.AutoF {
			box.Right = remaining - box.Left.V()
		}
	} else if pr.AutoF == box.MarginLeft || pr.AutoF == box.MarginRight {
		remaining := cbWidth - (box.BorderWidth() + box.Left.V() + box.Right.V())
		if box.MarginLeft == pr.AutoF && box.MarginRight == pr.AutoF {
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
		} else if box.MarginLeft == pr.AutoF {
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

	// https://www.w3.org/TR/CSS21/visudet.html#abs-replaced-height
	if box.Top == pr.AutoF && box.Bottom == pr.AutoF {
		box.Top = box.PositionY - cbY
	}
	if box.Top == pr.AutoF || box.Bottom == pr.AutoF {
		if box.MarginTop == pr.AutoF {
			box.MarginTop = pr.Float(0)
		}
		if box.MarginBottom == pr.AutoF {
			box.MarginBottom = pr.Float(0)
		}
		remaining := cbHeight - box.MarginHeight()
		if box.Top == pr.AutoF {
			box.Top = remaining - box.Bottom.V()
		}
		if box.Bottom == pr.AutoF {
			box.Bottom = remaining - box.Top.V()
		}
	} else if box.MarginTop == pr.AutoF || box.MarginBottom == pr.AutoF {
		remaining := cbHeight - (box.BorderHeight() + box.Top.V() + box.Bottom.V())
		if box.MarginTop == pr.AutoF && box.MarginBottom == pr.AutoF {
			box.MarginTop = intDiv(remaining, 2)
			box.MarginBottom = box.MarginTop
		} else if box.MarginTop == pr.AutoF {
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
