// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

// Package font holds the licence of the Bedern fonts: the text of the SIL
// Open Font License and the entries of the name table which state it.
package font

import _ "embed"

// OFL is the text of the SIL Open Font License, Version 1.1, with the
// copyright notice of the Bedern fonts.  It is written to OFL.txt beside
// the font files.
//
//go:embed OFL.txt
var OFL string

// LicenseFile is the name of the file which holds OFL.
const LicenseFile = "OFL.txt"

// Copyright is the copyright notice of the name table.
const Copyright = "Copyright 2026 Jochen Voss <voss@seehuhn.de>"

// License is the licence description of the name table.
const License = "This Font Software is licensed under the SIL Open Font License, Version 1.1. " +
	"This license is available with a FAQ at: https://openfontlicense.org"

// LicenseURL is the licence URL of the name table.
const LicenseURL = "https://openfontlicense.org"
