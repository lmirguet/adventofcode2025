package day1_test

import (
	"bytes"
	"fmt"
	"testing"

	"adventofcode2025/day1/src/day1"
)

func TestParseTokenValid(t *testing.T) {
	cases := []struct {
		in  string
		dir rune
		mag int
	}{
		{"L1", 'L', 1},
		{"R68", 'R', 68},
	}
	for _, tt := range cases {
		dir, mag, err := day1.ParseToken(tt.in)
		if err != nil {
			t.Fatalf("ParseToken(%s) unexpected error: %v", tt.in, err)
		}
		if dir != tt.dir || mag != tt.mag {
			t.Fatalf("ParseToken(%s) got (%c,%d), want (%c,%d)", tt.in, dir, mag, tt.dir, tt.mag)
		}
	}
}

func TestComputeIgnoresSpacing(t *testing.T) {
	input := "L1   R2\n\nL3\tR4"
	result, err := day1.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute unexpected error: %v", err)
	}
	if result.Position != 52 {
		t.Fatalf("Compute position got %d, want 52", result.Position)
	}
	if result.ZeroCrossings != 0 {
		t.Fatalf("Compute zero crossings got %d, want 0", result.ZeroCrossings)
	}
}

func TestComputeSampleSequence(t *testing.T) {
	input := "L68 L30 R48 L5 R60 L55 L1 L99 R14 L82"
	result, err := day1.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute unexpected error: %v", err)
	}
	if result.Position != 32 {
		t.Fatalf("Compute position got %d, want 32", result.Position)
	}
	if result.ZeroCrossings != 6 {
		t.Fatalf("Compute zero crossings got %d, want 6", result.ZeroCrossings)
	}
}

func TestInvalidTokens(t *testing.T) {
	cases := []string{
		"X10",
		"l5",
		"Lxx",
		"L-1",
		"L",
		"20",
	}
	for _, input := range cases {
		_, err := day1.Compute(bytes.NewBufferString(input))
		if err == nil {
			t.Fatalf("expected error for input %q", input)
		}
	}
}

func TestNoInstructions(t *testing.T) {
	_, err := day1.Compute(bytes.NewBufferString("   \n\t  "))
	if err == nil {
		t.Fatalf("expected error for empty input")
	}
}

func TestWrapBehavior(t *testing.T) {
	cases := []struct {
		input     string
		expectPos int
	}{
		{"L1", 49},  // 50-1
		{"L51", 99}, // 50-51 -> wrap to 99
		{"R50", 0},  // 50+50 -> wrap to 0
		{"L68 L30 R48 L5 R60 L55 L1 L99 R14 L82", 32}, // sample with wrap
	}
	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			got, err := day1.Compute(bytes.NewBufferString(tt.input))
			if err != nil {
				t.Fatalf("Compute(%q) error: %v", tt.input, err)
			}
			if got.Position != tt.expectPos {
				t.Fatalf("Compute(%q) position = %d, want %d", tt.input, got.Position, tt.expectPos)
			}
		})
	}
}

func TestZeroCrossingCounting(t *testing.T) {
	cases := []struct {
		input         string
		zeroCrossings int
		finalPos      int
	}{
		{"R50", 1, 0},
		{"R50 R50", 1, 50}, // only first hit 0
		{"R50 L50", 1, 50}, // one hit at 0
		{"L50", 1, 0},      // post-rotation hit counted
		{"L75", 1, 75},     // crosses zero once mid-motion
		{"L150", 2, 0},     // wraps and crosses zero twice
		{"R250", 3, 0},     // multiple wraps to the right
		{"L68 L30 R48 L5 R60 L55 L1 L99 R14 L82", 6, 32}, // sample
	}

	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			res, err := day1.Compute(bytes.NewBufferString(tt.input))
			if err != nil {
				t.Fatalf("Compute(%q) error: %v", tt.input, err)
			}
			if res.ZeroCrossings != tt.zeroCrossings {
				t.Fatalf("Compute(%q) zeroCrossings=%d, want %d", tt.input, res.ZeroCrossings, tt.zeroCrossings)
			}
			if res.Position != tt.finalPos {
				t.Fatalf("Compute(%q) position=%d, want %d", tt.input, res.Position, tt.finalPos)
			}
		})
	}
}

func BenchmarkComputeLargeSequence(b *testing.B) {
	// Build a 10k-token alternating sequence.
	var buf bytes.Buffer
	for i := 0; i < 10000; i++ {
		if i%2 == 0 {
			fmt.Fprint(&buf, "L1 ")
		} else {
			fmt.Fprint(&buf, "R2 ")
		}
	}
	data := buf.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := day1.Compute(bytes.NewReader(data)); err != nil {
			b.Fatalf("Compute error: %v", err)
		}
	}
}
