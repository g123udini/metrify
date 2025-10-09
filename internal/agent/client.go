package agent

import (
	"fmt"
	"net/http"
)

const host = "localhost:8080"
const header = "text/plain"

func SendMetric(metricType string, metricName string, value string) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", host, metricType, metricName, value)

	req, err := http.NewRequest(http.MethodPost, url, nil)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", header)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
