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
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, "<p> </p>")
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	tu.AssertEqual(t, len(paragraph.Box().Children), 0)
	tu.AssertEqual(t, paragraph.Box().Height, Fl(0))
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
//     html =  unpack1(page)
//     body :=  unpack1(html)
//     paragraph := unpack1(body)
//     // TODO: The second line should be removed
//     tu.AssertEqual(t, len(paragraph.Box().Children) , 1)

func TestBreakingLinebox(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	tu.AssertEqual(t, len(paragraph.Box().Children), 3)

	lines := paragraph.Box().Children
	for _, line := range lines {
		tu.AssertEqual(t, line.Box().Style.GetFontSize(), pr.FToV(13))
		tu.AssertEqual(t, line.Box().ElementTag(), "p")
		for _, child := range line.Box().Children {
			tu.AssertEqual(t, child.Box().ElementTag() == "em" || child.Box().ElementTag() == "p", true)
			tu.AssertEqual(t, child.Box().Style.GetFontSize(), pr.FToV(13))
			for _, childChild := range child.Box().Children {
				tu.AssertEqual(t, childChild.Box().ElementTag() == "em" || childChild.Box().ElementTag() == "strong" || childChild.Box().ElementTag() == "span", true)
				tu.AssertEqual(t, childChild.Box().Style.GetFontSize(), pr.FToV(13))
			}
		}
	}
}

func TestPositionXLtr(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        span {
          padding: 0 10px 0 15px;
          margin: 0 2px 0 3px;
          border: 1px solid;
         }
      </style>
      <body><span>a<br>b<br>c</span>`)
	html := unpack1(page)
	body := unpack1(html)
	line1, line2, line3 := unpack3(body)
	span1 := unpack1(line1)
	tu.AssertEqual(t, span1.Box().PositionX, Fl(0))
	text1, _ := unpack2(span1)
	tu.AssertEqual(t, text1.Box().PositionX, Fl(15+3+1))
	span2 := unpack1(line2)
	tu.AssertEqual(t, span2.Box().PositionX, Fl(0))
	text2, _ := unpack2(span2)
	tu.AssertEqual(t, text2.Box().PositionX, Fl(0))
	span3 := unpack1(line3)
	tu.AssertEqual(t, span3.Box().PositionX, Fl(0))
	text3 := unpack1(span3)
	tu.AssertEqual(t, text3.Box().PositionX, Fl(0))
}

func TestPositionXRtl(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	line1, line2, line3 := unpack3(body)
	span1 := unpack1(line1)
	text1, _ := unpack2(span1)
	tu.AssertEqual(t, span1.Box().PositionX, 100-text1.Box().Width.V()-(10+2+1))
	tu.AssertEqual(t, text1.Box().PositionX, 100-text1.Box().Width.V()-(10+2+1))
	span2 := unpack1(line2)
	text2, _ := unpack2(span2)
	tu.AssertEqual(t, span2.Box().PositionX, 100-text2.Box().Width.V())
	tu.AssertEqual(t, text2.Box().PositionX, 100-text2.Box().Width.V())
	span3 := unpack1(line3)
	text3 := unpack1(span3)
	tu.AssertEqual(t, span3.Box().PositionX, 100-text3.Box().Width.V()-(15+3+1))
	tu.AssertEqual(t, text3.Box().PositionX, 100-text3.Box().Width.V())
}

func TestBreakingLineboxRegression1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// See https://unicode.org/reports/tr14/
	page := renderOnePage(t, "<pre>a\nb\rc\r\nd\u2029e</pre>")
	html := unpack1(page)
	body := unpack1(html)
	pre := unpack1(body)
	lines := pre.Box().Children
	var texts []string
	for _, line := range lines {
		textBox := unpack1(line)
		texts = append(texts, textBox.(*bo.TextBox).TextS())
	}
	tu.AssertEqual(t, texts, []string{"a", "b", "c", "d", "e"})
}

func TestBreakingLineboxRegression2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	htmlSample := `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
      </style>
      <p style="width: %d.5em; font-family: weasyprint">ab
      <span style="padding-right: 1em; margin-right: 1em">c def</span>g
      hi</p>`
	for i := 0; i < 16; i++ {
		page := renderOnePage(t, fmt.Sprintf(htmlSample, i))
		html := unpack1(page)
		body := unpack1(html)
		p := unpack1(body)

		if i <= 3 {
			line1, line2, line3, line4 := unpack4(p)

			textbox1 := unpack1(line1)
			assertText(t, textbox1, "ab")

			span1 := unpack1(line2)
			textbox1 = unpack1(span1)
			assertText(t, textbox1, "c")

			span1, textbox2 := unpack2(line3)
			textbox1 = unpack1(span1)
			assertText(t, textbox1, "def")
			assertText(t, textbox2, "g")

			textbox1 = unpack1(line4)
			assertText(t, textbox1, "hi")
		} else if i <= 8 {
			line1, line2, line3 := unpack3(p)
			textbox1, span1 := unpack2(line1)
			assertText(t, textbox1, "ab ")
			textbox2 := unpack1(span1)
			assertText(t, textbox2, "c")

			span1, textbox2 = unpack2(line2)
			textbox1 = unpack1(span1)
			assertText(t, textbox1, "def")
			assertText(t, textbox2, "g")

			textbox1 = unpack1(line3)
			assertText(t, textbox1, "hi")
		} else if i <= 10 {
			line1, line2 := unpack2(p)

			textbox1, span1 := unpack2(line1)
			assertText(t, textbox1, "ab ")
			textbox2 := unpack1(span1)
			assertText(t, textbox2, "c")

			span1, textbox2 = unpack2(line2)
			textbox1 = unpack1(span1)
			assertText(t, textbox1, "def")
			assertText(t, textbox2, "g hi")
		} else if i <= 13 {
			line1, line2 := unpack2(p)

			textbox1, span1, textbox3 := unpack3(line1)
			assertText(t, textbox1, "ab ")
			textbox2 := unpack1(span1)
			assertText(t, textbox2, "c def")
			assertText(t, textbox3, "g")

			textbox1 = unpack1(line2)
			assertText(t, textbox1, "hi")
		} else {
			line1 := unpack1(p)

			textbox1, span1, textbox3 := unpack3(line1)
			assertText(t, textbox1, "ab ")
			textbox2 := unpack1(span1)
			assertText(t, textbox2, "c def")
			assertText(t, textbox3, "g hi")
		}
	}
}

func TestBreakingLineboxRegression3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test #1 for https://github.com/Kozea/WeasyPrint/issues/560
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <div style="width: 5.5em; font-family: weasyprint">
        aaaa aaaa a [<span>aaa</span>]`)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	line1, line2, line3, line4 := unpack4(div)
	tu.AssertEqual(t, unpack1(line1).(*bo.TextBox).TextS(), unpack1(line2).(*bo.TextBox).TextS())
	assertText(t, unpack1(line2), "aaaa")
	assertText(t, unpack1(line3), "a")
	text1, span, text2 := unpack3(line4)
	assertText(t, text1, "[")
	assertText(t, text2, "]")
	assertText(t, unpack1(span), "aaa")
}

func TestBreakingLineboxRegression4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test #2 for https://github.com/Kozea/WeasyPrint/issues/560
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <div style="width: 5.5em; font-family: weasyprint">
        aaaa a <span>b c</span>d`)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	line1, line2, line3 := unpack3(div)
	assertText(t, unpack1(line1), "aaaa")
	assertText(t, unpack1(line2), "a ")
	assertText(t, unpack1(line2.Box().Children[1]), "b")
	assertText(t, unpack1(line3.Box().Children[0]), "c")
	assertText(t, line3.Box().Children[1], "d")
}

func TestBreakingLineboxRegression5(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/580
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <div style="width: 5.5em; font-family: weasyprint">
        <span>aaaa aaaa a a a</span><span>bc</span>`)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	line1, line2, line3, line4 := unpack4(div)
	assertText(t, unpack1(line1.Box().Children[0]), "aaaa")
	assertText(t, unpack1(line2.Box().Children[0]), "aaaa")
	assertText(t, unpack1(line3.Box().Children[0]), "a a")
	assertText(t, unpack1(line4.Box().Children[0]), "a")
	assertText(t, unpack1(line4.Box().Children[1]), "bc")
}

func TestBreakingLineboxRegression6(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/586
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <div style="width: 5.5em; font-family: weasyprint">
        a a <span style="white-space: nowrap">/ccc</span>`)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	line1, line2 := unpack2(div)
	assertText(t, unpack1(line1), "a a")
	assertText(t, unpack1(line2.Box().Children[0]), "/ccc")
}

func TestBreakingLineboxRegression7(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/660
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <div style="width: 3.5em; font-family: weasyprint">
        <span><span>abc d e</span></span><span>f`)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	line1, line2, line3 := unpack3(div)
	assertText(t, unpack1(line1.Box().Children[0].Box().Children[0]), "abc")
	assertText(t, unpack1(line2.Box().Children[0].Box().Children[0]), "d")
	assertText(t, unpack1(line3.Box().Children[0].Box().Children[0]), "e")
	assertText(t, unpack1(line3.Box().Children[1]), "f")
}

func TestBreakingLineboxRegression8(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/783
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <p style="font-family: weasyprint"><span>
        aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
        bbbbbbbbbbb
        <b>cccc</b></span>ddd</p>`)
	html := unpack1(page)
	body := unpack1(html)
	p := unpack1(body)
	line1, line2 := unpack2(p)
	assertText(t, unpack1(unpack1(line1)), "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa bbbbbbbbbbb")
	assertText(t, unpack1(unpack1(unpack1(line2))), "cccc")
	assertText(t, line2.Box().Children[1], "ddd")
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
//     html =  unpack1(page)
//     body :=  unpack1(html)
//     p := unpack1(body)
//     line1, line2 = p.Box().Children
// assertText.AssertEqual(t, unpack1(line1.Box().Children[0]) , ()
//         "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbb")
//     tu.AssertEqual(t, unpack1(line2.Box().Children[0].Box().Children[0]).(*bo.TextBox).Text , "cccc")
//     tu.AssertEqual(t, line2.Box().Children[1].(*bo.TextBox).Text , "ddd")
// }

func TestBreakingLineboxRegression10(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	p := unpack1(body)
	line1, line2, line3, line4 := unpack4(p)
	assertText(t, unpack1(line1.Box().Children[0].Box().Children[0]), "xxxxxx YYY")
	assertText(t, unpack1(line2.Box().Children[0].Box().Children[0]), "yyyyyy yyy")
	assertText(t, unpack1(line3.Box().Children[0]), "ZZZZZZ zzzzz")
	assertText(t, unpack1(line4), ")x")
}

func TestBreakingLineboxRegression11(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/953
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <p style="width:10em; font-family: weasyprint">
          line 1<br><span>123 567 90</span>x
        </p>`)
	html := unpack1(page)
	body := unpack1(html)
	p := unpack1(body)
	line1, line2, line3 := unpack3(p)
	assertText(t, unpack1(line1), "line 1")
	assertText(t, unpack1(line2.Box().Children[0]), "123 567")
	assertText(t, unpack1(line3.Box().Children[0]), "90")
	assertText(t, line3.Box().Children[1], "x")
}

func TestBreakingLineboxRegression12(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/953
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <p style="width:10em; font-family: weasyprint">
          <br><span>123 567 90</span>x
        </p>`)
	html := unpack1(page)
	body := unpack1(html)
	p := unpack1(body)
	_, line2, line3 := unpack3(p)
	assertText(t, unpack1(line2.Box().Children[0]), "123 567")
	assertText(t, unpack1(line3.Box().Children[0]), "90")
	assertText(t, line3.Box().Children[1], "x")
}

func TestBreakingLineboxRegression13(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/953
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <p style="width:10em; font-family: weasyprint">
          123 567 90 <span>123 567 90</span>x
        </p>`)
	html := unpack1(page)
	body := unpack1(html)
	p := unpack1(body)
	line1, line2, line3 := unpack3(p)
	assertText(t, unpack1(line1), "123 567 90")
	assertText(t, unpack1(line2.Box().Children[0]), "123 567")
	assertText(t, unpack1(line3.Box().Children[0]), "90")
	assertText(t, line3.Box().Children[1], "x")
}

func TestBreakingLineboxRegression14(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1638
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
          body {font-family: weasyprint; width: 3em}
        </style>
        <span> <span>a</span> b</span><span>c</span>`)
	html := unpack1(page)
	body := unpack1(html)
	line1, line2 := unpack2(body)
	assertText(t, unpack1(line1.Box().Children[0].Box().Children[0]), "a")
	assertText(t, unpack1(line2.Box().Children[0]), "b")
	assertText(t, unpack1(line2.Box().Children[1]), "c")
}

func TestBreakingLineboxRegression15(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/ietf-tools/datatracker/issues/5507
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
          body {font-family: weasyprint; font-size: 4px}
          pre {float: left}
        </style>`+"<pre>ab©\ndéf\nghïj\nklm</pre>")
	html := unpack1(page)
	body := unpack1(html)
	pre := unpack1(body)
	line1, line2, line3, line4 := unpack4(pre)
	assertText(t, unpack1(line1), "ab©")
	assertText(t, unpack1(line2), "déf")
	assertText(t, unpack1(line3), "ghïj")
	assertText(t, unpack1(line4), "klm")
	tu.AssertEqual(t, unpack1(line1).Box().Width, Fl(4*3))
	tu.AssertEqual(t, unpack1(line2).Box().Width, Fl(4*3))
	tu.AssertEqual(t, unpack1(line3).Box().Width, Fl(4*4))
	tu.AssertEqual(t, unpack1(line4).Box().Width, Fl(4*3))
	tu.AssertEqual(t, pre.Box().Width, Fl(4*4))
}

func TestBreakingLineboxRegression_16(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1973
	page := renderOnePage(t,
		`<style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
          body {font-family: weasyprint; font-size: 4px}
          p {float: left}
        </style>`+
			"<p>tést</p><pre>ab©\ndéf\nghïj\nklm</pre>")

	html := unpack1(page)
	body := unpack1(html)
	p, pre := unpack2(body)
	line1 := unpack1(p)
	assertText(t, unpack1(line1), "tést")
	tu.AssertEqual(t, p.Box().Width, pr.Float(4*4))
	line1, line2, line3, line4 := unpack4(pre)
	assertText(t, unpack1(line1), "ab©")
	assertText(t, unpack1(line2), "déf")
	assertText(t, unpack1(line3), "ghïj")
	assertText(t, unpack1(line4), "klm")
	tu.AssertEqual(t, unpack1(line1).Box().Width, Fl(4*3))
	tu.AssertEqual(t, unpack1(line2).Box().Width, Fl(4*3))
	tu.AssertEqual(t, unpack1(line3).Box().Width, Fl(4*4))
	tu.AssertEqual(t, unpack1(line4).Box().Width, Fl(4*3))
}

func TestLineboxText(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, fmt.Sprintf(`
      <style>
        p { width: 165px; font-family:%s;}
      </style>
      <p><em>Lorem Ipsum</em>is very <strong>coool</strong></p>
    `, sansFonts))
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	lines := paragraph.Box().Children
	tu.AssertEqual(t, len(lines), 2)

	var chunks []string
	for _, line := range lines {
		s := ""
		for _, box := range bo.Descendants(line) {
			if box, ok := box.(*bo.TextBox); ok {
				s += box.TextS()
			}
		}
		chunks = append(chunks, s)
	}
	text := strings.Join(chunks, " ")
	tu.AssertEqual(t, text, "Lorem Ipsumis very coool")
}

func TestLineboxPositions(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range [][2]int{{165, 2}, {1, 5}, {0, 5}} {
		width, expectedLines := data[0], data[1]

		page := renderOnePage(t, fmt.Sprintf(`
		<style>
		  p { width:%dpx; font-family:%s;
			  line-height: 20px }
		</style>
		<p>this is test for <strong>Weasyprint</strong></p>`, width, sansFonts))
		html := unpack1(page)
		body := unpack1(html)
		paragraph := unpack1(body)
		lines := paragraph.Box().Children
		tu.AssertEqual(t, len(lines), expectedLines)

		refPositionY := lines[0].Box().PositionY
		refPositionX := lines[0].Box().PositionX
		for _, line := range lines {
			tu.AssertEqual(t, refPositionY, line.Box().PositionY)
			tu.AssertEqual(t, refPositionX, line.Box().PositionX)
			for _, box := range line.Box().Children {
				tu.AssertEqual(t, refPositionX, box.Box().PositionX)
				refPositionX += box.Box().Width.V()
				tu.AssertEqual(t, refPositionY, box.Box().PositionY)
			}
			tu.AssertEqual(t, refPositionX-line.Box().PositionX <= line.Box().Width.V(), true)
			refPositionX = line.Box().PositionX
			refPositionY += line.Box().Height.V()
		}
	}
}

func TestForcedLineBreaksPre(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// These lines should be small enough to fit on the default A4 page
	// with the default 12pt font-size.
	page := renderOnePage(t, `
      <style> pre { line-height: 42px }</style>
      <pre>Lorem ipsum dolor sit amet,
          consectetur adipiscing elit.


          Sed sollicitudin nibh

          et turpis molestie tristique.</pre>
	`)
	html := unpack1(page)
	body := unpack1(html)
	pre := unpack1(body)
	tu.AssertEqual(t, pre.Box().ElementTag(), "pre")
	lines := pre.Box().Children
	tu.AssertEqual(t, len(lines), 7)
	for _, line := range lines {
		if !bo.LineT.IsInstance(line) {
			t.Fatal()
		}
		tu.AssertEqual(t, line.Box().Height, Fl(42))
	}
}

func TestForcedLineBreaksParagraph(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style> p { line-height: 42px }</style>
      <p>Lorem ipsum dolor sit amet,<br>
        consectetur adipiscing elit.<br><br><br>
        Sed sollicitudin nibh<br>
        <br>
 
        et turpis molestie tristique.</p>
    `)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	tu.AssertEqual(t, paragraph.Box().ElementTag(), "p")
	lines := paragraph.Box().Children
	tu.AssertEqual(t, len(lines), 7)
	for _, line := range lines {
		if !bo.LineT.IsInstance(line) {
			t.Fatal()
		}
		tu.AssertEqual(t, line.Box().Height, Fl(42))
	}
}

func TestInlineboxSplitting(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// The text is strange to test some corner cases
	// See https://github.com/Kozea/WeasyPrint/issues/389
	for _, width := range []int{10000, 100, 10, 0} {
		page := renderOnePage(t, fmt.Sprintf(`
          <style>p { font-family:%s; width: %dpx; }</style>
          <p><strong>WeasyPrint is a frée softwäre ./ visual rendèring enginè
                     for HTML !!! && CSS.</strong></p>
        `, sansFonts, width))
		html := unpack1(page)
		body := unpack1(html)
		paragraph := unpack1(body)
		lines := paragraph.Box().Children
		if width == 10000 {
			tu.AssertEqual(t, len(lines), 1)
		} else {
			tu.AssertEqual(t, len(lines) > 1, true)
		}
		var textParts []string
		for _, line := range lines {
			strong := unpack1(line)
			text := unpack1(strong)
			textParts = append(textParts, text.(*bo.TextBox).TextS())
		}
		tu.AssertEqual(t, strings.Join(textParts, " "),
			"WeasyPrint is a frée softwäre ./ visual "+
				"rendèring enginè for HTML !!! && CSS.")
	}
}

func TestWhitespaceProcessing(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, source := range []string{"a", "  a  ", " \n  \ta", " a\t "} {
		page := renderOnePage(t, fmt.Sprintf("<p><em>%s</em></p>", source))
		html := unpack1(page)
		body := unpack1(html)
		p := unpack1(body)
		line := unpack1(p)
		em := unpack1(line)
		text := unpack1(em)
		assertText(t, text, "a")

		page = renderOnePage(t, fmt.Sprintf(
			`<p style="white-space: pre-line">
			
			<em>%s</em></pre>`, strings.ReplaceAll(source, "\n", " ")))
		html = unpack1(page)
		body = unpack1(html)
		p = unpack1(body)
		_, _, line3 := unpack3(p)
		em = unpack1(line3)
		text = unpack1(em)
		assertText(t, text, "a")
	}
}

func TestInlineReplacedAutoMargins(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 200px }
        img { display: inline; margin: auto; width: 50px }
      </style>
      <body><img src="pattern.png" />`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	img := unpack1(line)
	tu.AssertEqual(t, img.Box().MarginTop, Fl(0))
	tu.AssertEqual(t, img.Box().MarginRight, Fl(0))
	tu.AssertEqual(t, img.Box().MarginBottom, Fl(0))
	tu.AssertEqual(t, img.Box().MarginLeft, Fl(0))
}

func TestEmptyInlineAutoMargins(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 200px }
        span { margin: auto }
      </style>
      <body><span></span>`)
	html := unpack1(page)
	body := unpack1(html)
	block := unpack1(body)
	span := unpack1(block)
	tu.AssertEqual(t, span.Box().MarginTop != Fl(0), true)
	tu.AssertEqual(t, span.Box().MarginRight, Fl(0))
	tu.AssertEqual(t, span.Box().MarginBottom != Fl(0), true)
	tu.AssertEqual(t, span.Box().MarginLeft, Fl(0))
}

func TestFontStretch(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, fmt.Sprintf(`
      <style>
        p { float: left; font-family: %s }
      </style>
      <p>Hello, world!</p>
      <p style="font-stretch: condensed">Hello, world!</p>
    `, sansFonts))
	html := unpack1(page)
	body := unpack1(html)
	p1, p2 := unpack2(body)
	normal := p1.Box().Width.V()
	condensed := p2.Box().Width.V()
	tu.AssertEqual(t, condensed < normal, true)
}

func TestLineCount(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
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
		html := unpack1(page)
		body := unpack1(html)
		lines := body.Box().Children
		tu.AssertEqual(t, len(lines), data.lineCount)
	}
}

func TestVerticalAlign1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span := unpack1(line)
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, Fl(40))
	tu.AssertEqual(t, img2.Box().Height, Fl(60))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(20))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(0))
	// 60px + the descent of the font below the baseline
	tu.AssertEqual(t, 60 < line.Box().Height.V(), true)
	tu.AssertEqual(t, line.Box().Height.V() < 70, true)
	tu.AssertEqual(t, body.Box().Height, line.Box().Height)
}

func TestVerticalAlign2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span := unpack1(line)
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, Fl(40))
	tu.AssertEqual(t, img2.Box().Height, Fl(60))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(35))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(0))
	tu.AssertEqual(t, line.Box().Height, Fl(75))
	tu.AssertEqual(t, body.Box().Height, line.Box().Height)
}

func TestVerticalAlign3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Same as previously, but with percentages
	page := renderOnePage(t, `
      <span style="line-height: 10px">
        <img src="pattern.png" style="width: 40px; vertical-align: -150%"
        ><img src="pattern.png" style="width: 60px"></span>`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span := unpack1(line)
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, Fl(40))
	tu.AssertEqual(t, img2.Box().Height, Fl(60))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(35))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(0))
	tu.AssertEqual(t, line.Box().Height, Fl(75))
	tu.AssertEqual(t, body.Box().Height, line.Box().Height)
}

func TestVerticalAlign4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Same again, but have the vertical-align on an inline box.
	page := renderOnePage(t, `
      <span style="line-height: 10px">
        <span style="line-height: 10px; vertical-align: -15px">
          <img src="pattern.png" style="width: 40px"></span>
        <img src="pattern.png" style="width: 60px"></span>`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span1 := unpack1(line)
	span2, _, img2 := unpack3(span1)
	img1 := unpack1(span2)
	tu.AssertEqual(t, img1.Box().Height, Fl(40))
	tu.AssertEqual(t, img2.Box().Height, Fl(60))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(35))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(0))
	tu.AssertEqual(t, line.Box().Height, Fl(75))
	tu.AssertEqual(t, body.Box().Height, line.Box().Height)
}

func TestVerticalAlign5(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Same as previously, but with percentages
	page := renderOnePage(t, `
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <span style="line-height: 12px; font-size: 12px;
			font-family: weasyprint"><img src="pattern.png" 
			style="width: 40px; vertical-align: middle"><img src="pattern.png" 
			style="width: 60px"></span>`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span := unpack1(line)
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, Fl(40))
	tu.AssertEqual(t, img2.Box().Height, Fl(60))
	// middle of the image (positionY + 20) is at half the ex-height above
	// the baseline of the parent. The ex-height of weasyprint.otf is 0.8em
	tu.AssertEqual(t, img1.Box().PositionY, Fl(35.201202))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(0))
	tu.AssertEqual(t, line.Box().Height, Fl(75.2012))
	tu.AssertEqual(t, body.Box().Height, line.Box().Height)
}

func TestVerticalAlign6(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// sup and sub currently mean +/- 0.5 em
	// With the initial 16px font-size, that’s 8px.
	page := renderOnePage(t, `
      <span style="line-height: 10px">
        <img src="pattern.png" style="width: 60px"
        ><img src="pattern.png" style="width: 40px; vertical-align: super"
        ><img src="pattern.png" style="width: 40px; vertical-align: sub"
      ></span>`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span := unpack1(line)
	img1, img2, img3 := unpack3(span)
	tu.AssertEqual(t, img1.Box().Height, Fl(60))
	tu.AssertEqual(t, img2.Box().Height, Fl(40))
	tu.AssertEqual(t, img3.Box().Height, Fl(40))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(12))
	tu.AssertEqual(t, img3.Box().PositionY, Fl(28))
	tu.AssertEqual(t, line.Box().Height, Fl(68))
	tu.AssertEqual(t, body.Box().Height, line.Box().Height)
}

func TestVerticalAlign7(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <body style="line-height: 10px">
        <span>
          <img src="pattern.png" style="vertical-align: text-top"
          ><img src="pattern.png" style="vertical-align: text-bottom"
        ></span>`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span := unpack1(line)
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, Fl(4))
	tu.AssertEqual(t, img2.Box().Height, Fl(4))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(12))
	tu.AssertEqual(t, line.Box().Height, Fl(16))
	tu.AssertEqual(t, body.Box().Height, line.Box().Height)
}

func TestVerticalAlign8(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// This case used to cause an exception:
	// The second span has no children but should count for line heights
	// since it has padding.
	page := renderOnePage(t, `<span style="line-height: 1.5">
      <span style="padding: 1px"></span></span>`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span1 := unpack1(line)
	span2 := unpack1(span1)
	tu.AssertEqual(t, span1.Box().Height, Fl(16))
	tu.AssertEqual(t, span2.Box().Height, Fl(16))
	// The line’s strut does not has "line-height: normal" but the result should
	// be smaller than 1.5.
	tu.AssertEqual(t, span1.Box().MarginHeight(), Fl(24))
	tu.AssertEqual(t, span2.Box().MarginHeight(), Fl(24))
	tu.AssertEqual(t, line.Box().Height, Fl(24))
}

func TestVerticalAlign9(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span, div1 := unpack2(line)
	tu.AssertEqual(t, line.Box().Height, Fl(178))
	tu.AssertEqual(t, body.Box().Height, line.Box().Height)

	// Same as earlier
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().Height, Fl(40))
	tu.AssertEqual(t, img2.Box().Height, Fl(60))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(138))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(103))

	div2 := unpack1(div1)
	div3, div4 := unpack2(div2)
	divLine := unpack1(div4)
	divImg1, divImg2 := unpack2(divLine)
	tu.AssertEqual(t, div1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div1.Box().Height, Fl(175))
	tu.AssertEqual(t, div3.Box().Height, Fl(100))
	tu.AssertEqual(t, divLine.Box().Height, Fl(75))
	tu.AssertEqual(t, divImg1.Box().Height, Fl(40))
	tu.AssertEqual(t, divImg2.Box().Height, Fl(60))
	tu.AssertEqual(t, divImg1.Box().PositionY, Fl(135))
	tu.AssertEqual(t, divImg2.Box().PositionY, Fl(100))
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
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span1 := unpack1(line)
	img1, img2, span2, span4 := unpack4(span1)
	img3, span3 := unpack2(span2)
	img4 := unpack1(span3)
	img5 := unpack1(span4)
	tu.AssertEqual(t, body.Box().Height, line.Box().Height)
	tu.AssertEqual(t, line.Box().Height, Fl(40))
	tu.AssertEqual(t, img1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(36))
	tu.AssertEqual(t, img3.Box().PositionY, Fl(6))
	tu.AssertEqual(t, img4.Box().PositionY, Fl(36))
	tu.AssertEqual(t, img5.Box().PositionY, Fl(30))
}

func TestVerticalAlign11(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <span style="font-size: 0">
        <img src="pattern.png" style="vertical-align: bottom">
        <img src="pattern.png" style="vertical-align: top; height: 100px">
      </span>
    `)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span := unpack1(line)
	img1, img2 := unpack2(span)
	tu.AssertEqual(t, img1.Box().PositionY, Fl(96))
	tu.AssertEqual(t, img2.Box().PositionY, Fl(0))
}

func TestVerticalAlign12(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Reference for the next test
	page := renderOnePage(t, `
      <span style="font-size: 0; vertical-align: top">
        <img src="pattern.png">
      </span>
    `)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span := unpack1(line)
	img1 := unpack1(span)
	tu.AssertEqual(t, img1.Box().PositionY, Fl(0))
}

func TestVerticalAlign13(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Should be the same as above
	page := renderOnePage(t, `
      <span style="font-size: 0; vertical-align: top; display: inline-block">
        <img src="pattern.png">
      </span>`)
	html := unpack1(page)
	body := unpack1(html)
	line1 := unpack1(body)
	span := unpack1(line1)
	line2 := unpack1(span)
	img1 := unpack1(line2)
	tu.AssertEqual(t, img1.Box().ElementTag(), "img")
	tu.AssertEqual(t, img1.Box().PositionY, Fl(0))
}

func TestBoxDecorationBreakInlineSlice(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// https://www.w3.org/TR/css-backgrounds-3/#the-box-decoration-break
	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 100px }
        span { font-family: weasyprint; box-decoration-break: slice;
               padding: 5px; border: 1px solid black }
      </style>
      <span>a<br/>b<br/>c</span>`)
	html := unpack1(page1)
	body := unpack1(html)
	line1, line2, line3 := unpack3(body)
	span := unpack1(line1)
	tu.AssertEqual(t, span.Box().Width, Fl(16))
	tu.AssertEqual(t, span.Box().MarginWidth(), Fl(16+5+1))
	text, _ := unpack2(span)
	tu.AssertEqual(t, text.Box().PositionX, Fl(5+1))
	span = unpack1(line2)
	tu.AssertEqual(t, span.Box().Width, Fl(16))
	tu.AssertEqual(t, span.Box().MarginWidth(), Fl(16))
	text, _ = unpack2(span)
	tu.AssertEqual(t, text.Box().PositionX, Fl(0))
	span = unpack1(line3)
	tu.AssertEqual(t, span.Box().Width, Fl(16))
	tu.AssertEqual(t, span.Box().MarginWidth(), Fl(16+5+1))
	text = unpack1(span)
	tu.AssertEqual(t, text.Box().PositionX, Fl(0))
}

func TestBoxDecorationBreakInlineClone(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// https://www.w3.org/TR/css-backgrounds-3/#the-box-decoration-break
	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 100px }
        span { font-size: 12pt; font-family: weasyprint;
               box-decoration-break: clone;
               padding: 5px; border: 1px solid black }
      </style>
      <span>a<br/>b<br/>c</span>`)
	html := unpack1(page1)
	body := unpack1(html)
	line1, line2, line3 := unpack3(body)
	span := unpack1(line1)
	tu.AssertEqual(t, span.Box().Width, Fl(16))
	tu.AssertEqual(t, span.Box().MarginWidth(), Fl(16+2*(5+1)))
	text, _ := unpack2(span)
	tu.AssertEqual(t, text.Box().PositionX, Fl(5+1))
	span = unpack1(line2)
	tu.AssertEqual(t, span.Box().Width, Fl(16))
	tu.AssertEqual(t, span.Box().MarginWidth(), Fl(16+2*(5+1)))
	text, _ = unpack2(span)
	tu.AssertEqual(t, text.Box().PositionX, Fl(5+1))
	span = unpack1(line3)
	tu.AssertEqual(t, span.Box().Width, Fl(16))
	tu.AssertEqual(t, span.Box().MarginWidth(), Fl(16+2*(5+1)))
	text = unpack1(span)
	tu.AssertEqual(t, text.Box().PositionX, Fl(5+1))
}

func TestBidiPositionXInvariant(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        .float-border {
          float: left;
          border-right: 100px solid black;
        }
      </style>
      <div class="float-border" style="direction: ltr">abc</div>
      <div>&nbsp;</div>
      <div class="float-border" style="direction: rtl">abc</div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	block_ltr, _, block_rtl := unpack3(body)

	line_ltr := unpack1(block_ltr)
	text_ltr := unpack1(line_ltr)

	line_rtl := unpack1(block_rtl)
	text_rtl := unpack1(line_rtl)

	tu.AssertEqual(t, block_ltr.Box().PositionX, block_rtl.Box().PositionX)
	tu.AssertEqual(t, line_ltr.Box().PositionX, line_rtl.Box().PositionX)
	tu.AssertEqual(t, text_ltr.Box().PositionX, text_rtl.Box().PositionX)
}
