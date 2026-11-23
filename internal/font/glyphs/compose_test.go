// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"slices"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// composedNames returns the names of the composed letters.
func composedNames() []string {
	var out []string
	for _, c := range composedList {
		out = append(out, c.name)
	}
	return out
}

// TestComposedKeepTheBase: a composed letter is its rounded base, unchanged,
// with the mark, in the parameter sets of composedRoutine.  Where the mark is joined to the base, every node of the base
// which lies more than joinReach from the box of the mark is a node of the
// letter.
func TestComposedKeepTheBase(t *testing.T) {
	t.Parallel()
	const joinReach = 40
	for _, np := range allParams() {
		if !composedRoutine(np.name) {
			continue
		}
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, name := range composedNames() {
				c := composites[name]
				g := cachedRounded(np, name)
				b := cachedRounded(np, c.base)
				if g.Advance != b.Advance {
					t.Errorf("%s: the advance is %v, that of %s %v", name, g.Advance, c.base, b.Advance)
				}
				if !c.joined() {
					if len(g.Path) <= len(b.Path) || !pathsEqual(g.Path[:len(b.Path)], b.Path) {
						t.Errorf("%s: the letter does not begin with the contours of %s", name, c.base)
					}
					continue
				}
				dx, dy, f := c.place(np.params, cachedOutline1(np, c.base))
				mx0, my0, mx1, my1 := shiftInk(markOutline(np.params, c.mark, f), dx, dy).BBox()
				nodes := map[outline.Point]bool{}
				for _, ct := range g.Path {
					nodes[ct.Start] = true
					for _, s := range ct.Segs {
						nodes[s.P3] = true
					}
				}
				for _, ct := range b.Path {
					for _, s := range ct.Segs {
						q := s.P3
						if q.X > mx0-joinReach && q.X < mx1+joinReach && q.Y > my0-joinReach && q.Y < my1+joinReach {
							continue
						}
						if !nodes[q] {
							t.Errorf("%s: the node %v of %s is missing", name, q, c.base)
						}
					}
				}
			}
		})
	}
}

// cachedOutline1 returns the unrounded outline of the shared build.
func cachedOutline1(np namedParams, glyph string) outline.Path {
	p, _ := cachedOutline(np, glyph)
	return p
}

func pathsEqual(a, b outline.Path) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Start != b[i].Start || !slices.Equal(a[i].Segs, b[i].Segs) {
			return false
		}
	}
	return true
}

// markPart returns the box of the mark of a composed letter which is not
// joined: its last contours.
func markPart(np namedParams, name string) (x0, y0, x1, y1 float64) {
	c := composites[name]
	g := cachedRounded(np, name)
	n := len(markOutline(np.params, c.mark, formLower))
	return g.Path[len(g.Path)-n:].BBox()
}

// TestMarksStandOverTheMiddle: a mark over a symmetric letter (o, n, u, v, x,
// O, H, U, V, X and the others below) stands with the middle of its box within
// markCentreTol of the middle of the ink of the letter, the acute and the
// grave moved to the right and to the left by Acc.AcuteShift; over i, the
// dotless i and the dotless j, the middle of the mark stands where the middle
// of the dot of i stands.
func TestMarksStandOverTheMiddle(t *testing.T) {
	t.Parallel()
	const markCentreTol = 3
	symmetric := []string{"o", "n", "u", "v", "w", "x", "y", "z", "e", "c", "s",
		"A", "C", "E", "G", "H", "N", "O", "R", "S", "U", "V", "W", "Y", "Z", "D", "T", "I"}
	for _, np := range allParams() {
		if !composedRoutine(np.name) {
			continue
		}
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			ip, _ := cachedOutline(np, "i")
			iStem0, iStem1 := stemAt(ip, p.XHeight/2)
			dotMid := (iStem0+iStem1)/2 + design.DotOffset(p)
			for _, c := range composedList {
				if c.mark.below() || c.mark == markApostrophe || c.mark == markCommaAbove {
					continue
				}
				shift := 0.0
				switch c.mark {
				case markAcute:
					shift = p.Acc.AcuteShift
				case markGrave:
					shift = -p.Acc.AcuteShift
				}
				x0, _, x1, _ := markPart(np, c.name)
				mid := (x0+x1)/2 - shift
				var want float64
				switch {
				case c.base == "dotlessi":
					want = dotMid
				case slices.Contains(symmetric, c.base):
					// The middle of the ink above the baseline, without the
					// tail of y.
					b := cachedRounded(np, c.base)
					bx0, bx1 := inkSpan(b.Path, 0, p.CapHeight)
					want = (bx0 + bx1) / 2
				default:
					continue
				}
				if math.Abs(mid-want) > markCentreTol {
					t.Errorf("%s: the middle of the mark is at %.1f, want %.1f within %v", c.name, mid, want, markCentreTol)
				}
			}
		})
	}
}

// TestMarksClearTheBase: in every parameter set (in the routine tier the sets
// of composedRoutine) the lowest ink of a mark above a letter stands at least
// minGapLower above the highest ink of a lowercase letter and minGapCap above
// that of a capital or a letter with an ascender, and a mark below the
// baseline which is not joined stands at least minGapBelow below the lowest
// ink of its letter.
func TestMarksClearTheBase(t *testing.T) {
	t.Parallel()
	const minGapLower, minGapCap, minGapBelow = 36, 24, 36
	for _, np := range goldenParams() {
		if !composedRoutine(np.name) {
			continue
		}
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, c := range composedList {
				if c.joined() || c.mark == markApostrophe {
					continue
				}
				b := cachedRounded(np, c.base)
				_, by0, _, by1 := b.Path.BBox()
				_, my0, _, my1 := markPart(np, c.name)
				if c.mark.below() {
					if gap := by0 - my1; gap < minGapBelow {
						t.Errorf("%s: the mark is %.0f below the letter, want at least %v", c.name, gap, minGapBelow)
					}
					continue
				}
				want := float64(minGapLower)
				if isCapital(c.base) || slices.Contains(ascenderBases, c.base) {
					want = minGapCap
				}
				if gap := my0 - by1; gap < want {
					t.Errorf("%s: the mark is %.0f above the letter, want at least %v", c.name, gap, want)
				}
			}
		})
	}
}
