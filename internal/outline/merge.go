// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package outline

import "math"

// A mergeSample is a point of the two curves which a merged curve replaces,
// with the signed curvature there and the parameter of the merged curve
// nearest to it.
type mergeSample struct {
	q Point
	k float64
	u float64
}

// curvatureAt returns the signed curvature of the cubic c at t.
func curvatureAt(c [4]Point, t float64) float64 {
	d1, d2 := bezD(c, t), bezDD(c, t)
	l := math.Hypot(d1.X, d1.Y)
	if l == 0 {
		return 0
	}
	return (d1.X*d2.Y - d1.Y*d2.X) / (l * l * l)
}

// mergeCubics returns the cubic curve from the start of a to the end of b,
// which a and b continue, with the directions of a at its start and of b at
// its end, and the largest distance between it and the two.  The curve is
// the one, among those with the curvatures of a and b at its ends and the
// one closest to a and b in the least squares sense, which lies closest to
// them.  The result is false if none of them fits (mergeFits).
func mergeCubics(a, b [4]Point) ([4]Point, float64, bool) {
	return mergeCubicsWithin(a, b, MergeTolerance)
}

// mergeCubicsWithin is mergeCubics with the largest distance tol in place of
// MergeTolerance.
func mergeCubicsWithin(a, b [4]Point, tol float64) ([4]Point, float64, bool) {
	return mergeCubicsLoose(a, b, tol, false)
}

// mergeCubicsLoose is mergeCubicsWithin; with loose set the curvature of the
// merged curve is not bounded, only its distance from the two curves and its
// overshoot.
func mergeCubicsLoose(a, b [4]Point, tol float64, loose bool) ([4]Point, float64, bool) {
	p0, p3 := a[0], b[3]
	t0 := unitOf(bezD(a, 0))
	t3 := unitOf(bezD(b, 1))
	if a[1] == a[0] && a[2] == a[0] || b[2] == b[3] && b[1] == b[3] {
		return [4]Point{}, 0, false
	}

	// Samples of the two curves, with their share of the length.
	const per = 24
	var ss []mergeSample
	for ci, c := range [][4]Point{a, b} {
		for i := 0; i <= per; i++ {
			if ci == 1 && i == 0 {
				continue
			}
			t := float64(i) / per
			ss = append(ss, mergeSample{q: bez(c, t), k: curvatureAt(c, t)})
		}
	}
	var total float64
	for i := 1; i < len(ss); i++ {
		total += math.Hypot(ss[i].q.X-ss[i-1].q.X, ss[i].q.Y-ss[i-1].q.Y)
		ss[i].u = total
	}
	if total == 0 {
		return [4]Point{}, 0, false
	}
	for i := range ss {
		ss[i].u /= total
	}
	chordU := make([]float64, len(ss))
	for i := range ss {
		chordU[i] = ss[i].u
	}

	curve := func(al, be float64) [4]Point {
		return [4]Point{p0, {p0.X + al*t0.X, p0.Y + al*t0.Y}, {p3.X - be*t3.X, p3.Y - be*t3.Y}, p3}
	}
	project := func(m [4]Point) {
		for i := 1; i < len(ss)-1; i++ {
			u := ss[i].u
			for k := 0; k < 4; k++ {
				q, d1, d2 := bez(m, u), bezD(m, u), bezDD(m, u)
				ex, ey := q.X-ss[i].q.X, q.Y-ss[i].q.Y
				den := d1.X*d1.X + d1.Y*d1.Y + ex*d2.X + ey*d2.Y
				if den <= 0 {
					break
				}
				u = math.Max(0, math.Min(1, u-(ex*d1.X+ey*d1.Y)/den))
			}
			ss[i].u = u
		}
	}

	var cands [][4]Point

	// The curve closest in the least squares sense.
	lsq := func() ([4]Point, bool) {
		var m [4]Point
		for iter := 0; iter < 8; iter++ {
			var a11, a12, a22, r1, r2 float64
			for _, s := range ss {
				u := s.u
				w := 1 - u
				b0, b1, b2, b3 := w*w*w, 3*w*w*u, 3*w*u*u, u*u*u
				rx := s.q.X - (b0+b1)*p0.X - (b2+b3)*p3.X
				ry := s.q.Y - (b0+b1)*p0.Y - (b2+b3)*p3.Y
				a11 += b1 * b1
				a12 -= b1 * b2 * (t0.X*t3.X + t0.Y*t3.Y)
				a22 += b2 * b2
				r1 += b1 * (t0.X*rx + t0.Y*ry)
				r2 -= b2 * (t3.X*rx + t3.Y*ry)
			}
			det := a11*a22 - a12*a12
			if det == 0 {
				return m, false
			}
			al, be := (r1*a22-r2*a12)/det, (a11*r2-a12*r1)/det
			if !(al > 0 && be > 0) || al > total || be > total {
				return m, false
			}
			m = curve(al, be)
			project(m)
		}
		return m, true
	}
	if m, ok := lsq(); ok {
		cands = append(cands, m)
	}

	// The curves with the curvatures of a and b at the ends, from the roots
	// of 3/2 k0 al^2 = t0 x D - be t0 x t3 and 3/2 k3 be^2 = D x t3 - al t0 x t3.
	k0, k3 := ss[0].k, ss[len(ss)-1].k
	cr := func(u, v Point) float64 { return u.X*v.Y - u.Y*v.X }
	d := sub(p3, p0)
	s, c1, c2 := cr(t0, t3), cr(t0, d), cr(d, t3)
	if math.Abs(k3) > 1e-6 {
		beOf := func(al float64) float64 {
			v := 2.0 / 3 * (c2 - al*s) / k3
			if !(v > 0) {
				return math.NaN()
			}
			return math.Sqrt(v)
		}
		f := func(al float64) float64 { return 1.5*k0*al*al - c1 + beOf(al)*s }
		const n = 200
		for i := 0; i < n; i++ {
			lo, hi := total*float64(i)/n, total*float64(i+1)/n
			flo, fhi := f(lo), f(hi)
			if math.IsNaN(flo) || math.IsNaN(fhi) || (flo > 0) == (fhi > 0) {
				continue
			}
			for k := 0; k < 50; k++ {
				mid := (lo + hi) / 2
				if fm := f(mid); (fm > 0) == (flo > 0) {
					lo, flo = mid, fm
				} else {
					hi = mid
				}
			}
			al := (lo + hi) / 2
			if be := beOf(al); al > 0 && be > 0 && be < total {
				cands = append(cands, curve(al, be))
			}
		}
	}

	var best [4]Point
	bestDev, found := math.Inf(1), false
	for _, m := range cands {
		for i := range ss {
			ss[i].u = chordU[i]
		}
		project(m)
		if dev, ok := mergeFits(m, a, b, ss, tol, loose); ok && dev < bestDev {
			best, bestDev, found = m, dev, true
		}
	}
	return best, bestDev, found
}

// mergeFits reports whether the merged curve m may replace the curves a and
// b, sampled as ss with the parameters of m nearest to the samples, and
// returns the largest distance between m and the two.  It may if it lies
// within MergeTolerance of them, keeps its curvature within the bounds of
// mergeCurvatureShare and mergeCurvatureFloor, and runs beyond its end points
// by at most mergeOvershootExact, and by at most mergeOvershoot with its
// points rounded to the nearest integers.
func mergeFits(m, a, b [4]Point, ss []mergeSample, tol float64, loose bool) (float64, bool) {
	// The distance both ways.
	var dev float64
	for _, s := range ss {
		p := bez(m, s.u)
		dev = math.Max(dev, math.Hypot(p.X-s.q.X, p.Y-s.q.Y))
	}
	dense := func(c [4]Point) []Point {
		var ps []Point
		for i := 0; i <= 64; i++ {
			ps = append(ps, bez(c, float64(i)/64))
		}
		return ps
	}
	exact := append(dense(a), dense(b)...)
	for _, p := range dense(m) {
		best := math.Inf(1)
		for i := 1; i < len(exact); i++ {
			q := closest(exact[i-1], exact[i], p)
			best = math.Min(best, math.Hypot(p.X-q.X, p.Y-q.Y))
		}
		dev = math.Max(dev, best)
	}
	if dev > tol {
		return dev, false
	}

	// The curvature within the range of the exact curvature and at the ends
	// close to it.
	lo, hi, big := math.Inf(1), math.Inf(-1), 0.0
	for _, s := range ss {
		lo, hi, big = math.Min(lo, s.k), math.Max(hi, s.k), math.Max(big, math.Abs(s.k))
	}
	share, over, overRound := mergeCurvatureShare, mergeOvershootExact, mergeOvershoot
	if tol > MergeTolerance {
		share, over, overRound = 4*mergeCurvatureShare, 4*mergeOvershootExact, 4*mergeOvershoot
		if loose {
			over, overRound = tol/2, tol/2
		}
	}
	slack := share*big + mergeCurvatureFloor
	if !loose {
		for _, s := range ss {
			if k := curvatureAt(m, s.u); k < lo-slack || k > hi+slack {
				return dev, false
			}
		}
		for _, e := range []struct{ k, ke float64 }{{curvatureAt(m, 0), ss[0].k}, {curvatureAt(m, 1), ss[len(ss)-1].k}} {
			if math.Abs(e.k-e.ke) > share*math.Abs(e.ke)+mergeCurvatureFloor {
				return dev, false
			}
		}
	}

	// No new extreme beyond the end points, of the curve and of the curve
	// with its points rounded.
	if overshoot(m) > over {
		return dev, false
	}
	if overshoot([4]Point{roundPoint(m[0]), roundPoint(m[1]), roundPoint(m[2]), roundPoint(m[3])}) > overRound {
		return dev, false
	}
	return dev, true
}

// mergeCurvatureShare and mergeCurvatureFloor bound the curvature of a merged
// curve: it lies within the range of the curvature of the curves it replaces,
// widened by the share of their largest curvature and the floor, and at its
// ends it differs from theirs by at most the share of theirs and the floor.
const (
	mergeCurvatureShare = 0.15
	mergeCurvatureFloor = 0.001
)

// overshoot returns the largest distance by which the cubic c runs beyond its
// end points in x or in y.
func overshoot(c [4]Point) float64 {
	var worst float64
	for axis := 0; axis < 2; axis++ {
		get := func(q Point) float64 {
			if axis == 0 {
				return q.X
			}
			return q.Y
		}
		lo, hi := math.Min(get(c[0]), get(c[3])), math.Max(get(c[0]), get(c[3]))
		for _, t := range CubicExtrema(get(c[0]), get(c[1]), get(c[2]), get(c[3])) {
			v := get(bez(c, t))
			worst = math.Max(worst, math.Max(v-hi, lo-v))
		}
	}
	return worst
}
