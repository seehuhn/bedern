// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"
	"strings"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The sizes and assemblies of the parentheses, the brackets, the bars and the
// floors and ceilings.  Each size is the construction of the regular run at its height,
// centred on the axis, with the strokes, the thin ends and the reach of the
// arms grown by the height ratio to the power weightPower and the bow of the
// parenthesis grown by the square root of the height ratio; no outline is
// scaled.  The side spaces are those of the base.
//
// The largest size and the ends of the assembly come from one construction:
// the parenthesis of the largest weight, cut at the middle of its height,
// where both its edges are vertical, with a straight stroke put in.  An end
// of the assembly is one half with endRun of straight stroke at its joining
// end; the largest size is the two ends joined with the full connector, so
// that the assembly with no extender is the largest size and grows from it
// without a jump.  The extender is a straight piece extender long.  The
// brackets and the bars are built in the same way from their straight
// strokes.
const (
	// weightPower is the exponent of the height ratio by which the strokes
	// grow: 1.06 times the base at 1200 and 1.32 at 3000.
	weightPower = 0.24
	// endRun is the straight stroke at the joining end of an end of an
	// assembly, extender the length of the extender, and connector the
	// length of every connector.
	endRun    = 300.0
	extender  = 300.0
	connector = 150.0
)

// delimSizes are the heights of the larger sizes, in units: \big, \Big,
// \bigg and \Bigg at 10 pt ask for 12, 18, 24 and 30 pt, and 1500, 2100 and
// 2700 lie between them for \left and \right.
var delimSizes = []float64{1200, 1500, 1800, 2100, 2400, 2700, 3000}

// A Part is a part of a vertical assembly.
type Part struct {
	Name               string
	StartConn, EndConn float64
	FullAdvance        float64
	Extender           bool
}

// sized returns the parameters of the regular with the parentheses, brackets
// and bars h high, centred on the axis, their strokes grown by the factor f
// and the bow of the parenthesis by bow.
func sized(p design.Params, h, f, bow float64) design.Params {
	q := p
	axis := p.Punct.MathAxis
	q.Punct.BracketOver = axis + h/2 - p.CapHeight
	q.Punct.BracketUnder = h/2 - axis - p.Descender
	q.Punct.ParenThick *= f
	q.Punct.ParenEnd *= f
	q.Punct.ParenSagitta *= bow
	q.Punct.BracketStem *= f
	q.Punct.BracketArm *= f
	q.Punct.BracketReach *= f
	q.Punct.BarThick *= f
	q.Punct.MathThick *= f
	q.Sym.BraceInk *= bow
	q.Sym.BraceArm *= f
	q.Sym.BraceEnd *= f
	q.Sym.BraceTip *= f
	return q
}

// centredOnAxis returns the outline of a base delimiter moved by a whole
// number of units to centre its ink on the axis to the unit.  The text font's
// delimiters stand 12 units above it, centred on the middle of the capitals
// and the descenders; every larger size is built centred on the axis.
func centredOnAxis(path outline.Path, axis float64) outline.Path {
	_, y0, _, y1 := path.Round().BBox()
	return shifted(path, 0, math.Round(axis-(y0+y1)/2))
}

// placeDelimiters moves the base delimiters among the glyphs and their
// script forms onto the axis (centredOnAxis).
func placeDelimiters(gs []Glyph, axis float64) {
	for i := range gs {
		if n := strings.TrimSuffix(strings.TrimSuffix(gs[i].Name, ".s1"), ".s2"); delimiterNames[n] {
			gs[i].Path = centredOnAxis(gs[i].Path, axis)
		}
	}
}

// baseHeight returns the height of the base delimiters.
func baseHeight(p design.Params) float64 { return p.BracketTop() - p.BracketBottom() }

// grow returns the factor of the strokes at the height h.
func grow(p design.Params, h float64) float64 { return math.Pow(h/baseHeight(p), weightPower) }

// sizedParams returns the parameters of the delimiters h high.
func sizedParams(p design.Params, h float64) design.Params {
	return sized(p, h, grow(p, h), math.Sqrt(h/baseHeight(p)))
}

// The kinds of assembly of a delimiter: none, or two ends and an extender.
// LuaTeX refuses an assembly of extenders alone ("Extensible recipe has no
// fixed parts"), so the bars have ends too, which are straight.
const (
	noAssembly = iota
	withEnds
)

// A delimiter names a delimiter whose sizes the font has.
type delimiter struct {
	base     string // the glyph of the font
	assembly int
	draw     func(r *roman) (outline.Path, float64)
}

// delimiters are the delimiters of step 4.
var delimiters = []delimiter{
	{"parenleft", withEnds, copyOf("parenleft")},
	{"parenright", withEnds, copyOf("parenright")},
	{"bracketleft", withEnds, copyOf("bracketleft")},
	{"bracketright", withEnds, copyOf("bracketright")},
	{"bar", withEnds, copyOf("bar")},
	{"uni2016", withEnds, bars(vertGap)},
	{"uni2308", noAssembly, ceilFloor(false, false)},
	{"uni2309", noAssembly, ceilFloor(true, false)},
	{"uni230A", noAssembly, ceilFloor(false, true)},
	{"uni230B", noAssembly, ceilFloor(true, true)},
	{"braceleft", withMiddle, copyOf("braceleft")},
	{"braceright", withMiddle, copyOf("braceright")},
	{"uni27E8", noAssembly, angleBracket(false)},
	{"uni27E9", noAssembly, angleBracket(true)},
}

// node returns the i-th node of the contour: the start for 0 and the end of
// segment i-1 otherwise.
func node(c outline.Contour, i int) outline.Point {
	if i == 0 {
		return c.Start
	}
	return c.Segs[i-1].P3
}

// A run is a part of a contour: its start point and its segments.
type run struct {
	start outline.Point
	segs  []outline.Segment
}

func (r run) end() outline.Point { return r.segs[len(r.segs)-1].P3 }

func (r run) shifted(dy float64) run {
	c := mapPoints(outline.Path{{Start: r.start, Segs: r.segs}}, func(q outline.Point) outline.Point {
		return outline.Point{X: q.X, Y: q.Y + dy}
	})[0]
	return run{c.Start, c.Segs}
}

// halves returns the runs of the contour above and below the line y = yc,
// between its two nodes on the line.  The contour must be closed explicitly
// and must have exactly two nodes on the line, at which both runs leave it.
func halves(c outline.Contour, yc float64) (above, below run, err error) {
	n := len(c.Segs) // the nodes are 0 to n, node n equals node 0
	var at []int
	for i := 0; i < n; i++ {
		if math.Abs(node(c, i).Y-yc) < 1e-6 {
			at = append(at, i)
		}
	}
	if len(at) != 2 {
		return run{}, run{}, fmt.Errorf("the contour has %d nodes at %v, want 2", len(at), yc)
	}
	part := func(from, to int) run {
		r := run{start: node(c, from)}
		for k := from; k != to; k = (k + 1) % n {
			r.segs = append(r.segs, c.Segs[k])
		}
		return r
	}
	a, b := part(at[0], at[1]), part(at[1], at[0])
	mid := func(r run) float64 { return (r.start.Y + r.segs[0].P3.Y) / 2 }
	if mid(a) < yc {
		a, b = b, a
	}
	return a, b, nil
}

// stretched returns the contour with a straight piece s long put in at the
// line y = yc, the part above moved up by s/2 and the part below down.
func stretched(c outline.Contour, yc, s float64) (outline.Contour, error) {
	above, below, err := halves(c, yc)
	if err != nil {
		return outline.Contour{}, err
	}
	a, b := above.shifted(s/2), below.shifted(-s/2)
	out := outline.Contour{Start: a.start}
	out.Segs = append(out.Segs, a.segs...)
	out.Segs = append(out.Segs, outline.Segment{P3: b.start})
	out.Segs = append(out.Segs, b.segs...)
	out.Segs = append(out.Segs, outline.Segment{P3: a.start})
	return out, nil
}

// end returns the upper or the lower half of the contour, cut at y = yc,
// with a straight piece run long put in at the cut, moved to put its
// bottom at 0.
func end(c outline.Contour, yc, length float64, upper bool) (outline.Path, error) {
	above, below, err := halves(c, yc)
	if err != nil {
		return nil, err
	}
	var out outline.Contour
	if upper {
		a := above.shifted(length - yc)
		out = outline.Contour{Start: a.start, Segs: append([]outline.Segment(nil), a.segs...)}
		e := a.end()
		out.Segs = append(out.Segs,
			outline.Segment{P3: outline.Point{X: e.X, Y: 0}},
			outline.Segment{P3: outline.Point{X: a.start.X, Y: 0}},
			outline.Segment{P3: a.start})
		return outline.Path{out}, nil
	}
	b := below.shifted(-length - yc)
	out = outline.Contour{Start: b.start, Segs: append([]outline.Segment(nil), b.segs...)}
	e := b.end()
	out.Segs = append(out.Segs,
		outline.Segment{P3: outline.Point{X: e.X, Y: 0}},
		outline.Segment{P3: outline.Point{X: b.start.X, Y: 0}},
		outline.Segment{P3: b.start})
	_, y0, _, _ := outline.Path{out}.BBox()
	return shifted(outline.Path{out}, 0, -y0), nil
}

// delimGlyphs returns the sizes and the parts of the assemblies of the
// delimiters, and for each base the names of its sizes and its parts.
func delimGlyphs(p design.Params) (gs []Glyph, variants map[string][]string, parts map[string][]Part, err error) {
	variants = map[string][]string{}
	parts = map[string][]Part{}
	axis := p.Punct.MathAxis
	largest := delimSizes[len(delimSizes)-1]
	for _, d := range delimiters {
		for k, h := range delimSizes {
			q := sizedParams(p, h)
			var path outline.Path
			var adv float64
			if d.assembly == withMiddle && k == len(delimSizes)-1 {
				path, adv, err = braceLargest(p, d)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("%s: %w", d.base, err)
				}
			} else if d.assembly != noAssembly && k == len(delimSizes)-1 {
				// The largest size: the construction 2 endRun - connector
				// lower, stretched at the axis.
				s := 2*endRun - connector
				q = sized(p, h-s, grow(p, h), math.Sqrt(h/baseHeight(p)))
				path, adv = d.draw(newRoman(q))
				path, err = stretchPath(path, axis, s)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("%s: %w", d.base, err)
				}
			} else {
				path, adv = d.draw(newRoman(q))
			}
			name := fmt.Sprintf("%s.v%d", d.base, k+1)
			gs = append(gs, Glyph{Name: name, Advance: math.Round(adv), Path: path, Class: Fence, Like: likeOf(d.base)})
			variants[d.base] = append(variants[d.base], name)
		}
		if d.assembly == noAssembly {
			continue
		}
		if d.assembly == withMiddle {
			bg, ps, err := braceParts(p, d)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("%s: %w", d.base, err)
			}
			gs = append(gs, bg...)
			parts[d.base] = ps
			continue
		}
		// The ends and the extender, from the construction of the largest
		// size before it is stretched.
		s := 2*endRun - connector
		q := sized(p, largest-s, grow(p, largest), math.Sqrt(largest/baseHeight(p)))
		path, adv := d.draw(newRoman(q))
		adv = math.Round(adv)
		top, bot, ext, err := assemblyParts(path, axis, d.assembly == withEnds, endRun, extender)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s: %w", d.base, err)
		}
		var ps []Part
		if bot != nil {
			_, _, _, h := bot.BBox()
			gs = append(gs, Glyph{Name: d.base + ".bot", Advance: adv, Path: bot, Class: Fence, Like: likeOf(d.base)})
			ps = append(ps, Part{Name: d.base + ".bot", EndConn: connector, FullAdvance: h})
		}
		gs = append(gs, Glyph{Name: d.base + ".ext", Advance: adv, Path: ext, Class: Fence, Like: likeOf(d.base)})
		ps = append(ps, Part{Name: d.base + ".ext", StartConn: connector, EndConn: connector, FullAdvance: extender, Extender: true})
		if top != nil {
			_, _, _, h := top.BBox()
			gs = append(gs, Glyph{Name: d.base + ".top", Advance: adv, Path: top, Class: Fence, Like: likeOf(d.base)})
			ps = append(ps, Part{Name: d.base + ".top", StartConn: connector, FullAdvance: h})
		}
		parts[d.base] = ps
	}
	return gs, variants, parts, nil
}

// likeOf returns the glyph of the regular whose bounding-box rule applies to
// the sizes of a delimiter.
func likeOf(base string) string {
	switch base {
	case "uni2016":
		return "bar"
	case "uni2308", "uni230A":
		return "bracketleft"
	case "uni2309", "uni230B":
		return "bracketright"
	case "uni27E8":
		return "parenleft"
	case "uni27E9":
		return "parenright"
	}
	return base
}

// withNodesAt returns the contour, closed explicitly, with a node put in
// where a straight segment crosses the line y = yc.
func withNodesAt(c outline.Contour, yc float64) outline.Contour {
	if n := len(c.Segs); n > 0 && c.Segs[n-1].P3 != c.Start {
		c.Segs = append(c.Segs[:n:n], outline.Segment{P3: c.Start})
	}
	out := outline.Contour{Start: c.Start}
	cur := c.Start
	for _, sg := range c.Segs {
		a, b := cur, sg.P3
		if !sg.IsCurve && (a.Y-yc)*(b.Y-yc) < 0 {
			t := (yc - a.Y) / (b.Y - a.Y)
			out.Segs = append(out.Segs, outline.Segment{P3: outline.Point{X: a.X + t*(b.X-a.X), Y: yc}})
		}
		out.Segs = append(out.Segs, sg)
		cur = sg.P3
	}
	return out
}

// stretchPath stretches every contour of the path at y = yc.
func stretchPath(path outline.Path, yc, s float64) (outline.Path, error) {
	out := make(outline.Path, len(path))
	for i, c := range path {
		st, err := stretched(withNodesAt(c, yc), yc, s)
		if err != nil {
			return nil, err
		}
		out[i] = st
	}
	return out, nil
}

// assemblyParts returns the ends and the extender of an assembly from the
// construction of the largest size: the halves above and below the axis with
// run of straight stroke at the cut, and the extender, the stroke at the
// axis length long.  If ends is false the extender alone is returned, and
// top and bot are nil.
func assemblyParts(path outline.Path, axis float64, ends bool, run, length float64) (top, bot, ext outline.Path, err error) {
	for _, c := range path {
		c = withNodesAt(c, axis)
		above, below, err := halves(c, axis)
		if err != nil {
			return nil, nil, nil, err
		}
		lo, hi := math.Min(above.start.X, below.start.X), math.Max(above.start.X, below.start.X)
		ext = append(ext, box(lo, 0, hi, length))
		if !ends {
			continue
		}
		u, err := end(c, axis, run, true)
		if err != nil {
			return nil, nil, nil, err
		}
		l, err := end(c, axis, run, false)
		if err != nil {
			return nil, nil, nil, err
		}
		top = append(top, u...)
		bot = append(bot, l...)
	}
	return top, bot, ext, nil
}
