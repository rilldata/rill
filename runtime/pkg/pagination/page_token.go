package pagination

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

// MarshalPageToken encodes one or more args into a URL-safe base64 token.
func MarshalPageToken(args ...any) string {
	out := bytes.NewBuffer(nil)
	enc := json.NewEncoder(out)
	for _, arg := range args {
		if err := enc.Encode(arg); err != nil {
			panic(err)
		}
	}
	return base64.RawURLEncoding.EncodeToString(out.Bytes())
}

// UnmarshalPageToken decodes a token into the provided pointers in order.
func UnmarshalPageToken(token string, args ...any) error {
	res, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(res))
	for _, arg := range args {
		if err := dec.Decode(arg); err != nil {
			return err
		}
	}
	return nil
}

func ValidPageSize(pageSize uint32, defaultPageSize int) int {
	if pageSize == 0 {
		return defaultPageSize
	}
	return int(pageSize)
}
