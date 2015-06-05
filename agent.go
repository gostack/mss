package mss

import (
	"log"

	"golang.org/x/net/context"

	influx "github.com/influxdb/influxdb/client"
)

// Agent handles the tracking and persistence of measurements, allowing them
// to be batched for example.
type Agent struct {
	// influxClient holds an influxdb client instance to be used when persisting the
	// measurements.
	influxClient *influx.Client

	// influxDBName is the database name to be used
	influxDBName string

	// batchSize specifies how many measurements should be batched together to be persisted
	// in influxdb.
	batchSize uint

	// maxElapsedSecs specifies the max number of seconds to wait before persisting events.
	maxElapsedSecs uint

	// ch holds a channel to pass Measurement instances into the agent goroutine.
	ch chan *Measurement
}

// NewAgent creates a new Agent based on the provided configuration.
func NewAgent(ic *influx.Client, influxDBName string, batchSize uint, maxElapsedSecs uint) *Agent {
	return &Agent{
		influxClient:   ic,
		influxDBName:   influxDBName,
		batchSize:      batchSize,
		maxElapsedSecs: maxElapsedSecs,
		ch:             make(chan *Measurement),
	}
}

// Run loops continuosly processing batch until the context gets canceled.
func (a *Agent) Run(ctx context.Context, done chan<- interface{}) {
	log.Println("mss: agent started")

	b := make([]*Measurement, 0, a.batchSize)

	for {
		select {
		case <-ctx.Done():
			if len(b) > 0 {
				log.Printf("mss: persisting %d measurements", len(b))
				a.persist(b)
			}

			log.Println("mss: stopping agent due", ctx.Err())
			close(done)
			return
		case m := <-a.ch:
			b = append(b, m)

			if len(b) == cap(b) {
				log.Printf("mss: persisting %d measurements", len(b))
				a.persist(b)
				b = b[0:0]
			}
		}
	}
}

// Track takes a measurement and tracks it, and eventually persists it depending
// on batching and timing.
func (a *Agent) Track(m *Measurement) error {
	if err := m.Finish(); err != nil {
		return err
	}

	a.ch <- m
	return nil
}

// Persist writes all the tracked events to InfluxDB
func (a *Agent) persist(batch []*Measurement) error {
	points := make([]influx.Point, len(batch))

	for i, m := range batch {
		p := influx.Point{
			Name:      m.Name,
			Time:      m.FinishedAt,
			Precision: "n",
			Fields: map[string]interface{}{
				"duration": m.Duration().Nanoseconds(),
			},
		}

		for k, v := range m.Data {
			if _, exst := p.Fields[k]; !exst {
				p.Fields[k] = v
			}
		}

		points[i] = p
	}

	b := influx.BatchPoints{
		Database:        a.influxDBName,
		Points:          points,
		Precision:       "n",
		RetentionPolicy: "default",
	}

	resp, err := a.influxClient.Write(b)
	if err != nil {
		return err
	} else if resp != nil && resp.Error() != nil {
		return resp.Error()
	}

	return nil
}
