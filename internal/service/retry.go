package service

import (
	"fmt"
	"time"
)

func Retry[T any](attempts int, base time.Duration, step time.Duration, fn func() (T, error)) (T, error) {
	var (
		res T
		err error
	)

	for i := 1; i <= attempts; i++ {
		res, err = fn()
		if err == nil {
			return res, nil
		}

		delay := base + step*time.Duration(i)

		time.Sleep(delay)
	}

	return res, fmt.Errorf("after %d attempts, last error: %w", attempts, err)
}
