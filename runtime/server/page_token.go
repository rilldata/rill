package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

const _defaultPageSize = 100

func marshalPageToken(args ...any) string {
	out := bytes.NewBuffer(nil)

	enc := json.NewEncoder(out)
	for _, arg := range args {
		err := enc.Encode(arg)
		if err != nil {
			panic(err)
		}
	}

	return base64.RawURLEncoding.EncodeToString(out.Bytes())
}

func unmarshalPageToken(token string, args ...any) error {
	res, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(bytes.NewReader(res))
	for _, arg := range args {
		err = dec.Decode(arg)
		if err != nil {
			return err
		}
	}

	return nil
}

func validPageSize(pageSize uint32) int {
	if pageSize == 0 {
		return _defaultPageSize
	}
	return int(pageSize)
}
