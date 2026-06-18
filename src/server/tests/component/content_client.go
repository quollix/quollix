package component

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"server/tools"
	"strings"
	"time"

	u "github.com/quollix/common/utils"
)

const (
	sampleEndpoint       = "/"
	sampleBodyV20        = "this is version 2.0"
	contentRetryTimeout  = 3 * time.Second
	contentRetryInterval = 50 * time.Millisecond
)

type QuollixContentClient struct {
	quollix *QuollixClient
}

func (c *QuollixContentClient) GetSecret() (string, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSecret, nil)
	if err != nil {
		return "", err
	}

	var secret string
	err = json.Unmarshal(body, &secret)
	if err != nil {
		return "", err
	}

	return secret, nil
}

func (c *QuollixContentClient) AssertSampleContent(anonymousCall bool) error {
	if anonymousCall {
		return c.assertEndpoint(sampleBodyV20, nil)
	}
	return c.assertEndpoint(sampleBodyV20, c.quollix.Parent.Cookie)
}

func (c *QuollixContentClient) AssertContent(expected string) error {
	return c.assertEndpoint(expected, c.quollix.Parent.Cookie)
}

func (c *QuollixContentClient) AssertSampleContentWithCookie(cookie *http.Cookie) error {
	return c.assertEndpoint(sampleBodyV20, cookie)
}

func (c *QuollixContentClient) ReadSampleAppEnvValue(name string) (string, error) {
	var body string
	err := tools.EventuallyWithTimeout(contentRetryTimeout, contentRetryInterval, func() error {
		status, responseBody, requestErr := c.request("/env/"+name, c.quollix.Parent.Cookie)
		if requestErr != nil {
			return requestErr
		}
		if status != http.StatusOK {
			return u.Logger.NewError("request failed", "status_code", status, "response_body", responseBody)
		}
		body = responseBody
		return nil
	})
	if err != nil {
		return "", err
	}
	return body, nil
}

func (c *QuollixContentClient) AssertContentUsingSecret(secret string) error {
	if err := c.ExchangeSecretForAppCookie(secret); err != nil {
		return err
	}
	return c.AssertContent(sampleBodyV20)
}

func (c *QuollixContentClient) ExchangeSecretForAppCookie(secret string) error {
	return tools.EventuallyWithTimeout(contentRetryTimeout, contentRetryInterval, func() error {
		req, err := http.NewRequest("POST", getAppRequestUrlWithSecret(sampleEndpoint, secret), nil)
		if err != nil {
			return err
		}
		if c.quollix.Parent.Cookie != nil {
			req.AddCookie(c.quollix.Parent.Cookie)
		}
		client := &http.Client{
			CheckRedirect: func(request *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // #nosec G402: component tests intentionally connect to the local test certificate
				},
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer u.Close(resp.Body)
		if resp.StatusCode != http.StatusFound {
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return readErr
			}
			return u.Logger.NewError("request failed", "status_code", resp.StatusCode, "response_body", strings.TrimSuffix(string(body), "\n"))
		}
		cookies := resp.Cookies()
		if len(cookies) != 1 {
			return u.Logger.NewError("expected app cookie", "cookie_count", len(cookies))
		}
		c.quollix.Parent.Cookie = cookies[0]
		return nil
	})
}

func (c *QuollixContentClient) StoreStringInSampleApp(stringToStore string) error {
	req, err := http.NewRequest("POST", "http://sampleapp.localhost/save-string", strings.NewReader(stringToStore))
	if err != nil {
		return err
	}
	req.AddCookie(c.quollix.Parent.Cookie)
	_, err = httpClient.Do(req)
	return err
}

func (c *QuollixContentClient) ReadStringFromSampleApp() (string, error) {
	req, err := http.NewRequest("POST", "http://sampleapp.localhost/read-string", nil)
	if err != nil {
		return "", err
	}
	req.AddCookie(c.quollix.Parent.Cookie)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer u.Close(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *QuollixContentClient) assertEndpoint(expected string, cookie *http.Cookie) error {
	return tools.EventuallyWithTimeout(contentRetryTimeout, contentRetryInterval, func() error {
		status, body, err := c.request(sampleEndpoint, cookie)
		if err != nil {
			return err
		}
		return assertResponse(status, body, expected)
	})
}

func (c *QuollixContentClient) request(endpoint string, cookie *http.Cookie) (int, string, error) {
	url := "https://sampleapp.localhost" + endpoint
	return c.doGeneralGetContentRequest(url, cookie)
}

func (c *QuollixContentClient) doGeneralGetContentRequest(url string, cookie *http.Cookie) (int, string, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return 0, "", err
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer u.Close(resp.Body)
	if len(resp.Cookies()) != 0 {
		c.quollix.Parent.Cookie = resp.Cookies()[0]
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	trimmedBody := strings.TrimSuffix(string(body), "\n")
	return resp.StatusCode, trimmedBody, nil
}

func getAppRequestUrlWithSecret(endpoint, secret string) string {
	return fmt.Sprintf("https://sampleapp.localhost%s?%s=%s", endpoint, tools.BrandAppQuerySecretName, secret)
}

func assertResponse(actualStatusCode int, actualResponseBody string, expectedResponseBody string) error {
	if actualStatusCode != http.StatusOK || actualResponseBody != expectedResponseBody {
		return u.Logger.NewError("request failed", "status_code", actualStatusCode, "response_body", actualResponseBody)
	}
	return nil
}
