// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/stroke"
)

// The leg of k bows out to the right from the stem and turns the other way
// into the exit, so that each of its edges changes from convex to concave
// once.
func init() {
	why := "the leg of k bows out from the stem and turns the other way into the exit: each edge inflects once"
	intendedFindings["k"] = []intended{
		{Kind: fair.Inflection, X: 259, Y: 118, Dir: -90, Reason: why},
		{Kind: fair.Inflection, X: 327, Y: 80, Dir: 81, Reason: why},
	}
}

// k draws the k: the ascender with the entry at its top and a plain foot;
// the arm, a plain stroke of the pen from inside the stem at KArmY up to
// an end at the upper right, KArmX right of the stem and KArmDrop below the
// x-height, which bows a little down and to the right (its ends turn
// KArmBow from the chord) and ends in a cut KArmEndW of the minor axis of
// the pen thick; and the leg, which leaves the stem where the arm does,
// bows out to the right, runs vertically in the italic at KLegVert, where
// it has the stem weight, and runs on into the turn and the end of the
// exit, so that each of its edges inflects once on the way.  Its left side
// is that of h.
func k(p Params) letter {
	e := p.SideStemLeft - p.HLeftLess
	s := pt(e+p.Stem/2, p.KArmY)
	end := pt(e+p.Stem+p.KArmX, p.XHeight-p.KArmDrop)
	ch := degrees(math.Atan2(end.Y-s.Y, end.X-s.X))
	arm := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: s, Dir: ch - p.KArmBow, K: auto, W: auto, DW: auto},
			{P: end, Dir: ch + p.KArmBow, K: auto, W: p.KArmEndW * p.Thin, DW: 0},
		},
		Start: hidden, End: round,
	}
	e2 := e + p.Stem + p.KLegCounter
	c2 := e2 + p.Stem/2
	// The leg leaves the stem in the direction which makes the stretch to
	// the node where it runs vertically a circular arc.
	start := pt(e+p.Stem/2, p.KLegY)
	vert := p.vertNode(c2, p.KLegVert, p.KLegVert, p.across(p.Stem))
	chord := degrees(math.Atan2(vert.P.Y-start.Y, vert.P.X-start.X))
	leg := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: start, Dir: 2*chord - vert.Dir, K: auto, W: p.KLegStartW, DW: auto},
			vert,
		},
		Start: hidden, End: round,
	}
	leg.Nodes = append(leg.Nodes, p.exit(e2)[2:]...)
	return letter{
		strokes: []stroke.Stroke{
			p.stem(e, p.XHeight+p.AscenderEntryRise, false),
			arm,
			leg,
		},
		advance: e2 + p.Stem + p.SideStemRight,
	}
}
