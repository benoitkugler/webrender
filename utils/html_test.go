package utils

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var dates = [...]string{
	"1997",
	"1997-07",
	"1997-07-16",
	"1997-07-16T19:20+01:00",
	"1997-07-16T19:20:30+01:00",
	"1997-07-16T19:20:30.45+01:00",
	"1997-07-16T19:20:30.45-01:15",
	"2011-04-21T23:00:00Z",
}

func TestReW3CDate(t *testing.T) {
	for _, date := range dates {
		if !w3CDateRe.MatchString(date) {
			t.Fatalf("date %s not matching", date)
		}

		if parseW3cDate("test", date).IsZero() {
			t.Fatalf("date %s not parsed", date)
		}
	}
}

func TestWalkHtml(t *testing.T) {
	s := "<html><p>dlfkdfk</p><div><span>sdsd/<span><span></span></div></html>"
	root, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(root.FirstChild)
	iter := NewHtmlIterator(root.FirstChild)
	for iter.HasNext() {
		n := iter.Next()
		fmt.Printf("%p %v %s\n", n, n.DataAtom, n.Data)
	}
}

func TestHtmlIterator(t *testing.T) {
	n, err := html.Parse(strings.NewReader("<meta></meta><link></link><p><a>sldmskd</a></p>"))
	if err != nil {
		t.Fatal(err)
	}
	iter := NewHtmlIterator(n, atom.Meta, atom.Link)
	for iter.HasNext() {
		elem := iter.Next()
		if elem == nil {
			t.Fatal("inconstistent iterator")
		}
	}
}

func TestMeta(t *testing.T) {
	assertMeta := func(input string, exp DocumentMetadata) {
		if exp.Authors == nil {
			exp.Authors = []string{}
		}
		if exp.Keywords == nil {
			exp.Keywords = []string{}
		}
		if exp.Attachments == nil {
			exp.Attachments = []Attachment{}
		}
		root_, err := html.Parse(strings.NewReader(input))
		if err != nil || root_.FirstChild == nil {
			t.Fatalf("invalid html input : %s", err)
		}
		root := (*HTMLNode)(root_.FirstChild) // html.Parse wraps the <html> tag
		got := GetHtmlMetadata(root, FindBaseUrl(root_, ""))
		if !reflect.DeepEqual(got, exp) {
			t.Fatalf("expected %v, got%v", exp, got)
		}
	}
	assertMeta("<body>", DocumentMetadata{})
	assertMeta(`
            <meta name=author content="I Me &amp; Myself">
            <meta name=author content="Smith, John">
            <title>Test document</title>
            <h1>Another title</h1>
            <meta name=generator content="Human after all">
            <meta name=dummy content=ignored>
            <meta name=dummy>
            <meta content=ignored>
            <meta>
            <meta name=keywords content="html ,`+"\t"+`css,
                                         pdf,css">
            <meta name=dcterms.created content=2011-04>
            <meta name=dcterms.created content=2011-05>
            <meta name=dcterms.modified content=2013>
            <meta name=keywords content="Python; pydyf">
            <meta name=description content="Blah… ">
        `, DocumentMetadata{
		Authors:     []string{"I Me & Myself", "Smith, John"},
		Title:       "Test document",
		Generator:   "Human after all",
		Keywords:    []string{"html", "css", "pdf", "Python; pydyf"},
		Description: "Blah… ",
		Created:     time.Date(2011, 4, 1, 0, 0, 0, 0, time.UTC),
		Modified:    time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	assertMeta(`
            <title>One</title>
            <meta name=Author>
            <title>Two</title>
            <title>Three</title>
            <meta name=author content=Me>
        `, DocumentMetadata{
		Title:   "One",
		Authors: []string{"", "Me"},
	})
}
