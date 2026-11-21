// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"
)

// The foot serif of a diagonal, tested with made-up values: it is the head serif
// turned upside down, with a tip on the baseline, and it leaves the edge along
// the same bracket.
func TestDiagonalFootServif(t *testing.T) {
	t.Parallel()
	d := diagonal{X: 200, Y: 150, Lean: 0.5, Thick: 60}
	s := diagSerif{Over: 50, Thick: 33, Bracket: tdBracket}
	for _, e := range []edge{edgeLeft, edgeRight} {
		c := d.footServif(e, 0, s)
		if len(c) == 0 {
			t.Fatalf("edge %v: no foot serif", e)
		}
		if err := c.connected(); err != nil {
			t.Fatal(err)
		}
		// The chain starts on the edge above the tip.
		start := c.start()
		if start.Y <= s.Thick {
			t.Errorf("edge %v: the chain starts at height %.2f, want above the tip", e, start.Y)
		}
		// The tip stands on the baseline.
		var tipBottom, tipTop *piece
		for i := range c {
			p := c[i]
			if p.line && p.P0.X == p.P3.X {
				tipTop = &c[i]
			}
			if p.line && p.P0.Y == 0 && p.P3.Y == 0 {
				tipBottom = &c[i]
			}
		}
		if tipTop == nil || tipBottom == nil {
			t.Fatalf("edge %v: no tip or no base in %v", e, c)
		}
		// Tangent to the edge where the bracket arrives.
		tg := c.startTangent()
		want := d.down()
		if math.Abs(tg.X*want.Y-tg.Y*want.X) > 1e-6 {
			// the chain starts going down the edge or away from it, never across
			t.Errorf("edge %v: the bracket does not leave the edge along it: %v against %v", e, tg, want)
		}
	}
}
