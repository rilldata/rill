package worker

import (
	"testing"
)

func TestShouldScale(t *testing.T) {
	tests := []struct {
		name           string
		originSlots    int
		recommendSlots int
		want           bool
		wantReason     string
	}{
		{"No scaling down", 9, 8, false, scaledown},
		{"No scaling for small change", 50, 55, false, belowThreshold},
		{"No scaling for less than min scaling slots", 20, 24, false, belowThreshold},
		{"Scaling for significant change", 50, 60, true, ""},
		{"Scaling up for small services", 6, 10, true, ""},
		{"No scaling for the same quota", 77, 77, false, scaleMatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotReason := shouldScale(tt.originSlots, tt.recommendSlots)
			if got != tt.want || gotReason != tt.wantReason {
				t.Errorf("shouldScale(%d, %d) = (%v, %q); want (%v, %q)", tt.originSlots, tt.recommendSlots, got, gotReason, tt.want, tt.wantReason)
			}
		})
	}
}
