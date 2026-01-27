package router

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"metrify/internal/audit"
	"metrify/internal/handler"
	"metrify/internal/service"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMetric(t *testing.T) {
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
			content:  "text/plain",
		},
		{
			name:     "add gauge",
			url:      "/update/gauge/Alloc/22.2",
			method:   http.MethodPost,
			response: http.StatusOK,
			content:  "text/plain",
		},
		{
			name:     "wrong method",
			url:      "/update/counter/PollInterval/22",
			method:   http.MethodPut,
			response: http.StatusMethodNotAllowed,
			content:  "text/plain",
		},
		{
			name:     "not found",
			url:      "/update/counter/22",
			method:   http.MethodPost,
			response: http.StatusNotFound,
			content:  "text/plain",
		},
		{
			name:     "invalid request",
			url:      "/update/counter/PollInterval/22.2",
			method:   http.MethodPost,
			response: http.StatusBadRequest,
			content:  "text/plain",
		},
	}

	filepath := "./testdata/metric/test.json"
	dsn := "postgres://dev:dev@localhost:5432/dev"
	db, _ := sql.Open("pgx", dsn)
	defer os.Remove(filepath)
	ms := service.NewMemStorage(filepath, db)
	logger := service.NewLogger()
	p := audit.NewPublisher()

	h := handler.NewHandler(ms, logger, db, p, true, "")
	ts := httptest.NewServer(Metric(h))
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequest(t, ts, tt.method, tt.url)
			defer resp.Body.Close()

			assert.Equal(t, tt.response, resp.StatusCode)
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	t.Logf("value = %+v", req)

	return resp
}

func ExampleMetric_updateCounter() {
	ms := service.NewMemStorage("", nil)
	logger := service.NewLogger()
	h := handler.NewHandler(ms, logger, nil, nil, false, "")

	ts := httptest.NewServer(Metric(h))
	defer ts.Close()

	req, err := http.NewRequest(
		http.MethodPost,
		ts.URL+"/update/counter/PollInterval/22",
		nil,
	)
	if err != nil {
		fmt.Println("request error")
		return
	}

	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println("request error")
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
}

func ExampleMetric_updateGauge() {
	ms := service.NewMemStorage("", nil)
	logger := service.NewLogger()
	h := handler.NewHandler(ms, logger, nil, nil, false, "")

	ts := httptest.NewServer(Metric(h))
	defer ts.Close()

	req, err := http.NewRequest(
		http.MethodPost,
		ts.URL+"/update/gauge/Alloc/22.5",
		nil,
	)
	if err != nil {
		fmt.Println("request error")
		return
	}

	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println("request error")
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
}

func ExampleMetric_getCounter() {
	ms := service.NewMemStorage("", nil)
	logger := service.NewLogger()
	h := handler.NewHandler(ms, logger, nil, nil, false, "")

	ts := httptest.NewServer(Metric(h))
	defer ts.Close()

	_, err := ts.Client().Post(
		ts.URL+"/update/counter/Test/10",
		"text/plain",
		nil,
	)
	if err != nil {
		fmt.Println("request error")
		return
	}

	resp, err := ts.Client().Get(ts.URL + "/value/counter/Test")
	if err != nil {
		fmt.Println("request error")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read error")
		return
	}

	fmt.Println(string(body))
}

func ExampleMetric_invalidMetric() {
	ms := service.NewMemStorage("", nil)
	logger := service.NewLogger()
	h := handler.NewHandler(ms, logger, nil, nil, false, "")

	ts := httptest.NewServer(Metric(h))
	defer ts.Close()

	resp, err := ts.Client().Post(
		ts.URL+"/update/unknown/Test/10",
		"text/plain",
		nil,
	)
	if err != nil {
		fmt.Println("request error")
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
}
