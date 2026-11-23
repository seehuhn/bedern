// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capJ draws the capital J: the capital stem with head serifs, SerifJut on the
// left and JHeadRight on the right, which comes down to the baseline and curves
// to the left into a tail which ends in the ball terminal (ball.go), as the tail
// of j does.  J sits on the baseline: the lowest point of the tail lies
// OvershootUpper below it, as the round capitals do, and the leftmost ink, the
// ball, lies JReach left of the left-hand edge of the stem.  The left-hand edge
// of the stem begins to curve JInnerDrop above the lowest point and the
// right-hand edge JOuterDrop above it; each edge is one cubic curve from the stem,
// which it leaves vertically, to the section of the terminal, as on j, and the
// stroke swells as it turns from the stem into the tail and thins into the stalk.
// The ball is the part of the lowercase, JBall across on the stalk BallStalk,
// turned by capJBallAngle.
//
// The contour runs from the right-hand edge where it begins to curve up to the
// head serifs, along the top, down the left-hand edge, along the inner edge of the
// tail to the ball, round it, and back along the outer edge to the start.
//
// Side kinds: the ball on the left, JSideLeft from the edge, and a capital stem
// on the right.
func capJ(p design.Params) (outline.Path, float64) {
	c, g := p.Caps, p.Caps.CGSJQ
	s := capStem{Left: design.CapJStemLeft(p), HeadLeft: c.SerifJut, HeadRight: g.JHeadRight}
	low := -p.OvershootUpper
	outerY, innerY := low+g.JOuterDrop, low+g.JInnerDrop
	from := outline.Point{X: s.right(p), Y: outerY}
	top := s.top(p, from, innerY)
	tb := capJTerminal(p, s)
	return capContour(join(top, capJTail(p, s, tb))), design.CapJAdvance(p)
}

// capJBallAngle is the turn of the ball terminal of J (radians): that of j.
var capJBallAngle = jBallAngle + 20*math.Pi/180

// capJBall returns the ball terminal of J.
func capJBall(p design.Params) ballPart {
	b, err := makeBallPart(p.Caps.CGSJQ.JBall/2, p.BallStalk, ballCurl, ballSection, ballBlend)
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph J: %v; glyphs.Check reports such parameters", err))
	}
	return b
}

// capJTail returns the tail of J with the ball terminal tb: from the left-hand
// edge of the stem, where it begins to curve, along the inner edge, round the
// ball and back along the outer edge to the right-hand edge.
func capJTail(p design.Params, s capStem, tb placedBall) chain {
	g := p.Caps.CGSJQ
	low := -p.OvershootUpper
	return stalkChain("J",
		footRising(outline.Point{X: s.Left, Y: low + g.JInnerDrop}, -math.Pi/2),
		tb,
		footAt(outline.Point{X: s.right(p), Y: low + g.JOuterDrop}, math.Pi/2))
}

// capJTerminal returns the ball terminal of J: its leftmost ink JReach left of
// the stem, and the lowest point of the tail OvershootUpper below the baseline.
func capJTerminal(p design.Params, s capStem) placedBall {
	return must(placeAtExtreme(capJBall(p), ballPlace{Angle: capJBallAngle},
		func(x0, _ float64) float64 { return p.Caps.CGSJQ.JSideLeft - x0 }, -p.OvershootUpper, false,
		func(tb placedBall) chain { return capJTail(p, s, tb) }))
}
