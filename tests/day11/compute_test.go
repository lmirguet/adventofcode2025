package day11_test

import (
	"bytes"
	"os"
	"testing"

	"adventofcode2025/day1/src/day11"
)

func TestComputeSampleInputFile(t *testing.T) {
	payload, err := os.ReadFile("../../input11test.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	res, err := day11.Compute(bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.TotalPaths != 2 {
		t.Fatalf("TotalPaths=%d, want 2", res.TotalPaths)
	}
}

func TestComputeDeadEndIsZero(t *testing.T) {
	input := "svr: a\n"
	res, err := day11.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.TotalPaths != 0 {
		t.Fatalf("TotalPaths=%d, want 0", res.TotalPaths)
	}
}

func TestComputeCycleIsError(t *testing.T) {
	input := "svr: a\n" +
		"a: b\n" +
		"b: a out\n"
	_, err := day11.Compute(bytes.NewBufferString(input))
	if err == nil {
		t.Fatalf("expected error for cycle")
	}
}
