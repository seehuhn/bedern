// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The ligatures of f: the stems stand at the distance of two letters n, the
// advance is the advance of the last letter moved by the pitch of the stems,
// the ball takes the place of the dot of i and j, and the hook runs level into
// the top of f and l.

// ligatureNarrowing is the range of the advance of a ligature against the sum
// of the advances of its letters, and ligatureGapRatio that of the gap between
// its stems against the gap in the letters set without a kern.
var (
	ligatureNarrowing = [2]float64{0.87, 1.06}
	ligatureGapRatio  = [2]float64{0.75, 1.06}
)

// stemRuns returns the runs of ink of the path at the height y, as pairs of
// the left and the right edge.
func stemRuns(path outline.Path, y float64) [][2]float64 {
	xs := path.CrossingsX(y)
	var runs [][2]float64
	for i := 0; i+1 < len(xs); i += 2 {
		runs = append(runs, [2]float64{xs[i], xs[i+1]})
	}
	return runs
}

// nnGap returns the width of a stem of n and the distance between the stems of
// two adjacent letters n, measured at half the x-height.
func nnGap(t *testing.T, np namedParams) (stem, gap float64) {
	t.Helper()
	path, adv := cachedOutline(np, "n")
	runs := stemRuns(path, np.params.XHeight/2)
	if len(runs) != 2 {
		t.Fatalf("%s: n has %d runs at half the x-height", np.name, len(runs))
	}
	return runs[0][1] - runs[0][0], adv - runs[1][1] + runs[0][0]
}

func TestLigatureStemsKeepTheRhythmOfN(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		np := namedParams{name: s.Name, params: s.Params}
		_, gap := nnGap(t, np)
		fPath, fAdv := cachedOutline(np, "f")
		fRuns := stemRuns(fPath, np.params.XHeight/2)
		for _, l := range ligatureList {
			path, _ := cachedOutline(np, l.name)
			runs := stemRuns(path, np.params.XHeight/2)
			if len(runs) != len(l.parts) {
				t.Errorf("%s: %s has %d stems, want %d", np.name, l.name, len(runs), len(l.parts))
				continue
			}
			for k := 0; k+1 < len(runs); k++ {
				g := runs[k+1][0] - runs[k][1]
				if math.Abs(g-gap) > 1 {
					t.Errorf("%s: %s: the stems %d and %d are %.1f apart, want %.1f as in nn", np.name, l.name, k, k+1, g, gap)
				}
				// The same two letters set without a kern; the j of a ligature
				// stands in the cell of i, and is measured as i.
				nextName := l.parts[k+1]
				if nextName == "j" {
					nextName = "i"
				}
				next, _ := cachedOutline(np, nextName)
				nr := stemRuns(next, np.params.XHeight/2)
				plain := fAdv - fRuns[0][1] + nr[0][0]
				if r := g / plain; r < ligatureGapRatio[0] || r > ligatureGapRatio[1] {
					t.Errorf("%s: %s: the gap between the stems %d and %d is %.2f of that of the letters, want %v", np.name, l.name, k, k+1, r, ligatureGapRatio)
				}
			}
		}
	}
}

func TestLigatureAdvanceIsTheSumOfTheRule(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		np := namedParams{name: s.Name, params: s.Params}
		stem, gap := nnGap(t, np)
		for _, l := range ligatureList {
			_, adv := cachedOutline(np, l.name)
			n := len(l.parts)
			_, last := cachedOutline(np, l.parts[n-1])
			if l.parts[n-1] == "j" {
				_, last = cachedOutline(np, "i")
			}
			want := last + float64(n-1)*(stem+gap)
			if math.Abs(adv-want) > 0.5 {
				t.Errorf("%s: %s has the advance %.1f, want %.1f", np.name, l.name, adv, want)
			}
			sum := 0.0
			for _, part := range l.parts {
				_, a := cachedOutline(np, part)
				sum += a
			}
			if r := adv / sum; r < ligatureNarrowing[0] || r > ligatureNarrowing[1] {
				t.Errorf("%s: %s is %.3f of the advances of its letters, want %v", np.name, l.name, r, ligatureNarrowing)
			}
		}
	}
}

// The ball of an f before i and j takes the place of the dot: no contour stands apart
// above the x-height, and the ink covers the place of the dot of i over the
// last stem.  The hook of an f before f and l runs level into the top of the
// next letter: midway between the stems the ink reaches the top of f and is
// as thick as the thin part of o, within a unit, and an l stands as high as the f.  Where two f
// are joined by the bar, the white under the bridge is a counter.
func TestLigatureHooksJoin(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		np := namedParams{name: s.Name, params: s.Params}
		p := np.params
		// The top of f, the highest ink of its rounded outline.
		fPath := cachedRounded(np, "f").Path
		fTop := math.Inf(-1)
		x0, _, x1, _ := fPath.BBox()
		for x := x0; x <= x1; x++ {
			if ys := fPath.CrossingsY(x); len(ys) > 0 {
				fTop = math.Max(fTop, ys[len(ys)-1])
			}
		}
		for _, l := range ligatureList {
			path := cachedRounded(np, l.name).Path
			n := len(l.parts)
			counters := 0
			for _, ct := range path {
				_, y0, _, _ := outline.Path{ct}.BBox()
				if y0 > p.XHeight {
					t.Errorf("%s: %s has a contour of its own above the x-height", np.name, l.name)
				}
				if ct.SignedArea() < 0 {
					counters++
				}
			}
			wantCounters := 0
			for k := 0; k+1 < n; k++ {
				if l.parts[k+1] == "f" {
					wantCounters++
				}
			}
			if counters != wantCounters {
				t.Errorf("%s: %s has %d counters, want %d", np.name, l.name, counters, wantCounters)
			}
			last := l.parts[n-1]
			if last == "i" || last == "j" {
				dot := outline.Point{X: design.LigatureCell(p, n-1) + p.Stem/2 + design.DotOffset(p), Y: p.DotCentre}
				if !inkAt(path, dot) {
					t.Errorf("%s: %s has no ink where the dot of i stands, at %v", np.name, l.name, dot)
				}
			}
			for k := 0; k+1 < n; k++ {
				if l.parts[k+1] != "f" && l.parts[k+1] != "l" {
					continue
				}
				x := design.LigatureCell(p, k) + p.Stem + (p.SideStemLeft+p.SideStemRight)/2
				ys := path.CrossingsY(x)
				if len(ys) < 2 {
					t.Errorf("%s: %s has no bridge between the stems %d and %d", np.name, l.name, k, k+1)
					continue
				}
				top, under := ys[len(ys)-1], ys[len(ys)-2]
				if math.Abs(top-fTop) > 1 {
					t.Errorf("%s: %s: the bridge between the stems %d and %d reaches %.1f, want the top of f, %.1f", np.name, l.name, k, k+1, top, fTop)
				}
				if top-under < p.ThinLower-1 {
					t.Errorf("%s: %s: the bridge between the stems %d and %d is %.1f thick, want at least %v", np.name, l.name, k, k+1, top-under, p.ThinLower-1)
				}
			}
			if last == "l" {
				xl := design.LigatureCell(p, n-1) + design.AscenderInset(p) + p.AscenderStem/2
				ys := path.CrossingsY(xl)
				if top := ys[len(ys)-1]; math.Abs(top-fTop) > 1 {
					t.Errorf("%s: %s: the l stands %.1f high, want the top of f, %.1f", np.name, l.name, top, fTop)
				}
			}
		}
	}
}

// The carets stand between the stems, midway between the cells.
func TestLigatureCaretsStandBetweenTheStems(t *testing.T) {
	t.Parallel()
	np := namedParams{name: "regular", params: design.Regular()}
	for _, l := range ligatureList {
		path, _ := cachedOutline(np, l.name)
		runs := stemRuns(path, np.params.XHeight/2)
		cs := LigatureCarets(np.params, l.parts)
		if len(cs) != len(l.parts)-1 || len(runs) != len(l.parts) {
			t.Fatalf("%s: %d carets and %d stems", l.name, len(cs), len(runs))
		}
		for k, c := range cs {
			if mid := (runs[k][1] + runs[k+1][0]) / 2; math.Abs(c-mid) > 1 {
				t.Errorf("%s: caret %d at %.1f, want %.1f, midway between the stems", l.name, k, c, mid)
			}
		}
	}
}
