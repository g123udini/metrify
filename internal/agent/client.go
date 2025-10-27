package agent

import (
	"encoding/json"
	"fmt"
	"gopkg.in/resty.v1"
	models "metrify/internal/model"
)

func UpdateMetric(host string, metric models.Metrics) error {
	path := "/update"
	client := resty.New()
	client.
		SetHeader("Content-Type", "application/json").
		SetHostURL(fmt.Sprintf("http://%s", host))

	body, err := json.Marshal(metric)

	if err != nil {
		return err
	}

	client.R().SetBody(body).Post(path)

	return nil
}
