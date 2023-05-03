package boxes

import (
	"reflect"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/html/tree"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func TestCounters1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	exp := func(counter string) SerBox {
		return SerBox{"p", BlockT, BC{C: []SerBox{
			{"p", LineT, BC{C: []SerBox{{"p::before", InlineT, BC{C: []SerBox{{"p::before", TextT, BC{Text: counter}}}}}}}},
		}}}
	}
	var expected []SerBox
	for _, counter := range strings.Fields("0 1 3  2 4 6  -11 -9 -7  44 46 48") {
		expected = append(expected, exp(counter))
	}
	assertTree(t, parseAndBuild(t, `
      <style>
        p { counter-increment: p 2 }
        p:before { content: counter(p); }
        p:nth-child(1) { counter-increment: none; }
        p:nth-child(2) { counter-increment: p; }
      </style>
      <p></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p 117 p"></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p -13"></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p 42"></p>
      <p></p>
      <p></p>`), expected)
}

func TestCounters2(t *testing.T) {
	// cp := tu.CaptureLogs()
	// defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <ol style="list-style-position: inside">
        <li></li>
        <li></li>
        <li></li>
        <li><ol>
          <li></li>
          <li style="counter-increment: none"></li>
          <li></li>
        </ol></li>
        <li></li>
      </ol>`), []SerBox{
		{"ol", BlockT, BC{C: []SerBox{
			{"li", BlockT, BC{C: []SerBox{
				{"li", LineT, BC{C: []SerBox{{"li::marker", InlineT, BC{C: []SerBox{{"li::marker", TextT, BC{Text: "1. "}}}}}}}},
			}}},
			{"li", BlockT, BC{C: []SerBox{
				{"li", LineT, BC{C: []SerBox{{"li::marker", InlineT, BC{C: []SerBox{{"li::marker", TextT, BC{Text: "2. "}}}}}}}},
			}}},
			{"li", BlockT, BC{C: []SerBox{
				{"li", LineT, BC{C: []SerBox{{"li::marker", InlineT, BC{C: []SerBox{{"li::marker", TextT, BC{Text: "3. "}}}}}}}},
			}}},
			{"li", BlockT, BC{C: []SerBox{
				{"li", BlockT, BC{C: []SerBox{
					{"li", LineT, BC{C: []SerBox{{"li::marker", InlineT, BC{C: []SerBox{{"li::marker", TextT, BC{Text: "4. "}}}}}}}},
				}}},
				{"ol", BlockT, BC{C: []SerBox{
					{"li", BlockT, BC{C: []SerBox{
						{"li", LineT, BC{C: []SerBox{{"li::marker", InlineT, BC{C: []SerBox{{"li::marker", TextT, BC{Text: "1. "}}}}}}}},
					}}},
					{"li", BlockT, BC{C: []SerBox{
						{"li", LineT, BC{C: []SerBox{{"li::marker", InlineT, BC{C: []SerBox{{"li::marker", TextT, BC{Text: "1. "}}}}}}}},
					}}},
					{"li", BlockT, BC{C: []SerBox{
						{"li", LineT, BC{C: []SerBox{{"li::marker", InlineT, BC{C: []SerBox{{"li::marker", TextT, BC{Text: "2. "}}}}}}}},
					}}},
				}}},
			}}},
			{"li", BlockT, BC{C: []SerBox{
				{"li", LineT, BC{C: []SerBox{{"li::marker", InlineT, BC{C: []SerBox{{"li::marker", TextT, BC{Text: "5. "}}}}}}}},
			}}},
		}}},
	})
}

func TestCounters3(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        p { display: list-item; list-style: inside decimal }
      </style>
      <div>
        <p></p>
        <p></p>
        <p style="counter-reset: list-item 7 list-item -56"></p>
      </div>
      <p></p>`), []SerBox{
		{"div", BlockT, BC{C: []SerBox{
			{"p", BlockT, BC{C: []SerBox{
				{"p", LineT, BC{C: []SerBox{{"p::marker", InlineT, BC{C: []SerBox{{"p::marker", TextT, BC{Text: "1. "}}}}}}}},
			}}},
			{"p", BlockT, BC{C: []SerBox{
				{"p", LineT, BC{C: []SerBox{{"p::marker", InlineT, BC{C: []SerBox{{"p::marker", TextT, BC{Text: "2. "}}}}}}}},
			}}},
			{"p", BlockT, BC{C: []SerBox{
				{"p", LineT, BC{C: []SerBox{{"p::marker", InlineT, BC{C: []SerBox{{"p::marker", TextT, BC{Text: "-55. "}}}}}}}},
			}}},
		}}},
		{"p", BlockT, BC{C: []SerBox{
			{"p", LineT, BC{C: []SerBox{{"p::marker", InlineT, BC{C: []SerBox{{"p::marker", TextT, BC{Text: "1. "}}}}}}}},
		}}},
	})
}

func TestCounters4(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        section:before { counter-reset: h; content: "" }
        h1:before { counter-increment: h; content: counters(h, ".") }
      </style>
      <body>
        <section><h1></h1>
          <h1></h1>
          <section><h1></h1>
            <h1></h1>
          </section>
          <h1></h1>
        </section>
      </body>`), []SerBox{
		{"section", BlockT, BC{C: []SerBox{
			{"section", BlockT, BC{C: []SerBox{{"section", LineT, BC{C: []SerBox{{"section::before", InlineT, BC{C: []SerBox{}}}}}}}}},
			{"h1", BlockT, BC{C: []SerBox{
				{"h1", LineT, BC{C: []SerBox{{"h1::before", InlineT, BC{C: []SerBox{{"h1::before", TextT, BC{Text: "1"}}}}}}}},
			}}},
			{"h1", BlockT, BC{C: []SerBox{
				{"h1", LineT, BC{C: []SerBox{{"h1::before", InlineT, BC{C: []SerBox{{"h1::before", TextT, BC{Text: "2"}}}}}}}},
			}}},
			{"section", BlockT, BC{C: []SerBox{
				{"section", BlockT, BC{C: []SerBox{{"section", LineT, BC{C: []SerBox{{"section::before", InlineT, BC{C: []SerBox{}}}}}}}}},
				{"h1", BlockT, BC{C: []SerBox{
					{"h1", LineT, BC{C: []SerBox{{"h1::before", InlineT, BC{C: []SerBox{{"h1::before", TextT, BC{Text: "2.1"}}}}}}}},
				}}},
				{"h1", BlockT, BC{C: []SerBox{
					{"h1", LineT, BC{C: []SerBox{{"h1::before", InlineT, BC{C: []SerBox{{"h1::before", TextT, BC{Text: "2.2"}}}}}}}},
				}}},
			}}},
			{"h1", BlockT, BC{C: []SerBox{
				{"h1", LineT, BC{C: []SerBox{{"h1::before", InlineT, BC{C: []SerBox{{"h1::before", TextT, BC{Text: "3"}}}}}}}},
			}}},
		}}},
	})
}

func TestCounters5(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        p:before { content: counter(c) }
      </style>
      <div>
        <span style="counter-reset: c">
          Scope created now, deleted after the div
        </span>
      </div>
      <p></p>`), []SerBox{
		{"div", BlockT, BC{C: []SerBox{
			{"div", LineT, BC{C: []SerBox{{"span", InlineT, BC{C: []SerBox{{"span", TextT, BC{Text: "Scope created now, deleted after the div "}}}}}}}},
		}}},
		{"p", BlockT, BC{C: []SerBox{
			{"p", LineT, BC{C: []SerBox{{"p::before", InlineT, BC{C: []SerBox{{"p::before", TextT, BC{Text: "0"}}}}}}}},
		}}},
	})
}

func TestCounters6(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// counter-increment may interfere with display: list-item
	assertTree(t, parseAndBuild(t, `
      <p style="counter-increment: c;
                display: list-item; list-style: inside decimal">`), []SerBox{
		{"p", BlockT, BC{C: []SerBox{
			{"p", LineT, BC{C: []SerBox{{"p::marker", InlineT, BC{C: []SerBox{{"p::marker", TextT, BC{Text: "0. "}}}}}}}},
		}}},
	})
}

func TestCounters7(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	exp := func(counter string) SerBox {
		return SerBox{"p", BlockT, BC{C: []SerBox{
			{"p", LineT, BC{C: []SerBox{{"p::before", InlineT, BC{C: []SerBox{{"p::before", TextT, BC{Text: counter}}}}}}}},
		}}}
	}
	var expected []SerBox
	for _, counter := range strings.Fields("2.0 2.3 4.3") {
		expected = append(expected, exp(counter))
	}
	// Test that counters are case-sensitive
	// See https://github.com/Kozea/WeasyPrint/pull/827
	assertTree(t, parseAndBuild(t, `
      <style>
        p { counter-increment: p 2 }
        p:before { content: counter(p) "." counter(P); }
      </style>
      <p></p>
      <p style="counter-increment: P 3"></p>
      <p></p>`), expected)
}

func TestCounters8(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        p:before { content: 'a'; display: list-item }
      </style>
      <p></p>
      <p></p>`), []SerBox{
		{"p", BlockT, BC{C: []SerBox{
			{"p::before", BlockT, BC{C: []SerBox{
				{"p::marker", BlockT, BC{C: []SerBox{{"p::marker", LineT, BC{C: []SerBox{{"p::marker", TextT, BC{Text: "• "}}}}}}}},
				{"p::before", BlockT, BC{C: []SerBox{{"p::before", LineT, BC{C: []SerBox{{"p::before", TextT, BC{Text: "a"}}}}}}}},
			}}},
		}}},
		{"p", BlockT, BC{C: []SerBox{
			{"p::before", BlockT, BC{C: []SerBox{
				{"p::marker", BlockT, BC{C: []SerBox{{"p::marker", LineT, BC{C: []SerBox{{"p::marker", TextT, BC{Text: "• "}}}}}}}},
				{"p::before", BlockT, BC{C: []SerBox{{"p::before", LineT, BC{C: []SerBox{{"p::before", TextT, BC{Text: "a"}}}}}}}},
			}}},
		}}},
	})
}

func TestCounterStyles1(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	exp := func(counter string) SerBox {
		return SerBox{"p", BlockT, BC{C: []SerBox{
			{"p", LineT, BC{C: []SerBox{{"p::before", InlineT, BC{C: []SerBox{{"p::before", TextT, BC{Text: counter}}}}}}}},
		}}}
	}
	var expected []SerBox
	for _, counter := range strings.Fields("--  •  ◦  ▪  -7 Counter:-6 -5:Counter") {
		expected = append(expected, exp(counter))
	}
	assertTree(t, parseAndBuild(t, `
      <style>
        body { --var: 'Counter'; counter-reset: p -12 }
        p { counter-increment: p }
        p:nth-child(1):before { content: '-' counter(p, none) '-'; }
        p:nth-child(2):before { content: counter(p, disc); }
        p:nth-child(3):before { content: counter(p, circle); }
        p:nth-child(4):before { content: counter(p, square); }
        p:nth-child(5):before { content: counter(p); }
        p:nth-child(6):before { content: var(--var) ':' counter(p); }
        p:nth-child(7):before { content: counter(p) ':' var(--var); }
      </style>
      <p></p>
      <p></p>
      <p></p>
      <p></p>
      <p></p>
      <p></p>
      <p></p>
    `), expected)
}

func TestCounterStyles2(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	exp := func(counter string) SerBox {
		return SerBox{"p", BlockT, BC{C: []SerBox{
			{"p", LineT, BC{C: []SerBox{{"p::before", InlineT, BC{C: []SerBox{{"p::before", TextT, BC{Text: counter}}}}}}}},
		}}}
	}
	var expected []SerBox
	for _, counter := range strings.Fields("-1986 -1985  -11 -10 -9 -8  -1 00 01 02  09 10 11 99 100 101  4135 4136") {
		expected = append(expected, exp(counter))
	}

	assertTree(t, parseAndBuild(t, `
      <style>
        p { counter-increment: p }
        p::before { content: counter(p, decimal-leading-zero); }
      </style>
      <p style="counter-reset: p -1987"></p>
      <p></p>
      <p style="counter-reset: p -12"></p>
      <p></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p -2"></p>
      <p></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p 8"></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p 98"></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p 4134"></p>
      <p></p>
    `), expected)
}

func testCounterStyle(t *testing.T, style string, inputs []int, expected string) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	render := tree.UACounterStyle.RenderValue
	var results []string
	for _, value := range inputs {
		results = append(results, render(value, style))
	}
	if !reflect.DeepEqual(results, strings.Fields(expected)) {
		t.Fatalf("unexpected counters for style %s: %v", style, results)
	}
}

func TestCounterStyles(t *testing.T) {
	testCounterStyle(t, "decimal-leading-zero", []int{
		-1986, -1985,
		-11, -10, -9, -8,
		-1, 0, 1, 2,
		9, 10, 11,
		99, 100, 101,
		4135, 4136,
	}, `
        -1986 -1985  -11 -10 -9 -8  -1 00 01 02  09 10 11
        99 100 101  4135 4136
    `)

	testCounterStyle(t, "lower-roman", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		49, 50,
		389, 390,
		3489, 3490, 3491,
		4999, 5000, 5001,
	}, `
		-1986 -1985  -1 0 i ii iii iv v vi vii viii ix x xi xii
		xlix l  ccclxxxix cccxc  mmmcdlxxxix mmmcdxc mmmcdxci
		4999 5000 5001
    `)
	testCounterStyle(t, "upper-roman", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		49, 50,
		389, 390,
		3489, 3490, 3491,
		4999, 5000, 5001,
	}, `
	        -1986 -1985  -1 0 I II III IV V VI VII VIII IX X XI XII
	        XLIX L  CCCLXXXIX CCCXC  MMMCDLXXXIX MMMCDXC MMMCDXCI
	        4999 5000 5001
    `)

	testCounterStyle(t, "lower-alpha", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4,
		25, 26, 27, 28, 29,
		2002, 2003,
	}, `
		-1986 -1985  -1 0 a b c d  y z aa ab ac bxz bya
    `)

	testCounterStyle(t, "upper-alpha", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4,
		25, 26, 27, 28, 29,
		2002, 2003,
	}, `
		-1986 -1985  -1 0 A B C D  Y Z AA AB AC BXZ BYA
    `)

	testCounterStyle(t, "lower-latin", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4,
		25, 26, 27, 28, 29,
		2002, 2003,
	}, `
		-1986 -1985  -1 0 a b c d  y z aa ab ac bxz bya
    `)

	testCounterStyle(t, "lower-latin", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4,
		25, 26, 27, 28, 29,
		2002, 2003,
	}, `
		-1986 -1985  -1 0 a b c d  y z aa ab ac bxz bya
    `)

	testCounterStyle(t, "upper-latin", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4,
		25, 26, 27, 28, 29,
		2002, 2003,
	}, `
        -1986 -1985  -1 0 A B C D  Y Z AA AB AC BXZ BYA
    `)

	testCounterStyle(t, "georgian", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		20, 30, 40, 50, 60, 70, 80, 90, 100,
		200, 300, 400, 500, 600, 700, 800, 900, 1000,
		2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000,
		19999, 20000, 20001,
	}, `
        -1986 -1985  -1 0 ა
        ბ გ დ ე ვ ზ ჱ თ ი ია იბ
        კ ლ მ ნ ჲ ო პ ჟ რ
        ს ტ ჳ ფ ქ ღ ყ შ ჩ
        ც ძ წ ჭ ხ ჴ ჯ ჰ ჵ
        ჵჰშჟთ 20000 20001
    `)

	testCounterStyle(t, "armenian", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		20, 30, 40, 50, 60, 70, 80, 90, 100,
		200, 300, 400, 500, 600, 700, 800, 900, 1000,
		2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000,
		9999, 10000, 10001,
	}, `
        -1986 -1985  -1 0 Ա
        Բ Գ Դ Ե Զ Է Ը Թ Ժ ԺԱ ԺԲ
        Ի Լ Խ Ծ Կ Հ Ձ Ղ Ճ
        Մ Յ Ն Շ Ո Չ Պ Ջ Ռ
        Ս Վ Տ Ր Ց Ւ Փ Ք
        ՔՋՂԹ 10000 10001
    `)
}
