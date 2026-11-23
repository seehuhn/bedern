// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

// Package designtest supplies the perturbed parameter sets: the regular
// parameters with some values replaced.
package designtest

import (
	"math"
	"slices"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
)

// A Set is a named, valid parameter set.
type Set struct {
	Name   string
	Params design.Params
}

// routineOrder lists the perturbed sets in the order in which Sets takes them.
// The first sets move all lowercase proportions together (xheight430-serif90),
// all capitals (caps-down), and the weight and the counters
// (stem95-counter150).  Then follows one set each for the groups e-a-s, f-t-j,
// g, v-w-y, x-z-k and the punctuation, before the second set of a group and
// the sets which move one parameter alone.
var routineOrder = []string{
	"xheight430-serif90", "caps-down", "stem95-counter150",
	"e-a-s-up", "f-t-j-down", "g-up", "v-w-y-down", "x-z-k-up", "punct-up",
	"caps-up", "xheight480-serif40", "stem70-counter230",
	"e-a-s-down", "f-t-j-up", "g-down", "v-w-y-up", "x-z-k-down", "punct-down",
	"curve100", "thin26",
	"stem70", "stem95", "xheight430", "xheight480", "counter150", "counter230", "serif40", "serif90",
}

// Sets returns the perturbed sets: all of them, or if routine is set the first
// n of them in the order of routineOrder.  A negative n means all.
func Sets(routine bool, n int) []Set {
	all := Perturbed()
	if !routine || n < 0 {
		return all
	}
	var out []Set
	for _, name := range routineOrder {
		if len(out) == n {
			break
		}
		for _, s := range all {
			if s.Name == name {
				out = append(out, s)
			}
		}
	}
	return out
}

// Perturbed returns the parameter sets which differ from the regular
// parameters.  Every set is the regular parameters with some values replaced.
// The sets are: stem 70 and 95, x-height 430 and 480, counter 150 and 230,
// serif length 40 and 90, four combinations of these, and the sets which move
// the following.  The thickest round stroke (100) and the thin parts (26 and a
// hairline of 30).  One parameter each of e, a and s, up in one set and down in
// the other (the height of the bar of e, the reach of the arc of a and the drop
// of the upper terminal of s).  The same for f, t and j (the reach, hang and
// joins of the hooks and the tail, the reach of the bars and the height of the
// tip of t).  Two sets which move the parameters of g (the sizes and weights of
// the bowl and of the loop, the size, reach and height of the ear and its ball,
// and the place of the link).  Two sets which move the parameters of v, w and y
// (the widths of the ink, the serifs, the vertices, the apex of w and the tail
// of y), two for x, z and k, two for the capitals and two for the punctuation.
// In the sets which scale the x-height the bowl and the ear of g move with it.
// Every set passes Params.Validate and glyphs.Check.  A change of the serif
// length scales BracketReach, SerifJutInner and QSideRight in the same
// proportion and moves BSideLeft by the change.  A change of the x-height moves the notch and the join of the arm
// of r with it, and the upper notch and join of the bowls of p and q.  A change
// of the stem changes AscenderStem with it.
func Perturbed() []Set {
	return slices.Clone(perturbed())
}

// perturbed builds the sets of Perturbed once.
var perturbed = sync.OnceValue(func() []Set {
	var sets []Set
	add := func(name string, f ...func(p *design.Params)) {
		p := design.Regular()
		for _, g := range f {
			g(&p)
		}
		sets = append(sets, Set{name, p})
	}
	stem := func(v float64) func(*design.Params) { return func(p *design.Params) { p.Stem, p.AscenderStem = v, v } }
	// A change of the x-height moves the join and the notch of the arm of r by the
	// same amount, and the upper join and notch of the bowls of p and q.
	xh := func(v float64) func(*design.Params) {
		return func(p *design.Params) {
			d := v - p.XHeight
			p.XHeight, p.RNotch, p.RJoin = v, p.RNotch+d, p.RJoin+d
			// The top of t keeps its height: TRise changes by -d.
			p.TRise -= d
			p.QStemTop += d
			// The upper notch and join of the bowls of p and q move with the
			// x-height.
			p.PBowl.NotchUpper, p.PBowl.JoinUpper = p.PBowl.NotchUpper+d, p.PBowl.JoinUpper+d
			p.QBowl.NotchUpper, p.QBowl.JoinUpper = p.QBowl.NotchUpper+d, p.QBowl.JoinUpper+d
			// The bar of e, and with it the top of the bowl of a, moves with the
			// x-height.
			p.EBarHeight += d
			// The apex of w moves with the x-height.
			p.WApexHeight += d
			// The crossing of x moves by half the change and the junction of the
			// arm and the leg of k by 0.6 of the change.
			p.XCross += d / 2
			p.KJunction += 0.6 * d
		}
	}
	counter := func(v float64) func(*design.Params) { return func(p *design.Params) { p.NCounter = v } }
	serif := func(v float64) func(*design.Params) {
		return func(p *design.Params) {
			r := v / p.SerifJut
			// The head serif of b keeps its distance from the edge of the advance.
			p.BSideLeft += v - p.SerifJut
			p.SerifJut = v
			p.BracketReach = math.Round(p.BracketReach * r)
			p.SerifJutInner = math.Round(p.SerifJutInner * r)
			// QSideRight scales in the same proportion.
			p.QSideRight = math.Round(p.QSideRight * r)
			// The serifs of the diagonal letters scale by half the change.
			h := 1 + (r-1)/2
			p.DiagOverThick, p.DiagOverThin = math.Round(p.DiagOverThick*h), math.Round(p.DiagOverThin*h)
			// The serifs of x scale by half that change again.
			hx := 1 + (h-1)/2
			p.XOverThick, p.XOverThin = math.Round(p.XOverThick*hx), math.Round(p.XOverThin*hx)
			// KOverLeg does not change.
			p.KOverArm = math.Round(p.KOverArm * h)
		}
	}
	add("stem70", stem(70))
	add("stem95", stem(95))
	add("xheight430", xh(430))
	add("xheight480", xh(480))
	add("counter150", counter(150))
	add("counter230", counter(230))
	add("serif40", serif(40))
	add("serif90", serif(90))
	add("stem70-counter230", stem(70), counter(230))
	add("stem95-counter150", stem(95), counter(150))
	add("xheight430-serif90", xh(430), serif(90))
	add("xheight480-serif40", xh(480), serif(40))
	// The round stroke thicker by 6 and the thin parts thinner by 6.
	add("curve100", func(p *design.Params) { p.CurveLower = 100 })
	add("thin26", func(p *design.Params) {
		p.ThinLower, p.Hair, p.GLinkThick = 26, 30, 36
	})
	// e, a and s: the bar of e, the reach of the arc of a, and the beaks of s
	// with a longer and heavier serif hanging from a lower corner; and the
	// opposite.
	add("e-a-s-up", func(p *design.Params) {
		p.EBarHeight, p.AArcReach = 275, 26
		p.SUpperDrop, p.SBeakDrop, p.SSerifThick = 176, 98, 44
	})
	add("e-a-s-down", func(p *design.Params) {
		p.EBarHeight, p.AArcReach = 235, 14
		p.SUpperDrop, p.SBeakDrop, p.SSerifThick = 136, 86, 36
	})
	// f, t and j: the reach of the hooks and the tail, the
	// heights at which the edges leave the stems, the extent of the bars and the
	// height of the tip of t.
	add("f-t-j-up", func(p *design.Params) {
		p.FReach, p.FOuterDrop, p.FInnerDrop, p.FBarRight = 188, 130, 112, 120
		p.TReach, p.TTipHeight, p.TBarRight = 190, 125, 150
		p.JStemLeft, p.JReach, p.JOuterDrop, p.JInnerDrop = 135, 180, 95, 110
	})
	add("f-t-j-down", func(p *design.Params) {
		p.FReach, p.FOuterDrop, p.FInnerDrop, p.FBarRight = 172, 110, 90, 100
		p.TReach, p.TTipHeight, p.TBarRight = 154, 95, 128
		p.JStemLeft, p.JReach, p.JOuterDrop, p.JInnerDrop = 120, 160, 70, 85
	})
	// g: a larger, heavier bowl with a longer, higher ear and a larger ball, a
	// deeper and heavier loop and a link further to the right; and the
	// opposite.
	add("g-up", func(p *design.Params) {
		p.GBowlWidth, p.GBowlHeight, p.GBowlThick, p.GBowlInset = 335, 322, 86, 38
		p.GLoopWidth, p.GLoopDepth, p.GLoopFloor, p.GLoopTopThick = 450, 210, 64, 58
		p.GLoopLeftThick, p.GLoopRightThick, p.GLinkThick, p.GLinkLeft, p.GLinkDrop = 70, 72, 46, 50, 36
		p.GEarBall, p.GEarReach, p.GEarRise, p.GEarDrop, p.GEarThick = 70, 74, 24, 40, 34
	})
	add("g-down", func(p *design.Params) {
		p.GBowlWidth, p.GBowlHeight, p.GBowlThick, p.GBowlInset = 310, 300, 72, 26
		p.GLoopWidth, p.GLoopDepth, p.GLoopFloor, p.GLoopTopThick = 425, 194, 56, 50
		p.GLoopLeftThick, p.GLoopRightThick, p.GLinkThick, p.GLinkLeft, p.GLinkDrop = 58, 62, 38, 42, 44
		p.GEarBall, p.GEarReach, p.GEarRise, p.GEarDrop, p.GEarThick = 58, 70, 14, 46, 26
	})
	// v, w and y: wider letters with longer serifs, a wider flat of the vertices
	// which dips deeper, a wider apex of w and a lighter thick diagonal of w, and a
	// deeper and longer tail of y; and the opposite.  The sets which scale the
	// x-height move the apex of w with it.
	add("v-w-y-up", func(p *design.Params) {
		p.VInk, p.WInk, p.DiagOverThick, p.DiagOverThin = 500, 710, 58, 72
		p.DiagVertexFlat, p.DiagOvershoot, p.WApexFlat, p.WStemCut = 32, 16, 70, 8
		p.YDepth, p.YReach, p.YOuterDrop, p.YInnerDrop = 207, 151, 110, 150
	})
	add("v-w-y-down", func(p *design.Params) {
		p.VInk, p.WInk, p.DiagOverThick, p.DiagOverThin = 460, 660, 46, 56
		p.DiagVertexFlat, p.DiagOvershoot, p.WApexFlat, p.WStemCut = 16, 8, 46, 2
		p.YDepth, p.YReach, p.YOuterDrop, p.YInnerDrop = 197, 129, 90, 130
	})
	// x, z and k: wider letters with a thicker thin stroke offset, a higher crossing
	// and longer, heavier beaks and a leg set further out, and the opposite.
	add("x-z-k-up", func(p *design.Params) {
		p.XInk, p.XCross, p.XThinOffset, p.XLeanDiff, p.XThick = 490, 240, -8, 0.035, 78
		p.ZInk, p.ZInkTop, p.ZDiag, p.ZUpperBeakHeight, p.ZUpperBeakThick, p.ZLowerBeakHeight, p.ZLowerBeakThick = 390, 372, 84, 160, 44, 150, 46
		p.KReach, p.KJunction, p.KLeg, p.KArmAngle, p.KLegAngle = 345, 280, 80, 49, 37
	})
	add("x-z-k-down", func(p *design.Params) {
		p.XInk, p.XCross, p.XThinOffset, p.XLeanDiff, p.XThick = 455, 224, -2, 0.02, 70
		p.ZInk, p.ZInkTop, p.ZDiag, p.ZUpperBeakHeight, p.ZUpperBeakThick, p.ZLowerBeakHeight, p.ZLowerBeakThick = 352, 330, 74, 120, 32, 118, 34
		p.KReach, p.KJunction, p.KLeg, p.KArmAngle, p.KLegAngle = 322, 258, 72, 44, 33
	})
	// The capitals: a taller cap height with a larger overshoot, heavier stems,
	// round strokes and bars, longer and heavier serifs with a taller bracket,
	// deeper arm serifs and a higher bar of H, and the opposite: a lower cap
	// height and everything lighter and shorter.  The middle arms of E and F and
	// the bar of H keep their place against the middle of the cap height.
	add("caps-up", func(p *design.Params) {
		c := &p.Caps
		d := 717 - p.CapHeight
		p.CapHeight, p.OvershootUpper = 717, 19
		c.Stem, c.Round, c.RoundThin, c.Bar, c.TBar = 106, 128, 44, 46, 52
		c.SerifJut, c.SerifThick, c.BracketHeight, c.BracketReach = 104, 38, 72, 58
		c.TopDrop, c.BottomRise, c.MiddleReach, c.TDrop, c.ArmSerifThick = 150, 165, 105, 155, 46
		c.HBarHeight, c.EMiddleHeight, c.FMiddleHeight = c.HBarHeight+d/2+6, c.EMiddleHeight+d/2, c.FMiddleHeight+d/2
		c.HCounter, c.OWidth, c.ETopArm, c.EBottomArm, c.LArm, c.TArm = 300, 680, 410, 430, 365, 285
		// C, G, S, J and Q.
		g := &c.CGSJQ
		g.BeakX, g.BeakDrop, g.BeakArc, g.BeakEnd = 620, 142, 50, 440
		g.CBeakEnd = 462
		g.CTailHeight, g.CTailReach, g.CTailCut, g.CTailArc = 158, 10, 38, 56
		g.GStemTop, g.GSerifLeft, g.GSerifRight, g.GJoin = 320, 104, 74, 215
		g.SLowerWidth, g.SUpperWidth, g.SUpperShift, g.SSpineThick, g.SSpineAngle = 455, 420, 12, 110, 22
		g.SArcThick, g.SUpperArcThick, g.SSideThick, g.SUpperDrop, g.SLowerDrop = 46, 49, 96, 240, 270
		g.JReach, g.JOuterDrop, g.JInnerDrop, g.JHeadRight, g.JBall = 255, 205, 180, 68, 108
		g.QTailStart, g.QTailJoin, g.QTailEnd, g.QTailDepth, g.QTailTip = 370, 480, 630, 175, 24
		// B, D, P, R and U.
		c.BowlRound, c.BowlFlat, c.DFlat = 118, 0.50, 0.38
		c.BWaist, c.BUpper, c.BLower = c.BWaist+d/2+6, 395, 425
		c.PBar, c.PWidth, c.RBar, c.RWidth, c.DWidth = c.PBar+d/2, 390, c.RBar+d/2, 382, 490
		c.RBowlRound, c.BLowerRound, c.BUpperRound = c.BowlRound, c.BowlRound, c.BowlRound
		c.RLegTop, c.RLegFoot, c.RLeg, c.RLegSerifLeft, c.RLegSerifRight = 90, 335, 100, 66, 66
		c.UThin, c.UCurve = 48, 245
		capDiagUp(p)
		figuresUp(p)
		symbolsUp(p)
	})
	add("caps-down", func(p *design.Params) {
		c := &p.Caps
		d := 650 - p.CapHeight
		p.CapHeight, p.OvershootUpper = 650, 12
		c.Stem, c.Round, c.RoundThin, c.Bar, c.TBar = 86, 98, 30, 32, 36
		c.SerifJut, c.SerifThick, c.BracketHeight, c.BracketReach = 78, 28, 54, 40
		c.TopDrop, c.BottomRise, c.MiddleReach, c.TDrop, c.ArmSerifThick = 105, 120, 80, 110, 34
		c.HBarHeight, c.EMiddleHeight, c.FMiddleHeight = c.HBarHeight+d/2-4, c.EMiddleHeight+d/2, c.FMiddleHeight+d/2
		c.HCounter, c.OWidth, c.ETopArm, c.EBottomArm, c.LArm, c.TArm = 270, 610, 365, 375, 320, 245
		// C, G, S, J and Q.
		g := &c.CGSJQ
		g.BeakX, g.BeakDrop, g.BeakArc, g.BeakEnd = 566, 128, 38, 395
		g.CBeakEnd = 410
		g.CTailHeight, g.CTailReach, g.CTailCut, g.CTailArc = 152, 5, 28, 42
		g.GStemTop, g.GSerifLeft, g.GSerifRight, g.GJoin = 275, 78, 56, 180
		g.SLowerWidth, g.SUpperWidth, g.SUpperShift, g.SSpineThick, g.SSpineAngle = 405, 380, 10, 88, 18
		g.SArcThick, g.SUpperArcThick, g.SSideThick, g.SUpperDrop, g.SLowerDrop = 34, 37, 76, 210, 240
		g.JReach, g.JOuterDrop, g.JInnerDrop, g.JHeadRight, g.JBall = 210, 210, 180, 52, 100
		g.QTailStart, g.QTailJoin, g.QTailEnd, g.QTailDepth, g.QTailTip = 330, 430, 560, 140, 18
		// B, D, P, R and U.
		c.BowlRound, c.BowlFlat, c.DFlat = 92, 0.40, 0.28
		c.BWaist, c.BUpper, c.BLower = c.BWaist+d/2-4, 355, 378
		c.PBar, c.PWidth, c.RBar, c.RWidth, c.DWidth = c.PBar+d/2, 350, c.RBar+d/2, 342, 440
		c.RBowlRound, c.BLowerRound, c.BUpperRound = c.BowlRound, c.BowlRound, c.BowlRound
		c.RLegTop, c.RLegFoot, c.RLeg, c.RLegSerifLeft, c.RLegSerifRight = 70, 280, 82, 54, 54
		c.UThin, c.UCurve = 36, 200
		capDiagDown(p)
		figuresDown(p)
		symbolsDown(p)
	})
	// The punctuation and the signs of ASCII mathematics: larger and heavier, with a
	// deeper comma, rounder parentheses, longer arms of the brackets, a leaning
	// slash, a bigger asterisk and wider and heavier maths signs; and the opposite.
	add("punct-up", func(p *design.Params) {
		q := &p.Punct
		q.DotLow, q.ColonOver, q.ExclamTopWidth, q.ExclamBottomWidth, q.ExclamGap = 12, 8, 96, 40, 90
		q.CommaDepth, q.CommaRun, q.CommaEndAngle, q.CommaInnerAt = 180, 62, 52, 84
		q.QuoteTop, q.QuotePitch, q.StraightHeight, q.StraightTopWidth, q.StraightBottomWidth = 706, 200, 280, 66, 36
		q.HyphenLength, q.HyphenThick, q.DashThick, q.UnderThick = 280, 72, 54, 54
		q.BracketOver, q.BracketUnder, q.ParenSagitta, q.ParenThick, q.ParenEnd, q.ParenAngle = 70, 30, 190, 84, 54, 52
		q.BracketStem, q.BracketArm, q.BracketReach = 74, 44, 124
		q.SlashAngle, q.SlashThick, q.BarThick = 22, 58, 62
		q.AsteriskSize, q.AsteriskInner, q.AsteriskOuter = 440, 32, 78
		q.MathWidth, q.MathThick, q.EqualsGap, q.MathAngle, q.MathAngleWidth, q.MathVertexFlat = 600, 60, 150, 27, 560, 44
		q.NotdefThick = 44
	})
	add("punct-down", func(p *design.Params) {
		q := &p.Punct
		q.DotLow, q.ColonOver, q.ExclamTopWidth, q.ExclamBottomWidth, q.ExclamGap = 6, 0, 76, 30, 70
		q.CommaDepth, q.CommaRun, q.CommaEndAngle, q.CommaInnerAt = 156, 50, 46, 76
		q.QuoteTop, q.QuotePitch, q.StraightHeight, q.StraightTopWidth, q.StraightBottomWidth = 690, 180, 240, 50, 26
		q.HyphenLength, q.HyphenThick, q.DashThick, q.UnderThick = 240, 56, 40, 40
		q.BracketOver, q.BracketUnder, q.ParenSagitta, q.ParenThick, q.ParenEnd, q.ParenAngle = 44, 12, 160, 68, 40, 44
		q.BracketStem, q.BracketArm, q.BracketReach = 60, 34, 100
		q.SlashAngle, q.SlashThick, q.BarThick = 18, 46, 50
		q.AsteriskSize, q.AsteriskInner, q.AsteriskOuter = 380, 26, 64
		q.MathWidth, q.MathThick, q.EqualsGap, q.MathAngle, q.MathAngleWidth, q.MathVertexFlat = 520, 48, 120, 23, 470, 30
		q.NotdefThick = 32
	})
	return sets
})
