package services

import "github.com/flagsense/go-sdk/pkg/dto"

type UserVariantService interface {
	GetUserVariant(userVariantDTO *dto.UserVariantDTO) error
}
