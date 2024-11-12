package validation

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"

	pa "github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/selector"
)

// Expand shorthands and validate property values.
// See http://www.w3.org/TR/CSS21/propidx.html and various CSS3 modules.

// :copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

const proprietaryPrefix = "-weasy-"

var (
	ErrInvalidValue = errors.New("invalid or unsupported values for a known CSS property")

	LENGTHUNITS = map[string]pr.Unit{"ex": pr.Ex, "em": pr.Em, "ch": pr.Ch, "rem": pr.Rem, "px": pr.Px, "pt": pr.Pt, "pc": pr.Pc, "in": pr.In, "cm": pr.Cm, "mm": pr.Mm, "q": pr.Q}
	AngleUnits  = map[string]pr.Unit{"rad": pr.Rad, "turn": pr.Turn, "deg": pr.Deg, "grad": pr.Grad}
	// keyword -> (open, insert)
	contentQuoteKeywords = map[string]pr.Quote{
		"open-quote":     {Open: true, Insert: true},
		"close-quote":    {Open: false, Insert: true},
		"no-open-quote":  {Open: true, Insert: false},
		"no-close-quote": {Open: false, Insert: false},
	}

	zeroPercent    = pr.PercToD(0)
	fiftyPercent   = pr.PercToD(50)
	hundredPercent = pr.PercToD(100)

	backgroundPositionsPercentages = map[string]pr.Dimension{
		"top":    zeroPercent,
		"left":   zeroPercent,
		"center": fiftyPercent,
		"bottom": hundredPercent,
		"right":  hundredPercent,
	}

	// http://drafts.csswg.org/csswg/css3-values/#angles
	// 1<unit> is this many radians.
	ANGLETORADIANS = map[pr.Unit]utils.Fl{
		pr.Rad:  1,
		pr.Turn: 2 * math.Pi,
		pr.Deg:  math.Pi / 180,
		pr.Grad: math.Pi / 200,
	}

	// http://drafts.csswg.org/csswg/css-values/#resolution
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
	validators = [...]validator{
		pr.PAppearance:              appearance,
		pr.PBackgroundAttachment:    backgroundAttachment,
		pr.PBackgroundColor:         otherColors,
		pr.PBorderTopColor:          otherColors,
		pr.PBorderRightColor:        otherColors,
		pr.PBorderBottomColor:       otherColors,
		pr.PBorderLeftColor:         otherColors,
		pr.PColumnRuleColor:         otherColors,
		pr.PTextDecorationColor:     otherColors,
		pr.POutlineColor:            outlineColor,
		pr.PBorderCollapse:          borderCollapse,
		pr.PEmptyCells:              emptyCells,
		pr.PTransformOrigin:         transformOrigin,
		pr.PObjectPosition:          objectPosition,
		pr.PBackgroundPosition:      backgroundPosition,
		pr.PBackgroundRepeat:        backgroundRepeat,
		pr.PBackgroundSize:          backgroundSize,
		pr.PBackgroundClip:          box,
		pr.PBackgroundOrigin:        box,
		pr.PBorderSpacing:           borderSpacing,
		pr.PBorderTopRightRadius:    borderCornerRadius,
		pr.PBorderBottomRightRadius: borderCornerRadius,
		pr.PBorderBottomLeftRadius:  borderCornerRadius,
		pr.PBorderTopLeftRadius:     borderCornerRadius,
		pr.PBorderTopStyle:          borderStyle,
		pr.PBorderRightStyle:        borderStyle,
		pr.PBorderLeftStyle:         borderStyle,
		pr.PBorderBottomStyle:       borderStyle,
		pr.PColumnRuleStyle:         borderStyle,
		pr.PBreakBefore:             breakBeforeAfter,
		pr.PBreakAfter:              breakBeforeAfter,
		pr.PBreakInside:             breakInside,
		pr.PBoxDecorationBreak:      boxDecorationBreak,
		pr.PMarginBreak:             marginBreak,
		pr.PPage:                    page,
		pr.PBleedLeft:               bleed,
		pr.PBleedRight:              bleed,
		pr.PBleedTop:                bleed,
		pr.PBleedBottom:             bleed,
		pr.PMarks:                   marks,
		pr.POutlineStyle:            outlineStyle,
		pr.PBorderTopWidth:          borderWidth,
		pr.PBorderRightWidth:        borderWidth,
		pr.PBorderLeftWidth:         borderWidth,
		pr.PBorderBottomWidth:       borderWidth,
		pr.PBorderImageSlice:        borderImageSlice,
		pr.PBorderImageWidth:        borderImageWidth,
		pr.PBorderImageOutset:       borderImageOutset,
		pr.PBorderImageRepeat:       borderImageRepeat,
		pr.PColumnRuleWidth:         borderWidth,
		pr.POutlineWidth:            borderWidth,
		pr.PColumnWidth:             columnWidth,
		pr.PColumnSpan:              columnSpan,
		pr.PBoxSizing:               boxSizing,
		pr.PCaptionSide:             captionSide,
		pr.PClear:                   clear,
		pr.PClip:                    clip,
		pr.PTop:                     lengthPercOrAuto,
		pr.PRight:                   lengthPercOrAuto,
		pr.PLeft:                    lengthPercOrAuto,
		pr.PBottom:                  lengthPercOrAuto,
		pr.PMarginTop:               lengthPercOrAuto,
		pr.PMarginRight:             lengthPercOrAuto,
		pr.PMarginBottom:            lengthPercOrAuto,
		pr.PMarginLeft:              lengthPercOrAuto,
		pr.PHeight:                  widthHeight,
		pr.PWidth:                   widthHeight,
		pr.PColumnFill:              columnFill,
		pr.PDirection:               direction,
		pr.PDisplay:                 display,
		pr.PFloat:                   float,
		pr.PFontFamily:              fontFamily,
		pr.PFontKerning:             fontKerning,
		pr.PFontLanguageOverride:    fontLanguageOverride,
		pr.PFontVariantLigatures:    fontVariantLigatures,
		pr.PFontVariantPosition:     fontVariantPosition,
		pr.PFontVariantCaps:         fontVariantCaps,
		pr.PFontVariantNumeric:      fontVariantNumeric,
		pr.PFontFeatureSettings:     fontFeatureSettings,
		pr.PFontVariantAlternates:   fontVariantAlternates,
		pr.PFontVariantEastAsian:    fontVariantEastAsian,
		pr.PFontVariationSettings:   fontVariationSettings,
		pr.PFontStyle:               fontStyle,
		pr.PFontStretch:             fontStretch,
		pr.PFontWeight:              fontWeight,
		pr.PFootnoteDisplay:         footnoteDisplay,
		pr.PFootnotePolicy:          footnotePolicy,
		pr.PImageResolution:         imageResolution,
		pr.PLetterSpacing:           spacing,
		pr.PWordSpacing:             spacing,
		pr.PLineHeight:              lineHeight,
		pr.PListStylePosition:       listStylePosition,
		pr.PListStyleType:           listStyleType,
		pr.PPaddingTop:              lengthOrPercentage,
		pr.PPaddingRight:            lengthOrPercentage,
		pr.PPaddingBottom:           lengthOrPercentage,
		pr.PPaddingLeft:             lengthOrPercentage,
		pr.PMinWidth:                minWidthHeight,
		pr.PMinHeight:               minWidthHeight,
		pr.PMaxWidth:                maxWidthHeight,
		pr.PMaxHeight:               maxWidthHeight,
		pr.POpacity:                 opacity,
		pr.PZIndex:                  zIndex,
		pr.POrphans:                 orphansWidows,
		pr.PWidows:                  orphansWidows,
		pr.PColumnCount:             columnCount,
		pr.POverflow:                overflow,
		pr.PPosition:                position,
		pr.PQuotes:                  quotes,
		pr.PTableLayout:             tableLayout,
		pr.PTextAlignAll:            textAlignAll,
		pr.PTextAlignLast:           textAlignLast,
		pr.PTextDecorationLine:      textDecorationLine,
		pr.PTextDecorationStyle:     textDecorationStyle,
		pr.PTextIndent:              textIndent,
		pr.PTextTransform:           textTransform,
		pr.PVerticalAlign:           verticalAlign,
		pr.PVisibility:              visibility,
		pr.PWhiteSpace:              whiteSpace,
		pr.POverflowWrap:            overflowWrap,
		pr.PImageRendering:          imageRendering,
		pr.PImageOrientation:        imageOrientation,
		pr.PSize:                    size,
		pr.PTabSize:                 tabSize,
		pr.PHyphens:                 hyphens,
		pr.PHyphenateCharacter:      hyphenateCharacter,
		pr.PHyphenateLimitZone:      hyphenateLimitZone,
		pr.PHyphenateLimitChars:     hyphenateLimitChars,
		pr.PLang:                    lang,
		pr.PBookmarkLevel:           bookmarkLevel,
		pr.PBookmarkState:           bookmarkState,
		pr.PObjectFit:               objectFit,
		pr.PTextOverflow:            textOverflow,
		pr.PFlexBasis:               flexBasis,
		pr.PFlexDirection:           flexDirection,
		pr.PFlexGrow:                flexGrowShrink,
		pr.PFlexShrink:              flexGrowShrink,
		pr.POrder:                   order,
		pr.PColumnGap:               gap,
		pr.PRowGap:                  gap,
		pr.PFlexWrap:                flexWrap,
		pr.PAlignContent:            alignContent,
		pr.PAlignItems:              alignItems,
		pr.PAlignSelf:               alignSelf,
		pr.PJustifyContent:          justifyContent,
		pr.PJustifyItems:            justifyItems,
		pr.PJustifySelf:             justifySelf,
		pr.PAnchor:                  anchor,
		pr.PBlockEllipsis:           blockEllipsis,
		pr.PContinue:                continue_,
		pr.PMaxLines:                maxLines,
		pr.PWordBreak:               wordBreak,
		pr.PGridAutoColumns:         gridAuto,
		pr.PGridAutoRows:            gridAuto,
		pr.PGridAutoFlow:            gridAutoFlow,
		pr.PGridTemplateColumns:     gridTemplate,
		pr.PGridTemplateRows:        gridTemplate,
		pr.PGridTemplateAreas:       gridTemplateAreas,
		pr.PGridRowStart:            gridLine,
		pr.PGridColumnStart:         gridLine,
		pr.PGridRowEnd:              gridLine,
		pr.PGridColumnEnd:           gridLine,
	}
	validatorsError = map[pr.KnownProp]validatorError{
		pr.PBackgroundImage:   backgroundImage,
		pr.PBorderImageSource: borderImageSource,
		pr.PListStyleImage:    listStyleImage,
		pr.PContent:           content,
		pr.PCounterIncrement:  counterIncrement,
		pr.PCounterReset:      counterReset,
		pr.PCounterSet:        counterReset,
		pr.PFontSize:          fontSize,
		pr.PBookmarkLabel:     bookmarkLabel,
		pr.PTransform:         transform,
		pr.PStringSet:         stringSet,
		pr.PLink:              link,
	}

	// regroup the two cases (with error or without error)
	allValidators = pr.NewSetK(pr.PColor) // special case because of inherited

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
	for name, v := range validators {
		if v != nil {
			allValidators.Add(pr.KnownProp(name))
		}
	}
	for name := range validatorsError {
		allValidators.Add(name)
	}
}

type Token = pa.Token

type (
	validator      func(tokens []Token, baseUrl string) pr.CssProperty // dont support var(), attr()
	validatorError func(tokens []Token, baseUrl string) (pr.CssProperty, error)
)

// ValidateKnown validate one known, non shortand, property.
func ValidateKnown(name pr.KnownProp, tokens []Token, baseUrl string) (out pr.DeclaredValue, err error) {
	if name == pr.PColor { // special case to handle inherit
		return color(tokens, ""), nil
	}

	var value pr.CssProperty
	if function := validators[name]; function != nil {
		value = function(tokens, baseUrl)
	} else if functionE := validatorsError[name]; functionE != nil {
		value, err = functionE(tokens, baseUrl)
	}
	return value, err
}

func Validate(key pr.PropKey, tokens []Token) (pr.DeclaredValue, error) {
	out, err := validateNonShorthand("", key.String(), tokens, false)
	return out.property, err
}

// Default validator for non-shorthand pr.
// required = false
func validateNonShorthand(baseUrl string, name string, tokens []pa.Token, required bool) (out namedProperty, err error) {
	if strings.HasPrefix(name, "--") { // variable
		// can't validate variables contents before substitution
		return namedProperty{
			name:     pr.PropKey{Var: name},
			property: pr.RawTokens(tokens),
		}, nil
	}

	prop := pr.PropsFromNames[name]
	if !required && !pr.KnownProperties.Has(prop) {
		return out, errors.New("unknown property")
	}

	if _, isSupported := allValidators[prop]; !required && !isSupported {
		return out, fmt.Errorf("property %s not supported yet", name)
	}

	for _, token := range tokens {
		if HasVar(token) {
			// Found CSS variable, return pending-substitution values.
			return namedProperty{name: pr.PropKey{KnownProp: prop}, property: pr.RawTokens(tokens)}, nil
		}
	}

	var value pr.DeclaredValue
	keyword := getSingleKeyword(tokens)
	if keyword == "initial" || keyword == "inherit" {
		value = pr.NewDefaultValue(keyword)
	} else {
		value, err = ValidateKnown(prop, tokens, baseUrl)
		if err != nil {
			return out, err
		}
		if value == nil {
			return out, errors.New("invalid value (nil function return)")
		}
	}

	return namedProperty{name: pr.PropKey{KnownProp: prop}, property: value}, nil
}

// Not applicable to the print media
var notPrintMedia = utils.NewSet(
	// Aural media
	"azimuth",
	"cue",
	"cue-after",
	"cue-before",
	"elevation",
	"pause",
	"pause-after",
	"pause-before",
	"pitch-range",
	"pitch",
	"play-during",
	"richness",
	"speak-header",
	"speak-numeral",
	"speak-punctuation",
	"speak",
	"speech-rate",
	"stress",
	"voice-family",
	"volume",
	// Animations, transitions, timelines
	"animation",
	"animation-composition",
	"animation-delay",
	"animation-direction",
	"animation-duration",
	"animation-fill-mode",
	"animation-iteration-count",
	"animation-name",
	"animation-play-state",
	"animation-range",
	"animation-range-end",
	"animation-range-start",
	"animation-timeline",
	"animation-timing-function",
	"timeline-scope",
	"transition",
	"transition-delay",
	"transition-duration",
	"transition-property",
	"transition-timing-function",
	"view-timeline",
	"view-timeline-axis",
	"view-timeline-inset",
	"view-timeline-name",
	"view-transition-name",
	"will-change",
	// Dynamic and interactive
	"caret",
	"caret-color",
	"caret-shape",
	"cursor",
	"field-sizing",
	"pointer-event",
	"resize",
	"touch-action",
	// Browser viewport scrolling
	"overscroll-behavior",
	"overscroll-behavior-block",
	"overscroll-behavior-inline",
	"overscroll-behavior-x",
	"overscroll-behavior-y",
	"scroll-behavior",
	"scroll-margin",
	"scroll-margin-block",
	"scroll-margin-block-end",
	"scroll-margin-block-start",
	"scroll-margin-bottom",
	"scroll-margin-inline",
	"scroll-margin-inline-end",
	"scroll-margin-inline-start",
	"scroll-margin-left",
	"scroll-margin-right",
	"scroll-margin-top",
	"scroll-padding",
	"scroll-padding-block",
	"scroll-padding-block-end",
	"scroll-padding-block-start",
	"scroll-padding-bottom",
	"scroll-padding-inline",
	"scroll-padding-inline-end",
	"scroll-padding-inline-start",
	"scroll-padding-left",
	"scroll-padding-right",
	"scroll-padding-top",
	"scroll-snap-align",
	"scroll-snap-stop",
	"scroll-snap-type",
	"scroll-timeline",
	"scroll-timeline-axis",
	"scroll-timeline-name",
	"scrollbar-color",
	"scrollbar-gutter",
	"scrollbar-width",
)

// Declaration is the input form of a CSS property,
// possibly containing variables.
type Declaration struct {
	Name  pr.PropKey
	Value pr.DeclaredValue

	// Shortand is not zero for shortands containing 'var()' tokens, waiting to be expanded and validated
	// In this case, [Value] is [pr.RawTokens] and refers to the associate shorthand, not to the expanded [Name]
	Shortand pr.Shortand

	Important bool
}

type KeyedDeclarations struct {
	Selector     selector.SelectorGroup
	Declarations []Declaration
}

var (
	pos11 = pa.Pos{Line: 1, Column: 1}
	colon = pa.NewLiteral(":", pos11)
)

// See PreprocessDeclarationsPrelude
func PreprocessDeclarations(baseUrl string, declarations []pa.Compound) []Declaration {
	tmp, _ := PreprocessDeclarationsPrelude(baseUrl, declarations, nil)
	return tmp[0].Declarations
}

// PreprocessDeclarationsPrelude filter unsupported properties or parsing errors,
// and expand shortand properties.
//
// Properties containing var() tokens are not validated yet.
// Shortand containing var() tokens are not expanded.
//
// Log a warning for every ignored declaration.
//
// If [prelude] is nil, the returned error is always nil.
// The returned slice is never empty, and has always length 1 if [prelude] is nil.
func PreprocessDeclarationsPrelude(baseURL string, declarations []pa.Compound, prelude []pa.Token) ([]KeyedDeclarations, error) {
	// Compile list of selectors.
	var selectors selector.SelectorGroup
	if prelude != nil {
		// Handle & selector in non-nested rule. MDN explains that & is
		// then equivalent to :scope, and :scope is equivalent to :root
		// as we don’t support :scope yet.
		originalPrelude := prelude
		prelude = []Token{}
		for _, token := range originalPrelude {
			if pa.IsLiteral(token, "&") {
				prelude = append(prelude, colon, pa.NewIdent("root", pos11))
			} else {
				prelude = append(prelude, token)
			}
		}
		var err error
		selectors, err = selector.ParseGroup(pa.Serialize(prelude))
		if err != nil {
			return nil, err
		}
	}

	// Yield declarations.
	is := pa.NewFunctionBlock(pos11, "is", prelude)
	var (
		out      []KeyedDeclarations
		ownDecls []Declaration
	)
	for _, declaration := range declarations {
		if errToken, ok := declaration.(pa.ParseError); ok {
			logger.WarningLogger.Printf("Error: %s \n", errToken.Message)
		}

		if declaration, ok := declaration.(pa.QualifiedRule); ok {
			// Nested rule.
			if prelude == nil {
				continue
			}
			hasNesting := false
			// Replace & selector by parent.
			var declarationPrelude []Token
			for _, token := range declaration.Prelude {
				if pa.IsLiteral(token, "&") {
					hasNesting = true
					declarationPrelude = append(declarationPrelude, colon, is)
				} else {
					declarationPrelude = append(declarationPrelude, token)
				}
			}
			if !hasNesting {
				// No & selector, prepend parent.
				declarationPrelude = append([]Token{colon, is, pa.NewWhitespace(" ", pos11)},
					declaration.Prelude...)
			}
			contents, err := PreprocessDeclarationsPrelude(baseURL, pa.ParseBlocksContents(declaration.Content, false),
				declarationPrelude)
			if err != nil {
				return nil, err
			}
			out = append(out, contents...)
		}

		declaration, ok := declaration.(pa.Declaration)
		if !ok {
			continue
		}

		name := declaration.Name
		if !strings.HasPrefix(name, "--") { // check for non variable, case insensitive
			name = utils.AsciiLower(declaration.Name)
		}

		validationError := func(reason string) {
			logger.WarningLogger.Printf("Ignored `%s:%s` , %s. \n", declaration.Name, pa.Serialize(declaration.Value), reason)
		}

		if _, in := notPrintMedia[name]; in {
			validationError("the property does not apply for the print media")
			continue
		}

		if strings.HasPrefix(name, proprietaryPrefix) {
			unprefixedName := strings.TrimPrefix(name, proprietaryPrefix)
			if _, in := proprietary[unprefixedName]; in {
				name = unprefixedName
			} else if _, in := unstable[unprefixedName]; in {
				logger.WarningLogger.Printf("Deprecated `%s:%s`, prefixes on unstable attributes are deprecated, use `%s` instead. \n",
					declaration.Name, pa.Serialize(declaration.Value), unprefixedName)
				name = unprefixedName
			} else {
				logger.WarningLogger.Printf("Ignored `%s:%s`,prefix on this attribute is not supported, use `%s` instead. \n",
					declaration.Name, pa.Serialize(declaration.Value), unprefixedName)
				continue
			}
		}

		if strings.HasPrefix(name, "-") && !strings.HasPrefix(name, "--") {
			validationError("prefixed selectors are ignored")
			continue
		}

		tokens := pa.RemoveWhitespace(declaration.Value)

		// Having no tokens is allowed by grammar but refused by all
		// properties and expanders.
		if len(tokens) == 0 {
			validationError("no value")
			continue
		}

		var (
			result expandedProperties
			err    error
		)
		if sh := pr.NewShortand(name); sh != 0 {
			result, err = expanders[sh](baseURL, sh, tokens)
		} else {
			// validate without any expansion
			var r namedProperty
			r, err = validateNonShorthand(baseURL, name, tokens, false)
			result = append(result, r)
		}

		if err != nil {
			validationError(err.Error())
			continue
		}

		important := declaration.Important

		for _, np := range result {
			ownDecls = append(ownDecls, Declaration{
				Name:      np.name,
				Value:     np.property,
				Important: important,
				Shortand:  np.shortand,
			})
		}
	}

	out = append(out, KeyedDeclarations{selectors, ownDecls})

	return out, nil
}

// If `token` is [Ident], return its lower name.
// Otherwise return empty string.
func getKeyword(token Token) string {
	if ident, ok := token.(pa.Ident); ok {
		return utils.AsciiLower(ident.Value)
	}
	return ""
}

// If `tokens` is a 1-element list of [Ident], return its name.
// Otherwise return empty string.
func getSingleKeyword(tokens []Token) string {
	if len(tokens) == 1 {
		return getKeyword(tokens[0])
	}
	return ""
}

// negative  = true, percentage = false
func getLength(token Token, negative, percentage bool) pr.Dimension {
	switch token := token.(type) {
	case pa.Percentage:
		if percentage && (negative || token.ValueF >= 0) {
			return pr.PercToD(token.ValueF)
		}
	case pa.Dimension:
		unit, isKnown := LENGTHUNITS[string(token.Unit)]
		if isKnown && (negative || token.ValueF >= 0) {
			return pr.NewDim(pr.Float(token.ValueF), unit)
		}
	case pa.Number:
		if token.ValueF == 0 {
			return pr.NewDim(0, pr.Scalar)
		}
	}
	return pr.Dimension{}
}

// Return the value in radians of an <angle> token, or None.
func getAngle(token Token) (utils.Fl, bool) {
	if dim, ok := token.(pa.Dimension); ok {
		unit, in := AngleUnits[string(dim.Unit)]
		if in {
			return dim.ValueF * ANGLETORADIANS[unit], true
		}
	}
	return 0, false
}

// Return the value in dppx of a <resolution> token, or false.
func getResolution(token Token) (utils.Fl, bool) {
	if dim, ok := token.(pa.Dimension); ok {
		factor, in := RESOLUTIONTODPPX[string(dim.Unit)]
		if in {
			return dim.ValueF * factor, true
		}
	}
	return 0, false
}

// @validator()
// @commaSeparatedList
// @singleKeyword
// “background-attachment“ property validation.
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
	for _, part := range pa.SplitOnComma(tokens) {
		part = pa.RemoveWhitespace(part)
		result := _backgroundAttachment(part)
		if result == "" {
			return nil
		}
		out = append(out, result)
	}
	return out
}

// @validator("background-color")
// @validator("border-top-color")
// @validator("border-right-color")
// @validator("border-bottom-color")
// @validator("border-left-color")
// @validator("column-rule-color")
// @singleToken
func otherColors(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		c := pa.ParseColor(tokens[0])
		if !c.IsNone() {
			return pr.Color(c)
		}
	}
	return nil
}

// @validator()
// @singleToken
func outlineColor(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		token := tokens[0]
		if getKeyword(token) == "invert" {
			return pr.Color{Type: pa.ColorCurrentColor}
		} else {
			return pr.Color(pa.ParseColor(token))
		}
	}
	return nil
}

// @validator()
// @singleKeyword
func borderCollapse(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "separate", "collapse":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “empty-cells“ property validation.
func emptyCells(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "show", "hide":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator("color")
// @singleToken
// “*-color“ && “color“ properties validation.
func color(tokens []Token, _ string) pr.DeclaredValue {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	result := pa.ParseColor(token)
	if result.Type == pa.ColorCurrentColor {
		return pr.Inherit
	} else {
		return pr.Color(result)
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

	if getKeyword(token) == "none" {
		return pr.NoneImage{}, nil
	}
	return getImage(token, baseUrl)
}

func backgroundImage(tokens []Token, baseUrl string) (pr.CssProperty, error) {
	var out pr.Images
	for _, part := range pa.SplitOnComma(tokens) {
		part = pa.RemoveWhitespace(part)
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

// @validator("list-style-image", wantsBaseUrl=true)
// @singleToken
// “list-style-image“ property validation.
func listStyleImage(tokens []Token, baseUrl string) (pr.CssProperty, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]

	if token.Kind() != pa.KFunctionBlock {
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

var centerKeywordFakeToken = pa.NewIdent("center", pa.Pos{})

// @validator(unstable=true)
func transformOrigin(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 3 {
		// Ignore third parameter as 3D transforms are ignored.
		tokens = tokens[:2]
	}
	return parse2dPosition(tokens)
}

// @validator()
// @commaSeparatedList
// “background-position“ and “object-position“ property validation.
// See http://drafts.csswg.org/csswg/css-backgrounds-3/#the-background-position
func backgroundPosition(tokens []Token, _ string) pr.CssProperty {
	out := centers(tokens)

	if len(out) == 0 {
		return nil
	}

	return out
}

func centers(tokens []Token) pr.Centers {
	var out pr.Centers
	for _, part := range pa.SplitOnComma(tokens) {
		result := parsePosition(pa.RemoveWhitespace(part))
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

// @validator()
// @commaSeparatedList
// “background-repeat“ property validation.
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
	for _, part := range pa.SplitOnComma(tokens) {
		result := _backgroundRepeat(pa.RemoveWhitespace(part))
		if result == [2]string{} {
			return nil
		}
		out = append(out, result)
	}
	return out
}

// @validator()
// @commaSeparatedList
// Validation for “background-size“.
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
			return pr.Size{Width: length.ToValue(), Height: pr.SToV("auto")}
		}
	case 2:
		var out pr.Size
		lengthW := getLength(tokens[0], false, true)
		lengthH := getLength(tokens[1], false, true)
		if !lengthW.IsNone() {
			out.Width = lengthW.ToValue()
		} else if getKeyword(tokens[0]) == "auto" {
			out.Width = pr.SToV("auto")
		} else {
			return pr.Size{}
		}
		if !lengthH.IsNone() {
			out.Height = lengthH.ToValue()
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
	for _, part := range pa.SplitOnComma(tokens) {
		result := _backgroundSize(pa.RemoveWhitespace(part))
		if (result == pr.Size{}) {
			return nil
		}
		out = append(out, result)
	}
	return out
}

// @validator("background-clip")
// @validator("background-origin")
// @commaSeparatedList
// @singleKeyword
// Validation for the “<box>“ type used in “background-clip“
//
//	and ``background-origin``.
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
	for _, part := range pa.SplitOnComma(tokens) {
		result := _box(pa.RemoveWhitespace(part))
		if result == "" {
			return nil
		}
		out = append(out, result)
	}
	return out
}

func borderDims(tokens []Token, negative, percentage bool) pr.CssProperty {
	lengths := make([]pr.Dimension, len(tokens))
	allLengths := true
	for index, token := range tokens {
		lengths[index] = getLength(token, negative, percentage)
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

// @validator()
// Validator for the `border-spacing` property.
func borderSpacing(tokens []Token, _ string) pr.CssProperty {
	return borderDims(tokens, true, false)
}

// @validator("border-top-right-radius")
// @validator("border-bottom-right-radius")
// @validator("border-bottom-left-radius")
// @validator("border-top-left-radius")
// Validator for the `border-*-radius` pr.
func borderCornerRadius(tokens []Token, _ string) pr.CssProperty {
	return borderDims(tokens, false, true)
}

// @validator("border-top-style")
// @validator("border-right-style")
// @validator("border-left-style")
// @validator("border-bottom-style")
// @validator("column-rule-style")
// @singleKeyword
// “border-*-style“ properties validation.
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

// @validator("break-before")
// @validator("break-after")
// @singleKeyword
// “break-before“ && “break-after“ properties validation.
func breakBeforeAfter(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "avoid", "avoid-page", "page", "left", "right",
		"recto", "verso", "avoid-column", "column", "always":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “break-inside“ property validation.
func breakInside(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "avoid", "avoid-page", "avoid-column":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “box-decoration-break“ property validation.
func boxDecorationBreak(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "slice", "clone":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “margin-break“ property validation.
func marginBreak(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "keep", "discard":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator(unstable=true)
// @singleToken
// “page“ property validation.
func page(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if ident, ok := token.(pa.Ident); ok {
		if utils.AsciiLower(ident.Value) == "auto" {
			return pr.Page("auto")
		}
		return pr.Page(ident.Value)
	}
	return nil
}

// @validator("bleed-left")
// @validator("bleed-right")
// @validator("bleed-top")
// @validator("bleed-bottom")
// @singleToken
// “bleed“ property validation.
func bleed(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "auto" {
		return pr.DimOrS{S: "auto"}
	} else {
		return getLength(token, true, false).ToValue()
	}
}

// @validator()
// “marks“ property validation.
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

// @validator("outline-style")
// @singleKeyword
// “outline-style“ properties validation.
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

// @validator("border-top-width")
// @validator("border-right-width")
// @validator("border-left-width")
// @validator("border-bottom-width")
// @validator("column-rule-width")
// @validator("outline-width")
// @singleToken
// Border, column rule && outline widths properties validation.
func borderWidth(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, false)
	if !length.IsNone() {
		return length.ToValue()
	}
	keyword := getKeyword(token)
	if keyword == "thin" || keyword == "medium" || keyword == "thick" {
		return pr.DimOrS{S: keyword}
	}
	return nil
}

// @single_token
func borderImageSource(tokens []Token, baseURL string) (pr.CssProperty, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]
	if getKeyword(token) == "none" {
		return pr.NoneImage{}, nil
	}
	return getImage(token, baseURL)
}

func borderImageSlice(tokens []Token, _ string) pr.CssProperty {
	var (
		values pr.Values
		fill   bool
	)
	for i, token := range tokens {
		// Don't use get_length() because a dimension with a unit is disallowed.
		if v, ok := token.(pa.Percentage); ok && v.ValueF >= 0 {
			values = append(values, pr.PercToV(v.ValueF))
		} else if v, ok := token.(pa.Number); ok && v.ValueF >= 0 {
			values = append(values, pr.FToV(v.ValueF))
		} else if getKeyword(token) == "fill" && !fill && (i == 0 || i == len(tokens)-1) {
			fill = true
			values = append(values, pr.SToV("fill"))
		} else {
			return nil
		}
	}
	if L := len(values); (fill && 2 <= L && L <= 5) || (1 <= L && L <= 4) {
		return values
	}
	return nil
}

func borderImageWidth(tokens []Token, _ string) pr.CssProperty {
	var values pr.Values
	for _, token := range tokens {
		if getKeyword(token) == "auto" {
			values = append(values, pr.SToV("auto"))
		} else if v, ok := token.(pa.Number); ok && v.ValueF >= 0 {
			values = append(values, pr.FToV(v.ValueF))
		} else {
			if length := getLength(token, false, true); !length.IsNone() {
				values = append(values, length.ToValue())
			} else {
				return nil
			}
		}
	}

	if L := len(values); 1 <= L && L <= 4 {
		return values
	}
	return nil
}

func borderImageOutset(tokens []Token, _ string) pr.CssProperty {
	var values pr.Values
	for _, token := range tokens {
		if v, ok := token.(pa.Number); ok && v.ValueF >= 0 {
			values = append(values, pr.FToV(v.ValueF))
		} else {
			if length := getLength(token, false, false); !length.IsNone() {
				values = append(values, length.ToValue())
			} else {
				return nil
			}
		}
	}
	if L := len(values); 1 <= L && L <= 4 {
		return values
	}
	return nil
}

func borderImageRepeat(tokens []Token, _ string) pr.CssProperty {
	if L := len(tokens); 1 <= L && L <= 2 {
		var keywords pr.Strings
		for _, token := range tokens {
			switch k := getKeyword(token); k {
			case "stretch", "repeat", "round", "space":
				keywords = append(keywords, k)
			default:
				return nil
			}
		}
		return keywords
	}
	return nil
}

// @validator()
// @singleToken
// “column-width“ property validation.
func columnWidth(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, false)
	if !length.IsNone() {
		return length.ToValue()
	}
	keyword := getKeyword(token)
	if keyword == "auto" {
		return pr.DimOrS{S: keyword}
	}
	return nil
}

// @validator()
// @singleKeyword
// “column-span“ property validation.
func columnSpan(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "all", "none":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// Validation for the “box-sizing“ property from css3-ui
func boxSizing(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "padding-box", "border-box", "content-box":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “caption-side“ properties validation.
func captionSide(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "top", "bottom":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “clear“ property validation.
func clear(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "left", "right", "both", "none":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleToken
// Validation for the “clip“ property.
func clip(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	name, args := pa.ParseFunction(token)
	if name != "" {
		if name == "rect" && len(args) == 4 {
			var values pr.Values
			for _, arg := range args {
				if getKeyword(arg) == "auto" {
					values = append(values, pr.DimOrS{S: "auto"})
				} else {
					length := getLength(arg, true, false)
					if !length.IsNone() {
						values = append(values, length.ToValue())
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

// @validator(wantsBaseUrl=true)
// “content“ property validation.
func content(tokens []Token, baseUrl string) (pr.CssProperty, error) {
	var token Token
	for len(tokens) > 0 {
		if len(tokens) >= 2 {
			if lit, ok := tokens[1].(pa.Literal); ok && lit.Value == "," {
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
		lit, ok := tokens[len(tokens)-2].(pa.Literal)
		if tokens[len(tokens)-1].Kind() == pa.KString && ok && lit.Value == "/" {
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

// @validator()
// “counter-increment“ property validation.
func counterIncrement(tokens []Token, _ string) (pr.CssProperty, error) {
	ci, err := counter(tokens, 1)
	if err != nil || ci == nil {
		return nil, err
	}
	return pr.SIntStrings{Values: ci}, nil
}

// “counter-reset“ property validation.
// “counter-set“ property validation.
func counterReset(tokens []Token, _ string) (pr.CssProperty, error) {
	ci, err := counter(tokens, 0)
	if err != nil || ci == nil {
		return nil, err
	}
	return pr.SIntStrings{Values: ci}, err
}

// “counter-increment“ && “counter-reset“ properties validation.
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
	iter := pa.NewIter(tokens)
	token := iter.Next()
	for token != nil {
		ident, ok := token.(pa.Ident)
		if !ok {
			return nil, nil // expected a keyword here
		}
		counterName := ident.Value
		if counterName == "none" || counterName == "initial" || counterName == "inherit" {
			return nil, fmt.Errorf("invalid counter name: %s", counterName)
		}
		token = iter.Next()
		if number, ok := token.(pa.Number); ok && number.IsInt() { // implies token != nil
			// Found an integer. Use it and get the next token
			integer = number.Int()
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

// @validator("top")
// @validator("right")
// @validator("left")
// @validator("bottom")
// @validator("margin-top")
// @validator("margin-right")
// @validator("margin-bottom")
// @validator("margin-left")
// @singleToken
// “margin-*“ properties validation.
func lengthPercOrAuto(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, true, true)
	if !length.IsNone() {
		return length.ToValue()
	}
	if getKeyword(token) == "auto" {
		return pr.DimOrS{S: "auto"}
	}
	return nil
}

// @validator("height")
// @validator("width")
// @singleToken
// Validation for the “width“ && “height“ pr.
func widthHeight(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, true)
	if !length.IsNone() {
		return length.ToValue()
	}
	if getKeyword(token) == "auto" {
		return pr.DimOrS{S: "auto"}
	}
	return nil
}

// Validation for the “column-gap“ and "row-gap" property.
func gap(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, false)
	if !length.IsNone() {
		return length.ToValue()
	}
	if getKeyword(token) == "normal" {
		return pr.DimOrS{S: "normal"}
	}
	return nil
}

// @validator()
// @singleKeyword
// “column-fill“ property validation.
func columnFill(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "balance":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “direction“ property validation.
func direction(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "ltr", "rtl":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “display“ property validation.
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
		ident, ok := token.(pa.Ident)
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

// @validator("float")
// @singleKeyword
// “float“ property validation.
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
	if tt, ok := tokens[0].(pa.String); len(tokens) == 1 && ok {
		return tt.Value
	} else if len(tokens) > 0 {
		var values []string
		for _, token := range tokens {
			if tt, ok := token.(pa.Ident); ok {
				values = append(values, string(tt.Value))
			} else {
				return ""
			}
		}
		return strings.Join(values, " ")
	}
	return ""
}

// @validator()
// @commaSeparatedList
// “font-family“ property validation.
func fontFamily(tokens []Token, _ string) pr.CssProperty {
	var out pr.Strings
	for _, part := range pa.SplitOnComma(tokens) {
		result := _fontFamily(pa.RemoveWhitespace(part))
		if result == "" {
			return nil
		}
		out = append(out, result)
	}
	return out
}

// @validator()
// @singleKeyword
func fontKerning(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "normal", "none":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleToken
func fontLanguageOverride(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "normal" {
		return pr.String(keyword)
	}
	if tt, ok := token.(pa.String); ok {
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
		ident, isIdent := token.(pa.Ident)
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

// @validator()
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

// @validator()
// @singleKeyword
func fontVariantPosition(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "sub", "super":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
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

// @validator()
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

// @validator()
// “font-feature-settings“ property validation.
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
			case pa.Ident:
				if tt.Value == "on" {
					value = 1
				} else {
					value = 0
				}
			case pa.Number:
				if tt.IsInt() && tt.Int() >= 0 {
					value = tt.Int()
				}
			}
		} else if len(tokens) == 1 {
			value = 1
		}

		if len(tokens) == 1 {
			token = tokens[0]
			tt, ok := token.(pa.String)
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
	for _, part := range pa.SplitOnComma(tokens) {
		result := fontFeatureSettingsList(pa.RemoveWhitespace(part))
		if (result == pr.IntString{}) {
			return pr.SIntStrings{}
		}
		out.Values = append(out.Values, result)
	}
	return out
}

// @validator()
// @singleKeyword
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

// @validator()
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

// @property()
// “font-variation-settings“ property validation.
func fontVariationSettings(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 && getKeyword(tokens[0]) == "normal" {
		return pr.SFloatStrings{String: "normal"}
	}
	// @comma_separated_list
	fontVariationSettingsList := func(tokens []Token) (pr.FloatString, bool) {
		if len(tokens) == 2 {
			keyV, ok1 := tokens[0].(pa.String)
			valueV, ok2 := tokens[1].(pa.Number)
			if ok1 && ok2 {
				return pr.FloatString{String: keyV.Value, Float: valueV.ValueF}, true
			}
		}
		return pr.FloatString{}, false
	}

	var out pr.SFloatStrings
	for _, part := range pa.SplitOnComma(tokens) {
		result, ok := fontVariationSettingsList(pa.RemoveWhitespace(part))
		if !ok {
			return nil
		}
		out.Values = append(out.Values, result)
	}
	return out
}

// @validator()
// @singleToken
// “font-size“ property validation.
func fontSize(tokens []Token, _ string) (pr.CssProperty, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]
	length := getLength(token, false, true)
	if !length.IsNone() {
		return length.ToValue(), nil
	}
	fontSizeKeyword := getKeyword(token)
	if _, isIn := pr.FontSizeKeywords[fontSizeKeyword]; isIn || fontSizeKeyword == "smaller" || fontSizeKeyword == "larger" {
		return pr.DimOrS{S: fontSizeKeyword}, nil
	}
	return nil, nil
}

// @validator()
// @singleKeyword
// “font-style“ property validation.
func fontStyle(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "italic", "oblique":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// Validation for the “font-stretch“ property.
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

// @validator()
// @singleToken
// “font-weight“ property validation.
func fontWeight(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "normal" || keyword == "bold" || keyword == "bolder" || keyword == "lighter" {
		return pr.IntString{String: keyword}
	}
	if number, ok := token.(pa.Number); ok {
		intValue := number.Int()
		if number.IsInt() && (intValue == 100 || intValue == 200 || intValue == 300 || intValue == 400 || intValue == 500 || intValue == 600 || intValue == 700 || intValue == 800 || intValue == 900) {
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

// @validator(unstable=true)
// @singleToken
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

// @validator("letter-spacing")
// @validator("word-spacing")
// @singleToken
// Validation for “letter-spacing“ && “word-spacing“.
func spacing(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "normal" {
		return pr.DimOrS{S: "normal"}
	}
	length := getLength(token, true, false)
	if !length.IsNone() {
		return length.ToValue()
	}
	return nil
}

// @validator()
// @singleToken
// “line-height“ property validation.
func lineHeight(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "normal" {
		return pr.DimOrS{S: "normal"}
	}

	switch tt := token.(type) {
	case pa.Number:
		if tt.ValueF >= 0 {
			return pr.NewDim(pr.Float(tt.ValueF), pr.Scalar).ToValue()
		}
	case pa.Percentage:
		if tt.ValueF >= 0 {
			return pr.PercToV(tt.ValueF)
		}
	case pa.Dimension:
		if tt.ValueF >= 0 {
			l := getLength(token, true, false)
			if l.IsNone() {
				return nil
			}
			return l.ToValue()
		}
	}
	return nil
}

// @validator()
// @singleKeyword
// “list-style-position“ property validation.
func listStylePosition(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "inside", "outside":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleToken
// “list-style-type“ property validation.
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
	case pa.Ident:
		return pr.CounterStyleID{Name: string(token.Value)}, true
	case pa.String:
		return pr.CounterStyleID{Type: "string", Name: token.Value}, true
	case pa.FunctionBlock:
		if token.Name != "symbols" {
			return out, false
		}
		functionArguments := pa.RemoveWhitespace(token.Arguments)
		if len(functionArguments) == 0 {
			return out, false
		}
		arguments := []string{"symbolic"}
		if arg0, ok := functionArguments[0].(pa.Ident); ok {
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
			if str, ok := arg.(pa.String); ok {
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
// “min-width“ && “min-height“ properties validation.
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

// @validator("padding-top")
// @validator("padding-right")
// @validator("padding-bottom")
// @validator("padding-left")
// @singleToken
// “padding-*“ properties validation.
func lengthOrPercentage(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	l := getLength(token, false, true)
	if l.IsNone() {
		return nil
	}
	return l.ToValue()
}

// @validator("max-width")
// @validator("max-height")
// @singleToken
// Validation for max-width && max-height
func maxWidthHeight(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, true)
	if !length.IsNone() {
		return length.ToValue()
	}
	if getKeyword(token) == "none" {
		return pr.NewDim(pr.Inf, pr.Px).ToValue()
	}
	return nil
}

// @validator()
// @singleToken
// Validation for the “opacity“ property.
func opacity(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(pa.Number); ok {
		return pr.Float(utils.MinF(1, utils.MaxF(0, number.ValueF)))
	} else if perc, ok := token.(pa.Percentage); ok {
		return pr.Float(utils.MinF(1, utils.MaxF(0, perc.ValueF/100)))
	}

	return nil
}

// @validator()
// @singleToken
// Validation for the “z-index“ property.
func zIndex(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "auto" {
		return pr.IntString{String: "auto"}
	}
	if number, ok := token.(pa.Number); ok {
		if number.IsInt() {
			return pr.IntString{Int: number.Int()}
		}
	}
	return nil
}

// @validator("orphans")
// @validator("widows")
// @singleToken
// Validation for the “orphans“ && “widows“ pr.
func orphansWidows(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(pa.Number); ok {
		value := number.Int()
		if number.IsInt() && value >= 1 {
			return pr.Int(value)
		}
	}
	return nil
}

// @validator()
// @singleToken
// Validation for the “column-count“ property.
func columnCount(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(pa.Number); ok {
		value := number.Int()
		if number.IsInt() && value >= 1 {
			return pr.IntString{Int: value}
		}
	}
	if getKeyword(token) == "auto" {
		return pr.IntString{String: "auto"}
	}
	return nil
}

// @validator()
// @singleKeyword
// Validation for the “overflow“ property.
func overflow(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "visible", "hidden", "scroll":
		return pr.String(keyword)
	default:
		return nil
	}
}

// Validation for the “word-break“ property.
func wordBreak(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "break-all":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @single_keyword
// Validation for the “text-overflow“ property.
func textOverflow(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "clip", "ellipsis":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleToken
// “position“ property validation.
func position(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if fn, ok := token.(pa.FunctionBlock); ok && fn.Name == "running" && len(fn.Arguments) == 1 {
		if ident, ok := (fn.Arguments)[0].(pa.Ident); ok {
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

// @validator()
// “quotes“ property validation.
func quotes(tokens []Token, _ string) pr.CssProperty {
	var opens, closes []string
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		switch keyword {
		case "auto":
			return pr.Quotes{Tag: pr.Auto}
		case "none":
			return pr.Quotes{Tag: pr.None}
		}
	}
	if len(tokens) > 0 && len(tokens)%2 == 0 {
		// Separate open && close quotes.
		// eg.  ("«", "»", "“", "”")  -> (("«", "“"), ("»", "”"))
		for i := 0; i < len(tokens); i += 2 {
			open, ok1 := tokens[i].(pa.String)
			close_, ok2 := tokens[i+1].(pa.String)
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

// @validator()
// @singleKeyword
// Validation for the “table-layout“ property
func tableLayout(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "fixed", "auto":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “text-align-all“ property validation.
func textAlignAll(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "left", "right", "center", "justify", "start", "end":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “text-align-last“ property validation.
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
// “text-decoration-line“ property validation.
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
// “text-decoration-style“ property validation.
func textDecorationStyle(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "solid", "double", "dotted", "dashed", "wavy":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleToken
// “text-indent“ property validation.
func textIndent(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	l := getLength(token, true, true)
	if l.IsNone() {
		return nil
	}
	return l.ToValue()
}

// @validator()
// @singleKeyword
// “text-align“ property validation.
func textTransform(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "none", "uppercase", "lowercase", "capitalize", "full-width":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleToken
// Validation for the “vertical-align“ property
func verticalAlign(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, true, true)
	if !length.IsNone() {
		return length.ToValue()
	}
	keyword := getKeyword(token)
	if keyword == "baseline" || keyword == "middle" || keyword == "sub" || keyword == "super" || keyword == "text-top" || keyword == "text-bottom" || keyword == "top" || keyword == "bottom" {
		return pr.DimOrS{S: keyword}
	}
	return nil
}

// @validator()
// @singleKeyword
// “visibility“ property validation.
func visibility(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "visible", "hidden", "collapse":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “white-space“ property validation.
func whiteSpace(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "pre", "nowrap", "pre-wrap", "pre-line":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// “overflow-wrap“ property validation.
func overflowWrap(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "anywhere", "normal", "break-word":
		return pr.String(keyword)
	default:
		return nil
	}
}

// Validation for “footnote-display“.
func footnoteDisplay(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "block", "inline", "compact":
		return pr.String(keyword)
	default:
		return nil
	}
}

// Validation for “footnote-policy“.
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
// “flex-basis“ property validation.
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
// “flex-direction“ property validation.
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
	if f, ok := _flexGrowShrink(tokens); ok {
		return pr.Float(f)
	}
	return nil
}

func _flexGrowShrink(tokens []Token) (pr.Fl, bool) {
	if len(tokens) != 1 {
		return 0, false
	}
	token := tokens[0]
	if number, ok := token.(pa.Number); ok {
		return number.ValueF, true
	}
	return 0, false
}

// Parse “inflexible-breadth“.
func parseInflexibleBreadth(token Token) pr.DimOrS {
	keyword := getKeyword(token)
	switch keyword {
	case "auto", "min-content", "max-content":
		return pr.DimOrS{S: keyword}
	case "":
		length := getLength(token, false, true)
		if !length.IsNone() {
			return length.ToValue()
		}
	}

	return pr.DimOrS{}
}

// Parse “track-breadth“.
func parseTrackBreadth(token Token) pr.DimOrS {
	if dim, ok := token.(pa.Dimension); ok && dim.ValueF >= 0 && dim.Unit == "fr" {
		return pr.NewDim(pr.Float(dim.ValueF), pr.Fr).ToValue()
	}
	return parseInflexibleBreadth(token)
}

// Parse “track-size“.
func parseTrackSize(token Token) pr.GridDims {
	trackBreadth := parseTrackBreadth(token)
	if !trackBreadth.IsNone() {
		return pr.NewGridDimsValue(trackBreadth)
	}
	name, args := pa.ParseFunction(token)
	if name == "minmax" {
		if len(args) == 2 {
			inflexibleBreadth := parseInflexibleBreadth(args[0])
			trackBreadth := parseTrackBreadth(args[1])
			if !inflexibleBreadth.IsNone() && !trackBreadth.IsNone() {
				return pr.NewGridDimsMinmax(inflexibleBreadth, trackBreadth)
			}
		}
	} else if name == "fit-content" {
		if len(args) == 1 {
			length := getLength(args[0], false, true)
			if !length.IsNone() {
				return pr.NewGridDimsFitcontent(length)
			}
		}
	}

	return pr.GridDims{}
}

// Parse “fixed-size“.
func parseFixedSize(token Token) pr.GridDims {
	length := getLength(token, false, true)
	if !length.IsNone() {
		return pr.NewGridDimsValue(length.ToValue())
	}
	name, args := pa.ParseFunction(token)
	if name == "minmax" && len(args) == 2 {
		length := getLength(args[0], false, true)
		if !length.IsNone() {
			trackBreadth := parseTrackBreadth(args[1])
			if !trackBreadth.IsNone() {
				return pr.NewGridDimsMinmax(length.ToValue(), trackBreadth)
			}
		}
		keyword := getKeyword(args[0])
		if keyword == "min-content" || keyword == "max-content" || keyword == "auto" || !length.IsNone() {
			fixedBreadth := getLength(args[1], false, true)
			if !fixedBreadth.IsNone() {
				v1 := length.ToValue()
				if v1.IsNone() {
					v1 = pr.SToV(keyword)
				}
				return pr.NewGridDimsMinmax(v1, fixedBreadth.ToValue())
			}
		}
	}

	return pr.GridDims{}
}

// parse “line-names“, returning nil if invalid,
// but an empty list for '[]'
func parseLineNames(arg Token) []string {
	if arg, ok := arg.(pa.SquareBracketsBlock); ok {
		names := []string{}
		for _, token := range arg.Arguments {
			if ident, ok := token.(pa.Ident); ok {
				names = append(names, ident.Value)
			} else if _, ok := token.(pa.Whitespace); ok {
				continue
			} else {
				return nil
			}
		}
		return names
	}

	return nil
}

// @property("grid-auto-columns")
// @property("grid-auto-rows")
// “grid-auto-columns“ and “grid-auto-rows“ properties validation.
func gridAuto(tokens []Token, _ string) pr.CssProperty {
	var returnTokens pr.GridAuto
	for _, token := range tokens {
		trackSize := parseTrackSize(token)
		if trackSize.IsNone() {
			return nil
		}
		returnTokens = append(returnTokens, trackSize)
	}
	return returnTokens
}

// @property()
// “grid-auto-flow“ property validation.
func gridAutoFlow(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		switch keyword {
		case "row", "column":
			return pr.Strings{keyword}
		case "dense":
			return pr.Strings{keyword, "row"}
		}
	} else if len(tokens) == 2 {
		keywords := [2]string{getKeyword(tokens[0]), getKeyword(tokens[1])}
		switch keywords {
		case [2]string{"dense", "row"}, [2]string{"dense", "column"}, [2]string{"row", "dense"}, [2]string{"column", "dense"}:
			return pr.Strings(keywords[:])
		}
	}
	return nil
}

func parseRepeat(token Token, acceptAutoFit bool) (int, bool) {
	if nb, ok := token.(pa.Number); ok && nb.IsInt() && nb.ValueF >= 1 {
		return nb.Int(), true
	}
	switch getKeyword(token) {
	case "auto-fill":
		return pr.RepeatAutoFill, true
	case "auto-fit":
		if acceptAutoFit {
			return pr.RepeatAutoFit, true
		}
	}
	return 0, false
}

// [tokens] start right after 'subgrid'
func parseSubgrid(tokens []Token) ([]pr.GridSpec, bool) {
	var subgrid []pr.GridSpec
	for _, token := range tokens {
		lineNames := parseLineNames(token)
		if lineNames != nil {
			subgrid = append(subgrid, pr.GridNames(lineNames))
			continue
		}

		name, args := pa.ParseFunction(token)
		if !(name == "repeat" && len(args) >= 2) {
			return nil, false
		}

		reapeat, ok := parseRepeat(args[0], false)
		if !ok { // invalid
			return nil, false
		}
		var names [][]string
		for _, arg := range args[1:] {
			lineNames := parseLineNames(arg)
			if lineNames != nil {
				names = append(names, lineNames)
			}
		}
		subgrid = append(subgrid, pr.GridNameRepeat{Repeat: reapeat, Names: names})
	}
	return subgrid, true
}

// @property("grid-template-columns")
// @property("grid-template-rows")
// “grid-template-columns“ and “grid-template-rows“ validation.
func gridTemplate(tokens []Token, _ string) pr.CssProperty {
	if v, ok := gridTemplateImpl(tokens); ok {
		return v
	}
	return nil
}

func gridTemplateImpl(tokens []Token) (out pr.GridTemplate, _ bool) {
	if len(tokens) == 0 {
		return out, false
	}
	if len(tokens) == 1 && getKeyword(tokens[0]) == "none" {
		return pr.GridTemplate{Tag: pr.None}, true
	}
	if getKeyword(tokens[0]) == "subgrid" {
		if subgrid, ok := parseSubgrid(tokens[1:]); ok {
			return pr.GridTemplate{Tag: pr.Subgrid, Names: subgrid}, true
		}
		return out, false
	}

	var (
		returnTokens       []pr.GridSpec
		includesAutoRepeat = false
		includesTrack      = false
		lastIsLineName     = false
	)
	for _, token := range tokens {
		lineNames := parseLineNames(token)
		if lineNames != nil {
			if lastIsLineName {
				return out, false
			}
			lastIsLineName = true
			returnTokens = append(returnTokens, pr.GridNames(lineNames))
			continue
		}
		fixedSize := parseFixedSize(token)
		if !fixedSize.IsNone() {
			if !lastIsLineName {
				returnTokens = append(returnTokens, pr.GridNames{})
			}
			lastIsLineName = false
			returnTokens = append(returnTokens, fixedSize)
			continue
		}
		trackSize := parseTrackSize(token)
		if !trackSize.IsNone() {
			if !lastIsLineName {
				returnTokens = append(returnTokens, pr.GridNames{})
			}
			lastIsLineName = false
			returnTokens = append(returnTokens, trackSize)
			includesTrack = true
			continue
		}
		name, args := pa.ParseFunction(token)
		if name == "repeat" && len(args) >= 2 {
			number, ok := parseRepeat(args[0], true)
			if !ok {
				return out, false
			}
			if number <= -1 { // auto-repeat
				if includesAutoRepeat {
					return out, false
				}
				includesAutoRepeat = true
			}

			var (
				namesAndSizes        []pr.GridSpec
				repeatLastIsLineName = false
			)
			for _, arg := range args[1:] {
				lineNames = parseLineNames(arg)
				if lineNames != nil {
					if repeatLastIsLineName {
						return out, false
					}
					namesAndSizes = append(namesAndSizes, pr.GridNames(lineNames))
					repeatLastIsLineName = true
					continue
				}
				// fixed-repead
				fixedSize = parseFixedSize(arg)
				if !fixedSize.IsNone() {
					if !repeatLastIsLineName {
						namesAndSizes = append(namesAndSizes, pr.GridNames{})
					}
					repeatLastIsLineName = false
					namesAndSizes = append(namesAndSizes, fixedSize)
					continue
				}
				// track-repeat
				trackSize = parseTrackSize(arg)
				if !trackSize.IsNone() {
					includesTrack = true
					if !repeatLastIsLineName {
						namesAndSizes = append(namesAndSizes, pr.GridNames{})
					}
					repeatLastIsLineName = false
					namesAndSizes = append(namesAndSizes, trackSize)
					continue
				}
				return out, false
			}
			if !lastIsLineName {
				returnTokens = append(returnTokens, pr.GridNames{})
			}
			lastIsLineName = false
			if !repeatLastIsLineName {
				namesAndSizes = append(namesAndSizes, pr.GridNames{})
			}
			returnTokens = append(returnTokens, pr.GridRepeat{Names: namesAndSizes, Repeat: number})
			continue
		}
		return out, false
	}
	if includesAutoRepeat && includesTrack {
		return out, false
	}
	if !lastIsLineName {
		returnTokens = append(returnTokens, pr.GridNames{})
	}
	return pr.GridTemplate{Names: returnTokens}, true
}

// @property()
// “grid-template-areas“ property validation.
func gridTemplateAreas(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 && getKeyword(tokens[0]) == "none" {
		return pr.GridTemplateAreas{}
	}
	var gridAreas pr.GridTemplateAreas
	for _, token := range tokens {
		s, ok := token.(pa.String)
		if !ok {
			return nil
		}
		componentValues := pa.Tokenize([]byte(s.Value), true)
		var (
			row       []string
			lastIsDot = false
		)
		for _, value := range componentValues {
			switch value := value.(type) {
			case pa.Ident:
				row = append(row, value.Value)
				lastIsDot = false
			case pa.Literal:
				if value.Value != "." {
					return nil
				}
				if lastIsDot {
					continue
				}
				row = append(row, "")
				lastIsDot = true
			case pa.Whitespace:
				lastIsDot = false
			default:
				return nil
			}
		}
		if len(row) == 0 {
			return nil
		}
		gridAreas = append(gridAreas, row)
	}

	// check row / column have the same sizes
	L := len(gridAreas[0])
	for _, other := range gridAreas {
		if len(other) != L {
			return nil
		}
	}
	// check areas are continuous rectangles
	coordinates := make(map[[2]int]bool)
	areas := make(map[string]bool)
	for y, row := range gridAreas {
		for x, area := range row {
			if in := coordinates[[2]int{x, y}]; in || area == "" {
				continue
			}
			if in := areas[area]; in {
				return nil
			}
			areas[area] = true
			coordinates[[2]int{x, y}] = true
			nx := x + 1
			for ; nx < len(row); nx++ {
				narea := row[nx]
				if narea != area {
					break
				}
				coordinates[[2]int{nx, y}] = true
			}
			for ny := y + 1; ny < len(gridAreas); ny++ {
				nrow := gridAreas[ny]
				if set := utils.NewSet(nrow[x:nx]...); len(set) == 1 && set.Has(area) {
					for nnx := x; nnx < nx; nnx++ {
						coordinates[[2]int{nnx, ny}] = true
					}
				} else {
					break
				}
			}
		}
	}
	return gridAreas
}

// @property("grid-row-start")
// @property("grid-row-end")
// @property("grid-column-start")
// @property("grid-column-end")
// “grid-[row|column]-[start—end]“ properties validation.
func gridLine(tokens []Token, _ string) pr.CssProperty {
	v, ok := gridLineImpl(tokens)
	if ok {
		return v
	}
	return nil
}

func gridLineImpl(tokens []Token) (pr.GridLine, bool) {
	if len(tokens) == 1 {
		token := tokens[0]
		if keyword := getKeyword(token); keyword != "" {
			if keyword == "auto" {
				return pr.GridLine{Tag: pr.Auto}, true
			} else if keyword != "span" {
				return pr.GridLine{Ident: keyword}, true
			}
		} else if number, ok := token.(pa.Number); ok && number.IsInt() && number.ValueF != 0 {
			return pr.GridLine{Val: number.Int()}, true
		}
		return pr.GridLine{}, false
	}
	var (
		number int
		ident  string
		span   pr.Tag
	)

	for _, token := range tokens {
		if keyword := getKeyword(token); keyword != "" {
			if keyword == "auto" {
				return pr.GridLine{}, false
			}
			if keyword == "span" {
				if span == 0 {
					span = pr.Span
					continue
				}
			} else if ident == "" {
				ident = keyword
				continue
			}
		} else if nbT, ok := token.(pa.Number); ok && nbT.IsInt() && nbT.ValueF != 0 {
			if number == 0 {
				number = nbT.Int()
				continue
			}
		}
		return pr.GridLine{}, false
	}
	if span != 0 {
		if number < 0 {
			return pr.GridLine{}, false
		} else if ident != "" || number != 0 {
			return pr.GridLine{Tag: span, Val: number, Ident: ident}, true
		}
	} else if number != 0 {
		return pr.GridLine{Tag: span, Val: number, Ident: ident}, true
	}

	return pr.GridLine{}, false
}

// @validator()
// @singleToken
func order(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(pa.Number); ok && number.IsInt() {
		return pr.Int(number.Int())
	}
	return nil
}

// @validator()
// @singleKeyword
// “flex-wrap“ property validation.
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
// “justify-content“ property validation.
func justifyContent(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		switch keyword := getKeyword(tokens[0]); keyword {
		case "center", "space-between", "space-around", "space-evenly",
			"stretch", "normal", "flex-start", "flex-end",
			"start", "end", "left", "right":
			return pr.Strings{keyword}
		}
	} else if len(tokens) == 2 {
		kw1, kw2 := getKeyword(tokens[0]), getKeyword(tokens[1])
		if kw1 == "safe" || kw1 == "unsafe" {
			switch kw2 {
			case "center", "start", "end", "flex-start", "flex-end", "left",
				"right":
				return pr.Strings{kw1, kw2}
			}
		}
	}
	return nil
}

// @validator()
// “align-items“ property validation.
func justifyItems(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		switch keyword := getKeyword(tokens[0]); keyword {
		case "normal", "stretch", "center", "start", "end", "self-start",
			"self-end", "flex-start", "flex-end", "left", "right",
			"legacy":
			return pr.Strings{keyword}
		case "baseline":
			return pr.Strings{"first", keyword}
		}
	} else if len(tokens) == 2 {
		kw1, kw2 := getKeyword(tokens[0]), getKeyword(tokens[1])
		if kw1 == "safe" || kw1 == "unsafe" {
			switch kw2 {
			case "center", "start", "end", "self-start", "self-end",
				"flex-start", "flex-end", "left", "right":
				return pr.Strings{kw1, kw2}
			}
		} else if kw1 == "baseline" {
			if kw2 == "first" || kw2 == "last" {
				return pr.Strings{kw1, kw2}
			}
		} else if kw2 == "baseline" {
			if kw1 == "first" || kw1 == "last" {
				return pr.Strings{kw1, kw2}
			}
		} else if kw1 == "legacy" {
			if kw2 == "left" || kw2 == "right" || kw2 == "center" {
				return pr.Strings{kw1, kw2}
			}
		} else if kw2 == "legacy" {
			if kw1 == "left" || kw1 == "right" || kw1 == "center" {
				return pr.Strings{kw1, kw2}
			}
		}
	}
	return nil
}

// @validator()
// “align-items“ property validation.
func justifySelf(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		switch keyword := getKeyword(tokens[0]); keyword {
		case "auto", "normal", "stretch", "center", "start", "end",
			"self-start", "self-end", "flex-start", "flex-end", "left",
			"right":
			return pr.Strings{keyword}
		case "baseline":
			return pr.Strings{"first", keyword}
		}
	} else if len(tokens) == 2 {
		kw1, kw2 := getKeyword(tokens[0]), getKeyword(tokens[1])
		if kw1 == "safe" || kw1 == "unsafe" {
			switch kw2 {
			case "center", "start", "end", "self-start", "self-end",
				"flex-start", "flex-end", "left", "right":
				return pr.Strings{kw1, kw2}
			}
		} else if kw1 == "baseline" {
			if kw2 == "first" || kw2 == "last" {
				return pr.Strings{kw1, kw2}
			}
		} else if kw2 == "baseline" {
			if kw1 == "first" || kw1 == "last" {
				return pr.Strings{kw1, kw2}
			}
		}
	}
	return nil
}

// @validator()
// “align-items“ property validation.
func alignItems(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		switch keyword := getKeyword(tokens[0]); keyword {
		case "normal", "stretch", "center", "start", "end", "self-start",
			"self-end", "flex-start", "flex-end":
			return pr.Strings{keyword}
		case "baseline":
			return pr.Strings{"first", keyword}
		}
	} else if len(tokens) == 2 {
		kw1, kw2 := getKeyword(tokens[0]), getKeyword(tokens[1])
		if kw1 == "safe" || kw1 == "unsafe" {
			switch kw2 {
			case "center", "start", "end", "self-start", "self-end",
				"flex-start", "flex-end":
				return pr.Strings{kw1, kw2}
			}
		} else if kw1 == "baseline" {
			if kw2 == "first" || kw2 == "last" {
				return pr.Strings{kw1, kw2}
			}
		} else if kw2 == "baseline" {
			if kw1 == "first" || kw1 == "last" {
				return pr.Strings{kw1, kw2}
			}
		}
	}
	return nil
}

// @validator()
// @singleKeyword
// “align-self“ property validation.
func alignSelf(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		switch keyword := getKeyword(tokens[0]); keyword {
		case "auto", "normal", "stretch", "center", "start", "end",
			"self-start", "self-end", "flex-start", "flex-end":
			return pr.Strings{keyword}
		case "baseline":
			return pr.Strings{"first", keyword}
		}
	} else if len(tokens) == 2 {
		kw1, kw2 := getKeyword(tokens[0]), getKeyword(tokens[1])
		if kw1 == "safe" || kw1 == "unsafe" {
			switch kw2 {
			case "center", "start", "end", "self-start", "self-end",
				"flex-start", "flex-end":
				return pr.Strings{kw1, kw2}
			}
		} else if kw1 == "baseline" {
			if kw2 == "first" || kw2 == "last" {
				return pr.Strings{kw1, kw2}
			}
		} else if kw2 == "baseline" {
			if kw1 == "first" || kw1 == "last" {
				return pr.Strings{kw1, kw2}
			}
		}
	}
	return nil
}

// @validator()
// “align-content“ property validation.
func alignContent(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) == 1 {
		switch keyword := getKeyword(tokens[0]); keyword {
		case "center", "space-between", "space-around", "space-evenly",
			"stretch", "normal", "flex-start", "flex-end",
			"start", "end":
			return pr.Strings{keyword}
		case "baseline":
			return pr.Strings{"first", keyword}
		}
	} else if len(tokens) == 2 {
		kw1, kw2 := getKeyword(tokens[0]), getKeyword(tokens[1])
		if kw1 == "safe" || kw1 == "unsafe" {
			switch kw2 {
			case "center", "start", "end", "flex-start", "flex-end":
				return pr.Strings{kw1, kw2}
			}
		} else if kw1 == "baseline" {
			if kw2 == "first" || kw2 == "last" {
				return pr.Strings{kw1, kw2}
			}
		} else if kw2 == "baseline" {
			if kw1 == "first" || kw1 == "last" {
				return pr.Strings{kw1, kw2}
			}
		}
	}
	return nil
}

// @validator(unstable=true)
// @singleKeyword
// Validation for “image-rendering“.
func imageRendering(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "crisp-edges", "pixelated":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @property(unstable=True)
// Validation for “image-orientation“.
func imageOrientation(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	if keyword == "none" || keyword == "from-image" {
		return pr.SBoolFloat{String: keyword}
	}
	var (
		angle             pr.Fl
		flip              bool
		hasAngle, hasFlip bool
	)
	for _, token := range tokens {
		keyword = getKeyword(token)
		if keyword == "flip" {
			if hasFlip {
				return nil
			}
			flip, hasFlip = true, true
			continue
		}
		if !hasAngle {
			var ok bool
			angle, ok = getAngle(token)
			hasAngle = true
			if ok {
				continue
			}
		}
		return nil
	}

	return pr.SBoolFloat{Float: angle, Bool: flip}
}

// @validator(unstable=true)
// “size“ property validation.
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

// @validator(proprietary=true)
// @singleToken
// Validation for “anchor“.
func anchor(tokens []Token, _ string) (out pr.CssProperty) {
	if len(tokens) != 1 {
		return
	}
	token := tokens[0]
	if getKeyword(token) == "none" {
		return pr.String("none")
	}
	name, args := pa.ParseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(pa.Ident); ok && name == "attr" {
				return pr.AttrData{Name: string(ident.Value)}
			}
		}
	}
	return
}

// @validator(proprietary=true, wantsBaseUrl=true)
// @singleToken
// Validation for “link“.
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
	name, args := pa.ParseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(pa.Ident); ok && name == "attr" {
				attr = pr.AttrData{Name: string(ident.Value)}
			}
		}
	}
	if !attr.IsNone() {
		out = attr
	}
	return
}

// @validator()
// @singleToken
// Validation for “tab-size“.
// See https://www.w3.org/TR/css-text-3/#tab-size
func tabSize(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(pa.Number); ok {
		if number.IsInt() && number.ValueF >= 0 { // no unit means multiple of space width
			return pr.NewDim(pr.Float(number.ValueF), 0).ToValue()
		}
	}
	return getLength(token, false, false).ToValue()
}

// @validator(unstable=true)
// @singleToken
// Validation for “hyphens“.
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

// @validator(unstable=true)
// @singleToken
// Validation for “hyphenate-character“.
func hyphenateCharacter(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "auto" {
		return pr.String("‐")
	} else if str, ok := token.(pa.String); ok {
		return pr.String(str.Value)
	}
	return nil
}

// @validator(unstable=true)
// @singleToken
// Validation for “hyphenate-limit-zone“.
func hyphenateLimitZone(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	d := getLength(token, false, true)
	if d.IsNone() {
		return nil
	}
	return d.ToValue()
}

// @validator(unstable=true)
// Validation for “hyphenate-limit-chars“.
func hyphenateLimitChars(tokens []Token, _ string) pr.CssProperty {
	switch len(tokens) {
	case 1:
		token := tokens[0]
		keyword := getKeyword(token)
		if keyword == "auto" {
			return pr.Ints3{5, 2, 2}
		} else if number, ok := token.(pa.Number); ok && number.IsInt() {
			return pr.Ints3{number.Int(), 2, 2}
		}
	case 2:
		total, left := tokens[0], tokens[1]
		totalKeyword := getKeyword(total)
		leftKeyword := getKeyword(left)
		if totalNumber, ok := total.(pa.Number); ok && totalNumber.IsInt() {
			if leftNumber, ok := left.(pa.Number); ok && leftNumber.IsInt() {
				return pr.Ints3{totalNumber.Int(), leftNumber.Int(), leftNumber.Int()}
			} else if leftKeyword == "auto" {
				return pr.Ints3{totalNumber.Int(), 2, 2}
			}
		} else if totalKeyword == "auto" {
			if leftNumber, ok := left.(pa.Number); ok && leftNumber.IsInt() {
				return pr.Ints3{5, leftNumber.Int(), leftNumber.Int()}
			} else if leftKeyword == "auto" {
				return pr.Ints3{5, 2, 2}
			}
		}
	case 3:
		total, left, right := tokens[0], tokens[1], tokens[2]
		totalNumber, okT := total.(pa.Number)
		leftNumber, okL := left.(pa.Number)
		rightNumber, okR := right.(pa.Number)
		if ((okT && totalNumber.IsInt()) || getKeyword(total) == "auto") &&
			((okL && leftNumber.IsInt()) || getKeyword(left) == "auto") &&
			((okR && rightNumber.IsInt()) || getKeyword(right) == "auto") {
			totalInt := 5
			if okT {
				totalInt = totalNumber.Int()
			}
			leftInt := 2
			if okL {
				leftInt = leftNumber.Int()
			}
			rightInt := 2
			if okR {
				rightInt = rightNumber.Int()
			}
			return pr.Ints3{totalInt, leftInt, rightInt}
		}
	}
	return nil
}

// @validator(proprietary=true)
// @singleToken
// Validation for “lang“.
func lang(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "none" {
		return pr.TaggedString{Tag: pr.None}
	}
	name, args := pa.ParseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(pa.Ident); ok && name == "attr" {
				return pr.TaggedString{Tag: pr.Attr, S: ident.Value}
			}
		}
	} else if str, ok := token.(pa.String); ok {
		return pr.TaggedString{S: str.Value}
	}
	return nil
}

// @validator(unstable=true)
// Validation for “bookmark-label“.
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

// @validator(unstable=true)
// @singleToken
// Validation for “bookmark-level“.
func bookmarkLevel(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(pa.Number); ok && number.IsInt() && number.Int() >= 1 {
		return pr.TaggedInt{I: number.Int()}
	} else if getKeyword(token) == "none" {
		return pr.TaggedInt{Tag: pr.None}
	}
	return nil
}

// @validator(unstable=True)
// @single_keyword
// Validation for “bookmark-state“.
func bookmarkState(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "open", "closed":
		return pr.String(keyword)
	default:
		return nil
	}
}

// @validator(unstable=true)
// @commaSeparatedList
// Validation for “string-set“.
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
	for _, part := range pa.SplitOnComma(tokens) {
		result, err := _stringSet(pa.RemoveWhitespace(part), baseUrl)
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

// @validator()
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
// “box-ellipsis“ property validation.
func blockEllipsis(tokens []Token, _ string) pr.CssProperty {
	if v, ok := blockEllipsis_(tokens); ok {
		return v
	}
	return nil
}

func blockEllipsis_(tokens []Token) (out pr.TaggedString, ok bool) {
	if len(tokens) != 1 {
		return
	}
	token := tokens[0]
	if str, ok := token.(pa.String); ok {
		return pr.TaggedString{S: str.Value}, true
	}
	switch keyword := getKeyword(token); keyword {
	case "none":
		return pr.TaggedString{Tag: pr.None}, true
	case "auto":
		return pr.TaggedString{Tag: pr.Auto}, true
	}
	return
}

func transformFunction(token Token) (pr.SDimensions, error) {
	name, args := pa.ParseFunction(token)
	if name == "" {
		return pr.SDimensions{}, ErrInvalidValue
	}

	lengths, values := make([]pr.Dimension, len(args)), make([]pr.Dimension, len(args))
	isAllNumber, isAllLengths := true, true
	for index, a := range args {
		lengths[index] = getLength(a, true, true)
		isAllLengths = isAllLengths && !lengths[index].IsNone()
		if aNumber, ok := a.(pa.Number); ok {
			values[index] = pr.FToD(pr.Fl(aNumber.ValueF))
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
		case "skewx", "skew":
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
			if number, ok := args[0].(pa.Number); ok {
				return pr.SDimensions{String: "scale", Dimensions: []pr.Dimension{pr.FToD(number.ValueF), pr.FToD(1.)}}, nil
			}
		case "scaley":
			if number, ok := args[0].(pa.Number); ok {
				return pr.SDimensions{String: "scale", Dimensions: []pr.Dimension{pr.FToD(1.), pr.FToD(number.ValueF)}}, nil
			}
		case "scale":
			if number, ok := args[0].(pa.Number); ok {
				return pr.SDimensions{String: "scale", Dimensions: []pr.Dimension{pr.FToD(number.ValueF), pr.FToD(number.ValueF)}}, nil
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
	return pr.SDimensions{}, ErrInvalidValue
}

func maxLines(tokens []Token, _ string) pr.CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if token, ok := token.(pa.Number); ok {
		if token.IsInt() {
			return pr.TaggedInt{I: token.Int()}
		}
	}
	if keyword := getKeyword(token); keyword == "none" {
		return pr.TaggedInt{Tag: pr.None}
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

// @property()
// @single_token
// “appearance“ property validation.
func appearance(tokens []Token, _ string) pr.CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "none":
		return pr.String(keyword)
	default:
		return nil
	}
}
