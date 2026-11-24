// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// t draws the t: the stem from the oblique cut at TTop down into the exit,
// and the crossbar, a hairline whose top lies on the x-height, reaching
// CrossbarLeft and CrossbarRight beyond the edges of the stem.
func t(p Params) letter {
	e := p.TSideLeft
	c := e + p.Stem/2
	s := stroke.Stroke{
		Nodes: append([]stroke.Node{p.stemNode(c, p.TTop)}, p.exit(e)...),
		Start: stroke.End{Kind: stroke.Cut, Line: true, Angle: p.TCutAngle},
		End:   round,
	}
	y := p.XHeight - p.Hair/2
	bar := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(e-p.CrossbarLeft, y), Dir: 0, W: p.Hair},
			{P: pt(e+p.Stem+p.CrossbarRight, y), Dir: 0, W: p.Hair},
		},
	}
	return letter{strokes: []stroke.Stroke{s, bar}, advance: p.TWidth}
}
