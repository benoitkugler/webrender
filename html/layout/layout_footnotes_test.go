package layout

import (
	"fmt"
	"sort"
	"testing"

	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for footnotes layout.

func TestInlineFootnote(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>abc<span>de</span></div>`)
	html, footnoteArea := unpack2(page)
	body := unpack1(html)
	div := unpack1(body)
	divTextbox, footnoteCall := unpack2(unpack1(div))
	assertText(t, divTextbox, "abc")
	assertText(t, unpack1(footnoteCall), "1")
	tu.AssertEqual(t, divTextbox.Box().PositionY, Fl(0))

	footnoteMarker, footnoteTextbox := unpack2(unpack1(unpack1(footnoteArea)))
	assertText(t, unpack1(footnoteMarker), "1.")
	assertText(t, footnoteTextbox, "de")
	tu.AssertEqual(t, footnoteArea.Box().PositionY, Fl(5))
}

func TestBlockFootnote(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
         @font-face {src: url(weasyprint.otf); font-family: weasyprint}
         @page {
             size: 9px 7px;
         }
         div {
             font-family: weasyprint;
             font-size: 2px;
             line-height: 1;
         }
         div.footnote {
             float: footnote;
         }
        </style>
        <div>abc<div class="footnote">de</div></div>`)
	html, footnoteArea := unpack2(page)
	body := unpack1(html)
	div := unpack1(body)
	divTextbox, footnoteCall := unpack2(unpack1(div))
	assertText(t, divTextbox, "abc")
	assertText(t, unpack1(footnoteCall), "1")
	tu.AssertEqual(t, divTextbox.Box().PositionY, Fl(0))
	footnoteMarker, footnoteTextbox := unpack2(unpack1(unpack1(footnoteArea)))
	assertText(t, unpack1(footnoteMarker), "1.")
	assertText(t, footnoteTextbox, "de")
	tu.AssertEqual(t, footnoteArea.Box().PositionY, Fl(5))
}

func TestLongFootnote(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>abc<span>de f</span></div>`)
	html, footnoteArea := unpack2(page)
	body := unpack1(html)
	div := unpack1(body)
	divTextbox, footnoteCall := unpack2(unpack1(div))
	assertText(t, divTextbox, "abc")
	assertText(t, unpack1(footnoteCall), "1")
	tu.AssertEqual(t, divTextbox.Box().PositionY, Fl(0))
	footnoteLine1, footnoteLine2 := unpack2(unpack1(footnoteArea))
	footnoteMarker, footnoteContent1 := unpack2(footnoteLine1)
	footnoteContent2 := unpack1(footnoteLine2)
	assertText(t, unpack1(footnoteMarker), "1.")
	assertText(t, footnoteContent1, "de")
	tu.AssertEqual(t, footnoteArea.Box().PositionY, Fl(3))
	assertText(t, footnoteContent2, "f")
	tu.AssertEqual(t, footnoteContent2.Box().PositionY, Fl(5))
}

func TestSeveralFootnote(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>abcd e<span>fg</span> hijk l<span>mn</span></div>`)
	page1, page2 := pages[0], pages[1]

	html1, footnoteArea1 := unpack2(page1)
	body1 := unpack1(html1)
	div1 := unpack1(body1)
	div1Line1, _ := unpack2(div1)
	assertText(t, unpack1(div1Line1), "abcd")
	div1Line2Text, div1_footnoteCall := unpack2(div1.Box().Children[1])
	assertText(t, div1Line2Text, "e")
	assertText(t, unpack1(div1_footnoteCall), "1")
	footnoteMarker1, footnoteTextbox1 := unpack2(unpack1(unpack1(footnoteArea1)))
	assertText(t, unpack1(footnoteMarker1), "1.")
	assertText(t, footnoteTextbox1, "fg")

	html2, footnoteArea2 := unpack2(page2)
	body2 := unpack1(html2)
	div2 := unpack1(body2)
	div2Line1, _ := unpack2(div2)
	assertText(t, unpack1(div2Line1), "hijk")
	div2Line2Text, div2_footnoteCall := unpack2(div2.Box().Children[1])
	assertText(t, div2Line2Text, "l")
	assertText(t, unpack1(div2_footnoteCall), "2")
	footnoteMarker2, footnoteTextbox2 := unpack2(unpack1(unpack1(footnoteArea2)))
	assertText(t, unpack1(footnoteMarker2), "2.")
	assertText(t, footnoteTextbox2, "mn")
}

func TestReportedFootnote_1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>abc<span>f1</span> hij<span>f2</span></div>`)
	page1, page2 := pages[0], pages[1]

	html1, footnoteArea1 := unpack2(page1)
	body1 := unpack1(html1)
	div1 := unpack1(body1)
	div_line1, div_line2 := unpack2(div1)
	div_line1_text, div_footnoteCall1 := unpack2(div_line1)
	assertText(t, div_line1_text, "abc")
	assertText(t, unpack1(div_footnoteCall1), "1")
	div_line2_text, div_footnoteCall2 := unpack2(div_line2)
	assertText(t, div_line2_text, "hij")
	assertText(t, unpack1(div_footnoteCall2), "2")

	footnoteMarker1, footnoteTextbox1 := unpack2(unpack1(footnoteArea1.Box().Children[0]))
	assertText(t, unpack1(footnoteMarker1), "1.")
	assertText(t, footnoteTextbox1, "f1")

	html2, footnoteArea2 := unpack2(page2)
	tu.AssertEqual(t, len(html2.Box().Children), 0)
	footnoteMarker2, footnoteTextbox2 := unpack2(unpack1(unpack1(footnoteArea2)))
	assertText(t, unpack1(footnoteMarker2), "2.")
	assertText(t, footnoteTextbox2, "f2")
}

func TestReportedFootnote_2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>abc<span>f1</span> hij<span>f2</span> wow</div>`)
	page1, page2 := pages[0], pages[1]

	html1, footnoteArea1 := unpack2(page1)
	body1 := unpack1(html1)
	div1 := unpack1(body1)
	div_line1, div_line2 := unpack2(div1)
	div_line1_text, div_footnoteCall1 := unpack2(div_line1)
	assertText(t, div_line1_text, "abc")
	assertText(t, unpack1(div_footnoteCall1), "1")
	div_line2_text, div_footnoteCall2 := unpack2(div_line2)
	assertText(t, div_line2_text, "hij")
	assertText(t, unpack1(div_footnoteCall2), "2")
	footnoteMarker1, footnoteTextbox1 := unpack2(unpack1(footnoteArea1.Box().Children[0]))
	assertText(t, unpack1(footnoteMarker1), "1.")
	assertText(t, footnoteTextbox1, "f1")

	html2, footnoteArea2 := unpack2(page2)
	body2 := unpack1(html2)
	div2 := unpack1(body2)
	div2_line := unpack1(div2)
	assertText(t, unpack1(div2_line), "wow")
	footnoteMarker2, footnoteTextbox2 := unpack2(unpack1(unpack1(footnoteArea2)))
	assertText(t, unpack1(footnoteMarker2), "2.")
	assertText(t, footnoteTextbox2, "f2")
}

func TestReportedFootnote_3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 10px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>
          abc<span>1</span>
          def<span>v long 2</span>
          ghi<span>3</span>
        </div>`)
	page1, page2 := pages[0], pages[1]

	html1, footnoteArea1 := unpack2(page1)
	body1 := unpack1(html1)
	div1 := unpack1(body1)
	line1, line2, line3 := unpack3(div1)
	assertText(t, unpack1(line1), "abc")
	assertText(t, unpack1(line1.Box().Children[1]), "1")
	assertText(t, unpack1(line2), "def")
	assertText(t, unpack1(line2.Box().Children[1]), "2")
	assertText(t, unpack1(line3), "ghi")
	assertText(t, unpack1(line3.Box().Children[1]), "3")
	footnote1 := unpack1(footnoteArea1)
	assertText(t, unpack1(footnote1.Box().Children[0].Box().Children[0]), "1.")
	assertText(t, unpack1(footnote1).Box().Children[1], "1")

	_, footnoteArea2 := unpack2(page2)
	footnote2, footnote3 := unpack2(footnoteArea2)
	assertText(t, unpack1(footnote2.Box().Children[0].Box().Children[0]), "2.")
	assertText(t, unpack1(footnote2).Box().Children[1], "v")
	assertText(t, unpack1(footnote2.Box().Children[1]), "long")
	assertText(t, unpack1(footnote2.Box().Children[2]), "2")
	assertText(t, unpack1(footnote3.Box().Children[0].Box().Children[0]), "3.")
	assertText(t, unpack1(footnote3).Box().Children[1], "3")
}

func TestReportedSequentialFootnote(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>
            a<span>b</span><span>c</span><span>d</span><span>e</span>
        </div>`)

	var positions nodePosList
	for _, letter := range "abcde" {
		positions = append(positions, treePosition(asBoxes(pages), func(box Box) bool {
			if box, ok := box.(*bo.TextBox); ok {
				return box.TextS() == string(letter)
			}
			return false
		}))
	}
	tu.AssertEqual(t, sort.IsSorted(positions), true)
}

func TestReportedSequentialFootnoteSecondLine(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>
            aaa a<span>b</span><span>c</span><span>d</span><span>e</span>
        </div>`)

	var positions nodePosList
	for _, letter := range "abc" {
		positions = append(positions, treePosition(asBoxes(pages), func(box Box) bool {
			if box, ok := box.(*bo.TextBox); ok {
				return box.TextS() == string(letter)
			}
			return false
		}))
	}
	tu.AssertEqual(t, sort.IsSorted(positions), true)
}

func TestFootnoteAreaAfterCall(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, test := range []struct {
		css  string
		tail string
	}{
		{"p { break-inside: avoid }", "<br>e<br>f"},
		{"p { widows: 4 }", "<br>e<br>f"},
		{"p + p { break-before: avoid }", "</p><p>e<br>f"},
		{"p + p { break-before: avoid }", "<span>y</span><span>z</span></p><p>e"},
	} {
		pages := renderPages(t, fmt.Sprintf(`
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 10px;
                margin: 0;
            }
            body {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                orphans: 2;
                widows: 2;
                margin: 0;
            }
            span {
                float: footnote;
            }
            %s
        </style>
        <div>a<br>b</div>
        <p>c<br>d<span>x</span>%s</p>`, test.css, test.tail))

		footnoteCall := treePosition(asBoxes(pages), func(box Box) bool {
			return box.Box().ElementTag() == "p::footnote-call"
		})
		footnoteArea := treePosition(asBoxes(pages), func(box Box) bool {
			_, ok := box.(*bo.FootnoteAreaBox)
			return ok
		})

		tu.AssertEqual(t, footnoteCall.isLess(footnoteArea), true)
	}
}

func TestFootnoteDisplayInline(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 50px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
                footnote-display: inline;
            }
        </style>
        <div>abc<span>d</span> fgh<span>i</span></div>`)
	html, footnoteArea := unpack2(page)
	body := unpack1(html)
	div := unpack1(body)
	div_line1, div_line2 := unpack2(div)
	div_textbox1, footnoteCall1 := unpack2(div_line1)
	divTextbox2, footnoteCall2 := unpack2(div_line2)
	assertText(t, div_textbox1, "abc")
	assertText(t, divTextbox2, "fgh")
	assertText(t, unpack1(footnoteCall1), "1")
	assertText(t, unpack1(footnoteCall2), "2")
	line := unpack1(footnoteArea)
	footnote_mark1, footnoteTextbox1 := unpack2(unpack1(line))
	footnote_mark2, footnoteTextbox2 := unpack2(line.Box().Children[1])
	assertText(t, unpack1(footnote_mark1), "1.")
	assertText(t, footnoteTextbox1, "d")
	assertText(t, unpack1(footnote_mark2), "2.")
	assertText(t, footnoteTextbox2, "i")
}

func TestFootnoteLongerThanSpaceLeft(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>abc<span>def ghi jkl</span></div>`)
	page1, page2 := pages[0], pages[1]

	html1 := unpack1(page1)
	body1 := unpack1(html1)
	div := unpack1(body1)
	divTextbox, footnoteCall := unpack2(unpack1(div))
	assertText(t, divTextbox, "abc")
	assertText(t, unpack1(footnoteCall), "1")

	html2, footnoteArea := unpack2(page2)
	tu.AssertEqual(t, len(html2.Box().Children), 0)
	footnoteLine1, footnoteLine2, footnoteLine3 := unpack3(unpack1(footnoteArea))
	footnoteMarker, footnoteContent1 := unpack2(footnoteLine1)
	footnoteContent2 := unpack1(footnoteLine2)
	footnoteContent3 := unpack1(footnoteLine3)
	assertText(t, unpack1(footnoteMarker), "1.")
	assertText(t, footnoteContent1, "def")
	assertText(t, footnoteContent2, "ghi")
	assertText(t, footnoteContent3, "jkl")
}

func TestFootnoteLongerThanPage(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Nothing is defined for this use case in the specification. In WeasyPrint,
	// the content simply overflows.
	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>abc<span>def ghi jkl mno</span></div>`)
	page1, page2 := pages[0], pages[1]

	html1 := unpack1(page1)
	body1 := unpack1(html1)
	div := unpack1(body1)
	divTextbox, footnoteCall := unpack2(unpack1(div))
	assertText(t, divTextbox, "abc")
	assertText(t, unpack1(footnoteCall), "1")

	html2, footnoteArea2 := unpack2(page2)
	tu.AssertEqual(t, len(html2.Box().Children), 0)
	footnoteLine1, footnoteLine2, footnoteLine3, footnoteLine4 := unpack4(unpack1(footnoteArea2))
	footnoteMarker1, footnoteContent1 := unpack2(footnoteLine1)
	footnoteContent2 := unpack1(footnoteLine2)
	footnoteContent3 := unpack1(footnoteLine3)
	footnoteContent4 := unpack1(footnoteLine4)
	assertText(t, unpack1(footnoteMarker1), "1.")
	assertText(t, footnoteContent1, "def")
	assertText(t, footnoteContent2, "ghi")
	assertText(t, footnoteContent3, "jkl")
	assertText(t, footnoteContent4, "mno")
}

func TestFootnotePolicyLine(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 9px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
				orphans: 2;
                widows: 2;
            }
            span {
                float: footnote;
                footnote-policy: line;
            }
        </style>
        <div>abc def ghi jkl<span>1</span></div>`)
	page1, page2 := pages[0], pages[1]

	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	linebox1, linebox2 := unpack2(div)
	assertText(t, unpack1(linebox1), "abc")
	assertText(t, unpack1(linebox2), "def")

	html, footnoteArea := unpack2(page2)
	body = unpack1(html)
	div = unpack1(body)
	linebox1, linebox2 = unpack2(div)
	assertText(t, unpack1(linebox1), "ghi")
	assertText(t, unpack1(linebox2), "jkl")
	assertText(t, unpack1(linebox2.Box().Children[1]), "1")

	footnoteMarker, footnoteTextbox := unpack2(unpack1(unpack1(footnoteArea)))
	assertText(t, unpack1(footnoteMarker), "1.")
	assertText(t, footnoteTextbox, "1")
}

func TestFootnotePolicyBlock(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 9px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
                footnote-policy: block;
            }
        </style>
        <div>abc</div><div>def ghi jkl<span>1</span></div>`)
	page1, page2 := pages[0], pages[1]

	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	linebox1 := unpack1(div)
	assertText(t, unpack1(linebox1), "abc")

	html, footnoteArea := unpack2(page2)
	body = unpack1(html)
	div = unpack1(body)
	linebox1, linebox2, linebox3 := unpack3(div)
	assertText(t, unpack1(linebox1), "def")
	assertText(t, unpack1(linebox2), "ghi")
	assertText(t, unpack1(linebox3), "jkl")
	assertText(t, unpack1(linebox3.Box().Children[1]), "1")

	footnoteMarker, footnoteTextbox := unpack2(unpack1(unpack1(footnoteArea)))
	assertText(t, unpack1(footnoteMarker), "1.")
	assertText(t, footnoteTextbox, "1")
}

func TestFootnoteRepagination(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div::after {
                content: counter(pages);
            }
            span {
                float: footnote;
            }
        </style>
        <div>ab<span>de</span></div>`)
	html, footnoteArea := unpack2(page)
	body := unpack1(html)
	div := unpack1(body)
	divTextbox, footnoteCall, divAfter := unpack3(unpack1(div))
	assertText(t, divTextbox, "ab")
	assertText(t, unpack1(footnoteCall), "1")
	tu.AssertEqual(t, divTextbox.Box().PositionY, Fl(0))
	assertText(t, unpack1(divAfter), "1")

	footnoteMarker, footnoteTextbox := unpack2(unpack1(unpack1(footnoteArea)))
	assertText(t, unpack1(footnoteMarker), "1.")
	assertText(t, footnoteTextbox, "de")
	tu.AssertEqual(t, footnoteArea.Box().PositionY, Fl(5))
}

func TestReportedFootnoteRepagination(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1700
	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 5px;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
            }
            a::after {
                content: target-counter(attr(href), page);
            }
        </style>
        <div><a href="#i">a</a> bb<span>de</span> <i id="i">fg</i></div>`)
	page1, page2 := pages[0], pages[1]
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	line1, line2 := unpack2(div)
	a := unpack1(line1)
	assertText(t, unpack1(a), "a")
	assertText(t, unpack1(a.Box().Children[1]), "2")
	b, footnoteCall, _ := unpack3(line2)
	assertText(t, b, "bb")
	assertText(t, unpack1(footnoteCall), "1")

	html, footnoteArea := unpack2(page2)
	body = unpack1(html)
	div = unpack1(body)
	line1 = unpack1(div)
	i := unpack1(line1)
	assertText(t, unpack1(i), "fg")

	footnoteMarker, footnoteTextbox := unpack2(unpack1(unpack1(footnoteArea)))
	assertText(t, unpack1(footnoteMarker), "1.")
	assertText(t, footnoteTextbox, "de")
	tu.AssertEqual(t, footnoteArea.Box().PositionY, Fl(3))
}

func TestFootnoteMaxHeight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
          @page {
              size: 12px 6px;

              @footnote {
                  margin-left: 1px;
                  max-height: 4px;
              }
          }
          div {
              font-family: weasyprint;
              font-size: 2px;
              line-height: 1;
          }
          div.footnote {
              float: footnote;
          }
      </style>
      <div>ab<div class="footnote">c</div><div class="footnote">d</div>
      <div class="footnote">e</div></div>
      <div>fg</div>`)
	page1, page2 := pages[0], pages[1]

	html1, footnote_area1 := unpack2(page1)
	body1 := unpack1(html1)
	div := unpack1(body1)
	divTextbox, footnoteCall1, footnoteCall2, space, footnoteCall3 := unpack5(unpack1(div))
	assertText(t, divTextbox, "ab")
	assertText(t, unpack1(footnoteCall1), "1")
	assertText(t, unpack1(footnoteCall2), "2")
	assertText(t, space, " ")
	assertText(t, unpack1(footnoteCall3), "3")
	footnote1, footnote2 := unpack2(footnote_area1)
	footnoteLine1 := unpack1(footnote1)
	footnoteMarker1, footnoteContent1 := unpack2(footnoteLine1)
	assertText(t, unpack1(footnoteMarker1), "1.")
	assertText(t, footnoteContent1, "c")
	footnoteLine2 := unpack1(footnote2)
	footnoteMarker2, footnoteContent2 := unpack2(footnoteLine2)
	assertText(t, unpack1(footnoteMarker2), "2.")
	assertText(t, footnoteContent2, "d")

	html2, footnoteArea2 := unpack2(page2)
	body2 := unpack1(html2)
	div2 := unpack1(body2)
	divTextbox2 := unpack1(div2.Box().Children[0])
	assertText(t, divTextbox2, "fg")
	footnoteLine3 := unpack1(unpack1(footnoteArea2))
	footnoteMarker3, footnoteContent3 := unpack2(footnoteLine3)
	assertText(t, unpack1(footnoteMarker3), "3.")
	assertText(t, footnoteContent3, "e")
}

func TestFootnoteTableAbortedRow(t *testing.T) {
	pages := renderPages(t, `
      <style>
        @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        @page {size: 10px 35px}
        body {font-family: weasyprint; font-size: 2px}
        tr {height: 10px}
        .footnote {float: footnote}
      </style>
      <table><tbody>
        <tr><td>abc</td></tr>
        <tr><td>abc</td></tr>
        <tr><td>abc</td></tr>
        <tr><td>def<div class="footnote">f</div></td></tr>
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

	html, footnote_area := unpack2(page2)
	body = unpack1(html)
	table_wrapper = unpack1(body)
	table = unpack1(table_wrapper)
	tbody = unpack1(table)
	tr := unpack1(tbody)
	td := unpack1(tr)
	line := unpack1(td)
	textbox, _ := unpack2(line)
	assertText(t, textbox, "def")
	footnote := unpack1(footnote_area)
	line = unpack1(footnote)
	_, textbox = unpack2(line)
	assertText(t, textbox, "f")
}

func TestFootnoteTableAbortedGroup(t *testing.T) {
	pages := renderPages(t, `
      <style>
        @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        @page {size: 10px 35px}
        body {font-family: weasyprint; font-size: 2px}
        tr {height: 10px}
        tbody {break-inside: avoid}
        .footnote {float: footnote}
      </style>
      <table>
        <tbody>
          <tr><td>abc</td></tr>
          <tr><td>abc</td></tr>
        </tbody>
        <tbody>
          <tr><td>def<div class="footnote">f</div></td></tr>
          <tr><td>ghi</td></tr>
        </tbody>
      </table>
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

	html, footnote_area := unpack2(page2)
	body = unpack1(html)
	table_wrapper = unpack1(body)
	table = unpack1(table_wrapper)
	tbody = unpack1(table)
	tr1, tr2 := unpack2(tbody)
	td := unpack1(tr1)
	line := unpack1(td)
	textbox, _ := unpack2(line)
	assertText(t, textbox, "def")
	td = unpack1(tr2)
	line = unpack1(td)
	textbox = unpack1(line)
	assertText(t, textbox, "ghi")
	footnote := unpack1(footnote_area)
	line = unpack1(footnote)
	_, textbox = unpack2(line)
	assertText(t, textbox, "f")
}
