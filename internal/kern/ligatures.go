// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import "math"

// The kerning of the ligatures.
//
// A ligature (Glyph.First and Glyph.Last) is not kerned on its own: it takes
// the class and the single pairs of the left side of its first letter and of
// the right side of its last letter (inherit), and the kern of a pair of
// those letters where the marks raise it.  Then every pair of a ligature
// with a kerned glyph, in both orders, is checked against the outlines: where
// the two come closer than Kern.Floor with the kern the pair has, the kern is
// raised to the least value which keeps the floor, and the pair is written as
// a single pair.

// ligatures returns the ligatures of the font, and puts each into the classes
// of the sides of its letters.
func (t *Table) ligatures() []*Glyph {
	var ligs []*Glyph
	for _, g := range t.Font.Glyphs {
		if g.First < 0 {
			continue
		}
		ligs = append(ligs, g)
		if c, ok := t.Class1[g.Last]; ok {
			t.Class1[g.Index] = c
		}
		if c, ok := t.Class2[g.First]; ok {
			t.Class2[g.Index] = c
		}
	}
	return ligs
}

// ligatureFloor raises the kern of every pair of a ligature and a kerned
// glyph which comes closer than the floor.
func (t *Table) ligatureFloor(ligs []*Glyph) {
	f := t.Font
	floor := f.P.Kern.Floor
	for _, l := range ligs {
		for _, o := range f.Glyphs {
			if o.Kind == Skip {
				continue
			}
			for _, pr := range [2][2]*Glyph{{l, o}, {o, l}} {
				a, b := pr[0], pr[1]
				if a.First >= 0 && b.First >= 0 && b == l && a != l {
					continue // the pair of two ligatures is checked from the side of a
				}
				// The kern of the letters, as the marks may have raised it.
				k := t.Kern(a.Index, b.Index)
				base := k
				if a.First >= 0 {
					base = max(base, t.Kern(a.Last, b.Index))
				}
				if b.First >= 0 {
					base = max(base, t.Kern(a.Index, b.First))
				}
				if base != k {
					k = base
					t.Single[[2]int{a.Index, b.Index}] = k
				}
				if minDepth(a.Right)+minDepth(b.Left)+float64(k) >= floor {
					continue
				}
				d0 := f.Distance(a, b, float64(k))
				if d0 >= floor {
					continue
				}
				kk := k + int(math.Ceil(floor-d0))
				for kk < k+400 && f.Distance(a, b, float64(kk)) < floor {
					kk++
				}
				t.Single[[2]int{a.Index, b.Index}] = kk
				t.Marked = append(t.Marked, &Pair{
					Left: a, Right: b, Kern: kk, Floor: true,
					Dist0: d0, Dist1: f.Distance(a, b, float64(kk)),
					LClass: t.Class1[a.Index], RClass: t.Class2[b.Index],
				})
			}
		}
	}
}
