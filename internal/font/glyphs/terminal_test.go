// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// TestLowerTerminal checks the terminal of e on a curve of the o: the pieces
// join, the chain has an outer curve, a straight tip and an upper edge, the tip
// is at right angles to the stroke, and the upper edge is horizontal at the
// lowest point of the inner edge.
func TestLowerTerminal(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	f := newEFrame(p)
	outer := f.outerD.cubic(3, true) // bottom to the right extreme
	low := outline.Point{X: f.cx, Y: f.innerBottomY}
	yTip := f.barLow - p.EAperture
	ch := lowerTerminal(outer, yTip, p.ETerminalTip, low, eShape)
	if err := ch.connected(); err != nil {
		t.Fatal(err)
	}
	if len(ch) != 3 || !ch[1].line {
		t.Fatalf("want an outer curve, a straight tip and an upper edge, got %d pieces", len(ch))
	}
	tip := ch[1]
	// The tip is at right angles to the outer edge there.
	dir := ch[0].tangent(1)
	near(t, "tip at right angles", dir.X*(tip.P3.X-tip.P0.X)+dir.Y*(tip.P3.Y-tip.P0.Y), 0)
	// The upper edge leaves the tip along the stroke, backwards, and is
	// horizontal at the lowest point.
	e := ch.endTangent()
	near(t, "upper edge horizontal at the end", e.Y, 0)
}
