// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The letters of the trademark and of the service mark stand with their tops
// at the cap height and their feet above markFoot of the x-height; the white
// between the two letters lies between markWhiteMin and markWhiteMax, and the
// stem of the T is markStemMin to markStemMax of the stem of the capitals.
const (
	markFoot     = 0.5
	markWhiteMin = 20.0
	markWhiteMax = 60.0
	markStemMin  = 0.45
	markStemMax  = 0.82
)

func TestMarksAreSmallCapitals(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		for _, name := range []string{"trademark", "uni2120"} {
			path, _ := cachedOutline(np, name)
			if len(path) != 2 {
				t.Errorf("%s: %s has %d contours, want 2", np.name, name, len(path))
				continue
			}
			_, y0, ax1, _ := path[0:1].BBox()
			mx0, my0, _, my1 := path[1:2].BBox()
			if math.Abs(my1-p.CapHeight) > 1 {
				t.Errorf("%s: the top of the M of %s is at %.1f, the cap height %.0f", np.name, name, my1, p.CapHeight)
			}
			if lo := math.Min(y0, my0); lo < markFoot*p.XHeight {
				t.Errorf("%s: %s reaches down to %.0f, want at least %.0f", np.name, name, lo, markFoot*p.XHeight)
			}
			if w := mx0 - ax1; w < markWhiteMin || w > markWhiteMax {
				t.Errorf("%s: %.1f of white between the letters of %s, want %.0f to %.0f", np.name, w, name, markWhiteMin, markWhiteMax)
			}
		}
		path, _ := cachedOutline(np, "trademark")
		_, y0, _, y1 := path[0:1].BBox()
		rs := runs(path[0:1], (y0+y1)/2)
		if len(rs) != 1 {
			t.Errorf("%s: the T of the trademark has %d runs of ink at half its height", np.name, len(rs))
			continue
		}
		if w := (rs[0][1] - rs[0][0]) / p.Caps.Stem; w < markStemMin || w > markStemMax {
			t.Errorf("%s: the stem of the T of the trademark is %.2f capital stems, want %.2f to %.2f", np.name, w, markStemMin, markStemMax)
		}
	}
}

// boxes returns the boxes of the contours of the path, each as left, bottom,
// right and top.
func boxes(path outline.Path) [][4]float64 {
	var bs [][4]float64
	for i := range path {
		x0, y0, x1, y1 := path[i : i+1].BBox()
		bs = append(bs, [4]float64{x0, y0, x1, y1})
	}
	return bs
}

// The published sign and the copyleft sign have the advance and the ring of
// the copyright sign, and the letter of the copyleft sign is that of the
// copyright sign mirrored about the middle of the ring.
func TestCircledSignsShareTheRing(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		c, cadv := cachedOutline(np, "copyright")
		cb := boxes(c)
		for _, name := range []string{"uni2117", "u1F12F"} {
			path, adv := cachedOutline(np, name)
			if adv != cadv {
				t.Errorf("%s: %s has the advance %.0f, the copyright sign %.0f", np.name, name, adv, cadv)
			}
			b := boxes(path)
			if len(b) < 3 || b[0] != cb[0] || b[1] != cb[1] {
				t.Errorf("%s: %s does not have the ring of the copyright sign", np.name, name)
			}
		}
		l, _ := cachedOutline(np, "u1F12F")
		lb := boxes(l)
		if len(lb) != 3 || len(cb) != 3 {
			t.Errorf("%s: the copyleft sign has %d contours and the copyright sign %d, want 3", np.name, len(lb), len(cb))
			continue
		}
		mid := cb[0][0] + cb[0][2]
		want := [4]float64{mid - cb[2][2], cb[2][1], mid - cb[2][0], cb[2][3]}
		for i := range want {
			if math.Abs(lb[2][i]-want[i]) > 1 {
				t.Errorf("%s: the letter of the copyleft sign has the box %.1f, the mirrored letter of the copyright sign %.1f", np.name, lb[2], want)
				break
			}
		}
	}
}

// The ohm sign is symmetric about its middle, stands on two feet, each at
// least ohmFootMin of its ink long, with at least ohmGapMin of its ink
// between them, and the mho sign is the ohm sign turned through half a turn
// about the middle of the cap height.
const (
	ohmFootMin = 0.3
	ohmGapMin  = 0.09
)

func TestOhmAndMho(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		path, adv := cachedOutline(np, "Omega")
		x0, y0, x1, y1 := path.BBox()
		w := x1 - x0
		for _, y := range []float64{0.1 * p.CapHeight, 0.4 * p.CapHeight, 0.7 * p.CapHeight, 0.95 * p.CapHeight} {
			rs := runs(path, y)
			for i, r := range rs {
				m := rs[len(rs)-1-i]
				if math.Abs((r[0]-x0)-(x1-m[1])) > 1 {
					t.Errorf("%s: the ohm sign is not symmetric at the height %.0f", np.name, y)
					break
				}
			}
		}
		feet := runs(path, 10)
		if len(feet) != 2 {
			t.Errorf("%s: the ohm sign has %d runs of ink at its feet, want 2", np.name, len(feet))
			continue
		}
		if f := (feet[0][1] - feet[0][0]) / w; f < ohmFootMin {
			t.Errorf("%s: a foot of the ohm sign is %.2f of its ink long, want at least %.2f", np.name, f, ohmFootMin)
		}
		if g := (feet[1][0] - feet[0][1]) / w; g < ohmGapMin {
			t.Errorf("%s: the feet of the ohm sign are %.2f of its ink apart, want at least %.2f", np.name, g, ohmGapMin)
		}
		mho, madv := cachedOutline(np, "uni2127")
		mx0, my0, mx1, my1 := mho.BBox()
		if madv != adv || math.Abs(mx0-(adv-x1)) > 1e-6 || math.Abs(mx1-(adv-x0)) > 1e-6 ||
			math.Abs(my0-(p.CapHeight-y1)) > 1e-6 || math.Abs(my1-(p.CapHeight-y0)) > 1e-6 {
			t.Errorf("%s: the mho sign is not the ohm sign turned about the middle of the cap height", np.name)
		}
	}
}

// The small o of the numero sign stands numeroWhiteMin to numeroWhiteMax to
// the right of the right stem of the N, with its top at most numeroTop of
// the cap height, over an underline on the baseline which is as wide as the
// o and at least numeroUnder below it.
const (
	numeroWhiteMin = 50.0
	numeroWhiteMax = 125.0
	numeroTop      = 0.91
	numeroUnder    = 45.0
)

func TestNumero(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		path, _ := cachedOutline(np, "uni2116")
		b := boxes(path)
		if len(b) != 4 {
			t.Errorf("%s: the numero sign has %d contours, want 4", np.name, len(b))
			continue
		}
		o, line := b[1], b[3]
		if o[3] > numeroTop*p.CapHeight {
			t.Errorf("%s: the o of the numero sign reaches %.0f, want at most %.0f", np.name, o[3], numeroTop*p.CapHeight)
		}
		if line[1] != 0 || math.Abs(line[0]-o[0]) > 1 || math.Abs(line[2]-o[2]) > 1 {
			t.Errorf("%s: the underline of the numero sign is not on the baseline under the o", np.name)
		}
		if w := o[1] - line[3]; w < numeroUnder {
			t.Errorf("%s: %.0f of white between the underline and the o of the numero sign, want at least %.0f", np.name, w, numeroUnder)
		}
		rs := runs(path[0:1], (o[1]+o[3])/2)
		if len(rs) == 0 {
			t.Errorf("%s: the N of the numero sign has no ink beside the o", np.name)
			continue
		}
		if w := o[0] - rs[len(rs)-1][1]; w < numeroWhiteMin || w > numeroWhiteMax {
			t.Errorf("%s: %.0f of white between the N and the o of the numero sign, want %.0f to %.0f", np.name, w, numeroWhiteMin, numeroWhiteMax)
		}
	}
}

// The sides of the estimated sign are at least estContrast times as thick
// as its top, its bar and its bottom, which are no thinner than the thin
// strokes of the lowercase; the eye above the bar is closed, and below the
// bar the sign has no right side.
const estContrast = 2.0

func TestEstimated(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		path, _ := cachedOutline(np, "estimated")
		x0, y0, x1, y1 := path.BBox()
		h := y1 - y0
		thin := math.Inf(1)
		vr := vruns(path, (x0+x1)/2)
		if len(vr) != 3 {
			t.Errorf("%s: the estimated sign has %d runs of ink on its vertical axis, want 3", np.name, len(vr))
			continue
		}
		for _, r := range vr {
			thin = math.Min(thin, r[1]-r[0])
		}
		if thin < p.ThinLower-1 {
			t.Errorf("%s: the thinnest stroke of the estimated sign is %.1f, the thin strokes of the lowercase %.0f", np.name, thin, p.ThinLower)
		}
		upper := runs(path, y0+0.75*h)
		if len(upper) != 2 {
			t.Errorf("%s: the estimated sign has %d runs of ink above its bar, want 2", np.name, len(upper))
			continue
		}
		for _, r := range upper {
			if r[1]-r[0] < estContrast*thin {
				t.Errorf("%s: a side of the estimated sign is %.0f thick above the bar, the thin strokes %.0f", np.name, r[1]-r[0], thin)
			}
		}
		lower := runs(path, y0+0.35*h)
		if len(lower) != 1 || lower[0][1] > (x0+x1)/2 {
			t.Errorf("%s: the estimated sign has ink on its right side below the bar", np.name)
		}
	}
}

// The stroke of the prescription sign shows on both sides of the leg of the
// R: on each side of the middle line of the leg, at least recipeFree square
// units of the sign are ink which the R does not have.
const recipeFree = 1000.0

func TestPrescriptionStrokeCrossesTheLeg(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		p := np.params
		path, adv := cachedOutline(np, "prescription")
		r, radv := cachedOutline(np, "R")
		if adv != radv {
			t.Errorf("%s: the prescription sign has the advance %.0f, R %.0f", np.name, adv, radv)
		}
		leg := capRLeg(p)
		cy := recipeCross * p.CapHeight
		cx := (leg.edgeAt(edgeLeft, cy).X + leg.edgeAt(edgeRight, cy).X) / 2
		d := leg.down()
		const step = 5.0
		var free [2]float64
		reach := 0.3 * p.CapHeight
		for x := cx - reach; x <= cx+reach; x += step {
			for y := cy - reach; y <= cy+reach; y += step {
				q := outline.Point{X: x, Y: y}
				if !inkAt(path, q) || inkAt(r, q) {
					continue
				}
				if (x-cx)*d.Y-(y-cy)*d.X > 0 {
					free[0] += step * step
				} else {
					free[1] += step * step
				}
			}
		}
		for _, a := range free {
			if a < recipeFree {
				t.Errorf("%s: %.0f square units of the stroke of the prescription sign show on one side of the leg, want at least %.0f", np.name, a, recipeFree)
			}
		}
	}
}

// The ring and the C of the degree Celsius are celsiusWhiteMin to
// celsiusWhiteMax apart where they are nearest to each other.
const (
	celsiusWhiteMin = 68.0
	celsiusWhiteMax = 125.0
)

func TestCelsius(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		path, _ := cachedOutline(np, "uni2103")
		if len(path) != 3 {
			t.Errorf("%s: the degree Celsius has %d contours, want 3", np.name, len(path))
			continue
		}
		if w := pathsDistance(path[:2], path[2:], 2*celsiusWhiteMax); w < celsiusWhiteMin || w > celsiusWhiteMax {
			t.Errorf("%s: %.0f of white between the ring and the C of the degree Celsius, want %.0f to %.0f", np.name, w, celsiusWhiteMin, celsiusWhiteMax)
		}
	}
}
