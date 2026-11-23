// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

const tol = 2.0

const (
	// footScanY is the height of the scan through the foot serif: one unit
	// above the baseline.
	footScanY = 1
	// tipInset is how far inside the serif tip the vertical scan is placed.
	tipInset = 5
)

func near(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %.1f, want %.1f", what, got, want)
	}
}

func TestO(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, advance, ok := outlineOf("o", p)
	if !ok {
		t.Fatal("glyph o not found")
	}
	g := Glyph{Name: "o", Rune: 'o', Advance: advance, Path: path}

	if g.Advance != 486 || g.Advance != design.OAdvance(p) {
		t.Errorf("advance %v, want 486", g.Advance)
	}
	if len(g.Path) != 2 {
		t.Fatalf("o has %d contours, want 2", len(g.Path))
	}
	if g.Path[0].SignedArea() <= 0 {
		t.Error("outer contour of o is not counter-clockwise")
	}
	if g.Path[1].SignedArea() >= 0 {
		t.Error("inner contour of o is not clockwise")
	}

	x0, y0, x1, y1 := g.Path.BBox()
	near(t, "left sidebearing", x0, p.SideRoundLeft)
	near(t, "right sidebearing", design.OAdvance(p)-x1, p.SideRoundRight)
	near(t, "bottom", y0, -p.OvershootLower)
	near(t, "top", y1, p.XHeight+p.OvershootLower)

	// The thickest part of the bowl, at mid height.
	xs := g.Path.CrossingsX(p.XHeight / 2)
	if len(xs) != 4 {
		t.Fatalf("crossings at mid height: %v", xs)
	}
	near(t, "left curve thickness", xs[1]-xs[0], p.CurveLower)
	near(t, "right curve thickness", xs[3]-xs[2], p.CurveLower)

	// The thinnest part, at the top and bottom, on the vertical centre line.
	ys := g.Path.CrossingsY((x0 + x1) / 2)
	if len(ys) != 4 {
		t.Fatalf("crossings on the centre line: %v", ys)
	}
	near(t, "bottom thickness", ys[1]-ys[0], p.ThinLower)
	near(t, "top thickness", ys[3]-ys[2], p.ThinLower)
}

// diagonalThickness measures the stroke of o along the diagonals.  Between
// two super-elliptic contours the thickness at a diagonal is the distance from
// the point where the inner contour meets the diagonal of its own bounding box
// (the ray from the centre to a corner) to the nearest point of the outer
// contour.  The result has one value for each of the four quadrants.
func diagonalThickness(t *testing.T, path outline.Path) [4]float64 {
	t.Helper()
	if len(path) != 2 {
		t.Fatalf("o has %d contours, want 2", len(path))
	}
	out, inn := path[0].Flatten(), path[1].Flatten()
	bbox := func(pts []outline.Point) (x0, y0, x1, y1 float64) {
		x0, y0, x1, y1 = pts[0].X, pts[0].Y, pts[0].X, pts[0].Y
		for _, q := range pts {
			x0, x1 = math.Min(x0, q.X), math.Max(x1, q.X)
			y0, y1 = math.Min(y0, q.Y), math.Max(y1, q.Y)
		}
		return
	}
	x0, y0, x1, y1 := bbox(inn)
	cx, cy := (x0+x1)/2, (y0+y1)/2
	var res [4]float64
	for i, corner := range []outline.Point{{X: x1, Y: y1}, {X: x0, Y: y1}, {X: x0, Y: y0}, {X: x1, Y: y0}} {
		dx, dy := corner.X-cx, corner.Y-cy
		// Intersect the ray c + s*(dx, dy), 0 < s <= 1, with the inner contour.
		hit, found := outline.Point{}, false
		for j, a := range inn {
			b := inn[(j+1)%len(inn)]
			ex, ey := b.X-a.X, b.Y-a.Y
			det := dx*ey - dy*ex
			if det == 0 {
				continue
			}
			// c + s*d = a + u*e
			s := ((a.X-cx)*ey - (a.Y-cy)*ex) / det
			u := ((a.X-cx)*dy - (a.Y-cy)*dx) / det
			// The ray goes through a vertex of the flattened contour, the
			// middle of a quarter arc, and the ends of the segment are
			// accepted within a tolerance.
			const eps = 1e-9
			if s > 0 && u >= -eps && u <= 1+eps {
				hit, found = outline.Point{X: cx + s*dx, Y: cy + s*dy}, true
				break
			}
		}
		if !found {
			t.Fatalf("diagonal %d does not meet the inner contour", i)
		}
		best := math.Inf(1)
		for j, a := range out {
			b := out[(j+1)%len(out)]
			ex, ey := b.X-a.X, b.Y-a.Y
			u := ((hit.X-a.X)*ex + (hit.Y-a.Y)*ey) / (ex*ex + ey*ey)
			u = math.Max(0, math.Min(1, u))
			best = math.Min(best, math.Hypot(a.X+u*ex-hit.X, a.Y+u*ey-hit.Y))
		}
		res[i] = best
	}
	return res
}

// Range of the thickness along the diagonal as a fraction of the thickest
// part of the stroke.
const (
	diagRatioMin = 0.49
	diagRatioMax = 0.84
)

// diagonalProblems lists the quadrants of o in which the thickness along the
// diagonal is not strictly between the thin and the thick stroke, or is
// outside the range diagRatioMin to diagRatioMax relative to the thick stroke.
func diagonalProblems(t *testing.T, p design.Params) []string {
	t.Helper()
	path, _, _ := outlineOf("o", p)
	var problems []string
	for i, d := range diagonalThickness(t, path) {
		if !(d > p.ThinLower && d < p.CurveLower) {
			problems = append(problems, fmt.Sprintf("quadrant %d: thickness %.1f is not between %.0f and %.0f",
				i, d, p.ThinLower, p.CurveLower))
		}
		if r := d / p.CurveLower; r < diagRatioMin || r > diagRatioMax {
			problems = append(problems, fmt.Sprintf("quadrant %d: thickness %.1f is %.2f of the thickest part, outside %.2f to %.2f",
				i, d, r, diagRatioMin, diagRatioMax))
		}
	}
	return problems
}

func TestODiagonalThickness(t *testing.T) {
	t.Parallel()
	for _, v := range design.Variants() {
		path, _, _ := outlineOf("o", v.Params)
		t.Logf("%s: diagonal thickness (NE, NW, SW, SE): %.1f", v.Name, diagonalThickness(t, path))
		for _, msg := range diagonalProblems(t, v.Params) {
			t.Errorf("%s: %s", v.Name, msg)
		}
	}
}

// The check must be able to fail: a bad shape is reported.
func TestODiagonalThicknessCatchesBadShape(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	p.SupernessOuter, p.SupernessInner = 0.76, 0.66
	if len(diagonalProblems(t, p)) == 0 {
		t.Error("an inner contour at 0.66 inside an outer one at 0.76 was not reported")
	}
}

// mirrored reports whether the sorted values vs are symmetric about the
// centre c within tolerance one unit.
func mirrored(vs []float64, c float64) bool {
	for i := range vs {
		if math.Abs(vs[i]+vs[len(vs)-1-i]-2*c) > 1 {
			return false
		}
	}
	return true
}

func TestOMirror(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, g := range []Glyph{roundedOf("o", p)} {
		if g.Name != "o" {
			continue
		}
		x0, y0, x1, y1 := g.Path.BBox()
		cx, cy := (x0+x1)/2, (y0+y1)/2
		// The left and right halves: scan lines at many heights.
		for y := y0 + 1; y < y1; y += 3 {
			xs := g.Path.CrossingsX(y)
			if len(xs) == 0 || !mirrored(xs, cx) {
				t.Errorf("crossings at height %.0f are not symmetric about x = %.1f: %v", y, cx, xs)
			}
		}
		// The top and bottom halves: scan lines at many places across.
		for x := x0 + 1; x < x1; x += 3 {
			ys := g.Path.CrossingsY(x)
			if len(ys) == 0 || !mirrored(ys, cy) {
				t.Errorf("crossings at x = %.0f are not symmetric about height %.1f: %v", x, cy, ys)
			}
		}
	}
}

// checkStemFeet checks the foot serifs of a stem whose edges are at stemLeft
// and stemRight, and which has a serif of length SerifJut on both sides, in
// the path of a letter with no other ink near the baseline: the lengths, the
// thickness at the tips, the flat top, and the concave brackets (checkFoot).
func checkStemFeet(t *testing.T, p design.Params, path outline.Path, stemLeft, stemRight float64) {
	t.Helper()
	checkFoot(t, p, path, stemRight, true, p.SerifJut, p.SerifThick)
	checkFoot(t, p, path, stemLeft, false, p.SerifJut, p.SerifThick)
}

// checkHeadSerif checks the sloped head serif on the left of a stem with the
// edges stemLeft and stemRight and the flat top at height top: its length,
// the slope of its upper edge, its thickness at the tip, its concave
// underside, and that the stem has no serif on its right.  The path may hold
// other ink, also a foot serif of the stem and other stems, as long as it holds
// none on the vertical lines through the serif which lies above the serif and
// none to the right of the stem at the height of its top: the helper looks for
// the crossings near where the serif has them.
func checkHeadSerif(t *testing.T, p design.Params, path outline.Path, stemLeft, stemRight, top float64) {
	t.Helper()
	checkHeadSerifBelow(t, p, path, stemLeft, stemRight, top, true)
}

// checkHeadSerifBelow is checkHeadSerif with the check of the right-hand edge
// just below the head serif optional: the stem of p is joined by the upper
// stroke of its bowl 20 units below the underside of the serif and has no
// right-hand edge there.
func checkHeadSerifBelow(t *testing.T, p design.Params, path outline.Path, stemLeft, stemRight, top float64, rightBelow bool) {
	t.Helper()
	// The head serif, just below the tip: on the left only.
	tipTop := top - p.HeadSerifDrop
	hs := path.CrossingsX(tipTop - 1)
	near(t, "head serif left edge", stemLeft-nearest(t, "head serif tip", hs, stemLeft-p.SerifJut), p.SerifJut)
	nearest(t, "head serif right edge", hs, stemRight)
	// No serif on the right of the stem: the stem ends at its edge.
	nearest(t, "right edge of the stem just below the top", path.CrossingsX(top-1), stemRight)

	// The upper edge slopes: it leaves the top of the stem at the top height
	// and falls by HeadSerifDrop towards the tip.
	for _, f := range []float64{0.25, 0.5, 0.75} {
		us := path.CrossingsY(stemLeft - f*p.SerifJut)
		near(t, "upper edge of the head serif", nearest(t, "upper edge", us, top-f*p.HeadSerifDrop), top-f*p.HeadSerifDrop)
	}
	near(t, "top of the stem", lastOf(path.CrossingsY(stemLeft+p.Stem/2)), top)

	// The tip is SerifThick thick (the upper edge is a little higher five
	// units in from the tip).
	hts := path.CrossingsY(stemLeft - p.SerifJut + tipInset)
	upperAt := top - p.HeadSerifDrop*(1-tipInset/p.SerifJut)
	upper := nearest(t, "head serif tip height", hts, upperAt)
	var lower float64
	for _, v := range hts {
		if v < upper-1 && v > lower {
			lower = v
		}
	}
	near(t, "head serif thickness", upper-lower, p.SerifThick+p.HeadSerifDrop*tipInset/p.SerifJut)
	near(t, "head serif tip height", upper, upperAt)

	// The head serif bracket is concave, as the foot brackets are.
	hy := tipTop - p.SerifThick - headUnderRun/2
	hb := path.CrossingsX(hy)
	edge := math.NaN()
	for _, v := range hb {
		if v > stemLeft-p.SerifJut && v < stemLeft+0.5 && (math.IsNaN(edge) || v < edge) {
			edge = v
		}
	}
	if o := stemLeft - edge; math.IsNaN(edge) || o < 0.1*p.SerifJut || o > 0.4*p.SerifJut {
		t.Errorf("head serif underside is %.1f from the stem at half its rise, want between %.1f and %.1f (concave)",
			o, 0.1*p.SerifJut, 0.4*p.SerifJut)
	}

	// Just below the head serif bracket, only the stem remains on the left.
	bs := path.CrossingsX(tipTop - p.SerifThick - headUnderRun - 2)
	near(t, "stem left below head", nearest(t, "stem left below head", bs, stemLeft), stemLeft)
	if rightBelow {
		nearest(t, "stem right below head", bs, stemRight)
	}
}

func lastOf(vs []float64) float64 { return vs[len(vs)-1] }

func TestL(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, advance, ok := outlineOf("l", p)
	if !ok {
		t.Fatal("glyph l not found")
	}
	g := Glyph{Name: "l", Rune: 'l', Advance: advance, Path: path}

	if g.Advance != 270 || g.Advance != design.NarrowAdvance(p) {
		t.Errorf("advance %v, want 270", g.Advance)
	}
	if len(g.Path) != 1 {
		t.Fatalf("l has %d contours, want 1", len(g.Path))
	}

	_, y0, _, y1 := g.Path.BBox()
	near(t, "bottom", y0, 0)
	near(t, "top", y1, p.Ascender)

	// The stem, measured at half the x-height, above the foot serif.
	xs := g.Path.CrossingsX(p.XHeight / 2)
	if len(xs) != 2 {
		t.Fatalf("crossings at mid height: %v", xs)
	}
	near(t, "stem width", xs[1]-xs[0], p.Stem)
	stemLeft, stemRight := xs[0], xs[1]

	checkStemFeet(t, p, g.Path, stemLeft, stemRight)

	// The stem has the stem sides of n, less the cut: 99 and 90.
	near(t, "left side of the stem", stemLeft, design.NarrowSideLeft(p))
	near(t, "right side of the stem", g.Advance-stemRight, design.NarrowSideRight(p))

	checkHeadSerif(t, p, g.Path, stemLeft, stemRight, p.Ascender)
}

func TestSetOrder(t *testing.T) {
	t.Parallel()
	gs := Set(design.Regular())
	want := []string{".notdef", "space", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	// The 26 capitals follow the lowercase, in alphabetical order.
	want = append(want, strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ", "")...)
	// The ten figures follow the capitals.
	want = append(want, "zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine")
	// The basic punctuation and the signs of ASCII mathematics.
	want = append(want, strings.Fields("period comma colon semicolon exclam quotesingle quoteleft quoteright quotedbl quotedblleft quotedblright hyphen endash emdash parenleft parenright bracketleft bracketright slash backslash bar underscore asterisk plus minus equal less greater")...)
	// The question mark, the ampersand and the symbols of the basic set follow.
	want = append(want, symbolGlyphs...)
	// The dotless i and j, the spacing accents, the combining marks and the
	// composed letters follow.
	want = append(want, "dotlessi", "uni0237")
	for _, s := range spacingList {
		want = append(want, s.name)
	}
	for _, s := range combiningList {
		want = append(want, s.name)
	}
	for _, c := range composedList {
		want = append(want, c.name)
	}
	for _, l := range ligatureList {
		want = append(want, l.name)
	}
	if len(gs) != len(want) {
		t.Fatalf("%d glyphs, want %d", len(gs), len(want))
	}
	for i, n := range want {
		if gs[i].Name != n {
			t.Errorf("glyph %d is %q, want %q", i, gs[i].Name, n)
		}
	}
}

func TestTable(t *testing.T) {
	t.Parallel()
	if len(table) == 0 || table[0].name != ".notdef" {
		t.Fatal("the first table entry must be .notdef")
	}
	names := make(map[string]bool)
	runes := make(map[rune]bool)
	for i, e := range table {
		if e.name == "" {
			t.Errorf("entry %d has no name", i)
		}
		if names[e.name] {
			t.Errorf("duplicate name %q", e.name)
		}
		names[e.name] = true
		if e.r != 0 {
			if runes[e.r] {
				t.Errorf("duplicate rune %q", e.r)
			}
			runes[e.r] = true
		}
		if e.draw == nil {
			t.Errorf("entry %q has no drawing function", e.name)
		}
	}
}

func TestOutlineUnknown(t *testing.T) {
	t.Parallel()
	if _, _, ok := outlineOf("nonexistent", design.Regular()); ok {
		t.Error("Outline found a glyph which does not exist")
	}
}

// archLetters are the letters whose arch is that of n below the x-height.
var archLetters = []string{"n", "h"}

// archFrame holds the stem edges of n or h, read off at a height of 100, and
// the outline.  The tests of the arch measure everything else relative to
// these.
type archFrame struct {
	name             string
	p                design.Params
	path             outline.Path
	advance          float64
	l0, l1, r0, r1   float64 // edges of the left and the right stem
	xh               float64
	stemTop          float64         // the flat top of the left stem
	top              float64         // the highest point of the letter
	archTop, archBot []outline.Point // outer edge and underside of the arch
}

func newArchFrame(t *testing.T, name string, p design.Params) *archFrame {
	t.Helper()
	path, advance, ok := outlineOf(name, p)
	if !ok {
		t.Fatalf("glyph %s not found", name)
	}
	f := &archFrame{name: name, p: p, path: path, advance: advance, xh: p.XHeight}
	f.stemTop, f.top = p.XHeight, p.XHeight+p.ArchOvershoot
	if name == "h" {
		f.stemTop, f.top = p.Ascender, p.Ascender
	}
	xs := path.CrossingsX(100)
	if len(xs) != 4 {
		t.Fatalf("%s has %d crossings at height 100, want 4: %v", name, len(xs), xs)
	}
	f.l0, f.l1, f.r0, f.r1 = xs[0], xs[1], xs[2], xs[3]
	// The arch is sampled at every half unit
	// between the stems, the bottom and the top of the topmost run of the
	// vertical scan, where that run lies above 0.3 of the x-height.
	for x := f.l1 + 0.5; x < f.r0-0.25; x += 0.5 {
		ys := path.CrossingsY(x)
		if len(ys) >= 2 && ys[len(ys)-2] > 0.3*p.XHeight {
			f.archBot = append(f.archBot, outline.Point{X: x, Y: ys[len(ys)-2]})
			f.archTop = append(f.archTop, outline.Point{X: x, Y: ys[len(ys)-1]})
		}
	}
	if len(f.archTop) < 20 {
		t.Fatalf("%s has no arch between the stems (%d samples)", name, len(f.archTop))
	}
	return f
}

// eachArchLetter runs a test for n and for h, whose arches, stems and feet
// are measured in the same way.
func eachArchLetter(t *testing.T, test func(t *testing.T, f *archFrame)) {
	t.Helper()
	for _, name := range archLetters {
		t.Run(name, func(t *testing.T) {
			test(t, newArchFrame(t, name, design.Regular()))
		})
	}
}

// segDist is the distance from the point q to the line segment ab.
func segDist(q, a, b outline.Point) float64 {
	ex, ey := b.X-a.X, b.Y-a.Y
	u := ((q.X-a.X)*ex + (q.Y-a.Y)*ey) / (ex*ex + ey*ey)
	u = math.Max(0, math.Min(1, u))
	return math.Hypot(a.X+u*ex-q.X, a.Y+u*ey-q.Y)
}

// thickness returns, for each sample of the outer edge of the arch, the
// distance to the nearest point of the underside, which is the thickness of
// the arch at right angles to the stroke.
func (f *archFrame) thickness() []float64 {
	d := make([]float64, len(f.archTop))
	for i, q := range f.archTop {
		d[i] = math.Inf(1)
		for j := 0; j+1 < len(f.archBot); j++ {
			d[i] = math.Min(d[i], segDist(q, f.archBot[j], f.archBot[j+1]))
		}
	}
	return d
}

// joinIndex is the index of the sample of the outer edge at which the arch
// is said to leave the left stem: 0.35 stem width to the right of the stem.
func (f *archFrame) joinIndex() int {
	want := f.l1 + 0.35*f.p.Stem
	best := 0
	for i, q := range f.archTop {
		if math.Abs(q.X-want) < math.Abs(f.archTop[best].X-want) {
			best = i
		}
	}
	return best
}

func TestArchStems(t *testing.T) {
	t.Parallel()
	eachArchLetter(t, func(t *testing.T, f *archFrame) {
		p := f.p
		if f.advance != 540 || f.advance != design.NAdvance(p) {
			t.Errorf("advance %v, want 540", f.advance)
		}
		near(t, "left edge of the left stem", f.l0, p.SideStemLeft)
		near(t, "left stem width", f.l1-f.l0, p.Stem)
		near(t, "counter", f.r0-f.l1, p.NCounter)
		near(t, "right stem width", f.r1-f.r0, p.Stem)
		near(t, "right side", f.advance-f.r1, p.SideStemRight)
		if design.NGap(p) != p.NCounter {
			t.Errorf("NGap = %v, want the counter %v", design.NGap(p), p.NCounter)
		}
		// The outer edge of the right stem is straight up to ShoulderStart, and
		// the inner edge up to ShoulderStartInner.
		us := f.path.CrossingsX(p.ShoulderStart - 5)
		near(t, "outer edge below the shoulder", us[len(us)-1], f.r1)
		is := f.path.CrossingsX(p.ShoulderStartInner - 5)
		near(t, "inner edge below the shoulder", is[len(is)-2], f.r0)
		// The edge runs out into the stem with zero curvature.
		if above := f.path.CrossingsX(p.ShoulderStart + 50); above[len(above)-1] > f.r1-1 {
			t.Errorf("the outer edge is still at %.1f 50 units above ShoulderStart", above[len(above)-1])
		}
	})
}

func TestArchFeet(t *testing.T) {
	t.Parallel()
	eachArchLetter(t, func(t *testing.T, f *archFrame) {
		p := f.p
		fs := f.path.CrossingsX(footScanY)
		if len(fs) != 4 {
			t.Fatalf("crossings at the foot: %v", fs)
		}
		near(t, "left stem, outer jut", f.l0-fs[0], p.SerifJut)
		near(t, "left stem, inner jut", fs[1]-f.l1, p.SerifJutInner)
		near(t, "right stem, inner jut", f.r0-fs[2], p.SerifJutInner)
		near(t, "right stem, outer jut", fs[3]-f.r1, p.SerifJut)
		// Above the brackets only the stems remain.
		ss := f.path.CrossingsX(p.BracketHeight + 1)
		if len(ss) != 4 {
			t.Fatalf("crossings above the foot brackets: %v", ss)
		}
		near(t, "left stem above the bracket", ss[0], f.l0)
		near(t, "left stem above the bracket", ss[1], f.l1)
		// The serif is SerifThick thick at the tips and the brackets are concave.
		for _, x := range []float64{fs[0], fs[3]} {
			in := tipInset
			if x == fs[3] {
				in = -tipInset
			}
			ys := f.path.CrossingsY(x + float64(in))
			if len(ys) != 2 && len(ys) != 4 {
				t.Fatalf("crossings at a serif tip: %v", ys)
			}
			near(t, "serif thickness", ys[1]-ys[0], p.SerifThick)
		}
		midY := p.SerifThick + (p.BracketHeight-p.SerifThick)/2
		ms := f.path.CrossingsX(midY)
		if len(ms) != 4 {
			t.Fatalf("crossings at mid bracket: %v", ms)
		}
		for i, o := range []float64{f.l0 - ms[0], ms[1] - f.l1, f.r0 - ms[2], ms[3] - f.r1} {
			if o < 0.1*p.BracketReach || o > 0.4*p.BracketReach {
				t.Errorf("bracket %d offset at half height is %.1f, want between %.1f and %.1f (concave)",
					i, o, 0.1*p.BracketReach, 0.4*p.BracketReach)
			}
		}
	})
}

func TestArchSimple(t *testing.T) {
	t.Parallel()
	eachArchLetter(t, func(t *testing.T, f *archFrame) {
		p := f.p
		if len(f.path) != 1 {
			t.Fatalf("%s has %d contours, want 1", f.name, len(f.path))
		}
		// Between the notch and the top of the underside the arch is a band of
		// its own, beside the free top of the left stem and the right stem, and
		// the scan crosses six edges.  Everywhere else it crosses at most four,
		// and above the arch of h only the stem remains.
		top := p.XHeight + p.ArchOvershoot
		for y := 5.0; y < f.top; y += 10 {
			n := len(f.path.CrossingsX(y))
			limit := 4
			if y > p.ArchNotch && y < top-p.ThinLower {
				limit = 6
			}
			if y > top {
				limit = 2
			}
			if n%2 != 0 || n > limit || n == 0 {
				t.Errorf("the scan at height %.0f has %d crossings, want 2 to %d, an even number", y, n, limit)
			}
		}
	})
}

func TestArchHeight(t *testing.T) {
	t.Parallel()
	eachArchLetter(t, func(t *testing.T, f *archFrame) {
		_, y0, _, y1 := f.path.BBox()
		near(t, "bottom of "+f.name, y0, 0)
		near(t, "highest point of "+f.name, y1, f.top)
	})
}

func TestArchTop(t *testing.T) {
	t.Parallel()
	eachArchLetter(t, func(t *testing.T, f *archFrame) {
		p := f.p
		// The highest point of the outer edge, as the centre of the part within
		// half a unit of the top.
		var top float64
		for _, q := range f.archTop {
			top = math.Max(top, q.Y)
		}
		var sum, cnt float64
		for _, q := range f.archTop {
			if q.Y >= top-0.5 {
				sum += q.X
				cnt++
			}
		}
		xt := sum / cnt
		near(t, "height of the arch", top, p.XHeight+p.ArchOvershoot)
		c1 := (f.l0 + f.l1) / 2
		D := (f.r0+f.r1)/2 - c1
		if got := (xt - c1) / D; math.Abs(got-p.ArchPeak) > 0.01 {
			t.Errorf("the arch peaks at %.3f of the stem distance, want %.3f", got, p.ArchPeak)
		}
		// The vertical thickness on the vertical through the highest point.
		ys := f.path.CrossingsY(xt)
		if len(ys) < 2 {
			t.Fatalf("crossings on the vertical through the peak: %v", ys)
		}
		if got := ys[len(ys)-1] - ys[len(ys)-2]; math.Abs(got-p.ThinLower) > 3 {
			t.Errorf("arch thickness at the peak is %.1f, want %.1f within 3", got, p.ThinLower)
		}
	})
}

func TestArchNotchAndJoin(t *testing.T) {
	t.Parallel()
	eachArchLetter(t, func(t *testing.T, f *archFrame) {
		p := f.p

		// The notch: just below the x-height the left stem is a run of its own.
		ns := f.path.CrossingsX(p.XHeight - 8)
		if len(ns) != 4 {
			t.Fatalf("crossings just below the x-height: %v", ns)
		}
		near(t, "right edge of the stem below the x-height", ns[1], f.l1)
		if ns[2]-ns[1] < 15 {
			t.Errorf("the gap between the stem and the arch is %.1f at the x-height, want at least 15", ns[2]-ns[1])
		}
		// The notch height: the top of the
		// outer edge 2 and 5 units to the right of the stem, extended linearly
		// back to the stem.
		top := func(dx float64) float64 { c := f.path.CrossingsY(f.l1 + dx); return c[len(c)-1] }
		y2, y5 := top(2), top(5)
		notch := math.Min(p.XHeight, y2-(y5-y2)*2/3)
		if math.Abs(notch-p.ArchNotch) > 3 {
			t.Errorf("notch at %.1f, want %.0f within 3", notch, p.ArchNotch)
		}

		// The join: below it the right edge of the left stem is straight and the
		// counter reaches it; 15 units above, the underside is clearly off it.
		below := f.path.CrossingsX(p.ArchJoin - 3)
		if len(below) != 4 || math.Abs(below[1]-f.l1) > 1 {
			t.Errorf("3 units below ArchJoin the counter does not reach the stem: %v", below)
		}
		above := f.path.CrossingsX(p.ArchJoin + 15)
		if len(above) != 4 || above[1] < f.l1+3 {
			t.Errorf("15 units above ArchJoin the underside has not left the stem: %v", above)
		}
		// The join: the lowest height, going up from 0.3
		// of the x-height, at which the counter lies more than 4 units to the
		// right of the stem.
		join := math.NaN()
		for y := 0.3 * p.XHeight; y < p.XHeight; y += 0.25 {
			xs := f.path.CrossingsX(y)
			// The counter begins at the right end of the run before the last.
			if len(xs) >= 4 && xs[len(xs)-3] > f.l1+4 {
				join = y
				break
			}
		}
		if join < p.ArchJoin || join > p.ArchJoin+10 {
			t.Errorf("join measured at %.1f, want between %.0f and %.0f", join, p.ArchJoin, p.ArchJoin+10)
		}
	})
}

func TestArchThickness(t *testing.T) {
	t.Parallel()
	eachArchLetter(t, func(t *testing.T, f *archFrame) {
		p := f.p
		d := f.thickness()
		k := f.joinIndex()
		if testing.Verbose() {
			for i := 0; i < len(d); i += 10 {
				t.Logf("x %.1f outer %.1f under %.1f thickness %.1f", f.archTop[i].X, f.archTop[i].Y, f.archBot[i].Y, d[i])
			}
		}
		if math.Abs(d[k]-p.ArchThinJoin) > 3 {
			t.Errorf("arch thickness where it leaves the stem is %.1f, want %.1f within 3", d[k], p.ArchThinJoin)
		}
		// From the join to the right stem the arch only grows thicker: it never
		// falls below the thickness at the join, and never loses more than 1.5
		// from the greatest thickness so far.
		var run float64
		for i := k; i < len(d) && f.archTop[i].X < f.r0-10; i++ {
			run = math.Max(run, d[i])
			if d[i] < p.ArchThinJoin-1.5 {
				t.Errorf("at x = %.0f the arch is %.1f thick, below ArchThinJoin = %.0f", f.archTop[i].X, d[i], p.ArchThinJoin)
				break
			}
			if run-d[i] > 1.5 {
				t.Errorf("at x = %.0f the arch is %.1f thick after having been %.1f: a pinch", f.archTop[i].X, d[i], run)
				break
			}
		}
		// Next to the right stem the thickness at right angles is already
		// above half the stem.
		if last := d[len(d)-1]; last < 0.5*p.Stem {
			t.Errorf("the arch is only %.1f thick next to the right stem", last)
		}
		// Below the underside, the right stem is the full stem wide where its
		// edges are straight, and the width narrows steadily as the shoulder
		// rises into the arch.
		prev := math.Inf(1)
		for y := p.ShoulderStartInner; y < p.XHeight-p.ThinLower-15; y += 5 {
			xs := f.path.CrossingsX(y)
			if len(xs) < 4 {
				t.Fatalf("crossings at height %.0f: %v", y, xs)
			}
			w := xs[len(xs)-1] - xs[len(xs)-2]
			if w > prev+1 {
				t.Errorf("the shoulder widens from %.1f to %.1f at height %.0f", prev, w, y)
			}
			prev = w
		}
	})
}

func TestArchHeadSerif(t *testing.T) {
	t.Parallel()
	eachArchLetter(t, func(t *testing.T, f *archFrame) {
		checkHeadSerif(t, f.p, f.path, f.l0, f.l1, f.stemTop)
	})
}

// archMetrics measures the counter of n.  The peak of the underside is the
// centre of the part within half a unit of its highest point, as a fraction
// of the distance between the stem centres.  The roundness is the width of the counter 15 units
// below its highest point, as a fraction of the width between the stems.
func archMetrics(t *testing.T, f *archFrame) (peak, round float64) {
	t.Helper()
	var top float64
	for _, q := range f.archBot {
		top = math.Max(top, q.Y)
	}
	var sum, cnt float64
	for _, q := range f.archBot {
		if q.Y >= top-0.5 {
			sum += q.X
			cnt++
		}
	}
	c1 := (f.l0 + f.l1) / 2
	D := (f.r0+f.r1)/2 - c1
	peak = (sum/cnt - c1) / D

	// Six crossings: left stem, outer edge, underside, and the right stem
	// merged with the arch.  The counter lies between the middle two.
	xs := f.path.CrossingsX(top - 15)
	if len(xs) != 6 {
		t.Fatalf("crossings 15 units below the top of the counter: %v", xs)
	}
	round = (xs[4] - xs[3]) / (f.r0 - f.l1)
	return peak, round
}

// TestNCounterTop checks that the white inside the arch of n has a round top
// in the middle of the counter and not a point above its right half.
func TestArchCounterTop(t *testing.T) {
	t.Parallel()
	eachArchLetter(t, func(t *testing.T, f *archFrame) {
		peak, round := archMetrics(t, f)
		t.Logf("counter peak at %.3f, width 15 below the top %.3f of the counter", peak, round)
		if peak < 0.524 || peak > 0.600 {
			t.Errorf("the counter peaks at %.3f of the stem distance, want 0.524 to 0.600", peak)
		}
		if round < 0.47 {
			t.Errorf("the counter is %.3f of its width wide 15 units below its top, want at least 0.47 (a point fails)", round)
		}
	})
}

// headSerifMetrics returns the ink area of the head serif left of the stem
// and its vertical thickness 12 units from the stem and halfway to the tip.
func headSerifMetrics(t *testing.T, path outline.Path, stemLeft, jut float64) (area, near12, half float64) {
	t.Helper()
	thick := func(x float64) float64 {
		ys := path.CrossingsY(x)
		if len(ys) < 2 {
			t.Fatalf("no head serif at x = %.1f: %v", x, ys)
		}
		return ys[len(ys)-1] - ys[len(ys)-2]
	}
	for x := stemLeft - jut + 0.25; x < stemLeft; x += 0.5 {
		area += 0.5 * thick(x)
	}
	return area, thick(stemLeft - 12), thick(stemLeft - jut/2)
}

// TestHeadSerifShape checks that the head serif is a tapering serif with a
// curved underside and not a block of even thickness.
func TestHeadSerifShape(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, c := range []struct {
		name          string
		areaMax, nMax float64
	}{
		{"n", 3000, 52},
		{"h", 3000, 52},
		{"l", 3000, 52},
		{"i", 3000, 52},
	} {
		path, _, _ := outlineOf(c.name, p)
		stem := outline.Path{path[0]} // for i the dot is the second contour
		xs := stem.CrossingsX(p.XHeight / 2)
		area, n12, half := headSerifMetrics(t, stem, xs[0], p.SerifJut)
		t.Logf("%s: area %.0f, thickness 12 from the stem %.1f, halfway %.1f", c.name, area, n12, half)
		if area > c.areaMax {
			t.Errorf("%s: head serif area %.0f, want at most %.0f", c.name, area, c.areaMax)
		}
		if n12 > c.nMax {
			t.Errorf("%s: head serif %.1f thick 12 units from the stem, want at most %.0f", c.name, n12, c.nMax)
		}
	}
}

// TestHAscender checks the ascender of h: the highest point is the ascender
// height of l, the stem above the arch is as wide as the stem of l and stands
// where the left stem of n stands, and the head serif is that of l, measured
// from the top of the stem.
func TestHAscender(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	h, _, _ := outlineOf("h", p)
	l, _, _ := outlineOf("l", p)
	n, _, _ := outlineOf("n", p)
	_, _, _, hTop := h.BBox()
	_, _, _, lTop := l.BBox()
	near(t, "highest point of h", hTop, lTop)
	near(t, "highest point of h", hTop, p.Ascender)

	// Above the arch only the stem remains.
	hx := h.CrossingsX(p.XHeight + p.ArchOvershoot + 20)
	lx := l.CrossingsX(p.XHeight / 2)
	nx := n.CrossingsX(100)
	if len(hx) != 2 {
		t.Fatalf("crossings above the arch of h: %v", hx)
	}
	near(t, "stem width of the ascender", hx[1]-hx[0], lx[1]-lx[0])
	near(t, "left edge of the ascender", hx[0], nx[0])
	near(t, "right edge of the ascender", hx[1], nx[1])

	// The head serif has the shape of that of l, relative to the top of the
	// stem and to the left edge of the stem: the edges at fixed depths below
	// the top, and the heights of both edges at fixed distances from the stem.
	for _, d := range []float64{1, 5, 12, 20, 30, 40, 50, 60} {
		hs, ls := h.CrossingsX(p.Ascender-d), l.CrossingsX(p.Ascender-d)
		near(t, fmt.Sprintf("head serif tip %.0f below the top", d), hx[0]-hs[0], lx[0]-ls[0])
	}
	for _, f := range []float64{0.05, 0.25, 0.5, 0.75, 0.95} {
		hy := h.CrossingsY(hx[0] - f*p.SerifJut)
		ly := l.CrossingsY(lx[0] - f*p.SerifJut)
		if len(hy) != 4 || len(ly) != 4 {
			t.Fatalf("crossings on the head serif at %.2f: h %v, l %v", f, hy, ly)
		}
		near(t, "upper edge of the head serif", hy[3], ly[3])
		near(t, "underside of the head serif", hy[2], ly[2])
	}
}

// TestHJoinsLikeN checks that the arch of h is that of n: below the head serif
// of n the two outlines have the same crossings with every horizontal line, in
// the arch, in the right stem and in the left stem.
func TestHJoinsLikeN(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	h, ha, _ := outlineOf("h", p)
	n, na, _ := outlineOf("n", p)
	if ha != na {
		t.Errorf("advance of h is %v, of n %v", ha, na)
	}
	for y := 5.0; y < p.XHeight-60; y += 7 {
		hx, nx := h.CrossingsX(y), n.CrossingsX(y)
		if len(hx) != len(nx) {
			t.Errorf("at height %.0f h has %d crossings, n has %d", y, len(hx), len(nx))
			continue
		}
		for i := range hx {
			if math.Abs(hx[i]-nx[i]) > 0.01 {
				t.Errorf("at height %.0f crossing %d is at %.2f in h and %.2f in n", y, i, hx[i], nx[i])
			}
		}
	}
}

// dotOf returns the contour of the dot of i and the path of its stem.
func dotOf(t *testing.T, path outline.Path) (dot, stem outline.Path) {
	t.Helper()
	if len(path) != 2 {
		t.Fatalf("i has %d contours, want 2 (the stem and the dot)", len(path))
	}
	return outline.Path{path[1]}, outline.Path{path[0]}
}

func TestI(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, advance, ok := outlineOf("i", p)
	if !ok {
		t.Fatal("glyph i not found")
	}
	if advance != 270 || advance != design.NarrowAdvance(p) {
		t.Errorf("advance %v, want 270", advance)
	}
	dot, stem := dotOf(t, path)

	// The stem: as wide as that of l, with its sides, flat at the
	// x-height and with the foot serifs of l.
	_, y0, _, y1 := stem.BBox()
	near(t, "bottom of the stem", y0, 0)
	near(t, "top of the stem", y1, p.XHeight)
	xs := stem.CrossingsX(p.XHeight / 2)
	if len(xs) != 2 {
		t.Fatalf("crossings at mid height: %v", xs)
	}
	near(t, "stem width", xs[1]-xs[0], p.Stem)
	near(t, "left side of the stem", xs[0], design.NarrowSideLeft(p))
	near(t, "right side of the stem", advance-xs[1], design.NarrowSideRight(p))
	checkStemFeet(t, p, stem, xs[0], xs[1])
	checkHeadSerif(t, p, stem, xs[0], xs[1], p.XHeight)

	// The stem is that of l below the x-height.
	l, _, _ := outlineOf("l", p)
	for y := 1.0; y < p.XHeight-60; y += 7 {
		if is, ls := stem.CrossingsX(y), l.CrossingsX(y); len(is) != len(ls) || math.Abs(is[0]-ls[0]) > 0.01 || math.Abs(is[1]-ls[1]) > 0.01 {
			t.Errorf("at height %.0f the stem of i has the crossings %v, that of l %v", y, is, ls)
		}
	}

	// The dot: a separate contour, counter-clockwise like every outer contour.
	if a := path[1].SignedArea(); a <= 0 {
		t.Errorf("the dot has the signed area %.0f, want a counter-clockwise contour", a)
	}
	x0, dy0, x1, dy1 := dot.BBox()
	near(t, "width of the dot", x1-x0, p.DotDiameter)
	near(t, "height of the dot", dy1-dy0, p.DotDiameter)
	near(t, "height of the centre of the dot", (dy0+dy1)/2, p.DotCentre)
	near(t, "offset of the dot from the centre of the stem", (x0+x1)/2-(xs[0]+xs[1])/2, design.DotOffset(p))
	// The clear distance to the highest point of the stem and of its serif.
	_, _, _, top := stem.BBox()
	gap := dy0 - top
	near(t, "clear distance of the dot above the stem", gap, p.DotCentre-p.DotDiameter/2-p.XHeight)
	if gap < 60 {
		t.Errorf("the dot is only %.1f above the stem", gap)
	}
	// A circle fills pi/4 of its bounding box, and the dot is symmetrical.
	if fill := path[1].SignedArea() / ((x1 - x0) * (dy1 - dy0)); math.Abs(fill-math.Pi/4) > 0.005 {
		t.Errorf("the dot fills %.4f of its bounding box, want %.4f (a circle)", fill, math.Pi/4)
	}
	cx, cy := (x0+x1)/2, (dy0+dy1)/2
	for _, d := range []float64{10, 25, 40} {
		up := dot.CrossingsX(cy + d)
		down := dot.CrossingsX(cy - d)
		if len(up) != 2 || len(down) != 2 {
			t.Fatalf("crossings of the dot at %.0f from its centre: %v, %v", d, up, down)
		}
		near(t, "dot, left edge mirrored", cx-up[0], cx-down[0])
		near(t, "dot, right edge mirrored", up[1]-cx, down[1]-cx)
		// The horizontal half chord of a circle of radius r at distance d.
		r := p.DotDiameter / 2
		near(t, "dot, half chord", (up[1]-up[0])/2, math.Sqrt(r*r-d*d))
	}
}

// dotShiftOf returns the distance of the centre of the dot of the rounded
// glyph name ("i" or "j") to the left of the centre of its stem, measured at a
// height of 100 above the baseline.
func dotShiftOf(t *testing.T, p design.Params, name string) float64 {
	t.Helper()
	for _, g := range []Glyph{roundedOf(name, p)} {
		if g.Name != name {
			continue
		}
		dot, stem := dotOf(t, g.Path)
		x0, _, x1, _ := dot.BBox()
		xs := stem.CrossingsX(100)
		if len(xs) != 2 {
			t.Fatalf("%s: crossings of the stem at 100: %v", name, xs)
		}
		return (xs[0]+xs[1])/2 - (x0+x1)/2
	}
	t.Fatalf("glyph %s not found", name)
	return 0
}

// Without a head serif the dot of i and j stands exactly over the centre of
// the stem: the shift is a fraction of the length of the serif.  The serif
// parameters which a letter without serifs needs are zero.
func TestDotCentredWithoutSerif(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	p.SerifJut, p.SerifJutInner, p.HeadSerifDrop = 0, 0, 0
	for _, name := range []string{"i", "j"} {
		if s := dotShiftOf(t, p, name); math.Abs(s) > 0.5 {
			t.Errorf("%s: the dot is %.2f left of the centre of the stem without serifs, want 0", name, s)
		}
	}
}

// In the regular parameters the dot of i and j lies left of the centre of the
// stem by DotShift times the length of the head serif, to within the rounding
// of the outline.
func TestDotShiftIsFractionOfSerif(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	want := p.DotShift * p.SerifJut
	if want < 5 || want > 7 {
		t.Errorf("shift %.2f, want about 6", want)
	}
	for _, name := range []string{"i", "j"} {
		if s := dotShiftOf(t, p, name); math.Abs(s-want) > 1 {
			t.Errorf("%s: the dot is %.2f left of the centre of the stem, want %.2f", name, s, want)
		}
	}
}
