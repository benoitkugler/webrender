package layout

import (
	"fmt"
	"strings"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func TestVariableListContent(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1287
	page := renderOnePage(t, `
      <style>
        :root { --var: "Page " counter(page) "/" counter(pages) }
        div::before { content: var(--var) }
      </style>
      <div></div>
    `)
	html := unpack1(page)
	body := unpack1(html)
	div := unpack1(body)
	line := unpack1(div)
	before := unpack1(line)
	text := unpack1(before)
	assertText(t, text, "Page 1/1")
}

func TestVariableListDisplay(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, test := range []struct {
		name  string
		value string
	}{
		{"inline", "var(--var)"},
		{"inline-block", "var(--var)"},
		{"inline flow", "var(--var)"},
		{"inline", "var(--var) flow"},
		{"flow", "inline var(--var)"},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        html { --var: %s }
        div { display: %s }
      </style>
      <section><div></div></section>
    `, test.name, test.value))
		html := unpack1(page)
		body := unpack1(html)
		section := unpack1(body)
		child := unpack1(section)
		tu.AssertEqual(t, child.Type(), boxes.LineT)
	}
}

func TestVariableListFont(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, test := range []struct {
		name  string
		value string
	}{
		{`weasyprint`, "var(--var)"},
		{`"weasyprint"`, "var(--var)"},
		{`weasyprint`, "var(--var), monospace"},
		{`weasyprint, monospace`, "var(--var)"},
		{`monospace`, "weasyprint, var(--var)"},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        html { font-size: 2px; --var: %s }
        div { font-family: %s }
      </style>
      <div>aa</div>
    `, test.name, test.value))
		html := unpack1(page)
		body := unpack1(html)
		div := unpack1(body)
		line := unpack1(div)
		text := unpack1(line)
		tu.AssertEqual(t, text.Box().Width, pr.Float(4))
	}
}

func TestVariableInFunction(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        html { --var: title }
        h1 { counter-increment: var(--var) }
        div::before { content: counter(var(--var)) }
      </style>
      <section>
        <h1></h1>
        <div></div>
        <h1></h1>
        <div></div>
      </section>
    `)
	html := unpack1(page)
	body := unpack1(html)
	section := unpack1(body)
	_, div1, _, div2 := unpack4(section)
	assertText(t, unpack1(unpack1(unpack1(div1))), "1")
	assertText(t, unpack1(unpack1(unpack1(div2))), "2")
}

func TestVariableInFunctionMultipleValues(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        html { --name: title; --counter: title, upper-roman }
        h1 { counter-increment: var(--name) }
        div::before { content: counter(var(--counter)) }
      </style>
      <section>
        <h1></h1>
        <div></div>
        <h1></h1>
        <div></div>
        <h1></h1>
        <div style="--counter: var(--name), lower-roman"></div>
      </section>
    `)
	html := unpack1(page)
	body := unpack1(html)
	section := unpack1(body)
	_, div1, _, div2, _, div3 := unpack6(section)
	assertText(t, unpack1(unpack1(unpack1(div1))), "I")
	assertText(t, unpack1(unpack1(unpack1(div2))), "II")
	assertText(t, unpack1(unpack1(unpack1(div3))), "iii")
}

func TestVariableInVariableInFunction(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        html { --name: title; --counter: var(--name), upper-roman }
        h1 { counter-increment: var(--name) }
        div::before { content: counter(var(--counter)) }
      </style>
      <section>
        <h1></h1>
        <div></div>
        <h1></h1>
        <div></div>
        <h1></h1>
        <div style="--counter: var(--name), lower-roman"></div>
      </section>
    `)
	html := unpack1(page)
	body := unpack1(html)
	section := unpack1(body)
	_, div1, _, div2, _, div3 := unpack6(section)
	assertText(t, unpack1(unpack1(unpack1(div1))), "I")
	assertText(t, unpack1(unpack1(unpack1(div2))), "II")
	assertText(t, unpack1(unpack1(unpack1(div3))), "iii")
}

func TestVariableInFunctionMissing(t *testing.T) {
	capt := tu.CaptureLogs()
	page := renderOnePage(t, `
          <style>
            h1 { counter-increment: var(--var) }
            div::before { content: counter(var(--var)) }
          </style>
          <section>
            <h1></h1>
            <div></div>
            <h1></h1>
            <div></div>
          </section>
        `)
	gotL := capt.Logs()
	tu.AssertEqual(t, len(gotL), 4)
	tu.AssertEqual(t, strings.Contains(gotL[0], "no value"), true)
	tu.AssertEqual(t, strings.Contains(gotL[1], "invalid value"), true)
	html := unpack1(page)
	body := unpack1(html)
	section := unpack1(body)
	_, div1, _, div2 := unpack4(section)
	tu.AssertEqual(t, len(div1.Box().Children), 0)
	tu.AssertEqual(t, len(div2.Box().Children), 0)
}

func TestVariableInFunctionInVariable(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	page := renderOnePage(t, `
      <style>
        html { --name: title; --counter: counter(var(--name), upper-roman) }
        h1 { counter-increment: var(--name) }
        div::before { content: var(--counter) }
      </style>
      <section>
        <h1></h1>
        <div></div>
        <h1></h1>
        <div></div>
        <h1></h1>
        <div style="--counter: counter(var(--name), lower-roman)"></div>
      </section>
    `)
	html := unpack1(page)
	body := unpack1(html)
	section := unpack1(body)
	_, div1, _, div2, _, div3 := unpack6(section)
	assertText(t, unpack1(unpack1(unpack1(div1))), "I")
	assertText(t, unpack1(unpack1(unpack1(div2))), "II")
	assertText(t, unpack1(unpack1(unpack1(div3))), "iii")
}
