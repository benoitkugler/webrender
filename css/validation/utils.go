package validation

import (
	"fmt"
	"math"
	"strings"

	"github.com/benoitkugler/webrender/css/counters"
	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils"
)

// Default fallback values used in attr() functions
var attrFallbacks = map[string]pr.CssProperty{
	"string":  pr.String(""),
	"color":   pr.String("currentcolor"),
	"url":     pr.String("about:invalid"),
	"integer": pr.Dimension{Unit: pr.Scalar}.ToValue(),
	"number":  pr.Dimension{Unit: pr.Scalar}.ToValue(),
	"%":       pr.Dimension{Unit: pr.Scalar}.ToValue(),
}

func init() {
	for unitString, unit := range LENGTHUNITS {
		attrFallbacks[unitString] = pr.Dimension{Unit: unit}.ToValue()
	}
	for unitString, unit := range AngleUnits {
		attrFallbacks[unitString] = pr.Dimension{Unit: unit}.ToValue()
	}
}

// Split a list of tokens on commas, ie “parser.LiteralToken(",")“.
//
//	Only "top-level" comma tokens are splitting points, not commas inside a
//	function or blocks.
func SplitOnComma(tokens []Token) [][]Token {
	var parts [][]Token
	var thisPart []Token
	for _, token := range tokens {
		litteral, ok := token.(parser.LiteralToken)
		if ok && litteral.Value == "," {
			parts = append(parts, thisPart)
			thisPart = nil
		} else {
			thisPart = append(thisPart, token)
		}
	}
	parts = append(parts, thisPart)
	return parts
}

// Split a list of tokens on optional commas, ie “LiteralToken(",")“.
func splitOnOptionalComma(tokens []Token) (parts []Token) {
	for _, splitPart := range SplitOnComma(tokens) {
		if len(splitPart) == 0 {
			// Happens when there"s a comma at the beginning, at the end, or
			// when two commas are next to each other.
			return
		}
		parts = append(parts, splitPart...)
	}
	return parts
}

// If “token“ is a keyword, return its name. Otherwise return empty string.
func getCustomIdent(token Token) string {
	if ident, ok := token.(parser.IdentToken); ok {
		return string(ident.Value)
	}
	return ""
}

// Parse an <image> token.
func getImage(_token Token, baseUrl string) (pr.Image, error) {
	token, ok := _token.(parser.FunctionBlock)
	if !ok {
		parsed, _, err := getUrl(_token, baseUrl)
		if err != nil {
			return nil, err
		}
		if parsed.Name == "external" {
			return pr.UrlImage(parsed.String), nil
		}
		return nil, nil
	}
	arguments := SplitOnComma(RemoveWhitespace(*token.Arguments))
	name := token.Name.Lower()
	var err error
	switch name {
	case "linear-gradient", "repeating-linear-gradient":
		direction, colorStops := parseLinearGradientParameters(arguments)
		if len(colorStops) > 0 {
			parsedColorsStop := make([]pr.ColorStop, len(colorStops))
			for index, stop := range colorStops {
				parsedColorsStop[index], err = parseColorStop(stop)
				if err != nil {
					return nil, err
				}
			}
			return pr.LinearGradient{
				Direction:  direction,
				Repeating:  name == "repeating-linear-gradient",
				ColorStops: parsedColorsStop,
			}, nil
		}
	case "radial-gradient", "repeating-radial-gradient":
		result := parseRadialGradientParameters(arguments)
		if result.IsNone() {
			result.shape = "ellipse"
			result.size = pr.GradientSize{Keyword: "farthest-corner"}
			result.position = pr.Center{OriginX: "left", OriginY: "top", Pos: pr.Point{fiftyPercent, fiftyPercent}}
			result.colorStops = arguments
		}
		if len(result.colorStops) > 0 {
			parsedColorsStop := make([]pr.ColorStop, len(result.colorStops))
			for index, stop := range result.colorStops {
				parsedColorsStop[index], err = parseColorStop(stop)
				if err != nil {
					return nil, err
				}
			}
			return pr.RadialGradient{
				ColorStops: parsedColorsStop,
				Shape:      result.shape,
				Size:       result.size,
				Center:     result.position,
				Repeating:  name == "repeating-radial-gradient",
			}, nil
		}
	}
	return nil, nil
}

func parseLinearGradientParameters(arguments [][]Token) (pr.DirectionType, [][]Token) {
	firstArg := arguments[0]
	if len(firstArg) == 1 {
		angle, isNotNone := getAngle(firstArg[0])
		if isNotNone {
			return pr.DirectionType{Angle: angle}, arguments[1:]
		}
	} else {
		var mapped [3]string
		for index, token := range firstArg {
			if index < 3 {
				mapped[index] = getKeyword(token)
			}
		}
		result, isNotNone := directionKeywords[mapped]
		if isNotNone {
			return result, arguments[1:]
		}
	}
	return pr.DirectionType{Angle: math.Pi}, arguments // Default direction is "to bottom"
}

func reverse(a []Token) []Token {
	n := len(a)
	out := make([]Token, n)
	for i := range a {
		out[n-1-i] = a[i]
	}
	return out
}

type radialGradientParameters struct {
	shape      string
	colorStops [][]Token
	position   pr.Center
	size       pr.GradientSize
}

func (r radialGradientParameters) IsNone() bool {
	return r.shape == "" && r.size.IsNone() && r.position.IsNone() && r.colorStops == nil
}

func parseRadialGradientParameters(arguments [][]Token) radialGradientParameters {
	var shape, sizeShape string
	var position pr.Center
	var size pr.GradientSize
	stack := reverse(arguments[0])
	for len(stack) > 0 {
		token := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		keyword := getKeyword(token)
		if keyword == "at" {
			position = parsePosition(reverse(stack))
			if position.IsNone() {
				return radialGradientParameters{}
			}
			break
		} else if (keyword == "circle" || keyword == "ellipse") && shape == "" {
			shape = keyword
		} else if (keyword == "closest-corner" || keyword == "farthest-corner" || keyword == "closest-side" || keyword == "farthest-side") && size.IsNone() {
			size = pr.GradientSize{Keyword: keyword}
		} else {
			if len(stack) > 0 && size.IsNone() {
				length1 := getLength(token, true, true)
				length2 := getLength(stack[len(stack)-1], true, true)
				if !length1.IsNone() && !length2.IsNone() {
					size = pr.GradientSize{Explicit: [2]pr.Dimension{length1, length2}}
					sizeShape = "ellipse"
					i := utils.MaxInt(len(stack)-1, 0)
					stack = stack[:i]
				}
			}
			if size.IsNone() {
				length1 := getLength(token, true, false)
				if !length1.IsNone() {
					size = pr.GradientSize{Explicit: [2]pr.Dimension{length1, length1}}
					sizeShape = "circle"
				}
			}
			if size.IsNone() {
				return radialGradientParameters{}
			}
		}
	}
	if shape == "circle" && sizeShape == "ellipse" {
		return radialGradientParameters{}
	}
	out := radialGradientParameters{
		shape:      shape,
		size:       size,
		position:   position,
		colorStops: arguments[1:],
	}
	if shape == "" {
		if sizeShape != "" {
			out.shape = sizeShape
		} else {
			out.shape = "ellipse"
		}
	}
	if size.IsNone() {
		out.size = pr.GradientSize{Keyword: "farthest-corner"}
	}
	if position.IsNone() {
		out.position = pr.Center{
			OriginX: "left",
			OriginY: "top",
			Pos:     pr.Point{fiftyPercent, fiftyPercent},
		}
	}
	return out
}

func parseColorStop(tokens []Token) (pr.ColorStop, error) {
	switch len(tokens) {
	case 1:
		color := parser.ParseColor(tokens[0])
		if color.Type == parser.ColorCurrentColor {
			return pr.ColorStop{Color: pr.Color(parser.ParseColorString("black"))}, nil
		}
		if !color.IsNone() {
			return pr.ColorStop{Color: pr.Color(color)}, nil
		}
	case 2:
		color := parser.ParseColor(tokens[0])
		position := getLength(tokens[1], true, true)
		if !color.IsNone() && !position.IsNone() {
			return pr.ColorStop{Color: pr.Color(color), Position: position}, nil
		}
	}
	return pr.ColorStop{}, ErrInvalidValue
}

func parseURLToken(value, baseURL string) (url pr.NamedString, attr pr.AttrData, err error) {
	if strings.HasPrefix(value, "#") {
		return pr.NamedString{Name: "internal", String: utils.Unquote(value[1:])}, attr, nil
	} else {
		var joined string
		joined, err = utils.SafeUrljoin(baseURL, value, false)
		if err != nil {
			return
		}
		return pr.NamedString{Name: "external", String: joined}, attr, nil
	}
}

func getUrl(_token Token, baseUrl string) (url pr.NamedString, attr pr.AttrData, err error) {
	switch token := _token.(type) {
	case parser.URLToken:
		return parseURLToken(token.Value, baseUrl)
	case parser.FunctionBlock:
		if token.Name == "attr" {
			attr = checkAttrFunction(token, "url")
			return
		} else if L := len(*token.Arguments); token.Name == "url" && (L == 1 || L == 2) {
			val, _ := (*token.Arguments)[0].(parser.StringToken)
			return parseURLToken(val.Value, baseUrl)
		}
	}
	return
}

func checkStringOrElementFunction(stringOrElement string, token Token) (out pr.ContentProperty) {
	name, args := parseFunction(token)
	if name == "" {
		return
	}
	if name == stringOrElement && (len(args) == 1 || len(args) == 2) {
		customIdent_, ok := args[0].(parser.IdentToken)
		args = args[1:]
		if !ok {
			return
		}
		customIdent := customIdent_.Value

		var ident string
		if len(args) > 0 {
			ident_ := args[0]
			identToken, ok := ident_.(parser.IdentToken)
			val := identToken.Value.Lower()
			if !ok || (val != "first" && val != "start" && val != "last" && val != "first-except") {
				return
			}
			ident = val
		} else {
			ident = "first"
		}
		return pr.ContentProperty{Type: stringOrElement + "()", Content: pr.Strings{string(customIdent), ident}}
	}
	return
}

// CheckVarFunction parse a var(...) token
func CheckVarFunction(token Token) (out pr.VarData) {
	name, args := parseFunction(token)
	if name != "var" || len(args) == 0 {
		return
	}
	ident, ok := args[0].(parser.IdentToken)
	args = args[1:]
	if !ok || !strings.HasPrefix(string(ident.Value), "--") {
		return
	}
	// TODO: we should check authorized tokens
	// https://drafts.csswg.org/css-syntax-3/#typedef-declaration-value
	v := strings.ReplaceAll(string(ident.Value), "-", "_")
	return pr.VarData{
		Name:    v,
		Default: args,
	}
}

// Parse functional notation.
//
//	Return ``(name, args)`` if the given token is a function with comma- or
//	space-separated arguments. Return zero values otherwise.
func parseFunction(functionToken_ Token) (string, []Token) {
	functionToken, ok := functionToken_.(parser.FunctionBlock)
	if !ok {
		return "", nil
	}
	content := RemoveWhitespace(*functionToken.Arguments)
	var (
		arguments []Token
		token     Token
	)
	lastIsComma := false
	for len(content) > 0 {
		token, content = content[0], content[1:]
		lit, ok := token.(parser.LiteralToken)
		isComma := ok && lit.Value == ","
		if lastIsComma && isComma {
			return "", nil
		}
		if isComma {
			lastIsComma = true
		} else {
			lastIsComma = false
			if fn, isFunc := token.(parser.FunctionBlock); isFunc {
				innerName, _ := parseFunction(fn)
				if innerName == "" {
					return "", nil
				}
			}
			arguments = append(arguments, token)
		}
	}
	if lastIsComma {
		return "", nil
	}
	return functionToken.Name.Lower(), arguments
}

func checkAttrFunction(token parser.FunctionBlock, allowedType string) (out pr.AttrData) {
	name, args := parseFunction(token)
	if name == "" {
		return
	}
	la := len(args)
	if name == "attr" && (la == 1 || la == 2 || la == 3) {
		ident, ok := args[0].(parser.IdentToken)
		if !ok {
			return
		}
		attrName := ident.Value
		var (
			typeOrUnit string
			fallback   pr.CssProperty
		)
		if la == 1 {
			typeOrUnit = "string"
		} else {
			ident2, ok := args[1].(parser.IdentToken)
			if !ok {
				return
			}
			typeOrUnit = string(ident2.Value)
			fb, isIN := attrFallbacks[typeOrUnit]
			if !isIN {
				return
			}
			if la < 3 {
				fallback = fb
			} else {
				switch fbValue := args[2].(type) {
				case parser.StringToken:
					fallback = pr.String(fbValue.Value)
				default:
					// TODO: handle other fallback types
					return
				}
			}
		}
		if allowedType == "" || allowedType == typeOrUnit {
			return pr.AttrData{Name: string(attrName), TypeOrUnit: typeOrUnit, Fallback: fallback}
		}
	}
	return
}

// Parse background-position and object-position.
//
// See http://drafts.csswg.org/csswg/css-backgrounds-3/#the-background-position
// https://drafts.csswg.org/css-images-3/#propdef-object-position
func parsePosition(tokens []Token) pr.Center {
	center := parse2dPosition(tokens)
	if !center.IsNone() {
		return pr.Center{
			OriginX: "left",
			OriginY: "top",
			Pos:     center,
		}
	}

	if len(tokens) == 4 {
		keyword1 := getKeyword(tokens[0])
		keyword2 := getKeyword(tokens[2])
		length1 := getLength(tokens[1], true, true)
		length2 := getLength(tokens[3], true, true)
		if !length1.IsNone() && !length2.IsNone() {
			if (keyword1 == "left" || keyword1 == "right") && (keyword2 == "top" || keyword2 == "bottom") {
				return pr.Center{
					OriginX: keyword1,
					OriginY: keyword2,
					Pos:     pr.Point{length1, length2},
				}
			}
			if (keyword2 == "left" || keyword2 == "right") && (keyword1 == "top" || keyword1 == "bottom") {
				return pr.Center{
					OriginX: keyword2,
					OriginY: keyword1,
					Pos:     pr.Point{length2, length1},
				}
			}
		}
	}

	if len(tokens) == 3 {
		length := getLength(tokens[2], true, true)
		var keyword, otherKeyword string
		if !length.IsNone() {
			keyword = getKeyword(tokens[1])
			otherKeyword = getKeyword(tokens[0])
		} else {
			length = getLength(tokens[1], true, true)
			otherKeyword = getKeyword(tokens[2])
			keyword = getKeyword(tokens[0])
		}

		if !length.IsNone() {
			switch otherKeyword {
			case "center":
				switch keyword {
				case "top", "bottom":
					return pr.Center{OriginX: "left", OriginY: keyword, Pos: pr.Point{fiftyPercent, length}}
				case "left", "right":
					return pr.Center{OriginX: keyword, OriginY: "top", Pos: pr.Point{length, fiftyPercent}}
				}
			case "top", "bottom":
				if keyword == "left" || keyword == "right" {
					return pr.Center{OriginX: keyword, OriginY: otherKeyword, Pos: pr.Point{length, ZEROPERCENT}}
				}
			case "left", "right":
				if keyword == "top" || keyword == "bottom" {
					return pr.Center{OriginX: otherKeyword, OriginY: keyword, Pos: pr.Point{ZEROPERCENT, length}}
				}
			}
		}
	}
	return pr.Center{}
}

// Parse a <string> token.
func getString(_token Token) (out pr.ContentProperty) {
	switch token := _token.(type) {
	case parser.StringToken:
		return pr.ContentProperty{Type: "string", Content: pr.String(token.Value)}
	case parser.FunctionBlock:
		switch token.Name {
		case "attr":
			attr := checkAttrFunction(token, "string")
			if attr.IsNone() {
				return
			}
			return pr.ContentProperty{Type: "attr()", Content: attr}
		case "counter", "counters":
			return checkCounterFunction(token)
		case "content":
			return checkContentFunction(token)
		case "string":
			return checkStringOrElementFunction("string", token)
		}
	}
	return
}

func checkCounterFunction(token Token) (prop pr.ContentProperty) {
	name, args := parseFunction(token)
	if name == "" {
		return
	}
	var out pr.Counters
	LA := len(args)
	if (name == "counter" && (LA == 1 || LA == 2)) || (name == "counters" && (LA == 2 || LA == 3)) {
		ident, ok := args[0].(parser.IdentToken)
		args = args[1:]
		if !ok {
			return
		}
		out.Name = string(ident.Value)

		if name == "counters" {
			str, ok := args[0].(parser.StringToken)
			args = args[1:]
			if !ok {
				return
			}
			out.Separator = str.Value
		}

		if len(args) > 0 {
			counterStyle, ok := listStyleType_(args[0:1])
			if !ok {
				return
			}
			out.Style = counterStyle
		} else {
			out.Style.Name = "decimal"
		}

		return pr.ContentProperty{Type: fmt.Sprintf("%s()", name), Content: out}
	}
	return
}

func checkContentFunction(token Token) (out pr.ContentProperty) {
	name, args := parseFunction(token)
	if name == "" {
		return
	}
	if name == "content" {
		if len(args) == 0 {
			return pr.ContentProperty{Type: "content()", Content: pr.String("text")}
		} else if len(args) == 1 {
			ident, ok := args[0].(parser.IdentToken)
			v := ident.Value.Lower()
			if ok && (v == "text" || v == "before" || v == "after" || v == "first-letter" || v == "marker") {
				return pr.ContentProperty{Type: "content()", Content: pr.String(v)}
			}
		}
	}
	return
}

// Parse a <quote> token.
func getQuote(token Token) (pr.Quote, bool) {
	keyword := getKeyword(token)
	out, ok := contentQuoteKeywords[keyword]
	return out, ok
}

// Parse a <target> token.
func getTarget(token Token, baseUrl string) (out pr.ContentProperty, err error) {
	name, args := parseFunction(token)
	if name == "" {
		return
	}
	args = splitOnOptionalComma(args)
	la := len(args)
	if la == 0 {
		return
	}
	switch name {
	case "target-counter":
		if la != 2 && la != 3 {
			return
		}
	case "target-counters":
		if la != 3 && la != 4 {
			return
		}
	case "target-text":
		if la != 1 && la != 2 {
			return
		}
	default:
		return
	}

	var (
		values pr.SContentProps
		value  pr.SContentProp
	)

	link := args[0]
	args = args[1:]
	stringLink := getString(link)
	if stringLink.IsNone() {
		ur, attr, err := getUrl(link, baseUrl)
		if err != nil {
			return out, err
		}
		if !ur.IsNone() {
			value.ContentProperty = pr.ContentProperty{Type: "url", Content: ur}
		} else if !attr.IsNone() {
			value.ContentProperty = pr.ContentProperty{Type: "attr()", Content: attr}
		} else {
			return out, nil
		}
		values = append(values, value)
	} else {
		values = append(values, pr.SContentProp{ContentProperty: stringLink})
	}

	if strings.HasPrefix(name, "target-counter") {
		if len(args) == 0 {
			return
		}

		ident_ := args[0]
		args = args[1:]
		ident, ok := ident_.(parser.IdentToken)
		if !ok {
			return
		}
		values = append(values, pr.SContentProp{String: string(ident.Value)})

		if name == "target-counters" {
			string_ := getString(args[0])
			args = args[1:]
			if string_.IsNone() {
				return
			}
			values = append(values, pr.SContentProp{ContentProperty: string_})
		}

		var counterStyle string
		if len(args) > 0 {
			counterStyle = getKeyword(args[0])
		} else {
			counterStyle = "decimal"
		}
		values = append(values, pr.SContentProp{String: counterStyle})
	} else {
		var content string
		if len(args) > 0 {
			content = getKeyword(args[0])
			if content != "content" && content != "before" && content != "after" && content != "first-letter" {
				return
			}
		} else {
			content = "content"
		}
		values = append(values, pr.SContentProp{String: content})
	}
	return pr.ContentProperty{Type: fmt.Sprintf("%s()", name), Content: values}, nil
}

// Parse <content-list> tokens.
func getContentList(tokens []Token, baseUrl string) (out pr.ContentProperties, err error) {
	// See https://www.w3.org/TR/css-content-3/#typedef-content-list
	parsedTokens := make([]pr.ContentProperty, len(tokens))
	for i, token := range tokens {
		parsedTokens[i], err = getContentListToken(token, baseUrl)
		if err != nil {
			return nil, err
		}
		if parsedTokens[i].IsNone() {
			return nil, nil
		}
	}
	return parsedTokens, nil
}

// Parse one of the <content-list> tokens.
func getContentListToken(token Token, baseUrl string) (pr.ContentProperty, error) {
	// See https://www.w3.org/TR/css-content-3/#typedef-content-list

	// <string>
	string_ := getString(token)
	if !string_.IsNone() {
		return string_, nil
	}

	// <var>
	var_ := CheckVarFunction(token)
	if !var_.IsNone() {
		return pr.ContentProperty{Type: "var()", Content: var_}, nil
	}

	// contents
	if getKeyword(token) == "contents" {
		return pr.ContentProperty{Type: "content()", Content: pr.String("text")}, nil
	}

	// <uri>
	url, attr, err := getUrl(token, baseUrl)
	if err != nil {
		return pr.ContentProperty{}, err
	}
	if !url.IsNone() {
		return pr.ContentProperty{Type: "url", Content: url}, nil
	} else if !attr.IsNone() {
		return pr.ContentProperty{Type: "attr()", Content: attr}, nil
	}

	// <quote>
	quote, ok := getQuote(token)
	if ok {
		return pr.ContentProperty{Type: "quote", Content: quote}, nil
	}

	// <target>
	target, err := getTarget(token, baseUrl)
	if err != nil || !target.IsNone() {
		return target, err
	}

	// <leader>
	name, args := parseFunction(token)
	if name == "" {
		return pr.ContentProperty{}, nil
	}
	if name == "leader" {
		if len(args) != 1 {
			return pr.ContentProperty{}, nil
		}
		arg_ := args[0]
		var str string
		switch arg := arg_.(type) {
		case parser.IdentToken:
			switch arg.Value {
			case "dotted":
				str = "."
			case "solid":
				str = "_"
			case "space":
				str = " "
			default:
				return pr.ContentProperty{}, nil
			}
		case parser.StringToken:
			str = arg.Value
		}
		return pr.ContentProperty{Type: "leader()", Content: pr.Strings{"string", str}}, nil
	} else if name == "element" { // <element>
		return checkStringOrElementFunction("element", token), nil
	}
	return pr.ContentProperty{}, nil
}

func ParseCounterStyleName(tokens []parser.Token, cs counters.CounterStyle) string {
	tokens = RemoveWhitespace(tokens)
	if len(tokens) != 1 {
		return ""
	}

	token := tokens[0]
	if ident, ok := token.(parser.IdentToken); ok {
		if v := ident.Value.Lower(); v == "decimal" || v == "disc" {
			if _, ok := cs[v]; !ok {
				return string(ident.Value)
			}
		} else if ident.Value.Lower() != "none" {
			return string(ident.Value)
		}
	}

	return ""
}
