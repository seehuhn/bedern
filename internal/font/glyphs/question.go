// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"errors"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The question mark.
//
// A hook over a dot.  The dot is the dot of i (DotDiameter), standing on the
// baseline, as the full stop does.  The hook is one stroke (symstroke.go) which
// begins at the upper left in the ball terminal of the lowercase (ball.go),
// rises over the top, where it is thin (Figures.Thin), swells round the right-hand
// side to 0.9 of the round stroke of the lowercase (CurveLower) at the right
// extreme, where its weight lies, as on the rounds of the face (vertical
// stress), and sweeps down and to the left, thinner (0.64 of CurveLower),
// through an inflection, where both edges are straight for a moment
// (qInflection), and bends back in one curve into a short straight stem
// (qStemRun) above the dot, QStem thick, with its flat foot QGap above the dot.
// So each edge, from the right extreme to the foot, bends one way, passes
// through one inflection and bends the other way, with the curvature matched at
// every station and no angle where the sweep becomes the stem.
// The ball hangs from the stroke as on the lowercase: the part is mirrored, so
// that the stalk comes over the top from the right and curls down into the ball,
// and it is turned so that its two edges meet the stroke at the top extreme as
// fairly as they can (qBall).
//
// Side kinds: open on both sides; the side spaces are QSide to the leftmost ink
// (the ball) and to the rightmost ink (the right extreme of the hook).

// qFrame holds the measures of the question mark.
type qFrame struct {
	x0, x1 float64 // the leftmost and the rightmost ink
	cx     float64 // the centre line of the stem and of the dot
}

func newQFrame(p design.Params) qFrame {
	s := p.Sym
	x0 := s.QSide
	return qFrame{x0: x0, x1: x0 + s.QInk, cx: math.Round(x0 + 0.43*s.QInk)}
}

// qInflection is the curvature given to the centre line of the hook where it
// turns from bending right to bending left: nearly none, and both edges pass
// straight through the inflection and bend into the straight stem in one
// curve which has no curvature at either end.
const qInflection = 1e-7

// qCurlJoin is the angle (degrees) round the ball from its leftmost point at
// which the curl of the terminal joins its circle.
const qCurlJoin = 20.0

// qStemRun is the length of the straight foot of the stem.
const qStemRun = 30.0

// qStations returns the centre line of the hook, from its top to the foot of
// the stem.
func qStations(p design.Params, f qFrame) []station {
	s := p.Sym
	thick := p.CurveLower
	foot := p.DotDiameter + s.QGap
	return []station{
		{X: f.x0 + 0.50*s.QInk, Y: s.QTop - p.Figures.Thin/2, Dir: 0, W: p.Figures.Thin},
		{X: f.x1 - 0.45*thick, Y: 0.75 * s.QTop, Dir: -90, W: 0.9 * thick},
		{X: f.x0 + 0.58*s.QInk, Y: 0.49 * s.QTop, Dir: -128, W: 0.64 * thick, K: qInflection},
		{X: f.cx, Y: foot + qStemRun, Dir: -90, W: s.QStem, Flat: true},
		{X: f.cx, Y: foot, Dir: -90, W: s.QStem},
	}
}

// question draws the question mark, see the comment above.
func question(p design.Params) (outline.Path, float64) {
	f := newQFrame(p)
	sts := qStations(p, f)
	right, left := strokeOpen(sts)
	r := p.BallDiameter / 2
	fits := func(ball chain, tb placedBall) bool {
		_, err := qTerminalErr(ball, tb, right)
		return err == nil
	}
	ball, tb := ballStartPlaced("question", p, right, left, f.x0+r, span(0.80*p.Sym.QTop, 0.86*p.Sym.QTop, 0.01*p.Sym.QTop), span(-60, 40, 2), true, fits)
	hook := join(right, lineChain(right.end(), left.start()), left, qTerminal(ball, tb, right))
	return chainsPath(hook, dotChain(p, f.cx, p.DotDiameter/2)), f.x1 + p.Sym.QSide
}

// qTerminal returns the ball terminal of the hook on the ball tb, whose chain
// with its stalk is ball (ballStartPlaced): the outer edge of the stalk and the
// curl of the part to the start of its last curve before its leftmost point;
// from there one cubic curve, with the direction and the curvature of the curl,
// past the leftmost point to the circle of the ball qCurlJoin beyond its
// leftmost point, which it meets with the direction and the curvature of the
// circle (g2Cubic); the circle to the tip of the wedge; and from the tip the
// inner edge, which leaves with the direction and the curvature of the inner
// edge of the part and meets the start of the right-hand edge of the stroke
// with its direction and its curvature (edgeOut), or, where no such curve
// exists, leaves the tip with the handle of a circular arc (g2End).
func qTerminal(ball chain, tb placedBall, right chain) chain {
	c, err := qTerminalErr(ball, tb, right)
	if err != nil {
		panic(err.Error())
	}
	return c
}

// qTerminalErr is qTerminal which returns an error where no curve joins the
// parts.
func qTerminalErr(ball chain, tb placedBall, right chain) (chain, error) {
	r0 := right[0].cubic
	c, r := tb.Centre, tb.Part.R
	iL, iT := -1, -1
	for i, pc := range ball {
		if iL < 0 || pc.P3.X < ball[iL].P3.X {
			iL = i
		}
		if iT < 0 && isTip(pc.P3, tb.Tip) {
			iT = i
		}
	}
	if iT < 0 {
		return nil, errors.New("glyphs: glyph question: the ball terminal has no tip")
	}
	// The curl joins the circle qCurlJoin past its leftmost point.
	aJ := math.Pi + qCurlJoin*math.Pi/180
	join0 := outline.Point{X: c.X + r*math.Cos(aJ), Y: c.Y + r*math.Sin(aJ)}
	l := ball[iL].cubic
	curl, err := g2Cubic(l.P0, l.tangent(0), l.curvature(0), join0, dirOf(aJ+math.Pi/2), 1/r)
	if err != nil {
		return nil, errors.New("glyphs: glyph question: the curl cannot reach the ball; glyphs.Check reports such parameters")
	}
	aTip := math.Atan2(tb.Tip.Y-c.Y, tb.Tip.X-c.X)
	for aTip <= aJ {
		aTip += 2 * math.Pi
	}
	circ := arcAxes(c, r, aJ, aTip)
	circ[len(circ)-1].P3 = tb.Tip
	var tip stalkEnd
	for _, pc := range ball {
		if isTip(pc.P0, tb.Tip) {
			tip = stalkEnd{At: tb.Tip, Dir: pc.tangent(0), Kappa: pc.curvature(0)}
		}
	}
	foot := stalkFoot{At: r0.P0, Dir: r0.tangent(0), Kappa: r0.curvature(0)}
	in, err := edgeOut(tip, foot)
	if err != nil {
		// The tip is a corner: the curve may leave it with any curvature.
		in, err = g2End(tip.At, tip.Dir, arcHandle(tip.At, tip.Dir, foot.At, foot.Dir), foot.At, foot.Dir, foot.Kappa)
		if err != nil {
			return nil, errors.New("glyphs: glyph question: the inner edge cannot reach the ball terminal; glyphs.Check reports such parameters")
		}
	}
	out := append(chain(nil), ball[:iL]...)
	out = append(out, splitAtExtrema(curl)...)
	return join(out, circ, splitAtExtrema(in)), nil
}
