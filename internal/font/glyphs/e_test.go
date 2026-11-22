// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter e.  The tests measure the outline and compare it with the
// parameters and with fixed numbers.

// eGlyph returns the unrounded outline of e, which has an outer contour and a
// counter, the eye.
func eGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, advance, ok := outlineOf("e", p)
	if !ok {
		t.Fatal("glyph e not found")
	}
	if len(path) != 2 {
		t.Fatalf("e has %d contours, want an outer contour and the eye", len(path))
	}
	return path, advance
}

func TestEAdvance(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := eGlyph(t, p)
	// The advance is the round left side space, the bowl and the open right
	// side space.
	if want := p.SideRoundLeft + p.EBowlWidth + p.SideOpenRight; adv != want || adv != design.EAdvance(p) {
		t.Errorf("advance %v, want %v", adv, want)
	}
	ent, ok := lookup("e")
	if !ok || ent.r != 'e' || ent.left != sideRound || ent.right != sideOpen {
		t.Errorf("table entry of e: %+v, want rune e with a round left side and an open right side", ent)
	}
	// The ink ends at the extreme of the bowl, SideOpenRight before the
	// advance, and starts SideRoundLeft after the left edge.
	x0, _, x1, _ := path.BBox()
	near(t, "left side space", x0, p.SideRoundLeft)
	near(t, "right side space", adv-x1, p.SideOpenRight)
}

func TestEWidthAgainstO(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := eGlyph(t, p)
	o, _, _ := outlineOf("o", p)
	ex0, _, ex1, _ := path.BBox()
	ox0, _, ox1, _ := o.BBox()
	near(t, "bowl width", ex1-ex0, p.EBowlWidth)
	// The width of e is 0.874 to 0.995 of the width of o.
	r := (ex1 - ex0) / (ox1 - ox0)
	if r < 0.874 || r > 0.995 {
		t.Errorf("e is %.3f of the width of o, want inside 0.874 to 0.995", r)
	}
}

// barLevel returns the lower and the upper edge of the bar, from a vertical
// scan through the middle of the glyph, which crosses the bottom stroke, the bar
// and the top stroke.
func barLevel(t *testing.T, path outline.Path, x float64) (low, high float64) {
	t.Helper()
	ys := path.CrossingsY(x)
	if len(ys) != 6 {
		t.Fatalf("crossings at x = %v: %v, want the bottom stroke, the bar and the top stroke", x, ys)
	}
	return ys[2], ys[3]
}

func TestEBar(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := eGlyph(t, p)
	cx := (design.EBowlLeft(p) + design.EBowlRight(p)) / 2

	// EBarHeight is the height of the vertical centre of the bar, which is 247.
	// The thickness is 32, that of the thin parts of the round strokes, and a
	// parameter of its own.
	low, high := barLevel(t, path, cx+0.5)
	near(t, "bar centre", (low+high)/2, p.EBarHeight)
	near(t, "bar centre", (low+high)/2, 247)
	near(t, "bar thickness", high-low, p.EBarThick)
	near(t, "bar thickness", high-low, 32)
	if design.Regular().EBarThick != design.Regular().ThinLower {
		t.Errorf("EBarThick = %v, want it equal to ThinLower = %v", p.EBarThick, p.ThinLower)
	}

	// The bar is horizontal: both edges have the same height from the inner
	// left side to the right end, and its upper edge as far as the eye reaches.
	hasCrossing := func(ys []float64, y float64) bool {
		for _, v := range ys {
			if math.Abs(v-y) < 1e-6 {
				return true
			}
		}
		return false
	}
	_, _, eyeRight, _ := (outline.Path{path[1]}).BBox()
	eyeLeft, _, _, _ := (outline.Path{path[1]}).BBox()
	for x := eyeLeft + 4.13; x < eyeRight-4; x += 17.9 {
		if !hasCrossing(path.CrossingsY(x), low) || !hasCrossing(path.CrossingsY(x), high) {
			t.Errorf("at x = %.1f the bar has no edge at the heights %v and %v: crossings %v", x, low, high, path.CrossingsY(x))
		}
	}
	right := design.EBowlRight(p)
	for x := eyeRight + 5.3; x < right-4; x += 17.9 {
		if !hasCrossing(path.CrossingsY(x), low) {
			t.Errorf("at x = %.1f the underside of the bar is missing: crossings %v", x, path.CrossingsY(x))
		}
	}

	// The bar spans from the inner left side of the bowl to the right end: along
	// the middle of the bar the ink is one run from the left extreme of the bowl
	// to the right extreme, and the underside starts on the inner edge of the
	// left stroke and ends at the outer edge on the right.
	xs := path.CrossingsX(p.EBarHeight)
	if len(xs) != 2 {
		t.Fatalf("crossings along the middle of the bar: %v, want one run", xs)
	}
	near(t, "left end of the bar run", xs[0], design.EBowlLeft(p))
	near(t, "right end of the bar", xs[1], design.EBowlRight(p))
	under := path.CrossingsX(low - 3)
	above := path.CrossingsX(high + 3)
	if len(under) != 2 || len(above) != 4 {
		t.Fatalf("crossings below the bar %v and above it %v", under, above)
	}
	start := path[0].Start
	if math.Abs(start.Y-low) > 1e-6 || math.Abs(start.X-under[1]) > 1.5 {
		t.Errorf("the underside starts at %v, want height %v on the inner edge of the left stroke at x = %.1f", start, low, under[1])
	}
	if math.Abs(path[0].Segs[0].P3.Y-low) > 1e-6 || math.Abs(path[0].Segs[0].P3.X-xs[1]) > 1.5 || path[0].Segs[0].IsCurve {
		t.Errorf("the underside ends at %v, want a straight line to height %v at the right end x = %.1f", path[0].Segs[0].P3, low, xs[1])
	}
	near(t, "left end of the upper edge", above[1], under[1])
}

func TestEEye(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := eGlyph(t, p)
	eye := path[1]
	if path[0].SignedArea() <= 0 || eye.SignedArea() >= 0 {
		t.Error("the outer contour is not counter-clockwise or the eye not clockwise")
	}
	// The eye is a closed contour of its own, bounded below by the upper edge of
	// the bar and above by the inner edge of the top stroke.
	if last := eye.Segs[len(eye.Segs)-1]; last.P3 != eye.Start {
		t.Errorf("the eye ends at %v, want its start %v", last.P3, eye.Start)
	}
	x0, y0, x1, y1 := (outline.Path{eye}).BBox()
	barHigh := p.EBarHeight + p.EBarThick/2
	near(t, "bottom of the eye", y0, barHigh)
	near(t, "top of the eye", y1, p.XHeight+p.OvershootLower-p.ThinLower)
	near(t, "right of the eye", x1, design.EBowlRight(p)-p.CurveLower)
	if x0 < design.EBowlLeft(p)+p.CurveLower || x0 > design.EBowlLeft(p)+p.CurveLower+8 {
		t.Errorf("left of the eye at %.1f, want just right of %v", x0, design.EBowlLeft(p)+p.CurveLower)
	}
	// Its lower edge is the one straight segment at the height of the bar.
	flat := 0
	cur := eye.Start
	for _, s := range eye.Segs {
		if !s.IsCurve && s.P3.Y == barHigh && cur.Y == barHigh {
			flat++
		}
		cur = s.P3
	}
	if flat != 1 {
		t.Errorf("the eye has %d horizontal lines at the height of the bar, want 1", flat)
	}
}

// eTip returns the rightmost ink below the bar: the largest x over horizontal
// scans below the underside of the bar, with the mean height of the scans which
// reach it.
func eTip(path outline.Path, barLow float64) (x, y float64) {
	best := math.Inf(-1)
	for h := -11.5; h < barLow-1; h += 0.25 {
		if xs := path.CrossingsX(h); len(xs) > 0 {
			best = math.Max(best, xs[len(xs)-1])
		}
	}
	var sum, n float64
	for h := -11.5; h < barLow-1; h += 0.25 {
		if xs := path.CrossingsX(h); len(xs) > 0 && xs[len(xs)-1] >= best-0.5 {
			sum += h
			n++
		}
	}
	return best, sum / n
}

func TestEOpenBelowTheBar(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := eGlyph(t, p)
	barLow := p.EBarHeight - p.EBarThick/2
	xt, _ := eTip(path, barLow)
	barRight := path.CrossingsX(p.EBarHeight)[1]

	// The aperture is the vertical gap from the underside of the bar to the upper
	// edge of the ink below it, 2 units inside the nearer of the tip of the
	// terminal and the right end of the bar.
	xg := math.Min(xt, barRight) - 2
	ys := path.CrossingsY(xg)
	var below float64
	found := false
	for i := 0; i+1 < len(ys); i += 2 {
		if math.Abs(ys[i+1]-barLow) < 1e-6 {
			continue
		}
		if ys[i+1] < barLow && ys[i] < barLow {
			below, found = ys[i+1], true
		}
	}
	if !found {
		t.Fatalf("no ink below the bar at x = %v: crossings %v", xg, ys)
	}
	aperture := barLow - below
	near(t, "aperture", aperture, p.EAperture)
	near(t, "aperture", aperture, 121)

	// Open to the right: between the highest point of the terminal and the
	// underside of the bar only the left stroke is ink.
	topTerminal := math.Inf(-1)
	for y := -11.5; y < barLow-1; y += 0.25 {
		if len(path.CrossingsX(y)) > 2 {
			topTerminal = y
		}
	}
	if topTerminal < below || topTerminal > below+30 {
		t.Errorf("the terminal reaches %.1f, want a little above the %.1f of its cut", topTerminal, below)
	}
	for y := topTerminal + 1; y < barLow-1; y += 3.7 {
		if xs := path.CrossingsX(y); len(xs) != 2 {
			t.Errorf("at height %.1f, below the bar and above the terminal, the crossings are %v, want the left stroke only", y, xs)
		}
	}
	// There is no ink between the left stroke and the tip at the height of the top
	// of the terminal.
	xs := path.CrossingsX(topTerminal + 1)
	if len(xs) != 2 || xs[1] > design.EBowlLeft(p)+p.CurveLower+15 {
		t.Errorf("crossings at the top of the terminal %v", xs)
	}
}

// strokePerp measures the thickness of a roughly horizontal stroke at x: the
// lowest run of the vertical scans within span of x gives an upper and a lower
// edge, and the result is the mean of the distance from the upper edge at x to
// the lower edge and of the distance from the lower edge at x to the upper edge.
func strokePerp(path outline.Path, x, span float64) float64 {
	var up, lo []outline.Point
	for xx := x - span; xx <= x+span+0.05; xx += 0.5 {
		if ys := path.CrossingsY(xx); len(ys) >= 2 {
			lo = append(lo, outline.Point{X: xx, Y: ys[0]})
			up = append(up, outline.Point{X: xx, Y: ys[1]})
		}
	}
	k := 0
	for i := range up {
		if math.Abs(up[i].X-x) < math.Abs(up[k].X-x) {
			k = i
		}
	}
	nearest := func(q outline.Point, poly []outline.Point) float64 {
		best := math.Inf(1)
		for i := 0; i+1 < len(poly); i++ {
			best = math.Min(best, segDist(q, poly[i], poly[i+1]))
		}
		return best
	}
	return (nearest(up[k], lo) + nearest(lo[k], up)) / 2
}

func TestETerminal(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := eGlyph(t, p)
	barLow := p.EBarHeight - p.EBarThick/2
	xt, yt := eTip(path, barLow)

	// The tip lies on the bowl, not beyond its right extreme, and its outer
	// corner EAperture below the bar.
	if xt > design.EBowlRight(p)+0.5 || xt < design.EBowlRight(p)-60 {
		t.Errorf("the tip is at x = %.1f, want within 60 of the right extreme %v", xt, design.EBowlRight(p))
	}
	if yt < barLow-p.EAperture-12 || yt > barLow-p.EAperture+12 {
		t.Errorf("the tip is at height %.1f, want about %v", yt, barLow-p.EAperture)
	}

	// The thickness at right angles to the stroke 0.3 stem and one stem width
	// from the tip: 30 near the tip and 40 one stem back (the tip of
	// the cut is 34 wide).
	stem := p.Stem
	near30 := strokePerp(path, xt-0.3*stem, 0.6*stem)
	near38 := strokePerp(path, xt-1.0*stem, 0.6*stem)
	near(t, "thickness 0.3 stem from the tip", near30, 30)
	if near38 < 35 || near38 > 41 {
		t.Errorf("thickness one stem from the tip = %.1f, want 35 to 41", near38)
	}

	// The terminal thins monotonically towards its tip from one stem back.
	prev := math.Inf(1)
	for d := 1.0; d >= 0.3; d -= 0.05 {
		th := strokePerp(path, xt-d*stem, 0.6*stem)
		if th > prev+0.3 {
			// prev is the thickness one step further from the tip.
			t.Errorf("the terminal thickens towards the tip: %.2f at %.2f stem from it after %.2f", th, d, prev)
		}
		prev = math.Min(prev, th)
	}
	if !(near30 < near38-4) {
		t.Errorf("the terminal does not thin: %.1f near the tip against %.1f one stem back", near30, near38)
	}

	// The cut at the tip is a straight line of length ETerminalTip, at right
	// angles to the stroke.
	var cut float64
	cur := path[0].Start
	for _, s := range path[0].Segs {
		if !s.IsCurve && math.Abs(cur.X-xt) < 30 && cur.Y < barLow-20 && s.P3.Y < barLow-20 {
			cut = math.Hypot(s.P3.X-cur.X, s.P3.Y-cur.Y)
		}
		cur = s.P3
	}
	near(t, "length of the cut", cut, p.ETerminalTip)

	// Below the terminal the stroke has the thin part of the o at the bottom.
	cx := (design.EBowlLeft(p) + design.EBowlRight(p)) / 2
	ys := path.CrossingsY(cx + 0.5)
	near(t, "thin part at the bottom", ys[1]-ys[0], p.ThinLower)
	near(t, "overshoot at the bottom", ys[0], -p.OvershootLower)
}

func TestELeftSideAndTopLikeO(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := eGlyph(t, p)
	o, _, _ := outlineOf("o", p)
	bottom, top := -p.OvershootLower, p.XHeight+p.OvershootLower

	// The left extreme, an on-curve point, lies at EThickPos of the height, which
	// is 0.48, and the stroke there is as thick as that of the o.
	ext := path[0].Start
	for _, s := range path[0].Segs {
		if s.P3.X < ext.X {
			ext = s.P3
		}
	}
	near(t, "left extreme", ext.X, p.SideRoundLeft)
	if f := (ext.Y - bottom) / (top - bottom); math.Abs(f-p.EThickPos) > 0.005 || math.Abs(f-0.48) > 0.01 {
		t.Errorf("the left extreme is at %.3f of the height, want %v", f, 0.48)
	}
	xs := path.CrossingsX(ext.Y + 0.01)
	oxs := o.CrossingsX(p.XHeight / 2)
	near(t, "thickest part", xs[1]-xs[0], p.CurveLower)
	near(t, "thickest part against the o", xs[1]-xs[0], oxs[1]-oxs[0])
	// No other height has a thicker left stroke, outside the bar.
	barLow, barHigh := p.EBarHeight-p.EBarThick/2, p.EBarHeight+p.EBarThick/2
	for y := bottom + 60; y < top-60; y += 2.9 {
		if y > barLow-4 && y < barHigh+4 {
			continue
		}
		xs := path.CrossingsX(y)
		if w := xs[1] - xs[0]; w > p.CurveLower+0.1 {
			t.Errorf("at height %.1f the left stroke is %.2f thick, more than the thickest part %v", y, w, p.CurveLower)
		}
	}

	// The top has the thin part and the overshoot of the o, and so has the bottom.
	cx := (design.EBowlLeft(p) + design.EBowlRight(p)) / 2
	ocx := (design.OBowlLeft(p) + design.OBowlRight(p)) / 2
	ys, oys := path.CrossingsY(cx+0.5), o.CrossingsY(ocx+0.5)
	near(t, "thin part at the top", ys[5]-ys[4], oys[3]-oys[2])
	near(t, "thin part at the top against ThinLower", ys[5]-ys[4], p.ThinLower)
	near(t, "overshoot at the top", ys[5], oys[3])
	near(t, "overshoot at the top against the x-height", ys[5], p.XHeight+p.OvershootLower)
	near(t, "thin part at the bottom", ys[1]-ys[0], oys[1]-oys[0])
	near(t, "overshoot at the bottom", ys[0], oys[0])

	// The right side above the bar is as thick as that of the o.
	rx := path.CrossingsX(barHigh + 4)
	if len(rx) != 4 {
		t.Fatalf("crossings above the bar: %v", rx)
	}
	// A little above the bar it is 94 less the few units by which the inner edge
	// has already left its extreme.
	if w := rx[3] - rx[2]; w < p.CurveLower-5 || w > p.CurveLower+0.5 {
		t.Errorf("the right stroke above the bar is %.1f thick, want 89 to 94.5", w)
	}
}

// The eye of e leaves the bar vertically, at the upper edge of the bar, and has a
// round top: its sides are vertical there (the side 20 above the bar has moved by
// at most a unit) and nowhere in the upper part is the curve tighter than half
// the half width of the eye.
func TestEEyeIsARoundArch(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, _ := cachedOutline(np, "e")
			f := newEFrame(p)
			eye := outline.Path{path[1]}
			at := func(dy float64) (l, r float64) {
				xs := eye.CrossingsX(f.barHigh + dy)
				if len(xs) != 2 {
					t.Fatalf("the eye at %v above the bar is crossed at %v, want two places", dy, xs)
				}
				return xs[0], xs[1]
			}
			l0, r0 := at(0.5)
			for _, dy := range []float64{10, 20} {
				l, r := at(dy)
				if math.Abs(l-l0) > 1 || math.Abs(r-r0) > 1 {
					t.Errorf("%v above the bar the eye sides are at %.2f and %.2f, at the bar %.2f and %.2f: they lean", dy, l, r, l0, r0)
				}
			}
			half := (r0 - l0) / 2
			for i := range eye[0].Segs {
				c := segCubic(eye[0], i)
				for k := 0; k <= 100; k++ {
					tt := float64(k) / 100
					q := c.at(tt)
					if kap := math.Abs(c.curvature(tt)); q.Y > f.barHigh+60 && kap > 1e-9 && 1/kap < 0.5*half {
						t.Fatalf("at (%.0f, %.0f) the eye has a radius of %.1f, want at least %.1f (half its width %.1f): the top is pointed", q.X, q.Y, 1/kap, 0.5*half, half)
					}
				}
			}
		})
	}
}
