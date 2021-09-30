package tcp

import (
	"go.uber.org/ratelimit"
	"time"
)

type RateLimiterFactory func() RateLimiter

type RateLimiter interface {
	Take() time.Time
}

func NewUberRateLimiter(rps int) RateLimiter {
	return ratelimit.New(rps)
}

func NewUnlimitedRateLimitFactory() RateLimiterFactory {
	return func() RateLimiter {
		return NewUnlimitedRateLimit()
	}
}

func NewUnlimitedRateLimit() RateLimiter {
	return ratelimit.NewUnlimited()
}
