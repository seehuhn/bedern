// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"math"
	"strconv"
)

// Fix is a fix word: a signed number with 20 binary digits after the binary
// point.  The value of x is x/2^20.
type Fix int32

// Unity is the fix word 1.0.
const Unity Fix = 1 << 20

const (
	maxFix = Fix(math.MaxInt32)
	minFix = Fix(math.MinInt32)
)

// Units converts a length of n units into a fix word, where perEm units make
// 1.0.  The result is n·2^20/perEm, rounded to the nearest multiple of 2^-20;
// a value half way between two multiples is rounded away from zero.  A result
// outside the range of a fix word is replaced by the nearest fix word.  The
// argument perEm is positive.
//
// For a font whose design size is one em, Units(n, 1000) is the fix word of a
// length of n thousandths of an em.
func Units(n, perEm int) Fix {
	if perEm <= 0 {
		panic("tfm: the number of units per em is not positive")
	}
	neg := n < 0
	a := uint64(n)
	if neg {
		a = -a
	}
	p := uint64(perEm)
	if a > math.MaxUint64>>22 {
		if neg {
			return minFix
		}
		return maxFix
	}
	q := (a<<21 + p) / (2 * p)
	if neg {
		if q > 1<<31 {
			return minFix
		}
		return Fix(-int64(q))
	}
	if q > math.MaxInt32 {
		return maxFix
	}
	return Fix(q)
}

// FixFromFloat converts a number into a fix word.  The result is x·2^20,
// rounded to the nearest integer; a value half way between two integers is
// rounded away from zero.  A result outside the range of a fix word is
// replaced by the nearest fix word, and a NaN by zero.
func FixFromFloat(x float64) Fix {
	y := math.Round(x * (1 << 20))
	switch {
	case math.IsNaN(y):
		return 0
	case y >= math.MaxInt32:
		return maxFix
	case y <= math.MinInt32:
		return minFix
	}
	return Fix(y)
}

// Float returns the value of x as a floating point number.  The conversion is
// exact.
func (x Fix) Float() float64 {
	return float64(x) / (1 << 20)
}

// String returns the value of x in decimal notation, with the digits which
// TFtoPL prints: the shortest fraction from which PLtoTF recovers x.
func (x Fix) String() string {
	return string(appendFix(nil, x))
}

// appendFix appends the decimal notation of x to b.
func appendFix(b []byte, x Fix) []byte {
	u := uint32(x)
	a := int(u >> 20)
	f := int(u & 0xFFFFF)
	if a > 0x7FF {
		b = append(b, '-')
		a = 0x1000 - a
		if f > 0 {
			f = 0x100000 - f
			a--
		}
	}
	b = strconv.AppendInt(b, int64(a), 10)
	b = append(b, '.')
	f = 10*f + 5
	delta := 10
	for {
		if delta > 0x100000 {
			f += 0x80000 - delta/2
		}
		b = append(b, byte('0'+f/0x100000))
		f = 10 * (f % 0x100000)
		delta *= 10
		if f <= delta {
			break
		}
	}
	return b
}
