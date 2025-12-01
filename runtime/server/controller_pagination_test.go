package server

import (
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestModelPartitionPageToken_ExecutedOn(t *testing.T) {
	ts := time.Date(2024, time.March, 5, 12, 0, 0, 0, time.UTC)
	partition := drivers.ModelPartition{
		Key:        "abc123",
		ExecutedOn: &ts,
	}

	token := modelPartitionPageToken(partition)
	require.NotEmpty(t, token)

	var decoded time.Time
	var decodedKey string
	err := pagination.UnmarshalPageToken(token, &decoded, &decodedKey)
	require.NoError(t, err)
	require.True(t, ts.Equal(decoded))
	require.Equal(t, partition.Key, decodedKey)
}

func TestModelPartitionPageToken_NilExecutedOn(t *testing.T) {
	partition := drivers.ModelPartition{
		Key: "pending",
	}

	token := modelPartitionPageToken(partition)
	require.NotEmpty(t, token)

	var decoded time.Time
	var decodedKey string
	err := pagination.UnmarshalPageToken(token, &decoded, &decodedKey)
	require.NoError(t, err)
	require.True(t, decoded.IsZero())
	require.Equal(t, partition.Key, decodedKey)
}
