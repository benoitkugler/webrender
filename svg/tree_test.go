package svg

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func parseIcon(t *testing.T, iconPath string) {
	f, err := os.Open(iconPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img, err := Parse(f, "", nil)
	if err != nil {
		t.Fatal(iconPath, err)
	}

	img.Draw(outputPage{}, 100, 100) // just check for crashes
}

func corpusFiles() (out []string) {
	for _, p := range []string{
		"beach", "cape", "iceberg", "island",
		"mountains", "sea", "trees", "village",
	} {
		out = append(out, "testdata/landscapeIcons/"+p+".svg")
	}

	for _, p := range []string{
		"astronaut", "jupiter", "lander", "school-bus", "telescope", "content-cut-light", "defs",
		"24px",
	} {
		out = append(out, "testdata/testIcons/"+p+".svg")
	}

	for _, p := range []string{
		"OpacityStrokeDashTest.svg",
		"OpacityStrokeDashTest2.svg",
		"OpacityStrokeDashTest3.svg",
		"TestShapes.svg",
		"TestShapes2.svg",
		"TestShapes3.svg",
		"TestShapes4.svg",
		"TestShapes5.svg",
		"TestShapes6.svg",
		"go-logo-blue.svg",
	} {
		out = append(out, "testdata/"+p)
	}
	return out
}

func TestCorpus(t *testing.T) {
	for _, file := range corpusFiles() {
		parseIcon(t, file)
	}
}

func TestPercentages(t *testing.T) {
	parseIcon(t, "testdata/TestPercentages.svg")
}

func TestInvalidXML(t *testing.T) {
	_, err := Parse(strings.NewReader("dummy"), "", nil)
	if err == nil {
		t.Fatal("expected error on invalid input")
	}
	_, err = Parse(strings.NewReader("<not-svg></not-svg>"), "", nil)
	if err == nil {
		t.Fatal("expected error on invalid input")
	}
}

func TestBuildTree(t *testing.T) {
	input := `
	<svg viewBox="0 0 10 10">
	<style>
		path {
			color: red;
		}
	</style>
	<path style="fontsize: 10px">AA</path>

	</svg>
	`
	root, err := html.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	tree, err := buildSVGTree(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.root.children) != 1 {
		t.Fatalf("unexpected children %v", tree.root.children)
	}
	p := tree.root.children[0]
	if !reflect.DeepEqual(p.attrs, nodeAttributes{"fontsize": "10px", "color": "red"}) {
		t.Fatalf("unexpected attributes %v", p.attrs)
	}
}

func TestParseDefs(t *testing.T) {
	input := `
	<svg viewBox="0 0 10 10" xmlns="http://www.w3.org/2000/svg"
	xmlns:xlink="http://www.w3.org/1999/xlink">
	<!-- Some graphical objects to use -->
	<defs>
		<circle id="myCircle" cx="0" cy="0" r="5" />

		<linearGradient id="myGradient" gradientTransform="rotate(90)">
		<stop offset="20%" stop-color="gold" />
		<stop offset="90%" stop-color="red" />
		</linearGradient>
	</defs>

	<!-- using my graphical objects -->
	<use x="5" y="5" href="#myCircle" fill="url('#myGradient')" />
	</svg>
	`
	img, err := Parse(strings.NewReader(input), "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(img.definitions.nodes) != 1 {
		t.Fatal("defs")
	}
	if c, has := img.definitions.nodes["myCircle"]; !has || len(c.children) != 0 {
		t.Fatal("defs circle")
	}

	if len(img.definitions.paintServers) != 1 {
		t.Fatal("defs")
	}
	if _, has := img.definitions.paintServers["myGradient"]; !has {
		t.Fatal("defs circle")
	}
}

func TestTrefs(t *testing.T) {
	input := `
	<svg width="100%" height="100%" viewBox="0 0 1000 300"
     xmlns="http://www.w3.org/2000/svg"
     xmlns:xlink="http://www.w3.org/1999/xlink">
	<defs>
		<text id="ReferencedText">Referenced character data</text>
	</defs>

	<text x="100" y="100" font-size="45" >
		Inline character data
	</text>

	<text x="100" y="200" font-size="45" fill="red" >
		<tref xlink:href="#ReferencedText"/>
	</text>
	</svg>
	`
	root, err := html.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	img, err := buildSVGTree(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(img.root.children) != 3 {
		t.Fatalf("unexpected children %v", img.root.children)
	}
	if t1 := img.root.children[1]; string(t1.text) != "Inline character data" {
		t.Fatalf("unexpected text %s", t1.text)
	}

	t2 := img.root.children[2]
	if len(t2.children) != 0 {
		t.Fatalf("unexpected children %v", img.root.children)
	}
	if string(t2.text) != "Referenced character data" {
		t.Fatalf("unexpected text %s", t2.text)
	}
}
