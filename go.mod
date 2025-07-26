module github.com/benoitkugler/webrender

go 1.23.0

toolchain go1.24.1

require (
	github.com/benoitkugler/textlayout v0.3.1
	github.com/benoitkugler/textprocessing v0.0.3
	github.com/go-text/typesetting v0.3.0
	golang.org/x/image v0.29.0
	golang.org/x/net v0.42.0
	golang.org/x/text v0.27.0
)

require github.com/benoitkugler/pstokenizer v1.0.1 // indirect

// replace github.com/go-text/typesetting => ../../go-text/typesetting
