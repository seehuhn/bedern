// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// eszett draws the sharp s: the descending stem of f with its tail, which
// at the top turns over to the right into a 3, as the long s joined to a z.
// The first stroke starts at the waist in a cut which points back to the
// left (the spur), rises round the right side of the upper bowl, runs over
// the crest at the height of the hook of f and down the stem through the
// baseline into the tail with its ball (tail).  The second stroke leaves the
// waist inside the first and runs round the right side and the bottom of
// the lower bowl, which is wider than the upper, into an end cut square and
// turned up as the lower end of s, SzEndGap clear of the stem.
func eszett(p Params) letter {
	e := p.FSideLeft
	c := e + p.Stem/2
	r := e + p.Stem
	top, bot := p.XHeight+p.AscenderEntryRise+p.Overshoot, -p.Overshoot
	wy := p.SzWaistY * p.XHeight
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	waist := pt(r+p.SzWaistX, wy)
	ux := r + p.SzUpper - wv/2
	upper := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: waist, Dir: p.SzWaistDir, K: auto, W: p.SzWaistW * p.Thin, DW: auto},
			node(ux, wy+p.SzUpperSideY*(top-wy), 90),
			node(c+p.SzCrest*(ux-c), top-wh/2, 180),
			p.stemNode(c, p.SzStemTop),
		},
		Start: round,
	}
	tail, end := p.tail(e)
	upper.Nodes, upper.End = append(upper.Nodes, tail...), end

	d := p.SzWaistDir * math.Pi / 180
	start := pt(waist.X+20*math.Cos(d), waist.Y+20*math.Sin(d))
	w := p.SEndW * p.Thin
	lx := r + p.SzLower - wv/2
	lower := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: start, Dir: p.SzLowerDir, K: auto, W: p.Thin, DW: auto},
			node(lx, bot+p.SzLowerSideY*(wy-bot), -90),
			node(r+p.SzBottomAt*(lx-r), bot+wh/2, 180),
			{P: pt(r+p.SzEndGap+w/2, bot+p.SzEndDrop), Dir: p.SzEndDir, K: auto, W: w, DW: 0},
		},
		Start: hidden,
		End:   round,
	}
	return letter{strokes: []stroke.Stroke{upper, lower}, advance: p.SzWidth}
}
