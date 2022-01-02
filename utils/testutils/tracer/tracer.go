// Package tracer provides a function to dump the current layout tree,
// which may be used in debug mode.
package tracer

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/utils"
)

type Tracer struct {
	out *os.File
}

// NewTracer panics if an error occurs.
func NewTracer(outFile string) Tracer {
	f, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}

	return Tracer{out: f}
}

func FormatMaybeFloat(v properties.MaybeFloat) string {
	if v, ok := v.(properties.Float); ok {
		return strconv.FormatFloat(float64(utils.RoundPrec(float32(v), -1)), 'g', -1, 32)
	}
	return fmt.Sprintf("%v", v)
}

func (t Tracer) Dump(line string) {
	fmt.Fprintln(t.out, line)
}

func (t Tracer) DumpTree(box boxes.Box, context string) {
	fmt.Fprintln(t.out, context)

	var printer func(box boxes.Box, indent int)
	printer = func(box boxes.Box, indent int) {
		fmt.Fprint(t.out, strings.Repeat(" ", indent))
		fmt.Fprintf(t.out, "%s: %s %s %s %s\n", box.Type(),
			FormatMaybeFloat(box.Box().PositionX),
			FormatMaybeFloat(box.Box().PositionY),
			FormatMaybeFloat(box.Box().Width),
			FormatMaybeFloat(box.Box().Height),
		)
		if box, ok := box.(*boxes.TextBox); ok {
			fmt.Fprintln(t.out, box.Text)
		}

		for _, child := range box.Box().Children {
			printer(child, indent+1)
		}
	}

	printer(box, 0)

	fmt.Fprintln(t.out)
}
