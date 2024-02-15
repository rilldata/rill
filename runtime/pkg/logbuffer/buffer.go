package logbuffer

import (
	"context"
	"sync"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/bufferutil"
	"golang.org/x/exp/slices"
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

func (b *Buffer) AddEntry(lvl runtimev1.LogLevel, t time.Time, msg, payload string) error {
	message := &runtimev1.Log{
		Level:       lvl,
		Time:        timestamppb.New(t),
		Message:     msg,
		JsonPayload: payload,
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.messages.Push(bufferutil.Item[*runtimev1.Log]{
		Value: message,
		Size:  len(payload) + len(msg) + 8 + 12, // enum size (assuming upper bound for 64 bits system) + proto Timestamp size
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
		slices.Reverse(logs)
	}

	return logs
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
