// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// checkRoundDiagonal checks a letter of the family of v, whose last stroke
// rises round a free side into an end at the upper right: the bottom
// overshoots the baseline as o does, the free side is a light rising
// stroke and not a second stem (0.4 to 0.75 stems horizontally at its node),
// the ink stays below the top of o, without a stem there, and the
// ball which ends the free side keeps at least 40 of white from the stroke
// before it at the height of its centre, as the curl of k keeps from its
// stem.
func checkRoundDiagonal(t *testing.T, name string, adv float64) {
	t.Helper()
	p := Default()
	g := glyph(t, name)
	_, y0, _, y1 := bbox(g.Path)
	if math.Abs(y0+p.Overshoot) > 1 {
		t.Errorf("%s: the bottom is at %.1f, want %v", name, y0, -p.Overshoot)
	}
	if y1 > p.XHeight+p.Overshoot+0.5 {
		t.Errorf("%s: the ink rises to %.1f, above the top of o", name, y1)
	}
	ys := p.VSideY * p.XHeight
	rs := runs(g.Path, ys)
	last := rs[len(rs)-1]
	if w := last[1] - last[0]; w < 0.4*p.Stem || w > 0.75*p.Stem {
		t.Errorf("%s: the free side is %.1f wide at %.0f, want 0.4 to 0.75 stems", name, w, ys)
	}
	if right := adv - p.VSideX + (ys-p.ShearHeight)*p.slant(); math.Abs(last[1]-right) > 2 {
		t.Errorf("%s: the free side reaches %.1f at %.0f, want %.1f", name, last[1], ys, right)
	}
	// At the height of the centre of the ball the run of the ball is the
	// last, and the run before it is the stroke on its left.
	yb := p.XHeight - p.VBallTop - p.Ball/2
	rs = runs(g.Path, yb)
	if len(rs) < 2 {
		t.Fatalf("%s: %d runs at %.0f", name, len(rs), yb)
	}
	if gap := rs[len(rs)-1][0] - rs[len(rs)-2][1]; gap < 40 {
		t.Errorf("%s: the ball is %.1f from the stroke before it at %.0f, want at least 40", name, gap, yb)
	}
}

// TestVForm checks the round v.
func TestVForm(t *testing.T) {
	checkRoundDiagonal(t, "v", Default().VWidth)
}
