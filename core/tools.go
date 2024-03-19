package core

import (
	"strings"
)

const (
	// SharedMemoryKeyPrefix is the prefix for shared memory keys
	SharedMemoryKeyPrefix string = "shared://"
	JsonPathKeyPrefix     string = "jsonpath://"
)

func (v Variables) GetSharedKeys() []string {
	return v.getKeys(SharedMemoryKeyPrefix)
}
func (v Variables) GetSharedVariables() Variables {
	return v.getVariables(SharedMemoryKeyPrefix)
}

func (v Variables) GetJsonPathKeys() []string {
	return v.getKeys(JsonPathKeyPrefix)
}
func (v Variables) GetJsonPathVariables() Variables {
	return v.getVariables(JsonPathKeyPrefix)
}

func (v Variables) getVariables(prefix string) Variables {
	keys := v.getKeys(prefix)
	vars := make(Variables)
	for _, key := range keys {
		vars[key] = strings.TrimPrefix(v[key], prefix)
	}
	return vars
}

func (v Variables) getKeys(prefix string) []string {
	keys := []string{}
	for key, value := range v {
		if strings.HasPrefix(value, prefix) {
			keys = append(keys, key)
		}
	}
	return keys
}
