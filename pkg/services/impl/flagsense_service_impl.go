package impl

import (
	"context"
	"errors"
	"flagsense-go-sdk/config"
	"flagsense-go-sdk/pkg/dto"
	"flagsense-go-sdk/pkg/enums"
	"flagsense-go-sdk/pkg/infrastructure/http"
	"flagsense-go-sdk/pkg/model"
	"flagsense-go-sdk/pkg/services"
	"flagsense-go-sdk/third_party/assetmnger"
	"flagsense-go-sdk/third_party/logger"
	teltech "github.com/teltech/logger"
	"strings"
	"time"
)

type FlagsenseServiceImpl struct {
	SDKConfig          *model.SDKConfig
	Data               *dto.Data
	DataPollerService  services.DataPollerService
	UserVariantService services.UserVariantService
	EventService       services.EventService
	logger             *teltech.Log
}

const (
	ZER0    = 0
	DEFAULT = "default"
)

func NewFlagsenseService(sdkId string, sdkSecret string, environment *enums.Environment) *FlagsenseServiceImpl {
	// ---------------------  Initialize drivers  --------------------- //
	manager := assetmnger.NewManager()
	store := config.NewConfig(manager)
	sdkConfig := model.NewSDKConfig(sdkId, sdkSecret, environment)
	ctx := context.Background()
	log := logger.NewLogger()
	httpClient := httptrp.NewFlagSenseHttpClient()

	// ---------------------  Initialize Data  --------------------- //
	data := dto.Data{
		LastUpdatedOn: ZER0,
		Segments:      nil,
		Flags:         nil,
	}

	// ---------------------  Initialize Poller  --------------------- //
	poller := NewDataPollerService(
		sdkConfig, time.Duration(store.Constants.PollingInterval)*time.Minute, log, store, httpClient, &data)
	go poller.Start(ctx)

	// --------------------- Init User Variant ---------------------//
	userVariantService := NewUserVariantService(&data, log)

	// --------------------- Init Events Service ---------------------//
	eventsService := NewEventService(log, sdkConfig, httpClient, store)
	go eventsService.Start(ctx)

	flagsense := &FlagsenseServiceImpl{
		SDKConfig:          sdkConfig,
		Data:               &data,
		DataPollerService:  poller,
		UserVariantService: userVariantService,
		EventService:       eventsService,
		logger:             log,
	}
	return flagsense
}

func (fs *FlagsenseServiceImpl) WaitForInitializationComplete() {
	fs.DataPollerService.WaitForInitializationComplete()
}

func (fs *FlagsenseServiceImpl) InitializationComplete() bool {
	return fs.Data.LastUpdatedOn > ZER0
}

func (fs *FlagsenseServiceImpl) Close()  {
	ctx, cancel := context.WithCancel(context.Background())
	fs.EventService.ShutdownHook(ctx)
	cancel()
}

func (fs *FlagsenseServiceImpl) __evaluate(variantDTO *dto.UserVariantDTO) {
	var err error
	if fs.Data.LastUpdatedOn == 0 {
		err = errors.New("flag data is still loading, evaluation not called")
	} else {
		err = fs.UserVariantService.GetUserVariant(variantDTO)
	}
	if err != nil {
		//fs.logger.Errorf("Error in evaluation: %+v", err)
		variantDTO.Key = variantDTO.DefaultKey
		variantDTO.Value = variantDTO.DefaultValue

		defaultKey := DEFAULT
		if strings.TrimSpace(variantDTO.Key) == "" {
			defaultKey = variantDTO.Key
		}
		fs.EventService.AddEvaluationCount(variantDTO.FlagId, defaultKey)
		fs.EventService.AddErrorsCount(variantDTO.FlagId)
		return
	}
	fs.EventService.AddEvaluationCount(variantDTO.FlagId, variantDTO.Key)
}

func (fs *FlagsenseServiceImpl) _evaluate(fsFlag model.FSFlag, user model.FSUser, expectedVariantType string) model.FSVariation {
	userVariantDTO := dto.UserVariantDTO{
		FlagId:              fsFlag.FlagId,
		UserId:              user.UserId,
		Attributes:          user.Attributes,
		DefaultValue:        fsFlag.DefaultValue,
		DefaultKey:          fsFlag.DefaultKey,
		ExpectedVariantType: expectedVariantType,
	}
	fs.__evaluate(&userVariantDTO)
	return model.FSVariation{
		userVariantDTO.Key,
		userVariantDTO.Value,
	}
}

func (fs *FlagsenseServiceImpl) evaluateAndSetVariation(fsFlag model.FSFlag, user model.FSUser, expectedVariantType string, result *model.FSVariation) {
	var err error
	var variation model.FSVariation

	defer func() { //catch or finally
		if panicErr := recover(); panicErr != nil { //catch
			//fs.logger.Errorf("Panic: %v", panicErr)
			fs.EventService.AddEvaluationCount(fsFlag.FlagId, fsFlag.DefaultKey)
			fs.EventService.AddErrorsCount(fsFlag.FlagId)
		}
	}()

	variation = fs._evaluate(fsFlag, user, expectedVariantType)

	switch expectedVariantType {
	case dto.BOOL:
		variation, err = variation.ToBoolean()
		break
	case dto.INT:
		variation, err = variation.ToInteger()
		break
	case dto.DOUBLE:
		variation, err = variation.ToDouble()
		break
	case dto.STRING:
		variation, err = variation.ToString()
		break
	case dto.JSON:
		variation, err = variation.ToMap()
		break
	}

	if err == nil {
		result.Key = variation.Key
		result.Value = variation.Value
	}
}

func (fs *FlagsenseServiceImpl) BooleanVariation(fsFlag model.FSFlag, user model.FSUser) model.FSVariation {
	var result *model.FSVariation
	result = &model.FSVariation{
		Key: fsFlag.DefaultKey,
		Value: fsFlag.DefaultValue,
	}
	fs.evaluateAndSetVariation(fsFlag, user, dto.BOOL, result)
	return *result
}

func (fs *FlagsenseServiceImpl) StringVariation(fsFlag model.FSFlag, user model.FSUser) model.FSVariation {
	var result *model.FSVariation
	result = &model.FSVariation{
		Key: fsFlag.DefaultKey,
		Value: fsFlag.DefaultValue,
	}
	fs.evaluateAndSetVariation(fsFlag, user, dto.STRING, result)
	return *result
}

func (fs *FlagsenseServiceImpl) IntegerVariation(fsFlag model.FSFlag, user model.FSUser) model.FSVariation {
	var result *model.FSVariation
	result = &model.FSVariation{
		Key: fsFlag.DefaultKey,
		Value: fsFlag.DefaultValue,
	}
	fs.evaluateAndSetVariation(fsFlag, user, dto.INT, result)
	return *result
}

func (fs *FlagsenseServiceImpl) DecimalVariation(fsFlag model.FSFlag, user model.FSUser) model.FSVariation {
	var result *model.FSVariation
	result = &model.FSVariation{
		Key: fsFlag.DefaultKey,
		Value: fsFlag.DefaultValue,
	}
	fs.evaluateAndSetVariation(fsFlag, user, dto.DOUBLE, result)
	return *result
}

func (fs *FlagsenseServiceImpl) MapVariation(fsFlag model.FSFlag, user model.FSUser) model.FSVariation {
	var result *model.FSVariation
	result = &model.FSVariation{
		Key: fsFlag.DefaultKey,
		Value: fsFlag.DefaultValue,
	}
	fs.evaluateAndSetVariation(fsFlag, user, dto.JSON, result)
	return *result
}
