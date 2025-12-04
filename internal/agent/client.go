package agent

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/resty.v1"
	models "metrify/internal/model"
	"metrify/internal/service"
	"net/http"
	"time"
)

type Client struct {
	logger   *zap.SugaredLogger
	resty    *resty.Client
	host     string
	hashKey  string
	maxRetry int
}

func NewClient(host string, logger *zap.SugaredLogger, hashKey string) *Client {
	return &Client{
		logger:   logger,
		resty:    resty.New().SetTimeout(15 * time.Second),
		host:     host,
		hashKey:  hashKey,
		maxRetry: 3,
	}
}

func (client *Client) UpdateMetric(metric models.Metrics) error {
	path := "/update"
	body, err := json.Marshal(metric)

	if err != nil {
		return err
	}

	return client.sendRequest(path, body, client.maxRetry)
}

func (client *Client) UpdateMetrics(metrics []models.Metrics) error {
	path := "/updates"
	body, err := json.Marshal(metrics)

	if err != nil {
		client.logger.Errorw("failed to marshal metric", "error", err)
	}

	return client.sendRequest(path, body, client.maxRetry)
}

func (client *Client) sendRequest(path string, body []byte, maxRetry int) error {
	client.
		resty.
		SetHeader("Content-Type", "application/json").
		SetHostURL(fmt.Sprintf("http://%s", client.host))

	if client.hashKey != "" {
		client.resty.SetHeader("HashSHA256", service.SignData(body, client.hashKey))
	}

	resp, err := service.Retry(maxRetry, 1*time.Second, 2*time.Second, func() (*resty.Response, error) {
		return client.resty.R().SetBody(body).Post(path)
	})

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), resp.Body())
	}

	return nil
}
