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
        p { --var: 10px; width: var(--var); }
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
	// Regression test for #1656.
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
        html { --var1: var(--var2); --var2: var(--var1); padding: var(--var1) }
      </style>
    `)
}

func TestVariableChainRootMissing(t *testing.T) {
	// Regression test for #1656.
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

func TestVariableShorthandMargin(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	_, div := setupVar(t, `
      <style>
        html { --var: 10px }
        div { margin: 0 0 0 var(--var) }
      </style>
      <div></div>
    `)
	tu.AssertEqual(t, div.GetMarginTop(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetMarginRight(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetMarginBottom(), pr.FToPx(0))
	tu.AssertEqual(t, div.GetMarginLeft(), pr.FToPx(10))
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
	_ = div.GetMarginBottom()
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
	tu.AssertEqual(t, div.GetBorderTopWidth(), pr.FToV(1))
	tu.AssertEqual(t, div.GetBorderRightWidth(), pr.FToV(1))
	tu.AssertEqual(t, div.GetBorderBottomWidth(), pr.FToV(1))
	tu.AssertEqual(t, div.GetBorderLeftWidth(), pr.FToV(1))
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
	tu.AssertEqual(t, div.GetBorderTopWidth(), pr.FToV(1))
	tu.AssertEqual(t, div.GetBorderRightWidth(), pr.FToV(0))
	tu.AssertEqual(t, div.GetBorderBottomWidth(), pr.FToV(0))
	tu.AssertEqual(t, div.GetBorderLeftWidth(), pr.FToV(0))
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
	tu.AssertEqual(t, div.GetBorderTopWidth(), pr.FToV(1))
	tu.AssertEqual(t, div.GetBorderRightWidth(), pr.FToV(1))
	tu.AssertEqual(t, div.GetBorderBottomWidth(), pr.FToV(1))
	tu.AssertEqual(t, div.GetBorderLeftWidth(), pr.FToV(1))
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
	// trigger eval
	_ = div.GetBorderTopWidth()
	tu.AssertEqual(t, len(logs.Logs()), 2)
	tu.AssertEqual(t, strings.Contains(logs.Logs()[0], "multiple border-top-color values"), true)
	tu.AssertEqual(t, div.GetBorderTopWidth(), pr.FToV(0))
	tu.AssertEqual(t, div.GetBorderRightWidth(), pr.FToV(0))
	tu.AssertEqual(t, div.GetBorderBottomWidth(), pr.FToV(0))
	tu.AssertEqual(t, div.GetBorderLeftWidth(), pr.FToV(0))
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
		_, div := setupVar(t, fmt.Sprintf(`
			  <style>
				html { --v: %s }
				div { background: %s }
			  </style>
			  <div></div>
			`, test.var_, test.background))
		_ = div.GetBackgroundColor()
		tu.AssertEqual(t, len(logs.Logs()), 1)
		// tu.AssertEqual(t, strings.Contains(logs.Logs()[0], "invalid"), true)
	}
}

func TestVariableInitial(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for #2075.
	html, p := setupVar(t, `
      <style>
        html { --var: initial }
        p { width: var(--var) }
      </style>
      <p></p>
    `)
	tu.AssertEqual(t, html.GetWidth(), p.GetWidth())
}

func TestVariableInitialDefault(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for #2075.
	html, p := setupVar(t, `
	<style>
	p { --var: initial; width: var(--var, 10px) }
	</style>
	<p></p>
	`)
	tu.AssertEqual(t, html.GetWidth(), p.GetWidth())
	// tu.AssertEqual(t, style.GetWidth(), pr.FToPx(10))
}

func TestVariableInitialDefaultVar(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	// Regression test for #2075.
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
