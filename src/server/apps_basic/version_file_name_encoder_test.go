package apps_basic

import (
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

var (
	encoder   = &VersionFileNameEncoderImpl{}
	createdAt = time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
)

func TestEncodeComposeArchiveName(t *testing.T) {
	testCases := []struct {
		name          string
		input         ComposeArchiveName
		expectedFile  string
		expectedError string
	}{
		{
			name: "happy path",
			input: ComposeArchiveName{
				Maintainer:               "samplemaintainer",
				AppName:                  "sampleapp",
				Version:                  "v1.2.3",
				VersionCreationTimestamp: createdAt,
			},
			expectedFile: "samplemaintainer_sampleapp_v1.2.3_2025-01-02-03-04-05.yml",
		},
		{
			name: "empty maintainer",
			input: ComposeArchiveName{
				Maintainer:               "",
				AppName:                  "sampleapp",
				Version:                  "v1.2.3",
				VersionCreationTimestamp: createdAt,
			},
			expectedError: "maintainer, appName, version must be non-empty",
		},
		{
			name: "empty app name",
			input: ComposeArchiveName{
				Maintainer:               "samplemaintainer",
				AppName:                  "",
				Version:                  "v1.2.3",
				VersionCreationTimestamp: createdAt,
			},
			expectedError: "maintainer, appName, version must be non-empty",
		},
		{
			name: "empty version",
			input: ComposeArchiveName{
				Maintainer:               "samplemaintainer",
				AppName:                  "sampleapp",
				Version:                  "",
				VersionCreationTimestamp: createdAt,
			},
			expectedError: "maintainer, appName, version must be non-empty",
		},
		{
			name: "zero createdAt",
			input: ComposeArchiveName{
				Maintainer:               "samplemaintainer",
				AppName:                  "sampleapp",
				Version:                  "v1.2.3",
				VersionCreationTimestamp: time.Time{},
			},
			expectedError: "createdAt must be set",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualFile, err := encoder.EncodeComposeArchiveName(&testCase.input)

			if testCase.expectedError != "" {
				assert.NotNil(t, err)
				assert.Equal(t, testCase.expectedError, u.ExtractError(err))
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedFile, actualFile)
		})
	}
}

func TestDecodeComposeArchiveName(t *testing.T) {
	testCases := []struct {
		name          string
		inputFile     string
		expectedDto   *ComposeArchiveName
		expectedError string
	}{
		{
			name:      "happy path",
			inputFile: "samplemaintainer_sampleapp_v1.2.3_2025-01-02-03-04-05.yml",
			expectedDto: &ComposeArchiveName{
				Maintainer:               "samplemaintainer",
				AppName:                  "sampleapp",
				Version:                  "v1.2.3",
				VersionCreationTimestamp: createdAt,
			},
		},
		{
			name:      "happy path with directory prefix",
			inputFile: "/tmp/samplemaintainer_sampleapp_v1.2.3_2025-01-02-03-04-05.yml",
			expectedDto: &ComposeArchiveName{
				Maintainer:               "samplemaintainer",
				AppName:                  "sampleapp",
				Version:                  "v1.2.3",
				VersionCreationTimestamp: createdAt,
			},
		},
		{
			name:          "missing .yml suffix",
			inputFile:     "samplemaintainer_sampleapp_v1.2.3_2025-01-02-03-04-05.tar",
			expectedError: "file must end with .yml",
		},
		{
			name:          "wrong number of underscore parts",
			inputFile:     "samplemaintainer_sampleapp_2025-01-02-03-04-05.yml",
			expectedError: "expected 4 underscore-separated parts: maintainer_app_version_YYYY-MM-DD-HH-MM-SS.yml",
		},
		{
			name:          "invalid timestamp",
			inputFile:     "samplemaintainer_sampleapp_v1.2.3_2025-01-02.yml",
			expectedError: "invalid timestamp, expected YYYY-MM-DD-HH-MM-SS",
		},
		{
			name:          "empty maintainer",
			inputFile:     "_store_v1.2.3_2025-01-02-03-04-05.yml",
			expectedError: "maintainer, appName, version must be non-empty",
		},
		{
			name:          "empty app name",
			inputFile:     "samplemaintainer__v1.2.3_2025-01-02-03-04-05.yml",
			expectedError: "maintainer, appName, version must be non-empty",
		},
		{
			name:          "empty version",
			inputFile:     "samplemaintainer_sampleapp__2025-01-02-03-04-05.yml",
			expectedError: "maintainer, appName, version must be non-empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualDto, err := encoder.DecodeComposeArchiveName(testCase.inputFile)

			if testCase.expectedError != "" {
				assert.NotNil(t, err)
				assert.Equal(t, testCase.expectedError, u.ExtractError(err))
				assert.Nil(t, actualDto)
				return
			}

			assert.Nil(t, err)
			assert.NotNil(t, actualDto)
			assert.Equal(t, testCase.expectedDto.Maintainer, actualDto.Maintainer)
			assert.Equal(t, testCase.expectedDto.AppName, actualDto.AppName)
			assert.Equal(t, testCase.expectedDto.Version, actualDto.Version)
			assert.Equal(t, testCase.expectedDto.VersionCreationTimestamp, actualDto.VersionCreationTimestamp)
		})
	}
}

func TestComposeArchiveNameDecodeEncodeRoundTrip(t *testing.T) {
	inputFile := "samplemaintainer_sampleapp_v1.2.3_2025-01-02-03-04-05.yml"

	decodedDto, err := encoder.DecodeComposeArchiveName(inputFile)
	assert.Nil(t, err)
	assert.NotNil(t, decodedDto)

	encodedFile, err := encoder.EncodeComposeArchiveName(decodedDto)
	assert.Nil(t, err)
	assert.Equal(t, inputFile, encodedFile)
}
