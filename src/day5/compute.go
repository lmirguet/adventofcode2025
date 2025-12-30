package day5

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrNoRanges        = errors.New("no fresh ranges provided")
	ErrInvalidRange    = errors.New("invalid range")
	ErrInvalidBoundary = errors.New("invalid range boundary")
	ErrCountOverflow   = errors.New("fresh id count overflow")
)

type Result struct {
	TotalFreshIDs int64
}

type valueRange struct {
	Start int64
	End   int64
}

// Compute reads a list of inclusive fresh ranges and returns how many distinct
// ingredient IDs are considered fresh by the union of these ranges.
//
// Any content after the first blank line is ignored.
func Compute(r io.Reader) (Result, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	var ranges []valueRange
	line := 0

	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			if len(ranges) == 0 {
				continue
			}
			break
		}

		rg, err := parseRange(raw)
		if err != nil {
			return Result{}, fmt.Errorf("line %d: %w", line, err)
		}
		ranges = append(ranges, rg)
	}

	if err := scanner.Err(); err != nil {
		return Result{}, err
	}
	if len(ranges) == 0 {
		return Result{}, ErrNoRanges
	}

	merged := mergeRanges(ranges)
	var total int64
	for _, rg := range merged {
		span := rg.End - rg.Start + 1
		if span < 0 {
			return Result{}, ErrCountOverflow
		}
		if total > math.MaxInt64-span {
			return Result{}, ErrCountOverflow
		}
		total += span
	}
	return Result{TotalFreshIDs: total}, nil
}

func parseRange(token string) (valueRange, error) {
	parts := strings.Split(token, "-")
	if len(parts) != 2 {
		return valueRange{}, fmt.Errorf("%w: %s", ErrInvalidRange, token)
	}

	start, err := parseBoundary(parts[0])
	if err != nil {
		return valueRange{}, err
	}
	end, err := parseBoundary(parts[1])
	if err != nil {
		return valueRange{}, err
	}
	if start > end {
		return valueRange{}, fmt.Errorf("%w: start %d greater than end %d", ErrInvalidRange, start, end)
	}
	return valueRange{Start: start, End: end}, nil
}

func parseBoundary(part string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidBoundary, err)
	}
	return value, nil
}

// mergeRanges takes a slice of inclusive integer ranges and returns a new
// slice where all overlapping or adjacent (touching) ranges are merged.
//
// Behaviour details:
//   - Ranges are sorted by start (and by end for equal starts) so we can
//     iterate once and merge progressively.
//   - Ranges are inclusive, so e.g. [1,3] and [4,6] should be merged into
//     [1,6] because 4 == 3+1 (they "touch").
//   - The function is careful to avoid integer overflow when checking for
//     adjacency: it only computes last.End+1 when last.End < math.MaxInt64.
//
// This implementation returns ranges in ascending order and with no
// overlapping or adjacent entries.
func mergeRanges(ranges []valueRange) []valueRange {
	// Sort by start ascending. When starts are equal, sort by end ascending
	// so the shorter range comes first — this simplifies merging logic.
	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].Start == ranges[j].Start {
			return ranges[i].End < ranges[j].End
		}
		return ranges[i].Start < ranges[j].Start
	})

	out := make([]valueRange, 0, len(ranges))
	for _, rg := range ranges {
		// If output is empty just append the first range.
		if len(out) == 0 {
			out = append(out, rg)
			continue
		}

		// last is the most recently appended/merged range in the output.
		last := &out[len(out)-1]

		// Inclusive ranges: merge if overlapping or touching.
		// - overlapping: rg.Start <= last.End
		// - touching: rg.Start == last.End+1 (but avoid computing +1 when
		//   last.End == math.MaxInt64 to prevent overflow)
		canTouch := last.End < math.MaxInt64
		touchesOrOverlaps := rg.Start <= last.End || (canTouch && rg.Start == last.End+1)

		if touchesOrOverlaps {
			// Extend the last range end if the incoming range goes further.
			if rg.End > last.End {
				last.End = rg.End
			}
			continue
		}

		// Disjoint range: append as a new entry.
		out = append(out, rg)
	}
	return out
}
