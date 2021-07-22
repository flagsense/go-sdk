package client

import (
	"errors"
	"github.com/flagsense/go-sdk/constants"
	"github.com/flagsense/go-sdk/pkg/enums"
	"github.com/flagsense/go-sdk/pkg/model"
	"github.com/flagsense/go-sdk/pkg/services"
	"github.com/flagsense/go-sdk/pkg/services/impl"
	"strings"
	"sync"
)

var flagsenseServiceMap = make(map[string]services.FlagsenseService)
var lock = sync.Mutex{}
var shutdownInProgress = false

func CreateService(sdkId string, sdkSecret string, env string) (error, services.FlagsenseService) {
	if strings.TrimSpace(sdkId) == "" || strings.TrimSpace(sdkSecret) == "" {
		return errors.New("empty sdk params not allowed"), nil
	}

	lock.Lock()
	defer lock.Unlock()
	if flagsenseServiceMap[sdkId] != nil {
		fs, _ := flagsenseServiceMap[sdkId]
		return nil, fs
	}
	if !enums.NewEnvironment(env).IsValid(env) {
		env = constants.PROD
	}
	flagsenseServiceMap[sdkId] = impl.NewFlagsenseService(sdkId, sdkSecret, enums.NewEnvironment(env))
	fs, _ := flagsenseServiceMap[sdkId]
	return nil, fs
}

func User(userId string, attributes map[string]interface{}) model.FSUser {
	return model.FSUser{
		userId,
		attributes,
	}
}

func BooleanFlag(flagId string, defaultKey string, defaultValue bool) model.FSFlag {
	return model.FSFlag{
		flagId,
		defaultKey,
		defaultValue,
	}
}

func IntegerFlag(flagId string, defaultKey string, defaultValue int32) model.FSFlag {
	return model.FSFlag{
		flagId,
		defaultKey,
		defaultValue,
	}
}

func DecimalFlag(flagId string, defaultKey string, defaultValue float64) model.FSFlag {
	return model.FSFlag{
		flagId,
		defaultKey,
		defaultValue,
	}
}

func StringFlag(flagId string, defaultKey string, defaultValue string) model.FSFlag {
	return model.FSFlag{
		flagId,
		defaultKey,
		defaultValue,
	}
}

func MapFlag(flagId string, defaultKey string, defaultValue map[string]interface{}) model.FSFlag {
	return model.FSFlag{
		flagId,
		defaultKey,
		defaultValue,
	}
}

func Close()  {
	if shutdownInProgress {
		return
	}
	lock.Lock()
	defer lock.Unlock()

	shutdownInProgress = true
	for _, service := range flagsenseServiceMap {
		service.Close()
	}
}
