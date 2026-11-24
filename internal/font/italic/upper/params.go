// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

// Package upper makes the capitals, the figures and the signs of Bedern
// Italic from the glyph code of the roman: each glyph is drawn upright by the
// roman code with the weights of the italic, its rounds are turned to the
// stress of the italic pen, it is narrowed, its stems are put on the lattice
// of the slant, it is sheared by the slant 1/4 and moved back by ShearHeight
// times the slant, and its curves get nodes at the extremes of the slanted
// form.
package upper

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
)

// Params are the parameters of the capitals, figures and signs of the italic.
// Lengths are in font units, in the upright description unless a comment says
// otherwise.
type Params struct {
	// The values which the italic shares with its lowercase.
	SlantNum, SlantDen int     // the slant, the tangent SlantNum/SlantDen
	ShearHeight        float64 // the height at which the shear leaves a glyph in place

	// Turn is the angle, in degrees, by which the inner edges of the rounds
	// turn anticlockwise about their centres before the shear, so that the
	// line through the thinnest places of the top and the bottom of O, by the
	// least vertical thickness, leans 5 degrees less than the stems.
	Turn float64

	// Weight is the weight of the italic over that of the roman, at right
	// angles to the strokes: it sets the stems (given horizontally, which the
	// shear keeps), the rounds, the thin parts, the bars and the serif
	// tips of the capitals, the figures and the signs.
	Weight float64

	// Width is the ink width of the italic over that of the roman: the glyphs
	// are narrowed by this factor before the shear, with their side spaces.
	Width float64

	// CapStem and FigStem are the stems of the capitals and of the figures,
	// horizontally in the italic, whole units: both edges of a stem
	// lie on the lattice of the slant.
	CapStem, FigStem float64

	// DiagThick, DiagThin, WThick, NThick, MThick and KLeg are the factors by
	// which the diagonals of the capitals are given to the roman code over the
	// roman's weight times Weight, so that after the narrowing and the shear,
	// which make the thick diagonals steeper and the thin ones flatter, each
	// thick diagonal has the roman's weight times Weight at right angles and
	// each thin one the roman's weight: the thin diagonals, flatter than in
	// the roman, are lighter than their measure.
	DiagThick, DiagThin, WThick, NThick, MThick, KLeg float64

	// ThinFloor is the least thickness of the hairlines of the capitals, the
	// thin part of the rounds, the bars and the tips of the serifs: the script
	// forms of the maths italic have it from the pen (36 and 50 at the two
	// levels); 0 leaves them as the weight gives them.
	ThinFloor float64

	// ObtuseSerif is the length of a serif on the obtuse side of a slanted
	// stem (left at the foot, right at the head) over that of the others.
	ObtuseSerif float64

	// Lean is the extra slant of the diagonal capitals whose apex or vertex
	// the shear lets fall, by letter.
	Lean map[string]Lean
}

// Default returns the parameters of the italic.
func Default() Params {
	return Params{
		SlantNum: 1, SlantDen: 4,
		ShearHeight: 224,

		Turn: 2,

		Weight: 0.94,
		Width:  0.97,

		CapStem: 93,
		FigStem: 80,

		DiagThick: 0.94,
		DiagThin:  1.16,
		WThick:    1.00,
		NThick:    0.88,
		MThick:    0.93,
		KLeg:      0.90,

		ObtuseSerif: 0.92,

		Lean: map[string]Lean{},
	}
}

// slant returns the tangent of the slant.
func (p Params) slant() float64 { return float64(p.SlantNum) / float64(p.SlantDen) }

// shift returns the move to the left after the shear.
func (p Params) shift() float64 { return p.ShearHeight * p.slant() }

// Roman returns the parameters of the roman with which the roman code draws
// the glyphs of the italic before they are narrowed and sheared: the stems so
// that they are CapStem and FigStem wide after the narrowing, the other
// weights times Weight, those measured horizontally divided by Width, and the
// lengths of the serifs divided by Width, which the narrowing keeps.
func (p Params) Roman() design.Params {
	r := design.Regular()
	w, n := p.Weight, p.Width
	c := &r.Caps
	c.Stem = p.CapStem / n
	c.Round *= w / n
	c.RoundThin *= w
	c.Bar = roundTo(c.Bar*w, 4)
	c.TBar = roundTo(c.TBar*w, 4)
	c.SerifThick *= w
	c.SerifJut /= n
	c.BracketReach /= n
	c.ArmSerifThick *= w / n
	c.HBarHeight = 354
	c.BowlRound *= w / n
	c.BUpperRound *= w / n
	c.BLowerRound *= w / n
	c.RBowlRound *= w / n
	c.UThin *= 1.05 * w / n
	c.CGSJQ.GSerifLeft /= n
	c.CGSJQ.GSerifRight /= n
	c.CGSJQ.JHeadRight /= n
	c.RLegSerifLeft /= n
	c.RLegSerifRight /= n
	c.CGSJQ.SSpineThick *= w
	c.CGSJQ.SArcThick *= w
	c.CGSJQ.SUpperArcThick *= w
	c.CGSJQ.SSideThick *= w / n
	for _, v := range []*float64{&c.RoundThin, &c.Bar, &c.TBar, &c.UThin, &c.SerifThick,
		&c.CGSJQ.SArcThick, &c.CGSJQ.SUpperArcThick} {
		*v = math.Max(*v, p.ThinFloor)
	}

	d := &r.CapDiag
	d.Thick *= w * p.DiagThick / n
	d.WThick *= w * p.WThick / n
	d.XThick *= w * p.DiagThick / n
	d.MThick *= w * p.MThick / n
	d.NThick *= w * p.NThick / n
	d.KLeg *= w * p.KLeg / n
	d.ZDiag *= w * 0.95 / n
	d.ZInk *= 0.96
	d.ZInkTop *= 0.96
	d.KArmAngle = 40
	d.KJunction = 420
	d.Thin *= w * p.DiagThin / n
	d.ThinStem /= n // the thin stems of M and N keep the weight of the roman, as the thin diagonals do
	d.MStem = p.CapStem / n
	d.OverThick /= n
	d.OverThin /= n
	d.KOverLeg /= n
	d.KOverArm /= n

	q := &r.Punct
	q.CommaHead *= w
	q.HyphenThick *= w
	q.DashThick *= w
	// The ends of the sloped dashes are cut along the slant: the
	// white between two of them is a parallelogram, narrower than
	// the rectangle of the roman: the sides grow by 10.
	for _, d := range []*design.PunctSides{&q.Hyphen, &q.Dash, &q.EmDash} {
		d.Left += 10
		d.Right += 10
	}
	q.ParenThick *= w / n
	q.ParenEnd *= w / n
	q.BracketStem *= w / n
	q.BracketArm *= w
	q.ExclamTopWidth *= w / n
	q.ExclamBottomWidth *= w / n
	q.StraightTopWidth *= w / n
	q.StraightBottomWidth *= w / n
	q.PrimeTopWidth *= w / n
	q.PrimeBottomWidth *= w / n
	q.AsteriskInner *= w / n
	q.AsteriskOuter *= w / n

	y := &r.Sym
	y.QStem *= w / n
	y.BraceArm *= w / n
	y.BraceEnd *= w / n
	y.PctZeroThick *= w / n
	y.PctZeroThin *= w
	y.PctSlash *= w / n
	y.NumStroke *= w / n
	y.NumBar *= w
	y.DollarSpine *= w
	y.DollarArc *= w
	y.DollarUpperArc *= w
	y.DollarSideThick *= w / n
	y.DollarStroke *= w / n
	y.SterStem *= w / n
	y.SterBar *= w
	y.SterBase *= w
	y.EuroThick *= w / n
	y.EuroBar *= w
	y.SectSpine *= w
	y.SectHookThick *= w / n
	y.AtRing *= w
	y.AtBowlThick *= w / n
	y.AtStem *= w / n
	r.BallDiameter = r.Figures.Ball
	r.BallStalk *= r.Figures.Ball / design.Regular().BallDiameter

	f := &r.Figures
	f.Stem = p.FigStem / n
	f.Round *= w / n
	f.Thin *= w
	f.Heavy *= w / n
	f.Bowl *= w / n
	return r
}

// roundTo returns x rounded to a multiple of m.
func roundTo(x, m float64) float64 {
	return m * float64(int(x/m+0.5))
}

// A Lean is an extra slant of a letter, dx/dy upright, about the height At:
// a positive value moves what lies below At to the left.
type Lean struct {
	By, At float64
}
