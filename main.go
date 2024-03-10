package main

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
)

var (
	apiKeyRateLimit = 10 // 10 requests per second per API key
	ipRateLimit     = 5  // 5 requests per second per IP
	apiKeyMap       = make(map[string]int)
	ipMap           = make(map[string]int)
	mutex           = sync.Mutex{}
)

func main() {
	r := chi.NewRouter()

	// Middleware to enforce rate limits
	r.Use(rateLimitMiddleware)

	// Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	http.ListenAndServe(":8080", r)
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from request header
		apiKey := r.Header.Get("API-Key")

		// Check API key rate limit
		if apiKey != "" {
			if !checkRateLimit(apiKeyMap, apiKey, apiKeyRateLimit) {
				http.Error(w, "API Key rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		} else {
			// If no API key, check IP rate limit
			ip := getClientIP(r)
			if !checkRateLimit(ipMap, ip, ipRateLimit) {
				http.Error(w, "IP rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func checkRateLimit(limitMap map[string]int, key string, limit int) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if count, ok := limitMap[key]; ok {
		if count >= limit {
			return false
		}
		limitMap[key]++
	} else {
		limitMap[key] = 1
	}

	// Schedule removal of key after 1 second
	go func() {
		time.Sleep(time.Second)
		mutex.Lock()
		defer mutex.Unlock()
		delete(limitMap, key)
	}()

	return true
}

func getClientIP(r *http.Request) string {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
