package services

import "context"

type DataPollerService interface {
	Start(ctx context.Context)
	WaitForInitializationComplete()
}
