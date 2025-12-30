package day3_test

import (
	"bytes"
	"testing"

	"adventofcode2025/day1/src/day3"
)

func TestComputeExample(t *testing.T) {
	input := `987654321111111
811111111111119
234234234234278
818181911112111`

	result, err := day3.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if result.Banks != 4 {
		t.Fatalf("Banks = %d, want 4", result.Banks)
	}
	if result.Total != 3121910778619 {
		t.Fatalf("Total = %d, want 3121910778619", result.Total)
	}
}

func TestComputeIgnoresBlankLines(t *testing.T) {
	input := "\n  12345678901234  \n\n21098765432109\n"
	result, err := day3.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if result.Banks != 2 {
		t.Fatalf("Banks = %d, want 2", result.Banks)
	}
	// Best 12-digit subsequences: 345678901234 and 298765432109.
	if result.Total != 644444333343 {
		t.Fatalf("Total = %d, want 644444333343", result.Total)
	}
}

func TestComputeErrors(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"NoBanks", "   \n\t"},
		{"TooShort", "12345678901\n"},
		{"InvalidDigit", "1a23456789012\n"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := day3.Compute(bytes.NewBufferString(tt.input)); err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
		})
	}
}
