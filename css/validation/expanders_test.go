package validation

import (
	"fmt"
	"reflect"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func TestEmptyExpanderValue(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for exp := range expanders {
		if exp == 0 {
			continue
		}
		assertInvalid(t, fmt.Sprintf("%s:", pr.Shortand(exp)), "Ignored")
	}
}

func TestTextDecoration(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	assertValidDict(t, "text-decoration: none", toValidated(pr.Properties{
		pr.PTextDecorationLine: pr.Decorations(0),
	}))
	assertValidDict(t, "text-decoration: overline", toValidated(pr.Properties{
		pr.PTextDecorationLine: pr.Overline,
	}))
	assertValidDict(t, "text-decoration: overline solid", toValidated(pr.Properties{
		pr.PTextDecorationLine:  pr.Overline,
		pr.PTextDecorationStyle: pr.String("solid"),
	}))
	assertValidDict(t, "text-decoration: overline blink line-through", toValidated(pr.Properties{
		pr.PTextDecorationLine: pr.Blink | pr.LineThrough | pr.Overline,
	}))
	assertValidDict(t, "text-decoration: red", toValidated(pr.Properties{
		pr.PTextDecorationColor: pr.NewColor(1, 0, 0, 1),
	}))
	assertValidDict(t, "text-decoration: blue 1px", map[pr.KnownProp]pr.DeclaredValue{
		pr.PTextDecorationColor:     pr.NewColor(0, 0, 1, 1),
		pr.PTextDecorationThickness: pr.FToPx(1),
	})
	assertValidDict(t, "text-decoration: 100% none", map[pr.KnownProp]pr.DeclaredValue{
		pr.PTextDecorationLine:      pr.Decorations(0),
		pr.PTextDecorationThickness: pr.Dimension{Value: 100, Unit: pr.Perc}.ToValue(),
	})
	assertValidDict(t, "text-decoration: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PTextDecorationColor:     pr.Inherit,
		pr.PTextDecorationLine:      pr.Inherit,
		pr.PTextDecorationStyle:     pr.Inherit,
		pr.PTextDecorationThickness: pr.Inherit,
	})

	assertInvalid(t, "text-decoration: solid solid", "invalid")
	assertInvalid(t, "text-decoration: red red", "invalid")
	assertInvalid(t, "text-decoration: underline none", "invalid")
	assertInvalid(t, "text-decoration: 1px 100%", "invalid")
	assertInvalid(t, "text-decoration: none none", "invalid")
}

// Test the 4-value pr.
func TestFourSides(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "margin: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PMarginTop:    pr.Inherit,
		pr.PMarginRight:  pr.Inherit,
		pr.PMarginBottom: pr.Inherit,
		pr.PMarginLeft:   pr.Inherit,
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
func TestBorders(t *testing.T) {
	capt := tu.CaptureLogs()
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

// Test the “list_style“ property.
func TestListStyle(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "list-style: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PListStylePosition: pr.Inherit,
		pr.PListStyleImage:    pr.Inherit,
		pr.PListStyleType:     pr.Inherit,
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
		"got multiple list-style-type values in a list-style shorthand")
}

// Test the “background“ property.
func TestBackground(t *testing.T) {
	capt := tu.CaptureLogs()
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
	assertBackground(t, "inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PBackgroundRepeat:     pr.Inherit,
		pr.PBackgroundAttachment: pr.Inherit,
		pr.PBackgroundImage:      pr.Inherit,
		pr.PBackgroundPosition:   pr.Inherit,
		pr.PBackgroundSize:       pr.Inherit,
		pr.PBackgroundClip:       pr.Inherit,
		pr.PBackgroundOrigin:     pr.Inherit,
		pr.PBackgroundColor:      pr.Inherit,
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
	assertInvalid(t, "background: content-box red content-box", "invalid")
	// Color must be in the last layer :
	assertInvalid(t, "background: red, url(foo)", "invalid")
	assertInvalid(t, "background-image: inexistent-gradient(blue, green)", "invalid")
	assertInvalid(t, "background-position: 10px lipsum", "invalid")
}

func TestBorderRadius(t *testing.T) {
	capt := tu.CaptureLogs()
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
	assertValidDict(t, "border-radius: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PBorderTopLeftRadius:     pr.Inherit,
		pr.PBorderTopRightRadius:    pr.Inherit,
		pr.PBorderBottomRightRadius: pr.Inherit,
		pr.PBorderBottomLeftRadius:  pr.Inherit,
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

func TestBorderImage(t *testing.T) {
	capt := tu.CaptureLogs()

	for _, test := range []struct {
		css    string
		result map[pr.KnownProp]pr.DeclaredValue
	}{
		{"url(border.png) 27", map[pr.KnownProp]pr.DeclaredValue{
			pr.PBorderImageSource: pr.UrlImage("https://weasyprint.org/foo/border.png"),
			pr.PBorderImageSlice:  pr.Values{pr.FToV(27)},
		}},
		{"url(border.png) 10 / 4 / 2 round stretch", map[pr.KnownProp]pr.DeclaredValue{
			pr.PBorderImageSource: pr.UrlImage("https://weasyprint.org/foo/border.png"),
			pr.PBorderImageSlice:  pr.Values{pr.FToV(10)},
			pr.PBorderImageWidth:  pr.Values{pr.FToV(4)},
			pr.PBorderImageOutset: pr.Values{pr.FToV(2)},
			pr.PBorderImageRepeat: pr.Strings{"round", "stretch"},
		}},
		{"10 // 2", map[pr.KnownProp]pr.DeclaredValue{
			pr.PBorderImageSlice:  pr.Values{pr.FToV(10)},
			pr.PBorderImageOutset: pr.Values{pr.FToV(2)},
		}},
		{"5.5%", map[pr.KnownProp]pr.DeclaredValue{
			pr.PBorderImageSlice: pr.Values{pr.PercToV(5.5)},
		}},
		{`stretch 2 url("border.png")`, map[pr.KnownProp]pr.DeclaredValue{
			pr.PBorderImageSource: pr.UrlImage("https://weasyprint.org/foo/border.png"),
			pr.PBorderImageSlice:  pr.Values{pr.FToV(2)},
			pr.PBorderImageRepeat: pr.Strings{"stretch"},
		}},
		{"1/2 round", map[pr.KnownProp]pr.DeclaredValue{
			pr.PBorderImageSlice:  pr.Values{pr.FToV(1)},
			pr.PBorderImageWidth:  pr.Values{pr.FToV(2)},
			pr.PBorderImageRepeat: pr.Strings{"round"},
		}},
		{"none", map[pr.KnownProp]pr.DeclaredValue{
			pr.PBorderImageSource: pr.NoneImage{},
		}},
	} {
		assertValidDict(t, fmt.Sprintf("border-image: %s", test.css), test.result)
	}
	capt.AssertNoLogs(t)

	for _, test := range []struct {
		css    string
		reason string
	}{
		{"url(border.png) url(border.png)", "multiple border-image-source"},
		{"10 10 10 10 10", "multiple border-image-slice"},
		{"1 / 2 / 3 / 4", "invalid"},
		{"/1", "invalid"},
		{"/1", "invalid"},
		{"round round round", "invalid"},
		{"-1", "invalid"},
		{"1 repeat 2", "multiple border-image-slice"},
		{"1% // 1%", "invalid"},
		{"1 / repeat", "invalid"},
		{"", "no value"},
	} {
		assertInvalid(t, fmt.Sprintf("border-image: %s", test.css), test.reason)
	}
}

// Test the “font“ property.
func TestFont(t *testing.T) {
	capt := tu.CaptureLogs()
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
	capt := tu.CaptureLogs()
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

func TestWordWrap(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "word-wrap: normal", toValidated(pr.Properties{
		pr.POverflowWrap: pr.String("normal"),
	}))
	assertValidDict(t, "word-wrap: break-word", toValidated(pr.Properties{
		pr.POverflowWrap: pr.String("break-word"),
	}))
	assertValidDict(t, "word-wrap: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.POverflowWrap: pr.Inherit,
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "word-wrap: none", "invalid")
	assertInvalid(t, "word-wrap: normal, break-word", "invalid")
}

func TestFlex(t *testing.T) {
	capt := tu.CaptureLogs()

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
	assertValidDict(t, "flex: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PFlexGrow:   pr.Inherit,
		pr.PFlexShrink: pr.Inherit,
		pr.PFlexBasis:  pr.Inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, "flex: auto 0 0 0", "invalid")
	assertInvalid(t, "flex: 1px 2px", "invalid")
	assertInvalid(t, "flex: auto auto", "invalid")
	assertInvalid(t, "flex: auto 1 auto", "invalid")
}

func TestFlexFlow(t *testing.T) {
	capt := tu.CaptureLogs()

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
	assertValidDict(t, "flex-flow: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PFlexDirection: pr.Inherit,
		pr.PFlexWrap:      pr.Inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, "flex-flow: 1px", "invalid")
	assertInvalid(t, "flex-flow: wrap 1px", "invalid")
	assertInvalid(t, "flex-flow: row row", "invalid")
	assertInvalid(t, "flex-flow: wrap nowrap", "invalid")
	assertInvalid(t, "flex-flow: column wrap nowrap row", "invalid")
}

func TestGridColumnRow(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, test := range []struct {
		css        string
		start, end pr.GridLine
	}{
		{"auto", pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}},
		{"auto / auto", pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}},
		{"4", pr.GridLine{Val: 4}, pr.GridLine{Tag: pr.Auto}},
		{"c", pr.GridLine{Ident: "c"}, pr.GridLine{Ident: "c"}},
		{"4 / -4", pr.GridLine{Val: 4}, pr.GridLine{Val: -4}},
		{"c / d", pr.GridLine{Ident: "c"}, pr.GridLine{Ident: "d"}},
		{"ab / cd 4", pr.GridLine{Ident: "ab"}, pr.GridLine{Val: 4, Ident: "cd"}},
		{"ab 2 span", pr.GridLine{Tag: pr.Span, Val: 2, Ident: "ab"}, pr.GridLine{Tag: pr.Auto}},
	} {
		assertValidDict(t, fmt.Sprintf("grid-column: %s", test.css), map[pr.KnownProp]pr.DeclaredValue{
			pr.PGridColumnStart: test.start, pr.PGridColumnEnd: test.end,
		})
		assertValidDict(t, fmt.Sprintf("grid-row: %s", test.css), map[pr.KnownProp]pr.DeclaredValue{
			pr.PGridRowStart: test.start, pr.PGridRowEnd: test.end,
		})
	}

	for _, css := range [...]string{
		"auto auto",
		"4 / 2 / c",
		"span",
		"4 / span",
		"c /",
		"/4",
		"col / 2.1",
	} {
		assertInvalid(t, fmt.Sprintf("grid-column: %s", css), "invalid")
		assertInvalid(t, fmt.Sprintf("grid-row: %s", css), "invalid")
	}
}

func TestGridArea(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, test := range []struct {
		css                                      string
		rowStart, rowEnd, columnStart, columnEnd pr.GridLine
	}{
		{"auto", pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}},
		{"auto / auto", pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}},
		{"auto / auto / auto", pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}},
		{"auto / auto / auto / auto", pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}},
		{"1/c/2 d/span 2 ab", pr.GridLine{Val: 1}, pr.GridLine{Ident: "c"}, pr.GridLine{Val: 2, Ident: "d"}, pr.GridLine{Tag: pr.Span, Val: 2, Ident: "ab"}},
		{"1  /  c", pr.GridLine{Val: 1}, pr.GridLine{Ident: "c"}, pr.GridLine{Tag: pr.Auto}, pr.GridLine{Ident: "c"}},
		{"a / c 2", pr.GridLine{Ident: "a"}, pr.GridLine{Val: 2, Ident: "c"}, pr.GridLine{Ident: "a"}, pr.GridLine{Tag: pr.Auto}},
		{"a", pr.GridLine{Ident: "a"}, pr.GridLine{Ident: "a"}, pr.GridLine{Ident: "a"}, pr.GridLine{Ident: "a"}},
		{"span 2", pr.GridLine{Tag: pr.Span, Val: 2}, pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}, pr.GridLine{Tag: pr.Auto}},
	} {
		assertValidDict(t, fmt.Sprintf("grid-area: %s", test.css), map[pr.KnownProp]pr.DeclaredValue{
			pr.PGridRowStart:    test.rowStart,
			pr.PGridRowEnd:      test.rowEnd,
			pr.PGridColumnStart: test.columnStart,
			pr.PGridColumnEnd:   test.columnEnd,
		})
	}

	for _, css := range [...]string{
		"auto auto",
		"auto / auto auto",
		"4 / 2 / c / d / e",
		"span",
		"4 / span",
		"c /",
		"/4",
		"c//4",
		"/",
		"1 / 2 / 4 / 0.5",
	} {
		assertInvalid(t, fmt.Sprintf("grid-area: %s", css), "invalid")
	}
}

func TestPageBreak(t *testing.T) {
	capt := tu.CaptureLogs()

	assertValidDict(t, "page-break-after: left", toValidated(pr.Properties{
		pr.PBreakAfter: pr.String("left"),
	}))
	assertValidDict(t, "page-break-before: always", toValidated(pr.Properties{
		pr.PBreakBefore: pr.String("page"),
	}))
	assertValidDict(t, "page-break-after: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PBreakAfter: pr.Inherit,
	})
	assertValidDict(t, "page-break-before: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PBreakBefore: pr.Inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, "page-break-after: top", "invalid")
	assertInvalid(t, "page-break-before: 1px", "invalid")
}

func TestPageBreakInside(t *testing.T) {
	capt := tu.CaptureLogs()

	assertValidDict(t, "page-break-inside: avoid", toValidated(pr.Properties{
		pr.PBreakInside: pr.String("avoid"),
	}))
	assertValidDict(t, "page-break-inside: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PBreakInside: pr.Inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, "page-break-inside: top", "invalid")
}

func TestColumns(t *testing.T) {
	capt := tu.CaptureLogs()
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
	capt := tu.CaptureLogs()

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
	assertValidDict(t, "line-clamp: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PMaxLines:      pr.Inherit,
		pr.PContinue:      pr.Inherit,
		pr.PBlockEllipsis: pr.Inherit,
	})

	capt.AssertNoLogs(t)

	assertInvalid(t, `line-clamp: none none none`, "invalid")
	assertInvalid(t, `line-clamp: 1px`, "invalid")
	assertInvalid(t, `line-clamp: 0 "…"`, "invalid")
	assertInvalid(t, `line-clamp: 1px 2px`, "invalid")
}

func TestTextAlign(t *testing.T) {
	capt := tu.CaptureLogs()

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
	assertValidDict(t, "text-align: inherit", map[pr.KnownProp]pr.DeclaredValue{
		pr.PTextAlignAll:  pr.Inherit,
		pr.PTextAlignLast: pr.Inherit,
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
func assertBackground(t *testing.T, css string, expected map[pr.KnownProp]pr.DeclaredValue) {
	expanded := expandToDict(t, "background: "+css, "")
	col, in := expected[pr.PBackgroundColor]
	if !in {
		col = pr.InitialValues[pr.PBackgroundColor]
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

	nbLayers := len(bi.(pr.Images))
	for name, value := range expanded {
		initv := pr.InitialValues[pr.KnownProp(name)].(repeatable)
		ref := pr.DeclaredValue(initv.Repeat(nbLayers))
		if !reflect.DeepEqual(value, ref) {
			t.Fatalf("expected %v got %v", ref, value)
		}
	}
}
