// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import "testing"

// TestStemEdgeDoesNotFold checks that the edge of a straight stem whose
// thickness changes a little runs monotonically along the stem: with nearly
// zero curvatures at both ends, the handles which give those curvatures may
// overlap along the chord, and the edge then runs back on itself, which
// breaks the union of the letter (here a stem and a separate stem with a
// turned foot, the strokes of μ).
func TestStemEdgeDoesNotFold(t *testing.T) {
	nd := func(x, y, dir, w, k, dw float64) Node {
		return Node{P: Point{X: x, Y: y}, Dir: dir, W: w, K: k, DW: dw}
	}
	nib := &Nib{Major: 94, Minor: 32}
	stem := Stroke{Nib: nib, Nodes: []Node{
		nd(100, 448, -90, 72, 0, Auto), nd(100, 57, -90, 68, 0, Auto), nd(100, -154, -90, 52, 0, 0),
	}, End: End{Kind: Round}}
	foot := Stroke{Nib: nib, Nodes: []Node{
		nd(362, 448, -90, 66, 0, Auto), nd(362, 175, -90, 88, 0, Auto),
		nd(424, 4, 0, 32, Auto, Auto), nd(480, 48, 58, 24, Auto, 0),
	}}
	r, err := stem.Build()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range r.Contours {
		for _, p := range c {
			if p.Kind != EdgePiece {
				continue
			}
			// Along the stem every piece of an edge runs one way in y.
			if p.C[0].Y > 440 || p.C[0].Y < -100 {
				continue
			}
			up, down := false, false
			prev := p.C.At(0)
			for i := 1; i <= 64; i++ {
				q := p.C.At(float64(i) / 64)
				up = up || q.Y > prev.Y+1e-9
				down = down || q.Y < prev.Y-1e-9
				prev = q
			}
			if up && down {
				t.Errorf("edge piece %v runs back on itself", p.C)
			}
		}
	}
	if _, err := (Letter{Strokes: []Stroke{stem, foot}}).Build(); err != nil {
		t.Errorf("union: %v", err)
	}
}

// TestDoublesBack checks the test for handles which overlap along the chord.
func TestDoublesBack(t *testing.T) {
	down := Point{X: 0, Y: -1}
	d := Point{X: 0, Y: -100}
	if doublesBack(down, down, d, 33, 33) {
		t.Error("the handles of a straight line fold")
	}
	if !doublesBack(down, down, d, 150, 150) {
		t.Error("overlapping handles do not fold")
	}
	if doublesBack(Point{X: 1}, Point{X: -1}, Point{X: 0, Y: -50}, 60, 60) {
		t.Error("a hairpin turn folds")
	}
}
