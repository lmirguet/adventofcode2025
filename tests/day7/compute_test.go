package day7_test

import (
	"bytes"
	"testing"

	"adventofcode2025/day1/src/day7"
)

func TestComputeExample(t *testing.T) {
	input := `.......S.......
...............
.......^.......
...............
......^.^......
...............
.....^.^.^.....
...............
....^.^...^....
...............
...^.^...^.^...
...............
..^...^.....^..
...............
.^.^.^.^.^...^.
...............`

	res, err := day7.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.Timelines.String() != "40" {
		t.Fatalf("Timelines=%s, want 40", res.Timelines.String())
	}
}

func TestComputeStartOnLastRow(t *testing.T) {
	input := "...\n.S.\n"
	res, err := day7.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.Timelines.String() != "1" {
		t.Fatalf("Timelines=%s, want 1", res.Timelines.String())
	}
}

func TestComputeErrors(t *testing.T) {
	t.Run("NoGrid", func(t *testing.T) {
		if _, err := day7.Compute(bytes.NewBufferString("   \n\t\n")); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("NoStart", func(t *testing.T) {
		if _, err := day7.Compute(bytes.NewBufferString("...\n.^.\n...\n")); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("MultipleStart", func(t *testing.T) {
		if _, err := day7.Compute(bytes.NewBufferString("S.S\n...\n")); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("InvalidChar", func(t *testing.T) {
		if _, err := day7.Compute(bytes.NewBufferString("S..\n..x\n")); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("NonRectangular", func(t *testing.T) {
		if _, err := day7.Compute(bytes.NewBufferString("S..\n....\n")); err == nil {
			t.Fatalf("expected error")
		}
	})
}
