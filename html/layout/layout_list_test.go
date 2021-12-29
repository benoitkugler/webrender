package layout

import (
	"fmt"
	"testing"

	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for lists layout.

func TestListsStyle(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, inside := range []string{"inside", ""} {
		for _, sc := range [][2]string{
			{"circle", "◦ "},
			{"disc", "• "},
			{"square", "▪ "},
		} {
			style, character := sc[0], sc[1]
			page := renderOnePage(t, fmt.Sprintf(`
			<style>
				body { margin: 0 }
				ul { margin-left: 50px; list-style: %s %s }
			</style>
			<ul>
				<li>abc</li>
			</ul>
			`, inside, style))
			html := page.Box().Children[0]
			body := html.Box().Children[0]
			unorderedList := body.Box().Children[0]
			listItem := unorderedList.Box().Children[0]
			var content, markerText Box
			if inside != "" {
				var marker Box
				line := listItem.Box().Children[0]
				marker, content = unpack2(line)
				markerText = marker.Box().Children[0]
			} else {
				marker, lineContainer := unpack2(listItem)
				tu.AssertEqual(t, marker.Box().PositionX, listItem.Box().PositionX, "marker")
				tu.AssertEqual(t, marker.Box().PositionY, listItem.Box().PositionY, "marker")
				line := lineContainer.Box().Children[0]
				content = line.Box().Children[0]
				markerLine := marker.Box().Children[0]
				markerText = markerLine.Box().Children[0]
			}
			tu.AssertEqual(t, markerText.(*bo.TextBox).Text, character, "markerText")
			tu.AssertEqual(t, content.(*bo.TextBox).Text, "abc", "content")
		}
	}
}

func TestListsEmptyItem(t *testing.T) {
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/873
	page := renderOnePage(t, `
      <ul>
        <li>a</li>
        <li></li>
        <li>a</li>
      </ul>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	unorderedList := body.Box().Children[0]
	li1, li2, li3 := unpack3(unorderedList)
	tu.AssertEqual(t, li1.Box().PositionY != li2.Box().PositionY, true, "li1")
	tu.AssertEqual(t, li2.Box().PositionY != li3.Box().PositionY, true, "li1")
}

// @pytest.mark.xfail
// func TestListsWhitespaceItem(t *testing.T ) {
//     // Regression test for https://github.com/Kozea/WeasyPrint/issues/873
//     page := renderOnePage(t, `
//       <ul>
//         <li>a</li>
//         <li> </li>
//         <li>a</li>
//       </ul>
//     `)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     unorderedList := body.Box().Children[0]
//     li1, li2, li3 = unorderedList.Box().Children
//     tu.AssertEqual(t, li1.Box().PositionY != li2.Box().PositionY != li3.Box().PositionY, "li1")

func TestListsPageBreak(t *testing.T) {
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/945
	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 300px 100px }
        ul { font-size: 30px; font-family: weasyprint; margin: 0 }
      </style>
      <ul>
        <li>a</li>
        <li>a</li>
        <li>a</li>
        <li>a</li>
      </ul>
    `)
	page1, page2 := pages[0], pages[1]
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	ul := body.Box().Children[0]
	tu.AssertEqual(t, len(ul.Box().Children), 3, "len")
	for _, li := range ul.Box().Children {
		tu.AssertEqual(t, len(li.Box().Children), 2, "len")
	}

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	ul = body.Box().Children[0]
	tu.AssertEqual(t, len(ul.Box().Children), 1, "len")
	for _, li := range ul.Box().Children {
		tu.AssertEqual(t, len(li.Box().Children), 2, "len")
	}
}
