package service

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	p sync.Pool
}

func New[T Resettable](factory func() T) *Pool[T] {
	if factory == nil {
		panic("pool.New: factory is nil")
	}

	pp := &Pool[T]{}
	pp.p.New = func() any { return factory() }
	return pp
}

func (pp *Pool[T]) Get() T {
	return pp.p.Get().(T)
}

func (pp *Pool[T]) Put(v T) {
	v.Reset()
	pp.p.Put(v)
}
