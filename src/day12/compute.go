package day12

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrNoShapes         = errors.New("no shapes provided")
	ErrNoRegions        = errors.New("no regions provided")
	ErrInvalidShape     = errors.New("invalid shape")
	ErrInvalidRegion    = errors.New("invalid region")
	ErrDuplicateShape   = errors.New("duplicate shape index")
	ErrMissingShape     = errors.New("missing shape definition")
	ErrCountOverflow    = errors.New("count overflow")
	ErrStateKeyTooLarge = errors.New("state key too large")
)

type Result struct {
	FittingRegions int64
}

type point struct {
	x int
	y int
}

type shape struct {
	cells    []point
	variants [][]point
	area     int
}

type region struct {
	w      int
	h      int
	counts []uint16
	lineNo int
}

// Compute parses present shapes and regions, and returns how many regions can
// fit the requested presents (with rotations/flips allowed, no overlaps).
func Compute(r io.Reader) (Result, error) {
	shapes, regions, err := parseInput(r)
	if err != nil {
		return Result{}, err
	}

	workerCount := runtime.NumCPU()
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > 13 {
		workerCount = 13
	}

	type job struct {
		idx int
		rg  region
	}
	type res struct {
		idx    int
		ok     bool
		lineNo int
		err    error
	}

	jobCh := make(chan job)
	resCh := make(chan res, workerCount)

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for j := range jobCh {
				ok, err := regionFits(shapes, j.rg)
				resCh <- res{idx: j.idx, ok: ok, lineNo: j.rg.lineNo, err: err}
			}
		}()
	}

	go func() {
		for idx := range regions {
			jobCh <- job{idx: idx, rg: regions[idx]}
		}
		close(jobCh)
		wg.Wait()
		close(resCh)
	}()

	var total int64
	var firstErr *res
	for r := range resCh {
		if r.err != nil {
			if firstErr == nil || r.lineNo < firstErr.lineNo {
				tmp := r
				firstErr = &tmp
			}
			continue
		}
		if r.ok {
			total++
		}
	}

	if firstErr != nil {
		return Result{}, fmt.Errorf("line %d: %w", firstErr.lineNo, firstErr.err)
	}
	return Result{FittingRegions: total}, nil
}

var (
	shapeHeaderRE = regexp.MustCompile(`^(\d+):$`)
	regionLineRE  = regexp.MustCompile(`^(\d+)x(\d+):`)
)

func parseInput(r io.Reader) ([]shape, []region, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, strings.TrimRight(scanner.Text(), "\r"))
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	// Find the start of the region section.
	regionStart := -1
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if regionLineRE.MatchString(line) {
			regionStart = i
			break
		}
	}
	if regionStart == -1 {
		regionStart = len(lines)
	}

	shapeMap := make(map[int]shape)
	maxIndex := -1

	i := 0
	for i < regionStart {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}

		m := shapeHeaderRE.FindStringSubmatch(line)
		if m == nil {
			return nil, nil, fmt.Errorf("line %d: %w: expected shape header, got %q", i+1, ErrInvalidShape, line)
		}

		idx, err := strconv.Atoi(m[1])
		if err != nil || idx < 0 {
			return nil, nil, fmt.Errorf("line %d: %w: invalid shape index %q", i+1, ErrInvalidShape, m[1])
		}
		if _, exists := shapeMap[idx]; exists {
			return nil, nil, fmt.Errorf("line %d: %w: %d", i+1, ErrDuplicateShape, idx)
		}
		if idx > maxIndex {
			maxIndex = idx
		}

		i++
		var diagram []string
		for i < regionStart {
			row := strings.TrimSpace(lines[i])
			if row == "" {
				break
			}
			if shapeHeaderRE.MatchString(row) || regionLineRE.MatchString(row) {
				return nil, nil, fmt.Errorf("line %d: %w: missing blank line after shape %d", i+1, ErrInvalidShape, idx)
			}
			diagram = append(diagram, row)
			i++
		}

		sh, err := parseShapeDiagram(diagram)
		if err != nil {
			return nil, nil, fmt.Errorf("shape %d: %w", idx, err)
		}
		shapeMap[idx] = sh
		if i < regionStart && strings.TrimSpace(lines[i]) == "" {
			i++
		}
	}

	if len(shapeMap) == 0 {
		return nil, nil, ErrNoShapes
	}

	shapes := make([]shape, maxIndex+1)
	for idx := 0; idx <= maxIndex; idx++ {
		sh, ok := shapeMap[idx]
		if !ok {
			return nil, nil, fmt.Errorf("%w: %d", ErrMissingShape, idx)
		}
		sh.variants = uniqueVariants(sh.cells)
		shapes[idx] = sh
	}

	var regions []region
	for j := regionStart; j < len(lines); j++ {
		line := strings.TrimSpace(lines[j])
		if line == "" {
			continue
		}
		rg, err := parseRegionLine(line, len(shapes))
		if err != nil {
			return nil, nil, fmt.Errorf("line %d: %w", j+1, err)
		}
		rg.lineNo = j + 1
		regions = append(regions, rg)
	}
	if len(regions) == 0 {
		return nil, nil, ErrNoRegions
	}

	return shapes, regions, nil
}

func parseShapeDiagram(diagram []string) (shape, error) {
	if len(diagram) == 0 {
		return shape{}, fmt.Errorf("%w: empty diagram", ErrInvalidShape)
	}

	var cells []point
	for y, row := range diagram {
		for x, ch := range row {
			switch ch {
			case '#':
				cells = append(cells, point{x: x, y: y})
			case '.':
			default:
				return shape{}, fmt.Errorf("%w: invalid rune %q", ErrInvalidShape, ch)
			}
		}
	}
	if len(cells) == 0 {
		return shape{}, fmt.Errorf("%w: no occupied cells", ErrInvalidShape)
	}
	return shape{cells: cells, area: len(cells)}, nil
}

func parseRegionLine(line string, shapeCount int) (region, error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return region{}, fmt.Errorf("%w: %q", ErrInvalidRegion, line)
	}
	dims := strings.TrimSpace(parts[0])
	dimParts := strings.SplitN(dims, "x", 2)
	if len(dimParts) != 2 {
		return region{}, fmt.Errorf("%w: invalid dimensions %q", ErrInvalidRegion, dims)
	}
	w, err := strconv.Atoi(dimParts[0])
	if err != nil || w <= 0 {
		return region{}, fmt.Errorf("%w: invalid width %q", ErrInvalidRegion, dimParts[0])
	}
	h, err := strconv.Atoi(dimParts[1])
	if err != nil || h <= 0 {
		return region{}, fmt.Errorf("%w: invalid height %q", ErrInvalidRegion, dimParts[1])
	}

	rawCounts := strings.Fields(strings.TrimSpace(parts[1]))
	if len(rawCounts) != shapeCount {
		return region{}, fmt.Errorf("%w: expected %d counts, got %d", ErrInvalidRegion, shapeCount, len(rawCounts))
	}

	counts := make([]uint16, shapeCount)
	for i, tok := range rawCounts {
		v, err := strconv.Atoi(tok)
		if err != nil || v < 0 || v > math.MaxUint16 {
			return region{}, fmt.Errorf("%w: invalid count %q", ErrInvalidRegion, tok)
		}
		counts[i] = uint16(v)
	}

	return region{w: w, h: h, counts: counts}, nil
}

type bitset []uint64

func newBitset(bitCount int) bitset {
	return make([]uint64, (bitCount+63)/64)
}

func (b bitset) clone() bitset {
	out := make([]uint64, len(b))
	copy(out, b)
	return out
}

func (b bitset) overlaps(other bitset) bool {
	for i := range b {
		if b[i]&other[i] != 0 {
			return true
		}
	}
	return false
}

func (b bitset) orInPlace(other bitset) {
	for i := range b {
		b[i] |= other[i]
	}
}

func regionFits(shapes []shape, rg region) (bool, error) {
	cellCount := rg.w * rg.h
	if cellCount <= 0 {
		return false, fmt.Errorf("%w: invalid region dimensions", ErrInvalidRegion)
	}

	var totalArea int
	for i, sh := range shapes {
		c := int(rg.counts[i])
		if c == 0 {
			continue
		}
		if sh.area > 0 && totalArea > math.MaxInt-sh.area*c {
			return false, ErrCountOverflow
		}
		totalArea += sh.area * c
	}
	if totalArea > cellCount {
		return false, nil
	}

	placements := make([][]bitset, len(shapes))
	for i, sh := range shapes {
		if rg.counts[i] == 0 {
			continue
		}
		ps := placementsForShape(sh.variants, rg.w, rg.h)
		if len(ps) == 0 {
			return false, nil
		}
		placements[i] = ps
	}

	memo := make(map[string]bool)

	encodeKey := func(occ bitset, counts []uint16) (string, error) {
		// Guard against accidentally huge states (should stay small for AoC inputs).
		if len(occ) > 1024 || len(counts) > 1024 {
			return "", ErrStateKeyTooLarge
		}
		buf := make([]byte, 0, len(occ)*8+len(counts)*2)
		var tmp8 [8]byte
		for _, w := range occ {
			tmp8[0] = byte(w)
			tmp8[1] = byte(w >> 8)
			tmp8[2] = byte(w >> 16)
			tmp8[3] = byte(w >> 24)
			tmp8[4] = byte(w >> 32)
			tmp8[5] = byte(w >> 40)
			tmp8[6] = byte(w >> 48)
			tmp8[7] = byte(w >> 56)
			buf = append(buf, tmp8[:]...)
		}
		var tmp2 [2]byte
		for _, c := range counts {
			tmp2[0] = byte(c)
			tmp2[1] = byte(c >> 8)
			buf = append(buf, tmp2[:]...)
		}
		return string(buf), nil
	}

	allZero := func(counts []uint16) bool {
		for _, c := range counts {
			if c != 0 {
				return false
			}
		}
		return true
	}

	var canFit func(occ bitset, counts []uint16) (bool, error)
	canFit = func(occ bitset, counts []uint16) (bool, error) {
		if allZero(counts) {
			return true, nil
		}

		key, err := encodeKey(occ, counts)
		if err != nil {
			return false, err
		}
		if v, ok := memo[key]; ok {
			return v, nil
		}

		bestIdx := -1
		bestOptions := -1

		for i, remaining := range counts {
			if remaining == 0 {
				continue
			}
			opts := 0
			for _, p := range placements[i] {
				if !occ.overlaps(p) {
					opts++
					if bestOptions != -1 && opts >= bestOptions {
						break
					}
				}
			}
			if opts == 0 {
				memo[key] = false
				return false, nil
			}
			if bestIdx == -1 || opts < bestOptions {
				bestIdx = i
				bestOptions = opts
				if bestOptions == 1 {
					break
				}
			}
		}

		counts[bestIdx]--
		for _, p := range placements[bestIdx] {
			if occ.overlaps(p) {
				continue
			}
			nextOcc := occ.clone()
			nextOcc.orInPlace(p)
			ok, err := canFit(nextOcc, counts)
			if err != nil {
				return false, err
			}
			if ok {
				counts[bestIdx]++
				memo[key] = true
				return true, nil
			}
		}
		counts[bestIdx]++

		memo[key] = false
		return false, nil
	}

	startOcc := newBitset(cellCount)
	return canFit(startOcc, append([]uint16(nil), rg.counts...))
}

func placementsForShape(variants [][]point, w, h int) []bitset {
	cellCount := w * h

	seen := make(map[string]struct{})
	var out []bitset

	for _, v := range variants {
		vw, vh := bounds(v)
		if vw > w || vh > h {
			continue
		}

		for oy := 0; oy <= h-vh; oy++ {
			for ox := 0; ox <= w-vw; ox++ {
				bs := newBitset(cellCount)
				for _, p := range v {
					x := ox + p.x
					y := oy + p.y
					pos := y*w + x
					bs[pos/64] |= 1 << uint(pos%64)
				}

				key := bitsetKey(bs)
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				out = append(out, bs)
			}
		}
	}

	return out
}

func bitsetKey(bs bitset) string {
	buf := make([]byte, 0, len(bs)*8)
	var tmp8 [8]byte
	for _, w := range bs {
		tmp8[0] = byte(w)
		tmp8[1] = byte(w >> 8)
		tmp8[2] = byte(w >> 16)
		tmp8[3] = byte(w >> 24)
		tmp8[4] = byte(w >> 32)
		tmp8[5] = byte(w >> 40)
		tmp8[6] = byte(w >> 48)
		tmp8[7] = byte(w >> 56)
		buf = append(buf, tmp8[:]...)
	}
	return string(buf)
}

func uniqueVariants(cells []point) [][]point {
	base := normalizePoints(cells)
	seen := make(map[string]struct{})
	var out [][]point

	for t := 0; t < 8; t++ {
		pts := make([]point, len(base))
		for i, p := range base {
			x, y := p.x, p.y
			if t >= 4 {
				x = -x
			}
			rot := t % 4
			for r := 0; r < rot; r++ {
				x, y = y, -x
			}
			pts[i] = point{x: x, y: y}
		}

		pts = normalizePoints(pts)
		k := pointsKey(pts)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, pts)
	}

	return out
}

func normalizePoints(pts []point) []point {
	minX := pts[0].x
	minY := pts[0].y
	for _, p := range pts[1:] {
		if p.x < minX {
			minX = p.x
		}
		if p.y < minY {
			minY = p.y
		}
	}

	out := make([]point, len(pts))
	for i, p := range pts {
		out[i] = point{x: p.x - minX, y: p.y - minY}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].y == out[j].y {
			return out[i].x < out[j].x
		}
		return out[i].y < out[j].y
	})
	return out
}

func pointsKey(pts []point) string {
	var b strings.Builder
	for _, p := range pts {
		b.WriteString(strconv.Itoa(p.x))
		b.WriteByte(',')
		b.WriteString(strconv.Itoa(p.y))
		b.WriteByte(';')
	}
	return b.String()
}

func bounds(pts []point) (w int, h int) {
	maxX := 0
	maxY := 0
	for _, p := range pts {
		if p.x > maxX {
			maxX = p.x
		}
		if p.y > maxY {
			maxY = p.y
		}
	}
	return maxX + 1, maxY + 1
}
