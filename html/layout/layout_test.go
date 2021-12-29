package layout

import (
	"fmt"
	"io"
	"log"
	"testing"

	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango/fcfonts"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/layout/text"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

var baseUrl, _ = utils.PathToURL("../../resources_test/")

const fontmapCache = "text/test/cache.fc"

var fontconfig *text.FontConfiguration

func init() {
	logger.ProgressLogger.SetOutput(io.Discard)

	// this command has to run once
	// fmt.Println("Scanning fonts...")
	// _, err := fc.ScanAndCache(fontmapCache)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	fs, err := fc.LoadFontsetFile(fontmapCache)
	if err != nil {
		log.Fatal(err)
	}
	fontconfig = text.NewFontConfiguration(fcfonts.NewFontMap(fc.Standard.Copy(), fs))
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

// unpack 2 children
func unpack2(box Box) (c1, c2 Box) {
	return box.Box().Children[0], box.Box().Children[1]
}

// unpack 3 children
func unpack3(box Box) (c1, c2, c3 Box) {
	return box.Box().Children[0], box.Box().Children[1], box.Box().Children[2]
}

// unpack 4 children
func unpack4(box Box) (c1, c2, c3, c4 Box) {
	return box.Box().Children[0], box.Box().Children[1], box.Box().Children[2], box.Box().Children[3]
}

// unpack 5 children
func unpack5(box Box) (c1, c2, c3, c4, c5 Box) {
	return box.Box().Children[0], box.Box().Children[1], box.Box().Children[2], box.Box().Children[3], box.Box().Children[4]
}

// unpack 6 children
func unpack6(box Box) (c1, c2, c3, c4, c5, c6 Box) {
	return box.Box().Children[0], box.Box().Children[1], box.Box().Children[2], box.Box().Children[3], box.Box().Children[4], box.Box().Children[5]
}

// unpack 7 children
func unpack7(box Box) (c1, c2, c3, c4, c5, c6, c7 Box) {
	return box.Box().Children[0], box.Box().Children[1], box.Box().Children[2], box.Box().Children[3], box.Box().Children[4], box.Box().Children[5], box.Box().Children[6]
}

// unpack 8 children
func unpack8(box Box) (c1, c2, c3, c4, c5, c6, c7, c8 Box) {
	return box.Box().Children[0], box.Box().Children[1], box.Box().Children[2], box.Box().Children[3], box.Box().Children[4], box.Box().Children[5], box.Box().Children[6], box.Box().Children[7]
}

func TestMarginBoxes(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	tu.AssertEqual(t, page1.Children[0].Box().ElementTag(), "html", "page1")
	tu.AssertEqual(t, page2.Children[0].Box().ElementTag(), "html", "page2")

	var marginBoxes1, marginBoxes2 []string
	for _, box := range page1.Children[1:] {
		marginBoxes1 = append(marginBoxes1, box.(*bo.MarginBox).AtKeyword)
	}
	for _, box := range page2.Children[1:] {
		marginBoxes2 = append(marginBoxes2, box.(*bo.MarginBox).AtKeyword)
	}
	tu.AssertEqual(t, marginBoxes1, []string{"@top-center", "@bottom-left", "@bottom-left-corner"}, "marginBoxes1")
	tu.AssertEqual(t, marginBoxes2, []string{"@top-center"}, "marginBoxes2")

	if len(page2.Children) != 2 {
		t.Fatalf("expected two children, got %v", page2.Children)
	}
	_, topCenter := page2.Children[0], page2.Children[1]
	lineBox := topCenter.Box().Children[0]
	textBox, _ := lineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, textBox.Text, "Title", "textBox")
}

func TestMarginBoxStringSet1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	lineBox := bottomCenter.Box().Children[0]
	textBox, _ := lineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, textBox.Text, "first assignment", "textBox")

	if len(page1.Children) != 2 {
		t.Fatalf("expected two children, got %v", page1.Children)
	}
	_, bottomCenter = page1.Children[0], page1.Children[1]

	lineBox = bottomCenter.Box().Children[0]
	textBox, _ = lineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, textBox.Text, "first assignment", "textBox")
}

func TestMarginBoxStringSet2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
		lineBox := topCenter.Box().Children[0]
		textBox := lineBox.Box().Children[0].(*bo.TextBox)
		if contentVal == "before" || contentVal == "after" {
			tu.AssertEqual(t, textBox.Text, "pseudo", "textBox")
		} else {
			tu.AssertEqual(t, textBox.Text, "first assignment", "textBox")
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
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
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
	lineBox := topCenter.Box().Children[0]
	textBox := lineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, textBox.Text, "first assignment", "textBox")
}

func TestMarginBoxStringSet4(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	tu.AssertEqual(t, len(topCenter.Box().Children), 0, "Children")

	topCenter = page2.Box().Children[1]
	lineBox := topCenter.Box().Children[0]
	textBox := lineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, textBox.Text, "first_excepted", "textBox")
}

func TestMarginBoxStringSet5(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
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
	lineBox := topCenter.Box().Children[0]
	textBox := lineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, textBox.Text, "Second assignment", "textBox")
}

func TestMarginBoxStringSet6(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	topLineBox := topCenter.Box().Children[0]
	topTextBox := topLineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, topTextBox.Text, "before!-first-after!I.1", "topTextBox")

	bottomLineBox := bottomCenter.Box().Children[0]
	bottomTextBox := bottomLineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, bottomTextBox.Text, "before!last-secondclass2|1/I", "bottomTextBox")
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
	leftLineBox := topLeft.Box().Children[0]
	leftTextBox := leftLineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, leftTextBox.Text, "[Chocolate]", "leftTextBox")

	rightLineBox := topRight.Box().Children[0]
	rightTextBox := rightLineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, rightTextBox.Text, "{Cake}", "rightTextBox")
}

func TestMarginBoxStringSet8(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	leftLineBox := topLeft.Box().Children[0]
	leftTextBox := leftLineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, leftTextBox.Text, "[initial]", "leftTextBox")

	topLeft = page2.Box().Children[1]
	leftLineBox = topLeft.Box().Children[0]
	leftTextBox = leftLineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, leftTextBox.Text, "[]", "leftTextBox")

	topLeft = page3.Box().Children[1]
	leftLineBox = topLeft.Box().Children[0]
	leftTextBox = leftLineBox.Box().Children[0].(*bo.TextBox)
	tu.AssertEqual(t, leftTextBox.Text, "[ ]", "leftTextBox")
}

func TestMarginBoxStringSet9(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
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
	lineBox := topCenter.Box().Children[0]
	textBox := lineBox.Box().Children[0].(*bo.TextBox)

	tu.AssertEqual(t, textBox.Text, "first assignment second assignment", "textBox")
}

// Test page-based counters.
func TestPageCounters(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
		lineBox := bottomCenter.Box().Children[0]
		textBox := lineBox.Box().Children[0].(*bo.TextBox)
		exp := fmt.Sprintf("Page %d of 3.", pageNumber)
		tu.AssertEqual(t, textBox.Text, exp, fmt.Sprintf("page index %d", pageIndex))
	}
}
