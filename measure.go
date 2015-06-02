package mss

import (
	"errors"
	"time"
)

var (
	ErrMeasurementAlreadyFinished = errors.New("measurement has already finished")
)

type Measurement struct {
	Name       string
	StartedAt  time.Time
	FinishedAt time.Time
	Data       map[string]interface{}
}

func Measure(name string) *Measurement {
	return &Measurement{
		Name:      name,
		StartedAt: time.Now(),
		Data:      make(map[string]interface{}),
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
