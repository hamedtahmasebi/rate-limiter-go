package api

import (
	"context"
	"log"
	"rate-limiter-go/limiter"
)

type Server struct {
	UnimplementedRateLimiterServer
	Limiter         limiter.RateLimiter
	ServiceRegistry limiter.ServiceRegistry
}

func (s *Server) GetAccessStatus(ctx context.Context, req *GetAccessStatusRequest) (*GetAccessStatusResponse, error) {
	log.Printf("level=info event=get_access_status service_id=%s client_id=%s user_id=%s usage_amount=%d", req.ServiceID, req.ClientID, req.UserID, req.UsageAmountReq)
	_, err := s.ServiceRegistry.GetService(req.ServiceID)
	if err != nil {
		log.Printf("level=error event=get_service_by_id status=error service_id=%q: error=%q", req.ServiceID, err)
		return nil, err
	}

	res, err := s.Limiter.ConsumeService(limiter.RateLimitingReqeustContext{
		ServiceID:   req.ServiceID,
		ClientID:    req.ClientID,
		UserID:      req.UserID,
		UsageAmount: req.UsageAmountReq,
	})

	if err != nil {
		return nil, err
	}
	return &GetAccessStatusResponse{
		IsAllowed:         res.IsAllowed,
		RetryAfterSeconds: res.RetryAfterSeconds,
	}, nil
}
