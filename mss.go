package mss

import (
	"net/http"
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

func MeasureHTTP(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		m := Measure("http", Data{"url": req.RequestURI, "content-type": req.Header.Get("Content-Type")})
		defer Done(m)

		sw := statusResponseWriter{w, 0}
		h.ServeHTTP(&sw, req)

		m.Data["status"] = sw.Status
	}

	return http.HandlerFunc(fn)
}

type statusResponseWriter struct {
	http.ResponseWriter
	Status int
}

func (w *statusResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.Status = status
}
