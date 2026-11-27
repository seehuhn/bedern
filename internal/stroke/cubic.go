// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"math"

	"seehuhn.de/go/bedern/internal/outline"
	gpath "seehuhn.de/go/geom/path"
	"seehuhn.de/go/geom/vec"
)

// Point is a point in font units.
type Point = outline.Point

// A Cubic is a cubic Bézier curve.  A curve whose control points lie on its
// chord is a straight line.
type Cubic [4]Point

func add(a, b Point) Point             { return Point{X: a.X + b.X, Y: a.Y + b.Y} }
func sub(a, b Point) Point             { return Point{X: a.X - b.X, Y: a.Y - b.Y} }
func mul(a Point, s float64) Point     { return Point{X: a.X * s, Y: a.Y * s} }
func dot(a, b Point) float64           { return a.X*b.X + a.Y*b.Y }
func cross(a, b Point) float64         { return a.X*b.Y - a.Y*b.X }
func norm(a Point) float64             { return math.Hypot(a.X, a.Y) }
func dist(a, b Point) float64          { return norm(sub(a, b)) }
func rot90(a Point) Point              { return Point{X: -a.Y, Y: a.X} }
func lerp(a, b Point, t float64) Point { return add(a, mul(sub(b, a), t)) }
func unit(a Point) Point {
	n := norm(a)
	if n == 0 {
		return Point{}
	}
	return mul(a, 1/n)
}

// dirOf returns the unit vector of the direction a, in degrees.
func dirOf(a float64) Point {
	r := a * math.Pi / 180
	return Point{X: math.Cos(r), Y: math.Sin(r)}
}

// angleOf returns the direction of v in degrees.
func angleOf(v Point) float64 { return math.Atan2(v.Y, v.X) * 180 / math.Pi }

// At returns the point of the curve at t; t outside [0, 1] extrapolates.
func (c Cubic) At(t float64) Point {
	u := 1 - t
	a, b, cc, d := u*u*u, 3*u*u*t, 3*u*t*t, t*t*t
	return Point{
		X: a*c[0].X + b*c[1].X + cc*c[2].X + d*c[3].X,
		Y: a*c[0].Y + b*c[1].Y + cc*c[2].Y + d*c[3].Y,
	}
}

// D returns the first derivative of the curve at t.
func (c Cubic) D(t float64) Point {
	u := 1 - t
	return add(add(mul(sub(c[1], c[0]), 3*u*u), mul(sub(c[2], c[1]), 6*u*t)), mul(sub(c[3], c[2]), 3*t*t))
}

// DD returns the second derivative of the curve at t.
func (c Cubic) DD(t float64) Point {
	u := 1 - t
	a := add(sub(c[2], mul(c[1], 2)), c[0])
	b := add(sub(c[3], mul(c[2], 2)), c[1])
	return add(mul(a, 6*u), mul(b, 6*t))
}

// Curvature returns the signed curvature at t, positive where the curve turns
// to the left.
func (c Cubic) Curvature(t float64) float64 {
	d, dd := c.D(t), c.DD(t)
	n := norm(d)
	if n == 0 {
		return 0
	}
	return cross(d, dd) / (n * n * n)
}

// Tangent returns the unit tangent at t.  At an end whose handle is on the
// node it is the direction to the other control point.
func (c Cubic) Tangent(t float64) Point {
	d := c.D(t)
	if norm(d) < 1e-12 {
		switch {
		case t <= 0:
			d = sub(c[2], c[0])
		case t >= 1:
			d = sub(c[3], c[1])
		}
	}
	return unit(d)
}

// Split returns the two parts of the curve before and after t, by de
// Casteljau's construction; t outside [0, 1] extends the first part.
func (c Cubic) Split(t float64) (Cubic, Cubic) {
	a := lerp(c[0], c[1], t)
	b := lerp(c[1], c[2], t)
	d := lerp(c[2], c[3], t)
	e := lerp(a, b, t)
	f := lerp(b, d, t)
	m := lerp(e, f, t)
	return Cubic{c[0], a, e, m}, Cubic{m, f, d, c[3]}
}

// Sub returns the part of the curve between t0 and t1 (t0 < t1); the
// parameters may lie outside [0, 1], which extends the curve.
func (c Cubic) Sub(t0, t1 float64) Cubic {
	if math.Abs(t1) < 1e-12 {
		return c.Reversed().Sub(1-t1, 1-t0).Reversed()
	}
	head, _ := c.Split(t1)
	_, r := head.Split(t0 / t1)
	return r
}

// Reversed returns the curve walked from its end to its start.
func (c Cubic) Reversed() Cubic { return Cubic{c[3], c[2], c[1], c[0]} }

// Length returns the arc length of the curve, from 64 straight pieces.
func (c Cubic) Length() float64 { return c.lengthTo(1) }

func (c Cubic) lengthTo(t1 float64) float64 {
	const n = 64
	var l float64
	prev := c[0]
	for i := 1; i <= n; i++ {
		q := c.At(t1 * float64(i) / n)
		l += dist(q, prev)
		prev = q
	}
	return l
}

// IsStraight reports whether the control points lie on the chord.
func (c Cubic) IsStraight() bool {
	d := sub(c[3], c[0])
	l := norm(d)
	if l == 0 {
		return false
	}
	for _, q := range []Point{c[1], c[2]} {
		if math.Abs(cross(d, sub(q, c[0])))/l > 1e-7 {
			return false
		}
		if s := dot(d, sub(q, c[0])) / (l * l); s < -1e-9 || s > 1+1e-9 {
			return false
		}
	}
	return true
}

// Extrema returns the parameters in (0, 1) at which a coordinate of the curve
// is extreme, from the library.
func (c Cubic) Extrema() []float64 {
	g := gpath.Cubic{
		vec.Vec2{X: c[0].X, Y: c[0].Y}, vec.Vec2{X: c[1].X, Y: c[1].Y},
		vec.Vec2{X: c[2].X, Y: c[2].Y}, vec.Vec2{X: c[3].X, Y: c[3].Y},
	}
	return g.Extrema()
}

// doublesBack reports whether the curve from p0 in the direction t0 with the
// handle a to p3 = p0 + d in the direction t3 with the handle b runs back
// along the chord although both ends run forward along it.  With nearly
// straight ends the curvatures of the edge of a stem are tiny and allow
// handles which overlap along the chord: the curve then runs down, back up
// and down again on nearly one line, invisible, but a fold of the edge.
func doublesBack(t0, t3, d Point, a, b float64) bool {
	l := norm(d)
	u := mul(d, 1/l)
	c0, c3 := dot(t0, u), dot(t3, u)
	if c0 <= 0 || c3 <= 0 {
		return false
	}
	// The derivative of the projection onto the chord is a quadratic with
	// the Bernstein coefficients d0, d1, d2 (times 3).
	p1, p2 := a*c0, l-b*c3
	d0, d1, d2 := p1, p2-p1, l-p2
	return d1 < 0 && d1*d1 > d0*d2
}

// hermite returns the curve from p0 in the direction t0 with the curvature k0
// to p3 in the direction t3 with the curvature k3 (both unit directions,
// curvatures signed, positive to the left), and whether it has these
// curvatures.  The handle lengths a and b solve the two equations
//
//	3/2 k0 a^2 = cross(t0, d) - b cross(t0, t3)
//	3/2 k3 b^2 = cross(d, t3) - a cross(t0, t3)
//
// with d = p3 - p0.  Of the positive solutions whose curve does not run back
// along the chord (see doublesBack), the one nearest to the handles of a
// circular arc with the same end directions is taken.  Where there is
// none, the curve has those handles of the arc, is only tangent-continuous at
// its ends, and ok is false.
func hermite(p0, t0 Point, k0 float64, p3, t3 Point, k3 float64) (Cubic, bool) {
	d := sub(p3, p0)
	l := norm(d)
	mk := func(a, b float64) Cubic {
		return Cubic{p0, add(p0, mul(t0, a)), sub(p3, mul(t3, b)), p3}
	}
	if l < 1e-9 {
		return Cubic{p0, p0, p3, p3}, false
	}
	// The handles of the arc: 4/3 tan(theta/4) r on a chord 2 r sin(theta/2),
	// with theta the turn from t0 to t3.
	theta := math.Abs(math.Atan2(cross(t0, t3), dot(t0, t3)))
	h0 := l / 3
	if theta > 1e-6 {
		h0 = l * 2 / 3 * math.Tan(theta/4) / math.Sin(theta/2)
	}
	c := cross(t0, t3)
	A := cross(t0, d)
	B := cross(d, t3)
	if math.Abs(c) < 1e-9*l {
		// Parallel end directions: the two equations are separate.
		a, b := h0, h0
		okA, okB := false, false
		switch {
		case math.Abs(k0) > 1e-12 && A/k0 > 0:
			a, okA = math.Sqrt(A/(1.5*k0)), true
		case math.Abs(k0) <= 1e-12 && math.Abs(A) < 1e-9*l:
			okA = true
		}
		switch {
		case math.Abs(k3) > 1e-12 && B/k3 > 0:
			b, okB = math.Sqrt(B/(1.5*k3)), true
		case math.Abs(k3) <= 1e-12 && math.Abs(B) < 1e-9*l:
			okB = true
		}
		if doublesBack(t0, t3, d, a, b) {
			return mk(h0, h0), false
		}
		return mk(a, b), okA && okB
	}
	bOf := func(a float64) float64 { return (A - 1.5*k0*a*a) / c }
	f := func(a float64) float64 { b := bOf(a); return 1.5*k3*b*b - B + a*c }
	const n = 600
	hi := 2 * l
	best, bestA, bestB := math.Inf(1), 0.0, 0.0
	prevA, prevF := 0.0, f(0)
	prevF2 := prevF
	for i := 1; i <= n; i++ {
		a := hi * float64(i) / n
		fa := f(a)
		if (fa > 0) != (prevF > 0) || fa == 0 {
			lo, up, flo := prevA, a, prevF
			for j := 0; j < 80; j++ {
				m := (lo + up) / 2
				if fm := f(m); (fm > 0) == (flo > 0) {
					lo, flo = m, fm
				} else {
					up = m
				}
			}
			r := (lo + up) / 2
			if b := bOf(r); r > 1e-6 && b > 1e-6 && b < hi && !doublesBack(t0, t3, d, r, b) {
				if e := math.Abs(r-h0) + math.Abs(b-h0); e < best {
					best, bestA, bestB = e, r, b
				}
			}
		}
		// A double root, where f touches zero without changing sign.
		if i >= 2 {
			am := hi * float64(i-1) / n
			if fm := f(am); math.Abs(fm) <= math.Abs(prevF2) && math.Abs(fm) <= math.Abs(fa) && math.Abs(fm) < 1e-2*l {
				lo, up := hi*float64(i-2)/n, a
				for j := 0; j < 100; j++ {
					m1, m2 := lo+(up-lo)/3, up-(up-lo)/3
					if math.Abs(f(m1)) < math.Abs(f(m2)) {
						up = m2
					} else {
						lo = m1
					}
				}
				r := (lo + up) / 2
				if b := bOf(r); math.Abs(f(r)) < 1e-6*l && r > 1e-6 && b > 1e-6 && b < hi && !doublesBack(t0, t3, d, r, b) {
					if e := math.Abs(r-h0) + math.Abs(b-h0); e < best {
						best, bestA, bestB = e, r, b
					}
				}
			}
		}
		prevF2 = prevF
		prevA, prevF = a, fa
	}
	// Newton's method on both equations, from the solution of the equations
	// without the coupling term and from the handles of the arc, finds the
	// roots which the scan misses where cross(t0, t3) is small.
	newton := func(a, b float64) {
		for j := 0; j < 60; j++ {
			f1 := 1.5*k0*a*a - A + b*c
			f2 := 1.5*k3*b*b - B + a*c
			j11, j12, j21, j22 := 3*k0*a, c, c, 3*k3*b
			det := j11*j22 - j12*j21
			if det == 0 || math.IsNaN(det) {
				return
			}
			a -= (f1*j22 - f2*j12) / det
			b -= (j11*f2 - j21*f1) / det
		}
		f1 := 1.5*k0*a*a - A + b*c
		f2 := 1.5*k3*b*b - B + a*c
		if math.Abs(f1) < 1e-7*l && math.Abs(f2) < 1e-7*l && a > 1e-6 && b > 1e-6 && a < hi && b < hi && !doublesBack(t0, t3, d, a, b) {
			if e := math.Abs(a-h0) + math.Abs(b-h0); e < best {
				best, bestA, bestB = e, a, b
			}
		}
	}
	newton(h0, h0)
	if A/k0 > 0 && B/k3 > 0 {
		newton(math.Sqrt(A/(1.5*k0)), math.Sqrt(B/(1.5*k3)))
	}
	if math.IsInf(best, 1) {
		return mk(h0, h0), false
	}
	return mk(bestA, bestB), true
}
