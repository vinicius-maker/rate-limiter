package infra

import (
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/vinicius-maker/rate-limiter/internal/entity"
	"log"
	"os"
	"testing"
	"time"
)

// TestCreateLimiter ensures that the limiter is correctly created and stored in Redis.
func TestCreateLimiterSuccessfully(t *testing.T) {
	if err := godotenv.Load("./../../cmd/.env"); err != nil {
		log.Fatal("error trying to load env variables")
		return
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDRESS"),
	})

	repo := NewRepositoryRedis(rdb)

	limiter := entity.NewLimiter("unique_id_123", 10, 10*time.Minute)

	err := repo.Create(limiter)

	assert.NoError(t, err)

	result, err := repo.Find(limiter.Id)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, limiter.Id, result.Id)
	assert.Equal(t, limiter.Rate, result.Rate)
	assert.Equal(t, limiter.AccessCount, result.AccessCount)
	assert.Equal(t, limiter.BlockedAt, result.BlockedAt)
	assert.Equal(t, limiter.BlockDuration, result.BlockDuration)
	assert.WithinDuration(t, time.Now(), result.CreatedAt, time.Second)
}

// TestUpdateLimiter ensures that an existing limiter can be updated and retrieved successfully.
func TestUpdateLimiterSuccessfully(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDRESS"),
	})

	repo := NewRepositoryRedis(rdb)

	limiter := entity.NewLimiter("unique_id_456", 10, 10*time.Minute)

	err := repo.Create(limiter)

	err2 := limiter.IncrementAccessCount()

	err3 := repo.Update(limiter)

	assert.NoError(t, err)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	result, err := repo.Find(limiter.Id)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, limiter.Id, result.Id)
	assert.Equal(t, limiter.Rate, result.Rate)
	assert.Equal(t, limiter.AccessCount, result.AccessCount)
	assert.Equal(t, limiter.BlockedAt, result.BlockedAt)
	assert.Equal(t, limiter.BlockDuration, result.BlockDuration)
	assert.WithinDuration(t, time.Now(), result.CreatedAt, time.Second)
}
