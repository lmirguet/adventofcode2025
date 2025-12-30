package day3

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrNoBanks      = errors.New("no banks provided")
	ErrInvalidDigit = errors.New("invalid digit in bank")
	ErrBankTooShort = errors.New("bank must contain at least 12 batteries")
)

const requiredBatteries = 12

type Result struct {
	Banks int
	Total int64
}

// Compute reads digit-only banks, picks the optimal set of 12 ordered batteries
// per bank, and returns the total joltage sum.
func Compute(r io.Reader) (Result, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	var total int64
	banks := 0
	line := 0

	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}

		value, err := maxBankJoltage(raw)
		if err != nil {
			return Result{}, fmt.Errorf("line %d: %w", line, err)
		}
		total += int64(value)
		banks++
	}

	if err := scanner.Err(); err != nil {
		return Result{}, err
	}
	if banks == 0 {
		return Result{}, ErrNoBanks
	}

	return Result{Banks: banks, Total: total}, nil
}

func maxBankJoltage(bank string) (int64, error) {
	n := len(bank)
	if n < requiredBatteries {
		return 0, ErrBankTooShort
	}

	toDrop := n - requiredBatteries
	stack := make([]byte, 0, n)

	for _, ch := range bank {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("%w: %q", ErrInvalidDigit, ch)
		}
		d := byte(ch - '0')
		for toDrop > 0 && len(stack) > 0 && stack[len(stack)-1] < d {
			stack = stack[:len(stack)-1]
			toDrop--
		}
		stack = append(stack, d)
	}

	// Trim to exact length if necessary.
	if len(stack) > requiredBatteries {
		stack = stack[:requiredBatteries]
	}

	var value int64
	for _, d := range stack {
		value = value*10 + int64(d)
	}
	return value, nil
}
