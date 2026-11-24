// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// u draws the u: the entry and the left stem, which turns at the bottom into
// a thin rise into the right stem, and the right stem from a flat cut at the
// x-height down into the exit.  The turn is the arch of n turned round about
// the middle of the x-height between the two stems, so that it overshoots
// the baseline as the arch rises above the x-height, and its rise meets the
// right stem as far above the baseline as the arch meets the left stem below
// the x-height.
func u(p Params) letter {
	e1 := p.SideStemLeft
	e2 := e1 + p.Stem + p.NCounter
	c1, c2 := e1+p.Stem/2, e2+p.Stem/2
	turn := stroke.Stroke{
		Nib:   p.nib(),
		Nodes: p.entry(e1, p.XHeight),
		Start: round,
		End:   hidden,
	}
	q := p
	q.Branch = p.XHeight - p.URise
	q.ArchStartDir = p.URiseDir
	q.ArchStartBack = p.URiseBack
	q.ArchStartW = p.URiseW
	// The turn is the arch turned upside down about the x-height: its
	// bottom and its shoulder node keep their heights over the baseline.
	q.TopLift = 0
	q.ShoulderTop = p.ShoulderTop - p.TopLift
	q.Shoulder = p.Shoulder - p.TopLift
	arch := q.archNodes(e1, e2)
	for k := len(arch) - 1; k >= 0; k-- {
		nd := arch[k]
		nd.P = pt(c1+c2-nd.P.X, p.XHeight-nd.P.Y)
		turn.Nodes = append(turn.Nodes, nd)
	}
	right := stroke.Stroke{Nodes: []stroke.Node{p.stemNode(c2, p.XHeight)}, Start: flatTop, End: round}
	right.Nodes = append(right.Nodes, p.exit(e2)...)
	return letter{
		strokes: []stroke.Stroke{turn, right},
		advance: e2 + p.Stem + p.SideStemRight,
	}
}
