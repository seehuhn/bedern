// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import "seehuhn.de/go/bedern/internal/fair"

// The findings of the smoothness check on the glyphs of TS1 which the refits
// (ts1_repairs.go) leave: defects of the derivation or of the roman, none of
// which shows at 10 pt, each with its cause.  The regular's own judged findings reach the
// italic through the baseline of the roman (inherited).

const (
	corner5           = "the short round of the inner corner where the stem of the 5 meets its bowl, the defect of the roman at that place moved by the deformation, as on the lining 5"
	oldFive           = "the inner corner where the stem of the 5 meets its bowl, which the roman judges (short and inflection in its baseline); the zone of the ball which keeps the stem on the lattice (the others leave a handle off it) turns the corner tighter than the lining 5"
	oldFiveInflection = "the inflection which the roman judges where the bracket of the stroke runs into the inner edge of the bowl, 10 units right of the lowest point of the bracket as in the roman (9); the zone of the ball moves that point 17 units left of the place which the shear gives it, so that the roman's judgement, placed by the shear, misses it by 16 units; a longer first handle of the curve from the lowest point, which would move it into reach, leaves the curvature of the bracket there with a step (r 22 to r 36)"
	sterling          = "as on the sterling sign, whose ball and base the lira sign takes unchanged: "
	extreme06         = "the roman's node at an upright extreme with a short handle moves along the edge with the shear and lies too close to the next node for one of its own; the curve runs less than 0.7 beyond its nodes"
	dollarOld         = "the curve above the left edge of the upper tick runs into that edge, which leans 9 degrees and so is off the lattice, with a handle of 11 units; the rounding turns it by 4 degrees, and a longer handle on the edge, tried at 12 to 24 units, leaves a jump or a wiggle at the node at the extreme before it"
)

func init() {
	add := func(name string, es ...judged) { judgedFindings[name] = append(judgedFindings[name], es...) }

	add("five.taboldstyle", judged{Kind: fair.Short, X: 82, Y: 170, Dir: -34, Defect: true, Reason: corner5})
	add("five.prop", judged{Kind: fair.Short, X: 136, Y: 375, Dir: -53, Defect: true, Reason: corner5})
	add("five.oldstyle",
		judged{Kind: fair.Kink, X: 60, Y: 201, Dir: -106, Defect: true, Reason: oldFive},
		judged{Kind: fair.Jump, X: 48, Y: 172, Dir: -90, Defect: true, Reason: oldFive},
		judged{Kind: fair.Short, X: 49, Y: 170, Dir: -24, Defect: true, Reason: oldFive},
		judged{Kind: fair.Inflection, X: 63, Y: 170, Dir: 6, Defect: true, Reason: oldFiveInflection})

	add("uni20A9", judged{Kind: extremeKind, X: 277, Y: 617, Dir: 90, Reason: extreme06})
	add("florin", judged{Kind: fair.Wiggle, X: -52, Y: -167, Dir: 180, Defect: true, Reason: "the curvature of the tail of j, whose inner edge leaves the stem on a spiral in the roman, dips between two extremes after the shear"})

	add("lira",
		judged{Kind: fair.Jump, X: 147, Y: 13, Dir: -34, Defect: true, Reason: sterling + "the curvature of the base stroke jumps from r 631 to r 129"},
		judged{Kind: fair.Kink, X: 493, Y: 557, Dir: -79, Defect: true, Reason: sterling + "the short curve before the tip of the wedge of the ball has handles of 2 units, which the rounding turns by 7 degrees"},
		judged{Kind: fair.Short, X: 494, Y: 551, Dir: -81, Defect: true, Reason: sterling + "that short curve, 13 units"},
		judged{Kind: fair.Jump, X: 194, Y: 150, Dir: 62, Defect: true, Reason: sterling + "the curvature of the stem jumps from r 462 to r 97 where it enters the base"})

	add("dollar.oldstyle",
		judged{Kind: fair.Kink, X: 448, Y: 513, Dir: -95, Defect: true, Reason: dollarOld})
}
