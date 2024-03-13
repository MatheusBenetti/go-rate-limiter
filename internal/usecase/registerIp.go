package usecase

import (
	"context"
	"log"

	"github.com/MatheusBenetti/go-rate-limiter/config"
	"github.com/MatheusBenetti/go-rate-limiter/internal/dto"
	"github.com/MatheusBenetti/go-rate-limiter/internal/entity"
)

type RegisterIP struct {
	ipRepo entity.IPRepository
	config *config.Config
}

func NewRegisterIPUseCase(
	ipRepo entity.IPRepository,
	config *config.Config,
) *RegisterIP {
	return &RegisterIP{
		ipRepo: ipRepo,
		config: config,
	}
}

func (ipr *RegisterIP) Execute(
	ctx context.Context,
	input dto.IPRequestSave,
) (dto.IPRequestResult, error) {
	status, blockedErr := ipr.ipRepo.GetBlockedDuration(ctx, input.IP)
	if blockedErr != nil {
		return dto.IPRequestResult{}, blockedErr
	}

	if status == entity.StatusIPBlocked {
		log.Println("ip is blocked due to exceeding the maximum number of requests")
		return dto.IPRequestResult{}, entity.ErrIPExceededAmountRequest
	}

	getReq, getReqErr := ipr.ipRepo.GetRequest(ctx, input.IP)
	if getReqErr != nil {
		log.Printf("Error getting IP requests: %s \n", getReqErr.Error())
		return dto.IPRequestResult{}, getReqErr
	}

	getReq.TimeWindowSec = ipr.config.RateLimiter.ByIP.TimeWindow
	getReq.MaxRequests = ipr.config.RateLimiter.ByIP.MaxRequests
	if valErr := getReq.Validate(); valErr != nil {
		log.Printf("Error validation in rate limiter: %s \n", valErr.Error())
		return dto.IPRequestResult{}, valErr
	}

	getReq.AddRequests(input.TimeAdd)
	isAllowed := getReq.Allow(input.TimeAdd)
	if upsertErr := ipr.ipRepo.UpsertRequest(ctx, input.IP, getReq); upsertErr != nil {
		log.Printf("Error updating/inserting rate limit: %s \n", upsertErr.Error())
		return dto.IPRequestResult{}, upsertErr
	}

	if !isAllowed {
		if saveErr := ipr.ipRepo.SaveBlockedDuration(
			ctx,
			input.IP,
			ipr.config.RateLimiter.ByIP.BlockedDuration,
		); saveErr != nil {
			return dto.IPRequestResult{}, saveErr
		}
	}

	return dto.IPRequestResult{
		Allow: isAllowed,
	}, nil
}
