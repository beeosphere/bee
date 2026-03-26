package core

import (
	"encoding/json"
	"fmt"
)

func Serialize(v any) (bytes []byte, err error) {
	bytes, ok := v.([]byte)
	if ok {
		return bytes, nil
	}
	// bytes, err = json.Marshal(v)
	bytes, err = json.MarshalIndent(v, "", "  ")
	if err != nil {
		err = fmt.Errorf("failed to serialize data: %v", err)
	}
	return
}

func Deserialize(data []byte, v any) error {
	// bytes, ok := v.([]byte)
	// if ok {
	// 	v = bytes
	// 	return nil
	// }
	err := json.Unmarshal(data, &v)
	if err != nil {
		return fmt.Errorf("failed to deserialize data: %v", err)
	}
	return nil
}

func DirectDeserialize[T any](data any) *T {
	var result T
	err := Deserialize(data.([]byte), &result)
	if err != nil {
		return nil
	}
	return &result
}
