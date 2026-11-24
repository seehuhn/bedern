// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"seehuhn.de/go/bedern/internal/diag"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// The italic profile of the smoothness check, as a filter of the findings of
// fair.Default: an inflection whose two curved sides are separated by a
// straight segment at least inflLine long is not a finding, and every curve
// which runs beyond both its end points by more than extremeTol, in x or in
// y, is a finding of the kind fair.Extreme, except that a curve next to a
// stem of the slant, or next to the curve which touches it at a smooth node,
// may run up to stemExtremeTol beyond in x.
const (
	inflLine       = 40
	extremeTol     = 0.5
	stemExtremeTol = 2
)

// extremeKind is the kind of the findings of extremes without a node.
const extremeKind fair.Kind = "extreme"

// A judged entry is a finding of the smoothness check which is known: one
// which belongs to the form (Defect false), or a defect which is not repaired
// yet (Defect true), with the reason in words, as the baseline of the roman
// (bedern/internal/fair/testdata/baseline) judges the findings of the roman.
type judged struct {
	Kind   fair.Kind
	X, Y   float64
	Dir    float64
	Defect bool
	Reason string
}

// beakExtreme is the reason of an extreme next to the beak of an arm.
const beakExtreme = "the curve under the arm runs into the straight inner edge of the beak, which the shear turns past the vertical; the extreme lies within 15 units of that corner, where no node can stand, and 0.6 beyond the nodes"

// serifExtreme is the reason of an extreme in the bracket of a head serif of
// a diagonal.
const serifExtreme = "the roman's node where the bracket of the head serif meets the diagonal stands at its upright extreme with a handle of 4 units; the shear moves the extreme 4 units below that node, too close for a node of its own, and less than 0.7 beyond the nodes"

// sectionWeights is the reason of the findings of the section sign.
const sectionWeights = "the roman code of the section sign, which fits its hooks and its spine to its weights, draws at the weights of the italic (the spine and the hooks 0.94 of the roman) the inner turns of the hooks tighter (r 21 to 24 against r 52) and the spine with sharper turns; at its own weights it has none of these"

// archTurn is the reason of the inner turn at the top right of the arch of the
// superior 2.
const archTurn = "the inner edge at the top right of the arch of the superior 2 turns tighter (r 22 against r 72 on either side) than that of the 2 at its own size, which has no finding: the figure drawn with the heavier strokes of a superior figure and scaled has the smaller inner round, and the shear closes it; the 2 has no round to turn (figures.go), as the figure 2 has none"

// judgedFindings are the known findings, by glyph.
var judgedFindings = map[string][]judged{
	"uni018F": {{Kind: fair.Kink, X: 553, Y: 356, Dir: -104.5, Defect: true, Reason: "the roman's eye of the capital schwa ends in a straight piece of 17 units at its right extreme, which the slant turns by 1.9 degrees against the curve above it; the lattice puts the lower end of the piece at 339, between two multiples of 4"}},
	"E":       {{Kind: extremeKind, X: 649, Y: 615, Dir: -90, Reason: beakExtreme}},
	"F":       {{Kind: extremeKind, X: 639, Y: 615, Dir: -90, Reason: beakExtreme}},
	"W":       {{Kind: extremeKind, X: 276, Y: 618, Dir: 90, Reason: serifExtreme}},
	"A":       {{Kind: extremeKind, X: 478, Y: 62, Dir: -90, Reason: "the roman's node where the bracket of the right foot serif leaves the thick diagonal stands at its upright extreme with a short handle; the shear moves the extreme a few units along the bracket, too close to that node for one of its own, and 0.56 beyond the nodes"}},
	"V": {
		{Kind: extremeKind, X: 277, Y: 620, Dir: 90, Reason: serifExtreme},
		{Kind: fair.Wiggle, X: 290, Y: 641, Dir: 41, Defect: true, Reason: "the curvature of the bracket of the left head serif dips between the roman's node with the short handle and the flat of the serif, where the shear shortens the bracket on the acute side of the thin diagonal"},
	},
	"Y":          {{Kind: fair.Wiggle, X: 290, Y: 635, Dir: 68, Defect: true, Reason: "the curvature of the bracket of the left head serif dips between the roman's node with the short handle and the flat of the serif, as on V"}},
	"R":          {{Kind: fair.Kink, X: 437, Y: 67, Dir: -67, Defect: true, Reason: "the rounding turns the handle of 5 units where the bracket of the left foot serif of the leg leaves the straight leg by 1.6 degrees"}},
	"registered": {{Kind: fair.Jump, X: 519, Y: 191, Dir: 114, Defect: true, Reason: "a finding of the R drawn with heavier strokes and scaled, as in the roman: the curvature of the round of the bowl jumps to the straight edge of the leg where the joins are made for the strokes of the full size; the windows of the analysis are scaled with the letter"}},
	"five":       {{Kind: fair.Short, X: 105, Y: 375, Dir: -53, Defect: true, Reason: "the short round of the inner corner where the stem meets the bowl, the defect of the roman at that place (short 99 402 and jump 107 371 in its baseline)"}},
	"three": {
		{Kind: fair.Short, X: 166, Y: 608, Dir: 36, Defect: true, Reason: "a curve of 12 units in the neck of the upper ball, which the deformation that keeps the ball a circle leaves between two nodes"},
		{Kind: fair.Inflection, X: 162, Y: 606, Dir: 37, Defect: true, Reason: "the inflection in the neck of the upper ball, on the short curve beside it"},
	},
	"ampersand": {
		{Kind: fair.Jump, X: 717, Y: 610, Dir: 90, Reason: "the half circle of the round end of the stem of the t meets the side of the stem, as every round end does"},
		{Kind: fair.Jump, X: 387, Y: 663, Dir: 0, Defect: true, Reason: "the curvature of the counter of the upper bowl falls from r 57 to r 132 at its top node, where the automatic curvatures of the stroke tool share the turn between the top and the stalk unevenly; a stated curvature there is the repair"},
	},
	"uni00B2": {
		{Kind: fair.Spike, X: 300, Y: 667, Dir: -44, Defect: true, Reason: archTurn},
		{Kind: fair.Short, X: 61, Y: 300, Dir: -103, Reason: "the cut at the left end of the foot of the superior 2 is 13 units long, the thickness of the thin foot of a small figure, which the lattice moves up by 2 units at its lower edge; it is shorter than the 15 units of the check and no defect"},
	},
	"onehalf": {{Kind: fair.Spike, X: 636, Y: 377, Dir: -45, Defect: true, Reason: archTurn + ", here in the denominator"}},
	"questiondown": {
		{Kind: fair.Kink, X: 276, Y: 232, Dir: 75, Defect: true, Reason: "where the curve of the stem meets its straight end the roman code of the turned question mark has a vertical curve of 1.9 units and a line of 30 units, too short for a stem of the lattice; the shear and the nodes at the extremes leave the end of the stem 0.9 units off the slant, so that the edge turns by 12 degrees at the node"},
		{Kind: fair.Short, X: 276, Y: 233, Dir: 63, Defect: true, Reason: "the line of 2.2 units next to that kink, the vertical curve of 1.9 units of the roman code"},
	},
	"dollar": {{Kind: fair.Jump, X: 444, Y: 544, Dir: -90, Defect: true, Reason: "the defect of the roman (jump 378 515 in its baseline), where the upper arc of the S meets its straight part, which the nodes at the extremes of the slanted form move 29 units along the edge"}},
	"section": {
		{Kind: fair.Spike, X: 164, Y: 455, Dir: -73, Defect: true, Reason: sectionWeights},
		{Kind: fair.Spike, X: 288, Y: 85, Dir: 108, Defect: true, Reason: sectionWeights},
		{Kind: fair.Jump, X: 264, Y: 150, Dir: 180, Defect: true, Reason: sectionWeights},
		{Kind: fair.Spike, X: 264, Y: 150, Dir: 180, Defect: true, Reason: sectionWeights},
		{Kind: fair.Jump, X: 187, Y: 390, Dir: 0, Defect: true, Reason: sectionWeights},
		{Kind: fair.Spike, X: 187, Y: 390, Dir: 0, Defect: true, Reason: sectionWeights},
	},
	"dagger": {{Kind: fair.Inflection, X: 302, Y: 281, Dir: 72, Defect: true, Reason: "the roman's tail leaves its swelling with no curvature for its straight blade; the nodes at the extremes of the sheared tail cut the swelling where its curvature is nearly nothing (r 1261), and the curve on the far side of that node bends the other way by a trace"}},
	"daggerdbl": {
		{Kind: fair.Inflection, X: 209, Y: -59, Dir: 83, Defect: true, Reason: "the shear stretches the lower drop of the stem along its axis, and the node at its new extreme leaves the curve into the stem (r 220) bending the other way for a few units beside the drop (r 25)"},
		{Kind: fair.Jump, X: 382, Y: 135, Dir: 180, Defect: true, Reason: "the shear turns the drop at the end of the lower right arm into a leaning ellipse, whose top node, at the new extreme, joins a curve of r 61 to one of r 114"},
	},
	"sterling": {
		{Kind: fair.Jump, X: 147, Y: 13, Dir: -34, Defect: true, Reason: "the curvature of the base stroke jumps from r 631 to r 129 at a node of the roman code, which at the weights of the italic draws the base with a tighter turn than at its own"},
		{Kind: fair.Kink, X: 493, Y: 557, Dir: -79, Defect: true, Reason: "the short curve before the tip of the wedge of the ball has handles of 2 units, which the rounding turns by 7 degrees; one curve cannot replace it and its neighbour within 0.6 units"},
		{Kind: fair.Short, X: 494, Y: 551, Dir: -81, Defect: true, Reason: "that short curve, 13 units, before the tip of the wedge of the ball"},
	},
}

// Findings returns the findings of the italic profile of the smoothness check
// on the glyph which are not judged in judgedFindings or in the baseline of
// the roman parent.
func Findings(g *Glyph) []fair.Finding {
	var out []fair.Finding
	for _, f := range ProfileFindings(g.Name, g.Path) {
		ok := inherited(g, f)
		for _, e := range judgedFindings[g.Name] {
			if fair.Same(e.Kind, e.X, e.Y, e.Dir, f) {
				ok = true
			}
		}
		if !ok {
			out = append(out, f)
		}
	}
	return out
}

// ProfileFindings returns all findings of the italic profile on the path.
func ProfileFindings(name string, path outline.Path) []fair.Finding {
	a := fair.Analyse(name, path, fair.Default.Scaled(glyphs.AnalysisScale(design.Regular(), name)))
	var out []fair.Finding
	for _, f := range a.Findings {
		if f.Kind == fair.Inflection && onLongLine(path, pt(f.X, f.Y)) {
			continue
		}
		out = append(out, f)
	}
	return append(out, extremeFindings(name, path)...)
}

// onLongLine reports whether q lies within 1.5 units of a straight segment at
// least inflLine long.
func onLongLine(path outline.Path, q Point) bool {
	for _, c := range path {
		p0 := c.Start
		for _, s := range c.Segs {
			if !s.IsCurve && dist(p0, s.P3) >= inflLine && distToLine(p0, s.P3, q) <= 1.5 {
				return true
			}
			p0 = s.P3
		}
	}
	return false
}

// distToLine returns the distance from q to the segment from a to b.
func distToLine(a, b, q Point) float64 {
	d := sub(b, a)
	t := dot(sub(q, a), d) / dot(d, d)
	t = math.Max(0, math.Min(1, t))
	return dist(add(a, mul(d, t)), q)
}

// extremeFindings returns the curves of the path which run beyond their end
// points (see the profile above).
func extremeFindings(name string, path outline.Path) []fair.Finding {
	var out []fair.Finding
	for ci, c := range explicit(path) {
		n := len(c.Segs)
		start := func(i int) Point {
			i = (i + n) % n
			if i == 0 {
				return c.Start
			}
			return c.Segs[i-1].P3
		}
		seg := func(i int) (outline.Segment, Point) { i = (i + n) % n; return c.Segs[i], start(i) }
		stem := func(i int) bool {
			s, p0 := seg(i)
			d := sub(s.P3, p0)
			return !s.IsCurve && math.Abs(d.Y) >= minStem && math.Abs(d.X-0.25*d.Y) < 0.02*math.Abs(d.Y)
		}
		smooth := func(i int) bool { // the node at the start of segment i
			a, a0 := seg(i - 1)
			b, b0 := seg(i)
			u := sub(a.P3, a0)
			if a.IsCurve {
				u = sub(a.P3, a.P2)
			}
			v := sub(b.P3, b0)
			if b.IsCurve {
				v = sub(b.P1, b0)
			}
			return math.Abs(math.Atan2(cross(u, v), dot(u, v))) < 2*math.Pi/180
		}
		curve := func(i int) bool { s, _ := seg(i); return s.IsCurve }
		for i, s := range c.Segs {
			if !s.IsCurve {
				continue
			}
			p0 := start(i)
			near := stem(i-1) || stem(i+1) || curve(i-1) && smooth(i) && stem(i-2) || curve(i+1) && smooth(i+1) && stem(i+2)
			cu := cubic{p0, s.P1, s.P2, s.P3}
			for axis := 0; axis < 2; axis++ {
				get := func(q Point) float64 {
					if axis == 0 {
						return q.X
					}
					return q.Y
				}
				lo, hi := math.Min(get(cu[0]), get(cu[3])), math.Max(get(cu[0]), get(cu[3]))
				for _, t := range outline.CubicExtrema(get(cu[0]), get(cu[1]), get(cu[2]), get(cu[3])) {
					v := get(cu.at(t))
					over := math.Max(v-hi, lo-v)
					tol := extremeTol
					if axis == 0 && near {
						tol = stemExtremeTol
					}
					if over <= tol {
						continue
					}
					q := cu.at(t)
					d := cu.deriv(t)
					out = append(out, fair.Finding{Kind: extremeKind, Glyph: name, Contour: ci, X: q.X, Y: q.Y,
						Dir: math.Atan2(d.Y, d.X) * 180 / math.Pi, Size: over, Ratio: over / tol,
						Detail: fmt.Sprintf("the curve runs %.2f beyond its nodes in %s", over, []string{"x", "y"}[axis])})
				}
			}
		}
	}
	return out
}

var (
	romanOnce     sync.Once
	romanBaseline map[string][]fair.Entry
)

// inherited reports whether the finding is one which the baseline of the
// roman (bedern/internal/fair/testdata/baseline) has judged on the roman parent of
// the glyph, as intended or as a defect of the roman's construction, mapped
// to the italic by the affine part of the glyph's map.  A
// judged defect of the roman is repaired in the roman code, from which the
// italic then inherits the repair.
func inherited(g *Glyph, f fair.Finding) bool {
	name := g.Name
	romanOnce.Do(func() {
		romanBaseline = map[string][]fair.Entry{}
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			return
		}
		dir := filepath.Join(filepath.Dir(file), "..", "..", "..", "fair", "testdata", "baseline")
		for n, src := range sources {
			rn := src.roman
			if rn == "" {
				rn = n
			}
			r, err := os.Open(filepath.Join(dir, diag.FileName(rn)+".txt"))
			if err != nil {
				continue
			}
			es, err := fair.ParseBaseline(r)
			r.Close()
			if err == nil {
				romanBaseline[n] = es
			}
		}
	})
	m := g.Map
	for _, e := range romanBaseline[name] {
		if e.Reason == fair.Unjudged {
			continue
		}
		x := m[0]*e.X + m[1]*e.Y + m[2]
		s, c := math.Sincos(e.Dir * math.Pi / 180)
		dir := math.Atan2(s, m[0]*c+m[1]*s) * 180 / math.Pi
		if fair.Same(e.Kind, x, e.Y, dir, f) {
			return true
		}
	}
	return false
}
