package mapstructureutil

import (
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

// WeakDecode is similar to mapstructure.WeakDecode, but it also supports decoding RFC3339Nano-formatted timestamp strings to time.Time.
func WeakDecode(input, output any) error {
	config := &mapstructure.DecoderConfig{
		DecodeHook:       stringToTimeHook,
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

func stringToTimeHook(from, to reflect.Type, data any) (any, error) {
	if to == reflect.TypeOf(time.Time{}) && from == reflect.TypeOf("") {
		return time.Parse(time.RFC3339Nano, data.(string))
	}

	return data, nil
}
