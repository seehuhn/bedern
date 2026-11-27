// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['ι'] = iota_ }

// ι: a stem which leans a little and bends at the foot into an exit.
func iota_(c *ctx) letter {
	sx := 30 + penWidth(c.pen, -90)/2
	nodes := append([]stroke.Node{stemTopA(c, sx, 0)}, footA(c, sx, 120, 72)...)
	return letter{adv: advA(c, nodes[len(nodes)-1]), strokes: []stroke.Stroke{
		{Nodes: nodes, Start: topA, End: cut},
	}}
}
