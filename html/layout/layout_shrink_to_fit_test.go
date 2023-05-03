package layout

import (
	"fmt"
	"strings"
	"testing"

	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

//  Tests for shrink-to-fit algorithm.

func TestShrinkToFitFloatingPointError1(t *testing.T) {
	for marginLeft := 1; marginLeft < 10; marginLeft += 2 {
		for fontSize := 1; fontSize < 10; fontSize += 2 {
			testShrinkToFitFloatingPointError1(t, marginLeft, fontSize)
		}
	}
}

func testShrinkToFitFloatingPointError1(t *testing.T, marginLeft, fontSize int) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// See bugs #325 && #288, see commit fac5ee9.
	page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 100000px 100px }
        p { float: left; margin-left: 0.%din; font-size: 0.%dem;
            font-family: weasyprint }
      </style>
      <p>this parrot is dead</p>
    `, marginLeft, fontSize))
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	p := body.Box().Children[0]
	tu.AssertEqual(t, len(p.Box().Children), 1, fmt.Sprintf("margin: %d font size: %d", marginLeft, fontSize))
}

func TestShrinkToFitFloatingPointError2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, fontSize := range []int{1, 5, 10, 50, 100, 1000, 10000} {
		letters := 1
		for {
			page := renderOnePage(t, fmt.Sprintf(`
          <style>
            @font-face { src: url(weasyprint.otf); font-family: weasyprint }
            @page { size: %d0pt %d0px }
            p { font-size: %dpt; font-family: weasyprint }
          </style>
          <p>mmm <b>%s a</b></p>
        `, fontSize, fontSize, fontSize, strings.Repeat("i", letters)))
			html := page.Box().Children[0]
			body := html.Box().Children[0]
			p := body.Box().Children[0]
			tu.AssertEqual(t, len(p.Box().Children) == 1 || len(p.Box().Children) == 2, true, "")
			tu.AssertEqual(t, len(p.Box().Children[0].Box().Children), 2, "")
			text := p.Box().Children[0].Box().Children[1].Box().Children[0].(*bo.TextBox).Text
			tu.AssertEqual(t, len(text) > 0, true, "")
			if strings.HasSuffix(text, "i") {
				break
			} else {
				letters += 1
			}
		}
	}
}
