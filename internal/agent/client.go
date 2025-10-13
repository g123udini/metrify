package agent

import (
	"fmt"
	"gopkg.in/resty.v1"
	"metrify/internal/handler"
)

func UpdateMetric(host, metricType string, metricName string, value string) error {
	path := "/update/{metricType}/{metricName}/{value}"
	client := resty.New()
	client.
		SetHeader("Content-Type", handler.TextUpdateContentType).
		SetHostURL(fmt.Sprintf("http://%s", host)).
		SetPathParams(map[string]string{
			"metricName": metricName,
			"metricType": metricType,
			"value":      value,
		})

	_, err := client.R().Post(path)

	if err != nil {
		return err
	}

	return nil
}
