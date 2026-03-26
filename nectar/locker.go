package nectar

import (
	"sync"
	"time"

	"github.com/maypok86/otter"
)

// SIGNAL LOCKER

type signalLocker struct {
	cache *otter.Cache[string, *SignalCachedState]
}

func NewSignalLocker() (SignalLocker, error) {
	cache, err := otter.MustBuilder[string, *SignalCachedState](10_000).
		CollectStats().
		Cost(func(key string, value *SignalCachedState) uint32 {
			return 1
		}).
		WithTTL(365 * 24 * time.Hour).
		Build()
	if err != nil {
		return nil, err
	}
	locker := &signalLocker{
		cache: &cache,
	}
	return locker, nil
}

func (l *signalLocker) Value(signalId string) (value any, found bool) {
	if state, found := l.cache.Get(signalId); found {
		return state.Value, false
	}
	return nil, found
}

func (l *signalLocker) Update(signalId string, value any) error {
	if state, found := l.cache.Get(signalId); found {
		return state.Update(value)
	}
	l.cache.Set(signalId, NewSignalState(signalId, value))
	return nil
}

func (l *signalLocker) TryUpdate(signalId string, value any) UpdateResult {
	if state, found := l.cache.Get(signalId); found {
		return state.TryUpdate(value)
	}
	// Create a new state if not found
	l.cache.Set(signalId, NewSignalState(signalId, value))
	return l.TryUpdate(signalId, value)
}

// SIGNAL STATE

const (
	LockDuration = 2 * time.Second
)

type UpdateResult uint32

const (
	Updated UpdateResult = iota // Signal is updated
	Cached                      // Signal is cached and not updated
	Locked                      // Signal is locked and cannot be updated
)

type SignalCachedState struct {
	Id        string
	Value     any
	Timestamp time.Time
	err       error
	mux       sync.Mutex
}

func NewSignalState(id string, value any) *SignalCachedState {
	return &SignalCachedState{
		Id:        id,
		Value:     value,
		Timestamp: time.Now(),
	}
}

func (s *SignalCachedState) Update(value any) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	// Check if the signal is failed
	if s.err != nil {
		return s.err
	}

	// Sets the value and timestamp
	s.Value = value
	// Set the timestamp to the current time plus the lock duration
	s.Timestamp = time.Now().Add(LockDuration)
	return nil
}

func (s *SignalCachedState) TryUpdate(value any) UpdateResult {
	s.mux.Lock()
	defer s.mux.Unlock()

	now := time.Now()

	if s.Timestamp.After(now) {
		// If the signal timestamp is not expired, return Locked state
		return Locked
	}
	if s.Value == value {
		// If the value is the same, update the timestamp and return Cached state
		s.Timestamp = now
		return Cached
	}
	// Update the value, timestamp and return Updated state
	s.Value = value
	s.Timestamp = now
	return Updated
}
