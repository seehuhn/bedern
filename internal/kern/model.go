// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import (
	"math"
)

// ZoneOf returns the bands in which the pair a, b is seen, and the letters of
// the pair which sets the target.
//
// The zone is the band range from the baseline to the x-height for two
// lowercase letters, and from the baseline to the cap height as soon as a
// capital takes part.  The ascenders of lowercase letters do not widen the
// zone.  A pair with a punctuation mark is seen where the mark has ink (the
// dot of the full stop, the whole height of a quotation mark).
func (f *Font) ZoneOf(a, b *Glyph) (z Zone, ref string) {
	xrows := int(f.P.XHeight / f.P.Kern.Band)
	crows := int(f.P.CapHeight / f.P.Kern.Band)
	zero := f.Row(0)
	upper := a.Kind == Upper || b.Kind == Upper
	lower := a.Kind == Lower || b.Kind == Lower
	var m, o *Glyph
	switch {
	case a.Kind == Punct:
		m, o = a, b
	case b.Kind == Punct:
		m, o = b, a
	}
	top := zero + xrows
	if upper {
		top = zero + crows
	}
	switch {
	case upper && lower:
		ref = "Hn"
	case upper:
		ref = "HH"
	default:
		ref = "nn"
	}
	if m == nil {
		return Zone{zero, top}, ref
	}
	// The rows of the mark.
	lo := f.Row(math.Floor(m.Bottom/f.P.Kern.Band) * f.P.Kern.Band)
	hi := f.Row(math.Ceil(m.Top/f.P.Kern.Band) * f.P.Kern.Band)
	if m.Bottom > float64(xrows)*f.P.Kern.Band*0.8 {
		// A raised mark is seen against the whole height of the letter.
		if !upper {
			top = zero + crows
			ref = "ll"
		}
	}
	if lo < zero && (m.Bottom >= 0 || o.Bottom >= 0) {
		// The tail of a comma below a letter without a descender is left
		// out of the zone.
		lo = zero
	}
	if lo < 0 {
		lo = 0
	}
	if hi > top {
		hi = top
	}
	if lo > top {
		lo = top
	}
	return Zone{lo, hi}, ref
}

// Bands is the set of bands over which the white of a pair is measured: the
// bands Rows, and of them those from Lo up to but not including Hi for the
// narrowest place.
type Bands struct {
	Rows   []int
	Lo, Hi int
}

// shared returns the bands of the zone in which both glyphs have ink and so
// have both glyphs of the pair which sets the target, or the whole zone if
// there are none.  For two letters the narrowest place is looked for in these
// bands without those within design.Kern.Serif of their ends; a pair with a
// mark has none.
func (f *Font) shared(a, b *Glyph, z Zone) Bands {
	_, ref := f.ZoneOf(a, b)
	ta, tb := f.targetPair(a, b, ref)
	var bs Bands
	bs.Rows = make([]int, 0, max(z.Hi-z.Lo, 0))
	for i := z.Lo; i < z.Hi; i++ {
		if !math.IsNaN(a.Right[i]) && !math.IsNaN(b.Left[i]) &&
			(ta == nil || tb == nil || !math.IsNaN(ta.Right[i]) && !math.IsNaN(tb.Left[i])) {
			bs.Rows = append(bs.Rows, i)
		}
	}
	if len(bs.Rows) == 0 {
		for i := z.Lo; i < z.Hi; i++ {
			bs.Rows = append(bs.Rows, i)
		}
	}
	if a.Kind != Punct && b.Kind != Punct {
		m := int(f.P.Kern.Serif / f.P.Kern.Band)
		bs.Lo, bs.Hi = bs.Rows[0]+m, bs.Rows[len(bs.Rows)-1]+1-m
	}
	return bs
}

// white returns the white of the pair over the bands of the zone in which both
// glyphs have ink, when b is moved by k.
func (f *Font) white(a, b *Glyph, z Zone, k float64) float64 {
	return f.whiteRows(a, b, f.shared(a, b, z), k)
}

// whiteRows returns the white of the pair over the given bands when b is moved
// by k.  The white of a band is the sum of the open profiles of the two sides,
// each at most design.Kern.SideDepth, plus k, at least 0 and at most
// design.Kern.Depth; a band without ink on one side counts as
// design.Kern.Depth.  The white of the pair is the mean over the bands, and
// for two letters the weighted mean of that and the white of the narrowest
// band, with the weight design.Kern.Narrow of the narrowest band.
func (f *Font) whiteRows(a, b *Glyph, bs Bands, k float64) float64 {
	depth := f.P.Kern.Depth
	var sum float64
	least := math.Inf(1)
	for _, i := range bs.Rows {
		ra, lb := a.openRight[i], b.openLeft[i]
		w := depth
		if !math.IsNaN(ra) && !math.IsNaN(lb) {
			w = math.Max(0, math.Min(depth, ra+lb+k))
		}
		sum += w
		if i >= bs.Lo && i < bs.Hi {
			least = math.Min(least, w)
		}
	}
	mean := sum / float64(len(bs.Rows))
	if math.IsInf(least, 1) {
		return mean
	}
	nw := f.P.Kern.Narrow
	return (1-nw)*mean + nw*least
}

// glyphOf finds the glyph with the given letter.
func (f *Font) glyphOf(r rune) *Glyph {
	return f.byRune[r]
}

// Target returns the mean white which the pair a, b should have over the given
// bands.  For two capitals or two lowercase letters it is the white of the
// pair named by ref.  For a capital next to a lowercase letter it is the white
// of the lowercase letter next to H in the same order, and for a mark the
// white of the mark next to the letter named first by ref.
func (f *Font) Target(a, b *Glyph, ref string, bs Bands) float64 {
	ta, tb := f.targetPair(a, b, ref)
	return f.whiteRows(ta, tb, bs, 0)
}

// targetPair returns the glyphs of the pair whose white is the target of the
// pair a, b.
func (f *Font) targetPair(a, b *Glyph, ref string) (*Glyph, *Glyph) {
	g := func(s string) *Glyph { return f.glyphOf(rune(s[0])) }
	switch {
	case a.Kind == Punct:
		return a, g(ref[:1])
	case b.Kind == Punct:
		return g(ref[:1]), b
	case a.Kind == Upper && b.Kind == Lower:
		return g("H"), b
	case a.Kind == Lower && b.Kind == Upper:
		return a, g("H")
	}
	return g(ref[:1]), g(ref[1:])
}

// Solve returns the shift k of b for which the mean white over the bands is
// the target: the smallest k for which it is not less.
func (f *Font) Solve(a, b *Glyph, bs Bands, target float64) float64 {
	lo, hi := -600.0, 600.0
	if f.whiteRows(a, b, bs, hi) < target {
		return hi
	}
	if f.whiteRows(a, b, bs, lo) >= target {
		return lo
	}
	for i := 0; i < 24; i++ {
		mid := (lo + hi) / 2
		if f.whiteRows(a, b, bs, mid) >= target {
			hi = mid
		} else {
			lo = mid
		}
	}
	return hi
}
