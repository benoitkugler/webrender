// Package matrix provides 2D affine transformations.
package matrix

import (
	"errors"
	"math"

	"github.com/benoitkugler/webrender/utils"
)

type fl = utils.Fl

// Transform encode a (2D) linear transformation (Y = AX + B)
// The encoded transformation is given by :
// 		x_new = a * x + c * y + e
//		y_new = b * x + d * y + f
// which is equivalent to the vector notation
// 	A = | A C | ;  B = 	| E |
//		| B	D |			| F |
type Transform struct {
	a, b, c, d, e, f fl
}

func New(a, b, c, d, e, f fl) Transform {
	return Transform{a: a, b: b, c: c, d: d, e: e, f: f}
}

// Identity returns a new matrix initialized to the identity.
func Identity() Transform {
	return New(1, 0, 0, 1, 0, 0)
}

// Determinant returns the determinant of the matrix, which is
// non zero if and only if the transformation is reversible.
func (t Transform) Determinant() fl {
	return t.a*t.d - t.b*t.c
}

// Copy() method returns a copy of M
func (T Transform) Copy() *Transform {
	return &T
}

func (T Transform) Data() (a, b, c, d, e, f fl) {
	return T.a, T.b, T.c, T.d, T.e, T.f
}

// write t * t_ in out
func mult(t, t_ Transform, out *Transform) {
	out.a = t.a*t_.a + t.c*t_.b
	out.b = t.b*t_.a + t.d*t_.b
	out.c = t.a*t_.c + t.c*t_.d
	out.d = t.b*t_.c + t.d*t_.d
	out.e = t.a*t_.e + t.c*t_.f + t.e
	out.f = t.b*t_.e + t.d*t_.f + t.f
}

// Mul returns the transform T * U,
// which apply U then T.
func Mul(T, U Transform) Transform {
	out := Transform{}
	mult(T, U, &out)
	return out
}

// Mult update T in place with the result of U * T
func (T *Transform) RightMult(U Transform) {
	tmp := *T
	mult(U, tmp, T)
}

// Mult update T in place with the result of T * U
func (T *Transform) Mult(U Transform) {
	tmp := *T
	mult(tmp, U, T)
}

// Invert modify the matrix in place. Return an error
// if the transformation is not bijective.
func (T *Transform) Invert() error {
	det := T.Determinant()
	if det == 0 {
		return errors.New("The transformation is not invertible.")
	}
	T.a, T.d = T.d/det, T.a/det
	T.b = -T.b / det
	T.c = -T.c / det
	T.e = -(T.a*T.e + T.c*T.f)
	T.f = -(T.b*T.e + T.d*T.f)
	return nil
}

// Transforms the point `(x, y)` by this matrix, that is
// compute AX + B
func (T Transform) TransformPoint(x, y fl) (outX, outY fl) {
	outX = T.a*x + T.c*y + T.e
	outY = T.b*x + T.d*y + T.f
	return
}

// Applies a translation by `tx`, `ty`
// to the transformation in this matrix.
//
// The effect of the new transformation is to
// first translate the coordinates by `tx` and `ty`,
// then apply the original transformation to the coordinates.
//
// 	This changes the matrix in-place.
func (T *Transform) Translate(tx, ty fl) {
	T.e, T.f = T.TransformPoint(tx, ty)
}

func Translation(tx, ty fl) Transform {
	return Transform{1, 0, 0, 1, tx, ty}
}

func Scaling(sx, sy fl) Transform {
	return Transform{sx, 0, 0, sy, 0, 0}
}

// Applies scaling by `sx`, `sy`
// to the transformation in this matrix.
//
// The effect of the new transformation is to
// first scale the coordinates by `sx` and `sy`,
// then apply the original transformation to the coordinates.
//
// This changes the matrix in-place.
func (T *Transform) Scale(sx, sy fl) {
	mult(*T, Scaling(sx, sy), T)
}

// Applies a rotation by `radians`
// to the transformation in this matrix.
//
// The effect of the new transformation is to
// first rotate the coordinates by `radians`,
// then apply the original transformation to the coordinates.
//
// This is equivalent to computing T x Rotation(radians)
//
// This changes the matrix in-place.
func (T *Transform) Rotate(radians fl) {
	mult(*T, Rotation(radians), T)
}

// Rotation returns a rotation.
//
// `radians` is the angle of rotation, in radians.
// The direction of rotation is defined such that positive angles
// rotate in the direction from the positive X axis
// toward the positive Y axis.
func Rotation(radians fl) Transform {
	cos, sin := fl(math.Cos(float64(radians))), fl(math.Sin(float64(radians)))
	return Transform{cos, sin, -sin, cos, 0, 0}
}

// Skew returns a skew transformation
func Skew(thetax, thetay fl) Transform {
	b, c := fl(math.Tan(float64(thetax))), fl(math.Tan(float64(thetay)))
	return Transform{1, b, c, 1, 0, 0}
}

// Skew applies a skew transformation
func (T *Transform) Skew(thetax, thetay fl) {
	mult(*T, Skew(thetax, thetay), T)
}
