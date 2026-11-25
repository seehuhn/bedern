// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"slices"
	"strings"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/design/designtest"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/fair"
)

// horizontalRunes are the new code points of the horizontal constructions.
var horizontalRunes = []rune{0x305, 0x330, 0x332, 0x20D6, 0x20E1, 0x20EE, 0x20EF,
	0x23B4, 0x23B5, 0x23DC, 0x23DD, 0x23DE, 0x23DF}

// TestHorizontalConstructionsInTheFont reads the written font back and
// requires a horizontal construction for every glyph which has one, with its
// sizes in increasing order of their widths and every size the advance it
// is written with, and a top accent attachment for every size and part.
func TestHorizontalConstructionsInTheFont(t *testing.T) {
	if testing.Short() {
		t.Skip("the font is built in the full tier")
	}
	f, gs := written(t)
	id := map[string]int{}
	for i, g := range gs {
		id[g.Name] = i
	}
	for _, g := range gs {
		if !g.Horizontal {
			continue
		}
		c, ok := f.Math.Horizontal[glyphID(id[g.Name])]
		if !ok {
			t.Errorf("%s has no horizontal construction", g.Name)
			continue
		}
		if len(c.Variants) != len(g.Variants)+1 || (len(g.Parts) > 0) != (c.Assembly != nil) {
			t.Errorf("%s: the construction read back differs from that written", g.Name)
		}
		prev := -1.0
		for _, v := range c.Variants {
			x0, _, x1, _ := gs[v.Glyph].Path.Round().BBox()
			if float64(v.Advance) != x1-x0 || x1-x0 <= prev {
				t.Errorf("%s: the size %s is %v wide, written %v, after %v", g.Name, gs[v.Glyph].Name, x1-x0, v.Advance, prev)
			}
			prev = x1 - x0
		}
		names := append([]string(nil), g.Variants...)
		for _, pt := range g.Parts {
			names = append(names, pt.Name)
		}
		for _, n := range names {
			if _, ok := f.Math.TopAccent[glyphID(id[n])]; !ok {
				t.Errorf("%s has no top accent attachment", n)
			}
		}
	}
}

// horizontalOf holds the horizontal constructions by the name of their
// parameter set, each built once.
var horizontalOf sync.Map

type horizontalSet struct {
	m    map[string]Glyph
	cons map[string]construction
}

// horizontal returns the glyphs and the constructions of the horizontal
// constructions of the set.
func horizontal(t *testing.T, s designtest.Set) horizontalSet {
	t.Helper()
	v, _ := horizontalOf.LoadOrStore(s.Name, sync.OnceValues(func() (horizontalSet, error) {
		gs, cons, err := horizontalGlyphs(s.Params)
		return horizontalSet{byName(gs), cons}, err
	}))
	h, err := v.(func() (horizontalSet, error))()
	if err != nil {
		t.Fatal(err)
	}
	return h
}

// sameRuns reports whether two lists of runs agree within tol.
func sameRuns(a, b [][2]float64, tol float64) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if math.Abs(a[k][0]-b[k][0]) > tol || math.Abs(a[k][1]-b[k][1]) > tol {
			return false
		}
	}
	return true
}

// TestHorizontalAssemblies requires every part of a horizontal assembly to
// reach from the origin to its advance, every connector to be at least
// MinConnectorOverlap long, at most as long as its part and to lie on a
// straight run of the part, the ink at both ends of a joint to be the same
// runs across within one unit, the parts to be fixed at both ends of the
// assembly, and the widths of the assembly with n and with n + 1 extenders
// to overlap.  Where the construction has wider sizes, the
// assembly with no extender is no wider than the largest size.
func TestHorizontalAssemblies(t *testing.T) {
	for _, s := range derivedSets(1) {
		p := s.Params
		h := horizontal(t, s)
		overlap := math.Round(p.Punct.MathThick / 2)
		for base, c := range h.cons {
			ps := c.Parts
			if len(ps) == 0 {
				continue
			}
			if ps[0].Extender || ps[len(ps)-1].Extender {
				t.Errorf("%s: the assembly of %s has an extender at an end", s.Name, base)
			}
			for i, pt := range ps {
				g, ok := h.m[pt.Name]
				if !ok {
					t.Errorf("%s: no part %s", s.Name, pt.Name)
					continue
				}
				path := g.Path.Round()
				x0, _, x1, _ := path.BBox()
				if x0 != 0 || math.Abs(x1-pt.FullAdvance) > 0.5 || g.Advance != pt.FullAdvance {
					t.Errorf("%s: %s reaches from %v to %v with the advance %v, want 0 to %v", s.Name, pt.Name, x0, x1, g.Advance, pt.FullAdvance)
				}
				straight := func(lo, hi float64) bool {
					return sameRuns(runsV(path, lo+0.5), runsV(path, hi-0.5), 0.5)
				}
				for _, cn := range []struct {
					l      float64
					lo, hi float64
					want   bool
				}{{pt.StartConn, 0, pt.StartConn, i > 0}, {pt.EndConn, pt.FullAdvance - pt.EndConn, pt.FullAdvance, i < len(ps)-1}} {
					switch {
					case cn.want != (cn.l > 0):
						t.Errorf("%s: %s of %s has a connector %v at a free end or none at a joint", s.Name, pt.Name, base, cn.l)
					case cn.l > 0 && (cn.l < overlap || cn.l > pt.FullAdvance):
						t.Errorf("%s: a connector of %s is %v, not between %v and %v", s.Name, pt.Name, cn.l, overlap, pt.FullAdvance)
					case cn.l > 0 && !straight(cn.lo, cn.hi):
						t.Errorf("%s: %s is not straight from %v to %v", s.Name, pt.Name, cn.lo, cn.hi)
					}
				}
				if i > 0 {
					prev := h.m[ps[i-1].Name].Path.Round()
					a := runsV(prev, ps[i-1].FullAdvance-0.5)
					b := runsV(path, 0.5)
					if !sameRuns(a, b, 1) {
						t.Errorf("%s: %s ends in %v and %s starts in %v", s.Name, ps[i-1].Name, a, pt.Name, b)
					}
				}
			}
			// The widths with n extenders reach from the parts joined with
			// the longest overlaps to those joined with the shortest; those
			// with n and n + 1 extenders must overlap for n = 0 and 1.
			span := func(n int) (lo, hi float64) {
				var w float64
				joints := 0
				for _, pt := range ps {
					k := 1
					if pt.Extender {
						k = n
					}
					w += float64(k) * pt.FullAdvance
					joints += k
				}
				joints--
				var most float64
				for i := 1; i < len(ps); i++ {
					most = math.Max(most, math.Min(ps[i-1].EndConn, ps[i].StartConn))
				}
				return w - float64(joints)*most, w - float64(joints)*overlap
			}
			for n := range 2 {
				_, hi := span(n)
				lo, _ := span(n + 1)
				if lo > hi {
					t.Errorf("%s: the assembly of %s is at most %v wide with %d extenders and at least %v with %d", s.Name, base, hi, n, lo, n+1)
				}
			}
			if len(c.Variants) > 0 {
				lo, _ := span(0)
				x0, _, x1, _ := h.m[c.Variants[len(c.Variants)-1]].Path.Round().BBox()
				if lo > x1-x0+0.5 {
					t.Errorf("%s: the assembly of %s is at least %v wide, wider than its largest size %v", s.Name, base, lo, x1-x0)
				}
			}
		}
	}
}

// TestSpansJoinTheirLargestSize requires the largest size of every brace,
// parenthesis and bracket over and under to be its parts joined with the
// full connector, within one unit away from the joints, and the point of the
// middle piece of a brace to lie at its middle.
func TestSpansJoinTheirLargestSize(t *testing.T) {
	for _, s := range derivedSets(1) {
		h := horizontal(t, s)
		for _, sp := range spans {
			c := h.cons[sp.name]
			largest := h.m[c.Variants[len(c.Variants)-1]].Path.Round()
			var joined outline.Path
			var joints []float64
			x := 0.0
			for _, pt := range c.Parts {
				if pt.Extender {
					continue
				}
				if x > 0 {
					x -= pt.StartConn
					joints = append(joints, x)
				}
				joined = join(joined, shifted(h.m[pt.Name].Path.Round(), x, 0))
				x += pt.FullAdvance
			}
			lx0, _, lx1, _ := largest.BBox()
			for i := 1; i < 80; i++ {
				xx := lx0 + (lx1-lx0)*float64(i)/80
				if slices.ContainsFunc(joints, func(j float64) bool { return xx > j-1 && xx < j+hConn+1 }) {
					continue
				}
				if a, b := runsV(largest, xx), mergedRuns(runsV(joined, xx)); !sameRuns(a, b, 1) {
					t.Errorf("%s: at %v the largest %s is %v, its joined parts %v", s.Name, xx, sp.name, a, b)
				}
			}
			if sp.kind != withMiddle {
				continue
			}
			mid := h.m[sp.name+".mid"].Path.Round()
			x0, y0, x1, y1 := mid.BBox()
			point := y1
			if sp.under {
				point = y0
			}
			var at []float64
			for xx := x0; xx <= x1; xx++ {
				r := runsV(mid, xx)
				if len(r) > 0 && (!sp.under && math.Abs(r[len(r)-1][1]-point) < 0.5 || sp.under && math.Abs(r[0][0]-point) < 0.5) {
					at = append(at, xx)
				}
			}
			if len(at) == 0 || math.Abs((at[0]+at[len(at)-1])/2-(x0+x1)/2) > 1 {
				t.Errorf("%s: the point of %s.mid is not at its middle", s.Name, sp.name)
			}
		}
	}
}

// TestHorizontalSizesGrow requires the sizes of every horizontal
// construction to be wider each than the one before, the wide accents and
// the spans at the widths of wideWidths and spanWidths within one unit, the
// strokes and the heights of the wide accents not to shrink from size to
// size, and every size to stand on the line of its base: the accents over
// accentGap above AccentBaseHeight and those under accentGap below the
// baseline, the spans spanGap from it.
func TestHorizontalSizesGrow(t *testing.T) {
	for _, s := range derivedSets(1) {
		p := s.Params
		h := horizontal(t, s)
		r := newRoman(p)
		gap, abh := accentGap(r), accentBase(p)
		for base, c := range h.cons {
			var prev float64
			for k, n := range c.Variants {
				x0, _, x1, _ := h.m[n].Path.Round().BBox()
				if x1-x0 <= prev {
					t.Errorf("%s: %s is %v wide, no wider than the size before", s.Name, n, x1-x0)
				}
				prev = x1 - x0
				var want float64
				switch {
				case slices.ContainsFunc(wideAccents, func(a wideAccent) bool { return a.name == base }):
					want = wideWidths[k]
				case slices.ContainsFunc(spans, func(sp span) bool { return sp.name == base }):
					want = spanWidths[k+1]
				default:
					continue
				}
				if math.Abs(x1-x0-want) > 1 {
					t.Errorf("%s: %s is %v wide, want %v", s.Name, n, x1-x0, want)
				}
			}
		}
		for _, a := range wideAccents {
			var prevT, prevH float64
			for _, n := range h.cons[a.name].Variants {
				path := h.m[n].Path.Round()
				x0, y0, x1, y1 := path.BBox()
				if a.below {
					if y1 != -gap {
						t.Errorf("%s: %s hangs from %v, want %v", s.Name, n, y1, -gap)
					}
				} else if y0 != abh+gap {
					t.Errorf("%s: %s stands on %v, want %v", s.Name, n, y0, abh+gap)
				}
				// The stroke, upright at the middle of the
				// tilde and the breve and at a quarter of the circumflex
				// and the caron, where it is straight.
				at := (x0 + x1) / 2
				if a.mark == "uni0302" || a.mark == "uni030C" {
					at = x0 + (x1-x0)/4
				}
				rs, ra, rb := runsV(path, at), runsV(path, at-20), runsV(path, at+20)
				if len(rs) == 0 || len(ra) == 0 || len(rb) == 0 {
					t.Errorf("%s: no ink in %s at %v", s.Name, n, at)
					continue
				}
				// Across the stroke: the upright run times the cosine of
				// its slope.
				slope := ((rb[0][0] + rb[0][1]) - (ra[0][0] + ra[0][1])) / 80
				th := (rs[0][1] - rs[0][0]) * math.Cos(math.Atan(slope))
				if th < prevT-1 || y1-y0 < prevH-1 {
					t.Errorf("%s: %s is %v thick and %v high, less than the size before (%v, %v)", s.Name, n, th, y1-y0, prevT, prevH)
				}
				prevT, prevH = th, y1-y0
			}
		}
		for _, sp := range spans {
			for _, n := range append([]string{sp.name}, h.cons[sp.name].Variants...) {
				_, y0, _, y1 := h.m[n].Path.Round().BBox()
				if sp.under && y1 != -spanGap(p) || !sp.under && y0 != abh+spanGap(p) {
					t.Errorf("%s: %s reaches from %v to %v, %v from its argument", s.Name, n, y0, y1, spanGap(p))
				}
			}
		}
	}
}

// TestSpansHaveTheStrokesOfTheBase requires every size and every extender
// of the braces over and under to have the arm of the brace of the regular,
// within one unit, and the brackets their bar bracketWeight times the stem
// of the bracket.
func TestSpansHaveTheStrokesOfTheBase(t *testing.T) {
	for _, s := range derivedSets(1) {
		p := s.Params
		h := horizontal(t, s)
		for _, sp := range spans {
			var want float64
			switch sp.parent {
			case "braceleft", "braceright":
				want = p.Sym.BraceArm
			case "bracketleft", "bracketright":
				want = bracketWeight * p.Punct.BracketStem
			default:
				continue
			}
			names := append(append([]string{sp.name}, h.cons[sp.name].Variants...), sp.name+".ext")
			for _, n := range names {
				path := h.m[n].Path.Round()
				x0, _, x1, _ := path.BBox()
				at := x0 + (x1-x0)/4
				if strings.HasSuffix(n, ".ext") {
					at = (x0 + x1) / 2
				}
				rs := runsV(path, at)
				if len(rs) != 1 || math.Abs(rs[0][1]-rs[0][0]-want) > 1 {
					t.Errorf("%s: %s has the runs %v at %v, want one %v thick", s.Name, n, rs, at, want)
				}
			}
		}
	}
}

// TestHorizontalGlyphsPassTheChecks applies the structural checks of the
// regular, with the bounding-box rule of the glyph named by Like, and the
// fairness analysis to every size and part of the horizontal constructions,
// on the regular parameters; a size or a part has no more findings of each
// kind than its base has (the mark of the regular for the wide accents, the
// vertical delimiter for the spans), which are the intended inflections of
// the tilde and of the S-curves of the brace.
func TestHorizontalGlyphsPassTheChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("the horizontal constructions are checked in the full tier")
	}
	p := design.Regular()
	s := designtest.Set{Name: "regular", Params: p}
	h := horizontal(t, s)
	r := newRoman(p)
	allowed := map[string]map[fair.Kind]int{}
	count := func(name string, path outline.Path) map[fair.Kind]int {
		m := map[fair.Kind]int{}
		for _, f := range fair.Analyse(name, path.Round(), fair.Default).Findings {
			m[f.Kind]++
		}
		return m
	}
	for _, a := range wideAccents {
		path, _ := r.get(a.mark)
		allowed[a.name] = count(a.mark, path)
	}
	for _, sp := range spans {
		path, _ := r.get(sp.parent)
		allowed[sp.name] = count(sp.parent, path)
	}
	for n, g := range h.m {
		if err := glyphs.CheckOutline(p, g.Like, g.Path, g.Advance); err != nil {
			t.Errorf("%s: %v", n, err)
		}
		base, _, _ := strings.Cut(n, ".")
		for k, c := range count(n, g.Path) {
			if c > allowed[base][k] {
				t.Errorf("%s: %d findings %v of the fairness analysis, its base has %d", n, c, k, allowed[base][k])
			}
		}
	}
}
