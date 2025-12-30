package day4_test

import (
	"bytes"
	"testing"

	"adventofcode2025/day1/src/day4"
)

func TestComputeExample(t *testing.T) {
	input := `..@@.@@@@.
@@@.@.@.@@
@@@@@.@.@@
@.@@@@..@.
@@.@@@@.@@
.@@@@@@@.@
.@.@.@.@@@
@.@@@.@@@@
.@@@@@@@@.
@.@.@@@.@.`

	res, err := day4.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.TotalRemoved != 43 {
		t.Fatalf("TotalRemoved=%d, want 43", res.TotalRemoved)
	}
}

func TestComputeIgnoresBlankLines(t *testing.T) {
	input := "\n@.\n\n.@\n"
	res, err := day4.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.TotalRemoved != 2 {
		t.Fatalf("TotalRemoved=%d, want 2", res.TotalRemoved)
	}
}

func TestComputeErrors(t *testing.T) {
	t.Run("NoGrid", func(t *testing.T) {
		if _, err := day4.Compute(bytes.NewBufferString("   \n\t")); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("NonRectangular", func(t *testing.T) {
		if _, err := day4.Compute(bytes.NewBufferString("@.\n..@\n")); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("InvalidChar", func(t *testing.T) {
		if _, err := day4.Compute(bytes.NewBufferString("@x\n@@\n")); err == nil {
			t.Fatalf("expected error")
		}
	})
}
