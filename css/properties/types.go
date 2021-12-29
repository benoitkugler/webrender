package properties

import (
	"github.com/benoitkugler/webrender/css/parser"
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

type SContent struct {
	String   string
	Contents ContentProperties
}

type Display [3]string

type Decorations utils.Set

type Transforms []SDimensions

type Values []Value

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

type Quotes struct {
	Open  Strings
	Close Strings
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

type Page struct {
	String string
	Valid  bool
	Page   int
}

// Dimension or "auto" or "cover" or "contain"
type Size struct {
	String string
	Width  Value
	Height Value
}

type Center struct {
	OriginX string
	OriginY string
	Pos     Point
}

type Color parser.Color

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

type Point [2]Dimension

type Marks struct {
	Crop  bool
	Cross bool
}

type IntString struct {
	String string
	Int    int
}

type IntNamedString struct {
	NamedString
	Int int
}

type String string

type Value struct {
	String string
	Dimension
}

// OptionalRanges is either 'auto' or a slice of ranges.
type OptionalRanges struct {
	Ranges [][2]int
	Auto   bool
}

type NamedProperty struct {
	Name     string
	Property ValidatedProperty
}

type NamedProperties []NamedProperty

// ---------------------- helpers types -----------------------------------
type RawTokens []parser.Token

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

// Dimension without unit is interpreted as float
type Dimension struct {
	Unit  Unit
	Value Float
}

type BoolString struct {
	String string
	Bool   bool
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

func (Display) isCssProperty() {}

func (BoolString) isCssProperty() {}

func (v BoolString) IsNone() bool {
	return v == BoolString{}
}

func (Center) isCssProperty() {}
func (v Center) IsNone() bool {
	return v == Center{}
}

func (Centers) isCssProperty() {}

func (Color) isCssProperty() {}

func (ContentProperties) isCssProperty() {}

func (Float) isCssProperty() {}

func (Images) isCssProperty() {}

func (Int) isCssProperty() {}

func (IntString) isCssProperty() {}
func (v IntString) IsNone() bool {
	return v == IntString{}
}

func (Ints3) isCssProperty() {}
func (v Ints3) IsNone() bool {
	return v == Ints3{}
}

func (Marks) isCssProperty() {}
func (v Marks) IsNone() bool {
	return v == Marks{}
}

func (Decorations) isCssProperty() {}
func (v Decorations) IsNone() bool { return len(v) == 0 }

func (NamedString) isCssProperty() {}
func (v NamedString) IsNone() bool {
	return v == NamedString{}
}

func (CounterStyleID) isCssProperty() {}
func (v CounterStyleID) IsNone() bool {
	return v.Type == "" && v.Name == "" && v.Symbols == nil
}

func (Page) isCssProperty() {}
func (v Page) IsNone() bool {
	return v == Page{}
}

func (Point) isCssProperty() {}
func (v Point) IsNone() bool {
	return v == Point{}
}

func (Quotes) isCssProperty() {}
func (v Quotes) IsNone() bool {
	return v.Open == nil && v.Close == nil
}

func (Repeats) isCssProperty() {}

func (SContent) isCssProperty() {}
func (v SContent) IsNone() bool {
	return v.String == "" && v.Contents == nil
}

func (SIntStrings) isCssProperty() {}
func (v SIntStrings) IsNone() bool {
	return v.String == "" && v.Values == nil
}

func (SStrings) isCssProperty() {}
func (v SStrings) IsNone() bool {
	return v.String == "" && v.Strings == nil
}

func (Sizes) isCssProperty() {}

func (String) isCssProperty() {}

func (StringSet) isCssProperty() {}
func (v StringSet) IsNone() bool {
	return v.String == ""
}

func (Strings) isCssProperty() {}

func (Transforms) isCssProperty() {}

func (Value) isCssProperty() {}
func (v Value) IsNone() bool {
	return v == Value{}
}

func (Values) isCssProperty() {}

func (AttrData) isCssProperty() {}
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

func (v NamedProperty) IsNone() bool {
	return v.Name == "" && v.Property.IsNone()
}

func (v ColorStop) IsNone() bool {
	return v == ColorStop{}
}

func (v GradientSize) IsNone() bool {
	return v == GradientSize{}
}

func (v LinearGradient) IsNone() bool {
	return v.ColorStops == nil && v.Direction == DirectionType{} && v.Repeating == false
}

func (v Quote) IsNone() bool {
	return v == Quote{}
}

func (v OptionalRanges) IsNone() bool {
	return v.Ranges == nil && v.Auto == false
}

func (v Size) IsNone() bool {
	return v == Size{}
}

func (v RadialGradient) IsNone() bool {
	return v.ColorStops == nil && v.Shape == "" && v.Size == GradientSize{} && v.Center == Center{} && v.Repeating == false
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
