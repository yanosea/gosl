package slack

import (
	"context"
	"math"
	"time"

	"github.com/slack-go/slack"
)

const (
	DefaultMaxRetries = 3
	DefaultMaxBackoff = 30 * time.Second
)

type RetryConfig struct {
	MaxRetries int
	MaxBackoff time.Duration
}

func withRetry[T any](ctx context.Context, cfg RetryConfig, fn func() (T, error)) (T, error) {
	var result T
	var err error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}

		if rateLimitErr, ok := err.(*slack.RateLimitedError); ok {
			sleepDuration := rateLimitErr.RetryAfter
			if sleepDuration > cfg.MaxBackoff {
				sleepDuration = cfg.MaxBackoff
			}

			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(sleepDuration):
				continue
			}
		}

		if attempt < cfg.MaxRetries {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			if backoff > cfg.MaxBackoff {
				backoff = cfg.MaxBackoff
			}

			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}
	}

	return result, err
}
