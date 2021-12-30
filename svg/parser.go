package svg

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/benoitkugler/webrender/utils"
)

// provide low-level functions to read basic SVG data

type Fl = utils.Fl

var root2 = math.Sqrt(2)

// Unit is an enum type for units supported in SVG images.
type Unit uint8

// Units supported.
const (
	_ Unit = iota
	Px
	Cm
	Mm
	Pt
	In
	Q
	Pc

	// Special case : percentage (%) relative to the viewbox
	Perc
	// Special case : relative to the font size
	Em
	// Special case : relative to the font size
	Ex

	// special values for internal use
	auto
	autoStartReverse
)

var units = [...]string{Px: "px", Cm: "cm", Mm: "mm", Pt: "pt", In: "in", Q: "Q", Pc: "pc", Perc: "%", Em: "em", Ex: "ex"}

var toPx = [...]Fl{
	Px: 1, Cm: 96. / 2.54, Mm: 9.6 / 2.54, Pt: 96. / 72., In: 96., Q: 96. / 40. / 2.54, Pc: 96. / 6.,
	// other units depend on context
}

// 12pt
const defaultFontSize Fl = 96 * 12 / 72

// Value is a Value expressed in a unit.
// It may be relative, meaning that context is needed
// to obtain the actual Value (see the `resolve` method)
type Value struct {
	V Fl
	U Unit
}

// look for an absolute unit, or nothing (considered as pixels)
// % is also supported.
// it returns an empty value when 's' is empty
func parseValue(s string) (Value, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Value{}, nil
	}

	resolvedUnit := Px
	for u, suffix := range units {
		if u == 0 {
			continue
		}
		if strings.HasSuffix(s, suffix) {
			s = strings.TrimSpace(strings.TrimSuffix(s, suffix))
			resolvedUnit = Unit(u)
			break
		}
	}
	v, err := strconv.ParseFloat(s, 32)
	return Value{U: resolvedUnit, V: Fl(v)}, err
}

// Resolve convert `v` to pixels, resolving percentage and
// units relative to the font size.
func (v Value) Resolve(fontSize, percentageReference Fl) Fl {
	switch v.U {
	case Px, 0: // fast path for a common case
		return v.V
	case Perc:
		return v.V * percentageReference / 100
	case Em:
		return v.V * fontSize
	case Ex: // assume that 1em == 2ex
		return v.V * fontSize / 2
	default: // use the convertion table
		return v.V * toPx[v.U]
	}
}

// returns the index of the end of the first number starting at pos
// it is assumed that data[pos] is not a whitespace
// if isFlag is true only "0" or "1" are allowed,
// meaning for instance that 12 is parsed are 1 2, not 12.
func consumeNumber(data []byte, pos int, isFlag bool) int {
	if isFlag {
		return pos + 1
	}
	start := data[pos]
	seenDot := start == '.'
	pos++
	for ; pos < len(data); pos++ {
		c := data[pos]
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			continue
		case '.':
			if seenDot { // .5.5 is interpreted as 0.5 0.5
				return pos
			}
			// else continue: floating point
			seenDot = true
		case '-':
			// new number, expected on exponents
			if data[pos-1] == 'e' || data[pos-1] == 'E' {
				continue
			}
			return pos
		default:
			// accept numbers and exponents
			if c == 'e' || c == 'E' {
				continue
			}
			return pos
		}
	}
	return pos
}

// parsePoints reads a set of floating point values from the SVG format number string.
// units are not supported.
// to reduce allocations, values are appended to `points`, which is supposed to have 0 length,
// and is returned
// isEllipticalArc should be true for 'A' and 'a' commands, and is required
// to handle special case in number parsing
func parsePoints(dataPoints string, points []Fl, isEllipticalArc bool) ([]Fl, error) {
	data := []byte(dataPoints)

	for pos := 0; pos < len(data); {
		c := data[pos]
		if '0' <= c && c <= '9' || c == '.' || c == '-' || c == 'e' || c == 'E' {
			// for elliptical arc, arguments 4 and 5 are flags
			isFlag := isEllipticalArc && (len(points) == 3 || len(points) == 4)

			endNumber := consumeNumber(data, pos, isFlag)
			value, err := strconv.ParseFloat(dataPoints[pos:endNumber], 32)
			if err != nil {
				return nil, err
			}
			points = append(points, Fl(value))
			pos = endNumber
		} else {
			pos++ // skip "whitespaces"
		}
	}
	return points, nil
}

// parseValues reads a list of whitespace or comma-separated list of value,
// with units.
// the empty string or "none" are matched to a nil slice
func parseValues(dataPoints string) (points []Value, err error) {
	if dataPoints == "" || dataPoints == "none" {
		return nil, nil
	}

	fields := strings.FieldsFunc(dataPoints, func(r rune) bool { return r == ' ' || r == ',' })
	points = make([]Value, len(fields))
	for i, v := range fields {
		val, err := parseValue(v)
		if err != nil {
			return nil, err
		}
		points[i] = val
	}
	return points, nil
}

// parses opacity, stroke-opacity, fill-opacity attributes,
// returning 1 as a default value
func parseOpacity(op string) (Fl, error) {
	if op == "" {
		return 1, nil
	}
	out, err := strconv.ParseFloat(op, 32)
	return Fl(out), err
}

// if the URL is invalid, the empty string is returned
func parseURLFragment(url_ string) string {
	u, err := parseURL(url_)
	if err != nil {
		return ""
	}
	return u.Fragment
}

// parse a URL, possibly in a "url(…)" string.
func parseURL(url_ string) (*url.URL, error) {
	if strings.HasPrefix(url_, "url(") && strings.HasSuffix(url_, ")") {
		url_ = url_[4 : len(url_)-1]
		if len(url_) >= 2 {
			if (url_[0] == '"' && url_[len(url_)-1] == '"') || (url_[0] == '\'' && url_[len(url_)-1] == '\'') {
				url_ = url_[1 : len(url_)-1]
			}
		}
	}
	return url.Parse(url_)
}

func parseViewbox(attr string) (Rectangle, error) {
	points, err := parsePoints(attr, nil, false)
	if err != nil {
		return Rectangle{}, err
	}
	if len(points) != 4 {
		return Rectangle{}, fmt.Errorf("expected 4 numbers for viewbox, got %s", attr)
	}
	return Rectangle{points[0], points[1], points[2], points[3]}, nil
}

// return an empty list for empty attributes
func parseTransform(attr string) (out []transform, err error) {
	ts := strings.Split(attr, ")")
	for _, t := range ts {
		t = strings.TrimSpace(t)
		if len(t) == 0 {
			continue
		}

		d := strings.Split(t, "(")
		if len(d) != 2 || d[1] == "" {
			return nil, fmt.Errorf("invalid transformation: %s", t) // badly formed transformation
		}
		points, err := parseValues(d[1])
		if err != nil {
			return nil, fmt.Errorf("invalid transform: %s", err)
		}

		transformKind := strings.ToLower(strings.TrimSpace(d[0]))
		L := len(points)
		var tr transform
		copy(tr.args[:], points)

		switch transformKind {
		case "rotate":
			if L == 1 {
				tr.kind = rotate
			} else if L == 3 {
				tr.kind = rotateWithOrigin
			} else {
				return nil, fmt.Errorf("invalid transformation: %s", t)
			}
		case "translate":
			if L == 1 {
				tr.args[1] = Value{0, Px}
				tr.kind = translate
			} else if L == 2 {
				tr.kind = translate
			} else {
				return nil, fmt.Errorf("invalid transformation: %s", t)
			}
		case "skew":
			if L == 2 {
				tr.kind = skew
			} else {
				return nil, fmt.Errorf("invalid transformation: %s", t)
			}
		case "skewx":
			if L == 1 {
				tr.kind = skew
			} else {
				return nil, fmt.Errorf("invalid transformation: %s", t)
			}
		case "skewy":
			if L == 1 {
				tr.kind = skew
				tr.args[1] = tr.args[0]
				tr.args[0] = Value{0, Px}
			} else {
				return nil, fmt.Errorf("invalid transformation: %s", t)
			}
		case "scale":
			if L == 1 {
				tr.args[1] = Value{0, Px}
				tr.kind = scale
			} else if L == 2 {
				tr.kind = scale
			} else {
				return nil, fmt.Errorf("invalid transformation: %s", t)
			}
		case "matrix":
			if L == 6 {
				tr.kind = customMatrix
			} else {
				return nil, fmt.Errorf("invalid transformation: %s", t)
			}
		default:
			return nil, fmt.Errorf("invalid transformation: %s", t)
		}

		out = append(out, tr)

	}

	return out, nil
}

type preserveAspectRatio struct {
	xPosition, yPosition string
	none                 bool // align == "none"
	slice                bool // meet or slice
}

func parsePreserveAspectRatio(s string) (out preserveAspectRatio) {
	out.xPosition, out.yPosition = "min", "min"
	aspectRatio := strings.Split(s, " ")
	align := aspectRatio[0]
	if align != "none" || len(align) >= 5 {
		out.xPosition = strings.ToLower(align[1:4])
		out.yPosition = strings.ToLower(align[5:])
	}
	out.none = align == "none"
	out.slice = len(aspectRatio) >= 2 && aspectRatio[1] == "slice"
	return out
}

// accepts angle or "auto" or "auto-start-reverse"
// the angle is expressed in degrees
// the empty is matched to a 0 angle
func parseOrientation(attr string) (Value, error) {
	switch attr {
	case "":
		return Value{}, nil
	case "auto":
		return Value{U: auto}, nil
	case "auto-start-reverse":
		return Value{U: autoStartReverse}, nil
	default:
		f, err := strconv.ParseFloat(attr, 32)
		return Value{V: Fl(f)}, err
	}
}
