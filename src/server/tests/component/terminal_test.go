//go:build component

package component

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"server/terminal"
	"server/tools"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const (
	terminalReadTimeout = 10 * time.Second
	terminalCols        = 120
	terminalRows        = 40
)

const (
	expectedShellBanner = "Shell: /bin/sh"
	expectedPrompt      = "/opt/server #"
	testCommandInput    = "echo test-message\r"
	expectedCommandEcho = "echo test-message"
	expectedCommandOut  = "test-message"
)

var terminalControlSequencePattern = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

func normalizeNewlines(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return value
}

func normalizeTerminalOutput(value string) string {
	value = normalizeNewlines(value)
	return terminalControlSequencePattern.ReplaceAllString(value, "")
}

type terminalWebsocketClient struct {
	websocketConnection *websocket.Conn
	receivedBuilder     strings.Builder
}

func (client *terminalWebsocketClient) sendJson(message terminal.TerminalClientMessage) error {
	messageBytes, marshalError := json.Marshal(message)
	if marshalError != nil {
		return marshalError
	}
	return client.websocketConnection.WriteMessage(websocket.TextMessage, messageBytes)
}

func (client *terminalWebsocketClient) sendResize(cols int, rows int) error {
	return client.sendJson(terminal.TerminalClientMessage{Type: "resize", Cols: cols, Rows: rows})
}

func (client *terminalWebsocketClient) sendInput(input string) error {
	return client.sendJson(terminal.TerminalClientMessage{Type: "input", Data: input})
}

func (client *terminalWebsocketClient) waitUntil(
	readContext context.Context,
	predicate func(accumulated string) bool,
) (string, error) {
	for {
		select {
		case <-readContext.Done():
			return normalizeNewlines(client.receivedBuilder.String()), readContext.Err()
		default:
		}

		messageType, messageBytes, readError := client.websocketConnection.ReadMessage()
		if readError != nil {
			return normalizeNewlines(client.receivedBuilder.String()), readError
		}
		if messageType != websocket.BinaryMessage && messageType != websocket.TextMessage {
			continue
		}

		client.receivedBuilder.Write(messageBytes)
		normalized := normalizeTerminalOutput(client.receivedBuilder.String())
		if predicate(normalized) {
			return normalized, nil
		}
	}
}

func ensureContainsInOrder(output string, expectedSubstrings []string) error {
	searchStartIndex := 0
	for _, expectedSubstring := range expectedSubstrings {
		foundIndex := strings.Index(output[searchStartIndex:], expectedSubstring)
		if foundIndex < 0 {
			return fmt.Errorf("missing %q in output:\n%s", expectedSubstring, output)
		}
		searchStartIndex += foundIndex + len(expectedSubstring)
	}
	return nil
}

func TestTerminalWebsocket_Echo(t *testing.T) {
	client := GetClientAndLogin(t)

	parsedRootUrl, parseError := url.Parse(client.Parent.RootUrl)
	assert.Nil(t, parseError)

	websocketUrl, requestHeader, websocketDialer := prepateWebsocketConnection(parsedRootUrl, client.Parent.Cookie)

	websocketConnection, _, dialError := websocketDialer.Dial(websocketUrl.String(), requestHeader)
	assert.Nil(t, dialError)
	defer u.Close(websocketConnection)

	terminalClient := terminalWebsocketClient{websocketConnection: websocketConnection}

	readContext, cancel := context.WithTimeout(context.Background(), terminalReadTimeout)
	defer cancel()

	assert.Nil(t, terminalClient.sendResize(terminalCols, terminalRows))

	_, readyError := terminalClient.waitUntil(readContext, func(accumulated string) bool {
		return strings.Contains(accumulated, expectedShellBanner+"\n") && strings.Contains(accumulated, expectedPrompt)
	})
	assert.Nil(t, readyError)

	assert.Nil(t, terminalClient.sendInput(testCommandInput))

	expectedInOrder := []string{
		expectedPrompt,
		expectedCommandEcho,
		expectedCommandOut,
		expectedPrompt,
	}

	finalOutput, outputError := terminalClient.waitUntil(readContext, func(accumulated string) bool {
		return ensureContainsInOrder(accumulated, expectedInOrder) == nil
	})
	assert.Nil(t, outputError)
	assert.True(t, strings.Contains(finalOutput, expectedShellBanner))
	assert.Nil(t, ensureContainsInOrder(finalOutput, expectedInOrder))
}

func TestTerminalWebsocket_OnlyAdminCanConnect(t *testing.T) {
	adminClient := GetClientAndLogin(t)
	defer adminClient.Test.ResetTestState()

	InviteUserAndSetPassword(adminClient, SampleUsername, SampleUserPassword, SampleUserEmail)

	userClient := GetQuollixClient(t)
	assert.Nil(t, userClient.Auth.Login(SampleUsername, SampleUserPassword))

	parsedRootUrl, parseError := url.Parse(adminClient.Parent.RootUrl)
	assert.Nil(t, parseError)

	t.Run("admin", func(t *testing.T) {
		websocketUrl, requestHeader, websocketDialer := prepateWebsocketConnection(parsedRootUrl, adminClient.Parent.Cookie)
		websocketConnection, response, dialError := websocketDialer.Dial(websocketUrl.String(), requestHeader)
		assert.Nil(t, dialError)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusSwitchingProtocols, response.StatusCode)
		defer u.Close(websocketConnection)
	})

	t.Run("authenticated user", func(t *testing.T) {
		websocketUrl, requestHeader, websocketDialer := prepateWebsocketConnection(parsedRootUrl, userClient.Parent.Cookie)
		websocketConnection, response, dialError := websocketDialer.Dial(websocketUrl.String(), requestHeader)
		assert.NotNil(t, dialError)
		assert.NotNil(t, response)
		assert.NotEqual(t, http.StatusSwitchingProtocols, response.StatusCode)
		assert.Nil(t, websocketConnection)
	})

	t.Run("anonymous", func(t *testing.T) {
		websocketUrl, requestHeader, websocketDialer := prepateWebsocketConnection(parsedRootUrl, nil)
		websocketConnection, response, dialError := websocketDialer.Dial(websocketUrl.String(), requestHeader)
		assert.NotNil(t, dialError)
		assert.NotNil(t, response)
		assert.NotEqual(t, http.StatusSwitchingProtocols, response.StatusCode)
		assert.Nil(t, websocketConnection)
	})
}

func prepateWebsocketConnection(parsedRootUrl *url.URL, cookie *http.Cookie) (url.URL, http.Header, websocket.Dialer) {
	websocketUrl := url.URL{
		Scheme: "wss",
		Host:   parsedRootUrl.Host,
		Path:   tools.Paths.BackendTerminal,
		RawQuery: url.Values{
			"maintainer":  []string{"quollix"},
			"appName":     []string{"quollix"},
			"serviceName": []string{"quollix"},
		}.Encode(),
	}

	requestHeader := http.Header{}
	if cookie != nil {
		requestHeader.Add("Cookie", (&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		}).String())
	}

	websocketDialer := websocket.Dialer{}
	websocketDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return websocketUrl, requestHeader, websocketDialer
}
