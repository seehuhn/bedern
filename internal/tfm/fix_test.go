// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"math/rand"
	"strings"
	"testing"
)

func TestUnitsRounding(t *testing.T) {
	cases := []struct {
		n, perEm int
		want     Fix
	}{
		{0, 7, 0},
		{1, 1, Unity},
		{1, 2, Unity / 2},
		{500, 1000, Unity / 2},
		{1, 3, 349525},
		{-1, 3, -349525},
		{2, 3, 699051},
		{-2, 3, -699051},
		{1, 1 << 21, 1},
		{-1, 1 << 21, -1},
		{3, 1 << 21, 2},
		{-3, 1 << 21, -2},
		{1, 1 << 22, 0},
		{2047, 1, 2047 * Unity},
		{5000, 1, maxFix},
		{-5000, 1, minFix},
	}
	for _, c := range cases {
		if got := Units(c.n, c.perEm); got != c.want {
			t.Errorf("Units(%d, %d) is %d, want %d", c.n, c.perEm, got, c.want)
		}
	}
}

func TestFixFromFloat(t *testing.T) {
	cases := []struct {
		x    float64
		want Fix
	}{
		{0, 0},
		{1, Unity},
		{-0.25, -Unity / 4},
		{0.5 / (1 << 20), 1},
		{-0.5 / (1 << 20), -1},
		{0.49 / (1 << 20), 0},
		{3000, maxFix},
		{-3000, minFix},
	}
	for _, c := range cases {
		if got := FixFromFloat(c.x); got != c.want {
			t.Errorf("FixFromFloat(%v) is %d, want %d", c.x, got, c.want)
		}
		if c.want != maxFix && c.want != minFix && FixFromFloat(c.want.Float()) != c.want {
			t.Errorf("the value of %d does not convert back", c.want)
		}
	}
}

// TestFixString checks the decimal notation of some fix words, and that
// ReadPL recovers every fix word from its notation.
func TestFixString(t *testing.T) {
	cases := []struct {
		x    Fix
		want string
	}{
		{0, "0.0"},
		{Unity, "1.0"},
		{-Unity, "-1.0"},
		{Unity / 2, "0.5"},
		{-Unity / 4, "-0.25"},
		{1, "0.000001"},
		{-1, "-0.000001"},
		{10 * Unity, "10.0"},
		{maxFix, "2047.999999"},
		{Unity / 3, "0.333333"},
		{Unity/3 + 1, "0.333334"},
	}
	for _, c := range cases {
		if got := c.x.String(); got != c.want {
			t.Errorf("the notation of %d is %s, want %s", c.x, got, c.want)
		}
	}
	rng := rand.New(rand.NewSource(1))
	values := []Fix{0, 1, -1, maxFix, minFix + 1, Unity - 1, Unity + 1}
	for i := 0; i < 2000; i++ {
		values = append(values, Fix(rng.Uint32()), Fix(rng.Intn(1<<21)-1<<20))
	}
	for _, x := range values {
		if x == minFix {
			continue
		}
		f, err := ReadPL(strings.NewReader("(FONTDIMEN (SLANT R " + x.String() + "))"))
		if err != nil {
			t.Fatalf("the notation %s of %d does not read: %v", x, int32(x), err)
		}
		if f.Params[0] != x {
			t.Errorf("the notation %s of %d reads as %d", x, int32(x), f.Params[0])
		}
	}
}
