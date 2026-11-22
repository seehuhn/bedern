// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
)

// TestCapFitIsEven checks the fit of the capitals in the regular font with the
// measure of the kerning: for every capital the mean white of its pairs with
// the 26 capitals, as the left and as the right neighbour and without kerns,
// exceeds the white of HH by -5 to 80 per cent of it.
func TestCapFitIsEven(t *testing.T) {
	const lo, hi = -0.05, 0.80
	p := design.Regular()
	f := computed(p).Font
	z := Zone{f.Row(0), f.Row(0) + int(p.CapHeight/p.Kern.Band)}
	caps := make([]*Glyph, 0, 26)
	for r := 'A'; r <= 'Z'; r++ {
		g := f.glyphOf(r)
		if g == nil {
			t.Fatalf("no glyph for %c", r)
		}
		caps = append(caps, g)
	}
	H := f.glyphOf('H')
	hh := f.white(H, H, z, 0)
	for _, a := range caps {
		var left, right float64
		for _, b := range caps {
			right += f.white(a, b, z, 0)
			left += f.white(b, a, z, 0)
		}
		n := float64(len(caps))
		for _, s := range []struct {
			side string
			v    float64
		}{{"left neighbour", right/n/hh - 1}, {"right neighbour", left/n/hh - 1}} {
			if s.v < lo || s.v > hi {
				t.Errorf("%c as %s: mean white %.3f relative to HH, want %.2f to %.2f",
					a.Rune, s.side, s.v, lo, hi)
			}
		}
	}
}

// TestCapFitClearance checks that without kerns no pair of capitals and no
// capital before a lowercase letter comes closer than the floor of the
// kerning, in the regular font.
func TestCapFitClearance(t *testing.T) {
	p := design.Regular()
	f := computed(p).Font
	for _, a := range f.Glyphs {
		if a.Kind != Upper || a.Rune > 'Z' {
			continue
		}
		for _, b := range f.Glyphs {
			if b.Rune > 'z' || (b.Kind != Upper && b.Kind != Lower) {
				continue
			}
			if d := f.Distance(a, b, 0); d < p.Kern.Floor {
				t.Errorf("%c%c: closest distance %.1f without a kern, want at least %.0f",
					a.Rune, b.Rune, d, p.Kern.Floor)
			}
		}
	}
}

// TestCapFitAnchors checks that the anchors of the fit of the capitals keep
// their advances (H 747, O 746) and that the ten figures share one advance.
func TestCapFitAnchors(t *testing.T) {
	p := design.Regular()
	f := computed(p).Font
	for _, c := range []struct {
		r   rune
		adv float64
	}{{'H', 747}, {'O', 746}} {
		if g := f.glyphOf(c.r); g == nil || math.Abs(g.Advance-c.adv) > 0.5 {
			t.Errorf("advance of %c changed, want %.0f", c.r, c.adv)
		}
	}
	var adv float64
	for r := '0'; r <= '9'; r++ {
		g := f.glyphOf(r)
		if g == nil {
			t.Fatalf("no glyph for %c", r)
		}
		if r == '0' {
			adv = g.Advance
		} else if g.Advance != adv {
			t.Errorf("figure %c: advance %.1f, want %.1f as the figure 0", r, g.Advance, adv)
		}
	}
}
