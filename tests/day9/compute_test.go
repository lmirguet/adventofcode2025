package day9_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"adventofcode2025/day1/src/day9"
)

func TestCompute_Sample(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "input9test.txt"))
	if err != nil {
		t.Fatalf("failed to read input9test.txt: %v", err)
	}

	res, err := day9.Compute(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.MaxArea != 24 {
		t.Fatalf("maxArea=%d, want %d", res.MaxArea, 24)
	}
}

func TestCompute_Empty(t *testing.T) {
	if _, err := day9.Compute(bytes.NewBufferString("   \n\t\n")); err == nil {
		t.Fatalf("expected error on empty input")
	}
}
