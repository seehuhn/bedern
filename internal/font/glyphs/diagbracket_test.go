// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of this file measure the outlines of the letters, in every parameter
// set, for the brackets of the serifs on diagonals and on stems, the crotches of
// M and N and the brackets of the thin stems.

// A bracketRun is a run of consecutive curve segments of a contour between two
// straight segments, with the directions of the straight segments on both sides.
type bracketRun struct {
	segs       []cubic
	before     outline.Point // the direction of the line before the run
	after      outline.Point // the direction of the line after it
	start, end outline.Point
}

// bracketRuns returns the runs of curves of the contour ct.
func bracketRuns(ct outline.Contour) []bracketRun {
	n := len(ct.Segs)
	nodes := contourNodes(ct)
	isCurve := func(i int) bool { return ct.Segs[(i%n+n)%n].IsCurve }
	dir := func(i int) outline.Point {
		i = (i%n + n) % n
		return unit(nodes[i], ct.Segs[i].P3)
	}
	var runs []bracketRun
	for i := 0; i < n; i++ {
		if !isCurve(i) || isCurve(i-1) {
			continue
		}
		r := bracketRun{before: dir(i - 1), start: nodes[i]}
		j := i
		for ; isCurve(j); j++ {
			s := ct.Segs[j%n]
			r.segs = append(r.segs, cubic{nodes[j%n], s.P1, s.P2, s.P3})
		}
		r.end = r.segs[len(r.segs)-1].P3
		r.after = dir(j)
		runs = append(runs, r)
	}
	return runs
}

func (r bracketRun) startTangent() outline.Point {
	c := r.segs[0]
	if c.P1 == c.P0 {
		return unit(c.P0, c.P2)
	}
	return unit(c.P0, c.P1)
}

func (r bracketRun) endTangent() outline.Point {
	c := r.segs[len(r.segs)-1]
	if c.P2 == c.P3 {
		return unit(c.P1, c.P3)
	}
	return unit(c.P2, c.P3)
}

// concave reports whether the run turns one way only, by the sign of the turn
// from its start to its end.
func (r bracketRun) concave() bool {
	var heads []float64
	for _, c := range r.segs {
		pc := curveSeg(c)
		for k := 0; k <= 20; k++ {
			tg := pc.tangent(float64(k) / 20)
			heads = append(heads, math.Atan2(tg.Y, tg.X))
		}
	}
	total := turn(heads[0], heads[len(heads)-1])
	for k := 1; k < len(heads); k++ {
		if d := turn(heads[k-1], heads[k]); d*total < -1e-9 {
			return false
		}
	}
	return true
}

func parallel(a, b outline.Point) bool {
	return math.Abs(a.X*b.Y-a.Y*b.X) < 1e-6 && a.X*b.X+a.Y*b.Y > 0
}

// bracketSeen is the distance from the straight edge beyond which a bracket on a
// diagonal is measured: its run ends at the last point of the bracket which lies
// this far from the line of the edge.
const bracketSeen = 0.5

// seenRun returns the run of the bracket r from the face of the serif at the height
// face to the last point of it which lies bracketSeen or more from the line of the
// edge through q in the direction edge.
func (r bracketRun) seenRun(face float64, q, edge outline.Point) float64 {
	var run float64
	for _, c := range r.segs {
		pc := curveSeg(c)
		for k := 0; k <= 200; k++ {
			a := pc.at(float64(k) / 200)
			if math.Abs((a.X-q.X)*edge.Y-(a.Y-q.Y)*edge.X) >= bracketSeen {
				run = math.Max(run, math.Abs(a.Y-face))
			}
		}
	}
	return run
}

// diagBracketBand is the band in which the run of a bracket on a diagonal, along
// the edge from the face of the serif, lies as a multiple of the run of the bracket
// on a stem of the same case: as generous on the obtuse side, where the corner is
// open, as about twice that of the stem, and on the acute side at least about as
// generous as that of the stem.
var diagBracketBand = [2]float64{0.9, 2.2}

// The brackets of the serifs on the diagonals are tangent to the face of the serif
// and to the edge, concave, and have a run (seenRun) within diagBracketBand of the
// bracket of the stem, and on every diagonal letter the run on the obtuse side is
// longer than that on the stem.
func TestDiagonalBracketsAreThoseOfTheStem(t *testing.T) {
	t.Parallel()
	letters := []struct {
		name string
		caps bool
	}{
		{"A", true}, {"V", true}, {"W", true}, {"X", true}, {"Y", true}, {"K", true}, {"R", true},
		{"v", false}, {"w", false}, {"x", false}, {"y", false}, {"k", false},
	}
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		for _, l := range letters {
			thick, height, top := p.SerifThick, p.BracketHeight, p.XHeight
			if l.caps {
				thick, height, top = p.Caps.SerifThick, p.Caps.BracketHeight, p.CapHeight
			}
			stemRun := height - thick
			faces := []float64{top - thick, thick}
			path, _ := cachedOutline(np, l.name)
			var found int
			var longest float64
			for _, ct := range path {
				for _, r := range bracketRuns(ct) {
					// A bracket on a diagonal leaves a horizontal face at one
					// end and meets a straight edge which is neither horizontal
					// nor vertical at the other.
					face, edge, edgeEnd, faceEnd := r.before, r.after, r.end, r.start
					tFace, tEdge := r.startTangent(), r.endTangent()
					if math.Abs(r.before.Y) > 1e-9 {
						face, edge, edgeEnd, faceEnd = r.after, r.before, r.start, r.end
						tFace, tEdge = r.endTangent(), r.startTangent()
					}
					onFace := false
					for _, y := range faces {
						onFace = onFace || math.Abs(faceEnd.Y-y) < 1e-6
					}
					if math.Abs(face.Y) > 1e-9 || !onFace || math.Abs(edge.X) < 1e-9 || math.Abs(edge.Y) < 1e-9 {
						continue
					}
					found++
					if !parallel(tFace, face) {
						t.Errorf("%s: the bracket from %v leaves the face of the serif in the direction %v, not along it", l.name, faceEnd, tFace)
					}
					if !parallel(tEdge, edge) {
						t.Errorf("%s: the bracket from %v meets the edge at %v in the direction %v, not along it (%v)", l.name, faceEnd, edgeEnd, tEdge, edge)
					}
					if !r.concave() {
						t.Errorf("%s: the bracket from %v to %v is not concave", l.name, r.start, r.end)
					}
					run := r.seenRun(faceEnd.Y, edgeEnd, edge)
					if f := run / stemRun; f < diagBracketBand[0] || f > diagBracketBand[1] {
						t.Errorf("%s: the bracket from %v runs %.1f from the face, %.2f of the %.1f of the stem bracket, want %v to %v",
							l.name, faceEnd, run, f, stemRun, diagBracketBand[0], diagBracketBand[1])
					}
					longest = math.Max(longest, run)
				}
			}
			// R has one diagonal, the leg, with one foot serif.
			want := 4
			if l.name == "R" {
				want = 2
			}
			if found < want {
				t.Errorf("%s: %d brackets on diagonals found, want at least %d", l.name, found, want)
			}
			if longest <= stemRun {
				t.Errorf("%s: the longest bracket on a diagonal runs %.1f, not more than the %.1f of the stem bracket", l.name, longest, stemRun)
			}
		}
	})
}

// crotchMinInk is the least depth of ink over the point of a white wedge of M and N,
// along the line which halves the wedge.
var crotchMinInk = map[string]float64{"M": 55, "N": 40}

// inkAt reports whether the point q lies in the ink of the path, by the crossings
// of the horizontal line through it on its left.
func inkAt(path outline.Path, q outline.Point) bool {
	var n int
	for _, x := range path.CrossingsX(q.Y) {
		if x < q.X {
			n++
		}
	}
	return n%2 == 1
}

// The crotches of M and N are points: each is a node of the outline between two
// straight segments along its edges, and the ink stands at least crotchMinInk deep
// over it.
func TestCapMNCrotchesArePoints(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		for name, f := range map[string]func(design.Params) []sharpCrotch{"M": capMCrotches, "N": capNCrotches} {
			path, _ := cachedOutline(np, name)
			ct := path[0]
			n := len(ct.Segs)
			nodes := contourNodes(ct)
			for _, k := range f(np.params) {
				i := -1
				for j := 0; j < n; j++ {
					if math.Hypot(nodes[j].X-k.Tip.X, nodes[j].Y-k.Tip.Y) < 1e-6 {
						i = j
					}
				}
				if i < 0 {
					t.Errorf("%s, %s: no node of the outline at the tip %v", name, k.Name, k.Tip)
					continue
				}
				in, out := ct.Segs[(i+n-1)%n], ct.Segs[i]
				if in.IsCurve || out.IsCurve {
					t.Errorf("%s, %s: the tip %v is not between two straight segments", name, k.Name, k.Tip)
					continue
				}
				a, b := unit(k.Tip, nodes[(i+n-1)%n]), unit(k.Tip, out.P3)
				if !(parallel(a, k.Dirs[0]) && parallel(b, k.Dirs[1]) || parallel(a, k.Dirs[1]) && parallel(b, k.Dirs[0])) {
					t.Errorf("%s, %s: the segments leave the tip in the directions %v and %v, not along its edges %v", name, k.Name, a, b, k.Dirs)
				}
				// Over the tip: against the line which halves the wedge.
				u, v := k.Dirs[0], k.Dirs[1]
				h := math.Hypot(u.X+v.X, u.Y+v.Y)
				over := outline.Point{X: -(u.X + v.X) / h, Y: -(u.Y + v.Y) / h}
				for d := 0.5; d <= crotchMinInk[name]; d += 0.5 {
					q := outline.Point{X: k.Tip.X + d*over.X, Y: k.Tip.Y + d*over.Y}
					if !inkAt(path, q) {
						t.Errorf("%s, %s: white %.1f from the tip %v, want ink at least %v deep", name, k.Name, d, k.Tip, crotchMinInk[name])
						break
					}
				}
			}
		}
	})
}

// stemBracketRuns returns the vertical runs of the brackets which meet the vertical
// edge at x of a stem, tangent to it, from the horizontal face of a serif.
func stemBracketRuns(path outline.Path, x float64) []float64 {
	var out []float64
	for _, ct := range path {
		for _, r := range bracketRuns(ct) {
			if math.Abs(r.before.Y) > 1e-9 && math.Abs(r.after.Y) > 1e-9 {
				continue
			}
			for _, e := range []struct{ q, tg outline.Point }{{r.start, r.startTangent()}, {r.end, r.endTangent()}} {
				if math.Abs(e.q.X-x) < 1e-6 && math.Abs(e.tg.X) < 1e-9 {
					out = append(out, math.Abs(r.end.Y-r.start.Y))
				}
			}
		}
	}
	return out
}

// The thin stems of M, N and U share one rule for their brackets: every bracket on
// them runs as far, and less far than the bracket of the capital stem of H.
func TestCapThinStemsShareTheBracketRule(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		h := p.Caps.BracketHeight - p.Caps.SerifThick
		stems := map[string][]float64{}
		ml := design.CapMStemLeft(p)
		stems["M"] = []float64{ml, ml + p.CapDiag.ThinStem}
		nl, nr := capNStems(p)
		stems["N"] = []float64{nl.Left, nl.rightEdge(), nr.Left, nr.rightEdge()}
		_, _, rl, rr := capUStems(p)
		stems["U"] = []float64{rl, rr}
		var ref float64
		for _, g := range []string{"M", "N", "U"} {
			path, _ := cachedOutline(np, g)
			var any bool
			for _, x := range stems[g] {
				for _, run := range stemBracketRuns(path, x) {
					any = true
					if ref == 0 {
						ref = run
					}
					if math.Abs(run-ref) > 1e-6 {
						t.Errorf("%s: a bracket on the thin stem at %.1f runs %.2f, another thin stem %.2f", g, x, run, ref)
					}
				}
			}
			if !any {
				t.Errorf("%s: no bracket found on the thin stems", g)
			}
		}
		if ref >= h {
			t.Errorf("the brackets of the thin stems run %.2f, not less than the %.2f of the capital stem", ref, h)
		}
	})
}

// polygon returns the closed contour through the points.
func polygon(pts ...outline.Point) outline.Path {
	var path outline.Path
	path.MoveTo(pts[0].X, pts[0].Y)
	for _, q := range pts[1:] {
		path.LineTo(q.X, q.Y)
	}
	return path
}

// The exemption of a listed sharp crotch from the clearance check covers only the
// edges of that crotch near its tip: a second slit of the same shape, not listed,
// is still found, and a crotch listed a little beside the real one exempts nothing.
func TestSharpCrotchExemptsOnlyItsTip(t *testing.T) {
	t.Parallel()
	const floor = 20.0
	// An ink block with white wedges cut up from the baseline, each with its point
	// at the height h and the half-width w at the baseline, at x = 100 and, with
	// two, at x = 300.
	block := func(h, w float64, two bool) outline.Path {
		pts := []outline.Point{{X: 0, Y: 0}, {X: 100 - w, Y: 0}, {X: 100, Y: h}, {X: 100 + w, Y: 0}}
		if two {
			pts = append(pts, outline.Point{X: 300 - w, Y: 0}, outline.Point{X: 300, Y: h}, outline.Point{X: 300 + w, Y: 0})
		}
		pts = append(pts, outline.Point{X: 400, Y: 0}, outline.Point{X: 400, Y: 400}, outline.Point{X: 0, Y: 400})
		return polygon(pts...)
	}
	listed := func(h, w, dx float64) func(a0, a1, b0, b1 outline.Point) bool {
		tip := outline.Point{X: 100, Y: h}
		k := sharpCrotch{Name: "left", Tip: outline.Point{X: 100 + dx, Y: h},
			Dirs: [2]outline.Point{unit(tip, outline.Point{X: 100 - w}), unit(tip, outline.Point{X: 100 + w})}}
		return crotchSkip([]sharpCrotch{k}, floor)
	}
	// Wedges of 15 degrees, 150 high.
	h, w := 150.0, 150*math.Tan(7.5*math.Pi/180)
	if d, _, _ := block(h, w, false).MinDistance(clearanceRatio); d >= floor {
		t.Fatalf("the wedge is not a slit without the exemption (%.2f)", d)
	}
	if d, a, b := block(h, w, false).MinDistanceExcept(clearanceRatio, listed(h, w, 0)); d < floor {
		t.Errorf("the listed wedge fails: %.2f between %v and %v", d, a, b)
	}
	if d, a, _ := block(h, w, true).MinDistanceExcept(clearanceRatio, listed(h, w, 0)); d >= floor || a.X < 200 {
		t.Errorf("the unlisted wedge is not found: the smallest distance is %.2f at %v", d, a)
	}
	// A crotch listed 5 units beside the real tip exempts nothing.
	if d, _, _ := block(h, w, false).MinDistanceExcept(clearanceRatio, listed(h, w, 5)); d >= floor {
		t.Errorf("a crotch listed off the tip of the wedge exempts it")
	}
}
