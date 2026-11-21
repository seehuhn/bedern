// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests in this file build parts from stemSpec values made up for the
// test.  The serif thickness and the bracket come from the regular
// parameters.

// testStem is a stem unlike any in the font: wide, with a foot serif on the
// left which is shorter than the one on the right, and a head serif.
var testStem = stemSpec{
	Left: 210, Width: 74, Bottom: -100, Top: 590,
	FootLeft: 52, FootRight: 66,
	HeadJut: 58, HeadDrop: 9,
}

func stemPath(p design.Params, s stemSpec) outline.Path {
	var path outline.Path
	stemAlone(&path, p, s)
	return path
}

// TestStemAlone checks that a stem with a head serif and unequal foot serifs
// is one counter-clockwise contour with the shorter serif on the left.
func TestStemAlone(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	s := testStem
	path := stemPath(p, s)
	if len(path) != 1 || path[0].SignedArea() <= 0 {
		t.Fatalf("want one counter-clockwise contour, got %d contours", len(path))
	}
	xs := path.CrossingsX((s.Bottom + s.Top) / 2)
	if len(xs) != 2 {
		t.Fatalf("crossings at mid height: %v", xs)
	}

	fs := path.CrossingsX(s.Bottom + footScanY)
	if len(fs) != 2 {
		t.Fatalf("crossings at the foot: %v", fs)
	}
	// The left serif is the shorter one.
	if !(s.Left-fs[0] < fs[1]-(s.Left+s.Width)) {
		t.Error("the left foot serif is not the shorter one")
	}
	// Above the bracket only the stem remains.
	ss := path.CrossingsX(s.Bottom + p.BracketHeight + 1)
	if len(ss) != 2 {
		t.Fatalf("crossings above the bracket: %v", ss)
	}
}
