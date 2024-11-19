package layout

import (
	"testing"

	bo "github.com/benoitkugler/webrender/html/boxes"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Test the CSS cross references using target-*() functions.

func TestTargetCounter(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	html := unpack1(page)
	body := unpack1(html)
	div1, div2, div3, div4 := unpack4(body)
	before := unpack1(div1.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "4")
	before = unpack1(div2.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "test 1")
	before = unpack1(div3.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "iv")
	before = unpack1(div4.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "3")
}

func TestTargetCounterAttr(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	html := unpack1(page)
	body := unpack1(html)
	div1, div2, div3, div4 := unpack4(body)
	before := unpack1(div1.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "4")
	before = unpack1(div2.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "1")
	before = unpack1(div3.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "2")
	before = unpack1(div4.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "c")
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

	html := unpack1(page)
	body := unpack1(html)
	div1, div2, div3, div4 := unpack4(body)
	before := unpack1(div1.Box().Children[1].Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "4.2")
	before = unpack1(div2.Box().Children[0].Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "3")
	before = unpack1(div3.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "b.a")
	before = unpack1(div4.Box().Children[1].Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "12")
}

func TestTargetText(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	html := unpack1(page)
	body := unpack1(html)
	a1, _, a2, _, _, a3, _, a4 := unpack8(body)
	before := unpack1(a1.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "test 4 Chapter 4")
	before = unpack1(a2.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "wow")
	tu.AssertEqual(t, len(unpack1(unpack1(a3)).Box().Children), 0)
	before = unpack1(a4.Box().Children[0].Box().Children[0])
	tu.AssertEqual(t, before.(*bo.TextBox).Text, "1")
}

func TestTargetFloat(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	html := unpack1(page)
	body := unpack1(html)
	div, _ := unpack2(body)
	line := unpack1(div)
	inline := unpack1(line)
	textBox, after := unpack2(inline)
	tu.AssertEqual(t, textBox.(*bo.TextBox).Text, "link")
	tu.AssertEqual(t, unpack1(after.Box().Children[0]).(*bo.TextBox).Text, "1")
}

func TestTargetAbsolute(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	div, _ := unpack2(body)
	line := unpack1(div)
	inline := unpack1(line)
	textBox, after := unpack2(inline)
	assertText(t, textBox, "link")
	assertText(t, unpack1(after), "1")
}
