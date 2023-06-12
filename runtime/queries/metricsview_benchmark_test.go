package queries

import (
	"bufio"
	"fmt"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"io/ioutil"
	"os"
	"testing"
)

func Benchmark_writeCSV(b *testing.B) {
	meta := []*runtimev1.MetricsViewColumn{}
	fields := make(map[string]*structpb.Value)
	data := []*structpb.Struct{}

	for i := 0; i < 100; i++ {
		col := fmt.Sprintf("col%d", i)
		meta = append(meta, &runtimev1.MetricsViewColumn{
			Name: fmt.Sprintf("col%d", i),
		})

		fields[col] = structpb.NewStringValue(col)
		for j := 0; j < 10000; j++ {
			data = append(data, &structpb.Struct{
				Fields: fields,
			})
		}
	}

	file, err := ioutil.TempFile("", "output")
	defer os.Remove(file.Name())
	w := bufio.NewWriter(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = writeCSV(meta, data, w)
		require.NoError(b, err)
	}
}
