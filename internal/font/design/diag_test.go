// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import (
	"math"
	"testing"
)

// TestVLeanReproducesTheWidth checks that the outer edges of v stand VInk less
// the two jutes apart at the x-height and DiagVertexFlat apart at the flat, and
// that the outer edges of w and the flat of its apex have the widths which
// follow from WInk and WApexFlat.
func TestVLeanReproducesTheWidth(t *testing.T) {
	t.Parallel()
	p := Regular()
	l := VLean(p)
	top := 2*l*DiagHeight(p) + p.DiagVertexFlat
	jut := DiagJutOuter(p, p.DiagOverThick, l) + DiagJutOuter(p, p.DiagOverThin, l)
	if math.Abs(top-(p.VInk-jut)) > 1e-9 {
		t.Errorf("the outer edges are %v apart at the x-height, want %v", top, p.VInk-jut)
	}
	// The edges of the two outer diagonals of w stand WInk less the two jutes
	// apart, and the flat of the apex has its width.
	lw := WLean(p)
	jw := DiagJutOuter(p, p.DiagOverThick, lw) + DiagJutOuter(p, p.DiagOverThin, lw)
	ink := WPeriod(p) + p.DiagVertexFlat + 2*lw*DiagHeight(p)
	if math.Abs(ink-(p.WInk-jw)) > 1e-6 {
		t.Errorf("the outer edges of w are %v apart at the x-height, want %v", ink, p.WInk-jw)
	}
	apex := WPeriod(p) - p.DiagVertexFlat - 2*lw*(p.WApexHeight+p.DiagOvershoot) + DiagWidth(p.Hair, lw) + DiagWidth(WThick(p), lw)
	if math.Abs(apex-p.WApexFlat) > 1e-6 {
		t.Errorf("the flat of the apex is %v wide, want %v", apex, p.WApexFlat)
	}
}

// TestCrotchHeight checks that the inner edges of v meet at the height VCrotch.
func TestCrotchHeight(t *testing.T) {
	t.Parallel()
	p := Regular()
	l := VLean(p)
	y := VCrotch(p)
	// Right-hand edge of the thick diagonal and left-hand edge of the thin one, from the
	// outer edges which pass through the ends of the flat (x measured from the middle).
	s := y + p.DiagOvershoot
	left := -p.DiagVertexFlat/2 - l*s + DiagWidth(p.Stem, l)
	right := p.DiagVertexFlat/2 + l*s - DiagWidth(p.Hair, l)
	if math.Abs(left-right) > 1e-9 {
		t.Errorf("the inner edges are %v apart at the crotch, want 0", right-left)
	}
}
