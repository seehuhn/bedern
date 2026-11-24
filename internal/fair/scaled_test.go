// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package fair

import (
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// scaledPath returns p with every coordinate multiplied by s.
func scaledPath(p outline.Path, s float64) outline.Path {
	m := func(q Point) Point { return Point{X: q.X * s, Y: q.Y * s} }
	out := make(outline.Path, len(p))
	for i, c := range p {
		nc := outline.Contour{Start: m(c.Start)}
		for _, sg := range c.Segs {
			nc.Segs = append(nc.Segs, outline.Segment{IsCurve: sg.IsCurve, P1: m(sg.P1), P2: m(sg.P2), P3: m(sg.P3)})
		}
		out[i] = nc
	}
	return out
}

// arcsWithPiece draws two arcs joined by a straight piece of the given length
// between two corners.
func arcsWithPiece(length float64) outline.Path {
	var p outline.Path
	q := newPen(&p, 0, 0, 90)
	q.arc(200, -40)
	q.turn(-30)
	q.line(length)
	q.turn(-30)
	q.arc(200, -40)
	closeWith(q, &p)
	return p
}

func countKind(a *Analysis, k Kind) int {
	n := 0
	for _, f := range a.Findings {
		if f.Kind == k {
			n++
		}
	}
	return n
}

// TestScaledThresholds checks that thresholds scaled for a copy of a
// construction see the copy as the unscaled thresholds see the construction,
// and that the scaling leaves the thresholds of a glyph at its own size, and
// what they find in it, as they are.
func TestScaledThresholds(t *testing.T) {
	if Default.Scaled(1) != Default {
		t.Errorf("Scaled(1) changes the thresholds")
	}
	const s = 0.5
	// A piece of 20 units is not short; its copy of 10 units is short for the
	// thresholds of the full size and is not for the scaled ones.
	long := arcsWithPiece(20)
	if n := countKind(Analyse("full", long, Default), Short); n != 0 {
		t.Fatalf("the piece of 20 units is short at full size")
	}
	if n := countKind(Analyse("copy", scaledPath(long, s), Default), Short); n == 0 {
		t.Errorf("the copy of the piece is not short for the full-size thresholds: the test shows nothing")
	}
	if n := countKind(Analyse("copy", scaledPath(long, s), Default.Scaled(s)), Short); n != 0 {
		t.Errorf("the copy of the piece is short for the scaled thresholds")
	}
	// A piece of 10 units is short at full size, and its copy is short for
	// the scaled thresholds: the scaling hides nothing which the full size
	// finds.
	short := arcsWithPiece(10)
	if n := countKind(Analyse("full", short, Default), Short); n == 0 {
		t.Fatalf("the piece of 10 units is not short at full size")
	}
	if n := countKind(Analyse("copy", scaledPath(short, s), Default.Scaled(s)), Short); n == 0 {
		t.Errorf("the scaled thresholds miss the short piece of the copy")
	}
	// Angles and ratios stay.
	if sc := Default.Scaled(s); sc.CornerDeg != Default.CornerDeg || sc.KinkDeg != Default.KinkDeg ||
		sc.JumpRatio != Default.JumpRatio || sc.WiggleDepth != Default.WiggleDepth {
		t.Errorf("Scaled changes an angle or a ratio")
	}
}
