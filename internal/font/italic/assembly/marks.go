// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package assembly

import (
	"fmt"
	"math"
	"slices"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/stroke"
)

// The marks and the composed letters.
//
// The marks are those of the roman (glyphs.MarkOutline), at the roman's
// weights, sloped: every contour is sheared by the slant about the origin of
// the frame of the mark, its anchor, except the dots of the diaeresis and of
// the dot accent and the wedges of the acute and the double acute, which are
// moved by the slant at the height of their centre and keep their form.
//
// The anchors of a letter follow the rules of the roman (compose.go of the
// roman) in the upright description of the italic letter, x - y tan(slant),
// where a stem is vertical: the middle of the ink between the baseline and
// the x-height or the cap height, of the stem of l and h above the x-height,
// of the dot of i and j, of the stem of I, J and L below the cap height.  The
// anchor at the height y in the italic lies at the upright abscissa plus y
// times the slant, so that a mark stands on the centre line of the sloped
// letter extended to its height.

// rigid are the marks whose contours are moved by the slant and not sheared.
var rigid = []string{"dieresis", "dotaccent", "acute", "hungarumlaut"}

// sloped returns the mark of the roman, sloped.
func (p Params) sloped(mark, form string) outline.Path {
	m, ok := glyphs.MarkOutline(design.Regular(), mark, form)
	if !ok {
		panic(fmt.Sprintf("assembly: no mark %s in the form %s", mark, form))
	}
	k := p.slant()
	if mark == "dotbelow" {
		// The dot below is the dot of the italic, a circle of the diameter
		// of the dot of i and j, with its top where the roman's dot has it
		// and its centre moved by the slant as the roman's is.
		x0, _, x1, y1 := m.BBox()
		cy := y1 - p.Lower.Dot/2
		return outline.Path{circle((x0+x1)/2+k*cy, cy, p.Lower.Dot)}
	}
	var out outline.Path
	for _, c := range m {
		if len(m) == 1 && c.SignedArea() < 0 {
			c = reversed(c)
		}
		if slices.Contains(rigid, mark) {
			_, y0, _, y1 := outline.Path{c}.BBox()
			out = append(out, transformed(c, 0, k*(y0+y1)/2, 0))
		} else {
			out = append(out, transformed(c, k, 0, 0))
		}
	}
	return out
}

// transformed returns the contour under (x, y) -> (x + k y + dx, y + dy).
func transformed(c outline.Contour, k, dx, dy float64) outline.Contour {
	f := func(q outline.Point) outline.Point { return outline.Point{X: q.X + k*q.Y + dx, Y: q.Y + dy} }
	out := outline.Contour{Start: f(c.Start)}
	for _, s := range c.Segs {
		out.Segs = append(out.Segs, outline.Segment{IsCurve: s.IsCurve, P1: f(s.P1), P2: f(s.P2), P3: f(s.P3)})
	}
	return out
}

// moved returns the path moved by (dx, dy).
func moved(path outline.Path, dx, dy float64) outline.Path {
	out := make(outline.Path, len(path))
	for i, c := range path {
		out[i] = transformed(c, 0, dx, dy)
	}
	return out
}

// upright returns the path in the upright description, x - k y.
func upright(path outline.Path, k float64) outline.Path {
	out := make(outline.Path, len(path))
	for i, c := range path {
		out[i] = transformed(c, -k, 0, 0)
	}
	return out
}

// reversed returns the contour traversed the other way.
func reversed(c outline.Contour) outline.Contour {
	n := len(c.Segs)
	pts := make([]outline.Point, n+1)
	pts[0] = c.Start
	for i, s := range c.Segs {
		pts[i+1] = s.P3
	}
	out := outline.Contour{Start: pts[n]}
	for i := n - 1; i >= 0; i-- {
		s := c.Segs[i]
		if s.IsCurve {
			out.Segs = append(out.Segs, outline.Segment{IsCurve: true, P1: s.P2, P2: s.P1, P3: pts[i]})
		} else {
			out.Segs = append(out.Segs, outline.Segment{P3: pts[i]})
		}
	}
	return out
}

// inkSpan returns the least and the greatest abscissa of the outline between
// the heights y0 and y1, from the flattened contours and their crossings of
// the two heights.
func inkSpan(path outline.Path, y0, y1 float64) (lo, hi float64) {
	lo, hi = math.Inf(1), math.Inf(-1)
	for _, ct := range path {
		pts := ct.Flatten()
		n := len(pts)
		for i := range n {
			a, b := pts[i], pts[(i+1)%n]
			if a.Y >= y0 && a.Y <= y1 {
				lo, hi = math.Min(lo, a.X), math.Max(hi, a.X)
			}
			for _, y := range []float64{y0, y1} {
				if (a.Y < y) != (b.Y < y) {
					x := a.X + (y-a.Y)/(b.Y-a.Y)*(b.X-a.X)
					lo, hi = math.Min(lo, x), math.Max(hi, x)
				}
			}
		}
	}
	return lo, hi
}

// lowestInk returns the lowest height at which the vertical through x meets
// the outline.
func lowestInk(path outline.Path, x float64) float64 {
	low := math.Inf(1)
	for _, ct := range path {
		pts := ct.Flatten()
		n := len(pts)
		for i := range n {
			a, b := pts[i], pts[(i+1)%n]
			if (a.X < x) != (b.X < x) {
				low = math.Min(low, a.Y+(x-a.X)/(b.X-a.X)*(b.Y-a.Y))
			}
		}
	}
	return low
}

// leftStroke returns the edges of the leftmost stroke at the height y.
func leftStroke(path outline.Path, y float64) (lo, hi float64) {
	xs := path.CrossingsX(y)
	slices.Sort(xs)
	if len(xs) < 2 {
		return math.NaN(), math.NaN()
	}
	return xs[0], xs[1]
}

var ascenders = []string{"b", "d", "h", "k", "l"}

func isCapital(name string) bool { return len(name) == 1 && name[0] >= 'A' && name[0] <= 'Z' }

func isLetter(name string) bool {
	return len(name) == 1 && (isCapital(name) || name[0] >= 'a' && name[0] <= 'z') || name == "dotlessi" || name == "uni0237"
}

// anchorer computes the anchors of the bases of one table.
type anchorer struct {
	p    Params
	rp   design.Params
	k    float64
	dots map[string]float64 // the upright abscissa of the dot of i and j, by base
}

// height returns the height of the letter for the rules: the x-height or the
// cap height.
func (a *anchorer) height(name string) float64 {
	if isCapital(name) {
		return a.rp.CapHeight
	}
	return a.rp.XHeight
}

// topU returns the upright abscissa of the top anchor.
func (a *anchorer) topU(name string, u outline.Path) float64 {
	h := a.height(name)
	if x, ok := a.dots[name]; ok {
		return x
	}
	switch name {
	case "l", "h":
		lo, hi := inkSpan(u, a.rp.XHeight+60, a.rp.XHeight+120)
		return (lo + hi) / 2
	case "I", "J", "L":
		lo, hi := inkSpan(u, h-150, h-100)
		return (lo + hi) / 2
	}
	lo, hi := inkSpan(u, 0, h)
	return (lo + hi) / 2
}

// bottomU returns the upright abscissa of the bottom anchor.
func (a *anchorer) bottomU(name string, u outline.Path) float64 {
	h := a.height(name)
	switch name {
	case "t", "l", "r":
		lo, hi := leftStroke(u, 0.5*h)
		return (lo + hi) / 2
	case "T", "L":
		lo, hi := inkSpan(u, 0.5*h, 0.6*h)
		return (lo + hi) / 2
	}
	lo, hi := inkSpan(u, 0, h)
	return (lo + hi) / 2
}

// ogonekU returns the upright abscissa at which the ogonek leaves the base.
func (a *anchorer) ogonekU(name string, u outline.Path) float64 {
	h := a.height(name)
	switch name {
	case "a", "u":
		_, hi := inkSpan(u, 0.3*h, 0.5*h)
		return hi - a.p.Lower.Stem/2
	case "i", "I":
		lo, hi := inkSpan(u, 0.3*h, 0.5*h)
		return (lo + hi) / 2
	case "A":
		_, hi := inkSpan(u, 0, 0.02*h)
		return hi - a.rp.Caps.SerifJut*a.p.Caps.Width
	}
	lo, hi := inkSpan(u, 0, h)
	return lo + 0.68*(hi-lo)
}

// topLine returns the height at which the frame of a mark above the letter
// stands, and the height of the top anchor for the combining marks.
func (a *anchorer) topLine(name string) (frame, anchor float64) {
	acc := a.rp.Acc
	switch {
	case isCapital(name):
		return a.rp.CapHeight + acc.GapCap, a.rp.CapHeight + acc.GapCap - acc.Gap
	case slices.Contains(ascenders, name):
		return a.rp.Ascender + acc.GapCap, a.rp.Ascender + acc.GapCap - acc.Gap
	}
	return a.rp.XHeight + acc.Gap, a.rp.XHeight
}

// anchors returns the anchors of a base for the combining marks.
func (a *anchorer) anchors(name string, path outline.Path) []glyphs.Anchor {
	u := upright(path, a.k)
	_, y := a.topLine(name)
	return []glyphs.Anchor{
		{Class: "top", X: math.Round(a.topU(name, u) + a.k*y), Y: math.Round(y)},
		{Class: "bottom", X: math.Round(a.bottomU(name, u)), Y: 0},
		{Class: "ogonek", X: math.Round(a.ogonekU(name, u)), Y: 0},
	}
}

// place returns the offset of the frame of the mark in the composed letter.
func (a *anchorer) place(c glyphs.Composition, base outline.Path) (dx, dy float64) {
	u := upright(base, a.k)
	acc := a.rp.Acc
	switch c.Mark {
	case "apostrophe":
		_, _, _, top := base.BBox()
		x := 0.0
		switch c.Base {
		case "L":
			top = a.rp.CapHeight
			_, x = inkSpan(base, top-130, top-110)
		case "t":
			_, x = inkSpan(base, top-50, top-30)
			h := 0.0
			if m, ok := glyphs.MarkOutline(a.rp, "apostrophe", "cap"); ok {
				_, y0, _, y1 := m.BBox()
				h = y1 - y0
			}
			top = math.Max(top, a.rp.XHeight+acc.Gap/2+h)
		default:
			top = a.rp.Ascender
			_, x = inkSpan(base, top-110, top-80)
		}
		// The mark keeps clear of a head serif on the right of the stem.
		_, serif := inkSpan(base, top-70, top)
		x = math.Max(x+acc.ApoGap, serif+design.ClearanceFloor(a.rp)+4)
		return math.Round(x), math.Round(top)
	case "cedilla", "ogonek":
		x := a.bottomU(c.Base, u)
		if c.Mark == "ogonek" {
			x = a.ogonekU(c.Base, u)
		}
		y := lowestInk(base, x)
		if !c.Joined {
			_, y0, _, _ := base.BBox()
			y = math.Min(y, y0) - acc.CedOverlap - acc.CedApart
		}
		return math.Round(x + a.k*y), math.Round(y)
	case "commabelow", "dotbelow":
		return math.Round(a.bottomU(c.Base, u)), 0
	}
	y, _ := a.topLine(c.Base)
	return math.Round(a.topU(c.Base, u) + a.k*y), math.Round(y)
}

// like returns the glyph whose kerning a composed or dotless letter takes.
func like(name string, comps map[string]glyphs.Composition) string {
	switch name {
	case "dotlessi":
		return "i"
	case "uni0237":
		return "j"
	}
	if c, ok := comps[name]; ok {
		if l := like(c.Base, comps); l != "" {
			return l
		}
		return c.Base
	}
	return ""
}

// compose adds the spacing accents, the combining marks and the composed
// letters to the bases, and the anchors of the bases.
func compose(p Params, gs []glyphs.Glyph) ([]glyphs.Glyph, error) {
	rp := design.Regular()
	a := &anchorer{p: p, rp: rp, k: p.slant(), dots: map[string]float64{}}
	index := map[string]int{}
	for i, g := range gs {
		index[g.Name] = i
	}
	for _, d := range []struct{ base, from string }{{"i", "i"}, {"dotlessi", "i"}, {"j", "j"}, {"uni0237", "j"}} {
		_, dot := withoutDot(gs[index[d.from]].Path, rp.XHeight)
		x0, y0, x1, y1 := dot.BBox()
		a.dots[d.base] = (x0+x1)/2 - a.k*(y0+y1)/2
	}
	for i, g := range gs {
		if isLetter(g.Name) {
			gs[i].Anchors = a.anchors(g.Name, g.Path)
		}
	}
	nAdv := gs[index["n"]].Advance
	cx := -math.Round(nAdv / 2)
	acc := rp.Acc
	for _, m := range glyphs.MarkGlyphs() {
		mk := p.sloped(m.Mark, "lower")
		below := m.Mark == "cedilla" || m.Mark == "ogonek" || m.Mark == "commabelow" || m.Mark == "dotbelow"
		if !m.Combining {
			x0, _, x1, _ := mk.BBox()
			dx := math.Round(acc.MarkSide - x0)
			y := rp.XHeight + acc.Gap
			if below {
				y = 0
			}
			gs = append(gs, glyphs.Glyph{Name: m.Name, Rune: m.Rune,
				Advance: math.Ceil(x1 + dx + acc.MarkSide - 1e-6), Path: moved(mk, dx, y).Round()})
			continue
		}
		class := "top"
		switch m.Mark {
		case "cedilla", "commabelow", "dotbelow":
			class = "bottom"
		case "ogonek":
			class = "ogonek"
		}
		an := glyphs.Anchor{Class: class, X: cx, Y: rp.XHeight}
		path := moved(mk, cx+math.Round(a.k*acc.Gap), rp.XHeight+acc.Gap)
		if below {
			an.Y = 0
			path = moved(mk, cx, 0)
		}
		gs = append(gs, glyphs.Glyph{Name: m.Name, Rune: m.Rune, Path: path.Round(), Anchors: []glyphs.Anchor{an}})
	}
	gs = p.addTS1Marks(gs)
	gs = p.addCaseMarks(gs)
	comps := map[string]glyphs.Composition{}
	for _, c := range glyphs.Compositions() {
		comps[c.Name] = c
	}
	for _, c := range glyphs.Compositions() {
		bi, ok := index[c.Base]
		if !ok {
			return nil, fmt.Errorf("assembly: no base %s for %s", c.Base, c.Name)
		}
		b := gs[bi]
		mk := p.sloped(c.Mark, c.Form)
		dx, dy := a.place(c, b.Path)
		var path outline.Path
		if c.Joined {
			// Edges which run along each other are moved apart by a
			// fraction of a unit, which the rounding removes.
			var u outline.Path
			var err error
			for _, e := range []float64{0, 0.3, -0.3, 0.6} {
				u, err = union(append(append(outline.Path(nil), b.Path...), moved(mk, dx+e, dy+e)...))
				if err == nil {
					break
				}
			}
			if err != nil {
				return nil, fmt.Errorf("assembly: %s: %w", c.Name, err)
			}
			path = u.Round()
		} else {
			path = append(append(outline.Path(nil), b.Path...), moved(mk.Round(), dx, dy)...)
		}
		gs = append(gs, glyphs.Glyph{
			Name: c.Name, Rune: c.Rune, Advance: b.Advance, Path: path,
			Base: c.Base, Like: like(c.Name, comps), Marks: moved(mk.Round(), dx, dy),
		})
	}
	for i := range gs {
		if gs[i].Like == "" {
			gs[i].Like = like(gs[i].Name, comps)
		}
	}
	return gs, nil
}

// union returns the outline of the ink which the contours cover together.
func union(path outline.Path) (outline.Path, error) {
	var cs []stroke.Chain
	for _, c := range path {
		var ch stroke.Chain
		cur := c.Start
		for _, s := range c.Segs {
			cb := stroke.Cubic{cur, s.P1, s.P2, s.P3}
			if !s.IsCurve {
				d := outline.Point{X: (s.P3.X - cur.X) / 3, Y: (s.P3.Y - cur.Y) / 3}
				cb = stroke.Cubic{cur, {X: cur.X + d.X, Y: cur.Y + d.Y}, {X: cur.X + 2*d.X, Y: cur.Y + 2*d.Y}, s.P3}
			}
			if cb[0] != cb[3] || s.IsCurve {
				ch = append(ch, stroke.Piece{C: cb})
			}
			cur = s.P3
		}
		if cur != c.Start {
			d := outline.Point{X: (c.Start.X - cur.X) / 3, Y: (c.Start.Y - cur.Y) / 3}
			ch = append(ch, stroke.Piece{C: stroke.Cubic{cur, {X: cur.X + d.X, Y: cur.Y + d.Y}, {X: cur.X + 2*d.X, Y: cur.Y + 2*d.Y}, c.Start}})
		}
		cs = append(cs, ch)
	}
	u, err := stroke.Union(cs)
	if err != nil {
		return nil, err
	}
	var out outline.Path
	for _, ch := range u {
		ct := outline.Contour{Start: ch[0].C[0]}
		for i, pc := range ch {
			end := pc.C[3]
			if i == len(ch)-1 {
				end = ch[0].C[0]
			}
			if pc.C.IsStraight() {
				ct.Segs = append(ct.Segs, outline.Segment{P3: end})
			} else {
				ct.Segs = append(ct.Segs, outline.Segment{IsCurve: true, P1: pc.C[1], P2: pc.C[2], P3: end})
			}
		}
		out = append(out, ct)
	}
	return out, nil
}
