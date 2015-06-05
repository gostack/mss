package mss

import (
	"fmt"
	"net/url"

	influx "github.com/influxdb/influxdb/client"
)

type Driver interface {
	Persist([]*Measurement) error
}

// NewInfluxDBDriver initializes an InfluxDBDriver instance with the provided configuration.
func NewInfluxDBDriver(cfg *InfluxDBConfig) (*InfluxDBDriver, error) {
	clientCfg := influx.Config{Username: cfg.Username, Password: cfg.Password, UserAgent: "gostack/mss"}

	if u, err := url.Parse(cfg.URL); err != nil {
		return nil, err
	} else {
		clientCfg.URL = *u
	}

	c, err := influx.NewClient(clientCfg)
	if err != nil {
		return nil, err
	}

	return &InfluxDBDriver{Client: c, Database: cfg.Database}, nil
}

// InfluxDBConfig specifies the configuration for connecting to a InfluxDB instance
// that is a little bit more friendly to instantiate from.
type InfluxDBConfig struct {
	URL      string
	Username string
	Password string
	Database string
}

// InfluxDBDriver implements the Driver interface that knows how to persist
// data points into InfluxDB.
type InfluxDBDriver struct {
	Client   *influx.Client
	Database string
}

// Persist implements the Driver interface, persisting the data points to InfluxDB.
func (d *InfluxDBDriver) Persist(batch []*Measurement) error {
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
		Database:        d.Database,
		Points:          points,
		Precision:       "n",
		RetentionPolicy: "default",
	}

	resp, err := d.Client.Write(b)
	if err != nil {
		return err
	} else if resp != nil && resp.Error() != nil {
		return resp.Error()
	}

	return nil
}

// creates the InfluxDB database
func (d *InfluxDBDriver) CreateInfluxDB() error {
	d.DropInfluxDB()

	q := influx.Query{Command: fmt.Sprintf(`CREATE DATABASE "%s"`, d.Database)}
	resp, err := d.Client.Query(q)
	if err != nil {
		return err
	}
	if resp != nil && resp.Error() != nil {
		return resp.Error()
	}

	return nil
}

// drops the influx database
func (d *InfluxDBDriver) DropInfluxDB() error {
	q := influx.Query{Command: fmt.Sprintf(`DROP DATABASE "%s"`, d.Database)}
	if _, err := d.Client.Query(q); err != nil {
		return err
	}

	return nil
}
