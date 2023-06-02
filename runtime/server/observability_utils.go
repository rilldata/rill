package server

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func safeTimeStr(t *timestamppb.Timestamp) string {
	if t == nil {
		return ""
	}
	return t.AsTime().String()
}

func marshalProtoSlice[K proto.Message](s []K) []string {
	res := make([]string, len(s))
	for i := 0; i < len(res); i++ {
		res[i] = marshalProto(s[i])
	}
	return res
}

func marshalProto(pb proto.Message) string {
	b, err := protojson.Marshal(pb)
	if err != nil {
		return ""
	}
	return string(b)
}

func filterCount(m *runtimev1.MetricsViewFilter) int {
	if m == nil {
		return 0
	}
	return len(m.Include) + len(m.Exclude)
}
