package audit

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Event struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type Receiver interface {
	Receive(ctx context.Context, e Event) error
}

type Publisher struct {
	mu        sync.RWMutex
	receivers []Receiver
}

func NewPublisher(receivers ...Receiver) *Publisher {
	p := &Publisher{}
	for _, r := range receivers {
		if r != nil {
			p.receivers = append(p.receivers, r)
		}
	}
	return p
}

func (p *Publisher) Add(r Receiver) {
	if r == nil {
		return
	}
	p.mu.Lock()
	p.receivers = append(p.receivers, r)
	p.mu.Unlock()
}

func (p *Publisher) Enabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.receivers) > 0
}

func (p *Publisher) Publish(ctx context.Context, e Event) error {
	p.mu.RLock()
	rs := append([]Receiver(nil), p.receivers...) // копия
	p.mu.RUnlock()

	if len(rs) == 0 {
		return nil
	}

	var errs []error
	for _, r := range rs {
		if r == nil {
			continue
		}
		if err := r.Receive(ctx, e); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func NewEvent(metrics []string, ip string) Event {
	return Event{
		TS:        time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: ip,
	}
}
