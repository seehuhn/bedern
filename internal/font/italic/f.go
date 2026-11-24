// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// f draws the descending f: one stroke from the ball of the hook at the top
// down the stem through the x-height and the baseline into the tail with
// its ball (tail), and the crossbar, a hairline whose top lies on the
// x-height, reaching FBarLeft and FBarRight beyond the edges of the stem.
// The hook is the tail turned through half a turn about the centre line of
// the stem at the height FHookAxis.
func f(p Params) letter {
	e := p.FSideLeft
	c := e + p.Stem/2
	tail, end := p.tail(e)
	// The hook: the nodes of the tail turned about (c, FHookAxis), walked
	// from the ball towards the stem.
	var hook []stroke.Node
	for k := len(tail) - 1; k >= 0; k-- {
		nd := tail[k]
		nd.P = pt(2*c-nd.P.X, 2*p.FHookAxis-nd.P.Y)
		// Turned through half a turn and walked the other way: the
		// direction is the same, the curvature too.
		hook = append(hook, nd)
	}
	s := stroke.Stroke{Nodes: append(hook, tail...), Start: end, End: end}
	y := p.XHeight - p.Hair/2
	bar := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(e-p.FBarLeft, y), Dir: 0, W: p.Hair},
			{P: pt(e+p.Stem+p.FBarRight, y), Dir: 0, W: p.Hair},
		},
	}
	return letter{strokes: []stroke.Stroke{s, bar}, advance: p.FWidth}
}
