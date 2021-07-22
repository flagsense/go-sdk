package model

import (
	"errors"
	"github.com/flagsense/go-sdk/pkg/util"
	"fmt"
	"strings"
)

type FSVariation struct {
	Key   string
	Value interface{}
}

func (fsv *FSVariation) ToBoolean() (FSVariation, error) {
	if fmt.Sprintf("%T", fsv.Value) != "bool" {
		return FSVariation{}, errors.New(fmt.Sprintf("Value is not boolean, %+v", fsv.Value))
	}
	return FSVariation{
		Key:   fsv.Key,
		Value: fsv.Value.(bool),
	}, nil
}

func (fsv *FSVariation) ToInteger() (FSVariation, error) {
	valType := fmt.Sprintf("%T", fsv.Value)
	if !strings.Contains(valType, "int") && !strings.Contains(valType, "float") {
		return FSVariation{}, errors.New(fmt.Sprintf("Value is not int32, %+v", fsv.Value))
	}
	return FSVariation{
		Key:   fsv.Key,
		Value: util.ConvertToInt32(fsv.Value),
	}, nil
}

func (fsv *FSVariation) ToDouble() (FSVariation, error) {
	valType := fmt.Sprintf("%T", fsv.Value)
	if !strings.Contains(valType, "int") && !strings.Contains(valType, "float") {
		return FSVariation{}, errors.New(fmt.Sprintf("Value is not float64, %+v", fsv.Value))
	}
	return FSVariation{
		Key:   fsv.Key,
		Value: util.ConvertToFloat64(fsv.Value),
	}, nil
}

func (fsv *FSVariation) ToString() (FSVariation, error) {
	if fmt.Sprintf("%T", fsv.Value) != "string" {
		return FSVariation{}, errors.New(fmt.Sprintf("Value is not boolean, %+v", fsv.Value))
	}
	return FSVariation{
		Key:   fsv.Key,
		Value: fsv.Value.(string),
	}, nil
}

func (fsv *FSVariation) ToMap() (FSVariation, error) {
	if fmt.Sprintf("%T", fsv.Value) != "map[string]interface {}" {
		return FSVariation{}, errors.New(fmt.Sprintf("Value is not boolean, %+v", fsv.Value))
	}
	return FSVariation{
		Key:   fsv.Key,
		Value: fsv.Value.(map[string]interface{}),
	}, nil
}
