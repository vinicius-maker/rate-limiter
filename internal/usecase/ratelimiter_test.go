package usecase

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/vinicius-maker/rate-limiter/internal/entity"
	"github.com/vinicius-maker/rate-limiter/internal/infra"
	"log"
	"os"
	"testing"
	"time"
)

func setupRepo() entity.LimiterRepository {
	if err := godotenv.Load("./../../cmd/.env"); err != nil {
		log.Fatal("error trying to load env variables")
	}

	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDRESS"),
		DB:   0,
	})

	client.FlushDB(context.Background())

	return infra.NewRepositoryRedis(client)
}

// TestExecuteWhenLimiterIsNotFoundThenCreatesLimiter ensures the creation of a new limiter when it's not found in the repository.
func TestExecuteWhenLimiterNotFoundThenCreatesLimiter(t *testing.T) {
	repo := setupRepo()
	usecase := NewRateLimiter(repo)

	dto := RateLimiterDto{
		Id:            "new_id_123",
		Rate:          10,
		BlockDuration: 5 * time.Minute,
	}

	output, err := usecase.Execute(dto)
	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Nil(t, output.Error)

	limiter, err := repo.Find("new_id_123")
	assert.NoError(t, err)
	assert.Equal(t, 1, limiter.AccessCount)
}

// TestExecuteWhenLimiterExistsAndExpiredThenResetsLimiter ensures that an expired limiter gets reset.
func TestExecuteWhenLimiterExistsAndExpiredThenResetsLimiter(t *testing.T) {
	repo := setupRepo()
	usecase := NewRateLimiter(repo)

	dto := RateLimiterDto{
		Id:            "expired_id_456",
		Rate:          1,
		BlockDuration: 1 * time.Second,
	}

	limiter := entity.NewLimiter(dto.Id, dto.Rate, dto.BlockDuration)
	limiter.CreatedAt = time.Now().Add(-2 * time.Second)
	repo.Create(limiter)

	output, err := usecase.Execute(dto)
	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Nil(t, output.Error)

	limiter, err = repo.Find("expired_id_456")
	assert.NoError(t, err)
	assert.Equal(t, 1, limiter.AccessCount)
}

// TestExecuteWhenLimiterExistsAndNotExpiredThenIncrementsAccessCount ensures that a non-expired limiter increments its access count.
func TestExecuteWhenLimiterExistsAndNotExpiredThenIncrementsAccessCount(t *testing.T) {
	repo := setupRepo()
	usecase := NewRateLimiter(repo)

	dto := RateLimiterDto{
		Id:            "existing_id_789",
		Rate:          10,
		BlockDuration: 5 * time.Minute,
	}

	limiter := entity.NewLimiter(dto.Id, dto.Rate, dto.BlockDuration)
	repo.Create(limiter)

	output, err := usecase.Execute(dto)
	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Nil(t, output.Error)

	limiter, err = repo.Find("existing_id_789")
	assert.NoError(t, err)
	assert.Equal(t, 1, limiter.AccessCount)
}
