package keywords

// Keyword efficiently stores CSS keywords
type Keyword uint8

const (
	_ Keyword = iota
	Auto
	Baseline
	Center
	End
	First
	FlexEnd
	FlexStart
	Last
	Left
	Legacy
	Normal
	Right
	Safe
	SelfEnd
	SelfStart
	SpaceAround
	SpaceBetween
	SpaceEvenly
	Start
	Stretch
	Unsafe
)

func NewKeyword(s string) Keyword {
	switch s {
	case "auto":
		return Auto
	case "baseline":
		return Baseline
	case "center":
		return Center
	case "end":
		return End
	case "first":
		return First
	case "flex-end":
		return FlexEnd
	case "flex-start":
		return FlexStart
	case "last":
		return Last
	case "left":
		return Left
	case "legacy":
		return Legacy
	case "normal":
		return Normal
	case "right":
		return Right
	case "safe":
		return Safe
	case "self-end":
		return SelfEnd
	case "self-start":
		return SelfStart
	case "space-around":
		return SpaceAround
	case "space-between":
		return SpaceBetween
	case "space-evenly":
		return SpaceEvenly
	case "start":
		return Start
	case "stretch":
		return Stretch
	case "unsafe":
		return Unsafe
	}
	return 0
}
