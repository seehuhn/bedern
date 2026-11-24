// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

// Package fair measures the smoothness of glyph outlines.
//
// Each contour is cut at its corners into smooth runs.  Along each run the
// curvature is sampled by arc length, and the places where it misbehaves are
// reported as findings: jumps of curvature at nodes, near-corners, flats
// inside a curve, inflections, wiggles, spikes, very short segments and, from
// the thickness of the ink measured straight through the stroke, pinches and
// swellings.
//
// Every contour is first turned so that the ink lies on its left.  The
// curvature is then positive where the ink is convex (the outside of a bowl)
// and negative where it is concave (a counter, a bracket).
package fair

import (
	"fmt"
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/outline"
)

// Point is a point in font units.
type Point = outline.Point

// Kind is the kind of a finding.
type Kind string

// The kinds of finding.
const (
	Jump       Kind = "jump"       // the curvature jumps at a smooth node
	Kink       Kind = "kink"       // a node turns by a small but visible angle
	Flat       Kind = "flat"       // a nearly straight stretch between curved ones
	Inflection Kind = "inflection" // the curvature changes its sign
	Wiggle     Kind = "wiggle"     // the curvature rises, falls and rises within a short distance
	Spike      Kind = "spike"      // a short stretch curves much more than both neighbours
	Short      Kind = "short"      // a segment shorter than ShortSeg, not part of a circle
	Pinch      Kind = "pinch"      // the stroke is locally thinner than on both sides
	Swell      Kind = "swell"      // the stroke is locally thicker than on both sides

	// Extreme is reported only under a profile which sets ExtremeTol (the
	// italic): a curve runs beyond its end points, so that its extreme has no
	// node.  It is not in Kinds, so that the reports of the roman keep their
	// columns.
	Extreme Kind = "extreme"
)

// Kinds lists the kinds in the order of the reports.
var Kinds = []Kind{Jump, Kink, Flat, Inflection, Wiggle, Spike, Short, Pinch, Swell}

// Thresholds are the limits beyond which a place is a finding.  Curvatures
// are in 1/unit, lengths in font units of an em of 1000 units, angles in
// degrees.
type Thresholds struct {
	CornerDeg float64 // a node which turns by more than this is a corner, which ends a run
	KinkDeg   float64 // a node which turns by more than this, but no more than CornerDeg, is a kink

	JumpAbs   float64 // a jump needs at least this difference of curvature ...
	JumpRatio float64 // ... and this ratio of the larger to the smaller curvature (or a change of sign)

	FlatK    float64 // a stretch with |curvature| below this is flat ...
	FlatLen  float64 // ... if it is at least this long ...
	CurvedK  float64 // ... and is bounded on both sides by curvature of one sign above this ...
	FlatNear float64 // ... within this distance of its ends
	LineMax  float64 // a straight segment between curves shorter than this is a flat

	InflK float64 // curvature below this has no sign, for inflections

	WiggleLen   float64 // two maxima of curvature closer than this ...
	WiggleDepth float64 // ... with a minimum between them lower by this share ...
	WiggleAbs   float64 // ... and by at least this much, are a wiggle

	SpikeMul  float64 // a peak of curvature this many times the curvature ...
	SpikeDist float64 // ... at this distance on both sides ...
	SpikeMinK float64 // ... and above this, is a spike ...
	FilletK   float64 // ... unless it is above this, a deliberately rounded corner

	JointLen  float64 // jumps, spikes and wiggles closer than this to a corner, along the edge, belong to the joint and are not findings
	FilletLen float64 // curves between two straight segments or corners, shorter than this together, are a fillet (a bracket), judged as a joint: no jumps, spikes or wiggles

	ShortSeg float64 // a segment shorter than this is short

	PinchDist  float64 // the thickness at a place is compared with the mean at this distance on both sides ...
	PinchShare float64 // ... and a difference of more than this share is a pinch or a swelling
	ThickFind  bool    // report pinches and swellings as findings

	// The rules of the italic profile; zero leaves them out, as Default does.
	InflLine       float64 // an inflection whose two curved sides are separated by a straight segment at least this long is not a finding
	ExtremeTol     float64 // a curve which runs beyond both its end points by more than this, in x or in y, is an extreme without a node
	StemSlant      float64 // the tangent of the slant of the stems (dx/dy) ...
	StemExtremeTol float64 // ... beside which a curve may run beyond its end points in x by up to this
}

// Default are the thresholds of the gate.
var Default = Thresholds{
	CornerDeg: 15,
	KinkDeg:   1.5,

	JumpAbs:   0.005,
	JumpRatio: 1.8,

	FlatK:    0.0005,
	FlatLen:  25,
	CurvedK:  0.0025,
	FlatNear: 30,
	LineMax:  50,

	InflK: 0.0005,

	WiggleLen:   90,
	WiggleDepth: 0.4,
	WiggleAbs:   0.002,

	SpikeMul:  2.5,
	SpikeDist: 25,
	SpikeMinK: 0.006,
	FilletK:   1.0 / 12,

	FilletLen: 90,
	JointLen:  30,

	ShortSeg: 15,

	PinchDist:  25,
	PinchShare: 0.15,
	ThickFind:  false,
}

// Scaled returns the thresholds for an outline which is a copy of a
// construction drawn at its own size, scaled by s: the lengths of the windows
// are s times as long and the curvatures s times as large as in th, so that
// the analysis sees the outline as it sees the construction at its own size.
// The angles and the ratios stay.  Scaled(1) is th.
func (th Thresholds) Scaled(s float64) Thresholds {
	for _, v := range []*float64{&th.FlatLen, &th.FlatNear, &th.LineMax, &th.WiggleLen,
		&th.SpikeDist, &th.JointLen, &th.FilletLen, &th.ShortSeg, &th.PinchDist,
		&th.InflLine, &th.ExtremeTol, &th.StemExtremeTol} {
		*v *= s
	}
	for _, v := range []*float64{&th.JumpAbs, &th.FlatK, &th.CurvedK, &th.InflK,
		&th.WiggleAbs, &th.SpikeMinK, &th.FilletK} {
		*v /= s
	}
	return th
}

// Italic are the thresholds of the gate for the italic: those of Default,
// with an inflection across a straight piece of a stem allowed (the right stem
// of n runs from the convex shoulder straight into the concave exit) and the
// extremes of every curve required as nodes, except the extreme in x of a curve
// which leaves a stem slanted by 1/4, which may lie up to 2 units beyond the
// nodes.
var Italic = func() Thresholds {
	th := Default
	th.InflLine = 40
	th.ExtremeTol = 0.5
	th.StemSlant = 0.25
	th.StemExtremeTol = 2
	return th
}()

// A Finding is one place of an outline which the analysis reports.
type Finding struct {
	Kind    Kind
	Glyph   string
	Contour int
	X, Y    float64 // the place, in font units
	S       float64 // arc length from the start of the contour
	Dir     float64 // direction of the edge at the place, in degrees, with the ink on the left
	Size    float64 // the measured size, in the unit of the kind (see Detail)
	Ratio   float64 // the size divided by its threshold; above 1 is a finding
	Detail  string  // the size in words
	Number  int     // the number in the picture, from 1
}

// ID returns the identity of the finding: its kind, its place rounded to 10
// units and the compass direction of the edge there.  It does not depend on
// the contour index, the start of the contour or the rest of the glyph, so
// that it survives changes elsewhere in the glyph.  Baseline entries are
// matched with a tolerance (see Same), so that a place which moves by a unit
// or two across a rounding boundary still matches.
func (f Finding) ID() string {
	return fmt.Sprintf("%s %d %d %s", f.Kind, int(math.Round(f.X/10)*10), int(math.Round(f.Y/10)*10), Compass(f.Dir))
}

// Compass returns the direction d, in degrees, as one of eight compass points.
func Compass(d float64) string {
	names := []string{"E", "NE", "N", "NW", "W", "SW", "S", "SE"}
	i := int(math.Round(normDeg(d)/45)) % 8
	if i < 0 {
		i += 8
	}
	return names[i]
}

// normDeg brings an angle into [0, 360).
func normDeg(d float64) float64 {
	d = math.Mod(d, 360)
	if d < 0 {
		d += 360
	}
	return d
}

// angleDiff returns the absolute difference of two directions in degrees,
// from 0 to 180.
func angleDiff(a, b float64) float64 {
	d := math.Abs(normDeg(a - b))
	if d > 180 {
		d = 360 - d
	}
	return d
}

// MatchDist and MatchDeg are the tolerances of Same.
const (
	MatchDist = 15.0
	MatchDeg  = 50.0
)

// Same reports whether two findings are the same place: the same kind,
// at most MatchDist apart, on edges whose directions differ by at most
// MatchDeg (so that the two edges of a hairline are told apart).
func Same(kind Kind, x, y, dir float64, g Finding) bool {
	return kind == g.Kind && math.Hypot(x-g.X, y-g.Y) <= MatchDist && angleDiff(dir, g.Dir) <= MatchDeg
}

// Sample is a point of a contour with its arc length and curvature.
type Sample struct {
	P   Point
	S   float64 // arc length from the start of the contour
	K   float64 // signed curvature, positive where the ink is convex
	T   Point   // unit tangent, with the ink on the left
	Seg int     // index of the segment
}

// Node is the point between two segments of a contour.
type Node struct {
	P        Point
	S        float64 // arc length from the start of the contour
	Turn     float64 // turning angle from the incoming to the outgoing tangent, in degrees, positive to the left
	KIn      float64 // curvature at the end of the incoming segment
	KOut     float64 // curvature at the start of the outgoing segment
	Corner   bool    // |Turn| > CornerDeg
	InCurve  bool    // the incoming segment is a curve
	OutCurve bool    // the outgoing segment is a curve
}

// Run is a stretch of a contour between two corners, or a whole contour
// without corners.
type Run struct {
	Contour  int
	Samples  []Sample // in order, with S increasing (beyond the length of the contour where the run wraps round)
	Cyclic   bool     // the contour has no corner and the run is all of it
	Straight bool     // the run is one straight segment
	Nodes    []int    // the smooth nodes inside the run, as indices into Contour.Nodes
	Number   int      // the number in the picture, from 1
}

// Length returns the length of the run.
func (r Run) Length() float64 {
	if len(r.Samples) == 0 {
		return 0
	}
	return r.Samples[len(r.Samples)-1].S - r.Samples[0].S
}

// Thick is the thickness of the ink at a sample: from the point on the edge
// straight through the ink, along the normal, to the opposite edge.
type Thick struct {
	S     float64
	P, Q  Point // the point on the edge and the point where the ray leaves the ink
	T     float64
	Valid bool // the ray leaves the ink through an edge at a steep angle, so that T is the width of a stroke
}

// Contour is the analysis of one contour.
type Contour struct {
	Segs    []seg
	Fillet  []bool // Fillet[i] says that segment i is part of a fillet
	Samples []Sample
	Nodes   []Node // Nodes[i] is the node at the start of Segs[i]
	Length  float64
	Thick   []Thick
}

// Analysis is the analysis of one glyph.
type Analysis struct {
	Glyph    string
	Path     outline.Path // the outline, each contour turned so that the ink is on its left
	Contours []Contour
	Runs     []Run
	Findings []Finding
	Notes    []Finding // pinches and swellings, when they are not findings (Thresholds.ThickFind)
	Th       Thresholds

	combScale float64
}

// Analyse measures the outline p of the glyph called name.
func Analyse(name string, p outline.Path, th Thresholds) *Analysis {
	a := &Analysis{Glyph: name, Th: th}
	cs := orient(p)
	for _, c := range cs {
		a.Path = append(a.Path, toContour(c))
	}
	polys := flatPolys(cs)
	for ci, c := range cs {
		ca := Contour{Segs: c}
		ca.Samples, ca.Length = sampleContour(c)
		ca.Nodes = nodesOf(c, ca.Samples, th)
		ca.Fillet = fillets(c, ca.Nodes, th)
		ca.Thick = thickness(ca.Samples, polys)
		a.Contours = append(a.Contours, ca)
		a.Runs = append(a.Runs, runsOf(ci, &ca)...)
	}
	a.detect()
	sort.SliceStable(a.Findings, func(i, j int) bool {
		fi, fj := a.Findings[i], a.Findings[j]
		if fi.Contour != fj.Contour {
			return fi.Contour < fj.Contour
		}
		return fi.S < fj.S
	})
	for i := range a.Findings {
		a.Findings[i].Glyph = name
		a.Findings[i].Number = i + 1
	}
	for i := range a.Runs {
		a.Runs[i].Number = i + 1
	}
	return a
}

// Counts returns the number of findings of each kind.
func (a *Analysis) Counts() map[Kind]int {
	m := map[Kind]int{}
	for _, f := range a.Findings {
		m[f.Kind]++
	}
	return m
}
