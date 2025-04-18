package retrier

import (
	"context"
	"database/sql/driver"
	"time"
)

type Retrier struct {
	backoff        []time.Duration
	additionalTest AdditionalTest
}

type AdditionalTest interface {
	IsHardFailure(context.Context) (bool, error)
}

func NewRetrier(maxRetries int, initialBackoff time.Duration, additionalTest AdditionalTest) *Retrier {
	return &Retrier{
		backoff:        exponentialBackoff(maxRetries, initialBackoff),
		additionalTest: additionalTest,
	}
}

func exponentialBackoff(maxRetries int, initialBackoff time.Duration) []time.Duration {
	ret := make([]time.Duration, maxRetries)
	next := initialBackoff
	for i := range ret {
		ret[i] = next
		next *= 2
	}
	return ret
}

type Action int

const (
	Succeed         Action = iota // Succeed indicates the Retrier should treat this value as a success.
	Fail                          // Fail indicates the Retrier should treat this value as a hard failure and not retry.
	Retry                         // Retry indicates the Retrier should treat this value as a soft failure and retry.
	AdditionalCheck               // Additional check is required to determine if it's Fail or Retry
)

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
			fallthrough
		case Retry:
			if retries >= len(r.backoff) {
				return nil, err
			}
			err := sleep(ctx, r.backoff[retries])
			if err != nil {
				return nil, err
			}
			retries++
		}
	}
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
