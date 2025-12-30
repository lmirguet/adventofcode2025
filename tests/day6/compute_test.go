package day6_test

import (
	"bytes"
	"testing"

	"adventofcode2025/day1/src/day6"
)

func TestComputeExample(t *testing.T) {
	input := "123 328  51 64 \n 45 64  387 23 \n  6 98  215 314\n*   +   *   +  \n"
	res, err := day6.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.Total != 3263827 {
		t.Fatalf("Total=%d, want 3263827", res.Total)
	}
}

func TestComputeParsesColumnNumbers(t *testing.T) {
	// Two problems separated by an all-space column.
	// Problem 1 (cols 0-2): 4 + 23 + 1 = 28
	// Problem 2 (cols 4-5): 56 * 2 = 112
	input := " 2  5\n431 62\n+   *\n"
	res, err := day6.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.Total != 140 {
		t.Fatalf("Total=%d, want 140", res.Total)
	}
}

func TestComputeErrors(t *testing.T) {
	t.Run("NoWorksheet", func(t *testing.T) {
		if _, err := day6.Compute(bytes.NewBufferString("   \n\t\n")); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("MissingOperator", func(t *testing.T) {
		if _, err := day6.Compute(bytes.NewBufferString("1 2\n3 4\n   \n")); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("InvalidOperator", func(t *testing.T) {
		if _, err := day6.Compute(bytes.NewBufferString("1\nx\n")); err == nil {
			t.Fatalf("expected error")
		}
	})
}
