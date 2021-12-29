package layout

import (
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for position property.

func TestRelativePositioning1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        p { height: 20px }
      </style>
      <p>1</p>
      <div style="position: relative; top: 10px">
        <p>2</p>
        <p style="position: relative; top: -5px; left: 5px">3</p>
        <p>4</p>
        <p style="position: relative; bottom: 5px; right: 5px">5</p>
        <p style="position: relative">6</p>
        <p>7</p>
      </div>
      <p>8</p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	p1, div, p8 := unpack3(body)
	p2, p3, p4, p5, p6, p7 := unpack6(div)
	tu.AssertEqual(t, [2]pr.Float{p1.Box().PositionX, p1.Box().PositionY}, [2]pr.Float{0, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{div.Box().PositionX, div.Box().PositionY}, [2]pr.Float{0, 30}, "")
	tu.AssertEqual(t, [2]pr.Float{p2.Box().PositionX, p2.Box().PositionY}, [2]pr.Float{0, 30}, "")
	tu.AssertEqual(t, [2]pr.Float{p3.Box().PositionX, p3.Box().PositionY}, [2]pr.Float{5, 45}, "") // (0 + 5, 50 - 5}
	tu.AssertEqual(t, [2]pr.Float{p4.Box().PositionX, p4.Box().PositionY}, [2]pr.Float{0, 70}, "")
	tu.AssertEqual(t, [2]pr.Float{p5.Box().PositionX, p5.Box().PositionY}, [2]pr.Float{-5, 85}, "") // (0 - 5, 90 - 5}
	tu.AssertEqual(t, [2]pr.Float{p6.Box().PositionX, p6.Box().PositionY}, [2]pr.Float{0, 110}, "")
	tu.AssertEqual(t, [2]pr.Float{p7.Box().PositionX, p7.Box().PositionY}, [2]pr.Float{0, 130}, "")
	tu.AssertEqual(t, [2]pr.Float{p8.Box().PositionX, p8.Box().PositionY}, [2]pr.Float{0, 140}, "")
	tu.AssertEqual(t, div.Box().Height, pr.Float(120), "")
}

func TestRelativePositioning2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        img { width: 20px }
        body { font-size: 0 } /* Remove spaces */
      </style>
      <body>
      <span><img src=pattern.png></span>
      <span style="position: relative; left: 10px">
        <img src=pattern.png>
        <img src=pattern.png
             style="position: relative; left: -5px; top: 5px">
        <img src=pattern.png>
        <img src=pattern.png
             style="position: relative; right: 5px; bottom: 5px">
        <img src=pattern.png style="position: relative">
        <img src=pattern.png>
      </span>
      <span><img src=pattern.png></span>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span1, span2, span3 := unpack3(line)
	img1 := span1.Box().Children[0]
	img2, img3, img4, img5, img6, img7 := unpack6(span2)
	img8 := span3.Box().Children[0]
	tu.AssertEqual(t, [2]pr.Float{img1.Box().PositionX, img1.Box().PositionY}, [2]pr.Float{0, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{img2.Box().PositionX, img2.Box().PositionY}, [2]pr.Float{30, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{img3.Box().PositionX, img3.Box().PositionY}, [2]pr.Float{45, 5}, "") // (50 - 5, y + 5)
	tu.AssertEqual(t, [2]pr.Float{img4.Box().PositionX, img4.Box().PositionY}, [2]pr.Float{70, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{img5.Box().PositionX, img5.Box().PositionY}, [2]pr.Float{85, -5}, "") // (90 - 5, y - 5)
	tu.AssertEqual(t, [2]pr.Float{img6.Box().PositionX, img6.Box().PositionY}, [2]pr.Float{110, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{img7.Box().PositionX, img7.Box().PositionY}, [2]pr.Float{130, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{img8.Box().PositionX, img8.Box().PositionY}, [2]pr.Float{140, 0}, "")
	// Don't test the span2.Box().PositionY because it depends on fonts
	tu.AssertEqual(t, span2.Box().PositionX, pr.Float(30), "")
	tu.AssertEqual(t, span2.Box().Width, pr.Float(120), "")
}

func TestAbsolutePositioning1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <div style="margin: 3px">
        <div style="height: 20px; width: 20px; position: absolute"></div>
        <div style="height: 20px; width: 20px; position: absolute;
                    left: 0"></div>
        <div style="height: 20px; width: 20px; position: absolute;
                    top: 0"></div>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div1 := body.Box().Children[0]
	div2, div3, div4 := unpack3(div1)
	tu.AssertEqual(t, div1.Box().Height, pr.Float(0), "")
	tu.AssertEqual(t, [2]pr.Float{div1.Box().PositionX, div1.Box().PositionY}, [2]pr.Float{0, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{div2.Box().Width.V(), div2.Box().Height.V()}, [2]pr.Float{20, 20}, "")
	tu.AssertEqual(t, [2]pr.Float{div2.Box().PositionX, div2.Box().PositionY}, [2]pr.Float{3, 3}, "")
	tu.AssertEqual(t, [2]pr.Float{div3.Box().Width.V(), div3.Box().Height.V()}, [2]pr.Float{20, 20}, "")
	tu.AssertEqual(t, [2]pr.Float{div3.Box().PositionX, div3.Box().PositionY}, [2]pr.Float{0, 3}, "")
	tu.AssertEqual(t, [2]pr.Float{div4.Box().Width.V(), div4.Box().Height.V()}, [2]pr.Float{20, 20}, "")
	tu.AssertEqual(t, [2]pr.Float{div4.Box().PositionX, div4.Box().PositionY}, [2]pr.Float{3, 0}, "")
}

func TestAbsolutePositioning2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <div style="position: relative; width: 20px">
        <div style="height: 20px; width: 20px; position: absolute"></div>
        <div style="height: 20px; width: 20px"></div>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div1 := body.Box().Children[0]
	div2, div3 := unpack2(div1)
	for _, div := range []Box{div1, div2, div3} {
		tu.AssertEqual(t, [2]pr.Float{div.Box().PositionX, div.Box().PositionY}, [2]pr.Float{0, 0}, "")
		tu.AssertEqual(t, [2]pr.Float{div.Box().Width.V(), div.Box().Height.V()}, [2]pr.Float{20, 20}, "")
	}
}

func TestAbsolutePositioning3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <body style="font-size: 0">
        <img src=pattern.png>
        <span style="position: relative">
          <span style="position: absolute">2</span>
          <span style="position: absolute">3</span>
          <span>4</span>
        </span>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	_, span1 := unpack2(line)
	span2, span3, span4 := unpack3(span1)
	tu.AssertEqual(t, span1.Box().PositionX, pr.Float(4), "")
	tu.AssertEqual(t, [2]pr.Float{span2.Box().PositionX, span2.Box().PositionY}, [2]pr.Float{4, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{span3.Box().PositionX, span3.Box().PositionY}, [2]pr.Float{4, 0}, "")
	tu.AssertEqual(t, span4.Box().PositionX, pr.Float(4), "")
}

func TestAbsolutePositioning4(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style> img { width: 5px; height: 20px} </style>
      <body style="font-size: 0">
        <img src=pattern.png>
        <span style="position: absolute">2</span>
        <img src=pattern.png>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img1, span, img2 := unpack3(line)
	tu.AssertEqual(t, [2]pr.Float{img1.Box().PositionX, img1.Box().PositionY}, [2]pr.Float{0, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{span.Box().PositionX, span.Box().PositionY}, [2]pr.Float{5, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{img2.Box().PositionX, img2.Box().PositionY}, [2]pr.Float{5, 0}, "")
}

func TestAbsolutePositioning5(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style> img { width: 5px; height: 20px} </style>
      <body style="font-size: 0">
        <img src=pattern.png>
        <span style="position: absolute; display: block">2</span>
        <img src=pattern.png>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img1, span, img2 := unpack3(line)
	tu.AssertEqual(t, [2]pr.Float{img1.Box().PositionX, img1.Box().PositionY}, [2]pr.Float{0, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{span.Box().PositionX, span.Box().PositionY}, [2]pr.Float{0, 20}, "")
	tu.AssertEqual(t, [2]pr.Float{img2.Box().PositionX, img2.Box().PositionY}, [2]pr.Float{5, 0}, "")
}

func TestAbsolutePositioning6(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <div style="position: relative; width: 20px; height: 60px;
                  border: 10px solid; padding-top: 6px; top: 5px; left: 1px">
        <div style="height: 20px; width: 20px; position: absolute;
                    bottom: 50%"></div>
        <div style="height: 20px; width: 20px; position: absolute;
                    top: 13px"></div>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div1 := body.Box().Children[0]
	div2, div3 := unpack2(div1)
	tu.AssertEqual(t, [2]pr.Float{div1.Box().PositionX, div1.Box().PositionY}, [2]pr.Float{1, 5}, "")
	tu.AssertEqual(t, [2]pr.Float{div1.Box().Width.V(), div1.Box().Height.V()}, [2]pr.Float{20, 60}, "")
	tu.AssertEqual(t, [2]pr.Float{div1.Box().BorderWidth(), div1.Box().BorderHeight()}, [2]pr.Float{40, 86}, "")
	tu.AssertEqual(t, [2]pr.Float{div2.Box().PositionX, div2.Box().PositionY}, [2]pr.Float{11, 28}, "")
	tu.AssertEqual(t, [2]pr.Float{div2.Box().Width.V(), div2.Box().Height.V()}, [2]pr.Float{20, 20}, "")
	tu.AssertEqual(t, [2]pr.Float{div3.Box().PositionX, div3.Box().PositionY}, [2]pr.Float{11, 28}, "")
	tu.AssertEqual(t, [2]pr.Float{div3.Box().Width.V(), div3.Box().Height.V()}, [2]pr.Float{20, 20}, "")
}

func TestAbsolutePositioning7(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 1000px 2000px }
        html { font-size: 0 }
        p { height: 20px }
      </style>
      <p>1</p>
      <div style="width: 100px">
        <p>2</p>
        <p style="position: absolute; top: -5px; left: 5px">3</p>
        <p style="margin: 3px">4</p>
        <p style="position: absolute; bottom: 5px; right: 15px;
                  width: 50px; height: 10%;
                  padding: 3px; margin: 7px">5
          <span>
            <img src="pattern.png">
            <span style="position: absolute"></span>
            <span style="position: absolute; top: -10px; right: 5px;
                         width: 20px; height: 15px"></span>
          </span>
        </p>
        <p style="margin-top: 8px">6</p>
      </div>
      <p>7</p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	p1, div, p7 := unpack3(body)
	p2, p3, p4, p5, p6 := unpack5(div)
	line := p5.Box().Children[0]
	span1 := line.Box().Children[0]
	img, span2, span3 := unpack3(span1)
	tu.AssertEqual(t, [2]pr.Float{p1.Box().PositionX, p1.Box().PositionY}, [2]pr.Float{0, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{div.Box().PositionX, div.Box().PositionY}, [2]pr.Float{0, 20}, "")
	tu.AssertEqual(t, [2]pr.Float{p2.Box().PositionX, p2.Box().PositionY}, [2]pr.Float{0, 20}, "")
	tu.AssertEqual(t, [2]pr.Float{p3.Box().PositionX, p3.Box().PositionY}, [2]pr.Float{5, -5}, "")
	tu.AssertEqual(t, [2]pr.Float{p4.Box().PositionX, p4.Box().PositionY}, [2]pr.Float{0, 40}, "")
	// p5 x = page width - right - margin/padding/border - width
	//      = 1000       - 15    - 2 * 10                - 50
	//      = 915
	// p5 y = page height - bottom - margin/padding/border - height
	//      = 2000        - 5      - 2 * 10                - 200
	//      = 1775
	tu.AssertEqual(t, [2]pr.Float{p5.Box().PositionX, p5.Box().PositionY}, [2]pr.Float{915, 1775}, "")
	tu.AssertEqual(t, [2]pr.Float{img.Box().PositionX, img.Box().PositionY}, [2]pr.Float{925, 1785}, "")
	tu.AssertEqual(t, [2]pr.Float{span2.Box().PositionX, span2.Box().PositionY}, [2]pr.Float{929, 1785}, "")
	// span3 x = p5 right - p5 margin - span width - span right
	//         = 985      - 7         - 20         - 5
	//         = 953
	// span3 y = p5 y + p5 margin top + span top
	//         = 1775 + 7             + -10
	//         = 1772
	tu.AssertEqual(t, [2]pr.Float{span3.Box().PositionX, span3.Box().PositionY}, [2]pr.Float{953, 1772}, "")
	// p6 y = p4 y + p4 margin height - margin collapsing
	//      = 40   + 26               - 3
	//      = 63
	tu.AssertEqual(t, [2]pr.Float{p6.Box().PositionX, p6.Box().PositionY}, [2]pr.Float{0, 63}, "")
	tu.AssertEqual(t, div.Box().Height, pr.Float(71), "") // 20*3 + 2*3 + 8 - 3
	tu.AssertEqual(t, [2]pr.Float{p7.Box().PositionX, p7.Box().PositionY}, [2]pr.Float{0, 91}, "")
}

func TestAbsolutePositioning8(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1264
	page := renderOnePage(t, `
      <style>@page{ width: 50px; height: 50px }</style>
      <body style="font-size: 0">
        <div style="position: absolute; margin: auto;
                    left: 0; right: 10px;
                    top: 0; bottom: 10px;
                    width: 10px; height: 20px">
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tu.AssertEqual(t, [2]pr.Float{div.Box().ContentBoxX(), div.Box().ContentBoxY()}, [2]pr.Float{15, 10}, "")
	tu.AssertEqual(t, [2]pr.Float{div.Box().Width.V(), div.Box().Height.V()}, [2]pr.Float{10, 20}, "")
}

func TestAbsoluteImages(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        img { display: block; position: absolute }
      </style>
      <div style="margin: 10px">
        <img src=pattern.png />
        <img src=pattern.png style="left: 15px" />
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	img1, img2 := unpack2(div)
	tu.AssertEqual(t, div.Box().Height, pr.Float(0), "")
	tu.AssertEqual(t, [2]pr.Float{div.Box().PositionX, div.Box().PositionY}, [2]pr.Float{0, 0}, "")
	tu.AssertEqual(t, [2]pr.Float{img1.Box().PositionX, img1.Box().PositionY}, [2]pr.Float{10, 10}, "")
	tu.AssertEqual(t, [2]pr.Float{img1.Box().Width.V(), img1.Box().Height.V()}, [2]pr.Float{4, 4}, "")
	tu.AssertEqual(t, [2]pr.Float{img2.Box().PositionX, img2.Box().PositionY}, [2]pr.Float{15, 10}, "")
	tu.AssertEqual(t, [2]pr.Float{img2.Box().Width.V(), img2.Box().Height.V()}, [2]pr.Float{4, 4}, "")
	// TODO: test the various cases in absoluteReplaced()
}

func TestFixedPositioning(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// TODO:test page-break-before: left/right
	pages := renderPages(t, `
      a
      <div style="page-break-before: always; page-break-after: always">
        <p style="position: fixed">b</p>
      </div>
      c
    `)
	page1, page2, page3 := pages[0], pages[1], pages[2]

	tags := func(boxes []Box) []string {
		var out []string
		for _, b := range boxes {
			out = append(out, b.Box().ElementTag())
		}
		return out
	}
	html := page1.Box().Children[0]
	tu.AssertEqual(t, tags(html.Box().Children), []string{"body", "p"}, "")
	html = page2.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tu.AssertEqual(t, tags(div.Box().Children), []string{"p"}, "")
	html = page3.Box().Children[0]
	tu.AssertEqual(t, tags(html.Box().Children), []string{"p", "body"}, "")
}

func TestFixedPositioningRegression1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/641
	pages := renderPages(t, `
      <style>
        @page:first { size: 100px 200px }
        @page { size: 200px 100px; margin: 0 }
        article { break-after: page }
        .fixed { position: fixed; top: 10px; width: 20px }
      </style>
      <ul class="fixed" style="right: 0"><li>a</li></ul>
      <img class="fixed" style="right: 20px" src="pattern.png" />
      <div class="fixed" style="right: 40px">b</div>
      <article>page1</article>
      <article>page2</article>
    `)
	page1, page2 := pages[0], pages[1]

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	ul, img, div, article := unpack4(body)
	marker := ul.Box().Children[0]
	tu.AssertEqual(t, [2]pr.Float{ul.Box().PositionX, ul.Box().PositionY}, [2]pr.Float{80, 10}, "")
	tu.AssertEqual(t, [2]pr.Float{img.Box().PositionX, img.Box().PositionY}, [2]pr.Float{60, 10}, "")
	tu.AssertEqual(t, [2]pr.Float{div.Box().PositionX, div.Box().PositionY}, [2]pr.Float{40, 10}, "")
	tu.AssertEqual(t, [2]pr.Float{article.Box().PositionX, article.Box().PositionY}, [2]pr.Float{0, 0}, "")
	tu.AssertEqual(t, marker.Box().PositionX, ul.Box().PositionX, "")

	html = page2.Box().Children[0]
	ul, img, div, body = unpack4(html)
	marker = ul.Box().Children[0]
	tu.AssertEqual(t, [2]pr.Float{ul.Box().PositionX, ul.Box().PositionY}, [2]pr.Float{180, 10}, "")
	tu.AssertEqual(t, [2]pr.Float{img.Box().PositionX, img.Box().PositionY}, [2]pr.Float{160, 10}, "")
	tu.AssertEqual(t, [2]pr.Float{div.Box().PositionX, div.Box().PositionY}, [2]pr.Float{140, 10}, "")
	tu.AssertEqual(t, [2]pr.Float{article.Box().PositionX, article.Box().PositionY}, [2]pr.Float{0, 0}, "")
	tu.AssertEqual(t, marker.Box().PositionX, ul.Box().PositionX, "")
}

func TestFixedPositioningRegression2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/728
	pages := renderPages(t, `
      <style>
        @page { size: 100px 100px }
        section { break-after: page }
        .fixed { position: fixed; top: 10px; left: 15px; width: 20px }
      </style>
      <div class="fixed">
        <article class="fixed" style="top: 20px">
          <header class="fixed" style="left: 5px"></header>
        </article>
      </div>
      <section></section>
      <pre></pre>
    `)
	page1, page2 := pages[0], pages[1]

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	div, _ := unpack2(body)
	tu.AssertEqual(t, [2]pr.Float{div.Box().PositionX, div.Box().PositionY}, [2]pr.Float{15, 10}, "")
	article := div.Box().Children[0]
	tu.AssertEqual(t, [2]pr.Float{article.Box().PositionX, article.Box().PositionY}, [2]pr.Float{15, 20}, "")
	header := article.Box().Children[0]
	tu.AssertEqual(t, [2]pr.Float{header.Box().PositionX, header.Box().PositionY}, [2]pr.Float{5, 10}, "")

	html = page2.Box().Children[0]
	div, body = unpack2(html)
	tu.AssertEqual(t, [2]pr.Float{div.Box().PositionX, div.Box().PositionY}, [2]pr.Float{15, 10}, "")
	article = div.Box().Children[0]
	tu.AssertEqual(t, [2]pr.Float{article.Box().PositionX, article.Box().PositionY}, [2]pr.Float{15, 20}, "")
	header = article.Box().Children[0]
	tu.AssertEqual(t, [2]pr.Float{header.Box().PositionX, header.Box().PositionY}, [2]pr.Float{5, 10}, "")
}
