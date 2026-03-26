package streams

import (
	"errors"
	"fmt"
	"path"
	"sync"
)

var registry NodeProviderRegistry
var nodes map[string]*nodeWrapper

func RegisterNodeProvider(class, subclass string, provider NodeProvider) error {
	return registry.Register(class, subclass, provider)
}

func init() {
	registry = &providerRegistry{
		providers: sync.Map{},
	}
	nodes = make(map[string]*nodeWrapper)
}

type providerRegistry struct {
	providers sync.Map
}

func (r *providerRegistry) Register(class, subclass string, provider NodeProvider) error {
	key, err := r.key(class, subclass)
	if err != nil {
		return err
	}
	r.providers.Store(key, provider)
	return nil
}

func (r *providerRegistry) Find(class, subclass string) (provider NodeProvider, err error) {
	if subclass == "*" {
		return nil, errors.New("subclass (*) is only valid while registering")
	}
	matches := make(map[string]NodeProvider)

	// Found a key that matches the class and subclass parameters
	r.providers.Range(func(key, value any) bool {

		var searchKey string
		searchKey, err = r.key(class, subclass)
		if err != nil {
			return false
		}
		matched, _ := path.Match(key.(string), searchKey)
		if matched {
			matches[key.(string)] = value.(NodeProvider)
		}
		return err != nil // Finalizes the loop if there is an error
	})
	if err == nil {
		if len(matches) == 0 {
			err = errors.New("provider not found")
		}
		// Select the NodeProvider whose key is the largest
		var key string
		for providerKey := range matches {
			if len(providerKey) > len(key) {
				key = providerKey
			}
		}
		provider = matches[key]
	}
	return
}

func (r *providerRegistry) key(class, subclass string) (string, error) {
	if class == "" {
		return "", errors.New("invalid class")
	}
	if subclass == "" {
		return "", errors.New("invalid subclass")
	}
	return fmt.Sprintf("%s/%s", class, subclass), nil
}
