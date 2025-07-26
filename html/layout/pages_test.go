package layout

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for pages layout.

// Test the layout for “@page“ properties.
func TestPageSizeBasic(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	for _, data := range []struct {
		size          string
		width, height int
	}{
		{"auto", 793, 1122},
		{"2in 10in", 192, 960},
		{"242px", 242, 242},
		{"letter", 816, 1056},
		{"letter portrait", 816, 1056},
		{"letter landscape", 1056, 816},
		{"portrait", 793, 1122},
		{"landscape", 1122, 793},
	} {
		page := renderOnePage(t, fmt.Sprintf("<style>@page { size: %s; }</style>", data.size))
		tu.AssertEqual(t, int(page.Box().MarginWidth()), data.width)
		tu.AssertEqual(t, int(page.Box().MarginHeight()), data.height)
	}
}

func TestPageSizeWithMargin(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `<style>
      @page { size: 200px 300px; margin: 10px 10% 20% 1in }
      body { margin: 8px }
    </style>
    <p style="margin: 0">`)
	tu.AssertEqual(t, page.Box().MarginWidth(), Fl(200))
	tu.AssertEqual(t, page.Box().MarginHeight(), Fl(300))
	tu.AssertEqual(t, page.Box().PositionX, Fl(0))
	tu.AssertEqual(t, page.Box().PositionY, Fl(0))
	tu.AssertEqual(t, page.Box().Width, Fl(84))
	tu.AssertEqual(t, page.Box().Height, Fl(230))

	html := unpack1(page)
	tu.AssertEqual(t, html.Box().ElementTag(), "html")
	tu.AssertEqual(t, html.Box().PositionX, Fl(96))
	tu.AssertEqual(t, html.Box().PositionY, Fl(10))
	tu.AssertEqual(t, html.Box().Width, Fl(84))

	body := unpack1(html)
	tu.AssertEqual(t, body.Box().ElementTag(), "body")
	tu.AssertEqual(t, body.Box().PositionX, Fl(96))
	tu.AssertEqual(t, body.Box().PositionY, Fl(10))
	// body has margins in the UA stylesheet
	tu.AssertEqual(t, body.Box().MarginLeft, Fl(8))
	tu.AssertEqual(t, body.Box().MarginRight, Fl(8))
	tu.AssertEqual(t, body.Box().MarginTop, Fl(8))
	tu.AssertEqual(t, body.Box().MarginBottom, Fl(8))
	tu.AssertEqual(t, body.Box().Width, Fl(68))

	paragraph := unpack1(body)
	tu.AssertEqual(t, paragraph.Box().ElementTag(), "p")
	tu.AssertEqual(t, paragraph.Box().PositionX, Fl(104))
	tu.AssertEqual(t, paragraph.Box().PositionY, Fl(18))
	tu.AssertEqual(t, paragraph.Box().Width, Fl(68))
}

func TestPageSizeWithMarginBorderPadding(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `<style> @page {
      size: 100px; margin: 1px 2px; padding: 4px 8px;
      border-width: 16px 32px; border-style: solid;
    }</style>`)
	tu.AssertEqual(t, page.Box().Width, Fl(16))
	tu.AssertEqual(t, page.Box().Height, Fl(58))
	html := unpack1(page)
	tu.AssertEqual(t, html.Box().ElementTag(), "html")
	tu.AssertEqual(t, html.Box().PositionX, Fl(42))
	tu.AssertEqual(t, html.Box().PositionY, Fl(21))
}

func TestPageSizeMargins(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range []struct {
		margin                   string
		top, right, bottom, left pr.Float
	}{
		{"auto", 15, 10, 15, 10},
		{"5px 5px auto auto", 5, 5, 25, 15},
	} {
		page := renderOnePage(t, fmt.Sprintf(`<style>@page {
      size: 106px 206px; width: 80px; height: 170px;
      padding: 1px; border: 2px solid; margin: %s }</style>`, data.margin))
		tu.AssertEqual(t, page.Box().MarginTop, data.top)
		tu.AssertEqual(t, page.Box().MarginRight, data.right)
		tu.AssertEqual(t, page.Box().MarginBottom, data.bottom)
		tu.AssertEqual(t, page.Box().MarginLeft, data.left)
	}
}

func TestPageSizeOverConstrained(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range []struct {
		style         string
		width, height pr.Float
	}{
		{
			"size: 4px 10000px; width: 100px; height: 100px;" +
				"padding: 1px; border: 2px solid; margin: 3px",
			112, 112,
		},
		{
			"size: 1000px; margin: 100px; max-width: 500px; min-height: 1500px",
			700, 1700,
		},
		{
			"size: 1000px; margin: 100px; min-width: 1500px; max-height: 500px",
			1700, 700,
		},
	} {
		page := renderOnePage(t, fmt.Sprintf("<style>@page { %s }</style>", data.style))
		tu.AssertEqual(t, page.Box().MarginWidth(), data.width)
		tu.AssertEqual(t, page.Box().MarginHeight(), data.height)
	}
}

func TestPageBreaks(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, html := range []string{
		"<div>1</div>",
		"<div></div>",
		"<img src=pattern.png>",
	} {
		pages := renderPages(t, fmt.Sprintf(`
      <style>
        @page { size: 100px; margin: 10px }
        body { margin: 0 }
        div { height: 30px; font-size: 20px }
        img { height: 30px; display: block }
      </style>
      %s`, strings.Repeat(html, 5)))
		var posY [][]pr.Float
		for _, page := range pages {
			html := unpack1(page)
			body := unpack1(html)
			children := body.Box().Children
			var pos []pr.Float
			for _, child := range children {
				tu.AssertEqual(t, child.Box().ElementTag() == "div" || child.Box().ElementTag() == "img", true)
				tu.AssertEqual(t, child.Box().PositionX, Fl(10))
				pos = append(pos, child.Box().PositionY)
			}
			posY = append(posY, pos)
		}
		tu.AssertEqual(t, posY, [][]pr.Float{{10, 40}, {10, 40}, {10}})
	}
}

func TestPageBreaksBoxSplit(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// If floats round the wrong way, a block that gets filled to the end of a
	// page due to breaking over the page may be forced onto the next page
	// because it is slightly taller than can fit on the previous page, even if
	// it wouldn't have been without being filled. These numbers aren't ideal,
	// but they do seem to trigger the issue.
	page1, page2 := renderTwoPages(t, `
      <style>
        @page { size: 982.4146981627297px; margin: 0 }
        div { font-size: 5px; height: 200.0123456789px; margin: 0; padding: 0 }
        figure { margin: 0; padding: 0 }
      </style>
      <div>text</div>
      <div>text</div><!-- no page break here -->
      <section>
        <div>line1</div>
        <div>line2</div><!-- page break here -->
        <div>line3</div>
        <div>line4</div>
      </section>
     `)
	html := unpack1(page1)
	body := unpack1(html)
	tu.AssertEqual(t, len(body.Box().Children), 3)
	_, _, section := unpack3(body)
	tu.AssertEqual(t, len(section.Box().Children), 2)

	html = unpack1(page2)
	body = unpack1(html)
	section = unpack1(body)
	tu.AssertEqual(t, len(section.Box().Children), 2)
}

func TestPageBreaksComplex1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { margin: 10px }
        @page :left { margin-left: 50px }
        @page :right { margin-right: 50px }
        html { page-break-before: left }
        div { page-break-after: left }
        ul { page-break-before: always }
      </style>
      <div>1</div>
      <p>2</p>
      <p>3</p>
      <article>
        <section>
          <ul><li>4</li></ul>
        </section>
      </article>
    `)
	page1, page2, page3, page4 := pages[0], pages[1], pages[2], pages[3]

	// The first page is a right page on rtl, but not here because of
	// page-break-before on the root element.
	tu.AssertEqual(t, page1.Box().MarginLeft, Fl(50))
	tu.AssertEqual(t, page1.Box().MarginRight, Fl(10))
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	line := unpack1(div)
	text := unpack1(line)
	tu.AssertEqual(t, div.Box().ElementTag(), "div")
	assertText(t, text, "1")

	html = unpack1(page2)
	tu.AssertEqual(t, page2.Box().MarginLeft, Fl(10))
	tu.AssertEqual(t, page2.Box().MarginRight, Fl(50))
	tu.AssertEqual(t, len(html.Box().Children), 0) // empty page to get toe

	tu.AssertEqual(t, page3.Box().MarginLeft, Fl(50))
	tu.AssertEqual(t, page3.Box().MarginRight, Fl(10))
	html = unpack1(page3)
	body = unpack1(html)
	p1, p2 := unpack2(body)
	tu.AssertEqual(t, p1.Box().ElementTag(), "p")
	tu.AssertEqual(t, p2.Box().ElementTag(), "p")

	tu.AssertEqual(t, page4.Box().MarginLeft, Fl(10))
	tu.AssertEqual(t, page4.Box().MarginRight, Fl(50))
	html = unpack1(page4)
	body = unpack1(html)
	article := unpack1(body)
	section := unpack1(article)
	ulist := unpack1(section)
	tu.AssertEqual(t, ulist.Box().ElementTag(), "ul")
}

func TestPageBreaksComplex2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Reference for the following test:
	// Without any "avoid", this breaks after the <div>
	pages := renderPages(t, `
      <style>
        @page { size: 140px; margin: 0 }
        img { height: 25px; vertical-align: top }
      </style>
      <img src=pattern.png>
      <div>
        <p><img src=pattern.png><br/><img src=pattern.png><p>
        <p><img src=pattern.png><br/><img src=pattern.png><p>
      </div><!-- page break here -->
      <img src=pattern.png>
    `)
	page1, page2 := pages[0], pages[1]

	html := unpack1(page1)
	body := unpack1(html)
	img1, div := unpack2(body)
	tu.AssertEqual(t, img1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, img1.Box().Height, Fl(25))
	tu.AssertEqual(t, div.Box().PositionY, Fl(25))
	tu.AssertEqual(t, div.Box().Height, Fl(100))

	html = unpack1(page2)
	body = unpack1(html)
	img2 := unpack1(body)
	tu.AssertEqual(t, img2.Box().PositionY, Fl(0))
	tu.AssertEqual(t, img2.Box().Height, Fl(25))
}

func TestPageBreaksComplex3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Adding a few page-break-*: avoid, the only legal break is
	// before the <div>
	pages := renderPages(t, `
      <style>
        @page { size: 140px; margin: 0 }
        img { height: 25px; vertical-align: top }
      </style>
      <img src=pattern.png><!-- page break here -->
      <div>
        <p style="page-break-inside: avoid">
          <img src=pattern.png><br/><img src=pattern.png></p>
        <p style="page-break-before: avoid; page-break-after: avoid; widows: 2"
          ><img src=pattern.png><br/><img src=pattern.png></p>
      </div>
      <img src=pattern.png>
    `)
	page1, page2 := pages[0], pages[1]

	html := unpack1(page1)
	body := unpack1(html)
	img1 := unpack1(body)
	tu.AssertEqual(t, img1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, img1.Box().Height, Fl(25))

	html = unpack1(page2)
	body = unpack1(html)
	div, img2 := unpack2(body)
	tu.AssertEqual(t, div.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div.Box().Height, Fl(100))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(100))
	tu.AssertEqual(t, img2.Box().Height, Fl(25))
}

func TestPageBreaksComplex4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 140px; margin: 0 }
        img { height: 25px; vertical-align: top }
      </style>
      <img src=pattern.png><!-- page break here -->
      <div>
        <div>
          <p style="page-break-inside: avoid">
            <img src=pattern.png><br/><img src=pattern.png></p>
          <p style="page-break-before:avoid; page-break-after:avoid; widows:2"
            ><img src=pattern.png><br/><img src=pattern.png></p>
        </div>
        <img src=pattern.png>
      </div>
    `)
	page1, page2 := pages[0], pages[1]

	html := unpack1(page1)
	body := unpack1(html)
	img1 := unpack1(body)
	tu.AssertEqual(t, img1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, img1.Box().Height, Fl(25))

	html = unpack1(page2)
	body = unpack1(html)
	outerDiv := unpack1(body)
	innerDiv, img2 := unpack2(outerDiv)
	tu.AssertEqual(t, innerDiv.Box().PositionY, Fl(0))
	tu.AssertEqual(t, innerDiv.Box().Height, Fl(100))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(100))
	tu.AssertEqual(t, img2.Box().Height, Fl(25))
}

func TestPageBreaksComplex5(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Reference for the next test
	pages := renderPages(t, `
      <style>
        @page { size: 100px; margin: 0 }
        img { height: 30px; display: block; }
      </style>
      <div>
        <img src=pattern.png style="page-break-after: always">
        <section>
          <img src=pattern.png>
          <img src=pattern.png>
        </section>
      </div>
      <img src=pattern.png><!-- page break here -->
      <img src=pattern.png>
    `)
	page1, page2, page3 := pages[0], pages[1], pages[2]

	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().Height, Fl(100))
	html = unpack1(page2)
	body = unpack1(html)
	div, img4 := unpack2(body)
	tu.AssertEqual(t, div.Box().Height, Fl(60))
	tu.AssertEqual(t, img4.Box().Height, Fl(30))
	html = unpack1(page3)
	body = unpack1(html)
	img5 := unpack1(body)
	tu.AssertEqual(t, img5.Box().Height, Fl(30))
}

func TestPageBreaksComplex6(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 100px; margin: 0 }
        img { height: 30px; display: block; }
      </style>
      <div>
        <img src=pattern.png style="page-break-after: always">
        <section>
          <img src=pattern.png><!-- page break here -->
          <img src=pattern.png style="page-break-after: avoid">
        </section>
      </div>
      <img src=pattern.png style="page-break-after: avoid">
      <img src=pattern.png>
    `)
	page1, page2, page3 := pages[0], pages[1], pages[2]

	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().Height, Fl(100))
	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	section := unpack1(div)
	img2 := unpack1(section)
	tu.AssertEqual(t, img2.Box().Height, Fl(30))
	// TODO: currently this is 60: we do not increase the used height of blocks
	// to make them fill the blank space at the end of the age when we remove
	// children from them for some break-*: avoid.
	// See TODOs in blocks.blockContainerLayout
	// tu.AssertEqual(t, div.Box().Height , Fl(100))
	html = unpack1(page3)
	body = unpack1(html)
	div, img4, img5 := unpack3(body)
	tu.AssertEqual(t, div.Box().Height, Fl(30))
	tu.AssertEqual(t, img4.Box().Height, Fl(30))
	tu.AssertEqual(t, img5.Box().Height, Fl(30))
}

func TestPageBreaksComplex7(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { @bottom-center { content: counter(page) } }
        @page:blank { @bottom-center { content: none } }
      </style>
      <p style="page-break-after: right">foo</p>
      <p>bar</p>
    `)
	page1, page2, page3 := pages[0], pages[1], pages[2]

	tu.AssertEqual(t, len(page1.Box().Children), 2)
	tu.AssertEqual(t, len(page2.Box().Children), 1)
	tu.AssertEqual(t, len(page3.Box().Children), 2)
}

func TestPageBreaksComplex8(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 75px; margin: 0 }
        div { height: 20px }
      </style>
      <div></div>
      <section>
        <div></div>
        <div style="page-break-after: avoid">
          <div style="position: absolute"></div>
          <div style="position: fixed"></div>
        </div>
      </section>
      <div></div>
    `)
	page1, page2 := pages[0], pages[1]

	html := unpack1(page1)
	body, _ := unpack2(html)
	div1, section := unpack2(body)
	div2 := unpack1(section)
	tu.AssertEqual(t, div1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div2.Box().PositionY, Fl(20))
	tu.AssertEqual(t, div1.Box().Height, Fl(20))
	tu.AssertEqual(t, div2.Box().Height, Fl(20))
	html = unpack1(page2)
	body = unpack1(html)
	section, div4 := unpack2(body)
	div3 := unpack1(section)
	_, _ = unpack2(div3)
	tu.AssertEqual(t, div3.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div4.Box().PositionY, Fl(20))
	tu.AssertEqual(t, div3.Box().Height, Fl(20))
	tu.AssertEqual(t, div4.Box().Height, Fl(20))
}

func TestPageBreaksComplex_9(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for #1979
	pages := renderPages(t, `
      <style>
        @page { size: 75px; margin: 0 }
        div { height: 20px; margin: 10px }
      </style>
      <div style="height: 40px"></div>
      <div></div>
      <div style="break-before: left"></div>
      <div style="break-before: right"></div>
    `)
	page_1, page_2, page_3, page_4, page_5 := pages[0], pages[1], pages[2], pages[3], pages[4]
	html := page_1.Box().Children[0]
	body := html.Box().Children[0]
	div_1 := body.Box().Children[0]
	tu.AssertEqual(t, div_1.Box().ContentBoxX(), pr.Float(10))
	tu.AssertEqual(t, div_1.Box().ContentBoxY(), pr.Float(10))
	html = page_2.Box().Children[0]
	body = html.Box().Children[0]
	div_2 := body.Box().Children[0]
	tu.AssertEqual(t, div_2.Box().ContentBoxX(), pr.Float(10))
	tu.AssertEqual(t, div_2.Box().ContentBoxY(), pr.Float(0)) // Unforced page break
	html = page_3.Box().Children[0]
	tu.AssertEqual(t, len(html.Box().Children), 0) // Empty page
	html = page_4.Box().Children[0]
	body = html.Box().Children[0]
	div_3 := body.Box().Children[0]
	tu.AssertEqual(t, div_3.Box().ContentBoxX(), pr.Float(10))
	tu.AssertEqual(t, div_3.Box().ContentBoxY(), pr.Float(10)) // Forced page break
	html = page_5.Box().Children[0]
	body = html.Box().Children[0]
	div_4 := body.Box().Children[0]
	tu.AssertEqual(t, div_4.Box().ContentBoxX(), pr.Float(10))
	tu.AssertEqual(t, div_4.Box().ContentBoxY(), pr.Float(10)) // Forced page break
}

func TestMarginBreak(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range []struct {
		breakAfter, marginBreak string
		marginTop               pr.Float
	}{
		{"page", "auto", 5},
		{"auto", "auto", 0},
		{"page", "keep", 5},
		{"auto", "keep", 5},
		{"page", "discard", 0},
		{"auto", "discard", 0},
	} {
		pages := renderPages(t, fmt.Sprintf(`
		<style>
			@page { size: 70px; margin: 0 }
			div { height: 63px; margin: 5px 0 8px;
				break-after: %s; margin-break: %s }
		</style>
		<section>
			<div></div>
		</section>
		<section>
			<div></div>
		</section>
		`, data.breakAfter, data.marginBreak))
		page1, page2 := pages[0], pages[1]

		html := unpack1(page1)
		body := unpack1(html)
		section := unpack1(body)
		div := unpack1(section)
		tu.AssertEqual(t, div.Box().MarginTop, Fl(5))

		html = unpack1(page2)
		body = unpack1(html)
		section = unpack1(body)
		div = unpack1(section)
		tu.AssertEqual(t, div.Box().MarginTop, data.marginTop)
	}
}

// @pytest.mark.xfail

// func TestMarginBreakClearance(t *testing.T){
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     page1, page2 = renderPages(`
//       <style>
//         @page { size: 70px; margin: 0 }
//         div { height: 63px; margin: 5px 0 8px; break-after: page }
//       </style>
//       <section>
//         <div></div>
//       </section>
//       <section>
//         <div style="border-top: 1px solid black">
//           <div></div>
//         </div>
//       </section>
//     `)
//     html := unpack1(page1)
//     body :=  unpack1(html)
//     section, = body.Box().Children
//     div := unpack1(section)
//     tu.AssertEqual(t, div.Box().MarginTop , 5)
//
//     html := unpack1(page2)
//     body :=  unpack1(html)
//     section, = body.Box().Children
//     div1, = section.Box().Children
//     tu.AssertEqual(t, div1.Box().MarginTop , 0)
//     div2, = div1.Box().Children
//     tu.AssertEqual(t, div2.Box().MarginTop , 5)
//     tu.AssertEqual(t, div2.contentBoxY() , 5)

func TestRectoVersoBreak(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range []struct {
		direction, pageBreak string
		pagesNumber          int
	}{
		{"ltr", "recto", 3},
		{"ltr", "verso", 2},
		{"rtl", "recto", 3},
		{"rtl", "verso", 2},
		{"ltr", "right", 3},
		{"ltr", "left", 2},
		{"rtl", "right", 2},
		{"rtl", "left", 3},
	} {
		pages := renderPages(t, fmt.Sprintf(`
      <style>
        html { direction: %s }
        p { break-before: %s }
      </style>
      abc
      <p>def</p>
    `, data.direction, data.pageBreak))
		tu.AssertEqual(t, len(pages), data.pagesNumber)
	}
}

func TestRectoVersoBreakRoot(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, test := range []struct {
		direction string
		pageBreak string
		width     int
	}{
		{"ltr", "recto", 5},
		{"ltr", "verso", 4},
		{"rtl", "recto", 4},
		{"rtl", "verso", 5},
		{"ltr", "right", 5},
		{"ltr", "left", 4},
		{"rtl", "right", 5},
		{"rtl", "left", 4},
	} {

		page := renderOnePage(t, fmt.Sprintf(`
		<style>
        @page:left { size: 4px /* for 'left' */ }
        @page:right { size: 5px /* for 'right' */ }
        html { direction: %s; break-before: %s }
		</style>
		abc
		`, test.direction, test.pageBreak))
		tu.AssertEqual(t, page.Width, Fl(test.width))
	}
}

func TestPageNames1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 100px 100px }
        section { page: small }
      </style>
      <div>
        <section>large</section>
      </div>
    `)
	page1 := pages[0]
	tu.AssertEqual(t, page1.Box().Width, Fl(100))
	tu.AssertEqual(t, page1.Box().Height, Fl(100))
}

func TestPageNames2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 100px 100px }
        @page narrow { margin: 1px }
        section { page: small }
      </style>
      <div>
        <section>large</section>
      </div>
    `)
	page1 := pages[0]
	tu.AssertEqual(t, page1.Box().Width, Fl(100))
	tu.AssertEqual(t, page1.Box().Height, Fl(100))
}

func TestPageNames3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { margin: 0 }
        @page narrow { size: 100px 200px }
        @page large { size: 200px 100px }
        div { page: narrow }
        section { page: large }
      </style>
      <div>
        <section>large</section>
        <section>large</section>
        <p>narrow</p>
      </div>
    `)
	page1, page2 := pages[0], pages[1]

	tu.AssertEqual(t, page1.Box().Width, Fl(200))
	tu.AssertEqual(t, page1.Box().Height, Fl(100))
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	section1, section2 := unpack2(div)
	tu.AssertEqual(t, section1.Box().ElementTag(), "section")
	tu.AssertEqual(t, section2.Box().ElementTag(), "section")

	tu.AssertEqual(t, page2.Box().Width, Fl(100))
	tu.AssertEqual(t, page2.Box().Height, Fl(200))
	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	p := unpack1(div)
	tu.AssertEqual(t, p.Box().ElementTag(), "p")
}

func TestPageNames4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 200px 200px; margin: 0 }
        @page small { size: 100px 100px }
        p { page: small }
      </style>
      <section>normal</section>
      <section>normal</section>
      <p>small</p>
      <section>small</section>
    `)
	page1, page2 := pages[0], pages[1]

	tu.AssertEqual(t, page1.Box().Width, Fl(200))
	tu.AssertEqual(t, page1.Box().Height, Fl(200))
	html := unpack1(page1)
	body := unpack1(html)
	section1, section2 := unpack2(body)
	tu.AssertEqual(t, section1.Box().ElementTag(), "section")
	tu.AssertEqual(t, section2.Box().ElementTag(), "section")

	tu.AssertEqual(t, page2.Box().Width, Fl(100))
	tu.AssertEqual(t, page2.Box().Height, Fl(100))
	html = unpack1(page2)
	body = unpack1(html)
	p, section := unpack2(body)
	tu.AssertEqual(t, p.Box().ElementTag(), "p")
	tu.AssertEqual(t, section.Box().ElementTag(), "section")
}

func TestPageNames5(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 200px 200px; margin: 0 }
        @page small { size: 100px 100px }
        div { page: small }
      </style>
      <section><p>a</p>b</section>
      <section>c<div>d</div></section>
    `)
	page1, page2 := pages[0], pages[1]

	tu.AssertEqual(t, page1.Box().Width, Fl(200))
	tu.AssertEqual(t, page1.Box().Height, Fl(200))
	html := unpack1(page1)
	body := unpack1(html)
	section1, section2 := unpack2(body)
	tu.AssertEqual(t, section1.Box().ElementTag(), "section")
	tu.AssertEqual(t, section2.Box().ElementTag(), "section")
	_, _ = unpack2(section1)
	_ = unpack1(section2)

	tu.AssertEqual(t, page2.Box().Width, Fl(100))
	tu.AssertEqual(t, page2.Box().Height, Fl(100))
	html = unpack1(page2)
	body = unpack1(html)
	section2 = unpack1(body)
	_ = unpack1(section2)
}

func TestPageNames6(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { margin: 0 }
        @page large { size: 200px 200px }
        @page small { size: 100px 100px }
        section { page: large }
        div { page: small }
      </style>
      <section>a<p>b</p>c</section>
      <section>d<div>e</div>f</section>
    `)
	page1, page2, page3 := pages[0], pages[1], pages[2]

	tu.AssertEqual(t, page1.Box().Width, Fl(200))
	tu.AssertEqual(t, page1.Box().Height, Fl(200))
	html := unpack1(page1)
	body := unpack1(html)
	section1, section2 := unpack2(body)
	tu.AssertEqual(t, section1.Box().ElementTag(), "section")
	tu.AssertEqual(t, section2.Box().ElementTag(), "section")
	_, _, _ = unpack3(section1)
	_ = unpack1(section2)

	tu.AssertEqual(t, page2.Box().Width, Fl(100))
	tu.AssertEqual(t, page2.Box().Height, Fl(100))
	html = unpack1(page2)
	body = unpack1(html)
	section2 = unpack1(body)
	_ = unpack1(section2)

	tu.AssertEqual(t, page3.Box().Width, Fl(200))
	tu.AssertEqual(t, page3.Box().Height, Fl(200))
	html = unpack1(page3)
	body = unpack1(html)
	section2 = unpack1(body)
	_ = unpack1(section2)
}

func TestPageNames7(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 200px 200px; margin: 0 }
        @page small { size: 100px 100px }
        p { page: small; break-before: right }
      </style>
      <section>normal</section>
      <section>normal</section>
      <p>small</p>
      <section>small</section>
    `)
	page1, page2, page3 := pages[0], pages[1], pages[2]

	tu.AssertEqual(t, page1.Box().Width, Fl(200))
	tu.AssertEqual(t, page1.Box().Height, Fl(200))
	html := unpack1(page1)
	body := unpack1(html)
	section1, section2 := unpack2(body)
	tu.AssertEqual(t, section1.Box().ElementTag(), "section")
	tu.AssertEqual(t, section2.Box().ElementTag(), "section")

	tu.AssertEqual(t, page2.Box().Width, Fl(200))
	tu.AssertEqual(t, page2.Box().Height, Fl(200))
	html = unpack1(page2)
	tu.AssertEqual(t, len(html.Box().Children), 0)

	tu.AssertEqual(t, page3.Box().Width, Fl(100))
	tu.AssertEqual(t, page3.Box().Height, Fl(100))
	html = unpack1(page3)
	body = unpack1(html)
	p, section := unpack2(body)
	tu.AssertEqual(t, p.Box().ElementTag(), "p")
	tu.AssertEqual(t, section.Box().ElementTag(), "section")
}

func TestPageNames8(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page small { size: 100px 100px }
        section { page: small }
        p { line-height: 80px }
      </style>
      <section>
        <p>small</p>
        <p>small</p>
      </section>
    `)
	page1, page2 := pages[0], pages[1]

	tu.AssertEqual(t, page1.Box().Width, Fl(100))
	tu.AssertEqual(t, page1.Box().Height, Fl(100))
	html := unpack1(page1)
	body := unpack1(html)
	section := unpack1(body)
	p := unpack1(section)
	tu.AssertEqual(t, section.Box().ElementTag(), "section")
	tu.AssertEqual(t, p.Box().ElementTag(), "p")

	tu.AssertEqual(t, page2.Box().Width, Fl(100))
	tu.AssertEqual(t, page2.Box().Height, Fl(100))
	html = unpack1(page2)
	body = unpack1(html)
	section = unpack1(body)
	p = unpack1(section)
	tu.AssertEqual(t, section.Box().ElementTag(), "section")
	tu.AssertEqual(t, p.Box().ElementTag(), "p")
}

func TestPageNames9(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 200px 200px }
        @page small { size: 100px 100px }
        section { break-after: page; page: small }
        article { page: small }
      </style>
      <section>
        <div>big</div>
        <div>big</div>
      </section>
      <article>
        <div>small</div>
        <div>small</div>
      </article>
    `)
	page1, page2 := pages[0], pages[1]

	tu.AssertEqual(t, page1.Box().Width, Fl(100))
	tu.AssertEqual(t, page1.Box().Height, Fl(100))
	html := unpack1(page1)
	body := unpack1(html)
	section := unpack1(body)
	tu.AssertEqual(t, section.Box().ElementTag(), "section")

	tu.AssertEqual(t, page2.Box().Width, Fl(100))
	tu.AssertEqual(t, page2.Box().Height, Fl(100))
	html = unpack1(page2)
	body = unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().ElementTag(), "article")
}

func TestPageNames10(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	pages := renderPages(t, `
      <style>
        #running { position: running(running); }
        #fixed { position: fixed; }
        @page { size: 200px 200px; @top-center { content: element(header); }}
        section { page: small; }
        @page small { size: 100px 100px; }
        .pagebreak { break-after: page; }
      </style>
      <div id="running">running</div>
      <div id="fixed">fixed</div>
      <section>
        <h1>text</h1>
        <div class="pagebreak"></div>
        <article>text</article>
      </section>
    `)
	page1, page2 := pages[0], pages[1]

	tu.AssertEqual(t, page1.Box().Width, Fl(100))
	tu.AssertEqual(t, page1.Box().Height, Fl(100))
	html, _ := unpack2(page1)
	body := html.Box().Children[0]
	_, section := unpack2(body)
	h1, _ := unpack2(section)
	tu.AssertEqual(t, h1.Box().ElementTag(), "h1")

	tu.AssertEqual(t, page2.Box().Width, Fl(100))
	tu.AssertEqual(t, page2.Box().Height, Fl(100))
	html, _ = unpack2(page2)
	_, body = unpack2(html)
	section = body.Box().Children[0]
	article := section.Box().Children[0]
	tu.AssertEqual(t, article.Box().ElementTag(), "article")
}

func TestPageGroups(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 200px 200px }
        @page small { size: 100px 100px }
        @page :nth(1 of small) { size: 50px 50px }
        section { page: small }
        div, div section { break-after: page }
      </style>
      <div></div>
      <article></article>
      <section>
        <div></div>
        <div></div>
      </section>
      <section>
      </section>
      <div></div>
      <div></div>
      <section>
        <div></div>
      </section>
      <div>
        <section></section>
        <section></section>
      </div>
    `)
	page1, page2, page3, page4, page5, page6, page7, page8, page9 := pages[0], pages[1], pages[2], pages[3], pages[4], pages[5], pages[6], pages[7], pages[8]

	tu.AssertEqual(t, [2]pr.Float{page1.Box().Width.V(), page1.Box().Height.V()}, [2]pr.Float{200, 200})
	div := unpack1(unpack1(unpack1(page1)))
	tu.AssertEqual(t, div.Box().ElementTag, "div")

	tu.AssertEqual(t, [2]pr.Float{page2.Box().Width.V(), page2.Box().Height.V()}, [2]pr.Float{200, 200})
	article := unpack1(unpack1(unpack1(page2)))
	tu.AssertEqual(t, article.Box().ElementTag, "article")

	tu.AssertEqual(t, [2]pr.Float{page3.Box().Width.V(), page3.Box().Height.V()}, [2]pr.Float{50, 50})
	section := unpack1(unpack1(unpack1(page3)))
	tu.AssertEqual(t, section.Box().ElementTag, "section")
	div = unpack1(section)
	tu.AssertEqual(t, div.Box().ElementTag, "div")

	tu.AssertEqual(t, [2]pr.Float{page4.Box().Width.V(), page4.Box().Height.V()}, [2]pr.Float{100, 100})
	section = unpack1(unpack1(unpack1(page4)))
	tu.AssertEqual(t, section.Box().ElementTag, "section")
	div = unpack1(section)
	tu.AssertEqual(t, div.Box().ElementTag, "div")

	tu.AssertEqual(t, [2]pr.Float{page5.Box().Width.V(), page5.Box().Height.V()}, [2]pr.Float{50, 50})
	section, div = unpack2(unpack1(unpack1(page5)))
	tu.AssertEqual(t, section.Box().ElementTag, "section")
	tu.AssertEqual(t, div.Box().ElementTag, "div")

	tu.AssertEqual(t, [2]pr.Float{page6.Box().Width.V(), page6.Box().Height.V()}, [2]pr.Float{200, 200})
	div = unpack1(unpack1(unpack1(page6)))
	tu.AssertEqual(t, div.Box().ElementTag, "div")

	tu.AssertEqual(t, [2]pr.Float{page7.Box().Width.V(), page7.Box().Height.V()}, [2]pr.Float{50, 50})
	section = unpack1(unpack1(unpack1(page7)))
	tu.AssertEqual(t, section.Box().ElementTag, "section")
	div = unpack1(section)
	tu.AssertEqual(t, div.Box().ElementTag, "div")

	tu.AssertEqual(t, [2]pr.Float{page8.Box().Width.V(), page8.Box().Height.V()}, [2]pr.Float{50, 50})
	div = unpack1(unpack1(unpack1(page8)))
	tu.AssertEqual(t, div.Box().ElementTag, "div")
	section = unpack1(div)
	tu.AssertEqual(t, section.Box().ElementTag, "section")

	tu.AssertEqual(t, [2]pr.Float{page9.Box().Width.V(), page9.Box().Height.V()}, [2]pr.Float{50, 50})
	div = unpack1(unpack1(unpack1(page9)))
	tu.AssertEqual(t, div.Box().ElementTag, "div")
	section = unpack1(div)
	tu.AssertEqual(t, section.Box().ElementTag, "section")
}

// Regression test for #1076.
func TestPageGroupsBlankInside(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 100px }
        @page div { size: 50px }
        div { page: div }
        p { break-before: right }
      </style>
      <div>
        <p>1</p>
        <p>2</p>
      </div>
    `)
	tu.AssertEqual(t, len(pages), 3)
	for _, page := range pages {
		tu.AssertEqual(t, [2]pr.Float{page.Box().Width.V(), page.Box().Height.V()}, [2]pr.Float{50, 50})
	}
}

func TestPageGroupsBlankOutside(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 100px }
        @page p { size: 50px }
        p { page: p; break-before: right }
      </style>
      <div>
        <p>1</p>
        <p>2</p>
      </div>
    `)
	tu.AssertEqual(t, len(pages), 3)
	page1, page2, page3 := pages[0], pages[1], pages[2]
	tu.AssertEqual(t, [2]pr.Float{page1.Box().Width.V(), page1.Box().Height.V()}, [2]pr.Float{50, 50})
	tu.AssertEqual(t, [2]pr.Float{page2.Box().Width.V(), page2.Box().Height.V()}, [2]pr.Float{100, 100})
	tu.AssertEqual(t, [2]pr.Float{page3.Box().Width.V(), page3.Box().Height.V()}, [2]pr.Float{50, 50})
}

// Regression test for #2429.
func TestPageGroupsFirstNth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 100px }
        @page div { size: 50px }
        @page :nth(2n+1 of div) { size: 30px }
        div { page: div; break-before: right }
        p { break-before: page }
      </style>
      <div>
        <p>1</p>
        <p>2</p>
        <p>3</p>
      </div>
      <div>
        <p>4</p>
        <p>5</p>
      </div>
    `)
	tu.AssertEqual(t, [2]pr.Float{pages[0].Box().Width.V(), pages[0].Box().Height.V()}, [2]pr.Float{30, 30})
	tu.AssertEqual(t, [2]pr.Float{pages[1].Box().Width.V(), pages[1].Box().Height.V()}, [2]pr.Float{50, 50})
	tu.AssertEqual(t, [2]pr.Float{pages[2].Box().Width.V(), pages[2].Box().Height.V()}, [2]pr.Float{30, 30})
	tu.AssertEqual(t, [2]pr.Float{pages[3].Box().Width.V(), pages[3].Box().Height.V()}, [2]pr.Float{100, 100})
	tu.AssertEqual(t, [2]pr.Float{pages[4].Box().Width.V(), pages[4].Box().Height.V()}, [2]pr.Float{30, 30})
	tu.AssertEqual(t, [2]pr.Float{pages[5].Box().Width.V(), pages[5].Box().Height.V()}, [2]pr.Float{50, 50})
}

func TestOrphansWidowsAvoid(t *testing.T) {
	// capt := tu.CaptureLogs()
	// defer capt.AssertNoLogs(t)

	for _, data := range []struct {
		style      string
		lineCounts [2]int
	}{
		{"orphans: 2; widows: 2", [2]int{4, 3}},
		{"orphans: 5; widows: 2", [2]int{0, 7}},
		{"orphans: 2; widows: 4", [2]int{3, 4}},
		{"orphans: 4; widows: 4", [2]int{0, 7}},
		{"orphans: 2; widows: 2; page-break-inside: avoid", [2]int{0, 7}},
	} {
		pages := renderPages(t, fmt.Sprintf(`
		<style>
			@page { size: 200px }
			h1 { height: 120px }
			p { line-height: 20px;
				width: 1px; /* line break at each word */
				%s }
		</style>
		<h1>Tasty test</h1>
		<!-- There is room for 4 lines after h1 on the first page -->
		<p>one two three four five six seven</p>
		`, data.style))
		tu.AssertEqual(t, len(pages), 2)
		for i, page := range pages {
			html := unpack1(page)
			body := unpack1(html)
			bodyChildren := body.Box().Children
			if i == 0 {
				bodyChildren = bodyChildren[1:] // skip h1
			}
			var count int
			if len(bodyChildren) != 0 {
				count = len(bodyChildren[0].Box().Children)
			}
			tu.AssertEqual(t, count, data.lineCounts[i])
		}
	}
}

func TestPageAndLineboxBreaking(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Empty <span/> tests a corner case in skipFirstWhitespace()
	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 100px; margin: 2px; border: 1px solid }
        body { margin: 0 }
        div { font-family: weasyprint; font-size: 20px }
      </style>
      <div><span/>1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15</div>
    `)
	var texts []string
	for _, page := range pages {
		html := unpack1(page)
		body := unpack1(html)
		div := unpack1(body)
		lines := div.Box().Children
		for _, line := range lines {
			var lineTexts []string
			for _, child := range bo.Descendants(line) {
				if child, ok := child.(*bo.TextBox); ok {
					lineTexts = append(lineTexts, child.TextS())
				}
			}
			texts = append(texts, strings.Join(lineTexts, "a"))
		}
	}
	tu.AssertEqual(t, len(pages), 4)
	tu.AssertEqual(t, strings.Join(texts, ""), "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15")
}

func TestMarginBoxesFixedDimension1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Corner boxes
	page := renderOnePage(t, `
      <style>
        @page {
          @top-left-corner {
            content: "topLeft";
            padding: 10px;
          }
          @top-right-corner {
            content: "topRight";
            padding: 10px;
          }
          @bottom-left-corner {
            content: "bottomLeft";
            padding: 10px;
          }
          @bottom-right-corner {
            content: "bottomRight";
            padding: 10px;
          }
          size: 1000px;
          margin-top: 10%;
          margin-bottom: 40%;
          margin-left: 20%;
          margin-right: 30%;
        }
      </style>
    `)
	_, topLeft, topRight, bottomLeft, bottomRight := unpack5(page)
	for i, textExp := range []string{"topLeft", "topRight", "bottomLeft", "bottomRight"} {
		marginBox := []Box{topLeft, topRight, bottomLeft, bottomRight}[i]
		line := unpack1(marginBox)
		text := unpack1(line)
		assertText(t, text, textExp)
	}

	// Check positioning && Rule 1 for fixed dimensions
	tu.AssertEqual(t, topLeft.Box().PositionX, Fl(0))
	tu.AssertEqual(t, topLeft.Box().PositionY, Fl(0))
	tu.AssertEqual(t, topLeft.Box().MarginWidth(), Fl(200))
	tu.AssertEqual(t, topLeft.Box().MarginHeight(), Fl(100))

	tu.AssertEqual(t, topRight.Box().PositionX, Fl(700))
	tu.AssertEqual(t, topRight.Box().PositionY, Fl(0))
	tu.AssertEqual(t, topRight.Box().MarginWidth(), Fl(300))
	tu.AssertEqual(t, topRight.Box().MarginHeight(), Fl(100))

	tu.AssertEqual(t, bottomLeft.Box().PositionX, Fl(0))
	tu.AssertEqual(t, bottomLeft.Box().PositionY, Fl(600))
	tu.AssertEqual(t, bottomLeft.Box().MarginWidth(), Fl(200))
	tu.AssertEqual(t, bottomLeft.Box().MarginHeight(), Fl(400))

	tu.AssertEqual(t, bottomRight.Box().PositionX, Fl(700))
	tu.AssertEqual(t, bottomRight.Box().PositionY, Fl(600))
	tu.AssertEqual(t, bottomRight.Box().MarginWidth(), Fl(300))
	tu.AssertEqual(t, bottomRight.Box().MarginHeight(), Fl(400))
}

func TestMarginBoxesFixedDimension2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test rules 2 && 3
	page := renderOnePage(t, `
      <style>
        @page {
          margin: 100px 200px;
          @bottom-left-corner { content: ""; margin: 60px }
        }
      </style>
    `)
	_, marginBox := unpack2(page)
	tu.AssertEqual(t, marginBox.Box().MarginWidth(), Fl(200))
	tu.AssertEqual(t, marginBox.Box().MarginLeft, Fl(60))
	tu.AssertEqual(t, marginBox.Box().MarginRight, Fl(60))
	tu.AssertEqual(t, marginBox.Box().Width, Fl(80))

	tu.AssertEqual(t, marginBox.Box().MarginHeight(), Fl(100))
	// total was too big, the outside margin was ignored :
	tu.AssertEqual(t, marginBox.Box().MarginTop, Fl(60))
	tu.AssertEqual(t, marginBox.Box().MarginBottom, Fl(40))
	tu.AssertEqual(t, marginBox.Box().Height, Fl(0))
}

func TestMarginBoxesFixedDimension3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test rule 3 with a non-auto inner dimension
	page := renderOnePage(t, `
      <style>
        @page {
          margin: 100px;
          @left-middle { content: ""; margin: 10px; width: 130px }
        }
      </style>
    `)
	_, marginBox := unpack2(page)
	tu.AssertEqual(t, marginBox.Box().MarginWidth(), Fl(100))
	tu.AssertEqual(t, marginBox.Box().MarginLeft, Fl(-40))
	tu.AssertEqual(t, marginBox.Box().MarginRight, Fl(10))
	tu.AssertEqual(t, marginBox.Box().Width, Fl(130))
}

func TestMarginBoxesFixedDimension4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test rule 4
	page := renderOnePage(t, `
      <style>
        @page {
          margin: 100px;
          @left-bottom {
            content: "";
            margin-left: 10px;
            margin-right: auto;
            width: 70px;
          }
        }
      </style>
    `)
	_, marginBox := unpack2(page)
	tu.AssertEqual(t, marginBox.Box().MarginWidth(), Fl(100))
	tu.AssertEqual(t, marginBox.Box().MarginLeft, Fl(10))
	tu.AssertEqual(t, marginBox.Box().MarginRight, Fl(20))
	tu.AssertEqual(t, marginBox.Box().Width, Fl(70))
}

func TestMarginBoxesFixedDimension5(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test rules 2, 3 && 4
	page := renderOnePage(t, `
      <style>
        @page {
          margin: 100px;
          @right-top {
            content: "";
            margin-right: 10px;
            margin-left: auto;
            width: 130px;
          }
        }
      </style>
    `)
	_, marginBox := unpack2(page)
	tu.AssertEqual(t, marginBox.Box().MarginWidth(), Fl(100))
	tu.AssertEqual(t, marginBox.Box().MarginLeft, Fl(0))
	tu.AssertEqual(t, marginBox.Box().MarginRight, Fl(-30))
	tu.AssertEqual(t, marginBox.Box().Width, Fl(130))
}

func TestMarginBoxesFixedDimension6(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test rule 5
	page := renderOnePage(t, `
      <style>
        @page {
          margin: 100px;
          @top-left { content: ""; margin-top: 10px; margin-bottom: auto }
        }
      </style>
    `)
	_, marginBox := unpack2(page)
	tu.AssertEqual(t, marginBox.Box().MarginHeight(), Fl(100))
	tu.AssertEqual(t, marginBox.Box().MarginTop, Fl(10))
	tu.AssertEqual(t, marginBox.Box().MarginBottom, Fl(0))
	tu.AssertEqual(t, marginBox.Box().Height, Fl(90))
}

func TestMarginBoxesFixedDimension7(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test rule 5
	page := renderOnePage(t, `
      <style>
        @page {
          margin: 100px;
          @top-center { content: ""; margin: auto 0 }
        }
      </style>
    `)
	_, marginBox := unpack2(page)
	tu.AssertEqual(t, marginBox.Box().MarginHeight(), Fl(100))
	tu.AssertEqual(t, marginBox.Box().MarginTop, Fl(0))
	tu.AssertEqual(t, marginBox.Box().MarginBottom, Fl(0))
	tu.AssertEqual(t, marginBox.Box().Height, Fl(100))
}

func TestMarginBoxesFixedDimension8(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test rule 6
	page := renderOnePage(t, `
      <style>
        @page {
          margin: 100px;
          @bottom-right { content: ""; margin: auto; height: 70px }
        }
      </style>
    `)
	_, marginBox := unpack2(page)
	tu.AssertEqual(t, marginBox.Box().MarginHeight(), Fl(100))
	tu.AssertEqual(t, marginBox.Box().MarginTop, Fl(15))
	tu.AssertEqual(t, marginBox.Box().MarginBottom, Fl(15))
	tu.AssertEqual(t, marginBox.Box().Height, Fl(70))
}

func TestMarginBoxesFixedDimension9(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Rule 2 inhibits rule 6
	page := renderOnePage(t, `
      <style>
        @page {
          margin: 100px;
          @bottom-center { content: ""; margin: auto 0; height: 150px }
        }
      </style>
    `)
	_, marginBox := unpack2(page)
	tu.AssertEqual(t, marginBox.Box().MarginHeight(), Fl(100))
	tu.AssertEqual(t, marginBox.Box().MarginTop, Fl(0))
	tu.AssertEqual(t, marginBox.Box().MarginBottom, Fl(-50))
	tu.AssertEqual(t, marginBox.Box().Height, Fl(150))
}

func imagesFromW(widths ...int) string {
	var chunks []string
	for _, width := range widths {
		chunks = append(chunks, `url('data:image/svg+xml,`+url.PathEscape(fmt.Sprintf(`<svg width="%d" height="10"></svg>`, width))+`')`)
	}
	return strings.Join(chunks, " ")
}

func TestPageStyle(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range []struct {
		css    string
		widths []pr.Float
	}{
		{fmt.Sprintf(`@top-left { content: %s }
        @top-center { content: %s }
        @top-right { content: %s }
     `, imagesFromW(50, 50), imagesFromW(50, 50), imagesFromW(50, 50)), []pr.Float{100, 100, 100}}, // Use preferred widths if they fit
		{fmt.Sprintf(`@top-left { content: %s; margin: auto }
        @top-center { content: %s }
        @top-right { content: %s }
     `, imagesFromW(50, 50), imagesFromW(50, 50), imagesFromW(50, 50)), []pr.Float{100, 100, 100}}, // "auto" margins are set to 0
		{fmt.Sprintf(`@top-left { content: %s }
        @top-center { content: %s }
        @top-right { content: "foo"; width: 200px }
     `, imagesFromW(100, 50), imagesFromW(300, 150)), []pr.Float{150, 300, 200}}, // Use at least minimum widths, even if boxes overlap
		{fmt.Sprintf(`@top-left { content: %s }
        @top-center { content: %s }
        @top-right { content: %s }
     `, imagesFromW(150, 150), imagesFromW(150, 150), imagesFromW(150, 150)), []pr.Float{200, 200, 200}}, // Distribute remaining space proportionally
		{fmt.Sprintf(`@top-left { content: %s }
        @top-center { content: %s }
        @top-right { content: %s }
     `, imagesFromW(100, 100, 100), imagesFromW(100, 100), imagesFromW(10)), []pr.Float{220, 160, 10}},
		{fmt.Sprintf(`@top-left { content: %s; width: 205px }
        @top-center { content: %s }
        @top-right { content: %s }
     `, imagesFromW(100, 100, 100), imagesFromW(100, 100), imagesFromW(10)), []pr.Float{205, 190, 10}},
		{fmt.Sprintf(`@top-left { width: 1000px; margin: 1000px; padding: 1000px;
                    border: 1000px solid }
        @top-center { content: %s }
        @top-right { content: %s }
     `, imagesFromW(100, 100), imagesFromW(10)), []pr.Float{200, 10}}, // "width" && other have no effect without "content"
		{
			fmt.Sprintf(`@top-left { content: ""; width: 200px }
        @top-center { content: ""; width: 300px }
        @top-right { content: %s }
     `, imagesFromW(50, 50)), // This leaves 150px for @top-right’s shrink-to-fit
			[]pr.Float{200, 300, 100},
		},
		{
			fmt.Sprintf(`@top-left { content: ""; width: 200px }
        @top-center { content: ""; width: 300px }
        @top-right { content: %s }
     `, imagesFromW(100, 100, 100)),
			[]pr.Float{200, 300, 150},
		},
		{fmt.Sprintf(`@top-left { content: ""; width: 200px }
        @top-center { content: ""; width: 300px }
        @top-right { content: %s }
     `, imagesFromW(170, 175)), []pr.Float{200, 300, 175}},
		{fmt.Sprintf(`@top-left { content: ""; width: 200px }
        @top-center { content: ""; width: 300px }
        @top-right { content: %s }
     `, imagesFromW(170, 175)), []pr.Float{200, 300, 175}},
		{`@top-left { content: ""; width: 200px }
        @top-right { content: ""; width: 500px }
     `, []pr.Float{200, 500}},
		{fmt.Sprintf(`@top-left { content: ""; width: 200px }
        @top-right { content: %s }
     `, imagesFromW(150, 50, 150)), []pr.Float{200, 350}},
		{fmt.Sprintf(`@top-left { content: ""; width: 200px }
        @top-right { content: %s }
     `, imagesFromW(150, 50, 150, 200)), []pr.Float{200, 400}},
		{fmt.Sprintf(`@top-left { content: %s }
        @top-right { content: ""; width: 200px }
     `, imagesFromW(150, 50, 450)), []pr.Float{450, 200}},
		{fmt.Sprintf(`@top-left { content: %s }
        @top-right { content: %s }
     `, imagesFromW(150, 100), imagesFromW(10, 120)), []pr.Float{250, 130}},
		{fmt.Sprintf(`@top-left { content: %s }
        @top-right { content: %s }
     `, imagesFromW(550, 100), imagesFromW(10, 120)), []pr.Float{550, 120}},
		{fmt.Sprintf(`@top-left { content: %s }
        @top-right { content: %s }
     `, imagesFromW(250, 60), imagesFromW(250, 180)), []pr.Float{275, 325}}, // 250 + (100 * 1 / 4), 250 + (100 * 3 / 4)
	} {
		testPageStyle(t, data.css, data.widths)
	}
}

func testPageStyle(t *testing.T, css string, widths []pr.Float) {
	var expectedAtKeywords []string
	for _, atKeyword := range []string{"@top-left", "@top-center", "@top-right"} {
		if strings.Contains(css, atKeyword+" { content: ") {
			expectedAtKeywords = append(expectedAtKeywords, atKeyword)
		}
	}
	page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @page {
          size: 800px;
          margin: 100px;
          padding: 42px;
          border: 7px solid;
          %s
        }
      </style>
    `, css))
	tu.AssertEqual(t, unpack1(page).Box().ElementTag(), "html")
	marginBoxes := page.Box().Children[1:]
	tu.AssertEqual(t, len(marginBoxes), len(widths))
	var gotAtKeywords []string
	for _, box := range marginBoxes {
		gotAtKeywords = append(gotAtKeywords, box.(*bo.MarginBox).AtKeyword)
	}
	tu.AssertEqual(t, gotAtKeywords, expectedAtKeywords)

	offsets := map[string]pr.Float{"@top-left": 0, "@top-center": 0.5, "@top-right": 1}
	for i, box := range marginBoxes {
		tu.AssertEqual(t, box.Box().MarginWidth(), widths[i])
		tu.AssertEqual(t, box.Box().PositionX, 100+offsets[box.(*bo.MarginBox).AtKeyword]*(600-box.Box().MarginWidth()))
	}
}

func TestMarginBoxesVerticalAlign(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// 3 px ->    +-----+
	//            |  1  |
	//            +-----+
	//
	//        43 px ->   +-----+
	//        53 px ->   |  2  |
	//                   +-----+
	//
	//               83 px ->   +-----+
	//                          |  3  |
	//               103px ->   +-----+
	page := renderOnePage(t, `
      <style>
        @page {
          size: 800px;
          margin: 106px;  /* margin boxes’ content height is 100px */
 
          @top-left {
            content: "foo"; line-height: 20px; border: 3px solid;
            vertical-align: top;
          }
          @top-center {
            content: "foo"; line-height: 20px; border: 3px solid;
            vertical-align: middle;
          }
          @top-right {
            content: "foo"; line-height: 20px; border: 3px solid;
            vertical-align: bottom;
          }
        }
      </style>
    `)
	_, topLeft, topCenter, topRight := unpack4(page)
	line1 := unpack1(topLeft)
	line2 := unpack1(topCenter)
	line3 := unpack1(topRight)
	tu.AssertEqual(t, line1.Box().PositionY, Fl(3))
	tu.AssertEqual(t, line2.Box().PositionY, Fl(43))
	tu.AssertEqual(t, line3.Box().PositionY, Fl(83))
}

func textFromBoxes(boxes []Box) string {
	var s strings.Builder
	for _, box := range boxes {
		if box, ok := box.(*bo.TextBox); ok {
			s.WriteString(box.TextS())
		}
	}
	return s.String()
}

func TestMarginBoxesElement(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page {
          counter-increment: count;
          counter-reset: page pages;
          margin: 50px;
          size: 200px;
          @bottom-center {
            content: counter(page) ' of ' counter(pages)
                     ' (' counter(count) ')';
          }
        }
        h1 {
          height: 40px;
        }
      </style>
      <h1>test1</h1>
      <h1>test2</h1>
      <h1>test3</h1>
      <h1>test4</h1>
      <h1>test5</h1>
      <h1>test6</h1>
    `)
	footer1Text := textFromBoxes(bo.Descendants(pages[0].Box().Children[1]))
	tu.AssertEqual(t, footer1Text, "0 of 3 (1)")

	footer2Text := textFromBoxes(bo.Descendants(pages[1].Box().Children[1]))
	tu.AssertEqual(t, footer2Text, "0 of 3 (2)")

	footer3Text := textFromBoxes(bo.Descendants(pages[2].Box().Children[1]))
	tu.AssertEqual(t, footer3Text, "0 of 3 (3)")
}

func TestMarginBoxesRunningElement(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        footer {
          position: running(footer);
        }
        @page {
          margin: 50px;
          size: 200px;
          @bottom-center {
            content: element(footer);
          }
        }
		body {
			font-size: 1px
		}
        h1 {
          height: 40px;
        }
        .pages:before {
          content: counter(page);
        }
        .pages:after {
          content: counter(pages);
        }
      </style>
      <footer class="pages"> of </footer>
      <h1>test1</h1>
      <h1>test2</h1>
      <h1>test3</h1>
      <h1>test4</h1>
      <h1>test5</h1>
      <h1>test6</h1>
      <footer>Static</footer>
    `)

	var footer1Text string
	for _, node := range bo.Descendants(pages[0].Box().Children[1]) {
		if node, ok := node.(*bo.TextBox); ok {
			footer1Text += node.TextS()
		}
	}
	tu.AssertEqual(t, footer1Text, "1 of 3")

	var footer2Text string
	for _, node := range bo.Descendants(pages[1].Box().Children[1]) {
		if node, ok := node.(*bo.TextBox); ok {
			footer2Text += node.TextS()
		}
	}
	tu.AssertEqual(t, footer2Text, "2 of 3")

	var footer3Text string
	for _, node := range bo.Descendants(pages[2].Box().Children[1]) {
		if node, ok := node.(*bo.TextBox); ok {
			footer3Text += node.TextS()
		}
	}
	tu.AssertEqual(t, footer3Text, "Static")
}

func TestRunningElements(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range []struct {
		argument string
		texts    [5]string
	}{
		// TODO: start doesn’t work because running elements are removed from the
		// original tree, and the current implentation in
		// layout.getRunningElementFor uses the tree to know if it’s at the
		// beginning of the page
		// ("start", ("", "2-first", "2-last", "3-last", "5")),

		{"first", [5]string{"", "2-first", "3-first", "3-last", "5"}},
		{"last", [5]string{"", "2-last", "3-last", "3-last", "5"}},
		{"first-except", [5]string{"", "", "", "3-last", ""}},
	} {
		pages := renderPages(t, fmt.Sprintf(`
		<style>
			@page {
			margin: 50px;
			size: 200px;
			@bottom-center { content: element(title %s) }
			}
			article { break-after: page }
			h1 { position: running(title) }
		</style>
		<article>
			<div>1</div>
		</article>
		<article>
			<h1>2-first</h1>
			<h1>2-last</h1>
		</article>
		<article>
			<p>3</p>
			<h1>3-first</h1>
			<h1>3-last</h1>
		</article>
		<article>
		</article>
		<article>
			<h1>5</h1>
		</article>
		`, data.argument))
		tu.AssertEqual(t, len(pages), 5)
		for i, page := range pages {
			text := data.texts[i]
			_, margin := unpack2(page)
			if len(margin.Box().Children) != 0 {
				h1 := unpack1(margin)
				line := unpack1(h1)
				textbox := unpack1(line)
				assertText(t, textbox, text)
			} else {
				tu.AssertEqual(t, text, "")
			}
		}
	}
}

func TestRunningElementsDisplay(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page {
          margin: 50px;
          size: 200px;
          @bottom-left { content: element(inline) }
          @bottom-center { content: element(block) }
          @bottom-right { content: element(table) }
        }
        table { position: running(table) }
        div { position: running(block) }
        span { position: running(inline) }
      </style>
      text
      <table><tr><td>table</td></tr></table>
      <div>block</div>
      <span>inline</span>
    `)
	_, left, center, right := unpack4(page)
	var leftT, centerT, rightT string
	for _, node := range bo.Descendants(left) {
		if node, ok := node.(*bo.TextBox); ok {
			leftT += node.TextS()
		}
	}
	for _, node := range bo.Descendants(center) {
		if node, ok := node.(*bo.TextBox); ok {
			centerT += node.TextS()
		}
	}
	for _, node := range bo.Descendants(right) {
		if node, ok := node.(*bo.TextBox); ok {
			rightT += node.TextS()
		}
	}
	tu.AssertEqual(t, leftT, "inline")
	tu.AssertEqual(t, centerT, "block")
	tu.AssertEqual(t, rightT, "table")
}

func TestNoNewPage(t *testing.T) {
	pages := renderPages(t, `           
	<style>
		@page { size: 300px 30px }
		body { margin: 0; background: #fff }
	</style>
	<p><a href="another url"><span>[some url] </span>some content</p>
	`)
	if len(pages) != 1 {
		t.Fatalf("expected one page, got %d", len(pages))
	}
}

func TestRunningImg(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test regression
	_ = renderPages(t, `
      <style>
        img {
          position: running(img);
        }
        @page {
          @bottom-center {
            content: element(img);
          }
        }
      </style>
      <img src="pattern.png" />
    `)
}

func TestRunningAbsolute(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for #1540
	_ = renderPages(t, `
      <style>
        footer {
          position: running(footer);
        }
        p {
          position: absolute;
        }
        @page {
          @bottom-center {
            content: element(footer);
          }
        }
      </style>
      <footer>Hello!<p>Bonjour!</p></footer>
    `)
}

func TestRunningFlex(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test regression
	_ = renderPages(t, `
      <style>
        footer {
          display: flex;
          position: running(footer);
        }
        @page {
          @bottom-center {
            content: element(footer);
          }
        }
      </style>
      <footer>
        Hello!
      </footer>
    `)
}

func TestRunningFloat(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test regression
	_ = renderPages(t, `
      <style>
        footer {
          float: left;
          position: running(footer);
        }
        @page {
          @bottom-center {
            content: element(footer);
          }
        }
      </style>
      <footer>
        Hello!
      </footer>
    `)
}
