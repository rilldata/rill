package mapstructureutil

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

func decodeTimeHook(from, to reflect.Type, data any) (any, error) {
	if to != reflect.TypeOf(time.Time{}) {
		return data, nil
	}

	switch from.Kind() {
	case reflect.String:
		str, ok := data.(string)
		if !ok {
			return data, nil
		}
		t, err := time.Parse(time.RFC3339Nano, str)
		if err != nil {
			return data, nil
		}
		return t, nil

	case reflect.Map:
		m, ok := data.(map[string]any)
		if !ok {
			return data, nil
		}
		if tStr, ok := m["t"].(string); ok {
			t, err := time.Parse(time.RFC3339Nano, tStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse time from map: %w", err)
			}
			return t, nil
		}
	}
	return data, nil
}

func WeakDecode(input, output any) error {
	config := &mapstructure.DecoderConfig{
		DecodeHook:       decodeTimeHook,
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
