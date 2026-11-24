// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "testing"

// TestRBallClearsTheNext asks that the ball of r keeps at least 30 units
// from a following n, m, i or u without kerning.
func TestRBallClearsTheNext(t *testing.T) {
	r := glyph(t, "r")
	for _, name := range []string{"n", "m", "i", "u"} {
		if d := gapBetween(r.Path, glyph(t, name).Path, r.Advance); d < 30 {
			t.Errorf("r%s: the ball comes %.1f near the %s, want at least 30", name, d, name)
		}
	}
}

// TestRWedgeOpen asks that the white between the ball of r and its arm stays
// open at 8 pt: on some vertical line right of the stem the white between
// the arm above and the ball below is at least 8 units high.
func TestRWedgeOpen(t *testing.T) {
	p := Default()
	g := glyph(t, "r")
	_, _, x1, _ := bbox(g.Path)
	best := 0.0
	for x := italicX(p, p.SideStemLeft+p.Stem, p.XHeight) + 20.5; x < x1; x++ {
		var ys []float64
		for _, s := range crossings(g.Path, Point{X: x, Y: 0}, Point{Y: 1}) {
			if s > p.XHeight/2 {
				ys = append(ys, s)
			}
		}
		// Four crossings above the middle of the x-height: the ball
		// (bottom, top) and the arm (bottom, top).
		if len(ys) == 4 && ys[2]-ys[1] > best {
			best = ys[2] - ys[1]
		}
	}
	if best < 8 {
		t.Errorf("the wedge of r is at most %.1f high, want at least 8", best)
	}
}
