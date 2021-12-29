// Fetch and decode images in range various formats.
package images

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"image"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/svg"
	"github.com/benoitkugler/webrender/utils"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type Color = parser.RGBA

// Image is the common interface for supported image formats,
// such as gradients, SVG, or JPEG, PNG, etc...
type Image interface {
	backend.Image

	isImage()
}

var (
	_ Image = rasterImage{}
	_ Image = &SVGImage{}
	_ Image = LinearGradient{}
	_ Image = RadialGradient{}
)

// An error occured when loading an image.
// The image data is probably corrupted or in an invalid format.
func imageLoadingError(err error) error {
	return fmt.Errorf("error loading image : %s", err)
}

// Cache stores the result of fetching an image.
type Cache map[string]Image

func NewCache() Cache { return make(Cache) }

// Gets an image from an image URI.
// In case of an error, a log is printed and nil is returned
func GetImageFromUri(cache Cache, fetcher utils.UrlFetcher, optimizeSize bool, url, forcedMimeType string) Image {
	res, in := cache[url]
	if in {
		return res
	}

	img, err := getImageFromUri(fetcher, optimizeSize, url, forcedMimeType)

	cache[url] = img

	if err != nil {
		log.Println(err)
	}

	return img
}

func getImageFromUri(fetcher utils.UrlFetcher, optimizeSize bool, url, forcedMimeType string) (Image, error) {
	var (
		img     Image
		err     error
		content utils.RemoteRessource
	)

	content, err = fetcher(url)
	if err != nil {
		err = fmt.Errorf(`Failed to load image at "%s" (%s)`, url, err)
		return nil, err
	}

	mimeType := forcedMimeType
	if mimeType == "" {
		mimeType = content.MimeType
	}

	var errSvg error
	// Try to rely on given mimetype for SVG
	if mimeType == "image/svg+xml" {
		var svgIm *SVGImage
		svgIm, errSvg = NewSVGImage(content.Content, url, fetcher)
		if svgIm != nil {
			img = svgIm
		}
	}

	// Look for raster images, or for failing SVG
	if img == nil {
		content.Content.Seek(0, io.SeekStart)
		imageConfig, imageFormat, errRaster := image.DecodeConfig(content.Content)
		if errRaster != nil {
			if errSvg != nil { // Tried SVGImage then raster for a SVG, abort
				err = fmt.Errorf(`Failed to load image at "%s" (%s)`, url, errSvg)
				return nil, err
			}

			// Last chance, try SVG in case mime type is incorrect
			content.Content.Seek(0, io.SeekStart)
			img, errSvg = NewSVGImage(content.Content, url, fetcher)
			if errSvg != nil {
				err = fmt.Errorf(`Failed to load image at "%s" (%s)`, url, errRaster)
				return nil, err
			}
		} else {
			content.Content.Seek(0, io.SeekStart)
			img = newRasterImage(imageConfig, content.Content, "image/"+imageFormat, Hash(url), optimizeSize)
		}
	}

	return img, err
}

// Hash creates an ID from a string.
func Hash(s string) int {
	h := fnv.New32()
	h.Write([]byte(s))
	return int(h.Sum32())
}

type rasterImage struct {
	image backend.RasterImage

	intrinsicRatio  pr.Float
	intrinsicWidth  pr.Float
	intrinsicHeight pr.Float
	optimizeSize    bool
}

func newRasterImage(imageConfig image.Config, content io.ReadCloser, mimeType string, id int, optimizeSize bool) rasterImage {
	self := rasterImage{}
	self.optimizeSize = optimizeSize
	self.image.Content = content
	self.image.MimeType = mimeType
	self.image.ID = id
	self.intrinsicWidth = pr.Float(imageConfig.Width)
	self.intrinsicHeight = pr.Float(imageConfig.Height)
	self.intrinsicRatio = pr.Inf
	if self.intrinsicHeight != 0 {
		self.intrinsicRatio = self.intrinsicWidth / self.intrinsicHeight
	}
	return self
}

func (r rasterImage) isImage() {}

func (r rasterImage) GetIntrinsicSize(imageResolution, _ pr.Float) (width, height, ratio pr.MaybeFloat) {
	// Raster images are affected by the "image-resolution" property.
	return r.intrinsicWidth / imageResolution, r.intrinsicHeight / imageResolution, r.intrinsicRatio
}

func (r rasterImage) Draw(context backend.GraphicTarget, concreteWidth, concreteHeight pr.Fl, imageRendering string) {
	hasSize := concreteWidth > 0 && concreteHeight > 0 && r.intrinsicWidth > 0 && r.intrinsicHeight > 0
	if !hasSize {
		return
	}

	r.image.Rendering = string(imageRendering)
	context.DrawRasterImage(r.image, concreteWidth, concreteHeight)
}

type SVGImage struct {
	icon       *svg.SVGImage
	urlFetcher utils.UrlFetcher
	baseUrl    string
	svgData    []byte
	width      float64
	height     float64
}

func (SVGImage) isImage() {}

func NewSVGImage(svgData io.Reader, baseUrl string, urlFetcher utils.UrlFetcher) (*SVGImage, error) {
	self := new(SVGImage)
	// Don’t pass data URIs to CairoSVG.
	// They are useless for relative URIs anyway.
	if !strings.HasPrefix(strings.ToLower(baseUrl), "data:") {
		self.baseUrl = baseUrl
	}
	content, err := ioutil.ReadAll(svgData)
	if err != nil {
		return nil, imageLoadingError(err)
	}
	self.svgData = content
	self.urlFetcher = urlFetcher

	// FIXME: imageLoader
	self.icon, err = svg.Parse(bytes.NewReader(self.svgData), self.baseUrl, nil)
	if err != nil {
		return nil, imageLoadingError(err)
	}
	return self, nil
}

func (s *SVGImage) GetIntrinsicSize(_, fontSize pr.Float) (pr.MaybeFloat, pr.MaybeFloat, pr.MaybeFloat) {
	width, height := s.icon.DisplayedSize()

	var intrinsicWidth, intrinsicHeight, ratio pr.MaybeFloat
	if width.U != svg.Perc {
		intrinsicWidth = pr.Float(width.Resolve(pr.Fl(fontSize), 0))
	}
	if height.U != svg.Perc {
		intrinsicHeight = pr.Float(height.Resolve(pr.Fl(fontSize), 0))
	}

	if intrinsicWidth == nil || intrinsicHeight == nil {
		viewbox := s.icon.ViewBox
		if viewbox != nil && viewbox.Width != 0 && viewbox.Height != 0 {
			ratio = pr.Float(viewbox.Width / viewbox.Height)
			if pr.Is(intrinsicWidth) {
				intrinsicHeight = intrinsicWidth.V() / ratio.V()
			} else if pr.Is(intrinsicHeight) {
				intrinsicWidth = intrinsicHeight.V() * ratio.V()
			}
		}
	} else if pr.Is(intrinsicWidth) && pr.Is(intrinsicHeight) {
		ratio = intrinsicWidth.V() / intrinsicHeight.V()
	}

	return intrinsicWidth, intrinsicHeight, ratio
}

func (SVGImage) Draw(context backend.GraphicTarget, concreteWidth, concreteHeight pr.Fl, imageRendering string) {
	log.Println("SVG rendering not implemented yet")
	// FIXME:
	//         try {
	//             svg = ScaledSVGSurface(
	//                 cairosvg.parser.Tree(
	//                     bytestring=self.svgData, url=self.baseUrl,
	//                     urlFetcher=self.CairosvgUrlFetcher),
	//                 output=nil, dpi=96, outputWidth=concreteWidth,
	//                 outputHeight=concreteHeight)
	//             if svg.width && svg.height {
	//                 context.scale(
	//                     concreteWidth / svg.width, concreteHeight / svg.height)
	//                 context.setSourceSurface(svg.cairo)
	//                 context.paint()
	//             }
	//         } except Exception as e {
	//             LOGGER.error(
	//                 "Failed to draw an SVG image at %s : %s", self.baseUrl, e)
	//         }
}
