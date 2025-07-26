package validation

import (
	"errors"
	"fmt"
	"strings"

	pa "github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils"
)

type expander func(baseURL string, name pr.Shortand, tokens []Token) (expandedProperties, error)

var expanders = [...]expander{
	pr.SBorderColor: expandFourSides,
	pr.SBorderStyle: expandFourSides,
	pr.SBorderWidth: expandFourSides,
	pr.SBorderImage: genericExpander(pr.PBorderImageOutset, pr.PBorderImageRepeat, pr.PBorderImageSlice, pr.PBorderImageSource, pr.PBorderImageWidth)(_expandBorderImage),
	pr.SMargin:      expandFourSides,
	pr.SPadding:     expandFourSides,
	pr.SBleed:       expandFourSides,
	pr.SBorderRadius: genericExpander(
		pr.PBorderTopLeftRadius, pr.PBorderTopRightRadius,
		pr.PBorderBottomRightRadius, pr.PBorderBottomLeftRadius)(_borderRadius),
	pr.SPageBreakAfter:  genericExpander(pr.PBreakAfter)(_expandPageBreakBeforeAfter),
	pr.SPageBreakBefore: genericExpander(pr.PBreakBefore)(_expandPageBreakBeforeAfter),
	pr.SPageBreakInside: genericExpander(pr.PBreakInside)(_expandPageBreakInside),
	pr.SBackground:      expandBackground,
	pr.SWordWrap:        genericExpander(pr.POverflowWrap)(_expandWordWrap),
	pr.SListStyle:       genericExpander(pr.PListStyleType, pr.PListStylePosition, pr.PListStyleImage)(_expandListStyle),
	pr.SBorder:          expandBorder,
	pr.SBorderTop:       borderExpanders[0],
	pr.SBorderRight:     borderExpanders[1],
	pr.SBorderBottom:    borderExpanders[2],
	pr.SBorderLeft:      borderExpanders[3],
	pr.SColumnRule:      genericExpander(pr.PColumnRuleWidth, pr.PColumnRuleColor, pr.PColumnRuleStyle)(_expandBorderSide),
	pr.SOutline:         genericExpander(pr.POutlineWidth, pr.POutlineColor, pr.POutlineStyle)(_expandBorderSide),
	pr.SColumns:         genericExpander(pr.PColumnWidth, pr.PColumnCount)(_expandColumns),
	pr.SFontVariant: genericExpander(
		pr.PFontVariantAlternates, pr.PFontVariantCaps, pr.PFontVariantEastAsian, pr.PFontVariantLigatures,
		pr.PFontVariantNumeric, pr.PFontVariantPosition)(_fontVariant),
	pr.SFont: genericExpander(pr.PFontStyle, pr.PFontVariantCaps, pr.PFontWeight, pr.PFontStretch, pr.PFontSize,
		pr.PLineHeight, pr.PFontFamily)(_expandFont),
	pr.STextDecoration: genericExpander(pr.PTextDecorationLine, pr.PTextDecorationColor, pr.PTextDecorationStyle, pr.PTextDecorationThickness)(_expandTextDecoration),
	pr.SFlex:           genericExpander(pr.PFlexGrow, pr.PFlexShrink, pr.PFlexBasis)(_expandFlex),
	pr.SFlexFlow:       genericExpander(pr.PFlexDirection, pr.PFlexWrap)(_expandFlexFlow),
	pr.SLineClamp:      genericExpander(pr.PMaxLines, pr.PContinue, pr.PBlockEllipsis)(_expandLineClamp),
	pr.STextAlign:      genericExpander(pr.PTextAlignAll, pr.PTextAlignLast)(_expandTextAlign),
	pr.SGridColumn:     genericExpander(pr.PGridColumnStart, pr.PGridColumnEnd)(_expandGridColumnRow),
	pr.SGridRow:        genericExpander(pr.PGridRowStart, pr.PGridRowEnd)(_expandGridColumnRow),
	pr.SGridArea:       genericExpander(pr.PGridRowStart, pr.PGridRowEnd, pr.PGridColumnStart, pr.PGridColumnEnd)(_expandGridArea),
	pr.SGridTemplate:   genericExpander(pr.PGridTemplateColumns, pr.PGridTemplateRows, pr.PGridTemplateAreas)(_expandGridTemplate),
	pr.SGrid:           genericExpander(pr.PGridTemplateColumns, pr.PGridTemplateRows, pr.PGridTemplateAreas, pr.PGridAutoColumns, pr.PGridAutoRows, pr.PGridAutoFlow)(_expandGrid),
}

var borderExpanders = [...]expander{
	genericExpander(pr.PBorderTopWidth, pr.PBorderTopColor, pr.PBorderTopStyle)(_expandBorderSide),
	genericExpander(pr.PBorderRightWidth, pr.PBorderRightColor, pr.PBorderRightStyle)(_expandBorderSide),
	genericExpander(pr.PBorderBottomWidth, pr.PBorderBottomColor, pr.PBorderBottomStyle)(_expandBorderSide),
	genericExpander(pr.PBorderLeftWidth, pr.PBorderLeftColor, pr.PBorderLeftStyle)(_expandBorderSide),
}

// var expandBorderSide = genericExpander("-width", "-color", "-style")(_expandBorderSide)

func ExpandValidatePending(prop pr.KnownProp, from pr.Shortand, tokens []Token) (pr.DeclaredValue, error) {
	props, err := expanders[from]("", from, tokens)
	if err != nil {
		return nil, err
	}
	for _, expanded := range props {
		if expanded.name.KnownProp == prop {
			return expanded.property, nil
		}
	}
	return nil, fmt.Errorf("missing key %s in expanded property", prop)
}

// Return pending expanders when var is found in tokens.
func findVar(shortand pr.Shortand, tokens []Token, expandedNames []pr.KnownProp) ([]namedProperty, bool) {
	for _, token := range tokens {
		if HasVar(token) {
			// Found CSS variable, keep pending-substitution values.
			out := make([]namedProperty, len(expandedNames))
			for i, name := range expandedNames {
				out[i] = namedProperty{pr.PropKey{KnownProp: name}, pr.RawTokens(tokens), shortand}
			}
			return out, true
		}
	}
	return nil, false
}

// Expanders

type namedProperty struct {
	name     pr.PropKey
	property pr.DeclaredValue
	// if not empty, [property] is [pr.RawTokens] and refers to the associate shorthand, not to the expanded [name]
	shortand pr.Shortand
}

type expandedProperties []namedProperty

type namedTokens struct {
	name   pr.KnownProp
	tokens []Token
}

type beforeGeneric = func(baseURL string, name pr.Shortand, tokens []Token) ([]namedTokens, error)

// Decorator helping expanders to handle 'inherit' and 'initial'.
// Wrap an expander so that it does not have to handle the 'inherit' and
// 'initial' cases, and can just yield name suffixes. Missing suffixes
// get the initial value.
func genericExpander(expandedNames ...pr.KnownProp) func(beforeGeneric) expander {
	_expandedNames := pr.NewSetK(expandedNames...)

	// Decorate the 'wrapped' expander.
	genericExpanderDecorator := func(wrapped beforeGeneric) expander {
		// Wrap the expander.
		genericExpanderWrapper := func(baseURL string, shortand pr.Shortand, tokens []Token) (out expandedProperties, err error) {
			results := map[string]pr.DeclaredValue{} // FIXME: check if we can use Knownprop
			keyword := getSingleKeyword(tokens)
			var (
				skipValidation bool
				isPending      bool
			)
			if keyword == "inherit" || keyword == "initial" {
				val := pr.NewDefaultValue(keyword)
				for _, name := range expandedNames {
					results[name.String()] = val
				}
				skipValidation = true
			} else {
				if props, ok := findVar(shortand, tokens, expandedNames); ok {
					for _, prop := range props {
						results[prop.name.String()] = pr.RawTokens(tokens)
					}
					isPending = true
					skipValidation = true
				}
			}

			if !skipValidation {
				result, err := wrapped(baseURL, shortand, tokens)
				if err != nil {
					return nil, err
				}

				for _, nameToken := range result {
					newName, newToken := nameToken.name, nameToken.tokens
					if !_expandedNames.Has(newName) {
						return nil, fmt.Errorf("unknown expanded property %s", newName)
					}
					if _, isIn := results[newName.String()]; isIn {
						return nil, fmt.Errorf("got multiple %s values in a %s shorthand",
							newName, shortand)
					}
					results[newName.String()] = pr.RawTokens(newToken)
				}
			}

			for _, newName := range expandedNames {
				actualNewName := newName

				value, ok := results[newName.String()]
				if ok {
					if !skipValidation {
						np, err := validateNonShorthand(baseURL, actualNewName.String(), value.(pr.RawTokens), true)
						if err != nil {
							return nil, fmt.Errorf("validating %s: %s", actualNewName, err)
						}
						actualNewName, value = np.name.KnownProp, np.property
					}
				} else {
					value = pr.Initial
				}

				// actualNewName is now a valid prop name
				np := namedProperty{
					name:     pr.PropKey{KnownProp: actualNewName},
					property: value,
				}
				if isPending {
					np.shortand = shortand
				}
				out = append(out, np)
			}
			return out, nil
		}
		return genericExpanderWrapper
	}
	return genericExpanderDecorator
}

// Expand properties setting a token for the four sides of a box.
// "border-color", "border-style", "border-width", "margin", "padding", "bleed"
func expandFourSides(baseURL string, name pr.Shortand, tokens []Token) (out expandedProperties, err error) {
	// Define expanded names
	nameString := name.String()
	indexM := strings.LastIndex(nameString, "-")
	var expandedNames [4]pr.KnownProp
	for i, suffix := range [4]string{"-top", "-right", "-bottom", "-left"} {
		var newName string
		if indexM == -1 {
			newName = nameString + suffix
		} else {
			// eg. border-color becomes border-*-color, not border-color-*
			newName = nameString[:indexM] + suffix + nameString[indexM:]
		}
		expandedNames[i] = pr.PropsFromNames[newName]
	}

	if result, ok := findVar(name, tokens, expandedNames[:]); ok {
		return result, nil
	}

	// Make sure we have 4 tokens
	if len(tokens) == 1 {
		tokens = []Token{tokens[0], tokens[0], tokens[0], tokens[0]}
	} else if len(tokens) == 2 {
		tokens = []Token{tokens[0], tokens[1], tokens[0], tokens[1]} // (bottom, left) defaults to (top, right)
	} else if len(tokens) == 3 {
		tokens = append(tokens, tokens[1]) // left defaults to right
	} else if len(tokens) != 4 {
		return out, fmt.Errorf("expected 1 to 4 token components got %d", len(tokens))
	}

	for index, expandedName := range expandedNames {
		token := tokens[index]
		prop, err := validateNonShorthand(baseURL, expandedName.String(), []Token{token}, true)
		if err != nil {
			return nil, err
		}
		out = append(out, prop)
	}
	return out, nil
}

// Validator for the `border-radius` property.
func _borderRadius(baseURL string, _ pr.Shortand, tokens []Token) (out []namedTokens, err error) {
	var horizontal, vertical []Token
	current := &horizontal

	for index, token := range tokens {
		if lit, ok := token.(pa.Literal); ok && lit.Value == "/" {
			if current == &horizontal {
				if index == len(tokens)-1 {
					return nil, errors.New("expected value after '/' separator")
				} else {
					current = &vertical
				}
			} else {
				return nil, errors.New("expected only one '/' separator")
			}
		} else {
			*current = append(*current, token)
		}
	}

	if len(vertical) == 0 {
		vertical = append([]Token{}, horizontal...)
	}

	// Make sure we have 4 tokens
	for _, values := range [2]*[]Token{&horizontal, &vertical} {
		if len(*values) == 1 {
			*values = []Token{(*values)[0], (*values)[0], (*values)[0], (*values)[0]}
		} else if len(*values) == 2 {
			*values = []Token{(*values)[0], (*values)[1], (*values)[0], (*values)[1]} // (br, bl) defaults to (tl, tr)
		} else if len(*values) == 3 {
			*values = append(*values, (*values)[1]) // bl defaults to tr
		} else if len(*values) != 4 {
			return nil, fmt.Errorf("expected 1 to 4 token components got %d", len(*values))
		}
	}
	corners := [4]string{"top-left", "top-right", "bottom-right", "bottom-left"}
	for index, corner := range corners {
		name := fmt.Sprintf("border-%s-radius", corner)
		ts := []Token{horizontal[index], vertical[index]}
		_, err = validateNonShorthand(baseURL, name, ts, true)
		if err != nil {
			return nil, err
		}
		out = append(out, namedTokens{name: pr.PropsFromNames[name], tokens: ts})
	}
	return out, nil
}

// @expander("list-style")
// @genericExpander("-type", "-position", "-image", wantsBaseUrl=true)
// Expand the “list-style“ shorthand property.
//
//	See http://www.w3.org/TR/CSS21/generate.html#propdef-list-style
func _expandListStyle(baseURL string, _ pr.Shortand, tokens []Token) (out []namedTokens, err error) {
	var typeSpecified, imageSpecified bool
	noneCount := 0
	var noneToken Token
	for _, token := range tokens {
		var suffix string
		if getKeyword(token) == "none" {
			// Can be either -style or -image, see at the end which is not
			// otherwise specified.
			noneCount += 1
			noneToken = token
			continue
		}

		if image, _ := listStyleImage([]Token{token}, baseURL); image != nil {
			suffix = "-image"
			imageSpecified = true
		} else if listStylePosition([]Token{token}, "") != nil {
			suffix = "-position"
		} else if _, ok := listStyleType_([]Token{token}); ok {
			suffix = "-type"
			typeSpecified = true
		} else {
			return nil, ErrInvalidValue
		}
		out = append(out, namedTokens{name: pr.PropsFromNames["list-style"+suffix], tokens: []Token{token}})
	}

	if !typeSpecified && noneCount > 0 {
		out = append(out, namedTokens{name: pr.PListStyleType, tokens: []Token{noneToken}})
		noneCount -= 1
	}

	if !imageSpecified && noneCount > 0 {
		out = append(out, namedTokens{name: pr.PListStyleImage, tokens: []Token{noneToken}})
		noneCount -= 1
	}

	if noneCount > 0 {
		// Too many none tokens.
		return nil, ErrInvalidValue
	}
	return out, nil
}

// @expander("border")
// Expand the “border“ shorthand property.
//
//	See http://www.w3.org/TR/CSS21/box.html#propdef-border
func expandBorder(baseURL string, _ pr.Shortand, tokens []Token) (out expandedProperties, err error) {
	for suffix := pr.Shortand(0); suffix <= 3; suffix++ {
		prop := pr.SBorderTop + suffix
		expander := borderExpanders[suffix]
		props, err := expander(baseURL, prop, tokens)
		if err != nil {
			return nil, err
		}
		out = append(out, props...)
	}
	return out, nil
}

// Expand the “border-*“ shorthand pr.
// "border-top", "border-right", "border-bottom", "border-left", "column-rule", "outline"
//
//	See http://www.w3.org/TR/CSS21/box.html#propdef-border-top
func _expandBorderSide(_ string, shortand pr.Shortand, tokens []Token) ([]namedTokens, error) {
	out := make([]namedTokens, len(tokens))
	for index, token := range tokens {
		var suffix string
		if !pa.ParseColor(token).IsNone() {
			suffix = "-color"
		} else if borderWidth([]Token{token}, "") != nil {
			suffix = "-width"
		} else if borderStyle([]Token{token}, "") != nil {
			suffix = "-style"
		} else {
			return nil, ErrInvalidValue
		}
		out[index] = namedTokens{name: pr.PropsFromNames[shortand.String()+suffix], tokens: []Token{token}}
	}
	return out, nil
}

// @expander('border-image')
// @generic_expander('-outset', '-repeat', '-slice', '-source', '-width')
// Expand the ``border-image-*`` shorthand properties.

// See https://drafts.csswg.org/css-backgrounds/#the-border-image
func _expandBorderImage(baseURL string, _ pr.Shortand, tokens []Token) (out []namedTokens, err error) {
	for len(tokens) != 0 {
		source, err := borderImageSource(tokens[:1], baseURL)
		if err != nil {
			return nil, err
		}
		if source != nil {
			var res []Token
			res, tokens = tokens[:1], tokens[1:]
			out = append(out, namedTokens{pr.PBorderImageSource, res})
		} else if borderImageRepeat(tokens[:1], "") != nil {
			var repeats []Token
			repeats, tokens = tokens[:1], tokens[1:]
			for len(tokens) != 0 && borderImageRepeat(tokens[:1], "") != nil {
				var repeat Token
				repeat, tokens = tokens[0], tokens[1:]
				repeats = append(repeats, repeat)
			}
			out = append(out, namedTokens{pr.PBorderImageRepeat, repeats})
		} else if borderImageSlice(tokens[:1], "") != nil || getKeyword(tokens[0]) == "fill" {
			var slices []Token
			slices, tokens = tokens[:1], tokens[1:]
			for len(tokens) != 0 && borderImageSlice(append(slices, tokens[0]), "") != nil {
				var res Token
				res, tokens = tokens[0], tokens[1:]
				slices = append(slices, res)
			}
			out = append(out, namedTokens{pr.PBorderImageSlice, slices})
			if len(tokens) != 0 && tokens[0].Kind() == pa.KLitteral && tokens[0].(pa.Literal).Value == "/" {
				// slices / *
				tokens = tokens[1:]
			} else {
				// slices other
				continue
			}
			if len(tokens) == 0 {
				// slices /
				return nil, ErrInvalidValue
			}

			if borderImageWidth(tokens[:1], "") != nil {
				var widths []Token
				widths, tokens = tokens[:1], tokens[1:]
				for len(tokens) != 0 && borderImageWidth(append(widths, tokens[0]), "") != nil {
					var res Token
					res, tokens = tokens[0], tokens[1:]
					widths = append(widths, res)
				}
				out = append(out, namedTokens{pr.PBorderImageWidth, widths})
				if len(tokens) != 0 && tokens[0].Kind() == pa.KLitteral && tokens[0].(pa.Literal).Value == "/" {
					// slices / widths / slash *
					tokens = tokens[1:]
				} else {
					// slices / widths .
					continue
				}
			} else if len(tokens) != 0 && tokens[0].Kind() == pa.KLitteral && tokens[0].(pa.Literal).Value == "/" {
				// slices / / *
				tokens = tokens[1:]
			} else {
				// slices / other
				return nil, ErrInvalidValue
			}
			if len(tokens) == 0 {
				// slices / * /
				return nil, ErrInvalidValue
			}
			if borderImageOutset(tokens[:1], "") != nil {
				var outsets []Token
				outsets, tokens = tokens[:1], tokens[1:]
				for len(tokens) != 0 && borderImageOutset(append(outsets, tokens[0]), "") != nil {
					var res Token
					res, tokens = tokens[0], tokens[1:]
					outsets = append(outsets, res)
				}
				out = append(out, namedTokens{pr.PBorderImageOutset, outsets})
			} else {
				// slash / * / other
				return nil, ErrInvalidValue
			}
		} else {
			return nil, ErrInvalidValue
		}
	}

	return out, nil
}

type backgroundProps struct {
	color      pr.CssProperty
	image      pr.Image
	_keys      utils.Set
	repeat     [2]string
	attachment string
	clip       string
	origin     string
	size       pr.Size
	position   pr.Center
}

func (b backgroundProps) add(name string) error {
	name = "background_" + name
	if b._keys.Has(name) {
		return fmt.Errorf("invalid value : name %s already set", name)
	}
	b._keys.Add(name)
	return nil
}

func reverseLayers(a [][]Token) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}

var expandedBackgroundNames = [...]pr.KnownProp{
	pr.PBackgroundColor,
	pr.PBackgroundImage,
	pr.PBackgroundRepeat,
	pr.PBackgroundAttachment,
	pr.PBackgroundPosition,
	pr.PBackgroundSize,
	pr.PBackgroundClip,
	pr.PBackgroundOrigin,
}

// Expand the “background“ shorthand property.
// See http://drafts.csswg.org/csswg/css-backgrounds-3/#the-background
func expandBackground(baseURL string, shortand pr.Shortand, tokens []Token) (out expandedProperties, err error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "initial" || keyword == "inherit" {
		val := pr.NewDefaultValue(keyword)
		for prop := pr.PBackgroundColor; prop <= pr.PBackgroundOrigin; prop++ {
			out = append(out, namedProperty{name: pr.PropKey{KnownProp: prop}, property: val})
		}
		return
	}

	if out, ok := findVar(shortand, tokens, expandedBackgroundNames[:]); ok {
		return out, nil
	}

	parseLayer := func(tokens []Token, finalLayer bool) (pr.CssProperty, backgroundProps, error) {
		results := backgroundProps{_keys: utils.Set{}}

		// Make `tokens` a stack
		tokens = reverse(tokens)
		for len(tokens) > 0 {
			i := utils.MaxInt(len(tokens)-2, 0)
			repeat := _backgroundRepeat(reverse(tokens[i:]))
			if repeat != [2]string{} {
				if err = results.add("repeat"); err != nil {
					return pr.Color{}, backgroundProps{}, err
				}
				results.repeat = repeat
				tokens = tokens[:i]
				continue
			}

			lastToken := tokens[len(tokens)-1:]
			if finalLayer {
				color := otherColors(lastToken, "")
				if color != nil {
					if err = results.add("color"); err != nil {
						return pr.Color{}, backgroundProps{}, err
					}
					results.color = color
					tokens = tokens[:len(tokens)-1]
					continue
				}
			}

			image, err := _backgroundImage(lastToken, baseURL)
			if err != nil {
				return pr.Color{}, backgroundProps{}, err
			}
			if image != nil {
				if err = results.add("image"); err != nil {
					return pr.Color{}, backgroundProps{}, err
				}
				results.image = image
				tokens = tokens[:len(tokens)-1]
				continue
			}

			repeat = _backgroundRepeat(lastToken)
			if repeat != [2]string{} {
				if err = results.add("repeat"); err != nil {
					return pr.Color{}, backgroundProps{}, err
				}
				results.repeat = repeat
				tokens = tokens[:len(tokens)-1]
				continue
			}

			attachment := _backgroundAttachment(lastToken)
			if attachment != "" {
				if err = results.add("attachment"); err != nil {
					return pr.Color{}, backgroundProps{}, err
				}
				results.attachment = attachment
				tokens = tokens[:len(tokens)-1]
				continue
			}

			var position pr.Center
			start := len(tokens)
			if start > 4 {
				start = 4
			}
			for n := start; n >= 1; n-- {
				positionT := reverse(tokens[len(tokens)-n:])
				position = parsePosition(positionT)
				if !position.IsNone() {
					if err = results.add("position"); err != nil {
						return pr.Color{}, backgroundProps{}, err
					}
					results.position = position
					tokens = tokens[:len(tokens)-n]
					if len(tokens) > 0 {
						if lit, ok := tokens[len(tokens)-1].(pa.Literal); ok && lit.Value == "/" {
							start := len(tokens) + 1
							if start > 3 {
								start = 3
							}
							for n := start; n >= 2; n-- {
								// n includes the "/" delimiter.
								i, j := utils.MaxInt(0, len(tokens)-n), utils.MaxInt(0, len(tokens)-1)
								positionT = reverse(tokens[i:j])
								size := _backgroundSize(positionT)
								if !size.IsNone() {
									if err = results.add("size"); err != nil {
										return pr.Color{}, backgroundProps{}, err
									}
									results.size = size
									tokens = tokens[:i]
								}
							}
						}
					}
					break
				}
			}
			if !position.IsNone() {
				continue
			}

			origin := _box(lastToken)
			if origin != "" {
				if err = results.add("origin"); err != nil {
					return pr.Color{}, backgroundProps{}, err
				}
				results.origin = origin
				tokens = tokens[:len(tokens)-1]

				nextToken := tokens[utils.MaxInt(0, len(tokens)-1):]

				clip := _box(nextToken)
				if clip != "" {
					if err = results.add("clip"); err != nil {
						return pr.Color{}, backgroundProps{}, err
					}
					results.clip = clip
					tokens = tokens[:len(tokens)-1]
				} else {
					// The same keyword sets both:
					clip := _box(lastToken)
					if clip == "" {
						return pr.Color{}, backgroundProps{}, errors.New("clip shoudn't be empty")
					}
					if err = results.add("clip"); err != nil {
						return pr.Color{}, backgroundProps{}, err
					}
					results.clip = clip
				}
				continue
			}
			return pr.Color{}, backgroundProps{}, ErrInvalidValue
		}

		var color pr.CssProperty = pr.InitialValues.GetBackgroundColor()
		if results._keys.Has("background_color") {
			color = results.color
			delete(results._keys, "background_color")
		}

		if !results._keys.Has("background_image") {
			results.image = pr.InitialValues.GetBackgroundImage()[0]
		}
		if !results._keys.Has("background_repeat") {
			results.repeat = pr.InitialValues.GetBackgroundRepeat()[0]
		}
		if !results._keys.Has("background_attachment") {
			results.attachment = pr.InitialValues.GetBackgroundAttachment()[0]
		}
		if !results._keys.Has("background_position") {
			results.position = pr.InitialValues.GetBackgroundPosition()[0]
		}
		if !results._keys.Has("background_size") {
			results.size = pr.InitialValues.GetBackgroundSize()[0]
		}
		if !results._keys.Has("background_clip") {
			results.clip = pr.InitialValues.GetBackgroundClip()[0]
		}
		if !results._keys.Has("background_origin") {
			results.origin = pr.InitialValues.GetBackgroundOrigin()[0]
		}
		return color, results, nil
	}

	layers := pa.SplitOnComma(tokens)
	reverseLayers(layers)

	var resultColor pr.CssProperty

	n := len(layers)
	resultsImages := make(pr.Images, n)
	resultsRepeats := make(pr.Repeats, n)
	resultsAttachments := make(pr.Strings, n)
	resultsPositions := make(pr.Centers, n)
	resultsSizes := make(pr.Sizes, n)
	resultsClips := make(pr.Strings, n)
	resultsOrigins := make(pr.Strings, n)

	for i, tokens := range layers {
		layerColor, layer, err := parseLayer(tokens, i == 0)
		if i == 0 {
			resultColor = layerColor
		}
		if err != nil {
			return nil, err
		}
		resultsImages[i] = layer.image
		resultsRepeats[i] = layer.repeat
		resultsAttachments[i] = layer.attachment
		resultsPositions[i] = layer.position
		resultsSizes[i] = layer.size
		resultsClips[i] = layer.clip
		resultsOrigins[i] = layer.origin
	}

	// un-reverse
	for left, right := 0, n-1; left < right; left, right = left+1, right-1 {
		resultsImages[left], resultsImages[right] = resultsImages[right], resultsImages[left]
		resultsRepeats[left], resultsRepeats[right] = resultsRepeats[right], resultsRepeats[left]
		resultsAttachments[left], resultsAttachments[right] = resultsAttachments[right], resultsAttachments[left]
		resultsPositions[left], resultsPositions[right] = resultsPositions[right], resultsPositions[left]
		resultsSizes[left], resultsSizes[right] = resultsSizes[right], resultsSizes[left]
		resultsClips[left], resultsClips[right] = resultsClips[right], resultsClips[left]
		resultsOrigins[left], resultsOrigins[right] = resultsOrigins[right], resultsOrigins[left]
	}

	out = expandedProperties{
		{name: pr.PropKey{KnownProp: pr.PBackgroundImage}, property: resultsImages},
		{name: pr.PropKey{KnownProp: pr.PBackgroundRepeat}, property: resultsRepeats},
		{name: pr.PropKey{KnownProp: pr.PBackgroundAttachment}, property: resultsAttachments},
		{name: pr.PropKey{KnownProp: pr.PBackgroundPosition}, property: resultsPositions},
		{name: pr.PropKey{KnownProp: pr.PBackgroundSize}, property: resultsSizes},
		{name: pr.PropKey{KnownProp: pr.PBackgroundClip}, property: resultsClips},
		{name: pr.PropKey{KnownProp: pr.PBackgroundOrigin}, property: resultsOrigins},
		{name: pr.PropKey{KnownProp: pr.PBackgroundColor}, property: resultColor},
	}
	return out, nil
}

// @expander("text-decoration")
func _expandTextDecoration(_ string, _ pr.Shortand, tokens []Token) (out []namedTokens, err error) {
	var (
		lines      []Token
		colors     []Token
		styles     []Token
		thickness  []Token
		noneInLine bool
	)

	for _, token := range tokens {
		keyword := getKeyword(token)
		switch keyword {
		case "none", "underline", "overline", "line-through", "blink":
			lines = append(lines, token)
			if noneInLine {
				return nil, ErrInvalidValue
			} else if keyword == "none" {
				noneInLine = true
			}
		case "solid", "double", "dotted", "dashed", "wavy":
			if len(styles) != 0 {
				return nil, ErrInvalidValue
			} else {
				styles = append(styles, token)
			}
		default:
			if color := pa.ParseColor(token); !color.IsNone() {
				if len(colors) != 0 {
					return nil, ErrInvalidValue
				}
				colors = append(colors, token)
			} else if th := textDecorationThickness([]Token{token}, ""); th != nil {
				if len(thickness) != 0 {
					return nil, ErrInvalidValue
				}
				thickness = append(thickness, token)
			} else {
				return nil, ErrInvalidValue
			}
		}
	}

	if len(lines) != 0 {
		out = append(out, namedTokens{name: pr.PTextDecorationLine, tokens: lines})
	}
	if len(colors) != 0 {
		out = append(out, namedTokens{name: pr.PTextDecorationColor, tokens: colors})
	}
	if len(styles) != 0 {
		out = append(out, namedTokens{name: pr.PTextDecorationStyle, tokens: styles})
	}
	if len(thickness) != 0 {
		out = append(out, namedTokens{name: pr.PTextDecorationThickness, tokens: thickness})
	}

	return out, nil
}

// Expand legacy “page-break-before“ && “page-break-after“ pr.
// See https://www.w3.org/TR/css-break-3/#page-break-properties
func _expandPageBreakBeforeAfter(_ string, name pr.Shortand, tokens []Token) (out []namedTokens, err error) {
	newName := pr.PBreakBefore
	if name == pr.SPageBreakAfter {
		newName = pr.PBreakAfter
	}
	keyword := getSingleKeyword(tokens)
	if keyword == "auto" || keyword == "left" || keyword == "right" || keyword == "avoid" {
		out = append(out, namedTokens{name: newName, tokens: tokens})
	} else if keyword == "always" {
		out = append(out, namedTokens{name: newName, tokens: []Token{
			pa.NewIdent("page", tokens[0].Pos()),
		}})
	} else {
		return nil, ErrInvalidValue
	}
	return out, nil
}

// Expand the legacy “page-break-inside“ property.
// See https://www.w3.org/TR/css-break-3/#page-break-properties
func _expandPageBreakInside(_ string, _ pr.Shortand, tokens []Token) ([]namedTokens, error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "auto" || keyword == "avoid" {
		return []namedTokens{{name: pr.PBreakInside, tokens: tokens}}, nil
	}

	return nil, ErrInvalidValue
}

// Expand the “columns“ shorthand property.
func _expandColumns(_ string, _ pr.Shortand, tokens []Token) (out []namedTokens, err error) {
	if len(tokens) == 2 && getKeyword(tokens[0]) == "auto" {
		tokens = reverse(tokens)
	}
	var name pr.KnownProp
	for _, token := range tokens {
		l := []Token{token}
		if columnWidth(l, "") != nil && name != pr.PColumnWidth {
			name = pr.PColumnWidth
		} else if columnCount(l, "") != nil {
			name = pr.PColumnCount
		} else {
			return nil, ErrInvalidValue
		}
		out = append(out, namedTokens{name: name, tokens: l})
	}

	if len(tokens) == 1 {
		if name == pr.PColumnWidth {
			name = pr.PColumnCount
		} else {
			name = pr.PColumnWidth
		}
		out = append(out, namedTokens{name: name, tokens: []Token{
			pa.NewIdent("auto", tokens[0].Pos()),
		}})
	}
	return out, nil
}

var (
	noneFakeToken   = pa.NewIdent("none", pa.Pos{})
	normalFakeToken = pa.NewIdent("normal", pa.Pos{})
)

// Expand the “font-variant“ shorthand property.
// https://www.w3.org/TR/css-fonts-3/#font-variant-prop
func _fontVariant(_ string, name pr.Shortand, tokens []Token) (out []namedTokens, err error) {
	return expandFontVariant(tokens)
}

func expandFontVariant(tokens []Token) (out []namedTokens, err error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "normal" || keyword == "none" {
		out = make([]namedTokens, 6)
		for index, suffix := range [5]pr.KnownProp{
			pr.PFontVariantAlternates,
			pr.PFontVariantCaps,
			pr.PFontVariantEastAsian,
			pr.PFontVariantNumeric,
			pr.PFontVariantPosition,
		} {
			out[index] = namedTokens{name: suffix, tokens: []Token{normalFakeToken}}
		}
		token := noneFakeToken
		if keyword == "normal" {
			token = normalFakeToken
		}
		out[5] = namedTokens{name: pr.PFontVariantLigatures, tokens: []Token{token}}
	} else {
		features := map[pr.KnownProp][]Token{}
		for _, token := range tokens {
			keyword := getKeyword(token)
			if keyword == "normal" {
				// We don"t allow "normal", only the specific values
				return nil, errors.New("invalid : normal not allowed")
			}
			found := false
			for i, validator := range fontVariantMapper {
				feature := pr.PFontVariantAlternates + pr.KnownProp(i)
				if validator([]Token{token}, "") != nil {
					features[feature] = append(features[feature], token)
					found = true
					break
				}
			}
			if !found {
				return nil, errors.New("invalid : font variant not supported")
			}
		}
		for feature, tokens := range features {
			if len(tokens) > 0 {
				out = append(out, namedTokens{name: feature, tokens: tokens})
			}
		}
	}
	return out, nil
}

var fontVariantMapper = [...]validator{
	fontVariantAlternates,
	fontVariantCaps,
	fontVariantEastAsian,
	fontVariantLigatures,
	fontVariantNumeric,
	fontVariantPosition,
}

// ExpandFont expands the 'font' property, to be used in
// SVG documents. It returns a list of (name, property) pairs
func ExpandFont(tokens []Token) ([][2]string, error) {
	l, err := _expandFont("", pr.SFont, tokens)
	if err != nil {
		return nil, err
	}

	out := make([][2]string, len(l))
	for i, v := range l {
		// name := v.name
		// if strings.HasPrefix(v.name, "-") { // newName is a suffix
		// 	name = "font" + v.name
		// }
		out[i] = [2]string{v.name.String(), pa.Serialize(v.tokens)}
	}

	return out, nil
}

// Expand the “font“ shorthand property.
// https://www.w3.org/TR/css-fonts-3/#font-prop
func _expandFont(_ string, _ pr.Shortand, tokens []Token) ([]namedTokens, error) {
	expandFontKeyword := getSingleKeyword(tokens)
	if expandFontKeyword == "caption" || expandFontKeyword == "icon" || expandFontKeyword == "menu" || expandFontKeyword == "message-box" || expandFontKeyword ==
		"small-caption" || expandFontKeyword == "status-bar" {

		return nil, errors.New("system fonts are not supported")
	}
	var (
		out   []namedTokens
		token Token
	)
	// Make `tokens` a stack
	tokens = reverse(tokens)
	// Values for font-style, font-variant-caps, font-weight and font-stretch
	// can come in any order and are all optional.
	hasBroken := false
	for i := 0; i < 4; i++ {
		token, tokens = tokens[len(tokens)-1], tokens[:len(tokens)-1]

		kw := getKeyword(token)
		if kw == "normal" {
			// Just ignore "normal" keywords. Unspecified properties will get
			// their initial token, which is "normal" for all three here.
			continue
		}

		var suffix pr.KnownProp
		if fontStyle([]Token{token}, "") != nil {
			suffix = pr.PFontStyle
		} else if fontVariantCaps([]Token{token}, "") != nil {
			suffix = pr.PFontVariantCaps
		} else if fontWeight([]Token{token}, "") != nil {
			suffix = pr.PFontWeight
		} else if fontStretch([]Token{token}, "") != nil {
			suffix = pr.PFontStretch
		} else {
			// We’re done with these four, continue with font-size
			hasBroken = true
			break
		}
		out = append(out, namedTokens{name: suffix, tokens: []Token{token}})

		if len(tokens) == 0 {
			return nil, ErrInvalidValue
		}
	}
	if !hasBroken {
		token, tokens = tokens[len(tokens)-1], tokens[:len(tokens)-1]
	}

	// Then font-size is mandatory
	// Latest `token` from the loop.
	fs, err := fontSize([]Token{token}, "")
	if err != nil {
		return nil, err
	}
	if fs == nil {
		return nil, errors.New("invalid : font-size is mandatory for short font attribute")
	}
	out = append(out, namedTokens{name: pr.PFontSize, tokens: []Token{token}})

	// Then line-height is optional, but font-family is not so the list
	// must not be empty yet
	if len(tokens) == 0 {
		return nil, errors.New("invalid : font-familly is mandatory for short font attribute")
	}

	token = tokens[len(tokens)-1]
	tokens = tokens[:len(tokens)-1]
	if lit, ok := token.(pa.Literal); ok && lit.Value == "/" {
		token = tokens[len(tokens)-1]
		tokens = tokens[:len(tokens)-1]
		if lineHeight([]Token{token}, "") == nil {
			return nil, ErrInvalidValue
		}
		out = append(out, namedTokens{name: pr.PLineHeight, tokens: []Token{token}})
	} else {
		// We pop()ed a font-family, add it back
		tokens = append(tokens, token)
	}
	// Reverse the stack to get normal list
	tokens = reverse(tokens)
	if fontFamily(tokens, "") == nil {
		return nil, ErrInvalidValue
	}
	out = append(out, namedTokens{name: pr.PFontFamily, tokens: tokens})
	return out, nil
}

// Expand the “word-wrap“ legacy property.
// See https://www.w3.org/TR/css-text-3/#overflow-wrap
func _expandWordWrap(_ string, _ pr.Shortand, tokens []Token) ([]namedTokens, error) {
	keyword := overflowWrap(tokens, "")
	if keyword == nil {
		return nil, ErrInvalidValue
	}
	return []namedTokens{
		{name: pr.POverflowWrap, tokens: tokens},
	}, nil
}

// @expander("flex")
// Expand the “flex“ property.
func _expandFlex(_ string, _ pr.Shortand, tokens []Token) (out []namedTokens, err error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "none" {
		pos := tokens[0].Pos()
		zeroToken := pa.NewNumber(0, pos)
		autoToken := pa.NewIdent("auto", pos)
		out = append(out,
			namedTokens{name: pr.PFlexGrow, tokens: []Token{zeroToken}},
			namedTokens{name: pr.PFlexShrink, tokens: []Token{zeroToken}},
			namedTokens{name: pr.PFlexBasis, tokens: []Token{autoToken}},
		)
	} else {
		var (
			grow   utils.Fl = 1
			shrink utils.Fl = 1
			basis  Token
		)
		growFound, shrinkFound, basisFound := false, false, false
		for _, token := range tokens {
			// "A unitless zero that is not already preceded by two flex factors
			// must be interpreted as a flex factor."
			number, ok := token.(pa.Number)
			forcedFlexFactor := ok && number.Int() == 0 && !(growFound && shrinkFound)
			if !basisFound && !forcedFlexFactor {
				newBasis := flexBasis([]Token{token}, "")
				if newBasis != nil {
					basis = token
					basisFound = true
					continue
				}
			}
			if !growFound {
				newGrow, ok := _flexGrowShrink([]Token{token})
				if !ok {
					return nil, ErrInvalidValue
				} else {
					grow = newGrow
					growFound = true
					continue
				}
			} else if !shrinkFound {
				newShrink, ok := _flexGrowShrink([]Token{token})
				if !ok {
					return nil, ErrInvalidValue
				} else {
					shrink = newShrink
					shrinkFound = true
					continue
				}
			} else {
				return nil, ErrInvalidValue
			}
		}
		pos := tokens[0].Pos()
		growToken := pa.NewNumber(grow, pos)
		shrinkToken := pa.NewNumber(shrink, pos)
		if !basisFound {
			basis = pa.NewDimension(pa.NewNumber(0, pos), "px")
		}
		out = []namedTokens{
			{name: pr.PFlexGrow, tokens: []Token{growToken}},
			{name: pr.PFlexShrink, tokens: []Token{shrinkToken}},
			{name: pr.PFlexBasis, tokens: []Token{basis}},
		}
	}
	return out, nil
}

// @expander("flex-flow")
// Expand the “flex-flow“ property.
func _expandFlexFlow(_ string, _ pr.Shortand, tokens []Token) (out []namedTokens, err error) {
	if len(tokens) == 2 {
		hasBroken := false
		for _, sortedTokens := range [2][]Token{tokens, reverse(tokens)} {
			direction := flexDirection(sortedTokens[0:1], "")
			wrap := flexWrap(sortedTokens[1:2], "")
			if direction != nil && wrap != nil {
				out = append(out, namedTokens{name: pr.PFlexDirection, tokens: sortedTokens[0:1]})
				out = append(out, namedTokens{name: pr.PFlexWrap, tokens: sortedTokens[1:2]})
				hasBroken = true
				break
			}
		}
		if !hasBroken {
			return nil, ErrInvalidValue
		}
	} else if len(tokens) == 1 {
		direction := flexDirection(tokens[0:1], "")
		if direction != nil {
			out = append(out, namedTokens{name: pr.PFlexDirection, tokens: tokens[0:1]})
		} else {
			wrap := flexWrap(tokens[0:1], "")
			if wrap != nil {
				out = append(out, namedTokens{name: pr.PFlexWrap, tokens: tokens[0:1]})
			} else {
				return nil, ErrInvalidValue
			}
		}
	} else {
		return nil, ErrInvalidValue
	}
	return out, nil
}

// return tokens for "columns", "rows", "areas" (or a zero value)
func expandGridTemplateImpl(tokens []Token) ([3][]Token, error) {
	none := pa.NewIdent("none", tokens[0].Pos())
	if len(tokens) == 1 && getKeyword(tokens[0]) == "none" {
		return [...][]Token{{none}, {none}, {none}}, nil
	}
	chunks := [][]Token{{}}
	for _, token := range tokens {
		if lit, ok := token.(pa.Literal); ok && lit.Value == "/" {
			chunks = append(chunks, nil)
		} else {
			chunks[len(chunks)-1] = append(chunks[len(chunks)-1], token)
		}
	}
	var columns []Token
	if len(chunks) == 2 {
		_, okR := gridTemplateImpl(chunks[0])
		_, okC := gridTemplateImpl(chunks[1])
		if okC {
			if okR {
				return [...][]Token{chunks[1], chunks[0], {none}}, nil
			}
			columns = chunks[1]
		} else {
			return [3][]Token{}, ErrInvalidValue
		}
	} else if len(chunks) == 1 {
		columns = []Token{none}
	} else {
		return [3][]Token{}, ErrInvalidValue
	}
	// TODO: Handle last syntax.
	_ = columns
	return [3][]Token{}, ErrInvalidValue
}

// @expander("grid-template")
// @generic_expander("-columns", "-rows", "-areas")
// Expand the “grid-template“ property.
func _expandGridTemplate(_ string, _ pr.Shortand, tokens []Token) (out []namedTokens, _ error) {
	v, err := expandGridTemplateImpl(tokens)
	if err != nil {
		return nil, err
	}
	return []namedTokens{{pr.PGridTemplateColumns, v[0]}, {pr.PGridTemplateRows, v[1]}, {pr.PGridTemplateAreas, v[2]}}, nil
}

// @expander("grid")
// @generic_expander("-template-columns", "-template-rows", "-template-areas", "-auto-columns", "-auto-rows", "-auto-flow")
//
// Expand the “grid“ property.
func _expandGrid(_ string, _ pr.Shortand, tokens []Token) (out []namedTokens, _ error) {
	pos := tokens[0].Pos()
	auto := pa.NewIdent("auto", pos)
	none := pa.NewIdent("none", pos)
	row := pa.NewIdent("row", pos)
	column := pa.NewIdent("column", pos)

	template, err := expandGridTemplateImpl(tokens)
	if err == nil {
		out = append(out, namedTokens{pr.PGridTemplateColumns, template[0]}, namedTokens{pr.PGridTemplateRows, template[1]}, namedTokens{pr.PGridTemplateAreas, template[2]},
			namedTokens{pr.PGridAutoColumns, []Token{auto}}, namedTokens{pr.PGridAutoRows, []Token{auto}}, namedTokens{pr.PGridAutoFlow, []Token{row}})
		return
	}

	chunks := [][]Token{{}}
	for _, token := range tokens {
		if lit, ok := token.(pa.Literal); ok && lit.Value == "/" {
			chunks = append(chunks, nil)
			continue
		}
		chunks[len(chunks)-1] = append(chunks[len(chunks)-1], token)
	}
	if len(chunks) != 2 {
		return nil, ErrInvalidValue
	}

	var (
		autoTrack = -1
		dense     Token
		templates [2][]Token // "row", "column"
	)
	const (
		rowT    = 0
		columnT = 1
	)
	for track, tokens := range chunks {
		for _, token := range tokens {
			if getKeyword(token) == "dense" {
				if dense != nil || (autoTrack != -1 && autoTrack != track) {
					return nil, ErrInvalidValue
				}
				dense = token
			} else if getKeyword(token) == "auto-flow" {
				if autoTrack != -1 {
					return nil, ErrInvalidValue
				}
				autoTrack = track
			} else {
				templates[track] = append(templates[track], token)
			}
		}
	}
	if autoTrack == -1 {
		return nil, ErrInvalidValue
	}

	nonAutoTrack := columnT
	autoTrackToken := row
	if autoTrack == columnT {
		nonAutoTrack = rowT
		autoTrackToken = column
	}

	val := []Token{autoTrackToken}
	if dense != nil {
		val = []Token{autoTrackToken, dense}
	}

	names := [2]string{rowT: "row", columnT: "column"}
	return []namedTokens{
		{pr.PGridAutoFlow, val},
		{pr.PropsFromNames[fmt.Sprintf("grid-auto-%ss", names[autoTrack])], templates[autoTrack]},
		{pr.PropsFromNames[fmt.Sprintf("grid-auto-%ss", names[nonAutoTrack])], []Token{auto}},
		{pr.PropsFromNames[fmt.Sprintf("grid-template-%ss", names[autoTrack])], []Token{none}},
		{pr.PropsFromNames[fmt.Sprintf("grid-template-%ss", names[nonAutoTrack])], templates[nonAutoTrack]},
		{pr.PGridTemplateAreas, []Token{none}},
	}, nil
}

func expandGridColumnRowArea(tokens []Token, maxNumber int) (out [][]Token, _ error) {
	gridLines := [][]Token{{}}
	for _, token := range tokens {
		if lit, ok := token.(pa.Literal); ok && lit.Value == "/" {
			gridLines = append(gridLines, nil)
			continue
		}
		gridLines[len(gridLines)-1] = append(gridLines[len(gridLines)-1], token)
	}
	if !(1 <= len(gridLines) && len(gridLines) <= maxNumber) {
		return nil, ErrInvalidValue
	}
	var validations []pr.GridLine
	for _, tokens := range gridLines {
		validation, ok := gridLineImpl(tokens)
		if !ok {
			return nil, ErrInvalidValue
		}
		validations = append(validations, validation)
		out = append(out, tokens)
	}
	auto := pa.NewIdent("auto", tokens[0].Pos())
	lines := len(gridLines)
	if lines <= 1 {
		value := []Token{auto}
		if customIdent := validations[0].IsCustomIdent(); customIdent {
			value = gridLines[0]
		}
		gridLines = append(gridLines, tokens)
		validations = append(validations, validations[0])
		out = append(out, value)
	}
	if lines <= 2 && 2 < maxNumber {
		if customIdent := validations[0].IsCustomIdent(); customIdent {
			out = append(out, gridLines[0])
		} else {
			out = append(out, []Token{auto})
		}
	}
	if lines <= 3 && 3 < maxNumber {
		if customIdent := validations[1].IsCustomIdent(); customIdent {
			out = append(out, gridLines[1])
		} else {
			out = append(out, []Token{auto})
		}
	}
	return out, nil
}

// @expander('grid-column')
// @expander('grid-row')
// @generic_expander('-start', '-end')
// Expand the “grid-[column|row]“ properties.
func _expandGridColumnRow(_ string, shortand pr.Shortand, tokens []Token) (out []namedTokens, _ error) {
	tokens_list, err := expandGridColumnRowArea(tokens, 2)
	if err != nil {
		return nil, err
	}
	sides := [2]string{"-start", "-end"}
	for index, tokens := range tokens_list {
		out = append(out, namedTokens{name: pr.PropsFromNames[shortand.String()+sides[index]], tokens: tokens})
	}
	return
}

// @expander("grid-area")
// @generic_expander("grid-row-start", "grid-row-end", "grid-column-start", "grid-column-end")
// Expand the “grid-area“ property.
func _expandGridArea(_ string, _ pr.Shortand, tokens []Token) (out []namedTokens, _ error) {
	tokens_list, err := expandGridColumnRowArea(tokens, 4)
	if err != nil {
		return nil, err
	}
	sides := [4]string{"row-start", "row-end", "column-start", "column-end"}
	for index, tokens := range tokens_list {
		out = append(out, namedTokens{name: pr.PropsFromNames["grid-"+sides[index]], tokens: tokens})
	}
	return
}

// @expander('line-clamp')
// Expand the “line-clamp“ property.
func _expandLineClamp(_ string, _ pr.Shortand, tokens []Token) (out []namedTokens, err error) {
	if len(tokens) == 1 {
		keyword := getSingleKeyword(tokens)
		if keyword == "none" {
			pos := tokens[0].Pos()
			noneToken := pa.NewIdent("none", pos)
			autoToken := pa.NewIdent("auto", pos)
			return []namedTokens{
				{name: pr.PMaxLines, tokens: []Token{noneToken}},
				{name: pr.PContinue, tokens: []Token{autoToken}},
				{name: pr.PBlockEllipsis, tokens: []Token{noneToken}},
			}, nil
		} else if nb, ok := tokens[0].(pa.Number); ok && nb.IsInt() {
			pos := tokens[0].Pos()
			autoToken := pa.NewIdent("auto", pos)
			discardToken := pa.NewIdent("discard", pos)
			return []namedTokens{
				{name: pr.PMaxLines, tokens: tokens[0:1]},
				{name: pr.PContinue, tokens: []Token{discardToken}},
				{name: pr.PBlockEllipsis, tokens: []Token{autoToken}},
			}, nil
		}
	} else if len(tokens) == 2 {
		if nb, ok := tokens[0].(pa.Number); ok {
			maxLines := nb.Int()
			_, valid := blockEllipsis_(tokens[1:2])
			if maxLines != 0 && valid {
				pos := tokens[0].Pos()
				discardToken := pa.NewIdent("discard", pos)
				return []namedTokens{
					{name: pr.PMaxLines, tokens: tokens[0:1]},
					{name: pr.PContinue, tokens: []Token{discardToken}},
					{name: pr.PBlockEllipsis, tokens: tokens[1:2]},
				}, nil
			}
		}
	}
	return nil, ErrInvalidValue
}

// Expand the “text-align“ property.
func _expandTextAlign(_ string, _ pr.Shortand, tokens []Token) (out []namedTokens, err error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "" {
		return nil, ErrInvalidValue
	}

	pos := tokens[0].Pos()

	alignAll := tokens[0]
	if keyword == "justify-all" {
		alignAll = pa.NewIdent("justify", pos)
	}
	alignLast := alignAll
	if keyword == "justify" {
		alignLast = pa.NewIdent("start", pos)
	}

	return []namedTokens{
		{name: pr.PTextAlignAll, tokens: []Token{alignAll}},
		{name: pr.PTextAlignLast, tokens: []Token{alignLast}},
	}, nil
}
