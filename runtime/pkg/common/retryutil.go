package common

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

const (
	MaxSleepMillis  = 3000
	BaseSleepMillis = 300
)

// Task is a function that returns an error if it fails
type Task func() error

// ShouldRetryFunc is a function that returns true if the error should be retried
type ShouldRetryFunc func(error) bool

// Retry executes the task with exponential backoff
func Retry(ctx context.Context, task Task, shouldRetry ShouldRetryFunc, maxTries int) error {
	if maxTries <= 0 {
		return errors.New("maxTries must be greater than 0")
	}
	var err error
	for try := 0; try < maxTries; try++ {
		if ctx.Err() != nil {
			return errors.Join(ctx.Err(), err)
		}

		err = task()
		if err == nil {
			return nil
		}

		if !shouldRetry(err) {
			return err
		}

		sleepMillis := nextRetrySleepMillis(try)

		time.Sleep(time.Duration(sleepMillis) * time.Millisecond)
	}
	return err
}

func nextRetrySleepMillis(try int) int64 {
	fuzzyMultiplier := math.Min(math.Max(1+0.2*rand.NormFloat64(), 0), 2)
	sleepMillis := int64(math.Min(float64(MaxSleepMillis), float64(BaseSleepMillis)*math.Pow(2, float64(try))*fuzzyMultiplier))
	return sleepMillis
}
