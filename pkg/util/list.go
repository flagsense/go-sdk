package util

import (
	"fmt"
	"github.com/hashicorp/go-version"
)

func Contains(source []interface{}, search interface{}) bool {
	for _, find := range source {
		if fmt.Sprintf("%T", find) != fmt.Sprintf("%T", search) {
			return false
		}

		switch search.(type) {
		case int:
			if search.(int) == find.(int) {
				return true
			}
		case float64:
			if search.(float64) == find.(float64) {
				return true
			}
		case float32:
			if search.(float32) == find.(float32) {
				return true
			}
		case int64:
			if search.(int64) == find.(int64) {
				return true
			}
		case int32:
			if search.(int32) == find.(int32) {
				return true
			}
		case string:
			if search.(string) == find.(string) {
				return true
			}
		}
	}
	return false
}

func ContainsString(source []string, search string) bool {
	for _, find := range source {
		if find == search {
			return true
		}
	}
	return false
}

func ContainsVersion(source []interface{}, search *version.Version) bool {
	for _, versionString := range source {
		valueVersion, err := version.NewVersion(versionString.(string))
		if err != nil {
			return false
		}
		if valueVersion.Equal(search) {
			return true
		}
	}
	return false
}
