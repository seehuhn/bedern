// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The forms of the special letters which can go wrong when the parameters
// move: the bars must cross their stems and keep clear of the rest of the
// letter, the stroke of ø must cut the counter in two, ß must keep its open
// lower terminal and its waist, and the shared stroke of œ and æ must stay one
// stroke.

// barCase is a bar of a barred letter: its height, and the stretch of it which
// lies outside the stem, from x0 to x1.
type barCase struct {
	glyph  string
	y      func(p design.Params) float64
	x0, x1 func(p design.Params) float64
}

func barCases() []barCase {
	lc := func(f float64) func(design.Params) float64 {
		return func(p design.Params) float64 { return f * p.Ascender }
	}
	cp := func(f float64) func(design.Params) float64 {
		return func(p design.Params) float64 { return f * p.CapHeight }
	}
	dStem := func(p design.Params) float64 { return design.DStemLeft(p) + design.AscenderInset(p) }
	hStem := func(p design.Params) float64 { return p.SideStemLeft + design.AscenderInset(p) }
	dLeft := func(p design.Params) float64 { return capBowlStem(p, false).Left }
	return []barCase{
		{"dcroat", lc(dcroatCentre),
			func(p design.Params) float64 { return dStem(p) - 1.5*p.Stem },
			func(p design.Params) float64 { return dStem(p) - 10 }},
		{"hbar", lc(hbarCentre),
			func(p design.Params) float64 { return hStem(p) + p.AscenderStem + 10 },
			func(p design.Params) float64 { return hStem(p) + 1.5*p.Stem }},
		{"tbar", func(p design.Params) float64 { return tbarCentre * p.XHeight },
			func(p design.Params) float64 { return design.TStemLeft(p) + p.Stem + 5 },
			func(p design.Params) float64 { return design.TStemLeft(p) + p.Stem + p.TBarRight - 5 }},
		{"Dcroat", cp(capDcroatCentre),
			func(p design.Params) float64 { return dLeft(p) + p.Caps.Stem + 10 },
			func(p design.Params) float64 { return dLeft(p) + p.Caps.Stem + 1.5*p.Caps.Stem }},
		{"Hbar", cp(capHbarCentre),
			func(p design.Params) float64 { return design.CapStemLeft(p) + p.Caps.Stem + 10 },
			func(p design.Params) float64 { return design.CapStemLeft(p) + p.Caps.Stem + p.Caps.HCounter - 10 }},
	}
}

// TestBarsKeepClear asks that the bar of đ, ħ, ŧ, Đ and Ħ, where it lies
// outside the stem, keeps the clearance floor from the ink above and below it:
// the bowl of d, the arch of h, the crossbar of t and of H.
func TestBarsKeepClear(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		for _, c := range barCases() {
			path, _ := cachedOutline(np, c.glyph)
			y, x0, x1 := c.y(np.params), c.x0(np.params), c.x1(np.params)
			floor := design.ClearanceFloor(np.params)
			least := math.Inf(1)
			for x := x0; x <= x1; x += 4 {
				rs := vruns(path, x)
				for i, r := range rs {
					if r[0] > y || r[1] < y {
						continue
					}
					if i > 0 {
						least = math.Min(least, r[0]-rs[i-1][1])
					}
					if i+1 < len(rs) {
						least = math.Min(least, rs[i+1][0]-r[1])
					}
				}
			}
			if least < floor {
				t.Errorf("%s: the bar of %s comes %.1f from the rest of the letter, want at least %.1f", np.name, c.glyph, least, floor)
			}
		}
	}
}

// TestBarsCrossTheStem asks that the bars of ł, Ł, đ, Đ, ħ, Ħ, ŧ and Ŧ reach
// beyond both edges of the stems they cross, by at least a quarter of the
// stem, at the heights of the bar.
func TestBarsCrossTheStem(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		p := np.params
		for _, c := range []struct {
			glyph, base string
			y, band     float64 // the middle of the bar and half the band of heights it covers at the stem
			stems       []int   // the runs of the base at y which are stems the bar crosses; -1 is the last
		}{
			{"lslash", "l", lslashCentre * p.Ascender, 0.15 * p.Ascender, []int{0}},
			{"Lslash", "L", capLslashCentre * p.CapHeight, 0.15 * p.CapHeight, []int{0}},
			{"dcroat", "d", dcroatCentre * p.Ascender, 5, []int{-1}},
			{"Dcroat", "D", capDcroatCentre * p.CapHeight, 5, []int{0}},
			{"hbar", "h", hbarCentre * p.Ascender, 5, []int{0}},
			{"Hbar", "H", capHbarCentre * p.CapHeight, 5, []int{0, -1}},
			{"tbar", "t", tbarCentre * p.XHeight, 5, []int{0}},
			{"Tbar", "T", capTbarCentre * p.CapHeight, 5, []int{0}},
		} {
			barred, _ := cachedOutline(np, c.glyph)
			base, _ := cachedOutline(np, c.base)
			stems := runs(base, c.y)
			for _, i := range c.stems {
				if i < 0 {
					i += len(stems)
				}
				st := stems[i]
				left, right := st[0], st[1]
				for y := c.y - c.band; y <= c.y+c.band; y += 2 {
					for _, r := range runs(barred, y) {
						if r[0] <= st[1] && r[1] >= st[0] {
							left, right = math.Min(left, r[0]), math.Max(right, r[1])
						}
					}
				}
				w := st[1] - st[0]
				if st[0]-left < w/4 || right-st[1] < w/4 {
					t.Errorf("%s: the bar of %s reaches %.1f left and %.1f right of a stem of %.1f", np.name, c.glyph, st[0]-left, right-st[1], w)
				}
			}
		}
	}
}

// TestOslashCutsTheCounter asks that the stroke of ø and Ø stands free in the
// middle of the counter, with at least a third of the counter's width of
// white on either side at half the height.
func TestOslashCutsTheCounter(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		for _, c := range []struct {
			glyph string
			y     float64
		}{{"oslash", np.params.XHeight / 2}, {"Oslash", np.params.CapHeight / 2}} {
			path, _ := cachedOutline(np, c.glyph)
			rs := runs(path, c.y)
			if len(rs) != 3 {
				t.Errorf("%s: %s has %d runs of ink at half the height, want 3", np.name, c.glyph, len(rs))
				continue
			}
			counter := rs[2][0] - rs[0][1]
			if l, r := rs[1][0]-rs[0][1], rs[2][0]-rs[1][1]; l < counter/3 || r < counter/3 {
				t.Errorf("%s: the stroke of %s leaves %.1f and %.1f of white in a counter of %.1f", np.name, c.glyph, l, r, counter)
			}
		}
	}
}

// TestEszettForm asks that ß and ẞ keep their form: three runs of ink just
// above the baseline (the stem, the ball and the lower bowl), with the white
// between the stem and the ball at least the clearance floor; the waist on the
// right between 0.57 and 0.76 of the ink width from the left of the ink; and
// the upper bowl narrower than the lower one.
func TestEszettForm(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			eszettForm(t, np)
		})
	}
}

func eszettForm(t *testing.T, np namedParams) {
	{
		p := np.params
		for _, f := range []eszettFrame{newEszettFrame(p), newCapEszettFrame(p)} {
			path, _ := cachedOutline(np, f.name)
			x0, _, x1, _ := path.BBox()
			rs := runs(path, 0.1*p.XHeight)
			if len(rs) != 3 {
				t.Errorf("%s: %s has %d runs of ink at 0.1 of the x-height, want 3", np.name, f.name, len(rs))
			} else if gap := rs[1][0] - rs[0][1]; gap < design.ClearanceFloor(p) {
				t.Errorf("%s: the ball of %s is %.1f from the stem", np.name, f.name, gap)
			}
			waist, upper, lower := math.Inf(1), 0.0, 0.0
			for y := 0.3 * f.top; y < 0.9*f.top; y += 5 {
				rs := runs(path, y)
				if len(rs) == 0 {
					continue
				}
				r := rs[len(rs)-1][1]
				waist = math.Min(waist, r)
			}
			for y := f.waist; y < f.top-20; y += 5 {
				if rs := runs(path, y); len(rs) > 0 {
					upper = math.Max(upper, rs[len(rs)-1][1])
				}
			}
			for y := 20.0; y < f.waist; y += 5 {
				if rs := runs(path, y); len(rs) > 0 {
					lower = math.Max(lower, rs[len(rs)-1][1])
				}
			}
			w := (waist - x0) / (x1 - x0)
			t.Logf("%s: %s: waist at %.3f of the ink, upper bowl %.1f, lower %.1f", np.name, f.name, w, upper, lower)
			if w < 0.57 || w > 0.76 {
				t.Errorf("%s: the waist of %s lies at %.3f of the ink width, want 0.57 to 0.76", np.name, f.name, w)
			}
			if upper >= lower {
				t.Errorf("%s: the upper bowl of %s reaches %.1f, the lower %.1f", np.name, f.name, upper, lower)
			}
		}
	}
}

// TestSharedStroke asks that the stroke which o and e share in œ, and a and e
// in æ, is one stroke: at the height of the thickest part of e, or one unit
// below the bar of e where the bar stands on it, it is at least as thick as
// the round stroke of e and at most 1.15 times as thick.
func TestSharedStroke(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		p := np.params
		for _, c := range []struct {
			glyph string
			dx    float64
		}{
			{"oe", math.Round(design.OBowlRight(p) - p.CurveLower + joinOverlap - design.EBowlLeft(p))},
			{"ae", math.Round(design.AStemLeft(p) + aeOverlap - design.EBowlLeft(p))},
		} {
			path, _ := cachedOutline(np, c.glyph)
			x := c.dx + design.EBowlLeft(p) + p.CurveLower/2 // the middle of the left stroke of e
			w := math.NaN()
			y := eThickY(p)
			if low, _ := eBar(p); low <= y {
				y = low - 1
			}
			for _, r := range runs(path, y) {
				if r[0] <= x && r[1] >= x {
					w = r[1] - r[0]
				}
			}
			if !(w >= p.CurveLower && w <= 1.15*p.CurveLower) {
				t.Errorf("%s: the shared stroke of %s is %.1f thick, want %.1f to %.1f", np.name, c.glyph, w, p.CurveLower, 1.15*p.CurveLower)
			}
		}
	}
}

// TestSuperiorFigures asks that the stem of the superior one is 0.67 to 0.94
// as thick as the stem of one, which a plain scaling by supScale would miss,
// and that the figures of a fraction keep at least 0.9 fracGap from the
// fraction slash.
func TestSuperiorFigures(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		if !boldApplies(t, np) {
			continue
		}
		p := np.params
		one, _ := cachedOutline(np, "one")
		sup, _ := cachedOutline(np, "uni00B9")
		_, y0, _, y1 := one.BBox()
		_, s0, _, s1 := sup.BBox()
		widest := func(path outline.Path, y float64) float64 {
			w := 0.0
			for _, r := range runs(path, y) {
				w = math.Max(w, r[1]-r[0])
			}
			return w
		}
		if r := widest(sup, (s0+s1)/2) / widest(one, (y0+y1)/2); r < 0.67 || r > 0.94 {
			t.Errorf("%s: the stem of the superior one is %.3f of that of one, want 0.67 to 0.94", np.name, r)
		}
		for _, g := range []string{"onequarter", "onehalf", "threequarters"} {
			path, _ := cachedOutline(np, g)
			bar := fractionBar(p)
			var slash, others outline.Path
			for _, c := range path {
				if b := (outline.Path{c}); sameBox(b, bar) {
					slash = b
				} else {
					others = append(others, c)
				}
			}
			if slash == nil {
				t.Errorf("%s: %s has no fraction slash", np.name, g)
				continue
			}
			if d := pathsDistance(slash, others, fracGap); d < 0.9*fracGap {
				t.Errorf("%s: the figures of %s come %.1f from the slash, want at least %.1f", np.name, g, d, 0.9*fracGap)
			}
		}
	}
}

// sameBox reports whether the two paths have boxes of the same size.
func sameBox(a, b outline.Path) bool {
	ax0, ay0, ax1, ay1 := a.BBox()
	bx0, by0, bx1, by1 := b.BBox()
	return math.Abs((ax1-ax0)-(bx1-bx0)) < 0.5 && math.Abs((ay1-ay0)-(by1-by0)) < 0.5
}

// TestFurtherPunctuation asks that the guillemets pointing left and right are
// mirror images with mirrored side spaces, that the inverted marks stand
// between 1.5 descenders below the baseline and the top of the x-height, and
// that the double dagger is symmetric about its middle.
func TestFurtherPunctuation(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		p := np.params
		for _, pair := range [][2]string{{"guillemotleft", "guillemotright"}, {"guilsinglleft", "guilsinglright"}} {
			l, r := cachedRounded(np, pair[0]), cachedRounded(np, pair[1])
			lx0, ly0, lx1, ly1 := l.Path.BBox()
			rx0, ry0, rx1, ry1 := r.Path.BBox()
			if l.Advance != r.Advance || math.Abs(lx0-(r.Advance-rx1)) > 1 || math.Abs((l.Advance-lx1)-rx0) > 1 || ly0 != ry0 || ly1 != ry1 {
				t.Errorf("%s: %s and %s are not mirror images: advances %v and %v, boxes %v %v %v %v and %v %v %v %v",
					np.name, pair[0], pair[1], l.Advance, r.Advance, lx0, ly0, lx1, ly1, rx0, ry0, rx1, ry1)
			}
		}
		for _, g := range []string{"exclamdown", "questiondown"} {
			path, _ := cachedOutline(np, g)
			_, y0, _, y1 := path.BBox()
			if y0 < -1.5*p.Descender || y1 > p.XHeight+p.OvershootLower {
				t.Errorf("%s: %s reaches from %.1f to %.1f, want within %.1f to %.1f", np.name, g, y0, y1, -1.5*p.Descender, p.XHeight+p.OvershootLower)
			}
		}
		path, _ := cachedOutline(np, "daggerdbl")
		x0, y0, x1, y1 := path.BBox()
		back := turned(path, (x0+x1)/2, (y0+y1)/2)
		for y := y0 + 5; y < y1; y += 23 {
			a, b := runs(path, y), runs(back, y)
			if len(a) != len(b) {
				t.Errorf("%s: the double dagger has %d runs at %.0f and %d turned", np.name, len(a), y, len(b))
				continue
			}
			for i := range a {
				if math.Abs(a[i][0]-b[i][0]) > 1 || math.Abs(a[i][1]-b[i][1]) > 1 {
					t.Errorf("%s: the double dagger is not symmetric at %.0f: %v against %v", np.name, y, a[i], b[i])
				}
			}
		}
	}
}
