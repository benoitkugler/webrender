package layout

import (
	"fmt"
	"strings"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for inline layout.

var sansFonts = strings.Join(pr.Strings{"DejaVu Sans", "sans"}, " ")

func TestEmptyLinebox(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, "<p> </p>")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	tu.AssertEqual(t, len(paragraph.Box().Children), 0, "len")
	tu.AssertEqual(t, paragraph.Box().Height, pr.Float(0), "paragraph")
}

// @pytest.mark.xfail
// func TestEmptyLineboxRemovedSpace(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Whitespace removed at the beginning of the line => empty line => no line
//     page := renderOnePage(t, `
//       <style>
//         p { width: 1px }
//       </style>
//       <p><br>  </p>
//     `)
//     page := renderOnePage(t, "<p> </p>")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     // TODO: The second line should be removed
//     tu.AssertEqual(t, len(paragraph.Box().Children) , 1, "len")

func TestBreakingLinebox(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, fmt.Sprintf(`
      <style>
      p { font-size: 13px;
          width: 300px;
          font-family: %s;
          background-color: #393939;
          color: #FFFFFF;
          line-height: 1;
          text-decoration: underline overline line-through;}
      </style>
      <p><em>Lorem<strong> Ipsum <span>is very</span>simply</strong><em>
      dummy</em>text of the printing and. naaaa </em> naaaa naaaa naaaa
      naaaa naaaa naaaa naaaa naaaa</p>
    `, sansFonts))
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	tu.AssertEqual(t, len(paragraph.Box().Children), 3, "len")

	lines := paragraph.Box().Children
	for _, line := range lines {
		tu.AssertEqual(t, line.Box().Style.GetFontSize(), pr.FToV(13), "line")
		tu.AssertEqual(t, line.Box().ElementTag(), "p", "line")
		for _, child := range line.Box().Children {
			tu.AssertEqual(t, child.Box().ElementTag() == "em" || child.Box().ElementTag() == "p", true, "child")
			tu.AssertEqual(t, child.Box().Style.GetFontSize(), pr.FToV(13), "child")
			if bo.ParentBoxT.IsInstance(child) {
				for _, childChild := range child.Box().Children {
					tu.AssertEqual(t, childChild.Box().ElementTag() == "em" || childChild.Box().ElementTag() == "strong" || childChild.Box().ElementTag() == "span", true, "childChild")
					tu.AssertEqual(t, childChild.Box().Style.GetFontSize(), pr.FToV(13), "childChild")
				}
			}
		}
	}
}

func TestPositionXLtr(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        span {
          padding: 0 10px 0 15px;
          margin: 0 2px 0 3px;
          border: 1px solid;
         }
      </style>
      <body><span>a<br>b<br>c</span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line1, line2, line3 := unpack3(body)
	span1 := line1.Box().Children[0]
	tu.AssertEqual(t, span1.Box().PositionX, pr.Float(0), "span1")
	text1, _ := unpack2(span1)
	tu.AssertEqual(t, text1.Box().PositionX, pr.Float(15+3+1), "text1")
	span2 := line2.Box().Children[0]
	tu.AssertEqual(t, span2.Box().PositionX, pr.Float(0), "span2")
	text2, _ := unpack2(span2)
	tu.AssertEqual(t, text2.Box().PositionX, pr.Float(0), "text2")
	span3 := line3.Box().Children[0]
	tu.AssertEqual(t, span3.Box().PositionX, pr.Float(0), "span3")
	text3 := span3.Box().Children[0]
	tu.AssertEqual(t, text3.Box().PositionX, pr.Float(0), "text3")
}

func TestPositionXRtl(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        body {
          direction: rtl;
          width: 100px;
        }
        span {
          padding: 0 10px 0 15px;
          margin: 0 2px 0 3px;
          border: 1px solid;
         }
      </style>
      <body><span>a<br>b<br>c</span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line1, line2, line3 := unpack3(body)
	span1 := line1.Box().Children[0]
	text1, _ := unpack2(span1)
	tu.AssertEqual(t, span1.Box().PositionX, 100-text1.Box().Width.V()-(10+2+1), "span1")
	tu.AssertEqual(t, text1.Box().PositionX, 100-text1.Box().Width.V()-(10+2+1), "text1")
	span2 := line2.Box().Children[0]
	text2, _ := unpack2(span2)
	tu.AssertEqual(t, span2.Box().PositionX, 100-text2.Box().Width.V(), "span2")
	tu.AssertEqual(t, text2.Box().PositionX, 100-text2.Box().Width.V(), "text2")
	span3 := line3.Box().Children[0]
	text3 := span3.Box().Children[0]
	tu.AssertEqual(t, span3.Box().PositionX, 100-text3.Box().Width.V()-(15+3+1), "span3")
	tu.AssertEqual(t, text3.Box().PositionX, 100-text3.Box().Width.V(), "text3")
}

func TestBreakingLineboxRegression1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// See http://unicode.org/reports/tr14/
	page := renderOnePage(t, "<pre>a\nb\rc\r\nd\u2029e</pre>")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	pre := body.Box().Children[0]
	lines := pre.Box().Children
	var texts []string
	for _, line := range lines {
		textBox := line.Box().Children[0]
		texts = append(texts, textBox.(*bo.TextBox).Text)
	}
	tu.AssertEqual(t, texts, []string{"a", "b", "c", "d", "e"}, "texts")
}

func TestBreakingLineboxRegression2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	htmlSample := `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
      </style>
      <p style="width: %d.5em; font-family: weasyprint">ab
      <span style="padding-right: 1em; margin-right: 1em">c def</span>g
      hi</p>`
	for i := 0; i < 16; i++ {
		page := renderOnePage(t, fmt.Sprintf(htmlSample, i))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		p := body.Box().Children[0]

		if i <= 3 {
			line1, line2, line3, line4 := unpack4(p)

			textbox1 := line1.Box().Children[0]
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "ab", "textbox1")

			span1 := line2.Box().Children[0]
			textbox1 = span1.Box().Children[0]
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "c", "textbox1")

			span1, textbox2 := unpack2(line3)
			textbox1 = span1.Box().Children[0]
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "def", "textbox1")
			tu.AssertEqual(t, textbox2.(*bo.TextBox).Text, "g", "textbox2")

			textbox1 = line4.Box().Children[0]
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "hi", "textbox1")
		} else if i <= 8 {
			line1, line2, line3 := unpack3(p)
			textbox1, span1 := unpack2(line1)
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "ab ", "textbox1")
			textbox2 := span1.Box().Children[0]
			tu.AssertEqual(t, textbox2.(*bo.TextBox).Text, "c", "textbox2")

			span1, textbox2 = unpack2(line2)
			textbox1 = span1.Box().Children[0]
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "def", "textbox1")
			tu.AssertEqual(t, textbox2.(*bo.TextBox).Text, "g", "textbox2")

			textbox1 = line3.Box().Children[0]
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "hi", "textbox1")
		} else if i <= 10 {
			line1, line2 := unpack2(p)

			textbox1, span1 := unpack2(line1)
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "ab ", "textbox1")
			textbox2 := span1.Box().Children[0]
			tu.AssertEqual(t, textbox2.(*bo.TextBox).Text, "c", "textbox2")

			span1, textbox2 = unpack2(line2)
			textbox1 = span1.Box().Children[0]
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "def", "textbox1")
			tu.AssertEqual(t, textbox2.(*bo.TextBox).Text, "g hi", "textbox2")
		} else if i <= 13 {
			line1, line2 := unpack2(p)

			textbox1, span1, textbox3 := unpack3(line1)
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "ab ", "textbox1")
			textbox2 := span1.Box().Children[0]
			tu.AssertEqual(t, textbox2.(*bo.TextBox).Text, "c def", "textbox2")
			tu.AssertEqual(t, textbox3.(*bo.TextBox).Text, "g", "textbox3")

			textbox1 = line2.Box().Children[0]
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "hi", "textbox1")
		} else {
			line1 := p.Box().Children[0]

			textbox1, span1, textbox3 := unpack3(line1)
			tu.AssertEqual(t, textbox1.(*bo.TextBox).Text, "ab ", "textbox1")
			textbox2 := span1.Box().Children[0]
			tu.AssertEqual(t, textbox2.(*bo.TextBox).Text, "c def", "textbox2")
			tu.AssertEqual(t, textbox3.(*bo.TextBox).Text, "g hi", "textbox3")
		}
	}
}

func TestBreakingLineboxRegression3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test #1 for https://github.com/Kozea/WeasyPrint/issues/560
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <div style="width: 5.5em; font-family: weasyprint">
        aaaa aaaa a [<span>aaa</span>]`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	line1, line2, line3, line4 := unpack4(div)
	tu.AssertEqual(t, line1.Box().Children[0].(*bo.TextBox).Text, line2.Box().Children[0].(*bo.TextBox).Text, "line1")
	tu.AssertEqual(t, line2.Box().Children[0].(*bo.TextBox).Text, "aaaa", "line1")
	tu.AssertEqual(t, line3.Box().Children[0].(*bo.TextBox).Text, "a", "line3")
	text1, span, text2 := unpack3(line4)
	tu.AssertEqual(t, text1.(*bo.TextBox).Text, "[", "text1")
	tu.AssertEqual(t, text2.(*bo.TextBox).Text, "]", "text2")
	tu.AssertEqual(t, span.Box().Children[0].(*bo.TextBox).Text, "aaa", "span")
}

func TestBreakingLineboxRegression4(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test #2 for https://github.com/Kozea/WeasyPrint/issues/560
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <div style="width: 5.5em; font-family: weasyprint">
        aaaa a <span>b c</span>d`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	line1, line2, line3 := unpack3(div)
	tu.AssertEqual(t, line1.Box().Children[0].(*bo.TextBox).Text, "aaaa", "line1")
	tu.AssertEqual(t, line2.Box().Children[0].(*bo.TextBox).Text, "a ", "line2")
	tu.AssertEqual(t, line2.Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "b", "line2")
	tu.AssertEqual(t, line3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "c", "line3")
	tu.AssertEqual(t, line3.Box().Children[1].(*bo.TextBox).Text, "d", "line3")
}

func TestBreakingLineboxRegression5(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/580
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <div style="width: 5.5em; font-family: weasyprint">
        <span>aaaa aaaa a a a</span><span>bc</span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	line1, line2, line3, line4 := unpack4(div)
	tu.AssertEqual(t, line1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "aaaa", "line1")
	tu.AssertEqual(t, line2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "aaaa", "line2")
	tu.AssertEqual(t, line3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "a a", "line3")
	tu.AssertEqual(t, line4.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "a", "line4")
	tu.AssertEqual(t, line4.Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "bc", "line4")
}

func TestBreakingLineboxRegression6(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/586
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <div style="width: 5.5em; font-family: weasyprint">
        a a <span style="white-space: nowrap">/ccc</span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	line1, line2 := unpack2(div)
	tu.AssertEqual(t, line1.Box().Children[0].(*bo.TextBox).Text, "a a", "line1")
	tu.AssertEqual(t, line2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "/ccc", "line2")
}

func TestBreakingLineboxRegression7(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/660
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <div style="width: 3.5em; font-family: weasyprint">
        <span><span>abc d e</span></span><span>f`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	line1, line2, line3 := unpack3(div)
	tu.AssertEqual(t, line1.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "abc", "line1")
	tu.AssertEqual(t, line2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "d", "line2")
	tu.AssertEqual(t, line3.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "e", "line3")
	tu.AssertEqual(t, line3.Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "f", "line3")
}

func TestBreakingLineboxRegression8(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/783
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <p style="font-family: weasyprint"><span>
        aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
        bbbbbbbbbbb
        <b>cccc</b></span>ddd</p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	p := body.Box().Children[0]
	line1, line2 := unpack2(p)
	tu.AssertEqual(t, line1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa bbbbbbbbbbb", "line1")
	tu.AssertEqual(t, line2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "cccc", "line2")
	tu.AssertEqual(t, line2.Box().Children[1].(*bo.TextBox).Text, "ddd", "line2")
}

// @pytest.mark.xfail

// func TestBreakingLineboxRegression9(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Regression test for https://github.com/Kozea/WeasyPrint/issues/783
//     // TODO: inlines.canBreakInside return false for span but we can break
//     // before the <b> tag. canBreakInside should be fixed.
//     page := renderOnePage(t,
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<p style="font-family: weasyprint"><span>\n"
//         "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbb\n"
//         "<b>cccc</b></span>ddd</p>")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     p := body.Box().Children[0]
//     line1, line2 = p.Box().Children
//     tu.AssertEqual(t, line1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text , (, "line1")
//         "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbb")
//     tu.AssertEqual(t, line2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text , "cccc", "line2")
//     tu.AssertEqual(t, line2.Box().Children[1].(*bo.TextBox).Text , "ddd", "line2")
// }

func TestBreakingLineboxRegression10(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/923
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <p style="width:195px; font-family: weasyprint">
          <span>
            <span>xxxxxx YYY yyyyyy yyy</span>
            ZZZZZZ zzzzz
          </span> )x 
        </p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	p := body.Box().Children[0]
	line1, line2, line3, line4 := unpack4(p)
	tu.AssertEqual(t, line1.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "xxxxxx YYY", "line1")
	tu.AssertEqual(t, line2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "yyyyyy yyy", "line2")
	tu.AssertEqual(t, line3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "ZZZZZZ zzzzz", "line3")
	tu.AssertEqual(t, line4.Box().Children[0].(*bo.TextBox).Text, ")x", "line4")
}

func TestBreakingLineboxRegression11(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/953
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <p style="width:10em; font-family: weasyprint">
          line 1<br><span>123 567 90</span>x
        </p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	p := body.Box().Children[0]
	line1, line2, line3 := unpack3(p)
	tu.AssertEqual(t, line1.Box().Children[0].(*bo.TextBox).Text, "line 1", "line1")
	tu.AssertEqual(t, line2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "123 567", "line2")
	tu.AssertEqual(t, line3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "90", "line3")
	tu.AssertEqual(t, line3.Box().Children[1].(*bo.TextBox).Text, "x", "line3")
}

func TestBreakingLineboxRegression12(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/953
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <p style="width:10em; font-family: weasyprint">
          <br><span>123 567 90</span>x
        </p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	p := body.Box().Children[0]
	_, line2, line3 := unpack3(p)
	tu.AssertEqual(t, line2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "123 567", "line2")
	tu.AssertEqual(t, line3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "90", "line3")
	tu.AssertEqual(t, line3.Box().Children[1].(*bo.TextBox).Text, "x", "line3")
}

func TestBreakingLineboxRegression13(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/953
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <p style="width:10em; font-family: weasyprint">
          123 567 90 <span>123 567 90</span>x
        </p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	p := body.Box().Children[0]
	line1, line2, line3 := unpack3(p)
	tu.AssertEqual(t, line1.Box().Children[0].(*bo.TextBox).Text, "123 567 90", "line1")
	tu.AssertEqual(t, line2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "123 567", "line2")
	tu.AssertEqual(t, line3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "90", "line3")
	tu.AssertEqual(t, line3.Box().Children[1].(*bo.TextBox).Text, "x", "line3")
}

func TestLineboxText(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, fmt.Sprintf(`
      <style>
        p { width: 165px; font-family:%s;}
      </style>
      <p><em>Lorem Ipsum</em>is very <strong>coool</strong></p>
    `, sansFonts))
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	lines := paragraph.Box().Children
	tu.AssertEqual(t, len(lines), 2, "len")

	var chunks []string
	for _, line := range lines {
		s := ""
		for _, box := range bo.Descendants(line) {
			if box, ok := box.(*bo.TextBox); ok {
				s += box.Text
			}
		}
		chunks = append(chunks, s)
	}
	text := strings.Join(chunks, " ")
	tu.AssertEqual(t, text, "Lorem Ipsumis very coool", "text")
}

func TestLineboxPositions(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range [][2]int{{165, 2}, {1, 5}, {0, 5}} {
		width, expectedLines := data[0], data[1]

		page := renderOnePage(t, fmt.Sprintf(`
		<style>
		  p { width:%dpx; font-family:%s;
			  line-height: 20px }
		</style>
		<p>this is test for <strong>Weasyprint</strong></p>`, width, sansFonts))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		paragraph := body.Box().Children[0]
		lines := paragraph.Box().Children
		tu.AssertEqual(t, len(lines), expectedLines, "len")

		refPositionY := lines[0].Box().PositionY
		refPositionX := lines[0].Box().PositionX
		for _, line := range lines {
			tu.AssertEqual(t, refPositionY, line.Box().PositionY, "refPositionY")
			tu.AssertEqual(t, refPositionX, line.Box().PositionX, "refPositionX")
			for _, box := range line.Box().Children {
				tu.AssertEqual(t, refPositionX, box.Box().PositionX, "refPositionX")
				refPositionX += box.Box().Width.V()
				tu.AssertEqual(t, refPositionY, box.Box().PositionY, "refPositionY")
			}
			tu.AssertEqual(t, refPositionX-line.Box().PositionX <= line.Box().Width.V(), true, "refPositionX")
			refPositionX = line.Box().PositionX
			refPositionY += line.Box().Height.V()
		}
	}
}

func TestForcedLineBreaksPre(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// These lines should be small enough to fit on the default A4 page
	// with the default 12pt font-size.
	page := renderOnePage(t, `
      <style> pre { line-height: 42px }</style>
      <pre>Lorem ipsum dolor sit amet,
          consectetur adipiscing elit.


          Sed sollicitudin nibh

          et turpis molestie tristique.</pre>
	`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	pre := body.Box().Children[0]
	tu.AssertEqual(t, pre.Box().ElementTag(), "pre", "pre")
	lines := pre.Box().Children
	tu.AssertEqual(t, len(lines), 7, "len")
	for _, line := range lines {
		if !bo.LineBoxT.IsInstance(line) {
			t.Fatal()
		}
		tu.AssertEqual(t, line.Box().Height, pr.Float(42), "line")
	}
}

func TestForcedLineBreaksParagraph(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style> p { line-height: 42px }</style>
      <p>Lorem ipsum dolor sit amet,<br>
        consectetur adipiscing elit.<br><br><br>
        Sed sollicitudin nibh<br>
        <br>
 
        et turpis molestie tristique.</p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	tu.AssertEqual(t, paragraph.Box().ElementTag(), "p", "paragraph")
	lines := paragraph.Box().Children
	tu.AssertEqual(t, len(lines), 7, "len")
	for _, line := range lines {
		if !bo.LineBoxT.IsInstance(line) {
			t.Fatal()
		}
		tu.AssertEqual(t, line.Box().Height, pr.Float(42), "line")
	}
}

func TestInlineboxSplitting(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// The text is strange to test some corner cases
	// See https://github.com/Kozea/WeasyPrint/issues/389
	for _, width := range []int{10000, 100, 10, 0} {
		page := renderOnePage(t, fmt.Sprintf(`
          <style>p { font-family:%s; width: %dpx; }</style>
          <p><strong>WeasyPrint is a frée softwäre ./ visual rendèring enginè
                     for HTML !!! && CSS.</strong></p>
        `, sansFonts, width))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		paragraph := body.Box().Children[0]
		lines := paragraph.Box().Children
		if width == 10000 {
			tu.AssertEqual(t, len(lines), 1, "len")
		} else {
			tu.AssertEqual(t, len(lines) > 1, true, "len")
		}
		var textParts []string
		for _, line := range lines {
			strong := line.Box().Children[0]
			text := strong.Box().Children[0]
			textParts = append(textParts, text.(*bo.TextBox).Text)
		}
		tu.AssertEqual(t, strings.Join(textParts, " "),
			"WeasyPrint is a frée softwäre ./ visual "+
				"rendèring enginè for HTML !!! && CSS.", "")
	}
}

func TestWhitespaceProcessing(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, source := range []string{"a", "  a  ", " \n  \ta", " a\t "} {
		page := renderOnePage(t, fmt.Sprintf("<p><em>%s</em></p>", source))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		p := body.Box().Children[0]
		line := p.Box().Children[0]
		em := line.Box().Children[0]
		text := em.Box().Children[0]
		tu.AssertEqual(t, text.(*bo.TextBox).Text, "a", fmt.Sprintf("source was %s", source))

		page = renderOnePage(t, fmt.Sprintf(
			`<p style="white-space: pre-line">
			
			<em>%s</em></pre>`, strings.ReplaceAll(source, "\n", " ")))
		html = page.Box().Children[0]
		body = html.Box().Children[0]
		p = body.Box().Children[0]
		_, _, line3 := unpack3(p)
		em = line3.Box().Children[0]
		text = em.Box().Children[0]
		tu.AssertEqual(t, text.(*bo.TextBox).Text, "a", fmt.Sprintf("source was %s", source))
	}
}

func TestInlineReplacedAutoMargins(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 200px }
        img { display: inline; margin: auto; width: 50px }
      </style>
      <body><img src="pattern.png" />`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img := line.Box().Children[0]
	tu.AssertEqual(t, img.Box().MarginTop, pr.Float(0), "img")
	tu.AssertEqual(t, img.Box().MarginRight, pr.Float(0), "img")
	tu.AssertEqual(t, img.Box().MarginBottom, pr.Float(0), "img")
	tu.AssertEqual(t, img.Box().MarginLeft, pr.Float(0), "img")
}

func TestEmptyInlineAutoMargins(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 200px }
        span { margin: auto }
      </style>
      <body><span></span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	block := body.Box().Children[0]
	span := block.Box().Children[0]
	tu.AssertEqual(t, span.Box().MarginTop != pr.Float(0), true, "span")
	tu.AssertEqual(t, span.Box().MarginRight, pr.Float(0), "span")
	tu.AssertEqual(t, span.Box().MarginBottom != pr.Float(0), true, "span")
	tu.AssertEqual(t, span.Box().MarginLeft, pr.Float(0), "span")
}

func TestFontStretch(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, fmt.Sprintf(`
      <style>
        p { float: left; font-family: %s }
      </style>
      <p>Hello, world!</p>
      <p style="font-stretch: condensed">Hello, world!</p>
    `, sansFonts))
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	p1, p2 := unpack2(body)
	normal := p1.Box().Width.V()
	condensed := p2.Box().Width.V()
	tu.AssertEqual(t, condensed < normal, true, "condensed")
}

func TestLineCount(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	for _, data := range []struct {
		source    string
		lineCount int
	}{
		{`<body>hyphénation`, 1},                               // Default: no hyphenation
		{`<body lang=fr>hyphénation`, 1},                       // lang only: no hyphenation
		{`<body style="hyphens: auto">hyphénation`, 1},         // hyphens only: no hyph.
		{`<body style="hyphens: auto" lang=fr>hyphénation`, 4}, // both: hyph.
		{`<body>hyp&shy;hénation`, 2},                          // Hyphenation with soft hyphens
		{`<body style="hyphens: none">hyp&shy;hénation`, 1},    // … unless disabled
	} {
		page := renderOnePage(t, `
        <html style="width: 5em; font-family: weasyprint">
        <style>@font-face {
          src:url(weasyprint.otf); font-family :weasyprint
        }</style>`+data.source)
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		lines := body.Box().Children
		tu.AssertEqual(t, len(lines), data.lineCount, "len")
	}
}

func TestVerticalAlign1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	//            +-------+      <- positionY = 0
	//      +-----+       |
	// 40px |     |       | 60px
	//      |     |       |
	//      +-----+-------+      <- baseline
	page := renderOnePage(t, `
      <span>
        <img src="pattern.png" style="width: 40px"
        ><img src="pattern.png" style="width: 60px"
      ></span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span := line.Box().Children[0]
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, pr.Float(40), "img1")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(60), "img2")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(20), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(0), "img2")
	// 60px + the descent of the font below the baseline
	tu.AssertEqual(t, 60 < line.Box().Height.V(), true, "60")
	tu.AssertEqual(t, line.Box().Height.V() < 70, true, "60")
	tu.AssertEqual(t, body.Box().Height, line.Box().Height, "body")
}

func TestVerticalAlign2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	//            +-------+      <- positionY = 0
	//       35px |       |
	//      +-----+       | 60px
	// 40px |     |       |
	//      |     +-------+      <- baseline
	//      +-----+  15px
	page := renderOnePage(t, `
      <span>
        <img src="pattern.png" style="width: 40px; vertical-align: -15px"
        ><img src="pattern.png" style="width: 60px"></span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span := line.Box().Children[0]
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, pr.Float(40), "img1")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(60), "img2")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(35), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(0), "img2")
	tu.AssertEqual(t, line.Box().Height, pr.Float(75), "line")
	tu.AssertEqual(t, body.Box().Height, line.Box().Height, "body")
}

func TestVerticalAlign3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Same as previously, but with percentages
	page := renderOnePage(t, `
      <span style="line-height: 10px">
        <img src="pattern.png" style="width: 40px; vertical-align: -150%"
        ><img src="pattern.png" style="width: 60px"></span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span := line.Box().Children[0]
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, pr.Float(40), "img1")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(60), "img2")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(35), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(0), "img2")
	tu.AssertEqual(t, line.Box().Height, pr.Float(75), "line")
	tu.AssertEqual(t, body.Box().Height, line.Box().Height, "body")
}

func TestVerticalAlign4(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Same again, but have the vertical-align on an inline box.
	page := renderOnePage(t, `
      <span style="line-height: 10px">
        <span style="line-height: 10px; vertical-align: -15px">
          <img src="pattern.png" style="width: 40px"></span>
        <img src="pattern.png" style="width: 60px"></span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span1 := line.Box().Children[0]
	span2, _, img2 := unpack3(span1)
	img1 := span2.Box().Children[0]
	tu.AssertEqual(t, img1.Box().Height, pr.Float(40), "img1")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(60), "img2")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(35), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(0), "img2")
	tu.AssertEqual(t, line.Box().Height, pr.Float(75), "line")
	tu.AssertEqual(t, body.Box().Height, line.Box().Height, "body")
}

func TestVerticalAlign5(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Same as previously, but with percentages
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <span style="line-height: 12px; font-size: 12px;
			font-family: weasyprint"><img src="pattern.png" 
			style="width: 40px; vertical-align: middle"><img src="pattern.png" 
			style="width: 60px"></span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span := line.Box().Children[0]
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, pr.Float(40), "img1")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(60), "img2")
	// middle of the image (positionY + 20) is at half the ex-height above
	// the baseline of the parent. The ex-height of weasyprint.otf is 0.8em
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(35.201202), "img1") // 60 - 0.5 * 0.8 * font-size - 40/2
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(0), "img2")
	tu.AssertEqual(t, line.Box().Height, pr.Float(75.2012), "line")
	tu.AssertEqual(t, body.Box().Height, line.Box().Height, "body")
}

func TestVerticalAlign6(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// sup and sub currently mean +/- 0.5 em
	// With the initial 16px font-size, that’s 8px.
	page := renderOnePage(t, `
      <span style="line-height: 10px">
        <img src="pattern.png" style="width: 60px"
        ><img src="pattern.png" style="width: 40px; vertical-align: super"
        ><img src="pattern.png" style="width: 40px; vertical-align: sub"
      ></span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span := line.Box().Children[0]
	img1, img2, img3 := unpack3(span)
	tu.AssertEqual(t, img1.Box().Height, pr.Float(60), "img1")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(40), "img2")
	tu.AssertEqual(t, img3.Box().Height, pr.Float(40), "img3")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(0), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(12), "img2") // 20 - 16 * 0.5)
	tu.AssertEqual(t, img3.Box().PositionY, pr.Float(28), "img3") // 20 + 16 * 0.5)
	tu.AssertEqual(t, line.Box().Height, pr.Float(68), "line")
	tu.AssertEqual(t, body.Box().Height, line.Box().Height, "body")
}

func TestVerticalAlign7(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <body style="line-height: 10px">
        <span>
          <img src="pattern.png" style="vertical-align: text-top"
          ><img src="pattern.png" style="vertical-align: text-bottom"
        ></span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span := line.Box().Children[0]
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, pr.Float(4), "img1")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(4), "img2")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(0), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(12), "img2") // 16 - 4
	tu.AssertEqual(t, line.Box().Height, pr.Float(16), "line")
	tu.AssertEqual(t, body.Box().Height, line.Box().Height, "body")
}

func TestVerticalAlign8(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// This case used to cause an exception:
	// The second span has no children but should count for line heights
	// since it has padding.
	page := renderOnePage(t, `<span style="line-height: 1.5">
      <span style="padding: 1px"></span></span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span1 := line.Box().Children[0]
	span2 := span1.Box().Children[0]
	tu.AssertEqual(t, span1.Box().Height, pr.Float(16), "span1")
	tu.AssertEqual(t, span2.Box().Height, pr.Float(16), "span2")
	// The line’s strut does not has "line-height: normal" but the result should
	// be smaller than 1.5.
	tu.AssertEqual(t, span1.Box().MarginHeight(), pr.Float(24), "span1")
	tu.AssertEqual(t, span2.Box().MarginHeight(), pr.Float(24), "span2")
	tu.AssertEqual(t, line.Box().Height, pr.Float(24), "line")
}

func TestVerticalAlign9(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <span>
        <img src="pattern.png" style="width: 40px; vertical-align: -15px"
        ><img src="pattern.png" style="width: 60px"
      ></span><div style="display: inline-block; vertical-align: 3px">
        <div>
          <div style="height: 100px">foo</div>
          <div>
            <img src="pattern.png" style="
                 width: 40px; vertical-align: -15px"
            ><img src="pattern.png" style="width: 60px"
          ></div>
        </div>
      </div>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span, div1 := unpack2(line)
	tu.AssertEqual(t, line.Box().Height, pr.Float(178), "line")
	tu.AssertEqual(t, body.Box().Height, line.Box().Height, "body")

	// Same as earlier
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, pr.Float(40), "img1")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(60), "img2")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(138), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(103), "img2")

	div2 := div1.Box().Children[0]
	div3, div4 := unpack2(div2)
	divLine := div4.Box().Children[0]
	divImg1, divImg2 := unpack2(divLine)
	tu.AssertEqual(t, div1.Box().PositionY, pr.Float(0), "div1")
	tu.AssertEqual(t, div1.Box().Height, pr.Float(175), "div1")
	tu.AssertEqual(t, div3.Box().Height, pr.Float(100), "div3")
	tu.AssertEqual(t, divLine.Box().Height, pr.Float(75), "divLine")
	tu.AssertEqual(t, divImg1.Box().Height, pr.Float(40), "divImg1")
	tu.AssertEqual(t, divImg2.Box().Height, pr.Float(60), "divImg2")
	tu.AssertEqual(t, divImg1.Box().PositionY, pr.Float(135), "divImg1")
	tu.AssertEqual(t, divImg2.Box().PositionY, pr.Float(100), "divImg2")
}

func TestVerticalAlign10(t *testing.T) {
	// capt := tu.CaptureLogs()
	// defer capt.AssertNoLogs(t)

	// The first two images bring the top of the line box 30px above
	// the baseline and 10px below.
	// Each of the inner span
	page := renderOnePage(t, `
      <span style="font-size: 0">
        <img src="pattern.png" style="vertical-align: 26px">
        <img src="pattern.png" style="vertical-align: -10px">
        <span style="vertical-align: top">
          <img src="pattern.png" style="vertical-align: -10px">
          <span style="vertical-align: -10px">
            <img src="pattern.png" style="vertical-align: bottom">
          </span>
        </span>
        <span style="vertical-align: bottom">
          <img src="pattern.png" style="vertical-align: 6px">
        </span>
      </span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span1 := line.Box().Children[0]
	img1, img2, span2, span4 := unpack4(span1)
	img3, span3 := unpack2(span2)
	img4 := span3.Box().Children[0]
	img5 := span4.Box().Children[0]
	tu.AssertEqual(t, body.Box().Height, line.Box().Height, "body")
	tu.AssertEqual(t, line.Box().Height, pr.Float(40), "line")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(0), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(36), "img2")
	tu.AssertEqual(t, img3.Box().PositionY, pr.Float(6), "img3")
	tu.AssertEqual(t, img4.Box().PositionY, pr.Float(36), "img4")
	tu.AssertEqual(t, img5.Box().PositionY, pr.Float(30), "img5")
}

func TestVerticalAlign11(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <span style="font-size: 0">
        <img src="pattern.png" style="vertical-align: bottom">
        <img src="pattern.png" style="vertical-align: top; height: 100px">
      </span>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span := line.Box().Children[0]
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(96), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(0), "img2")
}

func TestVerticalAlign12(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Reference for the next test
	page := renderOnePage(t, `
      <span style="font-size: 0; vertical-align: top">
        <img src="pattern.png">
      </span>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span := line.Box().Children[0]
	img1 := span.Box().Children[0]
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(0), "img1")
}

func TestVerticalAlign13(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Should be the same as above
	page := renderOnePage(t, `
      <span style="font-size: 0; vertical-align: top; display: inline-block">
        <img src="pattern.png">
      </span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line1 := body.Box().Children[0]
	span := line1.Box().Children[0]
	line2 := span.Box().Children[0]
	img1 := line2.Box().Children[0]
	tu.AssertEqual(t, img1.Box().ElementTag(), "img", "img1")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(0), "img1")
}

func TestBoxDecorationBreakInlineSlice(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// http://www.w3.org/TR/css3-background/#the-box-decoration-break
	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 100px }
        span { font-family: weasyprint; box-decoration-break: slice;
               padding: 5px; border: 1px solid black }
      </style>
      <span>a<br/>b<br/>c</span>`)
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	line1, line2, line3 := unpack3(body)
	span := line1.Box().Children[0]
	tu.AssertEqual(t, span.Box().Width, pr.Float(16), "span")
	tu.AssertEqual(t, span.Box().MarginWidth(), pr.Float(16+5+1), "span")
	text, _ := unpack2(span)
	tu.AssertEqual(t, text.Box().PositionX, pr.Float(5+1), "text")
	span = line2.Box().Children[0]
	tu.AssertEqual(t, span.Box().Width, pr.Float(16), "span")
	tu.AssertEqual(t, span.Box().MarginWidth(), pr.Float(16), "span")
	text, _ = unpack2(span)
	tu.AssertEqual(t, text.Box().PositionX, pr.Float(0), "text")
	span = line3.Box().Children[0]
	tu.AssertEqual(t, span.Box().Width, pr.Float(16), "span")
	tu.AssertEqual(t, span.Box().MarginWidth(), pr.Float(16+5+1), "span")
	text = span.Box().Children[0]
	tu.AssertEqual(t, text.Box().PositionX, pr.Float(0), "text")
}

func TestBoxDecorationBreakInlineClone(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// http://www.w3.org/TR/css3-background/#the-box-decoration-break
	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 100px }
        span { font-size: 12pt; font-family: weasyprint;
               box-decoration-break: clone;
               padding: 5px; border: 1px solid black }
      </style>
      <span>a<br/>b<br/>c</span>`)
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	line1, line2, line3 := unpack3(body)
	span := line1.Box().Children[0]
	tu.AssertEqual(t, span.Box().Width, pr.Float(16), "span")
	tu.AssertEqual(t, span.Box().MarginWidth(), pr.Float(16+2*(5+1)), "span")
	text, _ := unpack2(span)
	tu.AssertEqual(t, text.Box().PositionX, pr.Float(5+1), "text")
	span = line2.Box().Children[0]
	tu.AssertEqual(t, span.Box().Width, pr.Float(16), "span")
	tu.AssertEqual(t, span.Box().MarginWidth(), pr.Float(16+2*(5+1)), "span")
	text, _ = unpack2(span)
	tu.AssertEqual(t, text.Box().PositionX, pr.Float(5+1), "text")
	span = line3.Box().Children[0]
	tu.AssertEqual(t, span.Box().Width, pr.Float(16), "span")
	tu.AssertEqual(t, span.Box().MarginWidth(), pr.Float(16+2*(5+1)), "span")
	text = span.Box().Children[0]
	tu.AssertEqual(t, text.Box().PositionX, pr.Float(5+1), "text")
}
