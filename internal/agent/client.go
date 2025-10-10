package agent

import (
	"fmt"
	"metrify/internal/handler"
	"net/http"
)

const host = "localhost:8080"

func UpdateMetric(metricType string, metricName string, value string) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", host, metricType, metricName, value)

	req, err := http.NewRequest(http.MethodPost, url, nil)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", handler.UpdateContentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
