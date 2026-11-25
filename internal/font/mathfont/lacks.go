// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

// RangeLackFile is the name of the file which lists the code points that
// Bedern Math lacks.
const RangeLackFile = "bedern-math-lacks.tex"

// KnownCodePoints returns the code points of the symbols which unicode-math
// knows, read from the lines \UnicodeMathSymbol{"XXXXX}{...} of its file
// unicode-math-table.tex, in increasing order.
func KnownCodePoints(table io.Reader) ([]rune, error) {
	data, err := io.ReadAll(table)
	if err != nil {
		return nil, err
	}
	const key = `\UnicodeMathSymbol{"`
	var out []rune
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, key) {
			continue
		}
		hex, _, ok := strings.Cut(line[len(key):], "}")
		if !ok {
			return nil, fmt.Errorf("unicode-math table: line %q has no closing brace", line)
		}
		var r rune
		if _, err := fmt.Sscanf(hex, "%X", &r); err != nil {
			return nil, fmt.Errorf("unicode-math table: line %q: %w", line, err)
		}
		out = append(out, r)
	}
	slices.Sort(out)
	return slices.Compact(out), nil
}

// lackedAlphabets are the alphabets which the font lacks whole, named as the
// option range of unicode-math names them.  A command such as \mathcal
// switches to the font of its alphabet, which the code points of the alphabet
// in a range do not give it; the range names the alphabets as well as
// the code points.  An alphabet which the font has in part (the italic
// Greek capitals, the double-struck lowercase and figures, the upright Greek
// lowercase) is named by style and alphabet.
//
// The double-struck lowercase and figures cannot be named: with bb/latin or
// bb/num in the range \mathbb sets no double-struck letter at all.
var lackedAlphabets = []string{
	"cal", "scr", "bfcal", "bfscr", "frak", "bffrak", "tt", "sfup", "sfit",
	"bfup", "bfit", "bfsfup", "bfsfit", "bbit", "up/greek", "it/Greek",
}

// inAlphabet holds the code points of the letters of the alphabets of
// unicode-math: the mathematical alphanumerics, the Greek letters and the
// letters of the Letterlike Symbols (the script, fraktur and double-struck
// letters which stand outside the block of the alphanumerics, and the
// double-struck pi, gamma and sum, which unicode-math counts as the
// miscellaneous letters of the double-struck style).  A range which contains
// one of them makes unicode-math switch the whole style of the letter to the
// font of the range; the code points of the alphabets are left to
// the names of the alphabets.
var inAlphabet = func() map[rune]bool {
	m := map[rune]bool{}
	add := func(lo, hi rune) {
		for r := lo; r <= hi; r++ {
			m[r] = true
		}
	}
	add(0x1D400, 0x1D7FF)
	add(0x391, 0x3FF)
	for _, x := range [][2]rune{{0x2102, 0x2102}, {0x210A, 0x2113}, {0x2115, 0x211D},
		{0x2124, 0x2124}, {0x2128, 0x2128}, {0x212C, 0x2131}, {0x2133, 0x2134},
		{0x213C, 0x2140}, {0x2145, 0x2149}} {
		add(x[0], x[1])
	}
	// the symbols among them
	for _, r := range []rune{0x210F, 0x2113, 0x2118, 0x211E} {
		delete(m, r)
	}
	return m
}()

// LackingRanges returns the code points among known which Bedern Math does
// not take as the first font: those without a glyph, and the delimiters
// without their sizes.  The ASCII signs which unicode-math sets as other
// characters count as taken.  The result is runs of code points in increasing
// lies between them; a run may contain code points which unicode-math
// does not know.
func LackingRanges(known []rune, gs []Glyph) [][2]rune {
	has := map[rune]bool{}
	var had []rune
	for _, g := range gs {
		if (inRange(g) || remapped[g.Rune]) && !has[g.Rune] {
			has[g.Rune] = true
			had = append(had, g.Rune)
		}
	}
	// a run does not reach over the glyphs of the font or over the letters
	// of the alphabets
	for r := range inAlphabet {
		had = append(had, r)
	}
	slices.Sort(had)
	var out [][2]rune
	for _, r := range known {
		if has[r] || inAlphabet[r] {
			continue
		}
		if n := len(out); n > 0 {
			// the first code point taken by the font after the end of the run
			i, _ := slices.BinarySearch(had, out[n-1][1])
			if i == len(had) || had[i] > r {
				out[n-1][1] = r
				continue
			}
		}
		out = append(out, [2]rune{r, r})
	}
	return out
}

// WriteLacking writes the file of the code points which Bedern Math lacks,
// for a second maths font under either engine.  It defines
// \BedernMathLacks, the list in the syntax of the option range of
// unicode-math (the alphabets which the font lacks by name, the other code
// points as runs), and \setBedernMathSecond, which loads the font named by its
// argument with that range.  A document loads Bedern Math first:
//
//	\usepackage{unicode-math}
//	\setmathfont{BedernMath-Regular.otf}
//	\input{bedern-math-lacks}
//	\setBedernMathSecond{second-math.otf}
func WriteLacking(w io.Writer, gs []Glyph, known []rune) error {
	var parts []string
	for _, x := range LackingRanges(known, gs) {
		if x[0] == x[1] {
			parts = append(parts, fmt.Sprintf(`"%X`, x[0]))
		} else {
			parts = append(parts, fmt.Sprintf(`"%X-"%X`, x[0], x[1]))
		}
	}
	_, err := fmt.Fprintf(w, `%% The code points of unicode-math which %s lacks.
\def\BedernMathLackedAlphabets{%s}
\def\BedernMathLackedCodes{%s}
\def\BedernMathLacks{\BedernMathLackedAlphabets,\BedernMathLackedCodes}
\protected\def\setBedernMathSecond#1{%%
  \begingroup\edef\x{\endgroup
    \noexpand\setmathfont{#1}[range={\BedernMathLacks}]}\x}
`, File, strings.Join(lackedAlphabets, ","), strings.Join(parts, ","))
	return err
}
