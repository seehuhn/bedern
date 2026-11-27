// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['υ'] = upsilon }

// υ: one stroke with its entry: a short rise from the left into a small
// hook, down the left stem, round the bowl and up the right to a cut end
// leaning to the left.
func upsilon(c *ctx) letter {
	nodes, rx, _ := cupNodes(c, 292, 112, 120)
	return letter{adv: rx + penWidth(c.pen, 90)/2 + 30, strokes: []stroke.Stroke{
		{Nodes: nodes, Start: cut, End: cut},
	}}
}

// cupNodes returns the nodes of the cup of υ and ψ, whose centre lines are
// width apart at the two sides, whose right arm ends reach to the left of its side, and the right side and the middle of the cup.
func cupNodes(c *ctx, width, reach, dir float64) (nodes []stroke.Node, rx, bx float64) {
	pen := c.pen
	x, o := c.x, c.o
	left := 25.0
	yt := x + o - penWidth(pen, 0)/2
	yb := -o + penWidth(pen, 180)/2
	// the hook
	ha, hb := 62.0, 60.0
	lx := left + 100.0 // the centre line of the left stem
	hx := lx - ha
	hy := yt - hb
	// the bowl
	rx = lx + width
	a := (rx - lx) / 2
	bx = lx + a
	b := 215.0
	by := yb + b
	nodes = []stroke.Node{
		ellipseAt(hx, hy, ha, hb, 195, false),
		ellipseNode(hx, hy, ha, hb, 0, false),
		ellipseNode(hx, hy, ha, hb, -90, false),
		ellipseNode(bx, by, a, b, -90, true),
		ellipseNode(bx, by, a, b, 0, true),
		ellipseNode(bx, by, a, b, 90, true),
		n(rx-reach, x+o-14, dir, auto),
	}
	nodes[0].DW = 0
	nodes[len(nodes)-1].DW = 0
	return nodes, rx, bx
}
