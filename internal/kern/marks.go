// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import (
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/outline"
)

// The kerning of the letters with marks.
//
// A glyph which takes the kerning of another glyph (Glyph.Like: a letter with
// a mark takes that of its base, the dotless i and j those of i and j) is not
// kerned on its own: it is put in the classes of that glyph, and every single
// pair of that glyph is repeated for it.  Then every pair of such a glyph with
// any kerned glyph, in both orders, is checked against its marks: where a mark
// comes closer to the other glyph than Acc.MarkFloor, with the kern the pair
// has, the kern is raised to the least value which keeps that distance, and
// the pair is written as a single pair.  A pair is checked only if a bound on
// the distance, from the sample points of the other glyph at the heights of
// the mark, does not already show that the mark keeps the floor, so that the
// cost grows with the number of glyphs and not with its square.

// inherit puts the glyphs which take the kerning of others into the classes
// and the single pairs of those glyphs, and raises the kerns of the pairs in
// which a mark comes too close.
func (t *Table) inherit() {
	f := t.Font
	var derived []*Glyph
	followers := map[int][]int{} // glyph index -> indices of the glyphs like it
	for _, g := range f.Glyphs {
		if g.Like >= 0 {
			derived = append(derived, g)
			followers[g.Like] = append(followers[g.Like], g.Index)
		}
	}
	ligs := t.ligatures()
	if len(derived) == 0 && len(ligs) == 0 {
		return
	}
	for _, g := range derived {
		if c, ok := t.Class1[g.Like]; ok {
			t.Class1[g.Index] = c
		}
		if c, ok := t.Class2[g.Like]; ok {
			t.Class2[g.Index] = c
		}
	}
	firsts, seconds := map[int][]int{}, map[int][]int{}
	for i, l := range followers {
		firsts[i] = append(firsts[i], l...)
		seconds[i] = append(seconds[i], l...)
	}
	for _, g := range ligs {
		firsts[g.Last] = append(firsts[g.Last], g.Index)
		seconds[g.First] = append(seconds[g.First], g.Index)
	}
	single := make(map[[2]int]int, len(t.Single))
	for key, v := range t.Single {
		for _, a := range append([]int{key[0]}, firsts[key[0]]...) {
			for _, b := range append([]int{key[1]}, seconds[key[1]]...) {
				single[[2]int{a, b}] = v
			}
		}
	}
	t.Single = single

	var others []*Glyph
	for _, g := range f.Glyphs {
		if g.Kind != Skip {
			others = append(others, g)
		}
	}
	floor := f.P.Acc.MarkFloor
	for _, d := range derived {
		if len(d.marks) == 0 {
			continue
		}
		for _, o := range others {
			if o.Like >= 0 && o.Index < d.Index && len(o.marks) > 0 {
				continue // the pair was checked from the side of o
			}
			for _, pr := range [2][2]*Glyph{{o, d}, {d, o}} {
				a, b := pr[0], pr[1]
				k := t.Kern(a.Index, b.Index)
				if f.markBound(a, b, float64(k)) >= floor {
					continue
				}
				d0 := f.MarkDistance(a, b, float64(k), floor)
				if d0 >= floor {
					continue
				}
				// The least kern which keeps the floor: a shift of b changes
				// the distance by at most its length, so that the kern must
				// grow by at least the floor less the distance.
				kk := k + int(math.Ceil(floor-d0))
				for kk < k+400 && f.MarkDistance(a, b, float64(kk), floor) < floor {
					kk++
				}
				t.Single[[2]int{a.Index, b.Index}] = kk
				t.Marked = append(t.Marked, &Pair{
					Left: a, Right: b, Kern: kk, Floor: true,
					Dist0: d0, Dist1: f.MarkDistance(a, b, float64(kk), floor),
					LClass: t.Class1[a.Index], RClass: t.Class2[b.Index],
				})
			}
		}
	}
	t.ligatureFloor(ligs)
	sort.Slice(t.Marked, func(i, j int) bool {
		a, b := t.Marked[i], t.Marked[j]
		if a.Left.Index != b.Left.Index {
			return a.Left.Index < b.Left.Index
		}
		return a.Right.Index < b.Right.Index
	})
}

// MarkDistance returns the closest distance between the marks of the glyphs
// and the other glyph of the pair a, b, with b set at the advance of a plus
// the kern, or limit if that is less: the marks of b against all of a, and the
// marks of a against all of b.  It is the distance of sample points at most
// ptSpacing apart.
func (f *Font) MarkDistance(a, b *Glyph, kern, limit float64) float64 {
	off := a.Advance + kern
	d := limit
	if len(b.marks) > 0 {
		d = pointsDistance(b.marks, off, a.rpts, a.rys, 0, d)
	}
	if len(a.marks) > 0 {
		d = pointsDistance(a.marks, 0, b.lpts, b.lys, off, d)
	}
	return d
}

// pointsDistance returns the closest distance between the points ps, moved by
// dp, and the points qs (sorted by height, with the heights ys), moved by dq,
// or limit if that is less.
func pointsDistance(ps []outline.Point, dp float64, qs []outline.Point, ys []float64, dq, limit float64) float64 {
	best := limit
	for _, p := range ps {
		r := best
		lo := sort.SearchFloat64s(ys, p.Y-r)
		b2 := best * best
		for j := lo; j < len(qs) && ys[j] <= p.Y+r; j++ {
			q := qs[j]
			dx := q.X + dq - p.X - dp
			if dx >= best || dx <= -best {
				continue
			}
			dy := q.Y - p.Y
			if d2 := dx*dx + dy*dy; d2 < b2 {
				b2 = d2
				best = math.Sqrt(d2)
			}
		}
	}
	return best
}

// markBound returns a lower bound of MarkDistance: the horizontal distance
// between each mark and the ink of the other glyph within the floor of the
// heights of the mark, or +Inf where the boxes show that no ink is that near.
func (f *Font) markBound(a, b *Glyph, kern float64) float64 {
	floor := f.P.Acc.MarkFloor
	off := a.Advance + kern
	bound := math.Inf(1)
	if len(b.marks) > 0 && a.Top > b.my0-floor && a.Bottom < b.my1+floor && b.mx0+off-a.xMax < floor {
		if xa, ok := extentIn(a.rpts, a.rys, b.my0-floor, b.my1+floor, true); ok {
			bound = math.Min(bound, b.mx0+off-xa)
		}
	}
	if len(a.marks) > 0 && b.Top > a.my0-floor && b.Bottom < a.my1+floor && b.xMin+off-a.mx1 < floor {
		if xb, ok := extentIn(b.lpts, b.lys, a.my0-floor, a.my1+floor, false); ok {
			bound = math.Min(bound, xb+off-a.mx1)
		}
	}
	return bound
}

// extentIn returns the greatest (right set) or the least abscissa of the
// points between the heights y0 and y1, and false if there are none.
func extentIn(pts []outline.Point, ys []float64, y0, y1 float64, right bool) (float64, bool) {
	lo := sort.SearchFloat64s(ys, y0)
	x, ok := math.Inf(-1), false
	if !right {
		x = math.Inf(1)
	}
	for j := lo; j < len(pts) && ys[j] <= y1; j++ {
		ok = true
		if right {
			x = math.Max(x, pts[j].X)
		} else {
			x = math.Min(x, pts[j].X)
		}
	}
	return x, ok
}
