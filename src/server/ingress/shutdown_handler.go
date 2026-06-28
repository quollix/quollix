package ingress

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"server/apps_basic"
	"syscall"
	"time"
)

var (
	serverShutdownTimeout       = 30 * time.Second
	appOperationShutdownTimeout = 2 * time.Minute
)

type ShutdownHandlerImpl struct {
	OperationRegistry apps_basic.OperationRegistry
}

func (h *ShutdownHandlerImpl) WaitAndShutdown(servers []*http.Server) {
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(shutdownSignal)

	<-shutdownSignal

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancelShutdown()

	for _, server := range servers {
		_ = server.Shutdown(shutdownContext)
	}

	if !h.waitUntilNoOperation() {
		for _, server := range servers {
			_ = server.Close()
		}
		os.Exit(1)
	}
}

func (h *ShutdownHandlerImpl) waitUntilNoOperation() bool {
	deadlineTimer := time.NewTimer(appOperationShutdownTimeout)
	defer deadlineTimer.Stop()

	retryTicker := time.NewTicker(200 * time.Millisecond)
	defer retryTicker.Stop()

	for {
		handle, err := h.OperationRegistry.TryBlockGlobalOperation("shutdown")
		if err == nil {
			handle.Done()
			return true
		}

		select {
		case <-retryTicker.C:
		case <-deadlineTimer.C:
			return false
		}
	}
}
