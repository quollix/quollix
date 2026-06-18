package terminal

import (
	"net/http"

	"github.com/gorilla/websocket"
	u "github.com/quollix/common/utils"
)

type TerminalHandlerImpl struct {
	DockerTerminalService DockerTerminalService
	// Leave CheckOrigin unset so gorilla/websocket keeps its default same-origin protection.
	// See https://pkg.go.dev/github.com/gorilla/websocket for the Upgrader docs.
	websocketUpgrader websocket.Upgrader `wire:"-"`
}

func (t *TerminalHandlerImpl) ServeTerminalWebsocket(responseWriter http.ResponseWriter, request *http.Request) {
	websocketConnection, websocketError := t.websocketUpgrader.Upgrade(responseWriter, request, nil)
	if websocketError != nil {
		return
	}
	defer u.Close(websocketConnection)

	maintainer := request.URL.Query().Get("maintainer")
	appName := request.URL.Query().Get("appName")
	serviceName := request.URL.Query().Get("serviceName")
	if maintainer == "" || appName == "" || serviceName == "" {
		_ = websocketConnection.WriteMessage(websocket.TextMessage, []byte("missing query params: maintainer, appName, serviceName\n"))
		return
	}

	containerName := maintainer + "_" + appName + "_" + serviceName

	readClientMessage := func() (messageType int, messageBytes []byte, readError error) {
		return websocketConnection.ReadMessage()
	}

	writeToClientBinary := func(messageBytes []byte) error {
		return websocketConnection.WriteMessage(websocket.BinaryMessage, messageBytes)
	}

	runError := t.DockerTerminalService.RunTerminalSession(
		request.Context(),
		containerName,
		readClientMessage,
		writeToClientBinary,
	)
	if runError != nil {
		_ = websocketConnection.WriteMessage(websocket.TextMessage, []byte(runError.Error()+"\n"))
	}
}
