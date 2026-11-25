// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"
	"strings"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The braces, brackets and parentheses over and under: U+23DE and U+23DF, U+23DC and U+23DD, U+23B4
// and U+23B5.  Each is the construction of the vertical delimiter of the
// regular run at the height of its width, with the strokes of the base
// delimiter at every width, turned by a quarter clockwise: the left
// brace, parenthesis and bracket give the forms over, whose point, bulge or
// bar is at the top, and the right ones the forms under; the ink stands
// spanGap above AccentBaseHeight or hangs spanGap below the baseline.  The bow of
// the parenthesis grows as the square root of the width over the height of
// the base delimiters, as in the vertical sizes, and the depth of the brace
// as the width to the power braceDepthPower.
//
// The sizes and the assembly are made as those of the vertical delimiters
// (delims.go, braces.go), with hEndRun, hConn and hExt: the largest size is
// the construction shorter by what joining the ends with the full connector
// adds, stretched at its cuts; the assembly with no extender is the
// largest size.
const (
	// braceDepth is the ink depth of the brace over the base, spanWidths[0]
	// wide.
	braceDepth = 200.0
	// braceDepthPower is the exponent of the width ratio by which the depth
	// of the largest brace grows.
	braceDepthPower = 0.25
	// parenDeepen is the bow of the parentheses over and under over that of
	// the vertical parenthesis at the same width.
	parenDeepen = 1.05
	// bracketWeight is the bar of the brackets over and under over the stem
	// of the vertical bracket.  bracketHook is the reach of its ends over
	// that of the vertical bracket: the depth grows from 186 to 210.
	bracketWeight = 0.9
	bracketHook   = 1.15
	// hSpan is what joining two ends with the full connector adds to the
	// construction at each cut.
	hSpan = 2*hEndRun - hConn
)

// spanWidths are the ink widths of the sizes of the spans, the base first;
// above the largest the assembly covers the argument exactly.
var spanWidths = []float64{520, 700, 950, 1300, 1750}

// A span names a brace, bracket or parenthesis over or under.
type span struct {
	name   string // the glyph of the font
	r      rune
	parent string // the vertical delimiter of the regular
	under  bool
	kind   int // withMiddle or withEnds
	sizes  int // the number of sizes with the base
}

var spans = []span{
	{"uni23DE", 0x23DE, "braceleft", false, withMiddle, 5},
	{"uni23DF", 0x23DF, "braceright", true, withMiddle, 5},
	{"uni23DC", 0x23DC, "parenleft", false, withEnds, 5},
	{"uni23DD", 0x23DD, "parenright", true, withEnds, 5},
	{"uni23B4", 0x23B4, "bracketleft", false, withEnds, 5},
	{"uni23B5", 0x23B5, "bracketright", true, withEnds, 5},
}

// spanParams returns the parameters of the vertical delimiter h high whose
// depth is that of the span w wide.
func spanParams(p design.Params, h, w float64) design.Params {
	q := sized(p, h, 1, math.Sqrt(w/baseHeight(p)))
	q.Sym.BraceInk = math.Round(braceDepth * math.Pow(w/spanWidths[0], braceDepthPower))
	q.Punct.ParenSagitta *= parenDeepen
	q.Punct.BracketStem *= bracketWeight
	q.Punct.BracketArm *= bracketWeight
	q.Punct.BracketReach *= bracketHook / bracketWeight
	return q
}

// turnedCW returns the path turned by a quarter clockwise about the origin.
func turnedCW(path outline.Path) outline.Path {
	return mapPoints(path, func(q outline.Point) outline.Point { return pt(q.Y, -q.X) })
}

// spanLift returns the shift which puts the turned path spanGap above
// AccentBaseHeight, where LuaTeX sets it spanGap above its argument, or
// spanGap below the baseline if under.
func spanLift(p design.Params, turned outline.Path, under bool) float64 {
	_, y0, _, y1 := turned.BBox()
	gap := spanGap(p)
	if under {
		return math.Round(-gap - y1)
	}
	return math.Round(accentBase(p) + gap - y0)
}

// spanGap returns the white between a span and its argument: two rules.
func spanGap(p design.Params) float64 { return 2 * p.Punct.MathThick }

// largest returns the construction of the assembly of the span, the heights
// of its cuts and its largest size, the construction stretched at the cuts.
func (sp span) largest(p design.Params) (cons, size outline.Path, adv, lo, hi float64, err error) {
	wl := spanWidths[sp.sizes-1]
	draw := func(h float64) (outline.Path, float64) {
		return newRomanNominal(spanParams(p, h, wl), spanWidths[sp.sizes-2]).get(sp.parent)
	}
	axis := p.Punct.MathAxis
	if sp.kind == withMiddle {
		cons, adv = draw(wl - 2*hSpan)
		lo, hi, err = braceCuts(cons, axis)
		if err != nil {
			return
		}
		size, err = stretchPath(cons, hi, hSpan)
		if err != nil {
			return
		}
		size, err = stretchPath(size, lo-hSpan/2, hSpan)
		return
	}
	cons, adv = draw(wl - hSpan)
	size, err = stretchPath(cons, axis, hSpan)
	return cons, size, adv, axis, axis, err
}

// spanGlyphs returns the spans with their sizes and the parts of their
// assemblies, and the construction of each.
func spanGlyphs(p design.Params) ([]Glyph, map[string]construction, error) {
	var gs []Glyph
	cons := map[string]construction{}
	for _, sp := range spans {
		cpath, lpath, adv, lo, hi, err := sp.largest(p)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", sp.name, err)
		}
		// The sizes, each placed by its own ink, and the parts placed as
		// the largest size.
		lturned := turnedCW(lpath)
		lift := spanLift(p, lturned, sp.under)
		place := func(path outline.Path) (outline.Path, float64) {
			tp := turnedCW(path)
			x0, _, x1, _ := tp.BBox()
			return shifted(tp, -x0, spanLift(p, tp, sp.under)).Round(), math.Round(x1 - x0)
		}
		var c construction
		for k := range sp.sizes {
			path := lpath
			if k < sp.sizes-1 {
				w := spanWidths[k]
				path, _ = newRoman(spanParams(p, w, w)).get(sp.parent)
			}
			tp, w := place(path)
			name := sp.name
			rn := sp.r
			if k > 0 {
				name, rn = fmt.Sprintf("%s.h%d", sp.name, k), 0
				c.Variants = append(c.Variants, name)
			}
			gs = append(gs, Glyph{Name: name, Rune: rn, Advance: w, Path: tp, Class: Accent, Like: "acutecomb"})
		}
		var pg []Glyph
		var ps []Part
		if sp.kind == withMiddle {
			pg, ps, err = cutBrace(cpath, adv, lo, hi, sp.name, "acutecomb", hEndRun, hConn, hExt)
			if err != nil {
				return nil, nil, fmt.Errorf("%s: %w", sp.name, err)
			}
		} else {
			top, bot, ext, err := assemblyParts(cpath, lo, true, hEndRun, hExt)
			if err != nil {
				return nil, nil, fmt.Errorf("%s: %w", sp.name, err)
			}
			height := func(q outline.Path) float64 { _, _, _, h := q.BBox(); return h }
			pg = []Glyph{{Name: sp.name + ".bot", Path: bot}, {Name: sp.name + ".ext", Path: ext}, {Name: sp.name + ".top", Path: top}}
			ps = []Part{
				{Name: sp.name + ".bot", EndConn: hConn, FullAdvance: height(bot)},
				{Name: sp.name + ".ext", StartConn: hConn, EndConn: hConn, FullAdvance: hExt, Extender: true},
				{Name: sp.name + ".top", StartConn: hConn, FullAdvance: height(top)},
			}
		}
		// The parts are drawn from the bottom up and stand on the baseline,
		// with the abscissae of the construction; turned, they run from the
		// left, start at the origin and lie at the height of the largest
		// size.
		rename := strings.NewReplacer(".bot", ".left", ".top", ".right")
		for _, g := range pg {
			tp := shifted(turnedCW(g.Path), 0, lift)
			_, _, x1, _ := tp.BBox()
			gs = append(gs, Glyph{Name: rename.Replace(g.Name), Advance: math.Round(x1), Path: tp.Round(), Class: Accent, Like: "acutecomb"})
		}
		for i := range ps {
			ps[i].Name = rename.Replace(ps[i].Name)
		}
		c.Parts = ps
		cons[sp.name] = c
	}
	return gs, cons, nil
}
