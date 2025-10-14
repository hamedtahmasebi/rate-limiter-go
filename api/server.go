package api

import (
	"context"
	"log"
	"rate-limiter-go/limiter"
)

type Server struct {
	UnimplementedRateLimiterServer
	BucketStorage   limiter.BucketStorage
	ServiceRegistry limiter.ServiceRegistry
}

func (s *Server) GetAccessStatus(ctx context.Context, req *GetAccessStatusRequest) (*GetAccessStatusResponse, error) {
	log.Printf("checking access for client_id=%q service_id=%q", req.ClientID, req.ServiceID)
	service, err := s.ServiceRegistry.GetService(req.ServiceID)
	if err != nil {
		log.Printf("error getting service for service_id=%q: %s", req.ServiceID, err)
		return nil, err
	}
	accessRes, err := s.BucketStorage.ConsumeService(req.ClientID, service.ID)
	if err != nil {
		log.Printf("error consuming service for client_id=%q service_id=%q: %s", req.ClientID, req.ServiceID, err)
		return nil, err
	}
	log.Printf(
		"access status for client_id=%q service_id=%q: allowed=%t retry_after=%d",
		req.ClientID,
		req.ServiceID,
		accessRes.IsAllowed,
		accessRes.RetryAfterSeconds,
	)
	return &GetAccessStatusResponse{
		IsAllowed:         accessRes.IsAllowed,
		RetryAfterSeconds: accessRes.RetryAfterSeconds,
	}, nil
}
