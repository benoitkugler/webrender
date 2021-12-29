package utils

import (
	"math"
	"testing"
)

func equals(x, y Fl) bool {
	return math.Abs(float64(x-y)) < 1e-6
}

func TestModulo(t *testing.T) {
	if v := FloatModulo(4.456, 3); !equals(v, 1.456) {
		t.Errorf("expected 1.456, got %f", v)
	}
	if v := FloatModulo(-2.456, 3); !equals(v, 0.544) {
		t.Errorf("expected 0.544, got %f", v)
	}
	if v := FloatModulo(-8, 5); !equals(v, 2) {
		t.Errorf("expected 2, got %f", v)
	}
	if v := FloatModulo(45, 7); !equals(v, 3) {
		t.Errorf("expected 3, got %f", v)
	}
}
