// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"slices"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The serifs on the diagonal strokes of v, w, x, y and k.  A head serif is a flat
// slab whose top lies on the x-height from tip to tip; its tip stands at the same
// distance from the edge of the stroke on both sides, measured at the level of the
// underside; and the serifs of neighbouring strokes keep a clear gap.

const (
	// diagOverhangBalance is the greatest difference, in units, between the overhangs
	// on the two sides of one stroke.
	diagOverhangBalance = 3.0
	// diagPairBalance is the greatest difference between the total widths of the serif
	// of the thick stroke and that of the thin stroke of a letter, serif and stroke
	// together.
	diagPairBalance = 30.0
	// diagSerifGapMin is the least white between the tips of the serifs of
	// neighbouring strokes at the level of the serifs.
	diagSerifGapMin = 60.0
)

// A serifSite is one serif of a letter: the stroke, its edge, and the height of the
// outer face of the slab (the x-height for a head serif, the baseline for a foot).
type serifSite struct {
	name  string
	d     diagonal
	e     edge
	foot  bool
	group string // the strokes which share a total: "thick" or "thin"
}

// serifSites returns the path of the letter and the serif sites it has.
func serifSites(t *testing.T, glyph string, p design.Params) (outline.Path, []serifSite) {
	t.Helper()
	path, _, ok := outlineOf(glyph, p)
	if !ok {
		t.Fatalf("glyph %s not found", glyph)
	}
	var sites []serifSite
	switch glyph {
	case "v", "y":
		vx := vVertex(p)
		sites = append(sites,
			serifSite{"thick diagonal, left", vx.Left, edgeLeft, false, "thick"},
			serifSite{"thick diagonal, right", vx.Left, edgeRight, false, "thick"},
			serifSite{"thin diagonal, left", vx.Right, edgeLeft, false, "thin"},
			serifSite{"thin diagonal, right", vx.Right, edgeRight, false, "thin"})
	case "w":
		v1, v2 := wVertices(p)
		sites = append(sites,
			serifSite{"first diagonal, left", v1.Left, edgeLeft, false, "thick"},
			serifSite{"first diagonal, right", v1.Left, edgeRight, false, "thick"},
			serifSite{"last diagonal, left", v2.Right, edgeLeft, false, "thin"},
			serifSite{"last diagonal, right", v2.Right, edgeRight, false, "thin"})
	case "x":
		thick, thinU, thinL := xDiagonals(p)
		sites = append(sites,
			serifSite{"upper thick, left", thick, edgeLeft, false, "thick"},
			serifSite{"upper thick, right", thick, edgeRight, false, "thick"},
			serifSite{"upper thin, left", thinU, edgeLeft, false, "thin"},
			serifSite{"upper thin, right", thinU, edgeRight, false, "thin"},
			serifSite{"lower thin, left", thinL, edgeLeft, true, "thin"},
			serifSite{"lower thin, right", thinL, edgeRight, true, "thin"},
			serifSite{"lower thick, left", thick, edgeLeft, true, "thick"},
			serifSite{"lower thick, right", thick, edgeRight, true, "thick"})
	case "k":
		arm, leg := kDiagonals(p)
		sites = append(sites,
			serifSite{"arm, left", arm, edgeLeft, false, "thin"},
			serifSite{"arm, right", arm, edgeRight, false, "thin"},
			serifSite{"leg, left", leg, edgeLeft, true, "thick"},
			serifSite{"leg, right", leg, edgeRight, true, "thick"})
	}
	return path, sites
}

// tipAt returns the x of the tip of the serif of the site: the crossing of the
// horizontal line through the middle of the slab which lies nearest to the edge on the
// side of the serif.
func (s serifSite) tipAt(t *testing.T, path outline.Path, p design.Params) (tip, over float64) {
	t.Helper()
	side := 1.0
	if s.e == edgeLeft {
		side = -1
	}
	under, mid := p.XHeight-p.SerifThick, p.XHeight-p.SerifThick/2
	if s.foot {
		under, mid = p.SerifThick, p.SerifThick/2
	}
	edgeMid := s.d.edgeAt(s.e, mid).X
	best := math.Inf(1)
	for _, x := range path.CrossingsX(mid) {
		if side*(x-edgeMid) > 5 && side*(x-edgeMid) < best {
			best = side * (x - edgeMid)
			tip = x
		}
	}
	if math.IsInf(best, 1) {
		t.Fatalf("%s: no tip found", s.name)
	}
	return tip, side * (tip - s.d.edgeAt(s.e, under).X)
}

// diagSerifSets are the parameter sets in which the serifs are checked: the regular
// ones and the perturbed ones which change the diagonal letters.
func diagSerifSets() map[string]design.Params {
	sets := map[string]design.Params{"regular": design.Regular()}
	for _, s := range allParams() {
		sets[s.name] = s.params
	}
	return sets
}

// The top of every head serif of a diagonal is horizontal and lies on the x-height
// from tip to tip: nothing of the letter stands above it there, and the ink reaches
// up to the line all along it.
func TestDiagonalHeadSerifTopsAreFlatOnTheXHeight(t *testing.T) {
	t.Parallel()
	for name, p := range diagSerifSets() {
		for _, g := range []string{"v", "w", "x", "y", "k"} {
			path, sites := serifSites(t, g, p)
			for _, s := range sites {
				if s.foot {
					continue
				}
				tip, _ := s.tipAt(t, path, p)
				side := 1.0
				if s.e == edgeLeft {
					side = -1
				}
				// From just inside the tip to the edge extended to the top, the
				// top of the ink is the line.
				reach := math.Abs(tip - s.d.edgeAt(s.e, p.XHeight).X)
				for i := 0; i <= 10; i++ {
					x := tip - side*(0.01+(reach-0.01)*float64(i)/10)
					ys := path.CrossingsY(x)
					if len(ys) == 0 {
						t.Fatalf("%s/%s/%s: nothing at x = %.1f", name, g, s.name, x)
					}
					if top := ys[len(ys)-1]; math.Abs(top-p.XHeight) > 1e-6 {
						t.Errorf("%s/%s/%s: the top of the serif is at %.4f at x = %.1f, want the x-height %v", name, g, s.name, top, x, p.XHeight)
					}
				}
			}
		}
	}
}

// The overhangs of the two sides of a stroke differ by at most diagOverhangBalance,
// and the thick and the thin stroke of a letter carry serifs of about the same total
// width.
func TestDiagonalSerifsAreBalanced(t *testing.T) {
	t.Parallel()
	for name, p := range diagSerifSets() {
		for _, g := range []string{"v", "w", "x", "y", "k"} {
			path, sites := serifSites(t, g, p)
			totals := map[string][]float64{}
			for i := 0; i+1 < len(sites); i += 2 {
				l, r := sites[i], sites[i+1]
				_, ol := l.tipAt(t, path, p)
				_, or := r.tipAt(t, path, p)
				if math.Abs(ol-or) > diagOverhangBalance {
					t.Errorf("%s/%s: the serif on the stroke %q overhangs %.1f on the left and %.1f on the right, want a difference of at most %v",
						name, g, l.name, ol, or, diagOverhangBalance)
				}
				under := p.XHeight - p.SerifThick
				if l.foot {
					under = p.SerifThick
				}
				width := l.d.edgeAt(edgeRight, under).X - l.d.edgeAt(edgeLeft, under).X
				totals[l.group] = append(totals[l.group], width+ol+or)
			}
			// The rule applies to the regular parameter set only.
			if tk, tn := totals["thick"], totals["thin"]; name == "regular" && len(tk) > 0 && len(tn) > 0 {
				if math.Abs(slices.Max(tk)-slices.Max(tn)) > diagPairBalance {
					t.Errorf("%s/%s: the serif of the thick stroke is %.1f wide and that of the thin one %.1f, want a difference of at most %v",
						name, g, slices.Max(tk), slices.Max(tn), diagPairBalance)
				}
			}
		}
	}
}

// The serifs of neighbouring strokes keep a clear white between their tips at the
// level of the serifs in the regular face: the inner serifs of v, y and x (at the head
// and at the feet), and the inner serif of w and the stroke which follows it.
func TestDiagonalSerifsKeepApart(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, g := range []string{"v", "w", "x", "y"} {
		path, _, _ := outlineOf(g, p)
		levels := []float64{p.XHeight - p.SerifThick/2}
		if g == "x" {
			levels = append(levels, p.SerifThick/2)
		}
		for _, y := range levels {
			xs := path.CrossingsX(y)
			for i := 1; i+1 < len(xs); i += 2 {
				if gap := xs[i+1] - xs[i]; gap < diagSerifGapMin {
					t.Errorf("%s: the white at height %.1f between x = %.1f and %.1f is %.1f wide, want at least %v",
						g, y, xs[i], xs[i+1], gap, diagSerifGapMin)
				}
			}
		}
	}
}
