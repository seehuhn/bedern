// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The least white between two bars or two strokes of a currency sign, and
// between a stroke and the sides of the counter in which it stands.  The
// bars of the lira, which stand between the loop and the ball of the
// sterling, keep liraBarWhite.
const (
	currencyBarWhite    = 40.0
	currencyStrokeWhite = 40.0
	liraBarWhite        = 30.0
)

// wideBands returns the bands of heights between lo and hi at which the ink
// of the path is one run at least frac of its ink width wide.
func wideBands(path outline.Path, lo, hi, frac float64) [][2]float64 {
	x0, _, x1, _ := path.BBox()
	var bands [][2]float64
	in := false
	for y := lo; y < hi; y++ {
		wide := false
		for _, r := range runs(path, y+0.5) {
			if r[1]-r[0] >= frac*(x1-x0) {
				wide = true
			}
		}
		switch {
		case wide && !in:
			bands, in = append(bands, [2]float64{y, y + 1}), true
		case wide:
			bands[len(bands)-1][1] = y + 1
		default:
			in = false
		}
	}
	return bands
}

// The won, the naira, the peso and the lira have two bars, which are at
// least as thick as the thin strokes of the lowercase, with at least
// currencyBarWhite of white between them, above them and below them.
func TestCurrencyBars(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		h := p.CapHeight
		pc, _, _ := outlineOf("P", p)
		_, py0, _, py1 := pc[1:2].BBox()
		for _, c := range []struct {
			name   string
			lo, hi float64
			frac   float64
			white  float64
		}{
			{"uni20A9", 0.15 * h, 0.85 * h, 0.9, currencyBarWhite},
			{"uni20A6", 0.15 * h, 0.85 * h, 0.9, currencyBarWhite},
			{"uni20B1", py0, py1, 0.9, currencyBarWhite},
			{"lira", 0.25 * p.Figures.Height, 0.7 * p.Figures.Height, 0.55, liraBarWhite},
		} {
			path, _ := cachedOutline(np, c.name)
			bars := wideBands(path, c.lo, c.hi, c.frac)
			if len(bars) != 2 {
				t.Errorf("%s: %s has %d bars between %.0f and %.0f, want 2", np.name, c.name, len(bars), c.lo, c.hi)
				continue
			}
			for _, b := range bars {
				if b[1]-b[0] < p.ThinLower-1 {
					t.Errorf("%s: a bar of %s is %.0f thick, the thin strokes %.0f", np.name, c.name, b[1]-b[0], p.ThinLower)
				}
			}
			for _, w := range []float64{bars[0][0] - c.lo, bars[1][0] - bars[0][1], c.hi - bars[1][1]} {
				if w < c.white {
					t.Errorf("%s: %s has %.0f of white beside a bar, want at least %.0f", np.name, c.name, w, c.white)
				}
			}
		}
	}
}

// The counters of the won and of the naira, the whites which the bars and
// the strokes of the letter enclose, are at least currencyCounter wide and
// high.
const currencyCounter = 20.0

func TestCurrencyCountersAreOpen(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		for _, name := range []string{"uni20A9", "uni20A6"} {
			path, _ := cachedOutline(np, name)
			for i := 1; i < len(path); i++ {
				x0, y0, x1, y1 := path[i : i+1].BBox()
				if x1-x0 < currencyCounter || y1-y0 < currencyCounter {
					t.Errorf("%s: a counter of %s is %.0f wide and %.0f high, want at least %.0f", np.name, name, x1-x0, y1-y0, currencyCounter)
				}
			}
		}
	}
}

// strokeRuns returns the runs of ink of the horizontal scan of the glyph at
// the height y.
func strokeRuns(np namedParams, glyph string, y float64) [][2]float64 {
	path, _ := cachedOutline(np, glyph)
	return runs(path, y)
}

// The strokes of the colon sign, the baht, the guarani and the dollar of the
// old-style figures stand free above the letter, and inside the letter they
// have white on both sides.
func TestCurrencyStrokesShow(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		h := p.CapHeight
		above := h + p.OvershootUpper + 15
		check := func(glyph string, y float64, want int, strokes []int, white float64) {
			rs := strokeRuns(np, glyph, y)
			if len(rs) != want {
				t.Errorf("%s: %s has %d runs of ink at the height %.0f, want %d", np.name, glyph, len(rs), y, want)
				return
			}
			for _, i := range strokes {
				if i > 0 && rs[i][0]-rs[i-1][1] < white {
					t.Errorf("%s: %s at the height %.0f: %.1f of white left of a stroke, want at least %.0f", np.name, glyph, y, rs[i][0]-rs[i-1][1], white)
				}
				if i+1 < len(rs) && rs[i+1][0]-rs[i][1] < white {
					t.Errorf("%s: %s at the height %.0f: %.1f of white right of a stroke, want at least %.0f", np.name, glyph, y, rs[i+1][0]-rs[i][1], white)
				}
			}
		}
		// The colon sign: two strokes above the C, and in the middle of its
		// height the left side of the C and the two strokes.
		check("colonmonetary", above, 2, []int{0}, currencyStrokeWhite)
		check("colonmonetary", h/2, 3, []int{1, 2}, currencyStrokeWhite)
		// The baht: the stem, the stroke and the side of the bowl, in both
		// bowls.
		check("uni0E3F", above, 1, nil, 0)
		check("uni0E3F", 0.25*h, 3, []int{1}, currencyStrokeWhite)
		check("uni0E3F", 0.77*h, 3, []int{1}, currencyStrokeWhite)
		// The guarani: the left side of the G, the stroke and the beak.
		check("uni20B2", above, 1, nil, 0)
		check("uni20B2", 0.62*h, 3, []int{1}, currencyStrokeWhite)
		// The dollar of the old-style figures: two strokes above the S, with
		// the clearance floor between them.
		check("dollar.oldstyle", p.Figures.Height+p.OvershootLower+15, 2, []int{0}, clearanceFloor(p))
	}
}

// The underline of the dong lies below the baseline and above the descender
// line, with at least dongWhite of white between it and the foot of the d.
const dongWhite = 30.0

func TestDongUnderline(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		path, _ := cachedOutline(np, "dong")
		x0, y0, x1, _ := path.BBox()
		if y0 < -p.Descender {
			t.Errorf("%s: the underline of the dong reaches %.0f, below the descender line %.0f", np.name, y0, -p.Descender)
		}
		rs := vruns(path, (x0+x1)/2)
		if len(rs) < 2 || rs[0][1] > 0 {
			t.Errorf("%s: the dong has no underline below the baseline in the middle of its ink", np.name)
			continue
		}
		if w := rs[1][0] - rs[0][1]; w < dongWhite {
			t.Errorf("%s: %.0f of white between the underline of the dong and the d, want at least %.0f", np.name, w, dongWhite)
		}
	}
}

// The stem of the florin is one stroke from the crossbar to the tail, no
// thinner than the stem of f and no thicker than the stem of j, and the
// tail reaches at least florinDepth of the descender below the baseline.
const florinDepth = 0.8

func TestFlorinStem(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		path, _ := cachedOutline(np, "florin")
		_, y0, _, _ := path.BBox()
		if y0 > -florinDepth*p.Descender {
			t.Errorf("%s: the florin reaches %.0f, want at most %.0f", np.name, y0, -florinDepth*p.Descender)
		}
		for y := -40.0; y < p.XHeight-p.Hair-5; y += 5 {
			rs := runs(path, y)
			if len(rs) == 0 {
				t.Errorf("%s: the florin has no ink at the height %.0f", np.name, y)
				continue
			}
			// The stem is the rightmost run below the baseline, where the
			// tail lies to its left.
			r := rs[len(rs)-1]
			if w := r[1] - r[0]; w < p.AscenderStem-1 || w > p.Stem+1 {
				t.Errorf("%s: the stem of the florin is %.1f wide at the height %.0f, want %.0f to %.0f", np.name, w, y, p.AscenderStem, p.Stem)
			}
		}
	}
}

// The stroke of the cent of the old-style figures leans to the right: above
// the c it stands at least centOldShift to the right of where it stands
// below the c.
const centOldShift = 40.0

func TestCentOldstyleStrokeLeans(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		up := strokeRuns(np, "cent.oldstyle", p.XHeight+p.OvershootLower+20)
		down := strokeRuns(np, "cent.oldstyle", -p.OvershootLower-20)
		if len(up) != 1 || len(down) != 1 {
			t.Errorf("%s: the stroke of the cent of the old-style figures does not stand free above and below the c", np.name)
			continue
		}
		if d := up[0][0] - down[0][0]; d < centOldShift {
			t.Errorf("%s: the stroke of the cent of the old-style figures stands %.1f further right above the c than below it, want at least %.0f", np.name, d, centOldShift)
		}
	}
}

// The currency signs which are capitals with strokes keep the advance of
// their capital, so that they space as the capital does.
func TestCurrencyCapitalsKeepTheAdvance(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		for _, c := range [][2]string{{"colonmonetary", "C"}, {"uni20A9", "W"}, {"uni20A6", "N"}, {"uni0E3F", "B"}, {"uni20B2", "G"}} {
			_, adv := cachedOutline(np, c[0])
			_, want := cachedOutline(np, c[1])
			if adv != want {
				t.Errorf("%s: %s has the advance %.0f, %s %.0f", np.name, c[0], adv, c[1], want)
			}
		}
	}
}
