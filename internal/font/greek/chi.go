// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

func init() { draw['χ'] = chi }

// χ: a stroke from the entry hook at the top left down to the right, which
// turns below the baseline into a short exit, and a straight stroke from the
// top right to the bottom left which crosses it.
func chi(c *ctx) letter {
	pen := c.pen
	top := c.x + c.o - penWidth(pen, 0)/2
	p1 := pt(268, 170)
	nodes := dgStemTo(c, top, p1.X, p1.Y, math.NaN(), 8)
	dl := nodes[len(nodes)-1].Dir
	cy0 := 130.0 // the crossing
	cx0 := dgLineX(p1.X, p1.Y, dl, cy0)
	// the exit turn, anticlockwise
	a2, b2 := 90.0, 110.0
	yb := c.desc + penWidth(pen, 0)/2 + 4
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
	// the straight stroke, with flat ends
	const d2 = -125.0
	t0 := pt(dgLineX(cx0, cy0, d2, top-12), top-12)
	t1 := pt(dgLineX(cx0, cy0, d2, c.desc+30), c.desc+30)
	diag := []stroke.Node{e(t0.X, t0.Y, d2, auto), e(t1.X, t1.Y, d2, auto)}
	sk := -penWidth(pen, d2) / 2 / math.Tan(d2*math.Pi/180)
	diag[0].Skew = sk
	diag[1].Skew = sk
	right := math.Max(ex.X+penWidth(pen, 45)/2, t0.X+penWidth(pen, d2)/2)
	return letter{adv: right + 25, strokes: []stroke.Stroke{
		{Nodes: nodes, Start: cut, End: cut},
		{Nodes: diag, Start: cut, End: cut},
	}}
}
