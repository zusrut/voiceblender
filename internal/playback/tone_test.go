package playback

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestLookupTone_ExactMatch(t *testing.T) {
	spec, ok := LookupTone("us_ringback")
	if !ok {
		t.Fatal("expected us_ringback to be found")
	}
	if len(spec.Frequencies) != 2 || spec.Frequencies[0] != 440 || spec.Frequencies[1] != 480 {
		t.Fatalf("unexpected frequencies: %v", spec.Frequencies)
	}
}

func TestLookupTone_BareNameDefaultsUS(t *testing.T) {
	spec, ok := LookupTone("busy")
	if !ok {
		t.Fatal("expected bare 'busy' to resolve to us_busy")
	}
	if len(spec.Frequencies) != 2 || spec.Frequencies[0] != 480 {
		t.Fatalf("unexpected frequencies for bare busy: %v", spec.Frequencies)
	}
}

func TestLookupTone_CaseInsensitive(t *testing.T) {
	_, ok := LookupTone("US_Ringback")
	if !ok {
		t.Fatal("expected case-insensitive match")
	}
}

func TestLookupTone_UKAlias(t *testing.T) {
	spec, ok := LookupTone("uk_ringback")
	if !ok {
		t.Fatal("expected uk_ringback to resolve via alias to gb_ringback")
	}
	gb, _ := LookupTone("gb_ringback")
	if spec.Frequencies[0] != gb.Frequencies[0] {
		t.Fatal("uk_ringback should match gb_ringback")
	}
}

func TestLookupTone_Unknown(t *testing.T) {
	_, ok := LookupTone("xx_unknown")
	if ok {
		t.Fatal("expected unknown tone to not be found")
	}
}

func TestToneNames(t *testing.T) {
	names := ToneNames()
	if len(names) != 41 { // 4 types × 10 countries + 1 (pl_ringback)
		t.Fatalf("expected 40 tone names, got %d", len(names))
	}
	// Check sorted order.
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Fatalf("names not sorted: %s before %s", names[i-1], names[i])
		}
	}
}

func TestToneReader_ProducesNonZeroDuringOn(t *testing.T) {
	spec := ToneSpec{
		Frequencies: []float64{440},
	}
	tr := NewToneReader(spec, 8000)
	buf := make([]byte, 320) // 160 samples = 20ms at 8kHz
	n, err := tr.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 320 {
		t.Fatalf("expected 320 bytes, got %d", n)
	}

	hasNonZero := false
	for i := 0; i < n; i += 2 {
		s := int16(binary.LittleEndian.Uint16(buf[i:]))
		if s != 0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		t.Fatal("expected non-zero samples during continuous tone")
	}
}

func TestToneReader_SilenceDuringOff(t *testing.T) {
	// Tone with 10ms on, then 100ms off.
	spec := ToneSpec{
		Frequencies: []float64{440},
		Cadence:     []CadenceSegment{{10, true}, {100, false}},
	}
	tr := NewToneReader(spec, 8000)

	// Read past the 10ms on segment (80 samples = 160 bytes).
	onBuf := make([]byte, 160)
	tr.Read(onBuf)

	// Skip past the 2ms fade-out ramp (16 samples = 32 bytes).
	fadeBuf := make([]byte, 32)
	tr.Read(fadeBuf)

	// Now read well into the off segment — should be silence.
	offBuf := make([]byte, 320) // 160 samples
	n, _ := tr.Read(offBuf)

	allZero := true
	for i := 0; i < n; i += 2 {
		s := int16(binary.LittleEndian.Uint16(offBuf[i:]))
		if s != 0 {
			allZero = false
			break
		}
	}
	if !allZero {
		t.Fatal("expected silence during off segment")
	}
}

func TestToneReader_DualToneDiffersFromSingle(t *testing.T) {
	single := NewToneReader(ToneSpec{Frequencies: []float64{440}}, 8000)
	dual := NewToneReader(ToneSpec{Frequencies: []float64{440, 480}}, 8000)

	buf1 := make([]byte, 320)
	buf2 := make([]byte, 320)
	single.Read(buf1)
	dual.Read(buf2)

	same := true
	for i := 0; i < 320; i++ {
		if buf1[i] != buf2[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("single and dual tone should produce different output")
	}
}

func TestToneReader_NeverReturnsEOF(t *testing.T) {
	tr := NewToneReader(ToneSpec{Frequencies: []float64{425}, Cadence: onOff(500, 500)}, 8000)
	buf := make([]byte, 1600) // 100ms
	for i := 0; i < 100; i++ {
		_, err := tr.Read(buf)
		if err != nil {
			t.Fatalf("unexpected error on read %d: %v", i, err)
		}
	}
}

func TestToneReader_PhaseContinuityAcrossReads(t *testing.T) {
	tr := NewToneReader(ToneSpec{Frequencies: []float64{440}}, 8000)

	// Read one big chunk vs two small chunks — last sample of first should
	// lead smoothly into first sample of second.
	buf1 := make([]byte, 320)
	buf2 := make([]byte, 320)
	tr.Read(buf1)
	tr.Read(buf2)

	lastSample := int16(binary.LittleEndian.Uint16(buf1[318:]))
	firstSample := int16(binary.LittleEndian.Uint16(buf2[0:]))

	// They should be close in value (smooth transition).
	diff := math.Abs(float64(lastSample) - float64(firstSample))
	// At 440Hz/8kHz, max delta between adjacent samples is about 2*pi*440/8000 * 23000 ≈ 7950
	// Allow generous threshold.
	if diff > 10000 {
		t.Fatalf("phase discontinuity: last=%d, first=%d, diff=%.0f", lastSample, firstSample, diff)
	}
}

func TestRegistryEntries_Valid(t *testing.T) {
	for name, spec := range toneRegistry {
		if len(spec.Frequencies) == 0 {
			t.Errorf("%s: no frequencies", name)
		}
		for _, f := range spec.Frequencies {
			if f <= 0 || f > 4000 {
				t.Errorf("%s: invalid frequency %.1f", name, f)
			}
		}
		if spec.ModulationHz < 0 {
			t.Errorf("%s: negative modulation", name)
		}
		for i, seg := range spec.Cadence {
			if seg.DurationMs <= 0 {
				t.Errorf("%s: cadence segment %d has non-positive duration %d", name, i, seg.DurationMs)
			}
		}
	}
}

func TestToneReader_AMModulation(t *testing.T) {
	// Verify that AM modulation produces varying amplitude.
	spec := ToneSpec{
		Frequencies:  []float64{400},
		ModulationHz: 17,
	}
	tr := NewToneReader(spec, 8000)

	// Read ~60ms (enough for one modulation cycle at 17Hz ≈ 59ms).
	buf := make([]byte, 960) // 480 samples
	tr.Read(buf)

	var minAbs, maxAbs int16
	for i := 0; i < len(buf); i += 2 {
		s := int16(binary.LittleEndian.Uint16(buf[i:]))
		abs := s
		if abs < 0 {
			abs = -abs
		}
		if abs > maxAbs {
			maxAbs = abs
		}
		if i == 0 || abs < minAbs {
			minAbs = abs
		}
	}
	// With AM, min amplitude should be noticeably less than max.
	if float64(minAbs) > float64(maxAbs)*0.8 {
		t.Fatalf("AM modulation not producing amplitude variation: min=%d max=%d", minAbs, maxAbs)
	}
}
