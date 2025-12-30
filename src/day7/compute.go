package day7

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
)

var (
	ErrNoGrid         = errors.New("no grid provided")
	ErrNonRectangular = errors.New("grid is not rectangular")
	ErrInvalidCell    = errors.New("invalid cell character")
	ErrNoStart        = errors.New("no start position (S) found")
	ErrMultipleStart  = errors.New("multiple start positions (S) found")
)

type Result struct {
	Timelines *big.Int
}

// Compute reads a room grid and counts how many distinct timelines (paths) a
// single downward-moving laser beam can take.
//
// When a beam hits a splitter ('^'), time splits: one timeline continues from
// the adjacent left column, another from the adjacent right column (both on the
// next row). Timelines are counted as distinct even if they later converge.
func Compute(r io.Reader) (Result, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	var grid [][]byte
	width := -1
	line := 0
	startX, startY := -1, -1

	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		if width == -1 {
			width = len(raw)
		}
		if len(raw) != width {
			return Result{}, fmt.Errorf("line %d: %w", line, ErrNonRectangular)
		}

		row := []byte(raw)
		for x, ch := range row {
			switch ch {
			case '.', '^':
				// ok
			case 'S':
				if startX != -1 {
					return Result{}, ErrMultipleStart
				}
				startX, startY = x, len(grid)
				row[x] = '.'
			default:
				return Result{}, fmt.Errorf("line %d col %d: %w: %q", line, x+1, ErrInvalidCell, ch)
			}
		}
		grid = append(grid, row)
	}

	if err := scanner.Err(); err != nil {
		return Result{}, err
	}
	if len(grid) == 0 {
		return Result{}, ErrNoGrid
	}
	if startX == -1 {
		return Result{}, ErrNoStart
	}

	h := len(grid)
	w := width
	if startY >= h-1 {
		return Result{Timelines: big.NewInt(1)}, nil
	}

	one := big.NewInt(1)
	next := make([]big.Int, w)
	curr := make([]big.Int, w)
	for i := range next {
		next[i].Set(one) // y==h base row: exiting yields one completed timeline
	}

	for y := h - 1; y >= startY+1; y-- {
		for x := 0; x < w; x++ {
			if grid[y][x] != '^' {
				curr[x].Set(&next[x])
				continue
			}

			left := one
			if x-1 >= 0 {
				left = &next[x-1]
			}
			right := one
			if x+1 < w {
				right = &next[x+1]
			}
			curr[x].Add(left, right)
		}
		curr, next = next, curr
	}

	return Result{Timelines: new(big.Int).Set(&next[startX])}, nil
}
