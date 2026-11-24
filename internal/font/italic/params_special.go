// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

// Special are the parameters of the letters which are not the twenty-six of
// the basic alphabet (ß, æ, œ, ð, þ, ŋ, ĸ, ſ and the others), in the upright
// description.  Weights are shares of the pen, which WithStem moves;
// heights are given as shares of the x-height or of the ascender where they
// belong to them.
type Special struct {
	SzWidth      float64 // the set width of ß
	SzLower      float64 // the outer right extreme of the lower bowl of ß right of the right edge of the stem
	SzUpper      float64 // the outer right extreme of the upper bowl of ß right of the right edge of the stem
	SzWaistY     float64 // the height of the waist of ß, where the upper bowl ends and the lower begins, a share of the x-height
	SzWaistX     float64 // the start of the stroke at the waist right of the right edge of the stem
	SzWaistDir   float64 // the direction in which the stroke leaves the waist into the upper bowl, upright, in degrees
	SzLowerDir   float64 // the direction in which the lower bowl leaves the waist, upright, in degrees
	SzUpperSideY float64 // the height of the right extreme of the upper bowl, a share of the way from the waist to the top
	SzLowerSideY float64 // the height of the right extreme of the lower bowl, a share of the way from the bottom to the waist
	SzStemTop    float64 // the height up to which the stem of ß runs straight, a multiple of SlantDen
	SzCrest      float64 // the crest of ß right of the centre of the stem, a share of the way to the right extreme of the upper bowl
	SzEndGap     float64 // the white between the right edge of the stem and the ink of the end of the lower bowl, at the height of the end
	SzWaistW     float64 // the thickness of the stroke where it leaves the waist, a share of the thin of the pen
	SzBottomAt   float64 // the bottom node of the lower bowl of ß, a share of the way from the right edge of the stem to the node of the right extreme
	SzEndDrop    float64 // the centre of the end of the lower bowl of ß above the bottom of the letter
	SzEndDir     float64 // the direction in which the lower bowl of ß arrives at its end, upright, in degrees

	OeGap  float64 // the outer left extreme of the e of œ right of the inner right extreme of the o
	OeJoin float64 // the angle above and below the middle of the e of œ at which the o ends inside its left side, in degrees
	OeTurn float64 // how far the nodes at the top and the bottom of the o of œ stand right of those of o
	OeIn   float64 // how far the ends of the o of œ lie right of the centre line of the left side of the e

	AeGap     float64 // the place of the e of æ: the outer left extreme of its round, if it were drawn, right of the left edge of the stem of a
	AeStemEnd float64 // the height down to which the stem of æ runs straight before it turns into the bottom of the e, a multiple of SlantDen
	AeTopIn   float64 // how far below AeStemEnd the turn runs vertically in the italic
	AeTopJoin float64 // the height at which the top of the e of æ runs into the stem
	AeTopDir  float64 // the direction in which it runs in, upright, in degrees
	AeLowJoin float64 // the height at which the bowl of a runs into the stem of æ
	AeLowDir  float64 // the direction in which it runs in, upright, in degrees
	AeBarRun  float64 // EBarRun of the e of æ, whose right side runs into the stem

	LigPitch             float64 // the distance between the centre lines of the stems of a drawn ligature of f
	LigBowlPitch         float64 // LigPitch of the ligatures into b, h and k
	LigCrestAt           float64 // the crest of the bridge of a ligature, a share of the pitch right of the stem of f
	LigFJoin             float64 // the end of the bridge into a second f right of the centre line of its stem
	LigFDrop             float64 // how far that end lies below the crest
	LigFDir              float64 // the direction in which the bridge into a second f runs into its stem, upright, in degrees
	LigFCrestAt          float64 // the crest of the bridge into a second f, a share of the pitch right of the stem of the first
	LigFLJoin, LigFLDrop float64 // LigFJoin and LigFDrop where the second f runs on into an ascender
	LigVert              float64 // the height above the join of the stem at which the bridge into an ascender runs vertically in the italic
	IJCloser             float64 // how much closer the j of ĳ stands to the i than after its advance
	LDotRight            float64 // the white between the right edge of the stem of ŀ and its dot, at the height of the dot
	LDotY                float64 // the height of the centre of the dot of ŀ, a share of the x-height
	LDotMore             float64 // how much wider ŀ is than l
}

// defaultSpecial returns the starting values of the special letters.
func defaultSpecial() Special {
	return Special{
		SzWidth:      490,
		SzLower:      286,
		SzUpper:      236,
		SzWaistY:     0.81,
		SzWaistX:     70,
		SzWaistDir:   20,
		SzLowerDir:   -15,
		SzUpperSideY: 0.52,
		SzLowerSideY: 0.48,
		SzStemTop:    584,
		SzCrest:      0.40,
		SzEndGap:     40,
		SzWaistW:     1.4,
		SzBottomAt:   0.6,
		SzEndDrop:    80,
		SzEndDir:     120,

		OeGap:  0,
		OeJoin: 45,
		OeTurn: 28,
		OeIn:   -4,

		AeGap:     10,
		AeStemEnd: 160,
		AeTopIn:   32,
		AeTopJoin: 350,
		AeTopDir:  235,
		AeLowJoin: 128,
		AeLowDir:  57,
		AeBarRun:  0.58,

		LigPitch:     245,
		LigBowlPitch: 270,
		LigCrestAt:   0.66,
		LigFJoin:     0,
		LigFDrop:     90,
		LigFDir:      -60,
		LigFCrestAt:  0.5,
		LigFLJoin:    8, LigFLDrop: 66,
		LigVert: 64,

		IJCloser:  20,
		LDotRight: 52,
		LDotY:     0.79,
		LDotMore:  60,
	}
}
