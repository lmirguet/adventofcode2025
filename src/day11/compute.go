package day11

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
)

var (
	ErrNoDevices         = errors.New("no devices provided")
	ErrMissingStart      = errors.New("missing start device: svr")
	ErrInvalidDefinition = errors.New("invalid device definition")
	ErrDuplicateDevice   = errors.New("duplicate device definition")
	ErrCycleDetected     = errors.New("cycle detected in dataflow graph")
	ErrOverflow          = errors.New("path count overflow")
)

type Result struct {
	TotalPaths int64
}

// Compute reads a directed dataflow graph (one device per line) and returns the
// number of distinct paths from "svr" to "out" that visit both "dac" and "fft".
func Compute(r io.Reader) (Result, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	adj := make(map[string][]string)
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}

		name, outputs, err := parseDefinition(raw)
		if err != nil {
			return Result{}, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if _, exists := adj[name]; exists {
			return Result{}, fmt.Errorf("line %d: %w: %s", lineNo, ErrDuplicateDevice, name)
		}
		adj[name] = outputs
	}

	if err := scanner.Err(); err != nil {
		return Result{}, err
	}
	if len(adj) == 0 {
		return Result{}, ErrNoDevices
	}
	if _, ok := adj["svr"]; !ok {
		return Result{}, ErrMissingStart
	}

	const (
		colorUnseen   = 0
		colorVisiting = 1
		colorDone     = 2
	)

	const (
		bitDAC = 1 << iota
		bitFFT
	)
	const wantMask = bitDAC | bitFFT

	type state struct {
		node string
		mask uint8
	}

	colors := make(map[state]uint8)
	memo := make(map[state]int64)

	setMask := func(mask uint8, node string) uint8 {
		switch node {
		case "dac":
			return mask | bitDAC
		case "fft":
			return mask | bitFFT
		default:
			return mask
		}
	}

	var dfs func(node string, mask uint8) (int64, error)
	dfs = func(node string, mask uint8) (int64, error) {
		mask = setMask(mask, node)
		if node == "out" {
			if mask == wantMask {
				return 1, nil
			}
			return 0, nil
		}

		st := state{node: node, mask: mask}
		if v, ok := memo[st]; ok {
			return v, nil
		}

		switch colors[st] {
		case colorVisiting:
			return 0, ErrCycleDetected
		case colorDone:
			return memo[st], nil
		}

		colors[st] = colorVisiting
		var total int64

		for _, next := range adj[node] {
			paths, err := dfs(next, mask)
			if err != nil {
				return 0, err
			}
			if total > math.MaxInt64-paths {
				return 0, ErrOverflow
			}
			total += paths
		}

		colors[st] = colorDone
		memo[st] = total
		return total, nil
	}

	total, err := dfs("svr", 0)
	if err != nil {
		return Result{}, err
	}
	return Result{TotalPaths: total}, nil
}

func parseDefinition(line string) (string, []string, error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("%w: %s", ErrInvalidDefinition, line)
	}

	name := strings.TrimSpace(parts[0])
	if name == "" {
		return "", nil, fmt.Errorf("%w: empty device name", ErrInvalidDefinition)
	}

	right := strings.TrimSpace(parts[1])
	if right == "" {
		return name, nil, nil
	}
	return name, strings.Fields(right), nil
}
