package limiter

import (
	"errors"
	"log"
	"sync"
	"time"
)

var ErrBucketNotFound = errors.New("bucket not found")
var ErrCreateBucketIdCollision = errors.New("a bucket already exists with this id")

type CreateBucketReqBody struct {
	ID                  string
	InitialTokens       uint64
	RefillRatePerSecond uint64
	MaxTokens           uint64
}

type Bucket struct {
	ID                  string
	Tokens              uint64     `json:"tokens"`
	RefillRatePerSecond uint64     `json:"refill_rate_per_second"`
	CreatedAt           time.Time  `json:"created_at"`
	LastRefill          time.Time  `json:"last_refill"`
	MaxTokens           uint64     `json:"max_tokens"`
	Mu                  sync.Mutex `json:"-"`
}

type AccessStatusResponse struct {
	IsAllowed         bool
	RetryAfterSeconds uint64
}

type ConsumeServiceRequest struct {
	ServiceID   string
	ClientID    string
	UserID      string
	UsageAmount uint64
}

type BucketStorage interface {
	CreateBucket(body CreateBucketReqBody) error
	RestoreBucket(body *Bucket) error
	ConsumeService(body ConsumeServiceRequest) (AccessStatusResponse, error)
	GetAllBuckets() []*Bucket
	GetBucket(ID string) (*Bucket, error)
}

type BucketStorageImpl struct {
	BucketsMap      map[string]*Bucket
	ServiceRegistry ServiceRegistry
}

func (bs *BucketStorageImpl) GetBucket(id string) (*Bucket, error) {
	b, exists := bs.BucketsMap[id]
	if !exists {
		return nil, ErrBucketNotFound
	}
	return b, nil
}

func (bs *BucketStorageImpl) RestoreBucket(bucket *Bucket) error {
	bucket.Mu.Lock()
	defer bucket.Mu.Unlock()

	_, exists := bs.BucketsMap[bucket.ID]
	if exists {
		log.Printf("level=warn event=restore_bucket bucket already exists")
		return ErrCreateBucketIdCollision
	}
	bs.BucketsMap[bucket.ID] = bucket
	return nil
}

func (bs *BucketStorageImpl) CreateBucket(body CreateBucketReqBody) error {
	if body.MaxTokens <= 0 {
		log.Fatalf("Max Tokens is not defined for bucket, bucket_id:%s", body.ID)
	}
	log.Printf("event=create_bucket bucket_id=%q initial_tokens=%d refill_rate_per_second=%d max_tokens=%d", body.ID, body.InitialTokens, body.RefillRatePerSecond, body.MaxTokens)
	_, bExists := bs.BucketsMap[body.ID]
	if bExists {
		log.Printf("event=create_bucket status=error errors=%q", ErrCreateBucketIdCollision)
		return ErrCreateBucketIdCollision
	}
	newBucket := &Bucket{
		ID:                  body.ID,
		Tokens:              body.InitialTokens,
		RefillRatePerSecond: body.RefillRatePerSecond,
		MaxTokens:           body.MaxTokens,
		Mu:                  sync.Mutex{},
		CreatedAt:           time.Now(),
		LastRefill:          time.Now(),
	}
	bs.BucketsMap[newBucket.ID] = newBucket
	log.Printf("event=bucket_created bucket_id=%q", newBucket.ID)

	return nil
}

func (bs *BucketStorageImpl) ConsumeService(body ConsumeServiceRequest) (accRes AccessStatusResponse, err error) {
	log.Printf("event=consume_service status=started client_id=%s user_id=%s", body.ClientID, body.UserID)
	requestedService, err := bs.ServiceRegistry.GetService(body.ServiceID)
	if err != nil {
		log.Printf("error=service_not_found client_id=%q service_id=%q err=%v", body.ClientID, body.ServiceID, err)
		return
	}

	bucketID := GetBucketID(GetBucketIDRequest{
		ClientID:  body.ClientID,
		ServiceID: body.ServiceID,
		UserID:    body.UserID,
	})
	b := bs.BucketsMap[bucketID]

	if b == nil {
		return accRes, ErrBucketNotFound
	}

	refill(b)

	b.Mu.Lock()
	defer b.Mu.Unlock()

	log.Printf("event=get_bucket_status service_id=%s client_id=%s user_id=%s tokens=%d usage_price=%d", body.ServiceID, body.ClientID, body.UserID, b.Tokens, requestedService.UsagePriceInTokens)
	consumeAmount := requestedService.UsagePriceInTokens * body.UsageAmount
	if b.Tokens < consumeAmount {
		accRes.IsAllowed = false
		accRes.RetryAfterSeconds = (consumeAmount - b.Tokens) / b.RefillRatePerSecond
		log.Printf("event=insufficient_tokens service_id=%s client_id=%s user_id=%s tokens=%d retry_after=%d", body.ServiceID, body.ClientID, body.UserID, b.Tokens, accRes.RetryAfterSeconds)
		return
	}
	b.Tokens -= consumeAmount
	accRes.IsAllowed = true
	accRes.RetryAfterSeconds = 0
	log.Printf("event=consume_tokens client_id=%s service_id=%s user_id=%s tokens_consumed=%d tokens_left=%d", body.ClientID, body.ServiceID, body.UserID, consumeAmount, b.Tokens)
	return
}

func (bs *BucketStorageImpl) GetAllBuckets() []*Bucket {
	buckets := make([]*Bucket, 0)
	for _, b := range bs.BucketsMap {
		buckets = append(buckets, b)
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
		if b.Tokens >= b.MaxTokens {
			b.Tokens = b.MaxTokens
		}
		log.Printf("event=bucket_refilled bucket_id=%s tokens_added=%d new_tokens=%d", b.ID, refilled, b.Tokens)
	}
	b.LastRefill = time.Now()
}

func NewBucketStorage(serviceRegistry ServiceRegistry) BucketStorage {
	return &BucketStorageImpl{
		BucketsMap:      make(map[string]*Bucket),
		ServiceRegistry: serviceRegistry,
	}
}

type GetBucketIDRequest struct {
	ServiceID string
	ClientID  string
	UserID    string
}

func GetBucketID(dto GetBucketIDRequest) string {
	return dto.ServiceID + "_" + dto.ClientID + "_" + dto.UserID
}
