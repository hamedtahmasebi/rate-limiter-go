package persist

import (
	"encoding/json"
	"log"
	"os"
	"rate-limiter-go/limiter"
)

func SaveToFile(bucket *limiter.Bucket) {
	bucket.Mu.Lock()
	defer bucket.Mu.Unlock()
	filename := "./persistence_files/" + bucket.ClientID + "_" + bucket.ServiceID + ".json"
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
