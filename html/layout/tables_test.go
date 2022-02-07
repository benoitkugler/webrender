package layout

import (
	"fmt"
	"strings"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Tests for layout of tables.

func TestInlineTable(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="display: inline-table; border-spacing: 10px; margin: 5px">
        <tr>
          <td><img src=pattern.png style="width: 20px"></td>
          <td><img src=pattern.png style="width: 30px"></td>
        </tr>
      </table>
      foo
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	tableWrapper, text := unpack2(line)
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(5), "") // 0 + margin-left
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(15), "")  // 0 + border-spacing
	tu.AssertEqual(t, td1.Box().Width, pr.Float(20), "")
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(45), "") // 15 + 20 + border-spacing
	tu.AssertEqual(t, td2.Box().Width, pr.Float(30), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(80), "")                // 20 + 30 + 3 * border-spacing
	tu.AssertEqual(t, tableWrapper.Box().MarginWidth(), pr.Float(90), "") // 80 + 2 * margin
	tu.AssertEqual(t, text.Box().PositionX, pr.Float(90), "")
}

func TestImplicitWidthTableColPercent(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// See https://github.com/Kozea/WeasyPrint/issues/169
	page := renderOnePage(t, `
      <table>
        <col style="width:25%"></col>
        <col></col>
        <tr>
          <td></td>
          <td></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	_, _ = unpack2(row)
}

func TestImplicitWidthTableTdPercent(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table>
        <tr>
          <td style="width:25%"></td>
          <td></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	_, _ = unpack2(row)
}

func TestLayoutTableFixed1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="table-layout: fixed; border-spacing: 10px; margin: 5px">
        <colgroup>
          <col style="width: 20px" />
        </colgroup>
        <tr>
          <td></td>
          <td style="width: 40px">a</td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(5), "") // 0 + margin-left
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(15), "")  // 5 + border-spacing
	tu.AssertEqual(t, td1.Box().Width, pr.Float(20), "")
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(45), "") // 15 + 20 + border-spacing
	tu.AssertEqual(t, td2.Box().Width, pr.Float(40), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(90), "") // 20 + 40 + 3 * border-spacing
}

func TestLayoutTableFixed2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="table-layout: fixed; border-spacing: 10px; width: 200px;
                    margin: 5px">
        <tr>
          <td style="width: 20px">a</td>
          <td style="width: 40px"></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(5), "") // 0 + margin-left
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(15), "")  // 5 + border-spacing
	tu.AssertEqual(t, td1.Box().Width, pr.Float(75), "")      // 20 + ((200 - 20 - 40 - 3 * border-spacing) / 2)
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(100), "") // 15 + 75 + border-spacing
	tu.AssertEqual(t, td2.Box().Width, pr.Float(95), "")      // 40 + ((200 - 20 - 40 - 3 * border-spacing) / 2)
	tu.AssertEqual(t, table.Box().Width, pr.Float(200), "")
}

func TestLayoutTableFixed3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="table-layout: fixed; border-spacing: 10px;
                    width: 110px; margin: 5px">
        <tr>
          <td style="width: 40px">a</td>
          <td>b</td>
        </tr>
        <tr>
          <td style="width: 50px">a</td>
          <td style="width: 30px">b</td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row1, row2 := unpack2(rowGroup)
	td1, td2 := unpack2(row1)
	td3, td4 := unpack2(row2)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(5), "") // 0 + margin-left
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(15), "")  // 0 + border-spacing
	tu.AssertEqual(t, td3.Box().PositionX, pr.Float(15), "")
	tu.AssertEqual(t, td1.Box().Width, pr.Float(40), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(40), "")
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(65), "") // 15 + 40 + border-spacing
	tu.AssertEqual(t, td4.Box().PositionX, pr.Float(65), "")
	tu.AssertEqual(t, td3.Box().Width, pr.Float(40), "")
	tu.AssertEqual(t, td4.Box().Width, pr.Float(40), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(110), "") // 20 + 40 + 3 * border-spacing
}

func TestLayoutTableFixed4(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="table-layout: fixed; border-spacing: 0;
                    width: 100px; margin: 10px">
        <colgroup>
          <col />
          <col style="width: 20px" />
        </colgroup>
        <tr>
          <td></td>
          <td style="width: 40px">a</td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(10), "") // 0 + margin-left
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(10), "")
	tu.AssertEqual(t, td1.Box().Width, pr.Float(80), "")     // 100 - 20
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(90), "") // 10 + 80
	tu.AssertEqual(t, td2.Box().Width, pr.Float(20), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(100), "")
}

func TestLayoutTableFixed5(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// With border-collapse
	page := renderOnePage(t, `
      <style>
        /* Do ! apply: */
        colgroup, col, tbody, tr, td { margin: 1000px }
      </style>
      <table style="table-layout: fixed;
                    border-collapse: collapse; border: 10px solid;
                    /* ignored with collapsed borders: */
                    border-spacing: 10000px; padding: 1000px">
        <colgroup>
          <col style="width: 30px" />
        </colgroup>
        <tbody>
          <tr>
            <td style="padding: 2px"></td>
            <td style="width: 34px; padding: 10px; border: 2px solid"></td>
          </tr>
        </tbody>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().BorderLeftWidth, pr.Float(5), "") // half of the collapsed 10px border
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(5), "")         // border-spacing is ignored
	tu.AssertEqual(t, td1.Box().MarginWidth(), pr.Float(30), "")    // as <col>
	tu.AssertEqual(t, td1.Box().Width, pr.Float(20), "")            // 30 - 5 (border-left) - 1 (border-right) - 2*2
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(35), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(34), "")
	tu.AssertEqual(t, td2.Box().MarginWidth(), pr.Float(60), "")    // 34 + 2*10 + 5 + 1
	tu.AssertEqual(t, table.Box().Width, pr.Float(90), "")          // 30 + 60
	tu.AssertEqual(t, table.Box().MarginWidth(), pr.Float(100), "") // 90 + 2*5 (border)
}

func TestLayoutTableAuto1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <body style="width: 100px">
      <table style="border-spacing: 10px; margin: auto">
        <tr>
          <td><img src=pattern.png></td>
          <td><img src=pattern.png></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, tableWrapper.Box().Width, pr.Float(38), "")      // Same as table, see below
	tu.AssertEqual(t, tableWrapper.Box().MarginLeft, pr.Float(31), "") // 0 + margin-left = (100 - 38) / 2
	tu.AssertEqual(t, tableWrapper.Box().MarginRight, pr.Float(31), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(31), "")
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(41), "") // 31 + spacing
	tu.AssertEqual(t, td1.Box().Width, pr.Float(4), "")
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(55), "") // 31 + 4 + spacing
	tu.AssertEqual(t, td2.Box().Width, pr.Float(4), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(38), "") // 3 * spacing + 2 * 4
}

func TestLayoutTableAuto2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <body style="width: 50px">
      <table style="border-spacing: 1px; margin: 10%">
        <tr>
          <td style="border: 3px solid black"><img src=pattern.png></td>
          <td style="border: 3px solid black">
            <img src=pattern.png><img src=pattern.png>
          </td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(5), "") // 0 + margin-left
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(6), "")   // 5 + border-spacing
	tu.AssertEqual(t, td1.Box().Width, pr.Float(4), "")
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(17), "") // 6 + 4 + spacing + 2 * border
	tu.AssertEqual(t, td2.Box().Width, pr.Float(8), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(27), "") // 3 * spacing + 4 + 8 + 4 * border
}

func TestLayoutTableAuto3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="border-spacing: 1px; margin: 5px; font-size: 0">
        <tr>
          <td></td>
          <td><img src=pattern.png><img src=pattern.png></td>
        </tr>
        <tr>
          <td>
            <img src=pattern.png>
            <img src=pattern.png>
            <img src=pattern.png>
          </td>
          <td><img src=pattern.png></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row1, row2 := unpack2(rowGroup)
	td11, td12 := unpack2(row1)
	td21, td22 := unpack2(row2)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(5), "")         // 0 + margin-left
	tu.AssertEqual(t, td11.Box().PositionX, td21.Box().PositionX, "") // 5 + spacing
	tu.AssertEqual(t, td21.Box().PositionX, pr.Float(6), "")          // 5 + spacing
	tu.AssertEqual(t, td11.Box().Width, td21.Box().Width, "")
	tu.AssertEqual(t, td21.Box().Width, pr.Float(12), "")
	tu.AssertEqual(t, td12.Box().PositionX, td22.Box().PositionX, "") // 6 + 12 + spacing
	tu.AssertEqual(t, td22.Box().PositionX, pr.Float(19), "")         // 6 + 12 + spacing
	tu.AssertEqual(t, td12.Box().Width, td22.Box().Width, "")
	tu.AssertEqual(t, td22.Box().Width, pr.Float(8), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(23), "") // 3 * spacing + 12 + 8
}

func TestLayoutTableAuto4(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="border-spacing: 1px; margin: 5px">
        <tr>
          <td style="border: 1px solid black"><img src=pattern.png></td>
          <td style="border: 2px solid black; padding: 1px">
            <img src=pattern.png>
          </td>
        </tr>
        <tr>
          <td style="border: 5px solid black"><img src=pattern.png></td>
          <td><img src=pattern.png></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row1, row2 := unpack2(rowGroup)
	td11, td12 := unpack2(row1)
	td21, td22 := unpack2(row2)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(5), "")         // 0 + margin-left
	tu.AssertEqual(t, td11.Box().PositionX, td21.Box().PositionX, "") // 5 + spacing
	tu.AssertEqual(t, td21.Box().PositionX, pr.Float(6), "")          // 5 + spacing
	tu.AssertEqual(t, td11.Box().Width, pr.Float(12), "")             // 4 + 2 * 5 - 2 * 1
	tu.AssertEqual(t, td21.Box().Width, pr.Float(4), "")
	tu.AssertEqual(t, td12.Box().PositionX, td22.Box().PositionX, "") // 6 + 4 + 2 * b1 + sp
	tu.AssertEqual(t, td22.Box().PositionX, pr.Float(21), "")         // 6 + 4 + 2 * b1 + sp
	tu.AssertEqual(t, td12.Box().Width, pr.Float(4), "")
	tu.AssertEqual(t, td22.Box().Width, pr.Float(10), "")  // 4 + 2 * 3
	tu.AssertEqual(t, table.Box().Width, pr.Float(27), "") // 3 * spacing + 4 + 4 + 2 * b1 + 2 * b2
}

func TestLayoutTableAuto5(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        * { font-family: weasyprint }
      </style>
      <table style="width: 1000px; font-family: weasyprint">
        <tr>
          <td style="width: 40px">aa aa aa aa</td>
          <td style="width: 40px">aaaaaaaaaaa</td>
          <td>This will take the rest of the width</td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2, td3 := unpack3(row)

	tu.AssertEqual(t, table.Box().Width, pr.Float(1000), "")
	tu.AssertEqual(t, td1.Box().Width, pr.Float(40), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(11)*pr.Float(16), "")
	tu.AssertEqual(t, td3.Box().Width, pr.Float(1000)-pr.Float(40)-pr.Float(11)*pr.Float(16), "")
}

func TestLayoutTableAuto6(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 100px 1000px; }
      </style>
      <table style="border-spacing: 1px; margin-right: 79px; font-size: 0">
        <tr>
          <td><img src=pattern.png></td>
          <td>
            <img src=pattern.png> <img src=pattern.png>
            <img src=pattern.png> <img src=pattern.png>
            <img src=pattern.png> <img src=pattern.png>
            <img src=pattern.png> <img src=pattern.png>
            <img src=pattern.png>
          </td>
        </tr>
        <tr>
          <td></td>
        </tr>
      </table>
    `)
	// Preferred minimum width is 2 * 4 + 3 * 1 = 11
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row1, row2 := unpack2(rowGroup)
	td11, td12 := unpack2(row1)
	td21 := row2.Box().Children[0]
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, td11.Box().PositionX, td21.Box().PositionX, "") // spacing
	tu.AssertEqual(t, td21.Box().PositionX, pr.Float(1), "")          // spacing
	tu.AssertEqual(t, td11.Box().Width, td21.Box().Width, "")         // minimum width
	tu.AssertEqual(t, td21.Box().Width, pr.Float(4), "")              // minimum width
	tu.AssertEqual(t, td12.Box().PositionX, pr.Float(6), "")          // 1 + 5 + sp
	tu.AssertEqual(t, td12.Box().Width, pr.Float(14), "")             // available width
	tu.AssertEqual(t, table.Box().Width, pr.Float(21), "")
}

func TestLayoutTableAuto7(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="border-spacing: 10px; margin: 5px">
        <colgroup>
          <col style="width: 20px" />
        </colgroup>
        <tr>
          <td></td>
          <td style="width: 40px">a</td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(5), "") // 0 + margin-left
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(15), "")  // 0 + border-spacing
	tu.AssertEqual(t, td1.Box().Width, pr.Float(20), "")
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(45), "") // 15 + 20 + border-spacing
	tu.AssertEqual(t, td2.Box().Width, pr.Float(40), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(90), "") // 20 + 40 + 3 * border-spacing
}

func TestLayoutTableAuto8(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="border-spacing: 10px; width: 120px; margin: 5px;
                    font-size: 0">
        <tr>
          <td style="width: 20px"><img src=pattern.png></td>
          <td><img src=pattern.png style="width: 40px"></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(5), "") // 0 + margin-left
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(15), "")  // 5 + border-spacing
	tu.AssertEqual(t, td1.Box().Width, pr.Float(20), "")      // fixed
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(45), "")  // 15 + 20 + border-spacing
	tu.AssertEqual(t, td2.Box().Width, pr.Float(70), "")      // 120 - 3 * border-spacing - 20
	tu.AssertEqual(t, table.Box().Width, pr.Float(120), "")
}

func TestLayoutTableAuto9(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="border-spacing: 10px; width: 110px; margin: 5px">
        <tr>
          <td style="width: 60px"></td>
          <td></td>
        </tr>
        <tr>
          <td style="width: 50px"></td>
          <td style="width: 30px"></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row1, row2 := unpack2(rowGroup)
	td1, td2 := unpack2(row1)
	td3, td4 := unpack2(row2)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(5), "") // 0 + margin-left
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(15), "")  // 0 + border-spacing
	tu.AssertEqual(t, td3.Box().PositionX, pr.Float(15), "")
	tu.AssertEqual(t, td1.Box().Width, pr.Float(60), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(30), "")
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(85), "") // 15 + 60 + border-spacing
	tu.AssertEqual(t, td4.Box().PositionX, pr.Float(85), "")
	tu.AssertEqual(t, td3.Box().Width, pr.Float(60), "")
	tu.AssertEqual(t, td4.Box().Width, pr.Float(30), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(120), "") // 60 + 30 + 3 * border-spacing
}

func TestLayoutTableAuto10(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="border-spacing: 0; width: 14px; margin: 10px">
        <colgroup>
          <col />
          <col style="width: 6px" />
        </colgroup>
        <tr>
          <td><img src=pattern.png><img src=pattern.png></td>
          <td style="width: 8px"></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(10), "") // 0 + margin-left
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(10), "")
	tu.AssertEqual(t, td1.Box().Width, pr.Float(6), "")      // 14 - 8
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(16), "") // 10 + 6
	tu.AssertEqual(t, td2.Box().Width, pr.Float(8), "")      // maximum of the minimum widths for the column
	tu.AssertEqual(t, table.Box().Width, pr.Float(14), "")
}

func TestLayoutTableAuto11(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="border-spacing: 0">
        <tr>
          <td style="width: 10px"></td>
          <td colspan="3"></td>
        </tr>
        <tr>
          <td colspan="2" style="width: 22px"></td>
          <td style="width: 8px"></td>
          <td style="width: 8px"></td>
        </tr>
        <tr>
          <td></td>
          <td></td>
          <td colspan="2"></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row1, row2, row3 := unpack3(rowGroup)
	td11, td12 := unpack2(row1)
	td21, td22, td23 := unpack3(row2)
	td31, td32, td33 := unpack3(row3)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, td11.Box().Width, pr.Float(10), "") // fixed
	tu.AssertEqual(t, td12.Box().Width, pr.Float(28), "") // 38 - 10
	tu.AssertEqual(t, td21.Box().Width, pr.Float(22), "") // fixed
	tu.AssertEqual(t, td22.Box().Width, pr.Float(8), "")  // fixed
	tu.AssertEqual(t, td23.Box().Width, pr.Float(8), "")  // fixed
	tu.AssertEqual(t, td31.Box().Width, pr.Float(10), "") // same as first line
	tu.AssertEqual(t, td32.Box().Width, pr.Float(12), "") // 22 - 10
	tu.AssertEqual(t, td33.Box().Width, pr.Float(16), "") // 8 + 8 from second line
	tu.AssertEqual(t, table.Box().Width, pr.Float(38), "")
}

func TestLayoutTableAuto12(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="border-spacing: 10px">
        <tr>
          <td style="width: 10px"></td>
          <td colspan="3"></td>
        </tr>
        <tr>
          <td colspan="2" style="width: 32px"></td>
          <td style="width: 8px"></td>
          <td style="width: 8px"></td>
        </tr>
        <tr>
          <td></td>
          <td></td>
          <td colspan="2"></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row1, row2, row3 := unpack3(rowGroup)
	td11, td12 := unpack2(row1)
	td21, td22, td23 := unpack3(row2)
	td31, td32, td33 := unpack3(row3)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, td11.Box().Width, pr.Float(10), "") // fixed
	tu.AssertEqual(t, td12.Box().Width, pr.Float(48), "") // 32 - 10 - sp + 2 * 8 + 2 * sp
	tu.AssertEqual(t, td21.Box().Width, pr.Float(32), "") // fixed
	tu.AssertEqual(t, td22.Box().Width, pr.Float(8), "")  // fixed
	tu.AssertEqual(t, td23.Box().Width, pr.Float(8), "")  // fixed
	tu.AssertEqual(t, td31.Box().Width, pr.Float(10), "") // same as first line
	tu.AssertEqual(t, td32.Box().Width, pr.Float(12), "") // 32 - 10 - sp
	tu.AssertEqual(t, td33.Box().Width, pr.Float(26), "") // 2 * 8 + sp
	tu.AssertEqual(t, table.Box().Width, pr.Float(88), "")
}

func TestLayoutTableAuto13(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression tests: these used to crash
	page := renderOnePage(t, `
      <table style="width: 30px">
        <tr>
          <td colspan=2></td>
          <td></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(20), "") // 2 / 3 * 30
	tu.AssertEqual(t, td2.Box().Width, pr.Float(10), "") // 1 / 3 * 30
	tu.AssertEqual(t, table.Box().Width, pr.Float(30), "")
}

func TestLayoutTableAuto14(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="width: 20px">
        <col />
        <col />
        <tr>
          <td></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1 := row.Box().Children[0]
	tu.AssertEqual(t, td1.Box().Width, pr.Float(20), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(20), "")
}

func TestLayoutTableAuto15(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="width: 20px">
        <col />
        <col />
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	columnGroup := table.(bo.TableBoxITF).Table().ColumnGroups[0]
	column1, column2 := unpack2(columnGroup)
	tu.AssertEqual(t, column1.Box().Width, pr.Float(0), "")
	tu.AssertEqual(t, column2.Box().Width, pr.Float(0), "")
}

func TestLayoutTableAuto16(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Absolute table
	page := renderOnePage(t, `
      <table style="width: 30px; position: absolute">
        <tr>
          <td colspan=2></td>
          <td></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(20), "") // 2 / 3 * 30
	tu.AssertEqual(t, td2.Box().Width, pr.Float(10), "") // 1 / 3 * 30
	tu.AssertEqual(t, table.Box().Width, pr.Float(30), "")
}

func TestLayoutTableAuto17(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// With border-collapse
	page := renderOnePage(t, `
      <style>
        /* Do ! apply: */
        colgroup, col, tbody, tr, td { margin: 1000px }
      </style>
      <table style="border-collapse: collapse; border: 10px solid;
                    /* ignored with collapsed borders: */
                    border-spacing: 10000px; padding: 1000px">
        <colgroup>
          <col style="width: 30px" />
        </colgroup>
        <tbody>
          <tr>
            <td style="padding: 2px"></td>
            <td style="width: 34px; padding: 10px; border: 2px solid"></td>
          </tr>
        </tbody>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, tableWrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().BorderLeftWidth, pr.Float(5), "") // half of the collapsed 10px border
	tu.AssertEqual(t, td1.Box().PositionX, pr.Float(5), "")         // border-spacing is ignored
	tu.AssertEqual(t, td1.Box().MarginWidth(), pr.Float(30), "")    // as <col>
	tu.AssertEqual(t, td1.Box().Width, pr.Float(20), "")            // 30 - 5 (border-left) - 1 (border-right) - 2*2
	tu.AssertEqual(t, td2.Box().PositionX, pr.Float(35), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(34), "")
	tu.AssertEqual(t, td2.Box().MarginWidth(), pr.Float(60), "")    // 34 + 2*10 + 5 + 1
	tu.AssertEqual(t, table.Box().Width, pr.Float(90), "")          // 30 + 60
	tu.AssertEqual(t, table.Box().MarginWidth(), pr.Float(100), "") // 90 + 2*5 (border)
}

func TestLayoutTableAuto18(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Column widths as percentage
	page := renderOnePage(t, `
      <table style="width: 200px">
        <colgroup>
          <col style="width: 70%" />
          <col style="width: 30%" />
        </colgroup>
        <tbody>
          <tr>
            <td>a</td>
            <td>abc</td>
          </tr>
        </tbody>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(140), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(60.000004), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(200), "")
}

func TestLayoutTableAuto19(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Column group width
	page := renderOnePage(t, `
      <table style="width: 200px">
        <colgroup style="width: 100px">
          <col />
          <col />
        </colgroup>
        <col style="width: 100px" />
        <tbody>
          <tr>
            <td>a</td>
            <td>a</td>
            <td>abc</td>
          </tr>
        </tbody>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2, td3 := unpack3(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(100), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(100), "")
	tu.AssertEqual(t, td3.Box().Width, pr.Float(100), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(300), "")
}

func TestLayoutTableAuto20(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Multiple column width
	page := renderOnePage(t, `
      <table style="width: 200px">
        <colgroup>
          <col style="width: 50px" />
          <col style="width: 30px" />
          <col />
        </colgroup>
        <tbody>
          <tr>
            <td>a</td>
            <td>a</td>
            <td>abc</td>
          </tr>
        </tbody>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2, td3 := unpack3(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(50), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(30), "")
	tu.AssertEqual(t, td3.Box().Width, pr.Float(120), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(200), "")
}

func TestLayoutTableAuto21(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Fixed-width table with column group with widths as percentages && pixels
	page := renderOnePage(t, `
      <table style="width: 500px">
        <colgroup style="width: 100px">
          <col />
          <col />
        </colgroup>
        <colgroup style="width: 30%">
          <col />
          <col />
        </colgroup>
        <tbody>
          <tr>
            <td>a</td>
            <td>a</td>
            <td>abc</td>
            <td>abc</td>
          </tr>
        </tbody>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2, td3, td4 := unpack4(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(100), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(100), "")
	tu.AssertEqual(t, td3.Box().Width, pr.Float(150), "")
	tu.AssertEqual(t, td4.Box().Width, pr.Float(150), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(500), "")
}

func TestLayoutTableAuto22(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Auto-width table with column group with widths as percentages && pixels
	page := renderOnePage(t, `
      <table>
        <colgroup style="width: 10%">
          <col />
          <col />
        </colgroup>
        <colgroup style="width: 200px">
          <col />
          <col />
        </colgroup>
        <tbody>
          <tr>
            <td>a a</td>
            <td>a b</td>
            <td>a c</td>
            <td>a d</td>
          </tr>
        </tbody>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2, td3, td4 := unpack4(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(50), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(50), "")
	tu.AssertEqual(t, td3.Box().Width, pr.Float(200), "")
	tu.AssertEqual(t, td4.Box().Width, pr.Float(200), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(500), "")
}

func TestLayoutTableAuto23(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Wrong column group width
	page := renderOnePage(t, `
      <table style="width: 200px">
        <colgroup style="width: 20%">
          <col />
          <col />
        </colgroup>
        <tbody>
          <tr>
            <td>a</td>
            <td>a</td>
          </tr>
        </tbody>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(100), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(100), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(200), "")
}

func TestLayoutTableAuto24(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Column width as percentage && cell width := range pixels
	page := renderOnePage(t, `
      <table style="width: 200px">
        <colgroup>
          <col style="width: 70%" />
          <col />
        </colgroup>
        <tbody>
          <tr>
            <td>a</td>
            <td style="width: 60px">abc</td>
          </tr>
        </tbody>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(140), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(60), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(200), "")
}

func TestLayoutTableAuto25(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Column width && cell width as percentage
	page := renderOnePage(t, `
      <div style="width: 400px">
        <table style="width: 50%">
          <colgroup>
            <col style="width: 70%" />
            <col />
          </colgroup>
          <tbody>
            <tr>
              <td>a</td>
              <td style="width: 30%">abc</td>
            </tr>
          </tbody>
        </table>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tableWrapper := div.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(140), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(60.000004), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(200), "")
}

func TestLayoutTableAuto26(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/307
	// Table with a cell larger than the table"s max-width
	_ = renderOnePage(t, `
      <table style="max-width: 300px">
        <td style="width: 400px"></td>
      </table>
    `)
}

func TestLayoutTableAuto27(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Table with a cell larger than the table"s width
	_ = renderOnePage(t, `
      <table style="width: 300px">
        <td style="width: 400px"></td>
      </table>
    `)
}

func TestLayoutTableAuto28(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Table with a cell larger than the table"s width && max-width
	_ = renderOnePage(t, `
      <table style="width: 300px; max-width: 350px">
        <td style="width: 400px"></td>
      </table>
    `)
}

func TestLayoutTableAuto29(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Table with a cell larger than the table"s width && max-width
	_ = renderOnePage(t, `
      <table style="width: 300px; max-width: 350px">
        <td style="padding: 50px">
          <div style="width: 300px"></div>
        </td>
      </table>
    `)
}

func TestLayoutTableAuto30(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Table with a cell larger than the table"s max-width
	_ = renderOnePage(t, `
      <table style="max-width: 300px; margin: 100px">
        <td style="width: 400px"></td>
      </table>
    `)
}

func TestLayoutTableAuto31(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Test a table with column widths < table width < column width + spacing
	_ = renderOnePage(t, `
      <table style="width: 300px; border-spacing: 2px">
        <td style="width: 299px"></td>
      </table>
    `)
}

func TestLayoutTableAuto32(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Table with a cell larger than the table"s width
	page := renderOnePage(t, `
      <table style="width: 300px; margin: 100px">
        <td style="width: 400px"></td>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	tu.AssertEqual(t, tableWrapper.Box().MarginWidth(), pr.Float(600), "") // 400 + 2 * 100
}

func TestLayoutTableAuto33(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Div with auto width containing a table with a min-width
	page := renderOnePage(t, `
      <div style="float: left">
        <table style="min-width: 400px; margin: 100px">
          <td></td>
        </table>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tableWrapper := div.Box().Children[0]
	tu.AssertEqual(t, div.Box().MarginWidth(), pr.Float(600), "")          // 400 + 2 * 100
	tu.AssertEqual(t, tableWrapper.Box().MarginWidth(), pr.Float(600), "") // 400 + 2 * 100
}

func TestLayoutTableAuto34(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Div with auto width containing an empty table with a min-width
	page := renderOnePage(t, `
      <div style="float: left">
        <table style="min-width: 400px; margin: 100px"></table>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tableWrapper := div.Box().Children[0]
	tu.AssertEqual(t, div.Box().MarginWidth(), pr.Float(600), "")          // 400 + 2 * 100
	tu.AssertEqual(t, tableWrapper.Box().MarginWidth(), pr.Float(600), "") // 400 + 2 * 100
}

func TestLayoutTableAuto35(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Div with auto width containing a table with a cell larger than the
	// table"s max-width
	page := renderOnePage(t, `
      <div style="float: left">
        <table style="max-width: 300px; margin: 100px">
          <td style="width: 400px"></td>
        </table>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tableWrapper := div.Box().Children[0]
	tu.AssertEqual(t, div.Box().MarginWidth(), pr.Float(600), "")          // 400 + 2 * 100
	tu.AssertEqual(t, tableWrapper.Box().MarginWidth(), pr.Float(600), "") // 400 + 2 * 100
}

func TestLayoutTableAuto36(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Test regression on a crash: https://github.com/Kozea/WeasyPrint/pull/152
	_ = renderOnePage(t, `
      <table>
        <td style="width: 50%">
      </table>
    `)
}

func TestLayoutTableAuto37(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Other crashes: https://github.com/Kozea/WeasyPrint/issues/305
	_ = renderOnePage(t, `
      <table>
        <tr>
          <td>
            <table>
              <tr>
                <th>Test</th>
              </tr>
              <tr>
                <td style="min-width: 100%;"></td>
                <td style="width: 48px;"></td>
              </tr>
            </table>
          </td>
        </tr>
      </table>
    `)
}

func TestLayoutTableAuto38(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	_ = renderOnePage(t, `
      <table>
        <tr>
          <td>
            <table>
              <tr>
                <td style="width: 100%;"></td>
                <td style="width: 48px;">
                  <img src="icon.png">
                </td>
              </tr>
            </table>
          </td>
        </tr>
      </table>
    `)
}

func TestLayoutTableAuto39(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	_ = renderOnePage(t, `
      <table>
        <tr>
          <td>
            <table style="display: inline-table">
              <tr>
                <td style="width: 100%;"></td>
                <td></td>
              </tr>
            </table>
          </td>
        </tr>
      </table>
    `)
}

func TestLayoutTableAuto40(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/368
	// Check that white-space is used for the shrink-to-fit algorithm
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
      </style>
      <table style="width: 0">
        <td style="font-family: weasyprint; white-space: nowrap">a a</td>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tableWrapper := div.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	tu.AssertEqual(t, table.Box().Width, pr.Float(16)*pr.Float(3), "")
}

func TestLayoutTableAuto41(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Cell width as percentage := range auto-width table
	page := renderOnePage(t, `
      <div style="width: 100px">
        <table>
          <tbody>
            <tr>
              <td>a a a a a a a a</td>
              <td style="width: 30%">a a a a a a a a</td>
            </tr>
          </tbody>
        </table>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tableWrapper := div.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(70), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(30.000002), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(100), "")
}

func TestLayoutTableAuto42(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Cell width as percentage in auto-width table
	page := renderOnePage(t, `
      <table>
        <tbody>
            <tr>
              <td style="width: 70px">a a a a a a a a</td>
              <td style="width: 30%">a a a a a a a a</td>
            </tr>
        </tbody>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2 := unpack2(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(70), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(30), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(100), "")
}

func TestLayoutTableAuto43(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Cell width as percentage on colspan cell := range auto-width table
	page := renderOnePage(t, `
      <div style="width: 100px">
        <table>
          <tbody>
            <tr>
              <td>a a a a a a a a</td>
              <td style="width: 30%" colspan=2>a a a a a a a a</td>
              <td>a a a a a a a a</td>
            </tr>
          </tbody>
        </table>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tableWrapper := div.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2, td3 := unpack3(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(35), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(30.000002), "")
	tu.AssertEqual(t, td3.Box().Width, pr.Float(35), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(100), "")
}

func TestLayoutTableAuto44(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Cells widths as percentages on normal && colspan cells
	page := renderOnePage(t, `
      <div style="width: 100px">
        <table>
          <tbody>
            <tr>
              <td>a a a a a a a a</td>
              <td style="width: 30%" colspan=2>a a a a a a a a</td>
              <td>a a a a a a a a</td>
              <td style="width: 40%">a a a a a a a a</td>
              <td>a a a a a a a a</td>
            </tr>
          </tbody>
        </table>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tableWrapper := div.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td1, td2, td3, td4, td5 := unpack5(row)
	tu.AssertEqual(t, td1.Box().Width, pr.Float(10), "")
	tu.AssertEqual(t, td2.Box().Width, pr.Float(30.000002), "")
	tu.AssertEqual(t, td3.Box().Width, pr.Float(10), "")
	tu.AssertEqual(t, td4.Box().Width, pr.Float(40), "")
	tu.AssertEqual(t, td5.Box().Width, pr.Float(10), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(100), "")
}

func TestLayoutTableAuto45(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Cells widths as percentage on multiple lines
	page := renderOnePage(t, `
      <div style="width: 1000px">
        <table>
          <tbody>
            <tr>
              <td>a a a a a a a a</td>
              <td style="width: 30%">a a a a a a a a</td>
              <td>a a a a a a a a</td>
              <td style="width: 40%">a a a a a a a a</td>
              <td>a a a a a a a a</td>
            </tr>
            <tr>
              <td style="width: 31%" colspan=2>a a a a a a a a</td>
              <td>a a a a a a a a</td>
              <td style="width: 42%" colspan=2>a a a a a a a a</td>
            </tr>
          </tbody>
        </table>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tableWrapper := div.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row1, row2 := unpack2(rowGroup)
	td11, td12, td13, td14, td15 := unpack5(row1)
	td21, td22, td23 := unpack3(row2)
	tu.AssertEqual(t, td11.Box().Width, pr.Float(10), "")  // 31% - 30%
	tu.AssertEqual(t, td12.Box().Width, pr.Float(300), "") // 30%
	tu.AssertEqual(t, td13.Box().Width, pr.Float(270), "") // 1000 - 31% - 42%
	tu.AssertEqual(t, td14.Box().Width, pr.Float(400), "") // 40%
	tu.AssertEqual(t, td15.Box().Width, pr.Float(20), "")  // 42% - 2%
	tu.AssertEqual(t, td21.Box().Width, pr.Float(310), "") // 31%
	tu.AssertEqual(t, td22.Box().Width, pr.Float(270), "") // 1000 - 31% - 42%
	tu.AssertEqual(t, td23.Box().Width, pr.Float(420), "") // 42%
	tu.AssertEqual(t, table.Box().Width, pr.Float(1000), "")
}

func TestLayoutTableAuto46(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Test regression:
	// http://test.weasyprint.org/suite-css21/chapter8/section2/test56/
	page := renderOnePage(t, `
      <div style="position: absolute">
        <table style="margin: 50px; border: 20px solid black">
          <tr>
            <td style="width: 200px; height: 200px"></td>
          </tr>
        </table>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tableWrapper := div.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td := row.Box().Children[0]
	tu.AssertEqual(t, td.Box().Width, pr.Float(200), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(200), "")
	tu.AssertEqual(t, div.Box().Width, pr.Float(340), "") // 200 + 2 * 50 + 2 * 20
}

func TestLayoutTableAuto47(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Test regression {
	// https://github.com/Kozea/WeasyPrint/issues/666
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
      </style>
      <table style="font-family: weasyprint">
        <tr>
          <td colspan=5>aaa</td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td := row.Box().Children[0]
	tu.AssertEqual(t, td.Box().Width, pr.Float(48), "") // 3 * font-size
}

func TestLayoutTableAuto48(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Related to:
	// https://github.com/Kozea/WeasyPrint/issues/685
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
      </style>
      <table style="font-family: weasyprint; border-spacing: 100px;
                    border-collapse: collapse">
        <tr>
          <td colspan=5>aaa</td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	td := row.Box().Children[0]
	tu.AssertEqual(t, td.Box().Width, pr.Float(48), "") // 3 * font-size
}

// xfail
// func TestLayoutTableAuto49(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	// Related to {
// 	// https://github.com/Kozea/WeasyPrint/issues/685
// 	// See TODO := range tableLayout.groupLayout
// 	page := renderOnePage(t, `
//       <style>
//         @font-face { src: url(weasyprint.otf); font-family: weasyprint }
//       </style>
//       <table style="font-family: weasyprint; border-spacing: 100px">
//         <tr>
//           <td colspan=5>aaa</td>
//         </tr>
//       </table>
//     `)
// 	html := page.Box().Children[0]
// 	body := html.Box().Children[0]
// 	tableWrapper := body.Box().Children[0]
// 	table := tableWrapper.Box().Children[0]
// 	rowGroup := table.Box().Children[0]
// 	row := rowGroup.Box().Children[0]
// 	td := row.Box().Children[0]
// 	tu.AssertEqual(t, td.Box().Width, pr.Float(48), "") // 3 * font-size
// }

func TestLayoutTableAuto50(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Test regression:
	// https://github.com/Kozea/WeasyPrint/issues/685
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
      </style>
      <table style="font-family: weasyprint; border-spacing: 5px">
       <tr><td>a</td><td>a</td><td>a</td><td>a</td><td>a</td></tr>
       <tr>
         <td colspan="5">aaa aaa aaa aaa</td>
       </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row1, row2 := unpack2(rowGroup)
	for _, td := range row1.Box().Children {
		tu.AssertEqual(t, td.Box().Width, pr.Float(44), "") // (15 * fontSize - 4 * sp) / 5
	}
	td21 := row2.Box().Children[0]
	tu.AssertEqual(t, td21.Box().Width, pr.Float(240), "") // 15 * fontSize
}

func TestExplicitWidthTablePercentRtl(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	for _, data := range []struct {
		bodyWidth, tableWidth string
		checkWidth            pr.Float
		positions, widths     []pr.Float
	}{
		{"500px", "230px", 220, []pr.Float{170, 5}, []pr.Float{45, 155}},
		{"530px", "100%", 520, []pr.Float{395, 5}, []pr.Float{120, 380}},
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        body { width: %s }
        table { width: %s; table-layout: fixed; direction: rtl;
                border-collapse: collapse; font-size: 1px }
        td, th { border: 10px solid }
      </style>
      <table style="">
        <col style="width: 25%%"></col>
        <col></col>
        <tr>
          <th>الاسم</th>
          <th>العائلة</th>
        </tr>
        <tr>
          <td>محمد يوسف</td>
          <td>29</td>
        </tr>
      </table>
    `, data.bodyWidth, data.tableWidth))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		wrapper := body.Box().Children[0]
		table := wrapper.Box().Children[0]
		rowGroup := table.Box().Children[0]
		row1, row2 := unpack2(rowGroup)

		tu.AssertEqual(t, table.Box().PositionX, pr.Float(0), "")
		tu.AssertEqual(t, table.Box().Width, data.checkWidth, "")
		var positionsX1, positionsX2, widths1, widths2 []pr.Float
		for _, child := range row1.Box().Children {
			positionsX1 = append(positionsX1, child.Box().PositionX)
			widths1 = append(widths1, child.Box().Width.V())
		}
		for _, child := range row2.Box().Children {
			positionsX2 = append(positionsX2, child.Box().PositionX)
			widths2 = append(widths2, child.Box().Width.V())
		}

		tu.AssertEqual(t, positionsX1, data.positions, "")
		tu.AssertEqual(t, positionsX2, data.positions, "")
		tu.AssertEqual(t, widths1, data.widths, "")
		tu.AssertEqual(t, widths2, data.widths, "")

	}
}

func TestTableColumnWidth1(t *testing.T) {
	source := `
      <style>
        body { width: 20000px; margin: 0 }
        table {
          width: 10000px; margin: 0 auto; border-spacing: 100px 0;
          table-layout: fixed
        }
        td { border: 10px solid; padding: 1px }
      </style>
      <table>
        <col style="width: 10%">
        <tr>
          <td style="width: 30%" colspan=3>
          <td>
        </tr>
        <tr>
          <td>
          <td>
          <td>
          <td>
        </tr>
        <tr>
          <td>
          <td colspan=12>This cell will be truncated to grid width
          <td>This cell will be removed as it is beyond the grid width
        </tr>
      </table>
    `
	capt := tu.CaptureLogs()
	page := renderOnePage(t, source)
	logs := capt.Logs()
	tu.AssertEqual(t, len(logs), 1, "")
	tu.AssertEqual(t, strings.Contains(logs[0], "This table row has more columns than the table, ignored 1 cell"), true, "")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	firstRow, secondRow, thirdRow := unpack3(rowGroup)
	cells := [][]Box{firstRow.Box().Children, secondRow.Box().Children, thirdRow.Box().Children}
	tu.AssertEqual(t, len(firstRow.Box().Children), 2, "")
	tu.AssertEqual(t, len(secondRow.Box().Children), 4, "")
	// Third cell here is completly removed
	tu.AssertEqual(t, len(thirdRow.Box().Children), 2, "")

	tu.AssertEqual(t, body.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, wrapper.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, wrapper.Box().MarginLeft, pr.Float(5000), "")
	tu.AssertEqual(t, wrapper.Box().ContentBoxX(), pr.Float(5000), "") // auto margin-left
	tu.AssertEqual(t, wrapper.Box().Width, pr.Float(10000), "")
	tu.AssertEqual(t, table.Box().PositionX, pr.Float(5000), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(10000), "")
	tu.AssertEqual(t, rowGroup.Box().PositionX, pr.Float(5100), "") // 5000 + borderSpacing
	tu.AssertEqual(t, rowGroup.Box().Width, pr.Float(9800), "")     // 10000 - 2*border-spacing
	tu.AssertEqual(t, firstRow.Box().PositionX, rowGroup.Box().PositionX, "")
	tu.AssertEqual(t, firstRow.Box().Width, rowGroup.Box().Width, "")

	// This cell has colspan=3
	tu.AssertEqual(t, cells[0][0].Box().PositionX, pr.Float(5100), "") // 5000 + border-spacing
	// `width` on a cell sets the content width
	tu.AssertEqual(t, cells[0][0].Box().Width, pr.Float(3000), "")         // 30% of 10000px
	tu.AssertEqual(t, cells[0][0].Box().BorderWidth(), pr.Float(3022), "") // 3000 + borders + padding

	// Second cell of the first line, but on the fourth && last column
	tu.AssertEqual(t, cells[0][1].Box().PositionX, pr.Float(8222), "")     // 5100 + 3022 + border-spacing
	tu.AssertEqual(t, cells[0][1].Box().BorderWidth(), pr.Float(6678), "") // 10000 - 3022 - 3*100
	tu.AssertEqual(t, cells[0][1].Box().Width, pr.Float(6656), "")         // 6678 - borders - padding

	tu.AssertEqual(t, cells[1][0].Box().PositionX, pr.Float(5100), "") // 5000 + border-spacing
	// `width` on a column sets the border width of cells
	tu.AssertEqual(t, cells[1][0].Box().BorderWidth(), pr.Float(1000), "") // 10% of 10000px
	tu.AssertEqual(t, cells[1][0].Box().Width, pr.Float(978), "")          // 1000 - borders - padding

	tu.AssertEqual(t, cells[1][1].Box().PositionX, pr.Float(6200), "")    // 5100 + 1000 + border-spacing
	tu.AssertEqual(t, cells[1][1].Box().BorderWidth(), pr.Float(911), "") // (3022 - 1000 - 2*100) / 2
	tu.AssertEqual(t, cells[1][1].Box().Width, pr.Float(889), "")         // 911 - borders - padding

	tu.AssertEqual(t, cells[1][2].Box().PositionX, pr.Float(7211), "")    // 6200 + 911 + border-spacing
	tu.AssertEqual(t, cells[1][2].Box().BorderWidth(), pr.Float(911), "") // (3022 - 1000 - 2*100) / 2
	tu.AssertEqual(t, cells[1][2].Box().Width, pr.Float(889), "")         // 911 - borders - padding

	// Same as cells[0][1]
	tu.AssertEqual(t, cells[1][3].Box().PositionX, pr.Float(8222), "") // Also 7211 + 911 + border-spacing
	tu.AssertEqual(t, cells[1][3].Box().BorderWidth(), pr.Float(6678), "")
	tu.AssertEqual(t, cells[1][3].Box().Width, pr.Float(6656), "")

	// Same as cells[1][0]
	tu.AssertEqual(t, cells[2][0].Box().PositionX, pr.Float(5100), "")
	tu.AssertEqual(t, cells[2][0].Box().BorderWidth(), pr.Float(1000), "")
	tu.AssertEqual(t, cells[2][0].Box().Width, pr.Float(978), "")

	tu.AssertEqual(t, cells[2][1].Box().PositionX, pr.Float(6200), "")     // Same as cells[1][1]
	tu.AssertEqual(t, cells[2][1].Box().BorderWidth(), pr.Float(8700), "") // 1000 - 1000 - 3*border-spacing
	tu.AssertEqual(t, cells[2][1].Box().Width, pr.Float(8678), "")         // 8700 - borders - padding
	tu.AssertEqual(t, cells[2][1].Box().Colspan, 3, "")                    // truncated to grid width
}

func TestTableColumnWidth2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        table { width: 1000px; border-spacing: 100px; table-layout: fixed }
      </style>
      <table>
        <tr>
          <td style="width: 50%">
          <td style="width: 60%">
          <td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	tu.AssertEqual(t, row.Box().Children[0].Box().Width, pr.Float(500), "")
	tu.AssertEqual(t, row.Box().Children[1].Box().Width, pr.Float(600), "")
	tu.AssertEqual(t, row.Box().Children[2].Box().Width, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(1500), "") // 500 + 600 + 4 * border-spacing
}

func TestTableColumnWidth3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Sum of columns width larger that the table width:
	// increase the table width
	page := renderOnePage(t, `
      <style>
        table { width: 1000px; border-spacing: 100px; table-layout: fixed }
        td { width: 60% }
      </style>
      <table>
        <tr>
          <td>
          <td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]
	row := rowGroup.Box().Children[0]
	cell1, cell2 := unpack2(row)
	tu.AssertEqual(t, cell1.Box().Width, pr.Float(600), "") // 60% of 1000px
	tu.AssertEqual(t, cell2.Box().Width, pr.Float(600), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(1500), "") // 600 + 600 + 3*border-spacing
	tu.AssertEqual(t, wrapper.Box().Width, table.Box().Width, "")
}

func TestTableRowHeight1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <table style="width: 1000px; border-spacing: 0 100px;
                    font: 20px/1em serif; margin: 3px; table-layout: fixed">
        <tr>
          <td rowspan=0 style="height: 420px; vertical-align: top"></td>
          <td>X<br>X<br>X</td>
          <td><table style="margin-top: 20px; border-spacing: 0">
            <tr><td>X</td></tr></table></td>
          <td style="vertical-align: top">X</td>
          <td style="vertical-align: middle">X</td>
          <td style="vertical-align: bottom">X</td>
        </tr>
        <tr>
          <!-- cells with no text (no line boxes) is a corner case
               := range cell baselines -->
          <td style="padding: 15px"></td>
          <td><div style="height: 10px"></div></td>
        </tr>
        <tr></tr>
        <tr>
            <td style="vertical-align: bottom"></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0]
	rowGroup := table.Box().Children[0]

	tu.AssertEqual(t, wrapper.Box().PositionY, pr.Float(0), "")
	tu.AssertEqual(t, table.Box().PositionY, pr.Float(3), "") // 0 + margin-top
	tu.AssertEqual(t, table.Box().Height, pr.Float(620), "")  // sum of row heigths + 5*border-spacing
	tu.AssertEqual(t, wrapper.Box().Height, table.Box().Height, "")
	tu.AssertEqual(t, rowGroup.Box().PositionY, pr.Float(103), "") // 3 + border-spacing
	tu.AssertEqual(t, rowGroup.Box().Height, pr.Float(420), "")    // 620 - 2*border-spacing
	var (
		heights, positionY                                          []pr.Float
		cells, borders, paddingTops, paddingBottoms, positionYCells [][]pr.Float
	)
	for _, row := range rowGroup.Box().Children {
		heights = append(heights, row.Box().Height.V())
		positionY = append(positionY, row.Box().PositionY)
		var cs, bs, pt, pb, py []pr.Float
		for _, cell := range row.Box().Children {
			cs = append(cs, cell.Box().Height.V())
			bs = append(bs, cell.Box().BorderHeight())
			pt = append(pt, cell.Box().PaddingTop.V())
			pb = append(pb, cell.Box().PaddingBottom.V())
			py = append(py, cell.Box().PositionY)
		}
		cells = append(cells, cs)
		borders = append(borders, bs)
		paddingTops = append(paddingTops, pt)
		paddingBottoms = append(paddingBottoms, pb)
		positionYCells = append(positionYCells, py)
	}
	tu.AssertEqual(t, heights, []pr.Float{80, 30, 0, 10}, "")
	tu.AssertEqual(t, positionY, []pr.Float{103, 283, 413, 513}, "") // cumulative sum of previous row heights && border-spacings

	tu.AssertEqual(t, cells, [][]pr.Float{
		{420, 60, 40, 20, 20, 20},
		{0, 10},
		nil,
		{0},
	}, "")
	tu.AssertEqual(t, borders, [][]pr.Float{
		{420, 80, 80, 80, 80, 80},
		{30, 30},
		nil,
		{10},
	}, "")
	// The baseline of the first row is at 40px because of the third column.
	// The second column thus gets a top padding of 20px pushes the bottom
	// to 80px.The middle is at 40px.
	tu.AssertEqual(t, paddingTops, [][]pr.Float{
		{0, 20, 0, 0, 30, 60},
		{15, 5},
		nil,
		{10},
	}, "")
	tu.AssertEqual(t, paddingBottoms, [][]pr.Float{
		{0, 0, 40, 60, 30, 0},
		{15, 15},
		nil,
		{0},
	}, "")
	tu.AssertEqual(t, positionYCells, [][]pr.Float{
		{103, 103, 103, 103, 103, 103},
		{283, 283},
		nil,
		{513},
	}, "")
}

func TestTableRowHeight2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// A cell box cannot extend beyond the last row box of a table.
	page := renderOnePage(t, `
      <table style="border-spacing: 0">
        <tr style="height: 10px">
          <td rowspan=5></td>
          <td></td>
        </tr>
        <tr style="height: 10px">
          <td></td>
        </tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0]
	_ = table.Box().Children[0]
}

func TestTableRowHeight3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
      </style>
      <table style="border-spacing: 0; font-family: weasyprint;
                    line-height: 20px">
        <tr><td>Table</td><td rowspan="2"></td></tr>
        <tr></tr>
      </table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0]
	tu.AssertEqual(t, table.Box().Height, pr.Float(20), "")
	rowGroup := table.Box().Children[0]
	tu.AssertEqual(t, rowGroup.Box().Height, pr.Float(20), "")
	row1, row2 := unpack2(rowGroup)
	tu.AssertEqual(t, row1.Box().Height, pr.Float(20), "")
	tu.AssertEqual(t, row2.Box().Height, pr.Float(0), "")
}

func TestTableWrapper(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 1000px }
        table { width: 600px; height: 500px; table-layout: fixed;
                  padding: 1px; border: 10px solid; margin: 100px; }
      </style>
      <table></table>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0]
	tu.AssertEqual(t, body.Box().Width, pr.Float(1000), "")
	tu.AssertEqual(t, wrapper.Box().Width, pr.Float(600), "") // Not counting borders || padding
	tu.AssertEqual(t, wrapper.Box().MarginLeft, pr.Float(100), "")
	tu.AssertEqual(t, table.Box().MarginWidth(), pr.Float(600), "")
	tu.AssertEqual(t, table.Box().Width, pr.Float(578), "") // 600 - 2*10 - 2*1, no margin
	// box-sizing := range the UA stylesheet  makes `height: 500px` set this
	tu.AssertEqual(t, table.Box().BorderHeight(), pr.Float(500), "")
	tu.AssertEqual(t, table.Box().Height, pr.Float(478), "")         // 500 - 2*10 - 2*1
	tu.AssertEqual(t, table.Box().MarginHeight(), pr.Float(500), "") // no margin
	tu.AssertEqual(t, wrapper.Box().Height, pr.Float(500), "")
	tu.AssertEqual(t, wrapper.Box().MarginHeight(), pr.Float(700), "") // 500 + 2*100
}

func TestTableHtmlTag(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Non-regression test: this used to cause an exception
	_ = renderOnePage(t, `<html style="display: table">`)
}

func TestTablePageBreaks(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range []struct {
		html      string
		rows      []int
		positions []pr.Float
	}{
		{
			`
			<style>
			  @page { size: 120px }
			  table { table-layout: fixed; width: 100% }
			  h1 { height: 30px }
			  td { height: 40px }
			</style>
			<h1>Dummy title</h1>
			<table>
			  <tr><td>row 1</td></tr>
			  <tr><td>row 2</td></tr>

			  <tr><td>row 3</td></tr>
			  <tr><td>row 4</td></tr>
			  <tr><td>row 5</td></tr>

			  <tr><td style="height: 300px"> <!-- overflow the page -->
				row 6</td></tr>
			  <tr><td>row 7</td></tr>
			  <tr><td>row 8</td></tr>
			</table>
		   `,
			[]int{2, 3, 1, 2},
			[]pr.Float{30, 70, 0, 40, 80, 0, 0, 40},
		},
		{
			`
			<style>
			  @page { size: 120px }
			  h1 { height: 30px}
			  td { height: 40px }
			  table { table-layout: fixed; width: 100%;
					  page-break-inside: avoid }
			</style>
			<h1>Dummy title</h1>
			<table>
			  <tr><td>row 1</td></tr>
			  <tr><td>row 2</td></tr>
			  <tr><td>row 3</td></tr>
			  <tr><td>row 4</td></tr>
		   </table>
		   `,
			[]int{0, 3, 1},
			[]pr.Float{0, 40, 80, 0},
		},
		{
			`
			<style>
			  @page { size: 120px }
			  h1 { height: 30px}
			  td { height: 40px }
			  table { table-layout: fixed; width: 100%;
					  page-break-inside: avoid }
			</style>
			<h1>Dummy title</h1>
			<table>
			  <tbody>
				<tr><td>row 1</td></tr>
				<tr><td>row 2</td></tr>
				<tr><td>row 3</td></tr>
			  </tbody>

			  <tr><td>row 4</td></tr>
			</table>
		   `,
			[]int{0, 3, 1},
			[]pr.Float{0, 40, 80, 0},
		},
		{
			`
			<style>
			  @page { size: 120px }
			  h1 { height: 30px}
			  td { height: 40px }
			  table { table-layout: fixed; width: 100% }
			</style>
			<h1>Dummy title</h1>
			<table>
			  <tr><td>row 1</td></tr>

			  <tbody style="page-break-inside: avoid">
				<tr><td>row 2</td></tr>
				<tr><td>row 3</td></tr>
			  </tbody>
			</table>
		   `,
			[]int{1, 2},
			[]pr.Float{30, 0, 40},
		},
		{
			`
		<style>
		  @page { size: 120px }
		  h1 { height: 30px}
		  td { line-height: 40px }
		  table { table-layout: fixed; width: 100%; orphans: 1; widows: 1 }
		</style>
		<h1>Dummy title</h1>
		<table>
		  <tr><td>r1l1</td></tr>
		  <tr style="break-inside: avoid"><td>r2l1<br>r2l2</td></tr>
		  <tr><td>r3l1</td></tr>
		</table>
	   `,
			[]int{1, 2},
			[]pr.Float{30, 0, 80},
		},
		{
			`
			<style>
			  @page { size: 120px }
			  h1 { height: 30px}
			  td { line-height: 40px }
			  table { table-layout: fixed; width: 100%; orphans: 1; widows: 1 }
			</style>
			<h1>Dummy title</h1>
			<table>
			  <tbody>
				<tr><td>r1l1</td></tr>
				<tr style="break-inside: avoid"><td>r2l1<br>r2l2</td></tr>
			  </tbody>
			  <tr><td>r3l1</td></tr>
			</table>
		   `,
			[]int{1, 2},
			[]pr.Float{30, 0, 80},
		},
	} {
		pages := renderPages(t, data.html)
		var (
			rowsPerPage   []int
			rowsPositionY []pr.Float
		)
		for i, page := range pages {
			html := page.Box().Children[0]
			body := html.Box().Children[0]
			bodyChildren := body.Box().Children
			if i == 0 {
				bodyChildren = bodyChildren[1:] // skip h1
			}

			if len(bodyChildren) == 0 {
				rowsPerPage = append(rowsPerPage, 0)
				continue
			}
			tableWrapper := bodyChildren[0]
			table := tableWrapper.Box().Children[0]
			rowsInThisPage := 0
			for _, group := range table.Box().Children {
				tu.AssertEqual(t, len(group.Box().Children) > 0, true, "found an empty table group")
				for _, row := range group.Box().Children {
					rowsInThisPage += 1
					rowsPositionY = append(rowsPositionY, row.Box().PositionY)
					_ = row.Box().Children[0]
				}
			}
			rowsPerPage = append(rowsPerPage, rowsInThisPage)
		}

		tu.AssertEqual(t, rowsPerPage, data.rows, "number of rows")
		tu.AssertEqual(t, rowsPositionY, data.positions, "positions")
	}
}

func TestTablePageBreaksComplex1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 100px }
      </style>
      <h1 style="margin: 0; height: 30px">Lipsum</h1>
      <!-- Leave 70px on the first page: enough for the header || row1
           but ! both.  -->
      <table style="border-spacing: 0; font-size: 5px">
        <thead>
          <tr><td style="height: 20px">Header</td></tr>
        </thead>
        <tbody>
          <tr><td style="height: 60px">Row 1</td></tr>
          <tr><td style="height: 10px">Row 2</td></tr>
          <tr><td style="height: 50px">Row 3</td></tr>
          <tr><td style="height: 61px">Row 4</td></tr>
          <tr><td style="height: 90px">Row 5</td></tr>
        </tbody>
        <tfoot>
          <tr><td style="height: 20px">Footer</td></tr>
        </tfoot>
      </table>
    `)
	var rowsPerPage [][][]string
	for i, page := range pages {
		var groups [][]string
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		tableWrapper := body.Box().Children[0]
		if i == 0 {
			tu.AssertEqual(t, tableWrapper.Box().ElementTag(), "h1", "")
		} else {
			table := tableWrapper.Box().Children[0]
			for _, group := range table.Box().Children {
				tu.AssertEqual(t, len(group.Box().Children) > 0, true, "found an empty table group")
				var rows []string
				for _, row := range group.Box().Children {
					cell := row.Box().Children[0]
					line := cell.Box().Children[0]
					text := line.Box().Children[0]
					rows = append(rows, text.(*bo.TextBox).Text)
				}
				groups = append(groups, rows)
			}
		}
		rowsPerPage = append(rowsPerPage, groups)
	}
	tu.AssertEqual(t, rowsPerPage, [][][]string{
		nil,
		{{"Header"}, {"Row 1"}, {"Footer"}},
		{{"Header"}, {"Row 2", "Row 3"}, {"Footer"}},
		{{"Header"}, {"Row 4"}},
		{{"Row 5"}},
	}, "")
}

func TestTablePageBreaksComplex2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 250px }
        td { height: 40px }
        table { table-layout: fixed; width: 100%; break-before: avoid }
      </style>
      <table>
        <thead>
          <tr><td>head 1</td></tr>
        </thead>
        <tbody>
          <tr><td>row 1 1</td></tr>
          <tr><td>row 1 2</td></tr>
          <tr><td>row 1 3</td></tr>
        </tbody>
        <tfoot>
          <tr><td>foot 1</td></tr>
        </tfoot>
      </table>
      <table>
        <thead>
          <tr><td>head 2</td></tr>
        </thead>
        <tbody>
          <tr><td>row 2 1</td></tr>
          <tr><td>row 2 2</td></tr>
          <tr><td>row 2 3</td></tr>
        </tbody>
        <tfoot>
          <tr><td>foot 2</td></tr>
        </tfoot>
      </table>
     `)
	var rowsPerPage [][][]string
	for _, page := range pages {
		var groups [][]string
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		for _, tableWrapper := range body.Box().Children {
			table := tableWrapper.Box().Children[0]
			for _, group := range table.Box().Children {
				tu.AssertEqual(t, len(group.Box().Children) > 0, true, "found an empty table group")
				var rows []string
				for _, row := range group.Box().Children {
					cell := row.Box().Children[0]
					line := cell.Box().Children[0]
					text := line.Box().Children[0]
					rows = append(rows, text.(*bo.TextBox).Text)
				}
				groups = append(groups, rows)
			}
		}
		rowsPerPage = append(rowsPerPage, groups)
	}
	tu.AssertEqual(t, rowsPerPage, [][][]string{
		{{"head 1"}, {"row 1 1", "row 1 2"}, {"foot 1"}},
		{{"head 1"}, {"row 1 3"}, {"foot 1"}, {"head 2"}, {"row 2 1"}, {"foot 2"}},
		{{"head 2"}, {"row 2 2", "row 2 3"}, {"foot 2"}},
	}, "")
	// TODO: test positions, the place of footer on the first page is wrong
}

func TestTablePageBreakAfter(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 1000px }
        h1 { height: 30px}
        td { height: 40px }
        table { table-layout: fixed; width: 100% }
      </style>
      <h1>Dummy title</h1>
      <table>

        <tbody>
          <tr><td>row 1</td></tr>
          <tr><td>row 2</td></tr>
          <tr><td>row 3</td></tr>
        </tbody>
        <tbody>
          <tr style="break-after: page"><td>row 1</td></tr>
          <tr><td>row 2</td></tr>
          <tr><td>row 3</td></tr>
        </tbody>
        <tbody>
          <tr><td>row 1</td></tr>
          <tr><td>row 2</td></tr>
          <tr style="break-after: page"><td>row 3</td></tr>
        </tbody>
        <tbody style="break-after: right">
          <tr><td>row 1</td></tr>
          <tr><td>row 2</td></tr>
          <tr><td>row 3</td></tr>
        </tbody>
        <tbody style="break-after: page">
          <tr><td>row 1</td></tr>
          <tr><td>row 2</td></tr>
          <tr><td>row 3</td></tr>
        </tbody>

      </table>
      <p>bla bla</p>
     `)
	page1, page2, page3, page4, page5, page6 := pages[0], pages[1], pages[2], pages[3], pages[4], pages[5]
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	_, tableWrapper := unpack2(body)
	table := tableWrapper.Box().Children[0]
	tableGroup1, tableGroup2 := unpack2(table)
	tu.AssertEqual(t, len(tableGroup1.Box().Children), 3, "")
	tu.AssertEqual(t, len(tableGroup2.Box().Children), 1, "")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	tableWrapper = body.Box().Children[0]
	table = tableWrapper.Box().Children[0]
	tableGroup1, tableGroup2 = unpack2(table)
	tu.AssertEqual(t, len(tableGroup1.Box().Children), 2, "")
	tu.AssertEqual(t, len(tableGroup2.Box().Children), 3, "")

	html = page3.Box().Children[0]
	body = html.Box().Children[0]
	tableWrapper = body.Box().Children[0]
	table = tableWrapper.Box().Children[0]
	tableGroup := table.Box().Children[0]
	tu.AssertEqual(t, len(tableGroup.Box().Children), 3, "")

	html = page4.Box().Children[0]
	tu.AssertEqual(t, len(html.Box().Children), 0, "")

	html = page5.Box().Children[0]
	body = html.Box().Children[0]
	tableWrapper = body.Box().Children[0]
	table = tableWrapper.Box().Children[0]
	tableGroup = table.Box().Children[0]
	tu.AssertEqual(t, len(tableGroup.Box().Children), 3, "")

	html = page6.Box().Children[0]
	body = html.Box().Children[0]
	p := body.Box().Children[0]
	tu.AssertEqual(t, p.Box().ElementTag(), "p", "")
}

func TestTablePageBreakBefore(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 1000px }
        h1 { height: 30px}
        td { height: 40px }
        table { table-layout: fixed; width: 100% }
      </style>
      <h1>Dummy title</h1>
      <table>
 
        <tbody>
          <tr style="break-before: page"><td>row 1</td></tr>
          <tr><td>row 2</td></tr>
          <tr><td>row 3</td></tr>
        </tbody>
        <tbody>
          <tr><td>row 1</td></tr>
          <tr style="break-before: page"><td>row 2</td></tr>
          <tr><td>row 3</td></tr>
        </tbody>
        <tbody>
          <tr style="break-before: page"><td>row 1</td></tr>
          <tr><td>row 2</td></tr>
          <tr><td>row 3</td></tr>
        </tbody>
        <tbody>
          <tr><td>row 1</td></tr>
          <tr><td>row 2</td></tr>
          <tr><td>row 3</td></tr>
        </tbody>
        <tbody style="break-before: left">
          <tr><td>row 1</td></tr>
          <tr><td>row 2</td></tr>
          <tr><td>row 3</td></tr>
        </tbody>

      </table>
      <p>bla bla</p>
     `)
	page1, page2, page3, page4, page5, page6 := pages[0], pages[1], pages[2], pages[3], pages[4], pages[5]

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	h1 := body.Box().Children[0]
	tu.AssertEqual(t, h1.Box().ElementTag(), "h1", "")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0]
	tableGroup1, tableGroup2 := unpack2(table)
	tu.AssertEqual(t, len(tableGroup1.Box().Children), 3, "")
	tu.AssertEqual(t, len(tableGroup2.Box().Children), 1, "")

	html = page3.Box().Children[0]
	body = html.Box().Children[0]
	tableWrapper = body.Box().Children[0]
	table = tableWrapper.Box().Children[0]
	tableGroup := table.Box().Children[0]
	tu.AssertEqual(t, len(tableGroup.Box().Children), 2, "")

	html = page4.Box().Children[0]
	body = html.Box().Children[0]
	tableWrapper = body.Box().Children[0]
	table = tableWrapper.Box().Children[0]
	tableGroup1, tableGroup2 = unpack2(table)
	tu.AssertEqual(t, len(tableGroup1.Box().Children), 3, "")
	tu.AssertEqual(t, len(tableGroup2.Box().Children), 3, "")

	html = page5.Box().Children[0]
	tu.AssertEqual(t, len(html.Box().Children), 0, "")

	html = page6.Box().Children[0]
	body = html.Box().Children[0]
	tableWrapper, p := unpack2(body)
	table = tableWrapper.Box().Children[0]
	tableGroup = table.Box().Children[0]
	tu.AssertEqual(t, len(tableGroup.Box().Children), 3, "")
	tu.AssertEqual(t, p.Box().ElementTag(), "p", "")
}

func TestTablePageBreakAvoid(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range []struct {
		html string
		rows []int
	}{
		{
			`
		<style>
		  @page { size: 100px }
		  table { table-layout: fixed; width: 100% }
		  tr { height: 26px }
		</style>
		<table>
		  <tbody>
			<tr><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr style="break-before: avoid"><td>row 2</td></tr>
			<tr style="break-before: avoid"><td>row 3</td></tr>
		  </tbody>
		</table>
	  `,
			[]int{1, 3},
		},
		{
			`
		<style>
		  @page { size: 100px }
		  table { table-layout: fixed; width: 100% }
		  tr { height: 26px }
		</style>
		<table>
		  <tbody>
			<tr><td>row 0</td></tr>
			<tr style="break-after: avoid"><td>row 1</td></tr>
			<tr><td>row 2</td></tr>
			<tr style="break-before: avoid"><td>row 3</td></tr>
		  </tbody>
		</table>
	   `,
			[]int{1, 3},
		},
		{
			`
		<style>
		  @page { size: 100px }
		  table { table-layout: fixed; width: 100% }
		  tr { height: 26px }
		</style>
		<table>
		  <tbody>
			<tr><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr style="break-after: avoid"><td>row 2</td></tr>
			<tr><td>row 3</td></tr>
		  </tbody>
		</table>
	   `,
			[]int{2, 2},
		},
		{
			`
		<style>
		  @page { size: 100px }
		  table { table-layout: fixed; width: 100% }
		  tr { height: 26px }
		</style>
		<table>
		  <tbody>
			<tr style="break-before: avoid"><td>row 0</td></tr>
			<tr style="break-before: avoid"><td>row 1</td></tr>
			<tr style="break-before: avoid"><td>row 2</td></tr>
			<tr style="break-before: avoid"><td>row 3</td></tr>
		  </tbody>
		</table>
	   `,
			[]int{3, 1},
		},
		{
			`
		<style>
		  @page { size: 100px }
		  table { table-layout: fixed; width: 100% }
		  tr { height: 26px }
		</style>
		<table>
		  <tbody>
			<tr style="break-after: avoid"><td>row 0</td></tr>
			<tr style="break-after: avoid"><td>row 1</td></tr>
			<tr style="break-after: avoid"><td>row 2</td></tr>
			<tr style="break-after: avoid"><td>row 3</td></tr>
		  </tbody>
		</table>
	   `,
			[]int{3, 1},
		},
		{
			`
		<style>
		  @page { size: 100px }
		  table { table-layout: fixed; width: 100% }
		  tr { height: 26px }
		  p { height: 26px }
		</style>
		<p>wow p</p>
		<table>
		  <tbody>
			<tr style="break-after: avoid"><td>row 0</td></tr>
			<tr style="break-after: avoid"><td>row 1</td></tr>
			<tr style="break-after: avoid"><td>row 2</td></tr>
			<tr style="break-after: avoid"><td>row 3</td></tr>
		  </tbody>
		</table>
	   `,
			[]int{1, 3, 1},
		},
		{
			`
		<style>
		  @page { size: 100px }
		  table { table-layout: fixed; width: 100% }
		  tr { height: 30px }
		</style>
		<table>
		  <tbody style="break-after: avoid">
			<tr><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr><td>row 2</td></tr>
		  </tbody>
		  <tbody>
			<tr><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr><td>row 2</td></tr>
		  </tbody>
		</table>
	   `,
			[]int{2, 3, 1},
		},
		{
			`
		<style>
		  @page { size: 100px }
		  table { table-layout: fixed; width: 100% }
		  tr { height: 30px }
		</style>
		<table>
		  <tbody>
			<tr><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr><td>row 2</td></tr>
		  </tbody>
		  <tbody style="break-before: avoid">
			<tr><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr><td>row 2</td></tr>
		  </tbody>
		</table>
	   `,
			[]int{2, 3, 1},
		},
		{
			`
		<style>
		  @page { size: 100px }
		  table { table-layout: fixed; width: 100% }
		  tr { height: 30px }
		</style>
		<table>
		  <tbody>
			<tr><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr><td>row 2</td></tr>
		  </tbody>
		  <tbody>
			<tr style="break-before: avoid"><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr><td>row 2</td></tr>
		  </tbody>
		</table>
	   `,
			[]int{2, 3, 1},
		},
		{
			`
		<style>
		  @page { size: 100px }
		  table { table-layout: fixed; width: 100% }
		  tr { height: 30px }
		</style>
		<table>
		  <tbody>
			<tr><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr style="break-after: avoid"><td>row 2</td></tr>
		  </tbody>
		  <tbody>
			<tr><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr><td>row 2</td></tr>
		  </tbody>
		</table>
	   `,
			[]int{2, 3, 1},
		},
		{
			`
		<style>
		  @page { size: 100px }
		  table { table-layout: fixed; width: 100% }
		  tr { height: 30px }
		</style>
		<table>
		  <tbody style="break-after: avoid">
			<tr><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr style="break-after: page"><td>row 2</td></tr>
		  </tbody>
		  <tbody>
			<tr><td>row 0</td></tr>
			<tr><td>row 1</td></tr>
			<tr><td>row 2</td></tr>
		  </tbody>
		</table>
	   `,
			[]int{3, 3},
		},
	} {
		testTablePageBreakAvoid(t, data.html, data.rows)
	}
}

func testTablePageBreakAvoid(t *testing.T, html string, rows []int) {
	pages := renderPages(t, html)
	tu.AssertEqual(t, len(pages), len(rows), "")
	var rowsPerPage []int
	for _, page := range pages {
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		if body.Box().Children[0].Box().ElementTag() == "p" {
			rowsPerPage = append(rowsPerPage, len(body.Box().Children))
			continue
		}

		tableWrapper := body.Box().Children[0]
		table := tableWrapper.Box().Children[0]
		rowsInThisPage := 0
		for _, group := range table.Box().Children {
			rowsInThisPage += len(group.Box().Children)
		}
		rowsPerPage = append(rowsPerPage, rowsInThisPage)
	}

	tu.AssertEqual(t, rowsPerPage, rows, "")
}

func TestInlineTableBaseline(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range []struct {
		verticalAlign  string
		tablePositionY pr.Float
	}{
		{"top", 8},
		{"bottom", 8},
		{"baseline", 10},
	} {
		// Check that inline table's baseline is its first row's baseline.
		// Div text's baseline is at 18px from the top (10px because of the
		// line-height, 8px as it's weasyprint.otf's baseline position).
		// When a row has vertical-align: baseline cells, its baseline is its cell's
		// baseline. The position of the table is thus 10px above the text's
		// baseline.
		// When a row has another value for vertical-align, the baseline is the
		// bottom of the row. The first cell's text is aligned with the div's text,
		// and the top of the table is thus 8px above the baseline.
		page := renderOnePage(t, fmt.Sprintf(`
		<style>
			@font-face { src: url(weasyprint.otf); font-family: weasyprint }
		</style>
		<div style="font-family: weasyprint; font-size: 10px; line-height: 30px">
			abc
			<table style="display: inline-table; border-collapse: collapse;
						line-height: 10px">
			<tr><td style="vertical-align: %s">a</td></tr>
			<tr><td>a</td></tr>
			</table>
			abc
		</div>
		`, data.verticalAlign))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		div := body.Box().Children[0]
		line := div.Box().Children[0]
		text1, tableWrapper, text2 := unpack3(line)
		table := tableWrapper.Box().Children[0]
		tu.AssertEqual(t, text1.Box().PositionY, text2.Box().PositionY, "")
		tu.AssertEqual(t, text2.Box().PositionY, pr.Float(0), "")
		tu.AssertEqual(t, table.Box().Height, pr.Float(10)*pr.Float(2), "")
		tu.AssertEqual(t, pr.Abs(table.Box().PositionY-data.tablePositionY) < pr.Float(0.1), true, "")
	}
}

func TestTableCaptionMarginTop(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        table { margin: 20px; }
        caption, h1, h2 { margin: 20px; height: 10px }
        td { height: 10px }
      </style>
      <h1></h1>
      <table>
        <caption></caption>
        <tr>
          <td></td>
        </tr>
      </table>
      <h2></h2>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	h1, wrapper, h2 := unpack3(body)
	caption, table := unpack2(wrapper)
	tbody := table.Box().Children[0]
	tu.AssertEqual(t, [2]pr.Float{h1.Box().ContentBoxX(), h1.Box().ContentBoxY()}, [2]pr.Float{20, 20}, "")
	tu.AssertEqual(t, [2]pr.Float{wrapper.Box().ContentBoxX(), wrapper.Box().ContentBoxY()}, [2]pr.Float{20, 50}, "")
	tu.AssertEqual(t, [2]pr.Float{caption.Box().ContentBoxX(), caption.Box().ContentBoxY()}, [2]pr.Float{40, 70}, "")
	tu.AssertEqual(t, [2]pr.Float{tbody.Box().ContentBoxX(), tbody.Box().ContentBoxY()}, [2]pr.Float{20, 100}, "")
	tu.AssertEqual(t, [2]pr.Float{h2.Box().ContentBoxX(), h2.Box().ContentBoxY()}, [2]pr.Float{20, 130}, "")
}

func TestTableCaptionMarginBottom(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        table { margin: 20px; }
        caption, h1, h2 { margin: 20px; height: 10px; caption-side: bottom }
        td { height: 10px }
      </style>
      <h1></h1>
      <table>
        <caption></caption>
        <tr>
          <td></td>
        </tr>
      </table>
      <h2></h2>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	h1, wrapper, h2 := unpack3(body)
	table, caption := unpack2(wrapper)
	tbody := table.Box().Children[0]
	tu.AssertEqual(t, [2]pr.Float{h1.Box().ContentBoxX(), h1.Box().ContentBoxY()}, [2]pr.Float{20, 20}, "")
	tu.AssertEqual(t, [2]pr.Float{wrapper.Box().ContentBoxX(), wrapper.Box().ContentBoxY()}, [2]pr.Float{20, 50}, "")
	tu.AssertEqual(t, [2]pr.Float{tbody.Box().ContentBoxX(), tbody.Box().ContentBoxY()}, [2]pr.Float{20, 50}, "")
	tu.AssertEqual(t, [2]pr.Float{caption.Box().ContentBoxX(), caption.Box().ContentBoxY()}, [2]pr.Float{40, 80}, "")
	tu.AssertEqual(t, [2]pr.Float{h2.Box().ContentBoxX(), h2.Box().ContentBoxY()}, [2]pr.Float{20, 130}, "")
}

func TestTableEmptyBody(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for testIndex, data := range []struct {
		rowsExpected [][]string
		thead, tfoot int
		content      string
	}{
		{[][]string{nil, {"Header", "Footer"}}, 45, 45, "<p>content</p>"},
		{[][]string{nil, {"Header", "Footer"}}, 85, 5, "<p>content</p>"},
		{[][]string{{"Header", "Footer"}}, 30, 30, "<p>content</p>"},
		{[][]string{nil, {"Header"}}, 30, 110, "<p>content</p>"},
		{[][]string{nil, {"Header", "Footer"}}, 30, 60, "<p>content</p>"},
		{[][]string{nil, {"Footer"}}, 110, 30, "<p>content</p>"},

		// We try to render the header and footer on the same page, but it does not
		// fit. So we try to render the header or the footer on the next one, but
		// nothing fit either.
		{[][]string{nil, nil}, 110, 110, "<p>content</p>"},

		{[][]string{{"Header", "Footer"}}, 30, 30, ""},
		{[][]string{{"Header"}}, 30, 110, ""},
		{[][]string{{"Header", "Footer"}}, 30, 60, ""},
		{[][]string{{"Footer"}}, 110, 30, ""},
		{[][]string{nil}, 110, 110, ""},
	} {
		html := fmt.Sprintf(`
      <style>
        @page { size: 100px }
        p { height: 20px }
        thead th { height: %dpx }
        tfoot th { height: %dpx }
      </style>
      %s
      <table>
        <thead><tr><th>Header</th></tr></thead>
        <tfoot><tr><th>Footer</th></tr></tfoot>
      </table>
    `, data.thead, data.tfoot, data.content)
		pages := renderPages(t, html)
		tu.AssertEqual(t, len(pages), len(data.rowsExpected), fmt.Sprintf("number of pages for test %d", testIndex))
		for i, page := range pages {
			var rows []string
			html := page.Box().Children[0]
			body := html.Box().Children[0]
			tableWrapper := body.Box().Children[len(body.Box().Children)-1]
			if !tableWrapper.Box().IsTableWrapper {
				tu.AssertEqual(t, rows, data.rowsExpected[i], "")
				continue
			}
			table := tableWrapper.Box().Children[0]
			for _, group := range table.Box().Children {
				for _, row := range group.Box().Children {
					cell := row.Box().Children[0]
					line := cell.Box().Children[0]
					text := line.Box().Children[0]
					rows = append(rows, text.(*bo.TextBox).Text)
				}
			}
			tu.AssertEqual(t, rows, data.rowsExpected[i], "")
		}
	}
}

func TestTableBreakChildrenMargin(t *testing.T) {
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1254
	html := `
      <style>
        @page { size: 100px }
        p { height: 20px; margin-top: 50px }
      </style>
      <table>
        <tr><td><p>Page1</p></td></tr>
        <tr><td><p>Page2</p></td></tr>
        <tr><td><p>Page3</p></td></tr>
      </table>
    `
	tu.AssertEqual(t, len(renderPages(t, html)), 3, "")
}

func TestTableTdBreakInsideAvoid(t *testing.T) {
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1547
	html := `
	<style>
	@page { size: 4cm }
	td { break-inside: avoid; line-height: 3cm }
  </style>
  <table>
	<tr>
	  <td>
		a<br>a
	  </td>
	</tr>
  </table>
    `
	tu.AssertEqual(t, len(renderPages(t, html)), 2, "")
}
