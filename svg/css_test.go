package svg

import (
	"reflect"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/utils"
	"golang.org/x/net/html"
)

func TestParseStyle(t *testing.T) {
	input := `
	<svg width="4cm" height="4cm" viewBox="0 0 400 400"
		xmlns="http://www.w3.org/2000/svg" version="1.1">
		<title>Example triangle01- simple example of a 'path'</title>
		<desc>Testing dashes around a square.</desc>

		<style>css1</style>
		<g>
			<style>css2</style>
		</g>
		<g>
			<style type="invalid">css3</style>
		</g>
		<g>
			<style type="text/css">css4</style>
		</g>
	</svg>
`
	root, err := html.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := fetchStyleAndTextRefs((*utils.HTMLNode)(root))
	if !reflect.DeepEqual(got, [][]byte{
		[]byte("css1"),
		[]byte("css2"),
		[]byte("css4"),
	}) {
		t.Fatalf("unexpected stylesheets %v", got)
	}
}

func TestProcessStyle(t *testing.T) {
	input := `
	<svg width="4cm" height="4cm" viewBox="0 0 400 400"
		xmlns="http://www.w3.org/2000/svg" version="1.1">
		<title>Example triangle01- simple example of a 'path'</title>
		<desc>Testing dashes around a square.</desc>

		<style>
		p {
			color: red;
		}
		</style> 
	</svg>
`

	root, err := html.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := fetchStyleAndTextRefs((*utils.HTMLNode)(root))
	normal, important := parseStylesheets(got, "")
	if len(normal) != 1 {
		t.Fatalf("unexpected normal style: %v", normal)
	}
	if len(important) != 0 {
		t.Fatalf("unexpected important style: %v", important)
	}
}
