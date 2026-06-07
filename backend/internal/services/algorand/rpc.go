package algorand

import (
	"context"
	"strings"
	"time"
)

const (
	rpcRetryPause  = 65 * time.Second
	rpcMaxAttempts = 4
)

func isRateLimitErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "exceeded your limit")
}

func (c *Client) throttle(ctx context.Context) error {
	if c.minInterval <= 0 {
		return nil
	}
	c.rpcMu.Lock()
	defer c.rpcMu.Unlock()
	if !c.lastRPC.IsZero() {
		wait := c.minInterval - time.Since(c.lastRPC)
		if wait > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}
	}
	c.lastRPC = time.Now()
	return nil
}

func retryRPC[T any](ctx context.Context, c *Client, op func() (T, error)) (T, error) {
	var zero T
	var lastErr error
	for attempt := 0; attempt < rpcMaxAttempts; attempt++ {
		if err := c.throttle(ctx); err != nil {
			return zero, err
		}
		result, err := op()
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !isRateLimitErr(err) || attempt == rpcMaxAttempts-1 {
			return zero, err
		}
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(rpcRetryPause):
		}
	}
	return zero, lastErr
}
