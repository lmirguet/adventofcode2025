package day9

import "math"

// minMax returns (min(a,b), max(a,b)).
func minMax(a, b int64) (int64, int64) {
	if a <= b {
		return a, b
	}
	return b, a
}

// uniqueSorted deduplicates a sorted slice in-place and returns the unique view.
func uniqueSorted(vals []int64) []int64 {
	if len(vals) == 0 {
		return nil
	}
	out := vals[:1]
	for i := 1; i < len(vals); i++ {
		if vals[i] == out[len(out)-1] {
			continue
		}
		out = append(out, vals[i])
	}
	return out
}

// absDiffInt64 returns |a-b| with overflow checking.
func absDiffInt64(a, b int64) (int64, error) {
	d, err := subInt64(a, b)
	if err != nil {
		return 0, err
	}
	if d == math.MinInt64 {
		return 0, ErrAreaOverflow
	}
	if d < 0 {
		d = -d
	}
	return d, nil
}

// subInt64 subtracts b from a with overflow checking.
func subInt64(a, b int64) (int64, error) {
	if (b > 0 && a < math.MinInt64+b) || (b < 0 && a > math.MaxInt64+b) {
		return 0, ErrAreaOverflow
	}
	return a - b, nil
}

// addInt64 adds two int64 values with overflow checking.
func addInt64(a, b int64) (int64, error) {
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < math.MinInt64-b) {
		return 0, ErrAreaOverflow
	}
	return a + b, nil
}

// mulInt64 multiplies two int64 values with overflow checking.
func mulInt64(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a == -1 && b == math.MinInt64 {
		return 0, ErrAreaOverflow
	}
	if b == -1 && a == math.MinInt64 {
		return 0, ErrAreaOverflow
	}
	if a > 0 {
		if b > 0 {
			if a > math.MaxInt64/b {
				return 0, ErrAreaOverflow
			}
		} else {
			if b < math.MinInt64/a {
				return 0, ErrAreaOverflow
			}
		}
	} else {
		if b > 0 {
			if a < math.MinInt64/b {
				return 0, ErrAreaOverflow
			}
		} else {
			if a != 0 && b < math.MaxInt64/a {
				return 0, ErrAreaOverflow
			}
		}
	}
	return a * b, nil
}
