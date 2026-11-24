// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"seehuhn.de/go/bedern/internal/stroke"
)

// v draws the v: one stroke from the entry down the stem, which turns round
// at the bottom as the turn of u does, rises round the free side, light
// against the stem, and turns over at the top to the left into the ball,
// which hangs inside the opening of the letter.
func v(p Params) letter {
	e := p.SideStemLeft
	s := stroke.Stroke{Nib: p.nib(), Nodes: p.entry(e, p.XHeight), Start: round}
	nodes, end := p.vTurn(e+p.Stem/2, p.VWidth)
	s.Nodes = append(s.Nodes, nodes...)
	s.End = end
	return letter{strokes: []stroke.Stroke{s}, advance: p.VWidth}
}

// vTurn returns the nodes of the turn and the free side of v and w at the
// foot of the stem whose centre line stands at c, the letter's set width
// being adv, and the end of the stroke: the stem straight down to VTurn, the
// node where the turn runs vertically in the italic, the lowest node, the
// free side, its outer edge VSideX inside the set width and VSideW of the
// pen's weight there, and the crest, VCrestIn left of the free side, where
// the stroke runs to the left as the stalk of the ball, its outer edge
// ArchOver above the x-height as on the arch of r; the ball hangs below it
// on the left, its top VBallTop below the x-height, by the rule of the
// spiral.  The pen gives the thickness at the turn.
func (p Params) vTurn(c, adv float64) ([]stroke.Node, stroke.End) {
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	free := adv - p.VSideX
	side := pt(free-p.VSideW*wv/2-oSide[1], p.VSideY*p.XHeight)
	crest := pt(side.X-p.VCrestIn, p.XHeight+p.TopLift+p.ArchOver-p.BallStalk/2)
	depth := p.ArchOver + p.VBallTop + p.Ball/2
	return []stroke.Node{
		p.stemNode(c, p.VTurn),
		p.vertNode(c, p.VTurn, p.VVert, p.ShoulderW),
		{P: pt(c+p.VBottom*(free-c), -p.Overshoot+wh/2), Dir: 0, K: auto, W: auto, DW: 0},
		{P: side, Dir: 90, K: auto, W: p.VSideW * wv, DW: auto},
		{P: crest, Dir: 180, K: auto, W: p.BallStalk, DW: 0},
	}, p.ballAtDepth(-depth)
}
