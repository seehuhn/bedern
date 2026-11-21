// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// Between its extreme and the stem, each edge of the bowl (the outer edge and the
// counter, above and below) is one convex curve: its tangent turns one way only,
// without an S-shape, and it has no straight stretch.  The
// constants are the limits which the test applies.
const (
	minJoinTurn   = 0.02 // radians over every stretch of joinStretch units
	joinStretch   = 25.0
	joinWrongWay  = 0.01 // radians of turning against the direction of the edge
	joinMinLength = 40.0
)

func polyHeadings(pts []outline.Point) []float64 {
	h := make([]float64, 0, len(pts)-1)
	for i := 1; i < len(pts); i++ {
		h = append(h, math.Atan2(pts[i].Y-pts[i-1].Y, pts[i].X-pts[i-1].X))
	}
	return h
}

func TestBowlJoinsAreConvexCurves(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			bowls := map[string]bowlSpec{"b": bBowl(p), "d": dBowl(p), "p": pBowl(p), "q": qBowl(p)}
			for name, b := range bowls {
				b = b.leftForm()
				oc := bowlOuterCurves(dBowlShape, b)
				cc := bowlCounter(dBowlShape, b)
				edges := map[string]cubic{
					"outer, upper":   oc[0],
					"outer, lower":   oc[3],
					"counter, lower": cc[1].cubic,
					"counter, upper": cc[4].cubic,
				}
				for what, c := range edges {
					var pts []outline.Point
					var s []float64
					length := 0.0
					for k := 0; k <= 200; k++ {
						q := c.at(float64(k) / 200)
						if k > 0 {
							length += math.Hypot(q.X-pts[k-1].X, q.Y-pts[k-1].Y)
						}
						pts = append(pts, q)
						s = append(s, length)
					}
					h := polyHeadings(pts)
					total := 0.0
					for i := 1; i < len(h); i++ {
						total += turn(h[i-1], h[i])
					}
					dir := 1.0
					if total < 0 {
						dir = -1
					}
					wrong := 0.0
					for i := 1; i < len(h); i++ {
						if d := -dir * turn(h[i-1], h[i]); d > 0 {
							wrong += d
						}
					}
					if wrong > joinWrongWay {
						t.Errorf("%s, %s %s: the edge turns by %.3f radians against its direction (an S-shape)", np.name, name, what, wrong)
					}
					if length < joinMinLength {
						continue
					}
					least, at := slidingTurns(pts, s, joinStretch)
					if least < minJoinTurn {
						t.Errorf("%s, %s %s: a stretch of %.0f units at (%.0f, %.0f) turns by only %.3f radians, want at least %v", np.name, name, what, joinStretch, at.X, at.Y, least, minJoinTurn)
					}
				}
			}
		})
	}
}
