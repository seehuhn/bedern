// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// This file holds the parameters of the basic punctuation and of the signs of
// ASCII mathematics: full stop, comma, colon, semicolon, exclamation mark,
// the apostrophe and the quotation marks, hyphen and dashes, parentheses and
// brackets, slash, backslash, vertical bar, underscore, asterisk, plus, minus,
// equals, less-than and greater-than, and the frame of .notdef.  The
// construction is in internal/glyphs (comma.go, dots.go, dash.go, bracket.go,
// slash.go, asterisk.go, mathsigns.go and notdef.go).
//
// Punctuation does not take part in the fit of the letters by side kinds: each
// sign has its own side spaces, a PunctSides value, measured to
// the outermost ink, and its advance is the ink between them.

// PunctSides holds the side spaces of a sign: the distances from the edge of the
// advance to the leftmost and to the rightmost ink.
type PunctSides struct {
	Left, Right float64
}

// Punct holds the parameters of the punctuation and the basic signs.
type Punct struct {
	// The dots.  The full stop is the dot of i (DotDiameter across, from the
	// same part), DotLow below the baseline at its lowest.  The upper dot of
	// the colon and the semicolon has its top ColonOver above the x-height.
	// ExclamTop is the height of the top of the stem of the exclamation mark
	// above the cap height; the stem is ExclamTopWidth wide at the top and
	// ExclamBottomWidth at its foot, which lies ExclamGap above the dot.
	DotLow            float64
	ColonOver         float64
	ExclamOver        float64
	ExclamTopWidth    float64
	ExclamBottomWidth float64
	ExclamGap         float64
	Period            PunctSides
	Colon             PunctSides
	Semicolon         PunctSides
	Exclam            PunctSides

	// The comma, in the frame in which the centre of the head is the origin.
	// The head is a circle of the diameter CommaHead; the lowest ink lies
	// CommaDepth below the baseline.  The outer (right) edge of the tail runs
	// into the head at its right extreme with the direction and the curvature
	// of the circle; the inner edge starts at a corner CommaInnerAt degrees
	// below the horizontal through the centre (the tip of the notch between the
	// tail and the head).  The tail ends in a
	// cut CommaTip wide whose centre lies CommaRun to the left of the centre of
	// the head, and it leaves in the direction CommaEndAngle degrees from the
	// vertical, to the left.  CommaOuterPull is the length of the handle of the
	// outer edge at the tip, and CommaInnerPull that of the inner edge at both
	// its ends, each as a fraction of the distance between the ends;
	// CommaInnerStart is the angle from the vertical at which the inner edge
	// leaves its corner, to the left.
	CommaHead       float64
	CommaDepth      float64
	CommaTip        float64
	CommaRun        float64
	CommaInnerAt    float64
	CommaEndAngle   float64
	CommaOuterPull  float64
	CommaInnerStart float64
	CommaInnerPull  float64
	Comma           PunctSides

	// The quotation marks.  The right single mark and the apostrophe are the
	// comma with its head's top at QuoteTop; the left one is that turned
	// through half a turn.  The double marks are two single marks QuotePitch
	// apart.  The straight marks are tapered bars StraightTopWidth wide at the
	// top, which is at StraightTop, and StraightBottomWidth at the foot,
	// StraightHeight high; the double one has two of them StraightPitch apart.
	QuoteTop            float64
	QuotePitch          float64
	StraightTop         float64
	StraightHeight      float64
	StraightTopWidth    float64
	StraightBottomWidth float64
	StraightPitch       float64
	Quote               PunctSides // the single marks: ' and the curly ones
	QuoteDouble         PunctSides // the curly double marks
	Straight            PunctSides // the straight single mark
	StraightDouble      PunctSides // the straight double mark

	// Hyphen and dashes.  The bars are plain, with flat ends: HyphenLength long
	// and HyphenThick thick with the centre at HyphenCentre; the en dash is
	// EnLength long, with the side space Dash, which gives it the advance of the
	// figures, and the em dash
	// EmLength with the side space EmDash, both DashThick thick with the centre at
	// DashCentre, which is the axis of the mathematics.  The underscore is
	// UnderLength long, UnderThick thick, its top UnderTop below the baseline.
	HyphenLength float64
	HyphenThick  float64
	HyphenCentre float64
	Hyphen       PunctSides
	EnLength     float64
	EmLength     float64
	DashThick    float64
	DashCentre   float64
	Dash         PunctSides
	EmDash       PunctSides
	UnderLength  float64
	UnderThick   float64
	UnderTop     float64

	// Parentheses and brackets stand BracketOver above the cap height and
	// BracketUnder below the descender line.  A parenthesis is the part of a stretched oval:
	// its outer edge reaches ParenSagitta from the line through its ends, the
	// strokes are ParenThick thick at the middle and ParenEnd thick (measured
	// horizontally) at the cut ends, and the oval is cut where its tangent is
	// ParenAngle degrees from the vertical.  A bracket has a stem BracketStem
	// wide and arms BracketArm thick which reach BracketReach beyond the inner
	// edge of the stem.
	BracketOver   float64
	BracketUnder  float64
	ParenSagitta  float64
	ParenThick    float64
	ParenEnd      float64
	ParenAngle    float64
	ParenLeft     PunctSides // the left parenthesis; the right one is its mirror image
	BracketStem   float64
	BracketArm    float64
	BracketReach  float64
	BracketLeft   PunctSides // the left bracket; the right one is its mirror image
	SlashAngle    float64
	SlashThick    float64
	SlashOver     float64 // the top of the slash above the cap height
	SlashUnder    float64 // its foot below the descender line
	Slash         PunctSides
	Backslash     PunctSides
	BarThick      float64
	Bar           PunctSides
	Underscore    PunctSides
	AsteriskTop   float64
	AsteriskSize  float64
	AsteriskInner float64
	AsteriskOuter float64
	Asterisk      PunctSides

	// The mathematical signs share one axis (the maths axis, MathAxis), one
	// stroke thickness, one ink width (MathWidth) and one advance: the plus is
	// a square of that width, the minus and the equals as wide, and the less-than
	// and greater-than MathAngleWidth wide with strokes at MathAngle degrees to
	// the horizontal, their vertex cut flat.  The two bars of the equals are
	// EqualsGap apart.
	MathAxis       float64
	MathThick      float64
	MathWidth      float64
	MathSide       float64 // the side space of every sign
	EqualsGap      float64
	MathAngle      float64
	MathAngleWidth float64
	MathVertexFlat float64

	// The frame of .notdef: NotdefThick thick, and the side space on both sides
	// of the ink; the glyph has the advance and the height of the figures.
	NotdefThick float64
	NotdefSide  float64
}

// regularPunct returns the punctuation parameters of Bedern Regular.
func regularPunct() Punct {
	return Punct{
		DotLow:            9,
		ColonOver:         4,
		ExclamOver:        4,
		ExclamTopWidth:    86,
		ExclamBottomWidth: 34,
		ExclamGap:         80,
		Period:            PunctSides{76, 76},
		Colon:             PunctSides{76, 76},
		Semicolon:         PunctSides{50, 76},
		Exclam:            PunctSides{76, 76},

		CommaHead:       110,
		CommaDepth:      168,
		CommaTip:        28,
		CommaRun:        55,
		CommaInnerAt:    90,
		CommaEndAngle:   50,
		CommaOuterPull:  0.38,
		CommaInnerStart: 20,
		CommaInnerPull:  0.1,
		Comma:           PunctSides{50, 76},

		QuoteTop:            698,
		QuotePitch:          190,
		StraightTop:         694,
		StraightHeight:      261,
		StraightTopWidth:    58,
		StraightBottomWidth: 30,
		StraightPitch:       156,
		Quote:               PunctSides{60, 60},
		QuoteDouble:         PunctSides{60, 60},
		Straight:            PunctSides{70, 70},
		StraightDouble:      PunctSides{70, 70},

		HyphenLength: 262,
		HyphenThick:  58,
		HyphenCentre: 228,
		Hyphen:       PunctSides{30, 31},
		EnLength:     410,
		EmLength:     940,
		DashThick:    52,
		DashCentre:   250,
		Dash:         PunctSides{40, 40},
		EmDash:       PunctSides{30, 30},
		UnderLength:  500,
		UnderThick:   46,
		UnderTop:     94,

		BracketOver:   57,
		BracketUnder:  21,
		ParenSagitta:  174,
		ParenThick:    76,
		ParenEnd:      46,
		ParenAngle:    48,
		ParenLeft:     PunctSides{78, 46},
		BracketStem:   74,
		BracketArm:    38,
		BracketReach:  112,
		BracketLeft:   PunctSides{76, 38},
		SlashAngle:    20,
		SlashThick:    52,
		SlashOver:     37,
		SlashUnder:    -54,
		Slash:         PunctSides{40, 40},
		Backslash:     PunctSides{40, 40},
		BarThick:      56,
		Bar:           PunctSides{106, 106},
		Underscore:    PunctSides{0, 0},
		AsteriskTop:   730,
		AsteriskSize:  340,
		AsteriskInner: 44,
		AsteriskOuter: 58,
		Asterisk:      PunctSides{57, 57},

		MathAxis:       250,
		MathThick:      54,
		MathWidth:      560,
		MathSide:       55,
		EqualsGap:      136,
		MathAngle:      25,
		MathAngleWidth: 515,
		MathVertexFlat: 36,

		NotdefThick: 38,
		NotdefSide:  40,
	}
}

// PeriodCentre returns the height of the centre of the full stop, which is that
// of the lower dot of the colon and of the head of the exclamation mark's dot.
func (p Params) PeriodCentre() float64 { return p.DotDiameter/2 - p.Punct.DotLow }

// ColonUpperCentre returns the height of the centre of the upper dot of the
// colon and the semicolon.
func (p Params) ColonUpperCentre() float64 {
	return p.XHeight + p.Punct.ColonOver - p.DotDiameter/2
}

// ExclamTop returns the height of the top of the stem of the exclamation mark.
func (p Params) ExclamTop() float64 { return p.CapHeight + p.Punct.ExclamOver }

// ExclamFoot returns the height of the foot of the stem of the exclamation
// mark: the gap above the dot.
func (p Params) ExclamFoot() float64 {
	return p.PeriodCentre() + p.DotDiameter/2 + p.Punct.ExclamGap
}

// BracketTop returns the height of the top of the parentheses and brackets.
func (p Params) BracketTop() float64 { return p.CapHeight + p.Punct.BracketOver }

// BracketBottom returns the height of the foot of the parentheses and
// brackets, which is negative.
func (p Params) BracketBottom() float64 { return -(p.Descender + p.Punct.BracketUnder) }

// SlashTop returns the height of the top of the slash and the backslash.
func (p Params) SlashTop() float64 { return p.CapHeight + p.Punct.SlashOver }

// SlashBottom returns the height of the foot of the slash and the backslash,
// which is negative.
func (p Params) SlashBottom() float64 { return -(p.Descender + p.Punct.SlashUnder) }

// MathAdvance returns the advance shared by plus, minus, equals, less-than and
// greater-than: the ink and a side space on each side.
func (p Params) MathAdvance() float64 { return p.Punct.MathWidth + 2*p.Punct.MathSide }
