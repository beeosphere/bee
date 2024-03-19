package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/samber/lo"
)

type SharedMemory interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Watch(keys []string, renew bool)
	Unwatch(key string)
	UnwatchAll()
	Watcher() <-chan *MemoryValue
}

type ResourceManager interface {
	SharedMemory
}

type resourceManager struct {
	sync.Mutex

	httpClient   *HttpClient
	protocol     *Protocol
	agentId      string
	memory       map[string]string
	watcher      chan *MemoryValue
	watchingKeys []string
}

func newResourceManager(agentId string, httpClient *HttpClient) *resourceManager {
	return &resourceManager{
		httpClient:   httpClient,
		protocol:     NewProtocol(),
		agentId:      agentId,
		memory:       make(map[string]string),
		watcher:      nil,
		watchingKeys: []string{},
	}
}

func (r *resourceManager) Run() error {

	// Protocol initialization
	proto := r.protocol

	proto.On("$AGENT.ID."+r.agentId+".MEMORY_CHANGED", func(proto *Protocol, message ProtoMessage) {
		r.memoryChanged(ProtoMap[MemoryValue](message))
	})
	return nil
}

func (r *resourceManager) Stop() error {
	if r.watcher != nil {
		close(r.watcher)
	}
	return r.protocol.Dispose()
}

func (r *resourceManager) Set(key, value string) error {
	// HTTP client call to set key-value pair in shared memory
	memory := &MemoryValue{Key: key, Value: value, Timestamp: time.Now()}
	if err := r.httpClient.Post(context.TODO(), "api/memory", memory, nil); err != nil {
		fmt.Println("Error setting shared memory key: ", err)
		return err
	}
	// Update local memory
	r.setValue(key, value)
	return nil
}

func (r *resourceManager) Get(key string) (string, error) {
	if value, exists := r.getValue(key); exists {
		return value, nil
	}
	// If the key does not exist, return value from server shared memory

	// Call HttpClient to get value from server shared memory
	var memory MemoryValue
	if err := r.httpClient.Get(context.TODO(), fmt.Sprintf("api/memory/%s", key), &memory); err != nil {
		fmt.Println("Error fetching shared memory key: ", err)
		return "", err
	}
	// Update local memory
	r.setValue(memory.Key, memory.Value)

	return memory.Value, nil
}
func (r *resourceManager) Watcher() <-chan *MemoryValue {
	if r.watcher == nil {
		r.watcher = make(chan *MemoryValue, 10)
	}
	return r.watcher
}

func (r *resourceManager) Watch(keys []string, renew bool) {
	if renew {
		r.UnwatchAll()
	}
	for _, key := range keys {
		r.watchMemory(key)
	}
}

func (r *resourceManager) Unwatch(key string) {
	// r.Lock()
	// defer r.Unlock()
	r.watchingKeys = lo.Filter(r.watchingKeys, func(k string, _ int) bool {
		return k != key
	})
}

func (r *resourceManager) UnwatchAll() {
	// r.Lock()
	// defer r.Unlock()
	r.watchingKeys = []string{}
}

func (r *resourceManager) watchMemory(key string) {
	// r.Lock()
	// defer r.Unlock()
	r.watchingKeys = lo.Uniq(append(r.watchingKeys, key))

	// Call HttpClient to get value from server shared memory
	var memory MemoryValue
	if err := r.httpClient.Get(context.TODO(), fmt.Sprintf("api/memory/%s?subscribe=%s", key, r.agentId), &memory); err != nil {
		fmt.Println("Error fetching shared memory key: ", err)
	}
	// Update local memory
	r.setValue(memory.Key, memory.Value)

	// handler(&memory)
	if r.watcher != nil {
		r.watcher <- &memory
	}
}

func (r *resourceManager) memoryChanged(data *MemoryValue) {
	// r.Lock()
	// defer r.Unlock()

	r.setValue(data.Key, data.Value)

	// // Notify subscribers of memory change...
	if r.watcher != nil && lo.Some(r.watchingKeys, []string{data.Key}) {
		r.watcher <- data
	}
}

func (r *resourceManager) setValue(key, value string) {
	r.Lock()
	defer r.Unlock()
	r.memory[key] = value
}

func (r *resourceManager) getValue(key string) (value string, exists bool) {
	r.Lock()
	defer r.Unlock()
	value, exists = r.memory[key]
	return
}
