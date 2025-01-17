package keywords

import (
	"fmt"
	"sort"
	"testing"

	"github.com/benoitkugler/webrender/utils"
)

func TestKeywordSorted(t *testing.T) {
	l := [...]string{
		"center", "space-between", "space-around", "space-evenly",
		"stretch", "normal", "flex-start", "flex-end",
		"start", "end", "left", "right",
		"safe", "unsafe",
		"center", "start", "end", "flex-start", "flex-end", "left",
		"right",
		"normal", "stretch", "center", "start", "end", "self-start",
		"self-end", "flex-start", "flex-end", "left", "right",
		"legacy",
		"baseline",
		"center", "start", "end", "self-start", "self-end",
		"flex-start", "flex-end", "left", "right",
		"auto", "normal", "stretch", "center", "start", "end",
		"self-start", "self-end", "flex-start", "flex-end", "left",
		"right",
		"center", "start", "end", "self-start", "self-end",
		"flex-start", "flex-end", "left", "right",
		"normal", "stretch", "center", "start", "end", "self-start",
		"self-end", "flex-start", "flex-end",
		"baseline",
		"center", "start", "end", "self-start", "self-end",
		"flex-start", "flex-end",
		"auto", "normal", "stretch", "center", "start", "end",
		"self-start", "self-end", "flex-start", "flex-end",
		"center", "start", "end", "self-start", "self-end",
		"flex-start", "flex-end",
		"center", "space-between", "space-around", "space-evenly",
		"stretch", "normal", "flex-start", "flex-end",
		"start", "end",
		"baseline",
		"center", "start", "end", "flex-start", "flex-end",
		"first", "last",
	}

	m := utils.NewSet(l[:]...)
	var out []string
	for l := range m {
		out = append(out, l)
	}
	sort.Strings(out)
	for _, s := range out {
		fmt.Printf("case %q:\nreturn %s\n", s, s)
	}
}
