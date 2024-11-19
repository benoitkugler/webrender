package layout

import (
	"fmt"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/text/hyphen"
	"github.com/benoitkugler/webrender/utils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func TestErrorRecovery(t *testing.T) {
	for _, style := range []string{
		`<style> html { color red; color: blue; color`,
		`<html style="color; color: blue; color red">`,
	} {
		capt := tu.CaptureLogs()
		page := renderOnePage(t, style)
		html := unpack1(page)
		tu.AssertEqual(t, html.Box().Style.GetColor(), pr.NewColor(0, 0, 1, 1)) // blue
		tu.AssertEqual(t, len(capt.Logs()), 2)
	}
}

type textContext struct {
	struts map[text.StrutLayoutKey][2]pr.Float
}

func (tc textContext) Fonts() text.FontConfiguration                          { return fontconfig }
func (tc textContext) HyphenCache() map[text.HyphenDictKey]hyphen.Hyphener    { return nil }
func (tc textContext) StrutLayoutsCache() map[text.StrutLayoutKey][2]pr.Float { return tc.struts }

func newTextContext() textContext {
	return textContext{struts: make(map[text.StrutLayoutKey][2]pr.Float)}
}

func TestLineHeightInheritance(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
		<style>
		html { font-size: 10px; line-height: 140% }
		section { font-size: 10px; line-height: 1.4 }
		div, p { font-size: 20px; vertical-align: 50% }
		</style>
		<body><div><section><p></p></section></div></body>
	`)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	section := unpack1(div)
	paragraph := unpack1(section)
	tu.AssertEqual(t, html.Box().Style.GetFontSize(), pr.FToV(10))
	tu.AssertEqual(t, div.Box().Style.GetFontSize(), pr.FToV(20))
	// 140% of 10px = 14px is inherited from html
	tu.AssertEqual(t, text.StrutLayout(div.Box().Style, newTextContext())[0], Fl(14))
	tu.AssertEqual(t, div.Box().Style.GetVerticalAlign(), pr.FToV(7)) // 50 % of 14px

	tu.AssertEqual(t, paragraph.Box().Style.GetFontSize(), pr.FToV(20))
	// 1.4 is inherited from p, 1.4 * 20px on em = 28px
	tu.AssertEqual(t, text.StrutLayout(paragraph.Box().Style, newTextContext())[0], Fl(28))
	tu.AssertEqual(t, paragraph.Box().Style.GetVerticalAlign(), pr.FToV(14)) // 50% of 28px,
}

func TestImportant(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	htmlContent := `
	<style>
			p:nth-child(1) { color: lime }
		body p:nth-child(2) { color: red }
	
		p:nth-child(3) { color: lime !important }
		body p:nth-child(3) { color: red }

		body p:nth-child(5) { color: lime }
		p:nth-child(5) { color: red }

		p:nth-child(6) { color: red }
		p:nth-child(6) { color: lime }
	</style>
	<p></p>
	<p></p>
	<p></p>
	<p></p>
	<p></p>
	<p></p>
	`

	style, err := tree.NewCSSDefault(utils.InputString(`
	body p:nth-child(1) { color: red }
	p:nth-child(2) { color: lime !important }

	p:nth-child(4) { color: lime !important }
	body p:nth-child(4) { color: red }
	`))
	if err != nil {
		t.Fatal(err)
	}

	pages := renderPages(t, htmlContent, style)

	page := pages[0]
	html := unpack1(page)
	body := unpack1(html)
	for _, paragraph := range body.Box().Children {
		tu.AssertEqual(t, paragraph.Box().Style.GetColor(), pr.NewColor(0, 1, 0, 1)) // lime (light green)
	}
}

func TestNamedPages(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
	<style>
	@page NARRow { size: landscape }
	div { page: AUTO }
	p { page: NARRow }
	</style>
	<div><p><span>a</span></p></div>
		`)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	p := unpack1(div)
	span := unpack1(p)
	tu.AssertEqual(t, html.Box().Style.GetPage(), pr.Page(""))
	tu.AssertEqual(t, body.Box().Style.GetPage(), pr.Page(""))
	tu.AssertEqual(t, div.Box().Style.GetPage(), pr.Page(""))
	tu.AssertEqual(t, p.Box().Style.GetPage(), pr.Page("NARRow"))
	tu.AssertEqual(t, span.Box().Style.GetPage(), pr.Page("NARRow"))
}

func TestUnits(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range []struct {
		value string
		width pr.Float
	}{
		{"96px", 96},
		{"1in", 96},
		{"72pt", 96},
		{"6pc", 96},
		{"2.54cm", 96},
		{"25.4mm", 96},
		{"101.6q", 96},
		{"1.1em", 11},
		{"1.1rem", 17.6},
		{"1.1ch", 11},
		{"1.5ex", 12},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
		<style>@font-face { src: url(AHEM____.TTF); font-family: ahem }</style>
		<body style="font: 10px ahem"><p style="margin-left: %s"></p>`, data.value))
		html := unpack1(page)
		body := unpack1(html)
		p := unpack1(body)
		tu.AssertEqual(t, p.Box().MarginLeft, data.width)
	}
}

func TestMediaQueries(t *testing.T) {
	for _, data := range []struct {
		media   string
		width   pr.Float
		warning bool
	}{
		{`@media screen { @page { size: 10px } }`, 20, false},
		{`@media print { @page { size: 10px } }`, 10, false},
		{`@media ("unknown content") { @page { size: 10px } }`, 20, true},
	} {
		logs := tu.CaptureLogs()

		style, err := tree.NewCSSDefault(utils.InputString(fmt.Sprintf("@page{size:20px}%s", data.media)))
		if err != nil {
			t.Fatal(err)
		}

		pages := renderPages(t, "<p>a<span>b", style)
		page := pages[0]
		html := unpack1(page)
		tu.AssertEqual(t, html.Box().Width, data.width)
		if !data.warning {
			logs.AssertNoLogs(t)
		}
	}
}

func TestNestingBlock(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, style := range [...]string{
		"div { p { width: 10px } }",
		"p { div & { width: 10px } }",
		"p { width: 20px; div & { width: 10px } }",
		"p { div & { width: 10px } width: 20px }",
		"div { & { & { p { & { width: 10px } } } } }",
		"@media print { div { p { width: 10px } } }",
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>%s</style>
      <div><p></p></div><p></p>
    `, style))
		html := unpack1(page)
		body := unpack1(html)
		div, p := unpack2(body)
		div_p := unpack1(div)
		tu.AssertEqual(t, div_p.Box().Width, pr.Float(10))
		tu.AssertEqual(t, p.Box().Width != pr.Float(10), true)
	}
}
