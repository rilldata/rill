package mapstructureutil

import (
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

func WeakDecode(input, output any) error {
	config := &mapstructure.DecoderConfig{
		DecodeHook:       timeDecodeHook(),
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

// timeDecodeHook handles decoding of time values
func timeDecodeHook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
		func(from reflect.Type, to reflect.Type, data any) (any, error) {
			// Handle map format {"t": "timestamp"} → time.Time
			if to == reflect.TypeOf(time.Time{}) && from.Kind() == reflect.Map {
				m, ok := data.(map[string]any)
				if !ok {
					return data, nil
				}
				if tStr, ok := m["t"].(string); ok {
					return time.Parse(time.RFC3339Nano, tStr)
				}
			}
			return data, nil
		},
	)
}
