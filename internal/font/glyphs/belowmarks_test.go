// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import "testing"

// TestMarksBelowClearTheInk sets the combining dot below and the combining
// comma below on the bottom anchor of every letter of the alphabet, as the
// feature mark does, and asks that the ink of the mark lies below all the ink
// of the letter by at least 30 units, descenders included.
func TestMarksBelowClearTheInk(t *testing.T) {
	t.Parallel()
	const clear = 30
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			var letters []string
			for c := 'a'; c <= 'z'; c++ {
				letters = append(letters, string(c), string(c-'a'+'A'))
			}
			deep := 0
			for _, mark := range []struct {
				name string
				kind markKind
			}{{"dotbelowcomb", markDotBelow}, {"uni0326", markCommaBelow}} {
				ma := markAnchor(np.params, mark.kind)
				_, _, _, my1 := cachedRounded(np, mark.name).Path.BBox()
				for _, name := range letters {
					raw, _ := cachedOutline(np, name)
					var at *Anchor
					anchors := baseAnchors(np.params, name, raw)
					for i, a := range anchors {
						if a.Class == "bottom" {
							at = &anchors[i]
						}
					}
					if at == nil {
						t.Fatalf("%s has no bottom anchor", name)
					}
					_, by0, _, _ := raw.BBox()
					top := my1 + at.Y - ma.Y
					if top > by0-clear {
						t.Errorf("%s on %s: the top of the mark is at %v, the ink of the letter reaches down to %v", mark.name, name, top, by0)
					}
					if at.Y < 0 {
						deep++
					}
				}
			}
			if deep < 8 {
				t.Errorf("only %d of the marks lie below a descender", deep)
			}
		})
	}
}
