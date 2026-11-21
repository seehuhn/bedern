// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package font

import (
	"strings"
	"testing"
)

// TestOFL checks the copyright line and the title of the embedded licence.
func TestOFL(t *testing.T) {
	t.Parallel()
	first, _, _ := strings.Cut(OFL, "\n")
	if want := "Copyright (c) 2026, Jochen Voss <voss@seehuhn.de>, with Reserved Font Name Bedern."; first != want {
		t.Errorf("first line %q, want %q", first, want)
	}
	if !strings.Contains(OFL, "SIL OPEN FONT LICENSE Version 1.1") {
		t.Error("the text has no title of the SIL Open Font License, Version 1.1")
	}
	if !strings.HasSuffix(OFL, "\n") {
		t.Error("the text does not end in a newline")
	}
}
