package persist

import (
	"encoding/json"
	"log"
	"os"
	"rate-limiter-go/limiter"
	"time"
)

func saveToFile(bucket *limiter.Bucket, dir string) {
	bucket.Mu.Lock()
	defer bucket.Mu.Unlock()
	filename := dir + bucket.ClientID + "_" + bucket.ServiceID + ".json"
	log.Printf("event=write_to_file filename=%s", filename)
	file, err := os.OpenFile(filename,
		os.O_CREATE| // create file if doesn't exist
			os.O_TRUNC| // truncate file to 0 bytes
			os.O_WRONLY, // write only, not append
		os.ModePerm)
	if err != nil {
		log.Printf("error=write_to_file filename=%s  err=%q", filename, err)
		return
	}

	err = json.NewEncoder(file).Encode(bucket)
	if err != nil {
		log.Printf("event=error_writing_json_to_file action=SaveToFile status=error err=%q", err)
	}
}

func loadFromFile(path string, storage limiter.BucketStorage) error {
	log.Printf("event=load_from_file status=started filename=%s", path)
	f, err := os.Open(path)
	if err != nil {
		log.Printf("error=load_from_file filename=%s err=%q", path, err)
		return err
	}
	defer f.Close()
	b := limiter.Bucket{}
	err = json.NewDecoder(f).Decode(&b)
	if err != nil {
		log.Printf("event=decode_bucket_json_failure status=error filename=%s err=%q", f.Name(), err)
		return err
	}
	err = storage.CreateBucket(limiter.CreateBucketReqBody{
		ClientID:            b.ClientID,
		ServiceID:           b.ServiceID,
		InitialTokens:       b.Tokens,
		RefillRatePerSecond: b.RefillRatePerSecond,
		MaxTokens:           b.MaxTokens,
	})
	if err != nil {
		log.Printf("event=create_bucket_from_file status=error filename=%s err=%q", f.Name(), err)
		return err
	}

	log.Printf("event=load_from_file status=success filename=%s", f.Name())
	return nil
}

func LoadFromFilesToStorage(dir string, storage limiter.BucketStorage) error {
	log.Printf("event=load_from_files_to_storage status=started dir=%s", dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("event=read_dir_failure action=LoadFromFilesToStorage dir=%q error=%q", dir, err)
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			LoadFromFilesToStorage(dir+"/"+e.Name(), storage)
		} else {
			loadFromFile(dir+"/"+e.Name(), storage)
		}
	}
	return nil
}

func InitializePersistenceDir(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	} else {
		if err != nil {
			panic(err)
		}
	}
}

func RunAutoSaveWorker(bs limiter.BucketStorage, interval time.Duration, dir string) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			allBuckets := bs.GetAllBuckets()
			for _, b := range allBuckets {
				saveToFile(b, dir+"/")
			}
		}
	}()
}
