// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"bytes"
	"strings"
	"testing"
)

// TestKnownCodePoints requires the code points of a table of unicode-math to
// be read in increasing order, once each.
func TestKnownCodePoints(t *testing.T) {
	table := `%% a comment
\UnicodeMathSymbol{"00021}{\mathexclam               }{\mathclose}{exclamation mark}%
\UnicodeMathSymbol{"1D400}{\mbfA                     }{\mathalpha}{bold capital a}%
\UnicodeMathSymbol{"021A6}{\mapsto                   }{\mathrel}{rightwards arrow from bar}%
\UnicodeMathSymbol{"021A6}{\mapsto                   }{\mathrel}{rightwards arrow from bar}%
`
	got, err := KnownCodePoints(strings.NewReader(table))
	if err != nil {
		t.Fatal(err)
	}
	want := []rune{0x21, 0x21A6, 0x1D400}
	if len(got) != len(want) {
		t.Fatalf("got %U, want %U", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %U, want %U", got, want)
		}
	}
}

// TestLackingRanges requires the list of the code points which the font lacks
// to name every symbol of unicode-math which the font has no glyph for, to
// name no code point which the font takes, and to leave the letters of the
// alphabets to the names of the alphabets.
func TestLackingRanges(t *testing.T) {
	gs := []Glyph{{Name: "space", Rune: ' '}, {Name: "colon", Rune: ':'}, {Name: "plus", Rune: '+'},
		{Name: "leq", Rune: 0x2264}, {Name: "mapsto", Rune: 0x21A6}, {Name: "bbR", Rune: 0x211D}}
	known := []rune{0x21, 0x2B, 0x3A, 0x3B1, 0x211C, 0x211D, 0x2135, 0x21A6, 0x2200, 0x2201, 0x2264, 0x2265, 0x2266, 0x1D49C}
	runs := LackingRanges(known, gs)
	in := func(r rune) bool {
		for _, x := range runs {
			if x[0] <= r && r <= x[1] {
				return true
			}
		}
		return false
	}
	for _, r := range []rune{0x21, 0x2200, 0x2201, 0x2265, 0x2266, 0x2135} {
		if !in(r) {
			t.Errorf("%U lacks a glyph and is not in the list", r)
		}
	}
	for _, r := range []rune{':', '+', 0x2264, 0x21A6, 0x211D} {
		if in(r) {
			t.Errorf("%U has a glyph and is in the list", r)
		}
	}
	for _, r := range []rune{0x211C, 0x3B1, 0x1D49C} {
		if in(r) {
			t.Errorf("%U is a letter of an alphabet and is in the list", r)
		}
	}
	// the run which starts at 2200 does not reach over the arrow ↦ at 21A6
	for _, x := range runs {
		if x[0] < 0x21A6 && x[1] > 0x21A6 {
			t.Errorf("run %U-%U reaches over the glyph ↦", x[0], x[1])
		}
	}
	var buf bytes.Buffer
	if err := WriteLacking(&buf, gs, known); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`\def\BedernMathLackedAlphabets{cal,`, `"2200-"2201`, `\setBedernMathSecond`} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("the file lacks %q:\n%s", want, buf.String())
		}
	}
}

// TestLackingRangesOfTheFont requires, for the font itself, that the list
// names no glyph of the font and no letter of an alphabet, and that the glyphs
// which the first tier has are not named by an alphabet in the list.
func TestLackingRangesOfTheFont(t *testing.T) {
	gs := derivedSigns(derivedSets(0)[0]) // the signs, to which the build of the whole font adds
	if !testing.Short() {
		var err error
		if gs, err = regularGlyphs(); err != nil {
			t.Fatal(err)
		}
	}
	// every symbol of the Mathematical Operators and Arrows blocks, as if
	// unicode-math knew them all
	var known []rune
	for r := rune(0x2190); r <= 0x22FF; r++ {
		known = append(known, r)
	}
	runs := LackingRanges(known, gs)
	for _, g := range gs {
		if !inRange(g) && !remapped[g.Rune] {
			continue
		}
		for _, x := range runs {
			if x[0] <= g.Rune && g.Rune <= x[1] {
				t.Errorf("%s (%U) is a glyph of the font and in the run %U-%U", g.Name, g.Rune, x[0], x[1])
			}
		}
	}
	for _, name := range lackedAlphabets {
		for _, a := range alphabets {
			if a.name == name {
				t.Errorf("the font has the alphabet %s whole, and the list names it as lacking", name)
			}
		}
	}
}

// TestMapstocharIsInTheFile requires the file of ranges to point
// \mapstochar at the bar of the font, and the bar to be a glyph with ink as
// high as the head of the arrows.
func TestMapstocharIsInTheFile(t *testing.T) {
	gs := derivedSigns(derivedSets(0)[0])
	var bar, arrow Glyph
	for _, g := range gs {
		switch g.Rune {
		case MapstocharRune:
			bar = g
		case 0x21A6:
			arrow = g
		}
	}
	if len(bar.Path) == 0 {
		t.Fatal("no bar at the code point of \\mapstochar")
	}
	bx0, by0, bx1, by1 := bar.Path.Round().BBox()
	_, ay0, _, ay1 := arrow.Path.Round().BBox()
	if by0 > ay0+1 || by1 < ay1-1 {
		t.Errorf("the bar stands from %.0f to %.0f, the arrow from %.0f to %.0f", by0, by1, ay0, ay1)
	}
	if bx1-bx0 > 80 {
		t.Errorf("the bar is %.0f wide, more than a stroke", bx1-bx0)
	}
	var buf bytes.Buffer
	if err := WriteRanges(&buf, gs); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `\Umathchardef\mapstochar=3 \BedernMathFamily\space"E500`) {
		t.Errorf("the file of ranges does not point \\mapstochar at the bar:\n%s", buf.String())
	}
}
