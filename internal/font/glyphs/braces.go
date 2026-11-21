// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The braces.
//
// A brace runs from the bottom to the top of the parentheses and is two
// halves, the lower the image of the upper in the horizontal through the
// middle.  Each half is one stroke (symstroke.go): from the end, BraceEnd thick
// and cut vertically, it turns into a straight vertical arm of BraceArm, and
// from the arm it turns the other way into the point at the middle, so that each
// half is one S-shaped curve whose turns pass into the straight arm without a
// jump of curvature.  At the point the stroke leans BraceTipAngle degrees from
// the horizontal and is cut along the horizontal through the middle: the outer
// edges of the two halves meet there in the point, the corner of the brace, and
// the inner edges in a notch BraceTip to the right of it.  The right brace is
// the mirror image of the left.
//
// Side kinds: open, BraceSide to the point and to the ends.

// braceLeftChain returns the outline of the left brace.
func braceLeftChain(p design.Params) chain {
	s := p.Sym
	x0, x1 := s.BraceSide, s.BraceSide+s.BraceInk
	top, mid := p.BracketTop(), (p.BracketTop()+p.BracketBottom())/2
	xa := x0 + 0.52*s.BraceInk // the centre line of the arm
	turn := x1 - xa            // the reach of the turn from the end into the arm
	sts := []station{
		{X: x1, Y: top - s.BraceEnd/2, Dir: 180, W: s.BraceEnd, Nib: -90},
		{X: xa, Y: top - 1.15*turn, Dir: -90, W: s.BraceArm, Flat: true},
		{X: xa, Y: mid + 1.0*turn, Dir: -90, W: s.BraceArm},
		{X: x0 + s.BraceTip/2, Y: mid, Dir: 180 + s.BraceTipAngle, W: s.BraceTip, Nib: 360},
	}
	right, left := strokeOpen(sts)
	return join(
		right,
		right.flipY(mid),
		lineChain(outline.Point{X: right.start().X, Y: 2*mid - right.start().Y}, outline.Point{X: left.end().X, Y: 2*mid - left.end().Y}),
		left.flipY(mid),
		left,
		lineChain(left.end(), right.start()),
	)
}

// braceleft draws the left brace, see the comment above.
func braceleft(p design.Params) (outline.Path, float64) {
	return chainsPath(braceLeftChain(p)), 2*p.Sym.BraceSide + p.Sym.BraceInk
}

// braceright draws the right brace, the mirror image of the left.
func braceright(p design.Params) (outline.Path, float64) {
	adv := 2*p.Sym.BraceSide + p.Sym.BraceInk
	return chainsPath(braceLeftChain(p).mirrorX(adv / 2)), adv
}
