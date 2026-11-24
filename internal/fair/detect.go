// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package fair

import (
	"fmt"
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/outline"
)

// radius writes a curvature as a radius of curvature, for the details.
func radius(k float64) string {
	if math.Abs(k) < 1e-5 {
		return "straight"
	}
	return fmt.Sprintf("r %.0f", 1/math.Abs(k))
}

// detect fills a.Findings and a.Notes.
func (a *Analysis) detect() {
	th := a.Th
	var fs []Finding
	for ci := range a.Contours {
		c := &a.Contours[ci]
		fs = append(fs, a.nodeFindings(ci, c)...)
		fs = append(fs, a.shortFindings(ci, c)...)
		if th.ExtremeTol > 0 {
			fs = append(fs, a.extremeFindings(ci, c)...)
		}
	}
	for _, r := range a.Runs {
		if r.Straight {
			continue
		}
		ext := r
		if r.Cyclic {
			ext = extend(r, a.Contours[r.Contour].Length, th)
		}
		fs = append(fs, a.flatFindings(ext)...)
		fs = append(fs, a.inflections(ext)...)
		fs = append(fs, a.extremes(ext)...)
		thick := a.thickFindings(ext)
		if th.ThickFind {
			fs = append(fs, thick...)
		} else {
			a.Notes = append(a.Notes, dedupe(a.Contours, thick)...)
		}
	}
	a.Findings = dedupe(a.Contours, fs)
}

// extend appends to a cyclic run its beginning again, so that every place
// has its neighbourhood on both sides somewhere in the run.
func extend(r Run, length float64, th Thresholds) Run {
	reach := max(th.WiggleLen, th.SpikeDist, th.FlatLen+th.FlatNear, th.PinchDist) + 20
	out := r
	out.Samples = append([]Sample(nil), r.Samples...)
	for _, s := range r.Samples {
		if s.S > reach {
			break
		}
		s.S += length
		out.Samples = append(out.Samples, s)
	}
	return out
}

// dedupe removes findings of the same kind on the same contour within 6
// units of a larger one, and brings the arc lengths of the extended cyclic
// runs back onto the contour.
func dedupe(cs []Contour, fs []Finding) []Finding {
	for i := range fs {
		if l := cs[fs[i].Contour].Length; fs[i].S >= l {
			fs[i].S -= l
		}
	}
	sort.SliceStable(fs, func(i, j int) bool { return fs[i].Ratio > fs[j].Ratio })
	var out []Finding
next:
	for _, f := range fs {
		for _, g := range out {
			if g.Kind == f.Kind && g.Contour == f.Contour && math.Hypot(g.X-f.X, g.Y-f.Y) < 6 {
				continue next
			}
		}
		out = append(out, f)
	}
	return out
}

// at returns the finding of kind k at the sample s.
func at(k Kind, ci int, s Sample) Finding {
	return Finding{Kind: k, Contour: ci, X: s.P.X, Y: s.P.Y, S: s.S, Dir: deg(s.T)}
}

// nodeFindings reports the kinks and the jumps of curvature at the nodes.
func (a *Analysis) nodeFindings(ci int, c *Contour) []Finding {
	th := a.Th
	var fs []Finding
	n := len(c.Segs)
	for i, nd := range c.Nodes {
		t := math.Abs(nd.Turn)
		if nd.Corner {
			continue
		}
		fillet := c.Fillet[i] || c.Fillet[(i+n-1)%n] || c.nearCorner(nd.S, th.JointLen)
		s := Sample{P: nd.P, S: nd.S, T: tangentAt(c, nd)}
		if t > th.KinkDeg {
			f := at(Kink, ci, s)
			f.Size, f.Ratio = t, t/th.KinkDeg
			f.Detail = fmt.Sprintf("the edge turns by %.1f degrees at the node", t)
			fs = append(fs, f)
			continue
		}
		ka, kb := math.Abs(nd.KIn), math.Abs(nd.KOut)
		diff := math.Abs(nd.KOut - nd.KIn)
		ratio := math.Inf(1)
		if nd.KIn*nd.KOut > 0 || (ka < th.InflK || kb < th.InflK) {
			if lo := min(ka, kb); lo > 1e-9 {
				ratio = max(ka, kb) / lo
			}
		}
		if diff > th.JumpAbs && ratio > th.JumpRatio && !fillet {
			f := at(Jump, ci, s)
			f.Size, f.Ratio = diff, diff/th.JumpAbs
			f.Detail = fmt.Sprintf("curvature %.4f to %.4f (%s to %s)", nd.KIn, nd.KOut, radius(nd.KIn), radius(nd.KOut))
			fs = append(fs, f)
		}
	}
	return fs
}

// tangentAt returns the outgoing tangent at a node.
func tangentAt(c *Contour, nd Node) Point {
	for _, s := range c.Samples {
		if s.S >= nd.S-1e-9 && s.P == nd.P {
			return s.T
		}
	}
	return Point{X: 1}
}

// shortFindings reports segments shorter than ShortSeg which are not part of
// a circle.
func (a *Analysis) shortFindings(ci int, c *Contour) []Finding {
	th := a.Th
	var fs []Finding
	n := len(c.Segs)
	for i, g := range c.Segs {
		l := g.length()
		if l >= th.ShortSeg {
			continue
		}
		var ks []float64
		var mid Sample
		for _, s := range c.Samples {
			if s.Seg == i {
				ks = append(ks, s.K)
				if s.S <= c.Nodes[i].S+l/2 {
					mid = s
				}
			}
		}
		if g.Curve && circular(append(ks, c.Nodes[i].KIn, c.Nodes[(i+1)%n].KOut)) {
			continue
		}
		f := at(Short, ci, mid)
		f.Size, f.Ratio = l, th.ShortSeg/max(l, 1)
		kind := "line"
		if g.Curve {
			kind = "curve"
		}
		f.Detail = fmt.Sprintf("a %s of length %.1f", kind, l)
		fs = append(fs, f)
	}
	return fs
}

// circular reports whether the curvatures are all within 4% of their mean,
// as on an arc of a circle.
func circular(ks []float64) bool {
	var m float64
	for _, k := range ks {
		m += k
	}
	m /= float64(len(ks))
	if math.Abs(m) < 1e-4 {
		return false
	}
	for _, k := range ks {
		if math.Abs(k-m) > 0.04*math.Abs(m) {
			return false
		}
	}
	return true
}

// flatFindings reports nearly straight stretches between curved stretches
// which bend the same way, and short straight segments between them.
func (a *Analysis) flatFindings(r Run) []Finding {
	th := a.Th
	c := &a.Contours[r.Contour]
	ss := r.Samples
	var fs []Finding
	for i := 0; i < len(ss); {
		if math.Abs(ss[i].K) >= th.FlatK {
			i++
			continue
		}
		j := i
		for j+1 < len(ss) && math.Abs(ss[j+1].K) < th.FlatK {
			j++
		}
		i0, i1 := i, j
		i = j + 1
		l := ss[i1].S - ss[i0].S
		if l < th.FlatLen {
			continue
		}
		long := false
		for k := i0; k <= i1; k++ {
			if g := c.Segs[ss[k].Seg]; !g.Curve && g.length() >= th.LineMax {
				long = true
				break
			}
		}
		if long {
			continue
		}
		left, right := 0.0, 0.0
		for k := i0 - 1; k >= 0 && ss[i0].S-ss[k].S <= th.FlatNear; k-- {
			if math.Abs(ss[k].K) > math.Abs(left) {
				left = ss[k].K
			}
		}
		for k := i1 + 1; k < len(ss) && ss[k].S-ss[i1].S <= th.FlatNear; k++ {
			if math.Abs(ss[k].K) > math.Abs(right) {
				right = ss[k].K
			}
		}
		if math.Abs(left) <= th.CurvedK || math.Abs(right) <= th.CurvedK || left*right < 0 {
			continue
		}
		mid := ss[i0]
		for k := i0; k <= i1; k++ {
			if ss[k].S <= (ss[i0].S+ss[i1].S)/2 {
				mid = ss[k]
			}
		}
		f := at(Flat, r.Contour, mid)
		f.Size, f.Ratio = l, l/th.FlatLen
		f.Detail = fmt.Sprintf("%.0f units straighter than %s between %s and %s", l, radius(th.FlatK), radius(left), radius(right))
		fs = append(fs, f)
	}
	return fs
}

// stretch is a part of a run in which the curvature has one sign.
type stretch struct {
	i0, i1 int
	sign   float64
	peak   float64
}

// inflections reports the places where the curvature changes its sign.
// Stretches of one sign which are both short (under 4 units) and weak (below
// four times InflK) do not count, so that noise about zero is no
// inflection.
func (a *Analysis) inflections(r Run) []Finding {
	th := a.Th
	ss := r.Samples
	var sts []stretch
	for i := 0; i < len(ss); {
		if math.Abs(ss[i].K) <= th.InflK {
			i++
			continue
		}
		sign := math.Copysign(1, ss[i].K)
		j := i
		peak := math.Abs(ss[i].K)
		for j+1 < len(ss) && math.Abs(ss[j+1].K) > th.InflK && math.Copysign(1, ss[j+1].K) == sign {
			j++
			peak = max(peak, math.Abs(ss[j].K))
		}
		if ss[j].S-ss[i].S >= 4 || peak >= 4*th.InflK {
			sts = append(sts, stretch{i, j, sign, peak})
		}
		i = j + 1
	}
	var fs []Finding
	for k := 1; k < len(sts); k++ {
		p, q := sts[k-1], sts[k]
		if p.sign == q.sign {
			continue
		}
		if th.InflLine > 0 && a.lineBetween(r, ss[p.i1].S, ss[q.i0].S) >= th.InflLine {
			continue
		}
		s0 := (ss[p.i1].S + ss[q.i0].S) / 2
		mid := ss[p.i1]
		for i := p.i1; i <= q.i0; i++ {
			if ss[i].S <= s0 {
				mid = ss[i]
			}
		}
		f := at(Inflection, r.Contour, mid)
		f.Size, f.Ratio = min(p.peak, q.peak), 1
		f.Detail = fmt.Sprintf("from %s to %s, peaks %s and %s", sgnWord(p.sign), sgnWord(q.sign), radius(p.peak), radius(q.peak))
		fs = append(fs, f)
	}
	// Two inflections closer than WiggleLen make the edge go one way and
	// back: a wiggle, however weak the curvature, at the peak between them.
	var infl []int // index of the stretch after each inflection
	for k := 1; k < len(sts); k++ {
		if sts[k-1].sign != sts[k].sign {
			infl = append(infl, k)
		}
	}
	for j := 1; j < len(infl); j++ {
		mid := sts[infl[j]-1] // the stretch between the two inflections
		a0, a1 := ss[sts[infl[j-1]-1].i1].S, ss[sts[infl[j]].i0].S
		if a1-a0 > th.WiggleLen {
			continue
		}
		peak := ss[mid.i0]
		for i := mid.i0; i <= mid.i1; i++ {
			if math.Abs(ss[i].K) > math.Abs(peak.K) {
				peak = ss[i]
			}
		}
		f := at(Wiggle, r.Contour, peak)
		f.Size = a1 - a0
		f.Ratio = th.WiggleLen / max(f.Size, 1)
		f.Detail = fmt.Sprintf("two inflections %.0f units apart, the edge bends %s (%s) between them", f.Size, sgnWord(mid.sign), radius(mid.peak))
		fs = append(fs, f)
	}
	return fs
}

func sgnWord(s float64) string {
	if s > 0 {
		return "convex"
	}
	return "concave"
}

// extreme is a local extreme of the curvature along a run.
type extreme struct {
	i   int
	max bool
}

// zigzag returns the extremes of the curvature which stand out from their
// surroundings by at least delta.
func zigzag(ss []Sample, delta float64) []extreme {
	var out []extreme
	if len(ss) == 0 {
		return nil
	}
	hi, lo := 0, 0
	dir := 0 // 1 looking for a maximum, -1 for a minimum
	for i := 1; i < len(ss); i++ {
		k := ss[i].K
		if k > ss[hi].K {
			hi = i
		}
		if k < ss[lo].K {
			lo = i
		}
		switch dir {
		case 0:
			if ss[hi].K-k >= delta && hi > 0 {
				out = append(out, extreme{hi, true})
				dir, lo = -1, i
			} else if k-ss[lo].K >= delta && lo > 0 {
				out = append(out, extreme{lo, false})
				dir, hi = 1, i
			}
		case 1:
			if ss[hi].K-k >= delta {
				out = append(out, extreme{hi, true})
				dir, lo = -1, i
			}
		case -1:
			if k-ss[lo].K >= delta {
				out = append(out, extreme{lo, false})
				dir, hi = 1, i
			}
		}
	}
	return out
}

// kAt returns the curvature at the arc length s, from the sample nearest to
// it, and false if s is outside the run.
func kAt(ss []Sample, s float64) (float64, bool) {
	if len(ss) == 0 || s < ss[0].S || s > ss[len(ss)-1].S {
		return 0, false
	}
	i := sort.Search(len(ss), func(i int) bool { return ss[i].S >= s })
	if i > 0 && s-ss[i-1].S < ss[i].S-s {
		i--
	}
	return ss[i].K, true
}

// extremes reports wiggles and spikes from the extremes of the curvature.
func (a *Analysis) extremes(r Run) []Finding {
	th := a.Th
	ss := r.Samples
	ex := zigzag(ss, th.WiggleAbs)
	cont := &a.Contours[r.Contour]
	fil := cont.Fillet
	var fs []Finding
	for k := 0; k+2 < len(ex); k++ {
		A, B, C := ss[ex[k].i], ss[ex[k+1].i], ss[ex[k+2].i]
		if C.S-A.S > th.WiggleLen || fil[A.Seg] || fil[B.Seg] || fil[C.Seg] || cont.nearCorner(B.S, th.JointLen) {
			continue
		}
		if math.Abs(A.K) > th.FilletK || math.Abs(B.K) > th.FilletK || math.Abs(C.K) > th.FilletK {
			continue
		}
		var depth float64
		if ex[k].max {
			depth = min(A.K, C.K) - B.K
		} else {
			depth = B.K - max(A.K, C.K)
		}
		base := max(math.Abs(A.K), math.Abs(C.K))
		if depth < th.WiggleAbs || depth < th.WiggleDepth*base {
			continue
		}
		f := at(Wiggle, r.Contour, B)
		f.Size, f.Ratio = depth/base, depth/base/th.WiggleDepth
		word := "falls"
		if !ex[k].max {
			word = "rises"
		}
		f.Detail = fmt.Sprintf("curvature %.4f, %.4f, %.4f within %.0f units: it %s by %.0f%% between two extremes", A.K, B.K, C.K, C.S-A.S, word, 100*depth/base)
		fs = append(fs, f)
	}
	for _, e := range ex {
		s := ss[e.i]
		if fil[s.Seg] || e.max != (s.K > 0) || cont.nearCorner(s.S, th.JointLen) {
			continue // a peak of |curvature| is a maximum of a convex or a minimum of a concave stretch
		}
		k0 := math.Abs(s.K)
		kl, okl := kAt(ss, s.S-th.SpikeDist)
		kr, okr := kAt(ss, s.S+th.SpikeDist)
		if !okl || !okr || k0 < th.SpikeMinK || k0 > th.FilletK {
			continue
		}
		if nearLine(ss, s.S, 2*th.SpikeDist, cont) {
			continue // the bracket of a serif, where a curve tightens into a straight stem
		}
		side := max(math.Abs(kl), math.Abs(kr))
		if k0 < th.SpikeMul*side || min(math.Abs(kl), math.Abs(kr)) < th.CurvedK {
			// A peak between nearly straight stretches is a bend, which
			// the wiggles and the jumps judge.
			continue
		}
		f := at(Spike, r.Contour, s)
		f.Size = k0 / max(side, 1e-6)
		f.Ratio = f.Size / th.SpikeMul
		f.Detail = fmt.Sprintf("%s against %s and %s at %.0f units on either side", radius(k0), radius(kl), radius(kr), th.SpikeDist)
		fs = append(fs, f)
	}
	return fs
}

// thickFindings reports where the thickness of the stroke is locally less
// (a pinch) or more (a swelling) than the mean of the thickness at PinchDist
// on both sides, by more than PinchShare.
func (a *Analysis) thickFindings(r Run) []Finding {
	th := a.Th
	c := &a.Contours[r.Contour]
	if len(r.Samples) == 0 {
		return nil
	}
	s0, s1 := r.Samples[0].S, r.Samples[len(r.Samples)-1].S
	// The thickness samples of the run, in order, unwrapped.
	var ts []Thick
	for _, t := range c.Thick {
		for _, off := range []float64{0, c.Length} {
			if s := t.S + off; s >= s0 && s <= s1 {
				t.S = s
				ts = append(ts, t)
			}
		}
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i].S < ts[j].S })
	// rel[i] is the mean of the neighbours at PinchDist less the thickness,
	// relative to the thickness; NaN where the neighbours are not valid.
	rel := make([]float64, len(ts))
	find := func(i int, target float64) (int, bool) {
		j := i
		step := 1
		if target < ts[i].S {
			step = -1
		}
		for j >= 0 && j < len(ts) {
			if !ts[j].Valid {
				return 0, false
			}
			if (step > 0 && ts[j].S >= target) || (step < 0 && ts[j].S <= target) {
				return j, math.Abs(ts[j].S-target) < thickStep*1.5
			}
			j += step
		}
		return 0, false
	}
	for i := range ts {
		rel[i] = math.NaN()
		if !ts[i].Valid {
			continue
		}
		l, okl := find(i, ts[i].S-th.PinchDist)
		h, okh := find(i, ts[i].S+th.PinchDist)
		if !okl || !okh {
			continue
		}
		rel[i] = ((ts[l].T+ts[h].T)/2 - ts[i].T) / ts[i].T
	}
	var fs []Finding
	for i := range ts {
		v := rel[i]
		if math.IsNaN(v) || math.Abs(v) <= th.PinchShare {
			continue
		}
		top := true
		for j := i - 1; j >= 0 && ts[i].S-ts[j].S <= th.PinchDist; j-- {
			if !math.IsNaN(rel[j]) && math.Abs(rel[j]) > math.Abs(v) {
				top = false
			}
		}
		for j := i + 1; j < len(ts) && ts[j].S-ts[i].S <= th.PinchDist; j++ {
			if !math.IsNaN(rel[j]) && math.Abs(rel[j]) >= math.Abs(v) {
				top = false
			}
		}
		if !top {
			continue
		}
		kind := Pinch
		if v < 0 {
			kind = Swell
		}
		k, _ := kAt(r.Samples, ts[i].S)
		_ = k
		var tan Point
		for _, s := range r.Samples {
			if s.S <= ts[i].S {
				tan = s.T
			}
		}
		f := Finding{Kind: kind, Contour: r.Contour, X: ts[i].P.X, Y: ts[i].P.Y, S: ts[i].S, Dir: deg(tan)}
		f.Size, f.Ratio = math.Abs(v), math.Abs(v)/th.PinchShare
		f.Detail = fmt.Sprintf("thickness %.0f against a mean of %.0f at %.0f units on either side", ts[i].T, ts[i].T*(1+v), th.PinchDist)
		fs = append(fs, f)
	}
	return fs
}

// nearCorner reports whether the arc length s is closer than d to a corner
// of the contour, along the contour.
func (c *Contour) nearCorner(s, d float64) bool {
	for _, nd := range c.Nodes {
		if !nd.Corner {
			continue
		}
		x := math.Mod(math.Abs(s-nd.S), c.Length)
		if min(x, c.Length-x) < d {
			return true
		}
	}
	return false
}

// nearLine reports whether a straight segment lies within d of the arc
// length s in the run.
func nearLine(ss []Sample, s, d float64, c *Contour) bool {
	for _, x := range ss {
		if math.Abs(x.S-s) < d && !c.Segs[x.Seg].Curve {
			return true
		}
	}
	return false
}

// lineBetween returns the length of the longest straight segment of the run
// whose samples lie between the arc lengths s0 and s1.
func (a *Analysis) lineBetween(r Run, s0, s1 float64) float64 {
	c := &a.Contours[r.Contour]
	var best float64
	for _, x := range r.Samples {
		if x.S < s0 || x.S > s1 {
			continue
		}
		if g := c.Segs[x.Seg]; !g.Curve {
			best = max(best, g.length())
		}
	}
	return best
}

// extremeFindings reports the curves which run beyond both their end points,
// in x or in y, by more than ExtremeTol: their extreme has no node.  The
// extreme in x of a curve next to a straight segment of the slant StemSlant
// (the curve touches the segment, or follows the curve which does at a smooth
// node) may run up to StemExtremeTol beyond.
func (a *Analysis) extremeFindings(ci int, c *Contour) []Finding {
	th := a.Th
	n := len(c.Segs)
	stem := func(i int) bool {
		g := c.Segs[(i+n)%n]
		d := sub(g.P3, g.P0)
		return !g.Curve && math.Abs(d.Y) > 1e-9 && math.Abs(d.X/d.Y-th.StemSlant) < 1e-3 && th.StemSlant != 0
	}
	smooth := func(i int) bool { return !c.Nodes[(i+n)%n].Corner }
	var fs []Finding
	for i, g := range c.Segs {
		if !g.Curve {
			continue
		}
		nearStem := stem(i-1) || stem(i+1) ||
			c.Segs[(i+n-1)%n].Curve && smooth(i) && stem(i-2) ||
			c.Segs[(i+1)%n].Curve && smooth(i+1) && stem(i+2)
		for axis := 0; axis < 2; axis++ {
			get := func(q Point) float64 {
				if axis == 0 {
					return q.X
				}
				return q.Y
			}
			v0, v1, v2, v3 := get(g.P0), get(g.P1), get(g.P2), get(g.P3)
			lo, hi := math.Min(v0, v3), math.Max(v0, v3)
			for _, t := range outline.CubicExtrema(v0, v1, v2, v3) {
				u := 1 - t
				v := u*u*u*v0 + 3*u*u*t*v1 + 3*u*t*t*v2 + t*t*t*v3
				over := math.Max(v-hi, lo-v)
				tol := th.ExtremeTol
				if axis == 0 && nearStem {
					tol = th.StemExtremeTol
				}
				if over <= tol {
					continue
				}
				p, d1, _ := g.at(t)
				f := Finding{Kind: Extreme, Contour: ci, X: p.X, Y: p.Y, Dir: deg(d1)}
				for _, s := range c.Samples {
					if s.Seg == i {
						f.S = s.S
						if dist(s.P, p) < 2 {
							break
						}
					}
				}
				f.Size, f.Ratio = over, over/tol
				name := "x"
				if axis == 1 {
					name = "y"
				}
				f.Detail = fmt.Sprintf("the curve runs %.2f beyond its nodes in %s", over, name)
				fs = append(fs, f)
			}
		}
	}
	return fs
}
