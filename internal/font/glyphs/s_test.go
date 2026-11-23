// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

func sGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, adv, ok := outlineOf("s", p)
	if !ok {
		t.Fatal("no glyph s")
	}
	if len(path) != 1 {
		t.Fatalf("s has %d contours, want 1 (it has no counter)", len(path))
	}
	return path, adv
}

// sPoints returns the outline of s as a polygon, sampled with no side
// longer than 2 units, and for each segment of the contour the index of the
// point at which it ends.
func sPoints(path outline.Path) (pts []outline.Point, end []int) {
	c := path[0]
	pts = []outline.Point{c.Start}
	cur := c.Start
	for _, s := range c.Segs {
		if s.IsCurve {
			cu := cubic{cur, s.P1, s.P2, s.P3}
			for i := 1; i <= 64; i++ {
				pts = append(pts, cu.at(float64(i)/64))
			}
		} else {
			k := int(math.Ceil(math.Hypot(s.P3.X-cur.X, s.P3.Y-cur.Y) / 2))
			for i := 1; i <= k; i++ {
				pts = append(pts, lerp(cur, s.P3, float64(i)/float64(k)))
			}
		}
		end = append(end, len(pts)-1)
		cur = s.P3
	}
	// The contour is closed: the last point is the first.
	pts = pts[:len(pts)-1]
	end[len(end)-1] = 0
	return pts, end
}

// sThickness returns the thickness of the stroke at the point i of the
// polygon: the distance to the nearest point of the other edge of the stroke,
// which is the part of the polygon from the point lo to the point hi.
func sThickness(pts []outline.Point, i, lo, hi int) float64 {
	best := math.Inf(1)
	for j := lo; j < hi; j++ {
		best = math.Min(best, segDist(pts[i], pts[j], pts[j+1]))
	}
	return best
}

// sJoints returns the on-curve points of the contour between two curves, with
// the unit directions in which the contour arrives and leaves.
type sJoint struct {
	At, In, Out outline.Point
}

func sJointsOf(path outline.Path) []sJoint {
	c := path[0]
	unit := func(a, b outline.Point) outline.Point {
		d := math.Hypot(b.X-a.X, b.Y-a.Y)
		return outline.Point{X: (b.X - a.X) / d, Y: (b.Y - a.Y) / d}
	}
	var js []sJoint
	cur := c.Start
	n := len(c.Segs)
	for k, s := range c.Segs {
		next := c.Segs[(k+1)%n]
		if s.IsCurve && next.IsCurve {
			js = append(js, sJoint{At: s.P3, In: unit(s.P2, s.P3), Out: unit(s.P3, next.P1)})
		}
		cur = s.P3
	}
	_ = cur
	return js
}

func TestSAdvance(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := sGlyph(t, p)
	near(t, "advance", adv, design.SAdvance(p))
	near(t, "advance from the parts", adv, p.SSideLeft+p.SLowerWidth+p.SSideRight)
	x0, _, x1, _ := path.BBox()
	near(t, "left side space", x0, p.SSideLeft)
	near(t, "right side space", adv-x1, p.SSideRight)
	for _, g := range []Glyph{roundedOf("s", p)} {
		{
			near(t, "rounded advance", g.Advance, math.Round(design.SAdvance(p)))
			if g.Rune != 's' {
				t.Errorf("rune %q, want 's'", g.Rune)
			}
		}
	}
	e, _ := lookup("s")
	if e.left != sideOpen || e.right != sideOpen {
		t.Errorf("sides of s are %v and %v, want open and open", e.left, e.right)
	}
}

// The overshoot at the top and at the bottom is that of the o, and the arcs
// have the thin value on the verticals through the two extremes.
func TestSOvershootAndThinArcs(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := sGlyph(t, p)
	_, y0, _, y1 := path.BBox()
	near(t, "top", y1, p.XHeight+p.OvershootLower)
	near(t, "bottom", y0, -p.OvershootLower)

	var xTop, xBottom float64
	for _, s := range path[0].Segs {
		if s.IsCurve && math.Abs(s.P3.Y-y1) < 1e-9 {
			xTop = s.P3.X
		}
		if s.IsCurve && math.Abs(s.P3.Y-y0) < 1e-9 {
			xBottom = s.P3.X
		}
	}
	ys := path.CrossingsY(xTop + 0.01)
	if len(ys) < 4 {
		t.Fatalf("%d crossings on the vertical through the top, want at least 4", len(ys))
	}
	near(t, "thin arc at the top", ys[len(ys)-1]-ys[len(ys)-2], p.SUpperArcThick)
	ys = path.CrossingsY(xBottom + 0.01)
	near(t, "thin arc at the bottom", ys[1]-ys[0], p.SArcThick)
	if p.SUpperArcThick < p.SArcThick || p.SUpperArcThick > p.SArcThick+6 || p.SArcThick < p.ThinLower-4 || p.SArcThick > p.Hair {
		t.Errorf("SArcThick = %v and SUpperArcThick = %v: the lower arc must lie within the thin values %v to %v of the o and the upper arc 0 to 6 above it", p.SArcThick, p.SUpperArcThick, p.ThinLower-4, p.Hair)
	}
}

// The ink above and below the middle height has the widths of the two halves,
// the lower half is wider by the ratio, and its centre lies to the right.
func TestSHalves(t *testing.T) {
	t.Parallel()
	for _, p := range []design.Params{design.Regular()} {
		path, _ := sGlyph(t, p)
		pts, _ := sPoints(path)
		mid := p.XHeight / 2
		lo := [2]float64{math.Inf(1), math.Inf(-1)}
		hi := lo
		for _, q := range pts {
			r := &hi
			if q.Y < mid {
				r = &lo
			}
			r[0], r[1] = math.Min(r[0], q.X), math.Max(r[1], q.X)
		}
		near(t, "width of the upper half", hi[1]-hi[0], p.SUpperWidth)
		near(t, "width of the lower half", lo[1]-lo[0], p.SLowerWidth)
		near(t, "ratio of the halves", (lo[1]-lo[0])/(hi[1]-hi[0]), p.SLowerWidth/p.SUpperWidth)
		near(t, "shift of the centres", (lo[0]+lo[1])/2-(hi[0]+hi[1])/2, p.SUpperShift)
		if r := p.SLowerWidth / p.SUpperWidth; r < 1.078 || r > 1.137 {
			t.Errorf("ratio %.3f lies outside the range 1.078 to 1.137", r)
		}
	}
}

// The spine: two on-curve points, one on each edge, a thickness apart at right
// angles to the direction of the edges there, which is the angle of the spine;
// the centre lies above the middle of the x-height in the ratio of the halves.
func TestSSpine(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := sGlyph(t, p)
	top, bottom := p.XHeight+p.OvershootLower, -p.OvershootLower
	r := p.SUpperWidth / p.SLowerWidth
	wantY := (top + r*bottom) / (1 + r)

	var spine []sJoint
	for _, j := range sJointsOf(path) {
		// The joints at the centre of the spine are the only ones whose
		// tangent is neither horizontal nor vertical.
		if a := math.Abs(j.In.Y); a > 0.2 && a < 0.87 && j.At.Y > 0.25*p.XHeight && j.At.Y < 0.75*p.XHeight {
			spine = append(spine, j)
		}
	}
	if len(spine) != 2 {
		t.Fatalf("%d joints at the spine, want 2", len(spine))
	}
	a, b := spine[0], spine[1] // the edge below, then the edge above
	if a.At.Y > b.At.Y {
		a, b = b, a
	}
	centre := outline.Point{X: (a.At.X + b.At.X) / 2, Y: (a.At.Y + b.At.Y) / 2}
	near(t, "height of the centre", centre.Y, wantY)
	if centre.Y <= p.XHeight/2 {
		t.Errorf("the centre of the spine is at %.1f, not above the middle %v", centre.Y, p.XHeight/2)
	}
	near(t, "thickness of the spine", math.Hypot(b.At.X-a.At.X, b.At.Y-a.At.Y), p.SSpineThick)
	for _, j := range spine {
		// Both edges run at the angle of the spine, falling to the right when
		// followed downwards; In and Out are parallel and equal up to sign.
		angle := math.Atan2(-math.Abs(j.Out.Y), math.Abs(j.Out.X)) * 180 / math.Pi
		near(t, "angle of the spine", -angle, p.SSpineAngle)
		if math.Abs(j.In.X*j.Out.Y-j.In.Y*j.Out.X) > 1e-6 {
			t.Errorf("the edge turns a corner at the spine: %v then %v", j.In, j.Out)
		}
	}
	// The line between the two points is at right angles to the spine.
	d := outline.Point{X: b.At.X - a.At.X, Y: b.At.Y - a.At.Y}
	dot := (d.X*a.Out.X + d.Y*a.Out.Y) / math.Hypot(d.X, d.Y)
	near(t, "cosine of the angle between the thickness and the spine", dot, 0)
}

// An sBeak holds the measurements of one beak of s or S in the frame in which
// it hangs at the top on the right (the lower beak turned through half a
// turn): the path in that frame, the abscissa of the cut, the top of the
// letter, the top extreme of the arc and the serif.
type sBeakView struct {
	Path        outline.Path
	Cut, Top, X float64 // the cut, the top of the letter, the abscissa of the top extreme
	Serif       hangingSerif
}

// sBeakViews returns the upper and the lower beak of a letter drawn by the
// frame f (s or S) whose outline is path.
func sBeakViews(t *testing.T, path outline.Path, f sFrame) [2]sBeakView {
	t.Helper()
	turn := func(q outline.Point) outline.Point { return outline.Point{X: -q.X, Y: -q.Y} }
	var v [2]sBeakView
	for i, w := range []struct {
		path outline.Path
		cut  float64
		top  outline.Point // the top extreme in the frame of the view
	}{
		{path, f.C.X + f.Upper.Lx, outline.Point{X: f.C.X - f.UpperE.Outer[1].P3.X, Y: f.C.Y - f.UpperE.Outer[1].P3.Y}},
		{mapPath(path, turn), -(f.C.X - f.Lower.Lx), turn(outline.Point{X: f.C.X + f.LowerE.Outer[1].P3.X, Y: f.C.Y + f.LowerE.Outer[1].P3.Y})},
	} {
		v[i] = sBeakView{Path: w.path, Cut: w.cut, Top: w.top.Y, X: w.top.X}
		v[i].Serif = measureHangingSerif(t, w.path, w.cut)
	}
	return v
}

var sBeakNames = [2]string{"upper", "lower"}

// The beaks of s are those of S at the size of the lowercase.  With k the ratio
// of the x-height to the cap height, the serif of each beak of s is 1.0 to 1.8
// times k times as wide as that of S halfway along its face, its end lies 0.85
// to 1.2 times k times as far from the top (bottom) of the letter, and it is
// bracketed: the ink widens to twice the serif not at once and within four times
// the height, scaled, in which the bracket of S does so.
func TestSBeaksAreThoseOfS(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			k := p.XHeight / p.CapHeight
			lc := sBeakViews(t, capOutline(np, "s"), newSFrame(p))
			uc := sBeakViews(t, capOutline(np, "S"), newCapSFrameRamp(p, beakRamp))
			for i := range lc {
				a, b := lc[i].Serif, uc[i].Serif
				if r := a.Thick / (k * b.Thick); r < 1.0 || r > 1.8 {
					t.Errorf("the %s serif of s is %.1f wide, %.2f times that of S (%.1f) scaled, want 1.0 to 1.8", sBeakNames[i], a.Thick, r, b.Thick)
				}
				da, db := lc[i].Top-a.End, uc[i].Top-b.End
				if r := da / (k * db); r < 0.85 || r > 1.2 {
					t.Errorf("the %s serif of s ends %.1f from the edge of the letter, %.2f times that of S (%.1f) scaled, want 0.85 to 1.2", sBeakNames[i], da, r, db)
				}
				if a.Bracket < 3 || a.Bracket > 4*k*b.Bracket {
					t.Errorf("the bracket of the %s serif of s is %.1f high, that of S %.1f: want 3 to %.1f", sBeakNames[i], a.Bracket, b.Bracket, 4*k*b.Bracket)
				}
				if np.name == "regular" {
					t.Logf("%s: thick %.1f drop %.1f face %.1f bracket %.1f (S %.1f %.1f %.1f %.1f, k %.3f)", sBeakNames[i], a.Thick, da, a.Face, a.Bracket, b.Thick, db, b.Face, b.Bracket, k)
				}
			}
		})
	}
}

// The arcs of s stay thin from the top (bottom) extreme to the beak: at every
// point of the outer edge from the extreme to the bracket, which starts where
// the serif widens to twice its width, the arc is at right angles at least as
// thick as the lower arc at its extreme (SArcThick) less two units, and at most
// 1.5 times as thick as at its own extreme.
func TestSBeakArcsAreThin(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			arcs := [2]float64{p.SUpperArcThick, p.SArcThick}
			for i, v := range sBeakViews(t, capOutline(np, "s"), newSFrame(p)) {
				// The bracket starts where the arc is twice the serif wide: the
				// left end of the run which reaches the cut at the height of the
				// end of the bracket.
				y := v.Serif.End + v.Serif.Face + v.Serif.Bracket
				r := runs(v.Path, y)
				stop := r[len(r)-1][0]
				// The outer and the inner edge of the arc on the verticals.
				memo := map[float64][2]outline.Point{}
				arc := func(x float64) (outline.Point, outline.Point) {
					if e, ok := memo[x]; ok {
						return e[0], e[1]
					}
					vr := vruns(v.Path, x)
					top := vr[len(vr)-1]
					e := [2]outline.Point{{X: x, Y: top[1]}, {X: x, Y: top[0]}}
					memo[x] = e
					return e[0], e[1]
				}
				for x := v.X; x < stop; x += 2 {
					out, _ := arc(x)
					th := math.Inf(1)
					for x2 := x - 100; x2 <= math.Min(x+100, stop); x2 += 0.5 {
						_, in := arc(x2)
						th = math.Min(th, math.Hypot(in.X-out.X, in.Y-out.Y))
					}
					if th < p.SArcThick-2 || th > 1.5*arcs[i] {
						t.Errorf("the %s arc is %.1f thick %.0f units from the extreme, want %.1f to %.1f", sBeakNames[i], th, x-v.X, p.SArcThick-2, 1.5*arcs[i])
						break
					}
				}
			}
		})
	}
}

// The apertures of s stay open: below the middle of each serif the white down
// to the next ink is at least 0.12 of the x-height (54 units in the regular
// set), one and a half times the hairline.
func TestSApertures(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			for i, v := range sBeakViews(t, capOutline(np, "s"), newSFrame(p)) {
				x := v.Cut - v.Serif.Thick/2
				gap := math.Inf(1)
				for _, r := range vruns(v.Path, x) {
					if r[1] < v.Serif.End-1 {
						gap = v.Serif.End - r[1]
					}
				}
				if !(gap >= 0.12*p.XHeight) {
					t.Errorf("the aperture below the %s serif is %.1f, want at least %.1f", sBeakNames[i], gap, 0.12*p.XHeight)
				}
				if np.name == "regular" {
					t.Logf("%s aperture %.1f", sBeakNames[i], gap)
				}
			}
		})
	}
}

// curvatureSigns returns the sign of the turning of the contour at many
// points of the run of cubic segments from ends[from] on, and the largest jump
// of direction at the joints between the segments.
func sRuns(path outline.Path) [][]int {
	c := path[0]
	n := len(c.Segs)
	var runs [][]int
	// Start after a line segment and collect consecutive curves.
	first := -1
	for k := 0; k < n; k++ {
		if !c.Segs[k].IsCurve {
			first = k
			break
		}
	}
	var cur []int
	for m := 1; m <= n; m++ {
		k := (first + m) % n
		if c.Segs[k].IsCurve {
			cur = append(cur, k)
			continue
		}
		if len(cur) > 0 {
			runs = append(runs, cur)
			cur = nil
		}
	}
	if len(cur) > 0 {
		runs = append(runs, cur)
	}
	return runs
}

func sCubicAt(path outline.Path, k int) cubic {
	c := path[0]
	n := len(c.Segs)
	from := c.Start
	if k > 0 {
		from = c.Segs[k-1].P3
	}
	s := c.Segs[k]
	_ = n
	return cubic{from, s.P1, s.P2, s.P3}
}

// Smoothness: along each of the two long edges the direction of the outline
// turns one way through the first bowl, has exactly one inflection, at the
// spine, and turns the other way through the second bowl, and has no jump at
// the joints of the segments.
func TestSSmoothEdges(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		path, _, _ := outlineOf("s", np.params)
		path0 := path
		runs := sRuns(path0)
		if len(runs) != 2 {
			t.Fatalf("%s: %d runs of curves, want 2 (the two long edges)", np.name, len(runs))
		}
		for _, run := range runs {
			if len(run) != 8 {
				t.Fatalf("%s: a long edge has %d segments, want 8 (the hollow at the beak, the round to it, the two quarters and the four of the inner edge: two quarters, the bracket and the curve into the face)", np.name, len(run))
			}
			var signs []float64
			var turned [2]float64
			for i, k := range run {
				c := sCubicAt(path0, k)
				if i > 0 {
					prev := sCubicAt(path0, run[i-1])
					a, b := prev.tangent(1), c.tangent(0)
					if jump := math.Abs(math.Atan2(a.X*b.Y-a.Y*b.X, a.X*b.X+a.Y*b.Y)); jump > 0.5*math.Pi/180 {
						t.Errorf("%s: jump of direction %.2f degrees at the joint of segments %d and %d", np.name, jump*180/math.Pi, run[i-1], k)
					}
				}
				const steps = 200
				prevT := c.tangent(0)
				for j := 1; j <= steps; j++ {
					cur := c.tangent(float64(j) / steps)
					turn := math.Atan2(prevT.X*cur.Y-prevT.Y*cur.X, prevT.X*cur.X+prevT.Y*cur.Y)
					if math.Abs(turn) > 1e-9 {
						signs = append(signs, turn)
					}
					if i < 4 {
						turned[0] += turn
					} else {
						turned[1] += turn
					}
					prevT = cur
				}
			}
			changes := 0
			for i := 1; i < len(signs); i++ {
				if (signs[i] > 0) != (signs[i-1] > 0) {
					changes++
				}
			}
			if changes != 2 {
				t.Errorf("%s: the direction of a long edge reverses its turning %d times, want twice (in the hollow at the beak and at the spine)", np.name, changes)
			}
			// The first four segments turn one way and the last four the
			// other, by more than 90 degrees each.
			if turned[0]*turned[1] >= 0 || math.Abs(turned[0]) < math.Pi/2 || math.Abs(turned[1]) < math.Pi/2 {
				t.Errorf("%s: a long edge turns %.0f and %.0f degrees through its two halves", np.name, turned[0]*180/math.Pi, turned[1]*180/math.Pi)
			}
		}
	}
}

// The thickness at right angles, followed along the edge which starts at the
// top, rises monotonically from the top extreme to the centre of the spine and
// falls monotonically to the bottom extreme.
func TestSThicknessProfile(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := sGlyph(t, p)
	pts, end := sPoints(path)
	// Segments 18 and 19 lead from the top extreme to the spine, 0 and 1 from
	// the spine to the bottom extreme (the edge wraps round the start); the
	// other edge runs from the end of the lower cut (5) to the end of the upper
	// arc (12).
	if len(path[0].Segs) != 20 {
		t.Fatalf("%d segments, want 20", len(path[0].Segs))
	}
	n := len(pts)
	var th []float64
	startIdx, spineIdx := end[17], end[19]
	for i := startIdx; ; i = (i + 1) % n {
		th = append(th, sThickness(pts, i, end[5], end[12]))
		if i == end[1] {
			break
		}
	}
	spineAt := (spineIdx - startIdx + n) % n
	if th[0] > p.SUpperArcThick+0.5 {
		t.Errorf("thickness at the top extreme %.1f", th[0])
	}
	near(t, "thickness at the bottom extreme", th[len(th)-1], p.SArcThick)
	near(t, "thickness at the centre of the spine", th[spineAt], p.SSpineThick)
	for i := 1; i <= spineAt; i++ {
		if th[i] < th[i-1]-0.5 {
			t.Errorf("the stroke thins from %.2f to %.2f before the spine, at %d of %d", th[i-1], th[i], i, spineAt)
			break
		}
	}
	for i := spineAt + 1; i < len(th); i++ {
		if th[i] > th[i-1]+0.5 {
			t.Errorf("the stroke thickens from %.2f to %.2f after the spine", th[i-1], th[i])
			break
		}
	}
	// Nowhere is the stroke thicker than the spine, nor thinner than the arcs.
	for i, v := range th {
		if v > p.SSpineThick+1 || v < p.SArcThick-1 {
			t.Errorf("thickness %.2f at %d of %d outside %v to %v", v, i, len(th), p.SArcThick, p.SSpineThick)
			break
		}
	}
	_ = fmt.Sprint
}

// The upper half is the lower half turned through 180 degrees about the
// centre, scaled: near the spine the outline is point-symmetric, and at the
// extremes it is not.
func TestSPointSymmetry(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := sGlyph(t, p)
	f := newSFrame(p)
	pts, _ := sPoints(path)
	reflect := func(q outline.Point) outline.Point {
		return outline.Point{X: 2*f.C.X - q.X, Y: 2*f.C.Y - q.Y}
	}
	nearest := func(q outline.Point) float64 {
		best := math.Inf(1)
		for i := range pts {
			best = math.Min(best, segDist(q, pts[i], pts[(i+1)%len(pts)]))
		}
		return best
	}
	for _, q := range pts {
		if math.Hypot(q.X-f.C.X, q.Y-f.C.Y) < 40 {
			if d := nearest(reflect(q)); d > 0.2 {
				t.Errorf("point (%.1f, %.1f) near the spine has no partner by the half turn, %.2f away", q.X, q.Y, d)
				break
			}
		}
	}
	// Away from the spine the outline is not point-symmetric: the right extreme
	// of the lower half has no partner.
	far := 0.0
	for _, q := range pts {
		far = math.Max(far, nearest(reflect(q)))
	}
	if far < 20 {
		t.Errorf("the whole letter is point-symmetric (largest distance of a reflected point %.1f), but the halves differ", far)
	}
}

// The outer edge of each arc ends at the cut of its beak in a true corner and
// not in a rounding: the cut is vertical (the outermost ink lies on the line of
// the cut from the corner down to the end of the serif, within 0.01 units), the
// outline turns by at least 20 degrees within 4 units on either side of the
// corner (the arc comes into its tip at about 30 degrees from the horizontal,
// a turn of 60 degrees), and the round
// from the start of the hollow to the extreme is one convex curve with
// neither a bump nor a straight stretch (it turns by at least 0.04 radians over
// every 40 units).
func TestSBeaksHaveCorners(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := sGlyph(t, p)
	if len(path[0].Segs) != 20 {
		t.Fatalf("%d segments, want 20", len(path[0].Segs))
	}
	for i, v := range sBeakViews(t, path, newSFrame(p)) {
		corner := v.Top - p.SBeakDrop
		for y := v.Serif.End + 1; y < corner-1; y++ {
			r := runs(v.Path, y)
			if d := math.Abs(r[len(r)-1][1] - v.Cut); d > 0.01 {
				t.Errorf("the cut of the %s beak is %.2f off the vertical at %.0f", sBeakNames[i], d, y)
				break
			}
		}
	}
	pts, end := sPoints(path)
	angle := func(a, b outline.Point) float64 { return math.Atan2(b.Y-a.Y, b.X-a.X) }
	wrap := func(d float64) float64 { return math.Mod(d+3*math.Pi, 2*math.Pi) - math.Pi }
	// Segments 5 and 15 are the cuts; the outer edges start at their ends with
	// the hollow (TestBeakHollow) and go on with the round (the segments 7 and
	// 17), whose edge is one convex curve.
	for _, tc := range []struct {
		name       string
		cut, first int
	}{{"lower", 5, 7}, {"upper", 15, 17}} {
		i := end[tc.cut]
		turn := wrap(angle(pts[i], pts[i+2]) - angle(pts[i-2], pts[i]))
		if math.Abs(turn) < 20*math.Pi/180 {
			t.Errorf("the outer corner of the %s beak turns by %.1f degrees within 4 units, want at least 20", tc.name, turn*180/math.Pi)
		}
		cu := sCubicAt(path, tc.first)
		var prev outline.Point
		var sign, length, lastDir float64
		var turned, lens []float64
		for k := 0; k <= 400; k++ {
			q := cu.at(float64(k) / 400)
			if k > 0 {
				dir := angle(prev, q)
				var d float64
				if k > 1 {
					d = wrap(dir - lastDir)
				}
				if sign == 0 && math.Abs(d) > 1e-9 {
					sign = d
				}
				if d*sign < -1e-9 {
					t.Errorf("the outer edge of the %s beak has a bump near %v", tc.name, q)
					break
				}
				length += math.Hypot(q.X-prev.X, q.Y-prev.Y)
				total := math.Abs(d)
				if len(turned) > 0 {
					total += turned[len(turned)-1]
				}
				turned = append(turned, total)
				lens = append(lens, length)
				lastDir = dir
			}
			prev = q
		}
		for a := range turned {
			b := a
			for b < len(turned) && lens[b]-lens[a] < 40 {
				b++
			}
			if b == len(turned) {
				break
			}
			if turned[b]-turned[a] < 0.04 {
				t.Errorf("the outer edge of the %s beak is straight over 40 units from %.0f of its length (turns %.3f)", tc.name, lens[a], turned[b]-turned[a])
				break
			}
		}
	}
}
