package day8

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
	ErrNoJunctionBoxes   = errors.New("no junction boxes provided")
	ErrInvalidPosition   = errors.New("invalid position")
	ErrNotEnoughCircuits = errors.New("not enough circuits to compute top 3 product")
	ErrOverflow          = errors.New("overflow")
)

type Result struct {
	Product int64
}

type point struct {
	x int64
	y int64
	z int64
}

type edge struct {
	dist2 int64
	a     int
	b     int
}

// Compute connects junction boxes in ascending straight-line distance order
// until every box belongs to a single circuit (i.e. the graph becomes
// connected).
//
// The output is the product of the X coordinates of the first connection that
// achieves a single circuit (the last union operation that reduces the number
// of circuits to 1).
func Compute(r io.Reader) (Result, error) {
	points, err := parsePoints(r)
	if err != nil {
		return Result{}, err
	}
	n := len(points)
	if n == 0 {
		return Result{}, ErrNoJunctionBoxes
	}
	if n == 1 {
		return Result{}, ErrNotEnoughCircuits
	}

	edges := make([]edge, 0, n*(n-1)/2)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			d2, err := squaredDistance(points[i], points[j])
			if err != nil {
				return Result{}, err
			}

			e := edge{dist2: d2, a: i, b: j}
			edges = append(edges, e)
		}
	}

	dsu := newDSU(n)
	sort.Slice(edges, func(i, j int) bool { return lessEdge(edges[i], edges[j]) })

	circuits := n
	lastA, lastB := -1, -1
	for _, e := range edges {
		if !dsu.Union(e.a, e.b) {
			continue
		}
		circuits--
		lastA, lastB = e.a, e.b
		if circuits == 1 {
			break
		}
	}
	if circuits != 1 || lastA < 0 || lastB < 0 {
		return Result{}, errors.New("failed to connect all circuits")
	}

	product, err := mulInt64(points[lastA].x, points[lastB].x)
	if err != nil {
		return Result{}, err
	}
	return Result{Product: product}, nil
}

// parsePoints reads one point per non-empty line in the form `X,Y,Z`.
func parsePoints(r io.Reader) ([]point, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	var points []point
	line := 0
	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}

		parts := strings.Split(raw, ",")
		if len(parts) != 3 {
			return nil, fmt.Errorf("line %d: %w: expected X,Y,Z", line, ErrInvalidPosition)
		}
		x, err := parseCoord(parts[0])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w: X: %v", line, ErrInvalidPosition, err)
		}
		y, err := parseCoord(parts[1])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w: Y: %v", line, ErrInvalidPosition, err)
		}
		z, err := parseCoord(parts[2])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w: Z: %v", line, ErrInvalidPosition, err)
		}

		points = append(points, point{x: x, y: y, z: z})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return points, nil
}

// parseCoord parses a single coordinate as a base-10 int64.
func parseCoord(s string) (int64, error) {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

// squaredDistance returns the squared Euclidean distance between two points,
// using overflow-checked arithmetic.
func squaredDistance(a, b point) (int64, error) {
	dx, err := subInt64(a.x, b.x)
	if err != nil {
		return 0, err
	}
	dy, err := subInt64(a.y, b.y)
	if err != nil {
		return 0, err
	}
	dz, err := subInt64(a.z, b.z)
	if err != nil {
		return 0, err
	}

	x2, err := sqInt64(dx)
	if err != nil {
		return 0, err
	}
	y2, err := sqInt64(dy)
	if err != nil {
		return 0, err
	}
	z2, err := sqInt64(dz)
	if err != nil {
		return 0, err
	}

	sum := x2
	sum, err = addInt64(sum, y2)
	if err != nil {
		return 0, err
	}
	sum, err = addInt64(sum, z2)
	if err != nil {
		return 0, err
	}
	return sum, nil
}

// subInt64 subtracts b from a with overflow checking.
func subInt64(a, b int64) (int64, error) {
	if (b > 0 && a < math.MinInt64+b) || (b < 0 && a > math.MaxInt64+b) {
		return 0, ErrOverflow
	}
	return a - b, nil
}

// addInt64 adds two int64 values with overflow checking.
func addInt64(a, b int64) (int64, error) {
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < math.MinInt64-b) {
		return 0, ErrOverflow
	}
	return a + b, nil
}

// sqInt64 returns v*v with overflow checking.
func sqInt64(v int64) (int64, error) {
	if v == math.MinInt64 {
		return 0, ErrOverflow
	}
	abs := v
	if abs < 0 {
		abs = -abs
	}
	if abs > 0 && abs > math.MaxInt64/abs {
		return 0, ErrOverflow
	}
	return v * v, nil
}

// mul3 multiplies three int64 values with overflow checking.
func mul3(a, b, c int64) (int64, error) {
	ab, err := mulInt64(a, b)
	if err != nil {
		return 0, err
	}
	return mulInt64(ab, c)
}

// mulInt64 multiplies two int64 values with overflow checking.
func mulInt64(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a == -1 && b == math.MinInt64 {
		return 0, ErrOverflow
	}
	if b == -1 && a == math.MinInt64 {
		return 0, ErrOverflow
	}
	if a > 0 {
		if b > 0 {
			if a > math.MaxInt64/b {
				return 0, ErrOverflow
			}
		} else {
			if b < math.MinInt64/a {
				return 0, ErrOverflow
			}
		}
	} else {
		if b > 0 {
			if a < math.MinInt64/b {
				return 0, ErrOverflow
			}
		} else {
			if a != 0 && b < math.MaxInt64/a {
				return 0, ErrOverflow
			}
		}
	}
	return a * b, nil
}

// normalizeEdge orders endpoints so comparisons are consistent (a <= b).
func normalizeEdge(e edge) edge {
	if e.a > e.b {
		e.a, e.b = e.b, e.a
	}
	return e
}

// lessEdge implements a deterministic ordering for edges:
// by squared distance, then by endpoints (lexicographically).
func lessEdge(a, b edge) bool {
	a = normalizeEdge(a)
	b = normalizeEdge(b)
	if a.dist2 != b.dist2 {
		return a.dist2 < b.dist2
	}
	if a.a != b.a {
		return a.a < b.a
	}
	return a.b < b.b
}

type dsu struct {
	parent []int
	size   []int
}

// newDSU creates a disjoint-set union structure for n singleton components.
func newDSU(n int) *dsu {
	parent := make([]int, n)
	size := make([]int, n)
	for i := 0; i < n; i++ {
		parent[i] = i
		size[i] = 1
	}
	return &dsu{parent: parent, size: size}
}

// Find returns the representative/root of x, using path compression.
func (d *dsu) Find(x int) int {
	for x != d.parent[x] {
		d.parent[x] = d.parent[d.parent[x]]
		x = d.parent[x]
	}
	return x
}

// Union merges the components of a and b and returns true if they were distinct.
func (d *dsu) Union(a, b int) bool {
	ra := d.Find(a)
	rb := d.Find(b)
	if ra == rb {
		return false
	}
	if d.size[ra] < d.size[rb] {
		ra, rb = rb, ra
	}
	d.parent[rb] = ra
	d.size[ra] += d.size[rb]
	return true
}
