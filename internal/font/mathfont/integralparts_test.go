// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/font/design/designtest"
)

// TestIntegralAssemblyJoins requires the top hook, the extender and the
// bottom hook of the integral to meet in the same upright stem: the run of
// ink at the joining end of each hook is that of the extender, the assembly
// without an extender is as high as the display size, and the hooks have no
// finding of the fairness analysis.
func TestIntegralAssemblyJoins(t *testing.T) {
	warm(derivedSets(1), func(s designtest.Set) { integralGlyphs(s.Params) })
	for _, s := range derivedSets(1) {
		gs, err := integralGlyphs(s.Params)
		if err != nil {
			t.Fatal(err)
		}
		m := byName(gs)
		bot, ext, top := m["uni222B.bot"], m["uni222B.ext"], m["uni222B.top"]
		if len(gs[0].Parts) != 3 {
			t.Fatalf("%s: the integral has %d parts, want 3", s.Name, len(gs[0].Parts))
		}
		_, _, _, bh := bot.Path.Round().BBox()
		_, ty0, _, th := top.Path.Round().BBox()
		_, _, _, eh := ext.Path.Round().BBox()
		cut := func(g Glyph, y float64) [2]float64 {
			rs := runsOf(g.Path, y)
			if len(rs) != 1 {
				t.Fatalf("%s: %s has %d runs at %.0f, want 1", s.Name, g.Name, len(rs), y)
			}
			return rs[0]
		}
		for _, c := range []struct {
			name string
			a, b [2]float64
		}{
			{"bottom hook", cut(bot, bh-1), cut(ext, 1)},
			{"extender", cut(ext, eh-1), cut(top, ty0+1)},
		} {
			if math.Abs(c.a[0]-c.b[0]) > 0.5 || math.Abs(c.a[1]-c.b[1]) > 0.5 {
				t.Errorf("%s: the %s ends in the run %.1f to %.1f, the next part starts with %.1f to %.1f", s.Name, c.name, c.a[0], c.a[1], c.b[0], c.b[1])
			}
		}
		if total := bh + th - connector; math.Abs(total-intDisplayHeight) > 2 {
			t.Errorf("%s: the assembly without an extender is %.0f high, want %v", s.Name, total, intDisplayHeight)
		}
		// The hooks are cut from the integral and have no finding of the
		// fairness analysis on any parameter set.
		for _, g := range []Glyph{bot, top} {
			if fs := fair.Analyse(g.Name, g.Path.Round(), fair.Default).Findings; len(fs) > 0 {
				t.Errorf("%s: %s has %d findings of the fairness analysis, the first %v", s.Name, g.Name, len(fs), fs[0])
			}
		}
	}
}

// TestIntegralIsFair requires the integral at the text and the display size
// to have no finding of the fairness analysis on any parameter set, but the
// two inflections of the S-curve of its stem near its middle, which are
// intended; the stroke of the integral takes the shape (integralShape) which
// the heavier and the lighter strokes of caps-up and caps-down need.
func TestIntegralIsFair(t *testing.T) {
	warm(derivedSets(3), func(s designtest.Set) { integralGlyphs(s.Params) })
	for _, s := range derivedSets(3) {
		gs, err := integralGlyphs(s.Params)
		if err != nil {
			t.Fatal(err)
		}
		for _, g := range gs[:2] {
			path := g.Path.Round()
			_, y0, _, y1 := path.BBox()
			n := 0
			for _, f := range fair.Analyse(g.Name, path, fair.Default).Findings {
				if f.Kind == fair.Inflection && math.Abs(f.Y-(y0+y1)/2) < 0.15*(y1-y0) {
					n++
					continue
				}
				t.Errorf("%s: %s: a finding of the fairness analysis: %v", s.Name, g.Name, f)
			}
			if n != 2 {
				t.Errorf("%s: %s has %d inflections of the S-curve near its middle, want 2", s.Name, g.Name, n)
			}
		}
	}
}
