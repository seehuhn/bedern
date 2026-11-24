// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

// Diag are the parameters of b, p, s, k, v, w, x and z, in the upright
// description.  Weights are given as shares of the pen and the stem, so
// that WithStem moves them; heights which belong to the x-height are given
// from the x-height or as shares of it, which WithXHeight moves.
type Diag struct {
	BWidth    float64 // the set width of b
	BLeftLess float64 // how much less the left side of b is than that of n
	BBottomAt float64 // the lowest node of the bowl of b, as a share of the way from the stem centre to the free side
	BTopAt    float64 // the top node of the bowl of b, likewise
	BSideY    float64 // the height of the node at the free side of the bowls of b and p, as a share of the x-height
	BBackDir  float64 // the direction, upright, opposite to that in which the bowl of b runs back into the stem
	BTurnVert float64 // the height of the node where the turn of b from the stem into the bowl runs vertically in the italic, as a share of the x-height
	BTurnW    float64 // the thickness of the turn there, as a share of the major axis of the pen

	PWidth    float64 // the set width of p
	PStartDir float64 // the direction, upright, in which the bowl of p leaves the stem
	PTopAt    float64 // the top node of the bowl of p, as a share of the way from the stem centre to the free side
	PBottomAt float64 // the lowest node of the bowl of p, likewise
	PJoin     float64 // the height at which the counter of p meets the stem at the bottom
	PJoinDir  float64 // the direction, upright, in which the bowl of p runs into the stem at the bottom
	PJoinBack float64 // how far beyond the join the centre line of the bowl of p runs on into the stem
	PBranch   float64 // the height at which the underside of the bowl of p meets the stem, as Branch does on n

	SWidth      float64 // the set width of s
	SLower      float64 // the ink width of the lower half of s
	SUpper      float64 // the ink width of the upper half of s
	SUpperShift float64 // how far the upper half of s stands right of the middle
	STopShift   float64 // how far the top node of s lies right of the middle of its half, and the bottom node left
	SSideAt     float64 // the height of the side nodes of s, as a share of the way from the spine to the top or bottom
	SSpineAngle float64 // the direction of the spine of s below the horizontal, upright, in degrees
	SEndW       float64 // the thickness of the ends of s, as a share of the minor axis of the pen
	SEndDrop    float64 // how far the centres of the ends of s lie below the top and above the bottom
	SEndDir     float64 // the direction, upright, in which the stroke of s leaves its upper end and arrives at its lower

	Descender  float64 // the descender line of p and q
	FootJut    float64 // the foot serif: its length beyond the right edge of the stem
	FootObtuse float64 // the length of the serif on the obtuse side, the left, over the other
	FootThick  float64 // the thickness of the foot serif, as a share of the hairline

	VWidth  float64 // the set width of v
	VTurn   float64 // the height down to which the stem of v runs straight, a multiple of SlantDen
	VVert   float64 // the height at which the turn of v runs vertically in the italic
	VBottom float64 // the lowest node of v right of the centre of the stem, as a share of the way to the free side
	VSideY  float64 // the height of the node at the free side of v, as a share of the x-height
	VSideX  float64 // the outer edge there, left of the set width
	VSideW  float64 // the thickness there, as a share of what the pen gives

	WCounter float64 // the counter of w between the stem edges of its first half, horizontally at ShearHeight
	WFree    float64 // the outer edge of the free side of w right of the centre of its second stem
	WBottom  float64 // the lowest node of the first turn of w, as a share of the way between the stem centres
	WJoinDir float64 // the direction, upright, in which the first turn of w runs into the foot of the second stem

	XWidth   float64 // the set width of x
	XMeetW   float64 // the thickness of each arc of x where the two meet, at ShearHeight
	XOverlap float64 // how far the two arcs of x overlap there, horizontally
	XTopIn   float64 // the top node of the left arc of x left of its node at ShearHeight
	XBotIn   float64 // the lowest node of the left arc of x left of its node at ShearHeight
	XEndY    float64 // the height of the centre of the lower end of the left arc of x
	XEndDir  float64 // the direction of that end, upright, in degrees
	XThickAt float64 // the place of the node of the left arc of x where it has the stem weight, an angle along the quarter of an ellipse from the top, in degrees
	XMeetDir float64 // the direction of both arcs of x at their meeting, taken from the vertical of the italic, in degrees; the meeting then leans with the slant
	XExitRun float64 // the end of the exit of x right of the lowest node of the right arc
	XExitTip float64 // the height of the top of the end of the exit of x

	ZWidth    float64 // the set width of z
	ZTopLeft  float64 // the left end of the top bar of z, in the italic
	ZTopRight float64 // the corner at the right of the top bar of z, on the x-height, in the italic
	ZBotLeft  float64 // the corner at the left of the bottom bar of z, on the baseline, in the italic
	ZExitRun  float64 // the end of the exit of z right of the lowest node of its bottom stroke, upright

	KLegCounter float64 // the white between the stem of k and the stem of its leg, horizontally at ShearHeight
	KLegY       float64 // the height of the start of the leg of k inside the stem
	KLegStartW  float64 // the thickness of the leg of k where it leaves the stem
	KLegVert    float64 // the height of the node where the leg of k runs vertically in the italic
	VCrestIn    float64 // the crest of the free side of v and w, where it turns into the ball, left of the node at the free side
	VBallTop    float64 // the top of the ball of v and w below the x-height
	KArmY       float64 // the arm of k: the height of its start on the centre line of the stem
	KArmX       float64 // its end right of the right edge of the stem
	KArmDrop    float64 // its end below the x-height
	KArmBow     float64 // how far its ends turn from its chord, in degrees, bowing down and to the right
	KArmEndW    float64 // the thickness of its end, as a share of the minor axis of the pen
	ZBarW       float64 // the thickness of the bars of z, as a share of the hairline
	ZDiagW      float64 // the thickness of the diagonal of z at right angles, as a share of what the pen gives
	ZStartY     float64 // the height of the centre of the start of the top stroke of z
	ZStartDir   float64 // the direction, upright, in which the top stroke of z leaves its start
	ZCrestIn    float64 // the crest of the top stroke of z right of its start, in the italic
	ZLowX       float64 // the lowest node of the bottom stroke of z, in the italic
	ZCrestOver  float64 // how far the outer edge of the top stroke of z rises above the x-height at its crest
	ZTopDir     float64 // the direction of the top stroke of z where it meets the diagonal, in the italic, in degrees
	ZBotDir     float64 // the direction of the bottom stroke of z where it leaves the diagonal, in the italic, in degrees
	ZDip        float64 // how far the outer edge of the bottom stroke of z reaches below the baseline
	ZBotRun     float64 // how far the bottom stroke of z runs straight from its start at the corner
	ZLowW       float64 // the thickness of the bottom stroke of z at its lowest node, as a share of the hairline
	ZCornerCut  float64 // the angle between the cut of each stroke of z at a corner and the outer edge of the other stroke, in the italic, in degrees
}

// defaultDiag returns the starting values of the parameters.
func defaultDiag() Diag {
	return Diag{
		BWidth:    490,
		BLeftLess: 0,
		BBottomAt: 0.54,
		BTopAt:    0.46,
		BSideY:    0.5,
		BBackDir:  64,
		BTurnVert: 0.18,
		BTurnW:    0.81,

		PWidth:    480,
		PStartDir: 62,
		PTopAt:    0.42,
		PBottomAt: 0.52,
		PJoin:     68,
		PJoinDir:  120,
		PJoinBack: 20,
		PBranch:   256,

		SWidth:      360,
		SLower:      300,
		SUpper:      290,
		SUpperShift: 0,
		STopShift:   -5,
		SSideAt:     0.5,
		SSpineAngle: 16,
		SEndW:       1,
		SEndDrop:    90,
		SEndDir:     125,

		Descender:  194,
		FootJut:    63,
		FootObtuse: 0.92,
		FootThick:  0.93,

		VWidth:  440,
		VTurn:   176,
		VVert:   80,
		VBottom: 0.42,
		VSideY:  0.5,
		VSideX:  25,
		VSideW:  0.5,

		WCounter: 185,
		WFree:    270,
		WBottom:  0.45,
		WJoinDir: 45,

		XWidth:   470,
		XMeetW:   44,
		XOverlap: 26,
		XTopIn:   140,
		XBotIn:   104,
		XEndY:    28,
		XEndDir:  136,
		XThickAt: 45,
		XMeetDir: -15,
		XExitRun: 55,
		XExitTip: 80,

		ZWidth:    395,
		ZTopLeft:  100,
		ZTopRight: 390,
		ZBotLeft:  32,
		ZExitRun:  150,

		KLegCounter: 130,
		KLegY:       240,
		KLegStartW:  40,
		KLegVert:    120,
		VCrestIn:    100,
		VBallTop:    41,
		KArmY:       230,
		KArmX:       250,
		KArmDrop:    10,
		KArmBow:     6,
		KArmEndW:    1,
		ZBarW:       1.5,
		ZDiagW:      0.62,
		ZStartY:     350,
		ZStartDir:   80,
		ZCrestIn:    100,
		ZLowX:       190,
		ZCrestOver:  8,
		ZTopDir:     -6,
		ZBotDir:     -20,
		ZDip:        20,
		ZBotRun:     20,
		ZLowW:       1.4,
		ZCornerCut:  10,
	}
}
