// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/outline"
)

// rayCrossings returns the distances from c, sorted, at which the ray from c
// in the direction a (radians) crosses the outline.
func rayCrossings(path outline.Path, c Point, a float64) []float64 {
	d := pt(math.Cos(a), math.Sin(a))
	var out []float64
	for _, ct := range path {
		poly := ct.Flatten()
		for i := range poly {
			p, q := poly[i], poly[(i+1)%len(poly)]
			e := sub(q, p)
			den := cross(d, e)
			if math.Abs(den) < 1e-12 {
				continue
			}
			w := sub(p, c)
			t := cross(w, e) / den
			s := cross(w, d) / den
			if t > 0 && s >= 0 && s < 1 {
				out = append(out, t)
			}
		}
	}
	sort.Float64s(out)
	return out
}

// Stress returns the angle from the vertical, in degrees (positive leans to
// the right), of the line through the thinnest places of the top and the
// bottom of a closed round whose counter contains c: on vertical lines every
// unit across the counter, the thickness of the top and of the bottom stroke
// is the distance between the two crossings of the outline above c and below
// c nearest to the counter; each thinnest place is the middle of the lines
// within 1 unit of the least thickness, at the middle of the stroke.  This is
// the least vertical thickness.
func Stress(path outline.Path, c Point) (angle float64, top, bottom Point, thinTop, thinBottom float64) {
	type cut struct{ x, w, y float64 }
	var tops, bots []cut
	x0, _, x1, _ := path.BBox()
	for x := math.Ceil(x0); x <= x1; x++ {
		ys := path.CrossingsY(x)
		sort.Float64s(ys)
		var above, below []float64
		for _, y := range ys {
			if y > c.Y {
				above = append(above, y)
			} else {
				below = append(below, y)
			}
		}
		if len(above) >= 2 && len(below) >= 2 {
			tops = append(tops, cut{x, above[1] - above[0], (above[0] + above[1]) / 2})
			nb := len(below)
			bots = append(bots, cut{x, below[nb-1] - below[nb-2], (below[nb-1] + below[nb-2]) / 2})
		}
	}
	place := func(cs []cut) (Point, float64) {
		least := math.Inf(1)
		for _, q := range cs {
			least = math.Min(least, q.w)
		}
		var sx, sy, n float64
		for _, q := range cs {
			if q.w <= least+1 {
				sx += q.x
				sy += q.y
				n++
			}
		}
		return pt(sx/n, sy/n), least
	}
	if len(tops) == 0 {
		return math.NaN(), c, c, 0, 0
	}
	top, thinTop = place(tops)
	bottom, thinBottom = place(bots)
	d := sub(top, bottom)
	return math.Atan2(d.X, d.Y) * 180 / math.Pi, top, bottom, thinTop, thinBottom
}
