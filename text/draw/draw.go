// Package draw use a backend and a layout object to
// draw glyphs on the ouput.
package draw

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/benoitkugler/textlayout/pango"
	"github.com/benoitkugler/webrender/backend"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/utils"
)

type Context struct {
	Output backend.Canvas          // where to draw the text
	Fonts  *text.FontConfiguration // used to find fonts
}

// CreateFirstLine create the text for the first line of `layout` starting at position `(x,y)`.
// It also register the font used.
func (ctx Context) CreateFirstLine(layout *text.TextLayout, style pr.StyleAccessor,
	textOverflow string, blockEllipsis pr.TaggedString, x, y, angle pr.Fl) backend.TextDrawing {
	pl := &layout.Layout
	pl.SetSingleParagraphMode(true)

	var ellipsis string
	if textOverflow == "ellipsis" || blockEllipsis.Tag != pr.None {
		// assert layout.maxWidth is not nil
		maxWidth := layout.MaxWidth.V()
		pl.SetWidth(pango.Unit(text.PangoUnitsFromFloat(pr.Fl(maxWidth))))
		if textOverflow == "ellipsis" {
			pl.SetEllipsize(pango.ELLIPSIZE_END)
		} else {
			ellipsis = blockEllipsis.S
			if blockEllipsis.Tag == pr.Auto {
				ellipsis = "…"
			}
			// Remove last word if hyphenated
			newText := pl.Text
			if hyph := string(style.GetHyphenateCharacter()); strings.HasSuffix(string(newText), hyph) {
				lastWordEnd := text.GetLastWordEnd(newText[:len(newText)-len([]rune(hyph))])
				if lastWordEnd != -1 && lastWordEnd != 0 {
					newText = newText[:lastWordEnd]
				}
			}
			layout.SetText(string(newText) + ellipsis)
		}
	}

	firstLine, secondLine := layout.GetFirstLine()
	if blockEllipsis.Tag != pr.None {
		for secondLine != 0 && secondLine != -1 {
			lastWordEnd := text.GetLastWordEnd(pl.Text[:len(pl.Text)-len([]rune(ellipsis))])
			if lastWordEnd == -1 {
				break
			}
			newText := pl.Text[:lastWordEnd]
			layout.SetText(string(newText) + ellipsis)
			firstLine, secondLine = layout.GetFirstLine()
		}
	}

	var (
		output               backend.TextDrawing
		inkRect, logicalRect pango.Rectangle
		lastFont             *backend.Font
		xAdvance             pr.Fl
	)

	fontSize := pr.Fl(style.GetFontSize().Value)

	output.FontSize = fontSize
	output.X, output.Y = x, y
	output.Angle = angle

	textRunes := pl.Text
	for run := firstLine.Runs; run != nil; run = run.Next {

		// Pango objects
		glyphItem := run.Data
		glyphString := glyphItem.Glyphs
		runStart := glyphItem.Item.Offset

		// Font content
		pangoFont := glyphItem.Item.Analysis.Font
		content := ctx.Fonts.FontContent(pangoFont.FaceID())
		outFont := ctx.Output.AddFont(pangoFont, content)

		if outFont != lastFont { // add a new "run"
			var outRun backend.TextRun
			outRun.Font = pangoFont
			output.Runs = append(output.Runs, outRun)
		} else { // use the last one
		}
		runDst := &output.Runs[len(output.Runs)-1]

		runDst.Glyphs = make([]backend.TextGlyph, len(glyphString.Glyphs))
		for i, glyphInfo := range glyphString.Glyphs {
			outGlyph := &runDst.Glyphs[i]
			width := glyphInfo.Geometry.Width
			glyph := glyphInfo.Glyph

			if glyph == pango.GLYPH_EMPTY {
				outGlyph.Offset = pr.Fl(width) / fontSize
				outGlyph.Glyph = fonts.EmptyGlyph
				continue
			}

			outGlyph.Offset = pr.Fl(glyphInfo.Geometry.XOffset) / fontSize
			outGlyph.Glyph = glyph.GID()

			// Ink bounding box and logical widths in font
			if _, in := outFont.Extents[outGlyph.Glyph]; !in {
				pangoFont.GlyphExtents(glyph, &inkRect, &logicalRect)
				x1, y1, x2, y2 := inkRect.X, -inkRect.Y-inkRect.Height,
					inkRect.X+inkRect.Width, -inkRect.Y
				if int(x1) < outFont.Bbox[0] {
					outFont.Bbox[0] = int(text.PangoUnitsToFloat(x1*1000) / fontSize)
				}
				if int(y1) < outFont.Bbox[1] {
					outFont.Bbox[1] = int(text.PangoUnitsToFloat(y1*1000) / fontSize)
				}
				if int(x2) > outFont.Bbox[2] {
					outFont.Bbox[2] = int(text.PangoUnitsToFloat(x2*1000) / fontSize)
				}
				if int(y2) > outFont.Bbox[3] {
					outFont.Bbox[3] = int(text.PangoUnitsToFloat(y2*1000) / fontSize)
				}
				outFont.Extents[outGlyph.Glyph] = backend.GlyphExtents{
					Width:  int(text.PangoUnitsToFloat(logicalRect.Width*1000) / fontSize),
					Y:      int(text.PangoUnitsToFloat(logicalRect.Y*1000) / fontSize),
					Height: int(text.PangoUnitsToFloat(logicalRect.Height*1000) / fontSize),
				}
			}

			// Kerning, word spacing, letter spacing
			outGlyph.Kerning = int(pr.Fl(outFont.Extents[outGlyph.Glyph].Width) - text.PangoUnitsToFloat(width*1000)/fontSize + outGlyph.Offset)

			// Mapping between glyphs and characters
			startPos := runStart + glyphString.LogClusters[i] // Positions of the glyphs in the UTF-8 string
			endPos := runStart + glyphItem.Item.Length
			if i < len(glyphString.Glyphs)-1 {
				endPos = runStart + glyphString.LogClusters[i+1]
			}
			if _, in := outFont.Cmap[outGlyph.Glyph]; !in {
				outFont.Cmap[outGlyph.Glyph] = textRunes[startPos:endPos]
			}

			// advance
			outGlyph.XAdvance = xAdvance
			xAdvance += pr.Fl(outFont.Extents[outGlyph.Glyph].Width) + outGlyph.Offset
		}
	}

	return output
}

// DrawEmoji loads and draws `glyph` onto `dst`.
// It may be used by backend implementations to render emojis.
func DrawEmoji(font *harfbuzz.Font, glyph fonts.GID, extents backend.GlyphExtents,
	fontSize, x, y, xAdvance utils.Fl, dst backend.Canvas) {
	face := font.Face()
	data := face.GlyphData(glyph, font.XPpem, font.YPpem)

	switch data := data.(type) {
	case fonts.GlyphBitmap:
		if data.Format == fonts.PNG {
			img := backend.RasterImage{
				Content:   bytes.NewReader(data.Data),
				MimeType:  "image/png",
				Rendering: "",
				ID:        utils.Hash(fmt.Sprintf("%p-%d", face, glyph)),
			}

			d := utils.Fl(extents.Width) / 1000
			a := utils.Fl(data.Width) / utils.Fl(data.Height) * d
			f := utils.Fl(-extents.Y-extents.Height)/1000 - fontSize
			f = y + f
			e := xAdvance / 1000
			e = x + e*fontSize

			dst.OnNewStack(func() {
				dst.State().Transform(matrix.New(a, 0, 0, d, e, f))
				dst.DrawRasterImage(img, fontSize, fontSize)
			})
		}
	}
	// TODO: support more formats
}
