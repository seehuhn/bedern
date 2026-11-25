// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/stroke"
)

// The remaining double-struck capitals, by the rule of doublestruck.go.  The
// strokes which are thick in the roman and stand upright or lean are doubled:
// the stems of I, L, F, H, B, D, T, K, J, U and Y, the diagonals of A, V, M, W
// and X, the right stem of M, the spine of S; every other stroke is one line.
// A leaning stroke has the serifs of the capital stems, sheared by its lean;
// the outer edges of the roman letter are kept where two strokes meet, which
// keeps the width and the vertices of its roman.

// dsI draws I: the doubled stem with the serifs of the capitals on all four
// corners.
func dsI(d dsFrame) outline.Path {
	cap := d.p.CapHeight
	x := d.stemLeft()
	var c pen
	d.footRight(&c, x)
	d.headRight(&c, x+d.d, cap)
	d.headLeft(&c, x, cap)
	d.footLeft(&c, x)
	return outline.Path{c.close(), d.doubleStem(x, 0, cap)}
}

// dsL draws L: the doubled stem with the serifs of E on the left and at the
// head on the right, and the bottom arm of one line with its beak.
func dsL(d dsFrame) outline.Path {
	c0 := d.p.Caps
	cap := d.p.CapHeight
	x := d.stemLeft()
	xs := x + d.d
	var c pen
	c.run(glyphs.CapArmBottom(d.q, xs, xs+d.x(c0.LArm), 0, d.t, c0.BottomRise))
	d.headRight(&c, x+d.d, cap)
	d.headLeft(&c, x, cap)
	d.footLeft(&c, x)
	return outline.Path{c.close(), d.doubleStem(x, 0, cap)}
}

// dsF draws F: the doubled stem with the foot serifs on both sides and the
// head serif on the left, the top arm of one line with its beak and the
// middle arm of one line, cut upright.
func dsF(d dsFrame) outline.Path {
	c0 := d.p.Caps
	cap := d.p.CapHeight
	x := d.stemLeft()
	xs := x + d.d
	ym := c0.FMiddleHeight
	var c pen
	d.footRight(&c, x)
	c.lineTo(pt(xs, ym-d.t/2))
	c.lineTo(pt(xs+d.x(c0.FMiddleArm), ym-d.t/2))
	c.lineTo(pt(xs+d.x(c0.FMiddleArm), ym+d.t/2))
	c.lineTo(pt(xs, ym+d.t/2))
	c.run(glyphs.CapArmTop(d.q, xs, xs+d.x(c0.FTopArm), cap-d.t, cap, c0.TopDrop+d.t-c0.Bar))
	d.headLeft(&c, x, cap)
	d.footLeft(&c, x)
	return outline.Path{c.close(), d.doubleStem(x, 0, cap)}
}

// thinStem returns a stem of one line with the left edge at x, from the
// baseline to top, with the serifs of the capitals of the given reaches at the
// foot on the left and the right and at the head on the left and the right.
func (d dsFrame) thinStem(x, top, fl, fr, hl, hr float64) outline.Path {
	q := d.q
	q.Caps.Stem = d.t
	return glyphs.CapStemOutline(q, x, top, fl, fr, hl, hr)
}

// dsH draws ℍ: the doubled left stem and the right stem of one line, both
// with the four serifs of H, and the bar of one line between them.
func dsH(d dsFrame) outline.Path {
	c0 := d.p.Caps
	cap := d.p.CapHeight
	j := c0.SerifJut
	x := d.stemLeft()
	xs := x + d.d
	r := d.romanRun(0.3*cap, lastRun)
	xr := d.x((r[0]+r[1])/2) + d.extra - d.t/2
	var c pen
	d.footRight(&c, x)
	d.headRight(&c, x+d.d, cap)
	d.headLeft(&c, x, cap)
	d.footLeft(&c, x)
	left := outline.Path{c.close(), d.doubleStem(x, 0, cap)}
	bar := outline.Path{box(xs-d.t/2, c0.HBarHeight-d.t/2, xr+d.t/2, c0.HBarHeight+d.t/2)}
	return glyphs.Union(left, bar, d.thinStem(xr, cap, j, j, j, j))
}

// dsT draws T: the doubled stem with the foot serifs on both sides, and the
// bar of one line at the top with the beaks of T hanging at both ends.
func dsT(d dsFrame) outline.Path {
	c0 := d.p.Caps
	cap := d.p.CapHeight
	r := d.romanRun(0.3*cap, firstRun)
	x := d.x((r[0]+r[1])/2) - d.d/2
	xs := x + d.d
	arm := d.x(c0.TArm)
	lower := cap - d.t
	drop := c0.TDrop + d.t - c0.TBar
	right := glyphs.CapArmTop(d.q, xs, xs+arm, lower, cap, drop)
	mirrored := glyphs.CapArmTop(d.q, -x, -x+arm, lower, cap, drop)
	left := openReversed(mapPoints(outline.Path{mirrored}, func(q outline.Point) outline.Point { return pt(-q.X, q.Y) })[0])
	var c pen
	d.footRight(&c, x)
	c.lineTo(pt(xs, lower))
	c.run(right)
	c.run(left)
	d.footLeft(&c, x)
	return outline.Path{c.close(), d.doubleStem(x, 0, cap)}
}

// dsD draws D: the doubled stem with the serifs of the capitals on the left
// and the bowl of one line from the baseline to the cap height.
func dsD(d dsFrame) outline.Path {
	cap := d.p.CapHeight
	x := d.stemLeft()
	xe := d.x(d.romanMaxX(0, cap)) + d.extra
	outer, counter := d.bowl(x+d.d, xe, 0, cap)
	var c pen
	c.run(outer)
	d.headLeft(&c, x, cap)
	d.footLeft(&c, x)
	return outline.Path{c.close(), d.doubleStem(x, 0, cap), counter}
}

// dsB draws B: ℙ and ℝ's stem with two bowls of one line, the lower from the
// baseline to the upper edge of the waist and the upper from the lower edge of
// the waist to the cap height, which share the waist line and whose outer
// edges cross in the notch of B.
func dsB(d dsFrame) outline.Path {
	cap := d.p.CapHeight
	wy := d.p.Caps.BWaist
	x := d.stemLeft()
	xs := x + d.d
	xeU := d.x(d.romanMaxX(wy+d.t, cap)) + d.extra
	xeL := d.x(d.romanMaxX(0, wy-d.t)) + d.extra
	outerL, counterL := d.bowl(xs, xeL, 0, wy+d.t/2)
	outerU, counterU := d.bowl(xs, xeU, wy-d.t/2, cap)
	ia, ta, ib, tb, ok := crossing(outerL, outerU)
	if !ok {
		panic("mathfont: the bowls of B do not cross")
	}
	var c pen
	c.run(headTo(outerL, ia, ta))
	c.run(tailFrom(outerU, ib, tb))
	d.headLeft(&c, x, cap)
	d.footLeft(&c, x)
	return outline.Path{c.close(), d.doubleStem(x, 0, cap), counterL, counterU}
}

// dsO draws O: the round of O as a closed line with its left side doubled by
// the lens, as Q has it.
func dsO(d dsFrame) outline.Path {
	c, a, b := d.round()
	var nodes []stroke.Node
	for _, deg := range []float64{0, 90, 180, 270} {
		q, dir := d.ellipsePoint(c, a, b, deg)
		nodes = append(nodes, stroke.Node{P: q, Dir: dir, K: stroke.Auto, W: d.t})
	}
	r, err := stroke.Stroke{Nodes: nodes, Closed: true}.Build()
	if err != nil {
		panic(err)
	}
	ln, err := d.lens(c, a, b)
	if err != nil {
		panic(err)
	}
	return glyphs.Union(r.Path(), ln)
}

// dsA draws 𝔸: the thin left diagonal of one line and the right diagonal
// doubled, which keep the outer edges of the roman and meet in the flat of its
// apex; the crossbar of one line; the serifs of the feet on both sides of both
// diagonals.  The white of the doubled diagonal stops a line from the left
// outer edge near the apex.
func dsA(d dsFrame) outline.Path {
	p := d.p
	cd := p.CapDiag
	_, _, _, top := d.roman.BBox()
	h0, h1 := d.heightsWith(2, 0.15, 0.85)
	e1 := d.edgeOf(h0, h1, firstRun, 0)
	e4 := d.edgeOf(h0, h1, lastRun, 1)
	wL, wR := across(d.t, e1.s), across(d.d, e4.s)
	i1, i4 := e1.shift(wL), e4.shift(-wR)
	ym := cd.ABar
	lo, up := ym-d.t/2, ym+d.t/2
	yc := crossY(i1, i4)
	var c pen
	d.diagFootLeft(&c, e1, d.diagJut())
	d.diagFootRight(&c, i1, d.diagJut())
	c.lineTo(pt(i1.at(lo), lo))
	c.lineTo(pt(i4.at(lo), lo))
	d.diagFootLeft(&c, i4, d.diagJut())
	d.diagFootRight(&c, e4, d.diagJut())
	c.lineTo(pt(e4.at(top), top))
	c.lineTo(pt(e1.at(top), top))
	path := outline.Path{c.close(), reversed(polygon(pt(i1.at(up), up), pt(i4.at(up), up), pt(i1.at(yc), yc)))}
	ht := across(d.t, e4.s)
	htL := across(d.t, e1.s)
	// The white of the doubled diagonal ends flat where it is still wide
	// enough, under the apex.
	l, r := i4.shift(ht), e4.shift(-ht)
	left := func(y float64) float64 { return math.Max(l.at(y), e1.at(y)+htL) }
	yt := top - d.t
	for ; yt > 0 && r.at(yt)-left(yt) < dsATip*d.g; yt-- {
	}
	return append(path, slotBetween(l, r, d.t, yt, rightOf(e1.shift(htL)))...)
}

// dsATip is the width of the white of the doubled diagonal of 𝔸 where it ends
// under the apex, over its width at the foot.
const dsATip = 0.85

// dsV draws V: the left diagonal doubled and the right one of one line, which
// keep the outer edges of the roman and meet in the flat of its vertex, the
// serifs at the heads on both sides of both.
func dsV(d dsFrame) outline.Path {
	p := d.p
	cap := p.CapHeight
	_, yv, _, _ := d.roman.BBox()
	h0, h1 := d.heightsWith(2, 0.15, 0.85)
	e1 := d.edgeOf(h0, h1, firstRun, 0)
	e4 := d.edgeOf(h0, h1, lastRun, 1)
	wL, wR := across(d.d, e1.s), across(d.t, e4.s)
	i1, i4 := e1.shift(wL), e4.shift(-wR)
	yc := crossY(i1, i4)
	var c pen
	c.lineTo(pt(e1.at(yv), yv))
	c.lineTo(pt(e4.at(yv), yv))
	d.diagHeadRight(&c, e4, cap, d.diagJut())
	d.diagHeadLeft(&c, i4, cap, d.diagJut())
	c.lineTo(pt(i4.at(yc), yc))
	d.diagHeadRight(&c, i1, cap, d.diagJut())
	d.diagHeadLeft(&c, e1, cap, d.diagJut())
	path := outline.Path{c.close()}
	ht, htR := across(d.t, e1.s), across(d.t, e4.s)
	return append(path, slotBetween(e1.shift(ht), e1.shift(wL-ht), yv+d.t, cap-d.t, leftOf(e4.shift(-htR)))...)
}

// dsX draws X: the thick diagonal doubled and the thin one of one line, each
// with the serifs of its four ends, crossing in the middle.
func dsX(d dsFrame) outline.Path {
	p := d.p
	cap := p.CapHeight
	// The strokes read where they are apart, above and below the crossing.
	yu, yl := d.heightWith(2, 0.62, 0.88), d.heightWith(2, 0.12, 0.38)
	thick := d.centreOf(yu, yl, 0, 1)
	thin := d.centreOf(yu, yl, 1, 0)
	wT, wN := across(d.d, thick.s), across(d.t, thin.s)
	ht := across(d.t, thick.s)
	tl, tr := thick.shift(-wT/2), thick.shift(wT/2)
	nl, nr := thin.shift(-wN/2), thin.shift(wN/2)
	big := outline.Path{d.rod(tl, tr, 0, cap, serifs{d.diagJut(), d.diagJut(), d.diagJut(), d.diagJut()})}
	big = append(big, slotBetween(tl.shift(ht), tr.shift(-ht), d.t, cap-d.t)...)
	small := outline.Path{d.rod(nl, nr, 0, cap, serifs{d.diagJut(), d.diagJut(), d.diagJut(), d.diagJut()})}
	return glyphs.Union(big, small)
}

// dsK draws K: the doubled stem with the serifs of the capitals, the arm and
// the leg of one line, each with the serifs of its end, from the stem and the
// arm of the roman, moved right by what the doubled stem adds.
func dsK(d dsFrame) outline.Path {
	p := d.p
	cap := p.CapHeight
	x := d.stemLeft()
	xs := x + d.d
	var c pen
	d.footRight(&c, x)
	d.headRight(&c, x+d.d, cap)
	d.headLeft(&c, x, cap)
	d.footLeft(&c, x)
	stem := outline.Path{c.close(), d.doubleStem(x, 0, cap)}
	// The arm leaves the stem inside its right line, cut upright.
	ya0, ya1 := d.heightsWith(2, 0.5, 0.88)
	arm := d.centreOf(ya0, ya1, 1, 1).shift(d.extra)
	wA := across(d.t, arm.s)
	al, ar := arm.shift(-wA/2), arm.shift(wA/2)
	xc := xs - d.t/2
	var a pen
	a.lineTo(pt(xc, ar.yAt(xc)))
	d.diagHeadRight(&a, ar, cap, d.diagJut())
	d.diagHeadLeft(&a, al, cap, d.diagJut())
	a.lineTo(pt(xc, al.yAt(xc)))
	armPath := outline.Path{a.close()}
	// The leg starts inside the arm, cut level, and ends on the baseline.
	yg0, yg1 := d.heightsWith(2, 0.12, 0.45)
	leg := d.centreOf(yg0, yg1, 1, 1).shift(d.extra)
	wG := across(d.t, leg.s)
	yj := crossY(arm, leg)
	legPath := outline.Path{d.legRod(leg.shift(-wG/2), leg.shift(wG/2), yj, d.diagJut())}
	return glyphs.Union(stem, armPath, legPath)
}

// yAt returns the height at which the line has the abscissa x.
func (e xline) yAt(x float64) float64 { return (x - e.x0) / e.s }

// legRod returns a stroke from the baseline, with the serifs of the foot on
// both sides, to the level cut at the height top.
func (d dsFrame) legRod(l, r xline, top, jut float64) outline.Contour {
	var c pen
	d.diagFootRight(&c, r, jut)
	c.lineTo(pt(r.at(top), top))
	c.lineTo(pt(l.at(top), top))
	d.diagFootLeft(&c, l, jut)
	return c.close()
}

// dsY draws Y: the doubled stem from the baseline to the crotch, the left
// arm doubled, with three lines at 0.7 of the cap height, and the right arm of
// one line, which leave the
// two lines of the stem and carry the serifs of their heads on both sides.
func dsY(d dsFrame) outline.Path {
	p := d.p
	cd := p.CapDiag
	cap := p.CapHeight
	r := d.romanRun(0.3*cap, firstRun)
	xs0 := d.x((r[0]+r[1])/2) - d.d/2
	ys0, ys1 := d.heightsWith(2, 0.45, 0.88)
	sl := d.centreOf(ys0, ys1, 0, 0).s
	sAbs := math.Abs(sl)
	wA := across(d.t, sAbs)
	wL := across(d.d, sAbs)
	yc := cd.YJunction
	ya := yc - (wL+wA-d.d)/(2*sAbs)
	le := xline{xs0 + sAbs*ya, -sAbs}      // the lower left edge of the left arm
	re := xline{xs0 + d.d - sAbs*ya, sAbs} // the lower right edge of the right arm
	li, ri := le.shift(wL), re.shift(-wA)
	var c pen
	d.footRight(&c, xs0)
	c.lineTo(pt(xs0+d.d, ya))
	d.diagHeadRight(&c, re, cap, d.diagJut())
	d.diagHeadLeft(&c, ri, cap, d.diagJut())
	c.lineTo(pt(ri.at(yc), yc))
	d.diagHeadRight(&c, li, cap, d.diagJut())
	d.diagHeadLeft(&c, le, cap, d.diagJut())
	c.lineTo(pt(xs0, ya))
	d.footLeft(&c, xs0)
	path := outline.Path{c.close()}
	// The white of the stem runs on up the left arm as one white: a line
	// from the left edge of the stem and of the arm, and a line from the right
	// edge of the stem and from the inner edge of the arm.
	ht := across(d.t, sAbs)
	y1 := ya + (ht-d.t)/sAbs
	y2 := ya + (wL-ht-d.d+d.t)/sAbs
	return append(path, reversed(polygon(
		pt(xs0+d.t, d.t), pt(xs0+d.d-d.t, d.t), pt(xs0+d.d-d.t, y2),
		pt(li.at(cap-d.t)-ht, cap-d.t), pt(le.at(cap-d.t)+ht, cap-d.t), pt(xs0+d.t, y1))))
}
