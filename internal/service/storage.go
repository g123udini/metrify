package service

import "encoding/json"

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

type Storage interface {
	GetCounter(key string) int64
	GetGauge(key string) float64
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, delta int64)
}

func (ms *MemStorage) GetCounter(key string) int64 {
	return ms.counters[key]
}

func (ms *MemStorage) GetGauge(key string) float64 {
	return ms.gauges[key]
}

func (ms *MemStorage) UpdateGauge(name string, value float64) {
	ms.gauges[name] = value
}

func (ms *MemStorage) UpdateCounter(name string, delta int64) {
	ms.counters[name] += delta
}

func (ms *MemStorage) MarshalJSON() ([]byte, error) {
	type dto struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}

	result := dto{
		Gauges:   ms.gauges,
		Counters: ms.counters,
	}

	return json.Marshal(result)
}
