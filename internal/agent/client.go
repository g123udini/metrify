package agent

import (
	"fmt"
	"gopkg.in/resty.v1"
	"metrify/internal/handler"
)

var host = "http://localhost"

func UpdateMetric(port, metricType string, metricName string, value string) error {
	path := "/update/{metricType}/{metricName}/{value}"
	client := resty.New()
	client.
		SetHeader("Content-Type", handler.TextUpdateContentType).
		SetHostURL(fmt.Sprintf(host + port)).
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
