// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// c draws the c: one stroke of the pen from the ball at the top right over
// the top, down the left side and round the bottom of o into the end of the
// exit.  The ball, placed by the rule of the spiral, has its lowest ink
// CBallLow below the top of the letter and its rightmost ink CBallIn inside
// the end of the lower stroke.
func c(p Params) letter {
	l := p.SideRoundLeft
	cx := l + p.CBowlWidth/2
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	cy := (top + bottom) / 2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	tipX := l + p.CTipInk
	r := p.Ball / 2
	return letter{
		strokes: []stroke.Stroke{{
			Nib: p.nib(),
			Nodes: []stroke.Node{
				{P: pt(cx+cCrest, top-p.BallStalk/2), Dir: 180, K: auto, W: p.BallStalk, DW: 0},
				node(l+wv/2+oSide[0], cy, 270),
				node(cx-oTurn, bottom+wh/2, 0),
				p.exitEnd(tipX),
			},
			Start: p.ballAtDepth(p.CBallLow - r),
			End:   round,
		}},
		advance: p.CWidth,
	}
}

// exitEnd returns the end node of an exit whose rightmost ink lies at x.
func (p Params) exitEnd(x float64) stroke.Node {
	return stroke.Node{P: pt(x-p.ExitEndW/2, p.ExitTipHeight-p.ExitEndW/2), Dir: p.ExitAngle, K: p.ExitEndK, W: p.ExitEndW, DW: 0}
}

// cCrest places the crest of c, the end node before the ball, right of the
// middle of the round as the top of o lies, so that the rule of the ball
// puts the ball over the end of the lower stroke.
const cCrest = 12.0
