// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// mirrorImage reports whether b is the mirror image of a in a vertical line:
// the same heights and the same ink width, with the side spaces exchanged.
func mirrorImage(a outline.Path, aadv float64, b outline.Path, badv float64) bool {
	ax0, ay0, ax1, ay1 := a.BBox()
	bx0, by0, bx1, by1 := b.BBox()
	return aadv == badv && ay0 == by0 && ay1 == by1 &&
		math.Abs(ax0-(badv-bx1)) < 1 && math.Abs(bx0-(aadv-ax1)) < 1 && math.Abs((ax1-ax0)-(bx1-bx0)) < 1e-6 &&
		math.Abs(inkArea(a)-inkArea(b)) < 1e-3
}

// regularSet returns Set for the regular parameters, built once.  The glyphs
// are shared and must not be changed.
var regularSet = sync.OnceValue(func() []Glyph { return Set(design.Regular()) })

// TestInklessHeights asks, for the regular parameters, that the three
// compound word marks have neither an outline nor a width, that their heights
// are those of the tops of x, H and l, and that every other glyph has the
// height zero.
func TestInklessHeights(t *testing.T) {
	t.Parallel()
	gs := regularSet()
	top := map[string]float64{}
	for _, g := range gs {
		if g.Name == "x" || g.Name == "H" || g.Name == "l" {
			_, _, _, top[g.Name] = g.Path.BBox()
		}
	}
	want := map[string]float64{"uni200C": top["x"], "cwmcapital": top["H"], "cwmascender": top["l"]}
	seen := 0
	for _, g := range gs {
		h, ok := want[g.Name]
		if ok {
			seen++
			if len(g.Path) != 0 || g.Advance != 0 || g.Height != h || h <= 0 {
				t.Errorf("%s: %d contours, advance %v, height %v, want none, 0 and %v", g.Name, len(g.Path), g.Advance, g.Height, h)
			}
		} else if g.Height != 0 {
			t.Errorf("%s has the height %v", g.Name, g.Height)
		}
	}
	if seen != len(want) {
		t.Errorf("%d compound word marks, want %d", seen, len(want))
	}
}

// TestStraightQuotesOnTheBaseline asks, in every parameter set, that the
// straight quotation marks on the baseline are the straight marks moved down,
// with their advances, and that they stand across the baseline: the top
// between the baseline and half the x-height, the foot below the baseline
// and above the foot of the brackets.
func TestStraightQuotesOnTheBaseline(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			for _, c := range [][2]string{{"quotesingle", "quotesinglbase.ts1"}, {"quotedbl", "quotedblbase.ts1"}} {
				high, hadv := cachedOutline(np, c[0])
				low, ladv := cachedOutline(np, c[1])
				_, _, _, hy1 := high.BBox()
				_, y0, _, y1 := low.BBox()
				if hadv != ladv || !samePath(shiftInk(high, 0, y1-hy1), low) {
					t.Errorf("%s is not %s moved down", c[1], c[0])
				}
				if y1 <= 0 || y1 >= p.XHeight/2 || y0 >= 0 || y0 <= p.BracketBottom() {
					t.Errorf("%s stands from %v to %v", c[1], y0, y1)
				}
			}
		})
	}
}

// TestPartDashes asks, in every parameter set, that the dashes of two thirds
// and of three quarters of an em stand at the height of the em dash with its
// thickness and its side spaces, and that their advances lie in this order
// between those of the en dash and the em dash, each a whole number.
func TestPartDashes(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			em, emAdv := cachedOutline(np, "emdash")
			mx0, my0, mx1, my1 := em.BBox()
			_, last := cachedOutline(np, "endash")
			for _, name := range []string{"twelveudash", "threequartersemdash"} {
				path, adv := cachedOutline(np, name)
				x0, y0, x1, y1 := path.BBox()
				if y0 != my0 || y1 != my1 || x0 != mx0 || adv-x1 != emAdv-mx1 {
					t.Errorf("%s: ink from %v to %v with the side spaces %v and %v; the em dash from %v to %v with %v and %v",
						name, y0, y1, x0, adv-x1, my0, my1, mx0, emAdv-mx1)
				}
				if adv <= last || adv >= emAdv || adv != math.Round(adv) {
					t.Errorf("%s has the advance %v, between %v and %v", name, adv, last, emAdv)
				}
				last = adv
			}
		})
	}
}

// TestDoubleHyphens asks, in every parameter set, that the double hyphen is
// two bars of the length and the thickness of the hyphen, as far above its
// middle as below, with at least a bar of white between them, between the
// baseline and the x-height, in the advance of the hyphen; and that the
// double hyphen as a hyphenation character has the same outline in a smaller
// advance, at least a third of that of the hyphen, beyond which its ink
// reaches.
func TestDoubleHyphens(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			hy, hadv := cachedOutline(np, "hyphen")
			hx0, hy0, hx1, hy1 := hy.BBox()
			path, adv := cachedOutline(np, "uni2E40")
			x0, y0, x1, y1 := path.BBox()
			if adv != hadv || x0 != hx0 || x1 != hx1 {
				t.Errorf("advance %v and ink from %v to %v; the hyphen %v and from %v to %v", adv, x0, x1, hadv, hx0, hx1)
			}
			v := vruns(path, (x0+x1)/2)
			if len(v) != 2 {
				t.Fatalf("%d bars, want 2", len(v))
			}
			thick := hy1 - hy0
			for _, r := range v {
				if math.Abs(r[1]-r[0]-thick) > 1e-6 {
					t.Errorf("a bar is %v thick, and the hyphen %v", r[1]-r[0], thick)
				}
			}
			if math.Abs((y0+y1)-(hy0+hy1)) > 1e-6 {
				t.Errorf("the bars stand from %v to %v, not about the middle of the hyphen, from %v to %v", y0, y1, hy0, hy1)
			}
			if white := v[1][0] - v[0][1]; white < thick {
				t.Errorf("the white between the bars is %v, less than a bar of %v", white, thick)
			}
			if y0 <= 0 || y1 >= p.XHeight {
				t.Errorf("the bars stand from %v to %v, outside the x-height", y0, y1)
			}
			char, cadv := cachedOutline(np, "hyphendbl.alt")
			if !samePath(char, path) {
				t.Errorf("the hyphenation character is not the double hyphen")
			}
			if cadv >= x1 || cadv < hadv/3 || cadv != math.Round(cadv) {
				t.Errorf("the hyphenation character has the advance %v, the hyphen %v, and its ink ends at %v", cadv, hadv, x1)
			}
		})
	}
}

// TestTextBrackets asks, in every parameter set, that the angle, the double
// and the quill brackets reach from the foot to the top of the square
// brackets, and that each right bracket is the mirror image of the left one;
// that the point of the angle bracket lies at the middle of its height and
// its strokes, across, are between the hairline and the stem of the bracket;
// that the double bracket has an outer stem of the vertical bar and an inner
// stem as thick as the arms and no thicker than the outer one, with at least
// 0.7 of the outer stem of white between them, closed into one counter by
// the arms, which reach at least the outer stem beyond the inner one; and
// that the quill bracket has the stem and the arms of the bracket and a third
// arm of their length and thickness about the middle of its height.
func TestTextBrackets(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			br, _ := cachedOutline(np, "bracketleft")
			bx0, by0, bx1, by1 := br.BBox()
			mid := (by0 + by1) / 2
			for _, c := range [][2]string{{"angleleft", "angleright"}, {"uni27E6", "uni27E7"}, {"uni2045", "uni2046"}} {
				l, ladv := cachedOutline(np, c[0])
				r, radv := cachedOutline(np, c[1])
				if _, y0, _, y1 := l.BBox(); y0 != by0 || y1 != by1 {
					t.Errorf("%s stands from %v to %v, and the bracket from %v to %v", c[0], y0, y1, by0, by1)
				}
				if !mirrorImage(l, ladv, r, radv) {
					t.Errorf("%s is not the mirror image of %s", c[1], c[0])
				}
			}

			angle, _ := cachedOutline(np, "angleleft")
			ax0, _, ax1, _ := angle.BBox()
			if h := runs(angle, mid); len(h) != 1 || math.Abs(h[0][0]-ax0) > 1 {
				t.Errorf("the point of the angle bracket: the ink at the middle is %v, and the sign starts at %v", h, ax0)
			}
			lo, hi := runs(angle, mid+0.25*(by1-mid)), runs(angle, mid+0.75*(by1-mid))
			if len(lo) != 1 || len(hi) != 1 {
				t.Fatalf("the upper stroke of the angle bracket: %v and %v", lo, hi)
			}
			lean := math.Atan2((hi[0][0]+hi[0][1]-lo[0][0]-lo[0][1])/2, 0.5*(by1-mid))
			across := (lo[0][1] - lo[0][0]) * math.Cos(lean)
			if across < p.Hair || across > p.Punct.BracketStem {
				t.Errorf("the strokes of the angle bracket are %.1f across, outside %v to %v", across, p.Hair, p.Punct.BracketStem)
			}
			if lean < 10*math.Pi/180 || lean > 35*math.Pi/180 || ax1-ax0 < 2*across {
				t.Errorf("the strokes of the angle bracket lean %.1f degrees in an ink %v wide", lean*180/math.Pi, ax1-ax0)
			}

			dbl, _ := cachedOutline(np, "uni27E6")
			bar, _ := cachedOutline(np, "bar")
			vx0, _, vx1, _ := bar.BBox()
			stems := runs(dbl, mid)
			if len(dbl) != 2 || len(stems) != 2 {
				t.Fatalf("the double bracket has %d contours and %d stems, want 2 and 2", len(dbl), len(stems))
			}
			if math.Abs(stems[0][1]-stems[0][0]-(vx1-vx0)) > 1e-6 {
				t.Errorf("the outer stem of the double bracket is %v wide, and the vertical bar %v", stems[0][1]-stems[0][0], vx1-vx0)
			}
			darms := vruns(dbl, stems[1][1]+1)
			if inner := stems[1][1] - stems[1][0]; len(darms) != 2 || math.Abs(inner-(darms[0][1]-darms[0][0])) > 1e-6 || inner > stems[0][1]-stems[0][0] {
				t.Errorf("the inner stem of the double bracket is %v wide, the outer stem %v, and the arms are %v", inner, stems[0][1]-stems[0][0], darms)
			}
			if white := stems[1][0] - stems[0][1]; white < 0.7*(vx1-vx0) {
				t.Errorf("the white between the stems of the double bracket is %v", white)
			}
			if _, _, x1, _ := dbl.BBox(); x1-stems[1][1] < vx1-vx0 {
				t.Errorf("the arms of the double bracket reach %v beyond the inner stem, less than a stem", x1-stems[1][1])
			}

			quill, _ := cachedOutline(np, "uni2045")
			if qx0, _, qx1, _ := quill.BBox(); qx0 != bx0 || qx1 != bx1 {
				t.Errorf("the quill bracket has its ink from %v to %v, and the bracket from %v to %v", qx0, qx1, bx0, bx1)
			}
			arms := vruns(quill, bx1-1)
			if len(arms) != 3 {
				t.Fatalf("the quill bracket has %d arms, want 3", len(arms))
			}
			if math.Abs(arms[1][0]+arms[1][1]-2*mid) > 1e-6 || math.Abs((arms[1][1]-arms[1][0])-(arms[0][1]-arms[0][0])) > 1e-6 {
				t.Errorf("the middle arm of the quill bracket stands from %v to %v, the lower arm from %v to %v, and the middle is %v",
					arms[1][0], arms[1][1], arms[0][0], arms[0][1], mid)
			}
			if h := runs(quill, mid); len(h) != 1 || h[0][0] != bx0 || h[0][1] != bx1 {
				t.Errorf("the middle arm of the quill bracket does not join the stem: %v", h)
			}
		})
	}
}

// TestDoubleBar asks, in every parameter set, that the double vertical bar
// is two bars of the width and the height of the vertical bar with its side
// spaces and with at least one bar of white between them.
func TestDoubleBar(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			bar, badv := cachedOutline(np, "bar")
			bx0, by0, bx1, by1 := bar.BBox()
			path, adv := cachedOutline(np, "uni2016")
			x0, y0, x1, y1 := path.BBox()
			if y0 != by0 || y1 != by1 || x0 != bx0 || adv-x1 != badv-bx1 {
				t.Errorf("from %v to %v with the side spaces %v and %v; the bar from %v to %v with %v and %v", y0, y1, x0, adv-x1, by0, by1, bx0, badv-bx1)
			}
			h := runs(path, (y0+y1)/2)
			if len(path) != 2 || len(h) != 2 {
				t.Fatalf("%d contours and %d bars, want 2 and 2", len(path), len(h))
			}
			w := bx1 - bx0
			if math.Abs(h[0][1]-h[0][0]-w) > 1e-6 || math.Abs(h[1][1]-h[1][0]-w) > 1e-6 || h[1][0]-h[0][1] < w {
				t.Errorf("the bars are %v, and the vertical bar is %v wide", h, w)
			}
		})
	}
}

// TestCircles asks, in every parameter set, that the large circle is the
// ring of the circled letters, contour for contour, in their advance; that
// the open bullet has the outer edge, the place and the advance of the
// bullet; and that its ring is of one thickness all round within a unit and
// a half, at least a hairline, and leaves a counter of at least half the
// diameter.
func TestCircles(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			c, cadv := cachedOutline(np, "copyright")
			large, ladv := cachedOutline(np, "uni25EF")
			if len(large) != 2 || len(c) < 2 || ladv != cadv {
				t.Errorf("the large circle is not a ring in the advance of the copyright sign")
			} else if p.Sym.LargeRing == 0 {
				if !samePath(large, c[:2]) {
					t.Errorf("the large circle is not the ring of the copyright sign")
				}
			} else {
				// A ring of its own thickness with the outer edge of the ring of
				// the copyright sign.
				lx0, ly0, lx1, ly1 := large.BBox()
				cx0, cy0, cx1, cy1 := c[:2].BBox()
				if lx0 != cx0 || ly0 != cy0 || lx1 != cx1 || ly1 != cy1 {
					t.Errorf("the large circle does not have the outer edge of the ring of the copyright sign")
				}
				xs := large.CrossingsX((ly0 + ly1) / 2)
				if len(xs) != 4 || math.Abs((xs[1]-xs[0])-p.Sym.LargeRing) > 1.5 {
					t.Errorf("the ring of the large circle is not %.0f thick: %v", p.Sym.LargeRing, xs)
				}
			}

			b, badv := cachedOutline(np, "bullet")
			o, oadv := cachedOutline(np, "openbullet")
			if len(o) != 2 || len(b) != 1 || oadv != badv || !samePath(o[:1], b) {
				t.Fatalf("the outer edge of the open bullet is not the bullet")
			}
			x0, _, x1, _ := o.BBox()
			outer := o[0].Flatten()
			lo, hi := math.Inf(1), 0.0
			for _, q := range o[1].Flatten() {
				d := math.Inf(1)
				for i, a := range outer {
					d = math.Min(d, segmentDistance(q, q, a, outer[(i+1)%len(outer)]))
				}
				lo, hi = math.Min(lo, d), math.Max(hi, d)
			}
			if hi-lo > 1.5 || lo < p.ThinLower || x1-x0-2*hi < (x1-x0)/2 {
				t.Errorf("the ring of the open bullet is %.1f to %.1f thick in a diameter of %v", lo, hi, x1-x0)
			}
		})
	}
}

// TestLowTildeAndDoubleGrave asks, in every parameter set, that the low tilde
// is the tilde moved down, in its advance, wholly below the baseline and
// above the descender line; and that the double grave is the mirror image of
// the double acute.
func TestLowTildeAndDoubleGrave(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			tilde, tadv := cachedOutline(np, "tilde")
			low, ladv := cachedOutline(np, "uni02F7")
			_, _, _, ty1 := tilde.BBox()
			_, y0, _, y1 := low.BBox()
			if tadv != ladv || !samePath(shiftInk(tilde, 0, y1-ty1), low) {
				t.Errorf("the low tilde is not the tilde moved down")
			}
			if y1 >= 0 || y0 <= -p.Descender {
				t.Errorf("the low tilde stands from %v to %v", y0, y1)
			}
			acute, aadv := cachedOutline(np, "hungarumlaut")
			grave, gadv := cachedOutline(np, "uni02F5")
			if len(grave) != len(acute) || !mirrorImage(acute, aadv, grave, gadv) {
				t.Errorf("the double grave is not the mirror image of the double acute")
			}
		})
	}
}
