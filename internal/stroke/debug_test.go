package stroke

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
)

func TestDebugL(t *testing.T) {
	p := design.Regular()
	s := italicLetters(p)["l"].Strokes[0].ShearedRatio(SlantNum, SlantDen)
	r, err := s.Build()
	if err != nil {
		t.Fatal(err)
	}
	for _, ch := range []Chain{r.Right, r.Left} {
		for _, pc := range ch {
			if pc.C[0].Y < 200 || pc.C[3].Y < 200 {
				t.Logf("%.2f fixed %v", pc.C, pc.Fixed)
			}
		}
		t.Log("--")
	}
	for _, c := range r.Path().Round() {
		for _, sg := range c.Segs {
			if sg.P3.Y < 200 {
				t.Logf("%v", sg)
			}
		}
	}
}
