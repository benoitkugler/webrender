package tree

// Test CSS custom proproperties, also known as CSS variables.

import (
	"fmt"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils"
)

// parse a simple html with style and an element and return
// the computed style for this element
func setupVar(t *testing.T, html string) pr.ElementStyle {
	page, err := newHtml(utils.InputString(html))
	if err != nil {
		t.Fatal(err)
	}

	styleFor := GetAllComputedStyles(page, nil, false, nil, nil, nil, nil, nil)
	p := page.Root.FirstChild.NextSibling.FirstChild
	return styleFor.Get((*utils.HTMLNode)(p), "")
}

func TestVariableSimple(t *testing.T) {
	style := setupVar(t, `
      <style>
        p { --var: 10px; width: var(--var); color: red }
      </style>
      <p></p>
    `)

	exp := pr.FToPx(10)
	if got := style.GetWidth(); got != exp {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}

func TestVariableInherit(t *testing.T) {
	style := setupVar(t, `
      <style>
        html { --var: 10px }
        p { width: var(--var) }
      </style>
      <p></p>
    `)
	exp := pr.FToPx(10)
	if got := style.GetWidth(); got != exp {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}

func TestVariableInheritOverride(t *testing.T) {
	style := setupVar(t, `
      <style>
        html { --var: 20px }
        p { width: var(--var); --var: 10px }
      </style>
      <p></p>
    `)
	exp := pr.FToPx(10)
	if got := style.GetWidth(); got != exp {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}

func TestVariableCaseSensitive1(t *testing.T) {
	style := setupVar(t, `
      <style>
        html { --VAR: 20px }
        p { width: var(--VAR) }
      </style>
      <p></p>
    `)
	exp := pr.FToPx(20)
	if got := style.GetWidth(); got != exp {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}

func TestVariableCaseSensitive2(t *testing.T) {
	style := setupVar(t, `
      <style>
        html { --var: 20px }
        body { --VAR: 10px }
        p { width: var(--VAR) }
      </style>
      <p></p>
    `)
	exp := pr.FToPx(10)
	if got := style.GetWidth(); got != exp {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}

func TestVariableChain(t *testing.T) {
	style := setupVar(t, `
      <style>
        html { --foo: 10px }
        body { --var: var(--foo) }
        p { width: var(--var) }
      </style>
      <p></p>
    `)
	exp := pr.FToPx(10)
	if got := style.GetWidth(); got != exp {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}

func TestVariablePartial1(t *testing.T) {
	style := setupVar(t, `
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
	style := setupVar(t, `
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
		style := setupVar(t, fmt.Sprintf(`
		  <style>
			div {
			  --var: improperValue;
			  %s: var(--var);
			}
		  </style>
		  <div></div>
		`, prop))
		_ = style.Get(prop) // just check for crashes
	}
}
