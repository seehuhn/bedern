// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "math"

// RoundParams are the parameters of the round letters c and e and of the
// bowl-and-stem letters a, d, g and q, in the upright description.
type RoundParams struct {
	CBowlWidth float64 // the width of the round of c between its outer left extreme and where its closed bowl has the right one
	CWidth     float64 // the set width of c
	CTipInk    float64 // the rightmost ink of the end of the lower stroke of c right of the left of the ink
	CBallLow   float64 // the lowest ink of the ball of c below the top of the letter
	CBallIn    float64 // how far the rightmost ink of the ball of c stays inside that of the lower stroke

	EBowlWidth float64 // the width of the round of e between its outer extremes
	EWidth     float64 // the set width of e
	EBar       float64 // the height of the centre line of the bar of e
	EBarW      float64 // the thickness of the bar of e
	EBarIn     float64 // how far the bar of e starts inside the outer left edge
	EEye       float64 // the height of the right extreme of the eye of e
	ETipInk    float64 // the rightmost ink of the end of the lower stroke of e right of the left of the ink

	BowlReach                   float64 // the outer left extreme of the bowl of a, d, g and q left of the left edge of the stem
	BowlTopJoin                 float64 // the depth below the x-height at which the bowl leaves the stem
	BowlTopDir                  float64 // the direction in which the bowl leaves the stem, upright, in degrees
	BowlLowJoin                 float64 // the height at which the bowl runs back into the stem
	BowlLowDir                  float64 // the direction in which it runs into the stem, upright, in degrees
	BowlPlainJoin, BowlPlainDir float64 // the same on a plain stem without exit, of q and g
	BowlJoinW                   float64 // the thickness of the bowl where it leaves and meets the stem
	BowlDepth                   float64 // how far inside the stem the ends of the bowl lie
	BowlStemTop                 float64 // the height of the flat cut at the top of the stem of a, g and q, a multiple of SlantDen
	BowlBottomDir               float64 // the turn of the bowl from the horizontal at its bottom node, upright; the extreme of its outer edge lies a whole piece away from it
	BowlTop, BowlBottom         float64 // the nodes at the top and the bottom of the bowl right of its middle; negative: its top and bottom lie further left than on o

	QWidth float64 // the set width of q

	GBowl float64 // the bowl of g over that of a, a fraction

	TailTurn      float64 // the tail of g, j and y: the depth below the baseline down to which the stem runs straight
	TailDepth     float64 // the lowest ink of the tail below the baseline
	TailReach     float64 // the leftmost ink of the ball of the tail left of the left edge of the stem
	TailBallDepth float64 // the depth of the centre of the ball above the tangent of the outer edge at the lowest point of the tail
	TailLow       float64 // the lowest node of the tail left of the centre of the stem

	JSideLeft  float64 // the side of j left of its stem, at ShearHeight
	JWidth     float64 // the set width of j
	YCounter   float64 // the counter of y between the stem edges, at ShearHeight
	YRiseDir   float64 // the direction of the rise of y where it meets the right stem, upright, in degrees
	YShoulder  float64 // the arch of n whose turning round is the turn of y: the height down to which it runs before the right stem is straight (Shoulder)
	YSideRight float64 // the side of y right of its right stem
	FSideLeft  float64 // the left edge of the stem of f, at ShearHeight
	FWidth     float64 // the set width of f
	FHookAxis  float64 // the height about which the hook at the top of f is the tail turned through half a turn
	FBarLeft   float64 // the crossbar of f: its reach left of the left edge of the stem
	FBarRight  float64 // its reach right of the right edge of the stem

	EBarRise float64 // the direction of the bar of e, upright, in degrees above the horizontal
	EBarRun  float64 // how far the bar of e runs straight from its start, as a share of the width of the round
	ETurn    float64 // how far the top node of e lies left of the middle of the round and the lowest node right of that of o

	TailWider  float64 // how much further left than TailLow the lowest node of a tail lies; set by the letter
	GTailWider float64 // that of the tail of g, which reaches under its bowl
	GTailTurn  float64 // TailTurn of the tail of g, which turns early under its bowl
}

// defaultRound returns the starting values of the round letters.
func defaultRound() RoundParams {
	return RoundParams{
		CBowlWidth: 336,
		CWidth:     356,
		CTipInk:    344,
		CBallLow:   138,
		CBallIn:    20,

		EBowlWidth: 330,
		EWidth:     365,
		EBar:       246,
		EBarW:      28.5,
		EBarIn:     30,
		EEye:       336,
		ETipInk:    337,

		BowlReach:     293,
		BowlTopJoin:   31,
		BowlTopDir:    155,
		BowlLowJoin:   100,
		BowlLowDir:    73,
		BowlPlainJoin: 84, BowlPlainDir: 60,
		BowlJoinW:   19,
		BowlDepth:   36,
		BowlStemTop: 464,
		BowlTop:     42, BowlBottom: 32, BowlBottomDir: -1,

		QWidth: 461,

		GBowl: 1,

		TailTurn:      80,
		TailDepth:     200,
		TailReach:     150,
		TailBallDepth: 88,
		TailLow:       80,

		JSideLeft:  100,
		JWidth:     265,
		YCounter:   156,
		YRiseDir:   70,
		YShoulder:  268,
		YSideRight: 60,
		FSideLeft:  110,
		FWidth:     300,
		FHookAxis:  252,
		FBarLeft:   70,
		FBarRight:  95,
		EBarRise:   12,
		EBarRun:    0.55,
		ETurn:      2,

		GTailWider: 100,
		GTailTurn:  40,
	}
}

// withXHeight returns the parameters for the x-height changed by the factor
// f: the heights of the bar and the eye of e and the ball of c move in
// proportion.
func (r RoundParams) withXHeight(f float64) RoundParams {
	r.EBar *= f
	r.EEye *= f
	r.CBallLow *= f
	r.BowlLowJoin = math.Round(r.BowlLowJoin*f/4) * 4
	r.BowlPlainJoin = math.Round(r.BowlPlainJoin*f/4) * 4
	r.FHookAxis = math.Round(r.FHookAxis*f/2) * 2
	return r
}
