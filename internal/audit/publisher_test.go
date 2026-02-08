package audit

import (
	"context"
	"errors"
	"testing"
	"time"
)

// generate:reset
type spyReceiver struct {
	calls int
	last  Event
	err   error
}

func (s *spyReceiver) Receive(ctx context.Context, e Event) error {
	s.calls++
	s.last = e
	return s.err
}

func TestNewPublisher_FiltersNilReceivers(t *testing.T) {
	r1 := &spyReceiver{}
	p := NewPublisher(nil, r1, nil)

	if !p.Enabled() {
		t.Fatalf("expected publisher to be enabled when non-nil receiver passed")
	}
}

func TestPublisher_Add_NilDoesNothing(t *testing.T) {
	p := NewPublisher()
	if p.Enabled() {
		t.Fatalf("expected publisher to be disabled initially")
	}

	p.Add(nil)

	if p.Enabled() {
		t.Fatalf("expected publisher to remain disabled after Add(nil)")
	}
}

func TestPublisher_Enabled(t *testing.T) {
	p := NewPublisher()
	if p.Enabled() {
		t.Fatalf("expected Enabled() == false when no receivers")
	}

	p.Add(&spyReceiver{})
	if !p.Enabled() {
		t.Fatalf("expected Enabled() == true when receiver added")
	}
}

func TestPublisher_Publish_NoReceivers_ReturnsNil(t *testing.T) {
	p := NewPublisher()

	if err := p.Publish(context.Background(), Event{TS: 1}); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestPublisher_Publish_CallsAllReceivers(t *testing.T) {
	r1 := &spyReceiver{}
	r2 := &spyReceiver{}
	p := NewPublisher(r1, nil, r2)

	e := Event{TS: 123, Metrics: []string{"Alloc"}, IPAddress: "1.2.3.4"}
	if err := p.Publish(context.Background(), e); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if r1.calls != 1 {
		t.Fatalf("expected r1 called once, got %d", r1.calls)
	}
	if r2.calls != 1 {
		t.Fatalf("expected r2 called once, got %d", r2.calls)
	}

	if r1.last.TS != e.TS || r1.last.IPAddress != e.IPAddress || len(r1.last.Metrics) != 1 || r1.last.Metrics[0] != "Alloc" {
		t.Fatalf("r1 received unexpected event: %+v", r1.last)
	}
	if r2.last.TS != e.TS || r2.last.IPAddress != e.IPAddress || len(r2.last.Metrics) != 1 || r2.last.Metrics[0] != "Alloc" {
		t.Fatalf("r2 received unexpected event: %+v", r2.last)
	}
}

func TestPublisher_Publish_JoinsErrors(t *testing.T) {
	wantErr1 := errors.New("receiver1 failed")
	wantErr2 := errors.New("receiver2 failed")

	r1 := &spyReceiver{err: wantErr1}
	r2 := &spyReceiver{err: wantErr2}
	p := NewPublisher(r1, r2)

	err := p.Publish(context.Background(), Event{TS: 1})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	// errors.Join() возвращает ошибку, которая удовлетворяет errors.Is для каждого исходного err
	if !errors.Is(err, wantErr1) {
		t.Fatalf("expected joined error to contain err1")
	}
	if !errors.Is(err, wantErr2) {
		t.Fatalf("expected joined error to contain err2")
	}
}

func TestNewEvent_SetsFieldsAndTimestamp(t *testing.T) {
	before := time.Now().Unix()
	e := NewEvent([]string{"Alloc", "Frees"}, "192.168.0.42")
	after := time.Now().Unix()

	if e.IPAddress != "192.168.0.42" {
		t.Fatalf("expected ip set, got %q", e.IPAddress)
	}
	if len(e.Metrics) != 2 || e.Metrics[0] != "Alloc" || e.Metrics[1] != "Frees" {
		t.Fatalf("unexpected metrics: %#v", e.Metrics)
	}
	if e.TS < before || e.TS > after {
		t.Fatalf("expected ts between %d and %d, got %d", before, after, e.TS)
	}
}
