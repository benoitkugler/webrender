package images

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils"
)

var (
	// re1 = regexp.MustCompile("(?<!e)-")
	re2 = regexp.MustCompile("[ \n\r\t,]+")
	// re3 = regexp.MustCompile(`(\.[0-9-]+)(?=\.)`)

	UNITS = map[string]pr.Float{
		"mm": 1 / 25.4,
		"cm": 1 / 2.54,
		"in": 1,
		"pt": 1 / 72.,
		"pc": 1 / 6.,
		"px": 1,
	}
)

// Normalize a string corresponding to an array of various values.
func normalize(str string) string {
	str = strings.ReplaceAll(str, "E", "e")
	// str = re1.ReplaceAllString(str, " -") // TODO:
	str = re2.ReplaceAllString(str, " ")
	// str = re3.ReplaceAllString(str, `\1 `)  // TODO:
	return strings.TrimSpace(str)
}

type floatOrString struct {
	s string
	f float64
}

func toFloat(s string, offset int) (pr.Float, error) {
	rs := []rune(s)
	s = string(rs[:len(rs)-offset])
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("wrong string for float : %s", s)
	}
	return pr.Float(f), nil
}

// Return ``viewbox`` of ``node``.
func getViewBox(node *utils.HTMLNode) []pr.Float {
	viewbox := node.Get("viewBox")
	if viewbox == "" {
		return nil
	}

	var out []pr.Float
	for _, position := range strings.Split(normalize(viewbox), " ") {
		f, err := strconv.ParseFloat(position, 64)
		if err != nil {
			log.Printf("wrong string for float %s in viewbox", position)
			return nil
		}
		out = append(out, pr.Float(f))
	}
	return out
}
