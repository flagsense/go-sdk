package config

import (
	"encoding/json"
	"flagsense-go-sdk/third_party/assetmnger"
	"fmt"
	"time"
)

const (
	prodFilePath = "config/prod.json"
)

type Store struct {
	Services  ServiceDefinitions
	Constants Constants
}

type ServiceDefinitions struct {
	SDKService    ServiceEndpoint `json:"sdk-service" validate:"required"`
	EventsService ServiceEndpoint `json:"events-service" validate:"required"`
}

type ServiceEndpoint struct {
	HttpEndpoint HTTPEndpoint `json:"http_endpoint" validate:"required"`
}

type HTTPEndpoint struct {
	Url            string        `validate:"required"`
	DefaultTimeout time.Duration `json:"timeout" validate:"required"`
}

type Constants struct {
	PollingInterval int  `json:"polling-interval" default:"5"`
	CaptureEvents   bool `json:"capture-events" default:"true"`
}

func NewConfig(am *assetmnger.Manager) *Store {
	var config Store
	fileToOpen := prodFilePath

	byteValue := am.Get(fileToOpen)
	if byteValue == nil {
		fmt.Println("FAILED_TO_GET_FLAGSENSE_CONFIG")
	}
	err := json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Printf("Error while parsing config, err: %+v", err)
		panic(err)
	}
	return &config
}
