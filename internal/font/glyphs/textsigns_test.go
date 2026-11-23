// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// runAt returns the run of ink which holds the abscissa x, and whether there
// is one.
func runAt(rs [][2]float64, x float64) ([2]float64, bool) {
	for _, r := range rs {
		if r[0] <= x && x <= r[1] {
			return r, true
		}
	}
	return [2]float64{}, false
}

// TestBlank asks, in every parameter set, that the blank symbol has the
// advance and the box of b; that a stroke crosses its stem between the
// x-height and the ascender, rising to the right, 0.8 to 1.3 thin strokes
// across; and that the stroke keeps at least two thin strokes of white from
// the head serif on the left and from the bowl on the right.
func TestBlank(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			b, badv := cachedOutline(np, "b")
			path, adv := cachedOutline(np, "uni2422")
			bx0, by0, bx1, by1 := b.BBox()
			x0, y0, x1, y1 := path.BBox()
			if adv != badv || math.Abs(x0-bx0) > 1e-6 || math.Abs(x1-bx1) > 1e-6 || y0 != by0 || y1 != by1 {
				t.Errorf("advance %v and box %v %v %v %v; b has %v and %v %v %v %v", adv, x0, y0, x1, y1, badv, bx0, by0, bx1, by1)
			}
			left := design.BStemLeft(p)
			right := left + p.AscenderStem
			xl, xr := left-p.SerifJut/2, right+0.75*p.Stem
			vl, vr := vruns(path, xl), vruns(path, xr)
			if len(vl) != 2 || len(vr) < 2 {
				t.Fatalf("the scans beside the stem meet %d and %d runs of ink", len(vl), len(vr))
			}
			sl, sr := vl[0], vr[len(vr)-1]
			if sl[0] <= p.XHeight || sr[1] >= p.Ascender {
				t.Errorf("the stroke stands from %v to %v, outside the x-height %v and the ascender %v", sl[0], sr[1], p.XHeight, p.Ascender)
			}
			rise := (sr[0] + sr[1] - sl[0] - sl[1]) / 2
			across := (sl[1] - sl[0]) * math.Cos(math.Atan2(rise, xr-xl))
			if rise <= 0 || across < 0.8*p.ThinLower || across > 1.3*p.ThinLower {
				t.Errorf("the stroke rises %.1f and is %.1f across; the thin stroke is %v", rise, across, p.ThinLower)
			}
			if w := vl[1][0] - sl[1]; w < 2*p.ThinLower {
				t.Errorf("the white between the stroke and the head serif is %v", w)
			}
			if w := sr[0] - vr[len(vr)-2][1]; w < 2*p.ThinLower {
				t.Errorf("the white between the stroke and the bowl is %v", w)
			}
		})
	}
}

// TestCommercialMinus asks, in every parameter set, that the commercial minus
// has the advance of the figures with its ink in the middle; that it is a
// stroke from the baseline to the height of the figures and two dots of the
// size of the full stop; that the upper dot stands at the left of the ink in
// its upper third and the lower dot at the right in its lower third, neither
// beyond the ends of the stroke; and that each dot keeps half its diameter
// of white from the stroke.
func TestCommercialMinus(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, adv := cachedOutline(np, "uni2052")
			_, zero := cachedOutline(np, "zero")
			stop, _ := cachedOutline(np, "period")
			sx0, sy0, sx1, sy1 := stop.BBox()
			x0, y0, x1, y1 := path.BBox()
			if adv != zero || math.Abs(x0-(adv-x1)) > 1 {
				t.Errorf("advance %v (the figures %v), side spaces %v and %v", adv, zero, x0, adv-x1)
			}
			if len(path) != 3 {
				t.Fatalf("%d contours, want 3", len(path))
			}
			stroke, upper, lower := path[:1], path[1:2], path[2:3]
			if _, a, _, b := stroke.BBox(); a != 0 || b != p.Figures.Height || y0 != a || y1 != b {
				t.Errorf("the stroke stands from %v to %v and the sign from %v to %v; the figures are %v high", a, b, y0, y1, p.Figures.Height)
			}
			h := y1 - y0
			ux0, uy0, ux1, uy1 := upper.BBox()
			lx0, ly0, lx1, ly1 := lower.BBox()
			for _, d := range [][4]float64{{ux0, uy0, ux1, uy1}, {lx0, ly0, lx1, ly1}} {
				if math.Abs(d[2]-d[0]-(sx1-sx0)) > 1e-6 || math.Abs(d[3]-d[1]-(sy1-sy0)) > 1e-6 {
					t.Errorf("a dot is %v by %v, and the full stop %v by %v", d[2]-d[0], d[3]-d[1], sx1-sx0, sy1-sy0)
				}
			}
			if math.Abs(ux0-x0) > 1e-6 || uy0 < y0+2*h/3 || uy1 > y1 {
				t.Errorf("the upper dot stands from %v at the heights %v to %v", ux0, uy0, uy1)
			}
			if math.Abs(lx1-x1) > 1e-6 || ly1 > y0+h/3 || ly0 < y0 {
				t.Errorf("the lower dot ends at %v at the heights %v to %v", lx1, ly0, ly1)
			}
			want := (sx1 - sx0) / 2
			for i, dot := range []outline.Path{upper, lower} {
				if d := pathsDistance(dot, stroke, want); d < want {
					t.Errorf("dot %d is %.1f from the stroke, want at least %v", i, d, want)
				}
			}
		})
	}
}

// TestReferenceMark asks, in every parameter set, that the mark ※ is
// a cross and four dots; that the cross is as wide as high within a unit,
// with its middle on the maths axis; that the dots are of one size, at most
// that of the full stop, and touch the box of the cross at the middles of
// its four sides; and that each dot keeps at least 40 units from the cross.
func TestReferenceMark(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, adv := cachedOutline(np, "uni203B")
			if len(path) != 5 {
				t.Fatalf("%d contours, want 5", len(path))
			}
			stop, _ := cachedOutline(np, "period")
			sx0, _, sx1, _ := stop.BBox()
			cross := path[:1]
			x0, y0, x1, y1 := cross.BBox()
			cx, cy := (x0+x1)/2, (y0+y1)/2
			if math.Abs((x1-x0)-(y1-y0)) > 1 || math.Abs(cy-p.Punct.MathAxis) > 0.5 || math.Abs(x0-(adv-x1)) > 1 {
				t.Errorf("the cross is %v by %v about the height %v, with the side spaces %v and %v", x1-x0, y1-y0, cy, x0, adv-x1)
			}
			size := 0.0
			for i := 1; i < 5; i++ {
				dot := path[i : i+1]
				dx0, dy0, dx1, dy1 := dot.BBox()
				d := dx1 - dx0
				if i == 1 {
					size = d
				}
				if math.Abs(d-size) > 1e-6 || math.Abs(dy1-dy0-d) > 1e-6 || d > sx1-sx0+1e-6 {
					t.Errorf("dot %d is %v by %v; the first is %v and the full stop %v", i, d, dy1-dy0, size, sx1-sx0)
				}
				mx, my := (dx0+dx1)/2-cx, (dy0+dy1)/2-cy
				reach := math.Max(math.Abs(mx), math.Abs(my)) + d/2
				if math.Min(math.Abs(mx), math.Abs(my)) > 0.5 || math.Abs(reach-(x1-x0)/2) > 1 {
					t.Errorf("dot %d stands at (%v, %v) from the middle and reaches %v; the cross reaches %v", i, mx, my, reach, (x1-x0)/2)
				}
				if dist := pathsDistance(dot, cross, 40); dist < 40 {
					t.Errorf("dot %d is %.1f from the cross, want at least 40", i, dist)
				}
			}
		})
	}
}

// TestPilcrowWithOneStem asks, in every parameter set, that the pilcrow with
// one stem is one contour as high and as deep as the pilcrow, in a smaller
// advance; that a scan a quarter of the way up meets one stem, 0.5 to 0.9
// stems wide, where the pilcrow has two; that the foot is at least 1.8 times
// as wide as the stem and reaches as far to its left as to its right; and
// that the bowl, four fifths of the way up, is one run of ink with the stem,
// at least three times as wide.
func TestPilcrowWithOneStem(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			two, tadv := cachedOutline(np, "paragraph")
			_, ty0, _, ty1 := two.BBox()
			path, adv := cachedOutline(np, "paragraph.alt")
			_, y0, _, y1 := path.BBox()
			if len(path) != 1 || y0 != ty0 || y1 != ty1 || adv >= tadv {
				t.Errorf("%d contours from %v to %v in the advance %v; the pilcrow from %v to %v in %v", len(path), y0, y1, adv, ty0, ty1, tadv)
			}
			h := y1 - y0
			stem := runs(path, y0+h/4)
			if len(stem) != 1 || len(runs(two, y0+h/4)) != 2 {
				t.Fatalf("%d stems, and the pilcrow %d", len(stem), len(runs(two, y0+h/4)))
			}
			w := stem[0][1] - stem[0][0]
			if w < 0.5*p.Stem || w > 0.9*p.Stem {
				t.Errorf("the stem is %v wide, the stem of the lowercase %v", w, p.Stem)
			}
			foot := runs(path, y0+2)
			if len(foot) != 1 || foot[0][1]-foot[0][0] < 1.8*w || math.Abs((stem[0][0]-foot[0][0])-(foot[0][1]-stem[0][1])) > 1 {
				t.Errorf("the foot is %v, and the stem %v", foot, stem)
			}
			bowl := runs(path, y0+0.8*h)
			if len(bowl) != 1 || math.Abs(bowl[0][1]-stem[0][1]) > 1e-6 || bowl[0][1]-bowl[0][0] < 3*w {
				t.Errorf("the bowl is %v, and the stem %v", bowl, stem)
			}
		})
	}
}

// TestInterrobang asks, in every parameter set, that the interrobang has the
// advance and the leftmost and rightmost ink of the question mark and its
// dot, moved sideways;
// that it is an outer contour, one counter and the dot; that its stem is one
// run of ink over the middle of the dot, from at least 0.4 dots above the dot
// to the top of the sign, which stands at least 10 units above the question
// mark; that the stem is alone and narrowest at its foot, where it is at most
// as wide as at its top; that the counter is at least a thin stroke wide; and
// that the inverted interrobang is the interrobang turned through half a turn
// in the same advance, as deep as the inverted question mark within a unit,
// with its dot at least as high as the upper dot of the colon and higher by
// at most what the stem stands above the question mark.
func TestInterrobang(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			q, qadv := cachedOutline(np, "question")
			path, adv := cachedOutline(np, "uni203D")
			if adv != qadv || len(path) != 3 || len(q) != 2 {
				t.Fatalf("advance %v and %d contours; the question mark %v and %d", adv, len(path), qadv, len(q))
			}
			qdx0, _, _, _ := q[1:].BBox()
			ddx0, _, _, _ := path[2:].BBox()
			if !samePath(path[2:], shiftInk(q[1:], ddx0-qdx0, 0)) {
				t.Errorf("the dot is not that of the question mark moved sideways")
			}
			qx0, _, qx1, qy1 := q.BBox()
			x0, y0, x1, y1 := path.BBox()
			if math.Abs(x0-qx0) > 1 || math.Abs(x1-qx1) > 1 || y1 < qy1+10 {
				t.Errorf("ink from %v to %v, top %v; the question mark from %v to %v, top %v", x0, x1, y1, qx0, qx1, qy1)
			}
			dx0, _, dx1, dy1 := path[2:].BBox()
			cx := (dx0 + dx1) / 2
			v := vruns(path, cx)
			if len(v) != 2 || v[1][0] < dy1+0.4*(dx1-dx0) || v[1][1] != y1 {
				t.Fatalf("the ink over the middle of the dot: %v; the dot ends at %v and the sign at %v", v, dy1, y1)
			}
			foot := runs(path, v[1][0]+2)
			top, ok := runAt(runs(path, y1-2), cx)
			if len(foot) != 1 || !ok || foot[0][1]-foot[0][0] > top[1]-top[0] || math.Abs(top[0]+top[1]-2*cx) > 1 {
				t.Errorf("the stem is %v at its foot and %v at its top, over the dot at %v", foot, top, cx)
			}
			counter := path[:1]
			if path[1].SignedArea() < 0 {
				counter = path[1:2]
			}
			if cx0, _, cx1, _ := counter.BBox(); counter[0].SignedArea() >= 0 || cx1-cx0 < p.ThinLower || cx1-cx0 > (x1-x0)/2 {
				t.Errorf("the counter is %v wide and has the area %v", cx1-cx0, counter[0].SignedArea())
			}

			colon, _ := cachedOutline(np, "colon")
			_, _, _, ctop := colon.BBox()
			qdown, _ := cachedOutline(np, "questiondown")
			_, qfoot, _, _ := qdown.BBox()
			down, dadv := cachedOutline(np, "uni2E18")
			_, dfoot, _, dtop := down.BBox()
			if dadv != adv || !samePath(down, turned(path, adv/2, (dfoot+y0+y1)/2)) {
				t.Errorf("the inverted interrobang is not the interrobang turned")
			}
			if math.Abs(dfoot-qfoot) > 1 || dtop < ctop || dtop > ctop+(y1-qy1)+1 {
				t.Errorf("the inverted interrobang stands from %v to %v; the inverted question mark reaches %v and the colon %v", dfoot, dtop, qfoot, ctop)
			}
		})
	}
}
