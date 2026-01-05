package pagination

import (
	"bytes"
	"context"
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

// CollectAll invokes pageFn repeatedly to retrieve all pages and returns a
// combined slice of results. Pagination stops when pageFn returns an empty
// nextToken. The pageFn callback must have the form:
//
//	pageFn(ctx, pageSize, token) → (items, nextToken, error)
func CollectAll[T any](ctx context.Context, pageFn func(context.Context, uint32, string) ([]T, string, error), pageSize uint32) ([]T, error) {
	var token string
	var out []T
	for {
		items, nextToken, err := pageFn(ctx, pageSize, token)
		if err != nil {
			return nil, err
		}

		if len(items) > 0 {
			out = append(out, items...)
		}

		if nextToken == "" {
			break
		}
		token = nextToken
	}
	return out, nil
}
