package layout

import (
	"fmt"
	"math"
	"strings"

	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/images"
	"github.com/benoitkugler/webrender/utils"
)

func boxRectangle(box bo.BoxFields, whichRectangle string) [4]pr.Float {
	switch whichRectangle {
	case "border-box":
		return [4]pr.Float{
			box.BorderBoxX(),
			box.BorderBoxY(),
			box.BorderWidth(),
			box.BorderHeight(),
		}
	case "padding-box":
		return [4]pr.Float{
			box.PaddingBoxX(),
			box.PaddingBoxY(),
			box.PaddingWidth(),
			box.PaddingHeight(),
		}
	case "content-box":
		return [4]pr.Float{
			box.ContentBoxX(),
			box.ContentBoxY(),
			box.Width.V(),
			box.Height.V(),
		}
	default:
		panic(fmt.Sprintf("unexpected whichRectangle : %s", whichRectangle))
	}
}

// emulate Python itertools.cycle
// i is the current iteration index, N the length of the target slice.
func cycle(i, N int) int { return i % N }

func resolveImage(image pr.Image, orientation pr.SBoolFloat, getImageFromUri bo.ImageFetcher) images.Image {
	switch img := image.(type) {
	case nil, pr.NoneImage:
		return nil
	case pr.UrlImage:
		return getImageFromUri(string(img), "", orientation)
	case pr.RadialGradient:
		return images.NewRadialGradient(img)
	case pr.LinearGradient:
		return images.NewLinearGradient(img)
	default:
		panic(fmt.Sprintf("unexpected type for image: %T %v", image, image))
	}
}

// Fetch and position background images.
func layoutBoxBackgrounds(page *bo.PageBox, box_ Box, getImageFromUri bo.ImageFetcher, layoutChildren bool, style pr.ElementStyle) {
	// Resolve percentages in border-radius properties
	box := box_.Box()
	resolveRadiiPercentages(box)

	if layoutChildren {
		for _, child := range box_.AllChildren() {
			layoutBoxBackgrounds(page, child, getImageFromUri, true, nil)
		}
	}

	if style == nil {
		style = box.Style
	}

	// This is for the border image, not the background, but this is a
	// convenient place to get the image.
	box.BorderImage = resolveImage(style.GetBorderImageSource(), pr.SBoolFloat{}, getImageFromUri)

	var (
		color     parser.RGBA // transparent
		images_   []images.Image
		anyImages = false
	)
	if style.GetVisibility() != "hidden" {
		orientation := style.GetImageOrientation()
		bs := style.GetBackgroundImage()
		images_ = make([]images.Image, len(bs))
		for i, v := range bs {
			images_[i] = resolveImage(v, orientation, getImageFromUri)
			if images_[i] != nil {
				anyImages = true
			}
		}
		color = tree.ResolveColor(style, pr.PBackgroundColor).RGBA
	}

	if color.A == 0 && !anyImages {
		if page != box_ { // Pages need a background for bleed box
			box.Background = nil
			return
		}
	}

	sizes := style.GetBackgroundSize()
	sizesN := len(sizes)
	clips := style.GetBackgroundClip()
	clipsN := len(clips)
	repeats := style.GetBackgroundRepeat()
	repeatsN := len(repeats)
	origins := style.GetBackgroundOrigin()
	originsN := len(origins)
	positions := style.GetBackgroundPosition()
	positionsN := len(positions)
	attachments := style.GetBackgroundAttachment()
	attachmentsN := len(attachments)

	ir := style.GetImageResolution()
	layers := make([]bo.BackgroundLayer, len(images_))
	for i, img := range images_ {
		layers[i] = layoutBackgroundLayer(box_, page, ir, img,
			sizes[cycle(i, sizesN)],
			clips[cycle(i, clipsN)],
			repeats[cycle(i, repeatsN)],
			origins[cycle(i, originsN)],
			positions[cycle(i, positionsN)],
			attachments[cycle(i, attachmentsN)],
		)
	}

	if traceMode {
		traceLogger.Dump(fmt.Sprintf("background: %v", layers))
	}

	box.Background = &bo.Background{Color: color, ImageRendering: style.GetImageRendering(), Layers: layers}
}

func layoutBackgroundLayer(box_ Box, page *bo.PageBox, resolution pr.DimOrS, image images.Image, size pr.Size, clip string, repeat [2]string,
	origin string, position pr.Center, attachment string,
) bo.BackgroundLayer {
	var (
		clippedBoxes []bo.RoundedBox
		paintingArea pr.Rectangle
	)
	box := box_.Box()
	if box_ == page {
		// [The page’s] background painting area is the bleed area […]
		// regardless of background-clip.
		// https://drafts.csswg.org/css-page-3/#painting
		paintingArea = page.BleedArea()
	} else if bo.TableRowGroupT.IsInstance(box_) {
		clippedBoxes = nil
		var totalHeight pr.Float
		for _, row_ := range box.Children {
			row := row_.Box()
			if len(row.Children) != 0 {
				var max pr.Float
				for _, cell := range row.Children {
					clippedBoxes = append(clippedBoxes, cell.Box().RoundedBorderBox())
					if v := cell.Box().BorderHeight(); v > max {
						max = v
					}
				}
				totalHeight = pr.Max(totalHeight, max)
			}
		}
		paintingArea = [4]pr.Float{
			box.BorderBoxX(), box.BorderBoxY(),
			box.BorderWidth(), totalHeight,
		}
	} else if bo.TableRowT.IsInstance(box_) {
		if len(box.Children) != 0 {
			clippedBoxes = nil
			var max pr.Float
			for _, cell := range box.Children {
				clippedBoxes = append(clippedBoxes, cell.Box().RoundedBorderBox())
				if v := cell.Box().BorderHeight(); v > max {
					max = v
				}
			}
			height := max
			paintingArea = [4]pr.Float{
				box.BorderBoxX(), box.BorderBoxY(),
				box.BorderWidth(), height,
			}
		}
	} else if bo.TableColumnGroupT.IsInstance(box_) || bo.TableColumnT.IsInstance(box_) {
		cells := box.GetCells()
		if len(cells) != 0 {
			clippedBoxes = nil
			maxX, minX := -pr.Inf, pr.Inf
			for _, cell := range cells {
				clippedBoxes = append(clippedBoxes, cell.Box().RoundedBorderBox())
				if v := cell.Box().BorderBoxX() + cell.Box().BorderWidth(); v > maxX {
					maxX = v
				}
				if v := cell.Box().BorderBoxX(); v < minX {
					minX = v
				}
			}
			paintingArea = [4]pr.Float{
				minX, box.BorderBoxY(),
				maxX - minX, box.BorderHeight(),
			}
		}
	} else {
		paintingArea = boxRectangle(*box, clip)
		switch clip {
		case "border-box":
			clippedBoxes = []bo.RoundedBox{box.RoundedBorderBox()}
		case "padding-box":
			clippedBoxes = []bo.RoundedBox{box.RoundedPaddingBox()}
		case "content-box":
			clippedBoxes = []bo.RoundedBox{box.RoundedContentBox()}
		default:
			// see validation
			panic(fmt.Sprintf("unexpected clip : %s", clip))
		}
	}

	var intrinsicWidth, intrinsicHeight, ratio pr.MaybeFloat
	if image != nil {
		intrinsicWidth, intrinsicHeight, ratio = image.GetIntrinsicSize(resolution.Value, box.Style.GetFontSize().Value)
	}
	if image == nil || (intrinsicWidth == pr.Float(0) || intrinsicHeight == pr.Float(0)) {
		return bo.BackgroundLayer{
			Image: nil, Unbounded: box_ == page, PaintingArea: paintingArea,
			Size: [2]pr.Float{}, Position: bo.Position{String: "unused"}, Repeat: bo.Repeat{String: "unused"},
			ClippedBoxes: clippedBoxes,
		}
	}

	var positioningArea [4]pr.Float
	if attachment == "fixed" {
		// Initial containing block
		if bo.PageT.IsInstance(box_) {
			// […] if background-attachment is fixed then the image is
			// positioned relative to the page box including its margins […].
			// https://drafts.csswg.org/css-page/#painting
			positioningArea = [4]pr.Float{0, 0, box.MarginWidth(), box.MarginHeight()}
		} else {
			positioningArea = boxRectangle(page.BoxFields, "content-box")
		}
	} else {
		positioningArea = boxRectangle(*box, origin)
	}

	_, _, positioningWidth, positioningHeight := positioningArea[0], positioningArea[1], positioningArea[2], positioningArea[3]
	var imageWidth, imageHeight pr.Float
	if size.String == "cover" {
		imageWidth, imageHeight = coverConstraintImageSizing(positioningWidth, positioningHeight, ratio)
	} else if size.String == "contain" {
		imageWidth, imageHeight = containConstraintImageSizing(positioningWidth, positioningHeight, ratio)
	} else {
		sizeWidth, sizeHeight := size.Width, size.Height
		iwidth, iheight, iratio := image.GetIntrinsicSize(resolution.Value, box.Style.GetFontSize().Value)
		imageWidth, imageHeight = DefaultImageSizing(iwidth, iheight, iratio,
			pr.ResolvePercentage(sizeWidth, positioningWidth), pr.ResolvePercentage(sizeHeight, positioningHeight), positioningWidth, positioningHeight)
	}

	originX, positionX_, originY, positionY_ := position.OriginX, position.Pos[0], position.OriginY, position.Pos[1]
	refX := positioningWidth - imageWidth
	refY := positioningHeight - imageHeight
	positionX := pr.ResolvePercentage(positionX_.ToValue(), refX)
	positionY := pr.ResolvePercentage(positionY_.ToValue(), refY)
	if originX == "right" {
		positionX = refX - positionX.V()
	}
	if originY == "bottom" {
		positionY = refY - positionY.V()
	}

	repeatX, repeatY := repeat[0], repeat[1]

	if repeatX == "round" {
		nRepeats := utils.MaxInt(1, int(math.Round(float64(positioningWidth/imageWidth))))
		newWidth := positioningWidth / pr.Float(nRepeats)
		positionX = pr.Float(0) // Ignore background-position for this dimension
		if repeatY != "round" && size.Height.S == "auto" {
			imageHeight *= newWidth / imageWidth
		}
		imageWidth = newWidth
	}
	if repeatY == "round" {
		nRepeats := utils.MaxInt(1, int(math.Round(float64(positioningHeight/imageHeight))))
		newHeight := positioningHeight / pr.Float(nRepeats)
		positionY = pr.Float(0) // Ignore background-position for this dimension
		if repeatX != "round" && size.Width.S == "auto" {
			imageWidth *= newHeight / imageHeight
		}
		imageHeight = newHeight
	}

	return bo.BackgroundLayer{
		Image:           image,
		Size:            [2]pr.Float{imageWidth, imageHeight},
		Position:        bo.Position{Point: bo.MaybePoint{positionX, positionY}},
		Repeat:          bo.Repeat{Reps: repeat},
		Unbounded:       false,
		PaintingArea:    paintingArea,
		PositioningArea: positioningArea,
		ClippedBoxes:    clippedBoxes,
	}
}

// Layout backgrounds on the page box and on its children.
//
// This function takes care of the canvas background, taken from the root
// elememt or a <body> child of the root element.
//
// See https://www.w3.org/TR/CSS21/colors.html#background
func layoutBackgrounds(page *bo.PageBox, getImageFromUri bo.ImageFetcher) {
	layoutBoxBackgrounds(page, page, getImageFromUri, true, nil)

	rootBox_ := page.Children[0]
	rootBox := rootBox_.Box()
	if bo.MarginT.IsInstance(rootBox_) {
		panic("unexpected margin box as first child of page")
	}
	chosenBox_ := rootBox_
	if strings.ToLower(rootBox.ElementTag()) == "html" && rootBox.Background == nil {
		for _, child := range rootBox.Children {
			if strings.ToLower(child.Box().ElementTag()) == "body" {
				chosenBox_ = child
				break
			}
		}
	}
	chosenBox := chosenBox_.Box()
	if chosenBox.Background != nil {
		paintingArea := boxRectangle(page.BoxFields, "border-box")
		originalBackground := page.Background
		layoutBoxBackgrounds(page, page, getImageFromUri, false, chosenBox.Style)
		canvasBg := *page.Background
		for i, l := range canvasBg.Layers {
			l.PaintingArea = paintingArea
			canvasBg.Layers[i] = l
		}
		page.CanvasBackground = &canvasBg
		page.Background = originalBackground
		chosenBox.Background = nil
	} else {
		page.CanvasBackground = nil
	}
}
