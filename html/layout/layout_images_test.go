package layout

import (
	"fmt"
	"math"
	"net/url"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/images"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

//  Tests for image layout.

func getImg(t *testing.T, input string) (Box, Box) {
	page := renderOnePage(t, input)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img := line.Box().Children[0]
	return body, img
}

func TestImages1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, html := range []string{
		fmt.Sprintf(`<img src="%s">`, "pattern.svg"),
		fmt.Sprintf(`<img src="%s">`, "pattern.png"),
		fmt.Sprintf(`<img src="%s">`, "pattern.gif"),
		fmt.Sprintf(`<img src="%s">`, "blue.jpg"),
		fmt.Sprintf(`<img src="%s">`, "data:image/svg+xml,"+url.PathEscape(`<svg width='4' height='4'></svg>`)),
		fmt.Sprintf(`<img src="%s">`, "DatA:image/svg+xml,"+url.PathEscape(`<svg width='4px' height='4px'></svg>`)),

		"<embed src=pattern.png>",
		"<embed src=pattern.svg>",
		"<embed src=really-a-png.svg type=image/png>",
		"<embed src=really-a-svg.png type=image/svg+xml>",

		"<object data=pattern.png>",
		"<object data=pattern.svg>",
		"<object data=really-a-png.svg type=image/png>",
		"<object data=really-a-svg.png type=image/svg+xml>",
	} {
		_, img := getImg(t, html)
		tu.AssertEqual(t, img.Box().Width, pr.Float(4), html+": width")
		tu.AssertEqual(t, img.Box().Height, pr.Float(4), html+": height")
	}
}

func TestImages2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// With physical units
	data := "data:image/svg+xml," + url.PathEscape(`<svg width="2.54cm" height="0.5in"></svg>`)
	_, img := getImg(t, fmt.Sprintf(`<img src="%s">`, data))
	tu.AssertEqual(t, img.Box().Width, pr.Float(96), "")
	tu.AssertEqual(t, img.Box().Height, pr.Float(48), "")
}

func TestImages3(t *testing.T) {
	// Invalid images
	for _, urlString := range []string{
		"nonexistent.png",
		"unknownprotocol://weasyprint.org/foo.png",
		"data:image/unknowntype,Not an image",
		// Invalid protocol
		"datå:image/svg+xml," + url.PathEscape(`<svg width="4" height="4"></svg>`),
		// zero-byte images
		"data:image/png,",
		"data:image/jpeg,",
		"data:image/svg+xml,",
		// Incorrect format
		"data:image/png,Not a PNG",
		"data:image/jpeg,Not a JPEG",
		"data:image/svg+xml,<sv invalid xml",
	} {
		capt := tu.CaptureLogs()
		_, img := getImg(t, fmt.Sprintf(`<img src="%s" alt="invalid image">`, urlString))
		tu.AssertEqual(t, len(capt.Logs()), 1, urlString)
		tu.AssertEqual(t, img.Type(), bo.InlineBoxT, urlString) // not a replaced box
		text := img.Box().Children[0]
		tu.AssertEqual(t, text.(*bo.TextBox).Text, "invalid image", "")
	}
}

func TestImages4(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, url := range []string{
		// GIF with JPEG mimetype
		"data:image/jpeg;base64," +
			"R0lGODlhAQABAIABAP///wAAACwAAAAAAQABAAACAkQBADs=",
		// GIF with PNG mimetype
		"data:image/png;base64," +
			"R0lGODlhAQABAIABAP///wAAACwAAAAAAQABAAACAkQBADs=",
		// PNG with JPEG mimetype
		"data:image/jpeg;base64," +
			"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC" +
			"0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=",
		// SVG with PNG mimetype
		"data:image/png," + url.PathEscape(`<svg width="1" height="1"></svg>`),
		"really-a-svg.png",
		// PNG with SVG
		"data:image/svg+xml;base64," +
			"R0lGODlhAQABAIABAP///wAAACwAAAAAAQABAAACAkQBADs=",
		"really-a-png.svg",
	} {
		// Sniffing, no logs
		_, _ = getImg(t, fmt.Sprintf(`<img src="%s">`, url))
	}
}

func TestImages5(t *testing.T) {
	capt := tu.CaptureLogs()
	_ = renderPages(t, "<img src=nonexistent.png><img src=nonexistent.png>")
	// Failures are cached too: only one error
	logs := capt.Logs()
	tu.AssertEqual(t, len(logs), 1, fmt.Sprintf("%v", logs))
	tu.AssertEqual(t, strings.Contains(logs[0], "Failed to load image"), true, "")
}

func TestImages6(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Layout rules try to preserve the ratio, so the height should be 40px too:
	body, img := getImg(t, `<body style="font-size: 0">
        <img src="pattern.png" style="width: 40px">`)
	tu.AssertEqual(t, body.Box().Height, pr.Float(40), "body")
	tu.AssertEqual(t, img.Box().PositionY, pr.Float(0), "img")
	tu.AssertEqual(t, img.Box().Width, pr.Float(40), "img")
	tu.AssertEqual(t, img.Box().Height, pr.Float(40), "img")
}

func TestImages7(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	body, img := getImg(t, `<body style="font-size: 0">
        <img src="pattern.png" style="height: 40px">`)
	tu.AssertEqual(t, body.Box().Height, pr.Float(40), "body")
	tu.AssertEqual(t, img.Box().PositionY, pr.Float(0), "img")
	tu.AssertEqual(t, img.Box().Width, pr.Float(40), "img")
	tu.AssertEqual(t, img.Box().Height, pr.Float(40), "img")
}

func TestImages8(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Same with percentages
	body, img := getImg(t, `<body style="font-size: 0"><p style="width: 200px">
        <img src="pattern.png" style="width: 20%">`)
	tu.AssertEqual(t, body.Box().Height, pr.Float(40), "body")
	tu.AssertEqual(t, img.Box().PositionY, pr.Float(0), "img")
	tu.AssertEqual(t, img.Box().Width, pr.Float(40), "img")
	tu.AssertEqual(t, img.Box().Height, pr.Float(40), "img")
}

func TestImages9(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	body, img := getImg(t, `<body style="font-size: 0">
        <img src="pattern.png" style="min-width: 40px">`)
	tu.AssertEqual(t, body.Box().Height, pr.Float(40), "body")
	tu.AssertEqual(t, img.Box().PositionY, pr.Float(0), "img")
	tu.AssertEqual(t, img.Box().Width, pr.Float(40), "img")
	tu.AssertEqual(t, img.Box().Height, pr.Float(40), "img")
}

func TestImages10(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	_, img := getImg(t, `<img src="pattern.png" style="max-width: 2px">`)
	tu.AssertEqual(t, img.Box().Width, pr.Float(2), "Width")
	tu.AssertEqual(t, img.Box().Height, pr.Float(2), "Height")
}

func TestImages11(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// display: table-cell is ignored. XXX Should it?
	page := renderOnePage(t, `<body style="font-size: 0">
        <img src="pattern.png" style="width: 40px">
        <img src="pattern.png" style="width: 60px; display: table-cell">
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img1, img2 := unpack2(line)
	tu.AssertEqual(t, body.Box().Height, pr.Float(60), "body")
	tu.AssertEqual(t, img1.Box().Width, pr.Float(40), "img1")
	tu.AssertEqual(t, img1.Box().Height, pr.Float(40), "img1")
	tu.AssertEqual(t, img2.Box().Width, pr.Float(60), "img2")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(60), "img2")
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(20), "img1")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(0), "img2")
}

func TestImages12(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Block-level image:
	page := renderOnePage(t, `
        <style>
            @page { size: 100px }
            img { width: 40px; margin: 10px auto; display: block }
        </style>
        <body>
            <img src="pattern.png">
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	img := body.Box().Children[0]
	tu.AssertEqual(t, img.Box().ElementTag(), "img", "")
	tu.AssertEqual(t, img.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, img.Box().PositionY, pr.Float(0), "")
	tu.AssertEqual(t, img.Box().Width, pr.Float(40), "")
	tu.AssertEqual(t, img.Box().Height, pr.Float(40), "")
	tu.AssertEqual(t, img.Box().ContentBoxX(), pr.Float(30), "") // (100 - 40) / 2 , 30px for margin-left
	tu.AssertEqual(t, img.Box().ContentBoxY(), pr.Float(10), "")
}

func TestImages13(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
            @page { size: 100px }
            img { min-width: 40%; margin: 10px auto; display: block }
        </style>
        <body>
            <img src="pattern.png">
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	img := body.Box().Children[0]
	tu.AssertEqual(t, img.Box().ElementTag(), "img", "")
	tu.AssertEqual(t, img.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, img.Box().PositionY, pr.Float(0), "")
	tu.AssertEqual(t, img.Box().Width, pr.Float(40), "")
	tu.AssertEqual(t, img.Box().Height, pr.Float(40), "")
	tu.AssertEqual(t, img.Box().ContentBoxX(), pr.Float(30), "") // (100 - 40) / 2 , 30px for margin-left
	tu.AssertEqual(t, img.Box().ContentBoxY(), pr.Float(10), "")
}

func TestImages14(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
            @page { size: 100px }
            img { min-width: 40px; margin: 10px auto; display: block }
        </style>
        <body>
            <img src="pattern.png">
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	img := body.Box().Children[0]
	tu.AssertEqual(t, img.Box().ElementTag(), "img", "")
	tu.AssertEqual(t, img.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, img.Box().PositionY, pr.Float(0), "")
	tu.AssertEqual(t, img.Box().Width, pr.Float(40), "")
	tu.AssertEqual(t, img.Box().Height, pr.Float(40), "")
	tu.AssertEqual(t, img.Box().ContentBoxX(), pr.Float(30), "") // (100 - 40) / 2 , 30px for margin-left
	tu.AssertEqual(t, img.Box().ContentBoxY(), pr.Float(10), "")
}

func TestImages15(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
        <style>
            @page { size: 100px }
            img { min-height: 30px; max-width: 2px;
                  margin: 10px auto; display: block }
        </style>
        <body>
            <img src="pattern.png">
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	img := body.Box().Children[0]
	tu.AssertEqual(t, img.Box().ElementTag(), "img", "")
	tu.AssertEqual(t, img.Box().PositionX, pr.Float(0), "")
	tu.AssertEqual(t, img.Box().PositionY, pr.Float(0), "")
	tu.AssertEqual(t, img.Box().Width, pr.Float(2), "")
	tu.AssertEqual(t, img.Box().Height, pr.Float(30), "")
	tu.AssertEqual(t, img.Box().ContentBoxX(), pr.Float(49), "") // (100 - 2) / 2 , 49px for margin-left
	tu.AssertEqual(t, img.Box().ContentBoxY(), pr.Float(10), "")
}

func TestImages16(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
        <body style="float: left">
        <img style="height: 200px; margin: 10px; display: block" src="
            data:image/svg+xml,`+url.PathEscape(`<svg width="150" height="100"></svg>`)+`           
        ">
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	img := body.Box().Children[0]
	tu.AssertEqual(t, img.Box().ElementTag(), "img", "")
	tu.AssertEqual(t, body.Box().Width, pr.Float(320), "body")
	tu.AssertEqual(t, body.Box().Height, pr.Float(220), "body")
	tu.AssertEqual(t, img.Box().Width, pr.Float(300), "img")
	tu.AssertEqual(t, img.Box().Height, pr.Float(200), "img")
}

func TestImages17(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
        <div style="width: 300px; height: 300px">
        <img src="data:image/svg+xml,`+url.PathEscape(`<svg viewBox="0 0 20 10"></svg>`)+`
        ">`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	line := div.Box().Children[0]
	img := line.Box().Children[0]
	tu.AssertEqual(t, img.Box().ElementTag(), "img", "")
	tu.AssertEqual(t, div.Box().Width, pr.Float(300), "div")
	tu.AssertEqual(t, div.Box().Height, pr.Float(300), "div")
	tu.AssertEqual(t, img.Box().Width, pr.Float(300), "img")
	tu.AssertEqual(t, img.Box().Height, pr.Float(150), "img")
}

func TestImages18(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1050
	_ = renderOnePage(t, `
        <img style="position: absolute" src="
            data:image/svg+xml,`+url.PathEscape(`<svg viewBox="0 0 20 10"></svg>`)+`
			">`)
}

func approxEqual(t *testing.T, got, exp pr.Fl, context string) {
	t.Helper()
	const float32EqualityThreshold = 1e-3

	if diff := math.Abs(float64(exp - got)); diff > float32EqualityThreshold {
		t.Fatalf("%s: expected %v, got %v (diff: %v)", context, exp, got, diff)
	}
}

func approxEqualSlice(t *testing.T, got, exp []pr.Fl, context string) {
	t.Helper()

	tu.AssertEqual(t, len(got), len(exp), context)
	for i := range got {
		approxEqual(t, got[i], exp[i], context)
	}
}

func approxEqualColor(t *testing.T, got, exp parser.RGBA, context string) {
	t.Helper()

	s1 := []pr.Fl{got.R, got.G, got.B, got.A}
	s2 := []pr.Fl{exp.R, exp.G, exp.B, exp.A}
	approxEqualSlice(t, s1, s2, context)
}

func TestLinearGradient(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	red := pr.NewColor(1, 0, 0, 1).RGBA
	lime := pr.NewColor(0, 1, 0, 1).RGBA
	blue := pr.NewColor(0, 0, 1, 1).RGBA

	// type="linear" positions=[0, 1] colors = [blue, lime]
	layout := func(gradientCss string, type_ string, init [6]pr.Fl,
		positions []pr.Fl, colors []parser.RGBA) {
		page := renderOnePage(t, "<style>@page { background: "+gradientCss)
		layer := page.Background.Layers[0]
		result := layer.Image.(images.LinearGradient).Layout(400, 300)
		tu.AssertEqual(t, result.ScaleY, pr.Fl(1), "result[0]")
		tu.AssertEqual(t, result.Kind, type_, "result[1]")
		approxEqualSlice(t, result.GradientKind.Coords[:], init[:], gradientCss+": Data")
		tu.AssertEqual(t, result.Positions, positions, "Positions")
		tu.AssertEqual(t, len(result.Colors) >= len(colors), true, "colors length")
		for i := range colors {
			color1, color2 := result.Colors[i], colors[i]
			approxEqualColor(t, color1, color2, "color for "+gradientCss)
		}
	}

	layout("linear-gradient(blue)", "solid", [6]pr.Fl{}, nil, []parser.RGBA{blue})
	layout("repeating-linear-gradient(blue)", "solid", [6]pr.Fl{}, nil, []parser.RGBA{blue})
	layout("linear-gradient(blue, lime)", "linear", [6]pr.Fl{200, 0, 200, 300}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("repeating-linear-gradient(blue, lime)", "linear", [6]pr.Fl{200, 0, 200, 300}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("repeating-linear-gradient(blue, lime 100px)", "linear", [6]pr.Fl{200, 0, 200, 300}, []pr.Fl{0, 1, 1, 2, 2, 3}, []parser.RGBA{blue, lime})

	layout("linear-gradient(to bottom, blue, lime)", "linear", [6]pr.Fl{200, 0, 200, 300}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(to top, blue, lime)", "linear", [6]pr.Fl{200, 300, 200, 0}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(to right, blue, lime)", "linear", [6]pr.Fl{0, 150, 400, 150}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(to left, blue, lime)", "linear", [6]pr.Fl{400, 150, 0, 150}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})

	layout("linear-gradient(to top left, blue, lime)", "linear", [6]pr.Fl{344, 342, 56, -42}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(to top right, blue, lime)", "linear", [6]pr.Fl{56, 342, 344, -42}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(to bottom left, blue, lime)", "linear", [6]pr.Fl{344, -42, 56, 342}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(to bottom right, blue, lime)", "linear", [6]pr.Fl{56, -42, 344, 342}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})

	layout("linear-gradient(270deg, blue, lime)", "linear", [6]pr.Fl{400, 150, 0, 150}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(.75turn, blue, lime)", "linear", [6]pr.Fl{400, 150, 0, 150}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(45deg, blue, lime)", "linear", [6]pr.Fl{25, 325, 375, -25}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(.125turn, blue, lime)", "linear", [6]pr.Fl{25, 325, 375, -25}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(.375turn, blue, lime)", "linear", [6]pr.Fl{25, -25, 375, 325}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(.625turn, blue, lime)", "linear", [6]pr.Fl{375, -25, 25, 325}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(.875turn, blue, lime)", "linear", [6]pr.Fl{375, 325, 25, -25}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})

	layout("linear-gradient(blue 2em, lime 20%)", "linear", [6]pr.Fl{200, 32, 200, 60}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime})
	layout("linear-gradient(blue 100px, red, blue, red 160px, lime)", "linear", [6]pr.Fl{200, 100, 200, 300}, []pr.Fl{0, .1, .2, .3, 1}, []parser.RGBA{blue, red, blue, red, lime})
	layout("linear-gradient(blue -100px, blue 0, red -12px, lime 50%)", "linear", [6]pr.Fl{200, -100, 200, 150}, []pr.Fl{0, .4, .4, 1}, []parser.RGBA{blue, blue, red, lime})
	layout("linear-gradient(blue, blue, red, lime -7px)", "linear", [6]pr.Fl{200, -1, 200, 1}, []pr.Fl{0, 0.5, 0.5, 0.5, 0.5, 1}, []parser.RGBA{blue, blue, blue, red, lime, lime})
	layout("repeating-linear-gradient(blue, blue, lime, lime -7px)", "solid", [6]pr.Fl{}, nil, []parser.RGBA{pr.NewColor(0, .5, .5, 1).RGBA})
}

func TestRadialGradient(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	red := pr.NewColor(1, 0, 0, 1).RGBA
	lime := pr.NewColor(0, 1, 0, 1).RGBA
	blue := pr.NewColor(0, 0, 1, 1).RGBA

	// type="radial" positions=[0, 1] colors = [blue, lime], 1
	layout := func(gradientCss string, type_ string, init [6]pr.Fl,
		positions []pr.Fl, colors []parser.RGBA, scaleY pr.Fl) {
		page := renderOnePage(t, "<style>@page { background: "+gradientCss)
		layer := page.Background.Layers[0]
		result := layer.Image.(images.RadialGradient).Layout(400, 300)
		tu.AssertEqual(t, result.ScaleY, scaleY, "scale for "+gradientCss)
		tu.AssertEqual(t, result.Kind, type_, "result[1]")

		if type_ == "radial" {
			centerX, centerY, radius0, radius1 := init[0], init[1], init[2], init[3]
			init = [6]pr.Fl{centerX, centerY / scaleY, radius0, centerX, centerY / scaleY, radius1}
		}
		approxEqualSlice(t, result.GradientKind.Coords[:], init[:], "Data for "+gradientCss)
		tu.AssertEqual(t, result.Positions, positions, "Positions for "+gradientCss)
		tu.AssertEqual(t, len(result.Colors) >= len(colors), true, "colors length")
		for i := range colors {
			color1, color2 := result.Colors[i], colors[i]
			approxEqualColor(t, color1, color2, "color for "+gradientCss)
		}
	}

	layout("radial-gradient(blue)", "solid", [6]pr.Fl{}, nil, []parser.RGBA{blue}, 1)
	layout("repeating-radial-gradient(blue)", "solid", [6]pr.Fl{}, nil, []parser.RGBA{blue}, 1)
	layout("radial-gradient(100px, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 100}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)

	layout("radial-gradient(100px at right 20px bottom 30px, lime, red)", "radial", [6]pr.Fl{380, 270, 0, 100}, []pr.Fl{0, 1}, []parser.RGBA{lime, red}, 1)
	layout("radial-gradient(0 0, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 1e-7}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(1px 0, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 1e7}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1e-14)
	layout("radial-gradient(0 1px, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 1e-7}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1e14)
	layout("repeating-radial-gradient(100px 200px, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 300}, []pr.Fl{0, 1, 1, 2, 2, 3}, []parser.RGBA{blue, lime}, 200./100)
	layout("repeating-radial-gradient(42px, blue -100px, lime 100px)", "radial", [6]pr.Fl{200, 150, 0, 300}, []pr.Fl{-0.5, 0, 0, 1}, []parser.RGBA{pr.NewColor(0, 0.5, 0.5, 1).RGBA, lime, blue, lime}, 1)
	layout("radial-gradient(42px, blue -20px, lime -1px)", "solid", [6]pr.Fl{}, nil, []parser.RGBA{lime}, 1)
	layout("radial-gradient(42px, blue -20px, lime 0)", "solid", [6]pr.Fl{}, nil, []parser.RGBA{lime}, 1)
	layout("radial-gradient(42px, blue -20px, lime 20px)", "radial", [6]pr.Fl{200, 150, 0, 20}, []pr.Fl{0, 1}, []parser.RGBA{pr.NewColor(0, .5, .5, 1).RGBA, lime}, 1)

	layout("radial-gradient(100px 120px, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 100}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 120./100)
	layout("radial-gradient(25% 40%, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 100}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 120./100)

	layout("radial-gradient(circle closest-side, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 150}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle closest-side at 150px 50px, blue, lime)", "radial", [6]pr.Fl{150, 50, 0, 50}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle closest-side at 45px 50px, blue, lime)", "radial", [6]pr.Fl{45, 50, 0, 45}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle closest-side at 420px 50px, blue, lime)", "radial", [6]pr.Fl{420, 50, 0, 20}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle closest-side at 420px 281px, blue, lime)", "radial", [6]pr.Fl{420, 281, 0, 19}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)

	layout("radial-gradient(closest-side, blue 20%, lime)", "radial", [6]pr.Fl{200, 150, 40, 200}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 150./200)
	layout("radial-gradient(closest-side at 300px 20%, blue, lime)", "radial", [6]pr.Fl{300, 60, 0, 100}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 60./100)
	layout("radial-gradient(closest-side at 10% 230px, blue, lime)", "radial", [6]pr.Fl{40, 230, 0, 40}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 70./40)

	layout("radial-gradient(circle farthest-side, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 200}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle farthest-side at 150px 50px, blue, lime)", "radial", [6]pr.Fl{150, 50, 0, 250}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle farthest-side at 45px 50px, blue, lime)", "radial", [6]pr.Fl{45, 50, 0, 355}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle farthest-side at 420px 50px, blue, lime)", "radial", [6]pr.Fl{420, 50, 0, 420}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle farthest-side at 220px 310px, blue, lime)", "radial", [6]pr.Fl{220, 310, 0, 310}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)

	layout("radial-gradient(farthest-side, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 200}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 150./200)
	layout("radial-gradient(farthest-side at 300px 20%, blue, lime)", "radial", [6]pr.Fl{300, 60, 0, 300}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 240./300)
	layout("radial-gradient(farthest-side at 10% 230px, blue, lime)", "radial", [6]pr.Fl{40, 230, 0, 360}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 230./360)

	layout("radial-gradient(circle closest-corner, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 250}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle closest-corner at 340px 80px, blue, lime)", "radial", [6]pr.Fl{340, 80, 0, 100}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle closest-corner at 0 342px, blue, lime)", "radial", [6]pr.Fl{0, 342, 0, 42}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)

	layout("radial-gradient(closest-corner, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 200 * math.Sqrt2}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 150./200)
	layout("radial-gradient(closest-corner at 450px 100px, blue, lime)", "radial", [6]pr.Fl{450, 100, 0, 50 * math.Sqrt2}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 100./50)
	layout("radial-gradient(closest-corner at 40px 210px, blue, lime)", "radial", [6]pr.Fl{40, 210, 0, 40 * math.Sqrt2}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 90./40)

	layout("radial-gradient(circle farthest-corner, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 250}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle farthest-corner at 300px -100px, blue, lime)", "radial", [6]pr.Fl{300, -100, 0, 500}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)
	layout("radial-gradient(circle farthest-corner at 400px 0, blue, lime)", "radial", [6]pr.Fl{400, 0, 0, 500}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 1)

	layout("radial-gradient(farthest-corner, blue, lime)", "radial", [6]pr.Fl{200, 150, 0, 200 * math.Sqrt2}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 150./200)
	layout("radial-gradient(farthest-corner at 450px 100px, blue, lime)", "radial", [6]pr.Fl{450, 100, 0, 450 * math.Sqrt2}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 200./450)
	layout("radial-gradient(farthest-corner at 40px 210px, blue, lime)", "radial", [6]pr.Fl{40, 210, 0, 360 * math.Sqrt2}, []pr.Fl{0, 1}, []parser.RGBA{blue, lime}, 210./360)
}

func TestImageMinMaxWidth(t *testing.T) {
	default_ := map[string]string{
		"min-width": "auto", "max-width": "none", "width": "auto",
		"min-height": "auto", "max-height": "none", "height": "auto",
	}
	for _, data := range []struct {
		props    map[string]string
		divWidth pr.Float
	}{
		{map[string]string{}, 4},
		{map[string]string{"min-width": "10px"}, 10},
		{map[string]string{"max-width": "1px"}, 1},
		{map[string]string{"width": "10px"}, 10},
		{map[string]string{"width": "1px"}, 1},
		{map[string]string{"min-height": "10px"}, 10},
		{map[string]string{"max-height": "1px"}, 1},
		{map[string]string{"height": "10px"}, 10},
		{map[string]string{"height": "1px"}, 1},
		{map[string]string{"min-width": "10px", "min-height": "1px"}, 10},
		{map[string]string{"min-width": "1px", "min-height": "10px"}, 10},
		{map[string]string{"max-width": "10px", "max-height": "1px"}, 1},
		{map[string]string{"max-width": "1px", "max-height": "10px"}, 1},
	} {
		var values []string
		for k, v := range default_ {
			if v2, has := data.props[k]; has {
				v = v2
			}
			values = append(values, fmt.Sprintf("%s: %s", k, v))
		}
		htmlInput := fmt.Sprintf(`
		<style> img { display: block; %s } </style>
		<div style="display: inline-block">
			<img src="pattern.png"><img src="pattern.svg">
		</div>`, strings.Join(values, ";"))
		page := renderOnePage(t, htmlInput)
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		line := body.Box().Children[0]
		div := line.Box().Children[0]
		tu.AssertEqual(t, div.Box().Width, data.divWidth, fmt.Sprintf("div width for %v", data.props))
	}
}

func TestSVGResizing(t *testing.T) {
	input := `
	<style>
		@page { size: 12px }
		body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
		svg { display: block; width: 8px }
	</style>
	<svg viewbox="0 0 4 4">
		<rect width="4" height="4" fill="#00f" />
		<rect width="1" height="1" fill="#f00" />
	</svg>`
	page := renderOnePage(t, input)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	img := body.Box().Children[0]
	if h := img.Box().Height; h != pr.Float(8) {
		t.Fatalf("invalid SVG height: %v", h)
	}
}
