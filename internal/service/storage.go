package service

import (
	"encoding/json"
	"os"
	"sync"
)

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
	mu       sync.RWMutex
	filepath string
}

func NewMemStorage(filepath string) *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
		filepath: filepath,
	}
}

type Storage interface {
	GetCounter(key string) (int64, bool)
	GetGauge(key string) (float64, bool)
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, delta int64)
	FlushToFile() error
}

func (ms *MemStorage) GetCounter(key string) (int64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, ok := ms.Counters[key]

	return val, ok
}

func (ms *MemStorage) GetGauge(key string) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, ok := ms.Gauges[key]

	return val, ok
}

func (ms *MemStorage) UpdateGauge(name string, value float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.Gauges[name] = value
}

func (ms *MemStorage) UpdateCounter(name string, delta int64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.Counters[name] += delta
}

func (ms *MemStorage) MarshalJSON() ([]byte, error) {
	type dto struct {
		Gauges   map[string]float64 `json:"Gauges"`
		Counters map[string]int64   `json:"Counters"`
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := dto{
		Gauges:   ms.Gauges,
		Counters: ms.Counters,
	}

	return json.Marshal(result)
}

func (ms *MemStorage) ReadFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)

	if err != nil {
		return err
	}

	return json.Unmarshal(data, ms)
}

func (ms *MemStorage) FlushToFile() error {
	data, err := json.MarshalIndent(ms, "", " ")

	if err != nil {
		return err
	}

	return os.WriteFile(ms.filepath, data, 0644)
}
