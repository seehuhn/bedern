// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capC draws the capital C: the round of O (capRound), open on the right, in
// one counter-clockwise contour without a counter.  The upper end is the upper
// beak (capUpperBeak), a serif on a thin arc of the family of the serifs on the
// arms of E; the lower end is a tapered tail without a serif (capCTail), the
// lower terminal of the lowercase c at the size and with the weights of the
// capitals.  The tail sweeps out and up to the right and ends in a fine cut a
// little beyond the vertical of the upper cut.  The contour runs from the top
// extreme of the outer edge over the top, down the left side and round the bottom
// to the tail, back along the counter round the left side and over the top, and
// round the upper beak to the top extreme.
//
// Side kinds: the round side of O on the left, and on the right the tip of the
// tail, CTailReach beyond the cut of the upper beak and CSideRight from the
// edge.
func capC(p design.Params) (outline.Path, float64) {
	adv := design.CapCAdvance(p)
	p = capCParams(p)
	f := capRound(p)
	return capContour(join(capRoundLeft(f), capCTail(p, f), capRoundLeftInner(f), capUpperBeak(p, f))), adv
}

// capCParams returns p with the end of the upper beak of C in the place of that
// of G, which the shared code of the beaks reads.
func capCParams(p design.Params) design.Params {
	p.Caps.CGSJQ.BeakEnd = p.Caps.CGSJQ.CBeakEnd
	return p
}

// capRoundLeft returns the outer edge of the round from the top extreme over the
// left side to the bottom extreme.
func capRoundLeft(f capRoundFrame) chain {
	return chain{curveSeg(f.outer.cubic(1, true)), curveSeg(f.outer.cubic(2, true))}
}

// capRoundLeftInner returns the counter from the bottom extreme round the left
// side to the top extreme, walked clockwise.
func capRoundLeftInner(f capRoundFrame) chain {
	return chain{curveSeg(f.inner.cubic(1, false)), curveSeg(f.inner.cubic(2, false))}
}

// The lower end of C.
//
// The tail is drawn in the frame in which it lies at the top on the right, as
// the upper beak does (the round of O is symmetric about its middle), and
// turned upside down.  In the natural frame: the outer edge is one curve from
// the bottom extreme of the round, which it leaves running east with the
// curvature of O there, to the outer corner of the cut, CTailReach beyond the
// vertical of the upper cut and CTailHeight above the baseline; of the curves
// which arrive there between capTailArriveMin and capTailArriveMax degrees above
// the horizontal the fairest is taken (bending), as for the outer edge of a beak.
// The cut is straight, at right angles to the outer edge there, CTailCut long
// towards the inside, as on the lowercase c.  The inner edge is one curve from
// the counter, which it leaves with the direction and the curvature of the
// counter, to the inner corner of the cut, which it reaches running towards the
// outer edge: its direction there is that of the outer edge turned towards it
// by the angle which makes the tail CTailArc thick where it is thickest (the
// diameter of the largest circle between the edges, searched between 0 and
// capTailTaperMax degrees).  Of the points at which the inner edge
// may leave the counter, the one whose curve changes its curvature least
// abruptly is taken (curvatureRate).
const (
	capTailArriveMin = 30
	capTailArriveMax = 75
	capTailTaperMax  = 40
)

// capTailCutCurvature are the curvatures with which the inner edge of the tail
// of C may arrive at the cut, as fractions of the curvature of the outer edge
// there.
var capTailCutCurvature = []float64{0.6, 0.7, 0.8, 0.9, 1.0}

// A capTailParts holds the curves and the points of the tail of C, in the
// frame in which it lies at the top on the right.
type capTailParts struct {
	Counter cubic         // the upper right quarter of the counter, top to right extreme
	TI      float64       // the parameter on Counter at which Inner leaves it
	Inner   cubic         // the inner edge, from the counter to Cut
	Cut     outline.Point // the inner corner of the cut
	Corner  outline.Point // the outer corner of the cut
	Outer   cubic         // the outer edge, from Corner to the top extreme
}

// capTailFrame returns the tail of C in the frame in which it lies at the top
// on the right, see the comment above.
func capTailFrame(p design.Params, f capRoundFrame) (capTailParts, error) {
	g := p.Caps.CGSJQ
	var b capTailParts
	mid := f.top + f.bottom
	b.Corner = outline.Point{X: f.beakX(p) + g.CTailReach, Y: mid - g.CTailHeight}
	top := outline.Point{X: f.cx, Y: f.top}
	qo := f.outer.cubic(0, true)
	out, ok := capOuterArriving(b.Corner, top, qo.curvature(1), capTailArriveMin, capTailArriveMax)
	if !ok {
		return b, fmt.Errorf("the outer edge of the tail cannot reach %v", b.Corner)
	}
	b.Outer = out
	dir := out.tangent(0) // away from the cut
	b.Cut = outline.Point{X: b.Corner.X - g.CTailCut*dir.Y, Y: b.Corner.Y + g.CTailCut*dir.X}
	toward := math.Atan2(-dir.Y, -dir.X) // the direction of the stroke into the cut

	qi := f.inner.cubic(3, false)
	b.Counter = qi
	// inner returns the inner edge which leaves the counter at u and arrives
	// at the cut turned by taper (radians) towards the outer edge.
	best := math.Inf(1)
	for _, u := range capCuts(0.02, 0.5) {
		for _, share := range capTailCutCurvature {
			kCut := -share * out.curvature(0)
			// thick returns the inner edge which leaves the counter at u and
			// arrives at the cut turned by taper (radians) towards the outer
			// edge, and whether the tail is CTailArc thick.
			thick := func(taper float64) (cubic, bool) {
				c, err := g2Cubic(qi.at(u), qi.tangent(u), qi.curvature(u), b.Cut, dirOf(toward+taper), kCut)
				if err != nil || !oneWay(c, -1) {
					return c, false
				}
				return c, capTailThickest(qi.sub(0, u), c, out) >= g.CTailArc
			}
			// The least taper, in steps of two degrees and then by bisection.
			step := 2 * math.Pi / 180
			hi := 0.0
			for ; hi <= capTailTaperMax*math.Pi/180; hi += step {
				if _, ok := thick(hi); ok {
					break
				}
			}
			if hi > capTailTaperMax*math.Pi/180 {
				continue
			}
			lo := math.Max(0, hi-step)
			for i := 0; i < 12 && hi > 0; i++ {
				m := (lo + hi) / 2
				if _, ok := thick(m); ok {
					hi = m
				} else {
					lo = m
				}
			}
			cv, _ := thick(hi)
			if e := bending(cv); e < best {
				best, b.Inner, b.TI = e, cv, u
			}
		}
	}
	if math.IsInf(best, 1) {
		return b, fmt.Errorf("no inner edge of the tail is %v thick", g.CTailArc)
	}
	return b, nil
}

// capTailThickest returns the diameter of the largest circle which fits
// between the outer edge out and the inner edge, the part of the counter c0
// followed by c1, measured from 200 points of the outer edge.
func capTailThickest(c0, c1, out cubic) float64 {
	var in []outline.Point
	for _, c := range []cubic{c0, c1} {
		for i := 0; i <= 60; i++ {
			in = append(in, c.at(float64(i)/60))
		}
	}
	// The distance of each point q of the outer edge to the nearest point of the
	// inner edge is the minimum over all of them.  The search starts with the point
	// which was nearest to the previous q, and passes over the points which are as
	// far away as that in x or in y alone.  The result is the same number as that of
	// the plain loop.
	var m float64
	near := 0
	for i := 0; i <= 80; i++ {
		q := out.at(float64(i) / 80)
		d := math.Hypot(q.X-in[near].X, q.Y-in[near].Y)
		for j, r := range in {
			dx, dy := q.X-r.X, q.Y-r.Y
			if dx >= d || -dx >= d || dy >= d || -dy >= d {
				continue
			}
			if h := math.Hypot(dx, dy); h < d {
				d, near = h, j
			} else if h != h {
				d = h // NaN, as math.Min gives
			}
		}
		m = math.Max(m, d)
	}
	return m
}

// capCTail returns the lower end of C as a chain from the bottom extreme of the
// outer edge to the bottom extreme of the counter, see the comment above.
func capCTail(p design.Params, f capRoundFrame) chain {
	b, err := capTailFrame(p, f)
	if err != nil {
		panic(fmt.Sprintf("glyphs: the tail of C: %v; glyphs.Check reports such parameters", err))
	}
	return b.chain().flipY((f.top + f.bottom) / 2)
}

// chain returns the tail as a chain, from the counter at the top extreme to the
// top extreme of the outer edge, in the frame in which it lies at the top.
func (b capTailParts) chain() chain {
	return join(
		chain{curveSeg(b.Counter.sub(0, b.TI))},
		splitAtExtrema(b.Inner),
		lineChain(b.Cut, b.Corner),
		chain{curveSeg(b.Outer)},
	)
}
