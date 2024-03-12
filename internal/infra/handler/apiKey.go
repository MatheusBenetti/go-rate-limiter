package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/MatheusBenetti/go-rate-limiter/internal/dto"
	"github.com/MatheusBenetti/go-rate-limiter/internal/entity"
	"github.com/MatheusBenetti/go-rate-limiter/internal/usecase"
)

type APIKeyHandler struct {
	repo entity.APIKeyRepository
}

func NewAPIKeyHandler(repo entity.APIKeyRepository) *APIKeyHandler {
	return &APIKeyHandler{repo: repo}
}

func (at *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	input := dto.APIKeyInput{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("error decoding input data:", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	apiKeyUseCase := usecase.NewCreateAPIKeyUseCase(at.repo)
	result, execErr := apiKeyUseCase.Execute(r.Context(), input)
	if execErr != nil {
		log.Println("error decoding input data:", execErr.Error())
		http.Error(w, execErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
