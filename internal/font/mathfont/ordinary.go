// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/font/italic"
	"seehuhn.de/go/bedern/internal/font/italic/assembly"
	"seehuhn.de/go/bedern/internal/stroke"
)

// The ordinary symbols which are drawn as signs: ∞, ∀, ∅ and ∇.  ∞ and ∅
// are monoline with the rule of the signs; ∀ is drawn without serifs, as ∃
// is, with its diagonals a little heavier than the rule; ∇ is a triangle with
// the weights of the capitals, its heavy stroke on the left, drawn and not
// turned from Δ.
const (
	// infWidth and infHeight are the ink of ∞; its stroke is the rule.
	infWidth  = 800.0
	infHeight = 400.0
	// infCross is the angle, in degrees, at which the centre line of ∞
	// crosses itself in the middle, from the level.
	infCross = 48.0
	// infTop is the distance of the top and the bottom of a loop of ∞ from
	// the middle over the half width of the centre line.
	infTop = 0.58
	// forallWidth is the ink width of ∀ over the cap height; forallStroke its
	// diagonals over the rule; forallBar the height of the middle of its bar
	// over its height.
	forallWidth  = 0.83
	forallStroke = 1.10
	forallBar    = 0.62
	// emptyRing is the outer diameter of the ring of ∅ over the height of
	// the figures; emptySlash the angle of its slash from the upright, in
	// degrees; emptyReach how far the slash reaches beyond the ring above and
	// below; emptySlashWeight the slash over the rule.
	emptyRing        = 0.94
	emptySlash       = 30.0
	emptyReach       = 40.0
	emptySlashWeight = 0.90
	// nablaWidth is the ink width of ∇ over its height; its left diagonal is
	// nablaThick of the capital stem and its right diagonal nablaThin of the
	// capital hairline, and its top bar is the bar of T.
	nablaWidth = 0.88
	nablaThick = 0.62
	nablaThin  = 1.00
)

// ordinaryDerivations are the ordinary symbols drawn as signs.
var ordinaryDerivations = []derivation{
	{"uni221E", 0x221E, Ord, "plus", sign(infinity)},
	{"uni2200", 0x2200, Ord, "V", forall},
	{"uni2205", 0x2205, Ord, "plus", sign(emptyset)},
	{"uni2207", 0x2207, Ord, "V", nabla},
	{"uni2113", 0x2113, Ord, "f", italicSymbol("uni2113", ellSides)},
	{"uni2202", 0x2202, Ord, "o", italicSymbol("uni2202", partialSides)},
}

// ellSides and partialSides are the side spaces of ℓ and ∂, left and right.
var (
	ellSides     = [2]float64{30, 22}
	partialSides = [2]float64{45, 48}
)

// italicSymbol returns ℓ or ∂, drawn with the pen of the italic
// (bedern/internal/font/italic, mathsigns.go), with the side spaces given.
func italicSymbol(name string, sides [2]float64) func(r *roman) (outline.Path, float64) {
	build := sync.OnceValue(func() *italic.Glyph {
		g, err := italic.BuildSymbol(italic.Default(), name)
		if err != nil {
			panic(err)
		}
		return g
	})
	return func(r *roman) (outline.Path, float64) {
		// The symbol is built once.
		return withSides(build().Path, sides[0], sides[1])
	}
}

func init() {
	derivations = append(derivations, ordinaryDerivations...)
}

// infinity draws ∞: one closed stroke of the rule through a crossing in the
// middle, round a loop on the right and through the crossing again round a
// loop on the left, the extremes of each loop on its centre line the nodes,
// the stations fitted to make the ink infWidth wide and infHeight high.
// The outline is the union of the edges of the stroke, whose loops wind in
// opposite senses.
func infinity(r *roman) outline.Path {
	t := r.p.Punct.MathThick
	w, h := infWidth-t, infHeight-t // the extent of the centre line
	var path outline.Path
	for range 8 {
		xc := infTop * w / 2 // the top and the bottom of a loop
		nd := func(x, y, dir float64) stroke.Node {
			return stroke.Node{P: pt(x, y), Dir: dir, K: stroke.Auto, W: t, DW: 0}
		}
		s := stroke.Stroke{Closed: true, Nodes: []stroke.Node{
			{P: pt(0, 0), Dir: -infCross, K: 0, W: t, DW: 0},
			nd(xc, -h/2, 0),
			nd(w/2, 0, 90),
			nd(xc, h/2, 180),
			{P: pt(0, 0), Dir: 180 + infCross, K: 0, W: t, DW: 0},
			nd(-xc, -h/2, 180),
			nd(-w/2, 0, 90),
			nd(-xc, h/2, 0),
		}}
		res, err := s.Build()
		if err != nil {
			panic(err)
		}
		path = glyphs.Union(res.Path())
		x0, y0, x1, y1 := path.BBox()
		w += infWidth - (x1 - x0)
		h += infHeight - (y1 - y0)
	}
	return path
}

// forall draws ∀: two straight strokes forallStroke of the rule across from
// the top corners of its ink to a vertex at the baseline, cut flat
// MathVertexFlat wide as the vertex of <, the free ends cut level at the cap
// height, and a bar of the rule between them; with the side spaces of V.
func forall(r *roman) (outline.Path, float64) {
	p := r.p
	t, f := p.Punct.MathThick, p.Punct.MathVertexFlat
	s := math.Round(forallStroke * t)
	h := p.CapHeight
	w := math.Round(forallWidth * h)
	xm := w / 2
	a := math.Atan2(h, xm-f/2)
	c := s / math.Sin(a) // the width of a stroke along a level line
	yv := (c - f/2) * math.Tan(a)
	vee := outline.Path{polygon(
		pt(0, h), pt(xm-f/2, 0), pt(xm+f/2, 0), pt(w, h),
		pt(w-c, h), pt(xm, yv), pt(c, h))}
	yb := math.Round(forallBar * h)
	// The bar runs between the centre lines of the strokes at its middle.
	xb := centreAt(xm, f, c, a, yb)
	bar := outline.Path{box(xb, yb-t/2, w-xb, yb+t/2)}
	path := glyphs.Union(vee, bar)
	v, adv := r.get("V")
	x0, _, x1, _ := v.BBox()
	return withSides(path, x0, adv-x1)
}

// centreAt returns the abscissa of the centre line of the left stroke of ∀
// at the height y: the stroke runs from the vertex up to the left top corner.
func centreAt(xm, f, c, a, y float64) float64 {
	// The outer edge passes (xm - f/2, 0) and rises at the angle a; the
	// centre line lies c/2 to its right along a level line.
	return xm - f/2 - y/math.Tan(a) + c/2
}

// emptyset draws ∅: a ring of the rule emptyRing of the height of the
// figures across, and a slash emptySlashWeight of the rule across, leaning
// by emptySlash from the upright through the middle of the ring and
// reaching emptyReach beyond it above and below, cut level.
func emptyset(r *roman) outline.Path {
	p := r.p
	t := p.Punct.MathThick
	rad := math.Round(emptyRing*p.Figures.Height) / 2
	a := emptySlash * math.Pi / 180
	sw := math.Round(emptySlashWeight*t) / math.Cos(a) // the slash along a level line
	h := rad + emptyReach
	run := h * math.Tan(a)
	slash := outline.Path{polygon(pt(-run-sw/2, -h), pt(-run+sw/2, -h), pt(run+sw/2, h), pt(run-sw/2, h))}
	return glyphs.Union(ring(pt(0, 0), rad, t), slash)
}

// nabla draws ∇: a triangle standing on its point, the cap height high and
// nablaWidth of that wide, its left side nablaThick of the capital stem
// across, its right side nablaThin of the capital hairline, its top the
// bar of T, and its point cut flat MathVertexFlat wide; with MathSide on
// each side.
func nabla(r *roman) (outline.Path, float64) {
	p := r.p
	h := p.CapHeight
	w := math.Round(nablaWidth * h)
	f := p.Punct.MathVertexFlat
	sl := math.Round(nablaThick * p.Caps.Stem)
	sr := math.Round(nablaThin * p.Caps.Bar)
	tb := p.Caps.TBar
	xm := w / 2
	outer := polygon(pt(0, h), pt(xm-f/2, 0), pt(xm+f/2, 0), pt(w, h))
	// The inner edges: the left edge moved sl to its right across, the
	// right edge sr to its left, the top tb down.
	al := math.Atan2(h, xm-f/2) // the angle of the left edge to the level
	cl := sl / math.Sin(al)     // its stroke along a level line
	cr := sr / math.Sin(al)
	yt := h - tb
	// The left inner edge passes (cl, h) parallel to the outer edge, the
	// right one (w - cr, h); they meet at the inner point.
	k := 1 / math.Tan(al) // dx per unit of height
	left := func(y float64) float64 { return cl + (h-y)*k }
	right := func(y float64) float64 { return w - cr - (h-y)*k }
	yi := h - (w-cr-cl)/(2*k) // where they meet
	counter := polygon(pt(left(yt), yt), pt(right(yt), yt), pt(left(yi), yi))
	path := outline.Path{outer, counter}
	return withSides(path, p.Punct.MathSide, p.Punct.MathSide)
}

// partialItalic returns the maths italic ∂ U+1D715: the construction of the
// upright ∂ drawn at the slant of the italic with the pen held level
// (bedern/internal/font/italic, mathsigns.go), with the side spaces of the upright, and
// the italics correction and the top accent attachment of the maths italic
// (italic.go), with d as its letter.
func partialItalic() Glyph {
	g, err := italic.BuildSymbol(italic.Default(), "u1D715")
	if err != nil {
		panic(err)
	}
	path, adv := withSides(g.Path, partialSides[0], partialSides[1])
	out := Glyph{Name: "u1D715", Rune: 0x1D715, Advance: math.Round(adv), Path: path, Class: Ord, Like: "o"}
	out.Italic = assembly.ItalicCorrection(glyphs.Glyph{Name: out.Name, Rune: out.Rune, Advance: out.Advance, Path: path.Round()}, icPad)
	out.Attach, out.Attached = accentAttachment("d", path), true
	return out
}
