// Package matrix provides 2D affine transformations.
package matrix

import (
	"errors"
	"math"

	"github.com/benoitkugler/webrender/utils"
)

type fl = utils.Fl

// Transform encode a (2D) linear transformation
//
// The encoded transformation is given by :
//		x_new = a * x + c * y + e
//		y_new = b * x + d * y + f
//
// which is equivalent to the vector notation Y = AX + B, with
// 	A = | a c | ;  B = 	| e |
//		| b	d |			| f |
//
// Transformation may also be viewed as 3D matrices of the form
// T =	| a c e |
//  	| b d f |
//  	| 0 0 1 |
// and P = | x |
//         | y |
//         | 1 |
// where | x_new | = T * P
//       | y_new |
//       |   1   |
type Transform struct {
	A, B, C, D, E, F fl
}

func New(a, b, c, d, e, f fl) Transform {
	return Transform{A: a, B: b, C: c, D: d, E: e, F: f}
}

// Identity returns a new matrix initialized to the identity.
func Identity() Transform {
	return New(1, 0, 0, 1, 0, 0)
}

// Translation returns the translation by (tx, ty).
func Translation(tx, ty fl) Transform {
	return Transform{1, 0, 0, 1, tx, ty}
}

// Scaling returns the scaling by (sx, sy).
func Scaling(sx, sy fl) Transform {
	return Transform{sx, 0, 0, sy, 0, 0}
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

// Determinant returns the determinant of the matrix, which is
// non zero if and only if the transformation is reversible.
func (t Transform) Determinant() fl {
	return t.A*t.D - t.B*t.C
}

// write t1 * t2 in out
func mult(t1, t2 Transform, out *Transform) {
	out.A = t1.A*t2.A + t1.C*t2.B
	out.B = t1.B*t2.A + t1.D*t2.B
	out.C = t1.A*t2.C + t1.C*t2.D
	out.D = t1.B*t2.C + t1.D*t2.D
	out.E = t1.A*t2.E + t1.C*t2.F + t1.E
	out.F = t1.B*t2.E + t1.D*t2.F + t1.F
}

// Mul returns the transform T * U,
// which apply U then T.
func Mul(T, U Transform) Transform {
	out := Transform{}
	mult(T, U, &out)
	return out
}

// Mul3 returns the transform R * S * T,
// which applies T, then S, then R.
func Mul3(R, S, T Transform) Transform {
	out := Transform{}
	mult(S, T, &out)
	mult(R, out, &out)
	return out
}

// LeftMultBy update T in place with the result of U * T
// The resulting transformation apply T first, then U.
func (T *Transform) LeftMultBy(U Transform) { mult(U, *T, T) }

// RightMultBy update T in place with the result of T * U
func (T *Transform) RightMultBy(U Transform) { mult(*T, U, T) }

// Invert modify the matrix in place. Return an error
// if the transformation is not bijective.
func (T *Transform) Invert() error {
	det := T.Determinant()
	if det == 0 {
		return errors.New("transformation is not invertible")
	}
	// if T = AX + Y;  T^-1 = A^-1 ; -A^-1 * B
	T.A, T.D = T.D/det, T.A/det
	T.B = -T.B / det
	T.C = -T.C / det
	e := -(T.A*T.E + T.C*T.F)
	f := -(T.B*T.E + T.D*T.F)
	T.E, T.F = e, f
	return nil
}

// Apply transforms the point `(x, y)` by this matrix, that is
// compute AX + B
func (T Transform) Apply(x, y fl) (outX, outY fl) {
	outX = T.A*x + T.C*y + T.E
	outY = T.B*x + T.D*y + T.F
	return
}

// Applies a translation by `tx`, `ty`
// to the transformation in this matrix.
//
// The effect of the new transformation is to
// first translate the coordinates by `tx` and `ty`,
// then apply the original transformation to the coordinates.
//
// This is equivalent to computing T x Translation(tx, ty)
//
// 	This changes the matrix in-place.
func (T *Transform) Translate(tx, ty fl) {
	T.E += T.A*tx + T.C*ty
	T.F += T.B*tx + T.D*ty
}

// Applies scaling by `sx`, `sy`
// to the transformation in this matrix.
//
// The effect of the new transformation is to
// first scale the coordinates by `sx` and `sy`,
// then apply the original transformation to the coordinates.
//
// This is equivalent to computing T x Scaling(sx, sy).
//
// This changes the matrix in-place.
func (T *Transform) Scale(sx, sy fl) {
	T.A *= sx
	T.B *= sx
	T.C *= sy
	T.D *= sy
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
func (T *Transform) Rotate(radians fl) { T.RightMultBy(Rotation(radians)) }

// Skew applies a skew transformation
//
// The effect of the new transformation is to
// first skew the coordinates,
// then apply the original transformation to the coordinates.
//
// This is equivalent to computing T x Skew(thetax, thetay)
func (T *Transform) Skew(thetax, thetay fl) { T.RightMultBy(Skew(thetax, thetay)) }
