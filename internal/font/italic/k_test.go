// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestKArm asks that the arm and the leg of k leave the stem together and
// open away from it: the white between the stem and
// the arm is at least 25 wide at 0.75 of the x-height and grows upwards,
// and the white between the arm and the leg, on the vertical line 40
// right of the stem, is at least 25 high.  The arm ends below the x-height.
func TestKArm(t *testing.T) {
	p := Default()
	g := glyph(t, "k")
	e := p.SideStemLeft - p.HLeftLess
	prev := 0.0
	for _, f := range []float64{0.75, 0.85, 0.95} {
		y := f * p.XHeight
		rs := runs(g.Path, y)
		if len(rs) < 2 {
			t.Fatalf("k: %d runs at %.0f, want the stem and the arm", len(rs), y)
		}
		gap := rs[1][0] - rs[0][1]
		if gap < 25 || gap <= prev {
			t.Errorf("k: the arm is %.1f from the stem at %.0f, want at least 25 and more than %.1f below", gap, y, prev)
		}
		prev = gap
	}
	if rs := runs(g.Path, p.XHeight+1); len(rs) > 1 {
		t.Errorf("k: the arm reaches above the x-height")
	}
	x := italicX(p, e+p.Stem+40, p.XHeight/2)
	ys := crossings(g.Path, Point{X: x, Y: 0}, Point{Y: 1})
	best := 0.0
	for i := 1; i+1 < len(ys); i += 2 {
		if ys[i] > p.XHeight/4 && ys[i+1] < p.XHeight && ys[i+1]-ys[i] > best {
			best = ys[i+1] - ys[i]
		}
	}
	if best < 25 {
		t.Errorf("k: the white between the arm and the leg is at most %.1f high at %.0f, want at least 25 (crossings %v)", best, x, ys)
	}
}

// TestKArmReach asks that the arm of k reach: its largest
// extent to the right of the right edge of the stem, between 0.6 and 1.02
// of the x-height and taken on the upright description, is 0.54 to 0.66 of
// the x-height and 3.3 to 4.2 stems.
func TestKArmReach(t *testing.T) {
	p := Default()
	g := glyph(t, "k")
	e := p.SideStemLeft - p.HLeftLess
	reach := 0.0
	for f := 0.6; f <= 1.02; f += 0.005 {
		y := f * p.XHeight
		rs := runs(g.Path, y)
		if len(rs) < 2 {
			continue
		}
		x := rs[len(rs)-1][1] - (y-p.ShearHeight)*p.slant()
		reach = math.Max(reach, x-(e+p.Stem))
	}
	if r := reach / p.XHeight; r < 0.54 || r > 0.66 {
		t.Errorf("k: the arm reaches %.0f beyond the stem, %.2f of the x-height, want 0.54 to 0.66", reach, r)
	}
	if r := reach / p.Stem; r < 3.3 || r > 4.2 {
		t.Errorf("k: the arm reaches %.0f beyond the stem, %.2f stems, want 3.3 to 4.2", reach, r)
	}
}
