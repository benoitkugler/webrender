package document

import (
	"reflect"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/utils"
	"github.com/benoitkugler/webrender/utils/testutils"
)

func round(x fl) fl { return utils.RoundPrec(x, 4) }

// Eliminate errors of floating point arithmetic for metadata.
func roundMeta(pages []Page) {
	for _, page := range pages {
		anc := page.anchors
		for anchorName, v := range anc {
			anc[anchorName] = [2]fl{round(v[0]), round(v[1])}
		}
		links := page.links
		for i, link := range links {
			r := link.Rectangle
			link.Rectangle = [4]fl{round(r[0]), round(r[1]), round(r[2]), round(r[3])}
			links[i] = link
		}
		bookmarks := page.bookmarks
		for i, v := range bookmarks {
			pos := v.position
			v.position = [2]fl{round(pos[0]), round(pos[1])}
			bookmarks[i] = v
		}
	}
}

func TestBookmarks(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range []struct {
		html           string
		expectedByPage [][]bookmarkData
		expectedTree   []backend.BookmarkNode
		round          bool
	}{
		{
			`
		        <style>* { height: 10px }</style>
		        <h1>a</h1>
		        <h4 style="page-break-after: always">b</h4>
		        <h3 style="position: relative; top: 2px; left: 3px">c</h3>
		        <h2>d</h2>
		        <h1>e</h1>
		    `,
			[][]bookmarkData{
				{{level: 1, label: "a", position: [2]fl{0, 0}, open: true}, {level: 4, label: "b", position: [2]fl{0, 10}, open: true}},
				{{level: 3, label: "c", position: [2]fl{3, 2}, open: true}, {level: 2, label: "d", position: [2]fl{0, 10}, open: true}, {level: 1, label: "e", position: [2]fl{0, 20}, open: true}},
			},
			[]backend.BookmarkNode{
				{Label: "a", PageIndex: 0, X: 0, Y: 0, Children: []backend.BookmarkNode{
					{Label: "b", PageIndex: 0, X: 0, Y: 10, Open: true},
					{Label: "c", PageIndex: 1, X: 3, Y: 2, Open: true},
					{Label: "d", PageIndex: 1, X: 0, Y: 10, Open: true},
				}, Open: true},
				{Label: "e", PageIndex: 1, X: 0, Y: 20, Open: true},
			},
			false,
		},
		{
			`
		        <style>
		            * { height: 90px; margin: 0 0 10px 0 }
		        </style>
		        <h1>Title 1</h1>
		        <h1>Title 2</h1>
		        <h2 style="position: relative; left: 20px">Title 3</h2>
		        <h2>Title 4</h2>
		        <h3>Title 5</h3>
		        <span style="display: block; page-break-before: always"></span>
		        <h2>Title 6</h2>
		        <h1>Title 7</h1>
		        <h2>Title 8</h2>
		        <h3>Title 9</h3>
		        <h1>Title 10</h1>
		        <h2>Title 11</h2>
		    `,
			[][]bookmarkData{
				{
					{level: 1, label: "Title 1", position: [2]fl{0, 0}, open: true},
					{level: 1, label: "Title 2", position: [2]fl{0, 100}, open: true},
					{level: 2, label: "Title 3", position: [2]fl{20, 200}, open: true},
					{level: 2, label: "Title 4", position: [2]fl{0, 300}, open: true},
					{level: 3, label: "Title 5", position: [2]fl{0, 400}, open: true},
				},
				{
					{level: 2, label: "Title 6", position: [2]fl{0, 100}, open: true},
					{level: 1, label: "Title 7", position: [2]fl{0, 200}, open: true},
					{level: 2, label: "Title 8", position: [2]fl{0, 300}, open: true},
					{level: 3, label: "Title 9", position: [2]fl{0, 400}, open: true},
					{level: 1, label: "Title 10", position: [2]fl{0, 500}, open: true},
					{level: 2, label: "Title 11", position: [2]fl{0, 600}, open: true},
				},
			},
			[]backend.BookmarkNode{
				{Label: "Title 1", PageIndex: 0, X: 0, Y: 0, Open: true},
				{Label: "Title 2", PageIndex: 0, X: 0, Y: 100, Children: []backend.BookmarkNode{
					{Label: "Title 3", PageIndex: 0, X: 20, Y: 200, Open: true},
					{Label: "Title 4", PageIndex: 0, X: 0, Y: 300, Children: []backend.BookmarkNode{
						{Label: "Title 5", PageIndex: 0, X: 0, Y: 400, Open: true},
					}, Open: true},
					{Label: "Title 6", PageIndex: 1, X: 0, Y: 100, Open: true},
				}, Open: true},
				{Label: "Title 7", PageIndex: 1, X: 0, Y: 200, Children: []backend.BookmarkNode{
					{Label: "Title 8", PageIndex: 1, X: 0, Y: 300, Children: []backend.BookmarkNode{
						{Label: "Title 9", PageIndex: 1, X: 0, Y: 400, Open: true},
					}, Open: true},
				}, Open: true},
				{Label: "Title 10", PageIndex: 1, X: 0, Y: 500, Children: []backend.BookmarkNode{
					{Label: "Title 11", PageIndex: 1, X: 0, Y: 600, Open: true},
				}, Open: true},
			},
			false,
		},
		{
			`
		        <style>* { height: 10px }</style>
		        <h2>A</h2> <p>depth 1</p>
		        <h4>B</h4> <p>depth 2</p>
		        <h2>C</h2> <p>depth 1</p>
		        <h3>D</h3> <p>depth 2</p>
		        <h4>E</h4> <p>depth 3</p>
		    `,
			[][]bookmarkData{{
				{level: 2, label: "A", position: [2]fl{0, 0}, open: true},
				{level: 4, label: "B", position: [2]fl{0, 20}, open: true},
				{level: 2, label: "C", position: [2]fl{0, 40}, open: true},
				{level: 3, label: "D", position: [2]fl{0, 60}, open: true},
				{level: 4, label: "E", position: [2]fl{0, 80}, open: true},
			}},
			[]backend.BookmarkNode{
				{Label: "A", PageIndex: 0, X: 0, Y: 0, Children: []backend.BookmarkNode{
					{Label: "B", PageIndex: 0, X: 0, Y: 20, Open: true},
				}, Open: true},
				{Label: "C", PageIndex: 0, X: 0, Y: 40, Children: []backend.BookmarkNode{
					{Label: "D", PageIndex: 0, X: 0, Y: 60, Children: []backend.BookmarkNode{
						{Label: "E", PageIndex: 0, X: 0, Y: 80, Open: true},
					}, Open: true},
				}, Open: true},
			},
			false,
		},
		{
			`
		        <style>* { height: 10px; font-size: 0 }</style>
		        <h2>A</h2> <p>h2 depth 1</p>
		        <h4>B</h4> <p>h4 depth 2</p>
		        <h3>C</h3> <p>h3 depth 2</p>
		        <h5>D</h5> <p>h5 depth 3</p>
		        <h1>E</h1> <p>h1 depth 1</p>
		        <h2>F</h2> <p>h2 depth 2</p>
		        <h2>G</h2> <p>h2 depth 2</p>
		        <h4>H</h4> <p>h4 depth 3</p>
		        <h1>I</h1> <p>h1 depth 1</p>
		    `,
			[][]bookmarkData{{
				{level: 2, label: "A", position: [2]fl{0, 0}, open: true},
				{level: 4, label: "B", position: [2]fl{0, 20}, open: true},
				{level: 3, label: "C", position: [2]fl{0, 40}, open: true},
				{level: 5, label: "D", position: [2]fl{0, 60}, open: true},
				{level: 1, label: "E", position: [2]fl{0, 70}, open: true},
				{level: 2, label: "F", position: [2]fl{0, 90}, open: true},
				{level: 2, label: "G", position: [2]fl{0, 110}, open: true},
				{level: 4, label: "H", position: [2]fl{0, 130}, open: true},
				{level: 1, label: "I", position: [2]fl{0, 150}, open: true},
			}},
			[]backend.BookmarkNode{
				{Label: "A", PageIndex: 0, X: 0, Y: 0, Children: []backend.BookmarkNode{
					{Label: "B", PageIndex: 0, X: 0, Y: 20, Open: true},
					{Label: "C", PageIndex: 0, X: 0, Y: 40, Children: []backend.BookmarkNode{
						{Label: "D", PageIndex: 0, X: 0, Y: 60, Open: true},
					}, Open: true},
				}, Open: true},
				{Label: "E", PageIndex: 0, X: 0, Y: 70, Children: []backend.BookmarkNode{
					{Label: "F", PageIndex: 0, X: 0, Y: 90, Open: true},
					{Label: "G", PageIndex: 0, X: 0, Y: 110, Children: []backend.BookmarkNode{
						{Label: "H", PageIndex: 0, X: 0, Y: 130, Open: true},
					}, Open: true},
				}, Open: true},
				{Label: "I", PageIndex: 0, X: 0, Y: 150, Open: true},
			},
			false,
		},
		{
			"<h1>é",
			[][]bookmarkData{
				{{level: 1, label: "é", position: [2]fl{0, 0}, open: true}},
			},
			[]backend.BookmarkNode{
				{Label: "é", PageIndex: 0, X: 0, Y: 0, Open: true},
			},
			false,
		},
		{
			`
		    <h1 style="transform: translateX(50px)">!
		`,
			[][]bookmarkData{
				{{level: 1, label: "!", position: [2]fl{50, 0}, open: true}},
			},
			[]backend.BookmarkNode{
				{Label: "!", PageIndex: 0, X: 50, Y: 0, Open: true},
			},
			false,
		},

		{
			`
		    <style>
		      img { display: block; bookmark-label: attr(alt); bookmark-level: 1 }
		    </style>
		    <img src="pattern.png" alt="Chocolate" />
		`,
			[][]bookmarkData{{
				{level: 1, label: "Chocolate", position: [2]fl{0, 0}, open: true},
			}},
			[]backend.BookmarkNode{
				{Label: "Chocolate", PageIndex: 0, X: 0, Y: 0, Open: true},
			},
			false,
		},

		{
			`
		        <h1 style="transform-origin: 0 0;
		                   transform: rotate(90deg) translateX(50px)">!
		    `,
			[][]bookmarkData{
				{
					{level: 1, label: "!", position: [2]fl{0, 50}, open: true},
				},
			},
			[]backend.BookmarkNode{{Label: "!", PageIndex: 0, X: 0, Y: 50, Open: true}},
			true,
		},
		{
			`
		    <body style="transform-origin: 0 0; transform: rotate(90deg)">
		    <h1 style="transform: translateX(50px)">!
		`,
			[][]bookmarkData{{{level: 1, label: "!", position: [2]fl{0, 50}, open: true}}},
			[]backend.BookmarkNode{{Label: "!", PageIndex: 0, X: 0, Y: 50, Open: true}},
			true,
		},
	} {
		assertBookmarks(t, data.html, data.expectedByPage, data.expectedTree, data.round)
	}
}

func renderHTML(t *testing.T, html string, baseUrl string, round bool) Document {
	doc, err := tree.NewHTML(utils.InputString(html), baseUrl, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	doc.UAStyleSheet = tree.TestUAStylesheet // fakeHTML

	document := Render(doc, nil, false, fc)
	if round {
		roundMeta(document.Pages)
	}
	return document
}

func assertBookmarks(t *testing.T, html string, expectedByPage [][]bookmarkData, expectedTree []backend.BookmarkNode, round bool) {
	t.Helper()

	document := renderHTML(t, html, baseUrl, round)

	var gotByPage [][]bookmarkData
	for _, page := range document.Pages {
		gotByPage = append(gotByPage, page.bookmarks)
	}
	if !reflect.DeepEqual(gotByPage, expectedByPage) {
		t.Fatalf("html %s : unexpected bookmark per page: %v \n %v", html, gotByPage, expectedByPage)
	}
	if got := document.makeBookmarkTree(); !reflect.DeepEqual(got, expectedTree) {
		t.Fatalf("unexpected bookmark tree %v", got)
	}
}

func TestLinks(t *testing.T) {
	// capt := testutils.CaptureLogs()
	// defer capt.AssertNoLogs(t)

	// baseUrl=resourceFilename("<inline HTML>"),
	// warnings=(), round=false
	assertLinks := func(html string, expectedLinksByPage [][]Link, expectedAnchorsByPage []anchors,
		expectedResolvedLinks [][]Link, expectedResolvedAnchors [][]backend.Anchor, baseUrl string, warnings []string, round bool) {
		t.Helper()

		capt := testutils.CaptureLogs()

		document := renderHTML(t, html, baseUrl, round)
		resolvedLinks, resolvedAnchors := document.resolveLinks()

		logs := capt.Logs()
		if len(logs) != len(warnings) {
			t.Fatalf("unexpected number of logs: %d", len(logs))
		}
		for i, expected := range warnings {
			if !strings.Contains(logs[i], expected) {
				t.Fatalf("invalid log: %s", logs[i])
			}
		}
		var (
			gotLinksByPage   [][]Link
			gotAnchorsByPage []map[string][2]fl
		)
		for _, p := range document.Pages {
			gotLinksByPage = append(gotLinksByPage, p.links)
			gotAnchorsByPage = append(gotAnchorsByPage, p.anchors)
		}
		if !reflect.DeepEqual(gotLinksByPage, expectedLinksByPage) {
			t.Fatalf("unexpected gotLinksByPage: %v", gotLinksByPage)
		}
		if !reflect.DeepEqual(gotAnchorsByPage, expectedAnchorsByPage) {
			t.Fatalf("unexpected gotAnchorsByPage: %v", gotAnchorsByPage)
		}
		if !reflect.DeepEqual(resolvedLinks, expectedResolvedLinks) {
			t.Fatalf("unexpected resolvedLinks: %v", resolvedLinks)
		}
		if !reflect.DeepEqual(resolvedAnchors, expectedResolvedAnchors) {
			t.Fatalf("unexpected resolvedAnchors: %v", resolvedAnchors)
		}
	}

	assertLinks(`
	    <style>
	        body { font-size: 10px; line-height: 2; width: 200px }
	        p { height: 90px; margin: 0 0 10px 0 }
	        img { width: 30px; vertical-align: top }
	    </style>
	    <p><a href="http://weasyprint.org"><img src=pattern.png></a></p>
	    <p style="padding: 0 10px"><a
	        href="#lipsum"><img style="border: solid 1px"
	                            src=pattern.png></a></p>
	    <p id=hello>Hello, World</p>
	    <p id=lipsum>
	        <a style="display: block; page-break-before: always; height: 30px"
	           href="#hel%6Co"></a>
	    </p>
	`, [][]Link{
		{
			{Type: "external", Target: "http://weasyprint.org", Rectangle: [4]fl{0, 0, 30, 20}},
			{Type: "external", Target: "http://weasyprint.org", Rectangle: [4]fl{0, 0, 30, 30}},
			{Type: "internal", Target: "lipsum", Rectangle: [4]fl{10, 100, 42, 120}},
			{Type: "internal", Target: "lipsum", Rectangle: [4]fl{10, 100, 42, 132}},
		},
		{
			{Type: "internal", Target: "hello", Rectangle: [4]fl{0, 0, 200, 30}},
		},
	},
		[]anchors{
			{"hello": [2]fl{0, 200}},
			{"lipsum": [2]fl{0, 0}},
		},
		[][]Link{
			{
				{Type: "external", Target: "http://weasyprint.org", Rectangle: [4]fl{0, 0, 30, 20}},
				{Type: "external", Target: "http://weasyprint.org", Rectangle: [4]fl{0, 0, 30, 30}},
				{Type: "internal", Target: "lipsum", Rectangle: [4]fl{10, 100, 42, 120}},
				{Type: "internal", Target: "lipsum", Rectangle: [4]fl{10, 100, 42, 132}},
			},
			{
				{Type: "internal", Target: "hello", Rectangle: [4]fl{0, 0, 200, 30}},
			},
		},
		[][]backend.Anchor{
			{
				{Name: "hello", X: 0, Y: 200},
			},
			{
				{Name: "lipsum", X: 0, Y: 0},
			},
		},
		baseUrl, nil, false)

	assertLinks(
		`
		        <body style="width: 200px">
		        <a href="../lipsum/é%E9" style="display: block; margin: 10px 5px">
		    `, [][]Link{{
			{Type: "external", Target: "http://weasyprint.org/foo/lipsum/%C3%A9%E9", Rectangle: [4]fl{5, 10, 195, 10}},
		}},
		[]anchors{{}},
		[][]Link{{
			{Type: "external", Target: "http://weasyprint.org/foo/lipsum/%C3%A9%E9", Rectangle: [4]fl{5, 10, 195, 10}},
		}},
		[][]backend.Anchor{nil},
		"http://weasyprint.org/foo/bar/", nil, false)

	assertLinks(
		`
		        <body style="width: 200px">
		        <div style="display: block; margin: 10px 5px;
		                    -weasy-link: url(../lipsum/é%E9)">
		    `,
		[][]Link{{
			{Type: "external", Target: "http://weasyprint.org/foo/lipsum/%C3%A9%E9", Rectangle: [4]fl{5, 10, 195, 10}},
		}},
		[]anchors{{}},
		[][]Link{{
			{Type: "external", Target: "http://weasyprint.org/foo/lipsum/%C3%A9%E9", Rectangle: [4]fl{5, 10, 195, 10}},
		}},
		[][]backend.Anchor{nil},
		"http://weasyprint.org/foo/bar/", nil, false)

	// Relative URI reference without a base URI: allowed for links
	assertLinks(
		`
		        <body style="width: 200px">
		        <a href="../lipsum" style="display: block; margin: 10px 5px">
		    `,
		[][]Link{{
			{Type: "external", Target: "../lipsum", Rectangle: [4]fl{5, 10, 195, 10}},
		}},
		[]anchors{{}},
		[][]Link{{
			{Type: "external", Target: "../lipsum", Rectangle: [4]fl{5, 10, 195, 10}},
		}},
		[][]backend.Anchor{nil},
		"", nil, false)

	// Relative URI reference without a base URI: not supported for -weasy-link
	assertLinks(
		`
		        <body style="width: 200px">
		        <div style="-weasy-link: url(../lipsum);
		                    display: block; margin: 10px 5px">
		    `, [][]Link{nil}, []anchors{{}}, [][]Link{nil}, [][]backend.Anchor{nil},
		"", []string{
			"Ignored `-weasy-link: url(../lipsum)` , Relative URI reference without a base URI",
		}, false)

	// Internal or absolute URI reference without a base URI: OK
	assertLinks(
		`
		        <body style="width: 200px">
		        <a href="#lipsum" id="lipsum"
		            style="display: block; margin: 10px 5px"></a>
		        <a href="http://weasyprint.org/" style="display: block"></a>
		    `, [][]Link{{
			{Type: "internal", Target: "lipsum", Rectangle: [4]fl{5, 10, 195, 10}},
			{Type: "external", Target: "http://weasyprint.org/", Rectangle: [4]fl{0, 10, 200, 10}},
		}},
		[]anchors{{"lipsum": [2]fl{5, 10}}},
		[][]Link{{
			{Type: "internal", Target: "lipsum", Rectangle: [4]fl{5, 10, 195, 10}},
			{Type: "external", Target: "http://weasyprint.org/", Rectangle: [4]fl{0, 10, 200, 10}},
		}},
		[][]backend.Anchor{{
			{Name: "lipsum", X: 5, Y: 10},
		}},
		"", nil, false)

	assertLinks(
		`
		        <body style="width: 200px">
		        <div style="-weasy-link: url(#lipsum);
		                    margin: 10px 5px" id="lipsum">
		    `,
		[][]Link{{
			{Type: "internal", Target: "lipsum", Rectangle: [4]fl{5, 10, 195, 10}},
		}},
		[]anchors{{"lipsum": [2]fl{5, 10}}},
		[][]Link{{
			{Type: "internal", Target: "lipsum", Rectangle: [4]fl{5, 10, 195, 10}},
		}},
		[][]backend.Anchor{{
			{Name: "lipsum", X: 5, Y: 10},
		}},
		"", nil, false)

	assertLinks(
		`
		        <style> a { display: block; height: 15px } </style>
		        <body style="width: 200px">
		            <a href="#lipsum"></a>
		            <a href="#missing" id="lipsum"></a>
		    `,
		[][]Link{{
			{Type: "internal", Target: "lipsum", Rectangle: [4]fl{0, 0, 200, 15}},
			{Type: "internal", Target: "missing", Rectangle: [4]fl{0, 15, 200, 30}},
		}},
		[]anchors{{"lipsum": [2]fl{0, 15}}},
		[][]Link{{
			{Type: "internal", Target: "lipsum", Rectangle: [4]fl{0, 0, 200, 15}},
		}},
		[][]backend.Anchor{{
			{Name: "lipsum", X: 0, Y: 15},
		}},
		"", []string{"No anchor #missing for internal URI reference"}, false)

	assertLinks(
		`
	        <body style="width: 100px; transform: translateY(100px)">
	        <a href="#lipsum" id="lipsum" style="display: block; height: 20px;
	            transform: rotate(90deg) scale(2)">
	    `,
		[][]Link{{
			{Type: "internal", Target: "lipsum", Rectangle: [4]fl{30, 10, 70, 210}},
		}},
		[]anchors{{"lipsum": [2]fl{70, 10}}},
		[][]Link{{
			{Type: "internal", Target: "lipsum", Rectangle: [4]fl{30, 10, 70, 210}},
		}},
		[][]backend.Anchor{{
			{Name: "lipsum", X: 70, Y: 10},
		}},
		"", nil, true)

	// Download for attachment
	assertLinks(
		`
	        <body style="width: 200px">
	        <a rel=attachment href="pattern.png" download="wow.png"
	            style="display: block; margin: 10px 5px">
	    `,
		[][]Link{{
			{Type: "attachment", Target: "pattern.png", Rectangle: [4]fl{5, 10, 195, 10}},
		}},
		[]anchors{{}},
		[][]Link{{
			{Type: "attachment", Target: "pattern.png", Rectangle: [4]fl{5, 10, 195, 10}},
		}},
		[][]backend.Anchor{nil},
		"", nil, false)
}
