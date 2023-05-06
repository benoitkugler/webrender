package tree

import (
	"fmt"

	"github.com/benoitkugler/webrender/css/counters"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/text"

	"github.com/benoitkugler/webrender/css/parser"

	"github.com/benoitkugler/webrender/css/selector"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/utils"
	"golang.org/x/net/html"

	_ "embed"
)

var (
	// Html5UAStylesheet is the user agent style sheet
	Html5UAStylesheet CSS

	// Html5UAFormsStylesheet is the user agent style sheet used when forms are enabled.
	Html5UAFormsStylesheet CSS

	// Html5PHStylesheet is the presentational hints style sheet
	Html5PHStylesheet CSS

	// TestUAStylesheet is a lightweight style sheet
	TestUAStylesheet CSS

	// The counters defined in the user agent style sheet
	UACounterStyle counters.CounterStyle
)

//go:embed tests_ua.css
var testUACSS string

//go:embed html5_ua.css
var html5UACSS string

//go:embed html5_ua_forms.css
var html5UAFormsCSS string

//go:embed html5_ph.css
var html5PHCSS string

func init() {
	var err error
	TestUAStylesheet, err = NewCSSDefault(utils.InputString(testUACSS))
	if err != nil {
		panic(fmt.Sprintf("invalid embedded stylesheet: %s", err))
	}
	UACounterStyle = make(counters.CounterStyle)
	Html5UAStylesheet, err = NewCSS(utils.InputString(html5UACSS), "", nil, false, "", nil, nil, nil, UACounterStyle)
	if err != nil {
		panic(fmt.Sprintf("invalid embedded stylesheet: %s", err))
	}
	Html5UAFormsStylesheet, err = NewCSS(utils.InputString(html5UAFormsCSS), "", nil, false, "", nil, nil, nil, UACounterStyle)
	if err != nil {
		panic(fmt.Sprintf("invalid embedded stylesheet: %s", err))
	}
	Html5PHStylesheet, err = NewCSSDefault(utils.InputString(html5PHCSS))
	if err != nil {
		panic(fmt.Sprintf("invalid embedded stylesheet: %s", err))
	}
}

// CSS represents a parsed CSS stylesheet.
// An instance is created in the same way as `HTML`, except that
// the “tree“ argument is not available. All other arguments are the same.
// An additional argument called “font_config“ must be provided to handle
// “@font-config“ rules. The same “fonts.FontConfiguration“ object must be
// used for different “CSS“ objects applied to the same document.
// “CSS“ objects have no public attribute or method. They are only meant to
// be used in the `HTML.WritePdf`, `HTML.WritePng` and
// `HTML.Render` methods of `HTML` objects.
type CSS struct {
	Matcher   matcher
	baseUrl   string
	pageRules []PageRule
}

// checkMimeType = false
func NewCSS(input utils.ContentInput, baseUrl string,
	urlFetcher utils.UrlFetcher, checkMimeType bool,
	mediaType string, fontConfig *text.FontConfiguration, matcher *matcher,
	pageRules *[]PageRule, counterStyle counters.CounterStyle,
) (CSS, error) {
	logger.ProgressLogger.Printf("Step 2 - Fetching and parsing CSS - %s", input)

	if urlFetcher == nil {
		urlFetcher = utils.DefaultUrlFetcher
	}
	if mediaType == "" {
		mediaType = "print"
	}

	ressource, err := utils.FetchSource(input, baseUrl, urlFetcher, checkMimeType)
	if err != nil {
		return CSS{}, fmt.Errorf("error fetching css input : %s", err)
	}

	stylesheet := parser.ParseStylesheetBytes(ressource.Content, false, false)

	if matcher == nil {
		matcher = newMatcher()
	}
	if pageRules == nil {
		pageRules = &[]PageRule{}
	}
	if counterStyle == nil {
		counterStyle = make(counters.CounterStyle)
	}

	out := CSS{baseUrl: ressource.BaseUrl}
	preprocessStylesheet(mediaType, ressource.BaseUrl, stylesheet, urlFetcher, matcher,
		pageRules, fontConfig, counterStyle, false)
	out.Matcher = *matcher
	out.pageRules = *pageRules
	return out, nil
}

func NewCSSDefault(input utils.ContentInput) (CSS, error) {
	return NewCSS(input, "", nil, false, "", nil, nil, nil, nil)
}

func (c CSS) IsNone() bool {
	return c.baseUrl == "" && c.Matcher == nil && c.pageRules == nil
}

type match struct {
	selector     selector.SelectorGroup
	declarations []validation.ValidatedProperty
}

type matcher []match

func newMatcher() *matcher {
	return &matcher{}
}

type matchResult struct {
	pseudoType  string
	payload     []validation.ValidatedProperty
	specificity selector.Specificity
}

func (m matcher) Match(element *html.Node) (out []matchResult) {
	for _, mat := range m {
		for _, sel := range mat.selector {
			if sel.Match(element) {
				out = append(out, matchResult{specificity: sel.Specificity(), pseudoType: sel.PseudoElement(), payload: mat.declarations})
			}
		}
	}
	return
}

type pageIndex struct {
	Group []parser.Token // TODO: handle groups
	A, B  int
}

func (p pageIndex) IsNone() bool {
	return p.A == 0 && p.B == 0 && p.Group == nil
}

type pageSelector struct {
	Side        string
	Name        string
	Index       pageIndex
	Specificity selector.Specificity
	Blank       bool
	First       bool
}
