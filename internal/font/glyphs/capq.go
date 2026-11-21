// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capQ draws the capital Q: the O (capRound, the outline of capO unchanged) with
// a sweeping tail which leaves the bottom of the bowl on the right of the middle,
// crosses the baseline and ends QTailDepth below it in a straight cut QTailTip
// long, at right angles to the stroke.  The tail is part of the outer contour:
// the outer edge of the round runs from the top round the left side to the point
// QTailStart from the left extreme, where the lower edge of the tail leaves it;
// the lower edge falls steeply and turns in one convex curve to the end, which it
// reaches rising at capQTail.EndAngle; from the upper corner of the cut the upper
// edge runs back in one concave curve and meets the round again QTailJoin from
// the left extreme, from where the outer edge goes on round the right side to
// the top.  The counter is that of O.
//
// Side kinds: the round sides of O; the tail ends inside the right extreme.
func capQ(p design.Params) (outline.Path, float64) {
	f := capRound(p)
	g := p.Caps.CGSJQ
	var path outline.Path
	start := f.outer.cubic(0, true).P3 // the top extreme
	path.MoveTo(start.X, start.Y)
	capRoundLeft(f).appendTo(&path)
	capQTail(p, f, g).appendTo(&path)
	path.CurveTo(f.outer.cubic(0, true).args())
	superEllipse(&path, f.left+p.Caps.Round, f.innerBottom, f.right-p.Caps.Round, f.innerTop, p.SupernessInner, false)
	return path, design.CapQAdvance(p)
}

// capQTailShape holds the shape constants of the tail of Q: the directions of
// the edges at their ends (degrees, counter-clockwise from east) and the handles,
// as fractions of the chord of each edge.
type capQTailShape struct {
	StartAngle float64 // the lower edge, where it leaves the round
	EndAngle   float64 // the direction of the tail at its end
	JoinAngle  float64 // the upper edge, where it meets the round
	LowerStart float64 // handle of the lower edge at the round
	LowerEnd   float64 // handle of the lower edge at the end
	UpperEnd   float64 // handle of the upper edge at the end
	UpperJoin  float64 // handle of the upper edge at the round
}

var capQTail0 = capQTailShape{
	StartAngle: -60,
	EndAngle:   -5,
	JoinAngle:  105,
	LowerStart: 0.45,
	LowerEnd:   0.45,
	UpperEnd:   0.45,
	UpperJoin:  0.30,
}

// capQTail returns the outer edge of the round from the bottom extreme to the
// right extreme with the tail in it.
func capQTail(p design.Params, f capRoundFrame, g design.CGSJQ) chain {
	sh := capQTail0
	q := f.outer.cubic(3, true) // bottom to the right extreme
	tA := q.tAtX(f.left + g.QTailStart)
	tB := q.tAtX(f.left + g.QTailJoin)
	a, b := q.at(tA), q.at(tB)
	deg := math.Pi / 180
	dEnd := dirOf(sh.EndAngle * deg)
	n := outline.Point{X: -dEnd.Y, Y: dEnd.X}
	e := outline.Point{X: f.left + g.QTailEnd, Y: -g.QTailDepth}
	el := outline.Point{X: e.X - g.QTailTip/2*n.X, Y: e.Y - g.QTailTip/2*n.Y}
	eu := outline.Point{X: e.X + g.QTailTip/2*n.X, Y: e.Y + g.QTailTip/2*n.Y}

	edge := func(p0 outline.Point, a0 float64, k0 float64, p3 outline.Point, a3 float64, k3 float64) cubic {
		ch := math.Hypot(p3.X-p0.X, p3.Y-p0.Y)
		u0, u3 := dirOf(a0*deg), dirOf(a3*deg)
		return cubic{p0,
			outline.Point{X: p0.X + k0*ch*u0.X, Y: p0.Y + k0*ch*u0.Y},
			outline.Point{X: p3.X - k3*ch*u3.X, Y: p3.Y - k3*ch*u3.Y},
			p3}
	}
	lower := edge(a, sh.StartAngle, sh.LowerStart, el, sh.EndAngle, sh.LowerEnd)
	upper := edge(eu, sh.EndAngle+180, sh.UpperEnd, b, sh.JoinAngle, sh.UpperJoin)
	return join(
		chain{curveSeg(q.sub(0, tA))},
		splitAtExtrema(lower),
		lineChain(el, eu),
		splitAtExtrema(upper),
		chain{curveSeg(q.sub(tB, 1))},
	)
}
