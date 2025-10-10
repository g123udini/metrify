package handler

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdate(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		method   string
		response int
		content  string
	}{
		{
			name:     "add counter",
			url:      "/update/counter/PollInterval/22",
			method:   http.MethodPost,
			response: http.StatusOK,
			content:  UpdateContentType,
		},
		{
			name:     "add gauge",
			url:      "/update/gauge/Alloc/22.2",
			method:   http.MethodPost,
			response: http.StatusOK,
			content:  UpdateContentType,
		},
		{
			name:     "wrong method",
			url:      "/update/counter/PollInterval/22",
			method:   http.MethodPut,
			response: http.StatusMethodNotAllowed,
			content:  UpdateContentType,
		},
		{
			name:     "not found",
			url:      "/update/counter/22",
			method:   http.MethodPost,
			response: http.StatusNotFound,
			content:  UpdateContentType,
		},
		{
			name:     "invalid request",
			url:      "/update/counter/PollInterval/22.2",
			method:   http.MethodPost,
			response: http.StatusBadRequest,
			content:  UpdateContentType,
		},
		{
			name:     "wrong content",
			url:      "/update/counter/PollInterval/22",
			method:   http.MethodPost,
			response: http.StatusUnsupportedMediaType,
			content:  "application/json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			req.Header.Set("Content-Type", tt.content)
			rec := httptest.NewRecorder()

			Update(rec, req)

			assert.Equal(t, tt.response, rec.Code)
		})
	}
}
