package agent

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/resty.v1"
	models "metrify/internal/model"
	"metrify/internal/service"
	"net/http"
	"time"
)

// generate:reset
type HTTPClient struct {
	logger    *zap.SugaredLogger
	resty     *resty.Client
	host      string
	hashKey   string
	maxRetry  int
	publicKey *rsa.PublicKey
}

func NewHTTPClient(host string, logger *zap.SugaredLogger, hashKey string, publicKey *rsa.PublicKey) *HTTPClient {
	return &HTTPClient{
		logger:    logger,
		resty:     resty.New().SetTimeout(8),
		host:      host,
		hashKey:   hashKey,
		maxRetry:  3,
		publicKey: publicKey,
	}
}

func (client *HTTPClient) Close() error {
	return nil
}

func (client *HTTPClient) UpdateMetric(metric models.Metrics) error {
	path := "/update"
	body, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	return client.sendRequest(path, body, client.maxRetry)
}

func (client *HTTPClient) UpdateMetrics(metrics []models.Metrics) error {
	path := "/updates"
	body, err := json.Marshal(metrics)
	if err != nil {
		client.logger.Errorw("failed to marshal metrics", "error", err)
		return err
	}

	return client.sendRequest(path, body, client.maxRetry)
}

func (client *HTTPClient) sendRequest(path string, body []byte, maxRetry int) error {
	client.resty.SetHostURL(fmt.Sprintf("http://%s", client.host))

	req := client.resty.R().
		SetHeader("Content-Type", "application/json")

	if client.hashKey != "" {
		req.SetHeader("HashSHA256", service.SignData(body, client.hashKey))
	}

	if client.publicKey != nil {
		encBody, err := client.encryptBody(body)
		if err != nil {
			return err
		}

		body = encBody
		req.SetHeader("Content-Type", "application/octet-stream")
		req.SetHeader("Content-Encryption", "RSA-PKCS1v15")
	}

	ip, err := getOutboundIP()
	if err == nil {
		req.SetHeader("X-Forwarded-For", ip)
	}

	resp, err := service.Retry(maxRetry, 1*time.Second, 2*time.Second, func() (*resty.Response, error) {
		return req.SetBody(body).Post(path)
	})
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), resp.Body())
	}

	return nil
}

func (client *HTTPClient) encryptBody(plain []byte) ([]byte, error) {
	if client.publicKey == nil {
		return plain, nil
	}

	ciphertext, err := rsa.EncryptPKCS1v15(
		rand.Reader,
		client.publicKey,
		plain,
	)
	if err != nil {
		return nil, err
	}

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(ciphertext)))
	base64.StdEncoding.Encode(encoded, ciphertext)

	return encoded, nil
}
