// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"fmt"
	"strconv"
	"time"
)

// Accepted range of SOURCE_DATE_EPOCH: from 1970-01-01 00:00:00 UTC to the
// end of the year 9999.
const maxEpoch = 253402300799

// SourceDate returns the time stamp for the font, in UTC.  The value is the
// content of the environment variable SOURCE_DATE_EPOCH, in seconds since
// 1970 in decimal.  If it is empty, the fallback is returned.
func SourceDate(epoch string, fallback time.Time) (time.Time, error) {
	if epoch == "" {
		return fallback.UTC(), nil
	}
	sec, err := strconv.ParseInt(epoch, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid SOURCE_DATE_EPOCH %q: want seconds since 1970 in decimal", epoch)
	}
	if sec < 0 || sec > maxEpoch {
		return time.Time{}, fmt.Errorf("invalid SOURCE_DATE_EPOCH %q: want a value from 0 (1970-01-01) to %d (end of the year 9999), in UTC", epoch, maxEpoch)
	}
	return time.Unix(sec, 0).UTC(), nil
}
