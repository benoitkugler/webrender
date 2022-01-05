package layout

import (
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for footnotes layout.

func TestInlineFootnote(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
                background: white;
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
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	divTextbox, footnoteCall := unpack2(div.Box().Children[0])
	tu.AssertEqual(t, divTextbox.(*bo.TextBox).Text, "abc", "")
	tu.AssertEqual(t, footnoteCall.Box().Children[0].(*bo.TextBox).Text, "1", "")
	tu.AssertEqual(t, divTextbox.Box().PositionY, pr.Float(0), "")

	footnoteMarker, footnoteTextbox := unpack2(footnoteArea.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, footnoteMarker.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteTextbox.(*bo.TextBox).Text, "de", "")
	tu.AssertEqual(t, footnoteArea.Box().PositionY, pr.Float(5), "")
}

func TestBlockFootnote(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
         @font-face {src: url(weasyprint.otf); font-family: weasyprint}
         @page {
             size: 9px 7px;
             background: white;
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
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	divTextbox, footnoteCall := unpack2(div.Box().Children[0])
	tu.AssertEqual(t, divTextbox.(*bo.TextBox).Text, "abc", "")
	tu.AssertEqual(t, footnoteCall.Box().Children[0].(*bo.TextBox).Text, "1", "")
	tu.AssertEqual(t, divTextbox.Box().PositionY, pr.Float(0), "")
	footnoteMarker, footnoteTextbox := unpack2(footnoteArea.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, footnoteMarker.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteTextbox.(*bo.TextBox).Text, "de", "")
	tu.AssertEqual(t, footnoteArea.Box().PositionY, pr.Float(5), "")
}

func TestLongFootnote(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
                background: white;
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
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	divTextbox, footnoteCall := unpack2(div.Box().Children[0])
	tu.AssertEqual(t, divTextbox.(*bo.TextBox).Text, "abc", "")
	tu.AssertEqual(t, footnoteCall.Box().Children[0].(*bo.TextBox).Text, "1", "")
	tu.AssertEqual(t, divTextbox.Box().PositionY, pr.Float(0), "")
	footnoteLine1, footnoteLine2 := unpack2(footnoteArea.Box().Children[0])
	footnoteMarker, footnoteContent1 := unpack2(footnoteLine1)
	footnoteContent2 := footnoteLine2.Box().Children[0]
	tu.AssertEqual(t, footnoteMarker.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteContent1.(*bo.TextBox).Text, "de", "")
	tu.AssertEqual(t, footnoteArea.Box().PositionY, pr.Float(3), "")
	tu.AssertEqual(t, footnoteContent2.(*bo.TextBox).Text, "f", "")
	tu.AssertEqual(t, footnoteContent2.Box().PositionY, pr.Float(5), "")
}

func TestSeveralFootnote(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
                background: white;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                orphans: 1;
                widows: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>abcd e<span>fg</span> hijk l<span>mn</span></div>`)
	page1, page2 := pages[0], pages[1]

	html1, footnoteArea1 := unpack2(page1)
	body1 := html1.Box().Children[0]
	div1 := body1.Box().Children[0]
	div1Line1, _ := unpack2(div1)
	tu.AssertEqual(t, div1Line1.Box().Children[0].(*bo.TextBox).Text, "abcd", "")
	div1Line2Text, div1_footnoteCall := unpack2(div1.Box().Children[1])
	tu.AssertEqual(t, div1Line2Text.(*bo.TextBox).Text, "e", "")
	tu.AssertEqual(t, div1_footnoteCall.Box().Children[0].(*bo.TextBox).Text, "1", "")
	footnoteMarker1, footnoteTextbox1 := unpack2(footnoteArea1.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, footnoteMarker1.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteTextbox1.(*bo.TextBox).Text, "fg", "")

	html2, footnoteArea2 := unpack2(page2)
	body2 := html2.Box().Children[0]
	div2 := body2.Box().Children[0]
	div2Line1, _ := unpack2(div2)
	tu.AssertEqual(t, div2Line1.Box().Children[0].(*bo.TextBox).Text, "hijk", "")
	div2Line2Text, div2_footnoteCall := unpack2(div2.Box().Children[1])
	tu.AssertEqual(t, div2Line2Text.(*bo.TextBox).Text, "l", "")
	tu.AssertEqual(t, div2_footnoteCall.Box().Children[0].(*bo.TextBox).Text, "2", "")
	footnoteMarker2, footnoteTextbox2 := unpack2(footnoteArea2.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, footnoteMarker2.Box().Children[0].(*bo.TextBox).Text, "2.", "")
	tu.AssertEqual(t, footnoteTextbox2.(*bo.TextBox).Text, "mn", "")
}

func TestReportedFootnote_1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
                background: white;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                orphans: 1;
                widows: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>abc<span>f1</span> hij<span>f2</span></div>`)
	page1, page2 := pages[0], pages[1]

	html1, footnoteArea1 := unpack2(page1)
	body1 := html1.Box().Children[0]
	div1 := body1.Box().Children[0]
	div_line1, div_line2 := unpack2(div1)
	div_line1_text, div_footnoteCall1 := unpack2(div_line1)
	tu.AssertEqual(t, div_line1_text.(*bo.TextBox).Text, "abc", "")
	tu.AssertEqual(t, div_footnoteCall1.Box().Children[0].(*bo.TextBox).Text, "1", "")
	div_line2_text, div_footnoteCall2 := unpack2(div_line2)
	tu.AssertEqual(t, div_line2_text.(*bo.TextBox).Text, "hij", "")
	tu.AssertEqual(t, div_footnoteCall2.Box().Children[0].(*bo.TextBox).Text, "2", "")

	footnoteMarker1, footnoteTextbox1 := unpack2(footnoteArea1.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, footnoteMarker1.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteTextbox1.(*bo.TextBox).Text, "f1", "")

	html2, footnoteArea2 := unpack2(page2)
	tu.AssertEqual(t, len(html2.Box().Children), 0, "")
	footnoteMarker2, footnoteTextbox2 := unpack2(footnoteArea2.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, footnoteMarker2.Box().Children[0].(*bo.TextBox).Text, "2.", "")
	tu.AssertEqual(t, footnoteTextbox2.(*bo.TextBox).Text, "f2", "")
}

func TestReportedFootnote_2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
                background: white;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                orphans: 1;
                widows: 1;
            }
            span {
                float: footnote;
            }
        </style>
        <div>abc<span>f1</span> hij<span>f2</span> wow</div>`)
	page1, page2 := pages[0], pages[1]

	html1, footnoteArea1 := unpack2(page1)
	body1 := html1.Box().Children[0]
	div1 := body1.Box().Children[0]
	div_line1, div_line2 := unpack2(div1)
	div_line1_text, div_footnoteCall1 := unpack2(div_line1)
	tu.AssertEqual(t, div_line1_text.(*bo.TextBox).Text, "abc", "")
	tu.AssertEqual(t, div_footnoteCall1.Box().Children[0].(*bo.TextBox).Text, "1", "")
	div_line2_text, div_footnoteCall2 := unpack2(div_line2)
	tu.AssertEqual(t, div_line2_text.(*bo.TextBox).Text, "hij", "")
	tu.AssertEqual(t, div_footnoteCall2.Box().Children[0].(*bo.TextBox).Text, "2", "")
	footnoteMarker1, footnoteTextbox1 := unpack2(footnoteArea1.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, footnoteMarker1.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteTextbox1.(*bo.TextBox).Text, "f1", "")

	html2, footnoteArea2 := unpack2(page2)
	body2 := html2.Box().Children[0]
	div2 := body2.Box().Children[0]
	div2_line := div2.Box().Children[0]
	tu.AssertEqual(t, div2_line.Box().Children[0].(*bo.TextBox).Text, "wow", "")
	footnoteMarker2, footnoteTextbox2 := unpack2(footnoteArea2.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, footnoteMarker2.Box().Children[0].(*bo.TextBox).Text, "2.", "")
	tu.AssertEqual(t, footnoteTextbox2.(*bo.TextBox).Text, "f2", "")
}

func TestReportedFootnote_3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 10px;
                background: white;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                orphans: 1;
                widows: 1;
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
	body1 := html1.Box().Children[0]
	div1 := body1.Box().Children[0]
	line1, line2, line3 := unpack3(div1)
	tu.AssertEqual(t, line1.Box().Children[0].(*bo.TextBox).Text, "abc", "")
	tu.AssertEqual(t, line1.Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "1", "")
	tu.AssertEqual(t, line2.Box().Children[0].(*bo.TextBox).Text, "def", "")
	tu.AssertEqual(t, line2.Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "2", "")
	tu.AssertEqual(t, line3.Box().Children[0].(*bo.TextBox).Text, "ghi", "")
	tu.AssertEqual(t, line3.Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "3", "")
	footnote1 := footnoteArea1.Box().Children[0]
	tu.AssertEqual(t, footnote1.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnote1.Box().Children[0].Box().Children[1].(*bo.TextBox).Text, "1", "")

	_, footnoteArea2 := unpack2(page2)
	footnote2, footnote3 := unpack2(footnoteArea2)
	tu.AssertEqual(t, footnote2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "2.", "")
	tu.AssertEqual(t, footnote2.Box().Children[0].Box().Children[1].(*bo.TextBox).Text, "v", "")
	tu.AssertEqual(t, footnote2.Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "long", "")
	tu.AssertEqual(t, footnote2.Box().Children[2].Box().Children[0].(*bo.TextBox).Text, "2", "")
	tu.AssertEqual(t, footnote3.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "3.", "")
	tu.AssertEqual(t, footnote3.Box().Children[0].Box().Children[1].(*bo.TextBox).Text, "3", "")
}

func TestFootnoteDisplayInline(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 50px;
                background: white;
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
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	div_line1, div_line2 := unpack2(div)
	div_textbox1, footnoteCall1 := unpack2(div_line1)
	div_textbox2, footnoteCall2 := unpack2(div_line2)
	tu.AssertEqual(t, div_textbox1.(*bo.TextBox).Text, "abc", "")
	tu.AssertEqual(t, div_textbox2.(*bo.TextBox).Text, "fgh", "")
	tu.AssertEqual(t, footnoteCall1.Box().Children[0].(*bo.TextBox).Text, "1", "")
	tu.AssertEqual(t, footnoteCall2.Box().Children[0].(*bo.TextBox).Text, "2", "")
	line := footnoteArea.Box().Children[0]
	footnote_mark1, footnoteTextbox1 := unpack2(line.Box().Children[0])
	footnote_mark2, footnoteTextbox2 := unpack2(line.Box().Children[1])
	tu.AssertEqual(t, footnote_mark1.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteTextbox1.(*bo.TextBox).Text, "d", "")
	tu.AssertEqual(t, footnote_mark2.Box().Children[0].(*bo.TextBox).Text, "2.", "")
	tu.AssertEqual(t, footnoteTextbox2.(*bo.TextBox).Text, "i", "")
}

func TestFootnoteLongerThanSpaceLeft(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
                background: white;
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

	html1 := page1.Box().Children[0]
	body1 := html1.Box().Children[0]
	div := body1.Box().Children[0]
	divTextbox, footnoteCall := unpack2(div.Box().Children[0])
	tu.AssertEqual(t, divTextbox.(*bo.TextBox).Text, "abc", "")
	tu.AssertEqual(t, footnoteCall.Box().Children[0].(*bo.TextBox).Text, "1", "")

	html2, footnoteArea := unpack2(page2)
	tu.AssertEqual(t, len(html2.Box().Children), 0, "")
	footnoteLine1, footnoteLine2, footnoteLine3 := unpack3(footnoteArea.Box().Children[0])
	footnoteMarker, footnoteContent1 := unpack2(footnoteLine1)
	footnoteContent2 := footnoteLine2.Box().Children[0]
	footnoteContent3 := footnoteLine3.Box().Children[0]
	tu.AssertEqual(t, footnoteMarker.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteContent1.(*bo.TextBox).Text, "def", "")
	tu.AssertEqual(t, footnoteContent2.(*bo.TextBox).Text, "ghi", "")
	tu.AssertEqual(t, footnoteContent3.(*bo.TextBox).Text, "jkl", "")
}

func TestFootnoteLongerThanPage(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Nothing is defined for this use case in the specification. In WeasyPrint,
	// the content simply overflows.
	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
                background: white;
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

	html1 := page1.Box().Children[0]
	body1 := html1.Box().Children[0]
	div := body1.Box().Children[0]
	divTextbox, footnoteCall := unpack2(div.Box().Children[0])
	tu.AssertEqual(t, divTextbox.(*bo.TextBox).Text, "abc", "")
	tu.AssertEqual(t, footnoteCall.Box().Children[0].(*bo.TextBox).Text, "1", "")

	html2, footnoteArea2 := unpack2(page2)
	tu.AssertEqual(t, len(html2.Box().Children), 0, "")
	footnoteLine1, footnoteLine2, footnoteLine3, footnoteLine4 := unpack4(footnoteArea2.Box().Children[0])
	footnoteMarker1, footnoteContent1 := unpack2(footnoteLine1)
	footnoteContent2 := footnoteLine2.Box().Children[0]
	footnoteContent3 := footnoteLine3.Box().Children[0]
	footnoteContent4 := footnoteLine4.Box().Children[0]
	tu.AssertEqual(t, footnoteMarker1.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteContent1.(*bo.TextBox).Text, "def", "")
	tu.AssertEqual(t, footnoteContent2.(*bo.TextBox).Text, "ghi", "")
	tu.AssertEqual(t, footnoteContent3.(*bo.TextBox).Text, "jkl", "")
	tu.AssertEqual(t, footnoteContent4.(*bo.TextBox).Text, "mno", "")
}

func TestFootnotePolicyLine(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 9px;
                background: white;
            }
            div {
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            span {
                float: footnote;
                footnote-policy: line;
            }
        </style>
        <div>abc def ghi jkl<span>1</span></div>`)
	page1, page2 := pages[0], pages[1]

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	linebox1, linebox2 := unpack2(div)
	tu.AssertEqual(t, linebox1.Box().Children[0].(*bo.TextBox).Text, "abc", "")
	tu.AssertEqual(t, linebox2.Box().Children[0].(*bo.TextBox).Text, "def", "")

	html, footnoteArea := unpack2(page2)
	body = html.Box().Children[0]
	div = body.Box().Children[0]
	linebox1, linebox2 = unpack2(div)
	tu.AssertEqual(t, linebox1.Box().Children[0].(*bo.TextBox).Text, "ghi", "")
	tu.AssertEqual(t, linebox2.Box().Children[0].(*bo.TextBox).Text, "jkl", "")
	tu.AssertEqual(t, linebox2.Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "1", "")

	footnoteMarker, footnoteTextbox := unpack2(footnoteArea.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, footnoteMarker.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteTextbox.(*bo.TextBox).Text, "1", "")
}

func TestFootnotePolicyBlock(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 9px;
                background: white;
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

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	linebox1 := div.Box().Children[0]
	tu.AssertEqual(t, linebox1.Box().Children[0].(*bo.TextBox).Text, "abc", "")

	html, footnoteArea := unpack2(page2)
	body = html.Box().Children[0]
	div = body.Box().Children[0]
	linebox1, linebox2, linebox3 := unpack3(div)
	tu.AssertEqual(t, linebox1.Box().Children[0].(*bo.TextBox).Text, "def", "")
	tu.AssertEqual(t, linebox2.Box().Children[0].(*bo.TextBox).Text, "ghi", "")
	tu.AssertEqual(t, linebox3.Box().Children[0].(*bo.TextBox).Text, "jkl", "")
	tu.AssertEqual(t, linebox3.Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "1", "")

	footnoteMarker, footnoteTextbox := unpack2(footnoteArea.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, footnoteMarker.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteTextbox.(*bo.TextBox).Text, "1", "")
}

func TestFootnoteRepagination(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
            @font-face {src: url(weasyprint.otf); font-family: weasyprint}
            @page {
                size: 9px 7px;
                background: white;
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
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	divTextbox, footnoteCall, divAfter := unpack3(div.Box().Children[0])
	tu.AssertEqual(t, divTextbox.(*bo.TextBox).Text, "ab", "")
	tu.AssertEqual(t, footnoteCall.Box().Children[0].(*bo.TextBox).Text, "1", "footnoteCall")
	tu.AssertEqual(t, divTextbox.Box().PositionY, pr.Float(0), "")
	tu.AssertEqual(t, divAfter.Box().Children[0].(*bo.TextBox).Text, "1", "divAfter")

	footnoteMarker, footnoteTextbox := unpack2(footnoteArea.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, footnoteMarker.Box().Children[0].(*bo.TextBox).Text, "1.", "")
	tu.AssertEqual(t, footnoteTextbox.(*bo.TextBox).Text, "de", "")
	tu.AssertEqual(t, footnoteArea.Box().PositionY, pr.Float(5), "")
}
