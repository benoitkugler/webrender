package layout

import (
	"fmt"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for blocks layout.

func TestBlockWidths(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { margin: 0; size: 120px 2000px }
        body { margin: 0 }
        div { margin: 10px }
        p { padding: 2px; border-width: 1px; border-style: solid }
      </style>
      <div>
        <p></p>
        <p style="width: 50px"></p>
      </div>
      <div style="direction: rtl">
        <p style="width: 50px; direction: rtl"></p>
      </div>
      <div>
        <p style="margin: 0 10px 0 20px"></p>
        <p style="width: 50px; margin-left: 20px; margin-right: auto"></p>
        <p style="width: 50px; margin-left: auto; margin-right: 20px"></p>
        <p style="width: 50px; margin: auto"></p>
  
        <p style="margin-left: 20px; margin-right: auto"></p>
        <p style="margin-left: auto; margin-right: 20px"></p>
        <p style="margin: auto"></p>

        <p style="width: 200px; margin: auto"></p>

        <p style="min-width: 200px; margin: auto"></p>
        <p style="max-width: 50px; margin: auto"></p>
        <p style="min-width: 50px; margin: auto"></p>

        <p style="width: 70%"></p>
      </div>
    `)
	html := unpack1(page)
	tu.AssertEqual(t, html.Box().ElementTag(), "html")
	body := unpack1(html)
	tu.AssertEqual(t, body.Box().ElementTag(), "body")
	tu.AssertEqual(t, body.Box().Width, Fl(120))

	divs := body.Box().Children

	var paragraphs []Box
	for _, div := range divs {
		tu.AssertEqual(t, bo.BlockT.IsInstance(div), true)
		tu.AssertEqual(t, div.Box().ElementTag(), "div")
		tu.AssertEqual(t, div.Box().Width, Fl(100))
		for _, paragraph := range div.Box().Children {
			tu.AssertEqual(t, bo.BlockT.IsInstance(paragraph), true)
			tu.AssertEqual(t, paragraph.Box().ElementTag(), "p")
			tu.AssertEqual(t, paragraph.Box().PaddingLeft, Fl(2))
			tu.AssertEqual(t, paragraph.Box().PaddingRight, Fl(2))
			tu.AssertEqual(t, paragraph.Box().BorderLeftWidth, Fl(1))
			tu.AssertEqual(t, paragraph.Box().BorderRightWidth, Fl(1))
			paragraphs = append(paragraphs, paragraph)
		}
	}
	tu.AssertEqual(t, len(paragraphs), 15)

	// width is "auto"
	tu.AssertEqual(t, paragraphs[0].Box().Width, Fl(94))
	tu.AssertEqual(t, paragraphs[0].Box().MarginLeft, Fl(0))
	tu.AssertEqual(t, paragraphs[0].Box().MarginRight, Fl(0))

	// No "auto", over-constrained equation with ltr, the initial
	// "margin-right: 0" was ignored.
	tu.AssertEqual(t, paragraphs[1].Box().Width, Fl(50))
	tu.AssertEqual(t, paragraphs[1].Box().MarginLeft, Fl(0))

	// No "auto", over-constrained equation with rtl, the initial
	// "margin-left: 0" was ignored.
	tu.AssertEqual(t, paragraphs[2].Box().Width, Fl(50))
	tu.AssertEqual(t, paragraphs[2].Box().MarginRight, Fl(0))

	// width is "auto"
	tu.AssertEqual(t, paragraphs[3].Box().Width, Fl(64))
	tu.AssertEqual(t, paragraphs[3].Box().MarginLeft, Fl(20))

	// margin-right is "auto"
	tu.AssertEqual(t, paragraphs[4].Box().Width, Fl(50))
	tu.AssertEqual(t, paragraphs[4].Box().MarginLeft, Fl(20))

	// margin-left is "auto"
	tu.AssertEqual(t, paragraphs[5].Box().Width, Fl(50))
	tu.AssertEqual(t, paragraphs[5].Box().MarginLeft, Fl(24))

	// Both margins are "auto", remaining space is split := range half
	tu.AssertEqual(t, paragraphs[6].Box().Width, Fl(50))
	tu.AssertEqual(t, paragraphs[6].Box().MarginLeft, Fl(22))

	// width is "auto", other "auto" are set to 0
	tu.AssertEqual(t, paragraphs[7].Box().Width, Fl(74))
	tu.AssertEqual(t, paragraphs[7].Box().MarginLeft, Fl(20))

	// width is "auto", other "auto" are set to 0
	tu.AssertEqual(t, paragraphs[8].Box().Width, Fl(74))
	tu.AssertEqual(t, paragraphs[8].Box().MarginLeft, Fl(0))

	// width is "auto", other "auto" are set to 0
	tu.AssertEqual(t, paragraphs[9].Box().Width, Fl(94))
	tu.AssertEqual(t, paragraphs[9].Box().MarginLeft, Fl(0))

	// sum of non-auto initially is too wide, set auto values to 0
	tu.AssertEqual(t, paragraphs[10].Box().Width, Fl(200))
	tu.AssertEqual(t, paragraphs[10].Box().MarginLeft, Fl(0))

	// Constrained by min-width, same as above
	tu.AssertEqual(t, paragraphs[11].Box().Width, Fl(200))
	tu.AssertEqual(t, paragraphs[11].Box().MarginLeft, Fl(0))

	// Constrained by max-width, same as paragraphs[6]
	tu.AssertEqual(t, paragraphs[12].Box().Width, Fl(50))
	tu.AssertEqual(t, paragraphs[12].Box().MarginLeft, Fl(22))

	// NOT constrained by min-width
	tu.AssertEqual(t, paragraphs[13].Box().Width, Fl(94))
	tu.AssertEqual(t, paragraphs[13].Box().MarginLeft, Fl(0))

	// 70%
	tu.AssertEqual(t, paragraphs[14].Box().Width, Fl(70))
	tu.AssertEqual(t, paragraphs[14].Box().MarginLeft, Fl(0))
}

func TestBlockHeightsP(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { margin: 0; size: 100px 20000px }
        html, body { margin: 0 }
        div { margin: 4px; border: 2px solid; padding: 4px }
        /* Use top margins so that margin collapsing doesn"t change result */
        p { margin: 16px 0 0; border: 4px solid; padding: 8px; height: 50px }
      </style>
      <div>
        <p></p>
        <!-- Not in normal flow: don't contribute to the parent’s height -->
        <p style="position: absolute"></p>
        <p style="float: left"></p>
      </div>
      <div> <p></p> <p></p> <p></p> </div>
      <div style="height: 20px"> <p></p> </div>
      <div style="height: 120px"> <p></p> </div>
      <div style="max-height: 20px"> <p></p> </div>
      <div style="min-height: 120px"> <p></p> </div>
      <div style="min-height: 20px"> <p></p> </div>
      <div style="max-height: 120px"> <p></p> </div>
    `)
	html := unpack1(page)
	body := unpack1(html)

	var heights []pr.Float
	for _, div := range body.Box().Children {
		heights = append(heights, div.Box().Height.V())
	}
	tu.AssertEqual(t, heights, []pr.Float{90, 90 * 3, 20, 120, 20, 120, 90, 90})
}

func TestBlockHeightsImg(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        body { height: 200px; font-size: 0 }
      </style>
      <div>
        <img src=pattern.png style="height: 40px">
      </div>
      <div style="height: 10%">
        <img src=pattern.png style="height: 40px">
      </div>
      <div style="max-height: 20px">
        <img src=pattern.png style="height: 40px">
      </div>
      <div style="max-height: 10%">
        <img src=pattern.png style="height: 40px">
      </div>
      <div style="min-height: 20px"></div>
      <div style="min-height: 10%"></div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	var heights []pr.Float
	for _, div := range body.Box().Children {
		heights = append(heights, div.Box().Height.V())
	}
	tu.AssertEqual(t, heights, []pr.Float{40, 20, 20, 20, 20, 20})
}

func TestBlockHeightsImgNoBodyHeight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Same but with no height on body: percentage *-height is ignored
	page := renderOnePage(t, `
      <style>
        body { font-size: 0 }
      </style>
        <div>
          <img src=pattern.png style="height: 40px">
        </div>
        <div style="height: 10%">
          <img src=pattern.png style="height: 40px">
        </div>
        <div style="max-height: 20px">
          <img src=pattern.png style="height: 40px">
        </div>
        <div style="max-height: 10%">
          <img src=pattern.png style="height: 40px">
        </div>
        <div style="min-height: 20px"></div>
        <div style="min-height: 10%"></div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	var heights []pr.Float
	for _, div := range body.Box().Children {
		heights = append(heights, div.Box().Height.V())
	}
	tu.AssertEqual(t, heights, []pr.Float{40, 40, 20, 40, 20, 0})
}

func TestBlockPercentageHeightsNoHtmlHeight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        html, body { margin: 0 }
        body { height: 50% }
      </style>
    `)
	html := unpack1(page)
	tu.AssertEqual(t, html.Box().ElementTag(), "html")
	body := unpack1(html)
	tu.AssertEqual(t, body.Box().ElementTag(), "body")

	// Since html’s height depend on body’s, body’s 50% means "auto"
	tu.AssertEqual(t, body.Box().Height, Fl(0))
}

func TestBlockPercentageHeights(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        html, body { margin: 0 }
        html { height: 300px }
        body { height: 50% }
      </style>
    `)
	html := unpack1(page)
	tu.AssertEqual(t, html.Box().ElementTag(), "html")
	body := unpack1(html)
	tu.AssertEqual(t, body.Box().ElementTag(), "body")

	// This time the percentage makes sense
	tu.AssertEqual(t, body.Box().Height, Fl(150))
}

func TestBoxSizing(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, size := range []string{
		"width: 10%; height: 1000px",
		"max-width: 10%; max-height: 1000px; height: 2000px",
		"width: 5%; min-width: 10%; min-height: 1000px",
		"width: 10%; height: 1000px; min-width: auto; max-height: none",
	} {
		testBoxSizing(t, size)
	}
}

func testBoxSizing(t *testing.T, size string) {
	// https://www.w3.org/TR/css-ui-3/#box-sizing
	page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @page { size: 100000px }
        body { width: 10000px; margin: 0 }
        div { %s; margin: 100px; padding: 10px; border: 1px solid }
      </style>
      <div></div>
 
      <div style="box-sizing: content-box"></div>
      <div style="box-sizing: padding-box"></div>
      <div style="box-sizing: border-box"></div>
    `, size))
	html := unpack1(page)
	body := unpack1(html)
	div1, div2, div3, div4 := unpack4(body)
	for _, div := range []Box{div1, div2} {
		tu.AssertEqual(t, div.Box().Style.GetBoxSizing(), pr.String("content-box"))
		tu.AssertEqual(t, div.Box().Width, Fl(1000))
		tu.AssertEqual(t, div.Box().Height, Fl(1000))
		tu.AssertEqual(t, div.Box().PaddingWidth(), Fl(1020))
		tu.AssertEqual(t, div.Box().PaddingHeight(), Fl(1020))
		tu.AssertEqual(t, div.Box().BorderWidth(), Fl(1022))
		tu.AssertEqual(t, div.Box().BorderHeight(), Fl(1022))
		tu.AssertEqual(t, div.Box().MarginHeight(), Fl(1222))
		// marginWidth() is the width of the containing block
	}
	// padding-box
	tu.AssertEqual(t, div3.Box().Style.GetBoxSizing(), pr.String("padding-box"))
	tu.AssertEqual(t, div3.Box().Width, Fl(980)) // 1000 - 20
	tu.AssertEqual(t, div3.Box().Height, Fl(980))
	tu.AssertEqual(t, div3.Box().PaddingWidth(), Fl(1000))
	tu.AssertEqual(t, div3.Box().PaddingHeight(), Fl(1000))
	tu.AssertEqual(t, div3.Box().BorderWidth(), Fl(1002))
	tu.AssertEqual(t, div3.Box().BorderHeight(), Fl(1002))
	tu.AssertEqual(t, div3.Box().MarginHeight(), Fl(1202))

	// border-box
	tu.AssertEqual(t, div4.Box().Style.GetBoxSizing(), pr.String("border-box"))
	tu.AssertEqual(t, div4.Box().Width, Fl(978)) // 1000 - 20 - 2
	tu.AssertEqual(t, div4.Box().Height, Fl(978))
	tu.AssertEqual(t, div4.Box().PaddingWidth(), Fl(998))
	tu.AssertEqual(t, div4.Box().PaddingHeight(), Fl(998))
	tu.AssertEqual(t, div4.Box().BorderWidth(), Fl(1000))
	tu.AssertEqual(t, div4.Box().BorderHeight(), Fl(1000))
	tu.AssertEqual(t, div4.Box().MarginHeight(), Fl(1200))
}

func TestBoxSizingZero(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, size := range []string{
		"width: 0; height: 0",
		"max-width: 0; max-height: 0",
		"min-width: 0; min-height: 0; width: 0; height: 0",
	} {
		testBoxSizingZero(t, size)
	}
}

func testBoxSizingZero(t *testing.T, size string) {
	// https://www.w3.org/TR/css-ui-3/#box-sizing
	page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @page { size: 100000px }
        body { width: 10000px; margin: 0 }
        div { %s; margin: 100px; padding: 10px; border: 1px solid }
      </style>
      <div></div>

      <div style="box-sizing: content-box"></div>
      <div style="box-sizing: padding-box"></div>
      <div style="box-sizing: border-box"></div>
    `, size))
	html := unpack1(page)
	body := unpack1(html)
	for _, div := range body.Box().Children {
		tu.AssertEqual(t, div.Box().Width, Fl(0))
		tu.AssertEqual(t, div.Box().Height, Fl(0))
		tu.AssertEqual(t, div.Box().PaddingWidth(), Fl(20))
		tu.AssertEqual(t, div.Box().PaddingHeight(), Fl(20))
		tu.AssertEqual(t, div.Box().BorderWidth(), Fl(22))
		tu.AssertEqual(t, div.Box().BorderHeight(), Fl(22))
		tu.AssertEqual(t, div.Box().MarginHeight(), Fl(222))
		// marginWidth() is the width of the containing block
	}
}

type collapseData struct {
	margin1, margin2 string
	result           pr.Float
}

var (
	COLLAPSING = [...]collapseData{
		{"10px", "15px", 15}, // ! 25
		// "The maximum of the absolute values of the negative adjoining margins is
		// deducted from the maximum of the positive adjoining margins"
		{"-10px", "15px", 5},
		{"10px", "-15px", -5},
		{"-10px", "-15px", -15},
		{"10px", "auto", 10}, // "auto" is 0
	}

	NOTCOLLAPSING = [...]collapseData{
		{"10px", "15px", 25},
		{"-10px", "15px", 5},
		{"10px", "-15px", -5},
		{"-10px", "-15px", -25},
		{"10px", "auto", 10}, // "auto" is 0
	}
)

func TestVerticalSpace1(t *testing.T) {
	for _, data := range COLLAPSING {
		// Siblings
		page := renderOnePage(t, fmt.Sprintf(`
		<style>
			p { font: 20px/1 serif } /* block height , 20px */
			#p1 { margin-bottom: %s }
			#p2 { margin-top: %s }
		</style>
		<p id=p1>Lorem ipsum
		<p id=p2>dolor sit amet
    `, data.margin1, data.margin2))
		html := unpack1(page)
		body := unpack1(html)
		p1, p2 := unpack2(body)
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result)
	}
}

func TestVerticalSpace2(t *testing.T) {
	for _, data := range COLLAPSING {

		// Not siblings, first is nested
		page := renderOnePage(t, fmt.Sprintf(`
		<style>
			p { font: 20px/1 serif } /* block height , 20px */
			#p1 { margin-bottom: %s }
			#p2 { margin-top: %s }
		</style>
		<div>
			<p id=p1>Lorem ipsum
		</div>
		<p id=p2>dolor sit amet
    `, data.margin1, data.margin2))
		html := unpack1(page)
		body := unpack1(html)
		div, p2 := unpack2(body)
		p1 := unpack1(div)
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result)
	}
}

func TestVerticalSpace3(t *testing.T) {
	for _, data := range COLLAPSING {
		// Not siblings, second is nested
		page := renderOnePage(t, fmt.Sprintf(`
		<style>
			p { font: 20px/1 serif } /* block height , 20px */
			#p1 { margin-bottom: %s }
			#p2 { margin-top: %s }
		</style>
		<p id=p1>Lorem ipsum
		<div>
			<p id=p2>dolor sit amet
		</div>
    `, data.margin1, data.margin2))
		html := unpack1(page)
		body := unpack1(html)
		p1, div := unpack2(body)
		p2 := unpack1(div)
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result)
	}
}

func TestVerticalSpace4(t *testing.T) {
	for _, data := range COLLAPSING {
		// Not siblings, second is doubly nested
		page := renderOnePage(t, fmt.Sprintf(`
		<style>
			p { font: 20px/1 serif } /* block height , 20px */
			#p1 { margin-bottom: %s }
			#p2 { margin-top: %s }
		</style>
		<p id=p1>Lorem ipsum
		<div>
			<div>
				<p id=p2>dolor sit amet
			</div>
		</div>
    `, data.margin1, data.margin2))
		html := unpack1(page)
		body := unpack1(html)
		p1, div1 := unpack2(body)
		div2 := unpack1(div1)
		p2 := unpack1(div2)
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result)
	}
}

func TestVerticalSpace5(t *testing.T) {
	for _, data := range COLLAPSING {
		// Collapsing with children
		page := renderOnePage(t, fmt.Sprintf(`
		<style>
			p { font: 20px/1 serif } /* block height , 20px */
			#div1 { margin-top: %s }
			#div2 { margin-top: %s }
		</style>
		<p>Lorem ipsum
		<div id=div1>
			<div id=div2>
			<p id=p2>dolor sit amet
			</div>
		</div>
    `, data.margin1, data.margin2))
		html := unpack1(page)
		body := unpack1(html)
		p1, div1 := unpack2(body)
		div2 := unpack1(div1)
		p2 := unpack1(div2)
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		// Parent and element edge are the same:
		tu.AssertEqual(t, div1.Box().BorderBoxY(), p2.Box().BorderBoxY())
		tu.AssertEqual(t, div2.Box().BorderBoxY(), p2.Box().BorderBoxY())
		tu.AssertEqual(t, p2Top-p1Bottom, data.result)
	}
}

func TestVerticalSpace6(t *testing.T) {
	for _, data := range NOTCOLLAPSING {
		// Block formatting context: Not collapsing with children
		page := renderOnePage(t, fmt.Sprintf(`
		<style>
			p { font: 20px/1 serif } /* block height , 20px */
			#div1 { margin-top: %s; overflow: hidden }
			#div2 { margin-top: %s }
		</style>
		<p>Lorem ipsum
		<div id=div1>
			<div id=div2>
			<p id=p2>dolor sit amet
			</div>
		</div>
    `, data.margin1, data.margin2))
		html := unpack1(page)
		body := unpack1(html)
		p1, div1 := unpack2(body)
		div2 := unpack1(div1)
		p2 := unpack1(div2)
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result)
	}
}

func TestVerticalSpace7(t *testing.T) {
	for _, data := range COLLAPSING {
		// Collapsing through an empty div
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        p { font: 20px/1 serif } /* block height , 20px */
        #p1 { margin-bottom: %s }
        #p2 { margin-top: %s }
        div { margin-bottom: %s; margin-top: %s }
      </style>
      <p id=p1>Lorem ipsum
      <div></div>
      <p id=p2>dolor sit amet
    `, data.margin1, data.margin2, data.margin1, data.margin2))
		html := unpack1(page)
		body := unpack1(html)
		p1, _, p2 := unpack3(body)
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result)
	}
}

func TestVerticalSpace8(t *testing.T) {
	for _, data := range NOTCOLLAPSING {
		// The root element does not collapse
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        html { margin-top: %s }
        body { margin-top: %s }
      </style>
      <p>Lorem ipsum
    `, data.margin1, data.margin2))
		html := unpack1(page)
		body := unpack1(html)
		p1 := unpack1(body)
		p1Top := p1.Box().ContentBoxY()
		// Vertical space from y=0
		tu.AssertEqual(t, p1Top, data.result)
	}
}

func TestVerticalSpace9(t *testing.T) {
	for _, data := range COLLAPSING {
		// <body> DOES collapse
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        body { margin-top: %s }
        div { margin-top: %s }
      </style>
      <div>
        <p>Lorem ipsum
    `, data.margin1, data.margin2))
		html := unpack1(page)
		body := unpack1(html)
		div := unpack1(body)
		p1 := unpack1(div)
		p1Top := p1.Box().ContentBoxY()
		// Vertical space from y=0
		tu.AssertEqual(t, p1Top, data.result)
	}
}

func TestBoxDecorationBreakBlockSlice(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// https://www.w3.org/TR/css-backgrounds-3/#the-box-decoration-break
	pages := renderPages(t, `
      <style>
        @page { size: 100px }
        p { padding: 2px; border: 3px solid; margin: 5px }
        img { display: block; height: 40px }
      </style>
      <p>
        <img src=pattern.png>
        <img src=pattern.png>
        <img src=pattern.png>
        <img src=pattern.png>`)
	page1, page2 := pages[0], pages[1]
	html := unpack1(page1)
	body := unpack1(html)
	paragraph := unpack1(body)
	img1, img2 := unpack2(paragraph)
	tu.AssertEqual(t, paragraph.Box().PositionY, Fl(0))
	tu.AssertEqual(t, paragraph.Box().MarginTop, Fl(5))
	tu.AssertEqual(t, paragraph.Box().BorderTopWidth, Fl(3))
	tu.AssertEqual(t, paragraph.Box().PaddingTop, Fl(2))
	tu.AssertEqual(t, paragraph.Box().ContentBoxY(), Fl(10))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(10))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(50))
	tu.AssertEqual(t, paragraph.Box().Height, Fl(90))
	tu.AssertEqual(t, paragraph.Box().MarginBottom, Fl(0))
	tu.AssertEqual(t, paragraph.Box().BorderBottomWidth, Fl(0))
	tu.AssertEqual(t, paragraph.Box().PaddingBottom, Fl(0))
	tu.AssertEqual(t, paragraph.Box().MarginHeight(), Fl(100))

	html = unpack1(page2)
	body = unpack1(html)
	paragraph = unpack1(body)
	img1, img2 = unpack2(paragraph)
	tu.AssertEqual(t, paragraph.Box().PositionY, Fl(0))
	tu.AssertEqual(t, paragraph.Box().MarginTop, Fl(0))
	tu.AssertEqual(t, paragraph.Box().BorderTopWidth, Fl(0))
	tu.AssertEqual(t, paragraph.Box().PaddingTop, Fl(0))
	tu.AssertEqual(t, paragraph.Box().ContentBoxY(), Fl(0))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(40))
	tu.AssertEqual(t, paragraph.Box().Height, Fl(80))
	tu.AssertEqual(t, paragraph.Box().PaddingBottom, Fl(2))
	tu.AssertEqual(t, paragraph.Box().BorderBottomWidth, Fl(3))
	tu.AssertEqual(t, paragraph.Box().MarginBottom, Fl(5))
	tu.AssertEqual(t, paragraph.Box().MarginHeight(), Fl(90))
}

func TestBoxDecorationBreakBlockClone(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// https://www.w3.org/TR/css-backgrounds-3/#the-box-decoration-break
	pages := renderPages(t, `
      <style>
        @page { size: 100px }
        p { padding: 2px; border: 3px solid; margin: 5px;
            box-decoration-break: clone }
        img { display: block; height: 40px }
      </style>
      <p>
        <img src=pattern.png>
        <img src=pattern.png>
        <img src=pattern.png>
        <img src=pattern.png>`)
	page1, page2 := pages[0], pages[1]
	html := unpack1(page1)
	body := unpack1(html)
	paragraph := unpack1(body)
	img1, img2 := unpack2(paragraph)
	tu.AssertEqual(t, paragraph.Box().PositionY, Fl(0))
	tu.AssertEqual(t, paragraph.Box().MarginTop, Fl(5))
	tu.AssertEqual(t, paragraph.Box().BorderTopWidth, Fl(3))
	tu.AssertEqual(t, paragraph.Box().PaddingTop, Fl(2))
	tu.AssertEqual(t, paragraph.Box().ContentBoxY(), Fl(10))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(10))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(50))
	tu.AssertEqual(t, paragraph.Box().Height, Fl(80))
	// TODO: bottom margin should be 0
	// https://www.w3.org/TR/css-break-3/#valdef-box-decoration-break-clone
	// "Cloned margins are truncated on block-level boxes."
	// See https://github.com/Kozea/WeasyPrint/issues/115
	tu.AssertEqual(t, paragraph.Box().MarginBottom, Fl(5))
	tu.AssertEqual(t, paragraph.Box().BorderBottomWidth, Fl(3))
	tu.AssertEqual(t, paragraph.Box().PaddingBottom, Fl(2))
	tu.AssertEqual(t, paragraph.Box().MarginHeight(), Fl(100))

	html = unpack1(page2)
	body = unpack1(html)
	paragraph = unpack1(body)
	img1, img2 = unpack2(paragraph)
	tu.AssertEqual(t, paragraph.Box().PositionY, Fl(0))
	tu.AssertEqual(t, paragraph.Box().MarginTop, Fl(0))
	tu.AssertEqual(t, paragraph.Box().BorderTopWidth, Fl(3))
	tu.AssertEqual(t, paragraph.Box().PaddingTop, Fl(2))
	tu.AssertEqual(t, paragraph.Box().ContentBoxY(), Fl(5))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(5))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(45))
	tu.AssertEqual(t, paragraph.Box().Height, Fl(80))
	tu.AssertEqual(t, paragraph.Box().PaddingBottom, Fl(2))
	tu.AssertEqual(t, paragraph.Box().BorderBottomWidth, Fl(3))
	tu.AssertEqual(t, paragraph.Box().MarginBottom, Fl(5))
	tu.AssertEqual(t, paragraph.Box().MarginHeight(), Fl(95))
}

func TestBoxDecorationBreakCloneBottomPadding(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 80px; margin: 0 }
        div { height: 20px }
        article { padding: 12px; box-decoration-break: clone }
      </style>
      <article>
        <div>a</div>
        <div>b</div>
        <div>c</div>
      </article>`)
	page1, page2 := pages[0], pages[1]
	html := unpack1(page1)
	body := unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().Height, Fl(80-2*12))
	div1, div2 := unpack2(article)
	tu.AssertEqual(t, div1.Box().PositionY, Fl(12))
	tu.AssertEqual(t, div2.Box().PositionY, Fl(12+20))

	html = unpack1(page2)
	body = unpack1(html)
	article = unpack1(body)
	tu.AssertEqual(t, article.Box().Height, Fl(20))
	div := unpack1(article)
	tu.AssertEqual(t, div.Box().PositionY, Fl(12))
}

// @pytest.mark.xfail
// func TestBoxDecorationBreakSliceBottomPadding():  // pragma: no cot*testing.Tver
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     // Last div fits := range first, but ! article"s padding. As it is impossible to
//     // break between a parent && its last child, put last child on next page.
//     // TODO: at the end of blockContainerLayout, we should check that the box
//     // with its bottom border/padding doesn"t cross the bottom line. If it does,
//     // we should re-render the box with a maxPositionY including the bottom
//     // border/padding.
//     page1, page2 = renderPages(`
//       <style>
//         @page { size: 80px; margin: 0 }
//         div { height: 20px }
//         article { padding: 12px; box-decoration-break: slice }
//       </style>
//       <article>
//         <div>a</div>
//         <div>b</div>
//         <div>c</div>
//       </article>`)
//     html := unpack1(page1)
//     body :=  unpack1(html)
//     article := unpack1(body)
//     tu.AssertEqual(t, article.Box().Height , 80 - 12)
//     div1, div2 = article.Box().Children
//     tu.AssertEqual(t, div1.Box().PositionY , 12)
//     tu.AssertEqual(t, div2.Box().PositionY , 12 + 20)

//     html := unpack1(page2)
//     body :=  unpack1(html)
//     article := unpack1(body)
//     tu.AssertEqual(t, article.Box().Height , 20)
//     div := unpack1(article)
//     tu.AssertEqual(t, div.Box().PositionY , 0)

func TestOverflowAuto(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="overflow: auto">
        <div style="float: left; height: 50px; margin: 10px">bla bla bla</div>
          toto toto`)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().Height, Fl(50+10+10))
}

func TestOverflowHiddenInFlowLayout(t *testing.T) {
	page := renderOnePage(t, `
      <div style="overflow: hidden; height: 3px;">
        <div>abc</div>
        <div style="height: 100px; margin: 50px;">def</div>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	parentDiv := unpack1(body)
	tu.AssertEqual(t, parentDiv.Box().Height, Fl(3))
}

func TestOverflowHiddenOutOfFlowLayout(t *testing.T) {
	page := renderOnePage(t, `
      <div style="overflow: hidden; height: 3px;">
        <div style="float: left;">abc</div>
        <div style="float: right; height: 100px; margin: 50px;">def</div>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	parentDiv := unpack1(body)
	tu.AssertEqual(t, parentDiv.Box().Height, Fl(3))
}

// Test regression: https://github.com/Kozea/WeasyPrint/issues/943
func TestBoxMarginTopRepagination(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 50px }
        :root { line-height: 1; font-size: 10px }
        a::before { content: target-counter(attr(href), page) }
        div { margin: 20px 0 0; background: yellow }
      </style>
      <p><a href="#title"></a></p>
      <div>1<br/>1<br/>2<br/>2</div>
      <h1 id="title">title</h1>
    `)
	page1, page2 := pages[0], pages[1]
	html := unpack1(page1)
	body := unpack1(html)
	_, div := unpack2(body)
	tu.AssertEqual(t, div.Box().MarginTop, Fl(20))
	tu.AssertEqual(t, div.Box().PaddingBoxY(), Fl(10+20))

	html = unpack1(page2)
	body = unpack1(html)
	div, _ = unpack2(body)
	tu.AssertEqual(t, div.Box().MarginTop, Fl(0))
	tu.AssertEqual(t, div.Box().PaddingBoxY(), Fl(0))
}

func TestContinueDiscard(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
      <style>
        @page { size: 80px; margin: 0 }
        div { display: inline-block; width: 100%; height: 25px }
        article { continue: discard; border: 1px solid; line-height: 1 }
      </style>
      <article>
        <div>a</div>
        <div>b</div>
        <div>c</div>
        <div>d</div>
        <div>e</div>
        <div>f</div>
      </article>`)
	html := unpack1(page1)
	body := unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().Height, Fl(3*25))
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().PositionY, Fl(1))
	tu.AssertEqual(t, div2.Box().PositionY, Fl(1+25))
	tu.AssertEqual(t, div3.Box().PositionY, Fl(1+25*2))
	tu.AssertEqual(t, article.Box().BorderBottomWidth, Fl(1))
}

func TestContinueDiscardChildren(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
  	<style>
        @page { size: 80px; margin: 0 }
        div { display: inline-block; width: 100%; height: 25px }
        section { border: 1px solid }
        article { continue: discard; border: 1px solid; line-height: 1 }
      </style>
      <article>
        <section>
          <div>a</div>
          <div>b</div>
          <div>c</div>
          <div>d</div>
          <div>e</div>
          <div>f</div>
        </section>
      </article>`)
	html := unpack1(page1)
	body := unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().Height, Fl(2+3*25))
	section := unpack1(article)
	tu.AssertEqual(t, section.Box().Height, Fl(3*25))
	div1, div2, div3 := unpack3(section)
	tu.AssertEqual(t, div1.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div2.Box().PositionY, Fl(2+25))
	tu.AssertEqual(t, div3.Box().PositionY, Fl(2+25*2))
	tu.AssertEqual(t, article.Box().BorderBottomWidth, Fl(1))
}

func TestBlockInBlockWithBottomPadding(t *testing.T) {
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1476
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 8em 3.5em }
        body { line-height: 1; font-family: weasyprint }
        div { padding-bottom: 1em }
      </style>
      abc def
      <div>
        <p>
          ghi jkl
          mno pqr
        </p>
      </div>
      stu vwx`)
	page1, page2 := pages[0], pages[1]

	html := unpack1(page1)
	body := unpack1(html)
	anonBody, div := unpack2(body)
	line := unpack1(anonBody)
	tu.AssertEqual(t, line.Box().Height, Fl(16))
	assertText(t, unpack1(line), "abc def")
	p := unpack1(div)
	line = unpack1(p)
	tu.AssertEqual(t, line.Box().Height, Fl(16))
	assertText(t, unpack1(line), "ghi jkl")

	html = unpack1(page2)
	body = unpack1(html)
	div, anonBody = unpack2(body)
	p = unpack1(div)
	line = unpack1(p)
	tu.AssertEqual(t, line.Box().Height, Fl(16))
	assertText(t, unpack1(line), "mno pqr")
	line = unpack1(anonBody)
	tu.AssertEqual(t, line.Box().Height, Fl(16))
	tu.AssertEqual(t, line.Box().ContentBoxY(), Fl(16+16)) // p content  + div padding
	assertText(t, unpack1(line), "stu vwx")
}

func TestPageBreaks1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	tu.AssertEqual(t, len(pages), 2)

	var positionsY [][]pr.Float
	for _, page := range pages {
		var divPos []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "div" {
				divPos = append(divPos, d.Box().PositionY)
			}
		}
		positionsY = append(positionsY, divPos)
	}
	tu.AssertEqual(t, positionsY, [][]pr.Float{{10, 70}, {10}})
}

func TestPageBreaks2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	tu.AssertEqual(t, len(pages), 2)

	var positionsY [][]pr.Float
	for _, page := range pages {
		var divPos []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "div" {
				divPos = append(divPos, d.Box().PositionY)
			}
		}
		positionsY = append(positionsY, divPos)
	}
	tu.AssertEqual(t, positionsY, [][]pr.Float{{10}, {10, 30}})
}

func TestPageBreaks3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	tu.AssertEqual(t, len(pages), 3)
	var positionsY [][]pr.Float
	for _, page := range pages {
		var divPos []pr.Float
		for _, d := range bo.Descendants(page) {
			if d.Box().ElementTag() == "div" {
				divPos = append(divPos, d.Box().PositionY)
			}
		}
		positionsY = append(positionsY, divPos)
	}
	tu.AssertEqual(t, positionsY, [][]pr.Float{{10}, {10}, {10}})
}
