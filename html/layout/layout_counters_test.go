package layout

import (
	"fmt"
	"testing"

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
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	html := unpack1(page)
	body := unpack1(html)
	ol := unpack1(body).Box()
	li1, li2, li3, li4 := ol.Children[0], ol.Children[1], ol.Children[2], ol.Children[3]
	assertText(t, unpack1(li1.Box().Children[0].Box().Children[0]), values[0])
	assertText(t, unpack1(li2.Box().Children[0].Box().Children[0]), values[1])
	assertText(t, unpack1(li3.Box().Children[0].Box().Children[0]), values[2])
	assertText(t, unpack1(li4.Box().Children[0].Box().Children[0]), values[3])
}

func TestCounterSet(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	html := unpack1(page)
	body := unpack1(html)
	chs := body.Box().Children
	art1, art2, art3, art4, art5, art6 := chs[0], chs[1], chs[2], chs[3], chs[4], chs[5]

	h1 := unpack1(art1)
	assertText(t, unpack1(unpack1(unpack1(h1))), "1")

	h2, h3 := unpack1(art2), art2.Box().Children[1]
	assertText(t, unpack1(unpack1(unpack1(h2))), "3")
	assertText(t, unpack1(unpack1(unpack1(h3))), "4")

	h3 = unpack1(art3)
	assertText(t, unpack1(unpack1(unpack1(h3))), "5")

	h2 = unpack1(art4)
	assertText(t, unpack1(unpack1(unpack1(h2))), "3")

	h31, h32 := unpack1(art5), art5.Box().Children[1]
	assertText(t, unpack1(unpack1(unpack1(h31))), "4")
	assertText(t, unpack1(unpack1(unpack1(h32))), "5")

	h1, h2, h3 = unpack1(art6), art6.Box().Children[1], art6.Box().Children[2]
	assertText(t, unpack1(unpack1(unpack1(h1))), "1")
	assertText(t, unpack1(unpack1(unpack1(h2))), "3")
	assertText(t, unpack1(unpack1(unpack1(h3))), "4")
}

func TestCounterMultipleExtends(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	html := unpack1(page)
	body := unpack1(html)
	olC := unpack1(body).Box().Children
	li1, li2, li3, li4, li5, li6, li7, li8 := olC[0], olC[1], olC[2], olC[3], olC[4], olC[5], olC[6], olC[7]
	assertText(t, unpack1(unpack1(unpack1(li1))), "a1b")
	assertText(t, unpack1(unpack1(unpack1(li2))), "2b")
	assertText(t, unpack1(unpack1(unpack1(li3))), "c3. ")
	assertText(t, unpack1(unpack1(unpack1(li4))), "d4. ")
	assertText(t, unpack1(unpack1(unpack1(li5))), "e5. ")
	assertText(t, unpack1(unpack1(unpack1(li6))), "6. ")
	assertText(t, unpack1(unpack1(unpack1(li7))), "7. ")
	assertText(t, unpack1(unpack1(unpack1(li8))), "8. ")
}

func TestCounters9(t *testing.T) {
	// See https://github.com/Kozea/WeasyPrint/issues/1685
	t.Skip("nested counters are broken")

	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
		  <ol>
			<li></li>
			<li>
			  <ol style="counter-reset: a">
				<li></li>
				<li></li>
			  </ol>
			</li>
			<li></li>
		  </ol>
		`)
	html := unpack1(page)
	body := unpack1(html)
	ol1 := unpack1(body)
	oli1, oli2, oli3 := unpack3(ol1)
	_, ol2 := unpack2(oli2)
	oli21, oli22 := unpack2(ol2)
	assertText(t, unpack1(oli1.Box().Children[0].Box().Children[0]), "1. ")
	assertText(t, unpack1(oli2.Box().Children[0].Box().Children[0]), "2. ")
	assertText(t, unpack1(oli21.Box().Children[0].Box().Children[0]), "1. ")
	assertText(t, unpack1(oli22.Box().Children[0].Box().Children[0]), "2. ")
	assertText(t, unpack1(oli3.Box().Children[0].Box().Children[0]), "3. ")
}
