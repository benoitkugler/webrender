package matrix

import (
	"math"
	"math/rand"
	"testing"
)

func randT() Transform {
	return New(rand.Float32(), rand.Float32(), rand.Float32(), rand.Float32(), rand.Float32(), rand.Float32())
}

func TestDeterminant(t *testing.T) {
	if det := Identity().Determinant(); det != 1 {
		t.Fatalf("unexpected derterminant: %f", det)
	}

	if det := Rotation(20).Determinant(); det != 1 {
		t.Fatalf("unexpected derterminant: %f", det)
	}

	if det := Translation(2, 2).Determinant(); det != 1 {
		t.Fatalf("unexpected derterminant: %f", det)
	}
}

func TestComposition(t *testing.T) {
	sc := Scaling(2, 3)
	rt := Rotation(30)
	tr := Translation(0.5, 1.5)

	// the composition of the three transformation is equal to
	c, s := fl(math.Cos(30)), fl(math.Sin(30))
	res1 := New(2*c, 3*s, -2*s, 3*c, 0.5, 1.5)

	// apply rt, then sc, then tr
	res2 := Mul(tr, Mul(sc, rt))

	// same
	rt.LeftMultBy(sc)
	rt.LeftMultBy(tr)
	res3 := rt

	if res1 != res2 || res1 != res3 {
		t.Fatalf("inconsistent results: %v %v %v", res1, res2, res3)
	}
}

func TestScale(t *testing.T) {
	for range [10]int{} {
		mat1 := randT()

		exp := Mul(mat1, Scaling(2, 3))

		mat1.Scale(2, 3)

		if mat1 != exp {
			t.Fatalf("unexpected Scale: %v", mat1)
		}
	}
}

func TestTranslate(t *testing.T) {
	for range [10]int{} {
		mat1 := randT()

		exp := Mul(mat1, Translation(2, 3))

		mat1.Translate(2, 3)

		if mat1 != exp {
			t.Fatalf("unexpected Scale: %v", mat1)
		}
	}
}

func TestSkew(t *testing.T) {
	res1 := Mul(Identity(), Skew(1, 2))

	res2 := Identity()
	res2.Skew(1, 2)

	if res1 != res2 {
		t.Fatalf("inconsitent Skew %v != %v", res1, res2)
	}
}

func TestRotate(t *testing.T) {
	res1 := Mul(Identity(), Rotation(10))

	res2 := Identity()
	res2.Rotate(10)

	if res1 != res2 {
		t.Fatalf("inconsitent Rotate %v != %v", res1, res2)
	}
}

func TestInvert(t *testing.T) {
	m1 := Translation(2, 2)
	m2 := Rotation(math.Pi / 2)

	prod1 := Mul(m1, m2)
	inv1 := prod1
	if err := inv1.Invert(); err != nil {
		t.Fatal(err)
	}

	if p := Mul(inv1, prod1); p != Identity() {
		t.Fatalf("%v %v %v", inv1, prod1, p)
	}

	if err := m1.Invert(); err != nil {
		t.Fatal(err)
	}
	if m1 != Translation(-2, -2) {
		t.Fatalf("%v", m1)
	}

	if err := m2.Invert(); err != nil {
		t.Fatal(err)
	}
	if m2 != Rotation(-math.Pi/2) {
		t.Fatalf("%v", m2)
	}

	// (AB)^-1 = B^-1 x A^-1
	inv2 := Mul(m2, m1)

	if inv1 != inv2 {
		t.Fatalf("%v %v", inv1, inv2)
	}
}

func TestInvertError(t *testing.T) {
	m := Scaling(1, 0)
	if m.Invert() == nil {
		t.Fatal("expected error on non invertible matrix")
	}
}

func TestApply(t *testing.T) {
	m := Rotation(math.Pi)
	if x, y := m.Apply(1, 1); math.Hypot(float64(x+1), float64(y+1)) > 1e-4 {
		t.Fatalf("%f %f != -1, -1", x, y)
	}
}
