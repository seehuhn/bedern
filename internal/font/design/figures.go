// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// This file holds the parameters of the figures 0 to 9 and the positions which
// follow from them.  The figures are lining and tabular: all ten stand on one
// advance, and each is centred in it with an optical shift of its own; they do
// not take part in the fit of the letters.  The construction is described in
// internal/glyphs/figures.go and the files of the figures.

// Figures holds the parameters of the figures.
type Figures struct {
	// Height is the height of the flat tops (1, 4, 7 and the flag of 5);
	// OverTop is how far the round tops (0, 2, 3, 6, 8 and 9) rise above it.
	// At the baseline the round figures reach OvershootUpper below it, as O
	// does.  Advance is the advance of every figure.
	Height  float64
	OverTop float64
	// SupWeight is the thickness of the strokes of a superior figure, in
	// strokes of the figure, and of the letters of © and ®, in strokes of
	// the letters.
	SupWeight float64
	Advance   float64

	// Strokes.  Stem is the stem of 1 and 4; Round the sides of 0 and Thin the
	// thin parts of the round figures (the top and bottom of 0, 3, 5, 6, 8 and
	// 9 and the necks of 2 and 3); Heavy the bar of 5 and of 7 and the base of
	// 2, which are heavy strokes and not thin parts; Bowl the thick parts of
	// the bowls of 3, 5, 6, 8 and 9; Ball the diameter of the ball terminal of
	// 2, 3, 5, 6 and 9, which is the part of the lowercase scaled from
	// BallDiameter to Ball, its stalk BallStalk with it.
	Stem  float64
	Round float64
	Thin  float64
	Heavy float64
	Bowl  float64
	Ball  float64

	// Ink is the width of the ink of each figure (but 1, whose ink follows from
	// its flag, stem and foot serif, and 9, which is 6 turned; see FigureInk),
	// and Shift how far the ink
	// stands right of the middle of the advance.
	Ink   [10]float64
	Shift [10]float64

	// 1.  The flag leaves the top of the stem at the right-hand edge and falls
	// to the left: OneFlag is its reach beyond the left-hand edge of the stem,
	// OneFlagDrop the depth of the middle of its tip below the top,
	// OneFlagThick its thickness at the tip, OneFlagAngle (degrees) the rise
	// of the flag from the tip towards the stem and OneFlagLean (degrees) the
	// fall of its upper edge where it leaves the top.  OneFlagJoin is how far
	// below the underside of the flag at the stem its lower edge runs into
	// the stem.  OneFoot is the jut of the foot serif on each side; the serif
	// is the bracketed foot of the capitals.  The top of 1 stands OneRise
	// above Height.
	OneFlag      float64
	OneFlagDrop  float64
	OneFlagThick float64
	OneFlagAngle float64
	OneFlagLean  float64
	OneFlagJoin  float64
	OneFoot      float64
	OneRise      float64

	// 2.  The upper bowl reaches from TwoBowlLeft (from the ink left; the
	// left extreme is cut off by the ball and lies outside the ink where the
	// value is negative) to the ink right; its left half is that of the oval
	// whose centre is TwoBowlMid above the baseline, and the right extreme of
	// its inner edge is TwoShoulder above the baseline; the right extreme of
	// its outer edge, the rightmost ink, is TwoOuterDrop lower.
	// TwoNeckThick is the thickness of its right side.  The outer edge of
	// the diagonal meets the top of the base TwoDiagFoot from the ink left,
	// at TwoDiagAngle (degrees from the vertical), and its inner edge meets
	// the left end of the base, at the ink left, TwoBaseEnd above the
	// baseline; the base is TwoBase long and ends in a vertical serif rising
	// TwoRise above it, TwoSerifThick wide.  The ball's leftmost ink is
	// TwoBallLeft from the ink left.
	TwoBowlLeft   float64
	TwoBowlMid    float64
	TwoShoulder   float64
	TwoOuterDrop  float64
	TwoNeckThick  float64
	TwoDiagFoot   float64
	TwoDiagAngle  float64
	TwoBase       float64
	TwoRise       float64
	TwoSerifThick float64
	TwoBallLeft   float64
	TwoBaseEnd    float64

	// 3.  ThreeWaist is the height of the middle, ThreeGap the thickness of
	// the middle stroke and ThreeOverlap how far the outer edges of the two
	// bowls reach past the middle into each other.  The lower bowl reaches
	// from ThreeLowerLeft (from the ink left; its left extreme is cut off by
	// the ball and lies outside the ink where the value is negative) to the
	// ink right.  The upper bowl is
	// ThreeUpperWidth wide and its right extreme ThreeUpperRight from the ink
	// left; ThreeUpperBowl is its thick part, Bowl that of the lower one.
	// ThreeOpen is how far the inner edge of a bowl lies inside the outer
	// one on the open side.  ThreeMiddle is the left end of the middle stroke from the ink
	// left, and ThreeUpperBall and ThreeLowerBall are the leftmost ink of the
	// two balls from the ink left.
	ThreeWaist      float64
	ThreeGap        float64
	ThreeOverlap    float64
	ThreeLowerLeft  float64
	ThreeUpperWidth float64
	ThreeUpperRight float64
	ThreeUpperBowl  float64
	ThreeOpen       float64
	ThreeMiddle     float64
	ThreeUpperBall  float64
	ThreeLowerBall  float64

	// The curled terminal of 6 and 9 is the part Ball across on a stalk
	// CurlStalk thick, turned so that the tip of its wedge lies CurlNeck from
	// the outer edge of the stroke; the centre of the ball stands CurlHeight
	// above the round bottom of 9 (on 6 below the round top).
	CurlHeight float64
	CurlStalk  float64
	CurlNeck   float64

	// 4.  The diagonal leans FourAngle (degrees) from the vertical and is
	// FourDiag thick; its outer edge runs from the left end of the bar, the ink
	// left, up to the top, where a flat FourTopFlat wide joins it to the
	// right-hand edge of the stem, which fixes the place of the stem.  The
	// bar is FourBar thick with its underside at FourBarLow and it reaches to
	// the right end of the ink; FourFoot is the jut of the foot serif on each
	// side of the stem.
	FourAngle   float64
	FourDiag    float64
	FourTopFlat float64
	FourBar     float64
	FourBarLow  float64
	FourFoot    float64

	// 5.  The vertical stroke is FiveStroke wide, its left-hand edge
	// FiveStrokeLeft from the ink left; the flag reaches FiveFlag from that
	// edge.  The bowl's outer edge is the oval from FiveOuterLeft (from the ink
	// left, negative where its left extreme lies outside the ink, where the
	// stroke cuts it off) to the ink right and from the overshoot below the
	// baseline to FiveBowlTop; its inner edge has its left extreme at
	// FiveInnerLeft.  The ball's leftmost ink is FiveBallLeft from the ink
	// left.
	FiveStroke     float64
	FiveStrokeLeft float64
	FiveFlag       float64
	FiveOuterLeft  float64
	FiveInnerLeft  float64
	FiveBowlTop    float64
	FiveBallLeft   float64

	// 6, and 9 which is 6 turned.  The bowl is the full width of the ink from
	// the overshoot below the baseline to SixBowlTop, and SixBowlTopThin thick
	// at the top; its counter has its left extreme SixCounterLeft from the ink
	// left.  The stem is the left side of an oval whose left extreme lies at
	// the ink left SixLeftY above the baseline, which is SixArcWidth wide and
	// reaches the round top; its inner edge lies SixStemThick inside it at the
	// left extreme and SixTopThin at the top.  From the top the edges open
	// into a hook whose outer edge ends with the radius SixHookRadius after a
	// turn of SixHookAngle (degrees); the outer extreme of the ball lies
	// SixBallIn inside the ink right.
	SixBowlTop     float64
	SixBowlTopThin float64
	SixCounterLeft float64
	SixLeftY       float64
	SixArcWidth    float64
	SixStemThick   float64
	SixTopThin     float64
	SixBallIn      float64
	SixHookRadius  float64
	SixHookAngle   float64

	// 7.  The bar carries a serif at its left end which hangs SevenDrop below
	// it, SevenSerifThick wide.  The diagonal leaves the bar at its right end
	// and ends on the baseline in a flat SevenFoot wide whose middle is
	// SevenFootX from the ink left; its edges lean SevenTopAngle from the
	// vertical at the bar and SevenFootAngle at the foot (degrees), and it is
	// SevenTopThick wide across under the bar.
	SevenDrop       float64
	SevenSerifThick float64
	SevenFoot       float64
	SevenFootX      float64
	SevenTopAngle   float64
	SevenFootAngle  float64
	SevenTopThick   float64

	// 8.  The spine is the ink between the counters: it lies between two
	// parallel lines EightGap apart about the point at the height EightWaist
	// in the middle of the ink, which fall to the right at EightSpine degrees
	// from the horizontal.  The upper bowl is EightUpperWidth wide, centred
	// over the lower one; the outer edge of the upper bowl ends EightWaistThin
	// below EightWaist and that of the lower bowl EightWaistThin above it.  The
	// sides of the bowls are EightUL, EightUR, EightLL and EightLR thick
	// (upper left, upper right, lower left, lower right), and the thin arcs at
	// the upper right and the lower left are EightJoin thick where they meet
	// the spine.
	EightWaist      float64
	EightGap        float64
	EightUpperWidth float64
	EightUL         float64
	EightUR         float64
	EightLL         float64
	EightLR         float64
	EightWaistThin  float64
	EightSpine      float64
	EightJoin       float64

	// Old holds the parameters of the old-style figures and of the
	// proportional figures.
	Old OldFigures
}

// regularFigures returns the parameters of the figures of Bedern Regular.
func regularFigures() Figures {
	return Figures{
		Height:    666,
		OverTop:   8,
		Advance:   490,
		SupWeight: 0.72,

		Stem:  83,
		Round: 94,
		Thin:  38,
		Heavy: 78,
		Bowl:  92,
		Ball:  90,

		//      0    1    2    3    4    5    6    7    8    9
		Ink:   [10]float64{416, 0, 395, 401, 458, 391, 408, 420, 407, 0},
		Shift: [10]float64{0, 0, -12, -14, 0, -12, 0, 0, 0, 0},

		OneFlag:      125,
		OneFlagDrop:  100,
		OneFlagThick: 40,
		OneFlagAngle: 8,
		OneFlagLean:  32,
		OneFlagJoin:  45,
		OneFoot:      118,
		OneRise:      15,

		TwoBowlLeft:   10,
		TwoBowlMid:    420,
		TwoShoulder:   476,
		TwoOuterDrop:  20,
		TwoNeckThick:  94,
		TwoDiagFoot:   128,
		TwoDiagAngle:  50,
		TwoBase:       392,
		TwoRise:       88,
		TwoSerifThick: 42,
		TwoBallLeft:   24,
		TwoBaseEnd:    24,

		ThreeWaist:      343,
		ThreeGap:        40,
		ThreeOverlap:    20,
		ThreeLowerLeft:  -56,
		ThreeUpperWidth: 385,
		ThreeUpperRight: 375,
		ThreeUpperBowl:  88,
		ThreeOpen:       60,
		ThreeMiddle:     122,
		ThreeUpperBall:  18,
		ThreeLowerBall:  0,

		CurlHeight: 132,
		CurlStalk:  31,
		CurlNeck:   35,

		FourAngle:   31,
		FourDiag:    36,
		FourTopFlat: 62,
		FourBar:     48,
		FourBarLow:  170,
		FourFoot:    97,

		FiveStroke:     44,
		FiveStrokeLeft: 49,
		FiveFlag:       315,
		FiveOuterLeft:  -82,
		FiveInnerLeft:  -82,
		FiveBowlTop:    412,
		FiveBallLeft:   0,

		SixBowlTop:     416,
		SixBowlTopThin: 42,
		SixCounterLeft: 86,
		SixLeftY:       300,
		SixArcWidth:    412,
		SixStemThick:   86,
		SixTopThin:     46,
		SixBallIn:      41,
		SixHookRadius:  260,
		SixHookAngle:   75,

		SevenDrop:       92,
		SevenSerifThick: 44,
		SevenFoot:       84,
		SevenFootX:      176,
		SevenTopAngle:   24,
		SevenFootAngle:  14,
		SevenTopThick:   66,

		EightWaist:      360,
		EightGap:        78,
		EightUpperWidth: 356,
		EightUL:         76,
		EightUR:         62,
		EightLL:         66,
		EightLR:         78,
		EightWaistThin:  14,
		EightSpine:      16,
		EightJoin:       36,

		Old: regularOldFigures(),
	}
}

// FigureInk returns the width of the ink of the figure d.
func FigureInk(p Params, d int) float64 {
	f := p.Figures
	switch d {
	case 1:
		return max(f.OneFlag, f.OneFoot) + f.Stem + f.OneFoot
	case 9:
		return f.Ink[6] // 9 is 6 turned
	}
	return f.Ink[d]
}

// FigureLeft returns the left side bearing of the figure d: the ink is
// centred in the advance and moved by its optical shift.
func FigureLeft(p Params, d int) float64 {
	f := p.Figures
	return (f.Advance-FigureInk(p, d))/2 + f.Shift[d]
}
