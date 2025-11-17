package agent

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/resty.v1"
	models "metrify/internal/model"
	"net/http"
)

type Client struct {
	logger *zap.SugaredLogger
	resty  *resty.Client
	host   string
}

func NewClient(host string, logger *zap.SugaredLogger) *Client {
	return &Client{
		logger: logger,
		resty:  resty.New(),
		host:   host,
	}
}

func (client *Client) UpdateMetric(metric models.Metrics) error {
	path := "/update"
	body, err := json.Marshal(metric)

	if err != nil {
		return err
	}

	return client.sendRequest(path, body)
}

func (client *Client) UpdateMetrics(metrics []models.Metrics) error {
	path := "/updates"
	body, err := json.Marshal(metrics)

	if err != nil {
		client.logger.Errorw("failed to marshal metric", "error", err)
	}

	return client.sendRequest(path, body)
}

func (client *Client) sendRequest(path string, body []byte) error {
	client.
		resty.
		SetHeader("Content-Type", "application/json").
		SetHostURL(fmt.Sprintf("http://%s", client.host))

	resp, err := client.resty.R().SetBody(body).Post(path)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), resp.Body())
	}

	return nil
}
