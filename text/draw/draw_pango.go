package draw

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textprocessing/pango"
	"github.com/benoitkugler/textprocessing/pango/fcfonts"
	"github.com/benoitkugler/webrender/backend"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/utils"
)

var _ backend.Font = (*pangoFont)(nil)

type pangoFont fcfonts.Font

func (f *pangoFont) Origin() text.FontOrigin {
	font := (*fcfonts.Font)(f)
	return text.FontOrigin(font.FaceID())
}

func (f *pangoFont) Description() backend.FontDescription {
	font := (*fcfonts.Font)(f)
	desc := font.Describe(false)
	fontSize := desc.Size
	metrics := font.GetMetrics("")

	out := backend.FontDescription{
		Style:  text.FontStyle(desc.Style),
		Family: desc.FamilyName,
		Size:   int(fontSize),
		Weight: int(desc.Weight),
	}
	if fontSize != 0 {
		out.Ascent = backend.Fl(metrics.Ascent * 1000 / pango.Unit(fontSize))
		out.Descent = backend.Fl(metrics.Descent * 1000 / pango.Unit(fontSize))
	}

	if face, ok := font.GetHarfbuzzFont().Face().(*truetype.Font); ok {
		out.IsOpentype = true
		out.IsOpentypeOpentype = face.Type == truetype.TypeOpenType
	}

	return out
}

func (ctx Context) createFirstLinePango(layout *text.TextLayoutPango,
	textOverflow string, blockEllipsis pr.TaggedString, scaleX, x, y, angle pr.Fl,
) backend.TextDrawing {
	style := layout.Style
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
			if hyph := style.HyphenateCharacter; strings.HasSuffix(string(newText), hyph) {
				lastWordEnd := text.GetLastWordEnd(ctx.Fonts, newText[:len(newText)-len([]rune(hyph))])
				if lastWordEnd != -1 && lastWordEnd != 0 {
					newText = newText[:lastWordEnd]
				}
			}
			layout.SetText(string(newText) + ellipsis)
		}
	}

	firstLine, index := layout.GetFirstLine()
	if blockEllipsis.Tag != pr.None {
		for index != 0 && index != -1 {
			lastWordEnd := text.GetLastWordEnd(ctx.Fonts, pl.Text[:len(pl.Text)-len([]rune(ellipsis))])
			if lastWordEnd == -1 {
				break
			}
			newText := pl.Text[:lastWordEnd]
			layout.SetText(string(newText) + ellipsis)
			firstLine, index = layout.GetFirstLine()
		}
	}

	var (
		output               backend.TextDrawing
		inkRect, logicalRect pango.Rectangle
		lastFont             *backend.FontChars
		xAdvance             pr.Fl
	)

	fontSize := style.FontSize

	output.FontSize = fontSize
	output.ScaleX = scaleX
	output.X, output.Y = x, y
	output.Angle = angle

	textRunes := pl.Text
	for run := firstLine.Runs; run != nil; run = run.Next {

		// Pango objects
		glyphItem := run.Data
		glyphString := glyphItem.Glyphs
		offset := glyphItem.Item.Offset

		// Font content
		pFont := glyphItem.Item.Analysis.Font.(*fcfonts.Font)
		content := ctx.Fonts.FontContent(text.FontOrigin(pFont.FaceID()))
		outFont := ctx.Output.AddFont((*pangoFont)(pFont), content)

		if outFont != lastFont { // add a new "run"
			var outRun backend.TextRun
			outRun.Font = (*pangoFont)(pFont)
			output.Runs = append(output.Runs, outRun)
		} // else use the last one

		runDst := &output.Runs[len(output.Runs)-1]

		// Positions of the glyphs in the UTF-8 string
		utf8Positions := make([]int, len(glyphString.Glyphs)-1)
		for i := range utf8Positions {
			utf8Positions[i] = offset + glyphString.LogClusters[i+1]
		}
		utf8Positions = append(utf8Positions, offset+glyphItem.Item.Length)

		runDst.Glyphs = make([]backend.TextGlyph, len(glyphString.Glyphs))
		var prevUtf8Position int
		for i, glyphInfo := range glyphString.Glyphs {
			outGlyph := &runDst.Glyphs[i]
			width := glyphInfo.Geometry.Width
			glyph := glyphInfo.Glyph

			if glyph == pango.GLYPH_EMPTY || glyph&pango.GLYPH_UNKNOWN_FLAG != 0 {
				outGlyph.Offset = pr.Fl(width) / fontSize
				outGlyph.Glyph = backend.GID(fonts.EmptyGlyph)
				continue
			}

			outGlyph.Offset = pr.Fl(glyphInfo.Geometry.XOffset) / fontSize
			outGlyph.Rise = pr.Fl(glyphInfo.Geometry.YOffset)
			outGlyph.Glyph = backend.GID(glyph.GID())

			// Ink bounding box and logical widths in font
			if _, in := outFont.Extents[outGlyph.Glyph]; !in {
				pFont.GlyphExtents(glyph, &inkRect, &logicalRect)
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
			utf8Position := utf8Positions[i]
			if _, in := outFont.Cmap[outGlyph.Glyph]; !in {
				outFont.Cmap[outGlyph.Glyph] = textRunes[prevUtf8Position:utf8Position]
			}
			prevUtf8Position = utf8Position

			// advance
			outGlyph.XAdvance = xAdvance
			xAdvance += pr.Fl(outFont.Extents[outGlyph.Glyph].Width) + outGlyph.Offset - pr.Fl(outGlyph.Kerning)
		}
	}

	return output
}

func drawEmojiPango(font_ *pangoFont, glyph backend.GID, extents backend.GlyphExtents,
	fontSize, x, y, xAdvance utils.Fl, dst backend.Canvas,
) {
	font := (*fcfonts.Font)(font_).GetHarfbuzzFont()
	face := font.Face()
	data := face.GlyphData(fonts.GID(glyph), font.XPpem, font.YPpem)

	switch data := data.(type) {
	// TODO: more formats
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
}
