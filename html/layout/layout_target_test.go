package layout

import (
	"testing"

	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Test the CSS cross references using target-*() functions.

func TestTargetCounter(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        div:first-child { counter-reset: div }
        div { counter-increment: div }
        #id1::before { content: target-counter("#id4", div) }
        #id2::before { content: "test " target-counter("#id1" div) }
        #id3::before { content: target-counter(url(#id4), div, lower-roman) }
        #id4::before { content: target-counter("#id3", div) }
      </style>
      <body>
        <div id="id1"></div>
        <div id="id2"></div>
        <div id="id3"></div>
        <div id="id4"></div>
    `)

	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div1, div2, div3, div4 := unpack4(body)
	before := div1.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "4", "before")
	before = div2.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "test 1", "before")
	before = div3.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "iv", "before")
	before = div4.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "3", "before")
}

func TestTargetCounterAttr(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        div:first-child { counter-reset: div }
        div { counter-increment: div }
        div::before { content: target-counter(attr(data-count), div) }
        #id2::before { content: target-counter(attr(data-count, url), div) }
        #id4::before {
          content: target-counter(attr(data-count), div, lower-alpha) }
      </style>
      <body>
        <div id="id1" data-count="#id4"></div>
        <div id="id2" data-count="#id1"></div>
        <div id="id3" data-count="#id2"></div>
        <div id="id4" data-count="#id3"></div>
    `)

	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div1, div2, div3, div4 := unpack4(body)
	before := div1.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "4", "before")
	before = div2.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "1", "before")
	before = div3.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "2", "before")
	before = div4.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "c", "before")
}

func TestTargetCounters(t *testing.T) {
	// cp := tu.CaptureLogs()
	// defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        div:first-child { counter-reset: div }
        div { counter-increment: div }
        #id1-2::before { content: target-counters("#id4-2", div, ".") }
        #id2-1::before { content: target-counters(url(#id3), div, "++") }
        #id3::before {
          content: target-counters("#id2-1", div, ".", lower-alpha) }
        #id4-2::before {
          content: target-counters(attr(data-count, url), div, "") }
      </style>
      <body>
        <div id="id1"><div></div><div id="id1-2"></div></div>
        <div id="id2"><div id="id2-1"></div><div></div></div>
        <div id="id3"></div>
        <div id="id4">
          <div></div><div id="id4-2" data-count="#id1-2"></div>
        </div>
    `)

	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div1, div2, div3, div4 := unpack4(body)
	before := div1.Box().Children[1].Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "4.2", "before")
	before = div2.Box().Children[0].Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "3", "before")
	before = div3.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "b.a", "before")
	before = div4.Box().Children[1].Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "12", "before")
}

func TestTargetText(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        a { display: block; color: red }
        div:first-child { counter-reset: div }
        div { counter-increment: div }
        #id2::before { content: "wow" }
        #link1::before { content: "test " target-text("#id4") }
        #link2::before { content: target-text(attr(data-count, url), before) }
        #link3::before { content: target-text("#id3", after) }
        #link4::before { content: target-text(url(#id1), first-letter) }
      </style>
      <body>
        <a id="link1"></a>
        <div id="id1">1 Chapter 1</div>
        <a id="link2" data-count="#id2"></a>
        <div id="id2">2 Chapter 2</div>
        <div id="id3">3 Chapter 3</div>
        <a id="link3"></a>
        <div id="id4">4 Chapter 4</div>
        <a id="link4"></a>
    `)

	html := page.Box().Children[0]
	body := html.Box().Children[0]
	a1, _, a2, _, _, a3, _, a4 := unpack8(body)
	before := a1.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "test 4 Chapter 4", "before")
	before = a2.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "wow", "before")
	tu.AssertEqual(t, len(a3.Box().Children[0].Box().Children[0].Box().Children), 0, "len")
	before = a4.Box().Children[0].Box().Children[0].Box().Children[0]
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "1", "before")
}

func TestTargetFloat(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        a::after {
          content: target-counter("#h", page);
          float: right;
        }
      </style>
      <div><a id="span">link</a></div>
      <h1 id="h">abc</h1>
    `)

	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div, _ := unpack2(body)
	line := div.Box().Children[0]
	inline := line.Box().Children[0]
	textBox, after := unpack2(inline)
	tu.AssertEqual(t, textBox.(*bo.TextBox).Text, "link", "textBox")
	tu.AssertEqual(t, after.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "1", "after")
}

func TestTargetAbsolute(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        a::after {
          content: target-counter('#h', page);
        }
        div {
           position: absolute;
        }
      </style>
      <div><a id="span">link</a></div>
      <h1 id="h">abc</h1>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div, _ := unpack2(body)
	line := div.Box().Children[0]
	inline := line.Box().Children[0]
	textBox, after := unpack2(inline)
	assertText(t, textBox, "link")
	assertText(t, after.Box().Children[0], "1")
}
