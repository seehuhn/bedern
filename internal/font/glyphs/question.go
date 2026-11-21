// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
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
	ball := ballStart("question", p, right, left, f.x0+r, span(0.80*p.Sym.QTop, 0.86*p.Sym.QTop, 0.01*p.Sym.QTop), span(-60, 40, 2), true)
	hook := join(right, lineChain(right.end(), left.start()), left, ball)
	return chainsPath(hook, dotChain(p, f.cx, p.DotDiameter/2)), f.x1 + p.Sym.QSide
}
