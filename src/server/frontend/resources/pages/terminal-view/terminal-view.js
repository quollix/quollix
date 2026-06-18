function createTerminal(terminalElement) {
    const terminal = new Terminal({
        cursorBlink: true,
        screenReaderMode: true,
    });
    const fitAddon = new FitAddon.FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.open(terminalElement);
    return { terminal, fitAddon };
}

function attachTerminalAccessibilityId(terminalElement) {
    const accessibilityTree = terminalElement.querySelector(".xterm-accessibility-tree")
    if (accessibilityTree) accessibilityTree.id = "terminal-output-accessibility"
}

function writeWebsocketPayloadToTerminal(terminal, event) {
    if (event.data instanceof ArrayBuffer) {
        terminal.write(new Uint8Array(event.data));
        return Promise.resolve();
    }
    if (event.data instanceof Blob) {
        return event.data.arrayBuffer().then((arrayBuffer) => {
            terminal.write(new Uint8Array(arrayBuffer));
        });
    }
    terminal.write(String(event.data));
    return Promise.resolve();
}

function sendWebsocketJson(websocketConnection, payload) {
    if (websocketConnection.readyState !== WebSocket.OPEN) return;
    websocketConnection.send(JSON.stringify(payload));
}

function attachTerminalResizeSync(terminalElement, terminal, fitAddon, websocketConnection) {
    const syncSizeToBackend = () => {
        fitAddon.fit();
        sendWebsocketJson(websocketConnection, {
            type: "resize",
            cols: terminal.cols,
            rows: terminal.rows,
        });
    };

    const resizeObserver = new ResizeObserver(() => syncSizeToBackend());
    resizeObserver.observe(terminalElement);
    window.addEventListener("resize", () => syncSizeToBackend());

    return { syncSizeToBackend };
}

function attachTerminalInput(terminal, websocketConnection) {
    terminal.onData((data) => {
        sendWebsocketJson(websocketConnection, { type: "input", data });
    });
}

function createTerminalWebsocket(websocketUrl, terminal) {
    const websocketConnection = new WebSocket(websocketUrl);
    websocketConnection.binaryType = "arraybuffer";

    websocketConnection.onmessage = (event) => {
        void writeWebsocketPayloadToTerminal(terminal, event);
    };

    websocketConnection.onclose = () => terminal.write("\r\n[disconnected]\r\n");
    websocketConnection.onerror = () => terminal.write("\r\n[websocket error]\r\n");

    return websocketConnection;
}

window.initTerminalViewPage = function (host, maintainer, appName, serviceName) {
    const terminalElement = document.getElementById("terminal");
    if (!terminalElement) return;

    const websocketQuery =
        `maintainer=${encodeURIComponent(maintainer)}` +
        `&appName=${encodeURIComponent(appName)}` +
        `&serviceName=${encodeURIComponent(serviceName)}`;

    const websocketUrl = `wss://${host}{{$.Paths.BackendTerminal}}?${websocketQuery}`;

    const { terminal, fitAddon } = createTerminal(terminalElement);
    attachTerminalAccessibilityId(terminalElement)
    window.activeTerminal = terminal;
    const websocketConnection = createTerminalWebsocket(websocketUrl, terminal);

    const { syncSizeToBackend } = attachTerminalResizeSync(
        terminalElement,
        terminal,
        fitAddon,
        websocketConnection,
    );

    attachTerminalInput(terminal, websocketConnection);

    websocketConnection.onopen = () => {
        terminal.focus();
        syncSizeToBackend();
        setTimeout(syncSizeToBackend, 0);
        setTimeout(() => attachTerminalAccessibilityId(terminalElement), 0);
    };
}

window.copyTerminalToClipboard = async function () {
    const terminal = window.activeTerminal;
    if (!terminal) return;

    const activeBuffer = terminal.buffer.active;
    const lines = [];

    for (let lineIndex = 0; lineIndex < activeBuffer.length; lineIndex++) {
        const line = activeBuffer.getLine(lineIndex);
        if (line) lines.push(line.translateToString(true));
    }

    const text = lines.join("\n").replace(/\s+$/g, "");
    await navigator.clipboard.writeText(text);
    window.showSnackbar("Terminal content copied to clipboard");
};
