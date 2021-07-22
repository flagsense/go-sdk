package services

import "flagsense-go-sdk/pkg/dto"

type UserVariantService interface {
	GetUserVariant(userVariantDTO *dto.UserVariantDTO) error
}
