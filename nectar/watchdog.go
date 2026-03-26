package nectar

import (
	"sync"
	"time"

	"github.com/maypok86/otter"
)

// ROUTES WATCHDOG

const (
	WatchdogTime = 2 * time.Second
)

type routeWatchdog struct {
	cache          *otter.Cache[string, *RouteState]
	ticker         *time.Ticker
	done           chan struct{}
	onMissedRoutes func(routes []string)
}

func NewRouteWatchdog(routesToGuard []string) (RouteWatchdog, error) {
	// Create the cache of routes
	cache, err := otter.MustBuilder[string, *RouteState](10_000).
		// CollectStats().
		Cost(func(key string, value *RouteState) uint32 {
			return 1
		}).
		WithTTL(365 * 24 * time.Hour).
		Build()
	if err != nil {
		return nil, err
	}
	// Initialize the cache with the routes to guard
	for _, route := range routesToGuard {
		cache.Set(route, NewRouteState(route))
	}
	return &routeWatchdog{
		cache: &cache,
	}, nil
}

func (w *routeWatchdog) Start() {
	if w.ticker == nil {
		w.ticker = time.NewTicker(WatchdogTime)
		w.done = make(chan struct{})

		// Start a goroutine to handle the ticker
		go func() {
			for {
				select {
				case <-w.done:
					return
				case <-w.ticker.C:
					w.checkRoutes()
				}
			}
		}()
	}
}

func (w *routeWatchdog) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
		w.ticker = nil
		close(w.done)
	}
}

func (w *routeWatchdog) RouteAlive(route string) {
	if state, found := w.cache.Get(route); found {
		state.Heartbeat()
	} // else {
	// 	w.cache.Set(route, NewRouteState(route))
	// }
}

func (w *routeWatchdog) OnMissedRoutes(fn func(routes []string)) {
	w.onMissedRoutes = fn
}

func (w *routeWatchdog) checkRoutes() {
	now := time.Now()
	missingRoutes := make([]string, 0)

	// Iterate over the cache and check each route
	// If the route is not alive, add it to the missing routes slice
	w.cache.Range(func(key string, value *RouteState) bool {
		// Check if the route is alive
		alive, downCount := value.IsAlive(now)
		if !alive && downCount == 1 {
			// If the route is not alive, add it to the missing routes slice only once
			// This prevents multiple notifications for the same route
			missingRoutes = append(missingRoutes, key)
		}
		return true // continue iterating
	})
	// If there are any routes that are not alive, notify it
	if len(missingRoutes) > 0 && w.onMissedRoutes != nil {
		w.onMissedRoutes(missingRoutes)
	}
}

type RouteState struct {
	Id        string
	Timestamp time.Time
	count     uint32
	err       error
	mux       sync.Mutex
}

func NewRouteState(id string) *RouteState {
	return &RouteState{
		Id:        id,
		Timestamp: time.Now(),
		count:     0,
	}
}

func (s *RouteState) Heartbeat() {
	s.mux.Lock()
	defer s.mux.Unlock()

	// Set the timestamp to the current time
	s.Timestamp = time.Now()
	s.count = 0
}

func (s *RouteState) IsAlive(checkTime time.Time) (alive bool, downCount uint32) {
	s.mux.Lock()
	defer s.mux.Unlock()

	// Timestamp should not be older than AliveTime from the current time
	if s.Timestamp.Add(WatchdogTime).After(checkTime) {
		return true, 0
	}
	// If the timestamp is older, increment the counter and return false
	s.count++
	return false, s.count
}
