package limiter

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

// Limiter mantém o estado do rate limiter
type Limiter interface {
	Limit(ip, token string) bool
}

// Middleware define um middleware de rate limiting
type Middleware struct {
	limiter Limiter
}

// NewMiddleware cria um novo middleware de rate limiting
func NewMiddleware(limiter Limiter) *Middleware {
	return &Middleware{limiter}
}

// Handle aplica o rate limiting ao handler HTTP
func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		token := r.Header.Get("API_KEY")

		if m.limiter.Limit(ip, token) {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, "You have reached the maximum number of requests or actions allowed within a certain time frame")
		}
	})
}

// RedisLimiter implementa o Limiter usando Redis
type RedisLimiter struct {
	client *redis.Client
}

// NewRedisLimiter cria um novo Limiter que armazena informações no Redis
func NewRedisLimiter(client *redis.Client) (*RedisLimiter, error) {
	return &RedisLimiter{client}, nil
}

// Limit implementa a função Limit do Limiter usando Redis
func (rl *RedisLimiter) Limit(ip, token string) bool {
	// Lógica de limitação usando Redis
	return true // substitua por lógica real de limitação
}

func loadEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func loadIntEnv(key string, defaultValue int) int {
	valueStr := loadEnv(key, strconv.Itoa(defaultValue))
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
