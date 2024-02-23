package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// Storage define os métodos necessários para interagir com o mecanismo de persistência.
type Storage interface {
	IncrementRequests(ctx context.Context, key string) (int64, error)
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

// RedisStorage é uma implementação de Storage para o Redis.
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage cria uma nova instância de RedisStorage.
func NewRedisStorage(client *redis.Client) *RedisStorage {
	return &RedisStorage{client: client}
}

// IncrementRequests incrementa o contador de solicitações para a chave fornecida.
func (rs *RedisStorage) IncrementRequests(ctx context.Context, key string) (int64, error) {
	return rs.client.Incr(ctx, key).Result()
}

// Get obtém o valor associado à chave fornecida.
func (rs *RedisStorage) Get(ctx context.Context, key string) (string, error) {
	return rs.client.Get(ctx, key).Result()
}

// Set define o valor associado à chave fornecida com um tempo de expiração.
func (rs *RedisStorage) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rs.client.Set(ctx, key, value, expiration).Err()
}

// RateLimiter representa um rate limiter.
type RateLimiter struct {
	storage              Storage
	defaultLimit         int
	defaultExpiration    time.Duration
	maxRequestsPerSecond int
	blockDuration        time.Duration
}

// NewRateLimiter cria um novo rate limiter com o limite padrão e o intervalo de expiração padrão.
func NewRateLimiter(storage Storage, defaultLimit int, defaultExpiration, blockDuration time.Duration, maxRequestsPerSecond int) *RateLimiter {
	return &RateLimiter{
		storage:              storage,
		defaultLimit:         defaultLimit,
		defaultExpiration:    defaultExpiration,
		maxRequestsPerSecond: maxRequestsPerSecond,
		blockDuration:        blockDuration,
	}
}

// Allow verifica se uma nova solicitação do IP fornecido ou com o token de acesso fornecido é permitida.
func (rl *RateLimiter) Allow(ip, apiKey string) bool {
	ctx := context.Background()

	// Verifica se o IP está bloqueado.
	blocked, err := rl.storage.Get(ctx, "blocked:"+ip)
	if err == nil && blocked == "true" {
		return false
	}

	// Verifica se há um limite definido para o token de acesso.
	limit := rl.defaultLimit
	if apiKey != "" {
		tokenLimit, err := rl.storage.Get(ctx, "limit:"+apiKey)
		if err == nil {
			limit, _ = strconv.Atoi(tokenLimit)
		}
	}

	// Atualiza o contador de solicitações para o IP fornecido.
	reqCount, err := rl.storage.IncrementRequests(ctx, "requests:"+ip)
	if err != nil {
		fmt.Println("Erro ao incrementar contador de solicitações para o IP:", err)
		return false
	}

	// Verifica se o número de solicitações excede o limite por segundo.
	if reqCount > int64(rl.maxRequestsPerSecond) {
		rl.storage.Set(ctx, "blocked:"+ip, true, rl.blockDuration)
		return false
	}

	// Verifica se o número de solicitações excede o limite.
	if reqCount > int64(limit) {
		rl.storage.Set(ctx, "blocked:"+ip, true, rl.blockDuration)
		return false
	}

	return true
}

func main() {
	// Carrega variáveis de ambiente de um arquivo .env na pasta raiz.
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Erro ao carregar variáveis de ambiente:", err)
	}

	// Obtém configurações de limite do ambiente ou define valores padrão.
	defaultLimit := getIntEnv("DEFAULT_LIMIT", 3)
	defaultExpiration := getDurationEnv("DEFAULT_EXPIRATION", time.Minute)
	maxRequestsPerSecond := getIntEnv("MAX_REQUESTS_PER_SECOND", 5)
	blockDuration := getDurationEnv("BLOCK_DURATION", time.Minute)

	// Configuração do cliente Redis.
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Crie uma nova instância de RedisStorage.
	redisStorage := NewRedisStorage(redisClient)

	// Crie um novo rate limiter com os limites configurados e o armazenamento Redis.
	rateLimiter := NewRateLimiter(redisStorage, defaultLimit, defaultExpiration, blockDuration, maxRequestsPerSecond)

	// Crie um manipulador simples que apenas imprime uma mensagem.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Olá, mundo!")
	})

	// Use o rate limiter como middleware.
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		apiKey := r.Header.Get("API_KEY")

		if !rateLimiter.Allow(ip, apiKey) {
			http.Error(w, "Você atingiu o número máximo de requisições ou ações permitidas dentro de um determinado período de tempo", http.StatusTooManyRequests)
			return
		}

		handler.ServeHTTP(w, r)
	}))

	// Inicie o servidor na porta 8080.
	fmt.Println("Servidor escutando na porta 8080...")
	http.ListenAndServe(":8080", nil)
}

// getIntEnv retorna o valor de uma variável de ambiente como um int, ou um valor padrão se não estiver definido.
func getIntEnv(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		result, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("Erro ao analisar a variável de ambiente %s: %s\n", key, err)
			return defaultValue
		}
		return result
	}
	return defaultValue
}

// getDurationEnv retorna o valor de uma variável de ambiente como uma duração, ou um valor padrão se não estiver definido.
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		result, err := time.ParseDuration(value)
		if err != nil {
			fmt.Printf("Erro ao analisar a variável de ambiente %s: %s\n", key, err)
			return defaultValue
		}
		return result
	}
	return defaultValue
}
