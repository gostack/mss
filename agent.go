package mss

import (
	"log"

	influx "github.com/influxdb/influxdb/client"
)

// Agent handles the tracking and persistence of measurements, allowing them
// to be batched for example.
type Agent struct {
	influxClient *influx.Client
	batch        []*Measurement
}

// NewAgent creates a new Agent based on the provided configuration.
func NewAgent(cfg *Config) (*Agent, error) {
	c, err := influx.NewClient(cfg.InfluxDB)
	if err != nil {
		return nil, err
	}

	if _, _, err := c.Ping(); err != nil {
		return nil, err
	}

	return &Agent{influxClient: c, batch: make([]*Measurement, 0)}, nil
}

// Track takes a measurement and tracks it, and eventually persists it depending
// on batching and timing.
func (a *Agent) Track(m *Measurement) error {
	if err := m.Finish(); err != nil {
		return err
	}

	a.batch = append(a.batch, m)
	return nil
}

// Persist writes all the tracked events to InfluxDB
func (a *Agent) Persist() error {
	points := make([]influx.Point, len(a.batch))

	for i, m := range a.batch {
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
		Database:        "mss",
		Points:          points,
		Precision:       "n",
		RetentionPolicy: "default",
	}

	log.Printf("%#v", b)

	resp, err := a.influxClient.Write(b)
	if err != nil {
		return err
	} else if resp != nil && resp.Error() != nil {
		return resp.Error()
	}

	return nil
}
