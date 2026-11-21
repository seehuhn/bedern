// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

// Package kern computes the pair kerning of the font from the outlines of the
// glyphs.
package kern

import (
	"math"
	"sort"
	"unicode"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// Kind is the kind of a glyph for kerning.
type Kind int

const (
	// Skip glyphs are not kerned: figures, symbols, the space and glyphs
	// without a code point.
	Skip Kind = iota
	Lower
	Upper
	Punct
)

func (k Kind) String() string {
	return [...]string{"skip", "lower", "upper", "punct"}[k]
}

// ptSpacing is the largest distance between the sample points of an outline
// used for the closest distance of two glyphs.
const ptSpacing = 3.0

// ptWindow is how far from the edge of its advance a glyph keeps sample
// points.
const ptWindow = 300.0

// Glyph is the data of one glyph used for kerning.
type Glyph struct {
	Index   int // position in the glyph list
	Name    string
	Rune    rune
	Advance float64
	Kind    Kind

	// Right and Left give for each band the distance from the edge of the
	// advance to the nearest ink in the band, or NaN if the band has no ink.
	// The band i covers the heights Lo + i*Band to Lo + (i+1)*Band.
	Right, Left []float64

	// Top and Bottom are the highest and lowest ink.
	Top, Bottom float64

	rpts, lpts []outline.Point // sample points near the edges, by height
	rys, lys   []float64       // their heights
}

// Font is a set of analysed glyphs.
type Font struct {
	P      design.Params
	Glyphs []*Glyph

	lo    float64 // height of the lower edge of band 0
	nrows int
}

// Analyse computes the profiles of the glyphs.
func Analyse(p design.Params, gs []glyphs.Glyph) *Font {
	k := p.Kern
	f := &Font{P: p}
	f.lo = -math.Ceil(p.Descender/k.Band) * k.Band
	top := math.Max(p.Ascender, p.CapHeight+p.OvershootUpper)
	f.nrows = int(math.Ceil((top - f.lo) / k.Band))
	punct := make(map[rune]bool)
	for _, r := range KernPunct {
		punct[r] = true
	}
	for i, g := range gs {
		kg := &Glyph{Index: i, Name: g.Name, Rune: g.Rune, Advance: g.Advance}
		switch {
		case g.Rune == 0 || len(g.Path) == 0:
		case unicode.IsUpper(g.Rune):
			kg.Kind = Upper
		case unicode.IsLetter(g.Rune):
			kg.Kind = Lower
		case punct[g.Rune]:
			kg.Kind = Punct
		}
		if len(g.Path) > 0 {
			f.profile(kg, g.Path)
		}
		f.Glyphs = append(f.Glyphs, kg)
	}
	return f
}

// KernPunct lists the punctuation which is kerned against letters: the full
// stop, comma, colon and semicolon, the question and exclamation marks, the
// quotation marks and the apostrophe, and the hyphen and dashes.
const KernPunct = ".,:;!?'\"‘’‚“”„-–—"

// profile fills in the profiles and the sample points of one glyph.
func (f *Font) profile(g *Glyph, path outline.Path) {
	k := f.P.Kern
	polys := make([][]outline.Point, len(path))
	for i, c := range path {
		polys[i] = c.Flatten()
	}
	g.Right = make([]float64, f.nrows)
	g.Left = make([]float64, f.nrows)
	g.Bottom, g.Top = math.Inf(1), math.Inf(-1)
	for _, poly := range polys {
		for _, q := range poly {
			g.Bottom = math.Min(g.Bottom, q.Y)
			g.Top = math.Max(g.Top, q.Y)
		}
	}
	const sub = 5
	for i := 0; i < f.nrows; i++ {
		r, l := math.NaN(), math.NaN()
		for j := 0; j < sub; j++ {
			y := f.lo + (float64(i)+(float64(j)+0.5)/sub)*k.Band
			lo, hi := math.Inf(1), math.Inf(-1)
			for _, poly := range polys {
				n := len(poly)
				for a := 0; a < n; a++ {
					p, q := poly[a], poly[(a+1)%n]
					if (p.Y <= y) == (q.Y <= y) {
						continue
					}
					x := p.X + (y-p.Y)/(q.Y-p.Y)*(q.X-p.X)
					lo, hi = math.Min(lo, x), math.Max(hi, x)
				}
			}
			if lo > hi {
				continue
			}
			dr, dl := g.Advance-hi, lo
			if math.IsNaN(r) || dr < r {
				r = dr
			}
			if math.IsNaN(l) || dl < l {
				l = dl
			}
		}
		g.Right[i], g.Left[i] = r, l
	}
	defer func() {
		sort.Slice(g.rpts, func(i, j int) bool { return g.rpts[i].Y < g.rpts[j].Y })
		sort.Slice(g.lpts, func(i, j int) bool { return g.lpts[i].Y < g.lpts[j].Y })
		for _, p := range g.rpts {
			g.rys = append(g.rys, p.Y)
		}
		for _, p := range g.lpts {
			g.lys = append(g.lys, p.Y)
		}
	}()
	for _, poly := range polys {
		n := len(poly)
		for a := 0; a < n; a++ {
			p, q := poly[a], poly[(a+1)%n]
			steps := int(math.Ceil(math.Hypot(q.X-p.X, q.Y-p.Y) / ptSpacing))
			if steps < 1 {
				steps = 1
			}
			for s := 0; s < steps; s++ {
				t := float64(s) / float64(steps)
				pt := outline.Point{X: p.X + t*(q.X-p.X), Y: p.Y + t*(q.Y-p.Y)}
				if g.Advance-pt.X <= ptWindow {
					g.rpts = append(g.rpts, pt)
				}
				if pt.X <= ptWindow {
					g.lpts = append(g.lpts, pt)
				}
			}
		}
	}
}

// Row returns the index of the band which starts at height y, which must be a
// multiple of the band height.
func (f *Font) Row(y float64) int {
	return int(math.Round((y - f.lo) / f.P.Kern.Band))
}

// Rows returns the number of bands.
func (f *Font) Rows() int { return f.nrows }

// Zone is a range of bands, from Lo up to but not including Hi.
type Zone struct{ Lo, Hi int }

// Distance returns the closest distance between the outlines of a and b, with
// b set at the advance of a plus the kern.  It is the distance of the sample
// points, which are at most 3 units apart along the outlines.  It is too
// large by at most 1.5 units; it is 0 where the ink of the two overlaps
// in a band.
func (f *Font) Distance(a, b *Glyph, kern float64) float64 {
	if len(a.rpts) == 0 || len(b.lpts) == 0 {
		return math.Inf(1)
	}
	best := math.Inf(1)
	for i := range a.Right {
		if math.IsNaN(a.Right[i]) || math.IsNaN(b.Left[i]) {
			continue
		}
		s := a.Right[i] + b.Left[i] + kern
		if s <= 0 {
			return 0
		}
		best = math.Min(best, s)
	}
	off := a.Advance + kern
	// The distance of the outermost points is an upper bound.
	pa, qb := a.rpts[0], b.lpts[0]
	for _, p := range a.rpts {
		if p.X > pa.X {
			pa = p
		}
	}
	for _, q := range b.lpts {
		if q.X < qb.X {
			qb = q
		}
	}
	best = math.Min(best, math.Hypot(qb.X+off-pa.X, qb.Y-pa.Y))
	best2 := best * best
	minLeft := math.Inf(1)
	for _, x := range b.Left {
		if !math.IsNaN(x) {
			minLeft = math.Min(minLeft, x)
		}
	}
	for _, p := range a.rpts {
		// The horizontal distance is a lower bound of the distance.
		if a.Advance-p.X+minLeft+kern >= math.Sqrt(best2) {
			continue
		}
		r := math.Sqrt(best2)
		lo := sort.SearchFloat64s(b.lys, p.Y-r)
		for j := lo; j < len(b.lpts) && b.lys[j] <= p.Y+r; j++ {
			q := b.lpts[j]
			dx := q.X + off - p.X
			if dx*dx >= best2 {
				continue
			}
			dy := q.Y - p.Y
			d2 := dx*dx + dy*dy
			if d2 < best2 {
				best2 = d2
				r = math.Sqrt(best2)
			}
		}
	}
	return math.Sqrt(best2)
}
