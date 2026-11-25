// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

// remapped are the ASCII signs which unicode-math sets as other characters
// (the hyphen as the minus, the asterisk as ∗, the colon as ∶).  A ranged font
// which names their code points sets the ASCII glyph in their place; the
// range leaves them out, and the file of ranges points them at the font's
// minus, ∗ and ∶ instead (WriteRanges).
var remapped = map[rune]bool{'*': true, ':': true}

// inRange reports whether a document takes the glyph from Bedern Math: a
// glyph with a code point other than the space and the remapped ASCII signs,
// unless it is a delimiter without its sizes.
func inRange(g Glyph) bool {
	switch {
	case g.Rune <= ' ' || remapped[g.Rune]:
		return false
	case delimiterNames[g.Name] && len(g.Variants) == 0:
		return false
	}
	return true
}

// Ranges returns the code points of the glyphs as runs of consecutive code
// points, in increasing order, of the glyphs which inRange takes.
func Ranges(gs []Glyph) [][2]rune {
	var rs []rune
	for _, g := range gs {
		if inRange(g) {
			rs = append(rs, g.Rune)
		}
	}
	slices.Sort(rs)
	rs = slices.Compact(rs)
	var out [][2]rune
	for _, r := range rs {
		if n := len(out); n > 0 && out[n-1][1]+1 == r {
			out[n-1][1] = r
			continue
		}
		out = append(out, [2]rune{r, r})
	}
	return out
}

// alphabets are the code points which unicode-math takes as alphabets and
// which the option range names by the style and the alphabet, not by code
// points.
var alphabets = []struct {
	cps  []rune
	name string
}{
	{runs('0', '9'), "up/num"}, {runs('A', 'Z'), "up/Latin"}, {runs('a', 'z'), "up/latin"},
	{runs(0x1D434, 0x1D44D), "it/Latin"},
	{append(append(runs(0x1D44E, 0x1D454), 0x210E), runs(0x1D456, 0x1D467)...), "it/latin"},
	{append(runs(0x391, 0x3A1), runs(0x3A3, 0x3A9)...), "up/Greek"},
	{runs(0x1D6FC, 0x1D714), "it/greek"},
	{doubleStruckCapitals(), "bb/Latin"},
}

// doubleStruckCapitals returns the code points of the double-struck capitals A
// to Z, the seven of the Letterlike Symbols among them.
func doubleStruckCapitals() []rune {
	var out []rune
	for _, l := range dsLetters {
		if l.base >= 'A' && l.base <= 'Z' {
			out = append(out, l.r)
		}
	}
	return out
}

// runs returns the code points from lo to hi.
func runs(lo, hi rune) []rune {
	var out []rune
	for r := lo; r <= hi; r++ {
		out = append(out, r)
	}
	return out
}

// RangeList returns the ranges in the syntax of the option range of
// unicode-math: the alphabets by name, and the other code points in
// hexadecimal with a double quote, a hyphen between the ends of a run, and
// commas between the runs.  An alphabet which the font has only in part is
// left out.
func RangeList(rs [][2]rune) string {
	has := func(r rune) bool {
		for _, x := range rs {
			if x[0] <= r && r <= x[1] {
				return true
			}
		}
		return false
	}
	var parts []string
	in := func(r rune) bool { return false }
	for _, a := range alphabets {
		all := true
		for _, r := range a.cps {
			all = all && has(r)
		}
		if all {
			parts = append(parts, a.name)
		}
		prev := in
		in = func(r rune) bool { return slices.Contains(a.cps, r) || prev(r) }
	}
	for _, x := range rs {
		var run []rune
		flush := func() {
			if len(run) == 0 {
				return
			}
			if len(run) == 1 {
				parts = append(parts, fmt.Sprintf(`"%X`, run[0]))
			} else {
				parts = append(parts, fmt.Sprintf(`"%X-"%X`, run[0], run[len(run)-1]))
			}
			run = nil
		}
		for r := x[0]; r <= x[1]; r++ {
			if in(r) {
				flush()
				continue
			}
			run = append(run, r)
		}
		flush()
	}
	return strings.Join(parts, ",")
}

// WriteRanges writes the file of ranges, which LuaLaTeX and XeLaTeX read: it
// finds the family of the font by the mathematical code of a character of the
// font (\Umathcodenum, which both engines have) and uses no \directlua.  It
// defines \BedernMathRanges as the list of the code points of the font and
// \setBedernMath, which loads the font with that range and then points the
// hyphen, the asterisk and the colon at the font's minus, ∗ and ∶, the
// primes at its primes, \mathbb{1} at its double-struck one, \widecheck at
// its caron and \mapstochar at the bar of \xmapsto; the double-struck
// capitals come through the alphabet bb/Latin of the range.  A document loads
// another maths font first for what Bedern Math lacks and Bedern Math last,
// whose constants then hold:
//
//	\usepackage{unicode-math}
//	\input{bedern-math-ranges}
//	\setmathfont{other-math.otf}
//	\setBedernMath
func WriteRanges(w io.Writer, gs []Glyph) error {
	has := map[rune]bool{}
	for _, g := range gs {
		if inRange(g) || remapped[g.Rune] {
			has[g.Rune] = true
		}
	}
	var remap []string
	for _, m := range []struct {
		from  rune
		class int
		to    rune
	}{{'-', 2, 0x2212}, {'*', 2, 0x2217}, {':', 3, 0x2236}} {
		if has[m.to] {
			remap = append(remap, fmt.Sprintf("  \\Umathcode%d=%d \\BedernMathFamily\\space\"%X \\relax\n", m.from, m.class, m.to))
		}
	}
	var primes []string
	for _, m := range []struct {
		name string
		to   rune
	}{{"single", 0x2032}, {"double", 0x2033}, {"triple", 0x2034}, {"quad", 0x2057}} {
		if has[m.to] {
			primes = append(primes, fmt.Sprintf(
				"  \\expandafter\\Umathchardef\\csname __um_prime_%s_mchar\\endcsname=0 \\BedernMathFamily\\space %d\n", m.name, m.to))
		}
	}
	// The capitals of \mathbb come from Bedern Math through the alphabet bb/Latin
	// of the range; unicode-math takes the figure of \mathbb{1} from the font
	// of the alphabet of the figures, which a ranged font does not replace;
	// the double-struck letters which are no capital are added to its
	// switch.
	var bb []string
	for _, l := range dsLetters {
		if has[l.r] && (l.base < 'A' || l.base > 'Z') {
			bb = append(bb, fmt.Sprintf("\\Umathcode%d=0 \\BedernMathFamily\\space\"%X ", l.base, l.r))
		}
	}
	if len(bb) > 0 {
		primes = append(primes, "  \\csname tl_gput_right:cn\\endcsname{g__um_switchto_bb_tl}{"+strings.Join(bb, "")+"}\n")
	}
	// unicode-math leaves \widecheck with the font loaded first when a
	// ranged font has U+030C; it is pointed at the font's caron.
	if has[0x30C] {
		primes = append(primes, "  \\protected\\def\\widecheck{\\Umathaccent 7 \\BedernMathFamily\\space\"030C }\n")
	}
	// \mapstochar, the bar which mathtools sets at the tail of \xmapsto, is
	// the character 0x37 of the symbol family, a figure 7 in a Unicode font;
	// it is pointed at the font's bar.
	if has[MapstocharRune] {
		primes = append(primes, fmt.Sprintf("  \\Umathchardef\\mapstochar=3 \\BedernMathFamily\\space\"%X \\relax\n", MapstocharRune))
	}
	_, err := fmt.Fprintf(w, `%% The code points of %s.
\def\BedernMathRanges{%s}
\def\BedernMathFamily{\number\numexpr(\Umathcodenum"%X-8388608)/16777216\relax}
\protected\def\setBedernMath{%%
  \begingroup\edef\x{\endgroup
    \noexpand\setmathfont{%s}[range={\BedernMathRanges}]}\x
%s%s}
`, File, RangeList(Ranges(gs)), probeRune, File, strings.Join(remap, ""), strings.Join(primes, ""))
	return err
}

// probeRune is a code point of the font which no other font of a document
// sets, by whose mathematical code the file of ranges finds the family of
// the font.
const probeRune = 0x2264
