package mapstructureutil

import (
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/metricsview"
)

// WeakDecode is similar to mapstructure.WeakDecode, but it also supports decoding RFC3339Nano-formatted timestamp strings to time.Time.
func WeakDecode(input, output any) error {
	config := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(mapstructure.StringToTimeHookFunc(time.RFC3339Nano), timeRangeDecodeHook),
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func timeRangeDecodeHook(from, to reflect.Type, data any) (any, error) {
	if from == reflect.TypeOf(&metricsview.TimeRange{}) {
		tr := data.(*metricsview.TimeRange)

		// first decode to map normally, this will not handle time fields correctly
		trMap := map[string]any{}
		err := mapstructure.WeakDecode(tr, &trMap)
		if err != nil {
			return nil, err
		}

		// now set the time fields correctly
		trMap["start"] = tr.Start.Format(time.RFC3339Nano)
		trMap["end"] = tr.End.Format(time.RFC3339Nano)
		return trMap, nil
	}
	return data, nil
}
