package day10_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"adventofcode2025/day1/src/day10"
)

func TestCompute_Sample(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "input10test.txt"))
	if err != nil {
		t.Fatalf("failed to read input10test.txt: %v", err)
	}

	res, err := day10.Compute(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.TotalPresses != 33 {
		t.Fatalf("totalPresses=%d, want %d", res.TotalPresses, 33)
	}
}

func TestCompute_Empty(t *testing.T) {
	if _, err := day10.Compute(bytes.NewBufferString("   \n\t\n")); err == nil {
		t.Fatalf("expected error on empty input")
	}
}

func TestCompute_LargerRequirements(t *testing.T) {
	input := `[..#.###.#.] (0,3,4,5) (1,4,6) (2,9) (0,4) (2,4,7,8) (0,2,3,4,5,6,7,8,9) (1,6) (1,2,5,6,7) (0,4,7,8) (0,1,2,3,5,8,9) (0,3,4,5,6,7,8,9) (4,6,9) {56,51,67,27,82,44,70,56,49,58}
`
	res, err := day10.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.TotalPresses != 132 {
		t.Fatalf("totalPresses=%d, want %d", res.TotalPresses, 132)
	}
}

func TestCompute_SingularSquareMatrix(t *testing.T) {
	input := `[.#####.] (1,2,3,5,6) (0,1) (0,1,3,5) (2,3) (0,3,6) (0,1,2,4,5,6) (0,4,5) {154,26,19,142,16,27,114}
`
	res, err := day10.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.TotalPresses != 172 {
		t.Fatalf("totalPresses=%d, want %d", res.TotalPresses, 172)
	}
}
