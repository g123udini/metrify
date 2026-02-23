package service

import (
	"encoding/json"
	"os"
)

func FromFile[T any](filePath string) (*T, error) {
	var cfg T

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
