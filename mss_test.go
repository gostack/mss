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

	a := mss.NewAgent(influxClient, "mss_test", 2, time.Duration(1*time.Millisecond))
	go a.Run(ctx, done)

	a.Track(mss.Measure("something", mss.Data{"index": 1}))

	a.Track(mss.Measure("something", mss.Data{"index": 2}))

	a.Track(mss.Measure("something", mss.Data{"index": 3}))

	time.Sleep(5 * time.Millisecond)

	a.Track(mss.Measure("something", mss.Data{"index": 4}))

	a.Track(mss.Measure("something", mss.Data{"index": 5}))

	stop()
	<-done

	time.Sleep(100 * time.Millisecond)

	q := influx.Query{Database: "mss_test", Command: "SELECT * FROM something"}
	resp, err := influxClient.Query(q)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected %d results but got %d", 1, len(resp.Results))
	}

	r := resp.Results[0]
	if len(r.Series) != 1 {
		t.Fatalf("expected %d series but got %d", 1, len(r.Series))
	}

	s := r.Series[0]
	if len(s.Values) != 5 {
		t.Fatalf("expected %d series but got %d", 5, len(s.Values))
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
