// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The rounds of a glyph are found on its upright outline, following
// the roman code: a counter is a contour whose orientation is opposite to that
// of the largest contour, and its round has the centre and the half axes of
// its bounding box.

// counters returns the rounds of the counters of the upright path, turning
// the directions from from to to with the given taper (0 and 360 for the
// whole counter).
func counters(path outline.Path, from, to, taper float64) []round {
	big, area := -1, 0.0
	for i, c := range path {
		if a := math.Abs(c.SignedArea()); a > area {
			big, area = i, a
		}
	}
	if big < 0 {
		return nil
	}
	sign := math.Signbit(path[big].SignedArea())
	var out []round
	for i, c := range path {
		if i == big || math.Signbit(c.SignedArea()) == sign {
			continue
		}
		x0, y0, x1, y1 := outline.Path{c}.BBox()
		out = append(out, round{C: pt((x0+x1)/2, (y0+y1)/2), A: (x1 - x0) / 2, B: (y1 - y0) / 2, From: from, To: to, Taper: taper})
	}
	return out
}

// roundsOf returns a rounds function which takes the counters of the glyph,
// all of each counter.
func allCounters(name string) func(Params, design.Params) []round {
	return func(_ Params, rp design.Params) []round {
		path, _, _ := romanOutline(name, rp)
		return counters(path, 0, 360, 0)
	}
}

// bowlCounters returns a rounds function which takes the counters of a
// glyph whose bowls hang on a stem on their left: the side of each counter
// away from the stem.
func bowlCounters(name string) func(Params, design.Params) []round {
	return func(_ Params, rp design.Params) []round {
		path, _, _ := romanOutline(name, rp)
		return counters(path, -70, 70, 25)
	}
}

// openRound returns a rounds function for a round like O which is open on the
// right (C and G): the inner edge of the round of O between the directions
// from and to.
func openRound(from, to float64) func(Params, design.Params) []round {
	return func(_ Params, rp design.Params) []round {
		r := capOCounter(rp)
		r.From, r.To, r.Taper = from, to, 20
		return []round{r}
	}
}

// bottomRound returns a rounds function for the round at the bottom of U and
// J: the counter between the stems at the height mid, from its lowest point
// on the vertical through its middle, and the directions round the bottom.
func bottomRound(name string, mid float64) func(Params, design.Params) []round {
	return func(_ Params, rp design.Params) []round {
		path, _, _ := romanOutline(name, rp)
		xs := sorted(path.CrossingsX(mid))
		if len(xs) < 4 {
			return nil
		}
		l, r := xs[1], xs[2]
		cx := (l + r) / 2
		ys := sorted(path.CrossingsY(cx))
		if len(ys) < 2 {
			return nil
		}
		bottom := ys[1]
		b := mid - bottom
		return []round{{C: pt(cx, mid), A: (r - l) / 2, B: b, From: 200, To: 340, Taper: 25}}
	}
}

// sRounds returns the rounds of S: the upper counter, turned over the top,
// and the lower counter, turned under the bottom.
func sRounds(name string) func(Params, design.Params) []round {
	return func(_ Params, rp design.Params) []round {
		path, _, _ := romanOutline(name, rp)
		x0, _, x1, _ := path.BBox()
		cx := (x0 + x1) / 2
		ys := sorted(path.CrossingsY(cx))
		if len(ys) < 6 {
			return nil
		}
		var out []round
		for _, part := range [][2]float64{{ys[1], ys[2]}, {ys[3], ys[4]}} {
			cy := (part[0] + part[1]) / 2
			xs := sorted(path.CrossingsX(cy))
			if len(xs) < 2 {
				continue
			}
			a := cx - xs[1]
			if part[0] > ys[2] {
				a = cx - xs[1]
			} else {
				a = xs[len(xs)-2] - cx
			}
			r := round{C: pt(cx, cy), A: math.Abs(a), B: (part[1] - part[0]) / 2, Taper: 25, Tol: 0.1}
			if part[0] > ys[2] {
				r.From, r.To = 40, 170
			} else {
				r.From, r.To = 220, 350
			}
			out = append(out, r)
		}
		return out
	}
}

// romanOutline returns the upright outline of the roman glyph in the
// parameters of the italic, with explicit closing segments.
func romanOutline(name string, rp design.Params) (outline.Path, float64, bool) {
	src := sources[name]
	rn := src.roman
	if rn == "" {
		rn = name
	}
	path, adv, ok := glyphsOutline(rn, rp)
	return explicit(path), adv, ok
}

// sorted returns xs sorted.
func sorted(xs []float64) []float64 {
	sort.Float64s(xs)
	return xs
}
