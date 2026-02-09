package service

import (
	"sync"
	"sync/atomic"
	"testing"
)

type testObj struct {
	resetCount int32
	value      int
}

func (o *testObj) Reset() {
	atomic.AddInt32(&o.resetCount, 1)
	o.value = 0
}

func TestNew_NilFactoryPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on nil factory, got none")
		}
	}()

	_ = New[*testObj](nil)
}

func TestPool_Get_ReturnsFromFactory(t *testing.T) {
	var created int32

	p := New(func() *testObj {
		atomic.AddInt32(&created, 1)
		return &testObj{value: 123}
	})

	obj := p.Get()
	if obj == nil {
		t.Fatalf("Get() returned nil")
	}

	if got := atomic.LoadInt32(&created); got != 1 {
		t.Fatalf("factory calls = %d, want %d", got, 1)
	}

	if obj.value != 123 {
		t.Fatalf("obj.value = %d, want %d", obj.value, 123)
	}
	if got := atomic.LoadInt32(&obj.resetCount); got != 0 {
		t.Fatalf("resetCount = %d, want %d", got, 0)
	}
}

func TestPool_Put_CallsReset(t *testing.T) {
	p := New(func() *testObj { return &testObj{} })

	obj := &testObj{value: 777}
	p.Put(obj)

	if got := atomic.LoadInt32(&obj.resetCount); got != 1 {
		t.Fatalf("resetCount = %d, want %d", got, 1)
	}
	if obj.value != 0 {
		t.Fatalf("obj.value = %d, want %d after Reset()", obj.value, 0)
	}
}

func TestPool_Concurrent_GetPut(t *testing.T) {
	var created int32
	p := New(func() *testObj {
		atomic.AddInt32(&created, 1)
		return &testObj{}
	})

	const (
		goroutines = 32
		iterations = 2000
	)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				obj := p.Get()
				if obj == nil {
					t.Errorf("Get() returned nil")
					return
				}
				obj.value = id*1000 + j
				p.Put(obj)
			}
		}(i)
	}

	wg.Wait()

	if atomic.LoadInt32(&created) == 0 {
		t.Fatalf("factory was never called, expected at least 1 call")
	}
}
