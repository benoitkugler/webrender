package selector

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func MustParseHTML(doc string) *html.Node {
	dom, err := html.Parse(strings.NewReader(doc))
	if err != nil {
		panic(err)
	}
	return dom
}

var (
	selector = MustCompile(`div.matched`)
	doc      = `<!DOCTYPE html>
<html>
<body>
<div class="matched">
  <div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
    <div class="matched"></div>
  </div>
</div>
</body>
</html>
`
)
var dom = MustParseHTML(doc)

func BenchmarkMatchAll(b *testing.B) {
	var matches []*html.Node
	for i := 0; i < b.N; i++ {
		matches = MatchAll(dom, selector)
	}
	_ = matches
}

func BenchmarkMatchAllW3(b *testing.B) {
	tests := loadValidSelectors(b)
	doc := parseReference("test_resources/content.xhtml")
	var allSelectors []Sel
	for _, test := range tests {
		if test.Xfail {
			continue
		}
		sels, err := ParseGroup(test.Selector)
		if err != nil {
			b.Fatalf("%s -> unable to parse valid selector : %s : %s", test.Name, test.Selector, err)
		}
		for _, sel := range sels {
			if sel.PseudoElement() != "" {
				continue // pseudo element doesn't count as a match in this test since they are not part of the document
			}
			allSelectors = append(allSelectors, sel)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, sel := range allSelectors {
			_ = MatchAll(doc, sel)
		}
	}
}

func TestMatchAllW3(t *testing.T) {
	tests := loadValidSelectors(t)
	doc := parseReference("test_resources/content.xhtml")
	var allSelectors []Sel
	for _, test := range tests {
		if test.Xfail {
			continue
		}
		sels, err := ParseGroup(test.Selector)
		if err != nil {
			t.Fatalf("%s -> unable to parse valid selector : %s : %s", test.Name, test.Selector, err)
		}
		for _, sel := range sels {
			if sel.PseudoElement() != "" {
				continue // pseudo element doesn't count as a match in this test since they are not part of the document
			}
			allSelectors = append(allSelectors, sel)
		}
	}

	for _, sel := range allSelectors {
		_ = MatchAll(doc, sel)
	}
}
