package tools

import (
	"os"
	"time"

	u "github.com/quollix/common/utils"
)

var (
	DefaultTime                     = time.Unix(0, 0).UTC()
	PrettyFrontendTimeLayout        = "2006-01-02 15:04:05"
	PrettyFrontendTimeLayoutWithDay = PrettyFrontendTimeLayout + ", Mon"
)

const ApplicationVersion = "1.0.3"

const (
	SampleTestRecipientEmail = "recipient@example.invalid"
	SampleTestEmailSubject   = "Test email"
	SampleTestEmailBody      = "This is a test message send by Quollix."
	SampleSMTPHost           = "smtp.example.com"
	SampleSMTPPort           = "587"
	SampleFromEmailAddress   = "test@example.com"
	SampleEmailUsername      = "user"
	SampleEmailPassword      = "password"
)

var FrontendResourceFilesystem = os.DirFS(FrontendResourcesPath)

// this type exists because "SA1029: should not use built-in type string as key for value; define your own type to avoid collisions (staticcheck)"
type AuthKeyType string

const AuthKey = AuthKeyType("auth")

func WaitOneHundredMilliseconds() {
	time.Sleep(100 * time.Millisecond)
}

func WaitOneSecond() {
	time.Sleep(1 * time.Second)
}

func Eventually(check func() error) error {
	return EventuallyWithTimeout(3*time.Second, 50*time.Millisecond, check)
}

func EventuallyWithTimeout(timeout time.Duration, interval time.Duration, check func() error) error {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		err := check()
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(interval)
	}

	if lastErr == nil {
		return nil
	}

	return lastErr
}

type UserAccessLevel int

const (
	AnonymousLevel UserAccessLevel = iota
	UserLevel
	AdminLevel
)

func (a UserAccessLevel) String() string {
	switch a {
	case AnonymousLevel:
		return "anonymous"
	case UserLevel:
		return "user"
	case AdminLevel:
		return "admin"
	default:
		u.Logger.Error("unknown user access level, to be fixed in code", "user_access_level", a)
		return "unknown"
	}
}

const (
	FrontendResourcesPath                 = "frontend/resources"
	FrontendResourcesPathWithLeadingSlash = "/" + FrontendResourcesPath
	FrontendResourcesPathWithSlash        = "/" + FrontendResourcesPath + "/"
	FrontendFramePathInResources          = "global/frame.html"
)

func GetStatusCodeError(expected, actual int) error {
	return u.Logger.NewError("unexpected status code", "expected", expected, "actual", actual)
}
