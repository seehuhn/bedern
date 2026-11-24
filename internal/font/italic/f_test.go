// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestFHookAndTail asks of f: the tail is the tail of every tail
// (checkTail); the crest of the hook stands at the top of the entry of an
// ascender, 704; the hook reaches as far right of the right edge of the
// stem, upright, as the tail reaches left of the left edge; both overhang
// the set width in the italic by at most maxOverhang; and the crossbar,
// which the union joins to the stem, has its top on the x-height and reaches
// FBarLeft and FBarRight beyond the edges of the stem at its middle (its
// ends are cut vertically in the italic).
func TestFHookAndTail(t *testing.T) {
	p := Default()
	e := p.FSideLeft
	checkTail(t, p, "f", e)
	g := glyph(t, "f")
	x0, _, x1, y1 := bbox(g.Path)
	if top := p.XHeight + p.AscenderEntryRise + p.Overshoot; math.Abs(y1-top) > 1 {
		t.Errorf("the crest of f stands at %.1f, want %v", y1, top)
	}
	if l, r := -x0, x1-g.Advance; l > maxOverhang || r > maxOverhang {
		t.Errorf("f overhangs %.0f before the origin and %.0f beyond the set width, want at most %v", l, r, maxOverhang)
	}
	up := unsheared(p, g.Path)
	left, _ := inkRange(up, -p.TailDepth, -p.Overshoot-10)
	_, right := inkRange(up, p.XHeight+40, y1)
	if d := (right - e - p.Stem) - (e - left); math.Abs(d) > 2 {
		t.Errorf("the hook of f reaches %.0f right of the stem and the tail %.0f left, want equal", right-e-p.Stem, e-left)
	}
	rs := runs(up, p.XHeight-p.Hair/2)
	if len(rs) != 1 {
		t.Fatalf("the crossbar of f is %d runs at its middle, want one", len(rs))
	}
	if l, r := e-rs[0][0], rs[0][1]-e-p.Stem; math.Abs(l-p.FBarLeft) > 1 || math.Abs(r-p.FBarRight) > 1 {
		t.Errorf("the crossbar of f reaches %.1f and %.1f beyond the stem, want %v and %v", l, r, p.FBarLeft, p.FBarRight)
	}
	if above := runs(up, p.XHeight+0.5); len(above) != 1 || above[0][1]-above[0][0] > p.Stem+1 {
		t.Errorf("above the x-height f is %v, want the stem alone", above)
	}
}
