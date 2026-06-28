package ingress

import (
	"sync"

	u "github.com/quollix/common/utils"
	"golang.org/x/time/rate"
)

var RateLimitExceededError = "rate limit exceeded"

type RateLimitChecker interface {
	CheckRequestAllowed(clientIp string) error
}

type RateLimitCheckerImpl struct {
	RequestsPerSec float64
	BurstPerIp     int

	mutex    sync.Mutex
	limiters map[string]*rate.Limiter
}

func NewRateLimitChecker(requestsPerSec float64, burstPerIp int) *RateLimitCheckerImpl {
	return &RateLimitCheckerImpl{
		RequestsPerSec: requestsPerSec,
		BurstPerIp:     burstPerIp,
		limiters:       map[string]*rate.Limiter{},
	}
}

func (r *RateLimitCheckerImpl) CheckRequestAllowed(clientIp string) error {
	if r.RequestsPerSec <= 0 || r.BurstPerIp <= 0 {
		return nil
	}

	limiter := r.getLimiter(clientIp)
	if !limiter.Allow() {
		return u.Logger.NewError(RateLimitExceededError)
	}
	return nil
}

func (r *RateLimitCheckerImpl) getLimiter(clientIp string) *rate.Limiter {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	limiter, ok := r.limiters[clientIp]
	if !ok {
		limiter = rate.NewLimiter(rate.Limit(r.RequestsPerSec), r.BurstPerIp)
		r.limiters[clientIp] = limiter
	}
	return limiter
}
