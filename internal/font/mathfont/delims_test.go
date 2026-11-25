// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"strings"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/fair"
)

// delimOf builds the sizes and parts of the delimiters of the parameters.
func delimOf(t *testing.T, p design.Params) (map[string]Glyph, map[string][]string, map[string][]Part) {
	t.Helper()
	gs, variants, parts, err := delimGlyphs(p)
	if err != nil {
		t.Fatal(err)
	}
	return byName(gs), variants, parts
}

// strokeAtAxis returns the width of the first run of ink along the axis: the
// crescent of the parenthesis, the stem of the bracket, the bar.
func strokeAtAxis(p outline.Path, axis float64) float64 {
	r := runsH(p, axis)
	if len(r) == 0 {
		return 0
	}
	return r[0][1] - r[0][0]
}

// TestDelimiterSizes requires every size of a delimiter to have its height
// and to be centred on the axis within one unit, the sizes to grow, and the
// stroke at the axis to be the stroke of the base times the height ratio to
// the power weightPower within one unit.
func TestDelimiterSizes(t *testing.T) {
	for _, s := range derivedSets(1) {
		p := s.Params
		m, variants, _ := delimOf(t, p)
		r := newRoman(p)
		for _, d := range delimiters {
			basePath, _ := d.draw(r)
			base := strokeAtAxis(basePath.Round(), p.BracketTop()-baseHeight(p)/2)
			for k, n := range variants[d.base] {
				g := m[n].Path.Round()
				_, y0, _, y1 := g.BBox()
				h := delimSizes[k]
				if math.Abs((y1-y0)-h) > 1 {
					t.Errorf("%s: %s is %v high, want %v", s.Name, n, y1-y0, h)
				}
				if c := (y0 + y1) / 2; math.Abs(c-p.Punct.MathAxis) > 1 {
					t.Errorf("%s: %s is centred at %v, off the axis", s.Name, n, c)
				}
				if d.base == "uni27E8" || d.base == "uni27E9" {
					continue // the angle brackets, TestAngleBracketSizes
				}
				want := base * math.Pow(h/baseHeight(p), weightPower)
				if got := strokeAtAxis(g, p.Punct.MathAxis); math.Abs(got-want) > 1 {
					t.Errorf("%s: the stroke of %s is %.1f at the axis, want %.1f", s.Name, n, got, want)
				}
			}
		}
	}
}

// TestDelimiterAssemblies requires the parts of every assembly to stand on
// the baseline with the advance which the table gives, every connector to be
// at least MinConnectorOverlap long and to lie on a straight vertical stroke
// of the part, the extender to be that stroke, and the largest size to be the
// two ends joined with the full connector, within one unit.
func TestDelimiterAssemblies(t *testing.T) {
	for _, s := range derivedSets(1) {
		p := s.Params
		m, variants, parts := delimOf(t, p)
		overlap := math.Round(p.Punct.MathThick / 2)
		for base, ps := range parts {
			if len(ps) != 3 {
				continue // the braces, TestBraceAssemblies
			}
			var bot, ext, top outline.Path
			for _, pt := range ps {
				g := m[pt.Name].Path.Round()
				_, y0, _, y1 := g.BBox()
				if y0 != 0 || math.Abs(y1-pt.FullAdvance) > 0.5 {
					t.Errorf("%s: %s reaches from %v to %v, want 0 to %v", s.Name, pt.Name, y0, y1, pt.FullAdvance)
				}
				for _, cn := range []float64{pt.StartConn, pt.EndConn} {
					if cn != 0 && cn < overlap {
						t.Errorf("%s: a connector of %s is %v, shorter than %v", s.Name, pt.Name, cn, overlap)
					}
				}
				// The ink along the connectors is a straight vertical stroke.
				check := func(lo, hi float64) {
					a, b := runsH(g, lo+0.5), runsH(g, hi-0.5)
					if len(a) != len(b) {
						t.Errorf("%s: %s is not straight from %v to %v", s.Name, pt.Name, lo, hi)
						return
					}
					for k := range a {
						if math.Abs(a[k][0]-b[k][0]) > 0.5 || math.Abs(a[k][1]-b[k][1]) > 0.5 {
							t.Errorf("%s: %s is not straight from %v to %v", s.Name, pt.Name, lo, hi)
						}
					}
				}
				if pt.StartConn > 0 {
					check(0, pt.StartConn)
				}
				if pt.EndConn > 0 {
					check(pt.FullAdvance-pt.EndConn, pt.FullAdvance)
				}
				switch {
				case strings.HasSuffix(pt.Name, ".bot"):
					bot = g
				case strings.HasSuffix(pt.Name, ".ext"):
					ext = g
				case strings.HasSuffix(pt.Name, ".top"):
					top = g
				}
			}
			if bot == nil || ext == nil || top == nil {
				t.Errorf("%s: the assembly of %s lacks a part", s.Name, base)
				continue
			}
			// The extender is the stroke at the joints.
			if a, b := runsH(ext, 1), runsH(top, 1); len(a) != len(b) || math.Abs(a[0][0]-b[0][0]) > 0.5 || math.Abs(a[0][1]-b[0][1]) > 0.5 {
				t.Errorf("%s: the extender of %s is not the stroke of its top part", s.Name, base)
			}
			// The largest size is the ends joined with the full connector.
			vs := variants[base]
			largest := m[vs[len(vs)-1]].Path.Round()
			_, _, _, hb := bot.BBox()
			_, ly0, _, _ := largest.BBox()
			joined := join(shifted(bot, 0, ly0), shifted(top, 0, ly0+hb-connector))
			for i := 1; i < 40; i++ {
				y := ly0 + 3000*float64(i)/40
				if y > ly0+hb-connector-1 && y < ly0+hb+1 {
					continue // the overlap, where the two ends lie on each other
				}
				a, b := runsH(largest, y), mergedRuns(runsH(joined, y))
				if len(a) != len(b) {
					t.Errorf("%s: at %v the largest %s has %d runs, the joined ends %d", s.Name, y, base, len(a), len(b))
					continue
				}
				for k := range a {
					if math.Abs(a[k][0]-b[k][0]) > 1 || math.Abs(a[k][1]-b[k][1]) > 1 {
						t.Errorf("%s: at %v the largest %s differs from its joined ends", s.Name, y, base)
					}
				}
			}
		}
	}
}

// mergedRuns joins runs which overlap or touch.
func mergedRuns(rs [][2]float64) [][2]float64 {
	var out [][2]float64
	for _, r := range rs {
		if n := len(out); n > 0 && r[0] <= out[n-1][1]+1e-6 {
			out[n-1][1] = math.Max(out[n-1][1], r[1])
			continue
		}
		out = append(out, r)
	}
	return out
}

// TestDelimitersPassTheChecks applies the structural checks of the regular,
// with the bounding-box rule of the base, and the fairness analysis to every
// size and part, on the regular parameters; a size or a part has no more
// findings of each kind of the fairness analysis than its base has, which
// are the intended inflections of the S-curves of the brace.
func TestDelimitersPassTheChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("the delimiters are checked in the full tier")
	}
	p := design.Regular()
	gs, _, _, err := delimGlyphs(p)
	if err != nil {
		t.Fatal(err)
	}
	r := newRoman(p)
	allowed := map[string]map[fair.Kind]int{}
	for _, d := range delimiters {
		path, _ := d.draw(r)
		allowed[d.base] = map[fair.Kind]int{}
		for _, f := range fair.Analyse(d.base, path.Round(), fair.Default).Findings {
			allowed[d.base][f.Kind]++
		}
	}
	for _, g := range gs {
		if strings.Contains(g.Name, ".v") {
			if err := glyphs.CheckOutline(p, g.Like, g.Path, g.Advance); err != nil {
				t.Errorf("%s: %v", g.Name, err)
			}
		}
		base, _, _ := strings.Cut(g.Name, ".")
		count := map[fair.Kind]int{}
		for _, f := range fair.Analyse(g.Name, g.Path.Round(), fair.Default).Findings {
			count[f.Kind]++
			if count[f.Kind] > allowed[base][f.Kind] {
				t.Errorf("%s: a finding of the fairness analysis which its base does not have: %v", g.Name, f)
			}
		}
	}
}

// baseDelimiterGlyphs returns the base delimiters and their script forms as
// the build of the font places them: in the full tier the glyphs of the font,
// in the routine tier the delimiters of the regular and of the derived signs
// with their script forms, placed by placeDelimiters, and their sizes.
func baseDelimiterGlyphs(t *testing.T) []Glyph {
	t.Helper()
	p := design.Regular()
	if !testing.Short() {
		gs, err := regularGlyphs()
		if err != nil {
			t.Fatal(err)
		}
		return gs
	}
	r := newRoman(p)
	var gs []Glyph
	for _, n := range romanNames() {
		if delimiterNames[n] {
			path, adv := r.get(n)
			gs = append(gs, Glyph{Name: n, Rune: 1, Advance: math.Round(adv), Path: path, Class: Ord, Like: n})
		}
	}
	for _, g := range derivedSigns(derivedSets(0)[0]) {
		if delimiterNames[g.Name] {
			g.Rune = 1
			gs = append(gs, g)
		}
	}
	gs, err := addScripts(gs, false)
	if err != nil {
		t.Fatal(err)
	}
	placeDelimiters(gs, p.Punct.MathAxis)
	sizes, _, _, err := delimGlyphs(p)
	if err != nil {
		t.Fatal(err)
	}
	return append(gs, sizes...)
}

// TestBaseDelimitersAreOnTheAxis requires the base delimiters of the font,
// the parentheses, brackets, braces, bars, floors, ceilings and angle
// brackets, and their script forms, to be centred on the axis to the unit like
// every larger size; \left( around a box which fits the base glyph stands
// level with the fraction bar and the minus (the text font's own delimiters
// stand 12 units higher and are not moved).
func TestBaseDelimitersAreOnTheAxis(t *testing.T) {
	axis := design.Regular().Punct.MathAxis
	found := 0
	for _, g := range baseDelimiterGlyphs(t) {
		base := g.Name
		if i := strings.Index(base, "."); i >= 0 {
			base = base[:i]
		}
		if !delimiterNames[base] || strings.HasSuffix(g.Name, ".top") || strings.HasSuffix(g.Name, ".bot") ||
			strings.HasSuffix(g.Name, ".ext") || strings.HasSuffix(g.Name, ".mid") || strings.HasSuffix(g.Name, ".ext2") {
			continue
		}
		_, y0, _, y1 := g.Path.Round().BBox()
		if c := (y0 + y1) / 2; math.Abs(c-axis) > 1 {
			t.Errorf("%s is centred at %v, off the axis at %v", g.Name, c, axis)
		}
		found++
	}
	if want := 14*(1+len(delimSizes)) + 2*7; found != want { // with the script forms of the seven which have them
		t.Errorf("%d delimiters checked, want %d", found, want)
	}
}
