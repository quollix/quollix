package api_client

import (
	"io"
	"net/http"
	"server/tools"

	u "github.com/quollix/common/utils"
)

type QuollixFrontendClient struct {
	quollix *QuollixClient
}

type FrontendResponse struct {
	StatusCode int
	Header     http.Header
	Body       string
}

func (c *QuollixFrontendClient) Reload() error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendReloadFrontendTemplatesFromFileSystem, nil)
	return err
}

func (c *QuollixFrontendClient) GetPage(path string) (*FrontendResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.quollix.Parent.RootUrl+path, nil)
	if err != nil {
		return nil, err
	}

	if c.quollix.Parent.SetCookieHeader && c.quollix.Parent.Cookie != nil {
		req.AddCookie(c.quollix.Parent.Cookie)
	}

	client := &http.Client{
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: httpClient.Transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer u.Close(resp.Body)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &FrontendResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       string(bodyBytes),
	}, nil
}
