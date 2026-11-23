// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"slices"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// TestCedillaJoinsTheLetter: in every parameter set (in the routine tier the
// sets of composedRoutine) the cedilla of Ç, ç, Ş, ş, Ţ and ţ is joined to the
// letter: the rounded letter has as many contours as its rounded base.
func TestCedillaJoinsTheLetter(t *testing.T) {
	t.Parallel()
	for _, np := range goldenParams() {
		if !composedRoutine(np.name) {
			continue
		}
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, c := range composedList {
				if c.mark != markCedilla {
					continue
				}
				g := cachedRounded(np, c.name)
				b := cachedRounded(np, c.base)
				if len(g.Path) != len(b.Path) {
					t.Errorf("%s: %d contours, %s has %d", c.name, len(g.Path), c.base, len(b.Path))
				}
			}
		})
	}
}

// TestOgonekStaysInItsStroke: in every parameter set (in the routine tier the
// sets of composedRoutine) no point of the ogonek lies inside a counter of its
// base, and where the ogonek leaves the base it is at most as thick as the
// stroke of the base there.  The thickness at a point is the shortest chord
// of the ink through it, over directions 3 degrees apart; that of the ogonek
// is taken joinProbe below the lowest ink of the base at the anchor, in the
// middle of the ink of the mark at that height, and that of the base
// joinProbe above that ink.
func TestOgonekStaysInItsStroke(t *testing.T) {
	t.Parallel()
	const joinProbe = 4
	for _, np := range goldenParams() {
		if !composedRoutine(np.name) {
			continue
		}
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, c := range composedList {
				if c.mark != markOgonek {
					continue
				}
				base := cachedOutline1(np, c.base)
				dx, dy, f := c.place(np.params, base)
				mark := flatten(shiftInk(markOutline(np.params, c.mark, f), dx, dy))
				basePolys := flatten(base)
				for _, poly := range basePolys {
					if polyArea(poly) >= 0 {
						continue
					}
				points:
					for _, m := range mark {
						for _, q := range m {
							if inside(poly, q) {
								t.Errorf("%s: the point %v of the ogonek lies in a counter of %s", c.name, q, c.base)
								break points
							}
						}
					}
				}
				y := lowestInk(base, dx)
				lo, hi := runAround(mark, y-joinProbe, dx)
				thick := shortestChord(mark, outline.Point{X: (lo + hi) / 2, Y: y - joinProbe})
				stroke := shortestChord(basePolys, outline.Point{X: dx, Y: y + joinProbe})
				if thick > stroke+0.5 {
					t.Errorf("%s: the ogonek is %.1f thick where it leaves, the stroke of %s %.1f", c.name, thick, c.base, stroke)
				}
			}
		})
	}
}

// flatten returns the contours of the path as polygons.
func flatten(path outline.Path) [][]outline.Point {
	var out [][]outline.Point
	for _, ct := range path {
		out = append(out, ct.Flatten())
	}
	return out
}

// polyArea returns the signed area of the polygon, positive for a
// counter-clockwise one.
func polyArea(poly []outline.Point) float64 {
	a := 0.0
	for i, p := range poly {
		q := poly[(i+1)%len(poly)]
		a += p.X*q.Y - q.X*p.Y
	}
	return a / 2
}

// lineCrossings returns the parameters s, sorted, at which the line q + s*d
// crosses the edges of the polygons.
func lineCrossings(polys [][]outline.Point, q, d outline.Point) []float64 {
	var ss []float64
	for _, poly := range polys {
		for i, a := range poly {
			b := poly[(i+1)%len(poly)]
			e := outline.Point{X: b.X - a.X, Y: b.Y - a.Y}
			den := d.X*e.Y - d.Y*e.X
			if den == 0 {
				continue
			}
			w := outline.Point{X: a.X - q.X, Y: a.Y - q.Y}
			s := (w.X*e.Y - w.Y*e.X) / den
			u := (w.X*d.Y - w.Y*d.X) / den
			if u >= 0 && u < 1 {
				ss = append(ss, s)
			}
		}
	}
	slices.Sort(ss)
	return ss
}

// runAround returns the ends of the run of ink on the horizontal at the
// height y which contains x, or which lies nearest to it.
func runAround(polys [][]outline.Point, y, x float64) (lo, hi float64) {
	ss := lineCrossings(polys, outline.Point{X: x, Y: y}, outline.Point{X: 1})
	best := math.Inf(1)
	for i := 0; i+1 < len(ss); i += 2 {
		d := math.Max(0, math.Max(ss[i], -ss[i+1]))
		if d < best {
			best, lo, hi = d, x+ss[i], x+ss[i+1]
		}
	}
	return lo, hi
}

// shortestChord returns the length of the shortest chord of the ink through
// q, over directions 3 degrees apart.
func shortestChord(polys [][]outline.Point, q outline.Point) float64 {
	best := math.Inf(1)
	for deg := 0; deg < 180; deg += 3 {
		a := float64(deg) * math.Pi / 180
		ss := lineCrossings(polys, q, outline.Point{X: math.Cos(a), Y: math.Sin(a)})
		for i := 0; i+1 < len(ss); i += 2 {
			if ss[i] <= 0 && ss[i+1] >= 0 {
				best = math.Min(best, ss[i+1]-ss[i])
			}
		}
	}
	return best
}

// TestOgonekTipHasNoBend asks that the outer edge of the ogonek, from its lowest point
// to the cut of its tip, turns one way only: the curvature does not change sign.
func TestOgonekTipHasNoBend(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		path, _ := cachedOutline(np, "ogonek")
		if len(path) != 1 {
			t.Fatalf("%s: the ogonek has %d contours, want 1", np.name, len(path))
		}
		pts := path[0].Flatten()
		n := len(pts)
		low, tip := 0, 0
		for i, q := range pts {
			if q.Y < pts[low].Y {
				low = i
			}
			if q.X > pts[tip].X {
				tip = i
			}
		}
		// The contour is counter-clockwise: the lower edge runs from the lowest point
		// up to the tip, turning left where the ink is convex.
		var last outline.Point
		have := false
		worst := 0.0
		for i := low; i != tip; i = (i + 1) % n {
			q := pts[i]
			if have && math.Hypot(q.X-last.X, q.Y-last.Y) < 10 {
				continue
			}
			j := (i + 1) % n
			for j != tip && math.Hypot(pts[j].X-q.X, pts[j].Y-q.Y) < 10 {
				j = (j + 1) % n
			}
			if have {
				a := math.Atan2(q.Y-last.Y, q.X-last.X)
				b := math.Atan2(pts[j].Y-q.Y, pts[j].X-q.X)
				d := math.Remainder(b-a, 2*math.Pi) * 180 / math.Pi
				worst = math.Min(worst, d)
			}
			last, have = q, true
		}
		if worst < -0.2 {
			t.Errorf("%s: the outer edge of the ogonek turns back by %.2f degrees between its lowest point and the tip", np.name, -worst)
		}
	}
}
