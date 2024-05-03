package layout

import (
	"fmt"
	"testing"

	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for lists layout.

func TestListsStyle(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
			html := unpack1(page)
			body := unpack1(html)
			unorderedList := unpack1(body)
			listItem := unpack1(unorderedList)
			var content, markerText Box
			if inside != "" {
				var marker Box
				line := unpack1(listItem)
				marker, content = unpack2(line)
				markerText = unpack1(marker)
			} else {
				marker, lineContainer := unpack2(listItem)
				tu.AssertEqual(t, marker.Box().PositionX, listItem.Box().PositionX)
				tu.AssertEqual(t, marker.Box().PositionY, listItem.Box().PositionY)
				line := unpack1(lineContainer)
				content = unpack1(line)
				markerLine := unpack1(marker)
				markerText = unpack1(markerLine)
			}
			tu.AssertEqual(t, markerText.(*bo.TextBox).Text, character)
			tu.AssertEqual(t, content.(*bo.TextBox).Text, "abc")
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
	html := unpack1(page)
	body := unpack1(html)
	unorderedList := unpack1(body)
	li1, li2, li3 := unpack3(unorderedList)
	tu.AssertEqual(t, li1.Box().PositionY != li2.Box().PositionY, true)
	tu.AssertEqual(t, li2.Box().PositionY != li3.Box().PositionY, true)
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
//     html =  unpack1(page)
//     body :=  unpack1(html)
//     unorderedList := unpack1(body)
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
	html := unpack1(page1)
	body := unpack1(html)
	ul := unpack1(body)
	tu.AssertEqual(t, len(ul.Box().Children), 3)
	for _, li := range ul.Box().Children {
		tu.AssertEqual(t, len(li.Box().Children), 2)
	}

	html = unpack1(page2)
	body = unpack1(html)
	ul = unpack1(body)
	tu.AssertEqual(t, len(ul.Box().Children), 1)
	for _, li := range ul.Box().Children {
		tu.AssertEqual(t, len(li.Box().Children), 2)
	}
}

func TestListsPageBreakMargin(t *testing.T) {
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1058
	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        @page { size: 300px 100px }
        ul { font-size: 30px; font-family: weasyprint; margin: 0 }
        p { margin: 10px 0 }
      </style>
      <ul>
        <li><p>a</p></li>
        <li><p>a</p></li>
        <li><p>a</p></li>
        <li><p>a</p></li>
      </ul>
   `)
	tu.AssertEqual(t, len(pages), 2)
	for _, page := range pages {
		html := unpack1(page)
		body := unpack1(html)
		ul := unpack1(body)
		tu.AssertEqual(t, len(ul.Box().Children), 2)
		for _, li := range ul.Box().Children {
			tu.AssertEqual(t, len(li.Box().Children), 2)
			tu.AssertEqual(t, unpack1(li).Box().PositionY,
				unpack1(li.Box().Children[1]).Box().PositionY)
		}
	}
}
