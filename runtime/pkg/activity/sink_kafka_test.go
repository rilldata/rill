package activity

import (
	"errors"
	"testing"
	"time"
)

func Test_retry(t *testing.T) {
	alwaysFail := errors.New("always fail")
	conditionalFail := errors.New("conditional fail")
	var fail = true

	tests := []struct {
		name         string
		maxRetries   int
		fn           func() error
		retryOnErrFn func(err error) bool
		expectedErr  error
	}{
		{
			name:       "success without retry",
			maxRetries: 3,
			fn: func() error {
				return nil
			},
			retryOnErrFn: func(err error) bool {
				return true
			},
			expectedErr: nil,
		},
		{
			name:       "always fail",
			maxRetries: 3,
			fn: func() error {
				return alwaysFail
			},
			retryOnErrFn: func(err error) bool {
				return true
			},
			expectedErr: alwaysFail,
		},
		{
			name:       "retry conditionally",
			maxRetries: 3,
			fn: func() error {
				return conditionalFail
			},
			retryOnErrFn: func(err error) bool {
				if err == conditionalFail {
					return true
				}
				return false
			},
			expectedErr: conditionalFail,
		},
		{
			name:       "fail first time, success afterwards",
			maxRetries: 3,
			fn: func() error {
				if fail {
					fail = false
					return conditionalFail
				}
				return nil
			},
			retryOnErrFn: func(err error) bool {
				return err == conditionalFail
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := retry(t.Context(), tt.maxRetries, 1*time.Millisecond, tt.fn, tt.retryOnErrFn); err != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
