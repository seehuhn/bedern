// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/fair"

// The inflections of the edges of four Greek letters which a search over the
// values of the Greek letters (a random search of the 180 values in 40
// minutes at a time, three times, which cleaned γ, ς, ξ and ζ) did not
// remove, and which belong to the forms: a stroke which changes the sense of
// its turn changes the sense of the turn of both its edges.
func init() {
	intendedFindings["kappaVar"] = []intended{
		{Kind: fair.Inflection, X: 43, Y: 380, Dir: 69.9, Reason: "the curved stem of ϰ turns out of the hook into its bow: the sense of its turn changes"},
		{Kind: fair.Inflection, X: 96, Y: 296, Dir: -111.3, Reason: "the curved stem of ϰ turns back at the bow"},
		{Kind: fair.Inflection, X: 59, Y: 159, Dir: -100.5, Reason: "the curved stem of ϰ turns again where it runs down to the foot"},
		{Kind: fair.Inflection, X: 259, Y: 118, Dir: -90, Reason: "the leg of ϰ leaves the stem against the sense of the turn of the stem"},
		{Kind: fair.Inflection, X: 327, Y: 82, Dir: 81.4, Reason: "the end of the leg of ϰ turns up against the sense of the leg"},
	}
	intendedFindings["nu"] = []intended{
		{Kind: fair.Inflection, X: 243, Y: 132, Dir: 27.1, Reason: "the light rise of ν leaves the point of the stem straight and turns into the free side: the sense of its turn changes once"},
	}
	// The script forms of ν, drawn with the heavier pen, show more
	// inflections of the same turns, whose peaks the text form keeps under
	// the threshold.
	rise := "the inner edge of the rise of ν turns into the free side as its outer edge does, so that it inflects once too"
	end := "the end of ν is bent back to the left at the x-height (NuEndDir), so that the outer edge of the free side inflects once below it, as the tips of υ do"
	intendedFindings["nu/1"] = []intended{
		{Kind: fair.Inflection, X: 262, Y: 166, Dir: -160, Reason: rise + " (peaks r 124 and r 463)"},
		{Kind: fair.Inflection, X: 517, Y: 392, Dir: 90, Reason: end + " (peaks r 141 and r 411)"},
	}
	intendedFindings["nu/2"] = []intended{
		{Kind: fair.Inflection, X: 269, Y: 167, Dir: -168, Reason: rise + " (peaks r 79 and r 378)"},
	}
	start := "the inner edge of the start of ζ and ξ, which turns to the left, runs into the half circle of the round end"
	intendedFindings["zeta"] = []intended{
		{Kind: fair.Inflection, X: 240, Y: 707, Dir: 69.7, Reason: start},
	}
	intendedFindings["xi"] = []intended{
		{Kind: fair.Inflection, X: 216, Y: 707, Dir: 69.7, Reason: start},
	}
	intendedFindings["rhoVar"] = []intended{
		{Kind: fair.Inflection, X: 161, Y: -183, Dir: 109.5, Reason: "the end of the hook of ϱ turns back against the sense of the hook, at a radius of 5 to 18"},
	}
	intendedFindings["thetaVar"] = []intended{
		{Kind: fair.Inflection, X: 84, Y: 240, Dir: -108.3, Reason: "the loop of ϑ turns from the bowl into the loop: the sense of its turn changes"},
		{Kind: fair.Inflection, X: 559, Y: 490, Dir: 69.3, Reason: "the right side of ϑ turns into the loop"},
		{Kind: fair.Inflection, X: 137, Y: 155, Dir: 75.9, Reason: "the loop of ϑ crosses the bowl: the edge turns the other way"},
	}
}
