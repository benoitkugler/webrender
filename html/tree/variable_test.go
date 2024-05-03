package tree

// Test CSS custom properties, also known as CSS variables.

import (
	"fmt"
	"strings"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// parse a simple html with style and an element and return
// the computed style for this element
func setupVar(t *testing.T, html string) (htmlS, elementS pr.ElementStyle) {
	page, err := newHtml(utils.InputString(html))
	if err != nil {
		t.Fatal(err)
	}

	styleFor := GetAllComputedStyles(page, nil, false, nil, nil, nil, nil, false, nil)
	htmlNode := page.Root
	elementNode := htmlNode.FirstChild.NextSibling.FirstChild

	htmlS = styleFor.Get((*utils.HTMLNode)(htmlNode), "")
	elementS = styleFor.Get((*utils.HTMLNode)(elementNode), "")
	return
}

func TestVariableSimple(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	_, style := setupVar(t, `
      <style>
        p { --var: 10px; width: var(--var); color: red }
      </style>
      <p></p>
    `)

	tu.AssertEqual(t, style.GetWidth(), pr.FToPx(10))
}

func TestVariableNotComputed(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	_, style := setupVar(t, `
	<style>
	p { --var: 1rem; width: var(--var) }
      </style>
      <p></p>
	  `)
	tu.AssertEqual(t, style.GetWidth(), pr.FToPx(16))
}

func TestVariableInherit(t *testing.T) {
	_, style := setupVar(t, `
      <style>
        html { --var: 10px }
        p { width: var(--var) }
      </style>
      <p></p>
    `)
	tu.AssertEqual(t, style.GetWidth(), pr.FToPx(10))
}

func TestVariableInheritOverride(t *testing.T) {
	_, style := setupVar(t, `
      <style>
        html { --var: 20px }
        p { width: var(--var); --var: 10px }
      </style>
      <p></p>
    `)
	tu.AssertEqual(t, style.GetWidth(), pr.FToPx(10))
}

func TestVariableDefaultUnknown(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	_, style := setupVar(t, `
      <style>
        p { width: var(--x, 10px) }
      </style>
      <p></p>
    `)
	tu.AssertEqual(t, style.GetWidth(), pr.FToPx(10))
}

func TestVariableDefaultVar(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	_, style := setupVar(t, `
      <style>
        p { --var: 10px; width: var(--x, var(--var)) }
      </style>
      <p></p>
    `)
	tu.AssertEqual(t, style.GetWidth(), pr.FToPx(10))
}

func TestVariableCaseSensitive1(t *testing.T) {
	_, style := setupVar(t, `
      <style>
        html { --VAR: 20px }
        p { width: var(--VAR) }
      </style>
      <p></p>
    `)
	tu.AssertEqual(t, style.GetWidth(), pr.FToPx(20))
}

func TestVariableCaseSensitive2(t *testing.T) {
	_, style := setupVar(t, `
      <style>
        html { --var: 20px }
        body { --VAR: 10px }
        p { width: var(--VAR) }
      </style>
      <p></p>
    `)
	exp := pr.FToPx(10)
	tu.AssertEqual(t, style.GetWidth(), exp)
}

func TestVariableChain(t *testing.T) {
	_, style := setupVar(t, `
      <style>
        html { --foo: 10px }
        body { --var: var(--foo) }
        p { width: var(--var) }
      </style>
      <p></p>
    `)
	exp := pr.FToPx(10)
	tu.AssertEqual(t, style.GetWidth(), exp)
}

func TestVariableChainRoot(t *testing.T) {
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1656
	style, _ := setupVar(t, `
      <style>
        html { --var2: 10px; --var1: var(--var2); width: var(--var1) }
      </style>
    `)
	exp := pr.FToPx(10)
	tu.AssertEqual(t, style.GetWidth(), exp)
}

func TestVariableSelf(t *testing.T) {
	_, _ = setupVar(t, `
      <style>
        html { --var1: var(--var1) }
      </style>
    `)
}

func TestVariableLoop(t *testing.T) {
	_, _ = setupVar(t, `
      <style>
        html { --var1: var(--var2); --var2: var(--var1) }
      </style>
    `)
}

func TestVariableChainRootMissing(t *testing.T) {
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1656
	_, _ = setupVar(t, `
      <style>
        html { --var1: var(--var-missing); width: var(--var1) }
      </style>
    `)
}

func TestVariablePartial1(t *testing.T) {
	_, style := setupVar(t, `
      <style>
        html { --var: 10px }
        div { margin: 0 0 0 var(--var) }
      </style>
      <div></div>
    `)
	exp0, exp10 := pr.FToPx(0), pr.FToPx(10)
	if got := style.GetMarginTop(); got != exp0 {
		t.Fatalf("expected %v, got %v", exp0, got)
	}
	if got := style.GetMarginRight(); got != exp0 {
		t.Fatalf("expected %v, got %v", exp0, got)
	}
	if got := style.GetMarginBottom(); got != exp0 {
		t.Fatalf("expected %v, got %v", exp0, got)
	}
	if got := style.GetMarginLeft(); got != exp10 {
		t.Fatalf("expected %v, got %v", exp10, got)
	}
}

func TestVariableShorthandMarginMultiple(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	_, div := setupVar(t, `
      <style>
        html { --var1: 10px; --var2: 20px }
        div { margin: var(--var2) 0 0 var(--var1) }
      </style>
      <div></div>
    `)
	tu.AssertEqual(t, div.GetMarginTop(), pr.FToPx(20))
	tu.AssertEqual(t, div.GetMarginRight(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetMarginBottom(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetMarginLeft(), pr.FToPx(10))
}

func TestVariableShorthandMarginInvalid(t *testing.T) {
	logs := tu.CaptureLogs()
	_, div := setupVar(t, `
          <style>
            html { --var: blue }
            div { margin: 0 0 0 var(--var) }
          </style>
          <div></div>
        `)
	tu.AssertEqual(t, len(logs.Logs()), 1)
	tu.AssertEqual(t, strings.Contains(logs.Logs()[0], "invalid value"), true)

	tu.AssertEqual(t, div.GetMarginTop(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetMarginRight(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetMarginBottom(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetMarginLeft(), pr.FToPx(0))
}

func TestVariableShorthandBorder(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	_, div := setupVar(t, `
      <style>
        html { --var: 1px solid blue }
        div { border: var(--var) }
      </style>
      <div></div>
    `)
	tu.AssertEqual(t, div.GetBorderTopWidth(), pr.FToPx(1))
	tu.AssertEqual(t, div.GetBorderRightWidth(), pr.FToPx(1))
	tu.AssertEqual(t, div.GetBorderBottomWidth(), pr.FToPx(1))
	tu.AssertEqual(t, div.GetBorderLeftWidth(), pr.FToPx(1))
}

func TestVariableShorthandBorderSide(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	_, div := setupVar(t, `
      <style>
        html { --var: 1px solid blue }
        div { border-top: var(--var) }
      </style>
      <div></div>
    `)
	tu.AssertEqual(t, div.GetBorderTopWidth(), pr.FToPx(1))
	tu.AssertEqual(t, div.GetBorderRightWidth(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetBorderBottomWidth(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetBorderLeftWidth(), pr.FToPx(0))
}

func TestVariableShorthandBorderMixed(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	_, div := setupVar(t, `
      <style>
        html { --var: 1px solid }
        div { border: blue var(--var) }
      </style>
      <div></div>
    `)
	tu.AssertEqual(t, div.GetBorderTopWidth(), pr.FToPx(1))
	tu.AssertEqual(t, div.GetBorderRightWidth(), pr.FToPx(1))
	tu.AssertEqual(t, div.GetBorderBottomWidth(), pr.FToPx(1))
	tu.AssertEqual(t, div.GetBorderLeftWidth(), pr.FToPx(1))
}

func TestVariableShorthandBorderMixedInvalid(t *testing.T) {
	logs := tu.CaptureLogs()
	_, div := setupVar(t, `
          <style>
            html { --var: 1px solid blue }
            div { border: blue var(--var) }
          </style>
          <div></div>
        `)
	// TODO: we should only get one warning here
	tu.AssertEqual(t, strings.Contains(logs.Logs()[0], "multiple color values"), true)
	tu.AssertEqual(t, div.GetBorderTopWidth(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetBorderRightWidth(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetBorderBottomWidth(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetBorderLeftWidth(), pr.FToPx(0))
}

func TestVariableShorthandBackground(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, test := range []struct {
		var_       string
		background string
	}{
		{"blue", "var(--v)"},
		{"padding-box url(pattern.png)", "var(--v)"},
		{"padding-box url(pattern.png)", "white var(--v) center"},
		{"100%", "url(pattern.png) var(--v) var(--v) / var(--v) var(--v)"},
		{"left / 100%", "url(pattern.png) top var(--v) 100%"},
	} {
		_, _ = setupVar(t, fmt.Sprintf(`
		  <style>
			html { --v: %s }
			div { background: %s }
		  </style>
		  <div></div>
		`, test.var_, test.background))
	}
}

func TestVariableShorthandBackgroundInvalid(t *testing.T) {
	for _, test := range []struct {
		var_       string
		background string
	}{
		{"invalid", "var(--v)"},
		{"blue", "var(--v) var(--v)"},
		{"100%", "url(pattern.png) var(--v) var(--v) var(--v)"},
	} {
		logs := tu.CaptureLogs()
		_, _ = setupVar(t, fmt.Sprintf(`
			  <style>
				html { --v: %s }
				div { background: %s }
			  </style>
			  <div></div>
			`, test.var_, test.background))

		tu.AssertEqual(t, len(logs.Logs()), 1)
		tu.AssertEqual(t, strings.Contains(logs.Logs()[0], "invalid value"), true)
	}
}

func TestVariableInitial(t *testing.T) {
	_, style := setupVar(t, `
      <style>
        html { --var: initial }
        p { width: var(--var, 10px) }
      </style>
      <p></p>
    `)
	tu.AssertEqual(t, style.GetWidth(), pr.FToPx(10))
}

func TestVariableInitialDefault(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/2075
	html, style := setupVar(t, `
      <style>
        p { --var: initial; width: var(--var, 10px) }
      </style>
      <p></p>
    `)
	tu.AssertEqual(t, html.GetWidth(), style.GetWidth())
}

func TestVariableInitialDefaultVar(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/2075
	html, style := setupVar(t, `
      <style>
        p { --var: initial; width: var(--var, var(--var)) }
      </style>
      <p></p>
    `)
	tu.AssertEqual(t, html.GetWidth(), style.GetWidth())
}

func TestVariableFallback(t *testing.T) {
	for prop := range pr.KnownProperties {
		_, style := setupVar(t, fmt.Sprintf(`
		  <style>
			div {
			  --var: improperValue;
			  %s: var(--var);
			}
		  </style>
		  <div></div>
		`, prop))
		_ = style.Get(prop.Key()) // just check for crashes
	}
}

// TODO:

// func TestVariableListContent(t *testing.T) {
// 	defer tu.CaptureLogs().AssertNoLogs(t)
//     // Regression test for https://github.com/Kozea/WeasyPrint/issues/1287
//     _, style := setupVar(t,`
//       <style>
//         :root { --var: "Page " counter(page) "/" counter(pages) }
//         div::before { content: var(--var) }
//       </style>
//       <div></div>
//     `)
//     html, = page.children
//     body, = html.children
//     div, = body.children
//     line, = div.children
//     before, = line.children
//     text, = before.children
//     assert text.text == "Page 1/1"
// }

// 	@pytest.mark.parametrize("var, display", (
// 		("inline", "var(--var)"),
// 		("inline-block", "var(--var)"),
// 		("inline flow", "var(--var)"),
// 		("inline", "var(--var) flow"),
// 		("flow", "inline var(--var)"),
// 		))

// func TestVariableListDisplay(t *testing.Tvar, display) {
// 	defer tu.CaptureLogs().AssertNoLogs(t)
//     _, style := setupVar(t,`
//       <style>
//         html { --var: %s }
//         div { display: %s }
//       </style>
//       <section><div></div></section>
//     ` % (var, display))
//     html, = page.children
//     body, = html.children
//     section, = body.children
//     child, = section.children
//     assert type(child).__name__ == "LineBox"

// 	@pytest.mark.parametrize("var, font", (
// 		("weasyprint", "var(--var)"),
// 		(""weasyprint"", "var(--var)"),
// 		("weasyprint", "var(--var), monospace"),
// 		("weasyprint, monospace", "var(--var)"),
// 		("monospace", "weasyprint, var(--var)"),
// 		))

// func TestVariableListFont(t *testing.Tvar, font) {
// 	defer tu.CaptureLogs().AssertNoLogs(t)
//     _, style := setupVar(t,`
//       <style>
//         @font-face {src: url(weasyprint.otf); font-family: weasyprint}
//         html { font-size: 2px; --var: %s }
//         div { font-family: %s }
//       </style>
//       <div>aa</div>
//     ` % (var, font))
//     html, = page.children
//     body, = html.children
//     div, = body.children
//     line, = div.children
//     text, = line.children
//     assert text.width == 4

// func TestVariableInFunction(t *testing.T) {
// 	defer tu.CaptureLogs().AssertNoLogs(t)
//     _, style := setupVar(t,`
//       <style>
//         html { --var: title }
//         h1 { counter-increment: var(--var) }
//         div::before { content: counter(var(--var)) }
//       </style>
//       <section>
//         <h1></h1>
//         <div></div>
//         <h1></h1>
//         <div></div>
//       </section>
//     `)
//     html, = page.children
//     body, = html.children
//     section, = body.children
//     h11, div1, h12, div2 = section.children
//     assert div1.children[0].children[0].children[0].text == "1"
//     assert div2.children[0].children[0].children[0].text == "2"

// func TestVariableInFunctionMultipleValues(t *testing.T) {
// 	defer tu.CaptureLogs().AssertNoLogs(t)
//     _, style := setupVar(t,`
//       <style>
//         html { --name: title; --counter: title, upper-roman }
//         h1 { counter-increment: var(--name) }
//         div::before { content: counter(var(--counter)) }
//       </style>
//       <section>
//         <h1></h1>
//         <div></div>
//         <h1></h1>
//         <div></div>
//         <h1></h1>
//         <div style="--counter: var(--name), lower-roman"></div>
//       </section>
//     `)
//     html, = page.children
//     body, = html.children
//     section, = body.children
//     h11, div1, h12, div2, h13, div3 = section.children
//     assert div1.children[0].children[0].children[0].text == 'I'
//     assert div2.children[0].children[0].children[0].text == 'II'
//     assert div3.children[0].children[0].children[0].text == 'iii'

// func TestVariableInVariableInFunction(t *testing.T) {
// 	defer tu.CaptureLogs().AssertNoLogs(t)
//     _, style := setupVar(t,`
//       <style>
//         html { --name: title; --counter: var(--name), upper-roman }
//         h1 { counter-increment: var(--name) }
//         div::before { content: counter(var(--counter)) }
//       </style>
//       <section>
//         <h1></h1>
//         <div></div>
//         <h1></h1>
//         <div></div>
//         <h1></h1>
//         <div style="--counter: var(--name), lower-roman"></div>
//       </section>
//     `)
//     html, = page.children
//     body, = html.children
//     section, = body.children
//     h11, div1, h12, div2, h13, div3 = section.children
//     assert div1.children[0].children[0].children[0].text == 'I'
//     assert div2.children[0].children[0].children[0].text == 'II'
//     assert div3.children[0].children[0].children[0].text == 'iii'

// func TestVariableInFunctionMissing(t *testing.T) {
//     with capture_logs() as logs:
//         _, style := setupVar(t,`
//           <style>
//             h1 { counter-increment: var(--var) }
//             div::before { content: counter(var(--var)) }
//           </style>
//           <section>
//             <h1></h1>
//             <div></div>
//             <h1></h1>
//             <div></div>
//           </section>
//         `)
//         assert len(logs) == 2
//         assert 'no value' in logs[0]
//         assert 'invalid value' in logs[1]
//     html, = page.children
//     body, = html.children
//     section, = body.children
//     h11, div1, h12, div2 = section.children
//     assert not div1.children
//     assert not div2.children

// func TestVariableInFunctionInVariable(t *testing.T) {
// 	defer tu.CaptureLogs().AssertNoLogs(t)
//     _, style := setupVar(t,`
//       <style>
//         html { --name: title; --counter: counter(var(--name), upper-roman) }
//         h1 { counter-increment: var(--name) }
//         div::before { content: var(--counter) }
//       </style>
//       <section>
//         <h1></h1>
//         <div></div>
//         <h1></h1>
//         <div></div>
//         <h1></h1>
//         <div style="--counter: counter(var(--name), lower-roman)"></div>
//       </section>
//     `)
//     html, = page.children
//     body, = html.children
//     section, = body.children
//     h11, div1, h12, div2, h13, div3 = section.children
//     assert div1.children[0].children[0].children[0].text == 'I'
//     assert div2.children[0].children[0].children[0].text == 'II'
//     assert div3.children[0].children[0].children[0].text == 'iii'
