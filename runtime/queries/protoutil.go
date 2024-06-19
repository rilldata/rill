package queries

import (
	"encoding/base64"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/gziputil"
	"google.golang.org/protobuf/proto"
)

func BakeQuery(qry *runtimev1.Query) (string, error) {
	if qry == nil {
		return "", errors.New("cannot bake nil query")
	}

	data, err := proto.Marshal(qry)
	if err != nil {
		return "", err
	}

	data, err = gziputil.GZipCompress(data)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(data), nil
}

func UnbakeQuery(bakedQry string) (*runtimev1.Query, error) {
	data, err := base64.URLEncoding.DecodeString(bakedQry)
	if err != nil {
		return nil, err
	}

	uncompressed, err := gziputil.GZipDecompress(data)
	if err != nil {
		// NOTE (2023-11-29): Backwards compatibility for when we didn't gzip baked queries. We can remove this in a few months.
		uncompressed = data
	}

	qry := &runtimev1.Query{}
	if err := proto.Unmarshal(uncompressed, qry); err != nil {
		return nil, err
	}

	return qry, nil
}

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
