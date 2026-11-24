// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestGTail asks of the open tail of g what checkTail asks of every tail,
// except that it reaches under the bowl: its leftmost ink lies within 0.06
// of the x-height of the leftmost ink of the bowl; and that the stem is the stem of the italic down to the turn.
func TestGTail(t *testing.T) {
	p := Default()
	e, _ := bowlEdge(p, "g")
	g := glyph(t, "g")
	up := unsheared(p, g.Path)
	bowlLeft, _ := inkRange(up, p.Overshoot, p.XHeight-p.Overshoot)
	d := 0.06 * p.XHeight
	checkTailReach(t, p, "g", e, e-bowlLeft-d, e-bowlLeft+d)
	for y := -p.GTailTurn + 4.5; y < p.XHeight-p.BowlTopJoin-30; y += 16 {
		if y > -10 && y < p.BowlPlainJoin+40 {
			continue
		}
		r := stemRun(t, p, g.Path, e+p.Stem/2, y)
		if w := r[1] - r[0]; math.Abs(w-p.Stem) > 1 {
			t.Errorf("the stem of g is %.1f wide at %.0f, want %v", w, y, p.Stem)
		}
	}
}

// checkTail asks of the tail of the letter name under the stem whose left
// edge stands at e, the tail of f, g, j and y: its lowest ink lies TailDepth
// below the baseline; its leftmost ink, upright, lies 130 to 170 left of
// the left edge of the stem; the ball
// stands at least 40 clear of the stem; and the white between the ball and the stroke above it
// stays open, at least 8 high on a vertical line.
func checkTail(t *testing.T, p Params, name string, e float64) {
	t.Helper()
	checkTailReach(t, p, name, e, 130, 170)
}

// checkTailReach is checkTail with the leftmost ink of the tail lo to hi
// left of the left edge of the stem.
func checkTailReach(t *testing.T, p Params, name string, e, lo, hi float64) {
	t.Helper()
	up := unsheared(p, glyph(t, name).Path)
	_, y0, _, _ := bbox(up)
	if math.Abs(-y0-p.TailDepth) > 1 {
		t.Errorf("%s: the tail reaches %.1f below the baseline, want %v", name, -y0, p.TailDepth)
	}
	left, _ := inkRange(up, -p.TailDepth, -p.Overshoot-10)
	if r := e - left; r < lo || r > hi {
		t.Errorf("%s: the tail reaches %.0f left of the stem, want %.0f to %.0f", name, r, lo, hi)
	}
	// The ball: the ink left of the stem between the lowest stroke and the
	// baseline, apart from the stem by the white of each line.
	clear := math.Inf(1)
	for y := -p.TailDepth + p.BallStalk + 5; y < -p.Overshoot-10; y += 2 {
		rs := runs(up, y+0.5)
		if len(rs) == 2 {
			clear = math.Min(clear, rs[1][0]-rs[0][1])
		}
	}
	if clear < 40 {
		t.Errorf("%s: the ball comes %.0f near the stem, want at least 40", name, clear)
	}
	best := 0.0
	for x := left; x < e; x += 2 {
		var ys []float64
		for _, s := range crossings(up, Point{X: x, Y: -p.TailDepth - 10}, Point{Y: 1}) {
			if y := s - p.TailDepth - 10; y < -p.Overshoot-5 {
				ys = append(ys, y)
			}
		}
		if len(ys) == 4 && ys[2]-ys[1] > best {
			best = ys[2] - ys[1]
		}
	}
	if best < 8 {
		t.Errorf("%s: the wedge of the tail is at most %.1f high, want at least 8", name, best)
	}
}
