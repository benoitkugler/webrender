package layout

import (
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/truetype"
	pr "github.com/benoitkugler/webrender/css/properties"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func TestLoadFont(t *testing.T) {
	f, err := os.Open("../../resources_test/weasyprint.otf")
	if err != nil {
		t.Fatal(err)
	}

	font, err := truetype.Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	gsub := font.LayoutTables().GSUB

	_, ok := gsub.FindFeatureIndex(truetype.MustNewTag("liga"))
	if !ok {
		t.Fatal()
	}
}

func TestFontFace(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body { font-family: weasyprint }
      </style>
      <span>abc</span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	tu.AssertEqual(t, line.Box().Width, pr.Float(3*16), "line")
}

func TestKerningDefault(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Kerning and ligatures are on by default
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body { font-family: weasyprint }
      </style>
      <span>kk</span><span>liga</span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span1, span2 := line.Box().Children[0], line.Box().Children[1]
	tu.AssertEqual(t, span1.Box().Width, pr.Float(1.5*16), "span1")
	tu.AssertEqual(t, span2.Box().Width, pr.Float(1.5*16), "span2")
}

func TestKerningDeactivate(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Deactivate kerning
	page := renderOnePage(t, `
      <style>
        @font-face {
          src: url(weasyprint.otf);
          font-family: no-kern;
          font-feature-settings: 'kern' off;
        }
        @font-face {
          src: url(weasyprint.otf);
          font-family: kern;
        }
        span:nth-child(1) { font-family: kern }
        span:nth-child(2) { font-family: no-kern }
      </style>
      <span>kk</span><span>kk</span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span1, span2 := line.Box().Children[0], line.Box().Children[1]
	tu.AssertEqual(t, span1.Box().Width, pr.Float(1.5*16), "span1")
	tu.AssertEqual(t, span2.Box().Width, pr.Float(2*16), "span2")
}

func TestKerningLigatureDeactivate(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Deactivate kerning and ligatures
	page := renderOnePage(t, `
      <style>
        @font-face {
          src: url(weasyprint.otf);
          font-family: no-kern-liga;
          font-feature-settings: 'kern' off;
          font-variant: no-common-ligatures;
        }
        @font-face {
          src: url(weasyprint.otf);
          font-family: kern-liga;
        }
        span:nth-child(1) { font-family: kern-liga }
        span:nth-child(2) { font-family: no-kern-liga }
      </style>
      <span>kk liga</span><span>kk liga</span>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	span1, span2 := line.Box().Children[0], line.Box().Children[1]
	tu.AssertEqual(t, span1.Box().Width, pr.Float((1.5+1+1.5)*16), "span1")
	tu.AssertEqual(t, span2.Box().Width, pr.Float((2+1+4)*16), "span2")
}

func TestFontFaceDescriptors(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t,
		`
        <style>
          @font-face {
            src: url(weasyprint.otf);
            font-family: weasyprint;
            font-variant: sub
                          discretionary-ligatures
                          oldstyle-nums
                          slashed-zero;
          }
          span { font-family: weasyprint }
        </style>`+
			"<span>kk</span>"+
			"<span>subs</span>"+
			"<span>dlig</span>"+
			"<span>onum</span>"+
			"<span>zero</span>'")

	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	kern, subs, dlig, onum, zero := unpack5(line)
	tu.AssertEqual(t, kern.Box().Width, pr.Float(1.5*16), "kern")
	tu.AssertEqual(t, subs.Box().Width, pr.Float(1.5*16), "subs")
	tu.AssertEqual(t, dlig.Box().Width, pr.Float(1.5*16), "dlig")
	tu.AssertEqual(t, onum.Box().Width, pr.Float(1.5*16), "onum")
	tu.AssertEqual(t, zero.Box().Width, pr.Float(1.5*16), "zero")
}
