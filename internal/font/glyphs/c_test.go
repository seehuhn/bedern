// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter c.  The tests measure the outline and compare it with the
// parameters: the left side, the top and the bottom are the curves of the o as e
// has them, the lower terminal is that of e, and the upper stroke ends in a
// ball like those of a and r.

// cWithin fails the test if got is not within eps of want.
func cWithin(t *testing.T, what string, got, want, eps float64) {
	t.Helper()
	if math.Abs(got-want) > eps {
		t.Errorf("%s: %.2f, want %.2f within %.2f", what, got, want, eps)
	}
}

func cGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, advance, ok := outlineOf("c", p)
	if !ok {
		t.Fatal("glyph c not found")
	}
	if len(path) != 1 {
		t.Fatalf("c has %d contours, want one and no counter", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Error("the contour of c is not counter-clockwise")
	}
	return path, advance
}

// cTop returns the position of the top extreme of c.
func cTop(path outline.Path) (x, y float64) {
	var top []outline.Point
	_, _, x1, _ := path.BBox()
	for xx := 0.0; xx <= x1; xx += 0.5 {
		if ys := path.CrossingsY(xx); len(ys) > 0 {
			top = append(top, outline.Point{X: xx, Y: ys[len(ys)-1]})
		}
	}
	return peakOf(top)
}

func TestCAdvanceAndSides(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := cGlyph(t, p)
	// The advance is the left side space of c, the ink and the open right side
	// space of e: 31 + 388 + 17 = 436.
	if want := p.CSideLeft + p.CBowlWidth + p.SideOpenRight; adv != want || adv != design.CAdvance(p) {
		t.Errorf("advance %v, want %v", adv, want)
	}
	if adv != 436 {
		t.Errorf("advance %v, want 436", adv)
	}
	ent, ok := lookup("c")
	if !ok || ent.r != 'c' || ent.left != sideBowlC || ent.right != sideOpen {
		t.Errorf("table entry of c: %+v, want rune c with a round left side and an open right side", ent)
	}
	x0, _, x1, _ := path.BBox()
	cWithin(t, "left side space", x0, p.CSideLeft, 0.01)
	cWithin(t, "right side space", adv-x1, p.SideOpenRight, 0.01)
	cWithin(t, "ink width", x1-x0, p.CBowlWidth, 0.01)
	cWithin(t, "ink width, absolute", x1-x0, 388, 0.01)
	// The advance is 0.727 to 0.888 of that of n.
	if r := adv / design.NAdvance(p); r < 0.727 || r > 0.888 {
		t.Errorf("c is %.3f of n wide, want 0.727 to 0.888", r)
	}
}

// TestCIsNotWiderThanE checks that e lies between c and o in width.
func TestCIsNotWiderThanE(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := cGlyph(t, p)
	e, eAdv := eGlyph(t, p)
	o, oAdv, _ := outlineOf("o", p)
	cx0, _, cx1, _ := path.BBox()
	ex0, _, ex1, _ := e.BBox()
	ox0, _, ox1, _ := o.BBox()
	cw, ew, ow := cx1-cx0, ex1-ex0, ox1-ox0
	if !(cw < ew && ew < ow) {
		t.Errorf("ink widths c %.1f, e %.1f, o %.1f: e must lie between c and o", cw, ew, ow)
	}
	if !(adv <= eAdv && eAdv < oAdv) {
		t.Errorf("advances c %v, e %v, o %v: e must lie between c and o", adv, eAdv, oAdv)
	}
	// The ink is 0.912 to 1.015 of that of e and 0.803 to 0.943 of that of o.
	if r := cw / ew; r < 0.912 || r > 1.015 {
		t.Errorf("c is %.3f of the ink width of e, want 0.912 to 1.015", r)
	}
	if r := cw / ow; r < 0.803 || r > 0.943 {
		t.Errorf("c is %.3f of the ink width of o, want 0.803 to 0.943", r)
	}
}

// The left side, the top and the bottom are the curves of the o as e has them:
// the thickest part is 94 at the left extreme, at EThickPos of the height, the
// thin parts at the top and the bottom are 32, the overshoots 12, and the
// squareness is that of the o at the arc midpoints of both edges.
func TestCFollowsTheCurvesOfO(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := cGlyph(t, p)
	bottom, top := -p.OvershootLower, p.XHeight+p.OvershootLower
	thickY := bottom + p.EThickPos*(top-bottom)

	// Thickest part: the left extreme is at EThickPos and the stroke is 94 thick
	// there and nowhere thicker on the left side.
	xs := path.CrossingsX(thickY)
	cWithin(t, "left extreme", xs[0], p.CSideLeft, 0.01)
	cWithin(t, "thickest part", xs[1]-xs[0], p.CurveLower, 0.05)
	cWithin(t, "thickest part, absolute", xs[1]-xs[0], 94, 0.05)
	for y := bottom + 40; y < top-40; y += 2.3 {
		if len(path.CrossingsX(y)) < 2 {
			continue
		}
		if x := path.CrossingsX(y); x[1]-x[0] > p.CurveLower+0.05 {
			t.Errorf("at height %.1f the left stroke is %.2f thick, more than %v", y, x[1]-x[0], p.CurveLower)
		}
	}

	// The top extreme, with the thin part and the overshoot, and the bottom.
	cx, topY := cTop(path)
	cWithin(t, "overshoot at the top", topY, top, 0.1)
	ys := path.CrossingsY(cx)
	if len(ys) != 4 {
		t.Fatalf("crossings above the middle of the bowl: %v, want the bottom and the top stroke", ys)
	}
	cWithin(t, "thin part at the bottom", ys[1]-ys[0], p.ThinLower, 0.1)
	cWithin(t, "overshoot at the bottom", ys[0], bottom, 0.1)
	cWithin(t, "thin part at the top", ys[3]-ys[2], p.ThinLower, 0.1)
	cWithin(t, "thin part at the top, absolute", ys[3]-ys[2], 32, 0.1)
	cWithin(t, "overshoot at the top, absolute", ys[3], 460, 0.1)
	cWithin(t, "overshoot at the bottom, absolute", ys[0], -12, 0.1)

	// Squareness: the quarter of a super-ellipse with superness s passes through
	// the point at s times the half axes from the centre on its diagonal.  The
	// outer edge has 0.72, the inner edge 0.745, in all four quarters of the left
	// half.
	left := p.CSideLeft
	a := cx - left
	check := func(what string, y, wantX float64, edge int) {
		t.Helper()
		xs := path.CrossingsX(y)
		cWithin(t, what, xs[edge], wantX, 0.35)
	}
	so, si := p.SupernessOuter, p.SupernessInner
	check("outer edge, upper left quarter", thickY+so*(top-thickY), cx-so*a, 0)
	check("outer edge, lower left quarter", thickY-so*(thickY-bottom), cx-so*a, 0)
	ai := cx - (left + p.CurveLower)
	ti, bi := top-p.ThinLower, bottom+p.ThinLower
	check("inner edge, upper left quarter", thickY+si*(ti-thickY), cx-si*ai, 1)
	check("inner edge, lower left quarter", thickY-si*(thickY-bi), cx-si*ai, 1)
}

// TestCClearance checks that the clearance between the ball and the inner curve
// and between the two terminals is at least the clearance floor, for the regular
// parameters.
func TestCClearance(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := cGlyph(t, p)
	d, a, b := path.MinDistance(design.SlitRatio)
	t.Logf("least clearance %.2f between (%.1f, %.1f) and (%.1f, %.1f)", d, a.X, a.Y, b.X, b.Y)
	if d < design.ClearanceFloor(p) {
		t.Errorf("clearance %.2f below the floor %v", d, design.ClearanceFloor(p))
	}
}

// The curvature of the inner edge of the upper stroke of c, from the o to the
// tip of the wedge, does not change faster than cKneeMax per unit of arc
// length: 0.0012 for the regular parameters and 0.003 for the other sets.
func TestCInnerEdgeHasNoKnee(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			cKneeMax := 0.003
			if np.name == "regular" {
				cKneeMax = 0.0012
			}
			f := newCFrame(np.params)
			tb := cTerminal(np.params, f)
			ch := cStalk(np.params, f)
			for _, pc := range ch[1:] {
				if r := curvatureRate(pc.cubic); r > cKneeMax {
					t.Errorf("the curvature of the inner edge changes by %.5f per unit, want at most %v", r, cKneeMax)
				}
				if isTip(pc.P3, tb.Tip) {
					break
				}
			}
		})
	}
}
