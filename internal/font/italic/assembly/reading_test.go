// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package assembly

import (
	"slices"
	"testing"

	"seehuhn.de/go/bedern/internal/kern"
)

// TestLowercaseKeepsTheReadingFloor checks that no two lowercase letters set
// with their kern come closer than ReadingFloor, and names the pairs ck,
// the pairs of x, and c and e before x.  The
// distance is that of the kerning model, which sets the floor.
func TestLowercaseKeepsTheReadingFloor(t *testing.T) {
	// The kerning gives pairs of one class the median kern of the class and
	// leaves a pair which differs from it by at most design.Kern.PairTolerance.
	tolerance := Kerning().Kern.PairTolerance + 0.5
	fullTable(t)
	tb := kerned()
	f := tb.Font
	byRune := map[rune]*kern.Glyph{}
	for _, g := range f.Glyphs {
		if g.Rune != 0 {
			byRune[g.Rune] = g
		}
	}
	for a := 'a'; a <= 'z'; a++ {
		for b := 'a'; b <= 'z'; b++ {
			k := tb.Kern(byRune[a].Index, byRune[b].Index)
			if d := f.Distance(byRune[a], byRune[b], float64(k)); d < ReadingFloor-tolerance {
				t.Errorf("%c%c with kern %d: closest distance %.1f, want at least %v", a, b, k, d, ReadingFloor)
			}
		}
	}
	for _, p := range []string{"ck", "ox", "xo", "ex", "cx", "xe", "xa", "xc", "ce", "co"} {
		a, b := byRune[rune(p[0])], byRune[rune(p[1])]
		if d := f.Distance(a, b, float64(tb.Kern(a.Index, b.Index))); d < ReadingFloor-tolerance {
			t.Errorf("%s: distance %.1f, want at least %v", p, d, ReadingFloor)
		}
	}
}

// TestFTakesNoNegativeKernBeforeClosingMarks checks that f is not drawn
// under the full stop: no negative kern before a closing mark.
func TestFTakesNoNegativeKernBeforeClosingMarks(t *testing.T) {
	fullTable(t)
	tb := kerned()
	var f *kern.Glyph
	for _, g := range tb.Font.Glyphs {
		if g.Rune == 'f' {
			f = g
		}
	}
	for _, g := range tb.Font.Glyphs {
		if g.Kind == kern.Punct && slices.Contains(closingMarks, g.Rune) {
			if k := tb.Kern(f.Index, g.Index); k < 0 {
				t.Errorf("f before %q has the kern %d", g.Rune, k)
			}
		}
	}
}
