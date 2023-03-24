package metrics

import (
	"context"

	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"google.golang.org/grpc"
)

type contextKey int

const OtelCfgKey contextKey = 0

type OtelCfg struct {
	CacheHits   instrument.Int64Counter
	CacheMisses instrument.Int64Counter
}

func customMetricsUnaryInterceptor(otelCfg *OtelCfg) (grpc.UnaryServerInterceptor, error) {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx = context.WithValue(ctx, OtelCfgKey, otelCfg)
		return handler(ctx, req)
	}, nil
}

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *serverStream) Context() context.Context {
	return ss.ctx
}

func wrapServerStream(ctx context.Context, ss grpc.ServerStream) *serverStream {
	return &serverStream{
		ServerStream: ss,
		ctx:          ctx,
	}
}

func customMetricsStreamInterceptor(otelCfg *OtelCfg) (grpc.StreamServerInterceptor, error) {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := context.WithValue(ss.Context(), OtelCfgKey, otelCfg)

		return handler(srv, wrapServerStream(ctx, ss))
	}, nil
}

func InitCustomMetricsInterceptors(ctx context.Context) (grpc.StreamServerInterceptor, grpc.UnaryServerInterceptor, error) {
	mp := global.Meter("cache")
	hits, err := mp.Int64Counter("hits")
	if err != nil {
		return nil, nil, err
	}

	misses, err := mp.Int64Counter("misses")
	if err != nil {
		return nil, nil, err
	}

	otelCfg := &OtelCfg{
		CacheHits:   hits,
		CacheMisses: misses,
	}

	si, err := customMetricsStreamInterceptor(otelCfg)
	if err != nil {
		return nil, nil, err
	}
	ui, err := customMetricsUnaryInterceptor(otelCfg)
	if err != nil {
		return nil, nil, err
	}

	return si, ui, nil
}
