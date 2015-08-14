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
	"fmt"
	"net/url"
	"time"

	influx "github.com/influxdb/influxdb/client"
	"github.com/rubyist/circuitbreaker"
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

	return &InfluxDBDriver{Client: c, Database: cfg.Database, cb: circuit.NewThresholdBreaker(10)}, nil
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
	cb       *circuit.Breaker
}

// Persist implements the Driver interface, persisting the data points to InfluxDB.
func (d *InfluxDBDriver) Persist(batch []*Measurement) error {
	points := make([]influx.Point, len(batch))

	for i, m := range batch {
		p := influx.Point{
			Measurement: m.Name,
			Time:        m.FinishedAt,
			Precision:   "n",
			Tags: map[string]string{
				"hostname":    m.Info.Hostname,
				"application": m.Info.Application,
			},
			Fields: map[string]interface{}{
				"transactionID": m.Info.TransactionID.String(),
				"value":         m.Duration().Nanoseconds(),
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
		Database:         d.Database,
		Points:           points,
		Precision:        "n",
		RetentionPolicy:  "default",
		WriteConsistency: influx.ConsistencyAny,
	}

	return d.cb.Call(func() error {
		resp, err := d.Client.Write(b)
		if err != nil {
			return err
		}
		if resp != nil && resp.Error() != nil {
			return resp.Error()
		}
		return nil
	}, 2*time.Second)
}

// creates the InfluxDB database
func (d *InfluxDBDriver) CreateInfluxDB() error {
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
