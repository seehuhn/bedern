// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The arrows of text: ← → ↑ ↓.
//
// An arrow is a shaft with a head.  The shaft is a bar of the stroke of the
// mathematical signs (Punct.MathThick).  The head is a chevron of two barbs
// whose edges are concentric circular arcs, concave on the outside: a barb
// leaves the tip at a shallow angle and opens by arrowBend degrees to its
// end, where it is cut along its radius.  The tip is cut flat,
// Punct.MathVertexFlat high, and the inner edges meet at a crotch on the
// middle line of the shaft.  The shaft runs into the head and is joined to it.
//
// The level arrows have their shaft on the maths axis (Punct.MathAxis), the
// advance of the em dash and the side space of the mathematical signs
// (Punct.MathSide), and ← is the mirror image of →.  The upright arrows are
// the level arrow turned by a quarter, with the middle of the ink on the
// maths axis, centred in the advance of the figures.

// Shape constants of the arrows.
const (
	arrowHeadHeight = 372.0 // the height of the head of →
	arrowHeadLength = 238.0 // the length of the head from the tip to the inner corner of the end of a barb
	arrowBarb       = 0.88  // the thickness of a barb, in strokes of the shaft
	arrowBend       = 30.0  // the turn of the edges of a barb from the tip to its end, degrees
)

// An arrowHead is the head of an arrow in its own frame: the tip at the
// origin, pointing to the right, the middle line of the shaft at y = 0.
type arrowHead struct {
	flat, barb float64       // the flat at the tip, the thickness of a barb
	c          outline.Point // the centre of the edges of the upper barb
	r          float64       // the radius of the outer edge of the upper barb
	t0, t1     float64       // the directions from the centre to the outer edge at the tip and at the end
}

// newArrowHead returns the head whose outer edge reaches the height h above
// the middle line and lies at the angles a0 at the tip and a1 at the end
// (radians) to it, with the flat and the barb given.
func newArrowHead(h, a0, a1, flat, barb float64) arrowHead {
	r := (h - flat/2) / (math.Cos(a0) - math.Cos(a1))
	c := outline.Point{X: r * math.Sin(a0), Y: flat/2 + r*math.Cos(a0)}
	t0 := math.Atan2(-math.Cos(a0), -math.Sin(a0))
	return arrowHead{flat: flat, barb: barb, c: c, r: r, t0: t0, t1: t0 - (a1 - a0)}
}

// arrowHeadOfLength returns the head of the half height h whose edges turn by
// bend (radians) from the tip to the end and which is l long from the tip to
// the inner corner of the end of a barb.
func arrowHeadOfLength(h, l, bend, flat, barb float64) arrowHead {
	lo, hi := -0.2, math.Pi/2-bend
	var hd arrowHead
	for range 60 {
		a := (lo + hi) / 2
		hd = newArrowHead(h, a, a+bend, flat, barb)
		if hd.length() > l {
			lo = a
		} else {
			hi = a
		}
	}
	if math.Abs(hd.length()-l) > 0.5 {
		panic(fmt.Sprintf("glyphs: no arrow head of the half height %.0f with a bend of %.0f degrees is %.0f long", h, bend*180/math.Pi, l))
	}
	return hd
}

// length returns the distance from the tip to the inner corner of the end of
// a barb.
func (hd arrowHead) length() float64 {
	return -(hd.c.X + (hd.r+hd.barb)*math.Cos(hd.t1))
}

// inner returns the abscissa of the inner edge of the upper barb at the
// height y.
func (hd arrowHead) inner(y float64) float64 {
	ri := hd.r + hd.barb
	return hd.c.X - math.Sqrt(ri*ri-(y-hd.c.Y)*(y-hd.c.Y))
}

// outer returns the abscissa of the outer edge of the upper barb, or of the
// flat of the tip, at the height y.
func (hd arrowHead) outer(y float64) float64 {
	if y <= hd.flat/2 {
		return 0
	}
	return hd.c.X - math.Sqrt(hd.r*hd.r-(y-hd.c.Y)*(y-hd.c.Y))
}

// chain returns the outline of the head, counter-clockwise from the lower
// corner of the flat of the tip.
func (hd arrowHead) chain() chain {
	ri := hd.r + hd.barb
	on := func(r, t float64) outline.Point {
		return outline.Point{X: hd.c.X + r*math.Cos(t), Y: hd.c.Y + r*math.Sin(t)}
	}
	// The direction from the centre to the crotch, where the inner edge
	// meets the middle line.
	tv := -math.Pi - math.Asin(-hd.c.Y/ri)
	outerArc := arcPieces(hd.c, hd.r, hd.t0, hd.t1)
	outerArc[0].P0 = outline.Point{X: 0, Y: hd.flat / 2}
	innerArc := arcPieces(hd.c, ri, hd.t1, tv)
	innerArc[len(innerArc)-1].P3 = outline.Point{X: innerArc[len(innerArc)-1].P3.X, Y: 0}
	upper := join(outerArc, lineChain(on(hd.r, hd.t1), on(ri, hd.t1)), innerArc)
	flat := lineChain(outline.Point{X: 0, Y: -hd.flat / 2}, outline.Point{X: 0, Y: hd.flat / 2})
	return join(flat, upper, upper.flipY(0))
}

// shaftEnd returns the abscissa at which a shaft t thick ends inside the
// head: halfway between the inner edge and the flat of the tip at a quarter
// of the thickness above the middle line.
func (hd arrowHead) shaftEnd(t float64) float64 {
	x := (hd.inner(t/4) + hd.outer(t/4)) / 2
	for _, y := range []float64{0, t / 2} {
		if !(hd.inner(y) < x && x < hd.outer(y)) {
			panic(fmt.Sprintf("glyphs: an arrow shaft %.0f thick does not end inside the head", t))
		}
	}
	return x
}

// textArrowHead returns the head of the arrows of text.
func textArrowHead(p design.Params) arrowHead {
	q := p.Punct
	return arrowHeadOfLength(arrowHeadHeight/2, arrowHeadLength, degrees(arrowBend), q.MathVertexFlat, math.Round(arrowBarb*q.MathThick))
}

// arrowParts returns the shaft and the head of an arrow which points to the
// right, width long, with its tail at the origin and its shaft on y = 0.
func arrowParts(p design.Params, width float64) (shaft, head chain) {
	t := p.Punct.MathThick
	hd := textArrowHead(p)
	return rectChain(0, -t/2, width+hd.shaftEnd(t), t/2), hd.chain().shifted(width, 0)
}

// arrowWidth returns the ink width of the arrows of text and the advance of
// the level ones: the advance of the em dash, and the ink which leaves the
// side space of the mathematical signs on each side.
func arrowWidth(p design.Params) (width, advance float64) {
	_, advance = emdash(p)
	return advance - 2*p.Punct.MathSide, advance
}

// levelArrow draws → or, mirrored, ←.
func levelArrow(p design.Params, mirror bool) (outline.Path, float64) {
	width, adv := arrowWidth(p)
	shaft, head := arrowParts(p, width)
	path := shiftInk(chainsPath(union(shaft, head)...), 0, p.Punct.MathAxis)
	if mirror {
		path = mirrorInk(path, 0)
	}
	return placeCentred(path, adv)
}

// uprightArrow draws ↑ or, with down set, ↓.
func uprightArrow(p design.Params, down bool) (outline.Path, float64) {
	width, _ := arrowWidth(p)
	shaft, head := arrowParts(p, width)
	turn := func(q outline.Point) outline.Point { return outline.Point{X: -q.Y, Y: q.X} }
	if down {
		turn = func(q outline.Point) outline.Point { return outline.Point{X: q.Y, Y: -q.X} }
	}
	path := chainsPath(union(shaft.mapPoints(turn), head.mapPoints(turn))...)
	_, y0, _, y1 := path.BBox()
	path = shiftInk(path, 0, math.Round(p.Punct.MathAxis-(y0+y1)/2))
	return placeCentred(path, p.Figures.Advance)
}

func arrowright(p design.Params) (outline.Path, float64) { return levelArrow(p, false) }
func arrowleft(p design.Params) (outline.Path, float64)  { return levelArrow(p, true) }
func arrowup(p design.Params) (outline.Path, float64)    { return uprightArrow(p, false) }
func arrowdown(p design.Params) (outline.Path, float64)  { return uprightArrow(p, true) }
