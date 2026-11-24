// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"math"

	"seehuhn.de/go/bedern/internal/outline"
)

// Point is a point in font units.
type Point = outline.Point

func pt(x, y float64) Point            { return Point{X: x, Y: y} }
func add(a, b Point) Point             { return Point{X: a.X + b.X, Y: a.Y + b.Y} }
func sub(a, b Point) Point             { return Point{X: a.X - b.X, Y: a.Y - b.Y} }
func mul(a Point, s float64) Point     { return Point{X: a.X * s, Y: a.Y * s} }
func dot(a, b Point) float64           { return a.X*b.X + a.Y*b.Y }
func cross(a, b Point) float64         { return a.X*b.Y - a.Y*b.X }
func norm(a Point) float64             { return math.Hypot(a.X, a.Y) }
func dist(a, b Point) float64          { return norm(sub(a, b)) }
func lerp(a, b Point, t float64) Point { return add(a, mul(sub(b, a), t)) }

// unit returns a scaled to length 1, or the zero vector.
func unit(a Point) Point {
	l := norm(a)
	if l == 0 {
		return a
	}
	return mul(a, 1/l)
}

// rotate returns v turned by a radians anticlockwise.
func rotate(v Point, a float64) Point {
	s, c := math.Sincos(a)
	return pt(c*v.X-s*v.Y, s*v.X+c*v.Y)
}

// A cubic is a cubic Bézier curve by its four control points.
type cubic [4]Point

// at returns the point of the curve at t.
func (c cubic) at(t float64) Point {
	u := 1 - t
	return add(add(mul(c[0], u*u*u), mul(c[1], 3*u*u*t)), add(mul(c[2], 3*u*t*t), mul(c[3], t*t*t)))
}

// deriv returns the derivative of the curve at t.
func (c cubic) deriv(t float64) Point {
	u := 1 - t
	return add(add(mul(sub(c[1], c[0]), 3*u*u), mul(sub(c[2], c[1]), 6*u*t)), mul(sub(c[3], c[2]), 3*t*t))
}

// split returns the two halves of the curve at t.
func (c cubic) split(t float64) (cubic, cubic) {
	a := lerp(c[0], c[1], t)
	b := lerp(c[1], c[2], t)
	d := lerp(c[2], c[3], t)
	ab := lerp(a, b, t)
	bd := lerp(b, d, t)
	m := lerp(ab, bd, t)
	return cubic{c[0], a, ab, m}, cubic{m, bd, d, c[3]}
}

// extremes returns the parameters strictly inside the curve, sorted, at which
// it runs horizontally or vertically.
func (c cubic) extremes() []float64 {
	var ts []float64
	for axis := 0; axis < 2; axis++ {
		get := func(q Point) float64 {
			if axis == 0 {
				return q.X
			}
			return q.Y
		}
		for _, t := range outline.CubicExtrema(get(c[0]), get(c[1]), get(c[2]), get(c[3])) {
			if t > 1e-3 && t < 1-1e-3 {
				ts = append(ts, t)
			}
		}
	}
	for i := 1; i < len(ts); i++ {
		for j := i; j > 0 && ts[j] < ts[j-1]; j-- {
			ts[j], ts[j-1] = ts[j-1], ts[j]
		}
	}
	return ts
}

// length returns the arc length of the curve.
func (c cubic) length() float64 {
	const n = 32
	var l float64
	prev := c[0]
	for i := 1; i <= n; i++ {
		q := c.at(float64(i) / n)
		l += dist(prev, q)
		prev = q
	}
	return l
}

// fit returns the cubic from a[0] to b[3] with the directions of a at its
// start and of b at its end which lies closest, in the least squares sense,
// to the two curves a and b, which meet at a smooth node, and the largest
// distance between it and them.  The result is false if no such curve with
// positive handles exists.
func fit(a, b cubic) (cubic, float64, bool) {
	p0, p3 := a[0], b[3]
	t0, t3 := unit(sub(a[1], a[0])), unit(sub(b[3], b[2]))
	if norm(t0) == 0 || norm(t3) == 0 {
		return cubic{}, 0, false
	}
	// Samples of the two curves.
	const per = 16
	var qs []Point
	for ci, c := range []cubic{a, b} {
		for i := 0; i <= per; i++ {
			if ci == 1 && i == 0 {
				continue
			}
			qs = append(qs, c.at(float64(i)/per))
		}
	}
	best, _, ok := fitPoints(p0, t0, p3, t3, qs)
	if !ok {
		return cubic{}, 0, false
	}
	var dev float64
	for _, c := range []cubic{a, b} {
		for i := 0; i <= 2*per; i++ {
			dev = math.Max(dev, distToCubic(best, c.at(float64(i)/(2*per))))
		}
	}
	return best, dev, true
}

// fitPoints returns the cubic from p0 in the direction t0 to p3 in the
// direction t3 which lies closest, in the least squares sense, to the points
// qs, which run from p0 to p3, and the largest distance from them to it.
func fitPoints(p0, t0, p3, t3 Point, qs []Point) (cubic, float64, bool) {
	us := make([]float64, len(qs))
	for i := 1; i < len(qs); i++ {
		us[i] = us[i-1] + dist(qs[i-1], qs[i])
	}
	total := us[len(us)-1]
	if total == 0 {
		return cubic{}, 0, false
	}
	for i := range us {
		us[i] /= total
	}
	var best cubic
	for iter := 0; iter < 6; iter++ {
		// The curve p0 + 3u²t... is linear in the handle lengths h0 and h3:
		// B(u) = p0 B0 + (p0 + h0 t0) B1 + (p3 - h3 t3) B2 + p3 B3.
		var m00, m01, m11, r0, r1 float64
		for i, u := range us {
			v := 1 - u
			b0, b1, b2, b3 := v*v*v, 3*v*v*u, 3*v*u*u, u*u*u
			base := add(mul(p0, b0+b1), mul(p3, b2+b3))
			e := sub(qs[i], base)
			g0, g1 := mul(t0, b1), mul(t3, -b2)
			m00 += dot(g0, g0)
			m01 += dot(g0, g1)
			m11 += dot(g1, g1)
			r0 += dot(g0, e)
			r1 += dot(g1, e)
		}
		det := m00*m11 - m01*m01
		if math.Abs(det) < 1e-12 {
			return cubic{}, 0, false
		}
		h0 := (r0*m11 - r1*m01) / det
		h3 := (m00*r1 - m01*r0) / det
		if h0 <= 0 || h3 <= 0 {
			return cubic{}, 0, false
		}
		best = cubic{p0, add(p0, mul(t0, h0)), sub(p3, mul(t3, h3)), p3}
		// Move each parameter to the nearest point of the curve.
		for i := range us {
			u := us[i]
			for k := 0; k < 3; k++ {
				d := best.deriv(u)
				e := sub(best.at(u), qs[i])
				dd := dot(d, d)
				if dd == 0 {
					break
				}
				u -= dot(e, d) / dd
				u = math.Max(0, math.Min(1, u))
			}
			us[i] = u
		}
	}
	var dev float64
	for _, q := range qs {
		dev = math.Max(dev, distToCubic(best, q))
	}
	return best, dev, true
}

// distToCubic returns the distance from q to the curve c.
func distToCubic(c cubic, q Point) float64 {
	const n = 64
	bt, bd := 0.0, math.Inf(1)
	for i := 0; i <= n; i++ {
		t := float64(i) / n
		if d := dist(c.at(t), q); d < bd {
			bt, bd = t, d
		}
	}
	for k := 0; k < 4; k++ {
		d := c.deriv(bt)
		e := sub(c.at(bt), q)
		dd := dot(d, d)
		if dd == 0 {
			break
		}
		bt = math.Max(0, math.Min(1, bt-dot(e, d)/dd))
	}
	return dist(c.at(bt), q)
}
