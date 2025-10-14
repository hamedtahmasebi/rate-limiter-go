package main

import (
	"log"
	"net"
	"rate-limiter-go/api"
	"rate-limiter-go/limiter"

	"google.golang.org/grpc"
)

func main() {
	log.Printf("event=init action=NewServiceRegistry")
	mainServiceRegistry := limiter.NewServiceRegistry()
	log.Printf("event=init action=NewBucketStorage")
	mainBucketStorage := limiter.NewBucketStorage(mainServiceRegistry)

	log.Printf("event=create_service id=%q usage_price_in_tokens=%d", "test_service", 2)
	testService, err := mainServiceRegistry.CreateService(limiter.CreateServiceReqBody{
		ID:                 "test_service",
		UsagePriceInTokens: 2,
	})
	if err != nil {
		log.Printf("event=create_service status=error error=%q", err)
		panic(err)
	}

	log.Printf("event=create_bucket client_id=%q service_id=%q initial_tokens=%d refill_rate_per_second=%d", "main_client", testService.ID, 100, 5)
	mainBucketStorage.CreateBucket(limiter.CreateBucketReqBody{
		ClientID:            "main_client",
		ServiceID:           testService.ID,
		InitialTokens:       100,
		RefillRatePerSecond: 5,
	})

	log.Printf("event=server_setup status=starting")
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Printf("event=server_setup status=error error=%q", err)
		panic(err)
	}
	grpcServer := grpc.NewServer()
	api.RegisterRateLimiterServer(grpcServer, &api.Server{
		BucketStorage:   mainBucketStorage,
		ServiceRegistry: mainServiceRegistry,
	})

	log.Printf("event=server status=listening port=%q", ":50051")

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Printf("event=server status=error error=%q", err)
		panic(err)
	}
}
