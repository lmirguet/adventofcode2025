package day2

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrNoRanges        = errors.New("no ranges")
	ErrInvalidRange    = errors.New("invalid range")
	ErrInvalidBoundary = errors.New("invalid range boundary")
)

type Result struct {
	InvalidIDs []int64
	Sum        int64
}

// Compute parses comma-separated ranges from the reader, identifies the invalid IDs,
// and returns them alongside their sum. Invalid IDs consist of any digit sequence
// repeated at least twice (e.g. 11, 6464, 123123, 123123123) without leading zeroes.
func Compute(r io.Reader) (Result, error) {
	payload, err := io.ReadAll(r)
	if err != nil {
		return Result{}, err
	}

	ranges, err := parseRanges(string(payload))
	if err != nil {
		return Result{}, err
	}
	if len(ranges) == 0 {
		return Result{}, ErrNoRanges
	}

	seen := make(map[int64]struct{})
	var invalidIDs []int64
	var total int64

	for _, rg := range ranges {
		ids := invalidInRange(rg.Start, rg.End)
		for _, id := range ids {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			invalidIDs = append(invalidIDs, id)
			total += id
		}
	}

	sort.Slice(invalidIDs, func(i, j int) bool { return invalidIDs[i] < invalidIDs[j] })
	return Result{InvalidIDs: invalidIDs, Sum: total}, nil
}

type valueRange struct {
	Start int64
	End   int64
}

func parseRanges(input string) ([]valueRange, error) {
	chunks := strings.Split(input, ",")
	res := make([]valueRange, 0, len(chunks))

	for idx, raw := range chunks {
		token := strings.TrimSpace(raw)
		if token == "" {
			continue
		}

		parts := strings.Split(token, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: %s", ErrInvalidRange, token)
		}

		minVal, err := parseBoundary(parts[0])
		if err != nil {
			return nil, fmt.Errorf("range %d (%s): %w", idx+1, token, err)
		}
		maxVal, err := parseBoundary(parts[1])
		if err != nil {
			return nil, fmt.Errorf("range %d (%s): %w", idx+1, token, err)
		}
		if minVal > maxVal {
			return nil, fmt.Errorf("%w: start %d greater than end %d", ErrInvalidRange, minVal, maxVal)
		}
		res = append(res, valueRange{Start: minVal, End: maxVal})
	}

	return res, nil
}

func parseBoundary(part string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidBoundary, err)
	}
	return value, nil
}
