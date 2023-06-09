package tree

import (
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/css/counters"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils/testutils"

	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/css/selector"
	"github.com/benoitkugler/webrender/utils"
)

// Test the CSS parsing, cascade, inherited && computed values.

func fakeHTML(html HTML) *HTML {
	html.UAStyleSheet = TestUAStylesheet
	return &html
}

func TestDescriptors(t *testing.T) {
	stylesheet := parser.ParseStylesheetBytes([]byte("@font-face{}"), false, false)
	logs := testutils.CaptureLogs()
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		nil, nil, false)
	logs.CheckEqual([]string{
		`Missing src descriptor in "@font-face" rule at 1:1`,
	}, t)

	stylesheet = parser.ParseStylesheetBytes([]byte("@font-face{src: url(test.woff)}"), false, false)
	logs = testutils.CaptureLogs()
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		nil, nil, false)
	logs.CheckEqual([]string{
		`Missing font-family descriptor in "@font-face" rule at 1:1`,
	}, t)

	stylesheet = parser.ParseStylesheetBytes([]byte("@font-face{font-family: test}"), false, false)
	logs = testutils.CaptureLogs()
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		nil, nil, false)
	logs.CheckEqual([]string{
		`Missing src descriptor in "@font-face" rule at 1:1`,
	}, t)

	stylesheet = parser.ParseStylesheetBytes([]byte("@font-face { font-family: test; src: wrong }"), false, false)
	logs = testutils.CaptureLogs()
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		nil, nil, false)
	logs.CheckEqual([]string{
		"Ignored `src: wrong ` at 1:33, invalid or unsupported values for a known CSS property.",
		`Missing src descriptor in "@font-face" rule at 1:1`,
	}, t)

	stylesheet = parser.ParseStylesheetBytes([]byte("@font-face { font-family: good, bad; src: url(test.woff) }"), false, false)
	logs = testutils.CaptureLogs()
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		nil, nil, false)
	logs.CheckEqual([]string{
		"Ignored `font-family: good, bad` at 1:14, invalid or unsupported values for a known CSS property.",
		`Missing font-family descriptor in "@font-face" rule at 1:1`,
	}, t)

	stylesheet = parser.ParseStylesheetBytes([]byte("@font-face { font-family: good, bad; src: really bad }"), false, false)
	logs = testutils.CaptureLogs()
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		nil, nil, false)
	logs.CheckEqual([]string{
		"Ignored `font-family: good, bad` at 1:14, invalid or unsupported values for a known CSS property.",
		"Ignored `src: really bad ` at 1:38, invalid or unsupported values for a known CSS property.",
		`Missing src descriptor in "@font-face" rule at 1:1`,
	}, t)
}

func resourceFilename(s string) string {
	return filepath.Join("../../resources_test", s)
}

// equivalent to python s.rsplit(sep, -1)[-1]
func rsplit(s, sep string) string {
	chunks := strings.Split(s, sep)
	return chunks[len(chunks)-1]
}

func TestFindStylesheets(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	html_, err := newHtml(utils.InputFilename(resourceFilename("doc1.html")))
	if err != nil {
		t.Fatal(err)
	}
	html := fakeHTML(*html_)
	sheets := findStylesheets(html.Root, "print", utils.DefaultUrlFetcher, html.BaseUrl, nil, nil, nil)

	if len(sheets) != 2 {
		t.Errorf("expected 2 sheets, got %d", len(sheets))
	}
	// Also test that stylesheets are in tree order
	var got [2]string
	for i, s := range sheets {
		got[i] = rsplit(rsplit(s.baseUrl, "/"), ",")
	}
	exp := [2]string{"a%7Bcolor%3AcurrentColor%7D", "doc1.html"}
	if got != exp {
		t.Errorf("expected %v got %v", exp, got)
	}

	var (
		rules      []selector.Sel
		pagesRules []PageRule
	)
	for _, sheet := range sheets {
		for _, sheetRules := range sheet.Matcher {
			rules = append(rules, sheetRules.selector...)
		}
		pagesRules = append(pagesRules, sheet.pageRules...)
	}
	if len(rules)+len(pagesRules) != 10 {
		t.Errorf("expected 10 rules, got %d", len(rules)+len(pagesRules))
	}
	// TODO: test that the values are correct too
}

// @assertNoLogs
func TestExpandShorthands(t *testing.T) {
	capt := testutils.CaptureLogs()
	sheet, err := NewCSSDefault(utils.InputFilename(resourceFilename("sheet2.css")))
	if err != nil {
		t.Fatal(err)
	}
	var sels []selector.Sel
	for _, match := range sheet.Matcher {
		sels = append(sels, match.selector...)
	}
	if len(sels) != 1 {
		t.Fatalf("expected ['li'] got %v", sels)
	}
	if sels[0].String() != "li" {
		t.Errorf("expected 'li' got %s", sels[0].String())
	}

	m := (sheet.Matcher)[0].declarations
	if m[0].Name != pr.PMarginBottom.Key() {
		t.Errorf("expected margin_bottom got %s", m[0].Name)
	}
	if (m[0].Value.ToCascaded().ToCSS().(pr.Value) != pr.Dimension{Value: 3, Unit: pr.Em}.ToValue()) {
		t.Errorf("expected got %v", m[0].Value)
	}
	if m[1].Name != pr.PMarginTop.Key() {
		t.Errorf("expected margin_top got %s", m[1].Name)
	}
	if (m[1].Value.ToCascaded().ToCSS().(pr.Value) != pr.Dimension{Value: 2, Unit: pr.Em}.ToValue()) {
		t.Errorf("expected got %v", m[1].Value)
	}
	if m[2].Name != pr.PMarginRight.Key() {
		t.Errorf("expected margin_right got %s", m[2].Name)
	}
	if (m[2].Value.ToCascaded().ToCSS().(pr.Value) != pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue()) {
		t.Errorf("expected got %v", m[2].Value)
	}
	if m[3].Name != pr.PMarginBottom.Key() {
		t.Errorf("expected margin_bottom got %s", m[3].Name)
	}
	if (m[3].Value.ToCascaded().ToCSS().(pr.Value) != pr.Dimension{Value: 2, Unit: pr.Em}.ToValue()) {
		t.Errorf("expected got %v", m[3].Value)
	}
	if m[4].Name != pr.PMarginLeft.Key() {
		t.Errorf("expected margin_left got %s", m[4].Name)
	}
	if (m[4].Value.ToCascaded().ToCSS().(pr.Value) != pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue()) {
		t.Errorf("expected got %v", m[4].Value)
	}
	if m[5].Name != pr.PMarginLeft.Key() {
		t.Errorf("expected margin_left got %s", m[5].Name)
	}
	if (m[5].Value.ToCascaded().ToCSS().(pr.Value) != pr.Dimension{Value: 4, Unit: pr.Em}.ToValue()) {
		t.Errorf("expected got %v", m[5].Value)
	}
	capt.AssertNoLogs(t)
	// TODO: test that the values are correct too
}

func assertProp(t *testing.T, got pr.ElementStyle, name pr.KnownProp, expected pr.CssProperty) {
	g := got.Get(name.Key())
	if !reflect.DeepEqual(g, expected) {
		t.Fatalf("%s - expected %v got %v", name, expected, g)
	}
}

// @assertNoLogs
func TestAnnotateDocument(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	document_, err := newHtml(utils.InputFilename(resourceFilename("doc1.html")))
	if err != nil {
		t.Fatal(err)
	}
	document := fakeHTML(*document_)
	document.UAStyleSheet, err = NewCSSDefault(utils.InputFilename(resourceFilename("mini_ua.css")))
	if err != nil {
		t.Fatal(err)
	}

	userStylesheet, err := NewCSSDefault(utils.InputFilename(resourceFilename("user.css")))
	if err != nil {
		t.Fatal(err)
	}

	styleFor := GetAllComputedStyles(document, []CSS{userStylesheet}, false, nil, nil, nil, nil, false, nil)
	// Element objects behave as lists of their children
	body := document.Root.NodeChildren(true)[1]
	children := body.NodeChildren(true)
	h1_, p_, ul_, div_ := children[0], children[1], children[2], children[3]
	li0_ := ul_.NodeChildren(true)[0]
	a_ := li0_.NodeChildren(true)[0]
	span1_ := div_.NodeChildren(true)[0]
	span2_ := span1_.NodeChildren(true)[0]

	h1 := styleFor.Get(h1_, "")
	p := styleFor.Get(p_, "")
	ul := styleFor.Get(ul_, "")
	li0 := styleFor.Get(li0_, "")
	div := styleFor.Get(div_, "")
	after := styleFor.Get(a_, "after")
	a := styleFor.Get(a_, "")
	span1 := styleFor.Get(span1_, "")
	span2 := styleFor.Get(span2_, "")

	u, err := utils.PathToURL(resourceFilename("logo_small.png"))
	if err != nil {
		t.Fatal(err)
	}
	assertProp(t, h1, pr.PBackgroundImage, pr.Images{pr.UrlImage(u)})

	assertProp(t, h1, pr.PFontWeight, pr.IntString{Int: 700})
	assertProp(t, h1, pr.PFontSize, pr.FToV(40)) // 2em

	// x-large * initial = 3/2 * 16 = 24
	assertProp(t, p, pr.PMarginTop, pr.Dimension{Value: 24, Unit: pr.Px}.ToValue())
	assertProp(t, p, pr.PMarginRight, pr.Dimension{Value: 0, Unit: pr.Px}.ToValue())
	assertProp(t, p, pr.PMarginBottom, pr.Dimension{Value: 24, Unit: pr.Px}.ToValue())
	assertProp(t, p, pr.PMarginLeft, pr.Dimension{Value: 0, Unit: pr.Px}.ToValue())
	assertProp(t, p, pr.PBackgroundColor, pr.CurrentColor)

	// 2em * 1.25ex = 2 * 20 * 1.25 * 0.8 = 40
	// 2.5ex * 1.25ex = 2.5 * 0.8 * 20 * 1.25 * 0.8 = 40
	// TODO: ex unit doesn"t work with @font-face fonts, see computedValues.py
	// assert ul["marginTop"] , pr.Dimension {Value:40,Unit: pr.Px}
	// .ToValue()assert ul["marginRight"] , pr.Dimension {Value:40,Unit: pr.Px}
	// .ToValue()assert ul["marginBottom"] , pr.Dimension {Value:40,Unit: pr.Px}
	// .ToValue()assert ul["marginLeft"] , pr.Dimension {Value:40,Unit: pr.Px}

	assertProp(t, ul, pr.PFontWeight, pr.IntString{Int: 400})
	// thick = 5px, 0.25 inches = 96*.25 = 24px
	assertProp(t, ul, pr.PBorderTopWidth, pr.FToV(0))
	assertProp(t, ul, pr.PBorderRightWidth, pr.FToV(5))
	assertProp(t, ul, pr.PBorderBottomWidth, pr.FToV(0))
	assertProp(t, ul, pr.PBorderLeftWidth, pr.FToV(24))

	assertProp(t, li0, pr.PFontWeight, pr.IntString{Int: 700})
	assertProp(t, li0, pr.PFontSize, pr.FToV(8))                                      // 6pt)
	assertProp(t, li0, pr.PMarginTop, pr.Dimension{Value: 16, Unit: pr.Px}.ToValue()) // 2em)
	assertProp(t, li0, pr.PMarginRight, pr.Dimension{Value: 0, Unit: pr.Px}.ToValue())
	assertProp(t, li0, pr.PMarginBottom, pr.Dimension{Value: 16, Unit: pr.Px}.ToValue())
	assertProp(t, li0, pr.PMarginLeft, pr.Dimension{Value: 32, Unit: pr.Px}.ToValue()) // 4em)

	assertProp(t, a, pr.PTextDecorationLine, pr.Decorations(utils.NewSet("underline")))
	assertProp(t, a, pr.PFontWeight, pr.IntString{Int: 900})
	assertProp(t, a, pr.PFontSize, pr.FToV(24)) // 300% of 8px)
	assertProp(t, a, pr.PPaddingTop, pr.Dimension{Value: 1, Unit: pr.Px}.ToValue())
	assertProp(t, a, pr.PPaddingRight, pr.Dimension{Value: 2, Unit: pr.Px}.ToValue())
	assertProp(t, a, pr.PPaddingBottom, pr.Dimension{Value: 3, Unit: pr.Px}.ToValue())
	assertProp(t, a, pr.PPaddingLeft, pr.Dimension{Value: 4, Unit: pr.Px}.ToValue())
	assertProp(t, a, pr.PBorderTopWidth, pr.FToV(42))
	assertProp(t, a, pr.PBorderBottomWidth, pr.FToV(42))

	assertProp(t, a, pr.PColor, pr.NewColor(1, 0, 0, 1))
	assertProp(t, a, pr.PBorderTopColor, pr.CurrentColor)

	assertProp(t, div, pr.PFontSize, pr.FToV(40))                                     // 2 * 20px)
	assertProp(t, span1, pr.PWidth, pr.Dimension{Value: 160, Unit: pr.Px}.ToValue())  // 10 * 16px (Root default is 16px))
	assertProp(t, span1, pr.PHeight, pr.Dimension{Value: 400, Unit: pr.Px}.ToValue()) // 10 * (2 * 20px))
	assertProp(t, span2, pr.PFontSize, pr.FToV(32))

	// The href attr should be as in the source, not made absolute.
	assertProp(t, after, pr.PBackgroundColor, pr.NewColor(1, 0, 0, 1))
	assertProp(t, after, pr.PBorderTopWidth, pr.FToV(42))
	assertProp(t, after, pr.PBorderBottomWidth, pr.FToV(3))
	assertProp(t, after, pr.PContent, pr.SContent{Contents: pr.ContentProperties{{Type: "string", Content: pr.String(" [")}, {Type: "string", Content: pr.String("home.html")}, {Type: "string", Content: pr.String("]")}}})

	// TODO: much more tests here: test that origin and selector precedence
	// and inheritance are correct…
}

// @assertNoLogs
func TestPage(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	document_, err := newHtml(utils.InputFilename(resourceFilename("doc1.html")))
	if err != nil {
		t.Fatal(err)
	}
	document := fakeHTML(*document_)
	css, err := NewCSSDefault(utils.InputString(`
		html { color: red }
		@page { margin: 10px }
		@page :right {
		color: blue;
		margin-bottom: 12pt;
		font-size: 20px;
		@top-left { width: 10em }
		@top-right { font-size: 10px}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	styleFor := GetAllComputedStyles(document, []CSS{css}, false, nil, nil, nil, nil, false, nil)

	pageType := utils.PageElement{Side: "left", First: true, Blank: false, Index: 0, Name: ""}
	styleFor.SetPageComputedStylesT(pageType, document)
	style := styleFor.Get(pageType, "")
	assertProp(t, style, pr.PMarginTop, pr.Dimension{Value: 5, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PMarginLeft, pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PMarginBottom, pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PColor, pr.NewColor(1, 0, 0, 1)) // red, inherited from html

	pageType = utils.PageElement{Side: "right", First: true, Blank: false, Index: 0, Name: ""}
	styleFor.SetPageComputedStylesT(pageType, document)
	style = styleFor.Get(pageType, "")
	assertProp(t, style, pr.PMarginTop, pr.Dimension{Value: 5, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PMarginLeft, pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PMarginBottom, pr.Dimension{Value: 16, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PColor, pr.NewColor(0, 0, 1, 1)) // blue

	pageType = utils.PageElement{Side: "left", First: false, Blank: false, Index: 1, Name: ""}
	styleFor.SetPageComputedStylesT(pageType, document)
	style = styleFor.Get(pageType, "")
	assertProp(t, style, pr.PMarginTop, pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PMarginLeft, pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PMarginBottom, pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PColor, pr.NewColor(1, 0, 0, 1)) // red, inherited from html

	pageType = utils.PageElement{Side: "right", First: false, Blank: false, Index: 1, Name: ""}
	styleFor.SetPageComputedStylesT(pageType, document)
	style = styleFor.Get(pageType, "")
	assertProp(t, style, pr.PMarginTop, pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PMarginLeft, pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PMarginBottom, pr.Dimension{Value: 16, Unit: pr.Px}.ToValue())
	assertProp(t, style, pr.PColor, pr.NewColor(0, 0, 1, 1)) // blue

	pageType = utils.PageElement{Side: "left", First: true, Blank: false, Index: 0, Name: ""}
	styleFor.SetPageComputedStylesT(pageType, document)
	style = styleFor.Get(pageType, "@top-left")
	if style != nil {
		t.Fatal("expected empty (nil) style")
	}

	pageType = utils.PageElement{Side: "right", First: true, Blank: false, Index: 0, Name: ""}
	styleFor.SetPageComputedStylesT(pageType, document)
	style = styleFor.Get(pageType, "@top-left")
	assertProp(t, style, pr.PFontSize, pr.FToV(20)) // inherited from @page
	assertProp(t, style, pr.PWidth, pr.Dimension{Value: 200, Unit: pr.Px}.ToValue())

	pageType = utils.PageElement{Side: "right", First: true, Blank: false, Index: 0, Name: ""}
	styleFor.SetPageComputedStylesT(pageType, document)
	style = styleFor.Get(pageType, "@top-right")
	assertProp(t, style, pr.PFontSize, pr.FToV(10))
}

type testPageSelector struct {
	sel string
	out []pageSelector
}

var tests = []testPageSelector{
	{sel: "@page {}", out: []pageSelector{
		{Specificity: selector.Specificity{0, 0, 0}},
	}},
	{sel: "@page :left {}", out: []pageSelector{
		{Side: "left", Specificity: selector.Specificity{0, 0, 1}},
	}},
	{sel: "@page:first:left {}", out: []pageSelector{
		{Side: "left", First: true, Specificity: selector.Specificity{0, 1, 1}},
	}},
	{sel: "@page pagename {}", out: []pageSelector{
		{Name: "pagename", Specificity: selector.Specificity{1, 0, 0}},
	}},
	{sel: "@page pagename:first:right:blank {}", out: []pageSelector{
		{Side: "right", Blank: true, First: true, Name: "pagename", Specificity: selector.Specificity{1, 2, 1}},
	}},
	{sel: "@page pagename, :first {}", out: []pageSelector{
		{Name: "pagename", Specificity: selector.Specificity{1, 0, 0}},
		{First: true, Specificity: selector.Specificity{0, 1, 0}},
	}},
	{sel: "@page :first:first {}", out: []pageSelector{
		{First: true, Specificity: selector.Specificity{0, 2, 0}},
	}},
	{sel: "@page :left:left {}", out: []pageSelector{
		{Side: "left", Specificity: selector.Specificity{0, 0, 2}},
	}},
	{sel: "@page :nth(2) {}", out: []pageSelector{
		{Index: pageIndex{A: 0, B: 2}, Specificity: selector.Specificity{0, 1, 0}},
	}},
	{sel: "@page :nth(2n + 4) {}", out: []pageSelector{
		{Index: pageIndex{A: 2, B: 4}, Specificity: selector.Specificity{0, 1, 0}},
	}},
	{sel: "@page :nth(3n) {}", out: []pageSelector{
		{Index: pageIndex{A: 3, B: 0}, Specificity: selector.Specificity{0, 1, 0}},
	}},
	{sel: "@page :nth( n+2 ) {}", out: []pageSelector{
		{Index: pageIndex{A: 1, B: 2}, Specificity: selector.Specificity{0, 1, 0}},
	}},
	{sel: "@page :nth(even) {}", out: []pageSelector{
		{Index: pageIndex{A: 2, B: 0}, Specificity: selector.Specificity{0, 1, 0}},
	}},
	{sel: "@page pagename:nth(2) {}", out: []pageSelector{
		{Name: "pagename", Index: pageIndex{A: 0, B: 2}, Specificity: selector.Specificity{1, 1, 0}},
	}},
	{sel: "@page page page {}"},
	{sel: "@page :left page {}"},
	{sel: "@page :left, {}"},
	{sel: "@page , {}"},
	{sel: "@page :left, test, {}"},
	{sel: "@page :wrong {}"},
	{sel: "@page :left:wrong {}"},
	{sel: "@page :left:right {}"},
}

func TestPageSelectors(t *testing.T) {
	capt := testutils.CaptureLogs()
	for _, te := range tests {
		atRule_ := parser.ParseStylesheetBytes([]byte(te.sel), false, false)[0]
		atRule, ok := atRule_.(parser.QualifiedRule)
		if !ok {
			atRule = atRule_.(parser.AtRule).QualifiedRule
		}
		res := parsePageSelectors(atRule)
		if !reflect.DeepEqual(res, te.out) {
			t.Fatalf("%s : expected %v got %v", te.sel, te.out, res)
		}
	}
	capt.AssertNoLogs(t)
}

type testWarnings struct {
	sel string
	out []string
}

var testsWarnings = [6]testWarnings{
	{
		sel: ":lipsum { margin: 2cm",
		out: []string{"Invalid or unsupported selector"},
	},
	{
		sel: "::lipsum { margin: 2cm",
		out: []string{"Invalid or unsupported selector"},
	},
	{
		sel: "foo { margin-color: red",
		out: []string{"Ignored", "unknown property"},
	},
	{
		sel: "foo { margin-top: red",
		out: []string{"Ignored", "invalid value"},
	},
	{
		sel: `@import "relative-uri.css"`,
		out: []string{"Relative URI reference without a base URI"},
	},
	{
		sel: `@import "invalid-protocol://absolute-URL"`,
		out: []string{"Failed to load stylesheet at"},
	},
}

// @assertNoLogs
// Check that appropriate warnings are logged.
func TestWarnings(t *testing.T) {
	for _, te := range testsWarnings {

		capt := testutils.CaptureLogs()
		_, err := NewCSSDefault(utils.InputString(te.sel))
		if err != nil {
			t.Fatal(err)
		}
		logs := capt.Logs()
		if len(logs) != 1 {
			t.Fatalf("%s : expected exactly 1 log, got %d", te.sel, len(logs))
		}
		for _, message := range te.out {
			if !strings.Contains(logs[0], message) {
				t.Fatalf("log should contain %s, got %s", message, logs[0])
			}
		}
	}
}

// @assertNoLogs
func TestWarningsStylesheet(t *testing.T) {
	ml := "<link rel=stylesheet href=invalid-protocol://absolute>"
	capt := testutils.CaptureLogs()
	html, err := newHtml(utils.InputString(ml))
	if err != nil {
		t.Fatal(err)
	}
	GetAllComputedStyles(html, nil, false, nil, nil, nil, nil, false, nil)
	logs := capt.Logs()
	if len(logs) != 1 {
		t.Fatalf("expected exactly 1 log, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "Failed to load stylesheet at") {
		t.Fatalf("log should contain 'Failed to load stylesheet at', got %s", logs[0])
	}
}

type testFontSize struct {
	parentCss, childCss   string
	parentSize, childSize pr.Float
}

var testsFs = []testFontSize{
	{parentCss: "10px", parentSize: 10, childCss: "10px", childSize: 10},
	{parentCss: "x-small", parentSize: 12, childCss: "xx-large", childSize: 32},
	{parentCss: "x-large", parentSize: 24, childCss: "2em", childSize: 48},
	{parentCss: "1em", parentSize: 16, childCss: "1em", childSize: 16},
	{parentCss: "1em", parentSize: 16, childCss: "larger", childSize: 6. / 5 * 16},
	{parentCss: "medium", parentSize: 16, childCss: "larger", childSize: 6. / 5 * 16},
	{parentCss: "x-large", parentSize: 24, childCss: "larger", childSize: 32},
	{parentCss: "xx-large", parentSize: 32, childCss: "larger", childSize: 1.2 * 32},
	{parentCss: "1px", parentSize: 1, childCss: "larger", childSize: 3. / 5 * 16},
	{parentCss: "28px", parentSize: 28, childCss: "larger", childSize: 32},
	{parentCss: "100px", parentSize: 100, childCss: "larger", childSize: 120},
	{parentCss: "xx-small", parentSize: 3. / 5 * 16, childCss: "larger", childSize: 12},
	{parentCss: "1em", parentSize: 16, childCss: "smaller", childSize: 8. / 9 * 16},
	{parentCss: "medium", parentSize: 16, childCss: "smaller", childSize: 8. / 9 * 16},
	{parentCss: "x-large", parentSize: 24, childCss: "smaller", childSize: 6. / 5 * 16},
	{parentCss: "xx-large", parentSize: 32, childCss: "smaller", childSize: 24},
	{parentCss: "xx-small", parentSize: 3. / 5 * 16, childCss: "smaller", childSize: 0.8 * 3. / 5 * 16},
	{parentCss: "1px", parentSize: 1, childCss: "smaller", childSize: 0.8},
	{parentCss: "28px", parentSize: 28, childCss: "smaller", childSize: 24},
	{parentCss: "100px", parentSize: 100, childCss: "smaller", childSize: 32},
}

func isClose(a, b pr.Float) bool {
	return math.Abs(math.Round(float64(a-b))) < 1e-5
}

func TestFontSize(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	html_, err := newHtml(utils.InputString("<p>a<span>b"))
	if err != nil {
		t.Fatal(err)
	}
	document := fakeHTML(*html_)
	for _, te := range testsFs {
		css, err := NewCSSDefault(utils.InputString(fmt.Sprintf("p{font-size:%s}span{font-size:%s}", te.parentCss, te.childCss)))
		if err != nil {
			t.Fatal(err)
		}
		styleFor := GetAllComputedStyles(document, []CSS{css}, false, nil, nil, nil, nil, false, nil)
		body := document.Root.NodeChildren(true)[1]
		p := body.NodeChildren(true)[0]
		span := p.NodeChildren(true)[1]
		if got := styleFor.Get(p, "").GetFontSize(); !isClose(got.Value, te.parentSize) {
			t.Fatalf("parent: expected %v got %v", te.parentSize, got)
		}
		if got := styleFor.Get(span, "").GetFontSize(); !isClose(got.Value, te.childSize) {
			t.Fatalf("child:expected %v got %v", te.childSize, got)
		}
	}
}

func TestCounterStyleInvalid(t *testing.T) {
	inputs := []string{
		"@counter-style test {system: alphabetic; symbols: a}",
		"@counter-style test {system: cyclic}",
		"@counter-style test {system: additive; additive-symbols: a 1}",
		"@counter-style test {system: additive; additive-symbols: 10 x, 1 i, 5 v}",
	}
	for _, rule := range inputs {
		stylesheet := parser.ParseStylesheetBytes([]byte(rule), false, false)
		cp := testutils.CaptureLogs()

		preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
			nil, make(counters.CounterStyle), false)
		if len(cp.Logs()) == 0 {
			t.Fatal("expected logs")
		}
	}
}
