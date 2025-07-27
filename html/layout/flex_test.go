package layout

import (
	"fmt"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

//  Tests for flex layout.

func assertEquals(t *testing.T, values ...pr.MaybeFloat) {
	for _, val := range values {
		if val != values[0] {
			t.Fatalf("different values %f %f", val, values[0])
		}
	}
}

func assertPosXEqual(t *testing.T, boxes ...Box) {
	for _, box := range boxes {
		if box.Box().PositionX != boxes[0].Box().PositionX {
			t.Fatal("different positionX")
		}
	}
}

func assertPosYEqual(t *testing.T, boxes ...Box) {
	for _, box := range boxes {
		if box.Box().PositionY != boxes[0].Box().PositionY {
			t.Fatal("different positionX")
		}
	}
}

func TestFlexDirectionRow(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "A")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "C")
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div1.Box().PositionX < div2.Box().PositionX, true)
	tu.AssertEqual(t, div2.Box().PositionX < div3.Box().PositionX, true)
}

func TestFlexDirectionRowMaxWidth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; max-width: 100px">
        <div></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().Width, 100)
}

func TestFlexDirectionRowMinHeight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; min-height: 100px">
        <div></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().Height, 100)
}

func TestFlexDirectionRowRtl(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; direction: rtl">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "A")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "C")
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div1.Box().PositionX+div1.Box().Width.V(), article.Box().PositionX+article.Box().Width.V())
	tu.AssertEqual(t, div1.Box().PositionX > div2.Box().PositionX, true)
	tu.AssertEqual(t, div2.Box().PositionX > div3.Box().PositionX, true)
}

func TestFlexDirectionRowReverse(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: row-reverse">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "C")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "A")
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div3.Box().PositionX+div3.Box().Width.V(), article.Box().PositionX+article.Box().Width.V())
	tu.AssertEqual(t, div1.Box().PositionX < div2.Box().PositionX, true)
	tu.AssertEqual(t, div2.Box().PositionX < div3.Box().PositionX, true)
}

func TestFlexDirectionRowReverseRtl(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: row-reverse;
      direction: rtl">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "C")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "A")
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div3.Box().PositionX, article.Box().PositionX)
	tu.AssertEqual(t, div1.Box().PositionX > div2.Box().PositionX, true)
	tu.AssertEqual(t, div2.Box().PositionX > div3.Box().PositionX, true)
}

func TestFlexDirectionColumn(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "A")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "C")
	assertPosXEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div1.Box().PositionY, article.Box().PositionY)
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true)
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true)
}

func TestFlexDirectionColumnMinWidth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column; min-height: 100px">
        <div></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().Height, 100)
}

func TestFlexDirectionColumnMaxHeight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; flex-flow: column wrap; max-height: 100px">
        <div style="height: 40px">A</div>
        <div style="height: 40px">B</div>
        <div style="height: 40px">C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().Height, 100)
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Height, 40)
	tu.AssertEqual(t, div1.Box().PositionX, 0)
	tu.AssertEqual(t, div1.Box().PositionY, 0)
	tu.AssertEqual(t, div2.Box().Height, 40)
	tu.AssertEqual(t, div2.Box().PositionX, 0)
	tu.AssertEqual(t, div2.Box().PositionY, 40)
	tu.AssertEqual(t, div3.Box().Height, 40)
	tu.AssertEqual(t, div3.Box().PositionX, div1.Box().Width)
	tu.AssertEqual(t, div3.Box().PositionY, 0)
}

func TestFlexDirectionColumnRtl(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column;
      direction: rtl">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "A")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "C")
	assertPosXEqual(t, div1, div2, div3, article)

	tu.AssertEqual(t, div1.Box().PositionY, article.Box().PositionY)
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true)
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true)
}

func TestFlexDirectionColumnReverse(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column-reverse">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "C")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "A")
	assertPosXEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div3.Box().PositionY+div3.Box().Height.V(), article.Box().PositionY+article.Box().Height.V())
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true)
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true)
}

func TestFlexDirectionColumnReverseRtl(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column-reverse;
      direction: rtl">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "C")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "A")
	assertPosXEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div3.Box().PositionY+div3.Box().Height.V(), article.Box().PositionY+article.Box().Height.V())
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true)
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true)
}

func TestFlexDirectionColumnBoxSizing(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        article {
          box-sizing: border-box;
          display: flex;
          flex-direction: column;
          height: 10px;
          padding-top: 5px;
          width: 10px;
        }
      </style>
      <article></article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().Width, 10)
	tu.AssertEqual(t, article.Box().Height, 5)
	tu.AssertEqual(t, article.Box().PaddingTop, 5)
}

func TestFlexRowWrap(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-flow: wrap; width: 50px">
        <div style="width: 20px">A</div>
        <div style="width: 20px">B</div>
        <div style="width: 20px">C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "A")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "C")
	tu.AssertEqual(t, div1.Box().PositionY, div2.Box().PositionY)
	tu.AssertEqual(t, div2.Box().PositionY, article.Box().PositionY)
	tu.AssertEqual(t, div3.Box().PositionY, article.Box().PositionY+div2.Box().Height.V())
	tu.AssertEqual(t, div1.Box().PositionX, div3.Box().PositionX)
	tu.AssertEqual(t, div3.Box().PositionX, article.Box().PositionX)
	tu.AssertEqual(t, div1.Box().PositionX < div2.Box().PositionX, true)
}

func TestFlexColumnWrap(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-flow: column wrap; height: 50px">
        <div style="height: 20px">A</div>
        <div style="height: 20px">B</div>
        <div style="height: 20px">C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "A")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "C")
	assertPosXEqual(t, div1, div2, article)
	tu.AssertEqual(t, div3.Box().PositionX, article.Box().PositionX+div2.Box().Width.V())
	assertPosYEqual(t, div1, div3, article)
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true)
}

func TestFlexRowWrapReverse(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-flow: wrap-reverse; width: 50px">
        <div style="width: 20px">A</div>
        <div style="width: 20px">B</div>
        <div style="width: 20px">C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "C")
	assertText(t, unpack1(div2.Box().Children[0]), "A")
	assertText(t, unpack1(div3.Box().Children[0]), "B")
	assertPosYEqual(t, div1, article)
	assertPosYEqual(t, div2, div3)
	tu.AssertEqual(t, div3.Box().PositionY, article.Box().PositionY+div1.Box().Height.V())
	assertPosXEqual(t, div1, div2, article)
	tu.AssertEqual(t, div2.Box().PositionX < div3.Box().PositionX, true)
}

func TestFlexColumnWrapReverse(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-flow: column wrap-reverse;
                      height: 50px">
        <div style="height: 20px">A</div>
        <div style="height: 20px">B</div>
        <div style="height: 20px">C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "C")
	assertText(t, unpack1(div2.Box().Children[0]), "A")
	assertText(t, unpack1(div3.Box().Children[0]), "B")
	assertPosXEqual(t, div1, article)
	assertPosXEqual(t, div2, div3)
	tu.AssertEqual(t, div3.Box().PositionX, article.Box().PositionX+div1.Box().Width.V())
	assertPosYEqual(t, div1, div2, article)
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true)
}

func TestFlexDirectionColumnFixedHeightContainer(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <section style="height: 10px">
        <article style="display: flex; flex-direction: column">
          <div>A</div>
          <div>B</div>
          <div>C</div>
        </article>
      </section>
    `)
	html := unpack1(page)
	body := unpack1(html)
	section := unpack1(body)
	article := unpack1(section)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "A")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "C")
	assertPosXEqual(t, div1, div2, div3, article)
	assertPosYEqual(t, div1, article)
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true)
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true)
	tu.AssertEqual(t, section.Box().Height, Fl(10))
	tu.AssertEqual(t, article.Box().Height.V() > 10, true)
}

func TestFlexDirectionColumnFixedHeight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column; height: 10px">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "A")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "C")
	assertPosXEqual(t, div1, div2, div3, article)
	assertPosYEqual(t, div1, article)
	tu.Assert(t, div1.Box().PositionY < div2.Box().PositionY && div2.Box().PositionY < div3.Box().PositionY)
	tu.AssertEqual(t, article.Box().Height, Fl(10))
	tu.Assert(t, div3.Box().PositionY > 10)
}

func TestFlexDirectionColumnFixedHeightWrap(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column; height: 10px;
                      flex-wrap: wrap">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "A")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "C")
	tu.AssertEqual(t, div1.Box().PositionX != div2.Box().PositionX, true)
	tu.AssertEqual(t, div2.Box().PositionX != div3.Box().PositionX, true)
	assertPosYEqual(t, div1, article)
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, article.Box().Height, Fl(10))
}

func TestFlexDirectionColumnBreak(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #2066.
	page1, page2 := renderTwoPages(t, `
      <style>
        @page { size: 4px 5px }
      </style>
      <article style="display: flex; flex-direction: column; font: 2px weasyprint">
        <div>A<br>B<br>C</div>
      </article>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	assertText(t, div.Box().Children[0].Box().Children[0], "A")
	assertText(t, div.Box().Children[1].Box().Children[0], "B")
	tu.AssertEqual(t, div.Box().Height, 5)
	html = unpack1(page2)
	body = unpack1(html)
	article = unpack1(body)
	div = unpack1(article)
	assertText(t, div.Box().Children[0].Box().Children[0], "C")
	tu.AssertEqual(t, div.Box().Height, 2)
}

func TestFlexDirectionColumnBreakMargin(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #1967.
	page1, page2 := renderTwoPages(t, `
      <style>
        @page { size: 4px 7px }
      </style>
      <body style="font: 2px weasyprint">
        <p style="margin: 1px 0">a</p>
        <article style="display: flex; flex-direction: column">
          <div>A<br>B<br>C</div>
        </article>
      </body>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	p, article := unpack2(body)
	tu.AssertEqual(t, p.Box().MarginHeight(), 4)
	tu.AssertEqual(t, article.Box().PositionY, 4)
	div := unpack1(article)
	assertText(t, div.Box().Children[0].Box().Children[0], "A")
	tu.AssertEqual(t, div.Box().Height, 3)
	html = unpack1(page2)
	body = unpack1(html)
	article = unpack1(body)
	div = unpack1(article)
	assertText(t, div.Box().Children[0].Box().Children[0], "B")
	assertText(t, div.Box().Children[1].Box().Children[0], "C")
	tu.AssertEqual(t, div.Box().Height, 4)
}

func TestFlexDirectionColumnBreakBorder(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page1, page2 := renderTwoPages(t, `
      <style>
        @page { size: 8px 7px }
        article, div { border: 1px solid black }
      </style>
      <article style="display: flex; flex-direction: column; font: 2px weasyprint">
        <div>A B C</div>
      </article>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().BorderHeight(), 7)
	tu.AssertEqual(t, article.Box().BorderTopWidth, 1)
	tu.AssertEqual(t, article.Box().BorderBottomWidth, 0)
	div := unpack1(article)
	assertText(t, div.Box().Children[0].Box().Children[0], "A")
	assertText(t, div.Box().Children[1].Box().Children[0], "B")
	tu.AssertEqual(t, div.Box().BorderHeight(), 6)
	tu.AssertEqual(t, div.Box().BorderTopWidth, 1)
	tu.AssertEqual(t, div.Box().BorderBottomWidth, 0)
	html = unpack1(page2)
	body = unpack1(html)
	article = unpack1(body)
	tu.AssertEqual(t, article.Box().BorderHeight(), 4)
	tu.AssertEqual(t, article.Box().BorderTopWidth, 0)
	tu.AssertEqual(t, article.Box().BorderBottomWidth, 1)
	div = unpack1(article)
	assertText(t, div.Box().Children[0].Box().Children[0], "C")
	tu.AssertEqual(t, div.Box().BorderHeight(), 3)
	tu.AssertEqual(t, div.Box().BorderTopWidth, 0)
	tu.AssertEqual(t, div.Box().BorderBottomWidth, 1)
}

func TestFlexDirectionColumnBreakMultipleChildren(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page1, page2 := renderTwoPages(t, `
      <style>
        @page { size: 4px 5px }
      </style>
      <article style="display: flex; flex-direction: column; font: 2px weasyprint">
        <div>A</div>
        <div>B<br>C</div>
      </article>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2 := unpack2(article)
	assertText(t, div1.Box().Children[0].Box().Children[0], "A")
	tu.AssertEqual(t, div1.Box().Height, Fl(2))
	assertText(t, div2.Box().Children[0].Box().Children[0], "B")
	tu.AssertEqual(t, div2.Box().Height, Fl(3))
	html = unpack1(page2)
	body = unpack1(html)
	article = unpack1(body)
	div2 = unpack1(article)
	assertText(t, div2.Box().Children[0].Box().Children[0], "C")
	tu.AssertEqual(t, div2.Box().Height, Fl(2))
}

func TestFlexItemMinWidth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex">
        <div style="min-width: 30px">A</div>
        <div style="min-width: 50px">B</div>
        <div style="min-width: 5px">C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "A")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "C")
	tu.AssertEqual(t, div1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div1.Box().Width, Fl(30))
	tu.AssertEqual(t, div2.Box().PositionX, Fl(30))
	tu.AssertEqual(t, div2.Box().Width, Fl(50))
	tu.AssertEqual(t, div3.Box().PositionX, Fl(80))
	tu.AssertEqual(t, div3.Box().Width.V() > Fl(5), true)
	assertPosYEqual(t, div1, div2, div3, article)
}

func TestFlexItemMinHeight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex">
        <div style="min-height: 30px">A</div>
        <div style="min-height: 50px">B</div>
        <div style="min-height: 5px">C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, unpack1(div1.Box().Children[0]), "A")
	assertText(t, unpack1(div2.Box().Children[0]), "B")
	assertText(t, unpack1(div3.Box().Children[0]), "C")
	tu.AssertEqual(t, div1.Box().Height.V(), div2.Box().Height.V())
	tu.AssertEqual(t, div2.Box().Height.V(), div3.Box().Height.V())
	tu.AssertEqual(t, div3.Box().Height.V(), article.Box().Height.V())
	tu.AssertEqual(t, article.Box().Height.V(), Fl(50))
}

func TestFlexAutoMargin(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/800
	_ = renderOnePage(t, `<div style="display: flex; margin: auto">`)
	_ = renderOnePage(t, `<div style="display: flex; flex-direction: column; margin: auto">`)
}

func TestFlexItemAutoMarginSized(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #2054.
	page := renderOnePage(t, `
      <style>
        div {
          margin: auto;
          display: flex;
          width: 160px;
          height: 160px;
        }
      </style>
      <article>
        <div></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	tu.Assert(t, div.Box().MarginLeft != Fl(0))
}

func TestFlexNoBaseline(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/765
	_ = renderOnePage(t, `
      <div class="references" style="display: flex; align-items: baseline;">
        <div></div>
      </div>`)
}

func TestFlexAlignContent(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range []struct {
		align  string
		height int
		y1, y2 Fl
	}{
		{"flex-start", 50, 0, 10},
		{"flex-end", 50, 30, 40},
		{"space-around", 60, 10, 40},
		{"space-between", 50, 0, 40},
		{"space-evenly", 50, 10, 30},
	} {
		// Regression test for https://github.com/Kozea/WeasyPrint/issues/811
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          align-content: %s;
          display: flex;
          flex-wrap: wrap;
          font-family: weasyprint;
          font-size: 10px;
          height: %dpx;
          line-height: 1;
        }
        section {
          width: 100%%;
        }
      </style>
      <article>
        <section><span>Lorem</span></section>
        <section><span>Lorem</span></section>
      </article>
    `, data.align, data.height))
		html := unpack1(page)
		body := unpack1(html)
		article := unpack1(body)
		section1, section2 := unpack2(article)
		line1 := unpack1(section1)
		line2 := unpack1(section2)
		span1 := unpack1(line1)
		span2 := unpack1(line2)
		tu.AssertEqual(t, section1.Box().PositionX, span1.Box().PositionX)
		tu.AssertEqual(t, span1.Box().PositionX, Fl(0))
		tu.AssertEqual(t, section1.Box().PositionY, span1.Box().PositionY)
		tu.AssertEqual(t, span1.Box().PositionY, data.y1)
		tu.AssertEqual(t, section2.Box().PositionX, span2.Box().PositionX)
		tu.AssertEqual(t, span2.Box().PositionX, Fl(0))
		tu.AssertEqual(t, section2.Box().PositionY, span2.Box().PositionY)
		tu.AssertEqual(t, span2.Box().PositionY, data.y2)
	}
}

func TestFlexItemPercentage(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/885
	page := renderOnePage(t, `
      <div style="display: flex; font-size: 15px; line-height: 1">
        <div style="height: 100%">a</div>
      </div>`)
	html := unpack1(page)
	body := unpack1(html)
	flex := unpack1(body)
	flexItem := unpack1(flex)
	tu.AssertEqual(t, flexItem.Box().Height, Fl(15))
}

func TestFlexUndefinedPercentageHeightMultipleLines(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1204
	_ = renderOnePage(t, `
      <div style="display: flex; flex-wrap: wrap; height: 100%">
        <div style="width: 100%">a</div>
        <div style="width: 100%">b</div>
      </div>`)
}

func TestFlexAbsolute(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1536
	_ = renderOnePage(t, `
      <div style="display: flex; position: absolute">
        <div>a</div>
      </div>`)
}

func TestFlexPercentHeight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
	  <style>
        .a { height: 10px; width: 10px; }
        .b { height: 10%; width: 100%; display: flex; flex-direction: column; }
      </style>
      <div class="a"">
        <div class="b"></div>
      </div>`)
	html := unpack1(page)
	body := unpack1(html)
	a := unpack1(body)
	b := unpack1(a)
	tu.AssertEqual(t, a.Box().Height, Fl(10))
	tu.AssertEqual(t, b.Box().Height, Fl(1))
}

func TestFlexColumnHeight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #2222.
	page := renderOnePage(t, `
      <section style="display: flex; width: 10em">
        <article style="display: flex; flex-direction: column">
          <div>
            Lorem ipsum dolor sit amet
          </div>
        </article>
        <article style="display: flex; flex-direction: column">
          <div>
            Lorem ipsum dolor sit amet
          </div>
        </article>
      </section>
    `)
	html := unpack1(page)
	body := unpack1(html)
	section := unpack1(body)
	article1, article2 := unpack2(section)
	tu.AssertEqual(t, article1.Box().Height, section.Box().Height)
	tu.AssertEqual(t, article2.Box().Height, section.Box().Height)
}

func TestFlexColumnHeightMargin(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #2222.
	page := renderOnePage(t, `
      <section style="display: flex; flex-direction: column; width: 10em">
        <article style="margin: 5px">
          Lorem ipsum dolor sit amet
        </article>
        <article style="margin: 10px">
          Lorem ipsum dolor sit amet
        </article>
      </section>
    `)
	html := unpack1(page)
	body := unpack1(html)
	section := unpack1(body)
	article1, article2 := unpack2(section)
	tu.AssertEqual(t, section.Box().Height, article1.Box().MarginHeight()+article2.Box().MarginHeight())
}

func TestFlexColumnWidth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #1171.
	page := renderOnePage(t, `
      <main style="display: flex; flex-direction: column;
                   width: 40px; height: 50px; font: 2px weasyprint">
        <section style="width: 100%; height: 5px">a</section>
        <section style="display: flex; flex: auto; flex-direction: column;
                        justify-content: space-between; width: 100%">
          <div>b</div>
          <div>c</div>
        </section>
      </main>
    `)
	html := unpack1(page)
	body := unpack1(html)
	main := unpack1(body)
	section1, section2 := unpack2(main)
	div1, div2 := unpack2(section2)
	tu.Assert(t, section1.Box().Width == section2.Box().Width && section2.Box().Width == main.Box().Width)
	tu.AssertEqual(t, div1.Box().Width, div2.Box().Width)
	tu.AssertEqual(t, div1.Box().PositionY, 5)
	tu.AssertEqual(t, div2.Box().PositionY, 48)
}

func TestFlexColumnInFlexRow(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <body style="display: flex; flex-wrap: wrap; font: 2px weasyprint">
        <article>1</article>
        <section style="display: flex; flex-direction: column">
          <div>2</div>
        </section>
      </body>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article, section := unpack2(body)
	tu.Assert(t, article.Box().PositionY == 0 && section.Box().PositionY == 0)
	tu.AssertEqual(t, article.Box().PositionX, Fl(0))
	tu.AssertEqual(t, section.Box().PositionX, Fl(2))
}

func TestFlexOverflow(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #2292.
	page := renderOnePage(t, `
      <style>
        article {
          display: flex;
        }
        section {
          overflow: hidden;
          width: 5px;
        }
      </style>
      <article>
        <section>A</section>
        <section>B</section>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	section_1, section_2 := unpack2(article)
	tu.AssertEqual(t, section_1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, section_2.Box().PositionX, Fl(5))
}

func TestFlexColumnOverflow(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #2304.
	renderPages(t, `
      <style>
        @page {size: 20px}
      </style>
      <div style="display: flex; flex-direction: column">
        <div></div>
        <div><div style="height: 40px"></div></div>
        <div><div style="height: 5px"></div></div>
      </div>
    `)
}

func TestInlineFlex(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	renderPages(t, `
      <style>
        @page {size: 20px}
      </style>
      <div style="display: inline-flex; flex-direction: column">
        <div>test</div>
      </div>
    `)
}

func TestInlineFlexEmptyChild(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	renderPages(t, `
      <style>
        @page {size: 20px}
      </style>
      <div style="display: inline-flex; flex-direction: column">
        <div></div>
      </div>
    `)
}

func TestInlineFlexAbsoluteBaseline(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	renderPages(t, `
      <style>
        @page {size: 20px}
      </style>
      <div style="display: inline-flex; flex-direction: column">
        <div style="position: absolute">abs</div>
      </div>
    `)
}

func TestFlexItemOverflow(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #2359.
	page := renderOnePage(t, `
      <div style="display: flex; font: 2px weasyprint; width: 12px">
        <div>ab</div>
        <div>c d e</div>
        <div>f</div>
      </div>`)
	html := unpack1(page)
	body := unpack1(html)
	flex := unpack1(body)
	div1, div2, div3 := unpack3(flex)
	tu.AssertEqual(t, div1.Box().Width, Fl(4))
	tu.AssertEqual(t, div2.Box().Width, Fl(6))
	tu.AssertEqual(t, div3.Box().Width, Fl(2))
	line1, line2 := unpack2(div2)
	text1 := unpack1(line1)
	text2 := unpack1(line2)
	assertText(t, text1, "c d")
	assertText(t, text2, "e")
}

func TestFlexItemChildBottomMargin(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #2449.
	for _, direction := range [...]string{"row", "column"} {
		page := renderOnePage(t, fmt.Sprintf(`
		  <div style="display: flex; font: 2px weasyprint; flex-direction: %s">
			<section>
			  <div style="margin: 2px 0">ab</div>
			</section>
		  </div>`, direction))
		html := unpack1(page)
		body := unpack1(html)
		flex := unpack1(body)
		tu.AssertEqual(t, flex.Box().ContentBoxY(), Fl(0))
		tu.AssertEqual(t, flex.Box().Height, Fl(6))
		section := unpack1(flex)
		tu.AssertEqual(t, section.Box().ContentBoxY(), Fl(0))
		tu.AssertEqual(t, section.Box().Height, Fl(6))
		div := unpack1(section)
		tu.AssertEqual(t, div.Box().ContentBoxY(), Fl(2))
		tu.AssertEqual(t, div.Box().Height, Fl(2))
	}
}

func TestFlexDirectionRowInlineBlock(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #1652.
	page := renderOnePage(t, `
      <article style="display: flex; font: 2px weasyprint; width: 14px">
        <div style="display: inline-block">A B C D E F</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	tu.AssertEqual(t, div.Box().Width, Fl(14))
	assertText(t, div.Box().Children[0].Box().Children[0], "A B C D")
	assertText(t, div.Box().Children[1].Box().Children[0], "E F")
}

func TestFlexFloat(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex">
        <div style="float: left">A</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	assertText(t, div.Box().Children[0].Box().Children[0], "A")
}

func TestFlexFloatInFlexItem(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #1356.
	page := renderOnePage(t, `
      <article style="display: flex; font: 2px weasyprint">
        <div style="width: 10px"><span style="float: right">abc</span></div>
        <div style="width: 10px"><span style="float: right">abc</span></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2 := unpack2(article)
	span1 := unpack1(div1)
	tu.AssertEqual(t, span1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, span1.Box().PositionX+span1.Box().Width.V(), Fl(10))
	span2 := unpack1(div2)
	tu.AssertEqual(t, span2.Box().PositionY, Fl(0))
	tu.AssertEqual(t, span2.Box().PositionX+span2.Box().Width.V(), Fl(20))
}

func TestFlexDirectionRowDefinedMain(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex">
        <div style="width: 10px; padding: 1px"></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	tu.AssertEqual(t, div.Box().Width, Fl(10))
	tu.AssertEqual(t, div.Box().MarginWidth(), Fl(12))
}

func TestFlexDirectionRowDefinedMainBorderBox(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex">
        <div style="box-sizing: border-box; width: 10px; padding: 1px"></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	tu.AssertEqual(t, div.Box().Width, Fl(8))
	tu.AssertEqual(t, div.Box().MarginWidth(), Fl(10))
}

func TestFlexDirectionColumnDefinedMain(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column">
        <div style="height: 10px; padding: 1px"></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	tu.AssertEqual(t, div.Box().Height, 10)
	tu.AssertEqual(t, div.Box().MarginHeight(), Fl(12))
}

func TestFlexDirectionColumnDefinedMainBorderBox(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column">
        <div style="box-sizing: border-box; height: 10px; padding: 1px"></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	tu.AssertEqual(t, div.Box().Height, 8)
	tu.AssertEqual(t, div.Box().MarginHeight(), Fl(10))
}

func TestFlexItemNegativeMargin(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex">
        <div style="margin-left: -20px; height: 10px; width: 10px"></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	tu.AssertEqual(t, div.Box().Height, 10)
	tu.AssertEqual(t, div.Box().Width, Fl(10))
}

func TestFlexItemAutoMarginMain(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; width: 100px">
        <div style="margin-left: auto; height: 10px; width: 10px"></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	tu.AssertEqual(t, div.Box().Height, 10)
	tu.AssertEqual(t, div.Box().Width, Fl(10))
	tu.AssertEqual(t, div.Box().MarginLeft, Fl(90))
}

func TestFlexItemAutoMarginCross(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; height: 100px">
        <div style="margin-top: auto; height: 10px; width: 10px"></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	tu.AssertEqual(t, div.Box().Height, 10)
	tu.AssertEqual(t, div.Box().Width, Fl(10))
	tu.AssertEqual(t, div.Box().MarginTop, Fl(90))
}

func TestFlexDirectionColumnItemAutoMargin(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <div style="font: 2px weasyprint; width: 30px; display: flex;
                  flex-direction: column; align-items: flex-start">
          <article style="margin: 0 auto">XXXX</article>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	article := unpack1(div)
	tu.AssertEqual(t, article.Box().Width, 8)
	tu.AssertEqual(t, article.Box().MarginLeft, Fl(11))
}

func TestFlexItemAutoMarginFlexBasis(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex">
        <div style="margin-left: auto; height: 10px; flex-basis: 10px"></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div := unpack1(article)
	tu.AssertEqual(t, div.Box().Height, 10)
	tu.AssertEqual(t, div.Box().Width, Fl(10))
}

// @pytest.mark.parametrize("align, x1, x2, x3", (
//
//		("start", 0, 2, 4),
//	    ("flex-start", 0, 2, 4),
//	    ("left", 0, 2, 4),
//	    ("end", 6, 8, 10),
//	    ("flex-end", 6, 8, 10),
//	    ("right", 6, 8, 10),
//	    ("center", 3, 5, 7),
//	    ("space-between", 0, 5, 10),
//	    ("space-around", 1, 5, 9),
//	    ("space-evenly", 1.5, 5, 8.5),
//	}))
func TestFlexDirectionRowJustify(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range [...]struct {
		align      string
		x1, x2, x3 Fl
	}{
		{"start", 0, 2, 4},
		{"flex-start", 0, 2, 4},
		{"left", 0, 2, 4},
		{"end", 6, 8, 10},
		{"flex-end", 6, 8, 10},
		{"right", 6, 8, 10},
		{"center", 3, 5, 7},
		{"space-between", 0, 5, 10},
		{"space-around", 1, 5, 9},
		{"space-evenly", 1.5, 5, 8.5},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
		  <article style="width: 12px; font: 2px weasyprint;
						  display: flex; justify-content: %s">
			<div>A</div>
			<div>B</div>
			<div>C</div>
		  </article>
		`, data.align))
		html := unpack1(page)
		body := unpack1(html)
		article := unpack1(body)
		div1, div2, div3 := unpack3(article)
		assertText(t, div1.Box().Children[0].Box().Children[0], "A")
		assertText(t, div2.Box().Children[0].Box().Children[0], "B")
		assertText(t, div3.Box().Children[0].Box().Children[0], "C")
		assertPosYEqual(t, div1, div2, div3, article)
		tu.AssertEqual(t, article.Box().PositionX, Fl(0))
		tu.AssertEqual(t, div1.Box().PositionX, data.x1)
		tu.AssertEqual(t, div2.Box().PositionX, data.x2)
		tu.AssertEqual(t, div3.Box().PositionX, data.x3)

	}
}

func TestFlexDirectionColumnJustify(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	for _, data := range [...]struct {
		align      string
		y1, y2, y3 Fl
	}{
		{"start", 0, 2, 4},
		{"flex-start", 0, 2, 4},
		{"left", 0, 2, 4},
		{"end", 6, 8, 10},
		{"flex-end", 6, 8, 10},
		{"right", 6, 8, 10},
		{"center", 3, 5, 7},
		{"space-between", 0, 5, 10},
		{"space-around", 1, 5, 9},
		{"space-evenly", 1.5, 5, 8.5},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <article style="height: 12px; font: 2px weasyprint;
                      display: flex; flex-direction: column; justify-content: %s">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `, data.align))
		html := unpack1(page)
		body := unpack1(html)
		article := unpack1(body)
		div1, div2, div3 := unpack3(article)
		assertText(t, div1.Box().Children[0].Box().Children[0], "A")
		assertText(t, div2.Box().Children[0].Box().Children[0], "B")
		assertText(t, div3.Box().Children[0].Box().Children[0], "C")
		assertPosXEqual(t, div1, div2, div3, article)
		tu.AssertEqual(t, article.Box().PositionY, Fl(0))
		tu.AssertEqual(t, div1.Box().PositionY, data.y1)
		tu.AssertEqual(t, div2.Box().PositionY, data.y2)
		tu.AssertEqual(t, div3.Box().PositionY, data.y3)
	}
}

func TestFlexDirectionRowJustifyGap(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range [...]struct {
		align      string
		x1, x2, x3 Fl
	}{
		{"start", 0, 4, 8},
		{"flex-start", 0, 4, 8},
		{"left", 0, 4, 8},
		{"end", 6, 10, 14},
		{"flex-end", 6, 10, 14},
		{"right", 6, 10, 14},
		{"center", 3, 7, 11},
		{"space-between", 0, 7, 14},
		{"space-around", 1, 7, 13},
		{"space-evenly", 1.5, 7, 12.5},
	} {

		page := renderOnePage(t, fmt.Sprintf(`
      <article style="width: 16px; font: 2px weasyprint; gap: 2px;
                      display: flex; justify-content: %s">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `, data.align))
		html := unpack1(page)
		body := unpack1(html)
		article := unpack1(body)
		div1, div2, div3 := unpack3(article)
		assertText(t, div1.Box().Children[0].Box().Children[0], "A")
		assertText(t, div2.Box().Children[0].Box().Children[0], "B")
		assertText(t, div3.Box().Children[0].Box().Children[0], "C")
		assertPosYEqual(t, div1, div2, div3, article)
		tu.AssertEqual(t, article.Box().PositionX, Fl(0))
		tu.AssertEqual(t, div1.Box().PositionX, data.x1)
		tu.AssertEqual(t, div2.Box().PositionX, data.x2)
		tu.AssertEqual(t, div3.Box().PositionX, data.x3)
	}
}

func TestFlexDirectionColumnJustifyGap(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range [...]struct {
		align      string
		y1, y2, y3 Fl
	}{
		{"start", 0, 4, 8},
		{"flex-start", 0, 4, 8},
		{"left", 0, 4, 8},
		{"end", 6, 10, 14},
		{"flex-end", 6, 10, 14},
		{"right", 6, 10, 14},
		{"center", 3, 7, 11},
		{"space-between", 0, 7, 14},
		{"space-around", 1, 7, 13},
		{"space-evenly", 1.5, 7, 12.5},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <article style="height: 16px; font: 2px weasyprint; gap: 2px;
                      display: flex; flex-direction: column; justify-content: %s">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `, data.align))
		html := unpack1(page)
		body := unpack1(html)
		article := unpack1(body)
		div1, div2, div3 := unpack3(article)
		assertText(t, div1.Box().Children[0].Box().Children[0], "A")
		assertText(t, div2.Box().Children[0].Box().Children[0], "B")
		assertText(t, div3.Box().Children[0].Box().Children[0], "C")
		assertPosXEqual(t, div1, div2, div3, article)
		tu.AssertEqual(t, article.Box().PositionY, Fl(0))
		tu.AssertEqual(t, div1.Box().PositionY, data.y1)
		tu.AssertEqual(t, div2.Box().PositionY, data.y2)
		tu.AssertEqual(t, div3.Box().PositionY, data.y3)
	}
}

func TestFlexDirectionRowJustifyGapWrap(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range [...]struct {
		align      string
		x1, x2, x3 Fl
	}{
		{"start", 0, 4, 0},
		{"flex-start", 0, 4, 0},
		{"left", 0, 4, 0},
		{"end", 3, 7, 7},
		{"flex-end", 3, 7, 7},
		{"right", 3, 7, 7},
		{"center", 1.5, 5.5, 3.5},
		{"space-between", 0, 7, 0},
		{"space-around", 0.75, 6.25, 3.5},
		{"space-evenly", 1, 6, 3.5},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <article style="width: 9px; font: 2px weasyprint; gap: 1px 2px;
                      display: flex; flex-wrap: wrap; justify-content: %s">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `, data.align))
		html := unpack1(page)
		body := unpack1(html)
		article := unpack1(body)
		div1, div2, div3 := unpack3(article)
		assertText(t, div1.Box().Children[0].Box().Children[0], "A")
		assertText(t, div2.Box().Children[0].Box().Children[0], "B")
		assertText(t, div3.Box().Children[0].Box().Children[0], "C")
		assertPosYEqual(t, div1, div2, article)
		tu.AssertEqual(t, article.Box().PositionY, Fl(0))
		tu.AssertEqual(t, div3.Box().PositionY, Fl(3))
		tu.AssertEqual(t, article.Box().PositionX, Fl(0))
		tu.AssertEqual(t, div1.Box().PositionX, data.x1)
		tu.AssertEqual(t, div2.Box().PositionX, data.x2)
		tu.AssertEqual(t, div3.Box().PositionX, data.x3)
	}
}

func TestFlexDirectionColumnJustifyGapWrap(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range [...]struct {
		align      string
		y1, y2, y3 Fl
	}{
		{"start", 0, 4, 0},
		{"flex-start", 0, 4, 0},
		{"left", 0, 4, 0},
		{"end", 3, 7, 7},
		{"flex-end", 3, 7, 7},
		{"right", 3, 7, 7},
		{"center", 1.5, 5.5, 3.5},
		{"space-between", 0, 7, 0},
		{"space-around", 0.75, 6.25, 3.5},
		{"space-evenly", 1, 6, 3.5},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <article style="height: 9px; width: 9px; font: 2px weasyprint; gap: 2px 1px;
                      display: flex; flex-wrap: wrap; flex-direction: column;
                      justify-content: %s">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `, data.align))
		html := unpack1(page)
		body := unpack1(html)
		article := unpack1(body)
		div1, div2, div3 := unpack3(article)
		assertText(t, div1.Box().Children[0].Box().Children[0], "A")
		assertText(t, div2.Box().Children[0].Box().Children[0], "B")
		assertText(t, div3.Box().Children[0].Box().Children[0], "C")

		assertPosXEqual(t, div1, div2, article)
		tu.AssertEqual(t, article.Box().PositionX, Fl(0))
		tu.AssertEqual(t, div3.Box().PositionX, Fl(5))
		tu.AssertEqual(t, article.Box().PositionY, Fl(0))
		tu.AssertEqual(t, div1.Box().PositionY, data.y1)
		tu.AssertEqual(t, div2.Box().PositionY, data.y2)
		tu.AssertEqual(t, div3.Box().PositionY, data.y3)
	}
}

func TestFlexDirectionRowStretchNoGrow(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="font: 2px weasyprint; width: 10px;
                      display: flex; justify-content: stretch">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, div1.Box().Children[0].Box().Children[0], "A")
	assertText(t, div2.Box().Children[0].Box().Children[0], "B")
	assertText(t, div3.Box().Children[0].Box().Children[0], "C")
	assertPosYEqual(t, div1, div2, div3, article)
	tu.Assert(t, div1.Box().Width == div2.Box().Width && div2.Box().Width == div3.Box().Width && div3.Box().Width == Fl(2))
	tu.Assert(t, div1.Box().PositionX == article.Box().PositionX && article.Box().PositionX == 0)
	tu.AssertEqual(t, div2.Box().PositionX, Fl(2))
	tu.AssertEqual(t, div3.Box().PositionX, Fl(4))
}

func TestFlexDirectionRowStretchGrow(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="font: 2px weasyprint; width: 10px;
                      display: flex; justify-content: stretch">
        <div>A</div>
        <div style="flex-grow: 3">B</div>
        <div style="flex-grow: 1">C</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2, div3 := unpack3(article)
	assertText(t, div1.Box().Children[0].Box().Children[0], "A")
	assertText(t, div2.Box().Children[0].Box().Children[0], "B")
	assertText(t, div3.Box().Children[0].Box().Children[0], "C")
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div1.Box().Width, Fl(2))
	tu.AssertEqual(t, div2.Box().Width, Fl(5))
	tu.AssertEqual(t, div3.Box().Width, Fl(3))
	tu.Assert(t, div1.Box().PositionX == article.Box().PositionX && article.Box().PositionX == 0)
	tu.AssertEqual(t, div2.Box().PositionX, Fl(2))
	tu.AssertEqual(t, div3.Box().PositionX, Fl(7))
}

func TestFlexDirectionRowJustifyMarginPadding(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range [...]struct {
		align      string
		x1, x2, x3 Fl
	}{
		{"start", 0, 6, 12},
		{"flex-start", 0, 6, 12},
		{"left", 0, 6, 12},
		{"end", 6, 12, 18},
		{"flex-end", 6, 12, 18},
		{"right", 6, 12, 18},
		{"center", 3, 9, 15},
		{"space-between", 0, 9, 18},
		{"space-around", 1, 9, 17},
		{"space-evenly", 1.5, 9, 16.5},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <article style="width: 20px; font: 2px weasyprint;
                      display: flex; justify-content: %s">
        <div style="margin: 0 1em">A</div>
        <div style="padding: 0 1em">B</div>
        <div>C</div>
      </article>
    `, data.align))
		html := unpack1(page)
		body := unpack1(html)
		article := unpack1(body)
		div1, div2, div3 := unpack3(article)
		assertText(t, div1.Box().Children[0].Box().Children[0], "A")
		assertText(t, div2.Box().Children[0].Box().Children[0], "B")
		assertText(t, div3.Box().Children[0].Box().Children[0], "C")
		assertPosYEqual(t, div1, div2, div3, article)
		tu.AssertEqual(t, article.Box().PositionX, 0)
		tu.AssertEqual(t, article.Box().Width, 20)
		tu.AssertEqual(t, div1.Box().PositionX, data.x1)
		tu.AssertEqual(t, div2.Box().PositionX, data.x2)
		tu.AssertEqual(t, div3.Box().PositionX, data.x3)
		tu.AssertEqual(t, div1.Box().MarginWidth(), 6)
		tu.AssertEqual(t, div2.Box().MarginWidth(), 6)
		tu.AssertEqual(t, div3.Box().MarginWidth(), 2)
	}
}

func TestFlexDirectionColumnJustifyMarginPadding(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range [...]struct {
		align      string
		y1, y2, y3 Fl
	}{
		{"start", 0, 6, 12},
		{"flex-start", 0, 6, 12},
		{"left", 0, 6, 12},
		{"end", 6, 12, 18},
		{"flex-end", 6, 12, 18},
		{"right", 6, 12, 18},
		{"center", 3, 9, 15},
		{"space-between", 0, 9, 18},
		{"space-around", 1, 9, 17},
		{"space-evenly", 1.5, 9, 16.5},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <article style="height: 20px; font: 2px weasyprint;
                      display: flex; flex-direction: column; justify-content: %s">
        <div style="margin: 1em 0">A</div>
        <div style="padding: 1em 0">B</div>
        <div>C</div>
      </article>
    `, data.align))
		html := unpack1(page)
		body := unpack1(html)
		article := unpack1(body)
		div1, div2, div3 := unpack3(article)
		assertText(t, div1.Box().Children[0].Box().Children[0], "A")
		assertText(t, div2.Box().Children[0].Box().Children[0], "B")
		assertText(t, div3.Box().Children[0].Box().Children[0], "C")
		assertPosXEqual(t, div1, div2, div3, article)
		tu.AssertEqual(t, article.Box().PositionY, 0)
		tu.AssertEqual(t, article.Box().Height, 20)
		tu.AssertEqual(t, div1.Box().PositionY, data.y1)
		tu.AssertEqual(t, div2.Box().PositionY, data.y2)
		tu.AssertEqual(t, div3.Box().PositionY, data.y3)
		tu.AssertEqual(t, div1.Box().MarginHeight(), 6)
		tu.AssertEqual(t, div2.Box().MarginHeight(), 6)
		tu.AssertEqual(t, div3.Box().MarginHeight(), 2)
	}
}

func TestFlexItemTable(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #1805.
	page := renderOnePage(t, `
      <article style="display: flex; font: 2px weasyprint">
        <table><tr><td>A</tr></td></table>
        <table><tr><td>B</tr></td></table>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	table_wrapper1, table_wrapper2 := unpack2(article)
	tu.Assert(t, table_wrapper1.Box().Width == table_wrapper2.Box().Width && table_wrapper2.Box().Width == Fl(2))
	tu.AssertEqual(t, table_wrapper1.Box().PositionX, 0)
	tu.AssertEqual(t, table_wrapper2.Box().PositionX, 2)
}

func TestFlexItemTableWidth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #1805.
	page := renderOnePage(t, `
      <article style="display: flex; font: 2px weasyprint; width: 40px">
        <table style="width: 25%"><tr><td>A</tr></td></table>
        <table style="width: 25%"><tr><td>B</tr></td></table>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	table_wrapper1, table_wrapper2 := unpack2(article)
	tu.Assert(t, table_wrapper1.Box().Width == table_wrapper2.Box().Width && table_wrapper2.Box().Width == Fl(10))
	tu.AssertEqual(t, table_wrapper1.Box().PositionX, 0)
	tu.AssertEqual(t, table_wrapper2.Box().PositionX, 10)
}

func TestFlexWidthOnParent(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <div style="font: 2px weasyprint; width: 30px; display: flex;
                  flex-direction: column; align-items: flex-start">
          <article>XXXX</article>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	article := unpack1(div)
	tu.AssertEqual(t, article.Box().Width, 8)
}

func TestFlexColumnItemFlex_1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <div style="font: 2px weasyprint; display: flex; flex-direction: column">
          <article>XXXX</article>
          <article style="flex: 1">XXXX</article>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().Height, 4)
}

func TestFlexRowItemFlex_0(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <div style="font: 2px weasyprint; display: flex">
          <article style="flex: 0">XXXX</article>
          <article style="flex: 0">XXXX</article>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	article1, article2 := unpack2(div)
	tu.AssertEqual(t, article1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, article1.Box().Width, Fl(8))
	tu.AssertEqual(t, article2.Box().PositionX, Fl(8))
	tu.AssertEqual(t, article2.Box().Width, Fl(8))
}

func TestFlexItemIntrinsicWidth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <div style="width: 100px; height: 100px;
                  display: flex; flex-direction: column; align-items: center">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 100">
          <rect width="200" height="100" fill="red" />
        </svg>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	svg := unpack1(div)
	tu.AssertEqual(t, svg.Box().Width, Fl(100))
}

func TestFlexAlignContentNegative(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <div style="height: 6px; width: 20px;
                  display: flex; flex-wrap: wrap; align-content: center">
        <span style="height: 2px; flex: none; margin: 1px; width: 8px"></span>
        <span style="height: 2px; flex: none; margin: 1px; width: 8px"></span>
        <span style="height: 2px; flex: none; margin: 1px; width: 8px"></span>
        <span style="height: 2px; flex: none; margin: 1px; width: 8px"></span>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	span1, span2, span3, span4 := unpack4(div)
	assertEquals(t, span1.Box().Height, span2.Box().Height, span3.Box().Height, span4.Box().Height, Fl(2))
	assertEquals(t, span1.Box().Width, span2.Box().Width, span3.Box().Width, span4.Box().Width, Fl(8))
	tu.Assert(t, span1.Box().PositionX == span3.Box().PositionX && span3.Box().PositionX == 0)
	tu.Assert(t, span2.Box().PositionX == span4.Box().PositionX && span4.Box().PositionX == 10)
	tu.Assert(t, span1.Box().PositionY == span2.Box().PositionY && span2.Box().PositionY == -1)
	tu.Assert(t, span3.Box().PositionY == span4.Box().PositionY && span4.Box().PositionY == 3)
}

func TestFlexShrink(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; width: 300px">
        <div style="flex: 0 2 auto; width: 300px"></div>
        <div style="width: 200px"></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2 := unpack2(article)
	tu.Assert(t, div1.Box().Width == div2.Box().Width && div2.Box().Width == Fl(150))
}

func TestFlexItemIntrinsicWidthShrink(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <div style="width: 10px; height: 100px; display: flex; flex-direction: column">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100">
          <rect width="100" height="100" fill="red" />
        </svg>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	svg := unpack1(div)
	tu.AssertEqual(t, svg.Box().Width, Fl(100))
}

func TestFlexItemIntrinsicHeightShrink(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <div style="width: 100px; height: 10px; display: flex; line-height: 0">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" height="100">
          <rect width="100" height="100" fill="red" />
        </svg>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	svg := unpack1(div)
	tu.AssertEqual(t, svg.Box().Height, Fl(100))
}

func TestFlexWrapInFlex(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <main style="display: flex; font: 2px weasyprint">
        <div style="display: flex; flex-wrap: wrap">
          <section style="width: 25%">A</section>
          <section style="flex: 1 75%">B</section>
        </div>
      </main>
    `)
	html := unpack1(page)
	body := unpack1(html)
	main := unpack1(body)
	div := unpack1(main)
	section1, section2 := unpack2(div)
	assertEquals(t, section1.Box().PositionY, section2.Box().PositionY, Fl(0))
	tu.AssertEqual(t, section1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, section2.Box().PositionX, Fl(1)) // 25% * 4
}

func TestFlexAutoBreakBefore(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page1, page2 := renderTwoPages(t, `
      <style>
        @page { size: 4px 5px }
        body { font: 2px weasyprint }
      </style>
      <p>A<br>B</p>
      <article style="display: flex">
        <div>A</div>
      </article>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	p := unpack1(body)
	tu.AssertEqual(t, p.Box().Height, Fl(4))
	html = unpack1(page2)
	body = unpack1(html)
	article := unpack1(body)
	tu.AssertEqual(t, article.Box().Height, Fl(2))
}

func TestFlexGrowInFlexColumn(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <html style="width: 14px">
        <body style="display: flex; flex-direction: column;
                     border: 1px solid; padding: 1px">
          <main style="flex: 1 1 auto; min-height: 0">
            <div style="height: 5px">
    `)
	html := unpack1(page)
	body := unpack1(html)
	main := unpack1(body)
	_, div, _ := unpack3(main)
	assertEquals(t, body.Box().Height, div.Box().Height, Fl(5))
	assertEquals(t, body.Box().Width, div.Box().Width, Fl(10))
	tu.AssertEqual(t, body.Box().MarginWidth(), Fl(14))
	tu.AssertEqual(t, body.Box().MarginHeight(), Fl(9))
}

func TestFlexCollapsingMargin(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <p style="margin-bottom: 20px; height: 100px">ABC</p>
      <article style="display: flex; margin-top: 10px">
        <div>A</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	p, article := unpack2(body)
	div := unpack1(article)
	tu.AssertEqual(t, p.Box().PositionY, Fl(0))
	tu.AssertEqual(t, p.Box().Height, Fl(100))
	tu.AssertEqual(t, article.Box().PositionY, Fl(110))
	tu.AssertEqual(t, div.Box().PositionY, Fl(120))
}

func TestFlexDirectionColumnNextPage(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #2414.
	page1, page2 := renderTwoPages(t, `
      <style>
        @page { size: 4px 5px }
        html { font: 2px/1 weasyprint }
      </style>
      <div>1</div>
      <article style="display: flex; flex-direction: column">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div, article := unpack2(body)
	assertText(t, div.Box().Children[0].Box().Children[0], "1")
	tu.AssertEqual(t, div.Box().Children[0].Box().Children[0].Box().PositionY, Fl(0))
	assertText(t, article.Box().Children[0].Box().Children[0].Box().Children[0], "A")
	tu.AssertEqual(t, article.Box().Children[0].Box().Children[0].Box().Children[0].Box().PositionY, Fl(2))
	html = unpack1(page2)
	body = unpack1(html)
	article = unpack1(body)
	assertText(t, article.Box().Children[0].Box().Children[0].Box().Children[0], "B")
	tu.AssertEqual(t, article.Box().Children[0].Box().Children[0].Box().Children[0].Box().PositionY, Fl(0))
	assertText(t, article.Box().Children[1].Box().Children[0].Box().Children[0], "C")
	tu.AssertEqual(t, article.Box().Children[1].Box().Children[0].Box().Children[0].Box().PositionY, Fl(2))
}

func TestFlex_1ItemPadding(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; width: 100px; font: 2px weasyprint">
        <div>abc</div>
        <div style="flex: 1; padding-right: 5em">def</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2 := unpack2(article)
	tu.AssertEqual(t, div1.Box().BorderWidth()+div2.Box().BorderWidth(), article.Box().Width)
}

func TestFlex_1ItemPaddingDirectionColumn(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column; height: 100px;
                      font: 2px weasyprint">
        <div>abc</div>
        <div style="flex: 1; padding-top: 5em">def</div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	div1, div2 := unpack2(article)
	tu.AssertEqual(t, div1.Box().BorderHeight()+div2.Box().BorderHeight(), article.Box().Height)
}

func TestFlexItemReplaced(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <div style="display: flex">
        <svg style="display: block" height="100" width="100" xmlns="http://www.w3.org/2000/svg">
          <circle r="45" cx="50" cy="50" fill="red" />
        </svg>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	svg := unpack1(div)
	assertEquals(t, svg.Box().Width, svg.Box().Height, Fl(100))
}

func TestFlexNestedColumn(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for issue #2442.
	page := renderOnePage(t, `
      <section style="display: flex; flex-direction: column; width: 200px">
        <div style="display: flex; flex-direction: column">
          <p>
            A
          </p>
        </div>
      </section>
    `)
	html := unpack1(page)
	body := unpack1(html)
	section := unpack1(body)
	div := unpack1(section)
	p := unpack1(div)
	tu.AssertEqual(t, p.Box().Width, Fl(200))
}

func TestFlexBlockifyImage(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; line-height: 2">
        <img src="pattern.png">
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	img := unpack1(article)
	assertEquals(t, article.Box().Height, img.Box().Height, Fl(4))
}

func TestFlexImageMaxWidth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex">
        <img src="pattern.png" style="max-width: 2px">
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	img := unpack1(article)
	assertEquals(t, article.Box().Height, img.Box().Height, img.Box().Width, Fl(2))
}

func TestFlexImageMaxHeight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex">
        <img src="pattern.png" style="max-height: 2px">
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	img := unpack1(article)
	assertEquals(t, article.Box().Height, img.Box().Height, img.Box().Width, Fl(2))
}

func TestFlexImageMinWidth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: flex; width: 20px">
        <img style="min-width: 10px; flex: 1 0 auto" src="pattern.png">
        <div style="flex: 1 0 1px"></div>
      </article>
    `)
	html := unpack1(page)
	body := unpack1(html)
	article := unpack1(body)
	img, div := unpack2(article)
	assertEquals(t, article.Box().Height, img.Box().Height, img.Box().Width, div.Box().Height, Fl(10))
}
