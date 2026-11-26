// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "math"

// GreekParams are the proportions of the Greek lowercase, in the upright description:
// the lengths, heights, directions and shares which the constructions of the
// letters (greek_*.go) use, as parameters; the script forms have values
// of their own (script_tuned.go) as the Latin letters have.
type GreekParams struct {
	ThetaWidth       float64 // the bowl of θ between its outer extremes
	ThetaBar         float64 // the centre line of the bar of θ, over the x-height
	ThetaTurn        float64 // the top node of θ right of the middle, the lowest node left of it
	PhiWidth         float64 // the bowl of ϕ
	PhiLow           float64 // how far the stem of ϕ stays below the ascender
	RhoWidth         float64 // the bowl of ρ
	RhoLeft          float64 // the left of the bowl of ρ right of that of o, for the descender which leans out to the left
	RhoStemIn        float64 // the centre line of the stem of ρ right of the middle of the left side
	RhoClose         float64 // the height of the start of the stroke inside the stem
	RhoCloseDir      float64 // its direction
	RhoJoin          float64 // the height from which the stem runs straight
	AlphaBowl        float64 // the crossing of α right of the left of the bowl, on the centre line
	AlphaTurn        float64 // the top node of the bowl left of its middle, the lowest node right of it
	AlphaTip         float64 // the tip right of the crossing
	AlphaTipDir      float64 // the direction in which the stroke leaves the tip
	AlphaCross       float64 // the height of the crossing
	AlphaDown        float64 // the direction of the stroke from the top through the crossing
	AlphaFoot        float64 // the lowest node of the foot right of the crossing
	AlphaFlick       float64 // the ink of the end of the foot right of its lowest node
	AlphaSide        float64 // the side space right of the tip or the foot
	AlphaRun         float64 // the straight run of the stroke below the crossing
	LambdaLeft       float64 // the left edge of the stem under the entry of λ right of that of n
	LambdaDir        float64 // the direction of the long stroke of λ, against the slant
	LambdaOut        float64 // the long stroke at greekLambdaHigh right of the centre line of the stem
	LambdaHigh       float64 // the height from which the long stroke runs straight
	LambdaSide       float64 // the side space right of the foot
	LambdaLegY       float64 // the height at which the leg leaves the long stroke
	LambdaLegDir     float64 // the direction of the leg, upward
	LambdaLegW       float64 // the thickness of the leg, in stems
	NuFoot           float64 // the height of the end node of the stem of ν
	NuCut            float64 // the direction of the cut at the foot of ν
	NuWidth          float64 // the free side of ν right of the centre line of its stem
	NuRiseW          float64 // the thickness of the rise of ν at the foot, in hairlines
	NuRiseDir        float64 // its direction there
	NuSide           float64 // the height of the node of the free side, over the x-height
	NuSideW          float64 // the thickness there, in stems
	NuBack           float64 // the end of ν left of the node of its free side
	NuEndDir         float64 // the direction of the end
	NuSideRight      float64 // the side space right of the free side
	ChiLeft          float64 // the left of χ against that of o
	ChiHook          float64 // the top of the hook of χ right of the left of its ink
	ChiRun           float64 // the heavy stroke at greekChiHigh right of the top of the hook
	ChiHigh          float64 // the height from which the heavy stroke runs straight
	ChiDir           float64 // its direction, against the slant
	ChiThin          float64 // the light stroke, in stems
	ChiThinDir       float64 // its direction, upward
	ChiCross         float64 // the height at which the strokes cross
	ChiSide          float64 // the side space right of the ink
	PsiWidth         float64 // ψ between the outer extremes of its cup
	PsiTop           float64 // the top of the stem of ψ
	PsiStemShift     float64 // the stem right of the middle of the cup
	KappaVarBow      float64 // the left extreme of the stroke of ϰ left of the centre line of the stem of κ
	KappaVarBowY     float64 // its height
	KappaVarEnd      float64 // the end of the curl left of the centre line
	KappaVarEndUp    float64 // its centre above the baseline, over half its thickness
	KappaVarEndDir   float64 // its direction, down from the horizontal to the left
	KappaVarEndW     float64 // its thickness over that of the end of the exit
	HiddenStart      float64 // the thickness of a stroke which starts inside another, over that of the other there
	DeltaWidth       float64 // the bowl of δ
	DeltaStart       float64 // the start of the neck right of the middle of the bowl
	DeltaStartDir    float64 // its direction
	DeltaNeck        float64 // the left extreme of the neck left of the middle of the bowl
	DeltaNeckY       float64 // its height
	DeltaCrest       float64 // the crest of the neck right of the middle of the bowl
	DeltaBall        float64 // the centre of the ball below the tangent at the crest
	EpsUpper         float64 // the width of the upper round of ε
	EpsUpperIn       float64 // its left extreme right of that of the lower round
	EpsLower         float64 // the width of the lower round
	EpsWaist         float64 // the height of the waist, over the x-height
	EpsWaistDir      float64 // the direction of the end of the upper stroke at the waist
	EpsWaistW        float64 // the thickness there, in thins of the pen
	EpsSpur          float64 // the end of the upper stroke right of the middle of the lower round
	EpsJoin          float64 // how far inside the end of the upper stroke the lower one starts
	EpsBallLow       float64 // the lowest ink of the ball below the top
	EpsTip           float64 // the rightmost ink of the end of the lower stroke right of the left of the ink
	PhiLoopWidth     float64 // the round of φ
	PhiLoopStart     float64 // the start of φ right of the left of the round
	PhiLoopStartDir  float64 // its direction
	PhiLoopSide      float64 // the height of the left extreme, over the x-height
	PhiLoopStem      float64 // the stem right of the middle of the round
	PhiLoopTop       float64 // the top node right of the stem
	PhiLoopJoin      float64 // the height from which the stem runs straight
	OmegaWidth       float64 // ω between its outer extremes
	OmegaTip         float64 // the tips inside the outer extremes
	OmegaTipDir      float64 // the direction of the right tip, and that of the left one reversed
	OmegaSide        float64 // the height of the outer extremes, over the x-height
	OmegaWaist       float64 // the height of the waist, over the x-height
	OmegaWaistDir    float64 // the direction of the left stroke at the waist
	OmegaJoin        float64 // the start of the right stroke back from the end of the left one, on its centre line
	BetaLeft         float64 // the stem of β against that of n
	BetaUpper        float64 // the outer right of the upper bowl right of the stem
	BetaLower        float64 // the outer right of the lower bowl right of the stem
	BetaWaist        float64 // the height of the waist
	BetaWaistIn      float64 // the start of the upper stroke at the waist right of the stem's centre line
	BetaWaistDir     float64 // its direction
	BetaWaistW       float64 // the thickness at the waist, in thins of the pen
	BetaJoinBack     float64 // how far inside the upper stroke the lower one starts
	BetaTopShift     float64 // the top node of the upper bowl right of its middle
	BetaJoin         float64 // the height from which the stem runs straight
	BetaLowSide      float64 // the height of the right extreme of the lower bowl, over the waist
	BetaClose        float64 // the height of the end of the lower bowl inside the stem
	BetaCloseDir     float64 // its direction
	GammaLean        float64 // the turn of the down stroke of γ to the right from the vertical, in degrees after the shear
	GammaBow         float64 // the turn of its direction to the right from the top to the crossing, in degrees
	GammaWeight      float64 // the factor of all thicknesses of γ
	GammaFall        float64 // the place of the fastest thinning, as a fraction of the down stroke
	GammaDownShift   float64 // the last node of the down stroke before the crossing
	GammaCrossShift  float64 // the first node of the rise beyond the crossing
	GammaCrossY      float64 // the height of the crossing of the loop
	GammaLoopW       float64 // the thickness of the loop, over that of the pen across a horizontal stroke
	GammaLoopTurn    float64 // the turn of the loop to the right from pointing straight down, in degrees
	GammaEye         float64 // the white between the strands of the loop
	GammaTurnIn      float64 // the radius of the turn less half the distance of the strands
	GammaTurnDir     float64 // the direction of the strands at the start and the end of the turn
	GammaRiseDir     float64 // the direction of the rise at the crossing
	GammaRiseChord   float64 // the direction from the start of the rise to the right side of the flourish
	GammaRiseW       float64 // its thickness, over that of the pen across a horizontal stroke
	GammaFlW         float64 // the thickness of the flourish at the top, over that of the pen across a horizontal stroke
	GammaFlTop       float64 // the top of its outer edge above the x-height
	GammaFlH         float64 // the height of the eye of its loop
	GammaFlEye       float64 // the width of the eye of its loop
	GammaFlBottom    float64 // the bottom of the loop right of its middle
	GammaLeg         float64 // the node of the leg right of the bottom of the loop
	GammaLegUp       float64 // its height above that bottom
	GammaLegDir      float64 // its direction
	GammaLegX        float64 // the end right of the bottom of the loop
	GammaLegY        float64 // its height above that bottom
	GammaEndDir      float64 // the direction at the end
	GammaEndW        float64 // the thickness at the end, over that of the pen across a horizontal stroke
	GammaSpan        float64 // the longest span between two nodes along the centre line
	GammaSide        float64 // the side space right of the leg
	SigmaBar         float64
	SigmaClose       float64
	SigmaDip         float64
	SigmaCloseDir    float64
	SigmaFall        float64
	SigmaFallDir     float64
	IotaLeftLess     float64 // how much less the left side of ι is than that of n
	MuDir            float64 // its direction there
	MuStart          float64 // the height of the start of the turn of μ inside its stem
	MuInside         float64 // how much thinner than the stem the turn of μ starts there
	BarW             float64 // the bar of τ and π, in thins of the pen
	BarDrop          float64 // how far below the bar its start lies
	BarRise          float64 // how far right of its start the bar runs level
	BarStartDir      float64 // the direction of the start of the bar
	TauWidth         float64 // the bar of τ, from the left of its ink to its end
	TauStem          float64 // the left edge of the stem of τ right of the left of the bar
	PiLeft           float64 // the left edge of the left leg of π right of the left of the bar
	PiCounter        float64 // the white between the legs of π
	PiOver           float64 // the bar of π beyond the right edge of its right leg
	UpsilonLeft      float64 // the left of υ right of that of o
	UpsilonWidth     float64 // υ between its outer extremes
	UpsilonTip       float64 // the tip at the upper left right of the left extreme
	UpsilonTipDir    float64 // the direction of the tip, reversed
	UpsilonSide      float64 // the height of the extremes, over the x-height
	UpsilonBack      float64 // the end of υ left of the right extreme
	UpsilonEndDir    float64 // the direction of the end of υ
	UpsilonRaise     float64 // how far the end of υ stands above the x-height less half its thickness
	LunateWidth      float64 // the round of ϵ
	LunateTip        float64 // the rightmost ink of its ends right of the left of the ink
	LunateTop        float64 // the top end of ϵ below the top, as the end of the exit stands above the bottom
	LunateBar        float64 // the height of the bar, over the x-height
	LunateBarEnd     float64 // the end of the bar right of the left of the ink
	HookRunX         float64 // the start of the straight run of the hook right of the middle of the round
	HookRunY         float64 // its height
	HookRunDir       float64 // its direction
	HookRun          float64 // its length
	HookSide         float64 // the right extreme of the hook right of the middle of the round
	HookSideY        float64 // its height
	HookLow          float64 // the lowest node of the hook right of the middle of the round
	HookEnd          float64 // the end of the hook right of the middle of the round
	HookEndUp        float64 // the end of the hook above the depth of the tails
	HookEndDir       float64 // the direction of the end, up from the horizontal to the left
	SigmaFinalLow    float64 // the bottom of the round of ς, above the baseline
	ZetaWidth        float64 // the round of ζ and ξ
	ZetaSide         float64 // the height of the node of the back of ζ, over the x-height
	ZetaSideX        float64 // the node of the back right of the left extreme of a round
	ZetaSideDir      float64 // its direction
	ZetaSideW        float64 // the thickness of the back there, over that of the pen across a vertical stroke
	ZetaRunW         float64 // the thickness of the run of the hook of ζ and ξ, over that of the pen across a vertical stroke
	ZetaCross        float64 // the crossing of the loop of ζ right of the node of the back, after the shear
	ZetaCrossShift   float64 // the node of the lower strand of a loop along it from the crossing
	ZetaCrossIn      float64 // the direction of the lower strand at the crossing, down from the horizontal
	ZetaCrossOut     float64 // the direction of the upper strand at the crossing, down from the horizontal to the left
	ZetaLoopY        float64 // the height of the lower strand of the loop of ζ at the turn
	ZetaLoop         float64 // the length of the loop of ζ from the crossing to the turn
	ZetaEye          float64 // the height of the eye of a loop, the distance of the strands less a thin
	ZetaTurnIn       float64 // the radius of the turn of a loop less than half the distance of the strands
	ZetaTurnDir      float64 // the direction of the strands at the start and the end of the turn
	ZetaLoopW        float64 // the thickness of a loop at its turn, over that of the pen across a horizontal stroke
	ZetaLoopTaper    float64 // the change of that thickness from one node of a loop to the next
	ZetaStart        float64 // the start of ζ right of the crossing, after the shear
	ZetaStartRise    float64 // its height above the ascender less half the thickness of the end of the exit
	ZetaStartDir     float64 // its direction, after the shear
	ZetaStartW       float64 // its thickness, over that of the pen across a horizontal stroke
	ZetaOver         float64 // the advance of ζ and ξ beyond the right side space of the round
	ZetaHookSide     float64 // the right extreme of the hook of ζ and ξ right of that of ς
	ZetaHookSideY    float64 // its height above that of ς
	ZetaHookSideW    float64 // its thickness over that of the pen across a vertical stroke, or 0 for that of the hook of ς
	ZetaHookLow      float64 // the lowest node of the hook of ζ and ξ right of that of ς
	XiHead           float64 // the crossing of the upper loop of ξ right of that of ζ
	XiSide           float64 // the height of the node of the back of ξ, over the x-height
	XiBowX           float64 // the left extreme of the bow of ξ right of that of a round
	XiBowY           float64 // its height
	XiBowDir         float64 // its direction
	XiBowW           float64 // the thickness of the bow there, over that of the pen across a vertical stroke
	XiWaist          float64 // the height of the crossing of the lower loop of ξ
	XiCross          float64 // that crossing right of the crossing of the upper loop, after the shear
	XiCrossShift     float64 // the nodes of both strands of the lower loop along them from the crossing
	XiLoop           float64 // the length of the lower loop of ξ from the crossing to the turn
	RhoVarSide       float64 // the height of the left extreme of the hook of ϱ
	RhoVarLow        float64 // its lowest node left of the middle of the round
	RhoVarEnd        float64 // its end right of the middle of the round
	RhoVarEndUp      float64 // the end above the depth of the tails
	RhoVarEndDir     float64 // the direction of the end, up from the horizontal to the right
	PiVarOver        float64 // how far the bar of ϖ reaches beyond the ω
	ThetaVarLeft     float64 // the stem under the entry of ϑ left of that of n
	ThetaVarTopK     float64 // the curvature of the centre line at the top node of the entry of ϑ, turning right
	ThetaVarIn       float64 // the left extreme of ϑ left of the centre line of the stem under its entry
	ThetaVarLowSide  float64 // the height of the left extreme
	ThetaVarWidth    float64 // the outer edge of the right side right of the centre line of the stem
	ThetaVarSide     float64 // the height of the node of the right side
	ThetaVarLoop     float64 // the left side of the loop left of the outer edge of the right side
	ThetaVarLoopSide float64 // the height of the left side of the loop
	ThetaVarEnd      float64 // the end right of the outer edge of the right side
	ThetaVarEndY     float64 // its height
	ThetaVarEndDir   float64 // its direction
	ThetaVarEndW     float64 // its thickness over that of the end of the exit
}

// defaultGreek returns the starting values of the Greek letters.
func defaultGreek() GreekParams {
	return GreekParams{
		ThetaWidth:       400.0,
		ThetaBar:         0.78,
		ThetaTurn:        4.0,
		PhiWidth:         470.0,
		PhiLow:           20.0,
		RhoWidth:         377.0,
		RhoLeft:          30.0,
		RhoStemIn:        4.0,
		RhoClose:         60.0,
		RhoCloseDir:      -40.0,
		RhoJoin:          240.0,
		AlphaBowl:        330.0,
		AlphaTurn:        10.0,
		AlphaTip:         100.0,
		AlphaTipDir:      200.0,
		AlphaCross:       230.0,
		AlphaDown:        -72.0,
		AlphaFoot:        125.0,
		AlphaFlick:       70.0,
		AlphaSide:        20.0,
		AlphaRun:         40.0,
		LambdaLeft:       -80.0,
		LambdaDir:        -56.0,
		LambdaOut:        80.0,
		LambdaHigh:       580.0,
		LambdaSide:       20.0,
		LambdaLegY:       380.0,
		LambdaLegDir:     72.0,
		LambdaLegW:       0.6,
		NuFoot:           -7.0,
		NuCut:            35.0,
		NuWidth:          250.0,
		NuRiseW:          1.12,
		NuRiseDir:        72.0,
		NuSide:           0.64,
		NuSideW:          0.46,
		NuBack:           17.0,
		NuEndDir:         101.0,
		NuSideRight:      30.0,
		ChiLeft:          -40.0,
		ChiHook:          60.0,
		ChiRun:           90.0,
		ChiHigh:          400.0,
		ChiDir:           -60.0,
		ChiThin:          0.5,
		ChiThinDir:       59.0,
		ChiCross:         160.0,
		ChiSide:          20.0,
		PsiWidth:         480.0,
		PsiTop:           652.0,
		PsiStemShift:     0.0,
		KappaVarBow:      5.0,
		KappaVarBowY:     210.0,
		KappaVarEnd:      63.0,
		KappaVarEndUp:    10.0,
		KappaVarEndDir:   26.0,
		KappaVarEndW:     15.0,
		HiddenStart:      0.64,
		DeltaWidth:       380.0,
		DeltaStart:       20.0,
		DeltaStartDir:    130.0,
		DeltaNeck:        60.0,
		DeltaNeckY:       580.0,
		DeltaCrest:       20.0,
		DeltaBall:        90.0,
		EpsUpper:         318.0,
		EpsUpperIn:       19.0,
		EpsLower:         325.0,
		EpsWaist:         0.55,
		EpsWaistDir:      -4.0,
		EpsWaistW:        1.26,
		EpsSpur:          -19.0,
		EpsJoin:          25.0,
		EpsBallLow:       138.0,
		EpsTip:           330.0,
		PhiLoopWidth:     500.0,
		PhiLoopStart:     60.0,
		PhiLoopStartDir:  240.0,
		PhiLoopSide:      0.45,
		PhiLoopStem:      30.0,
		PhiLoopTop:       50.0,
		PhiLoopJoin:      320.0,
		OmegaWidth:       607.0,
		OmegaTip:         60.0,
		OmegaTipDir:      110.0,
		OmegaSide:        0.68,
		OmegaWaist:       0.64,
		OmegaWaistDir:    88.0,
		OmegaJoin:        40.0,
		BetaLeft:         -20.0,
		BetaUpper:        200.0,
		BetaLower:        270.0,
		BetaWaist:        400.0,
		BetaWaistIn:      70.0,
		BetaWaistDir:     10.0,
		BetaWaistW:       1.2,
		BetaJoinBack:     10.0,
		BetaTopShift:     -20.0,
		BetaJoin:         520.0,
		BetaLowSide:      0.5,
		BetaClose:        50.0,
		BetaCloseDir:     135.0,
		GammaLean:        7.0,
		GammaBow:         12.0,
		GammaWeight:      1.1,
		GammaFall:        0.75,
		GammaDownShift:   13.0,
		GammaCrossShift:  30.0,
		GammaCrossY:      46.0,
		GammaLoopW:       0.88,
		GammaLoopTurn:    10.0,
		GammaEye:         24.0,
		GammaTurnIn:      0.0,
		GammaTurnDir:     10.0,
		GammaRiseDir:     58.0,
		GammaRiseChord:   70.0,
		GammaRiseW:       0.94,
		GammaFlW:         0.85,
		GammaFlTop:       0.0,
		GammaFlH:         41.0,
		GammaFlEye:       23.0,
		GammaFlBottom:    5.0,
		GammaLeg:         49.0,
		GammaLegUp:       13.0,
		GammaLegDir:      38.0,
		GammaLegX:        85.0,
		GammaLegY:        68.0,
		GammaEndDir:      73.0,
		GammaEndW:        0.77,
		GammaSpan:        150.0,
		GammaSide:        30.0,
		SigmaBar:         480.0,
		SigmaClose:       90.0,
		SigmaDip:         4.0,
		SigmaCloseDir:    145.0,
		SigmaFall:        6.0,
		SigmaFallDir:     0.0,
		IotaLeftLess:     40.0,
		MuDir:            -65.0,
		MuStart:          110.0,
		MuInside:         20.0,
		BarW:             1.3,
		BarDrop:          40.0,
		BarRise:          70.0,
		BarStartDir:      50.0,
		TauWidth:         440.0,
		TauStem:          170.0,
		PiLeft:           90.0,
		PiCounter:        210.0,
		PiOver:           50.0,
		UpsilonLeft:      10.0,
		UpsilonWidth:     400.0,
		UpsilonTip:       40.0,
		UpsilonTipDir:    70.0,
		UpsilonSide:      0.45,
		UpsilonBack:      80.0,
		UpsilonEndDir:    110.0,
		UpsilonRaise:     20.0,
		LunateWidth:      330.0,
		LunateTip:        290.0,
		LunateTop:        120.0,
		LunateBar:        0.5,
		LunateBarEnd:     220.0,
		HookRunX:         -26.0,
		HookRunY:         43.0,
		HookRunDir:       -21.0,
		HookRun:          70.0,
		HookSide:         151.0,
		HookSideY:        -90.0,
		HookLow:          50.0,
		HookEnd:          -80.0,
		HookEndUp:        64.0,
		HookEndDir:       35.0,
		SigmaFinalLow:    20.0,
		ZetaWidth:        340.0,
		ZetaSide:         0.45,
		ZetaSideX:        -15.0,
		ZetaSideDir:      275.0,
		ZetaSideW:        0.7149,
		ZetaRunW:         0.7539,
		ZetaCross:        230.0,
		ZetaCrossShift:   -8.0,
		ZetaCrossIn:      34.0,
		ZetaCrossOut:     18.0,
		ZetaLoopY:        582.0,
		ZetaLoop:         220.0,
		ZetaEye:          30.0,
		ZetaTurnIn:       3.0,
		ZetaTurnDir:      16.0,
		ZetaLoopW:        1.0033,
		ZetaLoopTaper:    0.1003,
		ZetaStart:        -55.0,
		ZetaStartRise:    30.0,
		ZetaStartDir:     -115.0,
		ZetaStartW:       0.7358,
		ZetaOver:         40.0,
		XiHead:           -24.0,
		XiSide:           0.44,
		XiBowX:           5.0,
		XiBowY:           495.0,
		XiBowDir:         265.0,
		XiBowW:           0.7539,
		XiWaist:          380.0,
		XiCross:          -80.0,
		XiCrossShift:     -10.0,
		XiLoop:           200.0,
		RhoVarSide:       -40.0,
		RhoVarLow:        24.0,
		RhoVarEnd:        57.0,
		RhoVarEndUp:      64.0,
		RhoVarEndDir:     35.0,
		PiVarOver:        50.0,
		ThetaVarLeft:     20.0,
		ThetaVarTopK:     -0.02,
		ThetaVarIn:       -5.0,
		ThetaVarLowSide:  164.0,
		ThetaVarWidth:    369.0,
		ThetaVarSide:     340.0,
		ThetaVarLoop:     220.0,
		ThetaVarLoopSide: 566.0,
		ThetaVarEnd:      40.0,
		ThetaVarEndY:     307.0,
		ThetaVarEndDir:   0.0,
		ThetaVarEndW:     10.0,
	}
}

// withStem returns the proportions for a stem d lighter than before (d is
// negative for a heavier stem), as WithStem moves the Latin letters: the
// centre lines of the stems stay where they are: a stem given by its
// left edge moves by half the change and a white between two stems takes
// the whole of it, and the widths between outer extremes, the side spaces,
// the directions and the heights stay.  Weights given in stems, thins and
// hairlines of the pen move with the pen.
func (g GreekParams) withStem(d float64) GreekParams {
	// τ and π: the edges of the stems right of the left of the bar, and
	// the bar beyond the right edge of the right leg.
	g.TauStem += d / 2
	g.PiLeft += d / 2
	g.PiCounter += d
	g.PiOver += d / 2
	// γ: the loop and the end of the rise, given from the left edge of the
	// stem under the entry, which WithStem moves by half the change.
	return g
}

// withXHeight returns the proportions for the x-height x of a letter whose
// x-height is x0 and whose ascender is asc, as WithXHeight moves the Latin
// letters: a height inside the x-height grows with it in proportion, a
// height between the x-height and the ascender keeps its share of the
// distance between them, a height at which a stroke runs into a straight
// stem is a multiple of SlantDen (den) as the stems' lattice needs, and
// the depths below the baseline, the heights of the turns at the baseline
// (the start of the turn of μ, the closing strokes of ρ and β), as those of
// the exit, and the heights given as shares of the x-height stay.
func (g GreekParams) withXHeight(x0, x, asc, den float64) GreekParams {
	f := x / x0
	in := func(v *float64) { *v *= f }
	above := func(v *float64) { *v = x + (*v-x0)*(asc-x)/(asc-x0) }
	lat := func(v *float64) { *v = math.Round(*v/den) * den }
	for _, v := range []*float64{
		&g.AlphaCross, &g.LambdaLegY, &g.ChiCross, &g.KappaVarBowY,
		&g.BetaWaist, &g.GammaCrossY, &g.ThetaVarLowSide, &g.ThetaVarSide, &g.ThetaVarEndY, &g.EpsBallLow,
	} {
		in(v)
	}
	for _, v := range []*float64{&g.RhoJoin, &g.ChiHigh, &g.PhiLoopJoin} {
		in(v)
		lat(v)
	}
	for _, v := range []*float64{&g.PsiTop, &g.DeltaNeckY, &g.SigmaBar, &g.ThetaVarLoopSide} {
		above(v)
	}
	for _, v := range []*float64{&g.LambdaHigh, &g.BetaJoin} {
		above(v)
		lat(v)
	}
	return g
}

// widen returns the proportions with the Greek letters widened by f, the
// rule of the script forms which Params.widen states for the Latin letters:
// the widths between outer extremes, the reach of a bowl or a side from the
// left of the letter and the bars grow by f, a counter between two stems so
// that the distance of their centre lines grows by f (stem0 is the stem
// before and stem the stem after the change of the pen), and a diagonal
// stroke turns such that its horizontal run grows by f.  The turns, the hooks,
// the loops, the tips and the ends, the offsets of a node from the middle of
// its round or from a stem, the side spaces and the heights keep their
// values: a turn which a heavier pen draws is not a stretched turn of a
// lighter one.
func (g GreekParams) widen(f, stem0, stem float64) GreekParams {
	if f == 1 {
		return g
	}
	for _, v := range []*float64{
		&g.ThetaWidth, &g.PhiWidth, &g.RhoWidth, &g.AlphaBowl, &g.AlphaTip, &g.AlphaFoot,
		&g.NuWidth, &g.PsiWidth, &g.DeltaWidth, &g.EpsUpper, &g.EpsLower, &g.EpsTip,
		&g.PhiLoopWidth, &g.OmegaWidth, &g.BetaUpper, &g.BetaLower,
		&g.SigmaBar, &g.TauWidth, &g.TauStem, &g.UpsilonWidth, &g.LunateWidth, &g.LunateTip,
		&g.LunateBarEnd, &g.ZetaWidth, &g.ThetaVarWidth,
		&g.ThetaVarLoop, &g.LambdaOut,
	} {
		*v = math.Round(*v * f)
	}
	g.PiCounter = f*(g.PiCounter+stem0) - stem
	turn := func(a *float64) {
		r := *a * math.Pi / 180
		*a = math.Atan2(math.Sin(r), f*math.Cos(r)) * 180 / math.Pi
	}
	for _, a := range []*float64{&g.AlphaDown, &g.LambdaDir, &g.LambdaLegDir, &g.ChiDir, &g.ChiThinDir, &g.GammaRiseDir, &g.NuRiseDir} {
		turn(a)
	}
	return g
}
