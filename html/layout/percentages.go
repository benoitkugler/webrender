package layout

import (
	"fmt"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/utils/testutils/tracer"
)

// Resolve percentages into fixed values.

// Compute a used length value from a computed length value.
//
// the return value should be set on the box
// mainFlexDirection is either 0, Width or Height
func resolveOnePercentage(value pr.Value, propertyName pr.KnownProp, referTo pr.Float, mainFlexDirection pr.KnownProp) pr.MaybeFloat {
	// box attributes are used values
	percent := pr.ResoudPercentage(value, referTo)

	if (propertyName == pr.PMinWidth || propertyName == pr.PMinHeight) && percent == pr.AutoF {
		if mainFlexDirection == 0 || propertyName != mainFlexDirection+2 {
			percent = pr.Float(0)
		}
	}

	if traceMode {
		traceLogger.Dump(fmt.Sprintf("resolveOnePercentage %s: %v %s -> %s", propertyName,
			tracer.FormatMaybeFloat(value.ToMaybeFloat()), tracer.FormatMaybeFloat(referTo),
			tracer.FormatMaybeFloat(percent)))
	}

	return percent
}

func resolvePositionPercentages(box *bo.BoxFields, containingBlock bo.Point) {
	cbWidth, cbHeight := containingBlock[0], containingBlock[1]
	box.Left = resolveOnePercentage(box.Style.GetLeft(), pr.PLeft, cbWidth, 0)
	box.Right = resolveOnePercentage(box.Style.GetRight(), pr.PRight, cbWidth, 0)
	box.Top = resolveOnePercentage(box.Style.GetTop(), pr.PTop, cbHeight, 0)
	box.Bottom = resolveOnePercentage(box.Style.GetBottom(), pr.PBottom, cbHeight, 0)
}

// mainFlexDirection is either 0, Width or Height
func resolvePercentagesBox(box Box, containingBlock containingBlock, mainFlexDirection pr.KnownProp) {
	w, h := containingBlock.ContainingBlock()
	resolvePercentages(box, bo.MaybePoint{w, h}, mainFlexDirection)
}

// Set used values as attributes of the box object.
// mainFlexDirection is either 0, Width or Height
func resolvePercentages(box_ Box, containingBlock bo.MaybePoint, mainFlexDirection pr.KnownProp) {
	cbWidth, cbHeight := containingBlock[0], containingBlock[1]

	if traceMode {
		traceLogger.Dump(fmt.Sprintf("resolvePercentages: %s %s",
			tracer.FormatMaybeFloat(cbWidth), tracer.FormatMaybeFloat(cbHeight)))
	}

	maybeHeight := cbWidth
	if bo.PageT.IsInstance(box_) {
		maybeHeight = cbHeight
	}
	box := box_.Box()
	box.MarginLeft = resolveOnePercentage(box.Style.GetMarginLeft(), pr.PMarginLeft, cbWidth.V(), 0)
	box.MarginRight = resolveOnePercentage(box.Style.GetMarginRight(), pr.PMarginRight, cbWidth.V(), 0)
	box.MarginTop = resolveOnePercentage(box.Style.GetMarginTop(), pr.PMarginTop, maybeHeight.V(), 0)
	box.MarginBottom = resolveOnePercentage(box.Style.GetMarginBottom(), pr.PMarginBottom, maybeHeight.V(), 0)
	box.PaddingLeft = resolveOnePercentage(box.Style.GetPaddingLeft(), pr.PPaddingLeft, cbWidth.V(), 0)
	box.PaddingRight = resolveOnePercentage(box.Style.GetPaddingRight(), pr.PPaddingRight, cbWidth.V(), 0)
	box.PaddingTop = resolveOnePercentage(box.Style.GetPaddingTop(), pr.PPaddingTop, maybeHeight.V(), 0)
	box.PaddingBottom = resolveOnePercentage(box.Style.GetPaddingBottom(), pr.PPaddingBottom, maybeHeight.V(), 0)
	box.Width = resolveOnePercentage(box.Style.GetWidth(), pr.PWidth, cbWidth.V(), 0)
	box.MinWidth = resolveOnePercentage(box.Style.GetMinWidth(), pr.PMinWidth, cbWidth.V(), mainFlexDirection)
	box.MaxWidth = resolveOnePercentage(box.Style.GetMaxWidth(), pr.PMaxWidth, cbWidth.V(), mainFlexDirection)

	// XXX later: top, bottom, left && right on positioned elements

	if cbHeight == pr.AutoF {
		// Special handling when the height of the containing block
		// depends on its content.
		height := box.Style.GetHeight()
		if height.String == "auto" || height.Unit == pr.Perc {
			box.Height = pr.AutoF
		} else {
			if height.Unit != pr.Px {
				panic(fmt.Sprintf("expected percentage, got %d", height.Unit))
			}
			box.Height = height.Value
		}
		box.MinHeight = resolveOnePercentage(box.Style.GetMinHeight(), pr.PMinHeight, pr.Float(0), mainFlexDirection)
		box.MaxHeight = resolveOnePercentage(box.Style.GetMaxHeight(), pr.PMaxHeight, pr.Inf, mainFlexDirection)
	} else {
		box.Height = resolveOnePercentage(box.Style.GetHeight(), pr.PHeight, cbHeight.V(), 0)
		box.MinHeight = resolveOnePercentage(box.Style.GetMinHeight(), pr.PMinHeight, cbHeight.V(), mainFlexDirection)
		box.MaxHeight = resolveOnePercentage(box.Style.GetMaxHeight(), pr.PMaxHeight, cbHeight.V(), mainFlexDirection)
	}

	// Used value == computed value
	box.BorderTopWidth = box.Style.GetBorderTopWidth().Value
	box.BorderRightWidth = box.Style.GetBorderRightWidth().Value
	box.BorderBottomWidth = box.Style.GetBorderBottomWidth().Value
	box.BorderLeftWidth = box.Style.GetBorderLeftWidth().Value

	// Shrink *content* widths and heights according to box-sizing
	// Thanks heavens and the spec: Our validator rejects negative values
	// for padding and border-width
	var horizontalDelta, verticalDelta pr.Float
	switch box.Style.GetBoxSizing() {
	case "border-box":
		horizontalDelta = box.PaddingLeft.V() + box.PaddingRight.V() + box.BorderLeftWidth.V() + box.BorderRightWidth.V()
		verticalDelta = box.PaddingTop.V() + box.PaddingBottom.V() + box.BorderTopWidth.V() + box.BorderBottomWidth.V()
	case "padding-box":
		horizontalDelta = box.PaddingLeft.V() + box.PaddingRight.V()
		verticalDelta = box.PaddingTop.V() + box.PaddingBottom.V()
	case "content-box":
		horizontalDelta = 0
		verticalDelta = 0
	default:
		panic(fmt.Sprintf("invalid box sizing %s", box.Style.GetBoxSizing()))
	}

	// Keep at least min* >= 0 to prevent funny output in case box.Width or
	// box.Height become negative.
	// Restricting max* seems reasonable, too.
	if horizontalDelta > 0 {
		if box.Width != pr.AutoF {
			box.Width = pr.Max(0, box.Width.V()-horizontalDelta)
		}
		box.MaxWidth = pr.Max(0, box.MaxWidth.V()-horizontalDelta)
		if box.MinWidth != pr.AutoF {
			box.MinWidth = pr.Max(0, box.MinWidth.V()-horizontalDelta)
		}
	}
	if verticalDelta > 0 {
		if box.Height != pr.AutoF {
			box.Height = pr.Max(0, box.Height.V()-verticalDelta)
		}
		box.MaxHeight = pr.Max(0, box.MaxHeight.V()-verticalDelta)
		if box.MinHeight != pr.AutoF {
			box.MinHeight = pr.Max(0, box.MinHeight.V()-verticalDelta)
		}
	}
}

func resoudRadius(box *bo.BoxFields, v pr.Point, side1, side2 bo.Side) bo.MaybePoint {
	// rx, ry = v
	if v[0] == pr.ZeroPixels || v[1] == pr.ZeroPixels { // Short track for common case
		return bo.MaybePoint{pr.Float(0), pr.Float(0)}
	}

	if box.RemoveDecorationSides[side1] || box.RemoveDecorationSides[side2] {
		return bo.MaybePoint{pr.Float(0), pr.Float(0)}
	}
	rx := pr.ResoudPercentage(v[0].ToValue(), box.BorderWidth())
	ry := pr.ResoudPercentage(v[1].ToValue(), box.BorderHeight())
	return bo.MaybePoint{rx, ry}
}

func resolveRadiiPercentages(box *bo.BoxFields) {
	box.BorderTopLeftRadius = resoudRadius(box, box.Style.GetBorderTopLeftRadius(), bo.STop, bo.SLeft).V()
	box.BorderTopRightRadius = resoudRadius(box, box.Style.GetBorderTopRightRadius(), bo.STop, bo.SRight).V()
	box.BorderBottomRightRadius = resoudRadius(box, box.Style.GetBorderBottomRightRadius(), bo.SBottom, bo.SRight).V()
	box.BorderBottomLeftRadius = resoudRadius(box, box.Style.GetBorderBottomLeftRadius(), bo.SBottom, bo.SLeft).V()
}
