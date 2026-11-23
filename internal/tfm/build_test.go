// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

// TestBuildGolden compares the files built from the fonts of the tests with
// the files which PLtoTF makes of their descriptions.
func TestBuildGolden(t *testing.T) {
	for _, tf := range testFonts() {
		_, _, data := build(t, tf.font)
		want := golden(t, tf.name+".tfm")
		if !bytes.Equal(data, want) {
			t.Errorf("%s: other bytes than PLtoTF, first at %d", tf.name, firstDiff(data, want))
		}
		var buf bytes.Buffer
		if _, err := tf.font.Write(&buf); err != nil {
			t.Errorf("%s: write failed: %v", tf.name, err)
		} else if !bytes.Equal(buf.Bytes(), want) {
			t.Errorf("%s: Font.Write gives other bytes than PLtoTF", tf.name)
		}
	}
}

// normal returns the description which the file built from f gives: with
// the check sum and the flag of the header, with fix words of the design
// size, without skips of no steps and with a stop at the end of the
// program.
func normal(f *Font, file *File, rep *Report) *Font {
	n := *f
	n.Checksum, n.HasChecksum = file.Checksum(), true
	n.SevenBitSafe = rep.SevenBitSafe
	units := f.DesignUnits
	n.DesignUnits = Unity
	scale := func(x Fix) Fix {
		if units == Unity {
			return x
		}
		return Fix(roundHalfAway(float64(x) / float64(units) * 1048576.0))
	}
	n.Params = append([]Fix(nil), f.Params...)
	for i := 1; i < len(n.Params); i++ {
		n.Params[i] = scale(n.Params[i])
	}
	n.LigTable = nil
	for _, in := range f.LigTable {
		if in.Kind == KindSkip && in.Skip == 0 {
			continue
		}
		if in.Kind == KindStep && in.Step.Op == OpKern {
			in.Step.Kern = scale(in.Step.Kern)
		}
		n.LigTable = append(n.LigTable, in)
	}
	if k := len(n.LigTable); k > 0 && n.LigTable[k-1].Kind == KindStep {
		n.LigTable = append(n.LigTable, Stop())
	}
	for c, ch := range f.Chars {
		if ch != nil {
			d := *ch
			d.Width, d.Height, d.Depth, d.Italic = scale(d.Width), scale(d.Height), scale(d.Depth), scale(d.Italic)
			n.Chars[c] = &d
		}
	}
	return &n
}

// TestBuildRoundTrip checks that the file built from a description has the
// same description, up to the reductions of the report.
func TestBuildRoundTrip(t *testing.T) {
	for _, tf := range testFonts() {
		file, rep, _ := build(t, tf.font)
		want := normal(tf.font, file, rep)
		got := file.Font()
		limit := map[string]Fix{}
		for _, r := range rep.Reductions {
			limit[r.Table] = r.MaxChange
		}
		largest := map[string]Fix{}
		for c, ch := range got.Chars {
			w := want.Chars[c]
			if ch == nil || w == nil {
				continue
			}
			for _, d := range []struct {
				name     string
				got, set *Fix
				want     Fix
			}{{"width", &ch.Width, &w.Width, w.Width}, {"height", &ch.Height, &w.Height, w.Height},
				{"depth", &ch.Depth, &w.Depth, w.Depth}, {"italic correction", &ch.Italic, &w.Italic, w.Italic}} {
				diff := *d.got - d.want
				if diff < 0 {
					diff = -diff
				}
				if diff > limit[d.name] {
					t.Errorf("%s: %s of %d is %v, want %v", tf.name, d.name, c, *d.got, d.want)
				}
				largest[d.name] = max(largest[d.name], diff)
				*d.set = *d.got
			}
		}
		if !reflect.DeepEqual(largest, limit) && len(limit) > 0 {
			t.Errorf("%s: largest changes %v, the report says %v", tf.name, largest, limit)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: the file has another description", tf.name)
		}
	}
}

// TestReduction checks the report of a font with too many values in every
// table of dimensions.
func TestReduction(t *testing.T) {
	file, rep, _ := build(t, reductionFont())
	want := []struct {
		table    string
		from, to int
	}{{"width", 257, 256}, {"height", 50, 16}, {"depth", 21, 16}, {"italic correction", 80, 64}}
	if len(rep.Reductions) != len(want) {
		t.Fatalf("%d reductions, want %d", len(rep.Reductions), len(want))
	}
	for i, w := range want {
		r := rep.Reductions[i]
		if r.Table != w.table || r.From != w.from || r.To != w.to {
			t.Errorf("reduction %d is %+v, want %s from %d to %d", i, r, w.table, w.from, w.to)
		}
		if r.MaxChange <= 0 || r.Tolerance <= 0 || 2*r.MaxChange > r.Tolerance+1 {
			t.Errorf("%s: largest change %v at tolerance %v", r.Table, r.MaxChange, r.Tolerance)
		}
	}
	if len(file.Width) != 256 || len(file.Height) != 16 || len(file.Depth) != 16 || len(file.Italic) != 64 {
		t.Errorf("tables of %d, %d, %d and %d entries", len(file.Width), len(file.Height), len(file.Depth), len(file.Italic))
	}
	for _, tf := range testFonts() {
		if _, rep, _ := build(t, tf.font); tf.name != "reduction" && len(rep.Reductions) != 0 {
			t.Errorf("%s: reductions %v", tf.name, rep.Reductions)
		}
	}
}

// TestSevenBitSafe checks the flag of the report.
func TestSevenBitSafe(t *testing.T) {
	want := map[string]bool{"text": true, "ops": false, "symbol": true, "extension": true, "big": true,
		"bigboundary": true, "reduction": true, "header": true, "empty": true}
	for _, tf := range testFonts() {
		file, rep, _ := build(t, tf.font)
		if rep.SevenBitSafe != want[tf.name] || file.SevenBitSafe() != want[tf.name] {
			t.Errorf("%s: flag %v in the report and %v in the file, want %v", tf.name, rep.SevenBitSafe, file.SevenBitSafe(), want[tf.name])
		}
	}
}

// TestBuildInvalid checks that Build refuses descriptions which the format
// cannot hold.
func TestBuildInvalid(t *testing.T) {
	bad := []func(f *Font){
		func(f *Font) { f.DesignSize = Unity - 1 },
		func(f *Font) { f.DesignUnits = -Unity },
		func(f *Font) { f.CodingScheme = "0123456789012345678901234567890123456789" },
		func(f *Font) { f.Family = "A(B" },
		func(f *Font) { f.Family = "\x80" },
		func(f *Font) { f.Params[0] = minFix },
		func(f *Font) { f.Params[1] = 16 * Unity },
		func(f *Font) { f.Chars['a'].Width = 16 * Unity },
		func(f *Font) { f.Chars['a'].Width = -16 * Unity },
		func(f *Font) { f.Chars['a'].Tag = LigTag },
		func(f *Font) { f.Chars['a'].Tag, f.Chars['a'].NextLarger = ListTag, 200 },
		func(f *Font) { f.Chars['f'].Tag, f.Chars['f'].NextLarger = ListTag, 'a' },
		func(f *Font) { f.Chars['a'].Tag, f.Chars['a'].Recipe = ExtTag, Recipe{Rep: 200} },
		func(f *Font) { f.LigTable = append(f.LigTable, Label('f'), Krn('a', 0)) },
		func(f *Font) { f.LigTable = append(f.LigTable, Label('b')) },
		func(f *Font) { f.LigTable = append(f.LigTable, Stop()) },
		func(f *Font) { f.LigTable = append([]Instr{Skip(1)}, f.LigTable...) },
		func(f *Font) { f.LigTable[len(f.LigTable)-1] = Skip(200) },
		func(f *Font) { f.LigTable[len(f.LigTable)-1] = Skip(3) },
		func(f *Font) { f.LigTable[1].Step.Op = 4 },
		func(f *Font) { f.LigTable[1].Kind = 99 },
		func(f *Font) { f.LigTable[1].Step.Next = 200 },
		func(f *Font) { f.LigTable[1].Step.Result = 200 },
		func(f *Font) { f.RecipeOrder = []uint8{'a'} },
		func(f *Font) { f.Header = make([]uint32, 40000) },
		func(f *Font) { f.Params = make([]Fix, 40000) },
	}
	for i, fn := range bad {
		f := textFont()
		fn(f)
		var invalid *InvalidFontError
		if _, _, err := f.Build(); !errors.As(err, &invalid) {
			t.Errorf("case %d: Build gives %v, want an InvalidFontError", i, err)
		}
	}
}

// TestPrograms checks that Programs gives every character its steps.
func TestPrograms(t *testing.T) {
	a := []Step{{Next: 'x', Op: OpKern, Kern: Unity}, {Next: 'y', Op: OpLig, Result: 'z'}}
	b := []Step{{Next: 'x', Op: OpKern, Kern: -Unity}}
	progs := map[int][]Step{'p': a, 'q': b, 'r': a, Boundary: b, 's': nil}
	f := NewFont()
	for _, c := range "pqrsxyz" {
		f.Chars[c] = &Char{Width: Unity}
	}
	f.LigTable = Programs(progs)
	want := []Instr{Label('p'), Label('r'), Krn('x', Unity), Lig(OpLig, 'y', 'z'), Stop(),
		BoundaryLabel(), Label('q'), Krn('x', -Unity), Stop()}
	if !reflect.DeepEqual(f.LigTable, want) {
		t.Errorf("program %v, want %v", f.LigTable, want)
	}
	file, _, _ := build(t, f)
	for c := 0; c <= Boundary; c++ {
		if got := file.Program(c); !reflect.DeepEqual(got, progs[c]) {
			t.Errorf("program of %d is %v, want %v", c, got, progs[c])
		}
	}
}
