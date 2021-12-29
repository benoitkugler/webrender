package layout

import (
	"fmt"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/css/properties"
	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

type fl = properties.Fl

// Return the (x, y, w, h) rectangle for the outer area of a box.
func outerArea(box Box) [4]fl {
	return [4]fl{
		fl(box.Box().PositionX), fl(box.Box().PositionY),
		fl(box.Box().MarginWidth()), fl(box.Box().MarginHeight()),
	}
}

func TestFloats1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// adjacent-floats-001
	page := renderOnePage(t, `
      <style>
        div { float: left }
        img { width: 100px; vertical-align: top }
      </style>
      <div><img src=pattern.png /></div>
      <div><img src=pattern.png /></div>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div1, div2 := body.Box().Children[0], body.Box().Children[1]
	tu.AssertEqual(t, outerArea(div1), [4]fl{0, 0, 100, 100}, "div1")
	tu.AssertEqual(t, outerArea(div2), [4]fl{100, 0, 100, 100}, "div2")
}

func TestFloats2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// c414-flt-fit-000
	page := renderOnePage(t, `
      <style>
        body { width: 290px }
        div { float: left; width: 100px;  }
        img { width: 60px; vertical-align: top }
      </style>
      <div><img src=pattern.png /><!-- 1 --></div>
      <div><img src=pattern.png /><!-- 2 --></div>
      <div><img src=pattern.png /><!-- 4 --></div>
      <img src=pattern.png /><!-- 3
      --><img src=pattern.png /><!-- 5 -->`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div1, div2, div4, anonBlock := unpack4(body)
	line3, line5 := anonBlock.Box().Children[0], anonBlock.Box().Children[1]
	img3 := line3.Box().Children[0]
	img5 := line5.Box().Children[0]
	tu.AssertEqual(t, outerArea(div1), [4]fl{0, 0, 100, 60}, "div1")
	tu.AssertEqual(t, outerArea(div2), [4]fl{100, 0, 100, 60}, "div2")
	tu.AssertEqual(t, outerArea(img3), [4]fl{200, 0, 60, 60}, "img3")

	tu.AssertEqual(t, outerArea(div4), [4]fl{0, 60, 100, 60}, "div4")
	tu.AssertEqual(t, outerArea(img5), [4]fl{100, 60, 60, 60}, "img5")
}

func TestFloats3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// c414-flt-fit-002
	page := renderOnePage(t, `
      <style type="text/css">
        body { width: 200px }
        p { width: 70px; height: 20px }
        .left { float: left }
        .right { float: right }
      </style>
      <p class="left"> ⇦ A 1 </p>
      <p class="left"> ⇦ B 2 </p>
      <p class="left"> ⇦ A 3 </p>
      <p class="right"> B 4 ⇨ </p>
      <p class="left"> ⇦ A 5 </p>
      <p class="right"> B 6 ⇨ </p>
      <p class="right"> B 8 ⇨ </p>
      <p class="left"> ⇦ A 7 </p>
      <p class="left"> ⇦ A 9 </p>
      <p class="left"> ⇦ B 10 </p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	var positions [][2]fl
	for _, paragraph := range body.Box().Children {
		positions = append(positions, [2]fl{fl(paragraph.Box().PositionX), fl(paragraph.Box().PositionY)})
	}
	tu.AssertEqual(t, positions, [][2]fl{
		{0, 0},
		{70, 0},
		{0, 20},
		{130, 20},
		{0, 40},
		{130, 40},
		{130, 60},
		{0, 60},
		{0, 80},
		{70, 80},
	}, "")
}

func TestFloats4(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// c414-flt-wrap-000 ... more || less
	page := renderOnePage(t, `
      <style>
        body { width: 100px }
        p { float: left; height: 100px }
        img { width: 60px; vertical-align: top }
      </style>
      <p style="width: 20px"></p>
      <p style="width: 100%"></p>
      <img src=pattern.png /><img src=pattern.png />
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	_, _, anonBlock := unpack3(body)
	line1, line2 := anonBlock.Box().Children[0], anonBlock.Box().Children[1]
	tu.AssertEqual(t, fl(anonBlock.Box().PositionY), fl(0), "anonBlock")
	tu.AssertEqual(t, [2]fl{fl(line1.Box().PositionX), fl(line1.Box().PositionY)}, [2]fl{20, 0}, "line1")
	tu.AssertEqual(t, [2]fl{fl(line2.Box().PositionX), fl(line2.Box().PositionY)}, [2]fl{0, 200}, "line2")
}

func TestFloats5(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// c414-flt-wrap-000 with text ... more || less
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body { width: 100px; font: 60px weasyprint; }
        p { float: left; height: 100px }
        img { width: 60px; vertical-align: top }
      </style>
      <p style="width: 20px"></p>
      <p style="width: 100%"></p>
      A B
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	_, _, anonBlock := unpack3(body)
	line1, line2 := anonBlock.Box().Children[0], anonBlock.Box().Children[1]
	tu.AssertEqual(t, fl(anonBlock.Box().PositionY), fl(0), "anonBlock")
	tu.AssertEqual(t, [2]fl{fl(line1.Box().PositionX), fl(line1.Box().PositionY)}, [2]fl{20, 0}, "line1")
	tu.AssertEqual(t, [2]fl{fl(line2.Box().PositionX), fl(line2.Box().PositionY)}, [2]fl{0, 200}, "line2")
}

func TestFloats6(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// floats-placement-vertical-001b
	page := renderOnePage(t, `
      <style>
        body { width: 90px; font-size: 0 }
        img { vertical-align: top }
      </style>
      <body>
      <span>
        <img src=pattern.png style="width: 50px" />
        <img src=pattern.png style="width: 50px" />
        <img src=pattern.png style="float: left; width: 30px" />
      </span>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line1, line2 := body.Box().Children[0], body.Box().Children[1]
	span1 := line1.Box().Children[0]
	span2 := line2.Box().Children[0]
	img1 := span1.Box().Children[0]
	img2, img3 := span2.Box().Children[0], span2.Box().Children[1]
	tu.AssertEqual(t, outerArea(img1), [4]fl{0, 0, 50, 50}, "img1")
	tu.AssertEqual(t, outerArea(img2), [4]fl{30, 50, 50, 50}, "img2")
	tu.AssertEqual(t, outerArea(img3), [4]fl{0, 50, 30, 30}, "img3")
}

func TestFloats7(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Variant of the above: no <span>
	page := renderOnePage(t, `
      <style>
        body { width: 90px; font-size: 0 }
        img { vertical-align: top }
      </style>
      <body>
      <img src=pattern.png style="width: 50px" />
      <img src=pattern.png style="width: 50px" />
      <img src=pattern.png style="float: left; width: 30px" />
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line1, line2 := body.Box().Children[0], body.Box().Children[1]
	img1 := line1.Box().Children[0]
	img2, img3 := line2.Box().Children[0], line2.Box().Children[1]
	tu.AssertEqual(t, outerArea(img1), [4]fl{0, 0, 50, 50}, "img1")
	tu.AssertEqual(t, outerArea(img2), [4]fl{30, 50, 50, 50}, "img2")
	tu.AssertEqual(t, outerArea(img3), [4]fl{0, 50, 30, 30}, "img3")
}

func TestFloats8(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Floats do no affect other pages
	pages := renderPages(t, `
      <style>
        body { width: 90px; font-size: 0 }
        img { vertical-align: top }
      </style>
      <body>
      <img src=pattern.png style="float: left; width: 30px" />
      <img src=pattern.png style="width: 50px" />
      <div style="page-break-before: always"></div>
      <img src=pattern.png style="width: 50px" />
    `)
	page1, page2 := pages[0], pages[1]

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	floatImg, anonBlock := body.Box().Children[0], body.Box().Children[1]
	line := anonBlock.Box().Children[0]
	img1 := line.Box().Children[0]
	tu.AssertEqual(t, outerArea(floatImg), [4]fl{0, 0, 30, 30}, "floatImg")
	tu.AssertEqual(t, outerArea(img1), [4]fl{30, 0, 50, 50}, "img1")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	_, anonBlock = body.Box().Children[0], body.Box().Children[1]
	line = anonBlock.Box().Children[0]
	_ = line.Box().Children[0]
}

func TestFloats9(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Regression test
	// https://github.com/Kozea/WeasyPrint/issues/263
	_ = renderOnePage(t, `<div style="top:100%; float:left">`)
}

func TestFloatsPageBreaks1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Tests floated images shorter than the page
	pages := renderPages(t, `
      <style>
        @page { size: 100px; margin: 10px }
        img { height: 45px; width:70px; float: left;}
      </style>
      <body>
        <img src=pattern.png>
          <!-- page break should be here !!! -->
        <img src=pattern.png>
    `)

	tu.AssertEqual(t, len(pages), 2, "number of pages")

	var pageImagesPosY [][]pr.Float
	for _, page := range pages {
		var images []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "img" {
				images = append(images, d.Box().PositionY)
				tu.AssertEqual(t, d.Box().PositionX, pr.Float(10), "img")
			}
		}
		pageImagesPosY = append(pageImagesPosY, images)
	}
	tu.AssertEqual(t, pageImagesPosY, [][]pr.Float{{10}, {10}}, "")
}

func TestFloatsPageBreaks2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Tests floated images taller than the page
	pages := renderPages(t, `
      <style>
        @page { size: 100px; margin: 10px }
        img { height: 81px; width:70px; float: left;}
      </style>
      <body>
        <img src=pattern.png>
          <!-- page break should be here !!! -->
        <img src=pattern.png>
    `)

	tu.AssertEqual(t, len(pages), 2, "")

	var pageImagesPosY [][]pr.Float
	for _, page := range pages {
		var images []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "img" {
				images = append(images, d.Box().PositionY)
				tu.AssertEqual(t, d.Box().PositionX, pr.Float(10), "img")
			}
		}
		pageImagesPosY = append(pageImagesPosY, images)
	}
	tu.AssertEqual(t, pageImagesPosY, [][]pr.Float{{10}, {10}}, "")
}

func TestFloatsPageBreaks3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Tests floated images shorter than the page
	pages := renderPages(t, `
      <style>
        @page { size: 100px; margin: 10px }
        img { height: 30px; width:70px; float: left;}
      </style>
      <body>
        <img src=pattern.png>
        <img src=pattern.png>
          <!-- page break should be here !!! -->
        <img src=pattern.png>
        <img src=pattern.png>
          <!-- page break should be here !!! -->
        <img src=pattern.png>
    `)

	tu.AssertEqual(t, len(pages), 3, "")

	var pageImagesPosY [][]pr.Float
	for _, page := range pages {
		var images []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "img" {
				images = append(images, d.Box().PositionY)
				tu.AssertEqual(t, d.Box().PositionX, pr.Float(10), "img")
			}
		}
		pageImagesPosY = append(pageImagesPosY, images)
	}
	tu.AssertEqual(t, pageImagesPosY, [][]pr.Float{{10, 40}, {10, 40}, {10}}, "")
}

func TestFloatsPageBreaks4(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// last float does not fit, pushed to next page
	pages := renderPages(t, `
      <style>
        @page{
          size: 110px;
          margin: 10px;
          padding: 0;
        }
        .large {
          width: 10px;
          height: 60px;
        }
        .small {
          width: 10px;
          height: 20px;
        }
      </style>
      <body>
        <div class="large"></div>
        <div class="small"></div>
        <div class="large"></div>
    `)

	tu.AssertEqual(t, len(pages), 2, "number of pages")

	var pageDivPosY [][]pr.Float
	for _, page := range pages {
		var images []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "div" {
				images = append(images, d.Box().PositionY)
			}
		}
		pageDivPosY = append(pageDivPosY, images)
	}
	tu.AssertEqual(t, pageDivPosY, [][]pr.Float{{10, 70}, {10}}, "")
}

func TestFloatsPageBreaks5(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// last float does not fit, pushed to next page
	// center div must not
	pages := renderPages(t, `
      <style>
        @page{
          size: 110px;
          margin: 10px;
          padding: 0;
        }
        .large {
          width: 10px;
          height: 60px;
        }
        .small {
          width: 10px;
          height: 20px;
          page-break-after: avoid;
        }
      </style>
      <body>
        <div class="large"></div>
        <div class="small"></div>
        <div class="large"></div>
    `)

	tu.AssertEqual(t, len(pages), 2, "number of pages")

	var pageDivPosY [][]pr.Float
	for _, page := range pages {
		var images []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "div" {
				images = append(images, d.Box().PositionY)
			}
		}
		pageDivPosY = append(pageDivPosY, images)
	}
	tu.AssertEqual(t, pageDivPosY, [][]pr.Float{{10}, {10, 30}}, "")
}

func TestFloatsPageBreaks6(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// center div must be the last element,
	// but float won't fit and will get pushed anyway
	pages := renderPages(t, `
      <style>
        @page{
          size: 110px;
          margin: 10px;
          padding: 0;
        }
        .large {
          width: 10px;
          height: 80px;
        }
        .small {
          width: 10px;
          height: 20px;
          page-break-after: avoid;
        }
      </style>
      <body>
        <div class="large"></div>
        <div class="small"></div>
        <div class="large"></div>
    `)

	tu.AssertEqual(t, len(pages), 3, "")
	var pageDivPosY [][]pr.Float
	for _, page := range pages {
		var images []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "div" {
				images = append(images, d.Box().PositionY)
			}
		}
		pageDivPosY = append(pageDivPosY, images)
	}
	tu.AssertEqual(t, pageDivPosY, [][]pr.Float{{10}, {10}, {10}}, "")
}

func TestPreferredWidths1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	getFloatWidth := func(bodyWidth int) pr.Float {
		page := renderOnePage(t, fmt.Sprintf(`
          <style>
            @font-face { src: url(weasyprint.otf); font-family: weasyprint }
          </style>
          <body style="width: %dpx; font-family: weasyprint">
          <p style="white-space: pre-line; float: left">
            Lorem ipsum dolor sit amet,
              consectetur elit
          </p>
                   <!--  ^  No-break space here  -->
        `, bodyWidth))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		paragraph := body.Box().Children[0]
		return paragraph.Box().Width.V()
	}
	// Preferred minimum width:
	tu.AssertEqual(t, getFloatWidth(10), pr.Float(len([]rune("consectetur elit"))*16), "10")
	// Preferred width:
	tu.AssertEqual(t, getFloatWidth(1000000), pr.Float(len([]rune("Lorem ipsum dolor sit amet,"))*16), "1000000")
}

func TestPreferredWidths2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Non-regression test:
	// Incorrect whitespace handling in preferred width used to cause
	// unnecessary line break.
	page := renderOnePage(t, `
      <p style="float: left">Lorem <em>ipsum</em> dolor.</p>
    } `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	tu.AssertEqual(t, len(paragraph.Box().Children), 1, "")
	tu.AssertEqual(t, bo.LineBoxT.IsInstance(paragraph.Box().Children[0]), true, "")
}

func TestPreferredWidths3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>img { width: 20px }</style>
      <p style="float: left">
        <img src=pattern.png><img src=pattern.png><br>
        <img src=pattern.png></p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	tu.AssertEqual(t, paragraph.Box().Width, pr.Float(40), "")
}

func TestPreferredWidths4(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
          @font-face { src: url(weasyprint.otf); font-family: weasyprint }
          p { font: 20px weasyprint }
        </style>
        <p style="float: left">XX<br>XX<br>X</p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	tu.AssertEqual(t, paragraph.Box().Width, pr.Float(40), "")
}

func TestPreferredWidths5(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// The space is the start of the line is collapsed.
	page := renderOnePage(t, `
        <style>
          @font-face { src: url(weasyprint.otf); font-family: weasyprint }
          p { font: 20px weasyprint }
        </style>
        <p style="float: left">XX<br> XX<br>X</p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	tu.AssertEqual(t, paragraph.Box().Width, pr.Float(40), "")
}

func TestFloatInInline1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body {
          font-family: weasyprint;
          font-size: 20px;
        }
        p {
          width: 14em;
          text-align: justify;
        }
        span {
          float: right;
        }
      </style>
      <p>
        aa bb <a><span>cc</span> ddd</a> ee ff
      </p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2 := paragraph.Box().Children[0], paragraph.Box().Children[1]

	p1, a, p2 := unpack3(line1)
	tu.AssertEqual(t, p1.Box().Width, pr.Float(6*20), "p1.width")
	tu.AssertEqual(t, p1.(*bo.TextBox).Text, "aa bb ", "p1.text")
	tu.AssertEqual(t, p1.Box().PositionX, pr.Float(0*20), "p1.positionX")
	tu.AssertEqual(t, p2.Box().Width, pr.Float(3*20), "p2.width")
	tu.AssertEqual(t, p2.(*bo.TextBox).Text, " ee", "p2.text")
	tu.AssertEqual(t, p2.Box().PositionX, pr.Float(9*20), "p2.positionX")
	span, aText := a.Box().Children[0], a.Box().Children[1]
	tu.AssertEqual(t, aText.Box().Width, pr.Float(3*20), "") // leading space collapse)
	tu.AssertEqual(t, aText.(*bo.TextBox).Text, "ddd", "")
	tu.AssertEqual(t, aText.Box().PositionX, pr.Float(6*20), "aText")
	tu.AssertEqual(t, span.Box().Width, pr.Float(2*20), "span")
	tu.AssertEqual(t, span.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "cc", "span")
	tu.AssertEqual(t, span.Box().PositionX, pr.Float(12*20), "span")

	p3 := line2.Box().Children[0]
	tu.AssertEqual(t, p3.Box().Width, pr.Float(2*20), "")
}

func TestFloatInInline_2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page {
          size: 10em;
        }
        article {
          font-family: weasyprint;
          line-height: 1;
        }
        div {
          float: left;
          width: 50%;
        }
      </style>
      <article>
        <span>
          <div>a b c</div>
          1 2 3 4 5 6
        </span>
      </article>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	line1, line2 := unpack2(article)
	span1 := line1.Box().Children[0]
	div, text := unpack2(span1)
	tu.AssertEqual(t, strings.TrimSpace(div.Box().Children[0].Box().Children[0].(*bo.TextBox).Text), "a b c", "")
	tu.AssertEqual(t, strings.TrimSpace(text.(*bo.TextBox).Text), "1 2 3", "")
	span2 := line2.Box().Children[0]
	text = span2.Box().Children[0]
	tu.AssertEqual(t, strings.TrimSpace(text.(*bo.TextBox).Text), "4 5 6", "")
}

func TestFloatInInline_3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page {
          size: 10em;
        }
        article {
          font-family: weasyprint;
          line-height: 1;
        }
        div {
          float: left;
          width: 50%;
        }
      </style>
      <article>
        <span>
          1 2 3 <div>a b c</div> 4 5 6
        </span>
      </article>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	line1, line2 := unpack2(article)
	span1 := line1.Box().Children[0]
	text, div := unpack2(span1)
	tu.AssertEqual(t, strings.TrimSpace(text.(*bo.TextBox).Text), "1 2 3", "")
	tu.AssertEqual(t, strings.TrimSpace(div.Box().Children[0].Box().Children[0].(*bo.TextBox).Text), "a b c", "")
	span2 := line2.Box().Children[0]
	text = span2.Box().Children[0]
	tu.AssertEqual(t, strings.TrimSpace(text.(*bo.TextBox).Text), "4 5 6", "")
}

func TestFloatInInline_4(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page {
          size: 10em;
        }
        article {
          font-family: weasyprint;
          line-height: 1;
        }
        div {
          float: left;
          width: 50%;
        }
      </style>
      <article>
        <span>
          1 2 3 4 <div>a b c</div> 5 6
        </span>
      </article>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	line1, line2 := unpack2(article)
	span1, div := unpack2(line1)
	text1, text2 := unpack2(span1)
	tu.AssertEqual(t, strings.TrimSpace(text1.(*bo.TextBox).Text), "1 2 3 4", "")
	tu.AssertEqual(t, strings.TrimSpace(text2.(*bo.TextBox).Text), "5", "")
	tu.AssertEqual(t, div.Box().PositionY, pr.Float(16), "")
	tu.AssertEqual(t, strings.TrimSpace(div.Box().Children[0].Box().Children[0].(*bo.TextBox).Text), "a b c", "")
	span2 := line2.Box().Children[0]
	text := span2.Box().Children[0]
	tu.AssertEqual(t, strings.TrimSpace(text.(*bo.TextBox).Text), "6", "")
}

func TestFloatNextLine(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body {
          font-family: weasyprint;
          font-size: 20px;
        }
        p {
          text-align: justify;
          width: 13em;
        }
        span {
          float: left;
        }
      </style>
      <p>pp pp pp pp <a><span>ppppp</span> aa</a> pp pp pp pp pp</p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2, line3 := unpack3(paragraph)
	tu.AssertEqual(t, len(line1.Box().Children), 1, "len")
	tu.AssertEqual(t, len(line3.Box().Children), 1, "len")
	a, p := line2.Box().Children[0], line2.Box().Children[1]
	span, aText := a.Box().Children[0], a.Box().Children[1]
	tu.AssertEqual(t, span.Box().PositionX, pr.Float(0), "span")
	tu.AssertEqual(t, span.Box().Width, pr.Float(5*20), "span")
	tu.AssertEqual(t, aText.Box().PositionX, pr.Float(5*20), "aText")
	tu.AssertEqual(t, a.Box().PositionX, pr.Float(5*20), "a")
	tu.AssertEqual(t, aText.Box().Width, pr.Float(2*20), "aText")
	tu.AssertEqual(t, a.Box().Width, pr.Float(2*20), "a")
	tu.AssertEqual(t, p.Box().PositionX, pr.Float(7*20), "p")
}

func TestFloatTextIndent1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body {
          font-family: weasyprint;
          font-size: 20px;
        }
        p {
          text-align: justify;
          text-indent: 1em;
          width: 14em;
        }
        span {
          float: left;
        }
      </style>
      <p><a>aa <span>float</span> aa</a></p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1 := paragraph.Box().Children[0]
	a := line1.Box().Children[0]
	a1, span, a2 := unpack3(a)
	spanText := span.Box().Children[0]
	tu.AssertEqual(t, span.Box().PositionX, pr.Float(0), "span")
	tu.AssertEqual(t, spanText.Box().PositionX, pr.Float(0), "spanText")
	tu.AssertEqual(t, span.Box().Width, pr.Float((1+5)*20), "span")         // text-indent + span text
	tu.AssertEqual(t, spanText.Box().Width, pr.Float((1+5)*20), "spanText") // text-indent + span text
	tu.AssertEqual(t, a1.Box().Width, pr.Float(3*20), "a1")
	tu.AssertEqual(t, a1.Box().PositionX, pr.Float((1+5+1)*20), "a1")   // span + a1 text-indent)
	tu.AssertEqual(t, a2.Box().Width, pr.Float(2*20), "a2")             // leading space collapse)
	tu.AssertEqual(t, a2.Box().PositionX, pr.Float((1+5+1+3)*20), "a2") // span + a1 t-i + a1)
}

func TestFloatTextIndent2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body {
          font-family: weasyprint;
          font-size: 20px;
        }
        p {
          text-align: justify;
          text-indent: 1em;
          width: 14em;
        }
        span {
          float: left;
        }
      </style>
      <p>
        oooooooooooo
        <a>aa <span>float</span> aa</a></p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2 := paragraph.Box().Children[0], paragraph.Box().Children[1]

	p1 := line1.Box().Children[0]
	tu.AssertEqual(t, p1.Box().PositionX, pr.Float(1*20), "p1") // text-indent
	tu.AssertEqual(t, p1.Box().Width, pr.Float(12*20), "p1")    // p text

	a := line2.Box().Children[0]
	a1, span, a2 := unpack3(a)
	spanText := span.Box().Children[0]
	tu.AssertEqual(t, span.Box().PositionX, pr.Float(0), "span")
	tu.AssertEqual(t, spanText.Box().PositionX, pr.Float(0), " spanText")
	tu.AssertEqual(t, span.Box().Width, pr.Float((1+5)*20), "span")           // text-indent + span text
	tu.AssertEqual(t, spanText.Box().Width, pr.Float((1+5)*20), "  spanText") // text-indent + span text
	tu.AssertEqual(t, a1.Box().Width, pr.Float(3*20), "a1")
	tu.AssertEqual(t, a1.Box().PositionX, pr.Float((1+5)*20), "a1")   // span)
	tu.AssertEqual(t, a2.Box().Width, pr.Float(2*20), "a2")           // leading space collapse)
	tu.AssertEqual(t, a2.Box().PositionX, pr.Float((1+5+3)*20), "a2") // span + a1)
}

func TestFloatTextIndent3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body {
          font-family: weasyprint;
          font-size: 20px;
        }
        p {
          text-align: justify;
          text-indent: 1em;
          width: 14em;
        }
        span {
          float: right;
        }
      </style>
      <p>
        oooooooooooo
        <a>aa <span>float</span> aa</a>
        oooooooooooo
      </p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2, line3 := unpack3(paragraph)

	p1 := line1.Box().Children[0]
	tu.AssertEqual(t, p1.Box().PositionX, pr.Float(1*20), " p1") // text-indent)
	tu.AssertEqual(t, p1.Box().Width, pr.Float(12*20), " p1")    // p text)

	a := line2.Box().Children[0]
	a1, span, a2 := unpack3(a)
	spanText := span.Box().Children[0]
	tu.AssertEqual(t, span.Box().PositionX, pr.Float((14-5-1)*20), " span")
	tu.AssertEqual(t, spanText.Box().PositionX, pr.Float((14-5-1)*20), "   spanText")
	tu.AssertEqual(t, span.Box().Width, pr.Float((1+5)*20), " span")           // text-indent + span text
	tu.AssertEqual(t, spanText.Box().Width, pr.Float((1+5)*20), "   spanText") // text-indent + span text
	tu.AssertEqual(t, a1.Box().PositionX, pr.Float(0), " a1")                  // span)
	tu.AssertEqual(t, a2.Box().Width, pr.Float(2*20), " a2")                   // leading space collapse)
	tu.AssertEqual(t, a2.Box().PositionX, pr.Float((14-5-1-2)*20), " a2")

	p2 := line3.Box().Children[0]
	tu.AssertEqual(t, p2.Box().PositionX, pr.Float(0), " p2")
	tu.AssertEqual(t, p2.Box().Width, pr.Float(12*20), " p2") // p text)
}

// @pytest.mark.xfail
// func TestFloatFail(t *testing.T) {
//   cp := tu.CaptureLogs()
//   defer cp.AssertNoLogs(t)
//     page := renderOnePage(t,`
//       <style>
//         @font-face { src: url(weasyprint.otf); font-family: weasyprint }
//         body {
//           font-family: weasyprint;
//           font-size: 20px;
//         }
//         p {
//           text-align: justify;
//           width: 12em;
//         }
//         span {
//           float: left;
//           background: red;
//         }
//         a {
//           background: yellow;
//         }
//       </style>
//       <p>bb bb pp bb pp pb <a><span>pp pp</span> apa</a> bb bb</p>`)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     line1, line2, line3 = paragraph.Box().Children
