package singleflight

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDo(t *testing.T) {
	var g Group
	v, err := g.Do(context.Background(), "key", func(context.Context) (interface{}, error) {
		return "bar", nil
	})
	if got, want := fmt.Sprintf("%v (%T)", v, v), "bar (string)"; got != want {
		t.Errorf("Do = %v; want %v", got, want)
	}
	if err != nil {
		t.Errorf("Do error = %v", err)
	}
}

func TestDoErr(t *testing.T) {
	var g Group
	someErr := errors.New("Some error")
	v, err := g.Do(context.Background(), "key", func(context.Context) (interface{}, error) {
		return nil, someErr
	})
	if err != someErr {
		t.Errorf("Do error = %v; want someErr %v", err, someErr)
	}
	if v != nil {
		t.Errorf("unexpected non-nil value %#v", v)
	}
}

func TestDoDupSuppress(t *testing.T) {
	var g Group
	var wg1, wg2 sync.WaitGroup
	c := make(chan string, 1)
	var calls int32
	fn := func(ctx context.Context) (interface{}, error) {
		if atomic.AddInt32(&calls, 1) == 1 {
			// First invocation.
			wg1.Done()
		}
		v := <-c
		c <- v // pump; make available for any future calls

		time.Sleep(10 * time.Millisecond) // let more goroutines enter Do

		return v, nil
	}

	const n = 10
	wg1.Add(1)
	for i := 0; i < n; i++ {
		wg1.Add(1)
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			wg1.Done()
			v, err := g.Do(context.Background(), "key", fn)
			if err != nil {
				t.Errorf("Do error: %v", err)
				return
			}
			if s, _ := v.(string); s != "bar" {
				t.Errorf("Do = %T %v; want %q", v, v, "bar")
			}
		}()
	}
	wg1.Wait()
	// At least one goroutine is in fn now and all of them have at
	// least reached the line before the Do.
	c <- "bar"
	wg2.Wait()
	if got := atomic.LoadInt32(&calls); got <= 0 || got >= n {
		t.Errorf("number of calls = %d; want over 0 and less than %d", got, n)
	}
}

// Test singleflight behaves correctly after Do panic.
// See https://github.com/golang/go/issues/41133
func TestPanicDo(t *testing.T) {
	var g Group
	fn := func(context.Context) (interface{}, error) {
		panic("invalid memory address or nil pointer dereference")
	}

	const n = 5
	waited := int32(n)
	panicCount := int32(0)
	done := make(chan struct{})
	for i := 0; i < n; i++ {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					// t.Logf("Got panic: %v\n%s", err, debug.Stack())
					atomic.AddInt32(&panicCount, 1)
				}

				if atomic.AddInt32(&waited, -1) == 0 {
					close(done)
				}
			}()

			g.Do(context.Background(), "key", fn)
		}()
	}

	select {
	case <-done:
		if panicCount != n {
			t.Errorf("Expect %d panic, but got %d", n, panicCount)
		}
	case <-time.After(time.Second):
		t.Fatalf("Do hangs")
	}
}

func TestGoexitDo(t *testing.T) {
	var g Group
	fn := func(ctx context.Context) (interface{}, error) {
		runtime.Goexit()
		return nil, nil
	}

	const n = 1
	waited := int32(n)
	done := make(chan struct{})
	for i := 0; i < n; i++ {
		go func() {
			var err error
			defer func() {
				if err != nil {
					t.Errorf("Error should be nil, but got: %v", err)
				}
				if atomic.AddInt32(&waited, -1) == 0 {
					close(done)
				}
			}()
			_, err = g.Do(context.Background(), "key", fn)
		}()
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("Do hangs")
	}
}

func TestContextCancelledForSome(t *testing.T) {
	s := Group{}

	f := func(ctx context.Context) (interface{}, error) {
		for {
			select {
			case <-ctx.Done():
				// Handle context cancellation
				return nil, ctx.Err()
			case <-time.After(200 * time.Millisecond):
				// Simulate some work
				return "hello", nil
			}
		}
	}
	errs := make([]error, 5)
	values := make([]interface{}, 5)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			var ctx context.Context
			var cancel context.CancelFunc
			if i%2 == 0 {
				// cancel all even requests
				ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
			} else {
				ctx, cancel = context.WithCancel(context.TODO())
			}
			defer cancel()
			defer wg.Done()
			values[i], errs[i] = s.Do(ctx, "key", func(ctx context.Context) (interface{}, error) {
				return f(ctx)
			})
		}(i)
		time.Sleep(10 * time.Millisecond) // ensure that first goroutine starts the work
	}
	wg.Wait()

	for i := 0; i < 5; i++ {
		if i%2 == 0 {
			require.Error(t, errs[i])
		} else {
			require.NoError(t, errs[i])
			require.Equal(t, values[i], "hello")
		}
	}
}

func TestContextCancelledForAll(t *testing.T) {
	s := Group{}

	f := func(ctx context.Context) (interface{}, error) {
		for {
			select {
			case <-ctx.Done():
				// Handle context cancellation
				return nil, ctx.Err()
			case <-time.After(200 * time.Millisecond):
				// Simulate some work
				return "hello", nil
			}
		}
	}
	errs := make([]error, 5)
	values := make([]interface{}, 5)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			defer wg.Done()
			values[i], errs[i] = s.Do(ctx, "key", func(ctx context.Context) (interface{}, error) {
				return f(ctx)
			})
		}(i)
	}
	wg.Wait()

	for i := 0; i < 5; i++ {
		require.Error(t, errs[i])
	}
}
