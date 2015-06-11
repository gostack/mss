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
	"errors"
	"time"

	"github.com/gostack/ctxinfo"
	"golang.org/x/net/context"
)

var (
	ErrMeasurementAlreadyFinished = errors.New("measurement has already finished")
)

type Data map[string]interface{}

type Measurement struct {
	EnvInfo ctxinfo.EnvInfo
	TxInfo  ctxinfo.TxInfo

	Name       string
	StartedAt  time.Time
	FinishedAt time.Time
	Data       Data
}

func NewMeasurement(ctx context.Context, name string, data Data) *Measurement {
	return &Measurement{
		EnvInfo:   ctxinfo.EnvFromContext(ctx),
		TxInfo:    ctxinfo.TxFromContext(ctx),
		Name:      name,
		StartedAt: time.Now(),
		Data:      data,
	}
}

func (m *Measurement) Finish() error {
	if !m.FinishedAt.IsZero() {
		return ErrMeasurementAlreadyFinished
	}

	m.FinishedAt = time.Now()
	return nil
}

func (m *Measurement) SetData(d map[string]interface{}) {
	for k, v := range d {
		m.Data[k] = v
	}
}

func (m Measurement) Duration() time.Duration {
	if m.FinishedAt.IsZero() {
		return time.Duration(0)
	}

	return m.FinishedAt.Sub(m.StartedAt)
}
