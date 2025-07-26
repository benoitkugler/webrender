package layout

import (
	"fmt"
	"strings"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

type fl = pr.Fl

// Return the (x, y, w, h) rectangle for the outer area of a box.
func outerArea(box Box) [4]fl {
	return [4]fl{
		fl(box.Box().PositionX), fl(box.Box().PositionY),
		fl(box.Box().MarginWidth()), fl(box.Box().MarginHeight()),
	}
}

func TestFloats1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// adjacent-floats-001
	page := renderOnePage(t, `
      <style>
        div { float: left }
        img { width: 100px; vertical-align: top }
      </style>
      <div><img src=pattern.png /></div>
      <div><img src=pattern.png /></div>`)
	html := unpack1(page)
	body := unpack1(html)
	div1, div2 := unpack1(body), body.Box().Children[1]
	tu.AssertEqual(t, outerArea(div1), [4]fl{0, 0, 100, 100})
	tu.AssertEqual(t, outerArea(div2), [4]fl{100, 0, 100, 100})
}

func TestFloats2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	div1, div2, div4, anonBlock := unpack4(body)
	line3, line5 := unpack1(anonBlock), anonBlock.Box().Children[1]
	img3 := unpack1(line3)
	img5 := unpack1(line5)
	tu.AssertEqual(t, outerArea(div1), [4]fl{0, 0, 100, 60})
	tu.AssertEqual(t, outerArea(div2), [4]fl{100, 0, 100, 60})
	tu.AssertEqual(t, outerArea(img3), [4]fl{200, 0, 60, 60})

	tu.AssertEqual(t, outerArea(div4), [4]fl{0, 60, 100, 60})
	tu.AssertEqual(t, outerArea(img5), [4]fl{100, 60, 60, 60})
}

func TestFloats3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
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
	})
}

func TestFloats4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	_, _, anonBlock := unpack3(body)
	line1, line2 := unpack1(anonBlock), anonBlock.Box().Children[1]
	tu.AssertEqual(t, fl(anonBlock.Box().PositionY), fl(0))
	tu.AssertEqual(t, [2]fl{fl(line1.Box().PositionX), fl(line1.Box().PositionY)}, [2]fl{20, 0})
	tu.AssertEqual(t, [2]fl{fl(line2.Box().PositionX), fl(line2.Box().PositionY)}, [2]fl{0, 200})
}

func TestFloats5(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
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
	html := unpack1(page)
	body := unpack1(html)
	_, _, anonBlock := unpack3(body)
	line1, line2 := unpack1(anonBlock), anonBlock.Box().Children[1]
	tu.AssertEqual(t, fl(anonBlock.Box().PositionY), fl(0))
	tu.AssertEqual(t, [2]fl{fl(line1.Box().PositionX), fl(line1.Box().PositionY)}, [2]fl{20, 0})
	tu.AssertEqual(t, [2]fl{fl(line2.Box().PositionX), fl(line2.Box().PositionY)}, [2]fl{0, 200})
}

func TestFloats6(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	line1, line2 := unpack1(body), body.Box().Children[1]
	span1 := unpack1(line1)
	span2 := unpack1(line2)
	img1 := unpack1(span1)
	img2, img3 := unpack1(span2), span2.Box().Children[1]
	tu.AssertEqual(t, outerArea(img1), [4]fl{0, 0, 50, 50})
	tu.AssertEqual(t, outerArea(img2), [4]fl{30, 50, 50, 50})
	tu.AssertEqual(t, outerArea(img3), [4]fl{0, 50, 30, 30})
}

func TestFloats7(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	line1, line2 := unpack1(body), body.Box().Children[1]
	img1 := unpack1(line1)
	img2, img3 := unpack1(line2), line2.Box().Children[1]
	tu.AssertEqual(t, outerArea(img1), [4]fl{0, 0, 50, 50})
	tu.AssertEqual(t, outerArea(img2), [4]fl{30, 50, 50, 50})
	tu.AssertEqual(t, outerArea(img3), [4]fl{0, 50, 30, 30})
}

func TestFloats8(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	html := unpack1(page1)
	body := unpack1(html)
	floatImg, anonBlock := unpack1(body), body.Box().Children[1]
	line := unpack1(anonBlock)
	img1 := unpack1(line)
	tu.AssertEqual(t, outerArea(floatImg), [4]fl{0, 0, 30, 30})
	tu.AssertEqual(t, outerArea(img1), [4]fl{30, 0, 50, 50})

	html = unpack1(page2)
	body = unpack1(html)
	_, anonBlock = unpack1(body), body.Box().Children[1]
	line = unpack1(anonBlock)
	_ = unpack1(line)
}

// Regression test for #263.
func TestFloats9(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	_ = renderOnePage(t, `<div style="top:100%; float:left">`)
}

func TestFloatsPageBreaks1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	tu.AssertEqual(t, len(pages), 2)

	var pageImagesPosY [][]pr.Float
	for _, page := range pages {
		var images []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "img" {
				images = append(images, d.Box().PositionY)
				tu.AssertEqual(t, d.Box().PositionX, Fl(10))
			}
		}
		pageImagesPosY = append(pageImagesPosY, images)
	}
	tu.AssertEqual(t, pageImagesPosY, [][]pr.Float{{10}, {10}})
}

func TestFloatsPageBreaks2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	tu.AssertEqual(t, len(pages), 2)

	var pageImagesPosY [][]pr.Float
	for _, page := range pages {
		var images []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "img" {
				images = append(images, d.Box().PositionY)
				tu.AssertEqual(t, d.Box().PositionX, Fl(10))
			}
		}
		pageImagesPosY = append(pageImagesPosY, images)
	}
	tu.AssertEqual(t, pageImagesPosY, [][]pr.Float{{10}, {10}})
}

func TestFloatsPageBreaks3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
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

	tu.AssertEqual(t, len(pages), 3)

	var pageImagesPosY [][]pr.Float
	for _, page := range pages {
		var images []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "img" {
				images = append(images, d.Box().PositionY)
				tu.AssertEqual(t, d.Box().PositionX, Fl(10))
			}
		}
		pageImagesPosY = append(pageImagesPosY, images)
	}
	tu.AssertEqual(t, pageImagesPosY, [][]pr.Float{{10, 40}, {10, 40}, {10}})
}

func TestPreferredWidths1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
		html := unpack1(page)
		body := unpack1(html)
		paragraph := unpack1(body)
		return paragraph.Box().Width.V()
	}
	// Preferred minimum width:
	tu.AssertEqual(t, getFloatWidth(10), Fl(len([]rune("consectetur elit"))*16))
	// Preferred width:
	tu.AssertEqual(t, getFloatWidth(1000000), Fl(len([]rune("Lorem ipsum dolor sit amet,"))*16))
}

func TestPreferredWidths2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Non-regression test:
	// Incorrect whitespace handling in preferred width used to cause
	// unnecessary line break.
	page := renderOnePage(t, `
      <p style="float: left">Lorem <em>ipsum</em> dolor.</p>
    } `)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	tu.AssertEqual(t, len(paragraph.Box().Children), 1)
	tu.AssertEqual(t, bo.LineT.IsInstance(unpack1(paragraph)), true)
}

func TestPreferredWidths3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>img { width: 20px }</style>
      <p style="float: left">
        <img src=pattern.png><img src=pattern.png><br>
        <img src=pattern.png></p>
    `)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	tu.AssertEqual(t, paragraph.Box().Width, Fl(40))
}

func TestPreferredWidths4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
          @font-face { src: url(weasyprint.otf); font-family: weasyprint }
          p { font: 20px weasyprint }
        </style>
        <p style="float: left">XX<br>XX<br>X</p>`)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	tu.AssertEqual(t, paragraph.Box().Width, Fl(40))
}

func TestPreferredWidths5(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// The space is the start of the line is collapsed.
	page := renderOnePage(t, `
        <style>
          @font-face { src: url(weasyprint.otf); font-family: weasyprint }
          p { font: 20px weasyprint }
        </style>
        <p style="float: left">XX<br> XX<br>X</p>`)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	tu.AssertEqual(t, paragraph.Box().Width, Fl(40))
}

func TestFloatInInline1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line1, line2 := unpack1(paragraph), paragraph.Box().Children[1]

	p1, a, p2 := unpack3(line1)
	tu.AssertEqual(t, p1.Box().Width, Fl(6*20))
	assertText(t, p1, "aa bb ")
	tu.AssertEqual(t, p1.Box().PositionX, Fl(0*20))
	tu.AssertEqual(t, p2.Box().Width, Fl(3*20))
	assertText(t, p2, " ee")
	tu.AssertEqual(t, p2.Box().PositionX, Fl(9*20))
	span, aText := unpack1(a), a.Box().Children[1]
	tu.AssertEqual(t, aText.Box().Width, Fl(3*20))
	assertText(t, aText, "ddd")
	tu.AssertEqual(t, aText.Box().PositionX, Fl(6*20))
	tu.AssertEqual(t, span.Box().Width, Fl(2*20))
	assertText(t, unpack1(span.Box().Children[0]), "cc")
	tu.AssertEqual(t, span.Box().PositionX, Fl(12*20))

	p3 := unpack1(line2)
	tu.AssertEqual(t, p3.Box().Width, Fl(2*20))
}

func TestFloatInInline_2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	line1, line2 := unpack2(article)
	span1 := unpack1(line1)
	div, text := unpack2(span1)
	tu.AssertEqual(t, strings.TrimSpace(unpack1(unpack1(div)).(*bo.TextBox).TextS()), "a b c")
	tu.AssertEqual(t, strings.TrimSpace(text.(*bo.TextBox).TextS()), "1 2 3")
	span2 := unpack1(line2)
	text = unpack1(span2)
	tu.AssertEqual(t, strings.TrimSpace(text.(*bo.TextBox).TextS()), "4 5 6")
}

func TestFloatInInline_3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	line1, line2 := unpack2(article)
	span1 := unpack1(line1)
	text, div := unpack2(span1)
	tu.AssertEqual(t, strings.TrimSpace(text.(*bo.TextBox).TextS()), "1 2 3")
	tu.AssertEqual(t, strings.TrimSpace(unpack1(unpack1(div)).(*bo.TextBox).TextS()), "a b c")
	span2 := unpack1(line2)
	text = unpack1(span2)
	tu.AssertEqual(t, strings.TrimSpace(text.(*bo.TextBox).TextS()), "4 5 6")
}

func TestFloatInInline_4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	line1, line2 := unpack2(article)
	span1, div := unpack2(line1)
	text1, text2 := unpack2(span1)
	tu.AssertEqual(t, strings.TrimSpace(text1.(*bo.TextBox).TextS()), "1 2 3 4")
	tu.AssertEqual(t, strings.TrimSpace(text2.(*bo.TextBox).TextS()), "5")
	tu.AssertEqual(t, div.Box().PositionY, Fl(16))
	tu.AssertEqual(t, strings.TrimSpace(unpack1(unpack1(div)).(*bo.TextBox).TextS()), "a b c")
	span2 := unpack1(line2)
	text := unpack1(span2)
	tu.AssertEqual(t, strings.TrimSpace(text.(*bo.TextBox).TextS()), "6")
}

func TestFloatNextLine(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line1, line2, line3 := unpack3(paragraph)
	tu.AssertEqual(t, len(line1.Box().Children), 1)
	tu.AssertEqual(t, len(line3.Box().Children), 1)
	a, p := unpack1(line2), line2.Box().Children[1]
	span, aText := unpack1(a), a.Box().Children[1]
	tu.AssertEqual(t, span.Box().PositionX, Fl(0))
	tu.AssertEqual(t, span.Box().Width, Fl(5*20))
	tu.AssertEqual(t, aText.Box().PositionX, Fl(5*20))
	tu.AssertEqual(t, a.Box().PositionX, Fl(5*20))
	tu.AssertEqual(t, aText.Box().Width, Fl(2*20))
	tu.AssertEqual(t, a.Box().Width, Fl(2*20))
	tu.AssertEqual(t, p.Box().PositionX, Fl(7*20))
}

func TestFloatTextIndent1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line1 := unpack1(paragraph)
	a := unpack1(line1)
	a1, span, a2 := unpack3(a)
	spanText := unpack1(span)
	tu.AssertEqual(t, span.Box().PositionX, Fl(0))
	tu.AssertEqual(t, spanText.Box().PositionX, Fl(0))
	tu.AssertEqual(t, span.Box().Width, Fl((1+5)*20))
	tu.AssertEqual(t, spanText.Box().Width, Fl((1+5)*20))
	tu.AssertEqual(t, a1.Box().Width, Fl(3*20))
	tu.AssertEqual(t, a1.Box().PositionX, Fl((1+5+1)*20))
	tu.AssertEqual(t, a2.Box().Width, Fl(2*20))
	tu.AssertEqual(t, a2.Box().PositionX, Fl((1+5+1+3)*20))
}

func TestFloatTextIndent2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line1, line2 := unpack1(paragraph), paragraph.Box().Children[1]

	p1 := unpack1(line1)
	tu.AssertEqual(t, p1.Box().PositionX, Fl(1*20))
	tu.AssertEqual(t, p1.Box().Width, Fl(12*20))

	a := unpack1(line2)
	a1, span, a2 := unpack3(a)
	spanText := unpack1(span)
	tu.AssertEqual(t, span.Box().PositionX, Fl(0))
	tu.AssertEqual(t, spanText.Box().PositionX, Fl(0))
	tu.AssertEqual(t, span.Box().Width, Fl((1+5)*20))
	tu.AssertEqual(t, spanText.Box().Width, Fl((1+5)*20))
	tu.AssertEqual(t, a1.Box().Width, Fl(3*20))
	tu.AssertEqual(t, a1.Box().PositionX, Fl((1+5)*20))
	tu.AssertEqual(t, a2.Box().Width, Fl(2*20))
	tu.AssertEqual(t, a2.Box().PositionX, Fl((1+5+3)*20))
}

func TestFloatTextIndent3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line1, line2, line3 := unpack3(paragraph)

	p1 := unpack1(line1)
	tu.AssertEqual(t, p1.Box().PositionX, Fl(1*20))
	tu.AssertEqual(t, p1.Box().Width, Fl(12*20))

	a := unpack1(line2)
	a1, span, a2 := unpack3(a)
	spanText := unpack1(span)
	tu.AssertEqual(t, span.Box().PositionX, Fl((14-5-1)*20))
	tu.AssertEqual(t, spanText.Box().PositionX, Fl((14-5-1)*20))
	tu.AssertEqual(t, span.Box().Width, Fl((1+5)*20))
	tu.AssertEqual(t, spanText.Box().Width, Fl((1+5)*20))
	tu.AssertEqual(t, a1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, a2.Box().Width, Fl(2*20))
	tu.AssertEqual(t, a2.Box().PositionX, Fl((14-5-1-2)*20))

	p2 := unpack1(line3)
	tu.AssertEqual(t, p2.Box().PositionX, Fl(0))
	tu.AssertEqual(t, p2.Box().Width, Fl(12*20))
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
//     html =  unpack1(page)
//     body :=  unpack1(html)
//     paragraph := unpack1(body)
//     line1, line2, line3 = paragraph.Box().Children

func TestFloatTableAbortedRow(t *testing.T) {
	pages := renderPages(t, `
      <style>
        @page {size: 10px 7px}
        body {font-family: weasyprint; font-size: 2px; line-height: 1}
        div {float: right; orphans: 1}
        td {break-inside: avoid}
      </style>
      <table><tbody>
        <tr><td>abc</td></tr>
        <tr><td>abc</td></tr>
        <tr><td>def <div>f<br>g</div> ghi</td></tr>
      </tbody></table>
    `)
	page1, page2 := pages[0], pages[1]

	html := unpack1(page1)
	body := unpack1(html)
	table_wrapper := unpack1(body)
	table := unpack1(table_wrapper)
	tbody := unpack1(table)
	for _, tr := range tbody.Box().Children {
		td := unpack1(tr)
		line := unpack1(td)
		textbox := unpack1(line)
		assertText(t, textbox, "abc")
	}
	html = unpack1(page2)
	body = unpack1(html)
	table_wrapper = unpack1(body)
	table = unpack1(table_wrapper)
	tbody = unpack1(table)
	tr := unpack1(tbody)
	td := unpack1(tr)
	line1, line2 := unpack2(td)
	textbox, div := unpack2(line1)
	assertText(t, textbox, "def ")
	textbox = unpack1(line2)
	assertText(t, textbox, "ghi")
	line1, line2 = unpack2(div)
	textbox, _ = unpack2(line1)
	assertText(t, textbox, "f")
	textbox = unpack1(line2)
	assertText(t, textbox, "g")
}

func TestFormattingContextAvoidRtl(t *testing.T) {
	renderPages(t, `
      <div style="direction: rtl">
        <div style="overflow: hidden"></div>
      </div>
    `)
}
