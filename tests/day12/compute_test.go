package day12_test

import (
	"bytes"
	"os"
	"testing"

	"adventofcode2025/day1/src/day12"
)

func TestComputeSampleInputFile(t *testing.T) {
	payload, err := os.ReadFile("../../input12test.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	res, err := day12.Compute(bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.FittingRegions != 2 {
		t.Fatalf("FittingRegions=%d, want 2", res.FittingRegions)
	}
}
