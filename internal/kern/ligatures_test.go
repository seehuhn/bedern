// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
)

// TestLigaturesTakeTheSidesOfTheirLetters: a ligature is kerned on its left as
// its first letter and on its right as its last letter, or more loosely where
// the floor asks for it, and keeps the floor from every kerned glyph on both
// sides.
func TestLigaturesTakeTheSidesOfTheirLetters(t *testing.T) {
	t.Parallel()
	for _, s := range allSets()[:4] {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			gs := glyphstest.Set(s.p)
			tb := computed(s.p)
			floor := s.p.Kern.Floor
			n := 0
			for _, l := range tb.Font.Glyphs {
				if l.First < 0 {
					continue
				}
				n++
				for _, o := range tb.Font.Glyphs {
					if o.Kind == Skip {
						continue
					}
					if k, base := tb.Kern(l.Index, o.Index), tb.Kern(l.Last, o.Index); k < base {
						t.Errorf("%s %s: kern %d, tighter than %d of %s %s", l.Name, o.Name, k, base, gs[l.Last].Name, o.Name)
					}
					if k, base := tb.Kern(o.Index, l.Index), tb.Kern(o.Index, l.First); k < base {
						t.Errorf("%s %s: kern %d, tighter than %d of %s %s", o.Name, l.Name, k, base, o.Name, gs[l.First].Name)
					}
					for _, pr := range [2][2]int{{l.Index, o.Index}, {o.Index, l.Index}} {
						k := tb.Kern(pr[0], pr[1])
						if tb.Font.Distance(tb.Font.Glyphs[pr[0]], tb.Font.Glyphs[pr[1]], float64(k)) >= floor+20 {
							continue
						}
						if d := exactDistance(gs[pr[0]], gs[pr[1]], float64(k)); d < floor-2 {
							t.Errorf("%s %s with kern %d: closest distance %.1f, floor %v", gs[pr[0]].Name, gs[pr[1]].Name, k, d, floor)
						}
					}
				}
			}
			if n == 0 {
				t.Fatal("no ligatures")
			}
		})
	}
}
