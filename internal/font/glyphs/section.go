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
// The frame of the eye is a super-ellipse with the squareness of the inner
// edge of o, SectEyeLong long along the spines and SectEyeShort across them,
// about O and turned to the angle of the spines; the middles of its long sides
// and its extremes place the parts.  The eye itself is a rounded lozenge whose
// long sides are the inner edges of the two spines and whose short ends are
// the inner edges of the thin arcs.  Each half of it, from the middle of one
// long side round one end to the middle of the other, is three cubic curves
// with the points, directions and curvatures given at their ends (g2Cubic):
// from the middle of the upper long side, in the direction of the spine with
// the curvature of the super-ellipse there, to the point ThinLower inside the
// outer edge of the arc where that edge has turned a quarter turn from the
// direction of the spine; on to the bottom of the eye, ThinLower above the
// bottom of the arc; and on to the middle of the lower long side.  At the two
// inner points the eye runs in the direction of the outer edge with the
// curvature of a circle concentric with it.  So each arc leaves its own spine
// with the spine's weight, falls to the thin weight by the middle of its turn
// and keeps it until it runs into the other spine.  The eye is one convex
// curve.
//
// The upper edge of the stroke is drawn as the edges of s are (s.go).  It has
// its inflection at the middle of the spine, SectSpine from the middle of the
// eye's upper long side at right angles to the spine, where it runs along the
// spine without curvature.  From there to each side extreme it is one cubic
// whose first handle lies on the line of the spine, sectSpineHandle of the
// way to its second control point, the corner where that line meets the
// vertical through the extreme.  On the right the extreme lies sectArcSide
// times Figures.Thin outside the frame's right extreme, and the quarter on to
// the bottom, Figures.Thin below the frame's bottom, has the handle of the
// quarters of s at the bottom and at the side the curvature of the curve
// before it (sideQuarter); the height of the side extreme is the one at which
// that handle is sectArcHandle of the quarter's height.  Beyond the bottom the
// edge turns on into the direction of the spine in one cubic, which leaves the
// bottom with its curvature and ends straight ThinLower below the middle of
// the eye's lower long side, and crosses the outer edge of the lower stroke's
// spine at a small angle, where the white beside the eye ends in a narrow
// crotch.  On the left the extreme is the inner side of the hook,
// SectHookThick right of its outer side and sectHookInY above its outer
// extreme, and the quarter on to the top of the hook, Figures.Thin below the
// top of the sign, is drawn the same way.
//
// The lower edge of the stroke runs down the outside of the hook, a quarter
// of a super-ellipse with the squareness of the outer edge of o from the top
// to the left extreme (sectHookLeft left of O and sectHookY above it, the
// super-ellipse sectHookW wide), and from the left extreme towards the middle
// of the eye's upper long side in one cubic which leaves the quarter with its
// direction and curvature and ends there in the direction of the spine with
// sectJoinBend of the curvature of the frame there (g2Cubic).  The outer edge
// of the lower stroke's arc crosses it above the eye's upper left end: the
// crotch on the left.

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

	// The outer edge of the hook, a quarter of a super-ellipse from the top to
	// the left extreme.
	xl, yl := -sectHookLeft, sectHookY
	so := newSuperShape(xl, 2*yl-top, xl+sectHookW, top, p.SupernessOuter)
	outTop := so.cubic(1, true) // the top to the left extreme
	down := outline.Point{Y: -1}
	// The lower edge of the spine runs from the outer left extreme of the hook
	// in one cubic which leaves the quarter with its direction and curvature
	// and ends in the middle of the eye's upper long side, in the direction of
	// the spine, with sectJoinBend of the curvature of the super-ellipse of
	// the eye there (g2Cubic, drawn from its end); beside the eye it is hidden.
	sl, err := g2Cubic(eyeA.P0, neg(d), -sectJoinBend*eyeA.curvature(0), outTop.P3, neg(down), -outTop.curvature(1))
	if err != nil {
		panic("glyphs: glyph section: the lower edge of the spine cannot be drawn; glyphs.Check reports such parameters")
	}
	outSpine := cubic{sl.P3, sl.P2, sl.P1, sl.P0}

	// The outer edge of the arc has its right extreme sectArcSide times the
	// thin part outside that of the frame of the eye at the same height and
	// its bottom the thin part below that of the frame.
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
	// s are (fromSpine, sideQuarter).
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
	// Beyond the bottom the outer edge turns on into the direction of the
	// spine and runs, the thin weight below the middle of the eye's lower long
	// side, into the lower stroke's spine: one cubic curve which leaves the
	// bottom with the curvature there and ends straight (g2Cubic, drawn from
	// its end).
	w := p.ThinLower
	endC := outline.Point{X: -eyeA.P0.X - w*n.X, Y: -eyeA.P0.Y - w*n.Y}
	ac, err := g2Cubic(endC, d, 0, bot, outline.Point{X: 1}, -arcB.curvature(1))
	if err != nil {
		panic("glyphs: glyph section: the arc cannot run into the spine; glyphs.Check reports such parameters")
	}
	arcC := cubic{ac.P3, ac.P2, ac.P1, ac.P0}

	// The hook.
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

	// The eye's boundary from the middle of its upper long side round its
	// right end to the middle of its lower long side, in three cubic curves
	// (see above); the other half is the same turned through half a turn.
	tMid := 0.0
	for lo, hi, k := 0.0, 1.0, 0; k < 60; k++ {
		tMid = (lo + hi) / 2
		if g := arcB.tangent(tMid); math.Atan2(-g.X, -g.Y) > th {
			hi = tMid
		} else {
			lo = tMid
		}
	}
	conc := func(k float64) float64 { return k / (1 + w*k) }
	inside := func(c cubic, tt float64) (outline.Point, outline.Point, float64) {
		q, g := c.at(tt), c.tangent(tt)
		return outline.Point{X: q.X + w*g.Y, Y: q.Y - w*g.X}, g, conc(c.curvature(tt))
	}
	g2or := func(p0, t0 outline.Point, k0 float64, p3, t3 outline.Point, k3 float64) cubic {
		if k3 == 0 {
			if c, err := g2Cubic(p3, neg(t3), 0, p0, neg(t0), -k0); err == nil {
				return cubic{c.P3, c.P2, c.P1, c.P0}
			}
		} else if c, err := g2Cubic(p0, t0, k0, p3, t3, k3); err == nil {
			return c
		}
		return nearG2(p0, t0, k0, p3, t3, k3)
	}
	pM, gM, kM := inside(arcB, tMid)
	pB, gB, kB := inside(arcB, 1)
	kT := eyeA.curvature(0)
	eyeHalf := curveChains(
		g2or(eyeA.P0, d, kT, pM, gM, kM),
		g2or(pM, gM, kM, pB, gB, kB),
		g2or(pB, gB, kB, neg(eyeA.P0), neg(d), kT),
	)
	// The edges of the upper stroke, walked from the top: the lower (outer
	// edge of the hook, lower edge of the spine) and the upper (inner edge of
	// the hook, upper edge of the spine, outer edge of the arc).
	lower := []cubic{outTop, outSpine}
	upper := []cubic{inTop, inSpine, arcA, arcB, arcC}

	// The crotch at the lower right: the first place along the upper edge
	// where it crosses the lower edge of the other stroke.
	iU, tU, iL, tL := -1, 0.0, -1, 0.0
	for i := 1; i < len(upper) && iU < 0; i++ {
		for j := 1; j < len(lower); j++ {
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
	counter := join(eyeHalf, eyeHalf.rotated180(outline.Point{}))

	// Place the ink SectSide from the left of the advance.
	x0, _, x1, _ := chainBBox(outer)
	width := math.Round(x1 - x0)
	dx := s.SectSide - x0 - (width-(x1-x0))/2
	return chainsPath(outer.shifted(dx, mid), counter.shifted(dx, mid)), width + 2*s.SectSide
}

// The shape constants of the section sign (see above): the fall of the spines
// (degrees); the outer left extreme of the hook, left of and above O, the
// width of its super-ellipse and the height of the inner extreme above the
// outer; the radius at the end of the hook and the angle through which it
// turns beyond the top (degrees); how far the rightmost ink of the ball lies
// inside the right extreme of the arc; the share of the curvature of the
// frame of the eye with which the lower edge of the spine ends; the outer
// right extreme of the arc, in thin parts outside the frame, and the handle
// which sets its height; the handles of the upper edge at the spine.
const (
	sectSpineAngle  = 40.0
	sectHookLeft    = 120.0
	sectHookY       = 230.0
	sectHookW       = 290.0
	sectHookRadius  = 240.0
	sectHookAngle   = 75.0
	sectJoinBend    = 0.3
	sectHookInY     = 70.0
	sectArcSide     = 1.2
	sectArcHandle   = 0.55
	sectSpineHandle = 0.6
	sectBallIn      = 22.0
)

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
// abscissa.  The chain runs from the top of the inner edge along it to a cut,
// where the inner edge of the stalk of the part would leave it most evenly, and
// from there in one curve to the tip of the wedge, round the ball with the
// outer edge of its stalk as the part draws them, and back along the outer edge
// to its top.  It keeps the stalk of the part and
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

	best, tI := math.Inf(1), 0.0
	for _, free := range []bool{false, true} {
		for _, u := range cCuts() {
			fi := stalkFoot{At: inner.at(u), Dir: inner.tangent(u), Kappa: inner.curvature(u)}
			if free {
				fi.Kappa = math.NaN()
			}
			if c, err := edgeIn(fi, tb.In); err == nil {
				if e := curvatureRate(c); e < best {
					best, tI = e, u
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
	// From the cut the inner edge runs in one curve to the tip of the wedge,
	// with the direction and the curvature of the round at the cut and with the
	// direction of the inner edge of the part and the handle of a circular arc
	// at the tip.
	tip := tipEnd(tb)
	q := inner.at(tI)
	in, err := g2Start(q, inner.tangent(tI), inner.curvature(tI), tip.At, tip.Dir, arcHandle(q, inner.tangent(tI), tip.At, tip.Dir))
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph %s: no inner edge reaches the tip of the wedge; glyphs.Check reports such parameters", name))
	}
	var rest chain
	for i, pc := range tb.Chain {
		if isTip(pc.P0, tb.Tip) {
			rest = tb.Chain[i:]
			break
		}
	}
	return join(chain{curveSeg(inner.sub(0, tI))}, splitAtExtrema(in), rest, chain{curveSeg(outer.sub(tO, 1))})
}
