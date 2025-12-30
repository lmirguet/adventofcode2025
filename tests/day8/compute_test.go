package day8_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"adventofcode2025/day1/src/day8"
)

func TestCompute_Sample(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "input8test.txt"))
	if err != nil {
		t.Fatalf("failed to read input8test.txt: %v", err)
	}

	res, err := day8.Compute(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if res.Product != 25272 {
		t.Fatalf("product=%d, want %d", res.Product, 25272)
	}
}

func TestCompute_Empty(t *testing.T) {
	if _, err := day8.Compute(bytes.NewBufferString("   \n\t\n")); err == nil {
		t.Fatalf("expected error on empty input")
	}
}
