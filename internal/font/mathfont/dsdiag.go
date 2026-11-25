// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// M, W and S of the double-struck capitals, by the rule of dscaps.go.

// dsM draws 𝕄: the thin left stem and the thin right diagonal of one line,
// the left diagonal and the right stem doubled, which keep the outer edges of
// the roman; the vertex has its flat and each white wedge between a stem and
// a diagonal its depth.
func dsM(d dsFrame) outline.Path {
	p := d.p
	cd := p.CapDiag
	cap := p.CapHeight
	jut := p.Caps.SerifJut
	ya, yb := d.heightsWith(4, 0.15, 0.85)
	ll1 := d.edgeIdx(ya, yb, 1, 0)
	rr2 := d.edgeIdx(ya, yb, 2, 1)
	w1, w2 := across(d.d, ll1.s), across(d.t, rr2.s)
	ur1, ul2 := ll1.shift(w1), rr2.shift(-w2)
	yv := cd.MHeight
	// The white wedges between a stem and a diagonal have the place of the
	// roman: its corners, which stand cd.MCrotchDepth under the cap height.
	yw := cap - cd.MCrotchDepth
	tl, tr := d.cornerNear(pt(ll1.at(yw), yw)), d.cornerNear(pt(rr2.at(yw), yw))
	a2 := d.x(tl.X)
	ll1 = ll1.shift(a2 - ll1.at(tl.Y))
	// The left stem keeps the left edge of the roman; the letter keeps its
	// left side space.
	a1 := d.x(d.romanRun(0.3*cap, firstRun)[0])
	b1 := d.x(tr.X)
	rr2 = rr2.shift(b1 - rr2.at(tr.Y))
	ur1, ul2 = ll1.shift(w1), rr2.shift(-w2)
	xr := b1 + d.d
	yw1, yw2 := tl.Y, tr.Y
	yc := crossY(ur1, ul2)
	var c pen
	c.run(glyphs.CapFootRight(d.q, a2, jut))
	c.lineTo(pt(a2, yw1))
	c.lineTo(pt(ll1.at(yv), yv))
	c.lineTo(pt(rr2.at(yv), yv))
	c.lineTo(pt(b1, yw2))
	c.run(glyphs.CapFootLeft(d.q, b1, jut))
	c.run(glyphs.CapFootRight(d.q, xr, jut))
	c.run(glyphs.CapHeadRight(d.q, xr, cap, jut))
	c.lineTo(pt(ul2.at(cap), cap))
	c.lineTo(pt(ul2.at(yc), yc))
	c.lineTo(pt(ur1.at(cap), cap))
	c.run(glyphs.CapHeadLeft(d.q, a1, cap, jut))
	c.run(glyphs.CapFootLeft(d.q, a1, jut))
	path := outline.Path{c.close()}
	h1, h2 := across(d.t, ll1.s), across(d.t, rr2.s)
	path = append(path, slotBetween(ll1.shift(h1), ur1.shift(-h1), yv+d.t, cap-d.t, leftOf(rr2.shift(-h2)))...)
	return append(path, d.doubleStem(b1, 0, cap))
}

// dsW draws 𝕎: the first and the third stroke doubled and the second and the
// fourth of one line, which keep the outer edges of the roman; each vertex
// has its flat and the middle apex its height.
func dsW(d dsFrame) outline.Path {
	p := d.p
	cd := p.CapDiag
	cap := p.CapHeight
	yb0, yb1 := d.heightsWith(4, 0.15, 0.85)
	ll1 := d.edgeIdx(yb0, yb1, 0, 0)
	rr2 := d.edgeIdx(yb0, yb1, 1, 1)
	ll3 := d.edgeIdx(yb0, yb1, 2, 0)
	rr4 := d.edgeIdx(yb0, yb1, 3, 1)
	w1, w2 := across(d.d, ll1.s), across(d.t, rr2.s)
	w3, w4 := across(d.d, ll3.s), across(d.t, rr4.s)
	ur1, ul2 := ll1.shift(w1), rr2.shift(-w2)
	ur3, ul4 := ll3.shift(w3), rr4.shift(-w4)
	_, yv, _, _ := d.roman.BBox()
	ya := cd.WApexHeight
	y1, ym, y2 := crossY(ur1, ul2), crossY(rr2, ll3), crossY(ur3, ul4)
	var c pen
	d.diagHeadLeft(&c, ll1, cap, d.diagJut())
	c.lineTo(pt(ll1.at(yv), yv))
	c.lineTo(pt(rr2.at(yv), yv))
	c.lineTo(pt(rr2.at(ym), ym))
	c.lineTo(pt(ll3.at(yv), yv))
	c.lineTo(pt(rr4.at(yv), yv))
	d.diagHeadRight(&c, rr4, cap, d.diagJut())
	d.diagHeadLeft(&c, ul4, cap, d.diagJut())
	c.lineTo(pt(ul4.at(y2), y2))
	c.lineTo(pt(ur3.at(ya), ya))
	c.lineTo(pt(ul2.at(ya), ya))
	c.lineTo(pt(ul2.at(y1), y1))
	d.diagHeadRight(&c, ur1, cap, d.diagJut())
	path := outline.Path{c.close()}
	h1, h2, h3, h4 := across(d.t, ll1.s), across(d.t, rr2.s), across(d.t, ll3.s), across(d.t, rr4.s)
	path = append(path, slotBetween(ll1.shift(h1), ur1.shift(-h1), yv+d.t, cap-d.t, leftOf(rr2.shift(-h2)))...)
	// The white of the third stroke ends a line from the outer edge of the
	// second under the apex and a line from the outer edge of the fourth.
	l, r := ll3.shift(h3), ur3.shift(-h3)
	left := func(y float64) float64 { return math.Max(l.at(y), ul2.at(y)+h2) }
	yt := ya - d.t
	for ; yt > 0 && r.at(yt)-left(yt) < d.g/2; yt-- {
	}
	return append(path, slotBetween(l, r, yv+d.t, yt, rightOf(ul2.shift(h2)), leftOf(rr4.shift(-h4)))...)
}
