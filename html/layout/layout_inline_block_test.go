package layout

import (
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for inline blocks layout.

func TestInlineBlockSizes(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { margin: 0; size: 200px 2000px }
        body { margin: 0 }
        div { display: inline-block; }
      </style>
      <div> </div>
      <div>a</div>
      <div style="margin: 10px; height: 100px"></div>
      <div style="margin-left: 10px; margin-top: -50px;
                  padding-right: 20px;"></div>
      <div>
        Ipsum dolor sit amet,
        consectetur adipiscing elit.
        Sed sollicitudin nibh
        et turpis molestie tristique.
      </div>
      <div style="width: 100px; height: 100px;
                  padding-left: 10px; margin-right: 10px;
                  margin-top: -10px; margin-bottom: 50px"></div>
      <div style="font-size: 0">
        <div style="min-width: 10px; height: 10px"></div>
        <div style="width: 10%">
          <div style="width: 10px; height: 10px"></div>
        </div>
      </div>
      <div style="min-width: 150px">foo</div>
      <div style="max-width: 10px
        ">Supercalifragilisticexpialidocious</div>`)
	html := page.Box().Children[0]
	tu.AssertEqual(t, html.Box().ElementTag(), "html", "html")
	body := html.Box().Children[0]
	tu.AssertEqual(t, body.Box().ElementTag(), "body", "body")
	tu.AssertEqual(t, body.Box().Width, pr.Float(200), "body")

	line1, line2, line3, line4 := unpack4(body)

	// First line:
	// White space in-between divs ends up preserved in TextBoxes
	div1, _, div2, _, div3, _, div4, _ := unpack8(line1)

	// First div, one ignored space collapsing with next space
	tu.AssertEqual(t, div1.Box().ElementTag(), "div", "div1")
	tu.AssertEqual(t, div1.Box().Width, pr.Float(0), "div1")

	// Second div, one letter
	tu.AssertEqual(t, div2.Box().ElementTag(), "div", "div2")
	tu.AssertEqual(t, 0 < div2.Box().Width.V(), true, "0")
	tu.AssertEqual(t, div2.Box().Width.V() < pr.Float(20), true, "0")

	// Third div, empty with margin
	tu.AssertEqual(t, div3.Box().ElementTag(), "div", "div3")
	tu.AssertEqual(t, div3.Box().Width, pr.Float(0), "div3")
	tu.AssertEqual(t, div3.Box().MarginWidth(), pr.Float(20), "div3")
	tu.AssertEqual(t, div3.Box().Height, pr.Float(100), "div3")

	// Fourth div, empty with margin && padding
	tu.AssertEqual(t, div4.Box().ElementTag(), "div", "div4")
	tu.AssertEqual(t, div4.Box().Width, pr.Float(0), "div4")
	tu.AssertEqual(t, div4.Box().MarginWidth(), pr.Float(30), "div4")

	// Second line :
	div5, _ := unpack2(line2)

	// Fifth div, long text, full-width div
	tu.AssertEqual(t, div5.Box().ElementTag(), "div", "div5")
	tu.AssertEqual(t, len(div5.Box().Children) > 1, true, "len")
	tu.AssertEqual(t, div5.Box().Width, pr.Float(200), "div5")

	// Third line :
	div6, _, div7, _ := unpack4(line3)

	// Sixth div, empty div with fixed width && height
	tu.AssertEqual(t, div6.Box().ElementTag(), "div", "div6")
	tu.AssertEqual(t, div6.Box().Width, pr.Float(100), "div6")
	tu.AssertEqual(t, div6.Box().MarginWidth(), pr.Float(120), "div6")
	tu.AssertEqual(t, div6.Box().Height, pr.Float(100), "div6")
	tu.AssertEqual(t, div6.Box().MarginHeight(), pr.Float(140), "div6")

	// Seventh div
	tu.AssertEqual(t, div7.Box().ElementTag(), "div", "div7")
	tu.AssertEqual(t, div7.Box().Width, pr.Float(20), "div7")
	childLine := div7.Box().Children[0]
	// Spaces have font-size: 0, they get removed
	childDiv1, childDiv2 := unpack2(childLine)
	tu.AssertEqual(t, childDiv1.Box().ElementTag(), "div", "childDiv1")
	tu.AssertEqual(t, childDiv1.Box().Width, pr.Float(10), "childDiv1")
	tu.AssertEqual(t, childDiv2.Box().ElementTag(), "div", "childDiv2")
	tu.AssertEqual(t, childDiv2.Box().Width, pr.Float(2), "childDiv2")
	grandchild := childDiv2.Box().Children[0]
	tu.AssertEqual(t, grandchild.Box().ElementTag(), "div", "grandchild")
	tu.AssertEqual(t, grandchild.Box().Width, pr.Float(10), "grandchild")

	div8, _, div9 := unpack3(line4)
	tu.AssertEqual(t, div8.Box().Width, pr.Float(150), "div8")
	tu.AssertEqual(t, div9.Box().Width, pr.Float(10), "div9")
}

func TestInlineBlockWithMargin(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1235
	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 100px }
        span { font-family: weasyprint; display: inline-block; margin: 0 30px }
      </style>
      <span>a b c d e f g h i j k l</span>`)
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	line1 := body.Box().Children[0]
	span := line1.Box().Children[0]
	tu.AssertEqual(t, span.Box().Width, pr.Float(40), "span") // 100 - 2 * 30
}
