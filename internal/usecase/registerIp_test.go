package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/MatheusBenetti/go-rate-limiter/config"
	"github.com/MatheusBenetti/go-rate-limiter/internal/dto"
	"github.com/MatheusBenetti/go-rate-limiter/internal/entity"
	"github.com/MatheusBenetti/go-rate-limiter/internal/entity/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterIPRequest_Execute(t *testing.T) {
	startTime := time.Date(2024, time.March, 12, 12, 34, 56, 0, time.UTC)
	cfg := &config.Config{
		RateLimiter: config.RateLimiter{
			ByIP: config.LimitValues{
				MaxRequests:     4,
				TimeWindow:      1,
				BlockedDuration: 600,
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIPRepo := mock.NewMockIPRepository(ctrl)
	ctx := context.Background()

	tests := []struct {
		name       string
		input      dto.IPRequestSave
		getRequest *entity.RateLimiter
		expected   dto.IPRequestResult
	}{
		{
			name: "with api key and result in getRequest",
			input: dto.IPRequestSave{
				IP:      "127.0.0.1",
				TimeAdd: startTime,
			},
			getRequest: func() *entity.RateLimiter {
				rt := &entity.RateLimiter{}
				rt.AddRequests(startTime.Add(-30 * time.Millisecond))
				rt.AddRequests(startTime.Add(-20 * time.Millisecond))
				rt.AddRequests(startTime.Add(-10 * time.Millisecond))
				return rt
			}(),
			expected: dto.IPRequestResult{
				Allow: true,
			},
		},
		{
			name: "with api key and result in getRequest",
			input: dto.IPRequestSave{
				IP:      "192.168.0.166",
				TimeAdd: startTime,
			},
			getRequest: func() *entity.RateLimiter {
				rt := &entity.RateLimiter{}
				rt.AddRequests(startTime.Add(-40 * time.Millisecond))
				rt.AddRequests(startTime.Add(-30 * time.Millisecond))
				rt.AddRequests(startTime.Add(-20 * time.Millisecond))
				rt.AddRequests(startTime.Add(-10 * time.Millisecond))
				return rt
			}(),
			expected: dto.IPRequestResult{
				Allow: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockIPRepo.EXPECT().GetBlockedDuration(ctx, tc.input.IP).Return("", nil).AnyTimes()
			mockIPRepo.EXPECT().GetRequest(ctx, tc.input.IP).Return(tc.getRequest, nil).AnyTimes()
			mockIPRepo.EXPECT().UpsertRequest(ctx, tc.input.IP, tc.getRequest).Return(nil).AnyTimes()
			mockIPRepo.EXPECT().
				SaveBlockedDuration(ctx, tc.input.IP, cfg.RateLimiter.ByIP.BlockedDuration).
				Return(nil).
				AnyTimes()
			rrq := NewRegisterIPUseCase(mockIPRepo, cfg)
			exRes, err := rrq.Execute(ctx, tc.input)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, exRes)
		})
	}
}
