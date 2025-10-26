package config

import (
	"encoding/json"
	"errors"
	"io"
	"log"
)

var ErrInvalidMaxTokens = errors.New("max token should be specified for every rule with a value > 0")

type LimitRule struct {
	ID                  string `json:"id"`
	ClientID            string `json:"client_id"`
	ServiceID           string `json:"service_id"`
	UsagePrice          uint64 `json:"usage_price"`
	RefillRatePerSecond uint64 `json:"refill_rate_per_second"`
	InitialTokens       uint64 `json:"initial_tokens"`
	MaxTokens           uint64 `json:"max_tokens"`
}

type PersistenceSettings struct {
	Disabled        bool  `json:"disabled"`
	IntervalSeconds uint8 `json:"interval_seconds"`
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
	validateConfig(&config)
	return &config
}

func validateConfig(c *Config) {
	for _, rule := range c.Rules {
		log.Printf("max tokens = %d", rule.MaxTokens)
		if rule.MaxTokens <= 0 {

			panic(ErrInvalidMaxTokens)
		}
	}
}

func NewJsonParser() ConfigParser {
	return &jsonParser{}
}
