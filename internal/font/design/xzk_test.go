// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import (
	"math"
	"testing"
)

// The lean of x reproduces the width of the ink, from the outermost tip of a serif
// on the left to that on the right.
func TestXLeanReproducesTheWidth(t *testing.T) {
	t.Parallel()
	p := Regular()
	lt, ln := XLeanThick(p), XLeanThin(p)
	wt, wn := DiagWidth(p.XThick, lt)/2, DiagWidth(p.Hair, ln)/2
	up, down := p.XHeight-p.XCross, p.XCross
	// The tips relative to the centre of the crossing.
	jt, jn := DiagJutOuter(p, p.XOverThick, lt), DiagJutOuter(p, p.XOverThin, ln)
	tl := -(lt*up + wt + jt)
	tr := ln*up + wn + jn
	bl := p.XThinOffset - (ln*down + wn + jn)
	br := lt*down + wt + jt
	if got := math.Max(tr, br) - math.Min(tl, bl); math.Abs(got-p.XInk) > 1e-6 {
		t.Errorf("the ink is %v wide, want XInk = %v", got, p.XInk)
	}
}

// The lean of z reproduces the width of the ink: from the corner of the diagonal at
// the baseline to the corner at the x-height, which lies over the end of the lower bar.
func TestZLeanReproducesTheWidth(t *testing.T) {
	t.Parallel()
	p := Regular()
	l := ZLean(p)
	if got := p.XHeight*l + DiagWidth(p.ZDiag, l); math.Abs(got-p.ZInk) > 1e-6 {
		t.Errorf("the ink of z is %v wide, want ZInk = %v", got, p.ZInk)
	}
}
