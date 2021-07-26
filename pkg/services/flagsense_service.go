package services

import "github.com/flagsense/go-sdk/pkg/model"

type FlagsenseService interface {
	InitializationComplete() bool
	WaitForInitializationComplete()
	Close()
	BooleanVariation(fsFlag model.FSFlag, user model.FSUser) model.FSVariation
	StringVariation(fsFlag model.FSFlag, user model.FSUser) model.FSVariation
	IntegerVariation(fsFlag model.FSFlag, user model.FSUser) model.FSVariation
	DecimalVariation(fsFlag model.FSFlag, user model.FSUser) model.FSVariation
	MapVariation(fsFlag model.FSFlag, user model.FSUser) model.FSVariation
	RecordCodeError(flagId string, variationKey string)
}
