// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// The summation, the product and the coproduct at text and display size.  Each size is
// drawn at its height h, centred on the axis, with the parts of the roman
// capitals: the stems and the thick diagonal are the capital stem grown by
// the law of the big operators, (h / CapHeight) to the power opPower, the
// thin strokes the arm of the capitals grown by the same law, and the serifs
// and the beaks of the arms the serifs of the capitals grown by the same
// law; each operator is the capitals drawn at its weight.
const (
	// opPower is the exponent of the law of the big operators; it puts the
	// thick diagonal of the summation at 1.17 cap stems in text and 1.35 in
	// display.
	opPower = 0.42
	// opTextHeight and opDisplayHeight are the heights of the summation and
	// the product: 1.46 and 2.05 cap heights.
	opTextHeight    = 1000.0
	opDisplayHeight = 1400.0
	// sumWidthText and sumWidthDisplay are the ink widths of the summation
	// over its height.
	sumWidthText    = 0.82
	sumWidthDisplay = 0.78
	// sumVertex is the distance of the point of the summation from the left
	// of its ink over the ink width.
	sumVertex = 0.56
	// sumFoot is the thickness of the bottom bar of the summation over that
	// of its thick diagonal.
	sumFoot = 0.80
	// prodWidthText and prodWidthDisplay are the ink widths of the product
	// over its height.
	prodWidthText    = 0.93
	prodWidthDisplay = 0.85
)

// opSize holds the weights and lengths of a big operator h high.
type opSize struct {
	h           float64
	bottom, top float64
	thick, thin float64
	grow        float64 // the factor of the weights and of the serifs
}

func newOpSize(p design.Params, h float64) opSize {
	g := math.Pow(h/p.CapHeight, opPower)
	return opSize{
		h: h, bottom: p.Punct.MathAxis - h/2, top: p.Punct.MathAxis + h/2,
		thick: p.Caps.Stem * g, thin: p.Caps.Bar * g,
		grow: g,
	}
}

// opParams returns the parameters of the capitals with the serifs, their
// brackets and the serifs of the arms of a big operator of the size s.
func opParams(p design.Params, s opSize) design.Params {
	q := p
	q.Caps.SerifJut *= s.grow
	q.Caps.SerifThick *= s.grow
	q.Caps.BracketHeight *= s.grow
	q.Caps.BracketReach *= s.grow
	q.Caps.ArmSerifThick *= s.grow
	return q
}

// A pen builds one closed contour from runs and lines.
type pen struct {
	c       outline.Contour
	started bool
}

func (b *pen) lineTo(q outline.Point) {
	if !b.started {
		b.c.Start, b.started = q, true
		return
	}
	if cur := b.cur(); math.Hypot(cur.X-q.X, cur.Y-q.Y) > 1e-9 {
		b.c.Segs = append(b.c.Segs, outline.Segment{P3: q})
	}
}

// run appends the open contour r, with a line to its start.
func (b *pen) run(r outline.Contour) {
	b.lineTo(r.Start)
	b.c.Segs = append(b.c.Segs, r.Segs...)
}

func (b *pen) cur() outline.Point {
	if len(b.c.Segs) == 0 {
		return b.c.Start
	}
	return b.c.Segs[len(b.c.Segs)-1].P3
}

// close returns the contour, closed explicitly.
func (b *pen) close() outline.Contour {
	b.lineTo(b.c.Start)
	if cur := b.cur(); cur != b.c.Start {
		b.c.Segs = append(b.c.Segs, outline.Segment{P3: b.c.Start})
	}
	return b.c
}

// meet returns the point where the line through a in the direction u meets
// the line through b in the direction v.
func meet(a, u, b, v outline.Point) outline.Point {
	d := u.X*v.Y - u.Y*v.X
	t := ((b.X-a.X)*v.Y - (b.Y-a.Y)*v.X) / d
	return pt(a.X+t*u.X, a.Y+t*u.Y)
}

// atY returns the point of the line through a in the direction u at the
// height y.
func atY(a, u outline.Point, y float64) outline.Point {
	return pt(a.X+(y-a.Y)/u.Y*u.X, y)
}

// slantedRun returns the horizontal width of a straight stroke t thick
// across which runs from a point on the line y = y0 to a point dx to the
// right of its start on the line y = y1.  The other edge is the line through
// the end point, parallel to it.
func slantedRun(t, dx, dy float64) float64 {
	// The width w satisfies w |dy| / hypot(dx - w, dy) = t.
	lo, hi := t, dx
	for range 60 {
		w := (lo + hi) / 2
		if w*math.Abs(dy)/math.Hypot(dx-w, dy) > t {
			hi = w
		} else {
			lo = w
		}
	}
	return (lo + hi) / 2
}

// opRightSide is the white after the ink of ∑, ∏ and ∐ at both sizes, where
// the scripts of the operator begin; the white on the left of the ink is MathSide.
const opRightSide = 66

// summation draws ∑ h high and w wide: a top bar of the thin stroke with the
// serif of the top arm of E hanging from its right end, a thick diagonal from
// the left end of the top bar down to the point at the axis, a thin diagonal
// from there to the left end of the bottom bar, and a bottom bar sumFoot of
// the thick diagonal with the serif of the bottom arm of E rising at its right
// end, with MathSide before and opRightSide after.
func summation(p design.Params, h, wShare float64) (outline.Path, float64) {
	s := newOpSize(p, h)
	q := opParams(p, s)
	w := math.Round(wShare * h)
	axis := p.Punct.MathAxis
	foot := sumFoot * s.thick
	v := pt(sumVertex*w, axis)
	// The thick diagonal: its upper edge from (a, top) to v, its lower edge
	// from (0, top) parallel to it; the thin one: its lower edge from v to
	// (b, bottom), its upper edge from (0, bottom) parallel to it.
	a := slantedRun(s.thick, v.X, s.top-axis)
	b := slantedRun(s.thin, v.X, axis-s.bottom)
	u := pt(v.X-a, axis-s.top)
	d := pt(b-v.X, s.bottom-axis)
	notch := meet(pt(0, s.top), u, pt(0, s.bottom), d)
	p1 := atY(v, u, s.top-s.thin)
	p2 := atY(v, d, s.bottom+foot)
	var c pen
	c.lineTo(pt(0, s.top))
	c.lineTo(notch)
	c.lineTo(pt(0, s.bottom))
	c.run(glyphs.CapArmBottom(q, p2.X, w, s.bottom, s.bottom+foot, p.Caps.BottomRise*s.grow))
	c.lineTo(v)
	c.run(glyphs.CapArmTop(q, p1.X, w, s.top-s.thin, s.top, p.Caps.TopDrop*s.grow))
	path := outline.Path{c.close()}
	return withSides(path, p.Punct.MathSide, opRightSide)
}

// product draws ∏ h high and w wide: two stems of the thick stroke with the
// foot serifs of the capitals on both sides, joined at the top by a bar of
// the thin stroke which runs out over each stem into the head serif of the
// capitals on its outer side, whose tip is as thick as the bar; MathSide before
// and opRightSide after.
func product(p design.Params, h, wShare float64) (outline.Path, float64) {
	s := newOpSize(p, h)
	q := opParams(p, s)
	head := q
	head.Caps.SerifThick = s.thin
	w := math.Round(wShare * h)
	jut := q.Caps.SerifJut
	xl, xr := jut, w-jut-s.thick // the left edges of the stems
	shift := func(c outline.Contour) outline.Contour {
		return shifted(outline.Path{c}, 0, s.bottom)[0]
	}
	top := s.h
	var c pen
	c.run(shift(glyphs.CapFootRight(q, xl+s.thick, jut)))
	c.lineTo(pt(xl+s.thick, s.top-s.thin))
	c.lineTo(pt(xr, s.top-s.thin))
	c.run(shift(glyphs.CapFootLeft(q, xr, jut)))
	c.run(shift(glyphs.CapFootRight(q, xr+s.thick, jut)))
	c.run(shift(glyphs.CapHeadRight(head, xr+s.thick, top, jut)))
	c.run(shift(glyphs.CapHeadLeft(head, xl, top, jut)))
	c.run(shift(glyphs.CapFootLeft(q, xl, jut)))
	path := outline.Path{c.close()}
	return withSides(path, p.Punct.MathSide, opRightSide)
}

// A bigOp names a big operator and how its sizes are drawn.
type bigOp struct {
	base string
	r    rune
	draw func(p design.Params, display bool) (outline.Path, float64)
}

// bigOps are the big operators; the coproduct is the product mirrored in the
// axis, with its bar at the foot.
var bigOps = []bigOp{
	{"uni2211", 0x2211, func(p design.Params, d bool) (outline.Path, float64) {
		if d {
			return summation(p, opDisplayHeight, sumWidthDisplay)
		}
		return summation(p, opTextHeight, sumWidthText)
	}},
	{"uni220F", 0x220F, prodOf},
	{"uni2210", 0x2210, func(p design.Params, d bool) (outline.Path, float64) {
		path, adv := prodOf(p, d)
		return mirroredY(path, p.Punct.MathAxis), adv
	}},
}

func prodOf(p design.Params, d bool) (outline.Path, float64) {
	if d {
		return product(p, opDisplayHeight, prodWidthDisplay)
	}
	return product(p, opTextHeight, prodWidthText)
}

// bigOpGlyphs returns the text and display sizes of the big operators, the
// text size with the display size as its variant.
func bigOpGlyphs(p design.Params) []Glyph {
	var gs []Glyph
	for _, o := range bigOps {
		path, adv := o.draw(p, false)
		dpath, dadv := o.draw(p, true)
		gs = append(gs,
			Glyph{Name: o.base, Rune: o.r, Advance: math.Round(adv), Path: path, Class: Ord, Like: "H", Variants: []string{o.base + ".d"}},
			Glyph{Name: o.base + ".d", Advance: math.Round(dadv), Path: dpath, Class: Ord, Like: "H"})
	}
	return gs
}
