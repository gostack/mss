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
	"os"
	"os/signal"
	"time"
)

var (
	running          bool
	measurementsChan = make(chan *Measurement)
	shutdownChan     = make(chan interface{}, 1)
)

// Set up signal handling to always properly shutdown
func init() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		<-c
		Shutdown()
	}()
}

// StartAgent starts running agent with the given configuration
func StartAgent(d Driver, maxBatchSize uint, maxElapsedTime time.Duration) {
	agent := agent{
		Driver:           d,
		MaxBatchSize:     maxBatchSize,
		MaxElapsedTime:   maxElapsedTime,
		MeasurementsChan: measurementsChan,
		ShutdownChan:     shutdownChan,
	}

	go func() {
		agent.Run()
		close(shutdownChan)
	}()

	running = true
}

func Shutdown() {
	shutdownChan <- nil
	<-shutdownChan
	running = false
}

func Record(m *Measurement) {
	m.Finish()

	if running {
		measurementsChan <- m
	}
}
