package pool

import (
	"errors"
	"sync"
)

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	p sync.Pool
}

func New[T Resettable](factory func() T) (*Pool[T], error) {
	if factory == nil {
		return nil, errors.New("pool.New: factory is nil")
	}

	pp := &Pool[T]{}
	pp.p.New = func() any { return factory() }
	return pp, nil
}

func (pp *Pool[T]) Get() T {
	return pp.p.Get().(T)
}

func (pp *Pool[T]) Put(v T) {
	v.Reset()
	pp.p.Put(v)
}
