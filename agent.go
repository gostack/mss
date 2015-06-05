package mss

import (
	"log"
	"os"
	"os/signal"
	"time"

	"golang.org/x/net/context"
)

var (
	// holds a CancelFunc when it's running
	shutdown context.CancelFunc

	// holds a channel that will be closed once shutdown is complete
	done chan interface{}
)

// Set up signal handling to always properly shutdown
func init() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		<-c
		shutdown()
	}()
}

func StartAgent(d Driver, maxBatchSize uint, maxElapsedTime time.Duration) chan<- *Measurement {
	var ctx context.Context

	ctx, shutdown = context.WithCancel(context.Background())
	done = make(chan interface{})

	agent := newAgent(d, maxBatchSize, maxElapsedTime)
	go func() {
		agent.Run(ctx)
		close(done)
	}()

	return agent.ch
}

func Shutdown() {
	if shutdown != nil {
		shutdown()
		<-done
	}
}

// agent handles the tracking and persistence of measurements, allowing them
// to be batched for example.
type agent struct {
	driver         Driver
	maxBatchSize   uint
	maxElapsedTime time.Duration
	ch             chan *Measurement
}

// NewAgent creates a new Agent based on the provided configuration.
func newAgent(d Driver, maxBatchSize uint, maxElapsedTime time.Duration) *agent {
	return &agent{
		driver:         d,
		maxBatchSize:   maxBatchSize,
		maxElapsedTime: maxElapsedTime,
		ch:             make(chan *Measurement),
	}
}

// Run loops continuosly processing batch until the context gets canceled.
func (a *agent) Run(ctx context.Context) {
	var (
		timeC = time.After(a.maxElapsedTime)
		batch = make([]*Measurement, 0, a.maxBatchSize)
	)

	log.Println("mss: agent started")

loop:
	for {
		select {
		case m := <-a.ch:
			batch = append(batch, m)
			if len(batch) == cap(batch) {
				log.Printf("mss: persisting %d measurements", len(batch))
				if err := a.driver.Persist(batch); err != nil {
					log.Printf("mss: [error] %s", err)
				}
				batch = batch[0:0]
				timeC = time.After(a.maxElapsedTime)
			}

		case <-timeC:
			if len(batch) > 0 {
				log.Printf("mss: %s passed, persisting %d measurements", a.maxElapsedTime, len(batch))
				if err := a.driver.Persist(batch); err != nil {
					log.Printf("mss: [error] %s", err)
				}
				batch = batch[0:0]
			}
			timeC = time.After(a.maxElapsedTime)

		case <-ctx.Done():
			if len(batch) > 0 {
				log.Printf("mss: shuttind down, persisting %d measurements", len(batch))
				if err := a.driver.Persist(batch); err != nil {
					log.Printf("mss: [error] %s", err)
				}
			}

			break loop
		}
	}

	log.Println("mss: agent stopped")
}

// Track takes a measurement and tracks it, and eventually persists it depending
// on batching and timing.
func (a *agent) Track(m *Measurement) error {
	if err := m.Finish(); err != nil {
		return err
	}

	a.ch <- m
	return nil
}
