module github.com/benoitkugler/webrender

go 1.16

require (
	github.com/benoitkugler/textlayout v0.0.10
	github.com/benoitkugler/textprocessing v0.0.0-20220428082259-6af2123ac267 // indirect
	golang.org/x/net v0.0.0-20211216030914-fe4d6282115f
	golang.org/x/text v0.3.7
)

replace github.com/benoitkugler/textlayout => ../textlayout

replace github.com/benoitkugler/textprocessing => ../textprocessing
