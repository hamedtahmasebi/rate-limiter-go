package limiter

import (
	"log"
	"sync"
	"time"
)

type CreateBucketReqBody struct {
	ClientID            string
	ServiceID           string
	InitialTokens       uint64
	RefillRatePerSecond uint64
}

type Bucket struct {
	ClientID            string     `json:"client_id"`
	ServiceID           string     `json:"service_id"`
	Tokens              uint64     `json:"tokens"`
	RefillRatePerSecond uint64     `json:"refill_rate_per_second"`
	CreatedAt           time.Time  `json:"created_at"`
	LastRefill          time.Time  `json:"last_refill"`
	Mu                  sync.Mutex `json:"-"`
}

type AccessStatusResponse struct {
	IsAllowed         bool
	RetryAfterSeconds uint64
}

type BucketStorage interface {
	CreateBucket(body CreateBucketReqBody) error
	ConsumeService(clientID string, serviceID string, usageAmount uint64) (AccessStatusResponse, error)
	GetAllBuckets() []*Bucket
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
		ClientID:            body.ClientID,
		ServiceID:           body.ServiceID,
		Tokens:              body.InitialTokens,
		RefillRatePerSecond: body.RefillRatePerSecond,
		CreatedAt:           time.Now(),
		LastRefill:          time.Now(),
		Mu:                  sync.Mutex{},
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
	clientBucket := bs.BucketsMap[clientID]
	if clientBucket == nil {
		return accRes, ErrClientNotFound
	}
	b := clientBucket[serviceID]
	if b == nil {
		return accRes, ErrServiceNotFound
	}
	refill(b)
	b.Mu.Lock()
	defer b.Mu.Unlock()
	log.Printf("event=get_bucket_status client_id=%q service_id=%q tokens=%d usage_price=%d", clientID, serviceID, b.Tokens, requestedService.UsagePriceInTokens)
	if b.Tokens < requestedService.UsagePriceInTokens {
		accRes.IsAllowed = false
		accRes.RetryAfterSeconds = requestedService.UsagePriceInTokens / b.RefillRatePerSecond
		log.Printf("access_denied client_id=%q service_id=%q tokens=%d retry_after=%d", clientID, serviceID, b.Tokens, accRes.RetryAfterSeconds)
		return
	}
	b.Tokens -= requestedService.UsagePriceInTokens * usageAmount
	accRes.IsAllowed = true
	accRes.RetryAfterSeconds = requestedService.UsagePriceInTokens / b.RefillRatePerSecond
	log.Printf("event=tokens_consumed client_id=%q service_id=%q tokens_left=%d", clientID, serviceID, b.Tokens)
	return
}

func (bs *BucketStorageImpl) GetAllBuckets() []*Bucket {
	buckets := make([]*Bucket, 0)
	for _, clientBuckets := range bs.BucketsMap {
		for _, bucket := range clientBuckets {
			buckets = append(buckets, bucket)
		}
	}
	return buckets
}

func refill(b *Bucket) {
	if b == nil {
		return
	}
	b.Mu.Lock()
	defer b.Mu.Unlock()
	refilled := uint64(time.Since(b.LastRefill).Seconds()) * b.RefillRatePerSecond
	if refilled > 0 {
		b.Tokens += refilled
		log.Printf("event=bucket_refilled client_id=%q service_id=%q tokens_added=%d new_tokens=%d", b.ClientID, b.ServiceID, refilled, b.Tokens)
	}
	b.LastRefill = time.Now()
}

func NewBucketStorage(serviceRegistry ServiceRegistry) BucketStorage {
	return &BucketStorageImpl{
		BucketsMap:      make(map[string]map[string]*Bucket),
		ServiceRegistry: serviceRegistry,
	}
}
