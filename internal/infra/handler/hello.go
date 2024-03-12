package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MatheusBenetti/go-rate-limiter/internal/entity"
)

func HelloWorld(w http.ResponseWriter, _ *http.Request) {
	output := struct {
		Message string `json:"message"`
	}{
		Message: "Hello world!",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(output); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HelloWorldWithAPIKey(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get(entity.APIKeyHeaderName)
	if apiKey == "" {
		http.Error(w, errors.New("you need an API key to perform this request").Error(), http.StatusBadRequest)
		return
	}

	output := struct {
		Message string `json:"message"`
		APIKey  string `json:"api_key"`
	}{
		Message: "Hello world with API key!",
		APIKey:  apiKey,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(output); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
