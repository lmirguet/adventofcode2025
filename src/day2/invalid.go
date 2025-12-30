package day2

import (
	"math"
)

func invalidInRange(start, end int64) []int64 {
	if end < start || end < 0 {
		return nil
	}
	if start < 0 {
		start = 0
	}

	minDigits := digitCount(start)
	if minDigits < 2 {
		minDigits = 2
	}
	maxDigits := digitCount(end)

	var result []int64
	for digits := minDigits; digits <= maxDigits; digits++ {
		maxChunkDigits := digits / 2
		if maxChunkDigits == 0 {
			continue
		}

		for chunkDigits := 1; chunkDigits <= maxChunkDigits; chunkDigits++ {
			if digits%chunkDigits != 0 {
				continue
			}

			repeat := digits / chunkDigits
			if repeat < 2 {
				continue
			}

			base := pow10(chunkDigits)
			if base == 0 {
				continue
			}
			factor := repeatFactor(base, repeat)
			if factor == 0 {
				continue
			}

			chunkMin := pow10(chunkDigits - 1)
			chunkMax := base - 1

			startChunk := ceilDiv(start, factor)
			if startChunk > chunkMin {
				chunkMin = startChunk
			}
			endChunk := end / factor
			if endChunk < chunkMax {
				chunkMax = endChunk
			}
			if chunkMin > chunkMax {
				continue
			}

			for chunk := chunkMin; chunk <= chunkMax; chunk++ {
				candidate := chunk * factor
				if candidate >= start && candidate <= end {
					result = append(result, candidate)
				}
			}
		}
	}

	return result
}

func digitCount(n int64) int {
	if n < 0 {
		n = -n
	}
	if n == 0 {
		return 1
	}
	count := 0
	for n > 0 {
		count++
		n /= 10
	}
	return count
}

func pow10(exp int) int64 {
	if exp <= 0 {
		return 1
	}
	result := int64(1)
	for i := 0; i < exp; i++ {
		if result > math.MaxInt64/10 {
			return 0
		}
		result *= 10
	}
	return result
}

func ceilDiv(a, b int64) int64 {
	if b == 0 {
		return 0
	}
	if a <= 0 {
		return 0
	}
	q := a / b
	if a%b != 0 {
		q++
	}
	return q
}

func repeatFactor(base int64, repeat int) int64 {
	factor := int64(0)
	for i := 0; i < repeat; i++ {
		if factor > (math.MaxInt64-1)/base {
			return 0
		}
		factor = factor*base + 1
	}
	return factor
}
