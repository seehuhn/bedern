// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"testing"
	"time"
)

func TestSourceDate(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

	got, err := SourceDate("1700000000", now)
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Unix(1700000000, 0); !got.Equal(want) {
		t.Errorf("valid value: got %v, want %v", got, want)
	}

	got, err = SourceDate("", now)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(now) {
		t.Errorf("empty value: got %v, want %v", got, now)
	}

	for _, bad := range []string{"abc", "12.5", "-", "1e9", " 5", "99999999999999999999",
		"-1", "-86400", "253402300800", "9223372036854775807"} {
		if _, err := SourceDate(bad, now); err == nil {
			t.Errorf("malformed value %q: no error", bad)
		}
	}
}

// TestSourceDateUTC changes the local time zone, a global, and does not run
// in parallel with the other tests.
func TestSourceDateUTC(t *testing.T) {
	old := time.Local
	time.Local = time.FixedZone("far", 14*3600)
	t.Cleanup(func() { time.Local = old })

	// 2023-11-14 23:59:59 UTC is already 15 November in the local zone.
	got, err := SourceDate("1700006399", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if got.Location() != time.UTC {
		t.Errorf("location %v, want UTC", got.Location())
	}
	if y, m, d := got.Date(); y != 2023 || m != time.November || d != 14 {
		t.Errorf("date %v, want 2023-11-14", got)
	}

	got, err = SourceDate("", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if got.Location() != time.UTC {
		t.Errorf("fallback location %v, want UTC", got.Location())
	}
	if _, err := SourceDate("253402300799", time.Now()); err != nil {
		t.Errorf("last second of 9999: %v", err)
	}
	if _, err := SourceDate("0", time.Now()); err != nil {
		t.Errorf("zero: %v", err)
	}
}
