// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The conditions for k: the stem of h with its head serif
// on the left and both feet; a thin arm of 36 at 47 degrees which runs from the
// stem to the upper right and ends in a serif at the x-height; a thick leg of 76 at
// 35 degrees which leaves the arm (not the stem) at 0.60 of the x-height, 25 to 108
// from the stem edge, and ends in a foot serif on the baseline; serifs 40 outside,
// 85 inside the arm, 64 inside the leg; the advance 512.

func kGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, adv, ok := outlineOf("k", p)
	if !ok {
		t.Fatal("glyph k not found")
	}
	if len(path) != 1 {
		t.Fatalf("k has %d contours, want 1", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Error("the contour of k is not counter-clockwise")
	}
	return path, adv
}

// kStem returns the edges of the stem of k, which is that of h.
func kStem(p design.Params) (left, right float64) {
	left = p.SideStemLeft + design.AscenderInset(p)
	return left, left + p.AscenderStem
}

func TestKAdvanceAndSides(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := kGlyph(t, p)
	tight(t, "advance", adv, 509)
	tight(t, "advance from side spaces and body", adv, p.SideStemLeft+p.Stem+p.KReach+p.KSideRight)
	if r := adv / design.NAdvance(p); r < 0.847 || r > 1.0 {
		t.Errorf("k is %.3f of n wide, want 0.847 to 1.000", r)
	}
	hPath, hAdv, _ := outlineOf("h", p)
	_ = hPath
	if r := adv / hAdv; r < 0.90 || r > 1.0 {
		t.Errorf("k is %.3f of h wide, want about 0.95", r)
	}
	x0, y0, x1, y1 := path.BBox()
	tight(t, "left end of the ink, the foot serif of the stem", x0, p.SideStemLeft-p.SerifJut)
	tight(t, "right space", adv-x1, p.KSideRight)
	tight(t, "top is the ascender", y1, p.Ascender)
	tight(t, "bottom is the baseline", y0, 0)
	if e, ok := lookup("k"); !ok || e.left != sideStem || e.right == sideUnknown {
		t.Error("k has no side kinds, or its left side is not that of h")
	}
}

// The stem is that of h: the head serif and the feet, and the stem as wide as the
// ascender stem.
func TestKStemIsThatOfH(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := kGlyph(t, p)
	left, right := kStem(p)
	checkHeadSerif(t, p, path, left, right, p.Ascender)
	checkFoot(t, p, path, left, false, p.SerifJut, p.SerifThick)
	checkFoot(t, p, path, right, true, p.SerifJutInner, p.SerifThick)
	xs := path.CrossingsX(p.XHeight + 100)
	if len(xs) != 2 {
		t.Fatalf("crossings above the arm: %v", xs)
	}
	tight(t, "stem width", xs[1]-xs[0], p.AscenderStem)
	tight(t, "left edge of the stem", xs[0], left)
	h, _, _ := outlineOf("h", p)
	hx := h.CrossingsX(p.XHeight + 100)
	tight(t, "the stem stands where that of h stands", xs[0], hx[0])
	tight(t, "left edge of the stem below the arm", path.CrossingsX(100)[0], left)
}

func TestKArmAndLeg(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := kGlyph(t, p)
	_, right := kStem(p)
	arm := readDiagonals(t, path, 290, 355, 2)[1]
	leg := readDiagonals(t, path, 90, 190, 2)[1]
	tight(t, "arm thickness", arm.thick(), p.Hair)
	tight(t, "arm is 36", arm.thick(), 36)
	tight(t, "arm angle", arm.angle(), 47)
	tight(t, "leg thickness", leg.thick(), p.KLeg)
	tight(t, "leg is 76", leg.thick(), 76)
	tight(t, "leg angle", leg.angle(), 35)
	if arm.lean <= 0 || leg.lean >= 0 {
		t.Errorf("the arm leans %.3f and the leg %.3f, want the arm up to the right and the leg down to the right", arm.lean, leg.lean)
	}
	if r := arm.thick() / 32; r < 0.99 || r > 1.46 {
		t.Errorf("arm over the thin part of o %.2f, want 0.99 to 1.46", r)
	}
	if r := leg.thick() / p.Stem; r < 0.91 || r > 1.02 {
		t.Errorf("leg over the stem %.2f, want 0.91 to 1.02", r)
	}
	// The centre lines of the arm and of the leg meet at the junction height, at
	// the distance from the stem which design.KJunctionX gives.
	yJ := 290 + (leg.centreAt(90, 290)-arm.centreAt(290, 290))/(arm.lean-leg.lean)
	tight(t, "height of the junction", yJ, p.KJunction)
	tight(t, "height of the junction is 269", yJ, 269)
	xJ := arm.centreAt(290, yJ)
	tight(t, "position of the junction", xJ, design.KJunctionX(p))
	if d := xJ - right; d < 25 || d > 108 {
		t.Errorf("the junction stands %.1f from the stem, want 25 to 108", d)
	}
	if r := yJ / p.XHeight; r < 0.552 || r > 0.639 {
		t.Errorf("the junction is at %.3f of the x-height, want 0.552 to 0.639", r)
	}
	// The arm meets the stem where its centre line reaches the edge of the stem.
	yS := 290 + (right-arm.centreAt(290, 290))/arm.lean
	if r := yS / p.XHeight; r < 0.366 || r > 0.540 {
		t.Errorf("the arm meets the stem at %.3f of the x-height, want 0.366 to 0.540", r)
	}
	for _, span := range [][2]float64{{290, 355}, {90, 190}} {
		for _, e := range edgeSamples2(path, span[0], span[1], 12) {
			if e > 1e-6 {
				t.Errorf("an edge of k deviates from a straight line by %v between %v and %v", e, span[0], span[1])
			}
		}
	}
}

// The junction: the leg leaves the arm and not the stem.  Between the stem and the
// leg there is white all the way from the baseline to the underside of the arm, and
// the outline has a sharp inner corner where the underside of the arm meets the
// edge of the leg, one where the right-hand edge of the leg meets the arm, one
// where the arm leaves the stem below and one above.
func TestKJunctionGeometry(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := kGlyph(t, p)
	_, right := kStem(p)
	for _, y := range []float64{60, 100, 150, 190} {
		xs := path.CrossingsX(y)
		if len(xs) != 4 {
			t.Fatalf("crossings at %v: %v, want the stem and the leg", y, xs)
		}
		if gap := xs[2] - xs[1]; gap < 40 {
			t.Errorf("at height %v the white between the stem and the leg is only %.1f wide", y, gap)
		}
	}
	// The corners.
	arm := readDiagonals(t, path, 290, 355, 2)[1]
	leg := readDiagonals(t, path, 90, 190, 2)[1]
	line := func(x, y0, lean float64) [2]float64 { return [2]float64{x - lean*y0, lean} }
	meet := func(a, b [2]float64) outline.Point {
		y := (b[0] - a[0]) / (a[1] - b[1])
		return outline.Point{X: a[0] + a[1]*y, Y: y}
	}
	armLower := line(arm.right, 290, arm.lean)
	armUpper := line(arm.left, 290, arm.lean)
	legLeft := line(leg.left, 90, leg.lean)
	legRight := line(leg.right, 90, leg.lean)
	stem := [2]float64{right, 0}
	corners := []struct {
		name   string
		q      outline.Point
		convex bool
	}{
		{"arm leaves the stem below", meet(stem, armLower), false},
		{"underside of the arm meets the leg", meet(armLower, legLeft), false},
		{"leg meets the arm on the right", meet(armLower, legRight), false},
		{"arm meets the stem above", meet(stem, armUpper), false},
	}
	ct := path[0]
	nodes := contourNodes(ct)
	n := len(ct.Segs)
	for _, c := range corners {
		i := nodeNear(t, ct, c.q)
		if ct.Segs[(i+n-1)%n].IsCurve || ct.Segs[i%n].IsCurve {
			t.Errorf("%s: a curve next to the corner", c.name)
		}
		a, b, d := nodes[(i+n-1)%n], nodes[i], nodes[i+1]
		if cross := (b.X-a.X)*(d.Y-b.Y) - (b.Y-a.Y)*(d.X-b.X); cross >= 0 {
			t.Errorf("%s: the corner is not an inner corner (turn %.1f)", c.name, cross)
		}
	}
	// The outer corner where the arm leaves the stem is below the underside of the
	// arm: the underside of the arm meets the stem below the junction.
	yStem := meet(stem, armLower).Y
	if yStem >= p.KJunction {
		t.Errorf("the arm leaves the stem at %.1f, not below the junction at %v", yStem, p.KJunction)
	}
	// The ink between the arm and the leg on the right of the junction is one
	// notch: the leg's right-hand edge meets the underside of the arm above the
	// meeting of its left-hand edge.
	if meet(armLower, legRight).Y <= meet(armLower, legLeft).Y {
		t.Error("the right-hand edge of the leg meets the arm below its left-hand edge")
	}
}

// The serifs: arm 40 outside and 85 inside, leg 40 outside and 64 inside, 33 thick
// at the tip.
func TestKSerifs(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := kGlyph(t, p)
	xb := path.CrossingsX(footScanY)
	if len(xb) != 4 {
		t.Fatalf("crossings just above the baseline: %v, want 4", xb)
	}
	// The thickness at the tip of the foot of the leg.
	for _, tip := range []struct{ x, side float64 }{{xb[2], 1}, {xb[3], -1}} {
		ys := path.CrossingsY(tip.x + tip.side*tipInset)
		tight(t, "foot serif bottom", ys[0], 0)
		if th := ys[1] - ys[0]; th < p.SerifThick-0.01 || th > p.SerifThick+1.5 {
			t.Errorf("foot serif thickness five units from the tip %.2f, want about %v", th, p.SerifThick)
		}
	}
	// The foot of the leg and the right foot of the stem leave room.
	_, right := kStem(p)
	if gap := xb[2] - (right + p.SerifJutInner); gap < 3*design.ClearanceFloor(p) {
		t.Errorf("the foot of the leg is %.1f from the right foot of the stem, want at least %v", gap, 3*design.ClearanceFloor(p))
	}
}
