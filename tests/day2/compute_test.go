package day2_test

import (
	"bytes"
	"testing"

	"adventofcode2025/day1/src/day2"
)

func TestComputeExample(t *testing.T) {
	input := `11-22,95-115,998-1012,1188511880-1188511890,222220-222224,
1698522-1698528,446443-446449,38593856-38593862,565653-565659,
824824821-824824827,2121212118-2121212124`

	result, err := day2.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	expected := []int64{11, 22, 99, 111, 999, 1010, 222222, 446446, 565656, 38593859, 824824824, 1188511885, 2121212121}
	if len(result.InvalidIDs) != len(expected) {
		t.Fatalf("got %d invalid IDs, want %d", len(result.InvalidIDs), len(expected))
	}
	for i, v := range expected {
		if result.InvalidIDs[i] != v {
			t.Fatalf("invalid ID %d: got %d, want %d", i, result.InvalidIDs[i], v)
		}
	}

	var expectedSum int64
	for _, v := range expected {
		expectedSum += v
	}
	if result.Sum != expectedSum {
		t.Fatalf("sum mismatch: got %d, want %d", result.Sum, expectedSum)
	}
}

func TestComputeDeduplicatesOverlaps(t *testing.T) {
	input := "11-22,20-44"
	result, err := day2.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	expected := []int64{11, 22, 33, 44}
	if len(result.InvalidIDs) != len(expected) {
		t.Fatalf("expected %d IDs, got %d", len(expected), len(result.InvalidIDs))
	}
	for i, v := range expected {
		if result.InvalidIDs[i] != v {
			t.Fatalf("invalid ID %d: got %d, want %d", i, result.InvalidIDs[i], v)
		}
	}

	var expectedSum int64
	for _, v := range expected {
		expectedSum += v
	}
	if result.Sum != expectedSum {
		t.Fatalf("sum mismatch: got %d, want %d", result.Sum, expectedSum)
	}
}

func TestComputeHandlesMultiRepeats(t *testing.T) {
	input := "111-111,123123123-123123123,1212121212-1212121212"
	result, err := day2.Compute(bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	expected := []int64{111, 123123123, 1212121212}
	if len(result.InvalidIDs) != len(expected) {
		t.Fatalf("expected %d IDs, got %d", len(expected), len(result.InvalidIDs))
	}
	for i, v := range expected {
		if result.InvalidIDs[i] != v {
			t.Fatalf("invalid ID %d: got %d, want %d", i, result.InvalidIDs[i], v)
		}
	}

	var expectedSum int64
	for _, v := range expected {
		expectedSum += v
	}
	if result.Sum != expectedSum {
		t.Fatalf("sum mismatch: got %d, want %d", result.Sum, expectedSum)
	}
}
