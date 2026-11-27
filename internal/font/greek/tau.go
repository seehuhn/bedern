// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['τ'] = tau }

// τ: a bar, bent down at its left end, and a stem with the exit of ι.
func tau(c *ctx) letter {
	yt := c.x + c.o - penWidth(c.pen, 0)/2
	sx := 245.0
	bar := stroke.Stroke{Nodes: barA(c, 25, 450, yt), Start: cut, End: cut}
	nodes := append([]stroke.Node{e(sx, yt-20, -90, auto)}, footA(c, sx, 120, 72)...)
	return letter{adv: 460 + 30, strokes: []stroke.Stroke{
		bar, {Nodes: nodes, Start: hidden, End: cut},
	}}
}
