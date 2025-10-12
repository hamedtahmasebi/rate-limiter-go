package main

import "rate-limiter-go/limiter"

func main() {
	mainServiceRegistry := limiter.NewServiceRegistry()
	mainBucketStorage := limiter.NewBucketStorage(mainServiceRegistry)

	testService, err := mainServiceRegistry.CreateService(limiter.CreateServiceReqBody{
		ID:                 "test_servie",
		UsagePriceInTokens: 2,
	})
	if err != nil {
		panic(err)
	}

	mainBucketStorage.CreateBucket(limiter.CreateBucketReqBody{
		ClientID:            "main_client",
		ServiceID:           testService.ID,
		InitialTokens:       100,
		RefillRatePerSecond: 5,
	})

}
