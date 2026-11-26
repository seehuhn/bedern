// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// loopNodes returns the nodes of a narrow-eyed loop whose centre lines cross
// at cross, in the upright description before the shear.  The loop points in
// the direction axis (in degrees, 0 to the right) and reaches length from the
// crossing to the tip of its turn.  The strands are wh+eye apart, the turn
// starts turnIn short of the full half distance and its strands leave in the
// direction turnDir against the axis.  The nodes are
//
//   - the lower strand, moved by in along it from the crossing, in the
//     direction crossIn below the axis,
//   - the three nodes of the turn, with Hairpin set,
//   - the upper strand, moved by out back along it from the crossing, in the
//     direction crossOut above the axis.
//
// The widths of the nodes are wh*(w0+(i-2)*taper) for node i.  Positions and
// directions are given before the shear and returned after it.
func (p Params) loopNodes(cross Point, axis, length, eye, turnIn, wh, w0, taper, turnDir, crossIn, crossOut, in, out float64) []stroke.Node {
	k := p.slant()
	ca, sa := cosd(axis), sind(axis)
	node := func(u, v, dir, w float64, hairpin bool) stroke.Node {
		dir += axis
		x := cross.X + u*ca - v*sa
		y := cross.Y + u*sa + v*ca
		d := math.Atan2(sind(dir), cosd(dir)-k*sind(dir)) * 180 / math.Pi
		return stroke.Node{P: pt(x-k*y, y), Dir: d, K: auto, W: w, DW: auto, Hairpin: hairpin}
	}
	r := (wh+eye)/2 - turnIn
	w := func(i int) float64 { return wh * (w0 + float64(i-2)*taper) }
	return []stroke.Node{
		node(in, -in*math.Tan(crossIn*math.Pi/180), -crossIn, w(0), false),
		node(length-r, -r, turnDir, w(1), true),
		node(length, 0, 90, w(2), true),
		node(length-r, r, 180-turnDir, w(3), true),
		node(-out, -out*math.Tan(crossOut*math.Pi/180), 180+crossOut, w(4), false),
	}
}
