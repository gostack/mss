package mss

import (
	"os"
	"os/signal"
	"time"
)

var (
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
}

func Shutdown() {
	shutdownChan <- nil
	<-shutdownChan
}

func Measure(name string, data Data) *Measurement {
	return &Measurement{
		Name:      name,
		StartedAt: time.Now(),
		Data:      data,
	}
}

func Done(m *Measurement) {
	m.Finish()
	measurementsChan <- m
}
