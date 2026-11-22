// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// sides returns the distances from the edges of the advance to the ink of the
// glyph, and its advance.
func sides(t *testing.T, name string, s namedParams) (left, right, adv float64) {
	t.Helper()
	path, adv, ok := outlineOf(name, s.params)
	if !ok {
		t.Fatalf("no glyph %s", name)
	}
	x0, _, x1, _ := path.BBox()
	return x0, adv - x1, adv
}

// The full stop, the colon, the exclamation mark, the comma and the semicolon
// share one advance, the full stop is centred in it, and the dot of the
// semicolon stands on the centre of the advance, where the full stop stands.
func TestDotsShareAdvance(t *testing.T) {
	t.Parallel()
	for _, s := range allParams() {
		pl, pr, want := sides(t, "period", s)
		if math.Abs(pl-pr) > 0.51 {
			t.Errorf("%s: the full stop has the side spaces %v and %v", s.name, pl, pr)
		}
		for _, name := range []string{"colon", "exclam", "comma", "semicolon"} {
			if _, _, adv := sides(t, name, s); adv != want {
				t.Errorf("%s: %s has the advance %v, the full stop %v", s.name, name, adv, want)
			}
		}
		path, _, _ := outlineOf("semicolon", s.params)
		found := false
		for _, ct := range path {
			x0, y0, x1, y1 := outline.Path{ct}.BBox()
			if math.Abs((x1-x0)-s.params.DotDiameter) < 0.01 && math.Abs((y1-y0)-s.params.DotDiameter) < 0.01 {
				found = true
				if c := (x0 + x1) / 2; math.Abs(c-want/2) > 0.01 {
					t.Errorf("%s: the dot of the semicolon is centred at %v in the advance %v", s.name, c, want)
				}
			}
		}
		if !found {
			t.Errorf("%s: the semicolon has no dot", s.name)
		}
	}
}

// Each pair of a left and a right sign is a mirror image: the side spaces of
// the right one are those of the left one exchanged, to a unit, and the outer
// side of the parentheses and of the brackets has the larger space.
func TestMirrorPairsHaveMirrorSides(t *testing.T) {
	t.Parallel()
	for _, s := range allParams() {
		for _, pair := range [][2]string{
			{"parenleft", "parenright"}, {"bracketleft", "bracketright"},
			{"braceleft", "braceright"}, {"slash", "backslash"},
			{"quoteleft", "quoteright"}, {"quotedblleft", "quotedblright"},
			{"less", "greater"},
		} {
			l0, r0, a0 := sides(t, pair[0], s)
			l1, r1, a1 := sides(t, pair[1], s)
			if a0 != a1 || math.Abs(l0-r1) > 1.01 || math.Abs(r0-l1) > 1.01 {
				t.Errorf("%s: %s has %v, %v on %v and %s has %v, %v on %v", s.name, pair[0], l0, r0, a0, pair[1], l1, r1, a1)
			}
		}
		for _, name := range []string{"parenleft", "bracketleft"} {
			if l, r, _ := sides(t, name, s); l <= r {
				t.Errorf("%s: %s has %v outside and %v inside, want more outside", s.name, name, l, r)
			}
		}
	}
}

// The signs which stand next to figures have, on the side which faces a figure,
// a side space within 45 units less and 60 units more than that of the zero
// (at the regular parameters, scaled with the advance of the figures).
func TestSignsSpaceLikeFigures(t *testing.T) {
	t.Parallel()
	for _, s := range allParams() {
		zl, zr, _ := sides(t, "zero", s)
		zs := (zl + zr) / 2
		scale := s.params.Figures.Advance / 490
		for _, c := range []struct {
			name  string
			right bool
		}{{"percent", false}, {"degree", false}, {"numbersign", true}, {"dollar", true}, {"sterling", true}, {"section", true}} {
			l, r, _ := sides(t, c.name, s)
			side := l
			if c.right {
				side = r
			}
			if d := side - zs; d < -45*scale || d > 60*scale {
				t.Errorf("%s: the side of %s facing a figure is %v, that of the zero %v", s.name, c.name, side, zs)
			}
		}
	}
}
