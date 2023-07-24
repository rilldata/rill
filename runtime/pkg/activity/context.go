package activity

import (
	"context"
)

type activityContextKey struct{}

type activityInfo struct {
	dims              *[]Dim
	requestInstanceID string
}

func WithDims(ctx context.Context, dims ...Dim) context.Context {
	info := getInfoFromContext(ctx)
	if info == nil {
		info = &activityInfo{}
		ctx = context.WithValue(ctx, activityContextKey{}, info)
	}
	if info.dims == nil {
		info.dims = &[]Dim{}
	}
	*info.dims = append(*info.dims, dims...)
	return ctx
}

func GetDimsFromContext(ctx context.Context) *[]Dim {
	info := getInfoFromContext(ctx)
	if info == nil {
		return nil
	}
	return info.dims
}

func WithRequestInstanceID(ctx context.Context, instanceID string) context.Context {
	info := getInfoFromContext(ctx)
	if info == nil {
		info = &activityInfo{}
		ctx = context.WithValue(ctx, activityContextKey{}, info)
	}
	info.requestInstanceID = instanceID
	return ctx
}

func GetRequestInstanceIDFromContext(ctx context.Context) string {
	info := getInfoFromContext(ctx)
	if info == nil {
		return ""
	}
	return info.requestInstanceID
}

func getInfoFromContext(ctx context.Context) *activityInfo {
	v, ok := ctx.Value(activityContextKey{}).(*activityInfo)
	if !ok {
		return nil
	}
	return v
}
