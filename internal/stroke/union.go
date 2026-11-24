// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"fmt"
	"math"
	"sort"
)

// The union of strokes.
//
// The strokes of a letter overlap: a branch runs into a stem, a bowl joins its
// stem twice, the bar of an italic e starts inside its own stroke.  Union
// returns the outline of the ink they cover together, on the principle of the
// union of the symbols of the font (symunion.go, written again here): every
// contour is closed and has the ink on its left, the pieces are cut where they
// cross, a cut piece is kept if there is ink on its left and none on its right,
// and the kept pieces are joined into contours again.  Where two edges cross,
// the outline has a true corner.  Edges must cross where they meet, not run
// along each other.

// unionSamples is the number of straight pieces by which a curve is followed
// when crossings are searched and winding numbers counted.
const unionSamples = 64

type cut struct {
	t  float64
	at Point
}

// Union returns the contours of the ink which the closed chains cover
// together.
func Union(contours []Chain) ([]Chain, error) {
	var pcs []Piece
	for _, c := range contours {
		for _, p := range c {
			if p.C.IsStraight() || p.Fixed {
				pcs = append(pcs, p)
				continue
			}
			for _, q := range splitExtremes(p.C) {
				pcs = append(pcs, Piece{C: q.C, Kind: p.Kind})
			}
		}
	}
	polys := make([][]Point, len(contours))
	for k, c := range contours {
		for _, p := range c {
			polys[k] = append(polys[k], pieceSamples(p.C)[1:]...)
		}
	}
	type box struct{ x0, y0, x1, y1 float64 }
	boxes := make([]box, len(pcs))
	for i, p := range pcs {
		b := box{math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)}
		for _, q := range p.C {
			b.x0, b.y0, b.x1, b.y1 = math.Min(b.x0, q.X), math.Min(b.y0, q.Y), math.Max(b.x1, q.X), math.Max(b.y1, q.Y)
		}
		boxes[i] = b
	}
	cuts := make([][]cut, len(pcs))
	for i := range pcs {
		for j := i + 1; j < len(pcs); j++ {
			a, b := boxes[i], boxes[j]
			if a.x1 < b.x0-1e-6 || b.x1 < a.x0-1e-6 || a.y1 < b.y0-1e-6 || b.y1 < a.y0-1e-6 {
				continue
			}
			for _, x := range crossings(pcs[i].C, pcs[j].C) {
				cuts[i] = append(cuts[i], cut{x[0], Point{X: x[2], Y: x[3]}})
				cuts[j] = append(cuts[j], cut{x[1], Point{X: x[2], Y: x[3]}})
			}
		}
	}
	var kept []Piece
	for i, p := range pcs {
		cs := cuts[i]
		sort.Slice(cs, func(a, b int) bool { return cs[a].t < cs[b].t })
		prevT, prevP := 0.0, p.C[0]
		for k := 0; k <= len(cs); k++ {
			t, at := 1.0, p.C[3]
			if k < len(cs) {
				t, at = cs[k].t, cs[k].at
				if dist(at, p.C[0]) < 1e-4 || dist(at, p.C[3]) < 1e-4 || dist(at, prevP) < 1e-4 {
					continue
				}
			}
			if t-prevT < 1e-9 {
				continue
			}
			c := p.C.Sub(prevT, t)
			c[0], c[3] = prevP, at
			if keep(c, polys) {
				kept = append(kept, Piece{C: c, Kind: p.Kind, Fixed: p.Fixed && len(cs) == 0})
			}
			prevT, prevP = t, at
		}
	}
	return link(kept)
}

// pieceSamples returns points of the curve, its start first.
func pieceSamples(c Cubic) []Point {
	if c.IsStraight() {
		return []Point{c[0], c[3]}
	}
	pts := make([]Point, unionSamples+1)
	for i := range pts {
		pts[i] = c.At(float64(i) / unionSamples)
	}
	return pts
}

// crossings returns the crossings of two curves as (ta, tb, x, y), apart from
// a common end point.
func crossings(a, b Cubic) [][4]float64 {
	pa, pb := pieceSamples(a), pieceSamples(b)
	var out [][4]float64
	near := func(u, v Point) bool { return dist(u, v) < 1e-6 }
	for i := 0; i+1 < len(pa); i++ {
		for j := 0; j+1 < len(pb); j++ {
			d1, d2 := sub(pa[i+1], pa[i]), sub(pb[j+1], pb[j])
			den := cross(d1, d2)
			if math.Abs(den) < 1e-12 {
				continue
			}
			w := sub(pb[j], pa[i])
			u, v := cross(w, d2)/den, cross(w, d1)/den
			if u < 0 || u >= 1 || v < 0 || v >= 1 {
				continue
			}
			ta := (float64(i) + u) / float64(len(pa)-1)
			tb := (float64(j) + v) / float64(len(pb)-1)
			ta, tb = refine(a, b, ta, tb)
			q := a.At(ta)
			if (near(q, a[0]) || near(q, a[3])) && (near(q, b[0]) || near(q, b[3])) {
				continue
			}
			dup := false
			for _, o := range out {
				if math.Abs(o[0]-ta) < 1e-7 && math.Abs(o[1]-tb) < 1e-7 {
					dup = true
				}
			}
			if !dup {
				out = append(out, [4]float64{ta, tb, q.X, q.Y})
			}
		}
	}
	return out
}

// refine improves the parameters of a crossing by Newton's method.
func refine(a, b Cubic, ta, tb float64) (float64, float64) {
	for k := 0; k < 30; k++ {
		e := sub(a.At(ta), b.At(tb))
		if norm(e) < 1e-11 {
			break
		}
		da, db := a.D(ta), b.D(tb)
		det := -da.X*db.Y + da.Y*db.X
		if math.Abs(det) < 1e-14 {
			break
		}
		ta = math.Min(1, math.Max(0, ta+(e.X*db.Y-e.Y*db.X)/det))
		tb = math.Min(1, math.Max(0, tb+(da.Y*e.X-da.X*e.Y)/det))
	}
	return ta, tb
}

// keep reports whether the ink lies on the left of the curve and none on its
// right, judged a quarter of a unit to either side of its middle.
func keep(c Cubic, polys [][]Point) bool {
	m, t := c.At(0.5), c.Tangent(0.5)
	n := mul(rot90(t), 0.25)
	return winding(polys, add(m, n)) > 0 && winding(polys, sub(m, n)) <= 0
}

// winding returns the sum of the winding numbers of the polygons about q.
func winding(polys [][]Point, q Point) int {
	w := 0
	for _, poly := range polys {
		for i, a := range poly {
			b := poly[(i+1)%len(poly)]
			c := (b.X-a.X)*(q.Y-a.Y) - (q.X-a.X)*(b.Y-a.Y)
			if a.Y <= q.Y {
				if b.Y > q.Y && c > 0 {
					w++
				}
			} else if b.Y <= q.Y && c < 0 {
				w--
			}
		}
	}
	return w
}

// link joins the pieces into closed chains.
func link(pcs []Piece) ([]Chain, error) {
	used := make([]bool, len(pcs))
	var out []Chain
	for s := range pcs {
		if used[s] {
			continue
		}
		var c Chain
		i := s
		for {
			used[i] = true
			c = append(c, pcs[i])
			end := pcs[i].C[3]
			if dist(end, pcs[s].C[0]) < 1e-3 {
				c[len(c)-1].C[3] = pcs[s].C[0]
				break
			}
			next, best := -1, 1e-3
			for j := range pcs {
				if !used[j] {
					if d := dist(pcs[j].C[0], end); d < best {
						next, best = j, d
					}
				}
			}
			if next < 0 {
				return nil, fmt.Errorf("stroke: union: the outline is open at %.2f,%.2f", end.X, end.Y)
			}
			pcs[next].C[0] = end
			i = next
		}
		out = append(out, c)
	}
	return out, nil
}

// signedArea returns the area enclosed by the chain, positive for a
// counter-clockwise chain.
func signedArea(c Chain) float64 {
	var pts []Point
	for _, p := range c {
		pts = append(pts, pieceSamples(p.C)[1:]...)
	}
	var a float64
	for i, p := range pts {
		q := pts[(i+1)%len(pts)]
		a += p.X*q.Y - q.X*p.Y
	}
	return a / 2
}
