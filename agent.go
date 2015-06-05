package mss

import (
	"log"
	"time"

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

	// maxElapsedTime specifies the max number of seconds to wait before persisting events.
	maxElapsedTime time.Duration

	// ch holds a channel to pass Measurement instances into the agent goroutine.
	ch chan *Measurement
}

// NewAgent creates a new Agent based on the provided configuration.
func NewAgent(ic *influx.Client, influxDBName string, batchSize uint, maxElapsedTime time.Duration) *Agent {
	return &Agent{
		influxClient:   ic,
		influxDBName:   influxDBName,
		batchSize:      batchSize,
		maxElapsedTime: maxElapsedTime,
		ch:             make(chan *Measurement),
	}
}

// Run loops continuosly processing batch until the context gets canceled.
func (a *Agent) Run(ctx context.Context, done chan<- interface{}) {
	var (
		timeC = time.After(a.maxElapsedTime)
		batch = make([]*Measurement, 0, a.batchSize)
	)

	log.Println("mss: agent started")

	for {
		select {
		case m := <-a.ch:
			batch = append(batch, m)
			if len(batch) == cap(batch) {
				log.Printf("mss: persisting %d measurements", len(batch))
				a.persist(batch)
				timeC = time.After(a.maxElapsedTime)
				batch = batch[0:0]
			}

		case <-timeC:
			if len(batch) > 0 {
				log.Printf("mss: %s passed, persisting %d measurements", a.maxElapsedTime, len(batch))
				a.persist(batch)
				timeC = time.After(a.maxElapsedTime)
				batch = batch[0:0]
			}

		case <-ctx.Done():
			if len(batch) > 0 {
				log.Printf("mss: shuttind down, persisting %d measurements", len(batch))
				a.persist(batch)
			}

			log.Println("mss: agent stopped")
			if done != nil {
				close(done)
			}

			return
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
