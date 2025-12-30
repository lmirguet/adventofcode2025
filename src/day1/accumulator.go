package day1

const dialSize = 100

func wrapDial(value int) int {
	r := value % dialSize
	if r < 0 {
		r += dialSize
	}
	return r
}

// Apply adjusts the accumulator based on direction and magnitude with wrap mod dialSize.
func Apply(acc int, dir rune, magnitude int) int {
	if dir == 'L' {
		return wrapDial(acc - magnitude)
	}
	return wrapDial(acc + magnitude)
}

// countZeroCrossings counts how many times a dial moving from acc by magnitude in dir
// lands on 0 (including when it stops there). Movement is stepwise +/-1 with wrap.
func countZeroCrossings(acc int, dir rune, magnitude int) int {
	if magnitude <= 0 {
		return 0
	}

	base := 0
	switch dir {
	case 'L':
		base = acc % dialSize
		if base == 0 {
			base = dialSize
		}
	case 'R':
		base = (dialSize - (acc % dialSize)) % dialSize
		if base == 0 {
			base = dialSize
		}
	default:
		return 0
	}

	if magnitude < base {
		return 0
	}
	return 1 + (magnitude-base)/dialSize
}
