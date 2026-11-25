// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/stroke"
)

// The double-struck letters of the first tier: the capitals A to
// Z and 𝟙, each with its two script forms.  ℂ ℕ ℙ ℚ ℝ ℤ 𝔼 and 𝟙 are drawn
// here and the other capitals in dscaps.go, dsdiag.go, dsround.go, dshooks.go
// and dss.go.
//
// A double-struck letter is the roman capital (or the figure one) drawn in
// lines.  One rule says which strokes are doubled: a stroke which the roman
// draws thick and which is straight (a stem or a diagonal), or which is the
// left side of a round, is two lines with a white between them, joined at its
// ends; every other stroke (the thin strokes of the roman, the right sides of
// the rounds and the bowls, the arms, the bars, the legs and the tail) is one
// line.  The left side of the round of C, G, O and Q is a second line inside
// the round, which leaves the round above and below and runs a white lens with
// it.  The serifs of the stems and the beaks of the arms are those of the roman
// capitals.  The positions are those of the roman letter, read from its
// outline, with the letter widened by what the doubled stroke adds over the
// roman stroke.
//
// The script forms are the construction run with heavier lines, a wider
// white and wider counters.
const (
	// dsGap is the white between the two lines of a doubled stroke at text
	// size.
	dsGap = 100.0
	// dsLine is the line over the thin stem of N: 49 units.
	dsLine = 1.07
	// dsLensJoin is the position angle (degrees from the right, on the
	// centre line of the round) at which the inner line of ℂ and ℚ leaves
	// the round above, and below by symmetry.
	dsLensJoin = 118.0
	// dsCOpen is the least position angle, in degrees from the right, of
	// the ends of the arms of ℂ.
	dsCOpen = 35.0
	// dsQTail is the depth of the tail of ℚ below the baseline over the cap
	// height, dsQTailAt the position angle on the round (degrees from the
	// right) at which it crosses the round, dsQTailLean its lean below the
	// level in degrees and dsQTailIn how far it starts inside the round, in
	// lines.
	dsQTail     = 0.22
	dsQTailAt   = 305.0
	dsQTailLean = 48.0
	dsQTailIn   = 0.0
)

// dsLevels are the weights and widths of the text form and the two script
// forms: the line over that of the text form, and the white and the
// counters over theirs.
var dsLevels = [3]struct{ weight, wide float64 }{{1, 1}, {1.13, 1.10}, {1.24, 1.22}}

// dsWiden holds the letters whose diagonals meet and which are drawn wider than
// their roman by this factor, which keeps the white between the doubled stroke
// and the line beside it open at the vertices (V, W, M and A keep the outer
// edges of their roman and let the doubled stroke grow inwards).
var dsWiden = map[string]float64{"A": dsDiagWiden, "V": dsDiagWiden}

// dsDiagWiden is the factor by which A and V are wider than their roman.
const dsDiagWiden = 1.08

// dsLetters are the double-struck letters: the name of the glyph, its code
// point, the letter of \mathbb, the roman glyph it is drawn from and its
// construction.
var dsLetters = []struct {
	name  string
	r     rune
	base  rune // the character which \mathbb maps to it
	roman string
	draw  func(d dsFrame) outline.Path
}{
	{"u1D538", 0x1D538, 'A', "A", dsA},
	{"u1D539", 0x1D539, 'B', "B", dsB},
	{"uni2102", 0x2102, 'C', "C", dsC},
	{"u1D53B", 0x1D53B, 'D', "D", dsD},
	{"u1D53C", 0x1D53C, 'E', "E", dsE},
	{"u1D53D", 0x1D53D, 'F', "F", dsF},
	{"u1D53E", 0x1D53E, 'G', "G", dsG},
	{"uni210D", 0x210D, 'H', "H", dsH},
	{"u1D540", 0x1D540, 'I', "I", dsI},
	{"u1D541", 0x1D541, 'J', "J", dsJ},
	{"u1D542", 0x1D542, 'K', "K", dsK},
	{"u1D543", 0x1D543, 'L', "L", dsL},
	{"u1D544", 0x1D544, 'M', "M", dsM},
	{"uni2115", 0x2115, 'N', "N", dsN},
	{"u1D546", 0x1D546, 'O', "O", dsO},
	{"uni2119", 0x2119, 'P', "P", dsP},
	{"uni211A", 0x211A, 'Q', "Q", dsQ},
	{"uni211D", 0x211D, 'R', "R", dsR},
	{"u1D54A", 0x1D54A, 'S', "S", dsS},
	{"u1D54B", 0x1D54B, 'T', "T", dsT},
	{"u1D54C", 0x1D54C, 'U', "U", dsU},
	{"u1D54D", 0x1D54D, 'V', "V", dsV},
	{"u1D54E", 0x1D54E, 'W', "W", dsW},
	{"u1D54F", 0x1D54F, 'X', "X", dsX},
	{"u1D550", 0x1D550, 'Y', "Y", dsY},
	{"uni2124", 0x2124, 'Z', "Z", dsZ},
	{"u1D7D9", 0x1D7D9, '1', "one", dsOne},
}

// A dsFrame holds what a double-struck letter is drawn from: the
// parameters, with the serifs grown by the weight of the level, the line,
// the white, the width of a doubled stroke, the roman outline and the
// factor of the widths.
type dsFrame struct {
	p       design.Params
	q       design.Params // the serifs of the level
	t, g, d float64       // the line, the white, the doubled stroke (2 t + g)
	roman   outline.Path
	adv     float64 // the advance of the roman
	wide    float64
	extra   float64 // the width added to the letter right of its doubled stroke
}

// newDSFrameOf returns the frame of a letter drawn from the roman outline
// path with the advance adv.
func newDSFrameOf(p design.Params, path outline.Path, adv float64, level int) dsFrame {
	lv := dsLevels[level]
	t := math.Round(dsLine * p.CapDiag.ThinStem * lv.weight)
	g := math.Round(dsGap * lv.wide)
	q := p
	q.Caps.SerifThick *= lv.weight
	d := dsFrame{p: p, q: q, t: t, g: g, d: 2*t + g, roman: path, adv: adv, wide: lv.wide}
	d.extra = d.d - p.Caps.Stem
	return d
}

// roundExtra returns the width which the doubled side adds to a round: the
// doubled stroke over the thickest part of the round of the roman.
func (d dsFrame) roundExtra() float64 { return d.d - d.p.Caps.Round }

// diagExtra returns the width which a doubled diagonal adds to N and Z: the
// doubled stroke over the diagonal of the roman, measured along a level
// line at the lean of the diagonal of N.
func (d dsFrame) diagExtra(thick float64) float64 {
	a := math.Atan2(d.p.CapHeight, design.CapNStemRight(d.p)-design.CapNStemLeft(d.p))
	return (d.d - thick) / math.Sin(a)
}

// x returns the abscissa of the roman widened about the origin.
func (d dsFrame) x(x float64) float64 { return x * d.wide }

// romanRun returns the run of ink of the roman at the height y which pick
// chooses.
func (d dsFrame) romanRun(y float64, pick func([][2]float64) [2]float64) [2]float64 {
	rs := runsOf(d.roman, y)
	if len(rs) == 0 {
		panic(fmt.Sprintf("mathfont: no ink at %v", y))
	}
	return pick(rs)
}

// doubleStem returns the counter of a doubled upright stroke whose outer
// left edge is at x, from the baseline to top: the white between its lines,
// a line above the baseline and below the top, as a clockwise contour.
func (d dsFrame) doubleStem(x, bottom, top float64) outline.Contour {
	return reversed(box(x+d.t, bottom+d.t, x+d.d-d.t, top-d.t))
}

// stemFeet runs the foot serifs of a doubled stem from x to x + d.d into the
// pen: the right one from (x + d.d, 0) up to its bracket; the left one is
// stemFootLeft.
func (d dsFrame) footRight(c *pen, x float64) {
	c.run(glyphs.CapFootRight(d.q, x+d.d, d.q.Caps.SerifJut))
}

func (d dsFrame) footLeft(c *pen, x float64) {
	c.run(glyphs.CapFootLeft(d.q, x, d.q.Caps.SerifJut))
}

func (d dsFrame) headLeft(c *pen, x, top float64) {
	c.run(glyphs.CapHeadLeft(d.q, x, top, d.q.Caps.SerifJut))
}

func (d dsFrame) headRight(c *pen, x, top float64) {
	c.run(glyphs.CapHeadRight(d.q, x, top, d.q.Caps.SerifJut))
}

// bowl returns the outer edge of a bowl of one line from the stem at xs to
// its outer extreme xe, between the heights yb and top, as the open contour
// from (xs, yb) round the right to (xs, top), counter-clockwise; and its
// counter, clockwise.  The bowl is the stroke of the symbols, straight at the
// top and at the bottom and round on the right, whose edges have no jump of
// curvature.
func (d dsFrame) bowl(xs, xe, yb, top float64) (outer, counter outline.Contour) {
	rad := (top - yb) / 2
	xf := math.Max(xe-rad, xs+d.t)
	t := d.t
	st := glyphs.Stroke([]glyphs.Station{
		{X: xs, Y: top - t/2, Dir: 0, W: t, Flat: true},
		{X: xf, Y: top - t/2, Dir: 0, W: t},
		{X: xe - t/2, Y: (top + yb) / 2, Dir: 270, W: t},
		{X: xf, Y: yb + t/2, Dir: 180, W: t, Flat: true},
		{X: xs, Y: yb + t/2, Dir: 180, W: t},
	})[0]
	// The contour runs along the inner edge from the start to the end, across
	// the end to (xs, yb), back along the outer edge and across the start.
	from := st.Start
	cut := -1
	for i, sg := range st.Segs {
		if !sg.IsCurve && math.Abs(from.X-xs) < 1e-6 && math.Abs(sg.P3.X-xs) < 1e-6 && math.Abs(sg.P3.Y-yb) < 1e-6 {
			cut = i
		}
		from = sg.P3
	}
	if cut < 0 {
		panic("mathfont: the bowl has no end cut")
	}
	counter = outline.Contour{Start: st.Start, Segs: append(append([]outline.Segment(nil), st.Segs[:cut]...), outline.Segment{P3: st.Start})}
	outer = outline.Contour{Start: st.Segs[cut].P3}
	for _, sg := range st.Segs[cut+1:] {
		outer.Segs = append(outer.Segs, sg)
		if !sg.IsCurve && math.Abs(sg.P3.X-xs) < 1e-6 && math.Abs(sg.P3.Y-top) < 1e-6 {
			break
		}
	}
	return outer, counter
}

// stemLeft returns the left edge of the stem of the roman, widened.
func (d dsFrame) stemLeft() float64 {
	return d.x(d.romanRun(0.3*d.p.CapHeight, firstRun)[0])
}

// dsP draws ℙ: the doubled stem with the foot serifs on both sides and the
// head serif on the left, and the bowl of one line from the top of the stem
// down to the bar of P.
func dsP(d dsFrame) outline.Path {
	return d.bowlLetter(d.p.Caps.PBar, false)
}

// dsR draws ℝ: ℙ with the bar of R and the leg of one line from the bar to
// the baseline, along the centre line of the leg of the roman.
func dsR(d dsFrame) outline.Path {
	return d.bowlLetter(d.p.Caps.RBar, true)
}

func (d dsFrame) bowlLetter(bar float64, leg bool) outline.Path {
	p := d.p
	cap := p.CapHeight
	x := d.stemLeft()
	yb := bar - p.Caps.Bar/2
	// The outer extreme of the bowl of the roman, at the middle of the bowl.
	xe := d.x(d.romanRun((cap+bar)/2, lastRun)[1]) + d.extra
	outer, counter := d.bowl(x+d.d, xe, yb, cap)
	var c pen
	d.footRight(&c, x)
	c.lineTo(pt(x+d.d, yb))
	if leg {
		// The centre line of the leg of the roman, through the middles of
		// its runs at two heights, moved by the added width.
		y0, y1 := 0.12*cap, 0.3*cap
		a, b := d.romanRun(y0, lastRun), d.romanRun(y1, lastRun)
		m0 := pt(d.x((a[0]+a[1])/2)+d.extra, y0)
		m1 := pt(d.x((b[0]+b[1])/2)+d.extra, y1)
		u := pt(m1.X-m0.X, m1.Y-m0.Y)
		l := math.Hypot(u.X, u.Y)
		hw := d.t / 2 * l / u.Y // the half width of the leg along a level line
		at := func(y float64) float64 { return m0.X + (y-m0.Y)*u.X/u.Y }
		// The straight bottom of the bowl right of the leg keeps the length
		// of the shortest segment.
		if dx := at(yb) + hw - (outer.Segs[0].P3.X - minSeg - 4); dx > 0 {
			at0 := at
			at = func(y float64) float64 { return at0(y) - dx }
		}
		c.lineTo(pt(at(yb)-hw, yb))
		c.lineTo(pt(at(0)-hw, 0))
		c.lineTo(pt(at(0)+hw, 0))
		c.lineTo(pt(at(yb)+hw, yb))
		// The outer edge continues from the leg, along the straight bottom.
		outer.Start, outer.Segs = outer.Segs[0].P3, outer.Segs[1:]
	}
	c.run(outer)
	d.headLeft(&c, x, cap)
	d.footLeft(&c, x)
	path := outline.Path{c.close(), d.doubleStem(x, 0, cap), counter}
	return path
}

// dsE draws 𝔼: the doubled stem with the serifs of E on the left, the top
// and bottom arms of E of one line with their beaks, and the middle arm of
// one line, cut upright.
func dsE(d dsFrame) outline.Path {
	p := d.p
	c0 := p.Caps
	cap := p.CapHeight
	x := d.stemLeft()
	xs := x + d.d
	ym := c0.EMiddleHeight
	var c pen
	c.run(glyphs.CapArmBottom(d.q, xs, xs+d.x(c0.EBottomArm), 0, d.t, c0.BottomRise))
	c.lineTo(pt(xs, ym-d.t/2))
	c.lineTo(pt(xs+d.x(c0.EMiddleArm), ym-d.t/2))
	c.lineTo(pt(xs+d.x(c0.EMiddleArm), ym+d.t/2))
	c.lineTo(pt(xs, ym+d.t/2))
	// The beak of the top arm reaches as far below the cap height as that
	// of E.
	c.run(glyphs.CapArmTop(d.q, xs, xs+d.x(c0.ETopArm), cap-d.t, cap, c0.TopDrop+d.t-c0.Bar))
	d.headLeft(&c, x, cap)
	d.footLeft(&c, x)
	return outline.Path{c.close(), d.doubleStem(x, 0, cap)}
}

// dsOne draws 𝟙: the doubled stem of the figure one with the foot serifs of
// the capitals on both sides and the flag of one line from the top of the
// stem down to the tip of the flag of the roman one.
func dsOne(d dsFrame) outline.Path {
	_, _, _, top := d.roman.BBox()
	x := d.stemLeft()
	// The tip of the flag: the leftmost ink of the upper half of the roman.
	tip := pt(math.Inf(1), 0)
	for y := 0.5 * top; y < top; y += 2 {
		if r := runsOf(d.roman, y); len(r) > 0 && r[0][0] < tip.X {
			tip = pt(r[0][0], y)
		}
	}
	tip.X = d.x(tip.X)
	// The upper edge of the flag runs from the top of the stem to the tip;
	// the lower edge is parallel to it, a line lower, across.
	u := pt(tip.X-x, tip.Y-top)
	l := math.Hypot(u.X, u.Y)
	dy := d.t * l / math.Abs(u.X) // the line across, measured upright
	var c pen
	d.footRight(&c, x)
	c.lineTo(pt(x+d.d, top))
	c.lineTo(pt(x, top))
	c.lineTo(tip)
	c.lineTo(pt(tip.X, tip.Y-dy))
	c.lineTo(pt(x, top-dy))
	d.footLeft(&c, x)
	return outline.Path{c.close(), d.doubleStem(x, 0, top)}
}

// dsN draws ℕ: the thin stems of N as lines with the serifs of N, and the
// diagonal doubled, from the top of the left stem to the foot of the right
// stem, its white cut level a line below the cap height and above the
// baseline.
func dsN(d dsFrame) outline.Path {
	p := d.p
	cap := p.CapHeight
	l := d.x(design.CapNStemLeft(p))
	xr := d.x(design.CapNStemRight(p)) + d.diagExtra(p.CapDiag.NThick)
	// The level width of the doubled diagonal, whose lean depends on it.
	dh := d.d
	for range 20 {
		a := math.Atan2(cap, xr-dh-l)
		dh = d.d / math.Sin(a)
	}
	ht := dh * d.t / d.d // the level width of one line of the diagonal
	// e1 is the upper edge of the diagonal, from (l, cap) to (xr - dh, 0),
	// e2 the lower edge, dh to the right.
	slope := (xr - dh - l) / cap // dx per unit of height, going down
	e1 := func(y float64) float64 { return l + (cap-y)*slope }
	yAt := func(x0, x float64) float64 { return cap - (x-x0)/slope }
	t := d.t
	var c pen
	c.run(glyphs.CapFootRight(d.q, l+t, d.q.Caps.SerifJut))
	c.lineTo(pt(l+t, yAt(l, l+t)))
	c.lineTo(pt(xr-dh, 0))
	c.lineTo(pt(xr, 0))
	d.headRight(&c, xr, cap)
	c.run(glyphs.CapHeadLeft(d.q, xr-t, cap, d.q.Caps.SerifJut))
	c.lineTo(pt(xr-t, yAt(l+dh, xr-t)))
	c.lineTo(pt(l+dh, cap))
	d.headLeft(&c, l, cap)
	c.run(glyphs.CapFootLeft(d.q, l, d.q.Caps.SerifJut))
	y0, y1 := t, cap-t
	hole := polygon(pt(e1(y1)+ht, y1), pt(e1(y1)+dh-ht, y1), pt(e1(y0)+dh-ht, y0), pt(e1(y0)+ht, y0))
	return outline.Path{c.close(), hole}
}

// dsZ draws ℤ: the arms of Z of one line, the top arm with the beak of Z
// hanging at its left end and the bottom arm with the beak rising at its
// right end, and the diagonal doubled from the right end of the top arm to
// the left end of the bottom arm, its white cut level by the arms.
func dsZ(d dsFrame) outline.Path {
	p := d.p
	cap := p.CapHeight
	t := d.t
	top := d.romanRun(cap-5, firstRun)
	bot := d.romanRun(5, firstRun)
	ex := d.diagExtra(p.CapDiag.ZDiag)
	xtl, xtr := d.x(top[0]), d.x(top[1])+ex
	xbl, xbr := d.x(bot[0]), d.x(bot[1])+ex
	dh := d.d
	for range 20 {
		a := math.Atan2(cap, xtr-dh-xbl)
		dh = d.d / math.Sin(a)
	}
	ht := dh * d.t / d.d
	// e1 is the upper left edge of the diagonal, from (xtr - dh, cap) to
	// (xbl, 0); e2 is dh to its right.
	e1 := func(y float64) float64 { return xbl + y/cap*(xtr-dh-xbl) }
	top1 := glyphs.CapArmTop(d.q, -(e1(cap - t)), -xtl, cap-t, cap, p.Caps.TopDrop)
	beak := openReversed(mapPoints(outline.Path{top1}, func(q outline.Point) outline.Point { return pt(-q.X, q.Y) })[0])
	var c pen
	c.lineTo(pt(xbl, 0))
	c.run(glyphs.CapArmBottom(d.q, e1(t)+dh, xbr, 0, t, p.Caps.BottomRise))
	c.lineTo(pt(xtr, cap))
	c.lineTo(pt(xtl, cap))
	c.run(beak)
	c.lineTo(pt(e1(cap-t), cap-t))
	c.lineTo(pt(e1(t), t))
	c.lineTo(pt(xbl, t))
	y0, y1 := t, cap-t
	hole := polygon(pt(e1(y1)+ht, y1), pt(e1(y1)+dh-ht, y1), pt(e1(y0)+dh-ht, y0), pt(e1(y0)+ht, y0))
	return outline.Path{c.close(), hole}
}

// dsRound returns the centre line of the round of ℂ and ℚ: the middle and
// the half axes of the ellipse a line inside the outer edge of the roman
// round.
func (d dsFrame) round() (c outline.Point, a, b float64) {
	// The round is that of O, from the left edge of the roman letter; the
	// tail of Q and the ends of C lie outside it.
	o, _, _ := glyphs.Outline("O", d.p)
	ox0, y0, ox1, y1 := o.BBox()
	x0, _, _, _ := d.roman.BBox()
	x1 := x0 + (ox1 - ox0) + d.roundExtra()
	c = pt(d.x((x0+x1)/2), (y0+y1)/2)
	return c, d.x((x1-x0)/2) - d.t/2, (y1-y0)/2 - d.t/2
}

func (d dsFrame) ellipsePoint(c outline.Point, a, b, deg float64) (outline.Point, float64) {
	th := deg * math.Pi / 180
	q := pt(c.X+a*math.Cos(th), c.Y+b*math.Sin(th))
	dir := math.Atan2(b*math.Cos(th), -a*math.Sin(th)) * 180 / math.Pi
	return q, dir
}

// lens returns the inner line of the doubled left side of a round: a
// circular arc of one line which leaves the centre line of the round at
// dsLensJoin degrees, passes a doubled stroke right of the outer edge at
// the middle height and joins the round again below.
func (d dsFrame) lens(c outline.Point, a, b float64) (outline.Path, error) {
	s, _ := d.ellipsePoint(c, a, b, dsLensJoin)
	xin := c.X - a - d.t/2 + d.d - d.t/2 // the centre line of the inner line at the middle
	dx, dy := s.X-xin, s.Y-c.Y
	rho := (dx*dx + dy*dy) / (2 * dx)
	cc := pt(xin+rho, c.Y)
	dirAt := func(q outline.Point) float64 {
		v := pt(q.X-cc.X, q.Y-cc.Y)
		return math.Atan2(v.X, -v.Y) * 180 / math.Pi
	}
	e := pt(s.X, 2*c.Y-s.Y)
	st := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: s, Dir: dirAt(s), K: 1 / rho, W: d.t},
			{P: pt(xin, c.Y), Dir: 270, K: 1 / rho, W: d.t},
			{P: e, Dir: dirAt(e), K: 1 / rho, W: d.t},
		},
		Start: stroke.End{Kind: stroke.Cut},
		End:   stroke.End{Kind: stroke.Cut},
	}
	r, err := st.Build()
	if err != nil {
		return nil, err
	}
	return r.Path(), nil
}

// dsC draws ℂ: the round of C as one line from the end of its upper arm,
// where the roman ends, over the top, down the left and round the bottom to
// the end of its lower arm, its left side doubled by the lens.
func dsC(d dsFrame) outline.Path {
	c, a, b := d.round()
	_, _, x1, _ := d.roman.BBox()
	ex := d.x(x1) + d.roundExtra() - d.t/2
	th := math.Acos(math.Max(-1, math.Min(1, (ex-c.X)/a))) * 180 / math.Pi
	// The arms end at least dsCOpen from the right; the round stays
	// open where the roman C is wide against its O.
	th = math.Max(th, dsCOpen)
	var nodes []stroke.Node
	for _, deg := range []float64{th, 90, 180, 270, 360 - th} {
		q, dir := d.ellipsePoint(c, a, b, deg)
		nodes = append(nodes, stroke.Node{P: q, Dir: dir, K: stroke.Auto, W: d.t})
	}
	st := stroke.Stroke{Nodes: nodes, Start: stroke.End{Kind: stroke.Cut}, End: stroke.End{Kind: stroke.Cut}}
	r, err := st.Build()
	if err != nil {
		panic(err)
	}
	ln, err := d.lens(c, a, b)
	if err != nil {
		panic(err)
	}
	return glyphs.Union(r.Path(), ln)
}

// dsQ draws ℚ: the round of O as a closed line with its left side doubled
// by the lens, and the tail of one line, which starts inside the bottom of
// the round and sweeps down to the right below the baseline.
func dsQ(d dsFrame) outline.Path {
	c, a, b := d.round()
	var nodes []stroke.Node
	for _, deg := range []float64{0, 90, 180, 270} {
		q, dir := d.ellipsePoint(c, a, b, deg)
		nodes = append(nodes, stroke.Node{P: q, Dir: dir, K: stroke.Auto, W: d.t})
	}
	st := stroke.Stroke{Nodes: nodes, Closed: true}
	r, err := st.Build()
	if err != nil {
		panic(err)
	}
	ln, err := d.lens(c, a, b)
	if err != nil {
		panic(err)
	}
	// The tail is a straight line of one line which crosses the round at
	// dsQTailAt degrees from the right, leans dsQTailLean degrees below the
	// level, starts dsQTailIn inside the centre line of the round and ends
	// dsQTail of the cap height below the baseline, cut across.
	at, _ := d.ellipsePoint(c, a, b, dsQTailAt)
	lean := dsQTailLean * math.Pi / 180
	u := pt(math.Cos(lean), -math.Sin(lean))
	s1 := (-dsQTail*d.p.CapHeight + d.t/2 - at.Y) / u.Y
	end := pt(at.X+s1*u.X, at.Y+s1*u.Y)
	start := pt(at.X-dsQTailIn*d.t*u.X, at.Y-dsQTailIn*d.t*u.Y)
	ts := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: start, Dir: math.Mod(-dsQTailLean+360, 360), K: 0, W: d.t},
			{P: end, Dir: math.Mod(-dsQTailLean+360, 360), K: 0, W: d.t},
		},
		Start: stroke.End{Kind: stroke.Cut},
		End:   stroke.End{Kind: stroke.Cut},
	}
	tr, err := ts.Build()
	if err != nil {
		panic(err)
	}
	return glyphs.Union(r.Path(), ln, tr.Path())
}

// dsRegular holds the double-struck letters of the regular parameters, which
// every build of the font and most tests ask for, built once.
var dsRegular = sync.OnceValue(func() []Glyph { return dsGlyphsOf(design.Regular()) })

// dsGlyphs returns the double-struck letters and their script forms, .s1
// and .s2, each with the side spaces of its roman, widened.
func dsGlyphs(p design.Params) []Glyph {
	if reflect.DeepEqual(p, design.Regular()) {
		return slices.Clone(dsRegular())
	}
	return dsGlyphsOf(p)
}

func dsGlyphsOf(p design.Params) []Glyph {
	var out []Glyph
	for _, l := range dsLetters {
		rp, radv, _ := glyphs.Outline(l.roman, p)
		for level := range dsLevels {
			d := newDSFrameOf(p, rp, radv, level)
			if w, ok := dsWiden[l.roman]; ok {
				d.wide *= w
			}
			path := l.draw(d)
			x0, _, x1, _ := d.roman.BBox()
			ux0, _, ux1, _ := path.BBox()
			left := d.x(x0)
			right := d.x(d.adv - x1)
			// The shift is a whole number of units; the rounding of the
			// outline is the rounding of the glyph.
			shift := math.Ceil(left - ux0)
			path = shifted(path, shift, 0)
			g := Glyph{Name: l.name, Advance: math.Ceil(ux1 + shift + right), Path: path, Class: Ord, Like: l.roman}
			switch level {
			case 0:
				g.Rune = l.r
				g.Script = [2]string{l.name + ".s1", l.name + ".s2"}
			case 1:
				g.Name += ".s1"
			case 2:
				g.Name += ".s2"
			}
			out = append(out, g)
		}
	}
	return out
}

// settled returns the outline of a letter moved to the place which dsGlyphsOf
// gives it, the left side of the roman widened; the fairness analysis of the
// candidates of fairest sees the outline which the font has.
func (d dsFrame) settled(path outline.Path) outline.Path {
	x0, _, _, _ := d.roman.BBox()
	ux0, _, _, _ := path.BBox()
	return shifted(path, math.Ceil(d.x(x0)-ux0), 0)
}

func firstRun(r [][2]float64) [2]float64 { return r[0] }
func lastRun(r [][2]float64) [2]float64  { return r[len(r)-1] }

// openReversed returns the open contour run backwards, from its end to its
// start.
func openReversed(c outline.Contour) outline.Contour {
	pts := []outline.Point{c.Start}
	for _, sg := range c.Segs {
		pts = append(pts, sg.P3)
	}
	out := outline.Contour{Start: pts[len(pts)-1]}
	for i := len(c.Segs) - 1; i >= 0; i-- {
		sg := c.Segs[i]
		r := outline.Segment{IsCurve: sg.IsCurve, P3: pts[i]}
		if sg.IsCurve {
			r.P1, r.P2 = sg.P2, sg.P1
		}
		out.Segs = append(out.Segs, r)
	}
	return out
}
