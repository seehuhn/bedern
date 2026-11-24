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
	// without a code point which do not take the kerning of another glyph.
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

	// openRight and openLeft are the profiles as far as ink of a neighbour
	// can enter, at most design.Kern.SideDepth: an opening of a side lower than
	// design.Kern.Mouth times the x-height counts only as deep as its
	// entrance.
	openRight, openLeft []float64

	// Top and Bottom are the highest and lowest ink.
	Top, Bottom float64

	// Like is the index of the glyph whose kerning the glyph takes, or -1;
	// such a glyph is not kerned on its own (see marks.go).  Marks holds the
	// sample points of its marks, by height.
	Like  int
	marks []outline.Point

	// First and Last are, for a ligature, the indices of its first and its
	// last letter, whose left and right sides it takes for kerning
	// (ligatures.go), or -1.
	First, Last int
	mys         []float64

	mx0, mx1, my0, my1 float64 // the box of the marks
	xMin, xMax         float64 // the least and the greatest abscissa of the ink

	rpts, lpts []outline.Point // sample points near the edges, by height
	rys, lys   []float64       // their heights

	// rmax is the first of the points rpts with the greatest abscissa, lmin
	// the first of lpts with the least, and minLeft the least value of Left,
	// all as Distance needs them.
	rmax, lmin outline.Point
	minLeft    float64
}

// Font is a set of analysed glyphs.
type Font struct {
	P      design.Params
	Glyphs []*Glyph

	lo     float64 // height of the lower edge of band 0
	nrows  int
	byRune map[rune]*Glyph
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
		kg := &Glyph{Index: i, Name: g.Name, Rune: g.Rune, Advance: g.Advance, Like: -1, First: -1, Last: -1}
		switch {
		case len(g.Parts) > 0:
			kg.Kind = Lower
		case g.Rune == 0 && len(g.Path) > 0 && figureForm(g.Name) != Skip:
			kg.Kind = figureForm(g.Name)
		case g.Rune == 0 || len(g.Path) == 0:
		case signKind[g.Rune] != Skip:
			kg.Kind = signKind[g.Rune]
		case unicode.In(g.Rune, unicode.Lm, unicode.Mn):
			// Modifier letters and combining marks are not kerned.
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
		kg.marks, kg.mys = samplePoints(g.Marks)
		if len(g.Marks) > 0 {
			kg.mx0, kg.my0, kg.mx1, kg.my1 = g.Marks.BBox()
		}
		if len(g.Path) > 0 {
			kg.xMin, _, kg.xMax, _ = g.Path.BBox()
		}
		f.Glyphs = append(f.Glyphs, kg)
		if f.byRune == nil {
			f.byRune = make(map[rune]*Glyph)
		}
		if _, ok := f.byRune[kg.Rune]; !ok && kg.Rune != 0 {
			f.byRune[kg.Rune] = kg
		}
	}
	index := make(map[string]int, len(gs))
	for i, g := range gs {
		index[g.Name] = i
	}
	for i, g := range gs {
		if len(g.Parts) > 0 {
			f.Glyphs[i].First = index[g.Parts[0]]
			f.Glyphs[i].Last = index[g.Parts[len(g.Parts)-1]]
		}
		if g.Like == "" {
			continue
		}
		j, ok := index[g.Like]
		if !ok {
			continue
		}
		if g.Rune == 0 && len(g.Path) > 0 {
			// A glyph without a code point has the kind of the glyph whose
			// kerning it takes.
			f.Glyphs[i].Kind = f.Glyphs[j].Kind
		}
		if f.Glyphs[i].Kind != Skip {
			f.Glyphs[i].Like = j
		}
	}
	return f
}

// samplePoints returns points of the outline at most ptSpacing apart, sorted by
// height, and their heights.
func samplePoints(path outline.Path) ([]outline.Point, []float64) {
	var pts []outline.Point
	for _, c := range path {
		poly := c.Flatten()
		n := len(poly)
		for a := 0; a < n; a++ {
			p, q := poly[a], poly[(a+1)%n]
			steps := max(int(math.Ceil(math.Hypot(q.X-p.X, q.Y-p.Y)/ptSpacing)), 1)
			for s := 0; s < steps; s++ {
				t := float64(s) / float64(steps)
				pts = append(pts, outline.Point{X: p.X + t*(q.X-p.X), Y: p.Y + t*(q.Y-p.Y)})
			}
		}
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i].Y < pts[j].Y })
	ys := make([]float64, len(pts))
	for i, p := range pts {
		ys[i] = p.Y
	}
	return pts, ys
}

// KernPunct lists the punctuation which is kerned against letters: the full
// stop, comma, colon and semicolon, the question and exclamation marks, the
// interrobang and the inverted ones, the quotation marks, the guillemets and
// the apostrophe, the hyphens and dashes with the dashes of two thirds and of
// three quarters of an em, and the brackets, which keep only the floor.
const KernPunct = ".,:;!?¡¿‽⸘'\"‘’‚“”„«»‹›-‐‑‒–—―()[]{}〈〉⟦⟧⁅⁆"

// figureNames lists the glyph names of the figures, without a suffix.
var figureNames = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}

// figureForm returns the kind of a proportional figure, by its glyph name: the
// lining form (.prop) is kerned as a capital and the old-style form
// (.oldstyle) as a lowercase letter.  The tabular forms keep their advance and
// are not kerned, and the result for every other name is Skip.
func figureForm(name string) Kind {
	for _, d := range figureNames {
		switch name {
		case d + ".prop":
			return Upper
		case d + ".oldstyle":
			return Lower
		}
	}
	return Skip
}

// signKind holds the signs which are kerned as capitals: the ampersand, the
// at sign and the currency signs.
var signKind = map[rune]Kind{
	'&': Upper, '@': Upper, '€': Upper, '£': Upper,
}

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
	m := int(math.Round(k.Mouth * f.P.XHeight / k.Band))
	g.openRight, g.openLeft = opening(g.Right, m, k.SideDepth), opening(g.Left, m, k.SideDepth)
	defer func() {
		sort.Slice(g.rpts, func(i, j int) bool { return g.rpts[i].Y < g.rpts[j].Y })
		sort.Slice(g.lpts, func(i, j int) bool { return g.lpts[i].Y < g.lpts[j].Y })
		for _, p := range g.rpts {
			g.rys = append(g.rys, p.Y)
		}
		for _, p := range g.lpts {
			g.lys = append(g.lys, p.Y)
		}
		if len(g.rpts) > 0 {
			g.rmax = g.rpts[0]
			for _, p := range g.rpts {
				if p.X > g.rmax.X {
					g.rmax = p
				}
			}
		}
		if len(g.lpts) > 0 {
			g.lmin = g.lpts[0]
			for _, q := range g.lpts {
				if q.X < g.lmin.X {
					g.lmin = q
				}
			}
		}
		g.minLeft = math.Inf(1)
		for _, x := range g.Left {
			if !math.IsNaN(x) {
				g.minLeft = math.Min(g.minLeft, x)
			}
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

// opening returns the profile p as far as a probe of m bands can enter it from
// the side, and at most limit: in each band the largest depth which the probe
// reaches in a position which covers the band.  A band without ink is open to
// any depth, and so is the space beyond the first and the last band; it stays
// without ink in the result.
func opening(p []float64, m int, limit float64) []float64 {
	n := len(p)
	if m < 1 {
		m = 1
	}
	at := func(i int) float64 {
		if i < 0 || i >= n || math.IsNaN(p[i]) {
			return math.Inf(1)
		}
		return p[i]
	}
	// reach[j] is the depth of the probe which covers the bands j-m+1 to j.
	reach := make([]float64, n+m-1)
	for j := range reach {
		v := math.Inf(1)
		for i := j - m + 1; i <= j; i++ {
			v = math.Min(v, at(i))
		}
		reach[j] = v
	}
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		if math.IsNaN(p[i]) {
			out[i] = math.NaN()
			continue
		}
		v := 0.0
		for j := i; j < i+m; j++ {
			v = math.Max(v, reach[j])
		}
		out[i] = math.Min(limit, math.Min(v, p[i]))
	}
	return out
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
	pa, qb := a.rmax, b.lmin
	best = math.Min(best, math.Hypot(qb.X+off-pa.X, qb.Y-pa.Y))
	best2 := best * best
	r := math.Sqrt(best2)
	minLeft := b.minLeft
	for _, p := range a.rpts {
		// The horizontal distance is a lower bound of the distance.
		if a.Advance-p.X+minLeft+kern >= r {
			continue
		}
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
