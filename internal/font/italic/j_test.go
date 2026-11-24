// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "testing"

// maxOverhang is how far f and j may
// reach beyond their set width.
const maxOverhang = 140.0

// TestJTail asks of j the tail of every tail (checkTail), and that the tail
// reaches no further than maxOverhang left of the origin in the italic, so
// that the floor of the kerning can keep it from the letter before.
func TestJTail(t *testing.T) {
	p := Default()
	checkTail(t, p, "j", p.JSideLeft)
	if x0, _, _, _ := bbox(glyph(t, "j").Path); -x0 > maxOverhang {
		t.Errorf("the tail of j reaches %.0f left of the origin, want at most %v", -x0, maxOverhang)
	}
}
