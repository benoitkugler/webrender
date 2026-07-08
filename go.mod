module github.com/benoitkugler/webrender

go 1.23.0

toolchain go1.24.1

require (
	github.com/benoitkugler/textlayout v0.3.2
	github.com/benoitkugler/textprocessing v0.0.6
	github.com/go-text/typesetting v0.3.5-0.20260506201825-684cf8ff69fd
	golang.org/x/image v0.29.0
	golang.org/x/net v0.42.0
	golang.org/x/text v0.27.0
)

require github.com/benoitkugler/pstokenizer v1.0.1 // indirect

// replace github.com/go-text/typesetting => ../../go-text/typesetting
