// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/design/designtest"
	"seehuhn.de/go/bedern/internal/outline"
)

// punctSets returns the regular parameters and the perturbed sets of the tier.
func punctSets() []designtest.Set {
	return append([]designtest.Set{{Name: "regular", Params: design.Regular()}}, designtest.Sets(testing.Short(), routineSets)...)
}

func punctOutline(t *testing.T, name string, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, adv, ok := outlineOf(name, p)
	if !ok {
		t.Fatalf("no glyph %s", name)
	}
	return path, adv
}

// The dots of the full stop, the colon, the semicolon and the exclamation mark are
// the dot of i: circles of DotDiameter, the lower dots at one height.
func TestPunctDotsAreTheDotOfI(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		p := s.Params
		for _, name := range []string{"period", "colon", "exclam", "semicolon"} {
			path, _ := punctOutline(t, name, p)
			found := 0
			for _, ct := range path {
				c := outline.Path{ct}
				x0, y0, x1, y1 := c.BBox()
				if math.Abs((x1-x0)-p.DotDiameter) > 0.01 || math.Abs((y1-y0)-p.DotDiameter) > 0.01 {
					continue
				}
				found++
				cy := (y0 + y1) / 2
				if cy != p.PeriodCentre() && cy != p.ColonUpperCentre() {
					t.Errorf("%s, %s: a dot is centred at the height %v", s.Name, name, cy)
				}
			}
			want := 1
			if name == "colon" {
				want = 2
			}
			if found != want {
				t.Errorf("%s, %s: %d dots of the size of the dot of i, want %d", s.Name, name, found, want)
			}
		}
	}
}

// The tail of the comma curves to the left only: its outer edge turns to the left
// all the way from the tip to the head, the curvature of the outer edge rises
// from the tip to the head evenly (it falls nowhere by more than six per cent),
// and the inner edge turns to the right only.
func TestCommaTailIsSmooth(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		p := s.Params
		ch := commaAt(p, 0, commaTop(p))
		// The pieces after the arc: the inner edge, the cut and the outer edge.
		var inner, outer []cubic
		seenCut := false
		for _, pc := range ch {
			if pc.line {
				seenCut = true
				continue
			}
			x, y := pc.P0, pc.P3
			if math.Hypot(x.X, x.Y-(commaTop(p)-p.Punct.CommaHead/2)) > p.Punct.CommaHead/2+0.01 ||
				math.Hypot(y.X, y.Y-(commaTop(p)-p.Punct.CommaHead/2)) > p.Punct.CommaHead/2+0.01 {
				if seenCut {
					outer = append(outer, pc.cubic)
				} else {
					inner = append(inner, pc.cubic)
				}
			}
		}
		if len(inner) == 0 || len(outer) == 0 {
			t.Fatalf("%s: no tail found", s.Name)
		}
		var ks []float64
		for _, c := range outer {
			for k := 0; k <= 50; k++ {
				ks = append(ks, c.curvature(float64(k)/50))
			}
		}
		for i, k := range ks {
			if k < -1e-9 {
				t.Errorf("%s: the outer edge of the tail turns to the right (curvature %g)", s.Name, k)
				break
			}
			// Along the contour the outer edge runs from the tip to the head.
			// A cubic curve which runs into the circle with its curvature dips a
			// little before the head, by at most a twentieth.
			if i > 0 && k < ks[i-1]*0.94 {
				t.Errorf("%s: the curvature of the outer edge falls towards the head (%g after %g)", s.Name, k, ks[i-1])
				break
			}
		}
		for _, c := range inner {
			for k := 0; k <= 50; k++ {
				if kk := c.curvature(float64(k) / 50); kk > 1e-9 {
					t.Errorf("%s: the inner edge of the tail turns to the left (curvature %g)", s.Name, kk)
				}
			}
		}
	}
}

// The quotation marks are the comma: the right single mark is the comma
// translated, and the left one has the same box.
func TestQuotesAreTheComma(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	c, _ := punctOutline(t, "comma", p)
	r, _ := punctOutline(t, "quoteright", p)
	l, _ := punctOutline(t, "quoteleft", p)
	cx0, cy0, cx1, cy1 := c.BBox()
	rx0, ry0, rx1, ry1 := r.BBox()
	lx0, ly0, lx1, ly1 := l.BBox()
	if math.Abs((cx1-cx0)-(rx1-rx0)) > 1e-6 || math.Abs((cy1-cy0)-(ry1-ry0)) > 1e-6 {
		t.Errorf("the right quotation mark has the box %vx%v, the comma %vx%v", rx1-rx0, ry1-ry0, cx1-cx0, cy1-cy0)
	}
	if math.Abs((lx1-lx0)-(rx1-rx0)) > 1e-6 || math.Abs((ly1-ly0)-(ry1-ry0)) > 1e-6 || math.Abs(ly1-ry1) > 1e-6 {
		t.Errorf("the left quotation mark does not have the box of the right one")
	}
	if math.Abs(ry1-p.Punct.QuoteTop) > 1e-6 {
		t.Errorf("the top of the quotation mark is %v, want %v", ry1, p.Punct.QuoteTop)
	}
	// The head of the left mark is at the bottom: a scan through the middle of the
	// head, at its height, crosses the head's full width.
	if n := len(l[0].Segs); n != len(r[0].Segs) {
		t.Errorf("the left mark has %d segments, the right one %d", n, len(r[0].Segs))
	}
}

// The parentheses have no straight stretch and no jump of curvature: along each of
// the two edges the curvature keeps its sign, stays above a floor and varies by
// less than a factor, and the right parenthesis is the mirror image of the left.
func TestParenthesisCurvature(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		p := s.Params
		ch := parenLeft(p)
		var curved []cubic
		for _, pc := range ch {
			if !pc.line {
				curved = append(curved, pc.cubic)
			}
		}
		if len(curved) != 4 {
			t.Fatalf("%s: %d curves, want 4", s.Name, len(curved))
		}
		for edge := 0; edge < 2; edge++ {
			lo, hi := math.Inf(1), 0.0
			sign := 0.0
			for _, c := range curved[2*edge : 2*edge+2] {
				for k := 0; k <= 100; k++ {
					v := c.curvature(float64(k) / 100)
					if sign == 0 {
						sign = math.Copysign(1, v)
					}
					if v*sign <= 0 {
						t.Fatalf("%s: the curvature of edge %d changes its sign", s.Name, edge)
					}
					lo, hi = math.Min(lo, math.Abs(v)), math.Max(hi, math.Abs(v))
				}
			}
			if lo < 1.0/4000 {
				t.Errorf("%s: edge %d has a nearly straight stretch, curvature %g", s.Name, edge, lo)
			}
			if hi/lo > 6 {
				t.Errorf("%s: the curvature of edge %d varies by the factor %.1f", s.Name, edge, hi/lo)
			}
		}
		l, la := punctOutline(t, "parenleft", p)
		r, ra := punctOutline(t, "parenright", p)
		lx0, ly0, lx1, ly1 := l.BBox()
		rx0, ry0, rx1, ry1 := r.BBox()
		if la != ra || math.Abs((lx1-lx0)-(rx1-rx0)) > 1e-6 || ly0 != ry0 || ly1 != ry1 {
			t.Errorf("%s: the parentheses are not mirror images", s.Name)
		}
		if math.Abs(lx0-(ra-rx1)) > 1e-6 {
			t.Errorf("%s: the sides of the parentheses are not exchanged", s.Name)
		}
	}
}

// Parentheses, brackets and the vertical bar enclose capitals, ascenders and
// descenders, and have the same height.
func TestBracketsEnclose(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		p := s.Params
		for _, name := range []string{"parenleft", "parenright", "bracketleft", "bracketright", "bar"} {
			path, _ := punctOutline(t, name, p)
			_, y0, _, y1 := path.BBox()
			if y1 < math.Max(p.CapHeight, p.Ascender) || y0 > -p.Descender {
				t.Errorf("%s, %s: reaches from %v to %v and does not enclose the letters", s.Name, name, y0, y1)
			}
			if math.Abs(y0-p.BracketBottom()) > 1e-6 || math.Abs(y1-p.BracketTop()) > 1e-6 {
				t.Errorf("%s, %s: reaches from %v to %v, want %v to %v", s.Name, name, y0, y1, p.BracketBottom(), p.BracketTop())
			}
		}
	}
}

// The five mathematical signs share one advance and one axis: the plus is a square
// whose bars cross on the axis, the minus is as wide as the plus, the equals is
// symmetric about the axis, and the less-than and greater-than are centred on it.
func TestMathSigns(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		p := s.Params
		axis := p.Punct.MathAxis
		var adv0 float64
		for i, name := range []string{"plus", "minus", "equal", "less", "greater"} {
			path, adv := punctOutline(t, name, p)
			if i == 0 {
				adv0 = adv
			} else if adv != adv0 {
				t.Errorf("%s: %s has the advance %v, the plus %v", s.Name, name, adv, adv0)
			}
			x0, y0, x1, y1 := path.BBox()
			if math.Abs((y0+y1)/2-axis) > 0.01 {
				t.Errorf("%s: %s is centred at %v, want the axis %v", s.Name, name, (y0+y1)/2, axis)
			}
			if math.Abs((x0+x1)/2-adv/2) > 0.51 {
				t.Errorf("%s: %s is centred at %v in the advance %v", s.Name, name, (x0+x1)/2, adv)
			}
			switch name {
			case "plus":
				if math.Abs((x1-x0)-(y1-y0)) > 0.01 {
					t.Errorf("%s: the plus is %vx%v, want a square", s.Name, x1-x0, y1-y0)
				}
			case "minus":
				if math.Abs((x1-x0)-p.Punct.MathWidth) > 0.01 || math.Abs((y1-y0)-p.Punct.MathThick) > 0.01 {
					t.Errorf("%s: the minus is %vx%v", s.Name, x1-x0, y1-y0)
				}
			}
		}
		if adv0 < p.Punct.MathWidth {
			t.Errorf("%s: the advance %v is less than the width of the signs", s.Name, adv0)
		}
	}
}

// The dashes are plain bars on the axis with side space, the en dash as wide as
// the figures with their space, the em dash
// longer than the en dash; the hyphen is lower than the dashes (at half the
// x-height of the regular parameters).
func TestDashes(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	h, _ := punctOutline(t, "hyphen", p)
	en, enAdv := punctOutline(t, "endash", p)
	_, emAdv := punctOutline(t, "emdash", p)
	_, hy0, _, hy1 := h.BBox()
	_, ey0, _, ey1 := en.BBox()
	if enAdv != p.Figures.Advance || emAdv <= enAdv {
		t.Errorf("the advances of the dashes are %v and %v, the advance of the figures is %v", enAdv, emAdv, p.Figures.Advance)
	}
	if (hy0+hy1)/2 >= (ey0+ey1)/2 {
		t.Errorf("the hyphen at %v is not below the dashes at %v", (hy0+hy1)/2, (ey0+ey1)/2)
	}
	if math.Abs((hy0+hy1)/2-p.XHeight/2) > 6 {
		t.Errorf("the hyphen is centred at %v, half the x-height is %v", (hy0+hy1)/2, p.XHeight/2)
	}
}

// The dashes keep side space: at least 40 on each side of the en dash and 30 on
// each side of the em dash (the gap between two dashes next to each other is at
// least 70).
func TestDashSideSpace(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		p := s.Params
		for _, c := range []struct {
			name string
			min  float64
		}{{"endash", 40}, {"emdash", 30}} {
			path, adv := punctOutline(t, c.name, p)
			x0, _, x1, _ := path.BBox()
			if x0 < c.min-0.5 || adv-x1 < c.min-0.5 {
				t.Errorf("%s: %s has side spaces %v and %v, want at least %v", s.Name, c.name, x0, adv-x1, c.min)
			}
		}
	}
}

// The hyphen and the dashes are one weight: their thicknesses differ by no more
// than 8 units, and the dashes are not thicker than the minus.
func TestHyphenAndDashesAreOneWeight(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		if s.Name != "regular" { // the variants of the tests perturb the thicknesses
			continue
		}
		p := s.Params
		var th []float64
		for _, name := range []string{"hyphen", "endash", "emdash"} {
			path, _ := punctOutline(t, name, p)
			_, y0, _, y1 := path.BBox()
			th = append(th, y1-y0)
		}
		if math.Abs(th[0]-th[1]) > 8 || math.Abs(th[1]-th[2]) > 0.5 {
			t.Errorf("%s: the thicknesses of hyphen, en dash and em dash are %v", s.Name, th)
		}
		if th[1] > p.Punct.MathThick+0.5 {
			t.Errorf("%s: the dashes %v are thicker than the minus (%v)", s.Name, th, p.Punct.MathThick)
		}
	}
}

// .notdef is a frame with a diagonal cross: one outer contour and four triangular
// counters, from the baseline to the height of the figures, on the advance of the
// figures.
func TestNotdef(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		p := s.Params
		path, adv := punctOutline(t, ".notdef", p)
		if len(path) != 5 {
			t.Fatalf("%s: .notdef has %d contours, want 5", s.Name, len(path))
		}
		for i, ct := range path[1:] {
			if len(ct.Segs) != 3 {
				t.Errorf("%s: counter %d has %d sides, want a triangle", s.Name, i, len(ct.Segs))
			}
		}
		x0, y0, x1, y1 := path.BBox()
		if y0 != 0 || y1 != p.Figures.Height || adv != p.Figures.Advance {
			t.Errorf("%s: .notdef is %vx%v from %v with advance %v, want the height %v and the advance %v of the figures", s.Name, x1-x0, y1-y0, y0, adv, p.Figures.Height, p.Figures.Advance)
		}
		if l, r := x0, adv-x1; math.Abs(l-r) > 1 {
			t.Errorf("%s: .notdef has side spaces %v and %v", s.Name, l, r)
		}
	}
}
