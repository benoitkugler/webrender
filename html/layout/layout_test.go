package layout

import (
	"fmt"
	"io"
	"testing"

	fc "github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/textprocessing/pango/fcfonts"
	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/utils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

var baseUrl, _ = utils.PathToURL("../../resources_test/")

const fontmapCache = "../../text/testdata/cache.fc"

var fontconfig *text.FontConfigurationPango

func init() {
	logger.ProgressLogger.SetOutput(io.Discard)

	// this command has to run once
	// fmt.Println("Scanning fonts...")
	// _, err := fc.ScanAndCache(fontmapCache)
	// if err != nil {
	// 	panic(err)
	// }

	fs, err := fc.LoadFontsetFile(fontmapCache)
	if err != nil {
		panic(err)
	}
	fontconfig = text.NewFontConfigurationPango(fcfonts.NewFontMap(fc.Standard.Copy(), fs))
}

func fakeHTML(html *tree.HTML) *tree.HTML {
	html.UAStyleSheet = tree.TestUAStylesheet
	return html
}

// lay out a document and return a list of PageBox objects
func renderPages(t *testing.T, htmlContent string, css ...tree.CSS) []*bo.PageBox {
	doc, err := tree.NewHTML(utils.InputString(htmlContent), baseUrl, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	doc = fakeHTML(doc)
	return Layout(doc, css, false, fontconfig)
}

// same as renderPages, but expects only on laid out page
func renderOnePage(t *testing.T, htmlContent string) *bo.PageBox {
	pages := renderPages(t, htmlContent)
	if len(pages) != 1 {
		t.Fatalf("expected one page, got %v", pages)
	}
	return pages[0]
}

func renderTwoPages(t *testing.T, htmlContent string) (page1, page2 *bo.PageBox) {
	pages := renderPages(t, htmlContent)
	if len(pages) != 2 {
		t.Fatalf("expected two pages, got %v", pages)
	}
	return pages[0], pages[1]
}

// unpack 1 children
func unpack1(box Box) (c1 Box) {
	return box.Box().Children[0]
}

// unpack 2 children
func unpack2(box Box) (c1, c2 Box) {
	return unpack1(box), box.Box().Children[1]
}

// unpack 3 children
func unpack3(box Box) (c1, c2, c3 Box) {
	return unpack1(box), box.Box().Children[1], box.Box().Children[2]
}

// unpack 4 children
func unpack4(box Box) (c1, c2, c3, c4 Box) {
	return unpack1(box), box.Box().Children[1], box.Box().Children[2], box.Box().Children[3]
}

// unpack 5 children
func unpack5(box Box) (c1, c2, c3, c4, c5 Box) {
	return unpack1(box), box.Box().Children[1], box.Box().Children[2], box.Box().Children[3], box.Box().Children[4]
}

// unpack 6 children
func unpack6(box Box) (c1, c2, c3, c4, c5, c6 Box) {
	return unpack1(box), box.Box().Children[1], box.Box().Children[2], box.Box().Children[3], box.Box().Children[4], box.Box().Children[5]
}

// unpack 7 children
func unpack7(box Box) (c1, c2, c3, c4, c5, c6, c7 Box) {
	return unpack1(box), box.Box().Children[1], box.Box().Children[2], box.Box().Children[3], box.Box().Children[4], box.Box().Children[5], box.Box().Children[6]
}

// unpack 8 children
func unpack8(box Box) (c1, c2, c3, c4, c5, c6, c7, c8 Box) {
	return unpack1(box), box.Box().Children[1], box.Box().Children[2], box.Box().Children[3], box.Box().Children[4], box.Box().Children[5], box.Box().Children[6], box.Box().Children[7]
}

func asBoxes(pages []*bo.PageBox) []Box {
	out := make([]Box, len(pages))
	for i, p := range pages {
		out[i] = p
	}
	return out
}

type nodePos []int

func (np nodePos) isLess(other nodePos) bool {
	for i := 0; i < len(np) && i < len(other); i++ {
		if np[i] < other[i] {
			return true
		} else if np[i] > other[i] {
			return false
		} else {
			continue
		}
	}
	return len(np) < len(other)
}

type nodePosList []nodePos

func (a nodePosList) Len() int           { return len(a) }
func (a nodePosList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a nodePosList) Less(i, j int) bool { return a[i].isLess(a[j]) }

// Return a list identifying the first matching box's tree position.
//
// Given a list of Boxes, this function returns a list containing the first
// (depth-first) Box that the matcher function identifies. This list can then
// be compared to another similarly-obtained list to assert that one Box is in
// the document tree before or after another.
func treePosition(boxList []Box, matcher func(Box) bool) nodePos {
	for i, box := range boxList {
		if matcher(box) {
			return []int{i}
		} else {
			position := treePosition(box.Box().Children, matcher)
			if len(position) != 0 {
				return append([]int{i}, position...)
			}
		}
	}
	return nil
}

func TestMarginBoxes(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
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
	  `)
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %v", pages)
	}
	page1, page2 := pages[0], pages[1]
	tu.AssertEqual(t, page1.Children[0].Box().ElementTag(), "html")
	tu.AssertEqual(t, page2.Children[0].Box().ElementTag(), "html")

	var marginBoxes1, marginBoxes2 []string
	for _, box := range page1.Children[1:] {
		marginBoxes1 = append(marginBoxes1, box.(*bo.MarginBox).AtKeyword)
	}
	for _, box := range page2.Children[1:] {
		marginBoxes2 = append(marginBoxes2, box.(*bo.MarginBox).AtKeyword)
	}
	tu.AssertEqual(t, marginBoxes1, []string{"@top-center", "@bottom-left", "@bottom-left-corner"})
	tu.AssertEqual(t, marginBoxes2, []string{"@top-center"})

	if len(page2.Children) != 2 {
		t.Fatalf("expected two children, got %v", page2.Children)
	}
	_, topCenter := page2.Children[0], page2.Children[1]
	lineBox := unpack1(topCenter)
	textBox, _ := unpack1(lineBox).(*bo.TextBox)
	assertText(t, textBox, "Title")
}

func TestMarginBoxStringSet1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test that both pages get string in the `bottom-center` margin box
	pages := renderPages(t, `
      <style>
        @page {
          @bottom-center { content: string(text_header) }
        }
        p {
          string-set: text_header content();
        }
        .page {
          page-break-before: always;
        }
      </style>
      <p>first assignment</p>
      <div class="page"></div>
    `)
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %v", pages)
	}
	page1, page2 := pages[0], pages[1]

	if len(page2.Children) != 2 {
		t.Fatalf("expected two children, got %v", page2.Children)
	}
	_, bottomCenter := page2.Children[0], page2.Children[1]
	lineBox := unpack1(bottomCenter)
	textBox, _ := unpack1(lineBox).(*bo.TextBox)
	assertText(t, textBox, "first assignment")

	if len(page1.Children) != 2 {
		t.Fatalf("expected two children, got %v", page1.Children)
	}
	_, bottomCenter = page1.Children[0], page1.Children[1]

	lineBox = unpack1(bottomCenter)
	textBox, _ = unpack1(lineBox).(*bo.TextBox)
	assertText(t, textBox, "first assignment")
}

func TestMarginBoxStringSet2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	simpleStringSetTest := func(contentVal, extraStyle string) {
		page1 := renderOnePage(t, fmt.Sprintf(`
          <style>
            @page {
              @top-center { content: string(text_header) }
            }
            p {
              string-set: text_header content(%s);
            }
            %s
          </style>
          <p>first assignment</p>
        `, contentVal, extraStyle))

		topCenter := page1.Children[1]
		lineBox := unpack1(topCenter)
		textBox := unpack1(lineBox).(*bo.TextBox)
		if contentVal == "before" || contentVal == "after" {
			assertText(t, textBox, "pseudo")
		} else {
			assertText(t, textBox, "first assignment")
		}
	}

	// Test each accepted value of `content()` as an arguemnt to `string-set`
	for _, value := range []string{"", "text", "before", "after"} {
		var extraStyle string
		if value == "before" || value == "after" {
			extraStyle = fmt.Sprintf("p:%s{content: 'pseudo'}", value)
		}
		simpleStringSetTest(value, extraStyle)
	}
}

func TestMarginBoxStringSet3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Test `first` (default value) ie. use the first assignment on the page
	page1 := renderOnePage(t, `
      <style>
        @page {
          @top-center { content: string(text_header, first) }
        }
        p {
          string-set: text_header content();
        }
      </style>
      <p>first assignment</p>
      <p>Second assignment</p>
    } `)

	topCenter := page1.Children[1]
	lineBox := unpack1(topCenter)
	textBox := unpack1(lineBox).(*bo.TextBox)
	assertText(t, textBox, "first assignment")
}

func TestMarginBoxStringSet4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// test `first-except` ie. exclude from page on which value is assigned
	pages := renderPages(t, `
		<style>
		@page {
		@top-center { content: string(header-nofirst, first-except) }
		}
		p{
		string-set: header-nofirst content();
		}
		.page{
		page-break-before: always;
		}
	</style>
	<p>first_excepted</p>
	<div class="page"></div>
	`)
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %v", pages)
	}
	page1, page2 := pages[0], pages[1]

	topCenter := page1.Box().Children[1]
	tu.AssertEqual(t, len(topCenter.Box().Children), 0)

	topCenter = page2.Box().Children[1]
	lineBox := unpack1(topCenter)
	textBox := unpack1(lineBox).(*bo.TextBox)
	assertText(t, textBox, "first_excepted")
}

func TestMarginBoxStringSet5(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Test `last` ie. use the most-recent assignment
	page1 := renderOnePage(t, `
      <style>
        @page {
          @top-center { content: string(headerLast, last) }
        }
        p {
          string-set: headerLast content();
        }
      </style>
      <p>String set</p>
      <p>Second assignment</p>
    `)

	topCenter := page1.Children[1]
	lineBox := unpack1(topCenter)
	textBox := unpack1(lineBox).(*bo.TextBox)
	assertText(t, textBox, "Second assignment")
}

func TestMarginBoxStringSet6(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test multiple complex string-set values
	page1 := renderOnePage(t, `
		<style>
		@page {
		@top-center { content: string(text_header, first) }
		@bottom-center { content: string(text_footer, last) }
		}
		html { counter-reset: a }
		body { counter-increment: a }
		ul { counter-reset: b }
		li {
		counter-increment: b;
		string-set:
			text_header content(before) "-" content() "-" content(after)
						counter(a, upper-roman) '.' counters(b, '|'),
			text_footer content(before) '-' attr(class)
						counters(b, '|') "/" counter(a, upper-roman);
		}
		li:before { content: 'before!' }
		li:after { content: 'after!' }
		li:last-child:before { content: 'before!last' }
		li:last-child:after { content: 'after!last' }
	</style>
	<ul>
		<li class="firstclass">first
		<li>
		<ul>
			<li class="secondclass">second
    `)

	topCenter, bottomCenter := page1.Children[1], page1.Children[2]
	topLineBox := unpack1(topCenter)
	topTextBox := unpack1(topLineBox).(*bo.TextBox)
	assertText(t, topTextBox, "before!-first-after!I.1")

	bottomLineBox := unpack1(bottomCenter)
	bottomTextBox := unpack1(bottomLineBox).(*bo.TextBox)
	assertText(t, bottomTextBox, "before!last-secondclass2|1/I")
}

func TestMarginBoxStringSet7(t *testing.T) {
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/722
	page1 := renderOnePage(t, `
      <style>
        img { string-set: left attr(alt) }
        img + img { string-set: right attr(alt) }
        @page { @top-left  { content: "[" string(left)  "]" }
                @top-right { content: "{" string(right) "}" } }
      </style>
      <img src=pattern.png alt="Chocolate">
      <img src=noSuchFile.png alt="Cake">
    `)

	topLeft, topRight := page1.Children[1], page1.Children[2]
	leftLineBox := unpack1(topLeft)
	leftTextBox := unpack1(leftLineBox).(*bo.TextBox)
	assertText(t, leftTextBox, "[Chocolate]")

	rightLineBox := unpack1(topRight)
	rightTextBox := unpack1(rightLineBox).(*bo.TextBox)
	assertText(t, rightTextBox, "{Cake}")
}

func TestMarginBoxStringSet8(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/726
	pages := renderPages(t, `
      <style>
        @page { @top-left  { content: "[" string(left) "]" } }
        p { page-break-before: always }
        .initial { string-set: left "initial" }
        .empty   { string-set: left ""        }
        .space   { string-set: left " "       }
      </style>

	  <p class="initial">Initial</p>
      <p class="empty">Empty</p>
      <p class="space">Space</p>
    `)
	if len(pages) != 3 {
		t.Fatalf("expected 3 page, got %v", pages)
	}
	page1, page2, page3 := pages[0], pages[1], pages[2]

	topLeft := page1.Box().Children[1]
	leftLineBox := unpack1(topLeft)
	leftTextBox := unpack1(leftLineBox).(*bo.TextBox)
	assertText(t, leftTextBox, "[initial]")

	topLeft = page2.Box().Children[1]
	leftLineBox = unpack1(topLeft)
	leftTextBox = unpack1(leftLineBox).(*bo.TextBox)
	assertText(t, leftTextBox, "[]")

	topLeft = page3.Box().Children[1]
	leftLineBox = unpack1(topLeft)
	leftTextBox = unpack1(leftLineBox).(*bo.TextBox)
	assertText(t, leftTextBox, "[ ]")
}

func TestMarginBoxStringSet9(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Test that named strings are case-sensitive
	// See https://github.com/Kozea/WeasyPrint/pull/827
	page1 := renderOnePage(t, `
      <style>
        @page {
          @top-center {
            content: string(text_header, first)
                     " " string(TEXTHeader, first)
          }
        }
        p { string-set: text_header content() }
        div { string-set: TEXTHeader content() }
      </style>
      <p>first assignment</p>
      <div>second assignment</div>
    `)

	topCenter := page1.Children[1]
	lineBox := unpack1(topCenter)
	textBox := unpack1(lineBox).(*bo.TextBox)

	assertText(t, textBox, "first assignment second assignment")
}

func TestMarginBoxStringSet10(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
	<style>
		@page { @top-left  { content: '[' string(p, start) ']' } }
		p { string-set: p content(); page-break-after: always }
	</style>
	<article></article>
	<p>1</p>
	<article></article>
	<p>2</p>
	<p>3</p>
	<article></article>
	`)
	if len(pages) != 4 {
		t.Fatalf("expected 4 pages, got %v", pages)
	}
	page1, page2, page3, page4 := pages[0], pages[1], pages[2], pages[3]

	topLeft := page1.Children[1]
	leftLineBox := unpack1(topLeft)
	assertText(t, unpack1(leftLineBox), "[]")

	topLeft = page2.Children[1]
	leftLineBox = unpack1(topLeft)
	assertText(t, unpack1(leftLineBox), "[1]")

	topLeft = page3.Children[1]
	leftLineBox = unpack1(topLeft)
	assertText(t, unpack1(leftLineBox), "[3]")

	topLeft = page4.Children[1]
	leftLineBox = unpack1(topLeft)
	assertText(t, unpack1(leftLineBox), "[3]")
}

// Test page-based counters.
func TestPageCounters(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page {
          /* Make the page content area only 10px high and wide,
             so every word in <p> end up on a page of its own. */
          size: 30px;
          margin: 10px;
          @bottom-center {
            content: "Page " counter(page) " of " counter(pages) ".";
          }
        }
      </style>
      <p>lorem ipsum dolor
    `)
	for pageIndex, page := range pages {
		pageNumber := pageIndex + 1
		bottomCenter := page.Box().Children[1]
		lineBox := unpack1(bottomCenter)
		exp := fmt.Sprintf("Page %d of 3.", pageNumber)
		assertText(t, unpack1(lineBox), exp)
	}
}

func TestBackground(t *testing.T) {
	page := renderOnePage(t, `
	<style>
	   @page { size: 4px; bleed: 1px; margin: 1px; marks: crop }
	</style>
	<body>`)
	layers := page.Background.Layers
	tu.AssertEqual(t, len(layers), 1)
	tu.AssertEqual(t, layers[0].PaintingArea, pr.Rectangle{-1, -1, 6, 6})
}

func TestCrashSplitFirstLine(t *testing.T) {
	const input = `
	<style>
	.wrapper {
		min-height: 100%;
		display: flex;
		flex-direction: column;
	}
	main {
		flex-grow: 1;
	}
	</style>
	<body class="wrapper">
		<main>
			<h5>Recommandations :</h5>
		</main>
	</body>
	`
	_ = renderPages(t, input)
}
