package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vinicius-maker/rate-limiter/internal/entity"
	"github.com/vinicius-maker/rate-limiter/internal/usecase"
)

type RateLimiterMiddleware struct {
	rateLimiter *usecase.RateLimiter
}

func NewRateLimiterMiddleware(rateLimiter *usecase.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{rateLimiter: rateLimiter}
}

func (m *RateLimiterMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientAPIKey := r.Header.Get("API_KEY")
			clientIP := strings.Split(r.RemoteAddr, ":")[0]

			id := clientIP
			rateLimit, err := strconv.Atoi(os.Getenv("RATE_LIMIT"))
			if err != nil {
				http.Error(w, "Invalid RATE_LIMIT environment variable value", http.StatusBadRequest)
				return
			}

			blockDurationMinutes, err := strconv.Atoi(os.Getenv("BLOCK_DURATION_TIME"))
			if err != nil {
				http.Error(w, "Invalid BLOCK_DURATION_TIME environment variable value", http.StatusBadRequest)
				return
			}

			// Override default rate and block duration if API key is provided
			if clientAPIKey != "" {
				id = clientAPIKey
				newRateLimit, newBlockDurationMinutes, err := parseApiKey(clientAPIKey)
				if err != nil {
					http.Error(w, "Invalid API key format", http.StatusBadRequest)
					return
				}

				rateLimit = newRateLimit
				blockDurationMinutes = newBlockDurationMinutes
			}

			dto := usecase.RateLimiterDto{
				Id:            id,
				Rate:          rateLimit,
				BlockDuration: time.Duration(blockDurationMinutes) * time.Minute,
			}

			log.Printf("[RateLimiterMiddleware] Processing request for ID: %s, Rate: %d, Block Duration: %d minutes", id, rateLimit, blockDurationMinutes)
			log.Println("[RateLimiterMiddleware] -------------------------------------------------------------------------------------------")

			output, err := m.rateLimiter.Execute(dto)

			log.Println("[RateLimiterMiddleware] -------------------------------------------------------------------------------------------")
			if err != nil {
				log.Printf("[RateLimiterMiddleware] Error executing rate limiter: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if output.Error != nil {
				var limiterErr *entity.LimiterError
				if errors.As(output.Error, &limiterErr) {
					log.Printf("[RateLimiterMiddleware] Rate limiter error for ID: %s, Error: %s", id, limiterErr.Err)
					switch limiterErr.Err {
					case "is_blocked":
						http.Error(w, limiterErr.Error(), http.StatusTooManyRequests)
						return
					case "entity_not_found", "expired_limiter":
						// Allow request to proceed for these cases.
						break
					default:
						http.Error(w, "Unexpected error occurred", http.StatusInternalServerError)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func parseApiKey(apiKey string) (int, int, error) {
	parts := strings.Split(apiKey, "_")
	if len(parts) < 5 {
		return 0, 0, fmt.Errorf("invalid API key format: expected at least 5 parts, got %d", len(parts))
	}

	rateLimit, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid rate value in API key: %v", err)
	}

	blockDurationMinutes, err := strconv.Atoi(parts[4])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid block duration value in API key: %v", err)
	}

	return rateLimit, blockDurationMinutes, nil
}
