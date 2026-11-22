// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// sterlingBase holds the base of the sterling measured on verticals from the
// middle of the stem at baseTop of the figure height to the end of the base:
// for each vertical the lowest run of ink, where it lies below baseTop of the
// figure height.
type sterlingBase struct {
	xs       []float64 // the verticals
	bot, top []float64 // the lowest run of ink on each
	stemMid  float64   // the middle of the stem at baseTop
	stemLeft float64   // the left edge of the stem at baseTop
}

const (
	baseTop  = 0.25 // the base lies below this fraction of the figure height
	baseStep = 2.0  // the distance of the verticals
)

func measureSterlingBase(t *testing.T, np namedParams) (outline.Path, sterlingBase) {
	t.Helper()
	path, _ := cachedOutline(np, "sterling")
	limit := baseTop * np.params.Figures.Height
	xs := path.CrossingsX(limit)
	if len(xs) != 2 {
		t.Fatalf("%s: %d crossings at %.0f, want the two edges of the stem", np.name, len(xs), limit)
	}
	var b sterlingBase
	b.stemLeft, b.stemMid = xs[0], (xs[0]+xs[1])/2
	_, _, x1, _ := path.BBox()
	for x := b.stemMid; x < x1; x += baseStep {
		ys := path.CrossingsY(x)
		if len(ys) < 2 || ys[1] > limit {
			if len(b.xs) > 0 {
				break
			}
			continue
		}
		b.xs = append(b.xs, x)
		b.bot = append(b.bot, ys[0])
		b.top = append(b.top, ys[1])
	}
	if len(b.xs) < 10 {
		t.Fatalf("%s: the base is found on %d verticals only", np.name, len(b.xs))
	}
	return path, b
}

// The sterling is one stroke: its outline is one contour of ink and one counter,
// that of the loop, which lies left of the stem, and on the vertical through
// the middle of the stem at baseTop of the figure height the ink runs without a
// gap from the bottom of the base up to sterlingJoin of the figure height.
const sterlingJoin = 0.3

func TestSterlingIsOneStroke(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		path, b := measureSterlingBase(t, np)
		var ink, counters int
		for _, ct := range path {
			if ct.SignedArea() > 0 {
				ink++
				continue
			}
			counters++
			for _, q := range ct.Flatten() {
				if q.X > b.stemLeft {
					t.Errorf("%s: the counter reaches %v, right of the stem at %.1f", np.name, q, b.stemLeft)
					break
				}
			}
		}
		if ink != 1 || counters != 1 {
			t.Errorf("%s: %d contours of ink and %d counters, want 1 and 1", np.name, ink, counters)
		}
		ys := path.CrossingsY(b.stemMid)
		if len(ys) < 2 || ys[1] < sterlingJoin*np.params.Figures.Height {
			t.Errorf("%s: on the vertical through the stem the ink from the bottom ends at %v", np.name, ys)
		}
	}
}

// The base is heavy: its greatest thickness, measured vertically, is between
// baseHeavyMin and baseHeavyMax of the stem and lies in the middle third of the
// base, from the middle of the stem to the end.
const (
	baseHeavyMin = 0.78
	baseHeavyMax = 0.95
)

func TestSterlingBaseIsHeavyInTheMiddle(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		_, b := measureSterlingBase(t, np)
		best, at := 0.0, 0.0
		for i, x := range b.xs {
			if th := b.top[i] - b.bot[i]; th > best {
				best, at = th, x
			}
		}
		stem := np.params.Sym.SterStem
		if r := best / stem; r < baseHeavyMin || r > baseHeavyMax {
			t.Errorf("%s: the base is %.1f thick at its heaviest, %.2f of the stem", np.name, best, r)
		}
		x0, x1 := b.stemMid, b.xs[len(b.xs)-1]
		if at < x0+(x1-x0)/3 || at > x0+2*(x1-x0)/3 {
			t.Errorf("%s: the base is heaviest at %.0f, outside the middle third of %.0f to %.0f", np.name, at, x0, x1)
		}
	}
}

// From its heaviest place to its end the base tapers: its thickness never grows
// by more than taperSlack from one vertical to the next and ends below
// taperEnd of the greatest; and the end lifts: the middle of the last vertical
// lies between liftMin and liftMax of the figure height above the middle of the
// base where it is heaviest.
const (
	taperSlack = 0.5
	taperEnd   = 0.6
	liftMin    = 0.07
	liftMax    = 0.16
)

func TestSterlingBaseTapersAndLifts(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		_, b := measureSterlingBase(t, np)
		k := 0
		for i := range b.xs {
			if b.top[i]-b.bot[i] > b.top[k]-b.bot[k] {
				k = i
			}
		}
		for i := k + 1; i < len(b.xs); i++ {
			if g := (b.top[i] - b.bot[i]) - (b.top[i-1] - b.bot[i-1]); g > taperSlack {
				t.Errorf("%s: the base grows by %.2f at %.0f", np.name, g, b.xs[i])
			}
		}
		n := len(b.xs) - 1
		peak, end := b.top[k]-b.bot[k], b.top[n]-b.bot[n]
		if end > taperEnd*peak {
			t.Errorf("%s: the base ends %.1f thick, %.2f of its greatest thickness", np.name, end, end/peak)
		}
		fh := np.params.Figures.Height
		lift := (b.top[n]+b.bot[n])/2 - (b.top[k]+b.bot[k])/2
		if lift < liftMin*fh || lift > liftMax*fh {
			t.Errorf("%s: the end of the base lifts by %.1f, %.3f of the figure height", np.name, lift, lift/fh)
		}
	}
}

// The counter of the loop stays open: it is at least counterWide of the
// figure height wide and counterHigh high.
const (
	counterWide = 0.09
	counterHigh = 0.11
)

func TestSterlingCounterIsOpen(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		path, _ := cachedOutline(np, "sterling")
		fh := np.params.Figures.Height
		for _, ct := range path {
			if ct.SignedArea() > 0 {
				continue
			}
			x0, y0, x1, y1 := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
			for _, q := range ct.Flatten() {
				x0, y0, x1, y1 = min(x0, q.X), min(y0, q.Y), max(x1, q.X), max(y1, q.Y)
			}
			if x1-x0 < counterWide*fh || y1-y0 < counterHigh*fh {
				t.Errorf("%s: the counter of the loop is %.0f wide and %.0f high", np.name, x1-x0, y1-y0)
			}
		}
	}
}

// The terminal of the hook is the ball of the lowercase: the outline has a
// circle of the radius BallDiameter/2, within ballRadiusSlack.
const ballRadiusSlack = 0.5

func TestSterlingBallIsTheSharedBall(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		want := np.params.BallDiameter / 2
		found := false
		for _, run := range circleRuns(cachedMerged(np, "sterling")) {
			if math.Abs(run.r-want) <= ballRadiusSlack {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no circle of the radius %.1f", np.name, want)
		}
	}
}

// The bar crosses the stem at SterBarY, SterBar thick (the two highest
// crossings of the vertical just inside its left end), and reaches between
// barLeftMin and barLeftMax of the advance left of the stem and at least
// barRightMin right of it.
const (
	barLeftMin  = 0.12
	barLeftMax  = 0.18
	barRightMin = 0.25
)

func TestSterlingBar(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		path, adv := cachedOutline(np, "sterling")
		s := np.params.Sym
		xs := path.CrossingsX(s.SterBarY)
		above := path.CrossingsX(s.SterBarY + s.SterBar)
		if len(xs) < 2 || len(above) < 2 {
			t.Fatalf("%s: %d and %d crossings at the bar and above it", np.name, len(xs), len(above))
		}
		bl, br := xs[0], xs[len(xs)-1]
		if l := (above[0] - bl) / adv; l < barLeftMin || l > barLeftMax {
			t.Errorf("%s: the bar reaches %.3f of the advance left of the stem", np.name, l)
		}
		if r := (br - above[1]) / adv; r < barRightMin {
			t.Errorf("%s: the bar reaches %.3f of the advance right of the stem", np.name, r)
		}
		ys := path.CrossingsY(bl + 2)
		if n := len(ys); n < 2 || math.Abs(ys[n-2]-(s.SterBarY-s.SterBar/2)) > 0.5 || math.Abs(ys[n-1]-(s.SterBarY+s.SterBar/2)) > 0.5 {
			t.Errorf("%s: the left end of the bar spans %v", np.name, ys)
		}
	}
}
