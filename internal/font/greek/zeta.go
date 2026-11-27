// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['ζ'] = zeta }

// zetaHead returns the two strokes of the head of ζ and ξ: the long diagonal, which begins at the top right with a short bar
// running to the left and falls to the left extreme (lx, ly) of the belly, and
// the small ear at the top left, which curls round and runs into the
// diagonal near the top.  The diagonal ends with the node at the left extreme.
func zetaHead(c *ctx, lx, ly float64) (diag []stroke.Node, ear stroke.Stroke) {
	diag = []stroke.Node{
		n(350, 664, -142, auto),
		n(200, 540, -135, auto),
		n(100, 400, -111, auto),
		n(lx, ly, -90, auto),
	}
	ear = zetaEar()
	return diag, ear
}

// ζ: the head and the outstroke.
func zeta(c *ctx) letter {
	pen := c.pen
	lx := 20 + penWidth(pen, 90)/2
	ly := 190.0
	nodes, ear := zetaHead(c, lx, ly)
	nodes = append(nodes, outstroke(c, lx, ly)...)
	return letter{adv: outRight + penWidth(pen, -90)/2 + 20, strokes: []stroke.Stroke{
		{Nodes: nodes, Start: cut, End: cut},
		ear,
	}}
}

// zetaEar returns the ear at the top left of ζ and ξ.
func zetaEar() stroke.Stroke {
	ex, ey, ea, eb := 140.0, 650.0, 62.0, 62.0
	return stroke.Stroke{
		Nodes: []stroke.Node{
			ellipseAt(ex, ey, ea, eb, 125, true),
			ellipseNode(ex, ey, ea, eb, 0, true),
			n(238, 598, -20, auto),
		},
		Start: cut, End: hidden,
	}
}
