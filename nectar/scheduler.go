package nectar

import (
	"sync"
	"time"
)

type scheduler struct {
	next      chan string
	stop      chan struct{}
	intervals []IntervalOptions
	closeOnce sync.Once
	mu        sync.RWMutex // Add mutex for thread safety
	stopped   bool         // Add flag to track stopped state
}

func NewScheduler() Scheduler {
	return &scheduler{
		next:      make(chan string),
		stop:      make(chan struct{}),
		intervals: []IntervalOptions{},
		stopped:   false,
	}
}

func (s *scheduler) Schedule(options SchedulerOptions) {
	s.next = make(chan string, options.QueueCapacity)
	s.stop = make(chan struct{})
	s.intervals = options.Intervals
}

func (s *scheduler) Next() <-chan string {
	return s.next
}

func (s *scheduler) Start(onScheduled func(string)) {
	go func() {
		for {
			select {
			case route, ok := <-s.Next():
				if !ok {
					// log.Info("Channel closed, stopping scheduler")
					return // Exit the goroutine when channel is closed
				}
				// log.Infof("Scheduler elapsed: %s", route)
				onScheduled(route)
			}
		}
	}()

	for _, interval := range s.intervals {

		// s.next <- interval.Route // Send the first signal immediately...

		go func(interval IntervalOptions) {
			ticker := time.NewTicker(time.Duration(interval.Interval) * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					// Check if scheduler is stopped before sending
					s.mu.RLock()
					isStopped := s.stopped
					s.mu.RUnlock()

					if !isStopped {
						select {
						case s.next <- interval.Route:
							// Successfully sent
						case <-s.stop:
							return
						}
					}
				case <-s.stop:
					return
				}
			}
		}(interval)
	}
}

func (s *scheduler) Stop() {
	// Mark as stopped first
	s.mu.Lock()
	s.stopped = true
	s.mu.Unlock()

	close(s.stop)
	s.closeOnce.Do(func() {
		close(s.next)
	})
}
