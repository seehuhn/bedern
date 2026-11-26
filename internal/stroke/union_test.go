// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"math"
	"testing"
)

// TestUnionShortPieceAtShallowCrossing joins a stem and a band whose edge
// crosses the left edge of the stem at 20 degrees, 0.77 units before a node
// of that edge: the short piece between the crossing and the node lies within
// a quarter of a unit of the crossing edge, and the outline must still close.
func TestUnionShortPieceAtShallowCrossing(t *testing.T) {
	line := func(a, b Point) Piece {
		d := sub(b, a)
		return Piece{C: Cubic{a, add(a, mul(d, 1.0/3)), add(a, mul(d, 2.0/3)), b}}
	}
	chain := func(pts ...Point) Chain {
		var c Chain
		for i, p := range pts {
			c = append(c, line(p, pts[(i+1)%len(pts)]))
		}
		return c
	}
	stem := chain(Point{X: 0, Y: -100}, Point{X: 20, Y: -100}, Point{X: 20, Y: 100},
		Point{X: 0, Y: 100}, Point{X: 0, Y: 0})
	a := 20 * math.Pi / 180
	d, n := Point{X: math.Sin(a), Y: math.Cos(a)}, Point{X: -math.Cos(a), Y: math.Sin(a)}
	p := Point{X: 0, Y: 0.77}
	lo, hi := add(p, mul(d, -160)), add(p, mul(d, 160))
	band := chain(lo, hi, add(hi, mul(n, 30)), add(lo, mul(n, 30)))
	out, err := Union([]Chain{stem, band})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Errorf("%d contours, want 1", len(out))
	}
}
