package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/flagsense/go-sdk/config"
	"github.com/flagsense/go-sdk/pkg/dto"
	flagsenseHttpClient "github.com/flagsense/go-sdk/pkg/infrastructure/http"
	"github.com/flagsense/go-sdk/pkg/model"
	"fmt"
	"github.com/teltech/logger"
	"net/http"
	"sync"
	"time"
)

const (
	CONTENT_TYPE      = "content-type"
	APPLICATION_JSON  = "application/json"
	HEADER_AUTH_TYPE  = "authType"
	HEADER_SDK_ID     = "sdkId"
	HEADER_SDK_SECRET = "sdkSecret"
	SDK               = "sdk"
)

type DataPollerServiceImpl struct {
	SDKConfig       *model.SDKConfig
	PollingInterval time.Duration
	logger          *logger.Log
	config          *config.Store
	client          *http.Client
	Data            *dto.Data
	Cond            *sync.Cond
	Mutex           *sync.Mutex
}

type DataPollerRequest struct {
	LastUpdatedOn float64 `json:"lastUpdatedOn" validate:"required"`
	Environment   string  `json:"environment" validate:"required"`
}

func NewDataPollerService(sdkConfig *model.SDKConfig, pollingInterval time.Duration, logger *logger.Log, config *config.Store,
	client *http.Client, data *dto.Data) *DataPollerServiceImpl {
	mutex := &sync.Mutex{}
	cond := sync.NewCond(mutex)
	return &DataPollerServiceImpl{
		SDKConfig:       sdkConfig,
		PollingInterval: pollingInterval,
		logger:          logger,
		config:          config,
		client:          client,
		Data:            data,
		Cond:            cond,
		Mutex:           mutex,
	}
}

func (dps *DataPollerServiceImpl) Start(ctx context.Context) {
	// for first time initialization at start
	dps.fetchLatest(ctx)

	tick := time.NewTicker(dps.PollingInterval)
	for {
		select {
		case <-tick.C:
			dps.fetchLatest(ctx)
		case <-ctx.Done():
			//dps.logger.Info("Polling Config Manager Stopped")
			return
		}
	}
}

func (dps *DataPollerServiceImpl) fetchLatest(ctx context.Context) {
	//fmt.Println("Fetching latest data at: ", time.Now().String())
	endpoint := fmt.Sprintf("%s/fetchLatest", dps.config.Services.SDKService.HttpEndpoint.Url)
	headers := map[string]string{
		CONTENT_TYPE:      APPLICATION_JSON,
		HEADER_AUTH_TYPE:  SDK,
		HEADER_SDK_ID:     dps.SDKConfig.SDKId,
		HEADER_SDK_SECRET: dps.SDKConfig.SDKSecret,
	}

	payload := DataPollerRequest{
		LastUpdatedOn: dps.Data.LastUpdatedOn,
		Environment:   dps.SDKConfig.Environment,
	}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		//dps.logger.Errorf("error while parsing request payload:%+v, error:%+v", payload, err)
		return
	}
	response, err := flagsenseHttpClient.MakeHttpRequest(ctx, "POST", endpoint, dps.client,
		bytes.NewBuffer(requestBody), headers)
	if err != nil {
		//dps.logger.Errorf("error while fetching latest flags data:%+v, error:%+v", payload, err)
		return
	}
	if response == nil || len(response) == 0 {
		return
	}
	var newData dto.Data
	err = json.Unmarshal(response, &newData)
	if err != nil {
		//dps.logger.Errorf("error while parsing response payload, error:%+v", err)
		return
	}

	if newData.LastUpdatedOn > 0 && newData.Flags != nil && newData.Segments != nil {
		newData.MatchType()
		if len(newData.Segments) > 0 {
			dps.Data.Segments = newData.Segments
		}
		if len(newData.Flags) > 0 {
			dps.Data.Flags = newData.Flags
		}

		dps.Mutex.Lock()
		dps.Data.LastUpdatedOn = newData.LastUpdatedOn
		dps.Mutex.Unlock()
		dps.Cond.Broadcast()
	}
}

func (dps *DataPollerServiceImpl) WaitForInitializationComplete() {
	dps.Mutex.Lock()
	defer dps.Mutex.Unlock()
	for dps.Data.LastUpdatedOn == ZER0 {
		dps.Cond.Wait()
	}
}
