package layout

import (
	"fmt"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for multicolumn layout.

func assertPos(t *testing.T, gotX, gotY, expX, expY pr.Float) {
	t.Helper()
	tu.AssertEqual(t, [2]pr.Float{gotX, gotY}, [2]pr.Float{expX, expY})
}

func assertText(t *testing.T, box Box, exp string) {
	t.Helper()
	tb, ok := box.(*bo.TextBox)
	tu.AssertEqual(t, ok, true)
	tu.AssertEqual(t, tb.TextS(), exp)
}

func columnsMetrics(columns []Box) (widths, heights, xs, ys []pr.Float) {
	for _, column := range columns {
		widths = append(widths, column.Box().Width.V())
		heights = append(heights, column.Box().Height.V())
		xs = append(xs, column.Box().PositionX.V())
		ys = append(ys, column.Box().PositionY.V())
	}
	return
}

func TestColumns(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, css := range []string{
		"columns: 4",
		"columns: 100px",
		"columns: 4 100px",
		"columns: 100px 4",
		"column-width: 100px",
		"column-count: 4",
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { %s; column-gap: 0 }
        body { margin: 0; font-family: weasyprint }
        @page { margin: 0; size: 400px 1000px }
      </style>
      <div>
        Ipsum dolor sit amet,
        consectetur adipiscing elit.
        Sed sollicitudin nibh
        et turpis molestie tristique.
      </div>
    `, css))
		html := unpack1(page)
		body := unpack1(html)
		div := unpack1(body)
		columns := div.Box().Children
		tu.AssertEqual(t, len(columns), 4)
		widths, _, xs, ys := columnsMetrics(columns)
		tu.AssertEqual(t, widths, []pr.Float{100, 100, 100, 100})
		tu.AssertEqual(t, xs, []pr.Float{0, 100, 200, 300})
		tu.AssertEqual(t, ys, []pr.Float{0, 0, 0, 0})
	}
}

func TestColumnGap(t *testing.T) {
	for _, data := range []struct {
		value string
		width pr.Float
	}{
		{"normal", 16},  // "normal" is 1em = 16px
		{"unknown", 16}, // default value is normal
		{"15px", 15},
		{"40%", 16},  // percentages are not allowed
		{"-1em", 16}, // negative values are not allowed
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 3; column-gap: %s }
        body { margin: 0; font-family: weasyprint }
        @page { margin: 0; size: 300px 1000px }
      </style>
      <div>
        Ipsum dolor sit amet,
        consectetur adipiscing elit.
        Sed sollicitudin nibh
        et turpis molestie tristique.
      </div>
    `, data.value))
		html := unpack1(page)
		body := unpack1(html)
		div := unpack1(body)
		columns := div.Box().Children
		tu.AssertEqual(t, len(columns), 3)
		widths, _, xs, ys := columnsMetrics(columns)

		tu.AssertEqual(t, widths, []pr.Float{100 - 2*data.width/3, 100 - 2*data.width/3, 100 - 2*data.width/3})
		tu.AssertEqual(t, xs, []pr.Float{0, 100 + data.width/3, 200 + 2*data.width/3})
		tu.AssertEqual(t, ys, []pr.Float{0, 0, 0})
	}
}

func TestColumnSpan1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body { margin: 0; font-family: weasyprint; line-height: 1 }
        div { columns: 2; width: 10em; column-gap: 0 }
        section { column-span: all; margin: 1em 0 }
      </style>
 
      <div>
        abc def
        <section>test</section>
        <section>test</section>
        ghi jkl
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	column1, column2, section1, section2, column3, column4 := unpack6(div)
	tu.AssertEqual(t, [2]pr.Float{column1.Box().PositionX, column1.Box().PositionY}, [2]pr.Float{0, 0})
	tu.AssertEqual(t, [2]pr.Float{column2.Box().PositionX, column2.Box().PositionY}, [2]pr.Float{5 * 16, 0})
	tu.AssertEqual(t, [2]pr.Float{section1.Box().ContentBoxX(), section1.Box().ContentBoxY()}, [2]pr.Float{0, 32})
	tu.AssertEqual(t, [2]pr.Float{section2.Box().ContentBoxX(), section2.Box().ContentBoxY()}, [2]pr.Float{0, 64})
	tu.AssertEqual(t, [2]pr.Float{column3.Box().PositionX, column3.Box().PositionY}, [2]pr.Float{0, 96})
	tu.AssertEqual(t, [2]pr.Float{column4.Box().PositionX, column4.Box().PositionY}, [2]pr.Float{5 * 16, 96})

	tu.AssertEqual(t, column1.Box().Height, Fl(16))
}

func TestColumnSpan2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
	<style>
		@font-face { src: url(weasyprint.otf); font-family: weasyprint }
		body { margin: 0; font-family: weasyprint; line-height: 1 }
		div { columns: 2; width: 10em; column-gap: 0 }
		section { column-span: all; margin: 1em 0 }
	</style>

	<div>
		<section>test</section>
		abc def
		ghi jkl
		mno pqr
		stu vwx
	</div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	section1, column1, column2 := unpack3(div)
	tu.AssertEqual(t, [2]pr.Float{section1.Box().ContentBoxX(), section1.Box().ContentBoxY()}, [2]pr.Float{0, 16})
	tu.AssertEqual(t, [2]pr.Float{column1.Box().PositionX, column1.Box().PositionY}, [2]pr.Float{0, 3 * 16})
	tu.AssertEqual(t, [2]pr.Float{column2.Box().PositionX, column2.Box().PositionY}, [2]pr.Float{5 * 16, 3 * 16})

	tu.AssertEqual(t, column1.Box().Height, Fl(16*4))
	tu.AssertEqual(t, column2.Box().Height, Fl(16*4))
}

func TestColumnSpan3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { margin: 0; size: 8px 3px }
        body { font-family: weasyprint; font-size: 1px }
        div { columns: 2; column-gap: 0; line-height: 1 }
        section { column-span: all }
      </style>
      <div>
        abc def
        ghi jkl
        <section>line1 line2</section>
        mno pqr
      </div>
    `)
	page1, page2 := pages[0], pages[1]
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	column1, column2, section := unpack3(div)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 0)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 4, 0)
	assertPos(t, section.Box().PositionX, section.Box().PositionY, 0, 2)

	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "abc")
	assertText(t, unpack1(column1.Box().Children[0].Box().Children[1]), "def")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "ghi")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[1]), "jkl")
	assertText(t, unpack1(section.Box().Children[0]), "line1")

	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	section, column1, column2 = unpack3(div)
	assertPos(t, section.Box().PositionX, section.Box().PositionY, 0, 0)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 1)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 4, 1)

	assertText(t, unpack1(section.Box().Children[0]), "line2")
	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "mno")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "pqr")
}

func TestColumnSpan4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1, page2 := renderTwoPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { margin: 0; size: 8px 3px }
        body { font-family: weasyprint; font-size: 1px }
        div { columns: 2; column-gap: 0; line-height: 1 }
        section { column-span: all }
      </style>
      <div>
        abc def
        <section>line1</section>
        ghi jkl
        mno pqr
      </div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	column1, column2, section, column3, column4 := unpack5(div)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 0)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 4, 0)
	assertPos(t, section.Box().PositionX, section.Box().PositionY, 0, 1)
	assertPos(t, column3.Box().PositionX, column3.Box().PositionY, 0, 2)
	assertPos(t, column4.Box().PositionX, column4.Box().PositionY, 4, 2)

	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "abc")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "def")
	assertText(t, unpack1(section.Box().Children[0]), "line1")
	assertText(t, unpack1(column3.Box().Children[0].Box().Children[0]), "ghi")
	assertText(t, unpack1(column4.Box().Children[0].Box().Children[0]), "jkl")

	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	column1, column2 = unpack2(div)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 0)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 4, 0)

	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "mno")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "pqr")
}

func TestColumnSpan5(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { margin: 0; size: 8px 3px }
        body { font-family: weasyprint; font-size: 1px }
        div { columns: 2; column-gap: 0; line-height: 1 }
        section { column-span: all }
      </style>
      <div>
        abc def
        ghi jkl
        <section>line1</section>
        mno pqr
      </div>
    `)
	page1, page2 := pages[0], pages[1]
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	column1, column2, section := unpack3(div)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 0)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 4, 0)
	assertPos(t, section.Box().PositionX, section.Box().PositionY, 0, 2)

	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "abc")
	assertText(t, unpack1(column1.Box().Children[0].Box().Children[1]), "def")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "ghi")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[1]), "jkl")
	assertText(t, unpack1(section.Box().Children[0]), "line1")

	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	column1, column2 = unpack2(div)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 0)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 4, 0)

	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "mno")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "pqr")
}

func TestColumnSpan6(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1, page2 := renderTwoPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { margin: 0; size: 8px 3px }
        body { font-family: weasyprint; font-size: 1px }
        div { columns: 2; column-gap: 0; line-height: 1 }
        section { column-span: all }
      </style>
      <div>
        abc def
        ghi jkl
        mno pqr
        <section>line1</section>
      </div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	column1, column2 := unpack2(div)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 0)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 4, 0)

	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "abc")
	assertText(t, unpack1(column1.Box().Children[0].Box().Children[1]), "def")
	assertText(t, unpack1(column1.Box().Children[0].Box().Children[2]), "ghi")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "jkl")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[1]), "mno")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[2]), "pqr")

	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	section := unpack1(div)
	assertText(t, unpack1(section.Box().Children[0]), "line1")
	assertPos(t, section.Box().PositionX, section.Box().PositionY, 0, 0)
}

func TestColumnSpan7(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { margin: 0; size: 8px 3px }
        body { font-family: weasyprint; font-size: 1px }
        div { columns: 2; column-gap: 0; line-height: 1 }
        section { column-span: all; font-size: 2px }
      </style>
      <div>
        abc def
        ghi jkl
        <section>l1</section>
        mno pqr
      </div>
    `)
	page1, page2 := pages[0], pages[1]

	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	column1, column2 := unpack2(div)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 0)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 4, 0)

	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "abc")
	assertText(t, unpack1(column1.Box().Children[0].Box().Children[1]), "def")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "ghi")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[1]), "jkl")

	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	section, column1, column2 := unpack3(div)
	assertPos(t, section.Box().PositionX, section.Box().PositionY, 0, 0)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 2)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 4, 2)

	assertText(t, unpack1(section.Box().Children[0]), "l1")
	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "mno")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "pqr")
}

func TestColumnSpan8(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { margin: 0; size: 8px 2px }
        body { font-family: weasyprint; font-size: 1px }
        div { columns: 2; column-gap: 0; line-height: 1 }
        section { column-span: all }
      </style>
      <div>
        abc def
        ghi jkl
        mno pqr
        <section>line1</section>
      </div>
    `)
	page1, page2 := pages[0], pages[1]

	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	column1, column2 := unpack2(div)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 0)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 4, 0)

	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "abc")
	assertText(t, unpack1(column1.Box().Children[0].Box().Children[1]), "def")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "ghi")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[1]), "jkl")

	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	column1, column2, section := unpack3(div)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 0)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 4, 0)
	assertPos(t, section.Box().PositionX, section.Box().PositionY, 0, 1)

	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "mno")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "pqr")
	assertText(t, unpack1(section.Box().Children[0]), "line1")
}

func TestColumnSpan9(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { margin: 0; size: 8px 3px }
        body { font-family: weasyprint; font-size: 1px }
        div { columns: 2; column-gap: 0; line-height: 1 }
        section { column-span: all }
      </style>
      <div>
        abc
        <section>line1</section>
        def ghi
      </div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	column1, section, column2, column3 := unpack4(div)
	assertPos(t, column1.Box().PositionX, column1.Box().PositionY, 0, 0)
	assertPos(t, section.Box().PositionX, section.Box().PositionY, 0, 1)
	assertPos(t, column2.Box().PositionX, column2.Box().PositionY, 0, 2)
	assertPos(t, column3.Box().PositionX, column3.Box().PositionY, 4, 2)

	assertText(t, unpack1(column1.Box().Children[0].Box().Children[0]), "abc")
	assertText(t, unpack1(section.Box().Children[0]), "line1")
	assertText(t, unpack1(column2.Box().Children[0].Box().Children[0]), "def")
	assertText(t, unpack1(column3.Box().Children[0].Box().Children[0]), "ghi")
}

func TestColumnSpanBalance(t *testing.T) {
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { margin: 0; size: 8px 5px }
        body { font-family: weasyprint; font-size: 1px }
        div { columns: 2; column-gap: 0; line-height: 1; column-fill: auto }
        section { column-span: all }
      </style>
      <div>
        abc def
        <section>line1</section>
        ghi jkl
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	column1, column2, section, column3 := unpack4(div)
	tu.AssertEqual(t, [2]pr.Float{column1.Box().PositionX, column1.Box().PositionY}, [2]pr.Float{0, 0})
	tu.AssertEqual(t, [2]pr.Float{column2.Box().PositionX, column2.Box().PositionY}, [2]pr.Float{4, 0})
	tu.AssertEqual(t, [2]pr.Float{section.Box().PositionX, section.Box().PositionY}, [2]pr.Float{0, 1})
	tu.AssertEqual(t, [2]pr.Float{column3.Box().PositionX, column3.Box().PositionY}, [2]pr.Float{0, 2})

	assertText(t, unpack1(unpack1(column1)).Box().Children[0], "abc")
	assertText(t, unpack1(unpack1(column2)).Box().Children[0], "def")
	assertText(t, unpack1(unpack1(section)), "line1")
	assertText(t, unpack1(unpack1(column3)).Box().Children[0], "ghi")
	assertText(t, unpack1(column3).Box().Children[1].Box().Children[0], "jkl")
}

func TestColumnsMultipage(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 2px }
      </style>
      <div>a b c d e f g</div>
    `)
	page1, page2 := pages[0], pages[1]
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	tu.AssertEqual(t, len(columns[0].Box().Children), 2)
	tu.AssertEqual(t, len(columns[1].Box().Children), 2)
	assertText(t, unpack1(columns[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[0].Box().Children[1]), "b")
	assertText(t, unpack1(columns[1].Box().Children[0]), "c")
	assertText(t, unpack1(columns[1].Box().Children[1]), "d")

	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	columns = div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	tu.AssertEqual(t, len(columns[0].Box().Children), 2)
	tu.AssertEqual(t, len(columns[1].Box().Children), 1)
	assertText(t, unpack1(columns[0].Box().Children[0]), "e")
	assertText(t, unpack1(columns[0].Box().Children[1]), "f")
	assertText(t, unpack1(columns[1].Box().Children[0]), "g")
}

func TestColumnsBreaks(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 2px }
        section { break-before: always; }
      </style>
      <div>a<section>b</section><section>c</section></div>
    `)
	page1, page2 := pages[0], pages[1]
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	tu.AssertEqual(t, len(columns[0].Box().Children), 1)
	tu.AssertEqual(t, len(columns[1].Box().Children), 1)
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[0]), "b")

	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	columns = div.Box().Children
	tu.AssertEqual(t, len(columns), 1)
	tu.AssertEqual(t, len(columns[0].Box().Children), 1)
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[0]), "c")
}

func TestColumnsBreakAfterColumn_1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 10px }
        section { break-after: column }
      </style>
      <div>a b <section>c</section> d</div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[1]), "b")
	assertText(t, unpack1(columns[0].Box().Children[1].Box().Children[0]), "c")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[0]), "d")
}

func TestColumnsBreakAfterColumn_2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 10px }
        section { break-after: column }
      </style>
      <div><section>a</section> b c d</div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[0]), "b")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[1]), "c")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[2]), "d")
}

func TestColumnsBreakAfterAvoidColumn(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 10px }
        section { break-after: avoid-column }
      </style>
      <div>a <section>b</section> c d</div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[0].Box().Children[1].Box().Children[0]), "b")
	assertText(t, unpack1(columns[0].Box().Children[2].Box().Children[0]), "c")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[0]), "d")
}

func TestColumnsBreakBeforeColumn_1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 10px }
        section { break-before: column }
      </style>
      <div>a b c <section>d</section></div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[1]), "b")
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[2]), "c")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[0]), "d")
}

func TestColumnsBreakBeforeColumn_2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 10px }
        section { break-before: column }
      </style>
      <div>a <section>b</section> c d</div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[0]), "b")
	assertText(t, unpack1(columns[1].Box().Children[1].Box().Children[0]), "c")
	assertText(t, unpack1(columns[1].Box().Children[1].Box().Children[1]), "d")
}

func TestColumnsBreakBeforeAvoidColumn(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 10px }
        section { break-before: avoid-column }
      </style>
      <div>a b <section>c</section> d</div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[1]), "b")
	assertText(t, unpack1(columns[0].Box().Children[1].Box().Children[0]), "c")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[0]), "d")
}

func TestColumnsBreakInsideColumn_1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 10px }
        section { break-inside: avoid-column }
      </style>
      <div><section>a b c</section> d</div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[1]), "b")
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[2]), "c")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[0]), "d")
}

func TestColumnsBreakInsideColumn_2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 10px }
        section { break-inside: avoid-column }
      </style>
      <div>a <section>b c d</section></div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[0]), "b")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[1]), "c")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[2]), "d")
}

func TestColumnsBreakInsideColumnNotEmptyPage(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page1 := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 10px }
        section { break-inside: avoid-column }
      </style>
      <p>p</p>
      <div><section>a b c</section> d</div>
    `)
	html := unpack1(page1)
	body := unpack1(html)
	p, div := unpack2(body)
	assertText(t, unpack1(p.Box().Children[0]), "p")
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2)
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[1]), "b")
	assertText(t, unpack1(columns[0].Box().Children[0].Box().Children[2]), "c")
	assertText(t, unpack1(columns[1].Box().Children[0].Box().Children[0]), "d")
}

func TestColumnsNotEnoughContent(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 5; column-gap: 0 }
        body { margin: 0; font-family: weasyprint; font-size: 1px }
        @page { margin: 0; size: 5px }
      </style>
      <div>a b c</div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().Width, Fl(5))
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 3)
	widths, _, xs, ys := columnsMetrics(columns)
	tu.AssertEqual(t, widths, []pr.Float{1, 1, 1})
	tu.AssertEqual(t, xs, []pr.Float{0, 1, 2})
	tu.AssertEqual(t, ys, []pr.Float{0, 0, 0})
}

func TestColumnsHigherThanPage(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 5; column-gap: 0 }
        body { margin: 0; font-family: weasyprint; font-size: 2px }
        @page { margin: 0; size: 5px 1px }
      </style>
      <div>a b c d e f g h</div>
    `)
	page1, page2 := pages[0], pages[1]
	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().Width, Fl(5))
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 5)
	assertText(t, unpack1(columns[0].Box().Children[0]), "a")
	assertText(t, unpack1(columns[1].Box().Children[0]), "b")
	assertText(t, unpack1(columns[2].Box().Children[0]), "c")
	assertText(t, unpack1(columns[3].Box().Children[0]), "d")
	assertText(t, unpack1(columns[4].Box().Children[0]), "e")

	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	tu.AssertEqual(t, div.Box().Width, Fl(5))
	columns = div.Box().Children
	tu.AssertEqual(t, len(columns), 3)
	assertText(t, unpack1(columns[0].Box().Children[0]), "f")
	assertText(t, unpack1(columns[1].Box().Children[0]), "g")
	assertText(t, unpack1(columns[2].Box().Children[0]), "h")
}

func TestColumnsEmpty(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 3 }
        body { margin: 0; font-family: weasyprint }
        @page { margin: 0; size: 3px; font-size: 1px }
      </style>
      <div></div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().Width, Fl(3))
	tu.AssertEqual(t, div.Box().Height, Fl(0))
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 0)
}

func TestColumnsFixedHeight(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, prop := range []string{"height", "min-height"} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 4; column-gap: 0; %s: 10px }
        body { margin: 0; font-family: weasyprint; line-height: 1px }
        @page { margin: 0; size: 4px 50px; font-size: 1px }
      </style>
      <div>a b c</div>
    `, prop))
		html := unpack1(page)
		body := unpack1(html)
		div := unpack1(body)
		tu.AssertEqual(t, div.Box().Width, Fl(4))
		columns := div.Box().Children
		tu.AssertEqual(t, len(columns), 3)

		widths, heights, xs, ys := columnsMetrics(columns)

		tu.AssertEqual(t, widths, []pr.Float{1, 1, 1})
		tu.AssertEqual(t, heights, []pr.Float{10, 10, 10})
		tu.AssertEqual(t, xs, []pr.Float{0, 1, 2})
		tu.AssertEqual(t, ys, []pr.Float{0, 0, 0})
	}
}

func TestColumnsPadding(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 4; column-gap: 0; padding: 1px }
        body { margin: 0; font-family: weasyprint; line-height: 1px }
        @page { margin: 0; size: 6px 50px; font-size: 1px }
      </style>
      <div>a b c</div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().Width, Fl(4))
	tu.AssertEqual(t, div.Box().Height, Fl(1))
	tu.AssertEqual(t, div.Box().PaddingWidth(), Fl(6))
	tu.AssertEqual(t, div.Box().PaddingHeight(), Fl(3))
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 3)
	widths, heights, xs, ys := columnsMetrics(columns)
	tu.AssertEqual(t, widths, []pr.Float{1, 1, 1})
	tu.AssertEqual(t, heights, []pr.Float{1, 1, 1})
	tu.AssertEqual(t, xs, []pr.Float{1, 2, 3})
	tu.AssertEqual(t, ys, []pr.Float{1, 1, 1})
}

func TestColumnsRelative(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article { position: absolute; top: 3px }
        div { columns: 4; column-gap: 0; position: relative;
              top: 1px; left: 2px }
        body { margin: 0; font-family: weasyprint; line-height: 1px }
        @page { margin: 0; size: 4px 50px; font-size: 1px }
      </style>
      <div>a b c d<article>e</article></div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().Width, Fl(4))
	columns := div.Box().Children
	widths, _, xs, ys := columnsMetrics(columns)
	tu.AssertEqual(t, widths, []pr.Float{1, 1, 1, 1})
	tu.AssertEqual(t, xs, []pr.Float{2, 3, 4, 5})
	tu.AssertEqual(t, ys, []pr.Float{1, 1, 1, 1})
	column4 := columns[len(columns)-1]
	columnLine := unpack1(column4)
	absoluteArticle := columnLine.Box().Children[1]
	absoluteLine := unpack1(absoluteArticle)
	span := unpack1(absoluteLine)
	tu.AssertEqual(t, span.Box().PositionX, Fl(5))
	tu.AssertEqual(t, span.Box().PositionY, Fl(4))
}

func TestColumnsRegression1(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test #1 for https://github.com/Kozea/WeasyPrint/issues/659
	pages := renderPages(t, `
      <style>
        @page {margin: 0; width: 100px; height: 100px}
        body {margin: 0; font-size: 1px}
      </style>
      <div style="height:95px">A</div>
      <div style="column-count:2">
        <div style="height:20px">B1</div>
        <div style="height:20px">B2</div>
        <div style="height:20px">B3</div>
      </div>
      <div style="height:95px">C</div>
    `)
	page1, page2, page3 := pages[0], pages[1], pages[2]

	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().PositionY, Fl(0))
	assertText(t, unpack1(div.Box().Children[0]), "A")

	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	tu.AssertEqual(t, div.Box().PositionY, Fl(0))
	column1, column2 := unpack2(div)
	tu.AssertEqual(t, column1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, column2.Box().PositionY, Fl(0))
	div1, div2 := unpack2(column1)
	div3 := unpack1(column2)
	tu.AssertEqual(t, div1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div3.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div2.Box().PositionY, Fl(20))
	assertText(t, unpack1(div1.Box().Children[0]), "B1")
	assertText(t, unpack1(div2.Box().Children[0]), "B2")
	assertText(t, unpack1(div3.Box().Children[0]), "B3")

	html = unpack1(page3)
	body = unpack1(html)
	div = unpack1(body)
	tu.AssertEqual(t, div.Box().PositionY, Fl(0))
	assertText(t, unpack1(div.Box().Children[0]), "C")
}

func TestColumnsRegression2(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test #2 for https://github.com/Kozea/WeasyPrint/issues/659
	pages := renderPages(t, `
      <style>
        @page {margin: 0; width: 100px; height: 100px}
        body {margin: 0; font-size: 1px}
      </style>
      <div style="column-count:2">
        <div style="height:20px">B1</div>
        <div style="height:60px">B2</div>
        <div style="height:60px">B3</div>
        <div style="height:60px">B4</div>
      </div>
    `)
	page1, page2 := pages[0], pages[1]

	html := unpack1(page1)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().PositionY, Fl(0))
	column1, column2 := unpack2(div)
	tu.AssertEqual(t, column1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, column2.Box().PositionY, Fl(0))
	div1, div2 := unpack2(column1)
	div3 := unpack1(column2)
	tu.AssertEqual(t, div1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div3.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div2.Box().PositionY, Fl(20))
	assertText(t, unpack1(div1.Box().Children[0]), "B1")
	assertText(t, unpack1(div2.Box().Children[0]), "B2")
	assertText(t, unpack1(div3.Box().Children[0]), "B3")

	html = unpack1(page2)
	body = unpack1(html)
	div = unpack1(body)
	tu.AssertEqual(t, div.Box().PositionY, Fl(0))
	column1 = unpack1(div)
	tu.AssertEqual(t, column1.Box().PositionY, Fl(0))
	div1 = unpack1(column1)
	tu.AssertEqual(t, div1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div3.Box().PositionY, Fl(0))
	assertText(t, unpack1(div1.Box().Children[0]), "B4")
}

func TestColumnsRegression3(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test #3 for https://github.com/Kozea/WeasyPrint/issues/659
	page := renderOnePage(t, `
      <style>
        @page {margin: 0; width: 100px; height: 100px}
        body {margin: 0; font-size: 10px}
      </style>
      <div style="column-count:2">
        <div style="height:20px; margin:5px">B1</div>
        <div style="height:60px">B2</div>
        <div style="height:60px">B3</div>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().PositionY, Fl(0))
	column1, column2 := unpack2(div)
	tu.AssertEqual(t, column1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, column2.Box().PositionY, Fl(0))
	div1, div2 := unpack2(column1)
	div3 := unpack1(column2)
	tu.AssertEqual(t, div1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div3.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div2.Box().PositionY, Fl(30))
	tu.AssertEqual(t, div.Box().Height, Fl(5+20+5+60))
	assertText(t, unpack1(div1.Box().Children[0]), "B1")
	assertText(t, unpack1(div2.Box().Children[0]), "B2")
	assertText(t, unpack1(div3.Box().Children[0]), "B3")
}

func TestColumnsRegression4(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/897
	page := renderOnePage(t, `
      <div style="position:absolute">
        <div style="column-count:2">
          <div>a</div>
        </div>
      </div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	tu.AssertEqual(t, div.Box().PositionY, Fl(0))
	column1 := unpack1(div)
	tu.AssertEqual(t, column1.Box().PositionY, Fl(0))
	div1 := unpack1(column1)
	tu.AssertEqual(t, div1.Box().PositionY, Fl(0))
}

func TestColumnsRegression5(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1191
	_ = renderPages(t, `
      <style>
        @page {width: 100px; height: 100px}
      </style>
      <div style="height: 1px"></div>
      <div style="columns: 2">
        <div style="break-after: avoid">
          <div style="height: 50px"></div>
        </div>
        <div style="break-after: avoid">
          <div style="height: 50px"></div>
          <p>a</p>
        </div>
      </div>
      <div style="height: 50px"></div>
    `)
}

func TestColumnsRegression_6(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/2103
	renderPages(t, `
      <style>
        @page {width: 100px; height: 100px}
      </style>
      <div style="columns: 2; column-width: 100px; width: 10px">abc def</div>
    `)
}
