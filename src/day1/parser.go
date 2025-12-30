package day1

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

var (
	ErrNoInstructions   = errors.New("no instructions")
	ErrInvalidDirection = errors.New("invalid direction")
	ErrInvalidMagnitude = errors.New("invalid magnitude")
)

const InitialValue = 50

type Result struct {
	Position      int
	ZeroCrossings int
}

// ParseToken validates a single token of the form `[L|R][0-9]+`.
func ParseToken(token string) (rune, int, error) {
	if len(token) < 2 {
		return 0, 0, ErrInvalidMagnitude
	}

	dir := rune(token[0])
	if dir != 'L' && dir != 'R' {
		return 0, 0, ErrInvalidDirection
	}

	mag, err := strconv.Atoi(token[1:])
	if err != nil || mag <= 0 {
		return 0, 0, ErrInvalidMagnitude
	}

	return dir, mag, nil
}

// Compute parses tokens from the reader and applies them to the accumulator.
// Returns an error with token position context on invalid input.
// Tracks how many times the dial passes through 0 while applying instructions.
func Compute(r io.Reader) (Result, error) {
	scanner := bufio.NewScanner(r)
	// Increase buffer in case of large magnitudes.
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)
	scanner.Split(bufio.ScanWords)

	position := 0
	acc := InitialValue
	sawAny := false
	totalZeroCrossings := 0

	for scanner.Scan() {
		position++
		token := scanner.Text()
		dir, mag, err := ParseToken(token)
		if err != nil {
			return Result{}, fmt.Errorf("token %d (%s): %w", position, token, err)
		}
		totalZeroCrossings += countZeroCrossings(acc, dir, mag)
		acc = Apply(acc, dir, mag)
		sawAny = true
	}

	if err := scanner.Err(); err != nil {
		return Result{}, err
	}
	if !sawAny {
		return Result{}, ErrNoInstructions
	}
	return Result{Position: acc, ZeroCrossings: totalZeroCrossings}, nil
}
