package validation

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"

	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
)

// Expand shorthands and validate property values.
// See http://www.w3.org/TR/CSS21/propidx.html and various CSS3 modules.

// :copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

const proprietaryPrefix = "-weasy-"

var (
	InvalidValue = errors.New("invalid or unsupported values for a known CSS property")

	LENGTHUNITS = map[string]pr.Unit{"ex": pr.Ex, "em": pr.Em, "ch": pr.Ch, "rem": pr.Rem, "px": pr.Px, "pt": pr.Pt, "pc": pr.Pc, "in": pr.In, "cm": pr.Cm, "mm": pr.Mm, "q": pr.Q}
	AngleUnits  = map[string]pr.Unit{"rad": pr.Rad, "turn": pr.Turn, "deg": pr.Deg, "grad": pr.Grad}
	// keyword -> (open, insert)
	contentQuoteKeywords = map[string]pr.Quote{
		"open-quote":     {Open: true, Insert: true},
		"close-quote":    {Open: false, Insert: true},
		"no-open-quote":  {Open: true, Insert: false},
		"no-close-quote": {Open: false, Insert: false},
	}

	ZEROPERCENT    = pr.Dimension{Value: 0, Unit: pr.Percentage}
	fiftyPercent   = pr.Dimension{Value: 50, Unit: pr.Percentage}
	HUNDREDPERCENT = pr.Dimension{Value: 100, Unit: pr.Percentage}

	backgroundPositionsPercentages = map[string]pr.Dimension{
		"top":    ZEROPERCENT,
		"left":   ZEROPERCENT,
		"center": fiftyPercent,
		"bottom": HUNDREDPERCENT,
		"right":  HUNDREDPERCENT,
	}

	// http://dev.w3.org/csswg/css3-values/#angles
	// 1<unit> is this many radians.
	ANGLETORADIANS = map[pr.Unit]utils.Fl{
		pr.Rad:  1,
		pr.Turn: 2 * math.Pi,
		pr.Deg:  math.Pi / 180,
		pr.Grad: math.Pi / 200,
	}

	// http://dev.w3.org/csswg/css-values/#resolution
	RESOLUTIONTODPPX = map[string]utils.Fl{
		"dppx": 1,
		"dpi":  utils.Fl(1 / pr.LengthsToPixels[pr.In]),
		"dpcm": utils.Fl(1 / pr.LengthsToPixels[pr.Cm]),
	}

	couplesLigatures = [][]string{
		{"common-ligatures", "no-common-ligatures"},
		{"historical-ligatures", "no-historical-ligatures"},
		{"discretionary-ligatures", "no-discretionary-ligatures"},
		{"contextual", "no-contextual"},
	}
	couplesNumeric = [][]string{
		{"lining-nums", "oldstyle-nums"},
		{"proportional-nums", "tabular-nums"},
		{"diagonal-fractions", "stacked-fractions"},
		{"ordinal"},
		{"slashed-zero"},
	}

	couplesEastAsian = [][]string{
		{"jis78", "jis83", "jis90", "jis04", "simplified", "traditional"},
		{"full-width", "proportional-width"},
		{"ruby"},
	}

	allLigaturesValues = utils.Set{}
	allNumericValues   = utils.Set{}
	allEastAsianValues = utils.Set{}

	// yes/no validators for non-shorthand properties
	// Maps property names to functions taking a property name and a value list,
	// returning a value or None for invalid.
	// Also transform values: keyword and URLs are returned as strings.
	// For properties that take a single value, that value is returned by itself
	// instead of a list.
	validators = map[string]validator{
		"background-attachment":      backgroundAttachment,
		"background-color":           otherColors,
		"border-top-color":           otherColors,
		"border-right-color":         otherColors,
		"border-bottom-color":        otherColors,
		"border-left-color":          otherColors,
		"column-rule-color":          otherColors,
		"text-decoration-color":      otherColors,
		"outline-color":              outlineColor,
		"border-collapse":            borderCollapse,
		"empty-cells":                emptyCells,
		"transform-origin":           transformOrigin,
		"object-position":            objectPosition,
		"background-position":        backgroundPosition,
		"background-repeat":          backgroundRepeat,
		"background-size":            backgroundSize,
		"background-clip":            box,
		"background-origin":          box,
		"border-spacing":             borderSpacing,
		"border-top-right-radius":    borderCornerRadius,
		"border-bottom-right-radius": borderCornerRadius,
		"border-bottom-left-radius":  borderCornerRadius,
		"border-top-left-radius":     borderCornerRadius,
		"border-top-style":           borderStyle,
		"border-right-style":         borderStyle,
		"border-left-style":          borderStyle,
		"border-bottom-style":        borderStyle,
		"column-rule-style":          borderStyle,
		"break-before":               breakBeforeAfter,
		"break-after":                breakBeforeAfter,
		"break-inside":               breakInside,
		"box-decoration-break":       boxDecorationBreak,
		"margin-break":               marginBreak,
		"page":                       page,
		"bleed-left":                 bleed,
		"bleed-right":                bleed,
		"bleed-top":                  bleed,
		"bleed-bottom":               bleed,
		"marks":                      marks,
		"outline-style":              outlineStyle,
		"border-top-width":           borderWidth,
		"border-right-width":         borderWidth,
		"border-left-width":          borderWidth,
		"border-bottom-width":        borderWidth,
		"column-rule-width":          borderWidth,
		"outline-width":              borderWidth,
		"column-width":               columnWidth,
		"column-span":                columnSpan,
		"box-sizing":                 boxSizing,
		"caption-side":               captionSide,
		"clear":                      clear,
		"clip":                       clip,
		"top":                        lengthPercOrAuto,
		"right":                      lengthPercOrAuto,
		"left":                       lengthPercOrAuto,
		"bottom":                     lengthPercOrAuto,
		"margin-top":                 lengthPercOrAuto,
		"margin-right":               lengthPercOrAuto,
		"margin-bottom":              lengthPercOrAuto,
		"margin-left":                lengthPercOrAuto,
		"height":                     widthHeight,
		"width":                      widthHeight,
		"column-gap":                 columnGap,
		"column-fill":                columnFill,
		"direction":                  direction,
		"display":                    display,
		"float":                      float,
		"font-family":                fontFamily,
		"font-kerning":               fontKerning,
		"font-language-override":     fontLanguageOverride,
		"font-variant-ligatures":     fontVariantLigatures,
		"font-variant-position":      fontVariantPosition,
		"font-variant-caps":          fontVariantCaps,
		"font-variant-numeric":       fontVariantNumeric,
		"font-feature-settings":      fontFeatureSettings,
		"font-variant-alternates":    fontVariantAlternates,
		"font-variant-east-asian":    fontVariantEastAsian,
		"font-style":                 fontStyle,
		"font-stretch":               fontStretch,
		"font-weight":                fontWeight,
		"footnote-display":           footnoteDisplay,
		"footnote-policy":            footnotePolicy,
		"image-resolution":           imageResolution,
		"letter-spacing":             spacing,
		"word-spacing":               spacing,
		"line-height":                lineHeight,
		"list-style-position":        listStylePosition,
		"list-style-type":            listStyleType,
		"padding-top":                lengthOrPercentage,
		"padding-right":              lengthOrPercentage,
		"padding-bottom":             lengthOrPercentage,
		"padding-left":               lengthOrPercentage,
		"min-width":                  minWidthHeight,
		"min-height":                 minWidthHeight,
		"max-width":                  maxWidthHeight,
		"max-height":                 maxWidthHeight,
		"opacity":                    opacity,
		"z-index":                    zIndex,
		"orphans":                    orphansWidows,
		"widows":                     orphansWidows,
		"column-count":               columnCount,
		"overflow":                   overflow,
		"position":                   position,
		"quotes":                     quotes,
		"table-layout":               tableLayout,
		"text-align-all":             textAlignAll,
		"text-align-last":            textAlignLast,
		"text-decoration-line":       textDecorationLine,
		"text-decoration-style":      textDecorationStyle,
		"text-indent":                textIndent,
		"text-transform":             textTransform,
		"vertical-align":             verticalAlign,
		"visibility":                 visibility,
		"white-space":                whiteSpace,
		"overflow-wrap":              overflowWrap,
		"image-rendering":            imageRendering,
		"size":                       size,
		"tab-size":                   tabSize,
		"hyphens":                    hyphens,
		"hyphenate-character":        hyphenateCharacter,
		"hyphenate-limit-zone":       hyphenateLimitZone,
		"hyphenate-limit-chars":      hyphenateLimitChars,
		"lang":                       lang,
		"bookmark-level":             bookmarkLevel,
		"bookmark-state":             bookmarkState,
		"object-fit":                 objectFit,
		"text-overflow":              textOverflow,
		"flex-basis":                 flexBasis,
		"flex-direction":             flexDirection,
		"flex-grow":                  flexGrowShrink,
		"flex-shrink":                flexGrowShrink,
		"order":                      order,
		"flex-wrap":                  flexWrap,
		"justify-content":            justifyContent,
		"align-items":                alignItems,
		"align-self":                 alignSelf,
		"align-content":              alignContent,
		"anchor":                     anchor,
		"block-ellipsis":             blockEllipsis,
		"continue":                   continue_,
		"max-lines":                  maxLines,
	}
	validatorsError = map[string]validatorError{
		"background-image":  backgroundImage,
		"list-style-image":  listStyleImage,
		"content":           content,
		"counter-increment": counterIncrement,
		"counter-reset":     counterReset,
		"counter-set":       counterReset,
		"font-size":         fontSize,
		"bookmark-label":    bookmarkLabel,
		"transform":         transform,
		"string-set":        stringSet,
		"link":              link,
	}

	// regroup the two cases (with error or without error)
	allValidators = utils.NewSet("color") // special case because of inherited

	proprietary = utils.NewSet(
		"anchor",
		"link",
		"lang",
	)
	unstable = utils.NewSet(
		"transform-origin",
		"size",
		"hyphens",
		"hyphenate-character",
		"hyphenate-limit-zone",
		"hyphenate-limit-chars",
		"bookmark-label",
		"bookmark-level",
		"bookmark-state",
		"string-set",
		"column-rule-color",
		"column-rule-style",
		"column-rule-width",
		"column-width",
		"column-span",
		"column-gap",
		"column-fill",
		"column-count",
		"bleed-left",
		"bleed-right",
		"bleed-top",
		"bleed-bottom",
		"marks",
		"continue",
		"max-lines",
	)
	multiValProperties = utils.NewSet(
		"content",
	)
)

func init() {
	for _, couple := range couplesLigatures {
		for _, cc := range couple {
			allLigaturesValues[cc] = utils.Has
		}
	}
	for _, couple := range couplesNumeric {
		for _, cc := range couple {
			allNumericValues[cc] = utils.Has
		}
	}
	for _, couple := range couplesEastAsian {
		for _, cc := range couple {
			allEastAsianValues[cc] = utils.Has
		}
	}
	for name := range validators {
		allValidators[name] = utils.Has
	}
	for name := range validatorsError {
		allValidators[name] = utils.Has
	}
}

type Token = parser.Token

type (
	validator      func(tokens []Token, baseUrl string) pr.CssProperty // dont support var(), attr()
	validatorError func(tokens []Token, baseUrl string) (pr.CssProperty, error)
	expander       func(baseUrl, name string, tokens []Token) (pr.NamedProperties, error)
)

type ValidatedProperty struct {
	Name      string
	Value     pr.ValidatedProperty
	Important bool
}

// Validate validate one property.
func Validate(name string, tokens []Token, baseUrl string) (out pr.CascadedProperty, err error) {
	if name == "color" { // special case to handle inherit inherit
		return color(tokens, ""), nil
	}

	var value pr.CssProperty
	if function := validators[name]; function != nil {
		value = function(tokens, baseUrl)
	} else if functionE := validatorsError[name]; functionE != nil {
		value, err = functionE(tokens, baseUrl)
	}
	return pr.AsCascaded(value), err
}

// Default validator for non-shorthand pr.
// required = false
func validateNonShorthand(baseUrl, name string, tokens []parser.Token, required bool) (out pr.NamedProperty, err error) {
	if strings.HasPrefix(name, "--") {
		return pr.NamedProperty{
			Name:     name,
			Property: pr.AsValidated(pr.RawTokens(tokens)),
		}, nil
	}

	if !required && !pr.KnownProperties.Has(name) {
		hyphensName := strings.ReplaceAll(name, "_", "-")
		if pr.KnownProperties.Has(hyphensName) {
			return out, fmt.Errorf("did you mean %s?", hyphensName)
		} else {
			return out, errors.New("unknown property")
		}
	}

	if _, isIn := allValidators[name]; !required && !isIn {
		return out, fmt.Errorf("property %s not supported yet", name)
	}

	if _, isIn := multiValProperties[name]; !isIn {
		for _, token := range tokens {
			var_ := CheckVarFunction(token)
			if !var_.IsNone() {
				return pr.NamedProperty{Name: name, Property: pr.AsValidated(var_)}, nil
			}
		}
	}

	var value pr.CascadedProperty
	keyword := getSingleKeyword(tokens)
	if keyword == "initial" || keyword == "inherit" {
		value = defaultFromString(keyword).AsCascaded()
	} else {
		value, err = Validate(name, tokens, baseUrl)
		if err != nil {
			return out, err
		}
		if value.IsNone() {
			return out, errors.New("invalid value (nil function return)")
		}
	}

	return pr.NamedProperty{Name: name, Property: value.AsValidated()}, nil
}

func defaultValidateShorthand(baseUrl, name string, tokens []parser.Token) (pr.NamedProperties, error) {
	np, err := validateNonShorthand(baseUrl, name, tokens, false)
	return pr.NamedProperties{np}, err
}

// Expand shorthand properties and filter unsupported properties and values.
// Log a warning for every ignored declaration.
// Return a iterable of ``(name, value, important)`` tuples.
//
func PreprocessDeclarations(baseUrl string, declarations []Token) []ValidatedProperty {
	var out []ValidatedProperty
	for _, _declaration := range declarations {
		if errToken, ok := _declaration.(parser.ParseError); ok {
			logger.WarningLogger.Printf("Error: %s \n", errToken.Message)
		}

		declaration, ok := _declaration.(parser.Declaration)
		if !ok {
			continue
		}

		name := string(declaration.Name)
		if !strings.HasPrefix(name, "--") { // check for non variable, case insensitive
			name = declaration.Name.Lower()
		}

		validationError := func(reason string) {
			logger.WarningLogger.Printf("Ignored `%s:%s` , %s. \n", declaration.Name, parser.Serialize(declaration.Value), reason)
		}

		if _, in := pr.NotPrintMedia[name]; in {
			validationError("the property does not apply for the print media")
			continue
		}

		if strings.HasPrefix(name, proprietaryPrefix) {
			unprefixedName := strings.TrimPrefix(name, proprietaryPrefix)
			if _, in := proprietary[unprefixedName]; in {
				name = unprefixedName
			} else if _, in := unstable[unprefixedName]; in {
				logger.WarningLogger.Printf("Deprecated `%s:%s`, prefixes on unstable attributes are deprecated, use `%s` instead. \n",
					declaration.Name, parser.Serialize(declaration.Value), unprefixedName)
				name = unprefixedName
			} else {
				logger.WarningLogger.Printf("Ignored `%s:%s`,prefix on this attribute is not supported, use `%s` instead. \n",
					declaration.Name, parser.Serialize(declaration.Value), unprefixedName)
				continue
			}
		}

		tokens := RemoveWhitespace(declaration.Value)

		expander_, in := expanders[name]
		if !in {
			expander_ = defaultValidateShorthand
		}
		result, err := expander_(baseUrl, name, tokens)
		if err != nil {
			validationError(err.Error())
			continue
		}

		important := declaration.Important

		for _, np := range result {
			out = append(out, ValidatedProperty{
				Name:      strings.ReplaceAll(np.Name, "-", "_"),
				Value:     np.Property,
				Important: important,
			})
		}
	}
	return out
}

// If `token` is a keyword, return its name.
// Otherwise return empty string.
func getKeyword(token Token) string {
	if ident, ok := token.(parser.IdentToken); ok {
		return ident.Value.Lower()
	}
	return ""
}

// If `tokens` is a 1-element list of keywords, return its name.
// Otherwise return empty string.
func getSingleKeyword(tokens []Token) string {
	if len(tokens) == 1 {
		return getKeyword(tokens[0])
	}
	return ""
}

// negative  = true, percentage = false
func getLength(_token Token, negative, percentage bool) pr.Dimension {
	switch token := _token.(type) {
	case parser.PercentageToken:
		if percentage && (negative || token.Value >= 0) {
			return pr.Dimension{Value: pr.Float(token.Value), Unit: pr.Percentage}
		}
	case parser.DimensionToken:
		unit, isKnown := LENGTHUNITS[string(token.Unit)]
		if isKnown && (negative || token.Value >= 0) {
			return pr.Dimension{Value: pr.Float(token.Value), Unit: unit}
		}
	case parser.NumberToken:
		if token.Value == 0 {
			return pr.Dimension{Unit: pr.Scalar}
		}
	}
	return pr.Dimension{}
}

// Return the value in radians of an <angle> token, or None.
func getAngle(token Token) (utils.Fl, bool) {
	if dim, ok := token.(parser.DimensionToken); ok {
		unit, in := AngleUnits[string(dim.Unit)]
		if in {
			return dim.Value * ANGLETORADIANS[unit], true
		}
	}
	return 0, false
}

// Return the value in dppx of a <resolution> token, or false.
func getResolution(token Token) (utils.Fl, bool) {
	if dim, ok := token.(parser.DimensionToken); ok {
		factor, in := RESOLUTIONTODPPX[string(dim.Unit)]
		if in {
			return dim.Value * factor, true
		}
	}
	return 0, false
}

//@validator()
//@commaSeparatedList
//@singleKeyword
// ``background-attachment`` property validation.
func _backgroundAttachment(tokens []Token) string {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "scroll", "fixed", "local":
		return keyword
	default:
		return ""
	}
}

func backgroundAttachment(tokens []Token, _ string) pr.CssProperty {
	var out pr.Strings
	for _, part := range SplitOnComma(tokens) {
		part = RemoveWhitespace(part)
		result := _backgroundAttachment(part)
		if result == "" {
			return nil
		}
		out = append(out, result)
	}
	return out
}

//@validator("background-color")
//@validator("border-top-color")
//@validator("border-right-color")
//@validator("border-bottom-color")
//@validator("border-left-color")
//@validator("column-rule-color")
//@singleToken
func otherColors(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		c := parser.ParseColor(tokens[0])
		if !c.IsNone() {
			return pr.Color(c)
		}
	}
	return nil
}

//@validator()
//@singleToken
func outlineColor(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		token := tokens[0]
		if getKeyword(token) == "invert" {
			return pr.Color{Type: parser.ColorCurrentColor}
		} else {
			return pr.Color(parser.ParseColor(token))
		}
	}
	return nil
}

//@validator()
//@singleKeyword
func borderCollapse(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "separate", "collapse":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``empty-cells`` property validation.
func emptyCells(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "show", "hide":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator("color")
//@singleToken
// ``*-color`` && ``color`` properties validation.
func color(tokens []Token, _ string) pr.CascadedProperty {
	if len(tokens) != 1 {
		return pr.CascadedProperty{}
	}
	token := tokens[0]
	result := parser.ParseColor(token)
	if result.Type == parser.ColorCurrentColor {
		return pr.Inherit.AsCascaded()
	} else {
		return pr.AsCascaded(pr.Color(result))
	}
}

// @validator("background-image", wantsBaseUrl=true)
// @commaSeparatedList
// @singleToken
func _backgroundImage(tokens []Token, baseUrl string) (pr.Image, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]

	if _, ok := token.(parser.FunctionBlock); !ok {
		if getKeyword(token) == "none" {
			return pr.NoneImage{}, nil
		}
	}
	return getImage(token, baseUrl)
}

func backgroundImage(tokens []Token, baseUrl string) (pr.CssProperty, error) {
	var out pr.Images
	for _, part := range SplitOnComma(tokens) {
		part = RemoveWhitespace(part)
		result, err := _backgroundImage(part, baseUrl)
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, nil
		}
		out = append(out, result)

	}
	return out, nil
}

var directionKeywords = map[[3]string]pr.DirectionType{
	// ("angle", radians)  0 upwards, then clockwise
	{"to", "top", ""}:    {Angle: 0},
	{"to", "right", ""}:  {Angle: pr.Fl(math.Pi) / 2},
	{"to", "bottom", ""}: {Angle: math.Pi},
	{"to", "left", ""}:   {Angle: math.Pi * 3 / 2},
	// ("corner", keyword)
	{"to", "top", "left"}:     {Corner: "top_left"},
	{"to", "left", "top"}:     {Corner: "top_left"},
	{"to", "top", "right"}:    {Corner: "top_right"},
	{"to", "right", "top"}:    {Corner: "top_right"},
	{"to", "bottom", "left"}:  {Corner: "bottom_left"},
	{"to", "left", "bottom"}:  {Corner: "bottom_left"},
	{"to", "bottom", "right"}: {Corner: "bottom_right"},
	{"to", "right", "bottom"}: {Corner: "bottom_right"},
}

//@validator("list-style-image", wantsBaseUrl=true)
//@singleToken
// ``list-style-image`` property validation.
func listStyleImage(tokens []Token, baseUrl string) (pr.CssProperty, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]

	if token.Type() != "function" {
		if getKeyword(token) == "none" {
			return pr.NoneImage{}, nil
		}
		parsedUrl, _, err := getUrl(token, baseUrl)
		if err != nil {
			return nil, err
		}
		if parsedUrl.Name == "external" {
			return pr.UrlImage(parsedUrl.String), nil
		}
	}
	return nil, nil
}

var centerKeywordFakeToken = parser.IdentToken{Value: "center"}

//@validator(unstable=true)
func transformOrigin(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 3 {
		// Ignore third parameter as 3D transforms are ignored.
		tokens = tokens[:2]
	}
	return parse2dPosition(tokens)
}

//@validator()
//@commaSeparatedList
// ``background-position`` and ``object-position`` property validation.
// See http://dev.w3.org/csswg/css3-background/#the-background-position
func backgroundPosition(tokens []Token, _ string) pr.CssProperty {
	out := centers(tokens)

	if len(out) == 0 {
		return nil
	}

	return out
}

func centers(tokens []Token) pr.Centers {
	var out pr.Centers
	for _, part := range SplitOnComma(tokens) {
		result := parsePosition(RemoveWhitespace(part))
		if result.IsNone() {
			return nil
		}
		out = append(out, result)
	}
	return out
}

// object-position property validation
func objectPosition(tokens []Token, _ string) pr.CssProperty {
	out := centers(tokens)

	if len(out) == 0 {
		return nil
	}
	return out[0]
}

// Common syntax of background-position and transform-origin.
func parse2dPosition(tokens []Token) pr.Point {
	if len(tokens) == 1 {
		tokens = []Token{tokens[0], centerKeywordFakeToken}
	} else if len(tokens) != 2 {
		return pr.Point{}
	}

	token1, token2 := tokens[0], tokens[1]
	length1 := getLength(token1, true, true)
	length2 := getLength(token2, true, true)
	if !length1.IsNone() && !length2.IsNone() {
		return pr.Point{length1, length2}
	}
	keyword1, keyword2 := getKeyword(token1), getKeyword(token2)
	if !length1.IsNone() && (keyword2 == "top" || keyword2 == "center" || keyword2 == "bottom") {
		return pr.Point{length1, backgroundPositionsPercentages[keyword2]}
	} else if !length2.IsNone() && (keyword1 == "left" || keyword1 == "center" || keyword1 == "right") {
		return pr.Point{backgroundPositionsPercentages[keyword1], length2}
	} else if (keyword1 == "left" || keyword1 == "center" || keyword1 == "right") &&
		(keyword2 == "top" || keyword2 == "center" || keyword2 == "bottom") {
		return pr.Point{backgroundPositionsPercentages[keyword1], backgroundPositionsPercentages[keyword2]}
	} else if (keyword1 == "top" || keyword1 == "center" || keyword1 == "bottom") &&
		(keyword2 == "left" || keyword2 == "center" || keyword2 == "right") {
		// Swap tokens. They need to be in (horizontal, vertical) order.
		return pr.Point{backgroundPositionsPercentages[keyword2], backgroundPositionsPercentages[keyword1]}
	}
	return pr.Point{}
}

//@validator()
//@commaSeparatedList
// ``background-repeat`` property validation.
func _backgroundRepeat(tokens []Token) [2]string {
	keywords := make([]string, len(tokens))
	for index, token := range tokens {
		keywords[index] = getKeyword(token)
	}

	switch len(keywords) {
	case 1:
		switch keywords[0] {
		case "repeat-x":
			return [2]string{"repeat", "no-repeat"}
		case "repeat-y":
			return [2]string{"no-repeat", "repeat"}
		case "no-repeat", "repeat", "space", "round":
			return [2]string{keywords[0], keywords[0]}
		}
	case 2:
		for _, k := range keywords {
			if !(k == "no-repeat" || k == "repeat" || k == "space" || k == "round") {
				return [2]string{}
			}
		}
		// OK
		return [2]string{keywords[0], keywords[1]}
	}
	return [2]string{}
}

func backgroundRepeat(tokens []Token, _ string) pr.CssProperty {
	var out pr.Repeats
	for _, part := range SplitOnComma(tokens) {
		result := _backgroundRepeat(RemoveWhitespace(part))
		if result == [2]string{} {
			return nil
		}
		out = append(out, result)
	}
	return out
}

//@validator()
//@commaSeparatedList
// Validation for ``background-size``.
func _backgroundSize(tokens []Token) pr.Size {
	switch len(tokens) {
	case 1:
		token := tokens[0]
		keyword := getKeyword(token)
		switch keyword {
		case "contain", "cover":
			return pr.Size{String: keyword}
		case "auto":
			return pr.Size{Width: pr.SToV("auto"), Height: pr.SToV("auto")}
		}
		length := getLength(token, false, true)
		if !length.IsNone() {
			return pr.Size{Width: pr.Value{Dimension: length}, Height: pr.SToV("auto")}
		}
	case 2:
		var out pr.Size
		lengthW := getLength(tokens[0], false, true)
		lengthH := getLength(tokens[1], false, true)
		if !lengthW.IsNone() {
			out.Width = pr.Value{Dimension: lengthW}
		} else if getKeyword(tokens[0]) == "auto" {
			out.Width = pr.SToV("auto")
		} else {
			return pr.Size{}
		}
		if !lengthH.IsNone() {
			out.Height = pr.Value{Dimension: lengthH}
		} else if getKeyword(tokens[1]) == "auto" {
			out.Height = pr.SToV("auto")
		} else {
			return pr.Size{}
		}
		return out
	}
	return pr.Size{}
}

func backgroundSize(tokens []Token, _ string) pr.CssProperty {
	var out pr.Sizes
	for _, part := range SplitOnComma(tokens) {
		result := _backgroundSize(RemoveWhitespace(part))
		if (result == pr.Size{}) {
			return nil
		}
		out = append(out, result)
	}
	return out
}

//@validator("background-clip")
//@validator("background-origin")
//@commaSeparatedList
//@singleKeyword
// Validation for the ``<box>`` type used in ``background-clip``
//     and ``background-origin``.
func _box(tokens []Token) string {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "border-box", "padding-box", "content-box":
		return keyword
	default:
		return ""
	}
}

func box(tokens []Token, _ string) pr.CssProperty {
	var out pr.Strings
	for _, part := range SplitOnComma(tokens) {
		result := _box(RemoveWhitespace(part))
		if result == "" {
			return nil
		}
		out = append(out, result)
	}
	return out
}

func borderDims(tokens []Token, negative bool) pr.CssProperty {
	lengths := make([]pr.Dimension, len(tokens))
	allLengths := true
	for index, token := range tokens {
		lengths[index] = getLength(token, negative, false)
		allLengths = allLengths && !lengths[index].IsNone()
	}
	if allLengths {
		if len(lengths) == 1 {
			return pr.Point{lengths[0], lengths[0]}
		} else if len(lengths) == 2 {
			return pr.Point{lengths[0], lengths[1]}
		}
	}
	return nil
}

//@validator()
// Validator for the `border-spacing` property.
func borderSpacing(tokens []Token, _ string) pr.CssProperty {
	return borderDims(tokens, true)
}

//@validator("border-top-right-radius")
//@validator("border-bottom-right-radius")
//@validator("border-bottom-left-radius")
//@validator("border-top-left-radius")
// Validator for the `border-*-radius` pr.
func borderCornerRadius(tokens []Token, _ string) pr.CssProperty {
	return borderDims(tokens, false)
}

//@validator("border-top-style")
//@validator("border-right-style")
//@validator("border-left-style")
//@validator("border-bottom-style")
//@validator("column-rule-style")
//@singleKeyword
// ``border-*-style`` properties validation.
func borderStyle(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "none", "hidden", "dotted", "dashed", "double",
		"inset", "outset", "groove", "ridge", "solid":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator("break-before")
//@validator("break-after")
//@singleKeyword
// ``break-before`` && ``break-after`` properties validation.
func breakBeforeAfter(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	// "always" is defined as an alias to "page" := range multi-column
	// https://www.w3.org/TR/css3-multicol/#column-breaks
	switch keyword {
	case "auto", "avoid", "avoid-page", "page", "left", "right",
		"recto", "verso", "avoid-column", "column", "always":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``break-inside`` property validation.
func breakInside(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "avoid", "avoid-page", "avoid-column":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``box-decoration-break`` property validation.
func boxDecorationBreak(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "slice", "clone":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``margin-break`` property validation.
func marginBreak(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "keep", "discard":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator(unstable=true)
//@singleToken
// ``page`` property validation.
func page(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if ident, ok := token.(parser.IdentToken); ok {
		if ident.Value.Lower() == "auto" {
			return pr.Page{String: "auto", Valid: true}
		}
		return pr.Page{String: string(ident.Value), Valid: true}
	}
	return nil
}

//@validator("bleed-left")
//@validator("bleed-right")
//@validator("bleed-top")
//@validator("bleed-bottom")
//@singleToken
// ``bleed`` property validation.
func bleed(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "auto" {
		return pr.Value{String: "auto"}
	} else {
		return pr.Value{Dimension: getLength(token, true, false)}
	}
}

//@validator()
// ``marks`` property validation.
func marks(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 2 {
		keywords := [2]string{getKeyword(tokens[0]), getKeyword(tokens[1])}
		if keywords == [2]string{"crop", "cross"} || keywords == [2]string{"cross", "crop"} {
			return pr.Marks{Crop: true, Cross: true}
		}
	} else if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		switch keyword {
		case "crop":
			return pr.Marks{Crop: true}
		case "cross":
			return pr.Marks{Cross: true}
		case "none":
			return pr.Marks{}
		}
	}
	return nil
}

//@validator("outline-style")
//@singleKeyword
// ``outline-style`` properties validation.
func outlineStyle(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "none", "dotted", "dashed", "double", "inset",
		"outset", "groove", "ridge", "solid":
		return pr.String(keyword)
	default:
		return nil
	}
}

// //@validator("border-top-width")
// //@validator("border-right-width")
// //@validator("border-left-width")
// //@validator("border-bottom-width")
// //@validator("column-rule-width")
// //@validator("outline-width")
// //@singleToken
// Border, column rule && outline widths properties validation.
func borderWidth(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, false)
	if !length.IsNone() {
		return pr.Value{Dimension: length}
	}
	keyword := getKeyword(token)
	if keyword == "thin" || keyword == "medium" || keyword == "thick" {
		return pr.Value{String: keyword}
	}
	return nil
}

// //@validator()
// //@singleToken
// ``column-width`` property validation.
func columnWidth(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, false)
	if !length.IsNone() {
		return pr.Value{Dimension: length}
	}
	keyword := getKeyword(token)
	if keyword == "auto" {
		return pr.Value{String: keyword}
	}
	return nil
}

// //@validator()
// //@singleKeyword
// ``column-span`` property validation.
func columnSpan(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "all", "none":
		return pr.String(keyword)
	default:
		return nil
	}
}

// //@validator()
// //@singleKeyword
// Validation for the ``box-sizing`` property from css3-ui
func boxSizing(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "padding-box", "border-box", "content-box":
		return pr.String(keyword)
	default:
		return nil
	}
}

// //@validator()
// //@singleKeyword
// ``caption-side`` properties validation.
func captionSide(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "top", "bottom":
		return pr.String(keyword)
	default:
		return nil
	}
}

// //@validator()
// //@singleKeyword
// ``clear`` property validation.
func clear(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "left", "right", "both", "none":
		return pr.String(keyword)
	default:
		return nil
	}
}

// //@validator()
// //@singleToken
// Validation for the ``clip`` property.
func clip(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	name, args := parseFunction(token)
	if name != "" {
		if name == "rect" && len(args) == 4 {
			var values pr.Values
			for _, arg := range args {
				if getKeyword(arg) == "auto" {
					values = append(values, pr.Value{String: "auto"})
				} else {
					length := getLength(arg, true, false)
					if !length.IsNone() {
						values = append(values, pr.Value{Dimension: length})
					}
				}
			}
			if len(values) == 4 {
				return values
			}
		}
	}
	if getKeyword(token) == "auto" {
		return pr.Values{}
	}
	return nil
}

// //@validator(wantsBaseUrl=true)
// ``content`` property validation.
func content(tokens []Token, baseUrl string) (pr.CssProperty, error) {
	var token Token
	for len(tokens) > 0 {
		if len(tokens) >= 2 {
			if lit, ok := tokens[1].(parser.LiteralToken); ok && lit.Value == "," {
				token, tokens = tokens[0], tokens[2:]
				im, err := getImage(token, baseUrl)
				if err != nil {
					return nil, err
				}
				if im == nil {
					ur, atr, err := getUrl(token, baseUrl)
					if err != nil || (ur.IsNone() && atr.IsNone()) {
						return nil, err
					}
				}
			} else {
				break
			}
		} else {
			break
		}
	}

	if len(tokens) == 0 {
		return nil, nil
	}
	if len(tokens) >= 3 {
		lit, ok := tokens[len(tokens)-2].(parser.LiteralToken)
		if tokens[len(tokens)-1].Type() == parser.StringTokenT && ok && lit.Value == "/" {
			// Ignore text for speech
			tokens = tokens[:len(tokens)-2]
		}
	}

	keyword := getSingleKeyword(tokens)
	if keyword == "normal" || keyword == "none" {
		return pr.SContent{String: keyword}, nil
	}
	l, err := getContentList(tokens, baseUrl)
	if l == nil || err != nil {
		return nil, err
	}
	return pr.SContent{Contents: l}, nil
}

// //@validator()
// ``counter-increment`` property validation.
func counterIncrement(tokens []Token, _ string) (pr.CssProperty, error) {
	ci, err := counter(tokens, 1)
	if err != nil || ci == nil {
		return nil, err
	}
	return pr.SIntStrings{Values: ci}, nil
}

// ``counter-reset`` property validation.
// ``counter-set`` property validation.
func counterReset(tokens []Token, _ string) (pr.CssProperty, error) {
	ci, err := counter(tokens, 0)
	if err != nil || ci == nil {
		return nil, err
	}
	return pr.SIntStrings{Values: ci}, err
}

// ``counter-increment`` && ``counter-reset`` properties validation.
func counter(tokens []Token, defaultInteger int) ([]pr.IntString, error) {
	if getSingleKeyword(tokens) == "none" {
		return []pr.IntString{}, nil
	}
	if len(tokens) == 0 {
		return nil, errors.New("got an empty token list")
	}
	var (
		results []pr.IntString
		integer int
	)
	iter := parser.NewTokenIterator(tokens)
	token := iter.Next()
	for token != nil {
		ident, ok := token.(parser.IdentToken)
		if !ok {
			return nil, nil // expected a keyword here
		}
		counterName := ident.Value
		if counterName == "none" || counterName == "initial" || counterName == "inherit" {
			return nil, fmt.Errorf("Invalid counter name: %s", counterName)
		}
		token = iter.Next()
		if number, ok := token.(parser.NumberToken); ok && number.IsInteger { // implies token != nil
			// Found an integer. Use it and get the next token
			integer = number.IntValue()
			token = iter.Next()
		} else {
			// Not an integer. Might be the next counter name.
			// Keep `token` for the next loop iteration.
			integer = defaultInteger
		}
		results = append(results, pr.IntString{String: string(counterName), Int: integer})
	}
	return results, nil
}

// //@validator("top")
// //@validator("right")
// //@validator("left")
// //@validator("bottom")
// //@validator("margin-top")
// //@validator("margin-right")
// //@validator("margin-bottom")
// //@validator("margin-left")
// //@singleToken
// ``margin-*`` properties validation.
func lengthPercOrAuto(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, true, true)
	if !length.IsNone() {
		return pr.Value{Dimension: length}
	}
	if getKeyword(token) == "auto" {
		return pr.Value{String: "auto"}
	}
	return nil
}

// //@validator("height")
// //@validator("width")
// //@singleToken
// Validation for the ``width`` && ``height`` pr.
func widthHeight(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, true)
	if !length.IsNone() {
		return pr.Value{Dimension: length}
	}
	if getKeyword(token) == "auto" {
		return pr.Value{String: "auto"}
	}
	return nil
}

// //@validator()
// //@singleToken
// Validation for the ``column-gap`` property.
func columnGap(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, false)
	if !length.IsNone() {
		return pr.Value{Dimension: length}
	}
	if getKeyword(token) == "normal" {
		return pr.Value{String: "normal"}
	}
	return nil
}

//@validator()
//@singleKeyword
// ``column-fill`` property validation.
func columnFill(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "balance":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``direction`` property validation.
func direction(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "ltr", "rtl":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``display`` property validation.
func display(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "none", "table-caption", "table-row-group", "table-cell",
		"table-header-group", "table-footer-group", "table-row",
		"table-column-group", "table-column":
		return pr.Display{keyword}
	case "inline-table", "inline-flex", "inline-grid":
		return pr.Display{"inline", keyword[7:]}
	case "inline-block":
		return pr.Display{"inline", "flow-root"}
	}

	var outside, inside, listItem string
	for _, token := range tokens {
		ident, ok := token.(parser.IdentToken)
		if !ok {
			return nil
		}
		value := string(ident.Value)
		switch value {
		case "block", "inline":
			if outside != "" {
				return nil
			}
			outside = value
		case "flow", "flow-root", "table", "flex", "grid":
			if inside != "" {
				return nil
			}
			inside = value
		case "list-item":
			if listItem != "" {
				return nil
			}
			listItem = value
		default:
			return nil
		}
	}

	if outside == "" {
		outside = "block"
	}
	if inside == "" {
		inside = "flow"
	}
	if listItem != "" {
		if inside == "flow" || inside == "flow-root" {
			return pr.Display{outside, inside, listItem}
		}
	} else {
		return pr.Display{outside, inside}
	}

	return nil
}

//@validator("float")
//@singleKeyword
// ``float`` property validation.
func float(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "left", "right", "footnote", "none":
		return pr.String(keyword)
	default:
		return nil
	}
}

func _fontFamily(tokens []Token) string {
	if len(tokens) == 0 {
		return ""
	}
	if tt, ok := tokens[0].(parser.StringToken); len(tokens) == 1 && ok {
		return tt.Value
	} else if len(tokens) > 0 {
		var values []string
		for _, token := range tokens {
			if tt, ok := token.(parser.IdentToken); ok {
				values = append(values, string(tt.Value))
			} else {
				return ""
			}
		}
		return strings.Join(values, " ")
	}
	return ""
}

// //@validator()
// //@commaSeparatedList
// ``font-family`` property validation.
func fontFamily(tokens []Token, _ string) pr.CssProperty {
	var out pr.Strings
	for _, part := range SplitOnComma(tokens) {
		result := _fontFamily(RemoveWhitespace(part))
		if result == "" {
			return nil
		}
		out = append(out, result)
	}
	return out
}

// //@validator()
// //@singleKeyword
func fontKerning(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "normal", "none":
		return pr.String(keyword)
	default:
		return nil
	}
}

// //@validator()
// //@singleToken
func fontLanguageOverride(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "normal" {
		return pr.String(keyword)
	}
	if tt, ok := token.(parser.StringToken); ok {
		return pr.String(tt.Value)
	}
	return nil
}

func parseFontVariant(tokens []Token, all utils.Set, couples [][]string) pr.SStrings {
	var values []string
	isInValues := func(s string, vs []string) bool {
		for _, v := range vs {
			if s == v {
				return true
			}
		}
		return false
	}
	for _, token := range tokens {
		ident, isIdent := token.(parser.IdentToken)
		if !isIdent {
			return pr.SStrings{}
		}
		identValue := string(ident.Value)
		if all.Has(identValue) {
			var concurrentValues []string
			for _, couple := range couples {
				if isInValues(identValue, couple) {
					concurrentValues = couple
					break
				}
			}
			for _, value := range concurrentValues {
				if isInValues(value, values) {
					return pr.SStrings{}
				}
			}
			values = append(values, identValue)
		} else {
			return pr.SStrings{}
		}
	}
	if len(values) > 0 {
		return pr.SStrings{Strings: values}
	}
	return pr.SStrings{}
}

// //@validator()
func fontVariantLigatures(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		if keyword == "normal" || keyword == "none" {
			return pr.SStrings{String: keyword}
		}
	}
	ss := parseFontVariant(tokens, allLigaturesValues, couplesLigatures)
	if ss.IsNone() {
		return nil
	}
	return ss
}

// //@validator()
// //@singleKeyword
func fontVariantPosition(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "sub", "super":
		return pr.String(keyword)
	default:
		return nil
	}
}

// //@validator()
// //@singleKeyword
func fontVariantCaps(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "small-caps", "all-small-caps", "petite-caps",
		"all-petite-caps", "unicase", "titling-caps":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
func fontVariantNumeric(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		if keyword == "normal" {
			return pr.SStrings{String: keyword}
		}
	}
	ss := parseFontVariant(tokens, allNumericValues, couplesNumeric)
	if ss.IsNone() {
		return nil
	}
	return ss
}

// //@validator()
// ``font-feature-settings`` property validation.
func fontFeatureSettings(tokens []Token, _ string) pr.CssProperty {
	s := _fontFeatureSettings(tokens)
	if s.IsNone() {
		return nil
	}
	return s
}

func _fontFeatureSettings(tokens []Token) pr.SIntStrings {
	if len(tokens) == 1 && getKeyword(tokens[0]) == "normal" {
		return pr.SIntStrings{String: "normal"}
	}

	fontFeatureSettingsList := func(tokens []Token) pr.IntString {
		var token Token
		feature, value := "", 0

		if len(tokens) == 2 {
			tokens, token = tokens[0:1], tokens[1]
			switch tt := token.(type) {
			case parser.IdentToken:
				if tt.Value == "on" {
					value = 1
				} else {
					value = 0
				}
			case parser.NumberToken:
				if tt.IsInteger && tt.IntValue() >= 0 {
					value = tt.IntValue()
				}
			}
		} else if len(tokens) == 1 {
			value = 1
		}

		if len(tokens) == 1 {
			token = tokens[0]
			tt, ok := token.(parser.StringToken)
			if ok && len(tt.Value) == 4 {
				ok := true
				for _, letter := range tt.Value {
					if !(0x20 <= letter && letter <= 0x7f) {
						ok = false
						break
					}
				}
				if ok {
					feature = tt.Value
				}
			}
		}

		if feature != "" {
			return pr.IntString{String: feature, Int: value}
		}
		return pr.IntString{}
	}

	var out pr.SIntStrings
	for _, part := range SplitOnComma(tokens) {
		result := fontFeatureSettingsList(RemoveWhitespace(part))
		if (result == pr.IntString{}) {
			return pr.SIntStrings{}
		}
		out.Values = append(out.Values, result)
	}
	return out
}

//@validator()
//@singleKeyword
func fontVariantAlternates(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	// TODO: support other values
	// See https://www.w3.org/TR/css-fonts-3/#font-variant-caps-prop
	switch keyword {
	case "normal", "historical-forms":
		return pr.String(keyword)
	default:
		return nil
	}
}

// //@validator()
func fontVariantEastAsian(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		if keyword == "normal" {
			return pr.SStrings{String: keyword}
		}
	}
	ss := parseFontVariant(tokens, allEastAsianValues, couplesEastAsian)
	if ss.IsNone() {
		return nil
	}
	return ss
}

//@validator()
//@singleToken
// ``font-size`` property validation.
func fontSize(tokens []Token, _ string) (pr.CssProperty, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]
	length := getLength(token, false, true)
	if !length.IsNone() {
		return pr.Value{Dimension: length}, nil
	}
	fontSizeKeyword := getKeyword(token)
	if _, isIn := pr.FontSizeKeywords[fontSizeKeyword]; isIn || fontSizeKeyword == "smaller" || fontSizeKeyword == "larger" {
		return pr.Value{String: fontSizeKeyword}, nil
	}
	return nil, nil
}

//@validator()
//@singleKeyword
// ``font-style`` property validation.
func fontStyle(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "italic", "oblique":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// Validation for the ``font-stretch`` property.
func fontStretch(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "ultra-condensed", "extra-condensed", "condensed", "semi-condensed",
		"normal", "semi-expanded", "expanded", "extra-expanded", "ultra-expanded":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleToken
// ``font-weight`` property validation.
func fontWeight(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "normal" || keyword == "bold" || keyword == "bolder" || keyword == "lighter" {
		return pr.IntString{String: keyword}
	}
	if number, ok := token.(parser.NumberToken); ok {
		intValue := number.IntValue()
		if number.IsInteger && (intValue == 100 || intValue == 200 || intValue == 300 || intValue == 400 || intValue == 500 || intValue == 600 || intValue == 700 || intValue == 800 || intValue == 900) {
			return pr.IntString{Int: intValue}
		}
	}
	return nil
}

// @validator()
// @single_keyword
func objectFit(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	// TODO: Figure out what the spec means by "'scale-down' flag".
	//  As of this writing, neither Firefox nor chrome support
	//  anything other than a single keyword as is done here.
	switch keyword {
	case "fill", "contain", "cover", "none", "scale-down":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator(unstable=true)
//@singleToken
func imageResolution(tokens []Token, _ string) pr.CssProperty {
	// TODO: support "snap" && "from-image"
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	value, ok := getResolution(token)
	if !ok {
		return nil
	}
	return pr.FToV(pr.Fl(value))
}

//@validator("letter-spacing")
//@validator("word-spacing")
//@singleToken
// Validation for ``letter-spacing`` && ``word-spacing``.
func spacing(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "normal" {
		return pr.Value{String: "normal"}
	}
	length := getLength(token, true, false)
	if !length.IsNone() {
		return pr.Value{Dimension: length}
	}
	return nil
}

//@validator()
//@singleToken
// ``line-height`` property validation.
func lineHeight(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "normal" {
		return pr.Value{String: "normal"}
	}

	switch tt := token.(type) {
	case parser.NumberToken:
		if tt.Value >= 0 {
			return pr.Value{Dimension: pr.Dimension{Value: pr.Float(tt.Value), Unit: pr.Scalar}}
		}
	case parser.PercentageToken:
		if tt.Value >= 0 {
			return pr.Value{Dimension: pr.Dimension{Value: pr.Float(tt.Value), Unit: pr.Percentage}}
		}
	case parser.DimensionToken:
		if tt.Value >= 0 {
			l := getLength(token, true, false)
			if l.IsNone() {
				return nil
			}
			return l.ToValue()
		}
	}
	return nil
}

//@validator()
//@singleKeyword
// ``list-style-position`` property validation.
func listStylePosition(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "inside", "outside":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleToken
// ``list-style-type`` property validation.
func listStyleType(tokens []Token, _ string) pr.CssProperty {
	out, ok := listStyleType_(tokens)
	if ok {
		return out
	}
	return nil
}

func listStyleType_(tokens []Token) (out pr.CounterStyleID, ok bool) {
	if len(tokens) != 1 {
		return out, false
	}
	token := tokens[0]
	switch token := token.(type) {
	case parser.IdentToken:
		return pr.CounterStyleID{Name: string(token.Value)}, true
	case parser.StringToken:
		return pr.CounterStyleID{Type: "string", Name: token.Value}, true
	case parser.FunctionBlock:
		if token.Name != "symbols" {
			return out, false
		}
		functionArguments := RemoveWhitespace(*token.Arguments)
		if len(functionArguments) == 0 {
			return out, false
		}
		arguments := []string{"symbolic"}
		if arg0, ok := functionArguments[0].(parser.IdentToken); ok {
			if arg0.Value == "cyclic" || arg0.Value == "numeric" || arg0.Value == "alphabetic" || arg0.Value == "symbolic" || arg0.Value == "fixed" {
				arguments = []string{string(arg0.Value)}
				functionArguments = functionArguments[1:]
			} else {
				return out, false
			}
		}

		if len(functionArguments) == 0 {
			return out, false
		}

		for _, arg := range functionArguments {
			if str, ok := arg.(parser.StringToken); ok {
				arguments = append(arguments, str.Value)
			} else {
				return out, false
			}
		}

		if arguments[0] == "alphabetic" || arguments[0] == "numeric" {
			if len(arguments) < 3 {
				return out, false
			}
		}
		return pr.CounterStyleID{Type: "symbols()", Name: arguments[0], Symbols: arguments[1:]}, true
	default:
		return out, false
	}
}

// @validator("min-width")
// @validator("min-height")
// @singleToken
// ``min-width`` && ``min-height`` properties validation.
func minWidthHeight(tokens []Token, _ string) pr.CssProperty {
	// See https://www.w3.org/TR/css-flexbox-1/#min-size-auto
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "auto" {
		return pr.SToV(keyword)
	} else {
		return lengthOrPercentage([]Token{token}, "")
	}
}

//@validator("padding-top")
//@validator("padding-right")
//@validator("padding-bottom")
//@validator("padding-left")
//@singleToken
// ``padding-*`` properties validation.
func lengthOrPercentage(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	l := getLength(token, false, true)
	if l.IsNone() {
		return nil
	}
	return pr.Value{Dimension: l}
}

//@validator("max-width")
//@validator("max-height")
//@singleToken
// Validation for max-width && max-height
func maxWidthHeight(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, true)
	if !length.IsNone() {
		return pr.Value{Dimension: length}
	}
	if getKeyword(token) == "none" {
		return pr.Value{Dimension: pr.Dimension{Value: pr.Inf, Unit: pr.Px}}
	}
	return nil
}

//@validator()
//@singleToken
// Validation for the ``opacity`` property.
func opacity(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok {
		return pr.Float(utils.MinF(1, utils.MaxF(0, number.Value)))
	}
	return nil
}

//@validator()
//@singleToken
// Validation for the ``z-index`` property.
func zIndex(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "auto" {
		return pr.IntString{String: "auto"}
	}
	if number, ok := token.(parser.NumberToken); ok {
		if number.IsInteger {
			return pr.IntString{Int: number.IntValue()}
		}
	}
	return nil
}

//@validator("orphans")
//@validator("widows")
//@singleToken
// Validation for the ``orphans`` && ``widows`` pr.
func orphansWidows(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok {
		value := number.IntValue()
		if number.IsInteger && value >= 1 {
			return pr.Int(value)
		}
	}
	return nil
}

//@validator()
//@singleToken
// Validation for the ``column-count`` property.
func columnCount(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok {
		value := number.IntValue()
		if number.IsInteger && value >= 1 {
			return pr.IntString{Int: value}
		}
	}
	if getKeyword(token) == "auto" {
		return pr.IntString{String: "auto"}
	}
	return nil
}

//@validator()
//@singleKeyword
// Validation for the ``overflow`` property.
func overflow(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "visible", "hidden", "scroll":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @single_keyword
// Validation for the ``text-overflow`` property.
func textOverflow(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "clip", "ellipsis":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleToken
// ``position`` property validation.
func position(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if fn, ok := token.(parser.FunctionBlock); ok && fn.Name == "running" && len(*fn.Arguments) == 1 {
		if ident, ok := (*fn.Arguments)[0].(parser.IdentToken); ok {
			return pr.BoolString{Bool: true, String: string(ident.Value)}
		}
	}
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "static", "relative", "absolute", "fixed":
		return pr.BoolString{String: keyword}
	default:
		return nil
	}
}

//@validator()
// ``quotes`` property validation.
func quotes(tokens []Token, _ string) pr.CssProperty {
	var opens, closes []string
	if len(tokens) > 0 && len(tokens)%2 == 0 {
		// Separate open && close quotes.
		// eg.  ("«", "»", "“", "”")  -> (("«", "“"), ("»", "”"))
		for i := 0; i < len(tokens); i += 2 {
			open, ok1 := tokens[i].(parser.StringToken)
			close_, ok2 := tokens[i+1].(parser.StringToken)
			if ok1 && ok2 {
				opens = append(opens, open.Value)
				closes = append(closes, close_.Value)
			} else {
				return nil
			}
		}
		return pr.Quotes{Open: opens, Close: closes}
	}
	return nil
}

//@validator()
//@singleKeyword
// Validation for the ``table-layout`` property
func tableLayout(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "fixed", "auto":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``text-align-all`` property validation.
func textAlignAll(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "left", "right", "center", "justify", "start", "end":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``text-align-last`` property validation.
func textAlignLast(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "left", "right", "center", "justify", "start", "end":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// ``text-decoration-line`` property validation.
func textDecorationLine(tokens []Token, _ string) pr.CssProperty {
	uniqKeywords := utils.Set{}
	valid := true
	for _, token := range tokens {
		keyword := getKeyword(token)
		if !(keyword == "underline" || keyword == "overline" || keyword == "line-through" || keyword == "blink") {
			valid = false
		}
		uniqKeywords.Add(keyword)
	}
	if _, in := uniqKeywords["none"]; len(uniqKeywords) == 1 && in { // then uniqKeywords == {"none"}
		return pr.Decorations{}
	}
	if valid && len(uniqKeywords) == len(tokens) {
		return pr.Decorations(uniqKeywords)
	}
	return nil
}

// @validator()
// @singleKeyword
// ``text-decoration-style`` property validation.
func textDecorationStyle(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "solid", "double", "dotted", "dashed", "wavy":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleToken
// ``text-indent`` property validation.
func textIndent(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	l := getLength(token, true, true)
	if l.IsNone() {
		return nil
	}
	return pr.Value{Dimension: l}
}

//@validator()
//@singleKeyword
// ``text-align`` property validation.
func textTransform(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "none", "uppercase", "lowercase", "capitalize", "full-width":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleToken
// Validation for the ``vertical-align`` property
func verticalAlign(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, true, true)
	if !length.IsNone() {
		return pr.Value{Dimension: length}
	}
	keyword := getKeyword(token)
	if keyword == "baseline" || keyword == "middle" || keyword == "sub" || keyword == "super" || keyword == "text-top" || keyword == "text-bottom" || keyword == "top" || keyword == "bottom" {
		return pr.Value{String: keyword}
	}
	return nil
}

//@validator()
//@singleKeyword
// ``visibility`` property validation.
func visibility(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "visible", "hidden", "collapse":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``white-space`` property validation.
func whiteSpace(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "pre", "nowrap", "pre-wrap", "pre-line":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``overflow-wrap`` property validation.
func overflowWrap(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "anywhere", "normal", "break-word":
		return pr.String(keyword)
	default:
		return nil
	}
}

// Validation for ``footnote-display``.
func footnoteDisplay(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "block", "inline", "compact":
		return pr.String(keyword)
	default:
		return nil
	}
}

// Validation for ``footnote-policy``.
func footnotePolicy(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "line", "block":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleToken
// ``flex-basis`` property validation.
func flexBasis(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	basis := widthHeight(tokens, "")
	if basis != nil {
		return basis
	}
	if getKeyword(token) == "content" {
		return pr.SToV("content")
	}
	return nil
}

// @validator()
// @singleKeyword
// ``flex-direction`` property validation.
func flexDirection(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "row", "row-reverse", "column", "column-reverse":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator("flex-grow")
// @validator("flex-shrink")
// @singleToken
func flexGrowShrink(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok {
		return pr.Float(number.Value)
	}
	return nil
}

// @validator()
// @singleToken
func order(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok && number.IsInteger {
		return pr.Int(number.IntValue())
	}
	return nil
}

// @validator()
// @singleKeyword
// ``flex-wrap`` property validation.
func flexWrap(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "nowrap", "wrap", "wrap-reverse":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// ``justify-content`` property validation.
func justifyContent(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "flex-start", "flex-end", "center", "space-between", "space-around", "space-evenly", "stretch":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// ``align-items`` property validation.
func alignItems(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "flex-start", "flex-end", "center", "baseline", "stretch":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// ``align-self`` property validation.
func alignSelf(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "flex-start", "flex-end", "center", "baseline", "stretch":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// ``align-content`` property validation.
func alignContent(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "flex-start", "flex-end", "center", "space-between", "space-around",
		"space-evenly", "stretch":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator(unstable=true)
//@singleKeyword
// Validation for ``image-rendering``.
func imageRendering(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "crisp-edges", "pixelated":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator(unstable=true)
// ``size`` property validation.
// See http://www.w3.org/TR/css3-page/#page-size-prop
func size(tokens []Token, _ string) pr.CssProperty {
	var (
		lengths        []pr.Dimension
		keywords       []string
		lengthsNotNone bool = true
	)
	for _, token := range tokens {
		length, keyword := getLength(token, false, false), getKeyword(token)
		lengthsNotNone = lengthsNotNone && !length.IsNone()
		lengths = append(lengths, length)
		keywords = append(keywords, keyword)
	}

	if lengthsNotNone {
		if len(lengths) == 1 {
			return pr.Point{lengths[0], lengths[0]}
		} else if len(lengths) == 2 {
			return pr.Point{lengths[0], lengths[1]}
		}
	}

	if len(keywords) == 1 {
		keyword := keywords[0]
		if psize, in := pr.PageSizes[keyword]; in {
			return psize
		} else if keyword == "auto" || keyword == "portrait" {
			return pr.A4
		} else if keyword == "landscape" {
			return pr.Point{pr.A4[1], pr.A4[0]}
		}
	}

	if len(keywords) == 2 {
		var orientation, pageSize string
		if keywords[0] == "portrait" || keywords[0] == "landscape" {
			orientation, pageSize = keywords[0], keywords[1]
		} else if keywords[1] == "portrait" || keywords[1] == "landscape" {
			pageSize, orientation = keywords[0], keywords[1]
		}
		if widthHeight, in := pr.PageSizes[pageSize]; in {
			if orientation == "portrait" {
				return widthHeight
			} else {
				return pr.Point{widthHeight[1], widthHeight[0]}
			}
		}
	}
	return nil
}

//@validator(proprietary=true)
//@singleToken
// Validation for ``anchor``.
func anchor(tokens []Token, _ string) (out pr.CssProperty) {
	if len(tokens) != 1 {
		return
	}
	token := tokens[0]
	if getKeyword(token) == "none" {
		return pr.String("none")
	}
	name, args := parseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(parser.IdentToken); ok && name == "attr" {
				return pr.AttrData{Name: string(ident.Value)}
			}
		}
	}
	return
}

//@validator(proprietary=true, wantsBaseUrl=true)
//@singleToken
// Validation for ``link``.
func link(tokens []Token, baseUrl string) (out pr.CssProperty, err error) {
	if len(tokens) != 1 {
		return
	}
	token := tokens[0]
	if getKeyword(token) == "none" {
		return pr.NamedString{Name: "none"}, nil
	}

	parsedUrl, attr, err := getUrl(token, baseUrl)
	if err != nil {
		return
	}
	if !parsedUrl.IsNone() {
		return parsedUrl, nil
	}
	name, args := parseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(parser.IdentToken); ok && name == "attr" {
				attr = pr.AttrData{Name: string(ident.Value)}
			}
		}
	}
	if !attr.IsNone() {
		out = attr
	}
	return
}

//@validator()
//@singleToken
// Validation for ``tab-size``.
// See https://www.w3.org/TR/css-text-3/#tab-size
func tabSize(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok {
		if number.IsInteger && number.Value >= 0 { // no unit means multiple of space width
			return pr.Value{Dimension: pr.Dimension{Value: pr.Float(number.Value)}}
		}
	}
	return pr.Value{Dimension: getLength(token, false, false)}
}

//@validator(unstable=true)
//@singleToken
// Validation for ``hyphens``.
func hyphens(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	switch keyword {
	case "none", "manual", "auto":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator(unstable=true)
//@singleToken
// Validation for ``hyphenate-character``.
func hyphenateCharacter(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "auto" {
		return pr.String("‐")
	} else if str, ok := token.(parser.StringToken); ok {
		return pr.String(str.Value)
	}
	return nil
}

//@validator(unstable=true)
//@singleToken
// Validation for ``hyphenate-limit-zone``.
func hyphenateLimitZone(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	d := getLength(token, false, true)
	if d.IsNone() {
		return nil
	}
	return pr.Value{Dimension: d}
}

//@validator(unstable=true)
// Validation for ``hyphenate-limit-chars``.
func hyphenateLimitChars(tokens []Token, _ string) pr.CssProperty {
	switch len(tokens) {
	case 1:
		token := tokens[0]
		keyword := getKeyword(token)
		if keyword == "auto" {
			return pr.Ints3{5, 2, 2}
		} else if number, ok := token.(parser.NumberToken); ok && number.IsInteger {
			return pr.Ints3{number.IntValue(), 2, 2}
		}
	case 2:
		total, left := tokens[0], tokens[1]
		totalKeyword := getKeyword(total)
		leftKeyword := getKeyword(left)
		if totalNumber, ok := total.(parser.NumberToken); ok && totalNumber.IsInteger {
			if leftNumber, ok := left.(parser.NumberToken); ok && leftNumber.IsInteger {
				return pr.Ints3{totalNumber.IntValue(), leftNumber.IntValue(), leftNumber.IntValue()}
			} else if leftKeyword == "auto" {
				return pr.Ints3{totalNumber.IntValue(), 2, 2}
			}
		} else if totalKeyword == "auto" {
			if leftNumber, ok := left.(parser.NumberToken); ok && leftNumber.IsInteger {
				return pr.Ints3{5, leftNumber.IntValue(), leftNumber.IntValue()}
			} else if leftKeyword == "auto" {
				return pr.Ints3{5, 2, 2}
			}
		}
	case 3:
		total, left, right := tokens[0], tokens[1], tokens[2]
		totalNumber, okT := total.(parser.NumberToken)
		leftNumber, okL := left.(parser.NumberToken)
		rightNumber, okR := right.(parser.NumberToken)
		if ((okT && totalNumber.IsInteger) || getKeyword(total) == "auto") &&
			((okL && leftNumber.IsInteger) || getKeyword(left) == "auto") &&
			((okR && rightNumber.IsInteger) || getKeyword(right) == "auto") {
			totalInt := 5
			if okT {
				totalInt = totalNumber.IntValue()
			}
			leftInt := 2
			if okL {
				leftInt = leftNumber.IntValue()
			}
			rightInt := 2
			if okR {
				rightInt = rightNumber.IntValue()
			}
			return pr.Ints3{totalInt, leftInt, rightInt}
		}
	}
	return nil
}

//@validator(proprietary=true)
//@singleToken
// Validation for ``lang``.
func lang(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "none" {
		return pr.NamedString{Name: "none"}
	}
	name, args := parseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(parser.IdentToken); ok && name == "attr" {
				return pr.NamedString{Name: "attr()", String: string(ident.Value)}
			}
		}
	} else if str, ok := token.(parser.StringToken); ok {
		return pr.NamedString{Name: "string", String: str.Value}
	}
	return nil
}

//@validator(unstable=true)
// Validation for ``bookmark-label``.
func bookmarkLabel(tokens []Token, baseUrl string) (out pr.CssProperty, err error) {
	parsedTokens := make(pr.ContentProperties, len(tokens))
	for index, v := range tokens {
		parsedTokens[index], err = getContentListToken(v, baseUrl)
		if err != nil {
			return nil, err
		}
		if parsedTokens[index].IsNone() {
			return nil, nil
		}
	}
	return parsedTokens, nil
}

//@validator(unstable=true)
//@singleToken
// Validation for ``bookmark-level``.
func bookmarkLevel(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok && number.IsInteger && number.IntValue() >= 1 {
		return pr.IntString{Int: number.IntValue()}
	} else if getKeyword(token) == "none" {
		return pr.IntString{String: "none"}
	}
	return nil
}

// @validator(unstable=True)
// @single_keyword
// Validation for ``bookmark-state``.
func bookmarkState(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "open", "closed":
		return pr.String(keyword)
	default:
		return nil
	}
}

//@validator(unstable=true)
//@commaSeparatedList
// Validation for ``string-set``.
func _stringSet(tokens []Token, baseUrl string) (out pr.SContent, err error) {
	// Spec asks for strings after custom keywords, but we allow content-lists
	if len(tokens) >= 2 {
		varName := getCustomIdent(tokens[0])
		if varName == "" {
			return
		}
		parsedTokens := make([]pr.ContentProperty, len(tokens)-1)
		for i, token := range tokens[1:] {
			parsedTokens[i], err = getContentListToken(token, baseUrl)
			if err != nil {
				return
			}
			if parsedTokens[i].IsNone() {
				return
			}
		}
		return pr.SContent{String: varName, Contents: parsedTokens}, nil
	} else if len(tokens) > 0 && getKeyword(tokens[0]) == "none" {
		return pr.SContent{String: "none"}, nil
	}
	return
}

func stringSet(tokens []Token, baseUrl string) (pr.CssProperty, error) {
	var out pr.StringSet
	for _, part := range SplitOnComma(tokens) {
		result, err := _stringSet(RemoveWhitespace(part), baseUrl)
		if err != nil {
			return nil, err
		}
		if result.IsNone() {
			return nil, nil
		}
		out.Contents = append(out.Contents, result)
	}
	return out, nil
}

//@validator()
func transform(tokens []Token, _ string) (pr.CssProperty, error) {
	if getSingleKeyword(tokens) == "none" {
		return pr.Transforms{}, nil
	}
	out := make(pr.Transforms, len(tokens))
	var err error
	for index, v := range tokens {
		out[index], err = transformFunction(v)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// @property()
// @single_token
// ``box-ellipsis`` property validation.
func blockEllipsis(tokens []Token, _ string) pr.CssProperty {
	if v, ok := blockEllipsis_(tokens); ok {
		return v
	}
	return nil
}

func blockEllipsis_(tokens []Token) (out pr.NamedString, ok bool) {
	if len(tokens) != 1 {
		return
	}
	token := tokens[0]
	if str, ok := token.(parser.StringToken); ok {
		return pr.NamedString{Name: "string", String: str.Value}, true
	}
	keyword := getKeyword(token)
	if keyword == "none" || keyword == "auto" {
		return pr.NamedString{Name: keyword}, true
	}
	return
}

func transformFunction(token Token) (pr.SDimensions, error) {
	name, args := parseFunction(token)
	if name == "" {
		return pr.SDimensions{}, InvalidValue
	}

	lengths, values := make([]pr.Dimension, len(args)), make([]pr.Dimension, len(args))
	isAllNumber, isAllLengths := true, true
	for index, a := range args {
		lengths[index] = getLength(a, true, true)
		isAllLengths = isAllLengths && !lengths[index].IsNone()
		if aNumber, ok := a.(parser.NumberToken); ok {
			values[index] = pr.FToD(pr.Fl(aNumber.Value))
		} else {
			isAllNumber = false
		}
	}
	switch len(args) {
	case 1:
		angle, notNone := getAngle(args[0])
		length := getLength(args[0], true, true)
		switch name {
		case "rotate":
			if notNone && angle != 0 {
				return pr.SDimensions{String: "rotate", Dimensions: []pr.Dimension{pr.FToD(pr.Fl(angle))}}, nil
			}
		case "skewx":
			if notNone && angle != 0 {
				return pr.SDimensions{String: "skew", Dimensions: []pr.Dimension{pr.FToD(pr.Fl(angle)), pr.ZeroPixels}}, nil
			}
		case "skewy":
			if notNone && angle != 0 {
				return pr.SDimensions{String: "skew", Dimensions: []pr.Dimension{pr.ZeroPixels, pr.FToD(pr.Fl(angle))}}, nil
			}
		case "translatex", "translate":
			if !length.IsNone() {
				return pr.SDimensions{String: "translate", Dimensions: []pr.Dimension{length, pr.ZeroPixels}}, nil
			}
		case "translatey":
			if !length.IsNone() {
				return pr.SDimensions{String: "translate", Dimensions: []pr.Dimension{pr.ZeroPixels, length}}, nil
			}
		case "scalex":
			if number, ok := args[0].(parser.NumberToken); ok {
				return pr.SDimensions{String: "scale", Dimensions: []pr.Dimension{pr.FToD(pr.Fl(number.Value)), pr.FToD(1.)}}, nil
			}
		case "scaley":
			if number, ok := args[0].(parser.NumberToken); ok {
				return pr.SDimensions{String: "scale", Dimensions: []pr.Dimension{pr.FToD(1.), pr.FToD(pr.Fl(number.Value))}}, nil
			}
		case "scale":
			if number, ok := args[0].(parser.NumberToken); ok {
				return pr.SDimensions{String: "scale", Dimensions: []pr.Dimension{pr.FToD(pr.Fl(number.Value)), pr.FToD(pr.Fl(number.Value))}}, nil
			}
		}
	case 2:
		if name == "scale" && isAllNumber {
			return pr.SDimensions{String: name, Dimensions: values}, nil
		}
		if name == "translate" && isAllLengths {
			return pr.SDimensions{String: name, Dimensions: lengths}, nil
		}
	case 6:
		if name == "matrix" && isAllNumber {
			return pr.SDimensions{String: name, Dimensions: values}, nil
		}
	}
	return pr.SDimensions{}, InvalidValue
}

func maxLines(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if token, ok := token.(parser.NumberToken); ok {
		if token.IsInteger {
			return pr.IntString{Int: token.IntValue()}
		}
	}
	if keyword := getKeyword(token); keyword == "none" {
		return pr.IntString{String: "none"}
	}
	return nil
}

func continue_(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "discard":
		return pr.String(keyword)
	default:
		return nil
	}
}

// Remove any top-level whitespace in a token list.
func RemoveWhitespace(tokens []Token) []Token {
	var out []Token
	for _, token := range tokens {
		if token.Type() != parser.WhitespaceTokenT && token.Type() != parser.CommentT {
			out = append(out, token)
		}
	}
	return out
}
