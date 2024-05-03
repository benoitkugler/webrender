package layout

import (
	"testing"

	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for inline blocks layout.

func TestInlineBlockSizes(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	tu.AssertEqual(t, html.Box().ElementTag(), "html")
	body := unpack1(html)
	tu.AssertEqual(t, body.Box().ElementTag(), "body")
	tu.AssertEqual(t, body.Box().Width, Fl(200))

	line1, line2, line3, line4 := unpack4(body)

	// First line:
	// White space in-between divs ends up preserved in TextBoxes
	div1, _, div2, _, div3, _, div4, _ := unpack8(line1)

	// First div, one ignored space collapsing with next space
	tu.AssertEqual(t, div1.Box().ElementTag(), "div")
	tu.AssertEqual(t, div1.Box().Width, Fl(0))

	// Second div, one letter
	tu.AssertEqual(t, div2.Box().ElementTag(), "div")
	tu.AssertEqual(t, 0 < div2.Box().Width.V(), true)
	tu.AssertEqual(t, div2.Box().Width.V() < Fl(20), true)

	// Third div, empty with margin
	tu.AssertEqual(t, div3.Box().ElementTag(), "div")
	tu.AssertEqual(t, div3.Box().Width, Fl(0))
	tu.AssertEqual(t, div3.Box().MarginWidth(), Fl(20))
	tu.AssertEqual(t, div3.Box().Height, Fl(100))

	// Fourth div, empty with margin && padding
	tu.AssertEqual(t, div4.Box().ElementTag(), "div")
	tu.AssertEqual(t, div4.Box().Width, Fl(0))
	tu.AssertEqual(t, div4.Box().MarginWidth(), Fl(30))

	// Second line :
	div5, _ := unpack2(line2)

	// Fifth div, long text, full-width div
	tu.AssertEqual(t, div5.Box().ElementTag(), "div")
	tu.AssertEqual(t, len(div5.Box().Children) > 1, true)
	tu.AssertEqual(t, div5.Box().Width, Fl(200))

	// Third line :
	div6, _, div7, _ := unpack4(line3)

	// Sixth div, empty div with fixed width && height
	tu.AssertEqual(t, div6.Box().ElementTag(), "div")
	tu.AssertEqual(t, div6.Box().Width, Fl(100))
	tu.AssertEqual(t, div6.Box().MarginWidth(), Fl(120))
	tu.AssertEqual(t, div6.Box().Height, Fl(100))
	tu.AssertEqual(t, div6.Box().MarginHeight(), Fl(140))

	// Seventh div
	tu.AssertEqual(t, div7.Box().ElementTag(), "div")
	tu.AssertEqual(t, div7.Box().Width, Fl(20))
	childLine := unpack1(div7)
	// Spaces have font-size: 0, they get removed
	childDiv1, childDiv2 := unpack2(childLine)
	tu.AssertEqual(t, childDiv1.Box().ElementTag(), "div")
	tu.AssertEqual(t, childDiv1.Box().Width, Fl(10))
	tu.AssertEqual(t, childDiv2.Box().ElementTag(), "div")
	tu.AssertEqual(t, childDiv2.Box().Width, Fl(2))
	grandchild := unpack1(childDiv2)
	tu.AssertEqual(t, grandchild.Box().ElementTag(), "div")
	tu.AssertEqual(t, grandchild.Box().Width, Fl(10))

	div8, _, div9 := unpack3(line4)
	tu.AssertEqual(t, div8.Box().Width, Fl(150))
	tu.AssertEqual(t, div9.Box().Width, Fl(10))
}

func TestInlineBlockWithMargin(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1235
	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 100px }
        span { font-family: weasyprint; display: inline-block; margin: 0 30px }
      </style>
      <span>a b c d e f g h i j k l</span>`)
	html := unpack1(page1)
	body := unpack1(html)
	line1 := unpack1(body)
	span := unpack1(line1)
	tu.AssertEqual(t, span.Box().Width, Fl(40)) // 100 - 2 * 30
}
