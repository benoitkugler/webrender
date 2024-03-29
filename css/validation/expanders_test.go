package validation

import (
	"reflect"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils"
	"github.com/benoitkugler/webrender/utils/testutils"
)

var inherit = pr.Inherit.AsCascaded().AsValidated()

// Test the 4-value pr.
func TestExpandFourSides(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "margin: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PMarginTop:    inherit,
		pr.PMarginRight:  inherit,
		pr.PMarginBottom: inherit,
		pr.PMarginLeft:   inherit,
	})
	assertValidDict(t, "margin: 1em", toValidated(pr.Properties{
		pr.PMarginTop:    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		pr.PMarginRight:  pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		pr.PMarginBottom: pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		pr.PMarginLeft:   pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
	}))
	assertValidDict(t, "margin: -1em auto 20%", toValidated(pr.Properties{
		pr.PMarginTop:    pr.Dimension{Value: -1, Unit: pr.Em}.ToValue(),
		pr.PMarginRight:  pr.SToV("auto"),
		pr.PMarginBottom: pr.Dimension{Value: 20, Unit: pr.Perc}.ToValue(),
		pr.PMarginLeft:   pr.SToV("auto"),
	}))
	assertValidDict(t, "padding: 1em 0", toValidated(pr.Properties{
		pr.PPaddingTop:    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		pr.PPaddingRight:  pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
		pr.PPaddingBottom: pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		pr.PPaddingLeft:   pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
	}))
	assertValidDict(t, "padding: 1em 0 2%", toValidated(pr.Properties{
		pr.PPaddingTop:    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		pr.PPaddingRight:  pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
		pr.PPaddingBottom: pr.Dimension{Value: 2, Unit: pr.Perc}.ToValue(),
		pr.PPaddingLeft:   pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
	}))
	assertValidDict(t, "padding: 1em 0 2em 5px", toValidated(pr.Properties{
		pr.PPaddingTop:    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		pr.PPaddingRight:  pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
		pr.PPaddingBottom: pr.Dimension{Value: 2, Unit: pr.Em}.ToValue(),
		pr.PPaddingLeft:   pr.Dimension{Value: 5, Unit: pr.Px}.ToValue(),
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
		pr.PBorderTopWidth: pr.Dimension{Value: 3, Unit: pr.Px}.ToValue(),
		pr.PBorderTopStyle: pr.String("dotted"),
		pr.PBorderTopColor: pr.NewColor(1, 0, 0, 1), // red
	}))
	assertValidDict(t, "border-top: 3px dotted", toValidated(pr.Properties{
		pr.PBorderTopWidth: pr.Dimension{Value: 3, Unit: pr.Px}.ToValue(),
		pr.PBorderTopStyle: pr.String("dotted"),
	}))
	assertValidDict(t, "border-top: 3px red", toValidated(pr.Properties{
		pr.PBorderTopWidth: pr.Dimension{Value: 3, Unit: pr.Px}.ToValue(),
		pr.PBorderTopColor: pr.NewColor(1, 0, 0, 1), // red
	}))
	assertValidDict(t, "border-top: solid", toValidated(pr.Properties{
		pr.PBorderTopStyle: pr.String("solid"),
	}))
	assertValidDict(t, "border: 6px dashed lime", toValidated(pr.Properties{
		pr.PBorderTopWidth: pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		pr.PBorderTopStyle: pr.String("dashed"),
		pr.PBorderTopColor: pr.NewColor(0, 1, 0, 1), // lime

		pr.PBorderLeftWidth: pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		pr.PBorderLeftStyle: pr.String("dashed"),
		pr.PBorderLeftColor: pr.NewColor(0, 1, 0, 1), // lime

		pr.PBorderBottomWidth: pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		pr.PBorderBottomStyle: pr.String("dashed"),
		pr.PBorderBottomColor: pr.NewColor(0, 1, 0, 1), // lime

		pr.PBorderRightWidth: pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		pr.PBorderRightStyle: pr.String("dashed"),
		pr.PBorderRightColor: pr.NewColor(0, 1, 0, 1), // lime
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "border: 6px dashed left", "invalid")
}

func TestExpandBorderRadius(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "border-radius: 1px", toValidated(pr.Properties{
		pr.PBorderTopLeftRadius:     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		pr.PBorderTopRightRadius:    pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		pr.PBorderBottomRightRadius: pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		pr.PBorderBottomLeftRadius:  pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
	}))
	assertValidDict(t, "border-radius: 1px 2em", toValidated(pr.Properties{
		pr.PBorderTopLeftRadius:     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		pr.PBorderTopRightRadius:    pr.Point{{Value: 2, Unit: pr.Em}, {Value: 2, Unit: pr.Em}},
		pr.PBorderBottomRightRadius: pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		pr.PBorderBottomLeftRadius:  pr.Point{{Value: 2, Unit: pr.Em}, {Value: 2, Unit: pr.Em}},
	}))
	assertValidDict(t, "border-radius: 1px / 2em", toValidated(pr.Properties{
		pr.PBorderTopLeftRadius:     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
		pr.PBorderTopRightRadius:    pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
		pr.PBorderBottomRightRadius: pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
		pr.PBorderBottomLeftRadius:  pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
	}))
	assertValidDict(t, "border-radius: 1px 3px / 2em 4%", toValidated(pr.Properties{
		pr.PBorderTopLeftRadius:     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
		pr.PBorderTopRightRadius:    pr.Point{{Value: 3, Unit: pr.Px}, {Value: 4, Unit: pr.Perc}},
		pr.PBorderBottomRightRadius: pr.Point{{Value: 1, Unit: pr.Px}, {Value: 2, Unit: pr.Em}},
		pr.PBorderBottomLeftRadius:  pr.Point{{Value: 3, Unit: pr.Px}, {Value: 4, Unit: pr.Perc}},
	}))
	assertValidDict(t, "border-radius: 1px 2em 3%", toValidated(pr.Properties{
		pr.PBorderTopLeftRadius:     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		pr.PBorderTopRightRadius:    pr.Point{{Value: 2, Unit: pr.Em}, {Value: 2, Unit: pr.Em}},
		pr.PBorderBottomRightRadius: pr.Point{{Value: 3, Unit: pr.Perc}, {Value: 3, Unit: pr.Perc}},
		pr.PBorderBottomLeftRadius:  pr.Point{{Value: 2, Unit: pr.Em}, {Value: 2, Unit: pr.Em}},
	}))
	assertValidDict(t, "border-radius: 1px 2em 3% 4rem", toValidated(pr.Properties{
		pr.PBorderTopLeftRadius:     pr.Point{{Value: 1, Unit: pr.Px}, {Value: 1, Unit: pr.Px}},
		pr.PBorderTopRightRadius:    pr.Point{{Value: 2, Unit: pr.Em}, {Value: 2, Unit: pr.Em}},
		pr.PBorderBottomRightRadius: pr.Point{{Value: 3, Unit: pr.Perc}, {Value: 3, Unit: pr.Perc}},
		pr.PBorderBottomLeftRadius:  pr.Point{{Value: 4, Unit: pr.Rem}, {Value: 4, Unit: pr.Rem}},
	}))
	assertValidDict(t, "border-radius: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PBorderTopLeftRadius:     inherit,
		pr.PBorderTopRightRadius:    inherit,
		pr.PBorderBottomRightRadius: inherit,
		pr.PBorderBottomLeftRadius:  inherit,
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
	assertValidDict(t, "list-style: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PListStylePosition: inherit,
		pr.PListStyleImage:    inherit,
		pr.PListStyleType:     inherit,
	})
	assertValidDict(t, "list-style: url(../bar/lipsum.png)", toValidated(pr.Properties{
		pr.PListStyleImage: pr.UrlImage("https://weasyprint.org/bar/lipsum.png"),
	}))
	assertValidDict(t, "list-style: square", toValidated(pr.Properties{
		pr.PListStyleType: pr.CounterStyleID{Name: "square"},
	}))
	assertValidDict(t, "list-style: circle inside", toValidated(pr.Properties{
		pr.PListStylePosition: pr.String("inside"),
		pr.PListStyleType:     pr.CounterStyleID{Name: "circle"},
	}))
	assertValidDict(t, "list-style: none circle inside", toValidated(pr.Properties{
		pr.PListStylePosition: pr.String("inside"),
		pr.PListStyleImage:    pr.NoneImage{},
		pr.PListStyleType:     pr.CounterStyleID{Name: "circle"},
	}))
	assertValidDict(t, "list-style: none inside none", toValidated(pr.Properties{
		pr.PListStylePosition: pr.String("inside"),
		pr.PListStyleImage:    pr.NoneImage{},
		pr.PListStyleType:     pr.CounterStyleID{Name: "none"},
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
		pr.PFontSize:   pr.Dimension{Value: 12, Unit: pr.Px}.ToValue(),
		pr.PFontFamily: pr.Strings{"My Fancy Font", "serif"},
	}))
	assertValidDict(t, `font: small/1.2 "Some Font", serif`, toValidated(pr.Properties{
		pr.PFontSize:   pr.SToV("small"),
		pr.PLineHeight: pr.Dimension{Value: 1.2, Unit: pr.Scalar}.ToValue(),
		pr.PFontFamily: pr.Strings{"Some Font", "serif"},
	}))
	assertValidDict(t, "font: small-caps italic 700 large serif", toValidated(pr.Properties{
		pr.PFontStyle:       pr.String("italic"),
		pr.PFontVariantCaps: pr.String("small-caps"),
		pr.PFontWeight:      pr.IntString{Int: 700},
		pr.PFontSize:        pr.SToV("large"),
		pr.PFontFamily:      pr.Strings{"serif"},
	}))
	assertValidDict(t, "font: small-caps condensed normal 700 large serif", toValidated(pr.Properties{
		// "font_style": String("normal"),  XXX shouldn’t this be here?
		pr.PFontStretch:     pr.String("condensed"),
		pr.PFontVariantCaps: pr.String("small-caps"),
		pr.PFontWeight:      pr.IntString{Int: 700},
		pr.PFontSize:        pr.SToV("large"),
		pr.PFontFamily:      pr.Strings{"serif"},
	}))
	assertValidDict(t, "font: italic 13px sans-serif", toValidated(pr.Properties{
		pr.PFontStyle:  pr.String("italic"),
		pr.PFontSize:   pr.FToPx(13),
		pr.PFontFamily: pr.Strings{"sans-serif"},
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
		pr.PFontVariantAlternates: pr.String("normal"),
		pr.PFontVariantCaps:       pr.String("normal"),
		pr.PFontVariantEastAsian:  pr.SStrings{String: "normal"},
		pr.PFontVariantLigatures:  pr.SStrings{String: "normal"},
		pr.PFontVariantNumeric:    pr.SStrings{String: "normal"},
		pr.PFontVariantPosition:   pr.String("normal"),
	}))
	assertValidDict(t, "font-variant: none", toValidated(pr.Properties{
		pr.PFontVariantAlternates: pr.String("normal"),
		pr.PFontVariantCaps:       pr.String("normal"),
		pr.PFontVariantEastAsian:  pr.SStrings{String: "normal"},
		pr.PFontVariantLigatures:  pr.SStrings{String: "none"},
		pr.PFontVariantNumeric:    pr.SStrings{String: "normal"},
		pr.PFontVariantPosition:   pr.String("normal"),
	}))
	assertValidDict(t, "font-variant: historical-forms petite-caps", toValidated(pr.Properties{
		pr.PFontVariantAlternates: pr.String("historical-forms"),
		pr.PFontVariantCaps:       pr.String("petite-caps"),
	}))
	assertValidDict(t, "font-variant: lining-nums contextual small-caps common-ligatures", toValidated(pr.Properties{
		pr.PFontVariantLigatures: pr.SStrings{Strings: []string{"contextual", "common-ligatures"}},
		pr.PFontVariantNumeric:   pr.SStrings{Strings: []string{"lining-nums"}},
		pr.PFontVariantCaps:      pr.String("small-caps"),
	}))
	assertValidDict(t, "font-variant: jis78 ruby proportional-width", toValidated(pr.Properties{
		pr.PFontVariantEastAsian: pr.SStrings{Strings: []string{"jis78", "ruby", "proportional-width"}},
	}))
	// CSS2-style font-variant
	assertValidDict(t, "font-variant: small-caps", toValidated(pr.Properties{
		pr.PFontVariantCaps: pr.String("small-caps"),
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
		pr.POverflowWrap: pr.String("normal"),
	}))
	assertValidDict(t, "overflow-wrap: break-word", toValidated(pr.Properties{
		pr.POverflowWrap: pr.String("break-word"),
	}))
	assertValidDict(t, "overflow-wrap: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.POverflowWrap: inherit,
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "overflow-wrap: none", "invalid")
	assertInvalid(t, "overflow-wrap: normal, break-word", "invalid")
}

func TestExpandWordWrap(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "word-wrap: normal", toValidated(pr.Properties{
		pr.POverflowWrap: pr.String("normal"),
	}))
	assertValidDict(t, "word-wrap: break-word", toValidated(pr.Properties{
		pr.POverflowWrap: pr.String("break-word"),
	}))
	assertValidDict(t, "word-wrap: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.POverflowWrap: inherit,
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "word-wrap: none", "invalid")
	assertInvalid(t, "word-wrap: normal, break-word", "invalid")
}

func TestExpandTextDecoration(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "text-decoration: none", toValidated(pr.Properties{
		pr.PTextDecorationLine: pr.Decorations{},
	}))
	assertValidDict(t, "text-decoration: overline", toValidated(pr.Properties{
		pr.PTextDecorationLine: pr.Decorations(utils.NewSet("overline")),
	}))
	assertValidDict(t, "text-decoration: overline solid", toValidated(pr.Properties{
		pr.PTextDecorationLine:  pr.Decorations(utils.NewSet("overline")),
		pr.PTextDecorationStyle: pr.String("solid"),
	}))
	assertValidDict(t, "text-decoration: overline blink line-through", toValidated(pr.Properties{
		pr.PTextDecorationLine: pr.Decorations(utils.NewSet("blink", "line-through", "overline")),
	}))
	assertValidDict(t, "text-decoration: red", toValidated(pr.Properties{
		pr.PTextDecorationColor: pr.NewColor(1, 0, 0, 1),
	}))
	assertValidDict(t, "text-decoration: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PTextDecorationColor: inherit,
		pr.PTextDecorationLine:  inherit,
		pr.PTextDecorationStyle: inherit,
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
		pr.PFlexGrow:   pr.Float(1),
		pr.PFlexShrink: pr.Float(1),
		pr.PFlexBasis:  pr.SToV("auto"),
	}))
	assertValidDict(t, "flex: none", toValidated(pr.Properties{
		pr.PFlexGrow:   pr.Float(0),
		pr.PFlexShrink: pr.Float(0),
		pr.PFlexBasis:  pr.SToV("auto"),
	}))
	assertValidDict(t, "flex: 10", toValidated(pr.Properties{
		pr.PFlexGrow:   pr.Float(10),
		pr.PFlexShrink: pr.Float(1),
		pr.PFlexBasis:  pr.ZeroPixels.ToValue(),
	}))
	assertValidDict(t, "flex: 2 2", toValidated(pr.Properties{
		pr.PFlexGrow:   pr.Float(2),
		pr.PFlexShrink: pr.Float(2),
		pr.PFlexBasis:  pr.ZeroPixels.ToValue(),
	}))
	assertValidDict(t, "flex: 2 2 1px", toValidated(pr.Properties{
		pr.PFlexGrow:   pr.Float(2),
		pr.PFlexShrink: pr.Float(2),
		pr.PFlexBasis:  pr.Dimension{Value: 1, Unit: pr.Px}.ToValue(),
	}))
	assertValidDict(t, "flex: 2 2 auto", toValidated(pr.Properties{
		pr.PFlexGrow:   pr.Float(2),
		pr.PFlexShrink: pr.Float(2),
		pr.PFlexBasis:  pr.SToV("auto"),
	}))
	assertValidDict(t, "flex: 2 auto", toValidated(pr.Properties{
		pr.PFlexGrow:   pr.Float(2),
		pr.PFlexShrink: pr.Float(1),
		pr.PFlexBasis:  pr.SToV("auto"),
	}))
	assertValidDict(t, "flex: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PFlexGrow:   inherit,
		pr.PFlexShrink: inherit,
		pr.PFlexBasis:  inherit,
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
		pr.PFlexDirection: pr.String("column"),
	}))
	assertValidDict(t, "flex-flow: wrap", toValidated(pr.Properties{
		pr.PFlexWrap: pr.String("wrap"),
	}))
	assertValidDict(t, "flex-flow: wrap column", toValidated(pr.Properties{
		pr.PFlexDirection: pr.String("column"),
		pr.PFlexWrap:      pr.String("wrap"),
	}))
	assertValidDict(t, "flex-flow: row wrap", toValidated(pr.Properties{
		pr.PFlexDirection: pr.String("row"),
		pr.PFlexWrap:      pr.String("wrap"),
	}))
	assertValidDict(t, "flex-flow: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PFlexDirection: inherit,
		pr.PFlexWrap:      inherit,
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
		pr.PBreakAfter: pr.String("left"),
	}))
	assertValidDict(t, "page-break-before: always", toValidated(pr.Properties{
		pr.PBreakBefore: pr.String("page"),
	}))
	assertValidDict(t, "page-break-after: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PBreakAfter: inherit,
	})
	assertValidDict(t, "page-break-before: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PBreakBefore: inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, "page-break-after: top", "invalid")
	assertInvalid(t, "page-break-before: 1px", "invalid")
}

func TestExpandPageBreakInside(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "page-break-inside: avoid", toValidated(pr.Properties{
		pr.PBreakInside: pr.String("avoid"),
	}))
	assertValidDict(t, "page-break-inside: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PBreakInside: inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, "page-break-inside: top", "invalid")
}

func TestExpandColumns(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "columns: 1em", toValidated(pr.Properties{
		pr.PColumnWidth: pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		pr.PColumnCount: pr.IntString{String: "auto"},
	}))
	assertValidDict(t, "columns: auto", toValidated(pr.Properties{
		pr.PColumnWidth: pr.SToV("auto"),
		pr.PColumnCount: pr.IntString{String: "auto"},
	}))
	assertValidDict(t, "columns: auto auto", toValidated(pr.Properties{
		pr.PColumnWidth: pr.SToV("auto"),
		pr.PColumnCount: pr.IntString{String: "auto"},
	}))

	capt.AssertNoLogs(t)

	assertInvalid(t, "columns: 1px 2px", "invalid")
	assertInvalid(t, "columns: auto auto auto", "multiple")
}

func TestLineClamp(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "line-clamp: none", toValidated(pr.Properties{
		pr.PMaxLines:      pr.TaggedInt{Tag: pr.None},
		pr.PContinue:      pr.String("auto"),
		pr.PBlockEllipsis: pr.TaggedString{Tag: pr.None},
	}))
	assertValidDict(t, "line-clamp: 2", toValidated(pr.Properties{
		pr.PMaxLines:      pr.TaggedInt{I: 2},
		pr.PContinue:      pr.String("discard"),
		pr.PBlockEllipsis: pr.TaggedString{Tag: pr.Auto},
	}))
	assertValidDict(t, `line-clamp: 3 "…"`, toValidated(pr.Properties{
		pr.PMaxLines:      pr.TaggedInt{I: 3},
		pr.PContinue:      pr.String("discard"),
		pr.PBlockEllipsis: pr.TaggedString{S: "…"},
	}))
	assertValidDict(t, "line-clamp: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PMaxLines:      inherit,
		pr.PContinue:      inherit,
		pr.PBlockEllipsis: inherit,
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
		pr.PTextAlignAll:  pr.String("start"),
		pr.PTextAlignLast: pr.String("start"),
	}))
	assertValidDict(t, "text-align: right", toValidated(pr.Properties{
		pr.PTextAlignAll:  pr.String("right"),
		pr.PTextAlignLast: pr.String("right"),
	}))
	assertValidDict(t, "text-align: justify", toValidated(pr.Properties{
		pr.PTextAlignAll:  pr.String("justify"),
		pr.PTextAlignLast: pr.String("start"),
	}))
	assertValidDict(t, "text-align: justify-all", toValidated(pr.Properties{
		pr.PTextAlignAll:  pr.String("justify"),
		pr.PTextAlignLast: pr.String("justify"),
	}))
	assertValidDict(t, "text-align: inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PTextAlignAll:  inherit,
		pr.PTextAlignLast: inherit,
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
func assertBackground(t *testing.T, css string, expected map[pr.KnownProp]pr.ValidatedProperty) {
	expanded := expandToDict(t, "background: "+css, "")
	col, in := expected[pr.PBackgroundColor]
	if !in {
		col = pr.AsCascaded(pr.InitialValues[pr.PBackgroundColor]).AsValidated()
	}
	if !reflect.DeepEqual(expanded[pr.PBackgroundColor], col) {
		t.Fatalf("expected %v got %v", col, expanded[pr.PBackgroundColor])
	}
	delete(expanded, pr.PBackgroundColor)
	delete(expected, pr.PBackgroundColor)

	bi := expanded[pr.PBackgroundImage]
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
		initv := pr.InitialValues[pr.KnownProp(name)].(repeatable)
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
		pr.PBackgroundColor: pr.NewColor(1, 0, 0, 1),
	}))
	assertBackground(t, "url(lipsum.png)", toValidated(pr.Properties{
		pr.PBackgroundImage: pr.Images{pr.UrlImage("https://weasyprint.org/foo/lipsum.png")},
	}))
	assertBackground(t, "no-repeat", toValidated(pr.Properties{
		pr.PBackgroundRepeat: pr.Repeats{{"no-repeat", "no-repeat"}},
	}))
	assertBackground(t, "fixed", toValidated(pr.Properties{
		pr.PBackgroundAttachment: pr.Strings{"fixed"},
	}))
	assertBackground(t, "repeat no-repeat fixed", toValidated(pr.Properties{
		pr.PBackgroundRepeat:     pr.Repeats{{"repeat", "no-repeat"}},
		pr.PBackgroundAttachment: pr.Strings{"fixed"},
	}))
	assertBackground(t, "inherit", map[pr.KnownProp]pr.ValidatedProperty{
		pr.PBackgroundRepeat:     inherit,
		pr.PBackgroundAttachment: inherit,
		pr.PBackgroundImage:      inherit,
		pr.PBackgroundPosition:   inherit,
		pr.PBackgroundSize:       inherit,
		pr.PBackgroundClip:       inherit,
		pr.PBackgroundOrigin:     inherit,
		pr.PBackgroundColor:      inherit,
	})
	assertBackground(t, "top", toValidated(pr.Properties{
		pr.PBackgroundPosition: pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top right", toValidated(pr.Properties{
		pr.PBackgroundPosition: pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 100, Unit: pr.Perc}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top right 20px", toValidated(pr.Properties{
		pr.PBackgroundPosition: pr.Centers{{OriginX: "right", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 20, Unit: pr.Px}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top 1% right 20px", toValidated(pr.Properties{
		pr.PBackgroundPosition: pr.Centers{{OriginX: "right", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 20, Unit: pr.Px}, pr.Dimension{Value: 1, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top no-repeat", toValidated(pr.Properties{
		pr.PBackgroundRepeat:   pr.Repeats{{"no-repeat", "no-repeat"}},
		pr.PBackgroundPosition: pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top right no-repeat", toValidated(pr.Properties{
		pr.PBackgroundRepeat:   pr.Repeats{{"no-repeat", "no-repeat"}},
		pr.PBackgroundPosition: pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 100, Unit: pr.Perc}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top right 20px no-repeat", toValidated(pr.Properties{
		pr.PBackgroundRepeat:   pr.Repeats{{"no-repeat", "no-repeat"}},
		pr.PBackgroundPosition: pr.Centers{{OriginX: "right", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 20, Unit: pr.Px}, pr.Dimension{Value: 0, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "top 1% right 20px no-repeat", toValidated(pr.Properties{
		pr.PBackgroundRepeat:   pr.Repeats{{"no-repeat", "no-repeat"}},
		pr.PBackgroundPosition: pr.Centers{{OriginX: "right", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 20, Unit: pr.Px}, pr.Dimension{Value: 1, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "url(bar) #f00 repeat-y center left fixed", toValidated(pr.Properties{
		pr.PBackgroundColor:      pr.NewColor(1, 0, 0, 1),
		pr.PBackgroundImage:      pr.Images{pr.UrlImage("https://weasyprint.org/foo/bar")},
		pr.PBackgroundRepeat:     pr.Repeats{{"no-repeat", "repeat"}},
		pr.PBackgroundAttachment: pr.Strings{"fixed"},
		pr.PBackgroundPosition:   pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 0, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "#00f 10% 200px", toValidated(pr.Properties{
		pr.PBackgroundColor:    pr.NewColor(0, 0, 1, 1),
		pr.PBackgroundPosition: pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 10, Unit: pr.Perc}, pr.Dimension{Value: 200, Unit: pr.Px}}}},
	}))
	assertBackground(t, "right 78px fixed", toValidated(pr.Properties{
		pr.PBackgroundAttachment: pr.Strings{"fixed"},
		pr.PBackgroundPosition:   pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 100, Unit: pr.Perc}, pr.Dimension{Value: 78, Unit: pr.Px}}}},
	}))
	assertBackground(t, "center / cover red", toValidated(pr.Properties{
		pr.PBackgroundSize:     pr.Sizes{{String: "cover"}},
		pr.PBackgroundPosition: pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}}},
		pr.PBackgroundColor:    pr.NewColor(1, 0, 0, 1),
	}))
	assertBackground(t, "center / auto red", toValidated(pr.Properties{
		pr.PBackgroundSize:     pr.Sizes{{Width: pr.SToV("auto"), Height: pr.SToV("auto")}},
		pr.PBackgroundPosition: pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}}},
		pr.PBackgroundColor:    pr.NewColor(1, 0, 0, 1),
	}))
	assertBackground(t, "center / 42px", toValidated(pr.Properties{
		pr.PBackgroundSize:     pr.Sizes{{Width: pr.Dimension{Value: 42, Unit: pr.Px}.ToValue(), Height: pr.SToV("auto")}},
		pr.PBackgroundPosition: pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "center / 7% 4em", toValidated(pr.Properties{
		pr.PBackgroundSize:     pr.Sizes{{Width: pr.Dimension{Value: 7, Unit: pr.Perc}.ToValue(), Height: pr.Dimension{Value: 4, Unit: pr.Em}.ToValue()}},
		pr.PBackgroundPosition: pr.Centers{{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}}},
	}))
	assertBackground(t, "red content-box", toValidated(pr.Properties{
		pr.PBackgroundColor:  pr.NewColor(1, 0, 0, 1),
		pr.PBackgroundOrigin: pr.Strings{"content-box"},
		pr.PBackgroundClip:   pr.Strings{"content-box"},
	}))
	assertBackground(t, "red border-box content-box", toValidated(pr.Properties{
		pr.PBackgroundColor:  pr.NewColor(1, 0, 0, 1),
		pr.PBackgroundOrigin: pr.Strings{"border-box"},
		pr.PBackgroundClip:   pr.Strings{"content-box"},
	}))
	assertBackground(t, "url(bar) center, no-repeat", toValidated(pr.Properties{
		pr.PBackgroundColor: pr.NewColor(0, 0, 0, 0),
		pr.PBackgroundImage: pr.Images{pr.UrlImage("https://weasyprint.org/foo/bar"), pr.NoneImage{}},
		pr.PBackgroundPosition: pr.Centers{
			{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 50, Unit: pr.Perc}, pr.Dimension{Value: 50, Unit: pr.Perc}}},
			{OriginX: "left", OriginY: "top", Pos: pr.Point{pr.Dimension{Value: 0, Unit: pr.Perc}, pr.Dimension{Value: 0, Unit: pr.Perc}}},
		},
		pr.PBackgroundRepeat: pr.Repeats{{"repeat", "repeat"}, {"no-repeat", "no-repeat"}},
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "background: 10px lipsum", "invalid")
	assertInvalid(t, "background-position: 10px lipsum", "invalid")
	assertInvalid(t, "background: content-box red content-box", "invalid")
	assertInvalid(t, "background-image: inexistent-gradient(blue, green)", "invalid")
	// Color must be in the last layer :
	assertInvalid(t, "background: red, url(foo)", "invalid")
}
