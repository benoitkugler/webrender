package layout

import (
	"fmt"
	"strings"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func TestLineBreakingNbsp(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1561
	page := renderOnePage(t, `
      <style>
        @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        body { font-family: weasyprint; width: 7.5em }
      </style>
      <body>a <span>b</span> c d&nbsp;<span>ef
    `)
	html := unpack1(page)
	body := unpack1(html)
	line1, line2 := unpack2(body)
	assertText(t, unpack1(line1), "a ")
	assertText(t, unpack1(line1.Box().Children[1]), "b")
	assertText(t, line1.Box().Children[2], " c")
	assertText(t, unpack1(line2), "d\u00a0")
	assertText(t, unpack1(line2.Box().Children[1]), "ef")
}

func TestLineBreakBeforeTrailingSpace(t *testing.T) {
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1852
	page := renderOnePage(t, `
        <p style="display: inline-block">test\u2028 </p>a
        <p style="display: inline-block">test\u2028</p>a
    `)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	p1, _, p2, _ := unpack4(line)
	tu.AssertEqual(t, p1.Box().Width, p2.Box().Width)
}

func TestTextFontSizeZero(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        p { font-size: 0; }
      </style>
      <p>test font size zero</p>
    `)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line := unpack1(paragraph)
	// zero-sized text boxes are removed
	tu.AssertEqual(t, len(line.Box().Children), 0)
	tu.AssertEqual(t, line.Box().Height, Fl(0))
	tu.AssertEqual(t, paragraph.Box().Height, Fl(0))
}

func TestTextFontSizeVerySmall(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1499
	page := renderOnePage(t, `
      <style>
        p { font-size: 0.00000001px }
      </style>
      <p>test font size zero</p>
    `)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line := unpack1(paragraph)
	tu.AssertEqual(t, line.Box().Height.V() < 0.001, true)
	tu.AssertEqual(t, paragraph.Box().Height.V() < 0.001, true)
}

func TestTextSpacedInlines(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
		<p>start <i><b>bi1</b> <b>bi2</b></i> <b>b1</b> end</p>
    `)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line := unpack1(paragraph)
	start, i, space, b, end := unpack5(line)

	assertText(t, start, "start ")
	assertText(t, space, " ")
	assertText(t, end, " end")
	if w := space.Box().Width.V(); w <= 0 {
		t.Fatalf("expected positive width, got %f", w)
	}

	bi1, space, bi2 := unpack3(i)
	bi1 = unpack1(bi1)
	bi2 = unpack1(bi2)
	assertText(t, bi1, "bi1")
	assertText(t, space, " ")
	assertText(t, bi2, "bi2")
	if w := space.Box().Width.V(); w <= 0 {
		t.Fatalf("expected positive width, got %f", w)
	}

	b1 := unpack1(b)
	assertText(t, b1, "b1")
}

func TestTextAlignLeft(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// <-------------------->  page, body
	//     +-----+
	// +---+     |
	// |   |     |
	// +---+-----+

	// ^   ^     ^          ^
	// x=0 x=40  x=100      x=200
	page := renderOnePage(t, `
      <style>
        @page { size: 200px }
      </style>
      <body>
        <img src="pattern.png" style="width: 40px"
        ><img src="pattern.png" style="width: 60px">`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	img1, img2 := unpack1(line), line.Box().Children[1]
	// initial value for text-align: left (in ltr text)
	tu.AssertEqual(t, img1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, img2.Box().PositionX, Fl(40))
}

func TestTextAlignRight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// <-------------------->  page, body
	//                +-----+
	//            +---+     |
	//            |   |     |
	//            +---+-----+

	// ^          ^   ^     ^
	// x=0        x=100     x=200
	//                x=140
	page := renderOnePage(t, `
      <style>
        @page { size: 200px }
        body { text-align: right }
      </style>
      <body>
        <img src="pattern.png" style="width: 40px"
        ><img src="pattern.png" style="width: 60px">`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	img1, img2 := unpack1(line), line.Box().Children[1]

	tu.AssertEqual(t, img1.Box().PositionX, Fl(100)) // 200 - 60 - 40
	tu.AssertEqual(t, img2.Box().PositionX, Fl(140)) // 200 - 60
}

func TestTextAlignCenter(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// <-------------------->  page, body
	//           +-----+
	//       +---+     |
	//       |   |     |
	//       +---+-----+

	// ^     ^   ^     ^
	// x=    x=50     x=150
	//           x=90
	page := renderOnePage(t, `
      <style>
        @page { size: 200px }
        body { text-align: center }
      </style>
      <body>
        <img src="pattern.png" style="width: 40px"
        ><img src="pattern.png" style="width: 60px">`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	img1, img2 := unpack1(line), line.Box().Children[1]

	tu.AssertEqual(t, img1.Box().PositionX, Fl(50))
	tu.AssertEqual(t, img2.Box().PositionX, Fl(90))
}

func TestTextAlignJustify(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 300px 1000px }
        body { text-align: justify }
      </style>
      <p><img src="pattern.png" style="width: 40px">
        <strong>
          <img src="pattern.png" style="width: 60px">
          <img src="pattern.png" style="width: 10px">
          <img src="pattern.png" style="width: 100px"
        ></strong><img src="pattern.png" style="width: 290px"
        ><!-- Last image will be on its own line. -->`)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line1, line2 := unpack1(paragraph), paragraph.Box().Children[1]
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5 := unpack1(line2)
	assertText(t, space1, " ")
	assertText(t, space2, " ")
	assertText(t, space3, " ")

	tu.AssertEqual(t, image1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, space1.Box().PositionX, Fl(40))
	tu.AssertEqual(t, strong.Box().PositionX, Fl(70))
	tu.AssertEqual(t, image2.Box().PositionX, Fl(70))
	tu.AssertEqual(t, space2.Box().PositionX, Fl(130))
	tu.AssertEqual(t, image3.Box().PositionX, Fl(160))
	tu.AssertEqual(t, space3.Box().PositionX, Fl(170))
	tu.AssertEqual(t, image4.Box().PositionX, Fl(200))
	tu.AssertEqual(t, strong.Box().Width.V(), Fl(230))
	tu.AssertEqual(t, image5.Box().PositionX, Fl(0))
}

func TestTextAlignJustifyAll(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 300px 1000px }
        body { text-align: justify-all }
      </style>
      <p><img src="pattern.png" style="width: 40px">
        <strong>
          <img src="pattern.png" style="width: 60px">
          <img src="pattern.png" style="width: 10px">
          <img src="pattern.png" style="width: 100px"
        ></strong><img src="pattern.png" style="width: 200px">
        <img src="pattern.png" style="width: 10px">`)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line1, line2 := unpack1(paragraph), paragraph.Box().Children[1]
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5, space4, image6 := unpack3(line2)
	assertText(t, space1, " ")
	assertText(t, space2, " ")
	assertText(t, space3, " ")
	assertText(t, space4, " ")

	tu.AssertEqual(t, image1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, space1.Box().PositionX, Fl(40))
	tu.AssertEqual(t, strong.Box().PositionX, Fl(70))
	tu.AssertEqual(t, image2.Box().PositionX, Fl(70))
	tu.AssertEqual(t, space2.Box().PositionX, Fl(130))
	tu.AssertEqual(t, image3.Box().PositionX, Fl(160))
	tu.AssertEqual(t, space3.Box().PositionX, Fl(170))
	tu.AssertEqual(t, image4.Box().PositionX, Fl(200))
	tu.AssertEqual(t, strong.Box().Width, Fl(230))

	tu.AssertEqual(t, image5.Box().PositionX, Fl(0))
	tu.AssertEqual(t, space4.Box().PositionX, Fl(200))
	tu.AssertEqual(t, image6.Box().PositionX, Fl(290))
}

func TestTextAlignAllLast(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 300px 1000px }
        body { text-align-all: justify; text-align-last: right }
      </style>
      <p><img src="pattern.png" style="width: 40px">
        <strong>
          <img src="pattern.png" style="width: 60px">
          <img src="pattern.png" style="width: 10px">
          <img src="pattern.png" style="width: 100px"
        ></strong><img src="pattern.png" style="width: 200px"
        ><img src="pattern.png" style="width: 10px">`)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line1, line2 := unpack1(paragraph), paragraph.Box().Children[1]
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5, image6 := unpack1(line2), line2.Box().Children[1]

	assertText(t, space1, " ")
	assertText(t, space2, " ")
	assertText(t, space3, " ")

	tu.AssertEqual(t, image1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, space1.Box().PositionX, Fl(40))
	tu.AssertEqual(t, strong.Box().PositionX, Fl(70))
	tu.AssertEqual(t, image2.Box().PositionX, Fl(70))
	tu.AssertEqual(t, space2.Box().PositionX, Fl(130))
	tu.AssertEqual(t, image3.Box().PositionX, Fl(160))
	tu.AssertEqual(t, space3.Box().PositionX, Fl(170))
	tu.AssertEqual(t, image4.Box().PositionX, Fl(200))
	tu.AssertEqual(t, strong.Box().Width, Fl(230))

	tu.AssertEqual(t, image5.Box().PositionX, Fl(90))
	tu.AssertEqual(t, image6.Box().PositionX, Fl(290))
}

func TestTextAlignNotEnoughSpace(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        p { text-align: center; width: 0 }
        span { display: inline-block }
      </style>
      <p><span>aaaaaaaaaaaaaaaaaaaaaaaaaa</span></p>`)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	span := unpack1(paragraph)
	tu.AssertEqual(t, span.Box().PositionX, Fl(0))
}

func TestTextAlignJustifyNoSpace(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// single-word line (zero spaces)
	page := renderOnePage(t, `
      <style>
        body { text-align: justify; width: 50px }
      </style>
      <p>Supercalifragilisticexpialidocious bar</p>
    `)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line1, _ := unpack1(paragraph), paragraph.Box().Children[1]
	text := unpack1(line1)
	tu.AssertEqual(t, text.Box().PositionX, Fl(0))
}

func TestTextAlignJustifyTextIndent(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// text-indent
	page := renderOnePage(t, `
      <style>
        @page { size: 300px 1000px }
        body { text-align: justify }
        p { text-indent: 3px }
      </style>
      <p><img src="pattern.png" style="width: 40px">
        <strong>
          <img src="pattern.png" style="width: 60px">
          <img src="pattern.png" style="width: 10px">
          <img src="pattern.png" style="width: 100px"
        ></strong><img src="pattern.png" style="width: 290px"
        ><!-- Last image will be on its own line. -->`)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line1, line2 := unpack1(paragraph), paragraph.Box().Children[1]
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5 := unpack1(line2)

	assertText(t, space1, " ")
	assertText(t, space2, " ")
	assertText(t, space3, " ")

	tu.AssertEqual(t, image1.Box().PositionX, Fl(3))
	tu.AssertEqual(t, space1.Box().PositionX, Fl(43))
	tu.AssertEqual(t, strong.Box().PositionX, Fl(72))
	tu.AssertEqual(t, image2.Box().PositionX, Fl(72))
	tu.AssertEqual(t, space2.Box().PositionX, Fl(132))
	tu.AssertEqual(t, image3.Box().PositionX, Fl(161))
	tu.AssertEqual(t, space3.Box().PositionX, Fl(171))
	tu.AssertEqual(t, image4.Box().PositionX, Fl(200))
	tu.AssertEqual(t, strong.Box().Width, Fl(228))

	tu.AssertEqual(t, image5.Box().PositionX, Fl(0))
}

func TestTextAlignJustifyNoBreakBetweenChildren(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test justification when line break happens between two inline children
	// that must stay together.
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/637
	page := renderOnePage(t, `
      <style>
        @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        p { text-align: justify; font-family: weasyprint; width: 7em }
      </style>
      <p>
        <span>a</span>
        <span>b</span>
        <span>bla</span><span>,</span>
        <span>b</span>
      </p>
    `)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line1, line2 := unpack1(paragraph), paragraph.Box().Children[1]
	span1, _, span2, _ := unpack4(line1)
	tu.AssertEqual(t, span1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, span2.Box().PositionX, Fl(6*16)) // 1 character + 5 spaces
	tu.AssertEqual(t, line1.Box().Width, Fl(7*16))     // 7em

	span1, span2, _, span3, _ := unpack5(line2)
	tu.AssertEqual(t, span1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, span2.Box().PositionX, Fl(3*16)) // 3 characters
	tu.AssertEqual(t, span3.Box().PositionX, Fl(5*16)) // (3 + 1) characters + 1 space
}

func TestWordSpacing(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// keep the empty <style> as a regression test: element.text is nil
	// (Not a string.)
	page := renderOnePage(t, `
      <style></style>
      <body><strong>Lorem ipsum dolor<em>sit amet</em></strong>`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	strong1 := unpack1(line)

	for _, text := range []string{
		"Lorem ipsum dolor<em>sit amet</em>",
		"Lorem ipsum <em>dolorsit</em> amet",
		"Lorem ipsum <em></em>dolorsit amet",
		"Lorem ipsum<em> </em>dolorsit amet",
		"Lorem ipsum<em> dolorsit</em> amet",
		"Lorem ipsum <em>dolorsit </em>amet",
	} {
		page = renderOnePage(t, fmt.Sprintf(`
		  <style>strong { word-spacing: 11px }</style>
		  <body><strong>%s</strong>`, text))
		html := unpack1(page)
		body = unpack1(html)
		line = unpack1(body)
		strong2 := unpack1(line)

		tu.AssertEqual(t, strong2.Box().Width.V()-strong1.Box().Width.V(), Fl(33))
	}
}

func TestLetterSpacing1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
        <body><strong>Supercalifragilisticexpialidocious</strong>`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	strong1 := unpack1(line)

	page = renderOnePage(t, `
        <style>strong { letter-spacing: 11px }</style>
        <body><strong>Supercalifragilisticexpialidocious</strong>`)
	html = unpack1(page)
	body = unpack1(html)
	line = unpack1(body)
	strong2 := unpack1(line)
	tu.AssertEqual(t, strong2.Box().Width.V()-strong1.Box().Width.V(), Fl(34*11))

	// an embedded tag should not affect the single-line letter spacing
	page = renderOnePage(t,
		"<style>strong { letter-spacing: 11px }</style>"+
			"<body><strong>Supercali<span>fragilistic</span>expialidocious"+
			"</strong>")
	html = unpack1(page)
	body = unpack1(html)
	line = unpack1(body)
	strong3 := unpack1(line)
	tu.AssertEqual(t, strong3.Box().Width, strong2.Box().Width)

	// duplicate wrapped lines should also have same overall width
	// Note work-around for word-wrap bug (issue #163) by marking word
	// as an inline-block
	page = renderOnePage(t, fmt.Sprintf(`<style>
          strong {
            letter-spacing: 11px;
            max-width: %fpx
        }
          span { display: inline-block }
        </style>
        <body><strong>
          <span>Supercali<i>fragilistic</i>expialidocious</span> 
          <span>Supercali<i>fragilistic</i>expialidocious</span>
        </strong>`, strong3.Box().Width.V()*1.5))
	html = unpack1(page)
	body = unpack1(html)
	line1, line2 := unpack1(body), body.Box().Children[1]
	tu.AssertEqual(t, unpack1(line1).Box().Width, unpack1(line2).Box().Width)
	tu.AssertEqual(t, unpack1(line1).Box().Width, strong2.Box().Width)
}

func TestSpacingEx(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test regression on ex units in spacing properties
	for _, spacing := range []string{"word-spacing", "letter-spacing"} {
		renderPages(t, fmt.Sprintf(`<div style="%s: 2ex">abc def`, spacing))
	}
}

func TestTextIndent(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, indent := range []string{"12px", "6%"} {
		page := renderOnePage(t, fmt.Sprintf(`
        <style>
            @page { size: 220px }
            body { margin: 10px; text-indent: %s }
        </style>
        <p>Some text that is long enough that it take at least three line,
           but maybe more.
    `, indent))
		html := unpack1(page)
		body := unpack1(html)
		paragraph := unpack1(body)
		lines := paragraph.Box().Children
		text1 := unpack1(lines[0])
		text2 := unpack1(lines[1])
		text3 := unpack1(lines[2])
		tu.AssertEqual(t, text1.Box().PositionX, Fl(22)) // 10px margin-left + 12px indent
		tu.AssertEqual(t, text2.Box().PositionX, Fl(10)) // No indent
		tu.AssertEqual(t, text3.Box().PositionX, Fl(10)) // No indent
	}
}

func TestTextIndentInline(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1000
	page := renderOnePage(t, `
        <style>
            @font-face { src: url(weasyprint.otf); font-family: weasyprint }
            p { display: inline-block; text-indent: 1em;
                font-family: weasyprint }
        </style>
        <p><span>text
    `)
	html := unpack1(page)
	body := unpack1(html)
	paragraph := unpack1(body)
	line := unpack1(paragraph)
	tu.AssertEqual(t, line.Box().Width, Fl((4+1)*16))
}

func TestTextIndentMultipage(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/706

	for _, indent := range []string{"12px", "6%"} {
		pages := renderPages(t, fmt.Sprintf(`
        <style>
            @page { size: 220px 1.5em; margin: 0 }
            body { margin: 10px; text-indent: %s }
        </style>
        <p>Some text that is long enough that it take at least three line,
           but maybe more.
    `, indent))
		page := pages[0]
		html := unpack1(page)
		body := unpack1(html)
		paragraph := unpack1(body)
		line := unpack1(paragraph)
		text := unpack1(line)
		tu.AssertEqual(t, text.Box().PositionX, Fl(22)) // 10px margin-left + 12px indent

		page = pages[1]
		html = unpack1(page)
		body = unpack1(html)
		paragraph = unpack1(body)
		line = unpack1(paragraph)
		text = unpack1(line)
		tu.AssertEqual(t, text.Box().PositionX, Fl(10)) // No indent
	}
}

func testHyphenateCharacter(t *testing.T, hyphChar string, replacer func(s string) string) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, fmt.Sprintf(`
        <html style="width: 5em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto;  hyphenate-character: '%s'" lang=fr>hyphénation`, hyphChar))
	html := unpack1(page)
	body := unpack1(html)
	lines := body.Box().Children
	if !(len(lines) > 1) {
		t.Fatalf("expected > 1, got %v", lines)
	}
	if text := unpack1(lines[0]).(*bo.TextBox).TextS(); !strings.HasSuffix(text, hyphChar) {
		t.Fatalf("unexpected %s", text)
	}
	fullText := ""
	for _, line := range lines {
		fullText += unpack1(line).(*bo.TextBox).TextS()
	}
	tu.AssertEqual(t, replacer(fullText), "hyphénation")
}

func TestHyphenateCharacter1(t *testing.T) {
	testHyphenateCharacter(t, "!", func(s string) string { return strings.ReplaceAll(s, "!", "") })
}

func TestHyphenateCharacter2(t *testing.T) {
	testHyphenateCharacter(t, "à", func(s string) string { return strings.ReplaceAll(s, "à", "") })
}

func TestHyphenateCharacter3(t *testing.T) {
	testHyphenateCharacter(t, "ù ù", func(s string) string { return strings.ReplaceAll(strings.ReplaceAll(s, "ù", ""), " ", "") })
}

func TestHyphenateCharacter4(t *testing.T) {
	testHyphenateCharacter(t, "", func(s string) string { return s })
}

func TestHyphenateCharacter5(t *testing.T) {
	testHyphenateCharacter(t, "———", func(s string) string { return strings.ReplaceAll(s, "—", "") })
}

func TestHyphenateManual1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	total := []rune("hyphénation")
	for i := 1; i < len(total); i++ {
		for _, hyphenateCharacter := range []string{"!", "ù ù"} {
			word := string(total[:i]) + "\u00ad" + string(total[i:])

			page := renderOnePage(t, fmt.Sprintf(`
			<html style="width: 5em; font-family: weasyprint" >
			<style>
			  @font-face {src: url(weasyprint.otf); font-family: weasyprint}
			</style>
			<body style="hyphens: manual;  hyphenate-character: '%s'" lang=fr>%s`, hyphenateCharacter, word))
			html := unpack1(page)
			body := unpack1(html)
			lines := body.Box().Children
			if !(len(lines) > 1) {
				t.Fatalf("expected > 1, got %v", lines)
			}
			if text := unpack1(lines[0]).(*bo.TextBox).TextS(); !strings.HasSuffix(text, hyphenateCharacter) {
				t.Fatalf("unexpected %s", text)
			}
			fullText := ""
			for _, line := range lines {
				fullText += unpack1(line).(*bo.TextBox).TextS()
			}
			tu.AssertEqual(t, strings.ReplaceAll(fullText, hyphenateCharacter, ""), word)

		}
	}
}

func TestHyphenateManual2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	total := []rune("hy phénation")
	for i := 1; i < len(total); i++ {
		for _, hyphenateCharacter := range []string{"!", "ù ù"} {
			word := string(total[:i]) + "\u00ad" + string(total[i:])

			page := renderOnePage(t, fmt.Sprintf(`
			<html style="width: 5em; font-family: weasyprint" >
			<style>
				@font-face {src: url(weasyprint.otf); font-family: weasyprint}
			</style>
			<body style="hyphens: manual;  hyphenate-character: '%s'" lang=fr>%s`, hyphenateCharacter, word))
			html := unpack1(page)
			body := unpack1(html)
			lines := body.Box().Children
			if L := len(lines); !(L == 2 || L == 3) {
				t.Fatalf("expected > 1, got %v", lines)
			}
			fullText := ""
			for _, line := range lines {
				fullText += textFromBoxes(line.Box().Children)
			}
			fullText = strings.ReplaceAll(fullText, hyphenateCharacter, "")
			if text := unpack1(lines[0]).(*bo.TextBox).TextS(); strings.HasSuffix(text, hyphenateCharacter) {
				tu.AssertEqual(t, fullText, word)
			} else {
				tu.AssertEqual(t, strings.HasSuffix(strings.TrimSuffix(text, "\u00ad"), "y"), true)
				if len(lines) == 3 {
					if text := unpack1(lines[1]).(*bo.TextBox).TextS(); !strings.HasSuffix(strings.TrimSuffix(text, "\u00ad"), hyphenateCharacter) {
						t.Fatalf("unexpected %s", text)
					}
				}
			}

		}
	}
}

func TestHyphenateManual3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Automatic hyphenation opportunities within a word must be ignored if the
	// word contains a conditional hyphen, in favor of the conditional
	// hyphen(s).
	page := renderOnePage(t,
		`<html style="width: 0.1em" lang="en">
        <body style="hyphens: auto">in&shy;lighten&shy;lighten&shy;in`)
	html := unpack1(page)
	body := unpack1(html)
	line1, line2, line3, line4 := unpack4(body)
	assertText(t, unpack1(line1), "in\u00ad-")
	assertText(t, unpack1(line2), "lighten\u00ad-")
	assertText(t, unpack1(line3), "lighten\u00ad-")
	assertText(t, unpack1(line4), "in")
}

func TestHyphenateManual4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1878
	page := renderOnePage(t,
		`<html style="width: 0.1em" lang="en">
        <body style="hyphens: auto">test&shy;`)
	html := unpack1(page)
	body := unpack1(html)
	line1 := unpack1(body)
	// TODO: should not end with an hyphen
	t.Skip()
	assertText(t, unpack1(line1), "test\xad")
}

func TestHyphenateLimitZone1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t,
		`<html style="width: 12em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto;
        hyphenate-limit-zone: 0" lang=fr>mmmmm hyphénation`)
	html := unpack1(page)
	body := unpack1(html)
	lines := body.Box().Children
	tu.AssertEqual(t, len(lines), 2)

	if text := unpack1(lines[0]).(*bo.TextBox).TextS(); !strings.HasSuffix(text, "-") {
		t.Fatalf("unexpected <%s>", text)
	}
	fullText := ""
	for _, line := range lines {
		fullText += unpack1(line).(*bo.TextBox).TextS()
	}
	tu.AssertEqual(t, strings.ReplaceAll(fullText, "-", ""), "mmmmm hyphénation")
}

func TestHyphenateLimitZone2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t,
		`<html style="width: 12em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto;
        hyphenate-limit-zone: 9em" lang=fr>mmmmm hyphénation`)
	html := unpack1(page)
	body := unpack1(html)
	lines := body.Box().Children
	tu.AssertEqual(t, len(lines), 2)

	if text := unpack1(lines[0]).(*bo.TextBox).TextS(); !strings.HasSuffix(text, "mm") {
		t.Fatalf("unexpected <%s>", text)
	}
	fullText := ""
	for _, line := range lines {
		fullText += unpack1(line).(*bo.TextBox).TextS()
	}
	tu.AssertEqual(t, fullText, "mmmmmhyphénation")
}

func TestHyphenateLimitZone3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t,
		`<html style="width: 12em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto;
        hyphenate-limit-zone: 5%" lang=fr>mmmmm hyphénation`)
	html := unpack1(page)
	body := unpack1(html)
	lines := body.Box().Children
	tu.AssertEqual(t, len(lines), 2)

	if text := unpack1(lines[0]).(*bo.TextBox).TextS(); !strings.HasSuffix(text, "-") {
		t.Fatalf("unexpected <%s>", text)
	}
	fullText := ""
	for _, line := range lines {
		fullText += unpack1(line).(*bo.TextBox).TextS()
	}
	tu.AssertEqual(t, strings.ReplaceAll(fullText, "-", ""), "mmmmm hyphénation")
}

func TestHyphenateLimitZone4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t,
		`<html style="width: 12em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto;
        hyphenate-limit-zone: 95%" lang=fr>mmmmm hyphénation`)
	html := unpack1(page)
	body := unpack1(html)
	lines := body.Box().Children
	tu.AssertEqual(t, len(lines), 2)

	if text := unpack1(lines[0]).(*bo.TextBox).TextS(); !strings.HasSuffix(text, "mm") {
		t.Fatalf("unexpected <%s>", text)
	}
	fullText := ""
	for _, line := range lines {
		fullText += unpack1(line).(*bo.TextBox).TextS()
	}
	tu.AssertEqual(t, fullText, "mmmmmhyphénation")
}

func TestHyphenateLimitChars(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, v := range []struct {
		css    string
		result int
	}{
		{"auto", 2},
		{"auto auto 0", 2},
		{"0 0 0", 2},
		{"4 4 auto", 1},
		{"6 2 4", 2},
		{"auto 1 auto", 2},
		{"7 auto auto", 1},
		{"6 auto auto", 2},
		{"5 2", 2},
		{"3", 2},
		{"2 4 6", 1},
		{"auto 4", 1},
		{"auto 2", 2},
	} {

		page := renderOnePage(t, fmt.Sprintf(`
        <html style="width: 1em; font-family: weasyprint">
        <style>
		@font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto; hyphenate-limit-chars: %s" lang=en>hyphen`, v.css))
		html := unpack1(page)
		body := unpack1(html)
		lines := body.Box().Children
		tu.AssertEqual(t, len(lines), v.result)
	}
}

func TestHyphenateLimitCharsPunctuation(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// See https://github.com/Kozea/WeasyPrint/issues/109
	for _, css := range []string{
		"3 3 3", // "en" is shorter than 3
		"3 6 2", // "light" is shorter than 6
		"8",     // "lighten" is shorter than 8
	} {
		page := renderOnePage(t, fmt.Sprintf(`
        <html style="width: 1em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto; hyphenate-limit-chars: %s" lang=en>..lighten..`, css))
		html := unpack1(page)
		body := unpack1(html)
		lines := body.Box().Children
		tu.AssertEqual(t, len(lines), 1)
	}
}

func TestOverflowWrap(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, v := range []struct {
		wrap, text string
		test       func(int) bool
		fullText   string
	}{
		{"anywhere", "aaaaaaaa", func(a int) bool { return a > 1 }, "aaaaaaaa"},
		{"break-word", "aaaaaaaa", func(a int) bool { return a > 1 }, "aaaaaaaa"},
		{"normal", "aaaaaaaa", func(a int) bool { return a == 1 }, "aaaaaaaa"},
		{"break-word", "hyphenations", func(a int) bool { return a > 3 }, "hy-phen-ations"},
		{"break-word", "A splitted word.  An hyphenated word.", func(a int) bool { return a > 8 }, "Asplittedword.Anhy-phen-atedword."},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        body {width: 80px; overflow: hidden; font-family: weasyprint; }
        span {overflow-wrap: %s; }
      </style>
      <body style="hyphens: auto;" lang="en"><span>%s`, v.wrap, v.text))
		html := unpack1(page)
		body := unpack1(html)
		var lines []string
		for _, line := range body.Box().Children {
			box := unpack1(line)
			textBox := unpack1(box).(*bo.TextBox)
			lines = append(lines, textBox.TextS())
		}
		if !v.test(len(lines)) {
			t.Fatal()
		}
		tu.AssertEqual(t, v.fullText, strings.Join(lines, ""))
	}
}

func TestOverflowWrap_2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, v := range []struct {
		wrap, text    string
		bodyWidth     int
		expectedWidth pr.Float
	}{
		{"anywhere", "aaaaaa", 10, 20},
		{"anywhere", "aaaaaa", 40, 40},
		{"break-word", "aaaaaa", 40, 120},
		{"normal", "aaaaaa", 40, 120},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
		<style>
		@font-face {src: url(weasyprint.otf); font-family: weasyprint}
		body {width: %dpx; font-family: weasyprint; font-size: 20px}
		table {overflow-wrap: %s}
	  </style>
	  <table><tr><td>%s`, v.bodyWidth, v.wrap, v.text))
		html := unpack1(page)
		body := unpack1(html)

		tableWrapper := unpack1(body)
		table := unpack1(tableWrapper)
		rowGroup := unpack1(table)
		tr := unpack1(rowGroup)
		td := unpack1(tr)
		tu.AssertEqual(t, td.Box().Width, v.expectedWidth)
	}
}

func TestWrapOverflowWordBreak(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range []struct {
		spanCSS       string
		expectedLines []string
	}{
		// overflow-wrap: anywhere and break-word are only allowed to break a word
		// "if there are no otherwise-acceptable break points in the line", which
		// means they should not split a word if it fits cleanly into the next line.
		// This can be done accidentally if it is in its own inline element.
		{"overflow-wrap: anywhere", []string{"aaa", "bbb"}},
		{"overflow-wrap: break-word", []string{"aaa", "bbb"}},

		// On the other hand, word-break: break-all mandates a break anywhere at the
		// end of a line, even if the word could fit cleanly onto the next line.
		{"word-break: break-all", []string{"aaa b", "bb"}},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
		<style>
			@font-face {src: url(weasyprint.otf); font-family: weasyprint}
			body {width: 80px; overflow: hidden; font-family: weasyprint}
			span {%s}
		</style>
		<body>
		<span>aaa </span><span>bbb
	`, data.spanCSS))
		html := unpack1(page)
		body := unpack1(html)
		var lines []string
		for _, line := range body.Box().Children {
			lineText := ""
			for _, span_box := range line.Box().Children {
				lineText += unpack1(span_box).(*bo.TextBox).TextS()
			}
			lines = append(lines, lineText)
		}
		tu.AssertEqual(t, lines, data.expectedLines)
	}
}

func TestOverflowWrapTrailingSpace(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, v := range []struct {
		wrap, text    string
		bodyWidth     int
		expectedWidth pr.Float
	}{
		{"anywhere", "aaaaaa", 10, 20},
		{"anywhere", "aaaaaa", 40, 40},
		{"break-word", "aaaaaa", 40, 120},
		{"normal", "abcdef", 40, 120},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
		<style>
		@font-face {src: url(weasyprint.otf); font-family: weasyprint}
		body {width: %dpx; font-family: weasyprint; font-size: 20px}
		table {overflow-wrap: %s}
	  </style>
	  <table><tr><td>%s `, v.bodyWidth, v.wrap, v.text))
		html := unpack1(page)
		body := unpack1(html)

		tableWrapper := unpack1(body)
		table := unpack1(tableWrapper)
		rowGroup := unpack1(table)
		tr := unpack1(rowGroup)
		td := unpack1(tr)
		tu.AssertEqual(t, td.Box().Width, v.expectedWidth)
	}
}

func testWhiteSpaceLines(t *testing.T, width int, space string, expected []string) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, fmt.Sprintf(`
      <style>
        body { font-size: 100px; width: %dpx }
        span { white-space: %s }
      </style>
      `, width, space)+"<body><span>This +    \n    is text")
	html := unpack1(page)
	body := unpack1(html)
	tu.AssertEqual(t, len(body.Box().Children), len(expected))
	for i, line := range body.Box().Children {
		box := unpack1(line)
		text := unpack1(box)
		assertText(t, text, expected[i])
	}
}

func TestWhiteSpace1(t *testing.T) {
	testWhiteSpaceLines(t, 1, "normal", []string{
		"This",
		"+",
		"is",
		"text",
	})
}

func TestWhiteSpace2(t *testing.T) {
	testWhiteSpaceLines(t, 1, "pre", []string{
		"This +    ",
		"    is text",
	})
}

func TestWhiteSpace3(t *testing.T) {
	testWhiteSpaceLines(t, 1, "nowrap", []string{"This + is text"})
}

func TestWhiteSpace4(t *testing.T) {
	testWhiteSpaceLines(t, 1, "pre-wrap", []string{
		"This ",
		"+    ",
		"    ",
		"is ",
		"text",
	})
}

func TestWhiteSpace5(t *testing.T) {
	testWhiteSpaceLines(t, 1, "pre-line", []string{
		"This",
		"+",
		"is",
		"text",
	})
}

func TestWhiteSpace6(t *testing.T) {
	testWhiteSpaceLines(t, 1000000, "normal", []string{"This + is text"})
}

func TestWhiteSpace7(t *testing.T) {
	testWhiteSpaceLines(t, 1000000, "pre", []string{
		"This +    ",
		"    is text",
	})
}

func TestWhiteSpace8(t *testing.T) {
	testWhiteSpaceLines(t, 1000000, "nowrap", []string{"This + is text"})
}

func TestWhiteSpace9(t *testing.T) {
	testWhiteSpaceLines(t, 1000000, "pre-wrap", []string{
		"This +    ",
		"    is text",
	})
}

func TestWhiteSpace10(t *testing.T) {
	testWhiteSpaceLines(t, 1000000, "pre-line", []string{
		"This +",
		"is text",
	})
}

func TestWhiteSpace11(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/813
	page := renderOnePage(t, `
      <style>
        pre { width: 0 }
      </style>
      <body><pre>This<br/>is text`)
	html := unpack1(page)
	body := unpack1(html)
	pre := unpack1(body)
	line1, line2 := unpack1(pre), pre.Box().Children[1]
	text1, box := unpack1(line1), line1.Box().Children[1]
	assertText(t, text1, "This")
	tu.AssertEqual(t, box.Box().ElementTag(), "br")
	text2 := unpack1(line2)
	assertText(t, text2, "is text")
}

func TestWhiteSpace12(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/813
	page := renderOnePage(t, `
      <style>
        pre { width: 0 }
      </style>
      <body><pre>This is <span>lol</span> text`)
	html := unpack1(page)
	body := unpack1(html)
	pre := unpack1(body)
	line1 := unpack1(pre)
	text1, span, text2 := unpack3(line1)
	assertText(t, text1, "This is ")
	tu.AssertEqual(t, span.Box().ElementTag(), "span")
	assertText(t, text2, " text")
}

func TestTabSize(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, v := range []struct {
		value string
		width pr.Float
	}{
		{"8", 144},   // (2 + (8 - 1)) * 16
		{"4", 80},    // (2 + (4 - 1)) * 16
		{"3em", 64},  // (2 + (3 - 1)) * 16
		{"25px", 41}, // 2 * 16 + 25 - 1 * 16
		// (0, 32),  // See Layout.setTabs
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        pre { tab-size: %s; font-family: weasyprint }
      </style>
      <pre>a&#9;a</pre>
    `, v.value))
		html := unpack1(page)
		body := unpack1(html)
		paragraph := unpack1(body)
		line := unpack1(paragraph)
		tu.AssertEqual(t, line.Box().Width, v.width)
	}
}

func TestTextTransform(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        p { text-transform: capitalize }
        p+p { text-transform: uppercase }
        p+p+p { text-transform: lowercase }
        p+p+p+p { text-transform: full-width }
        p+p+p+p+p { text-transform: none }
      </style>
<p>hé lO1</p><p>hé lO1</p><p>hé lO1</p><p>hé lO1</p><p>hé lO1</p>
    `)
	html := unpack1(page)
	body := unpack1(html)
	expected := []string{
		"Hé LO1",
		"HÉ LO1",
		"hé lo1",
		"\uff48é\u3000\uff4c\uff2f\uff11",
		"hé lO1",
	}
	tu.AssertEqual(t, len(body.Box().Children), len(expected))
	for i, child := range body.Box().Children {
		line := unpack1(child)
		assertText(t, unpack1(line), expected[i])
	}
}

func TestTextFloatingPreLine(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/610
	_ = renderOnePage(t, `
      <div style="float: left; white-space: pre-line">This is
      oh this end </div>
    `)
}

func TestLeaderContent(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, v := range []struct{ leader, content string }{
		{"dotted", "."},
		{"solid", "_"},
		{"space", " "},
		{`" .-"`, " .-"},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>div::after { content: leader(%s) }</style>
      <div></div>
    `, v.leader))
		html := unpack1(page)
		body := unpack1(html)
		div := unpack1(body)
		line := unpack1(div)
		after := unpack1(line)
		inline := unpack1(after)
		assertText(t, unpack1(inline), v.content)
	}
}

// expected fail
// func TestMaxLines(t *testing.T) {
// 	cp := tu.CaptureLogs()
// 	defer cp.AssertNoLogs(t)

// 	page := renderOnePage(t, `
//       <style>
//         @page {size: 10px 10px;}
//         @font-face {src: url(weasyprint.otf); font-family: weasyprint}
//         p {
//           font-family: weasyprint;
//           font-size: 2px;
//           max-lines: 2;
//         }
//       </style>
//       <p>
//         abcd efgh ijkl
//       </p>
//     `)
// 	html =  unpack1(page)
// 	body :=  unpack1(html)
// 	p1, p2 := unpack1(body), body.Box().Children[1]
// 	line1, line2 := unpack1(p1), p1.Box().Children[1]
// 	line3 := unpack1(p2)
// 	text1 := unpack1(line1)
// 	text2 := unpack1(line2)
// 	text3 := unpack1(line3)
// 	tu.AssertEqual(t, text1.(*bo.TextBox).Text, "abcd", "text1")
// 	tu.AssertEqual(t, text2.(*bo.TextBox).Text, "efgh", "text2")
// 	tu.AssertEqual(t, text3.(*bo.TextBox).Text, "ijkl", "text3")
// }

func TestContinue(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page {size: 10px 4px;}
        @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        div {
          continue: discard;
          font-family: weasyprint;
          font-size: 2px;
        }
      </style>
      <div>
        abcd efgh ijkl
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	p := unpack1(body)
	line1, line2 := unpack1(p), p.Box().Children[1]
	text1 := unpack1(line1)
	text2 := unpack1(line2)
	assertText(t, text1, "abcd")
	assertText(t, text2, "efgh")
}
