package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"golang.org/x/net/html/charset"
)

type ContentInput interface {
	isContentInput()
	String() string
}

type (
	InputFilename string
	InputUrl      string
	InputString   string
	InputReader   struct {
		io.ReadCloser
	}
)

func (c InputFilename) isContentInput() {}
func (c InputUrl) isContentInput()      {}
func (c InputString) isContentInput()   {}
func (c InputReader) isContentInput()   {}
func (c InputFilename) String() string {
	return string(c)
}

func (c InputUrl) String() string {
	if strings.HasPrefix(string(c), "data:") {
		return fmt.Sprintf("data url of len. %d", len(c))
	}
	return string(c)
}

func (c InputString) String() string {
	return fmt.Sprintf("string of len. %d", len(c))
}

func (c InputReader) String() string {
	return fmt.Sprintf("reader of type %T", c.ReadCloser)
}

type Source struct {
	BaseUrl string
	Content []byte // utf8 encoded
}

// FetchSource fetch the html input, and returns it with the
// normalized ``BaseUrl`` (checkCssMimeType=false).
func FetchSource(input ContentInput, baseUrl string, urlFetcher UrlFetcher,
	checkCssMimeType bool) (out Source, err error) {

	if baseUrl != "" {
		baseUrl, err = ensureUrl(baseUrl)
		if err != nil {
			return
		}
	}
	switch data := input.(type) {
	case InputFilename:
		if baseUrl == "" {
			baseUrl, err = PathToURL(string(data))
			if err != nil {
				return
			}
		}
		f, err := ioutil.ReadFile(string(data))
		if err != nil {
			return Source{}, err
		}
		return Source{Content: f, BaseUrl: baseUrl}, nil
	case InputUrl:
		result, err := urlFetcher(string(data))
		if err != nil {
			return Source{}, err
		}
		if result.RedirectedUrl == "" {
			result.RedirectedUrl = string(data)
		}
		if checkCssMimeType && result.MimeType != "text/css" {
			log.Printf("Unsupported stylesheet type %s for %s",
				result.MimeType, result.RedirectedUrl)
			return Source{BaseUrl: baseUrl}, nil
		} else {
			if baseUrl == "" {
				baseUrl = result.RedirectedUrl
			}
			decoded, err := decodeToUtf8(result.Content, result.ProtocolEncoding)
			if err != nil {
				return Source{}, err
			}
			return Source{Content: decoded, BaseUrl: baseUrl}, nil
		}
	case InputReader:
		bt, err := ioutil.ReadAll(data.ReadCloser)
		if err != nil {
			return Source{}, err
		}
		if err = data.ReadCloser.Close(); err != nil {
			return Source{}, err
		}
		return Source{Content: bt, BaseUrl: baseUrl}, nil
	case InputString:
		return Source{Content: []byte(data), BaseUrl: baseUrl}, nil
	default:
		return Source{}, fmt.Errorf("unexpected input type %T", input)
	}
}

func decodeToUtf8(data io.Reader, encod string) ([]byte, error) {
	if encod == "" { // assume UTF8
		return ioutil.ReadAll(data)
	}
	enc, _ := charset.Lookup(encod)
	if enc == nil {
		return nil, fmt.Errorf("unsupported encoding %s", encod)
	}
	return ioutil.ReadAll(enc.NewDecoder().Reader(data))
}
