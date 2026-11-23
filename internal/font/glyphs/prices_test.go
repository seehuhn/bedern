// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"
)

// The currency signs which stand in a column of prices, ¥, €, $, £ and ¢, have
// the advance of the figures and their ink centred in it within
// priceCentred, as the figures have, so that the figures after them align.
const priceCentred = 1.0

func TestPriceSignsHaveTheFiguresAdvance(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		fa := np.params.Figures.Advance
		for _, name := range []string{"yen", "Euro", "dollar", "sterling", "cent"} {
			path, adv, _ := outlineOf(name, np.params)
			if adv != fa {
				t.Errorf("%s: %s has the advance %.0f, the figures %.0f", np.name, name, adv, fa)
			}
			x0, _, x1, _ := path.BBox()
			if d := (x0 - (adv - x1)) / 2; math.Abs(d) > priceCentred {
				t.Errorf("%s: the ink of %s stands %.1f off the middle of its advance", np.name, name, d)
			}
		}
	}
}

// The bars of ¥ stand below its crotch, the top of the upper bar at least
// yenForkGap below it, with at least yenFootWhite of white between the lower
// bar and the foot serif of the stem.
const (
	yenForkGap   = 20.0
	yenFootWhite = 90.0
)

func TestYenBarsBelowTheFork(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		path, _, _ := outlineOf("yen", p)
		x0, _, x1, y1 := path.BBox()
		foot := yenScale * (p.Caps.SerifThick + p.Caps.BracketHeight)
		// The crotch is the lowest height above the foot at which the scan
		// meets the two arms.
		crotch := math.NaN()
		for y := foot; y < y1; y++ {
			if len(runs(path, y)) == 2 {
				crotch = y
				break
			}
		}
		if math.IsNaN(crotch) {
			t.Fatalf("%s: ¥ has no crotch", np.name)
		}
		// The bars are the bands of heights above the foot serif at which the
		// ink is one run wider than half the ink.
		var bars [][2]float64
		in := false
		for y := foot; y < crotch; y++ {
			rs := runs(path, y)
			wide := len(rs) == 1 && rs[0][1]-rs[0][0] > 0.5*(x1-x0)
			switch {
			case wide && !in:
				bars, in = append(bars, [2]float64{y, y}), true
			case wide:
				bars[len(bars)-1][1] = y
			default:
				in = false
			}
		}
		if len(bars) != 2 {
			t.Fatalf("%s: ¥ has %d bars, want 2", np.name, len(bars))
		}
		if g := crotch - bars[1][1]; g < yenForkGap {
			t.Errorf("%s: the upper bar of ¥ ends %.0f below the crotch, want at least %.0f", np.name, g, yenForkGap)
		}
		if g := bars[0][0] - foot; g < yenFootWhite {
			t.Errorf("%s: the lower bar of ¥ stands %.0f above the foot serif, want at least %.0f", np.name, g, yenFootWhite)
		}
	}
}

// The ring of ¤ has the stress of the face, heavier at its sides than at its
// top, where it is heavier than the hairline, and its spokes are heavier than
// the hairline and lighter than the stem.
func TestCurrencySignWeights(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		path, _, _ := outlineOf("currency", p)
		x0, y0, x1, y1 := path.BBox()
		side := runs(path, (y0+y1)/2)
		top := vruns(path, (x0+x1)/2)
		if len(side) != 2 || len(top) != 2 {
			t.Fatalf("%s: the ring of ¤ has %d and %d runs across its middle", np.name, len(side), len(top))
		}
		ws, wt := side[0][1]-side[0][0], top[1][1]-top[1][0]
		if wt <= p.ThinLower || ws <= wt {
			t.Errorf("%s: the ring of ¤ is %.1f at its side and %.1f at its top, the hairline %.1f", np.name, ws, wt, p.ThinLower)
		}
		if _, _, spoke := currencyWeights(p); spoke <= p.ThinLower || spoke >= p.Stem {
			t.Errorf("%s: the spokes of ¤ are %.1f, the hairline %.1f and the stem %.1f", np.name, spoke, p.ThinLower, p.Stem)
		}
	}
}

// The tilde is thin where it rises and thick where it falls: measured at right
// angles to its centre line over the middle tildeMiddle of its width, its
// thickest part is at least tildeContrast times its thinnest, and the
// thickest lies between the crest and the trough.  At its ends, tildeEndIn of
// its width from each end, it is at most tildeEndRatio of its thickest part,
// so that the ends are round and not cut square.
const (
	tildeMiddle   = 0.7
	tildeContrast = 1.5
	tildeEndIn    = 0.05
	tildeEndRatio = 0.75
)

func TestTildeContrast(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		for _, f := range []markForm{formLower, formCap} {
			path := markOutline(np.params, markTilde, f)
			x0, _, x1, _ := path.BBox()
			w := x1 - x0
			// thick returns the thickness at right angles to the centre line
			// at x, and the height of the centre line.
			thick := func(x float64) (float64, float64) {
				centre := func(x float64) (float64, float64) {
					vr := vruns(path, x)
					if len(vr) != 1 {
						return math.NaN(), math.NaN()
					}
					return (vr[0][0] + vr[0][1]) / 2, vr[0][1] - vr[0][0]
				}
				ya, _ := centre(x - 1)
				yb, _ := centre(x + 1)
				yc, v := centre(x)
				return v * math.Cos(math.Atan((yb-ya)/2)), yc
			}
			lo, hi, hiX := math.Inf(1), 0.0, 0.0
			crest, trough := math.Inf(-1), math.Inf(1)
			crestX, troughX := 0.0, 0.0
			for x := x0 + (1-tildeMiddle)/2*w; x <= x1-(1-tildeMiddle)/2*w; x++ {
				d, yc := thick(x)
				if math.IsNaN(d) {
					t.Fatalf("%s: the tilde is not one stroke at %.0f", np.name, x)
				}
				lo = math.Min(lo, d)
				if d > hi {
					hi, hiX = d, x
				}
				if yc > crest {
					crest, crestX = yc, x
				}
				if yc < trough {
					trough, troughX = yc, x
				}
			}
			if hi < tildeContrast*lo {
				t.Errorf("%s, form %d: the tilde is %.1f at its thickest and %.1f at its thinnest, want at least %.1f times", np.name, f, hi, lo, tildeContrast)
			}
			if hiX < crestX || hiX > troughX {
				t.Errorf("%s, form %d: the tilde is thickest at %.0f, outside the fall from %.0f to %.0f", np.name, f, hiX, crestX, troughX)
			}
			for _, x := range []float64{x0 + tildeEndIn*w, x1 - tildeEndIn*w} {
				if d, _ := thick(x); !(d <= tildeEndRatio*hi) {
					t.Errorf("%s, form %d: the tilde is %.1f near its end at %.0f, %.2f of its thickest", np.name, f, d, x, d/hi)
				}
			}
		}
	}
}

// The descender of µ ends round: muEndIn of the descender above its lowest
// point it is at most muEndRatio of the stem across.
const (
	muEndIn    = 0.02
	muEndRatio = 0.6
)

func TestMuEndsRound(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		path, _, _ := outlineOf("mu", p)
		_, y0, _, _ := path.BBox()
		rs := runs(path, y0+muEndIn*p.Descender)
		if len(rs) != 1 {
			t.Fatalf("%s: µ has %d runs near the end of its descender", np.name, len(rs))
		}
		if w := rs[0][1] - rs[0][0]; w > muEndRatio*p.Stem {
			t.Errorf("%s: the descender of µ is %.1f across near its end, the stem %.1f", np.name, w, p.Stem)
		}
	}
}
