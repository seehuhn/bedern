// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The comma and the quotation marks made of it.
//
// The comma is a round head, a circle of CommaHead across whose top is level
// with the top of the full stop, and a tail which curves to the left and ends in
// a short cut.  It is one contour, drawn in the manner of the ball terminal:
// the outer (right) edge of the tail runs into the head without a corner at the
// right extreme of the head, as one cubic curve which meets the circle with its
// direction and its curvature, and the curvature falls evenly from that of the
// head to that of the tail; the inner (left) edge meets the head at a corner, the tip of a wedge of white
// between the tail and the head.  The right single quotation mark is the comma
// raised to the quotation height, the left one is that turned through half a
// turn, and the double marks are two of them side by side.

// arcAxes returns the arc of the circle about c of the radius r from the
// direction a0 to the direction a1 (counter-clockwise, a1 > a0), cut at the
// extremes of the circle.
func arcAxes(c outline.Point, r, a0, a1 float64) chain {
	var ch chain
	cur := a0
	for k := math.Floor(a0/(math.Pi/2)) + 1; ; k++ {
		next := k * math.Pi / 2
		if next >= a1-1e-9 {
			break
		}
		ch = append(ch, arcPieces(c, r, cur, next)...)
		cur = next
	}
	return append(ch, arcPieces(c, r, cur, a1)...)
}

// commaChain returns the comma as a closed counter-clockwise chain, starting
// where the outer edge runs into the head.  The centre of the head is at (0, cy),
// and the lowest ink of the tip of the tail is at the height bottom.
func commaChain(p design.Params, cy, bottom float64) chain {
	q := p.Punct
	r := q.CommaHead / 2
	centre := outline.Point{Y: cy}
	aOut, aIn := 0.0, -degrees(q.CommaInnerAt)
	on := func(a float64) outline.Point {
		return outline.Point{X: centre.X + r*math.Cos(a), Y: centre.Y + r*math.Sin(a)}
	}
	A, W := on(aOut), on(aIn)

	// The tip: a cut of the width CommaTip across the direction h of the tail.
	phi := degrees(q.CommaEndAngle)
	h := outline.Point{X: -math.Sin(phi), Y: -math.Cos(phi)}
	nOut := outline.Point{X: math.Cos(phi), Y: -math.Sin(phi)}
	ty := bottom + q.CommaTip/2*math.Sin(phi)
	tc := outline.Point{X: -q.CommaRun, Y: ty}
	To := outline.Point{X: tc.X + q.CommaTip/2*nOut.X, Y: tc.Y + q.CommaTip/2*nOut.Y}
	Ti := outline.Point{X: tc.X - q.CommaTip/2*nOut.X, Y: tc.Y - q.CommaTip/2*nOut.Y}

	// The outer edge, from the tip to the head.
	// The handle at the tip is CommaOuterPull times the chord at most; where the
	// curve of the head does not allow it (a tail which turns less), it is shortened
	// until a curve exists.
	var outer cubic
	var err error
	chord := math.Hypot(A.X-To.X, A.Y-To.Y)
	for f := q.CommaOuterPull; f > 0.05; f *= 0.95 {
		outer, err = g2End(To, neg(h), f*chord, A, outline.Point{X: -math.Sin(aOut), Y: math.Cos(aOut)}, 1/r)
		if err == nil {
			break
		}
	}
	if err != nil {
		panic(fmt.Sprintf("glyphs: the comma: %v", err))
	}
	// The inner edge, from the corner at the head to the tip.
	psi := degrees(q.CommaInnerStart)
	dW := outline.Point{X: -math.Sin(psi), Y: -math.Cos(psi)}
	pull := q.CommaInnerPull * math.Hypot(Ti.X-W.X, Ti.Y-W.Y)
	inner := cubic{W,
		outline.Point{X: W.X + pull*dW.X, Y: W.Y + pull*dW.Y},
		outline.Point{X: Ti.X - pull*h.X, Y: Ti.Y - pull*h.Y},
		Ti}
	return join(
		arcAxes(centre, r, aOut, aIn+2*math.Pi),
		splitAtExtrema(inner),
		chain{lineSeg(Ti, To)},
		splitAtExtrema(outer),
	)
}

// commaTop returns the height of the top of the head of the comma, which is the
// top of the full stop.
func commaTop(p design.Params) float64 { return p.PeriodCentre() + p.DotDiameter/2 }

// commaAt returns the comma whose head has its top at the height top, as a
// chain, with the centre of the head at the abscissa cx.  The tail reaches as far
// below the head as that of the comma proper.
func commaAt(p design.Params, cx, top float64) chain {
	cy := top - p.Punct.CommaHead/2
	bottom := -p.Punct.CommaDepth - (commaTop(p) - top)
	return commaChain(p, cy, bottom).shifted(cx, 0)
}

// comma draws the comma.
func comma(p design.Params) (outline.Path, float64) {
	var path outline.Path
	commaAt(p, 0, commaTop(p)).addContour(&path)
	return placeOnAxis(path, p.DotsAdvance())
}

// quoteChain returns the right single quotation mark, as a chain, with the
// centre of the head at cx.
func quoteChain(p design.Params, cx float64) chain {
	return commaAt(p, cx, p.Punct.QuoteTop)
}

// quoteLeftChain returns the left single quotation mark: the right one turned
// through half a turn about the centre of its box, with the centre of the head
// of the right one at cx.
func quoteLeftChain(p design.Params, cx float64) chain {
	var path outline.Path
	quoteChain(p, 0).addContour(&path)
	x0, y0, x1, y1 := path.BBox()
	return quoteChain(p, cx).halfTurn(cx+(x0+x1)/2, (y0+y1)/2)
}

// quoteright draws the right single quotation mark, which is also the
// apostrophe.
func quoteright(p design.Params) (outline.Path, float64) {
	var path outline.Path
	quoteChain(p, 0).addContour(&path)
	return placeInk(path, p.Punct.Quote)
}

// quoteleft draws the left single quotation mark.
func quoteleft(p design.Params) (outline.Path, float64) {
	var path outline.Path
	quoteLeftChain(p, 0).addContour(&path)
	return placeInk(path, p.Punct.Quote)
}

// quotedblright draws the right double quotation mark: two single marks side by
// side.
func quotedblright(p design.Params) (outline.Path, float64) {
	var path outline.Path
	quoteChain(p, 0).addContour(&path)
	quoteChain(p, p.Punct.QuotePitch).addContour(&path)
	return placeInk(path, p.Punct.QuoteDouble)
}

// quotedblleft draws the left double quotation mark.
func quotedblleft(p design.Params) (outline.Path, float64) {
	var path outline.Path
	quoteLeftChain(p, 0).addContour(&path)
	quoteLeftChain(p, p.Punct.QuotePitch).addContour(&path)
	return placeInk(path, p.Punct.QuoteDouble)
}
