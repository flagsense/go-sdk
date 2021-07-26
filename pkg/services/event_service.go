package services

import (
	"context"
)

type EventService interface {
	Start(ctx context.Context)
	AddEvaluationCount(flagId string, variantKey string)
	AddErrorsCount(flagId string)
	ShutdownHook(ctx context.Context)
	AddCodeBugsCount(flagId string, variantKey string)
}
