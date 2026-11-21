// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// dGlyph returns the unrounded outline of d.
func dGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, advance, ok := outlineOf("d", p)
	if !ok {
		t.Fatal("glyph d not found")
	}
	if len(path) != 2 {
		t.Fatalf("d has %d contours, want an outer contour and a counter", len(path))
	}
	return path, advance
}

func TestDAdvance(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	_, adv := dGlyph(t, p)
	// The advance is the left side space of the bowl, the body (bowl and stem) and
	// the right side space of a stem.
	if want := p.DQSideLeft + p.DBowl.Width + p.Stem + p.DSideRight; adv != want || adv != design.DAdvance(p) {
		t.Errorf("advance %v, want %v", adv, want)
	}
	// The advance is 546.
	if math.Abs(adv-546) > 0.5 {
		t.Errorf("advance %v, want 546", adv)
	}
	e, ok := lookup("d")
	if !ok || e.r != 'd' || e.left != sideBowlDQ || e.right != sideStemD {
		t.Errorf("table entry of d: %+v, want rune d with a round left side and a stem on the right", e)
	}
}

func TestDStem(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := dGlyph(t, p)
	stemLeft := design.DStemLeft(p)
	stemRight := stemLeft + p.Stem
	if path[0].SignedArea() <= 0 || path[1].SignedArea() >= 0 {
		t.Error("the outer contour is not counter-clockwise or the counter not clockwise")
	}

	x0, y0, x1, y1 := path.BBox()
	near(t, "left extreme", x0, p.DQSideLeft)
	near(t, "bottom of the bowl", y0, -p.OvershootLower)
	near(t, "ascender", y1, p.Ascender)
	near(t, "right end of the foot serif", x1, stemRight+p.SerifJut)
	if x1 > adv-p.DSideRight+p.SerifJut+tol {
		t.Errorf("ink reaches %.1f, beyond the stem right side space", x1)
	}

	// The stem has the stem width above the upper notch, between the joins
	// and, on its left-hand edge, below the lower notch; its edges are upright.
	for y := p.XHeight + p.OvershootLower + 10; y < p.Ascender-p.HeadSerifDrop-p.SerifThick-20; y += 40 {
		xs := path.CrossingsX(y)
		if len(xs) != 2 {
			t.Fatalf("crossings above the bowl at %.0f: %v", y, xs)
		}
		near(t, "left edge above the bowl", xs[0], stemLeft)
		near(t, "right edge above the bowl", xs[1], stemRight)
	}
	for y := p.DBowl.JoinLower + 20; y < p.DBowl.JoinUpper-20; y += 40 {
		xs := path.CrossingsX(y)
		if len(xs) != 4 {
			t.Fatalf("crossings between the joins at %.0f: %v", y, xs)
		}
		near(t, "left edge between the joins", xs[2], stemLeft)
		near(t, "right edge between the joins", xs[3], stemRight)
		near(t, "stem width between the joins", xs[3]-xs[2], p.Stem)
	}
	// Between the lower notch and the foot the left edge is the stem's.
	xs := path.CrossingsX(p.DBowl.NotchLower / 2)
	if len(xs) < 2 {
		t.Fatalf("crossings below the lower notch: %v", xs)
	}
	near(t, "left edge below the lower notch", xs[len(xs)-2], stemLeft)

	// Foot serif on the right only, of length SerifJut, and none on the left:
	// at the baseline the stem starts at its own edge and the bowl is apart.
	fs := path.CrossingsX(footScanY)
	if len(fs) != 4 {
		t.Fatalf("crossings at the foot: %v", fs)
	}
	near(t, "left edge of the stem at the baseline", fs[2], stemLeft)
	near(t, "right foot serif", fs[3]-stemRight, p.SerifJut)
	if fs[2]-fs[1] < 20 {
		t.Errorf("the bowl comes within %.1f of the stem at the baseline", fs[2]-fs[1])
	}
	// The foot serif has the tip thickness and the bracket of the others.
	ts := path.CrossingsY(stemRight + p.SerifJut - tipInset)
	if len(ts) != 2 {
		t.Fatalf("crossings on the foot serif tip: %v", ts)
	}
	near(t, "foot serif thickness", ts[1]-ts[0], p.SerifThick)

	// The head serif is that of l, on the left at the ascender height, and
	// there is none on the right.  The scans of checkHeadSerif are repeated
	// here on the last crossings.
	top := p.Ascender
	hs := path.CrossingsX(top - p.HeadSerifDrop - 1)
	near(t, "head serif length", stemLeft-hs[0], p.SerifJut)
	near(t, "right edge at the head serif", hs[len(hs)-1], stemRight)
	for _, f := range []float64{0.25, 0.5, 0.75} {
		us := path.CrossingsY(stemLeft - f*p.SerifJut)
		near(t, "upper edge of the head serif", lastOf(us), top-f*p.HeadSerifDrop)
	}
	tip := path.CrossingsY(stemLeft - p.SerifJut + tipInset)
	near(t, "head serif thickness", tip[len(tip)-1]-tip[len(tip)-2], p.SerifThick+p.HeadSerifDrop*tipInset/p.SerifJut)
	near(t, "top of the stem", lastOf(path.CrossingsY(stemLeft+p.Stem/2)), top)
	below := path.CrossingsX(top - p.HeadSerifDrop - p.SerifThick - headUnderRun - 2)
	near(t, "stem left below the head", below[0], stemLeft)
	near(t, "stem right below the head", below[1], stemRight)
}

func TestDBowlIsTheBowlOfO(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := dGlyph(t, p)
	// The bowl of d is that of an o whose left side space is that of the bowl.
	po := p
	po.SideRoundLeft = p.DQSideLeft
	o, _, _ := outlineOf("o", po)

	// The thickest part, the thin parts and the overshoot of the o.
	mid := p.XHeight / 2
	dx, ox := path.CrossingsX(mid), o.CrossingsX(mid)
	near(t, "thickest part", dx[1]-dx[0], ox[1]-ox[0])
	near(t, "thickest part, CurveLower", dx[1]-dx[0], p.CurveLower)
	near(t, "free extreme", dx[0], ox[0])
	xt := p.DQSideLeft + p.OBowlWidth/2
	dy, oy := path.CrossingsY(xt), o.CrossingsY(xt)
	if len(dy) != 4 {
		t.Fatalf("crossings through the extremes: %v", dy)
	}
	near(t, "thin part at the top", dy[3]-dy[2], oy[3]-oy[2])
	near(t, "thin part at the bottom", dy[1]-dy[0], oy[1]-oy[0])
	near(t, "overshoot at the top", dy[3], oy[3])
	near(t, "overshoot at the bottom", dy[0], oy[0])
	near(t, "top of the bowl", dy[3], p.XHeight+p.OvershootLower)

	// The free side is drawn from the same curve code as the o: its segments
	// are those of the left half of the o.
	want := func(c outline.Contour, from, to int) []outline.Segment { return c.Segs[from:to] }
	sameRun := func(what string, c outline.Contour, run []outline.Segment) {
		t.Helper()
		for i := 0; i+len(run) <= len(c.Segs); i++ {
			ok := true
			for j, s := range run {
				if c.Segs[i+j] != s {
					ok = false
					break
				}
			}
			if ok {
				return
			}
		}
		t.Errorf("%s: the segments of the o are not in d", what)
	}
	sameRun("outer free side", path[0], want(o[0], 1, 3))
	sameRun("counter free side", path[1], want(o[1], 1, 3))
}

func TestDCounter(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := dGlyph(t, p)
	stemLeft := design.DStemLeft(p)

	// The counter is bounded on the right by the stem edge, between the joins.
	cx0, cy0, cx1, cy1 := (outline.Path{path[1]}).BBox()
	near(t, "right edge of the counter", cx1, stemLeft)
	near(t, "left edge of the counter", cx0, p.DQSideLeft+p.CurveLower)
	near(t, "top of the counter", cy1, p.XHeight+p.OvershootLower-p.ThinLower)
	near(t, "bottom of the counter", cy0, -p.OvershootLower+p.ThinLower)

	// Its width at mid height is 0.88 to 1.04 of that of the o, which is 245.
	o, _, _ := outlineOf("o", p)
	mid := p.XHeight / 2
	dx, ox := path.CrossingsX(mid), o.CrossingsX(mid)
	dw, ow := dx[2]-dx[1], ox[2]-ox[1]
	if r := dw / ow; r < 0.88 || r > 1.04 {
		t.Errorf("the counter is %.1f wide at mid height against %.1f for the o, a ratio of %.3f outside 0.88 to 1.04", dw, ow, r)
	}
	near(t, "counter width", dw, 244)
}

func TestDJoins(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := dGlyph(t, p)
	stemLeft := design.DStemLeft(p)
	b := dBowl(p)
	if b.X != stemLeft || b.NotchUpper != p.DBowl.NotchUpper || b.JoinLower != p.DBowl.JoinLower {
		t.Fatalf("the bowl of d does not take its heights from the parameters: %+v", b)
	}

	// The counter meets the stem at the joins and the outer edge at the notches.
	vertex := func(c outline.Contour, y float64) bool {
		pt := c.Start
		if pt.X == stemLeft && pt.Y == y {
			return true
		}
		for _, s := range c.Segs {
			if s.P3.X == stemLeft && s.P3.Y == y {
				return true
			}
		}
		return false
	}
	for _, y := range []float64{p.DBowl.JoinUpper, p.DBowl.JoinLower} {
		if !vertex(path[1], y) {
			t.Errorf("the counter has no corner on the stem at height %v", y)
		}
	}
	for _, y := range []float64{p.DBowl.NotchUpper, p.DBowl.NotchLower} {
		if !vertex(path[0], y) {
			t.Errorf("the outer contour has no corner on the stem at height %v", y)
		}
	}
	// The thickness at the joins, at right angles to the stroke.
	for _, y := range []float64{p.DBowl.NotchUpper, p.DBowl.NotchLower} {
		d := thicknessAt(outline.Point{X: stemLeft, Y: y}, path[1])
		if d < 0.9*p.BowlJoinThick || d > 1.3*p.BowlJoinThick {
			t.Errorf("thickness at the join at %v is %.1f, want %v to %v", y, d, 0.9*p.BowlJoinThick, 1.3*p.BowlJoinThick)
		}
	}
	// The stroke thickens from each join to the thickest part.
	checkBowlStroke(t, path, b)
}
