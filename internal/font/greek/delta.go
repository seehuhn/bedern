// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['δ'] = delta }

// δ: one stroke, from the tip of the head at the upper right, left over the
// top, down the left of the head into the diagonal, round the bowl
// clockwise and up into the diagonal again.
func delta(c *ctx) letter {
	cx, cy, a, b := omicronBowl(c)
	pen := c.pen
	top := c.asc - penWidth(pen, 180)/2
	tip := e(392, 575, 152, 44)
	nodes := []stroke.Node{
		tip,
		n(185, top, 178, auto),
		n(100, 610, -102, auto),
		tiltAt(cx, cy, a, b, omicronTilt, 40, false),
		tiltAt(cx, cy, a, b, omicronTilt, -90, false),
		tiltAt(cx, cy, a, b, omicronTilt, -180, false),
		tiltAt(cx, cy, a, b, omicronTilt, -270, false),
		tiltAt(cx, cy, a, b, omicronTilt, -286, false),
	}
	nodes[len(nodes)-1].W = 40
	nodes[len(nodes)-1].DW = 0
	return letter{adv: 415, strokes: []stroke.Stroke{{Nodes: nodes, Start: cut, End: hidden}}}
}
