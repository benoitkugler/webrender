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
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

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
	html := page.Box().Children[0]
	tu.AssertEqual(t, html.Box().ElementTag(), "html", "html")
	body := html.Box().Children[0]
	tu.AssertEqual(t, body.Box().ElementTag(), "body", "body")
	tu.AssertEqual(t, body.Box().Width, pr.Float(120), "body")

	divs := body.Box().Children

	var paragraphs []Box
	for _, div := range divs {
		tu.AssertEqual(t, bo.BlockBoxT.IsInstance(div), true, "isinstance")
		tu.AssertEqual(t, div.Box().ElementTag(), "div", "div")
		tu.AssertEqual(t, div.Box().Width, pr.Float(100), "div")
		for _, paragraph := range div.Box().Children {
			tu.AssertEqual(t, bo.BlockBoxT.IsInstance(paragraph), true, "isinstance")
			tu.AssertEqual(t, paragraph.Box().ElementTag(), "p", "paragraph")
			tu.AssertEqual(t, paragraph.Box().PaddingLeft, pr.Float(2), "paragraph")
			tu.AssertEqual(t, paragraph.Box().PaddingRight, pr.Float(2), "paragraph")
			tu.AssertEqual(t, paragraph.Box().BorderLeftWidth, pr.Float(1), "paragraph")
			tu.AssertEqual(t, paragraph.Box().BorderRightWidth, pr.Float(1), "paragraph")
			paragraphs = append(paragraphs, paragraph)
		}
	}
	tu.AssertEqual(t, len(paragraphs), 15, "len")

	// width is "auto"
	tu.AssertEqual(t, paragraphs[0].Box().Width, pr.Float(94), "paragraphs")
	tu.AssertEqual(t, paragraphs[0].Box().MarginLeft, pr.Float(0), "paragraphs")
	tu.AssertEqual(t, paragraphs[0].Box().MarginRight, pr.Float(0), "paragraphs")

	// No "auto", over-constrained equation with ltr, the initial
	// "margin-right: 0" was ignored.
	tu.AssertEqual(t, paragraphs[1].Box().Width, pr.Float(50), "paragraphs")
	tu.AssertEqual(t, paragraphs[1].Box().MarginLeft, pr.Float(0), "paragraphs")

	// No "auto", over-constrained equation with rtl, the initial
	// "margin-left: 0" was ignored.
	tu.AssertEqual(t, paragraphs[2].Box().Width, pr.Float(50), "paragraphs")
	tu.AssertEqual(t, paragraphs[2].Box().MarginRight, pr.Float(0), "paragraphs")

	// width is "auto"
	tu.AssertEqual(t, paragraphs[3].Box().Width, pr.Float(64), "paragraphs")
	tu.AssertEqual(t, paragraphs[3].Box().MarginLeft, pr.Float(20), "paragraphs")

	// margin-right is "auto"
	tu.AssertEqual(t, paragraphs[4].Box().Width, pr.Float(50), "paragraphs")
	tu.AssertEqual(t, paragraphs[4].Box().MarginLeft, pr.Float(20), "paragraphs")

	// margin-left is "auto"
	tu.AssertEqual(t, paragraphs[5].Box().Width, pr.Float(50), "paragraphs")
	tu.AssertEqual(t, paragraphs[5].Box().MarginLeft, pr.Float(24), "paragraphs")

	// Both margins are "auto", remaining space is split := range half
	tu.AssertEqual(t, paragraphs[6].Box().Width, pr.Float(50), "paragraphs")
	tu.AssertEqual(t, paragraphs[6].Box().MarginLeft, pr.Float(22), "paragraphs")

	// width is "auto", other "auto" are set to 0
	tu.AssertEqual(t, paragraphs[7].Box().Width, pr.Float(74), "paragraphs")
	tu.AssertEqual(t, paragraphs[7].Box().MarginLeft, pr.Float(20), "paragraphs")

	// width is "auto", other "auto" are set to 0
	tu.AssertEqual(t, paragraphs[8].Box().Width, pr.Float(74), "paragraphs")
	tu.AssertEqual(t, paragraphs[8].Box().MarginLeft, pr.Float(0), "paragraphs")

	// width is "auto", other "auto" are set to 0
	tu.AssertEqual(t, paragraphs[9].Box().Width, pr.Float(94), "paragraphs")
	tu.AssertEqual(t, paragraphs[9].Box().MarginLeft, pr.Float(0), "paragraphs")

	// sum of non-auto initially is too wide, set auto values to 0
	tu.AssertEqual(t, paragraphs[10].Box().Width, pr.Float(200), "paragraphs")
	tu.AssertEqual(t, paragraphs[10].Box().MarginLeft, pr.Float(0), "paragraphs")

	// Constrained by min-width, same as above
	tu.AssertEqual(t, paragraphs[11].Box().Width, pr.Float(200), "paragraphs")
	tu.AssertEqual(t, paragraphs[11].Box().MarginLeft, pr.Float(0), "paragraphs")

	// Constrained by max-width, same as paragraphs[6]
	tu.AssertEqual(t, paragraphs[12].Box().Width, pr.Float(50), "paragraphs")
	tu.AssertEqual(t, paragraphs[12].Box().MarginLeft, pr.Float(22), "paragraphs")

	// NOT constrained by min-width
	tu.AssertEqual(t, paragraphs[13].Box().Width, pr.Float(94), "paragraphs")
	tu.AssertEqual(t, paragraphs[13].Box().MarginLeft, pr.Float(0), "paragraphs")

	// 70%
	tu.AssertEqual(t, paragraphs[14].Box().Width, pr.Float(70), "paragraphs")
	tu.AssertEqual(t, paragraphs[14].Box().MarginLeft, pr.Float(0), "paragraphs")
}

func TestBlockHeightsP(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

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
	html := page.Box().Children[0]
	body := html.Box().Children[0]

	var heights []pr.Float
	for _, div := range body.Box().Children {
		heights = append(heights, div.Box().Height.V())
	}
	tu.AssertEqual(t, heights, []pr.Float{90, 90 * 3, 20, 120, 20, 120, 90, 90}, "heights")
}

func TestBlockHeightsImg(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	var heights []pr.Float
	for _, div := range body.Box().Children {
		heights = append(heights, div.Box().Height.V())
	}
	tu.AssertEqual(t, heights, []pr.Float{40, 20, 20, 20, 20, 20}, "heights")
}

func TestBlockHeightsImgNoBodyHeight(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	var heights []pr.Float
	for _, div := range body.Box().Children {
		heights = append(heights, div.Box().Height.V())
	}
	tu.AssertEqual(t, heights, []pr.Float{40, 40, 20, 40, 20, 0}, "heights")
}

func TestBlockPercentageHeightsNoHtmlHeight(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        html, body { margin: 0 }
        body { height: 50% }
      </style>
    `)
	html := page.Box().Children[0]
	tu.AssertEqual(t, html.Box().ElementTag(), "html", "html")
	body := html.Box().Children[0]
	tu.AssertEqual(t, body.Box().ElementTag(), "body", "body")

	// Since html’s height depend on body’s, body’s 50% means "auto"
	tu.AssertEqual(t, body.Box().Height, pr.Float(0), "body")
}

func TestBlockPercentageHeights(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        html, body { margin: 0 }
        html { height: 300px }
        body { height: 50% }
      </style>
    `)
	html := page.Box().Children[0]
	tu.AssertEqual(t, html.Box().ElementTag(), "html", "html")
	body := html.Box().Children[0]
	tu.AssertEqual(t, body.Box().ElementTag(), "body", "body")

	// This time the percentage makes sense
	tu.AssertEqual(t, body.Box().Height, pr.Float(150), "body")
}

func TestBoxSizing(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

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
	// http://www.w3.org/TR/css3-ui/#box-sizing
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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div1, div2, div3, div4 := unpack4(body)
	for _, div := range []Box{div1, div2} {
		tu.AssertEqual(t, div.Box().Style.GetBoxSizing(), pr.String("content-box"), "div")
		tu.AssertEqual(t, div.Box().Width, pr.Float(1000), "div")
		tu.AssertEqual(t, div.Box().Height, pr.Float(1000), "div")
		tu.AssertEqual(t, div.Box().PaddingWidth(), pr.Float(1020), "div")
		tu.AssertEqual(t, div.Box().PaddingHeight(), pr.Float(1020), "div")
		tu.AssertEqual(t, div.Box().BorderWidth(), pr.Float(1022), "div")
		tu.AssertEqual(t, div.Box().BorderHeight(), pr.Float(1022), "div")
		tu.AssertEqual(t, div.Box().MarginHeight(), pr.Float(1222), "div")
		// marginWidth() is the width of the containing block
	}
	// padding-box
	tu.AssertEqual(t, div3.Box().Style.GetBoxSizing(), pr.String("padding-box"), "div3")
	tu.AssertEqual(t, div3.Box().Width, pr.Float(980), "div3") // 1000 - 20
	tu.AssertEqual(t, div3.Box().Height, pr.Float(980), "div3")
	tu.AssertEqual(t, div3.Box().PaddingWidth(), pr.Float(1000), "div3")
	tu.AssertEqual(t, div3.Box().PaddingHeight(), pr.Float(1000), "div3")
	tu.AssertEqual(t, div3.Box().BorderWidth(), pr.Float(1002), "div3")
	tu.AssertEqual(t, div3.Box().BorderHeight(), pr.Float(1002), "div3")
	tu.AssertEqual(t, div3.Box().MarginHeight(), pr.Float(1202), "div3")

	// border-box
	tu.AssertEqual(t, div4.Box().Style.GetBoxSizing(), pr.String("border-box"), "div4")
	tu.AssertEqual(t, div4.Box().Width, pr.Float(978), "div4") // 1000 - 20 - 2
	tu.AssertEqual(t, div4.Box().Height, pr.Float(978), "div4")
	tu.AssertEqual(t, div4.Box().PaddingWidth(), pr.Float(998), "div4")
	tu.AssertEqual(t, div4.Box().PaddingHeight(), pr.Float(998), "div4")
	tu.AssertEqual(t, div4.Box().BorderWidth(), pr.Float(1000), "div4")
	tu.AssertEqual(t, div4.Box().BorderHeight(), pr.Float(1000), "div4")
	tu.AssertEqual(t, div4.Box().MarginHeight(), pr.Float(1200), "div4")
}

func TestBoxSizingZero(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, size := range []string{
		"width: 0; height: 0",
		"max-width: 0; max-height: 0",
		"min-width: 0; min-height: 0; width: 0; height: 0",
	} {
		testBoxSizingZero(t, size)
	}
}

func testBoxSizingZero(t *testing.T, size string) {
	// http://www.w3.org/TR/css3-ui/#box-sizing
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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	for _, div := range body.Box().Children {
		tu.AssertEqual(t, div.Box().Width, pr.Float(0), "div")
		tu.AssertEqual(t, div.Box().Height, pr.Float(0), "div")
		tu.AssertEqual(t, div.Box().PaddingWidth(), pr.Float(20), "div")
		tu.AssertEqual(t, div.Box().PaddingHeight(), pr.Float(20), "div")
		tu.AssertEqual(t, div.Box().BorderWidth(), pr.Float(22), "div")
		tu.AssertEqual(t, div.Box().BorderHeight(), pr.Float(22), "div")
		tu.AssertEqual(t, div.Box().MarginHeight(), pr.Float(222), "div")
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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		p1, p2 := unpack2(body)
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result, "p2Top")
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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		div, p2 := unpack2(body)
		p1 := div.Box().Children[0]
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result, "p2Top")
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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		p1, div := unpack2(body)
		p2 := div.Box().Children[0]
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result, "p2Top")
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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		p1, div1 := unpack2(body)
		div2 := div1.Box().Children[0]
		p2 := div2.Box().Children[0]
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result, "p2Top")
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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		p1, div1 := unpack2(body)
		div2 := div1.Box().Children[0]
		p2 := div2.Box().Children[0]
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		// Parent and element edge are the same:
		tu.AssertEqual(t, div1.Box().BorderBoxY(), p2.Box().BorderBoxY(), "div1")
		tu.AssertEqual(t, div2.Box().BorderBoxY(), p2.Box().BorderBoxY(), "div2")
		tu.AssertEqual(t, p2Top-p1Bottom, data.result, "p2Top")
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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		p1, div1 := unpack2(body)
		div2 := div1.Box().Children[0]
		p2 := div2.Box().Children[0]
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result, "p2Top")
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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		p1, _, p2 := unpack3(body)
		p1Bottom := p1.Box().ContentBoxY() + p1.Box().Height.V()
		p2Top := p2.Box().ContentBoxY()
		tu.AssertEqual(t, p2Top-p1Bottom, data.result, "p2Top")
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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		p1 := body.Box().Children[0]
		p1Top := p1.Box().ContentBoxY()
		// Vertical space from y=0
		tu.AssertEqual(t, p1Top, data.result, "p1Top")
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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		div := body.Box().Children[0]
		p1 := div.Box().Children[0]
		p1Top := p1.Box().ContentBoxY()
		// Vertical space from y=0
		tu.AssertEqual(t, p1Top, data.result, "p1Top")
	}
}

func TestBoxDecorationBreakBlockSlice(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// http://www.w3.org/TR/css3-background/#the-box-decoration-break
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
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	img1, img2 := unpack2(paragraph)
	tu.AssertEqual(t, paragraph.Box().PositionY, pr.Float(0), "paragraph")
	tu.AssertEqual(t, paragraph.Box().MarginTop, pr.Float(5), "paragraph")
	tu.AssertEqual(t, paragraph.Box().BorderTopWidth, pr.Float(3), "paragraph")
	tu.AssertEqual(t, paragraph.Box().PaddingTop, pr.Float(2), "paragraph")
	tu.AssertEqual(t, paragraph.Box().ContentBoxY(), pr.Float(10), "paragraph")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(10), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(50), "img2")
	tu.AssertEqual(t, paragraph.Box().Height, pr.Float(90), "paragraph")
	tu.AssertEqual(t, paragraph.Box().MarginBottom, pr.Float(0), "paragraph")
	tu.AssertEqual(t, paragraph.Box().BorderBottomWidth, pr.Float(0), "paragraph")
	tu.AssertEqual(t, paragraph.Box().PaddingBottom, pr.Float(0), "paragraph")
	tu.AssertEqual(t, paragraph.Box().MarginHeight(), pr.Float(100), "paragraph")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	paragraph = body.Box().Children[0]
	img1, img2 = unpack2(paragraph)
	tu.AssertEqual(t, paragraph.Box().PositionY, pr.Float(0), "paragraph")
	tu.AssertEqual(t, paragraph.Box().MarginTop, pr.Float(0), "paragraph")
	tu.AssertEqual(t, paragraph.Box().BorderTopWidth, pr.Float(0), "paragraph")
	tu.AssertEqual(t, paragraph.Box().PaddingTop, pr.Float(0), "paragraph")
	tu.AssertEqual(t, paragraph.Box().ContentBoxY(), pr.Float(0), "paragraph")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(0), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(40), "img2")
	tu.AssertEqual(t, paragraph.Box().Height, pr.Float(80), "paragraph")
	tu.AssertEqual(t, paragraph.Box().PaddingBottom, pr.Float(2), "paragraph")
	tu.AssertEqual(t, paragraph.Box().BorderBottomWidth, pr.Float(3), "paragraph")
	tu.AssertEqual(t, paragraph.Box().MarginBottom, pr.Float(5), "paragraph")
	tu.AssertEqual(t, paragraph.Box().MarginHeight(), pr.Float(90), "paragraph")
}

func TestBoxDecorationBreakBlockClone(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// http://www.w3.org/TR/css3-background/#the-box-decoration-break
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
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	img1, img2 := unpack2(paragraph)
	tu.AssertEqual(t, paragraph.Box().PositionY, pr.Float(0), "paragraph")
	tu.AssertEqual(t, paragraph.Box().MarginTop, pr.Float(5), "paragraph")
	tu.AssertEqual(t, paragraph.Box().BorderTopWidth, pr.Float(3), "paragraph")
	tu.AssertEqual(t, paragraph.Box().PaddingTop, pr.Float(2), "paragraph")
	tu.AssertEqual(t, paragraph.Box().ContentBoxY(), pr.Float(10), "paragraph")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(10), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(50), "img2")
	tu.AssertEqual(t, paragraph.Box().Height, pr.Float(80), "paragraph")
	// TODO: bottom margin should be 0
	// https://www.w3.org/TR/css-break-3/#valdef-box-decoration-break-clone
	// "Cloned margins are truncated on block-level boxes."
	// See https://github.com/Kozea/WeasyPrint/issues/115
	tu.AssertEqual(t, paragraph.Box().MarginBottom, pr.Float(5), "paragraph")
	tu.AssertEqual(t, paragraph.Box().BorderBottomWidth, pr.Float(3), "paragraph")
	tu.AssertEqual(t, paragraph.Box().PaddingBottom, pr.Float(2), "paragraph")
	tu.AssertEqual(t, paragraph.Box().MarginHeight(), pr.Float(100), "paragraph")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	paragraph = body.Box().Children[0]
	img1, img2 = unpack2(paragraph)
	tu.AssertEqual(t, paragraph.Box().PositionY, pr.Float(0), "paragraph")
	tu.AssertEqual(t, paragraph.Box().MarginTop, pr.Float(0), "paragraph")
	tu.AssertEqual(t, paragraph.Box().BorderTopWidth, pr.Float(3), "paragraph")
	tu.AssertEqual(t, paragraph.Box().PaddingTop, pr.Float(2), "paragraph")
	tu.AssertEqual(t, paragraph.Box().ContentBoxY(), pr.Float(5), "paragraph")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(5), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(45), "img2")
	tu.AssertEqual(t, paragraph.Box().Height, pr.Float(80), "paragraph")
	tu.AssertEqual(t, paragraph.Box().PaddingBottom, pr.Float(2), "paragraph")
	tu.AssertEqual(t, paragraph.Box().BorderBottomWidth, pr.Float(3), "paragraph")
	tu.AssertEqual(t, paragraph.Box().MarginBottom, pr.Float(5), "paragraph")
	tu.AssertEqual(t, paragraph.Box().MarginHeight(), pr.Float(95), "paragraph")
}

func TestBoxDecorationBreakCloneBottomPadding(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

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
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	tu.AssertEqual(t, article.Box().Height, pr.Float(80-2*12), "article")
	div1, div2 := unpack2(article)
	tu.AssertEqual(t, div1.Box().PositionY, pr.Float(12), "div1")
	tu.AssertEqual(t, div2.Box().PositionY, pr.Float(12+20), "div2")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	article = body.Box().Children[0]
	tu.AssertEqual(t, article.Box().Height, pr.Float(20), "article")
	div := article.Box().Children[0]
	tu.AssertEqual(t, div.Box().PositionY, pr.Float(12), "div")
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
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     article := body.Box().Children[0]
//     tu.AssertEqual(t, article.Box().Height , 80 - 12, "article")
//     div1, div2 = article.Box().Children
//     tu.AssertEqual(t, div1.Box().PositionY , 12, "div1")
//     tu.AssertEqual(t, div2.Box().PositionY , 12 + 20, "div2")

//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     article := body.Box().Children[0]
//     tu.AssertEqual(t, article.Box().Height , 20, "article")
//     div := article.Box().Children[0]
//     tu.AssertEqual(t, div.Box().PositionY , 0, "div")

func TestOverflowAuto(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="overflow: auto">
        <div style="float: left; height: 50px; margin: 10px">bla bla bla</div>
          toto toto`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	tu.AssertEqual(t, article.Box().Height, pr.Float(50+10+10), "article")
}

// Test regression: https://github.com/Kozea/WeasyPrint/issues/943
func TestBoxMarginTopRepagination(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

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
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	_, div := unpack2(body)
	tu.AssertEqual(t, div.Box().MarginTop, pr.Float(20), "div")
	tu.AssertEqual(t, div.Box().PaddingBoxY(), pr.Float(10+20), "div")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	div, _ = unpack2(body)
	tu.AssertEqual(t, div.Box().MarginTop, pr.Float(0), "div")
	tu.AssertEqual(t, div.Box().PaddingBoxY(), pr.Float(0), "div")
}

func TestContinueDiscard(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

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
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	tu.AssertEqual(t, article.Box().Height, pr.Float(3*25), "article")
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().PositionY, pr.Float(1), "div1")
	tu.AssertEqual(t, div2.Box().PositionY, pr.Float(1+25), "div2")
	tu.AssertEqual(t, div3.Box().PositionY, pr.Float(1+25*2), "div3")
	tu.AssertEqual(t, article.Box().BorderBottomWidth, pr.Float(1), "article")
}

func TestContinueDiscardChildren(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

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
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	tu.AssertEqual(t, article.Box().Height, pr.Float(2+3*25), "article")
	section := article.Box().Children[0]
	tu.AssertEqual(t, section.Box().Height, pr.Float(3*25), "section")
	div1, div2, div3 := unpack3(section)
	tu.AssertEqual(t, div1.Box().PositionY, pr.Float(2), "div1")
	tu.AssertEqual(t, div2.Box().PositionY, pr.Float(2+25), "div2")
	tu.AssertEqual(t, div3.Box().PositionY, pr.Float(2+25*2), "div3")
	tu.AssertEqual(t, article.Box().BorderBottomWidth, pr.Float(1), "article")
}

func TestBlockInBlockWithBottomPadding(t *testing.T) {
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1476
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 8em 3.5em }
        body { line-height: 1; orphans: 1; widows: 1; font-family: weasyprint }
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

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	anonBody, div := unpack2(body)
	line := anonBody.Box().Children[0]
	tu.AssertEqual(t, line.Box().Height, pr.Float(16), "line")
	tu.AssertEqual(t, line.Box().Children[0].(*bo.TextBox).Text, "abc def", "line")
	p := div.Box().Children[0]
	line = p.Box().Children[0]
	tu.AssertEqual(t, line.Box().Height, pr.Float(16), "line")
	tu.AssertEqual(t, line.Box().Children[0].(*bo.TextBox).Text, "ghi jkl", "line")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	div, anonBody = unpack2(body)
	p = div.Box().Children[0]
	line = p.Box().Children[0]
	tu.AssertEqual(t, line.Box().Height, pr.Float(16), "line")
	tu.AssertEqual(t, line.Box().Children[0].(*bo.TextBox).Text, "mno pqr", "line")
	line = anonBody.Box().Children[0]
	tu.AssertEqual(t, line.Box().Height, pr.Float(16), "line")
	tu.AssertEqual(t, line.Box().ContentBoxY(), pr.Float(16+16), "line") // p content + div padding
	tu.AssertEqual(t, line.Box().Children[0].(*bo.TextBox).Text, "stu vwx", "line")
}
