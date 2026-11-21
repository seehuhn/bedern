// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import (
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// Pair is the kern of one ordered pair of glyphs.
type Pair struct {
	Left, Right *Glyph
	Kern        int // the kern as written to the font

	Zone   Zone
	Ref    string  // the pair which sets the target
	Target float64 // the mean white the pair should have
	Allow  float64 // the allowance added to the target
	White0 float64 // the mean white without the kern
	White1 float64 // the mean white with the kern
	Model  float64 // the shift of the model before the limits
	Dist0  float64 // the closest distance of the outlines without the kern
	Dist1  float64 // and with it
	Floor  bool    // the floor limits the kern, or makes it positive
	Reason string  // the reason of the exception which sets the kern, if any

	LClass, RClass int // classes of the two sides, 0 if the glyph is in none
}

// Table is the result of the computation: the kern of every pair which has
// one, and the classes with the compact form of the kerns.
type Table struct {
	Font  *Font
	Pairs []*Pair // pairs with a kern, in the order of the glyph list

	// The compact form.  Class1 and Class2 map glyph indices to the class of
	// their right and left side (the first and the second glyph of a pair);
	// the glyphs not in the map have class 0, which has no kerns.  Values
	// holds the kern for each pair of classes, indexed by Class1 and then
	// Class2 (row 0 and column 0 are zero), and Single the pairs whose kern
	// differs from that of their classes, which are written first.
	Class1, Class2 map[int]int
	Values         [][]int
	Single         map[[2]int]int

	// NClass1 and NClass2 are the numbers of classes of the right and left
	// sides over all kernable glyphs, with or without kerns.
	NClass1, NClass2 int
}

// Compute computes the kerns of the glyphs.  The glyph index is the position
// in gs.
func Compute(p design.Params, gs []glyphs.Glyph) *Table {
	f := Analyse(p, gs)
	k := p.Kern
	t := &Table{Font: f}
	for _, r := range "Hln" {
		if f.glyphOf(r) == nil {
			return t
		}
	}
	var kernable []*Glyph
	for _, g := range f.Glyphs {
		if g.Kind != Skip {
			kernable = append(kernable, g)
		}
	}
	exc := make(map[[2]string]design.KernException)
	for _, e := range design.KernExceptions {
		exc[[2]string{e.Left, e.Right}] = e
	}
	type tkey struct {
		ref string
		z   Zone
	}
	targets := make(map[tkey]float64)

	all := make(map[[2]int]*Pair)
	for _, a := range kernable {
		for _, b := range kernable {
			if a.Kind == Punct && b.Kind == Punct {
				continue
			}
			if !f.markPair(a, b) {
				continue
			}
			z, ref := f.ZoneOf(a, b)
			if z.Hi <= z.Lo {
				continue
			}
			var target float64
			if a.Kind == Punct || b.Kind == Punct {
				// The target of a mark is the white of the mark next to the
				// letter of ref which is not the mark.
				r := f.glyphOf(rune(ref[0]))
				if a.Kind == Punct {
					target = f.white(a, r, z, 0)
				} else {
					target = f.white(r, b, z, 0)
				}
			} else {
				key := tkey{ref, z}
				var ok bool
				target, ok = targets[key]
				if !ok {
					target = f.Target(ref, z)
					targets[key] = target
				}
			}
			pr := &Pair{Left: a, Right: b, Zone: z, Ref: ref, Target: target}
			pr.White0 = f.white(a, b, z, 0)
			allow := k.Allowance
			if a.Kind != Punct && b.Kind != Punct && a.Kind != b.Kind {
				allow = k.AllowanceMixed
			}
			pr.Allow = allow
			pr.Model = f.Solve(a, b, z, target+allow)
			v := 0.0
			if pr.Model < 0 && f.interplay(a, b, z) {
				v = math.Max(k.Soft*math.Tanh(pr.Model/k.Soft), -k.MaxNegative)
			}
			pr.Dist0 = f.Distance(a, b, 0)
			if e, ok := exc[[2]string{a.Name, b.Name}]; ok {
				v = e.Value
				pr.Reason = e.Reason
			}
			kk := int(math.Round(v))
			if pr.Reason == "" && math.Abs(float64(kk)) < k.Threshold {
				kk = 0
			}
			if kk < 0 && pr.Reason == "" {
				kk = f.limitApproach(a, b, z, ref, kk)
				if math.Abs(float64(kk)) < k.Threshold {
					kk = 0
				}
			}
			kk, pr.Floor = f.applyFloor(a, b, kk, pr.Dist0)
			pr.Kern = kk
			pr.White1 = f.white(a, b, z, float64(kk))
			pr.Dist1 = pr.Dist0
			if kk != 0 {
				pr.Dist1 = f.Distance(a, b, float64(kk))
			}
			all[[2]int{a.Index, b.Index}] = pr
		}
	}

	t.classes(kernable, all)
	t.compact(all)
	for _, a := range kernable {
		for _, b := range kernable {
			if pr := all[[2]int{a.Index, b.Index}]; pr != nil && pr.Kern != 0 {
				t.Pairs = append(t.Pairs, pr)
			}
		}
	}
	return t
}

// markPair says whether the pair is kerned at all: a mark which stands above
// the baseline (the quotation marks, the hyphen and the dashes) is kerned
// against capitals only.
func (f *Font) markPair(a, b *Glyph) bool {
	for _, p := range [2][2]*Glyph{{a, b}, {b, a}} {
		if m, o := p[0], p[1]; m.Kind == Punct && o.Kind == Lower && m.Bottom >= 0.3*f.P.XHeight {
			return false
		}
	}
	return true
}

// interplay says whether the white responds to a kern in enough bands of the
// zone, at least the fraction design.Kern.Interplay of them: bands in which
// both glyphs have ink and the white is less than the cap.
func (f *Font) interplay(a, b *Glyph, z Zone) bool {
	n := 0
	for i := z.Lo; i < z.Hi; i++ {
		if !math.IsNaN(a.Right[i]) && !math.IsNaN(b.Left[i]) && a.Right[i]+b.Left[i] < f.P.Kern.Depth+f.P.Kern.MaxNegative {
			n++
		}
	}
	return float64(n) >= f.P.Kern.Interplay*float64(z.Hi-z.Lo)
}

// applyFloor changes the kern kk, so that the glyphs keep the floor distance:
// a negative kern which would bring them closer is reduced, and a pair which is
// closer than the floor without a kern gets a positive one.  A kern which the
// floor makes positive is not dropped below the threshold.  The result says
// whether the floor changed the kern.
func (f *Font) applyFloor(a, b *Glyph, kk int, d0 float64) (int, bool) {
	floor := f.P.Kern.Floor
	if d0 < floor {
		// The pair needs room: the smallest kern which clears the floor.
		lo, hi := 0, 400
		if f.Distance(a, b, float64(hi)) < floor {
			return hi, true
		}
		for lo < hi {
			mid := lo + (hi-lo)/2
			if f.Distance(a, b, float64(mid)) >= floor {
				hi = mid
			} else {
				lo = mid + 1
			}
		}
		if lo > kk {
			return lo, true
		}
		return kk, false
	}
	if kk >= 0 || f.Distance(a, b, float64(kk)) >= floor {
		return kk, false
	}
	// The most negative kern which keeps the floor.
	lo, hi := kk, 0
	for lo < hi {
		mid := lo + (hi-lo)/2
		if f.Distance(a, b, float64(mid)) >= floor {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	if math.Abs(float64(lo)) < f.P.Kern.Threshold {
		lo = 0
	}
	return lo, true
}

// profileZone returns the bands over which the glyph side is compared with
// others for the classes: from the descender to the top of the zone of the
// glyph, the x-height for lowercase letters and the cap height for the rest.
func (f *Font) profileZone(g *Glyph) Zone {
	top := f.Row(0) + int(f.P.XHeight/f.P.Kern.Band)
	switch g.Kind {
	case Upper:
		top = f.Row(0) + int(f.P.CapHeight/f.P.Kern.Band)
	case Punct:
		top = f.nrows
	}
	return Zone{0, top}
}

// sideDistance returns the largest difference of the capped depths of two
// profiles over the bands of the zone.
func (f *Font) sideDistance(x, y []float64, z Zone) float64 {
	depth := f.P.Kern.Depth
	var m float64
	for i := z.Lo; i < z.Hi; i++ {
		dx, dy := depth, depth
		if !math.IsNaN(x[i]) {
			dx = math.Min(depth, x[i])
		}
		if !math.IsNaN(y[i]) {
			dy = math.Min(depth, y[i])
		}
		m = math.Max(m, math.Abs(dx-dy))
	}
	return m
}

// classes divides the sides of the glyphs into classes: the first glyph of a
// pair by its right profile and the second by its left profile.  A glyph joins
// the first class, in the order of the glyphs, to whose first member it
// agrees within the class tolerance in every band; glyphs of different kinds
// share no class.
func (t *Table) classes(gs []*Glyph, all map[[2]int]*Pair) {
	f := t.Font
	tol := f.P.Kern.ClassTolerance
	assign := func(profile func(g *Glyph) []float64) (map[int]int, int) {
		cls := make(map[int]int)
		var reps []*Glyph
		for _, g := range gs {
			z := f.profileZone(g)
			found := 0
			for i, r := range reps {
				if r.Kind != g.Kind {
					continue
				}
				if f.sideDistance(profile(r), profile(g), z) <= tol {
					found = i + 1
					break
				}
			}
			if found == 0 {
				reps = append(reps, g)
				found = len(reps)
			}
			cls[g.Index] = found
		}
		return cls, len(reps)
	}
	t.Class1, t.NClass1 = assign(func(g *Glyph) []float64 { return g.Right })
	t.Class2, t.NClass2 = assign(func(g *Glyph) []float64 { return g.Left })
	for _, pr := range all {
		pr.LClass, pr.RClass = t.Class1[pr.Left.Index], t.Class2[pr.Right.Index]
	}
}

// compact finds the kern of each pair of classes, as the median of the kerns
// of the pairs it covers, and the pairs which are left over: those whose kern
// differs from that of their classes by more than the pair tolerance, and
// those at the floor, which the kern of the classes would take below it.  The
// kerns of the pairs are set to the kerns which the font gives.
func (t *Table) compact(all map[[2]int]*Pair) {
	tol := float64(t.Font.P.Kern.PairTolerance)
	type cp struct{ a, b int }
	groups := make(map[cp][]*Pair)
	for _, pr := range all {
		c := cp{pr.LClass, pr.RClass}
		groups[c] = append(groups[c], pr)
	}
	val := make(map[cp]int)
	for c, prs := range groups {
		ks := make([]int, len(prs))
		for i, pr := range prs {
			ks[i] = pr.Kern
		}
		sort.Ints(ks)
		val[c] = ks[(len(ks)-1)/2]
	}
	t.Single = make(map[[2]int]int)
	used1, used2 := map[int]int{}, map[int]int{}
	for c, prs := range groups {
		v := val[c]
		for _, pr := range prs {
			switch {
			case math.Abs(float64(pr.Kern-v)) > tol || (v < pr.Kern && (t.Font.Distance(pr.Left, pr.Right, float64(v)) < t.Font.P.Kern.Floor || (v < 0 && pr.Reason == "" && t.Font.limitApproach(pr.Left, pr.Right, pr.Zone, pr.Ref, v) != v))):
				t.Single[[2]int{pr.Left.Index, pr.Right.Index}] = pr.Kern
			default:
				pr.Kern = v
				pr.White1 = t.Font.white(pr.Left, pr.Right, pr.Zone, float64(v))
				pr.Dist1 = t.Font.Distance(pr.Left, pr.Right, float64(v))
			}
		}
		if v != 0 {
			used1[c.a] = 0
			used2[c.b] = 0
		}
	}
	// Number the classes which have a kern, in the order of first use by the
	// glyph list.
	n1, n2 := map[int]int{}, map[int]int{}
	for _, g := range t.Font.Glyphs {
		if c, ok := t.Class1[g.Index]; ok {
			if _, ok := used1[c]; ok {
				if _, ok := n1[c]; !ok {
					n1[c] = len(n1) + 1
				}
			}
		}
		if c, ok := t.Class2[g.Index]; ok {
			if _, ok := used2[c]; ok {
				if _, ok := n2[c]; !ok {
					n2[c] = len(n2) + 1
				}
			}
		}
	}
	t.Values = make([][]int, len(n1)+1)
	for i := range t.Values {
		t.Values[i] = make([]int, len(n2)+1)
	}
	for c, v := range val {
		if v != 0 {
			t.Values[n1[c.a]][n2[c.b]] = v
		}
	}
	t.Class1, t.Class2 = renumber(t.Class1, n1), renumber(t.Class2, n2)
	t.Single = t.dropRedundant(t.Single, all)
}

// renumber keeps the glyphs whose class has a number in m, with that number.
func renumber(cls map[int]int, m map[int]int) map[int]int {
	out := make(map[int]int)
	for g, c := range cls {
		if n, ok := m[c]; ok {
			out[g] = n
		}
	}
	return out
}

// dropRedundant removes the single pairs which equal what the classes give.
func (t *Table) dropRedundant(single map[[2]int]int, all map[[2]int]*Pair) map[[2]int]int {
	out := make(map[[2]int]int)
	for key, v := range single {
		c1, c2 := t.Class1[key[0]], t.Class2[key[1]]
		if t.Values[c1][c2] != v {
			out[key] = v
		}
	}
	return out
}

// Kern returns the kern the font gives the pair of glyphs, by index: a single
// pair first, then the classes.
func (t *Table) Kern(a, b int) int {
	if v, ok := t.Single[[2]int{a, b}]; ok {
		return v
	}
	return t.Values[t.Class1[a]][t.Class2[b]]
}

// Closest returns the smallest band gap of the pair with b moved by k: the
// least of the sums of the right profile of a and the left profile of b over
// the bands lo to hi (not including hi) in which both have ink.  It is +Inf if
// there is no such band.
func (f *Font) Closest(a, b *Glyph, k float64, lo, hi int) float64 {
	m := math.Inf(1)
	for i := lo; i < hi && i < len(a.Right); i++ {
		ra, lb := a.Right[i], b.Left[i]
		if math.IsNaN(ra) || math.IsNaN(lb) {
			continue
		}
		m = math.Min(m, ra+lb+k)
	}
	return m
}

// refPairs returns the glyphs of the pairs named by ref which stand on the
// side of a and of b: for a mark, the mark and the first letter of ref, and
// for Hn both orders.
func (f *Font) refPairs(a, b *Glyph, ref string) [][2]*Glyph {
	g := func(s string) *Glyph { return f.glyphOf(rune(s[0])) }
	switch {
	case a.Kind == Punct:
		return [][2]*Glyph{{a, g(ref[:1])}}
	case b.Kind == Punct:
		return [][2]*Glyph{{g(ref[:1]), b}}
	case ref == "Hn":
		return [][2]*Glyph{{g("H"), g("n")}, {g("n"), g("H")}}
	}
	return [][2]*Glyph{{g(ref[:1]), g(ref[1:])}}
}

// refClosest returns the closest approach of the pair named by ref over the bands
// lo to hi (the mean of the two orders for Hn).
func (f *Font) refClosest(a, b *Glyph, ref string, lo, hi int) float64 {
	var sum float64
	ps := f.refPairs(a, b, ref)
	for _, p := range ps {
		sum += f.Closest(p[0], p[1], 0, lo, hi)
	}
	return sum / float64(len(ps))
}

// straight says whether the profile is that of a straight stem over the bands
// lo to hi: at least the fraction design.Kern.StraightFrac of the bands have
// ink within design.Kern.StraightTol of the median distance.
func (f *Font) straight(p []float64, lo, hi int) bool {
	var v []float64
	for i := lo; i < hi && i < len(p); i++ {
		if !math.IsNaN(p[i]) {
			v = append(v, p[i])
		}
	}
	if hi <= lo || len(v) == 0 {
		return false
	}
	sort.Float64s(v)
	med := v[len(v)/2]
	n := 0
	for _, x := range v {
		if math.Abs(x-med) <= f.P.Kern.StraightTol {
			n++
		}
	}
	return float64(n) >= f.P.Kern.StraightFrac*float64(hi-lo)
}

// limitApproach limits the negative kern kk so that the pair keeps a fraction
// design.Kern.Approach of the closest approach of the pair named by ref, in two
// measures: over the whole height, and over the body of the zone, without the
// bands within design.Kern.Serif of its ends.  The second applies only if one
// of the two sides is a straight stem.  A pair already under the limit gets no
// negative kern.  A capital before a lowercase letter is also kept from having
// less mean white over the x-height than nn has.
func (f *Font) limitApproach(a, b *Glyph, z Zone, ref string, kk int) int {
	k := f.P.Kern
	limit := func(lo, hi int) {
		c0 := f.Closest(a, b, 0, lo, hi)
		if math.IsInf(c0, 1) || hi <= lo {
			return
		}
		need := k.Approach * f.refClosest(a, b, ref, lo, hi)
		lim := int(math.Ceil(need - c0))
		if lim > 0 {
			lim = 0
		}
		if kk < lim {
			kk = lim
		}
	}
	limit(0, f.nrows)
	if a.Kind != Punct && b.Kind != Punct {
		m := int(k.Serif / k.Band)
		if lo, hi := z.Lo+m, z.Hi-m; f.straight(a.Right, lo, hi) || f.straight(b.Left, lo, hi) {
			limit(lo, hi)
		}
	}
	if a.Kind == Upper && b.Kind == Lower {
		xz := Zone{f.Row(0), f.Row(0) + int(f.P.XHeight/f.P.Kern.Band)}
		n := f.glyphOf('n')
		t := f.white(n, n, xz, 0)
		for kk < 0 && f.white(a, b, xz, float64(kk)) < t {
			kk++
		}
	}
	return kk
}
