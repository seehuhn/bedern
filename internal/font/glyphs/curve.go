// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/outline"
)

// A cubic is one cubic Bézier curve.
type cubic struct{ P0, P1, P2, P3 outline.Point }

// args returns the three points after the start, in the order of
// outline.Path.CurveTo.
func (c cubic) args() (x1, y1, x2, y2, x3, y3 float64) {
	return c.P1.X, c.P1.Y, c.P2.X, c.P2.Y, c.P3.X, c.P3.Y
}

func lerp(a, b outline.Point, t float64) outline.Point {
	return outline.Point{X: a.X + t*(b.X-a.X), Y: a.Y + t*(b.Y-a.Y)}
}

// at returns the point of the curve at the parameter t.
func (c cubic) at(t float64) outline.Point {
	u := 1 - t
	return outline.Point{
		X: u*u*u*c.P0.X + 3*u*u*t*c.P1.X + 3*u*t*t*c.P2.X + t*t*t*c.P3.X,
		Y: u*u*u*c.P0.Y + 3*u*u*t*c.P1.Y + 3*u*t*t*c.P2.Y + t*t*t*c.P3.Y,
	}
}

// tangent returns the unit vector in which the curve runs at the parameter t.
func (c cubic) tangent(t float64) outline.Point {
	u := 1 - t
	dx := 3 * (u*u*(c.P1.X-c.P0.X) + 2*u*t*(c.P2.X-c.P1.X) + t*t*(c.P3.X-c.P2.X))
	dy := 3 * (u*u*(c.P1.Y-c.P0.Y) + 2*u*t*(c.P2.Y-c.P1.Y) + t*t*(c.P3.Y-c.P2.Y))
	n := math.Hypot(dx, dy)
	return outline.Point{X: dx / n, Y: dy / n}
}

// sub returns the part of the curve between the parameters t0 and t1.  It is the
// same curve as a shape, cut by de Casteljau's construction, and its end points
// lie exactly on the original.  If t0 is above t1 the part runs against the
// direction of the curve: it is the part between t1 and t0, reversed.  Parameters
// outside [0, 1] continue the curve.
func (c cubic) sub(t0, t1 float64) cubic {
	switch {
	case t0 > t1:
		return c.sub(t1, t0).reversed()
	case t0 >= 0 && t1 > 0 && t1 <= 1:
		if t1 < 1 {
			c = c.split(t1, true)
			t0 /= t1
		}
		if t0 > 0 {
			c = c.split(t0, false)
		}
		return c
	case t0 < 1:
		// Cut at t0 first, then at the parameter of t1 on what is left.
		return c.split(t0, false).split((t1-t0)/(1-t0), true)
	}
	// t0 = t1 = 1 or both beyond: seen from the other end the part starts below 1.
	return c.reversed().sub(1-t1, 1-t0).reversed()
}

// split returns the part of the curve before t (if head is true) or after it.
func (c cubic) split(t float64, head bool) cubic {
	p01, p12, p23 := lerp(c.P0, c.P1, t), lerp(c.P1, c.P2, t), lerp(c.P2, c.P3, t)
	p012, p123 := lerp(p01, p12, t), lerp(p12, p23, t)
	m := lerp(p012, p123, t)
	if head {
		return cubic{c.P0, p01, p012, m}
	}
	return cubic{m, p123, p23, c.P3}
}

// tAtY returns the parameter at which the curve is at the height y.  The curve
// must run monotonically in y between its end points, and y must lie between
// them.
func (c cubic) tAtY(y float64) float64 {
	lo, hi := 0.0, 1.0
	up := c.P3.Y > c.P0.Y
	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		if (c.at(mid).Y < y) == up {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// fromY returns the part of the curve from the height y to its end, with the
// start exactly at y.  The curve must run monotonically in y, see tAtY.
func (c cubic) fromY(y float64) cubic {
	return c.sub(c.tAtY(y), 1).snapStartY(y)
}

// toY returns the part of the curve from its start to the height y, with the end
// exactly at y; the other half of fromY.
func (c cubic) toY(y float64) cubic {
	return c.sub(0, c.tAtY(y)).snapEndY(y)
}

// snapEndY sets the height of the end of the curve to y.
func (c cubic) snapEndY(y float64) cubic {
	c.P3.Y = y
	return c
}

// snapStartY sets the height of the start of the curve to y.
func (c cubic) snapStartY(y float64) cubic {
	c.P0.Y = y
	return c
}

// curvature returns the signed curvature of the curve at the parameter t, which
// is positive where the curve turns to the left.
func (c cubic) curvature(t float64) float64 {
	u := 1 - t
	d1 := outline.Point{
		X: 3 * (u*u*(c.P1.X-c.P0.X) + 2*u*t*(c.P2.X-c.P1.X) + t*t*(c.P3.X-c.P2.X)),
		Y: 3 * (u*u*(c.P1.Y-c.P0.Y) + 2*u*t*(c.P2.Y-c.P1.Y) + t*t*(c.P3.Y-c.P2.Y)),
	}
	d2 := outline.Point{
		X: 6 * (u*(c.P2.X-2*c.P1.X+c.P0.X) + t*(c.P3.X-2*c.P2.X+c.P1.X)),
		Y: 6 * (u*(c.P2.Y-2*c.P1.Y+c.P0.Y) + t*(c.P3.Y-2*c.P2.Y+c.P1.Y)),
	}
	n := math.Hypot(d1.X, d1.Y)
	return (d1.X*d2.Y - d1.Y*d2.X) / (n * n * n)
}

// tAtX returns the parameter at which the curve is at the abscissa x.  The curve
// must run monotonically in x between its end points, and x must lie between them;
// it is tAtY for a curve which runs across.
func (c cubic) tAtX(x float64) float64 {
	lo, hi := 0.0, 1.0
	right := c.P3.X > c.P0.X
	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		if (c.at(mid).X < x) == right {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// offsetLeftGrowing returns a cubic curve which runs to the left of the curve c at
// a distance dist(t) measured at right angles to c, where dist is a function of
// the parameter, for the part of c from the parameter t0 to the end.  The exact curve,
// c(t) plus the unit normal times dist(t), is not in general a cubic; the result has
// the same ends and the other two control points which fit it best to 64 points of
// the exact curve in the sense of least squares.
func (c cubic) offsetLeftGrowing(dist func(t float64) float64, t0 float64) cubic {
	exact := func(u float64) outline.Point {
		t := t0 + u*(1-t0)
		q, tg := c.at(t), c.tangent(t)
		n := math.Hypot(tg.X, tg.Y)
		d := dist(t)
		return outline.Point{X: q.X - tg.Y/n*d, Y: q.Y + tg.X/n*d}
	}
	p0, p3 := exact(0), exact(1)
	// The curve is p0 (b0 + ...) with the Bernstein polynomials b0 to b3; the
	// unknown control points p1 and p2 solve the normal equations of the residual
	// of the points of the exact curve, in x and in y alike.
	var s11, s12, s22, r1x, r1y, r2x, r2y float64
	for k := 1; k < 64; k++ {
		t := float64(k) / 64
		m := 1 - t
		b0, b1, b2, b3 := m*m*m, 3*m*m*t, 3*m*t*t, t*t*t
		q := exact(t)
		rx := q.X - b0*p0.X - b3*p3.X
		ry := q.Y - b0*p0.Y - b3*p3.Y
		s11 += b1 * b1
		s12 += b1 * b2
		s22 += b2 * b2
		r1x += b1 * rx
		r1y += b1 * ry
		r2x += b2 * rx
		r2y += b2 * ry
	}
	den := s11*s22 - s12*s12
	return cubic{
		P0: p0,
		P1: outline.Point{X: (r1x*s22 - r2x*s12) / den, Y: (r1y*s22 - r2y*s12) / den},
		P2: outline.Point{X: (r2x*s11 - r1x*s12) / den, Y: (r2y*s11 - r1y*s12) / den},
		P3: p3,
	}
}

// crossing returns the parameters ta of a and tb of b at which the two curves
// meet, and ok is false if they do not.  It finds the first crossing along a by
// sampling both curves and refining the first crossing of the two polylines
// with Newton's method.
func crossing(a, b cubic) (ta, tb float64, ok bool) {
	const n = 256
	pa := make([]outline.Point, n+1)
	pb := make([]outline.Point, n+1)
	for i := 0; i <= n; i++ {
		pa[i], pb[i] = a.at(float64(i)/n), b.at(float64(i)/n)
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			// The segments pa[i]pa[i+1] and pb[j]pb[j+1].
			d1x, d1y := pa[i+1].X-pa[i].X, pa[i+1].Y-pa[i].Y
			d2x, d2y := pb[j+1].X-pb[j].X, pb[j+1].Y-pb[j].Y
			den := d1x*d2y - d1y*d2x
			if den == 0 {
				continue
			}
			wx, wy := pb[j].X-pa[i].X, pb[j].Y-pa[i].Y
			u := (wx*d2y - wy*d2x) / den
			v := (wx*d1y - wy*d1x) / den
			if u < 0 || u > 1 || v < 0 || v > 1 {
				continue
			}
			ta, tb = (float64(i)+u)/n, (float64(j)+v)/n
			// Newton's method on a(ta) - b(tb) = 0.
			for k := 0; k < 20; k++ {
				fa, fb := a.at(ta), b.at(tb)
				ex, ey := fa.X-fb.X, fa.Y-fb.Y
				if math.Hypot(ex, ey) < 1e-10 {
					break
				}
				da, db := a.deriv(ta), b.deriv(tb)
				// Solve da * dta - db * dtb = -e.
				det := -da.X*db.Y + da.Y*db.X
				if math.Abs(det) < 1e-12 {
					break
				}
				dta := (-ex*-db.Y - -ey*-db.X) / det
				dtb := (da.X*-ey - da.Y*-ex) / det
				ta, tb = math.Min(1, math.Max(0, ta+dta)), math.Min(1, math.Max(0, tb+dtb))
			}
			return ta, tb, true
		}
	}
	return 0, 0, false
}

// deriv returns the derivative of the curve at the parameter t.
func (c cubic) deriv(t float64) outline.Point {
	u := 1 - t
	return outline.Point{
		X: 3 * (u*u*(c.P1.X-c.P0.X) + 2*u*t*(c.P2.X-c.P1.X) + t*t*(c.P3.X-c.P2.X)),
		Y: 3 * (u*u*(c.P1.Y-c.P0.Y) + 2*u*t*(c.P2.Y-c.P1.Y) + t*t*(c.P3.Y-c.P2.Y)),
	}
}
