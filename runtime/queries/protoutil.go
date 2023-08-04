package queries

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/proto"
)

// safeFieldType returns the type of the field at index i, or nil if no type is found.
func safeFieldType(t *runtimev1.StructType, i int) *runtimev1.Type {
	if t != nil && len(t.Fields) > i {
		return t.Fields[i].Type
	}
	return nil
}

// sizeProtoMessage returns size of serialized proto message
func sizeProtoMessage(m proto.Message) int64 {
	bytes, err := proto.Marshal(m)
	if err != nil {
		panic(err)
	}

	return int64(len(bytes))
}
