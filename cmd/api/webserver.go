package main

import (
	"net/http"

	"github.com/MatheusBenetti/go-rate-limiter/config"
	"github.com/MatheusBenetti/go-rate-limiter/internal/infra/database"
	internalHandler "github.com/MatheusBenetti/go-rate-limiter/internal/infra/handler"
	"github.com/MatheusBenetti/go-rate-limiter/internal/infra/webserver"
	"github.com/MatheusBenetti/go-rate-limiter/internal/infra/webserver/middleware"
	"github.com/redis/go-redis/v9"
)

func CreateWebServer(cfg *config.Config, redisCli *redis.Client) *webserver.WebServer {
	newWebServer := webserver.NewWebServer(cfg.App.Port)
	newWebServer.InternalMiddleware = middleware.Middleware{
		RedisClient: redisCli,
		Config:      cfg,
	}
	apikeyHandler := internalHandler.NewAPIKeyHandler(database.NewAPIKeyRedis(redisCli))

	newWebServer.AddHandler(http.MethodPost, "/api-key", apikeyHandler.CreateAPIKey)
	newWebServer.AddHandler(http.MethodGet, "/hello-world", internalHandler.HelloWorld)
	newWebServer.AddHandler(http.MethodGet, "/hello-world-key", internalHandler.HelloWorldWithAPIKey)

	return newWebServer
}
