// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
)

// fitDepth is the cap of the white in one band for the measure of the fit, as
// a fraction of the x-height.
const fitDepth = 0.55

// fitBandLow and fitBandHigh bound the band within which the mean relative white of every lowercase
// letter, as a left and as a right neighbour of all 26 lowercase letters, lies
// around the white of nn, without kerning.
const fitBandLow, fitBandHigh = -0.20, 0.32

// TestLowercaseFitIsEven checks the fit of the lowercase letters without
// kerning.  The white of a pair is the mean over the bands from the baseline
// to the x-height of the horizontal distance between the two letters, capped
// at fitDepth of the x-height, and it is taken relative to the white of nn.
// For every letter the mean of this relative white minus one over its row (the
// letter on the left of each of the 26 letters) and over its column (on the
// right) lies within fitBandLow to fitBandHigh, and the anchors n and o keep
// their advances, so that the colour and the length of a line stay as they are.
func TestLowercaseFitIsEven(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	f := Analyse(p, glyphstest.Set(p))
	by := map[rune]*Glyph{}
	for _, g := range f.Glyphs {
		if g.Kind == Lower && g.Name == string(g.Rune) {
			by[g.Rune] = g
		}
	}
	lo := f.Row(0)
	hi := lo + int(p.XHeight/p.Kern.Band)
	depth := fitDepth * p.XHeight
	white := func(a, b *Glyph) float64 {
		var s float64
		for i := lo; i < hi; i++ {
			ra, lb := a.Right[i], b.Left[i]
			if math.IsNaN(ra) || math.IsNaN(lb) {
				s += depth
				continue
			}
			s += math.Max(0, math.Min(depth, ra+lb))
		}
		return s / float64(hi-lo)
	}
	nn := white(by['n'], by['n'])
	for x := 'a'; x <= 'z'; x++ {
		var row, col float64
		for y := 'a'; y <= 'z'; y++ {
			row += white(by[x], by[y])/nn - 1
			col += white(by[y], by[x])/nn - 1
		}
		row /= 26
		col /= 26
		if row < fitBandLow || row > fitBandHigh {
			t.Errorf("%c as the left letter: mean white %+.3f of nn, want %+.2f to %+.2f", x, row, fitBandLow, fitBandHigh)
		}
		if col < fitBandLow || col > fitBandHigh {
			t.Errorf("%c as the right letter: mean white %+.3f of nn, want %+.2f to %+.2f", x, col, fitBandLow, fitBandHigh)
		}
		if testing.Verbose() {
			t.Logf("%c row %+.3f col %+.3f", x, row, col)
		}
	}
	if a := design.NAdvance(p); a != 540 {
		t.Errorf("advance of n %v, want 540", a)
	}
	if a := design.OAdvance(p); a != 486 {
		t.Errorf("advance of o %v, want 486", a)
	}
}
