package pbutil

import (
	"encoding/binary"
	"fmt"
	"io"
	"slices"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/structpb"
)

// WriteHash writes the contents of a structpb.Value to a writer in a deterministic order.
// The output is not structured and can't be parsed back, so it's mainly suitable for writing to a hash writer.
func WriteHash(v *structpb.Value, w io.Writer) error {
	switch v2 := v.Kind.(type) {
	case *structpb.Value_NullValue:
		_, err := w.Write([]byte{0})
		return err
	case *structpb.Value_NumberValue:
		err := binary.Write(w, binary.BigEndian, v2.NumberValue)
		return err
	case *structpb.Value_StringValue:
		_, err := w.Write([]byte(v2.StringValue))
		return err
	case *structpb.Value_BoolValue:
		err := binary.Write(w, binary.BigEndian, v2.BoolValue)
		return err
	case *structpb.Value_ListValue:
		for _, v3 := range v2.ListValue.Values {
			err := WriteHash(v3, w)
			if err != nil {
				return err
			}
		}
	case *structpb.Value_StructValue:
		// Iterate over sorted keys
		ks := maps.Keys(v2.StructValue.Fields)
		slices.Sort(ks)
		for _, k := range ks {
			_, err := w.Write([]byte(k))
			if err != nil {
				return err
			}
			err = WriteHash(v2.StructValue.Fields[k], w)
			if err != nil {
				return err
			}
		}
	default:
		panic(fmt.Sprintf("unknown kind %T", v.Kind))
	}
	return nil
}
