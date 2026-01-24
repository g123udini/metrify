package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sync"
)

type FileReceiver struct {
	path string
	mu   sync.Mutex
}

func NewFileReceiver(path string) *FileReceiver {
	if path == "" {
		return nil
	}
	return &FileReceiver{path: path}
}

func (r *FileReceiver) Receive(ctx context.Context, e Event) error {
	_ = ctx

	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	f, err := os.OpenFile(r.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return nil
}

type HTTPReceiver struct {
	url    string
	client *http.Client
}

func NewHTTPReceiver(url string, client *http.Client) *HTTPReceiver {
	if url == "" {
		return nil
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPReceiver{url: url, client: client}
}

func (r *HTTPReceiver) Receive(ctx context.Context, e Event) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("audit http receiver: non-2xx status: " + resp.Status)
	}

	return nil
}
