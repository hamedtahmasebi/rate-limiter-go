package persist

import (
	"encoding/json"
	"log"
	"os"
)

type FileWriter[T any] interface {
	SaveToFile(entity T, filepath string) error
	LoadFromFile(filepath string) (*T, error)
}

type JsonWriter[T any] struct{}

var persistence_files_path string

func (jw *JsonWriter[T]) SaveToFile(entity T, filename string) error {
	filePath := persistence_files_path + "/" + filename + ".json"
	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Printf("level=error event=open_or_create_file status=error filepath=%s err=%q", filePath, err)
		return err
	}
	err = json.NewEncoder(fd).Encode(entity)
	if err != nil {
		log.Printf("level=error event=encode_json_to_file status=error err=%q", err)
		return err
	}
	return nil
}

func (jw *JsonWriter[T]) LoadFromFile(filename string) (*T, error) {
	filePath := persistence_files_path + "/" + filename
	var entity T
	fd, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Printf("level=error event=read_file filepath=%s err=%q", filePath, err)
		return nil, err
	}
	err = json.NewDecoder(fd).Decode(&entity)
	if err != nil {
		log.Printf("level=error event=decode_json_from_file filepath=%s err=%q", filePath, err)
		return nil, err
	}
	return &entity, nil
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
	persistence_files_path = dir
}
