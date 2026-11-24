// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// ink returns the area of the ink of the outline.
func ink(p outline.Path) float64 {
	var a float64
	for _, c := range p {
		a += c.SignedArea()
	}
	return math.Abs(a)
}

// Colour returns the colour of the string set in the italic and in the
// roman: the ink over the sum of the advances times the x-height.  Every letter of s must be
// drawn in the italic.
func Colour(p Params, s string) (italic, roman float64, err error) {
	rp := design.Regular()
	var ii, ia, ri, ra float64
	cache := map[rune]*Glyph{}
	for _, r := range s {
		g, ok := cache[r]
		if !ok {
			g, err = Build(p, string(r))
			if err != nil {
				return 0, 0, err
			}
			cache[r] = g
		}
		ii += ink(g.Path)
		ia += g.Advance
		path, adv, ok := glyphs.Outline(string(r), rp)
		if !ok {
			return 0, 0, fmt.Errorf("italic: no roman %q", r)
		}
		ri += ink(path)
		ra += adv
	}
	return ii / (ia * p.XHeight), ri / (ra * rp.XHeight), nil
}
