// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package assembly

import (
	"slices"
	"strings"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/kern"
)

// ReadingFloor is the least distance between the outlines of two lowercase
// letters set with their kern, higher than the KernFloor of the capitals and
// the signs.  45 is 0.10 of the x-height.
const ReadingFloor = 45

// RBallKern is the kern of r before the letters without an entry: the ball of
// r hangs at 0.75 of the x-height, where only the entries of the following
// letter come near it (TestRBallClearsTheNext keeps 30 clear without a kern
// there).
const RBallKern = -30

// roundAfterR are the letters before which r takes RBallKern.
const roundAfterR = "acdegoqs"

// closingMarks are the marks before which f takes no negative kern.  The
// floor still applies.
var closingMarks = []rune{'.', ',', ':', ';', '!', '?', '’', '”'}

// KernTable computes the kerns of the glyphs of the italic: the model of the
// roman with the floor of the italic, the ReadingFloor for the pairs of
// lowercase letters and no negative kern for f before the closing marks.
//
// The pairs which the reading floor and the closing marks set by hand are
// passed to the kerning as extra exceptions.
func KernTable(gs []glyphs.Glyph) *kern.Table {
	first := kern.Compute(Kerning(), gs)
	f := first.Font
	var exc []design.KernException
	var f1, r1 *kern.Glyph
	var marks []*kern.Glyph
	for _, g := range f.Glyphs {
		switch {
		case g.Rune == 'f' && g.Like < 0:
			f1 = g
		case g.Rune == 'r' && g.Like < 0 && g.First < 0:
			r1 = g
		case g.Kind == kern.Punct && slices.Contains(closingMarks, g.Rune):
			marks = append(marks, g)
		}
	}
	if f1 != nil {
		for _, m := range marks {
			exc = append(exc, design.KernException{Left: f1.Name, Right: m.Name, Value: 0,
				Reason: "f takes no negative kern before a closing mark"})
		}
	}
	for _, a := range f.Glyphs {
		if a.Kind != kern.Lower || a.Like >= 0 || a.First >= 0 {
			continue
		}
		for _, b := range f.Glyphs {
			if b.Kind != kern.Lower || b.Like >= 0 || b.First >= 0 {
				continue
			}
			k := first.Kern(a.Index, b.Index)
			tight := a == r1 && strings.ContainsRune(roundAfterR, b.Rune)
			if tight {
				k = min(k, RBallKern)
			}
			if f.Distance(a, b, float64(k)) >= ReadingFloor {
				if tight {
					exc = append(exc, design.KernException{Left: a.Name, Right: b.Name, Value: float64(k),
						Reason: "r before a letter without an entry stands closer, below its ball"})
				}
				continue
			}
			// The least kern which clears the floor.
			lo, hi := k, 400
			for lo < hi {
				mid := lo + (hi-lo)/2
				if f.Distance(a, b, float64(mid)) >= ReadingFloor {
					hi = mid
				} else {
					lo = mid + 1
				}
			}
			exc = append(exc, design.KernException{Left: a.Name, Right: b.Name, Value: float64(lo),
				Reason: "the closest approach of two lowercase letters is at least the reading floor"})
		}
	}
	return kern.ComputeWith(Kerning(), gs, exc)
}
