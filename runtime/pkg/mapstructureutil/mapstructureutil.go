package mapstructureutil

import (
	"time"

	"github.com/mitchellh/mapstructure"
)

func WeakDecode(input, output any) error {
	config := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
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
