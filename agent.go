package mss

import (
	"log"
	"time"
)

// agent handles the tracking and persistence of measurements, allowing them
// to be batched for example.
type agent struct {
	Driver           Driver
	MaxBatchSize     uint
	MaxElapsedTime   time.Duration
	MeasurementsChan <-chan *Measurement
	ShutdownChan     <-chan interface{}
}

// Run loops continuosly processing batch until the context gets canceled.
func (a *agent) Run() {
	var (
		timeC = time.After(a.MaxElapsedTime)
		batch = make([]*Measurement, 0, a.MaxBatchSize)
	)

	log.Println("mss: agent started")

loop:
	for {
		select {
		case m := <-a.MeasurementsChan:
			batch = append(batch, m)
			if len(batch) == cap(batch) {
				log.Printf("mss: persisting %d measurements", len(batch))
				if err := a.Driver.Persist(batch); err != nil {
					log.Printf("mss: [error] %s", err)
				}
				batch = batch[0:0]
				timeC = time.After(a.MaxElapsedTime)
			}

		case <-timeC:
			if len(batch) > 0 {
				log.Printf("mss: %s passed, persisting %d measurements", a.MaxElapsedTime, len(batch))
				if err := a.Driver.Persist(batch); err != nil {
					log.Printf("mss: [error] %s", err)
				}
				batch = batch[0:0]
			}
			timeC = time.After(a.MaxElapsedTime)

		case <-a.ShutdownChan:
			if len(batch) > 0 {
				log.Printf("mss: shuttind down, persisting %d measurements", len(batch))
				if err := a.Driver.Persist(batch); err != nil {
					log.Printf("mss: [error] %s", err)
				}
			}

			break loop
		}
	}

	log.Println("mss: agent stopped")
}
