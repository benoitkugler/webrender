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
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "div1")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "div2")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "div3")
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div1.Box().PositionX < div2.Box().PositionX, true, "div1")
	tu.AssertEqual(t, div2.Box().PositionX < div3.Box().PositionX, true, "div1")
}

func TestFlexDirectionRowRtl(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; direction: rtl">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "div1")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "div2")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "div3")
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div1.Box().PositionX+div1.Box().Width.V(), article.Box().PositionX+article.Box().Width.V(), "div1")
	tu.AssertEqual(t, div1.Box().PositionX > div2.Box().PositionX, true, "div1 > div2")
	tu.AssertEqual(t, div2.Box().PositionX > div3.Box().PositionX, true, "div2 > div3")
}

func TestFlexDirectionRowReverse(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: row-reverse">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div3.Box().PositionX+div3.Box().Width.V(), article.Box().PositionX+article.Box().Width.V(), "")
	tu.AssertEqual(t, div1.Box().PositionX < div2.Box().PositionX, true, "div1")
	tu.AssertEqual(t, div2.Box().PositionX < div3.Box().PositionX, true, "div1")
}

func TestFlexDirectionRowReverseRtl(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: row-reverse;
      direction: rtl">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div3.Box().PositionX, article.Box().PositionX, "")
	tu.AssertEqual(t, div1.Box().PositionX > div2.Box().PositionX, true, "div1 > div2")
	tu.AssertEqual(t, div2.Box().PositionX > div3.Box().PositionX, true, "div2 > div3")
}

func TestFlexDirectionColumn(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	assertPosXEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div1.Box().PositionY, article.Box().PositionY, "")
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true, "div1")
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true, "div1")
}

func TestFlexDirectionColumnRtl(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column;
      direction: rtl">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	assertPosXEqual(t, div1, div2, div3, article)

	tu.AssertEqual(t, div1.Box().PositionY, article.Box().PositionY, "")
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true, "div1")
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true, "div1")
}

func TestFlexDirectionColumnReverse(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column-reverse">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	assertPosXEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div3.Box().PositionY+div3.Box().Height.V(), article.Box().PositionY+article.Box().Height.V(), "")
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true, "div1")
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true, "div1")
}

func TestFlexDirectionColumnReverseRtl(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column-reverse;
      direction: rtl">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	assertPosXEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, div3.Box().PositionY+div3.Box().Height.V(), article.Box().PositionY+article.Box().Height.V(), "")
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true, "div1")
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true, "div1")
}

func TestFlexRowWrap(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-flow: wrap; width: 50px">
        <div style="width: 20px">A</div>
        <div style="width: 20px">B</div>
        <div style="width: 20px">C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	tu.AssertEqual(t, div1.Box().PositionY, div2.Box().PositionY, "")
	tu.AssertEqual(t, div2.Box().PositionY, article.Box().PositionY, "")
	tu.AssertEqual(t, div3.Box().PositionY, article.Box().PositionY+div2.Box().Height.V(), "")
	tu.AssertEqual(t, div1.Box().PositionX, div3.Box().PositionX, "")
	tu.AssertEqual(t, div3.Box().PositionX, article.Box().PositionX, "")
	tu.AssertEqual(t, div1.Box().PositionX < div2.Box().PositionX, true, "")
}

func TestFlexColumnWrap(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-flow: column wrap; height: 50px">
        <div style="height: 20px">A</div>
        <div style="height: 20px">B</div>
        <div style="height: 20px">C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	assertPosXEqual(t, div1, div2, article)
	tu.AssertEqual(t, div3.Box().PositionX, article.Box().PositionX+div2.Box().Width.V(), "")
	assertPosYEqual(t, div1, div3, article)
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true, "")
}

func TestFlexRowWrapReverse(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-flow: wrap-reverse; width: 50px">
        <div style="width: 20px">A</div>
        <div style="width: 20px">B</div>
        <div style="width: 20px">C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	assertPosYEqual(t, div1, article)
	assertPosYEqual(t, div2, div3)
	tu.AssertEqual(t, div3.Box().PositionY, article.Box().PositionY+div1.Box().Height.V(), "")
	assertPosXEqual(t, div1, div2, article)
	tu.AssertEqual(t, div2.Box().PositionX < div3.Box().PositionX, true, "")
}

func TestFlexColumnWrapReverse(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-flow: column wrap-reverse;
                      height: 50px">
        <div style="height: 20px">A</div>
        <div style="height: 20px">B</div>
        <div style="height: 20px">C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	assertPosXEqual(t, div1, article)
	assertPosXEqual(t, div2, div3)
	tu.AssertEqual(t, div3.Box().PositionX, article.Box().PositionX+div1.Box().Width.V(), "")
	assertPosYEqual(t, div1, div2, article)
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true, "")
}

func TestFlexDirectionColumnFixedHeightContainer(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <section style="height: 10px">
        <article style="display: flex; flex-direction: column">
          <div>A</div>
          <div>B</div>
          <div>C</div>
        </article>
      </section>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	section := body.Box().Children[0]
	article := section.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	assertPosXEqual(t, div1, div2, div3, article)
	assertPosYEqual(t, div1, article)
	tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY, true, "")
	tu.AssertEqual(t, div2.Box().PositionY < div3.Box().PositionY, true, "")
	tu.AssertEqual(t, section.Box().Height, pr.Float(10), "")
	tu.AssertEqual(t, article.Box().Height.V() > 10, true, "")
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
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     article := body.Box().Children[0]
//     div1, div2, div3 := unpack3(article)
//     tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text , "A", "")
//     tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text , "B", "")
//     tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text , "C", "")
//     tu.AssertEqual(t, (
//         div1.Box().PositionX ,
//         div2.Box().PositionX ,
//         div3.Box().PositionX ,
//         article.Box().PositionX)
//     tu.AssertEqual(t, div1.Box().PositionY , article.Box().PositionY
//     tu.AssertEqual(t, div1.Box().PositionY < div2.Box().PositionY < div3.Box().PositionY
//     tu.AssertEqual(t, article.Box().Height.V() , 10
//     tu.AssertEqual(t, div3.Box().PositionY > 10
// }

func TestFlexDirectionColumnFixedHeightWrap(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex; flex-direction: column; height: 10px;
                      flex-wrap: wrap">
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	tu.AssertEqual(t, div1.Box().PositionX != div2.Box().PositionX, true, "")
	tu.AssertEqual(t, div2.Box().PositionX != div3.Box().PositionX, true, "")
	assertPosYEqual(t, div1, article)
	assertPosYEqual(t, div1, div2, div3, article)
	tu.AssertEqual(t, article.Box().Height, pr.Float(10), "")
}

func TestFlexItemMinWidth(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex">
        <div style="min-width: 30px">A</div>
        <div style="min-width: 50px">B</div>
        <div style="min-width: 5px">C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	tu.AssertEqual(t, div1.Box().PositionX, pr.Float(0), "div1")
	tu.AssertEqual(t, div1.Box().Width, pr.Float(30), "div1")
	tu.AssertEqual(t, div2.Box().PositionX, pr.Float(30), "div2")
	tu.AssertEqual(t, div2.Box().Width, pr.Float(50), "div2")
	tu.AssertEqual(t, div3.Box().PositionX, pr.Float(80), "div3")
	tu.AssertEqual(t, div3.Box().Width.V() > pr.Float(5), true, "div3")
	assertPosYEqual(t, div1, div2, div3, article)
}

func TestFlexItemMinHeight(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <article style="display: flex">
        <div style="min-height: 30px">A</div>
        <div style="min-height: 50px">B</div>
        <div style="min-height: 5px">C</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div1, div2, div3 := unpack3(article)
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B", "")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "")
	tu.AssertEqual(t, div1.Box().Height.V(), div2.Box().Height.V(), "")
	tu.AssertEqual(t, div2.Box().Height.V(), div3.Box().Height.V(), "")
	tu.AssertEqual(t, div3.Box().Height.V(), article.Box().Height.V(), "")
	tu.AssertEqual(t, article.Box().Height.V(), pr.Float(50), "")
}

func TestFlexAutoMargin(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/800
	_ = renderOnePage(t, `<div style="display: flex; margin: auto">`)
	_ = renderOnePage(t, `<div style="display: flex; flex-direction: column; margin: auto">`)
}

func TestFlexNoBaseline(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/765
	_ = renderOnePage(t, `
      <div class="references" style="display: flex; align-items: baseline;">
        <div></div>
      </div>`)
}

func TestFlexAlignContent(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

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
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		article := body.Box().Children[0]
		section1, section2 := unpack2(article)
		line1 := section1.Box().Children[0]
		line2 := section2.Box().Children[0]
		span1 := line1.Box().Children[0]
		span2 := line2.Box().Children[0]
		tu.AssertEqual(t, section1.Box().PositionX, span1.Box().PositionX, "")
		tu.AssertEqual(t, span1.Box().PositionX, pr.Float(0), "")
		tu.AssertEqual(t, section1.Box().PositionY, span1.Box().PositionY, "")
		tu.AssertEqual(t, span1.Box().PositionY, data.y1, "")
		tu.AssertEqual(t, section2.Box().PositionX, span2.Box().PositionX, "")
		tu.AssertEqual(t, span2.Box().PositionX, pr.Float(0), "")
		tu.AssertEqual(t, section2.Box().PositionY, span2.Box().PositionY, "")
		tu.AssertEqual(t, span2.Box().PositionY, data.y2, "")
	}
}

func TestFlexItemPercentage(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/885
	page := renderOnePage(t, `
      <div style="display: flex; font-size: 15px; line-height: 1">
        <div style="height: 100%">a</div>
      </div>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	flex := body.Box().Children[0]
	flexItem := flex.Box().Children[0]
	tu.AssertEqual(t, flexItem.Box().Height, pr.Float(15), "")
}

func TestFlexUndefinedPercentageHeightMultipleLines(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1204
	_ = renderOnePage(t, `
      <div style="display: flex; flex-wrap: wrap; height: 100%">
        <div style="width: 100%">a</div>
        <div style="width: 100%">b</div>
      </div>`)
}

func TestFlexAbsolute(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1536
	_ = renderOnePage(t, `
      <div style="display: flex; position: absolute">
        <div>a</div>
      </div>`)
}
