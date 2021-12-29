package properties

import (
	"log"
	"math"

	"github.com/benoitkugler/webrender/css/parser"
)

var Inf = Float(math.Inf(+1))

// --------------- Values  -----------------------------------------------

func (d Dimension) ToPixels() Dimension {
	if c, ok := LengthsToPixels[d.Unit]; ok {
		return Dimension{Unit: Px, Value: d.Value * c}
	}
	return d
}

func (p Point) ToPixels() Point {
	return Point{p[0].ToPixels(), p[1].ToPixels()}
}

// ToValue wraps `d` to a Value object.
func (d Dimension) ToValue() Value {
	return Value{Dimension: d}
}

func FToD(f Fl) Dimension      { return Dimension{Value: Float(f), Unit: Scalar} }
func SToV(s string) Value      { return Value{String: s} }
func FToV(f Fl) Value          { return FToD(f).ToValue() }
func (f Float) ToValue() Value { return FToV(Fl(f)) }

func (v Value) ToMaybeFloat() MaybeFloat {
	if v.String == "auto" {
		return Auto
	}
	return v.Value
}

// FToPx returns `f` as pixels.
func FToPx(f Float) Value { return Dimension{Unit: Px, Value: f}.ToValue() }

func NewColor(r, g, b, a Fl) Color {
	return Color{RGBA: parser.RGBA{R: r, G: g, B: b, A: a}, Type: parser.ColorRGBA}
}

// Has returns `true` is v is one of the three elements.
func (d Display) Has(v string) bool {
	return d[0] == v || d[1] == v || d[2] == v
}

const (
	True  = Bool(true)
	False = Bool(false)
)

type Bool bool

func (Bool) isMaybeBool() {}

type MaybeBool interface {
	isMaybeBool()
}

// ----------------- misc  ---------------------------------------

func (p Point) ToSlice() []Dimension {
	return []Dimension{p[0], p[1]}
}

func (s GradientSize) IsExplicit() bool {
	return s.Keyword == ""
}

// -------------- Images ------------------------

func (i NoneImage) isCssProperty()      {}
func (i UrlImage) isCssProperty()       {}
func (i LinearGradient) isCssProperty() {}
func (i RadialGradient) isCssProperty() {}
func (i NoneImage) isImage()            {}
func (i UrlImage) isImage()             {}
func (i LinearGradient) isImage()       {}
func (i RadialGradient) isImage()       {}

// -------------------------- Content Property --------------------------
// func (i NoneImage) copyAsInnerContent() InnerContent      { return i }
// func (i UrlImage) copyAsInnerContent() InnerContent       { return i }
// func (i LinearGradient) copyAsInnerContent() InnerContent { return i.copy() }
// func (i RadialGradient) copyAsInnerContent() InnerContent { return i.copy() }

// contents,
func (s String) isInnerContent()  {}
func (s Strings) isInnerContent() {}

func (s Counters) isInnerContent() {}

// target
func (s SContentProps) isInnerContent() {}

// url
func (s NamedString) isInnerContent() {}

func (s Dimension) isInnerContent() {}
func (s Float) isInnerContent()     {}
func (s Int) isInnerContent()       {}
func (s Color) isInnerContent()     {}
func (s Quote) isInnerContent()     {}
func (s AttrData) isInnerContent()  {}
func (s VarData) isInnerContent()   {}

func (c ContentProperty) AsString() (value string) {
	return string(c.Content.(String))
}

func (c ContentProperty) AsLeader() string {
	return c.Content.(Strings)[1]
}

func (c ContentProperty) AsCounter() (counterName string, counterStyle CounterStyleID) {
	value, _ := c.Content.(Counters)
	return value.Name, value.Style
}

func (c ContentProperty) AsCounters() (counterName, separator string, counterStyle CounterStyleID) {
	value, _ := c.Content.(Counters)
	return value.Name, value.Separator, value.Style
}

func (c ContentProperty) AsStrings() []string {
	value, ok := c.Content.(Strings)
	if !ok {
		log.Fatalf("invalid content (expected []string): %v", c.Content)
	}
	return value
}

func (c ContentProperty) AsTargetCounter() (anchorToken ContentProperty, counterName, counterStyle string) {
	value, _ := c.Content.(SContentProps)
	if len(value) != 3 {
		log.Fatalf("invalid content (expected 3-list of String or ContentProperty): %v", c.Content)
	}
	return value[0].ContentProperty, value[1].String, value[2].String
}

func (c ContentProperty) AsTargetCounters() (anchorToken ContentProperty, counterName string, separator ContentProperty, counterStyle string) {
	value, _ := c.Content.(SContentProps)
	if len(value) != 4 {
		log.Fatalf("invalid content (expected 4-list of String or ContentProperty): %v", c.Content)
	}
	return value[0].ContentProperty, value[1].String, value[2].ContentProperty, value[3].String
}

func (c ContentProperty) AsTargetText() (anchorToken ContentProperty, textStyle string) {
	value, _ := c.Content.(SContentProps)
	if len(value) != 2 {
		log.Fatalf("invalid content (expected 2-list of String or ContentProperty): %v", c.Content)
	}
	return value[0].ContentProperty, value[1].String
}

func (c ContentProperty) AsQuote() Quote {
	return c.Content.(Quote)
}

// ------------------------- Usefull for test ---------------------------

func (bs Images) Repeat(n int) CssProperty {
	var out Images
	for i := 0; i < n; i++ {
		out = append(out, bs...)
	}
	return out
}

func (bs Centers) Repeat(n int) CssProperty {
	var out Centers
	for i := 0; i < n; i++ {
		out = append(out, bs...)
	}
	return out
}

func (bs Sizes) Repeat(n int) CssProperty {
	var out Sizes
	for i := 0; i < n; i++ {
		out = append(out, bs...)
	}
	return out
}

func (bs Repeats) Repeat(n int) CssProperty {
	var out Repeats
	for i := 0; i < n; i++ {
		out = append(out, bs...)
	}
	return out
}

func (bs Strings) Repeat(n int) CssProperty {
	var out Strings
	for i := 0; i < n; i++ {
		out = append(out, bs...)
	}
	return out
}

type (
	SContents  []SContent
	Dimensions []Dimension
)

// --------------- Geometry -------------------------

type Rectangle [4]Float

func (r Rectangle) ToFloat() [4]Fl {
	return [4]Fl{Fl(r[0]), Fl(r[1]), Fl(r[2]), Fl(r[3])}
}

func (r Rectangle) Unpack() (x, y, w, h Fl) {
	o := r.ToFloat()
	return o[0], o[1], o[2], o[3]
}

func (r Rectangle) Unpack2() (x, y, w, h Float) {
	return r[0], r[1], r[2], r[3]
}

func (r Rectangle) IsNone() bool {
	return r == Rectangle{}
}
