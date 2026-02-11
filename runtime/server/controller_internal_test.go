package server

import (
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestModelPartitionToPB_IncludesRetryFields(t *testing.T) {
	now := time.Now().UTC()
	p := drivers.ModelPartition{
		Key:        "partition-1",
		DataJSON:   []byte(`{"day":"2026-01-01"}`),
		Watermark:  &now,
		ExecutedOn: &now,
		Error:      "failed",
		Elapsed:    1500 * time.Millisecond,
		RetryUsed:  2,
		RetryMax:   3,
	}

	got := modelPartitionToPB(p)

	require.Equal(t, "partition-1", got.Key)
	require.Equal(t, "failed", got.Error)
	require.Equal(t, uint32(1500), got.ElapsedMs)
	require.Equal(t, uint32(2), got.RetryUsed)
	require.Equal(t, uint32(3), got.RetryMax)
}
