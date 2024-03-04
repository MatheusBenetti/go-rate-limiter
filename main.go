package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/MatheusBenetti/go-rate-limiter/limiter"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar configurações do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Criar cliente Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // use default DB
	})

	// Inicializar o limiter
	rateLimiter, err := limiter.NewRedisLimiter(redisClient)
	if err != nil {
		log.Fatalf("Error initializing rate limiter: %v", err)
	}

	// Definir o middleware
	rateLimitMiddleware := limiter.NewMiddleware(rateLimiter)

	// Handler para a rota raiz com o middleware aplicado
	http.HandleFunc("/", rateLimitMiddleware.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Request allowed")
	}))

	// Iniciar o servidor na porta 8080
	fmt.Println("Server listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
