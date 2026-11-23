// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter y.
//
// The y is the v whose thin diagonal goes on below the baseline into a tail which
// curves to the left and curls up into the ball terminal (ball.go).  The upper part
// is that of v (the same diagonals, the same lean, the same serifs), but the vertex
// is not drawn: the thick diagonal ends where it runs into the thin one.  Its outer
// edge, the left-hand edge, meets the left-hand edge of the thin diagonal at the
// foot of the thick stroke, a few units above the baseline, and from there the
// left-hand edge of the thin diagonal is the inner edge of the tail; the crotch
// above, where the inner edges meet, is the junction (design.VCrotch).
//
// The tail.  The straight part of the tail is the thin diagonal itself, down to the
// heights YOuterDrop and YInnerDrop above the lowest point; there both edges leave
// the diagonal in its direction and are each one cubic curve to the section of the
// terminal, with the handle at the diagonal which makes them fairest (yTail).  The
// inner edge leaves the diagonal without curvature: over the first stemRamp degrees
// of its turn it is a transition curve (straightFoot) from the straight diagonal
// into that cubic, as the edges of f and r leave their stems (windHook).  The
// terminal is turned as that of j (yBallAngle); its lowest ink lies YDepth below the
// baseline and its leftmost ink YReach left of the left-hand edge of the diagonal at
// that depth (yTerminal).  The tail is one round sweep, down the diagonal, round
// to the left and up into the ball, with no level or straight stretch.  The
// leftmost ink is the ball, left of the serif of the thick diagonal and inside the
// advance: side kind sideDiagTail.  The contour is one counter-clockwise outline:
//
//	the outer serif of the thick diagonal, down its outer edge to the foot of the
//	thick stroke, down the left-hand edge of the thin diagonal into the tail:
//	the inner edge, the terminal and the outer edge, which ends on the right-hand
//	edge of the thin diagonal, up that edge, its outer serif, along the top to its
//	inner serif, down its left-hand edge to the crotch, up the inner edge of the
//	thick diagonal, its inner serif and along the top.
//
// Side kinds: the right side is that of the diagonal letters, DiagSideRight to the serif;
// the left side is sideDiagTail: the serif of the thick diagonal stands DiagSideLeft from the
// edge of the advance and the end of the tail may reach the edge.

// yBallAngle is the angle by which the ball terminal of y is turned from its own
// frame (radians): that of j.
var yBallAngle = jBallAngle

// yFrame returns the vertex of the upper part, which y shares with v.
func yFrame(p design.Params) vertex { return vVertex(p) }

// yCentre returns the position of the centre line of the thin diagonal at the height
// of the lowest point of the tail.
func yCentre(p design.Params) float64 { return yFrame(p).Right.centreAt(-p.YDepth) }

// yHook returns the tail as a hook in the frame turned through half a turn about the
// point of the centre line of the thin diagonal at the depth of the lowest point:
// the edges of the diagonal at that depth are the "stem", the joins are the heights
// YOuterDrop and YInnerDrop above the lowest point, and Reach is measured from the
// left-hand edge of the diagonal, which in that frame is on the right.
func yHook(p design.Params) hookSpec {
	r := yFrame(p).Right
	cx, half := yCentre(p), r.width()/2
	return hookSpec{
		XL: cx - half, XR: cx + half, Cell: cx + half,
		Apex: p.YDepth, Reach: p.YReach,
		OuterJoin: p.YDepth - p.YOuterDrop, InnerJoin: p.YDepth - p.YInnerDrop,
	}
}

// yTail returns the tail as one chain with the ball terminal tb, from the left-hand
// edge of the thin diagonal, where the transition curve leaves it a little above the
// point YInnerDrop above the lowest point, along the inner edge, round the ball and
// back along the outer edge to the right-hand edge of the diagonal, YOuterDrop above
// the lowest point.  Both edges leave the diagonal in its direction.
func yTail(p design.Params, tb placedBall) chain {
	h := yHook(p)
	r, d := yFrame(p).Right, yFrame(p).Right.down()
	dn := math.Atan2(d.Y, d.X)
	in := footRising(r.edgeAt(edgeLeft, -h.InnerJoin), dn)
	a, err := edgeIn(in, tb.In)
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph y: the edge into the ball terminal: %v", err))
	}
	u := tTurned(a, 0, stemRamp)
	ramp := straightFoot(a.at(u), neg(a.tangent(u)), -a.curvature(u), in.At, neg(in.Dir)).reversed()
	b, err := edgeOut(tb.Out, footAt(r.edgeAt(edgeRight, -h.OuterJoin), dn+math.Pi))
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph y: the edge out of the ball terminal: %v", err))
	}
	return join(chain{curveSeg(ramp)}, splitAtExtrema(a.sub(u, 1)), tb.Chain, splitAtExtrema(b))
}

// yTerminal returns the ball terminal of y: its leftmost ink YReach left of the
// left-hand edge of the thin diagonal at the depth of the lowest point, which is
// YDepth below the baseline.
func yTerminal(p design.Params) placedBall {
	h := yHook(p)
	left := 2*yCentre(p) - (h.Cell + h.Reach)
	return must(placeAtExtreme(newBallPart(p), ballPlace{Angle: yBallAngle},
		func(x0, _ float64) float64 { return left - x0 }, -p.YDepth, false,
		func(tb placedBall) chain { return yTail(p, tb) }))
}

// y draws the lowercase letter y, see the comment above.
func y(p design.Params) (outline.Path, float64) {
	return yOutline(p), design.YAdvance(p)
}

// yOutline draws the letter y.
func yOutline(p design.Params) outline.Path {
	vx := yFrame(p)
	l, r := vx.Left, vx.Right
	top := p.XHeight
	thick, thin := diagSerifs(p)
	outL := l.headServif(edgeLeft, top, thick)
	outR := r.headServif(edgeRight, top, thin)
	inR := r.headServif(edgeLeft, top, thin)
	inL := l.headServif(edgeRight, top, thick)
	tail := yTail(p, yTerminal(p))
	foot := edgeMeet(l, edgeLeft, r, edgeLeft)
	c := join(
		outL,
		lineChain(outL.end(), foot),
		lineChain(foot, tail.start()),
		tail,
		lineChain(tail.end(), outR.end()),
		outR.reversed(),
		lineChain(outR.start(), inR.start()),
		inR,
		lineChain(inR.end(), vx.crotch()),
		lineChain(vx.crotch(), inL.end()),
		inL.reversed(),
		lineChain(inL.start(), outL.start()),
	)
	return contourPath(c)
}
