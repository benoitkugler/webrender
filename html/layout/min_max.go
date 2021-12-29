package layout

import (
	pr "github.com/benoitkugler/webrender/css/properties"
)

type block struct {
	X, Y, Width, Height pr.Float
}

func (b block) ContainingBlock() (width, height pr.MaybeFloat) { return b.Width, b.Height }

type containingBlock interface {
	ContainingBlock() (width, height pr.MaybeFloat)
}

type funcMinMax = func(box Box, context *layoutContext, containingBlock containingBlock) (bool, pr.Float)

// Decorate a function that sets the used width of a box to handle
// {min,max}-width.
func handleMinMaxWidth(function funcMinMax) funcMinMax {
	wrapper := func(box Box, context *layoutContext, containingBlock containingBlock) (bool, pr.Float) {
		computedMarginL, computedMarginR := box.Box().MarginLeft, box.Box().MarginRight
		res1, res2 := function(box, context, containingBlock)
		if box.Box().Width.V() > box.Box().MaxWidth.V() {
			box.Box().Width = box.Box().MaxWidth
			box.Box().MarginLeft, box.Box().MarginRight = computedMarginL, computedMarginR
			res1, res2 = function(box, context, containingBlock)
		}
		if box.Box().Width.V() < box.Box().MinWidth.V() {
			box.Box().Width = box.Box().MinWidth
			box.Box().MarginLeft, box.Box().MarginRight = computedMarginL, computedMarginR
			res1, res2 = function(box, context, containingBlock)
		}
		return res1, res2
	}
	// wrapper.WithoutMinMax = function
	return wrapper
}

// Decorate a function that sets the used height of a box to handle
// {min,max}-height.
func handleMinMaxHeight(function funcMinMax) funcMinMax {
	wrapper := func(box Box, context *layoutContext, containingBlock containingBlock) (bool, pr.Float) {
		computedMarginT, computedMarginB := box.Box().MarginTop, box.Box().MarginBottom
		res1, res2 := function(box, context, containingBlock)
		if box.Box().Height.V() > box.Box().MaxHeight.V() {
			box.Box().Height = box.Box().MaxHeight
			box.Box().MarginTop, box.Box().MarginBottom = computedMarginT, computedMarginB
			res1, res2 = function(box, context, containingBlock)
		}
		if box.Box().Height.V() < box.Box().MinHeight.V() {
			box.Box().Height = box.Box().MinHeight
			box.Box().MarginTop, box.Box().MarginBottom = computedMarginT, computedMarginB
			res1, res2 = function(box, context, containingBlock)
		}
		return res1, res2
	}
	//  wrapper.WithoutMinMax = function
	return wrapper
}
