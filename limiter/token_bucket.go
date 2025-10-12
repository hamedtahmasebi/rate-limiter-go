package limiter

import (
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
	isAllowed         bool
	retryAfterSeconds uint64
}

type BucketStorage interface {
	CreateBucket(body CreateBucketReqBody) error
	ConsumeService(clientID string, serviceID string) (AccessStatusResponse, error)
	// ConsumeTokens(clientID string, serviceID string, amount uint)
}

type BucketStorageImpl struct {
	BucketsMap      map[string]map[string]*Bucket
	ServiceRegistry ServiceRegistry
}

func (bs *BucketStorageImpl) CreateBucket(body CreateBucketReqBody) error {
	bs.BucketsMap[body.ClientID][body.ServiceID] = &Bucket{
		clientID:            body.ClientID,
		serviceID:           body.ServiceID,
		tokens:              body.InitialTokens,
		refillRatePerSecond: body.RefillRatePerSecond,
		CreatedAt:           time.Now(),
		LastRefill:          time.Now(),
	}
	return nil
}

func (bs *BucketStorageImpl) ConsumeService(clientID string, serviceID string) (accRes AccessStatusResponse, err error) {
	requestedService, err := bs.ServiceRegistry.GetService(serviceID)
	if err != nil {
		return
	}
	b := bs.BucketsMap[clientID][serviceID]
	refill(b)
	if b.tokens < requestedService.UsagePriceInTokens {
		accRes.isAllowed = false
		accRes.retryAfterSeconds = b.refillRatePerSecond * requestedService.UsagePriceInTokens
		return
	}
	b.tokens -= requestedService.UsagePriceInTokens
	accRes.isAllowed = false
	accRes.retryAfterSeconds = requestedService.UsagePriceInTokens / b.refillRatePerSecond
	return
}

func refill(b *Bucket) {
	b.tokens = uint64(time.Since(b.LastRefill).Seconds()) * b.refillRatePerSecond
	b.LastRefill = time.Now()
}

func NewBucketStorage(serviceRegistry ServiceRegistry) BucketStorage {
	return &BucketStorageImpl{
		BucketsMap:      make(map[string]map[string]*Bucket),
		ServiceRegistry: serviceRegistry,
	}
}
