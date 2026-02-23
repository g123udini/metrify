package agent

import (
	"encoding/json"
	"os"
)

type Config struct {
	Address        string `json:"address"`
	ReportInterval int    `json:"report_interval"`
	PollInterval   int    `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

func ConfigFromFile(filePath string) (*Config, error) {
	var config Config
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(jsonData, &config)

	return &config, nil
}
