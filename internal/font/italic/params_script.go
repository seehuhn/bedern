// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "math"

// widen returns the parameters with the horizontal positions of the letter
// widened by f, the rule of the script forms: the widths of the bowls and the
// set widths grow by f, and a counter between two stems grows such that the
// distance of the centre lines of its stems grows by f (stem0 is the stem
// before and p.Stem the stem after the change of the pen).  The turns, the
// entries and the exits, the sides and the heights keep the values of the
// parameters: the pen of a script form is heavier, and a turn which a
// heavier pen draws is not a stretched turn of a lighter one.
func (p Params) widen(f, stem0 float64) Params {
	if f == 1 {
		return p
	}
	counter := func(c float64) float64 { return f*(c+stem0) - p.Stem }
	p.NCounter = counter(p.NCounter)
	p.MCounter = counter(p.MCounter)
	p.WCounter = counter(p.WCounter)
	p.YCounter = counter(p.YCounter)
	p.KLegCounter = counter(p.KLegCounter)
	p.RCounter *= f
	for _, v := range []*float64{
		&p.OBowlWidth, &p.CBowlWidth, &p.CWidth, &p.EBowlWidth, &p.EWidth,
		&p.QWidth, &p.RWidth, &p.TWidth, &p.JWidth, &p.FWidth, &p.BWidth, &p.PWidth,
		&p.SWidth, &p.SLower, &p.SUpper, &p.VWidth, &p.XWidth, &p.ZWidth,
		&p.BowlReach, &p.CTipInk, &p.ETipInk, &p.WFree, &p.ZTopRight, &p.XTopIn, &p.XBotIn,
	} {
		*v = math.Round(*v * f)
	}
	p.Gk = p.Gk.widen(f, stem0, p.Stem)
	return p
}
