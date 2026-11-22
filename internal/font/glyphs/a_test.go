// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter a.  The tests measure the outline and compare it with the
// parameters.

// aGlyph returns the unrounded outline of a, which has an outer contour and a
// counter.
func aGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, advance, ok := outlineOf("a", p)
	if !ok {
		t.Fatal("glyph a not found")
	}
	if len(path) != 2 {
		t.Fatalf("a has %d contours, want an outer contour and a counter", len(path))
	}
	return path, advance
}

func TestAAdvance(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	_, adv := aGlyph(t, p)
	// The advance is the left side space of a, the bowl, the stem, the reach of
	// the tail and the space right of the tail.
	if want := p.ASideLeft + p.ABowl.Width + p.Stem + p.ATailReach + p.ATailSideRight; adv != want || adv != design.AAdvance(p) {
		t.Errorf("advance %v, want %v", adv, want)
	}
	if r := adv / design.NAdvance(p); r < 0.833 || r > 0.910 {
		t.Errorf("the advance is %.3f of that of n, outside the range 0.833 to 0.910", r)
	}
	e, ok := lookup("a")
	if !ok || e.r != 'a' || e.left != sideBowlA || e.right != sideTailA {
		t.Errorf("table entry of a: %+v, want rune a with the bowl of a on the left side and the stem with the tail of a on the right", e)
	}
}

// The bounding-box rule of a takes the left side space of a, ASideLeft, which
// the advance of a is made from.
func TestABoundingBoxUsesASideLeft(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	p.ASideLeft = p.SideRoundLeft - 2
	e, _ := lookup("a")
	if left, _ := e.inkMargins(p); left != p.ASideLeft {
		t.Errorf("the ink of a keeps %v from the left edge, want ASideLeft = %v", left, p.ASideLeft)
	}
	path, adv, _ := outlineOf("a", p)
	if problems := checkBBox(structureCase{params: p, glyph: "a", path: path, advance: adv}); len(problems) > 0 {
		t.Errorf("bounding box of a with ASideLeft = %v: %v", p.ASideLeft, problems)
	}
}

// The bowl, not the ball, governs the left side space: the bowl is the
// leftmost ink and the ball stands AArcReach inside it.
func TestALeftSide(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := aGlyph(t, p)
	x0, y0, x1, y1 := path.BBox()
	near(t, "left extreme of the glyph", x0, p.ASideLeft)
	near(t, "bottom with overshoot", y0, -p.OvershootLower)
	near(t, "top of the arc with overshoot", y1, p.XHeight+p.OvershootLower)
	near(t, "right end of the tail", x1, design.AStemLeft(p)+p.Stem+p.ATailReach)
	// The leftmost ink above the bowl is the ball.
	left := math.Inf(1)
	for _, q := range path[0].Flatten() {
		if q.Y > aBowl(p).Top+10 {
			left = math.Min(left, q.X)
		}
	}
	near(t, "reach of the arc inside the bowl", left-p.ASideLeft, p.AArcReach)
	if r := p.AArcReach; r < 17 || r > 40 {
		t.Errorf("the reach of the arc, %v, is outside the range 17 to 40", r)
	}
}

// The stem has the stem width on the right, from the foot up to the height at
// which the outer edge of the arc leaves it.
func TestAStem(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := aGlyph(t, p)
	stemLeft := design.AStemLeft(p)
	stemRight := stemLeft + p.Stem
	b := aBowl(p)
	// The right-hand edge is straight from the end of the tail to aOuterDrop below
	// ShoulderStart and leaves to the left above it.
	top := p.ShoulderStart - aOuterDrop
	for y := p.ATailTipHeight + 10; y <= top; y += 25 {
		near(t, "right edge of the stem", lastOf(path.CrossingsX(y)), stemRight)
	}
	if r := lastOf(path.CrossingsX(top + 80)); r > stemRight-2 {
		t.Errorf("80 above the straight part the right edge is still at %.1f, want it to have left the stem edge %v", r, stemRight)
	}
	// Between the joins the stem has the stem width: the counter lies to its
	// left.
	for y := b.JoinLower + 20; y < b.JoinUpper-20; y += 25 {
		xs := path.CrossingsX(y)
		if len(xs) != 4 {
			t.Fatalf("crossings between the joins at %.0f: %v", y, xs)
		}
		near(t, "left edge between the joins", xs[2], stemLeft)
		near(t, "stem width between the joins", xs[3]-xs[2], p.Stem)
	}
	// Above the upper notch the left-hand edge is the stem's up to the height at
	// which the underside of the arc meets it.
	for y := b.NotchUpper + 10; y < p.ShoulderStartInner+aInnerRaise-2; y += 20 {
		xs := path.CrossingsX(y)
		near(t, "left edge above the bowl", xs[len(xs)-2], stemLeft)
	}
}

func TestABowl(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := aGlyph(t, p)
	stemLeft := design.AStemLeft(p)
	outer, counter := path[0], path[1]
	if outer.SignedArea() <= 0 || counter.SignedArea() >= 0 {
		t.Error("the outer contour is not counter-clockwise or the counter not clockwise")
	}
	b := aBowl(p)
	// The extreme at the top is at the stem: the bowl is scanned just left of it.
	xt := b.X - 0.5

	// Width: from the free extreme to the stem edge.  Top: the outer edge at the
	// stem, which is the highest point of the bowl; the bowl is 0.63 of the width
	// of the o.
	cx0, cy0, cx1, cy1 := (outline.Path{counter}).BBox()
	near(t, "right edge of the counter", cx1, stemLeft)
	near(t, "left edge of the counter", cx0, p.ASideLeft+p.ABowl.Thick)
	near(t, "bowl width", stemLeft-p.ASideLeft, p.ABowl.Width)
	if r := p.ABowl.Width / p.OBowlWidth; r < 0.567 || r > 0.719 {
		t.Errorf("the bowl is %.3f of the o, outside the range 0.567 to 0.719", r)
	}
	// The extremes at the top and at the bottom lie at different distances from
	// the stem (aLowerHalf).
	ys := path.CrossingsY(xt)
	if len(ys) != 6 {
		t.Fatalf("crossings through the extreme at the top of the bowl: %v", ys)
	}
	near(t, "top of the bowl", ys[3], b.Top)
	near(t, "stroke at the top", ys[3]-ys[2], p.EBarThick)
	ys = path.CrossingsY(aLowerHalf(b).xExtreme())
	if len(ys) != 6 {
		t.Fatalf("crossings through the extreme at the bottom of the bowl: %v", ys)
	}
	near(t, "bottom of the bowl", ys[0], -p.OvershootLower)
	near(t, "thin part at the bottom", ys[1]-ys[0], p.ThinLower)
	near(t, "bottom of the counter", cy0, -p.OvershootLower+p.ThinLower)
	near(t, "top of the counter", cy1, b.JoinUpper)
	if r := b.Top / p.XHeight; r < 0.585 || r > 0.663 {
		t.Errorf("the bowl top is %.3f of the x-height, outside the range 0.585 to 0.663", r)
	}

	// The thickest part is on the free side, at ThickPos of the height of the
	// bowl, and nowhere is the stroke thicker.  It is measured at right angles
	// to the stroke.
	var best, at float64
	for _, q := range outer.Flatten() {
		if q.X >= aLowerHalf(b).xExtreme() || q.Y > b.Top+0.01 {
			continue
		}
		if d := thicknessAt(q, counter); d > best {
			best, at = d, q.Y
		}
	}
	near(t, "thickest part", best, p.ABowl.Thick)
	pos := (at - b.Bottom) / (b.Top - b.Bottom)
	if math.Abs(pos-p.ABowl.ThickPos) > 0.03 {
		t.Errorf("the thickest part lies at %.3f of the height, want %v", pos, p.ABowl.ThickPos)
	}
	if r := p.ABowl.Thick / p.CurveLower; r < 0.911 || r > 1.006 {
		t.Errorf("the thickest part is %.3f of that of the o, outside the range 0.911 to 1.006", r)
	}
	if p.ABowl.ThickPos < 0.28 || p.ABowl.ThickPos > 0.41 {
		t.Errorf("ABowl.ThickPos %v is outside the range 0.28 to 0.41", p.ABowl.ThickPos)
	}
}

func TestAJoins(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := aGlyph(t, p)
	stemLeft := design.AStemLeft(p)
	b := aBowl(p)
	if b.X != stemLeft || b.JoinLower != p.ABowl.JoinLower || b.NotchLower != p.ABowl.NotchLower {
		t.Fatalf("the bowl of a does not take its heights from the parameters: %+v", b)
	}
	vertex := func(c outline.Contour, y float64) bool {
		if c.Start.X == stemLeft && c.Start.Y == y {
			return true
		}
		for _, s := range c.Segs {
			if s.P3.X == stemLeft && s.P3.Y == y {
				return true
			}
		}
		return false
	}
	for _, y := range []float64{b.JoinUpper, b.JoinLower} {
		if !vertex(path[1], y) {
			t.Errorf("the counter has no corner on the stem at height %v", y)
		}
	}
	for _, y := range []float64{b.NotchUpper, b.NotchLower} {
		if !vertex(path[0], y) {
			t.Errorf("the outer contour has no corner on the stem at height %v", y)
		}
	}
	// Thickness at right angles to the stroke at the joins: below that of the
	// bowls, above that of the bar of e.
	near(t, "thickness at the lower join", thicknessAt(outline.Point{X: stemLeft, Y: b.NotchLower}, path[1]), p.BowlJoinThick)
	near(t, "thickness at the upper join", thicknessAt(outline.Point{X: stemLeft, Y: b.NotchUpper}, path[1]), p.EBarThick)
	// The upper join lies at 0.50 to 0.60 of the x-height and the lower join at
	// 0.19 to 0.27 of it.  Above, the counter's highest point at the stem is the
	// join.
	if r := b.JoinUpper / p.XHeight; r < 0.50 || r > 0.60 {
		t.Errorf("the upper join is at %.3f of the x-height, outside 0.50 to 0.60", r)
	}
	if r := p.ABowl.JoinLower / p.XHeight; r < 0.19 || r > 0.27 {
		t.Errorf("the lower join is at %.3f of the x-height, outside 0.19 to 0.27", r)
	}
	// The upper stroke leaves the stem more horizontally than that of d: the
	// outer edge runs from the notch at a smaller angle to the horizontal.
	cosA, _ := b.leaveAngle(b.NotchUpper, b.JoinUpper)
	cosD, _ := dBowl(p).leaveAngle(p.DBowl.NotchUpper, p.DBowl.JoinUpper)
	if !(cosA > cosD+0.2) {
		t.Errorf("the upper stroke of a leaves the stem at cos %.2f, that of d at %.2f: not more horizontal", cosA, cosD)
	}
	// The stroke grows monotonically from the upper join, where it is as thick
	// as the bar of e, to the thickest part, and shrinks from there to the thin
	// part at the extreme at the bottom, that of the lower half; between that
	// extreme and the lower join it neither pinches below the thin part nor swells
	// beyond the thickest part.  The thickness is measured at right angles to the
	// stroke.
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	lower := aLowerHalf(b)
	xb := lower.xExtreme()
	thick := func(from, to outline.Point) []float64 {
		pts, _ := outerEdge(t, path, from, to)
		var th []float64
		for _, q := range pts {
			if q.X < b.X-0.01 {
				th = append(th, thicknessAt(q, path[1]))
			}
		}
		if len(th) < 10 {
			t.Fatalf("only %d samples from %v to %v", len(th), from, to)
		}
		return th
	}
	monotone := func(what string, th []float64, sign, start, end float64) {
		for i := 1; i < len(th); i++ {
			if sign*(th[i]-th[i-1]) < -0.5 {
				t.Errorf("the stroke %s goes from %.2f to %.2f", what, th[i-1], th[i])
				break
			}
		}
		if math.Abs(th[0]-start) > 2 || math.Abs(th[len(th)-1]-end) > 2 {
			t.Errorf("the stroke %s runs from %.1f to %.1f, want %v to %v", what, th[0], th[len(th)-1], start, end)
		}
	}
	monotone("from the upper join to the free extreme", thick(pt(b.X, b.NotchUpper), pt(b.Left, b.thickY())), 1, p.EBarThick, b.Thick)
	monotone("from the free extreme to the bottom", thick(pt(b.Left, b.thickY()), pt(xb, b.Bottom)), -1, b.Thick, b.ThinBottom)
	for _, d := range thick(pt(xb, b.Bottom), pt(b.X, b.NotchLower)) {
		if d < math.Min(b.ThinBottom, b.JoinThick)-2 || d > b.Thick {
			t.Errorf("the lower stroke is %.1f thick between the extreme at the bottom and the join, outside %v to %v", d, math.Min(b.ThinBottom, b.JoinThick)-2, b.Thick)
			break
		}
	}
}

// heading returns the direction of the vector from a to b.
func heading(a, b outline.Point) float64 { return math.Atan2(b.Y-a.Y, b.X-a.X) }

// The top of a is one stroke which grows out of the stem and ends in the ball.
// Measured across it, from the tip of the wedge back along the underside to the
// stem, its thickness never falls by more than aTopFallBack, it changes nowhere
// by more than stalkRate units per unit of length (no jump), and where the
// underside meets the stem it has grown to at least aTopSwell of the stem.
const (
	aTopFallBack = 1.0
	aTopSwell    = 0.75
)

func TestATopSwellsIntoTheStem(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, _ := cachedOutline(np, "a")
			s, th := stalkProfile(t, p, "a", path)
			high := th[0]
			for i := range th {
				if th[i] < high-aTopFallBack {
					t.Errorf("parameter set %s: the top stroke thins from %.1f to %.1f at %.0f from the tip of the wedge", np.name, high, th[i], s[i])
					break
				}
				high = math.Max(high, th[i])
				if i > 0 {
					if ds := s[i] - s[i-1]; ds > 0.3 && math.Abs(th[i]-th[i-1])/ds > stalkRate {
						t.Errorf("parameter set %s: the thickness of the top stroke jumps by %.2f per unit at %.0f from the tip", np.name, math.Abs(th[i]-th[i-1])/ds, s[i])
						break
					}
				}
			}
			end := th[len(th)-1]
			if np.name == "regular" {
				t.Logf("the top stroke is %.1f thick at the ball and %.1f where it meets the stem (stem %v)", th[0], end, p.Stem)
			}
			if end < aTopSwell*p.Stem {
				t.Errorf("parameter set %s: the top stroke is %.1f thick where it meets the stem, want at least %.1f", np.name, end, aTopSwell*p.Stem)
			}
		})
	}
}

// The underside of the top runs into the left-hand edge of the stem as one
// curve: the contour turns by at most aKnuckleTurn where it meets the stem, and
// by at most aKnuckleSpan over the last aKnuckleLength units before it.
const (
	aKnuckleTurn   = 3 * math.Pi / 180
	aKnuckleSpan   = 20 * math.Pi / 180
	aKnuckleLength = 20.0
)

func TestATopHasNoKnuckle(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, _ := cachedOutline(np, "a")
			in, _ := stalkEdges(p, "a")
			ct := path[0]
			n := len(ct.Segs)
			i := nodeNear(t, ct, in[1])
			before, after := segCubic(ct, i+n-1), segCubic(ct, i)
			// The segment after the node is the straight edge of the stem.
			a := before.tangent(1)
			d := math.Abs(turn(math.Atan2(a.Y, a.X), heading(after.P0, after.P3)))
			pts, s := outerEdge(t, path, in[0], in[1])
			total := s[len(s)-1]
			k := len(pts) - 2
			for k > 0 && total-s[k] < aKnuckleLength {
				k--
			}
			m := len(pts) - 1
			span := math.Abs(turn(heading(pts[k], pts[k+1]), heading(pts[m-1], pts[m])))
			if np.name == "regular" {
				t.Logf("the underside turns by %.2f degrees at the stem and by %.1f over the last %v units", d*180/math.Pi, span*180/math.Pi, aKnuckleLength)
			}
			if d > aKnuckleTurn {
				t.Errorf("parameter set %s: the contour turns by %.1f degrees where the underside meets the stem, want at most %.0f", np.name, d*180/math.Pi, aKnuckleTurn*180/math.Pi)
			}
			if span > aKnuckleSpan {
				t.Errorf("parameter set %s: the underside turns by %.1f degrees over the last %v units before the stem, want at most %.0f", np.name, span*180/math.Pi, aKnuckleLength, aKnuckleSpan*180/math.Pi)
			}
		})
	}
}

// The bowl of a is an oval: between the stem and the free extreme both edges
// of its upper stroke turn one way only and have no straight stretch (no
// stretch of bowlStretch units turns by less than bowlStretchTurn), and the roof
// of the counter, from its free extreme to the stem, has no level stretch longer
// than maxLevelRoof: a stretch whose direction stays within levelAngle of the
// horizontal.
const (
	bowlStretch     = 30.0
	bowlStretchTurn = 0.04
	levelAngle      = 3 * math.Pi / 180
	maxLevelRoof    = 45.0
)

func TestABowlIsAnOval(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, _ := cachedOutline(np, "a")
			b := aBowl(p)
			pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
			counter := outline.Path{path[1]}
			for _, e := range []struct {
				what string
				path outline.Path
				a, b outline.Point
			}{
				{"outer edge", path, pt(b.X, b.NotchUpper), pt(b.Left, b.thickY())},
				{"counter", counter, pt(b.Left+b.Thick, b.thickY()), pt(b.X, b.JoinUpper)},
			} {
				pts, s := outerEdge(t, e.path, e.a, e.b)
				least, at := slidingTurns(pts, s, bowlStretch)
				// The edge turns the way of its total turn: left on the outer
				// contour, right on the counter, which runs clockwise.
				var turns []float64
				var total, wrong float64
				for i := 2; i < len(pts); i++ {
					if pts[i] == pts[i-1] || pts[i-1] == pts[i-2] {
						continue
					}
					d := turn(heading(pts[i-2], pts[i-1]), heading(pts[i-1], pts[i]))
					turns = append(turns, d)
					total += d
				}
				for _, d := range turns {
					if d*total < 0 {
						wrong += math.Abs(d)
					}
				}
				if np.name == "regular" {
					t.Logf("%s of the upper stroke: least turn %.3f over %v units at (%.0f, %.0f), wrong way %.4f", e.what, least, bowlStretch, at.X, at.Y, wrong)
				}
				if least < bowlStretchTurn {
					t.Errorf("parameter set %s: the %s of the upper stroke of the bowl turns by only %.3f radians over %v units at (%.0f, %.0f), want at least %v", np.name, e.what, least, bowlStretch, at.X, at.Y, bowlStretchTurn)
				}
				if wrong > edgeWrongWay {
					t.Errorf("parameter set %s: the %s of the upper stroke of the bowl turns the wrong way by %.3f radians", np.name, e.what, wrong)
				}
			}
			// The roof of the counter, from the free extreme to the join.
			pts, s := outerEdge(t, counter, pt(b.Left+b.Thick, b.thickY()), pt(b.X, b.JoinUpper))
			var longest, run float64
			for i := 1; i < len(pts); i++ {
				if pts[i] == pts[i-1] {
					continue
				}
				if math.Abs(heading(pts[i-1], pts[i])) < levelAngle {
					run += s[i] - s[i-1]
					longest = math.Max(longest, run)
				} else {
					run = 0
				}
			}
			if np.name == "regular" {
				t.Logf("the longest level stretch of the roof of the counter is %.1f", longest)
			}
			if longest > maxLevelRoof {
				t.Errorf("parameter set %s: the roof of the counter runs level for %.1f units, want at most %v", np.name, longest, maxLevelRoof)
			}
		})
	}
}

// The bottom of the bowl of a is one curve into the stem.  From the extreme at
// the bottom to the lower notch (outer edge) and from the lower join to the
// extreme at the bottom of the counter, each edge turns one way only (the wrong
// way by at most edgeWrongWay), has no straight stretch (every stretch of
// bowlStretch units turns by at least bowlStretchTurn, save within
// aLeaveStraight of the stem, where the edge leaves the stem as the arches
// leave theirs), its curvature changes by at most aBottomRate per unit of
// length along it, and it goes on from the free side at the extreme without a jump
// of its curvature: the turn per unit over the aBottomWindow units on either
// side of the extreme differs by at most aBottomJump of the larger.
const (
	aLeaveStraight = 30.0
	aBottomWindow  = 3.0
	aBottomJump    = 0.15
	aBottomRate    = 0.001
)

func TestABowlBottomRunsIntoTheStem(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, _ := cachedOutline(np, "a")
			lo := aLowerHalf(aBowl(p))
			xt := lo.xExtreme()
			pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
			for _, e := range []struct {
				what     string
				path     outline.Path
				extreme  outline.Point
				stem     outline.Point
				stemLast bool
			}{
				{"outer edge", path, pt(xt, lo.Bottom), pt(lo.X, lo.NotchLower), true},
				{"counter", outline.Path{path[1]}, pt(xt, lo.Bottom+lo.ThinBottom), pt(lo.X, lo.JoinLower), false},
			} {
				var pts []outline.Point
				var s []float64
				if e.stemLast {
					pts, s = outerEdge(t, e.path, e.extreme, e.stem)
				} else {
					pts, s = outerEdge(t, e.path, e.stem, e.extreme)
				}
				var turns []float64
				var total, wrong float64
				for i := 2; i < len(pts); i++ {
					if pts[i] == pts[i-1] || pts[i-1] == pts[i-2] {
						continue
					}
					d := turn(heading(pts[i-2], pts[i-1]), heading(pts[i-1], pts[i]))
					turns = append(turns, d)
					total += d
				}
				for _, d := range turns {
					if d*total < 0 {
						wrong += math.Abs(d)
					}
				}
				if wrong > edgeWrongWay {
					t.Errorf("%s: the %s of the bottom of the bowl turns the wrong way by %.3f radians", np.name, e.what, wrong)
				}
				var keep []outline.Point
				var ks []float64
				for i := range pts {
					fromStem := s[i]
					if e.stemLast {
						fromStem = s[len(s)-1] - s[i]
					}
					if fromStem >= aLeaveStraight {
						keep = append(keep, pts[i])
						ks = append(ks, s[i])
					}
				}
				least, at := slidingTurns(keep, ks, bowlStretch)
				if least < bowlStretchTurn {
					t.Errorf("%s: the %s of the bottom of the bowl turns by only %.3f radians over %v units at (%.0f, %.0f)", np.name, e.what, least, bowlStretch, at.X, at.Y)
				}
				var rate float64
				prevK := math.NaN()
				for i := 2; i < len(pts); i++ {
					ds := s[i] - s[i-1]
					if ds < 1e-9 || pts[i-1] == pts[i-2] {
						continue
					}
					k := turn(heading(pts[i-2], pts[i-1]), heading(pts[i-1], pts[i])) / ds
					if !math.IsNaN(prevK) {
						rate = math.Max(rate, math.Abs(k-prevK)/ds)
					}
					prevK = k
				}
				if rate > aBottomRate {
					t.Errorf("%s: the curvature of the %s of the bottom of the bowl changes by %.5f per unit, want at most %v", np.name, e.what, rate, aBottomRate)
				}
				k := turnRateAround(t, e.path[0], e.extreme, aBottomWindow)
				if d := math.Abs(k[0]-k[1]) / math.Max(math.Abs(k[0]), math.Abs(k[1])); d > aBottomJump {
					t.Errorf("%s: the curvature of the %s jumps at the bottom extreme from %.5f to %.5f", np.name, e.what, k[0], k[1])
				}
				if np.name == "regular" {
					t.Logf("%s: wrong way %.4f, least turn %.3f, curvature at the extreme %.5f and %.5f, its change %.5f per unit", e.what, wrong, least, k[0], k[1], rate)
				}
			}
		})
	}
}

// turnRateAround returns the turn per unit of the contour over the w units
// before and the w units after its node at q.
func turnRateAround(t *testing.T, ct outline.Contour, q outline.Point, w float64) [2]float64 {
	t.Helper()
	n := len(ct.Segs)
	i := nodeNear(t, ct, q)
	// rate returns the turn per unit of c over the first w units from the end
	// at t0 (0 or 1).
	rate := func(c cubic, t0 float64) float64 {
		var length float64
		prev, u := c.at(t0), t0
		for k := 1; k <= 1000 && length < w; k++ {
			u = t0 + (1-2*t0)*float64(k)/1000
			p := c.at(u)
			length += math.Hypot(p.X-prev.X, p.Y-prev.Y)
			prev = p
		}
		ta, tb := c.tangent(t0), c.tangent(u)
		return math.Abs(math.Atan2(ta.X*tb.Y-ta.Y*tb.X, ta.X*tb.X+ta.Y*tb.Y)) / length
	}
	return [2]float64{rate(segCubic(ct, (i+n-1)%n), 1), rate(segCubic(ct, i), 0)}
}

// The top of the bowl of a lies on the line of the bar of e: where the upper
// stroke of the bowl meets the stem, its upper edge and its underside are at
// the heights of the upper edge and of the underside of the bar, within
// aBarMatch.  Both edges are level at the stem: the counter is highest there,
// the outer edge is nowhere higher over the last aLevelSpan units before the
// stem, and over the last aLevelRun units each edge falls by at most
// aLevelDrop.
const (
	aBarMatch  = 1.0
	aLevelSpan = 50.0
	aLevelRun  = 5.0
	aLevelDrop = 0.25
)

func TestAJoinMeetsTheBarOfE(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			ePath, _ := cachedOutline(np, "e")
			low, high := barLevel(t, ePath, (design.EBowlLeft(p)+design.EBowlRight(p))/2+0.5)

			path, _ := cachedOutline(np, "a")
			stem := design.AStemLeft(p)
			// The roof of the counter: its highest point, which lies at the stem.
			roof := outline.Point{Y: math.Inf(-1)}
			for _, q := range path[1].Flatten() {
				if q.Y > roof.Y {
					roof = q
				}
			}
			if roof.X < stem-0.5 {
				t.Errorf("parameter set %s: the counter is highest at (%.1f, %.1f), not at the stem %v", np.name, roof.X, roof.Y, stem)
			}
			// The outer edge of the upper stroke at the distance d from the stem: the
			// lowest crossing above the middle of the bar and below the arc.
			edge := func(d float64) float64 {
				y := math.NaN()
				for _, v := range path.CrossingsY(stem - d) {
					if v > (low+high)/2 && v < high+10 && (math.IsNaN(y) || v < y) {
						y = v
					}
				}
				if math.IsNaN(y) {
					t.Fatalf("parameter set %s: no outer edge of the bowl %v left of the stem: %v", np.name, d, path.CrossingsY(stem-d))
				}
				return y
			}
			top := edge(0.01)
			if np.name == "regular" {
				t.Logf("the upper stroke of the bowl meets the stem from %.2f to %.2f, the bar of e runs from %.2f to %.2f", roof.Y, top, low, high)
			}
			if math.Abs(top-high) > aBarMatch || math.Abs(roof.Y-low) > aBarMatch {
				t.Errorf("parameter set %s: the upper stroke of the bowl of a meets the stem from %.2f to %.2f, the bar of e runs from %.2f to %.2f", np.name, roof.Y, top, low, high)
			}
			for d := 1.0; d <= aLevelSpan; d++ {
				if y := edge(d); y > top+0.01 {
					t.Errorf("parameter set %s: the outer edge of the bowl is at %.2f %v left of the stem, higher than %.2f at the stem", np.name, y, d, top)
					break
				}
			}
			if drop := top - edge(aLevelRun); drop > aLevelDrop {
				t.Errorf("parameter set %s: the outer edge of the bowl falls by %.2f over the last %v units before the stem", np.name, drop, aLevelRun)
			}
			under := math.Inf(-1)
			for _, v := range path.CrossingsY(stem - aLevelRun) {
				if v < (low+high)/2 && v > low-10 {
					under = math.Max(under, v)
				}
			}
			if drop := roof.Y - under; drop > aLevelDrop {
				t.Errorf("parameter set %s: the counter falls by %.2f over the last %v units before the stem", np.name, drop, aLevelRun)
			}
		})
	}
}
