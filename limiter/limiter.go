package limiter

type RateLimitingReqeustContext struct {
	ServiceID   string
	ClientID    string
	UserID      string
	UsageAmount uint64
}

type RateLimitResponse struct {
	IsAllowed         bool
	RetryAfterSeconds uint64
}

type RateLimiter interface {
	ConsumeService(RateLimitingReqeustContext) (RateLimitResponse, error)
}
