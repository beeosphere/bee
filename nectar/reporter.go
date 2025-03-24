package nectar

import (
	"fmt"
	"sync"
	"time"

	"github.com/beeosphere/bee/core"
)

type Stats struct {
	Operation  string
	signalsQty int
	triggers   []time.Time
	durations  []time.Duration
	discarded  int
}

type reporter struct {
	log     core.Logger
	ops     sync.Map
	counter uint64
}

func NewReporter(logger core.Logger) Reporter {
	return &reporter{
		log: logger,
		ops: sync.Map{},
	}
}

func (r *reporter) AddOperation(route string, operation string, signalsQty int, start time.Time) {
	elapsed := time.Since(start)
	stats, _ := r.ops.LoadOrStore(route, &Stats{Operation: operation, signalsQty: signalsQty})
	s := stats.(*Stats)
	s.durations = append(s.durations, elapsed)
	s.triggers = append(s.triggers, start)
}

func (r *reporter) AddDiscardedOperation(route string, operation string) {
	stats, _ := r.ops.LoadOrStore(route, &Stats{Operation: operation})
	s := stats.(*Stats)
	s.discarded++
}

func (r *reporter) Report() {
	r.counter++
	if r.counter%60 != 0 {
		return
	}

	var result string
	r.ops.Range(func(key, value interface{}) bool {
		route := key.(string)
		stats := value.(*Stats)
		var avgDuration float64
		var avgTriggers float64
		if len(stats.durations) > 0 {
			var total time.Duration
			for _, d := range stats.durations {
				total += d
			}
			if len(stats.durations) > 1 {
				avgDuration = float64(total.Milliseconds()) / float64(len(stats.durations))
			} else {
				avgDuration = float64(total.Milliseconds())
			}
			for i := 1; i < len(stats.triggers); i++ {
				avgTriggers += float64(stats.triggers[i].Sub(stats.triggers[i-1]).Milliseconds())
			}
			avgTriggers /= float64(len(stats.triggers) - 1)
		}
		result += fmt.Sprintf("(route: %s) %s %d signals %d times every %.2f ms (%.2f ms/op)\n",
			route, stats.Operation, stats.signalsQty, r.counter, avgTriggers, avgDuration)
		return true
	})
	if result != "" {
		r.log.Debug(result)
	}
}

func (r *reporter) AddError(err error) {
}
