// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestDIsAWithAscender asks that a and d differ only above the x-height: at
// every height from the baseline to 20 below the x-height the runs of ink
// agree within half a unit, and the stem of d runs straight at the weight
// of the stem from the x-height to the entry.
func TestDIsAWithAscender(t *testing.T) {
	p := Default()
	a, d := glyph(t, "a"), glyph(t, "d")
	if a.Advance != d.Advance {
		t.Errorf("a is %v wide and d %v, want one width", a.Advance, d.Advance)
	}
	for y := -p.Overshoot + 0.5; y < p.XHeight-20; y += 4 {
		ra, rd := runs(a.Path, y), runs(d.Path, y)
		same := len(ra) == len(rd)
		for i := 0; same && i < len(ra); i++ {
			same = math.Abs(ra[i][0]-rd[i][0]) < 0.5 && math.Abs(ra[i][1]-rd[i][1]) < 0.5
		}
		if !same {
			t.Errorf("a and d differ at %.1f: %v against %v", y, ra, rd)
			break
		}
	}
	c := p.bowlStemEdge() + p.Stem/2
	for y := p.XHeight + 4.5; y < p.XHeight+p.AscenderEntryRise-p.EntryJoin; y += 8 {
		r := stemRun(t, p, d.Path, c, y)
		if w := r[1] - r[0]; math.Abs(w-p.Stem) > 1 {
			t.Errorf("the stem of d is %.1f wide at %.0f, want %v", w, y, p.Stem)
		}
	}
}
