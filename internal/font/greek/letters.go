// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"fmt"
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
	"seehuhn.de/go/bedern/internal/stroke"
)

// Point is a point of the plane.
type Point = stroke.Point

var auto = stroke.Auto

func pt(x, y float64) Point { return Point{X: x, Y: y} }

// n returns a node at (x, y) travelling in the direction dir (degrees),
// w thick (auto for the pen), with the curvature by rule.
func n(x, y, dir, w float64) stroke.Node {
	return stroke.Node{P: pt(x, y), Dir: dir, W: w, K: auto, DW: auto}
}

// A letter is the advance width and the strokes of a drawn letter.
type letter struct {
	adv     float64
	strokes []stroke.Stroke
}

// names holds the glyph names of the letters, those of the italic Greek except
// for μ, whose name "mu" is that of the micro sign of the Latin glyphs.
var names = map[rune]string{
	'α': "alpha", 'β': "beta", 'γ': "gamma", 'δ': "delta", 'ε': "epsilon",
	'ζ': "zeta", 'η': "eta", 'θ': "theta", 'ι': "iota", 'κ': "kappa",
	'λ': "lambda", 'μ': "muGreek", 'ν': "nu", 'ξ': "xi", 'ο': "omicron",
	'π': "pi", 'ρ': "rho", 'σ': "sigma", 'ς': "sigmaFinal", 'τ': "tau",
	'υ': "upsilon", 'φ': "phiLoop", 'χ': "chi", 'ψ': "psi", 'ω': "omega",
}

// A Letter is a built Greek letter.
type Letter struct {
	Rune     rune
	Advance  float64
	Path     outline.Path // the outline, with integer coordinates
	Warnings []string     // the warnings of the stroke tool
}

// Metrics returns the x-height, the overshoot, the ascender height and the
// descender depth (negative) which the letters are drawn to.
func Metrics(p design.Params) (x, o, asc, desc float64) {
	c := newCtx(p)
	return c.x, c.o, c.asc, c.desc
}

// Letters builds the Greek lowercase letters in the order of the alphabet.
func Letters(p design.Params) ([]Letter, error) {
	return Build(p, alphabet)
}

// Build builds the letters of the string only, which are letters of the
// alphabet.
func Build(p design.Params, only string) ([]Letter, error) {
	c := newCtx(p)
	var rs []rune
	for _, r := range alphabet {
		if draw[r] != nil && containsRune(only, r) {
			rs = append(rs, r)
		}
	}
	out := make([]Letter, len(rs))
	errs := make([]error, len(rs))
	var wg sync.WaitGroup
	for i, r := range rs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l := draw[r](c)
			for i := range l.strokes {
				if l.strokes[i].Nib == nil {
					l.strokes[i].Nib = c.pen
				}
			}
			lr, err := stroke.Letter{Strokes: l.strokes}.Build()
			if err != nil {
				errs[i] = fmt.Errorf("%c: %w", r, err)
				return
			}
			out[i] = Letter{Rune: r, Advance: l.adv, Path: lr.Path().Round()}
			for _, sr := range lr.Strokes {
				out[i].Warnings = append(out[i].Warnings, sr.Warnings...)
			}
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func containsRune(s string, r rune) bool {
	for _, q := range s {
		if q == r {
			return true
		}
	}
	return false
}

// Glyphs returns the Greek lowercase letters as glyphs.
func Glyphs(p design.Params) ([]glyphs.Glyph, error) {
	ls, err := Letters(p)
	if err != nil {
		return nil, err
	}
	gs := make([]glyphs.Glyph, len(ls))
	for i, l := range ls {
		gs[i] = glyphs.Glyph{Name: names[l.Rune], Rune: l.Rune, Advance: math.Round(l.Advance), Path: l.Path}
	}
	return gs, nil
}

// alphabet is the Greek lowercase, in its order.
const alphabet = "αβγδεζηθικλμνξοπρσςτυφχψω"

// draw holds the function which draws each letter; every letter has its own
// file, which registers it here in an init function.
var draw = map[rune]func(c *ctx) letter{}

// ctx holds the measures of Bedern Regular which the letters are drawn to,
// in font units with y up and the baseline at 0.
type ctx struct {
	p    design.Params
	pen  *stroke.Nib // the broad pen of the Greek
	x, o float64     // the x-height and the overshoot of round letters
	asc  float64     // the height of the ascenders
	desc float64     // the depth of the descenders, negative
	thin float64     // the thin of the roman
	stem float64     // a stem of the roman
}

func newCtx(p design.Params) *ctx {
	return &ctx{
		p: p, pen: penNib(p),
		x: p.XHeight, o: p.OvershootLower,
		asc: p.Ascender, desc: -p.Descender,
		thin: p.ThinLower, stem: p.Stem + 3,
	}
}

// s is a node of a straight stroke; e an end node of a straight stroke whose
// thickness does not change; ne a node whose thickness does not change; nk
// a node with the curvature k of its centre line.
func s(x, y, dir, w float64) stroke.Node     { nd := n(x, y, dir, w); nd.K = 0; return nd }
func e(x, y, dir, w float64) stroke.Node     { nd := s(x, y, dir, w); nd.DW = 0; return nd }
func ne(x, y, dir, w float64) stroke.Node    { nd := n(x, y, dir, w); nd.DW = 0; return nd }
func nk(x, y, dir, w, k float64) stroke.Node { nd := n(x, y, dir, w); nd.K = k; return nd }

// The ends of strokes: cut straight across, round, or cut and hidden inside
// another stroke.
var (
	cut    = stroke.End{Kind: stroke.Cut}
	round  = stroke.End{Kind: stroke.Round}
	hidden = stroke.End{Kind: stroke.Cut, Hidden: true}
)
