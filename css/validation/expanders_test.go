package validation

import (
	"reflect"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils"
	"github.com/benoitkugler/webrender/utils/testutils"
)

var inherit = pr.CascadedProperty{Default: pr.Inherit}.AsValidated()

// Test the 4-value pr.
func TestExpandFourSides(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "margin: inherit", map[string]pr.ValidatedProperty{
		"margin_top":    pr.Inherit.AsCascaded().AsValidated(),
		"margin_right":  pr.Inherit.AsCascaded().AsValidated(),
		"margin_bottom": pr.Inherit.AsCascaded().AsValidated(),
		"margin_left":   pr.Inherit.AsCascaded().AsValidated(),
	})
	assertValidDict(t, "margin: 1em", toValidated(pr.Properties{
		"margin_top":    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"margin_right":  pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"margin_bottom": pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"margin_left":   pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
	}))
	assertValidDict(t, "margin: -1em auto 20%", toValidated(pr.Properties{
		"margin_top":    pr.Dimension{Value: -1, Unit: pr.Em}.ToValue(),
		"margin_right":  pr.SToV("auto"),
		"margin_bottom": pr.Dimension{Value: 20, Unit: pr.Perc}.ToValue(),
		"margin_left":   pr.SToV("auto"),
	}))
	assertValidDict(t, "padding: 1em 0", toValidated(pr.Properties{
		"padding_top":    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"padding_right":  pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
		"padding_bottom": pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"padding_left":   pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
	}))
	assertValidDict(t, "padding: 1em 0 2%", toValidated(pr.Properties{
		"padding_top":    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"padding_right":  pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
		"padding_bottom": pr.Dimension{Value: 2, Unit: pr.Perc}.ToValue(),
		"padding_left":   pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
	}))
	assertValidDict(t, "padding: 1em 0 2em 5px", toValidated(pr.Properties{
		"padding_top":    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"padding_right":  pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
		"padding_bottom": pr.Dimension{Value: 2, Unit: pr.Em}.ToValue(),
		"padding_left":   pr.Dimension{Value: 5, Unit: pr.Px}.ToValue(),
	}))
	capt.AssertNoLogs(t)

	assertInvalid(t, "padding: 1 2 3 4 5", "expected 1 to 4 token components got 5")
	assertInvalid(t, "margin: rgb(0, 0, 0)", "invalid")
	assertInvalid(t, "padding: auto", "invalid")
	assertInvalid(t, "padding: -12px", "invalid")
	assertInvalid(t, "border-width: -3em", "invalid")
	assertInvalid(t, "border-width: 12%", "invalid")
}

// Test the “border“ property.
func TestExpandBorders(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "border-top: 3px dotted red", toValidated(pr.Properties{
		"border_top_width": pr.Dimension{Value: 3, Unit: pr.Px}.ToValue(),
		"border_top_style": pr.String("dotted"),
		"border_top_color": pr.NewColor(1, 0, 0, 1), // red
	}))
	assertValidDict(t, "border-top: 3px dotted", toValidated(pr.Properties{
		"border_top_width": pr.Dimension{Value: 3, Unit: pr.Px}.ToValue(),
		"border_top_style": pr.String("dotted"),
	}))
	assertValidDict(t, "border-top: 3px red", toValidated(pr.Properties{
		"border_top_width": pr.Dimension{Value: 3, Unit: pr.Px}.ToValue(),
		"border_top_color": pr.NewColor(1, 0, 0, 1), // red
	}))
	assertValidDict(t, "border-top: solid", toValidated(pr.Properties{
		"border_top_style": pr.String("solid"),
	}))
	assertValidDict(t, "border: 6px dashed lime", toValidated(pr.Properties{
		"border_top_width": pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		"border_top_style": pr.String("dashed"),
		"border_top_color": pr.NewColor(0, 1, 0, 1), // lime

		"border_left_width": pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		"border_left_style": pr.String("dashed"),
		"border_left_color": pr.NewColor(0, 1, 0, 1), // lime

		"border_bottom_width": pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		"border_bottom_style": pr.String("dashed"),
		"border_bottom_color": pr.NewColor(0, 1, 0, 1), // lime

		"border_right_width": pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		"border_right_style": pr.String("dashed"),
		"border_right_color": pr.NewColor(0, 1, 0, 1), // lime
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "border: 6px dashed left", "invalid")
}

func TestExpandBorderRadius(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "border-radius: 1px", toValidated(pr.Properties{
		"border_top_left_radius":     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		"border_top_right_radius":    pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		"border_bottom_right_radius": pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		"border_bottom_left_radius":  pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
	}))
	assertValidDict(t, "border-radius: 1px 2em", toValidated(pr.Properties{
		"border_top_left_radius":     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		"border_top_right_radius":    pr.Point{{Value: 2, Unit: pr.Em}, {Value: 2, Unit: pr.Em}},
		"border_bottom_right_radius": pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		"border_bottom_left_radius":  pr.Point{{Value: 2, Unit: pr.Em}, {Value: 2, Unit: pr.Em}},
	}))
	assertValidDict(t, "border-radius: 1px / 2em", toValidated(pr.Properties{
		"border_top_left_radius":     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
		"border_top_right_radius":    pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
		"border_bottom_right_radius": pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
		"border_bottom_left_radius":  pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
	}))
	assertValidDict(t, "border-radius: 1px 3px / 2em 4%", toValidated(pr.Properties{
		"border_top_left_radius":     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
		"border_top_right_radius":    pr.Point{{Value: 3, Unit: pr.Px}, {Value: 4, Unit: pr.Perc}},
		"border_bottom_right_radius": pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
		"border_bottom_left_radius":  pr.Point{{Value: 3, Unit: pr.Px}, {Value: 4, Unit: pr.Perc}},
	}))
	assertValidDict(t, "border-radius: 1px 2em 3%", toValidated(pr.Properties{
		"border_top_left_radius":     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		"border_top_right_radius":    pr.Point{{Value: 2, Unit: pr.Em}, {Value: 2, Unit: pr.Em}},
		"border_bottom_right_radius": pr.Point{{Value: 3, Unit: pr.Perc}, {Value: 3, Unit: pr.Perc}},
		"border_bottom_left_radius":  pr.Point{{Value: 2, Unit: pr.Em}, {Value: 2, Unit: pr.Em}},
	}))
	assertValidDict(t, "border-radius: 1px 2em 3% 4rem", toValidated(pr.Properties{
		"border_top_left_radius":     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		"border_top_right_radius":    pr.Point{{Value: 2, Unit: pr.Em}, {Value: 2, Unit: pr.Em}},
		"border_bottom_right_radius": pr.Point{{Value: 3, Unit: pr.Perc}, {Value: 3, Unit: pr.Perc}},
		"border_bottom_left_radius":  pr.Point{{Value: 4, Unit: pr.Rem}, {Value: 4, Unit: pr.Rem}},
	}))
	assertValidDict(t, "border-radius: inherit", map[string]pr.ValidatedProperty{
		"border_top_left_radius":     inherit,
		"border_top_right_radius":    inherit,
		"border_bottom_right_radius": inherit,
		"border_bottom_left_radius":  inherit,
	})
	capt.AssertNoLogs(t)

	assertInvalid(t, "border-radius: 1px 1px 1px 1px 1px", "1 to 4 token")
	assertInvalid(t, "border-radius: 1px 1px 1px 1px 1px / 1px", "1 to 4 token")
	assertInvalid(t, "border-radius: 1px / 1px / 1px", `only one '/'`)
	assertInvalid(t, "border-radius: 12deg", "invalid")
	assertInvalid(t, "border-radius: 1px 1px 1px 12deg", "invalid")
	assertInvalid(t, "border-radius: super", "invalid")
	assertInvalid(t, "border-radius: 1px, 1px", "invalid")
	assertInvalid(t, "border-radius: 1px /", `value after '/'`)
}

// Test the “list_style“ property.
func TestExpandList_style(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "list-style: inherit", map[string]pr.ValidatedProperty{
		"list_style_position": pr.Inherit.AsCascaded().AsValidated(),
		"list_style_image":    pr.Inherit.AsCascaded().AsValidated(),
		"list_style_type":     pr.Inherit.AsCascaded().AsValidated(),
	})
	assertValidDict(t, "list-style: url(../bar/lipsum.png)", toValidated(pr.Properties{
		"list_style_image": pr.UrlImage("https://weasyprint.org/bar/lipsum.png"),
	}))
	assertValidDict(t, "list-style: square", toValidated(pr.Properties{
		"list_style_type": pr.CounterStyleID{Name: "square"},
	}))
	assertValidDict(t, "list-style: circle inside", toValidated(pr.Properties{
		"list_style_position": pr.String("inside"),
		"list_style_type":     pr.CounterStyleID{Name: "circle"},
	}))
	assertValidDict(t, "list-style: none circle inside", toValidated(pr.Properties{
		"list_style_position": pr.String("inside"),
		"list_style_image":    pr.NoneImage{},
		"list_style_type":     pr.CounterStyleID{Name: "circle"},
	}))
	assertValidDict(t, "list-style: none inside none", toValidated(pr.Properties{
		"list_style_position": pr.String("inside"),
		"list_style_image":    pr.NoneImage{},
		"list_style_type":     pr.CounterStyleID{Name: "none"},
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "list-style: none inside none none", "invalid")
	assertInvalid(t, "list-style: 1px", "invalid")
	assertInvalid(t, "list-style: circle disc",
		"got multiple type values in a list-style shorthand")
}

// Test the “font“ property.
func TestFont(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "font: 12px My Fancy Font, serif", toValidated(pr.Properties{
		"font_size":   pr.Dimension{Value: 12, Unit: pr.Px}.ToValue(),
		"font_family": pr.Strings{"My Fancy Font", "serif"},
	}))
	assertValidDict(t, `font: small/1.2 "Some Font", serif`, toValidated(pr.Properties{
		"font_size":   pr.SToV("small"),
		"line_height": pr.Dimension{Value: 1.2, Unit: pr.Scalar}.ToValue(),
		"font_family": pr.Strings{"Some Font", "serif"},
	}))
	assertValidDict(t, "font: small-caps italic 700 large serif", toValidated(pr.Properties{
		"font_style":        pr.String("italic"),
		"font_variant_caps": pr.String("small-caps"),
		"font_weight":       pr.IntString{Int: 700},
		"font_size":         pr.SToV("large"),
		"font_family":       pr.Strings{"serif"},
	}))
	assertValidDict(t, "font: small-caps condensed normal 700 large serif", toValidated(pr.Properties{
		// "font_style": String("normal"),  XXX shouldn’t this be here?
		"font_stretch":      pr.String("condensed"),
		"font_variant_caps": pr.String("small-caps"),
		"font_weight":       pr.IntString{Int: 700},
		"font_size":         pr.SToV("large"),
		"font_family":       pr.Strings{"serif"},
	}))
	assertValidDict(t, "font: italic 13px sans-serif", toValidated(pr.Properties{
		"font_style":  pr.String("italic"),
		"font_size":   pr.FToPx(13),
		"font_family": pr.Strings{"sans-serif"},
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, `font-family: "My" Font, serif`, "invalid")
	assertInvalid(t, `font-family: "My" "Font", serif`, "invalid")
	assertInvalid(t, `font-family: "My", 12pt, serif`, "invalid")
	assertInvalid(t, `font: menu`, "system fonts are not supported")
	assertInvalid(t, `font: 12deg My Fancy Font, serif`, "invalid")
	assertInvalid(t, `font: 12px`, "invalid")
	assertInvalid(t, `font: 12px/foo serif`, "invalid")
	assertInvalid(t, `font: 12px "Invalid" family`, "invalid")
	assertInvalid(t, "font: normal normal normal normal normal large serif", "invalid")
	assertInvalid(t, "font: normal small-caps italic 700 condensed large serif", "invalid")
	assertInvalid(t, "font: small-caps italic 700 normal condensed large serif", "invalid")
	assertInvalid(t, "font: small-caps italic 700 condensed normal large serif", "invalid")
}

// Test the “font-variant“ property.
func TestFontVariant(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "font-variant: normal", toValidated(pr.Properties{
		"font_variant_alternates": pr.String("normal"),
		"font_variant_caps":       pr.String("normal"),
		"font_variant_east_asian": pr.SStrings{String: "normal"},
		"font_variant_ligatures":  pr.SStrings{String: "normal"},
		"font_variant_numeric":    pr.SStrings{String: "normal"},
		"font_variant_position":   pr.String("normal"),
	}))
	assertValidDict(t, "font-variant: none", toValidated(pr.Properties{
		"font_variant_alternates": pr.String("normal"),
		"font_variant_caps":       pr.String("normal"),
		"font_variant_east_asian": pr.SStrings{String: "normal"},
		"font_variant_ligatures":  pr.SStrings{String: "none"},
		"font_variant_numeric":    pr.SStrings{String: "normal"},
		"font_variant_position":   pr.String("normal"),
	}))
	assertValidDict(t, "font-variant: historical-forms petite-caps", toValidated(pr.Properties{
		"font_variant_alternates": pr.String("historical-forms"),
		"font_variant_caps":       pr.String("petite-caps"),
	}))
	assertValidDict(t, "font-variant: lining-nums contextual small-caps common-ligatures", toValidated(pr.Properties{
		"font_variant_ligatures": pr.SStrings{Strings: []string{"contextual", "common-ligatures"}},
		"font_variant_numeric":   pr.SStrings{Strings: []string{"lining-nums"}},
		"font_variant_caps":      pr.String("small-caps"),
	}))
	assertValidDict(t, "font-variant: jis78 ruby proportional-width", toValidated(pr.Properties{
		"font_variant_east_asian": pr.SStrings{Strings: []string{"jis78", "ruby", "proportional-width"}},
	}))
	// CSS2-style font-variant
	assertValidDict(t, "font-variant: small-caps", toValidated(pr.Properties{
		"font_variant_caps": pr.String("small-caps"),
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "font-variant: normal normal", "invalid")
	assertInvalid(t, "font-variant: 2", "invalid")
	assertInvalid(t, `font-variant: ""`, "invalid")
	assertInvalid(t, "font-variant: extra", "invalid")
	assertInvalid(t, "font-variant: jis78 jis04", "invalid")
	assertInvalid(t, "font-variant: full-width lining-nums ordinal normal", "invalid")
	assertInvalid(t, "font-variant: diagonal-fractions stacked-fractions", "invalid")
	assertInvalid(t, "font-variant: common-ligatures contextual no-common-ligatures", "invalid")
	assertInvalid(t, "font-variant: sub super", "invalid")
	assertInvalid(t, "font-variant: slashed-zero slashed-zero", "invalid")
}

func TestExpandOverflowWrap(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "overflow-wrap: normal", toValidated(pr.Properties{
		"overflow_wrap": pr.String("normal"),
	}))
	assertValidDict(t, "overflow-wrap: break-word", toValidated(pr.Properties{
		"overflow_wrap": pr.String("break-word"),
	}))
	assertValidDict(t, "overflow-wrap: inherit", map[string]pr.ValidatedProperty{
		"overflow_wrap": inherit,
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "overflow-wrap: none", "invalid")
	assertInvalid(t, "overflow-wrap: normal, break-word", "invalid")
}

func TestExpandWordWrap(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "word-wrap: normal", toValidated(pr.Properties{
		"overflow_wrap": pr.String("normal"),
	}))
	assertValidDict(t, "word-wrap: break-word", toValidated(pr.Properties{
		"overflow_wrap": pr.String("break-word"),
	}))
	assertValidDict(t, "word-wrap: inherit", map[string]pr.ValidatedProperty{
		"overflow_wrap": inherit,
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "word-wrap: none", "invalid")
	assertInvalid(t, "word-wrap: normal, break-word", "invalid")
}

func TestExpandTextDecoration(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "text-decoration: none", toValidated(pr.Properties{
		"text_decoration_line": pr.Decorations{},
	}))
	assertValidDict(t, "text-decoration: overline", toValidated(pr.Properties{
		"text_decoration_line": pr.Decorations(utils.NewSet("overline")),
	}))
	assertValidDict(t, "text-decoration: overline solid", toValidated(pr.Properties{
		"text_decoration_line":  pr.Decorations(utils.NewSet("overline")),
		"text_decoration_style": pr.String("solid"),
	}))
	assertValidDict(t, "text-decoration: overline blink line-through", toValidated(pr.Properties{
		"text_decoration_line": pr.Decorations(utils.NewSet("blink", "line-through", "overline")),
	}))
	assertValidDict(t, "text-decoration: red", toValidated(pr.Properties{
		"text_decoration_color": pr.NewColor(1, 0, 0, 1),
	}))
	assertValidDict(t, "text-decoration: inherit", map[string]pr.ValidatedProperty{
		"text_decoration_color": inherit,
		"text_decoration_line":  inherit,
		"text_decoration_style": inherit,
	})

	assertInvalid(t, "text-decoration: solid solid", "invalid")
	assertInvalid(t, "text-decoration: red red", "invalid")
	assertInvalid(t, "text-decoration: 1px", "invalid")
	assertInvalid(t, "text-decoration: underline none", "invalid")
	assertInvalid(t, "text-decoration: none none", "invalid")

	capt.AssertNoLogs(t)
}

func TestExpandFlex(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "flex: auto", toValidated(pr.Properties{
		"flex_grow":   pr.Float(1),
		"flex_shrink": pr.Float(1),
		"flex_basis":  pr.SToV("auto"),
	}))
	assertValidDict(t, "flex: none", toValidated(pr.Properties{
		"flex_grow":   pr.Float(0),
		"flex_shrink": pr.Float(0),
		"flex_basis":  pr.SToV("auto"),
	}))
	assertValidDict(t, "flex: 10", toValidated(pr.Properties{
		"flex_grow":   pr.Float(10),
		"flex_shrink": pr.Float(1),
		"flex_basis":  pr.ZeroPixels.ToValue(),
	}))
	assertValidDict(t, "flex: 2 2", toValidated(pr.Properties{
		"flex_grow":   pr.Float(2),
		"flex_shrink": pr.Float(2),
		"flex_basis":  pr.ZeroPixels.ToValue(),
	}))
	assertValidDict(t, "flex: 2 2 1px", toValidated(pr.Properties{
		"flex_grow":   pr.Float(2),
		"flex_shrink": pr.Float(2),
		"flex_basis":  pr.Dimension{Value: 1, Unit: pr.Px}.ToValue(),
	}))
	assertValidDict(t, "flex: 2 2 auto", toValidated(pr.Properties{
		"flex_grow":   pr.Float(2),
		"flex_shrink": pr.Float(2),
		"flex_basis":  pr.SToV("auto"),
	}))
	assertValidDict(t, "flex: 2 auto", toValidated(pr.Properties{
		"flex_grow":   pr.Float(2),
		"flex_shrink": pr.Float(1),
		"flex_basis":  pr.SToV("auto"),
	}))
	assertValidDict(t, "flex: inherit", map[string]pr.ValidatedProperty{
		"flex_grow":   inherit,
		"flex_shrink": inherit,
		"flex_basis":  inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, "flex: auto 0 0 0", "invalid")
	assertInvalid(t, "flex: 1px 2px", "invalid")
	assertInvalid(t, "flex: auto auto", "invalid")
	assertInvalid(t, "flex: auto 1 auto", "invalid")
}

func TestExpandFlexFlow(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "flex-flow: column", toValidated(pr.Properties{
		"flex_direction": pr.String("column"),
	}))
	assertValidDict(t, "flex-flow: wrap", toValidated(pr.Properties{
		"flex_wrap": pr.String("wrap"),
	}))
	assertValidDict(t, "flex-flow: wrap column", toValidated(pr.Properties{
		"flex_direction": pr.String("column"),
		"flex_wrap":      pr.String("wrap"),
	}))
	assertValidDict(t, "flex-flow: row wrap", toValidated(pr.Properties{
		"flex_direction": pr.String("row"),
		"flex_wrap":      pr.String("wrap"),
	}))
	assertValidDict(t, "flex-flow: inherit", map[string]pr.ValidatedProperty{
		"flex_direction": inherit,
		"flex_wrap":      inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, "flex-flow: 1px", "invalid")
	assertInvalid(t, "flex-flow: wrap 1px", "invalid")
	assertInvalid(t, "flex-flow: row row", "invalid")
	assertInvalid(t, "flex-flow: wrap nowrap", "invalid")
	assertInvalid(t, "flex-flow: column wrap nowrap row", "invalid")
}

func TestExpandPageBreak(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "page-break-after: left", toValidated(pr.Properties{
		"break_after": pr.String("left"),
	}))
	assertValidDict(t, "page-break-before: always", toValidated(pr.Properties{
		"break_before": pr.String("page"),
	}))
	assertValidDict(t, "page-break-after: inherit", map[string]pr.ValidatedProperty{
		"break_after": inherit,
	})
	assertValidDict(t, "page-break-before: inherit", map[string]pr.ValidatedProperty{
		"break_before": inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, "page-break-after: top", "invalid")
	assertInvalid(t, "page-break-before: 1px", "invalid")
}

func TestExpandPageBreakInside(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "page-break-inside: avoid", toValidated(pr.Properties{
		"break_inside": pr.String("avoid"),
	}))
	assertValidDict(t, "page-break-inside: inherit", map[string]pr.ValidatedProperty{
		"break_inside": inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, "page-break-inside: top", "invalid")
}

func TestExpandColumns(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "columns: 1em", toValidated(pr.Properties{
		"column_width": pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"column_count": pr.IntString{String: "auto"},
	}))
	assertValidDict(t, "columns: auto", toValidated(pr.Properties{
		"column_width": pr.SToV("auto"),
		"column_count": pr.IntString{String: "auto"},
	}))
	assertValidDict(t, "columns: auto auto", toValidated(pr.Properties{
		"column_width": pr.SToV("auto"),
		"column_count": pr.IntString{String: "auto"},
	}))

	capt.AssertNoLogs(t)

	assertInvalid(t, "columns: 1px 2px", "invalid")
	assertInvalid(t, "columns: auto auto auto", "multiple")
}

func TestLineClamp(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "line-clamp: none", toValidated(pr.Properties{
		"max_lines":      pr.TaggedInt{Tag: pr.None},
		"continue":       pr.String("auto"),
		"block_ellipsis": pr.TaggedString{Tag: pr.None},
	}))
	assertValidDict(t, "line-clamp: 2", toValidated(pr.Properties{
		"max_lines":      pr.TaggedInt{I: 2},
		"continue":       pr.String("discard"),
		"block_ellipsis": pr.TaggedString{Tag: pr.Auto},
	}))
	assertValidDict(t, `line-clamp: 3 "…"`, toValidated(pr.Properties{
		"max_lines":      pr.TaggedInt{I: 3},
		"continue":       pr.String("discard"),
		"block_ellipsis": pr.TaggedString{S: "…"},
	}))
	assertValidDict(t, "line-clamp: inherit", map[string]pr.ValidatedProperty{
		"max_lines":      inherit,
		"continue":       inherit,
		"block_ellipsis": inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, `line-clamp: none none none`, "invalid")
	assertInvalid(t, `line-clamp: 1px`, "invalid")
	assertInvalid(t, `line-clamp: 0 "…"`, "invalid")
	assertInvalid(t, `line-clamp: 1px 2px`, "invalid")
}

func TestExpandTextAlign(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "text-align: start", toValidated(pr.Properties{
		"text_align_all":  pr.String("start"),
		"text_align_last": pr.String("start"),
	}))
	assertValidDict(t, "text-align: right", toValidated(pr.Properties{
		"text_align_all":  pr.String("right"),
		"text_align_last": pr.String("right"),
	}))
	assertValidDict(t, "text-align: justify", toValidated(pr.Properties{
		"text_align_all":  pr.String("justify"),
		"text_align_last": pr.String("start"),
	}))
	assertValidDict(t, "text-align: justify-all", toValidated(pr.Properties{
		"text_align_all":  pr.String("justify"),
		"text_align_last": pr.String("justify"),
	}))
	assertValidDict(t, "text-align: inherit", map[string]pr.ValidatedProperty{
		"text_align_all":  inherit,
		"text_align_last": inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, "text-align: none", "invalid")
	assertInvalid(t, "text-align: start end", "invalid")
	assertInvalid(t, "text-align: 1", "invalid")
	assertInvalid(t, `text-align: left left`, "invalid")
	assertInvalid(t, `text-align: top`, "invalid")
	assertInvalid(t, `text-align: "right"`, "invalid")
	assertInvalid(t, `text-align: 1px`, "invalid")
}

// Helper checking the background pr.
func assertBackground(t *testing.T, css string, expected map[string]pr.ValidatedProperty) {
	expanded := expandToDict(t, "background: "+css, "")
	col, in := expected["background_color"]
	if !in {
		col = pr.AsCascaded(pr.InitialValues["background_color"]).AsValidated()
	}
	if !reflect.DeepEqual(expanded["background_color"], col) {
		t.Fatalf("expected %v got %v", col, expanded["background_color"])
	}
	delete(expanded, "background_color")
	delete(expected, "background_color")

	bi := expanded["background_image"]
	for name, value := range expected {
		if !reflect.DeepEqual(expanded[name], value) {
			t.Fatalf("for %s expected %v got %v", name, value, expanded[name])
		}
		delete(expanded, name)
		delete(expected, name)
	}

	if len(expanded) == 0 {
		return
	}

	nbLayers := len(bi.ToCascaded().ToCSS().(pr.Images))
	for name, value := range expanded {
		initv := pr.InitialValues[name].(repeatable)
		ref := pr.AsCascaded(initv.Repeat(nbLayers)).AsValidated()
		if !reflect.DeepEqual(value, ref) {
			t.Fatalf("expected %v got %v", ref, value)
		}
	}
}

// Test the “background“ property.
func TestExpandBackground(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertBackground(t, "none", toValidated(pr.Properties{}))
	assertBackground(t, "red", toValidated(pr.Properties{
		"background_color": pr.NewColor(1, 0, 0, 1),
	}))
	assertBackground(t, "url(lipsum.png)", toValidated(pr.Properties{
		"background_image": pr.Images{pr.UrlImage("https://weasyprint.org/foo/lipsum.png")},
	}))
	assertBackground(t, "no-repeat", toValidated(pr.Properties{
		"background_repeat": pr.Repeats{{"no-repeat", "no-repeat"}},
	}))
	assertBackground(t, "fixed", toValidated(pr.Properties{
		"background_attachment": pr.Strings{"fixed"},
	}))
	assertBackground(t, "repeat no-repeat fixed", toValidated(pr.Properties{
		"background_repeat":     pr.Repeats{{"repeat", "no-repeat"}},
		"background_attachment": pr.Strings{"fixed"},
	}))
	assertBackground(t, "inherit", map[string]pr.ValidatedProperty{
		"background_repeat":     inherit,
		"background_attachment": inherit,
		"background_image":      inherit,
		"background_position":   inherit,
		"background_size":       inherit,
		"background_clip":       inherit,
		"background_origin":     inherit,
		"background_color":      inherit,
	})
	assertBackground(t, "top", toValidated(pr.Properties{
		"background_position": pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top right", toValidated(pr.Properties{
		"background_position": pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 100, Unit: pr.Perc}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top right 20px", toValidated(pr.Properties{
		"background_position": pr.Centers{{OriginX: "right", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 20, Unit: pr.Px}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top 1% right 20px", toValidated(pr.Properties{
		"background_position": pr.Centers{{OriginX: "right", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 20, Unit: pr.Px}, pr.Dimension{Value: 1, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top no-repeat", toValidated(pr.Properties{
		"background_repeat":   pr.Repeats{{"no-repeat", "no-repeat"}},
		"background_position": pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top right no-repeat", toValidated(pr.Properties{
		"background_repeat":   pr.Repeats{{"no-repeat", "no-repeat"}},
		"background_position": pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 100, Unit: pr.Perc}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top right 20px no-repeat", toValidated(pr.Properties{
		"background_repeat":   pr.Repeats{{"no-repeat", "no-repeat"}},
		"background_position": pr.Centers{{OriginX: "right", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 20, Unit: pr.Px}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top 1% right 20px no-repeat", toValidated(pr.Properties{
		"background_repeat":   pr.Repeats{{"no-repeat", "no-repeat"}},
		"background_position": pr.Centers{{OriginX: "right", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 20, Unit: pr.Px}, pr.Dimension{Value: 1, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "url(bar) #f00 repeat-y center left fixed", toValidated(pr.Properties{
		"background_color":      pr.NewColor(1, 0, 0, 1),
		"background_image":      pr.Images{pr.UrlImage("https://weasyprint.org/foo/bar")},
		"background_repeat":     pr.Repeats{{"no-repeat", "repeat"}},
		"background_attachment": pr.Strings{"fixed"},
		"background_position":   pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 0, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "#00f 10% 200px", toValidated(pr.Properties{
		"background_color":    pr.NewColor(0, 0, 1, 1),
		"background_position": pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 10, Unit: pr.Perc}, pr.Dimension{Value: 200, Unit: pr.Px}}}},
	}))
	assertBackground(t, "right 78px fixed", toValidated(pr.Properties{
		"background_attachment": pr.Strings{"fixed"},
		"background_position":   pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 100, Unit: pr.Perc}, pr.Dimension{Value: 78, Unit: pr.Px}}}},
	}))
	assertBackground(t, "center / cover red", toValidated(pr.Properties{
		"background_size":     pr.Sizes{{String: "cover"}},
		"background_position": pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}}},
		"background_color":    pr.NewColor(1, 0, 0, 1),
	}))
	assertBackground(t, "center / auto red", toValidated(pr.Properties{
		"background_size":     pr.Sizes{{Width: pr.SToV("auto"), Height: pr.SToV("auto")}},
		"background_position": pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}}},
		"background_color":    pr.NewColor(1, 0, 0, 1),
	}))
	assertBackground(t, "center / 42px", toValidated(pr.Properties{
		"background_size":     pr.Sizes{{Width: pr.Dimension{Value: 42, Unit: pr.Px}.ToValue(), Height: pr.SToV("auto")}},
		"background_position": pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "center / 7% 4em", toValidated(pr.Properties{
		"background_size":     pr.Sizes{{Width: pr.Dimension{Value: 7, Unit: pr.Perc}.ToValue(), Height: pr.Dimension{Value: 4, Unit: pr.Em}.ToValue()}},
		"background_position": pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "red content-box", toValidated(pr.Properties{
		"background_color":  pr.NewColor(1, 0, 0, 1),
		"background_origin": pr.Strings{"content-box"},
		"background_clip":   pr.Strings{"content-box"},
	}))
	assertBackground(t, "red border-box content-box", toValidated(pr.Properties{
		"background_color":  pr.NewColor(1, 0, 0, 1),
		"background_origin": pr.Strings{"border-box"},
		"background_clip":   pr.Strings{"content-box"},
	}))
	assertBackground(t, "url(bar) center, no-repeat", toValidated(pr.Properties{
		"background_color": pr.NewColor(0, 0, 0, 0),
		"background_image": pr.Images{pr.UrlImage("https://weasyprint.org/foo/bar"), pr.NoneImage{}},
		"background_position": pr.Centers{
			{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}},
			{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 0, Unit: pr.Perc}, pr.Dimension{Value: 0, Unit: pr.Perc}}},
		},
		"background_repeat": pr.Repeats{{"repeat", "repeat"}, {"no-repeat", "no-repeat"}},
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "background: 10px lipsum", "invalid")
	assertInvalid(t, "background-position: 10px lipsum", "invalid")
	assertInvalid(t, "background: content-box red content-box", "invalid")
	assertInvalid(t, "background-image: inexistent-gradient(blue, green)", "invalid")
	// Color must be in the last layer :
	assertInvalid(t, "background: red, url(foo)", "invalid")
}
