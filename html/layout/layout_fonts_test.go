package layout

import (
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/truetype"
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
	defer tu.CaptureLogs().AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body { font-family: weasyprint }
      </style>
      <span>abc</span>`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	tu.AssertEqual(t, line.Box().Width, Fl(3*16))
}

func TestKerningDefault(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	// Kerning and ligatures are on by default
	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body { font-family: weasyprint }
      </style>
      <span>kk</span><span>liga</span>`)
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span1, span2 := unpack1(line), line.Box().Children[1]
	tu.AssertEqual(t, span1.Box().Width, Fl(1.5*16))
	tu.AssertEqual(t, span2.Box().Width, Fl(1.5*16))
}

func TestKerningDeactivate(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span1, span2 := unpack1(line), line.Box().Children[1]
	tu.AssertEqual(t, span1.Box().Width, Fl(1.5*16))
	tu.AssertEqual(t, span2.Box().Width, Fl(2*16))
}

func TestKerningLigatureDeactivate(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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
	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	span1, span2 := unpack1(line), line.Box().Children[1]
	tu.AssertEqual(t, span1.Box().Width, Fl((1.5+1+1.5)*16))
	tu.AssertEqual(t, span2.Box().Width, Fl((2+1+4)*16))
}

func TestFontFaceDescriptors(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

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

	html := unpack1(page)
	body := unpack1(html)
	line := unpack1(body)
	kern, subs, dlig, onum, zero := unpack5(line)
	tu.AssertEqual(t, kern.Box().Width, Fl(1.5*16))
	tu.AssertEqual(t, subs.Box().Width, Fl(1.5*16))
	tu.AssertEqual(t, dlig.Box().Width, Fl(1.5*16))
	tu.AssertEqual(t, onum.Box().Width, Fl(1.5*16))
	tu.AssertEqual(t, zero.Box().Width, Fl(1.5*16))
}
