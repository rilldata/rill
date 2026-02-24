package mapstructureutil

import (
	"time"

	"github.com/mitchellh/mapstructure"
)

// WeakDecode is similar to mapstructure.WeakDecode, but it also supports decoding RFC3339Nano-formatted timestamp strings to time.Time.
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

// WeakDecodeWithWarnings is like WeakDecode but also returns any unused keys from the input.
func WeakDecodeWithWarnings(input, output any) ([]string, error) {
	md := &mapstructure.Metadata{}
	config := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
		Metadata:         md,
		Result:           output,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(input)
	if err != nil {
		return nil, err
	}

	return md.Unused, nil
}

// DecodeWithWarnings is like mapstructure.Decode but also returns any unused keys from the input.
func DecodeWithWarnings(input, output any) ([]string, error) {
	md := &mapstructure.Metadata{}
	config := &mapstructure.DecoderConfig{
		Metadata: md,
		Result:   output,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(input)
	if err != nil {
		return nil, err
	}

	return md.Unused, nil
}
