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
//
//	4.456 % 3
func FloatModulo(x Fl, i int) Fl {
	x2 := Floor(x)
	diff := x - x2
	return Fl(modLikePython(int(x2), i)) + diff
}

// RoundPrec rounds f with n digits precision
func RoundPrec(f Fl, n int) Fl {
	n10 := math.Pow10(n)
	return Fl(math.Round(float64(f)*n10) / n10)
}

// Round rounds f with 6 digits precision
func Round(f Fl) Fl {
	return RoundPrec(f, 6)
}

// Hypot returns SQRT(a^2 + b^2)
func Hypot(a, b Fl) Fl {
	return Fl(math.Hypot(float64(a), float64(b)))
}

func Abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
