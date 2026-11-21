// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"
)

// .notdef is a frame with a diagonal cross: one outer contour and four triangular
// counters, from the baseline to the height Notdef.Height, on the advance Notdef.Advance.
func TestNotdef(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		p := np.params
		path, adv, ok := outlineOf(".notdef", p)
		if !ok {
			t.Fatalf("%s: no glyph .notdef", np.name)
		}
		if len(path) != 5 {
			t.Fatalf("%s: .notdef has %d contours, want 5", np.name, len(path))
		}
		for i, ct := range path[1:] {
			if len(ct.Segs) != 3 {
				t.Errorf("%s: counter %d has %d sides, want a triangle", np.name, i, len(ct.Segs))
			}
		}
		x0, y0, x1, y1 := path.BBox()
		if y0 != 0 || y1 != p.Notdef.Height || adv != p.Notdef.Advance {
			t.Errorf("%s: .notdef is %vx%v from %v with advance %v, want the height %v and the advance %v", np.name, x1-x0, y1-y0, y0, adv, p.Notdef.Height, p.Notdef.Advance)
		}
		if l, r := x0, adv-x1; math.Abs(l-r) > 1 {
			t.Errorf("%s: .notdef has side spaces %v and %v", np.name, l, r)
		}
	}
}
