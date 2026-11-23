// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tex

import (
	"math"
	"slices"
)

// A Reduction says what the reduction of one table of dimensions did.
type Reduction struct {
	// Table is the name of the table: height, depth or italic.
	Table string

	// Before and After are the numbers of distinct values other than zero
	// before and after the reduction.
	Before, After int

	// Change is the largest change of a value, in units.
	Change float64
}

// reduce maps the values to at most n distinct values other than zero and
// returns the map and the largest change of a value.
//
// Zero and the values of keep are fixed: zero stays zero, no other value
// becomes zero, and a value of keep is a value which the result may hold and
// which never moves.  Every other value changes by no more than the least e
// for which the result fits: the values within e of a fixed value take it, and
// the rest are covered by intervals of width 2e, each of which takes its
// midpoint.  The fixed values which take a value count among the n, and the
// intervals have the rest.  Without values to keep this is the method of
// PLtoTF, which minimises the largest change of all values.
func reduce(values, keep []float64, n int) (map[float64]float64, float64) {
	out := make(map[float64]float64, len(values))
	var vs []float64
	for _, v := range values {
		if v == 0 {
			out[0] = 0
		} else {
			vs = append(vs, v)
		}
	}
	slices.Sort(vs)
	vs = slices.Compact(vs)
	ks := slices.DeleteFunc(slices.Clone(keep), func(k float64) bool { return k == 0 })
	slices.Sort(ks)
	ks = slices.Compact(ks)

	// plan assigns the values for the change e: the fixed value of each value
	// which takes one, and the ends of the intervals of the others
	plan := func(e float64) (fixed []float64, ends []int, used int) {
		fixed = make([]float64, len(vs))
		taken := make(map[float64]bool)
		for i, v := range vs {
			for _, k := range ks {
				if math.Abs(v-k) <= e && (fixed[i] == 0 || math.Abs(v-k) < math.Abs(v-fixed[i])) {
					fixed[i] = k
				}
			}
			if fixed[i] != 0 {
				taken[fixed[i]] = true
			}
		}
		for i := 0; i < len(vs); {
			if fixed[i] != 0 {
				i++
				continue
			}
			j := i + 1
			for j < len(vs) && fixed[j] == 0 && vs[j]-vs[i] <= 2*e {
				j++
			}
			ends = append(ends, j)
			i = j
		}
		return fixed, ends, len(taken)
	}

	cands := []float64{0}
	for i, v := range vs {
		for _, u := range vs[i+1:] {
			cands = append(cands, (u-v)/2)
		}
		for _, k := range ks {
			cands = append(cands, math.Abs(v-k))
		}
	}
	slices.Sort(cands)
	cands = slices.Compact(cands)
	e := cands[len(cands)-1]
	for _, c := range cands {
		if _, ends, used := plan(c); used+len(ends) <= n {
			e = c
			break
		}
	}

	fixed, _, _ := plan(e)
	change := 0.0
	i := 0
	for i < len(vs) {
		if fixed[i] != 0 {
			out[vs[i]] = fixed[i]
			change = math.Max(change, math.Abs(vs[i]-fixed[i]))
			i++
			continue
		}
		j := i + 1
		for j < len(vs) && fixed[j] == 0 && vs[j]-vs[i] <= 2*e {
			j++
		}
		rep := (vs[i] + vs[j-1]) / 2
		for _, v := range vs[i:j] {
			out[v] = rep
			change = math.Max(change, math.Abs(v-rep))
		}
		i = j
	}
	return out, change
}

// reduceTable reduces the dimension of the characters which get and set
// access to at most n values other than zero, and returns what it did.
func (m *Metrics) reduceTable(name string, n int, keep []float64, get func(*Char) *float64) Reduction {
	var values []float64
	for _, ch := range m.Chars {
		if ch != nil {
			values = append(values, *get(ch))
		}
	}
	distinct := func() int {
		seen := make(map[float64]bool)
		for _, ch := range m.Chars {
			if ch != nil && *get(ch) != 0 {
				seen[*get(ch)] = true
			}
		}
		return len(seen)
	}
	r := Reduction{Table: name, Before: distinct()}
	if r.Before > n {
		to, change := reduce(values, keep, n)
		for _, ch := range m.Chars {
			if ch != nil {
				*get(ch) = to[*get(ch)]
			}
		}
		r.Change = change
	}
	r.After = distinct()
	return r
}
