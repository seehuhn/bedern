// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/outline"
)

// The union of overlapping parts.
//
// The symbols are made of strokes which cross or run into each other: the
// diagonal of the ampersand crosses its thin stroke, the bars of the euro and of
// the sterling cross their rounds and stems, the stroke of the dollar runs
// through the S.  A glyph draws each part as a closed contour of its own, with
// the orientation of doc.go (counter-clockwise outside, clockwise round a
// counter), and union returns the outline of the ink they cover together: the
// pieces of the parts are cut where they cross each other, and a cut piece is
// kept if the ink lies on its left and no ink on its right.  The kept pieces are
// joined into contours again.  Where two edges cross, the outline has a corner,
// as at the crossbar of f.  The parts must cross where they meet, not touch: an
// edge which runs along another edge or touches it is a mistake of the glyph.

// unionSamples is the number of straight pieces by which a curve is followed
// when the crossings are searched for and the winding number is counted.
const unionSamples = 64

// A cutPoint is a point where a piece is cut: its parameter and the point.
type cutPoint struct {
	t  float64
	at outline.Point
}

// union returns the contours of the ink which the closed chains cover together.
// It panics if the cut pieces cannot be joined into closed contours.
func union(parts ...chain) []chain {
	// Every piece is cut at its extremes.
	var pcs []piece
	var pieceContour []int
	for k, c := range parts {
		for _, p := range c {
			if p.line {
				pcs = append(pcs, p)
				pieceContour = append(pieceContour, k)
				continue
			}
			for _, q := range splitAtExtrema(p.cubic) {
				pcs = append(pcs, q)
				pieceContour = append(pieceContour, k)
			}
		}
	}
	polys := make([][]outline.Point, len(parts))
	for k, c := range parts {
		for _, p := range c {
			polys[k] = append(polys[k], samples(p)[1:]...)
		}
	}

	cuts := make([][]cutPoint, len(pcs))
	for i := range pcs {
		for j := i + 1; j < len(pcs); j++ {
			for _, x := range pieceCrossings(pcs[i], pcs[j]) {
				cuts[i] = append(cuts[i], cutPoint{x[0], outline.Point{X: x[2], Y: x[3]}})
				cuts[j] = append(cuts[j], cutPoint{x[1], outline.Point{X: x[2], Y: x[3]}})
			}
		}
	}

	var kept []piece
	for i, p := range pcs {
		cs := cuts[i]
		sort.Slice(cs, func(a, b int) bool { return cs[a].t < cs[b].t })
		// A crossing at the common end of two pieces of the other contour is found
		// on both of them: it is one cut.
		var merged []cutPoint
		for _, c := range cs {
			if n := len(merged); n > 0 && math.Hypot(c.at.X-merged[n-1].at.X, c.at.Y-merged[n-1].at.Y) < 1e-4 {
				continue
			}
			if math.Hypot(c.at.X-p.P0.X, c.at.Y-p.P0.Y) < 1e-4 {
				c.at, c.t = p.P0, 0
			}
			if math.Hypot(c.at.X-p.P3.X, c.at.Y-p.P3.Y) < 1e-4 {
				c.at, c.t = p.P3, 1
			}
			merged = append(merged, c)
		}
		cs = merged
		prevT, prevP := 0.0, p.P0
		for k := 0; k <= len(cs); k++ {
			t, at := 1.0, p.P3
			if k < len(cs) {
				t, at = cs[k].t, cs[k].at
			}
			if t-prevT < 1e-9 {
				continue
			}
			var q piece
			if p.line {
				q = lineSeg(prevP, at)
			} else {
				c := p.cubic.sub(prevT, t)
				c.P0, c.P3 = prevP, at
				q = curveSeg(c)
			}
			if keepPiece(q, polys) {
				kept = append(kept, q)
			}
			prevT, prevP = t, at
		}
	}
	return linkPieces(kept)
}

// samples returns points of the piece, its start first.
func samples(p piece) []outline.Point {
	if p.line {
		return []outline.Point{p.P0, p.P3}
	}
	pts := make([]outline.Point, unionSamples+1)
	for i := range pts {
		pts[i] = p.at(float64(i) / unionSamples)
	}
	return pts
}

// pieceCrossings returns the crossings of two pieces as (ta, tb, x, y), apart
// from a common end point, where the pieces of one contour meet.
func pieceCrossings(a, b piece) [][4]float64 {
	pa, pb := samples(a), samples(b)
	var out [][4]float64
	near := func(u, v outline.Point) bool { return math.Hypot(u.X-v.X, u.Y-v.Y) < 1e-6 }
	for i := 0; i+1 < len(pa); i++ {
		for j := 0; j+1 < len(pb); j++ {
			d1 := outline.Point{X: pa[i+1].X - pa[i].X, Y: pa[i+1].Y - pa[i].Y}
			d2 := outline.Point{X: pb[j+1].X - pb[j].X, Y: pb[j+1].Y - pb[j].Y}
			den := d1.X*d2.Y - d1.Y*d2.X
			if math.Abs(den) < 1e-12 {
				continue
			}
			w := outline.Point{X: pb[j].X - pa[i].X, Y: pb[j].Y - pa[i].Y}
			u := (w.X*d2.Y - w.Y*d2.X) / den
			v := (w.X*d1.Y - w.Y*d1.X) / den
			if u < 0 || u >= 1 || v < 0 || v >= 1 {
				continue
			}
			ta := (float64(i) + u) / float64(len(pa)-1)
			tb := (float64(j) + v) / float64(len(pb)-1)
			ta, tb = refineCrossing(a, b, ta, tb)
			q := pieceAt(a, ta)
			// A common end of two pieces is not a crossing.
			if (ta < 1e-6 || ta > 1-1e-6) && (tb < 1e-6 || tb > 1-1e-6) &&
				(near(q, a.P0) || near(q, a.P3)) && (near(q, b.P0) || near(q, b.P3)) {
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

func pieceAt(p piece, t float64) outline.Point {
	if p.line {
		return lerp(p.P0, p.P3, t)
	}
	return p.at(t)
}

func pieceDeriv(p piece, t float64) outline.Point {
	if p.line {
		return outline.Point{X: p.P3.X - p.P0.X, Y: p.P3.Y - p.P0.Y}
	}
	return p.deriv(t)
}

// refineCrossing improves the parameters of a crossing by Newton's method.
func refineCrossing(a, b piece, ta, tb float64) (float64, float64) {
	for k := 0; k < 30; k++ {
		fa, fb := pieceAt(a, ta), pieceAt(b, tb)
		ex, ey := fa.X-fb.X, fa.Y-fb.Y
		if math.Hypot(ex, ey) < 1e-11 {
			break
		}
		da, db := pieceDeriv(a, ta), pieceDeriv(b, tb)
		det := -da.X*db.Y + da.Y*db.X
		if math.Abs(det) < 1e-14 {
			break
		}
		dta := (ex*db.Y - ey*db.X) / det
		dtb := (da.Y*ex - da.X*ey) / det
		ta = math.Min(1, math.Max(0, ta+dta))
		tb = math.Min(1, math.Max(0, tb+dtb))
	}
	return ta, tb
}

// keepPiece reports whether the ink lies on the left of the piece and none on
// its right, judged a quarter of a unit to either side of its middle.
func keepPiece(q piece, polys [][]outline.Point) bool {
	m := pieceAt(q, 0.5)
	d := pieceDeriv(q, 0.5)
	n := math.Hypot(d.X, d.Y)
	if n == 0 {
		return false
	}
	nl := outline.Point{X: -d.Y / n * 0.25, Y: d.X / n * 0.25}
	left := winding(polys, outline.Point{X: m.X + nl.X, Y: m.Y + nl.Y})
	right := winding(polys, outline.Point{X: m.X - nl.X, Y: m.Y - nl.Y})
	return left > 0 && right <= 0
}

// winding returns the sum of the winding numbers of the closed polygons about q.
func winding(polys [][]outline.Point, q outline.Point) int {
	w := 0
	for _, poly := range polys {
		for i, a := range poly {
			b := poly[(i+1)%len(poly)]
			if a.Y <= q.Y {
				if b.Y > q.Y && (b.X-a.X)*(q.Y-a.Y)-(q.X-a.X)*(b.Y-a.Y) > 0 {
					w++
				}
			} else if b.Y <= q.Y && (b.X-a.X)*(q.Y-a.Y)-(q.X-a.X)*(b.Y-a.Y) < 0 {
				w--
			}
		}
	}
	return w
}

// linkPieces joins the pieces into closed chains.
func linkPieces(pcs []piece) []chain {
	used := make([]bool, len(pcs))
	var out []chain
	for s := range pcs {
		if used[s] {
			continue
		}
		var c chain
		i := s
		for {
			used[i] = true
			c = append(c, pcs[i])
			end := pcs[i].P3
			if math.Hypot(end.X-pcs[s].P0.X, end.Y-pcs[s].P0.Y) < 1e-3 {
				c[len(c)-1].P3 = pcs[s].P0
				break
			}
			next, best := -1, 1e-3
			for j := range pcs {
				if used[j] {
					continue
				}
				if d := math.Hypot(pcs[j].P0.X-end.X, pcs[j].P0.Y-end.Y); d < best {
					next, best = j, d
				}
			}
			if next < 0 {
				panic(fmt.Sprintf("glyphs: union: the outline is open at %v", end))
			}
			pcs[next].P0 = end
			i = next
		}
		out = append(out, c)
	}
	return out
}
