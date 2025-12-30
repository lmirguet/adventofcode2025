package day6

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"math/bits"
	"strconv"
	"strings"
)

var (
	ErrNoWorksheet     = errors.New("no worksheet provided")
	ErrNoProblems      = errors.New("no problems found")
	ErrMissingOperator = errors.New("missing operator")
	ErrInvalidOperator = errors.New("invalid operator")
	ErrInvalidNumber   = errors.New("invalid number")
	ErrOverflow        = errors.New("overflow")
)

type Result struct {
	Total int64
}

// Compute parses a worksheet made of vertically-arranged problems separated by
// full columns of spaces and returns the grand total of all problem results.
func Compute(r io.Reader) (Result, error) {
	lines, err := readLines(r)
	if err != nil {
		return Result{}, err
	}
	if len(lines) == 0 {
		return Result{}, ErrNoWorksheet
	}

	// Trim leading/trailing empty (all-space) lines.
	for len(lines) > 0 && isAllSpaces(lines[0]) {
		lines = lines[1:]
	}
	for len(lines) > 0 && isAllSpaces(lines[len(lines)-1]) {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return Result{}, ErrNoWorksheet
	}

	width := 0
	for _, ln := range lines {
		if len(ln) > width {
			width = len(ln)
		}
	}
	if width == 0 {
		return Result{}, ErrNoWorksheet
	}
	for i := range lines {
		if len(lines[i]) < width {
			lines[i] = lines[i] + strings.Repeat(" ", width-len(lines[i]))
		}
	}

	opRow := len(lines) - 1
	if opRow < 0 {
		return Result{}, ErrNoWorksheet
	}

	segments := splitByBlankColumns(lines, width)
	if len(segments) == 0 {
		return Result{}, ErrNoProblems
	}

	var total int64
	for segIdx, seg := range segments {
		op, err := parseOperator(lines[opRow][seg.start:seg.end])
		if err != nil {
			return Result{}, fmt.Errorf("problem %d: %w", segIdx+1, err)
		}

		nums, err := parseColumnNumbers(lines[:opRow], seg.start, seg.end)
		if err != nil {
			return Result{}, fmt.Errorf("problem %d: %w", segIdx+1, err)
		}
		if len(nums) == 0 {
			return Result{}, fmt.Errorf("problem %d: %w", segIdx+1, ErrInvalidNumber)
		}

		value, err := eval(nums, op)
		if err != nil {
			return Result{}, fmt.Errorf("problem %d: %w", segIdx+1, err)
		}

		total, err = addInt64(total, value)
		if err != nil {
			return Result{}, fmt.Errorf("problem %d: %w", segIdx+1, err)
		}
	}

	return Result{Total: total}, nil
}

type segment struct {
	start int
	end   int // exclusive
}

func splitByBlankColumns(lines []string, width int) []segment {
	isBlankCol := make([]bool, width)
	for x := 0; x < width; x++ {
		allSpace := true
		for _, ln := range lines {
			if ln[x] != ' ' {
				allSpace = false
				break
			}
		}
		isBlankCol[x] = allSpace
	}

	var segs []segment
	in := false
	start := 0
	for x := 0; x < width; x++ {
		if isBlankCol[x] {
			if in {
				segs = append(segs, segment{start: start, end: x})
				in = false
			}
			continue
		}
		if !in {
			in = true
			start = x
		}
	}
	if in {
		segs = append(segs, segment{start: start, end: width})
	}

	return segs
}

func parseOperator(cell string) (rune, error) {
	first := rune(0)
	for _, ch := range cell {
		if ch == ' ' {
			continue
		}
		if ch != '+' && ch != '*' {
			return 0, fmt.Errorf("%w: %q", ErrInvalidOperator, ch)
		}
		if first == 0 {
			first = ch
			continue
		}
		// More than one operator char in the segment.
		return 0, ErrInvalidOperator
	}
	if first == 0 {
		return 0, ErrMissingOperator
	}
	return first, nil
}

func parseColumnNumbers(lines []string, start, end int) ([]int64, error) {
	var nums []int64
	for x := end - 1; x >= start; x-- {
		var digits strings.Builder
		for rowIdx, ln := range lines {
			ch := ln[x]
			if ch == ' ' {
				continue
			}
			if ch < '0' || ch > '9' {
				return nil, fmt.Errorf("row %d col %d: %w: %q", rowIdx+1, x+1, ErrInvalidNumber, ch)
			}
			digits.WriteByte(ch)
		}
		if digits.Len() == 0 {
			continue
		}
		n, err := strconv.ParseInt(digits.String(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("col %d: %w: %v", x+1, ErrInvalidNumber, err)
		}
		nums = append(nums, n)
	}
	return nums, nil
}

func eval(nums []int64, op rune) (int64, error) {
	switch op {
	case '+':
		var sum int64
		for _, n := range nums {
			var err error
			sum, err = addInt64(sum, n)
			if err != nil {
				return 0, err
			}
		}
		return sum, nil
	case '*':
		prod := int64(1)
		for _, n := range nums {
			var err error
			prod, err = mulInt64(prod, n)
			if err != nil {
				return 0, err
			}
		}
		return prod, nil
	default:
		return 0, ErrInvalidOperator
	}
}

func addInt64(a, b int64) (int64, error) {
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < math.MinInt64-b) {
		return 0, ErrOverflow
	}
	return a + b, nil
}

func mulInt64(a, b int64) (int64, error) {
	// Use unsigned multiplication on absolute values to detect overflow precisely.
	if a == 0 || b == 0 {
		return 0, nil
	}

	neg := (a < 0) != (b < 0)
	ua, ok := absToUint64(a)
	if !ok {
		return 0, ErrOverflow
	}
	ub, ok := absToUint64(b)
	if !ok {
		return 0, ErrOverflow
	}

	hi, lo := bits.Mul64(ua, ub)
	if hi != 0 {
		return 0, ErrOverflow
	}

	if neg {
		if lo > uint64(-math.MinInt64) {
			return 0, ErrOverflow
		}
		return -int64(lo), nil
	}

	if lo > uint64(math.MaxInt64) {
		return 0, ErrOverflow
	}
	return int64(lo), nil
}

func absToUint64(v int64) (uint64, bool) {
	if v >= 0 {
		return uint64(v), true
	}
	if v == math.MinInt64 {
		return 0, false
	}
	return uint64(-v), true
}

func readLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func isAllSpaces(s string) bool {
	for _, ch := range s {
		if ch != ' ' && ch != '\t' && ch != '\r' {
			return false
		}
	}
	return true
}
