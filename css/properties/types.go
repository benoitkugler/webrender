package properties

import (
	"fmt"

	pa "github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/utils"
)

// ------------- Top levels types, implementing CssProperty ------------

type StringSet struct {
	String   string
	Contents SContents
}

type Images []Image

type Centers []Center

type Sizes []Size

type Repeats [][2]string

type Strings []string

// Intersects returns true if at least one value in [values]
// is also in the list.
func (ss Strings) Intersects(values ...string) bool {
	for _, v1 := range ss {
		for _, v2 := range values {
			if v1 == v2 {
				return true
			}
		}
	}
	return false
}

type SContent struct {
	String   string
	Contents ContentProperties
}

type Display [3]string

type Decorations utils.Set

// Union return the union of  [s] and [other]
func (dec Decorations) Union(other Decorations) Decorations {
	out := utils.Set(dec).Copy()
	for k := range other {
		out[k] = utils.Has
	}
	return Decorations(out)
}

type Transforms []SDimensions

type Values []DimOrS

type SIntStrings struct {
	String string
	Values IntStrings
}

type SStrings struct {
	String  string
	Strings Strings
}

type CounterStyleID struct {
	Type    string // one of symbols(), string, or empty for an identifier
	Name    string
	Symbols Strings
}

type SDimensions struct {
	String     string
	Dimensions Dimensions
}

type IntStrings []IntString

type (
	DimOrS4 [4]DimOrS
	DimOrS5 [5]DimOrS
)

type Quotes struct {
	Open  Strings
	Close Strings
	Tag   Tag
}

type ContentProperties []ContentProperty

type AttrData struct {
	Fallback   CssProperty
	Name       string
	TypeOrUnit string
}

type Float Fl

type Int int

type Ints3 [3]int

type Page string

// Dimension or "auto" or "cover" or "contain"
type Size struct {
	String string
	Width  DimOrS
	Height DimOrS
}

type Center struct {
	OriginX string
	OriginY string
	Pos     Point
}

type Color pa.Color

type ContentProperty struct {
	// SStrings for type STRING, attr or string, counter, counters
	// Quote for type QUOTE
	// Url for URI
	// String for leader()
	Content InnerContent

	Type string
}

type NamedString struct {
	Name   string
	String string
}

type TaggedString struct {
	S   string
	Tag Tag
}

type Point [2]Dimension

type Marks struct {
	Crop  bool
	Cross bool
}

type IntString struct {
	String string
	Int    int
}

type TaggedInt struct {
	I   int
	Tag Tag
}

type IntNamedString struct {
	NamedString
	Int int
}

type String string

type DimOrS struct {
	S string
	Dimension
}

func (ds DimOrS) String() string {
	if ds.S != "" {
		return ds.S
	}
	return ds.Dimension.String()
}

// OptionalRanges is either 'auto' or a slice of ranges.
type OptionalRanges struct {
	Ranges [][2]int
	Auto   bool
}

// GridDims is a compact form for a grid template
// dimension. It is either :
//   - a single value V
//   - minmax(V, V2)
//   - fit-content(V)
type GridDims struct {
	V, v2 DimOrS
	tag   byte // 0, 'm' for minmax()' or 'f' for fit-content()
}

// NewGridDimsValue returns a non tagged value.
func NewGridDimsValue(v DimOrS) GridDims { return GridDims{V: v} }

// NewGridDimsMinmax returns minmax(...)
func NewGridDimsMinmax(v1, v2 DimOrS) GridDims { return GridDims{tag: 'm', V: v1, v2: v2} }

// NewGridDimsFitcontent returns fit-content(...)
func NewGridDimsFitcontent(v Dimension) GridDims { return GridDims{tag: 'f', V: v.ToValue()} }

func (size GridDims) SizingFunctions() [2]DimOrS {
	minSizing, maxSizing := size.V, size.V
	if size.tag == 'm' {
		minSizing, maxSizing = size.V, size.v2
	}
	if size.tag == 'f' {
		minSizing, maxSizing = SToV("auto"), SToV("auto")
	} else if minSizing.Unit == Fr {
		minSizing = SToV("auto")
	}
	return [2]DimOrS{minSizing, maxSizing}
}

func (size GridDims) IsMinmax() (min, max DimOrS, ok bool) {
	return size.V, size.v2, size.tag == 'm'
}

func (size GridDims) IsFitcontent() (v DimOrS, ok bool) {
	return size.V, size.tag == 'f'
}

type GridAuto []GridDims

func (ga GridAuto) Cycle() *GridAutoIter {
	return &GridAutoIter{ga, 0}
}

// Reverse returns a new, reversed slice
func (ga GridAuto) Reverse() GridAuto {
	out := make(GridAuto, len(ga))
	for i, v := range ga {
		out[len(ga)-1-i] = v
	}
	return out
}

type GridAutoIter struct {
	src GridAuto
	pos int
}

func (gai *GridAutoIter) Next() GridDims {
	out := gai.src[gai.pos%len(gai.src)]
	gai.pos++
	return out
}

// See https://developer.mozilla.org/en-US/docs/Web/CSS/grid-row-start
type GridLine struct {
	Ident string
	Val   int
	Tag   Tag // Auto, Span or 0
}

func (gl GridLine) IsCustomIdent() bool { return gl.Val == 0 && gl.Tag == 0 }

// Span returns true for "span" attributes. In this case, the [Val] field is valid.
func (gl *GridLine) IsSpan() bool { return gl.Tag == Span }

func (gl *GridLine) IsAuto() bool { return gl.Tag == Auto }

// An empty list means 'none'
type GridTemplateAreas [][]string

// IsNone returns true for the CSS 'none' keyword
func (gt GridTemplateAreas) IsNone() bool { return len(gt) == 0 }

type GridTemplate struct {
	Tag Tag
	// Every even value is a [GridNames]
	Names []GridSpec
}

type (
	GridSpec interface {
		isGridSpec()
	}
	GridNames      []string
	GridNameRepeat struct { // only found in subgrid
		Names  [][]string
		Repeat int // RepeatAutoFill, >= 1 otherwise
	}

	GridRepeat struct {
		// Every even value is a [GridNames]
		Names  []GridSpec
		Repeat int
	}
)

func (GridNames) isGridSpec()      {}
func (GridDims) isGridSpec()       {}
func (GridRepeat) isGridSpec()     {}
func (GridNameRepeat) isGridSpec() {}

const (
	RepeatAutoFill = -1
	RepeatAutoFit  = -2
)

// ---------------------- helpers types -----------------------------------

type SContentProp struct {
	ContentProperty ContentProperty
	String          string
}
type SContentProps []SContentProp

// Counters store a counter() or counters() attribute
type Counters struct {
	Name      string
	Separator string // optional, only valid for counters()
	Style     CounterStyleID
}

// guard for possible content properties
type InnerContent interface {
	isInnerContent()
}

type Unit uint8

func (u Unit) String() string {
	switch u {
	case Scalar: // means no unit, but a valid value
		return ""
	case Perc: // percentage (%)
		return "%"
	case Ex:
		return "ex"
	case Em:
		return "em"
	case Ch:
		return "ch"
	case Rem:
		return "rem"
	case Px:
		return "px"
	case Pt:
		return "pt"
	case Pc:
		return "pc"
	case In:
		return "in"
	case Cm:
		return "cm"
	case Mm:
		return "mm"
	case Q:
		return "q"
	case Rad:
		return "rad"
	case Turn:
		return "turn"
	case Deg:
		return "deg"
	case Grad:
		return "grad"
	case Fr:
		return "fr"
	default:
		return "<invalid unit>"
	}
}

// Dimension without unit is interpreted as float
type Dimension struct {
	Value Float
	Unit  Unit
}

func NewDim(v Float, u Unit) Dimension { return Dimension{v, u} }

func (d Dimension) String() string {
	return fmt.Sprintf("<%g %s>", d.Value, d.Unit)
}

type BoolString struct {
	String string
	Bool   bool
}

type FloatString struct {
	String string
	Float  Fl
}

type SBoolFloat struct {
	String string
	Bool   bool
	Float  Fl
}

// SFloatStrings is either a string or a list of (string, float) pairs
type SFloatStrings struct {
	String string
	Values []FloatString
}

type Quote struct {
	Open   bool
	Insert bool
}

// Might be an existing image or a gradient
type Image interface {
	// InnerContent
	CssProperty
	isImage()
}

type (
	NoneImage struct{}
	UrlImage  string
)

type ColorStop struct {
	Color    Color
	Position Dimension
}

type DirectionType struct {
	Corner string
	Angle  Fl
}

type GradientSize struct {
	Keyword  string
	Explicit Point
}

type ColorsStops []ColorStop

type LinearGradient struct {
	ColorStops ColorsStops
	Direction  DirectionType
	Repeating  bool
}

type RadialGradient struct {
	ColorStops ColorsStops
	Shape      string
	Size       GradientSize
	Center     Center
	Repeating  bool
}

func (v BoolString) IsNone() bool {
	return v == BoolString{}
}

func (v Center) IsNone() bool {
	return v == Center{}
}

func (v IntString) IsNone() bool {
	return v == IntString{}
}

func (v Ints3) IsNone() bool {
	return v == Ints3{}
}

func (v Marks) IsNone() bool {
	return v == Marks{}
}

func (v Decorations) IsNone() bool { return len(v) == 0 }

func (v NamedString) IsNone() bool {
	return v == NamedString{}
}

func (v CounterStyleID) IsNone() bool {
	return v.Type == "" && v.Name == "" && v.Symbols == nil
}

func (v Point) IsNone() bool {
	return v == Point{}
}

func (v Quotes) IsNone() bool {
	return v.Tag == 0 && v.Open == nil && v.Close == nil
}

func (v SContent) IsNone() bool {
	return v.String == "" && v.Contents == nil
}

func (v SIntStrings) IsNone() bool {
	return v.String == "" && v.Values == nil
}

func (v SStrings) IsNone() bool {
	return v.String == "" && v.Strings == nil
}

func (v StringSet) IsNone() bool {
	return v.String == ""
}

func (v DimOrS) IsNone() bool {
	return v == DimOrS{}
}

func (v AttrData) IsNone() bool {
	return v.Name == "" && v.TypeOrUnit == "" && v.Fallback == nil
}

func (v ContentProperty) IsNone() bool {
	return v.Type == ""
}

func (v DirectionType) IsNone() bool {
	return v == DirectionType{}
}

func (v Dimension) IsNone() bool {
	return v == Dimension{}
}

func (v ColorStop) IsNone() bool {
	return v == ColorStop{}
}

func (v GradientSize) IsNone() bool {
	return v == GradientSize{}
}

func (v LinearGradient) IsNone() bool {
	return v.ColorStops == nil && v.Direction == DirectionType{} && !v.Repeating
}

func (v Quote) IsNone() bool {
	return v == Quote{}
}

func (v OptionalRanges) IsNone() bool {
	return v.Ranges == nil && !v.Auto
}

func (v Size) IsNone() bool {
	return v == Size{}
}

func (v RadialGradient) IsNone() bool {
	return v.ColorStops == nil && v.Shape == "" && v.Size == GradientSize{} && v.Center == Center{} && !v.Repeating
}

func (v SContentProp) IsNone() bool {
	return v.String == "" && v.ContentProperty.IsNone()
}

func (v SDimensions) IsNone() bool {
	return v.String == "" && v.Dimensions == nil
}

func (v IntNamedString) IsNone() bool {
	return v == IntNamedString{}
}

func (v Counters) IsNone() bool {
	return v.Name == "" && v.Separator == "" && v.Style.IsNone()
}

func (v GridDims) IsNone() bool {
	return v.tag == 0 && v.V.IsNone() && v.v2.IsNone()
}

// method tags

func (TaggedString) isCssProperty()      {}
func (TaggedInt) isCssProperty()         {}
func (Display) isCssProperty()           {}
func (BoolString) isCssProperty()        {}
func (SFloatStrings) isCssProperty()     {}
func (SBoolFloat) isCssProperty()        {}
func (Center) isCssProperty()            {}
func (Centers) isCssProperty()           {}
func (Color) isCssProperty()             {}
func (ContentProperties) isCssProperty() {}
func (Float) isCssProperty()             {}
func (Images) isCssProperty()            {}
func (Int) isCssProperty()               {}
func (IntString) isCssProperty()         {}
func (Ints3) isCssProperty()             {}
func (Marks) isCssProperty()             {}
func (Decorations) isCssProperty()       {}
func (NamedString) isCssProperty()       {}
func (CounterStyleID) isCssProperty()    {}
func (Page) isCssProperty()              {}
func (Point) isCssProperty()             {}
func (Quotes) isCssProperty()            {}
func (Repeats) isCssProperty()           {}
func (SContent) isCssProperty()          {}
func (SIntStrings) isCssProperty()       {}
func (SStrings) isCssProperty()          {}
func (Sizes) isCssProperty()             {}
func (String) isCssProperty()            {}
func (StringSet) isCssProperty()         {}
func (Strings) isCssProperty()           {}
func (Transforms) isCssProperty()        {}
func (DimOrS) isCssProperty()            {}
func (Values) isCssProperty()            {}
func (DimOrS4) isCssProperty()           {}
func (DimOrS5) isCssProperty()           {}
func (AttrData) isCssProperty()          {}
func (NoneImage) isCssProperty()         {}
func (UrlImage) isCssProperty()          {}
func (LinearGradient) isCssProperty()    {}
func (RadialGradient) isCssProperty()    {}
func (GridAuto) isCssProperty()          {}
func (GridLine) isCssProperty()          {}
func (GridTemplateAreas) isCssProperty() {}
func (GridTemplate) isCssProperty()      {}

func (TaggedString) isDeclaredValue()      {}
func (TaggedInt) isDeclaredValue()         {}
func (Display) isDeclaredValue()           {}
func (BoolString) isDeclaredValue()        {}
func (SFloatStrings) isDeclaredValue()     {}
func (SBoolFloat) isDeclaredValue()        {}
func (Center) isDeclaredValue()            {}
func (Centers) isDeclaredValue()           {}
func (Color) isDeclaredValue()             {}
func (ContentProperties) isDeclaredValue() {}
func (Float) isDeclaredValue()             {}
func (Images) isDeclaredValue()            {}
func (Int) isDeclaredValue()               {}
func (IntString) isDeclaredValue()         {}
func (Ints3) isDeclaredValue()             {}
func (Marks) isDeclaredValue()             {}
func (Decorations) isDeclaredValue()       {}
func (NamedString) isDeclaredValue()       {}
func (CounterStyleID) isDeclaredValue()    {}
func (Page) isDeclaredValue()              {}
func (Point) isDeclaredValue()             {}
func (Quotes) isDeclaredValue()            {}
func (Repeats) isDeclaredValue()           {}
func (SContent) isDeclaredValue()          {}
func (SIntStrings) isDeclaredValue()       {}
func (SStrings) isDeclaredValue()          {}
func (Sizes) isDeclaredValue()             {}
func (String) isDeclaredValue()            {}
func (StringSet) isDeclaredValue()         {}
func (Strings) isDeclaredValue()           {}
func (Transforms) isDeclaredValue()        {}
func (DimOrS) isDeclaredValue()            {}
func (Values) isDeclaredValue()            {}
func (DimOrS4) isDeclaredValue()           {}
func (DimOrS5) isDeclaredValue()           {}
func (AttrData) isDeclaredValue()          {}
func (NoneImage) isDeclaredValue()         {}
func (UrlImage) isDeclaredValue()          {}
func (LinearGradient) isDeclaredValue()    {}
func (RadialGradient) isDeclaredValue()    {}
func (GridAuto) isDeclaredValue()          {}
func (GridLine) isDeclaredValue()          {}
func (GridTemplateAreas) isDeclaredValue() {}
func (GridTemplate) isDeclaredValue()      {}
