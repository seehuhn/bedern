// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// TestPlacementReturnsErrors asks that the placement of a ball terminal at an extreme
// of the stroke which a letter draws to it returns an error, and does not panic,
// where the construction of the stroke fails or the stroke does not settle at the
// height.
func TestPlacementReturnsErrors(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	b := newBallPart(p)
	pl := ballPlace{Angle: jBallAngle}
	left := func(x0, _ float64) float64 { return 100 - x0 }
	fixed := func(tb placedBall) chain { return chain{lineSeg(outline.Point{}, outline.Point{X: 1})} }
	if _, err := placeAtExtreme(b, pl, left, -200, false, func(placedBall) chain { panic("no stroke") }); err == nil {
		t.Error("a stroke which cannot be drawn gave no error")
	}
	if _, err := placeAtExtreme(b, pl, left, -200, false, fixed); err == nil {
		t.Error("a stroke which does not follow the ball gave no error")
	}
	if _, err := placeUnderCrest(b, fBallAngle, 300, 700, func(placedBall) chain { panic("no stroke") }); err == nil {
		t.Error("a crest which cannot be drawn gave no error")
	}
	// A ball placed for the regular parameters settles.
	jLeft := func(x0, _ float64) float64 { return design.JInkLeft(p) - x0 }
	if _, err := placeAtExtreme(b, pl, jLeft, -p.JDepth, false, func(tb placedBall) chain { return jTailChain(p, tb) }); err != nil {
		t.Errorf("the tail of j: %v", err)
	}
}

// TestBallTerminalsFarFromTheRegularParametersReturnErrors asks that the ball of c, the
// spiral into the ball of r and the terminals of the figures return an error, and
// do not panic, for parameters at which they cannot be drawn.
func TestBallTerminalsFarFromTheRegularParametersReturnErrors(t *testing.T) {
	t.Parallel()
	// The ball of c, too large to be placed with its ink at the tip of the lower
	// terminal.
	p := design.Regular()
	p.BallDiameter += 45
	f := newCFrame(p)
	if _, err := figBallPlace("c", newBallPart(p), f.outerU.cubic(0, true), f.tipX-p.CBallInset); err == nil {
		t.Error("the ball of c, 45 units too large, was placed")
	}

	// The spiral into the ball of r, with the ball too near the stem.
	p = design.Regular()
	p.RReach -= 12
	feet, notch := rArmFeet(p)
	if _, err := windHook("r", feet, rTerminal(p), notch, rWindRises); err == nil {
		t.Error("the spiral into the ball of r, with the ball 12 units nearer, was drawn")
	}

	// The terminals of 2, 3, 5, 6 and 9, with a stalk too thick for the ball.
	p = design.Regular()
	p.BallStalk += 20
	if _, err := figBallPart(p); err == nil {
		t.Error("a stalk 20 units too thick fitted the ball of the figures")
	}
	p = design.Regular()
	p.Figures.CurlStalk += 40
	if _, err := curlBallPart(p); err == nil {
		t.Error("a curl stalk 40 units too thick fitted the ball of 6 and 9")
	}

	// The lower terminal of 3, with a stalk too thin for its spiral.
	p = design.Regular()
	p.BallStalk -= 12.5
	f3 := p.Figures
	x0 := design.FigureLeft(p, 3)
	x1 := x0 + f3.Ink[3]
	bot, w := figBottom(p), f3.ThreeWaist
	lo := newSuperShape(x0+f3.ThreeLowerLeft, bot, x1, w+f3.ThreeOverlap, p.SupernessOuter)
	li := newSuperShape(x0+f3.ThreeLowerLeft+f3.ThreeOpen, bot+f3.Thin, x1-f3.Bowl, w-f3.ThreeGap/2, p.SupernessInner)
	_, err := roundTerminalTry("three", figBall(p), turnCubic(lo.cubic(2, true)), turnCubic(li.cubic(1, false)),
		-(x0 + f3.ThreeLowerBall), figStroke{Rise: true, Before: turnCubic(li.cubic(0, false))})
	if err == nil {
		t.Error("the lower terminal of 3, with the stalk 12.5 thinner, was drawn")
	}
}
