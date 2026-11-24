// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"math"
	"slices"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The signs of the encoding TS1 which the second part of the coverage work
// drew for the roman, and the old-style and proportional figures.  Each is
// either sloped or upright.
//
// Upright, as the arrows, the open circles, the brackets which are
// mathematics, the double bar and the signs of fixed shape: the four arrows, the asterisk and the radical of mathematics,
// the open bullet and the large circle, the double bar, the angle brackets,
// the white square brackets and the stroked brackets, the mark of reference and
// the estimated sign, the star of born and the leaf.
//
// Sloped from the roman by the method of the capitals, with the rounds
// turned and the dots, balls and rings which are circles kept: every sign
// which holds a letter or a vertical stroke (the currency signs but the
// currency sign itself, the letterlike signs, the interrobangs, the commercial
// minus, the dashes and the low quotes of TS1, the barred O, the rings of
// marriage and divorce, the cross of died and the note), and the figures, which are
// sloped from the roman's old-style and proportional figures as the lining
// figures are.  The old-style dollar and cent stand in the cell of the
// old-style figures.
func init() {
	for _, g := range []struct {
		name string
		r    rune
	}{
		{"arrowleft", '←'}, {"arrowright", '→'}, {"arrowup", '↑'}, {"arrowdown", '↓'},
		{"asteriskmath", '∗'}, {"radical", '√'},
		{"openbullet", '◦'}, {"uni25EF", '◯'},
		{"uni2016", '‖'}, {"angleleft", '\u2329'}, {"angleright", '\u232A'},
		{"uni27E6", '⟦'}, {"uni27E7", '⟧'}, {"uni2045", '⁅'}, {"uni2046", '⁆'},
		{"uni203B", '※'}, {"estimated", '℮'},
		{"born", 0xEB09}, {"leaf", 0xEB40},
	} {
		sources[g.name] = source{r: g.r, upright: true}
	}

	for _, g := range []struct {
		name string
		r    rune
		src  source
	}{
		{"quotesinglbase.ts1", 0xEB56, source{dots: signDots}},
		{"quotedblbase.ts1", 0xEB52, source{dots: signDots}},
		{"twelveudash", 0xE400, source{dots: punctDots}},
		{"threequartersemdash", 0xF6DE, source{dots: punctDots}},
		{"uni2E40", 0x2E40, source{}},
		{"hyphendbl.alt", 0xEB32, source{}},
		{"uni2422", 0x2422, source{draw: stemFloor("uni2422"), refit: refitBlank}},
		{"uni2052", 0x2052, source{dots: ringDots("uni2052")}},
		{"paragraph.alt", 0xEB4C, source{}},
		{"uni203D", 0x203D, source{dots: signDots}},
		{"uni2E18", 0x2E18, source{dots: signDots}},

		{"colonmonetary", 0x20A1, source{}},
		{"uni20A9", 0x20A9, source{}},
		{"uni20A6", 0x20A6, source{}},
		{"uni20B1", 0x20B1, source{rounds: bowlCounters("uni20B1")}},
		{"lira", 0x20A4, source{ball: letterBall, zone: 3, refit: refitLira}},
		{"dong", 0x20AB, source{rounds: bowlCounters("dong")}},
		{"uni0E3F", 0x0E3F, source{rounds: bowlCounters("uni0E3F")}},
		{"uni20B2", 0x20B2, source{}},
		{"florin", 0x0192, source{ball: figureBall, zone: 1.5, refit: refitFlorin}},
		{"dollar.oldstyle", 0xF724, source{dots: punctDots, tabular: true, oldstyle: true, refit: refitDollar}},
		{"cent.oldstyle", 0xF7A2, source{rounds: centRoundOf("cent.oldstyle", 50, 310), tabular: true, oldstyle: true}},

		{"trademark", 0x2122, source{}},
		{"uni2120", 0x2120, source{}},
		{"uni2117", 0x2117, source{rounds: letterCounters("uni2117")}},
		{"u1F12F", 0x1F12F, source{}},
		{"uni2116", 0x2116, source{rounds: allCounters("uni2116")}},
		{"prescription", 0x211E, source{rounds: allCounters("prescription")}},
		{"uni2103", 0x2103, source{dots: ringDots("uni2103")}},
		{"Omega", 0x2126, source{refit: refitOmega}},
		{"uni2127", 0x2127, source{refit: trimOvershoots}},

		{"Orogate", 0xEC10, source{rounds: allCounters("Orogate")}},
		{"orogate", 0xEC11, source{rounds: allCounters("orogate")}},
		{"uni26AD", 0x26AD, source{rounds: allCounters("uni26AD")}},
		{"uni26AE", 0x26AE, source{rounds: allCounters("uni26AE")}},
		{"died", 0xEB16, source{}},
		{"musicalnote", 0x266A, source{}},
	} {
		src := g.src
		src.r = g.r
		sources[g.name] = src
	}

	// The ties are marks, sheared in the assembly (internal/italic/assembly),
	// and the compound word marks have no ink.

	// The figures: the old-style tabular, old-style proportional and lining
	// proportional forms of the ten figures.  The tabular form keeps the
	// advance of the roman and stands centred in its cell, as the lining
	// figures do; the other two keep their side spaces.
	names := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
	for i, n := range names {
		for _, form := range []struct {
			suffix  string
			tabular bool
			old     bool
		}{{".taboldstyle", true, true}, {".oldstyle", false, true}, {".prop", false, false}} {
			name := n + form.suffix
			src := source{ball: figureBall, tabular: form.tabular, oldstyle: form.tabular && form.old}
			if i == 2 && form.old {
				src.ball = func(rp design.Params) float64 { return rp.Figures.Old.TwoBall }
			}
			switch n {
			case "zero", "six", "eight", "nine":
				src.rounds = allCounters(name)
			}
			switch n {
			case "six", "nine":
				src.zone = 3.5
			case "three", "five":
				src.zone = 1
			}
			// The zones of the balls of 3 and 5, chosen as those of the
			// lining figures were: those with the fewest findings of the
			// smoothness check among 0.5 to 6.
			if z, ok := ballZones[name]; ok {
				src.zone = z
			}
			src.refit = figureRefits[name]
			sources[name] = src
		}
	}
}

// ballZones are the zones of the balls of the figures 3 and 5 in the
// old-style and proportional forms (see circle).
var ballZones = map[string]float64{
	"three.taboldstyle": 2, "three.oldstyle": 2, "three.prop": 2,
	"five.taboldstyle": 1, "five.oldstyle": 8, "five.prop": 1,
}

// stemFloor returns the drawing of the roman glyph with every
// vertical line between 38 and minStem long made minStem long, by moving its
// last end up or down: the lattice treats a vertical line as a stem only from
// minStem on, and the rounding of the sheared outline makes a line of 39.7
// units a line of 40, which then lies off the lattice.
func stemFloor(name string) func(Params, design.Params) (outline.Path, float64) {
	return func(_ Params, rp design.Params) (outline.Path, float64) {
		path, adv, _ := glyphsOutline(name, rp)
		out := explicit(path)
		for ci, c := range out {
			segs := append([]outline.Segment(nil), c.Segs...)
			p0 := c.Start
			for i, s := range segs {
				d := s.P3.Y - p0.Y
				if !s.IsCurve && s.P3.X == p0.X && d != 0 && math.Abs(d) >= 38 && math.Abs(d) < minStem {
					segs[i].P3.Y = p0.Y + math.Copysign(minStem, d)
				}
				p0 = segs[i].P3
			}
			out[ci] = outline.Contour{Start: c.Start, Segs: segs}
		}
		return out, adv
	}
}

// ringDots returns the diameters of the rings (the contours which are
// no more than 300 units across and as wide as high) of the roman glyph,
// which stay circles.
func ringDots(name string) func(design.Params) []float64 {
	return func(rp design.Params) []float64 {
		path, _, _ := glyphsOutline(name, rp)
		var out []float64
		for _, c := range path {
			x0, y0, x1, y1 := outline.Path{c}.BBox()
			w, h := x1-x0, y1-y0
			if w <= 300 && h <= 300 && w-h < 1 && h-w < 1 {
				if !slices.Contains(out, w) {
					out = append(out, w)
				}
			}
		}
		return out
	}
}

// centRoundOf returns the rounds function of the old-style cent sign: the
// inner edge of its c, found from the counter of the sign, which the bar
// cuts, and the ink of the c.
func centRoundOf(name string, from, to float64) func(Params, design.Params) []round {
	return func(_ Params, rp design.Params) []round {
		path, _, _ := romanOutline(name, rp)
		cs := counters(path, 0, 360, 0)
		if len(cs) == 0 {
			return nil
		}
		x0, _, x1, _ := path.BBox()
		cx := (x0 + x1) / 2
		c := cs[0]
		return []round{{C: pt(cx, c.C.Y), A: cx - (c.C.X - c.A), B: c.B, From: from, To: to, Taper: 20}}
	}
}
