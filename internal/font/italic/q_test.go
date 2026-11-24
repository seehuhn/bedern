// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestQFoot asks that q stands on the foot of p: from the descender line to 40 above the serif the ink of
// q under its stem is that of p moved by the distance between the stems,
// within a unit, and the stem of q above the serif is the stem of the
// italic.
func TestQFoot(t *testing.T) {
	p := Default()
	gp, gq := glyph(t, "p"), glyph(t, "q")
	dx := p.bowlStemEdge() - p.SideStemLeft
	th := p.FootThick * p.Hair
	for y := -p.Descender + 0.5; y < -p.Descender+th+40; y += 2 {
		rp, rq := runs(gp.Path, y), runs(gq.Path, y)
		if len(rp) != 1 || len(rq) != 1 {
			t.Fatalf("at %.1f p has %d runs and q %d, want one each", y, len(rp), len(rq))
		}
		if math.Abs(rq[0][0]-rp[0][0]-dx) > 1 || math.Abs(rq[0][1]-rp[0][1]-dx) > 1 {
			t.Errorf("at %.1f the foot of q is %v, that of p %v moved by %v", y, rq[0], rp[0], dx)
		}
	}
	e := p.bowlStemEdge()
	for y := -p.Descender + th + 4.5; y < -20; y += 8 {
		run := stemRun(t, p, gq.Path, e+p.Stem/2, y)
		if w := run[1] - run[0]; math.Abs(w-p.Stem) > 1 {
			t.Errorf("the stem of q is %.1f wide at %.0f, want %v", w, y, p.Stem)
		}
	}
}
