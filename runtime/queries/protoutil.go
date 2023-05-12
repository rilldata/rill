package queries

import "google.golang.org/protobuf/proto"

// sizeProtoMessage returns approx size of proto message
func sizeProtoMessage(m proto.Message) int64 {
	bytes, err := proto.Marshal(m)
	if err != nil {
		panic(err)
	}

	return int64(len(bytes))
}
