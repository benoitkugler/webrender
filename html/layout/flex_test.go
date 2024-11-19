package layout

import (
	"fmt"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

//  Tests for flex layout.

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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "C")
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div1.Box().PositionX < div2.Box().PositionX, true)
	tu.AssertEqual(t, div2.Box().PositionX < div3.Box().PositionX, true)
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "C")
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "C")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "A")
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "C")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "A")
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "C")
	assertPosXEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div1.Box().PositionY, article.Box().PositionY)
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true)
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true)
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "C")
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "C")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "A")
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "C")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "A")
	assertPosXEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div3.Box().PositionY+div3.Box().Height.V(), article.Box().PositionY+article.Box().Height.V())
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true)
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true)
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "C")
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "C")
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "C")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "B")
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "C")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "B")
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "C")
	assertPosXEqual(t, div1, div2, div3, article)
	assertPosYEqual(t, div1, article)
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true)
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true)
	tu.AssertEqual(t, section.Box().Height, Fl(10))
	tu.AssertEqual(t, article.Box().Height.V() > 10, true)
}

// @pytest.mark.xfail
// func TestFlexDirectionColumnFixedHeight(t *testing.T){
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     page := renderOnePage(t,`
//       <article style="display: flex; flex-direction: column; height: 10px">
//         <div>A</div>
//         <div>B</div>
//         <div>C</div>
//       </article>
//     `)
//     html =  unpack1(page)
//     body :=  unpack1(html)
//     article := unpack1(body)
//     div1, div2, div3 := unpack3(article)
//     tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text , "A")
//     tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text , "B")
//     tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text , "C")
//     (
//         div1.Box().PositionX ,
//         div2.Box().PositionX ,
//         div3.Box().PositionX ,
//         article.Box().PositionX)
//     tu.AssertEqual(t, div1.Box().PositionY , article.Y
//     tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY < div3.Y
//     tu.AssertEqual(t, article.Box().Height.V0
//     tu.AssertEqual(t, div3.Box().Positio0
// }

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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "C")
	tu.AssertEqual(t, div1.Box().PositionX != div2.Box().PositionX, true)
	tu.AssertEqual(t, div2.Box().PositionX != div3.Box().PositionX, true)
	assertPosYEqual(t, div1, article)
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, article.Box().Height, Fl(10))
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "C")
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
	tu.AssertEqual(t, unpack1(div1.Box().Children[0]).(*bo.TextBox).Text, "A")
	tu.AssertEqual(t, unpack1(div2.Box().Children[0]).(*bo.TextBox).Text, "B")
	tu.AssertEqual(t, unpack1(div3.Box().Children[0]).(*bo.TextBox).Text, "C")
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
		y1, y2 pr.Float
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
