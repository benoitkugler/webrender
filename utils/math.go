package utils

import (
	"math"
)

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func MinF(x, y Fl) Fl {
	if x < y {
		return x
	}
	return y
}

func MaxF(x, y Fl) Fl {
	if x > y {
		return x
	}
	return y
}

type Fl = float32

func Maxs(values ...Fl) Fl {
	max := values[0]
	for _, w := range values {
		if w > max {
			max = w
		}
	}
	return max
}

func Mins(values ...Fl) Fl {
	min := values[0]
	for _, w := range values {
		if w < min {
			min = w
		}
	}
	return min
}

func modLikePython(d, m int) int {
	var res int = d % m
	if (res < 0 && m > 0) || (res > 0 && m < 0) {
		return res + m
	}
	return res
}

func Floor(x Fl) Fl {
	return Fl(math.Floor(float64(x)))
}

// FloatModulo implements Python modulo for float numbers, like
//	4.456 % 3
func FloatModulo(x Fl, i int) Fl {
	x2 := Floor(x)
	diff := x - x2
	return Fl(modLikePython(int(x2), i)) + diff
}

// Round rounds f with 6 digits precision
func Round(f Fl) Fl {
	n := math.Pow10(6)
	return Fl(math.Round(float64(f)*n) / n)
}
