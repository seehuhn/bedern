// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"fmt"

	"seehuhn.de/go/bedern/internal/stroke"
)

// The symbols of mathematics drawn with the pen of the italic: ℓ, the script
// l, and ∂, the partial differential.  Build draws them as it draws the
// letters; ∂ is drawn upright (U+2202) and at the slant of the italic
// (U+1D715, the maths italic ∂).

// symbols are the symbols drawn with the pen of the italic, by name.
var symbols = map[string]struct {
	r       rune
	draw    func(Params) letter
	upright bool
}{
	"uni2113": {0x2113, ell, false},
	"uni2202": {0x2202, partial, true},
	"u1D715":  {0x1D715, partial, false},
}

// The proportions of ℓ and ∂, over the ascender: ℓ is 0.99 of the
// italic l high (0.88 to 1.02) and its loop, the white above the crossing,
// 0.44 of its height (0.37 to 0.47); ∂ is 728 high (637 to 761) and 425 wide
// (383 to 473), its bowl as high as the x-height.
const (
	// ellStartX and ellStartY place the start of the upstroke of ℓ: its
	// distance left of the centre of the stem and its height, over the
	// ascender; ellStartDir is its direction, upright, in degrees.
	ellStartX, ellStartY, ellStartDir = 0.114, 0.279, 64.5
	// ellStem is the left edge of the stem of ℓ right of the side of n, over
	// the ascender, which leaves room for the start of the upstroke.
	ellStem = 0.12
	// ellLoopRight is the reach of the loop right of the centre of the stem,
	// ellLoopTop that of its top, and ellLoopSide the height of its right
	// side, over the ascender; its left side is the stem.  ellTopDir is the
	// direction of the stroke at the top node, a little downhill, so that
	// the top of the outer edge falls between the nodes.
	ellLoopRight, ellLoopTop, ellLoopSide, ellTopDir = 0.233, 0.052, 0.751, 172.7
	// partialWidth is the width of the bowl of ∂ over that of o, and
	// partialBack the reach of its back right of the bowl, over the ascender.
	partialWidth = 1.0
	partialBack  = 0.041
	// partialCrest is the crest of the top of ∂ right of the middle of the
	// bowl, over the width of the bowl.
	partialCrest = 0.115
	// partialSide is the height of the right extreme of ∂, where its back
	// turns into the bowl, over the ascender.
	partialSide = 0.475
	// partialEnd places the end of the stroke, inside the back: its
	// distance left of the centre line of the back and below the right
	// extreme, and its direction, upright, in degrees.
	partialEndIn, partialEndDepth, partialEndDir = 10.5, 2.6, -53.0
	// partialBottom and partialTop move the nodes at the bottom and the top
	// of the bowl right of its middle.
	partialBottom, partialTop = -7.6, -16.0
)

// ell draws ℓ: one stroke of the pen from a start low on the left, up to the
// right through the stem, round a loop over the top to the left and down the
// stem into the exit of l, crossing the upstroke.
func ell(p Params) letter {
	asc := p.XHeight + p.AscenderEntryRise
	e := p.SideStemLeft + ellStem*asc
	c := e + p.Stem/2
	wh := p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: auto}
	}
	nodes := []stroke.Node{
		{P: pt(c-ellStartX*asc, ellStartY*asc), Dir: ellStartDir, K: auto, W: p.EndW, DW: auto},
		node(c+ellLoopRight*asc, ellLoopSide*asc, 90),
		node(c+ellLoopTop*asc, asc-wh/2, ellTopDir),
	}
	nodes = append(nodes, p.exit(e)...)
	return letter{
		strokes: []stroke.Stroke{{Nib: p.nib(), Nodes: nodes, Start: round, End: round}},
		advance: e + p.Stem + p.SideStemRight,
	}
}

// partial draws ∂: one stroke of the pen from the ball at the upper left over
// the top to the right, down the back, round the bottom of the bowl, up its
// left side and over its top into the back, where it ends inside the stroke.
// The bowl is partialWidth of o wide and as high as the x-height.
func partial(p Params) letter {
	asc := p.XHeight + p.AscenderEntryRise
	l := p.SideRoundLeft
	w := partialWidth * p.OBowlWidth
	r := l + w
	cx := (l + r) / 2
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	cy := (top + bottom) / 2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	back := r + partialBack*asc
	yr := partialSide * asc // the height of the right extreme
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: auto}
	}
	return letter{
		strokes: []stroke.Stroke{{
			Nib: p.nib(),
			Nodes: []stroke.Node{
				{P: pt(cx+partialCrest*w, asc+p.Overshoot-p.BallStalk/2), Dir: 0, K: auto, W: p.BallStalk, DW: 0},
				node(back-wv/2, yr, 270),
				node(cx+partialBottom, bottom+wh/2, 180),
				node(l+wv/2, cy, 90),
				node(cx+partialTop, top-wh/2, 0),
				{P: pt(back-wv/2-partialEndIn, yr-partialEndDepth), Dir: partialEndDir, K: auto, W: auto, DW: auto},
			},
			Start: p.ballAtDepth(-(p.CBallLow - p.Ball/2)),
			End:   hidden,
		}},
		advance: back + p.SideRoundRight,
	}
}

// SymbolNames returns the names of the symbols drawn with the pen of the
// italic.
func SymbolNames() []string { return []string{"uni2113", "uni2202", "u1D715"} }

// BuildSymbol returns the symbol of the given name, as Build returns a
// letter; an upright symbol is drawn with the slant zero, and ∂ in both
// forms with the pen held level.
func BuildSymbol(p Params, name string) (*Glyph, error) {
	d, ok := symbols[name]
	if !ok {
		return nil, fmt.Errorf("italic: no symbol %q", name)
	}
	if d.upright {
		p.SlantNum = 0
	}
	if name != "uni2113" {
		p.Stress = 0
	}
	return buildLetter(p, name, d.r, d.draw(p))
}
