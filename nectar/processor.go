package nectar

import (
	"context"
	"sync"
	"time"

	"github.com/beeosphere/bee/core"
)

type processor[S, M, SigMeta any] struct {
	log         core.Logger
	driver      Driver
	scheduler   Scheduler
	reporter    Reporter
	runCancel   context.CancelFunc
	routeStates map[string]ProcessorState
	connState   ConnectionState
	routeLock   sync.RWMutex
	connLock    sync.RWMutex
	routeGroup  sync.WaitGroup
	jobChan     chan *SignalsJob
	collectChan chan *SignalsPack
	collectCb   func(data *SignalsPack)
	writeLock   sync.Mutex
}

func NewProcessor[S, M, SigMeta any](logger core.Logger, driver Driver, reporter Reporter) Processor[S, M, SigMeta] {
	return &processor[S, M, SigMeta]{
		log:         logger,
		driver:      driver,
		reporter:    reporter,
		jobChan:     make(chan *SignalsJob),
		collectChan: make(chan *SignalsPack),
		connState:   Disconnected,
	}
}

func (p *processor[S, M, SigMap]) IsRunning() bool {
	return p.ConnState() != Disconnected
}

func (p *processor[S, M, SigMeta]) SetScheduler(scheduler Scheduler) {
	p.scheduler = scheduler
}

func (p *processor[S, M, SigMeta]) Run(cfg *RuntimeSettings[S, M, SigMeta]) {
	// Initialize route states at Idle state
	p.routeStates = make(map[string]ProcessorState)
	for _, route := range cfg.RoutesInfo {
		p.routeStates[route.RouteId] = Idle
	}

	// Start processing signals
	ctx, cancel := context.WithCancel(context.Background())
	p.runCancel = cancel

	go p.doProcess(ctx)

	// Start the scheduler if it is available
	if p.scheduler != nil {
		p.scheduler.Start(func(route string) {
			if p.ConnState() == Connected {
				p.jobChan <- &SignalsJob{
					Route: route,
					Type:  Read,
				}
			}
		})
	}
}

func (p *processor[S, M, SigMeta]) Stop() {
	if p.scheduler != nil {
		p.scheduler.Stop()
	}
	p.setConnState(Disconnected)
	p.runCancel()

	// Wait for all running jobs to complete
	p.routeGroup.Wait()
}

func (p *processor[S, M, SigMeta]) DispatchSignals(pack *SignalsPack) {
	job := &SignalsJob{
		Type:    Write,
		Route:   pack.RouteId,
		Signals: pack.Signals,
	}
	p.jobChan <- job
}

func (p *processor[S, M, SigMeta]) OnSignalsCollected(fn func(pack *SignalsPack)) {
	p.collectCb = fn
}

func (p *processor[S, M, SigMeta]) doProcess(ctx context.Context) {
	defer p.disconnect() // Ensure disconnection when exiting
	for {

		// Checking state to know if it should connect or not...
		if p.ConnState() == Disconnected {
			if err := p.connect(ctx); err != nil { // It should try to connect until it succeeds
				return
			}
		}

		select {
		case job := <-p.jobChan:
			// if p.RouteState(job.Route) == Executing {
			// 	// TODO: Report it, discard or queue the job and continue
			// 	p.reporter.AddDiscardedOperation(job.Route, job.Type.String())
			// 	continue
			// }
			p.setRouteState(job.Route, Executing)
			go p.executeJob(job)
		case <-ctx.Done():
			// Exit the goroutine when context is done (processor is stopped).
			// This will also disconnect from the PLC because of the defer
			return
		}

		// Control the disconnection based on any route failed state
		if p.AnyFailedState() {
			p.disconnect()
		}
	}
}

func (p *processor[S, M, SigMeta]) connect(ctx context.Context) error {
	//  TODO: It should try to connect until it succeeds
	p.driver.Connect(ctx)

	p.setConnState(Connected)
	return nil
}

func (p *processor[S, M, SigMeta]) disconnect() {
	// TODO: It should work whatever the state is not Started
	p.driver.Disconnect()

	p.setConnState(Disconnected)
}

func (p *processor[S, M, SigMeta]) executeJob(job *SignalsJob) {
	// p.setRouteState(job.Route, Executing)
	p.routeGroup.Add(1)
	defer p.routeGroup.Done()

	// Execute the job
	if err := p.doExecute(job); err != nil {

		p.reporter.AddError(err)
		p.setRouteState(job.Route, Failed)
		return
	}
	p.setRouteState(job.Route, Idle)
}

func (p *processor[S, M, SigMeta]) doExecute(job *SignalsJob) error {
	if job.Type == Read && p.collectCb != nil {
		start := time.Now()
		signals, err := p.driver.ReadSignals(job.Route)

		if err != nil {
			elapsed := time.Since(start)
			p.log.Errorf("ReadSignals failed for route %s after %v: %v\n", job.Route, elapsed, err)
			return err
		}
		p.reporter.AddOperation(job.Route, "Read", len(signals), start)
		// p.log.Debug(p.reporter.Report())
		p.reporter.Report()

		pack := NewSignalsPack(job.Route, signals)
		p.collectCb(pack) // TODO: Evaluate if it should be done in a goroutine (or based on settings)

	} else if job.Type == Write {
		// TODO: Take into account queue jobs...
		p.writeLock.Lock()
		defer p.writeLock.Unlock()

		err := p.driver.WriteSignals(job.Signals)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *processor[S, M, SigMeta]) RouteState(key string) ProcessorState {
	p.routeLock.RLock()
	defer p.routeLock.RUnlock()
	return p.routeStates[key]
}

func (p *processor[S, M, SigMeta]) AnyFailedState() bool {
	p.routeLock.RLock()
	defer p.routeLock.RUnlock()
	for _, state := range p.routeStates {
		if state == Failed {
			return true
		}
	}
	return false
}

func (p *processor[S, M, SigMeta]) setRouteState(key string, state ProcessorState) {
	p.routeLock.Lock()
	defer p.routeLock.Unlock()
	p.routeStates[key] = state
}

func (p *processor[S, M, SigMeta]) ConnState() ConnectionState {
	p.connLock.RLock()
	defer p.connLock.RUnlock()
	return p.connState
}

func (p *processor[S, M, SigMeta]) setConnState(state ConnectionState) {
	p.connLock.Lock()
	defer p.connLock.Unlock()
	p.connState = state
}
