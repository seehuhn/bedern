// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"fmt"
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/outline"
	"seehuhn.de/go/bedern/internal/stroke"
)

// A letter is the upright description of a glyph: its strokes, its dots
// (circles which keep their form under the shear and move with their
// centre) and its set width.
type letter struct {
	strokes []stroke.Stroke
	dots    []dot
	advance float64
}

// A dot is a circle of the diameter D about C, upright.
type dot struct {
	C Point
	D float64
}

// A Glyph is a letter of the italic as outline.
type Glyph struct {
	Name    string
	Rune    rune
	Advance float64
	Path    outline.Path // the outline, rounded to whole units

	Strokes  []stroke.Stroke // the strokes, sheared and moved
	Nodes    int             // the number of nodes of the strokes
	Result   *stroke.LetterResult
	Warnings []string       // the warnings of the stroke tool
	Findings []fair.Finding // the findings of the italic profile of the smoothness check which are not marked as intended
}

// letters are the drawn letters, by name.
var letters = map[string]struct {
	r    rune
	draw func(Params) letter
}{
	"n": {'n', n},
	"o": {'o', o},
	"i": {'i', i},
	"u": {'u', u},
	"m": {'m', m},
	"h": {'h', h},
	"l": {'l', l},
	"t": {'t', t},
	"r": {'r', r},
	"c": {'c', c},
	"e": {'e', e},
	"a": {'a', a},
	"q": {'q', q},
	"d": {'d', d},
	"g": {'g', g},
	"b": {'b', b},
	"p": {'p', p},
	"s": {'s', s},
	"j": {'j', j},
	"y": {'y', y},
	"f": {'f', f},
	"v": {'v', v},
	"w": {'w', w},
	"x": {'x', x},
	"z": {'z', z},
	"k": {'k', k},

	"germandbls": {'ß', eszett},
	"oe":         {'œ', oe},
	"ae":         {'æ', ae},

	"longs":        {'ſ', longs},
	"kgreenlandic": {'ĸ', kgreenlandic},
	"thorn":        {'þ', thorn},
	"ij":           {'ĳ', ij},
	"ldot":         {'ŀ', ldot},
	"eng":          {'ŋ', eng},

	"eth":     {'ð', eth},
	"uni0259": {'ə', schwa},
	"micro":   {'µ', func(p Params) letter { return roundAdvance(micro(p)) }},
}

// roundAdvance returns the letter with its advance rounded to whole units.
func roundAdvance(l letter) letter {
	l.advance = math.Round(l.advance)
	return l
}

// GlyphName returns the name which the font gives the glyph of the letter of
// the given name: the name of the letter, except for µ, whose name in the
// font is that of the roman's (the letter of the Greek lowercase has the
// name mu here).
func GlyphName(name string) string {
	if name == "micro" {
		return "mu"
	}
	return name
}

// Names returns the names of the drawn letters, sorted.
func Names() []string {
	var out []string
	for n := range letters {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Build returns the glyph of the given name.
func Build(p Params, name string) (*Glyph, error) {
	d, ok := letters[name]
	if !ok {
		d, ok = greekLetters[name]
	}
	if !ok {
		return nil, fmt.Errorf("italic: no letter %q", name)
	}
	return buildLetter(p, name, d.r, d.draw(p))
}

// buildLetter returns the glyph of the letter, sheared by the slant.
func buildLetter(p Params, name string, r rune, l letter) (*Glyph, error) {
	k := p.slant()
	dx := -p.shift()
	var sl stroke.Letter
	g := &Glyph{Name: name, Rune: r, Advance: l.advance}
	for _, s := range l.strokes {
		s = moved(s, dx).ShearedRatio(p.SlantNum, p.SlantDen)
		sl.Strokes = append(sl.Strokes, s)
		g.Nodes += len(s.Nodes)
	}
	for _, dt := range l.dots {
		sl.Extra = append(sl.Extra, stroke.Disc(pt(dt.C.X+dx+k*dt.C.Y, dt.C.Y), dt.D/2))
	}
	res, err := sl.Build()
	if err != nil {
		return nil, fmt.Errorf("italic %s: %w", name, err)
	}
	g.Strokes, g.Result = sl.Strokes, res
	g.Path = res.Path().Round()
	g.Warnings = res.Warnings
	hooks := p.entryHooks(l, 1)
	for _, f := range fair.Analyse(name, g.Path, fair.Italic).Findings {
		if intendedFinding(name, f) == "" && !hookSpike(hooks, f) {
			g.Findings = append(g.Findings, f)
		}
	}
	return g, nil
}

// entrySpikeMax is the largest spike under the top of an entry hook which
// belongs to its form, over the threshold of the smoothness check: the
// inner edge of the hook turns over its top with a radius of 16 to 20
// against about 50 at 25 units on either side, 2.4 to 3.2 times as tight
// over the sweeps of the stem, the x-height and the contrast, and 2.7 times
// in the unrounded outline of the default parameters, which the rounding
// brings to 2.3.
const entrySpikeMax = 1.3

// entryHooks returns the places, in the italic, of the inner edge under the
// top node of every entry hook of the letter, for a letter widened
// horizontally by width before the shear.  An entry is the pair of nodes
// which p.entry starts with: the start, EndW thick in the direction
// EntryStartDir, and the top, EntryTopW thick and running horizontally.
func (p Params) entryHooks(l letter, width float64) []Point {
	var out []Point
	k := p.slant()
	for _, st := range l.strokes {
		ns := st.Nodes
		for i := 0; i+1 < len(ns); i++ {
			a, b := ns[i], ns[i+1]
			if a.W != p.EndW || a.Dir != p.EntryStartDir || b.Dir != 0 || b.W != p.EntryTopW {
				continue
			}
			q := pt(b.P.X*width, b.P.Y-b.W/2)
			out = append(out, pt(q.X-p.shift()+k*q.Y, q.Y))
		}
	}
	return out
}

// hookSpike reports whether the finding is the spike under the top of one
// of the entry hooks, within entrySpikeMax of the threshold: a tight round
// which is the form of the hook, where the rounding decides whether the gate
// sees it.
func hookSpike(hooks []Point, f fair.Finding) bool {
	if f.Kind != fair.Spike || f.Ratio > entrySpikeMax {
		return false
	}
	for _, h := range hooks {
		if math.Hypot(f.X-h.X, f.Y-h.Y) <= fair.MatchDist {
			return true
		}
	}
	return false
}

// moved returns the stroke moved by dx.
func moved(s stroke.Stroke, dx float64) stroke.Stroke {
	out := s
	out.Nodes = append([]stroke.Node(nil), s.Nodes...)
	for i := range out.Nodes {
		out.Nodes[i].P.X += dx
	}
	mv := func(e stroke.End) stroke.End {
		e.Q.X += dx
		e.Ball.Centre.X += dx
		e.Ball.InkX += dx
		return e
	}
	out.Start, out.End = mv(s.Start), mv(s.End)
	return out
}

// An intended entry is a finding of the smoothness check which belongs to the
// form, with its reason.
type intended struct {
	Kind   fair.Kind
	X, Y   float64
	Dir    float64
	Reason string
}

// intendedFindings are the findings marked as intended, by glyph.
var intendedFindings = map[string][]intended{}

// intendedFinding returns the reason if the finding is marked as intended.
func intendedFinding(name string, f fair.Finding) string {
	for _, e := range intendedFindings[name] {
		if fair.Same(e.Kind, e.X, e.Y, e.Dir, f) {
			return e.Reason
		}
	}
	return ""
}

// unshear returns the point of the italic in the upright description.
func (p Params) unshear(q Point) Point {
	return pt(q.X-p.slant()*q.Y+p.shift(), q.Y)
}

// degrees converts radians to degrees.
func degrees(a float64) float64 { return a * 180 / math.Pi }
