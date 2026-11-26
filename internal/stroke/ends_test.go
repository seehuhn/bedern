// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"math"
	"testing"
)

// TestRoundEndOnCurvedStroke ends a stroke which turns to the left with a
// round end whose right edge meets the half circle 17 degrees before its
// extreme in x.  The short arc to the extreme is drawn into the right edge:
// the half circle starts at the extreme, and the edge runs into it with its
// direction and its curvature.
func TestRoundEndOnCurvedStroke(t *testing.T) {
	const w = 24.0
	s := Stroke{
		Nib: &Nib{Major: w, Minor: w},
		Nodes: []Node{
			{P: Point{X: 0, Y: 0}, Dir: 35, K: Auto, W: w},
			{P: Point{X: 40, Y: 50}, Dir: 73, K: Auto, W: w},
		},
		End: End{Kind: Round},
	}
	r, err := s.Build()
	if err != nil {
		t.Fatal(err)
	}
	c := r.Contours[0]
	i := -1
	for j, pc := range c {
		if pc.Kind == EndPiece {
			i = j
			break
		}
	}
	if i < 1 {
		t.Fatal("no end piece after an edge")
	}
	arc, edge := c[i].C, c[i-1].C
	if l := arc.Length(); l < MinPiece {
		t.Errorf("the first piece of the half circle is %.1f units long", l)
	}
	if tx := arc.Tangent(0).X; math.Abs(tx) > 1e-6 {
		t.Errorf("the half circle starts off the extreme in x: the tangent there is %.4f,%.4f", tx, arc.Tangent(0).Y)
	}
	if d := math.Abs(angleBetween(edge.Tangent(1), arc.Tangent(0))); d > 0.01 {
		t.Errorf("the edge meets the half circle at an angle of %.3f degrees", d)
	}
	if ka, kb := edge.Curvature(1), arc.Curvature(0); math.Abs(ka-kb) > 0.05*kb {
		t.Errorf("the curvature jumps from %.4f to %.4f where the edge meets the half circle", ka, kb)
	}
}
