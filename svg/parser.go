package svg

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"unicode"

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

// // convert the unite to pixels. Return true if it is a %
// func parseUnit(s string) (Fl, bool, error) {
// 	value, err := parseValue(s)
// 	return value.v * toPx[value.u], value.u == Perc, err
// }

// type percentageReference uint8

// const (
// 	widthPercentage percentageReference = iota
// 	heightPercentage
// 	diagPercentage
// )

// // resolveUnit converts a length with a unit into its value in 'px'
// // percentage are supported, and refer to the viewBox
// // `asPerc` is only applied when `s` contains a percentage.
// func (viewBox Bounds) resolveUnit(s string, asPerc percentageReference) (Fl, error) {
// 	value, isPercentage, err := parseUnit(s)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if isPercentage {
// 		w, h := viewBox.W, viewBox.H
// 		switch asPerc {
// 		case widthPercentage:
// 			return value / 100 * w, nil
// 		case heightPercentage:
// 			return value / 100 * h, nil
// 		case diagPercentage:
// 			normalizedDiag := math.Sqrt(w*w+h*h) / root2
// 			return value / 100 * normalizedDiag, nil
// 		}
// 	}
// 	return value, nil
// }

// // parseUnit converts a length with a unit into its value in 'px'
// // percentage are supported, and refer to the current ViewBox
// func (c *iconCursor) parseUnit(s string, asPerc percentageReference) (Fl, error) {
// 	return c.icon.ViewBox.resolveUnit(s, asPerc)
// }

// func readFraction(v string) (f Fl, err error) {
// 	v = strings.TrimSpace(v)
// 	d := 1.0
// 	if strings.HasSuffix(v, "%") {
// 		d = 100
// 		v = strings.TrimSuffix(v, "%")
// 	}
// 	f, err = parseBasicFloat(v)
// 	f /= d
// 	return
// }

// func readAppendFloat(numStr string, points []Fl) ([]Fl, error) {
// 	fmt.Println(numStr)
// 	last := 0
// 	isFirst := true
// 	for i, n := range numStr {
// 		if n == '.' {
// 			if isFirst {
// 				isFirst = false
// 				continue
// 			}
// 			f, err := parseBasicFloat(numStr[last:i])
// 			if err != nil {
// 				return nil, err
// 			}
// 			points = append(points, f)
// 			last = i
// 		}
// 	}
// 	f, err := parseBasicFloat(numStr[last:])
// 	if err != nil {
// 		return nil, err
// 	}
// 	points = append(points, f)
// 	return points, nil
// }

// parsePoints reads a set of floating point values from the SVG format number string.
// units are not supported.
// values are appended to points, which is returned
func parsePoints(dataPoints string, points []Fl) ([]Fl, error) {
	lastIndex := -1
	lr := ' '
	for i, r := range dataPoints {
		if !unicode.IsNumber(r) && r != '.' && !(r == '-' && lr == 'e') && r != 'e' {
			if lastIndex != -1 {
				value, err := strconv.ParseFloat(dataPoints[lastIndex:i], 32)
				if err != nil {
					return nil, err
				}
				points = append(points, Fl(value))
			}
			if r == '-' {
				lastIndex = i
			} else {
				lastIndex = -1
			}
		} else if lastIndex == -1 {
			lastIndex = i
		}
		lr = r
	}
	if lastIndex != -1 && lastIndex != len(dataPoints) {
		value, err := strconv.ParseFloat(dataPoints[lastIndex:], 32)
		if err != nil {
			return nil, err
		}
		points = append(points, Fl(value))
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
	points, err := parsePoints(attr, nil)
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
			return nil, errParamMismatch // badly formed transformation
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
				return nil, errParamMismatch
			}
		case "translate":
			if L == 1 {
				tr.args[1] = Value{0, Px}
				tr.kind = translate
			} else if L == 2 {
				tr.kind = translate
			} else {
				return nil, errParamMismatch
			}
		case "skew":
			if L == 2 {
				tr.kind = skew
			} else {
				return nil, errParamMismatch
			}
		case "skewx":
			if L == 1 {
				tr.kind = skew
			} else {
				return nil, errParamMismatch
			}
		case "skewy":
			if L == 1 {
				tr.kind = skew
				tr.args[1] = tr.args[0]
				tr.args[0] = Value{0, Px}
			} else {
				return nil, errParamMismatch
			}
		case "scale":
			if L == 1 {
				tr.args[1] = Value{0, Px}
				tr.kind = scale
			} else if L == 2 {
				tr.kind = scale
			} else {
				return nil, errParamMismatch
			}
		case "matrix":
			if L == 6 {
				tr.kind = customMatrix
			} else {
				return nil, errParamMismatch
			}
		default:
			return nil, errParamMismatch
		}

		out = append(out, tr)

	}

	return out, nil
}
