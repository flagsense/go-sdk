package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/flagsense/go-sdk/config"
	flagsenseHttpClient "github.com/flagsense/go-sdk/pkg/infrastructure/http"
	"github.com/flagsense/go-sdk/pkg/model"
	guuid "github.com/google/uuid"
	"github.com/orcaman/concurrent-map"
	"github.com/teltech/logger"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type EventServiceImpl struct {
	logger         *logger.Log
	sdkConfig      *model.SDKConfig
	client         *http.Client
	requests       *cmap.ConcurrentMap
	errors         *cmap.ConcurrentMap
	data           *cmap.ConcurrentMap
	codeBugs       *cmap.ConcurrentMap
	timeslot       int64
	variantMapLock *sync.Mutex
	codeBugsLock   *sync.Mutex
	errorLock      *sync.Mutex
	refreshLock    *sync.Mutex
	config         *config.Store
	machineId      string
}

type VariantsRequest struct {
	MachineId   string                 `json:"machineId"`
	SdkType     string                 `json:"sdkType"`
	Environment string                 `json:"environment"`
	Data        map[string]interface{} `json:"data"`
	CodeBugs    map[string]interface{} `json:"codeBugs"`
	Errors      map[string]interface{} `json:"errors"`
	Time        int64                  `json:"time"`
}

const (
	EVENT_FLUSH_INTITAL_DELAY = 2
	EVENT_FLUSH_INTERVAL      = 5
	SDK_TYPE                  = "go"
)

var MILLIS_IN_EVENT_FLUSH_INTERVAL int64 = EVENT_FLUSH_INTERVAL * 60 * 1000

func NewEventService(logger *logger.Log, sdkConfig *model.SDKConfig, client *http.Client, config *config.Store) *EventServiceImpl {
	errors := cmap.New()
	data := cmap.New()
	codeBugs := cmap.New()
	timeslot := getTimeSlot(time.Now().Unix() * 1000)
	requests := cmap.New()
	return &EventServiceImpl{
		logger:         logger,
		sdkConfig:      sdkConfig,
		requests:       &requests,
		errors:         &errors,
		data:           &data,
		codeBugs:		&codeBugs,
		timeslot:       timeslot,
		variantMapLock: &sync.Mutex{},
		codeBugsLock:   &sync.Mutex{},
		errorLock:      &sync.Mutex{},
		refreshLock:    &sync.Mutex{},
		client:         client,
		config:         config,
		machineId:      guuid.NewString(),
	}
}

func (es *EventServiceImpl) Start(ctx context.Context) {
	time.Sleep(EVENT_FLUSH_INTITAL_DELAY * time.Minute)
	es.Run(ctx)
	tick := time.NewTicker(EVENT_FLUSH_INTERVAL * time.Minute)
	for {
		select {
		case <-tick.C:
			es.Run(ctx)
		case <-ctx.Done():
			//es.logger.Info("Events Service stopped")
			return
		}
	}
}

func (es *EventServiceImpl) AddEvaluationCount(flagId string, variantKey string) {
	defer func() { //catch or finally
	}()

	if !es.config.Constants.CaptureEvents {
		return
	}
	currentTimeSlot := getTimeSlot(time.Now().Unix() * 1000)
	if currentTimeSlot != es.timeslot {
		es.checkAndRefreshData(currentTimeSlot)
	}
	variantMap, present := es.data.Get(flagId)
	if !present {
		absent := es.data.SetIfAbsent(flagId, cmap.New())
		if absent {
			variantMap, _ = es.data.Get(flagId)
		}
	}

	es.variantMapLock.Lock()
	val, present := variantMap.(cmap.ConcurrentMap).Get(variantKey)
	if !present {
		variantMap.(cmap.ConcurrentMap).Set(variantKey, int64(1))
	} else {
		variantMap.(cmap.ConcurrentMap).Set(variantKey, val.(int64)+int64(1))
	}
	es.variantMapLock.Unlock()
}

func (es *EventServiceImpl) AddErrorsCount(flagId string) {
	defer func() { //catch or finally
	}()

	if !es.config.Constants.CaptureEvents {
		return
	}
	currentTimeSlot := getTimeSlot(time.Now().Unix() * 1000)
	if currentTimeSlot != es.timeslot {
		es.checkAndRefreshData(currentTimeSlot)
	}

	es.errorLock.Lock()
	val, present := es.errors.Get(flagId)
	if !present {
		es.errors.Set(flagId, int64(1))
	} else {
		es.errors.Set(flagId, val.(int64)+int64(1))
	}
	es.errorLock.Unlock()
}

func (es *EventServiceImpl) AddCodeBugsCount(flagId string, variantKey string)  {
	defer func() { //catch or finally
	}()

	if !es.config.Constants.CaptureEvents {
		return
	}
	currentTimeSlot := getTimeSlot(time.Now().Unix() * 1000)
	if currentTimeSlot != es.timeslot {
		es.checkAndRefreshData(currentTimeSlot)
	}
	variantMap, present := es.codeBugs.Get(flagId)
	if !present {
		absent := es.codeBugs.SetIfAbsent(flagId, cmap.New())
		if absent {
			variantMap, _ = es.codeBugs.Get(flagId)
		}
	}

	es.codeBugsLock.Lock()
	val, present := variantMap.(cmap.ConcurrentMap).Get(variantKey)
	if !present {
		variantMap.(cmap.ConcurrentMap).Set(variantKey, int64(1))
	} else {
		variantMap.(cmap.ConcurrentMap).Set(variantKey, val.(int64)+int64(1))
	}
	es.codeBugsLock.Unlock()
}

func (es *EventServiceImpl) checkAndRefreshData(timeslot int64) {
	es.refreshLock.Lock()
	if timeslot != es.timeslot {
		es.refreshData(timeslot)
	}
	es.refreshLock.Unlock()
}

func (es *EventServiceImpl) refreshData(currentTimeSlot int64) {

	variantRequest := VariantsRequest{
		MachineId:   es.machineId,
		Environment: es.sdkConfig.Environment,
		SdkType:     SDK_TYPE,
		Data:        make(map[string]interface{}),
		CodeBugs:    make(map[string]interface{}),
		Errors:      make(map[string]interface{}),
		Time:        es.timeslot,
	}
	for key, value := range es.data.Items() {
		variantRequest.Data[key] = value.(cmap.ConcurrentMap).Items()
	}
	for key, value := range es.codeBugs.Items() {
		variantRequest.CodeBugs[key] = value.(cmap.ConcurrentMap).Items()
	}
	variantRequest.Errors = es.errors.Items()

	if len(variantRequest.Data) != 0 || len(variantRequest.CodeBugs) != 0 || len(variantRequest.Errors) != 0 {
		es.requests.Set(strconv.FormatInt(es.timeslot, 10), variantRequest)
	}

	es.data.Clear()
	es.codeBugs.Clear()
	es.errors.Clear()
	es.timeslot = currentTimeSlot
}

func (es *EventServiceImpl) Run(ctx context.Context) {
	//fmt.Println("Sending events at: ", time.Now().String())
	timeKeys := es.requests.Keys()
	for _, key := range timeKeys {
		requestBody, present := es.requests.Get(key)
		if present {
			es.SendEvents(ctx, requestBody.(VariantsRequest))
		}
		es.requests.Remove(key)
	}
}

func (es *EventServiceImpl) SendEvents(ctx context.Context, requestBody VariantsRequest) {
	endpoint := fmt.Sprintf("%s/variantsData", es.config.Services.EventsService.HttpEndpoint.Url)
	headers := map[string]string{
		CONTENT_TYPE:      APPLICATION_JSON,
		HEADER_AUTH_TYPE:  SDK,
		HEADER_SDK_ID:     es.sdkConfig.SDKId,
		HEADER_SDK_SECRET: es.sdkConfig.SDKSecret,
	}

	es.checkAndRefreshData(getTimeSlot(time.Now().Unix() * 1000))

	body, err := json.Marshal(requestBody)
	if err != nil {
		//es.logger.Errorf("error while marshalling variant request, error:%+v", err)
		return
	}
	_, err = flagsenseHttpClient.MakeHttpRequest(ctx, "POST", endpoint, es.client,
		bytes.NewBuffer(body), headers)
	if err != nil {
		//es.logger.Errorf("error while sending events, error:%+v", err)
		return
	}
}

func (es *EventServiceImpl) ShutdownHook(ctx context.Context) {
	if !es.config.Constants.CaptureEvents {
		return
	}
	es.refreshData(getTimeSlot(time.Now().Unix() * 1000))
	es.Run(ctx)
}

func getTimeSlot(time int64) int64 {
	return (time / MILLIS_IN_EVENT_FLUSH_INTERVAL) * MILLIS_IN_EVENT_FLUSH_INTERVAL
}
