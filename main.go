package main

import (
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"path/filepath"
	"rate-limiter-go/api"
	"rate-limiter-go/config"
	"rate-limiter-go/limiter"
	"rate-limiter-go/persist"
	"strings"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// setup log file
	logFile, err := os.OpenFile("./logs/main.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	writer := io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(writer)
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.Printf("Logger Initialized")
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
		var persistInterval uint8 = 10
		if config.PersistenceSettings.IntervalSeconds > 0 {
			persistInterval = config.PersistenceSettings.IntervalSeconds
		}

		var persistence_dir = "./persistence_files"
		persist.InitializePersistenceDir(persistence_dir)
		jw := &persist.JsonWriter[limiter.Bucket]{}
		if err != nil {
			panic(err)
		}

		// Load persisted buckets
		filepath.WalkDir(persistence_dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".json") {
				bucket, err := jw.LoadFromFile(path)
				if err != nil {
					log.Printf("level=error event=load_persisted_bucket status=error error=%q", err)
					return err
				}
				err = mainBucketStorage.RestoreBucket(bucket)
				if err != nil {
					panic(err)
				}
			}
			return nil
		})

		// Save buckets on an interval
		go func() {
			ticker := time.NewTicker(time.Second * time.Duration(persistInterval))
			for range ticker.C {
				allBuckets := mainBucketStorage.GetAllBuckets()
				log.Printf("level=info event=periodic_save start saving %d buckets", len(allBuckets))
				for _, b := range allBuckets {
					err := jw.SaveToFile(*b, b.ID)
					if err != nil {
						log.Printf("level=error event=persist_to_json status=error err=%q", err)
					}
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
