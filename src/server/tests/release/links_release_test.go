//go:build release

package release

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"server/tools"
	"sort"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/utils"
)

const (
	linkCheckAttempts  = 10
	linkCheckRetryWait = time.Second
	linkCheckUserAgent = "quollix-link-checker"
)

func TestReleaseLinks_Accessible(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	links, err := collectLinksFromStruct(reflect.ValueOf(tools.Links))
	assert.Nil(t, err)

	for _, link := range links {
		assertLinkAccessible(t, client, link)
	}
}

func TestReleaseInstalledAppDocsLinks_Accessible(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	for _, appName := range tools.OfficialAppNames {
		assertLinkAccessible(t, client, tools.InstalledAppDocsUrl(appName))
	}
}

func collectLinksFromStruct(value reflect.Value) ([]string, error) {
	linksByURL := map[string]bool{}
	if err := collectLinksFromValue(value, "Links", linksByURL); err != nil {
		return nil, err
	}

	links := make([]string, 0, len(linksByURL))
	for link := range linksByURL {
		links = append(links, link)
	}

	sort.Strings(links)
	return links, nil
}

func collectLinksFromValue(value reflect.Value, fieldPath string, linksByURL map[string]bool) error {
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if err := collectLinksFromValue(value.Field(i), fieldPath+"."+field.Name, linksByURL); err != nil {
				return err
			}
		}
	case reflect.String:
		link := value.String()
		if link != "" {
			linksByURL[link] = true
		}
	default:
		return utils.Logger.NewError("unsupported field kind", "field_path", fieldPath, "kind", value.Kind())
	}

	return nil
}

func assertLinkAccessible(t *testing.T, client *http.Client, link string) {
	fmt.Printf("Checking link: %s\n", link)

	var lastResponse *http.Response
	var lastErr error
	for attempt := 1; attempt <= linkCheckAttempts; attempt++ {
		response, err := requestLink(client, link)
		lastResponse = response
		lastErr = err
		if err == nil && response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusInternalServerError {
			break
		}

		if err == nil && response.StatusCode < http.StatusInternalServerError {
			break
		}

		if attempt < linkCheckAttempts {
			time.Sleep(linkCheckRetryWait)
		}
	}

	if lastErr != nil {
		t.Fatalf("link %s request failed: %v", link, lastErr)
	}
	if lastResponse.StatusCode < http.StatusOK || lastResponse.StatusCode >= http.StatusBadRequest {
		t.Fatalf("link %s returned status %d from %s", link, lastResponse.StatusCode, lastResponse.Request.URL.String())
	}
}

func requestLink(client *http.Client, link string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", linkCheckUserAgent)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	_, copyErr := io.Copy(io.Discard, response.Body)
	closeErr := response.Body.Close()
	if copyErr != nil {
		return response, copyErr
	}
	if closeErr != nil {
		return response, closeErr
	}
	return response, nil
}
