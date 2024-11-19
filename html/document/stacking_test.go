package document

import (
	"testing"

	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/layout"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/utils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Test CSS stacking contexts.

var baseUrl, _ = utils.PathToURL("../../resources_test/")

// lay out a document and return a list of PageBox objects
func renderPages(t *testing.T, htmlContent string) []*bo.PageBox {
	doc, err := tree.NewHTML(utils.InputString(htmlContent), baseUrl, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	doc.UAStyleSheet = tree.TestUAStylesheet // fakeHTML
	return layout.Layout(doc, nil, false, fc)
}

type serializedStacking struct {
	tag           string
	blockAndCells []string
	zeroZs        []serializedStacking
}

func serializeStacking(context StackingContext) serializedStacking {
	out := serializedStacking{
		tag: context.box.Box().ElementTag(),
	}
	for _, b := range context.blocksAndCells {
		out.blockAndCells = append(out.blockAndCells, b.Box().ElementTag())
	}
	for _, c := range context.zeroZContexts {
		out.zeroZs = append(out.zeroZs, serializeStacking(c))
	}
	return out
}

func TestNested(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, data := range []struct {
		source   string
		contexts serializedStacking
	}{
		{
			`
      <p id=lorem></p>
      <div style="position: relative">
        <p id=lipsum></p>
      </div>`,
			serializedStacking{
				"html", []string{"body", "p"}, []serializedStacking{
					{"div", []string{"p"}, nil},
				},
			},
		},
		{
			`
      <div style="position: relative">
        <p style="position: relative"></p>
      </div>`,
			serializedStacking{
				"html", []string{"body"}, []serializedStacking{
					{"div", nil, nil},
					{"p", nil, nil},
				},
			},
		},
	} {
		page := renderPages(t, data.source)[0]
		html := page.Box().Children[0]
		tu.AssertEqual(t, serializeStacking(NewStackingContextFromBox(html, page, nil)), data.contexts)
	}
}

func TestImageContexts(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderPages(t, `
      <body>Some text: <img style="position: relative" src=pattern.png>`)[0]
	html := page.Box().Children[0]
	context := NewStackingContextFromBox(html, page, nil)
	// The image is *not* := range this context:
	tu.AssertEqual(t, bo.Serialize([]Box{context.box}), []bo.SerBox{
		{
			Tag: "html", Type: bo.BlockT, Content: bo.BC{
				C: []bo.SerBox{
					{
						Tag: "body", Type: bo.BlockT, Content: bo.BC{
							C: []bo.SerBox{
								{
									Tag: "body", Type: bo.LineT, Content: bo.BC{
										C: []bo.SerBox{
											{Tag: "body", Type: bo.TextT, Content: bo.BC{Text: "Some text: "}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	// ... but in a sub-context {
	var boxes []bo.Box
	for _, c := range context.zeroZContexts {
		boxes = append(boxes, c.box)
	}
	got := bo.Serialize(boxes)
	tu.AssertEqual(t, got, []bo.SerBox{
		{Tag: "img", Type: bo.InlineReplacedT, Content: bo.BC{Text: "<replaced>"}},
	})
}
