package publisher

import (
	"context"
)

type usageDimsContextKey struct{}

func WithDims(ctx context.Context, dims ...Dim) context.Context {
	ctxDims := GetDimsFromContext(ctx)
	if ctxDims == nil {
		return context.WithValue(ctx, usageDimsContextKey{}, &dims)
	}
	*ctxDims = append(*ctxDims, dims...)
	return ctx
}

func GetDimsFromContext(ctx context.Context) *[]Dim {
	v, ok := ctx.Value(usageDimsContextKey{}).(*[]Dim)
	if !ok {
		return nil
	}
	return v
}
