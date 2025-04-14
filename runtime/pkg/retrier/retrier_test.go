package retrier

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"
)

type mockAdditionalTest struct {
	isHardFailure bool
	err           error
}

func (m *mockAdditionalTest) IsHardFailure(ctx context.Context) (bool, error) {
	return m.isHardFailure, m.err
}

func TestRetrier_RunCtx(t *testing.T) {
	tests := []struct {
		name           string
		maxRetries     int
		initialBackoff time.Duration
		additionalTest AdditionalTest
		workResults    []struct {
			rows   driver.Rows
			action Action
			err    error
		}
		expectedError bool
	}{
		{
			name:           "succeed on first try",
			maxRetries:     3,
			initialBackoff: time.Millisecond,
			additionalTest: nil,
			workResults: []struct {
				rows   driver.Rows
				action Action
				err    error
			}{
				{rows: nil, action: Succeed, err: nil},
			},
			expectedError: false,
		},
		{
			name:           "fail on first try",
			maxRetries:     3,
			initialBackoff: time.Millisecond,
			additionalTest: nil,
			workResults: []struct {
				rows   driver.Rows
				action Action
				err    error
			}{
				{rows: nil, action: Fail, err: errors.New("hard failure")},
			},
			expectedError: true,
		},
		{
			name:           "retry and succeed",
			maxRetries:     3,
			initialBackoff: time.Millisecond,
			additionalTest: nil,
			workResults: []struct {
				rows   driver.Rows
				action Action
				err    error
			}{
				{rows: nil, action: Retry, err: errors.New("temporary error")},
				{rows: nil, action: Succeed, err: nil},
			},
			expectedError: false,
		},
		{
			name:           "retry and fail after max retries",
			maxRetries:     2,
			initialBackoff: time.Millisecond,
			additionalTest: nil,
			workResults: []struct {
				rows   driver.Rows
				action Action
				err    error
			}{
				{rows: nil, action: Retry, err: errors.New("temporary error 1")},
				{rows: nil, action: Retry, err: errors.New("temporary error 2")},
				{rows: nil, action: Retry, err: errors.New("temporary error 3")},
			},
			expectedError: true,
		},
		{
			name:           "additional check - hard failure",
			maxRetries:     3,
			initialBackoff: time.Millisecond,
			additionalTest: &mockAdditionalTest{isHardFailure: true, err: nil},
			workResults: []struct {
				rows   driver.Rows
				action Action
				err    error
			}{
				{rows: nil, action: AdditionalCheck, err: errors.New("ambiguous error")},
			},
			expectedError: true,
		},
		{
			name:           "additional check - retry and succeed",
			maxRetries:     3,
			initialBackoff: time.Millisecond,
			additionalTest: &mockAdditionalTest{isHardFailure: false, err: nil},
			workResults: []struct {
				rows   driver.Rows
				action Action
				err    error
			}{
				{rows: nil, action: AdditionalCheck, err: errors.New("ambiguous error")},
				{rows: nil, action: Succeed, err: nil},
			},
			expectedError: false,
		},
		{
			name:           "additional check - error in check",
			maxRetries:     3,
			initialBackoff: time.Millisecond,
			additionalTest: &mockAdditionalTest{isHardFailure: false, err: errors.New("check error")},
			workResults: []struct {
				rows   driver.Rows
				action Action
				err    error
			}{
				{rows: nil, action: AdditionalCheck, err: errors.New("ambiguous error")},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRetrier(tt.maxRetries, tt.initialBackoff, tt.additionalTest)

			callCount := 0
			work := func(ctx context.Context) (driver.Rows, Action, error) {
				if callCount >= len(tt.workResults) {
					t.Fatalf("work function called more times than expected: %d", callCount)
				}
				result := tt.workResults[callCount]
				callCount++
				return result.rows, result.action, result.err
			}

			_, err := r.RunCtx(context.Background(), work)

			if (err != nil) != tt.expectedError {
				t.Errorf("RunCtx() error = %v, expectedError %v", err, tt.expectedError)
			}

			expectedCalls := len(tt.workResults)
			if callCount != expectedCalls {
				t.Errorf("work function called %d times, expected %d", callCount, expectedCalls)
			}
		})
	}
}

func TestExponentialBackoff(t *testing.T) {
	initialBackoff := 10 * time.Millisecond
	maxRetries := 5

	backoffs := exponentialBackoff(maxRetries, initialBackoff)

	if len(backoffs) != maxRetries {
		t.Errorf("exponentialBackoff() returned %d durations, expected %d", len(backoffs), maxRetries)
	}

	expected := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		40 * time.Millisecond,
		80 * time.Millisecond,
		160 * time.Millisecond,
	}

	for i, b := range backoffs {
		if b != expected[i] {
			t.Errorf("backoff[%d] = %v, expected %v", i, b, expected[i])
		}
	}
}

func TestSleep(t *testing.T) {
	// Test normal sleep
	ctx := context.Background()
	start := time.Now()
	err := sleep(ctx, 10*time.Millisecond)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("sleep() error = %v, expected nil", err)
	}

	if elapsed < 10*time.Millisecond {
		t.Errorf("sleep() elapsed time %v, expected at least 10ms", elapsed)
	}

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	err = sleep(ctx, 100*time.Millisecond)
	if err != context.Canceled {
		t.Errorf("sleep() error = %v, expected context.Canceled", err)
	}
}
