package services

import (
	"fmt"
	"sync"
	"time"
)

type RateLimiterMemory struct {
	mu sync.Mutex
	limit int
	window time.Duration
	hits map[string][]time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiterMemory {
	return &RateLimiterMemory{limit: limit, window: window, hits: map[string][]time.Time{}}
}

func (r *RateLimiterMemory) Allow(key string) error {
	r.mu.Lock(); defer r.mu.Unlock()
	now := time.Now()
	arr := r.hits[key]
	// drop old
	var fresh []time.Time
	for _, t := range arr {
		if now.Sub(t) <= r.window { fresh = append(fresh, t) }
	}
	arr = fresh
	if len(arr) >= r.limit {
		return fmt.Errorf("rate limit exceeded: max %d in %s", r.limit, r.window)
	}
	r.hits[key] = append(arr, now)
	return nil
}
