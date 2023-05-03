module github.com/benoitkugler/webrender

go 1.16

require (
	github.com/benoitkugler/pstokenizer v1.0.1 // indirect
	github.com/benoitkugler/textlayout v0.3.0
	github.com/benoitkugler/textprocessing v0.0.0-20220428082259-6af2123ac267
	golang.org/x/image v0.0.0-20220413100746-70e8d0d3baa9 // indirect
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4
	golang.org/x/text v0.3.7
)

replace github.com/benoitkugler/textlayout => ../textlayout

replace github.com/benoitkugler/textprocessing => ../textprocessing
