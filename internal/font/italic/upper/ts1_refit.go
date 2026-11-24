// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"math"

	"seehuhn.de/go/bedern/internal/outline"
)

// The curves which the shear distorts, drawn again.  A glyph of TS1 whose
// source has a refit function passes the outline of the italic, sheared and
// with nodes at the extremes but before the rounding, through it; the
// functions here are the pieces which the refits are made of.  Each keeps the
// nodes it does not name, the directions at the nodes and the shape of the
// curves to within a unit or two.

// overTarget is how far a curve may run beyond its nodes after a refit; the
// profile of the smoothness check allows extremeTol, and the rounding adds a
// little.
const overTarget = 0.15

// segAt returns the contour and the index of the segment of the path which
// ends nearest to q, if one ends within 8 units of it.
func segAt(path outline.Path, q Point) (ci, si int) {
	best := 8.0
	ci, si = -1, -1
	for c := range path {
		for s := range path[c].Segs {
			if d := dist(path[c].Segs[s].P3, q); d < best {
				best, ci, si = d, c, s
			}
		}
	}
	if ci < 0 {
		panic("upper: no node near the place of a refit")
	}
	return ci, si
}

// curveAt returns the cubic at index si of the contour.
func curveAt(c outline.Contour, si int) cubic {
	p0 := c.Start
	if si > 0 {
		p0 = c.Segs[si-1].P3
	}
	s := c.Segs[si]
	return cubic{p0, s.P1, s.P2, s.P3}
}

// setCurve replaces the curve at index si of the contour.
func setCurve(c *outline.Contour, si int, cu cubic) {
	c.Segs[si] = outline.Segment{IsCurve: true, P1: cu[1], P2: cu[2], P3: cu[3]}
	if si == 0 {
		c.Start = cu[0]
	}
}

// trimOvershoot returns the curve with the handle next to its larger run
// beyond the end points shortened along its direction until the curve runs
// beyond them by no more than overTarget.  The handle which is shortened is
// the one at the end nearer to the extreme.
func trimOvershoot(cu cubic) cubic {
	over := func(c cubic) float64 {
		var worst float64
		for axis := 0; axis < 2; axis++ {
			get := func(q Point) float64 {
				if axis == 0 {
					return q.X
				}
				return q.Y
			}
			lo, hi := math.Min(get(c[0]), get(c[3])), math.Max(get(c[0]), get(c[3]))
			for _, t := range outline.CubicExtrema(get(c[0]), get(c[1]), get(c[2]), get(c[3])) {
				v := get(c.at(t))
				worst = math.Max(worst, math.Max(v-hi, lo-v))
			}
		}
		return worst
	}
	if over(cu) <= overTarget {
		return cu
	}
	// Of the two handles the one which needs the smaller shortening is
	// shortened.
	shorten := func(c cubic, end int, s float64) cubic {
		if end == 3 {
			c[2] = add(c[3], mul(sub(c[2], c[3]), s))
		} else {
			c[1] = add(c[0], mul(sub(c[1], c[0]), s))
		}
		return c
	}
	bestEnd, bestS := 3, -1.0
	for _, end := range []int{0, 3} {
		lo, hi := 0.0, 1.0
		for i := 0; i < 40; i++ {
			s := (lo + hi) / 2
			if over(shorten(cu, end, s)) <= overTarget {
				lo = s
			} else {
				hi = s
			}
		}
		if lo > bestS {
			bestEnd, bestS = end, lo
		}
	}
	return shorten(cu, bestEnd, bestS)
}

// trimOvershoots returns the path with every curve trimmed (trimOvershoot).
func trimOvershoots(path outline.Path) outline.Path {
	out := make(outline.Path, len(path))
	for ci, c := range path {
		nc := outline.Contour{Start: c.Start, Segs: append([]outline.Segment(nil), c.Segs...)}
		for si, s := range nc.Segs {
			if s.IsCurve {
				setCurve(&nc, si, trimOvershoot(curveAt(nc, si)))
			}
		}
		out[ci] = nc
	}
	return out
}
