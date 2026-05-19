package algoliautil

import (
	"context"
	"fmt"
	"time"
)

// Poll calls fn at increasing intervals until it returns done=true or maxAttempts
// is reached. Backoff doubles from 200ms up to a 5s cap. The context is respected.
// `name` is used only in the timeout error to identify the wait.
func Poll(ctx context.Context, name string, maxAttempts int, fn func() (done bool, err error)) error {
	delay := 200 * time.Millisecond
	for i := 0; i < maxAttempts; i++ {
		done, err := fn()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		if delay < 5*time.Second {
			delay *= 2
		}
		if delay > 5*time.Second {
			delay = 5 * time.Second
		}
	}
	return fmt.Errorf("%s did not complete after %d attempts", name, maxAttempts)
}
