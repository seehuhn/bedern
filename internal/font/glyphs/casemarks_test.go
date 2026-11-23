// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"slices"
	"testing"
)

// capitalCombiningName returns the name of the capital form of the combining
// mark of the kind k.
func capitalCombiningName(k markKind) string {
	for _, s := range combiningList {
		if s.mark == k {
			return s.name + caseSuffix
		}
	}
	return ""
}

// TestCapitalMarksStandAsOnTheComposedLetters: the capital form of a
// combining mark, set with its anchor on the top anchor of a capital or of a
// letter with an ascender, has the ink of the mark of the composed letter
// within a unit, which the rounding of the outlines leaves.
func TestCapitalMarksStandAsOnTheComposedLetters(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			n := 0
			for _, c := range composedList {
				if c.mark.below() || c.mark == markApostrophe || c.form() != formCap {
					continue
				}
				n++
				name := capitalCombiningName(c.mark)
				raw, _ := cachedOutline(np, c.base)
				var top Anchor
				for _, a := range baseAnchors(p, c.base, raw) {
					if a.Class == "top" {
						top = a
					}
				}
				m := capitalMarkAnchor(p)
				x0, y0, x1, y1 := cachedRounded(np, name).Path.BBox()
				dx, dy := top.X-m.X, top.Y-m.Y
				cx0, cy0, cx1, cy1 := markPart(np, c.name)
				for _, d := range []float64{x0 + dx - cx0, y0 + dy - cy0, x1 + dx - cx1, y1 + dy - cy1} {
					if math.Abs(d) > 1 {
						t.Errorf("%s on %s has the ink %v %v %v %v, and %s has its mark at %v %v %v %v",
							name, c.base, x0+dx, y0+dy, x1+dx, y1+dy, c.name, cx0, cy0, cx1, cy1)
						break
					}
				}
			}
			if n < 40 {
				t.Errorf("only %d composed letters carry a capital form", n)
			}
		})
	}
}

// TestCapitalAccentsStandAboveTheCapitals: a spacing capital accent has its
// lowest ink at least minGap above the top of the round capitals, the ink of
// the combining capital form, and the same space on both sides of its ink.
func TestCapitalAccentsStandAboveTheCapitals(t *testing.T) {
	t.Parallel()
	const minGap = 24
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			for _, s := range caseSpacingList {
				g := cachedRounded(np, s.name+caseSuffix)
				x0, y0, x1, y1 := g.Path.BBox()
				if gap := y0 - (p.CapHeight + p.OvershootUpper); gap < minGap {
					t.Errorf("%s stands %.0f above the top of O, want at least %v", g.Name, gap, minGap)
				}
				if l, r := x0, g.Advance-x1; math.Abs(l-r) > 1 || l < 1 {
					t.Errorf("%s has the side spaces %.0f and %.0f", g.Name, l, r)
				}
				mx0, my0, mx1, my1 := cachedRounded(np, capitalCombiningName(s.mark)).Path.BBox()
				if math.Abs((x1-x0)-(mx1-mx0)) > 0.01 || math.Abs((y1-y0)-(my1-my0)) > 0.01 || math.Abs(y0-my0) > 0.01 {
					t.Errorf("%s is %.0f wide and %.0f high from %.0f, and the combining form %.0f and %.0f from %.0f",
						g.Name, x1-x0, y1-y0, y0, mx1-mx0, my1-my0, my0)
				}
			}
		})
	}
}

// TestCapitalAccentsAreFlatter: the capital form of a mark is at most as
// high as the lowercase form, and lower for the acute, the grave, the
// circumflex, the caron, the breve and the double acute.
func TestCapitalAccentsAreFlatter(t *testing.T) {
	t.Parallel()
	lower := []markKind{markAcute, markGrave, markCircumflex, markCaron, markBreve, markDblAcute}
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			n := 0
			for _, s := range combiningList {
				if s.mark.below() {
					continue
				}
				n++
				_, y0, _, y1 := cachedRounded(np, s.name).Path.BBox()
				_, cy0, _, cy1 := cachedRounded(np, s.name+caseSuffix).Path.BBox()
				h, ch := y1-y0, cy1-cy0
				if ch > h+0.01 || slices.Contains(lower, s.mark) && ch > h-5 {
					t.Errorf("%s is %.0f high and %s %.0f", s.name+caseSuffix, ch, s.name, h)
				}
			}
			if n != len(caseSpacingList) {
				t.Errorf("%d combining marks above and %d spacing capital accents", n, len(caseSpacingList))
			}
		})
	}
}
