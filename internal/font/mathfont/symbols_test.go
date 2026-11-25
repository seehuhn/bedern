// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"strings"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/design/designtest"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/font/italic"
)

// The tests of the ordinary symbols, the big unions and circled operators
// and the double-struck letters (ordinary.go, bigsets.go,
// doublestruck.go, bedern/internal/font/italic/mathsigns.go).

// symbolRunes are the code points of the ordinary symbols, the big unions and
// circled operators and the double-struck letters.
var symbolRunes = []rune{
	0x221E, 0x2200, 0x2205, 0x2207, 0x2113, 0x2202,
	0x22C3, 0x22C2, 0x2A01, 0x2A02, 0x222E, 0x222C, 0x222D, 0x222F,
	0x1D715, 0x1D7D9,
}

func init() {
	for _, l := range dsLetters {
		symbolRunes = append(symbolRunes, l.r)
	}
}

// italicPen are the symbols drawn with the pen of the italic, whose
// smoothness is checked with the profile of the italic.
var italicPen = map[string]bool{"uni2113": true, "uni2202": true}

// area returns the area of the ink of the path.
func area(p outline.Path) float64 {
	var a float64
	for _, c := range p {
		a += c.SignedArea()
	}
	return math.Abs(a)
}

// dsOf holds the double-struck letters by the name of their parameter set.
var dsOf sync.Map

// bigSetsOf holds the big unions, circled operators and contour integrals by
// the name of their parameter set.
var bigSetsOf sync.Map

func bigSets(t *testing.T, s designtest.Set) map[string]Glyph {
	t.Helper()
	v, _ := bigSetsOf.LoadOrStore(s.Name, sync.OnceValues(func() ([]Glyph, error) { return bigSetGlyphs(s.Params) }))
	gs, err := v.(func() ([]Glyph, error))()
	if err != nil {
		t.Fatal(err)
	}
	return byName(gs)
}

func dsSet(s designtest.Set) map[string]Glyph {
	v, _ := dsOf.LoadOrStore(s.Name, sync.OnceValue(func() []Glyph { return dsGlyphs(s.Params) }))
	return byName(v.(func() []Glyph)())
}

// TestSymbolsHaveGlyphs requires a glyph with ink for every code point of
// the symbols among the derived signs, the operators and the
// double-struck letters, and the script forms of the double-struck letters.
func TestSymbolsHaveGlyphs(t *testing.T) {
	s := derivedSets(0)[0]
	m := map[rune]Glyph{}
	all := append(append(derivedSigns(s), dsGlyphs(s.Params)...), mapValues(bigSets(t, s))...)
	all = append(all, partialItalic())
	for _, g := range all {
		if g.Rune != 0 {
			m[g.Rune] = g
		}
	}
	for _, r := range symbolRunes {
		if g, ok := m[r]; !ok || len(g.Path) == 0 {
			t.Errorf("%U has no glyph", r)
		}
	}
	ds := dsSet(s)
	for _, l := range dsLetters {
		for _, suffix := range []string{".s1", ".s2"} {
			if _, ok := ds[l.name+suffix]; !ok {
				t.Errorf("%s has no script form %s", l.name, suffix)
			}
		}
	}
}

func mapValues(m map[string]Glyph) []Glyph {
	var out []Glyph
	for _, g := range m {
		out = append(out, g)
	}
	return out
}

// TestOrdinarySymbolsFollowTheirRules requires, within one unit: ∞ and ∅
// centred on the axis, ∞ infWidth by infHeight with the rule at its outer
// sides, the ring of ∅ of the rule and its slash emptyReach beyond it; ∀ and
// ∇ standing on the baseline and reaching the cap height, the diagonals of
// ∀ alike, and the left side of ∇ heavier than its right by the weights of
// its rule.
func TestOrdinarySymbolsFollowTheirRules(t *testing.T) {
	for _, s := range derivedSets(3) {
		p := s.Params
		m := byName(derivedSigns(s))
		rule := p.Punct.MathThick
		axis := p.Punct.MathAxis
		near := func(what string, got, want float64) {
			if math.Abs(got-want) > 1 {
				t.Errorf("%s: %s is %.1f, want %.1f", s.Name, what, got, want)
			}
		}
		inf := m["uni221E"].Path.Round()
		x0, y0, x1, y1 := inf.BBox()
		near("the middle of ∞", (y0+y1)/2, axis)
		near("the width of ∞", x1-x0, infWidth)
		near("the height of ∞", y1-y0, infHeight)
		if r := runsH(inf, axis); len(r) < 3 {
			t.Errorf("%s: ∞ has %d runs at the axis", s.Name, len(r))
		} else {
			near("the left side of ∞", r[0][1]-r[0][0], rule)
			near("the right side of ∞", r[len(r)-1][1]-r[len(r)-1][0], rule)
		}

		es := m["uni2205"].Path.Round()
		x0, y0, x1, y1 = es.BBox()
		near("the middle of ∅", (y0+y1)/2, axis)
		rad := math.Round(emptyRing*p.Figures.Height) / 2
		near("the height of ∅", y1-y0, 2*(rad+emptyReach))
		if r := runsH(es, axis); len(r) == 0 {
			t.Errorf("%s: ∅ has no ink at the axis", s.Name)
		} else {
			near("the ring of ∅", r[0][1]-r[0][0], rule)
		}

		for _, n := range []string{"uni2200", "uni2207"} {
			_, y0, _, y1 := m[n].Path.Round().BBox()
			near("the foot of "+n, y0, 0)
			near("the top of "+n, y1, p.CapHeight)
		}
		fa := m["uni2200"].Path.Round()
		y := 0.3 * p.CapHeight
		l, r := acrossAt(fa, y, 20, first), acrossAt(fa, y, 20, last)
		near("the left diagonal of ∀", l, math.Round(forallStroke*rule))
		near("the right diagonal of ∀", r, math.Round(forallStroke*rule))
		na := m["uni2207"].Path.Round()
		y = 0.5 * p.CapHeight
		near("the left side of ∇", acrossAt(na, y, 20, first), math.Round(nablaThick*p.Caps.Stem))
		near("the right side of ∇", acrossAt(na, y, 20, last), math.Round(nablaThin*p.Caps.Bar))
	}
}

// TestSymbolsHaveTheirWeight requires the ink of the ordinary symbols
// over that of n, and of the big unions and circled operators over that of
// the summation of their size, to lie in the range, on the regular
// parameters.
func TestSymbolsHaveTheirWeight(t *testing.T) {
	s := derivedSets(0)[0]
	m := byName(derivedSigns(s))
	n, _, _ := glyphs.Outline("n", s.Params)
	an := area(n.Round())
	for _, c := range []struct {
		name   string
		lo, hi float64
	}{
		{"uni221E", 0.99, 1.50}, {"uni2200", 0.87, 1.18}, {"uni2205", 1.17, 1.86}, {"uni2113", 0.74, 0.95},
	} {
		if v := area(m[c.name].Path.Round()) / an; v < c.lo || v > c.hi {
			t.Errorf("the ink of %s is %.2f of that of n, outside %.2f to %.2f", c.name, v, c.lo, c.hi)
		}
	}
	ops := operators(t, s)
	for n, g := range bigSets(t, s) {
		ops[n] = g
	}
	for _, c := range []struct {
		name   string
		lo, hi float64
	}{
		{"uni22C3", 0.54, 0.79}, {"uni2A01", 1.21, 1.65}, {"uni2A02", 1.24, 1.65},
	} {
		sum := area(ops["uni2211"].Path.Round())
		if v := area(ops[c.name].Path.Round()) / sum; v < c.lo || v > c.hi {
			t.Errorf("the ink of %s is %.2f of that of ∑, outside %.2f to %.2f", c.name, v, c.lo, c.hi)
		}
	}
}

// TestItalicSymbolsAreSmooth requires ℓ and ∂ to have no finding of the
// profile of the italic and no warning of the stroke tool, the maths italic
// ∂ among them.
func TestItalicSymbolsAreSmooth(t *testing.T) {
	for _, n := range italic.SymbolNames() {
		g, err := italic.BuildSymbol(italic.Default(), n)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range g.Warnings {
			t.Errorf("%s: %s", n, w)
		}
		for _, f := range g.Findings {
			t.Errorf("%s: a finding of the fairness analysis: %v", n, f)
		}
	}
}

// TestBigSetOperators requires the sizes of ⋃ ⋂ ⨁ ⨂ and ∮ to have their
// heights and to be centred on the axis within one unit, the display sizes
// to reach DisplayOperatorMinHeight and the text sizes not, ⋂ to be ⋃
// mirrored in the axis, the stems of ⋃ and the rings of ⨁ and ⨂ to have the
// weights of their rule, and the ring of ∮ to be centred on its stem on the
// axis; on the regular parameters in the routine tier and on every set in
// the full tier.
func TestBigSetOperators(t *testing.T) {
	for _, s := range derivedSets(0) {
		m := bigSets(t, s)
		p := s.Params
		axis := p.Punct.MathAxis
		min := float64(Constants(p).DisplayOperatorMinHeight)
		law := func(h float64) float64 { return p.Caps.Stem * math.Pow(h/p.CapHeight, opPower) }
		for _, c := range []struct {
			name      string
			text, dsp float64
		}{
			{"uni22C3", opTextHeight, opDisplayHeight},
			{"uni22C2", opTextHeight, opDisplayHeight},
			{"uni2A01", circOpText, circOpDisplay},
			{"uni2A02", circOpText, circOpDisplay},
			{"uni222E", intTextHeight, intDisplayHeight},
			{"uni222C", intTextHeight, intDisplayHeight},
			{"uni222D", intTextHeight, intDisplayHeight},
			{"uni222F", intTextHeight, intDisplayHeight},
		} {
			for _, sz := range []struct {
				name string
				h    float64
			}{{c.name, c.text}, {c.name + ".d", c.dsp}} {
				_, y0, _, y1 := m[sz.name].Path.Round().BBox()
				if math.Abs(y1-y0-sz.h) > 1 {
					t.Errorf("%s: %s is %v high, want %v", s.Name, sz.name, y1-y0, sz.h)
				}
				if d := (y0+y1)/2 - axis; math.Abs(d) > 1 {
					t.Errorf("%s: %s is centred %v from the axis", s.Name, sz.name, d)
				}
			}
			if c.dsp < min || c.text >= min {
				t.Errorf("%s: %s has %v and %v, DisplayOperatorMinHeight is %v", s.Name, c.name, c.text, c.dsp, min)
			}
		}
		for _, n := range []string{"uni22C3", "uni22C3.d"} {
			a := m[n].Path.Round()
			b := mirroredY(m[strings.Replace(n, "22C3", "22C2", 1)].Path, axis).Round()
			_, y0, _, y1 := a.BBox()
			for i := 1; i < 20; i++ {
				y := y0 + (y1-y0)*float64(i)/20
				ra, rb := runsH(a, y), runsH(b, y)
				if len(ra) != len(rb) {
					t.Errorf("%s: ⋂ is not ⋃ mirrored at %v", s.Name, y)
					break
				}
				for k := range ra {
					if math.Abs(ra[k][0]-rb[k][0]) > 1 || math.Abs(ra[k][1]-rb[k][1]) > 1 {
						t.Errorf("%s: ⋂ is not ⋃ mirrored at %v", s.Name, y)
					}
				}
			}
			h := opTextHeight
			if strings.HasSuffix(n, ".d") {
				h = opDisplayHeight
			}
			for _, r := range runsH(a, axis) {
				if got, want := r[1]-r[0], cupOpStem*law(h); math.Abs(got-want) > 1 {
					t.Errorf("%s: a stem of %s is %.1f, want %.1f", s.Name, n, got, want)
				}
			}
		}
		for _, c := range []struct {
			name string
			d    float64
		}{{"uni2A01", circOpText}, {"uni2A01.d", circOpDisplay}, {"uni2A02", circOpText}, {"uni2A02.d", circOpDisplay}} {
			// Above the bars, which end in the ring at 0.35 of the diameter
			// above the middle at the most.
			r := runsH(m[c.name].Path.Round(), axis+0.38*c.d)
			if len(r) < 2 {
				t.Errorf("%s: %s has %d runs above its middle", s.Name, c.name, len(r))
				continue
			}
			// The run of a ring of the stroke w at the height y above its
			// middle is longer than w by the slope of the ring there.
			ro := c.d / 2
			y := 0.38 * c.d
			want := math.Round(circOpRing * law(c.d))
			got := math.Sqrt(ro*ro-y*y) - math.Sqrt(math.Max(0, (ro-want)*(ro-want)-y*y))
			if math.Abs((r[0][1]-r[0][0])-got) > 1.5 {
				t.Errorf("%s: the ring of %s is %.1f along the level, want %.1f", s.Name, c.name, r[0][1]-r[0][0], got)
			}
		}
		for _, n := range []string{"uni222E", "uni222E.d"} {
			g := m[n]
			path := g.Path.Round()
			rs := runsH(path, axis)
			if len(rs) != 3 {
				t.Errorf("%s: %s has %d runs at the axis, want the ring, the stem and the ring", s.Name, n, len(rs))
				continue
			}
			stem := (rs[1][0] + rs[1][1]) / 2
			if math.Abs((rs[0][0]+rs[2][1])/2-stem) > 1 {
				t.Errorf("%s: the ring of %s is not centred on its stem", s.Name, n)
			}
			_, _, x1, _ := path.BBox()
			if g.Italic <= 0 || g.Advance+g.Italic < x1 {
				t.Errorf("%s: %s has the italics correction %v; its ink reaches %v beyond the advance", s.Name, n, g.Italic, x1-g.Advance)
			}
		}
	}
}

// TestMultipleIntegrals requires ∬ and ∭ to be the strokes of the integral
// side by side, their stems at the axis multiIntPitch of the ink of the
// integral apart, with at least a rule of white between neighbouring strokes,
// and ∬ ∭ and ∯ to have the side spaces and the italics correction of the
// integral; and ∯ to have the oval about its stems, centred between them on
// the axis; on the regular parameters in the routine tier and on every set in
// the full tier.
func TestMultipleIntegrals(t *testing.T) {
	warm(derivedSets(0), func(s designtest.Set) { integralGlyphs(s.Params) })
	for _, s := range derivedSets(0) {
		m := bigSets(t, s)
		p := s.Params
		axis := p.Punct.MathAxis
		ints, err := integralGlyphs(p)
		if err != nil {
			t.Fatal(err)
		}
		one := byName(ints)
		for _, c := range []struct {
			name   string
			n      int
			ring   bool
			single string
		}{{"uni222C", 2, false, "uni222B"}, {"uni222D", 3, false, "uni222B"}, {"uni222F", 2, true, "uni222B"}} {
			for _, suffix := range []string{"", ".d"} {
				name, base := c.name+suffix, one[c.single+suffix]
				g := m[name]
				path := g.Path.Round()
				bx0, _, bx1, _ := base.Path.Round().BBox()
				x0, _, x1, _ := path.BBox()
				if math.Abs(x0-bx0) > 1.5 || (!c.ring && math.Abs(g.Advance-x1-(base.Advance-bx1)) > 1.5) {
					t.Errorf("%s: %s has the side spaces %.0f and %.0f, the integral %.0f and %.0f", s.Name, name, x0, g.Advance-x1, bx0, base.Advance-bx1)
				}
				if math.Abs(g.Italic-base.Italic) > 1.5 {
					t.Errorf("%s: %s has the italics correction %.0f, the integral %.0f", s.Name, name, g.Italic, base.Italic)
				}
				rs := runsH(path, axis)
				want := c.n
				if c.ring {
					want += 2
				}
				if len(rs) != want {
					t.Errorf("%s: %s has %d runs at the axis, want %d", s.Name, name, len(rs), want)
					continue
				}
				stems := rs
				if c.ring {
					stems = rs[1 : len(rs)-1]
					mid := (stems[0][0] + stems[0][1] + stems[len(stems)-1][0] + stems[len(stems)-1][1]) / 4
					if got := (rs[0][0] + rs[len(rs)-1][1]) / 2; math.Abs(got-mid) > 1.5 {
						t.Errorf("%s: the oval of %s is centred %.1f from the middle of its stems", s.Name, name, got-mid)
					}
				}
				ink := bx1 - bx0
				for i := 1; i < len(stems); i++ {
					pitch := (stems[i][0]+stems[i][1])/2 - (stems[i-1][0]+stems[i-1][1])/2
					if math.Abs(pitch-math.Round(multiIntPitch*ink)) > 1.5 {
						t.Errorf("%s: the stems of %s are %.1f apart, want %.0f", s.Name, name, pitch, math.Round(multiIntPitch*ink))
					}
				}
				if c.ring {
					continue
				}
				for i := 0; i+1 < len(g.Path); i++ {
					a, b := g.Path[i].Flatten(), g.Path[i+1].Flatten()
					best := math.Inf(1)
					for _, u := range a {
						for _, v := range b {
							best = math.Min(best, math.Hypot(u.X-v.X, u.Y-v.Y))
						}
					}
					if best < p.Punct.MathThick {
						t.Errorf("%s: strokes %d and %d of %s are %.1f apart, less than the rule %.0f", s.Name, i, i+1, name, best, p.Punct.MathThick)
					}
				}
			}
		}
	}
}

// dsWhiteMin is the least white between the lines of a doubled stroke at
// each level: at least 40 at text size and in the first script form and 55
// in the second.
var dsWhiteMin = [3]float64{40, 40, 55}

// dsScanOf says where the doubled stroke of a letter is taken: at the
// fraction frac of the height of the letter, between the lines with the
// indices k and k + 1 of the runs of ink there.
type dsScanOf struct {
	frac float64
	k    int
}

// dsScan holds the places where the doubled stroke of a letter is taken
// when it is not at 0.45 of the height between the first two runs: the lens
// of ℂ ℚ G O and the spine of S at the middle, the diagonals of ℕ and ℤ, of A
// and X and of M beside a stem or a line, the leg of K clear of its arm and
// the stem of Y below its crotch.
var dsScan = map[string]dsScanOf{
	"C": {0.5, 0}, "Q": {0.5, 0}, "G": {0.5, 0}, "O": {0.5, 0}, "S": {0.5, 0},
	"N": {0.5, 1}, "Z": {0.5, 0}, "A": {0.45, 1}, "X": {0.2, 1}, "M": {0.7, 1},
	"K": {0.12, 0}, "Y": {0.25, 0}, "W": {0.75, 0},
}

// TestDoubleStruckLetters requires, at each level, the white between the
// lines of every doubled stroke (the stems, the diagonals taken across, the
// lens of ℂ ℚ G and O and the spine of S at the middle) to be at least
// dsWhiteMin and each line of the weight of the level within one unit; the
// letters to reach the height of their roman; and each script form to be
// heavier and wider than the form before it; on the regular parameters in the
// routine tier and on every set in the full tier.
func TestDoubleStruckLetters(t *testing.T) {
	warm(derivedSets(0), func(s designtest.Set) { dsSet(s) })
	for _, s := range derivedSets(0) {
		p := s.Params
		m := dsSet(s)
		for _, l := range dsLetters {
			prevW := 0.0
			for level, suffix := range []string{"", ".s1", ".s2"} {
				name := l.name + suffix
				g, ok := m[name]
				if !ok {
					t.Errorf("%s: no %s", s.Name, name)
					continue
				}
				path := g.Path.Round()
				x0, y0, x1, y1 := path.BBox()
				line := math.Round(dsLine * p.CapDiag.ThinStem * dsLevels[level].weight)
				// The pair of lines of the doubled stroke at its height, the
				// white between them and the lines taken across the stroke.
				sc, ok := dsScan[l.roman]
				if !ok {
					sc = dsScanOf{0.45, 0}
				}
				y := y0 + sc.frac*(y1-y0)
				r := runsH(path, y)
				a, b := runsH(path, y-20), runsH(path, y+20)
				k := sc.k
				if len(r) < k+2 || len(a) < k+2 || len(b) < k+2 {
					t.Errorf("%s: %s has %d runs at the height %.0f, want %d", s.Name, name, len(r), y, k+2)
					continue
				}
				dx := (b[k][0] + b[k][1] - a[k][0] - a[k][1]) / 2
				sin := 40 / math.Hypot(dx, 40)
				lines := [2]float64{(r[k][1] - r[k][0]) * sin, (r[k+1][1] - r[k+1][0]) * sin}
				white := (r[k+1][0] - r[k][1]) * sin
				if white < dsWhiteMin[level] {
					t.Errorf("%s: the white of %s is %.1f, want at least %v", s.Name, name, white, dsWhiteMin[level])
				}
				for _, w := range lines {
					if math.Abs(w-line) > 1.5 {
						t.Errorf("%s: a line of %s is %.1f, want %v", s.Name, name, w, line)
					}
				}
				if l.roman != "C" && l.roman != "Q" && l.roman != "one" {
					rp, _, _ := glyphs.Outline(l.roman, p)
					_, ry0, _, ry1 := rp.Round().BBox()
					if l.roman == "N" {
						ry0 = 0 // the vertex of ℕ lies on the baseline
					}
					if math.Abs(y0-ry0) > 0.5 || math.Abs(y1-ry1) > 0.5 {
						t.Errorf("%s: %s reaches from %v to %v, its roman from %v to %v", s.Name, name, y0, y1, ry0, ry1)
					}
				}
				if x1-x0 <= prevW {
					t.Errorf("%s: %s is no wider than the form before it", s.Name, name)
				}
				prevW = x1 - x0
			}
		}
	}
}

// TestDoubleStruckLettersPassTheChecks applies the structural checks of the
// regular, with the bounding-box rule of the roman letter, and the fairness
// analysis to the double-struck letters and their script forms, on the
// regular parameters in the routine tier and on every set in the full tier.
func TestDoubleStruckLettersPassTheChecks(t *testing.T) {
	for _, s := range derivedSets(0) {
		t.Run(s.Name, func(t *testing.T) {
			t.Parallel()
			for _, g := range dsSet(s) {
				if err := glyphs.CheckOutline(s.Params, g.Like, g.Path, g.Advance); err != nil {
					t.Errorf("%s: %v", g.Name, err)
				}
				var fs []fair.Finding
				rounded := g.Path.Round()
				base := strings.SplitN(g.Name, ".", 2)[0]
				spine := 0
				for _, f := range fair.Analyse(g.Name, rounded, fair.Default).Findings {
					// The upright edge of the stem of U and J runs from the concave
					// bracket of the serif straight into the convex round, as on
					// the right stem of n in the regular, where the baseline of
					// the fairness analysis judges it intended.
					if f.Kind == fair.Inflection && inflectionStems[base] && math.Abs(math.Mod(f.Dir+360, 180)-90) < 1 {
						continue
					}
					// The edges of the spine of 𝕊 and of its white run from one
					// bowl into the other, as the spine of S in the regular,
					// whose inflection the baseline judges intended.
					if base == "u1D54A" && sSpineInflection(rounded, f) {
						spine++
						continue
					}
					fs = append(fs, f)
				}
				if len(fs) > 0 {
					t.Errorf("%s: %d findings of the fairness analysis, the first %v", g.Name, len(fs), fs[0])
				}
				// One inflection on each edge of the spine and of its white.
				if spine > 4 {
					t.Errorf("%s: %d inflections in the spine, want at most 4", g.Name, spine)
				}
			}
		})
	}
}

// inflectionStems are the letters whose stem runs from the bracket of a serif
// straight into a round: ℙ and ℝ have no such stem, U and J have.
var inflectionStems = map[string]bool{"u1D541": true, "u1D54C": true}

// integralNames are the integrals whose stems have the two inflections of
// the S-curve near their middle.
var integralNames = map[string]bool{"uni222E": true, "uni222C": true, "uni222D": true, "uni222F": true}

// TestBigSetsPassTheChecks applies the structural checks of the regular,
// with the bounding-box rule of the glyph named by Like, and the fairness
// analysis to the big unions, circled operators and contour integrals, on
// the regular parameters; ∮ ∬ ∭ and ∯ have the two inflections of the S-curve
// of the stem of the integral, near its middle.
func TestBigSetsPassTheChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("the big unions are checked in the full tier")
	}
	for _, g := range bigSets(t, derivedSets(0)[0]) {
		if err := glyphs.CheckOutline(design.Regular(), g.Like, g.Path, g.Advance); err != nil {
			t.Errorf("%s: %v", g.Name, err)
		}
		_, y0, _, y1 := g.Path.Round().BBox()
		for _, f := range fair.Analyse(g.Name, g.Path.Round(), fair.Default).Findings {
			if integralNames[strings.TrimSuffix(g.Name, ".d")] && f.Kind == fair.Inflection && math.Abs(f.Y-(y0+y1)/2) < 0.15*(y1-y0) {
				continue
			}
			t.Errorf("%s: a finding of the fairness analysis: %v", g.Name, f)
		}
	}
}
