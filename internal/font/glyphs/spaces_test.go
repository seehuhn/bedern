// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"reflect"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// TestSpacesOfFixedWidth asks, in every parameter set, that the spaces have
// no outline; that the figure space has the advance of the figures and the
// punctuation space that of the full stop; that the en and the em spaces and
// quads are a half and the whole of the em and the spaces per em and the
// medium mathematical space their fractions of it, within half a unit; and
// that the compound word mark has neither an outline nor a width.
func TestSpacesOfFixedWidth(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			_, zero := cachedOutline(np, "zero")
			_, stop := cachedOutline(np, "period")
			em := p.UnitsPerEm
			for name, want := range map[string]float64{
				"uni2000": em / 2, "uni2001": em, "uni2002": em / 2, "uni2003": em,
				"uni2004": em / 3, "uni2005": em / 4, "uni2006": em / 6,
				"uni2007": zero, "uni2008": stop, "uni205F": em * 4 / 18,
				"cwm": 0,
			} {
				path, adv := cachedOutline(np, name)
				if len(path) != 0 {
					t.Errorf("%s has an outline", name)
				}
				if math.Abs(adv-want) > 0.5 || adv != math.Round(adv) {
					t.Errorf("%s: advance %v, want %.1f as a whole number", name, adv, want)
				}
			}
		})
	}
}

// TestHyphensAndDashes asks, in every parameter set, that the hyphen and the
// non-breaking hyphen have the outline and the advance of the hyphen-minus;
// that the figure dash has the advance of the figures, stands at the height
// of the en dash and has its ink in the middle of its advance; and that the
// horizontal bar stands at that height and is at least as long as the em dash.
func TestHyphensAndDashes(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			hyphen, hadv := cachedOutline(np, "hyphen")
			for _, name := range []string{"uni2010", "uni2011"} {
				path, adv := cachedOutline(np, name)
				if adv != hadv || !reflect.DeepEqual(path, hyphen) {
					t.Errorf("%s differs from the hyphen", name)
				}
			}
			en, _ := cachedOutline(np, "endash")
			_, ey0, _, ey1 := en.BBox()
			em, _ := cachedOutline(np, "emdash")
			mx0, _, mx1, _ := em.BBox()
			_, zero := cachedOutline(np, "zero")

			fig, adv := cachedOutline(np, "figuredash")
			x0, y0, x1, y1 := fig.BBox()
			if adv != zero {
				t.Errorf("the figure dash has the advance %v, and the figures %v", adv, zero)
			}
			if y0 != ey0 || y1 != ey1 {
				t.Errorf("the figure dash stands from %v to %v, and the en dash from %v to %v", y0, y1, ey0, ey1)
			}
			if math.Abs(x0-(adv-x1)) > 1 {
				t.Errorf("the figure dash has the side spaces %v and %v", x0, adv-x1)
			}

			bar, _ := cachedOutline(np, "uni2015")
			x0, y0, x1, y1 = bar.BBox()
			if y0 != ey0 || y1 != ey1 {
				t.Errorf("the horizontal bar stands from %v to %v, and the en dash from %v to %v", y0, y1, ey0, ey1)
			}
			if x1-x0 < mx1-mx0 {
				t.Errorf("the horizontal bar is %v long, shorter than the em dash, %v", x1-x0, mx1-mx0)
			}
		})
	}
}

// TestPerThousandZeroCompletesThePercent asks, in every parameter set, that
// the percent sign followed by the zero of the per mille sign is the per mille
// sign, and the per mille sign followed by that zero the per ten thousand
// sign: the same contours and the same advance.  The primes keep one pitch:
// the triple prime is as much wider than the double prime as that is wider
// than the prime.
func TestPerThousandZeroCompletesThePercent(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			zero, zadv := cachedOutline(np, "perthousandzero")
			for _, c := range [][2]string{{"percent", "perthousand"}, {"perthousand", "permyriad"}} {
				a, aadv := cachedOutline(np, c[0])
				b, badv := cachedOutline(np, c[1])
				set := append(append(outline.Path{}, a...), shiftInk(zero, aadv, 0)...)
				if aadv+zadv != badv || !samePath(set, b) {
					t.Errorf("%s followed by perthousandzero is not %s", c[0], c[1])
				}
			}
			_, a1 := cachedOutline(np, "minute")
			_, a2 := cachedOutline(np, "second")
			_, a3 := cachedOutline(np, "uni2034")
			if a3-a2 != a2-a1 || a2 <= a1 {
				t.Errorf("the primes have the advances %v, %v and %v", a1, a2, a3)
			}
		})
	}
}

// TestVisibleSpace asks, in every parameter set, that the visible space has
// the advance of the figures with its ink in the middle; that its bar lies
// below the baseline and above the descender line and is one run of ink
// across the sign; that its uprights rise above the baseline and end below
// half the x-height; and that they are as thick as the bar within a unit and
// the white between them at least four times as wide.
func TestVisibleSpace(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, adv := cachedOutline(np, "visiblespace")
			_, zero := cachedOutline(np, "zero")
			x0, y0, x1, y1 := path.BBox()
			if adv != zero || math.Abs(x0-(adv-x1)) > 1 || x0 < 0 {
				t.Errorf("advance %v (the figures %v), side spaces %v and %v", adv, zero, x0, adv-x1)
			}
			if y0 >= 0 || y0 <= -p.Descender || y1 <= 0 || y1 >= p.XHeight/2 {
				t.Errorf("the sign stands from %v to %v", y0, y1)
			}
			v := vruns(path, (x0+x1)/2)
			if len(v) != 1 || v[0][1] > 0 {
				t.Fatalf("the bar in the middle: %v", v)
			}
			thick := v[0][1] - v[0][0]
			if h := runs(path, y0+thick/2); len(h) != 1 || h[0][0] != x0 || h[0][1] != x1 {
				t.Errorf("the bar is not one run of ink across the sign: %v", h)
			}
			h := runs(path, y1-thick/2)
			if len(h) != 2 {
				t.Fatalf("%d runs of ink through the uprights, want 2", len(h))
			}
			for _, r := range h {
				if math.Abs(r[1]-r[0]-thick) > 1 {
					t.Errorf("an upright is %v thick, and the bar %v", r[1]-r[0], thick)
				}
			}
			if white := h[1][0] - h[0][1]; white < 4*thick {
				t.Errorf("the white between the uprights is %v, less than four times the stroke of %v", white, thick)
			}
		})
	}
}
