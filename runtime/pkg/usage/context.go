package usage

import (
	"context"
)

type usageDimsContextKey struct{}

func ContextWithUsageDims(ctx context.Context, dims ...Dim) context.Context {
	return context.WithValue(ctx, usageDimsContextKey{}, &dims)
}

func GetDimsFromContext(ctx context.Context) *[]Dim {
	v, ok := ctx.Value(usageDimsContextKey{}).(*[]Dim)
	if !ok {
		return nil
	}
	return v
}

func AddDimsToContext(ctx context.Context, dims ...Dim) {
	ctxDims := GetDimsFromContext(ctx)
	if ctxDims != nil {
		*ctxDims = append(*ctxDims, dims...)
	}
}
