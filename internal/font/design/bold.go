// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import "math"

// BoldFactors holds the factors from which the parameters of Bedern Bold are
// derived from those of Bedern Regular.
type BoldFactors struct {
	// Stem is the factor of the straight heavy strokes of the lowercase and
	// the figures, the thick diagonals and the heavy strokes of the signs.
	Stem float64
	// CapStem is the factor of the stems and the thick diagonals of the
	// capitals.
	CapStem float64
	// Round is the factor of the thickest parts of the round strokes.
	Round float64
	// Thin is the factor of the thin strokes: the thinnest parts of the
	// rounds, the hairlines, the bars, the joins and the tips of the serifs.
	Thin float64
	// Ball is the factor of the balls of the terminals and the heads of the
	// comma and the accents.
	Ball float64
	// Dot is the factor of the dot of i and j and of the full stop.
	Dot float64
	// Jut is the factor of the length of the serifs beyond the stem.
	Jut float64
	// Out is the share of the weight added to the straight strokes which a
	// horizontal line crosses in a letter that widens the letter; the rest
	// goes into its counters.  RoundOut is the same share for the round
	// strokes.
	Out, RoundOut float64
	// Side is the factor of the side spaces.
	Side float64
	// Sign is the factor of the strokes of the hyphen, the dashes, the
	// mathematical signs and the rules.
	Sign float64
	// ArchJoin, BowlJoinTop and BowlJoinBottom are the factors of the joins
	// of a thin stroke to a stem: the arches of n, m, h, r and the turn of
	// u, and the upper and the lower stroke of the bowls of a, b, d, p and q.
	// The vertical gap between the notch and the join and the thickness
	// across the stroke where it leaves the stem grow by the factor.
	ArchJoin, BowlJoinTop, BowlJoinBottom float64
	// Small is the factor of the heavy strokes round the small counters: the
	// bowl of a, the bowl of g and the upper bowl of the figure 8.
	Small float64
	// Mark is the factor of the widths and heights of the accents and of the
	// pitch of the diaeresis.
	Mark float64
	// TildeWidth and TildeHeight are the factors of the width and the height
	// of the tilde, and TildeThick that of its thick middle.
	TildeWidth, TildeHeight, TildeThick float64
	// EBar is the factor of the bar of e, and Eye the height of the counter
	// above the bar of e over that of the regular; the bar moves down to give
	// the counter this height.
	EBar, Eye float64
}

// BoldDefault returns the factors of Bedern Bold.
func BoldDefault() BoldFactors {
	return BoldFactors{
		Stem:     1.60,
		CapStem:  1.53,
		Round:    1.53,
		Thin:     1.40,
		Ball:     1.22,
		Dot:      1.35,
		Jut:      0.90,
		Out:      0.6,
		RoundOut: 0.40,
		Side:     1.00,
		Sign:     1.35,

		ArchJoin:       1.21,
		BowlJoinTop:    1.20,
		BowlJoinBottom: 1.09,
		Small:          1.40,
		EBar:           1.27,
		Mark:           1.08,
		TildeWidth:     1.11,
		TildeHeight:    1.20,
		TildeThick:     1.24,
		Eye:            0.97,
	}
}

// Bold returns the parameters of Bedern Bold.
func Bold() Params { return BoldFrom(BoldDefault()) }

// BoldFrom returns the parameters derived from the regular parameters by the
// factors r, with the fitted side spaces of the bold.  Every length is rounded
// to whole units.
func BoldFrom(r BoldFactors) Params {
	p := boldDerive(Regular(), r)
	boldRefit(&p)
	return p
}

// boldRefit moves the side spaces which the fit of the bold sets apart from
// the derived values.
func boldRefit(p *Params) {
	c, d := &p.Caps, &p.CapDiag
	for _, f := range []struct {
		v *float64
		d float64
	}{
		{&p.ASideLeft, 9}, {&p.CSideLeft, 5}, {&p.CSideRight, 11}, {&p.GSideLeft, -5},
		{&p.SSideLeft, -3}, {&p.SSideRight, 6},
		{&c.CapBowls.BSideRight, -4}, {&c.SideArmE, 4}, {&d.SideM.Left, 4},
		{&d.SideN.Left, 6}, {&d.SideN.Right, 4}, {&d.SideZ.Right, 16},
		{&d.SideA.Right, -10}, {&d.SideX.Right, -10}, {&c.CGSJQ.GSideRight, 8},
		{&c.CGSJQ.SSideRight, 4}, {&c.SideArmL, 4},
	} {
		*f.v += f.d
	}
}

// boldMul multiplies the lengths by k and rounds them to whole units.
func boldMul(k float64, fs ...*float64) {
	for _, f := range fs {
		*f = math.Round(*f * k)
	}
}

// boldAdd adds d to the lengths and rounds them to whole units.
func boldAdd(d float64, fs ...*float64) {
	for _, f := range fs {
		*f = math.Round(*f + d)
	}
}

// boldGrow returns the weight which a stroke of width w gains by the factor
// k.
func boldGrow(w, k float64) float64 { return w * (k - 1) }

// boldDerive applies the factors r to the regular parameters reg.
func boldDerive(reg Params, r BoldFactors) Params {
	p := reg
	mul, add, grow, whole := boldMul, boldAdd, boldGrow, math.Round
	out, in := r.Out, 1-r.Out

	// The added weight of the main strokes.
	dS := grow(reg.Stem, r.Stem)
	dR := grow(reg.CurveLower, r.Round)
	dH := grow(reg.Hair, r.Thin)
	dBall := grow(reg.BallDiameter, r.Ball)

	// Lowercase weights.
	mul(r.Stem, &p.Stem, &p.AscenderStem, &p.XThick, &p.ZDiag, &p.KLeg)
	mul(r.Small, &p.ABowl.Thick, &p.GBowlThick)
	mul(r.Round, &p.CurveLower, &p.SSpineThick, &p.SSideThick,
		&p.GLoopLeftThick, &p.GLoopRightThick)
	mul(r.Thin, &p.ThinLower, &p.Hair,
		&p.ETerminalTip, &p.ATailTipThick, &p.TTipThick, &p.SArcThick, &p.SUpperArcThick,
		&p.SBeakArc, &p.SSerifThick, &p.GLinkThick, &p.GLoopTopThick,
		&p.ZUpperBeakThick, &p.ZLowerBeakThick, &p.SerifThick, &p.DiagVertexFlat)
	mul(r.Ball, &p.BallDiameter, &p.GEarBall)
	// The stalk of the ear of g leaves the bowl as the lower stroke of a bowl
	// leaves the stem: it grows by the factor of that join.
	mul(r.BowlJoinBottom, &p.GEarThick)
	// The tip of the tail of t and the radii of its corners grow with the
	// thin strokes.
	p.TTipScale = r.Thin
	// The edges of the hook of f leave the stem lower by the growth of the
	// ball, and the reach of the hook loses half of it.
	p.FInnerDrop = whole(reg.FInnerDrop + dBall)
	p.FOuterDrop = whole(reg.FOuterDrop + dBall)
	p.FReach = whole(reg.FReach - dBall/2)
	// The counter above the bar of e has Eye of its regular height, and the
	// bar of e takes the rest of the added weight of the thin top and of the
	// bar.
	eyeTop := func(q Params) float64 { return q.XHeight + q.OvershootLower - q.ThinLower }
	eye := eyeTop(reg) - reg.EBarHeight - reg.EBarThick/2
	mul(r.EBar, &p.EBarThick)
	p.EBarHeight = whole(eyeTop(p) - r.Eye*eye - p.EBarThick/2)
	// An underside of the bar of e less than 15 units above the left extreme
	// of the bowl lies on the extreme.
	p.EBarSnap = 15
	// The stalk of the ball terminal grows by the mean of the factors of the
	// balls and of the thin strokes.
	mul((r.Ball+r.Thin)/2, &p.BallStalk)
	// The dot is rounded to an even number of units.
	p.DotDiameter = 2 * whole(reg.DotDiameter*r.Dot/2)
	// The distance from the centre of the ball of the ear of g to the crest
	// of the ear grows with the ball, and the reach of the ear grows by the
	// growth of the ball.
	p.GEarDrop = whole((reg.GEarDrop+reg.GEarRise)*r.Ball - reg.GEarRise)
	p.GEarReach = whole(reg.GEarReach + grow(reg.GEarBall, r.Ball))
	mul(r.Jut, &p.SerifJut, &p.BracketReach, &p.SerifJutInner, &p.FFootRight, &p.RFootRight,
		&p.DiagOverThick, &p.DiagOverThin)

	// Where a thin stroke meets a stem, the gap between the notch and the
	// join and the thickness across the stroke grow by the factors of the
	// joins.  The bowls share one thickness at the join, which grows by the
	// factor of the lower stroke.
	join := func(notch, j, k float64) float64 { return whole(notch + (j-notch)*k) }
	mul(r.ArchJoin, &p.ArchThinJoin)
	mul(r.BowlJoinBottom, &p.BowlJoinThick)
	p.ArchJoin = join(reg.ArchNotch, reg.ArchJoin, r.ArchJoin)
	p.RJoin = join(reg.RNotch, reg.RJoin, r.ArchJoin)
	p.UJoin = join(reg.UNotch, reg.UJoin, r.ArchJoin)
	for _, b := range []*Bowl{&p.DBowl, &p.BBowl, &p.PBowl, &p.QBowl, &p.ABowl} {
		if b.NotchUpper != 0 {
			b.JoinUpper = join(b.NotchUpper, b.JoinUpper, r.BowlJoinTop)
		}
		b.JoinLower = join(b.NotchLower, b.JoinLower, r.BowlJoinBottom)
	}

	// The joins of the bowls of b, d, p and q to their extremes and the upper
	// edge of the terminals of c and e take a share of the curvature of the
	// curve which continues them, and the tip and the low point of the tail
	// of a take rounder corners by the factor of the thin strokes.
	p.BowlMatch = 0.4
	p.TerminalMatch = 0.5
	p.ATailScale = r.Thin

	// Widths: a letter widens by Out of the weight added to the strokes which
	// a horizontal line crosses in it, and its counters take the rest.
	add(-2*in*dS, &p.NCounter, &p.MCounter, &p.UCounter)
	rout := r.RoundOut
	add(2*rout*dR, &p.OBowlWidth, &p.EBowlWidth, &p.CBowlWidth, &p.GBowlWidth, &p.GLoopWidth,
		&p.SLowerWidth, &p.SUpperWidth)
	// A bowl on a stem: the free extreme moves out by RoundOut of the round
	// stroke, the inner edge of the stem in by the rest of the stem.
	for _, b := range []*Bowl{&p.DBowl, &p.BBowl, &p.PBowl, &p.QBowl, &p.ABowl} {
		b.Width = whole(b.Width + rout*dR - in*dS)
	}
	// The arm of r reaches further by the growth of the ball.
	add(dBall, &p.RReach)
	// The ball of a grows about the centre of the regular ball.
	add(-dBall/2, &p.AArcReach)
	// The ink of the diagonal letters.
	add(out*(dS+dH), &p.VInk, &p.XInk)
	// The ink of w is 2 narrower than the sum of these growths: at this width
	// the rounding of the bracket of the head serif leaves no kink.
	add(out*(2*dS+2*dH)-2, &p.WInk)
	add(out*grow(reg.ZDiag, r.Stem), &p.ZInk, &p.ZInkTop)
	add(out*(grow(reg.KLeg, r.Stem)+dH), &p.KReach)

	// Side spaces.
	mul(r.Side, &p.SideStemLeft, &p.SideStemRight, &p.SideRoundLeft, &p.SideRoundRight,
		&p.DSideRight, &p.SideOpenRight, &p.BPSideLeft, &p.DQSideLeft, &p.BPSideRight,
		&p.CSideLeft, &p.CSideRight, &p.QSideRight, &p.RSideLeft, &p.RSideRight,
		&p.ASideLeft, &p.SSideLeft, &p.SSideRight, &p.GSideLeft, &p.GSideRight,
		&p.DiagSideLeft, &p.DiagSideRight, &p.ZSideLeft, &p.ZSideRight, &p.KSideRight)
	// The tail of j deepens by the weight added to its thin bottom; the tail
	// of y keeps its depth.
	add(grow(reg.ThinLower, r.Thin), &p.JDepth)
	// The loop of g deepens by the weight added to its top stroke.
	add(grow(reg.GLoopTopThick, r.Thin), &p.GLoopDepth)
	// The word space is that of the regular.
	p.SpaceAdvance = reg.SpaceAdvance
	// The ink of ß widens by Out of the stem and by RoundOut of the two sides
	// of the bowl, and EszettInk, its ink over that of n, follows.
	nInk := func(q Params) float64 {
		return NAdvance(q) - q.SideStemLeft - q.SideStemRight + 2*q.SerifJut
	}
	p.EszettInk = (reg.EszettInk*nInk(reg) + out*dS + 2*rout*dR) / nInk(p)

	boldCaps(&p, reg, r)
	boldFigures(&p, reg, r)
	boldSigns(&p, reg, r)
	boldDiagSerifs(&p, reg, r)
	return p
}

// boldDiagSerifs sets the overhangs of the serifs of x, k and the leg of R:
// the regular overhang times the jut factor, but with a flat of at least
// DiagBracketFlatMin.  On x and k the flat is the overhang less the
// thickness of the tip times the lean of the edge; on the leg of R it lies
// between the tip and the reach of the bracket.
func boldDiagSerifs(p *Params, reg Params, r BoldFactors) {
	for _, s := range []struct {
		v          *float64
		over, lean float64
	}{
		{&p.XOverThick, reg.XOverThick, XLeanThick(*p)},
		{&p.XOverThin, reg.XOverThin, XLeanThin(*p)},
		{&p.KOverArm, reg.KOverArm, KArmLean(*p)},
		{&p.KOverLeg, reg.KOverLeg, KLegLean(*p)},
	} {
		*s.v = math.Round(math.Max(r.Jut*s.over, DiagBracketFlatMin+p.SerifThick*math.Abs(s.lean)))
	}
	br := CapDiagBracket(*p)
	b, rb := &p.Caps.CapBowls, reg.Caps.CapBowls
	for _, s := range []struct {
		v    *float64
		over float64
	}{{&b.RLegSerifLeft, rb.RLegSerifLeft}, {&b.RLegSerifRight, rb.RLegSerifRight}} {
		*s.v = math.Round(math.Max(r.Jut*s.over, DiagBracketFlatMin+math.Max(br.Reach, br.ReachAcute)))
	}
}

// boldCaps applies the factors to the capitals.
func boldCaps(p *Params, reg Params, r BoldFactors) {
	mul, add, grow := boldMul, boldAdd, boldGrow
	out, in := r.Out, 1-r.Out
	c, rc := &p.Caps, reg.Caps
	g, b, d := &c.CGSJQ, &c.CapBowls, &p.CapDiag
	rd := reg.CapDiag

	dS := grow(rc.Stem, r.CapStem)
	dR := grow(rc.Round, r.Round)
	dThick := grow(rd.Thick, r.CapStem)
	dThin := grow(rd.Thin, r.Thin)

	mul(r.CapStem, &c.Stem, &d.Thick, &d.WThick, &d.XThick, &d.MThick, &d.NThick, &d.KLeg,
		&d.ZDiag, &d.MStem, &b.RLeg)
	mul(r.Round, &c.Round, &b.BowlRound, &b.BUpperRound, &b.BLowerRound, &b.RBowlRound,
		&g.SSpineThick, &g.SSideThick)
	mul(r.Thin, &c.RoundThin, &c.Bar, &c.TBar, &c.SerifThick, &c.ArmSerifThick,
		&g.BeakArc, &g.CTailCut, &g.SArcThick, &g.SUpperArcThick, &g.QTailTip,
		&b.UThin, &d.Thin, &d.ThinStem, &d.VertexFlat, &d.ApexFlat)
	mul(r.Ball, &g.JBall)
	// The tail of C, where it is thickest between its cut and the counter,
	// grows by the factor of the signs.
	mul(r.Sign, &g.CTailArc)
	mul(r.Jut, &c.SerifJut, &c.BracketReach, &g.GSerifLeft, &g.GSerifRight, &g.JHeadRight,
		&d.OverThick, &d.OverThin, &d.KOverLeg, &d.KOverArm)
	// The bar of Ł reaches as far beyond the stem on the right as in the
	// regular and starts at the tip of the left foot serif.
	c.LslashWidth = math.Round(CapLslashStems*rc.Stem + dS + c.SerifJut - rc.SerifJut)

	// Counters and widths.
	add(-2*in*dS, &c.HCounter)
	rout := r.RoundOut
	add(2*rout*dR, &c.OWidth, &g.SLowerWidth, &g.SUpperWidth)
	add(2*rout*dR, &g.BeakX)
	// An arm from the stem: the letter widens by Out of the stem.
	add(-in*dS, &c.ETopArm, &c.EMiddleArm, &c.EBottomArm, &c.FTopArm, &c.FMiddleArm, &c.LArm)
	add(-in*dS/2, &c.TArm)
	// A bowl on a stem: the letter widens by Out of the stem and RoundOut of
	// the round.
	add(out*dS+rout*dR-dS, &b.BUpper, &b.BLower, &b.PWidth, &b.RWidth, &b.DWidth)
	add(out*dS, &g.JReach)
	// The ink of the diagonal capitals.
	add(out*(dThick+dThin), &d.AInk, &d.VInk, &d.XInk, &d.YInk)
	add(out*(2*dThick+2*dThin), &d.WInk)
	add(out*grow(rd.ZDiag, r.CapStem), &d.ZInk, &d.ZInkTop)
	add(out*(dS+grow(rd.KLeg, r.CapStem)), &d.KInk)
	add(out*(2*grow(rd.ThinStem, r.Thin)+grow(rd.MThick, r.CapStem)+dThin), &d.MInk)

	// Side spaces.
	mul(r.Side, &c.SideStemLeft, &c.SideStemRight, &c.SideRoundLeft, &c.SideRoundRight,
		&c.SideArmE, &c.SideArmF, &c.SideArmL, &c.SideT,
		&g.CSideRight, &g.GSideRight, &g.SSideLeft, &g.SSideRight, &g.JSideLeft,
		&b.BowlSideLeft, &b.BowlSideRight, &b.BSideRight, &b.DSideRight, &b.RSideRight)
	for _, s := range []*CapSides{&d.SideA, &d.SideV, &d.SideW, &d.SideX, &d.SideY,
		&d.SideZ, &d.SideK, &d.SideM, &d.SideN} {
		mul(r.Side, &s.Left, &s.Right)
	}
}

// boldFigures applies the factors to the figures, which share one advance
// and widen as the round letters do.
func boldFigures(p *Params, reg Params, r BoldFactors) {
	mul, add := boldMul, boldAdd
	f := &p.Figures
	dR := reg.Figures.Round * (r.Round - 1)
	// The spine of 8 grows with the stem.
	mul(r.Stem, &f.Stem, &f.EightGap)
	// The heavy horizontal strokes (the base of 2, the bars of 5 and 7) grow
	// as the round strokes; the serifs at the ends of the base of 2 and of the
	// bar of 7 reach beyond the bar by half the added weight more.
	mul(r.Round, &f.Heavy)
	dHeavy := f.Heavy - reg.Figures.Heavy
	add(dHeavy/2, &f.TwoRise, &f.SevenDrop)
	// The foot of 7 grows with its diagonal.
	mul(r.Round, &f.Round, &f.Bowl, &f.ThreeUpperBowl, &f.TwoNeckThick, &f.SixStemThick,
		&f.SevenTopThick, &f.SevenFoot, &f.EightLL, &f.EightLR, &f.EightUL)
	// The thin side of the upper counter of 8 grows as the small bowls.
	mul(r.Small, &f.EightUR)
	mul(r.Thin, &f.Thin, &f.OneFlagThick, &f.TwoSerifThick, &f.FiveStroke, &f.SixBowlTopThin,
		&f.SixTopThin, &f.SevenSerifThick, &f.FourDiag, &f.FourBar, &f.EightJoin,
	)
	// The stalk of the ball of 6 and 9 grows as the stalk of the ball
	// terminal of the lowercase, and the distance of the tip of its wedge
	// from the outer edge of the stroke grows with the ball.
	mul((r.Ball+r.Thin)/2, &f.CurlStalk)
	mul(r.Ball, &f.Ball, &f.CurlNeck)
	mul(r.Jut, &f.OneFoot, &f.FourFoot)
	for i := range f.Ink {
		if f.Ink[i] != 0 {
			f.Ink[i] = math.Round(f.Ink[i] + 2*r.RoundOut*dR)
		}
	}
	f.Advance = math.Round(reg.Figures.Advance + 2*r.RoundOut*dR)
	// The upper bowls of 3 and 8 widen as the ink does; the upper bowl of 3
	// keeps its left edge.
	add(2*r.RoundOut*dR, &f.ThreeUpperWidth, &f.ThreeUpperRight, &f.EightUpperWidth)
	// The centres of the balls of 3, 6 and 9 move away from the outer edges
	// by the growth of their radius, and the arc of 6 widens as the ink does.
	dBall := reg.Figures.Ball * (r.Ball - 1)
	add(dBall/2, &f.CurlHeight, &f.SixBallIn, &f.ThreeLowerLeft)
	f.SixArcWidth = math.Round(reg.Figures.SixArcWidth + 2*r.RoundOut*dR)
	// The left end of the base of 2 rises by the weight added to the base.
	add(dHeavy, &f.TwoBaseEnd)
	// The strokes of the superior figures grow by the factor of the signs
	// over that of the stems.
	f.SupWeight = reg.Figures.SupWeight * r.Sign / r.Stem

	// The old-style figures 0, 1 and 2 follow the rules of the lining ones,
	// and the side spaces of the proportional figures those of the letters.
	o := &f.Old
	mul(r.Round, &o.ZeroRound)
	mul(r.Thin, &o.ZeroThin)
	add(2*r.RoundOut*dR, &o.TwoInk, &o.TwoBase)
	mul(r.Jut, &o.OneFoot)
	mul(r.Ball, &o.TwoBall)
	heavy := o.TwoHeavy
	mul(r.Round, &o.TwoHeavy)
	add((o.TwoHeavy-heavy)/2, &o.TwoRise)
	add(o.TwoHeavy-heavy, &o.TwoBaseEnd)
	// The diagonal of the old-style 2 thickens with the thin strokes.
	add((f.Thin-reg.Figures.Thin)/math.Cos(o.TwoDiagAngle*math.Pi/180), &o.TwoDiagFoot)
	for i := range o.Left {
		mul(r.Side, &o.Left[i], &o.Right[i], &o.LiningLeft[i], &o.LiningRight[i])
	}
}

// boldSigns applies the factors to the punctuation, the symbols and the
// accents.
func boldSigns(p *Params, reg Params, r BoldFactors) {
	mul, add, grow := boldMul, boldAdd, boldGrow
	q, s, a := &p.Punct, &p.Sym, &p.Acc
	// The signs widen as the letters do.
	dR := grow(reg.CurveLower, r.Round)
	dS := grow(reg.Stem, r.Stem)
	add(r.Out*grow(reg.BallDiameter, r.Ball)+r.RoundOut*0.9*dR, &s.QInk)
	add(r.RoundOut*2*dR, &s.PctInk, &s.AtInk, &s.AmpInk, &s.DegDiam)
	add(r.Out*2*dS, &s.NumInk)
	add(r.Out*dS, &s.SterInk)
	mul(r.Stem, &q.ExclamTopWidth, &q.BracketStem, &q.SlashThick, &q.BarThick,
		&s.SterStem, &s.PctSlash, &s.QStem)
	mul(r.Round, &q.ParenThick, &s.BraceArm, &s.BraceTip, &s.BraceEnd,
		&s.AmpBowlThick, &s.SectHookThick)
	// The ring of the circled signs is 40 units times the factor of the thin
	// strokes, so that it keeps its ratio to the thin of O; the ring of the
	// large circle is 44 units times the factor of the stems, 0.55 of the
	// stem of n.
	s.CircRing = math.Round(40 * r.Thin)
	s.LargeRing = math.Round(44 * r.Stem)
	mul(r.Small, &s.PctZeroThick)
	mul(r.Stem, &s.AtRing)
	mul(r.Sign, &s.DollarSpine, &s.DollarSideThick,
		&s.EuroThick, &s.AtBowlThick, &s.SterBase, &s.AtStem)
	// The spines of the ampersand and the section sign grow by half the
	// added weight of the round strokes.
	mul((1+r.Round)/2, &s.AmpThick, &s.SectSpine)
	// The loop and the base of the sterling follow the edges of its stem
	// from the stem and the frame of the regular.
	s.SterStemRef, s.SterInkRef = reg.Sym.SterStem, reg.Sym.SterInk
	mul(r.Thin, &q.ExclamBottomWidth, &q.CommaTip, &q.ParenEnd, &q.BracketArm,
		&q.StraightBottomWidth, &q.PrimeBottomWidth, &q.AsteriskInner, &q.NotdefThick,
		&s.DegThick, &s.TildeThick, &s.TildeThin, &s.CircThick, &s.GraveTip,
		&s.PctZeroThin, &s.NumStroke, &s.NumBar, &s.DollarArc,
		&s.DollarUpperArc, &s.DollarTip, &s.DollarStroke, &s.SterBar, &s.EuroBar,
		&s.AmpTail, &s.AmpArmThick,
		&a.CircThin, &a.CircThick, &a.TildeThin, &a.MacronThick,
		&a.BreveThick, &a.BreveThin, &a.RingThick, &a.RingThin, &a.CedThick,
		&a.OgoNeck, &a.OgoThick, &a.OgoThin, &a.OgoTip, &a.AcuteTip, &a.DblTip)
	// The apostrophe after d, l and t may reach beyond the advance by the
	// growth of its head as well.
	a.Overhang = math.Round(reg.Acc.Overhang + reg.Acc.ApoScale*grow(reg.Punct.CommaHead, r.Ball))
	// The top of the straight marks grows with the dot, so that it is 0.7 of
	// the stem of n, as it is 0.8 in the regular.
	mul(r.Dot, &q.StraightTopWidth)
	mul(r.Ball, &q.CommaHead, &q.PrimeTopWidth, &q.AsteriskOuter,
		&s.GraveHead, &a.AcuteHead, &a.DblHead)
	mul(r.Sign, &q.HyphenThick, &q.DashThick, &q.UnderThick, &q.MathThick)
	// The accents and the hooks of the cedilla and the ogonek grow by the
	// factor of the marks; their gaps above the letters and the cut at the
	// end of the tail of the cedilla stay.
	mul(r.Mark, &a.AcuteWidth, &a.AcuteHeight, &a.AcuteHeightCap, &a.CircWidth, &a.CircHeight,
		&a.CircHeightCap, &a.TildeHeight, &a.TildeHeightCap, &a.DiaPitch,
		&a.DiaPitchNarrow, &a.MacronWidth, &a.MacronWidthCap, &a.BreveWidth, &a.BreveHeight,
		&a.BreveHeightCap, &a.RingDiam, &a.RingDiamCap, &a.DblWidth, &a.DblHeight,
		&a.DblHeightCap, &a.DblPitch,
		&a.CedReach, &a.CedTail, &a.CedDepth, &a.OgoReach, &a.OgoDepth, &a.OgoTail)
	// The dot below grows about its centre.
	a.DotBelowGap = math.Round(reg.Acc.DotBelowGap - a.DotAboveScale*(p.DotDiameter-reg.DotDiameter)/2)
	// The reach of the hook of the ogonek grows by the weight added to its
	// thick side as well.
	add(a.OgoThick-reg.Acc.OgoThick, &a.OgoReach)
	// The tilde, the mark and the sign alike, grows by its own factors; the
	// sign keeps its width.
	a.TildeWidth = math.Round(reg.Acc.TildeWidth * r.TildeWidth)
	a.TildeThick = math.Round(reg.Acc.TildeThick * r.TildeThick)
	mul(r.TildeHeight/r.Mark, &a.TildeHeight, &a.TildeHeightCap)
	mul(r.TildeHeight, &s.TildeHeight)
}
