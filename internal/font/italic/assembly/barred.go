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

	"seehuhn.de/go/bedern/internal/font/italic"
)

// The barred letters ł, Ł, đ, Đ, Ð, ħ, Ħ, ŧ, Ŧ, ø and Ø, and the signs of the
// roman which the italic takes upright.
//
// A barred letter is its base with the bar of the roman (barred.go of the
// roman: the same heights in ascenders, cap heights and x-heights, the same
// angles above the horizontal and lengths in stems) set in the upright
// description of the italic letter, about the middle of the italic stem at
// the height of the bar, at the weight of the italic (the thin of the pen for
// the lowercase, the bar of the capitals for the capitals), and then sheared
// with the letter, so that a horizontal bar has its ends cut along the slant
// and the stroke of ø stands steeper by the slant.  The letter is the union
// of the two.

// A barSpec is a bar in the upright description: its middle, its horizontal
// extent, its angle above the horizontal and its thickness at right angles,
// or, for a stroke (slash), its ends' heights.
type barSpec struct {
	cx, cy, w, angle, thick float64
	slash                   bool
	y0, y1                  float64
}

// contour returns the bar as a closed contour, anticlockwise.
func (b barSpec) contour() outline.Contour {
	var pts []outline.Point
	if b.slash {
		a := b.angle * math.Pi / 180
		w := b.thick / math.Sin(a) / 2
		x := func(y float64) float64 { return b.cx + (y-b.cy)/math.Tan(a) }
		pts = []outline.Point{{X: x(b.y0) - w, Y: b.y0}, {X: x(b.y0) + w, Y: b.y0}, {X: x(b.y1) + w, Y: b.y1}, {X: x(b.y1) - w, Y: b.y1}}
	} else {
		t := math.Tan(b.angle * math.Pi / 180)
		h := b.thick / math.Cos(b.angle*math.Pi/180) / 2
		x0, x1 := b.cx-b.w/2, b.cx+b.w/2
		y0, y1 := b.cy-t*b.w/2, b.cy+t*b.w/2
		pts = []outline.Point{{X: x0, Y: y0 - h}, {X: x1, Y: y1 - h}, {X: x1, Y: y1 + h}, {X: x0, Y: y0 + h}}
	}
	c := outline.Contour{Start: pts[0]}
	for _, q := range pts[1:] {
		c.Segs = append(c.Segs, outline.Segment{P3: q})
	}
	c.Segs = append(c.Segs, outline.Segment{P3: pts[0]})
	return c
}

// barred lists the barred letters with their bases.
var barred = []struct {
	name string
	r    rune
	base string
}{
	{"oslash", 'ø', "o"}, {"Oslash", 'Ø', "O"},
	{"lslash", 'ł', "l"}, {"Lslash", 'Ł', "L"},
	{"dcroat", 'đ', "d"}, {"Dcroat", 'Đ', "D"}, {"Eth", 'Ð', "D"},
	{"hbar", 'ħ', "h"}, {"Hbar", 'Ħ', "H"},
	{"tbar", 'ŧ', "t"}, {"Tbar", 'Ŧ', "T"},
}

// bars returns the bars of the barred letter in the upright description of
// its base u.
func (p Params) bars(name string, u outline.Path) []barSpec {
	rp := design.Regular()
	st, cs := p.Lower.Stem, p.Caps.CapStem
	thin, cbar := p.Lower.Thin, rp.Caps.Bar*p.Caps.Weight
	mid := func(y float64) float64 { lo, hi := leftStroke(u, y); return (lo + hi) / 2 }
	right := func(y float64) float64 { _, hi := inkSpan(u, y-2, y+2); return hi }
	left := func(y0, y1 float64) float64 { lo, _ := inkSpan(u, y0, y1); return lo }
	x0, _, x1, _ := u.BBox()
	switch name {
	case "lslash":
		y := 0.525 * rp.Ascender
		return []barSpec{{cx: mid(y), cy: y, w: 2.4 * st, angle: 37, thick: thin}}
	case "Lslash":
		y := 0.54 * rp.CapHeight
		w := 3.69 * cs
		return []barSpec{{cx: left(0, 20) + w/2, cy: y, w: w, angle: 35, thick: cbar}}
	case "dcroat":
		y := 0.787 * rp.Ascender
		w := 3.82 * st
		return []barSpec{{cx: right(y) + 0.4*st - w/2, cy: y, w: w, thick: thin}}
	case "Dcroat", "Eth":
		y := 0.511 * rp.CapHeight
		w := 3.76 * cs
		return []barSpec{{cx: left(0, 20) + w/2, cy: y, w: w, thick: cbar}}
	case "hbar":
		y := 0.768 * rp.Ascender
		w := 3.7 * st
		lo, _ := leftStroke(u, y)
		return []barSpec{{cx: lo - 0.4*st + w/2, cy: y, w: w, thick: thin}}
	case "Hbar":
		y := 0.75 * rp.CapHeight
		lo, hi := inkSpan(u, y-150, y+150)
		return []barSpec{{cx: (lo + hi) / 2, cy: y, w: hi - lo, thick: cbar}}
	case "tbar":
		y := 0.6 * rp.XHeight
		lo, hi := inkSpan(u, rp.XHeight-p.Lower.Hair, rp.XHeight-2)
		return []barSpec{{cx: (lo + hi) / 2, cy: y, w: hi - lo, thick: p.Lower.Hair}}
	case "Tbar":
		y := 0.526 * rp.CapHeight
		w := 4.06 * cs
		return []barSpec{{cx: mid(0.5 * rp.CapHeight), cy: y, w: w, thick: cbar}}
	case "oslash":
		c := (x0 + x1) / 2
		return []barSpec{{cx: c, cy: rp.XHeight / 2, angle: 58, thick: thin, slash: true,
			y0: -rp.OvershootLower - 34, y1: rp.XHeight + rp.OvershootLower + 34}}
	case "Oslash":
		c := (x0 + x1) / 2
		return []barSpec{{cx: c, cy: rp.CapHeight / 2, angle: 53, thick: cbar, slash: true,
			y0: -rp.OvershootUpper - 30, y1: rp.CapHeight + rp.OvershootUpper + 30}}
	}
	return nil
}

// addBarred adds the barred letters to the table of bases.
func (p Params) addBarred(gs []glyphs.Glyph) ([]glyphs.Glyph, error) {
	k := p.slant()
	for _, b := range barred {
		i := slices.IndexFunc(gs, func(g glyphs.Glyph) bool { return g.Name == b.base })
		base := gs[i]
		path := append(outline.Path(nil), base.Path...)
		for _, bar := range p.bars(b.name, upright(base.Path, k)) {
			path = append(path, transformed(bar.contour(), k, 0, 0))
		}
		u, err := union(path)
		if err != nil {
			return nil, fmt.Errorf("assembly: %s: %w", b.name, err)
		}
		gs = append(gs, glyphs.Glyph{Name: b.name, Rune: b.r, Advance: base.Advance, Path: u.Round()})
	}
	return gs, nil
}

// The signs of the roman which the italic takes upright, as signs of
// mathematics.  The other signs are drawn by internal/italic/upper.
var (
	uprightSigns = []sign{
		{"multiply", '×'}, {"divide", '÷'}, {"plusminus", '±'}, {"logicalnot", '¬'}, {"brokenbar", '¦'},
		{"uni00A0", '\u00A0'}, {"uni2009", '\u2009'}, {"uni200A", '\u200A'}, {"uni202F", '\u202F'},
	}
)

// A sign is a glyph of the roman by name and code point.
type sign struct {
	name string
	r    rune
}

// addSigns adds the signs of the roman which the italic takes upright, and the
// soft hyphen as a copy of the hyphen.
func (p Params) addSigns(gs []glyphs.Glyph) []glyphs.Glyph {
	rp := design.Regular()
	for _, sg := range uprightSigns {
		path, adv, _ := glyphs.Outline(sg.name, rp)
		if sg.name == "uni00A0" && p.Space > 0 {
			adv = p.Space
		}
		gs = append(gs, glyphs.Glyph{Name: sg.name, Rune: sg.r, Advance: adv, Path: path.Round()})
	}
	i := slices.IndexFunc(gs, func(g glyphs.Glyph) bool { return g.Name == "hyphen" })
	h := gs[i]
	h.Name, h.Rune = "uni00AD", '­'
	return append(gs, h)
}

// The ligatures fi and ffi are the union of f and the dotless i, and of f, f
// and the dotless i, each set at the advance of the letters before it: the
// bar of f runs into the entry of i, and the hook of f stands where the dot
// of i would.  The union of f and l is not clean, the ball of the hook
// lying on the entry of the ascender, so that fl, ffl and the pairs of f
// with b, h, k and l are kerned instead.
var ligatures = []struct {
	name  string
	r     rune
	parts []string
}{
	{"fi", '\uFB01', []string{"f", "i"}},
	{"ffi", '\uFB03', []string{"ff", "i"}},
	{"f_j", 0, []string{"f", "j"}},
}

// addLigatures adds the ligatures, made from the fitted letters.  The drawn
// ligatures of italic (fl, ff, ffl, f_b, f_h and f_k) are among the
// letters; they get the letters they stand for, so that the feature liga
// sets them.  ffi is the drawn ff and the dotless i, and f_j the f and the
// dotless j, whose dot the ball of the hook replaces.
func addLigatures(gs []glyphs.Glyph) ([]glyphs.Glyph, error) {
	by := map[string]glyphs.Glyph{}
	for i, g := range gs {
		if parts, ok := italic.LigatureParts(g.Name); ok {
			gs[i].Parts = parts
		}
		by[g.Name] = g
	}
	for _, l := range ligatures {
		var path outline.Path
		x := 0.0
		var parts []string
		for _, part := range l.parts {
			g := by[part]
			switch part {
			case "i":
				g = by["dotlessi"]
			case "j":
				g = by["uni0237"]
			}
			if p, ok := italic.LigatureParts(part); ok {
				parts = append(parts, p...)
			} else {
				parts = append(parts, part)
			}
			path = append(path, moved(g.Path, x, 0)...)
			x += g.Advance
		}
		u, err := union(path)
		if err != nil {
			return nil, fmt.Errorf("assembly: %s: %w", l.name, err)
		}
		gs = append(gs, glyphs.Glyph{Name: l.name, Rune: l.r, Advance: x, Path: u.Round(), Parts: parts})
	}
	return gs, nil
}
