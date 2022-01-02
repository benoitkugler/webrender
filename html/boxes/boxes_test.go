package boxes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/css/counters"
	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/images"
	"github.com/benoitkugler/webrender/utils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

var (
	_ ReplacedBoxITF = (*ReplacedBox)(nil)
	_ ReplacedBoxITF = (*BlockReplacedBox)(nil)
	_ ReplacedBoxITF = (*InlineReplacedBox)(nil)

	_ BlockLevelBoxITF = (*BlockBox)(nil)
	_ BlockLevelBoxITF = (*BlockReplacedBox)(nil)
	_ BlockLevelBoxITF = (*TableBox)(nil)
	_ BlockLevelBoxITF = (*FlexBox)(nil)

	_ TableBoxITF = (*TableBox)(nil)
	_ TableBoxITF = (*InlineTableBox)(nil)
)

//  Test that the "before layout" box tree is correctly constructed.

func fakeHTML(html *tree.HTML) *tree.HTML {
	html.UAStyleSheet = tree.TestUAStylesheet
	return html
}

func parseBase(t testing.TB, content utils.ContentInput, baseUrl string) (*utils.HTMLNode, *tree.StyleFor, Gifu, string, *tree.TargetCollector, counters.CounterStyle) {
	t.Helper()

	html, err := tree.NewHTML(content, baseUrl, utils.DefaultUrlFetcher, "")
	if err != nil {
		t.Fatalf("parsing HTML failed: %s", err)
	}
	document := fakeHTML(html)
	cs := make(counters.CounterStyle)
	style := tree.GetAllComputedStyles(document, nil, false, nil, cs, nil, nil, nil)
	imgFetcher := func(url string, forcedMimeType string) images.Image {
		return images.GetImageFromUri(images.NewCache(), document.UrlFetcher, false, url, forcedMimeType)
	}
	tr := tree.NewTargetCollector()
	return document.Root, style, imgFetcher, html.BaseUrl, &tr, cs
}

func parse(t *testing.T, htmlContent string) BoxITF {
	a, b, c, d, e, f := parseBase(t, utils.InputString(htmlContent), baseUrl)
	boxes := elementToBox(a, b, c, d, e, f, nil)
	return boxes[0]
}

func parseAndBuild(t *testing.T, htmlContent string) BlockLevelBoxITF {
	return parseAndBuildExt(t, htmlContent, baseUrl)
}

func parseAndBuildExt(t *testing.T, htmlContent, baseUrl string) BlockLevelBoxITF {
	box := BuildFormattingStructure(parseBase(t, utils.InputString(htmlContent), baseUrl))
	if err := sanityChecks(box); err != nil {
		t.Fatalf("sanity check failed: %s", err)
	}
	return box
}

// Check the box tree equality.
//
// The obtained result is prettified in the message in case of failure.
//
// box: a Box object, starting with <html> and <body> blocks.
// expected: a list of serialized <body> children as returned by to_lists().
func assertTree(t *testing.T, box Box, expected []SerBox) {
	t.Helper()

	if tag := box.Box().ElementTag(); tag != "html" {
		t.Fatalf("unexpected element: %s", tag)
	}
	if !BlockBoxT.IsInstance(box) {
		t.Fatal("expected block box")
	}
	if L := len(box.Box().Children); L != 1 {
		t.Fatalf("expected one children, got %d", L)
	}

	box = box.Box().Children[0]
	if !BlockBoxT.IsInstance(box) {
		t.Fatal("expected block box")
	}
	if tag := box.Box().ElementTag(); tag != "body" {
		t.Fatalf("unexpected element: %s", tag)
	}

	if got := Serialize(box.Box().Children); !SerializedBoxEquals(got, expected) {
		t.Fatalf("expected \n%v\n, got\n%v", expected, got)
	}
}

var properChildren = map[BoxType][]BoxType{
	BlockContainerBoxT: {BlockLevelBoxT, LineBoxT},
	LineBoxT:           {InlineLevelBoxT},
	InlineBoxT:         {InlineLevelBoxT},
	TableBoxT: {
		TableCaptionBoxT,
		TableColumnGroupBoxT, TableColumnBoxT,
		TableRowGroupBoxT, TableRowBoxT,
	},
	InlineTableBoxT: {
		TableCaptionBoxT,
		TableColumnGroupBoxT, TableColumnBoxT,
		TableRowGroupBoxT, TableRowBoxT,
	},
	TableColumnGroupBoxT: {TableColumnBoxT},
	TableRowGroupBoxT:    {TableRowBoxT},
	TableRowBoxT:         {TableCellBoxT},
}

// Check that the rules regarding boxes are met.
//
// This is not required and only helps debugging.
//
// - A block container can contain either only block-level boxes or
//   only line boxes;
// - Line boxes and inline boxes can only contain inline-level
func sanityChecks(box Box) error {
	if !ParentBoxT.IsInstance(box) {
		return nil
	}

	acceptablesListsT, ok := properChildren[box.Type()]
	if !ok {
		return nil // this is less strict than the reference implementation
	}

	for _, child := range box.Box().Children {
		if !child.Box().IsInNormalFlow() {
			continue
		}
		isOk := false
		for _, typeOk := range acceptablesListsT {
			if typeOk.IsInstance(child) {
				isOk = true
				break
			}
		}
		if !isOk {
			return errors.New("invalid children check")
		}
	}

	for _, child := range box.Box().Children {
		if err := sanityChecks(child); err != nil {
			return err
		}
	}

	return nil
}

// func _parse_base(htmlContent, base_url=BASE_URL):
//     document = FakeHTML(string=htmlContent, base_url=base_url)
//     counter_style = CounterStyle()
//     style_for = get_all_computed_styles(document, counter_style=counter_style)
//     get_image_from_uri = functools.partial(
//         images.get_image_from_uri, cache={}, url_fetcher=document.url_fetcher,
//         optimize_size=())
//     target_collector = TargetCollector()
//     return (
//         document.etree_element, style_for, get_image_from_uri, base_url,
//         target_collector, counter_style)

var baseUrl, _ = utils.PathToURL("../../resources_test/")

func getGrid(t *testing.T, html string) ([][]Border, [][]Border) {
	root := parseAndBuild(t, html)
	body := root.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0].(TableBoxITF)

	buildGrid := func(bg [][]Border) (grid [][]Border /*maybe nil*/) {
		for _, column := range bg {
			out := make([]Border, len(column))
			for i, border := range column {
				if border.Width != 0 {
					border.Score = Score{}
					out[i] = border
				}
			}
			grid = append(grid, out)
		}
		return grid
	}

	return buildGrid(table.Table().CollapsedBorderGrid.Vertical),
		buildGrid(table.Table().CollapsedBorderGrid.Horizontal)
}

func TestBoxTree(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parse(t, "<p>"), []SerBox{{"p", BlockBoxT, BC{}}})
	assertTree(t, parse(t, `
	  <style>
	    span { display: inline-block }
	  </style>
	  <p>Hello <em>World <img src="pattern.png"><span>L</span></em>!</p>`),
		[]SerBox{
			{
				"p", BlockBoxT, BC{C: []SerBox{
					{"p", TextBoxT, BC{Text: "Hello "}},
					{"em", InlineBoxT, BC{C: []SerBox{
						{"em", TextBoxT, BC{Text: "World "}},
						{"img", InlineReplacedBoxT, BC{Text: "<replaced>"}},
						{"span", InlineBlockBoxT, BC{C: []SerBox{
							{"span", TextBoxT, BC{Text: "L"}},
						}}},
					}}},
					{"p", TextBoxT, BC{Text: "!"}},
				}},
			},
		})
}

func TestHtmlEntities(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	for _, quote := range []string{`"`, "&quot;", "&#x22;", "&#34;"} {
		assertTree(t, parse(t, fmt.Sprintf("<p>%sabc%s", quote, quote)), []SerBox{
			{"p", BlockBoxT, BC{C: []SerBox{
				{"p", TextBoxT, BC{Text: `"abc"`}},
			}}},
		})
	}
}

func TestInlineInBlock1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	source := "<div>Hello, <em>World</em>!\n<p>Lipsum.</p></div>"
	expected := []SerBox{
		{"div", BlockBoxT, BC{
			C: []SerBox{
				{
					"div", BlockBoxT,
					BC{C: []SerBox{
						{"div", LineBoxT, BC{C: []SerBox{
							{"div", TextBoxT, BC{Text: "Hello, "}},
							{"em", InlineBoxT, BC{C: []SerBox{
								{"em", TextBoxT, BC{Text: "World"}},
							}}},
							{"div", TextBoxT, BC{Text: "! "}},
						}}},
					}},
				},
				{"p", BlockBoxT, BC{C: []SerBox{
					{"p", LineBoxT, BC{C: []SerBox{
						{"p", TextBoxT, BC{Text: "Lipsum."}},
					}}},
				}}},
			},
		}},
	}
	box := parse(t, source)

	assertTree(t, box, []SerBox{
		{"div", BlockBoxT, BC{C: []SerBox{
			{"div", TextBoxT, BC{Text: "Hello, "}},
			{"em", InlineBoxT, BC{C: []SerBox{
				{"em", TextBoxT, BC{Text: "World"}},
			}}},
			{"div", TextBoxT, BC{Text: "! "}},
			{"p", BlockBoxT, BC{C: []SerBox{{"p", TextBoxT, BC{Text: "Lipsum."}}}}},
		}}},
	})

	box = InlineInBlock(box)
	assertTree(t, box, expected)
}

func TestInlineInBlock2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	source := "<div><p>Lipsum.</p>Hello, <em>World</em>!\n</div>"
	expected := []SerBox{
		{"div", BlockBoxT, BC{C: []SerBox{
			{"p", BlockBoxT, BC{C: []SerBox{{"p", LineBoxT, BC{C: []SerBox{{"p", TextBoxT, BC{Text: "Lipsum."}}}}}}}},
			{"div", BlockBoxT, BC{C: []SerBox{
				{"div", LineBoxT, BC{C: []SerBox{
					{"div", TextBoxT, BC{Text: "Hello, "}},
					{"em", InlineBoxT, BC{C: []SerBox{{"em", TextBoxT, BC{Text: "World"}}}}},
					{"div", TextBoxT, BC{Text: "! "}},
				}}},
			}}},
		}}},
	}
	box := parse(t, source)
	box = InlineInBlock(box)
	assertTree(t, box, expected)
}

func TestInlineInBlock3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Absolutes are left := range the lines to get their static position later.
	source := `<p>Hello <em style="position:absolute;
                                    display: block">World</em>!</p>`
	expected := []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{
			{"p", LineBoxT, BC{C: []SerBox{
				{"p", TextBoxT, BC{Text: "Hello "}},
				{"em", BlockBoxT, BC{C: []SerBox{{"em", LineBoxT, BC{C: []SerBox{{"em", TextBoxT, BC{Text: "World"}}}}}}}},
				{"p", TextBoxT, BC{Text: "!"}},
			}}},
		}}},
	}
	box := parse(t, source)
	box = InlineInBlock(box)
	assertTree(t, box, expected)
	box = BlockInInline(box)
	assertTree(t, box, expected)
}

func TestInlineInBlock4(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Floats are pull to the top of their containing blocks
	source := `<p>Hello <em style="float: left">World</em>!</p>`

	expected := []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{
			{"p", LineBoxT, BC{C: []SerBox{
				{"p", TextBoxT, BC{Text: "Hello "}},
				{"em", BlockBoxT, BC{C: []SerBox{{"em", LineBoxT, BC{C: []SerBox{{"em", TextBoxT, BC{Text: "World"}}}}}}}},
				{"p", TextBoxT, BC{Text: "!"}},
			}}},
		}}},
	}
	box := parse(t, source)
	box = InlineInBlock(box)
	box = BlockInInline(box)
	assertTree(t, box, expected)
}

func TestBlockInInline(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	box := parse(t, `
      <style>
        p { display: inline-block; }
        span, i { display: block; }
      </style>
      <p>Lorem <em>ipsum <strong>dolor <span>sit</span>
      <span>amet,</span></strong><span><em>conse<i>`)
	box = InlineInBlock(box)
	assertTree(t, box, []SerBox{
		{"body", LineBoxT, BC{C: []SerBox{
			{"p", InlineBlockBoxT, BC{C: []SerBox{
				{"p", LineBoxT, BC{C: []SerBox{
					{"p", TextBoxT, BC{Text: "Lorem "}},
					{"em", InlineBoxT, BC{C: []SerBox{
						{"em", TextBoxT, BC{Text: "ipsum "}},
						{"strong", InlineBoxT, BC{C: []SerBox{
							{"strong", TextBoxT, BC{Text: "dolor "}},
							{"span", BlockBoxT, BC{C: []SerBox{{"span", LineBoxT, BC{C: []SerBox{{"span", TextBoxT, BC{Text: "sit"}}}}}}}},
							{"strong", TextBoxT, BC{Text: " "}},
							{"span", BlockBoxT, BC{C: []SerBox{{"span", LineBoxT, BC{C: []SerBox{{"span", TextBoxT, BC{Text: "amet,"}}}}}}}},
						}}},
						{"span", BlockBoxT, BC{C: []SerBox{
							{"span", LineBoxT, BC{C: []SerBox{
								{"em", InlineBoxT, BC{C: []SerBox{
									{"em", TextBoxT, BC{Text: "conse"}},
									{"i", BlockBoxT, BC{C: []SerBox{}}},
								}}},
							}}},
						}}},
					}}},
				}}},
			}}},
		}}},
	})

	box = BlockInInline(box)
	assertTree(t, box, []SerBox{
		{"body", LineBoxT, BC{C: []SerBox{
			{"p", InlineBlockBoxT, BC{C: []SerBox{
				{"p", BlockBoxT, BC{C: []SerBox{
					{"p", LineBoxT, BC{C: []SerBox{
						{"p", TextBoxT, BC{Text: "Lorem "}},
						{"em", InlineBoxT, BC{C: []SerBox{
							{"em", TextBoxT, BC{Text: "ipsum "}},
							{"strong", InlineBoxT, BC{C: []SerBox{{"strong", TextBoxT, BC{Text: "dolor "}}}}},
						}}},
					}}},
				}}},
				{"span", BlockBoxT, BC{C: []SerBox{{"span", LineBoxT, BC{C: []SerBox{{"span", TextBoxT, BC{Text: "sit"}}}}}}}},
				{"p", BlockBoxT, BC{C: []SerBox{
					{"p", LineBoxT, BC{C: []SerBox{
						{"em", InlineBoxT, BC{C: []SerBox{{"strong", InlineBoxT, BC{C: []SerBox{{"strong", TextBoxT, BC{Text: " "}}}}}}}},
					}}},
				}}},
				{"span", BlockBoxT, BC{C: []SerBox{{"span", LineBoxT, BC{C: []SerBox{{"span", TextBoxT, BC{Text: "amet,"}}}}}}}},
				{"p", BlockBoxT, BC{C: []SerBox{
					{"p", LineBoxT, BC{C: []SerBox{{"em", InlineBoxT, BC{C: []SerBox{{"strong", InlineBoxT, BC{C: []SerBox{}}}}}}}}},
				}}},
				{"span", BlockBoxT, BC{C: []SerBox{
					{"span", BlockBoxT, BC{C: []SerBox{
						{"span", LineBoxT, BC{C: []SerBox{{"em", InlineBoxT, BC{C: []SerBox{{"em", TextBoxT, BC{Text: "conse"}}}}}}}},
					}}},
					{"i", BlockBoxT, BC{C: []SerBox{}}},
					{"span", BlockBoxT, BC{C: []SerBox{{"span", LineBoxT, BC{C: []SerBox{{"em", InlineBoxT, BC{C: []SerBox{}}}}}}}}},
				}}},
				{"p", BlockBoxT, BC{C: []SerBox{{"p", LineBoxT, BC{C: []SerBox{{"em", InlineBoxT, BC{C: []SerBox{}}}}}}}}},
			}}},
		}}},
	})
}

func TestStyles(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	box := parse(t, `
		  <style>
			span { display: block; }
			* { margin: 42px }
			html { color: blue }
		  </style>
		  <p>Lorem <em>ipsum <strong>dolor <span>sit</span>
			<span>amet,</span></strong><span>consectetur</span></em></p>`)
	box = InlineInBlock(box)
	box = BlockInInline(box)

	descendants := Descendants(box)
	if L := len(descendants); L != 31 {
		t.Fatalf("expected 31 descendants, got %d", L)
	}
	if d := descendants[0]; d != box {
		t.Fatalf("expected box to be the first descendant, got %v", d)
	}

	for _, child := range descendants {
		// All boxes inherit the color
		if c := child.Box().Style.GetColor(); c.RGBA != (parser.RGBA{R: 0, G: 0, B: 1, A: 1}) { // blue
			t.Fatal()
		}
		// Only non-anonymous boxes have margins
		if mt := child.Box().Style.GetMarginTop(); mt != pr.FToPx(0) && mt != pr.FToPx(42) {
			t.Fatal()
		}
	}
}

func TestWhitespaces(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// http://www.w3.org/TR/CSS21/text.html#white-space-model
	assertTree(t, parseAndBuild(t, "<p>Lorem \t\r\n  ipsum\t"+`<strong>  dolor
		<img src=pattern.png> sit
        <span style="position: absolute"></span> <em> amet </em>
        consectetur</strong>.</p>`+
		"<pre>\t  foo\n</pre>"+
		"<pre style=\"white-space: pre-wrap\">\t  foo\n</pre>"+
		"<pre style=\"white-space: pre-line\">\t  foo\n</pre>"), []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{
			{"p", LineBoxT, BC{C: []SerBox{
				{"p", TextBoxT, BC{Text: "Lorem ipsum "}},
				{"strong", InlineBoxT, BC{C: []SerBox{
					{"strong", TextBoxT, BC{Text: "dolor "}},
					{"img", InlineReplacedBoxT, BC{Text: "<replaced>"}},
					{"strong", TextBoxT, BC{Text: " sit "}},
					{"span", BlockBoxT, BC{C: []SerBox{}}},
					{"em", InlineBoxT, BC{C: []SerBox{{"em", TextBoxT, BC{Text: "amet "}}}}},
					{"strong", TextBoxT, BC{Text: "consectetur"}},
				}}},
				{"p", TextBoxT, BC{Text: "."}},
			}}},
		}}},
		{"pre", BlockBoxT, BC{C: []SerBox{{"pre", LineBoxT, BC{C: []SerBox{{"pre", TextBoxT, BC{Text: "\t  foo\n"}}}}}}}},
		{"pre", BlockBoxT, BC{C: []SerBox{{"pre", LineBoxT, BC{C: []SerBox{{"pre", TextBoxT, BC{Text: "\t  foo\n"}}}}}}}},
		{"pre", BlockBoxT, BC{C: []SerBox{{"pre", LineBoxT, BC{C: []SerBox{{"pre", TextBoxT, BC{Text: " foo\n"}}}}}}}},
	})
}

type pageStyleData struct {
	type_                    utils.PageElement
	top, right, bottom, left pr.Float
}

func testPageStyle(t *testing.T, data pageStyleData) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	document, err := tree.NewHTML(utils.InputString(`
      <style>
        @page { margin: 3px }
        @page name { margin-left: 15px; margin-top: 5px }
        @page :nth(3) { margin-bottom: 1px }
        @page :nth(5n+4) { margin-bottom: 2px }
        @page :first { margin-top: 20px }
        @page :right { margin-right: 10px; margin-top: 10px }
        @page :left { margin-left: 10px; margin-top: 10px }
      </style>
    `), "", utils.DefaultUrlFetcher, "")
	if err != nil {
		t.Fatal(err)
	}
	document = fakeHTML(document)
	styleFor := tree.GetAllComputedStyles(document, nil, false, nil, nil, nil, nil, nil)

	// Force the generation of the style for this page type as it"s generally
	// only done during the rendering.
	styleFor.SetPageComputedStylesT(data.type_, document)

	style := styleFor.Get(data.type_, "")
	if m := style.GetMarginTop(); m != pr.FToPx(data.top) {
		t.Fatalf("expected %f, got %v", data.top, m)
	}
	if m := style.GetMarginRight(); m != pr.FToPx(data.right) {
		t.Fatalf("expected %f, got %v", data.right, m)
	}
	if m := style.GetMarginBottom(); m != pr.FToPx(data.bottom) {
		t.Fatalf("expected %f, got %v", data.bottom, m)
	}
	if m := style.GetMarginLeft(); m != pr.FToPx(data.left) {
		t.Fatalf("expected %f, got %v", data.left, m)
	}
}

func TestPageStyle(t *testing.T) {
	for _, data := range []pageStyleData{
		{utils.PageElement{Side: "left", First: true, Index: 0, Blank: false, Name: ""}, 20, 3, 3, 10},
		{utils.PageElement{Side: "right", First: true, Index: 0, Blank: false, Name: ""}, 20, 10, 3, 3},
		{utils.PageElement{Side: "left", First: false, Index: 1, Blank: false, Name: ""}, 10, 3, 3, 10},
		{utils.PageElement{Side: "right", First: false, Index: 1, Blank: false, Name: ""}, 10, 10, 3, 3},
		{utils.PageElement{Side: "right", First: false, Index: 1, Blank: false, Name: "name"}, 5, 10, 3, 15},
		{utils.PageElement{Side: "right", First: false, Index: 2, Blank: false, Name: "name"}, 5, 10, 1, 15},
		{utils.PageElement{Side: "right", First: false, Index: 8, Blank: false, Name: "name"}, 5, 10, 2, 15},
	} {
		testPageStyle(t, data)
	}
}

func TestImages1(t *testing.T) {
	cp := tu.CaptureLogs()

	result := parseAndBuild(t, `
          <p><img src=pattern.png
            /><img alt="No src"
            /><img src=inexistent.jpg alt="Inexistent src" /></p>`)
	logs := cp.Logs()
	if L := len(logs); L != 1 {
		t.Fatalf("expected one log, got %d", L)
	}
	if !strings.Contains(logs[0], "Failed to load image") {
		t.Fatal(logs[0])
	}
	if !strings.Contains(logs[0], "inexistent.jpg") {
		t.Fatal(logs[0])
	}
	assertTree(t, result, []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{
			{"p", LineBoxT, BC{C: []SerBox{
				{"img", InlineReplacedBoxT, BC{Text: "<replaced>"}},
				{"img", InlineBoxT, BC{C: []SerBox{{"img", TextBoxT, BC{Text: "No src"}}}}},
				{"img", InlineBoxT, BC{C: []SerBox{{"img", TextBoxT, BC{Text: "Inexistent src"}}}}},
			}}},
		}}},
	})
}

func TestImages2(t *testing.T) {
	cp := tu.CaptureLogs()

	result := parseAndBuildExt(t, `<p><img src=pattern.png alt="No baseUrl">`, "")
	logs := cp.Logs()
	if L := len(logs); L != 1 {
		t.Fatalf("expected one log, got %d", L)
	}
	if !strings.Contains(logs[0], "Relative URI reference without a base URI") {
		t.Fatal(logs[0])
	}
	assertTree(t, result, []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{
			{"p", LineBoxT, BC{C: []SerBox{{"img", InlineBoxT, BC{C: []SerBox{{"img", TextBoxT, BC{Text: "No baseUrl"}}}}}}}},
		}}},
	})
}

func TestTables1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rules in http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes

	// Rule 1.3
	// Also table model: http://www.w3.org/TR/CSS21/tables.html#model
	assertTree(t, parseAndBuild(t, `
      <x-table>
        <x-tr>
          <x-th>foo</x-th>
          <x-th>bar</x-th>
        </x-tr>
        <x-tfoot></x-tfoot>
        <x-thead><x-th></x-th></x-thead>
        <x-caption style="caption-side: bottom"></x-caption>
        <x-thead></x-thead>
        <x-col></x-col>
        <x-caption>top caption</x-caption>
        <x-tr>
          <x-td>baz</x-td>
        </x-tr>
      </x-table>
    `), []SerBox{
		{"x-table", BlockBoxT, BC{C: []SerBox{
			{"x-caption", TableCaptionBoxT, BC{C: []SerBox{{"x-caption", LineBoxT, BC{C: []SerBox{{"x-caption", TextBoxT, BC{Text: "top caption"}}}}}}}},
			{"x-table", TableBoxT, BC{C: []SerBox{
				{"x-table", TableColumnGroupBoxT, BC{C: []SerBox{{"x-col", TableColumnBoxT, BC{C: []SerBox{}}}}}},
				{"x-thead", TableRowGroupBoxT, BC{C: []SerBox{{"x-thead", TableRowBoxT, BC{C: []SerBox{{"x-th", TableCellBoxT, BC{C: []SerBox{}}}}}}}}},
				{"x-table", TableRowGroupBoxT, BC{C: []SerBox{
					{"x-tr", TableRowBoxT, BC{C: []SerBox{
						{"x-th", TableCellBoxT, BC{C: []SerBox{{"x-th", LineBoxT, BC{C: []SerBox{{"x-th", TextBoxT, BC{Text: "foo"}}}}}}}},
						{"x-th", TableCellBoxT, BC{C: []SerBox{{"x-th", LineBoxT, BC{C: []SerBox{{"x-th", TextBoxT, BC{Text: "bar"}}}}}}}},
					}}},
				}}},
				{"x-thead", TableRowGroupBoxT, BC{C: []SerBox{}}},
				{"x-table", TableRowGroupBoxT, BC{C: []SerBox{
					{"x-tr", TableRowBoxT, BC{C: []SerBox{
						{"x-td", TableCellBoxT, BC{C: []SerBox{{"x-td", LineBoxT, BC{C: []SerBox{{"x-td", TextBoxT, BC{Text: "baz"}}}}}}}},
					}}},
				}}},
				{"x-tfoot", TableRowGroupBoxT, BC{C: []SerBox{}}},
			}}},
			{"x-caption", TableCaptionBoxT, BC{C: []SerBox{}}},
		}}},
	})
}

func TestTables2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rules 1.4 && 3.1
	assertTree(t, parseAndBuild(t, `
      <span style="display: table-cell">foo</span>
      <span style="display: table-cell">bar</span>
   `), []SerBox{
		{"body", BlockBoxT, BC{C: []SerBox{
			{"body", TableBoxT, BC{C: []SerBox{
				{"body", TableRowGroupBoxT, BC{C: []SerBox{
					{"body", TableRowBoxT, BC{C: []SerBox{
						{"span", TableCellBoxT, BC{C: []SerBox{{"span", LineBoxT, BC{C: []SerBox{{"span", TextBoxT, BC{Text: "foo"}}}}}}}},
						{"span", TableCellBoxT, BC{C: []SerBox{{"span", LineBoxT, BC{C: []SerBox{{"span", TextBoxT, BC{Text: "bar"}}}}}}}},
					}}},
				}}},
			}}},
		}}},
	})
}

func TestTables3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes
	// Rules 1.1 && 1.2
	// Rule XXX (not := range the spec): column groups have at least one column child
	assertTree(t, parseAndBuild(t, `
      <span style="display: table-column-group">
        1
        <em style="display: table-column">
          2
          <strong>3</strong>
        </em>
        <strong>4</strong>
      </span>
      <ins style="display: table-column-group"></ins>
    `), []SerBox{
		{"body", BlockBoxT, BC{C: []SerBox{
			{"body", TableBoxT, BC{C: []SerBox{
				{"span", TableColumnGroupBoxT, BC{C: []SerBox{{"em", TableColumnBoxT, BC{C: []SerBox{}}}}}},
				{"ins", TableColumnGroupBoxT, BC{C: []SerBox{{"ins", TableColumnBoxT, BC{C: []SerBox{}}}}}},
			}}},
		}}},
	})
}

func TestTables4(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rules 2.1 then 2.3
	assertTree(t, parseAndBuild(t, "<x-table>foo <div></div></x-table>"), []SerBox{
		{"x-table", BlockBoxT, BC{C: []SerBox{
			{"x-table", TableBoxT, BC{C: []SerBox{
				{"x-table", TableRowGroupBoxT, BC{C: []SerBox{
					{"x-table", TableRowBoxT, BC{C: []SerBox{
						{"x-table", TableCellBoxT, BC{C: []SerBox{
							{"x-table", BlockBoxT, BC{C: []SerBox{{"x-table", LineBoxT, BC{C: []SerBox{{"x-table", TextBoxT, BC{Text: "foo "}}}}}}}},
							{"div", BlockBoxT, BC{C: []SerBox{}}},
						}}},
					}}},
				}}},
			}}},
		}}},
	})
}

func TestTables5(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rule 2.2
	assertTree(t, parseAndBuild(t, `<x-thead style="display: table-header-group"><div></div><x-td></x-td></x-thead>`),
		[]SerBox{
			{"body", BlockBoxT, BC{C: []SerBox{
				{"body", TableBoxT, BC{C: []SerBox{
					{"x-thead", TableRowGroupBoxT, BC{C: []SerBox{
						{"x-thead", TableRowBoxT, BC{C: []SerBox{
							{"x-thead", TableCellBoxT, BC{C: []SerBox{{"div", BlockBoxT, BC{C: []SerBox{}}}}}},
							{"x-td", TableCellBoxT, BC{C: []SerBox{}}},
						}}},
					}}},
				}}},
			}}},
		})
}

func TestTables6(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rule 3.2
	assertTree(t, parseAndBuild(t, "<span><x-tr></x-tr></span>"), []SerBox{
		{"body", LineBoxT, BC{C: []SerBox{
			{"span", InlineBoxT, BC{C: []SerBox{
				{"span", InlineBlockBoxT, BC{C: []SerBox{
					{"span", InlineTableBoxT, BC{C: []SerBox{{"span", TableRowGroupBoxT, BC{C: []SerBox{{"x-tr", TableRowBoxT, BC{C: []SerBox{}}}}}}}}},
				}}},
			}}},
		}}},
	})
}

func TestTables7(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rule 3.1
	// Also, rule 1.3 does ! apply: whitespace before && after is preserved
	assertTree(t, parseAndBuild(t, `
		<span>
		  <em style="display: table-cell"></em>
		  <em style="display: table-cell"></em>
		</span>
	  `), []SerBox{
		{"body", LineBoxT, BC{C: []SerBox{
			{"span", InlineBoxT, BC{C: []SerBox{
				{"span", TextBoxT, BC{Text: " "}},
				{"span", InlineBlockBoxT, BC{C: []SerBox{
					{"span", InlineTableBoxT, BC{C: []SerBox{
						{"span", TableRowGroupBoxT, BC{C: []SerBox{
							{"span", TableRowBoxT, BC{C: []SerBox{
								{"em", TableCellBoxT, BC{C: []SerBox{}}},
								{"em", TableCellBoxT, BC{C: []SerBox{}}},
							}}},
						}}},
					}}},
				}}},
				{"span", TextBoxT, BC{Text: " "}},
			}}},
		}}},
	})
}

func TestTables8(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Rule 3.2
	assertTree(t, parseAndBuild(t, "<x-tr></x-tr>\t<x-tr></x-tr>"), []SerBox{
		{"body", BlockBoxT, BC{C: []SerBox{
			{"body", TableBoxT, BC{C: []SerBox{
				{"body", TableRowGroupBoxT, BC{C: []SerBox{
					{"x-tr", TableRowBoxT, BC{C: []SerBox{}}},
					{"x-tr", TableRowBoxT, BC{C: []SerBox{}}},
				}}},
			}}},
		}}},
	})
}

func TestTables9(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, "<x-col></x-col>\n<x-colgroup></x-colgroup>"), []SerBox{
		{"body", BlockBoxT, BC{C: []SerBox{
			{"body", TableBoxT, BC{C: []SerBox{
				{"body", TableColumnGroupBoxT, BC{C: []SerBox{{"x-col", TableColumnBoxT, BC{C: []SerBox{}}}}}},
				{"x-colgroup", TableColumnGroupBoxT, BC{C: []SerBox{{"x-colgroup", TableColumnBoxT, BC{C: []SerBox{}}}}}},
			}}},
		}}},
	})
}

func TestTableStyle(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	html := parseAndBuild(t, `<table style="margin: 1px; padding: 2px"></table>`)
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0]
	if !(BlockBoxT.IsInstance(wrapper)) {
		t.Fatal()
	}
	if !(TableBoxT.IsInstance(table)) {
		t.Fatal()
	}
	if !(wrapper.Box().Style.GetMarginTop() == pr.FToPx(1)) {
		t.Fatal()
	}
	if !(wrapper.Box().Style.GetPaddingTop() == pr.FToPx(0)) {
		t.Fatal()
	}
	if !(table.Box().Style.GetMarginTop() == pr.FToPx(0)) {
		t.Fatal()
	}
	if !(table.Box().Style.GetPaddingTop() == pr.FToPx(2)) {
		t.Fatal()
	}
}

func TestColumnStyle(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	html := parseAndBuild(t, `
      <table>
        <col span=3 style="width: 10px"></col>
        <col span=2></col>
      </table>
    `)
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0].(TableBoxITF)
	colgroup := table.Table().ColumnGroups[0]
	var (
		widths []pr.Value
		gridXs []int
	)
	for _, col := range colgroup.Box().Children {
		widths = append(widths, col.Box().Style.GetWidth())
		gridXs = append(gridXs, col.Box().GridX)
	}
	if !reflect.DeepEqual(widths, []pr.Value{
		pr.FToPx(10), pr.FToPx(10), pr.FToPx(10), pr.SToV("auto"), pr.SToV("auto"),
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(gridXs, []int{0, 1, 2, 3, 4}) {
		t.Fatal()
	}
	// copies, not the same box object
	if colgroup.Box().Children[0] == colgroup.Box().Children[1] {
		t.Fatal()
	}
}

func TestNestedGridX(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	html := parseAndBuild(t, `
      <table>
        <col span=2></col>
        <colgroup span=2></colgroup>
        <colgroup>
          <col></col>
          <col span=2></col>
        </colgroup>
        <col></col>
      </table>
    `)
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0].(TableBoxITF)

	type gridX struct {
		v    int
		cols []int
	}
	var gridXs []gridX
	for _, colgroup := range table.Table().ColumnGroups {
		v := gridX{v: colgroup.Box().GridX}
		for _, col := range colgroup.Box().Children {
			v.cols = append(v.cols, col.Box().GridX)
		}
		gridXs = append(gridXs, v)
	}
	if !reflect.DeepEqual(gridXs, []gridX{
		{0, []int{0, 1}}, {2, []int{2, 3}}, {4, []int{4, 5, 6}}, {7, []int{7}},
	}) {
		t.Fatal()
	}
}

func extractSpans(group Box) (gridXs, colspans, rowspans [][]int) {
	for _, row := range group.Box().Children {
		var gridX, colspan, rowspan []int
		for _, c := range row.Box().Children {
			gridX = append(gridX, c.Box().GridX)
			colspan = append(colspan, c.Box().Colspan)
			rowspan = append(rowspan, c.Box().Rowspan)
		}
		gridXs = append(gridXs, gridX)
		colspans = append(colspans, colspan)
		rowspans = append(rowspans, rowspan)
	}
	return
}

func TestColspanRowspan1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// +---+---+---+
	// | A | B | C | X
	// +---+---+---+
	// | D |     E | X
	// +---+---+   +---+
	// |  F ...|   |   |   <-- overlap
	// +---+---+---+   +
	// | H | X   X | G |
	// +---+---+   +   +
	// | I | J | X |   |
	// +---+---+   +---+

	// X: empty cells
	html := parseAndBuild(t, `
      <table>
        <tr>
          <td>A <td>B <td>C
        </tr>
        <tr>
          <td>D <td colspan=2 rowspan=2>E
        </tr>
        <tr>
          <td colspan=2>F <td rowspan=0>G
        </tr>
        <tr>
          <td>H
        </tr>
        <tr>
          <td>I <td>J
        </tr>
      </table>
    `)
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0].(TableBoxITF)
	group := table.Box().Children[0]

	gridXs, colspans, rowspans := extractSpans(group)

	if !reflect.DeepEqual(gridXs, [][]int{
		{0, 1, 2},
		{0, 1},
		{0, 3},
		{0},
		{0, 1},
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(colspans, [][]int{
		{1, 1, 1},
		{1, 2},
		{2, 1},
		{1},
		{1, 1},
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(rowspans, [][]int{
		{1, 1, 1},
		{1, 2},
		{1, 3},
		{1},
		{1, 1},
	}) {
		t.Fatal()
	}
}

func TestColspanRowspan2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// A cell box cannot extend beyond the last row box of a table.
	html := parseAndBuild(t, `
        <table>
            <tr>
                <td rowspan=5></td>
                <td></td>
            </tr>
            <tr>
                <td></td>
            </tr>
        </table>
    `)
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0].(TableBoxITF)
	group := table.Box().Children[0]

	gridXs, colspans, rowspans := extractSpans(group)

	if !reflect.DeepEqual(gridXs, [][]int{
		{0, 1},
		{1},
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(colspans, [][]int{
		{1, 1},
		{1},
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(rowspans, [][]int{
		{2, 1},
		{1},
	}) {
		t.Fatal()
	}
}

func TestBeforeAfter1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        p:before { content: normal }
        div:before { content: none }
        section::before { color: black }
      </style>
      <p></p>
      <div></div>
      <section></section>
    `), []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{}}},
		{"div", BlockBoxT, BC{C: []SerBox{}}},
		{"section", BlockBoxT, BC{C: []SerBox{}}},
	})
}

func TestBeforeAfter2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        p:before { content: "a" "b" }
        p::after { content: "d" "e" }
      </style>
      <p> c </p>
    `), []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{
			{"p", LineBoxT, BC{C: []SerBox{
				{"p::before", InlineBoxT, BC{C: []SerBox{{"p::before", TextBoxT, BC{Text: "ab"}}}}},
				{"p", TextBoxT, BC{Text: " c "}},
				{"p::after", InlineBoxT, BC{C: []SerBox{{"p::after", TextBoxT, BC{Text: "de"}}}}},
			}}},
		}}},
	})
}

func TestBeforeAfter3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	assertTree(t, parseAndBuild(t, `
      <style>
        a[href]:before { content: "[" attr(href) "] " }
      </style>
      <p><a href="some url">some text</a></p>
    `), []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{
			{"p", LineBoxT, BC{C: []SerBox{
				{"a", InlineBoxT, BC{C: []SerBox{
					{"a::before", InlineBoxT, BC{C: []SerBox{{"a::before", TextBoxT, BC{Text: "[some url] "}}}}},
					{"a", TextBoxT, BC{Text: "some text"}},
				}}},
			}}},
		}}},
	})
}

func TestBeforeAfter4(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
	<style>
		body { quotes: '«' '»' '“' '”' }
		q:before { content: open-quote ' '}
		q:after { content: ' ' close-quote }
	</style>
  	<p><q>Lorem ipsum <q>dolor</q> sit amet</q></p>
    `), []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{
			{"p", LineBoxT, BC{C: []SerBox{
				{"q", InlineBoxT, BC{C: []SerBox{
					{"q::before", InlineBoxT, BC{C: []SerBox{{"q::before", TextBoxT, BC{Text: "« "}}}}},
					{"q", TextBoxT, BC{Text: "Lorem ipsum "}},
					{"q", InlineBoxT, BC{C: []SerBox{
						{"q::before", InlineBoxT, BC{C: []SerBox{{"q::before", TextBoxT, BC{Text: "“ "}}}}},
						{"q", TextBoxT, BC{Text: "dolor"}},
						{"q::after", InlineBoxT, BC{C: []SerBox{{"q::after", TextBoxT, BC{Text: " ”"}}}}},
					}}},
					{"q", TextBoxT, BC{Text: " sit amet"}},
					{"q::after", InlineBoxT, BC{C: []SerBox{{"q::after", TextBoxT, BC{Text: " »"}}}}},
				}}},
			}}},
		}}},
	})
}

func TestBeforeAfter5(t *testing.T) {
	cp := tu.CaptureLogs()

	assertTree(t, parseAndBuild(t, `
          <style>
            p:before {
              content: "a" url(pattern.png) "b";

              /* Invalid, ignored in favor of the one above.
                 Regression test: this used to crash: */
              content: some-function(nested-function(something));
            }
          </style>
          <p>c</p>
        `), []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{
			{"p", LineBoxT, BC{C: []SerBox{
				{"p::before", InlineBoxT, BC{C: []SerBox{
					{"p::before", TextBoxT, BC{Text: "a"}},
					{"p::before", InlineReplacedBoxT, BC{Text: "<replaced>"}},
					{"p::before", TextBoxT, BC{Text: "b"}},
				}}},
				{"p", TextBoxT, BC{Text: "c"}},
			}}},
		}}},
	})

	logs := cp.Logs()
	if L := len(logs); L != 1 {
		t.Fatalf("expected 1 log, got %d", L)
	}
	if !strings.Contains(logs[0], "nested-function(") {
		t.Fatalf("unexpected log: %s", logs[0])
	}
	if !strings.Contains(logs[0], "invalid value") {
		t.Fatalf("unexpected log: %s", logs[0])
	}
}

var (
	black       = pr.NewColor(0, 0, 0, 1)
	red         = pr.NewColor(1, 0, 0, 1)
	green       = pr.NewColor(0, 1, 0, 1) // lime in CSS
	blue        = pr.NewColor(0, 0, 1, 1)
	yellow      = pr.NewColor(1, 1, 0, 1)
	black3      = Border{Style: "solid", Width: 3, Color: black}
	red1        = Border{Style: "solid", Width: 1, Color: red}
	yellow5     = Border{Style: "solid", Width: 5, Color: yellow}
	green5      = Border{Style: "solid", Width: 5, Color: green}
	dashedBlue5 = Border{Style: "dashed", Width: 5, Color: blue}
)

func TestBorderCollapse1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	html := parseAndBuild(t, "<table></table>")

	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0].(TableBoxITF)

	if !(table.Table().CollapsedBorderGrid.Horizontal == nil && table.Table().CollapsedBorderGrid.Vertical == nil) {
		t.Fatal()
	}

	gridH, gridV := getGrid(t, `<table style="border-collapse: collapse"></table>`)

	if len(gridH) != 0 || len(gridV) != 0 {
		t.Fatal()
	}
}

func TestBorderCollapse2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	verticalBorders, horizontalBorders := getGrid(t, `
      <style>td { border: 1px solid red }</style>
      <table style="border-collapse: collapse; border: 3px solid black">
        <tr> <td>A</td> <td>B</td> </tr>
        <tr> <td>C</td> <td>D</td> </tr>
      </table>
    `)
	if !reflect.DeepEqual(verticalBorders, [][]Border{
		{black3, red1, black3},
		{black3, red1, black3},
	}) {
		t.Fatalf("unexepected vertical borders %v", verticalBorders)
	}
	if !reflect.DeepEqual(horizontalBorders, [][]Border{
		{black3, black3},
		{red1, red1},
		{black3, black3},
	}) {
		t.Fatalf("unexepected horizontal borders: %v", horizontalBorders)
	}
}

func TestBorderCollapse3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// hidden vs. none
	verticalBorders, horizontalBorders := getGrid(t, `
      <style>table, td { border: 3px solid }</style>
      <table style="border-collapse: collapse">
        <tr> <td>A</td> <td style="border-style: hidden">B</td> </tr>
        <tr> <td>C</td> <td style="border-style: none">D</td> </tr>
      </table>
    `)
	if !reflect.DeepEqual(verticalBorders, [][]Border{
		{black3, Border{}, Border{}},
		{black3, black3, black3},
	}) {
		t.Fatalf("unexepected vertical borders %v", verticalBorders)
	}
	if !reflect.DeepEqual(horizontalBorders, [][]Border{
		{black3, Border{}},
		{black3, Border{}},
		{black3, black3},
	}) {
		t.Fatalf("unexepected horizontal borders: %v", horizontalBorders)
	}
}

func TestBorderCollapse4(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	verticalBorders, horizontalBorders := getGrid(t, `
      <style>td { border: 1px solid red }</style>
      <table style="border-collapse: collapse; border: 5px solid yellow">
        <col style="border: 3px solid black" />
        <tr> <td></td> <td></td> <td></td> </tr>
        <tr> <td></td> <td style="border: 5px dashed blue"></td>
          <td style="border: 5px solid lime"></td> </tr>
        <tr> <td></td> <td></td> <td></td> </tr>
        <tr> <td></td> <td></td> <td></td> </tr>
      </table>
    `)

	if !reflect.DeepEqual(verticalBorders, [][]Border{
		{yellow5, black3, red1, yellow5},
		{yellow5, dashedBlue5, green5, green5},
		{yellow5, black3, red1, yellow5},
		{yellow5, black3, red1, yellow5},
	}) {
		t.Fatalf("unexepected vertical borders %v", verticalBorders)
	}
	if !reflect.DeepEqual(horizontalBorders, [][]Border{
		{yellow5, yellow5, yellow5},
		{red1, dashedBlue5, green5},
		{red1, dashedBlue5, green5},
		{red1, red1, red1},
		{yellow5, yellow5, yellow5},
	}) {
		t.Fatalf("unexepected horizontal borders: %v", horizontalBorders)
	}
}

func TestBorderCollapse5(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// rowspan && colspan
	verticalBorders, horizontalBorders := getGrid(t, `
        <style>col, tr { border: 3px solid }</style>
        <table style="border-collapse: collapse">
            <col /><col /><col />
            <tr> <td rowspan=2></td> <td></td> <td></td> </tr>
            <tr>                     <td colspan=2></td> </tr>
        </table>
    `)

	if !reflect.DeepEqual(verticalBorders, [][]Border{
		{black3, black3, black3, black3},
		{black3, black3, Border{}, black3},
	}) {
		t.Fatalf("unexepected vertical borders %v", verticalBorders)
	}
	if !reflect.DeepEqual(horizontalBorders, [][]Border{
		{black3, black3, black3},
		{Border{}, black3, black3},
		{black3, black3, black3},
	}) {
		t.Fatalf("unexepected horizontal borders: %v", horizontalBorders)
	}
}

func testDisplayNoneRoot(t *testing.T, html string) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	box := parseAndBuild(t, html)
	if d := box.Box().Style.GetDisplay(); d != (pr.Display{"block", "flow"}) {
		t.Fatalf("unexpected display: %s", d)
	}
	if len(box.Box().Children) != 0 {
		t.Fatal()
	}
}

func TestDisplayNoneRoot(t *testing.T) {
	for _, html := range []string{
		`<html style="display: none">`,
		`<html style="display: none">abc`,
		`<html style="display: none"><p>abc`,
		`<body style="display: none"><p>abc`,
	} {
		testDisplayNoneRoot(t, html)
	}
}

func TestBuildPages(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
	<style>
		@page {
		/* Make the page content area only 10px high and wide,
			so every word in <p> end up on a page of its own. */
		size: 30px;
		margin: 10px;
		@top-center { content: "Title" }
		}
		@page :first {
		@bottom-left { content: "foo" }
		@bottom-left-corner { content: "baz" }
		}
	</style>
	<p>lorem ipsum
	`), []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{{"p", LineBoxT, BC{C: []SerBox{{"p", TextBoxT, BC{Text: "lorem ipsum "}}}}}}}},
	})
}

func TestInlineSpace(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
	<p>start <i><b>bi1</b> <b>bi2</b></i> <b>b1</b> end</p>
	`), []SerBox{
		{"p", BlockBoxT, BC{C: []SerBox{
			{"p", LineBoxT, BC{C: []SerBox{
				{"p", TextBoxT, BC{Text: "start "}},
				{"i", InlineBoxT, BC{C: []SerBox{
					{"b", InlineBoxT, BC{C: []SerBox{{"b", TextBoxT, BC{Text: "bi1"}}}}},
					{"i", TextBoxT, BC{Text: " "}},
					{"b", InlineBoxT, BC{C: []SerBox{{"b", TextBoxT, BC{Text: "bi2"}}}}},
				}}},
				{"p", TextBoxT, BC{Text: " "}},
				{"b", InlineBoxT, BC{C: []SerBox{{"b", TextBoxT, BC{Text: "b1"}}}}},
				{"p", TextBoxT, BC{Text: " end"}},
			}}},
		}}},
	})
}

func TestPhEmbedded(t *testing.T) {
	assertTree(t, parseAndBuild(t, `
	<object data="data:image/svg+xml,<svg></svg>"
			align=top hspace=10 vspace=20></object>
	<img src="data:image/svg+xml,<svg></svg>" alt=text
			align=right width=10 height=20 />
	<embed src="data:image/svg+xml,<svg></svg>" align=texttop />
  `), []SerBox{
		{"body", LineBoxT, BC{C: []SerBox{
			{"object", InlineReplacedBoxT, BC{Text: "<replaced>"}},
			{"body", TextBoxT, BC{Text: " "}},
			{"img", InlineReplacedBoxT, BC{Text: "<replaced>"}},
			{"body", TextBoxT, BC{Text: " "}},
			{"embed", InlineReplacedBoxT, BC{Text: "<replaced>"}},
			{"body", TextBoxT, BC{Text: " "}},
		}}},
	})
}

func TestNoNewLine(t *testing.T) {
	assertTree(t, parseAndBuild(t, `
	<style>
	@page { size: 300px 30px }
	body { margin: 0; background: #fff }
</style>
<p><a href="another url"><span>[some url] </span>some content</p>
`), []SerBox{
		{Tag: "p", Type: 0x2, Content: BC{
			C: []SerBox{{Tag: "p", Type: 0xf, Content: BC{C: []SerBox{
				{Tag: "a", Type: 0xa, Content: BC{C: []SerBox{
					{Tag: "span", Type: 0xa, Content: BC{C: []SerBox{
						{Tag: "span", Type: 0x1b, Content: BC{Text: "[some url] "}},
					}}},
					{Tag: "a", Type: 0x1b, Content: BC{Text: "some content"}},
				}}},
			}}}},
		}},
		{Tag: "body", Type: 0x2, Content: BC{C: []SerBox{
			{Tag: "body", Type: 0xf, Content: BC{C: []SerBox{
				{Tag: "a", Type: 0xa, Content: BC{C: []SerBox{
					{Tag: "a", Type: 0x1b, Content: BC{Text: " "}},
				}}},
			}}},
		}}},
	})
}

func buildFile(t testing.TB, source utils.ContentInput, baseURL string) []SerBox {
	var box Box = BuildFormattingStructure(parseBase(t, source, baseURL))
	if err := sanityChecks(box); err != nil {
		t.Fatalf("sanity check failed: %s", err)
	}

	if tag := box.Box().ElementTag(); tag != "html" {
		t.Fatalf("unexpected element: %s", tag)
	}
	if !BlockBoxT.IsInstance(box) {
		t.Fatal("expected block box")
	}
	if L := len(box.Box().Children); L != 1 {
		t.Fatalf("expected one children, got %d", L)
	}

	box = box.Box().Children[0]
	if !BlockBoxT.IsInstance(box) {
		t.Fatal("expected block box")
	}
	if tag := box.Box().ElementTag(); tag != "body" {
		t.Fatalf("unexpected element: %s", tag)
	}

	return Serialize(box.Box().Children)
}

func loadExpected(filename string) ([]SerBox, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	var out []SerBox
	err = json.NewDecoder(f).Decode(&out)
	return out, err
}

func TestRealPage(t *testing.T) {
	// Requires network to load the linked ressources

	// log.Default().SetOutput(io.Discard)
	got := buildFile(t, utils.InputFilename("../../resources_test/Wikipedia-Go.html"), "https://en.wikipedia.org/wiki/Go_(programming_language)")

	expected, err := loadExpected("../../resources_test/Wikipedia-Go-expected.json")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, got) {
		// ioutil.WriteFile("expected.tmp", []byte(fmt.Sprintf("%v", expected)), os.ModePerm)
		// ioutil.WriteFile("got.tmp", []byte(fmt.Sprintf("%v", got)), os.ModePerm)
		t.Fatal("diff")
	}
}

func BenchmarkRealPage(b *testing.B) {
	log.Default().SetOutput(io.Discard)

	for i := 0; i < b.N; i++ {
		buildFile(b, utils.InputFilename("../resources_test/Wikipedia-Go.html"), "https://en.wikipedia.org/wiki/Go_(programming_language)")
	}
}
