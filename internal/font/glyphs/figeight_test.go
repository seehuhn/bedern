// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sort"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The heavy stroke of 8 runs from the upper left through the waist to the
// lower right, and the upper right and the lower left of the bowls are thin
// arcs which run into it.
const (
	// The segment between the nearest points of the two counters leans from
	// the vertical, its upper end on the right, by this many degrees.
	eightLeanMin, eightLeanMax = 14.0, 35.0
	// The least distance between the counters, the thickness of the spine,
	// in stems of the figures.
	eightWaistMin, eightWaistMax = 0.86, 1.26
	// The height of the upper bowl over that of the lower, each measured from
	// the middle between the counters.
	eightBowlsMin, eightBowlsMax = 0.797, 0.921
	// The thin arc ends at most this share of its thickness at the extreme.
	eightArcThins = 0.75
)

// eightSamples returns points along the contour, edgeSamples on each
// segment, and for each point whether a corner of the contour (a turn of more
// than eightCornerTurn radians) follows it.
func eightSamples(ct outline.Contour) ([]outline.Point, []bool) {
	n := len(ct.Segs)
	var pts []outline.Point
	var corner []bool
	for i := 0; i < n; i++ {
		c := segCubic(ct, i)
		for s := 0; s < edgeSamples; s++ {
			pts = append(pts, c.at(float64(s)/edgeSamples))
			corner = append(corner, false)
		}
		next := segCubic(ct, (i+1)%n)
		ta, tb := c.tangent(1), next.tangent(0)
		if math.Abs(math.Atan2(ta.X*tb.Y-ta.Y*tb.X, ta.X*tb.X+ta.Y*tb.Y)) > eightCornerTurn {
			corner[len(corner)-1] = true
		}
	}
	return pts, corner
}

const eightCornerTurn = 0.1

// leastDistance returns the least distance between the point sets and the
// indices of the nearest pair.
func leastDistance(a, b []outline.Point) (float64, int, int) {
	best, ia, ib := math.Inf(1), 0, 0
	for i, p := range a {
		for j, q := range b {
			if d := math.Hypot(p.X-q.X, p.Y-q.Y); d < best {
				best, ia, ib = d, i, j
			}
		}
	}
	return best, ia, ib
}

func TestFigureEightSpine(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		path := capOutline(np, "eight")
		if len(path) != 3 {
			t.Fatalf("%s: 8 has %d contours, want 3", np.name, len(path))
		}
		type contour struct {
			pts    []outline.Point
			corner []bool
			x0, y0 float64
			x1, y1 float64
		}
		var cs []contour
		for _, ct := range path {
			pts, corner := eightSamples(ct)
			c := contour{pts: pts, corner: corner, x0: math.Inf(1), y0: math.Inf(1), x1: math.Inf(-1), y1: math.Inf(-1)}
			for _, q := range pts {
				c.x0, c.y0 = math.Min(c.x0, q.X), math.Min(c.y0, q.Y)
				c.x1, c.y1 = math.Max(c.x1, q.X), math.Max(c.y1, q.Y)
			}
			cs = append(cs, c)
		}
		// The outer contour is the tallest, the upper counter the higher one.
		sort.Slice(cs, func(i, j int) bool { return cs[i].y1-cs[i].y0 > cs[j].y1-cs[j].y0 })
		outer, up, lo := cs[0], cs[1], cs[2]
		if up.y0 < lo.y0 {
			up, lo = lo, up
		}

		d, iu, il := leastDistance(up.pts, lo.pts)
		p, q := up.pts[iu], lo.pts[il]
		lean := math.Atan2(p.X-q.X, p.Y-q.Y) * 180 / math.Pi
		if lean < eightLeanMin || lean > eightLeanMax {
			t.Errorf("%s: the counters of 8 are nearest along a line %.1f degrees from the vertical, want %.0f to %.0f", np.name, lean, eightLeanMin, eightLeanMax)
		}
		if w := d / np.params.Figures.Stem; w < eightWaistMin || w > eightWaistMax {
			t.Errorf("%s: the spine of 8 is %.1f thick, %.2f stems, want %.2f to %.2f", np.name, d, w, eightWaistMin, eightWaistMax)
		}
		mid := (p.Y + q.Y) / 2
		if r := (outer.y1 - mid) / (mid - outer.y0); r < eightBowlsMin || r > eightBowlsMax {
			t.Errorf("%s: the upper bowl of 8 is %.3f as high as the lower, want %.3f to %.3f", np.name, r, eightBowlsMin, eightBowlsMax)
		}

		// The thin arcs: from the right extreme of the upper counter down
		// (the left extreme of the lower counter up) to the first corner,
		// the distance to the outer contour does not grow by more than a
		// unit, and it ends at most eightArcThins of where it starts.
		arc := func(name string, c contour, start int) {
			var prev, first float64
			n := len(c.pts)
			for k := 0; k < n; k++ {
				i := (start + k) % n
				th, _, _ := leastDistance(c.pts[i:i+1], outer.pts)
				if k == 0 {
					first = th
				} else if th > prev+1 {
					t.Errorf("%s: the %s arc of 8 grows from %.1f to %.1f at (%.0f, %.0f)", np.name, name, prev, th, c.pts[i].X, c.pts[i].Y)
					return
				}
				prev = th
				if c.corner[i] {
					break
				}
			}
			if prev > eightArcThins*first {
				t.Errorf("%s: the %s arc of 8 thins only from %.1f to %.1f", np.name, name, first, prev)
			}
		}
		right, left := 0, 0
		for i, q := range up.pts {
			if q.X > up.pts[right].X {
				right = i
			}
		}
		for i, q := range lo.pts {
			if q.X < lo.pts[left].X {
				left = i
			}
		}
		arc("upper right", up, right)
		arc("lower left", lo, left)
	})
}
