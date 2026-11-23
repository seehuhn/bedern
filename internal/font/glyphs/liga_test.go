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
// the ball takes the place of the dot of i and j, an f before f ends in a free
// hook, and the hook of an f before l runs into the stem of l.

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

// The tops of the ligatures.  No contour stands apart above the x-height and
// none is a counter.  The ball of an f before i and j takes the place of the
// dot: the ink covers the place of the dot of i over the last stem.  An f
// before f ends in a free hook: between the bar and the crest the ink of the
// first f keeps at least ligaHookGap from the stem and the outer edge of the
// second, its crest stands within ligaCrestDrop below the top of f, and the
// second f stands as high as f.  The hook of an f before l runs into the stem
// of l: midway between the stems it is at least as thick as the thin part of
// o, its crest stands as high as that of f, the l keeps the ascender as its
// top, and the left-hand edge of l rises at least ligaLRise above the join.
func TestLigatureTops(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		np := namedParams{name: s.Name, params: s.Params}
		p := np.params
		fTop := topOfInk(cachedRounded(np, "f").Path)
		for _, l := range ligatureList {
			path := cachedRounded(np, l.name).Path
			n := len(l.parts)
			for _, ct := range path {
				_, y0, _, _ := outline.Path{ct}.BBox()
				if y0 > p.XHeight {
					t.Errorf("%s: %s has a contour of its own above the x-height", np.name, l.name)
				}
				if ct.SignedArea() < 0 {
					t.Errorf("%s: %s has a counter", np.name, l.name)
				}
			}
			last := l.parts[n-1]
			if last == "i" || last == "j" {
				dot := outline.Point{X: design.LigatureCell(p, n-1) + p.Stem/2 + design.DotOffset(p), Y: p.DotCentre}
				if !inkAt(path, dot) {
					t.Errorf("%s: %s has no ink where the dot of i stands, at %v", np.name, l.name, dot)
				}
			}
			runs := stemRuns(path, p.XHeight/2)
			if len(runs) != n {
				continue // reported by TestLigatureStemsKeepTheRhythmOfN
			}
			for k := 0; k+1 < n; k++ {
				x0, x1 := runs[k][0], runs[k+1][0]
				crest := topBetween(path, x0, x1)
				switch l.parts[k+1] {
				case "f":
					if d := fTop - crest; d < -1 || d > ligaCrestDrop {
						t.Errorf("%s: %s: the crest of the f %d stands %.1f below the top of f, want -1 to %v", np.name, l.name, k, d, ligaCrestDrop)
					}
					if g := hookGap(path, runs[k][1], x1, p.XHeight, crest); g < ligaHookGap {
						t.Errorf("%s: %s: the hook of the f %d comes within %.1f of the next f, want at least %v", np.name, l.name, k, g, ligaHookGap)
					}
					if k+1 < n-1 || last == "f" {
						if top := topBetween(path, x1, x1+design.LigaturePitch(p)); math.Abs(top-fTop) > 1 {
							t.Errorf("%s: %s: the f %d stands %.1f high, want the top of f, %.1f", np.name, l.name, k+1, top, fTop)
						}
					}
				case "l":
					if math.Abs(crest-fTop) > 1 {
						t.Errorf("%s: %s: the crest of the f %d stands at %.1f, want the top of f, %.1f", np.name, l.name, k, crest, fTop)
					}
					xm := design.LigatureCaret(p, k)
					ys := path.CrossingsY(xm)
					if len(ys) < 2 || ys[len(ys)-1]-ys[len(ys)-2] < p.ThinLower-1 {
						t.Errorf("%s: %s: the hook of the f %d is thinner than %v midway between the stems", np.name, l.name, k, p.ThinLower-1)
					}
					xl := design.LigatureCell(p, n-1) + design.AscenderInset(p)
					ys = path.CrossingsY(xl + p.AscenderStem/2)
					if top := ys[len(ys)-1]; math.Abs(top-p.Ascender) > 0.5 {
						t.Errorf("%s: %s: the l stands %.1f high, want the ascender, %v", np.name, l.name, top, p.Ascender)
					}
					ys = path.CrossingsY(xl - 0.5)
					if rise := p.Ascender - ys[len(ys)-1]; rise < ligaLRise {
						t.Errorf("%s: %s: the l rises %.1f above the hook, want at least %v", np.name, l.name, rise, ligaLRise)
					}
				}
			}
		}
	}
}

// ligaCrestDrop is the most by which the crest of an f before f stands below
// the top of f, ligaHookGap the least white between its hook and the next f,
// and ligaLRise the least by which the left-hand edge of l rises above the
// hook which runs into it.
const (
	ligaCrestDrop = 8.0
	ligaHookGap   = 40.0
	ligaLRise     = 12.0
)

// topOfInk returns the highest ink of the path.
func topOfInk(path outline.Path) float64 {
	x0, _, x1, _ := path.BBox()
	return topBetween(path, x0, x1)
}

// topBetween returns the highest ink of the path between the abscissae x0
// and x1, scanned at every unit.
func topBetween(path outline.Path, x0, x1 float64) float64 {
	top := math.Inf(-1)
	for x := math.Ceil(x0); x <= x1; x++ {
		if ys := path.CrossingsY(x); len(ys) > 0 {
			top = math.Max(top, ys[len(ys)-1])
		}
	}
	return top
}

// hookGap returns the least horizontal white, at the heights from the top
// of the bar to the crest, between the first run of ink which reaches beyond
// x1 (the left-hand edge of the next stem) and the run before it, where the
// first run of the line begins left of xs (the right-hand edge of the first
// stem).
func hookGap(path outline.Path, xs, x1, xh, crest float64) float64 {
	gap := math.Inf(1)
	for y := math.Ceil(xh) + 1; y < crest; y++ {
		runs := stemRuns(path, y)
		for i := 1; i < len(runs); i++ {
			if runs[i][1] > x1 {
				if runs[0][0] < xs {
					gap = math.Min(gap, runs[i][0]-runs[i-1][1])
				}
				break
			}
		}
	}
	return gap
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

// The outer edge of the hook of an f before i or j, from the ball to the
// stem, is one arch whose curvature changes evenly: it has one crest, its
// curvature falls from the ball to one least value and rises from there until
// it falls to zero at the stem in the transition curve, and no stretch longer
// than dotFlatRun of the width of the hook (from the left-hand edge of the stem
// to the rightmost ink of the ball) has a curvature below dotFlatShare of the
// mean curvature of the edge.
func TestDotHookIsOneEvenArch(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		p := s.Params
		x := design.LigatureCell(p, 1) + p.Stem/2 + design.DotOffset(p)
		hook := fDotHook(p, newFHookParts(p), x)
		r := p.BallDiameter / 2
		// The outer edge: the pieces after the last one on the circle of the ball.
		first := 0
		for i, pc := range hook {
			if !pc.line && math.Abs(pc.curvature(0)*r-1) < 0.01 && math.Abs(pc.curvature(1)*r-1) < 0.01 {
				first = i + 1
			}
		}
		var ks, ys, ss []float64
		length := 0.0
		for _, pc := range hook[first:] {
			prev := pc.at(0)
			for i := 0; i <= 200; i++ {
				q := pc.at(float64(i) / 200)
				length += math.Hypot(q.X-prev.X, q.Y-prev.Y)
				prev = q
				ks = append(ks, pc.curvature(float64(i)/200))
				ys = append(ys, q.Y)
				ss = append(ss, length)
			}
		}
		// The edge turns from its direction at the ball, which points up and
		// to the left, over the crest down to the stem.
		d0 := hook[first].tangent(0)
		turn := 3*math.Pi/2 - math.Atan2(d0.Y, d0.X)
		mean := turn / length
		_, _, xr, _ := chainBBox(hook)
		width := xr - hook.end().X
		run, longest := 0.0, 0.0
		for i := 1; i < len(ks); i++ {
			if ks[i] < dotFlatShare*mean {
				run += ss[i] - ss[i-1]
				longest = math.Max(longest, run)
			} else {
				run = 0
			}
		}
		if longest > dotFlatRun*width {
			t.Errorf("%s: the outer edge of the hook of fi is flatter than %v of its mean curvature over %.1f, %.2f of the width of the hook, want at most %v", s.Name, dotFlatShare, longest, longest/width, dotFlatRun)
		}
		// A crest is where the edge, walked from the ball, stops rising and
		// falls; steps of less than a thousandth of a unit do not count.
		crests, turns, rising := 0, 0, false
		for i := 1; i < len(ys); i++ {
			switch dy := ys[i] - ys[i-1]; {
			case dy > 1e-3:
				rising = true
			case dy < -1e-3 && rising:
				crests++
				rising = false
			}
		}
		for i := 1; i+1 < len(ks); i++ {
			if (ks[i]-ks[i-1])*(ks[i+1]-ks[i]) < -1e-14 {
				turns++
			}
		}
		if crests != 1 {
			t.Errorf("%s: the outer edge of the hook of fi has %d crests, want 1", s.Name, crests)
		}
		if turns > 2 {
			t.Errorf("%s: the curvature of the outer edge of the hook of fi turns %d times, want at most 2", s.Name, turns)
		}
	}
}

// dotFlatShare is the share of the mean curvature below which the outer edge
// of the hook of an f before i or j counts as flat, and dotFlatRun the longest
// flat stretch, as a share of the width of the hook.
const (
	dotFlatShare = 0.5
	dotFlatRun   = 0.1
)
