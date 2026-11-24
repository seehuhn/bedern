// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package fair

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Entry is a line of the baseline of a glyph: a finding which is known, and
// whether it is intended or a defect, with the reason in words.
type Entry struct {
	Kind     Kind
	X, Y     float64
	Dir      float64
	Intended bool
	Reason   string // "?" for an entry which nobody has judged yet
}

// Unjudged is the reason of an entry which the update of the baseline has
// added and nobody has judged yet.  The gate fails on it.
const Unjudged = "?"

// Matches reports whether the entry describes the finding f.
func (e Entry) Matches(f Finding) bool {
	return Same(e.Kind, e.X, e.Y, e.Dir, f)
}

// String returns the entry as a line of a baseline file.
func (e Entry) String() string {
	status := "defect"
	if e.Intended {
		status = "intended"
	}
	return fmt.Sprintf("%s %.0f %.0f %.0f %s: %s", e.Kind, e.X, e.Y, normDeg(e.Dir), status, e.Reason)
}

// EntryOf returns the entry for a finding, as a defect not yet judged.
func EntryOf(f Finding) Entry {
	return Entry{Kind: f.Kind, X: math.Round(f.X), Y: math.Round(f.Y), Dir: math.Round(normDeg(f.Dir)), Reason: Unjudged}
}

// ParseBaseline reads a baseline file: lines "kind x y dir status: reason",
// where dir is the direction of the edge in degrees (with the ink on the
// left) and status is "intended" or "defect".  Empty lines and lines starting
// with # are comments.
func ParseBaseline(r io.Reader) ([]Entry, error) {
	var out []Entry
	sc := bufio.NewScanner(r)
	n := 0
	for sc.Scan() {
		n++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		head, reason, ok := strings.Cut(line, ":")
		f := strings.Fields(head)
		if !ok || len(f) != 5 {
			return nil, fmt.Errorf("line %d: want \"kind x y dir intended|defect: reason\", got %q", n, line)
		}
		var e Entry
		e.Kind = Kind(f[0])
		known := false
		for _, k := range Kinds {
			known = known || k == e.Kind
		}
		if !known {
			return nil, fmt.Errorf("line %d: unknown kind %q", n, f[0])
		}
		var err error
		var v [3]float64
		for i := range v {
			if v[i], err = strconv.ParseFloat(f[1+i], 64); err != nil {
				return nil, fmt.Errorf("line %d: %v", n, err)
			}
		}
		e.X, e.Y, e.Dir = v[0], v[1], v[2]
		switch f[4] {
		case "intended":
			e.Intended = true
		case "defect":
		default:
			return nil, fmt.Errorf("line %d: status %q is neither intended nor defect", n, f[4])
		}
		e.Reason = strings.TrimSpace(reason)
		if e.Reason == "" {
			return nil, fmt.Errorf("line %d: no reason", n)
		}
		out = append(out, e)
	}
	return out, sc.Err()
}

// Compare matches the findings with the baseline entries.  It returns the
// findings which no entry describes, and the entries which describe no
// finding.  Each finding uses up at most one entry and each entry at most one
// finding, the nearest first.
func Compare(fs []Finding, es []Entry) (newFs []Finding, gone []Entry, matched map[int]int) {
	type pair struct {
		f, e int
		d    float64
	}
	var ps []pair
	for i, f := range fs {
		for j, e := range es {
			if e.Matches(f) {
				ps = append(ps, pair{i, j, math.Hypot(e.X-f.X, e.Y-f.Y)})
			}
		}
	}
	sort.SliceStable(ps, func(a, b int) bool { return ps[a].d < ps[b].d })
	matched = map[int]int{} // finding -> entry
	usedE := map[int]bool{}
	for _, p := range ps {
		if _, ok := matched[p.f]; ok || usedE[p.e] {
			continue
		}
		matched[p.f] = p.e
		usedE[p.e] = true
	}
	for i, f := range fs {
		if _, ok := matched[i]; !ok {
			newFs = append(newFs, f)
		}
	}
	for j, e := range es {
		if !usedE[j] {
			gone = append(gone, e)
		}
	}
	return newFs, gone, matched
}

// UpdateBaseline returns the baseline for the findings: the entries which
// describe a finding keep their status and reason and take the place of the
// finding, entries which describe no finding are dropped, and findings which
// no entry describes are added as unjudged defects.
func UpdateBaseline(fs []Finding, es []Entry) []Entry {
	_, _, matched := Compare(fs, es)
	var out []Entry
	for i, f := range fs {
		e := EntryOf(f)
		if j, ok := matched[i]; ok {
			e.Intended, e.Reason = es[j].Intended, es[j].Reason
		}
		out = append(out, e)
	}
	return out
}

// FormatBaseline writes the baseline file of a glyph.
func FormatBaseline(w io.Writer, glyph string, es []Entry) error {
	if _, err := fmt.Fprintf(w, "# Fairness findings of %s in the regular font which are known: kind x y direction status: reason.\n", glyph); err != nil {
		return err
	}
	for _, e := range es {
		if _, err := fmt.Fprintln(w, e.String()); err != nil {
			return err
		}
	}
	return nil
}
