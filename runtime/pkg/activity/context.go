package activity

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) context.Context {
	ctxAttrs := attrsFromContext(ctx)
	if ctxAttrs == nil {
		return context.WithValue(ctx, attrsContextKey{}, &attrs)
	}
	*ctxAttrs = append(*ctxAttrs, attrs...)
	return ctx
}

type attrsContextKey struct{}

func attrsFromContext(ctx context.Context) *[]attribute.KeyValue {
	v, ok := ctx.Value(attrsContextKey{}).(*[]attribute.KeyValue)
	if !ok {
		return nil
	}
	return v
}
