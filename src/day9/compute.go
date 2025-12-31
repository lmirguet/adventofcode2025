package day9

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
	ErrNoTiles        = errors.New("no tiles provided")
	ErrInvalidTile    = errors.New("invalid tile coordinate")
	ErrInvalidLoop    = errors.New("invalid loop")
	ErrAreaOverflow   = errors.New("area overflow")
	ErrNotEnoughTiles = errors.New("need at least two tiles")
)

type Result struct {
	MaxArea int64
}

type tile struct {
	x int64
	y int64
}

type block struct {
	start int64
	end   int64
}

// Compute reads an ordered loop of red tiles.
//
// Consecutive tiles (and the last->first) are connected by axis-aligned paths
// of green tiles; all tiles inside the loop are also green. The remaining tiles
// are neither red nor green.
//
// The chosen rectangle must have red tiles as opposite corners, and every tile
// it contains must be red or green (i.e. the rectangle must be fully contained
// within the loop region). The function returns the maximum such rectangle area
// using inclusive grid dimensions:
// width = |x1-x2| + 1, height = |y1-y2| + 1, area = width*height.
func Compute(r io.Reader) (Result, error) {
	tiles, err := parseTiles(r)
	if err != nil {
		return Result{}, err
	}
	if len(tiles) == 0 {
		return Result{}, ErrNoTiles
	}
	if len(tiles) < 2 {
		return Result{}, ErrNotEnoughTiles
	}
	if err := validateLoop(tiles); err != nil {
		return Result{}, err
	}

	minX, maxX := tiles[0].x, tiles[0].x
	minY, maxY := tiles[0].y, tiles[0].y
	for _, t := range tiles[1:] {
		if t.x < minX {
			minX = t.x
		}
		if t.x > maxX {
			maxX = t.x
		}
		if t.y < minY {
			minY = t.y
		}
		if t.y > maxY {
			maxY = t.y
		}
	}

	xBlocks, xIndex := buildBlocksWithIndex(tiles, func(t tile) int64 { return t.x }, minX, maxX)
	yBlocks, yIndex := buildBlocksWithIndex(tiles, func(t tile) int64 { return t.y }, minY, maxY)

	insidePrefix, err := buildInsidePrefixSums(tiles, xBlocks, yBlocks)
	if err != nil {
		return Result{}, err
	}

	var maxArea int64
	for i := 0; i < len(tiles); i++ {
		for j := i + 1; j < len(tiles); j++ {
			x1, x2 := tiles[i].x, tiles[j].x
			y1, y2 := tiles[i].y, tiles[j].y
			xMin, xMax := minMax(x1, x2)
			yMin, yMax := minMax(y1, y2)

			xiMin, ok := xIndex[xMin]
			if !ok {
				return Result{}, ErrInvalidLoop
			}
			xiMax, ok := xIndex[xMax]
			if !ok {
				return Result{}, ErrInvalidLoop
			}
			yiMin, ok := yIndex[yMin]
			if !ok {
				return Result{}, ErrInvalidLoop
			}
			yiMax, ok := yIndex[yMax]
			if !ok {
				return Result{}, ErrInvalidLoop
			}

			rectArea, err := inclusiveAreaFromBounds(xMin, xMax, yMin, yMax)
			if err != nil {
				return Result{}, err
			}

			inside := rectSum(insidePrefix, len(yBlocks), xiMin, yiMin, xiMax, yiMax)
			if inside == rectArea && rectArea > maxArea {
				maxArea = rectArea
			}
		}
	}

	return Result{MaxArea: maxArea}, nil
}

// parseTiles reads one coordinate per non-empty line in the form `x,y` and
// returns them in input order (which matters, because it defines the loop).
func parseTiles(r io.Reader) ([]tile, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	var tiles []tile
	line := 0
	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}

		parts := strings.Split(raw, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: %w: expected x,y", line, ErrInvalidTile)
		}

		x, err := parseCoord(parts[0])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w: x: %v", line, ErrInvalidTile, err)
		}
		y, err := parseCoord(parts[1])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w: y: %v", line, ErrInvalidTile, err)
		}

		tiles = append(tiles, tile{x: x, y: y})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return tiles, nil
}

// parseCoord parses a single coordinate as a base-10 int64.
func parseCoord(s string) (int64, error) {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

// validateLoop ensures the input describes an orthogonal closed loop: each
// consecutive pair (including last->first) must share X or Y, and must not be
// identical.
func validateLoop(tiles []tile) error {
	if len(tiles) < 4 {
		return ErrInvalidLoop
	}
	for i := 0; i < len(tiles); i++ {
		a := tiles[i]
		b := tiles[(i+1)%len(tiles)]
		if a.x != b.x && a.y != b.y {
			return fmt.Errorf("%w: consecutive tiles must share x or y", ErrInvalidLoop)
		}
		if a.x == b.x && a.y == b.y {
			return fmt.Errorf("%w: consecutive tiles must not be identical", ErrInvalidLoop)
		}
	}
	return nil
}

// inclusiveAreaFromBounds computes the inclusive area for the rectangle defined
// by (xMin,yMin) and (xMax,yMax), with overflow checks.
func inclusiveAreaFromBounds(xMin, xMax, yMin, yMax int64) (int64, error) {
	dx, err := absDiffInt64(xMax, xMin)
	if err != nil {
		return 0, ErrAreaOverflow
	}
	dy, err := absDiffInt64(yMax, yMin)
	if err != nil {
		return 0, ErrAreaOverflow
	}

	width, err := addInt64(dx, 1)
	if err != nil {
		return 0, ErrAreaOverflow
	}
	height, err := addInt64(dy, 1)
	if err != nil {
		return 0, ErrAreaOverflow
	}

	area, err := mulInt64(width, height)
	if err != nil {
		return 0, ErrAreaOverflow
	}
	return area, nil
}

// buildBlocksWithIndex builds a 1D coordinate compression for an axis.
//
// It returns:
//   - a list of contiguous blocks (each block represents many integer coordinates),
//   - and an index mapping from every "single coordinate" block (start==end)
//     to its block index.
//
// We include coordinates that appear in the input, plus +/-1 of those values
// (clamped to [minV,maxV]) so that rectangle edges align exactly with block
// boundaries and every block is uniformly inside/outside the loop.
func buildBlocksWithIndex(tiles []tile, coord func(tile) int64, minV, maxV int64) ([]block, map[int64]int) {
	breaks := make([]int64, 0, len(tiles)*3)
	for _, t := range tiles {
		v := coord(t)
		breaks = append(breaks, v)
		if v > math.MinInt64 {
			breaks = append(breaks, v-1)
		}
		if v < math.MaxInt64 {
			breaks = append(breaks, v+1)
		}
	}

	sort.Slice(breaks, func(i, j int) bool { return breaks[i] < breaks[j] })
	breaks = uniqueSorted(breaks)

	var inRange []int64
	inRange = append(inRange, minV, maxV)
	for _, v := range breaks {
		if v < minV || v > maxV {
			continue
		}
		inRange = append(inRange, v)
	}
	sort.Slice(inRange, func(i, j int) bool { return inRange[i] < inRange[j] })
	inRange = uniqueSorted(inRange)

	var blocks []block
	singles := make(map[int64]int, len(inRange))

	prev := int64(0)
	havePrev := false
	for _, b := range inRange {
		if !havePrev {
			if b > minV {
				blocks = append(blocks, block{start: minV, end: b - 1})
			}
		} else if b > prev+1 {
			blocks = append(blocks, block{start: prev + 1, end: b - 1})
		}
		blocks = append(blocks, block{start: b, end: b})
		singles[b] = len(blocks) - 1
		prev = b
		havePrev = true
	}
	if havePrev && prev < maxV {
		blocks = append(blocks, block{start: prev + 1, end: maxV})
	}

	return blocks, singles
}

// buildInsidePrefixSums builds a weighted 2D prefix-sum array over the
// coordinate-compressed grid.
//
// Each (xi,yi) cell represents a block of original tiles with size
// blockLen(xBlocks[xi]) * blockLen(yBlocks[yi]). We mark the cell as "inside"
// if a representative point of the block is inside or on the loop boundary.
//
// The returned prefix array is (nx+1) x (ny+1) flattened row-major, suitable
// for O(1) rectangle sum queries via rectSum().
func buildInsidePrefixSums(loop []tile, xBlocks, yBlocks []block) ([]int64, error) {
	nx := len(xBlocks)
	ny := len(yBlocks)
	prefix := make([]int64, (nx+1)*(ny+1))

	for xi := 0; xi < nx; xi++ {
		xCount, err := blockLen(xBlocks[xi])
		if err != nil {
			return nil, ErrAreaOverflow
		}
		rowSum := int64(0)
		for yi := 0; yi < ny; yi++ {
			yCount, err := blockLen(yBlocks[yi])
			if err != nil {
				return nil, ErrAreaOverflow
			}

			rep := tile{x: xBlocks[xi].start, y: yBlocks[yi].start}
			cell := int64(0)
			if pointInPolygonInclusive(rep, loop) {
				cell, err = mulInt64(xCount, yCount)
				if err != nil {
					return nil, ErrAreaOverflow
				}
			}

			rowSum, err = addInt64(rowSum, cell)
			if err != nil {
				return nil, ErrAreaOverflow
			}

			above := prefix[(xi)*(ny+1)+(yi+1)]
			prefix[(xi+1)*(ny+1)+(yi+1)], err = addInt64(above, rowSum)
			if err != nil {
				return nil, ErrAreaOverflow
			}
		}
	}

	return prefix, nil
}

// blockLen returns the number of integer coordinates covered by a block
// (inclusive), with overflow checks.
func blockLen(b block) (int64, error) {
	if b.end < b.start {
		return 0, ErrAreaOverflow
	}
	d, err := subInt64(b.end, b.start)
	if err != nil {
		return 0, ErrAreaOverflow
	}
	return addInt64(d, 1)
}

// rectSum returns the sum of weights within the inclusive rectangle
// [x1..x2] x [y1..y2] in block-index space, using the prefix-sum array.
func rectSum(prefix []int64, ny int, x1, y1, x2, y2 int) int64 {
	// inclusive indices in block space
	x1++
	y1++
	x2++
	y2++

	rowStride := ny + 1
	a := prefix[x2*rowStride+y2]
	b := prefix[(x1-1)*rowStride+y2]
	c := prefix[x2*rowStride+(y1-1)]
	d := prefix[(x1-1)*rowStride+(y1-1)]
	return a - b - c + d
}

// pointInPolygonInclusive returns true if p is inside the loop polygon or on
// its boundary.
//
// The polygon is orthogonal (axis-aligned edges). We use an even/odd rule with
// a horizontal ray cast to +infinity on X, counting crossings with vertical
// edges using a half-open Y-interval to avoid double-counting vertices.
func pointInPolygonInclusive(p tile, poly []tile) bool {
	inside := false
	for i := 0; i < len(poly); i++ {
		a := poly[i]
		b := poly[(i+1)%len(poly)]

		if onAxisAlignedSegment(p, a, b) {
			return true
		}

		if a.x != b.x { // horizontal edges do not affect horizontal ray crossings
			continue
		}

		x := a.x
		y1 := a.y
		y2 := b.y
		if y1 > y2 {
			y1, y2 = y2, y1
		}

		// Half-open interval on Y to avoid double-counting vertices.
		if p.y >= y1 && p.y < y2 && p.x < x {
			inside = !inside
		}
	}
	return inside
}

// onAxisAlignedSegment reports whether point p lies on the closed segment a->b,
// assuming a->b is axis-aligned (horizontal or vertical).
func onAxisAlignedSegment(p, a, b tile) bool {
	if a.x == b.x {
		if p.x != a.x {
			return false
		}
		lo, hi := minMax(a.y, b.y)
		return p.y >= lo && p.y <= hi
	}
	if a.y == b.y {
		if p.y != a.y {
			return false
		}
		lo, hi := minMax(a.x, b.x)
		return p.x >= lo && p.x <= hi
	}
	return false
}

// (utility functions were moved to utils.go)
