package entity

import (
	"log"
	"time"
)

type Limiter struct {
	Id            string
	Rate          int
	AccessCount   int
	BlockedAt     time.Time
	BlockDuration time.Duration
	CreatedAt     time.Time
}

func NewLimiter(id string, rate int, blockDuration time.Duration) *Limiter {
	return &Limiter{
		Id:            id,
		Rate:          rate,
		AccessCount:   0,
		BlockedAt:     time.Time{},
		BlockDuration: blockDuration,
		CreatedAt:     time.Now(),
	}
}

func (l *Limiter) IncrementAccessCount() error {
	if l.BlockedAt.IsZero() == false {
		if time.Since(l.BlockedAt) < l.BlockDuration {
			log.Printf("[Limiter] Access blocked for ID: %s, Blocked Until: %v", l.Id, l.BlockedAt.Add(l.BlockDuration))
			return NewIncrementBlockedError()
		}

		log.Printf("[Limiter] Limiter expired for ID: %s, resetting blocked status", l.Id)
		return NewExpiredLimiterError()
	}

	if time.Since(l.CreatedAt) > time.Second {
		log.Printf("[Limiter] Limiter expired for ID: %s, exceeded creation duration", l.Id)
		return NewExpiredLimiterError()
	}

	if (l.AccessCount + 1) > l.Rate {
		l.BlockedAt = time.Now()
		log.Printf("[Limiter] Rate exceeded for ID: %s, Blocking for duration: %v", l.Id, l.BlockDuration)
		return NewIncrementBlockedError()
	}
	l.AccessCount++

	log.Printf("[Limiter] Access count incremented for ID: %s, New Access Count: %d", l.Id, l.AccessCount)
	return nil
}
