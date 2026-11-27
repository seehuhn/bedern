// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['σ'] = sigma }

// σ: one stroke, from the tip of the outstroke at the top right, left
// along the top, round the bowl anticlockwise and up its right side into
// the outstroke.
func sigma(c *ctx) letter {
	cx, cy, a, b := omicronBowl(c)
	a *= 0.95
	cx -= 0.05 * a
	topEdge := c.x + c.o // the top of the ink
	tipX := cx + a + 100.0
	tipW := 60.0
	nodes := []stroke.Node{
		e(tipX, topEdge-tipW/2, 180, tipW),
	}
	for _, phi := range []float64{90, 180, 270, 360, 436} {
		nodes = append(nodes, tiltAt(cx, cy, a, b, omicronTilt, phi, true))
	}
	nodes[len(nodes)-1].W = 40
	nodes[len(nodes)-1].DW = 0
	return letter{adv: tipX + 20, strokes: []stroke.Stroke{{Nodes: nodes, Start: cut, End: hidden}}}
}
