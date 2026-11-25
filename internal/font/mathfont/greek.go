// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/font/italic"
	"seehuhn.de/go/bedern/internal/font/italic/assembly"
)

// The italic Greek lowercase: the letters of bedern/internal/font/italic at the code points of the
// mathematical italic, with mathSide on each side over their own side
// spaces, the overhangs of the maths italic, and the italics correction and
// the top accent attachment by the rules of the maths italic (italic.go).
// The Greek has no script forms.

// greekVariants are the code points of the mathematical italic of the
// variant letters, by their code points in the Greek block.
var greekVariants = map[rune]rune{
	'ϵ': 0x1D716, 'ϑ': 0x1D717, 'ϰ': 0x1D718, 'ϕ': 0x1D719, 'ϱ': 0x1D71A, 'ϖ': 0x1D71B,
}

// mathItalicGreek returns the code point of the mathematical italic of a
// letter of the Greek block, or 0.
func mathItalicGreek(r rune) rune {
	if 'α' <= r && r <= 'ω' {
		return 0x1D6FC + r - 'α'
	}
	return greekVariants[r]
}

// greekLower returns the italic Greek lowercase, built with the parameters
// of the italic.
func greekLower() ([]Glyph, error) {
	gs, err := italic.GreekGlyphs(italic.Default())
	if err != nil {
		return nil, err
	}
	out := make([]Glyph, 0, len(gs))
	for _, g := range gs {
		r := mathItalicGreek(g.Rune)
		if r == 0 {
			return nil, fmt.Errorf("mathfont: no mathematical italic for %s", g.Name)
		}
		path, adv := withinOverhang(shifted(g.Path, mathSide, 0), g.Advance+2*mathSide)
		ic := assembly.ItalicCorrection(glyphs.Glyph{Name: g.Name, Advance: adv, Path: path}, icPad)
		out = append(out, Glyph{
			Name: fmt.Sprintf("u%X", r), Rune: r, Advance: adv, Path: path, Class: Ord,
			Italic: ic, Attach: accentAttachment(g.Name, path), Attached: true,
		})
	}
	return out, nil
}

// rightmostAbove returns the rightmost ink of the outline at the height y or
// above, on the flattened curves.
func rightmostAbove(path outline.Path, y float64) float64 {
	x := math.Inf(-1)
	for _, c := range path {
		for _, q := range c.Flatten() {
			if q.Y >= y {
				x = math.Max(x, q.X)
			}
		}
	}
	return x
}
