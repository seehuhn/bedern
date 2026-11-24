// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// ð, ə and µ: the letters of the roman's special list which the italic takes
// from the letters it has.  ð is the bowl of o whose right side runs on up
// into an arm which bends over to the left, as ∂ is drawn, with a bar across
// the arm; ə is e turned through half a turn; µ is the μ of the Greek
// lowercase.

// The shape of ð.
const (
	// ethTopX is the end of the arm, in widths of the bowl from its left
	// extreme, and ethTopDir the direction in which the stroke leaves it,
	// upright, in degrees: the arm ends level, a little above the ascender.
	ethTopX, ethTopDir = 0.30, -8.0
	// ethMidX is the middle of the arm, where the bar crosses it, in widths
	// of the bowl from its right extreme, ethMidY its height in the space
	// between the top of the bowl and the ascender, and ethMidDir its
	// direction, upright, in degrees.
	ethMidX, ethMidY, ethMidDir = 0.36, 0.465, 287.0
	// ethBarLeft and ethBarRight are the reach of the bar left and right of
	// the centre line of the arm, in widths of the bowl, ethBarAngle its angle
	// above the horizontal, upright, in degrees, and ethBarOff the height of
	// its crossing of the arm over the middle of the arm, which keeps the
	// crossing of the edges away from the node there.
	ethBarLeft, ethBarRight, ethBarAngle = 0.28, 0.14, 25.0
	ethBarOff                            = -12.0
	// ethEndIn, ethEndDepth and ethEndDir place the end of the stroke, which
	// closes the bowl inside the arm: its distance left of the centre line of
	// the arm at the right extreme and below the top of the bowl, and its
	// direction, upright, in degrees.
	ethEndIn, ethEndDepth, ethEndDir = 0.0, 50.0, -30.0
	// ethTopOver is the height of the top of the end of the arm over the top
	// of the letters of l and d, 3.
	ethTopOver = 3.0
	// ethTop moves the node at the top of the bowl to the right.
	ethTop = 0.0
)

// eth draws ð: one stroke of the pen from the end of the arm, cut across, over
// the top of the arm to the right, down the arm into the right side of the
// bowl, round the bottom, up the left side and over the top, where it ends
// inside the arm; and the bar, a hairline rising to the right across the
// arm.
func eth(p Params) letter {
	asc := p.XHeight + p.AscenderEntryRise
	l := p.SideRoundLeft
	w := p.OBowlWidth
	r := l + w
	cx := (l + r) / 2
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	cy := (top + bottom) / 2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: auto}
	}
	crest := asc + ethTopOver
	midY := top + ethMidY*(crest-top)
	midX := r - ethMidX*w
	// armX is the centre line of the arm at the height y, a straight line
	// from the middle of the arm to the right extreme of the bowl.
	sideX := r - wv/2 - oSide[1]
	armX := func(y float64) float64 { return sideX + (midX-sideX)*(y-cy)/(midY-cy) }
	// The bowl is that of o, walked the other way: its nodes have the
	// curvatures and the rates of thickening which o has at its nodes, with
	// the sign of the other direction.
	ro, err := o(p).strokes[0].Resolved()
	if err != nil {
		panic(err)
	}
	same := func(n stroke.Node, k stroke.Node) stroke.Node {
		n.K, n.DW = -k.K, -k.DW
		return n
	}
	arm := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			node(l+ethTopX*w, crest-wh/2, ethTopDir),
			node(midX, midY, ethMidDir),
			same(node(r-wv/2-oSide[1], cy, 270), ro[0]),
			same(node(cx-oTurn, bottom+wh/2, 180), ro[3]),
			same(node(l+wv/2+oSide[0], cy, 90), ro[2]),
			same(node(cx+oTurn+ethTop, top-wh/2, 0), ro[1]),
			node(armX(top-ethEndDepth)-ethEndIn, top-ethEndDepth, ethEndDir),
		},
		Start: round,
		End:   hidden,
	}
	tan := math.Tan(ethBarAngle * math.Pi / 180)
	l0, r0 := ethBarLeft*w, ethBarRight*w
	by := midY + ethBarOff
	bx := armX(by)
	bar := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(bx-l0, by-l0*tan), Dir: ethBarAngle, W: p.Hair},
			{P: pt(bx+r0, by+r0*tan), Dir: ethBarAngle, W: p.Hair},
		},
	}
	return letter{strokes: []stroke.Stroke{arm, bar}, advance: r + p.SideRoundRight}
}

// turned returns the letter turned through half a turn about (cx, cy), in the
// upright description.  The pen is the same ellipse turned, the strokes keep
// their thickness and their curvature, and the shear of the italic commutes
// with the turn up to a move, so that the turned letter is drawn with the
// pen as the letter is.
func (l letter) turned(cx, cy float64) letter {
	out := letter{advance: l.advance}
	for _, s := range l.strokes {
		t := s
		t.Nodes = make([]stroke.Node, len(s.Nodes))
		for i, nd := range s.Nodes {
			nd.P = pt(2*cx-nd.P.X, 2*cy-nd.P.Y)
			nd.Dir = math.Mod(nd.Dir+180, 360)
			t.Nodes[i] = nd
		}
		turn := func(e stroke.End) stroke.End {
			if e.Kind == stroke.Cut && e.Line {
				e.Angle = math.Mod(e.Angle+180, 360)
			}
			return e
		}
		t.Start, t.End = turn(s.Start), turn(s.End)
		out.strokes = append(out.strokes, t)
	}
	for _, d := range l.dots {
		out.dots = append(out.dots, dot{C: pt(2*cx-d.C.X, 2*cy-d.C.Y), D: d.D})
	}
	return out
}

// schwa draws ə: the e turned through half a turn about the middle of its
// advance and of its height, so that the bar and the eye stand below the
// middle and the exit ends at the upper left; its sides are those of e
// exchanged.
func schwa(p Params) letter {
	l := e(p)
	return l.turned(l.advance/2, (p.XHeight+p.Overshoot-p.Overshoot)/2)
}

// micro draws µ: the μ of the Greek lowercase.
func micro(p Params) letter { return mu(p) }
