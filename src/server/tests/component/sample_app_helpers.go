package component

import (
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"time"

	"server/tests/api_client"
	"server/tools"

	u "github.com/quollix/common/utils"
)

const (
	sampleEndpoint       = "/"
	sampleAppHttpsUrl    = "https://sampleapp.localhost"
	sampleBodyV20        = "this is version 2.0"
	contentRetryTimeout  = 3 * time.Second
	contentRetryInterval = 50 * time.Millisecond
)

var sampleAppHttpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // #nosec G402: tests intentionally connect to the local test certificate
		},
	},
}

func AssertSampleAppDefaultContent(client *api_client.QuollixClient, anonymousCall bool) error {
	if anonymousCall {
		return assertSampleAppEndpoint(sampleBodyV20, nil)
	}
	return assertSampleAppEndpoint(sampleBodyV20, client.Parent.Cookie)
}

func AssertSampleAppContent(client *api_client.QuollixClient, expected string) error {
	return assertSampleAppEndpoint(expected, client.Parent.Cookie)
}

func AssertSampleAppContentWithCookie(cookie *http.Cookie) error {
	return assertSampleAppEndpoint(sampleBodyV20, cookie)
}

func AssertSampleAppContentUsingSecret(client *api_client.QuollixClient, secret string) error {
	if err := ExchangeAppAccessSecretForCookie(client, secret); err != nil {
		return err
	}
	return AssertSampleAppContent(client, sampleBodyV20)
}

func ExchangeAppAccessSecretForCookie(client *api_client.QuollixClient, secret string) error {
	return ExchangeAppAccessSecretForCookieWithUrl(client, secret, sampleAppHttpsUrl+sampleEndpoint)
}

func ExchangeAppAccessSecretForCookieWithUrl(client *api_client.QuollixClient, secret string, appUrl string) error {
	return tools.EventuallyWithTimeout(contentRetryTimeout, contentRetryInterval, func() error {
		cookie, err := client.AppAccess.ExchangeSecretForAppAccessCookie(secret, appUrl)
		if err != nil {
			return err
		}
		client.Parent.Cookie = cookie
		return nil
	})
}

func ReadSampleAppEnvValue(client *api_client.QuollixClient, name string) (string, error) {
	return readSampleAppUrl(client, sampleAppHttpsUrl+"/env/"+name)
}

func ReadSampleAppHeaderValue(client *api_client.QuollixClient, baseUrl string, name string) (string, error) {
	return readSampleAppUrl(client, strings.TrimSuffix(baseUrl, "/")+"/header/"+name)
}

func readSampleAppUrl(client *api_client.QuollixClient, appUrl string) (string, error) {
	var body string
	err := tools.EventuallyWithTimeout(contentRetryTimeout, contentRetryInterval, func() error {
		status, responseBody, requestErr := requestSampleAppWithClientUrl(client, appUrl)
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

func requestSampleAppWithClientUrl(client *api_client.QuollixClient, appUrl string) (int, string, error) {
	resp, err := client.AppAccess.DoAppRequest(client.Parent.Cookie, http.MethodPost, appUrl, nil)
	if err != nil {
		return 0, "", err
	}
	defer u.Close(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, strings.TrimSuffix(string(body), "\n"), nil
}

func StoreStringInSampleApp(client *api_client.QuollixClient, stringToStore string) error {
	resp, err := client.AppAccess.DoAppRequest(client.Parent.Cookie, http.MethodPost, "http://sampleapp.localhost/save-string", strings.NewReader(stringToStore))
	if resp != nil {
		defer u.Close(resp.Body)
	}
	return err
}

func ReadStringFromSampleApp(client *api_client.QuollixClient) (string, error) {
	resp, err := client.AppAccess.DoAppRequest(client.Parent.Cookie, http.MethodPost, "http://sampleapp.localhost/read-string", nil)
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

func assertSampleAppEndpoint(expected string, cookie *http.Cookie) error {
	return tools.EventuallyWithTimeout(contentRetryTimeout, contentRetryInterval, func() error {
		status, body, err := requestSampleApp(sampleEndpoint, cookie)
		if err != nil {
			return err
		}
		return assertResponse(status, body, expected)
	})
}

func requestSampleApp(endpoint string, cookie *http.Cookie) (int, string, error) {
	return requestSampleAppWithUrl(sampleAppHttpsUrl+endpoint, cookie)
}

func requestSampleAppWithUrl(appUrl string, cookie *http.Cookie) (int, string, error) {
	req, err := http.NewRequest("POST", appUrl, nil)
	if err != nil {
		return 0, "", err
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := sampleAppHttpClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer u.Close(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, strings.TrimSuffix(string(body), "\n"), nil
}

func assertResponse(actualStatusCode int, actualResponseBody string, expectedResponseBody string) error {
	if actualStatusCode != http.StatusOK || actualResponseBody != expectedResponseBody {
		return u.Logger.NewError("request failed", "status_code", actualStatusCode, "response_body", actualResponseBody)
	}
	return nil
}
