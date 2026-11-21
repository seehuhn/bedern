// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"seehuhn.de/go/postscript/type1"
	"seehuhn.de/go/sfnt"
	"seehuhn.de/go/sfnt/glyph"
	"seehuhn.de/go/sfnt/header"
	"seehuhn.de/go/sfnt/name"
	"seehuhn.de/go/sfnt/os2"

	bedern "seehuhn.de/go/bedern/internal/font"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
)

func TestPrivateDict(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	d := privateDict(p)

	bv := d.BlueValues
	if len(bv) == 0 || len(bv)%2 != 0 {
		t.Fatalf("BlueValues %v: need an even, non-zero number of values", bv)
	}
	scale := d.BlueScale
	if scale == 0 {
		scale = type1.DefaultBlueScale
	}
	if bv[0] != -p.OvershootLower || bv[1] != 0 {
		t.Errorf("first zone %v, %v, want the baseline zone (%v, 0)", bv[0], bv[1], -p.OvershootLower)
	}
	for i := 0; i < len(bv); i += 2 {
		lo, hi := bv[i], bv[i+1]
		if lo > hi {
			t.Errorf("zone %d (%v, %v) is not sorted", i/2, lo, hi)
		}
		if i > 0 && lo <= bv[i-1] {
			t.Errorf("zone %d (%v, %v) overlaps the previous zone", i/2, lo, hi)
		}
		if (hi-lo)*scale >= 1 {
			t.Errorf("zone %d (%v, %v) is too wide for BlueScale %v", i/2, lo, hi, scale)
		}
	}

	// OtherBlues has one zone: the descender line and the overshoot of a round
	// descender below it.  The zone lies below the baseline zone and is not too
	// wide for BlueScale.
	ob := d.OtherBlues
	if len(ob) != 2 {
		t.Fatalf("OtherBlues %v: want the one zone of the descender line", ob)
	}
	if ob[1] != -p.Descender || ob[0] != -p.Descender-p.OvershootLower {
		t.Errorf("descender zone %v, %v, want (%v, %v)", ob[0], ob[1], -p.Descender-p.OvershootLower, -p.Descender)
	}
	if ob[1] >= bv[0] {
		t.Errorf("descender zone (%v, %v) does not lie below the baseline zone (%v, %v)", ob[0], ob[1], bv[0], bv[1])
	}
	if (ob[1]-ob[0])*scale >= 1 {
		t.Errorf("descender zone (%v, %v) is too wide for BlueScale %v", ob[0], ob[1], scale)
	}

	// BlueShift and BlueFuzz have the values set in privateDict.
	if d.BlueShift != 7 {
		t.Errorf("BlueShift %v, want 7", d.BlueShift)
	}
	if d.BlueFuzz != 1 {
		t.Errorf("BlueFuzz %v, want 1", d.BlueFuzz)
	}
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	gs := glyphstest.Set(p)

	font, err := regularFont()
	if err != nil {
		t.Fatal(err)
	}

	got := roundTrip(t, font)

	if got.FamilyName != "Bedern" {
		t.Errorf("family name %q", got.FamilyName)
	}
	if got.UnitsPerEm != 1000 {
		t.Errorf("units per em %d", got.UnitsPerEm)
	}
	if !got.IsCFF() {
		t.Error("font does not have CFF outlines")
	}
	if n := got.NumGlyphs(); n != len(gs) {
		t.Fatalf("%d glyphs, want %d", n, len(gs))
	}
	for i, g := range gs {
		gid := glyph.ID(i)
		if name := got.GlyphName(gid); name != g.Name {
			t.Errorf("glyph %d: name %q, want %q", i, name, g.Name)
		}
		if w := got.GlyphWidth(gid); w != g.Advance {
			t.Errorf("glyph %s: width %v, want %v", g.Name, w, g.Advance)
		}
	}

	// Metrics and other fields.
	if float64(got.Ascent) != p.TypoAscender {
		t.Errorf("ascent %d, want %v", got.Ascent, p.TypoAscender)
	}
	if float64(got.Descent) != p.TypoDescender {
		t.Errorf("descent %d, want %v", got.Descent, p.TypoDescender)
	}
	if float64(got.LineGap) != p.LineGap {
		t.Errorf("line gap %d, want %v", got.LineGap, p.LineGap)
	}
	if float64(got.XHeight) != p.XHeight {
		t.Errorf("x-height %d, want %v", got.XHeight, p.XHeight)
	}
	if float64(got.CapHeight) != p.CapHeight {
		t.Errorf("cap height %d, want %v", got.CapHeight, p.CapHeight)
	}
	if got.Copyright != font.Copyright {
		t.Errorf("copyright %q, want %q", got.Copyright, font.Copyright)
	}

	// The font bounding box and the OS/2 metrics.
	var x0, y0, x1, y1 float64
	first := true
	for _, g := range gs {
		if len(g.Path) == 0 {
			continue
		}
		gx0, gy0, gx1, gy1 := g.Path.BBox()
		if first {
			x0, y0, x1, y1 = gx0, gy0, gx1, gy1
			first = false
			continue
		}
		x0, y0 = min(x0, gx0), min(y0, gy0)
		x1, y1 = max(x1, gx1), max(y1, gy1)
	}
	bbox := got.FontBBox()
	if bbox.IsZero() {
		t.Error("the font bounding box is empty")
	}
	if bbox.LLx > x0 || bbox.LLy > y0 || bbox.URx < x1 || bbox.URy < y1 {
		t.Errorf("font bounding box %v does not contain the glyph extents (%v, %v, %v, %v)",
			bbox, x0, y0, x1, y1)
	}
	if float64(got.WinAscent) < y1 {
		t.Errorf("WinAscent %d does not cover the top of the bounding box (%v)", got.WinAscent, y1)
	}
	if float64(got.WinDescent) < -y0 {
		t.Errorf("WinDescent %d does not cover the bottom of the bounding box (%v)", got.WinDescent, y0)
	}
	if got.UnicodeRange[0]&(1<<os2.URBasicLatin) == 0 {
		t.Error("the Basic Latin bit of the Unicode range is not set")
	}
	if got.IsFixedPitch() {
		t.Error("font claims to be fixed pitch")
	}
	if got.FontName != "Bedern-Regular" {
		t.Errorf("PostScript name %q", got.FontName)
	}

	// The character map.
	cm, _ := got.CMapTable.GetBest()
	if cm == nil {
		t.Fatal("no character map")
	}
	for _, c := range []struct {
		r    rune
		name string
	}{{'o', "o"}, {'l', "l"}, {'h', "h"}, {'d', "d"}, {'a', "a"}, {'s', "s"}, {'i', "i"}, {'b', "b"}, {'p', "p"}, {'q', "q"}, {' ', "space"}} {
		gid := cm.Lookup(c.r)
		if name := got.GlyphName(gid); name != c.name {
			t.Errorf("cmap %q: glyph %q, want %q", c.r, name, c.name)
		}
	}
}

// TestStyleNames checks that a style other than the regular one reaches the
// name table.
func TestStyleNames(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	style := design.Style{
		Family:     "Bedern Test",
		Subfamily:  "Fit190",
		PostScript: "BedernTest-Fit190",
		Weight:     400,
	}
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	font, err := FontWith(p, style, glyphstest.Set(p), now, Options{NoKern: true})
	if err != nil {
		t.Fatal(err)
	}
	got := roundTrip(t, font)
	if got.FamilyName != "Bedern Test" || got.Subfamily != "Fit190" || got.FontName != "BedernTest-Fit190" {
		t.Errorf("names: %q, %q, %q", got.FamilyName, got.Subfamily, got.FontName)
	}
}

// TestInvalidStyle checks that Font rejects an invalid style.
func TestInvalidStyle(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	style := design.RegularStyle()
	style.PostScript = "Bedern Regular"
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	if _, err := Font(p, style, glyphstest.Set(p), now); err == nil {
		t.Error("no error for an invalid style")
	}
}

// roundTrip writes the font to a temporary file and reads it back.
func roundTrip(t *testing.T, font *sfnt.Font) *sfnt.Font {
	t.Helper()
	got, err := sfnt.ReadFile(writeFont(t, font))
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// writeFont writes the font to a temporary file and returns the file name.
func writeFont(t *testing.T, font *sfnt.Font) string {
	t.Helper()
	fname := filepath.Join(t.TempDir(), "Bedern-Regular.otf")
	out, err := os.Create(fname)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := font.Write(out); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	return fname
}

// testNow is the time of the fonts of the tests.
var testNow = time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

// regularFont builds the regular font with kerning once per test binary.  The
// font is shared and must not be changed.
var regularFont = sync.OnceValues(func() (*sfnt.Font, error) {
	p := design.Regular()
	return Font(p, design.RegularStyle(), glyphstest.Set(p), testNow)
})

func buildTestFont(t *testing.T) string {
	t.Helper()
	font, err := regularFont()
	if err != nil {
		t.Fatal(err)
	}
	return writeFont(t, font)
}

// glyphAdvance is one entry of the hb-shape output.
type glyphAdvance struct {
	name    string
	advance string
}

// parseHBShape parses hb-shape output like "[l=0+270|o=1+486]" into glyph
// names and advances.  The cluster numbers are ignored.
func parseHBShape(out string) ([]glyphAdvance, error) {
	out = strings.TrimSpace(out)
	if !strings.HasPrefix(out, "[") || !strings.HasSuffix(out, "]") {
		return nil, fmt.Errorf("unexpected output %q", out)
	}
	var res []glyphAdvance
	for _, item := range strings.Split(out[1:len(out)-1], "|") {
		name, rest, ok := strings.Cut(item, "=")
		if !ok {
			return nil, fmt.Errorf("bad glyph entry %q", item)
		}
		_, adv, ok := strings.Cut(rest, "+")
		if !ok {
			return nil, fmt.Errorf("bad glyph entry %q", item)
		}
		res = append(res, glyphAdvance{name, adv})
	}
	return res, nil
}

// needTool returns the path of the named external tool.  If the tool is
// missing, the test is skipped, unless the environment variable
// BEDERN_REQUIRE_TOOLS is 1, in which case the test fails.
func needTool(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		if os.Getenv("BEDERN_REQUIRE_TOOLS") == "1" {
			t.Fatalf("required tool %s not found (BEDERN_REQUIRE_TOOLS=1): %v", name, err)
		}
		t.Skipf("%s not found", name)
	}
	return path
}

// TestHarfBuzz checks the advance widths as HarfBuzz sees them.
func TestHarfBuzz(t *testing.T) {
	t.Parallel()
	path := needTool(t, "hb-shape")
	fname := buildTestFont(t)
	cmd := exec.Command(path, fname, "lol o")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("hb-shape: %v\nstderr: %s", err, stderr.String())
	}
	got, err := parseHBShape(stdout.String())
	if err != nil {
		t.Fatalf("%v\nstderr: %s", err, stderr.String())
	}
	want := []glyphAdvance{{"l", "270"}, {"o", "486"}, {"l", "270"}, {"space", "323"}, {"o", "486"}}
	if !slices.Equal(got, want) {
		t.Errorf("hb-shape: got %v, want %v\nstderr: %s", got, want, stderr.String())
	}
}

// ttxValues dumps the given tables with ttx and returns the value attribute
// of the first element of each name.  Attribute order does not matter.
func ttxValues(t *testing.T, fname string, tables ...string) map[string]string {
	t.Helper()
	dump := filepath.Join(t.TempDir(), "dump.ttx")
	args := []string{"-q"}
	for _, tab := range tables {
		args = append(args, "-t", tab)
	}
	args = append(args, "-o", dump, fname)
	if out, err := exec.Command("ttx", args...).CombinedOutput(); err != nil {
		t.Fatalf("ttx: %v\n%s", err, out)
	}
	data, err := os.ReadFile(dump)
	if err != nil {
		t.Fatal(err)
	}
	res := make(map[string]string)
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("parsing ttx output: %v", err)
		}
		el, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		for _, a := range el.Attr {
			if a.Name.Local == "value" {
				if _, seen := res[el.Name.Local]; !seen {
					res[el.Name.Local] = a.Value
				}
			}
		}
	}
	return res
}

// TestTTX checks the OS/2, hhea, head and hmtx tables as fontTools sees them.
func TestTTX(t *testing.T) {
	t.Parallel()
	needTool(t, "ttx")
	fname := buildTestFont(t)
	v := ttxValues(t, fname, "OS/2", "hhea", "head")

	sel, err := strconv.ParseUint(strings.ReplaceAll(v["fsSelection"], " ", ""), 2, 16)
	if err != nil {
		t.Fatalf("fsSelection %q: %v", v["fsSelection"], err)
	}
	for bit, want := range map[uint]bool{0: false, 5: false, 6: true, 7: true} {
		if got := sel>>bit&1 == 1; got != want {
			t.Errorf("fsSelection bit %d is %t, want %t", bit, got, want)
		}
	}

	for _, c := range []struct{ name, want string }{
		{"sTypoAscender", "806"},
		{"sTypoDescender", "-194"},
		{"sTypoLineGap", "200"},
		{"ascent", "806"},
		{"descent", "-194"},
		{"lineGap", "200"},
		{"unitsPerEm", "1000"},
	} {
		if v[c.name] != c.want {
			t.Errorf("%s is %q, want %s", c.name, v[c.name], c.want)
		}
	}

}

// TestTTXHmtx checks the hmtx table as fontTools sees it.
func TestTTXHmtx(t *testing.T) {
	t.Parallel()
	needTool(t, "ttx")
	fname := buildTestFont(t)
	dump := filepath.Join(t.TempDir(), "hmtx.ttx")
	if out, err := exec.Command("ttx", "-q", "-t", "hmtx", "-o", dump, fname).CombinedOutput(); err != nil {
		t.Fatalf("ttx: %v\n%s", err, out)
	}
	data, err := os.ReadFile(dump)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`<mtx name="o" width="(\d+)" lsb="(-?\d+)"`)
	m := re.FindStringSubmatch(string(data))
	if m == nil {
		t.Fatalf("no hmtx entry for o in\n%s", data)
	}
	if m[1] != "486" {
		t.Errorf("advance of o is %s, want 486", m[1])
	}
}

// TestLicenceNames checks that every language table of the name table in the
// font file holds the copyright notice, the licence description and the
// licence URL.
func TestLicenceNames(t *testing.T) {
	t.Parallel()
	fname := buildTestFont(t)
	r, err := os.Open(fname)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	h, err := header.Read(r)
	if err != nil {
		t.Fatal(err)
	}
	data, err := h.ReadTableBytes(r, "name")
	if err != nil {
		t.Fatal(err)
	}
	info, err := name.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for platform, tables := range map[string]name.Tables{"Mac": info.Mac, "Windows": info.Windows} {
		for lang, tab := range tables {
			n++
			if tab.Copyright != bedern.Copyright {
				t.Errorf("%s %s: copyright %q, want %q", platform, lang, tab.Copyright, bedern.Copyright)
			}
			if tab.License != bedern.License {
				t.Errorf("%s %s: licence %q, want %q", platform, lang, tab.License, bedern.License)
			}
			if tab.LicenseURL != bedern.LicenseURL {
				t.Errorf("%s %s: licence URL %q, want %q", platform, lang, tab.LicenseURL, bedern.LicenseURL)
			}
		}
	}
	if n == 0 {
		t.Error("the name table has no entries")
	}
}

// Font refuses parameters which do not pass Validate.
func TestFontRejectsInvalidParams(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	p.ThinLower = p.CurveLower
	_, err := Font(p, design.RegularStyle(), glyphstest.Set(design.Regular()), time.Unix(0, 0))
	if err == nil {
		t.Error("Font accepted invalid parameters")
	}
}
