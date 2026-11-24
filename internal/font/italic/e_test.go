// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestEStress asks that the stress of e, the line through the thinnest
// places at its top and its bottom, leans 7 to 11 degrees forward in the
// italic, as that of o, with the thinnest place at the top left of the
// topmost point.  The bar crosses every ray from the middle of e; the rays at the top start in the middle of the eye and those at the
// bottom halfway under the bar.
func TestEStress(t *testing.T) {
	p := Default()
	g := glyph(t, "e")
	cx := p.SideRoundLeft + p.EBowlWidth/2
	_, qt := thinnest(g.Path, italicPoint(p, pt(cx, (p.EBar+p.XHeight)/2)), 30, 150)
	_, qb := thinnest(g.Path, italicPoint(p, pt(cx, p.EBar/2)), 220, 300)
	lean := math.Atan2(qt.X-qb.X, qt.Y-qb.Y) * 180 / math.Pi
	t.Logf("stress of e %.1f degrees", lean)
	if lean < 7 || lean > 11 {
		t.Errorf("the stress of e leans %.1f degrees, want 7 to 11", lean)
	}
	_, _, _, y1 := bbox(g.Path)
	l, r := inkRange(g.Path, y1-1, y1)
	if qt.X >= (l+r)/2 {
		t.Errorf("the thinnest place at the top of e is at %.0f, not left of the topmost point at %.0f", qt.X, (l+r)/2)
	}
}

// TestEEye asks that the bar of e runs through its own stroke without a
// stub, that e is one outer contour and the eye, and that the eye is
// round and open at 8 pt: at least 120 high and 150 wide inside, and the
// bar 28 to 34 thick.
func TestEEye(t *testing.T) {
	p := Default()
	g := glyph(t, "e")
	if len(g.Path) != 2 {
		t.Fatalf("e has %d contours, want the outer one and the eye", len(g.Path))
	}
	up := unsheared(p, g.Path)
	// The eye is the contour of the smaller box.
	eye := up[1:]
	if a, b := up[0].SignedArea(), up[1].SignedArea(); math.Abs(a) < math.Abs(b) {
		eye = up[:1]
	}
	x0, y0, x1, y1 := bbox(eye)
	if h, w := y1-y0, x1-x0; h < 120 || w < 150 {
		t.Errorf("the eye of e is %.0f high and %.0f wide inside, want at least 120 and 150", h, w)
	}
	// The bar, on the vertical line through the middle of the eye.
	ss := crossings(up, Point{X: (x0 + x1) / 2, Y: -50}, Point{Y: 1})
	if len(ss) != 6 {
		t.Fatalf("the vertical through the eye crosses e %d times, want 6", len(ss))
	}
	if w := ss[3] - ss[2]; w < 28 || w > 34 {
		t.Errorf("the bar of e is %.1f thick, want 28 to 34", w)
	}
}

// TestEAperture asks that the white under the bar of e, between the bar and
// the end of the lower stroke, is at least 90 high, and that the end of the lower stroke is an exit, its highest
// ink 90 to 120 above the baseline.
func TestEAperture(t *testing.T) {
	p := Default()
	up := unsheared(p, glyph(t, "e").Path)
	_, tipRight := inkRange(up, 0, p.XHeight/2-40)
	x := tipRight - 30
	ss := crossings(up, Point{X: x, Y: -50}, Point{Y: 1})
	if len(ss) < 4 {
		t.Fatalf("the vertical line at %.0f crosses e %d times, want 4", x, len(ss))
	}
	cx := p.SideRoundLeft + p.EBowlWidth/2
	high := 0.0
	for y := 0.0; y < p.EBar-50; y++ {
		for _, r := range runs(up, y+0.5) {
			if r[0] > cx {
				high = y + 1
			}
		}
	}
	if high < 90 || high > 120 {
		t.Errorf("the end of the lower stroke of e reaches %.0f, want 90 to 120", high)
	}
	if gap := ss[2] - ss[1]; gap < 90 {
		t.Errorf("the aperture of e is %.0f high, want at least 90", gap)
	}
}
