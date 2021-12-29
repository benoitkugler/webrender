package parser

import (
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/benoitkugler/webrender/utils"
)

var (
	// ColorKeywords maps color names to RGBA values
	ColorKeywords = map[string]Color{}

	hashRegexps = []hashRegexp{
		{multiplier: 2., regexp: regexp.MustCompile("(?i)^([\\da-f])([\\da-f])([\\da-f])$")},
		{multiplier: 1., regexp: regexp.MustCompile("(?i)^([\\da-f]{2})([\\da-f]{2})([\\da-f]{2})$")},
	}

	// (r, g, b) := range 0..255
	basicColorKeywords = map[string][3]uint8{
		"black":   {0, 0, 0},
		"silver":  {192, 192, 192},
		"gray":    {128, 128, 128},
		"white":   {255, 255, 255},
		"maroon":  {128, 0, 0},
		"red":     {255, 0, 0},
		"purple":  {128, 0, 128},
		"fuchsia": {255, 0, 255},
		"green":   {0, 128, 0},
		"lime":    {0, 255, 0},
		"olive":   {128, 128, 0},
		"yellow":  {255, 255, 0},
		"navy":    {0, 0, 128},
		"blue":    {0, 0, 255},
		"teal":    {0, 128, 128},
		"aqua":    {0, 255, 255},
	}

	// (r, g, b) := range 0..255
	extendedColorKeywords = map[string][3]uint8{
		"aliceblue":            {240, 248, 255},
		"antiquewhite":         {250, 235, 215},
		"aqua":                 {0, 255, 255},
		"aquamarine":           {127, 255, 212},
		"azure":                {240, 255, 255},
		"beige":                {245, 245, 220},
		"bisque":               {255, 228, 196},
		"black":                {0, 0, 0},
		"blanchedalmond":       {255, 235, 205},
		"blue":                 {0, 0, 255},
		"blueviolet":           {138, 43, 226},
		"brown":                {165, 42, 42},
		"burlywood":            {222, 184, 135},
		"cadetblue":            {95, 158, 160},
		"chartreuse":           {127, 255, 0},
		"chocolate":            {210, 105, 30},
		"coral":                {255, 127, 80},
		"cornflowerblue":       {100, 149, 237},
		"cornsilk":             {255, 248, 220},
		"crimson":              {220, 20, 60},
		"cyan":                 {0, 255, 255},
		"darkblue":             {0, 0, 139},
		"darkcyan":             {0, 139, 139},
		"darkgoldenrod":        {184, 134, 11},
		"darkgray":             {169, 169, 169},
		"darkgreen":            {0, 100, 0},
		"darkgrey":             {169, 169, 169},
		"darkkhaki":            {189, 183, 107},
		"darkmagenta":          {139, 0, 139},
		"darkolivegreen":       {85, 107, 47},
		"darkorange":           {255, 140, 0},
		"darkorchid":           {153, 50, 204},
		"darkred":              {139, 0, 0},
		"darksalmon":           {233, 150, 122},
		"darkseagreen":         {143, 188, 143},
		"darkslateblue":        {72, 61, 139},
		"darkslategray":        {47, 79, 79},
		"darkslategrey":        {47, 79, 79},
		"darkturquoise":        {0, 206, 209},
		"darkviolet":           {148, 0, 211},
		"deeppink":             {255, 20, 147},
		"deepskyblue":          {0, 191, 255},
		"dimgray":              {105, 105, 105},
		"dimgrey":              {105, 105, 105},
		"dodgerblue":           {30, 144, 255},
		"firebrick":            {178, 34, 34},
		"floralwhite":          {255, 250, 240},
		"forestgreen":          {34, 139, 34},
		"fuchsia":              {255, 0, 255},
		"gainsboro":            {220, 220, 220},
		"ghostwhite":           {248, 248, 255},
		"gold":                 {255, 215, 0},
		"goldenrod":            {218, 165, 32},
		"gray":                 {128, 128, 128},
		"green":                {0, 128, 0},
		"greenyellow":          {173, 255, 47},
		"grey":                 {128, 128, 128},
		"honeydew":             {240, 255, 240},
		"hotpink":              {255, 105, 180},
		"indianred":            {205, 92, 92},
		"indigo":               {75, 0, 130},
		"ivory":                {255, 255, 240},
		"khaki":                {240, 230, 140},
		"lavender":             {230, 230, 250},
		"lavenderblush":        {255, 240, 245},
		"lawngreen":            {124, 252, 0},
		"lemonchiffon":         {255, 250, 205},
		"lightblue":            {173, 216, 230},
		"lightcoral":           {240, 128, 128},
		"lightcyan":            {224, 255, 255},
		"lightgoldenrodyellow": {250, 250, 210},
		"lightgray":            {211, 211, 211},
		"lightgreen":           {144, 238, 144},
		"lightgrey":            {211, 211, 211},
		"lightpink":            {255, 182, 193},
		"lightsalmon":          {255, 160, 122},
		"lightseagreen":        {32, 178, 170},
		"lightskyblue":         {135, 206, 250},
		"lightslategray":       {119, 136, 153},
		"lightslategrey":       {119, 136, 153},
		"lightsteelblue":       {176, 196, 222},
		"lightyellow":          {255, 255, 224},
		"lime":                 {0, 255, 0},
		"limegreen":            {50, 205, 50},
		"linen":                {250, 240, 230},
		"magenta":              {255, 0, 255},
		"maroon":               {128, 0, 0},
		"mediumaquamarine":     {102, 205, 170},
		"mediumblue":           {0, 0, 205},
		"mediumorchid":         {186, 85, 211},
		"mediumpurple":         {147, 112, 219},
		"mediumseagreen":       {60, 179, 113},
		"mediumslateblue":      {123, 104, 238},
		"mediumspringgreen":    {0, 250, 154},
		"mediumturquoise":      {72, 209, 204},
		"mediumvioletred":      {199, 21, 133},
		"midnightblue":         {25, 25, 112},
		"mintcream":            {245, 255, 250},
		"mistyrose":            {255, 228, 225},
		"moccasin":             {255, 228, 181},
		"navajowhite":          {255, 222, 173},
		"navy":                 {0, 0, 128},
		"oldlace":              {253, 245, 230},
		"olive":                {128, 128, 0},
		"olivedrab":            {107, 142, 35},
		"orange":               {255, 165, 0},
		"orangered":            {255, 69, 0},
		"orchid":               {218, 112, 214},
		"palegoldenrod":        {238, 232, 170},
		"palegreen":            {152, 251, 152},
		"paleturquoise":        {175, 238, 238},
		"palevioletred":        {219, 112, 147},
		"papayawhip":           {255, 239, 213},
		"peachpuff":            {255, 218, 185},
		"peru":                 {205, 133, 63},
		"pink":                 {255, 192, 203},
		"plum":                 {221, 160, 221},
		"powderblue":           {176, 224, 230},
		"purple":               {128, 0, 128},
		"red":                  {255, 0, 0},
		"rosybrown":            {188, 143, 143},
		"royalblue":            {65, 105, 225},
		"saddlebrown":          {139, 69, 19},
		"salmon":               {250, 128, 114},
		"sandybrown":           {244, 164, 96},
		"seagreen":             {46, 139, 87},
		"seashell":             {255, 245, 238},
		"sienna":               {160, 82, 45},
		"silver":               {192, 192, 192},
		"skyblue":              {135, 206, 235},
		"slateblue":            {106, 90, 205},
		"slategray":            {112, 128, 144},
		"slategrey":            {112, 128, 144},
		"snow":                 {255, 250, 250},
		"springgreen":          {0, 255, 127},
		"steelblue":            {70, 130, 180},
		"tan":                  {210, 180, 140},
		"teal":                 {0, 128, 128},
		"thistle":              {216, 191, 216},
		"tomato":               {255, 99, 71},
		"turquoise":            {64, 224, 208},
		"violet":               {238, 130, 238},
		"wheat":                {245, 222, 179},
		"white":                {255, 255, 255},
		"whitesmoke":           {245, 245, 245},
		"yellow":               {255, 255, 0},
		"yellowgreen":          {154, 205, 50},
	}

	// (r, g, b, a) in 0..1 or a string marker
	specialColorKeywords = map[string]Color{
		"currentcolor": {Type: ColorCurrentColor},
		"transparent":  {Type: ColorRGBA, RGBA: RGBA{R: 0., G: 0., B: 0., A: 0.}},
	}
)

func init() {
	for k, v := range specialColorKeywords {
		ColorKeywords[k] = v
	}
	// 255 maps to 1, 0 to 0, the rest is linear.
	for k, v := range basicColorKeywords {
		ColorKeywords[k] = Color{Type: ColorRGBA, RGBA: RGBA{utils.Fl(v[0]) / 255., utils.Fl(v[1]) / 255., utils.Fl(v[2]) / 255., 1.}}
	}
	for k, v := range extendedColorKeywords {
		ColorKeywords[k] = Color{Type: ColorRGBA, RGBA: RGBA{utils.Fl(v[0]) / 255., utils.Fl(v[1]) / 255., utils.Fl(v[2]) / 255., 1.}}
	}
}

// values in [-1, 1]
type RGBA struct {
	R, G, B, A utils.Fl
}

func (color RGBA) Unpack() (r, g, b, a utils.Fl) {
	return color.R, color.G, color.B, color.A
}

func (color RGBA) IsNone() bool {
	return color == RGBA{}
}

func clamp(v utils.Fl) utils.Fl {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the color. Each value ranges within [0, 0xffff], but is represented
// by a uint32 so that multiplying by a blend factor up to 0xffff will not
// overflow.
//
// An alpha-premultiplied color component c has been scaled by alpha (a),
// so has valid values 0 <= c <= a.
func (c RGBA) RGBA() (r, g, b, a uint32) {
	c.R, c.G, c.B, c.A = clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A)
	sa := 0xFFFF * c.A
	return uint32(c.R * sa), uint32(c.G * sa), uint32(c.B * sa), uint32(sa)
}

type ColorType uint8

const (
	// ColorInvalid is an empty or invalid color specification.
	ColorInvalid ColorType = iota
	// ColorCurrentColor represents the special value "currentColor"
	// which need document context to be resolved.
	ColorCurrentColor
	// ColorRGBA is a standard rgba color.
	ColorRGBA
)

type Color struct {
	Type ColorType
	RGBA RGBA
}

func (c Color) IsNone() bool {
	return c.Type == ColorInvalid
}

func round(f utils.Fl) myFloat {
	return myFloat(utils.Round(f))
}

func (c Color) toJson() jsonisable {
	switch c.Type {
	case ColorInvalid:
		return nil
	case ColorCurrentColor:
		return myString("currentColor")
	default:
		return jsonList{round(c.RGBA.R), round(c.RGBA.G), round(c.RGBA.B), round(c.RGBA.A)}
	}
}

type hashRegexp struct {
	regexp     *regexp.Regexp
	multiplier int
}

func mustParseHexa(s string) utils.Fl {
	out, err := strconv.ParseInt(s, 16, 0)
	if err != nil {
		log.Fatal(err)
	}
	return utils.Fl(out)
}

// ParseColorString tokenize the input before calling `ParseColor`.
func ParseColorString(color string) Color {
	l := parseOneComponentValueString(color, true)
	return ParseColor(l)
}

// Parse a color value as defined in `CSS Color Level 3  <http://www.w3.org/TR/css3-color/>`.
// Returns :
//  - zero Color if the input is not a valid color value. (No error is returned.)
//  - CurrentColor for the *currentColor* keyword
//  - RGBA color for every other values (including keywords, HSL && HSLA.)
//    The alpha channel is clipped to [0, 1] but red, green, or blue can be out of range
//    (eg. ``rgb(-10%, 120%, 0%)`` is represented as ``(-0.1, 1.2, 0, 1)``.
func ParseColor(_token Token) Color {
	switch token := _token.(type) {
	case IdentToken:
		return ColorKeywords[token.Value.Lower()]
	case HashToken:
		for _, hashReg := range hashRegexps {
			match := hashReg.regexp.FindStringSubmatch(token.Value)
			if len(match) == 4 {
				r := mustParseHexa(strings.Repeat(match[1], hashReg.multiplier)) / 255
				g := mustParseHexa(strings.Repeat(match[2], hashReg.multiplier)) / 255
				b := mustParseHexa(strings.Repeat(match[3], hashReg.multiplier)) / 255
				return Color{Type: ColorRGBA, RGBA: RGBA{R: r, G: g, B: b, A: 1.}}
			}
		}
	case FunctionBlock:
		args := parseCommaSeparated(*token.Arguments)
		if len(args) != 0 {
			switch token.Name.Lower() {
			case "rgb":
				rgba, ok := parseRgb(args, 1.)
				if ok {
					return Color{Type: ColorRGBA, RGBA: rgba}
				}
			case "rgba":
				if len(args) < 3 {
					return Color{}
				}
				alpha, isNotNone := parseAlpha(args[3:])
				if isNotNone {
					rgba, ok := parseRgb(args[:3], alpha)
					if ok {
						return Color{Type: ColorRGBA, RGBA: rgba}
					}
				}
			case "hsl":
				rgba, ok := parseHsl(args, 1.)
				if ok {
					return Color{Type: ColorRGBA, RGBA: rgba}
				}
			case "hsla":
				if len(args) < 3 {
					return Color{}
				}
				alpha, isNotNone := parseAlpha(args[3:])
				if isNotNone {
					rgba, ok := parseHsl(args[:3], alpha)
					if ok {
						return Color{Type: ColorRGBA, RGBA: rgba}
					}
				}
			}
		}
	}
	return Color{}
}

// If args is a list of a single  NUMBER token,
// return its value clipped to the 0..1 range
//
func parseAlpha(args []Token) (utils.Fl, bool) {
	if len(args) == 1 {
		token, ok := args[0].(NumberToken)
		if ok {
			return utils.MinF(1., utils.MaxF(0., token.Value)), true
		}
	}
	return 0, false
}

// If args is a list of 3 NUMBER tokens or 3 PERCENTAGE tokens,
// return RGB values as a tuple of 3 floats := range 0..1.
//
func parseRgb(args []Token, alpha utils.Fl) (RGBA, bool) {
	if len(args) != 3 {
		return RGBA{}, false
	}
	nR, okR := args[0].(NumberToken)
	nG, okG := args[1].(NumberToken)
	nB, okB := args[2].(NumberToken)
	if okR && okG && okB && nR.IsInteger && nG.IsInteger && nB.IsInteger {
		return RGBA{R: nR.Value / 255, G: nG.Value / 255, B: nB.Value / 255, A: alpha}, true
	}

	pR, okR := args[0].(PercentageToken)
	pG, okG := args[1].(PercentageToken)
	pB, okB := args[2].(PercentageToken)
	if okR && okG && okB {
		return RGBA{R: pR.Value / 100, G: pG.Value / 100, B: pB.Value / 100, A: alpha}, true
	}
	return RGBA{}, false
}

// If args is a list of 1 NUMBER token && 2 PERCENTAGE tokens,
// return RGB values as a tuple of 3 floats := range 0..1.
func parseHsl(args []Token, alpha utils.Fl) (RGBA, bool) {
	if len(args) != 3 {
		return RGBA{}, false
	}
	h, okH := args[0].(NumberToken)
	s, okS := args[1].(PercentageToken)
	l, okL := args[2].(PercentageToken)
	if okH && okS && okL && h.IsInteger {
		r, g, b := hslToRgb(h.IntValue(), s.Value, l.Value)
		return RGBA{R: r, G: g, B: b, A: alpha}, true
	}
	return RGBA{}, false
}

//  returns (r, g, b) as floats in the 0..1 range
func hslToRgb(_hue int, saturation, lightness utils.Fl) (utils.Fl, utils.Fl, utils.Fl) {
	hue := float64(_hue) / 360
	hue = hue - math.Floor(hue)
	saturation = utils.MinF(1., utils.MaxF(0, saturation/100))
	lightness = utils.MinF(1, utils.MaxF(0, lightness/100))

	// Translated from ABC: http://www.w3.org/TR/css3-color/#hsl-color
	hueToRgb := func(m1, m2, h float64) utils.Fl {
		if h < 0 {
			h += 1.
		}
		if h > 1 {
			h -= 1.
		}
		if h*6 < 1 {
			return utils.Fl(m1 + (m2-m1)*h*6)
		}
		if h*2 < 1 {
			return utils.Fl(m2)
		}
		if h*3 < 2 {
			return utils.Fl(m1 + (m2-m1)*(2./3-h)*6)
		}
		return utils.Fl(m1)
	}
	var m1, m2 float64
	if lightness <= 0.5 {
		m2 = float64(lightness * (saturation + 1.))
	} else {
		m2 = float64(lightness + saturation - lightness*saturation)
	}
	m1 = float64(lightness*2) - m2
	return hueToRgb(m1, m2, hue+1./3), hueToRgb(m1, m2, hue), hueToRgb(m1, m2, hue-1./3)
}

// Parse a list of tokens (typically the content of a function token)
// as arguments made of a single token each, separated by mandatory commas,
// with optional white space around each argument.
// return the argument list without commas or white space;
// or None if the function token content do not match the description above.
//
func parseCommaSeparated(tokens []Token) []Token {
	var filtered []Token
	for _, token := range tokens {
		if token.Type() != WhitespaceTokenT && token.Type() != CommentT {
			filtered = append(filtered, token)
		}
	}

	if len(filtered)%2 == 1 {
		others := []Token{filtered[0]}
		isAll := true
		for i := 1; i < len(filtered); i += 2 {
			token := filtered[i]
			others = append(others, filtered[i+1])
			litteral, ok := token.(LiteralToken)
			if !ok || litteral.Value != "," {
				isAll = false
				break
			}
		}

		if isAll {
			return others
		}
	}
	return nil
}
