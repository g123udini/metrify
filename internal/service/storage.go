package service

import (
	"database/sql"
	"encoding/json"
	"os"
	"sync"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
	filepath string
	db       *sql.DB
}

func NewMemStorage(filepath string, db *sql.DB) *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
		filepath: filepath,
		db:       db,
	}
}

type Storage interface {
	GetCounter(key string) (int64, bool)
	GetGauge(key string) (float64, bool)
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, delta int64) error
	FlushToFile() error
	FlushToDB() error
}

func (ms *MemStorage) GetCounter(key string) (int64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, ok := ms.counters[key]

	return val, ok
}

func (ms *MemStorage) GetGauge(key string) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, ok := ms.gauges[key]

	return val, ok
}

func (ms *MemStorage) UpdateGauge(name string, value float64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.gauges[name] = value

	if ms.db != nil {
		_, err := ms.db.Exec("INSERT INTO metrics (name, value) VALUES ($1, $2)", name, value)

		return err
	}

	return nil
}

func (ms *MemStorage) UpdateCounter(name string, delta int64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.counters[name] += delta

	if ms.db != nil {
		_, err := ms.db.Exec("INSERT INTO metrics (name, value) VALUES ($1, $2)", name, delta)

		return err
	}

	return nil
}

func (ms *MemStorage) UnmarshalJSON(data []byte) error {
	type dto struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := dto{}

	err := json.Unmarshal(data, &result)

	ms.gauges = result.Gauges
	ms.counters = result.Counters

	return err
}

func (ms *MemStorage) MarshalJSON() ([]byte, error) {
	type dto struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := dto{
		Gauges:   ms.gauges,
		Counters: ms.counters,
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

func (ms *MemStorage) FlushToDB() error {
	if ms.db == nil {
		return nil
	}

	for name, value := range ms.gauges {
		_, err := ms.db.Exec("INSERT INTO metrics (name, value) VALUES ($1, $2)", name, value)

		return err
	}

	for name, value := range ms.counters {
		_, err := ms.db.Exec("INSERT INTO metrics (name, value) VALUES ($1, $2)", name, value)

		return err
	}

	return nil
}
