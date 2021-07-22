package enums

import (
	"flagsense-go-sdk/constants"
	"flagsense-go-sdk/pkg/util"
)

type Environment struct {
	Name string
}

func NewEnvironment(name string) *Environment {
	return &Environment{Name: name}
}

func (e Environment) IsValid(env string) bool {
	return util.ContainsString(constants.Environments, env)
}

func (e Environment) Equals(env string, name string) bool {
	return env == name
}
