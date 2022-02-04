package layout

import (
	"fmt"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/tree"
)

// Layout for images and other replaced elements.
// http://dev.w3.org/csswg/css-images-3/#sizing

// Default sizing algorithm for the concrete object size.
// http://dev.w3.org/csswg/css-images-3/#default-sizing
func defaultImageSizing(intrinsicWidth, intrinsicHeight, intrinsicRatio,
	specifiedWidth, specifiedHeight pr.MaybeFloat, defaultWidth, defaultHeight pr.Float) (concreteWidth, concreteHeight pr.Float) {

	if specifiedWidth == pr.AutoF {
		specifiedWidth = nil
	}
	if specifiedHeight == pr.AutoF {
		specifiedHeight = nil
	}

	if specifiedWidth != nil && specifiedHeight != nil {
		return specifiedWidth.V(), specifiedHeight.V()
	} else if specifiedWidth != nil {
		if intrinsicRatio != nil {
			concreteHeight = specifiedWidth.V() / intrinsicRatio.V()
		} else if intrinsicHeight != nil {
			concreteHeight = intrinsicHeight.V()
		} else {
			concreteHeight = defaultHeight
		}
		return specifiedWidth.V(), concreteHeight
	} else if specifiedHeight != nil {
		if intrinsicRatio != nil {
			concreteWidth = specifiedHeight.V() * intrinsicRatio.V()
		} else if intrinsicWidth != nil {
			concreteWidth = intrinsicWidth.V()
		} else {
			concreteWidth = defaultWidth
		}
		return concreteWidth, specifiedHeight.V()
	} else {
		if intrinsicWidth != nil || intrinsicHeight != nil {
			return defaultImageSizing(intrinsicWidth, intrinsicHeight, intrinsicRatio,
				intrinsicWidth, intrinsicHeight, defaultWidth, defaultHeight)
		} else {
			return containConstraintImageSizing(defaultWidth, defaultHeight, intrinsicRatio)
		}
	}
}

// Cover constraint sizing algorithm for the concrete object size.
// http://dev.w3.org/csswg/css-images-3/#contain-constraint
func containConstraintImageSizing(constraintWidth, constraintHeight pr.Float, intrinsicRatio pr.MaybeFloat) (concreteWidth, concreteHeight pr.Float) {
	return constraintImageSizing(constraintWidth, constraintHeight, intrinsicRatio, false)
}

// Cover constraint sizing algorithm for the concrete object size.
// http://dev.w3.org/csswg/css-images-3/#cover-constraint
func coverConstraintImageSizing(constraintWidth, constraintHeight pr.Float, intrinsicRatio pr.MaybeFloat) (concreteWidth, concreteHeight pr.Float) {
	return constraintImageSizing(constraintWidth, constraintHeight, intrinsicRatio, true)
}

func constraintImageSizing(constraintWidth, constraintHeight pr.Float, intrinsicRatio pr.MaybeFloat, cover bool) (concreteWidth, concreteHeight pr.Float) {
	if intrinsicRatio == nil {
		return constraintWidth, constraintHeight
	} else if cover != (constraintWidth > constraintHeight*intrinsicRatio.V()) {
		return constraintHeight * intrinsicRatio.V(), constraintHeight
	} else {
		return constraintWidth, constraintWidth / intrinsicRatio.V()
	}
}

// LayoutReplacedBox computes the dimension of the content of a replaced box.
func LayoutReplacedBox(box_ bo.ReplacedBoxITF) (drawWidth, drawHeight, positionX, positionY pr.Float) {
	box := box_.Replaced()
	// TODO: respect box-sizing ?
	objectFit := box.Style.GetObjectFit()
	position := box.Style.GetObjectPosition() // see validation

	image := box.Replacement
	intrinsicWidth, intrinsicHeight, ratio := image.GetIntrinsicSize(box.Style.GetImageResolution().Value, box.Style.GetFontSize().Value)
	if intrinsicWidth == nil || intrinsicHeight == nil {
		intrinsicWidth, intrinsicHeight = containConstraintImageSizing(box.Width.V(), box.Height.V(), ratio)
	}

	if objectFit == "fill" {
		drawWidth, drawHeight = box.Width.V(), box.Height.V()
	} else {
		if objectFit == "contain" || objectFit == "scale-down" {
			drawWidth, drawHeight = containConstraintImageSizing(box.Width.V(), box.Height.V(), ratio)
		} else if objectFit == "cover" {
			drawWidth, drawHeight = coverConstraintImageSizing(box.Width.V(), box.Height.V(), ratio)
		} else if objectFit == "none" {
			drawWidth, drawHeight = intrinsicWidth.V(), intrinsicHeight.V()
		} else {
			panic(fmt.Sprintf("unexpected objectFit %s", objectFit))
		}

		if objectFit == "scale-down" {
			drawWidth = pr.Min(drawWidth, intrinsicWidth.V())
			drawHeight = pr.Min(drawHeight, intrinsicHeight.V())
		}
	}

	originX, positionX_, originY, positionY_ := position.OriginX, position.Pos[0], position.OriginY, position.Pos[1]
	refX := box.Width.V() - drawWidth
	refY := box.Height.V() - drawHeight

	positionX = pr.ResoudPercentage(positionX_.ToValue(), refX).V()
	positionY = pr.ResoudPercentage(positionY_.ToValue(), refY).V()
	if originX == "right" {
		positionX = refX - positionX
	}
	if originY == "bottom" {
		positionY = refY - positionY
	}

	positionX += box.ContentBoxX()
	positionY += box.ContentBoxY()

	return drawWidth, drawHeight, positionX, positionY
}

var replacedBoxWidth = handleMinMaxWidth(replacedBoxWidth_)

// @handleMinMaxWidth
// Compute and set the used width for replaced boxes (inline- or block-level)
// containingBlock must be block
func replacedBoxWidth_(box_ Box, _ *layoutContext, containingBlock containingBlock) (bool, pr.Float) {
	box__, ok := box_.(bo.ReplacedBoxITF)
	if !ok {
		panic(fmt.Sprintf("expected ReplacedBox instance, got %s", box_))
	}
	box := box__.Replaced()
	intrinsicWidth, intrinsicHeight, ratio := box.Replacement.GetIntrinsicSize(box.Style.GetImageResolution().Value, box.Style.GetFontSize().Value)

	// This algorithm simply follows the different points of the specification
	// http://www.w3.org/TR/CSS21/visudet.html#inline-replaced-width
	if box.Height == pr.AutoF && box.Width == pr.AutoF {
		if intrinsicWidth != nil {
			// Point #1
			box.Width = intrinsicWidth
		} else if ratio != nil {
			if intrinsicHeight != nil {
				// Point #2 first part
				box.Width = intrinsicHeight.V() * ratio.V()
			} else {
				// Point #3
				blockLevelWidth(box, nil, containingBlock)
			}
		}
	}
	if box.Width == pr.AutoF {
		if ratio != nil {
			// Point #2 second part
			box.Width = box.Height.V() * ratio.V()
		} else if intrinsicWidth != nil {
			// Point #4
			box.Width = intrinsicWidth
		} else {
			// Point #5
			// It's pretty useless to rely on device size to set width.
			box.Width = pr.Float(300)
		}
	}

	return false, 0
}

var replacedBoxHeight = handleMinMaxHeight(replacedBoxHeight_)

// Compute and set the used height for replaced boxes (inline- or block-level)
// box_ must be a ReplacedBoxITF
func replacedBoxHeight_(box_ Box, _ *layoutContext, _ containingBlock) (bool, pr.Float) {
	box__, ok := box_.(bo.ReplacedBoxITF)
	if !ok {
		panic(fmt.Sprintf("expected ReplacedBox instance, got %s", box_))
	}
	box := box__.Replaced()
	// http://www.w3.org/TR/CSS21/visudet.html#inline-replaced-height
	_, intrinsicHeight, ratio := box.Replacement.GetIntrinsicSize(
		box.Style.GetImageResolution().Value, box.Style.GetFontSize().Value)

	// Test pr.Auto on the computed width, not the used width
	if box.Height == pr.AutoF && box.Width == pr.AutoF {
		box.Height = intrinsicHeight
	} else if box.Height == pr.AutoF && pr.Is(ratio) {
		box.Height = box.Width.V() / ratio.V()
	}

	if box.Height == pr.AutoF && box.Width == pr.AutoF && intrinsicHeight != nil {
		box.Height = intrinsicHeight
	} else if ratio != nil && box.Height == pr.AutoF {
		box.Height = box.Width.V() / ratio.V()
	} else if box.Height == pr.AutoF && intrinsicHeight != nil {
		box.Height = intrinsicHeight
	} else if box.Height == pr.AutoF {
		// It"s pretty useless to rely on device size to set width.
		box.Height = pr.Float(150)
	}

	return false, 0
}

// Resolve min/max constraints on replaced elements
// that have "auto" width or heights.
func minMaxAutoReplaced(box *bo.BoxFields) {
	width := box.Width.V()
	height := box.Height.V()
	minWidth := box.MinWidth.V()
	minHeight := box.MinHeight.V()
	maxWidth := pr.Max(minWidth, box.MaxWidth.V())
	maxHeight := pr.Max(minHeight, box.MaxHeight.V())

	// (violationWidth, violationHeight)
	var violationWidth, violationHeight string
	if width < minWidth {
		violationWidth = "min"
	} else if width > maxWidth {
		violationWidth = "max"
	}
	if height < minHeight {
		violationHeight = "min"
	} else if height > maxHeight {
		violationHeight = "max"
	}

	// Work around divisions by zero. These are pathological cases anyway.
	// TODO: is there a cleaner way?
	if width == 0 {
		width = 1e-6
	}
	if height == 0 {
		height = 1e-6
	}

	switch [2]string{violationWidth, violationHeight} {
	// ("", ""): nothing to do
	case [2]string{"max", ""}:
		box.Width = maxWidth
		box.Height = pr.Max(maxWidth*height/width, minHeight)
	case [2]string{"min", ""}:
		box.Width = minWidth
		box.Height = pr.Min(minWidth*height/width, maxHeight)
	case [2]string{"", "max"}:
		box.Width = pr.Max(maxHeight*width/height, minWidth)
		box.Height = maxHeight
	case [2]string{"", "min"}:
		box.Width = pr.Min(minHeight*width/height, maxWidth)
		box.Height = minHeight
	case [2]string{"max", "max"}:
		if maxWidth/width <= maxHeight/height {
			box.Width = maxWidth
			box.Height = pr.Max(minHeight, maxWidth*height/width)
		} else {
			box.Width = pr.Max(minWidth, maxHeight*width/height)
			box.Height = maxHeight
		}
	case [2]string{"min", "min"}:
		if minWidth/width <= minHeight/height {
			box.Width = pr.Min(maxWidth, minHeight*width/height)
			box.Height = minHeight
		} else {
			box.Width = minWidth
			box.Height = pr.Min(maxHeight, minWidth*height/width)
		}
	case [2]string{"min", "max"}:
		box.Width = minWidth
		box.Height = maxHeight
	case [2]string{"max", "min"}:
		box.Width = maxWidth
		box.Height = minHeight
	}
}

// Lay out an inline :class:`boxes.ReplacedBox` ``box``.
func inlineReplacedBoxLayout(box_ Box, containingBlock *bo.BoxFields) {
	resolveMarginAuto(box_.Box())
	inlineReplacedBoxWidthHeight(box_, containingBlock)
}

func inlineReplacedBoxWidthHeight(box Box, containingBlock containingBlock) {
	if style := box.Box().Style; style.GetWidth().String == "auto" && style.GetHeight().String == "auto" {
		replacedBoxWidth_(box, nil, containingBlock)
		replacedBoxHeight_(box, nil, nil)
		minMaxAutoReplaced(box.Box())
	} else {
		replacedBoxWidth(box, nil, containingBlock)
		replacedBoxHeight(box, nil, nil)
	}
}

// Lay out the block :class:`boxes.ReplacedBox` ``box``.
func blockReplacedBoxLayout(context *layoutContext, box_ bo.ReplacedBoxITF, containingBlock *bo.BoxFields) (bo.ReplacedBoxITF, blockLayout) {
	box_ = box_.Copy().(bo.ReplacedBoxITF) // Copy is type stable
	box := box_.Box()
	if box.Style.GetWidth().String == "auto" && box.Style.GetHeight().String == "auto" {
		computedMarginsL, computedMarginsR := box.MarginLeft, box.MarginRight
		blockReplacedWidth_(box_, nil, containingBlock)
		replacedBoxHeight_(box_, nil, nil)
		minMaxAutoReplaced(box)
		box.MarginLeft, box.MarginRight = computedMarginsL, computedMarginsR
		blockLevelWidth_(box_, nil, containingBlock)
	} else {
		blockReplacedWidth(box_, nil, containingBlock)
		replacedBoxHeight(box_, nil, nil)
	}

	// Don't collide with floats
	// http://www.w3.org/TR/CSS21/visuren.html#floats
	box.PositionX, box.PositionY, _ = avoidCollisions(context, box_, containingBlock, false)
	nextPage := tree.PageBreak{Break: "any"}

	if traceMode {
		traceLogger.DumpTree(box_, "blockReplacedBoxLayout-end")
	}

	return box_, blockLayout{resumeAt: nil, nextPage: nextPage, adjoiningMargins: nil, collapsingThrough: false}
}
