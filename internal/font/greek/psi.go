// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['ψ'] = psi }

// ψ: the cup of υ, with its entry, and the stem through the middle of the
// cup from the ascender to the tapered end of the descender.
func psi(c *ctx) letter {
	pen := c.pen
	nodes, rx, bx := cupNodes(c, 422, 80, 114)
	// the stem leans a little to the right at the top and to the left
	// below, and is calm at both ends
	stem := []stroke.Node{
		n(bx-4, c.asc-6, -93, auto),
		n(bx-8, 110, -90, auto),
		n(bx-22, c.desc+10, -87, 56),
	}
	stem[0].DW = 0
	return letter{adv: rx + penWidth(pen, 90)/2 + 20, strokes: []stroke.Stroke{
		{Nodes: nodes, Start: cut, End: cut},
		{Nodes: stem, Start: cut, End: cut},
	}}
}
