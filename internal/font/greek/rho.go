// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['ρ'] = rho }

// ρ: one stroke, from the soft end of the descender up the stem, over
// the bowl and round it back into the stem.
func rho(c *ctx) letter {
	pen := c.pen
	x, o := c.x, c.o
	left := 25.0
	yt := x + o - penWidth(pen, 0)/2
	yb := -o + penWidth(pen, 180)/2
	lx := left + penWidth(pen, 90)/2
	a := 155.0
	b := (yt - yb) / 2
	bx, by := lx+a, (yt+yb)/2
	nodes := []stroke.Node{
		n(lx+30, c.desc+16, 99, 62),
		ellipseNode(bx, by, a, b, 90, false),
		ellipseNode(bx, by, a, b, 0, false),
		ellipseNode(bx, by, a, b, -90, false),
		ellipseNode(bx, by, a, b, 180, false),
		ellipseAt(bx, by, a, b, 200, false),
	}
	nodes[len(nodes)-1].W = 34
	return letter{adv: bx + a + penWidth(pen, 90)/2 + 30, strokes: []stroke.Stroke{
		{Nodes: nodes, Start: cut, End: hidden},
	}}
}
