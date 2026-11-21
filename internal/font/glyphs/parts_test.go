// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// wantCurrent checks that the current point of the path is (x, y), where a
// part ends and the next one continues.
func wantCurrent(t *testing.T, path *outline.Path, what string, x, y float64) {
	t.Helper()
	if got := lastPoint(path); math.Abs(got.X-x) > 1e-9 || math.Abs(got.Y-y) > 1e-9 {
		t.Errorf("after %s the current point is %v, want (%v, %v)", what, got, x, y)
	}
}

// Every part ends where its comment says, since the next part starts there.
func TestPartsEndWhereTheyShould(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	a := testArch
	left := stemSpec{
		Left: a.X1 - 80, Width: 80, Top: 450, Bottom: -10,
		FootLeft: 50, FootRight: 40, HeadJut: 60, HeadDrop: 10,
	}
	right := stemSpec{Left: a.X2, Width: a.X3 - a.X2, Bottom: -10, FootLeft: 30, FootRight: 50}

	var path outline.Path
	path.MoveTo(left.start())
	wantCurrent(t, &path, "start", left.Left-left.FootLeft, left.Bottom)
	stemUp(&path, p, left, a.Join)
	wantCurrent(t, &path, "stemUp", left.right(), a.Join)
	archUnder(nArchShape, a).appendTo(&path)
	wantCurrent(t, &path, "archUnder", a.X2, a.ShoulderInner)
	stemDown(&path, p, right)
	wantCurrent(t, &path, "stemDown of the right stem", right.Left-right.FootLeft, right.Bottom)
	stemUp(&path, p, right, a.Shoulder)
	wantCurrent(t, &path, "stemUp of the right stem", right.right(), a.Shoulder)
	archOuter(nArchShape, a).appendTo(&path)
	wantCurrent(t, &path, "archOuter", a.X1, a.Notch)
	// The cap continues up the right-hand edge from the notch, which is below
	// the top of the stem.
	stemCap(&path, p, left)
	cur := lastPoint(&path)
	if cur.X != left.Left || cur.Y >= left.Top-left.HeadDrop {
		t.Errorf("after stemCap with a head serif the current point is %v, want on the left-hand edge below the head serif", cur)
	}
	stemDown(&path, p, left)
	wantCurrent(t, &path, "stemDown of the left stem", left.Left-left.FootLeft, left.Bottom)
	// The contour ends where it started.
	if path[0].Start != lastPoint(&path) {
		t.Errorf("the contour starts at %v and ends at %v", path[0].Start, lastPoint(&path))
	}

	// Without a head serif the cap ends at the top left corner.
	left.HeadJut = 0
	var plain outline.Path
	plain.MoveTo(left.start())
	stemUp(&plain, p, left, left.Top)
	wantCurrent(t, &plain, "stemUp to the top", left.right(), left.Top)
	stemCap(&plain, p, left)
	wantCurrent(t, &plain, "stemCap without a head serif", left.Left, left.Top)
}

func lastOfContour(c outline.Contour) outline.Point {
	if len(c.Segs) == 0 {
		return c.Start
	}
	return c.Segs[len(c.Segs)-1].P3
}
