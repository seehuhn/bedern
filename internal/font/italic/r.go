// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// r draws the r: the entry and the stem down to a plain foot, and the arm,
// the arch of n between stems RCounter apart cut at its peak, which winds
// from there into the ball, whose top lies RBallTop below the x-height.  The
// rule of the ball places the ball ahead of the peak where the spiral of the
// outer edge puts it (ballAtDepth).
func r(p Params) letter {
	e1 := p.SideStemLeft
	q := p
	if p.RStartDir > 0 {
		q.ArchStartDir = p.RStartDir
	}
	arm := q.archNodes(e1, e1+p.Stem+p.RCounter)[:2]
	if p.RStalkW > 0 {
		arm[1].W = p.RStalkW
		arm[1].P.Y = p.XHeight + p.TopLift + p.ArchOver - p.RStalkW/2
	}
	depth := p.ArchOver + p.RBallTop + p.Ball/2
	return letter{
		strokes: []stroke.Stroke{
			p.stem(e1, p.XHeight, false),
			{Nib: p.nib(), Nodes: arm, Start: hidden, End: p.ballAtDepth(depth)},
		},
		advance: p.RWidth,
	}
}
