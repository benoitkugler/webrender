// This package defines the types needed to handle the various CSS properties.
// There are 3 groups of types for a property, separated by 2 steps : cascading and computation.
// Thus the need of 3 types (see below).
// Schematically, the style computation is :
//
//	ValidatedProperty (ComputedFromCascaded)-> CascadedProperty (Compute)-> CssProperty
package properties

import (
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/utils"
)

type Fl = utils.Fl

// DeclaredValue is the most general CSS input for a property,
// one of:
//   - the special "initial" or "inherited" keywords.
//   - a validated [CssProperty]
//   - a raw slice of tokens (containing var() tokens), pending validation
type DeclaredValue interface {
	isDeclaredValue()
}

func (DefaultValue) isDeclaredValue() {}
func (RawTokens) isDeclaredValue()    {}

type RawTokens []parser.Token

func (rt RawTokens) String() string {
	return parser.Serialize(rt)
}

// CssProperty is the final form of a css input, a.k.a. the computed value.
// Default values and "var()" have been resolved, and the raw steam ok tokens has been
// validated.
type CssProperty interface {
	DeclaredValue

	isCssProperty()
}

type DefaultValue uint8

const (
	Inherit DefaultValue = iota + 1
	Initial
)

func NewDefaultValue(s string) DefaultValue {
	if s == "initial" {
		return Initial
	}
	return Inherit
}

func (d DefaultValue) String() string {
	switch d {
	case Inherit:
		return "<inherit>"
	case Initial:
		return "<initial>"
	default:
		return "invalid value"
	}
}

type VarData struct {
	Name    string // name of a custom property
	Default RawTokens
}

func (v VarData) IsNone() bool {
	return v.Name == "" && v.Default == nil
}

// KnownProp efficiently encode a known CSS property
type KnownProp uint8

func (p KnownProp) String() string { return propsNames[p] }

func (p KnownProp) Key() PropKey { return PropKey{KnownProp: p} }

// Properties is a general container for computed properties.
//
// In addition to the generic acces, an attempt to provide a "type safe" way is provided through the
// GetXXX and SetXXX methods. It relies on the convention than all the keys should be present,
// and values never be nil.
// Empty values are then encoded by the zero value of the concrete type.
type Properties map[KnownProp]CssProperty

// Copy return a shallow copy.
func (p Properties) Copy() Properties {
	out := make(Properties, len(p))
	for name, v := range p {
		out[name] = v
	}
	return out
}

// UpdateWith merge the entries from `other` to `p`.
func (p Properties) UpdateWith(other Properties) {
	for k, v := range other {
		p[k] = v
	}
}

// SpecifiedAttributes stores the value of
// CSS properties as specified.
type SpecifiedAttributes struct {
	Float    String
	Display  Display
	Position BoolString
}

// TextRatioCache stores the 1ex/font_size or 1ch/font_size
// ratios, for each font.
type TextRatioCache struct {
	ratioCh map[string]Float
	ratioEx map[string]Float
}

func NewTextRatioCache() TextRatioCache {
	return TextRatioCache{ratioCh: make(map[string]Float), ratioEx: make(map[string]Float)}
}

func (tr TextRatioCache) Get(fontKey string, isCh bool) (f Float, ok bool) {
	if isCh {
		f, ok = tr.ratioCh[fontKey]
	} else {
		f, ok = tr.ratioEx[fontKey]
	}
	return
}

func (tr TextRatioCache) Set(fontKey string, isCh bool, f Float) {
	if isCh {
		tr.ratioCh[fontKey] = f
	} else {
		tr.ratioEx[fontKey] = f
	}
}

// PropKey stores a CSS property name, supporting variables.
type PropKey struct {
	Var string // with leading --
	KnownProp
}

func (pr PropKey) String() string {
	if pr.KnownProp != 0 {
		return pr.KnownProp.String()
	}
	return pr.Var
}

// ElementStyle defines a common interface to access style properties.
// Implementations will typically compute the property on the fly and cache the result.
type ElementStyle interface {
	StyleAccessor

	// Set is the generic method to set an arbitrary property.
	// Type accessors should be used when possible.
	Set(key PropKey, value CssProperty)

	// Get is the generic method to access an arbitrary property.
	// Type accessors should be used when possible.
	Get(key PropKey) CssProperty

	// Copy returns a deep copy of the style.
	Copy() ElementStyle

	ParentStyle() ElementStyle
	Variables() map[string]RawTokens

	Specified() SpecifiedAttributes

	Cache() TextRatioCache
}

var _ StyleAccessor = Properties(nil)
