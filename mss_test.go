package mss_test

import (
	"os"
	"testing"
	"time"

	"github.com/gostack/mss"
	influx "github.com/influxdb/influxdb/client"
)

var driver *mss.InfluxDBDriver

func init() {
	var err error
	driver, err = mss.NewInfluxDBDriver(&mss.InfluxDBConfig{
		URL:      "http://localhost:8086",
		Database: "mss_test",
	})
	if err != nil {
		panic(err)
	}
}

// Define a custom global setup and teardown for the test suite
func TestMain(m *testing.M) {
	if err := driver.CreateInfluxDB(); err != nil {
		panic(err)
	}

	retCode := m.Run()

	if err := driver.DropInfluxDB(); err != nil {
		panic(err)
	}

	mss.Shutdown()
	os.Exit(retCode)
}

func TestBasicMeasurement(t *testing.T) {
	ch := mss.StartAgent(driver, 2, time.Duration(1*time.Millisecond))

	ch <- mss.Measure("something", mss.Data{"index": 1})
	ch <- mss.Measure("something", mss.Data{"index": 2})
	ch <- mss.Measure("something", mss.Data{"index": 3})

	time.Sleep(5 * time.Millisecond)

	ch <- mss.Measure("something", mss.Data{"index": 4})
	ch <- mss.Measure("something", mss.Data{"index": 5})

	time.Sleep(100 * time.Millisecond)

	q := influx.Query{Database: driver.Database, Command: "SELECT * FROM something"}
	resp, err := driver.Client.Query(q)
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
