package limiter

import (
	"log"
	"time"
)

type CreateBucketReqBody struct {
	ClientID            string
	ServiceID           string
	InitialTokens       uint64
	RefillRatePerSecond uint64
}

type Bucket struct {
	clientID            string
	serviceID           string
	tokens              uint64
	refillRatePerSecond uint64
	CreatedAt           time.Time
	LastRefill          time.Time
}

type AccessStatusResponse struct {
	IsAllowed         bool
	RetryAfterSeconds uint64
}

type BucketStorage interface {
	CreateBucket(body CreateBucketReqBody) error
	ConsumeService(clientID string, serviceID string, usageAmount uint64) (AccessStatusResponse, error)
	// ConsumeTokens(clientID string, serviceID string, amount uint)
}

type BucketStorageImpl struct {
	BucketsMap      map[string]map[string]*Bucket
	ServiceRegistry ServiceRegistry
}

func (bs *BucketStorageImpl) CreateBucket(body CreateBucketReqBody) error {
	log.Printf("event=create_bucket client_id=%q service_id=%q initial_tokens=%d refill_rate_per_second=%d", body.ClientID, body.ServiceID, body.InitialTokens, body.RefillRatePerSecond)
	clientServices, csExists := bs.BucketsMap[body.ClientID]
	if !csExists {
		clientServices = make(map[string]*Bucket)
		bs.BucketsMap[body.ClientID] = clientServices
	}
	clientServices[body.ServiceID] = &Bucket{
		clientID:            body.ClientID,
		serviceID:           body.ServiceID,
		tokens:              body.InitialTokens,
		refillRatePerSecond: body.RefillRatePerSecond,
		CreatedAt:           time.Now(),
		LastRefill:          time.Now(),
	}
	log.Printf("event=bucket_created client_id=%q service_id=%q", body.ClientID, body.ServiceID)
	return nil
}

func (bs *BucketStorageImpl) ConsumeService(clientID string, serviceID string, usageAmount uint64) (accRes AccessStatusResponse, err error) {
	log.Printf("event=consume_service client_id=%q service_id=%q", clientID, serviceID)
	requestedService, err := bs.ServiceRegistry.GetService(serviceID)
	if err != nil {
		log.Printf("error=service_not_found client_id=%q service_id=%q err=%v", clientID, serviceID, err)
		return
	}
	b := bs.BucketsMap[clientID][serviceID]
	if b == nil {
		return accRes, ErrServiceNotFound
	}
	refill(b)
	log.Printf("event=get_bucket_status client_id=%q service_id=%q tokens=%d usage_price=%d", clientID, serviceID, b.tokens, requestedService.UsagePriceInTokens)
	if b.tokens < requestedService.UsagePriceInTokens {
		accRes.IsAllowed = false
		accRes.RetryAfterSeconds = requestedService.UsagePriceInTokens / b.refillRatePerSecond
		log.Printf("access_denied client_id=%q service_id=%q tokens=%d retry_after=%d", clientID, serviceID, b.tokens, accRes.RetryAfterSeconds)
		return
	}
	b.tokens -= requestedService.UsagePriceInTokens * usageAmount
	accRes.IsAllowed = true
	accRes.RetryAfterSeconds = requestedService.UsagePriceInTokens / b.refillRatePerSecond
	log.Printf("event=tokens_consumed client_id=%q service_id=%q tokens_left=%d", clientID, serviceID, b.tokens)
	return
}

func refill(b *Bucket) {
	if b == nil {
		return
	}
	refilled := uint64(time.Since(b.LastRefill).Seconds()) * b.refillRatePerSecond
	if refilled > 0 {
		b.tokens += refilled
		log.Printf("event=bucket_refilled client_id=%q service_id=%q tokens_added=%d new_tokens=%d", b.clientID, b.serviceID, refilled, b.tokens)
	}
	b.LastRefill = time.Now()
}

func NewBucketStorage(serviceRegistry ServiceRegistry) BucketStorage {
	return &BucketStorageImpl{
		BucketsMap:      make(map[string]map[string]*Bucket),
		ServiceRegistry: serviceRegistry,
	}
}
