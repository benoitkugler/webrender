// This package defines the types needed to handle the various CSS properties.
// There are 3 groups of types for a property, separated by 2 steps : cascading and computation.
// Thus the need of 3 types (see below).
// Schematically, the style computation is :
//
//	ValidatedProperty (ComputedFromCascaded)-> CascadedProperty (Compute)-> CssProperty
package properties

import (
	"github.com/benoitkugler/webrender/utils"
)

type Fl = utils.Fl

// ValidatedProperty is the most general CSS input for a property.
// It covers the following cases:
//   - a plain CSS value, including "initial" or "inherited" special cases (CssProperty)
//   - a var() call (VarData)
//   - an input not yet validated, used as definition of variable (RawTokens)
type ValidatedProperty struct {
	SpecialProperty specialProperty
	prop            CascadedProperty
}

func (c ValidatedProperty) IsNone() bool {
	return c.prop.IsNone() && c.SpecialProperty == nil
}

// ToCascaded will panic if c.SpecialProperty is not nil.
func (c ValidatedProperty) ToCascaded() CascadedProperty {
	if c.SpecialProperty != nil {
		panic("attempted to bypass the SpecialProperty of a ValidatedProperty")
	}
	return c.prop
}

// CascadedProperty is the second form of a CSS input :
// var() calls have been resolved and the remaining raw properties have been checked.
// It is thus either a plain CSS property, or a default value.
type CascadedProperty struct {
	prop    CssProperty
	Default DefaultKind
}

func (v CascadedProperty) IsNone() bool {
	return v.prop == nil && v.Default == 0
}

// ToCSS will panic if c.Default is not 0.
func (c CascadedProperty) ToCSS() CssProperty {
	if c.Default != 0 {
		panic("attempted to bypass the Default of a CascadedProperty")
	}
	return c.prop
}

// AsValidated wraps the property into a ValidatedProperty
func (c CascadedProperty) AsValidated() ValidatedProperty {
	return ValidatedProperty{prop: c}
}

// CssProperty is final form of a css input :
// default values, "var()" and raw tokens have been resolved.
// Note than a CssProperty can naturally be seen as a CascadedProperty, but not the other way around.

type CssProperty interface {
	isCssProperty()
}

type specialProperty interface {
	isSpecialProperty()
}

func (v VarData) isSpecialProperty()   {}
func (v RawTokens) isSpecialProperty() {}

type DefaultKind uint8

const (
	Inherit DefaultKind = iota + 1
	Initial
)

// AsCascaded wraps the default to a CascadedProperty
func (d DefaultKind) AsCascaded() CascadedProperty { return CascadedProperty{Default: d} }

type VarData struct {
	Name    string // name of a custom property
	Default RawTokens
}

func (v VarData) IsNone() bool {
	return v.Name == "" && v.Default == nil
}

// ---------- Convenience constructor -------------------------------

func AsCascaded(prop CssProperty) CascadedProperty {
	return CascadedProperty{prop: prop}
}

func AsValidated(spe specialProperty) ValidatedProperty {
	return ValidatedProperty{SpecialProperty: spe}
}

// Properties is the general container for validated, cascaded and computed properties.
// In addition to the generic acces, an attempt to provide a "type safe" way is provided through the
// GetXXX and SetXXX methods. It relies on the convention than all the keys should be present,
// and values never be nil.
// "None" values are then encoded by the zero value of the concrete type.
type Properties map[string]CssProperty

func (p Properties) Keys() []string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	return keys
}

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

// ElementStyle defines a common interface to access style properties.
// Implementations will typically compute the property on the fly and cache the result.
type ElementStyle interface {
	StyleAccessor

	// Set is the generic method to set an arbitrary property.
	// Type accessors should be used when possible.
	Set(key string, value CssProperty)

	// Get is the generic method to access an arbitrary property.
	// Type accessors should be used when possible.
	Get(key string) CssProperty

	// Copy returns a deep copy of the style.
	Copy() ElementStyle

	ParentStyle() ElementStyle
	Variables() map[string]ValidatedProperty

	Specified() SpecifiedAttributes

	Cache() TextRatioCache
}

var _ StyleAccessor = Properties(nil)
