package layout

import (
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

type Fl = pr.Float

// Tests for grid layout.

func TestGridEmpty(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: grid">
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	tu.AssertEqual(t, article.Box().PositionX, Fl(0))
	tu.AssertEqual(t, article.Box().PositionY, Fl(0))
	tu.AssertEqual(t, article.Box().Width, html.Box().Width)
	tu.AssertEqual(t, article.Box().Height, Fl(0))
}

func TestGridSingleItem(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: grid">
        <div>a</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div := article.Box().Children[0]
	tu.AssertEqual(t, article.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div.Box().PositionX, Fl(0))
	tu.AssertEqual(t, article.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div.Box().PositionY, Fl(0))
	tu.AssertEqual(t, article.Box().Width, html.Box().Width)
	tu.AssertEqual(t, div.Box().Width, html.Box().Width)
}

func TestGridRows(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <article style="display: grid">
        <div>a</div>
        <div>b</div>
        <div>c</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c := unpack3(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_a.Box().PositionY < div_b.Box().PositionY, true)
	tu.AssertEqual(t, div_b.Box().PositionY < div_c.Box().PositionY, true)
	tu.AssertEqual(t, div_a.Box().Height, div_b.Box().Height)
	tu.AssertEqual(t, div_b.Box().Height, div_c.Box().Height)
	tu.AssertEqual(t, article.Box().Width, html.Box().Width)
	tu.AssertEqual(t, div_a.Box().Width, html.Box().Width)
	tu.AssertEqual(t, div_b.Box().Width, html.Box().Width)
	tu.AssertEqual(t, div_c.Box().Width, html.Box().Width)
}

func TestGridTemplateFr(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-rows: auto 1fr;
          grid-template-columns: auto 1fr;
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div>a</div> <div>b</div>
        <div>c</div> <div>d</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d := unpack4(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(2))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(2))
	tu.AssertEqual(t, div_a.Box().Height, Fl(2))
	tu.AssertEqual(t, div_b.Box().Height, Fl(2))
	tu.AssertEqual(t, div_c.Box().Height, Fl(2))
	tu.AssertEqual(t, div_d.Box().Height, Fl(2))
	tu.AssertEqual(t, div_a.Box().Width, Fl(2))
	tu.AssertEqual(t, div_c.Box().Width, Fl(2))
	tu.AssertEqual(t, div_b.Box().Width, Fl(8))
	tu.AssertEqual(t, div_d.Box().Width, Fl(8))
	tu.AssertEqual(t, article.Box().Width, Fl(10))
}

func TestGridTemplateAreas(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-areas: 'a b' 'c d';
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div>a</div> <div>b</div>
        <div>c</div> <div>d</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d := unpack4(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(5))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(5))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_d.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_a.Box().Height, Fl(2))
	tu.AssertEqual(t, div_b.Box().Height, Fl(2))
	tu.AssertEqual(t, div_c.Box().Height, Fl(2))
	tu.AssertEqual(t, div_d.Box().Height, Fl(2))
	tu.AssertEqual(t, div_a.Box().Width, Fl(5))
	tu.AssertEqual(t, div_b.Box().Width, Fl(5))
	tu.AssertEqual(t, div_c.Box().Width, Fl(5))
	tu.AssertEqual(t, div_d.Box().Width, Fl(5))
	tu.AssertEqual(t, article.Box().Width, Fl(10))
}

func TestGridTemplateAreasGridArea(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-areas: 'b a' 'd c';
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div style="grid-area: a">a</div> <div style="grid-area: b">b</div>
        <div style="grid-area: c">c</div> <div style="grid-area: d">d</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d := unpack4(article)
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(5))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(5))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_d.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_a.Box().Height, Fl(2))
	tu.AssertEqual(t, div_b.Box().Height, Fl(2))
	tu.AssertEqual(t, div_c.Box().Height, Fl(2))
	tu.AssertEqual(t, div_d.Box().Height, Fl(2))
	tu.AssertEqual(t, div_a.Box().Width, Fl(5))
	tu.AssertEqual(t, div_b.Box().Width, Fl(5))
	tu.AssertEqual(t, div_c.Box().Width, Fl(5))
	tu.AssertEqual(t, div_d.Box().Width, Fl(5))
	tu.AssertEqual(t, article.Box().Width, Fl(10))
}

func TestGridTemplateAreasEmptyRow(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-areas: 'b a' 'd a' 'd c';
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div style="grid-area: a">a</div> <div style="grid-area: b">b</div>
        <div style="grid-area: c">c</div> <div style="grid-area: d">d</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d := unpack4(article)
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(5))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(5))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_d.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_a.Box().Height, Fl(2))
	tu.AssertEqual(t, div_b.Box().Height, Fl(2))
	tu.AssertEqual(t, div_c.Box().Height, Fl(2))
	tu.AssertEqual(t, div_d.Box().Height, Fl(2))
	tu.AssertEqual(t, div_a.Box().Width, Fl(5))
	tu.AssertEqual(t, div_b.Box().Width, Fl(5))
	tu.AssertEqual(t, div_c.Box().Width, Fl(5))
	tu.AssertEqual(t, div_d.Box().Width, Fl(5))
	tu.AssertEqual(t, article.Box().Width, Fl(10))
}

func TestGridTemplateAreasMultipleRows(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-areas: 'b a' 'd a' '. c';
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div style="grid-area: a">a</div> <div style="grid-area: b">b</div>
        <div style="grid-area: c">c</div> <div style="grid-area: d">d</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d := unpack4(article)
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(5))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(5))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_d.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_a.Box().Height, Fl(4))
	tu.AssertEqual(t, div_b.Box().Height, Fl(2))
	tu.AssertEqual(t, div_c.Box().Height, Fl(2))
	tu.AssertEqual(t, div_d.Box().Height, Fl(2))
	tu.AssertEqual(t, div_a.Box().Width, Fl(5))
	tu.AssertEqual(t, div_b.Box().Width, Fl(5))
	tu.AssertEqual(t, div_c.Box().Width, Fl(5))
	tu.AssertEqual(t, div_d.Box().Width, Fl(5))
	tu.AssertEqual(t, article.Box().Width, Fl(10))
}

func TestGridTemplateAreasMultipleColumns(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-areas: 'b b' 'c a';
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div style="grid-area: a">a</div>
        <div style="grid-area: b">b</div>
        <div style="grid-area: c">c</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c := unpack3(article)
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(5))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_a.Box().Height, Fl(2))
	tu.AssertEqual(t, div_b.Box().Height, Fl(2))
	tu.AssertEqual(t, div_c.Box().Height, Fl(2))
	tu.AssertEqual(t, div_a.Box().Width, Fl(5))
	tu.AssertEqual(t, div_c.Box().Width, Fl(5))
	tu.AssertEqual(t, div_b.Box().Width, Fl(10))
	tu.AssertEqual(t, article.Box().Width, Fl(10))
}

func TestGridTemplateAreasOverlap(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-areas: 'a b' 'c d';
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div style="grid-area: a">a</div>
        <div style="grid-area: a">a</div>
        <div style="grid-area: a">a</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a1, div_a2, div_a3 := unpack3(article)
	tu.AssertEqual(t, div_a1.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_a2.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_a3.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_a1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_a2.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_a3.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_a1.Box().Width, Fl(6))
	tu.AssertEqual(t, div_a2.Box().Width, Fl(6))
	tu.AssertEqual(t, div_a3.Box().Width, Fl(6))
	tu.AssertEqual(t, div_a1.Box().Height, Fl(2))
	tu.AssertEqual(t, div_a2.Box().Height, Fl(2))
	tu.AssertEqual(t, div_a3.Box().Height, Fl(2))
	tu.AssertEqual(t, article.Box().Width, Fl(10))
}

func TestGridTemplateAreasExtraSpan(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-areas: 'a . b' 'c d d';
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div style="grid-area: a">a</div>
        <div style="grid-area: b">b</div>
        <div style="grid-area: c">c</div>
        <div style="grid-area: d">d</div>
        <div style="grid-row: span 2; grid-column: span 2">e</div>
        <div>f</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d, div_e, div_f := unpack6(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_e.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(4))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(6))
	tu.AssertEqual(t, div_f.Box().PositionX, Fl(6))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_d.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_e.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_f.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_a.Box().Width, Fl(4))
	tu.AssertEqual(t, div_b.Box().Width, Fl(4))
	tu.AssertEqual(t, div_c.Box().Width, Fl(4))
	tu.AssertEqual(t, div_f.Box().Width, Fl(4))
	tu.AssertEqual(t, div_d.Box().Width, Fl(6))
	tu.AssertEqual(t, div_e.Box().Width, Fl(6))

	for _, child := range article.Box().Children {
		tu.AssertEqual(t, child.Box().Height, Fl(2))
	}
	tu.AssertEqual(t, article.Box().Width, Fl(10))
}

func TestGridTemplateAreasExtraSpanDense(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-auto-flow: dense;
          grid-template-areas: 'a . b' 'c d d';
          line-height: 1;
          width: 9px;
        }
      </style>
      <article>
        <div style="grid-area: a">a</div>
        <div style="grid-area: b">b</div>
        <div style="grid-area: c">c</div>
        <div style="grid-area: d">d</div>
        <div style="grid-row: span 2; grid-column: span 2">e</div>
        <div>f</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d, div_e, div_f := unpack6(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_e.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(3))
	tu.AssertEqual(t, div_f.Box().PositionX, Fl(3))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(6))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_f.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_d.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_e.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_a.Box().Width, Fl(3))
	tu.AssertEqual(t, div_b.Box().Width, Fl(3))
	tu.AssertEqual(t, div_c.Box().Width, Fl(3))
	tu.AssertEqual(t, div_f.Box().Width, Fl(3))
	tu.AssertEqual(t, div_d.Box().Width, Fl(6))
	tu.AssertEqual(t, div_e.Box().Width, Fl(6))

	for _, child := range article.Box().Children {
		tu.AssertEqual(t, child.Box().Height, Fl(2))
	}
	tu.AssertEqual(t, article.Box().Width, Fl(9))
}

func TestGridTemplateRepeatFr(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-columns: repeat(2, 1fr 2fr);
          line-height: 1;
          width: 12px;
        }
      </style>
      <article>
        <div>a</div>
        <div>b</div>
        <div>c</div>
        <div>d</div>
        <div>e</div>
        <div>f</div>
        <div>g</div>
        <div>h</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d, div_e, div_f, div_g, div_h := unpack8(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_e.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(2))
	tu.AssertEqual(t, div_f.Box().PositionX, Fl(2))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(6))
	tu.AssertEqual(t, div_g.Box().PositionX, Fl(6))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(8))
	tu.AssertEqual(t, div_h.Box().PositionX, Fl(8))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_d.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_e.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_f.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_g.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_h.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_a.Box().Width, Fl(2))
	tu.AssertEqual(t, div_c.Box().Width, Fl(2))
	tu.AssertEqual(t, div_e.Box().Width, Fl(2))
	tu.AssertEqual(t, div_g.Box().Width, Fl(2))
	tu.AssertEqual(t, div_b.Box().Width, Fl(4))
	tu.AssertEqual(t, div_d.Box().Width, Fl(4))
	tu.AssertEqual(t, div_f.Box().Width, Fl(4))
	tu.AssertEqual(t, div_h.Box().Width, Fl(4))

	for _, child := range article.Box().Children {
		tu.AssertEqual(t, child.Box().Height, Fl(2))
	}
	tu.AssertEqual(t, article.Box().Width, Fl(12))
}

func TestGridTemplateShorthandFr(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template: auto 1fr / auto 1fr auto;
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div>a</div>
        <div>b</div>
        <div>c</div>
        <div>d</div>
        <div>e</div>
        <div>f</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d, div_e, div_f := unpack6(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(2))
	tu.AssertEqual(t, div_e.Box().PositionX, Fl(2))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(8))
	tu.AssertEqual(t, div_f.Box().PositionX, Fl(8))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_d.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_e.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_f.Box().PositionY, Fl(2))

	tu.AssertEqual(t, div_a.Box().Width, Fl(2))
	tu.AssertEqual(t, div_c.Box().Width, Fl(2))
	tu.AssertEqual(t, div_d.Box().Width, Fl(2))
	tu.AssertEqual(t, div_f.Box().Width, Fl(2))
	tu.AssertEqual(t, div_b.Box().Width, Fl(6))
	tu.AssertEqual(t, div_e.Box().Width, Fl(6))

	for _, child := range article.Box().Children {
		tu.AssertEqual(t, child.Box().Height, Fl(2))
	}
	tu.AssertEqual(t, article.Box().Width, Fl(10))
}

func TestGridShorthandAutoFlowRowsFrSize(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid: auto-flow 1fr / 6px;
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div>a</div>
        <div>b</div>
        <div>c</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c := unpack3(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(4))

	tu.AssertEqual(t, div_a.Box().Width, Fl(6))
	tu.AssertEqual(t, div_b.Box().Width, Fl(6))
	tu.AssertEqual(t, div_c.Box().Width, Fl(6))

	for _, child := range article.Box().Children {
		tu.AssertEqual(t, child.Box().Height, Fl(2))
	}
	tu.AssertEqual(t, article.Box().Width, Fl(10))
}

func TestGridShorthandAutoFlowColumnsNoneDense(t *testing.T) {
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid: none / auto-flow 1fr dense;
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div>a</div>
        <div>b</div>
        <div>c</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c := unpack3(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(2))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_a.Box().Width, Fl(10))
	tu.AssertEqual(t, div_b.Box().Width, Fl(10))
	tu.AssertEqual(t, div_c.Box().Width, Fl(10))

	for _, child := range article.Box().Children {
		tu.AssertEqual(t, child.Box().Height, Fl(2))
	}
	tu.AssertEqual(t, article.Box().Width, Fl(10))
}

func TestGridTemplateFrUndefinedFreeSpace(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-rows: 1fr 1fr;
          grid-template-columns: 1fr 1fr;
          line-height: 1;
          width: 10px;
        }
      </style>
      <article>
        <div>a</div> <div>b<br>b<br>b<br>b</div>
        <div>c</div> <div>d</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d := unpack4(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(5))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(5))
	tu.AssertEqual(t, div_a.Box().Height, Fl(8))
	tu.AssertEqual(t, div_c.Box().Height, Fl(8))
	tu.AssertEqual(t, div_b.Box().Height, Fl(8))
	tu.AssertEqual(t, div_d.Box().Height, Fl(8))
	tu.AssertEqual(t, div_a.Box().Width, Fl(5))
	tu.AssertEqual(t, div_c.Box().Width, Fl(5))
	tu.AssertEqual(t, div_b.Box().Width, Fl(5))
	tu.AssertEqual(t, div_d.Box().Width, Fl(5))
	tu.AssertEqual(t, article.Box().Width, Fl(10))
	tu.AssertEqual(t, article.Box().Height, Fl(16))
}

func TestGridColumnStart(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        dl {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-columns: max-content auto;
          line-height: 1;
          width: 10px;
        }
        dt {
          display: block;
          grid-column-start: 1;
        }
        dd {
          display: block;
          grid-column-start: 2;
        }
      </style>
      <dl>
        <dt>A</dt>
        <dd>A1</dd>
        <dd>A2</dd>
        <dt>B</dt>
        <dd>B1</dd>
        <dd>B2</dd>
      </dl>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	dl := body.Box().Children[0]
	dt_a, dd_a1, dd_a2, dt_b, dd_b1, dd_b2 := unpack6(dl)
	tu.AssertEqual(t, dt_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, dd_a1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, dd_a2.Box().PositionY, Fl(2))
	tu.AssertEqual(t, dt_b.Box().PositionY, Fl(4))
	tu.AssertEqual(t, dd_b1.Box().PositionY, Fl(4))
	tu.AssertEqual(t, dd_b2.Box().PositionY, Fl(6))
	tu.AssertEqual(t, dt_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, dt_b.Box().PositionX, Fl(0))
	tu.AssertEqual(t, dd_a1.Box().PositionX, Fl(2))
	tu.AssertEqual(t, dd_a2.Box().PositionX, Fl(2))
	tu.AssertEqual(t, dd_b1.Box().PositionX, Fl(2))
	tu.AssertEqual(t, dd_b2.Box().PositionX, Fl(2))
}

func TestGridColumnStartBlockified(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        dl {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-columns: max-content auto;
          line-height: 1;
          width: 10px;
        }
        dt {
          display: inline;
          grid-column-start: 1;
        }
        dd {
          display: inline;
          grid-column-start: 2;
        }
      </style>
      <dl>
        <dt>A</dt>
        <dd>A1</dd>
        <dd>A2</dd>
        <dt>B</dt>
        <dd>B1</dd>
        <dd>B2</dd>
      </dl>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	dl := body.Box().Children[0]
	dt_a, dd_a1, dd_a2, dt_b, dd_b1, dd_b2 := unpack6(dl)
	tu.AssertEqual(t, dt_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, dd_a1.Box().PositionY, Fl(0))
	tu.AssertEqual(t, dd_a2.Box().PositionY, Fl(2))
	tu.AssertEqual(t, dt_b.Box().PositionY, Fl(4))
	tu.AssertEqual(t, dd_b1.Box().PositionY, Fl(4))
	tu.AssertEqual(t, dd_b2.Box().PositionY, Fl(6))
	tu.AssertEqual(t, dt_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, dt_b.Box().PositionX, Fl(0))
	tu.AssertEqual(t, dd_a1.Box().PositionX, Fl(2))
	tu.AssertEqual(t, dd_a2.Box().PositionX, Fl(2))
	tu.AssertEqual(t, dd_b1.Box().PositionX, Fl(2))
	tu.AssertEqual(t, dd_b2.Box().PositionX, Fl(2))
}

func TestGridUndefinedFreeSpace(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body {
          font-family: weasyprint;
          font-size: 2px;
          line-height: 1;
        }
        .columns {
          display: grid;
          grid-template-columns: 1fr 1fr;
          width: 8px;
        }
        .rows {
          display: grid;
          grid-template-rows: 1fr 1fr;
        }
      </style>
      <div class="columns">
        <div class="rows">
          <div>aa</div>
          <div>b</div>
        </div>
        <div class="rows">
          <div>c<br>c</div>
          <div>d</div>
        </div>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div_c := body.Box().Children[0]
	div_c1, div_c2 := unpack2(div_c)
	div_r11, div_r12 := unpack2(div_c1)
	div_r21, div_r22 := unpack2(div_c2)
	tu.AssertEqual(t, div_r11.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_r12.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_r21.Box().PositionX, Fl(4))
	tu.AssertEqual(t, div_r22.Box().PositionX, Fl(4))
	tu.AssertEqual(t, div_r11.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_r21.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_r12.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_r22.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_r11.Box().Height, Fl(4))
	tu.AssertEqual(t, div_r21.Box().Height, Fl(4))
	tu.AssertEqual(t, div_r12.Box().Height, Fl(4))
	tu.AssertEqual(t, div_r22.Box().Height, Fl(4))
	tu.AssertEqual(t, div_r11.Box().Width, Fl(4))
	tu.AssertEqual(t, div_r21.Box().Width, Fl(4))
	tu.AssertEqual(t, div_r12.Box().Width, Fl(4))
	tu.AssertEqual(t, div_r22.Box().Width, Fl(4))
	tu.AssertEqual(t, div_c.Box().Width, Fl(8))
}

func TestGridPadding(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-rows: auto 1fr;
          grid-template-columns: auto 1fr;
          line-height: 1;
          width: 14px;
        }
      </style>
      <article>
        <div style="padding: 1px">a</div> <div>b</div>
        <div>c</div> <div style="padding: 2px">d</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d := unpack4(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().ContentBoxX(), Fl(0))
	tu.AssertEqual(t, div_a.Box().ContentBoxX(), Fl(1))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(4))
	tu.AssertEqual(t, div_b.Box().ContentBoxX(), Fl(4))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(4))
	tu.AssertEqual(t, div_d.Box().ContentBoxX(), Fl(6))
	tu.AssertEqual(t, div_a.Box().Width, Fl(2))
	tu.AssertEqual(t, div_b.Box().Width, Fl(10))
	tu.AssertEqual(t, div_c.Box().Width, Fl(4))
	tu.AssertEqual(t, div_d.Box().Width, Fl(6))
	tu.AssertEqual(t, article.Box().Width, Fl(14))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().ContentBoxY(), Fl(0))
	tu.AssertEqual(t, div_a.Box().ContentBoxY(), Fl(1))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_c.Box().ContentBoxY(), Fl(4))
	tu.AssertEqual(t, div_d.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_d.Box().ContentBoxY(), Fl(6))
	tu.AssertEqual(t, div_a.Box().Height, Fl(2))
	tu.AssertEqual(t, div_d.Box().Height, Fl(2))
	tu.AssertEqual(t, div_b.Box().Height, Fl(4))
	tu.AssertEqual(t, div_c.Box().Height, Fl(6))
	tu.AssertEqual(t, article.Box().Height, Fl(10))
}

func TestGridBorder(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-rows: auto 1fr;
          grid-template-columns: auto 1fr;
          line-height: 1;
          width: 14px;
        }
      </style>
      <article>
        <div style="border: 1px solid">a</div> <div>b</div>
        <div>c</div> <div style="border: 2px solid">d</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d := unpack4(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PaddingBoxX(), Fl(0))
	tu.AssertEqual(t, div_a.Box().PaddingBoxX(), Fl(1))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(4))
	tu.AssertEqual(t, div_b.Box().PaddingBoxX(), Fl(4))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(4))
	tu.AssertEqual(t, div_d.Box().PaddingBoxX(), Fl(6))
	tu.AssertEqual(t, div_a.Box().Width, Fl(2))
	tu.AssertEqual(t, div_b.Box().Width, Fl(10))
	tu.AssertEqual(t, div_c.Box().Width, Fl(4))
	tu.AssertEqual(t, div_d.Box().Width, Fl(6))
	tu.AssertEqual(t, article.Box().Width, Fl(14))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PaddingBoxY(), Fl(0))
	tu.AssertEqual(t, div_a.Box().PaddingBoxY(), Fl(1))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_c.Box().PaddingBoxY(), Fl(4))
	tu.AssertEqual(t, div_d.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_d.Box().PaddingBoxY(), Fl(6))
	tu.AssertEqual(t, div_a.Box().Height, Fl(2))
	tu.AssertEqual(t, div_d.Box().Height, Fl(2))
	tu.AssertEqual(t, div_b.Box().Height, Fl(4))
	tu.AssertEqual(t, div_c.Box().Height, Fl(6))
	tu.AssertEqual(t, article.Box().Height, Fl(10))
}

func TestGridMargin(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article {
          display: grid;
          font-family: weasyprint;
          font-size: 2px;
          grid-template-rows: auto 1fr;
          grid-template-columns: auto 1fr;
          line-height: 1;
          width: 14px;
        }
      </style>
      <article>
        <div style="margin: 1px">a</div> <div>b</div>
        <div>c</div> <div style="margin: 2px">d</div>
      </article>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	article := body.Box().Children[0]
	div_a, div_b, div_c, div_d := unpack4(article)
	tu.AssertEqual(t, div_a.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().PositionX, Fl(0))
	tu.AssertEqual(t, div_c.Box().BorderBoxX(), Fl(0))
	tu.AssertEqual(t, div_a.Box().BorderBoxX(), Fl(1))
	tu.AssertEqual(t, div_b.Box().PositionX, Fl(4))
	tu.AssertEqual(t, div_b.Box().BorderBoxX(), Fl(4))
	tu.AssertEqual(t, div_d.Box().PositionX, Fl(4))
	tu.AssertEqual(t, div_d.Box().BorderBoxX(), Fl(6))
	tu.AssertEqual(t, div_a.Box().Width, Fl(2))
	tu.AssertEqual(t, div_b.Box().Width, Fl(10))
	tu.AssertEqual(t, div_c.Box().Width, Fl(4))
	tu.AssertEqual(t, div_d.Box().Width, Fl(6))
	tu.AssertEqual(t, article.Box().Width, Fl(14))
	tu.AssertEqual(t, div_a.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().PositionY, Fl(0))
	tu.AssertEqual(t, div_b.Box().BorderBoxY(), Fl(0))
	tu.AssertEqual(t, div_a.Box().BorderBoxY(), Fl(1))
	tu.AssertEqual(t, div_c.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_c.Box().BorderBoxY(), Fl(4))
	tu.AssertEqual(t, div_d.Box().PositionY, Fl(4))
	tu.AssertEqual(t, div_d.Box().BorderBoxY(), Fl(6))
	tu.AssertEqual(t, div_a.Box().Height, Fl(2))
	tu.AssertEqual(t, div_d.Box().Height, Fl(2))
	tu.AssertEqual(t, div_b.Box().Height, Fl(4))
	tu.AssertEqual(t, div_c.Box().Height, Fl(6))
	tu.AssertEqual(t, article.Box().Height, Fl(10))
}
