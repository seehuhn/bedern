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
	sets := allSets()
	for _, s := range sets[:min(4, len(sets))] {
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

// TestGlyphsDrawnAsOthersTakeTheirKerning: the two letters S are kerned as S
// on both sides, the form of i of the feature locl as i, the hyphen and the
// non-breaking hyphen as the hyphen-minus, the figure dash as the en dash and
// the horizontal bar as the em dash, against every glyph and in both orders;
// and each of them is kerned at all.
func TestGlyphsDrawnAsOthersTakeTheirKerning(t *testing.T) {
	t.Parallel()
	sets := allSets()
	for _, s := range sets[:min(4, len(sets))] {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			tb := computed(s.p)
			index := map[string]int{}
			for _, g := range tb.Font.Glyphs {
				index[g.Name] = g.Index
			}
			for _, c := range [][2]string{
				{"S_S", "S"}, {"i.loclTRK", "i"}, {"uni2010", "hyphen"}, {"uni2011", "hyphen"},
				{"figuredash", "endash"}, {"uni2015", "emdash"},
			} {
				g, okG := index[c[0]]
				like, okL := index[c[1]]
				if !okG || !okL {
					t.Fatalf("no glyph %s or %s", c[0], c[1])
				}
				kerned := 0
				for _, o := range tb.Font.Glyphs {
					if k, want := tb.Kern(g, o.Index), tb.Kern(like, o.Index); k != want {
						t.Errorf("%s %s: kern %d, and %s %s has %d", c[0], o.Name, k, c[1], o.Name, want)
					} else if k != 0 {
						kerned++
					}
					if k, want := tb.Kern(o.Index, g), tb.Kern(o.Index, like); k != want {
						t.Errorf("%s %s: kern %d, and %s %s has %d", o.Name, c[0], k, o.Name, c[1], want)
					} else if k != 0 {
						kerned++
					}
				}
				if kerned == 0 {
					t.Errorf("%s has no kern", c[0])
				}
			}
		})
	}
}
