// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
	"seehuhn.de/go/geom/path"
)

// The structural checks do not depend on the shape of a letter.  Check applies
// them to every glyph of the table.

// Check reports whether the parameters are valid: every glyph of the font can
// be built from p and passes the structural checks, unrounded and rounded.
// The error names each glyph which fails, with the stage and the check.
func Check(p design.Params) error {
	// The glyphs are checked in parallel; the errors are in the order of the
	// table.  Each glyph which is not composed is built once, and the composed
	// letters use the builds of their bases.
	results := make([]error, len(table))
	bs := &checkBuilds{p: p, m: map[string]*checkBuild{}}
	var wg sync.WaitGroup
	gate := make(chan struct{}, runtime.GOMAXPROCS(0))
	for i, e := range table {
		wg.Add(1)
		gate <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-gate }()
			results[i] = checkGlyph(p, e.name, bs)
		}()
	}
	wg.Wait()
	return errors.Join(results...)
}

// CheckGlyph is Check for one glyph.  A panic of the construction, such as a
// terminal which cannot be placed, is returned as an error.
func CheckGlyph(p design.Params, name string) error {
	return checkGlyph(p, name, nil)
}

// checkGlyph is CheckGlyph with the builds bs, which may be nil.
func checkGlyph(p design.Params, name string, bs *checkBuilds) (err error) {
	e, ok := lookup(name)
	if !ok {
		return fmt.Errorf("glyphs: no glyph %s", name)
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("glyph %s: the construction fails: %v", name, r)
		}
	}()
	c := composites[name]
	var path, rounded outline.Path
	var adv float64
	switch {
	case bs == nil:
		path, adv = e.draw(p)
		rounded = path.Round()
		if c != nil {
			rounded = c.roundedFromScratch(p)
		}
	case c != nil:
		bp, badv := bs.get(c.base)
		dx, dy, f := c.place(p, bp)
		m := markOutline(p, c.mark, f)
		path, adv = c.assemble(bp, m, dx, dy), badv
		rounded = c.rounded(bp.Round(), m.Round(), bp, m, dx, dy)
	default:
		path, adv = bs.get(name)
		rounded = path.Round()
	}
	if problems := checkStages(p, name, path, adv, rounded); len(problems) > 0 {
		return fmt.Errorf("glyph %s: %s", name, problems[0])
	}
	return nil
}

// checkBuilds holds the outlines of the glyphs which are not composed, each
// drawn once for the parameters p when it is first asked for.  It is safe for
// concurrent use; the outlines are shared and are not changed.
type checkBuilds struct {
	p  design.Params
	mu sync.Mutex
	m  map[string]*checkBuild
}

type checkBuild struct {
	once  sync.Once
	path  outline.Path
	adv   float64
	fault any // what the construction panicked with
}

// get returns the outline and the advance of the glyph, and panics as its
// construction panics.
func (bs *checkBuilds) get(name string) (outline.Path, float64) {
	bs.mu.Lock()
	b := bs.m[name]
	if b == nil {
		b = new(checkBuild)
		bs.m[name] = b
	}
	bs.mu.Unlock()
	b.once.Do(func() {
		defer func() { b.fault = recover() }()
		b.path, b.adv = checkDraw(name, bs.p)
	})
	if b.fault != nil {
		panic(b.fault)
	}
	return b.path, b.adv
}

// checkDraw draws the glyph which is not composed for Check.  Tests replace it
// by a function which shares the builds.
var checkDraw = func(name string, p design.Params) (outline.Path, float64) {
	e, _ := lookup(name)
	return e.draw(p)
}

// checkStages runs all checks on the unrounded and on the rounded outline of a
// glyph and returns the problems, each with the stage and the check.  Rounding
// leaves the advance as it is.
func checkStages(p design.Params, glyph string, path outline.Path, adv float64, rounded outline.Path) []string {
	if c := composites[glyph]; c != nil && !c.joined() {
		return checkComposed(p, glyph, c, path, adv, rounded)
	}
	var problems []string
	for _, stage := range []string{"unrounded", "rounded"} {
		c := structureCase{params: p, glyph: glyph, path: path, advance: adv}
		if stage == "rounded" {
			c.path = rounded
			for _, msg := range checkIntegers(c) {
				problems = append(problems, stage+", check integers: "+msg)
			}
		}
		for _, ch := range structureChecks {
			for _, msg := range ch.run(c) {
				problems = append(problems, fmt.Sprintf("%s, check %s: %s", stage, ch.name, msg))
			}
		}
	}
	return problems
}

// checkComposed runs the checks on a composed letter whose mark is not joined
// to its base.  The base is unchanged and is checked on its own; the letter is
// checked for its mark, which must pass the checks of the outline, for its
// place in the advance, and for the distance between the mark and the base,
// which must be at least the clearance floor.  The mark is the last contours
// of the letter.
func checkComposed(p design.Params, glyph string, c *composite, path outline.Path, adv float64, rounded outline.Path) []string {
	n := len(markOutline(p, c.mark, formLower))
	var problems []string
	for _, stage := range []string{"unrounded", "rounded"} {
		full := path
		if stage == "rounded" {
			full = rounded
		}
		if len(full) <= n {
			problems = append(problems, fmt.Sprintf("%s: the letter has %d contours, not more than its mark", stage, len(full)))
			continue
		}
		base, mark := full[:len(full)-n], full[len(full)-n:]
		mc := structureCase{params: p, glyph: glyph, path: mark, advance: adv}
		if stage == "rounded" {
			for _, msg := range checkIntegers(mc) {
				problems = append(problems, stage+", check integers: "+msg)
			}
		}
		for _, ch := range structureChecks {
			if ch.name == "bbox" {
				mc.path = full
			}
			for _, msg := range ch.run(mc) {
				problems = append(problems, fmt.Sprintf("%s, check %s of the mark: %s", stage, ch.name, msg))
			}
			mc.path = mark
		}
		floor := clearanceFloor(p)
		if d := pathsDistance(base, mark, floor); d < floor {
			problems = append(problems, fmt.Sprintf("%s, check clearance: the mark is only %.2f from the base, want at least %v", stage, d, floor))
		}
	}
	return problems
}

// pathsDistance returns the least distance between the flattened outlines a and
// b, or limit if that is less.
func pathsDistance(a, b outline.Path, limit float64) float64 {
	bx0, by0, bx1, by1 := b.BBox()
	var bs [][2]outline.Point
	for _, ct := range b {
		pts := ct.Flatten()
		for i := range pts {
			bs = append(bs, [2]outline.Point{pts[i], pts[(i+1)%len(pts)]})
		}
	}
	best := limit
	for _, ct := range a {
		pts := ct.Flatten()
		for i := range pts {
			p, q := pts[i], pts[(i+1)%len(pts)]
			if math.Min(p.X, q.X) > bx1+best || math.Max(p.X, q.X) < bx0-best ||
				math.Min(p.Y, q.Y) > by1+best || math.Max(p.Y, q.Y) < by0-best {
				continue
			}
			for _, s := range bs {
				best = math.Min(best, segmentDistance(p, q, s[0], s[1]))
			}
		}
	}
	return best
}

// segmentDistance returns the distance between the segments pq and rs.
func segmentDistance(p, q, r, s outline.Point) float64 {
	cross := func(a, b, c outline.Point) float64 { return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X) }
	if cross(p, q, r)*cross(p, q, s) < 0 && cross(r, s, p)*cross(r, s, q) < 0 {
		return 0
	}
	ptSeg := func(a, b, c outline.Point) float64 {
		dx, dy := c.X-b.X, c.Y-b.Y
		l := dx*dx + dy*dy
		t := 0.0
		if l > 0 {
			t = math.Max(0, math.Min(1, ((a.X-b.X)*dx+(a.Y-b.Y)*dy)/l))
		}
		return math.Hypot(a.X-b.X-t*dx, a.Y-b.Y-t*dy)
	}
	return math.Min(math.Min(ptSeg(p, r, s), ptSeg(q, r, s)), math.Min(ptSeg(r, p, q), ptSeg(s, p, q)))
}

// A structureCase is one glyph in one parameter set.
type structureCase struct {
	params  design.Params
	glyph   string
	path    outline.Path
	advance float64
}

// A structureCheck returns the problems it finds, each as a sentence which
// says where.
type structureCheck struct {
	name string
	run  func(c structureCase) []string
}

var structureChecks = []structureCheck{
	{"contours", checkContours},
	{"orientation", checkOrientation},
	{"intersections", checkIntersections},
	{"scans", checkScans},
	{"bbox", checkBBox},
	{"extremes", checkExtremes},
	{"clearance", checkClearance},
}

func checkIntegers(c structureCase) []string {
	var problems []string
	for i, ct := range c.path {
		check := func(what string, q outline.Point) {
			if q.X != math.Round(q.X) || q.Y != math.Round(q.Y) {
				problems = append(problems, fmt.Sprintf("contour %d, %s has the non-integer coordinates %v", i, what, q))
			}
		}
		check("start", ct.Start)
		for j, s := range ct.Segs {
			check(fmt.Sprintf("segment %d end", j), s.P3)
			if s.IsCurve {
				check(fmt.Sprintf("segment %d first handle", j), s.P1)
				check(fmt.Sprintf("segment %d second handle", j), s.P2)
			}
		}
	}
	return problems
}

// checkContours requires finite coordinates, at least three segments, no
// segment of length zero (also not the one which closes the contour) and
// an area which is not zero.
func checkContours(c structureCase) []string {
	var problems []string
	for i, ct := range c.path {
		cur := ct.Start
		ok := true
		for j, s := range ct.Segs {
			for _, q := range []outline.Point{s.P1, s.P2, s.P3} {
				if math.IsNaN(q.X) || math.IsNaN(q.Y) || math.IsInf(q.X, 0) || math.IsInf(q.Y, 0) {
					problems = append(problems, fmt.Sprintf("contour %d, segment %d has the coordinate %v", i, j, q))
					ok = false
				}
			}
			if s.P3 == cur && (!s.IsCurve || (s.P1 == cur && s.P2 == cur)) {
				problems = append(problems, fmt.Sprintf("contour %d, segment %d has length zero at %v", i, j, cur))
			}
			cur = s.P3
		}
		if len(ct.Segs) < 3 {
			problems = append(problems, fmt.Sprintf("contour %d has only %d segments", i, len(ct.Segs)))
		}
		// The closing segment is implicit when the last point differs from
		// the start; the contour is closed in both cases.
		if ok && math.Abs(ct.SignedArea()) < 1e-9 {
			problems = append(problems, fmt.Sprintf("contour %d has area zero", i))
		}
	}
	return problems
}

// inside reports whether q lies inside the polygon (even-odd rule).
func inside(poly []outline.Point, q outline.Point) bool {
	in := false
	for i, a := range poly {
		b := poly[(i+1)%len(poly)]
		if (a.Y <= q.Y) == (b.Y <= q.Y) {
			continue
		}
		x := a.X + (q.Y-a.Y)/(b.Y-a.Y)*(b.X-a.X)
		if x > q.X {
			in = !in
		}
	}
	return in
}

// checkOrientation requires counter-clockwise contours at even nesting depth
// and clockwise contours at odd depth.  The depth of a contour is the number
// of other contours which contain its first point.
func checkOrientation(c structureCase) []string {
	var problems []string
	polys := make([][]outline.Point, len(c.path))
	for i, ct := range c.path {
		polys[i] = ct.Flatten()
	}
	for i, ct := range c.path {
		depth := 0
		for j := range c.path {
			if j != i && inside(polys[j], ct.Start) {
				depth++
			}
		}
		area := ct.SignedArea()
		switch {
		case depth%2 == 0 && area < 0:
			problems = append(problems, fmt.Sprintf("contour %d (outer, depth %d) is clockwise", i, depth))
		case depth%2 == 1 && area > 0:
			problems = append(problems, fmt.Sprintf("contour %d (inner, depth %d) is counter-clockwise", i, depth))
		}
	}
	return problems
}

func checkIntersections(c structureCase) []string {
	var problems []string
	for _, q := range c.path.SelfIntersections() {
		problems = append(problems, fmt.Sprintf("the outline crosses or touches itself at (%.2f, %.2f)", q.X, q.Y))
	}
	return problems
}

// scanPositions returns positions in [lo, hi), 6.71 apart from lo + 0.137,
// each moved by 0.0123 until it differs from every value in avoid by at least
// 1e-6.
func scanPositions(lo, hi float64, avoid []float64) []float64 {
	var out []float64
	for v := lo + 0.137; v < hi; v += 6.71 {
		for hit := true; hit; {
			hit = false
			for _, a := range avoid {
				if math.Abs(a-v) < 1e-6 {
					v += 0.0123
					hit = true
				}
			}
		}
		out = append(out, v)
	}
	return out
}

// checkScans requires an even number of crossings on horizontal and vertical
// scans, and that the contours, with their orientation, wind at most once:
// moving along the scan, outer contours lift the winding number from 0 to 1
// and inner ones lower it again.
func checkScans(c structureCase) []string {
	if len(c.path) == 0 {
		return nil
	}
	var problems []string
	var xs, ys []float64
	polys := make([][]outline.Point, len(c.path))
	for i, ct := range c.path {
		polys[i] = ct.Flatten()
		for _, q := range polys[i] {
			xs = append(xs, q.X)
			ys = append(ys, q.Y)
		}
	}
	x0, y0, x1, y1 := c.path.BBox()
	for _, y := range scanPositions(y0, y1, ys) {
		if n := len(c.path.CrossingsX(y)); n%2 != 0 {
			problems = append(problems, fmt.Sprintf("the horizontal scan at height %.2f has %d crossings", y, n))
		}
		if bad := windingProblem(polys, y, true); bad != "" {
			problems = append(problems, fmt.Sprintf("the horizontal scan at height %.2f: %s", y, bad))
		}
	}
	for _, x := range scanPositions(x0, x1, xs) {
		if n := len(c.path.CrossingsY(x)); n%2 != 0 {
			problems = append(problems, fmt.Sprintf("the vertical scan at x %.2f has %d crossings", x, n))
		}
		if bad := windingProblem(polys, x, false); bad != "" {
			problems = append(problems, fmt.Sprintf("the vertical scan at x %.2f: %s", x, bad))
		}
	}
	return problems
}

// windingProblem follows the scan line at v (horizontal if horiz) across the
// flattened contours polys and returns a description if the winding number
// leaves {0, 1} or does not end at 0.
func windingProblem(polys [][]outline.Point, v float64, horiz bool) string {
	type crossing struct {
		pos float64
		dir int
	}
	var cs []crossing
	for _, pts := range polys {
		for i, a := range pts {
			b := pts[(i+1)%len(pts)]
			if !horiz {
				a, b = outline.Point{X: a.Y, Y: a.X}, outline.Point{X: b.Y, Y: b.X}
			}
			if (a.Y <= v) == (b.Y <= v) {
				continue
			}
			pos := a.X + (v-a.Y)/(b.Y-a.Y)*(b.X-a.X)
			dir := 1
			if b.Y < a.Y {
				dir = -1
			}
			if !horiz {
				// Swapping the axes mirrors the plane.
				dir = -dir
			}
			cs = append(cs, crossing{pos, dir})
		}
	}
	for i := range cs {
		for j := i + 1; j < len(cs); j++ {
			if cs[j].pos < cs[i].pos {
				cs[i], cs[j] = cs[j], cs[i]
			}
		}
	}
	w := 0
	for _, c := range cs {
		// Moving along a horizontal scan, an upward crossing is the right
		// edge of a counter-clockwise contour and lowers the winding.
		w -= c.dir
		if w < 0 || w > 1 {
			return fmt.Sprintf("the winding number is %d at %.2f", w, c.pos)
		}
	}
	if w != 0 {
		return fmt.Sprintf("the winding number ends at %d", w)
	}
	return ""
}

// checkBBox requires the bounding box to lie within the advance and, on each
// side, to keep the distance from the edge of the advance which the kind of
// that side gives: see entry.inkMargins.  A glyph which is not in the table, or
// whose entry does not say the kind of its sides, has no rule and fails.
func checkBBox(c structureCase) []string {
	e, ok := lookup(c.glyph)
	if !ok || e.left == sideUnknown || e.right == sideUnknown {
		return []string{fmt.Sprintf("no bounding-box rule for glyph %s", c.glyph)}
	}
	if len(c.path) == 0 || e.left == sideMark {
		return nil
	}
	left, right := e.inkMargins(c.params)
	if composites[c.glyph] != nil {
		// The mark of a composed letter may reach Acc.Overhang beyond the
		// advance; the base is checked on its own.
		left = math.Min(left, -c.params.Acc.Overhang)
		right = math.Min(right, -c.params.Acc.Overhang)
	}
	var problems []string
	x0, _, x1, _ := c.path.BBox()
	if x0 < left-0.5 {
		problems = append(problems, fmt.Sprintf("the outline starts at %v, want at least %v", x0, left))
	}
	if c.advance-x1 < right-0.5 {
		problems = append(problems, fmt.Sprintf("the outline ends %v before the advance, want at least %v", c.advance-x1, right))
	}
	return problems
}

// clearanceRatio is the ratio of arc length to distance above which
// Path.MinDistance treats two parts of one contour as not neighbours.
const clearanceRatio = design.SlitRatio

// clearanceFloor returns the smallest distance allowed between a counter and
// the outer contour and between the two sides of a stroke: design.ClearanceFloor.
func clearanceFloor(p design.Params) float64 {
	return design.ClearanceFloor(p)
}

// A sharpCrotch is an inner corner of a glyph which the design makes narrower
// than the slit angle: a white wedge which ends in a point under solid ink,
// such as the wedges between the stems and the diagonals of M.  Near its tip
// the two edges are closer than the floor and far apart along the outline.  Tip
// is the corner and Dirs are the unit vectors along its two edges, away from
// the tip.
type sharpCrotch struct {
	Name string
	Tip  outline.Point
	Dirs [2]outline.Point
}

// sharpCrotches lists, for each glyph which has them, the sharp crotches of the
// design in the parameter set.  Only a glyph listed here is exempt, and only near
// these corners: see crotchSkip.
var sharpCrotches = map[string]func(p design.Params) []sharpCrotch{
	"M": capMCrotches,
}

// crotchTolerance is the distance from the line of an edge within which a piece of
// the outline counts as lying on it.
const crotchTolerance = 1.0

// crotchRadius returns the distance from the tip of the crotch k within which its
// edges are exempt from the clearance check: (floor + crotchTolerance) / sin(a) for
// the opening angle a.  Every point of one edge farther from the tip than this is
// at least the floor from the other edge, with crotchTolerance to spare.
func crotchRadius(k sharpCrotch, floor float64) float64 {
	u, v := k.Dirs[0], k.Dirs[1]
	return (floor + crotchTolerance) / math.Abs(u.X*v.Y-u.Y*v.X)
}

// crotchSkip returns the function which tells MinDistanceExcept to leave out a pair
// of pieces of the outline: both lie on the edges of one listed crotch (within
// crotchTolerance of the line of an edge, on the far side of the tip) and each
// starts or ends within crotchRadius of its tip.  A piece which reaches beyond the
// radius is followed by pieces wholly beyond it, which are measured.  Nothing else
// is left out.
func crotchSkip(ks []sharpCrotch, floor float64) func(a0, a1, b0, b1 outline.Point) bool {
	if len(ks) == 0 {
		return nil
	}
	onEdge := func(k sharpCrotch, q outline.Point) bool {
		dx, dy := q.X-k.Tip.X, q.Y-k.Tip.Y
		for _, u := range k.Dirs {
			if dx*u.X+dy*u.Y > -crotchTolerance && math.Abs(dx*u.Y-dy*u.X) <= crotchTolerance {
				return true
			}
		}
		return false
	}
	near := func(k sharpCrotch, r float64, q0, q1 outline.Point) bool {
		if !onEdge(k, q0) || !onEdge(k, q1) {
			return false
		}
		d0 := math.Hypot(q0.X-k.Tip.X, q0.Y-k.Tip.Y)
		d1 := math.Hypot(q1.X-k.Tip.X, q1.Y-k.Tip.Y)
		return math.Min(d0, d1) <= r
	}
	return func(a0, a1, b0, b1 outline.Point) bool {
		for _, k := range ks {
			r := crotchRadius(k, floor)
			if near(k, r, a0, a1) && near(k, r, b0, b1) {
				return true
			}
		}
		return false
	}
}

// checkClearance requires the distance between two contours, such as a counter
// and the outer contour, and between two parts of one contour which are not
// neighbours, such as the two sides of a stroke, to be at least the floor.  The
// edges of the sharp crotches which the design lists (sharpCrotches) are measured
// only beyond crotchRadius from their tips.
func checkClearance(c structureCase) []string {
	floor := clearanceFloor(c.params)
	var skip func(a0, a1, b0, b1 outline.Point) bool
	if f, ok := sharpCrotches[c.glyph]; ok {
		skip = crotchSkip(f(c.params), floor)
	}
	d, a, b := c.path.MinDistanceExcept(clearanceRatio, skip)
	if d < floor {
		return []string{fmt.Sprintf("two parts of the outline are only %.2f apart, at (%.1f, %.1f) and (%.1f, %.1f), want at least %v", d, a.X, a.Y, b.X, b.Y, floor)}
	}
	return nil
}

// extremeTolerance is the distance by which a curve may run beyond both of
// its end points before an on-curve point is required there.
const extremeTolerance = 0.5

// checkExtremes requires that no cubic segment leaves the range of its end
// points, in x or in y, by more than extremeTolerance: the extremes of every
// curve are on-curve points.
func checkExtremes(c structureCase) []string {
	var problems []string
	for i, ct := range c.path {
		cur := ct.Start
		for j, s := range ct.Segs {
			if s.IsCurve {
				for axis, name := range []string{"x", "y"} {
					get := func(q outline.Point) float64 {
						if axis == 0 {
							return q.X
						}
						return q.Y
					}
					v0, v1, v2, v3 := get(cur), get(s.P1), get(s.P2), get(s.P3)
					lo, hi := math.Min(v0, v3), math.Max(v0, v3)
					for _, t := range outline.CubicExtrema(v0, v1, v2, v3) {
						v := path.Cubic{{X: v0}, {X: v1}, {X: v2}, {X: v3}}.At(t).X
						if v > hi+extremeTolerance {
							problems = append(problems, fmt.Sprintf("contour %d, segment %d from %v to %v runs %.2f beyond its end points in the %s direction (maximum %.2f at t=%.2f)", i, j, cur, s.P3, v-hi, name, v, t))
						} else if v < lo-extremeTolerance {
							problems = append(problems, fmt.Sprintf("contour %d, segment %d from %v to %v runs %.2f beyond its end points in the %s direction (minimum %.2f at t=%.2f)", i, j, cur, s.P3, lo-v, name, v, t))
						}
					}
				}
			}
			cur = s.P3
		}
	}
	return problems
}
