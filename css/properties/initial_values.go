package properties

import (
	"github.com/benoitkugler/webrender/css/parser"
)

// InitialValues stores the default values for the CSS properties.
var InitialValues = Properties{
	"bottom":       SToV("auto"),
	"caption_side": String("top"),
	"clear":        String("none"),
	"clip":         Values{},                                // computed value for "auto"
	"color":        Color(parser.ParseColorString("black")), // chosen by the user agent

	// Means "none", but allow `display: list-item` to increment the
	// list-item counter. If we ever have a way for authors to query
	// computed values (JavaScript?), this value should serialize to "none".
	"counter_increment": SIntStrings{String: "auto"},
	"counter_reset":     SIntStrings{Values: IntStrings{}}, // parsed value for "none"
	"counter_set":       SIntStrings{Values: IntStrings{}}, // parsed value for "none"

	"direction":           String("ltr"),
	"display":             Display{"inline", "flow"},
	"empty_cells":         String("show"),
	"float":               String("none"),
	"height":              SToV("auto"),
	"left":                SToV("auto"),
	"right":               SToV("auto"),
	"line_height":         SToV("normal"),
	"list_style_image":    Image(NoneImage{}),
	"list_style_position": String("outside"),
	"list_style_type":     CounterStyleID{Name: "disc"},
	"margin_top":          zeroPixelsValue,
	"margin_right":        zeroPixelsValue,
	"margin_bottom":       zeroPixelsValue,
	"margin_left":         zeroPixelsValue,
	"max_height":          Value{Dimension: Dimension{Value: Inf, Unit: Px}}, // parsed value for "none}"
	"max_width":           Value{Dimension: Dimension{Value: Inf, Unit: Px}},
	"padding_top":         zeroPixelsValue,
	"padding_right":       zeroPixelsValue,
	"padding_bottom":      zeroPixelsValue,
	"padding_left":        zeroPixelsValue,
	"position":            BoolString{String: "static"},
	"table_layout":        String("auto"),
	"top":                 SToV("auto"),
	"unicode_bidi":        String("normal"),
	"vertical_align":      SToV("baseline"),
	"visibility":          String("visible"),
	"width":               SToV("auto"),
	"z_index":             IntString{String: "auto"},

	// Backgrounds and Borders 3 (CR): https://www.w3.org/TR/css3-background/
	"background_attachment": Strings{"scroll"},
	"background_clip":       Strings{"border-box"},
	"background_color":      Color(parser.ParseColorString("transparent")),
	"background_image":      Images{NoneImage{}},
	"background_origin":     Strings{"padding-box"},
	"background_position": Centers{
		Center{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Unit: Percentage}, Dimension{Unit: Percentage}}},
	},
	"background_repeat":   Repeats{{"repeat", "repeat"}},
	"background_size":     Sizes{Size{Width: SToV("auto"), Height: SToV("auto")}},
	"border_bottom_color": CurrentColor,
	"border_left_color":   CurrentColor,
	"border_right_color":  CurrentColor,
	"border_top_color":    CurrentColor,
	"border_bottom_style": String("none"),
	"border_left_style":   String("none"),
	"border_right_style":  String("none"),
	"border_top_style":    String("none"),
	"border_collapse":     String("separate"),
	"border_spacing":      Point{Dimension{Unit: Scalar}, Dimension{Unit: Scalar}},
	"border_bottom_width": FToV(3),
	"border_left_width":   FToV(3),
	"border_top_width":    FToV(3), // computed value for "medium"
	"border_right_width":  FToV(3),

	"border_bottom_left_radius":  Point{ZeroPixels, ZeroPixels},
	"border_bottom_right_radius": Point{ZeroPixels, ZeroPixels},
	"border_top_left_radius":     Point{ZeroPixels, ZeroPixels},
	"border_top_right_radius":    Point{ZeroPixels, ZeroPixels},

	// // Color 3 (REC): https://www.w3.org/TR/css3-color/
	"opacity": Float(1),

	// Multi-column Layout (WD): https://www.w3.org/TR/css-multicol-1/
	"column_width":      SToV("auto"),
	"column_count":      IntString{String: "auto"},
	"column_gap":        Value{Dimension: Dimension{Value: 1, Unit: Em}},
	"column_rule_color": CurrentColor,
	"column_rule_style": String("none"),
	"column_rule_width": SToV("medium"),
	"column_fill":       String("balance"),
	"column_span":       String("none"),

	// Fonts 3 (REC): https://www.w3.org/TR/css-fonts-3/
	"font_family":             Strings{"serif"}, // depends on user agent
	"font_feature_settings":   SIntStrings{String: "normal"},
	"font_kerning":            String("auto"),
	"font_language_override":  String("normal"),
	"font_size":               FToV(16), // actually medium, but we define medium from this
	"font_stretch":            String("normal"),
	"font_style":              String("normal"),
	"font_variant":            String("normal"),
	"font_variant_alternates": String("normal"),
	"font_variant_caps":       String("normal"),
	"font_variant_east_asian": SStrings{String: "normal"},
	"font_variant_ligatures":  SStrings{String: "normal"},
	"font_variant_numeric":    SStrings{String: "normal"},
	"font_variant_position":   String("normal"),
	"font_weight":             IntString{Int: 400},

	// Fragmentation 3/4 (CR/WD): https://www.w3.org/TR/css-break-4/
	"box_decoration_break": String("slice"),
	"break_after":          String("auto"),
	"break_before":         String("auto"),
	"break_inside":         String("auto"),
	"margin_break":         String("auto"),
	"orphans":              Int(2),
	"widows":               Int(2),

	// Generated Content 3 (WD): https://www.w3.org/TR/css-content-3/
	"bookmark_label": ContentProperties{{Type: "content", Content: String("text")}},
	"bookmark_level": IntString{String: "none"},
	"bookmark_state": String("open"),
	"content":        SContent{String: "normal"},
	"quotes":         Quotes{Open: []string{"“", "‘"}, Close: []string{"”", "’"}}, // chosen by the user agent
	"string_set":     StringSet{String: "none"},

	// // Images 3/4 (CR/WD): https://www.w3.org/TR/css4-images/
	"image_resolution": FToV(1), // dppx
	"image_rendering":  String("auto"),
	// https://drafts.csswg.org/css-images-3/
	"object_fit": String("fill"),
	"object_position": Center{OriginX: "left", OriginY: "top", Pos: Point{
		Dimension{Value: 50, Unit: Percentage}, Dimension{Value: 50, Unit: Percentage},
	}},

	// Paged Media 3 (WD): https://www.w3.org/TR/css-page-3/
	"size":         A4.ToPixels(),
	"page":         Page{String: "auto"},
	"bleed_left":   SToV("auto"),
	"bleed_right":  SToV("auto"),
	"bleed_top":    SToV("auto"),
	"bleed_bottom": SToV("auto"),
	"marks":        Marks{}, // computed value for 'none'

	// Text 3/4 (WD/WD): https://www.w3.org/TR/css-text-4/
	"hyphenate_character":   String("-"), // computed value chosen by the user agent
	"hyphenate_limit_chars": Ints3{5, 2, 2},
	"hyphenate_limit_zone":  zeroPixelsValue,
	"hyphens":               String("manual"),
	"letter_spacing":        SToV("normal"),
	"tab_size":              Value{Dimension: Dimension{Value: 8}},
	"text_align_all":        String("start"),
	"text_align_last":       String("auto"),
	"text_indent":           zeroPixelsValue,
	"text_transform":        String("none"),
	"white_space":           String("normal"),
	"word_spacing":          Value{}, // computed value for "normal"

	// Transforms 1 (CR): https://www.w3.org/TR/css-transforms-1/
	"transform_origin": Point{{Value: 50, Unit: Percentage}, {Value: 50, Unit: Percentage}},
	"transform":        Transforms{}, // computed value for "none"

	// User Interface 3 (REC): https://www.w3.org/TR/css-ui-3/
	"box_sizing":    String("content-box"),
	"outline_color": CurrentColor, // invert is not supported
	"outline_style": String("none"),
	"outline_width": Value{Dimension: Dimension{Value: 3}}, // computed value for "medium"
	"overflow_wrap": String("normal"),

	// Flexible Box Layout Module 1 (CR): https://www.w3.org/TR/css-flexbox-1/
	"align_content":   String("stretch"),
	"align_items":     String("stretch"),
	"align_self":      String("auto"),
	"flex_basis":      SToV("auto"),
	"flex_direction":  String("row"),
	"flex_grow":       Float(0),
	"flex_shrink":     Float(1),
	"flex_wrap":       String("nowrap"),
	"justify_content": String("flex-start"),
	"min_height":      zeroPixelsValue,
	"min_width":       zeroPixelsValue,
	"order":           Int(0),

	// Text Decoration Module 3 (CR): https://www.w3.org/TR/css-text-decor-3/
	"text_decoration_line":  Decorations{},
	"text_decoration_color": CurrentColor,
	"text_decoration_style": String("solid"),

	// Overflow Module 3 (WD): https://www.w3.org/TR/css-overflow-3/
	"block_ellipsis": NamedString{Name: "none"},
	"continue":       String("auto"),
	"max_lines":      IntString{String: "none"},
	"overflow":       String("visible"),
	"text_overflow":  String("clip"),

	// Proprietary
	"anchor": String(""),    // computed value of "none"
	"link":   NamedString{}, // computed value of "none"
	"lang":   NamedString{}, // computed value of "none"
}
