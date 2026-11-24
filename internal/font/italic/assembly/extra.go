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

// The glyphs which the italic makes of its other glyphs or takes from the
// roman: the ordinal indicators ª and º, the two letters S and the form of i
// of the feature locl, the spaces of fixed width and the compound word mark,
// and the dot below.

// The ordinal indicators are made as the roman makes them (signs.go of the
// roman): the letter, drawn with strokes ordWeight of those of a and o and
// scaled by ordScale, so that its strokes are ordWeight of theirs, over an
// underline of the thickness of the bar of e, ordGap below the letter, with
// its top at ordBottom plus that thickness, and ordSide of white on either
// side.  The roman's numbers hold: the letters are those of the italic, and
// the sloped letter keeps its slant under the scaling.
const (
	ordScale  = 0.7
	ordWeight = 0.76
	ordBottom = 331.0
	ordGap    = 55.0
	ordSide   = 30.0
)

// addOrdinals adds ª and º to the table of bases.
func (p Params) addOrdinals(gs []glyphs.Glyph) ([]glyphs.Glyph, error) {
	q := p.Lower.WithStem(p.Lower.Stem * ordWeight / ordScale)
	for _, o := range []struct {
		name   string
		r      rune
		letter string
	}{{"ordfeminine", 'ª', "a"}, {"ordmasculine", 'º', "o"}} {
		l, err := italic.Build(q, o.letter)
		if err != nil {
			return nil, fmt.Errorf("assembly: %s: %w", o.name, err)
		}
		if len(l.Warnings) > 0 {
			return nil, fmt.Errorf("assembly: %s: %v", o.name, l.Warnings)
		}
		x0, y0, x1, _ := l.Path.BBox()
		bar := p.Lower.EBarW
		path := moved(scaled(moved(l.Path, -(x0+x1)/2, -y0), ordScale), 0, ordBottom+bar+ordGap)
		lx0, _, lx1, _ := path.BBox()
		path = append(path, rectangle(lx0, ordBottom, lx1, ordBottom+bar))
		dx := ordSide - lx0
		path = moved(path, dx, 0).Round()
		gs = append(gs, glyphs.Glyph{Name: o.name, Rune: o.r, Advance: math.Round(lx1 - lx0 + 2*ordSide), Path: path})
	}
	return gs, nil
}

// scaled returns the path scaled by s about the origin.
func scaled(path outline.Path, s float64) outline.Path {
	f := func(q outline.Point) outline.Point { return outline.Point{X: q.X * s, Y: q.Y * s} }
	out := make(outline.Path, len(path))
	for i, c := range path {
		out[i] = outline.Contour{Start: f(c.Start)}
		for _, sg := range c.Segs {
			out[i].Segs = append(out[i].Segs, outline.Segment{IsCurve: sg.IsCurve, P1: f(sg.P1), P2: f(sg.P2), P3: f(sg.P3)})
		}
	}
	return out
}

// rectangle returns the contour of the rectangle, counterclockwise.
func rectangle(x0, y0, x1, y1 float64) outline.Contour {
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	return outline.Contour{Start: pt(x0, y0), Segs: []outline.Segment{
		{P3: pt(x1, y0)}, {P3: pt(x1, y1)}, {P3: pt(x0, y1)}, {P3: pt(x0, y0)},
	}}
}

// circle returns the circle of the diameter d about (cx, cy) as a contour
// counterclockwise, with nodes at its extremes.
func circle(cx, cy, d float64) outline.Contour {
	const k = 0.5522847498
	r := d / 2
	pt := func(x, y float64) outline.Point { return outline.Point{X: cx + x*r, Y: cy + y*r} }
	return outline.Contour{Start: pt(1, 0), Segs: []outline.Segment{
		{IsCurve: true, P1: pt(1, k), P2: pt(k, 1), P3: pt(0, 1)},
		{IsCurve: true, P1: pt(-k, 1), P2: pt(-1, k), P3: pt(-1, 0)},
		{IsCurve: true, P1: pt(-1, -k), P2: pt(-k, -1), P3: pt(0, -1)},
		{IsCurve: true, P1: pt(k, -1), P2: pt(1, -k), P3: pt(1, 0)},
	}}
}

// The spaces of fixed width and the compound word mark, which have no
// outline: the italic takes them from the roman with the roman's advances,
// which depend on the em and on the figures, which the italic keeps.  The
// punctuation space has the advance of the full stop (see Table).
var fixedSpaces = []sign{
	{"uni2000", ' '}, {"uni2001", ' '}, {"uni2002", ' '}, {"uni2003", ' '},
	{"uni2004", ' '}, {"uni2005", ' '}, {"uni2006", ' '}, {"uni2007", ' '},
	{"uni2008", ' '}, {"uni205F", ' '}, {"uni200C", '‌'},
}

// addSpaces adds the spaces of fixed width and the compound word mark.
func addSpaces(gs []glyphs.Glyph) []glyphs.Glyph {
	rp := design.Regular()
	for _, sg := range fixedSpaces {
		_, adv, ok := glyphs.Outline(sg.name, rp)
		if !ok {
			panic("assembly: no glyph " + sg.name + " in the roman")
		}
		gs = append(gs, glyphs.Glyph{Name: sg.name, Rune: sg.r, Advance: adv})
	}
	return gs
}

// derived adds, to the fitted bases, the glyphs which are made of them: the
// punctuation space set to the advance of the italic full stop, the two
// letters S, each in the advance of the S of the italic, and the form of i of
// the feature locl, which is the i.
func derived(gs []glyphs.Glyph) []glyphs.Glyph {
	find := func(name string) int {
		i := slices.IndexFunc(gs, func(g glyphs.Glyph) bool { return g.Name == name })
		if i < 0 {
			panic("assembly: no glyph " + name)
		}
		return i
	}
	gs[find("uni2008")].Advance = gs[find("period")].Advance
	s := gs[find("S")]
	gs = append(gs, glyphs.Glyph{
		Name: "S_S", Advance: 2 * s.Advance,
		Path: append(append(outline.Path(nil), s.Path...), moved(s.Path, s.Advance, 0)...),
	})
	i := gs[find("i")]
	i.Name, i.Rune = "i.loclTRK", 0
	return append(gs, i)
}

// likeOf are the glyphs which take the kerning of another glyph and are not
// composed (the roman's table of like, compose.go): the form of i of locl
// takes that of i, the two letters S that of S, the hyphens U+2010 and
// U+2011 that of the hyphen, the figure dash that of the en dash and the
// horizontal bar that of the em dash.
var likeOf = map[string]string{
	"i.loclTRK": "i", "S_S": "S", "uni2010": "hyphen", "uni2011": "hyphen",
	"figuredash": "endash", "uni2015": "emdash",
}
