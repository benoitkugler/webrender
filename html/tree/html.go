package tree

import (
	"bytes"
	"fmt"

	"github.com/benoitkugler/webrender/logger"

	"github.com/benoitkugler/webrender/utils"
	"golang.org/x/net/html"
)

// Represents an HTML document parsed by net/html.
type HTML struct {
	Root       *utils.HTMLNode
	mediaType  string
	UrlFetcher utils.UrlFetcher
	BaseUrl    string

	UAStyleSheet   CSS
	FormStyleSheet CSS
	PHStyleSheet   CSS
}

// `baseUrl` is the base used to resolve relative URLs
// (e.g. in “<img src="../foo.png">“). If not provided, is is infered from
// the input filename or the URL
//
// `urlFetcher` is a function called to fetch external resources such as stylesheets and images.
// and defaults to utils.DefaultUrlFetcher
//
// `mediaType` is the media type to use for “@media“, and defaults to "print".
func NewHTML(htmlContent utils.ContentInput, baseUrl string, urlFetcher utils.UrlFetcher, mediaType string) (*HTML, error) {
	logger.ProgressLogger.Println("Step 1 - Fetching and parsing HTML")
	if urlFetcher == nil {
		urlFetcher = utils.DefaultUrlFetcher
	}
	if mediaType == "" {
		mediaType = "print"
	}
	result, err := utils.FetchSource(htmlContent, baseUrl, urlFetcher, false)
	if err != nil {
		return nil, fmt.Errorf("can't fetch html input : %s", err)
	}

	root, err := html.ParseWithOptions(bytes.NewReader(result.Content), html.ParseOptionEnableScripting(false))
	if err != nil || root.FirstChild == nil {
		return nil, fmt.Errorf("invalid html input : %s", err)
	}

	var out HTML
	// html.Parse wraps the <html> tag
	out.Root = (*utils.HTMLNode)(root.FirstChild)
	if out.Root.Type == html.DoctypeNode {
		out.Root = (*utils.HTMLNode)(out.Root.NextSibling)
	}
	out.Root.Parent = nil
	out.BaseUrl = utils.FindBaseUrl(root, result.BaseUrl)
	out.UrlFetcher = urlFetcher
	out.mediaType = mediaType
	out.UAStyleSheet = Html5UAStylesheet
	out.PHStyleSheet = Html5PHStylesheet
	return &out, nil
}

func newHtml(htmlContent utils.ContentInput) (*HTML, error) {
	return NewHTML(htmlContent, "", nil, "")
}

func (h HTML) GetMetadata() utils.DocumentMetadata {
	return utils.GetHtmlMetadata(h.Root, h.BaseUrl)
}
