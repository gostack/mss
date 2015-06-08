package mss

import (
	"errors"
	"time"

	"github.com/gostack/web"
	"golang.org/x/net/context"
)

var (
	ErrMeasurementAlreadyFinished = errors.New("measurement has already finished")
)

type Data map[string]interface{}

type Measurement struct {
	Info web.Info

	Name       string
	StartedAt  time.Time
	FinishedAt time.Time
	Data       Data
}

func NewMeasurement(ctx context.Context, name string, data Data) *Measurement {
	return &Measurement{
		Info:      web.InfoFromContext(ctx),
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
