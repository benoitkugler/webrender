package utils

import (
	"io/ioutil"
	"net/url"
	"reflect"
	"testing"
)

func TestURLFetcher(t *testing.T) {
	u := url.PathEscape("<svg width='4' height='4'></svg>")
	content, err := DefaultUrlFetcher("data:image/svg+xml," + u)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(content.Content)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "<svg width='4' height='4'></svg>" {
		t.Fatalf("unexpected %s", b)
	}
}

func TestFetchUrlData(t *testing.T) {
	testData := func(input, expected string) {
		tmp, err := parseDataURL([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		res, err := tmp.toResource("")
		if err != nil {
			t.Fatal(err)
		}
		b, err := decodeToUtf8(res.Content, res.ProtocolEncoding)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(b); got != expected && expected != "XXX" {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}

	testData("data:image/svg+xml,<svg></svg>", "<svg></svg>")
	testData("data:image/svg+xml,<svg width='4' height='4'></svg>", "<svg width='4' height='4'></svg>")
	testData("data:text/css;charset=ASCII,a%7Bcolor%3AcurrentColor%7D", "a{color:currentColor}")
	testData(`data:text/css;charset=utf-16le;base64,bABpAHsAYwBvAGwAbwByADoAcgBlAGQAfQA=`, "li{color:red}")
	testData("data:,ul {border-width: 1000px !important}", "ul {border-width: 1000px !important}")
	testData(`data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAGAAAAAYCAYAAAFy7sgCAAAGsUlEQVRo3u2ZbWwcZxHHf3s%2B7LNbO3ZjXBtowprGODRX0qpNQCjmJKuVKhMl1P2AkCwhFOIKkCBSm9IXavGFKAixIAECwkmWo5MrhRI3Ub40IEwQgp6aIDg3Cd6eEqyIHEteah%2B1E69vhw%2BZtTaX8704ZzkKjHS6271nZ56ZZ%2BY%2F%2F%2BdZKF%2FCwYshx3EkkggLsD1v4FQkEZZYLCbAKyG9%2Ba9EIsG6hnUAf8x74K3aUC3j4%2BM54HcsR2oAIomwZOezkv%2FnSHpYNh%2BNCmAE7xv94zvFdd1bHsjMZmQkPSxAJP%2B%2FfuBLwK54PC7JZFKAVJmzXLBt2w%2FMvcDLwIb8QS8CeJ4nkURYIomw7J%2FYJ8BvSiiXptGGxWds2%2Fa9%2Bnaxh%2BYAD%2Bgt04NDgABTpQY2cvvSFLzw86gWeBVwC8SzlOSv2YeBPfmDBoBHgKmR9LBEEmHZfDTqGykqfkUE0nA78BzQGfSgUeP3wNeTXwXg7MwZDhw4UHL6ra2ti79%2FOvljgG8AZ4H64Lhm4MvAocxsRppGG%2FxcXihlwLIs6R%2FfKV2HO%2F26uA94pdDYUKUZUU7W1RQYXA98Gnhaf5%2FXWX0HeAHYoQonqa4sZSOsSWMCWeC9Yko%2BCQwBe4E6oNc0Tc91XTl1%2BaTsn9gnI%2Blhyc5nZWxsrBIkKSbl2tiic3tW53YDEwOKaoFBrcOfqKee53lG9xsPMjV784r%2F4lO%2FpPvyJ9iyZcuvFSaXK5XYeAZ4CDgGvB3MS4B54LQuWYPeuy4iRFsevsXqpuYoqVQKIH2bK1CuDQNo11o4XUzh%2FcDWYIe1LEtyuZx4niee54njOGKapgfsqlL%2Bl2OjEXg8nxrc1dJ0h3hbtL%2BGCtz7KPBF4CuBe9uB15VafE8hr9qylI3HgG8C2%2FK7VyHZoJj7MrBRm30qFotJMpkU27YlHo%2F7Ha5a%2BV%2FKRkSJ4KuKRLVLKapTjB1SzAVIjY2NSXY%2BKyPpYdk%2FsU9OXT4pruv6BdZbBQfKsVGnvWlIe1VB6VQO8JxC1vZYLCbZ%2BaxsPhpdZDyRRFhG0sPiOE6ldKBg2lRg4xF1YCDIIIKN7DGgD3gH%2BBXwejKZfPrs2tPs%2FvPN2bKuYR1nd7xLKBSSJeqoXKnERjPwNWAG%2BLn2rZuM%2B4Tpml6vaWlp4eLcxVusZq5lCgVgOVKJjRqdX86ffL4D5wIoZACnTpw4wRMdT96i%2FImOJxERAs4uVyqxUacF%2FPdiCj%2BjdRBRGFtwXVdG0sPSdbhTmkYbpH98p2RmM2JZlig1vl0GWo4NQ%2Fn%2Bs5pKRXfwjweaxy7TND3HcRZbfC6X8xVPVQlGy7WxVWlO5XRXFXm6EZmrQuSXYyPE3SiVoEhE6Wyr0u2rumO6zv%2B21AFdQAswC1wCMuUCXCmyWQus103Qg8qlDO0lxwOb%2Fl4FiK3AB3VS%2FuKKLtK%2FgbeAnwG%2FvUODuRw%2FFrR0H1UC75fwu8oJ%2FhFsW5VIG%2FBUgEIN6Y65O4AHu4Ap0zQ9y7LEcZyb9lRBUHQcRyzL8unZVBW5bFWAvAp%2BhDQ2g4F47dUYtlU6obXA54DnVdFLekjUGGifh4AFy7LEdV3xj3X9I66m0QZpGm2QrsOd0j%2B%2BU0bSw5KZzYjrun6HWlAd961i4FfCj0aN1Usau%2Bc1lmuXPFwvAEumUut7tQQvAb%2FXb%2FT0bCAej9cODg7yt%2Bm%2F8q2%2F7OUHZ76PnZ1k2p0mJzlykmPancbOTnL0whHs7CQfb%2B5mx2d3sH79%2BtCRI0c6FeaOr9ICrIQfLvA%2B8BGNXxi4R6HrisJVUWrxAVW2oMFf0Aczim8o3kV6enowDIPjF9%2Fk%2BMU3S3rrjzMMg56eHr%2BxP7qKFbASfojG6kpeDGs1tiW53RxwWT%2Bin5q8w4xpQK5evQpAR30H7ZH2khNvj7TTUd8BgD4rqmu1ZKX8qNeY%2BfHz4zlXDgT5E8tpCTUq7XSBC4Euv8227TV9fX1E73%2BYtvo27BmbS9cvFVTY3bSRFza9yOcf6Gfmygy7d%2B%2Fm%2FPnzF4DvrsBLhnJlJfwIKXxv1PheAE4qK6p4H9AGbNKTuhngBPBPXYRe4IemaT5kWZbR19fHNbmGnZ1k4r3U4glDR30Hm5qjbGjsImJEOHbsGHv27JFz5869o0eFq01Jq%2BmHAXwI6FFKagMTgHM7GzFDS%2BoeLSMv7zjzC9x4Y7gxFovVDAwMEI1GaWlpWSzRVCrFwYMH%2FXfxZ4AfAa8B%2F7lDaGg1%2FQgp43lfK0yqtRMuJa3ceKe5DfgYsCYAZ2ngD8CfAkzqTpW7xY%2F%2FSznyX%2FVeUb2kVmX4AAAAAElFTkSuQmCC`, "XXX")
}

func TestParseDataURL(t *testing.T) {
	type testCase struct {
		dataURL    string
		parsedData dataURI
	}

	testCases := []testCase{
		{
			"data:image/gif;base64,XXX",
			dataURI{
				mimeType: "image/gif", isBase64: true,
				data:   []byte("XXX"),
				params: make(map[string]string),
			},
		},
		{
			`data:,A%20brief%20note`,
			dataURI{
				mimeType: "text/plain", data: []byte(`A%20brief%20note`),
				params: map[string]string{"charset": "US-ASCII"},
			},
		},
		{
			`data:text/plain;charset=iso-8859-7,%be%fg%be`,
			dataURI{
				mimeType: "text/plain", data: []byte("%be%fg%be"),
				params: map[string]string{"charset": "iso-8859-7"},
			},
		},
		{
			`data:application/vnd-xxx-query,select_vcount,fcol_from_fieldtable/local`,
			dataURI{
				mimeType: "application/vnd-xxx-query",
				data:     []byte("select_vcount,fcol_from_fieldtable/local"),
				params:   make(map[string]string),
			},
		},
		{`data:,`, dataURI{
			mimeType: "text/plain",
			params:   map[string]string{"charset": "US-ASCII"},
			data:     []byte{},
		}},
		{`data:boo/foo;,`, dataURI{
			mimeType: "boo/foo",
			params:   make(map[string]string),
			data:     []byte{},
		}},
	}

	for _, test := range testCases {
		t.Run(test.dataURL, func(t *testing.T) {
			result, _ := parseDataURL([]byte(test.dataURL))
			if !reflect.DeepEqual(result, test.parsedData) {
				t.Errorf("ParseDataURL expected %+v\n, got %+v", test.parsedData, result)
			}
		})
	}
}

func TestJoinUrl(t *testing.T) {
	for _, data := range [][3]string{
		{"http://weasyprint.org/css/main.css", "/css/files/oter.css", "http://weasyprint.org/css/files/oter.css"},
		{"http://weasyprint.org/css/main.css", "css/files/oter.css", "http://weasyprint.org/css/css/files/oter.css"},
		{
			"https://en.wikipedia.org/wiki/Go_(programming_language)",
			"/w/load.php?lang=en&modules=ext.cite.styles%7Cext.pygments%2CwikimediaBadges%7Cext.uls.interlanguage%7Cext.visualEditor.desktopArticleTarget.noscript%7Cjquery.makeCollapsible.styles%7Cskins.vector.styles.legacy%7Cwikibase.client.init&only=styles&skin=vector",
			"https://en.wikipedia.org/w/load.php?lang=en&modules=ext.cite.styles%7Cext.pygments%2CwikimediaBadges%7Cext.uls.interlanguage%7Cext.visualEditor.desktopArticleTarget.noscript%7Cjquery.makeCollapsible.styles%7Cskins.vector.styles.legacy%7Cwikibase.client.init&only=styles&skin=vector",
		},
		{"https://en.wikipedia.org/wiki/Go_(programming_language)", "//upload.wikimedia.org/wikipedia/commons/thumb/0/05/Go_Logo_Blue.svg/215px-Go_Logo_Blue.svg.png", "https://upload.wikimedia.org/wikipedia/commons/thumb/0/05/Go_Logo_Blue.svg/215px-Go_Logo_Blue.svg.png"},
	} {
		baseUrl, urls, exp := data[0], data[1], data[2]
		parsedUrl, err := url.Parse(urls)
		if err != nil {
			t.Fatal(err)
		}
		if s, err := basicUrlJoin(baseUrl, parsedUrl); err != nil || s != exp {
			t.Fatalf("expected\n%s\n, got\n%s", exp, s)
		}
	}
}

func TestUrlJoin(t *testing.T) {
	got, err := SafeUrljoin("http://weasyprint.org/foo/bar/", "/foo/bar/#test", true)
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://weasyprint.org/foo/bar/#test" {
		t.Fatalf("unexpected joined url %s", got)
	}
}
