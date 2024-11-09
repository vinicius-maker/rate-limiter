package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/vinicius-maker/rate-limiter/internal/infra"
	"github.com/vinicius-maker/rate-limiter/internal/middleware"
	"github.com/vinicius-maker/rate-limiter/internal/usecase"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := godotenv.Load("./cmd/.env"); err != nil {
		log.Fatal("error loading env variables")
		return
	}

	chiRouter := chi.NewRouter()

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDRESS"),
	})

	repoRedis := infra.NewRepositoryRedis(redisClient)

	rateLimiterUseCase := usecase.NewRateLimiter(repoRedis)
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(rateLimiterUseCase)

	chiRouter.Use(rateLimiterMiddleware.Handler())

	chiRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("access granted"))
	})

	log.Fatal(http.ListenAndServe(":8080", chiRouter))
}
