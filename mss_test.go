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

package mss_test

import (
	"os"
	"testing"
	"time"

	"github.com/gostack/mss"
	influx "github.com/influxdb/influxdb/client"
	"golang.org/x/net/context"
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
	driver.DropInfluxDB()

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
	mss.StartAgent(driver, 2, time.Duration(1*time.Millisecond))

	ctx := context.Background()

	m := mss.NewMeasurement(ctx, "something", mss.Data{"index": 1})
	m.Data["extra"] = "info"
	mss.Record(m)

	mss.Record(mss.NewMeasurement(ctx, "something", mss.Data{"index": 2}))
	mss.Record(mss.NewMeasurement(ctx, "something", mss.Data{"index": 3}))

	time.Sleep(5 * time.Millisecond)

	mss.Record(mss.NewMeasurement(ctx, "something", mss.Data{"index": 4}))
	mss.Record(mss.NewMeasurement(ctx, "something", mss.Data{"index": 5}))

	time.Sleep(1500 * time.Millisecond)

	q := influx.Query{Database: driver.Database, Command: `SELECT * FROM "something"`}
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
