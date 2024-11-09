package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// TestLimiterCreation ensures the limiter is created with the correct values.
func TestLimiterCreationWithValidValues(t *testing.T) {
	limiter := NewLimiter("unique_id_123", 10, 10*time.Minute)

	assert.Equal(t, "unique_id_123", limiter.Id)
	assert.Equal(t, 10, limiter.Rate)
	assert.Equal(t, 0, limiter.AccessCount)
	assert.True(t, limiter.BlockedAt.IsZero())
	assert.Equal(t, 10*time.Minute, limiter.BlockDuration)
	assert.WithinDuration(t, time.Now(), limiter.CreatedAt, time.Second)
}

// TestLimiterAccessCountWhenBelowRateLimit ensures no errors occur if access count is within rate limit.
func TestLimiterAccessCountWhenUnderRateLimitThenSuccess(t *testing.T) {
	limiter := NewLimiter("unique_id_456", 10, 10*time.Minute)

	for i := 0; i < 10; i++ {
		err := limiter.IncrementAccessCount()
		assert.NoError(t, err)
		assert.Equal(t, i+1, limiter.AccessCount)
	}
}

// TestLimiterAccessCountWhenRateLimitExceededBlocksAccess ensures the limiter blocks access after exceeding the rate limit.
func TestLimiterAccessCountWhenLimitExceededThenBlocksAccess(t *testing.T) {
	limiter := NewLimiter("unique_id_789", 10, 10*time.Minute)

	for i := 0; i < 10; i++ {
		_ = limiter.IncrementAccessCount()
	}

	err := limiter.IncrementAccessCount()
	assert.Error(t, err)
	assert.Equal(t, "is_blocked", err.(*LimiterError).Err)
	assert.Equal(t, 10, limiter.AccessCount)
	assert.False(t, limiter.BlockedAt.IsZero())
}

// TestLimiterAccessCountWhenBlockExpiresResetsLimit ensures that the limiter can reset after the block period.
func TestLimiterAccessCountWhenBlockExpiresThenResetsLimit(t *testing.T) {
	limiter := NewLimiter("unique_id_101", 2, 2*time.Second)

	err := limiter.IncrementAccessCount()
	assert.NoError(t, err)
	assert.Equal(t, 1, limiter.AccessCount)

	_ = limiter.IncrementAccessCount()
	time.Sleep(3 * time.Second)
	err = limiter.IncrementAccessCount()
	assert.Equal(t, "expired_limiter", err.(*LimiterError).Err)
}

// TestLimiterAccessCountWhenCreatedAtIsExpired ensures the limiter is expired if created time is too old.
func TestLimiterAccessCountWhenCreatedAtExpiresThenExpiredLimitError(t *testing.T) {
	limiter := NewLimiter("unique_id_202", 1, 5*time.Minute)

	limiter.CreatedAt = time.Now().Add(-3 * time.Second)

	err := limiter.IncrementAccessCount()
	assert.Error(t, err)
	assert.Equal(t, "expired_limiter", err.(*LimiterError).Err)
}
