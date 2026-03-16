package admin

import "testing"

func TestCHCMinSlotsForMemory(t *testing.T) {
	tests := []struct {
		memoryGB float64
		want     int
	}{
		// Exact tier boundaries (1 slot = 4 GB)
		{8, 2},
		{12, 3},
		{16, 4},
		{32, 8},
		{64, 16},
		{120, 30},
		// Below smallest tier
		{4, 2},
		{1, 2},
		// Between tiers: rounds up to next tier
		{10, 3},
		{14, 4},
		{24, 8},
		{48, 16},
		{96, 30},
		// Above largest tier: clamps to biggest
		{128, 30},
		{256, 30},
		// Edge: zero
		{0, 2},
	}

	for _, tt := range tests {
		got := CHCMinSlotsForMemory(tt.memoryGB)
		if got != tt.want {
			t.Errorf("CHCMinSlotsForMemory(%v) = %d, want %d", tt.memoryGB, got, tt.want)
		}
	}
}
