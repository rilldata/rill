package admin

import "testing"

func TestCHCMinSlotsForMemory(t *testing.T) {
	tests := []struct {
		memoryGB float64
		want     int
	}{
		// Exact tier boundaries
		{8, 4},
		{12, 6},
		{16, 8},
		{32, 16},
		{64, 32},
		{120, 60},
		// Below smallest tier
		{4, 4},
		{1, 4},
		// Between tiers: rounds up to next tier
		{10, 6},
		{14, 8},
		{24, 16},
		{48, 32},
		{96, 60},
		// Above largest tier: clamps to biggest
		{128, 60},
		{256, 60},
		// Edge: zero
		{0, 4},
	}

	for _, tt := range tests {
		got := CHCMinSlotsForMemory(tt.memoryGB)
		if got != tt.want {
			t.Errorf("CHCMinSlotsForMemory(%v) = %d, want %d", tt.memoryGB, got, tt.want)
		}
	}
}
