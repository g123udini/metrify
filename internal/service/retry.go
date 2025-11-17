package service

import (
	"fmt"
	"time"
)

func Retry(attempts int, initialDelay time.Duration, fn func() error) error {
	delay := initialDelay
	var err error

	for i := 1; i <= attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}

		time.Sleep(delay)
		delay *= 2
	}

	return fmt.Errorf(
		"after %d attempts, last error: %w",
		attempts, err,
	)
}
