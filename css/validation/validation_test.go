package validation

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func toValidated(d pr.Properties) map[pr.KnownProp]pr.ValidatedProperty {
	out := make(map[pr.KnownProp]pr.ValidatedProperty)
	for k, v := range d {
		out[k] = pr.AsCascaded(v).AsValidated()
	}
	return out
}

// Helper to test shorthand properties expander functions.
func expandToDict(t *testing.T, css string, expectedError string) map[pr.KnownProp]pr.ValidatedProperty {
	t.Helper()

	declarations := parser.ParseDeclarationListString(css, false, false)

	capt := tu.CaptureLogs()
	baseUrl := "https://weasyprint.org/foo/"
	validated := PreprocessDeclarations(baseUrl, declarations)
	logs := capt.Logs()

	if expectedError != "" {
		if len(logs) != 1 || !strings.Contains(logs[0], expectedError) {
			t.Log(validated)

			t.Fatalf("for %s expected error \n%s\n got\n%v (len : %d)", css, expectedError, logs, len(logs))
		}
	} else {
		capt.AssertNoLogs(t)
	}
	out := map[pr.KnownProp]pr.ValidatedProperty{}
	for _, v := range validated {
		if v.Value.SpecialProperty != nil || v.Value.ToCascaded().Default != pr.Initial {
			out[v.Name.KnownProp] = v.Value
		}
	}
	return out
}

// message="invalid"
func assertInvalid(t *testing.T, css, message string) {
	t.Helper()

	d := expandToDict(t, css, message)
	if len(d) != 0 {
		t.Fatalf("expected no properties, got %v", d)
	}
}

func assertValidDict(t *testing.T, css string, ref map[pr.KnownProp]pr.ValidatedProperty) {
	t.Helper()

	got := expandToDict(t, css, "")
	if !reflect.DeepEqual(ref, got) {
		t.Fatalf("expected %v got %v", ref, got)
	}
}

func TestNotPrint(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertInvalid(t, "volume: 42", "the property does not apply for the print media")
}

func TestUnstablePrefix(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	d := expandToDict(t, "-weasy-max-lines: 3",
		"prefixes on unstable attributes are deprecated")

	tu.AssertEqual(t, d, toValidated(pr.Properties{pr.PMaxLines: pr.TaggedInt{I: 3}}), "unstable prefix")
}

func TestNormalPrefix(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertInvalid(t, "-weasy-display: block", "prefix on this attribute is not supported")
}

func TestUnknownPrefix(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertInvalid(t, "-unknown-display: block", "prefixed selectors are ignored")
}

func TestFunction(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "clip: rect(1px, 3em, auto, auto)", toValidated(pr.Properties{
		pr.PClip: pr.Values{
			pr.Dimension{Value: 1, Unit: pr.Px}.ToValue(),
			pr.Dimension{Value: 3, Unit: pr.Em}.ToValue(),
			pr.SToV("auto"),
			pr.SToV("auto"),
		},
	}))
	assertValidDict(t, "clip: rect(1px, 3em, auto auto)", toValidated(pr.Properties{
		pr.PClip: pr.Values{
			pr.Dimension{Value: 1, Unit: pr.Px}.ToValue(),
			pr.Dimension{Value: 3, Unit: pr.Em}.ToValue(),
			pr.SToV("auto"),
			pr.SToV("auto"),
		},
	}))
	assertValidDict(t, "clip: rect(1px 3em auto 1px)", toValidated(pr.Properties{
		pr.PClip: pr.Values{
			pr.Dimension{Value: 1, Unit: pr.Px}.ToValue(),
			pr.Dimension{Value: 3, Unit: pr.Em}.ToValue(),
			pr.SToV("auto"),
			pr.Dimension{Value: 1, Unit: pr.Px}.ToValue(),
		},
	}))
	assertInvalid(t, "clip: square(1px, 3em, auto, auto)", "invalid")
	assertInvalid(t, "clip: rect(1px, 3em, auto)", "invalid")
	assertInvalid(t, "clip: rect(1px, 3em / auto)", "invalid")
	capt.AssertNoLogs(t)
}

func TestCounters(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "counter-reset: foo bar 2 baz", toValidated(pr.Properties{
		pr.PCounterReset: pr.SIntStrings{Values: pr.IntStrings{{String: "foo", Int: 0}, {String: "bar", Int: 2}, {String: "baz", Int: 0}}},
	}))
	assertValidDict(t, "counter-increment: foo bar 2 baz", toValidated(pr.Properties{
		pr.PCounterIncrement: pr.SIntStrings{Values: pr.IntStrings{{String: "foo", Int: 1}, {String: "bar", Int: 2}, {String: "baz", Int: 1}}},
	}))
	assertValidDict(t, "counter-reset: foo", toValidated(pr.Properties{
		pr.PCounterReset: pr.SIntStrings{Values: pr.IntStrings{{String: "foo", Int: 0}}},
	}))
	assertValidDict(t, "counter-reset: FoO", toValidated(pr.Properties{
		pr.PCounterReset: pr.SIntStrings{Values: pr.IntStrings{{String: "FoO", Int: 0}}},
	}))
	assertValidDict(t, "counter-increment: foo bAr 2 Bar", toValidated(pr.Properties{
		pr.PCounterIncrement: pr.SIntStrings{Values: pr.IntStrings{{String: "foo", Int: 1}, {String: "bAr", Int: 2}, {String: "Bar", Int: 1}}},
	}))
	assertValidDict(t, "counter-reset: none", toValidated(pr.Properties{
		pr.PCounterReset: pr.SIntStrings{Values: pr.IntStrings{}},
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "counter-reset: foo initial", "invalid counter name: initial.")
	assertInvalid(t, "counter-reset: foo none", "invalid counter name: none.")
	assertInvalid(t, "counter-reset: foo 3px", "invalid")
	assertInvalid(t, "counter-reset: 3", "invalid")
}

func TestSpacing(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "letter-spacing: normal", toValidated(pr.Properties{
		pr.PLetterSpacing: pr.SToV("normal"),
	}))
	assertValidDict(t, "letter-spacing: 3px", toValidated(pr.Properties{
		pr.PLetterSpacing: pr.Dimension{Value: 3, Unit: pr.Px}.ToValue(),
	}))
	assertValidDict(t, "word-spacing: normal", toValidated(pr.Properties{
		pr.PWordSpacing: pr.SToV("normal"),
	}))
	assertValidDict(t, "word-spacing: 3px", toValidated(pr.Properties{
		pr.PWordSpacing: pr.Dimension{Value: 3, Unit: pr.Px}.ToValue(),
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "letter_spacing: normal", "did you mean letter-spacing")
	assertInvalid(t, "letter-spacing: 3", "invalid")
	assertInvalid(t, "word-spacing: 3", "invalid")
}

func TestFootnote(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "footnote-policy: auto", toValidated(pr.Properties{
		pr.PFootnotePolicy: pr.String("auto"),
	}))
	assertValidDict(t, "footnote-policy: line", toValidated(pr.Properties{
		pr.PFootnotePolicy: pr.String("line"),
	}))
	assertValidDict(t, "footnote-policy: block", toValidated(pr.Properties{
		pr.PFootnotePolicy: pr.String("block"),
	}))

	assertValidDict(t, "footnote-display: block", toValidated(pr.Properties{
		pr.PFootnoteDisplay: pr.String("block"),
	}))
	assertValidDict(t, "footnote-display: inline", toValidated(pr.Properties{
		pr.PFootnoteDisplay: pr.String("inline"),
	}))
	assertValidDict(t, "footnote-display: compact", toValidated(pr.Properties{
		pr.PFootnoteDisplay: pr.String("compact"),
	}))
	capt.AssertNoLogs(t)

	assertInvalid(t, "footnote_display: block", "did you mean footnote-display")
	assertInvalid(t, "footnote-display: 3", "invalid")
	assertInvalid(t, "footnote-policy: 3", "invalid")
	assertInvalid(t, "footnote-policy: normal", "invalid")
}

func TestDecoration(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "text-decoration-line: none", toValidated(pr.Properties{
		pr.PTextDecorationLine: pr.Decorations{},
	}))
	assertValidDict(t, "text-decoration-line: overline", toValidated(pr.Properties{
		pr.PTextDecorationLine: pr.Decorations(utils.NewSet("overline")),
	}))
	// blink is accepted but ignored
	assertValidDict(t, "text-decoration-line: overline blink line-through", toValidated(pr.Properties{
		pr.PTextDecorationLine: pr.Decorations(utils.NewSet("blink", "line-through", "overline")),
	}))

	assertValidDict(t, "text-decoration-style: solid", toValidated(pr.Properties{
		pr.PTextDecorationStyle: pr.String("solid"),
	}))
	assertValidDict(t, "text-decoration-style: double", toValidated(pr.Properties{
		pr.PTextDecorationStyle: pr.String("double"),
	}))
	assertValidDict(t, "text-decoration-style: dotted", toValidated(pr.Properties{
		pr.PTextDecorationStyle: pr.String("dotted"),
	}))
	assertValidDict(t, "text-decoration-style: dashed", toValidated(pr.Properties{
		pr.PTextDecorationStyle: pr.String("dashed"),
	}))

	capt.AssertNoLogs(t)
}

func TestSize(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "size: 200px", toValidated(pr.Properties{
		pr.PSize: pr.Point{{Value: 200, Unit: pr.Px}, {Value: 200, Unit: pr.Px}},
	}))
	assertValidDict(t, "size: 200px 300pt", toValidated(pr.Properties{
		pr.PSize: pr.Point{{Value: 200, Unit: pr.Px}, {Value: 300, Unit: pr.Pt}},
	}))
	assertValidDict(t, "size: auto", toValidated(pr.Properties{
		pr.PSize: pr.Point{{Value: 210, Unit: pr.Mm}, {Value: 297, Unit: pr.Mm}},
	}))
	assertValidDict(t, "size: portrait", toValidated(pr.Properties{
		pr.PSize: pr.Point{{Value: 210, Unit: pr.Mm}, {Value: 297, Unit: pr.Mm}},
	}))
	assertValidDict(t, "size: landscape", toValidated(pr.Properties{
		pr.PSize: pr.Point{{Value: 297, Unit: pr.Mm}, {Value: 210, Unit: pr.Mm}},
	}))
	assertValidDict(t, "size: A3 portrait", toValidated(pr.Properties{
		pr.PSize: pr.Point{{Value: 297, Unit: pr.Mm}, {Value: 420, Unit: pr.Mm}},
	}))
	assertValidDict(t, "size: A3 landscape", toValidated(pr.Properties{
		pr.PSize: pr.Point{{Value: 420, Unit: pr.Mm}, {Value: 297, Unit: pr.Mm}},
	}))
	assertValidDict(t, "size: portrait A3", toValidated(pr.Properties{
		pr.PSize: pr.Point{{Value: 297, Unit: pr.Mm}, {Value: 420, Unit: pr.Mm}},
	}))
	assertValidDict(t, "size: landscape A3", toValidated(pr.Properties{
		pr.PSize: pr.Point{{Value: 420, Unit: pr.Mm}, {Value: 297, Unit: pr.Mm}},
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "size: A3 landscape A3", "invalid")
	assertInvalid(t, "size: A12", "invalid")
	assertInvalid(t, "size: foo", "invalid")
	assertInvalid(t, "size: foo bar", "invalid")
	assertInvalid(t, "size: 20%", "invalid")
}

func TestTransforms(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "transform: none", toValidated(pr.Properties{
		pr.PTransform: pr.Transforms{},
	}))
	assertValidDict(t, "transform: translate(6px) rotate(90deg)", toValidated(pr.Properties{
		pr.PTransform: pr.Transforms{
			{String: "translate", Dimensions: []pr.Dimension{{Value: 6, Unit: pr.Px}, {Value: 0, Unit: pr.Px}}},
			{String: "rotate", Dimensions: []pr.Dimension{pr.FToD(math.Pi / 2)}},
		},
	}))
	assertValidDict(t, "transform: translate(-4px, 0)", toValidated(pr.Properties{
		pr.PTransform: pr.Transforms{{String: "translate", Dimensions: []pr.Dimension{{Value: -4, Unit: pr.Px}, {Value: 0, Unit: pr.Scalar}}}},
	}))
	assertValidDict(t, "transform: translate(6px, 20%)", toValidated(pr.Properties{
		pr.PTransform: pr.Transforms{{String: "translate", Dimensions: []pr.Dimension{{Value: 6, Unit: pr.Px}, {Value: 20, Unit: pr.Perc}}}},
	}))
	assertValidDict(t, "transform: translate(6px 20%)", toValidated(pr.Properties{
		pr.PTransform: pr.Transforms{{String: "translate", Dimensions: []pr.Dimension{{Value: 6, Unit: pr.Px}, {Value: 20, Unit: pr.Perc}}}},
	}))
	assertValidDict(t, "transform: scale(2)", toValidated(pr.Properties{
		pr.PTransform: pr.Transforms{{String: "scale", Dimensions: []pr.Dimension{pr.FToD(2), pr.FToD(2)}}},
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "transform: lipsumize(6px)", "invalid")
	assertInvalid(t, "transform: foo", "invalid")
	assertInvalid(t, "transform: scale(2) foo", "invalid")
	assertInvalid(t, "transform: 6px", "invalid")
}

type repeatable interface {
	Repeat(int) pr.CssProperty
}

func checkPosition(t *testing.T, css string, expected pr.Center) {
	l := expandToDict(t, "background-position:"+css, "")
	var (
		name pr.KnownProp
		v    pr.ValidatedProperty
	)
	for name_, v_ := range l {
		name = name_
		v = v_
	}
	if name != pr.PBackgroundPosition {
		t.Fatalf("expected background_position got %s", name)
	}
	exp := pr.AsCascaded(pr.Centers{expected}).AsValidated()
	if !reflect.DeepEqual(v, exp) {
		t.Fatalf("expected %v got %v", exp, v)
	}
}

// Test the “background-position“ property.
func TestExpandBackgroundPosition(t *testing.T) {
	capt := tu.CaptureLogs()

	css_xs := [5]string{"left", "center", "right", "4.5%", "12px"}
	val_xs := [5]pr.Dimension{{Value: 0, Unit: pr.Perc}, {Value: 50, Unit: pr.Perc}, {Value: 100, Unit: pr.Perc}, {Value: 4.5, Unit: pr.Perc}, {Value: 12, Unit: pr.Px}}
	css_ys := [5]string{"top", "center", "bottom", "7%", "1.5px"}
	val_ys := [5]pr.Dimension{{Value: 0, Unit: pr.Perc}, {Value: 50, Unit: pr.Perc}, {Value: 100, Unit: pr.Perc}, {Value: 7, Unit: pr.Perc}, {Value: 1.5, Unit: pr.Px}}
	for i, css_x := range css_xs {
		val_x := val_xs[i]
		for j, css_y := range css_ys {
			val_y := val_ys[j]
			// Two tokens:
			checkPosition(t, fmt.Sprintf("%s %s", css_x, css_y), pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{val_x, val_y}})
		}
		// One token:
		checkPosition(t, css_x, pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{val_x, {Value: 50, Unit: pr.Perc}}})
	}
	// One token, vertical
	checkPosition(t, "top", pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{{Value: 50, Unit: pr.Perc}, {Value: 0, Unit: pr.Perc}}})
	checkPosition(t, "bottom", pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{{Value: 50, Unit: pr.Perc}, {Value: 100, Unit: pr.Perc}}})

	// Three tokens:
	checkPosition(t, "center top 10%", pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{{Value: 50, Unit: pr.Perc}, {Value: 10, Unit: pr.Perc}}})
	checkPosition(t, "top 10% center", pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{{Value: 50, Unit: pr.Perc}, {Value: 10, Unit: pr.Perc}}})
	checkPosition(t, "center bottom 10%", pr.Center{OriginX: "left", OriginY: "bottom", Pos: pr.Point{{Value: 50, Unit: pr.Perc}, {Value: 10, Unit: pr.Perc}}})
	checkPosition(t, "bottom 10% center", pr.Center{OriginX: "left", OriginY: "bottom", Pos: pr.Point{{Value: 50, Unit: pr.Perc}, {Value: 10, Unit: pr.Perc}}})

	checkPosition(t, "right top 10%", pr.Center{OriginX: "right", OriginY: "top", Pos: pr.Point{{Value: 0, Unit: pr.Perc}, {Value: 10, Unit: pr.Perc}}})
	checkPosition(t, "top 10% right", pr.Center{OriginX: "right", OriginY: "top", Pos: pr.Point{{Value: 0, Unit: pr.Perc}, {Value: 10, Unit: pr.Perc}}})
	checkPosition(t, "right bottom 10%", pr.Center{OriginX: "right", OriginY: "bottom", Pos: pr.Point{{Value: 0, Unit: pr.Perc}, {Value: 10, Unit: pr.Perc}}})
	checkPosition(t, "bottom 10% right", pr.Center{OriginX: "right", OriginY: "bottom", Pos: pr.Point{{Value: 0, Unit: pr.Perc}, {Value: 10, Unit: pr.Perc}}})

	checkPosition(t, "center left 10%", pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 50, Unit: pr.Perc}}})
	checkPosition(t, "left 10% center", pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 50, Unit: pr.Perc}}})
	checkPosition(t, "center right 10%", pr.Center{OriginX: "right", OriginY: "top", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 50, Unit: pr.Perc}}})
	checkPosition(t, "right 10% center", pr.Center{OriginX: "right", OriginY: "top", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 50, Unit: pr.Perc}}})

	checkPosition(t, "bottom left 10%", pr.Center{OriginX: "left", OriginY: "bottom", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 0, Unit: pr.Perc}}})
	checkPosition(t, "left 10% bottom", pr.Center{OriginX: "left", OriginY: "bottom", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 0, Unit: pr.Perc}}})
	checkPosition(t, "bottom right 10%", pr.Center{OriginX: "right", OriginY: "bottom", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 0, Unit: pr.Perc}}})
	checkPosition(t, "right 10% bottom", pr.Center{OriginX: "right", OriginY: "bottom", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 0, Unit: pr.Perc}}})

	// Four tokens :
	checkPosition(t, "left 10% bottom 3px", pr.Center{OriginX: "left", OriginY: "bottom", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 3, Unit: pr.Px}}})
	checkPosition(t, "bottom 3px left 10%", pr.Center{OriginX: "left", OriginY: "bottom", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 3, Unit: pr.Px}}})
	checkPosition(t, "right 10% top 3px", pr.Center{OriginX: "right", OriginY: "top", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 3, Unit: pr.Px}}})
	checkPosition(t, "top 3px right 10%", pr.Center{OriginX: "right", OriginY: "top", Pos: pr.Point{{Value: 10, Unit: pr.Perc}, {Value: 3, Unit: pr.Px}}})

	capt.AssertNoLogs(t)

	assertInvalid(t, "background-position: left center 3px", "invalid")
	assertInvalid(t, "background-position: 3px left", "invalid")
	assertInvalid(t, "background-position: bottom 4%", "invalid")
	assertInvalid(t, "background-position: bottom top", "invalid")
}

// Test the “line-height“ property.
func TestLineHeight(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "line-height: 1px", toValidated(pr.Properties{
		pr.PLineHeight: pr.Dimension{Value: 1, Unit: pr.Px}.ToValue(),
	}))
	assertValidDict(t, "line-height: 1.1%", toValidated(pr.Properties{
		pr.PLineHeight: pr.Dimension{Value: 1.1, Unit: pr.Perc}.ToValue(),
	}))
	assertValidDict(t, "line-height: 1em", toValidated(pr.Properties{
		pr.PLineHeight: pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
	}))
	assertValidDict(t, "line-height: 1", toValidated(pr.Properties{
		pr.PLineHeight: pr.Dimension{Value: 1, Unit: pr.Scalar}.ToValue(),
	}))
	assertValidDict(t, "line-height: 1.3", toValidated(pr.Properties{
		pr.PLineHeight: pr.Dimension{Value: 1.3, Unit: pr.Scalar}.ToValue(),
	}))
	assertValidDict(t, "line-height: -0", toValidated(pr.Properties{
		pr.PLineHeight: pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
	}))
	assertValidDict(t, "line-height: 0px", toValidated(pr.Properties{
		pr.PLineHeight: pr.Dimension{Value: 0, Unit: pr.Px}.ToValue(),
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "line-height: 1deg", "invalid")
	assertInvalid(t, "line-height: -1px", "invalid")
	assertInvalid(t, "line-height: -1", "invalid")
	assertInvalid(t, "line-height: -0.5%", "invalid")
	assertInvalid(t, "line-height: 1px 1px", "invalid")
}

func TestImageResolution(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "image-resolution: .5dppx", toValidated(pr.Properties{
		pr.PImageResolution: pr.FToV(.5),
	}))
	capt.AssertNoLogs(t)

	assertInvalid(t, "image-resolution: 1deg", "invalid")
	assertInvalid(t, "image-resolution: -0.5%", "invalid")
	assertInvalid(t, "image-resolution: 1px 1px", "invalid")
}

func TestObjectFit(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "object-fit: cover", toValidated(pr.Properties{
		pr.PObjectFit: pr.String("cover"),
	}))
	capt.AssertNoLogs(t)

	assertInvalid(t, "object-fit: 1deg", "invalid")
	assertInvalid(t, "object-fit: -0.5%", "invalid")
	assertInvalid(t, "object-fit: 1px 1px", "invalid")
}

func TestMinMaxWidthHeight(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "min-width: 30px", toValidated(pr.Properties{
		pr.PMinWidth: pr.FToPx(30),
	}))
	assertValidDict(t, "min-height: 20px", toValidated(pr.Properties{
		pr.PMinHeight: pr.FToPx(20),
	}))
	assertValidDict(t, "max-width: 30px", toValidated(pr.Properties{
		pr.PMaxWidth: pr.FToPx(30),
	}))
	assertValidDict(t, "max-height: 20px", toValidated(pr.Properties{
		pr.PMaxHeight: pr.FToPx(20),
	}))
	capt.AssertNoLogs(t)

	assertInvalid(t, "min-width: red", "invalid")
	assertInvalid(t, "min-width: 1px 1px", "invalid")
	assertInvalid(t, "min-height: red", "invalid")
	assertInvalid(t, "min-height: 1px 1px", "invalid")
	assertInvalid(t, "max-width: red", "invalid")
	assertInvalid(t, "max-width: 1px 1px", "invalid")
	assertInvalid(t, "max-height: red", "invalid")
	assertInvalid(t, "max-height: 1px 1px", "invalid")
}

// Test the “string-set“ property.
func TestStringSet(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "string-set: test content(text)", toValidated(pr.Properties{
		pr.PStringSet: pr.StringSet{Contents: []pr.SContent{
			{String: "test", Contents: []pr.ContentProperty{{Type: "content()", Content: pr.String("text")}}},
		}},
	}))
	assertValidDict(t, "string-set: test content(before)", toValidated(pr.Properties{
		pr.PStringSet: pr.StringSet{Contents: []pr.SContent{
			{String: "test", Contents: []pr.ContentProperty{{Type: "content()", Content: pr.String("before")}}},
		}},
	}))
	assertValidDict(t, `string-set: test "string"`, toValidated(pr.Properties{
		pr.PStringSet: pr.StringSet{Contents: []pr.SContent{
			{String: "test", Contents: []pr.ContentProperty{{Type: "string", Content: pr.String("string")}}},
		}},
	}))
	assertValidDict(t, `string-set: test1 "string", test2 "string"`, toValidated(pr.Properties{
		pr.PStringSet: pr.StringSet{Contents: []pr.SContent{
			{String: "test1", Contents: []pr.ContentProperty{{Type: "string", Content: pr.String("string")}}},
			{String: "test2", Contents: []pr.ContentProperty{{Type: "string", Content: pr.String("string")}}},
		}},
	}))
	assertValidDict(t, "string-set: test attr(class)", toValidated(pr.Properties{
		pr.PStringSet: pr.StringSet{Contents: []pr.SContent{
			{String: "test", Contents: []pr.ContentProperty{{Type: "attr()", Content: pr.AttrData{Name: "class", TypeOrUnit: "string"}}}},
		}},
	}))
	assertValidDict(t, "string-set: test counter(count)", toValidated(pr.Properties{
		pr.PStringSet: pr.StringSet{Contents: []pr.SContent{
			{String: "test", Contents: []pr.ContentProperty{{Type: "counter()", Content: pr.Counters{Name: "count", Style: pr.CounterStyleID{Name: "decimal"}}}}},
		}},
	}))
	assertValidDict(t, "string-set: test counter(count, upper-roman)", toValidated(pr.Properties{
		pr.PStringSet: pr.StringSet{Contents: []pr.SContent{
			{String: "test", Contents: []pr.ContentProperty{{Type: "counter()", Content: pr.Counters{Name: "count", Style: pr.CounterStyleID{Name: "upper-roman"}}}}},
		}},
	}))
	assertValidDict(t, `string-set: test counters(count, ".")`, toValidated(pr.Properties{
		pr.PStringSet: pr.StringSet{Contents: []pr.SContent{
			{String: "test", Contents: []pr.ContentProperty{{Type: "counters()", Content: pr.Counters{Name: "count", Separator: ".", Style: pr.CounterStyleID{Name: "decimal"}}}}},
		}},
	}))
	assertValidDict(t, `string-set: test counters(count, ".", upper-roman)`, toValidated(pr.Properties{
		pr.PStringSet: pr.StringSet{Contents: []pr.SContent{
			{String: "test", Contents: []pr.ContentProperty{{Type: "counters()", Content: pr.Counters{Name: "count", Separator: ".", Style: pr.CounterStyleID{Name: "upper-roman"}}}}},
		}},
	}))
	assertValidDict(t, `string-set: test content(text) "string" attr(title) attr(title) counter(count)`, toValidated(pr.Properties{
		pr.PStringSet: pr.StringSet{Contents: []pr.SContent{
			{String: "test", Contents: []pr.ContentProperty{
				{Type: "content()", Content: pr.String("text")},
				{Type: "string", Content: pr.String("string")},
				{Type: "attr()", Content: pr.AttrData{Name: "title", TypeOrUnit: "string"}},
				{Type: "attr()", Content: pr.AttrData{Name: "title", TypeOrUnit: "string"}},
				{Type: "counter()", Content: pr.Counters{Name: "count", Style: pr.CounterStyleID{Name: "decimal"}}},
			}},
		}},
	}))

	capt.AssertNoLogs(t)
	assertInvalid(t, "string-set: test", "invalid")
	assertInvalid(t, "string-set: test test1", "invalid")
	assertInvalid(t, "string-set: test content(test)", "invalid")
	assertInvalid(t, "string-set: test unknown()", "invalid")
	assertInvalid(t, "string-set: test attr(id, class)", "invalid")
}

var (
	red           = pr.NewColor(1, 0, 0, 1)
	lime          = pr.NewColor(0, 1, 0, 1)
	blue          = pr.NewColor(0, 0, 1, 1)
	pi   utils.Fl = math.Pi
)

func checkGradientGeneric(t *testing.T, css string, expected pr.Image) {
	repeatings := [2]bool{false, true}
	prefixs := [2]string{"", "repeating-"}
	for i, repeating := range repeatings {
		prefix := prefixs[i]
		var mode string
		switch typed := expected.(type) {
		case pr.LinearGradient:
			typed.Repeating = repeating
			expected = typed
			mode = "linear"
		case pr.RadialGradient:
			typed.Repeating = repeating
			expected = typed
			mode = "radial"
		default:
			t.Fatalf("bad expected gradient !")
		}

		expanded := expandToDict(t, fmt.Sprintf("background-image: %s%s-gradient(%s)", prefix, mode, css), "")
		var image pr.Image
		for _, v := range expanded {
			image = v.ToCascaded().ToCSS().(pr.Images)[0]
		}
		if !reflect.DeepEqual(image, expected) {
			t.Fatalf("%s: expected %v got %v", css, expected, image)
		}
	}
}

func invalidGeneric(mode string, t *testing.T, css string) {
	assertInvalid(t, fmt.Sprintf("background-image: %s-gradient(%s)", mode, css), "invalid")
	assertInvalid(t, fmt.Sprintf("background-image: repeating-%s-gradient(%s)", mode, css), "invalid")
}

func TestLinearGradient(t *testing.T) {
	invalid := func(t *testing.T, css string) {
		invalidGeneric("linear", t, css)
	}

	gradient := func(t *testing.T, css string, direction pr.DirectionType, colors []pr.Color, stopPositions []pr.Dimension) {
		if colors == nil {
			colors = []pr.Color{blue}
		}
		if stopPositions == nil {
			stopPositions = []pr.Dimension{{}}
		}
		colorStops := make([]pr.ColorStop, len(colors))
		for i, s := range stopPositions {
			colorStops[i] = pr.ColorStop{Color: colors[i], Position: s}
		}
		checkGradientGeneric(t, css, pr.LinearGradient{ColorStops: colorStops, Direction: direction})
	}
	invalid(t, " ")
	invalid(t, "1% blue")
	invalid(t, "blue 10deg")
	invalid(t, "blue 4")
	invalid(t, "soylent-green 4px")
	invalid(t, "red 4px 2px")

	invalid(t, "18deg")

	invalid(t, "10arc-minutes, blue")
	invalid(t, "10px, blue")
	invalid(t, "to 90deg, blue")

	invalid(t, "to the top, blue")
	invalid(t, "to up, blue")
	invalid(t, "into top, blue")
	invalid(t, "top, blue")

	invalid(t, "to bottom up, blue")
	invalid(t, "bottom left, blue")

	capt := tu.CaptureLogs()
	gradient(t, "blue", pr.DirectionType{Angle: pi}, nil, nil)
	gradient(t, "red", pr.DirectionType{Angle: pi}, []pr.Color{red}, []pr.Dimension{{}})
	gradient(t, "blue 1%, lime,red 2em ", pr.DirectionType{Angle: pi},
		[]pr.Color{blue, lime, red}, []pr.Dimension{{Value: 1, Unit: pr.Perc}, {}, {Value: 2, Unit: pr.Em}})

	gradient(t, "18deg, blue", pr.DirectionType{Angle: pi / 10}, nil, nil)
	gradient(t, "4rad, blue", pr.DirectionType{Angle: 4}, nil, nil)
	gradient(t, ".25turn, blue", pr.DirectionType{Angle: pi / 2}, nil, nil)
	gradient(t, "100grad, blue", pr.DirectionType{Angle: (pi / 200) * 100}, nil, nil) // rounding error
	gradient(t, "12rad, blue 1%, lime,red 2em ", pr.DirectionType{Angle: 12},
		[]pr.Color{blue, lime, red}, []pr.Dimension{{Value: 1, Unit: pr.Perc}, {}, {Value: 2, Unit: pr.Em}})

	gradient(t, "to top, blue", pr.DirectionType{Angle: 0}, nil, nil)
	gradient(t, "to right, blue", pr.DirectionType{Angle: pi / 2}, nil, nil)
	gradient(t, "to bottom, blue", pr.DirectionType{Angle: pi}, nil, nil)
	gradient(t, "to left, blue", pr.DirectionType{Angle: pi * 3 / 2}, nil, nil)
	gradient(t, "to right, blue 1%, lime,red 2em ", pr.DirectionType{Angle: pi / 2},
		[]pr.Color{blue, lime, red}, []pr.Dimension{{Value: 1, Unit: pr.Perc}, {}, {Value: 2, Unit: pr.Em}})

	gradient(t, "to top left, blue", pr.DirectionType{Corner: "top_left"}, nil, nil)
	gradient(t, "to left top, blue", pr.DirectionType{Corner: "top_left"}, nil, nil)
	gradient(t, "to top right, blue", pr.DirectionType{Corner: "top_right"}, nil, nil)
	gradient(t, "to right top, blue", pr.DirectionType{Corner: "top_right"}, nil, nil)
	gradient(t, "to bottom left, blue", pr.DirectionType{Corner: "bottom_left"}, nil, nil)
	gradient(t, "to left bottom, blue", pr.DirectionType{Corner: "bottom_left"}, nil, nil)
	gradient(t, "to bottom right, blue", pr.DirectionType{Corner: "bottom_right"}, nil, nil)
	gradient(t, "to right bottom, blue", pr.DirectionType{Corner: "bottom_right"}, nil, nil)
	capt.AssertNoLogs(t)
}

func TestOverflowWrap(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "overflow-wrap: normal", toValidated(pr.Properties{
		pr.POverflowWrap: pr.String("normal"),
	}))
	assertValidDict(t, "overflow-wrap: break-word", toValidated(pr.Properties{
		pr.POverflowWrap: pr.String("break-word"),
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "overflow-wrap: none", "invalid")
	assertInvalid(t, "overflow-wrap: normal, break-word", "invalid")
}

func TestRadialGradient(t *testing.T) {
	capt := tu.CaptureLogs()

	gradient := func(t *testing.T, css string, shape string, size pr.GradientSize, center pr.Center, colors []pr.Color, stopPositions []pr.Dimension) {
		if colors == nil {
			colors = []pr.Color{blue}
		}
		if stopPositions == nil {
			stopPositions = []pr.Dimension{{}}
		}
		colorStops := make([]pr.ColorStop, len(colors))
		for i, s := range stopPositions {
			colorStops[i] = pr.ColorStop{Color: colors[i], Position: s}
		}
		if shape == "" {
			shape = "ellipse"
		}
		if size.IsNone() {
			size = pr.GradientSize{Keyword: "farthest-corner"}
		}
		if center.IsNone() {
			center = pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{{Value: 50, Unit: pr.Perc}, {Value: 50, Unit: pr.Perc}}}
		}
		checkGradientGeneric(t, css, pr.RadialGradient{ColorStops: colorStops, Shape: shape, Size: size, Center: center})
	}

	invalid := func(t *testing.T, css string) {
		invalidGeneric("radial", t, css)
	}

	invalid(t, " ")
	invalid(t, "1% blue")
	invalid(t, "blue 10deg")
	invalid(t, "blue 4")
	invalid(t, "soylent-green 4px")
	invalid(t, "red 4px 2px")

	invalid(t, "circle")
	invalid(t, "square, blue")
	invalid(t, "closest-triangle, blue")
	invalid(t, "center, blue")

	invalid(t, "ellipse 5ch")
	invalid(t, "5ch ellipse")

	invalid(t, "circle 10px 50px, blue")
	invalid(t, "10px 50px circle, blue")
	invalid(t, "10%, blue")
	invalid(t, "10% circle, blue")
	invalid(t, "circle 10%, blue")

	invalid(t, "at appex, blue")
	capt.AssertNoLogs(t)

	gradient(t, "blue", "", pr.GradientSize{}, pr.Center{}, nil, nil)
	gradient(t, "red", "", pr.GradientSize{}, pr.Center{}, []pr.Color{red}, nil)
	gradient(t, "blue 1%, lime,red 2em ", "", pr.GradientSize{}, pr.Center{},
		[]pr.Color{blue, lime, red},
		[]pr.Dimension{{Value: 1, Unit: pr.Perc}, {}, {Value: 2, Unit: pr.Em}})
	gradient(t, "circle, blue", "circle", pr.GradientSize{}, pr.Center{}, nil, nil)
	gradient(t, "ellipse, blue", "ellipse", pr.GradientSize{}, pr.Center{}, nil, nil)

	gradient(t, "ellipse closest-corner, blue",
		"ellipse", pr.GradientSize{Keyword: "closest-corner"}, pr.Center{}, nil, nil)
	gradient(t, "circle closest-side, blue",
		"circle", pr.GradientSize{Keyword: "closest-side"}, pr.Center{}, nil, nil)
	gradient(t, "farthest-corner circle, blue",
		"circle", pr.GradientSize{Keyword: "farthest-corner"}, pr.Center{}, nil, nil)
	gradient(t, "farthest-side, blue",
		"ellipse", pr.GradientSize{Keyword: "farthest-side"}, pr.Center{}, nil, nil)
	gradient(t, "5ch, blue",
		"circle", pr.GradientSize{Explicit: pr.Point{{Value: 5, Unit: pr.Ch}, {Value: 5, Unit: pr.Ch}}}, pr.Center{}, nil, nil)
	gradient(t, "5ch circle, blue",
		"circle", pr.GradientSize{Explicit: pr.Point{{Value: 5, Unit: pr.Ch}, {Value: 5, Unit: pr.Ch}}}, pr.Center{}, nil, nil)
	gradient(t, "circle 5ch, blue",
		"circle", pr.GradientSize{Explicit: pr.Point{{Value: 5, Unit: pr.Ch}, {Value: 5, Unit: pr.Ch}}}, pr.Center{}, nil, nil)

	gradient(t, "10px 50px, blue",
		"ellipse", pr.GradientSize{Explicit: pr.Point{{Value: 10, Unit: pr.Px}, {Value: 50, Unit: pr.Px}}}, pr.Center{}, nil, nil)
	gradient(t, "10px 50px ellipse, blue",
		"ellipse", pr.GradientSize{Explicit: pr.Point{{Value: 10, Unit: pr.Px}, {Value: 50, Unit: pr.Px}}}, pr.Center{}, nil, nil)
	gradient(t, "ellipse 10px 50px, blue",
		"ellipse", pr.GradientSize{Explicit: pr.Point{{Value: 10, Unit: pr.Px}, {Value: 50, Unit: pr.Px}}}, pr.Center{}, nil, nil)

	gradient(t, "10px 50px, blue",
		"ellipse", pr.GradientSize{Explicit: pr.Point{{Value: 10, Unit: pr.Px}, {Value: 50, Unit: pr.Px}}}, pr.Center{}, nil, nil)
	gradient(t, "at top 10% right, blue", "", pr.GradientSize{},
		pr.Center{OriginX: "right", OriginY: "top", Pos: pr.Point{{Value: 0, Unit: pr.Perc}, {Value: 10, Unit: pr.Perc}}}, nil, nil)
	gradient(t, "circle at bottom, blue", "circle", pr.GradientSize{},
		pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{{Value: 50, Unit: pr.Perc}, {Value: 100, Unit: pr.Perc}}}, nil, nil)
	gradient(t, "circle at 10px, blue", "circle", pr.GradientSize{},
		pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{{Value: 10, Unit: pr.Px}, {Value: 50, Unit: pr.Perc}}}, nil, nil)
	gradient(t, "closest-side circle at right 5em, blue",
		"circle", pr.GradientSize{Keyword: "closest-side"},
		pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{{Value: 100, Unit: pr.Perc}, {Value: 5, Unit: pr.Em}}}, nil, nil)
}

func TestImageOrientation(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertValidDict(t, "image-orientation: none", toValidated(pr.Properties{pr.PImageOrientation: pr.SBoolFloat{String: "none"}}))
	assertValidDict(t, "image-orientation: from-image", toValidated(pr.Properties{pr.PImageOrientation: pr.SBoolFloat{String: "from-image"}}))
	assertValidDict(t, "image-orientation: 90deg", toValidated(pr.Properties{pr.PImageOrientation: pr.SBoolFloat{Float: pi / 2, Bool: false}}))
	assertValidDict(t, "image-orientation: 30deg", toValidated(pr.Properties{pr.PImageOrientation: pr.SBoolFloat{Float: pi / 6, Bool: false}}))
	assertValidDict(t, "image-orientation: 180deg flip", toValidated(pr.Properties{pr.PImageOrientation: pr.SBoolFloat{Float: pi, Bool: true}}))
	assertValidDict(t, "image-orientation: 0deg flip", toValidated(pr.Properties{pr.PImageOrientation: pr.SBoolFloat{Float: 0, Bool: true}}))
	assertValidDict(t, "image-orientation: flip 90deg", toValidated(pr.Properties{pr.PImageOrientation: pr.SBoolFloat{Float: pi / 2, Bool: true}}))
	assertValidDict(t, "image-orientation: flip", toValidated(pr.Properties{pr.PImageOrientation: pr.SBoolFloat{Float: 0, Bool: true}}))

	assertInvalid(t, "image-orientation: none none", "invalid")
	assertInvalid(t, "image-orientation: unknown", "invalid")
	assertInvalid(t, "image-orientation: none flip", "invalid")
	assertInvalid(t, "image-orientation: from-image flip", "invalid")
	assertInvalid(t, "image-orientation: 10", "invalid")
	assertInvalid(t, "image-orientation: 10 flip", "invalid")
	assertInvalid(t, "image-orientation: flip 10", "invalid")
	assertInvalid(t, "image-orientation: flip flip", "invalid")
	assertInvalid(t, "image-orientation: 90deg flop", "invalid")
	assertInvalid(t, "image-orientation: 90deg 180deg", "invalid")
}

func TestBackgroundOriginClip(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "background-origin: content-box; background-clip: padding-box", toValidated(pr.Properties{
		pr.PBackgroundOrigin: pr.Strings{"content-box"},
		pr.PBackgroundClip:   pr.Strings{"padding-box"},
	}))
	assertValidDict(t, "background-origin: border-box;", toValidated(pr.Properties{
		pr.PBackgroundOrigin: pr.Strings{"border-box"},
	}))
	capt.AssertNoLogs(t)

	assertInvalid(t, "background-origin: 1deg", "invalid")
	assertInvalid(t, "background-origin: margin-ext-box", "invalid")
	assertInvalid(t, "background-clip: margin-ext-box", "invalid")
	assertInvalid(t, "background-clip: margin-ext-box", "invalid")
}

func TestBorderSpacing(t *testing.T) {
	capt := tu.CaptureLogs()
	assertValidDict(t, "border-spacing: 2px;", toValidated(pr.Properties{
		pr.PBorderSpacing: pr.Point{pr.FToPx(2).Dimension, pr.FToPx(2).Dimension},
	}))
	assertValidDict(t, "border-spacing: 1cm 2em;", toValidated(pr.Properties{
		pr.PBorderSpacing: pr.Point{pr.Dimension{Unit: pr.Cm, Value: 1}, pr.Dimension{Unit: pr.Em, Value: 2}},
	}))
	capt.AssertNoLogs(t)

	assertInvalid(t, "border-spacing:  eee", "invalid")
	assertInvalid(t, "border-spacing:  1cm 1cm 1cm", "invalid")
}
