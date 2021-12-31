package tree

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/html/layout/text"
	"github.com/benoitkugler/webrender/logger"

	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils"
)

// Convert *specified* property values (the result of the cascade and
//     inheritance) into *computed* values (that are inherited).

// :copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

var (
	// These are unspecified, other than 'thin' <='medium' <= 'thick'.
	// Values are in pixels.
	BorderWidthKeywords = map[string]pr.Float{
		"thin":   1,
		"medium": 3,
		"thick":  5,
	}

	// http://www.w3.org/TR/CSS21/fonts.html#propdef-font-weight
	FontWeightRelative = struct {
		bolder, lighter map[int]int
	}{
		bolder: map[int]int{
			100: 400,
			200: 400,
			300: 400,
			400: 700,
			500: 700,
			600: 900,
			700: 900,
			800: 900,
			900: 900,
		},
		lighter: map[int]int{
			100: 100,
			200: 100,
			300: 100,
			400: 100,
			500: 100,
			600: 400,
			700: 400,
			800: 700,
			900: 700,
		},
	}

	// Maps property names to functions returning the computed values
	computerFunctions = map[string]computerFunc{}

	// to avoid declaration cycle
	tmp = map[string]computerFunc{
		"background_image":    backgroundImage,
		"background_position": backgroundPosition,
		"object_position":     backgroundPosition,
		"transform_origin":    transformOrigin,

		"border_spacing":             borderSpacing,
		"size":                       size,
		"clip":                       clip,
		"border_top_left_radius":     borderRadius,
		"border_top_right_radius":    borderRadius,
		"border_bottom_left_radius":  borderRadius,
		"border_bottom_right_radius": borderRadius,

		"break_before": break_,
		"break_after":  break_,

		"top":                  length,
		"right":                length,
		"left":                 length,
		"bottom":               length,
		"margin_top":           length,
		"margin_right":         length,
		"margin_bottom":        length,
		"margin_left":          length,
		"height":               length,
		"width":                length,
		"min_width":            length,
		"min_height":           length,
		"max_width":            length,
		"max_height":           length,
		"padding_top":          length,
		"padding_right":        length,
		"padding_bottom":       length,
		"padding_left":         length,
		"text_indent":          length,
		"hyphenate_limit_zone": length,
		"flex_basis":           length,

		"bleed_left":          bleed,
		"bleed_right":         bleed,
		"bleed_top":           bleed,
		"bleed_bottom":        bleed,
		"letter_spacing":      pixelLength,
		"background_size":     backgroundSize,
		"border_top_width":    borderWidth,
		"border_right_width":  borderWidth,
		"border_left_width":   borderWidth,
		"border_bottom_width": borderWidth,
		"column_rule_width":   borderWidth,
		"outline_width":       borderWidth,
		"column_width":        columnWidth,
		"column_gap":          columnGap,
		"content":             content,
		"display":             display,
		"float":               floating,
		"font_size":           fontSize,
		"font_weight":         fontWeight,
		"line_height":         lineHeight,
		"anchor":              anchor,
		"lang":                lang,
		"tab_size":            tabSize,
		"transform":           transforms,
		"vertical_align":      verticalAlign,
		"word_spacing":        wordSpacing,
		"bookmark_label":      bookmarkLabel,
		"string_set":          stringSet,
		"link":                link,
	}

	keywordsValues []pr.Float
)

func init() {
	if pr.InitialValues.GetBorderTopWidth().Value != BorderWidthKeywords["medium"] {
		panic("border-top-width and medium should be the same !")
	}

	// In "portrait" orientation.
	for _, size := range pr.PageSizes {
		if size[0].Value > size[1].Value {
			panic("page size should be in portrait orientation")
		}
	}

	keywordsValues = make([]pr.Float, len(pr.FontSizeKeywords))
	for i, k := range pr.FontSizeKeywordsOrder {
		keywordsValues[i] = pr.FontSizeKeywords[k]
	}

	for k, v := range tmp {
		computerFunctions[k] = v
	}
}

type computerFunc = func(*ComputedStyle, string, pr.CssProperty) pr.CssProperty

func resolveVar(specified map[string]pr.ValidatedProperty, var_ pr.VarData) pr.RawTokens {
	knownVariableNames := utils.NewSet(var_.Name)
	default_ := var_.Default

	tmpVal, isSpecified := specified[var_.Name]

	computedValue, ok := tmpVal.SpecialProperty.(pr.RawTokens)

	// handle the initial case
	if ok && len(computedValue) == 1 {
		value := computedValue[0]
		if ident, ok := value.(parser.IdentToken); ok && ident.Value == "initial" {
			return default_
		}
	}

	if !isSpecified {
		computedValue = default_
	}
	var in bool
	// resolve potential variable cycle
	for len(computedValue) == 1 {
		varFunction := validation.CheckVarFunction(computedValue[0])
		if varFunction.IsNone() {
			break
		}
		if knownVariableNames.Has(varFunction.Name) {
			computedValue = default_
			break
		}
		knownVariableNames.Add(varFunction.Name)
		tmpVal, in = specified[varFunction.Name]
		if !in {
			computedValue = varFunction.Default
		} else {
			computedValue, _ = tmpVal.SpecialProperty.(pr.RawTokens)
		}
		default_ = varFunction.Default
	}
	return computedValue
}

// value is either a VarData or a normal property; it cannot be a RawToken
func computeVariable(varData pr.VarData, name string, computed map[string]pr.ValidatedProperty, baseUrl string, parentStyle pr.ElementStyle) (pr.CascadedProperty, bool) {
	alreadyComputedValue := false

	var newValue pr.CascadedProperty
	computedValue := resolveVar(computed, varData)
	if computedValue != nil {
		newValue, _ = validation.Validate(strings.ReplaceAll(name, "_", "-"), computedValue, baseUrl)
	}

	// See https://drafts.csswg.org/css-variables/#invalid-variables
	if newValue.IsNone() {
		logger.WarningLogger.Printf(`Unsupported computed value "%s" set in variable %s "
                "for property %s.`, parser.Serialize(computedValue), strings.ReplaceAll(varData.Name, "_", "-"), strings.ReplaceAll(name, "_", "-"))

		if _, in := pr.Inherited[name]; in && parentStyle != nil {
			alreadyComputedValue = true
			newValue = pr.AsCascaded(parentStyle.Get(name))
		} else {
			alreadyComputedValue = !pr.InitialNotComputed.Has(name)
			newValue = pr.AsCascaded(pr.InitialValues[name])
		}
	}

	return newValue, alreadyComputedValue
}

// backgroundImage computes lenghts in gradient background-image.
func backgroundImage(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Images)
	for i, image := range value {
		switch gradient := image.(type) {
		case pr.LinearGradient:
			for j, cl := range gradient.ColorStops {
				if !cl.Position.IsNone() {
					cl.Position = length2(computer, name, pr.Value{Dimension: cl.Position}, -1, false).Dimension
					gradient.ColorStops[j] = cl
				}
			}
			image = gradient
		case pr.RadialGradient:
			for j, cl := range gradient.ColorStops {
				if !cl.Position.IsNone() {
					cl.Position = length2(computer, name, pr.Value{Dimension: cl.Position}, -1, false).Dimension
					gradient.ColorStops[j] = cl
				}
			}
			gradient.Center = _backgroundPosition(computer, name, []pr.Center{gradient.Center})[0]
			if gradient.Size.IsExplicit() {
				l := _lengthOrPercentageTuple2(computer, name, gradient.Size.Explicit.ToSlice())
				gradient.Size.Explicit = pr.Point{l[0], l[1]}
			}
			image = gradient
		}
		value[i] = image
	}
	return value
}

func _backgroundPosition(computer *ComputedStyle, name string, value pr.Centers) pr.Centers {
	out := make(pr.Centers, len(value))
	for index, v := range value {
		out[index] = pr.Center{
			OriginX: v.OriginX,
			OriginY: v.OriginY,
			Pos: pr.Point{
				length2(computer, name, pr.Value{Dimension: v.Pos[0]}, -1, false).Dimension,
				length2(computer, name, pr.Value{Dimension: v.Pos[1]}, -1, false).Dimension,
			},
		}
	}
	return out
}

// backgroundPosition compute lengths in background-position.
func backgroundPosition(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Centers)
	return _backgroundPosition(computer, name, value)
}

func transformOrigin(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Point)
	l := _lengthOrPercentageTuple2(computer, name, value.ToSlice())
	return pr.Point{l[0], l[1]}
}

func clip(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	return lengths_(computer, name, _value.(pr.Values))
}

// Compute the lists of lengths that can be percentages.
// returns a slice with same length as input
func lengths_(computer *ComputedStyle, name string, value pr.Values) pr.Values {
	out := make(pr.Values, len(value))
	for index, v := range value {
		out[index] = length2(computer, name, v, -1, true)
	}
	return out
}

// Compute the lists of lengths that can be percentages.
func size(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Point)
	for index, v := range value {
		value[index] = length2(computer, name, v.ToValue(), -1, true).Dimension
	}
	return value
}

func borderSpacing(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Point)
	values := pr.Values{value[0].ToValue(), value[1].ToValue()}
	tmp := lengths_(computer, name, values)
	return pr.Point{tmp[0].Dimension, tmp[1].Dimension}
}

func borderRadius(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Point)
	var out pr.Point
	for index, v := range value {
		out[index] = length2(computer, name, v.ToValue(), -1, false).Dimension
	}
	return out
}

// Compute the lists of lengths that can be percentages.
func _lengthOrPercentageTuple2(computer *ComputedStyle, name string, value []pr.Dimension) []pr.Dimension {
	out := make([]pr.Dimension, len(value))
	for index, v := range value {
		out[index] = length2(computer, name, pr.Value{Dimension: v}, -1, false).Dimension
	}
	return out
}

// Compute the ``break-before`` and ``break-after`` pr.
func break_(_ *ComputedStyle, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.String)
	// "always" is defined as an alias to "page" in multi-column
	// https://www.w3.org/TR/css3-multicol/#column-breaks
	if value == "always" {
		return pr.String("page")
	}
	return value
}

func length(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	return length2(computer, name, value, -1, false)
}

func asPixels(v pr.Value, pixelsOnly bool) pr.Value {
	if pixelsOnly {
		v.Unit = pr.Scalar
	}
	return v
}

// Computes a length ``value``.
// passing a negative fontSize means null
// Always returns a Value which is interpreted as float64 if Unit is zero.
// pixelsOnly=false
func length2(computer *ComputedStyle, name string, value pr.Value, fontSize pr.Float, pixelsOnly bool) pr.Value {
	if value.String == "auto" || value.String == "content" {
		return value
	}
	if value.Value == 0 {
		return asPixels(pr.ZeroPixels.ToValue(), pixelsOnly)
	}

	var result pr.Float
	switch unit := value.Unit; unit {
	case pr.Px:
		return asPixels(value, pixelsOnly)
	case pr.Pt, pr.Pc, pr.In, pr.Cm, pr.Mm, pr.Q:
		// Convert absolute lengths to pixels
		result = value.Value * pr.LengthsToPixels[unit]
	case pr.Em, pr.Ex, pr.Ch, pr.Rem:
		if fontSize < 0 {
			fontSize = computer.GetFontSize().Value
		}
		switch unit {
		case pr.Ex:
			result = value.Value * fontSize * text.ExRatio(computer, computer.textContext)
		case pr.Ch:
			var letterSpacing pr.Value
			if name != "letter_spacing" { // avoid infinite recursion
				letterSpacing = computer.GetLetterSpacing()
			}
			result = value.Value * text.ChWidth(computer, fontSize, computer.textContext, letterSpacing)
		case pr.Em:
			result = value.Value * fontSize
		case pr.Rem:
			result = value.Value * computer.rootStyle.GetFontSize().Value
		}

	default:
		// A percentage or "auto": no conversion needed.
		return value
	}

	return asPixels(pr.Dimension{Value: result, Unit: pr.Px}.ToValue(), pixelsOnly)
}

func bleed(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if value.String == "auto" {
		if computer.GetMarks().Crop {
			return pr.Dimension{Value: 8, Unit: pr.Px}.ToValue() // 6pt
		}
		return pr.ZeroPixels.ToValue()
	}
	return length(computer, name, value)
}

func pixelLength(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if value.String == "normal" {
		return value
	}
	out := length2(computer, name, value, -1, true)
	return out
}

// Compute the ``background-size`` pr.
func backgroundSize(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Sizes)
	out := make(pr.Sizes, len(value))
	for index, v := range value {
		if v.String == "contain" || v.String == "cover" {
			out[index] = pr.Size{String: v.String}
		} else {
			out[index] = pr.Size{
				Width:  length2(computer, name, v.Width, -1, false),
				Height: length2(computer, name, v.Height, -1, false),
			}
		}
	}
	return out
}

// Compute the ``border-*-width`` pr.
// value.String may be the string representation of an int
func borderWidth(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	style := computer.Get(strings.ReplaceAll(name, "width", "style")).(pr.String)

	if style == "none" || style == "hidden" {
		return pr.FToV(0)
	}
	if bw, in := BorderWidthKeywords[value.String]; in {
		return bw.ToValue()
	}
	d := length2(computer, name, value, -1, true)
	return d
}

// Compute the ``column-width`` property.
func columnWidth(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	return length(computer, name, value)
}

// Compute the ``column-gap`` property.
func columnGap(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if value.String == "normal" {
		value = pr.Value{Dimension: pr.Dimension{Value: 1, Unit: pr.Em}}
	}
	return length(computer, name, value)
}

func computeAttrFunction(computer *ComputedStyle, values pr.AttrData) (out pr.ContentProperty, err error) {
	attrName, typeOrUnit, fallback := values.Name, values.TypeOrUnit, values.Fallback
	node, ok := computer.element.(*utils.HTMLNode)
	if !ok {
		return
	}

	var prop pr.InnerContent
	attrValue := node.Get(attrName)
	if attrValue == "" && fallback != nil {
		atrValue_, ok := fallback.(pr.InnerContent)
		if !ok {
			return out, fmt.Errorf("fallback type not supported : %T", fallback)
		}
		prop = atrValue_
	} else {
		switch typeOrUnit {
		case "string":
			prop = pr.String(attrValue) // Keep the string
		case "url":
			if strings.HasPrefix(attrValue, "#") {
				prop = pr.NamedString{Name: "internal", String: utils.Unquote(attrValue[1:])}
			} else {
				u, err := utils.SafeUrljoin(computer.baseUrl, attrValue, false)
				if err != nil {
					return out, err
				}
				prop = pr.NamedString{Name: "external", String: u}
			}
		case "color":
			prop = pr.Color(parser.ParseColorString(strings.TrimSpace(attrValue)))
		case "integer":
			i, err := strconv.Atoi(strings.TrimSpace(attrValue))
			if err != nil {
				return out, err
			}
			prop = pr.Int(i)
		case "number":
			f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 64)
			if err != nil {
				return out, err
			}
			prop = pr.Float(f)
		case "%":
			f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 64)
			if err != nil {
				return out, err
			}
			prop = pr.Dimension{Value: pr.Float(f), Unit: pr.Percentage}.ToValue()
			typeOrUnit = "length"
		default:
			unit, isUnit := validation.LENGTHUNITS[typeOrUnit]
			angle, isAngle := validation.AngleUnits[typeOrUnit]
			if isUnit {
				f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 64)
				if err != nil {
					return out, err
				}
				prop = pr.Dimension{Value: pr.Float(f), Unit: unit}.ToValue()
				typeOrUnit = "length"
			} else if isAngle {
				f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 64)
				if err != nil {
					return out, err
				}
				prop = pr.Dimension{Value: pr.Float(f), Unit: angle}.ToValue()
				typeOrUnit = "angle"
			}
		}
	}
	return pr.ContentProperty{Type: typeOrUnit, Content: prop}, nil
}

func contentList(computer *ComputedStyle, values pr.ContentProperties) (pr.ContentProperties, error) {
	var computedValues pr.ContentProperties
	for _, value := range values {
		var computedValue pr.ContentProperty
		switch value.Type {
		case "string", "content", "url", "quote", "leader()":
			computedValue = value
		case "attr()":
			attr, ok := value.Content.(pr.AttrData)
			if !ok || attr.TypeOrUnit != "string" {
				panic(fmt.Sprintf("invalid attr() property : %v", value.Content))
			}
			var err error
			computedValue, err = computeAttrFunction(computer, attr)
			if err != nil {
				return nil, err
			}
		case "counter()", "counters()", "content()", "element()", "string()":
			// Other values need layout context, their computed value cannot be
			// better than their specified value yet.
			// See build.computeContentList.
			computedValue = value
		case "target-counter()", "target-counters()", "target-text()":
			prop, ok := value.Content.(pr.SContentProps)
			if !ok || len(prop) == 0 {
				return nil, fmt.Errorf("expected a non empty list of String or ContentProperty, got %v", value.Content)
			}
			anchorToken := prop[0].ContentProperty
			if anchorToken.Type == "attr()" {
				proper, err := computeAttrFunction(computer, anchorToken.Content.(pr.AttrData))
				if err != nil {
					return nil, err
				}
				if !proper.IsNone() {
					computedValue = pr.ContentProperty{Type: value.Type, Content: append(pr.SContentProps{{ContentProperty: proper}}, prop[1:]...)}
				}
			} else {
				computedValue = value
			}
		}
		if computedValue.IsNone() {
			logger.WarningLogger.Printf("Unable to compute %v's value for content: %v\n", computer.element, value)
		} else {
			computedValues = append(computedValues, computedValue)
		}
	}
	return computedValues, nil
}

// Compute the ``bookmark-label`` property.
func bookmarkLabel(computer *ComputedStyle, _ string, _value pr.CssProperty) pr.CssProperty {
	if value, ok := _value.(pr.ContentProperties); ok {
		out, err := contentList(computer, value)
		if err != nil {
			logger.WarningLogger.Printf("error computing bookmark-label : %s\n", err)
			return pr.ContentProperties{}
		}
		return out
	}
	return pr.ContentProperties{}
}

// Compute the ``string-set`` property.
func stringSet(computer *ComputedStyle, _ string, _value pr.CssProperty) pr.CssProperty {
	// Spec asks for strings after custom keywords, but we allow content-lists
	if stringset, ok := _value.(pr.StringSet); ok {
		out := make(pr.SContents, len(stringset.Contents))
		for i, sset := range stringset.Contents {
			v, err := contentList(computer, sset.Contents)
			if err != nil {
				logger.WarningLogger.Printf("error computing string-set : %s \n", err)
				return pr.StringSet{}
			}
			out[i] = pr.SContent{String: sset.String, Contents: v}
		}
		return pr.StringSet{String: stringset.String, Contents: out}
	}
	return pr.StringSet{}
}

// Compute the ``content`` property.
func content(computer *ComputedStyle, _ string, _value pr.CssProperty) pr.CssProperty {
	if value, ok := _value.(pr.SContent); ok {
		if value.String == "normal" {
			if computer.pseudoType != "" {
				return pr.SContent{String: "inhibit"}
			} else {
				return pr.SContent{String: "contents"}
			}
		} else if value.String == "none" {
			return pr.SContent{String: "inhibit"}
		}
		props, err := contentList(computer, value.Contents)
		if err != nil {
			logger.WarningLogger.Printf("error computing content : %s\n", err)
			return pr.SContent{}
		}
		return pr.SContent{Contents: props}
	}
	return pr.SContent{}
}

// Compute the ``display`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func display(computer *ComputedStyle, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Display)
	float_ := computer.specified.Float
	position := computer.specified.Position
	if (!position.Bool && (position.String == "absolute" || position.String == "fixed")) || float_ != "none" || computer.isRootElement() {
		if value == (pr.Display{"inline-table"}) {
			return pr.Display{"block", "table"}
		} else if d := value[0]; value[1] == "" && value[2] == "" && strings.HasPrefix(d, "table-") {
			return pr.Display{"block", "flow"}
		} else if d == "inline" {
			if value.Has("list-item") {
				return pr.Display{"block", "flow", "list-item"}
			} else {
				return pr.Display{"block", "flow"}
			}
		}
	}
	return value
}

// Compute the ``float`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func floating(computer *ComputedStyle, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.String)
	position := computer.specified.Position
	if position.String == "absolute" || position.String == "fixed" {
		return pr.String("none")
	}
	return value
}

// Compute the ``font-size`` property.
func fontSize(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if fs, in := pr.FontSizeKeywords[value.String]; in {
		return fs.ToValue()
	}

	parentFontSize := pr.InitialValues.GetFontSize().Value
	if computer.parentStyle != nil {
		parentFontSize = computer.parentStyle.GetFontSize().Value
	}

	if value.String == "larger" {
		for _, keywordValue := range keywordsValues {
			if keywordValue > parentFontSize {
				return keywordValue.ToValue()
			}
		}
		return (parentFontSize * 1.2).ToValue()
	} else if value.String == "smaller" {
		for i := len(keywordsValues) - 1; i >= 0; i -= 1 {
			if keywordsValues[i] < parentFontSize {
				return (keywordsValues[i]).ToValue()
			}
		}
		return (parentFontSize * 0.8).ToValue()
	} else if value.Unit == pr.Percentage {
		return (value.Value * parentFontSize / 100.).ToValue()
	} else {
		return length2(computer, name, value, parentFontSize, true)
	}
}

// Compute the ``font-weight`` property.
func fontWeight(computer *ComputedStyle, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.IntString)
	var out int
	switch value.String {
	case "normal":
		out = 400
	case "bold":
		out = 700
	case "bolder":
		parentValue := computer.parentStyle.GetFontWeight().Int
		out = FontWeightRelative.bolder[parentValue]
	case "lighter":
		parentValue := computer.parentStyle.GetFontWeight().Int
		out = FontWeightRelative.lighter[parentValue]
	default:
		out = value.Int
	}
	return pr.IntString{Int: out}
}

// Compute the ``line-height`` property.
func lineHeight(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	var pixels pr.Float
	switch {
	case value.String == "normal":
		return value
	case value.Unit == pr.Scalar:
		return value
	case value.Unit == pr.Percentage:
		factor := value.Value / 100.
		fontSizeValue := computer.GetFontSize().Value
		pixels = factor * fontSizeValue
	default:
		pixels = length2(computer, name, value, -1, true).Value
	}
	return pr.Dimension{Value: pixels, Unit: pr.Px}.ToValue()
}

// Compute the ``anchor`` property.
func anchor(computer *ComputedStyle, _ string, _value pr.CssProperty) pr.CssProperty {
	// _value is either "none" or an AttrData
	attrData, ok := _value.(pr.AttrData)
	if !ok {
		return pr.String("")
	}
	if node, ok := computer.element.(*utils.HTMLNode); ok {
		anchorName := node.Get(attrData.Name)
		if anchorName == "" {
			return pr.String("")
		}
		return pr.String(anchorName)
	}
	return pr.String("")
}

// Compute the ``link`` property.
func link(computer *ComputedStyle, _ string, _value pr.CssProperty) pr.CssProperty {
	switch value := _value.(type) {
	case pr.NamedString:
		if value.String == "none" {
			return pr.NamedString{}
		} else {
			return value
		}

	case pr.AttrData:
		if node, ok := computer.element.(*utils.HTMLNode); ok {
			typeAttr, ok := utils.GetLinkAttribute(node, value.Name, computer.baseUrl)
			if !ok {
				return pr.NamedString{}
			}
			return pr.NamedString{Name: typeAttr[0], String: typeAttr[1]}
		}
	}
	return pr.NamedString{}
}

// Compute the ``lang`` property.
func lang(computer *ComputedStyle, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.NamedString)
	if value.String == "none" {
		return pr.NamedString{}
	}
	if node, ok := computer.element.(*utils.HTMLNode); ok && value.Name == "attr()" {
		s := node.Get(value.String)
		if s == "" {
			return pr.NamedString{}
		}
		return pr.NamedString{String: s}
	} else if value.Name == "string" {
		return pr.NamedString{String: value.String}
	}
	return pr.NamedString{}
}

// Compute the ``tab-size`` property.
func tabSize(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if value.Unit == pr.Scalar {
		return value
	}
	return length(computer, name, value)
}

// Compute the ``transform`` property.
func transforms(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Transforms)
	result := make(pr.Transforms, len(value))
	for index, tr := range value {
		if tr.String == "translate" {
			tr.Dimensions = _lengthOrPercentageTuple2(computer, name, tr.Dimensions)
		}
		result[index] = tr
	}
	return result
}

// Compute the ``vertical-align`` property.
func verticalAlign(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	// Use +/- half an em for super and sub, same as Pango.
	// (See the SUPERSUBRISE constant in pango-markup.c)
	var out pr.Value
	switch value.String {
	case "baseline", "middle", "text-top", "text-bottom", "top", "bottom":
		out.String = value.String
	case "super":
		out.Value = computer.GetFontSize().Value * 0.5
		out.Unit = pr.Scalar
	case "sub":
		out.Value = computer.GetFontSize().Value * -0.5
		out.Unit = pr.Scalar
	default:
		out.Unit = pr.Scalar
		if value.Unit == pr.Percentage {
			height := text.StrutLayout(computer, computer.textContext)[0]
			out.Value = height * value.Value / 100
		} else {
			out.Value = length2(computer, name, value, -1, true).Value
		}
	}
	return out
}

// Compute the ``word-spacing`` property.
func wordSpacing(computer *ComputedStyle, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if value.String == "normal" {
		return pr.Value{}
	}
	return length(computer, name, value)
}
