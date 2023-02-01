package gcs

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/c2h5oh/datasize"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/stretchr/testify/require"
)

type input struct {
	numFiles int
	path     string
}

func BenchmarkDownload10Files(b *testing.B) {
	benchmarkDownload(10, b)
}

func BenchmarkDownload20Files(b *testing.B) {
	benchmarkDownload(20, b)
}

func BenchmarkDownload100Files(b *testing.B) {
	benchmarkDownload(100, b)
}

func benchmarkDownload(num int, b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		c := connector{}
		iter, err := c.ConsumeAsIterator(ctx, &connectors.Env{}, &connectors.Source{
			Name:          "test",
			Connector:     "gcs",
			ExtractPolicy: extractPolicy(uint64(num)),
			Properties:    getMap("gs://ws1-teads.rilldata.com/beta-data/teads/teads_auction/v=1/y=2022/m=0[7-8]/**"),
		})
		require.NoError(b, err)
		defer iter.Close()

		files, err := iter.NextBatch(num)
		require.NoError(b, err)

		var size int64
		for _, f := range files {
			info, err := os.Stat(f)
			require.NoError(b, err)
			size += info.Size()
		}
		bytes := datasize.ByteSize(size)
		fmt.Printf("size of files %v\n", bytes.HumanReadable())
	}
}

func BenchmarkList10K(b *testing.B) {
	benchmarkList("gs://ws1-teads.rilldata.com/beta-data/teads/teads_auction/v=1/y=2022/m=0[7-8]/**", b)
}

func BenchmarkList100K(b *testing.B) {
	benchmarkList("gs://ws1-tvscientific.rilldata.com/beeswax-lambda-winlogs-prod/2021/08/**", b)
}

// func BenchmarkList500K(b *testing.B) {
// 	benchmarkList("gs://ws1-teads.rilldata.com/beta-data/teads/teads_auction/v=1/y=2022/m=0[7-8]/**", b)
// }

func benchmarkList(path string, b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		c := connector{}
		iter, err := c.ConsumeAsIterator(ctx, &connectors.Env{}, &connectors.Source{
			Name:          "test",
			Connector:     "gcs",
			ExtractPolicy: extractPolicy(uint64(20)),
			Properties:    getMap(path),
		})
		require.NoError(b, err)
		defer iter.Close()
	}
}

func getMap(path string) map[string]any {
	m := make(map[string]any)
	m["path"] = path
	m["glob.max_total_size"] = math.MaxInt64
	m["glob.max_objects_listed"] = math.MaxInt64
	m["glob.max_objects_matched"] = math.MaxInt64
	m["glob.page_size"] = 1000

	return m
}

func extractPolicy(n uint64) *runtimev1.Source_ExtractPolicy {
	return &runtimev1.Source_ExtractPolicy{FilesStrategy: runtimev1.Source_ExtractPolicy_STRATEGY_TAIL, FilesLimit: n}
}
