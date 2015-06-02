package mss

import (
	influx "github.com/influxdb/influxdb/client"
)

// Config defines a struct containing all the available configuration for
// this package.
type Config struct {
	// Embeds the (github.com/influxdb/influxdb/client).Config to be used
	// when connecting to the InfluxDB instance.
	InfluxDB influx.Config
}
