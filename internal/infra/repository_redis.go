package infra

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/vinicius-maker/rate-limiter/internal/entity"
	"log"
)

type RepositoryRedis struct {
	client *redis.Client
	ctx    context.Context
}

func NewRepositoryRedis(redisClient *redis.Client) *RepositoryRedis {
	return &RepositoryRedis{
		client: redisClient,
		ctx:    context.Background(),
	}
}

func (r *RepositoryRedis) Create(limiter *entity.Limiter) error {
	data, err := json.Marshal(limiter)
	if err != nil {
		log.Printf("[RepositoryRedis] Error marshalling limiter: %v", err)
		return err
	}

	err = r.client.Set(r.ctx, limiter.Id, data, 0).Err()
	if err != nil {
		log.Printf("[RepositoryRedis] Error setting data in Redis for ID: %s, Error: %v", limiter.Id, err)
	}
	return err
}

func (r *RepositoryRedis) Update(limiter *entity.Limiter) error {
	err := r.Create(limiter)
	if err != nil {
		log.Printf("[RepositoryRedis] Error updating limiter: %v", err)
		return err
	}
	return nil
}

func (r *RepositoryRedis) Find(id string) (*entity.Limiter, error) {
	data, err := r.client.Get(r.ctx, id).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			log.Printf("[RepositoryRedis] Limiter not found for ID: %s", id)
			return nil, entity.NewEntityNotFound()
		}

		log.Printf("[RepositoryRedis] Error retrieving data from Redis for ID: %s, Error: %v", id, err)
		return nil, err
	}

	var limiter entity.Limiter
	err = json.Unmarshal([]byte(data), &limiter)
	if err != nil {
		log.Printf("[RepositoryRedis] Error unmarshalling data for ID: %s, Error: %v", id, err)
		return nil, err
	}

	return &limiter, nil
}
