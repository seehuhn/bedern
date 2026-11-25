// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/stroke"
)

// The double-struck capital G, with a round open on the right, by the rule of
// dscaps.go (U and J are in dsround.go, S in dss.go).  Its round is that of C
// with the bar of G and the stem of G inside its right side.

// edgeIdx returns the left (side 0) or right (side 1) edge of the run with
// the index k of the roman, widened, through its points at the heights y0 and
// y1.
func (d dsFrame) edgeIdx(y0, y1 float64, k, side int) xline {
	pick := func(r [][2]float64) [2]float64 { return r[k] }
	return d.edgeOf(y0, y1, pick, side)
}

// dsG draws G: the round of C, whose lower arm runs on up the right side to
// its extreme at the middle height, with the bar of G, a line which leaves the
// arm to the left under the top of the arm; the corner at the right is a
// corner of the bar and the arm; the bar ends level dsGSpur of the width
// from the middle.  The left side of the round is doubled by the lens, and the
// stem of G, the thick right side of the roman, by a line inside the arm.
func dsG(d dsFrame) outline.Path {
	return fairest("G", len(dsGSpurHeights), func(i int) outline.Path { return dsGOf(d, dsGSpurHeights[i]) })
}

func dsGOf(d dsFrame, spur float64) outline.Path {
	cap := d.p.CapHeight
	c, a, b := d.round()
	ex := d.x(d.romanRun(0.74*cap, lastRun)[1]) + d.roundExtra() - d.t/2
	th := math.Acos(math.Max(-1, math.Min(1, (ex-c.X)/a))) * 180 / math.Pi
	th = math.Max(th, dsCOpen)
	var nodes []stroke.Node
	for _, deg := range []float64{th, 90, 180, 270, 360} {
		q, dir := d.ellipsePoint(c, a, b, deg)
		nodes = append(nodes, stroke.Node{P: q, Dir: dir, K: stroke.Auto, W: d.t})
	}
	st := stroke.Stroke{Nodes: nodes, Start: stroke.End{Kind: stroke.Cut}, End: stroke.End{Kind: stroke.Cut}}
	r, err := st.Build()
	if err != nil {
		panic(err)
	}
	ln, err := d.lens(c, a, b)
	if err != nil {
		panic(err)
	}
	// The bar lies under the middle height, its top level with the end of the
	// arm, and ends in the middle of the arm.
	yb := c.Y - math.Round(spur*d.t)
	bar := outline.Path{box(c.X+(1-2*dsGSpur)*a, yb-d.t, c.X+a, yb)}
	// The stem of G, which the roman draws thick: a line a white inside the
	// arm, from the bar down into the bottom of the round, which doubles the
	// right side, giving four lines at a fifth of the height of G where C
	// shows three.
	xs := c.X + a - (d.d - d.t)
	ys := c.Y - b*math.Sqrt(math.Max(0, 1-(xs-c.X)*(xs-c.X)/(a*a)))
	stem := outline.Path{box(xs-d.t/2, ys-0.2*d.t, xs+d.t/2, yb-d.t/2)}
	return glyphs.Union(r.Path(), ln, bar, stem)
}

// The length of the bar of G from the middle of the round over the half width
// of the round.
const dsGSpur = 0.45

// dsGSpurHeights are how far the arm of G stands over the bar, in lines, which
// dsG tries in turn: the rounding to whole units leaves a finding at a few.
var dsGSpurHeights = []float64{0.5, 0.45, 0.55, 0.4, 0.6, 0.35, 0.65}
