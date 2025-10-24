package config

import (
	"encoding/json"
	"io"
	"log"
)

type LimitRule struct {
	ID                  string `json:"id"`
	ClientID            string `json:"client_id"`
	ServiceID           string `json:"service_id"`
	UsagePrice          uint64 `json:"usage_price"`
	RefillRatePerSecond uint64 `json:"refill_rate_per_second"`
	InitialTokens       uint64 `json:"initial_tokens"`
}

type PersistenceSettings struct {
	Disabled bool
}

type Config struct {
	Rules               []LimitRule         `json:"rules"`
	PersistenceSettings PersistenceSettings `json:"persistence_settings"`
}

type ConfigParser interface {
	Parse(io.Reader) *Config
}

type jsonParser struct{}

func (j *jsonParser) Parse(in io.Reader) *Config {
	var config Config
	err := json.NewDecoder(in).Decode(&config)
	if err != nil {
		log.Printf("event=failed_to_parse_json_config err=%q", err)
		return nil
	}
	return &config
}

func NewJsonParser() ConfigParser {
	return &jsonParser{}
}
