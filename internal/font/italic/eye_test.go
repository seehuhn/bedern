// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
	"seehuhn.de/go/bedern/internal/stroke"
)

// eyeLoop returns a stroke of the italic pen which runs right and down, turns
// round at the right end through 180 degrees where it travels vertically and
// the pen is at its thickest, and runs back left over its own beginning: a
// long, low loop whose white eye is eye units high.  The three nodes of the
// turn are marked Hairpin and the pen is lowered there, to eyeW.
func eyeLoop(p Params, eye float64) stroke.Stroke {
	const eyeW = 16
	s := eyeW + eye
	nd := func(x, y, dir, w float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: w, DW: 0}
	}
	hp := func(n stroke.Node) stroke.Node { n.Hairpin = true; return n }
	return stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			nd(-100, s+30, -13, auto),
			hp(nd(115-s/2, 0, 0, eyeW)),
			hp(nd(115, s/2, 90, eyeW)),
			hp(nd(115-s/2, s, 180, eyeW)),
			nd(-100, s-30, 193, auto),
		},
		Start: stroke.End{Kind: stroke.Cut},
		End:   stroke.End{Kind: stroke.Cut},
	}
}

// TestLoopWithOpenEye asks that a single stroke which crosses itself and
// turns at a hairpin builds without a warning and without a finding of the
// smoothness check, and that the eye of the loop stays open: white, about
// as high as asked, and long.
func TestLoopWithOpenEye(t *testing.T) {
	p := Default()
	for _, eye := range []float64{10, 14, 20} {
		g, err := buildLetter(p, "eyeLoop", 'x', letter{strokes: []stroke.Stroke{eyeLoop(p, eye)}, advance: 300})
		if err != nil {
			t.Errorf("eye %.0f: %v", eye, err)
			continue
		}
		for _, w := range g.Warnings {
			t.Errorf("eye %.0f: warning: %s", eye, w)
		}
		for _, f := range g.Findings {
			t.Errorf("eye %.0f: %s at %.0f,%.0f: %s", eye, f.Kind, f.X, f.Y, f.Detail)
		}
		// The eye is the white between two runs of ink in a vertical line.
		var high, from, to float64
		to = math.Inf(-1)
		from = math.Inf(1)
		for x := -100.0; x < 200; x += 2 {
			cs := crossings(g.Path, pt(x, -300), pt(0, 1))
			if len(cs) != 4 {
				continue
			}
			if gap := cs[2] - cs[1]; gap > 1 {
				high = math.Max(high, gap)
				from, to = math.Min(from, x), math.Max(to, x)
			}
		}
		if high < eye-2 || high > eye+1 {
			t.Errorf("eye %.0f: the eye is %.1f high", eye, high)
		}
		if to-from < 50 {
			t.Errorf("eye %.0f: the eye is only %.0f long", eye, to-from)
		}
	}
}

// TestZetaXiLoops asks that ζ has one loop and ξ two, each with an open eye:
// the outline of ζ has one contour inside the outer one and that of ξ two,
// each of them high and long.
func TestZetaXiLoops(t *testing.T) {
	for name, loops := range map[string]int{"zeta": 1, "xi": 2} {
		g := greek(t, name)
		if len(g.Path) != loops+1 {
			t.Errorf("%s: %d contours, want %d", name, len(g.Path), loops+1)
			continue
		}
		outer, size := 0, 0.0
		for i, c := range g.Path {
			x0, y0, x1, y1 := outline.Path{c}.BBox()
			if a := (x1 - x0) * (y1 - y0); a > size {
				outer, size = i, a
			}
		}
		for i, c := range g.Path {
			if i == outer {
				continue
			}
			// The eye is the white between the two crossings of the
			// contour with a vertical line.
			p := outline.Path{c}
			x0, _, x1, _ := p.BBox()
			var high float64
			from, to := math.Inf(1), math.Inf(-1)
			for x := x0; x <= x1; x++ {
				cs := crossings(p, pt(x, -1000), pt(0, 1))
				if len(cs) == 2 {
					high = math.Max(high, cs[1]-cs[0])
					from, to = math.Min(from, x), math.Max(to, x)
				}
			}
			if high < 20 || to-from < 120 {
				t.Errorf("%s: an eye is %.0f high and %.0f long", name, high, to-from)
			}
		}
	}
}
