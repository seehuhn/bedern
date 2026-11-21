// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of r check the stem, the head serif and the left foot of n, a right
// foot of 80, an arm which leaves the stem like the arch of n and ends in a
// ball which hangs below its end, a reach of the arm from the stem to the
// rightmost ink, and the advance RAdvance.

type rFrame struct {
	p       design.Params
	path    outline.Path
	advance float64
}

func newRFrame(t *testing.T, p design.Params) *rFrame {
	t.Helper()
	path, advance, ok := outlineOf("r", p)
	if !ok {
		t.Fatal("glyph r not found")
	}
	return &rFrame{p, path, advance}
}

func (f *rFrame) stemRight() float64 { return f.p.RSideLeft + f.p.Stem }

func TestRAdvanceAndStem(t *testing.T) {
	t.Parallel()
	f := newRFrame(t, design.Regular())
	p := f.p
	if f.advance != design.RAdvance(p) {
		t.Errorf("advance %v, want %v", f.advance, design.RAdvance(p))
	}
	if ratio := f.advance / design.NAdvance(p); ratio < 0.628 || ratio > 0.727 {
		t.Errorf("r is %.3f of n wide, want 0.628 to 0.727", ratio)
	}
	xs := f.path.CrossingsX(100)
	if len(xs) != 2 {
		t.Fatalf("crossings at height 100: %v", xs)
	}
	near(t, "left edge of the stem", xs[0], p.RSideLeft)
	near(t, "stem width", xs[1]-xs[0], p.Stem)
	if len(f.path) != 1 {
		t.Errorf("%d contours, want 1", len(f.path))
	}
	// The ink ends RReach right of the stem, and RSideRight from the advance.
	_, _, x1, _ := f.path.BBox()
	near(t, "reach of the arm beyond the stem", x1-f.stemRight(), p.RReach)
	near(t, "space right of the ball", f.advance-x1, p.RSideRight)
}

func TestRSerifs(t *testing.T) {
	t.Parallel()
	f := newRFrame(t, design.Regular())
	p := f.p
	l, r := p.RSideLeft, f.stemRight()
	checkFoot(t, p, f.path, l, false, p.SerifJut, p.SerifThick)
	checkFoot(t, p, f.path, r, true, p.RFootRight, p.SerifThick)
	if p.RFootRight <= p.SerifJut {
		t.Errorf("the right foot %v is not longer than the left one %v", p.RFootRight, p.SerifJut)
	}
	checkHeadSerif(t, p, f.path, l, r, p.XHeight)
}

// TestRBallKeepsAwayFromTheNextHeadSerif checks that the ball of r is at least
// 38 from the head serif of a following n.  The distance is that of the
// flattened outlines, the outline of n shifted by the advance of r, from the ink
// of r right of its stem to the head serif of n left of its stem.
func TestRBallKeepsAwayFromTheNextHeadSerif(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	rp, adv, _ := outlineOf("r", p)
	np, _, _ := outlineOf("n", p)
	best := math.Inf(1)
	for _, a := range rp[0].Flatten() {
		if a.X < p.RSideLeft+p.Stem+20 {
			continue
		}
		for _, b := range np[0].Flatten() {
			if b.X > p.SideStemLeft+p.Stem {
				continue
			}
			best = math.Min(best, math.Hypot(a.X-(b.X+adv), a.Y-b.Y))
		}
	}
	if best < 38 {
		t.Errorf("the ball of r is %.1f from the head serif of a following n, want at least 38", best)
	}
}
