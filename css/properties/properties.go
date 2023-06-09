package properties

import "github.com/benoitkugler/webrender/css/parser"

// This file is used to generate typed accessors
//go:generate go run gen/gen.go

const (
	_ KnownProp = iota
	PBottom
	PCaptionSide
	PClear
	PClip
	PColor
	PDirection
	PDisplay
	PEmptyCells
	PFloat
	PLeft
	PRight
	PLineHeight

	PPosition
	PTableLayout
	PTop
	PUnicodeBidi
	PVerticalAlign
	PVisibility
	PZIndex

	// The following properties are all background related,
	// in the order expected by expandBackground
	PBackgroundColor
	PBackgroundImage
	PBackgroundRepeat
	PBackgroundAttachment
	PBackgroundPosition
	PBackgroundSize
	PBackgroundClip
	PBackgroundOrigin

	// the following properties are grouped by side,
	// in the [bottom, left, right, top] order,
	// so that, if side in an index (0, 1, 2 or 3),
	// the property is a PBorderBottomColor + side * 5
	// DO NOT CHANGE the order, because
	PBorderBottomColor
	PBorderBottomStyle
	PBorderBottomWidth
	PMarginBottom
	PPaddingBottom

	PBorderLeftColor
	PBorderLeftStyle
	PBorderLeftWidth
	PMarginLeft
	PPaddingLeft

	PBorderRightColor
	PBorderRightStyle
	PBorderRightWidth
	PMarginRight
	PPaddingRight

	PBorderTopColor
	PBorderTopStyle
	PBorderTopWidth
	PMarginTop
	PPaddingTop

	PBorderCollapse
	PBorderSpacing
	PBorderBottomLeftRadius
	PBorderBottomRightRadius
	PBorderTopLeftRadius
	PBorderTopRightRadius
	POpacity
	PColumnWidth
	PColumnCount
	PColumnGap
	PColumnRuleColor
	PColumnRuleStyle
	PColumnRuleWidth
	PColumnFill
	PColumnSpan
	PFontFamily
	PFontFeatureSettings
	PFontKerning
	PFontLanguageOverride
	PFontSize
	PFontStretch
	PFontStyle
	PFontVariant
	PFontVariantAlternates
	PFontVariantCaps
	PFontVariantEastAsian
	PFontVariantLigatures
	PFontVariantNumeric
	PFontVariantPosition
	PFontWeight
	PFontVariationSettings
	PBoxDecorationBreak
	PBreakAfter
	PBreakBefore
	PBreakInside
	PMarginBreak
	POrphans
	PWidows
	PBookmarkLabel
	PBookmarkLevel
	PBookmarkState
	PContent
	PFootnoteDisplay
	PFootnotePolicy
	PQuotes
	PStringSet
	PImageResolution
	PImageRendering
	PImageOrientation
	PObjectFit
	PObjectPosition
	PSize
	PPage
	PBleedLeft
	PBleedRight
	PBleedTop
	PBleedBottom
	PMarks
	PHyphenateCharacter
	PHyphenateLimitChars
	PHyphenateLimitZone
	PHyphens
	PLetterSpacing
	PTabSize
	PTextAlignAll
	PTextAlignLast
	PTextIndent
	PTextTransform
	PWhiteSpace
	PWordBreak
	PWordSpacing
	PTransformOrigin
	PTransform
	PAppearance
	POutlineColor
	POutlineStyle
	POutlineWidth
	PBoxSizing
	PHeight
	PMaxHeight
	PMaxWidth
	PMinHeight
	PMinWidth
	PWidth
	PAlignContent
	PAlignItems
	PAlignSelf
	PFlexBasis
	PFlexDirection
	PFlexGrow
	PFlexShrink
	PFlexWrap
	PJustifyContent
	POrder

	// text-decoration-XXX
	PTextDecorationLine
	PTextDecorationColor
	PTextDecorationStyle

	PBlockEllipsis
	PContinue
	PMaxLines
	POverflow
	POverflowWrap
	PTextOverflow
	PCounterIncrement
	PCounterReset
	PCounterSet
	PListStyleImage
	PListStylePosition
	PListStyleType
	PAnchor
	PLink
	PLang
)

// InitialValues stores the default values for the CSS properties.
var InitialValues = Properties{
	// CSS 2.1: https://www.w3.org/TR/CSS21/propidx.html
	PBottom:      SToV("auto"),
	PCaptionSide: String("top"),
	PClear:       String("none"),
	PClip:        Values{},                                // computed value for "auto"
	PColor:       Color(parser.ParseColorString("black")), // chosen by the user agent

	PDirection:    String("ltr"),
	PDisplay:      Display{"inline", "flow"},
	PEmptyCells:   String("show"),
	PFloat:        String("none"),
	PLeft:         SToV("auto"),
	PRight:        SToV("auto"),
	PLineHeight:   SToV("normal"),
	PMarginTop:    zeroPixelsValue,
	PMarginRight:  zeroPixelsValue,
	PMarginBottom: zeroPixelsValue,
	PMarginLeft:   zeroPixelsValue,

	PPaddingTop:    zeroPixelsValue,
	PPaddingRight:  zeroPixelsValue,
	PPaddingBottom: zeroPixelsValue,
	PPaddingLeft:   zeroPixelsValue,
	PPosition:      BoolString{String: "static"},
	PTableLayout:   String("auto"),
	PTop:           SToV("auto"),
	PUnicodeBidi:   String("normal"),
	PVerticalAlign: SToV("baseline"),
	PVisibility:    String("visible"),
	PZIndex:        IntString{String: "auto"},

	// Backgrounds and Borders 3 (CR): https://www.w3.org/TR/css-backgrounds-3/
	PBackgroundAttachment: Strings{"scroll"},
	PBackgroundClip:       Strings{"border-box"},
	PBackgroundColor:      Color(parser.ParseColorString("transparent")),
	PBackgroundImage:      Images{NoneImage{}},
	PBackgroundOrigin:     Strings{"padding-box"},
	PBackgroundPosition: Centers{
		Center{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Unit: Perc}, Dimension{Unit: Perc}}},
	},
	PBackgroundRepeat:  Repeats{{"repeat", "repeat"}},
	PBackgroundSize:    Sizes{Size{Width: SToV("auto"), Height: SToV("auto")}},
	PBorderBottomColor: CurrentColor,
	PBorderLeftColor:   CurrentColor,
	PBorderRightColor:  CurrentColor,
	PBorderTopColor:    CurrentColor,
	PBorderBottomStyle: String("none"),
	PBorderLeftStyle:   String("none"),
	PBorderRightStyle:  String("none"),
	PBorderTopStyle:    String("none"),
	PBorderCollapse:    String("separate"),
	PBorderSpacing:     Point{Dimension{Unit: Scalar}, Dimension{Unit: Scalar}},
	PBorderBottomWidth: FToV(3),
	PBorderLeftWidth:   FToV(3),
	PBorderTopWidth:    FToV(3), // computed value for "medium"
	PBorderRightWidth:  FToV(3),

	PBorderBottomLeftRadius:  Point{ZeroPixels, ZeroPixels},
	PBorderBottomRightRadius: Point{ZeroPixels, ZeroPixels},
	PBorderTopLeftRadius:     Point{ZeroPixels, ZeroPixels},
	PBorderTopRightRadius:    Point{ZeroPixels, ZeroPixels},

	// Color 3 (REC): https://www.w3.org/TR/css-color-3/
	POpacity: Float(1),

	// Multi-column Layout (WD): https://www.w3.org/TR/css-multicol-1/
	PColumnWidth:     SToV("auto"),
	PColumnCount:     IntString{String: "auto"},
	PColumnGap:       Value{Dimension: Dimension{Value: 1, Unit: Em}},
	PColumnRuleColor: CurrentColor,
	PColumnRuleStyle: String("none"),
	PColumnRuleWidth: SToV("medium"),
	PColumnFill:      String("balance"),
	PColumnSpan:      String("none"),

	// Fonts 3 (REC): https://www.w3.org/TR/css-fonts-3/
	PFontFamily:            Strings{"serif"}, // depends on user agent
	PFontFeatureSettings:   SIntStrings{String: "normal"},
	PFontKerning:           String("auto"),
	PFontLanguageOverride:  String("normal"),
	PFontSize:              FToV(16), // actually medium, but we define medium from this
	PFontStretch:           String("normal"),
	PFontStyle:             String("normal"),
	PFontVariant:           String("normal"),
	PFontVariantAlternates: String("normal"),
	PFontVariantCaps:       String("normal"),
	PFontVariantEastAsian:  SStrings{String: "normal"},
	PFontVariantLigatures:  SStrings{String: "normal"},
	PFontVariantNumeric:    SStrings{String: "normal"},
	PFontVariantPosition:   String("normal"),
	PFontWeight:            IntString{Int: 400},

	// Fonts 4 (WD): https://www.w3.org/TR/css-fonts-4/
	PFontVariationSettings: SFloatStrings{String: "normal"},

	// Fragmentation 3/4 (CR/WD): https://www.w3.org/TR/css-break-4/
	PBoxDecorationBreak: String("slice"),
	PBreakAfter:         String("auto"),
	PBreakBefore:        String("auto"),
	PBreakInside:        String("auto"),
	PMarginBreak:        String("auto"),
	POrphans:            Int(2),
	PWidows:             Int(2),

	// Generated Content 3 (WD): https://www.w3.org/TR/css-content-3/
	PBookmarkLabel:   ContentProperties{{Type: "content", Content: String("text")}},
	PBookmarkLevel:   IntString{String: "none"},
	PBookmarkState:   String("open"),
	PContent:         SContent{String: "normal"},
	PFootnoteDisplay: String("block"),
	PFootnotePolicy:  String("auto"),
	PQuotes:          Quotes{Open: []string{"“", "‘"}, Close: []string{"”", "’"}}, // chosen by the user agent
	PStringSet:       StringSet{String: "none"},

	// Images 3/4 (CR/WD): https://www.w3.org/TR/css-images-4/
	PImageResolution:  FToV(1), // dppx
	PImageRendering:   String("auto"),
	PImageOrientation: SBoolFloat{String: "from-image"},
	PObjectFit:        String("fill"),
	PObjectPosition: Center{OriginX: "left", OriginY: "top", Pos: Point{
		Dimension{Value: 50, Unit: Perc}, Dimension{Value: 50, Unit: Perc},
	}},

	// Paged Media 3 (WD): https://www.w3.org/TR/css-page-3/
	PSize:        A4.ToPixels(),
	PPage:        Page{String: "auto"},
	PBleedLeft:   SToV("auto"),
	PBleedRight:  SToV("auto"),
	PBleedTop:    SToV("auto"),
	PBleedBottom: SToV("auto"),
	PMarks:       Marks{}, // computed value for 'none'

	// Text 3/4 (WD/WD): https://www.w3.org/TR/css-text-4/
	PHyphenateCharacter:  String("-"), // computed value chosen by the user agent
	PHyphenateLimitChars: Ints3{5, 2, 2},
	PHyphenateLimitZone:  zeroPixelsValue,
	PHyphens:             String("manual"),
	PLetterSpacing:       SToV("normal"),
	PTabSize:             Value{Dimension: Dimension{Value: 8}},
	PTextAlignAll:        String("start"),
	PTextAlignLast:       String("auto"),
	PTextIndent:          zeroPixelsValue,
	PTextTransform:       String("none"),
	PWhiteSpace:          String("normal"),
	PWordBreak:           String("normal"),
	PWordSpacing:         Value{}, // computed value for "normal"

	// Transforms 1 (CR): https://www.w3.org/TR/css-transforms-1/
	PTransformOrigin: Point{{Value: 50, Unit: Perc}, {Value: 50, Unit: Perc}},
	PTransform:       Transforms{}, // computed value for "none"

	// User Interface 3/4 (REC/WD): https://www.w3.org/TR/css-ui-4/
	PAppearance:   String("none"),
	POutlineColor: CurrentColor, // invert is not supported
	POutlineStyle: String("none"),
	POutlineWidth: Value{Dimension: Dimension{Value: 3}}, // computed value for "medium"

	// Sizing 3 (WD): https://www.w3.org/TR/css-sizing-3/
	PBoxSizing: String("content-box"),
	PHeight:    SToV("auto"),
	PMaxHeight: Value{Dimension: Dimension{Value: Inf, Unit: Px}}, // parsed value for "none}"
	PMaxWidth:  Value{Dimension: Dimension{Value: Inf, Unit: Px}},
	PMinHeight: SToV("auto"),
	PMinWidth:  SToV("auto"),
	PWidth:     SToV("auto"),

	// Flexible Box Layout Module 1 (CR): https://www.w3.org/TR/css-flexbox-1/
	PAlignContent:   String("stretch"),
	PAlignItems:     String("stretch"),
	PAlignSelf:      String("auto"),
	PFlexBasis:      SToV("auto"),
	PFlexDirection:  String("row"),
	PFlexGrow:       Float(0),
	PFlexShrink:     Float(1),
	PFlexWrap:       String("nowrap"),
	PJustifyContent: String("flex-start"),
	POrder:          Int(0),

	// Text Decoration Module 3 (CR): https://www.w3.org/TR/css-text-decor-3/
	PTextDecorationLine:  Decorations{},
	PTextDecorationColor: CurrentColor,
	PTextDecorationStyle: String("solid"),

	// Overflow Module 3 (WD): https://www.w3.org/TR/css-overflow-3/
	PBlockEllipsis: TaggedString{Tag: None},
	PContinue:      String("auto"),
	PMaxLines:      TaggedInt{Tag: None},
	POverflow:      String("visible"),
	POverflowWrap:  String("normal"),
	PTextOverflow:  String("clip"),

	// Lists Module 3 (WD): https://drafts.csswg.org/css-lists-3/
	// Means "none", but allow `display: list-item` to increment the
	// list-item counter. If we ever have a way for authors to query
	// computed values (JavaScript?), this value should serialize to "none".
	PCounterIncrement:  SIntStrings{String: "auto"},
	PCounterReset:      SIntStrings{Values: IntStrings{}}, // parsed value for "none"
	PCounterSet:        SIntStrings{Values: IntStrings{}}, // parsed value for "none"
	PListStyleImage:    Image(NoneImage{}),
	PListStylePosition: String("outside"),
	PListStyleType:     CounterStyleID{Name: "disc"},

	// Proprietary
	PAnchor: String(""),    // computed value of "none"
	PLink:   NamedString{}, // computed value of "none"
	PLang:   NamedString{}, // computed value of "none"
}

func (pr KnownProp) IsTextDecoration() bool {
	return PTextDecorationLine <= pr && pr <= PTextDecorationStyle
}
