package svg

import (
	"strings"

	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/css/selector"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Apply CSS to SVG documents.

// http://www.w3.org/TR/SVG/styling.html#StyleElement
// n has tag style
func handleStyleElement(n *utils.HTMLNode) []byte {
	if n.DataAtom != atom.Style {
		return nil
	}
	for _, v := range n.Attr {
		if v.Key == "type" && v.Val != "text/css" {
			return nil
		}
	}

	// extract the css
	return n.GetChildrenText()
}

func fetchURL(url, baseURL string) ([]byte, string, error) {
	joinedUrl, err := utils.SafeUrljoin(baseURL, url, true)
	if err != nil {
		return nil, "", err
	}
	cssUrl, err := parseURL(joinedUrl)
	if err != nil {
		return nil, "", err
	}
	resolvedURL := cssUrl.String()
	content, err := utils.FetchSource(utils.InputUrl(resolvedURL), baseURL, utils.DefaultUrlFetcher, false)
	if err != nil {
		return nil, "", err
	}
	return content.Content, resolvedURL, nil
}

// Find rules among stylesheet rules and imports.
func findStylesheetsRules(rules []parser.Token, baseUrl string) (out []parser.QualifiedRule) {
	for _, rule := range rules {
		switch rule := rule.(type) {
		case parser.AtRule:
			if rule.AtKeyword.Lower() == "import" && rule.Content == nil {
				urlToken := parser.ParseOneComponentValue(*rule.Prelude)
				var url string
				switch urlToken := urlToken.(type) {
				case parser.StringToken:
					url = urlToken.Value
				case parser.URLToken:
					url = urlToken.Value
				default:
					continue
				}
				cssContent, resolvedURL, err := fetchURL(url, baseUrl)
				if err != nil {
					logger.WarningLogger.Printf("failed to load stylesheet: %s", err)
					continue
				}

				stylesheet := parser.ParseStylesheetBytes(cssContent, true, true)
				out = append(out, findStylesheetsRules(stylesheet, resolvedURL)...)
			}
			// if rule.AtKeyword.Lower() == "media":
		case parser.QualifiedRule:
			out = append(out, rule)
			// elif rule.type == "error":
		}
	}
	return out
}

type declaration struct {
	property string
	value    string
}

// Parse declarations in a given rule content.
func parseDeclarations(input []parser.Token) (normalDeclarations, importantDeclarations []declaration) {
	for _, decl := range parser.ParseDeclarationList(input, false, false) {
		if decl, ok := decl.(parser.Declaration); ok {
			if strings.HasPrefix(string(decl.Name), "-") {
				continue
			}
			if decl.Important {
				importantDeclarations = append(importantDeclarations, declaration{decl.Name.Lower(), parser.Serialize(decl.Value)})
			} else {
				normalDeclarations = append(normalDeclarations, declaration{decl.Name.Lower(), parser.Serialize(decl.Value)})
			}
		}
	}
	return normalDeclarations, importantDeclarations
}

type match struct {
	selector     selector.SelectorGroup
	declarations []declaration
}

type matcher []match

// Find stylesheets and return rule matchers.
func parseStylesheets(stylesheets [][]byte, url string) (matcher, matcher) {
	var normalMatcher, importantMatcher matcher
	// Parse rules and fill matchers
	for _, css := range stylesheets {
		stylesheet := parser.ParseStylesheetBytes(css, true, true)
		for _, rule := range findStylesheetsRules(stylesheet, url) {
			normalDeclarations, importantDeclarations := parseDeclarations(*rule.Content)
			prelude := parser.Serialize(*rule.Prelude)
			selector, err := selector.ParseGroup(prelude)
			if err != nil {
				logger.WarningLogger.Printf("Invalid or unsupported selector '%s', %s \n", prelude, err)
				continue
			}
			if len(normalDeclarations) != 0 {
				normalMatcher = append(normalMatcher, match{selector: selector, declarations: normalDeclarations})
			}
			if len(importantDeclarations) != 0 {
				importantMatcher = append(importantMatcher, match{selector: selector, declarations: importantDeclarations})
			}
		}
	}
	return normalMatcher, importantMatcher
}

// returns (property, value) pairs
func (m matcher) match(element *html.Node) (out []declaration) {
	for _, mat := range m {
		for _, sel := range mat.selector {
			if sel.Match(element) {
				out = append(out, mat.declarations...)
			}
		}
	}
	return
}

// replace `d` with the (potential) expanded properties
func expandProperty(d declaration) []declaration {
	if d.property != "font" {
		return []declaration{d}
	}

	tokens := validation.RemoveWhitespace(parser.Tokenize([]byte(d.value), true))
	expanded, err := validation.ExpandFont(tokens)
	if err != nil {
		logger.WarningLogger.Printf("ignoring %s property: %s", d.property, err)
		return nil
	}

	out := make([]declaration, len(expanded))
	for i, p := range expanded {
		out[i] = declaration{
			property: p.Name,
			value:    parser.Serialize(p.Tokens),
		}
	}
	return out
}

func (attrs nodeAttributes) applyStyle(baseURL string, node *html.Node, normal, important matcher) {
	var normalAttr, importantAttr []declaration
	if styleAttr := attrs["style"]; styleAttr != "" {
		normalAttr, importantAttr = parseDeclarations(parser.Tokenize([]byte(styleAttr), false))
	}
	delete(attrs, "style") // not useful anymore

	var allProps []declaration
	allProps = append(allProps, normal.match(node)...)
	allProps = append(allProps, normalAttr...)
	allProps = append(allProps, important.match(node)...)
	allProps = append(allProps, importantAttr...)
	for _, d := range allProps {
		expanded := expandProperty(d)
		for _, exp := range expanded {
			attrs[exp.property] = strings.TrimSpace(exp.value)
		}
	}
}
