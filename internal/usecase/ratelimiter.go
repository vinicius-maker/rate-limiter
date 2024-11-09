package usecase

import (
	"errors"
	"github.com/vinicius-maker/rate-limiter/internal/entity"
	"log"
	"sync"
	"time"
)

type RateLimiterDto struct {
	Id            string
	Rate          int
	BlockDuration time.Duration
}

type RateLimiterOutputDto struct {
	Error   error
	Success bool
}

type RateLimiter struct {
	limiterRepository entity.LimiterRepository
	mutexMap          map[string]*sync.Mutex
	mapAccessMutex    *sync.Mutex
}

func NewRateLimiter(limiterRepository entity.LimiterRepository) *RateLimiter {
	return &RateLimiter{
		limiterRepository: limiterRepository,
		mutexMap:          make(map[string]*sync.Mutex),
		mapAccessMutex:    &sync.Mutex{},
	}
}

func (rl *RateLimiter) Execute(dto RateLimiterDto) (RateLimiterOutputDto, error) {
	log.Printf("[RateLimiter] Start processing for ID: %s, Rate: %d, Block Duration: %s", dto.Id, dto.Rate, dto.BlockDuration)

	// Ensure mutex exists for the ID
	rl.mapAccessMutex.Lock()
	mutex, exists := rl.mutexMap[dto.Id]
	if !exists {
		mutex = &sync.Mutex{}
		rl.mutexMap[dto.Id] = mutex
	}
	rl.mapAccessMutex.Unlock()

	mutex.Lock()
	defer mutex.Unlock()

	limiterEntity, err := rl.limiterRepository.Find(dto.Id)
	if err != nil {
		var limiterErr *entity.LimiterError
		if errors.As(err, &limiterErr) && limiterErr.Err == "entity_not_found" {
			log.Printf("[RateLimiter] No limiter found for ID: %s. Creating a new limiter.", dto.Id)
			limiterEntity = entity.NewLimiter(dto.Id, dto.Rate, dto.BlockDuration)
			errOutput := limiterEntity.IncrementAccessCount()

			if err = rl.limiterRepository.Create(limiterEntity); err != nil {
				log.Printf("[RateLimiter] Error creating limiter for ID: %s: %v", dto.Id, err)
				return RateLimiterOutputDto{}, err
			}

			log.Printf("[RateLimiter] Limiter successfully created for ID: %s", dto.Id)
			return RateLimiterOutputDto{Error: errOutput, Success: true}, nil
		}

		log.Printf("[RateLimiter] Error fetching limiter for ID: %s: %v", dto.Id, err)
		return RateLimiterOutputDto{}, err
	}

	log.Printf("[RateLimiter] Limiter found for ID: %s. Incrementing access count.", dto.Id)
	incrementErr := limiterEntity.IncrementAccessCount()
	if incrementErr != nil {
		log.Printf("[RateLimiter] Error incrementing access count for ID: %s: %v", dto.Id, incrementErr)
		var limiterErr *entity.LimiterError
		if errors.As(incrementErr, &limiterErr) && limiterErr.Err == "expired_limiter" {
			log.Printf("[RateLimiter] Limiter expired for ID: %s. Recreating limiter.", dto.Id)
			limiterEntity = entity.NewLimiter(dto.Id, dto.Rate, dto.BlockDuration)
			incrementErr = limiterEntity.IncrementAccessCount()
			if err = rl.limiterRepository.Update(limiterEntity); err != nil {
				log.Printf("[RateLimiter] Error updating limiter for ID: %s after expiration: %v", dto.Id, err)
				return RateLimiterOutputDto{}, err
			}
			log.Printf("[RateLimiter] Limiter successfully updated after expiration for ID: %s", dto.Id)
			return RateLimiterOutputDto{Error: nil, Success: true}, nil
		}

		if err = rl.limiterRepository.Update(limiterEntity); err != nil {
			log.Printf("[RateLimiter] Error updating limiter for ID: %s: %v", dto.Id, err)
			return RateLimiterOutputDto{}, err
		}

		return RateLimiterOutputDto{Error: incrementErr, Success: false}, nil
	}

	if err = rl.limiterRepository.Update(limiterEntity); err != nil {
		log.Printf("[RateLimiter] Error updating limiter for ID: %s: %v", dto.Id, err)
		return RateLimiterOutputDto{}, err
	}

	log.Printf("[RateLimiter] Limiter successfully updated for ID: %s", dto.Id)
	return RateLimiterOutputDto{Error: nil, Success: true}, nil
}
