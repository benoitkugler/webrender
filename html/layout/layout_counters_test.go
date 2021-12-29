package layout

import (
	"fmt"
	"testing"

	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func TestTestCounterSymbols(t *testing.T) {
	for _, arg := range []struct {
		argument string
		values   [4]string
	}{
		{argument: `symbols(cyclic "a" "b" "c")`, values: [4]string{"a ", "b ", "c ", "a "}},
		{argument: `symbols(symbolic "a" "b")`, values: [4]string{"a ", "b ", "aa ", "bb "}},
		{argument: `symbols("a" "b")`, values: [4]string{"a ", "b ", "aa ", "bb "}},
		{argument: `symbols(alphabetic "a" "b")`, values: [4]string{"a ", "b ", "aa ", "ab "}},
		{argument: `symbols(fixed "a" "b")`, values: [4]string{"a ", "b ", "3 ", "4 "}},
		{argument: `symbols(numeric "0" "1" "2")`, values: [4]string{"1 ", "2 ", "10 ", "11 "}},
		{argument: `decimal`, values: [4]string{"1. ", "2. ", "3. ", "4. "}},
		{argument: `"/"`, values: [4]string{"/", "/", "/", "/"}},
	} {
		testCounterSymbols(t, arg.argument, arg.values)
	}
}

func testCounterSymbols(t *testing.T, arguments string, values [4]string) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, fmt.Sprintf(`
      <style>
        ol { list-style-type: %s }
      </style>
      <ol>
        <li>abc</li>
        <li>abc</li>
        <li>abc</li>
        <li>abc</li>
      </ol>
    `, arguments))

	html := page.Box().Children[0]
	body := html.Box().Children[0]
	ol := body.Box().Children[0].Box()
	li1, li2, li3, li4 := ol.Children[0], ol.Children[1], ol.Children[2], ol.Children[3]
	tu.AssertEqual(t, li1.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, values[0], fmt.Sprintf("symbols %s and item %d", arguments, 0))
	tu.AssertEqual(t, li2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, values[1], fmt.Sprintf("symbols %s and item %d", arguments, 1))
	tu.AssertEqual(t, li3.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, values[2], fmt.Sprintf("symbols %s and item %d", arguments, 2))
	tu.AssertEqual(t, li4.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, values[3], fmt.Sprintf("symbols %s and item %d", arguments, 3))
}

func TestCounterSet(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        body { counter-reset: h2 0 h3 4; font-size: 1px }
        article { counter-reset: h2 2 }
        h1 { counter-increment: h1 }
        h1::before { content: counter(h1) }
        h2 { counter-increment: h2; counter-set: h3 3 }
        h2::before { content: counter(h2) }
        h3 { counter-increment: h3 }
        h3::before { content: counter(h3) }
      </style>
      <article>
        <h1></h1>
      </article>
      <article>
        <h2></h2>
        <h3></h3>
      </article>
      <article>
        <h3></h3>
      </article>
      <article>
        <h2></h2>
      </article>
      <article>
        <h3></h3>
        <h3></h3>
      </article>
      <article>
        <h1></h1>
        <h2></h2>
        <h3></h3>
      </article>
    `)

	html := page.Box().Children[0]
	body := html.Box().Children[0]
	chs := body.Box().Children
	art1, art2, art3, art4, art5, art6 := chs[0], chs[1], chs[2], chs[3], chs[4], chs[5]

	h1 := art1.Box().Children[0]
	tu.AssertEqual(t, h1.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "1", "h1")

	h2, h3 := art2.Box().Children[0], art2.Box().Children[1]
	tu.AssertEqual(t, h2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "3", "h2")
	tu.AssertEqual(t, h3.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "4", "h3")

	h3 = art3.Box().Children[0]
	tu.AssertEqual(t, h3.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "5", "h3")

	h2 = art4.Box().Children[0]
	tu.AssertEqual(t, h2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "3", "h2")

	h31, h32 := art5.Box().Children[0], art5.Box().Children[1]
	tu.AssertEqual(t, h31.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "4", "h31")
	tu.AssertEqual(t, h32.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "5", "h32")

	h1, h2, h3 = art6.Box().Children[0], art6.Box().Children[1], art6.Box().Children[2]
	tu.AssertEqual(t, h1.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "1", "h1")
	tu.AssertEqual(t, h2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "3", "h2")
	tu.AssertEqual(t, h3.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "4", "h3")
}

func TestCounterMultipleExtends(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Inspired by W3C failing test system-extends-invalid
	page := renderOnePage(t, `
      <style>
        @counter-style a {
          system: extends b;
          prefix: a;
        }
        @counter-style b {
          system: extends c;
          suffix: b;
        }
        @counter-style c {
          system: extends b;
          pad: 2 c;
        }
        @counter-style d {
          system: extends d;
          prefix: d;
        }
        @counter-style e {
          system: extends unknown;
          prefix: e;
        }
        @counter-style f {
          system: extends decimal;
          symbols: a;
        }
        @counter-style g {
          system: extends decimal;
          additive-symbols: 1 a;
        }
      </style>
      <ol>
        <li style="list-style-type: a"></li>
        <li style="list-style-type: b"></li>
        <li style="list-style-type: c"></li>
        <li style="list-style-type: d"></li>
        <li style="list-style-type: e"></li>
        <li style="list-style-type: f"></li>
        <li style="list-style-type: g"></li>
        <li style="list-style-type: h"></li>
      </ol>
    `)

	html := page.Box().Children[0]
	body := html.Box().Children[0]
	olC := body.Box().Children[0].Box().Children
	li1, li2, li3, li4, li5, li6, li7, li8 := olC[0], olC[1], olC[2], olC[3], olC[4], olC[5], olC[6], olC[7]
	tu.AssertEqual(t, li1.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "a1b", "li1")
	tu.AssertEqual(t, li2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "2b", "li2")
	tu.AssertEqual(t, li3.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "c3. ", "li3")
	tu.AssertEqual(t, li4.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "d4. ", "li4")
	tu.AssertEqual(t, li5.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "e5. ", "li5")
	tu.AssertEqual(t, li6.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "6. ", "li6")
	tu.AssertEqual(t, li7.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "7. ", "li7")
	tu.AssertEqual(t, li8.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "8. ", "li8")
}
