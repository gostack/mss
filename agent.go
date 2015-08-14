/*
Copyright 2015 Rodrigo Rafael Monti Kochenburger

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
			if len(batch) >= cap(batch)/2 {
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
