package activity

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

type usageDimsContextKey struct{}

func WithDims(ctx context.Context, dims ...attribute.KeyValue) context.Context {
	ctxDims := GetDimsFromContext(ctx)
	if ctxDims == nil {
		return context.WithValue(ctx, usageDimsContextKey{}, &dims)
	}
	*ctxDims = append(*ctxDims, dims...)
	return ctx
}

func GetDimsFromContext(ctx context.Context) *[]attribute.KeyValue {
	v, ok := ctx.Value(usageDimsContextKey{}).(*[]attribute.KeyValue)
	if !ok {
		return nil
	}
	return v
}
