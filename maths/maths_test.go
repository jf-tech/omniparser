package maths

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinMaxInt(t *testing.T) {
	tests := []struct {
		name        string
		x           int
		y           int
		expectedMin int
		expectedMax int
	}{
		{
			name:        "x less than y",
			x:           1,
			y:           2,
			expectedMin: 1,
			expectedMax: 2,
		},
		{
			name:        "x greater than y",
			x:           2,
			y:           1,
			expectedMin: 1,
			expectedMax: 2,
		},
		{
			name:        "x equal to y",
			x:           2,
			y:           2,
			expectedMin: 2,
			expectedMax: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedMin, MinInt(test.x, test.y))
			assert.Equal(t, test.expectedMax, MaxInt(test.x, test.y))
		})
	}
}
