// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

func init() { draw['κ'] = kappa }

// κ: a stem leaning a little to the right; an
// arm in one long diagonal from the top right down into the foot of the
// stem, steepening as it goes; and a leg which leaves the arm near its
// middle, runs down to the baseline and turns into the exit, as that of ι.
func kappa(c *ctx) letter {
	pen := c.pen
	k := kappaShape
	stemDir := math.Atan2(-c.x, k.stemFoot-k.stemTop) * 180 / math.Pi
	stem := stroke.Stroke{Nodes: []stroke.Node{
		e(k.stemTop, c.x, stemDir, auto), e(k.stemFoot, 0, stemDir, auto),
	}, Start: cut, End: cut}
	// The arm ends tapered inside the stem, so that no corner of it shows.
	arm := stroke.Stroke{Nodes: []stroke.Node{
		n(k.armTop.X, k.armTop.Y, -140, auto),
		n(k.join.X, k.join.Y, -128, auto),
		n(k.armFoot.X, k.armFoot.Y, -112, 30),
	}, Start: cut, End: hidden}

	// The leg runs straight from the join at legDir into a turn whose
	// bottom lies on the baseline, less the overshoot.
	const legDir = -60.0
	yb := -c.o + penWidth(pen, 0)/2
	a, ey := 85.0, 130.0
	b := ey - yb
	p := ellipseNode(0, ey, a, b, legDir, true).P
	r := legDir * math.Pi / 180
	t := (k.join.Y - p.Y) / -math.Sin(r)
	ex := k.join.X + t*math.Cos(r) - p.X
	leg := stroke.Stroke{Nodes: []stroke.Node{
		s(k.join.X, k.join.Y, legDir, auto),
		ellipseNode(ex, ey, a, b, legDir, true),
		ellipseNode(ex, ey, a, b, 55, true),
	}, Start: hidden, End: cut}
	leg.Nodes[2].DW = 0
	return letter{adv: advA(c, leg.Nodes[len(leg.Nodes)-1]), strokes: []stroke.Stroke{stem, arm, leg}}
}

// kappaShape holds the points of the centre lines of κ.
var kappaShape = struct {
	stemTop, stemFoot     float64 // the stem's centre at the x-height and the baseline
	armTop, join, armFoot Point   // the arm's ends, and where the leg leaves it
}{
	stemTop: 92, stemFoot: 70,
	armTop: pt(410, 430), join: pt(238, 250), armFoot: pt(76, 70),
}
