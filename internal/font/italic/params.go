// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

// Package italic draws the letters of Bedern Italic from pen strokes
// (bedern/internal/stroke): each letter is described upright, sheared by the slant
// 1/4 on the lattice of whole units and moved back by ShearHeight times the
// slant: the shear leaves it in place at half the x-height.
package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
)

// Params are the parameters of the italic, all in the upright description and
// in font units unless a comment says otherwise.
type Params struct {
	// The heights, those of the roman.
	XHeight   float64 // the x-height
	Ascender  float64 // the ascender
	Overshoot float64 // the overshoot of round letters at the x-height and the baseline

	SlantNum, SlantDen int     // the slant, the tangent SlantNum/SlantDen
	ShearHeight        float64 // the height at which the shear leaves a letter in place, and the side spaces are measured

	Stem   float64 // the stem, measured horizontally in the italic
	Curve  float64 // the major axis of the pen: the thick of the round strokes
	Thin   float64 // the minor axis of the pen: every thinnest place, the start of every entry and the end of every exit
	Stress float64 // the turn of the major axis of the pen anticlockwise from the horizontal, upright, in degrees
	Hair   float64 // the hairline: the turn of every exit at the baseline, the bars of t

	Branch        float64 // the height at which the underside of the arch meets the left stem
	ArchPeak      float64 // the place of the peak of the arch between the stem centres, a fraction
	ArchStartW    float64 // the thickness of the arch where it starts inside the left stem
	ArchStartBack float64 // how far back along its direction the arch starts from where its underside meets the stem
	ArchStartDir  float64 // the direction in which the arch leaves the left stem, upright, in degrees
	ArchOver      float64 // how far the outer edge of the arch rises above the x-height
	Shoulder      float64 // the height down to which the arch runs before the right stem is straight
	ShoulderTop   float64 // the height of the shoulder node, where the arch runs vertically in the italic
	ShoulderW     float64 // the thickness of the arch at the shoulder node, at right angles

	NCounter          float64 // the counter of n between the stem edges, horizontally at ShearHeight
	MCounter          float64 // the counters of m
	SideStemLeft      float64 // the side of n left of its left stem, at ShearHeight
	SideStemRight     float64 // the side of n right of its right stem, at ShearHeight
	OBowlWidth        float64 // the width of the bowl of o between its outer extremes, upright
	SideRoundLeft     float64 // the side of o on the left
	SideRoundRight    float64 // the side of o on the right
	OSuperness        float64 // the squareness of the outer edge of o, upright
	URiseDir          float64 // the direction of the rise of u where it meets the right stem, upright, in degrees
	URiseBack         float64 // how far into the right stem the rise of u runs on beyond where it meets it
	URiseW            float64 // the thickness of the rise of u inside the right stem
	URise             float64 // the height at which the rise of u meets the right stem
	EndW              float64 // the thickness of the start of the entry, at its cut
	ExitEndW          float64 // the thickness of the end of the exit, at its cut
	EntryReach        float64 // the entry: the centre of its start left of the stem edge
	EntryLow          float64 // the height of the centre of the start of the entry
	EntryTop          float64 // the top of the outer edge of the entry at the x-height
	EntryTopW         float64 // the thickness of the entry at its top node
	EntryTopShift     float64 // the centre of the top node of the entry right of the stem edge, where the stroke runs horizontally
	EntryStartDir     float64 // the direction of the start of the entry, upright, in degrees
	EntryJoin         float64 // the depth below the top of the stem at which the entry joins the stem's centre line
	EntryVert         float64 // the height below the top of the stem at which the entry runs vertically in the italic
	ExitTurn          float64 // the height at which the stem begins to turn into the exit
	ExitVert          float64 // the height at which the exit runs vertically in the italic
	ExitReach         float64 // the ink of the exit right of the stem edge
	ExitTipHeight     float64 // the highest ink of the exit
	ExitAngle         float64 // the direction of the end of the exit, upright, in degrees
	ExitLowShift      float64 // the lowest node of the exit right of the centre of the stem
	ExitEndK          float64 // the curvature of the centre line of the exit at its end, turning left
	AscenderEntryRise float64 // how far the entry of an ascender stands above that at the x-height, a multiple of SlantDen
	TopLift           float64 // how far the parts at the top of the x-height (the entry, the peak and the shoulder of the arch) stand above their place by the x-height, moving only by multiples of SlantDen (WithXHeight)

	Ball                       float64 // the diameter of the ball terminal
	BallStalk                  float64 // the thickness of the stalk where the curl into the ball begins
	BallNeck                   float64 // the distance of the tip of the wedge from the outer edge of the curl
	BallOpen                   float64 // the opening of the wedge at its tip, in degrees
	BallPowerMin, BallPowerMax float64 // the range of the power of the length by which the curvature of the outer edge of the curl rises

	RBallTop   float64 // the top of the ball of r below the x-height
	RBallReach float64 // the rightmost ink of the ball of r right of the right edge of the stem
	RWidth     float64 // the set width of r
	RCounter   float64 // the counter of the arch of n of which the arm of r is the part up to the peak
	RStalkW    float64 // the thickness of the arm of r at its peak; 0: the pen's
	RStartDir  float64 // the direction in which the arm of r leaves the stem, upright, in degrees; 0: that of n

	Dot       float64 // the diameter of the dot of i
	DotCentre float64 // the height of the centre of the dot
	DotOffset float64 // the distance of the centre of the dot right of the vertical through the centre of the stem at the x-height, that of the roman (design.DotOffset)

	HLeftLess float64 // how much less the left side of h is than that of n

	CrossbarLeft, CrossbarRight float64 // the crossbar of t, left and right of the stem edges
	TSideLeft                   float64 // the left side of the stem of t, at ShearHeight
	TWidth                      float64 // the set width of t
	TTop                        float64 // the top of the stem of t, at its oblique cut
	TCutAngle                   float64 // the direction of the oblique cut at the top of t, upright, in degrees

	ExitLowW float64 // the thickness of the exit at its lowest node, where it turns at the baseline

	RoundParams             // c, e, a, d, g and q
	Diag                    // the parameters of b, p, s, k, v, w, x and z (params_diag.go)
	Special                 // the parameters of ß, æ, œ and the other special letters (params_special.go)
	Gk          GreekParams // the proportions of the Greek lowercase (params_greek.go)
}

// Default returns the starting values of the parameters, with the stem
// settled by the colour test.
func Default() Params {
	r := design.Regular()
	return Params{
		XHeight:   r.XHeight,
		Ascender:  r.Ascender,
		Overshoot: r.OvershootLower,

		SlantNum: 1, SlantDen: 4,
		ShearHeight: 224,

		Stem:   74,
		Curve:  81.6,
		Thin:   28.5,
		Stress: 7,
		Hair:   32.3,

		Branch:        232,
		ArchPeak:      0.58,
		ArchStartW:    34,
		ArchStartBack: 50,
		ArchStartDir:  65,
		ArchOver:      11,
		Shoulder:      272,
		ShoulderTop:   368,
		ShoulderW:     66.4,

		NCounter:          185,
		MCounter:          181,
		SideStemLeft:      93,
		SideStemRight:     92,
		OBowlWidth:        384,
		SideRoundLeft:     22,
		SideRoundRight:    23,
		OSuperness:        0.72,
		URiseDir:          72,
		URiseBack:         20,
		URiseW:            28.5,
		URise:             188,
		EndW:              22.8,
		ExitEndW:          18,
		EntryReach:        88,
		EntryLow:          378,
		EntryTop:          460,
		EntryTopW:         26,
		EntryTopShift:     -45,
		EntryStartDir:     85,
		EntryJoin:         120,
		EntryVert:         70,
		ExitTurn:          124,
		ExitVert:          92,
		ExitReach:         104,
		ExitTipHeight:     110,
		ExitAngle:         75,
		ExitLowShift:      78,
		AscenderEntryRise: 244,

		Ball:         80,
		BallStalk:    r.Figures.CurlStalk,
		BallNeck:     r.Figures.CurlNeck,
		BallOpen:     53,
		BallPowerMin: 1, BallPowerMax: 2.5,

		RCounter:   95,
		RStalkW:    31,
		RStartDir:  62,
		RBallTop:   41,
		RBallReach: 175,
		RWidth:     395,

		Dot:       100,
		DotCentre: 612,
		DotOffset: design.DotOffset(r),

		HLeftLess: 13,

		CrossbarLeft: 50, CrossbarRight: 95,
		TSideLeft: 70,
		TWidth:    300,
		TTop:      566,
		TCutAngle: 35,

		ExitLowW: 40,

		RoundParams: defaultRound(),
		Diag:        defaultDiag(),
		Special:     defaultSpecial(),
		Gk:          defaultGreek(),
	}
}

// slant returns the tangent of the slant.
func (p Params) slant() float64 { return float64(p.SlantNum) / float64(p.SlantDen) }

// shift returns the move to the left after the shear: ShearHeight times the
// slant.
func (p Params) shift() float64 { return p.ShearHeight * p.slant() }

// WithXHeight returns the parameters with the x-height x and the heights
// which belong to it moved with it: the dot, the top of t and the overshoot
// of the arch move by the change, the entry and the top of the arch (its
// peak, its shoulder node and its join) by the multiple of 2 SlantDen
// nearest to it (TopLift), moving along the lattice and keeping
// their shape after the rounding, the branch and the join of the rise of u
// in proportion, the heights where curves meet stems
// to the nearest multiple of SlantDen, and the entry of the ascender such that
// the ascender stays.
func (p Params) WithXHeight(x float64) Params {
	d, f := x-p.XHeight, x/p.XHeight
	den := float64(p.SlantDen)
	lat := func(y float64) float64 { return math.Round(y/den) * den }
	p.XHeight = x
	// The run-outs of the stems put the middle of their control points on
	// the lattice, which repeats every 2 SlantDen in height.
	e := math.Round((d-p.TopLift)/(2*den)) * 2 * den
	p.EntryLow += e
	p.EntryTop += e
	p.TopLift += e - d
	p.DotCentre += d
	p.TTop += d
	p.BowlStemTop = lat(p.BowlStemTop + d)
	p.Branch *= f
	p.PBranch *= f
	p.ZStartY += d
	p.Shoulder += e
	p.ShoulderTop += e
	p.AscenderEntryRise -= e
	p.RoundParams = p.RoundParams.withXHeight(f)
	p.Gk = p.Gk.withXHeight(x-d, x, p.Ascender, den)
	p.YShoulder += e
	return p
}

// WithStem returns the parameters with the stem w, measured horizontally,
// and the pen, the hairline and the ends moved with it in proportion: the
// colour is set by the stem, and the widths stay.  The centre lines of the stems stay where they are, to
// half a unit; the sides and the counters grow by what the stem
// loses.
func (p Params) WithStem(w float64) Params {
	f := w / p.Stem
	d := p.Stem - w
	// The left edge stays on whole units: the stem's edges are
	// lines of the lattice at the baseline; for an odd change the centre
	// lines move by half a unit.
	l := math.Floor(d / 2)
	p.SideStemLeft += l
	p.SideStemRight += d - l
	// The lowest node of the exit keeps its distance from the right edge
	// of the stem; the inner turn of the exit keeps its width.
	p.ExitLowShift -= d / 2
	p.NCounter += d
	p.MCounter += d
	p.Gk = p.Gk.withStem(d)
	p.Stem = w
	p.Curve *= f
	p.Thin *= f
	p.ArchStartW *= f
	p.URiseW *= f
	p.Hair *= f
	p.EntryTopW *= f
	p.EndW *= f
	p.ExitEndW *= f
	p.ShoulderW *= f
	p.ExitLowW *= f
	return p
}
