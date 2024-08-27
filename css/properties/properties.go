package properties

import "github.com/benoitkugler/webrender/css/parser"

// This file is used to generate typed accessors
//go:generate go run gen/gen.go

const (
	_ KnownProp = iota

	// DO NOT CHANGE the order, because
	// the following properties are grouped by side,
	// in the [bottom, left, right, top] order,
	// so that, if side in an index (0, 1, 2 or 3),
	// the property is a PBorderBottomColor + side * 5
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

	PBorderImageSource
	PBorderImageSlice
	PBorderImageWidth
	PBorderImageOutset
	PBorderImageRepeat

	// min-XXX is at +2, max-XXX is a + 4
	PWidth
	PHeight
	PMinWidth
	PMinHeight
	PMaxWidth
	PMaxHeight

	PColor
	PDirection
	PDisplay
	PFloat
	PLineHeight

	PPosition
	PTableLayout
	PTop
	PUnicodeBidi
	PVerticalAlign
	PVisibility
	PZIndex

	PBorderBottomLeftRadius
	PBorderBottomRightRadius
	PBorderTopLeftRadius
	PBorderTopRightRadius

	POpacity

	PColumnRuleStyle
	PColumnRuleWidth
	PColumnCount
	PColumnWidth

	PFontFamily
	PFontFeatureSettings
	PFontKerning
	PFontLanguageOverride
	PFontSize
	PFontStretch
	PFontStyle

	// The order for this group matters (see expandFontVariant)
	PFontVariantAlternates
	PFontVariantCaps
	PFontVariantEastAsian
	PFontVariantLigatures
	PFontVariantNumeric
	PFontVariantPosition

	PFontWeight
	PFontVariationSettings

	PHyphenateCharacter
	PHyphenateLimitChars
	PHyphens
	PLetterSpacing
	PTextAlignAll
	PTextAlignLast
	PTextIndent
	PTextTransform
	PWhiteSpace
	PWordBreak
	PWordSpacing
	PTransform

	PContinue
	PMaxLines
	POverflow
	POverflowWrap
	PCounterIncrement
	PCounterReset
	PCounterSet

	PAnchor
	PLink
	PLang

	PBoxDecorationBreak

	PBookmarkLabel
	PBookmarkLevel
	PBookmarkState
	PContent

	PStringSet
	PImageOrientation

	PPage
	PAppearance
	POutlineColor
	POutlineStyle
	POutlineWidth
	PBoxSizing

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

	// text-decoration-XXX
	PTextDecorationLine
	PTextDecorationColor
	PTextDecorationStyle

	PBreakAfter
	PBreakBefore
	PBreakInside

	PGridAutoColumns
	PGridAutoFlow
	PGridAutoRows
	// the order matter
	PGridTemplateColumns
	PGridTemplateRows
	PGridTemplateAreas
	PGridRowStart
	PGridColumnStart
	PGridRowEnd
	PGridColumnEnd

	PAlignContent
	PAlignItems
	PAlignSelf
	PFlexBasis
	PFlexDirection
	PFlexGrow
	PFlexShrink
	PFlexWrap
	PJustifyContent
	PJustifyItems
	PJustifySelf
	POrder
	PColumnGap
	PRowGap

	PBottom
	PCaptionSide
	PClear
	PClip
	PEmptyCells
	PLeft
	PRight

	PListStyleImage
	PListStylePosition
	PListStyleType

	PTextOverflow
	PBlockEllipsis
	PBorderCollapse
	PBorderSpacing

	PTransformOrigin

	PFontVariant

	PTabSize

	PMarginBreak
	POrphans
	PWidows

	PFootnoteDisplay
	PFootnotePolicy
	PQuotes

	PImageResolution
	PImageRendering

	PColumnFill
	PColumnSpan
	PColumnRuleColor

	PSize
	PBleedLeft
	PBleedRight
	PBleedTop
	PBleedBottom
	PMarks

	PObjectFit
	PObjectPosition

	PHyphenateLimitZone

	NbProperties
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

	PBorderImageSource: NoneImage{},
	PBorderImageSlice: Values{
		PercToV(100), PercToV(100), PercToV(100), PercToV(100),
		DimOrS{},
	},
	PBorderImageWidth: Values{FToV(1), FToV(1), FToV(1), FToV(1)},
	PBorderImageOutset: Values{
		FToV(0), FToV(0),
		FToV(0), FToV(0),
	},
	PBorderImageRepeat: Strings{"stretch", "stretch"},

	PBorderBottomLeftRadius:  Point{ZeroPixels, ZeroPixels},
	PBorderBottomRightRadius: Point{ZeroPixels, ZeroPixels},
	PBorderTopLeftRadius:     Point{ZeroPixels, ZeroPixels},
	PBorderTopRightRadius:    Point{ZeroPixels, ZeroPixels},

	// Color 3 (REC): https://www.w3.org/TR/css-color-3/
	POpacity: Float(1),

	// Multi-column Layout (WD): https://www.w3.org/TR/css-multicol-1/
	PColumnWidth:     SToV("auto"),
	PColumnCount:     IntString{String: "auto"},
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
	PBookmarkLevel:   TaggedInt{Tag: None},
	PBookmarkState:   String("open"),
	PContent:         SContent{String: "normal"},
	PFootnoteDisplay: String("block"),
	PFootnotePolicy:  String("auto"),
	PQuotes:          Quotes{Tag: Auto}, // chosen by the user agent
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
	PPage:        Page("auto"),
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
	PTabSize:             DimOrS{Dimension: Dimension{Value: 8}},
	PTextAlignAll:        String("start"),
	PTextAlignLast:       String("auto"),
	PTextIndent:          zeroPixelsValue,
	PTextTransform:       String("none"),
	PWhiteSpace:          String("normal"),
	PWordBreak:           String("normal"),
	PWordSpacing:         DimOrS{}, // computed value for "normal"

	// Transforms 1 (CR): https://www.w3.org/TR/css-transforms-1/
	PTransformOrigin: Point{{Value: 50, Unit: Perc}, {Value: 50, Unit: Perc}},
	PTransform:       Transforms{}, // computed value for "none"

	// User Interface 3/4 (REC/WD): https://www.w3.org/TR/css-ui-4/
	PAppearance:   String("none"),
	POutlineColor: CurrentColor, // invert is not supported
	POutlineStyle: String("none"),
	POutlineWidth: DimOrS{Dimension: Dimension{Value: 3}}, // computed value for "medium"

	// Sizing 3 (WD): https://www.w3.org/TR/css-sizing-3/
	PBoxSizing: String("content-box"),
	PHeight:    SToV("auto"),
	PMaxHeight: DimOrS{Dimension: Dimension{Value: Inf, Unit: Px}}, // parsed value for "none}"
	PMaxWidth:  DimOrS{Dimension: Dimension{Value: Inf, Unit: Px}},
	PMinHeight: SToV("auto"),
	PMinWidth:  SToV("auto"),
	PWidth:     SToV("auto"),

	// Flexible Box Layout Module 1 (CR): https://www.w3.org/TR/css-flexbox-1/
	PFlexBasis:     SToV("auto"),
	PFlexDirection: String("row"),
	PFlexGrow:      Float(0),
	PFlexShrink:    Float(1),
	PFlexWrap:      String("nowrap"),

	// Grid Layout Module Level 2 (CR): https://www.w3.org/TR/css-grid-2/
	PGridAutoFlow:        Strings{"row"},
	PGridAutoColumns:     GridAuto{NewGridDims(SToV("auto"))},
	PGridAutoRows:        GridAuto{NewGridDims(SToV("auto"))},
	PGridTemplateAreas:   GridTemplateAreas{},
	PGridTemplateColumns: GridTemplate{Tag: None},
	PGridTemplateRows:    GridTemplate{Tag: None},
	PGridRowStart:        GridLine{Tag: Auto},
	PGridColumnStart:     GridLine{Tag: Auto},
	PGridRowEnd:          GridLine{Tag: Auto},
	PGridColumnEnd:       GridLine{Tag: Auto},

	// CSS Box Alignment Module Level 3 (WD): https://www.w3.org/TR/css-align-3/
	PAlignContent:   String("normal"),
	PAlignItems:     String("normal"),
	PAlignSelf:      String("auto"),
	PJustifyContent: String("normal"),
	PJustifyItems:   String("normal"),
	PJustifySelf:    String("auto"),
	POrder:          Int(0),
	PColumnGap:      DimOrS{S: "normal"},
	PRowGap:         DimOrS{S: "normal"},

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
	PAnchor: String(""),     // computed value of "none"
	PLink:   NamedString{},  // computed value of "none"
	PLang:   TaggedString{}, // computed value of "none"
}

func (pr KnownProp) IsTextDecoration() bool {
	return PTextDecorationLine <= pr && pr <= PTextDecorationStyle
}

// Shortand is a compact representation of CSS keywords
// used as shortand for several properties
type Shortand uint8

const (
	_ Shortand = iota
	SBorderColor
	SBorderStyle
	SBorderWidth
	SBorderImage
	SMargin
	SPadding
	SBleed
	SBorderRadius
	SPageBreakAfter
	SPageBreakBefore
	SPageBreakInside
	SBackground
	SWordWrap
	SListStyle
	SBorder
	SBorderTop
	SBorderRight
	SBorderBottom
	SBorderLeft
	SColumnRule
	SOutline
	SColumns
	SFontVariant
	SFont
	STextDecoration
	SFlex
	SFlexFlow
	SLineClamp
	STextAlign
	SGridColumn
	SGridRow
	SGridArea
	SGridTemplate
	SGrid
)

// NewShortand return the tag for 's' or 0 if not supported
func NewShortand(s string) Shortand {
	switch s {
	case "border-color":
		return SBorderColor
	case "border-style":
		return SBorderStyle
	case "border-width":
		return SBorderWidth
	case "border-image":
		return SBorderImage
	case "margin":
		return SMargin
	case "padding":
		return SPadding
	case "bleed":
		return SBleed
	case "border-radius":
		return SBorderRadius
	case "page-break-after":
		return SPageBreakAfter
	case "page-break-before":
		return SPageBreakBefore
	case "page-break-inside":
		return SPageBreakInside
	case "background":
		return SBackground
	case "word-wrap":
		return SWordWrap
	case "list-style":
		return SListStyle
	case "border":
		return SBorder
	case "border-top":
		return SBorderTop
	case "border-right":
		return SBorderRight
	case "border-bottom":
		return SBorderBottom
	case "border-left":
		return SBorderLeft
	case "column-rule":
		return SColumnRule
	case "outline":
		return SOutline
	case "columns":
		return SColumns
	case "font-variant":
		return SFontVariant
	case "font":
		return SFont
	case "text-decoration":
		return STextDecoration
	case "flex":
		return SFlex
	case "flex-flow":
		return SFlexFlow
	case "line-clamp":
		return SLineClamp
	case "text-align":
		return STextAlign
	case "grid-column":
		return SGridColumn
	case "grid-row":
		return SGridRow
	case "grid-area":
		return SGridArea
	case "grid-template":
		return SGridTemplate
	case "grid":
		return SGrid
	default:
		return 0
	}
}

// String returns the CSS keyword.
func (sh Shortand) String() string {
	switch sh {
	case SBorderColor:
		return "border-color"
	case SBorderStyle:
		return "border-style"
	case SBorderWidth:
		return "border-width"
	case SBorderImage:
		return "border-image"
	case SMargin:
		return "margin"
	case SPadding:
		return "padding"
	case SBleed:
		return "bleed"
	case SBorderRadius:
		return "border-radius"
	case SPageBreakAfter:
		return "page-break-after"
	case SPageBreakBefore:
		return "page-break-before"
	case SPageBreakInside:
		return "page-break-inside"
	case SBackground:
		return "background"
	case SWordWrap:
		return "word-wrap"
	case SListStyle:
		return "list-style"
	case SBorder:
		return "border"
	case SBorderTop:
		return "border-top"
	case SBorderRight:
		return "border-right"
	case SBorderBottom:
		return "border-bottom"
	case SBorderLeft:
		return "border-left"
	case SColumnRule:
		return "column-rule"
	case SOutline:
		return "outline"
	case SColumns:
		return "columns"
	case SFontVariant:
		return "font-variant"
	case SFont:
		return "font"
	case STextDecoration:
		return "text-decoration"
	case SFlex:
		return "flex"
	case SFlexFlow:
		return "flex-flow"
	case SLineClamp:
		return "line-clamp"
	case STextAlign:
		return "text-align"
	case SGridColumn:
		return "grid-column"
	case SGridRow:
		return "grid-row"
	case SGridArea:
		return "grid-area"
	case SGridTemplate:
		return "grid-template"
	case SGrid:
		return "grid"
	default:
		return ""
	}
}
