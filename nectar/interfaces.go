package nectar

import (
	"context"
	"time"
)

// DRIVER

type Driver interface {
	Connect(ctx context.Context)
	Disconnect()
	Dispose()
	ReadSignals(routeId string) ([]*Signal, error)
	WriteSignals(signals []*Signal) error
	SignalsReceived() <-chan *SignalsPack
}

// SCHEDULER

type SchedulerOptions struct {
	QueueCapacity int
	Intervals     []IntervalOptions
}

type IntervalOptions struct {
	Interval int // in milliseconds
	Route    string
}

type Scheduler interface {
	Start(onScheduled func(string))
	Stop()
	Next() <-chan string
	Schedule(options SchedulerOptions)
}

// PROCESSOR

type RouteInfo[Mapping, SignalMetadata any] struct {
	RouteId   string
	Pattern   string
	IsEmitter bool
	Mapping   *Mapping
	Signals   []*SignalMetadata
}

type JobType int

const (
	Write JobType = iota
	Read
)

func (jt JobType) String() string {
	return [...]string{"Write", "Read"}[jt]
}

type SignalsJob struct {
	Route    string
	Type     JobType
	Sequence uint64
	Signals  []*Signal
}

type ProcessorState int

const (
	Idle ProcessorState = iota
	Executing
	Failed
)

type ConnectionState int

const (
	Disconnected ConnectionState = iota
	Connected
)

type SignalsPack struct {
	RouteId string
	Signals []*Signal
}

func NewSignalsPack(routeId string, signals []*Signal) *SignalsPack {
	return &SignalsPack{
		RouteId: routeId,
		Signals: signals,
	}
}

func (sp *SignalsPack) HasSignals() bool {
	return len(sp.Signals) > 0
}

type RuntimeSettings[Settings, Mapping, SignalMetadata any] struct {
	Settings   *Settings
	RoutesInfo []*RouteInfo[Mapping, SignalMetadata]
}

type Processor[Settings, Mapping, SignalMetadata any] interface {
	IsRunning() bool
	Run(cfg *RuntimeSettings[Settings, Mapping, SignalMetadata])
	Stop()
	SetScheduler(scheduler Scheduler)
	DispatchSignals(pack *SignalsPack)
	OnSignalsCollected(fn func(pack *SignalsPack))
}

// CACHE

type Cache interface {
	Get(key string) (value any, found bool)
	Set(key string, value any)
	SetAndLock(key string, value any, ttl time.Duration)
	SetIfNotLocked(key string, value any) bool
	Lock(key string, ttl time.Duration)
	Unlock(key string)
	Delete(key string)
	Clear()
}

// REPORTER

// TODO: Define methods and functionality to be reported
type Reporter interface {
	AddOperation(route string, operation string, signalsQty int, start time.Time)
	AddDiscardedOperation(route string, operation string)
	Report()
	AddError(err error)
}
