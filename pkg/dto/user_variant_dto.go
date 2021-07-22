package dto

type UserVariantDTO struct {
	UserId              string
	Attributes          map[string]interface{}
	FlagId              string
	DefaultKey          string
	DefaultValue        interface{}
	Key                 string
	Value               interface{}
	ExpectedVariantType string
}
