// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['ς'] = finalsigma }

// ς: a C-shaped bowl whose lower end runs on into the outstroke.
func finalsigma(c *ctx) letter {
	pen := c.pen
	yt := c.x + c.o - penWidth(pen, 0)/2
	lx := 20 + penWidth(pen, 90)/2
	cy := 215.0
	a := 200.0
	b := yt - cy
	cx := lx + a
	nodes := []stroke.Node{
		n(342, yt-8, 172, auto),
		ellipseNode(cx, cy, a, b, 180, true),
		ellipseNode(cx, cy, a, b, -90, true),
	}
	nodes = append(nodes, outstroke(c, lx, cy)...)
	return letter{adv: outRight + penWidth(pen, -90)/2 + 20, strokes: []stroke.Stroke{
		{Nodes: nodes, Start: stroke.End{Kind: stroke.Cut, Line: true, Angle: 80}, End: cut},
	}}
}
