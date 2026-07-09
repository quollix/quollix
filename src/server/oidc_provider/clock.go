package oidc_provider

import "time"

type Clock interface {
	Now() time.Time
}

type ClockImpl struct{}

func (c ClockImpl) Now() time.Time {
	return time.Now()
}
