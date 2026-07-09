package ingress

import (
	"net"
	"net/http"
)

var RequestBodyTooLargeError = "request body too large"

type Config struct {
	MaxBodyBytes   int64
	RequestsPerSec float64
	BurstPerIP     int
}

type Middleware struct {
	cfg              Config
	rateLimitChecker RateLimitChecker
}

func NewMiddleware(cfg Config) *Middleware {
	return &Middleware{
		cfg:              cfg,
		rateLimitChecker: NewRateLimitChecker(cfg.RequestsPerSec, cfg.BurstPerIP),
	}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isContentLengthWithinLimit(r.ContentLength, m.cfg.MaxBodyBytes) {
			http.Error(w, RequestBodyTooLargeError, http.StatusRequestEntityTooLarge)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, m.cfg.MaxBodyBytes)

		clientIp := clientIP(r)
		if err := m.rateLimitChecker.CheckRequestAllowed(clientIp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isContentLengthWithinLimit(contentLength int64, maxBodyBytes int64) bool {
	if maxBodyBytes <= 0 {
		return true
	}
	if contentLength < 0 {
		return true
	}
	return contentLength <= maxBodyBytes
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
