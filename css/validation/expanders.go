package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/utils"
)

var expanders = map[string]expander{
	"border-color": expandFourSides,
	"border-style": expandFourSides,
	"border-width": expandFourSides,
	"margin":       expandFourSides,
	"padding":      expandFourSides,
	"bleed":        expandFourSides,
	"border-radius": genericExpander(
		"border-top-left-radius", "border-top-right-radius",
		"border-bottom-right-radius", "border-bottom-left-radius")(_borderRadius),
	"page-break-after":  genericExpander("break-after")(_expandPageBreakBeforeAfter),
	"page-break-before": genericExpander("break-before")(_expandPageBreakBeforeAfter),
	"page-break-inside": genericExpander("break-inside")(_expandPageBreakInside),
	"background":        expandBackground,
	"word-wrap":         genericExpander("overflow-wrap")(_expandWordWrap),
	"list-style":        genericExpander("-type", "-position", "-image")(_expandListStyle),
	"border":            expandBorder,
	"border-top":        expandBorderSide,
	"border-right":      expandBorderSide,
	"border-bottom":     expandBorderSide,
	"border-left":       expandBorderSide,
	"column-rule":       expandBorderSide,
	"outline":           expandBorderSide,
	"columns":           genericExpander("column-width", "column-count")(_expandColumns),
	"font-variant": genericExpander("-alternates", "-caps", "-east-asian", "-ligatures",
		"-numeric", "-position")(_fontVariant),
	"font": genericExpander("-style", "-variant-caps", "-weight", "-stretch", "-size",
		"line-height", "-family")(_expandFont),
	"text-decoration": genericExpander("-line", "-color", "-style")(_expandTextDecoration),
	"flex":            genericExpander("-grow", "-shrink", "-basis")(_expandFlex),
	"flex-flow":       genericExpander("flex-direction", "flex-wrap")(_expandFlexFlow),
	"line-clamp":      genericExpander("max-lines", "continue", "block-ellipsis")(_expandLineClamp),
	"text-align":      genericExpander("-all", "-last")(_expandTextAlign),
}

var expandBorderSide = genericExpander("-width", "-color", "-style")(_expandBorderSide)

// Expanders

type NamedTokens struct {
	Name   string
	Tokens []parser.Token
}

type beforeGeneric = func(baseUrl, name string, tokens []parser.Token) ([]NamedTokens, error)

func defaultFromString(keyword string) pr.DefaultKind {
	val := pr.Inherit
	if keyword == "initial" {
		val = pr.Initial
	}
	return val
}

// Decorator helping expanders to handle ``inherit`` && ``initial``.
// Wrap an expander so that it does not have to handle the "inherit" and
// "initial" cases, and can just yield name suffixes. Missing suffixes
// get the initial value.
//
func genericExpander(expandedNames ...string) func(beforeGeneric) expander {
	_expandedNames := utils.Set{}
	for _, name := range expandedNames {
		_expandedNames.Add(name)
	}
	// Decorate the ``wrapped`` expander.
	genericExpanderDecorator := func(wrapped beforeGeneric) expander {
		// Wrap the expander.
		genericExpanderWrapper := func(baseUrl, name string, tokens []parser.Token) (out pr.NamedProperties, err error) {
			keyword := getSingleKeyword(tokens)
			results, toBeValidated := map[string]pr.ValidatedProperty{}, map[string][]parser.Token{}
			var skipValidation bool
			if keyword == "inherit" || keyword == "initial" {
				val := defaultFromString(keyword).AsCascaded().AsValidated()
				for _, name := range expandedNames {
					results[name] = val
				}
				skipValidation = true
			} else {
				skipValidation = false

				result, err := wrapped(baseUrl, name, tokens)
				if err != nil {
					return nil, err
				}

				for _, nameToken := range result {
					newName, newToken := nameToken.Name, nameToken.Tokens
					if !_expandedNames.Has(newName) {
						return nil, fmt.Errorf("unknown expanded property %s", newName)
					}
					if _, isIn := toBeValidated[newName]; isIn {
						return nil, fmt.Errorf("got multiple %s values in a %s shorthand",
							strings.Trim(newName, "-"), name)
					}
					toBeValidated[newName] = newToken
				}
			}

			for _, newName := range expandedNames {
				actualNewName := newName
				if strings.HasPrefix(newName, "-") {
					// newName is a suffix
					actualNewName = name + newName
				}
				var (
					value pr.ValidatedProperty
					in    bool
				)
				if skipValidation { // toBeValidated is empty -> ignore it
					value, in = results[newName]
				} else { // results is empty -> ignore it
					tokens, in = toBeValidated[newName]
					if in {
						np, err := validateNonShorthand(baseUrl, actualNewName, tokens, true)
						if err != nil {
							return nil, fmt.Errorf("validating %s: %s", actualNewName, err)
						}
						actualNewName = np.Name
						value = np.Property
					}
				}
				if !in {
					value = pr.Initial.AsCascaded().AsValidated()
				}

				out = append(out, pr.NamedProperty{Name: actualNewName, Property: value})
			}
			return out, nil
		}
		return genericExpanderWrapper
	}
	return genericExpanderDecorator
}

//@expander("border-color")
//@expander("border-style")
//@expander("border-width")
//@expander("margin")
//@expander("padding")
//@expander("bleed")
// Expand properties setting a token for the four sides of a box.
func expandFourSides(baseUrl, name string, tokens []parser.Token) (out pr.NamedProperties, err error) {
	// Make sure we have 4 tokens
	if len(tokens) == 1 {
		tokens = []parser.Token{tokens[0], tokens[0], tokens[0], tokens[0]}
	} else if len(tokens) == 2 {
		tokens = []parser.Token{tokens[0], tokens[1], tokens[0], tokens[1]} // (bottom, left) defaults to (top, right)
	} else if len(tokens) == 3 {
		tokens = append(tokens, tokens[1]) // left defaults to right
	} else if len(tokens) != 4 {
		return out, fmt.Errorf("Expected 1 to 4 token components got %d", len(tokens))
	}
	var newName string
	for index, suffix := range [4]string{"-top", "-right", "-bottom", "-left"} {
		token := tokens[index]
		i := strings.LastIndex(name, "-")
		if i == -1 {
			newName = name + suffix
		} else {
			// eg. border-color becomes border-*-color, not border-color-*
			newName = name[:i] + suffix + name[i:]
		}
		prop, err := validateNonShorthand(baseUrl, newName, []parser.Token{token}, true)
		if err != nil {
			return out, err
		}
		out = append(out, prop)
	}
	return out, nil
}

// Validator for the `border-radius` property.
func _borderRadius(_, _ string, tokens []parser.Token) (out []NamedTokens, err error) {
	var horizontal, vertical []parser.Token
	current := &horizontal

	for index, token := range tokens {
		if lit, ok := token.(parser.LiteralToken); ok && lit.Value == "/" {
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
		vertical = append([]parser.Token{}, horizontal...)
	}

	for _, values := range [2]*[]parser.Token{&horizontal, &vertical} {
		// Make sure we have 4 tokens
		if len(*values) == 1 {
			*values = []parser.Token{(*values)[0], (*values)[0], (*values)[0], (*values)[0]}
		} else if len(*values) == 2 {
			*values = []parser.Token{(*values)[0], (*values)[1], (*values)[0], (*values)[1]} // (br, bl) defaults to (tl, tr)
		} else if len(*values) == 3 {
			*values = append(*values, (*values)[1]) // bl defaults to tr
		} else if len(*values) != 4 {
			return nil, fmt.Errorf("expected 1 to 4 token components got %d", len(*values))
		}
	}
	corners := [4]string{"top-left", "top-right", "bottom-right", "bottom-left"}
	for index, corner := range corners {
		newName := fmt.Sprintf("border-%s-radius", corner)
		ts := []parser.Token{horizontal[index], vertical[index]}
		out = append(out, NamedTokens{Name: newName, Tokens: ts})
	}
	return out, nil
}

//@expander("list-style")
//@genericExpander("-type", "-position", "-image", wantsBaseUrl=true)
// Expand the ``list-style`` shorthand property.
//     See http://www.w3.org/TR/CSS21/generate.html#propdef-list-style
//
func _expandListStyle(baseUrl, _ string, tokens []parser.Token) (out []NamedTokens, err error) {
	var typeSpecified, imageSpecified bool
	noneCount := 0
	var noneToken parser.Token
	for _, token := range tokens {
		var suffix string
		if getKeyword(token) == "none" {
			// Can be either -style || -image, see at the end which is not
			// otherwise specified.
			noneCount += 1
			noneToken = token
			continue
		}

		if image, _ := listStyleImage([]parser.Token{token}, baseUrl); image != nil {
			suffix = "-image"
			imageSpecified = true
		} else if listStylePosition([]parser.Token{token}, "") != nil {
			suffix = "-position"
		} else if _, ok := listStyleType_([]parser.Token{token}); ok {
			suffix = "-type"
			typeSpecified = true
		} else {
			return nil, InvalidValue
		}
		out = append(out, NamedTokens{Name: suffix, Tokens: []parser.Token{token}})
	}

	if !typeSpecified && noneCount > 0 {
		out = append(out, NamedTokens{Name: "-type", Tokens: []parser.Token{noneToken}})
		noneCount -= 1
	}

	if !imageSpecified && noneCount > 0 {
		out = append(out, NamedTokens{Name: "-image", Tokens: []parser.Token{noneToken}})
		noneCount -= 1
	}

	if noneCount > 0 {
		// Too many none tokens.
		return nil, InvalidValue
	}
	return out, nil
}

//@expander("border")
// Expand the ``border`` shorthand property.
//     See http://www.w3.org/TR/CSS21/box.html#propdef-border
//
func expandBorder(baseUrl, name string, tokens []parser.Token) (out pr.NamedProperties, err error) {
	for _, suffix := range [4]string{"-top", "-right", "-bottom", "-left"} {
		props, err := expandBorderSide(baseUrl, name+suffix, tokens)
		if err != nil {
			return nil, err
		}
		out = append(out, props...)
	}
	return out, nil
}

//@expander("border-top")
//@expander("border-right")
//@expander("border-bottom")
//@expander("border-left")
//@expander("column-rule")
//@expander("outline")
//@genericExpander("-width", "-color", "-style")
// Expand the ``border-*`` shorthand pr.
//     See http://www.w3.org/TR/CSS21/box.html#propdef-border-top
//
func _expandBorderSide(_, _ string, tokens []parser.Token) ([]NamedTokens, error) {
	out := make([]NamedTokens, len(tokens))
	for index, token := range tokens {
		var suffix string
		if !parser.ParseColor(token).IsNone() {
			suffix = "-color"
		} else if borderWidth([]parser.Token{token}, "") != nil {
			suffix = "-width"
		} else if borderStyle([]parser.Token{token}, "") != nil {
			suffix = "-style"
		} else {
			return nil, InvalidValue
		}
		out[index] = NamedTokens{Name: suffix, Tokens: []parser.Token{token}}
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

func reverseLayers(a [][]parser.Token) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}

// Expand the ``background`` shorthand property.
// See http://dev.w3.org/csswg/css3-background/#the-background
func expandBackground(baseUrl, _ string, tokens []parser.Token) (out pr.NamedProperties, err error) {
	properties := [8]string{
		"background_color", "background_image", "background_repeat",
		"background_attachment", "background_position", "background_size",
		"background_clip", "background_origin",
	}
	keyword := getSingleKeyword(tokens)
	if keyword == "initial" || keyword == "inherit" {
		val := defaultFromString(keyword)
		for _, name := range properties {
			out = append(out, pr.NamedProperty{Name: name, Property: val.AsCascaded().AsValidated()})
		}
		return
	}

	parseLayer := func(tokens []parser.Token, finalLayer bool) (pr.CssProperty, backgroundProps, error) {
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

			token := tokens[len(tokens)-1:]
			if finalLayer {
				color := otherColors(token, "")
				if color != nil {
					if err = results.add("color"); err != nil {
						return pr.Color{}, backgroundProps{}, err
					}
					results.color = color
					tokens = tokens[:len(tokens)-1]
					continue
				}
			}

			image, err := _backgroundImage(token, baseUrl)
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

			repeat = _backgroundRepeat(token)
			if repeat != [2]string{} {
				if err = results.add("repeat"); err != nil {
					return pr.Color{}, backgroundProps{}, err
				}
				results.repeat = repeat
				tokens = tokens[:len(tokens)-1]
				continue
			}

			attachment := _backgroundAttachment(token)
			if attachment != "" {
				if err = results.add("attachment"); err != nil {
					return pr.Color{}, backgroundProps{}, err
				}
				results.attachment = attachment
				tokens = tokens[:len(tokens)-1]
				continue
			}

			index := 4 - len(tokens)
			if index < 0 {
				index = 0
			}
			var position pr.Center
			for _, n := range []int{4, 3, 2, 1}[index:] {
				nTokens := reverse(tokens[len(tokens)-n:])
				position = parsePosition(nTokens)
				if !position.IsNone() {
					if err = results.add("position"); err != nil {
						return pr.Color{}, backgroundProps{}, err
					}
					results.position = position
					tokens = tokens[:len(tokens)-n]
					if len(tokens) > 0 {
						if lit, ok := tokens[len(tokens)-1].(parser.LiteralToken); ok && lit.Value == "/" {
							index := 2 - len(tokens)
							if index < 0 {
								index = 0
							}
							for _, n := range []int{3, 2}[index:] {
								// n includes the "/" delimiter.
								i, j := utils.MaxInt(0, len(tokens)-n), utils.MaxInt(0, len(tokens)-1)
								nTokens = reverse(tokens[i:j])
								size := _backgroundSize(nTokens)
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

			origin := _box(token)
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
					clip := _box(token)
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
			return pr.Color{}, backgroundProps{}, InvalidValue
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

	layers := SplitOnComma(tokens)
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

	out = pr.NamedProperties{
		{Name: "background_image", Property: pr.AsCascaded(resultsImages).AsValidated()},
		{Name: "background_repeat", Property: pr.AsCascaded(resultsRepeats).AsValidated()},
		{Name: "background_attachment", Property: pr.AsCascaded(resultsAttachments).AsValidated()},
		{Name: "background_position", Property: pr.AsCascaded(resultsPositions).AsValidated()},
		{Name: "background_size", Property: pr.AsCascaded(resultsSizes).AsValidated()},
		{Name: "background_clip", Property: pr.AsCascaded(resultsClips).AsValidated()},
		{Name: "background_origin", Property: pr.AsCascaded(resultsOrigins).AsValidated()},
		{Name: "background-color", Property: pr.AsCascaded(resultColor).AsValidated()},
	}
	return out, nil
}

// @expander("text-decoration")
func _expandTextDecoration(_, _ string, tokens []parser.Token) (out []NamedTokens, err error) {
	var (
		textDecorationLine  []Token
		textDecorationColor []Token
		textDecorationStyle []Token
		noneInLine          bool
	)

	for _, token := range tokens {
		keyword := getKeyword(token)
		switch keyword {
		case "none", "underline", "overline", "line-through", "blink":
			textDecorationLine = append(textDecorationLine, token)
			if noneInLine {
				return nil, InvalidValue
			} else if keyword == "none" {
				noneInLine = true
			}
		case "solid", "double", "dotted", "dashed", "wavy":
			if len(textDecorationStyle) != 0 {
				return nil, InvalidValue
			} else {
				textDecorationStyle = append(textDecorationStyle, token)
			}
		default:
			color := parser.ParseColor(token)
			if color.IsNone() {
				return nil, InvalidValue
			} else if len(textDecorationColor) != 0 {
				return nil, InvalidValue
			} else {
				textDecorationColor = append(textDecorationColor, token)
			}
		}
	}

	if len(textDecorationLine) != 0 {
		out = append(out, NamedTokens{Name: "-line", Tokens: textDecorationLine})
	}
	if len(textDecorationColor) != 0 {
		out = append(out, NamedTokens{Name: "-color", Tokens: textDecorationColor})
	}
	if len(textDecorationStyle) != 0 {
		out = append(out, NamedTokens{Name: "-style", Tokens: textDecorationStyle})
	}

	return out, nil
}

// Expand legacy ``page-break-before`` && ``page-break-after`` pr.
// See https://www.w3.org/TR/css-break-3/#page-break-properties
func _expandPageBreakBeforeAfter(_, name string, tokens []parser.Token) (out []NamedTokens, err error) {
	keyword := getSingleKeyword(tokens)
	splits := strings.SplitN(name, "-", 2)
	if len(splits) < 2 {
		return nil, fmt.Errorf("bad format for name %s : should contain '-' ", name)
	}
	newName := splits[1]
	if keyword == "auto" || keyword == "left" || keyword == "right" || keyword == "avoid" {
		out = append(out, NamedTokens{Name: newName, Tokens: tokens})
	} else if keyword == "always" {
		out = append(out, NamedTokens{Name: newName, Tokens: []Token{parser.IdentToken{
			Value: "page",
			Pos:   tokens[0].Position(),
		}}})
	} else {
		return nil, InvalidValue
	}
	return out, nil
}

// Expand the legacy ``page-break-inside`` property.
// See https://www.w3.org/TR/css-break-3/#page-break-properties
func _expandPageBreakInside(_, _ string, tokens []parser.Token) ([]NamedTokens, error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "auto" || keyword == "avoid" {
		return []NamedTokens{{Name: "break-inside", Tokens: tokens}}, nil
	}

	return nil, InvalidValue
}

// Expand the ``columns`` shorthand property.
func _expandColumns(_, _ string, tokens []parser.Token) (out []NamedTokens, err error) {
	if len(tokens) == 2 && getKeyword(tokens[0]) == "auto" {
		tokens = reverse(tokens)
	}
	name := ""
	for _, token := range tokens {
		l := []parser.Token{token}
		if columnWidth(l, "") != nil && name != "column-width" {
			name = "column-width"
		} else if columnCount(l, "") != nil {
			name = "column-count"
		} else {
			return nil, InvalidValue
		}
		out = append(out, NamedTokens{Name: name, Tokens: l})
	}

	if len(tokens) == 1 {
		if name == "column-width" {
			name = "column-count"
		} else {
			name = "column-width"
		}
		out = append(out, NamedTokens{Name: name, Tokens: []Token{
			parser.IdentToken{Value: "auto", Pos: tokens[0].Position()},
		}})
	}
	return out, nil
}

var (
	noneFakeToken   = parser.IdentToken{Value: "none"}
	normalFakeToken = parser.IdentToken{Value: "normal"}
)

// Expand the ``font-variant`` shorthand property.
// https://www.w3.org/TR/css-fonts-3/#font-variant-prop
func _fontVariant(_, name string, tokens []parser.Token) (out []NamedTokens, err error) {
	return expandFontVariant(tokens)
}

func expandFontVariant(tokens []parser.Token) (out []NamedTokens, err error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "normal" || keyword == "none" {
		out = make([]NamedTokens, 6)
		for index, suffix := range [5]string{
			"-alternates", "-caps", "-east-asian", "-numeric",
			"-position",
		} {
			out[index] = NamedTokens{Name: suffix, Tokens: []parser.Token{normalFakeToken}}
		}
		token := noneFakeToken
		if keyword == "normal" {
			token = normalFakeToken
		}
		out[5] = NamedTokens{Name: "-ligatures", Tokens: []parser.Token{token}}
	} else {
		features := map[string][]parser.Token{}
		featuresKeys := [6]string{"alternates", "caps", "east-asian", "ligatures", "numeric", "position"}
		for _, token := range tokens {
			keyword := getKeyword(token)
			if keyword == "normal" {
				// We don"t allow "normal", only the specific values
				return nil, errors.New("invalid : normal not allowed")
			}
			found := false
			for _, feature := range featuresKeys {
				if fontVariantMapper[feature]([]parser.Token{token}, "") != nil {
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
				out = append(out, NamedTokens{Name: fmt.Sprintf("-%s", feature), Tokens: tokens})
			}
		}
	}
	return out, nil
}

var fontVariantMapper = map[string]func(tokens []parser.Token, _ string) pr.CssProperty{
	"alternates": fontVariantAlternates,
	"caps":       fontVariantCaps,
	"east-asian": fontVariantEastAsian,
	"ligatures":  fontVariantLigatures,
	"numeric":    fontVariantNumeric,
	"position":   fontVariantPosition,
}

// ExpandFont expands the 'font' property.
func ExpandFont(tokens []parser.Token) ([]NamedTokens, error) {
	l, err := _expandFont("", "", tokens)
	if err != nil {
		return nil, err
	}

	for i, v := range l {
		if strings.HasPrefix(v.Name, "-") { // newName is a suffix
			l[i].Name = "font" + v.Name
		}
	}
	return l, nil
}

// Expand the ``font`` shorthand property.
// https://www.w3.org/TR/css-fonts-3/#font-prop
func _expandFont(_, _ string, tokens []parser.Token) ([]NamedTokens, error) {
	expandFontKeyword := getSingleKeyword(tokens)
	if expandFontKeyword == "caption" || expandFontKeyword == "icon" || expandFontKeyword == "menu" || expandFontKeyword == "message-box" || expandFontKeyword ==
		"small-caption" || expandFontKeyword == "status-bar" {

		return nil, errors.New("system fonts are not supported")
	}
	var (
		out   []NamedTokens
		token parser.Token
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

		var suffix string
		if fontStyle([]parser.Token{token}, "") != nil {
			suffix = "-style"
		} else if kw == "normal" || kw == "small-caps" {
			suffix = "-variant-caps"
		} else if fontWeight([]parser.Token{token}, "") != nil {
			suffix = "-weight"
		} else if fontStretch([]parser.Token{token}, "") != nil {
			suffix = "-stretch"
		} else {
			// We’re done with these four, continue with font-size
			hasBroken = true
			break
		}
		out = append(out, NamedTokens{Name: suffix, Tokens: []parser.Token{token}})

		if len(tokens) == 0 {
			return nil, InvalidValue
		}
	}
	if !hasBroken {
		token, tokens = tokens[len(tokens)-1], tokens[:len(tokens)-1]
	}

	// Then font-size is mandatory
	// Latest `token` from the loop.
	fs, err := fontSize([]parser.Token{token}, "")
	if err != nil {
		return nil, err
	}
	if fs == nil {
		return nil, errors.New("invalid : font-size is mandatory for short font attribute !")
	}
	out = append(out, NamedTokens{Name: "-size", Tokens: []parser.Token{token}})

	// Then line-height is optional, but font-family is not so the list
	// must not be empty yet
	if len(tokens) == 0 {
		return nil, errors.New("invalid : font-familly is mandatory for short font attribute !")
	}

	token = tokens[len(tokens)-1]
	tokens = tokens[:len(tokens)-1]
	if lit, ok := token.(parser.LiteralToken); ok && lit.Value == "/" {
		token = tokens[len(tokens)-1]
		tokens = tokens[:len(tokens)-1]
		if lineHeight([]parser.Token{token}, "") == nil {
			return nil, InvalidValue
		}
		out = append(out, NamedTokens{Name: "line-height", Tokens: []parser.Token{token}})
	} else {
		// We pop()ed a font-family, add it back
		tokens = append(tokens, token)
	}
	// Reverse the stack to get normal list
	tokens = reverse(tokens)
	if fontFamily(tokens, "") == nil {
		return nil, InvalidValue
	}
	out = append(out, NamedTokens{Name: "-family", Tokens: tokens})
	return out, nil
}

// Expand the ``word-wrap`` legacy property.
// See http://http://www.w3.org/TR/css3-text/#overflow-wrap
func _expandWordWrap(_, _ string, tokens []parser.Token) ([]NamedTokens, error) {
	keyword := overflowWrap(tokens, "")
	if keyword == nil {
		return nil, InvalidValue
	}
	if len(tokens) == 1 {
		return []NamedTokens{
			{Name: "overflow-wrap", Tokens: tokens},
		}, nil
	}
	return nil, nil
}

// @expander("flex")
// Expand the ``flex`` property.
func _expandFlex(_, _ string, tokens []parser.Token) (out []NamedTokens, err error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "none" {
		pos := tokens[0].Position()
		zeroToken := parser.NumberToken{Value: 0, Representation: "0", IsInteger: true, Pos: pos}
		autoToken := parser.IdentToken{Value: "auto", Pos: pos}
		out = append(out,
			NamedTokens{Name: "-grow", Tokens: []Token{zeroToken}},
			NamedTokens{Name: "-shrink", Tokens: []Token{zeroToken}},
			NamedTokens{Name: "-basis", Tokens: []Token{autoToken}},
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
			number, ok := token.(parser.NumberToken)
			forcedFlexFactor := ok && number.IntValue() == 0 && !(growFound && shrinkFound)
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
					return nil, InvalidValue
				} else {
					grow = newGrow
					growFound = true
					continue
				}
			} else if !shrinkFound {
				newShrink, ok := _flexGrowShrink([]Token{token})
				if !ok {
					return nil, InvalidValue
				} else {
					shrink = newShrink
					shrinkFound = true
					continue
				}
			} else {
				return nil, InvalidValue
			}
		}
		pos := tokens[0].Position()
		growToken := parser.NewNumberToken(grow, pos)
		shrinkToken := parser.NewNumberToken(shrink, pos)
		if !basisFound {
			basis = parser.DimensionToken{
				Unit: "px",
				NumericToken: parser.NumericToken{
					Value:          0,
					Representation: "0",
					Pos:            pos,
					IsInteger:      true,
				},
			}
		}
		out = []NamedTokens{
			{Name: "-grow", Tokens: []Token{growToken}},
			{Name: "-shrink", Tokens: []Token{shrinkToken}},
			{Name: "-basis", Tokens: []Token{basis}},
		}
	}
	return out, nil
}

// @expander("flex-flow")
// Expand the ``flex-flow`` property.
func _expandFlexFlow(_, _ string, tokens []parser.Token) (out []NamedTokens, err error) {
	if len(tokens) == 2 {
		hasBroken := false
		for _, sortedTokens := range [2][]Token{tokens, reverse(tokens)} {
			direction := flexDirection(sortedTokens[0:1], "")
			wrap := flexWrap(sortedTokens[1:2], "")
			if direction != nil && wrap != nil {
				out = append(out, NamedTokens{Name: "flex-direction", Tokens: sortedTokens[0:1]})
				out = append(out, NamedTokens{Name: "flex-wrap", Tokens: sortedTokens[1:2]})
				hasBroken = true
				break
			}
		}
		if !hasBroken {
			return nil, InvalidValue
		}
	} else if len(tokens) == 1 {
		direction := flexDirection(tokens[0:1], "")
		if direction != nil {
			out = append(out, NamedTokens{Name: "flex-direction", Tokens: tokens[0:1]})
		} else {
			wrap := flexWrap(tokens[0:1], "")
			if wrap != nil {
				out = append(out, NamedTokens{Name: "flex-wrap", Tokens: tokens[0:1]})
			} else {
				return nil, InvalidValue
			}
		}
	} else {
		return nil, InvalidValue
	}
	return out, nil
}

// @expander('line-clamp')
// Expand the ``line-clamp`` property.
func _expandLineClamp(_, _ string, tokens []parser.Token) (out []NamedTokens, err error) {
	if len(tokens) == 1 {
		keyword := getSingleKeyword(tokens)
		if keyword == "none" {
			pos := tokens[0].Position()
			noneToken := parser.IdentToken{Value: "none", Pos: pos}
			autoToken := parser.IdentToken{Value: "auto", Pos: pos}
			return []NamedTokens{
				{Name: "max-lines", Tokens: []Token{noneToken}},
				{Name: "continue", Tokens: []Token{autoToken}},
				{Name: "block-ellipsis", Tokens: []Token{noneToken}},
			}, nil
		} else if nb, ok := tokens[0].(parser.NumberToken); ok && nb.IsInteger {
			pos := tokens[0].Position()
			autoToken := parser.IdentToken{Value: "auto", Pos: pos}
			discardToken := parser.IdentToken{Value: "discard", Pos: pos}
			return []NamedTokens{
				{Name: "max-lines", Tokens: tokens[0:1]},
				{Name: "continue", Tokens: []Token{discardToken}},
				{Name: "block-ellipsis", Tokens: []Token{autoToken}},
			}, nil
		}
	} else if len(tokens) == 2 {
		if nb, ok := tokens[0].(parser.NumberToken); ok {
			maxLines := nb.IntValue()
			_, valid := blockEllipsis_(tokens[1:2])
			if maxLines != 0 && valid {
				pos := tokens[0].Position()
				discardToken := parser.IdentToken{Value: "discard", Pos: pos}
				return []NamedTokens{
					{Name: "max-lines", Tokens: tokens[0:1]},
					{Name: "continue", Tokens: []Token{discardToken}},
					{Name: "block-ellipsis", Tokens: tokens[1:2]},
				}, nil
			}
		}
	}
	return nil, InvalidValue
}

// Expand the ``text-align`` property.
func _expandTextAlign(_, _ string, tokens []parser.Token) (out []NamedTokens, err error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "" {
		return nil, InvalidValue
	}

	pos := tokens[0].Position()

	alignAll := tokens[0]
	if keyword == "justify-all" {
		alignAll = parser.IdentToken{Value: "justify", Pos: pos}
	}
	alignLast := alignAll
	if keyword == "justify" {
		alignLast = parser.IdentToken{Value: "start", Pos: pos}
	}

	return []NamedTokens{
		{Name: "-all", Tokens: []Token{alignAll}},
		{Name: "-last", Tokens: []Token{alignLast}},
	}, nil
}
