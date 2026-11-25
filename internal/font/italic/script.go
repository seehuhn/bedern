// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/stroke"
)

// A Script gives the changes of a letter drawn for a smaller size: it is
// widened horizontally, a letter by its parameters and a symbol by its
// centre lines in the upright description, and it is drawn with a heavier
// pen whose thinnest strokes have a floor.  Heights do not change.
type Script struct {
	Width float64 // the factor of the widening
	Stem  float64 // the factor of the thickness of the stems
	Thin  float64 // the thickness of the thinnest strokes, at least
	// NoFindings leaves out the smoothness check of the outline, which
	// takes most of the time of a build.
	NoFindings bool
	// Level is the script level, 1 or 2, whose tuned parameters ParamsFor
	// applies (script_tuned.go); 0 leaves the parameters as Params gives them.
	Level int
	// Fixed, when set, gives the parameters of the form in place of those
	// which Params derives.
	Fixed *Params
}

// thick returns the thickness of a stroke drawn with the pen of s which has
// the thickness w with the pen of p: linear in w through the thin of the pen
// and the stem, both at right angles, and in proportion below the thin.
func (s Script) thick(p Params, w float64) float64 {
	t0, s0 := p.Thin, p.across(p.Stem)
	t1, s1 := math.Max(p.Thin*s.Stem, s.Thin), s0*s.Stem
	if w <= t0 {
		return w * t1 / t0
	}
	return t1 + (w-t0)*(s1-t1)/(s0-t0)
}

// BuildScript returns the glyph of the given name (a letter or a symbol)
// drawn as a script form by s: a letter with the parameters of ParamsFor, a
// symbol with its centre lines, ends, balls and dots widened horizontally by
// s.Width before the shear, and with every thickness of the pen and the
// strokes mapped by s.
func BuildScript(p Params, name string, s Script) (*Glyph, error) {
	d, ok := letters[name]
	if !ok {
		d, ok = greekLetters[name]
	}
	if !ok {
		// The symbols are drawn as BuildSymbol draws them.
		sym, found := symbols[name]
		if !found {
			return nil, fmt.Errorf("italic: no letter %q", name)
		}
		d.r, d.draw = sym.r, sym.draw
		if sym.upright {
			p.SlantNum = 0
		}
		if name != "uni2113" {
			p.Stress = 0
		}
	}
	q := s.ParamsFor(p, name)
	if s.Fixed != nil {
		q = *s.Fixed
	}
	if isLatin(name) || isGreek(name) {
		// The letters are widened by their parameters.
		s.Width = 1
	}
	l := d.draw(q)
	k := p.slant()
	dx := -p.shift()
	var sl stroke.Letter
	g := &Glyph{Name: name, Rune: d.r, Advance: l.advance * s.Width}
	for _, st := range l.strokes {
		st = moved(s.stroke(p, st), dx).ShearedRatio(p.SlantNum, p.SlantDen)
		sl.Strokes = append(sl.Strokes, st)
		g.Nodes += len(st.Nodes)
	}
	for _, dt := range l.dots {
		c := pt(dt.C.X*s.Width, dt.C.Y)
		sl.Extra = append(sl.Extra, stroke.Disc(pt(c.X+dx+k*c.Y, c.Y), dt.D/2))
	}
	r, err := sl.Build()
	if err != nil {
		return nil, fmt.Errorf("italic %s: %w", name, err)
	}
	g.Strokes, g.Result = sl.Strokes, r
	g.Path = r.Path().Round()
	g.Warnings = r.Warnings
	if s.NoFindings {
		return g, nil
	}
	// The findings marked as intended in a script form are those of its
	// level (script_intended.go): the places of a form drawn with another
	// pen and widened are not those of the text form.
	key := name
	if s.Level > 0 {
		key = fmt.Sprintf("%s/%d", name, s.Level)
	}
	hooks := p.entryHooks(l, s.Width)
	for _, f := range fair.Analyse(name, g.Path, fair.Italic).Findings {
		if intendedFinding(key, f) == "" && !hookSpike(hooks, f) && !atHairpin(g.Strokes, f) {
			g.Findings = append(g.Findings, f)
		}
	}
	return g, nil
}

// stroke returns the upright stroke widened and drawn with the pen of s.
func (s Script) stroke(p Params, st stroke.Stroke) stroke.Stroke {
	f := s.Width
	out := st
	// dir returns the direction a and the length g of the unit vector of the
	// direction a after the widening.
	dir := func(a float64) (float64, float64) {
		r := a * math.Pi / 180
		x, y := f*math.Cos(r), math.Sin(r)
		return math.Atan2(y, x) * 180 / math.Pi, math.Hypot(x, y)
	}
	out.Nodes = make([]stroke.Node, len(st.Nodes))
	for i, nd := range st.Nodes {
		a, g := dir(nd.Dir)
		// The letters are widened by their parameters (f is 1) and keep
		// their nodes as the text forms do; the nodes of a symbol widened
		// by its centre lines are put on whole units.
		if f != 1 {
			nd.P.X = math.Round(nd.P.X * f)
		}
		nd.Dir = a
		if !math.IsNaN(nd.K) {
			nd.K *= f / (g * g * g)
		}
		if nd.Edge != nil {
			nd.Edge = &stroke.EdgeK{Right: nd.Edge.Right * f / (g * g * g), Left: nd.Edge.Left * f / (g * g * g)}
		}
		slope := 1.0
		if !math.IsNaN(nd.DW) {
			nd.DW *= slope / g
		}
		nd.DDW *= slope / (g * g)
		out.Nodes[i] = nd
	}
	end := func(e stroke.End) stroke.End {
		if e.Kind == stroke.Clip || e.Line {
			e.Angle, _ = dir(e.Angle)
		}
		e.Q.X *= f
		if e.Kind == stroke.Ball {
			e.Ball.Centre.X *= f
			e.Ball.InkX *= f

		}
		return e
	}
	out.Start, out.End = end(st.Start), end(st.End)
	return out
}

// Params returns the parameters of the script form: the stem and the pen of
// WithStem with the stem of s, and every thickness of the pen and the strokes
// which belong to the pen mapped by thick, so that the thinnest strokes keep
// their floor.
func (s Script) Params(p Params) Params {
	q := p.WithStem(math.Round(p.Stem * s.Stem))
	q.Curve = s.thick(p, p.Curve)
	q.Thin = s.thick(p, p.Thin)
	q.Hair = s.thick(p, p.Hair)
	q.ArchStartW = s.thick(p, p.ArchStartW)
	q.URiseW = s.thick(p, p.URiseW)
	q.EntryTopW = s.thick(p, p.EntryTopW)
	q.EndW = s.thick(p, p.EndW)
	q.ExitEndW = s.thick(p, p.ExitEndW)
	q.ShoulderW = s.thick(p, p.ShoulderW)
	q.ExitLowW = s.thick(p, p.ExitLowW)
	q.Ball = s.thick(p, p.Ball)
	q.BallStalk = s.thick(p, p.BallStalk)
	q.Dot = s.thick(p, p.Dot)
	// The turn into the exit is a turn of the heavier stem: it starts higher
	// and runs vertically higher, by the share of the heavier stem.
	lat := func(v float64) float64 { return math.Round(v/4) * 4 }
	d := q.Stem - p.Stem
	q.ExitTurn += lat(0.9 * d)
	q.ExitVert += lat(d / 4)
	q.ExitLowShift = p.ExitLowShift + 2*math.Round(d/8)
	// The bottom of the bowl of a, d, g and q moves to the right by the
	// growth of the stem.
	q.BowlBottom += d
	// The hook of χ turns into its heavy stroke as the stem turns into the
	// exit: the straight run begins lower and further right, by the share
	// of the heavier stem.
	q.Gk.ChiHigh = lat(q.Gk.ChiHigh - 2*d)
	q.Gk.ChiRun += math.Round(2 * d / 3)
	// The loop of γ grows on both sides by half the growth of the stem, so
	// that its counter keeps its size.
	q.Gk.GammaLoopR += math.Round(d / 2)
	q.Gk.GammaLoopL += math.Round(d / 2)
	return q
}

// isLatin says whether name is a letter of the Latin lowercase or one of the
// other letters of the italic.  The script forms of the Latin and the Greek
// letters are widened by their parameters, those of the symbols by their
// centre lines.
func isLatin(name string) bool {
	_, ok := letters[name]
	return ok
}

// isGreek says whether name is a Greek letter.
func isGreek(name string) bool {
	_, ok := greekLetters[name]
	return ok
}

// ParamsFor returns the parameters of the letter name drawn by s: Params, for a
// Latin or a Greek letter the parameters widened by s.Width (the symbols are
// widened by their centre lines) and the tuned parameters of the level.
func (s Script) ParamsFor(p Params, name string) Params {
	q := s.Params(p)
	if !isLatin(name) && !isGreek(name) {
		return q
	}
	q = q.widen(s.Width, p.Stem)
	switch s.Level {
	case 1:
		q.tuneLatin1()
	case 2:
		q.tuneLatin2()
	}
	if isGreek(name) {
		// The Greek letters take the parameters of the level of the Latin
		// letters for the parts which they share with them (the entry, the
		// exit, the arch, the ball, the leg of k), and their own on top.
		switch s.Level {
		case 1:
			q.tuneGreek1()
		case 2:
			q.tuneGreek2()
		}
	}
	return q
}
