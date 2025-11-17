package service

import (
	"fmt"
	"time"
)

func Retry[T any](attempts int, initialDelay time.Duration, fn func() (T, error)) (T, error) {
	var (
		delay = initialDelay
		res   T
		err   error
	)

	for i := 1; i <= attempts; i++ {
		res, err = fn()
		if err == nil {
			return res, nil
		}

		time.Sleep(delay)
		delay *= 2
	}

	return res, fmt.Errorf("after %d attempts, last error: %w", attempts, err)
}
