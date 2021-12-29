package utils

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/webrender/version"
)

// warn if baseUrl is required but missing.
func UrlJoin(baseUrl, urlS string, allowRelative bool, context string) string {
	out, err := SafeUrljoin(baseUrl, urlS, allowRelative)
	if err != nil {
		log.Println(err, context)
	}
	return out
}

// micmic python urllib.urljoin behavior
func basicUrlJoin(baseUrl string, urls *url.URL) (string, error) {
	parsedBase, err := url.Parse(baseUrl)
	if err != nil {
		return "", fmt.Errorf("Invalid base url : %s", baseUrl)
	}
	if urls.Host != "" { // copy the scheme from base
		urls.Scheme = parsedBase.Scheme
		*parsedBase = *urls
	} else if path.IsAbs(urls.Path) { // join from the root
		parsedBase.Path = urls.Path
	} else { // join from the directory
		if path.Ext(parsedBase.Path) != "" {
			parsedBase.Path = path.Join(path.Dir(parsedBase.Path), urls.Path)
		} else {
			parsedBase.Path = path.Join(parsedBase.Path, urls.Path)
		}
	}
	parsedBase.RawQuery = urls.RawQuery
	parsedBase.RawFragment = urls.RawFragment
	parsedBase.Fragment = urls.Fragment
	return parsedBase.String(), nil
}

// defaut: allowRelative = false
func SafeUrljoin(baseUrl, urls string, allowRelative bool) (string, error) {
	parsed, err := url.Parse(urls)
	if err != nil {
		return "", fmt.Errorf("Invalid url : %s (%s)", urls, err)
	}
	if parsed.IsAbs() {
		return parsed.String(), nil
	} else if baseUrl != "" {
		return basicUrlJoin(baseUrl, parsed)
	} else if allowRelative {
		return parsed.String(), nil
	} else {
		return "", errors.New("Relative URI reference without a base URI: " + urls)
	}
}

// Get the URI corresponding to the ``attrName`` attribute.
// Return "" if:
// * the attribute is empty or missing or,
// * the value is a relative URI but the document has no base URI and
//   ``allowRelative`` is ``False``.
// Otherwise return an URI, absolute if possible.
func (element HTMLNode) GetUrlAttribute(attrName, baseUrl string, allowRelative bool) string {
	value := strings.TrimSpace(element.Get(attrName))
	if value != "" {
		return UrlJoin(baseUrl, value, allowRelative,
			fmt.Sprintf("<%s %s='%s'>", element.Data, attrName, value))
	}
	return ""
}

func Unquote(s string) string {
	unescaped, err := url.PathUnescape(s)
	if err != nil {
		log.Println(err)
		return ""
	}
	return unescaped
}

// Url represent an url which can be either internal or external
type Url struct {
	Url      string
	Internal bool
}

func (u Url) IsNone() bool {
	return u == Url{}
}

// Return ('external', absolute_uri) or
// ('internal', unquoted_fragment_id) or false
func GetLinkAttribute(element *HTMLNode, attrName string, baseUrl string) ([2]string, bool) {
	attrValue := strings.TrimSpace(element.Get(attrName))
	if strings.HasPrefix(attrValue, "#") && len(attrValue) > 1 {
		// Do not require a baseUrl when the value is just a fragment.
		unescaped := Unquote(attrValue[1:])
		return [2]string{"internal", unescaped}, true
	}

	uri := element.GetUrlAttribute(attrName, baseUrl, true)
	if uri == "" {
		return [2]string{}, false
	}
	if baseUrl != "" {
		parsed, err := url.Parse(uri)
		if err != nil {
			log.Println(err)
			return [2]string{}, false
		}
		baseParsed, err := url.Parse(baseUrl)
		if err != nil {
			log.Println(err)
			return [2]string{}, false
		}
		if parsed.Scheme == baseParsed.Scheme && parsed.Host == baseParsed.Host && parsed.Path == baseParsed.Path && parsed.RawQuery == baseParsed.RawQuery {
			// Compare with fragments removed
			return [2]string{"internal", parsed.Fragment}, true
		}
	}
	return [2]string{"external", uri}, true
}

// Return a file URL for the given `file` path.
func PathToURL(file string) (out string, err error) {
	file, err = filepath.Abs(file)
	if err != nil {
		return "", err
	}
	fileinfo, err := os.Lstat(file)
	if err != nil {
		return "", err
	}
	if fileinfo.IsDir() {
		// Make sure directory names have a trailing slash.
		// Otherwise relative URIs are resolved from the parent directory.
		file += string(filepath.Separator)
	}
	file = filepath.ToSlash(file)
	return "file://" + file, nil
}

// Get a ``scheme://path`` URL from ``string``.
//
// If ``string`` looks like an URL, return it unchanged. Otherwise assume a
// filename and convert it to a ``file://`` URL.
func ensureUrl(urlS string) (string, error) {
	parsed, err := url.Parse(urlS)
	if err != nil {
		return "", fmt.Errorf("Invalid url : %s (%s)", urlS, err)
	}
	if parsed.IsAbs() {
		return urlS, nil
	}
	return PathToURL(urlS)
}

type content interface {
	io.Closer
	fonts.Resource
}

type RemoteRessource struct {
	Content content

	// Optionnals values

	// MIME type extracted e.g. from a *Content-Type* header. If not provided, the type is guessed from the
	// 	file extension in the URL.
	MimeType string

	// actual URL of the resource
	// 	if there were e.g. HTTP redirects.
	RedirectedUrl string

	// filename of the resource. Usually
	// 	derived from the *filename* parameter in a *Content-Disposition*
	// 	header
	Filename string

	ProtocolEncoding string
}

type UrlFetcher = func(url string) (RemoteRessource, error)

type BytesCloser bytes.Reader

func (g *BytesCloser) Read(p []byte) (n int, err error) {
	return (*bytes.Reader)(g).Read(p)
}

func (g *BytesCloser) ReadAt(p []byte, off int64) (n int, err error) {
	return (*bytes.Reader)(g).ReadAt(p, off)
}

func (g *BytesCloser) Seek(off int64, whence int) (n int64, err error) {
	return (*bytes.Reader)(g).Seek(off, whence)
}

func (BytesCloser) Close() error { return nil }

// NewBytesCloser is a bytes.Reader with a Close method.
func NewBytesCloser(data []byte) *BytesCloser {
	return (*BytesCloser)(bytes.NewReader(data))
}

// Fetch an external resource such as an image or stylesheet.
func DefaultUrlFetcher(urlTarget string) (RemoteRessource, error) {
	if strings.HasPrefix(strings.ToLower(urlTarget), "data:") {
		// data url can't contains spaces and the strings comming from css
		// may contain tabs when separated on several lines with \
		urlTarget = htmlSpacesRe.ReplaceAllString(urlTarget, "")
		data, err := parseDataURL([]byte(urlTarget))
		if err != nil {
			return RemoteRessource{}, err
		}
		return data.toResource(urlTarget)
	}

	data, err := url.Parse(urlTarget)
	if err != nil {
		return RemoteRessource{}, err
	}
	if !data.IsAbs() {
		return RemoteRessource{}, fmt.Errorf("Not an absolute URI: %s", urlTarget)
	}
	urlTarget = data.String()

	if data.Scheme == "file" {
		var f content
		f, err = os.Open(data.Path)
		if err != nil {
			return RemoteRessource{}, fmt.Errorf("local file not found : %s", err)
		}
		return RemoteRessource{
			Content:       f,
			Filename:      filepath.Base(data.Path),
			MimeType:      mime.TypeByExtension(filepath.Ext(data.Path)),
			RedirectedUrl: urlTarget,
		}, nil
	}

	req, err := http.NewRequest(http.MethodGet, urlTarget, nil)
	if err != nil {
		return RemoteRessource{}, err
	}
	req.Header.Set("User-Agent", version.VersionString)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return RemoteRessource{}, err
	}
	defer response.Body.Close()

	result := RemoteRessource{}
	redirect, err := response.Location()
	if err == nil {
		result.RedirectedUrl = redirect.String()
	}
	mediaType, params, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err == nil {
		result.MimeType = mediaType
		result.ProtocolEncoding = params["charset"]
	}
	_, params, err = mime.ParseMediaType(response.Header.Get("Content-Disposition"))
	if err == nil {
		result.Filename = params["filename"]
	}

	contentEncoding := response.Header.Get("Content-Encoding")
	var r io.Reader
	if contentEncoding == "gzip" {
		r, err = gzip.NewReader(response.Body)
		if err != nil {
			return RemoteRessource{}, err
		}
	} else if contentEncoding == "deflate" {
		r, err = zlib.NewReader(response.Body)
		if err != nil {
			return RemoteRessource{}, err
		}
	} else {
		r = response.Body
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return RemoteRessource{}, err
	}
	result.Content = (*BytesCloser)(bytes.NewReader(buf.Bytes()))

	return result, nil
}

// dataURI represents the parsed "data" URL
type dataURI struct {
	params   map[string]string
	mimeType string
	data     []byte // before decoding
	isBase64 bool
}

// decode the base64 or ascii encoding, but not the charset
func (d dataURI) toResource(urlTarget string) (RemoteRessource, error) {
	var err error
	d.data, err = unescape(d.data)
	if err != nil {
		return RemoteRessource{}, err
	}
	if d.isBase64 {
		dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(d.data)))
		n, err := base64.StdEncoding.Decode(dbuf, d.data)
		if err != nil {
			return RemoteRessource{}, fmt.Errorf("invalid base64 data url: %s", err)
		}
		d.data = dbuf[:n]
	}
	return RemoteRessource{
		Content:          (*BytesCloser)(bytes.NewReader(d.data)),
		MimeType:         d.mimeType,
		RedirectedUrl:    urlTarget,
		ProtocolEncoding: d.params["charset"],
	}, nil
}

// parseDataURL parse the "data" URL into components.
func parseDataURL(url []byte) (dataURI, error) {
	// adapted from https://onethinglab.com/data-url-parse-in-golang
	const (
		dataURIPrefix   = "data:"
		defaultMimeType = "text/plain"
		defaultParam    = "charset=US-ASCII"
		base64Indicator = "base64"
	)

	data := url[len(dataURIPrefix):]
	// split properties and actual encoded data
	indexSep := bytes.IndexByte(data, ',')
	if indexSep == -1 {
		return dataURI{}, errors.New("data not found in Data URI")
	}
	properties, encodedData := string(data[:indexSep]), data[indexSep+1:]

	var result dataURI = dataURI{
		data:   encodedData,
		params: make(map[string]string),
	}
	for i, prop := range strings.Split(properties, ";") {
		if i == 0 {
			if strings.Contains(prop, "/") {
				result.mimeType = prop
			} else {
				params := strings.Split(defaultParam, "=")
				result.mimeType = defaultMimeType
				result.params[params[0]] = params[1]
			}
		} else {
			if prop == base64Indicator {
				result.isBase64 = true
			} else {
				// ignore if not valid properties assignment
				if strings.Contains(prop, "=") {
					propComponets := strings.SplitN(prop, "=", 2)
					result.params[propComponets[0]] = propComponets[1]
				}
			}
		}
	}

	return result, nil
}

func isHex(c byte) bool {
	switch {
	case c >= 'a' && c <= 'f':
		return true
	case c >= 'A' && c <= 'F':
		return true
	case c >= '0' && c <= '9':
		return true
	}
	return false
}

// borrowed from net/url/url.go
func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

// unescape unescapes a character sequence
// escaped with Escape(String?).
func unescape(s []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	reader := bytes.NewReader(s)

	for {
		r, size, err := reader.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if size > 1 {
			return nil, fmt.Errorf("rfc2396: non-ASCII char detected")
		}

		switch r {
		case '%':
			eb1, err := reader.ReadByte()
			if err == io.EOF {
				return nil, fmt.Errorf("rfc2396: unexpected end of unescape sequence")
			}
			if err != nil {
				return nil, err
			}
			if !isHex(eb1) {
				return nil, fmt.Errorf("rfc2396: invalid char 0x%x in unescape sequence", r)
			}
			eb0, err := reader.ReadByte()
			if err == io.EOF {
				return nil, fmt.Errorf("rfc2396: unexpected end of unescape sequence")
			}
			if err != nil {
				return nil, err
			}
			if !isHex(eb0) {
				return nil, fmt.Errorf("rfc2396: invalid char 0x%x in unescape sequence", r)
			}
			buf.WriteByte(unhex(eb0) + unhex(eb1)*16)
		default:
			buf.WriteByte(byte(r))
		}
	}
	return buf.Bytes(), nil
}
