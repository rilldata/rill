package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const DefaultHeartbeat = 15 * time.Second

// Client represents a connected client.
type Client struct {
	id      string
	eventCh chan Event
	done    chan struct{}
}

// Event represents a server-sent event.
type Event struct {
	Type string
	ID   string
	Data interface{}
}

// EventServer represents a server that sends events to clients.
type EventServer struct {
	clients   map[string]*Client
	mu        sync.RWMutex
	heartbeat time.Duration
}

// Option is a function that configures the EventServer.
type Option func(*EventServer)

// New creates a new EventServer with the given options.
func New(options ...Option) *EventServer {
	es := &EventServer{
		clients:   make(map[string]*Client),
		heartbeat: DefaultHeartbeat,
	}

	for _, opt := range options {
		opt(es)
	}

	return es
}

// ServeHTTP handles incoming HTTP requests and serves server-sent events.
func (es *EventServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") != "" && r.Header.Get("Accept") != "*/*" && r.Header.Get("Accept") != "text/event-stream" {
		http.Error(w, "This endpoint requires Accept: text/event-stream", http.StatusNotAcceptable)
		return
	}

	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	respController := http.NewResponseController(w)

	client := &Client{
		id:      fmt.Sprintf("%d", time.Now().UnixNano()),
		eventCh: make(chan Event, 10),
		done:    make(chan struct{}),
	}

	es.mu.Lock()
	es.clients[client.id] = client
	es.mu.Unlock()

	defer func() {
		es.mu.Lock()
		defer es.mu.Unlock()
		delete(es.clients, client.id)
		close(client.eventCh)
	}()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		ticker := time.NewTicker(es.heartbeat)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				client.eventCh <- Event{
					Type: "heartbeat",
					Data: time.Now().Unix(),
				}
			case <-ctx.Done():
				return
			case <-client.done:
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-client.eventCh:
			if !ok {
				return
			}

			if err := es.writeEvent(w, event); err != nil {
				return
			}

			if err := respController.Flush(); err != nil {
				return
			}
		}
	}
}

// Publish sends an event to all connected clients.
func (es *EventServer) Publish(event Event) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, client := range es.clients {
		select {
		case client.eventCh <- event:
		default:
		}
	}
}

func (es *EventServer) writeEvent(w http.ResponseWriter, event Event) error {
	if event.Type != "" {
		fmt.Fprintf(w, "event: %s\n", event.Type)
	}

	if event.ID != "" {
		fmt.Fprintf(w, "id: %s\n", event.ID)
	}

	switch data := event.Data.(type) {
	case string:
		fmt.Fprintf(w, "data: %s\n", data)
	default:
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "data: %s\n", jsonData)
	}

	fmt.Fprint(w, "\n")
	return nil
}
