// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import (
	"strings"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
)

// publicSets returns the regular parameters and, outside -short, the bold.
func publicSets() []named {
	sets := []named{{"regular", design.Regular()}}
	if !testing.Short() {
		sets = append(sets, named{"bold", design.Bold()})
	}
	return sets
}

// byName returns the glyphs of the font by name.
func byName(f *Font) map[string]*Glyph {
	m := make(map[string]*Glyph, len(f.Glyphs))
	for _, g := range f.Glyphs {
		m[g.Name] = g
	}
	return m
}

// TestTabularFiguresAreNotKerned: the figures which keep their advance (the
// tabular lining and old-style figures) have no kern with any glyph, on either
// side, so that a column of figures keeps its alignment; the proportional
// figures are kerned, the lining ones as capitals and the old-style ones as
// lowercase letters.
func TestTabularFiguresAreNotKerned(t *testing.T) {
	t.Parallel()
	for _, s := range publicSets() {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			tb := computed(s.p)
			f := tb.Font
			n := 0
			for _, g := range f.Glyphs {
				isFigure, tabular := false, false
				for _, d := range figureNames {
					switch g.Name {
					case d, d + ".taboldstyle":
						isFigure, tabular = true, true
					case d + ".prop", d + ".oldstyle":
						isFigure = true
					}
				}
				if !isFigure {
					continue
				}
				n++
				switch {
				case tabular && g.Kind != Skip:
					t.Errorf("%s: kind %v, want skip", g.Name, g.Kind)
				case !tabular && strings.HasSuffix(g.Name, ".prop") && g.Kind != Upper:
					t.Errorf("%s: kind %v, want upper", g.Name, g.Kind)
				case !tabular && strings.HasSuffix(g.Name, ".oldstyle") && g.Kind != Lower:
					t.Errorf("%s: kind %v, want lower", g.Name, g.Kind)
				}
				if !tabular {
					continue
				}
				for _, o := range f.Glyphs {
					if k := tb.Kern(g.Index, o.Index); k != 0 {
						t.Errorf("%s %s: kern %d, want none", g.Name, o.Name, k)
					}
					if k := tb.Kern(o.Index, g.Index); k != 0 {
						t.Errorf("%s %s: kern %d, want none", o.Name, g.Name, k)
					}
				}
			}
			if n != 40 {
				t.Errorf("found %d figures, want 40", n)
			}
		})
	}
}

// TestProportionalFiguresAreKerned: the proportional figures take kerns by the
// model of the capitals: 1 before 1 and 7, and 7 before the full stop and the
// comma, in the lining and the old-style form.
func TestProportionalFiguresAreKerned(t *testing.T) {
	t.Parallel()
	for _, s := range publicSets() {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			tb := computed(s.p)
			by := byName(tb.Font)
			for _, pair := range [][2]string{
				{"one.prop", "one.prop"}, {"one.prop", "seven.prop"},
				{"seven.prop", "period"}, {"seven.prop", "comma"},
				{"seven.oldstyle", "period"}, {"seven.oldstyle", "comma"},
			} {
				a, b := by[pair[0]], by[pair[1]]
				if k := tb.Kern(a.Index, b.Index); k >= 0 {
					t.Errorf("%s %s: kern %d, want a negative kern", pair[0], pair[1], k)
				}
			}
			for _, pair := range [][2]string{{"one.prop", "period"}, {"one.prop", "comma"}} {
				a, b := by[pair[0]], by[pair[1]]
				if k := tb.Kern(a.Index, b.Index); k > 0 {
					t.Errorf("%s %s: kern %d, want no positive kern", pair[0], pair[1], k)
				}
			}
		})
	}
}

// TestSignsAreKerned: the ampersand takes kerns as a capital does, before T
// and V, and the currency signs and the at sign are in the model, so that
// their pairs keep the floor and the other limits of the kerns.
func TestSignsAreKerned(t *testing.T) {
	t.Parallel()
	for _, s := range publicSets() {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			tb := computed(s.p)
			f := tb.Font
			by := byName(f)
			for _, r := range "&@€£" {
				if g := f.glyphOf(r); g == nil || g.Kind != Upper {
					t.Errorf("%c: not kerned as a capital", r)
				}
			}
			for _, b := range []string{"T", "V"} {
				if k := tb.Kern(by["ampersand"].Index, by[b].Index); k >= 0 {
					t.Errorf("& %s: kern %d, want a negative kern", b, k)
				}
			}
			for _, a := range []string{"ampersand", "at", "Euro", "sterling"} {
				for _, o := range f.Glyphs {
					if o.Kind == Skip {
						continue
					}
					for _, pr := range [2][2]*Glyph{{by[a], o}, {o, by[a]}} {
						k := tb.Kern(pr[0].Index, pr[1].Index)
						if f.Distance(pr[0], pr[1], float64(k)) < s.p.Kern.Floor-2 {
							t.Errorf("%s %s with kern %d: closest distance below the floor", pr[0].Name, pr[1].Name, k)
						}
					}
				}
			}
		})
	}
}

// TestBracesKeepTheFloorAtJ: the braces next to j and f, whose hooks come
// close to them, keep the floor, and a pair which is closer than the floor
// without a kern takes a positive one.
func TestBracesKeepTheFloorAtJ(t *testing.T) {
	t.Parallel()
	for _, s := range publicSets() {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			tb := computed(s.p)
			f := tb.Font
			by := byName(f)
			for _, br := range []string{"braceleft", "braceright"} {
				for _, l := range []string{"j", "f", "jdotless"} {
					a, b := by[br], by[l]
					if b == nil {
						continue
					}
					for _, pr := range [2][2]*Glyph{{a, b}, {b, a}} {
						k := tb.Kern(pr[0].Index, pr[1].Index)
						if d := f.Distance(pr[0], pr[1], float64(k)); d < s.p.Kern.Floor-2 {
							t.Errorf("%s %s with kern %d: closest distance %.1f, floor %v", pr[0].Name, pr[1].Name, k, d, s.p.Kern.Floor)
						}
						if d0 := f.Distance(pr[0], pr[1], 0); d0 < s.p.Kern.Floor && k <= 0 {
							t.Errorf("%s %s: %.1f apart without a kern, kern %d", pr[0].Name, pr[1].Name, d0, k)
						}
					}
				}
			}
		})
	}
}

// TestCapitalsKeepClearOfDashes: a capital which is closer than
// design.Kern.DashClear to a hyphen or a dash without a kern takes no
// negative kern before it, and one which is further away is left at least that
// distance.  L, whose serif of the arm stands as near to the dashes as the arm
// of E, has the same kern before them in every weight.
func TestCapitalsKeepClearOfDashes(t *testing.T) {
	t.Parallel()
	var lkerns []int
	for _, s := range publicSets() {
		tb := computed(s.p)
		f := tb.Font
		clear := s.p.Kern.DashClear
		for _, a := range f.Glyphs {
			if a.Kind != Upper {
				continue
			}
			for _, r := range dashes {
				b := f.glyphOf(r)
				k := tb.Kern(a.Index, b.Index)
				if k >= 0 {
					continue
				}
				if d := f.Distance(a, b, float64(k)); d < clear-2 {
					t.Errorf("%s: %s %s with kern %d: closest distance %.1f, want at least %v", s.name, a.Name, b.Name, k, d, clear)
				}
			}
		}
		l := f.glyphOf('L')
		for _, r := range dashes {
			lkerns = append(lkerns, tb.Kern(l.Index, f.glyphOf(r).Index))
		}
	}
	for _, k := range lkerns {
		if k != lkerns[0] {
			t.Errorf("L before the dashes: kerns %v differ between the dashes or the weights", lkerns)
			break
		}
	}
}
