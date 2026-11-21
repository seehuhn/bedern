// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The section sign.
//
// Two S-shaped strokes stand one above the other, pushed so far into each
// other that the lower bowl of the upper and the upper bowl of the lower are
// one counter, the eye, in the middle; the lower stroke is the upper turned
// through half a turn about the centre O of the sign, and the sign has
// point symmetry.  The upper stroke begins at the upper right in the ball
// terminal of the figures, runs thin over the round top of its hook and down
// its left side (SectHookThick), turns into its spine, which falls to the
// right at sectSpineAngle and carries the weight of the sign (SectSpine), and
// at the lower right end of the spine turns down and back to the left as a
// thin arc which closes the eye on the right and runs into the spine of the
// lower stroke.  So the stress lies in the two parallel spines, the sides of
// the eye are thin, and the eye is the widest part: its extremes lie outside
// the hooks.
//
// The eye is a super-ellipse with the squareness of the inner edge of o,
// SectEyeLong long along the spines and SectEyeShort across them, about O and
// turned to the angle of the spines: its long sides are the inner edges of
// the two spines, slightly curved and nowhere straight, and its round ends
// are the inner edges of the thin arcs.  The eye is one convex curve; the arc
// of each stroke runs into the other spine along it without a joint.
//
// The upper edge of the stroke is drawn as the edges of s are (s.go).  It has
// its inflection at the middle of the spine, SectSpine from the middle of the
// eye's upper long side at right angles to the spine, where it runs along the
// spine without curvature.  From there to each side extreme it is one cubic
// whose first handle lies on the line of the spine, sectSpineHandle of the
// way to its second control point, the corner where that line meets the
// vertical through the extreme.  On the right the extreme lies sectArcSide
// times Figures.Thin outside the eye's right extreme, and the quarter on to
// the bottom, Figures.Thin below the eye's bottom, has the handle of the
// quarters of s at the bottom and at the side the curvature of the curve
// before it (sideQuarter); the height of the side extreme is the one at which
// that handle is sectArcHandle of the quarter's height.  Beyond the bottom the
// edge runs on as the mirror image of that quarter, until it crosses the upper
// edge of the lower stroke's spine, where the white beside the eye ends in a
// crotch.  On the left the extreme is the inner side of the hook,
// SectHookThick right of its outer side and sectHookInY above its outer
// extreme, and the quarter on to the top of the hook, Figures.Thin below the
// top of the sign, is drawn the same way.
//
// The lower edge of the stroke runs down the outside of the hook, a quarter
// of a super-ellipse with the squareness of the outer edge of o from the top
// to the left extreme (sectHookLeft left of O and sectHookY above it, the
// super-ellipse sectHookW wide), and from the left extreme into the eye in
// one cubic which leaves the quarter with its direction and curvature and
// joins the eye with its direction and curvature (g2Cubic), sectJoin of the
// way from the middle of the eye's upper long side towards its upper end;
// from there it follows the eye round its lower right end.  The outer edge of
// the lower stroke's arc crosses it above the eye's upper left end: the
// crotch on the left.
//
// Over the top the stroke runs on to the right as the hook of 6 does
// (openingCurve): both edges leave their tops with the curvature there and
// open to a circle of sectHookRadius as they turn through sectHookAngle, and
// the ball terminal of the figures (figBall), the part of the lowercase
// scaled to Figures.Ball, hangs from the hook
// with its rightmost ink sectBallIn inside the right extreme of the arc
// (roundTerminal).
//
// The outline is drawn directly.  The outer contour runs, for the upper half,
// from the crotch at the lower right up the outer edge of the arc and the
// spine and down the inner edge of the hook to its top, round the ball, and
// down the outer edge of the hook and the lower edge of the spine to the
// crotch at the upper left; the lower half is the same turned through half a
// turn.  The eye is a contour of its own.
//
// Side kinds: open, SectSide to the ink.
func section(p design.Params) (outline.Path, float64) {
	s := p.Sym
	t := p.Figures.Thin
	mid := (s.SectTop + s.SectBottom) / 2
	top := s.SectTop - mid // the frame has O at the origin
	sp := s.SectSpine

	// The frame of the eye: u along the spine, falling to the right, v across
	// it, up.
	th := sectSpineAngle * math.Pi / 180
	d := outline.Point{X: math.Cos(th), Y: -math.Sin(th)}
	n := outline.Point{X: math.Sin(th), Y: math.Cos(th)}
	toXY := func(q outline.Point) outline.Point {
		return outline.Point{X: q.X*d.X + q.Y*n.X, Y: q.X*d.Y + q.Y*n.Y}
	}
	turned := func(c cubic) cubic { return cubic{toXY(c.P0), toXY(c.P1), toXY(c.P2), toXY(c.P3)} }

	turn := func(c cubic) cubic { return cubic{neg(c.P0), neg(c.P1), neg(c.P2), neg(c.P3)} }
	a, b := s.SectEyeLong/2, s.SectEyeShort/2
	eye := newSuperShape(-a, -b, a, b, p.SupernessInner)
	// Clockwise, segment 3 runs from the top (the middle of the upper long
	// side) to the right end, 0 from there to the bottom, 1 on to the left.
	eyeA, eyeB := turned(eye.cubic(3, false)), turned(eye.cubic(0, false))
	// The lower edge of the spine leaves the eye sectJoin of the way from the
	// middle of its long side towards its upper left end.
	eyeJ := turned(eye.cubic(2, false)).sub(1-sectJoin, 1)

	// The outer edge of the arc has its right extreme sectArcSide times the
	// thin part outside that of the eye at the same height and its bottom the
	// thin part below that of the eye.
	var eR, eB outline.Point
	eR.X, eB.Y = math.Inf(-1), math.Inf(1)
	for _, pc := range curveChains(eyeA, eyeB, turn(eyeA), turn(eyeB)) {
		if q := pc.P0; q.X > eR.X {
			eR = q
		}
		if q := pc.P0; q.Y < eB.Y {
			eB = q
		}
	}
	right := outline.Point{X: eR.X + sectArcSide*t, Y: eR.Y}
	bot := outline.Point{X: eB.X, Y: eB.Y - t}
	// The middle of the upper edge of the spine, SectSpine from the middle of
	// the eye's long side, where the edge has its inflection; from there to
	// the right extreme and on to the bottom the edge is drawn as the edges of
	// s are (fromSpine, sideQuarter), and beyond the bottom it runs on as the
	// mirror image of its last quarter.
	spineMid := outline.Point{X: eyeA.P0.X + sp*n.X, Y: eyeA.P0.Y + sp*n.Y}
	fromSpine := func(right outline.Point) cubic {
		corner := outline.Point{X: right.X, Y: spineMid.Y + (right.X-spineMid.X)*d.Y/d.X}
		return cubic{spineMid, lerp(spineMid, corner, sectSpineHandle), corner, right}
	}
	// The height of the right extreme: that at which the quarter below it
	// takes the curvature of the spine's curve with the handle sectArcHandle
	// of its height (sideQuarter), found by bisection between the bottom and
	// the line of the spine.
	handle := func(y float64) float64 {
		q := outline.Point{X: right.X, Y: y}
		a, w := y-bot.Y, right.X-bot.X
		return math.Sqrt(endRadius(fromSpine(q))*w*(1-sShape.Quarter)/1.5) / a
	}
	yLo, yHi := bot.Y+1, spineMid.Y+(right.X-spineMid.X)*d.Y/d.X-1
	if !(handle(yLo) > sectArcHandle && handle(yHi) < sectArcHandle) {
		panic("glyphs: glyph section: the arc cannot be drawn; glyphs.Check reports such parameters")
	}
	for i := 0; i < 60; i++ {
		if m := (yLo + yHi) / 2; handle(m) > sectArcHandle {
			yLo = m
		} else {
			yHi = m
		}
	}
	right.Y = (yLo + yHi) / 2
	arcA := fromSpine(right)
	arcB := sideQuarter(right, bot, sShape.Quarter, endRadius(arcA))
	arcC := mirrorCubic(arcB, bot.X)

	// The hook.
	xl, yl := -sectHookLeft, sectHookY
	so := newSuperShape(xl, 2*yl-top, xl+sectHookW, top, p.SupernessOuter)
	yi := yl + sectHookInY
	// The inner edge is drawn as the edges of s are: from the middle of the
	// spine to the inner left extreme with the corner handle (the curve has
	// its inflection at the spine), and on to the top with the handle of the
	// quarters of s at the top and the curvature of the curve before it at
	// the side (sideQuarter, turned).
	inLeft := outline.Point{X: xl + s.SectHookThick, Y: yi}
	inTopPt := outline.Point{X: xl + (s.SectHookThick+sectHookW-t)/2, Y: top - t}
	inCorner := outline.Point{X: inLeft.X, Y: spineMid.Y + (inLeft.X-spineMid.X)*d.Y/d.X}
	inSpine := cubic{spineMid, lerp(spineMid, inCorner, sectSpineHandle), inCorner, inLeft}.reversed()
	inTop := turn(sideQuarter(neg(inLeft), neg(inTopPt), sShape.Quarter, endRadius(inSpine.reversed()))).reversed()
	outTop := so.cubic(1, true) // the top to the left extreme
	down := outline.Point{Y: -1}
	outSpine := spineCurve(outTop.P3, down, outTop.curvature(1), eyeJ)

	// The edges of the upper stroke, walked from the top: the lower (outer
	// edge of the hook, lower edge of the spine, the eye) and the upper (inner
	// edge of the hook, upper edge of the spine, outer edge of the arc).
	lower := []cubic{outTop, outSpine, eyeJ, eyeA, eyeB}
	upper := []cubic{inTop, inSpine, arcA, arcB, arcC}

	// The crotch at the lower right: the first place along the upper edge
	// where it crosses the lower edge of the other stroke.
	iU, tU, iL, tL := -1, 0.0, -1, 0.0
	for i := 1; i < len(upper) && iU < 0; i++ {
		for j := 1; j < 4; j++ {
			if ta, tb, ok := crossing(upper[i], turn(lower[j])); ok {
				iU, tU, iL, tL = i, ta, j, tb
				break
			}
		}
	}
	if iU < 0 {
		panic("glyphs: glyph section: the arc does not meet the other spine; glyphs.Check reports such parameters")
	}
	var up, lo chain
	for i := 0; i < iU; i++ {
		up = join(up, curveChain(upper[i]))
	}
	// Both cuts end on the same point, the crotch.
	q := upper[iU].at(tU)
	cu, cl := upper[iU].sub(0, tU), lower[iL].sub(0, tL)
	cu.P3, cl.P3 = q, neg(q)
	up = join(up, curveChain(cu))
	for i := 0; i < iL; i++ {
		lo = join(lo, curveChain(lower[i]))
	}
	lo = join(lo, curveChain(cl))

	// The hook over the top and the ball.
	turnTo := sectHookAngle * math.Pi / 180
	ho := openingCurve("section", outTop.P0, outTop.curvature(0), 1/sectHookRadius, turnTo).reversed()
	hi := openingCurve("section", inTop.P0, inTop.curvature(0), 1/(sectHookRadius-t), turnTo)
	bp := figBall(p)
	_, _, eyeRight, _ := chainBBox(curveChains(arcA, arcB))
	ball := sectTerminal("section", bp, ho, hi, eyeRight-sectBallIn)

	half := join(up.reversed(), ball, lo)
	outer := join(half, half.rotated180(outline.Point{}))
	counter := curveChains(eyeA, eyeB, turn(eyeA), turn(eyeB))

	// Place the ink SectSide from the left of the advance.
	x0, _, x1, _ := chainBBox(outer)
	w := math.Round(x1 - x0)
	dx := s.SectSide - x0 - (w-(x1-x0))/2
	return chainsPath(outer.shifted(dx, mid), counter.shifted(dx, mid)), w + 2*s.SectSide
}

// The shape constants of the section sign (see above): the fall of the spines
// (degrees); the outer left extreme of the hook, left of and above O, the
// width of its super-ellipse and the height of the inner extreme above the
// outer; the radius at the end of the hook and the angle through which it
// turns beyond the top (degrees); how far the rightmost ink of the ball lies
// inside the right extreme of the arc; where the lower edge of the spine
// joins the eye; the outer right extreme of the arc, in thin parts outside
// the eye, and the handle which sets its height; the handles of the upper
// edge at the spine.
const (
	sectSpineAngle  = 40.0
	sectHookLeft    = 120.0
	sectHookY       = 230.0
	sectHookW       = 290.0
	sectHookRadius  = 240.0
	sectHookAngle   = 75.0
	sectJoin        = 0.2
	sectHookInY     = 70.0
	sectArcSide     = 1.2
	sectArcHandle   = 0.55
	sectSpineHandle = 0.6
	sectBallIn      = 22.0
)

// spineCurve returns the curve from the left extreme p0 of a hook, which it
// leaves in the direction t0 with the curvature k0, into the start of the
// quarter q in the middle of a spine, which it meets with the direction and
// the curvature of q.
func spineCurve(p0, t0 outline.Point, k0 float64, q cubic) cubic {
	if c, err := g2Cubic(p0, t0, k0, q.P0, q.tangent(0), q.curvature(0)); err == nil {
		return c
	}
	return nearG2(p0, t0, k0, q.P0, q.tangent(0), q.curvature(0))
}

// curveChains returns the curves as one chain, each cut at its extremes.
func curveChains(cs ...cubic) chain {
	var out chain
	for _, c := range cs {
		out = join(out, curveChain(c))
	}
	return out
}

// sectTerminal returns the end of a round stroke of the section sign in the
// ball terminal.  The curve outer is the outer edge of the hook from its right
// extreme up to its top, and inner the inner edge from its top to its right
// extreme, as the contour walks them; the part is turned so that the rightmost
// ink of the ball lies at right, with the centre of the ball on an integer
// abscissa.  The chain runs from the top of the inner edge along it and the inner
// edge of the stalk, round the ball with its stalk as the part draws them, and
// back along the outer edge to its top.  It keeps the stalk of the part and
// does not wind the stroke into the ball on a spiral as roundTerminal does.
func sectTerminal(name string, b ballPart, outer, inner cubic, right float64) chain {
	x1 := func(angle float64) float64 {
		tb, _ := roundTerminalAt(b, outer, angle)
		_, _, x1, _ := chainBBox(tb.Chain)
		return x1
	}
	solve := func(target float64) float64 {
		lo, hi := figBallAngles[0], figBallAngles[1]
		if !(x1(lo) >= target && x1(hi) <= target) {
			panic(fmt.Sprintf("glyphs: glyph %s: the ball cannot be placed with its ink at %.1f (%.1f to %.1f); glyphs.Check reports such parameters", name, target, x1(hi), x1(lo)))
		}
		for i := 0; i < 60; i++ {
			mid := (lo + hi) / 2
			if x1(mid) > target {
				lo = mid
			} else {
				hi = mid
			}
		}
		return (lo + hi) / 2
	}
	tb, _ := roundTerminalAt(b, outer, solve(right))
	target := right + math.Round(tb.Centre.X) - tb.Centre.X
	tb, tO := roundTerminalAt(b, outer, solve(target))

	best, in, tI := math.Inf(1), cubic{}, 0.0
	for _, free := range []bool{false, true} {
		for _, u := range cCuts() {
			fi := stalkFoot{At: inner.at(u), Dir: inner.tangent(u), Kappa: inner.curvature(u)}
			if free {
				fi.Kappa = math.NaN()
			}
			if c, err := edgeIn(fi, tb.In); err == nil {
				if e := curvatureRate(c); e < best {
					best, in, tI = e, c, u
				}
			}
		}
		if !math.IsInf(best, 1) {
			break
		}
	}
	if math.IsInf(best, 1) {
		panic(fmt.Sprintf("glyphs: glyph %s: no inner edge of the stalk reaches the ball terminal; glyphs.Check reports such parameters", name))
	}
	return join(chain{curveSeg(inner.sub(0, tI))}, splitAtExtrema(in), tb.Chain, chain{curveSeg(outer.sub(tO, 1))})
}
