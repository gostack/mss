package mss_test

import (
	"net/url"
	"testing"

	"github.com/gostack/mss"
	influx "github.com/influxdb/influxdb/client"
)

var cfg = &mss.Config{
	InfluxDB: influx.Config{
		URL:       *mustParseURL("http://localhost:8086"),
		UserAgent: "gostack mss (test)",
	},
}

func mustParseURL(u string) *url.URL {
	pu, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return pu
}

func TestBasicMeasurement(t *testing.T) {
	agent, err := mss.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}

	m := mss.Measure("db")
	m.SetData(map[string]interface{}{
		"query": "SELECT * FROM users WHERE id = 1",
	})

	if err := agent.Track(m); err != nil {
		t.Fatal(err)
	}

	if err := agent.Persist(); err != nil {
		t.Fatal(err)
	}
}
