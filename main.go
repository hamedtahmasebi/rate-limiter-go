package main

import (
	"log"
	"net"
	"os"
	"rate-limiter-go/api"
	"rate-limiter-go/config"
	"rate-limiter-go/limiter"
	"rate-limiter-go/persist"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// read config
	if len(os.Args) < 2 {
		panic("Please provide the path to the config")
	}
	configFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Printf("event=failed_to_open_config_file err=%q", err)
		panic(err)
	}
	defer configFile.Close()
	configParser := config.NewJsonParser()
	config := configParser.Parse(configFile)
	if config == nil {
		log.Printf("event=failed_to_parse_config err=%q", err)
		panic("failed to parse config")
	}

	log.Printf("event=init action=NewServiceRegistry")
	mainServiceRegistry := limiter.NewServiceRegistry()
	log.Printf("event=init action=NewBucketStorage")
	mainBucketStorage := limiter.NewBucketStorage(mainServiceRegistry)

	if !config.PersistenceSettings.Disabled {
		persistence_files_dir := "./persistence_files"
		_, err := os.Stat(persistence_files_dir)
		if os.IsNotExist(err) {
			err := os.Mkdir(persistence_files_dir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		} else {
			if err != nil {
				panic(err)
			}
		}

		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for _ = range ticker.C {
				allBuckets := mainBucketStorage.GetAllBuckets()
				for _, b := range allBuckets {
					persist.SaveToFile(b)
				}
			}
		}()
	}

	for _, rule := range config.Rules {
		log.Printf("event=create_service id=%q usage_price_in_tokens=%d", rule.ServiceID, rule.UsagePrice)
		_, err := mainServiceRegistry.CreateService((limiter.CreateServiceReqBody{
			ID:                 rule.ServiceID,
			UsagePriceInTokens: rule.UsagePrice,
		}))
		if err != nil {
			log.Printf("event=create_service status=error error=%q", err)
			panic(err)
		}

		log.Printf("event=create_bucket client_id=%q service_id=%q initial_tokens=%d refill_rate_per_second=%d", rule.ClientID, rule.ServiceID, rule.InitialTokens, rule.RefillRatePerSecond)
		err = mainBucketStorage.CreateBucket(limiter.CreateBucketReqBody{
			ClientID:            rule.ClientID,
			ServiceID:           rule.ServiceID,
			InitialTokens:       rule.InitialTokens,
			RefillRatePerSecond: rule.RefillRatePerSecond,
		})
		if err != nil {
			log.Printf("event=create_bucket status=error error=%q", err)
			panic(err)
		}

	}

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
