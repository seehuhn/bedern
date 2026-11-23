// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import "testing"

// TestBoldValidate checks that the bold parameters pass Validate.
func TestBoldValidate(t *testing.T) {
	t.Parallel()
	if err := Bold().Validate(); err != nil {
		t.Fatal(err)
	}
}
