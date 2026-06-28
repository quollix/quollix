package ingress

import (
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestRateLimitCheckerImpl_AllowsFirstRequest_RejectsSecondRequest_SameIp(t *testing.T) {
	checker := NewRateLimitChecker(0.000001, 1)
	assert.Nil(t, checker.CheckRequestAllowed("1.2.3.4"))
	err := checker.CheckRequestAllowed("1.2.3.4")
	assert.Equal(t, RateLimitExceededError, u.ExtractError(err))
}

func TestRateLimitCheckerImpl_SeparateLimitersPerIp(t *testing.T) {
	checker := NewRateLimitChecker(0.000001, 1)

	assert.Nil(t, checker.CheckRequestAllowed("1.2.3.4"))
	assert.Nil(t, checker.CheckRequestAllowed("5.6.7.8"))

	err := checker.CheckRequestAllowed("1.2.3.4")
	assert.Equal(t, RateLimitExceededError, u.ExtractError(err))
	err = checker.CheckRequestAllowed("5.6.7.8")
	assert.Equal(t, RateLimitExceededError, u.ExtractError(err))
}

func TestRateLimitCheckerImpl_RefillsOverTime(t *testing.T) {
	checker := NewRateLimitChecker(5, 1) // 5 req/sec → refill every 200ms
	ip := "1.2.3.4"
	assert.Nil(t, checker.CheckRequestAllowed(ip)) // consume burst token
	err := checker.CheckRequestAllowed(ip)
	assert.Equal(t, RateLimitExceededError, u.ExtractError(err))

	time.Sleep(250 * time.Millisecond)             // wait for refill
	assert.Nil(t, checker.CheckRequestAllowed(ip)) // allowed again
}
