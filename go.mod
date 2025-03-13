module github.com/benoitkugler/webrender

go 1.19

require (
	github.com/benoitkugler/textlayout v0.3.1
	github.com/benoitkugler/textprocessing v0.0.3
	github.com/go-text/typesetting v0.2.1
	golang.org/x/image v0.23.0
	golang.org/x/net v0.36.0
	golang.org/x/text v0.22.0
)

require github.com/benoitkugler/pstokenizer v1.0.1 // indirect

// replace github.com/go-text/typesetting => ../../go-text/typesetting
