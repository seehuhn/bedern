// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// testArch is an arch unlike the arch of n: the stems are far apart and the
// peak is nearly in the middle.
var testArch = archSpec{
	X1: 300, X2: 560, X3: 640,
	XPeak: 440, Peak: 480, Thick: 34,
	Notch: 410, Join: 300, ThinJoin: 26,
	Shoulder: 372, ShoulderInner: 356,
}

// archLetter joins two plain stems by an arch, in the order of the
// composition convention, and returns the stem edges as well.
func archLetter(p design.Params, a archSpec, shape archShape) outline.Path {
	left := stemSpec{Left: a.X1 - 80, Width: 80, Top: 450}
	right := stemSpec{Left: a.X2, Width: a.X3 - a.X2}
	var path outline.Path
	path.MoveTo(left.start())
	stemUp(&path, p, left, a.Join)
	archUnder(shape, a).appendTo(&path)
	stemDown(&path, p, right)
	stemUp(&path, p, right, a.Shoulder)
	archOuter(shape, a).appendTo(&path)
	stemCap(&path, p, left)
	stemDown(&path, p, left)
	return path
}

func TestArchBetweenStems(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	a := testArch
	path := archLetter(p, a, nArchShape)
	if len(path) != 1 || path[0].SignedArea() <= 0 {
		t.Fatalf("want one counter-clockwise contour, got %d contours", len(path))
	}

	// The join: below it the counter reaches the stem, above it the
	// underside has left.
	below := path.CrossingsX(a.Join - 3)
	if len(below) != 4 || math.Abs(below[1]-a.X1) > 1 {
		t.Errorf("3 units below the join the counter does not reach the stem: %v", below)
	}
	above := path.CrossingsX(a.Join + 15)
	if len(above) != 4 || above[1] < a.X1+3 {
		t.Errorf("15 units above the join the underside has not left the stem: %v", above)
	}
}

// A second shape gives a different arch from the same archSpec.
func TestArchShapeIsUsed(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	a := testArch
	shape := nArchShape
	shape.UnderPeakShift += 20
	base := archLetter(p, a, nArchShape)
	other := archLetter(p, a, shape)
	c0 := base.CrossingsY(a.XPeak)
	c1 := other.CrossingsY(a.XPeak)
	if math.Abs(c0[len(c0)-2]-c1[len(c1)-2]) < 0.5 {
		t.Error("the underside does not depend on the shape")
	}
}

// TestArchChains checks that the two edges of the arch are connected chains and
// that a mirrored arch edge runs from the image of its end.
func TestArchChains(t *testing.T) {
	t.Parallel()
	a := testArch
	under, outer := archUnder(nArchShape, a), archOuter(nArchShape, a)
	for name, c := range map[string]chain{"underside": under, "outer edge": outer} {
		if err := c.connected(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	want := func(what string, got outline.Point, x, y float64) {
		t.Helper()
		near(t, what+" x", got.X, x)
		near(t, what+" y", got.Y, y)
	}
	m := under.mirrorX(a.X1)
	want("mirrored underside start", m.start(), 2*a.X1-a.X2, a.ShoulderInner)
	want("mirrored underside end", m.end(), a.X1, a.Join)
}
