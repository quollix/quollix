package tools

import (
	"testing"

	"github.com/quollix/common/assert"
)

func TestIsProtectedDockerNetwork(t *testing.T) {
	assert.True(t, isProtectedDockerNetwork(OfficialDatabaseAppNetworkName))
	assert.False(t, isProtectedDockerNetwork(SampleAppDockerNetwork))
}

func TestIsDockerImageUnsupportedPlatformError(t *testing.T) {
	testCases := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "NoMatchingManifestArm64",
			output:   "no matching manifest for linux/arm64/v8 in the manifest list entries",
			expected: true,
		},
		{
			name:     "NoMatchForPlatform",
			output:   "failed to solve: no match for platform in manifest: not found",
			expected: true,
		},
		{
			name:     "ManifestListWithLinuxPlatform",
			output:   "image Error response from daemon: manifest list entries do not contain linux/arm/v8",
			expected: true,
		},
		{
			name:     "DifferentDockerError",
			output:   "pull access denied, repository does not exist or may require authorization",
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, isDockerImageUnsupportedPlatformError(testCase.output))
		})
	}
}

func TestIsDockerHubRateLimitError(t *testing.T) {
	testCases := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "TooManyRequests",
			output:   "toomanyrequests: You have reached your unauthenticated pull rate limit.",
			expected: true,
		},
		{
			name:     "PullRateLimit",
			output:   "ERROR: You have reached your pull rate limit.",
			expected: true,
		},
		{
			name:     "Http429",
			output:   "failed to copy: httpReadSeeker: failed open: 429 Too Many Requests",
			expected: true,
		},
		{
			name:     "DifferentDockerError",
			output:   "pull access denied, repository does not exist or may require authorization",
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, isDockerHubRateLimitError(testCase.output))
		})
	}
}
