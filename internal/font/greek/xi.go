// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['ξ'] = xi }

// ξ: three tiers.  The head is that of ζ: the ear,
// and a stroke which runs from the top right to the left, down to the upper
// belly and back right along the waist, where it ends.  The lower stroke
// begins inside the waist, falls to the lower belly at the left and runs on
// into the outstroke of ζ.
func xi(c *ctx) letter {
	pen := c.pen
	lx := 20 + penWidth(pen, 90)/2
	ly := 165.0
	tip := n(268, 343, 8, auto)
	tip.W = 56
	ua, ub := 112.0, 108.0
	ucx, ucy := lx+2+ua, 452.0
	up := []stroke.Node{
		n(366, 664, -168, auto),
		n(240, 598, -150, auto),
		n(128, 525, -127, auto),
		ellipseNode(ucx, ucy, ua, ub, -90, true),
		ellipseNode(ucx, ucy, ua, ub, -35, true),
		n(190, 346, -8, auto),
		tip,
	}
	la, lb := 150.0, 336-ly
	lcx := lx + la
	low := []stroke.Node{
		ellipseAt(lcx, ly, la, lb, 118, true),
		ellipseNode(lcx, ly, la, lb, -90, true),
	}
	low = append(low, outstroke(c, lx, ly)...)
	return letter{adv: outRight + penWidth(pen, -90)/2 + 20, strokes: []stroke.Stroke{
		{Nodes: up, Start: cut, End: cut},
		zetaEar(),
		{Nodes: low, Start: hidden, End: cut},
	}}
}
