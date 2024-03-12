package entity

import (
	"sync"
	"time"
)

type RateLimiter struct {
	Requests      []time.Time
	TimeWindowSec int64
	MaxRequests   int
	lock          sync.Mutex
}

func (rl *RateLimiter) Allow(fromTime time.Time) bool {
	rl.lock.Lock()
	defer rl.lock.Unlock()

	rl.removeOldRequests(fromTime)
	return len(rl.Requests) <= rl.MaxRequests
}

func (rl *RateLimiter) GetDurationTimeWindow() time.Duration {
	return time.Duration(rl.TimeWindowSec) * time.Second
}

func (rl *RateLimiter) removeOldRequests(fromTime time.Time) {
	threshold := fromTime.Add(-rl.GetDurationTimeWindow())
	start := 0
	for i, t := range rl.Requests {
		if t.After(threshold) {
			start = i
			break
		}
	}
	rl.Requests = rl.Requests[start:]
}

func (rl *RateLimiter) AddRequests(request time.Time) {
	rl.Requests = append(rl.Requests, request)
}

func (rl *RateLimiter) Validate() error {
	if rl.MaxRequests == 0 {
		return ErrRateLimiterMaxRequests
	}

	if rl.TimeWindowSec == 0 {
		return ErrRateLimiterTimeWindow
	}

	return nil
}
