// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"fmt"
	"strings"
	"testing"
)

// bcharWord0PL has a boundary character and no program for the left boundary.
const bcharWord0PL = "(BOUNDARYCHAR C x)\n(LIGTABLE (LABEL C a) (KRN C x R 0.1) (STOP))\n(CHARACTER C a)\n(CHARACTER C x)\n"

// sevenBitPL has a ligature from characters below 128 to character 200.
const sevenBitPL = "(LIGTABLE (LABEL C a) (LIG C b D 200) (STOP))\n(CHARACTER C a)\n(CHARACTER C b)\n(CHARACTER D 200)\n"

// kernHighPL returns a PL text with 600 different kerns in three programs
// of 200 steps, so that kern indices run above 255 and the programs of the
// second and the third character start beyond word 255.
func kernHighPL() string {
	var b strings.Builder
	b.WriteString("(LIGTABLE\n")
	for p := 0; p < 3; p++ {
		fmt.Fprintf(&b, "(LABEL D %d)\n", 200+p)
		for i := 0; i < 200; i++ {
			fmt.Fprintf(&b, "(KRN D %d R %.3f)\n", i, 0.001*float64(p*200+i+1))
		}
		b.WriteString("(STOP)\n")
	}
	b.WriteString(")\n")
	for c := 0; c < 203; c++ {
		fmt.Fprintf(&b, "(CHARACTER D %d)\n", c)
	}
	return b.String()
}

// skip127PL returns a PL text whose first step skips 127 steps.
func skip127PL() string {
	var b strings.Builder
	b.WriteString("(LIGTABLE\n(LABEL D 200)\n(KRN D 0 R 0.001)\n(SKIP D 127)\n")
	for i := 1; i <= 127; i++ {
		fmt.Fprintf(&b, "(KRN D %d R %.3f)\n", i, 0.001*float64(i+1))
	}
	b.WriteString("(KRN D 128 R 0.5)\n(STOP)\n)\n")
	for c := 0; c <= 128; c++ {
		fmt.Fprintf(&b, "(CHARACTER D %d)\n", c)
	}
	b.WriteString("(CHARACTER D 200)\n")
	return b.String()
}

func extraFile(t *testing.T, name string) *File {
	t.Helper()
	for _, src := range plSources {
		if src.name == name {
			font, err := ReadPL(strings.NewReader(src.text))
			if err != nil {
				t.Fatalf("%s: ReadPL failed: %v", name, err)
			}
			file, _, _ := build(t, font)
			return file
		}
	}
	t.Fatalf("no source %s", name)
	return nil
}

// TestKernIndexAbove255 checks kerns with an index above 255 and programs
// which start beyond word 255.
func TestKernIndexAbove255(t *testing.T) {
	file := extraFile(t, "kernhigh")
	if len(file.Kern) != 600 {
		t.Fatalf("%d kerns, want 600", len(file.Kern))
	}
	for p := 0; p < 3; p++ {
		for _, i := range []int{0, 55, 56, 199} {
			want := file.Kern[p*200+i]
			if got, ok := file.Step(200+p, i); !ok || got.Op != OpKern || got.Kern != want {
				t.Errorf("step of %d and %d is %+v, %v, want the kern %v", 200+p, i, got, ok, want)
			}
		}
	}
	if file.Kern[599] != Units(600, 1000) {
		t.Errorf("last kern is %v", file.Kern[599])
	}
	if _, ok := file.Step(202, 200); ok {
		t.Errorf("a step for a character which no step names")
	}
}

// TestBoundaryCharWord0 checks a boundary character which only word 0 of the
// program names: the font has no program for the left boundary.
func TestBoundaryCharWord0(t *testing.T) {
	file := extraFile(t, "bcharword0")
	if c, ok := file.BoundaryChar(); !ok || c != 'x' {
		t.Errorf("boundary character is %d, %v", c, ok)
	}
	if s := file.LigKern[len(file.LigKern)-1]; s.Skip == 255 {
		t.Errorf("the last word starts a program of the left boundary")
	}
	if got, ok := file.Step('a', Boundary); !ok || got.Op != OpKern || got.Next != 'x' || got.Kern != Units(1, 10) {
		t.Errorf("step of a and the boundary character is %+v, %v", got, ok)
	}
	if _, ok := file.Step(Boundary, 'a'); ok {
		t.Errorf("the font has a program for the left boundary")
	}
	if got := file.Program(Boundary); got != nil {
		t.Errorf("program of the left boundary is %v", got)
	}
}

// TestStepSkip127 checks File.Step on a skip of 127 steps.
func TestStepSkip127(t *testing.T) {
	file := extraFile(t, "skip127")
	if got, ok := file.Step(200, 128); !ok || got.Op != OpKern || got.Kern != Units(1, 2) {
		t.Errorf("step of 200 and 128 is %+v, %v", got, ok)
	}
	if got, ok := file.Step(200, 0); !ok || got.Kern != Units(1, 1000) {
		t.Errorf("step of 200 and 0 is %+v, %v", got, ok)
	}
	for _, c := range []int{1, 64, 127} {
		if got, ok := file.Step(200, c); ok {
			t.Errorf("step of 200 and %d is %+v, which the skip passes over", c, got)
		}
	}
}

// TestSevenBitFlagAndLabelOffset checks, without the converters, the flag
// of the header and the indirect address of a label above 255.
func TestSevenBitFlagAndLabelOffset(t *testing.T) {
	if extraFile(t, "sevenbit").SevenBitSafe() {
		t.Errorf("a ligature to character 200 leaves the flag set")
	}
	if !extraFile(t, "bcharword0").SevenBitSafe() {
		t.Errorf("the flag is not set")
	}
	file := extraFile(t, "kernhigh")
	if i := int(file.Chars[202-file.First].Remainder); file.LigKern[i].Skip <= 128 {
		t.Errorf("the label of character 202 gives word %d directly", i)
	}
}
