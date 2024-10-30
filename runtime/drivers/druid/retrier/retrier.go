package retrier

import (
	"context"
	"database/sql/driver"
	"math/rand"
	"time"
)

type Retrier struct {
	backoff        []time.Duration
	rand           *rand.Rand
	additionalTest AdditionalTest
}

type AdditionalTest interface {
	IsHardFailure(context.Context) (bool, error)
}

func NewRetrier(n int, initial time.Duration, at AdditionalTest) *Retrier {
	return &Retrier{
		// nolint:gosec // don't need it
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
		backoff: ExponentialBackoff(n, initial),
	}
}

func (r *Retrier) sleep(ctx context.Context, t <-chan time.Time) error {
	select {
	case <-t:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Retrier) RunCtx(ctx context.Context, work func(ctx context.Context) (driver.Rows, Action, error)) (driver.Rows, error) {
	retries := 0
	for {
		res, a, err := work(ctx)

		switch a {
		case Succeed:
			return res, nil
		case Fail:
			return nil, err
		case AdditionalCheck:
			if r.additionalTest != nil {
				/*
					For example, Druid datasource-unavailable error can have 2 cases:
						a) no datasource (should hard fail)
						b) coordinator is down (should retry)
					In both cases the error type is ambiguous - `invalidInput`.
					This aditional test makes sure if coordinator is OK (meaning the initial request had an invalid input).
				*/
				if ok, err2 := r.additionalTest.IsHardFailure(ctx); ok || err2 != nil {
					if err2 != nil {
						return nil, err2
					}
					return nil, err
				}
			}

			if retries >= len(r.backoff) {
				return nil, err
			}

			dur := r.backoff[retries]
			timeout := time.After(dur)
			if err := r.sleep(ctx, timeout); err != nil {
				return nil, err
			}

			retries++
		case Retry:
			if retries >= len(r.backoff) {
				return nil, err
			}

			dur := r.backoff[retries]
			timeout := time.After(dur)

			if err := r.sleep(ctx, timeout); err != nil {
				return nil, err
			}

			retries++
		}
	}
}

type Action int

const (
	Succeed         Action = iota // Succeed indicates the Retrier should treat this value as a success.
	Fail                          // Fail indicates the Retrier should treat this value as a hard failure and not retry.
	Retry                         // Retry indicates the Retrier should treat this value as a soft failure and retry.
	AdditionalCheck               // Additional check is required to determine if it's Fail or Retry
)

func ExponentialBackoff(n int, initialAmount time.Duration) []time.Duration {
	ret := make([]time.Duration, n)
	next := initialAmount
	for i := range ret {
		ret[i] = next
		next *= 2
	}
	return ret
}
