package day5_test

import (
	"bytes"
	"testing"

	"adventofcode2025/day1/src/day5"
)

func TestComputeExample(t *testing.T) {
	input := `3-5
10-14
16-20
12-18
`

	res, err := day5.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.TotalFreshIDs != 14 {
		t.Fatalf("TotalFreshIDs=%d, want 14", res.TotalFreshIDs)
	}
}

func TestComputeHandlesOverlapsAndAdjacency(t *testing.T) {
	input := `3-5
5-7
9-10
11-11
`

	res, err := day5.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	// Union: 3-7 and 9-11 => 5 + 3 = 8.
	if res.TotalFreshIDs != 8 {
		t.Fatalf("TotalFreshIDs=%d, want 8", res.TotalFreshIDs)
	}
}

func TestComputeIgnoresContentAfterBlankLine(t *testing.T) {
	input := "\n\n3-3\n\n\n3\n\n"
	res, err := day5.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.TotalFreshIDs != 1 {
		t.Fatalf("TotalFreshIDs=%d, want 1", res.TotalFreshIDs)
	}
}

func TestComputeErrors(t *testing.T) {
	t.Run("NoRanges", func(t *testing.T) {
		if _, err := day5.Compute(bytes.NewBufferString("\n\n1\n")); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("InvalidRangeLine", func(t *testing.T) {
		if _, err := day5.Compute(bytes.NewBufferString("1:2\n\n1\n")); err == nil {
			t.Fatalf("expected error")
		}
	})
}
