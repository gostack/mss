package mss_test

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gostack/mss"
	influx "github.com/influxdb/influxdb/client"
	"golang.org/x/net/context"
)

var (
	// Configuration to be used for testing purpose of this test
	influxConfig = influx.Config{
		URL:       *mustParseURL("http://localhost:8086"),
		UserAgent: "gostack mss (test)",
	}

	// influx client instance to be used for this test, see init()
	influxClient *influx.Client
)

// test package initialization
func init() {
	ic, err := influx.NewClient(influxConfig)
	if err != nil {
		panic(err)
	}

	if _, _, err := ic.Ping(); err != nil {
		panic(err)
	}

	influxClient = ic
}

// Define a custom global setup and teardown for the test suite
func TestMain(m *testing.M) {
	if err := createInfluxDB(); err != nil {
		panic(err)
	}

	retCode := m.Run()

	if err := dropInfluxDB(); err != nil {
		panic(err)
	}

	os.Exit(retCode)
}

func TestBasicMeasurement(t *testing.T) {
	ctx, stop := context.WithCancel(context.Background())
	done := make(chan interface{})

	a := mss.NewAgent(influxClient, "mss_test", 1, 5)
	go a.Run(ctx, done)

	m := mss.Measure("something", mss.Data{"extra": "information"})
	a.Track(m)

	stop()
	<-done

	time.Sleep(100 * time.Millisecond)

	q := influx.Query{Database: "mss_test", Command: "SELECT * FROM something"}
	resp, err := influxClient.Query(q)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Results) == 0 {
		t.Error("no results")
	}

	r := resp.Results[0]
	if len(r.Series) == 0 {
		t.Errorf("no series on %#v", r)
	}

	if len(r.Series[0].Values) == 0 {
		t.Errorf("no values on %#v", r.Series[0])
	}
}

// creates the InfluxDB database
func createInfluxDB() error {
	dropInfluxDB()

	q := influx.Query{Command: "CREATE DATABASE mss_test"}

	resp, err := influxClient.Query(q)
	if err != nil {
		return err
	}
	if resp != nil && resp.Error() != nil {
		return resp.Error()
	}

	return nil
}

// drops the influx database
func dropInfluxDB() error {
	q := influx.Query{Command: "DROP DATABASE mss_test"}
	if _, err := influxClient.Query(q); err != nil {
		return err
	}

	return nil
}

// mustParseURL simply parses url using url.Parse but panic if it fails
func mustParseURL(u string) *url.URL {
	pu, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return pu
}
