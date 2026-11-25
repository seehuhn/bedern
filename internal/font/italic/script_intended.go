// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "fmt"

// The findings of the script forms which are the intended findings of their
// text forms: the same inflection or wiggle of the same part, drawn with the
// heavier pen of the level and widened, so that it lies elsewhere than in
// the text form (BuildScript looks the findings of a script form up under
// "name/level").  Each place is given by the index of the entry of the text
// form whose reason it shares and by where the analysis finds it at the
// level; the kind is that of the text form.
func init() {
	type at struct {
		text      int
		x, y, dir float64
	}
	mark := func(name string, level int, places ...at) {
		key := fmt.Sprintf("%s/%d", name, level)
		for _, a := range places {
			// The files are initialised in the order of their names, so that
			// the entries of the text forms exist here.
			e := intendedFindings[name][a.text]
			intendedFindings[key] = append(intendedFindings[key], intended{
				Kind: e.Kind, X: a.x, Y: a.y, Dir: a.dir,
				Reason: e.Reason + fmt.Sprintf(" (the script form of level %d)", level),
			})
		}
	}
	mark("j", 1, at{0, 136, 378, -90})
	mark("k", 1, at{0, 274, 120, -90}, at{1, 351, 81, 80})
	mark("k", 2, at{0, 285, 112, -90}, at{1, 372, 82, 83})
	mark("s", 1, at{0, 183, 203, -17}, at{1, 208, 247, 163})
	mark("s", 2, at{0, 211, 199, -17}, at{1, 234, 255, 163})
	mark("kappa", 1, at{0, 266, 121, -90}, at{1, 343, 82, 82})
	mark("kappa", 2, at{0, 275, 110, -90}, at{1, 362, 85, 84})
	mark("kappaVar", 1, at{1, 96, 311, -109}, at{2, 62, 176, -99}, at{3, 266, 121, -90}, at{4, 343, 82, 82})
	mark("kappaVar", 2, at{1, 89, 306, -110}, at{2, 48, 159, -100}, at{3, 275, 110, -90}, at{4, 362, 85, 84})
	mark("nu", 1, at{0, 287, 140, 21})
	mark("nu", 2, at{0, 312, 139, 15})
	mark("thetaVar", 1, at{0, 62, 211, -104}, at{1, 590, 474, 71}, at{2, 141, 174, 76})
	mark("thetaVar", 2, at{0, 53, 198, -104}, at{2, 142, 169, 76})
	mark("mu", 1, at{0, 141, 77, 89})
	mark("mu", 2, at{0, 143, 78, 91})
	mark("upsilon", 1, at{3, 127, 247, 75})
	mark("upsilon", 2, at{0, 442, 299, -110}, at{1, 418, 226, -104}, at{2, 119, 213, 75}, at{3, 122, 284, 93})
	mark("psi", 1, at{2, 126, 243, 76})
	mark("psi", 2, at{0, 540, 299, -110}, at{1, 517, 228, -104}, at{2, 120, 215, 75}, at{3, 122, 282, 93})
	mark("phiLoop", 1, at{0, 103, 193, 75}, at{1, 113, 284, 88})
	mark("phiLoop", 2, at{0, 107, 202, 76}, at{1, 114, 294, 89})
	mark("omega", 1, at{0, 80, 364, -133}, at{2, 645, 364, -114}, at{3, 613, 270, -105}, at{1, 124, 275, 76})
	mark("omega", 2, at{0, 76, 368, -128}, at{2, 716, 357, -117}, at{3, 689, 286, -104}, at{1, 125, 274, 76})
	mark("piVar", 1, at{0, 125, 362, -133}, at{1, 172, 280, 76}, at{2, 692, 364, -114}, at{3, 660, 270, -105})
	mark("piVar", 2, at{1, 175, 274, 76}, at{3, 739, 286, -104})
	mark("zeta", 1, at{0, 249, 703, 69.7})
	mark("zeta", 2, at{0, 251, 698, 69.7})
	mark("xi", 1, at{0, 225, 703, 69.7})
	mark("xi", 2, at{0, 227, 698, 69.7})
}
