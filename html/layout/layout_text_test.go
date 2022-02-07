package layout

import (
	"fmt"
	"strings"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/utils/testutils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func TestTextFontSizeZero(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        p { font-size: 0; }
      </style>
      <p>test font size zero</p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line := paragraph.Box().Children[0]
	// zero-sized text boxes are removed
	tu.AssertEqual(t, len(line.Box().Children), 0, "children")
	tu.AssertEqual(t, line.Box().Height, pr.Float(0), "line")
	tu.AssertEqual(t, paragraph.Box().Height, pr.Float(0), "paragraph")
}

func TestTextFontSizeVerySmall(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1499
	page := renderOnePage(t, `
      <style>
        p { font-size: 0.00000001px }
      </style>
      <p>test font size zero</p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line := paragraph.Box().Children[0]
	tu.AssertEqual(t, line.Box().Height.V() < 0.001, true, "")
	tu.AssertEqual(t, paragraph.Box().Height.V() < 0.001, true, "")
}

func TestTextSpacedInlines(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
		<p>start <i><b>bi1</b> <b>bi2</b></i> <b>b1</b> end</p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line := paragraph.Box().Children[0]
	start, i, space, b, end := unpack5(line)

	tu.AssertEqual(t, start.(*bo.TextBox).Text, "start ", "start")
	tu.AssertEqual(t, space.(*bo.TextBox).Text, " ", "space")
	tu.AssertEqual(t, end.(*bo.TextBox).Text, " end", "end")
	if w := space.Box().Width.V(); w <= 0 {
		t.Fatalf("expected positive width, got %f", w)
	}

	bi1, space, bi2 := unpack3(i)
	bi1 = bi1.Box().Children[0]
	bi2 = bi2.Box().Children[0]
	tu.AssertEqual(t, bi1.(*bo.TextBox).Text, "bi1", "bi1")
	tu.AssertEqual(t, space.(*bo.TextBox).Text, " ", "space")
	tu.AssertEqual(t, bi2.(*bo.TextBox).Text, "bi2", "bi2")
	if w := space.Box().Width.V(); w <= 0 {
		t.Fatalf("expected positive width, got %f", w)
	}

	b1 := b.Box().Children[0]
	tu.AssertEqual(t, b1.(*bo.TextBox).Text, "b1", "b1")
}

func TestTextAlignLeft(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img1, img2 := line.Box().Children[0], line.Box().Children[1]
	// initial value for text-align: left (in ltr text)
	tu.AssertEqual(t, img1.Box().PositionX, pr.Float(0), "img1")
	tu.AssertEqual(t, img2.Box().PositionX, pr.Float(40), "img2")
}

func TestTextAlignRight(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img1, img2 := line.Box().Children[0], line.Box().Children[1]

	tu.AssertEqual(t, img1.Box().PositionX, pr.Float(100), "img1") // 200 - 60 - 40
	tu.AssertEqual(t, img2.Box().PositionX, pr.Float(140), "img2") // 200 - 60
}

func TestTextAlignCenter(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img1, img2 := line.Box().Children[0], line.Box().Children[1]

	tu.AssertEqual(t, img1.Box().PositionX, pr.Float(50), "img1")
	tu.AssertEqual(t, img2.Box().PositionX, pr.Float(90), "img2")
}

func TestTextAlignJustify(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2 := paragraph.Box().Children[0], paragraph.Box().Children[1]
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5 := line2.Box().Children[0]
	tu.AssertEqual(t, space1.(*bo.TextBox).Text, " ", "space1")
	tu.AssertEqual(t, space2.(*bo.TextBox).Text, " ", "space2")
	tu.AssertEqual(t, space3.(*bo.TextBox).Text, " ", "space3")

	tu.AssertEqual(t, image1.Box().PositionX, pr.Float(0), "image1")
	tu.AssertEqual(t, space1.Box().PositionX, pr.Float(40), "space1")
	tu.AssertEqual(t, strong.Box().PositionX, pr.Float(70), "strong")
	tu.AssertEqual(t, image2.Box().PositionX, pr.Float(70), "image2")
	tu.AssertEqual(t, space2.Box().PositionX, pr.Float(130), "space2")
	tu.AssertEqual(t, image3.Box().PositionX, pr.Float(160), "image3")
	tu.AssertEqual(t, space3.Box().PositionX, pr.Float(170), "space3")
	tu.AssertEqual(t, image4.Box().PositionX, pr.Float(200), "image4")
	tu.AssertEqual(t, strong.Box().Width.V(), pr.Float(230), "strong")
	tu.AssertEqual(t, image5.Box().PositionX, pr.Float(0), "image5")
}

func TestTextAlignJustifyAll(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2 := paragraph.Box().Children[0], paragraph.Box().Children[1]
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5, space4, image6 := unpack3(line2)
	tu.AssertEqual(t, space1.(*bo.TextBox).Text, " ", "space1")
	tu.AssertEqual(t, space2.(*bo.TextBox).Text, " ", "space2")
	tu.AssertEqual(t, space3.(*bo.TextBox).Text, " ", "space3")
	tu.AssertEqual(t, space4.(*bo.TextBox).Text, " ", "space4")

	tu.AssertEqual(t, image1.Box().PositionX, pr.Float(0), "image1")
	tu.AssertEqual(t, space1.Box().PositionX, pr.Float(40), "space1")
	tu.AssertEqual(t, strong.Box().PositionX, pr.Float(70), "strong")
	tu.AssertEqual(t, image2.Box().PositionX, pr.Float(70), "image2")
	tu.AssertEqual(t, space2.Box().PositionX, pr.Float(130), "space2")
	tu.AssertEqual(t, image3.Box().PositionX, pr.Float(160), "image3")
	tu.AssertEqual(t, space3.Box().PositionX, pr.Float(170), "space3")
	tu.AssertEqual(t, image4.Box().PositionX, pr.Float(200), "image4")
	tu.AssertEqual(t, strong.Box().Width, pr.Float(230), "strong")

	tu.AssertEqual(t, image5.Box().PositionX, pr.Float(0), "image5")
	tu.AssertEqual(t, space4.Box().PositionX, pr.Float(200), "space4")
	tu.AssertEqual(t, image6.Box().PositionX, pr.Float(290), "image6")
}

func TestTextAlignAllLast(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2 := paragraph.Box().Children[0], paragraph.Box().Children[1]
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5, image6 := line2.Box().Children[0], line2.Box().Children[1]

	tu.AssertEqual(t, space1.(*bo.TextBox).Text, " ", "space1")
	tu.AssertEqual(t, space2.(*bo.TextBox).Text, " ", "space2")
	tu.AssertEqual(t, space3.(*bo.TextBox).Text, " ", "space3")

	tu.AssertEqual(t, image1.Box().PositionX, pr.Float(0), "image1")
	tu.AssertEqual(t, space1.Box().PositionX, pr.Float(40), "space1")
	tu.AssertEqual(t, strong.Box().PositionX, pr.Float(70), "strong")
	tu.AssertEqual(t, image2.Box().PositionX, pr.Float(70), "image2")
	tu.AssertEqual(t, space2.Box().PositionX, pr.Float(130), "space2")
	tu.AssertEqual(t, image3.Box().PositionX, pr.Float(160), "image3")
	tu.AssertEqual(t, space3.Box().PositionX, pr.Float(170), "space3")
	tu.AssertEqual(t, image4.Box().PositionX, pr.Float(200), "image4")
	tu.AssertEqual(t, strong.Box().Width, pr.Float(230), "strong")

	tu.AssertEqual(t, image5.Box().PositionX, pr.Float(90), "image5")
	tu.AssertEqual(t, image6.Box().PositionX, pr.Float(290), "image6")
}

func TestTextAlignNotEnoughSpace(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        p { text-align: center; width: 0 }
        span { display: inline-block }
      </style>
      <p><span>aaaaaaaaaaaaaaaaaaaaaaaaaa</span></p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	span := paragraph.Box().Children[0]
	tu.AssertEqual(t, span.Box().PositionX, pr.Float(0), "span")
}

func TestTextAlignJustifyNoSpace(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// single-word line (zero spaces)
	page := renderOnePage(t, `
      <style>
        body { text-align: justify; width: 50px }
      </style>
      <p>Supercalifragilisticexpialidocious bar</p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, _ := paragraph.Box().Children[0], paragraph.Box().Children[1]
	text := line1.Box().Children[0]
	tu.AssertEqual(t, text.Box().PositionX, pr.Float(0), "text")
}

func TestTextAlignJustifyTextIndent(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2 := paragraph.Box().Children[0], paragraph.Box().Children[1]
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5 := line2.Box().Children[0]

	tu.AssertEqual(t, space1.(*bo.TextBox).Text, " ", "space1")
	tu.AssertEqual(t, space2.(*bo.TextBox).Text, " ", "space2")
	tu.AssertEqual(t, space3.(*bo.TextBox).Text, " ", "space3")

	tu.AssertEqual(t, image1.Box().PositionX, pr.Float(3), "image1")
	tu.AssertEqual(t, space1.Box().PositionX, pr.Float(43), "space1")
	tu.AssertEqual(t, strong.Box().PositionX, pr.Float(72), "strong")
	tu.AssertEqual(t, image2.Box().PositionX, pr.Float(72), "image2")
	tu.AssertEqual(t, space2.Box().PositionX, pr.Float(132), "space2")
	tu.AssertEqual(t, image3.Box().PositionX, pr.Float(161), "image3")
	tu.AssertEqual(t, space3.Box().PositionX, pr.Float(171), "space3")
	tu.AssertEqual(t, image4.Box().PositionX, pr.Float(200), "image4")
	tu.AssertEqual(t, strong.Box().Width, pr.Float(228), "strong")

	tu.AssertEqual(t, image5.Box().PositionX, pr.Float(0), "image5")
}

func TestTextAlignJustifyNoBreakBetweenChildren(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2 := paragraph.Box().Children[0], paragraph.Box().Children[1]
	span1, _, span2, _ := unpack4(line1)
	tu.AssertEqual(t, span1.Box().PositionX, pr.Float(0), "span1")
	tu.AssertEqual(t, span2.Box().PositionX, pr.Float(6*16), "span2") // 1 character + 5 spaces
	tu.AssertEqual(t, line1.Box().Width, pr.Float(7*16), "line1")     // 7em

	span1, span2, _, span3, _ := unpack5(line2)
	tu.AssertEqual(t, span1.Box().PositionX, pr.Float(0), "span1")
	tu.AssertEqual(t, span2.Box().PositionX, pr.Float(3*16), "span2") // 3 characters
	tu.AssertEqual(t, span3.Box().PositionX, pr.Float(5*16), "span3") // (3 + 1) characters + 1 space
}

func TestWordSpacing(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// keep the empty <style> as a regression test: element.text is nil
	// (Not a string.)
	page := renderOnePage(t, `
      <style></style>
      <body><strong>Lorem ipsum dolor<em>sit amet</em></strong>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	strong1 := line.Box().Children[0]

	page = renderOnePage(t, `
      <style>strong { word-spacing: 11px }</style>
      <body><strong>Lorem ipsum dolor<em>sit amet</em></strong>`)
	html = page.Box().Children[0]
	body = html.Box().Children[0]
	line = body.Box().Children[0]
	strong2 := line.Box().Children[0]
	tu.AssertEqual(t, strong2.Box().Width.V()-strong1.Box().Width.V(), pr.Float(33), "strong distance")
}

func TestLetterSpacing1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
        <body><strong>Supercalifragilisticexpialidocious</strong>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	strong1 := line.Box().Children[0]

	page = renderOnePage(t, `
        <style>strong { letter-spacing: 11px }</style>
        <body><strong>Supercalifragilisticexpialidocious</strong>`)
	html = page.Box().Children[0]
	body = html.Box().Children[0]
	line = body.Box().Children[0]
	strong2 := line.Box().Children[0]
	tu.AssertEqual(t, strong2.Box().Width.V()-strong1.Box().Width.V(), pr.Float(34*11), "strong distance")

	// an embedded tag should ! affect the single-line letter spacing
	page = renderOnePage(t,
		"<style>strong { letter-spacing: 11px }</style>"+
			"<body><strong>Supercali<span>fragilistic</span>expialidocious"+
			"</strong>")
	html = page.Box().Children[0]
	body = html.Box().Children[0]
	line = body.Box().Children[0]
	strong3 := line.Box().Children[0]
	tu.AssertEqual(t, strong3.Box().Width, strong2.Box().Width, "strong")

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
	html = page.Box().Children[0]
	body = html.Box().Children[0]
	line1, line2 := body.Box().Children[0], body.Box().Children[1]
	tu.AssertEqual(t, line1.Box().Children[0].Box().Width, line2.Box().Children[0].Box().Width, "")
	tu.AssertEqual(t, line1.Box().Children[0].Box().Width, strong2.Box().Width, "")
}

func TestSpacingEx(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Test regression on ex units in spacing properties
	for _, spacing := range []string{"word-spacing", "letter-spacing"} {
		renderPages(t, fmt.Sprintf(`<div style="%s: 2ex">abc def`, spacing))
	}
}

func TestTextIndent(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	for _, indent := range []string{"12px", "6%"} {
		page := renderOnePage(t, fmt.Sprintf(`
        <style>
            @page { size: 220px }
            body { margin: 10px; text-indent: %s }
        </style>
        <p>Some text that is long enough that it take at least three line,
           but maybe more.
    `, indent))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		paragraph := body.Box().Children[0]
		lines := paragraph.Box().Children
		text1 := lines[0].Box().Children[0]
		text2 := lines[1].Box().Children[0]
		text3 := lines[2].Box().Children[0]
		tu.AssertEqual(t, text1.Box().PositionX, pr.Float(22), "text1") // 10px margin-left + 12px indent
		tu.AssertEqual(t, text2.Box().PositionX, pr.Float(10), "text2") // No indent
		tu.AssertEqual(t, text3.Box().PositionX, pr.Float(10), "text3") // No indent
	}
}

func TestTextIndentInline(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1000
	page := renderOnePage(t, `
        <style>
            @font-face { src: url(weasyprint.otf); font-family: weasyprint }
            p { display: inline-block; text-indent: 1em;
                font-family: weasyprint }
        </style>
        <p><span>text
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line := paragraph.Box().Children[0]
	tu.AssertEqual(t, line.Box().Width, pr.Float((4+1)*16), "")
}

func TestTextIndentMultipage(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		paragraph := body.Box().Children[0]
		line := paragraph.Box().Children[0]
		text := line.Box().Children[0]
		tu.AssertEqual(t, text.Box().PositionX, pr.Float(22), "") // 10px margin-left + 12px indent

		page = pages[1]
		html = page.Box().Children[0]
		body = html.Box().Children[0]
		paragraph = body.Box().Children[0]
		line = paragraph.Box().Children[0]
		text = line.Box().Children[0]
		tu.AssertEqual(t, text.Box().PositionX, pr.Float(10), "") // No indent
	}
}

func testHyphenateCharacter(t *testing.T, hyphChar string, replacer func(s string) string) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, fmt.Sprintf(`
        <html style="width: 5em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto;  hyphenate-character: '%s'" lang=fr>hyphénation`, hyphChar))
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	lines := body.Box().Children
	if !(len(lines) > 1) {
		t.Fatalf("expected > 1, got %v", lines)
	}
	if text := lines[0].Box().Children[0].(*bo.TextBox).Text; !strings.HasSuffix(text, hyphChar) {
		t.Fatalf("unexpected %s", text)
	}
	fullText := ""
	for _, line := range lines {
		fullText += line.Box().Children[0].(*bo.TextBox).Text
	}
	tu.AssertEqual(t, replacer(fullText), "hyphénation", "")
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
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
			html := page.Box().Children[0]
			body := html.Box().Children[0]
			lines := body.Box().Children
			if !(len(lines) > 1) {
				t.Fatalf("expected > 1, got %v", lines)
			}
			if text := lines[0].Box().Children[0].(*bo.TextBox).Text; !strings.HasSuffix(text, hyphenateCharacter) {
				t.Fatalf("unexpected %s", text)
			}
			fullText := ""
			for _, line := range lines {
				fullText += line.Box().Children[0].(*bo.TextBox).Text
			}
			tu.AssertEqual(t, strings.ReplaceAll(fullText, hyphenateCharacter, ""), word, "")

		}
	}
}

func TestHyphenateManual2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
			html := page.Box().Children[0]
			body := html.Box().Children[0]
			lines := body.Box().Children
			if !(len(lines) > 1) {
				t.Fatalf("expected > 1, got %v", lines)
			}
			fullText := ""
			for _, line := range lines {
				fullText += line.Box().Children[0].(*bo.TextBox).Text
			}
			fullText = strings.ReplaceAll(fullText, hyphenateCharacter, "")
			if text := lines[0].Box().Children[0].(*bo.TextBox).Text; strings.HasSuffix(text, hyphenateCharacter) {
				tu.AssertEqual(t, fullText, word, "")
			} else {
				if !strings.HasSuffix(text, "y") {
					t.Fatal()
				}
				if len(lines) == 3 {
					if text := lines[1].Box().Children[0].(*bo.TextBox).Text; !strings.HasSuffix(text, hyphenateCharacter) {
						t.Fatalf("unexpected %s", text)
					}
				}
			}

		}
	}
}

func TestHyphenateManual3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Automatic hyphenation opportunities within a word must be ignored if the
	// word contains a conditional hyphen, in favor of the conditional
	// hyphen(s).
	page := renderOnePage(t,
		`<html style="width: 0.1em" lang="en">
        <body style="hyphens: auto">in&shy;lighten&shy;lighten&shy;in`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line1, line2, line3, line4 := unpack4(body)
	tu.AssertEqual(t, line1.Box().Children[0].(*bo.TextBox).Text, "in\u00ad-", "line1")
	tu.AssertEqual(t, line2.Box().Children[0].(*bo.TextBox).Text, "lighten\u00ad-", "line2")
	tu.AssertEqual(t, line3.Box().Children[0].(*bo.TextBox).Text, "lighten\u00ad-", "line3")
	tu.AssertEqual(t, line4.Box().Children[0].(*bo.TextBox).Text, "in", "line4")
}

func TestHyphenateLimitZone1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	page := renderOnePage(t,
		`<html style="width: 12em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto;
        hyphenate-limit-zone: 0" lang=fr>mmmmm hyphénation`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	lines := body.Box().Children
	tu.AssertEqual(t, len(lines), 2, "")

	if text := lines[0].Box().Children[0].(*bo.TextBox).Text; !strings.HasSuffix(text, "-") {
		t.Fatalf("unexpected <%s>", text)
	}
	fullText := ""
	for _, line := range lines {
		fullText += line.Box().Children[0].(*bo.TextBox).Text
	}
	tu.AssertEqual(t, strings.ReplaceAll(fullText, "-", ""), "mmmmm hyphénation", "")
}

func TestHyphenateLimitZone2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	page := renderOnePage(t,
		`<html style="width: 12em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto;
        hyphenate-limit-zone: 9em" lang=fr>mmmmm hyphénation`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	lines := body.Box().Children
	tu.AssertEqual(t, len(lines), 2, "")

	if text := lines[0].Box().Children[0].(*bo.TextBox).Text; !strings.HasSuffix(text, "mm") {
		t.Fatalf("unexpected <%s>", text)
	}
	fullText := ""
	for _, line := range lines {
		fullText += line.Box().Children[0].(*bo.TextBox).Text
	}
	tu.AssertEqual(t, fullText, "mmmmmhyphénation", "")
}

func TestHyphenateLimitZone3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	page := renderOnePage(t,
		`<html style="width: 12em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto;
        hyphenate-limit-zone: 5%" lang=fr>mmmmm hyphénation`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	lines := body.Box().Children
	tu.AssertEqual(t, len(lines), 2, "")

	if text := lines[0].Box().Children[0].(*bo.TextBox).Text; !strings.HasSuffix(text, "-") {
		t.Fatalf("unexpected <%s>", text)
	}
	fullText := ""
	for _, line := range lines {
		fullText += line.Box().Children[0].(*bo.TextBox).Text
	}
	tu.AssertEqual(t, strings.ReplaceAll(fullText, "-", ""), "mmmmm hyphénation", "")
}

func TestHyphenateLimitZone4(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t,
		`<html style="width: 12em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto;
        hyphenate-limit-zone: 95%" lang=fr>mmmmm hyphénation`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	lines := body.Box().Children
	tu.AssertEqual(t, len(lines), 2, "")

	if text := lines[0].Box().Children[0].(*bo.TextBox).Text; !strings.HasSuffix(text, "mm") {
		t.Fatalf("unexpected <%s>", text)
	}
	fullText := ""
	for _, line := range lines {
		fullText += line.Box().Children[0].(*bo.TextBox).Text
	}
	tu.AssertEqual(t, fullText, "mmmmmhyphénation", "")
}

func TestHyphenateLimitChars(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		lines := body.Box().Children
		tu.AssertEqual(t, len(lines), v.result, v.css)
	}
}

func TestHyphenateLimitCharsPunctuation(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		lines := body.Box().Children
		tu.AssertEqual(t, len(lines), 1, "")
	}
}

func TestOverflowWrap(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		var lines []string
		for _, line := range body.Box().Children {
			box := line.Box().Children[0]
			textBox := box.Box().Children[0].(*bo.TextBox)
			lines = append(lines, textBox.Text)
		}
		if !v.test(len(lines)) {
			t.Fatal()
		}
		tu.AssertEqual(t, v.fullText, strings.Join(lines, ""), fmt.Sprintf("input %s %s", v.wrap, v.text))
	}
}

func TestOverflowWrap_2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
		html := page.Box().Children[0]
		body := html.Box().Children[0]

		tableWrapper := body.Box().Children[0]
		table := tableWrapper.Box().Children[0]
		rowGroup := table.Box().Children[0]
		tr := rowGroup.Box().Children[0]
		td := tr.Box().Children[0]
		tu.AssertEqual(t, td.Box().Width, v.expectedWidth, fmt.Sprintf("input %s %s", v.wrap, v.text))
	}
}

func TestWrapOverflowWordBreak(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		var lines []string
		for _, line := range body.Box().Children {
			lineText := ""
			for _, span_box := range line.Box().Children {
				lineText += span_box.Box().Children[0].(*bo.TextBox).Text
			}
			lines = append(lines, lineText)
		}
		tu.AssertEqual(t, lines, data.expectedLines, "")
	}
}

func TestOverflowWrapTrailingSpace(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
		html := page.Box().Children[0]
		body := html.Box().Children[0]

		tableWrapper := body.Box().Children[0]
		table := tableWrapper.Box().Children[0]
		rowGroup := table.Box().Children[0]
		tr := rowGroup.Box().Children[0]
		td := tr.Box().Children[0]
		tu.AssertEqual(t, td.Box().Width, v.expectedWidth, fmt.Sprintf("input %s %s", v.wrap, v.text))
	}
}

func testWhiteSpaceLines(t *testing.T, width int, space string, expected []string) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, fmt.Sprintf(`
      <style>
        body { font-size: 100px; width: %dpx }
        span { white-space: %s }
      </style>
      `, width, space)+"<body><span>This +    \n    is text")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tu.AssertEqual(t, len(body.Box().Children), len(expected), "wrong length")
	for i, line := range body.Box().Children {
		box := line.Box().Children[0]
		text := box.Box().Children[0]
		tu.AssertEqual(t, text.(*bo.TextBox).Text, expected[i], "")
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
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/813
	page := renderOnePage(t, `
      <style>
        pre { width: 0 }
      </style>
      <body><pre>This<br/>is text`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	pre := body.Box().Children[0]
	line1, line2 := pre.Box().Children[0], pre.Box().Children[1]
	text1, box := line1.Box().Children[0], line1.Box().Children[1]
	tu.AssertEqual(t, text1.(*bo.TextBox).Text, "This", "text1")
	tu.AssertEqual(t, box.Box().ElementTag(), "br", "box")
	text2 := line2.Box().Children[0]
	tu.AssertEqual(t, text2.(*bo.TextBox).Text, "is text", "text2")
}

func TestWhiteSpace12(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/813
	page := renderOnePage(t, `
      <style>
        pre { width: 0 }
      </style>
      <body><pre>This is <span>lol</span> text`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	pre := body.Box().Children[0]
	line1 := pre.Box().Children[0]
	text1, span, text2 := unpack3(line1)
	tu.AssertEqual(t, text1.(*bo.TextBox).Text, "This is ", "text1")
	tu.AssertEqual(t, span.Box().ElementTag(), "span", "span")
	tu.AssertEqual(t, text2.(*bo.TextBox).Text, " text", "text2")
}

func TestTabSize(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		paragraph := body.Box().Children[0]
		line := paragraph.Box().Children[0]
		tu.AssertEqual(t, line.Box().Width, v.width, "for value "+v.value)
	}
}

func TestTextTransform(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	expected := []string{
		"Hé Lo1",
		"HÉ LO1",
		"hé lo1",
		"\uff48é\u3000\uff4c\uff2f\uff11",
		"hé lO1",
	}
	if len(body.Box().Children) != len(expected) {
		t.Fatal()
	}
	for i, child := range body.Box().Children {
		line := child.Box().Children[0]
		text := line.Box().Children[0].(*bo.TextBox)
		tu.AssertEqual(t, text.Text, expected[i], "")
	}
}

func TestTextFloatingPreLine(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/610
	_ = renderOnePage(t, `
      <div style="float: left; white-space: pre-line">This is
      oh this end </div>
    `)
}

func TestLeaderContent(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		div := body.Box().Children[0]
		line := div.Box().Children[0]
		after := line.Box().Children[0]
		inline := after.Box().Children[0]
		tu.AssertEqual(t, inline.Box().Children[0].(*bo.TextBox).Text, v.content, "")
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
// 	html := page.Box().Children[0]
// 	body := html.Box().Children[0]
// 	p1, p2 := body.Box().Children[0], body.Box().Children[1]
// 	line1, line2 := p1.Box().Children[0], p1.Box().Children[1]
// 	line3 := p2.Box().Children[0]
// 	text1 := line1.Box().Children[0]
// 	text2 := line2.Box().Children[0]
// 	text3 := line3.Box().Children[0]
// 	tu.AssertEqual(t, text1.(*bo.TextBox).Text, "abcd", "text1")
// 	tu.AssertEqual(t, text2.(*bo.TextBox).Text, "efgh", "text2")
// 	tu.AssertEqual(t, text3.(*bo.TextBox).Text, "ijkl", "text3")
// }

func TestContinue(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	p := body.Box().Children[0]
	line1, line2 := p.Box().Children[0], p.Box().Children[1]
	text1 := line1.Box().Children[0]
	text2 := line2.Box().Children[0]
	tu.AssertEqual(t, text1.(*bo.TextBox).Text, "abcd", "")
	tu.AssertEqual(t, text2.(*bo.TextBox).Text, "efgh", "")
}
