package logbuffer

import (
	"context"
	"encoding/json"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/bufferutil"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Buffer struct {
	clients  map[chan *runtimev1.Log]struct{}
	mu       sync.RWMutex
	messages *bufferutil.BoundedCircularBuffer[*runtimev1.Log]
}

type LogCallback func(item *runtimev1.Log)

func NewBuffer(maxMessageCount int, maxBufferSize int64) *Buffer {
	return &Buffer{
		clients:  make(map[chan *runtimev1.Log]struct{}),
		mu:       sync.RWMutex{},
		messages: bufferutil.NewBoundedCircularBuffer[*runtimev1.Log](maxMessageCount, maxBufferSize),
	}
}

func (b *Buffer) AddZapEntry(entry zapcore.Entry, coreFields []zapcore.Field, entryFields []zapcore.Field) error {
	size := 0
	attrs := make(map[string]any, len(coreFields)+len(entryFields))
	for _, field := range coreFields {
		attrs[field.Key] = field.String
		size += len(field.Key) + len(field.String) // approx size
	}

	for _, field := range entryFields {
		attrs[field.Key] = field.String
		size += len(field.Key) + len(field.String) // approx size
	}

	size += len(entry.Message)
	// add enum size, assuming upper bound for 64 bits system
	size += 8
	// add proto Timestamp size
	size += 12

	payload, err := json.Marshal(attrs)
	if err != nil {
		return err
	}

	message := &runtimev1.Log{
		Level:       zapLevelToPBLevel(entry.Level),
		Time:        timestamppb.New(entry.Time),
		Message:     entry.Message,
		JsonPayload: string(payload),
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.messages.Push(bufferutil.Item[*runtimev1.Log]{
		Value: message,
		Size:  size,
	})

	for client := range b.clients {
		client <- message
	}
	return nil
}

func (b *Buffer) WatchLogs(ctx context.Context, fn LogCallback) error {
	messageChannel := make(chan *runtimev1.Log)
	b.addClient(messageChannel)
	defer b.removeClient(messageChannel)

	for {
		select {
		case message, open := <-messageChannel:
			if !open {
				panic("client closed!")
			}
			fn(message)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *Buffer) GetLogs(asc bool, limit int) []*runtimev1.Log {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if limit < 0 {
		limit = b.messages.Count()
	}
	limit = min(limit, b.messages.Count())

	logs := make([]*runtimev1.Log, limit)
	i := 0
	if asc {
		b.messages.Iterate(func(item bufferutil.Item[*runtimev1.Log]) {
			logs[i] = item.Value
			i++
		}, limit)
	} else {
		b.messages.ReverseIterate(func(item bufferutil.Item[*runtimev1.Log]) {
			logs[i] = item.Value
			i++
		}, limit)
	}

	return logs
}

func zapLevelToPBLevel(level zapcore.Level) runtimev1.LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return runtimev1.LogLevel_LOG_LEVEL_DEBUG
	case zapcore.InfoLevel:
		return runtimev1.LogLevel_LOG_LEVEL_INFO
	case zapcore.WarnLevel:
		return runtimev1.LogLevel_LOG_LEVEL_WARN
	case zapcore.ErrorLevel:
		return runtimev1.LogLevel_LOG_LEVEL_ERROR
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return runtimev1.LogLevel_LOG_LEVEL_FATAL
	default:
		return runtimev1.LogLevel_LOG_LEVEL_UNSPECIFIED
	}
}

func (b *Buffer) addClient(client chan *runtimev1.Log) {
	b.mu.Lock()
	b.clients[client] = struct{}{}
	b.mu.Unlock()
}

func (b *Buffer) removeClient(client chan *runtimev1.Log) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, client)
	close(client) // close the channel here after the client is removed.
}
