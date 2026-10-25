package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	var last time.Time
	if clock != nil {
		last = clock.Now()
	}
	var rate float64
	if ratePerSec > 0 {
		rate = ratePerSec
	}
	return &Limiter{clock: clock, rate: rate, burst: burst, tokens: float64(burst), last: last}
}
func (l *Limiter) AllowN(n int) bool {
	if l.burst <= 0 || l.clock == nil || n <= 0 {
		return false
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock.Now()
	if timePassed := now.Sub(l.last); timePassed > 0 {
		l.tokens += l.rate * timePassed.Seconds()
		l.last = now

		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}
	}

	if float64(n) > l.tokens {
		return false
	}

	l.tokens -= float64(n)
	return true
}
