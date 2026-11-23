// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// aTailGlyph returns the unrounded a.
func aTailGlyph(t *testing.T) (design.Params, outline.Path, float64) {
	t.Helper()
	p := design.Regular()
	path, adv := aGlyph(t, p)
	if path[0].SignedArea() <= 0 || path[1].SignedArea() >= 0 {
		t.Error("the outer contour of the tailed a is not counter-clockwise or its counter not clockwise")
	}
	return p, path, adv
}

// The tail: the low point OvershootLower below the baseline, near the stem, and
// a foot which is a curve and not a serif.
func TestATailShape(t *testing.T) {
	t.Parallel()
	p, path, _ := aTailGlyph(t)
	xr := design.AStemLeft(p) + p.Stem
	var low outline.Point
	for x := xr - 10; x < xr+p.ATailReach; x += 0.25 {
		if ys := path.CrossingsY(x); len(ys) > 0 && ys[0] < low.Y {
			low = outline.Point{X: x, Y: ys[0]}
		}
	}
	near(t, "low point", low.Y, -p.OvershootLower)
	if low.X < xr+10 || low.X > xr+60 {
		t.Errorf("the low point of the tail is at %.1f, the stem ends at %.1f", low.X, xr)
	}

	// There is no foot serif: the foot is a curve, between the stem and the tip
	// the tail leaves a gap under its upper edge at the height of the inner low
	// point, and the ink at height 90 reaches the tip.
	xs := path.CrossingsX(p.ATailTipHeight - 20)
	if got := xs[len(xs)-1] - xr; got < p.ATailReach-30 {
		t.Errorf("at height %v the ink reaches %.1f beyond the stem, want the tail, about %v", p.ATailTipHeight-20, got, p.ATailReach)
	}
}

// The tail keeps its distance from every following letter.
func TestATailKeepsAwayFromWhatFollows(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	a, adv, _ := outlineOf("a", p)
	for _, name := range followers {
		next, _, ok := outlineOf(name, p)
		if !ok {
			continue
		}
		d, q, r := minDistanceBetween(a, shiftedPath(next, adv))
		t.Logf("a (tail) followed by %s: nearest approach %.1f, at (%.0f, %.0f) and (%.0f, %.0f)", name, d, q.X, q.Y, r.X, r.Y)
		if d < pairClearance {
			t.Errorf("a with a tail followed by %s: nearest approach %.1f, want at least %v", name, d, pairClearance)
		}
	}
}

// The tail of a is the rightmost ink and overhangs the advance by -ATailSideRight,
// which is at most MaxOverhang, in the regular and in the bold parameters.
func TestATailOverhangsTheAdvance(t *testing.T) {
	t.Parallel()
	for _, p := range []design.Params{design.Regular(), design.Bold()} {
		path, adv := aGlyph(t, p)
		_, _, x1, _ := path.BBox()
		if p.ATailSideRight >= 0 || -p.ATailSideRight > p.MaxOverhang {
			t.Fatalf("ATailSideRight = %v, want from -%v to 0 exclusive", p.ATailSideRight, p.MaxOverhang)
		}
		tight(t, "overhang of the tail", x1-adv, -p.ATailSideRight)
	}
}
