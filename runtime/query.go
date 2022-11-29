package runtime

import "context"

type CachingQuery interface {
	Resolve(ctx context.Context, rt Runtime, instanceID string, priority int) error
	Key() string
	NamesQueried() []string
	Marshal() any
	Unmarshal(v any) error
}
