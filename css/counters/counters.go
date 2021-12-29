package counters

import (
	"fmt"
	"math"
	"strings"

	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils"
)

// Implement the various counter types and list-style-type values.
//
// These are defined in the same terms as CSS 3 Lists:
// http://dev.w3.org/csswg/css3-lists/#predefined-counters

type CounterStyle map[string]CounterStyleDescriptors

// may return nil
func (c CounterStyle) resolveCounter(counterName string, previousTypes utils.Set) *CounterStyleDescriptors {
	counter, has := c[counterName]
	if !has {
		return nil
	}

	// Avoid circular fallbacks
	if previousTypes == nil {
		previousTypes = utils.NewSet()
	} else if previousTypes.Has(counterName) {
		return nil
	}
	previousTypes.Add(counterName)

	extends, system := "", "symbolic"
	if counter.System != (CounterStyleSystem{}) {
		extends, system = counter.System.Extends, counter.System.System
	}

	// Handle extends
	for extends != "" {
		if extendedCounter, has := c[system]; has {
			counter.System = extendedCounter.System
			previousTypes.Add(system)

			extends, system = "", "symbolic"
			if counter.System != (CounterStyleSystem{}) {
				extends, system = counter.System.Extends, counter.System.System
			}

			if extends != "" && previousTypes.Has(system) {
				extends, system = "extends", "decimal"
				continue
			}
			counter.merge(extendedCounter)
		} else {
			return &counter
		}
	}

	return &counter
}

func (c CounterStyle) resolveCounterStyle(counterStyle pr.CounterStyleID, previousTypes utils.Set) *CounterStyleDescriptors {
	if counterType := counterStyle.Type; counterType == "symbols()" || counterType == "string" {
		var out CounterStyleDescriptors
		if counterType == "string" {
			out.System = CounterStyleSystem{"", "cyclic", -1}
			out.Symbols = []pr.NamedString{{Name: "string", String: counterStyle.Name}}
			out.Suffix = pr.NamedString{Name: "string", String: ""}
		} else if counterType == "symbols()" {
			out.System = CounterStyleSystem{"", counterStyle.Name, -1}
			if counterStyle.Name == "fixed" {
				out.System.Number = 1
			}
			for _, argument := range counterStyle.Symbols {
				out.Symbols = append(out.Symbols, pr.NamedString{Name: "string", String: argument})
			}
			out.Suffix = pr.NamedString{Name: "string", String: " "}
		}
		out.Negative = [2]pr.NamedString{{Name: "string", String: "-"}, {Name: "string", String: ""}}
		out.Prefix = pr.NamedString{Name: "string", String: ""}
		out.Range.Auto = true
		out.Fallback = "decimal"
		return &out
	}

	counterName := counterStyle.Name
	return c.resolveCounter(counterName, previousTypes)
}

// Generate the counter representation.
//
// See https://www.w3.org/TR/css-counter-styles-3/#generate-a-counter
func (c CounterStyle) RenderValue(counterValue int, counterStyleName string) string {
	return c.renderValue(counterValue, c.resolveCounter(counterStyleName, nil), nil)
}

// RenderValueStyle is the same as `RenderValue`, for a general counter style.
func (c CounterStyle) RenderValueStyle(counterValue int, counterStyle pr.CounterStyleID) string {
	return c.renderValue(counterValue, c.resolveCounterStyle(counterStyle, nil), nil)
}

func (c CounterStyle) renderValue(counterValue int, counter *CounterStyleDescriptors, previousTypes utils.Set) string {
	if counter == nil {
		if _, has := c["decimal"]; has {
			return c.RenderValue(counterValue, "decimal")
		}
		// Could happen if the UA stylesheet is not used
		return ""
	}

	extends, system, fixedNumber := "", "symbolic", -1
	if counter.System != (CounterStyleSystem{}) {
		extends, system, fixedNumber = counter.System.Extends, counter.System.System, counter.System.Number
	}

	// Avoid circular fallbacks
	if previousTypes == nil {
		previousTypes = utils.NewSet()
	} else if previousTypes.Has(system) {
		return c.RenderValue(counterValue, "decimal")
	}

	// Handle extends
	for extends != "" {
		if extendedCounter, has := c[system]; has {
			counter.System = extendedCounter.System

			extends, system, fixedNumber = "", "symbolic", -1
			if counter.System != (CounterStyleSystem{}) {
				extends, system, fixedNumber = counter.System.Extends, counter.System.System, counter.System.Number
			}
			if previousTypes.Has(system) {
				return c.RenderValue(counterValue, "decimal")
			}
			previousTypes.Add(system)
			counter.merge(extendedCounter)
		} else {
			return c.RenderValue(counterValue, "decimal")
		}
	}

	// Step 2
	counterRanges := counter.Range.Ranges
	if counter.Range.Auto || counter.Range.IsNone() {
		minRange, maxRange := math.MinInt32, math.MaxInt32
		if system == "alphabetic" || system == "symbolic" {
			minRange = 1
		} else if system == "additive" {
			minRange = 0
		}
		counterRanges = [][2]int{{minRange, maxRange}}
	}
	var found bool
	for _, v := range counterRanges {
		if v[0] <= counterValue && counterValue <= v[1] {
			found = true
			break
		}
	}

	if !found {
		return c.renderValue(counterValue, c.resolveCounter(counter.fallback(), previousTypes), previousTypes)
	}

	// Step 3
	var (
		negativePrefix, negativeSuffix string
		useNegative                    bool
	)
	isNegative := counterValue < 0
	if isNegative {
		vs := counter.Negative
		if vs == ([2]pr.NamedString{}) {
			vs = [2]pr.NamedString{{Name: "string", String: "-"}, {Name: "string", String: ""}}
		}
		negativePrefix, negativeSuffix = symbol(vs[0]), symbol(vs[1])
		useNegative = system == "symbolic" || system == "alphabetic" || system == "numeric" || system == "additive"
		if useNegative {
			counterValue = abs(counterValue)
		}
	}

	var (
		initial string
		ok      bool
	)
	switch system {
	case "cyclic":
		initial, ok = repeating(counter.Symbols, counterValue)
		if !ok {
			return c.RenderValue(counterValue, "decimal")
		}
	case "fixed":
		if len(counter.Symbols) == 0 {
			return c.RenderValue(counterValue, "decimal")
		}
		initial, ok = nonRepeating(counter.Symbols, fixedNumber, counterValue)
		if !ok {
			return c.renderValue(counterValue, c.resolveCounter(counter.fallback(), previousTypes), previousTypes)
		}
	case "symbolic":
		initial, ok = symbolic(counter.Symbols, counterValue)
		if !ok {
			return c.RenderValue(counterValue, "decimal")
		}
	case "alphabetic":
		initial, ok = alphabetic(counter.Symbols, counterValue)
		if !ok {
			return c.RenderValue(counterValue, "decimal")
		}
	case "numeric":
		initial, ok = numeric(counter.Symbols, counterValue)
		if !ok {
			return c.RenderValue(counterValue, "decimal")
		}
	case "additive":
		if len(counter.AdditiveSymbols) == 0 {
			return c.RenderValue(counterValue, "decimal")
		}
		initial, ok = additive(counter.AdditiveSymbols, counterValue)
		if !ok {
			return c.renderValue(counterValue, c.resolveCounter(counter.fallback(), previousTypes), previousTypes)
		}
	}

	// Step 4
	pad := counter.Pad
	padDifference := pad.Int - len(initial)
	if isNegative && useNegative {
		padDifference -= len(negativePrefix) + len(negativeSuffix)
	}
	if padDifference > 0 {
		initial = strings.Repeat(symbol(pad.NamedString), padDifference) + initial
	}

	// Step 5
	if isNegative && useNegative {
		initial = negativePrefix + initial + negativeSuffix
	}

	// Step 6
	return initial
}

func symbol(value pr.NamedString) string {
	if value.Name == "string" {
		return value.String
	}
	return ""
}

// Implement the algorithm for `type: repeating`.
func repeating(symbols []pr.NamedString, value int) (string, bool) {
	if len(symbols) == 0 {
		return "", false
	}
	return symbol(symbols[(value-1)%len(symbols)]), true
}

// Implement the algorithm for `type: non-repeating`.
func nonRepeating(symbols []pr.NamedString, firstValue, value int) (string, bool) {
	L := len(symbols)
	value -= firstValue
	if 0 <= value && value < L {
		return symbol(symbols[value]), true
	}
	return "", false
}

// Implement the algorithm for `type: symbolic`.
func symbolic(symbols []pr.NamedString, value int) (string, bool) {
	if len(symbols) == 0 {
		return "", false
	}
	L := len(symbols)
	index := (value - 1) % L
	repeat := (value-1)/L + 1
	return strings.Repeat(symbol(symbols[index]), repeat), true
}

// Implement the algorithm for `type: alphabetic`.
func alphabetic(symbols []pr.NamedString, value int) (string, bool) {
	L := len(symbols)
	if L < 2 {
		return "", false
	}
	reversedParts := []string{}
	for value != 0 {
		value -= 1
		reversedParts = append(reversedParts, symbol(symbols[value%L]))
		value /= L
	}
	reverse(reversedParts)
	return strings.Join(reversedParts, ""), true
}

// Implement the algorithm for `type: numeric`.
func numeric(symbols []pr.NamedString, value int) (string, bool) {
	if value == 0 {
		return symbol(symbols[0]), true
	}
	if len(symbols) < 2 {
		return "", false
	}
	var reversedParts []string
	value = abs(value)
	L := len(symbols)
	for value != 0 {
		reversedParts = append(reversedParts, symbol(symbols[value%L]))
		value /= L
	}
	reverse(reversedParts)
	return strings.Join(reversedParts, ""), true
}

// Implement the algorithm for `type: additive`.
func additive(symbols []pr.IntNamedString, value int) (string, bool) {
	if value == 0 {
		for _, vs := range symbols {
			if vs.Int == 0 {
				return symbol(vs.NamedString), true
			}
		}
	}
	if len(symbols) == 0 {
		return "", false
	}
	var parts []string
	for _, vs := range symbols {
		repetitions := value / vs.Int
		parts = append(parts, strings.Repeat(symbol(vs.NamedString), repetitions))
		value -= vs.Int * repetitions
		if value == 0 {
			return strings.Join(parts, ""), true
		}
	}
	return "", false //  Failed to find a representation for this value
}

// Generates the content of a ::marker pseudo-element, for the given value.
func (c CounterStyle) RenderMarker(counterName pr.CounterStyleID, counterValue int) string {
	counter := c.resolveCounterStyle(counterName, nil)
	if counter == nil {
		if _, has := c["decimal"]; has {
			return c.RenderMarker(pr.CounterStyleID{Name: "decimal"}, counterValue)
		}
		// Could happen if the UA stylesheet is ! used
		return ""
	}

	prefix := symbol(counter.Prefix)
	suffix := counter.Suffix
	if suffix.IsNone() {
		suffix = pr.NamedString{Name: "string", String: ". "}
	}
	suffixS := symbol(suffix)

	value := c.renderValue(counterValue, counter, nil)
	return prefix + value + suffixS
}

type CounterStyleDescriptors struct {
	Negative        [2]pr.NamedString
	Prefix          pr.NamedString
	Suffix          pr.NamedString
	Fallback        string
	System          CounterStyleSystem
	Pad             pr.IntNamedString
	Symbols         []pr.NamedString
	AdditiveSymbols []pr.IntNamedString
	Range           pr.OptionalRanges
}

func (desc *CounterStyleDescriptors) Validate() error {
	system := desc.System
	if system == (CounterStyleSystem{}) {
		system = CounterStyleSystem{System: "symbolic"}
	}
	if system.Extends == "" {
		switch system.System {
		case "cyclic", "fixed", "symbolic":
			if len(desc.Symbols) == 0 {
				return fmt.Errorf("counter style %s needs at least one symbol", system.System)
			}
		case "alphabetic", "numeric":
			if len(desc.Symbols) < 2 {
				return fmt.Errorf("counter style %s needs at least two symbols", system.System)
			}
		case "additive":
			if len(desc.AdditiveSymbols) < 2 {
				return fmt.Errorf("counter style %s needs at least two additive symbols", system.System)
			}
		}
	}
	return nil
}

func (desc *CounterStyleDescriptors) fallback() string {
	if desc.Fallback != "" {
		return desc.Fallback
	}
	return "decimal"
}

// complete the fields of `desc` with those taken from `src`
func (desc *CounterStyleDescriptors) merge(src CounterStyleDescriptors) {
	if desc.Negative == ([2]pr.NamedString{}) {
		desc.Negative = src.Negative
	}
	if desc.System == (CounterStyleSystem{}) {
		desc.System = src.System
	}
	if desc.Prefix.IsNone() {
		desc.Prefix = src.Prefix
	}
	if desc.Suffix.IsNone() {
		desc.Suffix = src.Suffix
	}
	if desc.Range.IsNone() {
		desc.Range = src.Range
	}
	if desc.Fallback == "" {
		desc.Fallback = src.Fallback
	}
	if desc.Pad.IsNone() {
		desc.Pad = src.Pad
	}
	if desc.Symbols == nil {
		desc.Symbols = src.Symbols
	}
	if desc.AdditiveSymbols == nil {
		desc.AdditiveSymbols = src.AdditiveSymbols
	}
}

type CounterStyleSystem struct {
	Extends, System string
	Number          int
}

func reverse(a []string) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}

func abs(v int) int {
	return int(math.Abs(float64(v)))
}
