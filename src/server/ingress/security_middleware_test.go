package ingress

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

func newSecurityTestServer(maxBodySize int64, requestsPerSecond float64, burstLimit int) *httptest.Server {
	sec := NewMiddleware(Config{
		MaxBodyBytes:   maxBodySize,
		RequestsPerSec: requestsPerSecond,
		BurstPerIP:     burstLimit,
	})

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, RequestBodyTooLargeError, http.StatusRequestEntityTooLarge)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	wrapped := sec.Handler(baseHandler)
	return httptest.NewServer(wrapped)
}

func TestBodyLimit(t *testing.T) {
	ts := newSecurityTestServer(10, 0, 0)
	defer ts.Close()

	small := bytes.Repeat([]byte("x"), 5)
	large := bytes.Repeat([]byte("x"), 20)

	resp, err := http.Post(ts.URL, "text/plain", bytes.NewReader(small))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Post(ts.URL, "text/plain", bytes.NewReader(large))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestRateLimit(t *testing.T) {
	ts := newSecurityTestServer(0, 1, 1)
	defer ts.Close()

	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = client.Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRateLimitDisabled(t *testing.T) {
	ts := newSecurityTestServer(0, 0, 10)
	defer ts.Close()

	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 20; i++ {
		resp, err := client.Get(ts.URL)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}
