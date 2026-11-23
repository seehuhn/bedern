// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
)

// TestMarksKeepTheirFloor: in every pair of a letter with a mark and a kerned
// glyph, in both orders and kerned as the font does it, the marks keep at
// least Acc.MarkFloor from the other glyph, measured on the outlines; the
// distance of the computation may be 1.5 too large, and the test allows 2.
// The pairs whose bound already keeps the floor are not measured.
func TestMarksKeepTheirFloor(t *testing.T) {
	t.Parallel()
	for _, s := range allSets() {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			gs := glyphstest.Set(s.p)
			tb := computed(s.p)
			f := tb.Font
			floor := s.p.Acc.MarkFloor
			n := 0
			for _, d := range f.Glyphs {
				if d.Like < 0 || len(gs[d.Index].Marks) == 0 {
					continue
				}
				marks := glyphs.Glyph{Path: gs[d.Index].Marks, Advance: d.Advance}
				for _, o := range f.Glyphs {
					if o.Kind == Skip {
						continue
					}
					if k := tb.Kern(o.Index, d.Index); f.markBound(o, d, float64(k)) < floor {
						n++
						if dist := exactDistance(gs[o.Index], marks, float64(k)); dist < floor-2 {
							t.Errorf("%s %s with kern %d: the mark is %.1f from %s, want at least %v", o.Name, d.Name, k, dist, o.Name, floor)
						}
					}
					if k := tb.Kern(d.Index, o.Index); f.markBound(d, o, float64(k)) < floor {
						n++
						if dist := exactDistance(marks, gs[o.Index], float64(k)); dist < floor-2 {
							t.Errorf("%s %s with kern %d: the mark is %.1f from %s, want at least %v", d.Name, o.Name, k, dist, o.Name, floor)
						}
					}
				}
			}
			if n == 0 && s.p == design.Regular() {
				t.Error("no pair brings a mark near its neighbour; the test sees nothing")
			}
		})
	}
}
