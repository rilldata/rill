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

func (b *Buffer) AddZapEntry(entry zapcore.Entry, coreFields, entryFields []zapcore.Field) error {
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

func (b *Buffer) WatchLogs(ctx context.Context, fn LogCallback, minLvl runtimev1.LogLevel) error {
	messageChannel := make(chan *runtimev1.Log)
	b.addClient(messageChannel)
	defer b.removeClient(messageChannel)

	for {
		select {
		case message, open := <-messageChannel:
			if !open {
				panic("client closed!")
			}
			if message.Level < minLvl {
				continue
			}
			fn(message)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *Buffer) GetLogs(asc bool, limit int, minLvl runtimev1.LogLevel) []*runtimev1.Log {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if limit < 0 {
		limit = b.messages.Count()
	}
	limit = min(limit, b.messages.Count())

	logs := make([]*runtimev1.Log, limit)

	// always reverse iterate since we don't know how many logs will be skipped
	// so keep on going util we have enough logs. If we start from the beginning
	// we might iterate through entire buffer in worst case.
	b.messages.ReverseIterateUntil(func(item bufferutil.Item[*runtimev1.Log]) bool {
		if item.Value.Level < minLvl {
			// skip items having lower level than minLvl
			return true
		}
		// since default is asc=true fill the logs from the end so that we don't have to reverse it later
		logs[limit-1] = item.Value
		limit--
		return limit > 0
	})

	// truncate the logs from starting if it's not full in case some logs were skipped
	if limit > 0 {
		logs = logs[limit:]
	}

	if !asc {
		// this is a rare case, only when the user has specified asc=false
		// reverse the logs
		for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
			logs[i], logs[j] = logs[j], logs[i]
		}
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
