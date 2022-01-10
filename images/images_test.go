package images

import (
	"fmt"
	"os"
	"testing"

	"github.com/benoitkugler/webrender/svg"
	"github.com/benoitkugler/webrender/utils"
)

func TestLoadLocalImages(t *testing.T) {
	paths := []string{
		"../resources_test/blue.jpg",
		"../resources_test/icon.png",
		"../resources_test/pattern.gif",
		"../resources_test/pattern.svg",
	}
	for _, path := range paths {
		url, err := utils.PathToURL(path)
		if err != nil {
			t.Fatal(err)
		}
		out, err := getImageFromUri(utils.DefaultUrlFetcher, false, url, "")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%T\n", out)
	}
}

func TestSVGDisplayedSize(t *testing.T) {
	f, err := os.Open("../resources_test/pattern.svg")
	if err != nil {
		t.Fatal(err)
	}
	img, err := svg.Parse(f, "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	w, h := img.DisplayedSize()
	if w != (svg.Value{V: 4, U: svg.Px}) {
		t.Fatalf("unexpected width %v", w)
	}
	if h != (svg.Value{V: 4, U: svg.Px}) {
		t.Fatalf("unexpected height %v", h)
	}
}
