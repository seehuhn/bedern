// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// The arrows.
// → and ⇒ are new drawings and the rest of the family is made from their
// parts by rule: mirrored, turned by a quarter, with a bar or a hook at the
// tail, or with a longer shaft.
//
// A head is a chevron: two barbs whose edges are concentric circular arcs,
// concave on the outside; a barb leaves the tip at a shallow angle
// and opens towards its end, where it is cut along its radius; the tip is cut
// flat, MathVertexFlat high as the vertex of <, and the inner edges meet at a
// crotch on the axis.  The shafts are bars of the rule on the axis (→) or
// with the white of the equals between them (⇒), run into the head and
// joined to it by glyphs.Union.  The ink of an arrow has MathSide on each
// side.
const (
	// arrowWidth is the ink width of → and its family.
	arrowWidth = 880.0
	// longArrowWidth is the ink width of the long arrows.
	longArrowWidth = 1360.0
	// headHeight is the ink height of the head of →.
	headHeight = 372.0
	// headLength is the length of the head of → from the tip to the inner
	// corner of a barb's end.
	headLength = 238
	// barbShare is the thickness of a barb over the rule.
	barbShare = 0.88
	// barbBend is the turn of the edges of a barb from the tip to its end,
	// in degrees; the chord of the outer edge lies at about 40 degrees to
	// the shaft.
	barbBend = 30
	// hookTop is the top of the hook of ↪ above the axis.
	hookTop = 272.0
	// hookStretch is the length of the lower quarter of the hook over its
	// radius; the bend leaves the shaft without a jump of curvature.
	hookStretch = 1.7
	// hookEnd is the straight run at the end of the hook over the half
	// height of the hook.
	hookEnd = 0.4
	// vecWidth is the ink width of the vector accent; its head is vecShare
	// of the head of →.
	vecWidth = 420
	vecShare = 0.68
)

// A chevron is the head of an arrow in its own frame: the tip at the
// origin, pointing to the right, the axis at y = 0.
type chevron struct {
	f, w   float64 // the flat at the tip, the thickness of a barb across
	a0, a1 float64 // the angles of the outer edge to the axis at the tip and at the end, radians
	h      float64 // the top of the outer edge, half the height of the ink
	c      outline.Point
	r      float64 // the centre and the radius of the outer edge of the upper barb
	t0, t1 float64 // the position angles of the outer edge at the tip and at the end
}

// newChevron returns the chevron with the half height h, the angles a0 and
// a1, the flat f and the barb w.
func newChevron(h, a0, a1, f, w float64) chevron {
	r := (h - f/2) / (math.Cos(a0) - math.Cos(a1))
	c := pt(r*math.Sin(a0), f/2+r*math.Cos(a0))
	t0 := math.Atan2(-math.Cos(a0), -math.Sin(a0))
	return chevron{f: f, w: w, a0: a0, a1: a1, h: h, c: c, r: r, t0: t0, t1: t0 - (a1 - a0)}
}

// chevronOfLength returns the chevron with the half height h whose edges
// turn by bend from the tip to the end and whose head is l long from the tip
// to the inner corner of a barb's end.
func chevronOfLength(h, l, bend, f, w float64) chevron {
	lo, hi := -0.2, math.Pi/2-bend // the edge may leave the tip a little below the level
	var ch chevron
	for range 60 {
		a := (lo + hi) / 2
		ch = newChevron(h, a, a+bend, f, w)
		if ch.length() > l {
			lo = a
		} else {
			hi = a
		}
	}
	if math.Abs(ch.length()-l) > 0.5 {
		panic(fmt.Sprintf("mathfont: no head of the half height %.0f with a bend of %.0f degrees is %.0f long", h, bend*180/math.Pi, l))
	}
	return ch
}

// length returns the distance from the tip to the inner corner of the end.
func (ch chevron) length() float64 {
	return -(ch.c.X + (ch.r+ch.w)*math.Cos(ch.t1))
}

// innerAngle returns the position angle at which the inner edge of the upper
// barb is at the height y.
func (ch chevron) innerAngle(y float64) float64 {
	s := (y - ch.c.Y) / (ch.r + ch.w)
	if s < -1 {
		panic(fmt.Sprintf("mathfont: the inner edge of the head does not reach %.1f", y))
	}
	return -math.Pi - math.Asin(s)
}

// path returns the outline of the chevron: one contour, counter-clockwise.
func (ch chevron) path() outline.Path {
	ri := ch.r + ch.w
	tv := ch.innerAngle(0)
	var segs []outline.Segment
	segs = append(segs, arcSegs(ch.c, ch.r, ch.t0, ch.t1)...)
	segs = append(segs, outline.Segment{P3: pt(ch.c.X+ri*math.Cos(ch.t1), ch.c.Y+ri*math.Sin(ch.t1))})
	segs = append(segs, arcSegs(ch.c, ri, ch.t1, tv)...)
	// The lower barb is the upper one mirrored in the axis and traversed
	// from the crotch back to the tip.
	mirror := func(q outline.Point) outline.Point { return pt(q.X, -q.Y) }
	from := pt(0, ch.f/2)
	all := append([]outline.Segment{{P3: from}}, segs...)
	for i := len(segs) - 1; i >= 0; i-- {
		s := segs[i]
		prev := from
		if i > 0 {
			prev = segs[i-1].P3
		}
		all = append(all, outline.Segment{IsCurve: s.IsCurve, P1: mirror(s.P2), P2: mirror(s.P1), P3: mirror(prev)})
	}
	return outline.Path{{Start: pt(0, -ch.f/2), Segs: all}}
}

// spanAt returns the run of ink of the barb at the height y > 0, from the
// inner edge to the outer edge.
func (ch chevron) spanAt(y float64) (x0, x1 float64) {
	rs := runAt(ch.path(), y)
	if len(rs) == 0 {
		return 0, 0
	}
	return rs[len(rs)-1][0], rs[len(rs)-1][1]
}

// shaftEnd returns the abscissa at which a shaft whose edges lie at y0 and
// y1 above the axis ends inside the barb: halfway between the inner and the
// outer edge at the middle of the shaft, where both edges of the shaft cross
// the inner edge of the barb.
func (ch chevron) shaftEnd(y0, y1 float64) float64 {
	x0, x1 := ch.spanAt((y0 + y1) / 2)
	x := (x0 + x1) / 2
	for _, y := range []float64{y0, y1} {
		a, b := ch.spanAt(y)
		if !(a < x && x < b) {
			panic(fmt.Sprintf("mathfont: a shaft at %.0f to %.0f does not end inside the head", y0, y1))
		}
	}
	return x
}

// arrowParts holds the parts of the arrows of one set of parameters.
type arrowParts struct {
	t      float64 // the rule
	single chevron // the head of →
	double chevron // the head of ⇒
	gap    float64 // the white between the shafts of ⇒
}

func newArrowParts(r *roman) arrowParts {
	p := r.p.Punct
	t := p.MathThick
	w := math.Round(barbShare * t)
	bend := barbBend * math.Pi / 180
	single := chevronOfLength(headHeight/2, headLength, bend, p.MathVertexFlat, w)
	// The head of ⇒ is that of → grown by the white of the equals, with the
	// same angles.
	double := newChevron(headHeight/2+p.EqualsGap/2, single.a0, single.a1, p.MathVertexFlat, w)
	return arrowParts{t: t, single: single, double: double, gap: p.EqualsGap}
}

// arrow returns the ink of an arrow of the given ink width on the axis at
// y = 0, its tail at x = 0, with a head at the right end and, if both is
// true, at the left end; double gives the double arrow.
func (a arrowParts) arrow(width float64, double, both bool) outline.Path {
	ch, shafts := a.single, [][2]float64{{-a.t / 2, a.t / 2}}
	if double {
		ch = a.double
		shafts = [][2]float64{{-a.gap/2 - a.t, -a.gap / 2}, {a.gap / 2, a.gap/2 + a.t}}
	}
	parts := []outline.Path{shifted(ch.path(), width, 0)}
	x0, x1 := 0.0, width+ch.shaftEnd(a.shaftSpan(double))
	if both {
		parts = append(parts, mirroredX(parts[0], width/2))
		x0 = width - x1
	}
	for _, s := range shafts {
		parts = append(parts, outline.Path{box(x0, s[0], x1, s[1])})
	}
	return glyphs.Union(parts...)
}

// shaftSpan returns the edges of the upper shaft above the axis.
func (a arrowParts) shaftSpan(double bool) (y0, y1 float64) {
	if double {
		return a.gap / 2, a.gap/2 + a.t
	}
	return 0, a.t / 2
}

// arrowSign wraps the construction of an arrow pointing right as a glyph with
// MathSide on each side, a level arrow with its shaft on the axis and an
// upright one with its ink centred on the axis; turns gives the quarter turns
// counter-clockwise, mirror a mirror image in a vertical line first.
func arrowSign(draw func(a arrowParts, r *roman) outline.Path, mirror bool, turns int) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		a := newArrowParts(r)
		path := draw(a, r)
		if mirror {
			path = mirroredX(path, 0)
		}
		if turns != 0 {
			path = rotated(path, pt(0, 0), float64(turns)*math.Pi/2)
		}
		dy := r.p.Punct.MathAxis // a level arrow has its shaft on the axis
		if turns != 0 {
			_, y0, _, y1 := path.BBox()
			dy = math.Round(r.p.Punct.MathAxis - (y0+y1)/2)
		}
		return withSides(shifted(path, 0, dy), r.p.Punct.MathSide, r.p.Punct.MathSide)
	}
}

func plainArrow(width float64, double, both bool) func(a arrowParts, r *roman) outline.Path {
	return func(a arrowParts, r *roman) outline.Path { return a.arrow(width, double, both) }
}

// mapsto draws ↦, or ⟼ if long: the arrow with a bar of the rule across its
// tail, as high as the head.
func mapsto(width float64) func(a arrowParts, r *roman) outline.Path {
	return func(a arrowParts, r *roman) outline.Path {
		h := a.single.h
		bar := outline.Path{box(0, -h, a.t, h)}
		shaft := a.arrow(width-a.t/2, false, false)
		return glyphs.Union(bar, shifted(shaft, a.t/2, 0))
	}
}

// MapstocharRune is the code point of the bar of ↦ on its own, which the
// command \mapstochar of LaTeX sets for the extensible \xmapsto of mathtools:
// \mapstochar is the character 0x37 of the symbol family (the figure 7 in a
// Unicode font), so the file of ranges and the README point it at this
// glyph instead.
const MapstocharRune = 0xE500

// mapstochar draws the bar of ↦, as high as the head and as heavy as the
// rule, with the side space of the signs on its left and none on its right;
// the shaft of the arrow which follows it starts one side space inside its
// own advance and meets it.
func mapstochar(r *roman) (outline.Path, float64) {
	a := newArrowParts(r)
	h := a.single.h
	side := r.p.Punct.MathSide
	path, adv := withSides(outline.Path{box(0, -h, a.t, h)}, side, -side)
	return shifted(path, 0, r.p.Punct.MathAxis), adv
}

// hookArrow draws ↪: the arrow whose tail turns up through half a turn to
// the left and runs back to the right above the shaft, the top of its ink
// hookTop above the axis; the hook is a stroke of the rule (glyphs.Stroke)
// whose two quarters are hookStretch times as long as they are high and leave
// the straight runs without a jump of curvature.
func hookArrow(a arrowParts, r *roman) outline.Path {
	t := a.t
	rad := (hookTop - t/2) / 2 // half the height of the centre line of the hook
	x := t/2 + hookStretch*rad // the abscissa of the ends of the bend
	ch := a.single
	end := arrowWidth + ch.shaftEnd(0, t/2)
	hook := glyphs.Stroke([]glyphs.Station{
		{X: end, Y: 0, Dir: 180, W: t, Flat: true},
		{X: x, Y: 0, Dir: 180, W: t},
		{X: t / 2, Y: rad, Dir: 90, W: t},
		{X: x, Y: 2 * rad, Dir: 0, W: t, Flat: true},
		{X: x + hookEnd*rad, Y: 2 * rad, Dir: 0, W: t},
	})
	return glyphs.Union(shifted(ch.path(), arrowWidth, 0), hook)
}

// vectorAccent draws the vector accent U+20D7: the arrow → run with vecShare
// of the head, the stroke of the macron, its ink vecWidth wide from the
// origin, standing on the line of the accents, with an advance equal to its
// ink.
func vectorAccent(r *roman) (outline.Path, float64) {
	ch, t, y := vecFrame(r)
	return withSides(shifted(vecArrow(ch, t, vecWidth), 0, y).Round(), 0, 0)
}

// vecFrame returns the head and the stroke of the vector accent and the
// height of its shaft, at which its ink stands on the line of the accents.
func vecFrame(r *roman) (ch chevron, t, y float64) {
	p := r.p.Punct
	macron, _ := r.get("uni0304")
	_, my0, _, my1 := macron.BBox()
	t = math.Round(my1 - my0)
	a := newArrowParts(r)
	ch = chevronOfLength(vecShare*headHeight/2, vecShare*headLength, a.single.a1-a.single.a0,
		math.Round(vecShare*p.MathVertexFlat), math.Round(barbShare*t))
	_, y0, _, _ := vecArrow(ch, t, vecWidth).BBox()
	return ch, t, math.Round(my0 - y0)
}

// vecArrow returns the ink of the vector accent w wide from the origin, its
// shaft on y = 0.
func vecArrow(ch chevron, t, w float64) outline.Path {
	parts := []outline.Path{shifted(ch.path(), w, 0)}
	parts = append(parts, outline.Path{box(0, -t/2, w+ch.shaftEnd(0, t/2), t/2)})
	return glyphs.Union(parts...)
}

// The arrows, in the order of the font.
var arrowDerivations = []derivation{
	{"uni2192", 0x2192, Rel, "minus", arrowSign(plainArrow(arrowWidth, false, false), false, 0)},
	{"uni2190", 0x2190, Rel, "minus", arrowSign(plainArrow(arrowWidth, false, false), true, 0)},
	{"uni2194", 0x2194, Rel, "minus", arrowSign(plainArrow(arrowWidth, false, true), false, 0)},
	{"uni2191", 0x2191, Rel, "minus", arrowSign(plainArrow(arrowWidth, false, false), false, 1)},
	{"uni2193", 0x2193, Rel, "minus", arrowSign(plainArrow(arrowWidth, false, false), false, -1)},
	{"uni2195", 0x2195, Rel, "minus", arrowSign(plainArrow(arrowWidth, false, true), false, 1)},
	{"uni21D2", 0x21D2, Rel, "minus", arrowSign(plainArrow(arrowWidth, true, false), false, 0)},
	{"uni21D0", 0x21D0, Rel, "minus", arrowSign(plainArrow(arrowWidth, true, false), true, 0)},
	{"uni21D4", 0x21D4, Rel, "minus", arrowSign(plainArrow(arrowWidth, true, true), false, 0)},
	{"uni21D1", 0x21D1, Rel, "minus", arrowSign(plainArrow(arrowWidth, true, false), false, 1)},
	{"uni21D3", 0x21D3, Rel, "minus", arrowSign(plainArrow(arrowWidth, true, false), false, -1)},
	{"uni21A6", 0x21A6, Rel, "minus", arrowSign(mapsto(arrowWidth), false, 0)},
	{"uni21AA", 0x21AA, Rel, "minus", arrowSign(hookArrow, false, 0)},
	{"uni27F5", 0x27F5, Rel, "minus", arrowSign(plainArrow(longArrowWidth, false, false), true, 0)},
	{"uni27F6", 0x27F6, Rel, "minus", arrowSign(plainArrow(longArrowWidth, false, false), false, 0)},
	{"uni27F7", 0x27F7, Rel, "minus", arrowSign(plainArrow(longArrowWidth, false, true), false, 0)},
	{"uni27F8", 0x27F8, Rel, "minus", arrowSign(plainArrow(longArrowWidth, true, false), true, 0)},
	{"uni27F9", 0x27F9, Rel, "minus", arrowSign(plainArrow(longArrowWidth, true, false), false, 0)},
	{"uni27FA", 0x27FA, Rel, "minus", arrowSign(plainArrow(longArrowWidth, true, true), false, 0)},
	{"uni27FC", 0x27FC, Rel, "minus", arrowSign(mapsto(longArrowWidth), false, 0)},
	{"uni20D7", 0x20D7, Accent, "acutecomb", vectorAccent},
	{"mapstochar", MapstocharRune, Rel, "acutecomb", mapstochar},
}
