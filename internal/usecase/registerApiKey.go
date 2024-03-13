package usecase

import (
	"context"
	"log"

	"github.com/MatheusBenetti/go-rate-limiter/internal/dto"
	"github.com/MatheusBenetti/go-rate-limiter/internal/entity"
)

type RegisterAPIKey struct {
	apiRepo entity.APIKeyRepository
}

func NewRegisterAPIKeyUseCase(
	apiRepo entity.APIKeyRepository,
) *RegisterAPIKey {
	return &RegisterAPIKey{
		apiRepo: apiRepo,
	}
}

func (apk *RegisterAPIKey) Execute(
	ctx context.Context,
	input dto.APIKeyRequestSave,
) (dto.APIKeyOutput, error) {
	status, blockedErr := apk.apiRepo.GetBlockedDuration(ctx, input.Value)
	if blockedErr != nil {
		return dto.APIKeyOutput{}, blockedErr
	}

	if status == entity.StatusAPIKeyBlocked {
		log.Println("API key is blocked due to exceeding the maximum number of requests")
		return dto.APIKeyOutput{}, entity.ErrAPIKeyExceededAmountRequest
	}

	apiKeyConfig, getErr := apk.apiRepo.Get(ctx, input.Value)
	if getErr != nil {
		log.Println("API key get error:", getErr.Error())
		return dto.APIKeyOutput{}, getErr
	}

	rateLimReq, getReqErr := apk.apiRepo.GetRequest(ctx, input.Value)
	if getReqErr != nil {
		log.Printf("Error getting IP requests: %s \n", getReqErr.Error())
		return dto.APIKeyOutput{}, getReqErr
	}

	rateLimReq.TimeWindowSec = apiKeyConfig.RateLimiter.TimeWindowSec
	rateLimReq.MaxRequests = apiKeyConfig.RateLimiter.MaxRequests
	if valErr := rateLimReq.Validate(); valErr != nil {
		log.Printf("Error validation in rate limiter: %s \n", valErr.Error())
		return dto.APIKeyOutput{}, valErr
	}

	rateLimReq.AddRequests(input.TimeAdd)
	isAllowed := rateLimReq.Allow(input.TimeAdd)
	if upsertErr := apk.apiRepo.UpsertRequest(ctx, input.Value, rateLimReq); upsertErr != nil {
		log.Printf("Error updating/inserting rate limit: %s \n", upsertErr.Error())
		return dto.APIKeyOutput{}, upsertErr
	}

	if !isAllowed {
		if saveErr := apk.apiRepo.SaveBlockedDuration(
			ctx,
			input.Value,
			apiKeyConfig.BlockedDuration,
		); saveErr != nil {
			return dto.APIKeyOutput{}, saveErr
		}
	}

	return dto.APIKeyOutput{
		Allow: isAllowed,
	}, nil
}
