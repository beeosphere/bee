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
}

func NewScheduler() Scheduler {
	return &scheduler{
		next:      make(chan string),
		stop:      make(chan struct{}),
		intervals: []IntervalOptions{},
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
					s.next <- interval.Route
				case <-s.stop:
					return
				}
			}
		}(interval)
	}
}

func (s *scheduler) Stop() {
	close(s.stop)
	s.closeOnce.Do(func() {
		close(s.next)
	})
}
