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
	log.Printf("level=info event=get_access_status service_id=%s client_id=%s user_id=%s usage_amount=%d", req.ServiceID, req.ClientID, req.UserID, req.UsageAmountReq)
	_, err := s.ServiceRegistry.GetService(req.ServiceID)
	if err != nil {
		log.Printf("level=error event=get_service_by_id status=error service_id=%q: error=%q", req.ServiceID, err)
		return nil, err
	}

	bucketID := limiter.GetBucketID(limiter.GetBucketIDRequest{
		ServiceID: req.ServiceID,
		ClientID:  req.ClientID,
		UserID:    req.UserID,
	})
	bucket, err := s.BucketStorage.GetBucket(bucketID)

	if bucket == nil {
		s.BucketStorage.CreateBucket(limiter.CreateBucketReqBody{
			ID:                  bucketID,
			InitialTokens:       100,
			RefillRatePerSecond: 1,
			MaxTokens:           100,
		})
	}

	accessRes, err := s.BucketStorage.ConsumeService(limiter.ConsumeServiceRequest{
		ServiceID:   req.ServiceID,
		ClientID:    req.ClientID,
		UserID:      req.UserID,
		UsageAmount: req.UsageAmountReq,
	})
	if err != nil {
		log.Printf("level=error event=consume_service status=error client_id=%q service_id=%q: error=%q", req.ClientID, req.ServiceID, err)
		return nil, err
	}
	log.Printf(
		"level=info event=get_access_status status=success client_id=%q service_id=%q: allowed=%t retry_after=%d",
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
