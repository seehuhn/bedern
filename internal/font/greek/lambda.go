// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

func init() { draw['λ'] = lambda }

// λ: a long stroke from the entry hook at the top left, down to the right
// and into a short exit at the baseline, and a leg from the baseline at the
// left which joins it from below.
func lambda(c *ctx) letter {
	pen := c.pen
	// the long stroke: hook, gentle arc, line, turn
	p1 := pt(358, 230)
	nodes := dgStemTo(c, c.asc+14-penWidth(pen, 0)/2, p1.X, p1.Y, math.NaN(), 8)
	dl := nodes[len(nodes)-1].Dir
	a2, b2 := 90.0, 110.0
	yb := -c.o + penWidth(pen, 0)/2
	cy2 := yb + b2
	cx2 := 0.0
	for range 50 {
		p := ellipseNode(cx2, cy2, a2, b2, dl, true).P
		cx2 += dgLineX(p1.X, p1.Y, dl, p.Y) - p.X
	}
	q := ellipseNode(cx2, cy2, a2, b2, dl, true).P
	nodes = append(nodes,
		e(q.X, q.Y, dl, auto),
		ellipseNode(cx2, cy2, a2, b2, 0, true),
		ellipseNode(cx2, cy2, a2, b2, 45, true),
	)
	nodes[len(nodes)-1].DW = 0
	ex := nodes[len(nodes)-1].P
	// the leg: its foot is cut flat
	jy := 370.0
	jx := dgLineX(p1.X, p1.Y, dl, jy)
	lx, ly := 80.0, 14.0
	dleg := math.Atan2(jy-ly, jx-lx) * 180 / math.Pi
	leg := []stroke.Node{
		e(lx, ly, dleg, auto),
		e(jx, jy, dleg, auto),
	}
	leg[0].Skew = -penWidth(pen, dleg) / 2 / math.Tan(dleg*math.Pi/180)
	return letter{adv: ex.X + penWidth(pen, 45)/2 + 25, strokes: []stroke.Stroke{
		{Nodes: nodes, Start: cut, End: cut},
		{Nodes: leg, Start: cut, End: hidden},
	}}
}
