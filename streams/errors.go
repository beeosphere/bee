package streams

import (
	"errors"
	"fmt"
)

var (
	ErrInputNotFound = errors.New("input not found")
)

func ErrNodeNotFound(nodeId string) error {
	return errors.New(fmt.Sprintf("node '%s' not found: ", nodeId))
}
