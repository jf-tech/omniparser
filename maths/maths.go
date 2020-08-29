package maths

// MaxInt returns the bigger value of the two input ints.
func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// MinInt returns the smaller value of the two input ints.
func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// MaxIntValue is the max value for type int.
// https://groups.google.com/forum/#!msg/golang-nuts/a9PitPAHSSU/ziQw1-QHw3EJ
const MaxIntValue = int(^uint(0) >> 1)
