package model

import "github.com/flagsense/go-sdk/pkg/enums"

// TODO: It will be initialized once, one should never be able to initialize it again or change the attributes
type SDKConfig struct {
	SDKId       string
	SDKSecret   string
	Environment string
}

func NewSDKConfig(sdkId string, sdkSecret string, env *enums.Environment) *SDKConfig {
	return &SDKConfig{
		SDKId:       sdkId,
		SDKSecret:   sdkSecret,
		Environment: env.Name,
	}
}
