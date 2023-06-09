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
	_, style := setupVar(t, `
      <style>
        p { --var: 10px; width: var(--var); color: red }
      </style>
      <p></p>
    `)

	exp := pr.FToPx(10)
	tu.AssertEqual(t, style.GetWidth(), exp, "")
}

func TestVariableInherit(t *testing.T) {
	_, style := setupVar(t, `
      <style>
        html { --var: 10px }
        p { width: var(--var) }
      </style>
      <p></p>
    `)
	exp := pr.FToPx(10)
	tu.AssertEqual(t, style.GetWidth(), exp, "")
}

func TestVariableInheritOverride(t *testing.T) {
	_, style := setupVar(t, `
      <style>
        html { --var: 20px }
        p { width: var(--var); --var: 10px }
      </style>
      <p></p>
    `)
	exp := pr.FToPx(10)
	tu.AssertEqual(t, style.GetWidth(), exp, "")
}

func TestVariableCaseSensitive1(t *testing.T) {
	_, style := setupVar(t, `
      <style>
        html { --VAR: 20px }
        p { width: var(--VAR) }
      </style>
      <p></p>
    `)
	exp := pr.FToPx(20)
	tu.AssertEqual(t, style.GetWidth(), exp, "")
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
	tu.AssertEqual(t, style.GetWidth(), exp, "")
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
	tu.AssertEqual(t, style.GetWidth(), exp, "")
}

func TestVariableChainRoot(t *testing.T) {
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1656
	style, _ := setupVar(t, `
      <style>
        html { --var2: 10px; --var1: var(--var2); width: var(--var1) }
      </style>
    `)
	exp := pr.FToPx(10)
	tu.AssertEqual(t, style.GetWidth(), exp, "")
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

func TestVariableInitial(t *testing.T) {
	_, style := setupVar(t, `
      <style>
        html { --var: initial }
        p { width: var(--var, 10px) }
      </style>
      <p></p>
    `)
	exp := pr.FToPx(10)
	if got := style.GetWidth(); got != exp {
		t.Fatalf("expected %v, got %v", exp, got)
	}
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
		_ = style.Get(strings.ReplaceAll(prop, "-", "_")) // just check for crashes
	}
}
